# Economic Security Module - Quick Reference

## File Locations and Line Numbers

### Feature 1: Maximum Supply Cap Enforcement
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\keeper\keeper.go`
- CheckSupplyCap: Lines 105-130
- UpdateCirculatingSupply: Lines 132-157
- GetRemainingSupply: Lines 159-171
- **Test**: `keeper_test.go` Lines 11-63

### Feature 2: Inflation Rate Monitoring and Alerts
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\keeper\keeper.go`
- CheckInflation: Lines 177-254
- AdjustInflationRate: Lines 256-284
- createInflationAlert: Lines 224-243
- GetInflationAlerts: Lines 286-297
- **Test**: `keeper_test.go` Lines 65-103

### Feature 3: Liquidity Mining Reward Caps
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\keeper\liquidity_mining.go`
- DistributeLiquidityRewards: Lines 10-71
- GetLiquidityMiningStats: Lines 73-94
- CheckLiquidityRewardCap: Lines 96-112
- **Test**: `keeper_test.go` Lines 105-153

### Feature 4: Vesting Schedules for Team and Early Investors
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\keeper\vesting.go`
- CreateVestingSchedule: Lines 12-60
- ReleaseVestedTokens: Lines 62-98
- RevokeVestingSchedule: Lines 100-127
- GetVestingSchedule: Lines 129-135
- GetUserVestingSchedules: Lines 137-151
- calculateVestedAmount: Lines 153-186
- GetTotalVesting: Lines 188-210
- **Test**: `keeper_test.go` Lines 155-209

### Feature 5: Anti-Whale Mechanisms
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\keeper\whale_protection.go`
- CheckWhaleProtection: Lines 12-88
- UpdateAddressHolding: Lines 90-97
- recordLargeTx: Lines 99-127
- GetLargeTxRecords: Lines 129-141
- GetWhaleProtectionTriggers24h: Lines 143-157
- **Test**: `keeper_test.go` Lines 211-258

### Feature 6: Transfer Tax Options
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\keeper\transfer_tax.go`
- CalculateTransferTax: Lines 10-54
- ProcessTransferTax: Lines 56-71
- calculateDynamicTaxRate: Lines 73-82
- GetTaxCollected24h: Lines 84-89
- **Test**: `keeper_test.go` Lines 260-309

### Feature 7: Minimum Stake Requirements for Governance Proposals
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\keeper\governance.go`
- CheckProposalStake: Lines 18-31
- **Test**: `keeper_test.go` Lines 311-327

### Feature 8: Quadratic Voting for Fair Governance
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\keeper\governance.go`
- CalculateQuadraticVotingPower: Lines 33-52
- sqrt: Lines 162-176
- **Test**: `keeper_test.go` Lines 329-345

### Feature 9: Vote Locking Mechanisms
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\keeper\governance.go`
- LockVotingTokens: Lines 54-102
- UnlockVotingTokens: Lines 104-126
- GetVotingPower: Lines 128-157
- GetVoteLock: Lines 178-184
- GetUserVoteLocks: Lines 186-199
- GetTotalLockedGovernance: Lines 201-219
- **Test**: `keeper_test.go` Lines 347-396

### Feature 10: Treasury Multi-Signature Controls
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\keeper\treasury.go`
- ProposeTreasurySpend: Lines 12-59
- SignTreasurySpend: Lines 61-96
- ExecuteTreasurySpend: Lines 98-141
- GetPendingTreasuryTx: Lines 143-149
- GetAllPendingTreasuryTxs: Lines 151-163
- RejectTreasurySpend: Lines 165-177
- **Test**: `keeper_test.go` Lines 398-469

### Feature 11: Dynamic Fee Adjustment
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\keeper\dynamic_fees.go`
- RecordBlockUtilization: Lines 12-28
- AdjustDynamicFees: Lines 30-75
- CalculateDynamicFee: Lines 77-93
- GetCurrentFeeMultiplier: Lines 95-98
- GetAverageUtilization: Lines 100-112
- **Test**: `keeper_test.go` Lines 471-507

### Feature 12: MEV Redistribution
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\keeper\mev.go`
- CaptureMEV: Lines 12-32
- DistributeMEV: Lines 34-82
- distributeMEVToUsers: Lines 84-141
- addUserMEVBalance: Lines 143-148
- GetUserMEVBalance: Lines 150-158
- ClaimMEVRewards: Lines 160-173
- GetMEVStats: Lines 175-190
- **Test**: `keeper_test.go` Lines 509-565

## Proto Definitions

### Types
**File**: `C:\Users\decri\gitclones\aura\proto\aura\economicsecurity\v1beta1\types.proto`
- TokenomicsConfig: Lines 9-29
- VestingSchedule: Lines 31-54
- WhaleProtection: Lines 71-87
- TransferTaxConfig: Lines 89-117
- LiquidityMiningConfig: Lines 119-145
- GovernanceConfig: Lines 147-170
- VoteLock: Lines 172-189
- TreasuryMultisig: Lines 191-206
- PendingTreasuryTx: Lines 208-233
- DynamicFeeConfig: Lines 235-261
- MEVConfig: Lines 263-283
- InflationAlert: Lines 309-329
- LargeTxRecord: Lines 350-367
- Params: Lines 369-398

