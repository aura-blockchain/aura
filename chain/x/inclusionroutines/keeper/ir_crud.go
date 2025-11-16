package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

// GetIR retrieves an IR definition by ID
func (k *Keeper) GetIR(id string) (types.IRDefinition, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	ir, ok := k.irs[id]
	return ir, ok
}

// SetIR stores an IR definition
func (k *Keeper) SetIR(ir types.IRDefinition) error {
	if err := k.validateIR(ir); err != nil {
		return err
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	k.irs[ir.ID] = ir
	return nil
}

// DeleteIR removes an IR definition
func (k *Keeper) DeleteIR(id string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if IR exists
	if _, exists := k.irs[id]; !exists {
		return types.ErrIRNotFound
	}

	// Check if any other IRs depend on this one
	for _, prereq := range k.prerequisites {
		for _, reqID := range prereq.RequiredIRIDs {
			if reqID == id {
				return fmt.Errorf("cannot delete IR %s: required by %s", id, prereq.IRID)
			}
		}
	}

	delete(k.irs, id)
	delete(k.prerequisites, id)
	delete(k.rateLimits, id)
	return nil
}

// ListIRs returns a filtered list of IR definitions with pagination
func (k *Keeper) ListIRs(statusFilter inclusionroutinespb.IRStatus, arenaFilter inclusionroutinespb.Arena, localeFilter string, offset, limit int) ([]types.IRDefinition, int) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Collect all matching IRs
	matching := make([]types.IRDefinition, 0)
	for _, ir := range k.irs {
		// Apply filters
		if statusFilter != inclusionroutinespb.IRStatus_IR_STATUS_UNSPECIFIED && ir.Status != statusFilter {
			continue
		}
		if arenaFilter != inclusionroutinespb.Arena_ARENA_UNSPECIFIED && ir.Arena != arenaFilter {
			continue
		}
		if localeFilter != "" {
			hasLocale := false
			for _, tag := range ir.LocaleTags {
				if tag == localeFilter {
					hasLocale = true
					break
				}
			}
			if !hasLocale {
				continue
			}
		}
		matching = append(matching, ir)
	}

	total := len(matching)

	// Apply pagination
	if offset < 0 {
		offset = 0
	}
	if offset > len(matching) {
		return []types.IRDefinition{}, total
	}

	end := offset + limit
	if limit <= 0 || end > len(matching) {
		end = len(matching)
	}

	return matching[offset:end], total
}

// CreateIR creates a new IR definition
func (k *Keeper) CreateIR(ir types.IRDefinition) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if IR already exists
	if _, exists := k.irs[ir.ID]; exists {
		return types.ErrIRAlreadyExists
	}

	if err := k.validateIR(ir); err != nil {
		return err
	}

	// Set initial status if not set
	if ir.Status == inclusionroutinespb.IRStatus_IR_STATUS_UNSPECIFIED {
		ir.Status = inclusionroutinespb.IRStatus_IR_STATUS_DRAFT
	}

	k.irs[ir.ID] = ir
	return nil
}

// UpdateIR updates an existing IR definition
func (k *Keeper) UpdateIR(ir types.IRDefinition) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if IR exists
	existing, exists := k.irs[ir.ID]
	if !exists {
		return types.ErrIRNotFound
	}

	if err := k.validateIR(ir); err != nil {
		return err
	}

	// Preserve certain fields from the existing IR
	ir.Arena = existing.Arena                       // Arena cannot be changed after creation
	ir.ActivationHeight = existing.ActivationHeight // Activation height cannot be changed

	k.irs[ir.ID] = ir
	return nil
}

// SuspendIR suspends an IR
func (k *Keeper) SuspendIR(id string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	ir, exists := k.irs[id]
	if !exists {
		return types.ErrIRNotFound
	}

	if ir.Status == inclusionroutinespb.IRStatus_IR_STATUS_SUSPENDED {
		return fmt.Errorf("IR %s is already suspended", id)
	}

	ir.Status = inclusionroutinespb.IRStatus_IR_STATUS_SUSPENDED
	k.irs[id] = ir
	return nil
}

// ActivateIR activates a suspended or approved IR
func (k *Keeper) ActivateIR(id string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	ir, exists := k.irs[id]
	if !exists {
		return types.ErrIRNotFound
	}

	if ir.Status == inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE {
		return fmt.Errorf("IR %s is already active", id)
	}

	if ir.Status == inclusionroutinespb.IRStatus_IR_STATUS_RETIRED {
		return types.ErrIRRetired
	}

	ir.Status = inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE
	k.irs[id] = ir
	return nil
}

// validateIR performs validation on an IR definition
func (k *Keeper) validateIR(ir types.IRDefinition) error {
	if ir.ID == "" {
		return types.ErrIRInvalidID
	}

	if ir.Name == "" {
		return types.ErrInvalidName
	}

	if ir.Description == "" {
		return types.ErrInvalidDescription
	}

	if ir.Score < 0 {
		return types.ErrIRInvalidScore
	}

	if ir.POIReward < 0 {
		return types.ErrIRInvalidReward
	}

	if ir.Arena == inclusionroutinespb.Arena_ARENA_UNSPECIFIED {
		return types.ErrInvalidArena
	}

	if ir.PrivacyTier == inclusionroutinespb.PrivacyTier_PRIVACY_TIER_UNSPECIFIED {
		return types.ErrInvalidPrivacyTier
	}

	if len(ir.LocaleTags) == 0 {
		return types.ErrEmptyLocaleTag
	}

	if ir.SunsetHeight > 0 && ir.ActivationHeight > 0 && ir.SunsetHeight <= ir.ActivationHeight {
		return types.ErrSunsetBeforeActivation
	}

	return nil
}
