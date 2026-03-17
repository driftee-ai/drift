package llm

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestNewOpenAIClient(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	_, err := NewOpenAIClient()
	if err == nil {
		t.Error("expected error when OPENAI_API_KEY is not set")
	}

	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")
	client, err := NewOpenAIClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client to be constructed")
	}
}

func TestOpenAIClient_GenerateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if len(b) == 0 {
			t.Error("expected request body")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-3.5-turbo-0613",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "{\"is_in_sync\": true}"
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 5,
				"total_tokens": 15
			}
		}`))
	}))
	defer server.Close()

	config := openai.DefaultConfig("test-key")
	config.BaseURL = server.URL + "/v1"
	client := openai.NewClientWithConfig(config)
	
	llmClient := &OpenAIClient{client: client}

	res, usage, err := llmClient.GenerateJSON(context.Background(), "test prompt", nil)
	if err != nil {
		t.Fatalf("GenerateJSON failed: %v", err)
	}
	
	if usage.TotalTokens != 15 {
		t.Errorf("expected 15 total tokens, got %d", usage.TotalTokens)
	}
	if res != `{"is_in_sync": true}` {
		t.Errorf("unexpected response: %s", res)
	}
}
