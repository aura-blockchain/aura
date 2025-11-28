# Economics Module Proto Types - Quick Start Guide

## Import Statement

```go
import (
    "github.com/aequitas/aura/chain/x/economics/types"
    economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)
```

## Using Type Aliases (Recommended)

The `types` package re-exports all proto types for convenience:

```go
// Works with both local types and proto types
var params *types.Params
var schedule *types.VestingSchedule
var proposal *types.Proposal
var vote *types.Vote
```

## Common Operations

### Working with Params

```go
// Get params
params, err := keeper.GetParams(ctx)
if err != nil {
    return err
}

// Access fee params
baseFee := params.Fees.BaseFee
minGasPrice := params.Fees.MinGasPrice

// Update params
params.Fees.DynamicFeesEnabled = true
err = keeper.SetParams(ctx, params)
```

### Working with Vesting Schedules

```go
// Create a new vesting schedule
schedule := &economicspb.VestingSchedule{
    Id:             "schedule-001",
    Address:        "aura1...",
    OriginalAmount: sdk.NewCoin("uaura", math.NewInt(1000000)),
    VestedAmount:   sdk.NewCoin("uaura", math.NewInt(0)),
    StartTime:      timestamppb.New(startTime),
    EndTime:        timestamppb.New(endTime),
    CliffDuration:  uint64(30 * 24 * 60 * 60), // 30 days in seconds
    VestingType:    economicspb.VestingType_VESTING_TYPE_LINEAR,
    ScheduleType:   economicspb.ScheduleType_SCHEDULE_TYPE_TEAM,
    Revoked:        false,
}

// Save to store
err := keeper.SetVestingSchedule(ctx, schedule)

// Retrieve from store
schedule, err := keeper.GetVestingSchedule(ctx, "schedule-001")

// Calculate vested amount
vestedAmount, err := keeper.CalculateVestedAmount(schedule, time.Now())
```

### Working with Proposals

```go
// Create a proposal
proposal := &economicspb.Proposal{
    Id:            1,
    Title:         "Update Parameters",
    Description:   "Proposal to update fee parameters",
    Proposer:      "aura1...",
    Status:        economicspb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
    Category:      economicspb.ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE,
    SubmitTime:    timestamppb.New(time.Now()),
    DepositEndTime: timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
    TotalDeposit:  sdk.NewCoins(),
}

// Save proposal
err := keeper.SetProposal(ctx, proposal)

// Get proposal
proposal, err := keeper.GetProposal(ctx, 1)

// Check proposal status
if proposal.Status == economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
    // Proposal is in voting period
}
```

### Working with Votes

```go
// Create a vote
vote := &economicspb.Vote{
    ProposalId:   1,
    Voter:        "aura1...",
    Option:       economicspb.VoteOption_VOTE_OPTION_YES,
    Timestamp:    timestamppb.New(time.Now()),
    VotingPower:  math.NewInt(1000000),
    IsSecret:     false,
}

// Save vote
err := keeper.SetVote(ctx, vote)

// Get vote
vote, err := keeper.GetVote(ctx, 1, "aura1...")

// Calculate tally
tally, err := keeper.CalculateTally(ctx, 1)
fmt.Printf("Yes: %s, No: %s, Abstain: %s, Veto: %s\n",
    tally.YesCount, tally.NoCount, tally.AbstainCount, tally.NoWithVetoCount)
```

### Working with Vote Locks

```go
// Create a vote lock
lock := &economicspb.VoteLock{
    Id:          "lock-001",
    Owner:       "aura1...",
    Amount:      sdk.NewCoin("uaura", math.NewInt(1000000)),
    LockStart:   timestamppb.New(time.Now()),
    LockEnd:     timestamppb.New(time.Now().Add(30 * 24 * time.Hour)),
    VotingPower: math.NewInt(1200000), // With time multiplier
    Withdrawn:   false,
}

// Save lock
err := keeper.SetVoteLock(ctx, lock)

// Get lock
lock, err := keeper.GetVoteLock(ctx, "lock-001")
```

## Enum Values

### Vesting Types

```go
economicspb.VestingType_VESTING_TYPE_UNSPECIFIED
economicspb.VestingType_VESTING_TYPE_LINEAR
economicspb.VestingType_VESTING_TYPE_CLIFF
economicspb.VestingType_VESTING_TYPE_GRADED
economicspb.VestingType_VESTING_TYPE_MILESTONE

// Or use type aliases
types.VestingTypeLinear
types.VestingTypeCliff
```

### Schedule Types

```go
economicspb.ScheduleType_SCHEDULE_TYPE_TEAM
economicspb.ScheduleType_SCHEDULE_TYPE_INVESTOR
economicspb.ScheduleType_SCHEDULE_TYPE_ADVISOR
economicspb.ScheduleType_SCHEDULE_TYPE_ECOSYSTEM
economicspb.ScheduleType_SCHEDULE_TYPE_COMMUNITY

// Or use type aliases
types.ScheduleTypeTeam
types.ScheduleTypeInvestor
```

