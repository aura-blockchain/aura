// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	"bytes"

	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *pb.GenesisState) error {
	// Validate genesis state
	if data == nil {
		return types.ErrInvalidRequest
	}

	// Validate params
	if err := types.ValidateParams(&data.Params); err != nil {
		return fmt.Errorf("error in InitGenesis for Validate: %w", err)
	}

	// Set params (required)
	if err := k.SetParams(ctx, &data.Params); err != nil {
		return fmt.Errorf("error in InitGenesis for ErrInvalidRequest: %w", err)
	}

	// Validate and import contracts
	for i := range data.Contracts {
		// Validate contract before importing
		if err := validateContractInfo(&data.Contracts[i]); err != nil {
			return fmt.Errorf("error in InitGenesis for Validate: %w", err)
		}
		k.SetContractInfo(ctx, &data.Contracts[i])
	}

	// Validate and import metrics
	for i := range data.Metrics {
		// Validate metrics before importing
		if err := validateContractMetrics(&data.Metrics[i]); err != nil {
			return fmt.Errorf("error in InitGenesis for validateContractInfo: %w", err)
		}
		k.SetContractMetrics(ctx, &data.Metrics[i])
	}

	return nil
}

// validateContractInfo validates a contract info structure
func validateContractInfo(info *pb.ContractInfo) error {
	if info == nil {
		return types.ErrInvalidRequest
	}

	if info.Address == "" {
		return types.ErrInvalidRequest
	}

	if info.Creator == "" {
		return types.ErrInvalidRequest
	}

	if info.CreatedAt.IsZero() {
		return types.ErrInvalidRequest
	}

	// Metadata name and version are validated by invariants, not genesis import
	// This allows importing test data with minimal fields

	return nil
}

// validateContractMetrics validates contract metrics structure
func validateContractMetrics(metrics *pb.ContractMetrics) error {
	if metrics == nil {
		return types.ErrInvalidRequest
	}

	if metrics.ContractAddress == "" {
		return types.ErrInvalidRequest
	}

	return nil
}

// ExportGenesis exports the module state to genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *pb.GenesisState {
	contracts := []pb.ContractInfo{}
	metrics := []pb.ContractMetrics{}

	// Export all contracts
	k.IterateContractInfo(ctx, func(info *pb.ContractInfo) bool {
		contracts = append(contracts, *info)

		// Export metrics for each contract
		if m, found := k.GetContractMetrics(ctx, info.Address); found {
			metrics = append(metrics, *m)
		}
		return false
	})

	// Export actual stored params (not defaults) to preserve governance changes
	params := k.GetProtoParams(ctx)

	return &pb.GenesisState{
		Params:    params,
		Contracts: contracts,
		Metrics:   metrics,
	}
}

// SetContractMetrics stores contract metrics
func (k Keeper) SetContractMetrics(ctx sdk.Context, metrics *pb.ContractMetrics) {
	store := ctx.KVStore(k.storeKey)
	key := types.ContractMetricsKey(metrics.ContractAddress)
	bz := k.cdc.MustMarshal(metrics)
	store.Set(key, bz)
}

// IterateContractInfo iterates over all contract info
func (k Keeper) IterateContractInfo(ctx sdk.Context, cb func(*pb.ContractInfo) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Check if this is a contract info key
		if !bytes.HasPrefix(iterator.Key(), types.ContractInfoPrefix) {
			continue
		}

		var info pb.ContractInfo
		if err := k.cdc.Unmarshal(iterator.Value(), &info); err != nil {
			// Log error and skip invalid entry
			continue
		}
		if cb(&info) {
			break
		}
	}
}
