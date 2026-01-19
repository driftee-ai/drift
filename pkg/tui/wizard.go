package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type MainModel struct {
	session     *Session
	layout      LayoutModel
	currentPage Page
	pageIndex   int
	pages       []func(*Session, int, int) Page // Factory functions

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
		pages: []func(*Session, int, int) Page{
			func(s *Session, w, h int) Page { return NewDiscoveryPage(s, w, h) },
			func(s *Session, w, h int) Page { return NewProviderPage(s, w, h) },
			func(s *Session, w, h int) Page { return NewGroupingPage(s, w, h) },
			func(s *Session, w, h int) Page { return NewMappingPage(s, w, h) },
			func(s *Session, w, h int) Page { return NewFinalizePage(s, w, h) },
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

		if m.currentPage == nil && len(m.pages) > 0 {
			// Initialize first page
			m.currentPage = m.pages[0](m.session, m.width, m.height)
			return m, m.currentPage.Init()
		}

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

	case NextStepMsg:
		m.pageIndex++
		if m.pageIndex < len(m.pages) {
			m.currentPage = m.pages[m.pageIndex](m.session, m.width, m.height)
			return m, m.currentPage.Init()
		}
		return m, tea.Quit // Done
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

	return m.layout.RenderWithLayout(
		m.currentPage.Title(),
		m.currentPage.Step(),
		m.currentPage.Keys(),
		m.currentPage.View(),
	)
}
