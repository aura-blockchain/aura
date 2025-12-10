# Aura Block Explorer - Quick Reference

## One-Line Deployment

```bash
./start-explorer.sh
```

Access at: **http://localhost:8088**

---

## Essential Commands

### Start Explorer
```bash
# Standard start
docker compose -f docker-compose.explorer.yml up -d

# With rebuild
docker compose -f docker-compose.explorer.yml up -d --build

# Using helper script
./start-explorer.sh
./start-explorer.sh --build --logs
```

### Stop Explorer
```bash
docker compose -f docker-compose.explorer.yml down

# Remove volumes too
docker compose -f docker-compose.explorer.yml down -v
```

### View Logs
```bash
# Follow logs
docker logs aura-block-explorer -f

# Last 50 lines
docker logs aura-block-explorer --tail 50

# With timestamps
docker logs aura-block-explorer -f --timestamps
```

### Restart
```bash
docker compose -f docker-compose.explorer.yml restart
```

### Check Status
```bash
# Container status
docker ps | grep aura-block-explorer

# Health status
docker inspect aura-block-explorer | jq '.[0].State.Health.Status'

# Test HTTP
curl http://localhost:8088/health
```

---

## Configuration

### Chain Configuration File
```
explorer/ping-pub-explorer/chains/testnet/aura.json
```

### Current Settings
- **Chain ID:** aura-local-4
- **RPC:** http://172.26.0.10:26657 (validator-1)
- **API:** http://172.26.0.10:1317 (validator-1)
- **Port:** 8088
- **SDK Version:** 0.53.0
- **Denom:** uaura (6 decimals)

### Docker Compose File
```
docker-compose.explorer.yml
```

---

## Troubleshooting

### Explorer Won't Start

**Check if testnet is running:**
```bash
docker ps | grep validator
```

**Check network exists:**
```bash
docker network ls | grep aura-testnet
```

**Rebuild from scratch:**
```bash
docker compose -f docker-compose.explorer.yml down
docker compose -f docker-compose.explorer.yml up -d --build --force-recreate
```

### Can't Connect to RPC

**Test RPC from inside container:**
```bash
docker exec aura-block-explorer wget -qO- http://172.26.0.10:26657/status
```

**Check validator status:**
```bash
docker logs aura-validator-1 --tail 20
```

**Restart validators:**
```bash
docker compose -f docker-compose.testnet.yml restart
```

### Build Errors

**Clear build cache:**
```bash
docker builder prune -f
```

**Rebuild with no cache:**
```bash
docker compose -f docker-compose.explorer.yml build --no-cache
docker compose -f docker-compose.explorer.yml up -d
```

---

## Features Checklist

✅ **Block Explorer**
- View blocks in real-time
- Search by block height
- See validator signatures
- Block proposer information

✅ **Transactions**
- Search by hash
- View transaction details
- Decode messages
- See fees and gas used

✅ **Validators**
- Active validator set
- Voting power
- Commission rates
- Uptime/status

✅ **Accounts**
- Search by address
- View balances
- Transaction history
- Token holdings

✅ **Wallet (Keplr)**
- Connect wallet
- Send transactions
- Delegate to validators
- Claim rewards

✅ **Aura Custom Modules**
- Identity (DID)
- Privacy (ZK proofs)
- Compliance (KYC/AML)
- Security modules

---

## Access URLs

| Service | URL | Description |
|---------|-----|-------------|
| Explorer | http://localhost:8088 | Main web interface |
| Health Check | http://localhost:8088/health | Health endpoint |
| Validator 1 RPC | http://localhost:27657 | Direct RPC access |
| Validator 1 API | http://localhost:2317 | Direct API access |
| Prometheus | http://localhost:9094 | Testnet metrics |
| Grafana | http://localhost:3002 | Testnet dashboard |

---

## Network Information

| Parameter | Value |
|-----------|-------|
| Chain ID | aura-local-4 |
| Network | aura_aura-testnet |
| Subnet | 172.26.0.0/16 |
| Gateway | 172.26.0.1 |
| Validators | 4 active |

### Validator IPs

