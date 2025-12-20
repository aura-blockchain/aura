---
sidebar_position: 3
---

# SDK Integration

Integrate Aura blockchain functionality into your applications using official and community SDKs.

## JavaScript/TypeScript SDK

### Installation

```bash
npm install @aura/sdk
# or
yarn add @aura/sdk
```

### Basic Usage

```javascript
import { AuraClient } from '@aura/sdk';

// Connect to network
const client = new AuraClient({
  rpcEndpoint: 'https://rpc.aura.network',
  chainId: 'aura-mainnet-1'
});

await client.connect();

// Query balance
const balance = await client.getBalance('aura1...');
console.log(balance);

// Send transaction
const result = await client.sendTokens(
  fromAddress,
  toAddress,
  [{ denom: 'uaura', amount: '1000000' }],
  { memo: 'Payment' }
);
```

### Working with Credentials

```javascript
// Issue credential
const credential = await client.vcregistry.issueCredential({
  issuerDID: 'did:aura:issuer123',
  subjectDID: 'did:aura:subject456',
  credentialType: 'ProofOfHumanity',
  claims: { verified: true }
});

// Query credential
const vc = await client.vcregistry.getCredential(credentialId);

// Verify credential
const isValid = await client.vcregistry.verifyCredential(vc);
```

## Python SDK

### Installation

```bash
pip install aura-py
```

### Basic Usage

```python
from aura import AuraClient

# Initialize client
client = AuraClient(
    rpc_endpoint='https://rpc.aura.network',
    chain_id='aura-mainnet-1'
)

# Query balance
balance = client.get_balance('aura1...')
print(balance)

# Send transaction
tx = client.send_tokens(
    from_address='aura1...',
    to_address='aura1...',
    amount=1000000,
    denom='uaura'
)
```

## Go Integration

### Using Cosmos SDK

```go
import (
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/types"
)

// Create client context
clientCtx := client.Context{}.
    WithNodeURI("https://rpc.aura.network").
    WithChainID("aura-mainnet-1")

// Query balance
queryClient := banktypes.NewQueryClient(clientCtx)
res, err := queryClient.Balance(context.Background(), &banktypes.QueryBalanceRequest{
    Address: "aura1...",
    Denom:   "uaura",
})
```

## REST API

### Query Balance

```bash
curl https://api.aura.network/cosmos/bank/v1beta1/balances/aura1...
```

### Send Transaction

```bash
curl -X POST https://api.aura.network/cosmos/tx/v1beta1/txs \
  -H "Content-Type: application/json" \
  -d '{
    "tx_bytes": "...",
    "mode": "BROADCAST_MODE_SYNC"
  }'
```

## gRPC API

### Connect to gRPC

```go
conn, err := grpc.Dial(
    "grpc.aura.network:9090",
    grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
)
defer conn.Close()

client := banktypes.NewQueryClient(conn)
```

### Query Data

```go
res, err := client.Balance(ctx, &banktypes.QueryBalanceRequest{
    Address: "aura1...",
    Denom:   "uaura",
})
```

## WebSocket Subscriptions

### Subscribe to Events

```javascript
const ws = new WebSocket('wss://rpc.aura.network/websocket');

ws.on('open', () => {
  // Subscribe to new blocks
  ws.send(JSON.stringify({
    jsonrpc: '2.0',
    method: 'subscribe',
    params: ["tm.event='NewBlock'"],
    id: 1
  }));
});

ws.on('message', (data) => {
  const event = JSON.parse(data);
  console.log('New block:', event);
});
```

## Authentication

### Signing Transactions

```javascript
import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';

// Create wallet from mnemonic
const wallet = await DirectSecp256k1HdWallet.fromMnemonic(
  mnemonic,
  { prefix: 'aura' }
);

// Get accounts
const [account] = await wallet.getAccounts();

// Sign and broadcast
const result = await client.signAndBroadcast(
  account.address,
  [msg],
  fee,
  memo
);
```

## Error Handling

```javascript
try {
  const result = await client.sendTokens(...);
} catch (error) {
  if (error.code === 'INSUFFICIENT_FUNDS') {
    console.error('Not enough balance');
  } else if (error.code === 'INVALID_ADDRESS') {
    console.error('Invalid recipient address');
  } else {
    console.error('Transaction failed:', error.message);
  }
}
```

## Best Practices

- Always validate addresses before sending transactions
- Use environment variables for sensitive data (mnemonics, keys)
- Implement proper error handling
- Cache frequently accessed data
- Use pagination for large queries
- Monitor gas costs and optimize
- Test on testnet before mainnet

## Resources

- [JavaScript SDK Documentation](https://sdk-docs.aura.network/js)
- [Python SDK Documentation](https://sdk-docs.aura.network/python)
- [REST API Reference](https://api-docs.aura.network)
- [gRPC API Reference](https://grpc-docs.aura.network)
- [Example Applications](https://github.com/aura-blockchain/examples)

## Support

- [SDK Issues](https://github.com/aura-blockchain/sdk/issues)
- [Discord - Developer Channel](https://discord.gg/aura)
- [Stack Overflow](https://stackoverflow.com/questions/tagged/aura)
