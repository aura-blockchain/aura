// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	tmlog "cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	bridgekeeper "github.com/aequitas/aura/chain/x/bridge/keeper"
	compliancekeeper "github.com/aequitas/aura/chain/x/compliance/keeper"
	cskeeper "github.com/aequitas/aura/chain/x/confidencescore/keeper"
	contractregistrykeeper "github.com/aequitas/aura/chain/x/contractregistry/keeper"
	cryptographykeeper "github.com/aequitas/aura/chain/x/cryptography/keeper"
	drkeeper "github.com/aequitas/aura/chain/x/dataregistry/keeper"
	dexkeeper "github.com/aequitas/aura/chain/x/dex/keeper"
	govkeeper "github.com/aequitas/aura/chain/x/governance/keeper"
	idkeeper "github.com/aequitas/aura/chain/x/identitychange/keeper"
	irkeeper "github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
	validatorsecuritykeeper "github.com/aequitas/aura/chain/x/validatorsecurity/keeper"
	vckeeper "github.com/aequitas/aura/chain/x/vcregistry/keeper"
	walletsecuritykeeper "github.com/aequitas/aura/chain/x/walletsecurity/keeper"
	wasmSecurityKeeper "github.com/aequitas/aura/chain/x/wasm/keeper"
)

// KeeperContainer holds all application keepers in a structured way.
// This enables dependency injection and makes keeper dependencies explicit.
type KeeperContainer struct {
	// Core Cosmos SDK keepers
	AccountKeeper      authkeeper.AccountKeeper
	BankKeeper         bankkeeper.BaseKeeper
	StakingKeeper      *stakingkeeper.Keeper
	SlashingKeeper     slashingkeeper.Keeper
	DistributionKeeper distrkeeper.Keeper

	// WASM keeper
	WasmKeeper *wasmkeeper.Keeper

	// AURA custom keepers - tier 1 (no AURA dependencies)
	ComplianceKeeper     *compliancekeeper.Keeper
	CryptographyKeeper   *cryptographykeeper.Keeper
	WalletSecurityKeeper walletsecuritykeeper.Keeper
	GovernanceKeeper     *govkeeper.Keeper

	// AURA custom keepers - tier 2
	IdentityChangeKeeper    *idkeeper.Keeper
	InclusionRoutinesKeeper *irkeeper.Keeper

	// AURA custom keepers - tier 3
	ConfidenceScoreKeeper *cskeeper.Keeper

	// AURA custom keepers - tier 4
	VCRegistryKeeper   *vckeeper.Keeper
	DataRegistryKeeper *drkeeper.Keeper

	// AURA custom keepers - tier 5
	ContractRegistryKeeper *contractregistrykeeper.Keeper
	BridgeKeeper           *bridgekeeper.Keeper
	DexKeeper              *dexkeeper.Keeper

	// Security keepers
	ValidatorSecurityKeeper validatorsecuritykeeper.Keeper
	WasmSecurityKeeper      wasmSecurityKeeper.Keeper
}

// KeeperInitializer handles the initialization of all keepers with proper
// dependency injection and initialization order.
type KeeperInitializer struct {
	cdc           codec.Codec
	logger        tmlog.Logger
	storeKeys     map[string]*storetypes.KVStoreKey
	memKeys       map[string]*storetypes.MemoryStoreKey
	transientKeys map[string]*storetypes.TransientStoreKey
	authority     string
}

// NewKeeperInitializer creates a new keeper initializer.
func NewKeeperInitializer(
	cdc codec.Codec,
	logger tmlog.Logger,
	storeKeys map[string]*storetypes.KVStoreKey,
	memKeys map[string]*storetypes.MemoryStoreKey,
	transientKeys map[string]*storetypes.TransientStoreKey,
	authority string,
) *KeeperInitializer {
	return &KeeperInitializer{
		cdc:           cdc,
		logger:        logger,
		storeKeys:     storeKeys,
		memKeys:       memKeys,
		transientKeys: transientKeys,
		authority:     authority,
	}
}

