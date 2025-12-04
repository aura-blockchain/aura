#!/bin/bash
# ============================================================================
# AURA Testnet Monitoring & Health Check Script
# ============================================================================
# Comprehensive monitoring tool for the 4-validator local testnet
#
# Features:
#   - Quick health checks (single command)
#   - Continuous monitoring with real-time updates
#   - Individual node status checks
#   - Log analysis and error detection
#   - Performance metrics (block time, TPS, gas usage)
#   - Network health (peer connections, consensus)
#   - Troubleshooting diagnostics
# ============================================================================

set -euo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly CYAN='\033[0;36m'
readonly MAGENTA='\033[0;35m'
readonly NC='\033[0m'
readonly BOLD='\033[1m'

# Configuration
readonly VALIDATORS=("validator-1" "validator-2" "validator-3" "validator-4")
readonly RPC_PORTS=(27657 27757 27857 27957)
readonly API_PORTS=(2317 2417 2517 2617)
readonly GRPC_PORTS=(10090 10190 10290 10390)
readonly METRICS_PORTS=(27660 27760 27860 27960)
readonly CHAIN_ID="aura-local-4"
readonly COMPOSE_FILE="docker-compose.testnet.yml"

# State tracking for continuous monitoring
declare -A LAST_HEIGHT
declare -A LAST_TIMESTAMP
declare -A ERROR_COUNT

# ============================================================================
# Helper Functions
# ============================================================================

print_header() {
    echo -e "${BLUE}${BOLD}============================================================================${NC}"
    echo -e "${BLUE}${BOLD}$1${NC}"
    echo -e "${BLUE}${BOLD}============================================================================${NC}"
}

print_section() {
    echo -e "\n${CYAN}${BOLD}▶ $1${NC}"
    echo -e "${CYAN}──────────────────────────────────────────────────────────────${NC}"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "${CYAN}→${NC} $1"
}

print_metric() {
    local label=$1
    local value=$2
    local unit=${3:-}
    printf "  %-30s %s%s\n" "$label:" "$value" "$unit"
}

# ============================================================================
# Core Monitoring Functions
# ============================================================================

get_container_name() {
    local validator=$1
    echo "aura-${validator}"
}

is_container_running() {
    local validator=$1
    local container=$(get_container_name "$validator")
    docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${container}$"
}

get_rpc_port() {
    local index=$1
    echo "${RPC_PORTS[$index]}"
}

get_api_port() {
    local index=$1
    echo "${API_PORTS[$index]}"
}

get_metrics_port() {
    local index=$1
    echo "${METRICS_PORTS[$index]}"
}

# Query RPC endpoint with timeout and error handling
query_rpc() {
    local port=$1
    local endpoint=$2
    curl -sf --max-time 5 "http://localhost:${port}${endpoint}" 2>/dev/null
}

# Get block height for a validator
get_block_height() {
    local port=$1
    query_rpc "$port" "/status" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "0"
}

# Get block time for a validator
get_block_time() {
    local port=$1
    query_rpc "$port" "/status" | jq -r '.result.sync_info.latest_block_time' 2>/dev/null || echo "N/A"
}

# Get number of peers
get_peer_count() {
    local port=$1
    query_rpc "$port" "/net_info" | jq -r '.result.n_peers' 2>/dev/null || echo "0"
}

# Get validator voting power
get_voting_power() {
    local port=$1
    query_rpc "$port" "/status" | jq -r '.result.validator_info.voting_power' 2>/dev/null || echo "0"
}

# Check if validator is catching up
is_catching_up() {
    local port=$1
    local catching_up=$(query_rpc "$port" "/status" | jq -r '.result.sync_info.catching_up' 2>/dev/null)
    [[ "$catching_up" == "true" ]]
}

# Get container memory usage
get_memory_usage() {
    local container=$1
    docker stats --no-stream --format "{{.MemUsage}}" "$container" 2>/dev/null | awk '{print $1}'
}

# Get container CPU usage
get_cpu_usage() {
    local container=$1
    docker stats --no-stream --format "{{.CPUPerc}}" "$container" 2>/dev/null
}

# ============================================================================
# Quick Health Check (Single Command)
# ============================================================================

