package initwizard

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var defaultExcludes = []string{
	".git",
	"node_modules",
	"vendor",
	"dist",
	"build",
	"out",
	".next",
	".github",
}

// ScanProject finds all code and documentation files.
// In fastMode, it only collects file paths. Otherwise, it also reads file contents (up to 256KB).
func ScanProject(dir string, fastMode bool) (map[string]string, error) {
	files := make(map[string]string)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if path == dir {
				return nil
			}
			for _, exclude := range defaultExcludes {
				if d.Name() == exclude {
					return filepath.SkipDir
				}
			}
			// Skip hidden directories generally
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Check if it's a code or doc file based on extension
		ext := strings.ToLower(filepath.Ext(path))
		isDoc := ext == ".md" || ext == ".mdx" || ext == ".txt" || ext == ".rst"

		// Common programming languages
		isCode := ext == ".go" || ext == ".ts" || ext == ".js" || ext == ".jsx" || ext == ".tsx" || ext == ".py" ||
			ext == ".rb" || ext == ".java" || ext == ".c" || ext == ".cpp" || ext == ".rs" || ext == ".php" ||
			ext == ".cs" || ext == ".swift" || ext == ".kt" || ext == ".scala" || ext == ".sh"

		if isDoc || isCode {
			if fastMode {
				files[path] = ""
			} else {
				info, err := d.Info()
				// Read files smaller than 256KB
				if err == nil && info.Size() < 256*1024 {
					content, err := os.ReadFile(path)
					if err == nil {
						files[path] = string(content)
					}
				} else {
					files[path] = "" // Too large or error getting info, just include path
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan project: %w", err)
	}

	return files, nil
}
