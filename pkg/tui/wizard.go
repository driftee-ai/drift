package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/driftee-ai/drift/pkg/files"
)

// State represents the current step of the wizard.
type State int

const (
	StateLoading State = iota
	StateDiscovery
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
	Path        string
	Type        DocFileType
	Size        int64
	IsSelected  bool   // For interactive selection
	IsIgnored   bool   // For interactive ignore
	Description string // LLM generated description
}

// Model is the main Bubble Tea model for the init wizard.
type Model struct {
	state       State
	spinner     spinner.Model
	err         error
	width       int
	height      int
	allFiles    []FileInfo
	docFiles    []FileInfo
	codeFiles   []FileInfo
	loadingText string
}

// NewModel creates a new Model for the init wizard.
func NewModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(HighlightColor)
	return Model{
		state:       StateLoading,
		spinner:     s,
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
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case errMsg:
		m.err = msg.error
		m.state = StateError
		return m, nil

	case filesDiscoveredMsg:
		m.allFiles = msg.allFiles
		m.docFiles, m.codeFiles = classifyFiles(m.allFiles)
		m.state = StateDiscovery
		m.loadingText = "" // Clear loading text
		return m, nil
	}

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the TUI.
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	s := strings.Builder{}
	s.WriteString(HeaderStyle.Width(m.width).Render("Drift Init Wizard"))
	s.WriteString("\n\n")

	switch m.state {
	case StateLoading:
		s.WriteString(fmt.Sprintf("%s %s\n", m.spinner.View(), m.loadingText))
	case StateDiscovery:
		s.WriteString(fmt.Sprintf("Discovery: Found %d doc files and %d code files.\n\n", len(m.docFiles), len(m.codeFiles)))
		s.WriteString("Doc Files:\n")
		for i, file := range m.docFiles {
			s.WriteString(fmt.Sprintf("  %d: %s\n", i+1, file.Path))
		}
		s.WriteString("\nCode Files:\n")
		for i, file := range m.codeFiles {
			s.WriteString(fmt.Sprintf("  %d: %s\n", i+1, file.Path))
		}
		s.WriteString("\nPress 'q' or 'ctrl+c' to quit.\n")

	case StateGrouping:
		s.WriteString("Grouping files...\n")
	case StateMapping:
		s.WriteString("Mapping files...\n")
	case StateFinalize:
		s.WriteString("Finalizing configuration...\n")
	case StateDone:
		s.WriteString("Configuration created successfully!\n")
	}

	s.WriteString("\n")
	return AppStyle.Render(s.String())
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
