package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/inclusionroutines/params"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

// prefixEndBytes returns a []byte that would end a
// range query for all []byte with a certain prefix
// Copied from SDK since PrefixEndBytes was removed
func prefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)

	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end
		}
	}

	return nil
}

// Keeper manages the state of the inclusionroutines module using persistent KV store.
// All state is stored deterministically in the KV store to ensure consensus safety.
type Keeper struct {
	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	paramsStore  *params.Store
	authority    string // governance module account address
	logger       log.Logger
}

// NewKeeper creates a new Keeper instance with persistent KV store.
// All state is persisted to the KV store - no in-memory maps are used.
func NewKeeper(
	storeService store.KVStoreService,
	cdc codec.BinaryCodec,
	paramsStore *params.Store,
	authority string,
	logger log.Logger,
) *Keeper {
	if paramsStore == nil {
		paramsStore = params.NewStore(types.DefaultParams())
	}

	return &Keeper{
		storeService: storeService,
		cdc:          cdc,
		paramsStore:  paramsStore,
		authority:    authority,
		logger:       logger,
	}
}

// GetAuthority returns the governance module account address
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// GetParams returns the current module parameters
func (k *Keeper) GetParams() types.Params {
	return k.paramsStore.GetParams()
}

// SetParams sets new module parameters
func (k *Keeper) SetParams(params types.Params) error {
	if k.paramsStore == nil {
		return fmt.Errorf("params store not initialized")
	}
	return k.paramsStore.SetParams(params)
}

// ============================
// IR DEFINITION MANAGEMENT
// ============================

// GetIR retrieves an IR definition from KV store
func (k *Keeper) GetIR(ctx sdk.Context, irID string) (types.IRDefinition, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte(types.IRStoreKey(irID)))
	if err != nil || bz == nil {
		return types.IRDefinition{}, false
	}

	var ir types.IRDefinition
	if err := k.cdc.Unmarshal(bz, &ir); err != nil {
		return types.IRDefinition{}, false
	}

	return ir, true
}

// RegisterIR registers a new inclusion routine definition in KV store
func (k *Keeper) RegisterIR(ctx sdk.Context, ir types.IRDefinition) error {
	if ir.Id == "" {
		return types.ErrInvalidIRDefinition
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&ir)
	if err != nil {
		return fmt.Errorf("failed to marshal IR definition: %w", err)
	}

	if err := store.Set([]byte(types.IRStoreKey(ir.Id)), bz); err != nil {
		return fmt.Errorf("failed to store IR definition: %w", err)
	}

	return nil
}

// UpdateIRStatus updates the status of an inclusion routine in KV store
func (k *Keeper) UpdateIRStatus(ctx sdk.Context, irID string, status string) error {
	ir, exists := k.GetIR(ctx, irID)
	if !exists {
		return types.ErrIRNotFound
	}

	// Status would be from proto enum, storing as is for now
	// Update status field if it exists in IRDefinition
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&ir)
	if err != nil {
		return fmt.Errorf("failed to marshal IR definition: %w", err)
	}

	if err := store.Set([]byte(types.IRStoreKey(irID)), bz); err != nil {
		return fmt.Errorf("failed to update IR status: %w", err)
	}

	return nil
}

// GetAllIRs returns all registered inclusion routines from KV store
func (k *Keeper) GetAllIRs(ctx sdk.Context) []types.IRDefinition {
	store := k.storeService.OpenKVStore(ctx)

	prefix := []byte(types.IRStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, prefixEndBytes(prefix))
	if err != nil {
		return []types.IRDefinition{}
	}
	defer iterator.Close()

	irs := []types.IRDefinition{}
	for ; iterator.Valid(); iterator.Next() {
		var ir types.IRDefinition
		if err := k.cdc.Unmarshal(iterator.Value(), &ir); err != nil {
			continue
		}
		irs = append(irs, ir)
	}

	return irs
}

// ============================
// PREREQUISITE MANAGEMENT
// ============================

