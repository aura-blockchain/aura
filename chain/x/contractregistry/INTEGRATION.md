# Contract Registry Integration Guide

## Overview

The Contract Registry module provides security and compliance enforcement for WASM smart contracts on the AURA blockchain. This guide explains how to integrate the module with the WASM module and other components.

## Architecture

```
WASM Contract Upload/Execute
         ↓
Contract Registry Validation
         ↓
    ┌─────────────────────┐
    │ Check Registration  │
    │ Check Status        │
    │ Check Rate Limits   │
    │ Check Compliance    │
    │ Check VC Requirements│
    └─────────────────────┘
         ↓
    Approved/Rejected
```

## Integration Points

### 1. WASM Module Integration

#### Before Contract Upload Hook

Add validation before allowing contract code upload:

```go
// In WASM module's InstantiateContract function
func (k Keeper) InstantiateContract(ctx sdk.Context, codeID uint64, creator sdk.AccAddress, ...) {
    // ... existing code ...

    // Validate creator has required VCs (if contract registry requires)
    if k.contractRegistry != nil {
        // Check creator permissions
        // This can be added as a hook
    }

    // ... rest of instantiation ...
}
```

#### Before Contract Execution Hook

Add this to the WASM module's Execute function:

```go
// In WASM module's Execute function
func (k Keeper) Execute(ctx sdk.Context, contractAddr sdk.AccAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins) (*sdk.Result, error) {
    // Contract Registry validation
    if k.contractRegistry != nil {
        gasLimit := ctx.GasMeter().Limit()
        if err := k.contractRegistry.ValidateContractExecution(
            ctx,
            contractAddr.String(),
            caller.String(),
            gasLimit,
        ); err != nil {
            return nil, fmt.Errorf("contract execution validation failed: %w", err)
        }

        // Increment rate limit counter
        k.contractRegistry.IncrementRateLimit(ctx, contractAddr.String(), caller.String())
    }

    // ... existing execution code ...

    // Update metrics after execution
    if k.contractRegistry != nil {
        gasUsed := ctx.GasMeter().GasConsumed()
        success := (err == nil)
        k.contractRegistry.UpdateMetricsOnExecution(
            ctx,
            contractAddr.String(),
            gasUsed,
            success,
        )
    }

    return result, err
}
```

### 2. App Wiring

In `chain/app/app.go`, add the contract registry module:

```go
import (
    contractregistrykeeper "github.com/aequitas/aura/chain/x/contractregistry/keeper"
    contractregistrymodule "github.com/aequitas/aura/chain/x/contractregistry"
    contractregistrytypes "github.com/aequitas/aura/chain/x/contractregistry/types"
)

type AuraApp struct {
    // ... existing fields ...

    ContractRegistryKeeper contractregistrykeeper.Keeper
}

func NewAuraApp(...) *AuraApp {
    // ... existing code ...

    // Create contract registry keeper
    app.ContractRegistryKeeper = contractregistrykeeper.NewKeeper(
        appCodec,
        runtime.NewKVStoreService(keys[contractregistrytypes.StoreKey]),
        authtypes.NewModuleAddress(govtypes.ModuleName).String(),
    )

    // Set dependencies
    app.ContractRegistryKeeper.SetVCRegistryKeeper(app.VCRegistryKeeper)
    app.ContractRegistryKeeper.SetComplianceKeeper(app.ComplianceKeeper)
    app.ContractRegistryKeeper.SetConfidenceScoreKeeper(app.ConfidenceScoreKeeper)

    // ... create module manager ...

    app.ModuleManager = module.NewManager(
        // ... existing modules ...
        contractregistrymodule.NewAppModule(appCodec, app.ContractRegistryKeeper),
    )

    // Set WASM contract registry reference
    // app.WasmKeeper.SetContractRegistry(&app.ContractRegistryKeeper)

    return app
}
```

### 3. Store Keys

Add the store key in `chain/app/app.go`:

```go
keys := storetypes.NewKVStoreKeys(
    // ... existing keys ...
    contractregistrytypes.StoreKey,
)
```

### 4. Module Ordering

Configure genesis and begin/end block ordering:

```go
genesisModuleOrder := []string{
    // ... existing modules ...
    contractregistrytypes.ModuleName,
}

app.ModuleManager.SetOrderBeginBlockers(
    // ... existing modules ...
    contractregistrytypes.ModuleName,
)

app.ModuleManager.SetOrderEndBlockers(
    // ... existing modules ...
    contractregistrytypes.ModuleName,
)
```

