# AURA Block Explorer and Faucet Deployment Guide

This guide covers deploying a block explorer and faucet service for AURA local and cloud testnets.

## Overview

A complete testnet requires:
- **Block Explorer**: Web UI for browsing blocks, transactions, validators, and custom AURA modules
- **Faucet Service**: Automated token distribution for testnet users
- **Public APIs**: Exposed RPC, REST API, and gRPC endpoints with proper CORS configuration

## Prerequisites

- Running AURA testnet (local or cloud)
- Docker and Docker Compose (for containerized deployment)
- Node.js 18+ (for some explorer options)
- Domain names configured (for cloud deployment)

---

## Part 1: API Endpoint Configuration

### RPC/API/gRPC Endpoint Setup

Before deploying explorer and faucet, ensure your AURA node exposes the required endpoints.

#### 1.1 Configure app.toml

Edit `~/.aura/config/app.toml` (or your custom data directory):

```toml
###############################################################################
###                           gRPC Configuration                            ###
###############################################################################

[grpc]
# Enable defines if the gRPC server should be enabled.
enable = true

# Address defines the gRPC server address to bind to.
address = "0.0.0.0:9090"

###############################################################################
###                        gRPC Web Configuration                           ###
###############################################################################

[grpc-web]
# GRPCWebEnable defines if the gRPC-web should be enabled.
enable = true

# Address defines the gRPC-web server address to bind to.
address = "0.0.0.0:9091"

# EnableUnsafeCORS defines if CORS should be enabled (unsafe - use for local testnet only)
enable-unsafe-cors = true

###############################################################################
###                           API Configuration                             ###
###############################################################################

[api]
# Enable defines if the API server should be enabled.
enable = true

# Swagger defines if swagger documentation should automatically be registered.
swagger = true

# Address defines the API server to listen on.
address = "tcp://0.0.0.0:1317"

# MaxOpenConnections defines the number of maximum open connections.
max-open-connections = 1000

# RPCReadTimeout defines the Tendermint RPC read timeout (in seconds).
rpc-read-timeout = 10

# RPCWriteTimeout defines the Tendermint RPC write timeout (in seconds).
rpc-write-timeout = 10

# RPCMaxBodyBytes defines the Tendermint maximum response body (in bytes).
rpc-max-body-bytes = 1000000

# EnableUnsafeCORS defines if CORS should be enabled (unsafe - use for local testnet only)
enabled-unsafe-cors = true

# AllowedOrigins defines the CORS allowed origins (comma-separated list)
# For local testnet: "*"
# For production: "https://explorer.aura.network,https://faucet.aura.network"
allowed-origins = ["*"]
```

#### 1.2 Configure config.toml

Edit `~/.aura/config/config.toml`:

```toml
#######################################################
###       RPC Server Configuration Options          ###
#######################################################
[rpc]

# TCP or UNIX socket address for the RPC server to listen on
laddr = "tcp://0.0.0.0:26657"

# A list of origins a cross-domain request can be executed from
# Default value '[]' disables cors support
# Use '["*"]' to allow any origin
cors_allowed_origins = ["*"]

# A list of methods the client is allowed to use with cross-domain requests
cors_allowed_methods = ["HEAD", "GET", "POST"]

# A list of non simple headers the client is allowed to use with cross-domain requests
cors_allowed_headers = ["Origin", "Accept", "Content-Type", "X-Requested-With", "X-Server-Time"]

# Maximum number of unique clientIDs the WebSocket server will track
max_subscription_clients = 100

# Maximum number of unique queries a given client can have open at a time
max_subscriptions_per_client = 5

# How long to wait for a tx to be committed during /broadcast_tx_commit
# WARNING: Using a value larger than 10s will result in increasing the
# global HTTP write timeout, which applies to all connections and endpoints.
timeout_broadcast_tx_commit = "10s"

# Maximum size of request body, in bytes
max_body_bytes = 1000000

# Maximum size of request header, in bytes
max_header_bytes = 1048576

# The path to a file containing certificate that is used to create the HTTPS server.
# Might be either absolute path or path related to Tendermint's config directory.
# If the certificate is signed by a certificate authority,
# the certFile should be the concatenation of the server's certificate, any intermediates,
# and the CA's certificate.
# NOTE: both tls-cert-file and tls-key-file must be present for Tendermint to create HTTPS server.
# Otherwise, HTTP server is run.
tls_cert_file = ""

# The path to a file containing matching private key that is used to create the HTTPS server.
# Might be either absolute path or path related to tendermint's config directory.
# NOTE: both tls-cert-file and tls-key-file must be present for Tendermint to create HTTPS server.
# Otherwise, HTTP server is run.
tls_key_file = ""

# pprof listen address (https://golang.org/pkg/net/http/pprof)
pprof_laddr = "localhost:6060"
```

