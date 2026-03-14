package initwizard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/files"
)

var (
	docStyle         = lipgloss.NewStyle().Margin(1, 2)
	paneStyle        = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).Padding(1, 2)
	focusedPaneStyle = lipgloss.NewStyle().Border(lipgloss.DoubleBorder(), true).BorderForeground(lipgloss.Color("62")).Padding(1, 2)
	titleStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	subtextStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type focusState int

const (
	focusList focusState = iota
	focusInputName
	focusInputCode
	focusInputDocs
)

type ruleItem struct {
	rule config.Rule
}

func (i ruleItem) Title() string { return i.rule.Name }
func (i ruleItem) Description() string {
	return fmt.Sprintf("%d code globs, %d doc globs", len(i.rule.Code), len(i.rule.Docs))
}
func (i ruleItem) FilterValue() string { return i.rule.Name }

type tuiModel struct {
	rules      []config.Rule
	list       list.Model
	inputs     []textinput.Model
	focusIndex focusState

	viewport viewport.Model
	ready    bool

	width  int
	height int

	quitting bool
	saved    bool
}

func initialModel(rules []config.Rule) tuiModel {
	items := make([]list.Item, len(rules))
	for i, r := range rules {
		items[i] = ruleItem{rule: r}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Discovered Rules"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false) // Disable filtering to keep it clean for now

	var inputs []textinput.Model = make([]textinput.Model, 3)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Rule Name"
	inputs[0].Focus()
	inputs[0].PromptStyle = focusedPaneStyle
	inputs[0].TextStyle = focusedPaneStyle

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Code Globs (comma separated)"

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Doc Globs (comma separated)"

	// Load initial data if items exist
	if len(rules) > 0 {
		inputs[0].SetValue(rules[0].Name)
		inputs[1].SetValue(strings.Join(rules[0].Code, ", "))
		inputs[2].SetValue(strings.Join(rules[0].Docs, ", "))
	}

	return tuiModel{
		rules:      rules,
		list:       l,
		inputs:     inputs,
		focusIndex: focusList,
	}
}

func (m tuiModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			if m.focusIndex != focusList {
				m.setFocus(focusList)
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "ctrl+s":
			m.saved = true
			m.quitting = true
			return m, tea.Quit
		case "tab", "shift+tab", "up", "down":
			if m.focusIndex == focusList {
				// List specific navigation
				m.list, cmd = m.list.Update(msg)
				m.loadCurrentRuleIntoInputs()
				return m, cmd
			}

			// Form navigation
			s := msg.String()

			// Move up and down the form, and pop back out to the list
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
				if m.focusIndex < focusList {
					m.focusIndex = focusInputDocs
				}
			} else {
				m.focusIndex++
				if m.focusIndex > focusInputDocs {
					m.focusIndex = focusList
				}
			}

			m.setFocus(m.focusIndex)
			return m, nil

		case "enter":
			// If we hit enter on the list, transition into editing
			if m.focusIndex == focusList {
				m.setFocus(focusInputName)
				return m, nil
			}

			// If editing, save changes back to the list item
			m.saveInputsToCurrentRule()
			return m, nil

		case "d", "backspace", "delete":
			if m.focusIndex == focusList && len(m.list.Items()) > 0 {
				// Delete the current rule feature
				m.list.RemoveItem(m.list.Index())
				m.loadCurrentRuleIntoInputs()
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		h, v := docStyle.GetFrameSize()
		m.list.SetSize(m.width/3-h, m.height-v-4)

		if !m.ready {
			m.viewport = viewport.New((m.width*2)/3-h, m.height-v-15)
			m.viewport.SetContent("Glob matched files will appear here...")
			m.ready = true
		} else {
			m.viewport.Width = (m.width*2)/3 - h
			m.viewport.Height = m.height - v - 15
		}
	}

	// Handle list updates
	if m.focusIndex == focusList {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Handle input updates
	if m.focusIndex != focusList {
		cmd = m.updateInputs(msg)
		cmds = append(cmds, cmd)
		m.updateViewportContent() // Live preview capability
	}

	return m, tea.Batch(cmds...)
}

func (m *tuiModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m *tuiModel) setFocus(index focusState) {
	m.focusIndex = index
	for i := 0; i < len(m.inputs); i++ {
		if index == focusState(i+1) {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *tuiModel) loadCurrentRuleIntoInputs() {
	item := m.list.SelectedItem()
	if item == nil {
		for i := range m.inputs {
			m.inputs[i].SetValue("")
		}
		return
	}
	rule := item.(ruleItem).rule
	m.inputs[0].SetValue(rule.Name)
	m.inputs[1].SetValue(strings.Join(rule.Code, ", "))
	m.inputs[2].SetValue(strings.Join(rule.Docs, ", "))
	m.updateViewportContent()
}

func (m *tuiModel) saveInputsToCurrentRule() {
	idx := m.list.Index()
	item := m.list.SelectedItem()
	if item == nil {
		return
	}

	rule := item.(ruleItem).rule
	rule.Name = m.inputs[0].Value()

	rule.Code = []string{}
	for _, p := range strings.Split(m.inputs[1].Value(), ",") {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			rule.Code = append(rule.Code, trimmed)
		}
	}

	rule.Docs = []string{}
	for _, p := range strings.Split(m.inputs[2].Value(), ",") {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			rule.Docs = append(rule.Docs, trimmed)
		}
	}

	m.list.SetItem(idx, ruleItem{rule: rule})
}

func (m *tuiModel) updateViewportContent() {
	if !m.ready {
		return
	}

	// Live glob preview logic goes here!
	var out strings.Builder

	if m.focusIndex == focusInputCode {
		out.WriteString("🔍 Previewing Code Globs:\n")
		globs := splitGlobs(m.inputs[1].Value())
		matches, _ := files.FindFiles(globs)
		out.WriteString(fmt.Sprintf("%d files matched.\n\n", len(matches)))
		for _, match := range matches {
			out.WriteString("- " + match + "\n")
		}
	} else if m.focusIndex == focusInputDocs {
		out.WriteString("🔍 Previewing Doc Globs:\n")
		globs := splitGlobs(m.inputs[2].Value())
		matches, _ := files.FindFiles(globs)
		out.WriteString(fmt.Sprintf("%d files matched.\n\n", len(matches)))
		for _, match := range matches {
			out.WriteString("- " + match + "\n")
		}
	} else {
		out.WriteString("Select 'Code Globs' or 'Doc Globs' input to see live evaluations.")
	}

	m.viewport.SetContent(out.String())
}

func splitGlobs(in string) []string {
	var res []string
	for _, p := range strings.Split(in, ",") {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			res = append(res, trimmed)
		}
	}
	return res
}

func (m tuiModel) View() string {
	if m.quitting {
		return ""
	}

	// Left pane (List)
	listStyle := paneStyle
	if m.focusIndex == focusList {
		listStyle = focusedPaneStyle
	}
	leftPane := listStyle.Render(m.list.View())

	// Right pane (Editor)
	editorStyle := paneStyle
	if m.focusIndex != focusList {
		editorStyle = focusedPaneStyle
	}

	var editorContent strings.Builder
	editorContent.WriteString(titleStyle.Render("Rule Editor"))
	editorContent.WriteString("\n\n")

	editorContent.WriteString(m.inputs[0].View() + "\n\n")
	editorContent.WriteString(m.inputs[1].View() + "\n\n")
	editorContent.WriteString(m.inputs[2].View() + "\n\n")

	editorContent.WriteString(titleStyle.Render("Live Output Preview"))
	editorContent.WriteString("\n")
	editorContent.WriteString(m.viewport.View())

	rightPane := editorStyle.Width((m.width*2)/3 - 4).Height(m.height - 4).Render(editorContent.String())

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	helpBar := subtextStyle.Render("↑/↓: Navigate • TAB: Switch Pane • ENTER: Edit/Save • D: Delete Rule • CTRL+S: Finish & Generate • ESC: Quit")

	return docStyle.Render(mainView + "\n" + helpBar)
}

func runReviewApp(rules []config.Rule) ([]config.Rule, error) {
	p := tea.NewProgram(initialModel(rules), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return nil, err
	}

	model := m.(tuiModel)
	if !model.saved {
		return nil, nil // User quit without saving
	}

	var finalRules []config.Rule
	for _, item := range model.list.Items() {
		finalRules = append(finalRules, item.(ruleItem).rule)
	}
	return finalRules, nil
}
