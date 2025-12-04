# Aura Testnet Public RPC Endpoints

This document provides comprehensive information about the Aura blockchain testnet public RPC endpoints, including configuration, deployment, monitoring, and usage.

## Table of Contents

- [Overview](#overview)
- [Endpoint URLs](#endpoint-urls)
- [Configuration](#configuration)
- [Deployment Options](#deployment-options)
- [Security Measures](#security-measures)
- [Rate Limiting](#rate-limiting)
- [Monitoring](#monitoring)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)

---

## Overview

The Aura testnet provides public RPC endpoints to enable developers and users to interact with the blockchain without running their own nodes. These endpoints are production-grade, highly available, and secured with industry-standard practices.

### Supported Protocols

- **Tendermint RPC** (HTTP/WebSocket) - Port 26657
- **Cosmos REST API** (HTTP) - Port 1317
- **gRPC** (HTTP/2) - Port 9090
- **gRPC-Web** (HTTP/1.1 and HTTP/2) - Port 9091

### Key Features

- ✅ SSL/TLS encryption for all connections
- ✅ CORS support for browser-based applications
- ✅ Rate limiting to prevent abuse
- ✅ High availability with horizontal scaling
- ✅ Comprehensive monitoring and alerting
- ✅ WebSocket support for real-time event streaming
- ✅ Prometheus metrics exposure
- ✅ Automatic failover and health checks

---

## Endpoint URLs

### Production Endpoints (HTTPS)

```
Base URL: https://rpc.testnet.aura.network
```

| Service | Endpoint | Port | Protocol |
|---------|----------|------|----------|
| Tendermint RPC | `https://rpc.testnet.aura.network/rpc/` | 443 | HTTP/HTTPS, WebSocket |
| Cosmos REST API | `https://rpc.testnet.aura.network/api/` | 443 | HTTP/HTTPS |
| gRPC | `rpc.testnet.aura.network:9090` | 9090 | gRPC over TLS |
| gRPC-Web | `rpc.testnet.aura.network:9091` | 9091 | gRPC-Web over TLS |
| API Documentation | `https://rpc.testnet.aura.network/swagger/` | 443 | HTTP/HTTPS |
| Health Check | `https://rpc.testnet.aura.network/health` | 443 | HTTP/HTTPS |

### Local Development (HTTP)

When running locally via Docker Compose:

| Service | Endpoint | Port | Protocol |
|---------|----------|------|----------|
| Tendermint RPC | `http://localhost:26657` | 26657 | HTTP, WebSocket |
| Cosmos REST API | `http://localhost:1317` | 1317 | HTTP |
| gRPC | `localhost:9090` | 9090 | gRPC |
| gRPC-Web | `localhost:9091` | 9091 | gRPC-Web |
| Prometheus Metrics | `http://localhost:26660/metrics` | 26660 | HTTP |

---

## Configuration

### Node Configuration Files

Configuration files are located in `/networks/testnet/`:

- **config.toml** - Tendermint configuration (RPC, P2P, consensus)
- **app.toml** - Cosmos SDK application configuration (API, gRPC, modules)

### Key Configuration Settings

#### RPC Configuration (config.toml)

```toml
[rpc]
laddr = "tcp://0.0.0.0:26657"
cors_allowed_origins = ["*"]
cors_allowed_methods = ["HEAD", "GET", "POST"]
max_open_connections = 1000
max_subscription_clients = 100
```

#### API Configuration (app.toml)

```toml
[api]
enable = true
address = "tcp://0.0.0.0:1317"
max-open-connections = 2000
enabled-unsafe-cors = true
swagger = true
```

#### gRPC Configuration (app.toml)

```toml
[grpc]
enable = true
address = "0.0.0.0:9090"

[grpc-web]
enable = true
address = "0.0.0.0:9091"
enable-unsafe-cors = true
```

---

## Deployment Options

### Option 1: Docker Compose (Recommended for Testing)

Deploy using the provided Docker Compose configuration:

```bash
cd /home/decri/blockchain-projects/aura/docker

# Start all services (RPC node, nginx, prometheus, grafana)
docker-compose -f docker-compose.rpc.yml up -d

# View logs
docker-compose -f docker-compose.rpc.yml logs -f

# Stop services
docker-compose -f docker-compose.rpc.yml down
```

**Services deployed:**
- `aura-rpc-node` - Aura blockchain node
- `nginx-rpc-proxy` - Nginx reverse proxy with SSL/TLS
- `prometheus` - Metrics collection
- `grafana` - Metrics visualization (http://localhost:3001)
- `node-exporter` - System metrics

### Option 2: Kubernetes (Recommended for Production)

Deploy using Kubernetes manifests:

```bash
# Create namespace
kubectl create namespace aura-testnet

# Create TLS secret
kubectl create secret tls aura-rpc-tls \
  --cert=/home/decri/blockchain-projects/aura/nginx/ssl/aura-testnet.crt \
  --key=/home/decri/blockchain-projects/aura/nginx/ssl/aura-testnet.key \
  -n aura-testnet

# Deploy RPC nodes
kubectl apply -f /home/decri/blockchain-projects/aura/k8s/rpc-node-deployment.yaml

# Check deployment status
kubectl get pods -n aura-testnet
kubectl get svc -n aura-testnet

# View logs
kubectl logs -f -n aura-testnet -l app=aura-rpc-node
```

**Components deployed:**
- StatefulSet with 2 replicas (auto-scaling 2-10)
- LoadBalancer service for external access
- ConfigMaps for configuration
- PersistentVolumeClaims for data storage
- HorizontalPodAutoscaler for auto-scaling
- PodDisruptionBudget for high availability

### Option 3: Standalone (Manual)

Run the node directly:

```bash
cd /home/decri/blockchain-projects/aura/chain

# Build the binary
go build -o aurad ./cmd/aurad

# Initialize the node
./aurad init rpc-node --chain-id aura-testnet-1

# Copy configuration files
cp /home/decri/blockchain-projects/aura/networks/testnet/config.toml ~/.aura/config/
cp /home/decri/blockchain-projects/aura/networks/testnet/app.toml ~/.aura/config/

# Start the node
./aurad start
```

---

## Security Measures

### 1. SSL/TLS Encryption

All public endpoints use TLS 1.2+ with strong cipher suites:

```nginx
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:...';
ssl_prefer_server_ciphers off;
```

**Certificate Management:**
- Testnet: Self-signed certificates (generated via `scripts/generate-ssl-certs.sh`)
- Production: Use Let's Encrypt or commercial CA certificates

### 2. Rate Limiting

Nginx implements multiple rate limiting zones:

| Zone | Limit | Burst | Applies To |
|------|-------|-------|------------|
| `rpc_limit` | 30 req/s | 20 | Tendermint RPC endpoints |
| `api_limit` | 50 req/s | 30 | Cosmos REST API endpoints |
| `grpc_limit` | 100 req/s | 50 | gRPC endpoints |
| `conn_limit` | 100 connections | - | All endpoints (per IP) |

### 3. CORS Configuration

CORS headers allow browser-based applications to access the endpoints:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: DNT, User-Agent, X-Requested-With, ...
```

### 4. Request Size Limits

```nginx
client_max_body_size 2M;
client_body_buffer_size 128k;
```

### 5. Security Headers

```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
X-XSS-Protection: 1; mode=block
```

### 6. Access Control

- Unsafe RPC methods disabled (`unsafe = false`)
- WASM code uploads restricted on RPC nodes (`code_upload_access = "nobody"`)
- Prometheus metrics restricted to internal networks only

---

## Rate Limiting

### Per-IP Rate Limits

Rate limiting is enforced at the Nginx layer based on client IP addresses:

**Tendermint RPC:**
- Rate: 30 requests/second
- Burst: 20 additional requests
- Behavior: Delay (nodelay mode)

**Cosmos REST API:**
- Rate: 50 requests/second
- Burst: 30 additional requests
- Behavior: Delay (nodelay mode)

**gRPC:**
- Rate: 100 requests/second
- Burst: 50 additional requests
- Behavior: Delay (nodelay mode)

**Connection Limit:**
- Maximum: 100 concurrent connections per IP

### Rate Limit Response

When rate limit is exceeded, nginx returns:
- HTTP Status: `503 Service Temporarily Unavailable`
- Header: `Retry-After: <seconds>`

### Bypassing Rate Limits (for Trusted IPs)

To whitelist specific IPs, modify `/nginx/rpc-proxy.conf`:

```nginx
geo $limit {
    default 1;
    10.0.0.0/8 0;       # Internal network
    192.168.1.100 0;    # Trusted IP
}

map $limit $limit_key {
    0 "";
    1 $binary_remote_addr;
}

limit_req_zone $limit_key zone=rpc_limit:10m rate=30r/s;
```

---

## Monitoring

### Prometheus Metrics

Metrics are exposed on port 26660 (restricted to internal access):

```bash
# Access metrics endpoint (local only)
curl http://localhost:26660/metrics
```

**Available Metrics:**

| Category | Metrics | Description |
|----------|---------|-------------|
| Tendermint | `tendermint_consensus_*` | Consensus state, voting, rounds |
| P2P | `tendermint_p2p_*` | Peer connections, message rates |
| Mempool | `tendermint_mempool_*` | Transaction pool size, cache |
| Blockchain | `tendermint_blockchain_*` | Block height, sync status |
| State | `tendermint_state_*` | State sync, snapshots |

### Grafana Dashboards

Access Grafana at http://localhost:3001 (when running via Docker Compose)

**Default credentials:**
- Username: `admin`
- Password: `admin`

**Pre-configured dashboards:**
- Aura RPC Node Overview
- Tendermint Consensus
- API/RPC Request Rates
- System Resources

### Health Checks

**Endpoint Health:**
```bash
# Nginx health check
curl http://localhost/health
# Response: healthy

# Node health check
curl http://localhost:26657/health
# Response: {"result":{}}
```

**Kubernetes Probes:**
- Liveness Probe: `/health` endpoint
- Readiness Probe: `/status` endpoint

### Alerting (Optional)

Configure Alertmanager for production deployments:

```yaml
# prometheus-rpc.yml
alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

**Recommended Alerts:**
- Node down/unreachable
- High RPC error rate
- Memory/CPU usage exceeding thresholds
- P2P peer count below minimum
- Block production stalled

---

## Usage Examples

### Tendermint RPC

#### Get Node Status
```bash
# HTTP
curl https://rpc.testnet.aura.network/rpc/status

# Local
curl http://localhost:26657/status
```

#### Get Latest Block
```bash
curl https://rpc.testnet.aura.network/rpc/block
```

#### Query Transaction
```bash
curl https://rpc.testnet.aura.network/rpc/tx?hash=0x...
```

#### WebSocket Subscription
```javascript
const ws = new WebSocket('wss://rpc.testnet.aura.network/rpc/websocket');

ws.onopen = () => {
  ws.send(JSON.stringify({
    jsonrpc: "2.0",
    method: "subscribe",
    id: "1",
    params: {
      query: "tm.event='NewBlock'"
    }
  }));
};

ws.onmessage = (event) => {
  console.log('New block:', JSON.parse(event.data));
};
```

### Cosmos REST API

#### Get Account Balance
```bash
curl https://rpc.testnet.aura.network/api/cosmos/bank/v1beta1/balances/{address}
```

#### Get Latest Block
```bash
curl https://rpc.testnet.aura.network/api/cosmos/base/tendermint/v1beta1/blocks/latest
```

#### Query Module Params
```bash
curl https://rpc.testnet.aura.network/api/cosmos/staking/v1beta1/params
```

#### Swagger Documentation
Open in browser: https://rpc.testnet.aura.network/swagger/

### gRPC

#### Using grpcurl

**List available services:**
```bash
grpcurl -plaintext localhost:9090 list
```

**Query account:**
```bash
grpcurl -plaintext \
  -d '{"address": "aura1..."}' \
  localhost:9090 \
  cosmos.bank.v1beta1.Query/Balance
```

**Get node info:**
```bash
grpcurl -plaintext localhost:9090 cosmos.base.tendermint.v1beta1.Service/GetNodeInfo
```

#### Using Go Client

```go
import (
    "context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Connect to gRPC endpoint
conn, err := grpc.Dial(
    "localhost:9090",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
)
if err != nil {
    panic(err)
}
defer conn.Close()

// Create bank query client
bankClient := banktypes.NewQueryClient(conn)

// Query balance
res, err := bankClient.Balance(context.Background(), &banktypes.QueryBalanceRequest{
    Address: "aura1...",
    Denom:   "uaura",
})
```

### gRPC-Web (Browser)

#### Using @improbable-eng/grpc-web

```javascript
import { grpc } from "@improbable-eng/grpc-web";

const request = new QueryBalanceRequest();
request.setAddress("aura1...");
request.setDenom("uaura");

grpc.invoke(Query.Balance, {
  request: request,
  host: "https://rpc.testnet.aura.network:9091",
  onMessage: (message) => {
    console.log("Balance:", message.toObject());
  },
  onEnd: (code, msg, trailers) => {
    if (code === grpc.Code.OK) {
      console.log("Request completed successfully");
    }
  }
});
```

---

## Troubleshooting

### Common Issues

#### 1. Connection Refused

**Symptom:**
```
curl: (7) Failed to connect to localhost port 26657: Connection refused
```

**Solutions:**
- Check if the node is running: `docker ps` or `kubectl get pods`
- Verify port mappings in docker-compose or Kubernetes service
- Check firewall rules: `sudo ufw status`

#### 2. Rate Limit Exceeded

**Symptom:**
```
HTTP/1.1 503 Service Temporarily Unavailable
Retry-After: 1
```

**Solutions:**
- Reduce request rate
- Implement exponential backoff
- Contact administrators for rate limit increase
- Use multiple IPs if legitimate use case

#### 3. CORS Errors in Browser

**Symptom:**
```
Access to fetch at 'https://rpc.testnet.aura.network/api/...' from origin 'http://localhost:3000' has been blocked by CORS policy
```

**Solutions:**
- Ensure nginx CORS headers are configured
- Check browser console for specific CORS error
- Verify OPTIONS preflight requests are handled

#### 4. SSL Certificate Error

**Symptom:**
```
SSL certificate problem: self signed certificate
```

**Solutions:**
- For testing: Use `-k` flag with curl: `curl -k https://...`
- For browsers: Add certificate to trusted roots (see certificate generation script output)
- For production: Use proper CA-signed certificates

#### 5. gRPC Connection Issues

**Symptom:**
```
rpc error: code = Unavailable desc = connection error
```

**Solutions:**
- Verify gRPC port is accessible: `telnet localhost 9090`
- Check if gRPC is enabled in app.toml
- For TLS: Ensure proper certificates are configured
- For grpcurl: Use `-plaintext` for non-TLS connections

### Diagnostic Commands

```bash
# Check node is syncing
curl http://localhost:26657/status | jq '.result.sync_info'

# Check peer connections
curl http://localhost:26657/net_info | jq '.result.n_peers'

# View recent logs
docker logs -f --tail=100 aura-rpc-node

# Check nginx access logs
docker logs -f --tail=100 aura-nginx-proxy

# Test rate limiting
ab -n 100 -c 10 http://localhost/rpc/status

# Verify CORS headers
curl -H "Origin: http://example.com" -I https://rpc.testnet.aura.network/api/
```

### Getting Help

- **Documentation:** https://docs.aura.network
- **GitHub Issues:** https://github.com/aura/aura/issues
- **Discord:** https://discord.gg/aura
- **Telegram:** https://t.me/auranetwork

---

## Maintenance

### Certificate Renewal

Self-signed certificates expire after 365 days. Renew using:

```bash
/home/decri/blockchain-projects/aura/scripts/generate-ssl-certs.sh

# Restart nginx to load new certificates
docker-compose -f docker/docker-compose.rpc.yml restart nginx-rpc-proxy
# OR
kubectl rollout restart deployment/aura-rpc-node -n aura-testnet
```

### Updating Configuration

1. Edit configuration files in `/networks/testnet/`
2. Restart services:

```bash
# Docker Compose
docker-compose -f docker/docker-compose.rpc.yml restart aura-rpc-node

# Kubernetes
kubectl rollout restart deployment/aura-rpc-node -n aura-testnet
```

### Scaling

**Docker Compose:**
```bash
docker-compose -f docker/docker-compose.rpc.yml up -d --scale aura-rpc-node=3
```

**Kubernetes:**
```bash
kubectl scale deployment/aura-rpc-node --replicas=5 -n aura-testnet
```

---

## Production Checklist

Before deploying to production:

- [ ] Replace self-signed certificates with CA-signed certificates
- [ ] Configure proper DNS for RPC endpoint domain
- [ ] Set up monitoring and alerting
- [ ] Configure log aggregation (ELK, Loki, CloudWatch, etc.)
- [ ] Implement backup and disaster recovery procedures
- [ ] Review and adjust rate limits based on expected traffic
- [ ] Harden security (firewall rules, DDoS protection, etc.)
- [ ] Set up auto-scaling policies
- [ ] Configure persistent peers and seeds
- [ ] Test failover scenarios
- [ ] Document incident response procedures
- [ ] Schedule regular security audits

---

## License

This documentation is part of the Aura blockchain project.

---

**Last Updated:** 2025-12-03
**Version:** 1.0.0
**Maintained by:** Aura DevOps Team
