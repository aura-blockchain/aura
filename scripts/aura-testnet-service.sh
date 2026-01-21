#!/bin/bash
# Aura Testnet Service Script
# Manages the local Aura testnet for development

set -e

AURA_HOME="$HOME/.aura-testnet"
CHAIN_ID="aura-mvp-1"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED="\033[0;31m"
GREEN="\033[0;32m"
YELLOW="\033[0;33m"
NC="\033[0m"

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

check_prerequisites() {
    # Check if aurad binary or docker image exists
    if command -v aurad &> /dev/null; then
        AURAD_CMD="aurad"
        log_info "Using local aurad binary"
    elif docker images aurad:latest --format "{{.Repository}}" 2>/dev/null | grep -q aurad; then
        AURAD_CMD="docker run --rm -v $AURA_HOME:/home/aura/.aura aurad:latest aurad"
        log_info "Using aurad Docker image"
    else
        log_warn "aurad not found (neither binary nor Docker image)"
        log_warn "The Aura project needs proto generation before it can build."
        log_warn "Run 'cd ~/blockchain-projects/aura/chain && make proto-gen' first."
        log_warn "Then run 'docker build -t aurad:latest -f Dockerfile .' from aura directory"
        return 1
    fi
    return 0
}

init_chain() {
    if \! check_prerequisites; then
        return 1
    fi

    if [ -d "$AURA_HOME/config" ]; then
        log_warn "Chain already initialized at $AURA_HOME"
        log_warn "Run '$0 reset' to reset the chain"
        return 0
    fi

    log_info "Initializing Aura testnet..."
    mkdir -p "$AURA_HOME"

    $AURAD_CMD init testnet-node --chain-id $CHAIN_ID --home "$AURA_HOME"

    # Configure for local development
    sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0stake"/' "$AURA_HOME/config/app.toml"
    sed -i 's/enable = false/enable = true/' "$AURA_HOME/config/app.toml"
    sed -i 's/swagger = false/swagger = true/' "$AURA_HOME/config/app.toml"

    # Add test account
    echo "test test test test test test test test test test test junk" | $AURAD_CMD keys add validator --recover --keyring-backend test --home "$AURA_HOME"

    $AURAD_CMD genesis add-genesis-account validator 100000000000stake --keyring-backend test --home "$AURA_HOME"
    $AURAD_CMD genesis gentx validator 1000000stake --chain-id $CHAIN_ID --keyring-backend test --home "$AURA_HOME"
    $AURAD_CMD genesis collect-gentxs --home "$AURA_HOME"

    log_info "Aura testnet initialized at $AURA_HOME"
}

start() {
    if \! check_prerequisites; then
        log_error "Cannot start: prerequisites not met"
        return 1
    fi

    if [ \! -d "$AURA_HOME/config" ]; then
        log_info "Chain not initialized, initializing now..."
        init_chain
    fi

    if pgrep -f "aurad.*start.*$CHAIN_ID" > /dev/null; then
        log_warn "Aura testnet is already running"
        return 0
    fi

    log_info "Starting Aura testnet..."
    nohup $AURAD_CMD start --home "$AURA_HOME" > "$AURA_HOME/node.log" 2>&1 &
    echo $\! > "$AURA_HOME/node.pid"

    sleep 3
    if pgrep -f "aurad.*start.*$CHAIN_ID" > /dev/null; then
        log_info "Aura testnet started (PID: $(cat $AURA_HOME/node.pid))"
    else
        log_error "Failed to start Aura testnet"
        return 1
    fi
}

stop() {
    if [ -f "$AURA_HOME/node.pid" ]; then
        PID=$(cat "$AURA_HOME/node.pid")
        if kill -0 "$PID" 2>/dev/null; then
            log_info "Stopping Aura testnet (PID: $PID)..."
            kill "$PID"
            rm -f "$AURA_HOME/node.pid"
            log_info "Aura testnet stopped"
        else
            log_warn "Process $PID not running"
            rm -f "$AURA_HOME/node.pid"
        fi
    else
        # Try to find and kill any running aurad
        pkill -f "aurad.*start.*$CHAIN_ID" 2>/dev/null || true
        log_info "Aura testnet stopped"
    fi
}

status() {
    if \! check_prerequisites 2>/dev/null; then
        log_warn "Status: aurad not installed (build required)"
        return 1
    fi

    if pgrep -f "aurad.*start.*$CHAIN_ID" > /dev/null; then
        log_info "Status: Aura testnet is RUNNING"
        if [ -f "$AURA_HOME/node.pid" ]; then
            log_info "PID: $(cat $AURA_HOME/node.pid)"
        fi
        return 0
    else
        log_warn "Status: Aura testnet is STOPPED"
        return 1
    fi
}

case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        stop
        sleep 2
        start
        ;;
    status)
        status
        ;;
    init)
        init_chain
        ;;
    reset)
        stop
        rm -rf "$AURA_HOME"
        log_info "Chain data cleared"
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|init|reset}"
        exit 1
        ;;
esac
