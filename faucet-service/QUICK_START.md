# AURA Testnet Faucet - Quick Start Guide

Get the AURA Testnet Faucet running in 5 minutes!

## Prerequisites

- Docker and Docker Compose installed
- Access to an AURA testnet node
- (Optional) hCaptcha keys for production

## Quick Start - Development

### 1. Configure Environment

```bash
cd aura/faucet-service
cp .env.example .env
```

Edit `.env` and set these required values:
```env
NODE_RPC=http://your-aura-node:26657
CHAIN_ID=aura-testnet-1
FAUCET_MNEMONIC=your-mnemonic-phrase-here
FAUCET_ADDRESS=aura1...
```

### 2. Start Services

```bash
docker-compose up -d
```

### 3. Access Faucet

- **Web UI**: http://localhost:8080
- **API**: http://localhost:8080/api/v1
- **Health**: http://localhost:8080/api/v1/health

### 4. Test It Out

Open http://localhost:8080 in your browser:
1. Enter an AURA testnet address (starts with `aura1`)
2. Complete the captcha (skipped in development)
3. Click "Request Tokens"
4. Check the transaction in your wallet or explorer

## Quick Start - Local Development

### 1. Start Dependencies

```bash
docker-compose up -d postgres redis
```

### 2. Run Backend

```bash
cd backend
go mod download
go run main.go
```

### 3. Access Frontend

Open `frontend/index.html` in your browser or serve it:
```bash
cd frontend
python3 -m http.server 8000
```

## Quick Start - Production

### 1. Production Configuration

```bash
cp .env.example .env
```

Edit `.env` with production values:
```env
ENVIRONMENT=production
NODE_RPC=https://rpc.aura-testnet.com
CHAIN_ID=aura-testnet-1
FAUCET_MNEMONIC=use-secrets-manager
FAUCET_ADDRESS=aura1...
HCAPTCHA_SECRET=your-production-secret
LOG_LEVEL=info
CORS_ORIGINS=https://faucet.aura-chain.com
```

### 2. Deploy with Production Profile

```bash
docker-compose --profile production up -d
```

### 3. Verify Deployment

```bash
# Check status
docker-compose ps

# View logs
docker-compose logs -f faucet-backend

# Test health
curl http://localhost:8080/api/v1/health
```

## Common Commands

### View Logs
```bash
docker-compose logs -f faucet-backend
```

### Restart Service
```bash
docker-compose restart faucet-backend
```

### Stop All Services
```bash
docker-compose down
```

### Check Faucet Balance
```bash
curl http://localhost:8080/api/v1/faucet/info | jq .balance
```

### View Recent Transactions
```bash
curl http://localhost:8080/api/v1/faucet/recent | jq
```

### Run Tests
```bash
cd backend
go test ./... -v
```

## Troubleshooting

### Cannot connect to database
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# View PostgreSQL logs
docker-compose logs postgres
```

### Cannot connect to AURA node
```bash
# Test node connection
curl http://your-aura-node:26657/status

# Check node in faucet logs
docker-compose logs faucet-backend | grep "node"
```

### Faucet balance is low
```bash
# Check balance
curl http://localhost:8080/api/v1/faucet/info | jq .balance

# Send more tokens to faucet address
aurad tx bank send sender $FAUCET_ADDRESS 10000000000uaura --chain-id aura-testnet-1
```

### Rate limited
```bash
# Clear rate limits (development only!)
docker-compose exec redis redis-cli FLUSHDB
```

## Next Steps

1. **Read Full Documentation**: See [README.md](README.md)
2. **Deploy to Production**: See [DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)
3. **Set Up Monitoring**: See [MONITORING_ALERTING.md](MONITORING_ALERTING.md)
4. **Review Integration**: See [INTEGRATION_REPORT.md](INTEGRATION_REPORT.md)

## Getting Help

- **Documentation**: Start with README.md
- **Issues**: Check logs with `docker-compose logs`
- **GitHub**: Report issues or request features
- **Community**: Join AURA Discord/Telegram

---

**Happy Faucet Running! 💧**
