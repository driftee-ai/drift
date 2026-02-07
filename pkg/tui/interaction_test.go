package tui

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/llm"
	"github.com/stretchr/testify/assert"
)

// MockGenerator implements llm.Generator for testing
type MockGenerator struct{}

func (m *MockGenerator) Generate(ctx context.Context, prompt string) (string, error) {
	return "mock response", nil
}

func (m *MockGenerator) GenerateJSON(ctx context.Context, prompt string, schema *llm.Schema, result any) error {
	// Crude check to see if we are generating groups or mappings based on the schema or prompt
	// In a real mock we might check the schema structure more strictly.

	// If the prompt mentions "group them into logical features", return groups
	if len(prompt) > 20 && (string(prompt)[:20] == "You are a software a" || len(prompt) > 0) { // Simple heuristic
		// Determine which response to validation
		// The prompt for grouping contains "group them into logical features"
		// The prompt for mapping contains "identify which code files"

		// For Groups
		groupsResp := struct {
			Groups []config.Rule `json:"groups"`
		}{
			Groups: []config.Rule{
				{Name: "Core", Docs: []string{"README.md"}},
			},
		}

		// We can just rely on JSON marshaling to see if it matches the *result type
		// But in Go `result` is `any` (likely a pointer).

		val, _ := json.Marshal(groupsResp)
		if err := json.Unmarshal(val, result); err != nil {
			return err
		}

		// This generic mock is a bit messy. Let's make it smarter based on the prompt.
	}
	return nil
}

// SmartMockGenerator allows defining responses
type SmartMockGenerator struct {
	Groups   []config.Rule
	Mappings []struct {
		Name string   `json:"name"`
		Code []string `json:"code"`
	}
}

func (m *SmartMockGenerator) Generate(ctx context.Context, prompt string) (string, error) {
	return "mock", nil
}

func (m *SmartMockGenerator) GenerateJSON(ctx context.Context, prompt string, schema *llm.Schema, result any) error {
	// Check prompt content to decide response
	// Grouping prompt
	if len(prompt) > 100 && (strings.Contains(prompt, "group them into logical features")) {
		resp := struct {
			Groups []config.Rule `json:"groups"`
		}{Groups: m.Groups}
		data, _ := json.Marshal(resp)
		return json.Unmarshal(data, result)
	}

	// Mapping prompt
	if len(prompt) > 100 && (strings.Contains(prompt, "identify which code files")) {
		resp := struct {
			Mappings []struct {
				Name string   `json:"name"`
				Code []string `json:"code"`
			} `json:"mappings"`
		}{Mappings: m.Mappings}
		data, _ := json.Marshal(resp)
		return json.Unmarshal(data, result)
	}

	return nil
}

