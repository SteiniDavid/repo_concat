package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func NewModel(config Config) Model {
	// Initialize input fields
	urlInput := textinput.New()
	urlInput.Placeholder = "https://github.com/user/repo"
	urlInput.Focus()
	urlInput.CharLimit = 200
	urlInput.Width = 50
	if config.URL != "" {
		urlInput.SetValue(config.URL)
	}

	pathInput := textinput.New()
	pathInput.Placeholder = "/path/to/local/repo"
	pathInput.CharLimit = 200
	pathInput.Width = 50
	if config.Path != "" {
		pathInput.SetValue(config.Path)
	}

	includeInput := textinput.New()
	includeInput.Placeholder = "*.go,*.md (comma separated)"
	includeInput.CharLimit = 200
	includeInput.Width = 50
	if len(config.Include) > 0 {
		includeInput.SetValue(strings.Join(config.Include, ","))
	}

	excludeInput := textinput.New()
	excludeInput.Placeholder = "*.test.go,vendor/ (comma separated)"
	excludeInput.CharLimit = 200
	excludeInput.Width = 50
	if len(config.Exclude) > 0 {
		excludeInput.SetValue(strings.Join(config.Exclude, ","))
	}

	// Initialize file list
	fileList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	fileList.Title = "Repository Files"
	fileList.SetShowStatusBar(false)
	fileList.SetFilteringEnabled(true)
	fileList.Styles.Title = HeaderStyle

	// Initialize progress bar
	progressBar := progress.New(progress.WithDefaultGradient())

	return Model{
		state:         configView,
		config:        config,
		urlInput:      urlInput,
		pathInput:     pathInput,
		includeInput:  includeInput,
		excludeInput:  excludeInput,
		fileList:      fileList,
		progressBar:   progressBar,
		selectedFiles: make(map[string]bool),
		focused:       0,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case configView:
			return m.updateConfigView(msg)
		case peekView:
			return m.updatePeekView(msg)
		case fileBrowserView:
			return m.updateFileBrowserView(msg)
		case processingView:
			return m.updateProcessingView(msg)
		case resultsView:
			return m.updateResultsView(msg)
		}

	case tea.WindowSizeMsg:
		m.fileList.SetWidth(msg.Width - 4)
		m.fileList.SetHeight(msg.Height - 10)
		return m, nil

	case progressMsg:
		m.progress = float64(msg)
		cmd = m.progressBar.SetPercent(m.progress)
		return m, cmd

	case processingCompleteMsg:
		m.processing = false
		m.totalFiles = msg.files
		m.tokenCount = msg.tokens
		m.outputFile = msg.outputFile
		m.err = msg.err
		m.state = resultsView
		return m, nil

	case filesLoadedMsg:
		m.files = msg
		items := make([]list.Item, len(m.files))
		for i, file := range m.files {
			items[i] = file
		}
		m.fileList.SetItems(items)
		return m, nil

	case peekCompleteMsg:
		m.includedFiles = msg.includedFiles
		m.excludedFiles = msg.excludedFiles
		m.directoryTree = msg.directoryTree
		m.err = msg.err
		return m, nil

	case errorMsg:
		m.err = msg
		return m, nil
	}

	// Note: Component updates are handled in individual view handlers
	// to avoid double processing and escape key interference

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.state {
	case configView:
		return m.configViewRender()
	case peekView:
		return m.peekViewRender()
	case fileBrowserView:
		return m.fileBrowserViewRender()
	case processingView:
		return m.processingViewRender()
	case resultsView:
		return m.resultsViewRender()
	}
	return ""
}

func (m Model) configViewRender() string {
	var b strings.Builder

	b.WriteString(RenderTitle("Repo Concat Configuration"))
	b.WriteString("\n")

	// Create input rows with proper alignment
	inputs := []struct {
		label string
		input textinput.Model
		focused bool
	}{
		{"URL", m.urlInput, m.focused == 0},
		{"Path", m.pathInput, m.focused == 1},
		{"Include", m.includeInput, m.focused == 2},
		{"Exclude", m.excludeInput, m.focused == 3},
	}

	for _, input := range inputs {
		row := lipgloss.JoinHorizontal(lipgloss.Left,
			RenderLabel(input.label+":"),
			func() string {
				if input.focused {
					return RenderInput(input.input.View(), true)
				}
				return RenderInput(input.input.View(), false)
			}(),
		)
		b.WriteString(row)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	
	// Buttons
	b.WriteString(RenderButton("Peek", m.focused == 4))
	b.WriteString(RenderButton("Browse Files", m.focused == 5))
	b.WriteString(RenderButton("Process Now", m.focused == 6))
	b.WriteString("\n\n")

	b.WriteString(RenderHelp("↑/↓: Navigate • Tab: Next field • Enter: Select • Esc: Exit"))

	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(RenderError(fmt.Sprintf("Error: %v", m.err)))
	}

	return BaseStyle.Render(b.String())
}

