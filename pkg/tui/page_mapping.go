package tui

import (
	"context"
	"encoding/json"
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

type MappingPage struct {
	list       list.Model
	textInput  textinput.Model
	spinner    spinner.Model
	session    *Session
	width      int
	height     int
	isLoading  bool
	loadingMsg string

	isEditingCode bool
	editIndex     int
}

func NewMappingPage(s *Session, w, h int) *MappingPage {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), w, h)
	l.Title = "Review Code Mappings"
	l.SetStatusBarItemName("mapping", "mappings")

	ti := textinput.New()
	ti.CharLimit = 156
	ti.Width = 50

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(HighlightColor)

	return &MappingPage{
		list:       l,
		textInput:  ti,
		spinner:    sp,
		session:    s,
		width:      w,
		height:     h,
		isLoading:  true,
		loadingMsg: "Mapping code to features...",
	}
}

func (p *MappingPage) Init() tea.Cmd {
	// Simple check: if we have groups and the first one has code mapped, assume done.
	// Or just check if isLoading was set to false previously.
	if !p.isLoading {
		return nil
	}
	// Initial load
	return tea.Batch(
		p.spinner.Tick,
		p.generateMappingsCmd(),
	)
}

func (p *MappingPage) Update(msg tea.Msg) (Page, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case mappingsGeneratedMsg:
		p.session.Groups = msg.groups
		p.refreshList()
		p.isLoading = false
		return p, nil

	case tea.KeyMsg:
		if p.isLoading {
			return p, nil
		}

		if p.isEditingCode {
			switch msg.String() {
			case "esc":
				p.isEditingCode = false
				p.textInput.Blur()
				return p, nil
			case "enter":
				if p.editIndex >= 0 && p.editIndex < len(p.session.Groups) {
					globs := strings.Split(p.textInput.Value(), ",")
					for i := range globs {
						globs[i] = strings.TrimSpace(globs[i])
					}
					p.session.Groups[p.editIndex].Code = globs
					p.refreshList()
				}
				p.isEditingCode = false
				p.textInput.Blur()
				return p, nil
			}
			p.textInput, cmd = p.textInput.Update(msg)
			return p, cmd
		}

		switch msg.String() {
		case "enter":
			return p, func() tea.Msg { return NextStepMsg{} }
		case "e":
			if idx := p.list.Index(); idx >= 0 {
				p.startEditCode(idx)
				return p, textinput.Blink
			}
		case "ctrl+r":
			p.isLoading = true
			return p, tea.Batch(p.spinner.Tick, p.generateMappingsCmd())
		}

	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.list.SetSize(msg.Width, msg.Height)
	}

	if p.isLoading {
		p.spinner, cmd = p.spinner.Update(msg)
		cmds = append(cmds, cmd)
	} else if !p.isEditingCode {
		p.list, cmd = p.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return p, tea.Batch(cmds...)
}

func (p *MappingPage) View() string {
	if p.isLoading {
		return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center,
			fmt.Sprintf("%s %s", p.spinner.View(), p.loadingMsg),
		)
	}

	if p.isEditingCode {
		box := FocusedInputStyle.Render(p.textInput.View())
		label := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("Edit Code Patterns (comma separated):")

		content := lipgloss.JoinVertical(lipgloss.Left, label, box)
		return lipgloss.Place(p.width, p.height, lipgloss.Center, lipgloss.Center, content)
	}

	return p.list.View()
}

func (p *MappingPage) Title() string { return "Code Mapping" }
func (p *MappingPage) Step() string  { return "Step 4 of 5" }

func (p *MappingPage) Keys() []key.Binding {
	if p.isLoading {
		return []key.Binding{}
	}
	if p.isEditingCode {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "save")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		}
	}
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next")),
		key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit code")),
		key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "regenerate")),
	}
}

// Helpers

func (p *MappingPage) refreshList() {
	items := make([]list.Item, len(p.session.Groups))
	for i, g := range p.session.Groups {
		items[i] = MappingItem{Rule: g, Index: i}
	}
	p.list.SetItems(items)
}

func (p *MappingPage) startEditCode(idx int) {
	p.isEditingCode = true
	p.editIndex = idx
	p.textInput.SetValue(strings.Join(p.session.Groups[idx].Code, ", "))
	p.textInput.Focus()
}

type MappingItem struct {
	Rule  config.Rule
	Index int
}

func (m MappingItem) Title() string       { return m.Rule.Name }
func (m MappingItem) Description() string { return strings.Join(m.Rule.Code, ", ") }
func (m MappingItem) FilterValue() string { return m.Rule.Name }

// LLM Logic
type mappingsGeneratedMsg struct {
	groups []config.Rule
}

func (p *MappingPage) generateMappingsCmd() tea.Cmd {
	return func() tea.Msg {
		var codePaths []string
		for _, f := range p.session.CodeFiles {
			if len(codePaths) < 500 { // Keep safety limit for now
				codePaths = append(codePaths, f.Path)
			}
		}

		type GroupStub struct {
			Name string   `json:"name"`
			Docs []string `json:"docs"`
		}
		var groupStubs []GroupStub
		for _, g := range p.session.Groups {
			groupStubs = append(groupStubs, GroupStub{Name: g.Name, Docs: g.Docs})
		}

		groupsJSON, _ := json.Marshal(groupStubs)

		prompt := fmt.Sprintf(`You are a software architect. I will provide a list of code files and a set of feature groups. Please identify which code files likely belong to each feature group.

Requirements:
1. For each group, provide a list of 'code' glob patterns (e.g., 'src/auth/**/*.go') that cover the implementation.
2. Use wildcards ('*', '**') effectively.
3. If a group seems to have no corresponding code, leave the code list empty.
4. Return valid JSON.

Feature Groups:
%s

Code Files:
%s`, string(groupsJSON), strings.Join(codePaths, "\n"))

		schema := &llm.Schema{
			Type: llm.TypeObject,
			Properties: map[string]*llm.Schema{
				"mappings": {
					Type: llm.TypeArray,
					Items: &llm.Schema{
						Type: llm.TypeObject,
						Properties: map[string]*llm.Schema{
							"name": {Type: llm.TypeString},
							"code": {
								Type:  llm.TypeArray,
								Items: &llm.Schema{Type: llm.TypeString},
							},
						},
						Required: []string{"name", "code"},
					},
				},
			},
			Required: []string{"mappings"},
		}

		gen, err := llm.New(p.session.Provider)
		if err != nil {
			return mappingsGeneratedMsg{groups: p.session.Groups}
		}

		type mapping struct {
			Name string   `json:"name"`
			Code []string `json:"code"`
		}
		type response struct {
			Mappings []mapping `json:"mappings"`
		}
		var resp response
		err = gen.GenerateJSON(context.Background(), prompt, schema, &resp)
		if err != nil {
			return mappingsGeneratedMsg{groups: p.session.Groups}
		}

		newGroups := make([]config.Rule, len(p.session.Groups))
		copy(newGroups, p.session.Groups)

		for _, mapItem := range resp.Mappings {
			for i, g := range newGroups {
				if g.Name == mapItem.Name {
					newGroups[i].Code = mapItem.Code
					break
				}
			}
		}

		return mappingsGeneratedMsg{groups: newGroups}
	}
}