### Messages & Queries
**File**: `C:\Users\decri\gitclones\aura\proto\aura\economicsecurity\v1beta1\economic_security.proto`
- Msg Service: Lines 10-40
- Query Service: Lines 42-78
- Message Types: Lines 82-166
- Query Types: Lines 168-268

### Genesis
**File**: `C:\Users\decri\gitclones\aura\proto\aura\economicsecurity\v1beta1\genesis.proto`
- GenesisState: Lines 9-29

## Module Files

### Core Module
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\module.go`
- AppModule: Lines 19-67
- BeginBlock: Lines 44-56

### Message Server
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\msg_server.go`
- All message handlers: Lines 1-150

### Query Server
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\query_server.go`
- All query handlers: Lines 1-179

## Type Definitions

### Types
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\types\types.go`
- Type aliases: Lines 1-91

### Errors
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\types\errors.go`
- All error definitions: Lines 1-90

### Parameters
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\types\params.go`
- DefaultParams: Lines 10-72
- ValidateParams: Lines 74-130
- Validation functions: Lines 132-298

### Genesis
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\types\genesis.go`
- DefaultGenesis: Lines 10-21
- Validate: Lines 23-88

### Parameter Store
**File**: `C:\Users\decri\gitclones\aura\chain\x\economicsecurity\params\params.go`
- Store implementation: Lines 1-80

## Quick Command Reference

### Build & Test
```bash
# From chain directory
cd C:\Users\decri\gitclones\aura\chain

# Generate proto files (requires buf)
cd ../proto
buf generate aura/economicsecurity/v1beta1

# Run all tests
cd ../chain/x/economicsecurity
go test -v ./keeper

# Run specific test
go test -v ./keeper -run TestSupplyCapEnforcement
```

### Integration
```go
// In app.go
import (
    "github.com/aequitas/aura/chain/x/economicsecurity"
    eskeeper "github.com/aequitas/aura/chain/x/economicsecurity/keeper"
    esparams "github.com/aequitas/aura/chain/x/economicsecurity/params"
    estypes "github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// Create keeper
esParamsStore := esparams.NewStore(estypes.DefaultParams())
esKeeper := eskeeper.NewKeeper(esParamsStore)
esModule := economicsecurity.NewAppModule(esKeeper)
```

## Error Reference

### Supply Cap Errors
- `ErrMaxSupplyExceeded`: Minting would exceed max supply
- `ErrSupplyCapAlreadySet`: Cap is immutable
- `ErrInvalidSupplyCap`: Invalid supply value

### Inflation Errors
- `ErrInflationRateTooHigh`: Above maximum
- `ErrInflationRateTooLow`: Below minimum
- `ErrInvalidInflationRate`: Invalid rate value

### Vesting Errors
- `ErrVestingScheduleNotFound`: Schedule doesn't exist
- `ErrVestingAlreadyRevoked`: Already revoked
- `ErrNoVestedTokens`: Nothing to vest yet
- `ErrCliffNotReached`: Before cliff period

### Whale Protection Errors
- `ErrWhaleHoldingLimitExceeded`: Holding limit exceeded
- `ErrWhaleTxLimitExceeded`: Transaction limit exceeded
- `ErrLargeTxCooldownActive`: Cooldown still active

### Governance Errors
- `ErrInsufficientStake`: Below minimum stake
- `ErrVoteLockNotFound`: Lock doesn't exist
- `ErrVoteLockNotExpired`: Can't unlock yet
- `ErrLockDurationTooShort`: Below minimum
- `ErrLockDurationTooLong`: Above maximum

### Treasury Errors
- `ErrInsufficientSignatures`: Not enough signatures
- `ErrTxNotFound`: Transaction not found
- `ErrTxAlreadyExecuted`: Already executed
- `ErrTimelockNotExpired`: Timelock still active
- `ErrInvalidSigner`: Not authorized
- `ErrAlreadySigned`: Duplicate signature

### MEV Errors
- `ErrMEVRedistributionDisabled`: Feature disabled
- `ErrInvalidMEVConfig`: Invalid config
- `ErrInsufficientMEVBalance`: Nothing to claim

## Default Parameters

| Parameter | Default Value | Location |
|-----------|---------------|----------|
| Max Supply | 1,000,000,000 tokens | types/params.go:18 |
| Inflation Rate | 5% (500 bp) | types/params.go:20 |
| Min Inflation | 1% (100 bp) | types/params.go:22 |
| Max Inflation | 10% (1000 bp) | types/params.go:23 |
| Max Holding % | 5% (500 bp) | types/params.go:30 |
| Max Tx % | 1% (100 bp) | types/params.go:31 |
| Large Tx Cooldown | 1 hour | types/params.go:32 |
| LM Rewards | 100M tokens | types/params.go:47 |
| IR Multiplier | 1.2x (12000 bp) | types/params.go:52 |
| Min Proposal Stake | 10k tokens | types/params.go:56 |
| Quadratic Voting | Enabled | types/params.go:57 |
| Min Lock Duration | 30 days | types/params.go:59 |
| Max Lock Duration | 4 years | types/params.go:60 |
| Treasury Threshold | 3 of N | types/params.go:68 |
| Timelock | 24 hours | types/params.go:70 |
| Base Fee | 0.001 tokens | types/params.go:74 |
| Target Utilization | 75% | types/params.go:78 |
| MEV User Share | 40% (4000 bp) | types/params.go:86 |

---

**Quick Reference Version**: 1.0.0
**Last Updated**: 2025-11-13
