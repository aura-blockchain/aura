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

	// Set module parameters (gs.Params is already a pointer in GenesisState)
	if err := k.SetParams(ctx, gs.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// Import validator security info
	for _, info := range gs.Validators {
		if info != nil {
			k.SetValidatorSecurityInfo(ctx, *info) // Dereference pointer

			// Update region counts
			if gs.Params.EnableGeoDistribution && info.Region != "" {
				k.incrementRegionCount(ctx, info.Region)
			}
		}
	}

	// Import double sign evidences
	for _, evidence := range gs.DoubleSignEvidences {
		if evidence != nil {
			k.SetDoubleSignEvidence(ctx, *evidence) // Dereference pointer
		}
	}

	// Import downtime infractions
	for _, infraction := range gs.DowntimeInfractions {
		if infraction != nil {
			k.SetDowntimeInfraction(ctx, *infraction) // Dereference pointer
		}
	}

	// Import alerts
	for _, alert := range gs.Alerts {
		if alert != nil {
			k.SetValidatorAlert(ctx, *alert) // Dereference pointer
		}
	}

	// Import sentry nodes
	for _, sentry := range gs.SentryNodes {
		if sentry != nil {
			k.SetSentryNode(ctx, *sentry) // Dereference pointer
		}
	}

	k.Logger(ctx).Info("validator security genesis state initialized",
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

	k.Logger(ctx).Info("validator security genesis state exported",
		"validators", len(validators),
		"double_sign_evidences", len(doubleSignEvidences),
		"downtime_infractions", len(downtimeInfractions),
		"alerts", len(alerts),
		"sentry_nodes", len(sentryNodes),
	)

	// Convert slices to pointer slices for genesis state
	validatorPtrs := make([]*types.ValidatorSecurityInfo, len(validators))
	for i := range validators {
		validatorPtrs[i] = &validators[i]
	}

	evidencePtrs := make([]*types.DoubleSignEvidence, len(doubleSignEvidences))
	for i := range doubleSignEvidences {
		evidencePtrs[i] = &doubleSignEvidences[i]
	}

	infractionPtrs := make([]*types.DowntimeInfraction, len(downtimeInfractions))
	for i := range downtimeInfractions {
		infractionPtrs[i] = &downtimeInfractions[i]
	}

	alertPtrs := make([]*types.ValidatorAlert, len(alerts))
	for i := range alerts {
		alertPtrs[i] = &alerts[i]
	}

	sentryPtrs := make([]*types.SentryNodeInfo, len(sentryNodes))
	for i := range sentryNodes {
		sentryPtrs[i] = &sentryNodes[i]
	}

	return &types.GenesisState{
		Params:              params, // params is already a pointer
		Validators:          validatorPtrs,
		DoubleSignEvidences: evidencePtrs,
		DowntimeInfractions: infractionPtrs,
		Alerts:              alertPtrs,
		SentryNodes:         sentryPtrs,
	}
}

// GetAllDoubleSignEvidences retrieves all double sign evidences
func (k Keeper) GetAllDoubleSignEvidences(ctx context.Context) []types.DoubleSignEvidence {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.DoubleSignEvidenceKey)
	defer iterator.Close()

	var evidences []types.DoubleSignEvidence
	for ; iterator.Valid(); iterator.Next() {
		var evidence types.DoubleSignEvidence
		if err := k.cdc.Unmarshal(iterator.Value(), &evidence); err != nil {

			k.logger.Error("failed to unmarshal", "error", err)

			continue

		}
		evidences = append(evidences, evidence)
	}

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
	store := k.getStore(ctx)
	key := append(types.DoubleSignEvidenceKey, []byte(validatorAddr)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrEvidenceNotFound
	}

	var evidence types.DoubleSignEvidence
	if err := k.cdc.Unmarshal(bz, &evidence); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

		continue

	}
	return &evidence, nil
}

// GetAllDowntimeInfractions retrieves all downtime infractions
func (k Keeper) GetAllDowntimeInfractions(ctx context.Context) []types.DowntimeInfraction {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.DowntimeInfractionKey)
	defer iterator.Close()

	var infractions []types.DowntimeInfraction
	for ; iterator.Valid(); iterator.Next() {
		var infraction types.DowntimeInfraction
		if err := k.cdc.Unmarshal(iterator.Value(), &infraction); err != nil {

			k.logger.Error("failed to unmarshal", "error", err)

			continue

		}
		infractions = append(infractions, infraction)
	}

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
	store := k.getStore(ctx)
	key := append(types.DowntimeInfractionKey, []byte(validatorAddr)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrInfractionNotFound
	}

	var infraction types.DowntimeInfraction
	if err := k.cdc.Unmarshal(bz, &infraction); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

		continue

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
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorAlertKey)
	defer iterator.Close()

	var alerts []types.ValidatorAlert
	for ; iterator.Valid(); iterator.Next() {
		var alert types.ValidatorAlert
		if err := k.cdc.Unmarshal(iterator.Value(), &alert); err != nil {

			k.logger.Error("failed to unmarshal", "error", err)

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
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.SentryNodeKey)
	defer iterator.Close()

	var sentryNodes []types.SentryNodeInfo
	for ; iterator.Valid(); iterator.Next() {
		var node types.SentryNodeInfo
		if err := k.cdc.Unmarshal(iterator.Value(), &node); err != nil {

			k.logger.Error("failed to unmarshal", "error", err)

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
