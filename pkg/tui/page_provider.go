package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type ProviderPage struct {
	list    list.Model
	session *Session
	width   int
	height  int
}

func NewProviderPage(s *Session, w, h int) *ProviderPage {
	l := list.New([]list.Item{
		ProviderItem{Name: "Gemini", Desc: "Google's multimodal AI model (Requires GEMINI_API_KEY)"},
		ProviderItem{Name: "OpenAI", Desc: "GPT-3.5/4 models (Requires OPENAI_API_KEY)"},
	}, list.NewDefaultDelegate(), w, h)

	l.Title = "Select AI Provider"
	l.SetStatusBarItemName("provider", "providers")

	return &ProviderPage{
		list:    l,
		session: s,
		width:   w,
		height:  h,
	}
}

func (p *ProviderPage) Init() tea.Cmd { return nil }

func (p *ProviderPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			if selected, ok := p.list.SelectedItem().(ProviderItem); ok {
				p.session.Provider = strings.ToLower(selected.Name)
				return p, func() tea.Msg { return NextStepMsg{} }
			}
		}
	case tea.WindowSizeMsg:
		p.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

func (p *ProviderPage) View() string  { return p.list.View() }
func (p *ProviderPage) Title() string { return "Provider Selection" }
func (p *ProviderPage) Step() string  { return "Step 2 of 4" }
func (p *ProviderPage) Keys() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	}
}
