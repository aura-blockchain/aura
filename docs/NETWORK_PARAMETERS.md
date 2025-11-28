# AURA Network Parameters

**Version:** 1.0
**Last Updated:** 2025-11-25
**Chain ID**: aura-mainnet-1 (Mainnet) | aura-testnet-1 (Testnet)

---

## Table of Contents

1. [Overview](#overview)
2. [Consensus Parameters](#consensus-parameters)
3. [Core Cosmos Modules](#core-cosmos-modules)
4. [AURA Custom Modules](#aura-custom-modules)
5. [Genesis Configuration](#genesis-configuration)
6. [Querying Parameters](#querying-parameters)

---

## Overview

This document provides a comprehensive reference for all configurable parameters in the AURA blockchain. Parameters can be modified through governance proposals.

### Parameter Types

- **Consensus Parameters**: Block size, gas limits, validator set
- **Module Parameters**: Module-specific configuration
- **Genesis Parameters**: Initial network state

---

## Consensus Parameters

### Block Parameters

```json
{
  "max_bytes": "22020096",        // 21 MB maximum block size
  "max_gas": "-1",                // Unlimited gas per block
  "time_iota_ms": "1000"          // Minimum block time increment (1s)
}
```

**Query:**
```bash
aurad query consensus params
```

### Evidence Parameters

```json
{
  "max_age_num_blocks": "100000",    // Evidence validity: 100k blocks
  "max_age_duration": "172800000000000",  // ~48 hours in nanoseconds
  "max_bytes": "1048576"             // 1 MB max evidence size
}
```

### Validator Set

```json
{
  "max_validators": 125,             // Maximum active validators
  "pub_key_types": ["ed25519"]       // Supported key algorithms
}
```

**Recommended for Mainnet:**
- `max_validators`: 125-175
- `max_bytes`: 21 MB
- `max_gas`: -1 (unlimited)

**Recommended for Testnet:**
- `max_validators`: 50
- `max_bytes`: 10 MB

---

## Core Cosmos Modules

### Auth Module

```bash
aurad query auth params
```

**Parameters:**
```json
{
  "max_memo_characters": "256",
  "tx_sig_limit": "7",
  "tx_size_cost_per_byte": "10",
  "sig_verify_cost_ed25519": "590",
  "sig_verify_cost_secp256k1": "1000"
}
```

### Bank Module

```bash
aurad query bank params
```

**Parameters:**
```json
{
  "send_enabled": [],                // Tokens allowed to transfer
  "default_send_enabled": true       // Allow transfers by default
}
```

### Staking Module

```bash
aurad query staking params
```

**Parameters:**
```json
{
  "unbonding_time": "1814400s",      // 21 days unbonding period
  "max_validators": 125,             // Maximum active validators
  "max_entries": 7,                  // Max unbonding/redelegation entries
  "historical_entries": 10000,       // Historical info entries
  "bond_denom": "uaura",             // Staking token denomination
  "min_commission_rate": "0.050000000000000000"  // 5% minimum commission
}
```

**Recommended Mainnet:**
- `unbonding_time`: 1814400s (21 days)
- `max_validators`: 125
- `min_commission_rate`: 0.05 (5%)

**Recommended Testnet:**
- `unbonding_time`: 86400s (1 day)
- `max_validators`: 50

### Slashing Module

```bash
aurad query slashing params
```

**Parameters:**
```json
{
  "signed_blocks_window": "10000",            // 10,000 block sliding window
  "min_signed_per_window": "0.500000000000000000",  // Must sign 50% of blocks
  "downtime_jail_duration": "600s",           // 10 minute jail period
  "slash_fraction_double_sign": "0.050000000000000000",  // 5% slash for double-sign
  "slash_fraction_downtime": "0.000100000000000000"      // 0.01% slash for downtime
}
```

**Critical for Validators:**
- Missing > 5,000 blocks in 10,000 block window = jail
- Double-sign penalty: 5% stake + tombstone (permanent)
- Downtime penalty: 0.01% stake + 10 min jail

### Distribution Module

```bash
aurad query distribution params
```

**Parameters:**
```json
{
  "community_tax": "0.020000000000000000",    // 2% to community pool
  "base_proposer_reward": "0.010000000000000000",  // 1% base proposer bonus
  "bonus_proposer_reward": "0.040000000000000000", // 4% bonus (with precommits)
  "withdraw_addr_enabled": true                // Allow changing reward address
}
```

### Governance Module

```bash
aurad query gov params
```

**Parameters:**
```json
{
  "deposit_params": {
    "min_deposit": [
      {
        "denom": "uaura",
        "amount": "10000000000"              // 10,000 AURA minimum deposit
      }
    ],
    "max_deposit_period": "1209600s"         // 14 day deposit period
  },
  "voting_params": {
    "voting_period": "1209600s"              // 14 day voting period
  },
  "tally_params": {
    "quorum": "0.334000000000000000",        // 33.4% quorum
    "threshold": "0.500000000000000000",     // 50% yes threshold
    "veto_threshold": "0.334000000000000000" // 33.4% no-with-veto to reject
  }
}
```

**Governance Process:**
1. Submit proposal with 10,000 AURA deposit
2. 14-day deposit period for others to contribute
3. 14-day voting period
4. Quorum: 33.4% of voting power must vote
5. Pass: 50%+ yes (excluding abstain)
6. Veto: 33.4%+ no-with-veto rejects + burns deposit

---

## AURA Custom Modules

### Bridge Module

```bash
aurad query bridge params
```

**Parameters:**
```go
{
  "bridge_enabled": true,
  "min_confirmations": 1,                    // Block confirmations required
  "bridge_fee_basis_points": 30,             // 0.3% bridge fee
  "max_transfer_amount": "1000000000",       // Max transfer: 1000 AURA
  "validator_threshold_percentage": 67       // 67% validator approval needed
}
```

**Security Parameters:**
```go
{
  "timelock_duration": "24h",                // 24 hour withdrawal timelock
  "fraud_proof_window": "168h"               // 7 day fraud proof window
}
```

### DEX Module

```bash
aurad query dex params
```

**Parameters:**
```go
{
  "trading_fee": "0.003",                    // 0.3% trading fee
  "protocol_fee": "0.0005",                  // 0.05% protocol fee
  "max_slippage_bps": 10000,                 // 100% max slippage (basis points)
  "min_swap_amount": "1000000",              // 1 AURA minimum swap
  "ir_boost_enabled": true,                  // IR completion boost enabled
  "ir_boost_percent": 40,                    // 40% fee discount for IR completers
  "bonding_curve_enabled": true,             // Dynamic pricing enabled
  "buyback_burn_enabled": true,              // Token buyback enabled
  "buyback_percent": 100                     // 100% of fees for buyback
}
```

### Cryptography Module

```bash
aurad query cryptography params
```

**Parameters:**
```go
{
  "default_rotation_interval_days": 90,      // 90 day key rotation
  "enable_auto_rotation": true,              // Auto-rotate keys
  "min_threshold_participants": 2,           // Minimum threshold signers
  "max_threshold_participants": 100,         // Maximum threshold signers
  "min_entropy_bits": 256,                   // 256-bit minimum entropy
  "min_pbkdf2_iterations": 100000,           // PBKDF2 iteration minimum
  "min_argon2_memory_kb": 65536,             // 64 MB Argon2 memory
  "min_argon2_iterations": 3,                // Argon2 iteration minimum
  "enforce_certificate_pinning": true,       // Enforce cert pinning
  "certificate_pin_validity_days": 365,      // 1 year cert validity
  "min_salt_length_bytes": 16,               // 16 byte minimum salt
  "min_key_length_bits": 256                 // 256-bit minimum key length
}
```

### Compliance Module

```bash
aurad query compliance params
```

**Parameters:**
```go
{
  "kyc_required": false,                     // KYC requirement (governance controlled)
  "kyc_expiry_days": 365,                    // KYC validity period
  "aml_monitoring_enabled": true,            // AML monitoring active
  "transaction_limit_uaura": "100000000000", // 100,000 AURA daily limit
  "sanctions_check_enabled": true,           // Sanctions screening
  "gdpr_compliant": true,                    // GDPR compliance mode
  "data_retention_days": 2555                // 7 year data retention
}
```

### Confidence Score Module

```bash
aurad query confidencescore params
```

**Parameters:**
```go
{
  "min_confidence_score": 0,                 // Minimum score
  "max_confidence_score": 100,               // Maximum score
  "initial_score": 50,                       // New user starting score
  "decay_rate": "0.01",                      // Daily decay rate
  "ir_completion_boost": 10,                 // +10 points per IR completion
  "slash_penalty": 20,                       // -20 points for slashing
  "reward_multiplier": "1.5"                 // 1.5x rewards at max score
}
```

### VC Registry Module

```bash
aurad query vcregistry params
```

**Parameters:**
```go
{
  "issuance_fee": "1000000",                 // 1 AURA issuance fee
  "verification_fee": "500000",              // 0.5 AURA verification fee
  "revocation_fee": "250000",                // 0.25 AURA revocation fee
  "max_credential_size": 10240,              // 10 KB max credential size
  "credential_validity_days": 365,           // 1 year default validity
  "allow_self_signed": false,                // Require issuer authorization
  "enable_selective_disclosure": true        // Enable ZKP disclosure
}
```

### Validator Security Module

```bash
aurad query validatorsecurity params
```

**Parameters:**
```go
{
  "monitoring_enabled": true,                // Monitoring active
  "jailing_enabled": true,                   // Auto-jailing enabled
  "slashing_enabled": true,                  // Slashing enabled
  "double_sign_protection": true,            // Double-sign prevention
  "sentry_node_required": false,             // Sentry nodes optional
  "uptime_requirement": "0.95",              // 95% uptime required
  "max_downtime_blocks": 5000                // Max 5000 consecutive misses
}
```

### Network Security Module

```bash
aurad query networksecurity params
```

**Parameters:**
```go
{
  "rate_limit_enabled": true,                // Rate limiting enabled
  "max_connections_per_ip": 100,             // 100 connections per IP
  "gossip_duplicate_threshold": "0.9",       // 90% duplicate gossip threshold
  "mempool_rate_limit": 1000,                // 1000 tx/sec mempool limit
  "peer_reputation_enabled": true,           // Reputation tracking
  "min_peer_reputation": 50,                 // Minimum reputation score
  "sybil_detection_enabled": true            // Sybil attack detection
}
```

### Wallet Security Module

```bash
aurad query walletsecurity params
```

**Parameters:**
```go
{
  "spending_limit_enabled": true,            // Daily spending limits
  "default_daily_limit": "10000000000",      // 10,000 AURA daily limit
  "require_2fa": false,                      // 2FA optional (user choice)
  "session_timeout_minutes": 30,             // 30 minute session timeout
  "dust_filter_enabled": true,               // Dust attack protection
  "min_dust_amount": "1000",                 // 0.001 AURA dust threshold
  "biometric_enabled": true                  // Biometric auth supported
}
```

### Economic Security Module

```bash
aurad query economicsecurity params
```

**Parameters:**
```go
{
  "dynamic_fees_enabled": true,              // Dynamic fee adjustment
  "base_gas_price": "0.025",                 // 0.025 uaura base gas
  "gas_price_multiplier": "1.5",             // 1.5x max gas price
  "mev_protection_enabled": true,            // MEV protection active
  "whale_protection_enabled": true,          // Large tx monitoring
  "whale_threshold": "100000000000",         // 100,000 AURA whale threshold
  "circuit_breaker_enabled": true,           // Emergency pause capability
  "max_price_deviation": "0.1"               // 10% max price deviation
}
```

### Monitoring Module

```bash
aurad query monitoring params
```

**Parameters:**
```go
{
  "telemetry_enabled": true,                 // Telemetry collection
  "prometheus_enabled": true,                // Prometheus metrics
  "alerting_enabled": true,                  // Alert generation
  "anomaly_detection_enabled": true,         // ML anomaly detection
  "alert_threshold": "high",                 // Alert sensitivity
  "metrics_retention_hours": 720             // 30 day metrics retention
}
```

---

## Genesis Configuration

### Mainnet Genesis

Key genesis parameters for aura-mainnet-1:

```json
{
  "genesis_time": "2025-01-01T00:00:00Z",
  "chain_id": "aura-mainnet-1",
  "initial_height": "1",
  "consensus_params": {
    "block": {
      "max_bytes": "22020096",
      "max_gas": "-1"
    }
  },
  "app_state": {
    "staking": {
      "params": {
        "unbonding_time": "1814400s",
        "max_validators": 125,
        "bond_denom": "uaura"
      }
    },
    "gov": {
      "params": {
        "min_deposit": [{"denom": "uaura", "amount": "10000000000"}]
      }
    }
  }
}
```

### Testnet Genesis

Key genesis parameters for aura-testnet-1:

```json
{
  "genesis_time": "2024-11-01T00:00:00Z",
  "chain_id": "aura-testnet-1",
  "initial_height": "1",
  "consensus_params": {
    "block": {
      "max_bytes": "10485760",
      "max_gas": "-1"
    }
  },
  "app_state": {
    "staking": {
      "params": {
        "unbonding_time": "86400s",
        "max_validators": 50
      }
    },
    "gov": {
      "params": {
        "min_deposit": [{"denom": "uaura", "amount": "1000000000"}]
      }
    }
  }
}
```

---

## Querying Parameters

### Query All Module Parameters

```bash
# Core modules
aurad query auth params
aurad query bank params
aurad query staking params
aurad query slashing params
aurad query distribution params
aurad query gov params

# AURA modules
aurad query bridge params
aurad query dex params
aurad query cryptography params
aurad query compliance params
aurad query confidencescore params
aurad query vcregistry params
aurad query validatorsecurity params
aurad query networksecurity params
aurad query walletsecurity params
aurad query economicsecurity params
aurad query monitoring params
```

### Query Specific Parameter

```bash
# Example: Query staking unbonding time
aurad query staking params | jq -r .unbonding_time

# Example: Query bridge fee
aurad query bridge params | jq -r .bridge_fee_basis_points
```

### Modify Parameters via Governance

```bash
# Submit parameter change proposal
aurad tx gov submit-proposal param-change proposal.json \
  --from=proposer \
  --chain-id=aura-mainnet-1 \
  --fees=10000uaura

# Example proposal.json
{
  "title": "Increase Max Validators to 150",
  "description": "Proposal to increase maximum validators from 125 to 150",
  "changes": [
    {
      "subspace": "staking",
      "key": "MaxValidators",
      "value": "150"
    }
  ],
  "deposit": "10000000000uaura"
}
```

---

## Denomination Reference

### Native Denom

- **Base Denom**: `uaura` (micro-aura)
- **Display Denom**: `AURA`
- **Conversion**: 1 AURA = 1,000,000 uaura

### Common Amounts

| AURA | uaura |
|------|-------|
| 0.000001 | 1 |
| 0.001 | 1,000 |
| 1 | 1,000,000 |
| 10 | 10,000,000 |
| 100 | 100,000,000 |
| 1,000 | 1,000,000,000 |
| 10,000 | 10,000,000,000 |
| 100,000 | 100,000,000,000 |

---

## Recommendations

### For Mainnet Launch

**Conservative Settings:**
- `max_validators`: 100-125
- `unbonding_time`: 21 days
- `min_deposit`: 10,000 AURA
- `voting_period`: 14 days
- `bridge_fee`: 0.3%
- `trading_fee`: 0.3%

**Security-First:**
- `kyc_required`: Determine based on legal requirements
- `slashing_enabled`: true
- `double_sign_protection`: true
- `rate_limit_enabled`: true
- `mev_protection_enabled`: true

### For Testnet

**Faster Testing:**
- `unbonding_time`: 1 day
- `voting_period`: 3 days
- `min_deposit`: 1,000 AURA
- Lower fees for easier testing

---

**Document Status**: Production Ready
**Review Cycle**: After each governance parameter change
**Next Review**: With next network upgrade