cmd_quick_check() {
    print_header "AURA Testnet - Quick Health Check"

    local running=0
    local healthy=0
    local producing=0
    local total=${#VALIDATORS[@]}

    for i in "${!VALIDATORS[@]}"; do
        local validator="${VALIDATORS[$i]}"
        local container=$(get_container_name "$validator")
        local rpc_port=$(get_rpc_port "$i")

        if is_container_running "$validator"; then
            running=$((running + 1))

            # Check if producing blocks
            local height=$(get_block_height "$rpc_port")
            if [[ "$height" -gt 0 ]]; then
                producing=$((producing + 1))
                healthy=$((healthy + 1))
                print_success "${validator} - Running (Height: ${height})"
            else
                print_warning "${validator} - Running but not producing blocks"
            fi
        else
            print_error "${validator} - Not running"
        fi
    done

    echo ""
    print_section "Summary"
    print_metric "Containers Running" "${running}/${total}"
    print_metric "Healthy Validators" "${healthy}/${total}"
    print_metric "Producing Blocks" "${producing}/${total}"

    # Overall health status
    echo ""
    if [[ $healthy -ge 3 ]]; then
        print_success "Testnet is HEALTHY - Consensus operating normally"
        return 0
    elif [[ $healthy -ge 1 ]]; then
        print_warning "Testnet is DEGRADED - Some validators offline"
        return 1
    else
        print_error "Testnet is DOWN - No validators producing blocks"
        return 2
    fi
}

# ============================================================================
# Continuous Monitoring (Real-time Updates)
# ============================================================================

cmd_watch() {
    print_header "AURA Testnet - Continuous Monitoring"
    print_info "Updating every 3 seconds. Press Ctrl+C to exit."
    echo ""

    # Initialize last heights
    for i in "${!VALIDATORS[@]}"; do
        LAST_HEIGHT["${VALIDATORS[$i]}"]=0
        ERROR_COUNT["${VALIDATORS[$i]}"]=0
    done

    local iteration=0
    while true; do
        iteration=$((iteration + 1))

        # Clear screen on non-first iterations
        if [[ $iteration -gt 1 ]]; then
            clear
            print_header "AURA Testnet - Live Monitor (Update #${iteration})"
            echo -e "${CYAN}Last updated: $(date '+%Y-%m-%d %H:%M:%S')${NC}"
        fi

        # Monitor each validator
        for i in "${!VALIDATORS[@]}"; do
            local validator="${VALIDATORS[$i]}"
            local container=$(get_container_name "$validator")
            local rpc_port=$(get_rpc_port "$i")

            echo ""
            print_section "${validator} (RPC: ${rpc_port})"

            if ! is_container_running "$validator"; then
                print_error "Container not running"
                ERROR_COUNT["$validator"]=$((ERROR_COUNT["$validator"] + 1))
                continue
            fi

            # Get current metrics
            local height=$(get_block_height "$rpc_port")
            local peers=$(get_peer_count "$rpc_port")
            local memory=$(get_memory_usage "$container")
            local cpu=$(get_cpu_usage "$container")
            local voting_power=$(get_voting_power "$rpc_port")

            # Calculate block progress
            local last_height=${LAST_HEIGHT["$validator"]:-0}
            local height_diff=$((height - last_height))
            LAST_HEIGHT["$validator"]=$height

            # Display metrics
            if [[ $height_diff -gt 0 ]]; then
                print_metric "Status" "${GREEN}PRODUCING BLOCKS${NC}"
                print_metric "Block Height" "$height (+${height_diff})"
            elif [[ $height -gt 0 ]]; then
                print_metric "Status" "${YELLOW}STALLED${NC}"
                print_metric "Block Height" "$height (no change)"
            else
                print_metric "Status" "${RED}NOT SYNCING${NC}"
                print_metric "Block Height" "N/A"
            fi

            print_metric "Peers" "$peers"
            print_metric "Voting Power" "$voting_power"
            print_metric "Memory Usage" "$memory"
            print_metric "CPU Usage" "$cpu"

            # Check for catching up
            if is_catching_up "$rpc_port"; then
                print_warning "Validator is catching up with the network"
            fi
        done

        # Network summary
        echo ""
        print_section "Network Summary"

        local heights=()
        for i in "${!VALIDATORS[@]}"; do
            local rpc_port=$(get_rpc_port "$i")
            local h=$(get_block_height "$rpc_port")
            heights+=("$h")
        done

        # Check if heights are consistent
        local max_height=$(printf '%s\n' "${heights[@]}" | sort -nr | head -1)
        local min_height=$(printf '%s\n' "${heights[@]}" | grep -v "^0$" | sort -n | head -1)
        local height_diff=$((max_height - min_height))

        print_metric "Max Height" "$max_height"
        print_metric "Min Height" "$min_height"
        print_metric "Height Variance" "$height_diff blocks"

        if [[ $height_diff -le 1 ]]; then
            print_success "All validators in sync"
        else
            print_warning "Validators have different heights (may indicate sync issues)"
        fi

        sleep 3
    done
}

# ============================================================================
# Individual Validator Status
# ============================================================================

cmd_validator_status() {
    local validator=${1:-validator-1}

    # Find validator index
    local index=-1
    for i in "${!VALIDATORS[@]}"; do
        if [[ "${VALIDATORS[$i]}" == "$validator" ]]; then
            index=$i
            break
        fi
    done

    if [[ $index -eq -1 ]]; then
        print_error "Invalid validator: $validator"
        print_info "Valid validators: ${VALIDATORS[*]}"
        return 1
    fi

    print_header "Detailed Status - ${validator}"

    local container=$(get_container_name "$validator")
    local rpc_port=$(get_rpc_port "$index")
    local api_port=$(get_api_port "$index")
    local metrics_port=$(get_metrics_port "$index")

    # Container status
    print_section "Container Status"
    if is_container_running "$validator"; then
        print_success "Container is running"

        local health=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null || echo "unknown")
        print_metric "Health Status" "$health"

        local uptime=$(docker inspect --format='{{.State.StartedAt}}' "$container" 2>/dev/null)
        print_metric "Started At" "$uptime"
    else
        print_error "Container is not running"
        return 1
    fi

    # Chain status
    print_section "Chain Status"
    local status=$(query_rpc "$rpc_port" "/status")

    if [[ -n "$status" ]]; then
        local node_id=$(echo "$status" | jq -r '.result.node_info.id')
        local version=$(echo "$status" | jq -r '.result.node_info.version')
        local network=$(echo "$status" | jq -r '.result.node_info.network')
        local moniker=$(echo "$status" | jq -r '.result.node_info.moniker')

        print_metric "Node ID" "$node_id"
        print_metric "Moniker" "$moniker"
        print_metric "Chain ID" "$network"
        print_metric "Version" "$version"

        local height=$(echo "$status" | jq -r '.result.sync_info.latest_block_height')
        local block_time=$(echo "$status" | jq -r '.result.sync_info.latest_block_time')
        local catching_up=$(echo "$status" | jq -r '.result.sync_info.catching_up')

        print_metric "Block Height" "$height"
        print_metric "Last Block Time" "$block_time"
        print_metric "Catching Up" "$catching_up"

        local address=$(echo "$status" | jq -r '.result.validator_info.address')
        local voting_power=$(echo "$status" | jq -r '.result.validator_info.voting_power')

        print_metric "Validator Address" "$address"
        print_metric "Voting Power" "$voting_power"
    else
        print_error "Unable to query RPC endpoint"
    fi

    # Network info
    print_section "Network Info"
    local net_info=$(query_rpc "$rpc_port" "/net_info")

    if [[ -n "$net_info" ]]; then
        local listening=$(echo "$net_info" | jq -r '.result.listening')
        local n_peers=$(echo "$net_info" | jq -r '.result.n_peers')

        print_metric "Listening" "$listening"
        print_metric "Connected Peers" "$n_peers"

        echo ""
        print_info "Peer List:"
        echo "$net_info" | jq -r '.result.peers[] | "  - \(.node_info.moniker) (\(.node_info.id))"'
    else
        print_error "Unable to query network info"
    fi

    # Resource usage
    print_section "Resource Usage"
    local memory=$(get_memory_usage "$container")
    local cpu=$(get_cpu_usage "$container")

    print_metric "Memory" "$memory"
    print_metric "CPU" "$cpu"

    # Endpoints
    print_section "Access Endpoints"
    print_metric "RPC" "http://localhost:${rpc_port}"
    print_metric "API" "http://localhost:${api_port}"
    print_metric "gRPC" "localhost:${GRPC_PORTS[$index]}"
    print_metric "Metrics" "http://localhost:${metrics_port}"
}

