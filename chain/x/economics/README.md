# Economics Module

The Economics module consolidates the functionality of the `economicsecurity` and `governance` modules into a single, cohesive module for managing all economic and governance aspects of the Aura blockchain.

## Overview

This module provides comprehensive economic and governance features including:
- **Dynamic Fee Management**: Automatic fee adjustment based on network utilization
- **Transfer Tax**: Optional tax on token transfers
- **Vesting Schedules**: Token vesting with multiple schedule types (linear, milestone, cliff)
- **Vote Locking**: Lock tokens to gain voting power
- **Treasury Management**: Multisig treasury with timelock protection
- **Governance**: Complete proposal and voting system with multiple voting types
- **Economic Monitoring**: Track inflation, large transactions, and whale protection
- **MEV Management**: MEV redistribution and tracking

## Module Structure

```
economics/
├── keeper/
│   ├── keeper.go       # Core keeper with genesis and parameter operations
│   ├── fees.go         # Dynamic fee and transfer tax management
│   ├── vesting.go      # Vesting schedules and vote locks
│   └── governance.go   # Proposals, voting, treasury, and economic monitoring
├── types/
│   ├── keys.go         # Store keys and key prefixes
│   ├── errors.go       # Error definitions
│   ├── types.go        # Core data structures
│   └── genesis.go      # Genesis state and parameters
└── module.go           # Module implementation with ABCI hooks
```

## Key Features

### 1. Dynamic Fee Management

The module supports dynamic fee adjustment based on network congestion:

```go
// Record block utilization
keeper.RecordBlockUtilization(ctx, utilization)

// Get current fee multiplier
multiplier, _ := keeper.GetCurrentFeeMultiplier(ctx)

// Calculate adjusted fee
adjustedFee, _ := keeper.CalculateFee(ctx, baseFee)
```

**Parameters:**
- `DynamicFeesEnabled`: Enable/disable dynamic fees
- `BaseFeeMultiplier`: Current fee multiplier
- `MaxFeeMultiplier`: Maximum fee multiplier
- `TargetUtilization`: Target network utilization (%)
- `AdjustmentSpeed`: Speed of fee adjustments

### 2. Transfer Tax

Optional tax on token transfers with configurable rate and recipient:

```go
// Calculate transfer tax
tax, _ := keeper.CalculateTransferTax(ctx, amount)

// Update tax configuration
keeper.SetTransferTaxConfig(ctx, enabled, rate, recipient)
```

**Parameters:**
- `TransferTaxEnabled`: Enable/disable transfer tax
- `TransferTaxRate`: Tax rate (e.g., "0.001" for 0.1%)
- `TransferTaxRecipient`: Address to receive tax

### 3. Vesting Schedules

Support for multiple vesting types:

- **Linear**: Tokens vest linearly over time
- **Cliff-then-Linear**: Cliff period followed by linear vesting
- **Milestone**: Vesting based on milestones

```go
// Create vesting schedule
schedule := &types.VestingSchedule{
    ScheduleId: "schedule-1",
    BeneficiaryAddress: "aura1...",
    TotalAmount: "1000000",
    VestingType: types.VestingTypeLinear,
    StartTime: startTime,
    EndTime: endTime,
}
keeper.SetVestingSchedule(ctx, schedule)

// Calculate vested amount
vestedAmount, _ := keeper.CalculateVestedAmount(schedule, currentTime)

// Release vested tokens
releasable, _ := keeper.ReleaseVestedTokens(ctx, scheduleID)
```

### 4. Vote Locking

Lock tokens to gain voting power with duration-based bonuses:

```go
// Create vote lock
lock := &types.VoteLock{
    LockId: "lock-1",
    Owner: "aura1...",
    Amount: "1000",
    UnlockTime: unlockTime,
}
keeper.SetVoteLock(ctx, lock)

// Calculate voting power (includes duration bonus)
votingPower := keeper.CalculateVotingPower(amount, lockDuration)
```

### 5. Governance

Comprehensive proposal and voting system:

**Proposal Types:**
- Text proposals
- Parameter changes
- Software upgrades
- Spending proposals
- Emergency proposals
- Constitution changes

**Voting Options:**
- Yes
- No
- Abstain
- No with Veto

**Advanced Features:**
- Weighted voting (split vote across options)
- Secret ballot voting (commit-reveal)
- Quadratic voting
- Vote delegation
- Snapshot voting

```go
// Create proposal
proposal := &types.Proposal{
    Title: "Upgrade Proposal",
    Description: "Upgrade to v2.0",
    Category: types.ProposalCategoryParameterChange,
    Proposer: "aura1...",
}
keeper.SetProposal(ctx, proposal)

// Submit vote
vote := &types.Vote{
    ProposalId: proposalID,
    Voter: "aura1...",
    Option: types.VoteOptionYes,
    Weight: "1000",
}
keeper.SetVote(ctx, vote)

// Calculate tally
tally, _ := keeper.CalculateTally(ctx, proposalID)
```

