# Identity Management Tutorial

## Overview
Create and manage decentralized identities (DIDs) on Aura.

## Prerequisites
```bash
# Create a key
aurad keys add mykey --keyring-backend test

# Get testnet tokens
curl -X POST https://faucet.aura.network/request \
  -d '{"address": "aura1..."}'
```

## Create Identity Record

### CLI
```bash
aurad tx identity create-record \
  --did "did:aura:$(aurad keys show mykey -a)" \
  --metadata-hash "sha256:abc123" \
  --from mykey \
  --chain-id aura-testnet-1
```

### REST API
```bash
curl -X POST http://localhost:1317/aura/identity/v1beta1/records \
  -H "Content-Type: application/json" \
  -d '{
    "did": "did:aura:aura1...",
    "owner": "aura1...",
    "metadata_hash": "sha256:abc123"
  }'
```

### Go SDK
```go
import "github.com/aura-chain/aura/sdk/go/pkg/modules/identity"

client := identity.NewClient(auraClient)
resp, err := client.CreateIdentityRecord(ctx, &identity.CreateRecordParams{
    DID:          "did:aura:aura1...",
    MetadataHash: "sha256:abc123",
})
```

## Query Identity

### CLI
```bash
aurad query identity record did:aura:aura1...
```

### REST
```bash
curl http://localhost:1317/aura/identity/v1beta1/records/did:aura:aura1...
```

## Request Identity Change

```bash
aurad tx identity request-change \
  --target-did "did:aura:aura1..." \
  --metadata-hash "sha256:newmeta" \
  --from mykey
```

## Assign Roles

```bash
aurad tx identity assign-role \
  --address aura1user... \
  --role admin \
  --from mykey
```
