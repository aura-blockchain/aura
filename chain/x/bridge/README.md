# Bridge Module

## Overview

The Bridge module enables secure cross-chain token transfers and identity linkage between Aura, PAW, and XAI blockchains. It implements lock/mint and burn/unlock mechanisms with validator multi-signatures, fraud proof windows, Merkle proof verification, and relayer infrastructure for trustless cross-chain operations.

## Features

- **Token Lock/Mint**: Lock tokens on source chain, mint wrapped tokens on destination
- **Token Burn/Unlock**: Burn wrapped tokens, unlock original tokens on source chain
- **Shared Identity**: Link addresses across chains for unified identity
- **Cross-Chain Swaps**: Atomic swaps across multiple chains with routing
- **Fraud Proof System**: Challenge period for detecting fraudulent transfers
- **Validator Multi-Sig**: Require threshold of validator signatures for minting
- **Relayer Network**: Decentralized relayers for transfer completion
- **Merkle Proof Verification**: Trustless verification using block headers
- **Emergency Pause**: Circuit breaker for security incidents

## State

### CrossChainTransfer
- **Transfer ID**: Unique identifier for each transfer
- **Status**: Pending, completed, fraud_detected, expired
- **Source/Target Chain**: Chain identifiers (aura, paw, xai)
- **Sender/Recipient**: Cross-chain addresses
- **Amount**: Transferred token amount and denomination
- **Validator Signatures**: Multi-sig requirements
- **Fraud Proof Window**: Time period for challenges (default 24 hours)
- **Merkle Proof**: Proof of inclusion in source block

### WrappedToken
- **Wrapped Denom**: Token denomination on destination (e.g., "paw.token")
- **Original Denom**: Source token denomination  
- **Total Supply**: Total wrapped tokens minted
- **Source Chain**: Origin blockchain

### SharedIdentity
- **Aura Address**: Primary identity address
- **PAW Address**: Linked PAW address
- **XAI Address**: Linked XAI address
- **Verification Signatures**: Signatures proving ownership

## Messages

### MsgLockTokens
Lock tokens on Aura for transfer to PAW or XAI.

**Example**:
```json
{
  "sender": "aura1...",
  "target_chain": "paw",
  "recipient": "paw1...",
  "amount": {"denom": "uaura", "amount": "1000000"}
}
```

**Response**:
```json
{
  "transfer_id": "xfer_abc123",
  "estimated_completion": 300
}
```

### MsgMintTokens
Mint wrapped tokens on Aura (validator-signed).

**Example**:
```json
{
  "validator": "auravaloper1...",
  "source_chain": "paw",
  "source_tx_hash": "0x123...",
  "recipient": "aura1...",
  "amount": "1000000",
  "denom": "token",
  "validator_signature": "signature_bytes"
}
```

### MsgUnlockTokens
Unlock tokens after burn proof from target chain.

**Example**:
```json
{
  "sender": "aura1...",
  "source_chain": "paw",
  "burn_tx_hash": "0x456...",
  "amount": "1000000",
  "denom": "token",
  "validator_signatures": ["sig1", "sig2", "sig3"],
  "merkle_proof": "proof_bytes",
  "merkle_root": "root_bytes",
  "source_block_height": 12345,
  "source_block_hash": "hash_bytes"
}
```

### MsgBurnTokens
Burn wrapped tokens to unlock on source chain.

**Example**:
```json
{
  "sender": "aura1...",
  "target_chain": "paw",
  "recipient": "paw1...",
  "amount": {"denom": "paw.token", "amount": "1000000"}
}
```

### MsgLinkAddress
Link addresses across chains for shared identity.

**Example**:
```json
{
  "aura_address": "aura1...",
  "paw_address": "paw1...",
  "xai_address": "xai1...",
  "paw_signature": "sig_bytes",
  "xai_signature": "sig_bytes",
  "signer": "aura1..."
}
```

**Response**:
```json
{
  "success": true,
  "linked_identity_id": "identity_xyz"
}
```

### MsgCrossChainSwap
Initiate cross-chain swap with routing.

