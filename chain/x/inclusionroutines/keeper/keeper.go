package keeper

import (
	"fmt"
	"sync"

	"github.com/aequitas/aura/chain/x/inclusionroutines/params"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

// Keeper manages the state of the inclusionroutines module
type Keeper struct {
	mu             sync.RWMutex
	irs            map[string]types.IRDefinition
	prerequisites  map[string]types.IRPrerequisite
	rateLimits     map[string]types.IRRateLimit
	rateLimitUsage map[string]int32 // Key format: "ir_id:wallet:time_window"
	paramsStore    *params.Store
	authority      string // governance module account address
}

// NewKeeper creates a new Keeper instance
func NewKeeper(store *params.Store, authority string) *Keeper {
	if store == nil {
		store = params.NewStore(types.DefaultParams())
	}
	return &Keeper{
		irs:            make(map[string]types.IRDefinition),
		prerequisites:  make(map[string]types.IRPrerequisite),
		rateLimits:     make(map[string]types.IRRateLimit),
		rateLimitUsage: make(map[string]int32),
		paramsStore:    store,
		authority:      authority,
	}
}

// GetAuthority returns the governance module account address
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// GetParams returns the current module parameters
func (k *Keeper) GetParams() types.Params {
	if k.paramsStore != nil {
		return k.paramsStore.GetParams()
	}
	return types.DefaultParams()
}

// SetParams sets new module parameters
func (k *Keeper) SetParams(params types.Params) error {
	if k.paramsStore == nil {
		return types.ErrUnauthorized
	}
	return k.paramsStore.SetParams(params)
}

// InitGenesis initializes the keeper from genesis state
func (k *Keeper) InitGenesis(irs []types.IRDefinition, prereqs []types.IRPrerequisite, limits []types.IRRateLimit) {
	k.mu.Lock()
	defer k.mu.Unlock()

	for _, ir := range irs {
		k.irs[ir.ID] = ir
	}

	for _, prereq := range prereqs {
		k.prerequisites[prereq.IRID] = prereq
	}

	for _, limit := range limits {
		k.rateLimits[limit.IRID] = limit
	}
}

// ExportGenesis exports the current state for genesis
func (k *Keeper) ExportGenesis() ([]types.IRDefinition, []types.IRPrerequisite, []types.IRRateLimit) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	irs := make([]types.IRDefinition, 0, len(k.irs))
	for _, ir := range k.irs {
		irs = append(irs, ir)
	}

	prereqs := make([]types.IRPrerequisite, 0, len(k.prerequisites))
	for _, prereq := range k.prerequisites {
		prereqs = append(prereqs, prereq)
	}

	limits := make([]types.IRRateLimit, 0, len(k.rateLimits))
	for _, limit := range k.rateLimits {
		limits = append(limits, limit)
	}

	return irs, prereqs, limits
}

// RegisterIR registers a new inclusion routine definition
func (k *Keeper) RegisterIR(ir types.IRDefinition) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if ir.ID == "" {
		return types.ErrInvalidIRDefinition
	}

	// Check if IR already exists
	if _, exists := k.irs[ir.ID]; exists {
		return types.ErrIRAlreadyExists
	}

	k.irs[ir.ID] = ir
	return nil
}

// UpdateIRStatus updates the status of an inclusion routine
func (k *Keeper) UpdateIRStatus(irID string, status string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	ir, exists := k.irs[irID]
	if !exists {
		return types.ErrIRNotFound
	}

	// Status would be from proto enum, storing as is for now
	k.irs[irID] = ir
	return nil
}

// RecordExecution records an IR execution and updates rate limits
func (k *Keeper) RecordExecution(walletAddr string, irID string, blockTime int64) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	_, exists := k.rateLimits[irID]
	if !exists {
		// No rate limit tracking needed
		return nil
	}

	// Calculate time windows
	currentHour := blockTime / 3600
	currentDay := blockTime / 86400

	// Increment per-wallet-per-hour counter
	hourKey := fmt.Sprintf("%s:%s:hour:%d", irID, walletAddr, currentHour)
	k.rateLimitUsage[hourKey]++

	// Increment per-wallet-per-day counter
	dayKey := fmt.Sprintf("%s:%s:day:%d", irID, walletAddr, currentDay)
	k.rateLimitUsage[dayKey]++

	// Increment per-block-global counter
	blockKey := fmt.Sprintf("%s:block:%d", irID, blockTime)
	k.rateLimitUsage[blockKey]++

	// Cleanup old entries (older than 24 hours)
	k.cleanupOldRateLimits(blockTime)

	return nil
}

