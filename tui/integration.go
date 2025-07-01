package tui

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// PerformDryRun performs a dry run to show what files would be processed (exported for testing)
func PerformDryRun(rootPath string, exclusionPatterns []string, inclusionPatterns []string) ([]string, []string, error) {
	return performDryRun(rootPath, exclusionPatterns, inclusionPatterns)
}

// performDryRun performs a dry run to show what files would be processed
func performDryRun(rootPath string, exclusionPatterns []string, inclusionPatterns []string) ([]string, []string, error) {
	// Validate exclusion patterns
	var validExclusionPatterns []string
	for _, pattern := range exclusionPatterns {
		if !isPathPattern(pattern) {
			// Test if it's a valid regex or can be converted from glob
			testPattern := pattern
			if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
				testPattern = globToRegex(pattern)
			}
			if _, err := regexp.Compile(testPattern); err != nil {
				return nil, nil, fmt.Errorf("invalid exclusion pattern '%s': %v", pattern, err)
			}
		}
		validExclusionPatterns = append(validExclusionPatterns, pattern)
	}

	// Validate inclusion patterns
	var validInclusionPatterns []string
	for _, pattern := range inclusionPatterns {
		if !isPathPattern(pattern) {
			// Test if it's a valid regex or can be converted from glob
			testPattern := pattern
			if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
				testPattern = globToRegex(pattern)
			}
			if _, err := regexp.Compile(testPattern); err != nil {
				return nil, nil, fmt.Errorf("invalid inclusion pattern '%s': %v", pattern, err)
			}
		}
		validInclusionPatterns = append(validInclusionPatterns, pattern)
	}

	// Default exclusion patterns
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

	// Combine with user exclusions
	allExclusionPatterns := append(defaultExclusionPatterns, validExclusionPatterns...)

	var includedFiles []string
	var excludedFiles []string

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !isTextFile(path) {
			excludedFiles = append(excludedFiles, path)
			return nil
		}

		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		baseName := filepath.Base(path)

		// Check exclusion patterns
		for _, pattern := range allExclusionPatterns {
			if matchesPattern(pattern, relativePath, baseName) {
				excludedFiles = append(excludedFiles, path)
				return nil
			}
		}

		// Check inclusion patterns (if any)
		if len(validInclusionPatterns) > 0 {
			included := false
			for _, pattern := range validInclusionPatterns {
				if matchesPattern(pattern, relativePath, baseName) {
					included = true
					break
				}
			}
			if !included {
				excludedFiles = append(excludedFiles, path)
				return nil
			}
		}

		includedFiles = append(includedFiles, path)
		return nil
	})

	return includedFiles, excludedFiles, err
}

// isPathPattern determines if a pattern is a path-based pattern
func isPathPattern(pattern string) bool {
	return strings.HasPrefix(pattern, "/")
}

// matchesPattern checks if a file matches a given pattern
func matchesPattern(pattern, relativePath, baseName string) bool {
	if isPathPattern(pattern) {
		return matchesPathPattern(pattern, relativePath)
	}

	// Convert glob patterns to regex if needed
	regexPattern := pattern
	if strings.Contains(pattern, "*") || strings.Contains(pattern, "?") {
		regexPattern = globToRegex(pattern)
	}

	// Try regex matching on both relative path and base name
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return false
	}

	return regex.MatchString(relativePath) || regex.MatchString(baseName)
}

// globToRegex converts a glob pattern to a regex pattern
func globToRegex(glob string) string {
	// Escape regex special characters except * and ?
	result := regexp.QuoteMeta(glob)
	
	// Replace escaped glob characters with regex equivalents
	result = strings.ReplaceAll(result, "\\*", ".*")
	result = strings.ReplaceAll(result, "\\?", ".")
	
	// Anchor the pattern
	if !strings.HasPrefix(result, ".*") {
		result = "^" + result
	}
	if !strings.HasSuffix(result, ".*") {
		result = result + "$"
	}
	
	return result
}

// matchesPathPattern handles path-based pattern matching
func matchesPathPattern(pattern, relativePath string) bool {
	pattern = strings.TrimPrefix(pattern, "/")
	if strings.HasSuffix(pattern, "/") {
		// Directory pattern
		pattern = strings.TrimSuffix(pattern, "/")
		return strings.HasPrefix(relativePath, pattern+"/") || relativePath == pattern
	}
	// File pattern
	return strings.HasPrefix(relativePath, pattern)
}

// isTextFile determines if a file is likely a text file
func isTextFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first 512 bytes to check for binary content
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return false
	}

	// Check for null bytes (common in binary files)
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return false
		}
	}

	return true
}

