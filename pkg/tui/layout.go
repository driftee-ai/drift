package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// LayoutModel wraps proper pages with a consistent header and footer.
type LayoutModel struct {
	width  int
	height int
	help   help.Model
}

func NewLayoutModel() LayoutModel {
	return LayoutModel{
		help: help.New(),
	}
}

func (l *LayoutModel) SetSize(w, h int) {
	l.width = w
	l.height = h
	l.help.Width = w
}

// RenderWithLayout wraps the content string with the standard header and footer.
// title: The main title of the current step
// step: "Step X of Y" or similar context
// keys: The keybindings active for this page
// content: The rendered string from the sub-model
func (l LayoutModel) RenderWithLayout(title, step string, keys []key.Binding, content string) string {
	header := l.renderHeader(title, step)
	footer := l.renderFooter(keys)

	// Calculate content height
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	contentHeight := l.height - headerHeight - footerHeight

	// Ensure content respects the available space
	// We use lipgloss.Place to align it nicely, but we don't want to force it if it's too big.
	styledContent := lipgloss.NewStyle().
		Width(l.width).
		Height(contentHeight).
		Render(content)

	return fmt.Sprintf("%s\n%s\n%s", header, styledContent, footer)
}

func (l LayoutModel) renderHeader(title, step string) string {
	titleStyle := HeaderStyle.Width(l.width / 2).Align(lipgloss.Left).Render(title)
	stepStyle := HeaderStyle.Width(l.width / 2).Align(lipgloss.Right).Render(step)

	return lipgloss.JoinHorizontal(lipgloss.Center, titleStyle, stepStyle)
}

func (l LayoutModel) renderFooter(keys []key.Binding) string {
	return l.help.ShortHelpView(keys)
}
