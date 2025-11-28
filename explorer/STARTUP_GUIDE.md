# AURA Block Explorer - Startup Guide

## Quick Verification Checklist

Before starting the explorer, verify these components:

### 1. AURA Node Status

```bash
# Check if AURA node is running
curl http://localhost:26657/health
# Expected: {"jsonrpc":"2.0","id":-1,"result":{}}

# Check node info
curl http://localhost:26657/status
# Should return node status with chain_id

# Check REST API
curl http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info
# Should return node information
```

### 2. File Structure

Ensure all files are present:

```
explorer/
├── config.py                    ✓ Configuration
├── explorer_backend.py          ✓ Main application
├── requirements.txt             ✓ Dependencies
├── Dockerfile                   ✓ Container build
├── docker-compose.yml           ✓ Orchestration
├── test_explorer.py             ✓ Tests
├── verify_setup.py              ✓ Setup verification
├── README.md                    ✓ Documentation
└── STARTUP_GUIDE.md            ✓ This file
```

### 3. Configuration

Check environment variables or edit `config.py`:

```bash
# Required settings
export NODE_RPC_URL="http://localhost:26657"
export NODE_API_URL="http://localhost:1317"
export CHAIN_ID="aura-testnet-1"
export DENOM="uaura"

# Optional settings
export EXPLORER_PORT="8082"
export EXPLORER_DB_PATH="./explorer.db"
```

## Startup Methods

### Method 1: Direct Python (Development)

```bash
# 1. Install dependencies
pip install -r requirements.txt

# 2. Verify setup (optional but recommended)
python verify_setup.py

# 3. Start explorer
python explorer_backend.py

# Expected output:
# INFO - Starting AURA Block Explorer
# INFO - Chain ID: aura-testnet-1
# INFO - Node RPC URL: http://localhost:26657
# INFO - Port: 8082
# * Running on http://0.0.0.0:8082
```

### Method 2: Docker (Production)

```bash
# 1. Build image
docker build -t aura-explorer .

# 2. Run container
docker run -d \
  --name aura-explorer \
  -p 8082:8082 \
  -e NODE_RPC_URL=http://host.docker.internal:26657 \
  -e NODE_API_URL=http://host.docker.internal:1317 \
  -e CHAIN_ID=aura-testnet-1 \
  -v $(pwd)/data:/data \
  aura-explorer

# 3. Check logs
docker logs -f aura-explorer

# 4. Test health
curl http://localhost:8082/health
```

### Method 3: Docker Compose (Recommended)

```bash
# 1. Edit docker-compose.yml if needed
# Update NODE_RPC_URL and NODE_API_URL to point to your AURA node

# 2. Start services
docker-compose up -d

# 3. Check status
docker-compose ps

# 4. View logs
docker-compose logs -f explorer

# 5. Test endpoints
curl http://localhost:8082/
curl http://localhost:8082/health
```

## Post-Startup Verification

### Test Basic Endpoints

```bash
# 1. Explorer info
curl http://localhost:8082/
# Should return explorer metadata

# 2. Health check
curl http://localhost:8082/health
# Should return {"status": "healthy"}

# 3. Analytics dashboard
curl http://localhost:8082/api/analytics/dashboard
# Should return network metrics

# 4. Search (block height)
curl -X POST http://localhost:8082/api/search \
  -H "Content-Type: application/json" \
  -d '{"query": "1"}'
# Should return block information

# 5. Rich list
curl http://localhost:8082/api/richlist?limit=10
# Should return top addresses
```

### Test Search Functionality

```bash
# Search by block height
curl -X POST http://localhost:8082/api/search \
  -H "Content-Type: application/json" \
  -d '{"query": "100"}'

# Search by address (use actual AURA address)
curl -X POST http://localhost:8082/api/search \
  -H "Content-Type: application/json" \
  -d '{"query": "aura1..."}'

# Search by transaction hash (use actual tx hash)
curl -X POST http://localhost:8082/api/search \
  -H "Content-Type: application/json" \
  -d '{"query": "ABC123..."}'
```

### Test WebSocket Connection

```javascript
// In browser console or Node.js
const ws = new WebSocket('ws://localhost:8082/api/ws/updates');

ws.onopen = () => {
    console.log('Connected');
    ws.send('ping');
};

ws.onmessage = (event) => {
    console.log('Received:', event.data);
};

ws.onerror = (error) => {
    console.error('Error:', error);
};
```

## Troubleshooting

### Issue: "Connection refused" when accessing explorer

**Cause**: Explorer not running or wrong port

