# TODO: Implement 21 missing bridge unlock security tests

---
status: pending
priority: p1
issue_id: "006"
tags: [code-review, security, testing, bridge, critical]
dependencies: []
---

## Problem Statement

The bridge module has 21 security test functions marked with `// TODO: Implement test`. These tests are critical for verifying bridge security against replay attacks, signature manipulation, and validator consensus bypass.

**Impact:** Untested security-critical code paths. Unknown vulnerabilities in cross-chain asset transfers.

## Findings

**Location:** `/home/decri/blockchain-projects/aura/chain/x/bridge/keeper/msg_server_unlock_security_test.go`

**Missing Tests (21 total):**
```go
func TestUnlock_ReplayAttack(t *testing.T) { /* TODO: Implement test */ }
func TestUnlock_SignatureReuse(t *testing.T) { /* TODO: Implement test */ }
func TestUnlock_InvalidValidatorSet(t *testing.T) { /* TODO: Implement test */ }
func TestUnlock_ThresholdManipulation(t *testing.T) { /* TODO: Implement test */ }
func TestUnlock_DoubleSigning(t *testing.T) { /* TODO: Implement test */ }
// ... 16 more
```

**Attack Vectors Not Tested:**
1. Replay attacks (reuse of burn transaction hashes)
2. Signature replay across validator set changes
3. Front-running unlock requests
4. Time-of-check-time-of-use vulnerabilities
5. Validator set manipulation during unlock

## Proposed Solutions

### Option 1: Implement all 21 tests systematically (Recommended)
**Pros:** Complete security coverage
**Cons:** Time intensive
**Effort:** Large (3-5 days)
**Risk:** Low

### Option 2: Implement critical subset (6 tests)
**Pros:** Faster, covers highest risk
**Cons:** Incomplete coverage
**Effort:** Medium (1-2 days)
**Risk:** Medium

**Critical 6:**
1. TestUnlock_ReplayAttack
2. TestUnlock_SignatureReuse
3. TestUnlock_InvalidValidatorSet
4. TestUnlock_ThresholdManipulation
5. TestUnlock_ValidatorSetChangeDuring
6. TestUnlock_ExpiredSignatures

## Recommended Action

Option 1 - Implement all 21 tests before mainnet. Option 2 acceptable for testnet.

## Technical Details

**Test Template:**
```go
func TestUnlock_ReplayAttack(t *testing.T) {
    suite := setupBridgeTestSuite(t)

    // 1. Create valid unlock with signatures
    unlock := createValidUnlock(suite)

    // 2. Process unlock successfully
    _, err := suite.msgServer.Unlock(suite.ctx, unlock)
    require.NoError(t, err)

    // 3. Attempt replay with same signatures
    _, err = suite.msgServer.Unlock(suite.ctx, unlock)
    require.ErrorIs(t, err, types.ErrAlreadyProcessed)

    // 4. Verify funds not double-released
    balance := suite.bankKeeper.GetBalance(...)
    require.Equal(t, expectedBalance, balance)
}
```

## Acceptance Criteria

- [ ] All 21 test functions implemented
- [ ] Tests cover attack scenarios listed above
- [ ] Tests verify both rejection AND state not corrupted
- [ ] 100% pass rate on bridge security tests
- [ ] Documentation of each attack scenario

## Work Log

| Date | Action | Notes |
|------|--------|-------|
| 2025-12-08 | Identified | Security Sentinel agent review |

## Resources

- Test file: `chain/x/bridge/keeper/msg_server_unlock_security_test.go`
- Bridge keeper: `chain/x/bridge/keeper/msg_server.go`
- Bridge types: `chain/x/bridge/types/`