# ============================================================================
# Block Production Monitor
# ============================================================================

cmd_watch_blocks() {
    local validator=${1:-validator-1}

    # Find validator index
    local index=-1
    for i in "${!VALIDATORS[@]}"; do
        if [[ "${VALIDATORS[$i]}" == "$validator" ]]; then
            index=$i
            break
        fi
    done

    if [[ $index -eq -1 ]]; then
        print_error "Invalid validator: $validator"
        return 1
    fi

    local rpc_port=$(get_rpc_port "$index")

    print_header "Block Production Monitor - ${validator}"
    print_info "Watching RPC: http://localhost:${rpc_port}"
    print_info "Press Ctrl+C to exit"
    echo ""

    local last_height=0
    local start_time=$(date +%s)
    local start_height=$(get_block_height "$rpc_port")

    while true; do
        local height=$(get_block_height "$rpc_port")
        local current_time=$(date +%s)

        if [[ $height -gt $last_height ]]; then
            local block_time=$(get_block_time "$rpc_port")
            local elapsed=$((current_time - start_time))
            local blocks_produced=$((height - start_height))
            local avg_block_time=0

            if [[ $blocks_produced -gt 0 ]]; then
                avg_block_time=$(awk "BEGIN {printf \"%.2f\", $elapsed / $blocks_produced}")
            fi

            echo -e "${GREEN}▲${NC} Block ${BOLD}${height}${NC} at ${block_time} (avg: ${avg_block_time}s/block)"
            last_height=$height
        fi

        sleep 1
    done
}

