---
id: "033"
title: "Governance No Double Vote Prevention"
status: ready
priority: p1
category: security
module: governance
severity: HIGH
cvss: 8.5
source: governance-security-audit
---

# Governance No Double Vote Prevention

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

- [ ] Duplicate vote detection implemented
- [ ] Vote storage uses voter+proposalId as key (not array append)
- [ ] Decide: allow vote updates or reject duplicates
- [ ] Signer verification added
- [ ] Tests for duplicate vote handling
- [ ] Tests for vote update (if allowed)
