#!/bin/bash
# Aura Project - Run Go Tests (Bash version)
# This script runs Go tests from the correct directory

set -e

echo -e "\033[36mAura Project - Running Go Tests\033[0m"
echo -e "\033[36m================================\n\033[0m"

# Change to the chain directory (where go.mod is located)
cd "/c/Users/decri/GitClones/aura/chain"

echo -e "\033[33mWorking directory: $PWD\033[0m"
echo -e "\n\033[36mRunning go test...\033[0m"

# Run go test with any additional arguments
go test ./... "$@"

echo -e "\n\033[32mDone!\033[0m"
