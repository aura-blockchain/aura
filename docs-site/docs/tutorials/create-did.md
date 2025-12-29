---
sidebar_position: 2
---

# Create a DID

**Difficulty:** Beginner | **Time:** 5 minutes

Create your Decentralized Identifier (DID) on Aura.

## What is a DID?

A DID is a globally unique identifier that you control. Unlike traditional identifiers (email, phone), DIDs are:
- Self-sovereign (you own it)
- Cryptographically verifiable
- Not dependent on any central authority

## Steps

### 1. Generate a Wallet

```bash
aurad keys add my-wallet
```

Save the mnemonic phrase securely.

### 2. Fund Your Wallet

Get testnet tokens:
```bash
# Get your address
aurad keys show my-wallet -a

# Visit faucet.aura.network and request tokens
```

### 3. Create Your DID

```bash
aurad tx identity create-did \
  --from my-wallet \
  --chain-id aura-testnet-1 \
  --gas auto \
  --gas-adjustment 1.3 \
  -y
```

### 4. Verify Your DID

```bash
# Get your DID from the transaction
aurad query identity did-by-address $(aurad keys show my-wallet -a)
```

## DID Document

Your DID resolves to a DID Document containing:
- **ID**: Your unique DID (e.g., `did:aura:1abc...`)
- **Verification Methods**: Your public keys
- **Authentication**: Methods for proving control
- **Service Endpoints**: How to interact with your DID

## Next Steps

- [Issue a Credential](./issue-credential) - Become a credential issuer
- [Stake Tokens](./stake-tokens) - Secure the network
