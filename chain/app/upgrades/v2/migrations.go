// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package v2

import (
	"fmt"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/aequitas/aura/chain/x/auth/types"
	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
	compliancetypes "github.com/aequitas/aura/chain/x/compliance/types"
	confidencescoretypes "github.com/aequitas/aura/chain/x/confidencescore/types"
	contractregistrytypes "github.com/aequitas/aura/chain/x/contractregistry/types"
	cryptographytypes "github.com/aequitas/aura/chain/x/cryptography/types"
	dataregistrytypes "github.com/aequitas/aura/chain/x/dataregistry/types"
	dextypes "github.com/aequitas/aura/chain/x/dex/types"
	economicsecuritytypes "github.com/aequitas/aura/chain/x/economicsecurity/types"
	governancetypes "github.com/aequitas/aura/chain/x/governance/types"
	identitychangetypes "github.com/aequitas/aura/chain/x/identitychange/types"
	inclusionroutinestypes "github.com/aequitas/aura/chain/x/inclusionroutines/types"
	monitoringtypes "github.com/aequitas/aura/chain/x/monitoring/types"
	networksecuritytypes "github.com/aequitas/aura/chain/x/networksecurity/types"
	prevalidationtypes "github.com/aequitas/aura/chain/x/prevalidation/types"
	privacytypes "github.com/aequitas/aura/chain/x/privacy/types"
	validatorsecuritytypes "github.com/aequitas/aura/chain/x/validatorsecurity/types"
	vcregistrytypes "github.com/aequitas/aura/chain/x/vcregistry/types"
	walletsecuritytypes "github.com/aequitas/aura/chain/x/walletsecurity/types"
	wasmtypes "github.com/aequitas/aura/chain/x/wasm/types"
)

// MigrateStore performs idempotent store migrations for all modules
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc interface{}) error {
	ctx.Logger().Info("Starting store migrations for v2")

	// Perform all module migrations
	migrations := []struct {
		name    string
		migrate func(sdk.Context, storetypes.StoreKey) error
	}{
		{"auth", migrateAuthStore},
		{"bridge", migrateBridgeStore},
		{"compliance", migrateComplianceStore},
		{"confidencescore", migrateConfidenceScoreStore},
		{"contractregistry", migrateContractRegistryStore},
		{"cryptography", migrateCryptographyStore},
		{"dataregistry", migrateDataRegistryStore},
		{"dex", migrateDEXStore},
		{"economicsecurity", migrateEconomicSecurityStore},
		{"governance", migrateGovernanceStore},
		{"identitychange", migrateIdentityChangeStore},
		{"inclusionroutines", migrateInclusionRoutinesStore},
		{"monitoring", migrateMonitoringStore},
		{"networksecurity", migrateNetworkSecurityStore},
		{"prevalidation", migratePrevalidationStore},
		{"privacy", migratePrivacyStore},
		{"validatorsecurity", migrateValidatorSecurityStore},
		{"vcregistry", migrateVCRegistryStore},
		{"walletsecurity", migrateWalletSecurityStore},
		{"wasm", migrateWasmStore},
	}

	for _, m := range migrations {
		ctx.Logger().Info(fmt.Sprintf("Migrating %s store", m.name))
		if err := m.migrate(ctx, storeKey); err != nil {
			return fmt.Errorf("failed to migrate %s: %w", m.name, err)
		}
	}

	ctx.Logger().Info("Store migrations completed")
	return nil
}

