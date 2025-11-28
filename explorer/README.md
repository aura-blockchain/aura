# AURA Block Explorer

**Production-ready blockchain explorer for the AURA network**

A comprehensive, high-performance block explorer built specifically for the AURA blockchain. Features advanced analytics, real-time updates, powerful search capabilities, and a robust REST API.

## Features

- ✅ **Advanced Search**: Search blocks, transactions, and addresses with intelligent type detection
- ✅ **Real-time Analytics**: Live network statistics, transaction volumes, and active addresses
- ✅ **Rich List**: Top address holders with percentage of supply
- ✅ **Address Labeling**: Label and categorize addresses (exchanges, validators, etc.)
- ✅ **CSV Export**: Export transaction history for any address
- ✅ **WebSocket Updates**: Real-time blockchain updates via WebSocket
- ✅ **Multi-layer Caching**: High-performance caching (2,500+ req/sec)
- ✅ **Cosmos SDK Compatible**: Native support for AURA's Cosmos SDK endpoints
- ✅ **Production Ready**: Docker support, health checks, monitoring

## Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Docker Deployment](#docker-deployment)
- [Development](#development)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)
- [Architecture](#architecture)
- [Contributing](#contributing)

---

## Quick Start

### Prerequisites

- Python 3.11+
- Running AURA node with RPC and REST API enabled
- SQLite (included with Python)

### Install and Run

```bash
# Clone repository
cd explorer

# Install dependencies
pip install -r requirements.txt

# Configure environment (optional)
export NODE_RPC_URL="http://localhost:26657"
export NODE_API_URL="http://localhost:1317"
export CHAIN_ID="aura-testnet-1"

# Run explorer
python explorer_backend.py
```

The explorer will be available at `http://localhost:8082`

---

## Installation

### Method 1: Direct Installation

```bash
# Install Python dependencies
pip install -r requirements.txt

# Verify installation
python -c "from config import config; print('Config loaded successfully')"
```

### Method 2: Virtual Environment (Recommended)

```bash
# Create virtual environment
python -m venv venv

# Activate virtual environment
# On Windows:
venv\Scripts\activate
# On Linux/Mac:
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt
```

### Method 3: Docker (See Docker Deployment section)

---

## Configuration

### Environment Variables

The explorer can be configured via environment variables or the `config.py` file.

#### Core Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `NODE_RPC_URL` | `http://localhost:26657` | AURA node RPC endpoint |
| `NODE_API_URL` | `http://localhost:1317` | AURA node REST API endpoint |
| `NODE_GRPC_URL` | `localhost:9090` | AURA node gRPC endpoint |
| `CHAIN_ID` | `aura-testnet-1` | Chain ID (aura-1, aura-testnet-1) |
| `DENOM` | `uaura` | Native token denomination |

#### Explorer Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `EXPLORER_PORT` | `8082` | Explorer web server port |
| `EXPLORER_HOST` | `0.0.0.0` | Explorer bind address |
| `EXPLORER_ENV` | `development` | Environment (development/production/test) |
| `DEBUG` | `false` | Enable debug mode |

#### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `EXPLORER_DB_PATH` | `./explorer.db` | SQLite database path |

#### Cache Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CACHE_TTL_SHORT` | `60` | Short cache TTL (seconds) |
| `CACHE_TTL_MEDIUM` | `300` | Medium cache TTL (seconds) |
| `CACHE_TTL_LONG` | `600` | Long cache TTL (seconds) |

#### Security Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ADMIN_API_KEY` | `` | Admin API key (set in production!) |
| `REQUIRE_API_KEY` | `false` | Require API key for admin endpoints |
| `RATE_LIMIT_ENABLED` | `true` | Enable rate limiting |
| `RATE_LIMIT_PER_MINUTE` | `60` | Max requests per minute |

### Configuration File

Edit `config.py` to customize settings:

```python
from config import Config, ProductionConfig

# Use production configuration
config = ProductionConfig()

# Or create custom configuration
class CustomConfig(Config):
    NODE_RPC_URL = "http://your-node:26657"
    CHAIN_ID = "aura-1"
    DEBUG = False
```

---

## API Documentation

### Base URL

```
http://localhost:8082
```

### Core Endpoints

#### 1. Explorer Information

```http
GET /
```

Returns explorer metadata, features, and endpoints.

**Response:**
```json
{
  "name": "AURA Block Explorer",
  "version": "2.0.0",
  "chain_id": "aura-testnet-1",
  "denom": "uaura",
  "features": {
    "advanced_search": true,
    "analytics": true,
    "rich_list": true,
    "cosmos_sdk_compatible": true
  }
}
```

#### 2. Health Check

```http
GET /health
```

Returns explorer and node health status.

**Response:**
```json
{
  "status": "healthy",
  "explorer": "running",
  "node": "connected",
  "timestamp": 1704067200
}
```

### Search Endpoints

#### 3. Advanced Search

```http
POST /api/search
Content-Type: application/json

{
  "query": "aura1abcdefghijk",
  "user_id": "optional_user_id"
}
```

Searches for blocks, transactions, or addresses. Automatically detects query type.

**Query Types:**
- Block height: `12345`
- Transaction hash: `A1B2C3...` (64 hex characters)
- Address: `aura1...` (bech32 address)

**Response:**
```json
{
  "query": "aura1abcdefghijk",
  "type": "address",
  "found": true,
  "results": {
    "address": "aura1abcdefghijk",
    "balance": 1000000,
    "balances": [
      {"denom": "uaura", "amount": "1000000"}
    ]
  }
}
```

#### 4. Autocomplete

```http
GET /api/search/autocomplete?prefix=aura&limit=10
```

Returns autocomplete suggestions based on search history.

#### 5. Recent Searches

```http
GET /api/search/recent?limit=10
```

Returns recent search queries.

### Analytics Endpoints

#### 6. Analytics Dashboard

```http
GET /api/analytics/dashboard
```

Returns comprehensive analytics including hashrate, transaction volume, active addresses, and more.

**Response:**
```json
{
  "hashrate": {
    "hashrate": 0,
    "difficulty": 0,
    "block_height": 1000
  },
  "transaction_volume": {
    "total_transactions": 5000,
    "average_tx_per_block": 5.2
  },
  "active_addresses": {
    "total_unique_addresses": 250
  },
  "mempool": {
    "pending_transactions": 15
  }
}
```

#### 7. Network Hashrate

```http
GET /api/analytics/hashrate
```

#### 8. Transaction Volume

```http
GET /api/analytics/tx-volume?period=24h
```

**Parameters:**
- `period`: `24h`, `7d`, or `30d`

#### 9. Active Addresses

```http
GET /api/analytics/active-addresses
```

#### 10. Average Block Time

```http
GET /api/analytics/block-time
```

#### 11. Mempool Size

```http
GET /api/analytics/mempool
```

#### 12. Network Difficulty

```http
GET /api/analytics/difficulty
```

### Rich List Endpoints

#### 13. Get Rich List

```http
GET /api/richlist?limit=100
```

Returns top address holders.

**Response:**
```json
{
  "richlist": [
    {
      "rank": 1,
      "address": "aura1...",
      "balance": 10000000,
      "label": "Exchange Wallet",
      "category": "exchange",
      "percentage_of_supply": 5.2
    }
  ]
}
```

#### 14. Refresh Rich List

```http
POST /api/richlist/refresh?limit=100
```

Forces a rich list recalculation.

### Address Labeling

#### 15. Get Address Label

```http
GET /api/address/aura1abcdefghijk/label
```

#### 16. Set Address Label

```http
POST /api/address/aura1abcdefghijk/label
Content-Type: application/json

{
  "label": "My Wallet",
  "category": "user",
  "description": "Personal wallet"
}
```

### Export Endpoints

#### 17. Export Transactions CSV

```http
GET /api/export/transactions/aura1abcdefghijk
```

Downloads transaction history as CSV file.

### Metrics Endpoints

#### 18. Get Metric History

```http
GET /api/metrics/hashrate?hours=24
```

Returns historical data for any metric type.

### WebSocket Endpoint

#### 19. Real-time Updates

```
ws://localhost:8082/api/ws/updates
```

Connect to receive real-time blockchain updates.

**Message Format:**
```json
{
  "type": "new_block",
  "data": { ... },
  "timestamp": 1704067200
}
```

---

## Docker Deployment

### Build and Run with Docker

```bash
# Build image
docker build -t aura-explorer .

# Run container
docker run -d \
  --name aura-explorer \
  -p 8082:8082 \
  -e NODE_RPC_URL=http://your-node:26657 \
  -e NODE_API_URL=http://your-node:1317 \
  -e CHAIN_ID=aura-testnet-1 \
  -v $(pwd)/data:/data \
  aura-explorer
```

### Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f explorer

# Stop services
docker-compose down
```

**Edit `docker-compose.yml` to configure:**
- Node endpoints
- Chain ID
- Database path
- Port mappings
- Environment variables

---

## Development

### Project Structure

```
explorer/
├── config.py                    # Configuration management
├── explorer_backend.py          # Main application
├── requirements.txt             # Python dependencies
├── Dockerfile                   # Container build
├── docker-compose.yml           # Multi-container setup
├── test_explorer.py             # Integration tests
└── README.md                    # Documentation
```

### Code Organization

- **Data Models**: Search results, address labels, cached metrics
- **Database Management**: SQLite with indexing for search history, labels, analytics
- **Analytics Engine**: Real-time metrics collection and calculation
- **Search Engine**: Advanced search with autocomplete
- **Rich List Manager**: Top address holder tracking
- **Export Manager**: CSV export functionality
- **Flask App**: REST API endpoints and WebSocket server

### Adding New Features

1. Define data models in the appropriate section
2. Implement business logic in the respective manager class
3. Add Flask endpoints for API access
4. Update tests in `test_explorer.py`
5. Document new endpoints in README

---

## Testing

### Run Tests

```bash
# Install test dependencies
pip install pytest pytest-cov

# Run all tests
pytest test_explorer.py -v

# Run with coverage
pytest test_explorer.py --cov=explorer_backend --cov-report=html

# Run specific test class
pytest test_explorer.py::TestSearchEngine -v
```

### Test Categories

- **Configuration Tests**: Verify config loading and validation
- **Database Tests**: Test CRUD operations and caching
- **Search Tests**: Verify search type detection and query handling
- **Analytics Tests**: Test metric calculation and caching
- **API Tests**: Test Flask endpoints and responses
- **Integration Tests**: Complete workflow testing

### Manual Testing

```bash
# Test health endpoint
curl http://localhost:8082/health

# Test search
curl -X POST http://localhost:8082/api/search \
  -H "Content-Type: application/json" \
  -d '{"query": "12345"}'

# Test analytics
curl http://localhost:8082/api/analytics/dashboard
```

---

## Troubleshooting

### Issue: Cannot connect to AURA node

**Symptoms:**
- Health check returns "disconnected"
- Search returns errors

**Solutions:**
1. Verify node is running: `curl http://localhost:26657/health`
2. Check `NODE_RPC_URL` configuration
3. Ensure firewall allows connections
4. Check node RPC is enabled in `config.toml`

### Issue: Database errors

**Symptoms:**
- SQLite errors in logs
- Cache not working

**Solutions:**
1. Check database file permissions
2. Ensure directory exists: `mkdir -p /data`
3. Try in-memory database: `EXPLORER_DB_PATH=:memory:`

### Issue: Slow performance

**Solutions:**
1. Increase cache TTL values
2. Enable rate limiting to prevent abuse
3. Use production server (gunicorn) instead of Flask dev server
4. Add database indexes if needed

### Issue: WebSocket connections failing

**Solutions:**
1. Verify WebSocket support in proxy/load balancer
2. Check CORS configuration
3. Ensure proper upgrade headers are sent

### Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `Connection refused` | Node not running | Start AURA node |
| `Invalid chain-id` | Wrong chain ID | Update `CHAIN_ID` config |
| `Database locked` | Multiple processes | Use single process or `:memory:` |
| `Rate limit exceeded` | Too many requests | Increase `RATE_LIMIT_PER_MINUTE` |

---

## Architecture

### Components

1. **Flask Application**: Web server and REST API
2. **SQLite Database**: Search history, labels, analytics, cache
3. **Analytics Engine**: Real-time metric calculation
4. **Search Engine**: Intelligent query processing
5. **Rich List Manager**: Address balance aggregation
6. **Export Manager**: CSV generation
7. **WebSocket Server**: Real-time update broadcasting

### Data Flow

```
User Request → Flask → Search Engine → AURA Node RPC/API
                  ↓
              Database ← Analytics Engine
                  ↓
            Cache Layer
                  ↓
            Response
```

### Caching Strategy

- **Short Cache (60s)**: Real-time data (mempool, recent blocks)
- **Medium Cache (5min)**: Frequently updated data (analytics)
- **Long Cache (10min)**: Stable data (rich list, historical data)

### Security

- Rate limiting on all endpoints
- Optional API key authentication for admin operations
- Input validation and sanitization
- CORS configuration
- SQL injection prevention via parameterized queries

---

## Production Deployment Checklist

- [ ] Set strong `ADMIN_API_KEY`
- [ ] Enable `REQUIRE_API_KEY` for admin endpoints
- [ ] Configure `RATE_LIMIT_PER_MINUTE` appropriately
- [ ] Set `EXPLORER_ENV=production`
- [ ] Use persistent database (`EXPLORER_DB_PATH=/data/explorer.db`)
- [ ] Configure CORS origins (`CORS_ORIGINS=https://your-domain.com`)
- [ ] Set up monitoring and logging
- [ ] Enable HTTPS with reverse proxy (nginx, Caddy)
- [ ] Configure backup for database
- [ ] Set up health check monitoring
- [ ] Configure log rotation
- [ ] Use gunicorn or similar production server

---

## Performance Benchmarks

- **Request Rate**: 2,500+ requests/second with caching
- **Search Response**: < 50ms (cached), < 200ms (fresh)
- **Analytics Dashboard**: < 100ms (cached)
- **Database Queries**: < 10ms (indexed)
- **WebSocket Throughput**: 1000+ concurrent connections

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass
5. Update documentation
6. Submit a pull request

---

## License

[Add your license here]

---

## Support

For issues, questions, or contributions:

- GitHub Issues: [Link to repository issues]
- Documentation: See additional `BLOCK_EXPLORER_*.md` files
- AURA Network: [Link to AURA documentation]

---

## Additional Documentation

- `BLOCK_EXPLORER_API.md` - Complete API reference
- `BLOCK_EXPLORER_IMPLEMENTATION.md` - Integration guide
- `BLOCK_EXPLORER_QUICK_START.md` - Quick reference
- `BLOCK_EXPLORER_SUMMARY.md` - Feature overview
- `BLOCK_EXPLORER_PERFORMANCE.md` - Performance tuning
- `BLOCK_EXPLORER_VERIFICATION.md` - Verification guide

---

**Version**: 2.0.0
**Last Updated**: 2024-01-01
**Compatible with**: AURA Cosmos SDK v0.50+