func TestWizardInteraction_HappyPath(t *testing.T) {
	// 1. Setup Mock LLM
	mockLLM := &SmartMockGenerator{
		Groups: []config.Rule{{Name: "MockFeature", Docs: []string{"docs/feature.md"}}},
		Mappings: []struct {
			Name string   `json:"name"`
			Code []string `json:"code"`
		}{
			{Name: "MockFeature", Code: []string{"src/feature.go"}},
		},
	}

	// Inject Mock
	llm.TestingFactory = func(provider string) (llm.Generator, error) {
		return mockLLM, nil
	}
	defer func() { llm.TestingFactory = nil }()

	// 2. Setup Temp Dir
	tmpDir, err := os.MkdirTemp("", "drift-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Logf("failed to restore working directory: %v", err)
		}
	}()

	// Create dummy files for discovery
	// Create dummy files for discovery
	if err := os.Mkdir("docs", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("src", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("docs/feature.md", []byte("doc content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("src/feature.go", []byte("code content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("README.md", []byte("readme"), 0644); err != nil {
		t.Fatal(err)
	}

	// 3. Initialize Model
	m := NewModel()

	// Bootstrap Init (triggers discovery)
	// Manually run Init() logic because the test harness update loop handles it slightly differently
	// But NewModel setup is:
	// m.currentPage = m.pages[0]... which happens in Update(WindowSizeMsg)

	// Send WindowSize to start
	model, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = model.(MainModel)

	// Handle Discovery Cmd (async file walk)
	// We need to execute the command returned by DiscoveryPage.Init()
	// Since we can't easily execute tea.Cmd in test without Program, we can cheat:
	// The Init() returns discoverFilesCmd(). We can manually trigger the Msg it produces.
	// OR we can just inject the state directly.

	// Let's manually trigger the file discovery completion
	filesMsg := filesDiscoveredMsg{
		allFiles: []FileInfo{
			{Path: "docs/feature.md", Type: TypeDoc},
			{Path: "src/feature.go", Type: TypeCode},
			{Path: "README.md", Type: TypeDoc},
		},
	}
	model, _ = m.Update(filesMsg)
	m = model.(MainModel)

	// --- Step 1: Discovery ---
	// Assert we are on Discovery Page
	assert.Contains(t, m.currentPage.Title(), "Discovery")

	// Press Enter to proceed (Accept default files)
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(MainModel)

	// Check Transition to Provider
	// The Update loop processes the NextStepMsg internally
	// But `m.Update` returns `cmd` if the page returns `NextStep`.
	// In `wizard.go`:
	// case NextStepMsg: m.pageIndex++ ... return m, m.currentPage.Init()

	// WAIT: `tea.KeyMsg` -> `currentPage.Update` -> returns `NextStepMsg` as a Cmd.
	// `MainModel.Update` receives `tea.KeyMsg`, passes to `currentPage`.
	// `currentPage` returns `(newPage, func() Msg { return NextStepMsg{} })`.
	// `MainModel` batch returns that command.
	// The Bubble Tea runtime would execute that command and feed `NextStepMsg` back to `MainModel.Update`.

	// In this "Manual" test harness, WE must execute the command or manually send the msg.
	// Since `NextStepMsg` is simple, we can just send it.

	model, _ = m.Update(NextStepMsg{})
	m = model.(MainModel)

	// --- Step 2: Provider ---
	assert.Contains(t, m.currentPage.Title(), "Provider")

	// Select "Gemini" (Default selection is index 0) and press Enter
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(MainModel)

	// Trigger Next Step
	model, _ = m.Update(NextStepMsg{})
	m = model.(MainModel)

	// --- Step 3: Grouping ---
	assert.Contains(t, m.currentPage.Title(), "Grouping")

	// Grouping page initializes with `generateGroupsCmd`.
	// We must simulate the result of that command.
	// The Mock LLM logic in this file will be used if we ran the cmd, but we are skipping the cmd execution.
	// So we simulate `groupsGeneratedMsg`.

	// However, to verify the mock hook works, we *could* run the cmd.
	// But tea.Cmd is opaque.
	// Let's just assume the cmd works and send the success msg.
	groupsMsg := groupsGeneratedMsg{groups: mockLLM.Groups}
	model, _ = m.Update(groupsMsg)
	m = model.(MainModel)

	// Verify list is populated
	// We can't easily inspect the internal list model without casting to *GroupingPage, but we can check View() output
	assert.Contains(t, m.View(), "MockFeature")

	// Press Enter to confirm groups
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(MainModel)
	model, _ = m.Update(NextStepMsg{})
	m = model.(MainModel)

	// --- Step 4: Mapping ---
	assert.Contains(t, m.currentPage.Title(), "Code Mapping")

	// Simulate mappings generated
	newGroups := mockLLM.Groups
	newGroups[0].Code = []string{"src/feature.go"}
	mapsMsg := mappingsGeneratedMsg{groups: newGroups}

	model, _ = m.Update(mapsMsg)
	m = model.(MainModel)

	// Verify
	assert.Contains(t, m.View(), "src/feature.go")

	// Press Enter to confirm
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(MainModel)
	model, _ = m.Update(NextStepMsg{})
	m = model.(MainModel)

	// --- Step 5: Finalize ---
	assert.Contains(t, m.currentPage.Title(), "Finalizing")

	// Init triggers saveConfigCmd.
	// We simulate success.
	// But here, let's actually RUN the command to verify it writes the file!
	// The command is returned by Init.
	finalPage := m.currentPage.(*FinalizePage) // Check cast
	saveCmd := finalPage.saveConfigCmd()

	// Execute the command (it's a function)
	msg := saveCmd()

	// Assert msg is success
	assert.IsType(t, configSavedMsg{}, msg)

	// Assert file exists
	_, err = os.Stat(".drift.yaml")
	assert.NoError(t, err)

	// Update model with success
	model, _ = m.Update(msg)
	m = model.(MainModel)

	assert.Contains(t, m.View(), "saved")
}
