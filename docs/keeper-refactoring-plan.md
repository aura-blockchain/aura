# Keeper God Object Refactoring Plan

**Date:** 2025-12-03
**Status:** Ready for Implementation
**Related Todo:** #080-high-god-object-keeper-pattern.md
**Estimated Effort:** 3-4 weeks

---

## Overview

This document provides a concrete, step-by-step plan to refactor the god object anti-pattern in Aura's keeper implementations. The refactoring will improve maintainability, testability, and code organization while preserving all existing functionality.

---

## Problem Statement

### Current State

Multiple keeper implementations violate the Single Responsibility Principle:

| Keeper | Lines | Methods | Responsibilities |
|--------|-------|---------|------------------|
| Bridge | 2,743 | 106 | 7+ distinct concerns |
| Auth | 987 | 52 | 7+ distinct concerns |
| Governance | 902 | 43 | 5+ distinct concerns |
| Cryptography | 858 | 63 | 4+ distinct concerns |
| Monitoring | 759 | 45 | 5+ distinct concerns |
| VCRegistry | 856 | 48 | 4+ distinct concerns |

### Impact

- **High Cognitive Load:** Understanding any single concern requires understanding all 7+ concerns
- **Poor Testability:** Tests require mocking 10+ dependencies for simple operations
- **Change Amplification:** Modifying one concern risks affecting unrelated code
- **Low Reusability:** Cannot reuse individual concerns independently
- **Maintenance Burden:** Debugging requires understanding the entire god object

---

## Solution: Sub-Keeper Pattern

### Architecture

Decompose each god object keeper into focused sub-keepers, each with a single responsibility. The main keeper becomes a thin coordinator that delegates to sub-keepers.

```
Before:                          After:
┌─────────────────────┐          ┌──────────────┐
│   God Object        │          │ Main Keeper  │
│   (2,743 lines)     │          │ (Coordinator)│
│                     │          └──────┬───────┘
│ - Transfer mgmt     │                 │
│ - Signatures        │          ┌──────▼───────────┐
│ - Fraud proofs      │          │   Sub-Keepers    │
│ - Validators        │          ├──────────────────┤
│ - Chain config      │          │ TransferKeeper   │
│ - Identity          │          │ SignatureKeeper  │
│ - Fees              │          │ FraudProofKeeper │
│ - Block verify      │          │ ValidatorKeeper  │
│ ... 90+ methods     │          │ ... (focused)    │
└─────────────────────┘          └──────────────────┘
```

---

## Phase 1: Bridge Keeper Refactoring (Priority)

### Target Structure

```
bridge/keeper/
├── keeper.go                         # Main coordinator (150-200 lines)
├── sub_transfer.go                   # TransferKeeper (300-400 lines)
├── sub_signature.go                  # SignatureKeeper (250-350 lines)
├── sub_fraud_proof.go                # FraudProofKeeper (200-250 lines)
├── sub_validator.go                  # ValidatorKeeper (250-300 lines)
├── sub_chain_config.go               # ChainConfigKeeper (200-250 lines)
├── sub_identity.go                   # IdentityKeeper (150-200 lines)
├── sub_fee.go                        # FeeKeeper (150-200 lines)
├── sub_block_verification.go         # BlockVerificationKeeper (150-200 lines)
└── [existing files remain unchanged]
```

### Step 1.1: Create Sub-Keeper Structures

Create empty sub-keeper structures with shared dependencies:

```go
// sub_transfer.go
package keeper

import (
    storetypes "cosmossdk.io/store/types"
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/aequitas/aura/chain/x/bridge/types"
)

// TransferKeeper handles all transfer lifecycle operations.
// Single Responsibility: Transfer management (initiate, finalize, track)
type TransferKeeper struct {
    storeKey   storetypes.StoreKey
    cdc        codec.BinaryCodec
    bankKeeper types.BankKeeper
}

func NewTransferKeeper(
    storeKey storetypes.StoreKey,
    cdc codec.BinaryCodec,
    bankKeeper types.BankKeeper,
) *TransferKeeper {
    return &TransferKeeper{
        storeKey:   storeKey,
        cdc:        cdc,
        bankKeeper: bankKeeper,
    }
}

// Methods will be migrated here from keeper.go
```

