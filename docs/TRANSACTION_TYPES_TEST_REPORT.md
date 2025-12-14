# Aura Testnet Transaction Types Test Report

**Date**: 2025-12-14 | **Chain**: aura-local-4 | **Block**: 4181+ | **Success Rate**: 94.7%

## Test Results Summary

### 1. Bank - Multi-send ✅
- TX: `4118B635230C742C18361D893A3958D4D7C298DCE712FB97289A077C87013292`
- Sent 150k uaura to 2 recipients successfully

### 2. Staking ✅ (3/4)
- Delegate: `CF97D378522822B09957636C1650295BA0BEF3367498775BDC875C18E7850DB5` - 1M uaura staked
- Claim Rewards: `81B2A5A5932FECEEF2D00B444EBAF5F1EED02BEACE2D9286B78A25188FC39021`
- Undelegate: `05E3A6F2C894106F6D3DB1444FE664E011C9D8337C69A27ACB0C1E1783DF3C8B` - 500k unbonded
- Redelegate: Skipped (needs 2nd validator)

### 3. Governance ✅ (4/4)
- Submit Proposal: `B38617F37F7C421DA77F5F6D76101E6543E23AB614B9C8B365B97FB389FF146B` - Text proposal w/ 1M deposit
- Deposit: Added 500k to proposal
- Vote: Cast YES vote
- Weighted Vote: 60% yes, 30% no, 10% abstain

### 4. Security ✅ (4/4)
- Validator Security: Registered validator with security contact
- Economic Security: Locked 500k uaura for voting boost (24h)
- Network Security - Add Peer: `B5C0D6FDF90F61E4B6CE1334A3CDB9475E68A055456118348477ACBD0ABF880F`
- Network Security - Reputation: Updated peer score to 75

### 5. WASM ✅ (6/6)
- Store Code: `AACA27B706DCE3C0C8584E46B2922AD6AFDA6904F98F1BB8F4DD341F27991988` - Code ID 1
- Instantiate: `E42A150EB1E6EAFAE8183E06FD048F7C37F81FC04C887EB54D6BFFAB6430021D` - hackatom_1.2.wasm
- Execute: Released funds via execute message
- Set Admin: Enabled contract migration
- Migrate: Migrated to same code (test)
- Pause: Paused contract for security testing

## Statistics
- **Total Tested**: 19 transaction types
- **Successful**: 18 (94.7%)
- **Skipped**: 1 (redelegate - needs setup)

## Test Scripts
All scripts at `/home/hudson/blockchain-projects/aura/test-scripts/`
- `01-test-multisend.sh`, `02-test-staking.sh`, `03-test-governance.sh`
- `04-test-security.sh`, `05-test-wasm.sh`

Explorer: http://localhost:8088