### 6. Treasury Management

Multisig treasury with timelock protection:

```go
// Configure treasury multisig
multisig := &types.TreasuryMultisig{
    TreasuryAddress: "aura1...",
    Signers: []string{"aura1...", "aura2...", "aura3..."},
    Threshold: 2,
    TimelockDelay: 86400, // 1 day
}
keeper.SetTreasuryMultisig(ctx, multisig)

// Propose treasury transaction
tx := &types.PendingTreasuryTx{
    TxId: "tx-1",
    Proposer: "aura1...",
    Recipient: "aura2...",
    Amount: "10000",
    Description: "Development funding",
}
keeper.SetPendingTreasuryTx(ctx, tx)
```

### 7. Economic Monitoring

Track inflation, large transactions, and whale activity:

```go
// Record inflation alert
alert := &types.InflationAlert{
    AlertId: "alert-1",
    AlertType: types.InflationAlertTypeAboveTarget,
    CurrentRate: 8000,
    TargetRate: 7000,
}
keeper.SetInflationAlert(ctx, alert)

// Record large transaction
record := &types.LargeTxRecord{
    TxHash: "0x123...",
    Sender: "aura1...",
    Amount: "1000000",
    Flagged: true,
}
keeper.SetLargeTxRecord(ctx, record)
```

### 8. MEV Management

MEV redistribution and tracking:

```go
// Set user MEV balance
keeper.SetUserMEVBalance(ctx, address, balance)

// Track total MEV pending
keeper.SetTotalMEVPending(ctx, amount)
```

## Store Keys

The module uses the following key prefixes:

### Fee Management (0x01-0x04)
- `0x01`: DynamicFeeConfig
- `0x02`: TransferTaxConfig
- `0x03`: FeeMultiplier
- `0x04`: UtilizationHistory

### Vesting (0x10-0x13)
- `0x10`: VestingSchedule
- `0x11`: UserVestingIndex
- `0x12`: VoteLock
- `0x13`: UserVoteLockIndex

### Treasury (0x20-0x23)
- `0x20`: TreasuryMultisig
- `0x21`: PendingTreasuryTx
- `0x22`: TreasuryBalance
- `0x23`: TreasuryTransactionCounter

### Governance (0x30-0x38)
- `0x30`: Proposal
- `0x31`: Vote
- `0x32`: Deposit
- `0x33`: NextProposalID
- `0x34`: VoteDelegation
- `0x35`: SnapshotVote
- `0x36`: VoteCommitment
- `0x37`: VetoRequest
- `0x38`: TokenLock

### Economic Monitoring (0x40-0x44)
- `0x40`: InflationAlert
- `0x41`: LargeTxRecord
- `0x42`: LastLargeTxTime
- `0x43`: AddressHolding
- `0x44`: PreviousInflation

### MEV (0x50-0x52)
- `0x50`: UserMEVBalance
- `0x51`: TotalMEVPending
- `0x52`: TotalBurned

### State Tracking (0x60-0x62)
- `0x60`: CurrentHeight
- `0x61`: CurrentTime
- `0x62`: Params

## Parameters

The module maintains a comprehensive set of parameters in `types.Params`:

```go
type Params struct {
    // Fee parameters
    DynamicFeesEnabled     bool
    BaseFeeMultiplier      string
    MaxFeeMultiplier       string
    TargetUtilization      uint64
    AdjustmentSpeed        string
    TransferTaxEnabled     bool
    TransferTaxRate        string
    TransferTaxRecipient   string

    // Vesting parameters
    MinVestingDuration     int64
    MaxVestingDuration     int64
    MinCliffDuration       int64
    MinLockDuration        int64
    MaxLockDuration        int64

    // Treasury parameters
    TreasuryMinSigners     uint32
    TreasuryMaxSigners     uint32
    TreasuryTimelockDelay  int64

    // Governance parameters
    MinDepositAmount       string
    DepositPeriod          int64
    VotingPeriod           int64
    QuorumThreshold        string
    PassThreshold          string
    VetoThreshold          string
    ExecutionDelay         int64
    VetoCoSignersRequired  uint32
    WeightedVotingEnabled  bool
    SecretBallotEnabled    bool
    QuadraticVotingEnabled bool
    CategoryParams         []*CategoryParams

    // Inflation parameters
    TargetInflationRate    uint64
    MaxInflationRate       uint64
    MinInflationRate       uint64
    InflationCheckInterval uint64

    // Whale protection parameters
    WhaleProtectionEnabled bool
    MaxHoldingPercentage   string
    MaxTxPercentage        string
    LargeTxCooldown        int64

    // Liquidity mining parameters
    LiquidityMiningEnabled bool
    RewardCapPerEpoch      string

    // MEV parameters
    MevRedistributionEnabled  bool
    MevRedistributionStrategy MEVRedistributionStrategy
}
```

