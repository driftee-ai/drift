# Drift

[**Full Documentation**](https://driftee-ai.github.io/drift)

## Why use Drift?

Documentation falls behind. It's inevitable. 

Drift automates the synchronization process by using AI to semantically compare your codebase against your documentation, instantly catching discrepancies before they ship. No more outdated READMEs or misleading API references in production.

- **Never Ship Outdated Docs:** Catch documentation drift the moment it happens. Keep your users happy and your API references perfectly accurate.
- **Automated PR Enforcement:** Put documentation review on autopilot. Drift runs natively in your CI/CD pipeline to flag missing updates before they merge.
- **Zero Friction Setup:** Map your entire codebase to your documentation in minutes using our interactive CLI wizard and wildcard-friendly YAML.
- **Model Agnostic:** Plug and play with your preferred LLM provider. First-class support for Gemini, OpenAI, and Anthropic.

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

All `drift` commands support a global `--version` or `-v` flag to print the version and exit.

```bash
drift --version
```

### `drift init`

Initializes a new project by launching an interactive setup wizard. It automatically discovers and maps your codebase to your documentation.

```bash
drift init
```

*(You can use `drift init --fast` to only use file paths for discovery, making it quicker and cheaper if you have a large repository.)*

**Specify a target directory:**

```bash
drift init --dir /path/to/target/directory
```

### `drift check`

Evaluates your codebase against your documentation to detect out-of-sync content based on your `.drift.yaml` configuration.

**Check all files:**

```bash
drift check
```

**Use a custom configuration file:**

```bash
drift check --config /path/to/your/config.yaml
```

**Check only changed files:**

For lightning-fast CI/CD pipelines, use `--changed-files` to verify only the code modified in an active pull request. See the [full documentation](https://driftee-ai.github.io/drift) for advanced Github Actions integration.

## Configuration

Drift uses a `.drift.yaml` file to map code to documentation. *(Note: Drift will also automatically discover `.drift.yml`, `drift.yaml`, and `drift.yml` if you prefer).*

- **`provider`**: The backend LLM (`gemini`, `openai`, or `anthropic`).
- **`rules`**: A list of mappings to evaluate.
  - **`name`**: A readable identifier for the rule.
  - **`code`**: Glob patterns targeting your source code files.
  - **`docs`**: Glob patterns targeting the corresponding documentation files.

### Example `.drift.yaml`

```yaml
version: 1
provider: gemini
rules:
  - name: "User API"
    code:
      - "src/api/user.go"
    docs:
      - "docs/api/users.md"
  - name: "Authentication"
    code:
      - "src/auth/**/*.go"
    docs:
      - "docs/auth.md"
```

## Providers

### Gemini

### Gemini

Authenticate the Google Gemini API:

```bash
export GEMINI_API_KEY="your-api-key"
```

### OpenAI

Authenticate the OpenAI API:

```bash
export OPENAI_API_KEY="your-api-key"
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
