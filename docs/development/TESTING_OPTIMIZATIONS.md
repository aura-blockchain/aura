# Go Test Performance Optimizations

## Quick Test Commands (Fastest to Slowest)

### 1. **Fastest: Short Tests Only** (~2-5 minutes)
```bash
make test-short
# OR
go test -v -short -parallel=8 -timeout=10m ./...
```
Use this when you want quick feedback. Skips integration tests.

### 2. **Fast: Parallel Tests** (~5-10 minutes)
```bash
make test-fast
# OR
go test -v -parallel=8 -timeout=20m -count=1 ./...
```
Runs all tests with 8 parallel workers. Doesn't use cache.

### 3. **Cached: Fast Reruns** (~3-8 minutes after first run)
```bash
make test-cached
# OR
go test -v -parallel=8 -timeout=20m ./...
```
Uses Go test cache. Much faster on subsequent runs.

### 4. **Standard: Default Tests** (~10-15 minutes)
```bash
make test
# OR
go test -v -parallel=4 -timeout=30m ./...
```
Updated default with parallelization.

### 5. **Thorough: Race Detection** (~20-30 minutes)
```bash
make test-race
# OR
go test -v -race -parallel=4 -timeout=30m ./...
```
Finds race conditions. Slower but thorough.

## Test Specific Packages

To test only one module:
```bash
make test-pkg PKG=./x/aura
# OR
go test -v -parallel=4 -timeout=10m ./x/aura
```

## For Agents with 10-Minute Timeouts

**Recommended approach:**
1. Use `make test-short` (usually finishes in <10 mins)
2. Or break tests by package:
```bash
# List all packages with tests
go list -f '{{.Dir}}' ./... | grep -E 'x/|app/|cmd/'

# Test each major package separately
go test -v -parallel=4 -timeout=10m ./x/aura/...
go test -v -parallel=4 -timeout=10m ./x/dex/...
go test -v -parallel=4 -timeout=10m ./app/...
```

## Performance Tips

1. **Use build cache**: Go caches test results. Only changed tests re-run.
2. **Parallel workers**: `-parallel=N` runs N tests simultaneously
3. **Short tests**: Add `-short` flag to skip long integration tests
4. **Specific packages**: Test only what changed
5. **Disable verbose**: Remove `-v` for less output (faster)

## Environment Variables Set

The following are now in your `~/.bashrc`:
- `GOMAXPROCS=3` - Uses all WSL2 processors
- `GOFLAGS="-buildvcs=false"` - Faster builds

## Troubleshooting Timeouts

If tests still timeout:
1. Identify slow tests: `go test -v ./... | grep -E '(PASS|FAIL).*[0-9]+\..*s'`
2. Run them separately with higher timeout
3. Consider using `t.Parallel()` in test functions
4. Check for tests waiting on network/database

## Original Makefile Backup

A backup was saved at: `Makefile.backup`
