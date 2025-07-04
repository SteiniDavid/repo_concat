package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"bufio"

	"flag"
	
	"github.com/fatih/color"
	"repo-concat/cli"
	"repo-concat/tui"
)

type Config struct {
	githubURL    string
	localPath    string
	exclusions   []string
	inclusions   []string
	peek         bool
	outputDir    string
	tokenEst     bool
	enableTUI    bool
}

type CacheEntry struct {
	URL        string    `json:"url"`
	CachedAt   time.Time `json:"cached_at"`
	RepoPath   string    `json:"repo_path"`
	ExpiresAt  time.Time `json:"expires_at"`
}


func main() {
	var config Config
	var exclusionFlags stringSlice
	var inclusionFlags stringSlice

	flag.StringVar(&config.githubURL, "url", "", "GitHub repository URL")
	flag.StringVar(&config.localPath, "path", "", "Local directory path")
	flag.Var(&exclusionFlags, "exclude", "Regex patterns or path patterns (/dir) to exclude files (can be used multiple times)")
	flag.Var(&inclusionFlags, "include", "Regex patterns or path patterns (/dir) to include files (if specified, only matching files are included)")
	flag.BoolVar(&config.peek, "peek", false, "Show folder structure and dry run before processing")
	flag.StringVar(&config.outputDir, "output", ".", "Output directory for concatenated file")
	flag.BoolVar(&config.tokenEst, "tokens", true, "Estimate token count")
	flag.BoolVar(&config.enableTUI, "tui", false, "Enable modern TUI interface")

	flag.Parse()

	config.exclusions = []string(exclusionFlags)
	config.inclusions = []string(inclusionFlags)

	// Launch TUI mode if requested
	if config.enableTUI {
		tuiConfig := tui.Config{
			URL:       config.githubURL,
			Path:      config.localPath,
			Include:   config.inclusions,
			Exclude:   config.exclusions,
			Output:    config.outputDir,
			EnableTUI: true,
		}
		
		if err := tui.RunTUI(tuiConfig); err != nil {
			fmt.Printf("TUI error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if config.githubURL == "" && config.localPath == "" {
		fmt.Println(cli.ErrorMsg("Configuration Error", 
			"Either GitHub URL or local directory path is required",
			"Use -url for GitHub repositories or -path for local directories"))
		flag.Usage()
		os.Exit(1)
	}

	if config.githubURL != "" && config.localPath != "" {
		fmt.Println(cli.ErrorMsg("Configuration Error", 
			"Cannot specify both GitHub URL and local directory path",
			"Use either -url OR -path, not both"))
		flag.Usage()
		os.Exit(1)
	}

	if err := processRepository(config); err != nil {
		log.Fatal(err)
	}
}

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func getTmpCacheDir() string {
	return filepath.Join("/tmp", "repo-concat-cache")
}

func urlToHash(githubURL string) string {
	hash := md5.Sum([]byte(githubURL))
	return hex.EncodeToString(hash[:])
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "just now"
	}
	if d < time.Minute {
		seconds := int(d.Seconds())
		if seconds == 1 {
			return "1 second"
		}
		return fmt.Sprintf("%d seconds", seconds)
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	hours := int(d.Hours())
	if hours == 1 {
		return "1 hour"
	}
	return fmt.Sprintf("%d hours", hours)
}

func getCachedRepo(githubURL string) (string, bool, time.Time, error) {
	cacheDir := getTmpCacheDir()
	urlHash := urlToHash(githubURL)
	metadataPath := filepath.Join(cacheDir, urlHash+".json")

	// Check if metadata file exists
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return "", false, time.Time{}, nil
	}

	// Read metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return "", false, time.Time{}, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return "", false, time.Time{}, err
	}

	// Check if cache is still valid
	if time.Now().After(entry.ExpiresAt) {
		// Cache expired, clean up
		os.Remove(metadataPath)
		os.RemoveAll(entry.RepoPath)
		return "", false, time.Time{}, nil
	}

	// Check if repo directory still exists
	if _, err := os.Stat(entry.RepoPath); os.IsNotExist(err) {
		// Repo directory missing, clean up metadata
		os.Remove(metadataPath)
		return "", false, time.Time{}, nil
	}

	return entry.RepoPath, true, entry.CachedAt, nil
}

