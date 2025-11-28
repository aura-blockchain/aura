# AURA Block Explorer

Production-ready blockchain explorer for AURA network, adapted from XAI blockchain.

## Quick Start

### Configuration

Update the explorer backend with AURA endpoints:

```python
# In explorer_backend.py, update these variables:
NODE_RPC_URL = "http://localhost:26657"  # Your AURA node RPC
CHAIN_ID = "aura-1"  # Your AURA chain ID
EXPLORER_PORT = 8082
```

### Install Dependencies

```bash
pip install flask flask-cors flask-sock requests
```

### Run Explorer

```bash
python explorer_backend.py
```

The explorer will be available at http://localhost:8082

## Features

- ✅ Advanced search (blocks, transactions, addresses)
- ✅ Real-time analytics dashboard
- ✅ Rich list / Top holders
- ✅ Address labeling system
- ✅ CSV export functionality
- ✅ WebSocket real-time updates
- ✅ Multi-layer caching (2,500+ req/sec)
- ✅ SQLite database with indexing

## API Endpoints

- `GET /health` - Health check
- `GET /api/analytics/dashboard` - All metrics
- `POST /api/search` - Advanced search
- `GET /api/richlist` - Top addresses
- `GET /api/export/transactions/{address}` - CSV export
- `ws://localhost:8082/api/ws/updates` - WebSocket updates

## Documentation

See the following files for complete documentation:
- `BLOCK_EXPLORER_API.md` - Complete API reference
- `BLOCK_EXPLORER_IMPLEMENTATION.md` - Integration guide
- `BLOCK_EXPLORER_QUICK_START.md` - Quick reference
- `BLOCK_EXPLORER_SUMMARY.md` - Feature overview

## Docker Deployment

```bash
docker build -t aura-explorer .
docker run -p 8082:8082 \
  -e NODE_RPC_URL=http://your-node:26657 \
  -e CHAIN_ID=aura-1 \
  aura-explorer
```
