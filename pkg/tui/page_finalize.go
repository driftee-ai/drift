package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/driftee-ai/drift/pkg/config"
	"gopkg.in/yaml.v3"
)

type FinalizePage struct {
	spinner  spinner.Model
	session  *Session
	width    int
	height   int
	isSaving bool
	isDone   bool
	err      error
}

func NewFinalizePage(s *Session, w, h int) *FinalizePage {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(HighlightColor)

	return &FinalizePage{
		spinner:  sp,
		session:  s,
		width:    w,
		height:   h,
		isSaving: true,
	}
}

func (p *FinalizePage) Init() tea.Cmd {
	return tea.Batch(
		p.spinner.Tick,
		p.saveConfigCmd(),
	)
}

func (p *FinalizePage) Update(msg tea.Msg) (Page, tea.Cmd) {
	switch msg := msg.(type) {
	case configSavedMsg:
		p.isSaving = false
		p.isDone = true
		return p, nil

	case errMsg:
		p.isSaving = false
		p.err = msg.error
		return p, nil

	case tea.KeyMsg:
		if p.isDone || p.err != nil {
			if msg.String() == "enter" || msg.String() == "q" {
				return p, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	p.spinner, cmd = p.spinner.Update(msg)
	return p, cmd
}

func (p *FinalizePage) View() string {
	content := ""
	if p.isSaving {
		content = fmt.Sprintf("%s Saving configuration...", p.spinner.View())
	} else if p.err != nil {
		content = fmt.Sprintf("Error saving config: %v\nPress Enter to exit.", p.err)
	} else {
		content = "Configuration saved to .drift.yaml!\nPress Enter to exit."
	}

	return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center, content)
}

func (p *FinalizePage) Title() string { return "Finalizing" }
func (p *FinalizePage) Step() string  { return "Step 5 of 5" }

func (p *FinalizePage) Keys() []key.Binding {
	if p.isDone || p.err != nil {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "exit")),
		}
	}
	return []key.Binding{}
}

// Logic

type configSavedMsg struct{}

func (p *FinalizePage) saveConfigCmd() tea.Cmd {
	return func() tea.Msg {
		conf := config.Config{
			Version:  1,
			Provider: p.session.Provider,
			Rules:    p.session.Groups,
		}

		data, err := yaml.Marshal(conf)
		if err != nil {
			return errMsg{err}
		}

		// Add comments
		commentedData := []byte("# .drift.yaml\n# This file defines the rules for checking drift between your code and documentation.\n\n" + string(data))

		err = os.WriteFile(".drift.yaml", commentedData, 0644)
		if err != nil {
			return errMsg{err}
		}

		return configSavedMsg{}
	}
}
