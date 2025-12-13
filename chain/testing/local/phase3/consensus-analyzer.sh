#!/bin/bash
#
# Consensus Analyzer
# Provides detailed analysis of consensus state and voting power distribution
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

VAL1_RPC="http://localhost:26657"
VAL2_RPC="http://localhost:26667"
VAL3_RPC="http://localhost:26677"
VAL4_RPC="http://localhost:26687"

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

# Check if node is running
is_node_running() {
    local rpc_url=$1
    curl -s "$rpc_url/health" > /dev/null 2>&1
}

# Get block height
get_block_height() {
    local rpc_url=$1
    curl -s "$rpc_url/status" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "0"
}

# Get validator info from RPC
get_validator_info() {
    local rpc_url=$1
    curl -s "$rpc_url/validators" 2>/dev/null || echo "{}"
}

# Get consensus state
get_consensus_state() {
    local rpc_url=$1
    curl -s "$rpc_url/consensus_state" 2>/dev/null || echo "{}"
}

# Get net info
get_net_info() {
    local rpc_url=$1
    curl -s "$rpc_url/net_info" 2>/dev/null || echo "{}"
}

# Get block at height
get_block() {
    local rpc_url=$1
    local height=$2
    curl -s "$rpc_url/block?height=$height" 2>/dev/null || echo "{}"
}

# Analyze validator set
analyze_validators() {
    print_header "VALIDATOR SET ANALYSIS"

    # Try each RPC until we find a running node
    local active_rpc=""
    for rpc in "$VAL1_RPC" "$VAL2_RPC" "$VAL3_RPC" "$VAL4_RPC"; do
        if is_node_running "$rpc"; then
            active_rpc="$rpc"
            break
        fi
    done

    if [ -z "$active_rpc" ]; then
        print_error "No validators are running"
        return 1
    fi

    local val_info=$(get_validator_info "$active_rpc")
    local total_validators=$(echo "$val_info" | jq -r '.result.total' 2>/dev/null || echo "0")

    echo "Total Validators in Set: $total_validators"
    echo ""

    if [ "$total_validators" -gt 0 ]; then
        echo "Voting Power Distribution:"
        echo "$val_info" | jq -r '.result.validators[] | "  \(.address[0:12])... : \(.voting_power) (\(.voting_power | tonumber * 100 / 100)%)"' 2>/dev/null || echo "  Unable to parse validator data"

        local total_power=$(echo "$val_info" | jq -r '[.result.validators[].voting_power | tonumber] | add' 2>/dev/null || echo "0")
        echo ""
        echo "Total Voting Power: $total_power"

        # Check if consensus is possible
        local required_power=$(echo "scale=0; $total_power * 2 / 3 + 1" | bc)
        echo "Required for Consensus: $required_power (>2/3)"
    fi
}

# Analyze consensus state
analyze_consensus() {
    print_header "CONSENSUS STATE ANALYSIS"

    local active_count=0
    local validators_info=""

    # Check each validator
    for i in 1 2 3 4; do
        local rpc_var="VAL${i}_RPC"
        local rpc="${!rpc_var}"
        local status="STOPPED"
        local height="N/A"
        local peers="N/A"

        if is_node_running "$rpc"; then
            status="${GREEN}RUNNING${NC}"
            height=$(get_block_height "$rpc")
            local net_info=$(get_net_info "$rpc")
            peers=$(echo "$net_info" | jq -r '.result.n_peers' 2>/dev/null || echo "N/A")
            ((active_count++))
        else
            status="${RED}STOPPED${NC}"
        fi

        validators_info+="Validator $i: $status"
        if [ "$height" != "N/A" ]; then
            validators_info+=" | Height: $height | Peers: $peers"
        fi
        validators_info+="\n"
    done

    echo -e "$validators_info"

    local voting_power_pct=$((active_count * 25))
    echo "Active Validators: $active_count/4"
    echo "Active Voting Power: ${voting_power_pct}%"
    echo ""

    if [ $voting_power_pct -gt 66 ]; then
        print_success "Consensus is ACTIVE (>66% voting power)"
    else
        print_error "Consensus is HALTED (<66% voting power)"
    fi
}

