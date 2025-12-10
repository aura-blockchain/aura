# Aura Testnet Public Endpoints

This document describes the public RPC and API endpoints for the Aura testnet, configured using an Nginx reverse proxy.

## Overview

The public endpoint proxy provides:
- **HTTP RPC endpoints** for blockchain queries and transaction submission
- **REST API endpoints** for Cosmos SDK modules
- **gRPC endpoints** for native gRPC clients
- **WebSocket support** for real-time block and event subscriptions
- **CORS headers** for browser-based applications
- **Rate limiting** to prevent abuse
- **Security headers** and connection limits

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Internet / Users                        │
└───────────────────────────┬─────────────────────────────────┘
                            │
                ┌───────────▼───────────┐
                │   Nginx Reverse Proxy  │
                │   (aura-testnet-proxy) │
                │                        │
                │  - Rate Limiting       │
                │  - CORS Headers        │
                │  - Security Headers    │
                │  - Load Balancing      │
                └───────────┬───────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
    ┌───▼────┐         ┌────▼───┐         ┌────▼───┐
    │ Val-1  │         │ Val-2  │         │ Val-3  │
    │ Primary│         │ Backup │         │ Backup │
    └────────┘         └────────┘         └────────┘
     (Active)          (Available)       (Available)
```

## Prerequisites

1. **Testnet must be running**:
   ```bash
   docker-compose -f docker-compose.testnet.yml up -d
   ```

2. **Validators must be healthy**:
   ```bash
   docker-compose -f docker-compose.testnet.yml ps
   # All validators should show "healthy" status
   ```

3. **Check validator connectivity**:
   ```bash
   # Test validator-1 RPC directly
   curl http://localhost:27657/status

   # Test validator-1 API directly
   curl http://localhost:2317/cosmos/base/tendermint/v1beta1/blocks/latest
   ```

## Setup Instructions

### Step 1: Start the Proxy

```bash
cd /home/decri/blockchain-projects/aura
docker-compose -f docker-compose.proxy.yml up -d
```

### Step 2: Verify Proxy Health

```bash
# Check proxy container status
docker-compose -f docker-compose.proxy.yml ps

# Check health endpoint
curl http://localhost/health
# Expected: {"status":"healthy","service":"aura-testnet-proxy","version":"1.0.0"}

# Check root endpoint for service info
curl http://localhost/
# Expected: JSON with endpoint information
```

### Step 3: Test Endpoints

**Test RPC endpoint:**
```bash
curl http://localhost/rpc/status | jq
```

**Test API endpoint:**
```bash
curl http://localhost/api/cosmos/base/tendermint/v1beta1/blocks/latest | jq
```

**Test gRPC endpoint (requires grpcurl):**
```bash
grpcurl -plaintext localhost:9090 list
```

## Available Endpoints

### HTTP Endpoints (Port 80)

| Endpoint | Description | Example |
|----------|-------------|---------|
| `http://localhost/` | Service information | `curl http://localhost/` |
| `http://localhost/health` | Health check | `curl http://localhost/health` |
| `http://localhost/rpc/` | Tendermint RPC | `curl http://localhost/rpc/status` |
| `http://localhost/api/` | Cosmos REST API | `curl http://localhost/api/cosmos/bank/v1beta1/supply` |
| `http://localhost/api/swagger/` | API documentation | Open in browser |
| `http://localhost/metrics` | Prometheus metrics (private) | Only from local networks |

### gRPC Endpoint (Port 9090)

| Endpoint | Protocol | Example |
|----------|----------|---------|
| `localhost:9090` | gRPC (HTTP/2) | `grpcurl -plaintext localhost:9090 list` |

### WebSocket Endpoints

| Endpoint | Description | Example |
|----------|-------------|---------|
| `ws://localhost/rpc/websocket` | Tendermint event subscriptions | See WebSocket section below |

## Usage Examples

### RPC Examples

**Get node status:**
```bash
curl http://localhost/rpc/status | jq
```

**Get latest block:**
```bash
curl http://localhost/rpc/block | jq
```

**Get blockchain info:**
```bash
curl http://localhost/rpc/blockchain?minHeight=1&maxHeight=100 | jq
```