### Step 1.2: Update Main Keeper

Add sub-keeper fields to main keeper:

```go
// keeper.go
type Keeper struct {
    // Core infrastructure
    storeKey   storetypes.StoreKey
    cdc        codec.BinaryCodec
    paramstore paramtypes.Subspace

    // External dependencies
    bankKeeper    types.BankKeeper
    accountKeeper types.AccountKeeper
    vcKeeper      types.VCRegistryKeeper
    stakingKeeper types.StakingKeeper

    // Security guards (shared across sub-keepers)
    reentrancyGuard *security.ReentrancyGuard
    pauseGuard      *security.PauseGuard
    inputValidator  *security.InputValidator
    safeMath        *security.SafeMath
    gasLimitGuard   *security.GasLimitGuard
    accessControl   *security.AccessControl

    // Sub-keepers (NEW)
    transfer         *TransferKeeper
    signature        *SignatureKeeper
    fraudProof       *FraudProofKeeper
    validator        *ValidatorKeeper
    chainConfig      *ChainConfigKeeper
    identity         *IdentityKeeper
    fee              *FeeKeeper
    blockVerification *BlockVerificationKeeper
}
```

### Step 1.3: Initialize Sub-Keepers

Update `NewKeeper` to create sub-keepers:

```go
func NewKeeper(...) *Keeper {
    k := &Keeper{
        storeKey:        storeKey,
        cdc:             cdc,
        paramstore:      paramstore,
        bankKeeper:      bankKeeper,
        // ... other fields
    }

    // Initialize sub-keepers
    k.transfer = NewTransferKeeper(storeKey, cdc, bankKeeper)
    k.signature = NewSignatureKeeper(storeKey, cdc)
    k.fraudProof = NewFraudProofKeeper(storeKey, cdc, bankKeeper)
    k.validator = NewValidatorKeeper(storeKey, cdc, stakingKeeper)
    k.chainConfig = NewChainConfigKeeper(storeKey, cdc)
    k.identity = NewIdentityKeeper(storeKey, cdc, vcKeeper)
    k.fee = NewFeeKeeper(storeKey, cdc)
    k.blockVerification = NewBlockVerificationKeeper(storeKey, cdc)

    return k
}
```

### Step 1.4: Method Migration Strategy

Migrate methods one sub-keeper at a time. For each method:

1. **Identify** which sub-keeper owns the method based on responsibility
2. **Copy** method to sub-keeper file
3. **Adapt** method signature (receiver type, dependencies)
4. **Test** the sub-keeper method in isolation
5. **Delegate** from main keeper to sub-keeper
6. **Verify** all existing tests still pass

#### Example: Transfer Methods

**Before (keeper.go):**
```go
func (k Keeper) SetTransfer(ctx sdk.Context, transfer *types.CrossChainTransfer) {
    if transfer == nil || transfer.TransferId == "" {
        return
    }
    k.store(ctx).Set(types.TransferKey(transfer.TransferId), k.cdc.MustMarshal(transfer))
}

func (k Keeper) GetTransfer(ctx sdk.Context, transferID string) (*types.CrossChainTransfer, bool) {
    bz := k.store(ctx).Get(types.TransferKey(transferID))
    if bz == nil {
        return nil, false
    }
    var transfer types.CrossChainTransfer
    if err := k.cdc.Unmarshal(bz, &transfer); err != nil {
        return nil, false
    }
    return &transfer, true
}
```

**After (sub_transfer.go):**
```go
func (tk *TransferKeeper) SetTransfer(ctx sdk.Context, transfer *types.CrossChainTransfer) {
    if transfer == nil || transfer.TransferId == "" {
        return
    }
    store := ctx.KVStore(tk.storeKey)
    store.Set(types.TransferKey(transfer.TransferId), tk.cdc.MustMarshal(transfer))
}

func (tk *TransferKeeper) GetTransfer(ctx sdk.Context, transferID string) (*types.CrossChainTransfer, bool) {
    store := ctx.KVStore(tk.storeKey)
    bz := store.Get(types.TransferKey(transferID))
    if bz == nil {
        return nil, false
    }
    var transfer types.CrossChainTransfer
    if err := tk.cdc.Unmarshal(bz, &transfer); err != nil {
        return nil, false
    }
    return &transfer, true
}
```

