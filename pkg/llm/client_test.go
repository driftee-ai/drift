package llm

import (
	"context"
	"testing"
)

func TestDummyClient(t *testing.T) {
	client := NewDummyClient()
	res, usage, err := client.GenerateJSON(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if usage.TotalTokens != 0 {
		t.Errorf("expected 0 usage")
	}
	if res == "" {
		t.Errorf("expected non-empty response")
	}
}
