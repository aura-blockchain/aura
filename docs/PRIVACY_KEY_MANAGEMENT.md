# Privacy Module Key Management Security

## Critical Security Principle

**PRIVATE KEYS MUST NEVER BE STORED ON-CHAIN OR TRANSMITTED TO THE BLOCKCHAIN**

This document explains the secure key management model for the Aura privacy module.

---

## Key Types

### Public Keys (Stored On-Chain)

These keys are stored in the blockchain state and are publicly visible:

- **Public View Key**: Allows anyone to see that encrypted transactions are directed to you, but not decrypt them
- **Public Spend Key**: Used in stealth address generation and transaction validation

Public keys are safe to store on-chain because they cannot be used to:
- Decrypt private transaction data
- Spend funds
- Reveal private information

### Private Keys (Client-Side Only)

These keys MUST NEVER leave the client device:

- **Private View Key**: Used by the owner to decrypt and view their transaction data
- **Private Spend Key**: Used by the owner to spend funds and create transactions

Private keys must be:
- Generated client-side using cryptographically secure random number generators
- Stored securely in the client's local keystore (hardware wallet, encrypted file, etc.)
- Never transmitted to the blockchain
- Never included in transaction messages
- Never stored in blockchain state

---

## Security Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    CLIENT DEVICE                            │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  PRIVATE KEY STORAGE (Hardware Wallet / Keystore)    │  │
│  │                                                       │  │
│  │  • Private View Key    (32 bytes)                    │  │
│  │  • Private Spend Key   (32 bytes)                    │  │
│  │                                                       │  │
│  │  NEVER TRANSMITTED TO BLOCKCHAIN                     │  │
│  └──────────────────────────────────────────────────────┘  │
│                          ↓                                  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  CLIENT-SIDE OPERATIONS                              │  │
│  │                                                       │  │
│  │  • Key Generation                                    │  │
│  │  • Transaction Decryption (using Private View Key)  │  │
│  │  • Transaction Signing (using Private Spend Key)    │  │
│  │  • Derive Public Keys from Private Keys             │  │
│  └──────────────────────────────────────────────────────┘  │
│                          ↓                                  │
│                   PUBLIC KEYS ONLY                          │
└─────────────────────────┼───────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│                    BLOCKCHAIN                               │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  ON-CHAIN STATE (Publicly Readable)                  │  │
│  │                                                       │  │
│  │  • Public View Key     (32 bytes)                    │  │
│  │  • Public Spend Key    (32 bytes)                    │  │
│  │  • Encrypted Transaction Data                        │  │
│  │  • Commitments                                       │  │
│  │  • Zero-Knowledge Proofs                             │  │
│  │                                                       │  │
│  │  NO PRIVATE KEYS EVER                                │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## Client-Side Operations

### 1. Key Generation

```go
// Client-side pseudocode (NEVER on blockchain)
func GenerateViewKeys() (publicKey, privateKey []byte) {
    privateKey = generateSecureRandom(32)  // Cryptographically secure
    publicKey = derivePublicKey(privateKey) // Elliptic curve operation

    // Store privateKey in secure local keystore
    // Transmit ONLY publicKey to blockchain
    return publicKey, privateKey
}
```

### 2. View Key Registration (On-Chain)

```go
// Transaction message sent to blockchain
msg := MsgRegisterViewKey{
    Owner: userAddress,
    ViewKey: ViewKey{
        PublicViewKey: publicViewKey,  // Safe to transmit
        KeyType:       "INCOMING",
        // NO private_view_key field - removed for security
    },
}
```

### 3. Transaction Decryption (Client-Side)

```go
// Client-side pseudocode (NEVER on blockchain)
func DecryptTransaction(encryptedData []byte, privateViewKey []byte) ([]byte, error) {
    // 1. Download encrypted transaction from blockchain
    encryptedTx := blockchain.QueryTransaction(txHash)

    // 2. Decrypt locally using private view key
    decryptedData := decryptWithPrivateKey(encryptedTx, privateViewKey)

    // 3. Parse and display transaction details
    return decryptedData
}
```

---

## What Changed (Security Fix)

### BEFORE (VULNERABLE):

```protobuf
// WRONG - CRITICAL VULNERABILITY
message ViewKey {
    bytes public_view_key = 1;
    bytes private_view_key = 2;  // ❌ STORED ON-CHAIN - ANYONE CAN READ
    // ...
}
```

**Attack**: Anyone could query the blockchain and retrieve all private view keys, completely defeating privacy.

### AFTER (SECURE):

```protobuf
// CORRECT - SECURE
message ViewKey {
    bytes public_view_key = 1;  // ✅ Public only, safe to store on-chain
    // private_view_key field removed entirely
    // ...
}
```

**Protection**: No private keys are ever stored on-chain or transmitted in messages.

---

## Validation & Security Checks

The privacy module now enforces multiple layers of security:

### 1. Proto-Level Protection
- `private_view_key` field removed from `ViewKey` message
- No way to accidentally include private keys in transactions

### 2. Message Handler Validation
```go
// In RegisterViewKey handler
if len(msg.ViewKey.PublicViewKey) == 0 {
    return error("public view key cannot be empty")
}

// Validate key length (32, 33, or 64 bytes for common curves)
if keyLen != 32 && keyLen != 33 && keyLen != 64 {
    return error("invalid public key length")
}

// Defensive check against abuse
if msg.ViewKey.KeyType == "PRIVATE" || msg.ViewKey.KeyType == "SECRET" {
    return error("private keys cannot be registered on-chain")
}
```

### 3. Query Endpoint Removal
- `DecryptWithViewKey` query removed entirely
- No on-chain decryption operations
- All decryption must be client-side

---

## Migration Notes

### For Existing Deployments

If you have an existing chain with private keys already stored:

1. **IMMEDIATELY rotate all keys**:
   - Generate new view key pairs client-side
   - Register new public view keys on-chain
   - Assume old private keys are compromised

2. **Run state migration** to clear old private keys:
```go
// Migration handler
func MigrateViewKeys(ctx sdk.Context, keeper Keeper) error {
    // For all stored view keys, clear private_view_key field
    // This requires a chain upgrade/migration
}
```

3. **Notify all users** that their old view keys are compromised and they must generate new ones

### For New Deployments

- No migration needed
- Private keys never stored
- Follow this documentation for proper key management

---

## Security Audit Checklist

- [x] Private view keys removed from proto definitions
- [x] Private spend keys never included in any messages
- [x] Query endpoints do not accept private keys as parameters
- [x] Message handlers validate only public keys are provided
- [x] Key length validation implemented
- [x] Defensive checks against "PRIVATE"/"SECRET" key types
- [x] Client-side decryption documented
- [x] On-chain decryption methods removed
- [x] Security documentation created

---

## References

### Privacy Protocols Using This Model

- **Monero**: Uses view keys and spend keys with same separation
- **Zcash**: Viewing keys are public, spending keys are private
- **Mimblewimble**: Blinding factors never leave client

### Standards

- [BIP32](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki) - Hierarchical Deterministic Wallets
- [BIP44](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki) - Multi-Account Hierarchy
- [SLIP-0010](https://github.com/satoshilabs/slips/blob/master/slip-0010.md) - Key Derivation for Ed25519

---

## Support

For questions about secure key management:

1. Read this document thoroughly
2. Review the privacy module source code
3. Consult with a blockchain security auditor
4. Never compromise on the principle: **PRIVATE KEYS NEVER ON-CHAIN**

---

**Last Updated**: 2025-12-02
**Security Level**: CRITICAL
**Review Status**: Post-vulnerability-fix
