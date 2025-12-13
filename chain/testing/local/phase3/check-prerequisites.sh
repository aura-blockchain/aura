#!/bin/bash
#
# Prerequisites Checker
# Verifies that all requirements are met before running consensus tests
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

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Check counter
checks_passed=0
checks_failed=0
checks_total=0

check() {
    local description=$1
    local command=$2
    ((checks_total++))

    if eval "$command" > /dev/null 2>&1; then
        print_success "$description"
        ((checks_passed++))
        return 0
    else
        print_error "$description"
        ((checks_failed++))
        return 1
    fi
}

check_with_output() {
    local description=$1
    shift
    local command="$@"
    ((checks_total++))

    if $command > /dev/null 2>&1; then
        print_success "$description"
        ((checks_passed++))
        return 0
    else
        print_error "$description"
        ((checks_failed++))
        return 1
    fi
}

# Main checks
main() {
    print_header "PHASE 3 PREREQUISITES CHECKER"
    echo "Checking prerequisites for 4-validator consensus testing..."
    echo ""

    # 1. Environment checks
    print_header "1. ENVIRONMENT CHECKS"

    check "Current directory is aura project" "[ -d /home/hudson/blockchain-projects/aura ]"

    if [ "${GOCACHE:-}" = "/home/hudson/blockchain-projects/aura/.cache/go-build" ]; then
        print_success "Environment sourced (GOCACHE set correctly)"
        ((checks_passed++))
    else
        print_error "Environment not sourced (run: source env.sh)"
        print_info "  Current GOCACHE: ${GOCACHE:-not set}"
        print_info "  Expected: /home/hudson/blockchain-projects/aura/.cache/go-build"
        ((checks_failed++))
    fi
    ((checks_total++))

    # 2. Binary checks
    print_header "2. BINARY CHECKS"

    if [ -f "$BINARY" ]; then
        print_success "Binary exists: $BINARY"
        ((checks_passed++))

        if [ -x "$BINARY" ]; then
            print_success "Binary is executable"
            ((checks_passed++))
        else
            print_error "Binary is not executable"
            print_info "  Run: chmod +x $BINARY"
            ((checks_failed++))
        fi
    else
        print_error "Binary not found: $BINARY"
        print_info "  Build with: cd chain && make build"
        ((checks_failed++))
        checks_total=$((checks_total + 1))
    fi
    checks_total=$((checks_total + 1))

    # 3. Validator home directories
    print_header "3. VALIDATOR HOME DIRECTORIES"

    for i in 1 2 3 4; do
        local home="/home/hudson/.aura/validator$i"
        if [ -d "$home" ]; then
            print_success "Validator $i home exists: $home"
            ((checks_passed++))

            # Check critical files
            if [ -f "$home/config/genesis.json" ]; then
                print_success "  genesis.json exists"
                ((checks_passed++))
            else
                print_error "  genesis.json missing"
                ((checks_failed++))
            fi

            if [ -f "$home/config/config.toml" ]; then
                print_success "  config.toml exists"
                ((checks_passed++))
            else
                print_error "  config.toml missing"
                ((checks_failed++))
            fi

            if [ -f "$home/config/node_key.json" ]; then
                print_success "  node_key.json exists"
                ((checks_passed++))
            else
                print_error "  node_key.json missing"
                ((checks_failed++))
            fi
        else
            print_error "Validator $i home not found: $home"
            print_info "  Run the 4-validator setup script first"
            ((checks_failed++))
            checks_total=$((checks_total + 3))
        fi
        checks_total=$((checks_total + 4))
    done

    # 4. Port availability
    print_header "4. PORT AVAILABILITY"

    local ports=(26657 26667 26677 26687 26656 26666 26676 26686 9090 9092 9093 9094)
    for port in "${ports[@]}"; do
        if ! netstat -tln 2>/dev/null | grep -q ":$port "; then
            print_success "Port $port available"
            ((checks_passed++))
        else
            print_warning "Port $port in use (validator may be running)"
            # Don't count as failure - might be intentional
            ((checks_passed++))
        fi
        ((checks_total++))
    done

    # 5. Required tools
    print_header "5. REQUIRED TOOLS"

    check_with_output "curl installed" command -v curl
    check_with_output "jq installed" command -v jq
    check_with_output "bc installed" command -v bc
    check_with_output "netstat available" command -v netstat

    # 6. Validator configurations
    print_header "6. VALIDATOR CONFIGURATIONS"

    # Check chain ID consistency
    local chain_id_consistent=true
    for i in 1 2 3 4; do
        local genesis="/home/hudson/.aura/validator$i/config/genesis.json"
        if [ -f "$genesis" ]; then
            local cid=$(jq -r '.chain_id' "$genesis" 2>/dev/null || echo "")
            if [ "$cid" = "$CHAIN_ID" ]; then
                print_success "Validator $i has correct chain ID: $CHAIN_ID"
                ((checks_passed++))
            else
                print_error "Validator $i has incorrect chain ID: $cid (expected: $CHAIN_ID)"
                ((checks_failed++))
                chain_id_consistent=false
            fi
        else
            print_error "Validator $i genesis.json not found"
            ((checks_failed++))
            chain_id_consistent=false
        fi
        ((checks_total++))
    done

    # Check validator set in genesis
    local genesis1="/home/hudson/.aura/validator1/config/genesis.json"
    if [ -f "$genesis1" ]; then
        local val_count=$(jq '.validators | length' "$genesis1" 2>/dev/null || echo "0")
        if [ "$val_count" -eq 4 ]; then
            print_success "Genesis has 4 validators"
            ((checks_passed++))
        else
            print_error "Genesis has $val_count validators (expected 4)"
            ((checks_failed++))
        fi
        ((checks_total++))
    fi

    # 7. Test scripts
    print_header "7. TEST SCRIPTS"

    local scripts=(
        "4-validator-consensus-test.sh"
        "consensus-analyzer.sh"
        "validator-control.sh"
    )

    for script in "${scripts[@]}"; do
        local script_path="/home/hudson/blockchain-projects/aura/chain/testing/local/phase3/$script"
        if [ -f "$script_path" ]; then
            if [ -x "$script_path" ]; then
                print_success "$script exists and is executable"
                ((checks_passed++))
            else
                print_warning "$script exists but is not executable"
                print_info "  Run: chmod +x $script_path"
                ((checks_failed++))
            fi
        else
            print_error "$script not found"
            ((checks_failed++))
        fi
        ((checks_total++))
    done

    # Summary
    print_header "SUMMARY"

    echo "Checks passed: $checks_passed/$checks_total"
    echo ""

    if [ $checks_failed -eq 0 ]; then
        print_success "All prerequisites met! Ready to run consensus tests."
        echo ""
        echo "Next steps:"
        echo "  1. cd /home/hudson/blockchain-projects/aura/chain/testing/local/phase3"
        echo "  2. ./4-validator-consensus-test.sh"
        echo ""
        echo "Or for manual testing:"
        echo "  ./validator-control.sh start all"
        echo "  ./validator-control.sh status"
        echo ""
        return 0
    else
        print_error "$checks_failed checks failed"
        echo ""
        echo "Fix the issues above before running consensus tests."
        echo ""

        if ! [ -f "$BINARY" ]; then
            echo "Most common fix: Build the binary"
            echo "  cd /home/hudson/blockchain-projects/aura/chain"
            echo "  make build"
            echo ""
        fi

        if ! [ -d "/home/hudson/.aura/validator1" ]; then
            echo "Most common fix: Run 4-validator setup script"
            echo "  (Should be provided by Phase 3 setup agent)"
            echo ""
        fi

        return 1
    fi
}

main "$@"
