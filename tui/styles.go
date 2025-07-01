package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Enhanced color palette - vibrant but professional
	primaryColor   = lipgloss.Color("#3B82F6") // Bright blue
	secondaryColor = lipgloss.Color("#10B981") // Emerald green  
	accentColor    = lipgloss.Color("#8B5CF6") // Purple
	successColor   = lipgloss.Color("#059669") // Dark green
	warningColor   = lipgloss.Color("#F59E0B") // Amber
	errorColor     = lipgloss.Color("#EF4444") // Red
	infoColor      = lipgloss.Color("#06B6D4") // Cyan
	textColor      = lipgloss.Color("#F9FAFB") // Almost white
	mutedColor     = lipgloss.Color("#6B7280") // Medium gray
	bgColor        = lipgloss.Color("#111827") // Dark blue-gray
	highlightColor = lipgloss.Color("#FBBF24") // Golden yellow
	
	// Base styles - clean and minimal
	BaseStyle = lipgloss.NewStyle().
		Padding(1, 2)
	
	// Header styles - enhanced with gradients and borders
	HeaderStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Padding(0, 2).
		MarginBottom(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor)
	
	TitleStyle = lipgloss.NewStyle().
		Foreground(textColor).
		Background(primaryColor).
		Bold(true).
		Padding(1, 3).
		MarginBottom(1).
		Align(lipgloss.Center).
		Border(lipgloss.ThickBorder()).
		BorderForeground(accentColor)
	
	// Input styles - enhanced with glow effects
	InputStyle = lipgloss.NewStyle().
		BorderForeground(mutedColor).
		Padding(0, 2).
		Width(48).
		Foreground(textColor).
		Border(lipgloss.RoundedBorder())
	
	InputFocusedStyle = InputStyle.Copy().
		BorderForeground(highlightColor).
		Foreground(textColor).
		Bold(true)
	
	LabelStyle = lipgloss.NewStyle().
		Foreground(textColor).
		Bold(false).
		Width(15).
		Align(lipgloss.Right).
		MarginRight(2).
		Padding(1, 0)
	
	// List styles - minimal borders
	ListStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(mutedColor).
		Padding(1).
		Height(20)
	
	SelectedItemStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		PaddingLeft(2)
	
	ItemStyle = lipgloss.NewStyle().
		Foreground(textColor).
		PaddingLeft(2)
	
	DirectoryStyle = lipgloss.NewStyle().
		Foreground(secondaryColor).
		Bold(true).
		Italic(false)
	
	FileStyle = lipgloss.NewStyle().
		Foreground(textColor)
	
	HighlightedFileStyle = lipgloss.NewStyle().
		Foreground(highlightColor).
		Bold(true)
	
	// Status styles
	StatusStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Padding(0, 1)
	
	SuccessStyle = lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true)
	
	ErrorStyle = lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true)
	
	WarningStyle = lipgloss.NewStyle().
		Foreground(warningColor).
		Bold(true)
	
	// Progress styles - clean
	ProgressStyle = lipgloss.NewStyle().
		Padding(0, 1).
		MarginBottom(1)
	
	// Button styles - enhanced with gradients and shadows
	ButtonStyle = lipgloss.NewStyle().
		Foreground(textColor).
		Background(mutedColor).
		Padding(0, 4).
		MarginRight(2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(mutedColor)
	
	ButtonFocusedStyle = ButtonStyle.Copy().
		Foreground(textColor).
		Background(primaryColor).
		BorderForeground(primaryColor).
		Bold(true)
	
	ButtonActiveStyle = ButtonStyle.Copy().
		Foreground(textColor).
		Background(successColor).
		BorderForeground(successColor).
		Bold(true)
	
	// Help styles
	HelpStyle = lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Padding(1, 0)
)

func RenderTitle(text string) string {
	return TitleStyle.Render(text)
}

func RenderHeader(text string) string {
	return HeaderStyle.Render(text)
}

func RenderSuccess(text string) string {
	return SuccessStyle.Render(text)
}

func RenderError(text string) string {
	return ErrorStyle.Render(text)
}

func RenderWarning(text string) string {
	return WarningStyle.Render(text)
}

func RenderStatus(text string) string {
	return StatusStyle.Render(text)
}

func RenderButton(text string, focused bool) string {
	if focused {
		return ButtonFocusedStyle.Render(text)
	}
	return ButtonStyle.Render(text)
}

func RenderInput(text string, focused bool) string {
	if focused {
		return InputFocusedStyle.Border(lipgloss.NormalBorder()).Render(text)
	}
	return InputStyle.Border(lipgloss.NormalBorder()).Render(text)
}

func RenderLabel(text string) string {
	return LabelStyle.Render(text)
}

func RenderHelp(text string) string {
	return HelpStyle.Render(text)
}

func RenderHighlight(text string) string {
	return lipgloss.NewStyle().
		Foreground(highlightColor).
		Bold(true).
		Render(text)
}

func RenderInfo(text string) string {
	return lipgloss.NewStyle().
		Foreground(infoColor).
		Bold(true).
		Render(text)
}

func RenderAccent(text string) string {
	return lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		Render(text)
}

func RenderGradientText(text string) string {
	// Simple gradient effect using different shades
	return lipgloss.NewStyle().
		Foreground(primaryColor).
		Background(lipgloss.Color("#1E3A8A")).
		Bold(true).
		Padding(0, 1).
		Render(text)
}