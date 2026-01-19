package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/driftee-ai/drift/pkg/config"
)

// Session holds the shared state for the wizard.
type Session struct {
	AllFiles  []FileInfo
	DocFiles  []FileInfo
	CodeFiles []FileInfo
	Groups    []config.Rule
	Provider  string
	Config    config.Config
}

// Page is a sub-model that represents a single step in the wizard.
type Page interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Page, tea.Cmd)
	View() string
	Keys() []key.Binding
	Title() string
	Step() string // e.g. "Step 1/5"
}

// Common messages
type NextStepMsg struct{}
type PrevStepMsg struct{}
type ErrorMsg struct{ Err error }