# ============================================================================
# Log Analysis and Error Detection
# ============================================================================

cmd_check_logs() {
    local validator=${1:-all}
    local lines=${2:-100}

    print_header "Log Analysis - Error Detection"

    local validators_to_check=()
    if [[ "$validator" == "all" ]]; then
        validators_to_check=("${VALIDATORS[@]}")
    else
        validators_to_check=("$validator")
    fi

    for val in "${validators_to_check[@]}"; do
        print_section "Checking ${val}"

        local container=$(get_container_name "$val")

        if ! is_container_running "$val"; then
            print_error "Container not running"
            continue
        fi

        # Get logs
        local logs=$(docker logs --tail "$lines" "$container" 2>&1)

        # Error patterns to detect
        local error_patterns=(
            "ERROR"
            "FATAL"
            "panic"
            "failed"
            "error"
            "timeout"
            "connection refused"
            "consensus failure"
            "validator.*double.*sign"
            "insufficient.*voting.*power"
        )

        local total_errors=0

        for pattern in "${error_patterns[@]}"; do
            local count=$(echo "$logs" | grep -ci "$pattern" || true)
            if [[ $count -gt 0 ]]; then
                total_errors=$((total_errors + count))
                print_warning "Found ${count} occurrences of '${pattern}'"
            fi
        done

        if [[ $total_errors -eq 0 ]]; then
            print_success "No errors detected in last ${lines} log lines"
        else
            print_error "Total error occurrences: ${total_errors}"
            echo ""
            print_info "Recent errors:"
            echo "$logs" | grep -Ei "error|fatal|panic|failed" | tail -5 | sed 's/^/  /'
        fi

        echo ""
    done
}

cmd_tail_logs() {
    local validator=${1:-validator-1}
    local container=$(get_container_name "$validator")

    print_header "Tailing Logs - ${validator}"
    print_info "Press Ctrl+C to exit"
    echo ""

    docker logs -f --tail 50 "$container"
}

# ============================================================================
# Performance Metrics
# ============================================================================