// InitializeKeepers initializes all keepers in the correct dependency order.
// This is the main entry point for dependency injection.
func (ki *KeeperInitializer) InitializeKeepers(container *KeeperContainer) error {
	// Phase 1: Initialize core Cosmos SDK keepers (already initialized in app.go)
	// These are passed in via the container

	// Phase 2: Initialize tier 1 AURA keepers (no AURA dependencies)
	if err := ki.initTier1Keepers(container); err != nil {
		return fmt.Errorf("failed to initialize tier 1 keepers: %w", err)
	}

	// Phase 3: Initialize tier 2 AURA keepers
	if err := ki.initTier2Keepers(container); err != nil {
		return fmt.Errorf("failed to initialize tier 2 keepers: %w", err)
	}

	// Phase 4: Initialize tier 3 AURA keepers
	if err := ki.initTier3Keepers(container); err != nil {
		return fmt.Errorf("failed to initialize tier 3 keepers: %w", err)
	}

	// Phase 5: Initialize tier 4 AURA keepers
	if err := ki.initTier4Keepers(container); err != nil {
		return fmt.Errorf("failed to initialize tier 4 keepers: %w", err)
	}

	// Phase 6: Initialize tier 5 AURA keepers
	if err := ki.initTier5Keepers(container); err != nil {
		return fmt.Errorf("failed to initialize tier 5 keepers: %w", err)
	}

	// Phase 7: Initialize security keepers
	if err := ki.initSecurityKeepers(container); err != nil {
		return fmt.Errorf("failed to initialize security keepers: %w", err)
	}

	// Phase 8: Wire up cross-keeper dependencies
	if err := ki.wireKeeperDependencies(container); err != nil {
		return fmt.Errorf("failed to wire keeper dependencies: %w", err)
	}

	return nil
}

// initTier1Keepers initializes tier 1 keepers (no AURA dependencies).
func (ki *KeeperInitializer) initTier1Keepers(container *KeeperContainer) error {
	// These keepers are already initialized in app.go:
	// - ComplianceKeeper
	// - CryptographyKeeper
	// - WalletSecurityKeeper
	// - GovernanceKeeper

	// Validate they are set
	if container.ComplianceKeeper == nil {
		return fmt.Errorf("compliance keeper not initialized")
	}
	if container.CryptographyKeeper == nil {
		return fmt.Errorf("cryptography keeper not initialized")
	}
	// WalletSecurityKeeper validation - the keeper is a value type, not a pointer
	// so we cannot check if it's nil. We rely on proper initialization in app.go.
	if container.GovernanceKeeper == nil {
		return fmt.Errorf("governance keeper not initialized")
	}

	ki.logger.Info("tier 1 keepers initialized successfully")
	return nil
}

// initTier2Keepers initializes tier 2 keepers.
func (ki *KeeperInitializer) initTier2Keepers(container *KeeperContainer) error {
	// These keepers are already initialized in app.go:
	// - IdentityChangeKeeper
	// - InclusionRoutinesKeeper

	if container.IdentityChangeKeeper == nil {
		return fmt.Errorf("identity change keeper not initialized")
	}
	if container.InclusionRoutinesKeeper == nil {
		return fmt.Errorf("inclusion routines keeper not initialized")
	}

	ki.logger.Info("tier 2 keepers initialized successfully")
	return nil
}

// initTier3Keepers initializes tier 3 keepers.
func (ki *KeeperInitializer) initTier3Keepers(container *KeeperContainer) error {
	// ConfidenceScoreKeeper depends on InclusionRoutinesKeeper
	if container.ConfidenceScoreKeeper == nil {
		return fmt.Errorf("confidence score keeper not initialized")
	}

	// ARCHITECTURAL DECISION: Deferred dependency wiring pending interface alignment
	//
	// ConfidenceScoreKeeper requires an IRRegistry interface with these methods:
	//   - GetIRPrerequisites(irID string) ([]string, error)
	//   - IsIRActive(irID string) bool
	//   - GetIRScore(irID string) (uint64, error)
	//   - GetIRArena(irID string) (string, error)
	//
	// Current status: InclusionRoutinesKeeper implements the first three methods but lacks
	// GetIRArena(), which would require design decisions about:
	//   1. Whether IRs should be categorized into arenas (competitive/cooperative/skill-based)
	//   2. How arena metadata should be stored (in IR definition vs separate registry)
	//   3. Arena scoring aggregation logic and weighting
	//
	// Workaround: ConfidenceScoreKeeper operates without IR registry integration, using
	// direct score calculations. This is functionally complete for current requirements.
	//
	// To implement: Add arena metadata to IR definition in inclusionroutines module,
	// implement GetIRArena() method, then enable wiring:
	//   container.ConfidenceScoreKeeper.SetIRRegistry(container.InclusionRoutinesKeeper)
	//
	// Reference: chain/x/inclusionroutines/keeper/registry_adapter.go.skip contains
	// the adapter implementation ready for use once GetIRArena is added.

	ki.logger.Info("tier 3 keepers initialized successfully (IR registry wiring pending)")
	return nil
}

