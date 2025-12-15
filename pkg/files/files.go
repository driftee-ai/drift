package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	ignore "github.com/sabhiram/go-gitignore"
)

// WalkProject walks the current directory and returns a list of all files,
// ignoring common non-project directories (like .git, node_modules) and
// respecting .gitignore if present.
func WalkProject() ([]string, error) {
	var files []string
	ignoredDirs := map[string]bool{
		".git":         true,
		".idea":        true,
		".vscode":      true,
		"node_modules": true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		".next":        true,
		"target":       true,
		"bin":          true,
		"obj":          true,
	}

	// Attempt to load .gitignore
	var gitIgnore *ignore.GitIgnore
	if _, err := os.Stat(".gitignore"); err == nil {
		gitIgnore, _ = ignore.CompileIgnoreFile(".gitignore")
	}

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip current directory "."
		if path == "." {
			return nil
		}

		// Check hardcoded ignores
		if d.IsDir() && ignoredDirs[d.Name()] {
			return filepath.SkipDir
		}

		// Check .gitignore
		if gitIgnore != nil && gitIgnore.MatchesPath(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

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
