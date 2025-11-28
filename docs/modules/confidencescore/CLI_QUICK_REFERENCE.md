# Confidence Score CLI - Quick Reference

## Query Commands

| Command | Description | Example |
|---------|-------------|---------|
| `score [address]` | Get user's confidence score | `aurad query confidencescore score aura1abc...` |
| `completions [address]` | List IR completions | `aurad query confidencescore completions aura1abc... --arena Biometric` |
| `history [address]` | Get score change history | `aurad query confidencescore history aura1abc...` |
| `thresholds` | Get verification thresholds | `aurad query confidencescore thresholds` |
| `verified-users` | List verified users | `aurad query confidencescore verified-users --min-score 15000` |
| `arena-breakdown [address]` | Get arena score breakdown | `aurad query confidencescore arena-breakdown aura1abc...` |
| `slash-records [address]` | Get slash history | `aurad query confidencescore slash-records aura1abc...` |
| `params` | Get module parameters | `aurad query confidencescore params` |
| `ir-completion [address] [ir-id]` | Get specific IR completion | `aurad query confidencescore ir-completion aura1abc... IR-102` |

## Transaction Commands

| Command | Who Can Use | Description |
|---------|-------------|-------------|
| `record-completion [address] [ir-id] [proof] [verifier]` | AI Assistants | Record IR completion |
| `recalculate-score [address]` | Governance | Recalculate user's score |
| `slash [address] [ir-id] [amount] [reason]` | Governance | Slash score for fraud |
| `appeal [slash-tx] [deposit]` | Users | Appeal a slash |
| `resolve-appeal [address] [slash-tx] [restore]` | Governance | Resolve an appeal |

## Common Flags

### Query Flags
- `--page [n]` - Page number for pagination
- `--limit [n]` - Results per page
- `--height [n]` - Query at specific block height
- `--output json` - JSON output format

### Transaction Flags
- `--from [key]` - Key to sign transaction
- `--chain-id [id]` - Chain ID
- `--gas auto` - Auto-estimate gas
- `--gas-prices [amount]` - Gas prices
- `--yes` - Skip confirmation prompt

## Arena Types

- `Biometric` - Biometric verification (fingerprint, face, etc.)
- `Possession` - Device/credential possession
- `Knowledge` - Knowledge-based verification
- `Social` - Social graph verification
- `GeoLocation` - Location-based verification
- `HighAssurance` - High-assurance methods
- `Persistence` - Time-persistence verification
- `Specialized` - Specialized verification methods

## Slash Reasons

- `fraud_detected` - Fraud detected in verification
- `false_attestation` - False attestation by assistant
- `collusion` - Collusion detected
- `duplicate_completion` - Duplicate/replay completion

## Score Multipliers

### Velocity Bonus (VBT)
- Within 7 days: **1.25x**
- Within 30 days: **1.10x**
- Otherwise: 1.0x

### Arena Multiplier
- Arena score >= 5000: **1.5x** (focus bonus)
- Arena score >= 4000: **1.2x**
- Arena score >= 3000: **1.1x**
- Otherwise: 1.0x

### Jackpot Bonus
- 1 in 1000: **25.0x** 🎰
- 1 in 100: **5.0x** 🎰
- Otherwise: 1.0x

## Thresholds

- **Verified Human**: 10,000 CS
- **High Assurance**: 15,000 CS
- **Arena Focus**: 5,000 CS per arena

## Examples by Use Case

### User Checking Their Score
```bash
# Get basic score info
aurad query confidencescore score aura1myaddress...

# Get detailed arena breakdown
aurad query confidencescore arena-breakdown aura1myaddress...

# Check specific IR completion
aurad query confidencescore ir-completion aura1myaddress... IR-102
```

### AI Assistant Recording Completion
```bash
# Record IR completion with proof
aurad tx confidencescore record-completion \
  aura1user... IR-102 \
  a1b2c3d4... 9f8e7d6c... \
  --from assistant-key \
  --chain-id aura-testnet-1 \
  --gas auto
```

### User Appealing a Slash
```bash
# Check slash records first
aurad query confidencescore slash-records aura1myaddress...

# File appeal with evidence
aurad tx confidencescore appeal \
  ABC123TXHASH 1000aeq \
  --evidence QmIPFSHash... \
  --from my-key \
  --chain-id aura-mainnet-1
```

### Governance Operations
```bash
# Recalculate a user's score
aurad tx confidencescore recalculate-score aura1user... \
  --from governance-key

# Slash for fraud
aurad tx confidencescore slash \
  aura1user... IR-305 5000 fraud_detected \
  --evidence QmEvidence... \
  --from governance-key

# Resolve appeal - restore score
aurad tx confidencescore resolve-appeal \
  aura1user... ABC123TXHASH true \
  --notes "Valid appeal" \
  --from governance-key
```

### Querying Lists
```bash
# Get verified users
aurad query confidencescore verified-users --limit 100

# Get completion history with pagination
aurad query confidencescore completions aura1user... \
  --page 1 --limit 50

# Get score history for specific block range
aurad query confidencescore history aura1user... \
  --from-height 1000 --to-height 5000
```

## Tips

1. **Always check help**: Add `--help` to any command for detailed information
   ```bash
   aurad query confidencescore score --help
   ```

2. **Use JSON output for scripting**:
   ```bash
   aurad query confidencescore score aura1... --output json | jq '.total_score'
   ```

3. **Test transactions first**: Use `--dry-run` flag to test without broadcasting
   ```bash
   aurad tx confidencescore record-completion ... --dry-run
   ```

4. **Check gas requirements**: Use `--gas auto` for automatic gas estimation
   ```bash
   aurad tx confidencescore slash ... --gas auto --gas-adjustment 1.3
   ```

5. **Filter arena completions**: Use `--arena` flag to see specific arena progress
   ```bash
   aurad query confidencescore completions aura1... --arena Biometric
   ```

## Common Workflows

### New User Journey
1. Complete IR-000 (anchor) → Get initial CS
2. Check score: `query confidencescore score`
3. Review arena breakdown: `query confidencescore arena-breakdown`
4. Complete more IRs to reach 10,000 CS (Verified Human)
5. Monitor progress: `query confidencescore history`

### AI Assistant Workflow
1. Verify user's IR completion offline
2. Generate proof hash (SHA256)
3. Record completion: `tx confidencescore record-completion`
4. Verify recording: `query confidencescore ir-completion`

### Fraud Detection Workflow
1. Detect fraudulent completion
2. Gather evidence (IPFS upload)
3. Submit slash: `tx confidencescore slash`
4. If appealed: Review and resolve: `tx confidencescore resolve-appeal`
