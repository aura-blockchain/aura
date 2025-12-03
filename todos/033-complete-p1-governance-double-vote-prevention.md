---
id: "033"
title: "Governance Double Vote Prevention"
status: complete
priority: p1
category: security
module: governance
severity: HIGH
cvss: 8.5
source: governance-security-audit
completed: 2025-12-03
---

# Governance Double Vote Prevention - VERIFIED SECURE

## Problem

Users can vote multiple times on the same proposal, with all votes counted in the tally.

## Affected Files

- `chain/x/governance/keeper/msg_server.go:158-180`
- `chain/x/governance/keeper/keeper.go:307-340`

## Vulnerability

```go
func (ms msgServer) Vote(goCtx context.Context, msg *govpb.MsgVote) (*govpb.MsgVoteResponse, error) {
    // ...
    vote := &types.Vote{
        ProposalId: msg.ProposalId,
        Voter:      msg.Voter,
        Option:     int32(msg.Option),
        Timestamp:  &timestamp,
    }

    // APPENDS new vote without checking if voter already voted
    if err := ms.Keeper.AddVote(ctx, msg.ProposalId, vote); err != nil {
        return nil, status.Error(codes.Internal, "failed to add vote")
    }
}

func (k *Keeper) AddVote(ctx sdk.Context, proposalId uint64, vote *types.Vote) error {
    proposal, err := k.GetProposal(ctx, proposalId)
    if err != nil {
        return err
    }
    // JUST APPENDS - no duplicate check
    proposal.Votes = append(proposal.Votes, vote)
    return k.SetProposal(ctx, *proposal)
}
```

## Impact

- Vote manipulation by repeated voting
- Governance outcome manipulation
- Unfair voting advantage

## Required Fix

```go
func (ms msgServer) Vote(goCtx context.Context, msg *govpb.MsgVote) (*govpb.MsgVoteResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Verify signer matches voter
    signers := msg.GetSigners()
    if len(signers) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no signers")
    }

    voterAddr, err := sdk.AccAddressFromBech32(msg.Voter)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid voter address")
    }

    if !signers[0].Equals(voterAddr) {
        return nil, status.Error(codes.PermissionDenied, "signer does not match voter")
    }

    // Check proposal exists and is in voting period
    proposal, err := ms.Keeper.GetProposal(ctx, msg.ProposalId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "proposal not found")
    }

    if proposal.Status != types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
        return nil, status.Errorf(codes.FailedPrecondition,
            "proposal not in voting period, status: %s", proposal.Status)
    }

    // CHECK FOR EXISTING VOTE
    existingVote, err := ms.Keeper.GetVote(ctx, msg.ProposalId, msg.Voter)
    if err == nil && existingVote != nil {
        // User already voted - update vote or reject
        // Option 1: Allow vote change (more flexible)
        existingVote.Option = int32(msg.Option)
        existingVote.Timestamp = timestamppb.New(ctx.BlockTime())
        if err := ms.Keeper.SetVote(ctx, msg.ProposalId, existingVote); err != nil {
            return nil, status.Error(codes.Internal, "failed to update vote")
        }

        ctx.EventManager().EmitEvent(
            sdk.NewEvent(
                types.EventTypeVoteUpdated,
                sdk.NewAttribute(types.AttributeKeyProposalId, fmt.Sprintf("%d", msg.ProposalId)),
                sdk.NewAttribute(types.AttributeKeyVoter, msg.Voter),
                sdk.NewAttribute(types.AttributeKeyOption, msg.Option.String()),
            ),
        )

        return &govpb.MsgVoteResponse{}, nil

        // Option 2: Reject duplicate (stricter)
        // return nil, status.Error(codes.AlreadyExists, "voter has already voted on this proposal")
    }

    // Create new vote
    vote := &types.Vote{
        ProposalId: msg.ProposalId,
        Voter:      msg.Voter,
        Option:     int32(msg.Option),
        Timestamp:  timestamppb.New(ctx.BlockTime()),
    }

    if err := ms.Keeper.SetVote(ctx, msg.ProposalId, vote); err != nil {
        return nil, status.Error(codes.Internal, "failed to add vote")
    }

    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeVote,
            sdk.NewAttribute(types.AttributeKeyProposalId, fmt.Sprintf("%d", msg.ProposalId)),
            sdk.NewAttribute(types.AttributeKeyVoter, msg.Voter),
            sdk.NewAttribute(types.AttributeKeyOption, msg.Option.String()),
        ),
    )

    return &govpb.MsgVoteResponse{}, nil
}
```

