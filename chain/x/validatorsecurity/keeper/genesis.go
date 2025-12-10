package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

// InitGenesis initializes the validator security module's state from genesis
func (k Keeper) InitGenesis(ctx context.Context, gs *types.GenesisState) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Validate genesis state
	if err := types.ValidateGenesisState(gs); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	// Set module parameters (gs.Params is a value type with nullable=false)
	if err := k.SetParams(ctx, &gs.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// Import validator security info (nullable=false means slice contains values, not pointers)
	for i := range gs.Validators {
		info := &gs.Validators[i]
		k.SetValidatorSecurityInfo(ctx, *info)

		// Update region counts
		if gs.Params.EnableGeoDistribution && info.Region != "" {
			k.incrementRegionCount(ctx, info.Region)
		}
	}

	// Import double sign evidences (nullable=false means slice contains values, not pointers)
	for i := range gs.DoubleSignEvidences {
		evidence := &gs.DoubleSignEvidences[i]
		k.SetDoubleSignEvidence(ctx, *evidence)
	}

	// Import downtime infractions (nullable=false means slice contains values, not pointers)
	for i := range gs.DowntimeInfractions {
		infraction := &gs.DowntimeInfractions[i]
		k.SetDowntimeInfraction(ctx, *infraction)
	}

	// Import alerts (nullable=false means slice contains values, not pointers)
	for i := range gs.Alerts {
		alert := &gs.Alerts[i]
		k.SetValidatorAlert(ctx, *alert)
	}

	// Import sentry nodes (nullable=false means slice contains values, not pointers)
	for i := range gs.SentryNodes {
		sentry := &gs.SentryNodes[i]
		k.SetSentryNode(ctx, *sentry)
	}

	k.Logger(sdkCtx).Info("validator security genesis state initialized",
		"validators", len(gs.Validators),
		"double_sign_evidences", len(gs.DoubleSignEvidences),
		"downtime_infractions", len(gs.DowntimeInfractions),
		"alerts", len(gs.Alerts),
		"sentry_nodes", len(gs.SentryNodes),
	)

	// Emit genesis initialization event
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validatorsecurity_genesis",
			sdk.NewAttribute("validators_count", fmt.Sprintf("%d", len(gs.Validators))),
			sdk.NewAttribute("alerts_count", fmt.Sprintf("%d", len(gs.Alerts))),
		),
	)

	return nil
}

// ExportGenesis exports the validator security module's state for genesis
func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get current parameters
	params := k.GetParams(ctx)

	// Export validators
	validators := k.GetAllValidators(ctx)

	// Export double sign evidences
	doubleSignEvidences := k.GetAllDoubleSignEvidences(ctx)

	// Export downtime infractions
	downtimeInfractions := k.GetAllDowntimeInfractions(ctx)

	// Export alerts (only active ones for genesis)
	alerts := k.GetActiveValidatorAlerts(ctx)

	// Export sentry nodes
	sentryNodes := k.GetAllSentryNodes(ctx)

	k.Logger(sdkCtx).Info("validator security genesis state exported",
		"validators", len(validators),
		"double_sign_evidences", len(doubleSignEvidences),
		"downtime_infractions", len(downtimeInfractions),
		"alerts", len(alerts),
		"sentry_nodes", len(sentryNodes),
	)

	// GenesisState expects value types (nullable=false), so return slices directly
	return &types.GenesisState{
		Params:              *params, // Dereference params to value type
		Validators:          validators,
		DoubleSignEvidences: doubleSignEvidences,
		DowntimeInfractions: downtimeInfractions,
		Alerts:              alerts,
		SentryNodes:         sentryNodes,
	}
}

// GetAllDoubleSignEvidences retrieves all double sign evidences in deterministic order.
// Results are ordered lexicographically by validator address to ensure consensus determinism.
func (k Keeper) GetAllDoubleSignEvidences(ctx context.Context) []types.DoubleSignEvidence {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.DoubleSignEvidenceKey)
	defer iterator.Close()

	var evidences []types.DoubleSignEvidence
	for ; iterator.Valid(); iterator.Next() {
		var evidence types.DoubleSignEvidence
		if err := k.cdc.Unmarshal(iterator.Value(), &evidence); err != nil {
			k.Logger(sdkCtx).Error("failed to unmarshal double sign evidence", "error", err)
			continue
		}
		evidences = append(evidences, evidence)
	}

	// KVStorePrefixIterator returns keys in lexicographic order.
	return evidences
}

// SetDoubleSignEvidence stores double sign evidence
func (k Keeper) SetDoubleSignEvidence(ctx context.Context, evidence types.DoubleSignEvidence) {
	store := k.getStore(ctx)
	key := append(types.DoubleSignEvidenceKey, []byte(evidence.ValidatorAddress)...)
	bz := k.cdc.MustMarshal(&evidence)
	store.Set(key, bz)
}