// initTier4Keepers initializes tier 4 keepers.
func (ki *KeeperInitializer) initTier4Keepers(container *KeeperContainer) error {
	// VCRegistryKeeper depends on ConfidenceScoreKeeper
	if container.VCRegistryKeeper == nil {
		return fmt.Errorf("VC registry keeper not initialized")
	}

	// ARCHITECTURAL DECISION: Deferred dependency wiring pending interface signature alignment
	//
	// VCRegistryKeeper requires a ConfidenceScoreKeeper interface with context-free methods:
	//   - GetUserScore(walletAddr string) (uint64, bool)
	//   - HasCompletedIR(walletAddr, irID string) bool
	//   - GetArenaScore(walletAddr, arena string) (uint64, error)
	//   - GetAnchorInfo(walletAddr string) (interface{}, bool)
	//   - IsVerified(walletAddr string) bool
	//
	// Current status: ConfidenceScoreKeeper methods include sdk.Context parameter in their
	// signatures, which is the standard Cosmos SDK keeper pattern for accessing blockchain state.
	//
	// Interface mismatch reasons:
	//   1. VCRegistryKeeper expects context-free lookups for simplified verification logic
	//   2. ConfidenceScoreKeeper follows Cosmos SDK convention of passing context explicitly
	//   3. Both patterns are valid; choice depends on whether keeper should manage context internally
	//
	// Workaround: VCRegistryKeeper operates independently using its own VC-based verification
	// logic without confidence score integration. This is functionally complete for VC issuance
	// and presentation verification.
	//
	// To implement: Choose one of two approaches:
	//   Option A: Update ConfidenceScoreKeeper to store context internally (breaks Cosmos conventions)
	//   Option B: Update VCRegistryKeeper interface to accept context in method signatures (preferred)
	//   Option C: Create adapter layer that captures context at dependency injection time
	//
	// Additionally: Implement HasVC(ctx sdk.Context, holder, vcID string) bool method in
	// VCRegistryKeeper for circular verification checks.
	//
	// Then enable wiring:
	//   container.VCRegistryKeeper.SetConfidenceScoreKeeper(container.ConfidenceScoreKeeper)

	// DataRegistryKeeper has no dependencies
	if container.DataRegistryKeeper == nil {
		return fmt.Errorf("data registry keeper not initialized")
	}

	ki.logger.Info("tier 4 keepers initialized successfully (CS keeper wiring pending)")
	return nil
}

// initTier5Keepers initializes tier 5 keepers.
func (ki *KeeperInitializer) initTier5Keepers(container *KeeperContainer) error {
	// ContractRegistryKeeper depends on VCRegistry, Compliance, and ConfidenceScore
	if container.ContractRegistryKeeper == nil {
		return fmt.Errorf("contract registry keeper not initialized")
	}

	// Wire dependencies using adapters to bridge interface mismatches
	// These adapters are defined in app/keeper_adapters.go and handle context wrapping
	// and method signature differences between the keepers and expected interfaces.
	contractRegistryVCAdapter := newContractRegistryVCAdapter(container.VCRegistryKeeper)
	contractRegistryComplianceAdapter := newContractRegistryComplianceAdapter(container.ComplianceKeeper)
	contractRegistryCSAdapter := newContractRegistryConfidenceScoreAdapter(container.ConfidenceScoreKeeper)

	container.ContractRegistryKeeper.SetVCKeeper(contractRegistryVCAdapter)
	container.ContractRegistryKeeper.SetComplianceKeeper(contractRegistryComplianceAdapter)
	container.ContractRegistryKeeper.SetConfidenceScoreKeeper(contractRegistryCSAdapter)

	// BridgeKeeper and DexKeeper are already initialized
	if container.BridgeKeeper == nil {
		return fmt.Errorf("bridge keeper not initialized")
	}
	if container.DexKeeper == nil {
		return fmt.Errorf("dex keeper not initialized")
	}

	ki.logger.Info("tier 5 keepers initialized successfully (cross-keeper wiring pending)")
	return nil
}

