# Aura Testnet Endpoints

**Chain ID:** `aura-testnet-1`
**Network:** Testnet
**Status:** Live

---

## Public Endpoints

### RPC (Tendermint/CometBFT)
```
https://rpc.aura-testnet.com:26657
```
- WebSocket: `wss://rpc.aura-testnet.com:26657/websocket`
- Used for: Block queries, transaction broadcast, event subscriptions

### REST API (Cosmos SDK)
```
https://api.aura-testnet.com:1317
```
- OpenAPI docs: `/swagger/`
- Used for: Account queries, token balances, module state

### gRPC
```
grpc.aura-testnet.com:9090
```
- TLS enabled
- Used for: High-performance queries, streaming

---

## Local Development Endpoints

When running a local node:

| Service | Endpoint |
|---------|----------|
| RPC | `http://localhost:26657` |
| REST API | `http://localhost:1317` |
| gRPC | `localhost:9090` |
| gRPC-Web | `localhost:9091` |
| Prometheus | `localhost:26660` |
| pprof | `localhost:6060` |

---

## User Services

| Service | URL |
|---------|-----|
| Block Explorer | https://explorer.aura-testnet.com |
| Testnet Faucet | https://faucet.aura-testnet.com |
| Developer Playground | https://playground.aura-testnet.com |
| Status Page | https://status.aura-testnet.com |

---

## Chain Configuration

### For Keplr Wallet
```javascript
{
  chainId: "aura-testnet-1",
  chainName: "Aura Testnet",
  rpc: "https://rpc.aura-testnet.com:26657",
  rest: "https://api.aura-testnet.com:1317",
  bip44: { coinType: 118 },
  bech32Config: {
    bech32PrefixAccAddr: "aura",
    bech32PrefixAccPub: "aurapub",
    bech32PrefixValAddr: "auravaloper",
    bech32PrefixValPub: "auravaloperpub",
    bech32PrefixConsAddr: "auravalcons",
    bech32PrefixConsPub: "auravalconspub"
  },
  currencies: [{
    coinDenom: "AURA",
    coinMinimalDenom: "uaura",
    coinDecimals: 6
  }],
  feeCurrencies: [{
    coinDenom: "AURA",
    coinMinimalDenom: "uaura",
    coinDecimals: 6,
    gasPriceStep: { low: 0.015, average: 0.025, high: 0.04 }
  }],
  stakeCurrency: {
    coinDenom: "AURA",
    coinMinimalDenom: "uaura",
    coinDecimals: 6
  }
}
```

### Genesis File
```
https://raw.githubusercontent.com/aura-chain/networks/main/aura-testnet-1/genesis.json
```

### Persistent Peers
```
<node-id>@seed1.aura-testnet.com:26656,
<node-id>@seed2.aura-testnet.com:26656
```

---

## SDK Configuration

### JavaScript (CosmJS)
```javascript
import { SigningStargateClient } from "@cosmjs/stargate";

const rpcEndpoint = "https://rpc.aura-testnet.com:26657";
const client = await SigningStargateClient.connectWithSigner(rpcEndpoint, signer);
```

### Python
```python
from cosmospy import cosmos

client = cosmos.Client(
    lcd_url="https://api.aura-testnet.com:1317",
    rpc_url="https://rpc.aura-testnet.com:26657"
)
```

### Go
```go
import "github.com/aura-chain/aura/sdk/go/client"

cfg := client.Config{
    RPCEndpoint: "https://rpc.aura-testnet.com:26657",
    GRPCEndpoint: "grpc.aura-testnet.com:9090",
    ChainID: "aura-testnet-1",
}
auraClient, err := client.NewClient(cfg)
```

---

## Rate Limits

| Endpoint | Limit |
|----------|-------|
| RPC | 100 req/min per IP |
| REST API | 60 req/min per IP |
| Faucet | 1 request/24h per address |

---

## Support

- Issues: https://github.com/aura-chain/aura/issues
- Discord: https://discord.gg/aura
- Email: support@aura.dev
