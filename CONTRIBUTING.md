# Contributing to AURA

Thank you for your interest in contributing to AURA! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

### Reporting Bugs

- Check existing issues to avoid duplicates
- Use the bug report template
- Include steps to reproduce
- Provide system information (OS, Go version, etc.)

### Suggesting Features

- Check existing issues and discussions
- Clearly describe the feature and its use case
- Consider implementation complexity

### Submitting Code

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes**
4. **Run tests locally**
   ```bash
   cd chain
   go test ./...
   make test-authz   # quick auth/PermissionDenied checks
   pre-commit run --all-files
   ```
5. **Commit with clear messages**
   ```bash
   git commit -m "feat: add new feature description"
   ```
6. **Push and create a Pull Request**

## Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/aura.git
cd aura/chain

# Install Go dependencies
go mod download

# Install pre-commit hooks
cd ..
pre-commit install

# Build the daemon
cd chain
go build -o aurad ./cmd/aurad
```

## Code Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use meaningful variable and function names

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation only
- `style:` - Formatting, no code change
- `refactor:` - Code change that neither fixes a bug nor adds a feature
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

### Testing

- Write tests for new features
- Use table-driven tests where appropriate
- All tests must pass before merge
- Include integration tests for module changes

### Security

- Never commit secrets or private keys
- Run `gosec` security scanner
- Follow Cosmos SDK security best practices
- Review WASM contracts carefully
- Report vulnerabilities privately (see SECURITY.md)

### Module Development

When modifying Cosmos SDK modules in `chain/x/`:
1. Update protobuf definitions in `proto/`
2. Regenerate proto files (`make proto-gen`)
3. Implement keeper methods
4. Add genesis import/export
5. Register in `app/app.go`
6. Write comprehensive tests
7. Update documentation

## Pull Request Process

1. Update documentation if needed
2. Add tests for new functionality
3. Ensure all checks pass
4. Request review from maintainers
5. Address review feedback
6. Squash commits if requested

## Questions?

- Open a GitHub Discussion
- Check existing documentation

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see LICENSE file).
