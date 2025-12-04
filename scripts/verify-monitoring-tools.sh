#!/bin/bash
# ============================================================================
# AURA Testnet Monitoring Tools Verification Script
# ============================================================================
# Verifies that all monitoring tools are installed and functional
# ============================================================================

set -euo pipefail

# Colors
readonly GREEN='\033[0;32m'
readonly RED='\033[0;31m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

CHECKS_PASSED=0
CHECKS_FAILED=0

# ============================================================================
# Helper Functions
# ============================================================================

print_header() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}\n"
}

print_check() {
    echo -n "$1... "
}

print_pass() {
    echo -e "${GREEN}✓ PASS${NC}"
    CHECKS_PASSED=$((CHECKS_PASSED + 1))
}

print_fail() {
    echo -e "${RED}✗ FAIL${NC}"
    if [[ -n "${1:-}" ]]; then
        echo -e "${RED}  Error: $1${NC}"
    fi
    CHECKS_FAILED=$((CHECKS_FAILED + 1))
}

print_summary() {
    local total=$((CHECKS_PASSED + CHECKS_FAILED))
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}VERIFICATION SUMMARY${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════${NC}"
    echo -e "Total checks:  $total"
    echo -e "${GREEN}Passed:        $CHECKS_PASSED${NC}"
    echo -e "${RED}Failed:        $CHECKS_FAILED${NC}"
    echo ""

    if [[ $CHECKS_FAILED -eq 0 ]]; then
        echo -e "${GREEN}✓ All monitoring tools verified and operational!${NC}"
        return 0
    else
        echo -e "${RED}✗ Some checks failed. Please review errors above.${NC}"
        return 1
    fi
}

# ============================================================================
# Verification Checks
# ============================================================================

check_script_exists() {
    local script=$1
    print_check "Checking if ${script} exists"

    if [[ -f "$script" ]]; then
        print_pass
    else
        print_fail "Script not found: $script"
    fi
}

check_script_executable() {
    local script=$1
    print_check "Checking if ${script} is executable"

    if [[ -x "$script" ]]; then
        print_pass
    else
        print_fail "Script not executable: $script"
    fi
}

check_script_help() {
    local script=$1
    local help_cmd=${2:-help}
    print_check "Checking if ${script} ${help_cmd} works"

    if timeout 5 "$script" "$help_cmd" >/dev/null 2>&1; then
        print_pass
    else
        print_fail "Help command failed or timed out"
    fi
}

check_documentation() {
    local doc=$1
    print_check "Checking if ${doc} exists"

    if [[ -f "$doc" ]]; then
        print_pass
    else
        print_fail "Documentation not found: $doc"
    fi
}

check_documentation_content() {
    local doc=$1
    local keyword=$2
    print_check "Checking if ${doc} contains '${keyword}'"

    if grep -q "$keyword" "$doc" 2>/dev/null; then
        print_pass
    else
        print_fail "Keyword not found in documentation"
    fi
}

# ============================================================================
# Main Verification
# ============================================================================

main() {
    print_header "AURA Testnet Monitoring Tools Verification"

    # Check main monitoring script
    print_header "Checking testnet-monitor.sh"
    check_script_exists "./scripts/testnet-monitor.sh"
    check_script_executable "./scripts/testnet-monitor.sh"
    check_script_help "./scripts/testnet-monitor.sh" "help"

    # Verify individual commands exist in help output
    print_check "Verifying all monitoring commands are documented"
    local help_output=$(./scripts/testnet-monitor.sh help 2>/dev/null || echo "")
    local commands=("quick" "watch" "watch-blocks" "status" "logs" "check-logs" "performance" "network" "diagnose" "restart-failed" "shutdown")
    local all_found=true

    for cmd in "${commands[@]}"; do
        if ! echo "$help_output" | grep -q "$cmd"; then
            all_found=false
            break
        fi
    done

    if [[ "$all_found" == true ]]; then
        print_pass
    else
        print_fail "Not all commands found in help output"
    fi

    # Check continuous monitor script
    print_header "Checking continuous-monitor.sh"
    check_script_exists "./scripts/continuous-monitor.sh"
    check_script_executable "./scripts/continuous-monitor.sh"

    # Check updated testnet-manage.sh
    print_header "Checking testnet-manage.sh integration"
    check_script_exists "./scripts/testnet-manage.sh"
    check_documentation_content "./scripts/testnet-manage.sh" "testnet-monitor.sh"

    # Check documentation
    print_header "Checking Documentation"
    check_documentation "TESTNET_MONITORING_GUIDE.md"
    check_documentation "MONITORING_CHEATSHEET.md"
    check_documentation "MONITORING_TOOLS_SUMMARY.md"
    check_documentation "docs/runbooks/TESTNET_MONITORING.md"

    # Verify key content in documentation
    print_header "Verifying Documentation Content"
    check_documentation_content "TESTNET_MONITORING_GUIDE.md" "Quick Start"
    check_documentation_content "TESTNET_MONITORING_GUIDE.md" "Health Check"
    check_documentation_content "TESTNET_MONITORING_GUIDE.md" "Performance"
    check_documentation_content "TESTNET_MONITORING_GUIDE.md" "Troubleshooting"

    check_documentation_content "MONITORING_CHEATSHEET.md" "Essential Commands"
    check_documentation_content "MONITORING_CHEATSHEET.md" "Cheat Sheet"

    check_documentation_content "MONITORING_TOOLS_SUMMARY.md" "What Was Delivered"
    check_documentation_content "MONITORING_TOOLS_SUMMARY.md" "Quick Start Examples"

    check_documentation_content "docs/runbooks/TESTNET_MONITORING.md" "Daily Monitoring Workflow"
    check_documentation_content "docs/runbooks/TESTNET_MONITORING.md" "Incident Response"

    # Check Docker Compose integration
    print_header "Checking Docker Compose Configuration"
    check_documentation "docker-compose.testnet.yml"
    check_documentation_content "docker-compose.testnet.yml" "validator-1"
    check_documentation_content "docker-compose.testnet.yml" "validator-2"
    check_documentation_content "docker-compose.testnet.yml" "validator-3"
    check_documentation_content "docker-compose.testnet.yml" "validator-4"
    check_documentation_content "docker-compose.testnet.yml" "prometheus-testnet"
    check_documentation_content "docker-compose.testnet.yml" "grafana-testnet"

    # Check Prometheus configuration
    print_header "Checking Prometheus Configuration"
    check_documentation "prometheus/prometheus-testnet.yml"
    check_documentation_content "prometheus/prometheus-testnet.yml" "validator-1"

    # Check ROADMAP update
    print_header "Checking ROADMAP Update"
    check_documentation "ROADMAP_PRODUCTION.md"
    check_documentation_content "ROADMAP_PRODUCTION.md" "testnet-monitor.sh"

    # Print summary
    print_summary
}

main

exit $CHECKS_FAILED
