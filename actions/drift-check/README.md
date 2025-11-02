# Drift Check GitHub Action

This action runs `drift` to check for documentation drift in your repository.

## Usage

```yaml
- name: Drift Check
  uses: driftee-ai/drift/actions/drift-check@v1
  with:
    gemini-api-key: ${{ secrets.GEMINI_API_KEY }}
```

## Inputs

- `config`: Path to the `.drift.yaml` config file. Defaults to `.drift.yaml`.
- `gemini-api-key`: API key for the Gemini provider. This is a required input.
- `version`: The version of `drift` to install. Defaults to the latest version.
