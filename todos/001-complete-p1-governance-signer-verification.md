---
status: ready
priority: p1
issue_id: "001"
tags: [code-review, security, governance]
dependencies: []
---

# Missing Sender Authorization in Governance Message Handlers

## Problem Statement

The governance module message handlers (SubmitProposal, Deposit, Vote, DelegateVote) do NOT verify that `msg.Sender`, `msg.Proposer`, `msg.Depositor`, `msg.Voter`, or `msg.Delegator` match the authenticated transaction signer.

**Why it matters:** This allows any attacker to impersonate any address when submitting governance actions, enabling governance manipulation, vote theft, and reputation damage.

## Findings

### Evidence
- **File:** `chain/x/governance/keeper/msg_server.go`
- **Lines:** 29-88, 91-134, 137-182, 232-264

```go
// VULNERABLE CODE - Line 29-88
func (ms msgServer) SubmitProposal(goCtx context.Context, msg *govpb.MsgSubmitProposal) (*govpb.MsgSubmitProposalResponse, error) {
    // ... validation ...
    // NO CHECK: Is the tx signer actually msg.Proposer?
    proposal := &types.Proposal{
        Proposer:    msg.Proposer,  // Attacker can set ANY address here
        // ...
    }
```

### Attack Vector
1. Attacker submits a governance proposal with `proposer` set to a high-reputation validator's address
2. Attacker deposits funds with `depositor` set to the governance module's address
3. Attacker votes with `voter` set to a whale's address

### Impact
- Governance manipulation
- Unauthorized proposal submission under victim's identity
- Vote theft/impersonation
- Reputation damage

## Proposed Solutions

### Option A: Add Signer Verification to Each Handler (Recommended)
**Pros:** Direct fix, minimal changes
**Cons:** Repetitive code across handlers
**Effort:** Small (1-2 hours)
**Risk:** Low

```go
func (ms msgServer) SubmitProposal(goCtx context.Context, msg *govpb.MsgSubmitProposal) (*govpb.MsgSubmitProposalResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Extract signer from transaction context
    signers := msg.GetSigners()
    if len(signers) == 0 {
        return nil, status.Error(codes.Unauthenticated, "no signers")
    }

    // Verify proposer matches signer
    proposerAddr, err := sdk.AccAddressFromBech32(msg.Proposer)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, err.Error())
    }

    if !proposerAddr.Equals(signers[0]) {
        return nil, status.Error(codes.PermissionDenied, "proposer must be transaction signer")
    }

    // ... rest of implementation
}
```

### Option B: Use SDK's GetSigners() Pattern
**Pros:** SDK-standard approach, auto-verified by ante handler
**Cons:** Requires implementing GetSigners() on message types
**Effort:** Medium (2-4 hours)
**Risk:** Low

## Recommended Action
Implement Option A immediately. Apply same fix to: Vote, Deposit, DelegateVote, UndelegateVote, SubmitVeto, CosignVeto, ExecuteProposal handlers.

## Technical Details

### Affected Files
- `chain/x/governance/keeper/msg_server.go` (all message handlers)

### Acceptance Criteria
- [ ] All governance message handlers verify signer matches claimed address
- [ ] Unit tests verify signer impersonation is rejected
- [ ] Integration test confirms governance flow still works with valid signers

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in security review | Critical security vulnerability |

## Resources
- [Cosmos SDK Message Signing](https://docs.cosmos.network/main/build/building-modules/msg-services)
- Related: P1-002 (Bridge signature verification), P1-003 (DEX order cancellation)
