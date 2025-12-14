#!/bin/bash
# ============================================================================
# AURA Blockchain Soak Testing - Extended Duration Testing
# ============================================================================
# Tests blockchain stability, performance, and resource usage over 24-48 hours
#
# Usage:
#   ./scripts/soak-test.sh [duration_hours] [output_dir]
#
# Example:
#   ./scripts/soak-test.sh 24 soak-test-results
#   (runs for 24 hours, saves results to soak-test-results/)
# ============================================================================

set -euo pipefail

# Configuration
readonly DURATION_HOURS=${1:-24}
readonly OUTPUT_DIR=${2:-"soak-test-results-$(date +%Y%m%d-%H%M%S)"}
readonly CHECK_INTERVAL=300  # 5 minutes
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Metrics tracking
TOTAL_CHECKS=0
FAILED_CHECKS=0
START_HEIGHT=0
START_TIME=$(date +%s)
BASELINE_MEMORY=()
BASELINE_CPU=()

# ============================================================================
# Setup and Initialization
# ============================================================================

setup() {
    echo -e "${BLUE}=== AURA Blockchain Soak Testing ===${NC}"
    echo -e "Duration:     ${DURATION_HOURS} hours"
    echo -e "Interval:     ${CHECK_INTERVAL} seconds ($(($CHECK_INTERVAL / 60)) minutes)"
    echo -e "Output dir:   ${OUTPUT_DIR}"
    echo ""

    # Create output directory
    mkdir -p "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR/metrics"
    mkdir -p "$OUTPUT_DIR/logs"
    mkdir -p "$OUTPUT_DIR/snapshots"

    # Initialize log files
    echo "Soak Test Started: $(date)" > "$OUTPUT_DIR/soak-test.log"
    echo "Duration: ${DURATION_HOURS} hours" >> "$OUTPUT_DIR/soak-test.log"
    echo "Check Interval: ${CHECK_INTERVAL} seconds" >> "$OUTPUT_DIR/soak-test.log"
    echo "" >> "$OUTPUT_DIR/soak-test.log"

    # Check testnet is running
    if ! docker ps --filter "name=aura-validator-1" --format "{{.Names}}" | grep -q "^aura-validator-1$"; then
        echo -e "${RED}ERROR: Testnet is not running${NC}"
        exit 1
    fi

    # Get baseline metrics
    echo -e "${GREEN}Collecting baseline metrics...${NC}"
    START_HEIGHT=$(curl -s localhost:27657/status 2>&1 | jq -r '.result.sync_info.latest_block_height // "0"')
    collect_baseline_metrics

    echo -e "${GREEN}Starting height: ${START_HEIGHT}${NC}"
    echo "Start Height: ${START_HEIGHT}" >> "$OUTPUT_DIR/soak-test.log"
    echo ""
}

collect_baseline_metrics() {
    local validators=("aura-validator-1" "aura-validator-2" "aura-validator-3" "aura-validator-4")

    for val in "${validators[@]}"; do
        if docker ps --format "{{.Names}}" | grep -q "^${val}$"; then
            local mem=$(docker stats --no-stream --format "{{.MemUsage}}" "$val" | awk '{print $1}')
            local cpu=$(docker stats --no-stream --format "{{.CPUPerc}}" "$val" | awk -F'%' '{print $1}')
            BASELINE_MEMORY+=("$val:$mem")
            BASELINE_CPU+=("$val:$cpu")
        fi
    done
}

# ============================================================================
# Logging Functions
# ============================================================================

log() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[${timestamp}] $*" | tee -a "$OUTPUT_DIR/soak-test.log"
}

log_metric() {
    local metric_file="$1"
    shift
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "${timestamp},$*" >> "$OUTPUT_DIR/metrics/${metric_file}.csv"
}

# ============================================================================
# Monitoring Functions
# ============================================================================

