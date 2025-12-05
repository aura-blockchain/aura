#!/bin/bash
# ============================================================================
# AURA Testnet Management Script
# ============================================================================
# Provides convenient commands for managing the multi-node testnet
#
# Features:
# - Store initialization verification
# - AppHash consistency checks
# - Start → Stop → Start cycle testing
# ============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Source verification library
source "${SCRIPT_DIR}/lib-store-verification.sh"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

COMPOSE_FILE="docker-compose.testnet.yml"
VALIDATORS=("validator-1" "validator-2" "validator-3" "validator-4")

# ============================================================================
# Helper Functions
# ============================================================================

print_header() {
    echo -e "${BLUE}============================================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}============================================================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# ============================================================================
# Command Functions
# ============================================================================

cmd_start() {
    print_header "Starting AURA Testnet"
    docker-compose -f "${COMPOSE_FILE}" up -d
    print_success "Testnet started"
    echo ""
    cmd_status
}

cmd_stop() {
    print_header "Stopping AURA Testnet"
    docker-compose -f "${COMPOSE_FILE}" down
    print_success "Testnet stopped"
}

cmd_restart() {
    print_header "Restarting AURA Testnet"
    docker-compose -f "${COMPOSE_FILE}" restart
    print_success "Testnet restarted"
}

cmd_status() {
    print_header "Testnet Status"

    # Check if containers are running
    for validator in "${VALIDATORS[@]}"; do
        CONTAINER_NAME="aura-${validator}"
        if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
            # Get block height
            PORT=$((26657 + (${validator: -1} - 1) * 100))
            if [ "${validator}" == "validator-1" ]; then
                PORT=26657
            fi

            HEIGHT=$(curl -s "http://localhost:${PORT}/status" 2>/dev/null | \
                     jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "N/A")

            echo -e "${GREEN}✓ ${validator}${NC} - Running (Height: ${HEIGHT})"
        else
            echo -e "${RED}✗ ${validator}${NC} - Stopped"
        fi
    done

    echo ""
    print_info "Monitoring:"
    echo "  Prometheus: http://localhost:9091"
    echo "  Grafana: http://localhost:3001 (admin/aura-testnet-admin)"
}

cmd_logs() {
    VALIDATOR=${1:-validator-1}
    print_header "Logs for ${VALIDATOR}"
    docker-compose -f "${COMPOSE_FILE}" logs -f "${VALIDATOR}"
}

cmd_exec() {
    VALIDATOR=${1:-validator-1}
    shift
    COMMAND=${@:-sh}

    print_header "Executing on ${VALIDATOR}"
    docker-compose -f "${COMPOSE_FILE}" exec "${VALIDATOR}" ${COMMAND}
}

cmd_query() {
    VALIDATOR=${1:-validator-1}
    PORT=$((26657 + (${VALIDATOR: -1} - 1) * 100))
    if [ "${VALIDATOR}" == "validator-1" ]; then
        PORT=26657
    fi

    print_header "Query RPC Status - ${VALIDATOR}"
    curl -s "http://localhost:${PORT}/status" | jq '.'
}

cmd_health() {
    print_header "Health Check for All Validators"

    for validator in "${VALIDATORS[@]}"; do
        CONTAINER_NAME="aura-${validator}"
        HEALTH=$(docker inspect --format='{{.State.Health.Status}}' "${CONTAINER_NAME}" 2>/dev/null || echo "not_running")

        if [ "${HEALTH}" == "healthy" ]; then
            echo -e "${GREEN}✓ ${validator}${NC} - Healthy"
        elif [ "${HEALTH}" == "unhealthy" ]; then
            echo -e "${RED}✗ ${validator}${NC} - Unhealthy"
        elif [ "${HEALTH}" == "starting" ]; then
            echo -e "${YELLOW}⚠ ${validator}${NC} - Starting..."
        else
            echo -e "${RED}✗ ${validator}${NC} - Not Running"
        fi
    done
}

cmd_bft_test() {
    print_header "Byzantine Fault Tolerance Test"
    print_info "Stopping validator-4 to test consensus with 3/4 validators..."

    docker-compose -f "${COMPOSE_FILE}" stop validator-4
    print_success "validator-4 stopped"

    echo ""
    print_info "Waiting 10 seconds for network to adjust..."
    sleep 10

    echo ""
    print_info "Checking if chain is still producing blocks..."

    HEIGHT_BEFORE=$(curl -s "http://localhost:26657/status" 2>/dev/null | \
                    jq -r '.result.sync_info.latest_block_height' 2>/dev/null)

    sleep 5

    HEIGHT_AFTER=$(curl -s "http://localhost:26657/status" 2>/dev/null | \
                   jq -r '.result.sync_info.latest_block_height' 2>/dev/null)

    if [ "${HEIGHT_AFTER}" -gt "${HEIGHT_BEFORE}" ]; then
        print_success "Chain is still producing blocks! (${HEIGHT_BEFORE} → ${HEIGHT_AFTER})"
        print_success "BFT test PASSED: 3/4 validators maintaining consensus"
    else
        print_error "Chain stopped producing blocks"
        print_error "BFT test FAILED"
    fi

    echo ""
    print_info "Restarting validator-4..."
    docker-compose -f "${COMPOSE_FILE}" start validator-4
    print_success "validator-4 restarted"
}

cmd_clean() {
    print_header "Cleaning Testnet Data"
    print_info "This will remove all containers, volumes, and testnet data"
    read -p "Are you sure? (yes/no): " -r
    echo

    if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        docker-compose -f "${COMPOSE_FILE}" down -v
        rm -rf testnet-data
        print_success "Testnet data cleaned"
        print_info "Run './scripts/testnet-init.sh' to reinitialize"
    else
        print_info "Clean operation cancelled"
    fi
}

cmd_ports() {
    print_header "Port Mappings"
    echo ""
    echo "Validator 1:"
    echo "  RPC:     http://localhost:26657"
    echo "  API:     http://localhost:1317"
    echo "  gRPC:    localhost:9090"
    echo "  P2P:     localhost:26656"
    echo "  Metrics: http://localhost:26660"
    echo ""
    echo "Validator 2:"
    echo "  RPC:     http://localhost:26757"
    echo "  API:     http://localhost:1417"
    echo "  gRPC:    localhost:9190"
    echo "  P2P:     localhost:26756"
    echo "  Metrics: http://localhost:26760"
    echo ""
    echo "Validator 3:"
    echo "  RPC:     http://localhost:26857"
    echo "  API:     http://localhost:1517"
    echo "  gRPC:    localhost:9290"
    echo "  P2P:     localhost:26856"
    echo "  Metrics: http://localhost:26860"
    echo ""
    echo "Validator 4:"
    echo "  RPC:     http://localhost:26957"
    echo "  API:     http://localhost:1617"
    echo "  gRPC:    localhost:9390"
    echo "  P2P:     localhost:26956"
    echo "  Metrics: http://localhost:26960"
    echo ""
    echo "Monitoring:"
    echo "  Prometheus: http://localhost:9091"
    echo "  Grafana:    http://localhost:3001"
    echo ""
}

cmd_verify_stores() {
    print_header "Store Initialization Verification"

    local validator=${1:-validator-1}
    local testnet_dir="testnet-data"

    if [ ! -d "$testnet_dir" ]; then
        print_error "Testnet data directory not found: $testnet_dir"
        print_info "Run ./scripts/testnet-init.sh first"
        return 1
    fi

    local node_home="${testnet_dir}/${validator}"

    if [ ! -d "$node_home" ]; then
        print_error "Validator home not found: $node_home"
        return 1
    fi

    print_info "Verifying stores for: $validator"
    echo ""

    verify_store_initialization "$node_home"
}

cmd_check_apphash() {
    print_header "AppHash Consistency Check"

    local validator=${1:-validator-1}
    local port=26657

    # Determine RPC port based on validator
    case $validator in
        validator-1) port=26657 ;;
        validator-2) port=26757 ;;
        validator-3) port=26857 ;;
        validator-4) port=26957 ;;
        *)
            print_error "Unknown validator: $validator"
            return 1
            ;;
    esac

    local rpc="http://localhost:${port}"
    local testnet_dir="testnet-data"
    local node_home="${testnet_dir}/${validator}"

    print_info "Checking AppHash for: $validator"
    print_info "RPC endpoint: $rpc"
    echo ""

    # Check if baseline exists
    local baseline_file="${node_home}/apphash_baseline.txt"
    if [ -f "$baseline_file" ]; then
        local baseline=$(cat "$baseline_file")
        verify_apphash_consistency "$node_home" "$rpc" "$baseline"
    else
        # Record baseline
        verify_apphash_consistency "$node_home" "$rpc"
    fi
}