func cacheRepo(githubURL, repoPath string) error {
	cacheDir := getTmpCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	urlHash := urlToHash(githubURL)
	metadataPath := filepath.Join(cacheDir, urlHash+".json")

	entry := CacheEntry{
		URL:       githubURL,
		CachedAt:  time.Now(),
		RepoPath:  repoPath,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

func processRepository(config Config) error {
	var repoPath string
	var shouldCleanup bool

	// Handle local directory path
	if config.localPath != "" {
		// Validate local path exists
		if _, err := os.Stat(config.localPath); os.IsNotExist(err) {
			return fmt.Errorf("local directory does not exist: %s", config.localPath)
		}
		fmt.Println(cli.StatusMsg("info", "Processing local directory: "+config.localPath))
		repoPath = config.localPath
		shouldCleanup = false
	} else {
		// Handle GitHub URL - check tmp cache first
		if cachedPath, found, cachedAt, err := getCachedRepo(config.githubURL); err != nil {
			fmt.Println(cli.StatusMsg("warning", fmt.Sprintf("Cache check failed: %v", err)))
		} else if found {
			age := time.Since(cachedAt)
			fmt.Println(cli.StatusMsg("success", fmt.Sprintf("Using cached repository (cached %s ago)", formatDuration(age))))
			repoPath = cachedPath
			shouldCleanup = false
		}

		if repoPath == "" {
			// No cache found, clone repository
			tempDir, err := os.MkdirTemp("", "repo-concat-*")
			if err != nil {
				return fmt.Errorf("failed to create temp directory: %w", err)
			}

			fmt.Println(cli.StatusMsg("loading", "Cloning repository: "+config.githubURL))
			
			if err := cloneRepository(config.githubURL, tempDir); err != nil {
				os.RemoveAll(tempDir)
				return fmt.Errorf("failed to clone repository: %w", err)
			}
			
			fmt.Println(cli.StatusMsg("success", "Repository cloned successfully"))

			repoName := extractRepoName(config.githubURL)
			repoPath = filepath.Join(tempDir, repoName)
			shouldCleanup = true

			// Cache the repository in tmp
			cacheDir := getTmpCacheDir()
			if err := os.MkdirAll(cacheDir, 0755); err == nil {
				cachedRepoPath := filepath.Join(cacheDir, urlToHash(config.githubURL))
				if err := os.RemoveAll(cachedRepoPath); err == nil {
					if err := os.Rename(repoPath, cachedRepoPath); err == nil {
						repoPath = cachedRepoPath
						shouldCleanup = false
						if err := cacheRepo(config.githubURL, cachedRepoPath); err != nil {
							color.Yellow("⚠️  Warning: failed to cache repository metadata: %v", err)
						}
					}
				}
			}

			if shouldCleanup {
				defer os.RemoveAll(tempDir)
			}
		}
	}

	if config.peek {
		fmt.Println()
		fmt.Println(cli.SimpleHeader("📋 Repository Preview"))
		fmt.Println()
		
		dryRunFiles, excludedFiles, err := performDryRun(repoPath, config.exclusions, config.inclusions)
		if err != nil {
			return fmt.Errorf("failed to perform dry run: %w", err)
		}

		// Convert file paths to relative paths for cleaner display
		var relativeFiles []string
		for _, file := range dryRunFiles {
			if rel, err := filepath.Rel(repoPath, file); err == nil {
				relativeFiles = append(relativeFiles, rel)
			} else {
				relativeFiles = append(relativeFiles, file)
			}
		}
		
		// Show simple tree with meaningful name
		displayName := filepath.Base(repoPath)
		if config.localPath != "" {
			displayName = config.localPath
		} else if config.githubURL != "" {
			displayName = extractRepoName(config.githubURL)
		}
		fmt.Println(cli.SimpleTree(displayName, relativeFiles, nil))
		fmt.Println()

		// Simple summary
		fmt.Println(cli.SimpleSummary(int64(len(dryRunFiles)), int64(len(excludedFiles)), 0))
		fmt.Println()
		
		if len(dryRunFiles) == 0 {
			fmt.Println(cli.StatusMsg("error", "No files would be included with current filters"))
			return nil
		}

		// Simple confirmation
		fmt.Print(cli.ConfirmPrompt(fmt.Sprintf("Proceed with concatenation of %d files?", len(dryRunFiles)))) 
		fmt.Print(": ")
		
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println(cli.StatusMsg("warning", "Operation cancelled"))
			return nil
		}
	}

	fmt.Println(cli.StatusMsg("loading", "Collecting files..."))
	files, err := collectFiles(repoPath, config.exclusions, config.inclusions)
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}
	fmt.Println(cli.StatusMsg("success", fmt.Sprintf("Found %d files to process", len(files))))

	var outputFileName string
	if config.localPath != "" {
		outputFileName = generateOutputFileNameForPath(config.localPath)
	} else {
		outputFileName = generateOutputFileName(config.githubURL)
	}
	
	// Create output directory structure
	outputSubDir := filepath.Join(config.outputDir, "repo-concat-output")
	if err := os.MkdirAll(outputSubDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	outputPath := filepath.Join(outputSubDir, outputFileName)

	fmt.Println(cli.StatusMsg("loading", "Concatenating files..."))
	content, err := concatenateFiles(files, repoPath)
	if err != nil {
		return fmt.Errorf("failed to concatenate files: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	var tokenCount int
	if config.tokenEst {
		tokenCount = estimateTokens(content)
	}
	
	fmt.Println()
	fmt.Println(cli.Done(outputPath, len(files), tokenCount))

	if err := copyToClipboard(content); err != nil {
		fmt.Println(cli.StatusMsg("warning", "Could not copy to clipboard"))
		fmt.Println(cli.Subtle("  Install xclip (Linux) or use the output file above"))
	} else {
		fmt.Println(cli.StatusMsg("success", "Content copied to clipboard"))
	}

	return nil
}

func cloneRepository(githubURL, destDir string) error {
	cmd := exec.Command("git", "clone", githubURL)
	cmd.Dir = destDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extractRepoName(githubURL string) string {
	parsedURL, err := url.Parse(githubURL)
	if err != nil {
		parts := strings.Split(githubURL, "/")
		if len(parts) > 0 {
			return strings.TrimSuffix(parts[len(parts)-1], ".git")
		}
		return "repository"
	}

	parts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(parts) >= 2 {
		return strings.TrimSuffix(parts[1], ".git")
	}
	return "repository"
}

func showFilteredDirectoryStructure(rootPath string, relevantFiles []string, depth, maxDepth int) error {
	if depth > maxDepth {
		return nil
	}

	// Build a set of relevant directories based on the included files
	relevantDirs := make(map[string]bool)
	relevantDirs[rootPath] = true

	for _, file := range relevantFiles {
		dir := filepath.Dir(file)
		for dir != rootPath && dir != "." && dir != "/" {
			relevantDirs[dir] = true
			dir = filepath.Dir(dir)
		}
	}

	return showFilteredDirectoryStructureRecursive(rootPath, relevantDirs, relevantFiles, depth, maxDepth, rootPath)
}

func showFilteredDirectoryStructureRecursive(path string, relevantDirs map[string]bool, relevantFiles []string, depth, maxDepth int, rootPath string) error {
	if depth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		indent := strings.Repeat("  ", depth)

		if entry.IsDir() {
			// Only show directories that contain relevant files
			if relevantDirs[fullPath] {
				color.HiBlue("%s📁 %s/", indent, entry.Name())
				if depth < maxDepth {
					showFilteredDirectoryStructureRecursive(fullPath, relevantDirs, relevantFiles, depth+1, maxDepth, rootPath)
				}
			}
		} else {
			// Only show files that are in the relevant files list
			for _, relevantFile := range relevantFiles {
				if relevantFile == fullPath {
					color.HiGreen("%s📄 %s", indent, entry.Name())
					break
				}
			}
		}
	}
	return nil
}

func showDirectoryStructure(path string, depth, maxDepth int) error {
	if depth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		indent := strings.Repeat("  ", depth)
		if entry.IsDir() {
			color.HiBlue("%s📁 %s/", indent, entry.Name())
			if depth < maxDepth {
				showDirectoryStructure(filepath.Join(path, entry.Name()), depth+1, maxDepth)
			}
		} else {
			color.HiGreen("%s📄 %s", indent, entry.Name())
		}
	}
	return nil
}

func isPathPattern(pattern string) bool {
	return strings.HasPrefix(pattern, "/")
}

func matchesPathPattern(pattern, relativePath string) bool {
	if !isPathPattern(pattern) {
		return false
	}
	
	// Remove leading slash from pattern
	cleanPattern := strings.TrimPrefix(pattern, "/")
	
	// Split path into components
	pathParts := strings.Split(relativePath, string(filepath.Separator))
	
	// For top-level directory matching, check if first component starts with pattern
	if len(pathParts) > 0 {
		return strings.HasPrefix(pathParts[0], cleanPattern)
	}
	
	return false
}

func matchesPattern(pattern, relativePath, baseName string) bool {
	if isPathPattern(pattern) {
		return matchesPathPattern(pattern, relativePath)
	}
	
	// Regular regex matching
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return compiled.MatchString(relativePath) || compiled.MatchString(baseName)
}

func performDryRun(rootPath string, exclusionPatterns []string, inclusionPatterns []string) ([]string, []string, error) {
	var includedFiles []string
	var excludedFiles []string

	// Validate regex patterns (skip path patterns starting with /)
	for _, pattern := range exclusionPatterns {
		if !isPathPattern(pattern) {
			_, err := regexp.Compile(pattern)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid exclusion regex pattern '%s': %w", pattern, err)
			}
		}
	}

	for _, pattern := range inclusionPatterns {
		if !isPathPattern(pattern) {
			_, err := regexp.Compile(pattern)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid inclusion regex pattern '%s': %w", pattern, err)
			}
		}
	}

	defaultExclusionPatterns := []string{
		`\.git/`,
		`\.gitignore$`,
		`\.DS_Store$`,
		`node_modules/`,
		`\.env$`,
		`\.(jpg|jpeg|png|gif|svg|ico|bmp|tiff|webp)$`,
		`\.(mp4|mov|avi|mkv|webm|flv)$`,
		`\.(mp3|wav|flac|aac|ogg)$`,
		`\.(zip|tar|gz|rar|7z|exe|dmg|pkg)$`,
		`\.(pdf|doc|docx|xls|xlsx|ppt|pptx)$`,
	}

	allExclusionPatterns := append(exclusionPatterns, defaultExclusionPatterns...)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		// Check if it's a text file first
		if !isTextFile(path) {
			excludedFiles = append(excludedFiles, path)
			return nil
		}

		// Check exclusions
		excluded := false
		for _, pattern := range allExclusionPatterns {
			if matchesPattern(pattern, relativePath, filepath.Base(path)) {
				excludedFiles = append(excludedFiles, path)
				excluded = true
				break
			}
		}

		if excluded {
			return nil
		}

		// If inclusions are specified, file must match at least one inclusion pattern
		if len(inclusionPatterns) > 0 {
			matched := false
			for _, pattern := range inclusionPatterns {
				if matchesPattern(pattern, relativePath, filepath.Base(path)) {
					matched = true
					break
				}
			}
			if !matched {
				excludedFiles = append(excludedFiles, path)
				return nil
			}
		}

		includedFiles = append(includedFiles, path)
		return nil
	})

	return includedFiles, excludedFiles, err
}