check_consensus_health() {
    log "Checking consensus health..."

    local validators=("aura-validator-1" "aura-validator-2" "aura-validator-3" "aura-validator-4")
    local heights=()
    local consensus_ok=true

    for val in "${validators[@]}"; do
        if docker ps --format "{{.Names}}" | grep -q "^${val}$"; then
            local status=$(docker exec "$val" aurad status 2>&1 || echo "{}")
            local height=$(echo "$status" | jq -r '.sync_info.latest_block_height // "0"')
            local catching_up=$(echo "$status" | jq -r '.sync_info.catching_up // "true"')

            heights+=("$val:$height")

            if [[ "$catching_up" == "true" ]]; then
                log "  WARNING: $val is catching up (height: $height)"
                consensus_ok=false
            fi
        fi
    done

    # Check if heights are within reasonable range (5 blocks)
    local min_height=999999999
    local max_height=0

    for h in "${heights[@]}"; do
        local height=$(echo "$h" | cut -d: -f2)
        if [[ $height -gt 0 ]]; then
            if [[ $height -lt $min_height ]]; then
                min_height=$height
            fi
            if [[ $height -gt $max_height ]]; then
                max_height=$height
            fi
        fi
    done

    local height_diff=$((max_height - min_height))
    if [[ $height_diff -gt 5 ]]; then
        log "  WARNING: Height difference between validators: $height_diff blocks"
        consensus_ok=false
    fi

    if $consensus_ok; then
        log "  OK: Consensus healthy (height: $max_height, diff: $height_diff)"
    fi

    # Log metrics
    log_metric "consensus" "$max_height,$height_diff,$consensus_ok"

    return $([ "$consensus_ok" = true ] && echo 0 || echo 1)
}

check_memory_usage() {
    log "Checking memory usage..."

    local validators=("aura-validator-1" "aura-validator-2" "aura-validator-3" "aura-validator-4")
    local memory_ok=true

    for val in "${validators[@]}"; do
        if docker ps --format "{{.Names}}" | grep -q "^${val}$"; then
            local mem_usage=$(docker stats --no-stream --format "{{.MemUsage}}" "$val")
            local mem_perc=$(docker stats --no-stream --format "{{.MemPerc}}" "$val" | awk -F'%' '{print $1}')

            log "  $val: $mem_usage (${mem_perc}%)"
            log_metric "memory" "$val,$mem_usage,$mem_perc"

            # Alert if memory usage exceeds 80%
            if (( $(echo "$mem_perc > 80" | bc -l) )); then
                log "  WARNING: High memory usage on $val: ${mem_perc}%"
                memory_ok=false
            fi
        fi
    done

    return $([ "$memory_ok" = true ] && echo 0 || echo 1)
}

check_cpu_usage() {
    log "Checking CPU usage..."

    local validators=("aura-validator-1" "aura-validator-2" "aura-validator-3" "aura-validator-4")

    for val in "${validators[@]}"; do
        if docker ps --format "{{.Names}}" | grep -q "^${val}$"; then
            local cpu_perc=$(docker stats --no-stream --format "{{.CPUPerc}}" "$val" | awk -F'%' '{print $1}')

            log "  $val: ${cpu_perc}%"
            log_metric "cpu" "$val,$cpu_perc"
        fi
    done
}

check_disk_usage() {
    log "Checking disk usage..."

    local validators=("aura-validator-1" "aura-validator-2" "aura-validator-3" "aura-validator-4")

    for val in "${validators[@]}"; do
        if docker ps --format "{{.Names}}" | grep -q "^${val}$"; then
            local data_size=$(docker exec "$val" du -sh /root/.aura/data 2>/dev/null | awk '{print $1}')
            local db_size=$(docker exec "$val" du -sh /root/.aura/data/application.db 2>/dev/null | awk '{print $1}' || echo "N/A")

            log "  $val data: $data_size (DB: $db_size)"
            log_metric "disk" "$val,$data_size,$db_size"
        fi
    done
}

