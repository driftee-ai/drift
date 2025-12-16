package llm

import (
	"context"
	"encoding/json"
)

// DummyGenerator is a mock generator.
type DummyGenerator struct{}

// NewDummyGenerator creates a new DummyGenerator.
func NewDummyGenerator() *DummyGenerator {
	return &DummyGenerator{}
}

// Generate returns a fixed string.
func (g *DummyGenerator) Generate(ctx context.Context, prompt string) (string, error) {
	return "This is a dummy response.", nil
}

// GenerateJSON returns a fixed JSON structure.
// It assumes the result is compatible with the dummy data.
func (g *DummyGenerator) GenerateJSON(ctx context.Context, prompt string, schema *Schema, result any) error {
	// For now, we try to unmarshal a simple success case.
	// If the caller expects something else, this might fail, but for drift it expects {is_in_sync: true, reason: ...}
	// We'll return a generic JSON that fits that.
	data := `{"is_in_sync": true, "reason": "This is a dummy assessment."}`
	return json.Unmarshal([]byte(data), result)
}