#### 1.3 Restart Node

After configuration changes:

```bash
# If running as systemd service
sudo systemctl restart aurad

# If running in Docker
docker-compose restart validator-1

# If running manually
pkill aurad
./aurad start
```

#### 1.4 Verify Endpoints

```bash
# Test RPC endpoint
curl http://localhost:26657/status | jq .

# Test REST API endpoint
curl http://localhost:1317/cosmos/base/tendermint/v1beta1/blocks/latest | jq .

# Test gRPC endpoint (requires grpcurl)
grpcurl -plaintext localhost:9090 list

# Test AURA custom modules
curl http://localhost:1317/aura/identity/v1/identities | jq .
curl http://localhost:1317/aura/vcregistry/v1/credentials | jq .
curl http://localhost:1317/aura/inclusionroutines/v1/routines | jq .
```

#### 1.5 CORS Configuration Notes

**For Local Testnet:**
- Use `allowed-origins = ["*"]` for maximum compatibility
- Enable `enable-unsafe-cors = true`
- This is acceptable for local development ONLY

**For Cloud/Production Testnet:**
- Restrict origins to specific domains:
  ```toml
  allowed-origins = [
    "https://explorer.testnet.aura.network",
    "https://faucet.testnet.aura.network",
    "https://wallet.testnet.aura.network"
  ]
  ```
- Set `enable-unsafe-cors = false`
- Use reverse proxy (nginx) for additional security

---

## Part 2: Block Explorer Deployment

### Option A: Big Dipper (Recommended for AURA)

Big Dipper is a popular Cosmos SDK block explorer with extensive customization options.

#### 2.1 Prerequisites

```bash
# Install dependencies
sudo apt-get update
sudo apt-get install -y nodejs npm postgresql docker.io docker-compose

# Install Yarn
npm install -g yarn
```

#### 2.2 Clone and Configure Big Dipper

```bash
# Clone Big Dipper repository
git clone https://github.com/forbole/big-dipper-2.0-cosmos.git
cd big-dipper-2.0-cosmos

# Install dependencies
yarn install

# Create environment file
cp .env.example .env
```

Edit `.env`:

```env
# Chain configuration
NEXT_PUBLIC_CHAIN_ID=aura-local-4
NEXT_PUBLIC_CHAIN_NAME=AURA Testnet
NEXT_PUBLIC_CHAIN_TYPE=mainnet

# RPC endpoints
NEXT_PUBLIC_RPC_WEBSOCKET=ws://localhost:26657/websocket
NEXT_PUBLIC_RPC_URL=http://localhost:26657

# REST API
NEXT_PUBLIC_API_URL=http://localhost:1317

# GraphQL endpoint (optional)
NEXT_PUBLIC_GRAPHQL_URL=http://localhost:3000/v1/graphql
NEXT_PUBLIC_GRAPHQL_WS=ws://localhost:3000/v1/graphql

# Database (Hasura + PostgreSQL)
POSTGRES_USER=aura
POSTGRES_PASSWORD=aura-secure-password
POSTGRES_DB=aura_explorer
HASURA_GRAPHQL_DATABASE_URL=postgres://aura:aura-secure-password@postgres:5432/aura_explorer
HASURA_GRAPHQL_ENABLE_CONSOLE=true
HASURA_GRAPHQL_ADMIN_SECRET=aura-hasura-secret
HASURA_GRAPHQL_UNAUTHORIZED_ROLE=user

# BDJuno (blockchain data indexer)
CHAIN_ID=aura-local-4
RPC_ADDR=http://host.docker.internal:26657
GRPC_ADDR=localhost:9090
API_ADDR=http://localhost:1317
```

