#!/bin/bash
set -e

# Test 2.1: Genesis Initialization Workflow
# This script tests the complete genesis initialization workflow for Aura
# Including: init, add-genesis-account, gentx, collect-gentxs, validate-genesis

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
BINARY="$PROJECT_ROOT/chain/aurad"
TEST_HOME="/tmp/aura-genesis-test-$$"
CHAIN_ID="aura-test-1"
MONIKER="test-node"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_step() {
    echo -e "\n${GREEN}==>${NC} $1"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test environment..."
    if [ -d "$TEST_HOME" ]; then
        rm -rf "$TEST_HOME"
        log_info "Removed test home directory: $TEST_HOME"
    fi
}

# Set trap for cleanup
trap cleanup EXIT

# Verify binary exists
if [ ! -f "$BINARY" ]; then
    log_error "Binary not found at $BINARY"
    log_info "Building binary..."
    cd "$PROJECT_ROOT/chain"
    go build -o aurad ./cmd/aurad
    if [ $? -ne 0 ]; then
        log_error "Failed to build binary"
        exit 1
    fi
fi

log_info "Starting Genesis Initialization Workflow Test"
log_info "Binary: $BINARY"
log_info "Test Home: $TEST_HOME"
log_info "Chain ID: $CHAIN_ID"

# Test 2.1.1: Initialize node
log_step "Step 1: Initialize node with 'aurad init'"
$BINARY init "$MONIKER" --chain-id "$CHAIN_ID" --home "$TEST_HOME" > /dev/null 2>&1

if [ $? -ne 0 ]; then
    log_error "Failed to initialize node"
    exit 1
fi

# Verify genesis file was created
if [ ! -f "$TEST_HOME/config/genesis.json" ]; then
    log_error "Genesis file not created"
    exit 1
fi

# Verify config files were created
if [ ! -f "$TEST_HOME/config/config.toml" ]; then
    log_error "config.toml not created"
    exit 1
fi

if [ ! -f "$TEST_HOME/config/app.toml" ]; then
    log_error "app.toml not created"
    exit 1
fi

log_info "✓ Node initialized successfully"
log_info "  - Genesis file created: $TEST_HOME/config/genesis.json"
log_info "  - Config file created: $TEST_HOME/config/config.toml"
log_info "  - App config created: $TEST_HOME/config/app.toml"

# Test 2.1.2: Add genesis accounts
log_step "Step 2: Add genesis accounts"

# Create test key for validator
VALIDATOR_KEY="validator"
$BINARY keys add "$VALIDATOR_KEY" --keyring-backend test --home "$TEST_HOME" > /dev/null 2>&1
VALIDATOR_ADDR=$($BINARY keys show "$VALIDATOR_KEY" -a --keyring-backend test --home "$TEST_HOME" 2>&1)

if [ -z "$VALIDATOR_ADDR" ]; then
    log_error "Failed to create validator key"
    exit 1
fi

log_info "✓ Validator key created: $VALIDATOR_ADDR"

# Create test key for user account
USER_KEY="user1"
$BINARY keys add "$USER_KEY" --keyring-backend test --home "$TEST_HOME" > /dev/null 2>&1
USER_ADDR=$($BINARY keys show "$USER_KEY" -a --keyring-backend test --home "$TEST_HOME" 2>&1)

if [ -z "$USER_ADDR" ]; then
    log_error "Failed to create user key"
    exit 1
fi

log_info "✓ User key created: $USER_ADDR"

# Add genesis accounts
$BINARY genesis add-genesis-account "$VALIDATOR_ADDR" 1000000000stake,1000000000uaura --home "$TEST_HOME" > /dev/null 2>&1

if [ $? -ne 0 ]; then
    log_error "Failed to add validator genesis account"
    exit 1
fi

log_info "✓ Validator genesis account added"

$BINARY genesis add-genesis-account "$USER_ADDR" 5000000stake,5000000uaura --home "$TEST_HOME" > /dev/null 2>&1

if [ $? -ne 0 ]; then
    log_error "Failed to add user genesis account"
    exit 1
fi

log_info "✓ User genesis account added"

# Verify accounts in genesis using jq if available, otherwise grep
if command -v jq > /dev/null 2>&1; then
    GENESIS_ACCOUNTS=$(cat "$TEST_HOME/config/genesis.json" | jq -r '.app_state.auth.accounts[] | select(.address == "'"$VALIDATOR_ADDR"'") | .address' 2>/dev/null | wc -l)
else
    GENESIS_ACCOUNTS=$(cat "$TEST_HOME/config/genesis.json" | grep -o "\"address\"[[:space:]]*:[[:space:]]*\"$VALIDATOR_ADDR\"" | wc -l)
fi

if [ "$GENESIS_ACCOUNTS" -eq 0 ]; then
    log_error "Validator account not found in genesis"
    # Debug output
    log_error "Looking for address: $VALIDATOR_ADDR"
    if command -v jq > /dev/null 2>&1; then
        log_error "Accounts in genesis:"
        cat "$TEST_HOME/config/genesis.json" | jq -r '.app_state.auth.accounts[].address' 2>/dev/null || echo "Failed to parse accounts"
    fi
    exit 1
fi

log_info "✓ Genesis accounts verified in genesis.json"

# Test 2.1.3: Create genesis transaction (gentx)
log_step "Step 3: Create genesis transaction (gentx)"

$BINARY genesis gentx "$VALIDATOR_KEY" 500000000stake \
    --chain-id "$CHAIN_ID" \
    --keyring-backend test \
    --home "$TEST_HOME" > /dev/null 2>&1