// CacheEntry represents cached repository metadata
type CacheEntry struct {
	URL        string    `json:"url"`
	CachedAt   time.Time `json:"cached_at"`
	RepoPath   string    `json:"repo_path"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// getTmpCacheDir returns the cache directory path
func getTmpCacheDir() string {
	return filepath.Join("/tmp", "repo-concat-cache")
}

// urlToHash converts a URL to a hash for cache identification
func urlToHash(githubURL string) string {
	hash := md5.Sum([]byte(githubURL))
	return hex.EncodeToString(hash[:])
}

// getCachedRepo checks if a repository is already cached and valid
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

// cacheRepo stores repository information in cache
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

// cloneRepository clones a GitHub repository to a destination directory
func cloneRepository(githubURL, destDir string) error {
	cmd := exec.Command("git", "clone", githubURL)
	cmd.Dir = destDir
	// Don't pipe stdout/stderr to avoid issues in TUI mode
	return cmd.Run()
}

// extractRepoName extracts repository name from GitHub URL
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

// resolveRepositoryPath resolves either local path or GitHub URL to a local path
func resolveRepositoryPath(config Config) (string, error) {
	if config.Path != "" {
		return config.Path, nil
	}
	
	if config.URL != "" {
		// Check cache first
		if cachedPath, found, _, err := getCachedRepo(config.URL); err != nil {
			return "", fmt.Errorf("cache check failed: %v", err)
		} else if found {
			return cachedPath, nil
		}

		// Need to clone the repository
		repoName := extractRepoName(config.URL)
		tempDir := getTmpCacheDir()
		repoPath := filepath.Join(tempDir, repoName)

		// Create temp directory
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create temp directory: %v", err)
		}

		// Remove existing directory if it exists
		os.RemoveAll(repoPath)

		// Clone repository
		if err := cloneRepository(config.URL, tempDir); err != nil {
			return "", fmt.Errorf("failed to clone repository: %v", err)
		}

		// Cache the cloned repository
		if err := cacheRepo(config.URL, repoPath); err != nil {
			return "", fmt.Errorf("failed to cache repository: %v", err)
		}

		return repoPath, nil
	}

	return "", fmt.Errorf("please specify either a repository URL or local path")
}

// processRepositoryTUI handles the full repository processing for TUI
func processRepositoryTUI(config Config, statusCallback func(string), progressCallback func(float64)) (int, int, string, error) {
	statusCallback("Resolving repository...")
	progressCallback(0.05)
	
	// Resolve repository path (local or GitHub URL)
	rootPath, err := resolveRepositoryPath(config)
	if err != nil {
		return 0, 0, "", err
	}

	statusCallback("Collecting files...")
	progressCallback(0.1)

	// Collect files using the same logic as the CLI
	files, err := collectFiles(rootPath, config.Exclude, config.Include)
	if err != nil {
		return 0, 0, "", fmt.Errorf("Failed to collect files: %v", err)
	}

	statusCallback(fmt.Sprintf("Processing %d files...", len(files)))
	progressCallback(0.3)

	// Concatenate files
	content, err := concatenateFiles(files, rootPath)
	if err != nil {
		return 0, 0, "", fmt.Errorf("Failed to concatenate files: %v", err)
	}

	statusCallback("Generating output...")
	progressCallback(0.8)

	// Generate output file
	timestamp := time.Now().Format("20060102_150405")
	outputFileName := fmt.Sprintf("repo_concat_%s.txt", timestamp)
	outputPath := filepath.Join(config.Output, "repo-concat-output", outputFileName)

	// Create output directory
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return 0, 0, "", fmt.Errorf("Failed to create output directory: %v", err)
	}

	// Write output file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return 0, 0, "", fmt.Errorf("Failed to write output file: %v", err)
	}

	// Estimate token count (rough approximation: 1 token ≈ 4 characters)
	tokenCount := len(content) / 4

	statusCallback("Complete!")
	progressCallback(1.0)

	return len(files), tokenCount, outputPath, nil
}

// collectFiles collects all files that should be processed
func collectFiles(rootPath string, exclusionPatterns []string, inclusionPatterns []string) ([]string, error) {
	includedFiles, _, err := performDryRun(rootPath, exclusionPatterns, inclusionPatterns)
	return includedFiles, err
}

// concatenateFiles concatenates all files with headers
func concatenateFiles(files []string, rootPath string) (string, error) {
	var result strings.Builder
	
	// Add header
	result.WriteString("# Repository Concatenation\n")
	result.WriteString(fmt.Sprintf("# Generated on: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("# Total files: %d\n\n", len(files)))

	for _, filePath := range files {
		relativePath, err := filepath.Rel(rootPath, filePath)
		if err != nil {
			relativePath = filePath
		}

		result.WriteString(fmt.Sprintf("# File: %s\n", relativePath))
		result.WriteString("```\n")

		// Read file content
		file, err := os.Open(filePath)
		if err != nil {
			result.WriteString(fmt.Sprintf("Error reading file: %v\n", err))
		} else {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				result.WriteString(scanner.Text())
				result.WriteString("\n")
			}
			file.Close()
		}

		result.WriteString("```\n\n")
	}

	return result.String(), nil
}