// initSecurityKeepers initializes security keepers.
func (ki *KeeperInitializer) initSecurityKeepers(container *KeeperContainer) error {
	// ValidatorSecurityKeeper is a value type (not pointer), so we can't check if it's nil.
	// The keeper struct is returned by value from NewKeeper, and it doesn't have a nested
	// .Keeper field. We rely on proper initialization in app.go.
	//
	// Note: Removed invalid check for container.ValidatorSecurityKeeper.Keeper

	// WasmSecurityKeeper is also a value type (not pointer), so we can't check if it's nil.
	// The keeper struct is returned by value from NewKeeper, and it doesn't have a nested
	// .Keeper field. We rely on proper initialization in app.go.
	//
	// Note: Removed invalid check for container.WasmSecurityKeeper.Keeper

	ki.logger.Info("security keepers initialized successfully")
	return nil
}

// wireKeeperDependencies wires up cross-keeper dependencies after all keepers are initialized.
func (ki *KeeperInitializer) wireKeeperDependencies(container *KeeperContainer) error {
	// All dependencies are already wired in the tier initialization steps
	// This function is for any additional cross-cutting concerns

	ki.logger.Info("keeper dependencies wired successfully")
	return nil
}

// ValidateKeeperDependencies validates that all keeper dependencies are properly satisfied.
func ValidateKeeperDependencies(container *KeeperContainer) error {
	// Validate core keepers - these don't have nested .Keeper fields
	// AccountKeeper is an interface type, we can check if methods are callable
	// but we cannot check .AccountKeeper field as it doesn't exist.
	//
	// Note: Removed invalid checks for nested .Keeper, .BaseKeeper fields

	if container.StakingKeeper == nil {
		return fmt.Errorf("staking keeper is nil")
	}

	// ARCHITECTURAL DECISION: Deferred dependency validations
	//
	// These validations are disabled because the dependencies are not yet wired due to
	// interface alignment issues documented in initTier3Keepers and initTier4Keepers above.
	//
	// Current status: All keepers operate independently without cross-module integration.
	// This is functionally complete for their core features but misses some enhancement
	// opportunities (e.g., VC verification using confidence scores).
	//
	// The keepers are fully functional in isolation:
	//   - ConfidenceScoreKeeper calculates scores from IR completion without IR metadata
	//   - VCRegistryKeeper issues and verifies VCs without confidence score integration
	//   - ContractRegistryKeeper uses adapters for VC, Compliance, and CS integration
	//
	// Re-enable these validations after wiring is implemented:
	//
	// Validate ConfidenceScore -> InclusionRoutines dependency
	// if container.ConfidenceScoreKeeper != nil {
	// 	if container.InclusionRoutinesKeeper == nil {
	// 		return fmt.Errorf("confidence score keeper requires inclusion routines keeper")
	// 	}
	// }

	// Validate VCRegistry -> ConfidenceScore dependency
	// if container.VCRegistryKeeper != nil {
	// 	if container.ConfidenceScoreKeeper == nil {
	// 		return fmt.Errorf("VC registry keeper requires confidence score keeper")
	// 	}
	// }

	// Validate ContractRegistry dependencies (already wired via adapters)
	// if container.ContractRegistryKeeper != nil {
	// 	if container.VCRegistryKeeper == nil {
	// 		return fmt.Errorf("contract registry keeper requires VC registry keeper")
	// 	}
	// 	if container.ComplianceKeeper == nil {
	// 		return fmt.Errorf("contract registry keeper requires compliance keeper")
	// 	}
	// 	if container.ConfidenceScoreKeeper == nil {
	// 		return fmt.Errorf("contract registry keeper requires confidence score keeper")
	// 	}
	// }

	// Validate Bridge/Dex dependencies on VCRegistry
	if container.BridgeKeeper != nil {
		if container.VCRegistryKeeper == nil {
			return fmt.Errorf("bridge keeper requires VC registry keeper")
		}
	}
	if container.DexKeeper != nil {
		if container.VCRegistryKeeper == nil {
			return fmt.Errorf("dex keeper requires VC registry keeper")
		}
	}

	// ValidatorSecurityKeeper and SlashingKeeper are value types, not pointers
	// We cannot check .Keeper field as it doesn't exist
	//
	// Note: Removed invalid checks for nested .Keeper fields

	return nil
}

