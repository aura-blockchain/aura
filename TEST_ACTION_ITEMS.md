# Test Coverage Action Items - IMMEDIATE

**Date:** 2025-12-08
**Priority:** CRITICAL - Blocks production readiness
**Full Report:** See `TEST_COVERAGE_ANALYSIS.md`

---

## 🚨 CRITICAL BLOCKER (Must Fix Before ANY Testing)

### Compilation Errors Preventing All Tests from Running

**Issue:** Protobuf type mismatches prevent test execution.

**Impact:** ZERO tests can run. Claims of "~98% pass rate" cannot be verified.

**Fix Required:**
```bash
cd /home/decri/blockchain-projects/aura/chain

# Regenerate all protobuf files
make proto-gen

# Verify dependencies
go mod tidy

# Test compilation
go build ./...

# Run tests
go test ./... -v -cover -coverprofile=coverage.out

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html
```

**Affected Files:**
- `proto/aura/bridge/v1beta1/*.pb.go` - Int/Dec type errors
- `proto/aura/dex/v1beta1/*.pb.go` - Dec type undefined
- `proto/aura/wasm/v1beta1/*.pb.go` - RawContractMessage undefined
- `chain/x/dataregistry/keeper/*.go` - Timestamp.AsTime() missing
- Multiple codec.go files - ServiceDesc naming issues

**Estimated Time:** 2-4 hours

---

## P1: CRITICAL TEST GAPS (Week 1-2)

### 1. Missing Message Handler Tests (3 days)

**Required Files:**
```
/chain/x/security/keeper/msg_server_test.go       [NEW]
/chain/x/identity/keeper/msg_server_test.go       [NEW]
```

**Test Coverage Required:**
- All message validation (invalid inputs, edge cases)
- Authorization checks (verify signers)
- State changes (create/update/delete)
- Error handling (missing fields, duplicates)
- Event emission

**Template:**
```go
func TestMsgServer<Operation>(t *testing.T) {
    tests := []struct {
        name      string
        msg       *pb.Msg<Operation>
        shouldErr bool
        errMsg    string
    }{
        {"valid operation", validMsg, false, ""},
        {"invalid - empty field", invalidMsg, true, "invalid"},
        // ... more cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, err := msgServer.<Operation>(ctx, tt.msg)
            if tt.shouldErr {
                require.Error(t, err)
                require.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
                require.NotNil(t, resp)
            }
        })
    }
}
```

### 2. Missing Query Handler Tests (2 days)

**Required Files:**
```
/chain/x/security/keeper/query_server_test.go     [NEW]
/chain/x/identity/keeper/query_server_test.go     [NEW]
/chain/x/aura-bindings/keeper/query_server_test.go [NEW]
```

**Test Coverage Required:**
- All query endpoints
- Pagination (offset, limit)
- Not found scenarios
- Invalid input handling

### 3. Integration Tests (5 days)

**Current State:** Placeholder implementations only (assertions on `NotNil(ctx)`).

**Required:** Replace placeholders with actual tests in:
```
/chain/testing/integration/comprehensive_integration_test.go
```

**Critical Flows:**
1. **DEX ↔ Bridge Integration**
   - Bridge tokens from external chain
   - Add liquidity to DEX pool
   - Execute swap
   - Bridge tokens back
   - Verify all balances

2. **VC Registry ↔ Identity Change**
   - Create verifiable credential
   - Request identity change
   - Update VC with new identity
   - Verify consistency

3. **Governance ↔ Economic Security**
   - Submit proposal to change economic params
   - Vote and execute
   - Verify economic security module reflects changes

4. **Validator Security ↔ Network Security**
   - Detect malicious validator
   - Network security triggers rate limiting
   - Validator security slashes and jails
   - Test unjail process

### 4. E2E Tests (5 days)

**Current State:** Skeleton implementations only.

**Required:** Implement actual scenarios in:
```
/chain/testing/e2e/e2e_test.go
```

**Critical Scenarios:**
1. **Complete Transaction Lifecycle**
   - Create transaction
   - Sign with private key
   - Broadcast to network
   - Verify inclusion in block
   - Query transaction result

2. **Multi-Validator Consensus**
   - Start 4 validator nodes
   - Submit transactions
   - Verify consensus achieved
   - Check all nodes have same state

3. **Byzantine Fault Tolerance**
   - Start 10 validators
   - Stop 3 validators (< 1/3)
   - Verify chain continues producing blocks
   - Restart stopped validators
   - Verify they sync

---

## P2: IMPORTANT GAPS (Week 3-4)

### 1. Economics Module Tests (5 days)

**Current:** Only msg_server tested (missing query and keeper).

**Required Files:**
```
/chain/x/economics/keeper/query_server_test.go    [NEW]
/chain/x/economics/keeper/keeper_test.go          [NEW]
/chain/x/economics/keeper/invariants_test.go      [NEW]
```

**Test Coverage:**
- 22 query handlers (Params, VestingSchedule, Proposals, etc.)
- Vesting calculation logic
- Inflation rate adjustments
- Treasury operations
- Economic invariants (total supply, vesting consistency)

### 2. Security Module Tests (3 days)

**Current:** Only 2 test files, no msg/query coverage.