## Genesis State

The module's genesis state includes all economic and governance data:

```go
type GenesisState struct {
    Params             *Params
    VestingSchedules   []*VestingSchedule
    VoteLocks          []*VoteLock
    TreasuryMultisig   *TreasuryMultisig
    PendingTreasuryTxs []*PendingTreasuryTx
    TreasuryBalance    string
    Proposals          []*Proposal
    Votes              []*Vote
    Deposits           []*Deposit
    VoteDelegations    []*VoteDelegation
    VetoRequests       []*VetoRequest
    SnapshotVotes      []*SnapshotVote
    VoteCommitments    []*VoteCommitment
    TokenLocks         []*TokenLock
    NextProposalId     uint64
    InflationAlerts    []*InflationAlert
    LargeTxRecords     []*LargeTxRecord
    LastLargeTxTimes   map[string]int64
    AddressHoldings    map[string]string
    UserMevBalances    map[string]string
    TotalMevPending    string
    TotalBurned        string
    PreviousInflation  uint64
}
```

## ABCI Hooks

### BeginBlock

The module's `BeginBlock` performs:
1. Sets current block height and time
2. Records block utilization for dynamic fees
3. Checks inflation periodically
4. Processes vesting schedule releases
5. Processes expired vote locks

### EndBlock

The module's `EndBlock` performs:
1. Updates proposal statuses based on time
2. Finalizes voting periods
3. Triggers proposal execution

## Integration

To integrate the economics module into your app:

```go
import (
    "github.com/aequitas/aura/chain/x/economics"
    economicskeeper "github.com/aequitas/aura/chain/x/economics/keeper"
    economicstypes "github.com/aequitas/aura/chain/x/economics/types"
)

// Create keeper
economicsKeeper := economicskeeper.NewKeeper(
    appCodec,
    storeService,
    storeKey,
    authority,
)

// Create module
economicsModule := economics.NewAppModule(appCodec, economicsKeeper)

// Register in module manager
app.ModuleManager = module.NewManager(
    // ... other modules
    economicsModule,
)
```

## Error Handling

The module provides comprehensive error definitions in `types/errors.go`:
- General errors (unauthorized, invalid amount, etc.)
- Fee management errors
- Vesting errors
- Treasury errors
- Governance errors (proposal, voting, deposit, delegation, veto)
- Secret ballot voting errors
- Quadratic voting errors
- Token lock errors
- Inflation errors
- Whale protection errors
- Liquidity mining errors
- MEV errors

## Production Considerations

1. **Protobuf Definitions**: This implementation uses Go structs. In production, generate these from `.proto` files.

2. **Message Handlers**: Implement `MsgServer` for handling transactions.

3. **Query Handlers**: Implement `QueryServer` for gRPC queries.

4. **Events**: Emit events for state changes (proposals created, votes cast, etc.).

5. **Hooks**: Implement hooks for cross-module communication (e.g., bank module for token transfers).

6. **Invariants**: Add invariant checks for state consistency.

7. **Migrations**: Implement upgrade handlers for parameter changes.

## Code Statistics

- **Total Lines**: ~2,793
- **Files**: 9
- **Keeper Functions**: 80+
- **Type Definitions**: 30+
- **Parameters**: 35+

## Events

### EventCreateVestingSchedule
Emitted when vesting schedule is created.

**Attributes**: `beneficiary`, `total_amount`, `vesting_type`

### EventReleaseVestedTokens
Emitted when vested tokens are released.

**Attributes**: `beneficiary`, `amount`, `schedule_id`

### EventRevokeVestingSchedule
Emitted when vesting schedule is revoked.

**Attributes**: `schedule_id`, `beneficiary`, `revoker`

### EventSubmitProposal
Emitted when governance proposal is submitted.

**Attributes**: `proposal_id`, `proposer`, `category`

### EventVote
Emitted when vote is cast on proposal.

**Attributes**: `proposal_id`, `voter`, `option`

### EventDelegateVote
Emitted when voting power is delegated.

**Attributes**: `delegator`, `delegate`

### EventLockVotingTokens
Emitted when tokens are locked for voting power.

**Attributes**: `owner`, `amount`, `lock_duration`

### EventProposeTreasurySpend
Emitted when treasury spend is proposed.

**Attributes**: `tx_id`, `proposer`, `amount`, `recipient`

### EventExecuteTreasurySpend
Emitted when treasury spend is executed.

**Attributes**: `tx_id`, `amount`, `recipient`

### EventAdjustInflationRate
Emitted when inflation rate is adjusted.

**Attributes**: `old_rate`, `new_rate`, `reason`

## Cosmos SDK Patterns

This module follows Cosmos SDK v0.50+ patterns:
- Uses `cosmossdk.io/core/store` for KV store operations
- Implements `appmodule.AppModule` interface
- Uses context-based operations
- Follows deterministic execution principles
- Implements proper ABCI hooks