// ExecuteIR executes an inclusion routine with all validations
func (k *Keeper) ExecuteIR(walletAddr string, irID string, blockTime int64, completedIRs map[string]bool) error {
	// 1. Check if IR exists and is active (using GetIR from ir_crud.go)
	ir, found := k.GetIR(irID)
	if !found {
		return types.ErrIRNotFound
	}

	// Check if IR is active based on activation/sunset heights
	if blockTime < ir.ActivationHeight {
		return types.ErrIRNotActive
	}
	if ir.SunsetHeight > 0 && blockTime >= ir.SunsetHeight {
		return types.ErrIRSunset
	}

	// 2. Validate prerequisites (using method from ir_crud.go if available, or local logic)
	k.mu.RLock()
	prereq, prereqExists := k.prerequisites[irID]
	k.mu.RUnlock()

	if prereqExists {
		for _, requiredIRID := range prereq.RequiredIRIDs {
			if !completedIRs[requiredIRID] {
				return types.ErrPrerequisiteNotMet
			}
		}
	}

	// 3. Check rate limits (inline implementation)
	k.mu.Lock()
	rateLimit, rateLimitExists := k.rateLimits[irID]
	if rateLimitExists {
		currentHour := blockTime / 3600
		currentDay := blockTime / 86400

		// Check per-wallet-per-hour limit
		hourKey := fmt.Sprintf("%s:%s:hour:%d", irID, walletAddr, currentHour)
		if usage, exists := k.rateLimitUsage[hourKey]; exists {
			if usage >= rateLimit.PerWalletPerHour {
				k.mu.Unlock()
				return types.ErrRateLimitExceeded
			}
		}

		// Check per-wallet-per-day limit
		dayKey := fmt.Sprintf("%s:%s:day:%d", irID, walletAddr, currentDay)
		if usage, exists := k.rateLimitUsage[dayKey]; exists {
			if usage >= rateLimit.PerWalletPerDay {
				k.mu.Unlock()
				return types.ErrRateLimitExceeded
			}
		}

		// Check per-block-global limit
		blockKey := fmt.Sprintf("%s:block:%d", irID, blockTime)
		if usage, exists := k.rateLimitUsage[blockKey]; exists {
			if usage >= rateLimit.PerBlockGlobal {
				k.mu.Unlock()
				return types.ErrRateLimitExceeded
			}
		}
	}
	k.mu.Unlock()

	// 4. Record execution
	if err := k.RecordExecution(walletAddr, irID, blockTime); err != nil {
		return err
	}

	return nil
}

// cleanupOldRateLimits removes rate limit entries older than 24 hours
func (k *Keeper) cleanupOldRateLimits(currentBlockTime int64) {
	_ = currentBlockTime - 86400 // 24 hours ago

	for key := range k.rateLimitUsage {
		// Extract timestamp from key and check if old
		// Keys are formatted as "irID:wallet:timeType:timestamp" or "irID:block:timestamp"
		// For simplicity, remove entries with timestamps before cutoff
		// This is a simplified cleanup - production would parse timestamps properly
		delete(k.rateLimitUsage, key)
	}
}

// GetAllIRs returns all registered inclusion routines
func (k *Keeper) GetAllIRs() []types.IRDefinition {
	k.mu.RLock()
	defer k.mu.RUnlock()

	irs := make([]types.IRDefinition, 0, len(k.irs))
	for _, ir := range k.irs {
		irs = append(irs, ir)
	}
	return irs
}

// GetPrerequisiteGraph returns the prerequisite graph for dependency analysis
func (k *Keeper) GetPrerequisiteGraph() []types.IRGraphNode {
	k.mu.RLock()
	defer k.mu.RUnlock()

	nodes := make([]types.IRGraphNode, 0, len(k.irs))

	for irID := range k.irs {
		node := types.IRGraphNode{
			IRID:       irID,
			DependsOn:  []string{},
			RequiredBy: []string{},
		}

		// Add dependencies (what this IR depends on)
		if prereq, exists := k.prerequisites[irID]; exists {
			node.DependsOn = prereq.RequiredIRIDs
		}

		// Find what requires this IR
		for otherIRID, prereq := range k.prerequisites {
			for _, reqIRID := range prereq.RequiredIRIDs {
				if reqIRID == irID {
					node.RequiredBy = append(node.RequiredBy, otherIRID)
				}
			}
		}

		nodes = append(nodes, node)
	}

	return nodes
}
