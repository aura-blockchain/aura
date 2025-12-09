---
status: pending
priority: p1
issue_id: "106"
tags: [code-review, security, panic, chain-halt, critical]
dependencies: ["100"]
---

# P1 CRITICAL: MustMarshal/MustUnmarshal Panic Usage - Chain Halt Risk

## Problem Statement

The codebase uses `MustMarshal` and `MustUnmarshal` extensively (95+ instances), which panic on error. In a consensus system, **panics cause chain halts**.

**Why it matters:** A malformed state or invalid data causes ALL validators to panic simultaneously, halting the network. This is a denial-of-service vulnerability.

## Findings

### Evidence

**Pattern found across all modules:**

```go
// DANGEROUS - Panics on error
bz := k.cdc.MustMarshal(&value)
k.cdc.MustUnmarshal(bz, &result)
```

### Affected Modules (Partial List)

| Module | MustMarshal Count | MustUnmarshal Count |
|--------|------------------|---------------------|
| dex | 15+ | 20+ |
| identity | 10+ | 12+ |
| bridge | 8+ | 10+ |
| vcregistry | 5+ | 8+ |
| compliance | 6+ | 7+ |
| confidencescore | 4+ | 5+ |

### Risk Scenarios

1. **State corruption** - Corrupted data in state causes unmarshal panic
2. **Malicious proposals** - Attacker submits proposal with invalid encoding
3. **Upgrade bugs** - Migration creates incompatible state
4. **Network partition** - Different nodes have different state versions

## Proposed Solutions

### Solution A: Replace with Safe Error Handling (Recommended)
**Effort:** 2-3 days | **Risk:** Low

Replace all `Must*` variants with error-returning versions:

```go
// SAFE - Returns error
bz, err := k.cdc.Marshal(&value)
if err != nil {
    return errorsmod.Wrap(ErrMarshalFailed, err.Error())
}

var result MyType
if err := k.cdc.Unmarshal(bz, &result); err != nil {
    return errorsmod.Wrap(ErrUnmarshalFailed, err.Error())
}
```

### Solution B: Panic Recovery Middleware
**Effort:** 1 day | **Risk:** Medium

Add panic recovery in ABCI handlers (not recommended as primary solution).

## Recommended Action

**GO WITH SOLUTION A**: Replace all `Must*` calls with safe error handling. This is the Cosmos SDK best practice for production chains.

## Technical Details

### Search Pattern

```bash
# Find all instances
grep -rn "MustMarshal\|MustUnmarshal" chain/x/
```

### Affected Files

All keeper files across all modules containing state operations.

### Database/State Changes

None - code changes only.

## Acceptance Criteria

- [ ] Zero uses of MustMarshal in production code paths
- [ ] Zero uses of MustUnmarshal in production code paths
- [ ] All marshal/unmarshal operations return proper errors
- [ ] Error messages include context (module, key, operation)
- [ ] Unit tests verify error handling works correctly

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Security audit identified vulnerability | P1 Critical |

## Resources

- [Cosmos SDK Error Handling](https://docs.cosmos.network/main/building-modules/errors)
- [Chain Halt Post-Mortems](https://github.com/cosmos/cosmos-sdk/issues?q=panic+chain+halt)