cmd_performance() {
    print_header "Performance Metrics"

    print_section "Block Production Statistics"

    # Sample block production over 30 seconds
    print_info "Sampling block production over 30 seconds..."

    local samples=10
    local interval=3
    local heights=()

    for i in $(seq 1 $samples); do
        local height=$(get_block_height "${RPC_PORTS[0]}")
        heights+=("$height")
        if [[ $i -lt $samples ]]; then
            sleep $interval
        fi
    done

    # Calculate metrics
    local start_height=${heights[0]}
    local end_height=${heights[-1]}
    local blocks_produced=$((end_height - start_height))
    local total_time=$((samples * interval))
    local avg_block_time=$(awk "BEGIN {printf \"%.2f\", $total_time / $blocks_produced}")
    local blocks_per_min=$(awk "BEGIN {printf \"%.1f\", $blocks_produced * 60 / $total_time}")

    print_metric "Start Height" "$start_height"
    print_metric "End Height" "$end_height"
    print_metric "Blocks Produced" "$blocks_produced"
    print_metric "Time Period" "${total_time}s"
    print_metric "Average Block Time" "${avg_block_time}s"
    print_metric "Blocks per Minute" "$blocks_per_min"

    # Transaction throughput (would need to query actual tx count)
    print_section "Resource Usage Across Validators"

    for i in "${!VALIDATORS[@]}"; do
        local validator="${VALIDATORS[$i]}"
        local container=$(get_container_name "$validator")

        if is_container_running "$validator"; then
            local memory=$(get_memory_usage "$container")
            local cpu=$(get_cpu_usage "$container")

            printf "  %-15s Memory: %-10s CPU: %s\n" "$validator" "$memory" "$cpu"
        fi
    done
}

# ============================================================================
# Network Health
# ============================================================================