#### 2.3 Customize for AURA Modules

Create `config.yaml`:

```yaml
chain:
  id: aura-local-4
  name: AURA Testnet
  prefix: aura
  logo: /aura-logo.png
  denom: uaura
  exponent: 6
  symbol: AURA

modules:
  - name: bank
    enabled: true
  - name: staking
    enabled: true
  - name: gov
    enabled: true
  - name: distribution
    enabled: true
  - name: slashing
    enabled: true
  - name: identity
    enabled: true
    path: /aura/identity/v1
    description: "AURA Identity Module - DID and identity management"
  - name: vcregistry
    enabled: true
    path: /aura/vcregistry/v1
    description: "Verifiable Credentials Registry"
  - name: inclusionroutines
    enabled: true
    path: /aura/inclusionroutines/v1
    description: "AI-powered Inclusion Routines"
  - name: compliance
    enabled: true
    path: /aura/compliance/v1
    description: "Compliance and AML screening"
  - name: bridge
    enabled: true
    path: /aura/bridge/v1
    description: "Cross-chain bridge"
  - name: dex
    enabled: true
    path: /aura/dex/v1
    description: "Decentralized exchange"

endpoints:
  rpc: http://localhost:26657
  api: http://localhost:1317
  grpc: localhost:9090
  websocket: ws://localhost:26657/websocket

ui:
  theme: dark
  logo: /aura-logo.png
  primaryColor: "#6366f1"
  accentColor: "#8b5cf6"
```

#### 2.4 Deploy with Docker Compose

Create `docker-compose-explorer.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: aura-explorer-db
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    networks:
      - aura-explorer

  hasura:
    image: hasura/graphql-engine:v2.36.0
    container_name: aura-explorer-hasura
    restart: unless-stopped
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    environment:
      HASURA_GRAPHQL_DATABASE_URL: ${HASURA_GRAPHQL_DATABASE_URL}
      HASURA_GRAPHQL_ENABLE_CONSOLE: ${HASURA_GRAPHQL_ENABLE_CONSOLE}
      HASURA_GRAPHQL_ADMIN_SECRET: ${HASURA_GRAPHQL_ADMIN_SECRET}
      HASURA_GRAPHQL_UNAUTHORIZED_ROLE: ${HASURA_GRAPHQL_UNAUTHORIZED_ROLE}
      HASURA_GRAPHQL_ENABLE_TELEMETRY: "false"
    networks:
      - aura-explorer

  bdjuno:
    image: forbole/bdjuno:cosmos-v4.0.0
    container_name: aura-explorer-indexer
    restart: unless-stopped
    depends_on:
      - postgres
      - hasura
    environment:
      CHAIN_ID: ${CHAIN_ID}
      RPC_ADDR: ${RPC_ADDR}
      GRPC_ADDR: ${GRPC_ADDR}
      DATABASE_URL: ${HASURA_GRAPHQL_DATABASE_URL}
    volumes:
      - ./bdjuno-config.yaml:/config.yaml:ro
    command: bdjuno parse --config /config.yaml
    networks:
      - aura-explorer
    extra_hosts:
      - "host.docker.internal:host-gateway"

  frontend:
    image: forbole/big-dipper-2.0-cosmos:latest
    container_name: aura-explorer-ui
    restart: unless-stopped
    ports:
      - "3000:3000"
    depends_on:
      - hasura
    environment:
      NEXT_PUBLIC_CHAIN_ID: ${NEXT_PUBLIC_CHAIN_ID}
      NEXT_PUBLIC_CHAIN_NAME: ${NEXT_PUBLIC_CHAIN_NAME}
      NEXT_PUBLIC_RPC_URL: ${NEXT_PUBLIC_RPC_URL}
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL}
      NEXT_PUBLIC_GRAPHQL_URL: ${NEXT_PUBLIC_GRAPHQL_URL}
    networks:
      - aura-explorer

networks:
  aura-explorer:
    driver: bridge

volumes:
  postgres_data:
```

