#!/bin/bash
# Quick 4-validator local testnet - no Docker, runs directly
# Usage: ./scripts/local-4-validators.sh [start|stop|status|clean]

set -e

CHAIN_ID="aura-local-4val"
BASE_DIR="/tmp/aura-testnet"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
AURAD="$SCRIPT_DIR/../aurad"
DENOM="uaura"

# Ports: each validator gets unique ports
# Note: gRPC uses 19090+ range to avoid conflict with Prometheus (9091)
declare -a P2P_PORTS=(26656 26666 26676 26686)
declare -a RPC_PORTS=(26657 26667 26677 26687)
declare -a API_PORTS=(1317 1318 1319 1320)
declare -a GRPC_PORTS=(19090 19091 19092 19093)

log() { echo "[$(date +%H:%M:%S)] $1"; }

init_validators() {
    log "Initializing 4 validators..."
    rm -rf "$BASE_DIR"

    # Step 1: Initialize all nodes (each creates its own validator in genesis via priv_validator_key)
    for i in 0 1 2 3; do
        HOME_DIR="$BASE_DIR/val$i"
        mkdir -p "$HOME_DIR"
        log "  Initializing val$i..."
        $AURAD init "validator-$i" --chain-id "$CHAIN_ID" --home "$HOME_DIR" -y > /dev/null 2>&1
        # Wait for config.toml to be complete (Cosmos SDK v0.50+ uses ~105 lines)
        while [ ! -f "$HOME_DIR/config/config.toml" ] || [ "$(wc -l < "$HOME_DIR/config/config.toml")" -lt 100 ]; do
            sleep 0.5
        done
    done

    # Step 2: Merge all validators into a single genesis
    # Our custom init creates validators directly in genesis using priv_validator_key
    # We need to merge validators, delegations, last_validator_powers, and accounts from all 4
    log "Merging validators into single genesis..."

    # Extract data from all 4 genesis files and merge
    jq -s '
        # Merge validators array
        .[0].app_state.staking.validators = [.[0,1,2,3].app_state.staking.validators[]] |
        # Merge delegations array
        .[0].app_state.staking.delegations = [.[0,1,2,3].app_state.staking.delegations[]] |
        # Merge last_validator_powers array
        .[0].app_state.staking.last_validator_powers = [.[0,1,2,3].app_state.staking.last_validator_powers[]] |
        # Sum last_total_power (each validator has same power, so multiply by 4)
        .[0].app_state.staking.last_total_power = (.[0].app_state.staking.last_total_power | tonumber | . * 4 | tostring) |
        # Merge bank balances
        .[0].app_state.bank.balances = [.[0,1,2,3].app_state.bank.balances[]] |
        # Merge auth accounts
        .[0].app_state.auth.accounts = [.[0,1,2,3].app_state.auth.accounts[]] |
        # Return the merged first genesis
        .[0]
    ' "$BASE_DIR/val0/config/genesis.json" \
      "$BASE_DIR/val1/config/genesis.json" \
      "$BASE_DIR/val2/config/genesis.json" \
      "$BASE_DIR/val3/config/genesis.json" > "$BASE_DIR/merged_genesis.json"

    # Step 3: Distribute merged genesis to all validators
    log "Distributing merged genesis..."
    for i in 0 1 2 3; do
        cp "$BASE_DIR/merged_genesis.json" "$BASE_DIR/val$i/config/genesis.json"
    done
    rm "$BASE_DIR/merged_genesis.json"

    # Step 4: Get node IDs using aurad command
    log "Getting node IDs..."
    declare -a NODE_IDS
    for i in 0 1 2 3; do
        HOME_DIR="$BASE_DIR/val$i"
        NODE_ID=$($AURAD tendermint show-node-id --home "$HOME_DIR" 2>/dev/null)
        NODE_IDS[$i]="$NODE_ID"
        log "  Val$i node ID: $NODE_ID"
    done

    # Build peers string
    log "Configuring peers..."
    PEERS=""
    for i in 0 1 2 3; do
        if [ -n "$PEERS" ]; then
            PEERS="$PEERS,"
        fi
        PEERS="${PEERS}${NODE_IDS[$i]}@127.0.0.1:${P2P_PORTS[$i]}"
    done

    # Step 5: Configure each validator's ports and peers
    for i in 0 1 2 3; do
        HOME_DIR="$BASE_DIR/val$i"
        CONFIG="$HOME_DIR/config/config.toml"
        APP_CONFIG="$HOME_DIR/config/app.toml"

        # Set unique ports
        sed -i "s/laddr = \"tcp:\/\/0.0.0.0:26656\"/laddr = \"tcp:\/\/0.0.0.0:${P2P_PORTS[$i]}\"/" "$CONFIG"
        sed -i "s/laddr = \"tcp:\/\/127.0.0.1:26657\"/laddr = \"tcp:\/\/127.0.0.1:${RPC_PORTS[$i]}\"/" "$CONFIG"
        sed -i "s/address = \"tcp:\/\/localhost:1317\"/address = \"tcp:\/\/localhost:${API_PORTS[$i]}\"/" "$APP_CONFIG"
        sed -i "s/address = \"localhost:9090\"/address = \"localhost:${GRPC_PORTS[$i]}\"/" "$APP_CONFIG"

        # Set persistent peers (exclude self using NODE_IDS array)
        SELF_ID="${NODE_IDS[$i]}"
        OTHER_PEERS=$(echo "$PEERS" | sed "s/${SELF_ID}@127.0.0.1:${P2P_PORTS[$i]},\\?//g" | sed 's/^,//' | sed 's/,$//')
        sed -i "s/persistent_peers = \"\"/persistent_peers = \"$OTHER_PEERS\"/" "$CONFIG"

        # Enable API
        sed -i 's/enable = false/enable = true/' "$APP_CONFIG"

        # Fast block time for testing (1s)
        sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/' "$CONFIG"
    done

    log "Initialization complete!"
}

