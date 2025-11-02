# Drift

Drift is a command-line tool to detect and prevent drift between your code and your documentation.
It uses large language models to assess if your documentation accurately reflects your code.

## Installation

### Homebrew (macOS and Linux)

```bash
brew tap driftee-ai/drift
brew install drift
```

### Go

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

**Check all files:**

```bash
drift check
```

**Use a custom configuration file:**

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

## Community & Support

- **Found a bug?** [File an issue](https://github.com/driftee-ai/drift/issues)
- **Have a question or suggestion?** [Start a discussion](https://github.com/driftee-ai/drift/discussions)

## Contributing

Contributions are welcome! We'd love your help in making `drift` even better.

Please see our [Contributing Guidelines](CONTRIBUTING.md) for more information.

Here's a quick guide to get you started:

- **Run tests:** `make test`
- **Run linter:** `make lint`

## License

`drift` is licensed under the [Apache License 2.0](LICENSE).
