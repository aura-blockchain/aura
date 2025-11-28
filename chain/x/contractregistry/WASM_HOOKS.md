# WASM Module Integration Hooks

This document describes the specific hooks needed in the WASM module to integrate with the Contract Registry.

## Required Changes to WASM Module

### 1. Add Contract Registry Keeper to WASM Keeper

**File**: `chain/x/wasm/keeper/keeper.go`

```go
import (
    contractregistrykeeper "github.com/aequitas/aura/chain/x/contractregistry/keeper"
    contractregistrytypes "github.com/aequitas/aura/chain/x/contractregistry/types"
)

type Keeper struct {
    // ... existing fields ...

    contractRegistry *contractregistrykeeper.Keeper
}

// SetContractRegistry sets the contract registry keeper
func (k *Keeper) SetContractRegistry(registry *contractregistrykeeper.Keeper) {
    k.contractRegistry = registry
}
```

### 2. Hook: Before Contract Instantiation

**File**: `chain/x/wasm/keeper/keeper.go`

Add validation before instantiating a contract:

```go
func (k Keeper) Instantiate(
    ctx sdk.Context,
    codeID uint64,
    creator, admin sdk.AccAddress,
    initMsg []byte,
    label string,
    deposit sdk.Coins,
) (sdk.AccAddress, []byte, error) {
    // ... existing code to validate codeID, creator, etc. ...

    // Generate contract address
    contractAddress := k.generateContractAddress(ctx, codeID, instanceNumber)

    // ===== CONTRACT REGISTRY HOOK =====
    // Auto-register contract with default settings
    if k.contractRegistry != nil {
        info := contractregistrytypes.ContractInfo{
            Address: contractAddress.String(),
            CodeId:  codeID,
            Creator: creator.String(),
            Admin:   admin.String(),
            Label:   label,
            Metadata: contractregistrytypes.ContractMetadata{
                Name:        label,
                Description: "Auto-registered WASM contract",
            },
            SecurityPolicy: contractregistrytypes.SecurityPolicy{
                AllowPause:     true,
                MaxGasPerTx:    k.contractRegistry.GetParams(ctx).DefaultMaxGas,
                RateLimitPerUser: k.contractRegistry.GetParams(ctx).DefaultRateLimit,
            },
            Compliance: contractregistrytypes.ComplianceRequirements{},
            Status:     contractregistrytypes.ContractStatus_CONTRACT_STATUS_ACTIVE,
        }

        if err := k.contractRegistry.RegisterContract(ctx, info); err != nil {
            return nil, nil, fmt.Errorf("contract registry registration failed: %w", err)
        }
    }
    // ===== END CONTRACT REGISTRY HOOK =====

    // ... rest of instantiation logic ...

    return contractAddress, data, nil
}
```

### 3. Hook: Before Contract Execution

**File**: `chain/x/wasm/keeper/keeper.go`

Add validation before executing a contract:

```go
func (k Keeper) Execute(
    ctx sdk.Context,
    contractAddress sdk.AccAddress,
    caller sdk.AccAddress,
    msg []byte,
    coins sdk.Coins,
) (*sdk.Result, error) {
    // ===== CONTRACT REGISTRY HOOK - BEFORE EXECUTION =====
    if k.contractRegistry != nil {
        // Get gas limit from context
        gasLimit := ctx.GasMeter().Limit()

        // Validate execution
        if err := k.contractRegistry.ValidateContractExecution(
            ctx,
            contractAddress.String(),
            caller.String(),
            gasLimit,
        ); err != nil {
            // Increment compliance failure metric
            k.contractRegistry.IncrementMetricsCounter(
                ctx,
                contractAddress.String(),
                "compliance_failure",
            )
            return nil, fmt.Errorf("contract execution validation failed: %w", err)
        }

        // Increment rate limit counter
        k.contractRegistry.IncrementRateLimit(
            ctx,
            contractAddress.String(),
            caller.String(),
        )
    }
    // ===== END CONTRACT REGISTRY HOOK - BEFORE =====

    // ... existing execution code ...
    gasConsumedBefore := ctx.GasMeter().GasConsumed()

    result, err := k.wasmVM.Execute(/* ... */)

    // ===== CONTRACT REGISTRY HOOK - AFTER EXECUTION =====
    if k.contractRegistry != nil {
        gasUsed := ctx.GasMeter().GasConsumed() - gasConsumedBefore
        success := (err == nil)

        // Update execution metrics
        k.contractRegistry.UpdateMetricsOnExecution(
            ctx,
            contractAddress.String(),
            gasUsed,
            success,
        )
    }
    // ===== END CONTRACT REGISTRY HOOK - AFTER =====

    return result, err
}
```

