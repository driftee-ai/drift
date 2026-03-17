package llm

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type rewriteTransport struct {
	transport http.RoundTripper
	serverURL string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.serverURL, "http://")
	return t.transport.RoundTrip(req)
}

func TestNewGeminiClient(t *testing.T) {
	os.Unsetenv("GEMINI_API_KEY")
	_, err := NewGeminiClient()
	if err == nil {
		t.Error("expected error when GEMINI_API_KEY is not set")
	}

	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")
	client, err := NewGeminiClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Error("expected client to be constructed")
	}
}

func TestGeminiClient_GenerateJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if string(b) == "" {
			t.Error("expected request body")
		}
		w.Header().Set("Content-Type", "application/json")
		// Mock response structure matching Gemini REST API
		w.Write([]byte(`{
			"candidates": [
				{
					"content": {
						"parts": [
							{"text": "{\"is_in_sync\": true}"}
						]
					}
				}
			],
			"usageMetadata": {
				"promptTokenCount": 5,
				"candidatesTokenCount": 10,
				"totalTokenCount": 15
			}
		}`))
	}))
	defer server.Close()

	mockHTTPClient := &http.Client{
		Transport: &rewriteTransport{
			transport: http.DefaultTransport,
			serverURL: server.URL,
		},
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("test-key"), option.WithHTTPClient(mockHTTPClient), option.WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("Failed to create genai client: %v", err)
	}

	model := client.GenerativeModel("gemini-2.5-flash")
	model.ResponseMIMEType = "application/json"

	llmClient := &GeminiClient{client: model}

	// Make sure we pass a genai.Schema or a generic struct to test GenerateJSON schema attachment logic
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"is_in_sync": {Type: genai.TypeBoolean},
		},
	}

	res, usage, err := llmClient.GenerateJSON(context.Background(), "test prompt", schema)
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