**Get validators:**
```bash
curl http://localhost/rpc/validators | jq
```

**Get network info:**
```bash
curl http://localhost/rpc/net_info | jq
```

**Get genesis:**
```bash
curl http://localhost/rpc/genesis | jq
```

### API Examples

**Get latest block (REST):**
```bash
curl http://localhost/api/cosmos/base/tendermint/v1beta1/blocks/latest | jq
```

**Get node info:**
```bash
curl http://localhost/api/cosmos/base/tendermint/v1beta1/node_info | jq
```

**Get validators:**
```bash
curl http://localhost/api/cosmos/base/tendermint/v1beta1/validatorsets/latest | jq
```

**Get total supply:**
```bash
curl http://localhost/api/cosmos/bank/v1beta1/supply | jq
```

**Get account balance:**
```bash
curl http://localhost/api/cosmos/bank/v1beta1/balances/aura1... | jq
```

**Get staking parameters:**
```bash
curl http://localhost/api/cosmos/staking/v1beta1/params | jq
```

**Get all staking validators:**
```bash
curl http://localhost/api/cosmos/staking/v1beta1/validators | jq
```

### gRPC Examples (requires grpcurl)

**List all services:**
```bash
grpcurl -plaintext localhost:9090 list
```

**Get node info:**
```bash
grpcurl -plaintext localhost:9090 cosmos.base.tendermint.v1beta1.Service/GetNodeInfo
```

**Get latest block:**
```bash
grpcurl -plaintext localhost:9090 cosmos.base.tendermint.v1beta1.Service/GetLatestBlock
```

**Get account balance:**
```bash
grpcurl -plaintext \
  -d '{"address":"aura1...","denom":"uaura"}' \
  localhost:9090 \
  cosmos.bank.v1beta1.Query/Balance
```

### WebSocket Examples (requires wscat or similar)

**Install wscat:**
```bash
npm install -g wscat
```

**Subscribe to new blocks:**
```bash
wscat -c ws://localhost/rpc/websocket
# After connection, send:
{"jsonrpc":"2.0","method":"subscribe","params":["tm.event='NewBlock'"],"id":1}
```

**Subscribe to new transactions:**
```bash
wscat -c ws://localhost/rpc/websocket
# After connection, send:
{"jsonrpc":"2.0","method":"subscribe","params":["tm.event='Tx'"],"id":1}
```

**Subscribe to validator set updates:**
```bash
wscat -c ws://localhost/rpc/websocket
# After connection, send:
{"jsonrpc":"2.0","method":"subscribe","params":["tm.event='ValidatorSetUpdates'"],"id":1}
```

## Browser-Based Applications

The proxy includes CORS headers to support browser-based applications:

```javascript
// Example: Fetch latest block from browser
fetch('http://localhost/api/cosmos/base/tendermint/v1beta1/blocks/latest')
  .then(response => response.json())
  .then(data => console.log('Latest block:', data))
  .catch(error => console.error('Error:', error));

// Example: WebSocket connection from browser
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

## Rate Limits

The proxy enforces the following rate limits per IP address:

| Endpoint Type | Rate Limit | Burst Allowance |
|---------------|------------|-----------------|
| RPC endpoints | 30 req/s | 20 additional |
| API endpoints | 50 req/s | 30 additional |
| gRPC endpoints | 100 req/s | 50 additional |
| Connection limit | 100 concurrent | N/A |
| Max body size | 2 MB | N/A |

If you exceed these limits, you'll receive HTTP 503 (Service Temporarily Unavailable) responses.

### Rate Limit Response

```json
{
  "error": "503 Service Temporarily Unavailable"
}
```

### Adjusting Rate Limits

Edit `/home/decri/blockchain-projects/aura/nginx/testnet-proxy.conf`:

```nginx
# Increase RPC rate limit from 30r/s to 100r/s
limit_req_zone $binary_remote_addr zone=testnet_rpc_limit:10m rate=100r/s;