### 4. Hook: Before Contract Migration

**File**: `chain/x/wasm/keeper/keeper.go`

Add validation before migrating a contract:

```go
func (k Keeper) Migrate(
    ctx sdk.Context,
    contractAddress sdk.AccAddress,
    caller sdk.AccAddress,
    newCodeID uint64,
    msg []byte,
) (*sdk.Result, error) {
    // ===== CONTRACT REGISTRY HOOK =====
    if k.contractRegistry != nil {
        // Check if contract is registered
        if !k.contractRegistry.IsContractRegistered(ctx, contractAddress.String()) {
            return nil, fmt.Errorf("contract not registered in contract registry")
        }

        // Get contract info to check if migration is allowed
        info, found := k.contractRegistry.GetContractInfo(ctx, contractAddress.String())
        if found && !info.SecurityPolicy.AllowMigration {
            return nil, fmt.Errorf("contract migration not allowed by security policy")
        }

        // Verify caller is admin
        if found && info.Admin != caller.String() {
            return nil, fmt.Errorf("only contract admin can migrate")
        }
    }
    // ===== END CONTRACT REGISTRY HOOK =====

    // ... existing migration logic ...

    return result, nil
}
```

### 5. Optional Hook: Before Code Upload

**File**: `chain/x/wasm/keeper/keeper.go`

Optionally validate code uploads:

```go
func (k Keeper) Create(
    ctx sdk.Context,
    creator sdk.AccAddress,
    wasmCode []byte,
    instantiateAccess *types.AccessConfig,
) (codeID uint64, checksum []byte, err error) {
    // ===== CONTRACT REGISTRY HOOK (Optional) =====
    if k.contractRegistry != nil {
        params := k.contractRegistry.GetParams(ctx)

        // Check if open registration is disabled
        if !params.OpenRegistration {
            // Only allow uploads from authorized addresses
            // (This would require additional registry functionality)
        }

        // Check creator contract limit
        count := k.contractRegistry.GetCreatorContractCount(ctx, creator.String())
        if params.MaxContractsPerCreator > 0 && count >= params.MaxContractsPerCreator {
            return 0, nil, contractregistrytypes.ErrTooManyContracts
        }
    }
    // ===== END CONTRACT REGISTRY HOOK =====

    // ... existing code upload logic ...

    return codeID, checksum, nil
}
```

## Configuration Example

**File**: `chain/app/app.go`

Complete wiring in the application:

