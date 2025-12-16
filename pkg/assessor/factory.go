package assessor

import (
	"fmt"

	"github.com/driftee-ai/drift/pkg/llm"
)

// New creates a new DocAssessor based on the provided provider name.
func New(provider string) (DocAssessor, error) {
	generator, err := llm.New(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM generator: %w", err)
	}
	return NewDriftAssessor(generator), nil
}