**Required Files:**
```
/chain/x/security/keeper/msg_server_test.go       [NEW]
/chain/x/security/keeper/query_server_test.go     [NEW]
/chain/x/security/keeper/guards_extended_test.go  [ENHANCE]
```

### 3. Fix Skipped Tests (2 days)

**Tests currently skipped:**
- `x/governance/keeper/query_server_test.go` - "Test setup needs to be implemented"
- `x/cryptography/keeper/comprehensive_test.go` - "DeleteZKProofConfig not implemented"
- `x/wasm/keeper/msg_server_test.go` - 7 tests skipped (wasmd keeper required)
- `x/vcregistry/query_server_test.go` - "GetDisclosureRequest not implemented"

**Action:** Either implement missing functionality or document deferral reason in ROADMAP.

### 4. Security Attack Tests (3 days)

**Add tests for:**
- Gas griefing attacks (unbounded loops)
- Denial of Service (spam, rate limiting)
- Timestamp manipulation
- Flash loan attacks (DEX)
- Front-running attacks (explicit test for commit-reveal)
- 51% validator attack (bridge)
- Replay attacks (bridge nonce checking)

---

## P3: ENHANCEMENTS (Week 5-6)

### 1. Activate Fuzz Tests (2 days)

**Current:** Fuzz test functions exist but use `_ = variable` pattern (no actual testing).

**Fix:** Integrate with real validation functions:
```go
// Before (placeholder)
func FuzzAddressValidation(f *testing.F) {
    f.Fuzz(func(t *testing.T, addrStr string) {
        _ = addrStr  // ← Does nothing
    })
}

// After (actual test)
func FuzzAddressValidation(f *testing.F) {
    f.Fuzz(func(t *testing.T, addrStr string) {
        addr, err := sdk.AccAddressFromBech32(addrStr)
        if err == nil {
            require.NotNil(t, addr)
            require.Equal(t, addrStr, addr.String())
        }
    })
}
```

### 2. Property-Based Testing (2 days)

**Add invariant tests:**
- DEX: `x * y = k` after every swap
- Bridge: Total supply conservation
- Economics: Vesting schedule totals
- Staking: Total tokens = bonded + unbonding + unbonded

### 3. CI/CD Pipeline (1 day)

**Re-enable GitHub Actions:**
```yaml
# .github/workflows/test.yml
name: Test Suite
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: make proto-gen
      - run: go test -race -cover ./...
      - uses: codecov/codecov-action@v3
```

### 4. Performance Benchmarks (1 day)

**Add benchmarks:**
```go
func BenchmarkAMMSwap(b *testing.B) {
    // Setup pool once
    setup()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        keeper.Swap(ctx, poolID, amountIn, minAmountOut)
    }
    // Target: <1ms per swap
}
```

---

## Quick Reference: File Creation Checklist

### Must Create (P1)
- [ ] `/chain/x/security/keeper/msg_server_test.go`
- [ ] `/chain/x/security/keeper/query_server_test.go`
- [ ] `/chain/x/identity/keeper/msg_server_test.go`
- [ ] `/chain/x/identity/keeper/query_server_test.go`

### Must Enhance (P1)
- [ ] `/chain/testing/integration/comprehensive_integration_test.go` - Replace placeholders
- [ ] `/chain/testing/e2e/e2e_test.go` - Replace skeletons

### Should Create (P2)
- [ ] `/chain/x/economics/keeper/query_server_test.go`
- [ ] `/chain/x/economics/keeper/keeper_test.go`
- [ ] `/chain/x/economics/keeper/invariants_test.go`
- [ ] `/chain/x/aura-bindings/keeper/query_server_test.go`

### Should Enhance (P2)
- [ ] All fuzz tests in `/chain/testing/fuzz/fuzz_test.go`
- [ ] Security attack tests (new files in relevant modules)

---

## Success Criteria

### Before Mainnet Launch:
1. ✅ All compilation errors fixed
2. ✅ `go test ./...` exits with 0 (all tests pass)
3. ✅ Coverage >80% (measured, not claimed)
4. ✅ All P1 tests implemented
5. ✅ Integration tests actually test cross-module flows
6. ✅ E2E tests actually test transaction lifecycles
7. ✅ Security attack tests pass
8. ✅ No skipped tests (all t.Skip() resolved)
9. ✅ CI/CD pipeline green

### Timeline:
- **Week 1:** Fix compilation, verify actual coverage
- **Week 2:** Complete P1 tests (security, identity, integration)
- **Week 3:** Complete P2 tests (economics, skipped tests)
- **Week 4:** Security hardening tests
- **Week 5-6:** Enhancements (fuzz, property-based, CI)

**Total:** 5-6 weeks to production-ready test coverage.

---

## Commands to Run Daily

```bash
# Check compilation
go build ./...

# Run all tests
go test ./... -v

# Check coverage
go test ./... -cover -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total

# Run only fast tests (when available)
go test -short ./...

# Run race detector
go test -race ./...

# Run fuzz tests (1 minute each)
go test -fuzz=. -fuzztime=1m ./testing/fuzz/...
```

---

**Next Steps:**
1. Fix compilation errors (TODAY)
2. Run `go test ./...` and document results
3. Begin P1 test implementation
4. Update ROADMAP_PRODUCTION.md with test coverage tasks

**Status Tracking:** Mark tasks complete in ROADMAP_PRODUCTION.md under new "Test Coverage Phase".
