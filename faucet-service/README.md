# AURA Testnet Faucet

A production-ready testnet faucet for the AURA blockchain with rate limiting, captcha protection, and comprehensive monitoring.

## Features

- **Modern Web UI**: Clean, responsive interface built with vanilla JavaScript
- **Rate Limiting**: Per-IP and per-address rate limiting using Redis
- **Captcha Protection**: hCaptcha integration to prevent abuse
- **Database Tracking**: PostgreSQL database for request history and statistics
- **Real-time Statistics**: Live faucet balance and distribution metrics
- **Docker Support**: Complete Docker Compose setup for easy deployment
- **Health Monitoring**: Built-in health checks and status endpoints
- **Comprehensive Tests**: Unit, integration, and E2E tests
- **Security Features**: Input validation, SQL injection prevention, XSS protection

## Architecture

### Frontend
- Pure HTML/CSS/JavaScript (no framework dependencies)
- Responsive design with dark theme
- Real-time updates
- Client-side validation
- hCaptcha integration

### Backend
- Go 1.23.1+
- Gin web framework
- PostgreSQL for data persistence
- Redis for rate limiting
- hCaptcha for bot protection
- Structured logging with logrus
- Graceful shutdown
- Health checks

## Prerequisites

- Docker and Docker Compose (recommended)
- Go 1.23.1+ (for local development)
- PostgreSQL 15+ (for local development)
- Redis 7+ (for local development)
- Access to an AURA testnet node
- hCaptcha account (for production)

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
cd aura/faucet-service
```

2. Copy the environment file and configure it:
```bash
cp .env.example .env
# Edit .env with your configuration
```

Required configuration:
```env
NODE_RPC=http://aura-observer-1:26657
NODE_API=http://aura-observer-1:1317
NODE_GRPC=aura-observer-1:9090
CHAIN_ID=aura-testnet-1
FAUCET_MNEMONIC=your-mnemonic-phrase-here
FAUCET_ADDRESS=aura1...
HCAPTCHA_SECRET=your-hcaptcha-secret
```

Make sure the validator gRPC endpoint is reachable from the faucet container (set `grpc.address = "0.0.0.0:9090"` in `app.toml`).

3. Start the services:
```bash
docker-compose up -d
```

4. Access the faucet:
- Web UI: http://localhost:8080
- API: http://localhost:8080/api/v1
- Health check: http://localhost:8080/api/v1/health

### Local Development

1. Install dependencies:
```bash
cd backend
go mod download
```

2. Start PostgreSQL and Redis:
```bash
docker-compose up -d postgres redis
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run the server:
```bash
cd backend
go run main.go
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | Server port | `8080` | No |
| `ENVIRONMENT` | Environment (development/production) | `development` | No |
| `CORS_ORIGINS` | Allowed CORS origins | `*` | No |
| `NODE_RPC` | Tendermint RPC endpoint (used for tx broadcast) | `http://aura-observer-1:26657` | Yes |
| `NODE_API` | Optional REST endpoint (not required; balance uses gRPC) | `http://aura-observer-1:1317` | No |
| `NODE_GRPC` | gRPC endpoint used for account queries/balance checks | `aura-observer-1:9090` | Yes |
| `CHAIN_ID` | Chain ID | `aura-testnet-1` | Yes |
| `FAUCET_MNEMONIC` | Faucet wallet mnemonic | - | Yes |
| `FAUCET_ADDRESS` | Faucet wallet address | - | Yes |
| `DENOM` | Token denomination | `uaura` | No |
| `AMOUNT_PER_REQUEST` | Amount to send per request (in micro-units) | `100000000` (100 AURA) | No |
| `DATABASE_URL` | PostgreSQL connection string | See `.env.example` | Yes |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379/0` | Yes |
| `RATE_LIMIT_PER_IP` | Max requests per IP per window | `10` | No |
| `RATE_LIMIT_PER_ADDRESS` | Max requests per address per window | `1` | No |
| `RATE_LIMIT_WINDOW_HOURS` | Rate limit window in hours | `24` | No |
| `HCAPTCHA_SECRET` | hCaptcha secret key | - | Yes (production) |
| `GAS_LIMIT` | Transaction gas limit | `200000` | No |
| `GAS_PRICE` | Transaction gas price | `0.025uaura` | No |
| `TRANSACTION_MEMO` | Transaction memo | `AURA Testnet Faucet` | No |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | `info` | No |

### Frontend Configuration

Update the hCaptcha site key in `frontend/index.html`:
```html
<div class="h-captcha" data-sitekey="YOUR_HCAPTCHA_SITE_KEY"></div>
```

## API Documentation

### Endpoints

#### Health Check
```
GET /api/v1/health
```
Returns the health status of the faucet service.

**Response:**
```json
{
  "status": "healthy",
  "network": "aura-testnet-1",
  "height": "12345"
}
```

#### Get Faucet Info
```
GET /api/v1/faucet/info
```
Returns faucet configuration and statistics.

**Response:**
```json
{
  "amount_per_request": 100000000,
  "denom": "uaura",
  "balance": 1000000000000,
  "total_distributed": 50000000000,
  "unique_recipients": 125,
  "requests_last_24h": 45,
  "chain_id": "aura-testnet-1"
}
```

#### Request Tokens
```
POST /api/v1/faucet/request
```
Request tokens from the faucet.

**Request Body:**
```json
{
  "address": "aura1...",
  "captcha_token": "hcaptcha-token"
}
```

**Response (Success):**
```json
{
  "tx_hash": "ABCD1234...",
  "recipient": "aura1...",
  "amount": 100000000,
  "message": "Tokens sent successfully"
}
```

**Response (Rate Limited):**
```json
{
  "error": "Too many requests from your IP address. Please try again later."
}
```

#### Get Recent Transactions
```
GET /api/v1/faucet/recent
```
Returns recent faucet transactions (last 50).

**Response:**
```json
{
  "transactions": [
    {
      "recipient": "aura1...",
      "amount": 100000000,
      "tx_hash": "ABCD1234...",
      "timestamp": "2025-11-20T12:00:00Z"
    }
  ]
}
```

#### Get Statistics
```
GET /api/v1/faucet/stats
```
Returns detailed faucet statistics.

**Response:**
```json
{
  "total_requests": 1000,
  "successful_requests": 950,
  "failed_requests": 50,
  "total_distributed": 95000000000,
  "unique_recipients": 125,
  "requests_last_24h": 45,
  "requests_last_hour": 3
}
```

## Testing

### Run All Tests
```bash
cd backend
go test ./... -v
```

### Run Specific Test Suite
```bash
# Unit tests
go test ./pkg/... -v

