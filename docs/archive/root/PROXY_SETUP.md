# Aura Testnet Public Endpoint Proxy Setup

This document provides a quick setup guide for the public RPC/API endpoint proxy for the Aura testnet.

## Overview

The proxy service provides public HTTP and gRPC endpoints for accessing the Aura testnet, with features including:
- Rate limiting to prevent abuse
- CORS headers for browser-based applications
- Security headers
- WebSocket support for subscriptions
- Load balancing capability (configurable)

## Quick Start

### 1. Ensure Testnet is Running

```bash
# Check if testnet validators are running
docker compose -f docker-compose.testnet.yml ps

# All 4 validators should be listed with "Up" status
# They may show as "unhealthy" initially - this is normal during startup
```

### 2. Start the Proxy

**Option A: Using the helper script (recommended)**
```bash
./scripts/start-proxy.sh
```

**Option B: Using docker compose directly**
```bash
docker compose -f docker-compose.proxy.yml up -d
```

### 3. Verify Proxy is Running

```bash
# Check proxy status
docker compose -f docker-compose.proxy.yml ps

# Test health endpoint
curl http://localhost/health

# Expected output:
# {"status":"healthy","service":"aura-testnet-proxy","version":"1.0.0"}
```

### 4. Test the Endpoints

```bash
# Test RPC endpoint
curl http://localhost/rpc/status | jq

# Test API endpoint
curl http://localhost/api/cosmos/base/tendermint/v1beta1/node_info | jq

# Test in browser
# Open: http://localhost/api/swagger/
```

## Available Endpoints

Once the proxy is running, you can access:

| Endpoint | URL | Description |
|----------|-----|-------------|
| **Service Info** | `http://localhost/` | JSON with endpoint information |
| **Health Check** | `http://localhost/health` | Proxy health status |
| **RPC** | `http://localhost/rpc` | Tendermint RPC endpoint |
| **API** | `http://localhost/api` | Cosmos SDK REST API |
| **gRPC** | `localhost:9090` | Native gRPC endpoint (requires gRPC client) |
| **Swagger Docs** | `http://localhost/api/swagger/` | Interactive API documentation |
| **WebSocket** | `ws://localhost/rpc/websocket` | Real-time event subscriptions |

## Backend Configuration

The proxy currently forwards requests to **validator-1**:
- **RPC**: validator-1:26657
- **API**: validator-1:1317
- **gRPC**: validator-1:9090

The configuration supports load balancing across all validators. To enable it, see the Load Balancing section below.

## Common Operations

### View Proxy Logs

```bash
# All logs
docker compose -f docker-compose.proxy.yml logs -f

# Access logs only
docker exec aura-testnet-proxy tail -f /var/log/nginx/testnet-access.log

# Error logs only
docker exec aura-testnet-proxy tail -f /var/log/nginx/testnet-error.log
```

### Restart Proxy

```bash
docker compose -f docker-compose.proxy.yml restart
```

### Stop Proxy

```bash
docker compose -f docker-compose.proxy.yml down
```

### Reload Configuration (without downtime)

```bash
# After editing nginx/testnet-proxy.conf
docker exec aura-testnet-proxy nginx -s reload
```

## Example Usage

### Query Blockchain Data

```bash
# Get node status
curl http://localhost/rpc/status | jq

# Get latest block
curl http://localhost/api/cosmos/base/tendermint/v1beta1/blocks/latest | jq

# Get all validators
curl http://localhost/api/cosmos/staking/v1beta1/validators | jq

# Get token supply
curl http://localhost/api/cosmos/bank/v1beta1/supply | jq

# Get account balance (replace with actual address)
curl http://localhost/api/cosmos/bank/v1beta1/balances/aura1... | jq
```

### Subscribe to Events (WebSocket)

```bash
# Install wscat if needed
npm install -g wscat

# Connect to WebSocket
wscat -c ws://localhost/rpc/websocket

# Subscribe to new blocks (after connection)
{"jsonrpc":"2.0","method":"subscribe","params":["tm.event='NewBlock'"],"id":1}

# Subscribe to transactions
{"jsonrpc":"2.0","method":"subscribe","params":["tm.event='Tx'"],"id":2}
```

