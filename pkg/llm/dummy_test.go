package llm_test

import (
	"context"
	"testing"

	"github.com/driftee-ai/drift/pkg/llm"
	"github.com/stretchr/testify/assert"
)

func TestDummyGenerator_Generate(t *testing.T) {
	gen := llm.NewDummyGenerator()
	resp, err := gen.Generate(context.Background(), "test prompt")
	assert.NoError(t, err)
	assert.Equal(t, "This is a dummy response.", resp)
}

func TestDummyGenerator_GenerateJSON(t *testing.T) {
	gen := llm.NewDummyGenerator()

	type Result struct {
		IsInSync bool   `json:"is_in_sync"`
		Reason   string `json:"reason"`
	}

	var res Result
	err := gen.GenerateJSON(context.Background(), "test prompt", nil, &res)
	assert.NoError(t, err)
	assert.True(t, res.IsInSync)
	assert.Equal(t, "This is a dummy assessment.", res.Reason)
}