func (m Model) peekViewRender() string {
	var b strings.Builder

	b.WriteString(RenderTitle("Repository Preview"))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(RenderError(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
		b.WriteString(RenderHelp("b: Back • Esc: Exit"))
		return BaseStyle.Render(b.String())
	}

	// Show statistics
	b.WriteString(RenderSuccess(fmt.Sprintf("Found %d files to include", len(m.includedFiles))))
	b.WriteString("\n")
	b.WriteString(RenderWarning(fmt.Sprintf("Found %d files to exclude", len(m.excludedFiles))))
	b.WriteString("\n\n")

	// Show sample included files
	b.WriteString(RenderHeader("Files to Include:"))
	b.WriteString("\n")
	for i, file := range m.includedFiles {
		if i >= 10 { // Show only first 10
			b.WriteString("  ... and ")
			b.WriteString(fmt.Sprintf("%d more files", len(m.includedFiles)-10))
			b.WriteString("\n")
			break
		}
		b.WriteString("  ")
		b.WriteString(file)
		b.WriteString("\n")
	}

	if len(m.excludedFiles) > 0 {
		b.WriteString("\n")
		b.WriteString(RenderHeader("Sample Excluded Files:"))
		b.WriteString("\n")
		for i, file := range m.excludedFiles {
			if i >= 5 { // Show only first 5 excluded
				b.WriteString("  ... and ")
				b.WriteString(fmt.Sprintf("%d more excluded files", len(m.excludedFiles)-5))
				b.WriteString("\n")
				break
			}
			b.WriteString("  ")
			b.WriteString(RenderStatus(file))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n")
	b.WriteString(RenderHelp("Enter: Process • b: Back • Esc: Exit"))

	return BaseStyle.Render(b.String())
}

func (m Model) fileBrowserViewRender() string {
	var b strings.Builder

	b.WriteString(RenderTitle("File Browser"))
	b.WriteString("\n")

	b.WriteString(ListStyle.Render(m.fileList.View()))
	b.WriteString("\n")

	selectedCount := len(m.selectedFiles)
	b.WriteString(RenderStatus(fmt.Sprintf("Selected: %d files", selectedCount)))
	b.WriteString("\n\n")

	b.WriteString(RenderHelp("↑/↓: Navigate • Space: Toggle selection • Enter: Process • b: Back • Esc: Exit"))

	return BaseStyle.Render(b.String())
}

func (m Model) processingViewRender() string {
	var b strings.Builder

	b.WriteString(RenderTitle("Processing Repository"))
	b.WriteString("\n")

	b.WriteString(ProgressStyle.Render(m.progressBar.View()))
	b.WriteString("\n")

	b.WriteString(RenderStatus(m.statusMessage))
	b.WriteString("\n\n")

	b.WriteString(RenderHelp("Please wait while processing..."))

	return BaseStyle.Render(b.String())
}

func (m Model) resultsViewRender() string {
	var b strings.Builder

	if m.err != nil {
		b.WriteString(RenderTitle("Processing Failed"))
		b.WriteString("\n")
		b.WriteString(RenderError(fmt.Sprintf("Error: %v", m.err)))
	} else {
		b.WriteString(RenderTitle("Processing Complete"))
		b.WriteString("\n")
		b.WriteString(RenderSuccess(fmt.Sprintf("Successfully processed %d files", m.totalFiles)))
		b.WriteString("\n")
		b.WriteString(RenderSuccess(fmt.Sprintf("Estimated tokens: %d", m.tokenCount)))
		b.WriteString("\n")
		b.WriteString(RenderSuccess(fmt.Sprintf("Output saved to: %s", m.outputFile)))
	}

	b.WriteString("\n\n")
	b.WriteString(RenderHelp("Enter: Process another • Esc: Exit"))

	return BaseStyle.Render(b.String())
}

func RunTUI(config Config) error {
	m := NewModel(config)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}