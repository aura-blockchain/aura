# Economics Module Protobuf Definitions

This directory contains the comprehensive protobuf definitions for the consolidated Economics module, which merges the functionality of the `economicsecurity` and `governance` modules into a unified economics management system.

## Files

### 1. `economics.proto` - Main Types File

**Purpose**: Defines all core data types, parameters, and enums for the economics module.

**Key Components**:

#### Parameters
- **`Params`**: Unified economics module parameters containing:
  - `FeeParams`: Fee-related settings (base fee, min gas price, dynamic fees, burn percentage)
  - `VestingParams`: Vesting configuration (min/max duration, early unlock penalty)
  - `TreasuryParams`: Treasury settings (multisig, timelock, spending limits)
  - `GovernanceParams`: Governance rules (voting periods, quorum, thresholds, delegation)
  - `MEVParams`: MEV protection and redistribution settings
  - `WhaleProtectionParams`: Anti-whale mechanisms (transfer limits, cooldowns)
  - `LiquidityMiningParams`: Liquidity incentive configuration
  - `TokenomicsParams`: Overall tokenomics (max supply, inflation rates)

#### Core Types
- **`VestingSchedule`**: Token vesting schedules with cliff and duration support
- **`Proposal`**: Governance proposals with categories, status, and tally tracking
- **`Vote`**: Individual votes with support for secret ballots and weighted voting
- **`Deposit`**: Proposal deposits
- **`VoteLock`**: Locked tokens for voting power boost
- **`VoteDelegation`**: Voting power delegation with category support
- **`PendingTreasuryTx`**: Multi-signature treasury transactions

#### Statistics Types
- **`InflationMetrics`**: Tracks inflation rates and circulating supply
- **`MEVStats`**: MEV capture and redistribution statistics
- **`LiquidityMiningStats`**: Liquidity mining rewards tracking

#### Enums
- **`VestingType`**: Linear, cliff, graded, milestone
- **`ScheduleType`**: Team, investor, advisor, ecosystem, community
- **`ProposalStatus`**: Full lifecycle tracking (deposit → voting → execution)
- **`ProposalCategory`**: Text, parameter change, upgrade, spending, emergency, constitution
- **`VoteOption`**: Yes, no, abstain, no with veto
- **`MEVRedistributionStrategy`**: Various redistribution methods

### 2. `genesis.proto` - Genesis State

**Purpose**: Defines the initial state of the economics module at chain genesis.

**Contents**:
- Module parameters
- Initial vesting schedules
- Initial proposals and votes
- Vote locks and delegations
- Pending treasury transactions
- Inflation and MEV statistics
- Counter states (next IDs for proposals, schedules, etc.)

### 3. `query.proto` - Query Service

**Purpose**: Defines all gRPC query endpoints with HTTP annotations.

**Query Categories**:

#### Vesting Queries
- `VestingSchedule`: Get schedule by ID
- `VestingSchedulesByAddress`: Get all schedules for an address
- `AllVestingSchedules`: Get all schedules

#### Governance Queries
- `Proposal`: Get proposal by ID
- `Proposals`: List all proposals with filters
- `Vote`: Get specific vote
- `Votes`: Get all votes for a proposal
- `Deposit`: Get specific deposit
- `Deposits`: Get all deposits for a proposal
- `TallyResult`: Get current tally

#### Vote Lock Queries
- `VoteLock`: Get lock by ID
- `VoteLocksByOwner`: Get all locks for an owner
- `VotingPower`: Calculate voting power (including locks and delegation)
- `VoteDelegations`: Get all delegations

#### Treasury Queries
- `PendingTreasuryTx`: Get pending transaction
- `PendingTreasuryTxs`: Get all pending transactions

#### Statistics Queries
- `InflationMetrics`: Current inflation statistics
- `MEVStats`: MEV redistribution statistics
- `UserMEVBalance`: User's MEV balance
- `LiquidityMiningStats`: Liquidity mining statistics
- `TokenomicsStats`: Overall tokenomics dashboard

