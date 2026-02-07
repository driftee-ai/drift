package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/llm"
)

type GroupingPage struct {
	list       list.Model
	textInput  textinput.Model
	spinner    spinner.Model
	session    *Session
	width      int
	height     int
	isLoading  bool
	loadingMsg string

	// Edit states
	isEditingName bool
	isEditingDocs bool
	editIndex     int
}

func NewGroupingPage(s *Session, w, h int) *GroupingPage {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), w, h)
	l.Title = "Review Feature Groups"
	l.SetStatusBarItemName("group", "groups")

	ti := textinput.New()
	ti.CharLimit = 156
	ti.Width = 50

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(HighlightColor)

	return &GroupingPage{
		list:       l,
		textInput:  ti,
		spinner:    sp,
		session:    s,
		width:      w,
		height:     h,
		isLoading:  true,
		loadingMsg: "Analyzing documentation to identify features...",
	}
}

func (p *GroupingPage) Init() tea.Cmd {
	if len(p.session.Groups) > 0 {
		p.isLoading = false
		p.refreshList()
		return nil
	}
	return tea.Batch(
		p.spinner.Tick,
		p.generateGroupsCmd(),
	)
}

func (p *GroupingPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case groupsGeneratedMsg:
		p.session.Groups = msg.groups
		p.refreshList()
		p.isLoading = false
		return p, nil

	case tea.KeyMsg:
		if p.isLoading {
			return p, nil // Ignore keys while loading
		}

		if p.isEditingName || p.isEditingDocs {
			switch msg.String() {
			case "esc":
				p.isEditingName = false
				p.isEditingDocs = false
				p.textInput.Blur()
				return p, nil
			case "enter":
				// Save changes
				if p.editIndex >= 0 && p.editIndex < len(p.session.Groups) {
					if p.isEditingName {
						p.session.Groups[p.editIndex].Name = p.textInput.Value()
					} else if p.isEditingDocs {
						globs := strings.Split(p.textInput.Value(), ",")
						for i := range globs {
							globs[i] = strings.TrimSpace(globs[i])
						}
						p.session.Groups[p.editIndex].Docs = globs
					}
					p.refreshList()
				}
				p.isEditingName = false
				p.isEditingDocs = false
				p.textInput.Blur()
				return p, nil
			}
			p.textInput, cmd = p.textInput.Update(msg)
			return p, cmd
		}

		switch msg.String() {
		case "enter":
			return p, func() tea.Msg { return NextStepMsg{} }
		case "r":
			if idx := p.list.Index(); idx >= 0 {
				p.startEditName(idx)
				return p, textinput.Blink
			}
		case "e":
			if idx := p.list.Index(); idx >= 0 {
				p.startEditDocs(idx)
				return p, textinput.Blink
			}
		case "d":
			idx := p.list.Index()
			if idx >= 0 && idx < len(p.session.Groups) {
				p.session.Groups = append(p.session.Groups[:idx], p.session.Groups[idx+1:]...)
				p.list.Select(idx - 1)
				p.refreshList()
			}
		case "a":
			newRule := config.Rule{Name: "New Feature", Docs: []string{}}
			p.session.Groups = append(p.session.Groups, newRule)
			p.refreshList()
			p.list.Select(len(p.session.Groups) - 1)
			p.startEditName(len(p.session.Groups) - 1)
			return p, textinput.Blink

		case "ctrl+r":
			p.isLoading = true
			p.session.Groups = []config.Rule{} // Clear existing
			return p, tea.Batch(p.spinner.Tick, p.generateGroupsCmd())
		}

	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.list.SetSize(msg.Width, msg.Height)
	}

	if p.isLoading {
		p.spinner, cmd = p.spinner.Update(msg)
		cmds = append(cmds, cmd)
	} else if !p.isEditingName && !p.isEditingDocs {
		p.list, cmd = p.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return p, tea.Batch(cmds...)
}

func (p *GroupingPage) View() string {
	if p.isLoading {
		return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center,
			fmt.Sprintf("%s %s", p.spinner.View(), p.loadingMsg),
		)
	}

	if p.isEditingName || p.isEditingDocs {
		var title string
		if p.isEditingName {
			title = "Rename Feature:"
		} else {
			title = "Edit Doc Patterns (comma separated):"
		}

		box := FocusedInputStyle.Render(p.textInput.View())
		label := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(title)

		content := lipgloss.JoinVertical(lipgloss.Left, label, box)
		return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center, content)
	}

	return p.list.View()
}

