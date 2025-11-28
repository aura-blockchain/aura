# Confidence Score Aggregation Module

## Overview

The `confidencescore` module manages the aggregation and verification of Confidence Scores (CS) across the Aura network. It tracks user completion of Inclusion Routines (IRs), calculates cumulative scores with bonuses, and determines verification status for Verifiable Credential (VC) issuance.

**Key Features:**
- IR completion recording with assistant attestation
- Score calculation with velocity, arena, and jackpot multipliers
- Verification status management (CS ≥ 10,000 = Verified)
- Arena focus bonuses (1.1× to 1.5× multipliers)
- Slashing and appeal mechanisms for fraud prevention
- Comprehensive query API for wallets and verifiers

## Architecture

```
confidencescore/
├── keeper/              # Core state management and business logic
│   ├── keeper.go        # Main keeper with state access
│   ├── ir_completion.go # IR completion recording and validation
│   ├── score_calculation.go # Score calculation with multipliers
│   ├── slash.go         # Slashing and appeal logic
│   ├── queries.go       # Query helper methods
│   └── keeper_test.go   # Comprehensive keeper tests
├── params/              # Module parameters
│   ├── store.go         # Params store
│   └── store_test.go    # Params tests
├── types/               # Type definitions and converters
│   ├── params.go        # Module parameters and defaults
│   ├── keys.go          # Store keys
│   ├── errors.go        # Module-specific errors
│   ├── converters.go    # Proto <-> internal type converters
│   ├── genesis.go       # Genesis state handling
│   └── events.go        # Event types and attributes
├── module.go            # AppModule interface implementation
├── msg_server.go        # Message handler implementations
├── query_server.go      # Query handler implementations
└── README.md            # This file
```

## Core Concepts

### Confidence Score (CS)

An aggregate numerical metric representing the statistical reliability of a user's verified identity claims. Each completed IR adds points to the total CS.

**Formula:**
```
TotalCS = Σ(IRᵢ.base_score × velocity_bonus × arena_bonus × jackpot_bonus)
```

### Verification Thresholds

| Status | CS Threshold | Description |
|--------|--------------|-------------|
| Unverified | < 10,000 | Default state for new users |
| Verified | ≥ 10,000 | Eligible for VC:isVerifiedHuman |
| High Assurance | ≥ 15,000 | Eligible for VC:hasHighAssurance |
| Arena Focus | ≥ 5,000 in single arena | Specialized expertise badge |

### Multipliers

**Velocity Bonus (Time-Based):**
- Complete verification in ≤7 days: **1.25× multiplier**
- Complete verification in ≤30 days: **1.10× multiplier**
- After 30 days: **1.0× (no bonus)**

**Arena Focus Bonus:**
- ≥3,000 CS in arena: **1.1× multiplier**
- ≥4,000 CS in arena: **1.2× multiplier**
- ≥5,000 CS in arena: **1.5× multiplier**

**Jackpot Bonus (Probabilistic):**
- 1 in 100 chance: **5.0× multiplier**
- 1 in 1,000 chance: **25.0× multiplier**

## State Schema

### UserConfidenceRecord

```protobuf
message UserConfidenceRecord {
  string wallet_address = 1;              // Bech32 Aura address
  uint64 total_score = 2;                 // Aggregate CS value
  repeated IRCompletion completed_irs = 3; // Full completion history
  bool has_anchor = 4;                    // IR-000 completed
  uint64 last_updated_height = 5;         // Block height of last update
  map<string, ArenaScore> arena_scores = 6; // Per-arena totals
  VerificationStatus status = 7;          // Verification status
  google.protobuf.Timestamp last_updated = 8;
  uint64 verification_achieved_height = 9;
  AnchorInfo anchor_info = 11;            // IR-000 metadata
}
```

### IRCompletion

