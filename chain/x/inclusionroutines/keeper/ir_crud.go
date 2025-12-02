package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CreateIR creates a new IR definition in KV store
func (k *Keeper) CreateIR(ctx sdk.Context, ir types.IRDefinition) error {
	// Check if IR already exists
	if _, exists := k.GetIR(ctx, ir.Id); exists {
		return types.ErrIRAlreadyExists
	}

	if err := k.validateIR(ir); err != nil {
		return err
	}

	// Set initial status if not set
	if ir.Status == 0 {
		ir.Status = inclusionroutinespb.IRStatus_IR_STATUS_DRAFT
	}

	// Register the IR
	return k.RegisterIR(ctx, ir)
}

// UpdateIR updates an existing IR definition in KV store
func (k *Keeper) UpdateIR(ctx sdk.Context, ir types.IRDefinition) error {
	// Check if IR exists
	existing, exists := k.GetIR(ctx, ir.Id)
	if !exists {
		return types.ErrIRNotFound
	}

	if err := k.validateIR(ir); err != nil {
		return err
	}

	// Preserve certain immutable fields from the existing IR
	ir.Arena = existing.Arena                       // Arena cannot be changed after creation
	ir.ActivationHeight = existing.ActivationHeight // Activation height cannot be changed

	return k.RegisterIR(ctx, ir)
}

// DeleteIR removes an IR definition from KV store
func (k *Keeper) DeleteIR(ctx sdk.Context, id string) error {
	// Check if IR exists
	if _, exists := k.GetIR(ctx, id); !exists {
		return types.ErrIRNotFound
	}

	// Check if any other IRs depend on this one
	// Must iterate to check dependencies
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.IRStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, prefixEndBytes(prefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var checkIR types.IRDefinition
		if err := k.cdc.Unmarshal(iterator.Value(), &checkIR); err != nil {
			continue
		}

		if prereq, exists := k.GetPrerequisite(ctx, checkIR.Id); exists {
			for _, reqID := range prereq.RequiredIrIds {
				if reqID == id {
					return fmt.Errorf("cannot delete IR %s: required by %s", id, checkIR.Id)
				}
			}
		}
	}

	// Delete from store
	if err := store.Delete([]byte(types.IRStoreKey(id))); err != nil {
		return fmt.Errorf("failed to delete IR: %w", err)
	}

	// Delete related data
	store.Delete([]byte(types.PrerequisiteStoreKey(id)))
	store.Delete([]byte(types.RateLimitStoreKey(id)))

	return nil
}

// ListIRs returns a filtered list of IR definitions with pagination
// Filters at storage iteration level instead of loading all IRs into memory
func (k *Keeper) ListIRs(ctx sdk.Context, statusFilter inclusionroutinespb.IRStatus, arenaFilter inclusionroutinespb.Arena, localeFilter string, offset, limit int) ([]types.IRDefinition, int) {
	if limit <= 0 {
		limit = 100 // Default page size
	}
	if offset < 0 {
		offset = 0
	}

	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.IRStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, prefixEndBytes(prefix))
	if err != nil {
		k.logger.Error("failed to create IR iterator", "error", err)
		return []types.IRDefinition{}, 0
	}
	defer iterator.Close()

	matching := make([]types.IRDefinition, 0, limit)
	totalCount := 0
	skipped := 0

	for ; iterator.Valid(); iterator.Next() {
		var ir types.IRDefinition
		if err := k.cdc.Unmarshal(iterator.Value(), &ir); err != nil {
			k.logger.Error("failed to unmarshal IR in ListIRs", "error", err)
			continue
		}

		// Apply filters at iteration level
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

		// This IR matches filters
		totalCount++

		// Skip until we reach offset
		if skipped < offset {
			skipped++
			continue
		}

		// Stop if we've collected enough for this page
		if len(matching) >= limit {
			continue
		}

		matching = append(matching, ir)
	}

	return matching, totalCount
}

