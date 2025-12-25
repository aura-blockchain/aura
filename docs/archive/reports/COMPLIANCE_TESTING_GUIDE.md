# Compliance Module Testing Guide

## Overview

This guide documents the end-to-end testing of the Aura Compliance module performed on the testnet.

## Setup

### 1. Build the CLI

```bash
cd /home/decri/blockchain-projects/aura/chain
go build -o aurad ./cmd/aurad
```

### 2. Initialize Node

```bash
./aurad init testnode --chain-id test-1
```

### 3. Create Validator Key

```bash
./aurad keys add validator --keyring-backend test
```

### 4. Setup Genesis

```bash
# Add validator to genesis
./aurad genesis add-genesis-account aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc 100000000stake

# Generate gentx
./aurad genesis gentx validator 90000000stake --chain-id test-1 --keyring-backend test

# Collect gentxs
./aurad genesis collect-gentxs
```

### 5. Start Node

```bash
./aurad start
```

Node will be available at:
- REST API: http://localhost:1317
- RPC: tcp://localhost:26657

## Testing Commands

### Test 1: Node Connectivity

```bash
# Check REST API
curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info

# Check latest block
curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/blocks/latest
```

### Test 2: List Compliance Commands

```bash
# Query commands
./aurad query compliance --help

# Transaction commands
./aurad tx compliance --help
```

### Test 3: Submit KYC Record

```bash
# Generate unsigned transaction
./aurad tx compliance submit-kyc \
  aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc \
  3 \
  aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc \
  35b5ff306c0e207ccdf38baeb726df259adc4446ef0d1a6ea011b0b99ba38c53 \
  US \
  --from validator \
  --keyring-backend test \
  --chain-id test-1 \
  --node tcp://localhost:26657 \
  --generate-only

# Submit transaction (requires funded account)
./aurad tx compliance submit-kyc \
  aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc \
  3 \
  aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc \
  35b5ff306c0e207ccdf38baeb726df259adc4446ef0d1a6ea011b0b99ba38c53 \
  US \
  --from validator \
  --keyring-backend test \
  --chain-id test-1 \
  --node tcp://localhost:26657 \
  --gas auto \
  -y
```

### Test 4: Query KYC Record

```bash
./aurad query compliance kyc-record \
  aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc \
  --node tcp://localhost:26657
```

### Test 5: Query AML Profile

```bash
./aurad query compliance aml-profile \
  aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc \
  --node tcp://localhost:26657
```

### Test 6: Screen Sanctions

```bash
# Query existing sanctions screening
./aurad query compliance sanctions \
  aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc \
  --node tcp://localhost:26657

# Submit new sanctions screening
./aurad tx compliance screen-sanctions \
  aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc \
  --from validator \
  --keyring-backend test \
  --chain-id test-1 \
  --node tcp://localhost:26657 \
  --gas auto \
  -y
```

### Test 7: Query Transaction Alerts

```bash
./aurad query compliance alerts \
  aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc \
  --node tcp://localhost:26657
```

### Test 8: Record GDPR Consent

```bash
./aurad tx compliance record-consent \
  aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc \
  "Marketing" \
  true \
  --from validator \
  --keyring-backend test \
  --chain-id test-1 \
  --node tcp://localhost:26657 \
  --gas auto \
  -y
```

### Test 9: Request GDPR Data

```bash
./aurad tx compliance request-data \
  "Personal Data" \
  --from validator \
  --keyring-backend test \
  --chain-id test-1 \
  --node tcp://localhost:26657 \
  --gas auto \
  -y
```

### Test 10: Report Suspicious Activity

```bash
./aurad tx compliance report-suspicious \
  aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc \
  "STRUCTURING" \
  "Multiple rapid transactions in sequence" \
  5.5 \
  --from validator \
  --keyring-backend test \
  --chain-id test-1 \
  --node tcp://localhost:26657 \
  --gas auto \
  -y
```

### Test 11: Generate Tax Report

```bash
./aurad tx compliance generate-tax-report \
  aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc \
  "US" \
  2025 \
  "aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc" \
  --from validator \
  --keyring-backend test \
  --chain-id test-1 \
  --node tcp://localhost:26657 \
  --gas auto \
  -y
```

## KYC Levels Reference

| Level | Name | Description |
|-------|------|-------------|
| 1 | NONE | No KYC verification |
| 2 | BASIC | Basic identity verification |
| 3 | INTERMEDIATE | Government ID verification |
| 4 | ADVANCED | Enhanced due diligence |

