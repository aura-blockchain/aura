# Aura Block Explorer - Deployment Guide

## Overview

This guide covers deploying the Ping.pub block explorer for the Aura blockchain testnet (chain-id: `aura-local-4`).

**Status:** Ready for deployment
**Technology:** Ping.pub Explorer (Vue.js + Nginx)
**Network:** Connects to existing 4-validator testnet

---

## Quick Start

### Prerequisites

1. **Aura testnet must be running:**
   ```bash
   docker-compose -f docker-compose.testnet.yml ps
   # Should show 4 validators running
   ```

2. **Docker and Docker Compose installed**

### Deploy Explorer (5 minutes)

```bash
cd /home/decri/blockchain-projects/aura

# Build and start the explorer
docker-compose -f docker-compose.explorer.yml up -d --build

# Monitor the build progress (first time takes 3-5 minutes)
docker-compose -f docker-compose.explorer.yml logs -f explorer

# Wait for "Configuration complete; ready for start up" message
```

### Access Explorer

Once deployed, access the explorer at:

**🌐 http://localhost:8088**

---

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Aura Block Explorer                      │
│                     (localhost:8088)                        │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            │ Docker Network: aura_aura-testnet
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
┌───────────────┐   ┌───────────────┐   ┌───────────────┐
│ Validator 1   │   │ Validator 2   │   │ Validator 3   │
│ 172.26.0.10   │   │ 172.26.0.11   │   │ 172.26.0.12   │
│ RPC: 26657    │   │ RPC: 26657    │   │ RPC: 26657    │
│ API: 1317     │   │ API: 1317     │   │ API: 1317     │
└───────────────┘   └───────────────┘   └───────────────┘
```

### Network Configuration

- **Explorer Container:** `aura-block-explorer`
- **Network:** `aura_aura-testnet` (external, created by testnet compose)
- **Primary RPC:** `http://172.26.0.10:26657` (validator-1)
- **Primary API:** `http://172.26.0.10:1317` (validator-1)
- **Fallback RPC:** `http://172.26.0.11:26657` (validator-2)
- **Fallback API:** `http://172.26.0.11:1317` (validator-2)

---

## Configuration Details

### Chain Configuration

The explorer uses the chain configuration at:
```
explorer/ping-pub-explorer/chains/testnet/aura.json
```

**Current Configuration:**
```json
{
  "chain_name": "aura",
  "registry_name": "aura-local-4",
  "api": [
    { "provider": "validator-1", "address": "http://172.26.0.10:1317" },
    { "provider": "validator-2", "address": "http://172.26.0.11:1317" }
  ],
  "rpc": [
    { "provider": "validator-1", "address": "http://172.26.0.10:26657" },
    { "provider": "validator-2", "address": "http://172.26.0.11:26657" }
  ],
  "sdk_version": "0.53.0",
  "coin_type": "118",
  "min_tx_fee": "800",
  "addr_prefix": "aura",
  "assets": [
    {
      "base": "uaura",
      "symbol": "AURA",
      "exponent": "6"
    }
  ]
}
```

### Docker Compose Configuration

File: `docker-compose.explorer.yml`

Key settings:
- **Port:** 8088 (host) → 80 (container)
- **Memory:** 512MB limit (512MB swap)
- **CPU:** 1 core
- **Network:** Connects to external `aura_aura-testnet` network
- **Health Check:** HTTP check on port 80 every 30s

---

## Features

### What the Explorer Provides

✅ **Block Explorer**
- Real-time block updates
- Block details (height, hash, proposer, time)
- Transaction list per block
- Validator signatures

✅ **Transaction Browser**
- Search by transaction hash
- Decode transaction messages
- View gas used and fees
- Transaction status (success/failed)

✅ **Validator Information**
- Active validator set
- Voting power distribution
- Validator details and status
- Commission rates
- Uptime statistics

✅ **Chain Statistics**
- Current block height
- Average block time
- Total transactions
- Active validators count
- Chain parameters

✅ **Wallet Integration** (Keplr)
- Connect Keplr wallet
- View account balances
- Send transactions
- Delegate to validators
- Claim rewards

### Aura-Specific Modules

The explorer will show Aura's custom modules:
- **Identity Module:** DID registration, verification
- **Privacy Module:** Zero-knowledge proofs, selective disclosure
- **Compliance Module:** KYC verification, AML checks
- **Network Security Module:** Attack detection, security scoring
- **Validator Security Module:** Validator reputation tracking
- **Economics Security Module:** Economic attack detection

---

## Usage

### Basic Navigation

1. **Home Page:** Chain overview and recent blocks
2. **Blocks:** Browse all blocks with pagination
3. **Transactions:** Search and view all transactions
4. **Validators:** List of active validators with details
5. **Governance:** View proposals (if any)
6. **Parameters:** Chain parameters and module configs

### Search Functionality

The explorer supports searching by:
- Block height (e.g., `541`)
- Transaction hash (e.g., `0x123...`)
- Account address (e.g., `aura1...`)
- Validator address (e.g., `auravaloper1...`)

### Wallet Connection

1. Click "Connect Wallet" in top-right
2. Select Keplr wallet
3. Approve connection to `aura-local-4`
4. View your account balance and transactions

**Note:** For testnet, you'll need to manually add the chain to Keplr using the chain info from the explorer.

---

## Management Commands

### View Logs

```bash
# Follow explorer logs
docker-compose -f docker-compose.explorer.yml logs -f

# View last 50 lines
docker logs aura-block-explorer --tail 50
```

### Restart Explorer

