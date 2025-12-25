# Governance Module

## Overview

The Governance module provides advanced on-chain governance capabilities for the Aura blockchain, including proposal submission with categories, voting with delegation, secret ballot voting, quadratic voting, emergency veto mechanisms, execution delays, token locks for voting power, and snapshot-based governance.

## Features

- **Proposal Categories**: Text, parameter change, software upgrade, spending, emergency, constitution proposals
- **Voting Methods**: Standard voting, weighted voting, secret ballot (commit-reveal), quadratic voting
- **Vote Delegation**: Delegate voting power to trusted addresses with revocation capability
- **Emergency Veto**: Multi-signature veto mechanism for emergency governance actions
- **Execution Delay**: Mandatory delay between proposal passage and execution
- **Token Locking**: Lock tokens for increased voting power with configurable duration
- **Snapshot Voting**: Off-chain vote aggregation with on-chain verification
- **Tally Mechanisms**: Standard tallying and quadratic voting power calculation

## State

### Proposals
- **Proposal**: Core proposal with category, title, description, proposer, and voting status
- **ProposalStatus**: Enum - DEPOSIT_PERIOD, VOTING_PERIOD, PASSED, REJECTED, FAILED, VETOED, EXECUTION_DELAY, READY_FOR_EXECUTION, EXECUTED
- **ProposalCategory**: Enum - TEXT, PARAMETER_CHANGE, SOFTWARE_UPGRADE, SPENDING, EMERGENCY, CONSTITUTION
- **TallyResult**: Vote tallying with yes/no/abstain/veto counts

### Voting
- **Vote**: Individual vote with option (YES, NO, ABSTAIN, NO_WITH_VETO)
- **WeightedVoteOption**: Weighted voting across multiple options
- **VoteOption**: Enum - YES, ABSTAIN, NO, NO_WITH_VETO
- **VoteDelegation**: Delegation from delegator to delegate address

### Secret Ballot
- **SnapshotVote**: Off-chain vote submission with commitment hash
- **VoteCommitment**: Commit-reveal vote with hash and reveal status

### Emergency Controls
- **VetoRequest**: Multi-signature veto request with cosigner tracking

### Token Mechanics
- **Deposit**: Proposal deposit tracking per depositor
- **TokenLock**: Locked tokens for voting power with unlock time

### Parameters
- **GovernanceParams**: Module configuration including voting periods, quorums, thresholds
- **CategoryParams**: Per-category parameter overrides

## Messages

### MsgSubmitProposal
Submit governance proposal with deposit.

**Fields**: `proposer`, `category`, `title`, `description`, `initial_deposit`

### MsgVote
Vote on proposal with single option.

**Fields**: `proposal_id`, `voter`, `option`

### MsgVoteWeighted
Vote on proposal with weighted options.

**Fields**: `proposal_id`, `voter`, `options`

### MsgDeposit
Deposit tokens to proposal.

**Fields**: `proposal_id`, `depositor`, `amount`

### MsgSubmitSnapshotVote
Submit off-chain snapshot vote.

**Fields**: `proposal_id`, `voter`, `vote_hash`, `signature`

### MsgRevealSecretVote
Reveal secret ballot vote.

**Fields**: `proposal_id`, `voter`, `option`, `salt`

### MsgDelegateVote
Delegate voting power to another address.

**Fields**: `delegator`, `delegate`

### MsgUndelegateVote
Undelegate voting power.

**Fields**: `delegator`, `delegate`

### MsgSubmitVeto
Initiate emergency veto request.

**Fields**: `proposal_id`, `vetoer`, `reason`

### MsgCosignVeto
Cosign emergency veto.

**Fields**: `veto_id`, `cosigner`

### MsgExecuteProposal
Execute passed proposal after delay.

**Fields**: `proposal_id`, `executor`
