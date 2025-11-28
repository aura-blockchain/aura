package app

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	contractregistrytypes "github.com/aequitas/aura/chain/x/contractregistry/types"
	walletsecuritytypes "github.com/aequitas/aura/chain/x/walletsecurity/types"
)

// UpgradeInfo contains metadata about a chain upgrade.
type UpgradeInfo struct {
	// Name is the unique identifier for this upgrade
	Name string

	// Height is the block height at which the upgrade should occur
	// If 0, the upgrade will occur based on time
	Height int64

	// Info is a URL or description providing upgrade information
	Info string

	// StoreUpgrades defines store additions, deletions, and renames
	StoreUpgrades *storetypes.StoreUpgrades
}

const (
	// Upgrade names - must be unique and follow semantic versioning
	UpgradeV1_0_0 = "v1.0.0" // Initial mainnet launch
	UpgradeV1_1_0 = "v1.1.0" // Add contract registry and enhanced security
	UpgradeV1_2_0 = "v1.2.0" // Add privacy features and cross-chain
)

// RegisterUpgradeHandlers registers all upgrade handlers for the chain.
// This must be called during app initialization before starting the chain.
//
// Note: Upgrade handlers are currently prepared but not active until
// the upgrade keeper is integrated into the app. This will be done in
// a future enhancement.
func (app *App) RegisterUpgradeHandlers() {
	// TODO: Integrate upgrade keeper into app.go
	// For now, upgrade handlers are defined but not registered
	// Uncomment when upgrade keeper is added to App struct

	/*
		// Register v1.1.0 upgrade - adds contract registry and security enhancements
		app.RegisterUpgradeHandler(
			UpgradeV1_1_0,
			app.CreateUpgradeHandler(UpgradeV1_1_0, &storetypes.StoreUpgrades{
				Added: []string{
					contractregistrytypes.StoreKey,
					walletsecuritytypes.StoreKey,
					cryptographytypes.StoreKey,
				},
			}),
		)

		// Register v1.2.0 upgrade - adds privacy and cross-chain features
		app.RegisterUpgradeHandler(
			UpgradeV1_2_0,
			app.CreateUpgradeHandler(UpgradeV1_2_0, &storetypes.StoreUpgrades{
				// No new stores in this upgrade, only parameter updates
			}),
		)
	*/
}

// RegisterUpgradeHandler is a wrapper around the upgrade handler registration.
// It ensures proper logging and error handling for upgrade execution.
//
// Note: This will be activated when upgrade keeper is integrated.
func (app *App) RegisterUpgradeHandler(planName string, handler upgradetypes.UpgradeHandler) {
	// TODO: Integrate upgrade keeper
	// app.upgradeKeeper.SetUpgradeHandler(planName, handler)
	app.Logger().Info("upgrade handler prepared", "name", planName)
}

// CreateUpgradeHandler creates an upgrade handler function for a specific upgrade.
// The handler performs the necessary state migrations and module initialization.
//
// Upgrade handlers must be deterministic and idempotent:
// - They should produce the same result regardless of how many times they run
// - They should not rely on external state or randomness
// - They should handle both fresh upgrades and recovery scenarios
func (app *App) CreateUpgradeHandler(
	planName string,
	storeUpgrades *storetypes.StoreUpgrades,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		app.Logger().Info(
			"starting upgrade execution",
			"upgrade", planName,
			"height", sdkCtx.BlockHeight(),
		)

		// Execute upgrade-specific logic based on the plan name
		switch planName {
		case UpgradeV1_1_0:
			if err := app.upgradeV1_1_0(sdkCtx); err != nil {
				return nil, fmt.Errorf("failed to execute v1.1.0 upgrade: %w", err)
			}

		case UpgradeV1_2_0:
			if err := app.upgradeV1_2_0(sdkCtx); err != nil {
				return nil, fmt.Errorf("failed to execute v1.2.0 upgrade: %w", err)
			}

		default:
			// For unknown upgrades, just run module migrations
			app.Logger().Info("executing default upgrade handler", "upgrade", planName)
		}

		// TODO: Run module migrations - requires configurator to be set up
		// Run module migrations - this handles consensus version updates
		// and calls each module's migration functions
		// versionMap, err := app.moduleManager.RunMigrations(ctx, configurator, fromVM)
		// if err != nil {
		// 	return nil, fmt.Errorf("failed to run module migrations: %w", err)
		// }
		versionMap := fromVM

		app.Logger().Info(
			"upgrade execution completed successfully",
			"upgrade", planName,
			"height", sdkCtx.BlockHeight(),
		)

		return versionMap, nil
	}
}

