package llm

import (
	"fmt"
)

// New creates a new Client based on the provided provider name.
func New(provider string) (Client, error) {
	switch provider {
	case "gemini":
		return NewGeminiClient()
	case "openai":
		return NewOpenAIClient()
	case "anthropic":
		return NewAnthropicClient()
	case "dummy":
		return NewDummyClient(), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
