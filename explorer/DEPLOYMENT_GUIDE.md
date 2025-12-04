# Aura Block Explorer - Production Deployment Guide

## Overview

Production-grade blockchain explorer for the Aura network with:
- **Database Indexer**: PostgreSQL with full historical data
- **WebSocket Manager**: Real-time blockchain updates
- **Load-Balanced API**: High-availability REST API
- **Redis Caching**: High-performance query caching
- **Nginx Reverse Proxy**: Load balancing and SSL termination

---

## Architecture

```
Internet
   │
   ├─> Nginx (Load Balancer + SSL)
   │      │
   │      ├─> Frontend (Ping.pub)
   │      ├─> API-1 (Explorer Backend)
   │      ├─> API-2 (Explorer Backend)
   │      └─> WebSocket Server
   │
   ├─> PostgreSQL (Indexed blockchain data)
   ├─> Redis (Query cache)
   └─> Indexer (Background blockchain indexing)
```

---

## Prerequisites

### Hardware Requirements
- **CPU**: 4+ cores
- **RAM**: 8GB minimum, 16GB recommended
- **Storage**: 100GB+ SSD (grows with blockchain)
- **Network**: 100Mbps+

### Software Requirements
- Docker 24.0+
- Docker Compose 2.0+
- Access to Aura node RPC (port 26657) and REST API (port 1317)

---

## Quick Start (Development)

### 1. Clone and Configure

```bash
cd /home/decri/blockchain-projects/aura/explorer

# Create environment file
cp .env.example .env

# Edit configuration
nano .env
```

### 2. Configure Environment

```bash
# .env file
NODE_RPC_URL=http://localhost:26657
NODE_API_URL=http://localhost:1317
CHAIN_ID=aura-testnet-1
DENOM=uaura
DB_PASSWORD=changeme123
```

### 3. Start Services

```bash
# Start all services
docker-compose -f docker-compose.production.yml up -d

# Check logs
docker-compose -f docker-compose.production.yml logs -f

# Check service status
docker-compose -f docker-compose.production.yml ps
```

### 4. Verify Deployment

```bash
# Check API health
curl http://localhost/api/health

# Check WebSocket
wscat -c ws://localhost/ws/updates

# Check frontend
curl http://localhost/

# Check database indexing
docker logs aura-explorer-indexer
```

---

## Production Deployment

### Step 1: Server Setup

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### Step 2: SSL Certificates (Let's Encrypt)

```bash
# Install Certbot
sudo apt install certbot -y

# Get SSL certificate
sudo certbot certonly --standalone -d explorer.aura.network

# Copy certificates
sudo mkdir -p /home/decri/blockchain-projects/aura/explorer/ssl
sudo cp /etc/letsencrypt/live/explorer.aura.network/fullchain.pem ./ssl/
sudo cp /etc/letsencrypt/live/explorer.aura.network/privkey.pem ./ssl/
```

### Step 3: Configure for Production

```bash
cd /home/decri/blockchain-projects/aura/explorer

# Set strong passwords
echo "DB_PASSWORD=$(openssl rand -hex 32)" >> .env

# Configure domains
echo "DOMAIN=explorer.aura.network" >> .env

# Enable HTTPS in nginx.prod.conf
nano nginx.prod.conf
# Uncomment HTTPS server block
```

### Step 4: Deploy

```bash
# Build and start
docker-compose -f docker-compose.production.yml build
docker-compose -f docker-compose.production.yml up -d

# Monitor deployment
docker-compose -f docker-compose.production.yml logs -f
```

### Step 5: Verify Production

```bash
# Check all services are running
docker-compose -f docker-compose.production.yml ps

# Test HTTPS
curl https://explorer.aura.network/api/health

# Check indexer progress
docker exec aura-explorer-indexer python -c "
from indexer import BlockchainIndexer
import asyncio
indexer = BlockchainIndexer('...', '...', '...')
asyncio.run(indexer.initialize())
print(asyncio.run(indexer.get_latest_indexed_height()))
"
```

---

## Service Management

### Start/Stop Services

```bash
# Start all
docker-compose -f docker-compose.production.yml up -d

# Stop all
docker-compose -f docker-compose.production.yml down

# Restart specific service
docker-compose -f docker-compose.production.yml restart explorer-api-1

# View logs
docker-compose -f docker-compose.production.yml logs -f indexer
```

### Database Management

```bash
# Connect to PostgreSQL
docker exec -it aura-explorer-db psql -U explorer -d aura_explorer

# Backup database
docker exec aura-explorer-db pg_dump -U explorer aura_explorer > backup.sql

# Restore database
cat backup.sql | docker exec -i aura-explorer-db psql -U explorer aura_explorer

# Check database size
docker exec aura-explorer-db psql -U explorer -d aura_explorer -c "
SELECT pg_size_pretty(pg_database_size('aura_explorer'));"
```

### Redis Cache Management

```bash
# Connect to Redis
docker exec -it aura-explorer-redis redis-cli

# Clear cache
docker exec aura-explorer-redis redis-cli FLUSHALL

# Check cache size
docker exec aura-explorer-redis redis-cli INFO memory
```

---

## Monitoring

### Check Indexer Status

```bash
# View indexer logs
docker logs aura-explorer-indexer -f

# Check indexing progress
docker exec aura-explorer-db psql -U explorer -d aura_explorer -c "
SELECT
    MAX(height) as latest_block,
    COUNT(*) as total_blocks,
    COUNT(DISTINCT proposer_address) as unique_proposers
FROM blocks;"
```