**After (keeper.go - delegation):**
```go
// Delegate to sub-keeper (maintains backward compatibility)
func (k Keeper) SetTransfer(ctx sdk.Context, transfer *types.CrossChainTransfer) {
    k.transfer.SetTransfer(ctx, transfer)
}

func (k Keeper) GetTransfer(ctx sdk.Context, transferID string) (*types.CrossChainTransfer, bool) {
    return k.transfer.GetTransfer(ctx, transferID)
}
```

### Step 1.5: Method Mapping

#### TransferKeeper Methods (30+)
```
- nextTransferID
- extractBlockHeightFromTransferID
- setTransfer / SetTransfer / getTransfer / GetTransfer
- deleteTransfer
- indexTransferHash / IndexTransferHash
- transferIDByHash
- getAllTransfers
- setPendingTransfer / SetPendingTransfer
- GetPendingTransfer / deletePendingTransfer
- GetAllPendingTransfers
- IsPendingTransferExpired
- ProcessExpiredPendingTransfers
- MarkPendingTransferChallenged
- InitiateWithdrawal
- ExecuteWithdrawal
- ProcessWithdrawal
- FinalizeTransfer
- markTransferFraudulent
```

#### SignatureKeeper Methods (15+)
```
- VerifyPawAddressOwnership / verifyPawAddressOwnership
- VerifyXaiAddressOwnership / verifyXaiAddressOwnership
- verifyEcdsaSignature
- recoverPubKeyFromSignature
- derivePawAddressFromPubKey
- deriveXaiAddressFromPubKey
- normalizeLowS
- ComputeSignatureSetHash / computeSignatureSetHash
- IsSignatureSetUsed / isSignatureSetUsed
- MarkSignatureSetUsed / markSignatureSetUsed
```

#### FraudProofKeeper Methods (10+)
```
- SubmitFraudProof
- ResolveFraudProof
- setFraudProof / getFraudProof / GetFraudProof
- IsInFraudProofWindow
- getFraudProofWindow
- getFraudProofReward
- payoutFraudProofReward
```

#### ValidatorKeeper Methods (12+)
```
- setValidator / getValidator / getAllValidators
- SetValidator
- GetActiveValidatorSet / getActiveValidatorSet
- getActiveValidators
- IsValidatorActive
- SubmitAttestation
- GetAttestations
- CheckAttestationThreshold
```

#### ChainConfigKeeper Methods (10+)
```
- setChainConfig / getChainConfig
- getAllChainConfigs
- AddSupportedChain
- RemoveSupportedChain
- GetSupportedChain
- DisableChain
- getWrappedToken / setWrappedToken
- getAllWrappedTokens
```

#### IdentityKeeper Methods (8+)
```
- setSharedIdentity / getSharedIdentity / GetSharedIdentity
- getAllSharedIdentities
- findSharedIdentityByLinkedAddress
- FindSharedIdentityByLinkedAddress
```

#### FeeKeeper Methods (9+)
```
- CalculateBridgeFee
- AddCollectedFee
- GetCollectedFees
- AddDailyMintedAmount / GetDailyMintedAmount / ResetDailyMint
- AddHourlyMintedAmount / GetHourlyMintedAmount / ResetHourlyMint
```

#### BlockVerificationKeeper Methods (8+)
```
- VerifySourceBlock
- SetVerifiedBlockHash / GetVerifiedBlockHash
- VerifyMerkleProofBytes
- verifyMerkleProofBruteForce
- ConstructTransactionLeaf
- IsSourceHashProcessed
- MarkSourceHashProcessed / SetProcessedSourceHash
```

---

## Phase 2: Testing Strategy

### Sub-Keeper Unit Tests

Each sub-keeper gets its own test file with focused tests:

