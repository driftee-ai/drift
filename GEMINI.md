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

- Reads flags like `--changed-files`
- Calls `config.Load()` to get the config
- Calls [`assessor.New`](http://assessor.New)`()` to get the right assessor
- Calls `drift.RunCheck()` to orchestrate the check

### `/pkg/config` - Configuration Management

**config.go**

- Defines `Config` and `Mapping` structs (to match YAML)
- `Load(path string) (*Config, error)` - Uses viper to find and unmarshal `.drift.yaml`
- `CreateScaffold(path string) error` - Creates blank, commented `.drift.yaml` for `drift init`

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

### `/pkg/drift` - Core Orchestrator

**check.go**

`RunCheck(config *config.Config, changedFiles []string) error`:

1. Takes the config and the list of changed files
2. Loops through changed files to find which `config.Mapping`s were "triggered"
3. For each triggered mapping, reads the doc/code files from disk
4. Calls `assessor.Assess()` function
5. Reads the `AssessmentResult`
6. If `IsInSync == false`, prints the `Reason` and returns an error (which causes `exit 1`)
7. If all checks pass, returns `nil`

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