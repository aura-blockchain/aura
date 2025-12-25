# Economic Security Module

## Overview

The Economic Security module provides comprehensive tokenomics and economic protection mechanisms for the Aura blockchain, including vesting schedules, vote locking, dynamic fees, transfer taxes, liquidity mining, MEV redistribution, whale protection, inflation monitoring with circuit breakers, treasury management, and economic attack detection.

## Features

- **Vesting Schedules**: Linear, milestone-based, and cliff-then-linear vesting for team/investor/ecosystem allocations
- **Vote Locking**: Lock tokens for increased governance voting power with duration-based multipliers
- **Dynamic Fee Adjustment**: Automatic fee adjustments based on network utilization and gas price volatility
- **Transfer Taxes**: Configurable transfer taxes with exemptions and redistribution mechanisms
- **Liquidity Mining**: Automated reward distribution for liquidity providers with staking integration
- **MEV Redistribution**: Configurable MEV capture and redistribution strategies (stake-weighted, activity-weighted, equal, IR-weighted)
- **Whale Protection**: Large transaction monitoring with circuit breakers and rate limiting
- **Inflation Monitoring**: Real-time inflation tracking with alerts and automatic adjustment mechanisms
- **Circuit Breakers**: Automated protection against price volatility, liquidity crises, supply changes, and gas spikes
- **Attack Detection**: Pattern recognition for pump-and-dump, flash loan attacks, sybil attacks, wash trading, and front-running
- **Treasury Management**: Multi-signature treasury with proposal-based spending controls

## State

### Vesting
- **VestingSchedule**: Vesting with type (LINEAR, MILESTONE, CLIFF_THEN_LINEAR), beneficiary, amount, and release schedule
- **ScheduleType**: Enum - TEAM, INVESTOR, ADVISOR, ECOSYSTEM, COMMUNITY

### Vote Locking
- **VoteLock**: Locked tokens with owner, amount, lock duration, and voting power multiplier

### Tokenomics Configuration
- **TokenomicsConfig**: Global tokenomics parameters including supply caps, inflation rates, burn rates
- **DynamicFeeConfig**: Dynamic fee adjustment parameters and utilization thresholds
- **TransferTaxConfig**: Transfer tax rates, exemptions, and redistribution percentages
- **LiquidityMiningConfig**: Liquidity mining reward rates and staking requirements
- **MEVConfig**: MEV capture and redistribution strategy configuration

### Economic Protection
- **WhaleProtection**: Large transaction thresholds and rate limiting configuration
- **LargeTxRecord**: Tracking of large transactions for monitoring and analysis
- **CircuitBreakerConfig**: Circuit breaker thresholds and triggered events
- **AttackDetectionConfig**: Detected attacks with evidence and mitigation actions

### Inflation Management
- **InflationAlert**: Inflation monitoring alerts with severity levels
- **InflationAlertType**: Enum - ABOVE_TARGET, BELOW_TARGET, ABOVE_MAX, BELOW_MIN, RAPID_CHANGE
- **AlertSeverity**: Enum - INFO, WARNING, CRITICAL, EMERGENCY

### Treasury
- **TreasuryMultisig**: Multi-signature treasury configuration with signers and threshold
- **PendingTreasuryTx**: Pending treasury spend proposals with approval tracking

## Messages

### MsgCreateVestingSchedule
Create vesting schedule for beneficiary.

**Fields**: `creator`, `beneficiary`, `vesting_type`, `schedule_type`, `total_amount`, `start_time`, `cliff_duration`, `vesting_duration`

### MsgRevokeVestingSchedule
Revoke existing vesting schedule.

**Fields**: `revoker`, `schedule_id`

### MsgReleaseVestedTokens
Release vested tokens that are unlocked.

**Fields**: `beneficiary`, `schedule_id`

### MsgLockVotingTokens
Lock tokens for voting power multiplier.

**Fields**: `owner`, `amount`, `lock_duration`

### MsgUnlockVotingTokens
Unlock previously locked voting tokens.

**Fields**: `owner`, `lock_id`

### MsgUpdateParams
Update module parameters (authority only).

**Fields**: `authority`, `params`

### MsgAdjustInflationRate
Manually adjust inflation rate.

**Fields**: `authority`, `new_rate`, `reason`

### MsgProposeTreasurySpend
Propose treasury spend transaction.

**Fields**: `proposer`, `recipient`, `amount`, `description`

### MsgSignTreasurySpend
Sign pending treasury spend proposal.

**Fields**: `signer`, `tx_id`

### MsgExecuteTreasurySpend
Execute approved treasury spend.

**Fields**: `executor`, `tx_id`