// upgradeV1_1_0 performs the v1.1.0 upgrade logic.
// This upgrade adds:
// - Contract registry module for smart contract governance
// - Security module (consolidated: replaces walletsecurity, validatorsecurity, cryptography, etc.)
func (app *App) upgradeV1_1_0(ctx sdk.Context) error {
	app.Logger().Info("executing v1.1.0 upgrade")

	// Initialize contract registry with default params
	if app.contractRegistryKeeper != nil {
		params := contractregistrytypes.DefaultParams()
		if err := app.contractRegistryKeeper.SetParams(ctx, params); err != nil {
			return fmt.Errorf("failed to initialize contract registry params: %w", err)
		}
		app.Logger().Info("initialized contract registry module")
	}

	// Wallet security module is initialized via genesis
	app.Logger().Info("wallet security module ready")

	// TODO: Update DEX params for enhanced security
	// Requires EnableCircuitBreaker and MaxPriceImpact fields to be added to Params
	// if app.dexKeeper != nil {
	// 	dexParams := app.dexKeeper.GetParams(ctx)
	// 	// Enable circuit breaker for large price movements
	// 	dexParams.EnableCircuitBreaker = true
	// 	dexParams.MaxPriceImpact = math.LegacyNewDecWithPrec(10, 2) // 10% max impact
	// 	if err := app.dexKeeper.SetParams(ctx, dexParams); err != nil {
	// 		return fmt.Errorf("failed to update dex params: %w", err)
	// 	}
	// 	app.Logger().Info("updated dex module parameters")
	// }

	app.Logger().Info("v1.1.0 upgrade completed successfully")
	return nil
}

// upgradeV1_2_0 performs the v1.2.0 upgrade logic.
// This upgrade adds:
// - Privacy feature enhancements
// - Cross-chain bridge improvements
// - Economics (governance) parameter updates
func (app *App) upgradeV1_2_0(ctx sdk.Context) error {
	app.Logger().Info("executing v1.2.0 upgrade")

	// TODO: Update bridge params for cross-chain improvements
	// Requires MinTransferDelay and EnableFastTransfers fields to be added to Params
	// if app.bridgeKeeper != nil {
	// 	bridgeParams := app.bridgeKeeper.GetParams(ctx)
	// 	// Reduce transfer delays for trusted chains
	// 	bridgeParams.MinTransferDelay = 100 // blocks
	// 	bridgeParams.EnableFastTransfers = true
	// 	if err := app.bridgeKeeper.SetParams(ctx, bridgeParams); err != nil {
	// 		return fmt.Errorf("failed to update bridge params: %w", err)
	// 	}
	// 	app.Logger().Info("updated bridge module parameters")
	// }

	// TODO: Update economics (governance) params
	// Economics module is consolidated (replaces governance module)
	// if app.economicsKeeper != nil {
	// 	// Get and update economics params
	// 	app.Logger().Info("updated economics module parameters")
	// }

	// TODO: Update compliance params for enhanced privacy
	// Requires EnableZKProofs field to be added to ComplianceParams
	// if app.complianceKeeper != nil {
	// 	complianceParams := compliancetypes.DefaultParams()
	// 	// Enable zero-knowledge proof support
	// 	complianceParams.EnableZKProofs = true
	// 	if err := app.complianceKeeper.SetParams(ctx, complianceParams); err != nil {
	// 		return fmt.Errorf("failed to update compliance params: %w", err)
	// 	}
	// 	app.Logger().Info("updated compliance module parameters")
	// }

	app.Logger().Info("v1.2.0 upgrade completed successfully")
	return nil
}

