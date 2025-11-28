# Confidence Score Module - CLI Commands Summary

## Overview
Comprehensive CLI commands have been created for the `x/confidencescore` module, providing full command-line access to all query and transaction functionality.

## Files Created

### 1. Query Commands
**File:** `C:\Users\decri\GitClones\aura\chain\x\confidencescore\client\cli\query.go`

#### Available Query Commands:

1. **`score [wallet-address]`** - Query a user's confidence score and verification status
   - Returns: Total score, verification status, anchor info, arena scores, IR count, last update

2. **`completions [wallet-address]`** - Query a user's IR completion history
   - Flags: `--arena` (filter by arena type), pagination flags
   - Returns: List of all IR completions with scores, bonuses, timestamps, verifier info

3. **`history [wallet-address]`** - Query score change history
   - Flags: `--from-height`, `--to-height`, pagination flags
   - Returns: Complete history of score changes with reasons and deltas

4. **`thresholds`** - Query verification thresholds
   - Returns: Verified Human threshold, VC thresholds, arena focus thresholds

5. **`verified-users`** - Query list of verified users (CS >= 10,000)
   - Flags: `--min-score` (minimum score filter), pagination flags
   - Returns: List of verified user addresses and scores

6. **`arena-breakdown [wallet-address]`** - Query score breakdown by arena
   - Returns: Per-arena scores, IR counts, focus bonus status, list of focus arenas

7. **`slash-records [wallet-address]`** - Query slash records for a user
   - Returns: Slash history with amounts, reasons, evidence, appeal status

8. **`params`** - Query module parameters
   - Returns: All module parameters including thresholds, bonuses, rate limits, jackpot config

9. **`ir-completion [wallet-address] [ir-id]`** - Query specific IR completion
   - Returns: Detailed completion record with all multipliers and metadata

### 2. Transaction Commands
**File:** `C:\Users\decri\GitClones\aura\chain\x\confidencescore\client\cli\tx.go`

#### Available Transaction Commands:

1. **`record-completion [wallet-address] [ir-id] [proof-hash] [verifier-hash]`**
   - Record an IR completion (AI assistant only)
   - Arguments: wallet address, IR ID, proof hash (hex), verifier hash (hex)
   - Automatically calculates: velocity bonus, arena multiplier, jackpot bonus
   - Returns: Score earned, new total, verification achievement status, multipliers applied

2. **`recalculate-score [wallet-address]`**
   - Recalculate user's total score from history (governance only)
   - Use cases: Fix discrepancies, audit scores, recover from corruption
   - Returns: Previous score, recalculated score, list of discrepancies

3. **`slash [wallet-address] [ir-id] [slash-amount] [reason]`**
   - Slash confidence score for fraud (governance only)
   - Flags: `--evidence` (IPFS hash or URL)
   - Reasons: fraud_detected, false_attestation, collusion, duplicate_completion
   - Returns: Previous score, new score, verification revocation status, slash tx hash

4. **`appeal [slash-tx-hash] [deposit]`**
   - Appeal a confidence score slash
   - Flags: `--evidence` (IPFS hash or URL)
   - Requirements: Appeal before deadline, provide deposit (e.g., "1000aeq")
   - Returns: Appeal acceptance status, review deadline

5. **`resolve-appeal [wallet-address] [slash-tx-hash] [restore]`**
   - Resolve a slash appeal (governance only)
   - Arguments: wallet address, slash tx hash, restore (true/false)
   - Flags: `--notes` (resolution notes)
   - Returns: Restored amount, deposit return status

## Command Usage Examples

### Query Examples

```bash
# Query a user's confidence score
$ aurad query confidencescore score aura1abc123...

# Query completions filtered by arena with pagination
$ aurad query confidencescore completions aura1abc123... --arena Biometric --page 1 --limit 50

# Query score history for specific block range
$ aurad query confidencescore history aura1abc123... --from-height 1000 --to-height 5000

# Query arena breakdown
$ aurad query confidencescore arena-breakdown aura1abc123...

# Query verified users with minimum score
$ aurad query confidencescore verified-users --min-score 15000

# Query specific IR completion
$ aurad query confidencescore ir-completion aura1abc123... IR-102

# Query slash records
$ aurad query confidencescore slash-records aura1abc123...

# Query thresholds and params
$ aurad query confidencescore thresholds
$ aurad query confidencescore params
```

