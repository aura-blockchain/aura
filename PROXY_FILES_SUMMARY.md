# Aura Testnet Proxy - Files Summary

This document lists all files created for the public endpoint proxy setup.

## Files Created

### Configuration Files

#### `/home/decri/blockchain-projects/aura/docker-compose.proxy.yml`
Docker Compose configuration for the nginx reverse proxy service.

**Key Features:**
- Nginx 1.25-alpine container
- Connects to testnet network (`aura_aura-testnet`)
- Exposes ports 80 (HTTP) and 9090 (gRPC)
- Health check configured
- Log volume for nginx logs
- Resource limits (256MB RAM, 0.5 CPU)

**Usage:**
```bash
docker compose -f docker-compose.proxy.yml up -d
```

#### `/home/decri/blockchain-projects/aura/nginx/testnet-proxy.conf`
Nginx reverse proxy configuration with comprehensive security and performance settings.

**Key Features:**
- Rate limiting zones (RPC: 30r/s, API: 50r/s, gRPC: 100r/s)
- CORS headers for browser access
- WebSocket support for subscriptions
- Security headers (X-Frame-Options, X-Content-Type-Options, etc.)
- Connection limits (100 concurrent per IP)
- Load balancing configuration (validator-1 primary, others as backup)
- Health check endpoint
- Prometheus metrics endpoint (restricted to private networks)
- Swagger documentation proxy

**Upstream Backends:**
- RPC: validator-1:26657
- API: validator-1:1317
- gRPC: validator-1:9090

**Endpoints:**
- `/` - Service information
- `/health` - Health check
- `/rpc` - Tendermint RPC
- `/api` - Cosmos REST API
- `/metrics` - Prometheus metrics (private)
- `/swagger` - API documentation

### Scripts

#### `/home/decri/blockchain-projects/aura/scripts/start-proxy.sh`
Helper script to start the proxy with validation and health checks.

**Features:**
- Checks if testnet network exists
- Verifies validator health
- Starts proxy with docker compose
- Waits for proxy to become healthy
- Tests endpoints automatically
- Displays connection information

**Usage:**
```bash
./scripts/start-proxy.sh
```

**Permissions:**
```bash
chmod +x scripts/start-proxy.sh
```

#### `/home/decri/blockchain-projects/aura/scripts/test-proxy.sh`
Comprehensive test script for all proxy endpoints.

**Tests Performed:**
1. Proxy container running
2. Health endpoint
3. Service info endpoint
4. RPC status endpoint
5. RPC block endpoint
6. API node info endpoint
7. API latest block endpoint
8. API validators endpoint
9. API supply endpoint
10. CORS headers
11. gRPC endpoint (if grpcurl available)
12. WebSocket upgrade
13. Rate limiting
14. Swagger documentation

**Usage:**
```bash
./scripts/test-proxy.sh
```

**Permissions:**
```bash
chmod +x scripts/test-proxy.sh
```

### Documentation

#### `/home/decri/blockchain-projects/aura/PROXY_SETUP.md`
Quick setup and usage guide for the proxy.

**Contents:**
- Quick start instructions
- Available endpoints table
- Common operations
- Usage examples (curl, JavaScript, gRPC)
- Configuration options
- Troubleshooting guide
- Architecture diagram
- Security considerations

**Target Audience:**
Developers who need to quickly set up and use the proxy.

#### `/home/decri/blockchain-projects/aura/docs/TESTNET_PUBLIC_ENDPOINTS.md`
Comprehensive documentation for all public endpoints.

**Contents:**
- Detailed architecture
- Prerequisites and setup
- Complete endpoint reference
- Extensive usage examples
- Rate limit documentation
- Security considerations
- Load balancing configuration
- Monitoring and troubleshooting
- Production deployment checklist

**Target Audience:**
Developers, operators, and security teams who need detailed information.

#### `/home/decri/blockchain-projects/aura/nginx/README.md`
Documentation specific to nginx configurations.

**Contents:**
- Configuration file descriptions
- SSL/TLS certificate setup
- Testing configurations
- Customization guide (rate limits, CORS, load balancing)
- Monitoring with logs
- Troubleshooting nginx-specific issues

**Target Audience:**
DevOps engineers and system administrators.

#### `/home/decri/blockchain-projects/aura/PROXY_FILES_SUMMARY.md`
This file - comprehensive listing of all created files.

## File Structure

```
/home/decri/blockchain-projects/aura/
├── docker-compose.proxy.yml          # Main proxy service configuration
├── PROXY_SETUP.md                    # Quick setup guide
├── PROXY_FILES_SUMMARY.md           # This file
├── nginx/
│   ├── testnet-proxy.conf           # Nginx configuration
│   ├── README.md                     # Nginx-specific docs
│   ├── rpc-proxy.conf               # Existing production config (SSL)
│   ├── ssl-config.conf              # Existing SSL config
│   └── ssl/                         # SSL certificates directory
├── scripts/
│   ├── start-proxy.sh               # Helper to start proxy
│   └── test-proxy.sh                # Endpoint testing script
└── docs/
    └── TESTNET_PUBLIC_ENDPOINTS.md  # Comprehensive endpoint docs
```

## Quick Reference

### Start the Proxy
```bash
# Option 1: Using helper script (recommended)
./scripts/start-proxy.sh

# Option 2: Direct docker compose
docker compose -f docker-compose.proxy.yml up -d
```