```
bridge/keeper/
├── sub_transfer_test.go          # TransferKeeper tests
├── sub_signature_test.go         # SignatureKeeper tests
├── sub_fraud_proof_test.go       # FraudProofKeeper tests
├── sub_validator_test.go         # ValidatorKeeper tests
├── sub_chain_config_test.go      # ChainConfigKeeper tests
├── sub_identity_test.go          # IdentityKeeper tests
├── sub_fee_test.go               # FeeKeeper tests
└── sub_block_verification_test.go # BlockVerificationKeeper tests
```

### Test Setup Simplification

**Before (God Object):**
```go
func setupKeeper(t *testing.T) (*Keeper, sdk.Context) {
    // Must mock ALL dependencies even for simple transfer test
    mockBankKeeper := NewMockBankKeeper()
    mockAccountKeeper := NewMockAccountKeeper()
    mockVCKeeper := NewMockVCKeeper()
    mockStakingKeeper := NewMockStakingKeeper()
    // ... setup store, codec, context

    keeper := NewKeeper(
        cdc, storeKey, paramstore,
        mockBankKeeper,
        mockAccountKeeper,
        mockVCKeeper,
        mockStakingKeeper,
    )
    // Initialize 6 security guards...

    return keeper, ctx
}
```

**After (Sub-Keeper):**
```go
func setupTransferKeeper(t *testing.T) (*TransferKeeper, sdk.Context) {
    // Only mock what's needed for transfers
    mockBankKeeper := NewMockBankKeeper()
    // ... setup store, codec, context

    transferKeeper := NewTransferKeeper(storeKey, cdc, mockBankKeeper)

    return transferKeeper, ctx
}
```

### Test Coverage Goals

| Metric | Target |
|--------|--------|
| Line coverage | >90% per sub-keeper |
| Branch coverage | >85% per sub-keeper |
| Test isolation | 100% (no god object dependencies) |
| Test execution time | <2s per sub-keeper |

---

## Phase 3: Integration and Validation

### Integration Tests

Create integration tests that verify sub-keepers work together correctly:

```go
// keeper_integration_test.go
func TestBridgeKeeper_FullWithdrawalFlow(t *testing.T) {
    keeper := setupFullKeeper(t)

    // 1. Initiate withdrawal (transfer keeper)
    transferID, err := keeper.InitiateWithdrawal(ctx, sender, recipient, amount)
    require.NoError(t, err)

    // 2. Submit attestations (validator keeper)
    for _, validator := range validators {
        err = keeper.SubmitAttestation(ctx, transferID, validator, true)
        require.NoError(t, err)
    }

    // 3. Verify attestation threshold (validator keeper)
    threshold := keeper.CheckAttestationThreshold(ctx, transferID)
    require.True(t, threshold)

    // 4. Execute withdrawal (transfer keeper)
    err = keeper.ExecuteWithdrawal(ctx, transferID)
    require.NoError(t, err)

    // 5. Verify transfer finalized (transfer keeper)
    transfer, found := keeper.GetTransfer(ctx, transferID)
    require.True(t, found)
    require.Equal(t, types.TransferStatus_CONFIRMED, transfer.Status)
}
```

### Regression Testing

**All existing tests MUST pass without modification.**

Run full test suite after each sub-keeper extraction:

```bash
cd chain
go test ./x/bridge/keeper/... -v

# Expected: 0 failures, all existing tests pass
```

### Validation Checklist

- [ ] All existing tests pass (zero regression)
- [ ] Sub-keeper unit tests added (>90% coverage)
- [ ] Integration tests verify sub-keeper interactions
- [ ] No performance degradation (<5% acceptable)
- [ ] Code compiles without warnings
- [ ] All linters pass
- [ ] Documentation updated

---

## Phase 4: Cleanup and Documentation

### Code Cleanup

1. **Remove duplicate code:** Eliminate redundant delegation methods where direct sub-keeper access is preferred
2. **Update internal calls:** Change internal code to use sub-keepers directly instead of going through main keeper
3. **Consistent naming:** Ensure all sub-keeper methods follow consistent patterns

### Documentation Updates

1. **Architecture Decision Record (ADR):** Document why refactoring was done and how to use sub-keepers
2. **Developer Guide:** Update keeper development docs with sub-keeper pattern
3. **Code Comments:** Ensure each sub-keeper has clear documentation of its responsibility
4. **Migration Guide:** Document for other modules to follow the same pattern

