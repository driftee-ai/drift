package llm_test

import (
	"testing"

	"github.com/driftee-ai/drift/pkg/llm"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
		wantType interface{} // Expected type of the returned client
	}{
		{
			name:     "Gemini provider - no api key",
			provider: "gemini",
			wantErr:  true, // expects GEMINI_API_KEY environment variable not set error
			wantType: nil,
		},
		{
			name:     "Dummy provider",
			provider: "dummy",
			wantErr:  false,
			wantType: &llm.DummyClient{},
		},
		{
			name:     "Unknown provider",
			provider: "unknown",
			wantErr:  true,
			wantType: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Gemini provider - no api key" {
				t.Setenv("GEMINI_API_KEY", "")
			}

			got, err := llm.New(tt.provider)

			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantType != nil {
				// Check if the returned client is of the expected type
				if _, ok := tt.wantType.(*llm.DummyClient); ok {
					if _, ok := got.(*llm.DummyClient); !ok {
						t.Errorf("New() got = %T, want %T", got, tt.wantType)
					}
				}
			}
		})
	}
}
