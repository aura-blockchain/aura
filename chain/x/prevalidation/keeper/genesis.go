// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
	pb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
)

// InitGenesis initializes the module state from genesis data
func (k Keeper) InitGenesis(ctx context.Context, data *pb.GenesisState) error {
	if err := types.ValidateGenesis(data); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)

	// Set params (Params is a value type, not a pointer)
	paramsBytes, err := k.cdc.Marshal(&data.Params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}
	store.Set(types.ParamsKey, paramsBytes)

	// Import pre-validated transactions
	for _, tx := range data.PreValidatedTransactions {
		if tx == nil {
			continue
		}
		txBytes, err := k.cdc.Marshal(tx)
		if err != nil {
			return fmt.Errorf("failed to marshal transaction: %w", err)
		}
		key := types.GetPreValidatedTxKey(tx.Id)
		store.Set(key, txBytes)
	}

	// Import validation templates
	for _, template := range data.Templates {
		if template == nil {
			continue
		}
		templateBytes, err := k.cdc.Marshal(template)
		if err != nil {
			return fmt.Errorf("failed to marshal template: %w", err)
		}
		key := types.GetValidationTemplateKey(template.Id)
		store.Set(key, templateBytes)
	}

	// Import metrics
	if data.Metrics != nil {
		metricsBytes, err := k.cdc.Marshal(data.Metrics)
		if err != nil {
			return fmt.Errorf("failed to marshal metrics: %w", err)
		}
		store.Set(types.MetricsKey, metricsBytes)
	}

	if k.logger != nil {
		k.logger.Info("Prevalidation genesis imported",
			"transactions", len(data.PreValidatedTransactions),
			"templates", len(data.Templates))
	}

	return nil
}

// ExportGenesis exports the current module state to genesis
func (k Keeper) ExportGenesis(ctx context.Context) *pb.GenesisState {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(k.storeKey)

	// Export params
	var params pb.Params
	paramsBytes := store.Get(types.ParamsKey)
	if paramsBytes != nil {
		if err := k.cdc.Unmarshal(paramsBytes, &params); err != nil {
			if k.logger != nil {
				k.logger.Error("failed to unmarshal params during export", "error", err)
			}
			params = *types.DefaultParams()
		}
	} else {
		params = *types.DefaultParams()
	}

	genesis := &pb.GenesisState{
		Params:                   params,
		PreValidatedTransactions: []*pb.PreValidatedTransaction{},
		Templates:                []*pb.ValidationTemplate{},
		Metrics:                  &pb.PreValidationMetrics{},
	}

	// Export all pre-validated transactions
	iter := storetypes.KVStorePrefixIterator(store, types.PreValidatedTxPrefix)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var tx pb.PreValidatedTransaction
		if err := k.cdc.Unmarshal(iter.Value(), &tx); err == nil {
			genesis.PreValidatedTransactions = append(genesis.PreValidatedTransactions, &tx)
		}
	}

	// Export all validation templates
	iter2 := storetypes.KVStorePrefixIterator(store, types.ValidationTemplatePrefix)
	defer iter2.Close()
	for ; iter2.Valid(); iter2.Next() {
		var template pb.ValidationTemplate
		if err := k.cdc.Unmarshal(iter2.Value(), &template); err == nil {
			genesis.Templates = append(genesis.Templates, &template)
		}
	}

	// Export metrics
	var metrics pb.PreValidationMetrics
	metricsBytes := store.Get(types.MetricsKey)
	if metricsBytes != nil {
		if err := k.cdc.Unmarshal(metricsBytes, &metrics); err != nil {
			if k.logger != nil {
				k.logger.Error("failed to unmarshal metrics during export", "error", err)
			}
		} else {
			genesis.Metrics = &metrics
		}
	}

	return genesis
}
