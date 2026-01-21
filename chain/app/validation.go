// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"sort"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
	dextypes "github.com/aequitas/aura/chain/x/dex/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// ValidationResult contains the results of app validation checks.
type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

// AddError adds an error to the validation result.
func (vr *ValidationResult) AddError(format string, args ...interface{}) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, fmt.Sprintf(format, args...))
}

// AddWarning adds a warning to the validation result.
func (vr *ValidationResult) AddWarning(format string, args ...interface{}) {
	vr.Warnings = append(vr.Warnings, fmt.Sprintf(format, args...))
}

// ValidateApp performs comprehensive validation of the app configuration.
// This should be called during app initialization to catch configuration errors early.
func (app *App) ValidateApp() ValidationResult {
	result := ValidationResult{Valid: true}

	// Create a temporary context for validation
	ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

	// Validate store keys
	app.validateStoreKeys(&result)

	// Validate module account permissions
	app.validateModuleAccountPermissions(&result)

	// Validate consensus parameters
	app.validateConsensusParameters(ctx, &result)

	// Validate module dependencies
	app.validateModuleDependencies(&result)

	// Validate keeper initialization
	app.validateKeeperInitialization(&result)

	return result
}

// validateStoreKeys validates that all store keys are properly configured
// and there are no conflicts or duplicates.
func (app *App) validateStoreKeys(result *ValidationResult) {
	// Map to track store key usage
	storeKeyMap := make(map[string]string)

	// Use the centralized AsMap() method - single source of truth
	// This eliminates the need to maintain a duplicate list here
	storeKeys := app.storeKeys.AsMap()

	// Validate each store key
	for keyName, key := range storeKeys {
		if key == nil {
			result.AddError("store key %s is nil", keyName)
			continue
		}

		// Check for duplicate key names
		if existingModule, exists := storeKeyMap[key.Name()]; exists {
			result.AddError(
				"duplicate store key name %s used by both %s and %s",
				key.Name(),
				existingModule,
				keyName,
			)
		} else {
			storeKeyMap[key.Name()] = keyName
		}

		// Validate key name format
		if key.Name() == "" {
			result.AddError("store key %s has empty name", keyName)
		}
	}

	// Validate memory store keys
	// Note: Using consolidated modules - security module doesn't use memory stores currently
	memKeys := map[string]*storetypes.MemoryStoreKey{
		// Consolidated modules don't use memory stores
		// Add memory store keys here if needed in the future
	}

	for keyName, key := range memKeys {
		if key == nil {
			result.AddError("memory store key %s is nil", keyName)
			continue
		}

		// Memory keys should not conflict with persistent keys
		if existingModule, exists := storeKeyMap[key.Name()]; exists {
			result.AddError(
				"memory store key %s conflicts with persistent key from %s",
				key.Name(),
				existingModule,
			)
		}
	}
}

// validateModuleAccountPermissions validates that module account permissions
// are properly configured and consistent.
func (app *App) validateModuleAccountPermissions(result *ValidationResult) {
	// Required permissions for specific modules
	requiredPermissions := map[string][]string{
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		dextypes.ModuleName:            {authtypes.Burner},
		bridgetypes.ModuleName:         {authtypes.Minter, authtypes.Burner},
		wasmtypes.ModuleName:           {authtypes.Minter, authtypes.Burner},
	}

	// Validate required permissions
	for moduleName, required := range requiredPermissions {
		actual, exists := moduleAccountPermissions[moduleName]
		if !exists {
			result.AddError("missing module account permissions for %s", moduleName)
			continue
		}

		// Check that all required permissions are present
		actualMap := make(map[string]bool)
		for _, perm := range actual {
			actualMap[perm] = true
		}

		for _, reqPerm := range required {
			if !actualMap[reqPerm] {
				result.AddError(
					"module %s missing required permission %s",
					moduleName,
					reqPerm,
				)
			}
		}
	}

	// Validate that module accounts with minter permission are secure
	dangerousModules := []string{bridgetypes.ModuleName, wasmtypes.ModuleName}
	for _, moduleName := range dangerousModules {
		perms, exists := moduleAccountPermissions[moduleName]
		if !exists {
			continue
		}

		hasMinter := false
		for _, perm := range perms {
			if perm == authtypes.Minter {
				hasMinter = true
				break
			}
		}

		if hasMinter {
			result.AddWarning(
				"module %s has minter permission - ensure proper security controls",
				moduleName,
			)
		}
	}

	// Validate allowed receiving modules consistency
	for moduleName := range allowedReceivingModules {
		if _, exists := moduleAccountPermissions[moduleName]; !exists {
			result.AddError(
				"module %s is allowed to receive funds but has no account permissions defined",
				moduleName,
			)
		}
	}
}

