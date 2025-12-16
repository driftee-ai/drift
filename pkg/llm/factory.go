package llm

import "fmt"

// New creates a new Generator based on the provider name.
func New(provider string) (Generator, error) {
	switch provider {
	case "gemini":
		return NewGeminiGenerator()
	case "openai":
		return NewOpenAIGenerator()
	case "dummy":
		return NewDummyGenerator(), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