```go
func NewAuraApp(...) *AuraApp {
    // ... existing setup ...

    // Create keepers in order of dependencies

    // 1. Create VC Registry Keeper
    app.VCRegistryKeeper = vcregistrykeeper.NewKeeper(...)

    // 2. Create Compliance Keeper
    app.ComplianceKeeper = compliancekeeper.NewKeeper(...)

    // 3. Create Confidence Score Keeper
    app.ConfidenceScoreKeeper = confidencescorekeeper.NewKeeper(...)

    // 4. Create Contract Registry Keeper
    app.ContractRegistryKeeper = contractregistrykeeper.NewKeeper(
        appCodec,
        runtime.NewKVStoreService(keys[contractregistrytypes.StoreKey]),
        authtypes.NewModuleAddress(govtypes.ModuleName).String(),
    )

    // 5. Wire dependencies
    app.ContractRegistryKeeper.SetVCRegistryKeeper(app.VCRegistryKeeper)
    app.ContractRegistryKeeper.SetComplianceKeeper(app.ComplianceKeeper)
    app.ContractRegistryKeeper.SetConfidenceScoreKeeper(app.ConfidenceScoreKeeper)

    // 6. Create WASM Keeper
    wasmOpts := []wasmkeeper.Option{}
    app.WasmKeeper = wasmkeeper.NewKeeper(
        appCodec,
        runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
        app.AccountKeeper,
        app.BankKeeper,
        app.StakingKeeper,
        distrkeeper.NewQuerier(app.DistrKeeper),
        app.IBCFeeKeeper,
        app.IBCKeeper.ChannelKeeper,
        app.IBCKeeper.PortKeeper,
        scopedWasmKeeper,
        app.TransferKeeper,
        app.MsgServiceRouter(),
        app.GRPCQueryRouter(),
        wasmDir,
        wasmConfig,
        availableCapabilities,
        authtypes.NewModuleAddress(govtypes.ModuleName).String(),
        wasmOpts...,
    )

    // 7. Set contract registry in WASM keeper
    app.WasmKeeper.SetContractRegistry(&app.ContractRegistryKeeper)

    // ... rest of setup ...

    return app
}
```

## Testing Integration

Create integration tests to verify the hooks work correctly:

**File**: `chain/x/wasm/keeper/integration_test.go`

```go
func TestContractRegistryIntegration(t *testing.T) {
    // Setup test environment with both WASM and Contract Registry
    ctx, wasmKeeper, registryKeeper := setupTest(t)

    // Upload code
    creator := sdk.AccAddress("creator")
    codeID, _, err := wasmKeeper.Create(ctx, creator, wasmCode, nil)
    require.NoError(t, err)

    // Instantiate contract (should auto-register)
    contractAddr, _, err := wasmKeeper.Instantiate(
        ctx,
        codeID,
        creator,
        creator,
        initMsg,
        "test-contract",
        nil,
    )
    require.NoError(t, err)

    // Verify contract is registered
    require.True(t, registryKeeper.IsContractRegistered(ctx, contractAddr.String()))

    // Test execution with rate limiting
    caller := sdk.AccAddress("caller")

    // Should succeed initially
    _, err = wasmKeeper.Execute(ctx, contractAddr, caller, execMsg, nil)
    require.NoError(t, err)

    // Exceed rate limit
    params := registryKeeper.GetParams(ctx)
    for i := 0; i < int(params.DefaultRateLimit); i++ {
        registryKeeper.IncrementRateLimit(ctx, contractAddr.String(), caller.String())
    }

    // Should fail due to rate limit
    _, err = wasmKeeper.Execute(ctx, contractAddr, caller, execMsg, nil)
    require.Error(t, err)
    require.Contains(t, err.Error(), "rate limit exceeded")

    // Verify metrics were updated
    metrics, found := registryKeeper.GetContractMetrics(ctx, contractAddr.String())
    require.True(t, found)
    require.Greater(t, metrics.TotalExecutions, uint64(0))
}
```

## Deployment Checklist

When deploying the integration:

- [ ] Contract Registry module added to app
- [ ] Store key allocated for Contract Registry
- [ ] Contract Registry keeper created with dependencies
- [ ] WASM keeper has reference to Contract Registry
- [ ] Hooks added to WASM Instantiate function
- [ ] Hooks added to WASM Execute function
- [ ] Hooks added to WASM Migrate function
- [ ] Module ordering configured (genesis, begin/end block)
- [ ] Integration tests passing
- [ ] Default params configured appropriately
- [ ] Documentation updated

## Performance Considerations

The hooks add minimal overhead:

1. **Instantiate Hook**: ~1-2ms (one-time per contract)
2. **Execute Hook**: ~0.5-1ms per call (validation + rate limit check)
3. **Metrics Update**: ~0.2-0.5ms (async after execution)

Total overhead: <2ms per transaction, negligible compared to WASM execution time.

## Rollback Plan

If issues arise, the Contract Registry can be disabled by:

1. Setting `contractRegistry = nil` in WASM keeper
2. Adding feature flag to skip hooks
3. Reverting to previous version without hooks

All hooks are defensive and won't break existing functionality if the keeper is nil.
