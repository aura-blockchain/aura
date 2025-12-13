# Aura Testnet Validator Reconfiguration

## Problem Identified
The Aura testnet was configured with only 1 actual validator (validator-1) having 100% voting power. The other 3 nodes (validator-2, validator-3, validator-4) were running as full nodes but not as validators.

## Root Causes Found

### 1. Bond Denomination Mismatch
- **Issue**: Genesis file had `bond_denom: "stake"` in staking module parameters
- **Impact**: Gentx transactions using "uaura" were rejected, preventing validator creation
- **Fix**: Changed `bond_denom` to "uaura" in genesis.json

### 2. File Permissions in Docker Volumes
- **Issue**: Files copied to Docker volumes had root ownership (UID 0) instead of aura user (UID 1000)
- **Impact**: aurad process couldn't read config files, causing constant restarts
- **Fix**: Added `chown -R 1000:1000 /home/aura/.aura` to populate-volumes.sh script

### 3. Default Genesis Validator
- **Issue**: `aurad init` creates a default "genesis-validator" with 90,000 AURA staked
- **Impact**: This validator was present alongside the 4 intended validators
- **Status**: Needs verification after fixes are applied

## Completed Steps

1. **Stopped current testnet**: `docker compose -f docker-compose.testnet.yml down -v`
2. **Cleaned old data**: `sudo rm -rf testnet-data`
3. **Rebuilt aurad binary**: Ensured latest version
4. **Ran testnet-init.sh**: Created fresh genesis with 4 validators
   - Each validator stakes 900,000 AURA (900,000,000,000 uaura)
   - Equal voting power distribution (25% each)
5. **Fixed bond_denom**: Updated genesis to use "uaura" instead of "stake"
6. **Fixed populate-volumes.sh**: Added chown command to set correct file ownership
7. **Updated testnet-init.sh**: Permanent fix for future initializations

## Validators Configuration

### Validator Details (from genesis)
```
validator-1:
  Address: aura16npecr9x24xzkuxv0nhxmyynmctejum3g86l82
  Operator: auravaloper16npecr9x24xzkuxv0nhxmyynmctejum3n4thl5
  Node ID: 58dbac6d0b3e2fbcf45b6e3842d7f2b4ced5e30b
  IP: 172.26.0.10
  Stake: 900,000,000,000 uaura

validator-2:
  Address: aura1wm8w7ml8elxq9tpc3hph0xgydfkvjuy253eeuv
  Operator: auravaloper1wm8w7ml8elxq9tpc3hph0xgydfkvjuy20rg3yj
  Node ID: 457803b60b975e1c06cc1c17e4d4c83d40cad4d4
  IP: 172.26.0.11
  Stake: 900,000,000,000 uaura

validator-3:
  Address: aura1gpu66zdke67uluqnh3jk6a46jfd5a8z47uqgks
  Operator: auravaloper1gpu66zdke67uluqnh3jk6a46jfd5a8z49w3qww
  Node ID: 8f49e3207fbd6c64d3eacc204148433b60152215
  IP: 172.26.0.12
  Stake: 900,000,000,000 uaura

validator-4:
  Address: aura13kzmhhxvy2d3zvx0zxzt0e2q3awfhpc7ndeg0w
  Operator: auravaloper13kzmhhxvy2d3zvx0zxzt0e2q3awfhpc7glgqhs
  Node ID: f807cfdb325afd804d8e42de83bc30e1cc880a93
  IP: 172.26.0.13
  Stake: 900,000,000,000 uaura
```

## Modified Files

1. **scripts/testnet-init.sh**
   - Added chown command to populate-volumes.sh template (line 401-402)

2. **testnet-data/populate-volumes.sh** (generated)
   - Includes permission fix: `chown -R 1000:1000 /home/aura/.aura`

3. **testnet-data/validator-*/config/genesis.json**
   - Fixed `app_state.staking.params.bond_denom` from "stake" to "uaura"
   - Contains 4 gentx transactions

## Verification Steps (To Be Completed)

1. **Check all containers are running**:
   ```bash
   docker ps --filter "name=aura-validator"
   ```

2. **Verify 4 validators are active**:
   ```bash
   curl -s http://localhost:27657/validators | jq '.result.validators | length'
   ```

3. **Check voting power distribution** (should be 90000 each for 25%):
   ```bash
   curl -s http://localhost:27657/validators | jq '.result.validators[] | {voting_power}'
   ```

4. **Query via REST API**:
   ```bash
   curl -s http://localhost:2317/cosmos/staking/v1beta1/validators | jq '.validators | map({moniker: .description.moniker, tokens, status})'
   ```

## Expected Results

- **Total validators**: 4
- **Voting power per validator**: 90,000 (representing 900,000 AURA staked)
- **Voting power distribution**: 25% each
- **Consensus**: Should work with 3/4 validators (BFT tolerance)

## Remaining Issues

1. **Container startup problems**: Validators may still be experiencing issues starting
2. **Genesis processing**: Need to verify all 4 gentx transactions are processed during InitGenesis
3. **Minimum stake requirement**: Cosmos SDK requires 824,645,180,800 uaura minimum per validator

## Next Steps if Issues Persist

1. Increase staking amount to 1,000,000 AURA (1,000,000,000,000 uaura) for safety margin
2. Verify gentx processing in logs: `docker logs aura-validator-1 | grep genutil`
3. Check for any consensus errors: `docker logs aura-validator-1 | grep -i "error\|fail"`
4. Consider using `aurad testnet init-files` command if available
5. Manual verification of all 4 priv_validator_key.json files have unique keys

## Files Modified in This Session

- `/home/hudson/blockchain-projects/aura/scripts/testnet-init.sh` - Added chown to populate script
- `/home/hudson/blockchain-projects/aura/testnet-data/populate-volumes.sh` - Added chown command
- `/home/hudson/blockchain-projects/aura/testnet-data/validator-*/config/genesis.json` - Fixed bond_denom

## Port Mappings

```
Validator-1: RPC=27657, API=2317, gRPC=10090, P2P=27656, Metrics=27660
Validator-2: RPC=27757, API=2417, gRPC=10190, P2P=27756, Metrics=27760
Validator-3: RPC=27857, API=2517, gRPC=10290, P2P=27856, Metrics=27860
Validator-4: RPC=27957, API=2617, gRPC=10390, P2P=27956, Metrics=27960
```

## Date
Reconfiguration performed: 2025-12-13
