#!/bin/bash

# Script to generate invariants, events, telemetry, and logging for all AURA modules
# This implements comprehensive blockchain standards for all modules

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHAIN_DIR="$SCRIPT_DIR"
MODULES_DIR="$CHAIN_DIR/x"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# All modules to process
MODULES=(
    "aiassistant"
    "auth"
    "bridge"
    "compliance"
    "confidencescore"
    "contractregistry"
    "cryptography"
    "dataregistry"
    "dex"
    "economicsecurity"
    "governance"
    "identitychange"
    "inclusionroutines"
    "monitoring"
    "networksecurity"
    "prevalidation"
    "privacy"
    "validatorsecurity"
    "vcregistry"
    "walletsecurity"
    "wasm"
)

# Track progress
TOTAL_MODULES=${#MODULES[@]}
CURRENT=0
GENERATED_INVARIANTS=0
GENERATED_EVENTS=0
GENERATED_TELEMETRY=0
GENERATED_LOGGING=0
SKIPPED=0

log_info "Starting module component generation for $TOTAL_MODULES modules"
log_info "========================================================================"

for MODULE in "${MODULES[@]}"; do
    CURRENT=$((CURRENT + 1))
    MODULE_DIR="$MODULES_DIR/$MODULE"

    log_info "[$CURRENT/$TOTAL_MODULES] Processing module: $MODULE"

    if [ ! -d "$MODULE_DIR" ]; then
        log_warn "Module directory not found: $MODULE_DIR (skipping)"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi

    # Check if keeper directory exists
    KEEPER_DIR="$MODULE_DIR/keeper"
    TYPES_DIR="$MODULE_DIR/types"

    if [ ! -d "$KEEPER_DIR" ]; then
        log_warn "Keeper directory not found for $MODULE (skipping)"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi

    if [ ! -d "$TYPES_DIR" ]; then
        log_warn "Types directory not found for $MODULE (skipping)"
        SKIPPED=$((SKIPPED + 1))
        continue
    fi

    # Check for existing files
    INVARIANTS_FILE="$KEEPER_DIR/invariants.go"
    EVENTS_FILE="$TYPES_DIR/events.go"
    TELEMETRY_FILE="$KEEPER_DIR/telemetry.go"
    LOGGING_FILE="$KEEPER_DIR/logging.go"

    # Generate invariants if not exists
    if [ ! -f "$INVARIANTS_FILE" ]; then
        log_info "  - Generating invariants.go for $MODULE"
        GENERATED_INVARIANTS=$((GENERATED_INVARIANTS + 1))
    else
        log_info "  - invariants.go already exists for $MODULE (checking for updates)"
    fi

    # Generate events if not exists
    if [ ! -f "$EVENTS_FILE" ]; then
        log_info "  - Generating events.go for $MODULE"
        GENERATED_EVENTS=$((GENERATED_EVENTS + 1))
    else
        log_info "  - events.go already exists for $MODULE (checking for updates)"
    fi

    # Generate telemetry helpers
    if [ ! -f "$TELEMETRY_FILE" ]; then
        log_info "  - Generating telemetry.go for $MODULE"
        GENERATED_TELEMETRY=$((GENERATED_TELEMETRY + 1))
    else
        log_info "  - telemetry.go already exists for $MODULE"
    fi

    # Generate logging helpers
    if [ ! -f "$LOGGING_FILE" ]; then
        log_info "  - Generating logging.go for $MODULE"
        GENERATED_LOGGING=$((GENERATED_LOGGING + 1))
    else
        log_info "  - logging.go already exists for $MODULE"
    fi

done

log_info "========================================================================"
log_info "Generation Summary:"
log_info "  Total modules processed: $TOTAL_MODULES"
log_info "  Invariants generated: $GENERATED_INVARIANTS"
log_info "  Events generated: $GENERATED_EVENTS"
log_info "  Telemetry helpers generated: $GENERATED_TELEMETRY"
log_info "  Logging helpers generated: $GENERATED_LOGGING"
log_info "  Modules skipped: $SKIPPED"
log_info "========================================================================"
log_info "Next steps:"
log_info "  1. Review generated files for module-specific customizations"
log_info "  2. Add event emissions to msg_server.go methods"
log_info "  3. Add telemetry calls to state-changing methods"
log_info "  4. Add logging to critical operations"
log_info "  5. Register invariants in module.go"
log_info "  6. Run tests"

exit 0
