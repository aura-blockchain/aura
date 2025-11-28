# AURA Multi-Node Local Testnet Setup

This guide walks you through setting up a 4-validator local testnet for the AURA blockchain using Docker Compose.

## Overview

- **Chain ID**: `aura-local-4`
- **Validators**: 4 nodes
- **Consensus**: CometBFT (Byzantine Fault Tolerant)
- **Monitoring**: Prometheus + Grafana
- **Network**: Isolated Docker bridge network

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ installed (for building the binary)
- jq (recommended, for JSON processing)
- At least 8GB RAM and 20GB disk space

## Quick Start

### 1. Initialize the Testnet

```bash
# Make the initialization script executable
chmod +x scripts/testnet-init.sh

# Run the initialization script
./scripts/testnet-init.sh
```

This will:
- Build the `aurad` binary
- Initialize 4 validator nodes
- Create validator keys for each node
- Generate genesis file with all validators
- Configure persistent peers for P2P discovery
- Create initialization data in `./testnet-data/`

### 2. Populate Docker Volumes

```bash
# Navigate to testnet data directory
cd testnet-data

# Run the volume population script
./populate-volumes.sh

# Return to project root
cd ..
```

### 3. Start the Testnet

```bash
# Start all validators and monitoring services
docker-compose -f docker-compose.testnet.yml up -d
```

### 4. Verify the Testnet is Running

```bash
# Check status using the management script
./scripts/testnet-manage.sh status

# Or manually check RPC endpoint
curl http://localhost:26657/status | jq .
```

## Management Commands

A convenient management script is provided at `scripts/testnet-manage.sh`:

```bash
# Make it executable
chmod +x scripts/testnet-manage.sh

# Show all available commands
./scripts/testnet-manage.sh help

# Start the testnet
./scripts/testnet-manage.sh start

# Check status of all validators
./scripts/testnet-manage.sh status

# View logs for a specific validator
./scripts/testnet-manage.sh logs validator-1

# Check health of all validators
./scripts/testnet-manage.sh health

# Test Byzantine Fault Tolerance
./scripts/testnet-manage.sh bft-test

# Stop the testnet
./scripts/testnet-manage.sh stop

# Clean all data (requires confirmation)
./scripts/testnet-manage.sh clean
```

## Port Mappings

### Validator 1 (Primary)
- **RPC**: http://localhost:26657
- **REST API**: http://localhost:1317
- **gRPC**: localhost:9090
- **P2P**: localhost:26656
- **Metrics**: http://localhost:26660

### Validator 2
- **RPC**: http://localhost:26757
- **REST API**: http://localhost:1417
- **gRPC**: localhost:9190
- **P2P**: localhost:26756
- **Metrics**: http://localhost:26760

### Validator 3
- **RPC**: http://localhost:26857
- **REST API**: http://localhost:1517
- **gRPC**: localhost:9290
- **P2P**: localhost:26856
- **Metrics**: http://localhost:26860

### Validator 4
- **RPC**: http://localhost:26957
- **REST API**: http://localhost:1617
- **gRPC**: localhost:9390
- **P2P**: localhost:26956
- **Metrics**: http://localhost:26960

### Monitoring Services
- **Prometheus**: http://localhost:9091
- **Grafana**: http://localhost:3001 (admin/aura-testnet-admin)

## Network Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Bridge Network                     │
│                    172.26.0.0/16                             │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ Validator-1  │  │ Validator-2  │  │ Validator-3  │       │
│  │ 172.26.0.10  │  │ 172.26.0.11  │  │ 172.26.0.12  │       │
│  └──────────────┘  └──────────────┘  └──────────────┘       │
│         │                 │                 │                │
│         └─────────────────┼─────────────────┘                │
│                           │                                  │
│                  ┌──────────────┐                            │
│                  │ Validator-4  │                            │
│                  │ 172.26.0.13  │                            │
│                  └──────────────┘                            │
│                                                               │
│  ┌────────────────┐           ┌─────────────┐               │
│  │  Prometheus    │           │   Grafana   │               │
│  │  (Monitoring)  │───────────│ (Dashboard) │               │
│  └────────────────┘           └─────────────┘               │
└─────────────────────────────────────────────────────────────┘
```

## Testing Scenarios

### 1. Query Chain Status

```bash
# Get current block height
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'

# Get validator info
curl -s http://localhost:26657/validators | jq '.result.validators'

# Query via REST API
curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/blocks/latest | jq .
```

### 2. Test Byzantine Fault Tolerance

```bash
# Use the built-in BFT test
./scripts/testnet-manage.sh bft-test

# Or manually stop one validator
docker-compose -f docker-compose.testnet.yml stop validator-4

# Verify the chain continues to produce blocks (3/4 validators = consensus)
watch -n 1 'curl -s http://localhost:26657/status | jq .result.sync_info.latest_block_height'

# Restart the validator
docker-compose -f docker-compose.testnet.yml start validator-4
```

### 3. Test State Synchronization

```bash
# Stop a validator
docker-compose -f docker-compose.testnet.yml stop validator-3

# Wait for other validators to produce blocks (30+ seconds)
sleep 30

# Restart the stopped validator
docker-compose -f docker-compose.testnet.yml start validator-3

# Check logs to verify state sync
docker-compose -f docker-compose.testnet.yml logs -f validator-3
```

### 4. Execute Transactions

```bash
# Enter validator-1 container
docker-compose -f docker-compose.testnet.yml exec validator-1 sh

