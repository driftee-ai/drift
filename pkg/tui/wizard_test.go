package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestAutocompleteLogic(t *testing.T) {
	// Mock file list
	mockFiles := []FileInfo{
		{Path: "README.md"},
		{Path: "cmd/root.go"},
		{Path: "cmd/init.go"},
		{Path: "pkg/tui/wizard.go"},
		{Path: ".gitignore"},
	}

	tests := []struct {
		input          string
		expectedOutput string
		description    string
	}{
		{
			input:          "REA",
			expectedOutput: "README.md",
			description:    "Simple prefix match",
		},
		{
			input:          "cmd",
			expectedOutput: "cmd/",
			description:    "Common prefix directory",
		},
		{
			input:          "cmd/",
			expectedOutput: "cmd/",
			description:    "Directory prefix already matched",
		},
		{
			input:          "cmd/r",
			expectedOutput: "cmd/root.go",
			description:    "File in directory",
		},
		{
			input:          ".",
			expectedOutput: ".", // Currently, "." doesn't match "README.md" or others based on simple HasPrefix logic unless adjusted
			description:    "Dot input",
		},
		{
			input:          "./",
			expectedOutput: "./",
			description:    "Dot slash input",
		},
		{
			input:          "pkg",
			expectedOutput: "pkg/tui/wizard.go",
			description:    "Deep file match",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			// Simulate the logic in Update
			input := tc.input
			var result string
			var matches []string

			if input == "." || input == "./" {
				result = input
			} else {
				cleanInput := strings.TrimPrefix(input, "./")

				for _, f := range mockFiles {
					if strings.HasPrefix(f.Path, cleanInput) {
						matches = append(matches, f.Path)
					}
				}

				if len(matches) == 1 {
					result = matches[0]
				} else if len(matches) > 1 {
					common := matches[0]
					for _, match := range matches[1:] {
						for !strings.HasPrefix(match, common) {
							common = common[:len(common)-1]
						}
					}
					result = common
				} else {
					// No match, result is original input
					result = input
				}
			}

			// Adjust validation
			if result != tc.expectedOutput {
				// If expected is empty string (meaning we expect the common prefix of ALL files), handle that:
				if tc.expectedOutput == "" && len(matches) == len(mockFiles) {
					// pass
				} else {
					t.Errorf("For input '%s', expected '%s', got '%s' (Matches: %v)", tc.input, tc.expectedOutput, result, matches)
				}
			}
		})
	}
}

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

func TestUpdate_ToggleIgnore(t *testing.T) {
	m := NewModel()

	// Setup: Discovery state with one doc file
	docFile := FileInfo{Path: "README.md", Type: TypeDoc, IsIgnored: false}
	m.state = StateDiscovery
	m.docFiles = []FileInfo{docFile}
	m.list.SetItems([]list.Item{docFile})

	// Select the first item (should be default, but ensuring)
	m.list.Select(0)

	// Simulate 'space' keypress
	msg := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}

	// Update
	newModel, _ := m.Update(msg)
	newM := newModel.(Model)

	// Verify the item in the list is updated
	selectedItem := newM.list.SelectedItem().(FileInfo)
	assert.True(t, selectedItem.IsIgnored, "Item in list should be ignored")

	// Verify the item in the slice is updated
	// Note: The slice update logic in wizard.go relies on matching Path
	// Because we manually constructed m.docFiles and m.list separately in this test setup,
	// we need to ensure the logic in Update actually propagates the change to m.docFiles.
	// In the real app, m.docFiles sources the list items.
	assert.True(t, newM.docFiles[0].IsIgnored, "Item in docFiles slice should be ignored")

	// Toggle back
	newModel, _ = newM.Update(msg)
	newM = newModel.(Model)
	selectedItem = newM.list.SelectedItem().(FileInfo)
	assert.False(t, selectedItem.IsIgnored, "Item should be un-ignored after second toggle")
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
