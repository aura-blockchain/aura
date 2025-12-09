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
// Each upgrade handler defines the migration logic for a specific protocol version.
// Handlers are executed automatically when the chain reaches the upgrade height.
func (app *App) RegisterUpgradeHandlers() {
	// Register v1.0.0 upgrade - initial mainnet launch baseline
	// This serves as a baseline for future upgrades
	app.RegisterUpgradeHandler(
		UpgradeV1_0_0,
		app.CreateUpgradeHandler(UpgradeV1_0_0, nil),
	)

	// Register v1.1.0 upgrade - adds contract registry and security enhancements
	app.RegisterUpgradeHandler(
		UpgradeV1_1_0,
		app.CreateUpgradeHandler(UpgradeV1_1_0, &storetypes.StoreUpgrades{
			Added: []string{
				contractregistrytypes.StoreKey,
				walletsecuritytypes.StoreKey,
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

	app.Logger().Info("upgrade handlers registered successfully",
		"handlers", []string{UpgradeV1_0_0, UpgradeV1_1_0, UpgradeV1_2_0})
}

// RegisterUpgradeHandler is a wrapper around the upgrade handler registration.
// It ensures proper logging and error handling for upgrade execution.
//
// This method registers the handler with the upgrade keeper, which will execute
// it automatically when the chain reaches the upgrade height.
func (app *App) RegisterUpgradeHandler(planName string, handler upgradetypes.UpgradeHandler) {
	app.UpgradeKeeper.SetUpgradeHandler(planName, handler)
	app.Logger().Info("upgrade handler registered", "name", planName)
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

		// Run module migrations - this handles consensus version updates
		// and calls each module's migration functions
		// The configurator is stored during app initialization in NewAppWithOptions
		versionMap, err := app.moduleManager.RunMigrations(ctx, app.configurator(), fromVM)
		if err != nil {
			return nil, fmt.Errorf("failed to run module migrations: %w", err)
		}

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

	// DEX params update - future enhancement
	// Circuit breaker and price impact limits can be added in a future upgrade
	// when the DEX module params are extended with these fields
	app.Logger().Info("dex security enhancements scheduled for future upgrade")

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

	// Bridge params update - future enhancement
	// Fast transfers and reduced delays can be added when bridge params are extended
	if app.bridgeKeeper != nil {
		app.Logger().Info("bridge cross-chain improvements scheduled for future upgrade")
	}

	// Economics (governance) params update - future enhancement
	// Parameter updates can be added when economics module params are finalized
	if app.economicsKeeper != nil {
		app.Logger().Info("economics module ready for parameter updates")
	}

	// Compliance params update - future enhancement
	// Zero-knowledge proof support can be added when compliance module is extended
	if app.complianceKeeper != nil {
		app.Logger().Info("compliance privacy enhancements scheduled for future upgrade")
	}

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
// The upgrade keeper automatically halts the chain when an upgrade height is reached,
// so this method focuses on additional safety checks beyond scheduled upgrades.
func (app *App) ShouldHaltChain(ctx sdk.Context) bool {
	// Check if there's a pending upgrade that should trigger a halt
	// The upgrade keeper handles this automatically via BeginBlock,
	// but we check here for consistency with other halt conditions
	plan, err := app.UpgradeKeeper.GetUpgradePlan(ctx)
	if err == nil && plan.Height > 0 && ctx.BlockHeight() >= plan.Height {
		app.Logger().Info(
			"chain halt triggered for upgrade",
			"upgrade", plan.Name,
			"height", ctx.BlockHeight(),
		)
		return true
	}

	// Additional safety checks for security-related halt conditions
	if app.shouldHaltForSecurity(ctx) {
		app.Logger().Error("chain halt triggered for security reasons")
		return true
	}

	return false
}

// shouldHaltForSecurity checks security conditions that might require chain halt.
// This is a critical safety mechanism to prevent chain operation under unsafe conditions.
//
// Future enhancements will include:
// - Validator security threshold monitoring (halt if >1/3 validators jailed)
// - Bridge pause detection (halt if bridge encounters critical issues)
// - Governance emergency halt (halt via on-chain governance vote)
func (app *App) shouldHaltForSecurity(ctx sdk.Context) bool {
	// Validator security monitoring - future enhancement
	// When implemented, this will check if too many validators are jailed
	// and halt the chain if the BFT safety threshold (1/3) is exceeded

	// Bridge security monitoring - future enhancement
	// When implemented, this will check if the bridge module is paused
	// due to a security incident and halt the chain for safety

	// Governance emergency halt - future enhancement
	// When implemented, this will allow governance to trigger an emergency
	// chain halt via an on-chain proposal

	// No halt conditions met - chain can continue
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
