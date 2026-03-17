package llm

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewAnthropicClient(t *testing.T) {
	os.Unsetenv("ANTHROPIC_API_KEY")
	_, err := NewAnthropicClient()
	if err == nil {
		t.Error("expected error when ANTHROPIC_API_KEY is not set")
	}

	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	defer os.Unsetenv("ANTHROPIC_API_KEY")
	client, err := NewAnthropicClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client to be constructed")
	}
}

func TestAnthropicClient_GenerateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if string(b) == "" {
			t.Error("expected request body")
		}
		
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("missing or incorrect api key header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"content": [
				{
					"text": "{\"is_in_sync\": true}"
				}
			],
			"usage": {
				"input_tokens": 5,
				"output_tokens": 10
			}
		}`))
	}))
	defer server.Close()

	os.Setenv("ANTHROPIC_BASE_URL", server.URL)
	defer os.Unsetenv("ANTHROPIC_BASE_URL")

	client := &AnthropicClient{apiKey: "test-key"}

	res, usage, err := client.GenerateJSON(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}

	if usage.TotalTokens != 15 {
		t.Errorf("expected 15 total tokens, got %d", usage.TotalTokens)
	}
	if !strings.Contains(res, "is_in_sync") {
		t.Errorf("unexpected response: %s", res)
	}
}
