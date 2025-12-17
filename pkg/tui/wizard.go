package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/driftee-ai/drift/pkg/files"
)

// State represents the current step of the wizard.
type State int

const (
	StateLoading State = iota
	StateDiscovery
	StateAddingFile
	StateGrouping
	StateMapping
	StateFinalize
	StateDone
	StateError
)

// DocFileType represents the classification of a file.
type DocFileType int

const (
	TypeDoc DocFileType = iota
	TypeCode
	TypeIgnored // Files that are explicitly ignored by the user or default rules
	TypeOther   // Files that don't fit into doc or code (e.g., config files, images)
)

// FileInfo holds information about a discovered file.
type FileInfo struct {
	Path       string
	Type       DocFileType
	Size       int64
	IsSelected bool   // For interactive selection
	IsIgnored  bool   // For interactive ignore
	LLMSummary string // LLM generated description
}

// FilterValue implements list.Item.
func (f FileInfo) FilterValue() string { return f.Path }

// Title implements list.Item.
func (f FileInfo) Title() string {
	if f.IsIgnored {
		return "[ ] " + f.Path
	}
	return "[x] " + f.Path
}

// Description implements list.Item.
func (f FileInfo) Description() string {
	return fmt.Sprintf("Size: %d bytes", f.Size)
}

// Model is the main Bubble Tea model for the init wizard.
type Model struct {
	state       State
	spinner     spinner.Model
	list        list.Model
	textInput   textinput.Model
	err         error
	width       int
	height      int
	allFiles    []FileInfo
	docFiles    []FileInfo
	codeFiles   []FileInfo
	loadingText string
	feedbackMsg string
}

// NewModel creates a new Model for the init wizard.
func NewModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(HighlightColor)

	// Initialize list with an empty set of items and a default delegate
	// This ensures m.list is never nil when SetSize is called
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Review Documentation Files" // Set initial title
	l.SetStatusBarItemName("file", "files")
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add file")),
			key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "toggle ignore")),
		}
	}

	ti := textinput.New()
	ti.Placeholder = "Enter file path (Tab to complete)"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return Model{
		state:       StateLoading,
		spinner:     s,
		list:        l, // Initialize the list model
		textInput:   ti,
		loadingText: "Starting wizard...",
	}
}