### Use from JavaScript/Browser

```javascript
// Fetch latest block
fetch('http://localhost/api/cosmos/base/tendermint/v1beta1/blocks/latest')
  .then(response => response.json())
  .then(data => console.log('Latest block:', data))
  .catch(error => console.error('Error:', error));

// WebSocket subscription
const ws = new WebSocket('ws://localhost/rpc/websocket');

ws.onopen = () => {
  console.log('Connected to Aura testnet');
  ws.send(JSON.stringify({
    jsonrpc: '2.0',
    method: 'subscribe',
    params: ["tm.event='NewBlock'"],
    id: 1
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('New block:', data);
};
```

### Use with gRPC (requires grpcurl)

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List all services
grpcurl -plaintext localhost:9090 list

# Get node info
grpcurl -plaintext localhost:9090 \
  cosmos.base.tendermint.v1beta1.Service/GetNodeInfo

# Get latest block
grpcurl -plaintext localhost:9090 \
  cosmos.base.tendermint.v1beta1.Service/GetLatestBlock

# Get account balance
grpcurl -plaintext \
  -d '{"address":"aura1...","denom":"uaura"}' \
  localhost:9090 \
  cosmos.bank.v1beta1.Query/Balance
```

## Configuration

### Rate Limits

Current rate limits (per IP address):
- **RPC**: 30 requests/second (burst: +20)
- **API**: 50 requests/second (burst: +30)
- **gRPC**: 100 requests/second (burst: +50)
- **Connections**: 100 concurrent connections
- **Max body size**: 2 MB

To adjust, edit `/home/decri/blockchain-projects/aura/nginx/testnet-proxy.conf` and restart the proxy.

### CORS Configuration

Currently configured to allow all origins (`*`) for development.

For production, restrict to specific domains by editing `nginx/testnet-proxy.conf`:
```nginx
# Change from:
add_header Access-Control-Allow-Origin "*" always;

# To:
add_header Access-Control-Allow-Origin "https://yourdomain.com" always;
```

### Load Balancing

To distribute traffic across all validators:

1. Edit `/home/decri/blockchain-projects/aura/nginx/testnet-proxy.conf`
2. Uncomment the backup validator lines in the upstream blocks:

```nginx
upstream testnet_rpc_backend {
    server validator-1:26657 max_fails=3 fail_timeout=30s weight=10;
    server validator-2:26657 max_fails=3 fail_timeout=30s weight=5;  # Uncomment this
    server validator-3:26657 max_fails=3 fail_timeout=30s weight=5;  # Uncomment this
    server validator-4:26657 max_fails=3 fail_timeout=30s weight=5;  # Uncomment this
    keepalive 32;
}
```

3. Restart the proxy:
```bash
docker compose -f docker-compose.proxy.yml restart
```

## Troubleshooting

### Proxy Won't Start

**Error: "network aura_aura-testnet not found"**

Solution: Start the testnet first:
```bash
docker compose -f docker-compose.testnet.yml up -d
```

**Error: "port 80 already in use"**

Solution: Another service is using port 80. Either stop that service or change the proxy port in `docker-compose.proxy.yml`.

### Connection Errors

**502 Bad Gateway**

Cause: Validators are not responding

Solution:
```bash
# Check if validators are running
docker compose -f docker-compose.testnet.yml ps

# Check validator logs
docker compose -f docker-compose.testnet.yml logs validator-1

# Restart validators if needed
docker compose -f docker-compose.testnet.yml restart
```

**503 Service Unavailable**

Cause: Rate limit exceeded

Solution:
- Reduce request frequency
- Increase rate limits in `nginx/testnet-proxy.conf`
- Implement request batching in your application

### CORS Errors

Check browser console for the specific CORS error. Common issues:
- Origin not allowed (if you've restricted CORS)
- Missing required headers
- Wrong request method

Test CORS headers:
```bash
curl -H "Origin: http://example.com" -I http://localhost/rpc/status
```

### WebSocket Connection Fails

Test WebSocket upgrade:
```bash
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
  http://localhost/rpc/websocket