```protobuf
message IRCompletion {
  string ir_id = 1;                      // e.g., "IR-102"
  uint64 base_score = 2;                 // Base score from IR definition
  uint64 final_score = 3;                // Final score after multipliers
  google.protobuf.Timestamp completed_at = 4;
  string assistant_address = 6;          // AI assistant who verified
  bytes proof_hash = 7;                  // SHA256 of proof data
  bytes verifier_hash = 8;               // SHA256 of verifier plug-in
  float velocity_bonus = 10;             // Velocity multiplier applied
  float arena_bonus = 11;                // Arena multiplier applied
  float jackpot_bonus = 12;              // Jackpot multiplier applied
  string arena = 14;                     // Arena type
}
```

## Message Types

### RecordIRCompletion

Records an IR completion with assistant attestation.

```protobuf
message MsgRecordIRCompletion {
  string wallet_address = 1;
  string ir_id = 2;
  bytes proof_hash = 3;                  // SHA256 of proof
  bytes verifier_hash = 4;               // SHA256 of verifier plug-in
  string assistant_address = 5;          // Signer (AI assistant)
  google.protobuf.Timestamp timestamp = 6;
}
```

**Validation Rules:**
- Signer must be registered AI Assistant
- IR must exist in registry and be Active
- User must have completed IR-000 (except for IR-000 itself)
- IR must not already be completed by this user
- Prerequisites must be satisfied
- Rate limits must not be exceeded
- Proof hash must be 32 bytes (SHA256)

### RecalculateScore

Recalculates a user's total score (admin/governance).

```protobuf
message MsgRecalculateScore {
  string wallet_address = 1;
  string authority = 2;                  // Governance module address
}
```

### SlashScore

Slashes a user's score for fraud (governance).

```protobuf
message MsgSlashScore {
  string wallet_address = 1;
  string ir_id = 2;                      // IR being disputed
  string reason = 3;
  uint64 slash_amount = 4;               // CS points to deduct
  string authority = 5;                  // Governance or fraud detector
  string evidence = 6;                   // IPFS hash or metadata
}
```

### AppealSlash

Allows a user to appeal a slash.

```protobuf
message MsgAppealSlash {
  string wallet_address = 1;
  string slash_tx_hash = 2;
  string evidence = 3;                   // IPFS hash or metadata
  string deposit = 4;                    // AURA deposit for appeal
}
```

### ResolveAppeal

Resolves an appeal (governance).

```protobuf
message MsgResolveAppeal {
  string wallet_address = 1;
  string slash_tx_hash = 2;
  bool restore_score = 3;                // Whether to restore slashed score
  string authority = 4;                  // Governance module address
  string resolution_notes = 5;
}
```

## Query Types

### UserScore

Get a user's confidence score and status.

```bash
# Example CLI query
aura query confidencescore user-score <wallet-address>
```

### UserCompletions

Get a user's completed IRs with pagination.

```bash
# Example CLI query
aura query confidencescore user-completions <wallet-address> --arena Biometric
```

### ScoreHistory

Get a user's score change history.

```bash
# Example CLI query
aura query confidencescore score-history <wallet-address> --from-height 100 --to-height 200
```

### Thresholds

Get verification thresholds.

```bash
# Example CLI query
aura query confidencescore thresholds
```

### VerifiedUsers

List verified users with pagination.

```bash
# Example CLI query
aura query confidencescore verified-users --min-score 10000 --limit 100
```

### ArenaBreakdown

Get a user's score breakdown by arena.

```bash
# Example CLI query
aura query confidencescore arena-breakdown <wallet-address>
```

### SlashRecord

Get slash records for a user.

```bash
# Example CLI query
aura query confidencescore slash-record <wallet-address>
```

### Params

Get module parameters.

```bash
# Example CLI query
aura query confidencescore params
```

## Module Parameters

```go
type Params struct {
    VerificationThreshold   uint64    // Default: 10,000
    HighAssuranceThreshold  uint64    // Default: 15,000
    ArenaFocusThreshold     uint64    // Default: 5,000
    VelocityBonusDays       []uint64  // [7, 30]
    VelocityBonusMultipliers []float32 // [1.25, 1.10]
    ArenaMultipliers        map[uint64]float32 // {3000: 1.1, 4000: 1.2, 5000: 1.5}
    SlashPercentage         uint64    // Default: 25%
    AppealDeposit           string    // Default: 1000 AURA
    MaxIRsPerDay            uint64    // Default: 10
    MaxIRsPerHour           uint64    // Default: 3
    JackpotOdds             []uint64  // [100, 1000]
    JackpotMultipliers      []float32 // [5.0, 25.0]
    StalenessEnabled        bool      // Default: false
    DegradationRatePerYear  uint64    // Default: 0
}
```

