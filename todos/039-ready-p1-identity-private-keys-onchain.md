---
id: "039"
title: "Privacy Module Private View Keys Stored On-Chain"
status: ready
priority: p1
category: security
module: privacy
severity: CRITICAL
cvss: 10.0
source: identity-privacy-audit
---

# Privacy Module Private View Keys Stored On-Chain

## Problem

Private view keys are stored directly on the blockchain, defeating the entire purpose of the privacy module. Anyone can read the chain state and see all "private" information.

## Affected Files

- `chain/x/privacy/types/privacy.proto`
- `chain/x/privacy/keeper/msg_server.go`

## Vulnerability

```protobuf
// From privacy.proto
message ShieldedAccount {
    string address = 1;
    bytes public_key = 2;
    bytes private_view_key = 3;  // CRITICAL: PRIVATE KEY ON CHAIN!
    bytes private_spend_key = 4; // CRITICAL: PRIVATE KEY ON CHAIN!
    // ...
}
```

## Impact

- **Complete privacy failure** - All "private" transactions visible
- Private view keys allow anyone to see transaction details
- Private spend keys allow anyone to SPEND funds
- Entire privacy system is theater, not actual privacy

## Required Fix

**NEVER store private keys on chain.** Redesign the privacy system:

```protobuf
// Fixed: Only public keys on chain
message ShieldedAccount {
    string address = 1;
    bytes public_view_key = 2;   // Public key for creating shielded payments
    bytes public_spend_key = 3;  // Public key for verification
    bytes commitment = 4;        // Pedersen commitment to balance
    // NO PRIVATE KEYS
}
```

```go
// Client-side key derivation (never touches chain)
type PrivacyClient struct {
    // Keys derived from seed phrase, stored in local wallet
    privateViewKey  []byte
    privateSpendKey []byte
    publicViewKey   []byte
    publicSpendKey  []byte
}

func (c *PrivacyClient) GenerateKeys(seed []byte) {
    // Derive keys locally
    c.privateSpendKey = hash(seed || "spend")
    c.privateViewKey = hash(seed || "view")
    c.publicSpendKey = scalarMultBase(c.privateSpendKey)
    c.publicViewKey = scalarMultBase(c.privateViewKey)
}

// On-chain: Only register public keys
func (ms msgServer) CreateShieldedAccount(goCtx context.Context, msg *privacypb.MsgCreateShieldedAccount) (*privacypb.MsgCreateShieldedAccountResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // ONLY store public keys
    account := &privacypb.ShieldedAccount{
        Address:        msg.Address,
        PublicViewKey:  msg.PublicViewKey,  // NOT private!
        PublicSpendKey: msg.PublicSpendKey, // NOT private!
        Commitment:     msg.InitialCommitment,
    }

    // Validate keys are valid public keys (on curve)
    if !isValidPublicKey(msg.PublicViewKey) {
        return nil, status.Error(codes.InvalidArgument, "invalid public view key")
    }

    ms.Keeper.SetShieldedAccount(ctx, account)

    return &privacypb.MsgCreateShieldedAccountResponse{}, nil
}
```

## Acceptance Criteria

- [ ] All private keys removed from proto definitions
- [ ] All private key storage removed from keeper
- [ ] Documentation clarifies client-side key management
- [ ] Migration to remove any stored private keys
- [ ] Tests verify no private keys in state
- [ ] Security audit of new design