func (p *GroupingPage) Title() string { return "Feature Grouping" }

func (p *GroupingPage) Step() string { return "Step 3 of 5" }

func (p *GroupingPage) Keys() []key.Binding {
	if p.isLoading {
		return []key.Binding{}
	}
	if p.isEditingName || p.isEditingDocs {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "save")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		}
	}
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename")),
		key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit docs")),
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
		key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "regenerate")),
	}
}

// Helpers

func (p *GroupingPage) refreshList() {
	items := make([]list.Item, len(p.session.Groups))
	for i, g := range p.session.Groups {
		items[i] = RuleItem{Rule: g, Index: i}
	}
	p.list.SetItems(items)
}

func (p *GroupingPage) startEditName(idx int) {
	p.isEditingName = true
	p.editIndex = idx
	p.textInput.SetValue(p.session.Groups[idx].Name)
	p.textInput.Focus()
}

func (p *GroupingPage) startEditDocs(idx int) {
	p.isEditingDocs = true
	p.editIndex = idx
	p.textInput.SetValue(strings.Join(p.session.Groups[idx].Docs, ", "))
	p.textInput.Focus()
}

// Structs for list items (reused across pages potentially, but defined locally if needed or in common)
// We need RuleItem in common or here. It was in wizard.go before.
// We should probably move RuleItem to pages.go or keep it here if only used here.
// But MappingPage uses it too? MappingPage uses MappingItem.
// Let's define it here for now, or check generic implementation.
// Previously in wizard.go:
/*
type RuleItem struct {
	Rule  config.Rule
	Index int
}
func (r RuleItem) Title() string       { return r.Rule.Name }
func (r RuleItem) Description() string { return strings.Join(r.Rule.Docs, ", ") }
func (r RuleItem) FilterValue() string { return r.Rule.Name }
*/

// Re-defining locally for simplicity, or we could put in pages.go
type RuleItem struct {
	Rule  config.Rule
	Index int
}

func (r RuleItem) Title() string       { return r.Rule.Name }
func (r RuleItem) Description() string { return strings.Join(r.Rule.Docs, ", ") }
func (r RuleItem) FilterValue() string { return r.Rule.Name }

// LLM Logic
type groupsGeneratedMsg struct {
	groups []config.Rule
}

func (p *GroupingPage) generateGroupsCmd() tea.Cmd {
	return func() tea.Msg {
		// Collect selected doc files
		var docPaths []string
		for _, f := range p.session.DocFiles {
			if !f.IsIgnored {
				docPaths = append(docPaths, f.Path)
			}
		}

		if len(docPaths) == 0 {
			return groupsGeneratedMsg{groups: []config.Rule{{Name: "All Docs", Docs: []string{"**/*.md"}}}}
		}

		gen, err := llm.New(p.session.Provider)
		if err != nil {
			return groupsGeneratedMsg{groups: []config.Rule{{Name: "General Documentation", Docs: []string{"**/*.md"}}}}
		}

		prompt := fmt.Sprintf(`You are a software architect. I will provide a list of documentation files. Please group them into logical features (e.g., 'Authentication', 'Billing', 'API').

Requirements:
1. Create a 'General' group for generic files (README, contributing, etc) if they exist.
2. For each group, provide a short 'name' and a list of 'docs' glob patterns (e.g., 'docs/auth/**/*.md') that cover the files in that group. Use wildcards ('*', '**') effectively.
3. Return valid JSON matching the requested schema.

Files:
%s`, strings.Join(docPaths, "\n"))

		schema := &llm.Schema{
			Type: llm.TypeObject,
			Properties: map[string]*llm.Schema{
				"groups": {
					Type: llm.TypeArray,
					Items: &llm.Schema{
						Type: llm.TypeObject, // Add name and docs properties
						Properties: map[string]*llm.Schema{
							"name": {Type: llm.TypeString},
							"docs": {
								Type:  llm.TypeArray,
								Items: &llm.Schema{Type: llm.TypeString},
							},
						},
						Required: []string{"name", "docs"},
					},
				},
			},
			Required: []string{"groups"},
		}

		type response struct {
			Groups []config.Rule `json:"groups"`
		}
		var resp response
		err = gen.GenerateJSON(context.Background(), prompt, schema, &resp)
		if err != nil {
			return groupsGeneratedMsg{groups: []config.Rule{{Name: "General Documentation", Docs: []string{"**/*.md"}}}}
		}

		return groupsGeneratedMsg{groups: resp.Groups}
	}
}
