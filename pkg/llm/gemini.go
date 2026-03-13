package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiClient uses the Gemini API.
type GeminiClient struct {
	client *genai.GenerativeModel
}

// NewGeminiClient reads the Gemini API key from the GEMINI_API_KEY environment variable.
func NewGeminiClient() (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.5-flash")
	model.ResponseMIMEType = "application/json"

	return &GeminiClient{client: model}, nil
}

// GenerateJSON generates content returning it natively as a JSON string format representing the schema.
func (c *GeminiClient) GenerateJSON(ctx context.Context, prompt string, schema interface{}) (string, Usage, error) {
	// Only assign the schema if it perfectly matches genai.Schema
	if s, ok := schema.(*genai.Schema); ok {
		c.client.ResponseSchema = s
	} else {
		// If custom or generic structs are passed, append it directly into the prompt asking Gemini to match it
		prompt += fmt.Sprintf("\n\nRespond with a JSON object strictly matching this schema structure: %+v", schema)
	}

	resp, err := c.client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", Usage{}, fmt.Errorf("failed to generate content: %w", err)
	}

	var usage Usage
	if resp.UsageMetadata != nil {
		usage.PromptTokens = int(resp.UsageMetadata.PromptTokenCount)
		usage.CompletionTokens = int(resp.UsageMetadata.CandidatesTokenCount)
		usage.TotalTokens = int(resp.UsageMetadata.TotalTokenCount)
	}

	if len(resp.Candidates) > 0 {
		content := resp.Candidates[0].Content
		if len(content.Parts) > 0 {
			if textVal, ok := content.Parts[0].(genai.Text); ok {
				return string(textVal), usage, nil
			}
			return fmt.Sprintf("%v", content.Parts[0]), usage, nil
		}
	}

	return "", Usage{}, fmt.Errorf("empty response from Gemini API")
}
