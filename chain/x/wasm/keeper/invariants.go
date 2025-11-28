package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all wasm module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "security-stats-valid", SecurityStatsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "paused-contracts-valid", PausedContractsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "authorized-uploaders-valid", AuthorizedUploadersInvariant(k))
}

// AllInvariants runs all invariants of the wasm module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg, broken := ParamsInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		msg, broken = SecurityStatsInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		msg, broken = PausedContractsInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		msg, broken = AuthorizedUploadersInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		return "", false
	}
}

// ParamsInvariant checks that module parameters are valid
func ParamsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params := k.GetParams(ctx)
		if err := params.Validate(); err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("invalid params: %s", err.Error()),
			), true
		}
		return "", false
	}
}

// SecurityStatsInvariant checks that security statistics are consistent
func SecurityStatsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		stats := k.GetSecurityStats(ctx)

		// Count actual paused contracts
		actualPausedCount := k.countPausedContracts(ctx)

		if stats.TotalPausedContracts != actualPausedCount {
			return sdk.FormatInvariant(
				types.ModuleName,
				"security-stats-valid",
				fmt.Sprintf(
					"paused contracts count mismatch: stats=%d, actual=%d",
					stats.TotalPausedContracts,
					actualPausedCount,
				),
			), true
		}

		// Validate all counters are non-negative (implicit by uint64 type)
		// Additional checks could include:
		// - Total executions should be >= total contracts instantiated
		// - Reentrancy blocks should be reasonable

		return "", false
	}
}

// PausedContractsInvariant checks that paused contracts state is valid
func PausedContractsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)

		// Iterate all paused contracts and validate addresses
		iterator := storetypes.KVStorePrefixIterator(store, types.ContractPauseKey)
		defer iterator.Close()

		count := 0
		for ; iterator.Valid(); iterator.Next() {
			key := iterator.Key()
			address := string(key[len(types.ContractPauseKey):])

			// Validate address format
			if len(address) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"paused-contracts-valid",
					"found paused contract with empty address",
				), true
			}

			// Validate address is valid bech32
			if _, err := sdk.AccAddressFromBech32(address); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"paused-contracts-valid",
					fmt.Sprintf("invalid paused contract address: %s", address),
				), true
			}

			count++
		}

		return "", false
	}
}

// AuthorizedUploadersInvariant checks that authorized uploaders state is valid
func AuthorizedUploadersInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)

		// Iterate all authorized uploaders and validate addresses
		iterator := storetypes.KVStorePrefixIterator(store, types.ContractAuthKey)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			key := iterator.Key()
			address := string(key[len(types.ContractAuthKey):])

			// Validate address format
			if len(address) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"authorized-uploaders-valid",
					"found authorized uploader with empty address",
				), true
			}

			// Validate address is valid bech32
			if _, err := sdk.AccAddressFromBech32(address); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"authorized-uploaders-valid",
					fmt.Sprintf("invalid authorized uploader address: %s", address),
				), true
			}
		}

		return "", false
	}
}

// countPausedContracts counts the number of paused contracts
func (k Keeper) countPausedContracts(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ContractPauseKey)
	defer iterator.Close()

	count := uint64(0)
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return count
}
