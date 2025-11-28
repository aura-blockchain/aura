# Economic Security Module - RPC Method Mapping

This document provides a complete mapping between proto RPC definitions and their implementations.

## Query Service Methods

| # | RPC Method | Proto Definition | Implementation | Keeper Method(s) | Status |
|---|------------|-----------------|----------------|------------------|--------|
| 1 | Params | `economic_security.proto:45` | `query_server.go:22` | `GetParams()` | ✅ |
| 2 | VestingSchedule | `economic_security.proto:48` | `query_server.go:30` | `GetVestingSchedule()` | ✅ |
| 3 | VestingSchedulesByBeneficiary | `economic_security.proto:51` | `query_server.go:47` | `GetUserVestingSchedules()`, `GetTotalVesting()` | ✅ |
| 4 | VoteLock | `economic_security.proto:54` | `query_server.go:59` | `GetVoteLock()` | ✅ |
| 5 | VoteLocksByOwner | `economic_security.proto:57` | `query_server.go:71` | `GetUserVoteLocks()`, `GetVotingPower()` | ✅ |
| 6 | VotingPower | `economic_security.proto:60` | `query_server.go:83` | `GetVotingPower()` | ✅ |
| 7 | PendingTreasuryTx | `economic_security.proto:63` | `query_server.go:94` | `GetPendingTreasuryTx()` | ✅ |
| 8 | PendingTreasuryTxs | `economic_security.proto:66` | `query_server.go:106` | `GetAllPendingTreasuryTxs()` | ✅ |
| 9 | InflationMetrics | `economic_security.proto:69` | `query_server.go:115` | `GetParams()` | ✅ |
| 10 | InflationAlerts | `economic_security.proto:72` | `query_server.go:128` | `GetInflationAlerts()` | ✅ |
| 11 | LiquidityMiningStats | `economic_security.proto:75` | `query_server.go:142` | `GetLiquidityMiningStats()` | ✅ |
| 12 | MEVStats | `economic_security.proto:78` | `query_server.go:157` | `GetMEVStats()` | ✅ |
| 13 | UserMEVBalance | `economic_security.proto:81` | `query_server.go:171` | `GetUserMEVBalance()` | ✅ |
| 14 | TokenomicsStats | `economic_security.proto:84` | `query_server.go:181` | Multiple getters | ✅ |

## Message Service Methods

| # | RPC Method | Proto Definition | Implementation | Keeper Method(s) | Status |
|---|------------|-----------------|----------------|------------------|--------|
| 1 | CreateVestingSchedule | `economic_security.proto:12` | `msg_server.go:21` | `CreateVestingSchedule()` | ✅ |
| 2 | ReleaseVestedTokens | `economic_security.proto:15` | `msg_server.go:41` | `ReleaseVestedTokens()` | ✅ |
| 3 | RevokeVestingSchedule | `economic_security.proto:18` | `msg_server.go:53` | `RevokeVestingSchedule()` | ✅ |
| 4 | LockVotingTokens | `economic_security.proto:21` | `msg_server.go:65` | `LockVotingTokens()` | ✅ |
| 5 | UnlockVotingTokens | `economic_security.proto:24` | `msg_server.go:78` | `UnlockVotingTokens()` | ✅ |
| 6 | ProposeTreasurySpend | `economic_security.proto:27` | `msg_server.go:90` | `ProposeTreasurySpend()` | ✅ |
| 7 | SignTreasurySpend | `economic_security.proto:30` | `msg_server.go:103` | `SignTreasurySpend()` | ✅ |
| 8 | ExecuteTreasurySpend | `economic_security.proto:33` | `msg_server.go:116` | `ExecuteTreasurySpend()` | ✅ |
| 9 | UpdateParams | `economic_security.proto:36` | `msg_server.go:131` | `SetParams()` | ✅ |
| 10 | AdjustInflationRate | `economic_security.proto:39` | `msg_server.go:146` | `AdjustInflationRate()` | ✅ |

## Keeper Method Locations

