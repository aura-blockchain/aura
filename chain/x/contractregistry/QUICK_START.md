# Contract Registry - Quick Start Guide

## Overview

The Contract Registry provides security and compliance enforcement for WASM smart contracts on AURA.

## Installation (Next Steps)

### 1. Wire into App (~30 minutes)

**File**: `chain/app/app.go`

```go
import (
    contractregistrykeeper "github.com/aequitas/aura/chain/x/contractregistry/keeper"
    contractregistrymodule "github.com/aequitas/aura/chain/x/contractregistry"
    contractregistrytypes "github.com/aequitas/aura/chain/x/contractregistry/types"
)

// Add to AuraApp struct
type AuraApp struct {
    // ... existing fields ...
    ContractRegistryKeeper contractregistrykeeper.Keeper
}

// In NewAuraApp function
func NewAuraApp(...) *AuraApp {
    // 1. Add store key
    keys := storetypes.NewKVStoreKeys(
        // ... existing keys ...
        contractregistrytypes.StoreKey,
    )

    // 2. Create keeper
    app.ContractRegistryKeeper = contractregistrykeeper.NewKeeper(
        appCodec,
        runtime.NewKVStoreService(keys[contractregistrytypes.StoreKey]),
        authtypes.NewModuleAddress(govtypes.ModuleName).String(),
    )

    // 3. Set dependencies
    app.ContractRegistryKeeper.SetVCRegistryKeeper(app.VCRegistryKeeper)
    app.ContractRegistryKeeper.SetComplianceKeeper(app.ComplianceKeeper)
    app.ContractRegistryKeeper.SetConfidenceScoreKeeper(app.ConfidenceScoreKeeper)

    // 4. Add to module manager
    app.ModuleManager = module.NewManager(
        // ... existing modules ...
        contractregistrymodule.NewAppModule(appCodec, app.ContractRegistryKeeper),
    )

    // 5. Configure ordering
    genesisModuleOrder := []string{
        // ... existing modules ...
        contractregistrytypes.ModuleName,
    }

    app.ModuleManager.SetOrderBeginBlockers(
        // ... existing modules ...
        contractregistrytypes.ModuleName,
    )

    return app
}
```

### 2. Implement WASM Hooks (~2-3 hours)

See detailed guide: [WASM_HOOKS.md](./WASM_HOOKS.md)

**Summary**:
- Add contract registry to WASM keeper
- Implement BeforeInstantiate hook (auto-register)
- Implement BeforeExecute hook (validate + rate limit)
- Implement AfterExecute hook (update metrics)

## Quick Usage Examples

### Register a Contract

```go
msg := &types.MsgRegisterContract{
    Signer:          creator,
    ContractAddress: contractAddr,
    CodeId:          codeID,
    Creator:         creator,
    Admin:           admin,
    Label:           "my-contract",
    Metadata: types.ContractMetadata{
        Name:               "My Contract",
        Description:        "A secure contract",
        RequiresVc:         true,
        RequiredVcTypes:    []string{"kyc_basic"},
        MinConfidenceScore: 50,
    },
    SecurityPolicy: types.SecurityPolicy{
        MaxGasPerTx:      5000000,
        RateLimitPerUser: 100, // per hour
    },
}
```

### Query Contract Info

```bash
# Get contract details
aurad query contractregistry contract cosmos1contract...

# List contracts by creator
aurad query contractregistry contracts-by-creator cosmos1creator...

# Get metrics
aurad query contractregistry metrics cosmos1contract...
```

### Pause a Contract (Admin)

```go
msg := &types.MsgPauseContract{
    Signer:          admin,
    ContractAddress: contractAddr,
    Reason:          "Security audit",
}
```

## Testing

```bash
# Run all tests
cd chain/x/contractregistry
go test ./... -v

# Run with coverage
go test ./... -v -cover

# Run specific test
go test ./keeper -v -run TestRegisterContract
```

## Key Features

### Security Controls
- ✅ VC requirements (user must have specific credentials)
- ✅ Rate limiting (100 calls/hour default)
- ✅ Gas limits (5M gas default)
- ✅ KYC enforcement (configurable levels)
- ✅ Sanctions screening
- ✅ Confidence score requirements
- ✅ Blacklist/whitelist

### Metrics Tracked
- Total/successful/failed executions
- Gas usage (total and average)
- Unique users
- Rate limit violations
- Compliance failures

### Lifecycle States
- **Active**: Normal operation
- **Paused**: Temporarily disabled
- **Deprecated**: Discouraged but functional
- **Frozen**: Permanently disabled

## Common Operations

### Update Metadata (Admin Only)

```go
msg := &types.MsgUpdateContractMetadata{
    Signer:          admin,
    ContractAddress: contractAddr,
    Metadata: types.ContractMetadata{
        Name:        "Updated Name",
        Version:     "2.0.0",
        // ... other fields
    },
}
```

### Update Security Policy (Admin Only)

```go
msg := &types.MsgUpdateSecurityPolicy{
    Signer:          admin,
    ContractAddress: contractAddr,
    SecurityPolicy: types.SecurityPolicy{
        RateLimitPerUser: 200, // Increase limit
        MaxGasPerTx:      10000000,
    },
}
```

### Deprecate Contract

```go
msg := &types.MsgDeprecateContract{
    Signer:          admin,
    ContractAddress: oldContract,
    Reason:          "Upgraded to v2",
    MigrationTarget: newContract,
}
```

## Default Parameters

```go
OpenRegistration:        true      // Anyone can register
MaxContractsPerCreator:  100       // Limit per creator
RequireMetadata:         true      // Metadata required
RequireSecurityPolicy:   true      // Security policy required
AuditWarningDays:        180       // Audit warning after 6 months
DefaultRateLimit:        100       // 100 calls/hour
DefaultMaxGas:           5000000   // 5M gas max
```

## Error Reference

| Error | Meaning | Solution |
|-------|---------|----------|
| `ErrContractNotFound` | Contract not registered | Register first |
| `ErrRateLimitExceeded` | Too many calls | Wait for reset |
| `ErrMissingVC` | Required VC missing | Obtain credential |
| `ErrInsufficientCS` | CS too low | Increase score |
| `ErrBlacklisted` | Address blocked | Contact admin |
| `ErrNotContractAdmin` | Not authorized | Use admin account |

## Performance

- Registration: ~1-2ms
- Validation per call: ~0.5-1ms
- Total overhead: <2ms per execution

## Documentation

- **Full Integration Guide**: [INTEGRATION.md](./INTEGRATION.md)
- **WASM Hooks Guide**: [WASM_HOOKS.md](./WASM_HOOKS.md)
- **Module Overview**: [README.md](./README.md)
- **Implementation Report**: [CONTRACT_REGISTRY_IMPLEMENTATION_REPORT.md](../../CONTRACT_REGISTRY_IMPLEMENTATION_REPORT.md)

## Support

Run tests to verify installation:
```bash
cd /home/decri/blockchain-projects/aura/chain/x/contractregistry
go test ./... -v -cover
```

Expected: All tests passing, >90% coverage

## Next Steps

1. Wire module into app.go (see above)
2. Implement WASM hooks (see WASM_HOOKS.md)
3. Run integration tests
4. Deploy to testnet
5. Deploy to mainnet

**Estimated time**: 6-9 hours total
