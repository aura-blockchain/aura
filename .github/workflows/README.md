# GitHub Workflows - DISABLED

**Status:** All GitHub Actions workflows in this directory are **DISABLED** for this repository.

## Why Workflows Are Disabled

This project uses local testing exclusively. All CI/CD, testing, and validation is performed on the development machine (bcpc) before commits are pushed to GitHub.

## Available Workflows (For Reference)

These workflow files exist for documentation and potential future use, but are **not active**:

1. **ci.yml** - Continuous Integration
   - Linting (golangci-lint, gofmt)
   - Testing (Go tests with race detection, coverage)
   - Building (multi-platform builds: linux/amd64, linux/arm64)
   - Proto verification (buf linting and breaking change detection)
   - Integration tests

2. **security.yml** - Security Scanning
   - gosec security scanning for Go code
   - Dependency vulnerability scanning
   - CodeQL analysis
   - SARIF report generation
   - Scheduled weekly scans

3. **release.yml** - Release Automation
   - Version validation
   - Multi-platform binary builds (linux, darwin, windows)
   - Docker image builds and publishing
   - GitHub release creation with changelog
   - Artifact uploads

4. **deploy-docs.yml** - Documentation Deployment
   - Builds documentation site from `docs-site/`
   - Deploys to GitHub Pages
   - Triggered on main branch pushes to docs

## Local Testing Commands

Instead of relying on GitHub Actions, use these local commands:

```bash
# Lint
cd chain && golangci-lint run --timeout=10m

# Test with coverage
cd chain && go test ./... -race -coverprofile=coverage.out -timeout=20m

# Build
cd chain && go build -o aurad ./cmd/aurad

# Security scan
cd chain && gosec -exclude-dir=tests -exclude-dir=testutil ./...

# Proto validation
cd proto && buf lint && buf breaking --against .git#branch=main
```

## Future Use

If GitHub Actions are enabled in the future, these workflows are production-ready and tested. Simply enable Actions in the repository settings.

## See Also

- **CLAUDE.md** - Section "Git" mentions "GitHub Actions are DISABLED. Test locally."
- **Local testing setup** - See `/home/hudson/blockchain-projects/CLAUDE.md`
