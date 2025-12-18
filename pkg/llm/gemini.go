package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiGenerator uses the Gemini API.
type GeminiGenerator struct {
	client *genai.GenerativeModel
}

// NewGeminiGenerator creates a new GeminiGenerator.
func NewGeminiGenerator() (*GeminiGenerator, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-3-flash-preview")
	return &GeminiGenerator{client: model}, nil
}

// Generate generates text content.
func (g *GeminiGenerator) Generate(ctx context.Context, prompt string) (string, error) {
	g.client.ResponseMIMEType = "text/plain"
	g.client.ResponseSchema = nil

	resp, err := g.client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0]), nil
	}

	return "", fmt.Errorf("no content generated")
}

// GenerateJSON generates a JSON response.
func (g *GeminiGenerator) GenerateJSON(ctx context.Context, prompt string, schema *Schema, result any) error {
	g.client.ResponseMIMEType = "application/json"
	g.client.ResponseSchema = toGenaiSchema(schema)

	resp, err := g.client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) > 0 {
		content := resp.Candidates[0].Content
		if len(content.Parts) > 0 {
			jsonStr := fmt.Sprintf("%s", content.Parts[0])
			if err := json.Unmarshal([]byte(jsonStr), result); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("no content generated")
}

func toGenaiSchema(s *Schema) *genai.Schema {
	if s == nil {
		return nil
	}

	gs := &genai.Schema{
		Type:        mapType(s.Type),
		Description: s.Description,
		Required:    s.Required,
	}

	if s.Properties != nil {
		gs.Properties = make(map[string]*genai.Schema)
		for k, v := range s.Properties {
			gs.Properties[k] = toGenaiSchema(v)
		}
	}

	if s.Items != nil {
		gs.Items = toGenaiSchema(s.Items)
	}

	return gs
}

func mapType(t SchemaType) genai.Type {
	switch t {
	case TypeObject:
		return genai.TypeObject
	case TypeArray:
		return genai.TypeArray
	case TypeString:
		return genai.TypeString
	case TypeBoolean:
		return genai.TypeBoolean
	case TypeInteger:
		return genai.TypeInteger
	case TypeNumber:
		return genai.TypeNumber
	default:
		return genai.TypeString // Default to string if unknown
	}
}
