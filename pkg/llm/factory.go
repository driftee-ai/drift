package llm

import "fmt"

// TestingFactory can be set by tests to override the default generator creation.
var TestingFactory func(string) (Generator, error)

// New creates a new Generator based on the provider name.
func New(provider string) (Generator, error) {
	if TestingFactory != nil {
		return TestingFactory(provider)
	}
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
