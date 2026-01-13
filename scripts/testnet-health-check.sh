#!/bin/bash
# ============================================================================
# AURA Testnet Health Check (Public vs Local)
# ============================================================================
# Defaults to PUBLIC testnet health check (aura-mvp-1).
# Use --local to target the local 4-validator docker testnet.
#
# Usage:
#   ./testnet-health-check.sh [--json] [--quiet] [--local]
# ============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TARGET="public"
PASS_THROUGH=()

while [[ $# -gt 0 ]]; do
    case "$1" in
        --local)
            TARGET="local"
            shift
            ;;
        --public)
            TARGET="public"
            shift
            ;;
        *)
            PASS_THROUGH+=("$1")
            shift
            ;;
    esac
done

if [[ "$TARGET" == "local" ]]; then
    exec "${SCRIPT_DIR}/health-check-all.sh" "${PASS_THROUGH[@]}"
fi

exec "${SCRIPT_DIR}/public-testnet-health-check.sh" "${PASS_THROUGH[@]}"
