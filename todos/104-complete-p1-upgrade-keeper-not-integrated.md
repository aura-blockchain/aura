---
status: pending
priority: p1
issue_id: "104"
tags: [code-review, architecture, upgrade, testnet-blocker]
dependencies: ["100"]
---

# P1 CRITICAL: Upgrade Keeper Not Integrated - No Protocol Upgrades Possible

## Problem Statement

The upgrade keeper framework is prepared but **not integrated** into app.go, meaning the chain **cannot perform protocol upgrades without hard forks**.

**Why it matters:** Any bug fix, security patch, or feature addition will require all validators to manually update and coordinate restart - extremely risky for production.

## Findings

### Evidence

**File:** `/home/decri/blockchain-projects/aura/chain/app/upgrades.go` (11 TODOs)

```go
// Line 47: TODO: Integrate upgrade keeper into app.go
// Line 79: TODO: Integrate upgrade keeper
// Line 121: TODO: Run module migrations (requires configurator setup)
// Lines 159-317: Multiple upgrade handler implementations waiting for activation
```

### Current State
- Upgrade handlers defined but not registered
- Upgrade keeper not wired into app initialization
- No way to schedule or execute protocol upgrades
- Module migrations cannot run

### Impact
- Cannot fix bugs without hard fork
- Cannot add features without hard fork
- Validator coordination nightmare for any change
- Security patches require emergency coordination

## Proposed Solutions

### Solution A: Complete Upgrade Keeper Integration (Recommended)
**Effort:** 1-2 days | **Risk:** Low

1. Wire upgrade keeper into app.go
2. Register upgrade handlers
3. Configure module migrations
4. Test upgrade flow on local testnet

```go
// In app.go
func (a *App) setupUpgradeKeeper() {
    a.UpgradeKeeper = upgradekeeper.NewKeeper(
        skipUpgradeHeights,
        runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
        a.appCodec,
        a.homePath,
        a,
        a.AccountKeeper.AddressCodec(),
    )

    // Register handlers
    a.registerUpgradeHandlers()
}

func (a *App) registerUpgradeHandlers() {
    a.UpgradeKeeper.SetUpgradeHandler("v1.0.0", func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
        // Run migrations
        return a.ModuleManager.RunMigrations(ctx, a.configurator, fromVM)
    })
}
```

## Recommended Action

**GO WITH SOLUTION A**: Complete the integration. The framework is already prepared.

## Technical Details

### Affected Files
- `chain/app/app.go` - Wire upgrade keeper
- `chain/app/upgrades.go` - Activate handlers

### Database/State Changes
- Upgrade store already defined in keys

## Acceptance Criteria

- [ ] Upgrade keeper initialized in NewApp()
- [ ] At least one upgrade handler registered
- [ ] `aurad query upgrade plan` works
- [ ] Can schedule and execute test upgrade on local testnet
- [ ] Module migrations run successfully

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Git history analysis identified gap | P1 Critical |

## Resources

- [Cosmos SDK Upgrades](https://docs.cosmos.network/main/building-apps/upgrades)
- [Module Migrations](https://docs.cosmos.network/main/building-modules/upgrade)
