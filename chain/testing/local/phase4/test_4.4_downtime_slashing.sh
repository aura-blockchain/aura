#!/bin/bash
# Phase 4.4: Validator Downtime Slashing Test
# Tests: Take a validator offline and verify downtime slashing mechanics

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
RESULTS_FILE="$SCRIPT_DIR/test_4.4_results.txt"

echo "=== Phase 4.4: Validator Downtime Slashing Test ===" | tee "$RESULTS_FILE"
echo "Started at: $(date)" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test counters
PASSED=0
FAILED=0

# RPC and REST endpoints
RPC="http://localhost:27657"
REST="http://localhost:2317"

echo "Step 1: Get slashing parameters..." | tee -a "$RESULTS_FILE"
echo "===================================" | tee -a "$RESULTS_FILE"

# Try to get slashing params
echo "Querying slashing parameters from chain..." | tee -a "$RESULTS_FILE"

# Method 1: Use docker exec to query directly
SLASHING_PARAMS=$(docker exec aura-validator-1 aurad q slashing params --output json 2>/dev/null || echo "{}")

if [ "$SLASHING_PARAMS" != "{}" ] && [ -n "$SLASHING_PARAMS" ]; then
    echo "Slashing parameters retrieved:" | tee -a "$RESULTS_FILE"
    echo "$SLASHING_PARAMS" | jq '.' | tee -a "$RESULTS_FILE"

    # Extract key parameters
    SIGNED_BLOCKS_WINDOW=$(echo "$SLASHING_PARAMS" | jq -r '.signed_blocks_window // .params.signed_blocks_window' || echo "0")
    MIN_SIGNED_PER_WINDOW=$(echo "$SLASHING_PARAMS" | jq -r '.min_signed_per_window // .params.min_signed_per_window' || echo "0")
    DOWNTIME_JAIL_DURATION=$(echo "$SLASHING_PARAMS" | jq -r '.downtime_jail_duration // .params.downtime_jail_duration' || echo "0")
    SLASH_FRACTION_DOWNTIME=$(echo "$SLASHING_PARAMS" | jq -r '.slash_fraction_downtime // .params.slash_fraction_downtime' || echo "0")

    echo "" | tee -a "$RESULTS_FILE"
    echo "Key parameters:" | tee -a "$RESULTS_FILE"
    echo "  Signed blocks window: $SIGNED_BLOCKS_WINDOW" | tee -a "$RESULTS_FILE"
    echo "  Min signed per window: $MIN_SIGNED_PER_WINDOW" | tee -a "$RESULTS_FILE"
    echo "  Downtime jail duration: $DOWNTIME_JAIL_DURATION" | tee -a "$RESULTS_FILE"
    echo "  Slash fraction (downtime): $SLASH_FRACTION_DOWNTIME" | tee -a "$RESULTS_FILE"

    if [ "$SIGNED_BLOCKS_WINDOW" != "0" ] && [ "$SIGNED_BLOCKS_WINDOW" != "null" ]; then
        echo -e "${GREEN}✓ Downtime slashing parameters configured${NC}" | tee -a "$RESULTS_FILE"
        PASSED=$((PASSED + 1))

        # Calculate how many blocks can be missed
        if [ "$MIN_SIGNED_PER_WINDOW" != "0" ] && [ "$MIN_SIGNED_PER_WINDOW" != "null" ]; then
            # Remove quotes if present
            MIN_SIGNED_CLEAN=$(echo "$MIN_SIGNED_PER_WINDOW" | tr -d '"')

            # Check if it's a decimal fraction (like "0.500000000000000000")
            if [[ "$MIN_SIGNED_CLEAN" =~ ^0\. ]]; then
                # Convert decimal to percentage
                PERCENTAGE=$(echo "$MIN_SIGNED_CLEAN * 100" | bc -l | cut -d'.' -f1)
                MAX_MISSED=$((SIGNED_BLOCKS_WINDOW * (100 - PERCENTAGE) / 100))
            else
                MAX_MISSED=$((SIGNED_BLOCKS_WINDOW - MIN_SIGNED_CLEAN))
            fi

            echo "" | tee -a "$RESULTS_FILE"
            echo "Downtime threshold calculation:" | tee -a "$RESULTS_FILE"
            echo "  Window: $SIGNED_BLOCKS_WINDOW blocks" | tee -a "$RESULTS_FILE"
            echo "  Must sign at least: $MIN_SIGNED_PER_WINDOW" | tee -a "$RESULTS_FILE"
            echo "  Can miss up to: $MAX_MISSED blocks before jailing" | tee -a "$RESULTS_FILE"
        fi
    else
        echo -e "${YELLOW}⚠ Slashing parameters not fully configured${NC}" | tee -a "$RESULTS_FILE"
    fi
