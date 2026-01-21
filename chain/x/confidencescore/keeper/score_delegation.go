// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"

	"cosmossdk.io/math"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================
// SCORE DELEGATION
// Allows users to delegate their confidence score for various purposes
// ============================

// DelegationType defines the purpose of delegation
type DelegationType string

const (
	DelegationTypeValidation DelegationType = "validation" // Delegate for IR validation
	DelegationTypeGovernance DelegationType = "governance" // Delegate voting power
	DelegationTypeReputation DelegationType = "reputation" // Lend reputation
	DelegationTypeCollateral DelegationType = "collateral" // Use as collateral
)

// ScoreDelegation represents a delegation of confidence score
type ScoreDelegation struct {
	DelegationID   string
	Delegator      string // User delegating score
	Delegate       string // User receiving delegation
	DelegatedScore uint64
	DelegationType DelegationType
	StartHeight    uint64
	EndHeight      uint64 // 0 = indefinite
	Active         bool
	Revocable      bool
	RewardSharePct uint64 // Percentage of rewards delegate shares back to delegator
	CreatedHeight  uint64
	LastUpdated    uint64
}

// CreateDelegation creates a new score delegation
func (k *Keeper) CreateDelegation(
	ctx sdk.Context,
	delegator string,
	delegate string,
	delegatedScore uint64,
	delegationType DelegationType,
	duration uint64, // blocks
	revocable bool,
	rewardSharePct uint64,
) (*ScoreDelegation, error) {
	// Validate inputs
	if delegator == "" || delegate == "" {
		return nil, types.ErrInvalidWalletAddress
	}

	if delegator == delegate {
		return nil, fmt.Errorf("cannot delegate to self")
	}

	if delegatedScore == 0 {
		return nil, fmt.Errorf("delegation amount must be positive")
	}

	if rewardSharePct > 100 {
		return nil, fmt.Errorf("reward share percentage cannot exceed 100")
	}

	// Get delegator record
	record, ok := k.GetUserRecord(ctx, delegator)
	if !ok {
		return nil, types.ErrUserRecordNotFound
	}

	// Check if delegator has enough score
	if record.TotalScore < delegatedScore {
		return nil, fmt.Errorf("insufficient score: have %d, need %d", record.TotalScore, delegatedScore)
	}

	// Calculate end height
	currentHeight := uint64(ctx.BlockHeight())
	endHeight := uint64(0)
	if duration > 0 {
		endHeight = currentHeight + duration
	}

	// Create delegation
	delegationID := fmt.Sprintf("del-%s-%s-%d", delegator, delegate, currentHeight)
	delegation := &ScoreDelegation{
		DelegationID:   delegationID,
		Delegator:      delegator,
		Delegate:       delegate,
		DelegatedScore: delegatedScore,
		DelegationType: delegationType,
		StartHeight:    currentHeight,
		EndHeight:      endHeight,
		Active:         true,
		Revocable:      revocable,
		RewardSharePct: rewardSharePct,
		CreatedHeight:  currentHeight,
		LastUpdated:    currentHeight,
	}

	// Deduct score from delegator (locked in delegation)
	record.TotalScore -= delegatedScore
	if err := k.SetUserRecord(ctx, record); err != nil {
		return nil, err
	}

	// Update delegate's delegated score (virtual score)
	delegateRecord := k.GetOrCreateUserRecord(ctx, delegate)
	delegateRecord.TotalScore += delegatedScore
	if err := k.SetUserRecord(ctx, delegateRecord); err != nil {
		return nil, err
	}

	// Store delegation
	if err := k.storeDelegation(ctx, delegation); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"score_delegated",
			sdk.NewAttribute("delegation_id", delegationID),
			sdk.NewAttribute("delegator", delegator),
			sdk.NewAttribute("delegate", delegate),
			sdk.NewAttribute("amount", fmt.Sprintf("%d", delegatedScore)),
			sdk.NewAttribute("type", string(delegationType)),
			sdk.NewAttribute("duration", fmt.Sprintf("%d", duration)),
		),
	)

	return delegation, nil
}

