#!/bin/bash
#
# Verify Phase 3 Deliverables
# Ensures all required files are present and properly configured
#

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

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

PHASE3_DIR="/home/hudson/blockchain-projects/aura/chain/testing/local/phase3"
checks_passed=0
checks_total=0

check_file() {
    file=$1
    description=$2
    executable=${3:-false}

    checks_total=$((checks_total + 1))

    if [ -f "$PHASE3_DIR/$file" ]; then
        if [ "$executable" = "true" ]; then
            if [ -x "$PHASE3_DIR/$file" ]; then
                print_success "$description (executable)"
                checks_passed=$((checks_passed + 1))
                return 0
            else
                print_error "$description (not executable)"
                return 1
            fi
        else
            print_success "$description"
            checks_passed=$((checks_passed + 1))
            return 0
        fi
    else
        print_error "$description (missing)"
        return 1
    fi
}

print_header "PHASE 3 DELIVERABLES VERIFICATION"

print_header "1. EXECUTABLE SCRIPTS"

check_file "4-validator-consensus-test.sh" "Main test suite" true
check_file "consensus-analyzer.sh" "Analysis tool" true
check_file "validator-control.sh" "Validator control" true
check_file "check-prerequisites.sh" "Prerequisites checker" true

print_header "2. DOCUMENTATION FILES"

check_file "README.md" "Comprehensive guide" false
check_file "QUICK_START.md" "Quick start guide" false
check_file "INDEX.md" "File index" false
check_file "REPORT_TEMPLATE.md" "Test report template" false
check_file "PHASE3_SUMMARY.md" "Phase 3 summary" false

print_header "3. SCRIPT SYNTAX CHECK"

# Check bash syntax
for script in 4-validator-consensus-test.sh consensus-analyzer.sh validator-control.sh check-prerequisites.sh; do
    checks_total=$((checks_total + 1))
    if [ -f "$PHASE3_DIR/$script" ]; then
        if bash -n "$PHASE3_DIR/$script" 2>/dev/null; then
            print_success "$script has valid bash syntax"
            checks_passed=$((checks_passed + 1))
        else
            print_error "$script has syntax errors"
        fi
    else
        print_error "$script not found for syntax check"
    fi
done

print_header "SUMMARY"

echo "Checks passed: $checks_passed/$checks_total"
echo ""

if [ $checks_passed -eq $checks_total ]; then
    print_success "All deliverables verified successfully!"
    echo ""
    echo "Phase 3 testing suite is complete and ready for use."
    echo ""
    echo "Next steps:"
    echo "  1. Wait for 4-validator setup completion"
    echo "  2. Run: ./check-prerequisites.sh"
    echo "  3. Run: ./4-validator-consensus-test.sh"
    echo ""
    exit 0
else
    failed=$((checks_total - checks_passed))
    print_error "$failed checks failed"
    echo ""
    echo "Some deliverables are missing or incomplete."
    echo "Review the errors above."
    echo ""
    exit 1
fi
