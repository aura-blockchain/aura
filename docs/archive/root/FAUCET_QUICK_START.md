# AURA Faucet - Quick Start Guide

## Prerequisites

- Docker and Docker Compose installed
- AURA testnet running (4 validators)

## Setup in 3 Steps

### Step 1: Start the Testnet (if not already running)

```bash
# Initialize testnet
./scripts/testnet-init.sh

# Build Docker image
docker build -t aurad:latest -f docker/Dockerfile.testnet .

# Start validators
docker-compose -f docker-compose.testnet.yml up -d

# Verify validators are running
docker-compose -f docker-compose.testnet.yml ps
```

### Step 2: Create and Fund Faucet

```bash
# Run setup script (creates wallet and funds it)
./scripts/faucet-setup.sh
```

**This will:**
- Create a new faucet wallet
- Fund it with 100,000,000 AURA
- Generate `.env.faucet` configuration

### Step 3: Start Faucet Service

```bash
# Start faucet services
docker-compose -f docker-compose.faucet.yml --env-file .env.faucet up -d

# Verify it's running
docker-compose -f docker-compose.faucet.yml ps
```

## Access the Faucet

- **Web UI**: http://localhost:8081
- **API**: http://localhost:8081/api/v1
- **Health Check**: http://localhost:8081/api/v1/health

## Test It

```bash
# Get faucet info
curl http://localhost:8081/api/v1/faucet/info

# Request tokens (replace with your address)
curl -X POST http://localhost:8081/api/v1/faucet/request \
  -H "Content-Type: application/json" \
  -d '{"address": "aura1your_address_here"}'

# Check balance
docker exec aura-validator-1 aurad query bank balances aura1your_address_here \
  --chain-id aura-local-4
```

## Common Commands

```bash
# View logs
docker-compose -f docker-compose.faucet.yml logs -f faucet-backend

# Stop faucet
docker-compose -f docker-compose.faucet.yml down

# Restart faucet
docker-compose -f docker-compose.faucet.yml restart faucet-backend

# Check faucet balance
curl http://localhost:8081/api/v1/faucet/info | jq '.balance'

# Clear rate limits (for testing)
docker exec aura-faucet-redis redis-cli FLUSHDB
```

## Troubleshooting

**Faucet won't start?**
```bash
# Check testnet is running
docker ps | grep validator

# Check network exists
docker network ls | grep aura-testnet
```

**Can't connect to validator?**
```bash
# Test from faucet container
docker exec aura-faucet-backend wget -qO- http://aura-observer-1:26657/status
```

**Transactions failing?**
```bash
# Check faucet balance
curl http://localhost:8081/api/v1/faucet/info

# Refill if needed
FAUCET_ADDR=$(grep FAUCET_ADDRESS .env.faucet | cut -d'=' -f2)
docker exec aura-validator-1 bash -c "
  echo 'password123' | aurad tx bank send \
    \$(aurad keys show validator-1 --keyring-backend test --address) \
    ${FAUCET_ADDR} \
    50000000000000uaura \
    --chain-id aura-local-4 \
    --keyring-backend test \
    --gas 200000 \
    --gas-prices 0.025uaura \
    --yes
"
```

## Configuration

Default settings in `.env.faucet`:
- **Amount per request**: 100 AURA
- **Rate limit (IP)**: 20 requests per 24h
- **Rate limit (address)**: 3 requests per 24h

To customize, edit `.env.faucet` and restart:
```bash
vim .env.faucet
docker-compose -f docker-compose.faucet.yml restart faucet-backend
```

## More Details

For comprehensive documentation, see [FAUCET_DEPLOYMENT.md](./FAUCET_DEPLOYMENT.md)
