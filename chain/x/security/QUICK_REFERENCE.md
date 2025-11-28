# Security Module Quick Reference

## File Location
```
/home/decri/blockchain-projects/aura/chain/x/security/module.go
```

## Module Identity
- **Name**: `security`
- **Store Key**: `security`
- **Package**: `github.com/aequitas/aura/chain/x/security`

## Key Types

### AppModuleBasic
```go
type AppModuleBasic struct {
    cdc codec.Codec
}
```

### AppModule
```go
type AppModule struct {
    AppModuleBasic
    keeper *keeper.Keeper
}
```

## Constructor
```go
func NewAppModule(cdc codec.Codec, keeper *keeper.Keeper) AppModule
```

## Core Methods

### Genesis
```go
InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage)
ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage
DefaultGenesis(cdc codec.JSONCodec) json.RawMessage
ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error
```

### Registration
```go
RegisterLegacyAminoCodec(cdc *codec.LegacyAmino)
RegisterInterfaces(registry codectypes.InterfaceRegistry)
RegisterServices(cfg module.Configurator)
RegisterInvariants(ir sdk.InvariantRegistry)
RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux)
```

### ABCI
```go
BeginBlock(ctx context.Context) error
EndBlock(ctx context.Context) error
```

### CLI
```go
GetTxCmd() *cobra.Command
GetQueryCmd() *cobra.Command
```

## BeginBlock Operations (in order)
1. **processKeyRotations** - Crypto key rotation
2. **updateNetworkMetrics** - Network security metrics
3. **checkValidatorSecurity** - Validator monitoring
4. **processWalletSecurity** - Wallet operations
5. **updateIncidentState** - Incident tracking
6. **refreshPrivacyPools** - Privacy maintenance

## EndBlock Operations (in order)
1. **cleanupExpiredSessions** - Session cleanup
2. **finalizeSecurityMetrics** - Metrics aggregation

## Integration Example

### In app.go
```go
import (
    "github.com/aequitas/aura/chain/x/security"
    securitykeeper "github.com/aequitas/aura/chain/x/security/keeper"
    securitytypes "github.com/aequitas/aura/chain/x/security/types"
)

// Create keeper
app.SecurityKeeper = securitykeeper.NewKeeper(
    appCodec,
    keys[securitytypes.StoreKey],
    memKeys[securitytypes.MemStoreKey],
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
    app.BankKeeper,
    app.StakingKeeper,
    app.AccountKeeper,
)

// Create module
securityModule := security.NewAppModule(
    appCodec,
    app.SecurityKeeper,
)

// Add to module manager
app.ModuleManager = module.NewManager(
    // ... other modules
    securityModule,
)
```

## Consolidated Modules
This module replaces:
- `networksecurity`
- `validatorsecurity`
- `walletsecurity`
- `incidentresponse`
- `cryptography`
- `privacy`

## Interface Compliance
- ✅ `module.AppModuleBasic`
- ✅ `module.HasGenesis`
- ✅ `module.HasServices`
- ✅ `appmodule.AppModule`
- ✅ `appmodule.HasBeginBlocker`
- ✅ `appmodule.HasEndBlocker`

## Consensus Version
```go
ConsensusVersion() uint64 // Returns 1
```

## Error Handling
- All BeginBlock/EndBlock operations log errors but don't halt the chain
- Genesis validation panics on invalid state (fail-fast)
- Structured logging with context throughout

## Status
✅ Production-ready implementation
