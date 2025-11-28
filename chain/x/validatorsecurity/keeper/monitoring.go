package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

// TrackBlockSign tracks block signing for downtime detection
func (k Keeper) TrackBlockSign(ctx context.Context, validatorAddr string, signed bool) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get validator info
	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return err
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
		info.LastSeen = timestamppb.New(now)
	}

	// Increment index
	info.IndexOffset++

	// Save updated info
	k.SetValidatorSecurityInfo(ctx, info)

	// Check for downtime violation
	if err := k.HandleDowntime(ctx, validatorAddr); err != nil {
		k.Logger(ctx).Error("error handling downtime", "validator", validatorAddr, "error", err)
	}

	return nil
}

// MonitorValidator performs comprehensive monitoring checks
func (k Keeper) MonitorValidator(ctx context.Context, validatorAddr string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return err
	}

	params := k.GetParams(ctx)

	// Check last seen time
	if info.LastSeen != nil {
		timeSinceLastSeen := sdkCtx.BlockTime().Sub(info.LastSeen.AsTime())
		monitoringThreshold := params.MonitoringInterval.AsDuration() * 2
		if timeSinceLastSeen > monitoringThreshold {
			k.CreateAlert(ctx, types.ValidatorAlert{
				Id:               fmt.Sprintf("inactive-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
				ValidatorAddress: validatorAddr,
				AlertType:        types.ValidatorAlert_DOWNTIME,
				Severity:         types.ValidatorAlert_WARNING,
				Message:          fmt.Sprintf("Validator inactive for %s", timeSinceLastSeen),
				Timestamp:        timestamppb.New(sdkCtx.BlockTime()),
				Acknowledged:     false,
			})
		}
	}

	// Check minimum stake
	if err := k.ValidateMinimumStake(ctx, validatorAddr); err != nil {
		k.Logger(ctx).Warn("validator below minimum stake", "validator", validatorAddr)
	}

	// Check sentry nodes
	if params.RequireSentryNodes {
		sentryNodes := k.GetValidatorSentryNodes(ctx, validatorAddr)
		activeCount := 0
		for _, node := range sentryNodes {
			if node.IsActive {
				// Check sentry node heartbeat
				if node.LastHeartbeat != nil {
					timeSinceHeartbeat := sdkCtx.BlockTime().Sub(node.LastHeartbeat.AsTime())
					monitoringThreshold := params.MonitoringInterval.AsDuration() * 2
					if timeSinceHeartbeat > monitoringThreshold {
						k.CreateAlert(ctx, types.ValidatorAlert{
							Id:               fmt.Sprintf("sentry-offline-%s-%s", validatorAddr, node.Address),
							ValidatorAddress: validatorAddr,
							AlertType:        types.ValidatorAlert_SENTRY_NODE_OFFLINE,
							Severity:         types.ValidatorAlert_CRITICAL,
							Message:          fmt.Sprintf("Sentry node %s offline for %s", node.Address, timeSinceHeartbeat),
							Timestamp:        timestamppb.New(sdkCtx.BlockTime()),
							Acknowledged:     false,
						})
					} else {
						activeCount++
					}
				}
			}
		}

		if int32(activeCount) < params.MinSentryNodes {
			k.CreateAlert(ctx, types.ValidatorAlert{
				Id:               fmt.Sprintf("sentry-insufficient-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
				ValidatorAddress: validatorAddr,
				AlertType:        types.ValidatorAlert_SENTRY_NODE_OFFLINE,
				Severity:         types.ValidatorAlert_CRITICAL,
				Message:          fmt.Sprintf("Only %d active sentry nodes, minimum required: %d", activeCount, params.MinSentryNodes),
				Timestamp:        timestamppb.New(sdkCtx.BlockTime()),
				Acknowledged:     false,
			})
		}
	}

	// Check geographic distribution compliance
	if params.EnableGeoDistribution && info.Region != "" {
		regionCount := k.getRegionValidatorCount(ctx, info.Region)
		if regionCount > int64(params.MaxValidatorsPerRegion) {
			k.CreateAlert(ctx, types.ValidatorAlert{
				Id:               fmt.Sprintf("geo-violation-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
				ValidatorAddress: validatorAddr,
				AlertType:        types.ValidatorAlert_GEOGRAPHIC_VIOLATION,
				Severity:         types.ValidatorAlert_WARNING,
				Message:          fmt.Sprintf("Region %s has %d validators, exceeding limit of %d", info.Region, regionCount, params.MaxValidatorsPerRegion),
				Timestamp:        timestamppb.New(sdkCtx.BlockTime()),
				Acknowledged:     false,
			})
		}
	}

	// Check if failover is active
	if info.FailoverActive {
		k.CreateAlert(ctx, types.ValidatorAlert{
			Id:               fmt.Sprintf("failover-active-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
			ValidatorAddress: validatorAddr,
			AlertType:        types.ValidatorAlert_FAILOVER_TRIGGERED,
			Severity:         types.ValidatorAlert_INFO,
			Message:          fmt.Sprintf("Failover active, using backup: %s", info.ActiveBackup),
			Timestamp:        timestamppb.New(sdkCtx.BlockTime()),
			Acknowledged:     false,
		})
	}

	return nil
}

// MonitorAllValidators runs monitoring for all validators
func (k Keeper) MonitorAllValidators(ctx context.Context) {
	validators := k.GetAllValidators(ctx)

	for _, info := range validators {
		if err := k.MonitorValidator(ctx, info.ValidatorAddress); err != nil {
			k.Logger(ctx).Error("error monitoring validator",
				"validator", info.ValidatorAddress,
				"error", err,
			)
		}
	}
}

// CreateAlert creates a new validator alert
func (k Keeper) CreateAlert(ctx context.Context, alert types.ValidatorAlert) {
	store := k.getStore(ctx)

	// Generate ID if not provided
	if alert.Id == "" {
		counter := k.getAlertCounter(ctx)
		alert.Id = fmt.Sprintf("alert-%d", counter)
		k.incrementAlertCounter(ctx)
	}

	key := types.GetValidatorAlertKey(alert.Id)
	bz := k.cdc.MustMarshal(&alert)
	store.Set(key, bz)

	k.Logger(ctx).Info("alert created",
		"id", alert.Id,
		"validator", alert.ValidatorAddress,
		"type", alert.AlertType,
		"severity", alert.Severity,
	)
}

// GetValidatorAlerts returns all alerts for a validator
func (k Keeper) GetValidatorAlerts(ctx context.Context, validatorAddr string) []types.ValidatorAlert {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorAlertKey)
	defer iterator.Close()

	var alerts []types.ValidatorAlert
	for ; iterator.Valid(); iterator.Next() {
		var alert types.ValidatorAlert
		k.cdc.MustUnmarshal(iterator.Value(), &alert)
		if alert.ValidatorAddress == validatorAddr {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// GetAllAlerts returns all alerts
func (k Keeper) GetAllAlerts(ctx context.Context) []types.ValidatorAlert {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorAlertKey)
	defer iterator.Close()

	var alerts []types.ValidatorAlert
	for ; iterator.Valid(); iterator.Next() {
		var alert types.ValidatorAlert
		k.cdc.MustUnmarshal(iterator.Value(), &alert)
		alerts = append(alerts, alert)
	}

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
	k.cdc.MustUnmarshal(bz, &alert)

	now := sdkCtx.BlockTime()
	alert.Acknowledged = true
	alert.AcknowledgedAt = timestamppb.New(now)
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
