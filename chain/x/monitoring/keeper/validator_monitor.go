package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// GetValidatorUptime retrieves uptime data for a validator from the KV store
func (k Keeper) GetValidatorUptime(ctx context.Context, validatorAddr string) (*types.ValidatorUptime, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(ValidatorUptimeKeyPrefix, []byte(validatorAddr)...)

	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrValidatorNotFound
	}

	var uptime types.ValidatorUptime
	if err := json.Unmarshal(bz, &uptime); err != nil {
		return nil, err
	}

	return &uptime, nil
}

// UpdateValidatorUptime updates uptime data for a validator
func (k Keeper) UpdateValidatorUptime(ctx context.Context, validatorAddr, moniker string, blockHeight int64, signed bool) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	if !params.EnableValidatorMonitoring {
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()

	// Try to get existing uptime
	uptime, err := k.GetValidatorUptime(ctx, validatorAddr)
	if err != nil {
		// Create new uptime record
		uptime = &types.ValidatorUptime{
			ValidatorAddress:  validatorAddr,
			Moniker:           moniker,
			TotalBlocks:       0,
			SignedBlocks:      0,
			MissedBlocks:      0,
			UptimePercentage:  0,
			LastSeen:          blockTime,
			Status:            "active",
			ConsecutiveMisses: 0,
			Jailed:            false,
		}
	}

	uptime.TotalBlocks++
	uptime.LastSeen = blockTime

	if signed {
		uptime.SignedBlocks++
		uptime.ConsecutiveMisses = 0
	} else {
		uptime.MissedBlocks++
		uptime.ConsecutiveMisses++

		// Update metrics
		if k.metrics != nil {
			k.metrics.ValidatorMissedBlocks.WithLabelValues(validatorAddr, moniker).Inc()
		}

		// Check if validator should be jailed
		if uptime.ConsecutiveMisses >= params.MaxConsecutiveMisses {
			uptime.Jailed = true
			uptime.Status = "jailed"

			// Create alert
			if params.EnableAlerts {
				if err := k.createValidatorDownAlert(ctx, uptime); err != nil {
					sdkCtx.Logger().Error("failed to create validator down alert", "validator", validatorAddr, "err", err)
				}
			}
		}
	}

	// Calculate uptime percentage
	if uptime.TotalBlocks > 0 {
		uptime.UptimePercentage = float64(uptime.SignedBlocks) / float64(uptime.TotalBlocks) * 100
	}

	// Persist to KV store
	if err := k.SetValidatorUptime(ctx, uptime); err != nil {
		return err
	}

	// Update Prometheus metrics
	if k.metrics != nil {
		k.metrics.ValidatorUptime.WithLabelValues(validatorAddr, moniker).Set(uptime.UptimePercentage)
	}

	return nil
}

// GetAllValidatorUptimes returns uptime data for all validators
func (k Keeper) GetAllValidatorUptimes(ctx context.Context) ([]*types.ValidatorUptime, error) {
	var uptimes []*types.ValidatorUptime

	err := k.IterateValidatorUptimes(ctx, func(uptime *types.ValidatorUptime) bool {
		uptimes = append(uptimes, uptime)
		return false
	})

	return uptimes, err
}

// GetJailedValidators returns all jailed validators
func (k Keeper) GetJailedValidators(ctx context.Context) ([]*types.ValidatorUptime, error) {
	var jailed []*types.ValidatorUptime

	err := k.IterateValidatorUptimes(ctx, func(uptime *types.ValidatorUptime) bool {
		if uptime.Jailed {
			jailed = append(jailed, uptime)
		}
		return false
	})

	return jailed, err
}

// createValidatorDownAlert creates an alert when a validator goes down
func (k Keeper) createValidatorDownAlert(ctx context.Context, uptime *types.ValidatorUptime) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()

	alert := &types.Alert{
		ID:       k.generateID(ctx, "alert-validator-down"),
		Type:     types.AlertTypeValidatorDown,
		Severity: types.SeverityHigh,
		Message:  fmt.Sprintf("Validator %s is down after %d consecutive misses", uptime.Moniker, uptime.ConsecutiveMisses),
		Details: map[string]interface{}{
			"validator_address":  uptime.ValidatorAddress,
			"moniker":            uptime.Moniker,
			"consecutive_misses": uptime.ConsecutiveMisses,
			"uptime_percentage":  uptime.UptimePercentage,
			"total_blocks":       uptime.TotalBlocks,
			"missed_blocks":      uptime.MissedBlocks,
		},
		Timestamp:        blockTime,
		Acknowledged:     false,
		Resolved:         false,
		NotificationSent: false,
	}

	// Store alert in KV store
	if err := k.SetAlert(ctx, alert); err != nil {
		return err
	}

	// Update metrics
	if k.metrics != nil {
		k.metrics.TotalAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
		k.metrics.ActiveAlerts.WithLabelValues(string(alert.Type), string(alert.Severity)).Inc()
	}

	return nil
}

// GetValidatorStats returns validator statistics
func (k Keeper) GetValidatorStats(ctx context.Context) (map[string]interface{}, error) {
	totalValidators := 0
	activeValidators := 0
	jailedValidators := 0
	var avgUptime float64

	err := k.IterateValidatorUptimes(ctx, func(uptime *types.ValidatorUptime) bool {
		totalValidators++
		if uptime.Status == "active" && !uptime.Jailed {
			activeValidators++
		}
		if uptime.Jailed {
			jailedValidators++
		}
		avgUptime += uptime.UptimePercentage
		return false
	})

	if err != nil {
		return nil, err
	}

	if totalValidators > 0 {
		avgUptime /= float64(totalValidators)
	}

	return map[string]interface{}{
		"total_validators":  totalValidators,
		"active_validators": activeValidators,
		"jailed_validators": jailedValidators,
		"average_uptime":    avgUptime,
	}, nil
}

// ResetValidatorJailStatus resets the jail status for a validator
func (k Keeper) ResetValidatorJailStatus(ctx context.Context, validatorAddr string) error {
	uptime, err := k.GetValidatorUptime(ctx, validatorAddr)
	if err != nil {
		return err
	}

	uptime.Jailed = false
	uptime.Status = "active"
	uptime.ConsecutiveMisses = 0

	return k.SetValidatorUptime(ctx, uptime)
}

// NOTE: DeleteValidatorUptime is implemented in keeper.go

// ExportValidatorMetrics exports validator metrics for external monitoring
func (k Keeper) ExportValidatorMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	err := k.IterateValidatorUptimes(ctx, func(uptime *types.ValidatorUptime) bool {
		metrics[uptime.ValidatorAddress] = map[string]interface{}{
			"moniker":            uptime.Moniker,
			"uptime_percentage":  uptime.UptimePercentage,
			"total_blocks":       uptime.TotalBlocks,
			"signed_blocks":      uptime.SignedBlocks,
			"missed_blocks":      uptime.MissedBlocks,
			"consecutive_misses": uptime.ConsecutiveMisses,
			"jailed":             uptime.Jailed,
			"status":             uptime.Status,
		}
		return false
	})

	if err != nil {
		return nil, err
	}

	return metrics, nil
}
