---
sidebar_position: 2
---

# API Reference

AURA provides multiple ways to interact with the blockchain.

## REST API

The REST API is available at `https://testnet-api.aurablockchain.org`.

### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/cosmos/bank/v1beta1/balances/{address}` | GET | Query account balances |
| `/cosmos/staking/v1beta1/validators` | GET | List all validators |
| `/aura/credential/v1/credentials` | GET | List all credentials |
| `/aura/credential/v1/credentials/{id}` | GET | Get credential by ID |

### Example

```bash
# Get account balances
curl -s https://testnet-api.aurablockchain.org/cosmos/bank/v1beta1/balances/aura1...

# List credentials
curl -s https://testnet-api.aurablockchain.org/aura/credential/v1/credentials
```

## gRPC

The gRPC endpoint is available at `testnet-grpc.aurablockchain.org:443`.

### Using grpcurl

```bash
# List services
grpcurl testnet-grpc.aurablockchain.org:443 list

# Query credentials
grpcurl testnet-grpc.aurablockchain.org:443 aura.credential.v1.Query/Credentials
```

## RPC

The Tendermint RPC is available at `https://testnet-rpc.aurablockchain.org`.

### Endpoints

| Endpoint | Description |
|----------|-------------|
| `/status` | Node status |
| `/block` | Latest block |
| `/validators` | Current validator set |
| `/tx?hash=0x...` | Transaction by hash |

### Example

```bash
# Get node status
curl -s https://testnet-rpc.aurablockchain.org/status | jq '.result.sync_info'

# Get latest block
curl -s https://testnet-rpc.aurablockchain.org/block | jq '.result.block.header.height'
```

## WebSocket

WebSocket connections are available at `wss://testnet-ws.aurablockchain.org`.

### Subscribe to Events

```javascript
const ws = new WebSocket('wss://testnet-ws.aurablockchain.org/websocket');

ws.onopen = () => {
  ws.send(JSON.stringify({
    jsonrpc: '2.0',
    method: 'subscribe',
    params: ["tm.event='NewBlock'"],
    id: 1
  }));
};

ws.onmessage = (event) => {
  console.log(JSON.parse(event.data));
};
```

## SDKs

- [JavaScript SDK](https://github.com/aura-blockchain/aura-sdk)
- [Go SDK](https://github.com/aura-blockchain/aura/tree/main/sdk/go)