**Solution**:
```bash
# Check if explorer is running
netstat -an | grep 8082  # Linux/Mac
netstat -an | findstr 8082  # Windows

# Check Docker container
docker ps | grep aura-explorer

# Check logs
docker logs aura-explorer
```

### Issue: Health check returns "disconnected"

**Cause**: Cannot connect to AURA node

**Solution**:
```bash
# Verify node is running
curl http://localhost:26657/health

# Check node RPC configuration in node's config.toml
# Ensure: laddr = "tcp://0.0.0.0:26657"

# If using Docker, ensure proper network connectivity
# Use host.docker.internal instead of localhost
```

### Issue: "Database locked" error

**Cause**: Multiple processes accessing same database

**Solution**:
```bash
# Option 1: Stop other explorer instances
pkill -f explorer_backend.py

# Option 2: Use in-memory database (no persistence)
export EXPLORER_DB_PATH=":memory:"

# Option 3: Use different database path
export EXPLORER_DB_PATH="./explorer_$(date +%s).db"
```

### Issue: Import errors or missing dependencies

**Cause**: Dependencies not installed

**Solution**:
```bash
# Install all dependencies
pip install -r requirements.txt

# Or install individually
pip install flask flask-cors flask-sock requests

# For production deployment
pip install gunicorn gevent

# For testing
pip install pytest pytest-cov
```

### Issue: Configuration errors

**Cause**: Invalid configuration values

**Solution**:
```bash
# Check configuration
python -c "from config import config; config.validate()"

# Reset to defaults
unset NODE_RPC_URL NODE_API_URL CHAIN_ID

# Set valid values
export NODE_RPC_URL="http://localhost:26657"
export CHAIN_ID="aura-testnet-1"
```

## Performance Tuning

### For High Traffic

```bash
# Use production server with multiple workers
gunicorn \
  --bind 0.0.0.0:8082 \
  --workers 4 \
  --worker-class gevent \
  --timeout 120 \
  --log-level info \
  explorer_backend:app
```

### For Development

```bash
# Use Flask development server
export FLASK_ENV=development
export DEBUG=true
python explorer_backend.py
```

### Optimize Caching

Edit `config.py` or set environment variables:

```bash
# Increase cache TTL for stable data
export CACHE_TTL_LONG=1800  # 30 minutes

# Reduce for real-time data
export CACHE_TTL_SHORT=30  # 30 seconds
```

## Monitoring

### Health Check Endpoint

```bash
# Automated health monitoring
while true; do
  curl -sf http://localhost:8082/health || echo "Explorer down!"
  sleep 30
done
```

### Log Monitoring

```bash
# Monitor Docker logs
docker logs -f aura-explorer

# Monitor file logs (if configured)
tail -f /var/log/aura-explorer.log
```

### Metrics Collection

The explorer tracks:
- Search queries
- Analytics calculations
- Cache hit/miss rates
- API request counts

Access metrics via:
```bash
curl http://localhost:8082/api/metrics/hashrate?hours=24
```

## Security Checklist for Production

- [ ] Set strong `ADMIN_API_KEY`
- [ ] Enable `REQUIRE_API_KEY=true`
- [ ] Configure rate limiting
- [ ] Set `DEBUG=false`
- [ ] Use HTTPS with reverse proxy
- [ ] Restrict CORS origins
- [ ] Enable firewall rules
- [ ] Set up monitoring alerts
- [ ] Configure log rotation
- [ ] Use persistent database with backups

## Next Steps

1. ✅ Start the explorer
2. ✅ Verify all endpoints work
3. ⬜ Connect frontend (if applicable)
4. ⬜ Set up monitoring
5. ⬜ Configure backups
6. ⬜ Deploy to production

## Resources

- **Full Documentation**: See `README.md`
- **API Reference**: See `BLOCK_EXPLORER_API.md`
- **Implementation Guide**: See `BLOCK_EXPLORER_IMPLEMENTATION.md`
- **Performance Guide**: See `BLOCK_EXPLORER_PERFORMANCE.md`

## Support

If you encounter issues:

1. Check logs for error messages
2. Run `verify_setup.py` to diagnose issues
3. Review troubleshooting section above
4. Check AURA node is accessible
5. Verify configuration settings

---

**Quick Start Summary**:
```bash
pip install -r requirements.txt
python explorer_backend.py
curl http://localhost:8082/health
```

**Docker Quick Start**:
```bash
docker-compose up -d
docker-compose logs -f explorer
curl http://localhost:8082/health
```