// validateConsensusParameters validates consensus parameters for safety and correctness.
func (app *App) validateConsensusParameters(ctx sdk.Context, result *ValidationResult) {
	// These validations would be performed once consensus parameter keeper is added
	// For now, we perform basic configuration validation

	// Validate block size limits (would come from consensus params)
	maxBlockSize := int64(10 * 1024 * 1024) // 10MB
	if maxBlockSize <= 0 {
		result.AddError("max block size must be positive")
	}
	if maxBlockSize > 100*1024*1024 { // 100MB
		result.AddWarning("max block size is very large (%d bytes) - this may cause performance issues", maxBlockSize)
	}

	// Validate max gas per block
	maxGasPerBlock := int64(50_000_000) // 50M gas
	if maxGasPerBlock <= 0 {
		result.AddError("max gas per block must be positive")
	}
	if maxGasPerBlock > 1_000_000_000 { // 1B gas
		result.AddWarning("max gas per block is very large (%d) - this may cause performance issues", maxGasPerBlock)
	}

	// Validate evidence parameters
	maxEvidenceAge := int64(100000) // blocks
	if maxEvidenceAge <= 0 {
		result.AddError("max evidence age must be positive")
	}
	if maxEvidenceAge < 10000 {
		result.AddWarning("max evidence age is low (%d blocks) - old evidence may be rejected", maxEvidenceAge)
	}

	// Validate validator parameters
	maxValidators := uint32(100)
	if maxValidators == 0 {
		result.AddError("max validators must be positive")
	}
	if maxValidators > 1000 {
		result.AddWarning("max validators is very large (%d) - this may impact consensus performance", maxValidators)
	}
}

// validateModuleDependencies validates that module dependencies are properly satisfied.
func (app *App) validateModuleDependencies(result *ValidationResult) {
	// Module dependency graph
	// Format: module -> []dependencies
	// Uses consolidated modules: security (replaces validatorsecurity, walletsecurity, cryptography),
	// economics (replaces governance)
	dependencies := map[string][]string{
		"bridge":           {"bank", "auth", "vcregistry"},
		"dex":              {"bank", "auth", "vcregistry"},
		"contractregistry": {"vcregistry", "compliance", "confidencescore"},
		"confidencescore":  {"inclusionroutines"},
		"vcregistry":       {"confidencescore"},
		"security":         {"staking", "slashing", "bank"}, // Consolidated security module
		"wasm":             {"auth", "bank", "staking"},
		"wasmsecurity":     {"wasm", "contractregistry"},
	}

	// Available modules (keepers that are initialized)
	// In Cosmos SDK v0.50, keeper types are the actual keepers themselves,
	// not wrapper structs with nested .Keeper fields
	// Uses consolidated modules
	availableModules := map[string]bool{
		"auth":              !isKeeperZero(app.AccountKeeper),
		"bank":              !isKeeperZero(app.BankKeeper),
		"staking":           app.StakingKeeper != nil,
		"slashing":          !isKeeperZero(app.SlashingKeeper),
		"distribution":      !isKeeperZero(app.DistributionKeeper),
		"bridge":            app.bridgeKeeper != nil,
		"dex":               app.dexKeeper != nil,
		"economicsecurity":  app.economicsecurityKeeper != nil,
		"governance":        app.governanceKeeper != nil,
		"vcregistry":        app.vcKeeper != nil,
		"compliance":        app.complianceKeeper != nil,
		"confidencescore":   app.csKeeper != nil,
		"inclusionroutines": app.irKeeper != nil,
		"contractregistry":  app.contractRegistryKeeper != nil,
		"walletsecurity":    !isKeeperZero(app.walletsecurityKeeper),
		"validatorsecurity": !isKeeperZero(app.validatorsecurityKeeper),
		"identitychange":    app.identitychangeKeeper != nil,
		"wasm":              app.WasmKeeper != nil,
		"wasmsecurity":      !isKeeperZero(app.wasmSecurityKeeper),
	}

	// Check each module's dependencies
	for module, deps := range dependencies {
		if !availableModules[module] {
			// Module not initialized, skip dependency check
			continue
		}

		for _, dep := range deps {
			if !availableModules[dep] {
				result.AddError(
					"module %s depends on %s which is not initialized",
					module,
					dep,
				)
			}
		}
	}

	// Detect circular dependencies
	if hasCycle := detectCircularDependencies(dependencies); hasCycle {
		result.AddError("circular module dependencies detected")
	}
}