#### 2.5 Start Explorer

```bash
# Start all services
docker-compose -f docker-compose-explorer.yml up -d

# Check status
docker-compose -f docker-compose-explorer.yml ps

# View logs
docker-compose -f docker-compose-explorer.yml logs -f

# Access explorer at http://localhost:3000
```

### Option B: Ping.pub (Lightweight Alternative)

Ping.pub is a lightweight, client-side block explorer that doesn't require backend infrastructure.

#### 2.6 Deploy Ping.pub

```bash
# Clone repository
git clone https://github.com/ping-pub/explorer.git
cd explorer

# Install dependencies
yarn install

# Create AURA chain configuration
mkdir -p chains/mainnet
cat > chains/mainnet/aura.json <<EOF
{
  "chain_name": "aura",
  "coingecko": "",
  "api": ["http://localhost:1317"],
  "rpc": ["http://localhost:26657"],
  "snapshot_provider": "",
  "sdk_version": "0.53.4",
  "coin_type": "118",
  "min_tx_fee": "800",
  "addr_prefix": "aura",
  "logo": "/logos/aura.svg",
  "assets": [{
    "base": "uaura",
    "symbol": "AURA",
    "exponent": "6",
    "coingecko_id": "",
    "logo": "/logos/aura.svg"
  }]
}
EOF

# Start development server
yarn dev

# Or build for production
yarn build
yarn start

# Access at http://localhost:5173
```

### Option C: Mintscan (Professional-Grade)

For production deployments, consider Mintscan (commercial option with support).

Contact: https://www.mintscan.io/

---

## Part 3: Faucet Service Deployment

### Option A: Cosmos Faucet (Go-based)

#### 3.1 Install Cosmos Faucet

```bash
# Clone repository
git clone https://github.com/cosmos/faucet.git
cd faucet

# Build
go build -o faucet ./cmd/faucet
```

#### 3.2 Create Configuration

Create `config.yaml`:

```yaml
# Chain configuration
chain_id: "aura-local-4"
denom: "uaura"
amount: 10000000  # 10 AURA (10^6 uaura)
max_credit: 100000000  # 100 AURA max per address

# Node connection
node: "http://localhost:26657"

# Faucet account (must have tokens)
mnemonic: "your faucet account mnemonic phrase here"
keyring_backend: "test"

# Rate limiting
cooldown: 86400  # 24 hours in seconds
ip_limit: 10     # Max requests per IP per day

# Server configuration
port: 8000
host: "0.0.0.0"

# Security
allowed_origins:
  - "*"  # Local testnet only
  # - "https://testnet.aura.network"  # Production

# Captcha (optional)
recaptcha:
  enabled: false
  site_key: ""
  secret_key: ""

# Logging
log_level: "info"
log_format: "json"
```

#### 3.3 Fund Faucet Account

```bash
# Create faucet account if needed
./aurad keys add faucet --keyring-backend test

# Get address
FAUCET_ADDR=$(./aurad keys show faucet --keyring-backend test -a)

# Fund from validator account (on testnet)
./aurad tx bank send validator-1 $FAUCET_ADDR 1000000000uaura \
  --keyring-backend test \
  --chain-id aura-local-4 \
  --yes

# Verify balance
./aurad query bank balances $FAUCET_ADDR
```

#### 3.4 Run Faucet Service