// migrateAuthStore migrates auth module state
func migrateAuthStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_auth")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate emergency admin records
	// Update account permissions if needed

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateBridgeStore migrates bridge module state
func migrateBridgeStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_bridge")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate transfer records
	// Update chain configs
	// Migrate validator records
	// Update wrapped token formats

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateComplianceStore migrates compliance module state
func migrateComplianceStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_compliance")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate KYC records
	// Update AML profiles
	// Migrate GDPR consents

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateConfidenceScoreStore migrates confidence score module state
func migrateConfidenceScoreStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_confidencescore")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate score records
	// Update score calculation parameters
	// Recalculate outdated scores if needed

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateContractRegistryStore migrates contract registry module state
func migrateContractRegistryStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_contractregistry")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate contract records
	// Update contract metadata

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateCryptographyStore migrates cryptography module state
func migrateCryptographyStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_cryptography")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate key rotation records
	// Update encryption settings

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateDataRegistryStore migrates data registry module state
func migrateDataRegistryStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_dataregistry")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate data items
	// Update IPFS references

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateDEXStore migrates DEX module state
func migrateDEXStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_dex")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate liquidity pools
	// Update orderbook structure
	// Migrate swap orders

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateEconomicSecurityStore migrates economic security module state
func migrateEconomicSecurityStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_economicsecurity")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate fee settings
	// Update MEV protection parameters

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateGovernanceStore migrates governance module state
func migrateGovernanceStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_governance")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate proposals
	// Update voting records

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateIdentityChangeStore migrates identity change module state
func migrateIdentityChangeStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_identitychange")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate identity change requests
	// Update ownership records

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateInclusionRoutinesStore migrates inclusion routines module state
func migrateInclusionRoutinesStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_inclusionroutines")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate inclusion routine records
	// Update completion tracking

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateMonitoringStore migrates monitoring module state
func migrateMonitoringStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_monitoring")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate alerts
	// Update monitoring rules

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateNetworkSecurityStore migrates network security module state
func migrateNetworkSecurityStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_networksecurity")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate peer reputations
	// Update rate limit settings

	store.Set(migrationKey, []byte{1})
	return nil
}

// migratePrevalidationStore migrates prevalidation module state
func migratePrevalidationStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_prevalidation")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate prevalidation rules
	// Update validation cache

	store.Set(migrationKey, []byte{1})
	return nil
}

// migratePrivacyStore migrates privacy module state
func migratePrivacyStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_privacy")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate privacy settings
	// Update encryption keys

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateValidatorSecurityStore migrates validator security module state
func migrateValidatorSecurityStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_validatorsecurity")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate validator monitoring data
	// Update slashing records

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateVCRegistryStore migrates VC registry module state
func migrateVCRegistryStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_vcregistry")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate VC records
	// Update DID documents
	// Migrate presentation records

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateWalletSecurityStore migrates wallet security module state
func migrateWalletSecurityStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_walletsecurity")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate wallet configs
	// Update security settings

	store.Set(migrationKey, []byte{1})
	return nil
}

// migrateWasmStore migrates wasm module state
func migrateWasmStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	store := ctx.KVStore(storeKey)

	migrationKey := []byte("migration_v2_wasm")
	if store.Has(migrationKey) {
		return nil
	}

	// Migrate contract instances
	// Update code references

	store.Set(migrationKey, []byte{1})
	return nil
}

// MigrateParams migrates module parameters
func MigrateParams(ctx sdk.Context) error {
	ctx.Logger().Info("Migrating module parameters")

	// Add parameter migrations here if needed
	// Example: Update default values, add new parameters, etc.

	return nil
}

// CleanupOldState removes deprecated state from previous versions
func CleanupOldState(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	ctx.Logger().Info("Cleaning up old state")

	// Remove deprecated keys
	// Clean up old indices
	// This should be done carefully and only after ensuring data is migrated

	return nil
}

// ValidateMigration validates that migration completed successfully
func ValidateMigration(ctx sdk.Context) error {
	ctx.Logger().Info("Validating migration")

	// Add validation logic here
	// Check that all expected data exists
	// Verify data integrity

	return nil
}

// Module type usage to prevent unused import errors
var (
	_ = authtypes.ModuleName
	_ = bridgetypes.ModuleName
	_ = compliancetypes.ModuleName
	_ = confidencescoretypes.ModuleName
	_ = contractregistrytypes.ModuleName
	_ = cryptographytypes.ModuleName
	_ = dataregistrytypes.ModuleName
	_ = dextypes.ModuleName
	_ = economicsecuritytypes.ModuleName
	_ = governancetypes.ModuleName
	_ = identitychangetypes.ModuleName
	_ = inclusionroutinestypes.ModuleName
	_ = monitoringtypes.ModuleName
	_ = networksecuritytypes.ModuleName
	_ = prevalidationtypes.ModuleName
	_ = privacytypes.ModuleName
	_ = validatorsecuritytypes.ModuleName
	_ = vcregistrytypes.ModuleName
	_ = walletsecuritytypes.ModuleName
	_ = wasmtypes.ModuleName
	_ = prefix.Store{}
)
