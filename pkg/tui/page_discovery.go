package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DiscoveryPage struct {
	list        list.Model
	textInput   textinput.Model
	session     *Session
	width       int
	height      int
	isAdding    bool
	feedbackMsg string
}

func NewDiscoveryPage(s *Session, w, h int) *DiscoveryPage {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), w, h)
	l.Title = "Review Documentation Files"
	l.SetStatusBarItemName("file", "files")

	ti := textinput.New()
	ti.Placeholder = "Enter file path (Tab to complete)"
	ti.CharLimit = 156
	ti.Width = 40

	// Set initial items if available
	items := make([]list.Item, len(s.DocFiles))
	for i, f := range s.DocFiles {
		items[i] = f
	}
	l.SetItems(items)

	return &DiscoveryPage{
		list:      l,
		textInput: ti,
		session:   s,
		width:     w,
		height:    h,
	}
}

func (p *DiscoveryPage) Init() tea.Cmd {
	if len(p.session.AllFiles) == 0 {
		return discoverFilesCmd()
	}
	return nil
}

func (p *DiscoveryPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case filesDiscoveredMsg:
		p.session.AllFiles = msg.allFiles
		p.session.DocFiles, p.session.CodeFiles = classifyFiles(msg.allFiles)

		items := make([]list.Item, len(p.session.DocFiles))
		for i, f := range p.session.DocFiles {
			items[i] = f
		}
		p.list.SetItems(items)
		return p, nil

	case tea.KeyMsg:
		if p.isAdding {
			switch msg.String() {
			case "esc":
				p.isAdding = false
				p.textInput.Blur()
				return p, nil
			case "enter":
				// Handle adding file (logic migrated from wizard.go)
				path := strings.TrimSpace(p.textInput.Value())
				if err := p.addFile(path); err != nil {
					p.feedbackMsg = err.Error()
				} else {
					p.isAdding = false
					p.textInput.Blur()
					p.textInput.SetValue("")
					p.feedbackMsg = ""
				}
				return p, nil
			case "tab":
				p.autocomplete()
				return p, nil
			}
			p.textInput, cmd = p.textInput.Update(msg)
			return p, cmd
		}

		switch msg.String() {
		case "enter":
			return p, func() tea.Msg { return NextStepMsg{} }
		case "a":
			p.isAdding = true
			p.textInput.Focus()
			return p, textinput.Blink
		case "space":
			if selectedItem, ok := p.list.SelectedItem().(FileInfo); ok {
				index := p.list.Index()
				selectedItem.IsIgnored = !selectedItem.IsIgnored
				p.list.SetItem(index, selectedItem)
				p.session.DocFiles[index].IsIgnored = selectedItem.IsIgnored
			}
		}
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.list.SetSize(msg.Width, msg.Height)
	}

	if !p.isAdding {
		p.list, cmd = p.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return p, tea.Batch(cmds...)
}

func (p *DiscoveryPage) View() string {
	if p.isAdding {
		box := FocusedInputStyle.Render(p.textInput.View())
		label := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("Add a file path manually:")
		feedback := ErrorStyle.Render(p.feedbackMsg)

		// Center the modal
		content := lipgloss.JoinVertical(lipgloss.Left, label, box, feedback)
		return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center, content)
	}
	return p.list.View()
}

func (p *DiscoveryPage) Title() string { return "Discovery" }
func (p *DiscoveryPage) Step() string  { return "Step 1 of 4" }
func (p *DiscoveryPage) Keys() []key.Binding {
	if p.isAdding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "complete")),
		}
	}
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next")),
		key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add file")),
		key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "toggle ignore")),
	}
}

// Logic helpers
func (p *DiscoveryPage) addFile(path string) error {
	// ... Logic from wizard.go ...
	// Simplified for brevity, will implement full logic
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("file not found")
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory")
	}

	newFile := FileInfo{Path: path, Size: info.Size(), Type: TypeDoc}
	p.session.DocFiles = append(p.session.DocFiles, newFile)
	p.list.InsertItem(len(p.list.Items()), newFile)
	return nil
}

func (p *DiscoveryPage) autocomplete() {
	// ... Logic from wizard.go ...
}