```bash
# Start faucet
./faucet --config config.yaml

# Or run with Docker
docker run -d \
  --name aura-faucet \
  -p 8000:8000 \
  -v $(pwd)/config.yaml:/config.yaml:ro \
  -e FAUCET_CONFIG=/config.yaml \
  cosmos/faucet:latest

# Test faucet
curl -X POST http://localhost:8000/claim \
  -H "Content-Type: application/json" \
  -d '{"address": "aura1..."}'
```

### Option B: Custom Faucet (Node.js)

#### 3.5 Create Custom Faucet

Create `faucet-server.js`:

```javascript
const express = require('express');
const cors = require('cors');
const rateLimit = require('express-rate-limit');
const { DirectSecp256k1HdWallet } = require('@cosmjs/proto-signing');
const { SigningStargateClient } = require('@cosmjs/stargate');

const app = express();
app.use(express.json());
app.use(cors());

// Configuration
const CONFIG = {
  chainId: 'aura-local-4',
  rpcEndpoint: 'http://localhost:26657',
  faucetMnemonic: process.env.FAUCET_MNEMONIC,
  amount: '10000000',  // 10 AURA
  denom: 'uaura',
  prefix: 'aura',
  gasPrice: '0.025uaura'
};

// Rate limiting
const limiter = rateLimit({
  windowMs: 24 * 60 * 60 * 1000, // 24 hours
  max: 5, // 5 requests per IP
  message: 'Too many requests, please try again later'
});

app.use('/claim', limiter);

// Faucet claim endpoint
app.post('/claim', async (req, res) => {
  try {
    const { address } = req.body;

    // Validate address
    if (!address || !address.startsWith(CONFIG.prefix)) {
      return res.status(400).json({
        error: 'Invalid address format'
      });
    }

    // Create wallet from mnemonic
    const wallet = await DirectSecp256k1HdWallet.fromMnemonic(
      CONFIG.faucetMnemonic,
      { prefix: CONFIG.prefix }
    );

    // Connect to chain
    const client = await SigningStargateClient.connectWithSigner(
      CONFIG.rpcEndpoint,
      wallet,
      { gasPrice: CONFIG.gasPrice }
    );

    // Get faucet address
    const [faucetAccount] = await wallet.getAccounts();

    // Send tokens
    const result = await client.sendTokens(
      faucetAccount.address,
      address,
      [{ denom: CONFIG.denom, amount: CONFIG.amount }],
      {
        amount: [{ denom: CONFIG.denom, amount: '500' }],
        gas: '200000'
      },
      'AURA Testnet Faucet'
    );

    res.json({
      success: true,
      txHash: result.transactionHash,
      amount: CONFIG.amount,
      denom: CONFIG.denom
    });

  } catch (error) {
    console.error('Faucet error:', error);
    res.status(500).json({
      error: 'Failed to send tokens',
      message: error.message
    });
  }
});

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

// Status endpoint
app.get('/status', async (req, res) => {
  try {
    const wallet = await DirectSecp256k1HdWallet.fromMnemonic(
      CONFIG.faucetMnemonic,
      { prefix: CONFIG.prefix }
    );

    const client = await SigningStargateClient.connectWithSigner(
      CONFIG.rpcEndpoint,
      wallet
    );

    const [faucetAccount] = await wallet.getAccounts();
    const balance = await client.getBalance(faucetAccount.address, CONFIG.denom);

    res.json({
      address: faucetAccount.address,
      balance: balance,
      chainId: CONFIG.chainId,
      amount: CONFIG.amount,
      denom: CONFIG.denom
    });

  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

const PORT = process.env.PORT || 8000;
app.listen(PORT, () => {
  console.log(`AURA Faucet running on port ${PORT}`);
});
```

Create `package.json`:

```json
{
  "name": "aura-faucet",
  "version": "1.0.0",
  "main": "faucet-server.js",
  "scripts": {
    "start": "node faucet-server.js",
    "dev": "nodemon faucet-server.js"
  },
  "dependencies": {
    "@cosmjs/proto-signing": "^0.32.0",
    "@cosmjs/stargate": "^0.32.0",
    "cors": "^2.8.5",
    "express": "^4.18.2",
    "express-rate-limit": "^7.1.0"
  },
  "devDependencies": {
    "nodemon": "^3.0.0"
  }
}
```

