#!/bin/bash
# Test script to verify the governance delegation key construction fix

set -e

echo "=================================="
echo "Governance Delegation Test Suite"
echo "=================================="
echo ""

cd "$(dirname "$0")"

echo "Running TestMultipleDelegations (the failing test)..."
go test -v -run TestMultipleDelegations ./x/governance/keeper/

echo ""
echo "Running all voting power tests..."
go test -v -run "TestVotingPower|TestSybil|TestWhale" ./x/governance/keeper/

echo ""
echo "Running all governance keeper tests..."
go test -v ./x/governance/keeper/

echo ""
echo "=================================="
echo "All tests passed!"
echo "=================================="
