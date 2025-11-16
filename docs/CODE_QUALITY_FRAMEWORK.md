# Code Quality Framework

## Overview

This document establishes comprehensive standards and procedures for code quality, coverage testing, fuzz testing, and static analysis for the Aura blockchain. These practices ensure high code reliability, security, and maintainability throughout the development lifecycle.

---

## Table of Contents

1. [Code Coverage Testing](#1-code-coverage-testing)
2. [Fuzz Testing](#2-fuzz-testing)
3. [Static Analysis](#3-static-analysis)
4. [Integration and Automation](#4-integration-and-automation)
5. [Quality Metrics and Reporting](#5-quality-metrics-and-reporting)
6. [Continuous Improvement](#6-continuous-improvement)

---

## 1. Code Coverage Testing

### 1.1 Coverage Requirements

#### Coverage Targets by Module Type

| Module Type | Line Coverage | Branch Coverage | Function Coverage |
|-------------|--------------|----------------|-------------------|
| Critical (Identity, VC, Tokens) | 95%+ | 90%+ | 100% |
| Important (DEX, Bridge, Confidence) | 90%+ | 85%+ | 95% |
| Supporting (Monitoring, Utilities) | 85%+ | 80%+ | 90% |
| **Overall Project** | **90%+** | **85%+** | **95%+** |

#### Critical Modules

The following modules require 95%+ coverage:
- `x/identitychange` - Identity transition logic
- `x/vcregistry` - Verifiable credential management
- `x/confidencescore` - Trust score calculation
- Core token transfer logic
- Staking and governance modules
- Cryptographic implementations

### 1.2 Testing Infrastructure Setup

#### Go Testing Tools Installation

```bash
# Standard Go testing tools (built-in)
go test -cover ./...

# Coverage visualization
go get -u github.com/axw/gocov/gocov
go get -u github.com/AlekSi/gocov-xml
go get -u github.com/matm/gocov-html

# Coverage badge generation
go get -u github.com/jpoles1/gopherbadger

# Enhanced coverage tools
go get -u github.com/boumenot/gocover-cobertura
```

#### Directory Structure

```
chain/
├── coverage/                    # Coverage reports
│   ├── coverage.out            # Raw coverage data
│   ├── coverage.html           # HTML report
│   ├── coverage.xml            # XML report (for CI)
│   └── coverage-by-module/     # Per-module reports
├── .github/
│   └── workflows/
│       └── coverage.yml        # Coverage CI workflow
└── scripts/
    ├── run-coverage.sh         # Coverage generation script
    └── coverage-report.sh      # Report generation
```

### 1.3 Coverage Generation Scripts

#### Basic Coverage Generation

Create `scripts/run-coverage.sh`:

```bash
#!/bin/bash
# run-coverage.sh - Generate comprehensive code coverage

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COVERAGE_DIR="${PROJECT_ROOT}/coverage"

# Create coverage directory
mkdir -p "${COVERAGE_DIR}"
mkdir -p "${COVERAGE_DIR}/coverage-by-module"

echo "Running tests with coverage..."

# Generate coverage for all packages
go test -v -coverprofile="${COVERAGE_DIR}/coverage.out" \
  -covermode=atomic \
  -timeout=30m \
  ./...

# Generate HTML report
go tool cover -html="${COVERAGE_DIR}/coverage.out" \
  -o "${COVERAGE_DIR}/coverage.html"

# Calculate overall coverage percentage
COVERAGE=$(go tool cover -func="${COVERAGE_DIR}/coverage.out" | \
  grep total | awk '{print substr($3, 1, length($3)-1)}')

echo "Overall Coverage: ${COVERAGE}%"

# Check if coverage meets threshold
THRESHOLD=90
if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
  echo "ERROR: Coverage ${COVERAGE}% is below threshold ${THRESHOLD}%"
  exit 1
fi

echo "Coverage check passed!"
```

Make it executable:
```bash
chmod +x scripts/run-coverage.sh
```

#### Per-Module Coverage

Create `scripts/coverage-by-module.sh`:

```bash
#!/bin/bash
# coverage-by-module.sh - Generate coverage reports per module

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COVERAGE_DIR="${PROJECT_ROOT}/coverage/coverage-by-module"

mkdir -p "${COVERAGE_DIR}"

# List of modules to analyze
MODULES=(
  "x/identitychange"
  "x/vcregistry"
  "x/confidencescore"
  "x/dataregistry"
  "x/prevalidation"
  "x/dex"
  "x/bridge"
  "x/validatorsecurity"
  "x/networksecurity"
  "x/privacy"
  "x/cryptography"
  "x/monitoring"
  "x/compliance"
)

echo "Generating per-module coverage reports..."

# Generate coverage for each module
for module in "${MODULES[@]}"; do
  echo "Processing module: ${module}"

  MODULE_PATH="${PROJECT_ROOT}/${module}"
  if [ ! -d "${MODULE_PATH}" ]; then
    echo "Warning: Module ${module} not found, skipping..."
    continue
  fi

  MODULE_NAME=$(basename "${module}")
  COVERAGE_FILE="${COVERAGE_DIR}/${MODULE_NAME}-coverage.out"

  # Generate coverage
  go test -v -coverprofile="${COVERAGE_FILE}" \
    -covermode=atomic \
    "./${module}/..."

  # Calculate coverage percentage
  if [ -f "${COVERAGE_FILE}" ]; then
    COVERAGE=$(go tool cover -func="${COVERAGE_FILE}" | \
      grep total | awk '{print substr($3, 1, length($3)-1)}')
    echo "${MODULE_NAME}: ${COVERAGE}%"

    # Generate HTML report
    go tool cover -html="${COVERAGE_FILE}" \
      -o "${COVERAGE_DIR}/${MODULE_NAME}-coverage.html"
  fi
done

echo "Per-module coverage reports generated in ${COVERAGE_DIR}"
```

### 1.4 Test Organization

#### Test File Structure

```
x/identitychange/
├── keeper/
│   ├── keeper.go
│   ├── keeper_test.go           # Unit tests for keeper
│   ├── msg_server.go
│   ├── msg_server_test.go       # Unit tests for message handlers
│   ├── query.go
│   └── query_test.go            # Unit tests for queries
├── types/
│   ├── types.go
│   ├── types_test.go            # Unit tests for types
│   ├── validation.go
│   └── validation_test.go       # Validation tests
└── integration_test.go          # Integration tests
```

#### Test Naming Conventions

```go
// Unit test example
func TestMsgSubmitIdentityChange_ValidateBasic(t *testing.T) {
    tests := []struct {
        name    string
        msg     MsgSubmitIdentityChange
        wantErr bool
    }{
        {
            name: "valid message",
            msg: MsgSubmitIdentityChange{
                OldIdentityId: "alice",
                NewIdentityId: "alice-new",
                Reason:        "passport-update",
            },
            wantErr: false,
        },
        {
            name: "empty old identity",
            msg: MsgSubmitIdentityChange{
                OldIdentityId: "",
                NewIdentityId: "alice-new",
                Reason:        "update",
            },
            wantErr: true,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.msg.ValidateBasic()
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

#### Table-Driven Tests

```go
// Table-driven test for comprehensive coverage
func TestConfidenceScore_Calculate(t *testing.T) {
    testCases := []struct {
        name           string
        verifications  int
        endorsements   int
        age            time.Duration
        expected       int
        expectError    bool
    }{
        {"zero verifications", 0, 0, 0, 0, false},
        {"single verification", 1, 0, 24 * time.Hour, 10, false},
        {"multiple verifications", 5, 3, 720 * time.Hour, 45, false},
        {"max score", 100, 50, 8760 * time.Hour, 100, false},
        {"negative values", -1, 0, 0, 0, true},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            score, err := CalculateConfidenceScore(
                tc.verifications,
                tc.endorsements,
                tc.age,
            )

            if tc.expectError {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.Equal(t, tc.expected, score)
        })
    }
}
```

### 1.5 Coverage Analysis

#### Identifying Coverage Gaps

```bash
# Generate detailed function coverage
go tool cover -func=coverage/coverage.out | grep -v "100.0%" | sort -k3 -n

# Find uncovered lines
go tool cover -func=coverage/coverage.out | grep "0.0%"

# Generate annotated source with coverage
go tool cover -html=coverage/coverage.out -o coverage/annotated.html
```

#### Coverage Report Parsing

Create `scripts/analyze-coverage.sh`:

```bash
#!/bin/bash
# analyze-coverage.sh - Analyze coverage and identify gaps

COVERAGE_FILE="coverage/coverage.out"

echo "=== Coverage Analysis ==="
echo ""

echo "Overall Coverage:"
go tool cover -func="${COVERAGE_FILE}" | grep total

echo ""
echo "Modules below 90% coverage:"
go tool cover -func="${COVERAGE_FILE}" | \
  grep -v "100.0%" | \
  awk '{if ($3 < 90) print $1, $3}' | \
  grep -v "total:"

echo ""
echo "Uncovered functions:"
go tool cover -func="${COVERAGE_FILE}" | grep "0.0%"

echo ""
echo "Detailed report available at: coverage/coverage.html"
```

### 1.6 GitHub Actions Integration

Create `.github/workflows/coverage.yml`:

```yaml
name: Code Coverage

on:
  push:
    branches: [main, master, develop]
  pull_request:
    branches: [main, master]

jobs:
  coverage:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests with coverage
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Generate coverage report
        run: |
          go tool cover -func=coverage.out -o=coverage.txt
          go tool cover -html=coverage.out -o=coverage.html

      - name: Check coverage threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
          echo "Total coverage: ${COVERAGE}%"

          THRESHOLD=90
          if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
            echo "ERROR: Coverage ${COVERAGE}% is below threshold ${THRESHOLD}%"
            exit 1
          fi

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          fail_ci_if_error: true
          verbose: true

      - name: Upload coverage report
        uses: actions/upload-artifact@v3
        with:
          name: coverage-report
          path: |
            coverage.out
            coverage.txt
            coverage.html

      - name: Comment coverage on PR
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v6
        with:
          script: |
            const fs = require('fs');
            const coverage = fs.readFileSync('coverage.txt', 'utf8');

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '## Coverage Report\n\n```\n' + coverage + '\n```'
            });
```

### 1.7 Pre-Commit Hooks

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Pre-commit hook to check test coverage

echo "Running tests with coverage check..."

# Run tests
go test -cover ./... > /dev/null 2>&1
TEST_EXIT_CODE=$?

if [ $TEST_EXIT_CODE -ne 0 ]; then
  echo "ERROR: Tests failed. Commit aborted."
  echo "Run 'go test ./...' to see failures."
  exit 1
fi

# Check coverage on changed files
CHANGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' | grep -v '_test\.go$')

if [ -z "$CHANGED_FILES" ]; then
  echo "No Go files changed, skipping coverage check."
  exit 0
fi

for file in $CHANGED_FILES; do
  PACKAGE=$(dirname "$file")
  if [ "$PACKAGE" != "." ]; then
    COVERAGE=$(go test -coverprofile=/tmp/cover.out "./$PACKAGE" 2>/dev/null | \
      grep -oP 'coverage: \K[0-9.]+')

    if [ ! -z "$COVERAGE" ]; then
      if (( $(echo "$COVERAGE < 85" | bc -l) )); then
        echo "WARNING: Coverage for $PACKAGE is ${COVERAGE}% (below 85%)"
      fi
    fi
  fi
done

echo "Coverage check passed!"
exit 0
```

---

## 2. Fuzz Testing

### 2.1 Fuzzing Strategy

#### Fuzzing Targets

| Component | Fuzzing Priority | Tool | Duration |
|-----------|-----------------|------|----------|
| Message handlers | Critical | Go native fuzzing | Continuous |
| Serialization | Critical | Go native fuzzing | Continuous |
| Cryptographic functions | Critical | Go native fuzzing | 24/7 |
| State transitions | High | Custom fuzzer | Daily |
| Query handlers | Medium | Go native fuzzing | Weekly |
| Utility functions | Low | Go native fuzzing | As needed |

### 2.2 Go Native Fuzzing Setup

#### Fuzz Test Example

Create `x/identitychange/types/fuzz_test.go`:

```go
package types_test

import (
    "testing"

    "github.com/aura/chain/x/identitychange/types"
)

// FuzzMsgSubmitIdentityChange fuzzes the MsgSubmitIdentityChange validation
func FuzzMsgSubmitIdentityChange(f *testing.F) {
    // Seed corpus with valid examples
    f.Add("alice", "alice-new", "passport-update")
    f.Add("bob", "bob-v2", "legal-name-change")
    f.Add("", "", "")

    // Fuzz function
    f.Fuzz(func(t *testing.T, oldID, newID, reason string) {
        msg := types.NewMsgSubmitIdentityChange(oldID, newID, reason)

        // Should never panic
        err := msg.ValidateBasic()

        // If validation passes, message should be well-formed
        if err == nil {
            if msg.OldIdentityId == "" || msg.NewIdentityId == "" {
                t.Error("ValidateBasic passed with empty IDs")
            }
        }
    })
}

// FuzzIdentityChangeProto fuzzes protobuf deserialization
func FuzzIdentityChangeProto(f *testing.F) {
    // Seed with valid serialized message
    validMsg := types.IdentityChange{
        OldIdentityId: "alice",
        NewIdentityId: "alice-new",
        Timestamp:     1234567890,
    }
    seed, _ := validMsg.Marshal()
    f.Add(seed)

    f.Fuzz(func(t *testing.T, data []byte) {
        msg := &types.IdentityChange{}

        // Should never panic on invalid input
        _ = msg.Unmarshal(data)
    })
}

// FuzzConfidenceScore fuzzes confidence score calculation
func FuzzConfidenceScore(f *testing.F) {
    f.Add(int64(0), int64(0), int64(0))
    f.Add(int64(5), int64(3), int64(720))
    f.Add(int64(100), int64(50), int64(8760))

    f.Fuzz(func(t *testing.T, verifications, endorsements, ageHours int64) {
        // Should never panic
        score, err := types.CalculateConfidenceScore(
            int(verifications),
            int(endorsements),
            time.Duration(ageHours) * time.Hour,
        )

        // Score should be in valid range
        if err == nil && (score < 0 || score > 100) {
            t.Errorf("Invalid score: %d", score)
        }
    })
}
```

#### Running Fuzz Tests

```bash
# Run fuzz test for 30 seconds
go test -fuzz=FuzzMsgSubmitIdentityChange -fuzztime=30s

# Run until failure or manual stop
go test -fuzz=FuzzMsgSubmitIdentityChange

# Run with specific timeout
go test -fuzz=FuzzIdentityChangeProto -fuzztime=5m

# Run all fuzz tests briefly
go test -fuzz=. -fuzztime=10s ./...
```

### 2.3 Continuous Fuzzing

#### GitHub Actions Fuzzing

Create `.github/workflows/fuzzing.yml`:

```yaml
name: Continuous Fuzzing

on:
  schedule:
    - cron: '0 */6 * * *'  # Every 6 hours
  workflow_dispatch:

jobs:
  fuzz:
    runs-on: ubuntu-latest
    timeout-minutes: 360  # 6 hours

    strategy:
      matrix:
        target:
          - FuzzMsgSubmitIdentityChange
          - FuzzIdentityChangeProto
          - FuzzConfidenceScore
          - FuzzVCRegistration
          - FuzzDEXSwap

    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Restore fuzz corpus
        uses: actions/cache@v3
        with:
          path: |
            **/testdata/fuzz/**/corpus
          key: fuzz-corpus-${{ matrix.target }}-${{ github.sha }}
          restore-keys: |
            fuzz-corpus-${{ matrix.target }}-

      - name: Run fuzzing
        run: |
          go test -fuzz=${{ matrix.target }} -fuzztime=5h ./...
        continue-on-error: true

      - name: Upload crashes
        if: failure()
        uses: actions/upload-artifact@v3
        with:
          name: fuzz-crashes-${{ matrix.target }}
          path: '**/testdata/fuzz/**/crash-*'

      - name: Save fuzz corpus
        uses: actions/cache@v3
        with:
          path: |
            **/testdata/fuzz/**/corpus
          key: fuzz-corpus-${{ matrix.target }}-${{ github.sha }}
```

### 2.4 OSS-Fuzz Integration

#### OSS-Fuzz Configuration

Create `oss-fuzz/Dockerfile`:

```dockerfile
FROM gcr.io/oss-fuzz-base/base-builder-go

# Clone repository
RUN git clone --depth 1 https://github.com/aura-blockchain/aura /src/aura

# Set working directory
WORKDIR /src/aura

# Copy build script
COPY build.sh /src/
```

Create `oss-fuzz/build.sh`:

```bash
#!/bin/bash -eu

# Build fuzz targets
cd /src/aura/chain

# Compile fuzz targets
compile_native_go_fuzzer ./x/identitychange/types FuzzMsgSubmitIdentityChange fuzz_msg_submit_identity_change
compile_native_go_fuzzer ./x/identitychange/types FuzzIdentityChangeProto fuzz_identity_change_proto
compile_native_go_fuzzer ./x/vcregistry/types FuzzVCRegistration fuzz_vc_registration
compile_native_go_fuzzer ./x/dex/types FuzzDEXSwap fuzz_dex_swap

# Copy seed corpus
cp -r /src/aura/chain/testdata/fuzz/* $OUT/
```

### 2.5 Crash Triage Workflow

#### Automated Crash Detection

Create `scripts/triage-fuzz-crashes.sh`:

```bash
#!/bin/bash
# triage-fuzz-crashes.sh - Analyze and categorize fuzz crashes

CRASH_DIR="testdata/fuzz"
REPORT_FILE="fuzz-crash-report.md"

echo "# Fuzz Crash Report" > "$REPORT_FILE"
echo "Generated: $(date)" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# Find all crash files
CRASHES=$(find "$CRASH_DIR" -name "crash-*")

if [ -z "$CRASHES" ]; then
  echo "No crashes found!"
  exit 0
fi

echo "Found $(echo "$CRASHES" | wc -l) crash(es)"
echo ""

for crash in $CRASHES; do
  echo "## Crash: $(basename $crash)" >> "$REPORT_FILE"
  echo "" >> "$REPORT_FILE"

  # Extract fuzz target
  TARGET=$(dirname "$crash" | xargs basename)
  echo "**Fuzz Target**: $TARGET" >> "$REPORT_FILE"

  # Attempt to reproduce
  echo "**Reproduction**:" >> "$REPORT_FILE"
  echo '```' >> "$REPORT_FILE"
  go test -run="$TARGET" -fuzz="$TARGET" -fuzztime=1s 2>&1 || true >> "$REPORT_FILE"
  echo '```' >> "$REPORT_FILE"
  echo "" >> "$REPORT_FILE"

  # Show crash input
  echo "**Crash Input**:" >> "$REPORT_FILE"
  echo '```' >> "$REPORT_FILE"
  xxd "$crash" | head -20 >> "$REPORT_FILE"
  echo '```' >> "$REPORT_FILE"
  echo "" >> "$REPORT_FILE"
done

echo "Crash report generated: $REPORT_FILE"
```

#### Crash Prioritization

| Crash Type | Priority | Response Time |
|------------|----------|---------------|
| Panic in critical module | P0 | Immediate |
| Memory corruption | P0 | < 4 hours |
| DoS via resource exhaustion | P1 | < 24 hours |
| Logic error | P2 | < 1 week |
| Invalid input handling | P3 | < 2 weeks |

### 2.6 Fuzzing Best Practices

1. **Start with seed corpus**: Provide valid examples
2. **Minimize crashes**: Use `-fuzzminimizetime` to reduce crash inputs
3. **Monitor coverage**: Track code coverage growth from fuzzing
4. **Continuous operation**: Run fuzzing 24/7 on CI infrastructure
5. **Corpus management**: Store and version corpus in repository
6. **Quick reproducibility**: Ensure crashes can be reproduced locally

---

## 3. Static Analysis

### 3.1 Static Analysis Tools

#### Tool Overview

| Tool | Purpose | Severity | Frequency |
|------|---------|----------|-----------|
| gosec | Security vulnerabilities | High | Every commit |
| golangci-lint | Code quality & style | Medium | Every commit |
| govulncheck | Dependency vulnerabilities | High | Daily |
| SonarQube | Comprehensive analysis | Medium | Per PR |
| CodeQL | Semantic analysis | High | Weekly |

### 3.2 gosec Configuration

Create `.gosec.json`:

```json
{
  "global": {
    "nosec": false,
    "show-ignored": false,
    "audit-tags": "security,compliance"
  },
  "severity": "medium",
  "confidence": "medium",
  "exclude": [
    "G104",
    "G304"
  ],
  "exclude-dirs": [
    "vendor",
    "testdata",
    "proto"
  ],
  "tests": true
}
```

#### Running gosec

```bash
# Basic scan
gosec ./...

# With JSON output
gosec -fmt=json -out=gosec-report.json ./...

# With SARIF output (for GitHub)
gosec -fmt=sarif -out=gosec.sarif ./...

# Fail on high severity only
gosec -severity=high ./...
```

### 3.3 golangci-lint Configuration

Create `.golangci.yml`:

```yaml
run:
  timeout: 5m
  tests: true
  skip-dirs:
    - vendor
    - proto
    - testdata
  skip-files:
    - ".*\\.pb\\.go$"
    - ".*_test\\.go$"  # Exclude for some linters

linters:
  enable:
    - errcheck        # Check for unchecked errors
    - gosec           # Security issues
    - govet           # Go vet analysis
    - ineffassign     # Detect ineffectual assignments
    - staticcheck     # Static analysis
    - unused          # Unused code
    - gosimple        # Simplification suggestions
    - misspell        # Spelling mistakes
    - gocyclo         # Cyclomatic complexity
    - dupl            # Duplicate code
    - goconst         # Repeated strings
    - goimports       # Import formatting
    - revive          # Replacement for golint
    - stylecheck      # Style issues
    - unconvert       # Unnecessary conversions
    - unparam         # Unused function parameters
    - gofmt           # Format checking
    - bodyclose       # HTTP body close
    - noctx           # HTTP request without context
    - rowserrcheck    # SQL rows.Err check
    - sqlclosecheck   # SQL Close() check
    - exportloopref   # Loop variable export
    - gocritic        # Various checks
    - predeclared     # Predeclared identifier usage
    - whitespace      # Whitespace issues

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true

  govet:
    check-shadowing: true
    enable-all: true

  gocyclo:
    min-complexity: 15

  dupl:
    threshold: 100

  goconst:
    min-len: 3
    min-occurrences: 3

  misspell:
    locale: US

  gosec:
    severity: medium
    confidence: medium
    excludes:
      - G104  # Audit errors not checked

  staticcheck:
    checks: ["all"]

  revive:
    rules:
      - name: var-naming
      - name: exported
      - name: indent-error-flow
      - name: blank-imports
      - name: context-as-argument
      - name: dot-imports
      - name: error-return
      - name: error-strings
      - name: error-naming
      - name: if-return
      - name: increment-decrement
      - name: receiver-naming
      - name: time-naming
      - name: unexported-return
      - name: indent-error-flow

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0

  exclude-rules:
    # Exclude some linters from running on tests files
    - path: _test\.go
      linters:
        - gocyclo
        - dupl
        - gosec

    # Exclude known issues
    - text: "weak cryptographic primitive"
      linters:
        - gosec
      path: "test.*"

output:
  format: colored-line-number
  print-issued-lines: true
  print-linter-name: true
```

#### Running golangci-lint

```bash
# Basic run
golangci-lint run ./...

# With all linters
golangci-lint run --enable-all ./...

# Fast mode (for pre-commit)
golangci-lint run --fast ./...

# Generate report
golangci-lint run --out-format=json > golangci-report.json
```

### 3.4 govulncheck Integration

```bash
# Install
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan for vulnerabilities
govulncheck ./...

# JSON output
govulncheck -json ./... > vulns.json

# Scan only code that actually uses vulnerable functions
govulncheck -mode=binary ./cmd/aurad
```

### 3.5 SonarQube Setup

#### sonar-project.properties

```properties
# Project identification
sonar.projectKey=aura-blockchain
sonar.projectName=Aura Blockchain
sonar.projectVersion=1.0

# Source code
sonar.sources=.
sonar.exclusions=**/vendor/**,**/testdata/**,**/*.pb.go,**/node_modules/**

# Test sources
sonar.tests=.
sonar.test.inclusions=**/*_test.go

# Coverage
sonar.go.coverage.reportPaths=coverage/coverage.out

# Test execution
sonar.go.tests.reportPaths=test-report.json

# External reports
sonar.go.golangci-lint.reportPaths=golangci-report.xml
sonar.go.govet.reportPaths=govet-report.xml

# Language
sonar.language=go
sonar.sourceEncoding=UTF-8

# Quality gates
sonar.qualitygate.wait=true
```

#### Running SonarQube

```bash
# Install sonar-scanner
# Download from https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/

# Run analysis
sonar-scanner \
  -Dsonar.host.url=http://localhost:9000 \
  -Dsonar.login=$SONAR_TOKEN

# Or via Docker
docker run \
  --rm \
  -e SONAR_HOST_URL="http://sonarqube:9000" \
  -e SONAR_LOGIN="$SONAR_TOKEN" \
  -v "$(pwd):/usr/src" \
  sonarsource/sonar-scanner-cli
```

### 3.6 CodeQL Configuration

Create `.github/workflows/codeql.yml`:

```yaml
name: CodeQL Analysis

on:
  push:
    branches: [main, master, develop]
  pull_request:
    branches: [main, master]
  schedule:
    - cron: '0 6 * * 1'  # Weekly on Monday

jobs:
  analyze:
    name: Analyze
    runs-on: ubuntu-latest
    permissions:
      actions: read
      contents: read
      security-events: write

    strategy:
      fail-fast: false
      matrix:
        language: ['go']

    steps:
      - name: Checkout repository
        uses: actions/checkout@v3

      - name: Initialize CodeQL
        uses: github/codeql-action/init@v2
        with:
          languages: ${{ matrix.language }}
          queries: security-and-quality

      - name: Autobuild
        uses: github/codeql-action/autobuild@v2

      - name: Perform CodeQL Analysis
        uses: github/codeql-action/analyze@v2
        with:
          category: "/language:${{ matrix.language }}"
```

### 3.7 Integrated Static Analysis Workflow

Create `.github/workflows/static-analysis.yml`:

```yaml
name: Static Analysis

on:
  push:
    branches: [main, master, develop]
  pull_request:
    branches: [main, master]

jobs:
  gosec:
    name: Security Scan (gosec)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run gosec
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt=sarif -out=gosec.sarif ./...'

      - name: Upload SARIF file
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: gosec.sarif

  golangci-lint:
    name: Code Quality (golangci-lint)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m

  govulncheck:
    name: Vulnerability Scan (govulncheck)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest

      - name: Run govulncheck
        run: govulncheck ./...

  report:
    name: Analysis Report
    needs: [gosec, golangci-lint, govulncheck]
    runs-on: ubuntu-latest
    if: always()
    steps:
      - name: Generate summary
        run: |
          echo "# Static Analysis Summary" >> $GITHUB_STEP_SUMMARY
          echo "All static analysis checks completed." >> $GITHUB_STEP_SUMMARY
```

---

## 4. Integration and Automation

### 4.1 Pre-Commit Hooks

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Pre-commit hook for code quality checks

set -e

echo "Running pre-commit checks..."

# 1. Format check
echo "Checking code formatting..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
  echo "ERROR: The following files are not formatted:"
  echo "$UNFORMATTED"
  echo "Run: gofmt -w ."
  exit 1
fi

# 2. Quick lint
echo "Running quick lint..."
golangci-lint run --fast --timeout=2m ./... || {
  echo "ERROR: Linting failed. Run 'golangci-lint run ./...' to see details."
  exit 1
}

# 3. Security scan
echo "Running security scan..."
gosec -quiet ./... || {
  echo "ERROR: Security issues found. Run 'gosec ./...' to see details."
  exit 1
}

# 4. Tests
echo "Running tests..."
go test -short ./... > /dev/null || {
  echo "ERROR: Tests failed. Run 'go test ./...' to see failures."
  exit 1
}

echo "All pre-commit checks passed!"
```

### 4.2 GitHub Pull Request Checks

Create `.github/workflows/pr-checks.yml`:

```yaml
name: Pull Request Checks

on:
  pull_request:
    branches: [main, master]

jobs:
  quality-gate:
    name: Quality Gate
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # Full history for better analysis

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

      - name: Run tests with coverage
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
          echo "COVERAGE=$COVERAGE" >> $GITHUB_ENV

      - name: Check coverage threshold
        run: |
          if (( $(echo "$COVERAGE < 90" | bc -l) )); then
            echo "::error::Coverage $COVERAGE% is below 90% threshold"
            exit 1
          fi

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

      - name: Run gosec
        uses: securego/gosec@master
        with:
          args: '-fmt=json -out=gosec-report.json ./...'

      - name: Check for high severity issues
        run: |
          HIGH_COUNT=$(jq '[.Issues[] | select(.Severity == "HIGH")] | length' gosec-report.json)
          if [ "$HIGH_COUNT" -gt "0" ]; then
            echo "::error::Found $HIGH_COUNT high severity security issues"
            exit 1
          fi

      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      - name: Comment PR with results
        if: always()
        uses: actions/github-script@v6
        with:
          script: |
            const coverage = process.env.COVERAGE;
            const body = `## Quality Gate Results

            - ✅ Tests: Passed
            - 📊 Coverage: ${coverage}%
            - 🔍 Linting: Passed
            - 🔒 Security: Passed
            - 🆘 Vulnerabilities: None

            All quality checks passed! Ready for review.`;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body
            });
```

### 4.3 Continuous Integration Pipeline

Create `.github/workflows/ci.yml`:

```yaml
name: Continuous Integration

on:
  push:
    branches: [main, master, develop]
  pull_request:
    branches: [main, master]

jobs:
  test:
    name: Test
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ['1.20', '1.21']

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go }}

      - name: Run tests
        run: go test -v -race ./...

  coverage:
    name: Coverage
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Generate coverage
        run: |
          ./scripts/run-coverage.sh

      - name: Upload to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage/coverage.out

  lint:
    name: Lint
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3

  security:
    name: Security
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3
      - uses: securego/gosec@master
        with:
          args: './...'

  build:
    name: Build
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build
        run: make build
```

---

## 5. Quality Metrics and Reporting

### 5.1 Quality Metrics Dashboard

#### Key Metrics

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Test Coverage | 90%+ | - | 🔴 |
| Critical Module Coverage | 95%+ | - | 🔴 |
| Linter Issues | 0 | - | 🔴 |
| Security Issues (High) | 0 | - | 🔴 |
| Cyclomatic Complexity | < 15 avg | - | 🔴 |
| Code Duplication | < 3% | - | 🔴 |
| Known Vulnerabilities | 0 | - | 🔴 |
| Fuzz Crashes | 0 | - | 🔴 |

### 5.2 Weekly Quality Report

Create `scripts/generate-quality-report.sh`:

```bash
#!/bin/bash
# generate-quality-report.sh - Generate weekly quality report

REPORT_FILE="quality-report-$(date +%Y-%m-%d).md"

echo "# Code Quality Report" > "$REPORT_FILE"
echo "Generated: $(date)" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# Coverage
echo "## Test Coverage" >> "$REPORT_FILE"
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# Linting
echo "## Linting Issues" >> "$REPORT_FILE"
golangci-lint run ./... >> "$REPORT_FILE" 2>&1 || true
echo "" >> "$REPORT_FILE"

# Security
echo "## Security Issues" >> "$REPORT_FILE"
gosec -fmt=text ./... >> "$REPORT_FILE" 2>&1 || true
echo "" >> "$REPORT_FILE"

# Vulnerabilities
echo "## Dependency Vulnerabilities" >> "$REPORT_FILE"
govulncheck ./... >> "$REPORT_FILE" 2>&1 || true
echo "" >> "$REPORT_FILE"

# Fuzz crashes
echo "## Fuzz Testing" >> "$REPORT_FILE"
CRASHES=$(find testdata/fuzz -name "crash-*" 2>/dev/null | wc -l)
echo "Crash count: $CRASHES" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

echo "Quality report generated: $REPORT_FILE"
```

---

## 6. Continuous Improvement

### 6.1 Monthly Reviews

- Review quality metrics trends
- Identify recurring issues
- Update tooling configurations
- Adjust coverage targets
- Plan remediation sprints

### 6.2 Tool Updates

```bash
# Update all tools monthly
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### 6.3 Team Training

- Quarterly security training
- Best practices workshops
- Tool usage sessions
- Knowledge sharing

---

## Appendix: Quick Reference

### Essential Commands

```bash
# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Fuzzing
go test -fuzz=FuzzMyFunction -fuzztime=30s

# Security scan
gosec ./...

# Linting
golangci-lint run ./...

# Vulnerabilities
govulncheck ./...

# All checks
./scripts/run-coverage.sh && golangci-lint run ./... && gosec ./...
```

### CI/CD Badge Examples

```markdown
![Coverage](https://codecov.io/gh/aura-blockchain/aura/branch/main/graph/badge.svg)
![Go Report](https://goreportcard.com/badge/github.com/aura-blockchain/aura)
![Security](https://img.shields.io/badge/security-gosec-blue)
```

---

**Document Version**: 1.0
**Last Updated**: 2025-11-13
**Maintained By**: Aura Engineering Team