// RevokeDelegation revokes an active delegation
func (k *Keeper) RevokeDelegation(ctx sdk.Context, delegationID string, caller string) error {
	delegation, ok := k.getDelegation(ctx, delegationID)
	if !ok {
		return fmt.Errorf("delegation not found")
	}

	// Check if caller is delegator
	if delegation.Delegator != caller {
		return types.ErrUnauthorized
	}

	// Check if revocable
	if !delegation.Revocable {
		return fmt.Errorf("delegation is not revocable")
	}

	// Check if still active
	if !delegation.Active {
		return fmt.Errorf("delegation already inactive")
	}

	// Return score to delegator
	delegatorRecord, ok := k.GetUserRecord(ctx, delegation.Delegator)
	if !ok {
		return types.ErrUserRecordNotFound
	}

	delegatorRecord.TotalScore += delegation.DelegatedScore
	if err := k.SetUserRecord(ctx, delegatorRecord); err != nil {
		return fmt.Errorf("error in RevokeDelegation: %w", err)
	}

	// Deduct from delegate
	delegateRecord, ok := k.GetUserRecord(ctx, delegation.Delegate)
	if ok {
		if delegateRecord.TotalScore >= delegation.DelegatedScore {
			delegateRecord.TotalScore -= delegation.DelegatedScore
		} else {
			delegateRecord.TotalScore = 0
		}
		if err := k.SetUserRecord(ctx, delegateRecord); err != nil {
			return fmt.Errorf("error in RevokeDelegation: %w", err)
		}
	}

	// Mark as inactive
	delegation.Active = false
	delegation.LastUpdated = uint64(ctx.BlockHeight())
	if err := k.storeDelegation(ctx, delegation); err != nil {
		return fmt.Errorf("error in RevokeDelegation: %w", err)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"delegation_revoked",
			sdk.NewAttribute("delegation_id", delegationID),
			sdk.NewAttribute("delegator", delegation.Delegator),
			sdk.NewAttribute("delegate", delegation.Delegate),
			sdk.NewAttribute("amount", fmt.Sprintf("%d", delegation.DelegatedScore)),
		),
	)

	return nil
}

// ProcessExpiredDelegations processes expired delegations using the expiration index
// Should be called in EndBlocker
// Uses O(k) index lookup instead of O(n) full scan for scalability
func (k *Keeper) ProcessExpiredDelegations(ctx sdk.Context) (int, error) {
	currentHeight := uint64(ctx.BlockHeight())
	processed := 0

	// Use expiration index for efficient lookup - only fetch delegations expiring at current height
	delegations, err := k.getDelegationsExpiringAtHeight(ctx, currentHeight)
	if err != nil {
		k.logger.Error("failed to fetch expiring delegations", "height", currentHeight, "error", err)
		return 0, err
	}

	for _, delegation := range delegations {
		if !delegation.Active {
			// Already processed, skip
			continue
		}

		// Return score to delegator
		delegatorRecord, ok := k.GetUserRecord(ctx, delegation.Delegator)
		if ok {
			delegatorRecord.TotalScore += delegation.DelegatedScore
			if err := k.SetUserRecord(ctx, delegatorRecord); err != nil {
				k.logger.Error("failed to return score to delegator", "delegation_id", delegation.DelegationID, "error", err)
				continue
			}
		}

		// Deduct from delegate
		delegateRecord, ok := k.GetUserRecord(ctx, delegation.Delegate)
		if ok {
			if delegateRecord.TotalScore >= delegation.DelegatedScore {
				delegateRecord.TotalScore -= delegation.DelegatedScore
			} else {
				delegateRecord.TotalScore = 0
			}
			if err := k.SetUserRecord(ctx, delegateRecord); err != nil {
				k.logger.Error("failed to deduct score from delegate", "delegation_id", delegation.DelegationID, "error", err)
				continue
			}
		}

		// Mark as inactive and update
		delegation.Active = false
		delegation.LastUpdated = currentHeight
		if err := k.storeDelegation(ctx, &delegation); err != nil {
			k.logger.Error("failed to update delegation status", "delegation_id", delegation.DelegationID, "error", err)
			continue
		}

		processed++
	}

	if processed > 0 {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"delegations_expired",
				sdk.NewAttribute("count", fmt.Sprintf("%d", processed)),
				sdk.NewAttribute("height", fmt.Sprintf("%d", currentHeight)),
			),
		)
	}

	return processed, nil
}

