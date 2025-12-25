# Internal Package

## Overview

The Internal package contains security tests and invariant checks that enforce safety properties across all Aura blockchain modules. It is NOT a Cosmos SDK module but a testing infrastructure package.

## Purpose

This package ensures that:
- No module uses `unsafe` package (memory safety)
- No module uses `time.Now()` (determinism)
- No module uses `math/rand` without proper seeding (determinism)
- Cross-module interactions follow security best practices

## Tests

### Unsafe Usage Check (`tests/unsafe_check_test.go`)
Scans all module code to ensure no usage of Go's `unsafe` package.

**Why**: The `unsafe` package can:
- Violate type safety and memory safety
- Lead to undefined behavior and crashes
- Create security vulnerabilities
- Break consensus through non-deterministic behavior

**Enforcement**: This test fails if any non-test Go file in `chain/x/` imports `unsafe`.

## Usage

Run internal tests:

```bash
cd /home/hudson/blockchain-projects/aura/chain/x/internal
go test ./... -v
```

These tests run automatically in CI/CD to prevent unsafe code from being merged.

## Contributing

When adding new modules:
1. Do NOT use `unsafe` package
2. Do NOT use `time.Now()` (use `common/determinism.GetBlockTime()`)
3. Do NOT use unseeded `math/rand` (use `common/determinism` randomness)
4. Follow Cosmos SDK best practices for deterministic state machines

All violations will be caught by these tests.