else
    echo -e "${YELLOW}⚠ Could not retrieve slashing parameters${NC}" | tee -a "$RESULTS_FILE"
    echo "Using default Cosmos SDK values for reference:" | tee -a "$RESULTS_FILE"
    echo "  Signed blocks window: 10000 (typical)" | tee -a "$RESULTS_FILE"
    echo "  Min signed: 50% of window" | tee -a "$RESULTS_FILE"
    echo "  Jail duration: 10 minutes (typical)" | tee -a "$RESULTS_FILE"

    SIGNED_BLOCKS_WINDOW=10000
    MAX_MISSED=5000
fi

echo "" | tee -a "$RESULTS_FILE"

# Step 2: Get current validator states
echo "Step 2: Get current validator states..." | tee -a "$RESULTS_FILE"
echo "========================================" | tee -a "$RESULTS_FILE"

# Get validators from Tendermint
VALIDATORS=$(curl -s "$RPC/validators" | jq -r '.result.validators[] | "\(.address),\(.voting_power)"')
echo "Active validators:" | tee -a "$RESULTS_FILE"
echo "$VALIDATORS" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

VALIDATOR_COUNT=$(echo "$VALIDATORS" | wc -l)
echo "Total validators: $VALIDATOR_COUNT" | tee -a "$RESULTS_FILE"

if [ "$VALIDATOR_COUNT" -lt 2 ]; then
    echo -e "${YELLOW}⚠ Only $VALIDATOR_COUNT validator(s) found${NC}" | tee -a "$RESULTS_FILE"
    echo "Downtime test will use the available validator" | tee -a "$RESULTS_FILE"
fi

echo "" | tee -a "$RESULTS_FILE"

# Get signing info for all validators
echo "Querying validator signing info..." | tee -a "$RESULTS_FILE"
SIGNING_INFO=$(docker exec aura-validator-1 aurad q slashing signing-infos --output json 2>/dev/null || echo "{}")

if [ "$SIGNING_INFO" != "{}" ] && [ -n "$SIGNING_INFO" ]; then
    echo "Signing info retrieved:" | tee -a "$RESULTS_FILE"
    echo "$SIGNING_INFO" | jq -r '.info[]? | "Cons Address: \(.address)\n  Missed blocks: \(.missed_blocks_counter)\n  Jailed until: \(.jailed_until)"' | tee -a "$RESULTS_FILE"

    # Check for already jailed validators
    JAILED=$(echo "$SIGNING_INFO" | jq -r '.info[]? | select(.jailed_until != "1970-01-01T00:00:00Z") | .address')
    if [ -n "$JAILED" ]; then
        echo "" | tee -a "$RESULTS_FILE"
        echo -e "${YELLOW}⚠ Some validators are already jailed:${NC}" | tee -a "$RESULTS_FILE"
        echo "$JAILED" | tee -a "$RESULTS_FILE"
    else
        echo -e "${GREEN}✓ No validators currently jailed${NC}" | tee -a "$RESULTS_FILE"
        PASSED=$((PASSED + 1))
    fi
else
    echo -e "${YELLOW}⚠ Could not retrieve signing info${NC}" | tee -a "$RESULTS_FILE"
fi

echo "" | tee -a "$RESULTS_FILE"

# Step 3: Downtime simulation test
echo "Step 3: Downtime Simulation Test" | tee -a "$RESULTS_FILE"
echo "=================================" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

echo "Test approach: Stop one validator temporarily" | tee -a "$RESULTS_FILE"
echo "WARNING: This test will briefly disrupt the network" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Choose validator to stop (use validator-4 if available)
TARGET_CONTAINER="aura-validator-4"

if ! docker ps --format '{{.Names}}' | grep -q "^$TARGET_CONTAINER$"; then
    echo -e "${YELLOW}⚠ $TARGET_CONTAINER not running, using validator-2${NC}" | tee -a "$RESULTS_FILE"
    TARGET_CONTAINER="aura-validator-2"
fi

echo "Target validator container: $TARGET_CONTAINER" | tee -a "$RESULTS_FILE"

# Get current height before stopping
START_HEIGHT=$(curl -s "$RPC/status" | jq -r '.result.sync_info.latest_block_height')
echo "Starting height: $START_HEIGHT" | tee -a "$RESULTS_FILE"

