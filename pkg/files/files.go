package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// FindFiles takes a base directory and a list of glob patterns and returns a list of matching file paths relative to the baseDir.
func FindFiles(baseDir string, patterns []string) ([]string, error) {
	var matchingFiles []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// doublestar.Glob walks the file system and returns matching files
		matches, err := doublestar.Glob(os.DirFS(baseDir), pattern)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			// We need to return the full absolute or relative path from process CWD, so we join with baseDir
			fullPath := filepath.Join(baseDir, match)

			// Ensure it's a file and not a directory
			info, err := os.Stat(fullPath)
			if err != nil {
				// If file doesn't exist or other error, skip
				continue
			}
			if info.IsDir() {
				continue
			}

			// Add to list if not already seen
			if !seen[fullPath] {
				matchingFiles = append(matchingFiles, fullPath)
				seen[fullPath] = true
			}
		}
	}
	return matchingFiles, nil
}

// ReadAndConcatenate takes a list of file paths, reads each file, and returns a single string with all the content.s
func ReadAndConcatenate(paths []string) (string, error) {
	var builder strings.Builder
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", path, err)
		}
		builder.WriteString(string(content))
		builder.WriteString("\n--- End of file: ")
		builder.WriteString(path)
		builder.WriteString(" ---\n")
	}
	return builder.String(), nil
}

// ReadFiles takes a list of file paths and returns a map of file paths to their contents.
func ReadFiles(paths []string) (map[string]string, error) {
	contents := make(map[string]string)
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", path, err)
		}
		contents[path] = string(content)
	}
	return contents, nil
}
