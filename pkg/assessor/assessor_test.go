package assessor_test

import (
	"testing"

	"github.com/driftee-ai/drift/pkg/assessor"
)

func TestNewAssessor(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
	}{
		{
			name:     "Gemini provider - no api key",
			provider: "gemini",
			wantErr:  true,
		},
		{
			name:     "Dummy provider",
			provider: "dummy",
			wantErr:  false,
		},
		{
			name:     "Unknown provider",
			provider: "unknown",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Gemini provider - no api key" {
				t.Setenv("GEMINI_API_KEY", "")
			}

			got, err := assessor.New(tt.provider)

			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Errorf("New() returned nil, want non-nil")
			}
		})
	}
}