## Blocked Jurisdictions

The following jurisdictions are OFAC-sanctioned by default:
- KP (North Korea)
- IR (Iran)
- SY (Syria)
- CU (Cuba)
- RU (Russia)
- BY (Belarus)

Addresses in these jurisdictions will be rejected during sanctions screening.

## Important Notes

### PII Commitment Hash

The `pii-commitment` parameter must be:
- Exactly 64 hexadecimal characters
- Represents the SHA-256 hash of off-chain PII data
- Never store actual PII on-chain

Example generation:
```bash
# Generate SHA-256 hash
echo -n "John Doe DOB:1990-01-15" | sha256sum

# Use the 64-char hex output
```

### Address Format

All addresses must be valid bech32-encoded Aura addresses:
- Prefix: `aura`
- Example: `aura1u49gs78yaaex077cygcw94c9y0ln7lduvg32fc`

### Transaction Submission Requirements

Full transaction submission requires:
1. A funded validator account
2. Proper keyring configuration
3. Network connectivity to the blockchain

For testing transaction structure without submission:
```bash
# Add --generate-only flag to see transaction without broadcasting
--generate-only

# Or use --dry-run for simulation
--dry-run
```

## Module Architecture

### Query Endpoints

- `compliance/kyc-record/{address}` - Get KYC record for address
- `compliance/aml-profile/{address}` - Get AML risk profile
- `compliance/sanctions/{address}` - Get sanctions screening results
- `compliance/alerts/{address}` - Get transaction monitoring alerts
- `compliance/tax-report/{address}` - Get tax report

### Transaction Types

- `submit-kyc` - Submit KYC verification
- `screen-sanctions` - Screen address against sanctions lists
- `record-consent` - Record GDPR consent
- `request-data` - Request GDPR data
- `report-suspicious` - Report suspicious activity
- `generate-tax-report` - Generate tax compliance report

## Genesis Configuration

Key compliance parameters in genesis:

```json
{
  "compliance": {
    "params": {
      "kyc_required": false,
      "minimum_kyc_level": "KYC_LEVEL_UNSPECIFIED",
      "kyc_expiry_days": "365",
      "approved_kyc_providers": [],
      "blocked_jurisdictions": ["KP", "IR", "SY", "CU", "RU", "BY"],
      "transaction_monitoring_enabled": false,
      "velocity_limit_24h": "1000000",
      "single_transaction_limit": "100000",
      "sanctions_screening_enabled": false,
      "gdpr_enabled": false
    }
  }
}
```

## Troubleshooting

### Keyring Issues

If you encounter "key not found" errors:

```bash
# Check if key exists
./aurad keys list --keyring-backend test

# Recreate key if needed
./aurad keys add validator --keyring-backend test --recover
```

### Invalid Address Format

Ensure addresses use correct bech32 format:
```bash
# Check address is valid
./aurad keys show validator --keyring-backend test
```

### Transaction Simulation Errors

For testing without execution:
```bash
# Use --generate-only to create unsigned transaction
--generate-only

# Or use --dry-run for gas estimation
--dry-run
```

## Success Indicators

### Successful KYC Query
- Returns address, KYC level, provider, commitment, jurisdiction
- Or "not found" if no record exists (normal for new address)

### Successful Sanctions Screening
- Returns address, matches, status
- Status: SANCTIONS_CLEAR or SANCTIONS_MATCH

### Successful Transaction Submission
- Returns txhash
- Included in block after confirmation
- Status queryable via: `./aurad query tx {txhash}`

## Integration Testing

### DEX-Compliance Integration

To test KYC gating on DEX swaps:

1. Submit KYC record for user address
2. Query compliance status
3. Attempt DEX swap
4. Verify swap is allowed/denied based on KYC level

### Inter-Module Queries

Test module cross-communication:

```bash
# Compliance module can be queried by other modules
./aurad query compliance kyc-record [address]
```

## Performance Benchmarks

Typical latencies observed during testing:
- Query endpoints: <100ms
- Sanctions screening: <100ms
- Transaction submission: <1 second
- Block confirmation: ~2-6 seconds

## References

- Compliance Module Source: `/home/decri/blockchain-projects/aura/chain/x/compliance/`
- Test Report: `/home/decri/blockchain-projects/aura/chain/COMPLIANCE_TEST_REPORT.md`
- Implementation Details: See module keeper and message server code

