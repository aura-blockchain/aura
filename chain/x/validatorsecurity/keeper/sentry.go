package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

// RegisterSentryNode registers a sentry node for a validator
func (k Keeper) RegisterSentryNode(
	ctx context.Context,
	validatorAddr string,
	sentryAddr string,
	ipAddress string,
	port int32,
) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Validate validator exists
	if !k.HasValidatorSecurityInfo(ctx, validatorAddr) {
		return types.ErrValidatorNotFound
	}

	// Validate IP and port
	if ipAddress == "" {
		return types.ErrInvalidSentryNode
	}
	if port <= 0 || port > 65535 {
		return types.ErrInvalidSentryNode
	}

	// Create sentry node info
	blockTime := sdkCtx.BlockTime()
	sentryNode := types.SentryNodeInfo{
		Address:          sentryAddr,
		ValidatorAddress: validatorAddr,
		IpAddress:        ipAddress,
		Port:             port,
		IsActive:         true,
		LastHeartbeat:    &blockTime,
		RequestCount:     0,
		BlockedRequests:  0,
	}

	k.SetSentryNodeInfo(ctx, sentryNode)

	// Update validator's sentry node list
	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to get for validator: %w", err)
	}

	// Check if already registered
	for _, addr := range info.SentryNodeAddresses {
		if addr == sentryAddr {
			return nil // Already registered
		}
	}

	info.SentryNodeAddresses = append(info.SentryNodeAddresses, sentryAddr)
	k.SetValidatorSecurityInfo(ctx, info)

	k.Logger(sdkCtx).Info("sentry node registered",
		"validator", validatorAddr,
		"sentry", sentryAddr,
		"ip", ipAddress,
		"port", port,
	)

	return nil
}

// SetSentryNodeInfo stores sentry node information
func (k Keeper) SetSentryNodeInfo(ctx context.Context, node types.SentryNodeInfo) {
	store := k.getStore(ctx)
	key := types.GetSentryNodeInfoKey(node.Address)
	bz := k.cdc.MustMarshal(&node)
	store.Set(key, bz)
}

// GetSentryNodeInfo retrieves sentry node information
func (k Keeper) GetSentryNodeInfo(ctx context.Context, sentryAddr string) (types.SentryNodeInfo, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)
	key := types.GetSentryNodeInfoKey(sentryAddr)
	bz := store.Get(key)
	if bz == nil {
		return types.SentryNodeInfo{}, types.ErrSentryNodeNotFound
	}

	var node types.SentryNodeInfo
	if err := k.cdc.Unmarshal(bz, &node); err != nil {
		k.Logger(sdkCtx).Error("failed to unmarshal sentry node", "error", err)
		return types.SentryNodeInfo{}, err
	}
	return node, nil
}

// GetValidatorSentryNodes retrieves all sentry nodes for a validator in deterministic order.
// Results are ordered lexicographically by sentry node address to ensure consensus determinism.
func (k Keeper) GetValidatorSentryNodes(ctx context.Context, validatorAddr string) []types.SentryNodeInfo {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.SentryNodeInfoKey)
	defer iterator.Close()

	nodes := make([]types.SentryNodeInfo, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var node types.SentryNodeInfo
		if err := k.cdc.Unmarshal(iterator.Value(), &node); err != nil {
			k.Logger(sdkCtx).Error("failed to unmarshal", "error", err)
			continue
		}
		if node.ValidatorAddress == validatorAddr {
			nodes = append(nodes, node)
		}
	}

	// KVStorePrefixIterator returns keys in lexicographic order.
	return nodes
}

// UpdateSentryHeartbeat updates the heartbeat timestamp for a sentry node
func (k Keeper) UpdateSentryHeartbeat(ctx context.Context, sentryAddr string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	node, err := k.GetSentryNodeInfo(ctx, sentryAddr)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	blockTime := sdkCtx.BlockTime()
	node.LastHeartbeat = &blockTime
	node.IsActive = true

	k.SetSentryNodeInfo(ctx, node)
	return nil
}

// RecordSentryRequest records a request processed by a sentry node
func (k Keeper) RecordSentryRequest(ctx context.Context, sentryAddr string, blocked bool) error {
	node, err := k.GetSentryNodeInfo(ctx, sentryAddr)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	node.RequestCount++
	if blocked {
		node.BlockedRequests++
	}

	k.SetSentryNodeInfo(ctx, node)
	return nil
}

