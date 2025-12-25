# Aura Block Explorer - README

## What is This?

This is a **Ping.pub block explorer** configured for the Aura blockchain testnet. It provides a web interface to browse blocks, transactions, validators, and interact with the Aura chain using Keplr wallet.

---

## Quick Start

### Deploy (From project root)

```bash
cd /home/decri/blockchain-projects/aura
./scripts/start-explorer.sh
```

**Access:** http://localhost:8088

---

## Requirements

1. **Aura testnet running:**
   ```bash
   docker compose -f docker-compose.testnet.yml up -d
   ```

2. **Docker and Docker Compose installed**

---

## Manual Deployment

```bash
cd /home/decri/blockchain-projects/aura

# Build and start
docker compose -f docker-compose.explorer.yml up -d --build

# View logs
docker logs aura-block-explorer -f

# Stop
docker compose -f docker-compose.explorer.yml down
```

---

## Configuration

### Chain Config
```
ping-pub-explorer/chains/testnet/aura.json
```

Current settings:
- **Chain ID:** aura-local-4
- **RPC:** http://172.26.0.50:26657 (observer-1)
- **API:** http://172.26.0.50:1317 (observer-1)

### Docker Compose
```
../docker-compose.explorer.yml
```

Settings:
- **Port:** 8088
- **Network:** aura_aura-testnet (external)
- **Memory:** 512MB

---

## Features

✅ Browse blocks and transactions
✅ Search by block height, tx hash, or address
✅ View validator information
✅ Connect Keplr wallet
✅ Send transactions
✅ Delegate and claim rewards
✅ View Aura custom modules (Identity, Privacy, Compliance, etc.)

---

## Directory Structure

```
explorer/
├── README_DEPLOYMENT.md           # This file
├── ping-pub-explorer/              # Ping.pub source code
│   ├── Dockerfile                  # Build instructions
│   ├── chains/testnet/aura.json   # Aura chain configuration
│   ├── package.json                # Dependencies
│   └── src/                        # Vue.js source code
└── [legacy files]                  # Not used (old implementation)
```

---

## Troubleshooting

### Explorer won't load

**Check testnet:**
```bash
docker ps | grep validator
```

**Check network:**
```bash
docker network ls | grep aura-testnet
```

**Rebuild:**
```bash
docker compose -f ../docker-compose.explorer.yml down
docker compose -f ../docker-compose.explorer.yml up -d --build
```

### Can't connect to RPC

**Test connectivity:**
```bash
docker exec aura-block-explorer wget -qO- http://172.26.0.50:26657/status
```

**Check validators:**
```bash
docker logs aura-observer-1 --tail 20
```

---

## Documentation

Full documentation in project root:
- **EXPLORER_DEPLOYMENT_GUIDE.md** - Complete deployment guide
- **EXPLORER_QUICK_REFERENCE.md** - Command reference
- **EXPLORER_DEPLOYMENT_SUMMARY.md** - Summary and overview

---

## Technology

**Frontend:**
- Vue.js 3 (composition API)
- TailwindCSS + DaisyUI
- TypeScript
- Vite build system

**Backend:**
- None! (Light explorer)
- Direct RPC/API calls to validators
- No indexer or database needed

**Deployment:**
- Docker multi-stage build
- Nginx web server
- Alpine Linux base images

---

## Development

### Local Development (without Docker)

```bash
cd ping-pub-explorer

# Install dependencies
yarn install

# Start dev server
yarn dev

# Build for production
yarn build
```

Dev server runs on http://localhost:5173

### Modifying Chain Config

1. Edit `chains/testnet/aura.json`
2. Rebuild docker image:
   ```bash
   docker compose -f ../docker-compose.explorer.yml up -d --build
   ```

---

## Support

### Ping.pub
- **GitHub:** https://github.com/ping-pub/explorer
- **Discord:** https://discord.gg/CmjYVSr6GW
- **Docs:** https://ping.pub

### Aura
- **Testnet Guide:** ../TESTNET_QUICKSTART.md
- **Monitoring:** ../TESTNET_MONITORING_GUIDE.md
- **Architecture:** ../ARCHITECTURE_REVIEW_TESTNET_READINESS.md

---

## Status

✅ **Ready for deployment**
✅ **Tested with Cosmos SDK 0.53.0**
✅ **Configured for aura-local-4**
✅ **Full feature set enabled**

---

## Quick Commands

```bash
# Deploy
./scripts/start-explorer.sh

# Access
http://localhost:8088

# Logs
docker logs aura-block-explorer -f

# Restart
docker compose -f docker-compose.explorer.yml restart

# Stop
docker compose -f docker-compose.explorer.yml down

# Status
docker ps | grep explorer
```

---

**Need help?** See `EXPLORER_DEPLOYMENT_GUIDE.md` for detailed documentation.