cmd_test_restart_consistency() {
    print_header "Start → Stop → Start Consistency Test"

    print_info "This test verifies AppHash consistency across node restarts"
    print_info "It will:"
    print_info "  1. Record baseline AppHash"
    print_info "  2. Stop all validators"
    print_info "  3. Restart all validators"
    print_info "  4. Verify AppHash matches baseline"
    echo ""

    read -p "Continue? (yes/no): " -r
    echo

    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        print_info "Test cancelled"
        return 0
    fi

    # Step 1: Record baseline AppHash for all validators
    print_info "Step 1: Recording baseline AppHash..."
    declare -A baseline_hashes

    for validator in "${VALIDATORS[@]}"; do
        local port=26657
        case $validator in
            validator-1) port=26657 ;;
            validator-2) port=26757 ;;
            validator-3) port=26857 ;;
            validator-4) port=26957 ;;
        esac

        local rpc="http://localhost:${port}"
        local apphash=$(get_apphash "$rpc")

        if [ $? -eq 0 ]; then
            baseline_hashes[$validator]=$apphash
            echo -e "  ${validator}: ${apphash}"

            # Save to file
            local node_home="testnet-data/${validator}"
            echo "$apphash" > "${node_home}/apphash_baseline.txt"
        else
            print_error "Failed to get AppHash for ${validator}"
            return 1
        fi
    done

    echo ""

    # Step 2: Stop all validators
    print_info "Step 2: Stopping all validators..."
    docker-compose -f "${COMPOSE_FILE}" stop
    sleep 3
    print_success "All validators stopped"
    echo ""

    # Step 3: Restart all validators
    print_info "Step 3: Restarting all validators..."
    docker-compose -f "${COMPOSE_FILE}" start
    print_info "Waiting 15 seconds for validators to sync..."
    sleep 15
    print_success "All validators restarted"
    echo ""

    # Step 4: Verify AppHash consistency
    print_info "Step 4: Verifying AppHash consistency..."
    local all_consistent=true

    for validator in "${VALIDATORS[@]}"; do
        local port=26657
        case $validator in
            validator-1) port=26657 ;;
            validator-2) port=26757 ;;
            validator-3) port=26857 ;;
            validator-4) port=26957 ;;
        esac

        local rpc="http://localhost:${port}"
        local baseline="${baseline_hashes[$validator]}"
        local current=$(get_apphash "$rpc")

        if [ $? -ne 0 ]; then
            print_error "${validator}: Failed to get current AppHash"
            all_consistent=false
            continue
        fi

        if [ "$baseline" = "$current" ]; then
            print_success "${validator}: AppHash consistent (${current})"
        else
            print_error "${validator}: AppHash MISMATCH"
            echo -e "  Expected: ${baseline}"
            echo -e "  Current:  ${current}"
            all_consistent=false
        fi
    done

    echo ""
    if [ "$all_consistent" = true ]; then
        print_success "✓ Restart consistency test PASSED"
        print_success "All validators have consistent AppHash after restart"
    else
        print_error "✗ Restart consistency test FAILED"
        print_error "One or more validators have inconsistent AppHash"
        return 1
    fi
}