#### 3.6 Deploy Custom Faucet

```bash
# Install dependencies
npm install

# Set environment variable
export FAUCET_MNEMONIC="your faucet mnemonic here"

# Start faucet
npm start

# Or with Docker
docker build -t aura-faucet .
docker run -d \
  --name aura-faucet \
  -p 8000:8000 \
  -e FAUCET_MNEMONIC="your mnemonic" \
  aura-faucet
```

### Option C: Discord/Telegram Faucet Bot

For community engagement, deploy a Discord or Telegram bot:

```bash
# Clone Discord faucet bot
git clone https://github.com/cosmos/discord-faucet-bot.git
cd discord-faucet-bot

# Configure
cp config.example.yaml config.yaml
# Edit config.yaml with your Discord bot token and AURA settings

# Run bot
npm install
npm start
```

---

## Part 4: Integrated Deployment

### 4.1 Combined Docker Compose

Create `docker-compose-services.yml`:

```yaml
version: '3.8'

services:
  # AURA Node (from existing testnet setup)
  validator-1:
    extends:
      file: docker-compose.testnet.yml
      service: validator-1

  # PostgreSQL for Explorer
  explorer-db:
    image: postgres:15-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: aura
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: aura_explorer
    volumes:
      - explorer_db:/var/lib/postgresql/data
    networks:
      - aura-services

  # Hasura GraphQL Engine
  hasura:
    image: hasura/graphql-engine:v2.36.0
    restart: unless-stopped
    ports:
      - "8080:8080"
    depends_on:
      - explorer-db
    environment:
      HASURA_GRAPHQL_DATABASE_URL: postgres://aura:${DB_PASSWORD}@explorer-db:5432/aura_explorer
      HASURA_GRAPHQL_ENABLE_CONSOLE: "true"
      HASURA_GRAPHQL_ADMIN_SECRET: ${HASURA_SECRET}
    networks:
      - aura-services

  # BDJuno Indexer
  indexer:
    image: forbole/bdjuno:cosmos-v4.0.0
    restart: unless-stopped
    depends_on:
      - explorer-db
      - validator-1
    volumes:
      - ./explorer-config.yaml:/config.yaml:ro
    command: bdjuno parse --config /config.yaml
    networks:
      - aura-services

  # Explorer Frontend
  explorer:
    image: forbole/big-dipper-2.0-cosmos:latest
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      NEXT_PUBLIC_CHAIN_ID: aura-local-4
      NEXT_PUBLIC_RPC_URL: http://validator-1:26657
      NEXT_PUBLIC_API_URL: http://validator-1:1317
      NEXT_PUBLIC_GRAPHQL_URL: http://hasura:8080/v1/graphql
    networks:
      - aura-services

  # Faucet Service
  faucet:
    build: ./faucet
    restart: unless-stopped
    ports:
      - "8000:8000"
    environment:
      FAUCET_MNEMONIC: ${FAUCET_MNEMONIC}
      CHAIN_ID: aura-local-4
      RPC_ENDPOINT: http://validator-1:26657
    depends_on:
      - validator-1
    networks:
      - aura-services

networks:
  aura-services:
    driver: bridge

volumes:
  explorer_db:
```

### 4.2 Start All Services

```bash
# Create .env file
cat > .env <<EOF
DB_PASSWORD=secure-db-password
HASURA_SECRET=secure-hasura-secret
FAUCET_MNEMONIC=your faucet mnemonic phrase
EOF

# Start all services
docker-compose -f docker-compose-services.yml up -d

# Check status
docker-compose -f docker-compose-services.yml ps

# Access services:
# - Explorer: http://localhost:3000
# - Faucet: http://localhost:8000
# - Hasura Console: http://localhost:8080/console (admin secret required)
```

