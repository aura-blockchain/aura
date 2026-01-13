# AURA Public Testnet Security Baseline

This document summarizes the expected security controls for the public AURA testnet (aura-mvp-1)
exposed via registered endpoints and a reverse proxy.

## Scope
- Public RPC / REST / gRPC / WebSocket endpoints
- Public explorer, faucet, monitoring, and artifacts endpoints
- Validator nodes are **not** directly exposed to the public Internet

## Baseline Expectations
1. **Sentry/Observer architecture**
   - Public traffic terminates on non-validator full nodes (observer/sentry).
   - Validators are isolated from direct Internet exposure.

2. **Unsafe RPC endpoints disabled**
   - Tendermint/CometBFT unsafe RPC endpoints must never be publicly exposed.

3. **Rate limiting and connection limits**
   - Per-IP request rate limits for RPC/REST/gRPC.
   - Tighter limits for transaction broadcast endpoints to reduce spam.

4. **TLS everywhere**
   - HTTPS for RPC/REST.
   - TLS for gRPC and gRPC-web.
   - HSTS enabled on public HTTPS endpoints.

5. **Access control for metrics**
   - Prometheus metrics should be restricted to private networks.

6. **Monitoring and health checks**
   - Public endpoint availability and health should be continuously monitored.

## Implementation in this repo
- Reverse proxy configs:
  - `nginx/rpc-proxy.conf`
  - `nginx/testnet-proxy.conf`
- Public endpoint checks:
  - `scripts/public-testnet-health-check.sh`

## Node Configuration (in aura-testnets repo)
The following node-level settings should be enforced in the testnet config repo:

- `rpc.unsafe = false`
- `rpc.max_open_connections` set to a reasonable bound for the host
- gRPC connection limits sized to OS limits
- Sentry architecture for validators (validator RPC not public)

## How to verify
Run the public health check:

```bash
./scripts/public-testnet-health-check.sh
```

For JSON output:

```bash
./scripts/public-testnet-health-check.sh --json
```
