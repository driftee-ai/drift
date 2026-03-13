package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

// OpenAIClient uses the OpenAI API.
type OpenAIClient struct {
	client *openai.Client
}

// NewOpenAIClient creates a generic adapter.
func NewOpenAIClient() (*OpenAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	client := openai.NewClient(apiKey)
	return &OpenAIClient{client: client}, nil
}

// GenerateJSON passes prompts natively executing as JSON.
func (c *OpenAIClient) GenerateJSON(ctx context.Context, prompt string, schema interface{}) (string, error) {
	prompt += fmt.Sprintf("\n\nThe system requires you to output the answer strictly as a JSON object resolving to this schema mapping: %+v", schema)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a senior software engineer that outputs only valid structured JSON matching requested schema inputs.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) > 0 {
		return resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("empty response from OpenAI API")
}
