package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutocomplete(t *testing.T) {
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
			input:          "cmd/r",
			expectedOutput: "cmd/root.go",
			description:    "File in directory",
		},
		{
			input:          "pkg",
			expectedOutput: "pkg/tui/wizard.go",
			description:    "Deep file match",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			result, _ := Autocomplete(tc.input, mockFiles)
			assert.Equal(t, tc.expectedOutput, result)
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
			},
			wantDocCount:  0,
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