# Analyze block production
analyze_block_production() {
    print_header "BLOCK PRODUCTION ANALYSIS"

    # Find active validator
    local active_rpc=""
    for rpc in "$VAL1_RPC" "$VAL2_RPC" "$VAL3_RPC" "$VAL4_RPC"; do
        if is_node_running "$rpc"; then
            active_rpc="$rpc"
            break
        fi
    done

    if [ -z "$active_rpc" ]; then
        print_error "No validators are running"
        return 1
    fi

    local current_height=$(get_block_height "$active_rpc")
    echo "Current Block Height: $current_height"

    # Get last 10 blocks to analyze proposers
    echo ""
    echo "Recent Block Proposers:"

    local start_height=$((current_height - 9))
    if [ $start_height -lt 1 ]; then
        start_height=1
    fi

    for ((h=start_height; h<=current_height; h++)); do
        local block=$(get_block "$active_rpc" "$h")
        local proposer=$(echo "$block" | jq -r '.result.block.header.proposer_address' 2>/dev/null || echo "unknown")
        local time=$(echo "$block" | jq -r '.result.block.header.time' 2>/dev/null || echo "unknown")
        echo "  Block $h: ${proposer:0:12}... at $time"
    done

    # Monitor production for 10 seconds
    echo ""
    print_info "Monitoring block production for 10 seconds..."
    local start_height=$(get_block_height "$active_rpc")
    sleep 10
    local end_height=$(get_block_height "$active_rpc")
    local blocks_produced=$((end_height - start_height))

    if [ $blocks_produced -gt 0 ]; then
        print_success "Produced $blocks_produced blocks in 10 seconds"
        local rate=$(echo "scale=2; $blocks_produced / 10" | bc)
        echo "Block rate: ${rate} blocks/sec"
    else
        print_error "No blocks produced (chain may be halted)"
    fi
}

# Analyze peer connectivity
analyze_peers() {
    print_header "PEER CONNECTIVITY ANALYSIS"

    for i in 1 2 3 4; do
        local rpc_var="VAL${i}_RPC"
        local rpc="${!rpc_var}"

        echo "Validator $i:"

        if is_node_running "$rpc"; then
            local net_info=$(get_net_info "$rpc")
            local n_peers=$(echo "$net_info" | jq -r '.result.n_peers' 2>/dev/null || echo "0")

            echo "  Status: ${GREEN}RUNNING${NC}"
            echo "  Connected Peers: $n_peers"

            if [ "$n_peers" -gt 0 ]; then
                echo "  Peer Details:"
                echo "$net_info" | jq -r '.result.peers[] | "    - \(.node_info.moniker) (\(.remote_ip))"' 2>/dev/null || echo "    Unable to parse peer data"
            fi
        else
            echo "  Status: ${RED}STOPPED${NC}"
        fi
        echo ""
    done
}

