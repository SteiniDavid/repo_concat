package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
)

type sessionState int

const (
	configView sessionState = iota
	peekView
	fileBrowserView
	processingView
	resultsView
)

type Config struct {
	URL         string
	Path        string
	Include     []string
	Exclude     []string
	Output      string
	EnableTUI   bool
}

type FileItem struct {
	Path     string
	IsDir    bool
	Size     int64
	ModTime  time.Time
	Selected bool
}

func (f FileItem) Title() string       { return f.Path }
func (f FileItem) Description() string { 
	if f.IsDir {
		return "Directory"
	}
	return formatFileSize(f.Size)
}
func (f FileItem) FilterValue() string { return f.Path }

type Model struct {
	state           sessionState
	config          Config
	
	// Components
	urlInput        textinput.Model
	pathInput       textinput.Model
	includeInput    textinput.Model
	excludeInput    textinput.Model
	fileList        list.Model
	progressBar     progress.Model
	
	// Data
	files           []FileItem
	selectedFiles   map[string]bool
	currentDir      string
	
	// Peek data
	includedFiles   []string
	excludedFiles   []string
	directoryTree   string
	
	// UI State
	focused         int
	err             error
	processing      bool
	progress        float64
	statusMessage   string
	
	// Results
	totalFiles      int
	tokenCount      int
	outputFile      string
}

type progressMsg float64
type processingCompleteMsg struct {
	files      int
	tokens     int
	outputFile string
	err        error
}
type filesLoadedMsg []FileItem
type peekCompleteMsg struct {
	includedFiles []string
	excludedFiles []string
	directoryTree string
	err           error
}
type errorMsg error

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return "< 1KB"
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}