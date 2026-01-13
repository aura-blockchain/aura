#!/bin/bash
#
# Aura Testnet BFT Reconfiguration Script
# Converts single-validator testnet to 4-validator BFT setup
#
# This script creates a proper BFT testnet with 4 validators @ 25% voting power each
#

set -e

CHAIN_ID="aura-mvp-1"
DENOM="uaura"
TOTAL_SUPPLY="1000000000000000"  # 1 billion tokens
VALIDATOR_STAKE="250000000000"    # 250k tokens per validator (25% each)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if aurad is available
if ! command -v aurad &> /dev/null; then
    log_error "aurad not found. Please build it first:"
    log_error "cd chain && go build -o aurad ./cmd/aurad"
    exit 1
fi

log_info "=== Aura BFT Testnet Reconfiguration ==="
log_info "This will create a 4-validator testnet with equal voting power"
log_info ""

# Confirm action
read -p "This will DESTROY existing testnet data. Continue? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    log_warn "Aborted by user"
    exit 0
fi

# Stop existing testnet
log_info "Stopping existing testnet containers..."
docker-compose -f docker-compose.yml down 2>/dev/null || true

# Clean up old data
log_info "Cleaning up old validator data..."
rm -rf ~/.aura
rm -rf ./data/validator-*
rm -rf ./testnet

# Create fresh directory structure
log_info "Creating directory structure..."
mkdir -p testnet
cd testnet

# Initialize validators
log_info "Initializing 4 validators..."
for i in {1..4}; do
    VALIDATOR_HOME="validator-$i"
    mkdir -p "$VALIDATOR_HOME"

    log_info "Initializing validator-$i..."
    aurad init "validator-$i" --chain-id "$CHAIN_ID" --home "$VALIDATOR_HOME" > /dev/null 2>&1

    # Generate validator key
    log_info "Generating keys for validator-$i..."
    aurad keys add "validator-$i" --keyring-backend test --home "$VALIDATOR_HOME" > "$VALIDATOR_HOME/key_info.txt" 2>&1

    # Get validator address
    VALIDATOR_ADDR=$(aurad keys show "validator-$i" -a --keyring-backend test --home "$VALIDATOR_HOME")
    echo "$VALIDATOR_ADDR" > "$VALIDATOR_HOME/address.txt"

    log_info "Validator-$i address: $VALIDATOR_ADDR"
done

log_info ""
log_info "All validators initialized successfully!"
log_info ""

# Use validator-1's genesis as the base
BASE_GENESIS="validator-1/config/genesis.json"
log_info "Using validator-1 genesis as base template..."

# Add genesis accounts for all validators
log_info "Adding genesis accounts..."
for i in {1..4}; do
    VALIDATOR_HOME="validator-$i"
    VALIDATOR_ADDR=$(cat "$VALIDATOR_HOME/address.txt")

    log_info "Adding account for validator-$i: $VALIDATOR_ADDR"
    aurad add-genesis-account "$VALIDATOR_ADDR" "${TOTAL_SUPPLY}${DENOM}" \
        --keyring-backend test \
        --home "validator-1" > /dev/null 2>&1
done