# Increase connection limit from 100 to 500
limit_conn testnet_conn_limit 500;
```

Then reload the proxy:
```bash
docker-compose -f docker-compose.proxy.yml restart
```

## Security Considerations

### CORS Configuration

The proxy allows all origins (`Access-Control-Allow-Origin: *`) for development. For production:

1. Edit `nginx/testnet-proxy.conf`
2. Change `add_header Access-Control-Allow-Origin "*"` to specific domains
3. Restart the proxy

### Metrics Endpoint

The `/metrics` endpoint is restricted to private networks only:
- `10.0.0.0/8`
- `172.16.0.0/12`
- `192.168.0.0/16`
- `127.0.0.1`

Public access is denied for security.

### SSL/TLS

For production deployments with SSL/TLS:

1. Obtain SSL certificates (Let's Encrypt recommended)
2. Update `nginx/testnet-proxy.conf` to enable SSL
3. Uncomment SSL-related sections and configure certificate paths
4. Update `docker-compose.proxy.yml` to mount certificate directory
5. Change port 80 listener to 443 with SSL

## Load Balancing

The current configuration uses **validator-1** as the primary backend. To enable load balancing across all validators:

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

Then restart:
```bash
docker-compose -f docker-compose.proxy.yml restart
```

## Monitoring

### View Proxy Logs

```bash
# All logs
docker-compose -f docker-compose.proxy.yml logs -f

# Access logs only
docker exec aura-testnet-proxy tail -f /var/log/nginx/testnet-access.log

# Error logs only
docker exec aura-testnet-proxy tail -f /var/log/nginx/testnet-error.log

# gRPC logs
docker exec aura-testnet-proxy tail -f /var/log/nginx/testnet-grpc-access.log
```

### Check Connection Stats

```bash
# Get nginx status
docker exec aura-testnet-proxy nginx -s status

# Check active connections
docker exec aura-testnet-proxy ps aux | grep nginx
```

### Prometheus Metrics

Access validator metrics (from local network only):
```bash
curl http://localhost/metrics
```

## Troubleshooting

### Proxy won't start

**Error: "network aura_aura-testnet not found"**

Solution: Start the testnet first:
```bash
docker-compose -f docker-compose.testnet.yml up -d
```

**Error: "port 80 already in use"**

Solution: Stop conflicting service or change proxy port in `docker-compose.proxy.yml`.

### Connection refused errors

**Check if validators are running:**
```bash
docker-compose -f docker-compose.testnet.yml ps
```

**Check if proxy can reach validators:**
```bash
docker exec aura-testnet-proxy wget -O- http://validator-1:26657/status
```

### Rate limit errors (HTTP 503)

Requests are being rate limited. Either:
1. Reduce request frequency
2. Increase rate limits in `nginx/testnet-proxy.conf`
3. Implement request batching

### CORS errors in browser

Check browser console for specific CORS error. Ensure:
1. Using correct endpoint URL
2. CORS headers are enabled in nginx config
3. Request method is GET or POST
4. Required headers are allowed

### WebSocket connection fails

**Check WebSocket upgrade:**
```bash
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" http://localhost/rpc/websocket
```

Should return HTTP 101 Switching Protocols.

## Stopping the Proxy

```bash
docker-compose -f docker-compose.proxy.yml down
```

To remove volumes (logs):
```bash
docker-compose -f docker-compose.proxy.yml down -v
```

## Production Deployment

For production deployments:

1. **Enable SSL/TLS** (required for secure connections)
2. **Restrict CORS** to specific domains
3. **Increase resource limits** in docker-compose
4. **Enable load balancing** across validators
5. **Configure firewall rules**
6. **Set up monitoring and alerting**
7. **Implement request authentication** if needed
8. **Use production-grade SSL certificates** (not self-signed)
9. **Configure DDoS protection**
10. **Set up log aggregation and analysis**

## Additional Resources

- [Tendermint RPC Documentation](https://docs.tendermint.com/master/rpc/)
- [Cosmos SDK REST API](https://docs.cosmos.network/main/run-node/interact-node#using-the-rest-api)
- [Nginx Configuration](https://nginx.org/en/docs/)
- [CORS Documentation](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [gRPC Documentation](https://grpc.io/docs/)

## Contact

For issues or questions, refer to the main project documentation or open an issue on GitHub.
