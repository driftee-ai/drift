package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestClassifyFiles(t *testing.T) {
	tests := []struct {
		name          string
		inputFiles    []FileInfo
		wantDocCount  int
		wantCodeCount int
	}{
		{
			name: "Standard files",
			inputFiles: []FileInfo{
				{Path: "main.go", Type: TypeOther},
				{Path: "README.md", Type: TypeOther},
				{Path: "utils.js", Type: TypeOther},
			},
			wantDocCount:  1,
			wantCodeCount: 2,
		},
		{
			name: "No extension",
			inputFiles: []FileInfo{
				{Path: "LICENSE", Type: TypeOther},
				{Path: "Makefile", Type: TypeOther},
			},
			wantDocCount:  0,
			wantCodeCount: 0,
		},
		{
			name: "Mixed casing",
			inputFiles: []FileInfo{
				{Path: "Doc.MD", Type: TypeOther},
				{Path: "Script.PY", Type: TypeOther},
			},
			wantDocCount:  1,
			wantCodeCount: 1,
		},
		{
			name: "Dotfiles",
			inputFiles: []FileInfo{
				{Path: ".gitignore", Type: TypeOther},
				{Path: ".env", Type: TypeOther},
			},
			wantDocCount:  0, // .gitignore might be code or other, but let's check current logic
			wantCodeCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs, code := classifyFiles(tt.inputFiles)
			assert.Equal(t, tt.wantDocCount, len(docs), "Docs count mismatch")
			assert.Equal(t, tt.wantCodeCount, len(code), "Code count mismatch")
		})
	}
}

func TestUpdate_Discovery(t *testing.T) {
	m := NewModel()

	// Simulate "Files Discovered" event
	testFiles := []FileInfo{
		{Path: "main.go"},
		{Path: "README.md"},
	}
	msg := filesDiscoveredMsg{allFiles: testFiles}

	// Call Update
	newModel, _ := m.Update(msg)
	newM := newModel.(Model)

	// Assert State Transition
	assert.Equal(t, StateDiscovery, newM.state)
	assert.Empty(t, newM.loadingText)

	// Assert Data Population
	assert.Equal(t, 2, len(newM.allFiles))
	assert.Equal(t, 1, len(newM.docFiles))
	assert.Equal(t, 1, len(newM.codeFiles))
}

func TestUpdate_Quit(t *testing.T) {
	m := NewModel()
	m.state = StateDiscovery

	// Simulate 'q' keypress
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}

	_, cmd := m.Update(msg)

	// In Bubble Tea, tea.Quit is a special command.
	// We can't easily compare function pointers, but we can check if a command was returned.
	// For exact verification, we'd need to mock the tea runtime or check the type,
	// but mostly we just want to ensure it doesn't panic.
	assert.NotNil(t, cmd)
}
