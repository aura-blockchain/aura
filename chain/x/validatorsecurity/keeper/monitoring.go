// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

// TrackBlockSign tracks block signing for downtime detection
func (k Keeper) TrackBlockSign(ctx context.Context, validatorAddr string, signed bool) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get validator info
	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to get for validatorsecurity: %w", err)
	}

	// Don't track if tombstoned
	if info.IsTombstoned {
		return nil
	}

	params := k.GetParams(ctx)

	// Update index
	index := info.IndexOffset % params.SignedBlocksWindow

	// Get previous signed status
	store := k.getStore(ctx)
	key := types.GetValidatorMissedBlockBitKey(validatorAddr, index)
	previouslySigned := !store.Has(key)

	// Update missed blocks counter
	if !signed && previouslySigned {
		info.MissedBlocksCounter++
		store.Set(key, []byte{1}) // Mark as missed
	} else if signed && !previouslySigned {
		if info.MissedBlocksCounter > 0 {
			info.MissedBlocksCounter--
		}
		store.Delete(key) // Mark as signed
	}

	// Update last seen time if signed
	if signed {
		now := sdkCtx.BlockTime()
		info.LastSeen = &now // LastSeen is *time.Time with stdtime=true
	}

	// Increment index
	info.IndexOffset++

	// Save updated info
	k.SetValidatorSecurityInfo(ctx, info)

	// Check for downtime violation
	if err := k.HandleDowntime(ctx, validatorAddr); err != nil {
		k.Logger(sdkCtx).Error("error handling downtime", "validator", validatorAddr, "error", err)
	}

	return nil
}

// TrackSigningFromVote tracks block signing using consensus address from vote info
// This is called by BeginBlocker to process CometBFT last commit votes
func (k Keeper) TrackSigningFromVote(ctx context.Context, consAddr sdk.ConsAddress, signed bool) error {
	// Look up validator by consensus address
	validator, err := k.stakingKeeper.ValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		// Validator not found in staking - may not be bonded, skip
		return nil
	}

	valAddr := validator.GetOperator()
	if valAddr == nil {
		return nil
	}

	validatorAddr := valAddr.String()

	// Check if registered in validatorsecurity
	if !k.HasValidatorSecurityInfo(ctx, validatorAddr) {
		// Not registered, skip tracking
		return nil
	}

	return k.TrackBlockSign(ctx, validatorAddr, signed)
}

