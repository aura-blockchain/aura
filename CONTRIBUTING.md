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

## Good First Issues

New to the project? These beginner-friendly tasks are great starting points:

1. **Add unit tests to auth module** - Increase coverage for `chain/x/auth/keeper/`
2. **Improve error messages in identity module** - Make errors in `chain/x/identity/types/errors.go` more descriptive
3. **Add CLI help examples** - Expand `--help` text in `chain/x/governance/client/cli/`
4. **Document a keeper method** - Add godoc comments to unexported functions in any `x/*/keeper/`
5. **Add validation tests** - Write tests for `ValidateBasic()` methods in `chain/x/*/types/msgs.go`
6. **Improve logging** - Add structured logging to `chain/x/compliance/keeper/`
7. **Add table-driven tests** - Convert simple tests to table-driven format in any module
8. **Fix linter warnings** - Run `golangci-lint` and fix reported issues

Look for issues labeled `good first issue` or `help wanted` on GitHub.

## Questions?

- Open a GitHub Discussion
- Check existing documentation

## Developer Certificate of Origin (DCO)

All contributions to this project must be signed off to indicate that you have the right to submit the work under our open source license. This is done using the `--signoff` (or `-s`) flag when committing:

```bash
git commit -s -m "feat: your feature description"
```

This adds a `Signed-off-by:` line to your commit message, certifying that you agree to the [Developer Certificate of Origin](https://developercertificate.org/):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

### Configuring Git for Sign-off

To automatically sign off all commits, configure your Git identity:

```bash
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

### Amending Unsigned Commits

If you forgot to sign off a commit, you can amend it:

```bash
git commit --amend -s
```

For multiple commits, use interactive rebase with the `exec` command.

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0 (see LICENSE file).
