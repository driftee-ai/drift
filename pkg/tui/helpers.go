package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/driftee-ai/drift/pkg/files"
)

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

// Autocomplete returns the completed path and a feedback message/suggestion.
func Autocomplete(input string, allFiles []FileInfo) (string, string) {
	if input == "." || input == "./" {
		return input, ""
	}

	cleanInput := strings.TrimPrefix(input, "./")
	var matches []string
	for _, f := range allFiles {
		if strings.HasPrefix(f.Path, cleanInput) {
			matches = append(matches, f.Path)
		}
	}

	if len(matches) == 1 {
		return matches[0], ""
	} else if len(matches) > 1 {
		// Find common prefix
		common := matches[0]
		for _, match := range matches[1:] {
			for !strings.HasPrefix(match, common) {
				common = common[:len(common)-1]
			}
		}

		msg := "Matches: " + strings.Join(matches, ", ")
		if len(matches) > 3 {
			msg = fmt.Sprintf("Matches: %s... (%d more)", strings.Join(matches[:3], ", "), len(matches)-3)
		}
		return common, msg
	}

	return input, ""
}
