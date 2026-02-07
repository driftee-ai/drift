package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type MainModel struct {
	session     *Session
	layout      LayoutModel
	currentPage Page
	pageIndex   int
	pages       []Page

	width    int
	height   int
	quitting bool
}

func NewModel() MainModel {
	layout := NewLayoutModel()
	session := &Session{}

	return MainModel{
		session: session,
		layout:  layout,
		pages: []Page{
			NewDiscoveryPage(session, 0, 0),
			NewProviderPage(session, 0, 0),
			NewGroupingPage(session, 0, 0),
			NewMappingPage(session, 0, 0),
			NewFinalizePage(session, 0, 0),
		},
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout.SetSize(msg.Width, msg.Height)

		m.layout.SetSize(msg.Width, msg.Height)

		// Propagate resize to all pages
		var cmds []tea.Cmd
		for i, page := range m.pages {
			var cmd tea.Cmd
			m.pages[i], cmd = page.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		if m.currentPage == nil && len(m.pages) > 0 {
			// Initialize first page
			m.currentPage = m.pages[0]
			cmds = append(cmds, m.currentPage.Init())
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		// Global Back handlers
		// Note: We should be careful not to override page-specific controls if they use these keys.
		// Most pages use 'esc' for cancelling internal states (editing, adding).
		// We can check if the current page handled the message or not, but Bubble Tea
		// doesn't support that easily (Update consummates the msg).
		// For now, let's bind 'ctrl+b' as a safe global back.
		if msg.String() == "ctrl+b" {
			return m.Update(PrevStepMsg{})
		}

	case NextStepMsg:
		m.pageIndex++
		if m.pageIndex < len(m.pages) {
			m.currentPage = m.pages[m.pageIndex]
			return m, m.currentPage.Init()
		}
		return m, tea.Quit // Done

	case PrevStepMsg:
		if m.pageIndex > 0 {
			m.pageIndex--
			m.currentPage = m.pages[m.pageIndex]
			// We might want to re-init or just let it be.
			// Calling Init() again might be useful for some pages to refresh data.
			return m, m.currentPage.Init()
		}
	}

	if m.currentPage != nil {
		var cmd tea.Cmd
		var newPage Page
		newPage, cmd = m.currentPage.Update(msg)
		m.currentPage = newPage
		return m, cmd
	}

	return m, nil
}

func (m MainModel) View() string {
	if m.quitting {
		return "Bye!\n"
	}
	if m.currentPage == nil {
		return "Initializing..."
	}

	keys := m.currentPage.Keys()
	if m.pageIndex > 0 {
		keys = append(keys, key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("ctrl+b", "back")))
	}
	keys = append(keys, key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")))

	return m.layout.RenderWithLayout(
		m.currentPage.Title(),
		m.currentPage.Step(),
		keys,
		m.currentPage.View(),
	)
}
