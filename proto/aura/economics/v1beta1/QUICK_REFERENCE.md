# Economics Module Quick Reference

## Transactions

### Vesting
```bash
# Create vesting schedule
aurad tx economics create-vesting-schedule [beneficiary] [amount] [start-time] [cliff-duration] [vesting-duration] [type] [schedule-type]

# Release vested tokens
aurad tx economics release-vested-tokens [schedule-id]

# Revoke vesting schedule
aurad tx economics revoke-vesting-schedule [schedule-id] [reason]
```

### Governance
```bash
# Submit proposal
aurad tx economics submit-proposal [title] [description] [category] --deposit=[amount] --emergency=[bool]

# Deposit on proposal
aurad tx economics deposit [proposal-id] [amount]

# Vote on proposal
aurad tx economics vote [proposal-id] [yes|no|abstain|no-with-veto] --secret=[bool]

# Vote with weights
aurad tx economics vote-weighted [proposal-id] [yes=0.5,no=0.3,abstain=0.2]

# Delegate voting power
aurad tx economics delegate-vote [delegate-address] --categories=[text,parameter-change]

# Undelegate voting power
aurad tx economics undelegate-vote [delegate-address]

# Execute passed proposal
aurad tx economics execute-proposal [proposal-id]

# Reveal secret vote
aurad tx economics reveal-secret-vote [proposal-id] [option] [reveal-key]
```

### Vote Locks
```bash
# Lock tokens for voting boost
aurad tx economics lock-voting-tokens [amount] [duration-seconds]

# Unlock tokens
aurad tx economics unlock-voting-tokens [lock-id]
```

### Treasury
```bash
# Propose treasury spend
aurad tx economics propose-treasury-spend [recipient] [amount] [description]

# Sign treasury spend
aurad tx economics sign-treasury-spend [tx-id]

# Execute treasury spend
aurad tx economics execute-treasury-spend [tx-id]
```

### Admin
```bash
# Update params (governance only)
aurad tx economics update-params [params-json]

# Adjust inflation rate (governance only)
aurad tx economics adjust-inflation-rate [new-rate-bp] [reason]
```

## Queries

### Parameters
```bash
aurad query economics params
```

### Vesting
```bash
aurad query economics vesting-schedule [schedule-id]
aurad query economics vesting-schedules-by-address [address]
aurad query economics all-vesting-schedules
```

### Governance
```bash
aurad query economics proposal [proposal-id]
aurad query economics proposals --status=[deposit-period|voting-period|passed|rejected]
aurad query economics vote [proposal-id] [voter]
aurad query economics votes [proposal-id]
aurad query economics deposit [proposal-id] [depositor]
aurad query economics deposits [proposal-id]
aurad query economics tally [proposal-id]
```

### Vote Locks
```bash
aurad query economics vote-lock [lock-id]
aurad query economics vote-locks-by-owner [owner]
aurad query economics voting-power [address]
aurad query economics vote-delegations [delegator]
```

### Treasury
```bash
aurad query economics pending-treasury-tx [tx-id]
aurad query economics pending-treasury-txs
```

### Statistics
```bash
aurad query economics inflation-metrics
aurad query economics mev-stats
aurad query economics user-mev-balance [address]
aurad query economics liquidity-mining-stats
aurad query economics tokenomics-stats
```

## HTTP REST API

### Vesting
```
GET /aura/economics/v1beta1/vesting/{schedule_id}
GET /aura/economics/v1beta1/vesting/address/{address}
GET /aura/economics/v1beta1/vesting
```

### Governance
```
GET /aura/economics/v1beta1/proposals/{proposal_id}
GET /aura/economics/v1beta1/proposals
GET /aura/economics/v1beta1/proposals/{proposal_id}/votes/{voter}
GET /aura/economics/v1beta1/proposals/{proposal_id}/votes
GET /aura/economics/v1beta1/proposals/{proposal_id}/deposits/{depositor}
GET /aura/economics/v1beta1/proposals/{proposal_id}/deposits
GET /aura/economics/v1beta1/proposals/{proposal_id}/tally
```

### Vote Locks
```
GET /aura/economics/v1beta1/vote_locks/{lock_id}
GET /aura/economics/v1beta1/vote_locks/owner/{owner}
GET /aura/economics/v1beta1/voting_power/{address}
GET /aura/economics/v1beta1/vote_delegations/{delegator}
```

