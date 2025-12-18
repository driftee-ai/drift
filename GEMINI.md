# Technical Implementation

This document provides technical details and implementation guidance for our products.

## General Principles

- **Planning First:** Before implementing any feature or performing complex refactoring, a detailed plan must be formulated and presented to the user for approval. No code should be written until the plan is approved.
- **Verification After:** Upon completing a feature, bug fix, or any stage of a multi-step plan, always execute the full project verification suite (e.g., `make test integration-test lint build`) to ensure system integrity and performance.

## Tech Stack Overview

### Open-Source CLI

**Language:** Go

- Fast compilation
- Single binary distribution
- Excellent CI/CD support
- Strong parsing library ecosystem

**Key Dependencies:**

- LLM API clients (Google AI, OpenAI)
- YAML parser for config files (`gopkg.in/yaml.v3`)
- Git integration for diff detection (handled via command-line arguments)

**Repository:** [`github.com/driftee-ai/drift`](http://github.com/driftee-ai/drift) (Apache 2.0 License)

## Third-Party Services & Vendors

### Development & Version Control

- **GitHub Organization:** For code hosting and version control
- **Domain Registrar:** For [driftee.com](http://driftee.com) and related domains

### Infrastructure & Hosting

- **Vercel:** Frontend hosting (Next.js app)
- **Google Cloud Platform (GCP):**
    - Cloud Run for API server
    - Cloud SQL for PostgreSQL
    - Pub/Sub for job queues
- **GitHub Container Registry ([ghcr.io](http://io)):** Docker image hosting

### Data & Storage

- **Supabase:** Database (PostgreSQL) and authentication
- Alternative: Google Cloud SQL

### Authentication & Security

- **Supabase Auth** or **Auth0:** User authentication
- **Stripe:** Payment processing
- **PostHog** or **Mixpanel:** Product analytics

### LLM Providers

- **Google AI (Gemini):** Primary LLM provider
- **OpenAI (GPT-4):** Alternative/fallback
- **Anthropic (Claude):** Alternative for specific use cases

### CI/CD Platform Support

We support integration with:

- GitHub Actions
- GitLab CI
- Jenkins
- CircleCI
- Bitbucket Pipelines

## Go Client Architecture

This section provides the complete package structure and implementation details for the open-source CLI.

### Repository Structure

```
github.com/driftee/drift/
├── main.go                    # Entry point - calls cmd.Execute()
├── .drift.yaml               # Example config file
├── go.mod                    # Go module definitions
├── Dockerfile                # For Docker builds
├── README.md                 # Project documentation
├── cmd/                      # CLI commands (no business logic)
│   ├── root.go              # Root 'drift' command
│   ├── init.go              # 'drift init' command
│   └── check.go             # 'drift check' command
├── pkg/
│   ├── assessor/            # Assessment logic
│   │   ├── assessor.go      # DriftAssessor implementation
│   │   └── factory.go       # Factory for creating assessors
│   ├── config/              # Config file handling
│   │   └── config.go
│   ├── files/               # File searching and reading
│   │   └── files.go
│   ├── llm/                 # LLM provider abstraction
│   │   ├── llm.go           # Generator interface
│   │   ├── factory.go       # Factory for creating generators
│   │   ├── gemini.go        # Google AI implementation
│   │   ├── openai.go        # OpenAI implementation
│   │   └── dummy.go         # Mock implementation for tests
│   ├── rules/               # Rule filtering logic
│   │   └── filter.go
│   └── tui/                 # Terminal UI for interactive commands
│       ├── wizard.go        # drift init wizard logic
│       └── styles.go        # TUI styling
```

### Package Details

### `/cmd` - CLI Commands

Uses Cobra for command-line interface. Contains zero business logic.

**cmd/root.go** - Defines the root `drift` command.

**cmd/init.go** - Defines `drift init`.
- Launches an interactive TUI wizard (via `pkg/tui`) to guide the user through discovery and configuration.

**cmd/check.go** - Defines `drift check`.
- Reads flags like `--config` and `--changed-files`.
- Calls `config.Load()` to get the config.
- Calls `rules.FilterTriggeredRules()` to get the list of rules to check based on changed files.
- Calls `files.FindFiles()` and `files.ReadFiles()` to get the content of code and docs.
- Calls `assessor.New()` to get the right assessor.
- Loops through triggered rules, reads files, and calls `assessor.Assess()` for each.

### `/pkg/config` - Configuration Management

**config.go**

- Defines `Config` and `Rule` structs (to match YAML).
- `Load(path string) (*Config, error)` - Uses `gopkg.in/yaml.v3` to find and unmarshal `.drift.yaml`.
- `CreateScaffold(path string) error` - Creates blank, commented `.drift.yaml` for `drift init`.

**Config Schema:**

```yaml
version: 1
provider: gemini
rules:
- name: Example API Documentation
  code:
  - src/api/**/*.go
  docs:
  - docs/api/**/*.md
```

### `/pkg/tui` - Terminal User Interface

**wizard.go**
- Implements the Bubble Tea model for the `drift init` wizard.
- Manages a state machine (`StateDiscovery`, `StateGrouping`, etc.) to guide the user.
- **StateDiscovery**: Walks the file system using `pkg/files` and allows the user to review/ignore discovered files.

### `/pkg/rules` - Rule Filtering

**filter.go**

- `FilterTriggeredRules(rules []config.Rule, changedFiles []string) ([]config.Rule, error)`: Filters rules from the config against a list of changed file paths. It uses glob matching (with support for `**`) to determine which rules are "triggered". If `changedFiles` is empty, it returns all rules.

### `/pkg/files` - File Handling

**files.go**

- `FindFiles(globs []string) ([]string, error)`: Takes a slice of glob patterns and returns a list of matching file paths.
- `ReadFiles(paths []string) (map[string]string, error)`: Reads a list of file paths and returns a map of file paths to their content.
- `ReadAndConcatenate(paths []string) (string, error)`: Reads a list of file paths and returns their concatenated content.
- `WalkProject()`: Recursively walks the project directory, respecting `.gitignore` and skipping common non-code directories (e.g., `.git`, `node_modules`).

### `/pkg/llm` - LLM Provider Abstraction

**llm.go**
- Defines the `Generator` interface for LLM providers.
    - `Generate(ctx, prompt)`: Basic text generation.
    - `GenerateJSON(ctx, prompt, schema, result)`: Structured JSON generation using provider-specific schema mapping.
- Defines a provider-agnostic `Schema` struct.

**factory.go**
- `New(provider string) (Generator, error)`: Factory for creating specific LLM generators (`gemini`, `openai`, `dummy`).

### `/pkg/assessor` - Logic Layer

**assessor.go**
- `DocAssessor` interface: `Assess(docContent, codeContents)`.
- `DriftAssessor` struct: Implements `DocAssessor` using an `llm.Generator`.
    - Handles prompt construction and response parsing for drift detection.

**factory.go**
- `New(provider string) (DocAssessor, error)`: Factory that initializes a `DriftAssessor` with the requested LLM provider.

### Core Go Libraries

**CLI Dependencies:**

- **cobra** - For CLI commands (`init`, `check`).
- **bubbletea**, **lipgloss**, **bubbles** - For the interactive TUI wizard.
- **gopkg.in/yaml.v3** - For reading `.drift.yaml` config.
- **doublestar** - For glob matching.
- **go-gitignore** - For respecting `.gitignore` during file discovery.
- **golangci-lint** - For linting the Go code.

### End-to-End Testing

Our end-to-end (E2E) tests are located in `main_integration_test.go` and use a structured `testdata/e2e` directory to organize test cases. These tests run the compiled `drift` binary against specific code and documentation examples to verify the tool's behavior.

**`testdata/e2e` Directory Structure:**

Test cases are categorized by their expected outcome:

-   **`true_positives/`**: Contains test cases where a drift *exists*, and the tool is expected to *correctly detect it*.
    -   Example: `missing_param_in_docs/` (code has a parameter not in docs).
-   **`true_negatives/`**: Contains test cases where code and documentation are *in sync*, and the tool is expected to *correctly confirm that*.
    -   Example: `in_sync_example/` (code and docs match perfectly).
-   **`false_positives/`**: Contains test cases where a naive check might incorrectly flag a drift, but the LLM-based tool should *correctly identify them as in sync*.
    -   Example: `cosmetic_diff_example/` (semantically equivalent code/docs with minor wording differences).
-   **`false_negatives/`**: Contains test cases where a subtle drift *exists*, and the tool is expected to *correctly detect it* (i.e., it's a real drift that might be missed by simpler checks).
    -   Example: `subtle_drift_example/` (code returns a pointer, docs say struct).

**Adding New E2E Tests:**

1.  Create a new subdirectory under the appropriate classification (e.g., `testdata/e2e/true_positives/my_new_test`).
2.  Inside this directory, create:
    -   A `.drift.yaml` file configured for the `gemini` provider, pointing to `code.go` and `docs.md` within the same directory.
    -   A `code.go` file with the relevant code.
    -   A `docs.md` file with the corresponding documentation.
3.  Add a new test function to `main_integration_test.go` that:
    -   Runs the `drift check` command with the `.drift.yaml` from your new test case.
    -   Asserts the expected outcome (e.g., "Result: In Sync" or "Result: Out of Sync").

**Conditional Execution of Live Tests:**

Tests that interact with the Gemini API require the `GEMINI_API_KEY` environment variable to be set. If the key is not present, these tests will fail, indicating that the live API could not be reached. This allows for fast local development (e.g. using the `dummy` provider) without requiring an API key, while ensuring live tests run in CI environments where the key is configured as a secret.