# Integration tests
go test ./tests/integration/... -v

# E2E tests
go test ./tests/e2e/... -v
```

### Run Tests with Coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test with Local AURA Node

1. Start your AURA node:
```bash
cd /path/to/aura/chain
make install
aurad init testnode --chain-id aura-testnet-1
aurad start
```

2. Update faucet configuration:
```env
NODE_RPC=http://localhost:26657
NODE_API=http://localhost:1317
NODE_GRPC=localhost:9090
CHAIN_ID=aura-testnet-1
```

3. Test token requests via API or UI

## Database Schema

### faucet_requests Table
```sql
CREATE TABLE faucet_requests (
    id SERIAL PRIMARY KEY,
    recipient VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    tx_hash VARCHAR(255),
    ip_address VARCHAR(45) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_recipient ON faucet_requests(recipient);
CREATE INDEX idx_ip_address ON faucet_requests(ip_address);
CREATE INDEX idx_created_at ON faucet_requests(created_at);
CREATE INDEX idx_status ON faucet_requests(status);
```

## Rate Limiting

The faucet implements two-tier rate limiting:

1. **IP-based**: Limits requests from the same IP address (default: 10 per 24h)
2. **Address-based**: Limits requests to the same wallet address (default: 1 per 24h)

Rate limits are enforced using Redis with a sliding window algorithm. Both limits are checked before processing a request.

## Security Features

### Built-in Security
- hCaptcha verification in production mode
- Two-tier rate limiting (IP and address-based)
- Input validation and sanitization
- SQL injection prevention (parameterized queries)
- XSS protection
- CORS configuration
- Request audit logging
- Error message sanitization

### Recommendations
- Use strong database passwords
- Enable SSL/TLS for production
- Configure firewall rules
- Regular security audits
- Monitor logs for suspicious activity
- Keep dependencies updated
- Use secrets management for sensitive data

## Monitoring

### Health Checks
- API health endpoint: `/api/v1/health`
- Docker health checks for all services
- Database connection monitoring
- Redis connection monitoring
- Blockchain node status monitoring

### Logging
Structured JSON logging with different log levels:
- `debug`: Detailed debugging information
- `info`: General information (default)
- `warn`: Warning messages
- `error`: Error messages

**Example log entry:**
```json
{
  "level": "info",
  "msg": "Token request received",
  "address": "aura1...",
  "ip": "192.168.1.1",
  "time": "2025-11-20T12:00:00Z"
}
```

### Metrics to Monitor
- Request rate (requests per minute)
- Success/failure rate
- Response times
- Faucet balance
- Database connection pool
- Redis memory usage
- Error rates

## Troubleshooting

### Common Issues

1. **Cannot connect to database**
   - Check PostgreSQL is running: `docker-compose ps postgres`
   - Verify DATABASE_URL configuration
   - Check network connectivity
   - View logs: `docker-compose logs postgres`

2. **Cannot connect to Redis**
   - Check Redis is running: `docker-compose ps redis`
   - Verify REDIS_URL configuration
   - Test connection: `redis-cli ping`
   - View logs: `docker-compose logs redis`

3. **Transactions failing**
   - Check faucet account balance
   - Verify NODE_RPC/NODE_API/NODE_GRPC endpoints are accessible
   - Check faucet mnemonic/address configuration
   - Ensure node is synced and not catching up

4. **Captcha not working**
   - Verify hCaptcha site key in frontend
   - Check HCAPTCHA_SECRET in backend
   - Ensure ENVIRONMENT=production
   - Check network access to hcaptcha.com

5. **Build errors**
   ```bash
   cd backend
   go mod tidy
   go build -o ../bin/faucet-server main.go
   ```

## Production Deployment

### Pre-deployment Checklist

- [ ] Update `.env` with production values
- [ ] Set strong database password
- [ ] Configure hCaptcha keys
- [ ] Set up SSL/TLS certificates
- [ ] Configure firewall rules
- [ ] Set up monitoring and alerts
- [ ] Configure backup strategy
- [ ] Review and update rate limits
- [ ] Set `ENVIRONMENT=production`
- [ ] Set `LOG_LEVEL=info` or `warn`
- [ ] Update CORS_ORIGINS to specific domains
- [ ] Test all endpoints
- [ ] Verify faucet has sufficient balance

### Using Docker Compose

1. Configure production environment:
```bash
cp .env.example .env
# Update all values for production
```

2. Generate SSL certificates (optional, if using nginx):
```bash
mkdir -p ssl
# Add your SSL certificates
# ssl/cert.pem
# ssl/key.pem
```

3. Start with production profile:
```bash
docker-compose --profile production up -d
```

4. Verify deployment:
```bash
docker-compose ps
docker-compose logs -f faucet-backend
curl http://localhost:8080/api/v1/health
```

### Monitoring Production

```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f faucet-backend

# Check resource usage
docker stats

# View recent transactions
curl http://localhost:8080/api/v1/faucet/recent
```

## Maintenance

### Database Backup
```bash
docker-compose exec postgres pg_dump -U faucet faucet > backup_$(date +%Y%m%d).sql
```

### Database Restore
```bash
cat backup_20251120.sql | docker-compose exec -T postgres psql -U faucet faucet
```

### Update Application
```bash
# Pull latest changes
git pull

# Rebuild and restart
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Verify
docker-compose logs -f faucet-backend
```

### Clear Rate Limits (Emergency)
```bash
docker-compose exec redis redis-cli FLUSHDB
```

## Development

### Project Structure
```
faucet-service/
├── frontend/              # Web UI
│   ├── index.html        # Main page
│   ├── styles.css        # Styling
│   └── app.js            # Application logic
├── backend/              # Go API server
│   ├── main.go          # Entry point
│   ├── pkg/             # Packages
│   │   ├── api/         # HTTP handlers
│   │   ├── config/      # Configuration
│   │   ├── database/    # Database layer
│   │   ├── faucet/      # Faucet service
│   │   └── ratelimit/   # Rate limiting
│   ├── tests/           # Tests
│   ├── go.mod           # Go modules
│   └── Dockerfile       # Container build
├── scripts/             # Utility scripts
├── docker-compose.yml   # Deployment config
├── nginx.conf          # Nginx configuration
├── .env.example        # Environment template
└── README.md          # This file
```

### Adding Features

1. Backend changes: Update relevant files in `backend/pkg/`
2. Frontend changes: Update `frontend/` files
3. Add tests: Create test files with `_test.go` suffix
4. Update documentation
5. Test locally
6. Create pull request

## Support

- **Documentation**: This README and inline code comments
- **Issues**: GitHub Issues
- **AURA Chain**: [AURA GitHub](https://github.com/aura-chain/aura)

## License

This project is licensed under the MIT License.

## Acknowledgments

- Based on the PAW testnet faucet implementation
- Integrated with AURA blockchain
- Uses Cosmos SDK standards

---

**Version**: 1.0.0
**Last Updated**: 2025-11-20
**Status**: Production Ready