# Create gentx for each validator
log_info ""
log_info "Creating genesis transactions (gentx) for each validator..."
for i in {1..4}; do
    VALIDATOR_HOME="validator-$i"

    # Copy the base genesis to each validator
    cp validator-1/config/genesis.json "$VALIDATOR_HOME/config/genesis.json"

    log_info "Creating gentx for validator-$i..."
    aurad gentx "validator-$i" "${VALIDATOR_STAKE}${DENOM}" \
        --chain-id "$CHAIN_ID" \
        --keyring-backend test \
        --home "$VALIDATOR_HOME" \
        --moniker "validator-$i" \
        --commission-rate "0.10" \
        --commission-max-rate "0.20" \
        --commission-max-change-rate "0.01" \
        --min-self-delegation "1" > /dev/null 2>&1

    # Copy gentx to validator-1 for collection
    cp "$VALIDATOR_HOME/config/gentx"/*.json "validator-1/config/gentx/"
done

# Collect all gentx files
log_info ""
log_info "Collecting all genesis transactions..."
aurad collect-gentxs --home "validator-1" > /dev/null 2>&1

# Validate final genesis
log_info "Validating genesis file..."
if aurad validate-genesis --home "validator-1" > /dev/null 2>&1; then
    log_info "✓ Genesis file validated successfully!"
else
    log_error "Genesis validation failed!"
    exit 1
fi

# Copy final genesis to all validators
log_info ""
log_info "Distributing final genesis to all validators..."
for i in {2..4}; do
    cp validator-1/config/genesis.json "validator-$i/config/genesis.json"
    log_info "Copied genesis to validator-$i"
done

# Configure persistent peers
log_info ""
log_info "Configuring persistent peers..."

# Get node IDs
declare -a NODE_IDS
for i in {1..4}; do
    NODE_ID=$(aurad tendermint show-node-id --home "validator-$i")
    NODE_IDS[$i]="$NODE_ID"
    log_info "Validator-$i node ID: $NODE_ID"
done

# Set persistent peers for each validator (connect to all others)
for i in {1..4}; do
    PEERS=""
    for j in {1..4}; do
        if [ $i -ne $j ]; then
            # Use container names as hostnames
            PEER="${NODE_IDS[$j]}@aura-validator-$j:26656"
            if [ -z "$PEERS" ]; then
                PEERS="$PEER"
            else
                PEERS="$PEERS,$PEER"
            fi
        fi
    done

    # Update config.toml
    sed -i "s/persistent_peers = \"\"/persistent_peers = \"$PEERS\"/" "validator-$i/config/config.toml"
    log_info "Configured persistent peers for validator-$i"
done

# Update configuration for all validators
log_info ""
log_info "Updating validator configurations..."
for i in {1..4}; do
    # Enable API
    sed -i 's/enable = false/enable = true/' "validator-$i/config/app.toml"

    # Enable unsafe CORS (testnet only)
    sed -i 's/enabled-unsafe-cors = false/enabled-unsafe-cors = true/' "validator-$i/config/app.toml"

    # Set minimum gas price
    sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.025uaura"/' "validator-$i/config/app.toml"

    # Enable Prometheus
    sed -i 's/prometheus = false/prometheus = true/' "validator-$i/config/config.toml"

    log_info "Updated configuration for validator-$i"
done

# Create docker-compose.yml for BFT testnet
log_info ""
log_info "Creating docker-compose.yml for BFT testnet..."

cat > docker-compose.bft.yml <<'COMPOSE_EOF'
version: '3.8'

services:
  validator-1:
    image: aura:latest
    container_name: aura-validator-1-bft
    command: aurad start --home /aura
    ports:
      - "26657:26657"  # RPC
      - "26656:26656"  # P2P
      - "1317:1317"    # API
      - "9090:9090"    # gRPC
    volumes:
      - ./testnet/validator-1:/aura
    networks:
      - aura-bft-network
    restart: unless-stopped

  validator-2:
    image: aura:latest
    container_name: aura-validator-2-bft
    command: aurad start --home /aura
    ports:
      - "26667:26657"
      - "26666:26656"
      - "1327:1317"
      - "9091:9090"
    volumes:
      - ./testnet/validator-2:/aura
    networks:
      - aura-bft-network
    restart: unless-stopped

  validator-3:
    image: aura:latest
    container_name: aura-validator-3-bft
    command: aurad start --home /aura
    ports:
      - "26677:26657"
      - "26676:26656"
      - "1337:1317"
      - "9092:9090"
    volumes:
      - ./testnet/validator-3:/aura
    networks:
      - aura-bft-network
    restart: unless-stopped

  validator-4:
    image: aura:latest
    container_name: aura-validator-4-bft
    command: aurad start --home /aura
    ports:
      - "26687:26657"
      - "26686:26656"
      - "1347:1317"
      - "9093:9090"
    volumes:
      - ./testnet/validator-4:/aura
    networks:
      - aura-bft-network
    restart: unless-stopped

networks:
  aura-bft-network:
    driver: bridge

COMPOSE_EOF

# Create README
cat > README_BFT_TESTNET.md <<'README_EOF'
# Aura BFT Testnet

This testnet has 4 validators with equal voting power (25% each).

## Validator Distribution
- validator-1: 25% voting power (250,000 tokens staked)
- validator-2: 25% voting power (250,000 tokens staked)
- validator-3: 25% voting power (250,000 tokens staked)
- validator-4: 25% voting power (250,000 tokens staked)

## Starting the Testnet

```bash
docker-compose -f docker-compose.bft.yml up -d
```

## Checking Validator Status

```bash
# Check all validators
curl -s http://localhost:26657/validators | jq '.result.validators'

# Should show 4 validators with equal voting power
```

## BFT Testing

### Test 1: 4/4 Validators (100% voting power)
- All validators running
- Expected: Consensus active, blocks produced

### Test 2: 3/4 Validators (75% voting power)
- Stop 1 validator
- Expected: Consensus active (>66.67% threshold met)

```bash
docker stop aura-validator-4-bft
```

### Test 3: 2/4 Validators (50% voting power)
- Stop 2 validators
- Expected: **Consensus HALTED** (<66.67% threshold)

```bash
docker stop aura-validator-3-bft aura-validator-4-bft
```

### Test 4: 1/4 Validator (25% voting power)
- Stop 3 validators
- Expected: **Consensus HALTED**

## Recovering from Halt

Start validators to restore >66.67% voting power:

```bash
docker start aura-validator-3-bft aura-validator-4-bft
```

## Port Mapping

| Validator | RPC | P2P | API | gRPC |
|-----------|-----|-----|-----|------|
| validator-1 | 26657 | 26656 | 1317 | 9090 |
| validator-2 | 26667 | 26666 | 1327 | 9091 |
| validator-3 | 26677 | 26676 | 1337 | 9092 |
| validator-4 | 26687 | 26686 | 1347 | 9093 |

README_EOF

log_info ""
log_info "========================================="
log_info "✓ BFT Testnet Configuration Complete!"
log_info "========================================="
log_info ""
log_info "Testnet location: ./testnet/"
log_info "Docker compose file: docker-compose.bft.yml"
log_info "README: README_BFT_TESTNET.md"
log_info ""
log_info "Validator Summary:"
for i in {1..4}; do
    ADDR=$(cat "validator-$i/address.txt")
    log_info "  validator-$i: $ADDR"
done
log_info ""
log_info "Next steps:"
log_info "1. Build Docker image: docker build -t aura:latest ."
log_info "2. Start testnet: docker-compose -f docker-compose.bft.yml up -d"
log_info "3. Verify validators: curl http://localhost:26657/validators | jq"
log_info ""
log_info "See README_BFT_TESTNET.md for BFT testing instructions"
log_info "========================================="
