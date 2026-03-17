package initwizard

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/files"
	"github.com/driftee-ai/drift/pkg/llm"
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

type screenState int

const (
	stateProvider screenState = iota
	stateLoading
	stateReview
	stateSummary
)

type providerItem struct {
	id, title, desc string
}

func (i providerItem) Title() string       { return i.title }
func (i providerItem) Description() string { return i.desc }
func (i providerItem) FilterValue() string { return i.title }

type ruleItem struct {
	rule config.Rule
}

func (i ruleItem) Title() string { return i.rule.Name }
func (i ruleItem) Description() string {
	return fmt.Sprintf("%d code globs, %d doc globs", len(i.rule.Code), len(i.rule.Docs))
}
func (i ruleItem) FilterValue() string { return i.rule.Name }

type tuiModel struct {
	state      screenState
	spinner    spinner.Model
	prog       progress.Model
	loadingMsg string
	err        error

	dir      string
	fastMode bool
	provider string
	usage    llm.Usage

	providerList list.Model
	rules        []config.Rule
	list         list.Model
	inputs       []textinput.Model
	focusIndex   focusState

	viewport viewport.Model
	ready    bool

	width  int
	height int

	quitting bool
	saved    bool
}

func initialModel(dir string, fastMode bool) tuiModel {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	prog := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))

	items := []list.Item{
		providerItem{id: "gemini", title: "Google Gemini (Default)", desc: "Uses the Gemini API for fast, reliable auto-discovery (requires GEMINI_API_KEY)"},
		providerItem{id: "openai", title: "OpenAI", desc: "Uses OpenAI's models (requires OPENAI_API_KEY)"},
		providerItem{id: "anthropic", title: "Anthropic", desc: "Uses Claude models (requires ANTHROPIC_API_KEY)"},
	}

	pList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	pList.Title = "Select an LLM Provider for Auto-Discovery"
	pList.Styles.Title = titleStyle
	pList.SetShowStatusBar(false)
	pList.SetFilteringEnabled(false)

	return tuiModel{
		state:        stateProvider,
		spinner:      s,
		prog:         prog,
		loadingMsg:   "Analyzing repository...",
		dir:          dir,
		fastMode:     fastMode,
		providerList: pList,
		focusIndex:   focusList,
	}
}

type scanResultMsg struct {
	files map[string]string
	err   error
}

type mapResultMsg struct {
	rules []config.Rule
	usage llm.Usage
	err   error
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) startScan() tea.Cmd {
	return func() tea.Msg {
		files, err := ScanProject(m.dir, m.fastMode)
		return scanResultMsg{files: files, err: err}
	}
}

func (m tuiModel) startMap(files map[string]string) tea.Cmd {
	return func() tea.Msg {
		client, err := llm.New(m.provider)
		if err != nil {
			return mapResultMsg{err: err}
		}
		mapper := NewMapper(client)
		rules, usage, err := mapper.MapFiles(context.Background(), files, m.fastMode)
		return mapResultMsg{rules: rules, usage: usage, err: err}
	}
}

func (m *tuiModel) initReviewState() {
	items := make([]list.Item, len(m.rules))
	for i, r := range m.rules {
		items[i] = ruleItem{rule: r}
	}

	m.list = list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.list.SetShowTitle(false)
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(false)

	var inputs []textinput.Model = make([]textinput.Model, 3)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Rule Name"
	inputs[0].Focus()
	inputs[0].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	inputs[0].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Code Globs (comma separated)"

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Doc Globs (comma separated)"

	if len(m.rules) > 0 {
		inputs[0].SetValue(m.rules[0].Name)
		inputs[1].SetValue(strings.Join(m.rules[0].Code, ", "))
		inputs[2].SetValue(strings.Join(m.rules[0].Docs, ", "))
	}
	m.inputs = inputs
	m.focusIndex = focusList
	m.ready = false
}