# Generate comprehensive report
generate_report() {
    local output_file="/home/hudson/blockchain-projects/aura/chain/testing/local/phase3/consensus_analysis_$(date +%Y%m%d_%H%M%S).md"

    {
        echo "# Consensus Analysis Report"
        echo ""
        echo "Generated: $(date)"
        echo ""
        echo "## Executive Summary"
        echo ""

        # Count active validators
        local active_count=0
        for rpc in "$VAL1_RPC" "$VAL2_RPC" "$VAL3_RPC" "$VAL4_RPC"; do
            is_node_running "$rpc" && ((active_count++)) || true
        done

        local voting_power=$((active_count * 25))
        echo "- Active Validators: $active_count/4"
        echo "- Active Voting Power: ${voting_power}%"

        if [ $voting_power -gt 66 ]; then
            echo "- Consensus Status: **ACTIVE** (can produce blocks)"
        else
            echo "- Consensus Status: **HALTED** (insufficient voting power)"
        fi
        echo ""

        # Find active RPC
        local active_rpc=""
        for rpc in "$VAL1_RPC" "$VAL2_RPC" "$VAL3_RPC" "$VAL4_RPC"; do
            if is_node_running "$rpc"; then
                active_rpc="$rpc"
                break
            fi
        done

        if [ -n "$active_rpc" ]; then
            local height=$(get_block_height "$active_rpc")
            echo "- Current Block Height: $height"
        else
            echo "- Current Block Height: N/A (no validators running)"
        fi

        echo ""
        echo "## Validator Details"
        echo ""

        for i in 1 2 3 4; do
            local rpc_var="VAL${i}_RPC"
            local rpc="${!rpc_var}"

            echo "### Validator $i"
            echo ""

            if is_node_running "$rpc"; then
                echo "- **Status**: Running"
                echo "- **Block Height**: $(get_block_height "$rpc")"

                local net_info=$(get_net_info "$rpc")
                local n_peers=$(echo "$net_info" | jq -r '.result.n_peers' 2>/dev/null || echo "0")
                echo "- **Connected Peers**: $n_peers"

                local val_info=$(get_validator_info "$rpc")
                local val_count=$(echo "$val_info" | jq -r '.result.total' 2>/dev/null || echo "0")
                echo "- **Validators in Set**: $val_count"
            else
                echo "- **Status**: Stopped"
            fi
            echo ""
        done

        if [ -n "$active_rpc" ]; then
            echo "## Recent Blocks"
            echo ""

            local current_height=$(get_block_height "$active_rpc")
            local start_height=$((current_height - 9))
            if [ $start_height -lt 1 ]; then
                start_height=1
            fi

            echo "| Height | Proposer | Time |"
            echo "|--------|----------|------|"

            for ((h=start_height; h<=current_height; h++)); do
                local block=$(get_block "$active_rpc" "$h")
                local proposer=$(echo "$block" | jq -r '.result.block.header.proposer_address' 2>/dev/null || echo "unknown")
                local time=$(echo "$block" | jq -r '.result.block.header.time' 2>/dev/null || echo "unknown")
                echo "| $h | ${proposer:0:16}... | $time |"
            done
        fi

        echo ""
        echo "## BFT Consensus Properties"
        echo ""
        echo "- **Total Validators**: 4"
        echo "- **Voting Power per Validator**: 25%"
        echo "- **Required for Consensus**: >66% (3+ validators)"
        echo ""
        echo "### Consensus Thresholds"
        echo ""
        echo "| Active Validators | Voting Power | Can Produce Blocks? |"
        echo "|------------------|--------------|---------------------|"
        echo "| 4 | 100% | ✓ Yes |"
        echo "| 3 | 75% | ✓ Yes (>66%) |"
        echo "| 2 | 50% | ✗ No (<66%) |"
        echo "| 1 | 25% | ✗ No (<66%) |"
        echo ""

    } > "$output_file"

    print_success "Report generated: $output_file"
}

# Show menu
show_menu() {
    print_header "CONSENSUS ANALYZER"
    echo "1. Analyze Validator Set"
    echo "2. Analyze Consensus State"
    echo "3. Analyze Block Production"
    echo "4. Analyze Peer Connectivity"
    echo "5. Generate Full Report"
    echo "6. Run All Analyses"
    echo "0. Exit"
    echo ""
}

# Main
main() {
    while true; do
        show_menu
        read -p "Select option: " choice

        case $choice in
            1) analyze_validators ;;
            2) analyze_consensus ;;
            3) analyze_block_production ;;
            4) analyze_peers ;;
            5) generate_report ;;
            6)
                analyze_validators
                analyze_consensus
                analyze_block_production
                analyze_peers
                generate_report
                ;;
            0) exit 0 ;;
            *) print_error "Invalid option" ;;
        esac

        echo ""
        read -p "Press Enter to continue..."
    done
}

# If arguments provided, run specific analysis
if [ $# -gt 0 ]; then
    case $1 in
        validators) analyze_validators ;;
        consensus) analyze_consensus ;;
        blocks) analyze_block_production ;;
        peers) analyze_peers ;;
        report) generate_report ;;
        all)
            analyze_validators
            analyze_consensus
            analyze_block_production
            analyze_peers
            generate_report
            ;;
        *) echo "Usage: $0 [validators|consensus|blocks|peers|report|all]" ;;
    esac
else
    main
fi