cmd_help() {
    cat << EOF
AURA Testnet Management Script

Usage: $0 <command> [options]

Commands:
    start                       Start all validators and monitoring
    stop                        Stop all validators
    restart                     Restart all validators
    status                      Show status of all validators
    logs [validator]            Show logs (default: validator-1)
    exec [validator]            Execute command in validator container
    query [validator]           Query RPC status (default: validator-1)
    health                      Check health status of all validators
    bft-test                    Test Byzantine Fault Tolerance (stops 1 validator)
    clean                       Remove all containers, volumes, and data
    ports                       Show port mappings for all services
    verify-stores [validator]   Verify store initialization (default: validator-1)
    check-apphash [validator]   Check AppHash consistency (default: validator-1)
    test-restart                Test AppHash consistency across restart
    help                        Show this help message

Examples:
    $0 start
    $0 logs validator-2
    $0 exec validator-1 aurad status
    $0 query validator-3
    $0 bft-test
    $0 verify-stores validator-1
    $0 check-apphash validator-2
    $0 test-restart

Advanced Monitoring:
    For comprehensive monitoring and health checks, use:
        ./scripts/testnet-monitor.sh <command>

    Available monitoring commands:
        quick               Quick health check (single command)
        watch               Live monitoring dashboard
        watch-blocks        Real-time block production
        performance         Performance metrics
        network             Network health check
        diagnose            Full diagnostic report
        check-logs          Scan logs for errors

    See: TESTNET_MONITORING_GUIDE.md or MONITORING_CHEATSHEET.md

EOF
}

# ============================================================================
# Main
# ============================================================================

COMMAND=${1:-help}

case "${COMMAND}" in
    start)
        cmd_start
        ;;
    stop)
        cmd_stop
        ;;
    restart)
        cmd_restart
        ;;
    status)
        cmd_status
        ;;
    logs)
        cmd_logs "$2"
        ;;
    exec)
        shift
        cmd_exec "$@"
        ;;
    query)
        cmd_query "$2"
        ;;
    health)
        cmd_health
        ;;
    bft-test)
        cmd_bft_test
        ;;
    clean)
        cmd_clean
        ;;
    ports)
        cmd_ports
        ;;
    verify-stores)
        cmd_verify_stores "$2"
        ;;
    check-apphash)
        cmd_check_apphash "$2"
        ;;
    test-restart)
        cmd_test_restart_consistency
        ;;
    help|--help|-h)
        cmd_help
        ;;
    *)
        print_error "Unknown command: ${COMMAND}"
        echo ""
        cmd_help
        exit 1
        ;;
esac
