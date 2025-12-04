#!/bin/bash
# ============================================================================
# AURA Testnet Management Script
# ============================================================================
# Provides convenient commands for managing the multi-node testnet
# ============================================================================

set -e

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

cmd_help() {
    cat << EOF
AURA Testnet Management Script

Usage: $0 <command> [options]

Commands:
    start               Start all validators and monitoring
    stop                Stop all validators
    restart             Restart all validators
    status              Show status of all validators
    logs [validator]    Show logs (default: validator-1)
    exec [validator]    Execute command in validator container
    query [validator]   Query RPC status (default: validator-1)
    health              Check health status of all validators
    bft-test            Test Byzantine Fault Tolerance (stops 1 validator)
    clean               Remove all containers, volumes, and data
    ports               Show port mappings for all services
    help                Show this help message

Examples:
    $0 start
    $0 logs validator-2
    $0 exec validator-1 aurad status
    $0 query validator-3
    $0 bft-test

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