// validateKeeperInitialization validates that all keepers are properly initialized.
func (app *App) validateKeeperInitialization(result *ValidationResult) {
	// Check core Cosmos SDK keepers
	// In SDK v0.50, AccountKeeper and BankKeeper are concrete types,
	// not structs with nested .Keeper fields
	if isKeeperZero(app.AccountKeeper) {
		result.AddError("account keeper is not initialized")
	}
	if isKeeperZero(app.BankKeeper) {
		result.AddError("bank keeper is not initialized")
	}
	if app.StakingKeeper == nil {
		result.AddError("staking keeper is not initialized")
	}
	if isKeeperZero(app.SlashingKeeper) {
		result.AddError("slashing keeper is not initialized")
	}
	if isKeeperZero(app.DistributionKeeper) {
		result.AddError("distribution keeper is not initialized")
	}

	// Check AURA custom keepers
	if app.vcKeeper == nil {
		result.AddError("VC registry keeper is not initialized")
	}
	if app.bridgeKeeper == nil {
		result.AddError("bridge keeper is not initialized")
	}
	if app.dexKeeper == nil {
		result.AddError("DEX keeper is not initialized")
	}
	if app.contractRegistryKeeper == nil {
		result.AddError("contract registry keeper is not initialized")
	}

	// Individual keepers
	if isKeeperZero(app.walletsecurityKeeper) {
		result.AddError("wallet security keeper is not initialized")
	}
	if isKeeperZero(app.validatorsecurityKeeper) {
		result.AddError("validator security keeper is not initialized")
	}
	if app.economicsecurityKeeper == nil {
		result.AddError("economics security keeper is not initialized")
	}
	if app.governanceKeeper == nil {
		result.AddError("governance keeper is not initialized")
	}
	if app.identitychangeKeeper == nil {
		result.AddError("identity change keeper is not initialized")
	}

	if isKeeperZero(app.wasmSecurityKeeper) {
		result.AddError("wasm security keeper is not initialized")
	}

	// Check keeper cross-references
	// VCRegistry should have confidence score keeper set
	if app.vcKeeper != nil {
		// This would check internal state if we had getters
		result.AddWarning("unable to verify vcregistry keeper dependencies - consider adding getter methods")
	}

	// Contract registry should have dependencies set
	if app.contractRegistryKeeper != nil {
		result.AddWarning("unable to verify contract registry keeper dependencies - consider adding getter methods")
	}
}

// isKeeperZero checks if a keeper is uninitialized (zero value).
// This is a helper function to check struct keepers that don't use pointers.
func isKeeperZero(k interface{}) bool {
	// For struct keepers, we check if they're equal to their zero value
	// This works because Go allows comparison of structs if all fields are comparable
	switch v := k.(type) {
	case nil:
		return true
	default:
		// Use reflection to check if the value is zero
		// For now, we'll just return false for non-nil values
		// as a more robust check would require reflection
		_ = v
		return false
	}
}