// GetPrerequisite retrieves IR prerequisites from KV store
func (k *Keeper) GetPrerequisite(ctx sdk.Context, irID string) (types.IRPrerequisite, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte(types.PrerequisiteStoreKey(irID)))
	if err != nil || bz == nil {
		return types.IRPrerequisite{}, false
	}

	var prereq types.IRPrerequisite
	if err := k.cdc.Unmarshal(bz, &prereq); err != nil {
		return types.IRPrerequisite{}, false
	}

	return prereq, true
}

// SetPrerequisite stores IR prerequisites to KV store
func (k *Keeper) SetPrerequisite(ctx sdk.Context, prereq types.IRPrerequisite) error {
	if prereq.IrId == "" {
		return types.ErrInvalidIRDefinition
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&prereq)
	if err != nil {
		return fmt.Errorf("failed to marshal prerequisite: %w", err)
	}

	if err := store.Set([]byte(types.PrerequisiteStoreKey(prereq.IrId)), bz); err != nil {
		return fmt.Errorf("failed to store prerequisite: %w", err)
	}

	return nil
}

// ============================
// RATE LIMIT MANAGEMENT
// ============================

// GetRateLimit retrieves IR rate limit configuration from KV store
func (k *Keeper) GetRateLimit(ctx sdk.Context, irID string) (types.IRRateLimit, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte(types.RateLimitStoreKey(irID)))
	if err != nil || bz == nil {
		return types.IRRateLimit{}, false
	}

	var rateLimit types.IRRateLimit
	if err := k.cdc.Unmarshal(bz, &rateLimit); err != nil {
		return types.IRRateLimit{}, false
	}

	return rateLimit, true
}

// SetRateLimit stores IR rate limit configuration to KV store
func (k *Keeper) SetRateLimit(ctx sdk.Context, rateLimit types.IRRateLimit) error {
	if rateLimit.IrId == "" {
		return types.ErrInvalidIRDefinition
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&rateLimit)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit: %w", err)
	}

	if err := store.Set([]byte(types.RateLimitStoreKey(rateLimit.IrId)), bz); err != nil {
		return fmt.Errorf("failed to store rate limit: %w", err)
	}

	return nil
}

// GetRateLimitUsage retrieves rate limit usage counter from KV store
func (k *Keeper) GetRateLimitUsage(ctx sdk.Context, irID, wallet, timeWindow string) int32 {
	store := k.storeService.OpenKVStore(ctx)
	key := types.RateLimitUsageKey(irID, wallet, timeWindow)

	bz, err := store.Get([]byte(key))
	if err != nil || bz == nil {
		return 0
	}

	// Simple counter stored as int32 (4 bytes)
	if len(bz) != 4 {
		return 0
	}

	return int32(bz[0]) | int32(bz[1])<<8 | int32(bz[2])<<16 | int32(bz[3])<<24
}

// IncrementRateLimitUsage increments a rate limit usage counter in KV store
func (k *Keeper) IncrementRateLimitUsage(ctx sdk.Context, irID, wallet, timeWindow string) error {
	current := k.GetRateLimitUsage(ctx, irID, wallet, timeWindow)
	newCount := current + 1

	store := k.storeService.OpenKVStore(ctx)
	key := types.RateLimitUsageKey(irID, wallet, timeWindow)

	// Encode as 4-byte little-endian
	bz := []byte{
		byte(newCount),
		byte(newCount >> 8),
		byte(newCount >> 16),
		byte(newCount >> 24),
	}

	if err := store.Set([]byte(key), bz); err != nil {
		return fmt.Errorf("failed to increment rate limit usage: %w", err)
	}

	return nil
}

// ============================
// IR EXECUTION
// ============================