// DeactivateSentryNode marks a sentry node as inactive
func (k Keeper) DeactivateSentryNode(ctx context.Context, sentryAddr string) error {
	node, err := k.GetSentryNodeInfo(ctx, sentryAddr)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	node.IsActive = false
	k.SetSentryNodeInfo(ctx, node)

	// Check if validator still has enough active sentry nodes
	params := k.GetParams(ctx)
	if params.RequireSentryNodes {
		sentryNodes := k.GetValidatorSentryNodes(ctx, node.ValidatorAddress)
		activeCount := 0
		for _, n := range sentryNodes {
			if n.IsActive {
				activeCount++
			}
		}

			if clampIntToInt32(activeCount) < params.MinSentryNodes {
			// Create critical alert
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			blockTime := sdkCtx.BlockTime()
			k.CreateAlert(ctx, types.ValidatorAlert{
				Id:               fmt.Sprintf("sentry-critical-%s-%d", node.ValidatorAddress, sdkCtx.BlockHeight()),
				ValidatorAddress: node.ValidatorAddress,
				AlertType:        types.ValidatorAlert_SENTRY_NODE_OFFLINE,
				Severity:         types.ValidatorAlert_CRITICAL,
				Message:          fmt.Sprintf("Active sentry nodes (%d) below minimum (%d)", activeCount, params.MinSentryNodes),
				Timestamp:        &blockTime,
				Acknowledged:     false,
			})

			// Attempt failover if enabled
			if params.EnableAutoFailover {
				if err := k.TriggerFailover(ctx, node.ValidatorAddress); err != nil {
					k.Logger(sdkCtx).Error("failover failed", "validator", node.ValidatorAddress, "error", err)
				}
			}
		}
	}

	return nil
}

// TriggerFailover triggers failover to a backup validator
func (k Keeper) TriggerFailover(ctx context.Context, validatorAddr string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to get for validator: %w", err)
	}

	// Already in failover
	if info.FailoverActive {
		return nil
	}

	// No backup validators
	if len(info.BackupValidatorAddresses) == 0 {
		return types.ErrNoBackupValidators
	}

	// Find first available backup
	var selectedBackup string
	for _, backupAddr := range info.BackupValidatorAddresses {
		if k.HasValidatorSecurityInfo(ctx, backupAddr) {
			backupInfo, err := k.GetValidatorSecurityInfo(ctx, backupAddr)
			if err == nil && !backupInfo.IsJailed && !backupInfo.IsTombstoned {
				selectedBackup = backupAddr
				break
			}
		}
	}

	if selectedBackup == "" {
		return types.ErrInvalidBackupValidator
	}

	// Activate failover
	info.FailoverActive = true
	info.ActiveBackup = selectedBackup
	k.SetValidatorSecurityInfo(ctx, info)

	blockTime := sdkCtx.BlockTime()
	k.CreateAlert(ctx, types.ValidatorAlert{
		Id:               fmt.Sprintf("failover-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
		ValidatorAddress: validatorAddr,
		AlertType:        types.ValidatorAlert_FAILOVER_TRIGGERED,
		Severity:         types.ValidatorAlert_CRITICAL,
		Message:          fmt.Sprintf("Failover triggered to backup validator: %s", selectedBackup),
		Timestamp:        &blockTime,
		Acknowledged:     false,
	})

	k.Logger(sdkCtx).Info("failover triggered",
		"validator", validatorAddr,
		"backup", selectedBackup,
	)

	return nil
}

// RestoreFromFailover restores validator from failover state
func (k Keeper) RestoreFromFailover(ctx context.Context, validatorAddr string) error {
	info, err := k.GetValidatorSecurityInfo(ctx, validatorAddr)
	if err != nil {
		return fmt.Errorf("failed to get for validator: %w", err)
	}

	if !info.FailoverActive {
		return nil
	}

	// Verify validator is healthy before restoring
	params := k.GetParams(ctx)
	if params.RequireSentryNodes {
		sentryNodes := k.GetValidatorSentryNodes(ctx, validatorAddr)
		activeCount := 0
		for _, node := range sentryNodes {
			if node.IsActive {
				activeCount++
			}
		}

			if clampIntToInt32(activeCount) < params.MinSentryNodes {
			return types.ErrInsufficientSentryNodes
		}
	}

	// Deactivate failover
	info.FailoverActive = false
	info.ActiveBackup = ""
	k.SetValidatorSecurityInfo(ctx, info)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()
	k.CreateAlert(ctx, types.ValidatorAlert{
		Id:               fmt.Sprintf("failover-restored-%s-%d", validatorAddr, sdkCtx.BlockHeight()),
		ValidatorAddress: validatorAddr,
		AlertType:        types.ValidatorAlert_FAILOVER_TRIGGERED,
		Severity:         types.ValidatorAlert_INFO,
		Message:          "Validator restored from failover",
		Timestamp:        &blockTime,
		Acknowledged:     false,
	})

	k.Logger(sdkCtx).Info("validator restored from failover", "validator", validatorAddr)

	return nil
}