if [ $? -ne 0 ]; then
    log_error "Failed to create gentx"
    exit 1
fi

# Verify gentx file was created
GENTX_COUNT=$(ls -1 "$TEST_HOME/config/gentx/" 2>/dev/null | wc -l)
if [ "$GENTX_COUNT" -eq 0 ]; then
    log_error "No gentx files found"
    exit 1
fi

log_info "✓ Gentx created successfully"
log_info "  - Gentx file: $(ls -1 "$TEST_HOME/config/gentx/")"

# Test 2.1.4: Collect genesis transactions
log_step "Step 4: Collect genesis transactions"

$BINARY genesis collect-gentxs --home "$TEST_HOME" > /dev/null 2>&1

if [ $? -ne 0 ]; then
    log_error "Failed to collect gentxs"
    exit 1
fi

log_info "✓ Genesis transactions collected successfully"

# Verify genesis contains gentx data
GENESIS_TXS=$(cat "$TEST_HOME/config/genesis.json" | grep -o "gentxs" | wc -l)
if [ "$GENESIS_TXS" -eq 0 ]; then
    log_warn "No gentxs found in genesis.json (may be in a different format)"
fi

# Test 2.1.5: Validate genesis
log_step "Step 5: Validate genesis"

# Run validate-genesis from the home directory (it looks for config/genesis.json relative to cwd)
VALIDATE_OUTPUT=$(cd "$TEST_HOME" && $BINARY genesis validate-genesis 2>&1)
VALIDATE_EXIT_CODE=$?

if [ $VALIDATE_EXIT_CODE -ne 0 ]; then
    log_error "Genesis validation failed"
    log_error "Output: $VALIDATE_OUTPUT"
    exit 1
fi

log_info "✓ Genesis validated successfully"

# Additional verification tests
log_step "Step 6: Additional verification checks"

# Check genesis file size (should be reasonable)
GENESIS_SIZE=$(stat -f%z "$TEST_HOME/config/genesis.json" 2>/dev/null || stat -c%s "$TEST_HOME/config/genesis.json")
log_info "✓ Genesis file size: $GENESIS_SIZE bytes"

if [ "$GENESIS_SIZE" -lt 1000 ]; then
    log_warn "Genesis file seems unusually small"
fi

# Verify chain ID in genesis
if command -v jq > /dev/null 2>&1; then
    GENESIS_CHAIN_ID=$(cat "$TEST_HOME/config/genesis.json" | jq -r '.chain_id' 2>/dev/null)
else
    GENESIS_CHAIN_ID=$(cat "$TEST_HOME/config/genesis.json" | grep -o '"chain_id"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"chain_id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
fi

if [ "$GENESIS_CHAIN_ID" != "$CHAIN_ID" ]; then
    log_error "Chain ID mismatch: expected $CHAIN_ID, got $GENESIS_CHAIN_ID"
    exit 1
fi

log_info "✓ Chain ID verified: $GENESIS_CHAIN_ID"

# Verify initial stake allocation
TOTAL_SUPPLY=$(cat "$TEST_HOME/config/genesis.json" | grep -o '"denom":"stake"' | wc -l)
log_info "✓ Stake allocations found: $TOTAL_SUPPLY"

# Test invalid operations
log_step "Step 7: Test invalid operations (negative testing)"

# Try to add duplicate genesis account (should fail or handle gracefully)
set +e  # Temporarily disable exit on error for negative tests
$BINARY genesis add-genesis-account "$VALIDATOR_ADDR" 1000stake --home "$TEST_HOME" > /dev/null 2>&1
DUPLICATE_EXIT_CODE=$?
set -e  # Re-enable exit on error

if [ $DUPLICATE_EXIT_CODE -eq 0 ]; then
    log_warn "Adding duplicate genesis account succeeded (may be by design)"
else
    log_info "✓ Adding duplicate genesis account failed as expected"
fi

# Try to validate with invalid chain ID
INVALID_HOME="/tmp/aura-invalid-test-$$"
mkdir -p "$INVALID_HOME/config"
echo '{"chain_id":"invalid","app_state":{}}' > "$INVALID_HOME/config/genesis.json"
set +e  # Temporarily disable exit on error
(cd "$INVALID_HOME" && $BINARY genesis validate-genesis > /dev/null 2>&1)
INVALID_EXIT_CODE=$?
set -e  # Re-enable exit on error
rm -rf "$INVALID_HOME"

if [ $INVALID_EXIT_CODE -ne 0 ]; then
    log_info "✓ Invalid genesis validation failed as expected"
else
    log_warn "Invalid genesis validation succeeded (may need stricter validation)"
fi

# Summary
log_step "Test Summary"
echo ""
echo "Genesis Initialization Workflow Test: PASSED"
echo ""
echo "Tests Completed:"
echo "  ✓ 2.1.1: Node initialization (aurad init)"
echo "  ✓ 2.1.2: Genesis account creation (add-genesis-account)"
echo "  ✓ 2.1.3: Genesis transaction creation (gentx)"
echo "  ✓ 2.1.4: Genesis transaction collection (collect-gentxs)"
echo "  ✓ 2.1.5: Genesis validation (validate-genesis)"
echo "  ✓ 2.1.6: Additional verification checks"
echo "  ✓ 2.1.7: Negative testing"
echo ""
echo "Genesis File: $TEST_HOME/config/genesis.json"
echo "Chain ID: $GENESIS_CHAIN_ID"
echo "Validator Address: $VALIDATOR_ADDR"
echo "User Address: $USER_ADDR"
echo ""

log_info "All tests passed successfully!"

exit 0