// MonitorValidator performs comprehensive monitoring checks
func (k Keeper) MonitorValidator(ctx context.Context, validatorAddr string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to get for MonitorValidator: %w", err)
	}

	params := k.GetParams(ctx)

	// Check last seen time
	if info.LastSeen != nil {
		// LastSeen is *time.Time with stdtime=true, MonitoringInterval is time.Duration
		timeSinceLastSeen := sdkCtx.BlockTime().Sub(*info.LastSeen)
		monitoringThreshold := params.MonitoringInterval * 2
		if timeSinceLastSeen > monitoringThreshold {
			alertTime := sdkCtx.BlockTime()
			k.CreateAlert(ctx, types.ValidatorAlert{
				Id:               fmt.Sprintf("inactive-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
				ValidatorAddress: validatorAddr,
				AlertType:        types.ValidatorAlert_DOWNTIME,
				Severity:         types.ValidatorAlert_WARNING,
				Message:          fmt.Sprintf("Validator inactive for %s", timeSinceLastSeen),
				Timestamp:        &alertTime, // Timestamp is *time.Time with stdtime=true
				Acknowledged:     false,
			})
		}
	}

	// Check minimum stake
	if err := k.ValidateMinimumStake(ctx, validatorAddr); err != nil {
		k.Logger(sdkCtx).Warn("validator below minimum stake", "validator", validatorAddr)
	}

	// Check sentry nodes
	if params.RequireSentryNodes {
		sentryNodes := k.GetValidatorSentryNodes(ctx, validatorAddr)
		activeCount := 0
		for _, node := range sentryNodes {
			if node.IsActive {
				// Check sentry node heartbeat
				if node.LastHeartbeat != nil {
					// LastHeartbeat is *time.Time with stdtime=true, MonitoringInterval is time.Duration
					timeSinceHeartbeat := sdkCtx.BlockTime().Sub(*node.LastHeartbeat)
					monitoringThreshold := params.MonitoringInterval * 2
					if timeSinceHeartbeat > monitoringThreshold {
						alertTime := sdkCtx.BlockTime()
						k.CreateAlert(ctx, types.ValidatorAlert{
							Id:               fmt.Sprintf("sentry-offline-%s-%s", validatorAddr, node.Address),
							ValidatorAddress: validatorAddr,
							AlertType:        types.ValidatorAlert_SENTRY_NODE_OFFLINE,
							Severity:         types.ValidatorAlert_CRITICAL,
							Message:          fmt.Sprintf("Sentry node %s offline for %s", node.Address, timeSinceHeartbeat),
							Timestamp:        &alertTime, // Timestamp is *time.Time with stdtime=true
							Acknowledged:     false,
						})
					} else {
						activeCount++
					}
				}
			}
		}

		if clampIntToInt32(activeCount) < params.MinSentryNodes {
			alertTime := sdkCtx.BlockTime()
			k.CreateAlert(ctx, types.ValidatorAlert{
				Id:               fmt.Sprintf("sentry-insufficient-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
				ValidatorAddress: validatorAddr,
				AlertType:        types.ValidatorAlert_SENTRY_NODE_OFFLINE,
				Severity:         types.ValidatorAlert_CRITICAL,
				Message:          fmt.Sprintf("Only %d active sentry nodes, minimum required: %d", activeCount, params.MinSentryNodes),
				Timestamp:        &alertTime, // Timestamp is *time.Time with stdtime=true
				Acknowledged:     false,
			})
		}
	}

	// Check geographic distribution compliance
	if params.EnableGeoDistribution && info.Region != "" {
		regionCount := k.getRegionValidatorCount(ctx, info.Region)
		if regionCount > int64(params.MaxValidatorsPerRegion) {
			alertTime := sdkCtx.BlockTime()
			k.CreateAlert(ctx, types.ValidatorAlert{
				Id:               fmt.Sprintf("geo-violation-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
				ValidatorAddress: validatorAddr,
				AlertType:        types.ValidatorAlert_GEOGRAPHIC_VIOLATION,
				Severity:         types.ValidatorAlert_WARNING,
				Message:          fmt.Sprintf("Region %s has %d validators, exceeding limit of %d", info.Region, regionCount, params.MaxValidatorsPerRegion),
				Timestamp:        &alertTime, // Timestamp is *time.Time with stdtime=true
				Acknowledged:     false,
			})
		}
	}

	// Check if failover is active
	if info.FailoverActive {
		alertTime := sdkCtx.BlockTime()
		k.CreateAlert(ctx, types.ValidatorAlert{
			Id:               fmt.Sprintf("failover-active-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
			ValidatorAddress: validatorAddr,
			AlertType:        types.ValidatorAlert_FAILOVER_TRIGGERED,
			Severity:         types.ValidatorAlert_INFO,
			Message:          fmt.Sprintf("Failover active, using backup: %s", info.ActiveBackup),
			Timestamp:        &alertTime, // Timestamp is *time.Time with stdtime=true
			Acknowledged:     false,
		})
	}

	return nil
}

// MonitorAllValidators runs monitoring for all validators.
// DEPRECATED: Use MonitorValidatorsBatched for production to prevent consensus failure
// with large validator sets. This function iterates over ALL validators every call.
func (k Keeper) MonitorAllValidators(ctx context.Context) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	validators := k.GetAllValidators(ctx)

	for _, info := range validators {
		if err := k.MonitorValidator(ctx, info.ValidatorAddress); err != nil {
			k.Logger(sdkCtx).Error("error monitoring validator",
				"validator", info.ValidatorAddress,
				"error", err,
			)
		}
	}
}

// MonitorValidatorsBatched monitors up to 'limit' validators per call using cursor-based
// pagination. This prevents consensus timeout with large validator sets by spreading
// monitoring across multiple blocks.
//
// PERFORMANCE: With 1000+ validators, unbounded monitoring causes block production
// to exceed timeout, halting the chain. This function ensures EndBlocker completes
// in <500ms regardless of validator set size.
//
// Returns: number of validators processed in this batch
func (k Keeper) MonitorValidatorsBatched(ctx context.Context, limit int) int {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if limit <= 0 {
		limit = types.MaxValidatorsPerMonitoringBatch
	}

	store := k.getStore(ctx)

	// Get cursor (last processed validator address)
	cursorBytes := store.Get(types.MonitoringCursorKey)

	// Iterate validators starting from cursor
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorSecurityInfoKey)
	defer iterator.Close()

	// Skip to cursor position if we have one
	if cursorBytes != nil {
		cursorKey := types.GetValidatorSecurityInfoKey(string(cursorBytes))
		for ; iterator.Valid(); iterator.Next() {
			if string(iterator.Key()) == string(cursorKey) {
				iterator.Next() // Move past the cursor
				break
			}
		}
	}

	processed := 0
	var lastProcessedAddr string

	for ; iterator.Valid() && processed < limit; iterator.Next() {
		var info types.ValidatorSecurityInfo
		if err := k.cdc.Unmarshal(iterator.Value(), &info); err != nil {
			k.Logger(sdkCtx).Error("failed to unmarshal validator info", "error", err)
			continue
		}

		lastProcessedAddr = info.ValidatorAddress

		if err := k.MonitorValidator(ctx, info.ValidatorAddress); err != nil {
			k.Logger(sdkCtx).Error("error monitoring validator",
				"validator", info.ValidatorAddress,
				"error", err,
			)
		}

		processed++
	}

	// Update cursor for next block
	if processed > 0 {
		if !iterator.Valid() {
			// Completed full pass, reset cursor for next iteration
			store.Delete(types.MonitoringCursorKey)
		} else {
			// Save cursor pointing to last processed validator
			store.Set(types.MonitoringCursorKey, []byte(lastProcessedAddr))
		}
	}

	return processed
}

// CreateAlert creates a new validator alert
func (k Keeper) CreateAlert(ctx context.Context, alert types.ValidatorAlert) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)

	// Generate ID if not provided
	if alert.Id == "" {
		counter := k.getAlertCounter(ctx)
		alert.Id = fmt.Sprintf("alert-%d", counter)
		k.incrementAlertCounter(ctx)
	}

	// Store alert in primary index
	key := types.GetValidatorAlertKey(alert.Id)
	bz := k.cdc.MustMarshal(&alert)
	store.Set(key, bz)

	// Add to secondary index by validator address for O(1) lookup
	if alert.ValidatorAddress != "" {
		indexKey := types.GetValidatorAlertByAddrKey(alert.ValidatorAddress, alert.Id)
		store.Set(indexKey, []byte(alert.Id))
	}

	k.Logger(sdkCtx).Info("alert created",
		"id", alert.Id,
		"validator", alert.ValidatorAddress,
		"type", alert.AlertType,
		"severity", alert.Severity,
	)
}