### Proposal Status

```go
economicspb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD
economicspb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
economicspb.ProposalStatus_PROPOSAL_STATUS_PASSED
economicspb.ProposalStatus_PROPOSAL_STATUS_REJECTED
economicspb.ProposalStatus_PROPOSAL_STATUS_FAILED
economicspb.ProposalStatus_PROPOSAL_STATUS_VETOED
economicspb.ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY
economicspb.ProposalStatus_PROPOSAL_STATUS_EXECUTED

// Or use type aliases
types.ProposalStatusVotingPeriod
types.ProposalStatusPassed
```

### Vote Options

```go
economicspb.VoteOption_VOTE_OPTION_YES
economicspb.VoteOption_VOTE_OPTION_NO
economicspb.VoteOption_VOTE_OPTION_ABSTAIN
economicspb.VoteOption_VOTE_OPTION_NO_WITH_VETO

// Or use type aliases
types.VoteOptionYes
types.VoteOptionNo
```

## Time Handling

### Converting to/from Timestamps

```go
import "google.golang.org/protobuf/types/known/timestamppb"

// Go time.Time to proto Timestamp
goTime := time.Now()
protoTime := timestamppb.New(goTime)

// Proto Timestamp to Go time.Time
schedule.StartTime.AsTime()

// Comparing times
if time.Now().After(schedule.EndTime.AsTime()) {
    // Vesting period ended
}
```

### Converting to/from Durations

```go
import "google.golang.org/protobuf/types/known/durationpb"

// Go time.Duration to proto Duration
goDuration := 30 * 24 * time.Hour
protoDuration := durationpb.New(goDuration)

// Proto Duration to Go time.Duration
params.Governance.VotingPeriod.AsDuration()
```

## Amount Handling

### Working with Coins

```go
import (
    "cosmossdk.io/math"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// Create a single coin
coin := sdk.NewCoin("uaura", math.NewInt(1000000))

// Create multiple coins
coins := sdk.NewCoins(
    sdk.NewCoin("uaura", math.NewInt(1000000)),
    sdk.NewCoin("uosmo", math.NewInt(500000)),
)

// Access amount
amount := schedule.OriginalAmount.Amount
amountInt := amount.BigInt()
amountString := amount.String()

// Compare amounts
if schedule.VestedAmount.Amount.LT(schedule.OriginalAmount.Amount) {
    // Still vesting
}

// Math operations
newAmount := schedule.OriginalAmount.Amount.Add(additionalAmount)
```

## Iteration Patterns

### Iterate Over Vesting Schedules

```go
err := keeper.IterateVestingSchedules(ctx, func(schedule *economicspb.VestingSchedule) bool {
    // Process schedule
    fmt.Printf("Schedule %s: %s\n", schedule.Id, schedule.Address)

    // Return true to stop iteration, false to continue
    return false
})
```

### Iterate Over Proposals

```go
err := keeper.IterateProposals(ctx, func(proposal *economicspb.Proposal) bool {
    // Process proposal
    fmt.Printf("Proposal %d: %s\n", proposal.Id, proposal.Title)

    return false
})
```

### Iterate Over Votes

```go
err := keeper.IterateVotes(ctx, proposalID, func(vote *economicspb.Vote) bool {
    // Process vote
    fmt.Printf("Vote from %s: %s\n", vote.Voter, vote.Option)

    return false
})
```

## Default Values

```go
// Get default params
params := types.DefaultParams()

// Get default for specific category
feeParams := types.DefaultFeeParams()
vestingParams := types.DefaultVestingParams()
governanceParams := types.DefaultGovernanceParams()
```

## Validation

```go
// Validate all params
err := types.ValidateParams(params)

// Validate specific category
err := types.ValidateFeeParams(params.Fees)
err := types.ValidateGovernanceParams(params.Governance)
```

## Common Mistakes to Avoid

1. **Don't forget to convert timestamps**:
   ```go
   // Wrong
   if time.Now().Unix() > schedule.EndTime

   // Correct
   if time.Now().After(schedule.EndTime.AsTime())
   ```

2. **Use proto field names, not old JSON names**:
   ```go
   // Wrong
   schedule.ScheduleId
   schedule.TotalAmount

   // Correct
   schedule.Id
   schedule.OriginalAmount
   ```

3. **Use codec marshaling, not JSON**:
   ```go
   // Wrong
   bz, _ := json.Marshal(schedule)

   // Correct
   bz, _ := keeper.cdc.Marshal(schedule)
   ```

4. **Access Amount properly**:
   ```go
   // Wrong
   amount := schedule.OriginalAmount // This is a Coin

   // Correct
   amount := schedule.OriginalAmount.Amount // This is math.Int
   amountString := amount.String()
   ```

## Resources

- Full migration guide: `MIGRATION_SUMMARY.md`
- Proto definitions: `/proto/aura/economics/v1beta1/`
- Type aliases: `/chain/x/economics/types/aliases.go`
- Validation: `/chain/x/economics/types/validation.go`
- Defaults: `/chain/x/economics/types/defaults.go`
