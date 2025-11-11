# Technical Implementation

This document provides technical details and implementation guidance for both v1 and v2 products.

## Tech Stack Overview

### v1: Open-Source CLI

**Language:** Go

- Fast compilation
- Single binary distribution
- Excellent CI/CD support
- Strong parsing library ecosystem

**Key Dependencies:**

- `tree-sitter` (with Go bindings) for code parsing
- LLM API clients (Google AI, OpenAI, Anthropic)
- YAML parser for config files
- Git integration for diff detection

**Repository:** [`github.com/driftee-ai/drift`](http://github.com/driftee-ai/drift) (Apache 2.0 License)

### v2: Platform

**Full-stack TypeScript** for consistency and developer velocity

**Frontend:**

- Marketing site: Framer/Webflow for fast iteration
- Web app: Next.js on Vercel
- Dashboard: React with data visualization libraries

**Backend:**

- API Server: NestJS on Google Cloud Run (Serverless)
- Database: PostgreSQL (Supabase or Google Cloud SQL)
- Queue: Google Cloud Pub/Sub for job handling
- Auth: Supabase Auth or Auth0

**Agent:**

- Docker container with proprietary Go binary
- Includes tree-sitter libraries
- Runs in customer's CI environment

**Repository:** [`github.com/driftee-ai/platform`](http://github.com/driftee-ai/platform) (Proprietary)

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
- **GitHub Container Registry ([ghcr.io](http://ghcr.io)):** Docker image hosting for the v2 agent

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

## v1 Go Client Architecture

### Proposed Package Structure

```

drift/

├── cmd/

│   └── drift/           # Main CLI entry point

│       ├── main.go

│       ├── check.go     # 'drift check' command

│       └── init.go      # 'drift init' command

├── internal/

│   ├── config/          # Config file parsing (.drift.yaml)

│   ├── parser/          # tree-sitter integration

│   ├── differ/          # Git diff detection

│   ├── llm/             # LLM API clients

│   └── reporter/        # Output formatting

└── pkg/

└── drift/           # Public API (if needed for plugins)

```

### Key Interfaces

**Parser Interface:**

```go

type Parser interface {

ParseFile(path string) (*APISignatures, error)

ExtractPublicAPI(ast *AST) []Symbol

}

```

**LLM Client Interface:**

```go

type LLMClient interface {

CheckDrift(code, docs string) (*DriftReport, error)

SuggestFix(drift *DriftReport) (string, error)

}

```

**Config Schema:**

```yaml

version: 1

mappings:

- code: src/api/user.go
    
    docs: docs/api/[users.md](http://users.md)
    
- code: src/api/auth.go
    
    docs: docs/api/[authentication.md](http://authentication.md)
    

```

This is customer-facing documentation for installing the v2 agent (running inside Docker as `drift`) in CI/CD platforms.

### Installation Pattern (All Platforms)

1. Pull the Docker image: `docker pull [ghcr.io/driftee-ai/agent:latest](http://ghcr.io/driftee-ai/agent:latest)`
2. Set environment variables: `DRIFTEE_API_KEY`
3. Run the agent: `docker run` [`](http://ghcr.io/driftee-ai/agent:latest)[ghcr.io/driftee-ai/agent:latest`](http://ghcr.io/driftee-ai/agent:latest`) `drift check`

### Platform-Specific Examples

### GitHub Actions

```yaml

```

name: Drift Check

on: [pull_request]

jobs:

drift:

runs-on: ubuntu-latest

steps:

- uses: actions/checkout@v3
- name: Run Driftee Agent
    
    env:
    
    DRIFTEE_API_KEY: $ secrets.DRIFTEE_API_KEY
    
    run: |
    
    docker pull [ghcr.io/driftee-ai/agent:latest](http://ghcr.io/driftee-ai/agent:latest)
    
    docker run [ghcr.io/driftee-ai/agent:latest](http://ghcr.io/driftee-ai/agent:latest) drift check
    

```

```

```

### GitLab CI

```yaml

```

drift_check:

image: [ghcr.io/driftee-ai/agent:latest](http://ghcr.io/driftee-ai/agent:latest)

script:

- drift check

only:

- merge_requests

variables:

DRIFTEE_API_KEY: $DRIFTEE_API_KEY

```

```

```

### Jenkins

```groovy

```

pipeline {

agent any

stages {

stage('Drift Check') {

steps {

docker.image('[ghcr.io/driftee-ai/agent:latest').inside](http://ghcr.io/driftee-ai/agent:latest').inside) {

sh 'drift check'

}

}

}

}

}

}

}

}

}

```

```

}

```

### CircleCI

```yaml

```

version: 2.1

jobs:

drift-check:

docker:

- image: [ghcr.io/driftee-ai/agent:latest](http://ghcr.io/driftee-ai/agent:latest)

steps:

- checkout
- run: drift check

```

```

```

### Bitbucket Pipelines

```yaml

```

pipelines:

pull-requests:

'':

- step:
    
    name: Drift Check
    
    image: [ghcr.io/driftee-ai/agent:latest](http://ghcr.io/driftee-ai/agent:latest)
    
    script:
    
    - drift check

```

```

```

## v2 Init Strategy (Core IP)

Our "Automated Init" feature uses a 4-phase funnel to automatically generate the `.drift.yaml` config file.

### Phase 1: Repo Analysis

- Scan directory structure
- Identify documentation folders (docs/, [README.md](http://README.md), etc.)
- Identify code folders (src/, lib/, pkg/, etc.)
- Detect programming languages

### Phase 2: Test-First Hack

- Parse test files to identify API boundaries
- Tests often import and use the public API
- Extract import patterns and function calls

### Phase 3: Code-First Map

- Use tree-sitter to parse code files
- Extract public APIs (exported functions, classes, interfaces)
- Map code symbols to likely documentation files

### Phase 4: LLM Mop-Up

- Use LLM to handle edge cases
- Match ambiguous documentation sections to code
- Generate confidence scores for each mapping

**Output:** A complete `.drift.yaml` file with high-confidence mappings

## Detailed v1 Go Package Architecture

This section provides the complete package structure and implementation details for the open-source CLI.

### Repository Structure

```
[github.com/driftee/drift/](http://github.com/driftee/drift/)
├── main.go                    # Entry point - calls cmd.Execute()
├── .drift.yaml               # Example config file
├── go.mod                    # Go module definitions
├── Dockerfile                # For Docker builds
├── [README.md](http://README.md)                 # Project documentation
├── cmd/                      # CLI commands (no business logic)
│   ├── root.go              # Root 'drift' command
│   ├── init.go              # 'drift init' command
│   └── check.go             # 'drift check' command
├── pkg/
│   ├── config/              # Config file handling
│   │   └── config.go
│   ├── assessor/            # LLM integration
│   │   ├── assessor.go      # Interface definition
│   │   ├── gemini.go        # Google AI implementation
│   │   └── openai.go        # OpenAI implementation
│   ├── drift/               # Core business logic
│   │   └── check.go
│   └── local/               # Smart Context optimization
│       └── parser.go
```

### Package Details

### `/cmd` - CLI Commands

Uses Cobra for command-line interface. Contains zero business logic.

**cmd/root.go** - Defines the root `drift` command

**cmd/init.go** - Defines `drift init`

- Calls `config.CreateScaffold()` from `/pkg`

**cmd/check.go** - Defines `drift check`

- Reads flags like `--config` and `--changed-files`.
- Calls `config.Load()` to get the config.
- Calls `rules.FilterTriggeredRules()` to get the list of rules to check based on changed files.
- Calls `assessor.New()` to get the right assessor.
- Loops through triggered rules, reads files, and calls `assessor.Assess()` for each.

### `/pkg/config` - Configuration Management

**config.go**

- Defines `Config` and `Mapping` structs (to match YAML)
- `Load(path string) (*Config, error)` - Uses viper to find and unmarshal `.drift.yaml`
- `CreateScaffold(path string) error` - Creates blank, commented `.drift.yaml` for `drift init`

### `/pkg/rules` - Rule Filtering

**filter.go**

- `FilterTriggeredRules(rules []config.Rule, changedFiles []string) ([]config.Rule, error)`: Filters rules from the config against a list of changed file paths. It uses glob matching (with support for `**`) to determine which rules are "triggered". If `changedFiles` is empty, it returns all rules.

### `/pkg/assessor` - LLM Interface

**assessor.go** - Core interface definition:

```go
type AssessmentResult struct {
    IsInSync bool   `json:"is_in_sync"`
    Reason   string `json:"reason"`
}

type DocAssessor interface {
    Assess(docContent string, codeContents map[string]string) (*AssessmentResult, error)
}
```

`New(config *config.Config) (DocAssessor, error)` - Factory function that reads the config (`provider: gemini`) and returns the correct concrete assessor.

**gemini.go** - Google AI implementation

- `type GeminiAssessor struct { ... }`
- `NewGeminiAssessor() *GeminiAssessor` - Reads `GOOGLE_API_KEY` from env
- `Assess(...)` - Formats Gemini-specific prompt, makes API call, parses JSON response

**openai.go** - OpenAI implementation

- `type OpenAIAssessor struct { ... }`
- `NewOpenAIAssessor() *OpenAIAssessor` - Reads `OPENAI_API_KEY` from env
- `Assess(...)` - Formats OpenAI-specific prompt, makes API call

### `/pkg/local` - Smart Context (v1.1 Optimization)

This package will be added when we build the "Smart Context" optimization.

**parser.go**

`ExtractSymbols(filepath string, content []byte) ([]string, error)`:

1. Uses `go-tree-sitter`
2. Detects the language from the file extension
3. Applies the correct tree-sitter query to extract function names, class names, etc.
4. Returns a list of symbols to be used in the "Smart Prompt"

### Core Go Libraries

**v1 CLI Dependencies:**

- **cobra** - For CLI commands (`init`, `check`)
- **viper** - For reading `.drift.yaml` config
- **go-tree-sitter** - For the "Smart Context" optimization
- **golangci-lint** - For linting the Go code

### End-to-End Testing

Our end-to-end (E2E) tests are located in `main_test.go` and use a structured `testdata/e2e` directory to organize test cases. These tests run the compiled `drift` binary against specific code and documentation examples to verify the tool's behavior.

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
3.  Add a new test function to `main_test.go` that:
    -   Sets `GEMINI_API_KEY` if it's a live test.
    -   Runs the `drift check` command with the `.drift.yaml` from your new test case.
    -   Asserts the expected outcome (e.g., "Result: In Sync" or "Result: Out of Sync").

**Conditional Execution of Live Tests:**

Tests that interact with the Gemini API (i.e., those using the `gemini` provider in their `.drift.yaml`) are skipped if the `GEMINI_API_KEY` environment variable is not set. This allows for fast local development without requiring an API key, while ensuring live tests run in CI environments where the key is configured as a secret.