#!/bin/bash
# Simple helper to get current block height from any validator
# Usage: ./get_block_height.sh [rpc_port]
# Default port: 27657 (validator-1)

PORT=${1:-27657}
curl -s localhost:${PORT}/status 2>/dev/null | jq -r '.result.sync_info.latest_block_height // "0"' 2>/dev/null || echo "0"