// RecordExecution records an IR execution and updates rate limits in KV store
func (k *Keeper) RecordExecution(ctx sdk.Context, walletAddr string, irID string) error {
	_, exists := k.GetRateLimit(ctx, irID)
	if !exists {
		// No rate limit tracking needed
		return nil
	}

	blockTime := ctx.BlockTime().Unix()

	// Calculate time windows
	currentHour := blockTime / 3600
	currentDay := blockTime / 86400

	// Increment per-wallet-per-hour counter
	hourKey := fmt.Sprintf("hour:%d", currentHour)
	if err := k.IncrementRateLimitUsage(ctx, irID, walletAddr, hourKey); err != nil {
		return fmt.Errorf("failed to Sprintf: %w", err)
	}

	// Increment per-wallet-per-day counter
	dayKey := fmt.Sprintf("day:%d", currentDay)
	if err := k.IncrementRateLimitUsage(ctx, irID, walletAddr, dayKey); err != nil {
		return fmt.Errorf("failed to Sprintf: %w", err)
	}

	// Increment per-block-global counter
	blockKey := fmt.Sprintf("block:%d", blockTime)
	if err := k.IncrementRateLimitUsage(ctx, irID, "global", blockKey); err != nil {
		return fmt.Errorf("failed to Sprintf: %w", err)
	}

	return nil
}

// ExecuteIR executes an inclusion routine with all validations
func (k *Keeper) ExecuteIR(ctx sdk.Context, walletAddr string, irID string, completedIRs map[string]bool) error {
	// 1. Check if IR exists and is active
	ir, found := k.GetIR(ctx, irID)
	if !found {
		return types.ErrIRNotFound
	}

	blockHeight := ctx.BlockHeight()

	// Check if IR is active based on activation/sunset heights
	if blockHeight < ir.ActivationHeight {
		return types.ErrIRNotActive
	}
	if ir.SunsetHeight > 0 && blockHeight >= ir.SunsetHeight {
		return types.ErrIRSunset
	}

	// 2. Validate prerequisites
	prereq, prereqExists := k.GetPrerequisite(ctx, irID)
	if prereqExists {
		for _, requiredIRID := range prereq.RequiredIrIds {
			if !completedIRs[requiredIRID] {
				return types.ErrPrerequisiteNotMet
			}
		}
	}

	// 3. Check rate limits
	rateLimit, rateLimitExists := k.GetRateLimit(ctx, irID)
	if rateLimitExists {
		blockTime := ctx.BlockTime().Unix()
		currentHour := blockTime / 3600
		currentDay := blockTime / 86400

		// Check per-wallet-per-hour limit
		hourKey := fmt.Sprintf("hour:%d", currentHour)
		hourUsage := k.GetRateLimitUsage(ctx, irID, walletAddr, hourKey)
		if hourUsage >= rateLimit.PerWalletPerHour {
			return types.ErrRateLimitExceeded
		}

		// Check per-wallet-per-day limit
		dayKey := fmt.Sprintf("day:%d", currentDay)
		dayUsage := k.GetRateLimitUsage(ctx, irID, walletAddr, dayKey)
		if dayUsage >= rateLimit.PerWalletPerDay {
			return types.ErrRateLimitExceeded
		}

		// Check per-block-global limit
		blockKey := fmt.Sprintf("block:%d", blockTime)
		blockUsage := k.GetRateLimitUsage(ctx, irID, "global", blockKey)
		if blockUsage >= rateLimit.PerBlockGlobal {
			return types.ErrRateLimitExceeded
		}
	}

	// 4. Record execution
	if err := k.RecordExecution(ctx, walletAddr, irID); err != nil {
		return fmt.Errorf("operation failed: %w", err)
	}

	return nil
}

// ============================
// PREREQUISITE GRAPH
// ============================

// GetPrerequisiteGraph returns the prerequisite graph for dependency analysis
func (k *Keeper) GetPrerequisiteGraph(ctx sdk.Context) []types.IRGraphNode {
	store := k.storeService.OpenKVStore(ctx)

	// Get all IRs
	prefix := []byte(types.IRStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, prefixEndBytes(prefix))
	if err != nil {
		return []types.IRGraphNode{}
	}
	defer iterator.Close()

	// Build IR ID list
	irIDs := []string{}
	for ; iterator.Valid(); iterator.Next() {
		var ir types.IRDefinition
		if err := k.cdc.Unmarshal(iterator.Value(), &ir); err != nil {
			continue
		}
		irIDs = append(irIDs, ir.Id)
	}

	// Build graph nodes
	nodes := []types.IRGraphNode{}
	for _, irID := range irIDs {
		node := types.IRGraphNode{
			IrId:       irID,
			DependsOn:  []string{},
			RequiredBy: []string{},
		}

		// Add dependencies (what this IR depends on)
		if prereq, exists := k.GetPrerequisite(ctx, irID); exists {
			node.DependsOn = prereq.RequiredIrIds
		}

		// Find what requires this IR
		for _, otherIRID := range irIDs {
			if otherIRID == irID {
				continue
			}
			if prereq, exists := k.GetPrerequisite(ctx, otherIRID); exists {
				for _, reqIRID := range prereq.RequiredIrIds {
					if reqIRID == irID {
						node.RequiredBy = append(node.RequiredBy, otherIRID)
					}
				}
			}
		}

		nodes = append(nodes, node)
	}

	return nodes
}