// GetUserDelegations returns all delegations for a user using pagination
// For compatibility, this returns all results, but uses pagination internally
// For new code, use getDelegationsByUser with explicit pagination
func (k *Keeper) GetUserDelegations(ctx sdk.Context, walletAddr string, activeOnly bool) []ScoreDelegation {
	result := []ScoreDelegation{}
	offset := 0
	limit := 100

	for {
		delegations, hasMore, err := k.getDelegationsByUser(ctx, walletAddr, activeOnly, offset, limit)
		if err != nil {
			k.logger.Error("failed to fetch user delegations", "wallet", walletAddr, "error", err)
			return result
		}

		result = append(result, delegations...)

		if !hasMore {
			break
		}

		offset += limit
	}

	return result
}

// GetDelegatedScore returns total score delegated by a user
func (k *Keeper) GetDelegatedScore(ctx sdk.Context, walletAddr string) uint64 {
	delegations := k.GetUserDelegations(ctx, walletAddr, true)
	total := uint64(0)

	for _, delegation := range delegations {
		if delegation.Delegator == walletAddr {
			total += delegation.DelegatedScore
		}
	}

	return total
}

// GetReceivedDelegations returns total score delegated to a user
func (k *Keeper) GetReceivedDelegations(ctx sdk.Context, walletAddr string) uint64 {
	delegations := k.GetUserDelegations(ctx, walletAddr, true)
	total := uint64(0)

	for _, delegation := range delegations {
		if delegation.Delegate == walletAddr {
			total += delegation.DelegatedScore
		}
	}

	return total
}

// GetEffectiveScore returns user's effective score including delegations
func (k *Keeper) GetEffectiveScore(ctx sdk.Context, walletAddr string) uint64 {
	record, ok := k.GetUserRecord(ctx, walletAddr)
	if !ok {
		return 0
	}

	return record.TotalScore + k.GetReceivedDelegations(ctx, walletAddr)
}

// DistributeDelegationRewards distributes rewards to delegators
func (k *Keeper) DistributeDelegationRewards(ctx sdk.Context, delegateAddr string, totalReward math.Int, bankKeeper BankKeeper) error {
	delegations := k.GetUserDelegations(ctx, delegateAddr, true)

	for _, delegation := range delegations {
		if delegation.Delegate != delegateAddr || delegation.RewardSharePct == 0 {
			continue
		}

		// Calculate delegator share
		delegatorShare := totalReward.Mul(math.NewInt(int64(delegation.RewardSharePct))).Quo(math.NewInt(100))

		if delegatorShare.IsPositive() {
			// Send reward to delegator
			delegatorAddr, err := sdk.AccAddressFromBech32(delegation.Delegator)
			if err != nil {
				continue
			}

			coins := sdk.NewCoins(sdk.NewCoin("uaura", delegatorShare))
			if err := bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, delegatorAddr, coins); err != nil {
				continue
			}

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"delegation_reward_distributed",
					sdk.NewAttribute("delegation_id", delegation.DelegationID),
					sdk.NewAttribute("delegator", delegation.Delegator),
					sdk.NewAttribute("delegate", delegateAddr),
					sdk.NewAttribute("amount", delegatorShare.String()),
				),
			)
		}
	}

	return nil
}

// Helper functions for storing/retrieving delegations using KV store

func (k *Keeper) storeDelegation(ctx sdk.Context, delegation *ScoreDelegation) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.DelegationStoreKey(delegation.DelegationID)

	// Use JSON marshaling to preserve ALL fields
	// This is appropriate for non-proto Go structs
	bz, err := json.Marshal(delegation)
	if err != nil {
		return fmt.Errorf("failed to marshal delegation: %w", err)
	}

	if err := store.Set([]byte(key), bz); err != nil {
		return fmt.Errorf("failed to store delegation: %w", err)
	}

	// Store expiration index if delegation has end height and is active
	// This allows O(k) lookup of delegations expiring at a specific height
	if delegation.EndHeight > 0 && delegation.Active {
		expirationKey := types.ExpirationIndexKey(delegation.EndHeight, delegation.DelegationID)
		if err := store.Set([]byte(expirationKey), []byte(delegation.DelegationID)); err != nil {
			return fmt.Errorf("failed to store expiration index: %w", err)
		}
	} else if delegation.EndHeight > 0 && !delegation.Active {
		// Remove expiration index if delegation became inactive
		expirationKey := types.ExpirationIndexKey(delegation.EndHeight, delegation.DelegationID)
		if err := store.Delete([]byte(expirationKey)); err != nil {
			return fmt.Errorf("failed to delete expiration index: %w", err)
		}
	}

	return nil
}