// GetDoubleSignEvidence retrieves double sign evidence for a validator
func (k Keeper) GetDoubleSignEvidence(ctx context.Context, validatorAddr string) (*types.DoubleSignEvidence, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)
	key := append(types.DoubleSignEvidenceKey, []byte(validatorAddr)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrEvidenceNotFound
	}

	var evidence types.DoubleSignEvidence
	if err := k.cdc.Unmarshal(bz, &evidence); err != nil {
		k.Logger(sdkCtx).Error("failed to unmarshal double sign evidence", "error", err)
		return nil, err
	}
	return &evidence, nil
}

// GetAllDowntimeInfractions retrieves all downtime infractions in deterministic order.
// Results are ordered lexicographically by validator address to ensure consensus determinism.
func (k Keeper) GetAllDowntimeInfractions(ctx context.Context) []types.DowntimeInfraction {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.DowntimeInfractionKey)
	defer iterator.Close()

	var infractions []types.DowntimeInfraction
	for ; iterator.Valid(); iterator.Next() {
		var infraction types.DowntimeInfraction
		if err := k.cdc.Unmarshal(iterator.Value(), &infraction); err != nil {
			k.Logger(sdkCtx).Error("failed to unmarshal downtime infraction", "error", err)
			continue
		}
		infractions = append(infractions, infraction)
	}

	// KVStorePrefixIterator returns keys in lexicographic order.
	return infractions
}

// SetDowntimeInfraction stores downtime infraction
func (k Keeper) SetDowntimeInfraction(ctx context.Context, infraction types.DowntimeInfraction) {
	store := k.getStore(ctx)
	key := append(types.DowntimeInfractionKey, []byte(infraction.ValidatorAddress)...)
	bz := k.cdc.MustMarshal(&infraction)
	store.Set(key, bz)
}

// GetDowntimeInfraction retrieves downtime infraction for a validator
func (k Keeper) GetDowntimeInfraction(ctx context.Context, validatorAddr string) (*types.DowntimeInfraction, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)
	key := append(types.DowntimeInfractionKey, []byte(validatorAddr)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrInfractionNotFound
	}

	var infraction types.DowntimeInfraction
	if err := k.cdc.Unmarshal(bz, &infraction); err != nil {
		k.Logger(sdkCtx).Error("failed to unmarshal downtime infraction", "error", err)
		return nil, err
	}
	return &infraction, nil
}

// GetActiveValidatorAlerts retrieves all active (unacknowledged) validator alerts
func (k Keeper) GetActiveValidatorAlerts(ctx context.Context) []types.ValidatorAlert {
	allAlerts := k.GetAllValidatorAlerts(ctx)
	activeAlerts := []types.ValidatorAlert{}

	for _, alert := range allAlerts {
		if !alert.Acknowledged { // Use Acknowledged field instead of Resolved
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// GetAllValidatorAlerts retrieves all validator alerts
func (k Keeper) GetAllValidatorAlerts(ctx context.Context) []types.ValidatorAlert {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorAlertKey)
	defer iterator.Close()

	var alerts []types.ValidatorAlert
	for ; iterator.Valid(); iterator.Next() {
		var alert types.ValidatorAlert
		if err := k.cdc.Unmarshal(iterator.Value(), &alert); err != nil {
			k.Logger(sdkCtx).Error("failed to unmarshal validator alert", "error", err)
			continue
		}
		alerts = append(alerts, alert)
	}

	return alerts
}

// SetValidatorAlert stores a validator alert
func (k Keeper) SetValidatorAlert(ctx context.Context, alert types.ValidatorAlert) {
	store := k.getStore(ctx)
	key := append(types.ValidatorAlertKey, []byte(alert.Id)...)
	bz := k.cdc.MustMarshal(&alert)
	store.Set(key, bz)
}

// GetAllSentryNodes retrieves all sentry nodes
func (k Keeper) GetAllSentryNodes(ctx context.Context) []types.SentryNodeInfo {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.SentryNodeKey)
	defer iterator.Close()

	var sentryNodes []types.SentryNodeInfo
	for ; iterator.Valid(); iterator.Next() {
		var node types.SentryNodeInfo
		if err := k.cdc.Unmarshal(iterator.Value(), &node); err != nil {
			k.Logger(sdkCtx).Error("failed to unmarshal sentry node", "error", err)
			continue
		}
		sentryNodes = append(sentryNodes, node)
	}

	return sentryNodes
}

// SetSentryNode stores sentry node info
func (k Keeper) SetSentryNode(ctx context.Context, node types.SentryNodeInfo) {
	store := k.getStore(ctx)
	key := append(types.SentryNodeKey, []byte(node.Address)...) // Use Address field instead of NodeId
	bz := k.cdc.MustMarshal(&node)
	store.Set(key, bz)
}