func collectFiles(rootPath string, exclusionPatterns []string, inclusionPatterns []string) ([]string, error) {
	var files []string

	// Validate regex patterns (skip path patterns starting with /)
	for _, pattern := range exclusionPatterns {
		if !isPathPattern(pattern) {
			_, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid exclusion regex pattern '%s': %w", pattern, err)
			}
		}
	}

	for _, pattern := range inclusionPatterns {
		if !isPathPattern(pattern) {
			_, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid inclusion regex pattern '%s': %w", pattern, err)
			}
		}
	}

	defaultExclusionPatterns := []string{
		`\.git/`,
		`\.gitignore$`,
		`\.DS_Store$`,
		`node_modules/`,
		`\.env$`,
		`\.(jpg|jpeg|png|gif|svg|ico|bmp|tiff|webp)$`,
		`\.(mp4|mov|avi|mkv|webm|flv)$`,
		`\.(mp3|wav|flac|aac|ogg)$`,
		`\.(zip|tar|gz|rar|7z|exe|dmg|pkg)$`,
		`\.(pdf|doc|docx|xls|xlsx|ppt|pptx)$`,
	}

	allExclusionPatterns := append(exclusionPatterns, defaultExclusionPatterns...)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		// Check exclusions first
		for _, pattern := range allExclusionPatterns {
			if matchesPattern(pattern, relativePath, filepath.Base(path)) {
				return nil
			}
		}

		// If inclusions are specified, file must match at least one inclusion pattern
		if len(inclusionPatterns) > 0 {
			matched := false
			for _, pattern := range inclusionPatterns {
				if matchesPattern(pattern, relativePath, filepath.Base(path)) {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		if isTextFile(path) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func isTextFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return false
		}
	}

	return true
}