### Treasury
```
GET /aura/economics/v1beta1/treasury/pending/{tx_id}
GET /aura/economics/v1beta1/treasury/pending
```

### Statistics
```
GET /aura/economics/v1beta1/inflation/metrics
GET /aura/economics/v1beta1/mev/stats
GET /aura/economics/v1beta1/mev/balance/{address}
GET /aura/economics/v1beta1/liquidity_mining/stats
GET /aura/economics/v1beta1/tokenomics/stats
```

## Message Examples

### Submit Proposal (JSON)
```json
{
  "title": "Adjust Fee Parameters",
  "description": "Increase base fee to reduce spam",
  "category": "PROPOSAL_CATEGORY_PARAMETER_CHANGE",
  "proposer": "aura1...",
  "initial_deposit": [
    {
      "denom": "uaura",
      "amount": "10000000"
    }
  ],
  "is_emergency": false
}
```

### Create Vesting Schedule (JSON)
```json
{
  "creator": "aura1...",
  "beneficiary_address": "aura1...",
  "total_amount": {
    "denom": "uaura",
    "amount": "1000000000"
  },
  "start_time": "2025-01-01T00:00:00Z",
  "cliff_duration": 31536000,
  "vesting_duration": 126144000,
  "vesting_type": "VESTING_TYPE_LINEAR",
  "schedule_type": "SCHEDULE_TYPE_TEAM"
}
```

### Lock Voting Tokens (JSON)
```json
{
  "owner": "aura1...",
  "amount": {
    "denom": "uaura",
    "amount": "100000000"
  },
  "lock_duration": 31536000
}
```

## Enums

### VestingType
- `VESTING_TYPE_UNSPECIFIED` (0)
- `VESTING_TYPE_LINEAR` (1)
- `VESTING_TYPE_CLIFF` (2)
- `VESTING_TYPE_GRADED` (3)
- `VESTING_TYPE_MILESTONE` (4)

### ScheduleType
- `SCHEDULE_TYPE_UNSPECIFIED` (0)
- `SCHEDULE_TYPE_TEAM` (1)
- `SCHEDULE_TYPE_INVESTOR` (2)
- `SCHEDULE_TYPE_ADVISOR` (3)
- `SCHEDULE_TYPE_ECOSYSTEM` (4)
- `SCHEDULE_TYPE_COMMUNITY` (5)

### ProposalStatus
- `PROPOSAL_STATUS_UNSPECIFIED` (0)
- `PROPOSAL_STATUS_DEPOSIT_PERIOD` (1)
- `PROPOSAL_STATUS_VOTING_PERIOD` (2)
- `PROPOSAL_STATUS_PASSED` (3)
- `PROPOSAL_STATUS_REJECTED` (4)
- `PROPOSAL_STATUS_FAILED` (5)
- `PROPOSAL_STATUS_VETOED` (6)
- `PROPOSAL_STATUS_EXECUTION_DELAY` (7)
- `PROPOSAL_STATUS_EXECUTED` (8)

### ProposalCategory
- `PROPOSAL_CATEGORY_UNSPECIFIED` (0)
- `PROPOSAL_CATEGORY_TEXT` (1)
- `PROPOSAL_CATEGORY_PARAMETER_CHANGE` (2)
- `PROPOSAL_CATEGORY_SOFTWARE_UPGRADE` (3)
- `PROPOSAL_CATEGORY_SPENDING` (4)
- `PROPOSAL_CATEGORY_EMERGENCY` (5)
- `PROPOSAL_CATEGORY_CONSTITUTION` (6)

### VoteOption
- `VOTE_OPTION_UNSPECIFIED` (0)
- `VOTE_OPTION_YES` (1)
- `VOTE_OPTION_ABSTAIN` (2)
- `VOTE_OPTION_NO` (3)
- `VOTE_OPTION_NO_WITH_VETO` (4)

### MEVRedistributionStrategy
- `MEV_STRATEGY_UNSPECIFIED` (0)
- `MEV_STRATEGY_PROPORTIONAL_TO_STAKE` (1)
- `MEV_STRATEGY_PROPORTIONAL_TO_ACTIVITY` (2)
- `MEV_STRATEGY_EQUAL_DISTRIBUTION` (3)
- `MEV_STRATEGY_IR_WEIGHTED` (4)