check_block_production() {
    log "Checking block production rate..."

    local current_height=$(curl -s localhost:27657/status 2>&1 | jq -r '.result.sync_info.latest_block_height // "0"')
    local elapsed_seconds=$(($(date +%s) - START_TIME))
    local blocks_produced=$((current_height - START_HEIGHT))

    if [[ $elapsed_seconds -gt 0 ]]; then
        local blocks_per_second=$(echo "scale=3; $blocks_produced / $elapsed_seconds" | bc)
        local expected_rate=0.200  # ~5 second blocks = 0.2 blocks/sec

        log "  Current height: $current_height"
        log "  Blocks produced: $blocks_produced"
        log "  Rate: $blocks_per_second blocks/sec (expected: $expected_rate)"

        log_metric "block_production" "$current_height,$blocks_produced,$blocks_per_second"

        # Check if production rate is significantly lower than expected
        if (( $(echo "$blocks_per_second < 0.15" | bc -l) )); then
            log "  WARNING: Block production slower than expected"
            return 1
        fi
    fi

    return 0
}

check_error_logs() {
    log "Scanning for errors in logs..."

    local validators=("aura-validator-1" "aura-validator-2" "aura-validator-3" "aura-validator-4")
    local errors_found=false

    for val in "${validators[@]}"; do
        if docker ps --format "{{.Names}}" | grep -q "^${val}$"; then
            # Check for critical errors in last 100 lines
            local panic_count=$(docker logs "$val" --tail 100 2>&1 | grep -ci "panic" || echo 0)
            local error_count=$(docker logs "$val" --tail 100 2>&1 | grep -ci "error" || echo 0)
            local fatal_count=$(docker logs "$val" --tail 100 2>&1 | grep -ci "fatal" || echo 0)

            if [[ $panic_count -gt 0 ]] || [[ $fatal_count -gt 0 ]]; then
                log "  WARNING: $val - Panics: $panic_count, Fatals: $fatal_count, Errors: $error_count"
                errors_found=true
            fi

            log_metric "errors" "$val,$panic_count,$error_count,$fatal_count"
        fi
    done

    return $([ "$errors_found" = true ] && echo 1 || echo 0)
}

take_snapshot() {
    log "Taking system snapshot..."

    local snapshot_time=$(date +%Y%m%d-%H%M%S)
    local snapshot_file="$OUTPUT_DIR/snapshots/snapshot-${snapshot_time}.txt"

    {
        echo "=== System Snapshot: $(date) ==="
        echo ""
        echo "=== Docker Containers ==="
        docker ps --filter "name=aura-" --format "table {{.Names}}\t{{.Status}}\t{{.Size}}"
        echo ""
        echo "=== Validator Status ==="
        curl -s localhost:27657/status 2>&1 | jq '.result.sync_info'
        echo ""
        echo "=== Resource Usage ==="
        docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" | grep aura-
        echo ""
    } > "$snapshot_file"

    log "  Snapshot saved to: $snapshot_file"
}

# ============================================================================
# Main Test Loop
# ============================================================================