---

## Part 5: Production Considerations

### 5.1 Security Hardening

**For Production/Cloud Testnet:**

1. **Disable CORS wildcard:**
   ```toml
   allowed-origins = ["https://explorer.aura.network", "https://faucet.aura.network"]
   ```

2. **Enable HTTPS:**
   - Use Let's Encrypt for SSL certificates
   - Configure nginx reverse proxy
   - Enforce HTTPS redirects

3. **Rate Limiting:**
   - Implement aggressive rate limiting on faucet
   - Use Cloudflare or similar DDoS protection
   - Add CAPTCHA to faucet frontend

4. **Faucet Security:**
   - Store mnemonic in encrypted vault (Vault, AWS Secrets Manager)
   - Implement IP-based and address-based rate limiting
   - Monitor for abuse patterns
   - Set daily distribution limits

5. **Database Security:**
   - Use strong passwords
   - Enable SSL for database connections
   - Regular backups
   - Network isolation

### 5.2 Monitoring

Add monitoring for explorer and faucet:

```yaml
# prometheus/prometheus-services.yml
scrape_configs:
  - job_name: 'faucet'
    static_configs:
      - targets: ['faucet:8000']
    metrics_path: '/metrics'

  - job_name: 'explorer-indexer'
    static_configs:
      - targets: ['indexer:26660']
```

### 5.3 Nginx Reverse Proxy

For production, use nginx:

```nginx
# /etc/nginx/sites-available/aura-testnet

# Explorer
server {
    listen 80;
    server_name explorer.testnet.aura.network;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# Faucet
server {
    listen 80;
    server_name faucet.testnet.aura.network;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=faucet:10m rate=5r/h;
    limit_req zone=faucet burst=2 nodelay;

    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# RPC (read-only)
server {
    listen 80;
    server_name rpc.testnet.aura.network;

    location / {
        proxy_pass http://localhost:26657;
        proxy_set_header Host $host;
    }
}

# API (read-only)
server {
    listen 80;
    server_name api.testnet.aura.network;

    location / {
        proxy_pass http://localhost:1317;
        proxy_set_header Host $host;
    }
}
```

---

## Part 6: Testing and Verification

### 6.1 Explorer Testing

```bash
# Verify explorer is indexing
curl http://localhost:8080/v1/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ block(limit: 1, order_by: {height: desc}) { height timestamp } }"}'

# Check explorer UI
curl http://localhost:3000

# Verify AURA custom modules are displayed
# Navigate to: http://localhost:3000/modules
```

### 6.2 Faucet Testing

```bash
# Create test account
TEST_ADDR=$(./aurad keys show test-account --keyring-backend test -a 2>/dev/null || \
  (./aurad keys add test-account --keyring-backend test && \
   ./aurad keys show test-account --keyring-backend test -a))

# Request tokens from faucet
curl -X POST http://localhost:8000/claim \
  -H "Content-Type: application/json" \
  -d "{\"address\": \"$TEST_ADDR\"}"

# Verify balance
./aurad query bank balances $TEST_ADDR

# Check faucet status
curl http://localhost:8000/status | jq .
```

### 6.3 API Endpoint Testing

```bash
# Test all public endpoints
curl http://localhost:26657/status | jq .result.node_info
curl http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info | jq .
curl http://localhost:1317/aura/identity/v1/identities | jq .
curl http://localhost:1317/aura/vcregistry/v1/credentials | jq .

# Test WebSocket connection
wscat -c ws://localhost:26657/websocket
# Send: {"jsonrpc":"2.0","method":"subscribe","id":1,"params":{"query":"tm.event='NewBlock'"}}
```

---

## Part 7: Maintenance and Operations

### 7.1 Backup and Restore

```bash
# Backup explorer database
docker exec explorer-db pg_dump -U aura aura_explorer > explorer_backup.sql

# Restore
cat explorer_backup.sql | docker exec -i explorer-db psql -U aura aura_explorer
```