## Common Workflows

### Create and Vote on Proposal
```bash
# 1. Submit proposal
aurad tx economics submit-proposal \
  "Increase Block Gas Limit" \
  "Proposal to increase max block gas from 30M to 50M" \
  parameter-change \
  --deposit=10000000uaura \
  --from=proposer

# 2. Add deposit (if needed)
aurad tx economics deposit 1 5000000uaura --from=depositor

# 3. Vote
aurad tx economics vote 1 yes --from=voter

# 4. Check tally
aurad query economics tally 1

# 5. Execute (after timelock)
aurad tx economics execute-proposal 1 --from=executor
```

### Setup Vesting for Team Member
```bash
# 1. Create vesting schedule
# 4-year linear vesting with 1-year cliff
aurad tx economics create-vesting-schedule \
  aura1teamaddress... \
  1000000000uaura \
  2025-01-01T00:00:00Z \
  31536000 \
  126144000 \
  linear \
  team \
  --from=admin

# 2. Query schedule
aurad query economics vesting-schedules-by-address aura1teamaddress...

# 3. Release vested tokens (beneficiary)
aurad tx economics release-vested-tokens schedule-1 --from=aura1teamaddress...
```

### Lock Tokens for Governance
```bash
# 1. Lock tokens for 1 year
aurad tx economics lock-voting-tokens 100000000uaura 31536000 --from=voter

# 2. Check voting power
aurad query economics voting-power aura1voter...

# 3. Vote with boosted power
aurad tx economics vote 5 yes --from=voter

# 4. Unlock after period
aurad query economics vote-locks-by-owner aura1voter...
aurad tx economics unlock-voting-tokens lock-123 --from=voter
```

### Treasury Multi-Sig Spend
```bash
# 1. Propose spend
aurad tx economics propose-treasury-spend \
  aura1recipient... \
  50000000uaura \
  "Q1 Development Grant" \
  --from=signer1

# 2. Sign by other signers
aurad tx economics sign-treasury-spend tx-456 --from=signer2
aurad tx economics sign-treasury-spend tx-456 --from=signer3

# 3. Check status
aurad query economics pending-treasury-tx tx-456

# 4. Execute (after timelock and signatures)
aurad tx economics execute-treasury-spend tx-456 --from=executor
```

## Parameter Examples

### Fee Parameters
```json
{
  "base_fee": "1000",
  "min_gas_price": "0.025",
  "dynamic_fees_enabled": true,
  "fee_burn_percentage": 5000,
  "min_fee_multiplier": 5000,
  "max_fee_multiplier": 20000,
  "target_block_utilization": 7000,
  "fee_adjustment_speed": 1000
}
```

### Governance Parameters
```json
{
  "min_deposit": [{"denom": "uaura", "amount": "10000000"}],
  "max_deposit_period": "604800s",
  "voting_period": "604800s",
  "quorum": 3340,
  "threshold": 5000,
  "veto_threshold": 3340,
  "execution_delay": "172800s",
  "emergency_voting_period": "86400s",
  "emergency_quorum": 5000,
  "emergency_threshold": 6670,
  "quadratic_voting_enabled": false,
  "vote_locking_enabled": true,
  "min_lock_duration": "2592000s",
  "max_lock_duration": "126144000s",
  "lock_multiplier_per_year": 10000
}
```

## Error Handling

### Common Errors
- `proposal not found`: Proposal ID doesn't exist
- `voting period ended`: Can't vote after period ends
- `insufficient deposit`: Deposit below minimum
- `unauthorized`: Not authorized for operation
- `schedule not found`: Vesting schedule doesn't exist
- `vesting not ready`: Tokens not yet vested
- `lock still active`: Can't unlock before period ends
- `insufficient signatures`: Treasury tx needs more signatures
- `timelock not expired`: Treasury tx in timelock period

## Best Practices

1. **Proposals**: Always include detailed description and rationale
2. **Vesting**: Use appropriate schedule type for transparency
3. **Vote Locks**: Lock longer for more voting power
4. **Treasury**: Allow sufficient timelock for review
5. **Parameters**: Test changes on testnet first
6. **Queries**: Use pagination for large result sets
7. **Gas**: Estimate gas before submitting transactions
