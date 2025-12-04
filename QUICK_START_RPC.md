# Quick Start Guide - Aura Public RPC Endpoints

## TL;DR - Start in 30 Seconds

```bash
# Verify configuration
./scripts/verify-rpc-configuration.sh

# Start the RPC node stack
docker compose -f docker/docker-compose.rpc.yml up -d

# Wait 30 seconds for initialization, then test
sleep 30
./scripts/test-rpc-endpoints.sh

# Access services
# - RPC: http://localhost:26657
# - API: http://localhost:1317
# - Grafana: http://localhost:3001 (admin/admin)
# - Prometheus: http://localhost:9092
```

## What You Get

### Endpoints
- ✅ Tendermint RPC on port 26657
- ✅ Cosmos REST API on port 1317
- ✅ gRPC on port 9090
- ✅ gRPC-Web on port 9091
- ✅ Prometheus metrics on port 26660
- ✅ Nginx reverse proxy on ports 80/443

### Services
- ✅ Aura blockchain node
- ✅ Nginx with SSL/TLS and rate limiting
- ✅ Prometheus metrics collection
- ✅ Grafana dashboards
- ✅ Node exporter for system metrics

### Security
- ✅ Rate limiting (30-100 req/s depending on endpoint)
- ✅ SSL/TLS encryption (self-signed certs)
- ✅ CORS enabled for browser access
- ✅ Security headers configured
- ✅ Unsafe RPC methods disabled

## Quick Tests

### RPC Status
```bash
curl http://localhost:26657/status | jq
```

### Latest Block
```bash
curl http://localhost:1317/cosmos/base/tendermint/v1beta1/blocks/latest | jq
```

### Health Check
```bash
curl http://localhost/health
```

### Metrics
```bash
curl http://localhost:26660/metrics
```

## View Logs

```bash
# All services
docker compose -f docker/docker-compose.rpc.yml logs -f

# Just the node
docker compose -f docker/docker-compose.rpc.yml logs -f aura-rpc-node

# Just nginx
docker compose -f docker/docker-compose.rpc.yml logs -f nginx-rpc-proxy
```

## Stop Services

```bash
docker compose -f docker/docker-compose.rpc.yml down
```

## Full Documentation

See `/docs/ops/PUBLIC_RPC_ENDPOINTS.md` for complete documentation.

## Troubleshooting

**Port already in use?**
```bash
# Check what's using the port
sudo lsof -i :26657

# Stop the service or change the port in docker-compose.rpc.yml
```

**Docker compose not found?**
```bash
# Use newer syntax
docker compose -f docker/docker-compose.rpc.yml up -d
```

**Node won't start?**
```bash
# Check logs
docker compose -f docker/docker-compose.rpc.yml logs aura-rpc-node

# Common fix: Remove old data and restart
docker compose -f docker/docker-compose.rpc.yml down -v
docker compose -f docker/docker-compose.rpc.yml up -d
```

## Production Deployment

For production, see:
- Kubernetes: `/k8s/rpc-node-deployment.yaml`
- Full docs: `/docs/ops/PUBLIC_RPC_ENDPOINTS.md`
- Summary: `/RPC_ENDPOINT_DEPLOYMENT_SUMMARY.md`