func concatenateFiles(files []string, rootPath string) (string, error) {
	var result strings.Builder
	
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	result.WriteString(fmt.Sprintf("# Repository Concatenation\n"))
	result.WriteString(fmt.Sprintf("# Generated on: %s\n", timestamp))
	result.WriteString(fmt.Sprintf("# Total files: %d\n\n", len(files)))

	for _, filePath := range files {
		relativePath, err := filepath.Rel(rootPath, filePath)
		if err != nil {
			relativePath = filePath
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Warning: failed to read file %s: %v\n", relativePath, err)
			continue
		}

		result.WriteString(fmt.Sprintf("# File: %s\n", relativePath))
		result.WriteString("```\n")
		result.Write(content)
		if !strings.HasSuffix(string(content), "\n") {
			result.WriteString("\n")
		}
		result.WriteString("```\n\n")
	}

	return result.String(), nil
}



func generateOutputFileName(githubURL string) string {
	repoName := extractRepoName(githubURL)
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-concat-%s.txt", repoName, timestamp)
}

func generateOutputFileNameForPath(localPath string) string {
	dirName := filepath.Base(localPath)
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-concat-%s.txt", dirName, timestamp)
}

func estimateTokens(content string) int {
	words := strings.Fields(content)
	return len(words) * 4 / 3
}

func copyToClipboard(content string) error {
	var cmd *exec.Cmd
	
	switch {
	case commandExists("pbcopy"):
		cmd = exec.Command("pbcopy")
	case commandExists("xclip"):
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case commandExists("xsel"):
		cmd = exec.Command("xsel", "--clipboard", "--input")
	default:
		return fmt.Errorf("no clipboard utility found")
	}

	cmd.Stdin = strings.NewReader(content)
	return cmd.Run()
}

func commandExists(cmdName string) bool {
	_, err := exec.LookPath(cmdName)
	return err == nil
}