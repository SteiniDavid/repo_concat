package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Color palette - softer, more readable colors
	primaryColor   = lipgloss.Color("#60A5FA") // Soft blue
	secondaryColor = lipgloss.Color("#34D399") // Soft green  
	successColor   = lipgloss.Color("#22C55E") // Green
	warningColor   = lipgloss.Color("#FBBF24") // Yellow
	errorColor     = lipgloss.Color("#F87171") // Soft red
	textColor      = lipgloss.Color("#E5E7EB") // Light gray
	mutedColor     = lipgloss.Color("#9CA3AF") // Medium gray
	bgColor        = lipgloss.Color("#1F2937") // Dark background
	
	// Base styles - clean and minimal
	BaseStyle = lipgloss.NewStyle().
		Padding(1, 2)
	
	// Header styles - clean and readable
	HeaderStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)
	
	TitleStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Padding(1, 0).
		MarginBottom(1).
		Align(lipgloss.Left)
	
	// Input styles - subtle and clean
	InputStyle = lipgloss.NewStyle().
		BorderForeground(mutedColor).
		Padding(0, 1).
		Width(50).
		Foreground(textColor)
	
	InputFocusedStyle = InputStyle.Copy().
		BorderForeground(primaryColor).
		Foreground(textColor)
	
	LabelStyle = lipgloss.NewStyle().
		Foreground(textColor).
		Bold(false).
		Width(12).
		Align(lipgloss.Right).
		MarginRight(1)
	
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
		Bold(false)
	
	FileStyle = lipgloss.NewStyle().
		Foreground(textColor)
	
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
	
	// Button styles - simple and clean
	ButtonStyle = lipgloss.NewStyle().
		Foreground(textColor).
		Background(mutedColor).
		Padding(0, 3).
		MarginRight(2)
	
	ButtonFocusedStyle = ButtonStyle.Copy().
		Foreground(textColor).
		Background(primaryColor)
	
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