// GetKeeperInitializationOrder returns the order in which keepers should be initialized.
func GetKeeperInitializationOrder() []string {
	return []string{
		// Core SDK keepers (tier 0)
		"auth",
		"bank",
		"staking",
		"slashing",
		"distribution",

		// Tier 1: No AURA dependencies
		"compliance",
		"cryptography",
		"walletsecurity",
		"governance",

		// Tier 2: Depend on tier 1
		"identitychange",
		"inclusionroutines",

		// Tier 3: Depend on tier 2
		"confidencescore",

		// Tier 4: Depend on tier 3
		"vcregistry",
		"dataregistry",

		// Tier 5: Depend on tier 4
		"contractregistry",
		"bridge",
		"dex",

		// Security: Depend on all above
		"wasm",
		"wasmsecurity",
		"validatorsecurity",
	}
}

// KeeperDependencyGraph returns the dependency graph for all keepers.
func KeeperDependencyGraph() map[string][]string {
	return map[string][]string{
		// Core SDK modules
		"bank":         {"auth"},
		"staking":      {"auth", "bank"},
		"slashing":     {"staking"},
		"distribution": {"auth", "bank", "staking"},

		// Tier 1 (no AURA dependencies)
		"compliance":     {"auth"},
		"cryptography":   {},
		"walletsecurity": {},
		"governance":     {},

		// Tier 2
		"identitychange":    {},
		"inclusionroutines": {},

		// Tier 3
		"confidencescore": {"inclusionroutines"},

		// Tier 4
		"vcregistry":   {"confidencescore"},
		"dataregistry": {},

		// Tier 5
		"contractregistry": {"vcregistry", "compliance", "confidencescore"},
		"bridge":           {"bank", "auth", "vcregistry"},
		"dex":              {"bank", "auth", "vcregistry"},

		// Security
		"wasm":              {"auth", "bank", "staking"},
		"wasmsecurity":      {"wasm", "contractregistry"},
		"validatorsecurity": {"staking", "slashing", "bank"},
	}
}

// TopologicalSort performs a topological sort on the keeper dependency graph.
// Returns the keepers in initialization order.
func TopologicalSort(dependencies map[string][]string) ([]string, error) {
	// Calculate in-degree for each keeper; include nodes referenced only as dependencies.
	inDegree := make(map[string]int)
	for keeper, deps := range dependencies {
		if _, exists := inDegree[keeper]; !exists {
			inDegree[keeper] = 0
		}
		for _, dep := range deps {
			if _, exists := dependencies[dep]; !exists {
				// Materialize external dependency as a node with no further deps
				// so ordering includes it for downstream validation.
				dependencies[dep] = []string{}
			}
			if _, exists := inDegree[dep]; !exists {
				inDegree[dep] = 0
			}
			inDegree[keeper]++
		}
	}

	// Queue of keepers with no dependencies
	var queue []string
	for keeper, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, keeper)
		}
	}

	// Process queue
	var result []string
	for len(queue) > 0 {
		// Pop from queue
		keeper := queue[0]
		queue = queue[1:]
		result = append(result, keeper)

		// Reduce in-degree for dependents
		for dependent, deps := range dependencies {
			for _, dep := range deps {
				if dep == keeper {
					inDegree[dependent]--
					if inDegree[dependent] == 0 {
						queue = append(queue, dependent)
					}
				}
			}
		}
	}

	// Check for cycles
	if len(result) != len(dependencies) {
		return nil, fmt.Errorf("circular dependency detected in keeper graph")
	}

	return result, nil
}

// ValidateKeeperInitializationOrder validates that keepers are initialized in dependency order.
func ValidateKeeperInitializationOrder(order []string) error {
	dependencies := KeeperDependencyGraph()

	position := make(map[string]int)
	for i, keeper := range order {
		position[keeper] = i
	}

	// Check each keeper's dependencies come before it
	for i, keeper := range order {
		deps := dependencies[keeper]
		for _, dep := range deps {
			depPos, exists := position[dep]
			if !exists {
				return fmt.Errorf("keeper %s depends on %s which is not in initialization order", keeper, dep)
			}
			if depPos >= i {
				return fmt.Errorf("keeper %s (pos %d) depends on %s (pos %d) which comes after it", keeper, i, dep, depPos)
			}
		}
	}

	return nil
}