## Usage Examples

### Registering a Contract

```go
// Create registration message
msg := &types.MsgRegisterContract{
    Signer:          creatorAddr,
    ContractAddress: contractAddr,
    CodeId:          codeID,
    Creator:         creatorAddr,
    Admin:           adminAddr,
    Label:           "my-dapp",
    Metadata: types.ContractMetadata{
        Name:               "My DApp",
        Description:        "A decentralized application",
        Version:            "1.0.0",
        Homepage:           "https://mydapp.com",
        SourceCodeUrl:      "https://github.com/mydapp/contracts",
        Tags:               []string{"defi", "dao"},
        RequiresVc:         true,
        RequiredVcTypes:    []string{"kyc_basic", "verified_human"},
        MinConfidenceScore: 50,
        RequiredKycLevel:   1,
        CheckSanctions:     true,
    },
    SecurityPolicy: types.SecurityPolicy{
        AllowPause:       true,
        AllowMigration:   true,
        MaxGasPerTx:      5000000,
        RateLimitPerUser: 100, // 100 calls per hour
        BlacklistedAddresses: []string{},
        WhitelistedAddresses: []string{}, // Empty = all allowed
    },
    Compliance: types.ComplianceRequirements{
        EnforceKyc:             true,
        MinKycLevel:            1,
        EnforceSanctionsCheck:  true,
        RequireAudit:           false,
    },
}

// Submit transaction
_, err := msgClient.RegisterContract(ctx, msg)
```

### Querying Contract Info

```bash
# Query contract information
aurad query contractregistry contract <contract-address>

# Query contracts by creator
aurad query contractregistry contracts-by-creator <creator-address>

# Query contracts by tag
aurad query contractregistry contracts-by-tag defi

# Query all registered contracts
aurad query contractregistry contracts

# Query contract metrics
aurad query contractregistry metrics <contract-address>
```

### Updating Contract Settings

```go
// Update metadata (admin only)
msgMetadata := &types.MsgUpdateContractMetadata{
    Signer:          adminAddr,
    ContractAddress: contractAddr,
    Metadata: types.ContractMetadata{
        Name:        "Updated Name",
        Description: "Updated description",
        Version:     "2.0.0",
        // ... other fields ...
    },
}

// Update security policy (admin only)
msgSecurity := &types.MsgUpdateSecurityPolicy{
    Signer:          adminAddr,
    ContractAddress: contractAddr,
    SecurityPolicy: types.SecurityPolicy{
        MaxGasPerTx:      10000000,
        RateLimitPerUser: 200,
        // ... other fields ...
    },
}

// Pause contract (admin or governance)
msgPause := &types.MsgPauseContract{
    Signer:          adminAddr,
    ContractAddress: contractAddr,
    Reason:          "Security audit in progress",
}

// Unpause contract
msgUnpause := &types.MsgUnpauseContract{
    Signer:          adminAddr,
    ContractAddress: contractAddr,
}

// Deprecate contract
msgDeprecate := &types.MsgDeprecateContract{
    Signer:          adminAddr,
    ContractAddress: contractAddr,
    Reason:          "Upgraded to v2",
    MigrationTarget: newContractAddr,
}
```

## Security Features

### 1. VC Verification

Contracts can require users to have specific Verifiable Credentials:

```go
Metadata: types.ContractMetadata{
    RequiresVc:      true,
    RequiredVcTypes: []string{
        "kyc_basic",        // Basic KYC verification
        "verified_human",   // Proof of humanity
        "age_over_18",      // Age verification
    },
}
```

### 2. Confidence Score Enforcement

Require minimum confidence score:

```go
Metadata: types.ContractMetadata{
    MinConfidenceScore: 50, // 0-100 scale
}
```

### 3. Rate Limiting

Prevent DoS attacks via rate limiting:

```go
SecurityPolicy: types.SecurityPolicy{
    RateLimitPerUser: 100, // Max calls per user per hour
}
```

Rate limits are checked before execution and automatically reset every hour.

### 4. Gas Limits

Enforce maximum gas per transaction:

```go
SecurityPolicy: types.SecurityPolicy{
    MaxGasPerTx: 5000000, // Maximum gas allowed
}
```

### 5. Address Lists

Use blacklists and whitelists:

```go
SecurityPolicy: types.SecurityPolicy{
    BlacklistedAddresses: []string{
        "cosmos1malicious...",
    },
    WhitelistedAddresses: []string{
        "cosmos1trusted1...",
        "cosmos1trusted2...",
    },
}
```