func (m *tuiModel) resizeComponents() {
	if m.width == 0 || m.height == 0 {
		return
	}
	h, _ := docStyle.GetFrameSize()

	targetPaneHeight := m.height - 12
	paneContentHeight := targetPaneHeight - 4 // border + padding = 4

	listHeight := paneContentHeight - 3 // title and spacing = 3 lines
	if listHeight < 0 {
		listHeight = 0
	}
	m.list.SetSize(m.width/3-h-2, listHeight)

	viewportHeight := paneContentHeight - 17 // inputs, labels, titles and newlines = 17 lines
	if viewportHeight < 0 {
		viewportHeight = 0
	}

	if !m.ready {
		m.viewport = viewport.New((m.width*2)/3-h-2, viewportHeight)
		m.viewport.SetContent("Glob matched files will appear here...")
		m.ready = true
		m.updateViewportContent()
	} else {
		m.viewport.Width = (m.width*2)/3 - h - 2
		m.viewport.Height = viewportHeight
	}
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
			if m.state == stateReview && m.focusIndex != focusList {
				m.setFocus(focusList)
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "tab", "shift+tab":
			if m.state == stateReview {
				s := msg.String()
				if s == "shift+tab" {
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
			}
		case "up", "down":
			if m.state == stateReview && m.focusIndex != focusList {
				s := msg.String()
				if s == "up" {
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
			}
		case "enter":
			if m.state == stateProvider {
				if item, ok := m.providerList.SelectedItem().(providerItem); ok {
					m.provider = item.id
					m.state = stateLoading
					return m, tea.Batch(m.spinner.Tick, m.startScan())
				}
				return m, nil
			}
			if m.state == stateReview {
				if m.focusIndex == focusList {
					if m.list.Index() == len(m.list.Items())-1 {
						m.state = stateSummary
						return m, nil
					}
					m.list.CursorDown()
					m.loadCurrentRuleIntoInputs()
					return m, nil
				}
				m.saveInputsToCurrentRule()
				m.setFocus(focusList)
				return m, nil
			}
			if m.state == stateSummary {
				m.saved = true
				m.quitting = true
				return m, tea.Quit
			}
		case "d", "backspace", "delete":
			if m.state == stateReview && m.focusIndex == focusList && len(m.list.Items()) > 0 {
				m.list.RemoveItem(m.list.Index())
				m.loadCurrentRuleIntoInputs()
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.providerList.SetSize(m.width-h, m.height-v)
		m.prog.Width = m.width - h - 4
		if m.state == stateReview {
			m.resizeComponents()
		}

	case spinner.TickMsg:
		if m.state == stateLoading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case scanResultMsg:
		if msg.err != nil {
			m.err = msg.err
			m.quitting = true
			return m, tea.Quit
		}
		if m.fastMode {
			m.loadingMsg = fmt.Sprintf("Found %d relevant files. Asking %s LLM to map documentation via file paths...", len(msg.files), m.provider)
		} else {
			m.loadingMsg = fmt.Sprintf("Read contents of %d relevant files. Asking %s LLM to map documentation...", len(msg.files), m.provider)
		}
		return m, m.startMap(msg.files)

	case mapResultMsg:
		if msg.err != nil {
			m.err = msg.err
			m.quitting = true
			return m, tea.Quit
		}
		m.rules = msg.rules
		m.usage = msg.usage
		if len(m.rules) == 0 {
			m.rules = []config.Rule{
				{
					Name: "Example API Documentation",
					Code: []string{"src/api/**/*.go"},
					Docs: []string{"docs/api/**/*.md"},
				},
			}
		}
		m.initReviewState()
		m.resizeComponents()
		m.state = stateReview
		return m, textinput.Blink
	}

	// Route updates by state
	if m.state == stateProvider {
		m.providerList, cmd = m.providerList.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.state == stateReview {
		if m.focusIndex == focusList {
			var wasIndex = m.list.Index()
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			if m.list.Index() != wasIndex {
				m.loadCurrentRuleIntoInputs()
			}
		} else {
			cmd = m.updateInputs(msg)
			cmds = append(cmds, cmd)
			if _, ok := msg.(tea.KeyMsg); ok {
				m.updateViewportContent()
			}
		}
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

	var out strings.Builder
	if m.focusIndex == focusInputCode {
		out.WriteString("🔍 Previewing Code Globs:\n")
		globs := splitGlobs(m.inputs[1].Value())
		matches, _ := files.FindFiles(".", globs)
		out.WriteString(fmt.Sprintf("%d files matched.\n\n", len(matches)))
		for _, match := range matches {
			out.WriteString("> " + match + "\n")
		}
	} else if m.focusIndex == focusInputDocs {
		out.WriteString("🔍 Previewing Doc Globs:\n")
		globs := splitGlobs(m.inputs[2].Value())
		matches, _ := files.FindFiles(".", globs)
		out.WriteString(fmt.Sprintf("%d files matched.\n\n", len(matches)))
		for _, match := range matches {
			out.WriteString("> " + match + "\n")
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

	if m.state == stateProvider {
		return docStyle.Render(m.providerList.View())
	}

	if m.state == stateLoading {
		loadingView := fmt.Sprintf("\n\n  %s %s\n\n", m.spinner.View(), m.loadingMsg)
		return docStyle.Render(loadingView)
	}

	if m.state == stateSummary {
		summary := fmt.Sprintf("\n\n  %s\n\n  %s %d rules configured.\n\n  ⚡ %d tokens used during discovery.\n\n  %s\n\n",
			titleStyle.Render("Configuration Ready"),
			"✅", len(m.list.Items()),
			m.usage.TotalTokens,
			subtextStyle.Render("Press ENTER to generate .drift.yaml, or ESC to cancel."),
		)
		return docStyle.Render(summary)
	}

	// State Review
	pct := 0.0
	if len(m.list.Items()) > 0 {
		pct = float64(m.list.Index()+1) / float64(len(m.list.Items()))
	}

	headerStr := fmt.Sprintf("%s\n\n%s %d of %d Rules\n%s\n\n",
		titleStyle.Render("Drift: Auto-Discovery Wizard"),
		subtextStyle.Render("Progress:"),
		m.list.Index()+1,
		len(m.list.Items()),
		m.prog.ViewAs(pct),
	)

	listStyle := paneStyle
	if m.focusIndex == focusList {
		listStyle = focusedPaneStyle
	}

	var listContent strings.Builder
	listContent.WriteString(titleStyle.Render("Discovered Rules"))
	listContent.WriteString("\n\n")
	listContent.WriteString(m.list.View())
	leftPane := listStyle.Width(m.width/3 - 4).Height(m.height - 12).Render(listContent.String())

	editorStyle := paneStyle
	if m.focusIndex != focusList {
		editorStyle = focusedPaneStyle
	}

	var editorContent strings.Builder
	editorContent.WriteString(titleStyle.Render("Rule Editor"))
	editorContent.WriteString("\n\n")

	editorContent.WriteString(subtextStyle.Render("Rule Name") + "\n")
	editorContent.WriteString(m.inputs[0].View() + "\n\n")

	editorContent.WriteString(subtextStyle.Render("Code Files (Globs)") + "\n")
	editorContent.WriteString(m.inputs[1].View() + "\n\n")

	editorContent.WriteString(subtextStyle.Render("Documentation Files (Globs)") + "\n")
	editorContent.WriteString(m.inputs[2].View() + "\n\n")

	editorContent.WriteString(titleStyle.Render("Live Output Preview"))
	editorContent.WriteString("\n")
	editorContent.WriteString(m.viewport.View())

	rightPane := editorStyle.Width((m.width*2)/3 - 4).Height(m.height - 12).Render(editorContent.String())

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	helpBar := subtextStyle.Render("↑/↓: Navigate • TAB: Edit Rule • ENTER: Accept & Next • D: Delete • ESC: Quit")

	return docStyle.Render(headerStr + mainView + "\n\n" + helpBar)
}

func runReviewApp(dir string, fastMode bool) ([]config.Rule, llm.Usage, string, error) {
	p := tea.NewProgram(initialModel(dir, fastMode), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return nil, llm.Usage{}, "", err
	}

	model := m.(tuiModel)
	if model.err != nil {
		return nil, llm.Usage{}, "", model.err
	}

	if !model.saved {
		return nil, llm.Usage{}, "", nil // User quit without saving
	}

	var finalRules []config.Rule
	for _, item := range model.list.Items() {
		finalRules = append(finalRules, item.(ruleItem).rule)
	}
	return finalRules, model.usage, model.provider, nil
}
