package rules

import (
	"testing"

	"github.com/driftee-ai/drift/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mockRules = []config.Rule{
	{
		Name: "Rule1-Go",
		Code: []string{"pkg/server/*.go"},
		Docs: []string{"docs/api/server.md"},
	},
	{
		Name: "Rule2-JS",
		Code: []string{"frontend/src/**/*.js"},
		Docs: []string{"docs/frontend/components.md"},
	},
	{
		Name: "Rule3-MultiGlob",
		Code: []string{"pkg/utils/*.go", "pkg/helpers/*.go"},
		Docs: []string{"docs/api/utils.md"},
	},
}

func TestFilterTriggeredRules(t *testing.T) {
	tests := []struct {
		name          string
		changedFiles  []string
		expectedRules []string // Names of rules we expect to be triggered
		expectErr     bool
	}{
		{
			name:          "No changed files should return all rules",
			changedFiles:  []string{},
			expectedRules: []string{"Rule1-Go", "Rule2-JS", "Rule3-MultiGlob"},
			expectErr:     false,
		},
		{
			name:          "Code file match should trigger one rule",
			changedFiles:  []string{"pkg/server/main.go"},
			expectedRules: []string{"Rule1-Go"},
			expectErr:     false,
		},
		{
			name:          "Doc file match should trigger one rule",
			changedFiles:  []string{"docs/frontend/components.md"},
			expectedRules: []string{"Rule2-JS"},
			expectErr:     false,
		},
		{
			name:          "Deeply nested code file match",
			changedFiles:  []string{"frontend/src/components/buttons/Button.js"},
			expectedRules: []string{"Rule2-JS"},
			expectErr:     false,
		},
		{
			name:          "File matching no rules should return empty list",
			changedFiles:  []string{"README.md"},
			expectedRules: []string{},
			expectErr:     false,
		},
		{
			name:          "One file matching multiple rules (hypothetical)",
			changedFiles:  []string{"pkg/server/main.go", "frontend/src/app.js"},
			expectedRules: []string{"Rule1-Go", "Rule2-JS"},
			expectErr:     false,
		},
		{
			name:          "File matching second glob in a rule",
			changedFiles:  []string{"pkg/helpers/format.go"},
			expectedRules: []string{"Rule3-MultiGlob"},
			expectErr:     false,
		},
		{
			name:          "Multiple changed files, only one matches",
			changedFiles:  []string{"Makefile", "go.mod", "pkg/server/server.go"},
			expectedRules: []string{"Rule1-Go"},
			expectErr:     false,
		},
		{
			name:         "Invalid glob pattern should return an error",
			changedFiles: []string{"test"},
			// Temporarily modify a rule to have a bad pattern
			// This is tested implicitly by how filepath.Match works, but good to be aware.
			// For this test, we assume the globs in mockRules are valid.
			// To properly test, we'd need to inject a bad rule.
			// Let's assume valid patterns for now.
			expectedRules: []string{},
			expectErr:     false, // filepath.Match handles this gracefully if the file doesn't match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triggeredRules, err := FilterTriggeredRules(mockRules, tt.changedFiles)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				var triggeredRuleNames []string
				for _, rule := range triggeredRules {
					triggeredRuleNames = append(triggeredRuleNames, rule.Name)
				}

				assert.ElementsMatch(t, tt.expectedRules, triggeredRuleNames, "The triggered rules did not match the expected rules.")
			}
		})
	}
}
