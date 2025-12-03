---
id: "045"
title: "Governance Delegations Not Exported in Genesis"
status: ready
priority: p1
category: data-integrity
module: governance
severity: CRITICAL
data_loss_risk: 100%
source: data-integrity-review
---

# Governance Delegations Not Exported in Genesis

## Problem

The `ExportGenesis` function only exports `Params`. All vote delegations, proposals, votes, and deposits are lost on every chain upgrade.

## Affected Files

- `chain/x/governance/keeper/genesis.go`

## Vulnerability

```go
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
    params := k.GetParams(ctx)
    return types.GenesisState{Params: params}  // ONLY PARAMS!
    // MISSING: Delegations
    // MISSING: Proposals
    // MISSING: Votes
    // MISSING: Deposits
}
```

## Impact

- **100% data loss** of vote delegations on chain upgrade
- All active proposals lost
- All votes lost
- All deposits permanently locked (never refunded)
- Governance state completely wiped

## Required Fix

```go
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
    params := k.GetParams(ctx)

    // Export all vote delegations
    delegations := k.getAllVoteDelegations(ctx)

    // Export all proposals
    proposals := k.GetAllProposals(ctx)

    // Export all votes
    votes := make(map[uint64][]*types.Vote)
    for _, proposal := range proposals {
        proposalVotes := k.GetProposalVotes(ctx, proposal.Id)
        votes[proposal.Id] = proposalVotes
    }

    // Export all deposits
    deposits := make(map[uint64][]*types.Deposit)
    for _, proposal := range proposals {
        proposalDeposits := k.GetProposalDeposits(ctx, proposal.Id)
        deposits[proposal.Id] = proposalDeposits
    }

    return types.GenesisState{
        Params:      params,
        Delegations: delegations,
        Proposals:   proposals,
        Votes:       votes,
        Deposits:    deposits,
    }
}

func (k *Keeper) getAllVoteDelegations(ctx sdk.Context) []*types.VoteDelegation {
    store := ctx.KVStore(k.storeKey)
    iterator := storetypes.KVStorePrefixIterator(store, DelegationsKeyPrefix)
    defer iterator.Close()

    var delegations []*types.VoteDelegation
    for ; iterator.Valid(); iterator.Next() {
        var delegation types.VoteDelegation
        if err := k.cdc.Unmarshal(iterator.Value(), &delegation); err != nil {
            ctx.Logger().Error("failed to unmarshal delegation during export", "error", err)
            continue
        }
        delegations = append(delegations, &delegation)
    }
    return delegations
}

func (k Keeper) InitGenesis(ctx sdk.Context, gen types.GenesisState) error {
    // Set params
    k.SetParams(ctx, gen.Params)

    // Import delegations
    for _, delegation := range gen.Delegations {
        if delegation == nil {
            continue
        }
        if err := k.SetVoteDelegation(ctx, delegation); err != nil {
            return fmt.Errorf("failed to import delegation: %w", err)
        }
    }

    // Import proposals
    for _, proposal := range gen.Proposals {
        if proposal == nil {
            continue
        }
        if err := k.SetProposal(ctx, *proposal); err != nil {
            return fmt.Errorf("failed to import proposal: %w", err)
        }
    }

    // Import votes
    for proposalId, proposalVotes := range gen.Votes {
        for _, vote := range proposalVotes {
            if vote == nil {
                continue
            }
            if err := k.SetVote(ctx, proposalId, vote); err != nil {
                return fmt.Errorf("failed to import vote: %w", err)
            }
        }
    }

    // Import deposits
    for proposalId, proposalDeposits := range gen.Deposits {
        for _, deposit := range proposalDeposits {
            if deposit == nil {
                continue
            }
            if err := k.SetDeposit(ctx, proposalId, deposit); err != nil {
                return fmt.Errorf("failed to import deposit: %w", err)
            }
        }
    }

    return nil
}
```

## Update GenesisState Proto

```protobuf
message GenesisState {
    Params params = 1;
    repeated VoteDelegation delegations = 2;
    repeated Proposal proposals = 3;
    map<uint64, VoteList> votes = 4;
    map<uint64, DepositList> deposits = 5;
}

message VoteList {
    repeated Vote votes = 1;
}

message DepositList {
    repeated Deposit deposits = 1;
}
```

## Acceptance Criteria

- [ ] All delegations exported
- [ ] All proposals exported
- [ ] All votes exported
- [ ] All deposits exported
- [ ] InitGenesis imports all data
- [ ] Genesis round-trip test (export → import → export identical)
- [ ] Upgrade simulation test