### Transaction Examples

```bash
# Record IR completion (AI assistant)
$ aurad tx confidencescore record-completion \
    aura1user123... IR-102 \
    a1b2c3d4e5f6... 9f8e7d6c5b4a... \
    --from assistant-key \
    --chain-id aura-testnet-1

# Recalculate score (governance)
$ aurad tx confidencescore recalculate-score aura1user123... \
    --from governance-key \
    --chain-id aura-mainnet-1

# Slash score for fraud (governance)
$ aurad tx confidencescore slash \
    aura1user123... IR-305 5000 fraud_detected \
    --evidence QmX3d9f2h... \
    --from governance-key \
    --chain-id aura-mainnet-1

# Appeal a slash (user)
$ aurad tx confidencescore appeal \
    A1B2C3D4E5F6... 1000aeq \
    --evidence QmY4e8g3k... \
    --from user-key \
    --chain-id aura-mainnet-1

# Resolve appeal - restore score (governance)
$ aurad tx confidencescore resolve-appeal \
    aura1user123... A1B2C3D4E5F6... true \
    --notes "Evidence validates user claim" \
    --from governance-key \
    --chain-id aura-mainnet-1

# Resolve appeal - deny (governance)
$ aurad tx confidencescore resolve-appeal \
    aura1user123... A1B2C3D4E5F6... false \
    --notes "Insufficient evidence provided" \
    --from governance-key \
    --chain-id aura-mainnet-1
```

## Integration

The CLI commands have been integrated into the main Aura daemon:

**Modified Files:**
- `C:\Users\decri\GitClones\aura\chain\cmd\aurad\cmd\query.go` - Integrated query commands
- `C:\Users\decri\GitClones\aura\chain\cmd\aurad\cmd\tx.go` - Integrated transaction commands

The commands are accessible via:
```bash
$ aurad query confidencescore [subcommand]
$ aurad tx confidencescore [subcommand]
```

## Arena Types

The following arena types are supported for filtering and breakdown:
- Biometric
- Possession
- Knowledge
- Social
- GeoLocation
- HighAssurance
- Persistence
- Specialized

## Score Calculation Details

### Velocity Bonus (VBT)
- 1.25x multiplier if IR completed within 7 days of previous IR
- 1.10x multiplier if IR completed within 30 days of previous IR
- 1.0x (no bonus) otherwise

### Arena Multiplier
- 1.5x multiplier if arena score >= 5,000 (focus bonus active)
- 1.2x multiplier if arena score >= 4,000
- 1.1x multiplier if arena score >= 3,000
- 1.0x (no bonus) otherwise

### Jackpot Bonus
- Rare 25.0x multiplier (1 in 1000 odds)
- Rare 5.0x multiplier (1 in 100 odds)
- 1.0x (no bonus) most of the time

### Verification Thresholds
- Verified Human: 10,000 CS
- High Assurance: 15,000 CS
- Arena Focus: 5,000 CS per arena

## Slashing System

### Slash Reasons
- `fraud_detected` - Fraud detected in verification
- `false_attestation` - Assistant provided false attestation
- `collusion` - Collusion detected
- `duplicate_completion` - Duplicate/replay completion

### Appeal Process
1. User has 30 days (configurable) to appeal after slash
2. User must provide deposit (default 1000 AEQ)
3. Governance reviews evidence
4. If successful: score restored, deposit returned
5. If unsuccessful: slash stands, deposit forfeit

## Rate Limiting
- Max IRs per day: 10 (configurable)
- Max IRs per hour: 3 (configurable)

## Compilation Status

✅ **SUCCESS** - All CLI commands compile successfully without errors.

The confidence score CLI commands are fully implemented and ready for use. The compilation was verified with:
```bash
cd C:\Users\decri\GitClones\aura\chain\x\confidencescore\client\cli
go build -v .
```

## Next Steps

1. Test commands with a running node
2. Add bash completion scripts for CLI
3. Create user documentation with more examples
4. Add CLI integration tests
5. Consider adding JSON output format options for scripting

## Notes

- All commands use gRPC to communicate with the blockchain
- Commands support standard Cosmos SDK flags (--chain-id, --from, --gas, etc.)
- Query commands support pagination for large result sets
- All commands include comprehensive help text accessible via --help flag