```bash
# Restart the container
docker-compose -f docker-compose.explorer.yml restart

# Or rebuild and restart
docker-compose -f docker-compose.explorer.yml up -d --build
```

### Stop Explorer

```bash
# Stop the explorer
docker-compose -f docker-compose.explorer.yml down

# Stop and remove volumes
docker-compose -f docker-compose.explorer.yml down -v
```

### Check Status

```bash
# Check if explorer is running
docker ps | grep aura-block-explorer

# Check health status
docker inspect aura-block-explorer | jq '.[0].State.Health.Status'

# Test HTTP endpoint
curl http://localhost:8088/health
```

---

## Troubleshooting

### Explorer Not Loading

**Symptom:** Browser shows "Unable to connect" or blank page

**Solutions:**
1. Check container is running:
   ```bash
   docker ps | grep aura-block-explorer
   ```

2. Check logs for errors:
   ```bash
   docker logs aura-block-explorer --tail 100
   ```

3. Verify nginx is serving files:
   ```bash
   curl -I http://localhost:8088/
   # Should return HTTP 200 OK
   ```

4. Rebuild if needed:
   ```bash
   docker-compose -f docker-compose.explorer.yml up -d --build --force-recreate
   ```

### "Cannot Connect to Node" Error

**Symptom:** Explorer loads but shows "Cannot connect to RPC node"

**Solutions:**
1. Verify validators are running:
   ```bash
   docker-compose -f docker-compose.testnet.yml ps
   ```

2. Test RPC connectivity from explorer container:
   ```bash
   docker exec aura-block-explorer wget -qO- http://172.26.0.10:26657/status
   ```

3. Check testnet network exists:
   ```bash
   docker network ls | grep aura-testnet
   ```

4. Restart validators if needed:
   ```bash
   docker-compose -f docker-compose.testnet.yml restart
   ```

### Build Fails

**Symptom:** Docker build fails with yarn/npm errors

**Solutions:**
1. Clean docker build cache:
   ```bash
   docker builder prune -f
   ```

2. Rebuild with no cache:
   ```bash
   docker-compose -f docker-compose.explorer.yml build --no-cache
   ```

3. Check disk space:
   ```bash
   df -h
   ```

4. If memory issues on WSL2, increase Docker memory limit in Docker Desktop settings

### Slow Performance

**Symptom:** Explorer is slow or unresponsive

**Solutions:**
1. Check resource usage:
   ```bash
   docker stats aura-block-explorer
   ```

2. Increase memory limit in `docker-compose.explorer.yml`:
   ```yaml
   mem_limit: 1g  # Increase from 512m
   memswap_limit: 1536m
   ```

3. Check validator performance:
   ```bash
   docker stats aura-validator-1
   ```

---

## Advanced Configuration

### Adding More RPC Endpoints

Edit `explorer/ping-pub-explorer/chains/testnet/aura.json`:

```json
{
  "rpc": [
    { "provider": "validator-1", "address": "http://172.26.0.10:26657" },
    { "provider": "validator-2", "address": "http://172.26.0.11:26657" },
    { "provider": "validator-3", "address": "http://172.26.0.12:26657" },
    { "provider": "validator-4", "address": "http://172.26.0.13:26657" }
  ]
}
```

Then rebuild:
```bash
docker-compose -f docker-compose.explorer.yml up -d --build
```

### Custom Port

Change the port in `docker-compose.explorer.yml`:

```yaml
ports:
  - "3000:80"  # Use port 3000 instead of 8088
```

### Connecting to External Validators

If validators are on external servers:

```yaml
environment:
  - RPC_ENDPOINT=http://your-validator-ip:26657
  - API_ENDPOINT=http://your-validator-ip:1317
```

---

## Production Considerations

### For Production Deployment

1. **SSL/TLS:**
   - Add nginx SSL configuration
   - Use Let's Encrypt for certificates
   - Enable HTTPS on port 443

2. **Reverse Proxy:**
   - Deploy behind Cloudflare or similar CDN
   - Enable caching for static assets
   - Configure rate limiting

3. **Monitoring:**
   - Add Prometheus metrics export
   - Configure uptime monitoring
   - Set up alerting for downtime

4. **Security:**
   - Limit CORS origins
   - Enable security headers (CSP, HSTS)
   - Regular security updates

5. **Performance:**
   - Increase memory limits based on load
   - Add multiple explorer instances with load balancing
   - Use CDN for static assets

---

## Files and Locations

| File | Purpose |
|------|---------|
| `docker-compose.explorer.yml` | Main docker compose configuration |
| `explorer/ping-pub-explorer/Dockerfile` | Explorer build instructions |
| `explorer/ping-pub-explorer/chains/testnet/aura.json` | Aura chain configuration |
| `explorer/Dockerfile` | Legacy dockerfile (not used) |
| `explorer/docker-compose.yml` | Legacy compose file (not used) |

---

## Support and Documentation

### Ping.pub Documentation
- GitHub: https://github.com/ping-pub/explorer
- Discord: https://discord.gg/CmjYVSr6GW
- Website: https://ping.pub

### Aura Documentation
- Architecture: `ARCHITECTURE_REVIEW_TESTNET_READINESS.md`
- Testnet Guide: `TESTNET_QUICKSTART.md`
- Monitoring: `TESTNET_MONITORING_GUIDE.md`

---

## Summary

**✅ Deployment Time:** 5 minutes (first build may take longer)
**✅ Resource Usage:** ~200MB RAM, minimal CPU
**✅ Access URL:** http://localhost:8088
**✅ Chain ID:** aura-local-4
**✅ Features:** Full block explorer + wallet integration

The explorer is now ready for use with the Aura testnet!
