package keeper

import (
        storetypes "cosmossdk.io/store/types"
	"fmt"

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
	k.storeDelegation(ctx, delegation)

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
		return err
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
			return err
		}
	}

	// Mark as inactive
	delegation.Active = false
	delegation.LastUpdated = uint64(ctx.BlockHeight())
	k.storeDelegation(ctx, delegation)

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

// ProcessExpiredDelegations processes all expired delegations
// Should be called in EndBlocker
func (k *Keeper) ProcessExpiredDelegations(ctx sdk.Context) (int, error) {
	currentHeight := uint64(ctx.BlockHeight())
	processed := 0

	delegations := k.getAllDelegations(ctx)

	for _, delegation := range delegations {
		if !delegation.Active {
			continue
		}

		// Check if expired
		if delegation.EndHeight > 0 && currentHeight >= delegation.EndHeight {
			// Return score to delegator
			delegatorRecord, ok := k.GetUserRecord(ctx, delegation.Delegator)
			if ok {
				delegatorRecord.TotalScore += delegation.DelegatedScore
				k.SetUserRecord(ctx, delegatorRecord)
			}

			// Deduct from delegate
			delegateRecord, ok := k.GetUserRecord(ctx, delegation.Delegate)
			if ok {
				if delegateRecord.TotalScore >= delegation.DelegatedScore {
					delegateRecord.TotalScore -= delegation.DelegatedScore
				} else {
					delegateRecord.TotalScore = 0
				}
				k.SetUserRecord(ctx, delegateRecord)
			}

			// Mark as inactive
			delegation.Active = false
			delegation.LastUpdated = currentHeight
			k.storeDelegation(ctx, &delegation)

			processed++
		}
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

// GetUserDelegations returns all delegations for a user
func (k *Keeper) GetUserDelegations(ctx sdk.Context, walletAddr string, activeOnly bool) []ScoreDelegation {
	delegations := k.getAllDelegations(ctx)
	result := []ScoreDelegation{}

	for _, delegation := range delegations {
		if delegation.Delegator == walletAddr || delegation.Delegate == walletAddr {
			if !activeOnly || delegation.Active {
				result = append(result, delegation)
			}
		}
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

func (k *Keeper) storeDelegation(ctx sdk.Context, delegation *ScoreDelegation) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.DelegationStoreKey(delegation.DelegationID)

	// Marshal delegation (simplified - in production use protobuf)
	bz := make([]byte, 0)
	bz = append(bz, []byte(delegation.Delegator)...)
	bz = append(bz, 0) // delimiter
	bz = append(bz, []byte(delegation.Delegate)...)

	store.Set([]byte(key), bz)
}

func (k *Keeper) getDelegation(ctx sdk.Context, delegationID string) (*ScoreDelegation, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.DelegationStoreKey(delegationID)
	bz, err := store.Get([]byte(key))
	if err != nil || len(bz) == 0 {
		return nil, false
	}

	// Unmarshal delegation (simplified)
	// In production, use proper protobuf unmarshaling
	return &ScoreDelegation{DelegationID: delegationID}, true
}

func (k *Keeper) getAllDelegations(ctx sdk.Context) []ScoreDelegation {
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.DelegationStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return []ScoreDelegation{}
	}
	defer iterator.Close()

	delegations := []ScoreDelegation{}
	for ; iterator.Valid(); iterator.Next() {
		// Parse delegation from key/value
		// In production, properly unmarshal from protobuf
		delegations = append(delegations, ScoreDelegation{})
	}

	return delegations
}