### 7.2 Monitoring Dashboard

Add Grafana dashboard for services:

```json
{
  "dashboard": {
    "title": "AURA Services Dashboard",
    "panels": [
      {
        "title": "Faucet Requests",
        "targets": [{"expr": "rate(faucet_requests_total[5m])"}]
      },
      {
        "title": "Explorer Indexer Lag",
        "targets": [{"expr": "indexer_blocks_behind"}]
      },
      {
        "title": "API Response Time",
        "targets": [{"expr": "http_request_duration_seconds"}]
      }
    ]
  }
}
```

### 7.3 Upgrade Procedures

When upgrading AURA node:

1. Stop services that depend on node
2. Upgrade node
3. Restart node and verify
4. Update explorer indexer if schema changed
5. Restart all services

---

## Quick Reference

### Service URLs (Local Testnet)

| Service | URL | Purpose |
|---------|-----|---------|
| Explorer | http://localhost:3000 | Browse blocks, txs, validators |
| Faucet | http://localhost:8000 | Request testnet tokens |
| RPC | http://localhost:26657 | CometBFT RPC endpoint |
| REST API | http://localhost:1317 | Cosmos SDK REST API |
| gRPC | localhost:9090 | gRPC endpoint |
| Hasura | http://localhost:8080/console | GraphQL admin |
| Prometheus | http://localhost:9091 | Metrics |
| Grafana | http://localhost:3001 | Dashboards |

### Service URLs (Cloud Testnet - Example)

| Service | URL | Purpose |
|---------|-----|---------|
| Explorer | https://explorer.testnet.aura.network | Public explorer |
| Faucet | https://faucet.testnet.aura.network | Public faucet |
| RPC | https://rpc.testnet.aura.network | Public RPC |
| REST API | https://api.testnet.aura.network | Public API |
| gRPC | grpc.testnet.aura.network:443 | Public gRPC |

### Common Commands

```bash
# Start all services
docker-compose -f docker-compose-services.yml up -d

# Stop all services
docker-compose -f docker-compose-services.yml down

# View logs
docker-compose -f docker-compose-services.yml logs -f

# Restart specific service
docker-compose -f docker-compose-services.yml restart faucet

# Check faucet balance
curl http://localhost:8000/status | jq .balance

# Fund faucet
./aurad tx bank send validator-1 <faucet-address> 1000000000uaura \
  --keyring-backend test --chain-id aura-local-4 --yes
```

---

## Resources

### Explorer Projects
- **Big Dipper**: https://github.com/forbole/big-dipper-2.0-cosmos
- **Ping.pub**: https://github.com/ping-pub/explorer
- **Mintscan**: https://www.mintscan.io/ (commercial)
- **ATOMScan**: https://atomscan.com/

### Faucet Projects
- **Cosmos Faucet**: https://github.com/cosmos/faucet
- **CosmJS Faucet**: https://github.com/cosmos/cosmjs/tree/main/packages/faucet
- **Discord Faucet Bot**: https://github.com/cosmos/discord-faucet-bot

### Documentation
- **Cosmos SDK**: https://docs.cosmos.network/
- **CometBFT RPC**: https://docs.cometbft.com/v0.38/core/rpc
- **Hasura**: https://hasura.io/docs/latest/index/
- **AURA Modules**: `/docs/modules/`

---

## Next Steps

After deploying explorer and faucet:

1. **Test all AURA modules** through explorer UI
2. **Distribute tokens** to community testers via faucet
3. **Monitor usage** through Grafana dashboards
4. **Collect feedback** on module functionality
5. **Progress to cloud deployment** (Phase 2 in ROADMAP_PRODUCTION.md)

For production deployment, see:
- `/docs/ops/PRODUCTION_DEPLOYMENT.md`
- `/docs/runbooks/EXPLORER_DEPLOYMENT.md` (to be created)
- `/docs/runbooks/FAUCET_DEPLOYMENT.md` (to be created)
