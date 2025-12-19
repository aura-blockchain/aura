#!/usr/bin/env bash
set -euo pipefail

# Runs full test suite plus fuzz corpus replay via make test-with-fuzz.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT_DIR}/chain"

echo "Running full test suite + fuzz corpus replay..."
make test-with-fuzz
