#!/bin/bash
# Source this file to set up isolated environment for aura
# Usage: source env.sh

export PROJECT_NAME="aura"
export PROJECT_ROOT="/home/hudson/blockchain-projects/aura"

# Go setup
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go

# Go caches - isolated per project
export GOCACHE="$PROJECT_ROOT/.cache/go-build"
export GOTESTCACHE="$PROJECT_ROOT/.cache/go-test"

# Log directory
export LOG_DIR="$PROJECT_ROOT/logs"

# Docker network
export DOCKER_NETWORK="aura-testnet"

# Convenience aliases
alias logs="cd $LOG_DIR"
alias proj="cd $PROJECT_ROOT"
alias dc="docker compose -f $PROJECT_ROOT/docker/docker-compose.yml"

# Silent environment setup - no output