// SuspendIR suspends an IR
func (k *Keeper) SuspendIR(ctx sdk.Context, id string) error {
	ir, exists := k.GetIR(ctx, id)
	if !exists {
		return types.ErrIRNotFound
	}

	if ir.Status == inclusionroutinespb.IRStatus_IR_STATUS_SUSPENDED {
		return fmt.Errorf("IR %s is already suspended", id)
	}

	ir.Status = inclusionroutinespb.IRStatus_IR_STATUS_SUSPENDED
	return k.RegisterIR(ctx, ir)
}

// ActivateIR activates a suspended or approved IR
func (k *Keeper) ActivateIR(ctx sdk.Context, id string) error {
	ir, exists := k.GetIR(ctx, id)
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
	return k.RegisterIR(ctx, ir)
}

// RetireIR retires an IR permanently
func (k *Keeper) RetireIR(ctx sdk.Context, id string) error {
	ir, exists := k.GetIR(ctx, id)
	if !exists {
		return types.ErrIRNotFound
	}

	if ir.Status == inclusionroutinespb.IRStatus_IR_STATUS_RETIRED {
		return fmt.Errorf("IR %s is already retired", id)
	}

	ir.Status = inclusionroutinespb.IRStatus_IR_STATUS_RETIRED
	return k.RegisterIR(ctx, ir)
}

// GetIRByArena returns all IRs in a specific arena
// Filters at storage iteration level for efficiency
func (k *Keeper) GetIRByArena(ctx sdk.Context, arena inclusionroutinespb.Arena) []types.IRDefinition {
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.IRStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, prefixEndBytes(prefix))
	if err != nil {
		k.logger.Error("failed to create IR iterator", "error", err)
		return []types.IRDefinition{}
	}
	defer iterator.Close()

	result := make([]types.IRDefinition, 0)
	for ; iterator.Valid(); iterator.Next() {
		var ir types.IRDefinition
		if err := k.cdc.Unmarshal(iterator.Value(), &ir); err != nil {
			k.logger.Error("failed to unmarshal IR", "error", err)
			continue
		}

		if ir.Arena == arena {
			result = append(result, ir)
		}
	}

	return result
}

// GetActiveIRs returns all active IRs at the current block height
// Filters at storage iteration level for efficiency
func (k *Keeper) GetActiveIRs(ctx sdk.Context) []types.IRDefinition {
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.IRStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, prefixEndBytes(prefix))
	if err != nil {
		k.logger.Error("failed to create IR iterator", "error", err)
		return []types.IRDefinition{}
	}
	defer iterator.Close()

	result := make([]types.IRDefinition, 0)
	currentHeight := ctx.BlockHeight()

	for ; iterator.Valid(); iterator.Next() {
		var ir types.IRDefinition
		if err := k.cdc.Unmarshal(iterator.Value(), &ir); err != nil {
			k.logger.Error("failed to unmarshal IR", "error", err)
			continue
		}

		// Check if IR is active
		if ir.Status == inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE {
			// Check activation height
			if currentHeight >= ir.ActivationHeight {
				// Check sunset height
				if ir.SunsetHeight == 0 || currentHeight < ir.SunsetHeight {
					result = append(result, ir)
				}
			}
		}
	}

	return result
}

// validateIR performs comprehensive validation on an IR definition
func (k *Keeper) validateIR(ir types.IRDefinition) error {
	if ir.Id == "" {
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

	if ir.PoiReward < 0 {
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

// GetIRCount returns the total number of registered IRs
func (k *Keeper) GetIRCount(ctx sdk.Context) int {
	return len(k.GetAllIRs(ctx))
}

// GetIRsByStatus returns all IRs with a specific status
// Filters at storage iteration level for efficiency
func (k *Keeper) GetIRsByStatus(ctx sdk.Context, status inclusionroutinespb.IRStatus) []types.IRDefinition {
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.IRStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, prefixEndBytes(prefix))
	if err != nil {
		k.logger.Error("failed to create IR iterator", "error", err)
		return []types.IRDefinition{}
	}
	defer iterator.Close()

	result := make([]types.IRDefinition, 0)
	for ; iterator.Valid(); iterator.Next() {
		var ir types.IRDefinition
		if err := k.cdc.Unmarshal(iterator.Value(), &ir); err != nil {
			k.logger.Error("failed to unmarshal IR", "error", err)
			continue
		}

		if ir.Status == status {
			result = append(result, ir)
		}
	}

	return result
}
