package assessor

import (
	"fmt"
)

// New creates a new DocAssessor based on the provided provider name.
func New(provider string) (DocAssessor, error) {
	switch provider {
	case "gemini":
		return NewGeminiAssessor()
	case "openai":
		return NewOpenAIAssessor()
	case "dummy":
		return NewDummyAssessor(), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
