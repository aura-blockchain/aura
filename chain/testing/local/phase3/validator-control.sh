#!/bin/bash
#
# Validator Control Script
# Simplified control and monitoring of 4-validator testnet
#

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
BINARY="/home/hudson/blockchain-projects/aura/chain/aurad"
CHAIN_ID="aura-local-testnet"

# Validator homes and ports
declare -A VAL_HOMES=(
    [1]="/home/hudson/.aura/validator1"
    [2]="/home/hudson/.aura/validator2"
    [3]="/home/hudson/.aura/validator3"
    [4]="/home/hudson/.aura/validator4"
)

declare -A VAL_RPC_PORTS=(
    [1]="26657"
    [2]="26667"
    [3]="26677"
    [4]="26687"
)

declare -A VAL_P2P_PORTS=(
    [1]="26656"
    [2]="26666"
    [3]="26676"
    [4]="26686"
)

declare -A VAL_GRPC_PORTS=(
    [1]="9090"
    [2]="9092"
    [3]="9093"
    [4]="9094"
)

# Helper functions
print_header() {
    echo -e "\n${CYAN}========================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# Check if validator is running
is_validator_running() {
    local val_num=$1
    local rpc_port=${VAL_RPC_PORTS[$val_num]}
    curl -s "http://localhost:$rpc_port/health" > /dev/null 2>&1
}

# Get block height
get_block_height() {
    local val_num=$1
    local rpc_port=${VAL_RPC_PORTS[$val_num]}
    curl -s "http://localhost:$rpc_port/status" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "0"
}

# Get peer count
get_peer_count() {
    local val_num=$1
    local rpc_port=${VAL_RPC_PORTS[$val_num]}
    curl -s "http://localhost:$rpc_port/net_info" | jq -r '.result.n_peers' 2>/dev/null || echo "0"
}

# Start validator
start_validator() {
    local val_num=$1
    local home=${VAL_HOMES[$val_num]}
    local rpc_port=${VAL_RPC_PORTS[$val_num]}
    local p2p_port=${VAL_P2P_PORTS[$val_num]}
    local grpc_port=${VAL_GRPC_PORTS[$val_num]}

    if is_validator_running "$val_num"; then
        print_info "Validator $val_num is already running"
        return 0
    fi

    print_info "Starting validator $val_num..."

    nohup "$BINARY" start \
        --home "$home" \
        --rpc.laddr "tcp://0.0.0.0:$rpc_port" \
        --p2p.laddr "tcp://0.0.0.0:$p2p_port" \
        --grpc.address "0.0.0.0:$grpc_port" \
        > "/home/hudson/blockchain-projects/aura/chain/testing/local/phase3/validator${val_num}.log" 2>&1 &

    sleep 3

    if is_validator_running "$val_num"; then
        print_success "Validator $val_num started"
        return 0
    else
        print_error "Validator $val_num failed to start"
        return 1
    fi
}

# Stop validator
stop_validator() {
    local val_num=$1
    local home=${VAL_HOMES[$val_num]}

    if ! is_validator_running "$val_num"; then
        print_info "Validator $val_num is not running"
        return 0
    fi

    print_info "Stopping validator $val_num..."

    # Find and kill process
    local pid=$(ps aux | grep "aurad start --home $home" | grep -v grep | awk '{print $2}' | head -1)

    if [ -n "$pid" ]; then
        kill "$pid" 2>/dev/null || true
        sleep 2

        if ! is_validator_running "$val_num"; then
            print_success "Validator $val_num stopped"
            return 0
        else
            print_error "Validator $val_num failed to stop gracefully, forcing..."
            kill -9 "$pid" 2>/dev/null || true
            sleep 1
            print_success "Validator $val_num force stopped"
            return 0
        fi
    else
        print_error "Could not find PID for validator $val_num"
        return 1
    fi
}

# Restart validator
restart_validator() {
    local val_num=$1
    stop_validator "$val_num"
    sleep 2
    start_validator "$val_num"
}

# Show status
show_status() {
    print_header "VALIDATOR STATUS"

    local active_count=0
    local total_height=0

    for val_num in 1 2 3 4; do
        local status="STOPPED"
        local height="N/A"
        local peers="N/A"
        local status_color="$RED"

        if is_validator_running "$val_num"; then
            status="RUNNING"
            status_color="$GREEN"
            height=$(get_block_height "$val_num")
            peers=$(get_peer_count "$val_num")
            ((active_count++))
            ((total_height += height))
        fi

        echo -e "Validator $val_num: ${status_color}$status${NC}"
        if [ "$height" != "N/A" ]; then
            echo "  Height: $height | Peers: $peers"
            echo "  RPC: http://localhost:${VAL_RPC_PORTS[$val_num]}"
        fi
    done

    echo ""
    local voting_power=$((active_count * 25))
    echo "Active Validators: $active_count/4"
    echo "Active Voting Power: ${voting_power}%"

    if [ $voting_power -gt 66 ]; then
        print_success "Consensus: ACTIVE (can produce blocks)"
    elif [ $voting_power -eq 0 ]; then
        print_error "Consensus: NO VALIDATORS RUNNING"
    else
        print_error "Consensus: HALTED (insufficient voting power)"
    fi
}

# Show quick status (one line per validator)
quick_status() {
    for val_num in 1 2 3 4; do
        if is_validator_running "$val_num"; then
            local height=$(get_block_height "$val_num")
            local peers=$(get_peer_count "$val_num")
            echo -e "V$val_num: ${GREEN}●${NC} H:$height P:$peers"
        else
            echo -e "V$val_num: ${RED}●${NC} STOPPED"
        fi
    done
}

# Start all validators
start_all() {
    print_header "Starting All Validators"
    for val_num in 1 2 3 4; do
        start_validator "$val_num"
    done
    sleep 3
    show_status
}

# Stop all validators
stop_all() {
    print_header "Stopping All Validators"
    for val_num in 1 2 3 4; do
        stop_validator "$val_num"
    done
    show_status
}

# Restart all validators
restart_all() {
    print_header "Restarting All Validators"
    stop_all
    sleep 3
    start_all
}

# Monitor block production
monitor() {
    print_header "Monitoring Block Production"
    print_info "Press Ctrl+C to stop"
    echo ""

    while true; do
        # Clear previous lines
        tput cuu 6 2>/dev/null || true
        tput ed 2>/dev/null || true

        quick_status

        # Calculate voting power
        local active_count=0
        for val_num in 1 2 3 4; do
            is_validator_running "$val_num" && ((active_count++)) || true
        done
        local voting_power=$((active_count * 25))

        echo ""
        if [ $voting_power -gt 66 ]; then
            echo -e "Consensus: ${GREEN}ACTIVE${NC} (${voting_power}%)"
        else
            echo -e "Consensus: ${RED}HALTED${NC} (${voting_power}%)"
        fi

        sleep 2
    done
}

# Tail logs
tail_logs() {
    local val_num=${1:-1}
    local log_file="/home/hudson/blockchain-projects/aura/chain/testing/local/phase3/validator${val_num}.log"

    if [ -f "$log_file" ]; then
        tail -f "$log_file"
    else
        print_error "Log file not found: $log_file"
        exit 1
    fi
}

# Show help
show_help() {
    cat << EOF
Validator Control Script

Usage: $0 <command> [options]

Commands:
    status              Show validator status
    quick               Show quick status (compact)
    monitor             Monitor block production (live)

    start <1-4|all>     Start validator(s)
    stop <1-4|all>      Stop validator(s)
    restart <1-4|all>   Restart validator(s)

    logs <1-4>          Tail validator logs

Examples:
    $0 status           # Show full status
    $0 start 1          # Start validator 1
    $0 stop all         # Stop all validators
    $0 restart 2        # Restart validator 2
    $0 monitor          # Live monitoring
    $0 logs 3           # Tail validator 3 logs

EOF
}

# Main
main() {
    if [ $# -eq 0 ]; then
        show_help
        exit 0
    fi

    local command=$1
    shift

    case $command in
        status)
            show_status
            ;;
        quick)
            quick_status
            ;;
        monitor)
            monitor
            ;;
        start)
            if [ $# -eq 0 ]; then
                print_error "Specify validator number (1-4) or 'all'"
                exit 1
            fi
            if [ "$1" = "all" ]; then
                start_all
            elif [[ "$1" =~ ^[1-4]$ ]]; then
                start_validator "$1"
            else
                print_error "Invalid validator number. Use 1-4 or 'all'"
                exit 1
            fi
            ;;
        stop)
            if [ $# -eq 0 ]; then
                print_error "Specify validator number (1-4) or 'all'"
                exit 1
            fi
            if [ "$1" = "all" ]; then
                stop_all
            elif [[ "$1" =~ ^[1-4]$ ]]; then
                stop_validator "$1"
            else
                print_error "Invalid validator number. Use 1-4 or 'all'"
                exit 1
            fi
            ;;
        restart)
            if [ $# -eq 0 ]; then
                print_error "Specify validator number (1-4) or 'all'"
                exit 1
            fi
            if [ "$1" = "all" ]; then
                restart_all
            elif [[ "$1" =~ ^[1-4]$ ]]; then
                restart_validator "$1"
            else
                print_error "Invalid validator number. Use 1-4 or 'all'"
                exit 1
            fi
            ;;
        logs)
            if [ $# -eq 0 ]; then
                print_error "Specify validator number (1-4)"
                exit 1
            fi
            if [[ "$1" =~ ^[1-4]$ ]]; then
                tail_logs "$1"
            else
                print_error "Invalid validator number. Use 1-4"
                exit 1
            fi
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