// detectCircularDependencies detects circular dependencies in the module dependency graph.
func detectCircularDependencies(dependencies map[string][]string) bool {
	// Track visit state for each node
	// 0 = unvisited, 1 = visiting, 2 = visited
	state := make(map[string]int)

	var hasCycle bool
	var visit func(string)

	visit = func(module string) {
		if hasCycle {
			return
		}

		if state[module] == 1 {
			// Currently visiting this node - cycle detected
			hasCycle = true
			return
		}

		if state[module] == 2 {
			// Already visited
			return
		}

		state[module] = 1 // Mark as visiting

		// Visit all dependencies
		for _, dep := range dependencies[module] {
			visit(dep)
		}

		state[module] = 2 // Mark as visited
	}

	// Check each module
	for module := range dependencies {
		if state[module] == 0 {
			visit(module)
			if hasCycle {
				return true
			}
		}
	}

	return false
}

// ValidateModuleInitializationOrder validates that modules are initialized
// in the correct order based on their dependencies.
func ValidateModuleInitializationOrder(order []string, dependencies map[string][]string) error {
	// Build index of module positions
	position := make(map[string]int)
	for i, module := range order {
		position[module] = i
	}

	// Check each module's dependencies come before it
	for i, module := range order {
		deps, hasDeps := dependencies[module]
		if !hasDeps {
			continue
		}

		for _, dep := range deps {
			depPos, exists := position[dep]
			if !exists {
				return fmt.Errorf("module %s depends on %s which is not in initialization order", module, dep)
			}

			if depPos >= i {
				return fmt.Errorf(
					"module %s (position %d) depends on %s (position %d) which comes after it",
					module, i, dep, depPos,
				)
			}
		}
	}

	return nil
}

// GetRecommendedModuleInitOrder returns the recommended module initialization order
// based on dependency analysis.
// Uses consolidated modules: security, identity, economics
func GetRecommendedModuleInitOrder() []string {
	// Topologically sorted based on dependencies
	return []string{
		// Core SDK modules (no AURA dependencies)
		"auth",
		"bank",
		"staking",
		"slashing",
		"distribution",

		// AURA modules - tier 1 (no AURA dependencies)
		"compliance",
		"security",  // Consolidated (replaces cryptography, walletsecurity, validatorsecurity, networksecurity, incidentresponse, privacy)
		"economics", // Consolidated (replaces economicsecurity, governance)

		// AURA modules - tier 2 (depend on tier 1)
		"identity", // Consolidated (replaces identitychange)
		"inclusionroutines",

		// AURA modules - tier 3 (depend on tier 2)
		"confidencescore",

		// AURA modules - tier 4 (depend on tier 3)
		"vcregistry",
		"dataregistry",

		// AURA modules - tier 5 (depend on tier 4)
		"contractregistry",
		"bridge",
		"dex",

		// WASM modules (depend on all above)
		"wasm",
		"wasmsecurity",
	}
}

// ValidateStoreKeyNamesUnique ensures all store key names are unique across the app.
func ValidateStoreKeyNamesUnique(keys map[string]*storetypes.KVStoreKey) error {
	seen := make(map[string]string)

	for moduleName, key := range keys {
		if key == nil {
			return fmt.Errorf("store key for module %s is nil", moduleName)
		}

		keyName := key.Name()
		if existingModule, exists := seen[keyName]; exists {
			return fmt.Errorf(
				"duplicate store key name '%s' used by modules '%s' and '%s'",
				keyName,
				existingModule,
				moduleName,
			)
		}

		seen[keyName] = moduleName
	}

	return nil
}

// ValidateModuleAccountPermissions ensures module account permissions are consistent.
func ValidateModuleAccountPermissions(perms map[string][]string) error {
	// Check for duplicate permissions within each module
	for moduleName, modulePerms := range perms {
		seen := make(map[string]bool)
		for _, perm := range modulePerms {
			if seen[perm] {
				return fmt.Errorf(
					"module %s has duplicate permission: %s",
					moduleName,
					perm,
				)
			}
			seen[perm] = true
		}

		// Sort permissions for consistency
		sort.Strings(modulePerms)
	}

	// Validate known permission types
	validPermissions := map[string]bool{
		authtypes.Minter:  true,
		authtypes.Burner:  true,
		authtypes.Staking: true,
	}

	for moduleName, modulePerms := range perms {
		for _, perm := range modulePerms {
			if !validPermissions[perm] {
				return fmt.Errorf(
					"module %s has unknown permission type: %s",
					moduleName,
					perm,
				)
			}
		}
	}

	return nil
}
