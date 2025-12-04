#!/bin/bash
# ============================================================================
# AURA Testnet Continuous Monitoring Automation
# ============================================================================
# Runs periodic health checks and logs results for long-running test sessions
#
# Usage:
#   ./scripts/continuous-monitor.sh [interval_seconds] [output_file]
#
# Example:
#   ./scripts/continuous-monitor.sh 60 monitoring.log
#   (checks every 60 seconds, logs to monitoring.log)
# ============================================================================

set -euo pipefail

# Configuration
readonly INTERVAL=${1:-60}  # Default: check every 60 seconds
readonly OUTPUT_FILE=${2:-testnet-monitoring.log}
readonly MONITOR_SCRIPT="./scripts/testnet-monitor.sh"

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Counters
CHECKS_RUN=0
HEALTHY_COUNT=0
DEGRADED_COUNT=0
DOWN_COUNT=0
ERROR_COUNT=0

# ============================================================================
# Logging Functions
# ============================================================================

log() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[${timestamp}] $*" | tee -a "$OUTPUT_FILE"
}

log_header() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "" | tee -a "$OUTPUT_FILE"
    echo "==============================================================================" | tee -a "$OUTPUT_FILE"
    echo "[${timestamp}] $*" | tee -a "$OUTPUT_FILE"
    echo "==============================================================================" | tee -a "$OUTPUT_FILE"
}

# ============================================================================
# Monitoring Functions
# ============================================================================

run_health_check() {
    CHECKS_RUN=$((CHECKS_RUN + 1))

    log_header "Health Check #${CHECKS_RUN}"

    # Run quick health check
    if ${MONITOR_SCRIPT} quick >> "$OUTPUT_FILE" 2>&1; then
        log "✓ Status: HEALTHY"
        HEALTHY_COUNT=$((HEALTHY_COUNT + 1))
        return 0
    else
        local exit_code=$?
        if [[ $exit_code -eq 1 ]]; then
            log "⚠ Status: DEGRADED"
            DEGRADED_COUNT=$((DEGRADED_COUNT + 1))
        else
            log "✗ Status: DOWN"
            DOWN_COUNT=$((DOWN_COUNT + 1))
        fi
        return $exit_code
    fi
}

run_detailed_check() {
    log_header "Detailed Monitoring Pass"

    # Network health
    log "Running network health check..."
    ${MONITOR_SCRIPT} network >> "$OUTPUT_FILE" 2>&1 || true

    # Performance metrics
    log "Collecting performance metrics..."
    ${MONITOR_SCRIPT} performance >> "$OUTPUT_FILE" 2>&1 || true

    # Log analysis
    log "Scanning logs for errors..."
    ${MONITOR_SCRIPT} check-logs all 50 >> "$OUTPUT_FILE" 2>&1 || true
}

run_error_detection() {
    log "Scanning for critical errors..."

    local critical_errors=0

    # Check for consensus failures
    if docker-compose -f docker-compose.testnet.yml logs --tail=100 2>&1 | grep -qi "consensus failure"; then
        log "✗ CRITICAL: Consensus failure detected"
        critical_errors=$((critical_errors + 1))
    fi

    # Check for double signing
    if docker-compose -f docker-compose.testnet.yml logs --tail=100 2>&1 | grep -qi "double.*sign"; then
        log "✗ CRITICAL: Double signing detected"
        critical_errors=$((critical_errors + 1))
    fi

    # Check for panics
    if docker-compose -f docker-compose.testnet.yml logs --tail=100 2>&1 | grep -qi "panic"; then
        log "✗ ERROR: Panic detected in logs"
        critical_errors=$((critical_errors + 1))
    fi

    if [[ $critical_errors -gt 0 ]]; then
        ERROR_COUNT=$((ERROR_COUNT + 1))
        log "✗ Total critical errors detected: ${critical_errors}"
        return 1
    else
        log "✓ No critical errors detected"
        return 0
    fi
}

print_summary() {
    log_header "Monitoring Summary"
    log "Total Checks Run:      ${CHECKS_RUN}"
    log "Healthy:               ${HEALTHY_COUNT}"
    log "Degraded:              ${DEGRADED_COUNT}"
    log "Down:                  ${DOWN_COUNT}"
    log "Critical Errors:       ${ERROR_COUNT}"

    if [[ $CHECKS_RUN -gt 0 ]]; then
        local uptime_percent=$(awk "BEGIN {printf \"%.1f\", ($HEALTHY_COUNT * 100.0) / $CHECKS_RUN}")
        log "Uptime:                ${uptime_percent}%"
    fi
}

# ============================================================================
# Signal Handling
# ============================================================================

cleanup() {
    echo ""
    log "Received interrupt signal, stopping monitoring..."
    print_summary
    log "Monitoring stopped. Full log saved to: ${OUTPUT_FILE}"
    exit 0
}

trap cleanup SIGINT SIGTERM

# ============================================================================
# Main Monitoring Loop
# ============================================================================

main() {
    echo -e "${BLUE}AURA Testnet Continuous Monitoring${NC}"
    echo -e "${BLUE}====================================${NC}"
    echo ""
    echo -e "Interval:     ${INTERVAL} seconds"
    echo -e "Log file:     ${OUTPUT_FILE}"
    echo -e "Started:      $(date)"
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to stop monitoring and see summary${NC}"
    echo ""

    # Initialize log file
    log_header "Monitoring Session Started"
    log "Interval: ${INTERVAL} seconds"
    log "Monitor script: ${MONITOR_SCRIPT}"

    local iteration=0
    local next_detailed_check=5  # Run detailed check every 5 iterations

    while true; do
        iteration=$((iteration + 1))

        # Basic health check every iteration
        run_health_check

        # Error detection
        run_error_detection || true

        # Detailed check periodically
        if [[ $((iteration % next_detailed_check)) -eq 0 ]]; then
            run_detailed_check
        fi

        # Print current stats to console
        echo -e "\r${GREEN}Checks: ${CHECKS_RUN}${NC} | ${GREEN}Healthy: ${HEALTHY_COUNT}${NC} | ${YELLOW}Degraded: ${DEGRADED_COUNT}${NC} | ${RED}Down: ${DOWN_COUNT}${NC} | ${RED}Errors: ${ERROR_COUNT}${NC} | Last: $(date '+%H:%M:%S')\c"

        # Sleep until next check
        sleep "$INTERVAL"
    done
}

# ============================================================================
# Entry Point
# ============================================================================

# Verify monitor script exists
if [[ ! -x "$MONITOR_SCRIPT" ]]; then
    echo "Error: Monitor script not found or not executable: ${MONITOR_SCRIPT}"
    exit 1
fi

main
