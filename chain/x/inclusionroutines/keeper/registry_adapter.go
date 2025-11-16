package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

// GetIRPrerequisites returns the list of prerequisite IR IDs for the given IR
// This method is part of the IRRegistry interface used by the confidencescore module
func (k *Keeper) GetIRPrerequisites(irID string) ([]string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Check if IR exists
	if _, exists := k.irs[irID]; !exists {
		return nil, types.ErrIRNotFound
	}

	// Get prerequisites
	prereq, hasPrereqs := k.prerequisites[irID]
	if !hasPrereqs {
		return []string{}, nil // No prerequisites
	}

	return prereq.RequiredIRIDs, nil
}

// IsIRActive checks if an IR is currently active
// This method is part of the IRRegistry interface used by the confidencescore module
func (k *Keeper) IsIRActive(irID string) bool {
	k.mu.RLock()
	defer k.mu.RUnlock()

	ir, exists := k.irs[irID]
	if !exists {
		return false
	}

	return ir.Status == inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE
}

// GetIRScore returns the score value for completing an IR
// This method is part of the IRRegistry interface used by the confidencescore module
func (k *Keeper) GetIRScore(irID string) (uint64, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	ir, exists := k.irs[irID]
	if !exists {
		return 0, types.ErrIRNotFound
	}

	if ir.Score < 0 {
		return 0, fmt.Errorf("invalid score for IR %s: %d", irID, ir.Score)
	}

	return uint64(ir.Score), nil
}

// GetIRArena returns the arena type for an IR
// This method is part of the IRRegistry interface used by the confidencescore module
func (k *Keeper) GetIRArena(irID string) (string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	ir, exists := k.irs[irID]
	if !exists {
		return "", types.ErrIRNotFound
	}

	// Convert Arena enum to string
	arenaStr := ir.Arena.String()
	return arenaStr, nil
}
