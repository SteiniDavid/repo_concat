package cli

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Minimal, elegant styling functions that enhance without overwhelming

var (
	// Simple, elegant colors
	blue   = color.New(color.FgBlue)
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
	cyan   = color.New(color.FgCyan)
	gray   = color.New(color.FgHiBlack)
	white  = color.New(color.FgWhite)
	bold   = color.New(color.Bold)
)

// Simple status messages with subtle icons
func StatusMsg(msgType, text string) string {
	switch msgType {
	case "success":
		return green.Sprint("✓") + " " + text
	case "error":
		return red.Sprint("✗") + " " + text
	case "warning":
		return yellow.Sprint("!") + " " + text
	case "info":
		return blue.Sprint("→") + " " + text
	case "loading":
		return blue.Sprint("•") + " " + text
	default:
		return text
	}
}

// Simple colored text
func Success(text string) string { return green.Sprint(text) }
func Error(text string) string   { return red.Sprint(text) }
func Warning(text string) string { return yellow.Sprint(text) }
func Info(text string) string    { return blue.Sprint(text) }
func Subtle(text string) string  { return gray.Sprint(text) }
func Highlight(text string) string { return cyan.Sprint(text) }

// Simple progress indicator (single line, no screen clearing)
func Progress(current, total int, message string) string {
	if total == 0 {
		return blue.Sprint("•") + " " + message
	}
	
	percentage := float64(current) / float64(total) * 100
	return fmt.Sprintf("%s %s (%.0f%% - %d/%d)", 
		blue.Sprint("•"), message, percentage, current, total)
}

// Elegant tree display with minimal borders
func SimpleTree(rootPath string, files []string, excluded []string) string {
	var lines []string
	
	// Just show a clean, simple tree with meaningful name
	lines = append(lines, Info("📁 " + rootPath))
	
	// Group files by directory
	dirs := make(map[string][]string)
	for _, file := range files {
		parts := strings.Split(file, "/")
		if len(parts) > 1 {
			dir := parts[0]
			filename := strings.Join(parts[1:], "/")
			dirs[dir] = append(dirs[dir], filename)
		} else {
			dirs["."] = append(dirs["."], file)
		}
	}
	
	// Show directories and files cleanly
	for dir, dirFiles := range dirs {
		if dir != "." {
			lines = append(lines, "  " + cyan.Sprint("📁 " + dir + "/"))
			for _, file := range dirFiles {
				icon := getSimpleIcon(file)
				lines = append(lines, "    " + gray.Sprint(icon + " " + file))
			}
		}
	}
	
	// Show root files
	if rootFiles, ok := dirs["."]; ok {
		for _, file := range rootFiles {
			icon := getSimpleIcon(file)
			lines = append(lines, "  " + white.Sprint(icon + " " + file))
		}
	}
	
	return strings.Join(lines, "\n")
}

func getSimpleIcon(filename string) string {
	if strings.HasSuffix(filename, ".go") {
		return "🔧"
	} else if strings.HasSuffix(filename, ".md") {
		return "📝"
	} else if strings.HasSuffix(filename, ".json") || strings.HasSuffix(filename, ".yml") {
		return "⚙️"
	}
	return "📄"
}

// Simple summary table (no heavy borders)
func SimpleSummary(included, excluded, totalSize int64) string {
	var lines []string
	
	lines = append(lines, bold.Sprint("Summary:"))
	lines = append(lines, fmt.Sprintf("  Files to include: %s", green.Sprint(fmt.Sprintf("%d", included))))
	lines = append(lines, fmt.Sprintf("  Files excluded:   %s", gray.Sprint(fmt.Sprintf("%d", excluded))))
	if totalSize > 0 {
		lines = append(lines, fmt.Sprintf("  Total size:       %s", gray.Sprint(formatSize(totalSize))))
	}
	
	return strings.Join(lines, "\n")
}

// Clean error display
func ErrorMsg(title, message, suggestion string) string {
	var lines []string
	
	lines = append(lines, red.Sprint("✗ " + title))
	lines = append(lines, "  " + message)
	if suggestion != "" {
		lines = append(lines, "  " + gray.Sprint("→ " + suggestion))
	}
	
	return strings.Join(lines, "\n")
}

// Simple confirmation prompt
func ConfirmPrompt(question string) string {
	return yellow.Sprint("? ") + question + " " + gray.Sprint("(y/N)")
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Simple header (just a colored title, no heavy borders)
func SimpleHeader(text string) string {
	return bold.Sprint(text)
}

// Completion message
func Done(outputPath string, fileCount int, tokenCount int) string {
	var lines []string
	
	lines = append(lines, green.Sprint("✓ Concatenation complete!"))
	lines = append(lines, fmt.Sprintf("  Output: %s", highlight.Sprint(outputPath)))
	lines = append(lines, fmt.Sprintf("  Files:  %d", fileCount))
	if tokenCount > 0 {
		lines = append(lines, fmt.Sprintf("  Tokens: ~%d", tokenCount))
	}
	
	return strings.Join(lines, "\n")
}

var highlight = cyan // alias for consistency