// LoadUpgradeStoreLoader returns a store loader for handling store upgrades.
// This is called during app initialization to prepare for pending upgrades.
func (app *App) LoadUpgradeStoreLoader(upgradeInfo upgradetypes.Plan) *storetypes.StoreUpgrades {
	// Determine which store upgrades to apply based on the upgrade plan
	switch upgradeInfo.Name {
	case UpgradeV1_1_0:
		return &storetypes.StoreUpgrades{
			Added: []string{
				contractregistrytypes.StoreKey,
				walletsecuritytypes.StoreKey,
			},
		}

	case UpgradeV1_2_0:
		// No store changes in v1.2.0
		return &storetypes.StoreUpgrades{}

	default:
		// No store upgrades for unknown plans
		return &storetypes.StoreUpgrades{}
	}
}

// ShouldHaltChain determines if the chain should halt based on upgrade conditions.
// This implements safety checks to prevent chain progression under unsafe conditions.
//
// Note: Upgrade checking is disabled until upgrade keeper is integrated.
func (app *App) ShouldHaltChain(ctx sdk.Context) bool {
	// TODO: Enable when upgrade keeper is integrated
	/*
		// Check if there's a pending upgrade
		plan, found := app.upgradeKeeper.GetUpgradePlan(ctx)
		if !found {
			return false
		}

		// Halt if we've reached the upgrade height
		if plan.ShouldExecute(ctx) {
			app.Logger().Info(
				"chain halt triggered for upgrade",
				"upgrade", plan.Name,
				"height", ctx.BlockHeight(),
			)
			return true
		}
	*/

	// Additional safety checks
	if app.shouldHaltForSecurity(ctx) {
		app.Logger().Error("chain halt triggered for security reasons")
		return true
	}

	return false
}

// shouldHaltForSecurity checks security conditions that might require chain halt.
// This is a critical safety mechanism to prevent chain operation under unsafe conditions.
func (app *App) shouldHaltForSecurity(ctx sdk.Context) bool {
	// TODO: Check validator security status
	// Requires fixing GetLastValidators signature and Keeper accessor
	// Check validator security status
	// if app.validatorSecurityKeeper != nil {
	// 	// Halt if too many validators are jailed
	// 	totalValidators, err := app.StakingKeeper.GetLastValidators(ctx)
	// 	if err != nil {
	// 		app.Logger().Error("failed to get validators", "error", err)
	// 		return false
	// 	}
	// 	jailedCount := 0
	// 	for _, val := range totalValidators {
	// 		if val.IsJailed() {
	// 			jailedCount++
	// 		}
	// 	}
	//
	// 	// Halt if more than 1/3 of validators are jailed (BFT safety threshold)
	// 	if len(totalValidators) > 0 && jailedCount*3 > len(totalValidators) {
	// 		app.Logger().Error(
	// 			"too many validators jailed, halting chain",
	// 			"total", len(totalValidators),
	// 			"jailed", jailedCount,
	// 		)
	// 		return true
	// 	}
	// }

	// TODO: Check for critical bridge issues
	// Requires IsPaused method to be implemented
	// if app.bridgeKeeper != nil {
	// 	// Halt if bridge is paused due to security incident
	// 	if app.bridgeKeeper.IsPaused(ctx) {
	// 		app.Logger().Error("bridge paused, halting chain for safety")
	// 		return true
	// 	}
	// }

	// TODO: Check for governance emergency halt
	// Requires IsEmergencyHalt method to be implemented
	// if app.govKeeper != nil {
	// 	if app.govKeeper.IsEmergencyHalt(ctx) {
	// 		app.Logger().Error("emergency halt activated by governance")
	// 		return true
	// 	}
	// }

	return false
}

// ValidateUpgradePlan validates an upgrade plan before it's scheduled.
// This ensures upgrade plans are well-formed and safe to execute.
func ValidateUpgradePlan(plan upgradetypes.Plan) error {
	if plan.Name == "" {
		return fmt.Errorf("upgrade plan name cannot be empty")
	}

	if plan.Height <= 0 {
		return fmt.Errorf("upgrade height must be positive, got %d", plan.Height)
	}

	// Validate that the upgrade name is recognized
	knownUpgrades := map[string]bool{
		UpgradeV1_0_0: true,
		UpgradeV1_1_0: true,
		UpgradeV1_2_0: true,
	}

	if !knownUpgrades[plan.Name] {
		return fmt.Errorf("unknown upgrade plan: %s", plan.Name)
	}

	return nil
}
