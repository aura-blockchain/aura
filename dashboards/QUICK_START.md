# AURA Dashboards Quick Start Guide

## Overview

This guide will help you get all three AURA dashboards up and running in minutes.

## Prerequisites

- Node.js 16+ installed
- Python 3+ (for simple HTTP servers)
- AURA blockchain node running (or access to AURA testnet/mainnet)
- Git (for cloning the repository)

## Quick Setup - All Dashboards

### Option 1: Individual Setup

#### Validator Dashboard (Port 8080)

```bash
cd dashboards/validator
npm install
npm start
```

Open browser: http://localhost:8080

#### Staking Dashboard (Port 8081)

```bash
cd dashboards/staking
npm install
npm start
```

Open browser: http://localhost:8081

#### Governance Dashboard (Port 8082)

```bash
cd dashboards/governance
npm install
npm start
```

Open browser: http://localhost:8082

### Option 2: Batch Setup Script

Create a file `start-all-dashboards.sh`:

```bash
#!/bin/bash

echo "Starting all AURA dashboards..."

# Start Validator Dashboard
cd dashboards/validator
npm install
npm start &
VALIDATOR_PID=$!

# Start Staking Dashboard
cd ../staking
npm install
npm start &
STAKING_PID=$!

# Start Governance Dashboard
cd ../governance
npm install
npm start &
GOVERNANCE_PID=$!

echo "All dashboards started!"
echo "Validator Dashboard: http://localhost:8080"
echo "Staking Dashboard: http://localhost:8081"
echo "Governance Dashboard: http://localhost:8082"
echo ""
echo "Press Ctrl+C to stop all dashboards"

# Wait for Ctrl+C
trap "kill $VALIDATOR_PID $STAKING_PID $GOVERNANCE_PID; exit" INT
wait
```

Make it executable and run:

```bash
chmod +x start-all-dashboards.sh
./start-all-dashboards.sh
```

## Configuration

### Configure Endpoints

Edit `dashboards/config.js`:

```javascript
const AuraConfig = {
    endpoints: {
        // Your AURA node REST API
        rest: 'http://your-aura-node:1317',

        // Your AURA node RPC
        rpc: 'http://your-aura-node:26657',

        // Your AURA node gRPC (if needed)
        grpc: 'http://your-aura-node:9090'
    }
};
```

### Or use Environment Variables

```bash
export AURA_REST_ENDPOINT=http://your-aura-node:1317
export AURA_RPC_ENDPOINT=http://your-aura-node:26657
export AURA_GRPC_ENDPOINT=http://your-aura-node:9090
```

### Mock Mode for Development

To test without a running AURA node:

```bash
export AURA_MOCK_MODE=true
npm start
```

## Testing

### Run All Tests

```bash
# Validator Dashboard
cd dashboards/validator
npm test

# Staking Dashboard
cd dashboards/staking
npm test

# Governance Dashboard
cd dashboards/governance
npm test
```

### Watch Mode (Development)

```bash
npm run test:watch
```

### Coverage Report

```bash
npm run test:coverage
```

## Docker Deployment

### Validator Dashboard

```bash
cd dashboards/validator
docker-compose up -d
```

### All Dashboards with Docker Compose

Create `docker-compose.yml` in `dashboards/`:

```yaml
version: '3.8'

services:
  validator-dashboard:
    build: ./dashboards/validator
    ports:
      - "8080:80"
    environment:
      - AURA_REST_ENDPOINT=http://aura-node:1317
      - AURA_RPC_ENDPOINT=http://aura-node:26657
    networks:
      - aura-network

  staking-dashboard:
    build: ./dashboards/staking
    ports:
      - "8081:80"
    environment:
      - AURA_REST_ENDPOINT=http://aura-node:1317
      - AURA_RPC_ENDPOINT=http://aura-node:26657
    networks:
      - aura-network

  governance-dashboard:
    build: ./dashboards/governance
    ports:
      - "8082:80"
    environment:
      - AURA_REST_ENDPOINT=http://aura-node:1317
      - AURA_RPC_ENDPOINT=http://aura-node:26657
    networks:
      - aura-network

networks:
  aura-network:
    external: true
```

Run:

```bash
docker-compose up -d
```

## Troubleshooting

### Port Already in Use

```bash
# Find process using port
netstat -ano | findstr :8080  # Windows
lsof -i :8080                  # Linux/Mac

# Kill process
kill -9 <PID>
```

Or change the port in package.json scripts.

### Cannot Connect to AURA Node

1. Verify AURA node is running:
   ```bash
   curl http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info
   ```

2. Check firewall settings

3. Enable CORS in AURA node config (`app.toml`):
   ```toml
   [api]
   enable = true
   swagger = true
   address = "tcp://0.0.0.0:1317"

   [api.cors]
   allowed-origins = ["*"]
   ```

### Tests Failing

1. Install dependencies:
   ```bash
   npm install
   ```

2. Clear cache:
   ```bash
   npm cache clean --force
   rm -rf node_modules package-lock.json
   npm install
   ```

3. Check Node version:
   ```bash
   node --version  # Should be 16+
   ```

## Next Steps

1. **Customize Configuration**
   - Edit `config.js` for your network
   - Set environment variables
   - Configure alerts and notifications

2. **Add Validators**
   - In Validator Dashboard, click "Add Validator"
   - Enter validator operator address
   - Configure monitoring settings

3. **Start Staking**
   - Browse validators in Staking Dashboard
   - Compare validators
   - Delegate tokens

4. **Participate in Governance**
   - View active proposals
   - Vote on proposals
   - Create new proposals (if you have enough stake)

## Production Deployment

### Nginx Configuration

Example nginx config for all dashboards:

```nginx
server {
    listen 80;
    server_name dashboards.aura.network;

    location /validator {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /staking {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /governance {
        proxy_pass http://localhost:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### SSL/TLS

Use Let's Encrypt for free SSL:

```bash
sudo certbot --nginx -d dashboards.aura.network
```

## Support

- **Documentation**: See individual dashboard READMEs
- **Issues**: https://github.com/aura-network/aura/issues
- **Discord**: Join AURA Network community

## Useful Commands

```bash
# Development
npm run dev           # Start with auto-reload
npm run lint          # Check code quality
npm run format        # Format code

# Testing
npm test              # Run tests
npm run test:watch    # Watch mode
npm run test:coverage # Coverage report

# Production
npm run build         # Build for production
npm start             # Start production server
```

## Dashboard URLs

Once running:

- **Validator Dashboard**: http://localhost:8080
- **Staking Dashboard**: http://localhost:8081
- **Governance Dashboard**: http://localhost:8082

Enjoy your AURA dashboards!
