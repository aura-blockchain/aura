# AURA Core Chain Integration - Developer Guide

## Quick Start

This guide covers the core chain integration components implemented for AURA blockchain.

---

## Architecture Overview

### Transaction Processing Pipeline

```
User Transaction
    ↓
[Ante Handler Chain]
    ↓
├─ Setup Context
├─ Extension Options
├─ Transaction Validation
├─ Fee Deduction
├─ Wallet Security Check ←── AURA Custom
├─ Compliance Check ←── AURA Custom
├─ Signature Verification
├─ Sequence Increment
└─ WASM Gas Limiting
    ↓
[Message Execution]
    ↓
[State Changes]
```

---

## Components

### 1. Ante Handler (`ante.go`)

**Purpose**: Pre-execution transaction validation and fee processing

**Key Functions**:
```go
// Create ante handler with all decorators
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error)

// Validate ante handler options
func ValidateAnteHandlerOptions(options HandlerOptions) error
```

**Custom Decorators**:
```go
// Wallet security checks
type WalletSecurityDecorator struct {
    keeper walletsecuritykeeper.Keeper
}

// Compliance checks (sanctions, AML)
type ComplianceDecorator struct {
    keeper *compliancekeeper.Keeper
}
```

**Usage Example**:
```go
// In app initialization
app.SetupAnteHandler()

// The ante handler automatically:
// 1. Validates transaction structure
// 2. Checks wallet rate limits
// 3. Performs compliance screening
// 4. Deducts fees
// 5. Verifies signatures
```

---

### 2. Upgrade Handlers (`upgrades.go`)

**Purpose**: Safe chain upgrades with state migrations

**Defined Upgrades**:
- `v1.0.0` - Initial mainnet
- `v1.1.0` - Contract registry and security
- `v1.2.0` - Privacy and cross-chain

**Key Functions**:
```go
// Register all upgrade handlers
func (app *App) RegisterUpgradeHandlers()

// Create handler for specific upgrade
func (app *App) CreateUpgradeHandler(
    planName string,
    storeUpgrades *storetypes.StoreUpgrades,
) upgradetypes.UpgradeHandler

// Check if chain should halt
func (app *App) ShouldHaltChain(ctx sdk.Context) bool
```

**Adding New Upgrade**:
```go
const UpgradeV1_3_0 = "v1.3.0"

func (app *App) RegisterUpgradeHandlers() {
    app.RegisterUpgradeHandler(
        UpgradeV1_3_0,
        app.CreateUpgradeHandler(UpgradeV1_3_0, &storetypes.StoreUpgrades{
            Added: []string{"newmodule"},
        }),
    )
}

func (app *App) upgradeV1_3_0(ctx sdk.Context) error {
    // Your upgrade logic here
    return nil
}
```

**Safety Features**:
- Automatic halt if >1/3 validators jailed
- Bridge pause detection
- Governance emergency halt

---

### 3. Validation (`validation.go`)

**Purpose**: Startup validation to catch configuration errors

**What's Validated**:
1. Store key uniqueness
2. Module account permissions
3. Consensus parameters
4. Module dependencies
5. Keeper initialization

**Key Functions**:
```go
// Run all validations
func (app *App) ValidateApp() ValidationResult

// Validate store keys
func (app *App) validateStoreKeys(result *ValidationResult)

// Validate module dependencies
func (app *App) validateModuleDependencies(result *ValidationResult)

// Detect circular dependencies
func detectCircularDependencies(dependencies map[string][]string) bool
```

**Usage**:
```go
// Validation runs automatically in NewAppWithLogger()
validationResult := app.ValidateApp()
if !validationResult.Valid {
    for _, err := range validationResult.Errors {
        logger.Error("validation error", "error", err)
    }
    panic("app validation failed")
}
```

**Module Dependency Graph**:
```go
dependencies := map[string][]string{
    "bank":             {"auth"},
    "staking":          {"auth", "bank"},
    "confidencescore":  {"inclusionroutines"},
    "vcregistry":       {"confidencescore"},
    "contractregistry": {"vcregistry", "compliance", "confidencescore"},
    "bridge":           {"bank", "auth", "vcregistry"},
    "dex":              {"bank", "auth", "vcregistry"},
}
```

---

### 4. Dependency Injection (`depinject.go`)

**Purpose**: Explicit keeper dependency management

**Keeper Tiers**:
```
Tier 0: Core SDK (auth, bank, staking, slashing, distribution)
Tier 1: No AURA deps (compliance, cryptography, walletsecurity, governance)
Tier 2: Basic AURA (identitychange, inclusionroutines)
Tier 3: Intermediate (confidencescore)
Tier 4: Advanced (vcregistry, dataregistry)
Tier 5: Complex (contractregistry, bridge, dex, aiassistant)
Security: Final (wasm, wasmsecurity, validatorsecurity)
```

**Key Functions**:
```go
// Get dependency graph
func KeeperDependencyGraph() map[string][]string

// Get initialization order
func GetKeeperInitializationOrder() []string

// Topological sort
func TopologicalSort(dependencies map[string][]string) ([]string, error)

// Validate dependencies
func ValidateKeeperDependencies(container *KeeperContainer) error
```

**Adding New Module**:
```go
// 1. Add to dependency graph
graph["newmodule"] = []string{"dependency1", "dependency2"}

// 2. Add to initialization order (respect tier)
order := append(order, "newmodule")

// 3. Validate new order
err := ValidateKeeperInitializationOrder(order)
```

---

## Common Tasks

### Adding Custom Ante Decorator