func (k *Keeper) getDelegation(ctx sdk.Context, delegationID string) (*ScoreDelegation, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.DelegationStoreKey(delegationID)
	bz, err := store.Get([]byte(key))
	if err != nil || len(bz) == 0 {
		return nil, false
	}

	var delegation ScoreDelegation
	if err := json.Unmarshal(bz, &delegation); err != nil {
		k.logger.Error("failed to unmarshal delegation", "delegation_id", delegationID, "error", err)
		return nil, false
	}

	return &delegation, true
}

// getDelegationsPaginated retrieves delegations with pagination to prevent unbounded iteration
// Returns delegations and whether there are more results
func (k *Keeper) getDelegationsPaginated(ctx sdk.Context, offset, limit int) ([]ScoreDelegation, bool, error) {
	if limit <= 0 {
		limit = 100 // Default page size
	}
	if offset < 0 {
		offset = 0
	}

	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.DelegationStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, false, fmt.Errorf("failed to create delegation iterator: %w", err)
	}
	defer iterator.Close()

	delegations := make([]ScoreDelegation, 0, limit)
	count := 0
	skipped := 0

	for ; iterator.Valid(); iterator.Next() {
		// Skip until we reach offset
		if skipped < offset {
			skipped++
			continue
		}

		// Stop if we've collected enough
		if count >= limit {
			// Check if there are more results
			return delegations, true, nil
		}

		var delegation ScoreDelegation
		if err := json.Unmarshal(iterator.Value(), &delegation); err != nil {
			k.logger.Error("failed to unmarshal delegation in getDelegationsPaginated", "error", err)
			continue
		}

		delegations = append(delegations, delegation)
		count++
	}

	// No more results
	return delegations, false, nil
}

// getDelegationsByUser retrieves delegations for a specific user with pagination
// Filters at storage iteration level for efficiency
func (k *Keeper) getDelegationsByUser(ctx sdk.Context, walletAddr string, activeOnly bool, offset, limit int) ([]ScoreDelegation, bool, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.DelegationStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, false, fmt.Errorf("failed to create delegation iterator: %w", err)
	}
	defer iterator.Close()

	delegations := make([]ScoreDelegation, 0, limit)
	count := 0
	skipped := 0

	for ; iterator.Valid(); iterator.Next() {
		var delegation ScoreDelegation
		if err := json.Unmarshal(iterator.Value(), &delegation); err != nil {
			k.logger.Error("failed to unmarshal delegation", "error", err)
			continue
		}

		// Filter by user
		if delegation.Delegator != walletAddr && delegation.Delegate != walletAddr {
			continue
		}

		// Filter by active status
		if activeOnly && !delegation.Active {
			continue
		}

		// Skip until offset
		if skipped < offset {
			skipped++
			continue
		}

		// Check limit
		if count >= limit {
			return delegations, true, nil
		}

		delegations = append(delegations, delegation)
		count++
	}

	return delegations, false, nil
}

// getDelegationsExpiringAtHeight retrieves delegations expiring at a specific height
// This uses the expiration index for O(k) complexity instead of O(n)
// This is the CORRECT way to process expirations in EndBlocker
func (k *Keeper) getDelegationsExpiringAtHeight(ctx sdk.Context, height uint64) ([]ScoreDelegation, error) {
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.ExpirationIndexPrefix(height))
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, fmt.Errorf("failed to create expiration iterator: %w", err)
	}
	defer iterator.Close()

	delegations := []ScoreDelegation{}
	for ; iterator.Valid(); iterator.Next() {
		// Value is delegation ID
		delegationID := string(iterator.Value())

		// Fetch full delegation
		if delegation, ok := k.getDelegation(ctx, delegationID); ok {
			delegations = append(delegations, *delegation)
		} else {
			k.logger.Warn("expiration index references missing delegation", "delegation_id", delegationID, "height", height)
		}
	}

	return delegations, nil
}
