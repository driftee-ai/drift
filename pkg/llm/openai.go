package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

// OpenAIGenerator uses the OpenAI API.
type OpenAIGenerator struct {
	client *openai.Client
}

// NewOpenAIGenerator creates a new OpenAIGenerator.
func NewOpenAIGenerator() (*OpenAIGenerator, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	client := openai.NewClient(apiKey)
	return &OpenAIGenerator{client: client}, nil
}

// Generate generates text content.
func (g *OpenAIGenerator) Generate(ctx context.Context, prompt string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := g.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

// GenerateJSON generates a JSON response.
// Note: OpenAI implementation currently relies on the prompt to define the schema,
// but enforces JSON mode. The schema argument is currently ignored but could be
// used to generate a JSON schema description in the future.
func (g *OpenAIGenerator) GenerateJSON(ctx context.Context, prompt string, schema *Schema, result any) error {
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	}

	resp, err := g.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create chat completion: %w", err)
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
