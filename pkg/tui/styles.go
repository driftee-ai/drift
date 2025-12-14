package tui

import "github.com/charmbracelet/lipgloss"

var (
	// General styles
	AppStyle  = lipgloss.NewStyle().Padding(1, 2)
	HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(0, 1).
			Bold(true)

	// List styles
	SelectedItemStyle   = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	UnselectedItemStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("252"))

	// Status messages
	StatusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("57"))

	// Highlight color for important text or actions
	HighlightColor = lipgloss.Color("205") // Pink
	GreenColor     = lipgloss.Color("10")
	RedColor       = lipgloss.Color("9")

	// Buttons
	ButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("237")).
			Padding(0, 2).
			Margin(1, 1).
			Height(2).
			Align(lipgloss.Center)

	ActiveButtonStyle = ButtonStyle.
				Foreground(lipgloss.Color("236")).
				Background(HighlightColor)
)
