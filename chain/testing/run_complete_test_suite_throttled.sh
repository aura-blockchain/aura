#!/bin/bash
#
# Complete Test Suite Runner with CPU Throttling
# This script runs all tests with CPU throttling to ensure stable completion
# even under resource constraints, and continues running if the user logs off.
#
# Usage: ./run_complete_test_suite_throttled.sh [cpu_limit_percent]
#
# Example: ./run_complete_test_suite_throttled.sh 50  # Limit to 50% CPU

set -e

# Configuration
CPU_LIMIT="${1:-75}"  # Default to 75% CPU if not specified
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_DIR="/home/hudson/blockchain-projects/aura/chain/testing/test-logs"
MAIN_LOG="$LOG_DIR/complete_suite_${TIMESTAMP}.log"
SUMMARY_LOG="$LOG_DIR/summary_${TIMESTAMP}.log"

# Create log directory
mkdir -p "$LOG_DIR"

# Function to log with timestamp
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$MAIN_LOG"
}

# Function to run tests with CPU throttling using cpulimit
run_with_cpulimit() {
    local test_name="$1"
    local test_cmd="$2"

    log "Starting: $test_name"
    log "Command: $test_cmd"

    # Run the command with cpulimit if available, otherwise run directly
    if command -v cpulimit &> /dev/null; then
        log "Running with CPU limit: ${CPU_LIMIT}%"
        # Use timeout to prevent infinite hangs, and cpulimit to throttle CPU
        timeout 3600 cpulimit -l "$CPU_LIMIT" -f -- bash -c "cd /home/hudson/blockchain-projects/aura/chain && source ../env.sh && $test_cmd" 2>&1 | tee -a "$MAIN_LOG"
        local exit_code=${PIPESTATUS[0]}
    else
        log "cpulimit not found, running without CPU throttling"
        bash -c "cd /home/hudson/blockchain-projects/aura/chain && source ../env.sh && $test_cmd" 2>&1 | tee -a "$MAIN_LOG"
        local exit_code=$?
    fi

    if [ $exit_code -eq 0 ]; then
        log "✓ PASSED: $test_name"
        return 0
    else
        log "✗ FAILED: $test_name (exit code: $exit_code)"
        return 1
    fi
}

# Start
log "========================================="
log "Aura Blockchain - Complete Test Suite"
log "CPU Throttling: ${CPU_LIMIT}%"
log "Log File: $MAIN_LOG"
log "========================================="
log ""

# Check for cpulimit
if ! command -v cpulimit &> /dev/null; then
    log "WARNING: cpulimit not installed. Installing..."
    sudo apt-get update && sudo apt-get install -y cpulimit 2>&1 | tee -a "$MAIN_LOG" || log "Failed to install cpulimit, continuing without throttling"
fi

# Initialize counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test 1: Unit Tests (all packages)
log ""
log "===== Test 1/5: Unit Tests (all packages) ====="
if run_with_cpulimit "Unit Tests" "go test ./... -v -count=1 -timeout=30m"; then
    ((TESTS_PASSED++))
else
    ((TESTS_FAILED++))
fi
((TESTS_RUN++))

# Test 2: Integration Tests
log ""
log "===== Test 2/5: Integration Tests ====="
if run_with_cpulimit "Integration Tests" "go test -tags=integration ./... -v -count=1 -timeout=30m"; then
    ((TESTS_PASSED++))
else
    ((TESTS_FAILED++))
fi
((TESTS_RUN++))

# Test 3: Crypto Primitives
log ""
log "===== Test 3/5: Crypto Primitives ====="
if run_with_cpulimit "Crypto Primitives" "go test ./testing/integration/crypto_primitives_test.go -v"; then
    ((TESTS_PASSED++))
else
    ((TESTS_FAILED++))
fi
((TESTS_RUN++))

# Test 4: Encoding Primitives
log ""
log "===== Test 4/5: Encoding Primitives ====="
if run_with_cpulimit "Encoding Primitives" "go test ./testing/integration/encoding_primitives_test.go -v"; then
    ((TESTS_PASSED++))
else
    ((TESTS_FAILED++))
fi
((TESTS_RUN++))

# Test 5: Race Detection (on critical packages)
log ""
log "===== Test 5/5: Race Detection (critical packages) ====="
if run_with_cpulimit "Race Detection" "go test -race ./app/... ./x/security/... ./x/bridge/... ./x/dex/... -timeout=30m"; then
    ((TESTS_PASSED++))
else
    ((TESTS_FAILED++))
fi
((TESTS_RUN++))

# Generate Summary
log ""
log "========================================="
log "Test Suite Complete"
log "========================================="
log "Tests Run:    $TESTS_RUN"
log "Tests Passed: $TESTS_PASSED"
log "Tests Failed: $TESTS_FAILED"
log "Success Rate: $((TESTS_PASSED * 100 / TESTS_RUN))%"
log ""
log "Main Log:    $MAIN_LOG"
log "========================================="

# Create summary file
cat > "$SUMMARY_LOG" <<EOF
Aura Blockchain - Complete Test Suite Summary
==============================================
Date: $(date)
CPU Limit: ${CPU_LIMIT}%

Results:
--------
Tests Run:    $TESTS_RUN
Tests Passed: $TESTS_PASSED
Tests Failed: $TESTS_FAILED
Success Rate: $((TESTS_PASSED * 100 / TESTS_RUN))%

Main Log: $MAIN_LOG

Test Details:
-------------
1. Unit Tests: $(grep -c "✓ PASSED: Unit Tests" "$MAIN_LOG" && echo "PASSED" || echo "FAILED")
2. Integration Tests: $(grep -c "✓ PASSED: Integration Tests" "$MAIN_LOG" && echo "PASSED" || echo "FAILED")
3. Crypto Primitives: $(grep -c "✓ PASSED: Crypto Primitives" "$MAIN_LOG" && echo "PASSED" || echo "FAILED")
4. Encoding Primitives: $(grep -c "✓ PASSED: Encoding Primitives" "$MAIN_LOG" && echo "PASSED" || echo "FAILED")
5. Race Detection: $(grep -c "✓ PASSED: Race Detection" "$MAIN_LOG" && echo "PASSED" || echo "FAILED")
EOF

log "Summary saved to: $SUMMARY_LOG"

# Exit with appropriate code
if [ $TESTS_FAILED -eq 0 ]; then
    log "✓ ALL TESTS PASSED"
    exit 0
else
    log "✗ SOME TESTS FAILED"
    exit 1
fi
