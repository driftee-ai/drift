package llm_test

import (
	"testing"

	"github.com/driftee-ai/drift/pkg/llm"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		provider    string
		envKey      string
		envValue    string
		expectError bool
		expectType  interface{}
	}{
		{
			name:        "Dummy Provider",
			provider:    "dummy",
			expectError: false,
			expectType:  &llm.DummyGenerator{},
		},
		{
			name:        "Gemini Provider - Success",
			provider:    "gemini",
			envKey:      "GEMINI_API_KEY",
			envValue:    "fake-key",
			expectError: false,
			expectType:  &llm.GeminiGenerator{},
		},
		{
			name:        "Gemini Provider - Missing Key",
			provider:    "gemini",
			envKey:      "GEMINI_API_KEY",
			envValue:    "",
			expectError: true,
			expectType:  nil,
		},
		{
			name:        "OpenAI Provider - Success",
			provider:    "openai",
			envKey:      "OPENAI_API_KEY",
			envValue:    "fake-key",
			expectError: false,
			expectType:  &llm.OpenAIGenerator{},
		},
		{
			name:        "OpenAI Provider - Missing Key",
			provider:    "openai",
			envKey:      "OPENAI_API_KEY",
			envValue:    "",
			expectError: true,
			expectType:  nil,
		},
		{
			name:        "Unknown Provider",
			provider:    "unknown",
			expectError: true,
			expectType:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				t.Setenv(tt.envKey, tt.envValue)
			}

			generator, err := llm.New(tt.provider)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, generator)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tt.expectType, generator)
			}
		})
	}
}