**HTTP Endpoints**: All queries include RESTful HTTP annotations under `/aura/economics/v1beta1/`

### 4. `tx.proto` - Transaction Messages

**Purpose**: Defines all transaction message types for state changes.

**Message Categories**:

#### Vesting Operations
- `CreateVestingSchedule`: Create new vesting schedule
- `ReleaseVestedTokens`: Release vested tokens to beneficiary
- `RevokeVestingSchedule`: Revoke a vesting schedule

#### Governance Operations
- `SubmitProposal`: Submit new governance proposal
- `Deposit`: Add deposit to proposal
- `Vote`: Cast vote (simple or secret ballot)
- `VoteWeighted`: Cast weighted vote
- `DelegateVote`: Delegate voting power
- `UndelegateVote`: Remove delegation
- `ExecuteProposal`: Execute passed proposal after timelock
- `RevealSecretVote`: Reveal secret ballot vote

#### Vote Lock Operations
- `LockVotingTokens`: Lock tokens for voting power boost
- `UnlockVotingTokens`: Unlock tokens after lock period

#### Treasury Operations
- `ProposeTreasurySpend`: Propose treasury spend (multisig)
- `SignTreasurySpend`: Sign treasury transaction
- `ExecuteTreasurySpend`: Execute approved treasury spend

#### Admin Operations
- `UpdateParams`: Update module parameters (governance only)
- `AdjustInflationRate`: Manually adjust inflation rate (governance only)

## Design Principles

### 1. Consolidation
Merges economicsecurity and governance into a cohesive economic management system, eliminating redundancy and improving maintainability.

### 2. Type Safety
Uses `gogoproto` annotations for:
- Custom types: `cosmossdk.io/math.Int`, `cosmossdk.io/math.LegacyDec`
- Non-nullable fields where appropriate
- Standard duration and timestamp conversions
- Coin type casting

### 3. Flexibility
Supports advanced features:
- **Governance**: Secret ballots, weighted voting, delegation, snapshot voting, emergency proposals
- **Vesting**: Multiple vesting types with cliff support
- **Treasury**: Multi-signature with timelock
- **MEV Protection**: Multiple redistribution strategies
- **Whale Protection**: Transfer limits and cooldowns

### 4. Compatibility
Follows Cosmos SDK conventions:
- Standard imports (gogoproto, google/protobuf, cosmos/base)
- Proper go_package declaration
- HTTP annotations for queries
- Message options for equality and getters

## Usage

### Generate Go Code

```bash
# From proto directory
buf generate
```

### Integration Points

The economics module integrates with:
- **Bank Module**: For transfers and balances
- **Auth Module**: For account management
- **Staking Module**: For validator voting power
- **Distribution Module**: For fee distribution
- **Gov Module**: For parameter changes (if using standard gov)

## Migration Notes

When migrating from separate `economicsecurity` and `governance` modules:

1. **State Migration**: Create migration scripts to:
   - Merge vesting schedules
   - Migrate governance proposals
   - Consolidate parameters
   - Preserve vote locks and delegations

2. **Parameter Mapping**:
   - Map old economicsecurity params to `Params.fees`, `Params.mev`, etc.
   - Map old governance params to `Params.governance`

3. **Query Compatibility**: Update clients to use new query paths:
   - Old: `/aura/economicsecurity/v1beta1/*`, `/aura/governance/v1beta1/*`
   - New: `/aura/economics/v1beta1/*`

4. **Message Handlers**: Implement handlers for all message types in the keeper

## Future Enhancements

Potential additions:
- Cross-chain governance integration
- Advanced inflation models (PID controllers)
- Automated treasury rebalancing
- DAO-specific governance features
- Quadratic funding mechanisms
- Conviction voting

## License

Copyright (c) 2024 Aura Blockchain
