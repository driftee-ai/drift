package files_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/driftee-ai/drift/pkg/files"
)

// setupTestFiles creates a temporary directory and some dummy files for testing.
// It returns the path to the temporary directory and a cleanup function.
func setupTestFiles(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "files_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create dummy files
	if err := os.MkdirAll(filepath.Join(tmpDir, "src", "api"), 0755); err != nil {
		t.Fatalf("Failed to create src/api dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "src", "api", "user.go"), []byte("package api\n// User struct"), 0644); err != nil {
		t.Fatalf("Failed to write user.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "src", "api", "auth.go"), []byte("package api\n// Auth struct"), 0644); err != nil {
		t.Fatalf("Failed to write auth.go: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "docs", "api"), 0755); err != nil {
		t.Fatalf("Failed to create docs/api dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "docs", "api", "users.md"), []byte("# Users API"), 0644); err != nil {
		t.Fatalf("Failed to write users.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "docs", "api", "auth.md"), []byte("# Auth API"), 0644); err != nil {
		t.Fatalf("Failed to write auth.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Project README"), 0644); err != nil {
		t.Fatalf("Failed to write README.md: %v", err)
	}

	// Change to the temporary directory for globbing to work relative to it
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", tmpDir, err)
	}

	cleanup := func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestFindFiles(t *testing.T) {
	_, cleanup := setupTestFiles(t)
	defer cleanup()

	tests := []struct {
		name     string
		patterns []string
		want     []string
	}{
		{
			name:     "single glob pattern",
			patterns: []string{"src/api/*.go"},
			want:     []string{"src/api/auth.go", "src/api/user.go"}, // Order might vary
		},
		{
			name:     "double star glob pattern",
			patterns: []string{"**/*.md"},
			want:     []string{"docs/api/auth.md", "docs/api/users.md", "README.md"}, // Order might vary
		},
		{
			name:     "multiple patterns",
			patterns: []string{"src/api/*.go", "docs/api/*.md"},
			want:     []string{"src/api/auth.go", "src/api/user.go", "docs/api/auth.md", "docs/api/users.md"}, // Order might vary
		},
		{
			name:     "no matching files",
			patterns: []string{"nonexistent/*.txt"},
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := files.FindFiles(tt.patterns)
			if err != nil {
				t.Fatalf("FindFiles() error = %v", err)
			}

			// Sort both slices to ensure consistent order for comparison
			sort.Strings(got)
			sort.Strings(tt.want)

			if !compareStringSlices(got, tt.want) {
				t.Errorf("FindFiles() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadAndConcatenate(t *testing.T) {
	_, cleanup := setupTestFiles(t)
	defer cleanup()

	// Paths are relative to the temporary directory
	paths := []string{
		filepath.Join("src", "api", "user.go"),
		filepath.Join("docs", "api", "users.md"),
	}

	expectedContent := `package api
// User struct
--- End of file: src/api/user.go ---
# Users API
--- End of file: docs/api/users.md ---
`
	got, err := files.ReadAndConcatenate(paths)
	if err != nil {
		t.Fatalf("ReadAndConcatenate() error = %v", err)
	}

	if got != expectedContent {
		t.Errorf("ReadAndConcatenate() got = %q, want %q", got, expectedContent)
	}
}

// Helper to compare string slices (order-independent)
func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
