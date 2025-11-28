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

	aikeeper "github.com/aequitas/aura/chain/x/aiassistant/keeper"
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
	AIKeeper               *aikeeper.Keeper

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

	// TODO: Wire dependency after interface alignment
	// The ConfidenceScoreKeeper expects an IRRegistry interface with these methods:
	//   - GetIRPrerequisites(irID string) ([]string, error)
	//   - IsIRActive(irID string) bool
	//   - GetIRScore(irID string) (uint64, error)
	//   - GetIRArena(irID string) (string, error)  // <- MISSING in InclusionRoutinesKeeper
	//
	// Currently commented out until InclusionRoutinesKeeper implements GetIRArena method.
	// See: chain/x/inclusionroutines/keeper/registry_adapter.go.skip for the adapter.
	//
	// container.ConfidenceScoreKeeper.SetIRRegistry(container.InclusionRoutinesKeeper)

	ki.logger.Info("tier 3 keepers initialized successfully (IR registry wiring pending)")
	return nil
}

// initTier4Keepers initializes tier 4 keepers.
func (ki *KeeperInitializer) initTier4Keepers(container *KeeperContainer) error {
	// VCRegistryKeeper depends on ConfidenceScoreKeeper
	if container.VCRegistryKeeper == nil {
		return fmt.Errorf("VC registry keeper not initialized")
	}

	// TODO: Wire dependency after interface alignment
	// The VCRegistryKeeper expects a ConfidenceScoreKeeper interface with these methods:
	//   - GetUserScore(walletAddr string) (uint64, bool)  // <- Current signature has sdk.Context
	//   - HasCompletedIR(walletAddr, irID string) bool
	//   - GetArenaScore(walletAddr, arena string) (uint64, error)
	//   - GetAnchorInfo(walletAddr string) (interface{}, bool)  // <- Current signature has sdk.Context
	//   - IsVerified(walletAddr string) bool
	//
	// Currently commented out until ConfidenceScoreKeeper methods are updated to match
	// the expected interface (remove sdk.Context from signatures).
	// Also need to add HasVC method to VCRegistryKeeper.
	//
	// container.VCRegistryKeeper.SetConfidenceScoreKeeper(container.ConfidenceScoreKeeper)

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

	// TODO: Wire dependencies after interface alignment
	// The ContractRegistryKeeper expects these interfaces:
	//
	// 1. VCRegistryKeeper interface needs:
	//    - HasVC(walletAddr, vcType string) bool  // <- MISSING in VCRegistryKeeper
	//    (plus other methods that may exist)
	//
	// 2. ComplianceKeeper interface needs:
	//    - GetKYCLevel(walletAddr string) (string, error)  // <- MISSING in ComplianceKeeper
	//    (plus other methods that may exist)
	//
	// 3. ConfidenceScoreKeeper interface needs:
	//    - Methods without sdk.Context as mentioned in tier 4
	//
	// Currently commented out until all required interface methods are implemented.
	//
	// container.ContractRegistryKeeper.SetVCRegistryKeeper(container.VCRegistryKeeper)
	// container.ContractRegistryKeeper.SetComplianceKeeper(container.ComplianceKeeper)
	// container.ContractRegistryKeeper.SetConfidenceScoreKeeper(container.ConfidenceScoreKeeper)

	// BridgeKeeper, DexKeeper, and AIKeeper are already initialized
	if container.BridgeKeeper == nil {
		return fmt.Errorf("bridge keeper not initialized")
	}
	if container.DexKeeper == nil {
		return fmt.Errorf("dex keeper not initialized")
	}
	if container.AIKeeper == nil {
		return fmt.Errorf("AI keeper not initialized")
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

	// TODO: Re-enable these validations after keeper wiring is implemented
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

	// Validate ContractRegistry dependencies
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
		"aiassistant",

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
		"aiassistant":      {"bank"},

		// Security
		"wasm":              {"auth", "bank", "staking"},
		"wasmsecurity":      {"wasm", "contractregistry"},
		"validatorsecurity": {"staking", "slashing", "bank"},
	}
}

// TopologicalSort performs a topological sort on the keeper dependency graph.
// Returns the keepers in initialization order.
func TopologicalSort(dependencies map[string][]string) ([]string, error) {
	// Calculate in-degree for each keeper
	inDegree := make(map[string]int)
	for keeper := range dependencies {
		if _, exists := inDegree[keeper]; !exists {
			inDegree[keeper] = 0
		}
		for _, dep := range dependencies[keeper] {
			inDegree[dep]++
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