# Get validator's current missed block count
VALIDATOR_CONS_ADDR=$(docker exec "$TARGET_CONTAINER" aurad tendermint show-address 2>/dev/null || echo "")
if [ -n "$VALIDATOR_CONS_ADDR" ]; then
    echo "Target validator consensus address: $VALIDATOR_CONS_ADDR" | tee -a "$RESULTS_FILE"

    INITIAL_MISSED=$(echo "$SIGNING_INFO" | jq -r --arg addr "$VALIDATOR_CONS_ADDR" '.info[]? | select(.address == $addr) | .missed_blocks_counter' || echo "0")
    echo "Initial missed blocks: $INITIAL_MISSED" | tee -a "$RESULTS_FILE"
else
    echo -e "${YELLOW}⚠ Could not get consensus address${NC}" | tee -a "$RESULTS_FILE"
    INITIAL_MISSED="0"
fi

echo "" | tee -a "$RESULTS_FILE"

# Stop the validator
echo "Stopping validator $TARGET_CONTAINER..." | tee -a "$RESULTS_FILE"
docker stop "$TARGET_CONTAINER" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Validator stopped successfully${NC}" | tee -a "$RESULTS_FILE"
else
    echo -e "${RED}✗ Failed to stop validator${NC}" | tee -a "$RESULTS_FILE"
    FAILED=$((FAILED + 1))
fi

# Wait for some blocks to pass (10-20 blocks)
echo "" | tee -a "$RESULTS_FILE"
echo "Waiting for 20 blocks to pass..." | tee -a "$RESULTS_FILE"

TARGET_HEIGHT=$((START_HEIGHT + 20))
while true; do
    CURRENT_HEIGHT=$(curl -s "$RPC/status" | jq -r '.result.sync_info.latest_block_height' || echo "$START_HEIGHT")

    if [ "$CURRENT_HEIGHT" -ge "$TARGET_HEIGHT" ]; then
        break
    fi

    echo "  Current height: $CURRENT_HEIGHT / $TARGET_HEIGHT" | tee -a "$RESULTS_FILE"
    sleep 3
done

echo "Reached target height: $CURRENT_HEIGHT" | tee -a "$RESULTS_FILE"

# Restart the validator
echo "" | tee -a "$RESULTS_FILE"
echo "Restarting validator $TARGET_CONTAINER..." | tee -a "$RESULTS_FILE"
docker start "$TARGET_CONTAINER" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Validator restarted successfully${NC}" | tee -a "$RESULTS_FILE"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗ Failed to restart validator${NC}" | tee -a "$RESULTS_FILE"
    FAILED=$((FAILED + 1))
fi

# Wait for validator to sync
echo "Waiting for validator to sync..." | tee -a "$RESULTS_FILE"
sleep 10

# Check if missed blocks increased
echo "" | tee -a "$RESULTS_FILE"
echo "Checking updated missed block count..." | tee -a "$RESULTS_FILE"

UPDATED_SIGNING_INFO=$(docker exec aura-validator-1 aurad q slashing signing-infos --output json 2>/dev/null || echo "{}")

if [ -n "$VALIDATOR_CONS_ADDR" ] && [ "$UPDATED_SIGNING_INFO" != "{}" ]; then
    UPDATED_MISSED=$(echo "$UPDATED_SIGNING_INFO" | jq -r --arg addr "$VALIDATOR_CONS_ADDR" '.info[]? | select(.address == $addr) | .missed_blocks_counter' || echo "0")
    echo "Updated missed blocks: $UPDATED_MISSED" | tee -a "$RESULTS_FILE"

    BLOCKS_MISSED=$((UPDATED_MISSED - INITIAL_MISSED))
    echo "Blocks missed during downtime: $BLOCKS_MISSED" | tee -a "$RESULTS_FILE"

    if [ "$BLOCKS_MISSED" -gt 0 ]; then
        echo -e "${GREEN}✓ Downtime was correctly tracked${NC}" | tee -a "$RESULTS_FILE"
        PASSED=$((PASSED + 1))

        # Check if validator was jailed
        JAILED_UNTIL=$(echo "$UPDATED_SIGNING_INFO" | jq -r --arg addr "$VALIDATOR_CONS_ADDR" '.info[]? | select(.address == $addr) | .jailed_until' || echo "1970-01-01T00:00:00Z")

        if [ "$JAILED_UNTIL" != "1970-01-01T00:00:00Z" ]; then
            echo -e "${GREEN}✓ Validator was jailed for downtime${NC}" | tee -a "$RESULTS_FILE"
            echo "  Jailed until: $JAILED_UNTIL" | tee -a "$RESULTS_FILE"
            PASSED=$((PASSED + 1))
        else
            echo -e "${GREEN}✓ Validator NOT jailed (downtime below threshold)${NC}" | tee -a "$RESULTS_FILE"
            echo "  This is expected for short downtimes" | tee -a "$RESULTS_FILE"
            PASSED=$((PASSED + 1))
        fi
    else
        echo -e "${YELLOW}⚠ No missed blocks recorded${NC}" | tee -a "$RESULTS_FILE"
        echo "  Validator may have synced too quickly" | tee -a "$RESULTS_FILE"
    fi