run_soak_test() {
    local end_time=$((START_TIME + DURATION_HOURS * 3600))
    local next_snapshot=$((START_TIME + 3600))  # Snapshot every hour

    log "=== Starting Soak Test ==="
    log "End time: $(date -d @$end_time '+%Y-%m-%d %H:%M:%S')"

    # Initialize metric CSV files
    echo "timestamp,height,diff,healthy" > "$OUTPUT_DIR/metrics/consensus.csv"
    echo "timestamp,validator,memory_usage,memory_perc" > "$OUTPUT_DIR/metrics/memory.csv"
    echo "timestamp,validator,cpu_perc" > "$OUTPUT_DIR/metrics/cpu.csv"
    echo "timestamp,validator,data_size,db_size" > "$OUTPUT_DIR/metrics/disk.csv"
    echo "timestamp,height,blocks_produced,rate" > "$OUTPUT_DIR/metrics/block_production.csv"
    echo "timestamp,validator,panics,errors,fatals" > "$OUTPUT_DIR/metrics/errors.csv"

    while [[ $(date +%s) -lt $end_time ]]; do
        TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

        local current_time=$(date +%s)
        local elapsed_hours=$(echo "scale=2; ($current_time - $START_TIME) / 3600" | bc)
        local remaining_hours=$(echo "scale=2; ($end_time - $current_time) / 3600" | bc)

        echo ""
        log "=== Check #${TOTAL_CHECKS} (Elapsed: ${elapsed_hours}h, Remaining: ${remaining_hours}h) ==="

        # Run health checks
        local check_failed=false

        check_consensus_health || { check_failed=true; FAILED_CHECKS=$((FAILED_CHECKS + 1)); }
        check_memory_usage || check_failed=true
        check_cpu_usage
        check_disk_usage
        check_block_production || check_failed=true
        check_error_logs || check_failed=true

        # Take hourly snapshots
        if [[ $current_time -ge $next_snapshot ]]; then
            take_snapshot
            next_snapshot=$((next_snapshot + 3600))
        fi

        # Sleep until next check
        log "Next check in ${CHECK_INTERVAL} seconds..."
        sleep "$CHECK_INTERVAL"
    done

    log "=== Soak Test Complete ==="
}

# ============================================================================
# Results Summary
# ============================================================================

generate_summary() {
    log "=== Generating Summary Report ==="

    local summary_file="$OUTPUT_DIR/SUMMARY.txt"
    local end_height=$(curl -s localhost:27657/status 2>&1 | jq -r '.result.sync_info.latest_block_height // "0"')
    local total_blocks=$((end_height - START_HEIGHT))
    local elapsed_seconds=$(($(date +%s) - START_TIME))
    local avg_block_time=$(echo "scale=3; $elapsed_seconds / $total_blocks" | bc)
    local uptime_percent=$(echo "scale=2; (($TOTAL_CHECKS - $FAILED_CHECKS) * 100) / $TOTAL_CHECKS" | bc)

    {
        echo "=== AURA Blockchain Soak Test Summary ==="
        echo "=========================================="
        echo ""
        echo "Test Duration:      ${DURATION_HOURS} hours"
        echo "Start Time:         $(date -d @$START_TIME '+%Y-%m-%d %H:%M:%S')"
        echo "End Time:           $(date '+%Y-%m-%d %H:%M:%S')"
        echo ""
        echo "=== Block Production ==="
        echo "Start Height:       $START_HEIGHT"
        echo "End Height:         $end_height"
        echo "Total Blocks:       $total_blocks"
        echo "Average Block Time: ${avg_block_time}s"
        echo ""
        echo "=== Health Checks ==="
        echo "Total Checks:       $TOTAL_CHECKS"
        echo "Failed Checks:      $FAILED_CHECKS"
        echo "Uptime:             ${uptime_percent}%"
        echo ""
        echo "=== Results Location ==="
        echo "Summary:            $summary_file"
        echo "Detailed Log:       $OUTPUT_DIR/soak-test.log"
        echo "Metrics:            $OUTPUT_DIR/metrics/"
        echo "Snapshots:          $OUTPUT_DIR/snapshots/"
        echo ""
    } | tee "$summary_file"

    log "Summary saved to: $summary_file"
}

# ============================================================================
# Signal Handling
# ============================================================================

cleanup() {
    echo ""
    log "Received interrupt signal, stopping test..."
    generate_summary
    exit 0
}

trap cleanup SIGINT SIGTERM

# ============================================================================
# Main Entry Point
# ============================================================================

main() {
    setup
    run_soak_test
    generate_summary

    echo ""
    echo -e "${GREEN}=== Soak Test Completed Successfully ===${NC}"
    echo -e "Results directory: ${OUTPUT_DIR}"
    echo ""
}

main