```

Should return HTTP 101 Switching Protocols.

## Files Created

This setup consists of the following files:

| File | Purpose |
|------|---------|
| `docker-compose.proxy.yml` | Docker Compose configuration for the proxy service |
| `nginx/testnet-proxy.conf` | Nginx reverse proxy configuration with rate limiting and CORS |
| `nginx/README.md` | Nginx configuration documentation |
| `scripts/start-proxy.sh` | Helper script to start the proxy with validation |
| `docs/TESTNET_PUBLIC_ENDPOINTS.md` | Comprehensive endpoint documentation |
| `PROXY_SETUP.md` | This file - quick setup guide |

## Documentation

- **Quick Setup**: `PROXY_SETUP.md` (this file)
- **Detailed Endpoint Docs**: `docs/TESTNET_PUBLIC_ENDPOINTS.md`
- **Nginx Configuration**: `nginx/README.md`
- **Testnet Setup**: `docker-compose.testnet.yml` (see header comments)

## Security Considerations

This configuration is designed for **development and testing**. For production:

1. **Enable SSL/TLS** (HTTPS)
2. **Restrict CORS** to specific domains
3. **Implement authentication** if needed
4. **Configure firewall rules**
5. **Set up DDoS protection**
6. **Use production SSL certificates**
7. **Monitor and log all access**
8. **Implement request rate limiting per user (not just per IP)**
9. **Disable metrics endpoint or restrict to internal networks**
10. **Regular security audits**

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Internet / Browser / CLI                  │
└───────────────────────────┬─────────────────────────────────┘
                            │
                ┌───────────▼──────────┐
                │  Nginx Reverse Proxy  │ :80, :9090
                │ (aura-testnet-proxy)  │
                │                       │
                │  • Rate Limiting      │
                │  • CORS Headers       │
                │  • Security Headers   │
                │  • WebSocket Support  │
                │  • Load Balancing     │
                └───────────┬──────────┘
                            │
            ┌───────────────┴───────────────┐
            │   Docker Network              │
            │   aura_aura-testnet           │
            │   (172.26.0.0/16)             │
            │                               │
            │  ┌─────────────────────────┐  │
            │  │   Validator-1           │  │ :26657 (RPC)
            │  │   172.26.0.10           │  │ :1317  (API)
            │  │   (Primary)             │  │ :9090  (gRPC)
            │  └─────────────────────────┘  │
            │                               │
            │  ┌─────────────────────────┐  │
            │  │   Validator-2           │  │ :26657 (RPC)
            │  │   172.26.0.11           │  │ :1317  (API)
            │  │   (Backup - optional)   │  │ :9090  (gRPC)
            │  └─────────────────────────┘  │
            │                               │
            │  ┌─────────────────────────┐  │
            │  │   Validator-3           │  │
            │  │   172.26.0.12           │  │
            │  │   (Backup - optional)   │  │
            │  └─────────────────────────┘  │
            │                               │
            │  ┌─────────────────────────┐  │
            │  │   Validator-4           │  │
            │  │   172.26.0.13           │  │
            │  │   (Backup - optional)   │  │
            │  └─────────────────────────┘  │
            │                               │
            └───────────────────────────────┘
```

## Next Steps

After setting up the proxy:

1. **Test all endpoints** to ensure they work correctly
2. **Monitor logs** for errors or rate limit warnings
3. **Configure load balancing** if you need high availability
4. **Set up SSL/TLS** for production deployments
5. **Implement monitoring and alerting**
6. **Document your specific use cases** for team members

## Support

For issues or questions:
- Check the detailed documentation in `docs/TESTNET_PUBLIC_ENDPOINTS.md`
- Review nginx logs for errors
- Ensure testnet validators are healthy and producing blocks
- Check Docker network connectivity

---

**Last Updated**: 2025-12-10