cmd_network_health() {
    print_header "Network Health Check"

    print_section "Peer Connectivity"

    for i in "${!VALIDATORS[@]}"; do
        local validator="${VALIDATORS[$i]}"
        local rpc_port=$(get_rpc_port "$i")

        if ! is_container_running "$validator"; then
            print_error "${validator}: Not running"
            continue
        fi

        local peers=$(get_peer_count "$rpc_port")
        local expected_peers=$((${#VALIDATORS[@]} - 1))

        if [[ $peers -eq $expected_peers ]]; then
            print_success "${validator}: ${peers}/${expected_peers} peers connected"
        elif [[ $peers -gt 0 ]]; then
            print_warning "${validator}: ${peers}/${expected_peers} peers connected"
        else
            print_error "${validator}: No peers connected"
        fi
    done

    print_section "Consensus Status"

    # Check if all validators are at similar heights
    local heights=()
    local max_height=0

    for i in "${!VALIDATORS[@]}"; do
        local rpc_port=$(get_rpc_port "$i")
        local height=$(get_block_height "$rpc_port")
        heights+=("$height")

        if [[ $height -gt $max_height ]]; then
            max_height=$height
        fi
    done

    for i in "${!VALIDATORS[@]}"; do
        local validator="${VALIDATORS[$i]}"
        local height=${heights[$i]}
        local diff=$((max_height - height))

        if [[ $diff -eq 0 ]]; then
            print_success "${validator}: Height ${height} (in sync)"
        elif [[ $diff -le 5 ]]; then
            print_warning "${validator}: Height ${height} (${diff} blocks behind)"
        else
            print_error "${validator}: Height ${height} (${diff} blocks behind)"
        fi
    done

    print_section "RPC Endpoint Availability"

    for i in "${!VALIDATORS[@]}"; do
        local validator="${VALIDATORS[$i]}"
        local rpc_port=$(get_rpc_port "$i")
        local api_port=$(get_api_port "$i")

        # Test RPC
        if query_rpc "$rpc_port" "/health" >/dev/null 2>&1; then
            print_success "${validator} RPC (${rpc_port}): Available"
        else
            print_error "${validator} RPC (${rpc_port}): Unavailable"
        fi

        # Test API
        if curl -sf --max-time 5 "http://localhost:${api_port}/cosmos/base/tendermint/v1beta1/node_info" >/dev/null 2>&1; then
            print_success "${validator} API (${api_port}): Available"
        else
            print_error "${validator} API (${api_port}): Unavailable"
        fi
    done
}

# ============================================================================
# Troubleshooting Commands
# ============================================================================

cmd_diagnose() {
    local validator=${1:-all}

    print_header "Diagnostic Report"

    print_section "Docker Environment"
    print_metric "Docker Version" "$(docker --version | awk '{print $3}' | tr -d ',')"
    print_metric "Docker Compose" "$(docker-compose --version | awk '{print $4}' | tr -d ',')"

    print_section "Container Status"

    local validators_to_check=()
    if [[ "$validator" == "all" ]]; then
        validators_to_check=("${VALIDATORS[@]}")
    else
        validators_to_check=("$validator")
    fi

    for val in "${validators_to_check[@]}"; do
        local container=$(get_container_name "$val")

        echo ""
        print_info "Diagnosing ${val}"

        if ! docker ps -a --format '{{.Names}}' | grep -q "^${container}$"; then
            print_error "Container does not exist"
            print_info "Run: ./scripts/testnet-init.sh && docker-compose -f docker-compose.testnet.yml up -d"
            continue
        fi

        local status=$(docker inspect --format='{{.State.Status}}' "$container" 2>/dev/null)
        print_metric "Status" "$status"

        if [[ "$status" != "running" ]]; then
            print_error "Container is not running"

            # Check exit code
            local exit_code=$(docker inspect --format='{{.State.ExitCode}}' "$container" 2>/dev/null)
            print_metric "Exit Code" "$exit_code"

            # Show recent logs
            print_info "Last 20 log lines:"
            docker logs --tail 20 "$container" 2>&1 | sed 's/^/  /'
            continue
        fi

        local health=$(docker inspect --format='{{.State.Health.Status}}' "$container" 2>/dev/null || echo "none")
        print_metric "Health" "$health"

        # Check if ports are accessible
        local rpc_port=$(get_rpc_port $(printf "%s\n" "${VALIDATORS[@]}" | grep -n "^${val}$" | cut -d: -f1 | head -1 | awk '{print $1-1}'))

        if nc -z localhost "$rpc_port" 2>/dev/null; then
            print_success "RPC port ${rpc_port} is accessible"
        else
            print_error "RPC port ${rpc_port} is not accessible"
        fi

        # Check recent errors
        local error_count=$(docker logs --tail 100 "$container" 2>&1 | grep -ci "error" || true)
        print_metric "Recent Errors" "$error_count (in last 100 lines)"
    done

    print_section "Recommendations"

    local running_count=0
    for val in "${VALIDATORS[@]}"; do
        if is_container_running "$val"; then
            running_count=$((running_count + 1))
        fi
    done

    if [[ $running_count -eq 0 ]]; then
        print_info "No validators running. Start with: docker-compose -f docker-compose.testnet.yml up -d"
    elif [[ $running_count -lt 3 ]]; then
        print_warning "Only ${running_count}/4 validators running. Consensus requires 3/4."
        print_info "Start missing validators with: docker-compose -f docker-compose.testnet.yml up -d"
    else
        print_success "Sufficient validators running (${running_count}/4)"
    fi
}

cmd_restart_failed() {
    print_header "Restarting Failed Validators"

    local restarted=0

    for validator in "${VALIDATORS[@]}"; do
        if ! is_container_running "$validator"; then
            print_info "Starting ${validator}..."
            docker-compose -f "$COMPOSE_FILE" up -d "$validator"
            restarted=$((restarted + 1))
        fi
    done

    if [[ $restarted -eq 0 ]]; then
        print_success "All validators are already running"
    else
        print_success "Restarted ${restarted} validator(s)"
        print_info "Waiting 10 seconds for startup..."
        sleep 10
        cmd_quick_check
    fi
}

# ============================================================================
# Clean Shutdown Procedure
# ============================================================================

cmd_shutdown() {
    print_header "Clean Testnet Shutdown Procedure"

    print_info "Step 1: Stopping validators gracefully..."
    for validator in "${VALIDATORS[@]}"; do
        if is_container_running "$validator"; then
            print_info "Stopping ${validator}..."
            docker-compose -f "$COMPOSE_FILE" stop "$validator"
            print_success "${validator} stopped"
        fi
    done

    print_info "Step 2: Waiting for containers to stop..."
    sleep 5

    print_info "Step 3: Stopping monitoring services..."
    docker-compose -f "$COMPOSE_FILE" stop prometheus-testnet grafana-testnet 2>/dev/null || true

    print_info "Step 4: Removing containers (keeping volumes)..."
    docker-compose -f "$COMPOSE_FILE" down

    print_success "Testnet shut down cleanly"
    print_info "To restart: docker-compose -f docker-compose.testnet.yml up -d"
    print_info "To wipe data: ./scripts/testnet-manage.sh clean"
}

# ============================================================================
# Help and Usage
# ============================================================================

cmd_help() {
    cat << 'EOF'
AURA Testnet Monitoring & Health Check Script

USAGE:
    ./scripts/testnet-monitor.sh <command> [options]

QUICK HEALTH CHECKS:
    quick               Quick health check of all validators (single command)
    status <validator>  Detailed status of a specific validator

CONTINUOUS MONITORING:
    watch               Real-time monitoring dashboard (updates every 3s)
    watch-blocks [val]  Watch block production in real-time

LOG ANALYSIS:
    logs <validator>    Tail logs for a validator
    check-logs [val]    Scan logs for errors (val=all for all validators)

PERFORMANCE:
    performance         Measure block time, TPS, resource usage
    network             Check network health and peer connectivity

TROUBLESHOOTING:
    diagnose [val]      Run diagnostic checks (val=all for all validators)
    restart-failed      Restart any stopped validators
    shutdown            Clean shutdown of entire testnet

COMMON ERROR PATTERNS:
    - "consensus failure"       → Validator not participating in consensus
    - "connection refused"      → Peer connectivity issues
    - "insufficient voting"     → Not enough validators online (need 3/4)
    - "double sign"            → Validator signing multiple blocks
    - "timeout"                → Network latency or slow queries

EXAMPLES:
    # Quick health check
    ./scripts/testnet-monitor.sh quick

    # Watch blocks being produced
    ./scripts/testnet-monitor.sh watch-blocks validator-1

    # Check logs for errors
    ./scripts/testnet-monitor.sh check-logs all

    # Full diagnostic on validator-2
    ./scripts/testnet-monitor.sh status validator-2

    # Performance metrics
    ./scripts/testnet-monitor.sh performance

TROUBLESHOOTING STEPS:

    If nodes fail to start:
        1. ./scripts/testnet-monitor.sh diagnose all
        2. Check logs: ./scripts/testnet-monitor.sh logs validator-1
        3. Restart: ./scripts/testnet-monitor.sh restart-failed

    If consensus stalls:
        1. ./scripts/testnet-monitor.sh quick
        2. Ensure 3/4 validators running
        3. Check network: ./scripts/testnet-monitor.sh network

    If blocks stop producing:
        1. ./scripts/testnet-monitor.sh watch-blocks validator-1
        2. Check peer connectivity
        3. Verify voting power distribution

    Clean restart:
        1. ./scripts/testnet-monitor.sh shutdown
        2. docker-compose -f docker-compose.testnet.yml up -d
        3. ./scripts/testnet-monitor.sh quick

MONITORING ENDPOINTS:
    - Prometheus: http://localhost:9094
    - Grafana: http://localhost:3002 (admin/aura-testnet-admin)
    - Validators: RPC ports 27657, 27757, 27857, 27957

For more information, see: /docs/runbooks/LOCAL_TESTNET_DOCKER.md
EOF
}

# ============================================================================
# Main Command Dispatcher
# ============================================================================

main() {
    local command=${1:-help}

    case "$command" in
        quick|health)
            cmd_quick_check
            ;;
        watch|monitor)
            cmd_watch
            ;;
        watch-blocks|blocks)
            cmd_watch_blocks "${2:-validator-1}"
            ;;
        status|info)
            cmd_validator_status "${2:-validator-1}"
            ;;
        logs|tail)
            cmd_tail_logs "${2:-validator-1}"
            ;;
        check-logs|errors)
            cmd_check_logs "${2:-all}" "${3:-100}"
            ;;
        performance|perf|metrics)
            cmd_performance
            ;;
        network|peers)
            cmd_network_health
            ;;
        diagnose|debug)
            cmd_diagnose "${2:-all}"
            ;;
        restart-failed|restart)
            cmd_restart_failed
            ;;
        shutdown|stop)
            cmd_shutdown
            ;;
        help|--help|-h)
            cmd_help
            ;;
        *)
            print_error "Unknown command: $command"
            echo ""
            cmd_help
            exit 1
            ;;
    esac
}

# Run main if not sourced
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
