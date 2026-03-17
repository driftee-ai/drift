package checker

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/llm"
)

// MockClient allows us to hardcode JSON responses for testing
type MockClient struct {
	Response    string
	ShouldError bool
}

func (m *MockClient) GenerateJSON(ctx context.Context, prompt string, schema interface{}) (string, llm.Usage, error) {
	if m.ShouldError {
		return "", llm.Usage{}, fmt.Errorf("mock error")
	}
	return m.Response, llm.Usage{}, nil
}

func TestEvaluateRules(t *testing.T) {
	rule := config.Rule{
		Name: "Test Rule",
		Code: []string{filepath.Join("testdata", "code", "*.go")},
		Docs: []string{filepath.Join("testdata", "docs", "*.md")},
	}
	missingBothRule := config.Rule{
		Name: "Missing Both",
		Code: []string{filepath.Join("testdata", "code", "nonexistent.go")},
		Docs: []string{filepath.Join("testdata", "docs", "nonexistent.md")},
	}
	missingCodeRule := config.Rule{
		Name: "Missing Code Only",
		Code: []string{filepath.Join("testdata", "code", "nonexistent.go")},
		Docs: []string{filepath.Join("testdata", "docs", "*.md")},
	}
	missingDocsRule := config.Rule{
		Name: "Missing Docs Only",
		Code: []string{filepath.Join("testdata", "code", "*.go")},
		Docs: []string{filepath.Join("testdata", "docs", "nonexistent.md")},
	}

	tests := []struct {
		name           string
		rule           config.Rule
		mockResponse   string
		mockError      bool
		diffOnly       bool
		diffContext    string
		expectedInSync bool
		expectedReason string
		expectError    bool
		expectSkipped  bool
		expectIgnored  bool
	}{
		{
			name:           "In Sync",
			rule:           rule,
			mockResponse:   "```json\n{\n  \"is_in_sync\": true,\n  \"reason\": \"Matches perfectly\"\n}\n```",
			expectedInSync: true,
			expectedReason: "Matches perfectly",
		},
		{
			name:           "Out of Sync",
			rule:           rule,
			mockResponse:   "```json\n{\n  \"is_in_sync\": false,\n  \"reason\": \"Missing parameter\"\n}\n```",
			expectedInSync: false,
			expectedReason: "Missing parameter",
		},
		{
			name:           "Malformed JSON",
			rule:           rule,
			mockResponse:   "Not a json response",
			expectedInSync: false,
			expectError:    true,
		},
		{
			name:           "LLM Error",
			rule:           rule,
			mockError:      true,
			expectedInSync: false,
			expectError:    true,
		},
		{
			name:           "Diff Ignored",
			rule:           rule,
			diffOnly:       true,
			diffContext:    "--- a/file +++ b/file",
			mockResponse:   "```json\n{\n  \"is_in_sync\": false,\n  \"reason\": \"Legacy drift\",\n  \"is_drift_caused_by_diff\": false\n}\n```",
			expectedInSync: true,
			expectIgnored:  true,
		},
		{
			name:           "Diff Causes Drift",
			rule:           rule,
			diffOnly:       true,
			diffContext:    "--- a/file +++ b/file",
			mockResponse:   "```json\n{\n  \"is_in_sync\": false,\n  \"reason\": \"Diff broke it\",\n  \"is_drift_caused_by_diff\": true\n}\n```",
			expectedInSync: false,
			expectIgnored:  false,
		},
		{
			name:          "Missing Both - Skipped",
			rule:          missingBothRule,
			expectSkipped: true,
		},
		{
			name:           "Missing Code Only - Error",
			rule:           missingCodeRule,
			expectedInSync: false,
			expectError:    true, // ErrMissingFiles
		},
		{
			name:           "Missing Docs Only - Error",
			rule:           missingDocsRule,
			expectedInSync: false,
			expectError:    true, // ErrMissingFiles
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &MockClient{
				Response:    tt.mockResponse,
				ShouldError: tt.mockError,
			}

			// Test with both providers to ensure schema generation doesn't crash
			for _, provider := range []string{"openai", "gemini"} {
				t.Run(provider, func(t *testing.T) {
					checker := New(client, provider)

					results := checker.EvaluateRules(context.Background(), []config.Rule{tt.rule}, tt.diffOnly, tt.diffContext)
					if len(results) != 1 {
						t.Fatalf("expected 1 result, got %d", len(results))
					}

					res := results[0]

					if tt.expectSkipped {
						if !res.Skipped {
							t.Errorf("expected rule to be skipped")
						}
						return
					}

					if tt.expectError {
						if res.Error == nil {
							t.Errorf("expected an error, but got nil")
						}
						return
					} else if res.Error != nil {
						t.Fatalf("unexpected error: %v", res.Error)
					}

					if res.IsInSync != tt.expectedInSync {
						t.Errorf("expected IsInSync to be %v, got %v\nReason: %s", tt.expectedInSync, res.IsInSync, res.Reason)
					}

					if res.IgnoredDueToDiff != tt.expectIgnored {
						t.Errorf("expected IgnoredDueToDiff to be %v, got %v", tt.expectIgnored, res.IgnoredDueToDiff)
					}

					if tt.expectedReason != "" && res.Reason != tt.expectedReason {
						t.Errorf("expected reasoning %q, got %q", tt.expectedReason, res.Reason)
					}
				})
			}
		})
	}
}
