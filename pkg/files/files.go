package files

import (
	"fmt"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// FindFiles takes a list of glob patterns and returns a list of matching file paths.
func FindFiles(patterns []string) ([]string, error) {
	var matchingFiles []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		// doublestar.Glob walks the file system and returns matching files
		// Use os.DirFS(".") to glob the current directory
		matches, err := doublestar.Glob(os.DirFS("."), pattern)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			// doublestar.Glob returns paths relative to the root of the FS (os.DirFS(".")).
			// We need to prepend the current directory to make them absolute or relative to the project root.
			// For now, let's assume the patterns are relative to the project root.
			// The match is already relative to the current directory.

			// Ensure it's a file and not a directory
			info, err := os.Stat(match)
			if err != nil {
				// If file doesn't exist or other error, skip
				continue
			}
			if info.IsDir() {
				continue
			}

			// Add to list if not already seen
			if !seen[match] {
				matchingFiles = append(matchingFiles, match)
				seen[match] = true
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
