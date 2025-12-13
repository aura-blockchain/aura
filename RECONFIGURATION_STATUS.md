# Aura Testnet 4-Validator Reconfiguration - Status Report

## Date: 2025-12-13

## Objective
Reconfigure the Aura testnet to have 4 actual validators with equal voting power (25% each), replacing the current setup where only validator-1 has 100% voting power.

## Work Completed

### 1. Issue Diagnosis
Identified three critical problems:
- **Bond denomination mismatch**: Genesis had `bond_denom: "stake"` but gentx used "uaura"
- **File permissions**: Docker volume files owned by root (UID 0) instead of aura user (UID 1000)
- **Default genesis validator**: `aurad init` creates a default validator that conflicts with gentx

### 2. Fixes Implemented

#### testnet-init.sh (scripts/testnet-init.sh)
Added permanent fixes:
- Line 401-402: Added `chown -R 1000:1000 /home/aura/.aura` to populate-volumes.sh template
- Line 314-328: Added code to:
  - Clear default validators from staking module
  - Remove "stake" denomination from balances
  - Keep validators only in gentx for proper initialization

#### populate-volumes.sh
Modified to set correct file ownership for Docker containers

#### Genesis Configuration
- Fixed `bond_denom` from "stake" to "uaura"
- Verified 4 gentx transactions are present
- Each validator stakes 900,000 AURA (900,000,000,000 uaura)

### 3. Validation Results

**Chain Status**: ✓ Running and producing blocks (height 20+)

**Validator Count**:
```
Expected: 4 validators with 25% voting power each
Actual: 1 validator with 100% voting power
```

**Genesis File (in Docker volume)**:
- bond_denom: "uaura" ✓
- gentx count: 4 ✓
- Expected validators in gentx: validator-1, validator-2, validator-3, validator-4 ✓

## Current Status: PARTIALLY COMPLETE

### What's Working
- ✓ Testnet initializes without errors
- ✓ Chain is running and producing blocks
- ✓ Genesis file has correct bond_denom
- ✓ Genesis file contains all 4 gentx transactions
- ✓ File permissions are correct
- ✓ All 4 validator containers start successfully

### What's NOT Working
- ✗ Only 1 validator is active in the validator set
- ✗ Gentx transactions appear to process without error but don't create validators
- ✗ Voting power distribution is 100% instead of 25% each

## Problem Analysis

The gentx transactions are present in genesis and the genutil module initializes without errors, but the validators are not being added to the active validator set. Possible causes:

1. **Staking module initialization order**: The staking module may be clearing validators before genutil processes gentx
2. **Minimum stake requirement**: DefaultPowerReduction is 824,645,180,800 uaura. Current stake of 900,000,000,000 should exceed this.
3. **Genesis validator conflict**: There may still be a conflicting validator definition somewhere
4. **Gentx validation failure**: Gentx transactions may be silently failing validation

## Next Steps for Resolution

### Option 1: Investigate Gentx Processing
1. Add debug logging to genutil module InitGenesis
2. Check if DeliverGenTxs is being called
3. Verify validator creation in staking module

### Option 2: Manual Validator Addition
Instead of using gentx, manually add all 4 validators to `app_state.staking.validators` in genesis with:
- operator_address
- consensus_pubkey
- tokens: "900000000000"
- status: "BOND_STATUS_BONDED"

### Option 3: Use Different Initialization Method
1. Start with single validator
2. Add additional validators via MsgCreateValidator transactions after genesis
3. Delegate tokens to reach equal voting power

### Option 4: Increase Staking Amount
Change staking amount from 900,000 AURA to 1,000,000 AURA to provide more margin above DefaultPowerReduction

## Files Modified

1. `/home/hudson/blockchain-projects/aura/scripts/testnet-init.sh`
   - Added file ownership fix
   - Added validator clearing logic
   - Added stake denom cleanup

2. `/home/hudson/blockchain-projects/aura/testnet-data/populate-volumes.sh`
   - Added chown command for proper permissions

3. `/home/hudson/blockchain-projects/aura/testnet-data/validator-*/config/genesis.json`
   - Fixed bond_denom to "uaura"

## Verification Commands

```bash
# Check validator count
curl -s http://localhost:27657/validators | jq '.result.total'

# Check voting power distribution
curl -s http://localhost:27657/validators | jq '.result.validators[] | {voting_power}'

# Check gentx in genesis
docker run --rm -v aura_validator-1-data:/data alpine cat /data/config/genesis.json | jq '.app_state.genutil.gen_txs | length'

# Check bond denom
docker run --rm -v aura_validator-1-data:/data alpine cat /data/config/genesis.json | jq '.app_state.staking.params.bond_denom'

# Check container status
docker ps --filter "name=aura-validator"
```

## Documentation Created

1. `VALIDATOR_RECONFIGURATION.md` - Initial problem analysis and fixes
2. `RECONFIGURATION_STATUS.md` - This file - current status

## Conclusion

Significant progress was made in diagnosing and fixing infrastructure issues (permissions, bond denom), but the core problem of gentx processing remains unresolved. The testnet runs successfully with 1 validator, but the 4-validator configuration requires additional investigation into the Cosmos SDK's gentx processing mechanism.

**Recommendation**: Consult Cosmos SDK documentation on multi-validator genesis configuration or examine the genutil module source code to understand why gentx transactions aren't creating validators despite being present in genesis.