| Keeper Method | File | Line | Purpose |
|---------------|------|------|---------|
| CreateVestingSchedule | `keeper/vesting.go` | 19 | Creates new vesting schedule |
| ReleaseVestedTokens | `keeper/vesting.go` | 68 | Releases vested tokens |
| RevokeVestingSchedule | `keeper/vesting.go` | 111 | Revokes vesting schedule |
| GetVestingSchedule | `keeper/vesting.go` | 150 | Retrieves vesting schedule |
| GetUserVestingSchedules | `keeper/vesting.go` | 159 | Gets all user's schedules |
| GetTotalVesting | `keeper/vesting.go` | 232 | Gets total vesting amounts |
| LockVotingTokens | `keeper/governance.go` | 60 | Locks tokens for voting |
| UnlockVotingTokens | `keeper/governance.go` | 116 | Unlocks voting tokens |
| GetVotingPower | `keeper/governance.go` | 146 | Calculates voting power |
| GetVoteLock | `keeper/governance.go` | 182 | Retrieves vote lock |
| GetUserVoteLocks | `keeper/governance.go` | 191 | Gets all user's locks |
| GetTotalLockedGovernance | `keeper/governance.go` | 232 | Total locked for governance |
| ProposeTreasurySpend | `keeper/treasury.go` | 18 | Proposes treasury spend |
| SignTreasurySpend | `keeper/treasury.go` | 71 | Signs treasury transaction |
| ExecuteTreasurySpend | `keeper/treasury.go` | 117 | Executes treasury spend |
| GetPendingTreasuryTx | `keeper/treasury.go` | 166 | Gets pending transaction |
| GetAllPendingTreasuryTxs | `keeper/treasury.go` | 175 | Gets all pending txs |
| CheckInflation | `keeper/keeper.go` | 185 | Monitors inflation |
| AdjustInflationRate | `keeper/keeper.go` | 298 | Adjusts inflation rate |
| GetInflationAlerts | `keeper/keeper.go` | 331 | Gets inflation alerts |
| GetLiquidityMiningStats | `keeper/liquidity_mining.go` | 85 | Gets liquidity stats |
| GetMEVStats | `keeper/mev.go` | 206 | Gets MEV statistics |
| GetUserMEVBalance | `keeper/mev.go` | 177 | Gets user MEV balance |
| GetWhaleProtectionTriggers24h | `keeper/whale_protection.go` | 155 | Whale protection triggers |
| GetTaxCollected24h | `keeper/transfer_tax.go` | 84 | Tax collected stats |

## Type Conversion Functions

| Function | File | Purpose |
|----------|------|---------|
| ParamsFromProto | `types/conversions.go` | Converts proto Params to internal |
| ParamsToProto | `types/conversions.go` | Converts internal Params to proto |
| VestingScheduleToProto | `types/conversions.go` | Converts vesting schedule to proto |
| VestingSchedulesSliceToProto | `types/conversions.go` | Converts slice of schedules |
| VoteLockToProto | `types/conversions.go` | Converts vote lock to proto |
| VoteLocksSliceToProto | `types/conversions.go` | Converts slice of locks |
| PendingTreasuryTxToProto | `types/conversions.go` | Converts treasury tx to proto |
| PendingTreasuryTxsSliceToProto | `types/conversions.go` | Converts slice of treasury txs |
| InflationAlertToProto | `types/conversions.go` | Converts inflation alert to proto |
| InflationAlertsSliceToProto | `types/conversions.go` | Converts slice of alerts |

## Error Types

All error types are defined in `types/errors.go`:

- ErrVestingScheduleNotFound
- ErrVoteLockNotFound
- ErrTxNotFound
- ErrInflationRateTooHigh
- ErrInflationRateTooLow
- ErrMaxSupplyExceeded
- ErrInvalidAmount
- ErrUnauthorized
- And 30+ more...

## Implementation Statistics

- **Total RPC Methods:** 24 (14 queries + 10 messages)
- **Implemented Methods:** 24 (100%)
- **Keeper Methods:** 35+
- **Type Conversions:** 10
- **Error Types:** 40+
- **Proto Files:** 3
- **Implementation Files:** 14

## Verification Commands

```bash
# Build module
cd chain/x/economicsecurity && go build

# Run tests
go test ./...

# Check interfaces
go vet

# Full chain build
cd chain && go build ./cmd/aurad
```

## gRPC Service Registration

Both services are registered in `module.go`:

```go
func (m AppModule) RegisterServices(config ModuleServices) {
    config.RegisterMsgServer(NewMsgServer(m.keeper))
    config.RegisterQueryServer(NewQueryServer(m.keeper))
}
```

The servers implement the generated interfaces from `economic_security_grpc.pb.go`:
- `economicsecuritypb.QueryServer`
- `economicsecuritypb.MsgServer`

## Completion Checklist

- [x] All query methods implemented
- [x] All message methods implemented
- [x] All keeper methods implemented
- [x] All type conversions implemented
- [x] All errors defined
- [x] Module registration complete
- [x] Successful compilation
- [x] Interface compliance verified
- [x] Error handling in place
- [x] Documentation complete

**Status: 100% Complete ✅**
