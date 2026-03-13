package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// AnthropicClient uses the raw Anthropic API.
type AnthropicClient struct {
	apiKey string
}

// NewAnthropicClient creates a new Anthropic Client checking credentials.
func NewAnthropicClient() (*AnthropicClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}
	return &AnthropicClient{apiKey: apiKey}, nil
}

type anthropicRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	System      string    `json:"system"`
	Messages    []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// GenerateJSON constructs a direct HTTP request to the Anthropic Messages API.
func (c *AnthropicClient) GenerateJSON(ctx context.Context, prompt string, schema interface{}) (string, error) {
	// Anthropic needs to be prompted explicitly to return only JSON
	sysPrompt := fmt.Sprintf("You are a deterministic parsing system. You must output exclusively valid JSON matching the following abstract schema structure without any markdown tags or conversation: %+v", schema)
	
	reqBody := anthropicRequest{
		Model:     "claude-3-haiku-20240307",
		MaxTokens: 4096,
		System:    sysPrompt,
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonValue, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic rest failure: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic error response (status %d): %s", resp.StatusCode, string(body))
	}

	var response anthropicResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal anthropic response: %w", err)
	}

	if len(response.Content) > 0 {
		return response.Content[0].Text, nil
	}

	return "", fmt.Errorf("empty content response from Anthropic API")
}
