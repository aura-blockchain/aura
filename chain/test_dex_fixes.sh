#!/bin/bash

# Test script to verify DEX test fixes
# Tests the four previously failing tests

set -e

cd /home/decri/blockchain-projects/aura/chain

echo "Testing DEX keeper fixes..."
echo ""

# Test 1: Pool creation limit enforcement
echo "1. Testing pool creation limit enforcement..."
go test -v ./x/dex/keeper/... -run TestPoolCreationLimit_Enforcement -count=1

# Test 2: Pool creation cooldown with proper time advancement
echo ""
echo "2. Testing pool creation cooldown respects cooldown period..."
go test -v ./x/dex/keeper/... -run TestPoolCreationCooldown_RespectsCooldownPeriod -count=1

# Test 3: Query market price with funded account
echo ""
echo "3. Testing query market price uses stored value..."
go test -v ./x/dex/keeper/... -run TestQueryMarketPriceUsesStoredValue -count=1

# Test 4: Query spot price with funded account
echo ""
echo "4. Testing query spot price..."
go test -v ./x/dex/keeper/... -run TestQuerySpotPrice -count=1

echo ""
echo "All tests passed successfully!"
