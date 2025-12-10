# Nginx Configurations for Aura

This directory contains nginx reverse proxy configurations for various Aura deployments.

## Configuration Files

### `testnet-proxy.conf`

Public endpoint proxy for the local testnet (4 validators).

**Features:**
- HTTP RPC endpoint at `/rpc`
- REST API endpoint at `/api`
- gRPC endpoint on port 9090
- WebSocket support for subscriptions
- CORS headers for browser clients
- Rate limiting per IP
- Security headers

**Usage:**
```bash
# Start with docker-compose
docker-compose -f docker-compose.proxy.yml up -d

# Or use the helper script
./scripts/start-proxy.sh
```

**Endpoints:**
- RPC: `http://localhost/rpc`
- API: `http://localhost/api`
- gRPC: `localhost:9090`
- Swagger: `http://localhost/api/swagger/`
- Health: `http://localhost/health`

### `rpc-proxy.conf`

Production-ready RPC proxy configuration with SSL/TLS support.

**Features:**
- HTTPS with TLS 1.2/1.3
- SSL certificate configuration
- HTTP to HTTPS redirect
- gRPC over TLS
- gRPC-Web for browser clients
- OCSP stapling
- Security headers (HSTS, etc.)

**Usage:**
```bash
# Start with docker-compose
docker-compose -f docker/docker-compose.rpc.yml up -d
```

**Note:** Requires SSL certificates in `nginx/ssl/` directory.

### `ssl-config.conf`

Shared SSL/TLS configuration snippets for nginx.

## SSL Certificate Setup

For production deployments with SSL/TLS:

### Using Let's Encrypt

```bash
# Install certbot
sudo apt-get update
sudo apt-get install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot certonly --nginx -d rpc.testnet.aura.network

# Certificates will be placed in:
# /etc/letsencrypt/live/rpc.testnet.aura.network/
```

### Using Self-Signed Certificates (Development Only)

```bash
# Generate self-signed certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout nginx/ssl/aura-testnet.key \
  -out nginx/ssl/aura-testnet.crt \
  -subj "/C=US/ST=State/L=City/O=Aura/CN=localhost"

# Set proper permissions
chmod 600 nginx/ssl/aura-testnet.key
chmod 644 nginx/ssl/aura-testnet.crt
```

## Testing Configurations

### Test nginx syntax

```bash
# Test configuration file syntax
docker run --rm -v $(pwd)/nginx:/etc/nginx/conf.d:ro nginx:1.25-alpine nginx -t
```

### Test with curl

```bash
# Test RPC endpoint
curl http://localhost/rpc/status

# Test API endpoint
curl http://localhost/api/cosmos/base/tendermint/v1beta1/node_info

# Test health check
curl http://localhost/health

# Test CORS headers
curl -H "Origin: http://example.com" -I http://localhost/rpc/status

# Test WebSocket upgrade
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  http://localhost/rpc/websocket
```

### Test gRPC endpoint

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:9090 list

# Get node info
grpcurl -plaintext localhost:9090 \
  cosmos.base.tendermint.v1beta1.Service/GetNodeInfo
```

## Customization

### Adjusting Rate Limits

Edit the rate limit zones at the top of the configuration:

```nginx
# Increase RPC rate limit from 30r/s to 100r/s
limit_req_zone $binary_remote_addr zone=testnet_rpc_limit:10m rate=100r/s;

# Increase API rate limit from 50r/s to 200r/s
limit_req_zone $binary_remote_addr zone=testnet_api_limit:10m rate=200r/s;

# Increase connection limit from 100 to 500
limit_conn testnet_conn_limit 500;
```

### Enabling Load Balancing

Uncomment backup validators in the upstream blocks:

```nginx
upstream testnet_rpc_backend {
    server validator-1:26657 max_fails=3 fail_timeout=30s weight=10;
    server validator-2:26657 max_fails=3 fail_timeout=30s weight=5;  # Uncomment
    server validator-3:26657 max_fails=3 fail_timeout=30s weight=5;  # Uncomment
    server validator-4:26657 max_fails=3 fail_timeout=30s weight=5;  # Uncomment
    keepalive 32;
}
```

### Restricting CORS

Change the CORS headers to allow only specific origins:

```nginx
# Instead of:
add_header Access-Control-Allow-Origin "*" always;

# Use:
add_header Access-Control-Allow-Origin "https://yourdomain.com" always;
```

### Adding IP Whitelist

Add IP restrictions to specific endpoints:

```nginx
location /rpc {
    # Only allow specific IPs
    allow 1.2.3.4;
    allow 5.6.7.8;
    deny all;

    # Rest of configuration...
}
```

## Monitoring

### View Access Logs

```bash
# Real-time access logs
docker exec aura-testnet-proxy tail -f /var/log/nginx/testnet-access.log

# Error logs
docker exec aura-testnet-proxy tail -f /var/log/nginx/testnet-error.log

# gRPC access logs
docker exec aura-testnet-proxy tail -f /var/log/nginx/testnet-grpc-access.log
```

### Check Nginx Status

```bash
# Get nginx process status
docker exec aura-testnet-proxy ps aux | grep nginx

# Check configuration
docker exec aura-testnet-proxy nginx -t

# Reload configuration without downtime
docker exec aura-testnet-proxy nginx -s reload
```

## Troubleshooting

### Configuration Errors

**Syntax error:**
```bash
# Test configuration
docker exec aura-testnet-proxy nginx -t
```

**Reload failed:**
```bash
# Check error logs
docker exec aura-testnet-proxy cat /var/log/nginx/testnet-error.log
```

### Connection Issues

**502 Bad Gateway:**
- Check if validators are running and healthy
- Verify network connectivity between proxy and validators
- Check upstream configuration

**503 Service Unavailable:**
- Rate limit exceeded
- Too many connections
- Increase limits or reduce request frequency

**CORS errors:**
- Check CORS headers are present: `curl -I http://localhost/rpc/status`
- Verify origin is allowed in configuration
- Check browser console for specific error

### Performance Issues

**High latency:**
- Enable keepalive connections
- Increase worker processes
- Add more validators to upstream

**High memory usage:**
- Reduce buffer sizes
- Decrease connection limits
- Enable caching for read-only endpoints

## Documentation

For detailed information, see:
- [Testnet Public Endpoints Documentation](../docs/TESTNET_PUBLIC_ENDPOINTS.md)
- [Nginx Documentation](https://nginx.org/en/docs/)
- [Tendermint RPC API](https://docs.tendermint.com/master/rpc/)
- [Cosmos SDK REST API](https://docs.cosmos.network/main/run-node/interact-node)