| Validator | IP | RPC | API |
|-----------|----|----|-----|
| validator-1 | 172.26.0.10 | 26657 | 1317 |
| validator-2 | 172.26.0.11 | 26657 | 1317 |
| validator-3 | 172.26.0.12 | 26657 | 1317 |
| validator-4 | 172.26.0.13 | 26657 | 1317 |

---

## Resource Usage

| Resource | Limit |
|----------|-------|
| Memory | 512MB |
| Swap | 768MB |
| CPU | 1 core |
| Disk | ~500MB (build) + 50MB (runtime) |

---

## Files Location

```
aura/
├── docker-compose.explorer.yml          # Main compose file
├── start-explorer.sh                    # Helper script
├── EXPLORER_DEPLOYMENT_GUIDE.md         # Full documentation
├── EXPLORER_QUICK_REFERENCE.md          # This file
└── explorer/
    ├── ping-pub-explorer/
    │   ├── Dockerfile                   # Build instructions
    │   └── chains/testnet/aura.json    # Chain config
    └── [legacy files not used]
```

---

## Quick Tests

### Verify Explorer is Working
```bash
# 1. Check container is running
docker ps | grep aura-block-explorer

# 2. Test HTTP endpoint
curl -I http://localhost:8088/

# 3. Check health
curl http://localhost:8088/health

# 4. View recent logs
docker logs aura-block-explorer --tail 20

# 5. Open in browser
xdg-open http://localhost:8088  # Linux
# or just visit http://localhost:8088
```

### Test RPC Connectivity
```bash
# From host (won't work - wrong network)
curl http://172.26.0.10:26657/status

# From explorer container (should work)
docker exec aura-block-explorer wget -qO- http://172.26.0.10:26657/status | jq .

# From validator (should work)
docker exec aura-validator-1 curl -s http://localhost:26657/status | jq .
```

---

## Common Workflows

### Fresh Deployment
```bash
./start-explorer.sh --build
```

### Update Chain Config
```bash
# 1. Edit config
nano explorer/ping-pub-explorer/chains/testnet/aura.json

# 2. Rebuild
docker compose -f docker-compose.explorer.yml up -d --build
```

### Switch to Different Validator
```bash
# Edit docker-compose.explorer.yml, change:
# RPC_ENDPOINT=http://172.26.0.11:26657  # Use validator-2
# API_ENDPOINT=http://172.26.0.11:1317

docker compose -f docker-compose.explorer.yml up -d
```

### Debug Connection Issues
```bash
# 1. Check testnet
docker compose -f docker-compose.testnet.yml ps

# 2. Check network
docker network inspect aura_aura-testnet

# 3. Test from explorer
docker exec aura-block-explorer sh -c "
  echo 'Testing RPC...'
  wget -qO- http://172.26.0.10:26657/status | head -20
  echo 'Testing API...'
  wget -qO- http://172.26.0.10:1317/cosmos/base/tendermint/v1beta1/node_info | head -20
"

# 4. Check explorer logs
docker logs aura-block-explorer --tail 50
```

---

## Support

### Documentation
- **Full Guide:** `EXPLORER_DEPLOYMENT_GUIDE.md`
- **Testnet Guide:** `TESTNET_QUICKSTART.md`
- **Monitoring:** `TESTNET_MONITORING_GUIDE.md`

### Ping.pub Resources
- **GitHub:** https://github.com/ping-pub/explorer
- **Discord:** https://discord.gg/CmjYVSr6GW

### Log Locations
- **Explorer:** `docker logs aura-block-explorer`
- **Validator 1:** `docker logs aura-validator-1`
- **All Services:** `docker compose -f docker-compose.testnet.yml logs`

---

## Summary

✅ **Deploy in 5 minutes:** `./start-explorer.sh`
✅ **Access:** http://localhost:8088
✅ **Chain:** aura-local-4
✅ **Validators:** 4 active nodes
✅ **Features:** Full explorer + wallet + custom modules
✅ **Resource:** ~200MB RAM, minimal CPU
✅ **Status:** Production ready

**Need help?** See `EXPLORER_DEPLOYMENT_GUIDE.md` for detailed documentation.