start_validators() {
    log "Starting 4 validators..."

    for i in 0 1 2 3; do
        HOME_DIR="$BASE_DIR/val$i"
        LOG_FILE="$BASE_DIR/val$i.log"
        PID_FILE="$BASE_DIR/val$i.pid"

        if [ -f "$PID_FILE" ] && kill -0 $(cat "$PID_FILE") 2>/dev/null; then
            log "  Val$i already running (PID $(cat $PID_FILE))"
            continue
        fi

        $AURAD start --home "$HOME_DIR" > "$LOG_FILE" 2>&1 &
        echo $! > "$PID_FILE"
        log "  Val$i started (PID $!, RPC: ${RPC_PORTS[$i]})"
    done

    log "Waiting for blocks..."
    sleep 5

    # Check if producing blocks
    HEIGHT=$(curl -s "http://127.0.0.1:${RPC_PORTS[0]}/status" | grep -o '"latest_block_height":"[0-9]*"' | grep -o '[0-9]*' || echo "0")
    log "Current block height: $HEIGHT"
}

stop_validators() {
    log "Stopping validators..."

    for i in 0 1 2 3; do
        PID_FILE="$BASE_DIR/val$i.pid"
        if [ -f "$PID_FILE" ]; then
            PID=$(cat "$PID_FILE")
            if kill -0 "$PID" 2>/dev/null; then
                kill "$PID"
                log "  Val$i stopped (PID $PID)"
            fi
            rm "$PID_FILE"
        fi
    done
}

status_validators() {
    log "Validator status:"

    for i in 0 1 2 3; do
        PID_FILE="$BASE_DIR/val$i.pid"
        RPC="http://127.0.0.1:${RPC_PORTS[$i]}"

        if [ -f "$PID_FILE" ] && kill -0 $(cat "$PID_FILE") 2>/dev/null; then
            HEIGHT=$(curl -s "$RPC/status" 2>/dev/null | grep -o '"latest_block_height":"[0-9]*"' | grep -o '[0-9]*' || echo "?")
            CATCHING=$(curl -s "$RPC/status" 2>/dev/null | grep -o '"catching_up":[a-z]*' | cut -d: -f2 || echo "?")
            log "  Val$i: RUNNING (height: $HEIGHT, catching_up: $CATCHING)"
        else
            log "  Val$i: STOPPED"
        fi
    done
}

clean_testnet() {
    log "Cleaning testnet data..."
    stop_validators
    rm -rf "$BASE_DIR"
    log "Cleaned!"
}

case "${1:-}" in
    start)
        if [ ! -d "$BASE_DIR/val0" ]; then
            init_validators
        fi
        start_validators
        ;;
    stop)
        stop_validators
        ;;
    status)
        status_validators
        ;;
    clean)
        clean_testnet
        ;;
    restart)
        stop_validators
        sleep 2
        start_validators
        ;;
    init)
        clean_testnet
        init_validators
        ;;
    *)
        echo "Usage: $0 {start|stop|status|clean|restart|init}"
        echo ""
        echo "Commands:"
        echo "  start   - Initialize (if needed) and start 4 validators"
        echo "  stop    - Stop all validators"
        echo "  status  - Show validator status and block heights"
        echo "  clean   - Stop and remove all testnet data"
        echo "  restart - Stop and start validators"
        echo "  init    - Clean and reinitialize (no start)"
        echo ""
        echo "Ports:"
        echo "  Val0: RPC 26657, API 1317, gRPC 19090, P2P 26656"
        echo "  Val1: RPC 26667, API 1318, gRPC 19091, P2P 26666"
        echo "  Val2: RPC 26677, API 1319, gRPC 19092, P2P 26676"
        echo "  Val3: RPC 26687, API 1320, gRPC 19093, P2P 26686"
        echo ""
        echo "Logs: $BASE_DIR/val*.log"
        ;;
esac
