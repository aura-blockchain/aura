# AURA Local Testnet - Quick Start Guide

## TL;DR - Get Running in 3 Steps

```bash
# 1. Initialize the testnet
./scripts/testnet-init.sh

# 2. Populate Docker volumes
cd testnet-data && ./populate-volumes.sh && cd ..

# 3. Start the testnet
docker-compose -f docker-compose.testnet.yml up -d
```

## Verify It's Working

```bash
# Check status
./scripts/testnet-manage.sh status

# View logs
./scripts/testnet-manage.sh logs validator-1

# Query the chain
curl http://localhost:26657/status | jq '.result.sync_info'
```

## Essential Commands

```bash
# Management script (recommended)
./scripts/testnet-manage.sh <command>

# Available commands:
#   start      - Start all validators
#   stop       - Stop all validators
#   status     - Show status of all nodes
#   logs       - View logs for a validator
#   health     - Check health of all validators
#   bft-test   - Test Byzantine Fault Tolerance
#   ports      - Show all port mappings
#   clean      - Remove all data (requires confirmation)
#   help       - Show all commands
```

## Access Points

### Validator Endpoints
- **Validator 1**: RPC=:26657, API=:1317, gRPC=:9090
- **Validator 2**: RPC=:26757, API=:1417, gRPC=:9190
- **Validator 3**: RPC=:26857, API=:1517, gRPC=:9290
- **Validator 4**: RPC=:26957, API=:1617, gRPC=:9390

### Monitoring
- **Prometheus**: http://localhost:9091
- **Grafana**: http://localhost:3001 (admin/aura-testnet-admin)

## Quick Tests

```bash
# Test Byzantine Fault Tolerance
./scripts/testnet-manage.sh bft-test

# Query a specific validator
./scripts/testnet-manage.sh query validator-2

# Execute command in container
./scripts/testnet-manage.sh exec validator-1 aurad status
```

## Troubleshooting

**Containers not starting?**
```bash
docker-compose -f docker-compose.testnet.yml logs
```

**Need to reset everything?**
```bash
./scripts/testnet-manage.sh clean
./scripts/testnet-init.sh
cd testnet-data && ./populate-volumes.sh && cd ..
docker-compose -f docker-compose.testnet.yml up -d
```

**Chain not producing blocks?**
- Ensure at least 3 of 4 validators are running
- Check logs for consensus errors
- Verify persistent_peers are configured

## Configuration Details

- **Chain ID**: aura-local-4
- **Validators**: 4 nodes with 900,000 AURA staked each
- **Consensus**: CometBFT (requires 3/4 validators for consensus)
- **Block Time**: ~3 seconds
- **Network**: Isolated Docker bridge (172.26.0.0/16)

## Next Steps

1. Test all 27 AURA modules (identity, vcregistry, DEX, bridge, etc.)
2. Deploy smart contracts (vc-issuer, binding-tester)
3. Run load/benchmark tests
4. Review Phase 1 tasks in `/ROADMAP_PRODUCTION.md`

## Full Documentation

See `/TESTNET_SETUP.md` for detailed documentation including:
- Architecture diagrams
- Testing scenarios
- Advanced configuration
- Security considerations
- Troubleshooting guide

---

**Chain ID**: aura-local-4 | **Validators**: 4 | **Consensus**: 3/4 required