### Check API Performance

```bash
# Test API response time
time curl http://localhost/api/analytics/dashboard

# Check connection count
docker exec aura-explorer-nginx nginx -T | grep worker_connections
```

### Check System Resources

```bash
# Container resource usage
docker stats

# Disk usage
df -h

# Database connections
docker exec aura-explorer-db psql -U explorer -d aura_explorer -c "
SELECT count(*) FROM pg_stat_activity;"
```

---

## Troubleshooting

### Indexer Not Syncing

```bash
# Check node connectivity
curl http://localhost:26657/status

# Check indexer logs
docker logs aura-explorer-indexer --tail 100

# Restart indexer
docker-compose -f docker-compose.production.yml restart indexer
```

### Database Connection Errors

```bash
# Check PostgreSQL is running
docker-compose -f docker-compose.production.yml ps postgres

# Check connection string
docker exec aura-explorer-indexer env | grep DATABASE_URL

# Test connection
docker exec aura-explorer-db psql -U explorer -d aura_explorer -c "SELECT 1;"
```

### High Memory Usage

```bash
# Check Redis memory
docker exec aura-explorer-redis redis-cli INFO memory

# Restart Redis with new limit
docker-compose -f docker-compose.production.yml stop redis
# Edit docker-compose.production.yml to adjust maxmemory
docker-compose -f docker-compose.production.yml up -d redis
```

### Slow API Responses

```bash
# Check database indexes
docker exec aura-explorer-db psql -U explorer -d aura_explorer -c "
SELECT schemaname, tablename, indexname
FROM pg_indexes
WHERE schemaname = 'public';"

# Analyze query performance
docker exec aura-explorer-db psql -U explorer -d aura_explorer -c "
EXPLAIN ANALYZE SELECT * FROM transactions WHERE height > 1000 LIMIT 100;"
```

---

## Maintenance

### Regular Tasks

**Daily:**
- Check service health
- Monitor disk usage
- Review error logs

**Weekly:**
- Backup database
- Review API performance metrics
- Update SSL certificates if needed

**Monthly:**
- Update Docker images
- Optimize database (VACUUM, ANALYZE)
- Review and archive old logs

### Database Optimization

```bash
# Vacuum and analyze
docker exec aura-explorer-db psql -U explorer -d aura_explorer -c "
VACUUM ANALYZE;"

# Reindex
docker exec aura-explorer-db psql -U explorer -d aura_explorer -c "
REINDEX DATABASE aura_explorer;"
```

### Log Rotation

```bash
# Configure Docker log rotation
cat > /etc/docker/daemon.json <<EOF
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

sudo systemctl restart docker
```

---

## Scaling

### Horizontal Scaling (More API Instances)

Edit `docker-compose.production.yml`:

```yaml
# Add more API instances
explorer-api-3:
  # Copy explorer-api-1 configuration
  ...

# Update nginx upstream
upstream explorer_api {
  server explorer-api-1:8082;
  server explorer-api-2:8082;
  server explorer-api-3:8082;  # Add new instance
}
```

### Vertical Scaling (More Resources)

```yaml
# Add resource limits in docker-compose.production.yml
services:
  postgres:
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
```

---

## Security Checklist

- [ ] SSL/TLS enabled (HTTPS)
- [ ] Strong database password
- [ ] Firewall configured (only ports 80, 443 open)
- [ ] Regular security updates
- [ ] Database backups automated
- [ ] API rate limiting enabled
- [ ] Monitoring and alerting configured
- [ ] Non-root Docker containers
- [ ] Secrets in environment variables (not in code)
- [ ] Log aggregation configured

---

## Performance Tuning

### PostgreSQL

```sql
-- Increase shared_buffers (25% of RAM)
ALTER SYSTEM SET shared_buffers = '2GB';

-- Increase work_mem
ALTER SYSTEM SET work_mem = '64MB';

-- Increase maintenance_work_mem
ALTER SYSTEM SET maintenance_work_mem = '512MB';

-- Reload configuration
SELECT pg_reload_conf();
```

### Redis

```bash
# Increase max memory
docker-compose -f docker-compose.production.yml stop redis
# Edit docker-compose: --maxmemory 4gb
docker-compose -f docker-compose.production.yml up -d redis
```

---

## Support

**Documentation:**
- API Documentation: http://localhost/api/
- Frontend Guide: ./ping-pub-explorer/README.md
- Architecture: ./EXPLORER_ANALYSIS_AND_IMPLEMENTATION_PLAN.md

**Logs:**
- API: `docker logs aura-explorer-api-1`
- Indexer: `docker logs aura-explorer-indexer`
- WebSocket: `docker logs aura-explorer-websocket`
- Database: `docker logs aura-explorer-db`

---

## Deployment Checklist

### Pre-Deployment
- [ ] Node RPC/API accessible
- [ ] Environment variables configured
- [ ] SSL certificates obtained
- [ ] Domain DNS configured
- [ ] Firewall rules set

### Deployment
- [ ] Services started successfully
- [ ] Database schema created
- [ ] Indexer syncing blocks
- [ ] API responding to requests
- [ ] WebSocket connections working
- [ ] Frontend accessible

### Post-Deployment
- [ ] Monitor for 24 hours
- [ ] Set up automated backups
- [ ] Configure monitoring alerts
- [ ] Document any custom changes
- [ ] Create runbook for on-call

---

**Version:** 1.0
**Last Updated:** December 4, 2025
**Status:** Production Ready
