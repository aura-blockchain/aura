# Intermodule Guard Testing

- **Targeted fuzz replay (all targets):** `../scripts/run_intermodule_fuzz_regression.sh` (iterates fuzz targets)
- **Full suite + corpora:** `make test-with-fuzz` (wraps tests and fuzz corpus replay).
- Corpora live under `x/common/testing/testdata/` to keep regressions deterministic.
- **Replay specific seed:** `go test ./x/common/testing -run Fuzz<Case>/seed#0`
- **Replay failing corpus file:** `go test ./x/common/testing -run Fuzz<Case>/<hash>`