## Acceptance Criteria

- [x] Duplicate vote detection implemented
- [x] Vote storage uses voter+proposalId as key (not array append)
- [x] Decision made: allow vote updates (standard governance behavior)
- [x] Signer verification added
- [x] Tests for duplicate vote handling
- [x] Tests for vote update behavior

## Implementation Status

### Already Implemented (Verified 2025-12-03)

The governance module already has complete double vote prevention:

1. **Map-Based Storage** (`keeper.go:194-212`)
   - Votes are stored using composite key: `VotesKeyPrefix + proposalID + voter`
   - This key structure inherently prevents duplicate votes
   - `SetVote()` overwrites existing votes instead of appending

2. **Vote Update Logic** (`msg_server.go:246-274`)
   - Checks for existing vote before creating new one
   - Allows vote updates during voting period (standard governance feature)
   - Updates timestamp when vote is changed

3. **Signer Verification** (`msg_server.go:218-230`)
   - Validates that signer matches voter address
   - Rejects votes with no signers
   - Prevents voting on behalf of others

4. **Voting Period Check** (`msg_server.go:241-243`)
   - Only allows votes during `VOTING_PERIOD` status
   - Prevents votes before/after voting window

### Comprehensive Test Coverage

New test file: `chain/x/governance/keeper/double_vote_verified_test.go`

Tests verify:
- Duplicate vote rejection
- Vote update behavior
- Multi-voter scenarios
- Storage key uniqueness across proposals
- Secret ballot protection
- Vote count accuracy in tally
- Map-based storage mechanism
- Signer verification (documented)
- Timestamp preservation (documented)

All 10 tests pass successfully.

### Test Results

```
=== RUN   TestDoubleVote_PreventsDuplicateVoting
--- PASS: TestDoubleVote_PreventsDuplicateVoting (0.00s)
=== RUN   TestDoubleVote_MultipleVotersOnSameProposal
--- PASS: TestDoubleVote_MultipleVotersOnSameProposal (0.00s)
=== RUN   TestDoubleVote_VoteStorageKeyUniqueness
--- PASS: TestDoubleVote_VoteStorageKeyUniqueness (0.00s)
=== RUN   TestDoubleVote_AttemptMultipleVotesOnSameProposal
--- PASS: TestDoubleVote_AttemptMultipleVotesOnSameProposal (0.00s)
=== RUN   TestDoubleVote_SecretBallotUpdate
--- PASS: TestDoubleVote_SecretBallotUpdate (0.00s)
=== RUN   TestDoubleVote_VoteCountAccuracy
--- PASS: TestDoubleVote_VoteCountAccuracy (0.00s)
=== RUN   TestDoubleVote_ProposalNotFound
--- PASS: TestDoubleVote_ProposalNotFound (0.00s)
=== RUN   TestDoubleVote_EmptyVoter
--- PASS: TestDoubleVote_EmptyVoter (0.00s)
=== RUN   TestDoubleVote_MapBasedStorage
--- PASS: TestDoubleVote_MapBasedStorage (0.00s)
=== RUN   TestDoubleVote_TimestampPreservation
--- PASS: TestDoubleVote_TimestampPreservation (0.00s)
PASS
ok      github.com/aequitas/aura/chain/x/governance/keeper      0.122s
```

## Security Assessment

**SECURE** - The implementation follows blockchain security best practices:

- Uses composite keys for vote storage (ProposalID + Voter)
- Prevents vote duplication at storage level
- Allows vote updates (standard governance feature, not a vulnerability)
- Verifies signers to prevent unauthorized voting
- Enforces voting period restrictions

This is the standard pattern used in Cosmos SDK governance and other production blockchain governance systems.
