package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNavigation(t *testing.T) {
	model := NewModel()

	// Initial state
	assert.Equal(t, 0, model.pageIndex)
	// Send WindowSizeMsg to initialize currentPage
	res, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = res.(MainModel)

	// Initial state
	assert.Equal(t, 0, model.pageIndex)
	assert.NotNil(t, model.currentPage)

	// Move Next
	res, cmd := model.Update(NextStepMsg{})
	_ = cmd // ignore cmd
	model = res.(MainModel)
	assert.Equal(t, 1, model.pageIndex)

	// Move Next again
	res, _ = model.Update(NextStepMsg{})
	model = res.(MainModel)
	assert.Equal(t, 2, model.pageIndex)

	// Move Back
	res, _ = model.Update(PrevStepMsg{})
	model = res.(MainModel)
	assert.Equal(t, 1, model.pageIndex)

	// Move Back to start
	res, _ = model.Update(PrevStepMsg{})
	model = res.(MainModel)
	assert.Equal(t, 0, model.pageIndex)

	// Try moving back from start (should verify no change)
	res, _ = model.Update(PrevStepMsg{})
	model = res.(MainModel)
	assert.Equal(t, 0, model.pageIndex)
}
