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
		wantType interface{} // Expected type of the returned assessor
	}{
		{
			name:     "Gemini provider",
			provider: "gemini",
			wantErr:  true, // Expect error because GEMINI_API_KEY is not set
			wantType: &assessor.GeminiAssessor{},
		},
		{
			name:     "Dummy provider",
			provider: "dummy",
			wantErr:  false,
			wantType: &assessor.DummyAssessor{},
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
			got, err := assessor.New(tt.provider)

			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantType != nil {
				// Check if the returned assessor is of the expected type
				// This is a basic type assertion, more robust checks might be needed for interfaces
				if _, ok := tt.wantType.(*assessor.GeminiAssessor); ok {
					if _, ok := got.(*assessor.GeminiAssessor); !ok {
						t.Errorf("New() got = %T, want %T", got, tt.wantType)
					}
				} else if _, ok := tt.wantType.(*assessor.DummyAssessor); ok {
					if _, ok := got.(*assessor.DummyAssessor); !ok {
						t.Errorf("New() got = %T, want %T", got, tt.wantType)
					}
				}
			}
		})
	}
}