```go
// 1. Create decorator struct
type MyCustomDecorator struct {
    keeper MyKeeper
}

// 2. Implement AnteHandle
func (d MyCustomDecorator) AnteHandle(
    ctx sdk.Context,
    tx sdk.Tx,
    simulate bool,
    next sdk.AnteHandler,
) (sdk.Context, error) {
    // Your validation logic

    // Continue chain
    return next(ctx, tx, simulate)
}

// 3. Add to ante handler chain in ante.go
anteDecorators := []sdk.AnteDecorator{
    // ... existing decorators
    NewMyCustomDecorator(options.MyKeeper),
    // ... remaining decorators
}
```

### Adding Module with Dependencies

```go
// 1. Add to dependency graph in depinject.go
func KeeperDependencyGraph() map[string][]string {
    return map[string][]string{
        // ... existing
        "mymodule": {"dependency1", "dependency2"},
    }
}

// 2. Add to initialization order
func GetKeeperInitializationOrder() []string {
    return []string{
        // ... existing in order
        "mymodule", // Add in correct tier
    }
}

// 3. Update validation in validation.go
availableModules := map[string]bool{
    // ... existing
    "mymodule": app.myKeeper != nil,
}
```

### Creating Chain Upgrade

```go
// 1. Define upgrade name
const UpgradeV2_0_0 = "v2.0.0"

// 2. Register handler
func (app *App) RegisterUpgradeHandlers() {
    app.RegisterUpgradeHandler(
        UpgradeV2_0_0,
        app.CreateUpgradeHandler(UpgradeV2_0_0, &storetypes.StoreUpgrades{
            Added:   []string{"newstore"},
            Deleted: []string{"oldstore"},
            Renamed: []storetypes.StoreRename{
                {OldKey: "old", NewKey: "new"},
            },
        }),
    )
}

// 3. Implement upgrade logic
func (app *App) upgradeV2_0_0(ctx sdk.Context) error {
    // Migrate params
    if app.myKeeper != nil {
        params := app.myKeeper.GetParams(ctx)
        params.NewField = defaultValue
        app.myKeeper.SetParams(ctx, params)
    }

    // Migrate state
    // ...

    return nil
}

// 4. Add to CreateUpgradeHandler switch
switch planName {
    case UpgradeV2_0_0:
        if err := app.upgradeV2_0_0(sdkCtx); err != nil {
            return nil, err
        }
}
```

---

## Testing

### Unit Tests

```bash
# Test ante handler
go test -v -run TestNewAnteHandler ./app

# Test validation
go test -v -run TestValidate ./app

# Test dependency injection
go test -v -run TestKeeperDependency ./app

# All app tests
go test -v ./app/...
```

### Integration Tests

```go
func TestChainStartup(t *testing.T) {
    app := NewApp()
    require.NotNil(t, app)

    // Validate app initialized correctly
    ctx := app.BaseApp.NewContext(false, tmproto.Header{})
    // Test module interactions
}
```

---

## Debugging

### Validation Failures

```go
// Enable verbose logging
validationResult := app.ValidateApp()
if !validationResult.Valid {
    for _, err := range validationResult.Errors {
        log.Printf("ERROR: %s\n", err)
    }
    for _, warn := range validationResult.Warnings {
        log.Printf("WARN: %s\n", warn)
    }
}
```

### Ante Handler Issues

```go
// Add logging to custom decorators
func (d MyDecorator) AnteHandle(...) (sdk.Context, error) {
    logger := ctx.Logger().With("decorator", "MyDecorator")
    logger.Info("processing transaction")

    // Your logic

    logger.Info("transaction processed successfully")
    return next(ctx, tx, simulate)
}
```

### Dependency Issues

```go
// Validate initialization order
order := GetKeeperInitializationOrder()
if err := ValidateKeeperInitializationOrder(order); err != nil {
    log.Fatalf("Invalid initialization order: %v", err)
}

// Check for circular dependencies
graph := KeeperDependencyGraph()
if detectCircularDependencies(graph) {
    log.Fatal("Circular dependency detected!")
}
```

---

## Performance Tips

### Ante Handler
- Keep decorators lightweight
- Use early returns for invalid transactions
- Minimize state reads
- Use context caching

### Validation
- Validation runs once at startup
- Pre-compute dependency graphs
- Cache validation results

### Dependency Injection
- Initialize keepers in dependency order
- Avoid redundant initialization
- Use lazy initialization where appropriate

---

## Security Best Practices

### Ante Handler
1. Always validate inputs
2. Check gas consumption
3. Prevent reentrancy
4. Rate limit expensive operations
5. Clear error messages (no sensitive data)

### Upgrades
1. Test on testnet first
2. Have rollback plan
3. Coordinate with validators
4. Monitor upgrade progress
5. Validate state after upgrade

### Validation
1. Fail fast on errors
2. Log all validation failures
3. No silent failures
4. Comprehensive coverage

---

## References

- [Cosmos SDK Ante Handler Docs](https://docs.cosmos.network/main/basics/gas-fees.html#antehandler)
- [Upgrade Module Docs](https://docs.cosmos.network/main/modules/upgrade)
- [Module Dependency Best Practices](https://docs.cosmos.network/main/building-modules/intro.html)

---

## Getting Help

1. Check implementation in `chain/app/*.go`
2. Review tests in `chain/app/*_test.go`
3. Read full report in `CORE_CHAIN_INTEGRATION_REPORT.md`
4. Check remaining work in `REMAINING_WORK.md`

---

**Last Updated**: 2025-11-25
**Version**: 1.0
