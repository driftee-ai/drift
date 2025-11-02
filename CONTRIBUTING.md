# Contributing to Drift

First off, thank you for considering contributing to Drift! It's people like you that make open source software such a great thing.

We welcome any and all contributions. Here are a few ways you can help:

- [Reporting Bugs](#reporting-bugs)
- [Suggesting Enhancements](#suggesting-enhancements)
- [Code Contributions](#code-contributions)

## Reporting Bugs

If you find a bug, please [file an issue](https://github.com/driftee-ai/drift/issues) and provide the following information:

- A clear and descriptive title.
- A detailed description of the problem, including steps to reproduce it.
- The version of `drift` you are using (`drift --version`).
- Your operating system.
- Any relevant logs or error messages.

## Suggesting Enhancements

If you have an idea for a new feature or an improvement to an existing one, please [start a discussion](https://github.com/driftee-ai/drift/discussions) to share your idea.

## Code Contributions

We love pull requests! Here's a quick guide to get you started:

1.  **Fork the repository** and create your branch from `main`.
2.  **Set up your development environment** (see below).
3.  **Make your changes** and ensure that the tests and linter pass.
4.  **Submit a pull request** with a clear description of your changes.

### Development Environment

`drift` is written in Go. You will need to have Go installed on your system.

1.  Clone your fork of the repository:
    ```bash
    git clone https://github.com/YOUR_USERNAME/drift.git
    ```
2.  Install the dependencies:
    ```bash
    go mod tidy
    ```

### Running Tests and Linters

To run the tests:

```bash
make test
```

To run the linter:

```bash
make lint
```
