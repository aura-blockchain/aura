#!/bin/bash
# Test Helpers for Phase 5
# Common functions used across all Phase 5 tests

# Get block height from RPC endpoint (host machine)
get_height_from_host_rpc() {
    local PORT=${1:-27657}  # Default to validator-1's RPC port
    curl -s localhost:${PORT}/status 2>/dev/null | jq -r '.result.sync_info.latest_block_height // "0"' 2>/dev/null || echo "0"
}

# Get block height using aurad query inside container
get_height_from_container() {
    local CONTAINER=${1:-aura-validator-1}
    docker exec ${CONTAINER} aurad q block --output json 2>&1 | jq -r '.block.header.height // "0"' 2>/dev/null || echo "0"
}

# Export this for sed replacement
export -f get_height_from_host_rpc
export -f get_height_from_container
