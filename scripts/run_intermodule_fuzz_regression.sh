#!/usr/bin/env bash
set -euo pipefail

# Replays all stored fuzz corpora for inter-module guard suites.
# Use this in regression runs to ensure prior discoveries stay fixed.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "${REPO_ROOT}/chain"

# Avoid flaky failures if a background cleanup job prunes the repo-local build cache
# while the linker is running. Use a dedicated temp GOCACHE for this run.
GOCACHE_DIR="$(mktemp -d)"
trap 'rm -rf "${GOCACHE_DIR}"' EXIT
export GOCACHE="${GOCACHE_DIR}"

# Go requires -fuzz to match exactly one fuzz target. For regression we want to
# replay *all* stored corpora, so iterate targets explicitly with -fuzztime=0
# (replay corpus only, no additional random fuzzing).
FUZZ_TARGETS="$(go test ./x/common/testing -list '^Fuzz' | rg '^Fuzz')"
if [ -z "${FUZZ_TARGETS}" ]; then
  echo "No fuzz targets found in ./x/common/testing"
  exit 1
fi

while IFS= read -r target; do
  echo "Replaying corpus for ${target}..."
  # Go 1.24 does not accept zero fuzztime; keep this minimal so we replay corpora
  # and do a very small amount of additional fuzzing.
  go test ./x/common/testing -run=^$ -fuzz="^${target}$" -fuzztime="${FUZZTIME:-1s}"
done <<< "${FUZZ_TARGETS}"
