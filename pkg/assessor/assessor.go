package assessor

// AssessmentResult holds the result of a drift assessment.
type AssessmentResult struct {
	IsInSync bool   `json:"is_in_sync"`
	Reason   string `json:"reason"`
}

// DocAssessor is the interface for assessing drift between code and documentation.
type DocAssessor interface {
	Assess(docContent string, codeContents map[string]string) (*AssessmentResult, error)
}

// DummyAssessor is a mock assessor for testing purposes.
type DummyAssessor struct{}

// NewDummyAssessor creates a new DummyAssessor.
func NewDummyAssessor() *DummyAssessor {
	return &DummyAssessor{}
}

// Assess returns a hardcoded assessment result.
func (a *DummyAssessor) Assess(docContent string, codeContents map[string]string) (*AssessmentResult, error) {
	return &AssessmentResult{
		IsInSync: true,
		Reason:   "This is a dummy assessment.",
	}, nil
}
