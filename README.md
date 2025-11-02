# Drift

Drift is a command-line tool to detect and prevent drift between your code and your documentation.
It uses large language models to assess if your documentation accurately reflects your code.

## Installation

You can install `drift` using `go install`:

```bash
go install github.com/driftee-ai/drift@latest
```

## Usage

### `drift init`

Initializes a new project by creating a `.drift.yaml` configuration file in the current directory.

```bash
drift init
```

### `drift check`

Checks for drift between your code and documentation based on the rules in your `.drift.yaml` file.

```bash
drift check
```

You can also specify a custom configuration file using the `--config` flag:

```bash
drift check --config /path/to/your/config.yaml
```

## Configuration

The `.drift.yaml` file defines the rules for checking drift.

- **`provider`**: The backend provider to use for assessing drift. Currently, only `"gemini"` is supported.
- **`rules`**: A list of rules to check.
  - **`name`**: A descriptive name for the rule.
  - **`code`**: A list of glob patterns for the code files.
  - **`docs`**: A list of glob patterns for the documentation files.

### Example `.drift.yaml`

```yaml
version: 1
provider: gemini
rules:
  - name: "User API Documentation"
    code:
      - "src/api/user.go"
    docs:
      - "docs/api/users.md"
  - name: "Authentication Service"
    code:
      - "src/auth/**/*.go"
    docs:
      - "docs/auth.md"
```

## Providers

### Gemini

To use the Gemini provider, you need to set the `GEMINI_API_KEY` environment variable to your Gemini API key.

```bash
export GEMINI_API_KEY="your-api-key"
```

## Contributing

Contributions are welcome! Here's how to get started:

- **Run tests:** `make test`
- **Run linter:** `make lint`