**Note**: If `WhitelistedAddresses` is non-empty, ONLY those addresses can interact with the contract.

### 6. Compliance Checks

Enforce KYC and sanctions screening:

```go
Compliance: types.ComplianceRequirements{
    EnforceKyc:            true,
    MinKycLevel:           2, // 1=basic, 2=intermediate, 3=advanced
    EnforceSanctionsCheck: true,
}
```

## Metrics and Monitoring

The module tracks comprehensive metrics for each contract:

- `TotalExecutions` - Total number of calls
- `SuccessfulExecutions` - Successful calls
- `FailedExecutions` - Failed calls
- `TotalGasUsed` - Cumulative gas consumption
- `AvgGasPerExecution` - Average gas per call
- `UniqueUsers` - Number of unique callers
- `RateLimitViolations` - Rate limit hits
- `ComplianceFailures` - Compliance check failures
- `LastExecution` - Timestamp of last execution

Query metrics:

```bash
aurad query contractregistry metrics <contract-address>
```

## Governance Integration

The module supports governance actions via the authority address (governance module):

### Update Module Parameters

```go
// Governance proposal to update params
proposal := &types.MsgUpdateParams{
    Authority: govModuleAddr,
    Params: types.ContractRegistryParams{
        OpenRegistration:        true,
        MaxContractsPerCreator:  100,
        RequireMetadata:         true,
        RequireSecurityPolicy:   true,
        RequireComplianceConfig: false,
        AuditWarningDays:        180,
        DefaultRateLimit:        100,
        DefaultMaxGas:           5000000,
    },
}
```

### Emergency Actions

Governance can pause or freeze contracts:

```go
// Pause via governance
msgPause := &types.MsgPauseContract{
    Signer:          govModuleAddr,
    ContractAddress: contractAddr,
    Reason:          "Security vulnerability discovered",
}
```

## Best Practices

### For Contract Developers

1. **Register Immediately**: Register your contract right after deployment
2. **Set Appropriate Limits**: Configure realistic gas and rate limits
3. **Keep Metadata Updated**: Maintain accurate version and documentation links
4. **Monitor Metrics**: Regularly check contract metrics for anomalies
5. **Plan Deprecation**: Use deprecation feature before upgrading contracts

### For Contract Users

1. **Check Registration**: Verify contracts are registered before use
2. **Review Requirements**: Check VC and compliance requirements
3. **Monitor Rate Limits**: Be aware of rate limit constraints
4. **Verify Status**: Ensure contract is active, not paused or deprecated

### For Validators/Operators

1. **Enable All Checks**: Configure all security modules (VC, compliance, CS)
2. **Set Reasonable Defaults**: Configure sensible default limits
3. **Monitor Metrics**: Track overall registry metrics
4. **Regular Cleanup**: Rate limit cleanup runs automatically every 100 blocks

## Error Handling

Common errors and solutions:

| Error | Cause | Solution |
|-------|-------|----------|
| `ErrContractNotFound` | Contract not registered | Register contract first |
| `ErrContractPaused` | Contract is paused | Wait for unpause or contact admin |
| `ErrRateLimitExceeded` | Too many calls | Wait for hourly reset |
| `ErrGasLimitExceeded` | Gas request too high | Reduce gas or contact admin |
| `ErrMissingVC` | User lacks required VC | Obtain required credentials |
| `ErrInsufficientCS` | Confidence score too low | Increase confidence score |
| `ErrKYCRequired` | KYC verification needed | Complete KYC process |
| `ErrBlacklisted` | Address is blacklisted | Contact contract admin |

## Testing

Run tests:

```bash
cd chain/x/contractregistry
go test ./... -v -cover
```

Expected coverage: >90% for all packages.

## Future Enhancements

Planned features:

1. **Analytics Dashboard**: Web-based metrics visualization
2. **Auto-pause**: Automatic pause on anomaly detection
3. **Multi-signature Admin**: Require multiple signatures for admin actions
4. **Tiered Rate Limits**: Different limits based on user CS or VCs
5. **Contract Categories**: Enhanced categorization and discovery
6. **Audit Integration**: Automated audit report verification
7. **Cross-chain Registry**: Support for cross-chain contract references

## Support

For issues or questions:
- GitHub: [AURA Blockchain Issues](https://github.com/aequitas/aura/issues)
- Documentation: [docs/modules/contractregistry/](../../../docs/modules/contractregistry/)
- Community: AURA Discord/Telegram
