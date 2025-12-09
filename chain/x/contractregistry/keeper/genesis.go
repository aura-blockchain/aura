package keeper

import (
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

	// Set params (required)
	if data.Params == nil {
		return types.ErrInvalidRequest
	}
	if err := k.SetParams(ctx, data.Params); err != nil {
		return err
	}

	// Import contracts
	for _, contract := range data.Contracts {
		if contract != nil {
			k.SetContractInfo(ctx, contract)
		}
	}

	// Import metrics
	for _, metrics := range data.Metrics {
		if metrics != nil {
			k.SetContractMetrics(ctx, metrics)
		}
	}

	return nil
}

// ExportGenesis exports the module state to genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *pb.GenesisState {
	contracts := []*pb.ContractInfo{}
	metrics := []*pb.ContractMetrics{}

	// Export all contracts
	k.IterateContractInfo(ctx, func(info *pb.ContractInfo) bool {
		contracts = append(contracts, info)

		// Export metrics for each contract
		if m, found := k.GetContractMetrics(ctx, info.Address); found {
			metrics = append(metrics, m)
		}
		return false
	})

	return &pb.GenesisState{
		Params:    types.DefaultParams(),
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