## Events

### EventIRCompleted

```protobuf
message EventIRCompleted {
  string wallet_address = 1;
  string ir_id = 2;
  uint64 score_earned = 3;
  uint64 new_total_score = 4;
  string assistant_address = 5;
  string arena = 6;
  uint64 block_height = 7;
  float velocity_multiplier = 8;
  float arena_multiplier = 9;
  float jackpot_multiplier = 10;
}
```

### EventVerificationAchieved

```protobuf
message EventVerificationAchieved {
  string wallet_address = 1;
  uint64 final_score = 2;
  google.protobuf.Timestamp achieved_at = 3;
  uint32 ir_count = 4;
  uint64 days_since_anchor = 5;
}
```

### EventArenaFocusAchieved

```protobuf
message EventArenaFocusAchieved {
  string wallet_address = 1;
  string arena = 2;
  uint64 arena_score = 3;
  google.protobuf.Timestamp achieved_at = 4;
}
```

### EventScoreSlashed

```protobuf
message EventScoreSlashed {
  string wallet_address = 1;
  string ir_id = 2;
  uint64 slash_amount = 3;
  uint64 new_score = 4;
  string reason = 5;
  bool verification_revoked = 6;
  string slash_tx_hash = 7;
}
```

## Integration Points

### With Inclusion Routines Module

The confidencescore module depends on the `inclusionroutines` module for:
- IR definitions (score values, prerequisites, arena)
- IR status (Active, Suspended, Retired)
- Rate limit configurations

### With VC Registry Module

The confidencescore module provides to the `vc_registry` module:
- User verification status (CS ≥ 10,000)
- Arena focus achievements
- Completed IR list for policy checks

### With AI Assistant Network

The module receives attestations from AI Assistants:
- Proof of IR completion
- Verifier plug-in hash
- Assistant signature

## Security Considerations

### Sybil Resistance

- **IR-000 Anchor:** Government ID verification prevents easy identity creation
- **High CS Threshold:** 10,000 points requires significant effort
- **Prerequisite Chains:** Forces sequential completion
- **Rate Limiting:** Prevents rapid farming
- **AI Assistant Staking:** Assistants risk stake for false attestations

### Replay Attack Prevention

```go
func CheckReplay(walletAddr string, proofHash []byte) error {
    // Check if this exact proof hash already used
    if k.proofHashes[walletAddr][proofHash] {
        return ErrReplayDetected
    }
    // Store hash to prevent reuse
    k.proofHashes[walletAddr][proofHash] = true
    return nil
}
```

### Front-Running Protection

- IR completions are atomic transactions
- No mempool-visible state changes
- Assistant signatures time-bound (5 min expiry)
- Block-based jackpot randomness (not predictable pre-commit)

## Testing

Run module tests:

```bash
cd chain/x/confidencescore
go test ./...
```

Run keeper tests:

```bash
go test ./keeper
```

Run integration tests:

```bash
go test -tags=integration ./...
```

## Example Usage

### Recording an IR Completion

```bash
aura tx confidencescore record-ir-completion \
  --wallet-address aura1abc... \
  --ir-id IR-102 \
  --proof-hash 0x123... \
  --verifier-hash 0x456... \
  --assistant-address aura1assistant... \
  --from assistant-key
```

### Querying User Score

```bash
aura query confidencescore user-score aura1abc...
```

### Slashing a Score (Governance)

```bash
aura tx confidencescore slash-score \
  --wallet-address aura1abc... \
  --ir-id IR-102 \
  --slash-amount 2500 \
  --reason "fraud_detected" \
  --evidence ipfs://QmXYZ... \
  --from governance-key
```

## Contributing

See the main Aura repository CONTRIBUTING.md for guidelines.

## License

Copyright © 2025 Aequitas Foundation