# Inside the container:
# Check balance
aurad query bank balances $(aurad keys show validator-1 --keyring-backend test -a)

# Send tokens
aurad tx bank send \
  $(aurad keys show validator-1 --keyring-backend test -a) \
  $(aurad keys show validator-2 --keyring-backend test -a) \
  1000uaura \
  --keyring-backend test \
  --chain-id aura-local-4 \
  --yes

# Query transaction
aurad query tx <TX_HASH>
```

### 5. Test Custom Modules

```bash
# Query Identity module
curl http://localhost:1317/aura/identity/v1/identities

# Query VC Registry
curl http://localhost:1317/aura/vcregistry/v1/credentials

# Query Inclusion Routines
curl http://localhost:1317/aura/inclusionroutines/v1/routines

# Query Governance proposals
curl http://localhost:1317/cosmos/gov/v1beta1/proposals
```

## Monitoring

### Prometheus Metrics

Access Prometheus at http://localhost:9091

Available metrics include:
- `tendermint_consensus_height` - Current block height
- `tendermint_consensus_validators` - Number of validators
- `tendermint_consensus_missing_validators` - Missing validators
- `tendermint_p2p_peers` - Number of connected peers
- `go_memstats_alloc_bytes` - Memory usage

### Grafana Dashboards

Access Grafana at http://localhost:3001

Default credentials:
- **Username**: admin
- **Password**: aura-testnet-admin

Import AURA-specific dashboards from `/grafana/dashboards/`

## Troubleshooting

### Validators Not Starting

```bash
# Check container logs
docker-compose -f docker-compose.testnet.yml logs validator-1

# Check if genesis file is valid
docker-compose -f docker-compose.testnet.yml exec validator-1 aurad validate-genesis

# Verify persistent peers are configured
docker-compose -f docker-compose.testnet.yml exec validator-1 cat /home/aura/.aura/config/config.toml | grep persistent_peers
```

### Validators Not Connecting

```bash
# Check peer connectivity
docker-compose -f docker-compose.testnet.yml exec validator-1 aurad status | jq .SyncInfo

# Verify network connectivity
docker network inspect aura_aura-testnet

# Check if ports are accessible
docker-compose -f docker-compose.testnet.yml exec validator-1 nc -zv validator-2 26656
```

### Chain Halted

```bash
# Check if minimum validators are running (need 3/4 for consensus)
docker-compose -f docker-compose.testnet.yml ps

# Restart all validators
docker-compose -f docker-compose.testnet.yml restart

# If issue persists, check logs for consensus failures
docker-compose -f docker-compose.testnet.yml logs | grep -i error
```

### Clean Restart

```bash
# Stop and remove all containers and volumes
docker-compose -f docker-compose.testnet.yml down -v

# Remove testnet data
rm -rf testnet-data

# Reinitialize from scratch
./scripts/testnet-init.sh
cd testnet-data && ./populate-volumes.sh && cd ..
docker-compose -f docker-compose.testnet.yml up -d
```

## Advanced Configuration

### Modifying Genesis Parameters

Edit the initialization script (`scripts/testnet-init.sh`) and modify the jq commands in Step 5 to adjust:
- Staking parameters (unbonding time, max validators)
- Governance parameters (voting period, deposit amounts)
- Mint parameters (inflation rates)
- Custom module parameters

### Adding More Validators

1. Update `NUM_VALIDATORS` in `testnet-init.sh`
2. Add validator service in `docker-compose.testnet.yml`
3. Add IP to `VALIDATOR_IPS` array
4. Add scrape config in `prometheus/prometheus-testnet.yml`
5. Reinitialize the testnet

### Custom Network Configuration

Edit `docker-compose.testnet.yml` to modify:
- Subnet and IP ranges
- Port mappings
- Resource limits
- Volume configurations

## Security Notes

**This is a LOCAL TESTNET only.** Not suitable for production use.

- Keyring backend is set to `test` (insecure)
- CORS is enabled for all origins
- Default passwords are used
- No TLS encryption
- Duplicate IP connections allowed

For production deployment, see `/docs/ops/PRODUCTION_DEPLOYMENT.md`.

## Next Steps

After successfully running the local testnet:

1. **Test Custom Modules**: Execute transactions and queries for all 27 AURA modules
2. **Deploy Smart Contracts**: Deploy and test WASM contracts (vc-issuer, binding-tester)
3. **Benchmark Performance**: Run load tests to measure TPS and latency
4. **Test Upgrades**: Practice coordinated network upgrades
5. **Progress to Cloud Testnet**: Deploy to cloud infrastructure (Phase 2 in ROADMAP_PRODUCTION.md)

## Resources

- **Main Roadmap**: `/ROADMAP_PRODUCTION.md`
- **Production Deployment**: `/docs/ops/PRODUCTION_DEPLOYMENT.md`
- **Validator Onboarding**: `/docs/validators/ONBOARDING.md`
- **Module Documentation**: `/docs/modules/`
- **Docker Compose Reference**: `docker-compose.testnet.yml`

## Support

For issues or questions:
1. Check logs: `./scripts/testnet-manage.sh logs <validator>`
2. Review troubleshooting section above
3. Check AURA documentation in `/docs/`
4. Review Cosmos SDK documentation: https://docs.cosmos.network/

---

**Happy Testing!**
