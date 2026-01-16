#!/bin/bash
# AURA MVP Coordinated Launch Script
# Starts all validators in sequence for chain launch

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Configuration
CHAIN_ID="${CHAIN_ID:-aura-mvp-1}"
START_DELAY="${START_DELAY:-30}"  # Seconds between validator starts

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "[INFO] $1"; }
log_ok() { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo "=========================================="
echo "AURA MVP Coordinated Launch"
echo "=========================================="
echo "Chain ID: $CHAIN_ID"
echo "Start delay: ${START_DELAY}s between validators"
echo "=========================================="
echo

# Pre-flight checks
log_info "Running pre-flight checks..."

# Check SSH connectivity
for server in aura-testnet services-testnet; do
    if ssh -o ConnectTimeout=5 "$server" "echo OK" &>/dev/null; then
        log_ok "SSH to $server"
    else
        log_error "Cannot connect to $server"
        exit 1
    fi
done

# Check genesis files are in place
log_info "Verifying genesis files..."
GENESIS_HASH=""
for server_val in "aura-testnet:val1" "aura-testnet:val2" "services-testnet:val3" "services-testnet:val4"; do
    IFS=':' read -r server val <<< "$server_val"
    home="~/.aura-$val"

    hash=$(ssh "$server" "sha256sum $home/config/genesis.json 2>/dev/null | awk '{print \$1}'" 2>/dev/null || echo "NOT_FOUND")

    if [ -z "$GENESIS_HASH" ]; then
        GENESIS_HASH="$hash"
    fi

    if [ "$hash" = "$GENESIS_HASH" ] && [ "$hash" != "NOT_FOUND" ]; then
        log_ok "Genesis on $server ($val)"
    else
        log_error "Genesis mismatch or missing on $server ($val)"
        exit 1
    fi
done

echo
log_info "Genesis checksum: $GENESIS_HASH"
echo

# Confirm launch
read -p "Ready to launch chain? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "Launch aborted"
    exit 0
fi

echo
log_info "Starting coordinated launch..."
echo

# Start validators in sequence
# Validator 1 (seed node)
log_info "Starting validator 1 on aura-testnet..."
ssh aura-testnet "nohup ~/.aura/cosmovisor/genesis/bin/aurad start --home ~/.aura-val1 > ~/.aura-val1/node.log 2>&1 &"
log_ok "Validator 1 started"
echo "Waiting ${START_DELAY}s..."
sleep "$START_DELAY"

# Validator 2
log_info "Starting validator 2 on aura-testnet..."
ssh aura-testnet "nohup ~/.aura/cosmovisor/genesis/bin/aurad start --home ~/.aura-val2 > ~/.aura-val2/node.log 2>&1 &"
log_ok "Validator 2 started"
echo "Waiting ${START_DELAY}s..."
sleep "$START_DELAY"

# Validator 3
log_info "Starting validator 3 on services-testnet..."
ssh services-testnet "nohup ~/.aura/cosmovisor/genesis/bin/aurad start --home ~/.aura-val3 > ~/.aura-val3/node.log 2>&1 &"
log_ok "Validator 3 started"
echo "Waiting ${START_DELAY}s..."
sleep "$START_DELAY"

# Validator 4
log_info "Starting validator 4 on services-testnet..."
ssh services-testnet "nohup ~/.aura/cosmovisor/genesis/bin/aurad start --home ~/.aura-val4 > ~/.aura-val4/node.log 2>&1 &"
log_ok "Validator 4 started"

echo
log_info "All validators started. Waiting for chain to produce blocks..."
sleep 15

# Verify chain is running
echo
log_info "Verifying chain status..."

HEIGHT=$(ssh aura-testnet "curl -s http://127.0.0.1:10657/status 2>/dev/null | jq -r '.result.sync_info.latest_block_height'" 2>/dev/null || echo "0")
VALIDATORS=$(ssh aura-testnet "curl -s http://127.0.0.1:10657/validators 2>/dev/null | jq '.result.validators | length'" 2>/dev/null || echo "0")

echo
echo "=========================================="
echo "Launch Status"
echo "=========================================="
echo "Block height: $HEIGHT"
echo "Active validators: $VALIDATORS"
echo "=========================================="

if [ "$HEIGHT" -gt "0" ] && [ "$VALIDATORS" -ge "3" ]; then
    log_ok "Chain is producing blocks!"
    echo
    echo "Next steps:"
    echo "1. Monitor chain: ssh aura-testnet 'tail -f ~/.aura-val1/node.log'"
    echo "2. Run integration tests: ./scripts/mvp-integration-tests/run-all-tests.sh"
    echo "3. Update infrastructure (explorer, faucet, etc.)"
else
    log_warn "Chain may not be running correctly"
    echo "Check validator logs:"
    echo "  ssh aura-testnet 'tail -50 ~/.aura-val1/node.log'"
    echo "  ssh services-testnet 'tail -50 ~/.aura-val3/node.log'"
fi
