#!/bin/bash
#
# Test 6.2.1: Start Bitcoin regtest node and fund wallet
#
# This test verifies that Bitcoin Core can be started in regtest mode
# and that we can create addresses and mine blocks to fund wallets.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RESULTS_FILE="${SCRIPT_DIR}/test_6.2.1_results.txt"
BTC_CLI="bitcoin-cli -regtest"

echo "=================================="
echo "Test 6.2.1: Bitcoin Regtest Setup"
echo "=================================="
echo ""

# Clear previous results
> "$RESULTS_FILE"

log_result() {
    echo "$1" | tee -a "$RESULTS_FILE"
}

log_result "Test Start: $(date)"
log_result ""

# Check if bitcoind is already running
if pgrep -x bitcoind > /dev/null; then
    log_result "✓ Bitcoin daemon already running"
else
    log_result "Starting Bitcoin daemon in regtest mode..."
    bitcoind -regtest -daemon -fallbackfee=0.00001
    sleep 3

    if pgrep -x bitcoind > /dev/null; then
        log_result "✓ Bitcoin daemon started successfully"
    else
        log_result "✗ Failed to start Bitcoin daemon"
        exit 1
    fi
fi

log_result ""

# Wait for bitcoind to be ready
log_result "Waiting for Bitcoin daemon to be ready..."
for i in {1..30}; do
    if $BTC_CLI getblockchaininfo &> /dev/null; then
        log_result "✓ Bitcoin daemon is ready (${i}s)"
        break
    fi
    if [ $i -eq 30 ]; then
        log_result "✗ Bitcoin daemon did not become ready in time"
        exit 1
    fi
    sleep 1
done

log_result ""

# Get blockchain info
log_result "=== Blockchain Information ==="
CHAIN_INFO=$($BTC_CLI getblockchaininfo)
BLOCKS=$(echo "$CHAIN_INFO" | grep '"blocks"' | awk '{print $2}' | tr -d ',')
log_result "Current block height: $BLOCKS"
log_result ""

# Create or get a wallet
log_result "=== Wallet Setup ==="
WALLET_NAME="testwallet"

# Try to load wallet if it exists
if $BTC_CLI loadwallet "$WALLET_NAME" 2>/dev/null; then
    log_result "✓ Loaded existing wallet: $WALLET_NAME"
elif $BTC_CLI createwallet "$WALLET_NAME" 2>/dev/null; then
    log_result "✓ Created new wallet: $WALLET_NAME"
else
    # Wallet might already be loaded
    log_result "✓ Wallet already loaded or exists: $WALLET_NAME"
fi

log_result ""

# Generate a new address
log_result "=== Address Generation ==="
ADDRESS=$($BTC_CLI -rpcwallet="$WALLET_NAME" getnewaddress "test-address")
log_result "Generated address: $ADDRESS"
log_result ""

# Mine blocks to the address to fund it
log_result "=== Mining Blocks ==="
log_result "Mining 101 blocks to address (needed for coinbase maturity)..."
BLOCK_HASHES=$($BTC_CLI -rpcwallet="$WALLET_NAME" generatetoaddress 101 "$ADDRESS")
NEW_BLOCKS=$(echo "$BLOCK_HASHES" | wc -l)
log_result "✓ Mined $NEW_BLOCKS blocks"
log_result ""

# Check the balance
BALANCE=$($BTC_CLI -rpcwallet="$WALLET_NAME" getbalance)
log_result "Wallet balance: $BALANCE BTC"

if (( $(echo "$BALANCE > 0" | bc -l) )); then
    log_result "✓ Wallet successfully funded"
else
    log_result "✗ Wallet not funded"
    exit 1
fi

log_result ""

# Get updated blockchain info
CHAIN_INFO=$($BTC_CLI getblockchaininfo)
NEW_BLOCK_HEIGHT=$(echo "$CHAIN_INFO" | grep '"blocks"' | awk '{print $2}' | tr -d ',')
log_result "New block height: $NEW_BLOCK_HEIGHT"
log_result "Blocks mined: $((NEW_BLOCK_HEIGHT - BLOCKS))"
log_result ""

# Create a second address for testing
log_result "=== Second Address for Testing ==="
ADDRESS2=$($BTC_CLI -rpcwallet="$WALLET_NAME" getnewaddress "test-address-2")
log_result "Generated second address: $ADDRESS2"
log_result ""

# Test sending BTC to the second address
log_result "=== Test Transaction ==="
SEND_AMOUNT="10.0"
log_result "Sending $SEND_AMOUNT BTC to second address..."
TXID=$($BTC_CLI -rpcwallet="$WALLET_NAME" sendtoaddress "$ADDRESS2" "$SEND_AMOUNT")
log_result "Transaction ID: $TXID"
log_result ""

# Mine a block to confirm the transaction
log_result "Mining 1 block to confirm transaction..."
$BTC_CLI -rpcwallet="$WALLET_NAME" generatetoaddress 1 "$ADDRESS" > /dev/null
log_result "✓ Transaction confirmed"
log_result ""

# Verify the transaction
TX_INFO=$($BTC_CLI gettransaction "$TXID")
CONFIRMATIONS=$(echo "$TX_INFO" | grep '"confirmations"' | awk '{print $2}' | tr -d ',')
log_result "Transaction confirmations: $CONFIRMATIONS"

if [ "$CONFIRMATIONS" -ge 1 ]; then
    log_result "✓ Transaction successfully confirmed"
else
    log_result "✗ Transaction not confirmed"
    exit 1
fi

log_result ""

# Final balance check
FINAL_BALANCE=$($BTC_CLI -rpcwallet="$WALLET_NAME" getbalance)
log_result "Final wallet balance: $FINAL_BALANCE BTC"
log_result ""

# Display connection info for atomic swap tests
log_result "=== Bitcoin RPC Connection Info ==="
log_result "RPC URL: http://127.0.0.1:18443"
log_result "RPC User: btcuser"
log_result "RPC Password: btcpass123"
log_result "Wallet: $WALLET_NAME"
log_result "Test Address 1: $ADDRESS"
log_result "Test Address 2: $ADDRESS2"
log_result ""

log_result "Test End: $(date)"
log_result ""
log_result "=== OVERALL RESULT: PASS ==="
log_result ""

echo ""
echo "Results saved to: $RESULTS_FILE"
echo ""
echo "✓ Bitcoin regtest node is ready for atomic swap testing"
echo "✓ Wallet funded with $(echo "$FINAL_BALANCE" | awk '{printf "%.2f", $1}') BTC"