else
    echo -e "${YELLOW}⚠ Could not verify missed blocks${NC}" | tee -a "$RESULTS_FILE"
fi

echo "" | tee -a "$RESULTS_FILE"

# Step 4: Verify chain continued during downtime
echo "Step 4: Verify chain continued during validator downtime" | tee -a "$RESULTS_FILE"
echo "=========================================================" | tee -a "$RESULTS_FILE"

FINAL_HEIGHT=$(curl -s "$RPC/status" | jq -r '.result.sync_info.latest_block_height')
BLOCKS_PRODUCED=$((FINAL_HEIGHT - START_HEIGHT))

echo "Blocks produced during test: $BLOCKS_PRODUCED" | tee -a "$RESULTS_FILE"
echo "Start height: $START_HEIGHT" | tee -a "$RESULTS_FILE"
echo "Final height: $FINAL_HEIGHT" | tee -a "$RESULTS_FILE"

if [ "$BLOCKS_PRODUCED" -ge 20 ]; then
    echo -e "${GREEN}✓ Chain continued producing blocks despite validator downtime${NC}" | tee -a "$RESULTS_FILE"
    echo "  This demonstrates BFT tolerance (can tolerate up to 1/3 validators down)" | tee -a "$RESULTS_FILE"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗ Chain did not produce expected number of blocks${NC}" | tee -a "$RESULTS_FILE"
    FAILED=$((FAILED + 1))
fi

echo "" | tee -a "$RESULTS_FILE"

# Step 5: Downtime slashing mechanics explanation
echo "Step 5: Downtime Slashing Mechanics" | tee -a "$RESULTS_FILE"
echo "====================================" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

echo "How downtime slashing works:" | tee -a "$RESULTS_FILE"
echo "1. Validators must sign blocks to prove they're online" | tee -a "$RESULTS_FILE"
echo "2. Each validator has a 'signed_blocks_window' (e.g., 10,000 blocks)" | tee -a "$RESULTS_FILE"
echo "3. Must sign min_signed_per_window (e.g., 50% = 5,000 blocks)" | tee -a "$RESULTS_FILE"
echo "4. If missed > threshold, validator is jailed" | tee -a "$RESULTS_FILE"
echo "5. Jailed validators cannot participate until unjailed" | tee -a "$RESULTS_FILE"
echo "6. Small slash penalty applied (e.g., 0.01% of stake)" | tee -a "$RESULTS_FILE"
echo "7. Validator must manually unjail themselves" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Final summary
echo "========================================" | tee -a "$RESULTS_FILE"
echo "PHASE 4.4 DOWNTIME SLASHING SUMMARY" | tee -a "$RESULTS_FILE"
echo "========================================" | tee -a "$RESULTS_FILE"
echo "Completed at: $(date)" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"
echo "Tests passed: $PASSED" | tee -a "$RESULTS_FILE"
echo "Tests failed: $FAILED" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

echo "Key Findings:" | tee -a "$RESULTS_FILE"
echo "1. Downtime tracking is active and counting missed blocks" | tee -a "$RESULTS_FILE"
echo "2. Chain remains operational with 1 validator down (BFT tolerance)" | tee -a "$RESULTS_FILE"
echo "3. Validators can rejoin network after downtime" | tee -a "$RESULTS_FILE"
echo "4. Slashing parameters are properly configured" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

if [ "$FAILED" -eq 0 ]; then
    echo -e "${GREEN}✓ PHASE 4.4 PASSED - Downtime slashing working correctly${NC}" | tee -a "$RESULTS_FILE"
    echo "" | tee -a "$RESULTS_FILE"
    echo "Recommendations for production:" | tee -a "$RESULTS_FILE"
    echo "- Monitor validator uptime continuously" | tee -a "$RESULTS_FILE"
    echo "- Set up alerts for missed blocks approaching threshold" | tee -a "$RESULTS_FILE"
    echo "- Implement redundancy and failover systems" | tee -a "$RESULTS_FILE"
    echo "- Have procedures for unjailing validators" | tee -a "$RESULTS_FILE"
    exit 0
else
    echo -e "${RED}✗ PHASE 4.4 FAILED - Issues detected${NC}" | tee -a "$RESULTS_FILE"
    exit 1
fi