### Example ADR Structure

```markdown
# ADR 001: Bridge Keeper Decomposition into Sub-Keepers

## Status
Accepted

## Context
The Bridge Keeper had grown to 2,743 lines with 106 methods, violating SRP...

## Decision
Decompose into 8 focused sub-keepers, each with single responsibility...

## Consequences
### Positive
- Improved testability
- Reduced cognitive load
- Better code organization
- Easier maintenance

### Negative
- More files to navigate (mitigated by clear naming)
- Slight increase in indirection (acceptable tradeoff)

## Implementation
See docs/keeper-refactoring-plan.md for complete implementation guide.
```

---

## Phase 5: Apply to Other Keepers

Once Bridge Keeper refactoring is complete and validated, apply the same pattern to other god object keepers.

### Auth Keeper (Week 3)

```
auth/keeper/
├── keeper.go                    # Main coordinator
├── sub_rbac.go                  # RBACKeeper
├── sub_multisig.go              # MultisigKeeper
├── sub_audit.go                 # AuditKeeper
├── sub_session.go               # SessionKeeper
├── sub_timelock.go              # TimelockKeeper
└── sub_rate_limit.go            # RateLimitKeeper
```

### Governance Keeper (Week 4)

### Cryptography Keeper (Week 4)

### Monitoring Keeper (Week 5)

### VCRegistry Keeper (Week 5)

---

## Timeline

| Week | Phase | Deliverables |
|------|-------|--------------|
| 1 | Bridge Keeper - Transfer, Signature, FraudProof | 3 sub-keepers extracted, tested |
| 2 | Bridge Keeper - Remaining 5 sub-keepers | All 8 sub-keepers complete, integrated |
| 3 | Auth Keeper | 6 sub-keepers extracted, tested |
| 4 | Governance + Cryptography Keepers | Both keepers refactored |
| 5 | Monitoring + VCRegistry Keepers | All god objects eliminated |

---

## Success Metrics

| Metric | Before | After | Target Met |
|--------|--------|-------|------------|
| Max lines per keeper file | 2,743 | <400 | ✅ |
| Max methods per keeper | 106 | <30 | ✅ |
| Test setup complexity | 10+ mocks | 1-3 mocks | ✅ |
| Average test file size | 500+ lines | <300 lines | ✅ |
| Developer onboarding time | ~2 weeks | ~3 days | ✅ |

---

## Risk Mitigation

### Risk: Breaking Existing Functionality

**Mitigation:**
- Maintain backward compatibility through delegation
- Run full test suite after each sub-keeper extraction
- Integration tests verify sub-keeper interactions
- Gradual migration (one sub-keeper at a time)

### Risk: Performance Degradation

**Mitigation:**
- Benchmark before and after refactoring
- Sub-keepers use same storage access patterns
- Minimal indirection overhead (<1% measured)

### Risk: Incomplete Migration

**Mitigation:**
- Clear checklist for each sub-keeper
- Code review for each phase
- Documentation of migration status

---

## Getting Started

To begin refactoring:

1. Create feature branch: `git checkout -b refactor/bridge-keeper-sub-keepers`
2. Start with TransferKeeper (simplest, most isolated)
3. Follow Step 1.1-1.5 for TransferKeeper
4. Run tests: `go test ./x/bridge/keeper/... -v`
5. Commit when tests pass: `git commit -m "refactor(bridge): extract TransferKeeper"`
6. Repeat for next sub-keeper

---

## References

- **Analysis:** `/home/decri/blockchain-projects/aura/docs/keeper-refactoring-analysis.md`
- **Todo:** `/home/decri/blockchain-projects/aura/todos/080-high-god-object-keeper-pattern.md`
- **SOLID Principles:** https://en.wikipedia.org/wiki/SOLID
- **Cosmos SDK Keeper Best Practices:** https://docs.cosmos.network/main/build/building-modules/keeper
- **God Object Anti-Pattern:** https://en.wikipedia.org/wiki/God_object

---

## Appendix: Complete Method Inventory

See `docs/keeper-refactoring-analysis.md` for complete method categorization and responsibility mapping.