// ============================
// GENESIS MANAGEMENT
// ============================

// InitGenesis initializes the keeper from genesis state
func (k *Keeper) InitGenesis(ctx sdk.Context, genesis types.GenesisState) error {
	// Validate incoming state before mutating store
	if err := types.ValidateGenesisState(&genesis); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	// Set params
	if err := k.paramsStore.SetParams(genesis.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	for _, ir := range genesis.Irs {
		if ir == nil {
			continue
		}
		if err := k.RegisterIR(ctx, *ir); err != nil {
			return fmt.Errorf("failed to import IR %s: %w", ir.Id, err)
		}
	}

	for _, prereq := range genesis.Prerequisites {
		if prereq == nil {
			continue
		}
		if err := k.SetPrerequisite(ctx, *prereq); err != nil {
			return fmt.Errorf("failed to import prerequisite for %s: %w", prereq.IrId, err)
		}
	}

	for _, rateLimit := range genesis.RateLimits {
		if rateLimit == nil {
			continue
		}
		if err := k.SetRateLimit(ctx, *rateLimit); err != nil {
			return fmt.Errorf("failed to import rate limit for %s: %w", rateLimit.IrId, err)
		}
	}

	return nil
}

// ExportGenesis exports the current state for genesis
func (k *Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	store := k.storeService.OpenKVStore(ctx)

	// Export IR definitions
	irDefinitions := []*types.IRDefinition{}
	prefix := []byte(types.IRStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, prefixEndBytes(prefix))
	if err == nil {
		defer iterator.Close()
		for ; iterator.Valid(); iterator.Next() {
			var ir types.IRDefinition
			if err := k.cdc.Unmarshal(iterator.Value(), &ir); err == nil {
				irCopy := ir
				irDefinitions = append(irDefinitions, &irCopy)
			}
		}
	}

	// Export prerequisites
	prerequisites := []*types.IRPrerequisite{}
	prereqPrefix := []byte(types.PrerequisiteStoreKeyPrefix)
	prereqIterator, err := store.Iterator(prereqPrefix, prefixEndBytes(prereqPrefix))
	if err == nil {
		defer prereqIterator.Close()
		for ; prereqIterator.Valid(); prereqIterator.Next() {
			var prereq types.IRPrerequisite
			if err := k.cdc.Unmarshal(prereqIterator.Value(), &prereq); err == nil {
				prereqCopy := prereq
				prerequisites = append(prerequisites, &prereqCopy)
			}
		}
	}

	// Export rate limits
	rateLimits := []*types.IRRateLimit{}
	rateLimitPrefix := []byte(types.RateLimitStoreKeyPrefix)
	rateLimitIterator, err := store.Iterator(rateLimitPrefix, prefixEndBytes(rateLimitPrefix))
	if err == nil {
		defer rateLimitIterator.Close()
		for ; rateLimitIterator.Valid(); rateLimitIterator.Next() {
			var rateLimit types.IRRateLimit
			if err := k.cdc.Unmarshal(rateLimitIterator.Value(), &rateLimit); err == nil {
				rateLimitCopy := rateLimit
				rateLimits = append(rateLimits, &rateLimitCopy)
			}
		}
	}

	params := k.GetParams()
	return types.GenesisState{
		Params:        params,
		Irs:           irDefinitions,
		Prerequisites: prerequisites,
		RateLimits:    rateLimits,
	}
}