**Example**:
```json
{
  "sender": "aura1...",
  "source_chain": "aura",
  "input_coin": {"denom": "uaura", "amount": "1000000"},
  "target_chain": "paw",
  "target_denom": "ptoken",
  "min_target_amount": "950000",
  "recipient": "paw1...",
  "max_slippage_bps": 500
}
```

**Response**:
```json
{
  "swap_id": "swap_abc",
  "route": ["aura", "osmosis", "paw"],
  "estimated_completion": 600
}
```

### MsgFinalizeTransfer
Finalize pending transfer after fraud proof window expires.

**Example**:
```json
{
  "sender": "aura1...",
  "transfer_id": "xfer_abc123"
}
```

**Response**:
```json
{
  "success": true,
  "amount": "1000000",
  "recipient": "aura1..."
}
```

### MsgSubmitFraudProof
Challenge fraudulent transfer during fraud proof window.

**Example**:
```json
{
  "challenger": "aura1...",
  "transfer_id": "xfer_abc123",
  "fraud_type": "double_spend",
  "evidence": "evidence_bytes",
  "description": "Transaction was double-spent on source chain"
}
```

**Response**:
```json
{
  "success": true,
  "fraud_proof_id": "fraud_xyz"
}
```

## Queries

### QueryTransfer
```bash
aurad query bridge transfer xfer_abc123
```

### QueryUserTransfers
```bash
aurad query bridge user-transfers aura1... --chain paw
```

### QueryWrappedToken
```bash
aurad query bridge wrapped paw.token
```

### QuerySharedIdentity
```bash
aurad query bridge identity aura1...
```

### QueryBridgeStats
```bash
aurad query bridge stats
```

## Events

| Event Type | Attributes | Description |
|------------|------------|-------------|
| `tokens_locked` | `transfer_id`, `sender`, `target_chain`, `amount` | Tokens locked for transfer |
| `tokens_minted` | `recipient`, `source_chain`, `wrapped_denom`, `amount` | Wrapped tokens minted |
| `tokens_unlocked` | `recipient`, `source_chain`, `amount` | Original tokens unlocked |
| `tokens_burned` | `sender`, `target_chain`, `amount` | Wrapped tokens burned |
| `address_linked` | `aura_address`, `paw_address`, `xai_address` | Addresses linked across chains |
| `cross_chain_swap_initiated` | `swap_id`, `route`, `input`, `target` | Cross-chain swap started |
| `transfer_finalized` | `transfer_id`, `recipient`, `amount` | Transfer completed |
| `fraud_proof_submitted` | `fraud_proof_id`, `transfer_id`, `challenger` | Fraud challenge submitted |
| `bridge_paused` | `chains`, `reason` | Bridge operations paused |

## Errors

| Code | Name | Description |
|------|------|-------------|
| 1 | `ErrTransferNotFound` | Transfer ID not found |
| 2 | `ErrInvalidChain` | Chain identifier invalid |
| 3 | `ErrInsufficientSignatures` | Not enough validator signatures |
| 4 | `ErrInvalidMerkleProof` | Merkle proof verification failed |
| 5 | `ErrFraudProofWindowActive` | Cannot finalize during fraud window |
| 6 | `ErrFraudProofWindowExpired` | Cannot challenge after window |
| 7 | `ErrBridgePaused` | Bridge operations paused |
| 8 | `ErrWrappedTokenNotFound` | Wrapped token not registered |

## Integration Notes

### For Wallet Developers

1. **Transfer Tracking**: Display transfer status with estimated completion
2. **Fraud Window**: Show countdown timer for finalization eligibility
3. **Multi-Chain Display**: Show linked addresses across all chains
4. **Relayer Selection**: Allow users to specify preferred relayer
5. **Route Visualization**: Display swap route across chains

### Security Considerations

- **Fraud Proof Window**: Wait full 24 hours before finalizing high-value transfers
- **Validator Threshold**: Verify sufficient validator signatures
- **Merkle Proofs**: Validate proof structure before submission
- **Emergency Pause**: Monitor for pause events, halt operations if detected

### Best Practices

- **Timeouts**: Set appropriate timeouts for cross-chain operations
- **Gas Reserves**: Maintain gas on destination chain for completion
- **Identity Linking**: Link addresses early to streamline cross-chain operations
- **Route Optimization**: Use QueryBestRoute for optimal cross-chain paths
