# Aura Explorer - Quick Start Guide

## 🚀 Start in 5 Minutes

### Prerequisites
- Aura node running (RPC: 26657, API: 1317)
- Docker & Docker Compose installed

### Start Production Explorer
```bash
cd /home/decri/blockchain-projects/aura/explorer

# 1. Configure (one-time)
cat > .env << 'ENVFILE'
NODE_RPC_URL=http://localhost:26657
NODE_API_URL=http://localhost:1317
CHAIN_ID=aura-testnet-1
DENOM=uaura
DB_PASSWORD=changeme123
ENVFILE

# 2. Start all services
docker-compose -f docker-compose.production.yml up -d

# 3. Check status (wait ~30 seconds for startup)
docker-compose -f docker-compose.production.yml ps
```

### Verify It's Working
```bash
# API health
curl http://localhost/api/health

# Check indexer is syncing blocks
docker logs aura-explorer-indexer --tail 20

# Access frontend
open http://localhost/
```

## 📊 What You Get

- **Frontend:** http://localhost/ (Ping.pub UI)
- **API:** http://localhost/api/ (REST API)
- **WebSocket:** ws://localhost/ws/updates (Real-time)
- **Database:** PostgreSQL on port 5432
- **Cache:** Redis on port 6379

## 🔧 Common Commands

### View Logs
```bash
# All services
docker-compose -f docker-compose.production.yml logs -f

# Specific service
docker logs aura-explorer-indexer -f
docker logs aura-explorer-api-1 -f
```

### Stop/Start
```bash
# Stop all
docker-compose -f docker-compose.production.yml down

# Start all
docker-compose -f docker-compose.production.yml up -d

# Restart single service
docker-compose -f docker-compose.production.yml restart indexer
```

### Check Database
```bash
# Connect to database
docker exec -it aura-explorer-db psql -U explorer -d aura_explorer

# Check indexed blocks
docker exec aura-explorer-db psql -U explorer -d aura_explorer -c "
SELECT COUNT(*) as blocks, MIN(height) as first, MAX(height) as latest FROM blocks;
"
```

## 🐛 Troubleshooting

### Indexer Not Syncing
```bash
# Check node is reachable
curl http://localhost:26657/status

# Check indexer logs
docker logs aura-explorer-indexer --tail 50

# Restart indexer
docker-compose -f docker-compose.production.yml restart indexer
```

### API Errors
```bash
# Check API logs
docker logs aura-explorer-api-1 --tail 50

# Check database connection
docker exec aura-explorer-api-1 env | grep DATABASE_URL
```

### Frontend Not Loading
```bash
# Check nginx
docker logs aura-explorer-nginx

# Verify frontend container
docker-compose -f docker-compose.production.yml ps frontend
```

## 📖 Documentation

- **Complete Guide:** `DEPLOYMENT_GUIDE.md`
- **Detailed Report:** `EXPLORER_COMPLETION_REPORT.md`
- **Feature Summary:** `IMPLEMENTATION_SUMMARY.md`
- **User Manual:** `README.md`

## 🎯 Next Steps

1. ✅ Start explorer (done above)
2. Monitor indexer progress for 5-10 minutes
3. Access frontend and test search
4. Review `DEPLOYMENT_GUIDE.md` for production setup
5. Configure SSL for production (Let's Encrypt)

---

**Status:** Production Ready
**Time to Deploy:** 5 minutes
**Support:** See documentation files