### Test the Proxy
```bash
# Run all tests
./scripts/test-proxy.sh

# Quick manual test
curl http://localhost/health
curl http://localhost/rpc/status | jq
```

### View Logs
```bash
# All logs
docker compose -f docker-compose.proxy.yml logs -f

# Access logs only
docker exec aura-testnet-proxy tail -f /var/log/nginx/testnet-access.log
```

### Stop the Proxy
```bash
docker compose -f docker-compose.proxy.yml down
```

### Restart the Proxy
```bash
docker compose -f docker-compose.proxy.yml restart
```

## Endpoints

| Endpoint | URL | Description |
|----------|-----|-------------|
| Service Info | `http://localhost/` | JSON with endpoint information |
| Health Check | `http://localhost/health` | Proxy health status |
| RPC | `http://localhost/rpc` | Tendermint RPC endpoint |
| API | `http://localhost/api` | Cosmos SDK REST API |
| gRPC | `localhost:9090` | Native gRPC endpoint |
| Swagger | `http://localhost/api/swagger/` | API documentation |
| WebSocket | `ws://localhost/rpc/websocket` | Event subscriptions |

## Backend Mapping

The proxy forwards requests to validator-1:

| Service | Frontend | Backend |
|---------|----------|---------|
| RPC | `http://localhost/rpc` | `validator-1:26657` |
| API | `http://localhost/api` | `validator-1:1317` |
| gRPC | `localhost:9090` | `validator-1:9090` |

## Rate Limits

| Endpoint | Rate | Burst | Connection Limit |
|----------|------|-------|------------------|
| RPC | 30 req/s | +20 | 100 per IP |
| API | 50 req/s | +30 | 100 per IP |
| gRPC | 100 req/s | +50 | 200 per IP |

## Configuration Changes

### Enable Load Balancing

Edit `nginx/testnet-proxy.conf` and uncomment backup validators:

```nginx
upstream testnet_rpc_backend {
    server validator-1:26657 max_fails=3 fail_timeout=30s weight=10;
    server validator-2:26657 max_fails=3 fail_timeout=30s weight=5;  # Uncomment
    server validator-3:26657 max_fails=3 fail_timeout=30s weight=5;  # Uncomment
    server validator-4:26657 max_fails=3 fail_timeout=30s weight=5;  # Uncomment
    keepalive 32;
}
```

### Adjust Rate Limits

Edit `nginx/testnet-proxy.conf`:

```nginx
# Change from:
limit_req_zone $binary_remote_addr zone=testnet_rpc_limit:10m rate=30r/s;

# To (example - increase to 100r/s):
limit_req_zone $binary_remote_addr zone=testnet_rpc_limit:10m rate=100r/s;
```

### Restrict CORS

Edit `nginx/testnet-proxy.conf`:

```nginx
# Change from:
add_header Access-Control-Allow-Origin "*" always;

# To:
add_header Access-Control-Allow-Origin "https://yourdomain.com" always;
```

After any configuration change, reload nginx:
```bash
docker exec aura-testnet-proxy nginx -s reload
```

## Dependencies

### Required
- Docker with Compose plugin
- Running testnet (docker-compose.testnet.yml)

### Optional (for testing)
- `curl` - HTTP testing
- `jq` - JSON parsing
- `grpcurl` - gRPC testing
- `wscat` - WebSocket testing

## Security Notes

**Current configuration is for DEVELOPMENT/TESTING only.**

For production:
1. Enable SSL/TLS (HTTPS)
2. Restrict CORS to specific domains
3. Implement request authentication
4. Configure firewall rules
5. Set up DDoS protection
6. Use production SSL certificates
7. Disable metrics endpoint for public access
8. Implement proper monitoring and alerting

## Integration Examples

### JavaScript/Browser
```javascript
fetch('http://localhost/api/cosmos/base/tendermint/v1beta1/blocks/latest')
  .then(r => r.json())
  .then(data => console.log(data));
```

### curl
```bash
curl http://localhost/rpc/status | jq
```

### gRPC (grpcurl)
```bash
grpcurl -plaintext localhost:9090 list
```

### WebSocket (wscat)
```bash
wscat -c ws://localhost/rpc/websocket
{"jsonrpc":"2.0","method":"subscribe","params":["tm.event='NewBlock'"],"id":1}
```

## Troubleshooting

### Proxy won't start
- Ensure testnet is running: `docker compose -f docker-compose.testnet.yml ps`
- Check if port 80 is available: `sudo lsof -i :80`

### Connection refused
- Check validator status: `docker compose -f docker-compose.testnet.yml ps`
- Test validator directly: `curl http://localhost:27657/status`

### Rate limited (503)
- Reduce request frequency
- Increase rate limits in `nginx/testnet-proxy.conf`

## Next Steps

1. Start the proxy: `./scripts/start-proxy.sh`
2. Run tests: `./scripts/test-proxy.sh`
3. Try example queries from `PROXY_SETUP.md`
4. Read detailed docs in `docs/TESTNET_PUBLIC_ENDPOINTS.md`

## Support

For detailed information:
- Quick guide: `PROXY_SETUP.md`
- Full documentation: `docs/TESTNET_PUBLIC_ENDPOINTS.md`
- Nginx specifics: `nginx/README.md`
