---
id: "031"
title: "Governance No Deposit Locking"
status: ready
priority: p1
category: security
module: governance
severity: CRITICAL
cvss: 9.5
source: governance-security-audit
---

# Governance No Deposit Locking

## Problem

The `SubmitProposal` function stores deposit as a string without actually transferring tokens from the proposer. Anyone can create proposals with zero cost.

## Affected Files

- `chain/x/governance/keeper/msg_server.go:79-90`

## Vulnerability

```go
func (ms msgServer) SubmitProposal(goCtx context.Context, msg *govpb.MsgSubmitProposal) (*govpb.MsgSubmitProposalResponse, error) {
    // ...
    if msg.InitialDeposit != "" && msg.InitialDeposit != "0" {
        // NO ACTUAL TOKEN TRANSFER - just stores a string!
        deposit := &types.Deposit{
            ProposalId: id,
            Depositor:  msg.Creator,
            Amount:     msg.InitialDeposit,  // String stored, not coins transferred
        }
        if err := k.SetDeposit(ctx, id, deposit); err != nil {
            return nil, status.Error(codes.Internal, "failed to set deposit")
        }
    }
    // ...
}
```

## Impact

- Spam governance with unlimited proposals
- No economic cost to attack governance
- Governance DOS attacks trivial
- No deposit refund mechanism needed (never locked)

## Required Fix

```go
func (ms msgServer) SubmitProposal(goCtx context.Context, msg *govpb.MsgSubmitProposal) (*govpb.MsgSubmitProposalResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Parse deposit amount
    deposit, err := sdk.ParseCoinsNormalized(msg.InitialDeposit)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid deposit amount")
    }

    // Check minimum deposit
    params := ms.Keeper.GetParams(ctx)
    minDeposit, _ := sdk.ParseCoinsNormalized(params.MinDeposit)
    if deposit.IsAllLT(minDeposit) {
        return nil, status.Errorf(codes.InvalidArgument,
            "deposit %s below minimum %s", deposit, minDeposit)
    }

    // Get proposer address
    proposerAddr, err := sdk.AccAddressFromBech32(msg.Creator)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "invalid creator address")
    }

    // ACTUALLY TRANSFER TOKENS TO MODULE ACCOUNT
    err = ms.Keeper.bankKeeper.SendCoinsFromAccountToModule(
        ctx,
        proposerAddr,
        types.ModuleName,
        deposit,
    )
    if err != nil {
        return nil, status.Errorf(codes.FailedPrecondition,
            "failed to transfer deposit: %s", err)
    }

    // Create proposal
    proposal := &types.Proposal{
        Id:            id,
        Creator:       msg.Creator,
        Title:         msg.Title,
        Description:   msg.Description,
        Status:        types.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
        TotalDeposit:  deposit.String(),
        // ...
    }

    // Store deposit record
    depositRecord := &types.Deposit{
        ProposalId: id,
        Depositor:  msg.Creator,
        Amount:     deposit.String(),
    }
    if err := ms.Keeper.SetDeposit(ctx, id, depositRecord); err != nil {
        // Refund on failure
        ms.Keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, proposerAddr, deposit)
        return nil, status.Error(codes.Internal, "failed to set deposit")
    }

    // Emit event
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            types.EventTypeSubmitProposal,
            sdk.NewAttribute(types.AttributeKeyProposalId, fmt.Sprintf("%d", id)),
            sdk.NewAttribute(types.AttributeKeyDepositor, msg.Creator),
            sdk.NewAttribute(types.AttributeKeyDeposit, deposit.String()),
        ),
    )

    return &govpb.MsgSubmitProposalResponse{ProposalId: id}, nil
}
```

## Acceptance Criteria

- [ ] Deposits actually transferred to module account
- [ ] Minimum deposit enforced
- [ ] Deposit refund on proposal rejection
- [ ] Deposit slash on spam/malicious proposals
- [ ] Tests for deposit transfer
- [ ] Tests for minimum deposit rejection
- [ ] Tests for deposit refund
