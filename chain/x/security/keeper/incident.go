package keeper

import (
	"encoding/binary"
	"fmt"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// =============================================================================
// Incident Response Operations
// =============================================================================

// SetIncident stores an incident
func (k Keeper) SetIncident(ctx sdk.Context, incident *securitypb.Incident) {
	store := k.GetStore(ctx)
	key := append(types.IncidentKey, []byte(incident.IncidentId)...)
	bz := k.cdc.MustMarshal(incident)
	store.Set(key, bz)
}

// GetIncident retrieves an incident
func (k Keeper) GetIncident(ctx sdk.Context, id string) (*securitypb.Incident, bool) {
	store := k.GetStore(ctx)
	key := append(types.IncidentKey, []byte(id)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var incident securitypb.Incident
	if err := k.cdc.Unmarshal(bz, &incident); err != nil {
		k.Logger(ctx).Error("failed to unmarshal incident", "error", err, "id", id)
		return nil, false
	}
	return &incident, true
}

// GetAllIncidents returns all incidents
func (k Keeper) GetAllIncidents(ctx sdk.Context) []*securitypb.Incident {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.IncidentKey)
	defer iterator.Close()

	var incidents []*securitypb.Incident
	for ; iterator.Valid(); iterator.Next() {
		var incident securitypb.Incident
		if err := k.cdc.Unmarshal(iterator.Value(), &incident); err != nil {
			k.Logger(ctx).Error("failed to unmarshal incident during iteration", "error", err)
			continue
		}
		incidents = append(incidents, &incident)
	}
	return incidents
}

// SetPauseState stores the system pause state
func (k Keeper) SetPauseState(ctx sdk.Context, state *types.PauseState) {
	store := k.GetStore(ctx)
	bz := k.cdc.MustMarshal(state)
	store.Set(types.PauseStateKey, bz)
}

// GetPauseState retrieves the system pause state
func (k Keeper) GetPauseState(ctx sdk.Context) *types.PauseState {
	store := k.GetStore(ctx)
	bz := store.Get(types.PauseStateKey)
	if bz == nil {
		return &types.PauseState{IsPaused: false, PauseLevel: 0}
	}
	var state types.PauseState
	if err := k.cdc.Unmarshal(bz, &state); err != nil {
		k.Logger(ctx).Error("failed to unmarshal pause state", "error", err)
		return &types.PauseState{IsPaused: false, PauseLevel: 0}
	}
	return &state
}

// SetWalletLimit stores a wallet limit
func (k Keeper) SetWalletLimit(ctx sdk.Context, limit *types.WalletLimit) {
	store := k.GetStore(ctx)
	key := append(types.WalletLimitKey, []byte(limit.WalletAddress)...)
	bz := k.cdc.MustMarshal(limit)
	store.Set(key, bz)
}

// GetWalletLimit retrieves a wallet limit
func (k Keeper) GetWalletLimit(ctx sdk.Context, walletAddr string) (*types.WalletLimit, bool) {
	store := k.GetStore(ctx)
	key := append(types.WalletLimitKey, []byte(walletAddr)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	var limit types.WalletLimit
	if err := k.cdc.Unmarshal(bz, &limit); err != nil {
		k.Logger(ctx).Error("failed to unmarshal wallet limit", "error", err, "wallet", walletAddr)
		return nil, false
	}
	return &limit, true
}

// GetAllWalletLimits returns all wallet limits
func (k Keeper) GetAllWalletLimits(ctx sdk.Context) []*types.WalletLimit {
	store := k.GetStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.WalletLimitKey)
	defer iterator.Close()

	var limits []*types.WalletLimit
	for ; iterator.Valid(); iterator.Next() {
		var limit types.WalletLimit
		if err := k.cdc.Unmarshal(iterator.Value(), &limit); err != nil {
			k.Logger(ctx).Error("failed to unmarshal wallet limit during iteration", "error", err)
			continue
		}
		limits = append(limits, &limit)
	}
	return limits
}

// DeleteWalletLimit removes a wallet limit
func (k Keeper) DeleteWalletLimit(ctx sdk.Context, walletAddr string) {
	store := k.GetStore(ctx)
	key := append(types.WalletLimitKey, []byte(walletAddr)...)
	store.Delete(key)
}

// SetNextIncidentID stores the next incident ID
func (k Keeper) SetNextIncidentID(ctx sdk.Context, id uint64) {
	store := k.GetStore(ctx)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	store.Set(types.NextIncidentIDKey, bz)
}

// GetNextIncidentID retrieves and increments the next incident ID
func (k Keeper) GetNextIncidentID(ctx sdk.Context) uint64 {
	store := k.GetStore(ctx)
	bz := store.Get(types.NextIncidentIDKey)
	if bz == nil {
		return 1
	}
	return binary.BigEndian.Uint64(bz)
}

// CreateIncident creates a new incident
func (k Keeper) CreateIncident(ctx sdk.Context, incidentType, severity, description, reportedBy string) *securitypb.Incident {
	id := k.GetNextIncidentID(ctx)
	k.SetNextIncidentID(ctx, id+1)

	incident := &securitypb.Incident{
		IncidentId:  fmt.Sprintf("INC-%d", id),
		Title:       incidentType,
		Description: description,
		DetectedAt:  timestamppb.New(ctx.BlockTime()),
		Status:      securitypb.IncidentStatus_INCIDENT_STATUS_DETECTED,
	}

	k.SetIncident(ctx, incident)

	k.Logger(ctx).Info("incident created",
		"id", incident.IncidentId,
		"type", incidentType,
		"severity", severity,
	)

	return incident
}

// PauseSystem pauses the system
func (k Keeper) PauseSystem(ctx sdk.Context, level uint32, reason, pausedBy string) error {
	state := k.GetPauseState(ctx)
	if state.IsPaused {
		return types.ErrSystemPaused
	}

	blockTime := ctx.BlockTime()
	newState := &types.PauseState{
		IsPaused:   true,
		PauseLevel: level,
		PausedAt:   &blockTime,
		PausedBy:   pausedBy,
		Reason:     reason,
	}
	k.SetPauseState(ctx, newState)

	k.Logger(ctx).Warn("system paused",
		"level", level,
		"reason", reason,
		"paused_by", pausedBy,
	)

	return nil
}

// ResumeSystem resumes the system
func (k Keeper) ResumeSystem(ctx sdk.Context) error {
	state := k.GetPauseState(ctx)
	if !state.IsPaused {
		return nil
	}

	newState := &types.PauseState{
		IsPaused:   false,
		PauseLevel: 0,
	}
	k.SetPauseState(ctx, newState)

	k.Logger(ctx).Info("system resumed")
	return nil
}

// IsSystemPaused checks if the system is paused
func (k Keeper) IsSystemPaused(ctx sdk.Context) bool {
	state := k.GetPauseState(ctx)
	return state.IsPaused
}

// GetPauseLevel returns the current pause level
func (k Keeper) GetPauseLevel(ctx sdk.Context) uint32 {
	state := k.GetPauseState(ctx)
	return state.PauseLevel
}

// CheckTransactionAllowed checks if a transaction is allowed given current security state
func (k Keeper) CheckTransactionAllowed(ctx sdk.Context, sender string, amount sdk.Coins) error {
	// Check if system is paused
	if k.IsSystemPaused(ctx) {
		pauseLevel := k.GetPauseLevel(ctx)
		if pauseLevel >= 2 {
			// Level 2+: All transactions blocked
			return types.ErrSystemPaused
		}
		// Level 1: Only transfers blocked, other txs allowed
	}

	// Check wallet limits
	limit, hasLimit := k.GetWalletLimit(ctx, sender)
	if hasLimit {
		// Check if limit has expired
		if limit.ExpiresAt != nil && ctx.BlockTime().After(*limit.ExpiresAt) {
			k.DeleteWalletLimit(ctx, sender)
		} else {
			// Check amount against limit
			for _, coin := range amount {
				maxAmount, ok := sdkmath.NewIntFromString(limit.MaxTxAmount)
				if !ok || coin.Amount.GT(maxAmount) {
					return types.ErrWalletLimitExceeded
				}
			}
		}
	}

	return nil
}

// ResolveIncident marks an incident as resolved
func (k Keeper) ResolveIncident(ctx sdk.Context, incidentID string, actionsTaken []string) error {
	incident, found := k.GetIncident(ctx, incidentID)
	if !found {
		return types.ErrIncidentNotFound
	}

	incident.Status = securitypb.IncidentStatus_INCIDENT_STATUS_RESOLVED
	incident.ResolvedAt = timestamppb.New(ctx.BlockTime())

	k.SetIncident(ctx, incident)

	k.Logger(ctx).Info("incident resolved",
		"id", incidentID,
		"actions", actionsTaken,
	)

	return nil
}