// GetValidatorAlerts returns all alerts for a validator in deterministic order.
// Results are ordered lexicographically by alert ID to ensure consensus determinism.
//
// PERFORMANCE: Uses secondary index by validator address for O(k) lookup where
// k = number of alerts for this validator, instead of O(n) scan of all alerts.
func (k Keeper) GetValidatorAlerts(ctx context.Context, validatorAddr string) []types.ValidatorAlert {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)

	// Use secondary index to iterate only alerts for this validator
	prefix := types.GetValidatorAlertByAddrPrefix(validatorAddr)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	alerts := make([]types.ValidatorAlert, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		// Value contains the alert ID
		alertID := string(iterator.Value())

		// Fetch the full alert from primary storage
		alertKey := types.GetValidatorAlertKey(alertID)
		bz := store.Get(alertKey)
		if bz == nil {
			// Alert was deleted but index not cleaned up - skip and log
			k.Logger(sdkCtx).Debug("stale alert index entry", "alert_id", alertID, "validator", validatorAddr)
			continue
		}

		var alert types.ValidatorAlert
		if err := k.cdc.Unmarshal(bz, &alert); err != nil {
			k.Logger(sdkCtx).Error("failed to unmarshal alert", "error", err, "alert_id", alertID)
			continue
		}
		alerts = append(alerts, alert)
	}

	// KVStorePrefixIterator returns keys in lexicographic order (by alert ID).
	return alerts
}

// GetAllAlerts returns all alerts in deterministic order.
// Results are ordered lexicographically by alert ID to ensure consensus determinism.
func (k Keeper) GetAllAlerts(ctx context.Context) []types.ValidatorAlert {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorAlertKey)
	defer iterator.Close()

	alerts := make([]types.ValidatorAlert, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var alert types.ValidatorAlert
		if err := k.cdc.Unmarshal(iterator.Value(), &alert); err != nil {
			k.Logger(sdkCtx).Error("failed to unmarshal alert", "error", err)
			continue
		}
		alerts = append(alerts, alert)
	}

	// KVStorePrefixIterator returns keys in lexicographic order (by alert ID).
	return alerts
}

// AcknowledgeAlert acknowledges an alert
func (k Keeper) AcknowledgeAlert(ctx context.Context, alertID, acknowledgerAddr string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)
	key := types.GetValidatorAlertKey(alertID)

	bz := store.Get(key)
	if bz == nil {
		return types.ErrAlertNotFound
	}

	var alert types.ValidatorAlert
	if err := k.cdc.Unmarshal(bz, &alert); err != nil {
		k.Logger(sdkCtx).Error("failed to unmarshal alert", "error", err)
		return fmt.Errorf("error in AcknowledgeAlert for GetValidatorAlertKey: %w", err)
	}

	now := sdkCtx.BlockTime()
	alert.Acknowledged = true
	alert.AcknowledgedAt = &now // AcknowledgedAt is *time.Time with stdtime=true
	alert.AcknowledgedBy = acknowledgerAddr

	bz = k.cdc.MustMarshal(&alert)
	store.Set(key, bz)

	return nil
}

func (k Keeper) getAlertCounter(ctx context.Context) uint64 {
	store := k.getStore(ctx)
	bz := store.Get(types.AlertCounterKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) incrementAlertCounter(ctx context.Context) {
	store := k.getStore(ctx)
	counter := k.getAlertCounter(ctx)
	store.Set(types.AlertCounterKey, sdk.Uint64ToBigEndian(counter+1))
}
