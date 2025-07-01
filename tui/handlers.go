package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateConfigView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle escape and quit keys first, before any component processing
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	}

	// Handle navigation keys
	switch msg.String() {
	case "tab", "down":
		m.focused = (m.focused + 1) % 7
		m = m.updateFocus()
		return m, nil

	case "shift+tab", "up":
		m.focused = (m.focused - 1 + 7) % 7
		m = m.updateFocus()
		return m, nil

	case "enter":
		// Update config from form inputs
		m = m.updateConfigFromInputs()
		
		switch m.focused {
		case 4: // Peek/Preview
			m.state = peekView
			return m, m.startPeek()
		case 5: // Browse Files
			m.state = fileBrowserView
			return m, m.loadFiles()
		case 6: // Process Now
			m.state = processingView
			m.processing = true
			return m, m.startProcessing()
		}
		return m, nil
	}

	// Update active input only if not a navigation/control key
	switch m.focused {
	case 0:
		m.urlInput, cmd = m.urlInput.Update(msg)
	case 1:
		m.pathInput, cmd = m.pathInput.Update(msg)
	case 2:
		m.includeInput, cmd = m.includeInput.Update(msg)
	case 3:
		m.excludeInput, cmd = m.excludeInput.Update(msg)
	}

	return m, cmd
}

func (m Model) updatePeekView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit

	case "b":
		m.state = configView
		return m, nil

	case "enter":
		// Proceed with processing
		m.state = processingView
		m.processing = true
		return m, m.startProcessing()
	}

	return m, nil
}

func (m Model) updateFileBrowserView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle escape and quit keys first, before any component processing
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	}

	// Handle our custom navigation keys
	switch msg.String() {
	case "b":
		m.state = configView
		return m, nil

	case "enter":
		m.state = processingView
		m.processing = true
		return m, m.startProcessing()

	case " ":
		// Toggle selection for current item
		if selectedItem, ok := m.fileList.SelectedItem().(FileItem); ok {
			if m.selectedFiles[selectedItem.Path] {
				delete(m.selectedFiles, selectedItem.Path)
			} else {
				m.selectedFiles[selectedItem.Path] = true
			}
		}
		return m, nil
	}

	// Update file list only for other keys (arrow keys, etc.)
	m.fileList, cmd = m.fileList.Update(msg)
	return m, cmd
}

func (m Model) updateProcessingView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Allow escape to quit even during processing
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	}

	// Ignore other keys during processing
	return m, nil
}

func (m Model) updateResultsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit

	case "enter":
		m.state = configView
		m.err = nil
		m.processing = false
		m.progress = 0
		m.selectedFiles = make(map[string]bool)
		return m, nil
	}

	return m, nil
}

func (m Model) updateFocus() Model {
	// Reset all focus states
	m.urlInput.Blur()
	m.pathInput.Blur()
	m.includeInput.Blur()
	m.excludeInput.Blur()

	// Set focus on current field
	switch m.focused {
	case 0:
		m.urlInput.Focus()
	case 1:
		m.pathInput.Focus()
	case 2:
		m.includeInput.Focus()
	case 3:
		m.excludeInput.Focus()
	}

	return m
}

func (m Model) updateConfigFromInputs() Model {
	// Update config from input fields
	m.config.URL = strings.TrimSpace(m.urlInput.Value())
	m.config.Path = strings.TrimSpace(m.pathInput.Value())
	
	// Parse include patterns
	includeStr := strings.TrimSpace(m.includeInput.Value())
	if includeStr != "" {
		m.config.Include = strings.Split(includeStr, ",")
		for i := range m.config.Include {
			m.config.Include[i] = strings.TrimSpace(m.config.Include[i])
		}
	} else {
		m.config.Include = []string{}
	}
	
	// Parse exclude patterns
	excludeStr := strings.TrimSpace(m.excludeInput.Value())
	if excludeStr != "" {
		m.config.Exclude = strings.Split(excludeStr, ",")
		for i := range m.config.Exclude {
			m.config.Exclude[i] = strings.TrimSpace(m.config.Exclude[i])
		}
	} else {
		m.config.Exclude = []string{}
	}
	
	return m
}

func (m Model) startPeek() tea.Cmd {
	return func() tea.Msg {
		// Resolve repository path (local or GitHub URL)
		rootPath, err := resolveRepositoryPath(m.config)
		if err != nil {
			return peekCompleteMsg{err: err}
		}

		// Perform dry run to get files that would be included/excluded
		includedFiles, excludedFiles, err := performDryRun(rootPath, m.config.Exclude, m.config.Include)
		if err != nil {
			return peekCompleteMsg{err: fmt.Errorf("Failed to scan files: %v", err)}
		}

		// Convert to relative paths for display
		var relIncluded []string
		var relExcluded []string

		for _, filePath := range includedFiles {
			if relPath, err := filepath.Rel(rootPath, filePath); err == nil {
				relIncluded = append(relIncluded, relPath)
			} else {
				relIncluded = append(relIncluded, filePath)
			}
		}

		for _, filePath := range excludedFiles {
			if relPath, err := filepath.Rel(rootPath, filePath); err == nil {
				relExcluded = append(relExcluded, relPath)
			} else {
				relExcluded = append(relExcluded, filePath)
			}
		}

		return peekCompleteMsg{
			includedFiles: relIncluded,
			excludedFiles: relExcluded,
			directoryTree: "", // Could add directory tree later
			err:           nil,
		}
	}
}

func (m Model) loadFiles() tea.Cmd {
	return func() tea.Msg {
		// Resolve repository path (local or GitHub URL)
		rootPath, err := resolveRepositoryPath(m.config)
		if err != nil {
			return errorMsg(err)
		}

		// Perform dry run to get files that would be included/excluded
		includedFiles, excludedFiles, err := performDryRun(rootPath, m.config.Exclude, m.config.Include)
		if err != nil {
			return errorMsg(fmt.Errorf("Failed to scan files: %v", err))
		}

		var files []FileItem
		
		// Add included files
		for _, filePath := range includedFiles {
			info, err := os.Stat(filePath)
			if err != nil {
				continue // Skip files we can't stat
			}
			
			relPath, _ := filepath.Rel(rootPath, filePath)
			files = append(files, FileItem{
				Path:     relPath,
				IsDir:    info.IsDir(),
				Size:     info.Size(),
				ModTime:  info.ModTime(),
				Selected: false,
			})
		}

		// Add some excluded files for context (marked as excluded)
		for i, filePath := range excludedFiles {
			if i >= 10 { // Limit to first 10 excluded files
				break
			}
			info, err := os.Stat(filePath)
			if err != nil {
				continue
			}
			
			relPath, _ := filepath.Rel(rootPath, filePath)
			files = append(files, FileItem{
				Path:     fmt.Sprintf("[EXCLUDED] %s", relPath),
				IsDir:    info.IsDir(),
				Size:     info.Size(),
				ModTime:  info.ModTime(),
				Selected: false,
			})
		}

		return filesLoadedMsg(files)
	}
}

func (m Model) startProcessing() tea.Cmd {
	return func() tea.Msg {
		statusCallback := func(status string) {
			// TODO: Could send status updates via channels in future
		}

		progressCallback := func(progress float64) {
			// TODO: Could send progress updates via channels in future  
		}

		// Process the repository using actual logic
		files, tokens, outputFile, err := processRepositoryTUI(
			m.config,
			statusCallback,
			progressCallback,
		)

		return processingCompleteMsg{
			files:      files,
			tokens:     tokens,
			outputFile: outputFile,
			err:        err,
		}
	}
}