// Init initializes the model, performing initial file discovery.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, discoverFilesCmd())
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		// Handle State-specific keys
		if m.state == StateDiscovery {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "enter":
				// Confirm selection and move to next state
				m.state = StateGrouping
				// TODO: Initiate Grouping logic here (LLM call)
				return m, nil
			case "a":
				m.state = StateAddingFile
				m.textInput.SetValue("")
				m.textInput.Focus()
				m.feedbackMsg = ""
				return m, textinput.Blink
			case " ":
				// Toggle ignore status
				if selectedItem, ok := m.list.SelectedItem().(FileInfo); ok {
					index := m.list.Index()
					// Toggle status
					selectedItem.IsIgnored = !selectedItem.IsIgnored

					// Update in the list
					m.list.SetItem(index, selectedItem)

					// Update in our local state
					for i, f := range m.docFiles {
						if f.Path == selectedItem.Path {
							m.docFiles[i].IsIgnored = selectedItem.IsIgnored
							break
						}
					}

					return m, nil
				}
			}
		} else if m.state == StateAddingFile {
			switch msg.String() {
			case "esc":
				m.state = StateDiscovery
				m.textInput.Blur()
				m.feedbackMsg = ""
				return m, nil
			case "enter":
				path := m.textInput.Value()
				path = strings.TrimSpace(path) // basic cleanup

				var fileToAdd FileInfo
				found := false

				// 1. Check if it's already in our scanned files
				for _, f := range m.allFiles {
					if f.Path == path {
						fileToAdd = f
						found = true
						break
					}
				}

				// 2. If not found, check the filesystem directly (allows adding ignored files)
				if !found {
					info, err := os.Stat(path)
					if err == nil && !info.IsDir() {
						fileToAdd = FileInfo{
							Path: path,
							Size: info.Size(),
							Type: TypeDoc, // Force to Doc since user is explicitly adding it
						}
						found = true
					}
				}

				if found {
					// Check if already in docFiles to avoid duplicates
					alreadyInDocs := false
					for _, f := range m.docFiles {
						if f.Path == fileToAdd.Path {
							alreadyInDocs = true
							break
						}
					}

					if !alreadyInDocs {
						fileToAdd.Type = TypeDoc     // Ensure it's treated as a doc
						fileToAdd.IsIgnored = false  // Ensure it's active
						fileToAdd.IsSelected = false // Default state

						m.docFiles = append(m.docFiles, fileToAdd)
						m.list.InsertItem(len(m.list.Items()), fileToAdd)
					}
					m.state = StateDiscovery
					m.textInput.Blur()
					m.feedbackMsg = "" // Clear any previous error
				} else {
					m.feedbackMsg = fmt.Sprintf("Error: File '%s' not found or is a directory.", path)
				}
				return m, nil
			case "tab":
				// Autocomplete
				input := m.textInput.Value()
				if input == "." || input == "./" {
					return m, nil // Don't match everything starting with "."
				}

				cleanInput := strings.TrimPrefix(input, "./")
				var matches []string
				for _, f := range m.allFiles {
					if strings.HasPrefix(f.Path, cleanInput) {
						matches = append(matches, f.Path)
					}
				}

				if len(matches) == 1 {
					m.textInput.SetValue(matches[0])
					m.textInput.SetCursor(len(matches[0]))
					m.feedbackMsg = ""
				} else if len(matches) > 1 {
					// Find common prefix
					common := matches[0]
					for _, match := range matches[1:] {
						for !strings.HasPrefix(match, common) {
							common = common[:len(common)-1]
						}
					}
					m.textInput.SetValue(common)
					m.textInput.SetCursor(len(common))

					// Show some matches as feedback
					maxToShow := 3
					displayMatches := matches
					if len(matches) > maxToShow {
						displayMatches = matches[:maxToShow]
					}
					m.feedbackMsg = fmt.Sprintf("Matches: %s...", strings.Join(displayMatches, ", "))
				}
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Ensure list is not nil before calling SetSize
		if m.list.Width() == 0 && m.list.Height() == 0 { // Check if it's still at initial empty size
			m.list.SetSize(msg.Width, msg.Height)
		} else {
			m.list.SetSize(msg.Width, msg.Height)
		}

	case errMsg:
		m.err = msg.error
		m.state = StateError
		return m, nil

	case filesDiscoveredMsg:
		m.allFiles = msg.allFiles
		m.docFiles, m.codeFiles = classifyFiles(m.allFiles)

		// Prepare items for the list
		items := make([]list.Item, len(m.docFiles))
		for i, f := range m.docFiles {
			items[i] = f
		}
		m.list.SetItems(items)

		// Set up delegate styles (only need to do once unless we change delegates)
		delegate := list.NewDefaultDelegate()
		delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("205")).Bold(true)
		delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("205"))
		m.list.SetDelegate(delegate)

		m.list.Title = fmt.Sprintf("Review Documentation Files (Found %d docs, %d code)", len(m.docFiles), len(m.codeFiles))
		m.list.SetStatusBarItemName("file", "files")

		m.state = StateDiscovery
		m.loadingText = "" // Clear loading text
		return m, nil
	}

	// Route updates based on state
	if m.state == StateDiscovery {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.state == StateAddingFile {
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI.
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	switch m.state {
	case StateLoading:
		s := strings.Builder{}
		s.WriteString(HeaderStyle.Width(m.width).Render("Drift Init Wizard"))
		s.WriteString("\n\n")
		s.WriteString(fmt.Sprintf("%s %s\n", m.spinner.View(), m.loadingText))
		return AppStyle.Render(s.String())

	case StateDiscovery:
		return AppStyle.Render(m.list.View())

	case StateAddingFile:
		return fmt.Sprintf(
			"Add a file manually:\n\n%s\n\n%s\n(Esc to cancel, Enter to add, Tab to complete)",
			m.textInput.View(),
			m.feedbackMsg,
		)

	case StateGrouping:
		s := strings.Builder{}
		s.WriteString(HeaderStyle.Width(m.width).Render("Drift Init Wizard"))
		s.WriteString("\n\n")
		s.WriteString("Grouping files...\n")
		return AppStyle.Render(s.String())

	case StateMapping:
		s := strings.Builder{}
		s.WriteString(HeaderStyle.Width(m.width).Render("Drift Init Wizard"))
		s.WriteString("\n\n")
		s.WriteString("Mapping files...\n")
		return AppStyle.Render(s.String())

	case StateFinalize:
		s := strings.Builder{}
		s.WriteString(HeaderStyle.Width(m.width).Render("Drift Init Wizard"))
		s.WriteString("\n\n")
		s.WriteString("Finalizing configuration...\n")
		return AppStyle.Render(s.String())

	case StateDone:
		s := strings.Builder{}
		s.WriteString(HeaderStyle.Width(m.width).Render("Drift Init Wizard"))
		s.WriteString("\n\n")
		s.WriteString("Configuration created successfully!\n")
		return AppStyle.Render(s.String())
	}

	return ""
}

// messages for async operations
type filesDiscoveredMsg struct {
	allFiles []FileInfo
}

type errMsg struct {
	error error
}

// discoverFilesCmd performs the initial file discovery.
func discoverFilesCmd() tea.Cmd {
	return func() tea.Msg {
		allFilePaths, err := files.WalkProject()
		if err != nil {
			return errMsg{err}
		}

		var allFiles []FileInfo
		for _, path := range allFilePaths {
			info, err := os.Stat(path)
			if err != nil {
				// Log error and skip file, or handle as needed
				continue
			}
			allFiles = append(allFiles, FileInfo{
				Path: path,
				Size: info.Size(),
				Type: TypeOther, // Default, will be classified later
			})
		}
		return filesDiscoveredMsg{allFiles: allFiles}
	}
}

// classifyFiles classifies the discovered files into docFiles and codeFiles.
// This is a basic implementation and will be expanded.
func classifyFiles(allFiles []FileInfo) ([]FileInfo, []FileInfo) {
	docFiles := []FileInfo{}
	codeFiles := []FileInfo{}

	docExtensions := map[string]bool{
		".md":   true,
		".mdx":  true,
		".txt":  true,
		".rst":  true,
		".adoc": true,
	}

	codeExtensions := map[string]bool{
		".go":     true,
		".js":     true,
		".ts":     true,
		".jsx":    true,
		".tsx":    true,
		".py":     true,
		".java":   true,
		".kt":     true,
		".c":      true,
		".cpp":    true,
		".h":      true,
		".hpp":    true,
		".rb":     true,
		".php":    true,
		".cs":     true,
		".sh":     true,
		".yaml":   true, // often used for config/code-like declarations
		".yml":    true,
		".json":   true,
		".xml":    true,
		".html":   true,
		".css":    true,
		".scss":   true,
		".less":   true,
		".vue":    true,
		".svelte": true,
		".swift":  true,
		".rs":     true,
		".sql":    true,
	}

	for _, file := range allFiles {
		dotIndex := strings.LastIndex(file.Path, ".")
		var ext string
		if dotIndex != -1 {
			ext = strings.ToLower(file.Path[dotIndex:])
		} else {
			ext = "" // No extension
		}
		if docExtensions[ext] {
			file.Type = TypeDoc
			docFiles = append(docFiles, file)
		} else if codeExtensions[ext] {
			file.Type = TypeCode
			codeFiles = append(codeFiles, file)
		}
	}

	return docFiles, codeFiles
}
