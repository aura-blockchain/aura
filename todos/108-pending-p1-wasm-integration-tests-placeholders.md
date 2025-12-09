---
status: pending
priority: p1
issue_id: "108"
tags: [code-review, testing, wasm, integration, critical]
dependencies: ["100"]
---

# P1 CRITICAL: WASM Integration Tests Are Placeholders

## Problem Statement

The WASM module integration tests are placeholder implementations that don't actually test the smart contract execution functionality.

**Why it matters:** WASM contracts are a core feature. Without proper tests, contract deployment and execution bugs won't be caught before production.

## Findings

### Evidence

**File:** `/home/decri/blockchain-projects/aura/chain/x/wasm/keeper/keeper_test.go`

```go
func TestStoreCode(t *testing.T) {
    // TODO: implement actual test
    t.Skip("not implemented")
}

func TestInstantiateContract(t *testing.T) {
    // TODO: implement actual test
    t.Skip("not implemented")
}

func TestExecuteContract(t *testing.T) {
    // TODO: implement actual test
    t.Skip("not implemented")
}
```

### Missing Test Coverage

| Test Category | Status | Priority |
|---------------|--------|----------|
| Store code | Missing | Critical |
| Instantiate contract | Missing | Critical |
| Execute contract | Missing | Critical |
| Migrate contract | Missing | High |
| Contract queries | Missing | High |
| Gas metering | Missing | High |
| Memory limits | Missing | High |
| Permission checks | Missing | Critical |

## Proposed Solutions

### Solution A: Implement Full Test Suite (Recommended)
**Effort:** 3-5 days | **Risk:** Low

Implement comprehensive integration tests using CosmWasm test framework:

```go
func TestStoreCode(t *testing.T) {
    ctx, keeper := setupKeeper(t)

    // Load actual WASM bytecode
    wasmCode := loadTestContract(t, "cw20_base.wasm")

    // Store code
    codeID, checksum, err := keeper.StoreCode(ctx, creatorAddr, wasmCode)
    require.NoError(t, err)
    require.Equal(t, uint64(1), codeID)

    // Verify storage
    storedCode, err := keeper.GetByteCode(ctx, codeID)
    require.NoError(t, err)
    require.Equal(t, wasmCode, storedCode)
}

func TestInstantiateContract(t *testing.T) {
    ctx, keeper := setupKeeper(t)

    // Store code first
    codeID := storeTestCode(t, ctx, keeper)

    // Instantiate
    initMsg := []byte(`{"name":"Test Token","symbol":"TEST","decimals":6}`)
    contractAddr, _, err := keeper.InstantiateContract(
        ctx, codeID, creatorAddr, nil, initMsg, "test-contract", nil,
    )
    require.NoError(t, err)
    require.NotEmpty(t, contractAddr)

    // Verify contract state
    state := keeper.QueryContractState(ctx, contractAddr, []byte("config"))
    require.Contains(t, string(state), "Test Token")
}
```

## Recommended Action

**GO WITH SOLUTION A**: Implement full integration test suite with real WASM contracts.

## Technical Details

### Affected Files

- `chain/x/wasm/keeper/keeper_test.go`
- `chain/x/wasm/keeper/msg_server_test.go`
- `chain/x/wasm/keeper/query_server_test.go`
- `chain/x/wasm/testdata/` (test contracts)

### Test Contracts Needed

1. `cw20_base.wasm` - Standard token contract
2. `cw721_base.wasm` - NFT contract
3. `simple_counter.wasm` - Minimal test contract

## Acceptance Criteria

- [ ] StoreCode test with actual WASM bytecode
- [ ] InstantiateContract test with initialization
- [ ] ExecuteContract test with state changes
- [ ] QueryContract test for state reads
- [ ] MigrateContract test for upgrades
- [ ] Permission tests (creator-only, admin-only)
- [ ] Gas metering tests
- [ ] Memory limit tests
- [ ] Error handling tests for invalid bytecode
- [ ] Test coverage >80% for WASM module

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Test coverage analysis identified gap | P1 Critical |

## Resources

- [CosmWasm Testing Guide](https://docs.cosmwasm.com/docs/getting-started/testing)
- [wasmd Integration Tests](https://github.com/CosmWasm/wasmd/tree/main/x/wasm/keeper)
