# Drift Check GitHub Action

This action runs `drift` to check for documentation drift in your repository.

## Usage

### With Gemini

```yaml
- name: Drift Check (Gemini)
  uses: driftee-ai/drift/actions/drift-check@v1
  with:
    gemini-api-key: ${{ secrets.GEMINI_API_KEY }}
```

### With OpenAI

```yaml
- name: Drift Check (OpenAI)
  uses: driftee-ai/drift/actions/drift-check@v1
  with:
    openai-api-key: ${{ secrets.OPENAI_API_KEY }}
```

## Inputs

- `config`: Path to the `.drift.yaml` config file. Defaults to `.drift.yaml`.
- `gemini-api-key`: API key for the Gemini provider.
- `openai-api-key`: API key for the OpenAI provider.
- `version`: The version of `drift` to install. Defaults to the latest version.

**Note:** You must provide an API key for the provider specified in your `.drift.yaml` file.

## Versioning

It is recommended to pin the action to a specific version to ensure stability. You can do this by specifying a version number in the `uses` field:

```yaml
- name: Drift Check
  uses: driftee-ai/drift/actions/drift-check@v1.2.3 # Replace with your desired version
  with:
    gemini-api-key: ${{ secrets.GEMINI_API_KEY }}
```
