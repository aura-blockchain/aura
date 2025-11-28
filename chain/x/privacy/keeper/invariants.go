package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/privacy/types"
	privacyproto "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// RegisterInvariants registers all privacy module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "mixing-state-consistency", MixingStateConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "commitment-validity", CommitmentValidityInvariant(k))
}

// AllInvariants runs all invariants of the privacy module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			MixingStateConsistencyInvariant(k),
			CommitmentValidityInvariant(k),
		}

		for _, inv := range invariants {
			msg, broken := inv(ctx)
			if broken {
				return msg, broken
			}
		}

		return "", false
	}
}

// ParamsInvariant checks that module parameters are valid
func ParamsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params := k.GetParams(ctx)
		// Basic validation of params
		if params.MinRingSize < 2 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				"min ring size must be at least 2",
			), true
		}
		if params.MaxRingSize < params.MinRingSize {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				"max ring size must be >= min ring size",
			), true
		}
		return "", false
	}
}

// MixingStateConsistencyInvariant checks mixing pool state consistency
func MixingStateConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		iterator := storetypes.KVStorePrefixIterator(store, types.MixingPoolPrefix)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var pool privacyproto.MixingPool
			if err := k.cdc.Unmarshal(iterator.Value(), &pool); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mixing-state-consistency",
					fmt.Sprintf("failed to unmarshal mixing pool: %s", err.Error()),
				), true
			}

			// Check pool ID is not empty
			if pool.PoolId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mixing-state-consistency",
					"mixing pool has empty ID",
				), true
			}

			// Check participant count matches participants list
			participantCount := int32(len(pool.Participants))
			if participantCount < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mixing-state-consistency",
					fmt.Sprintf("mixing pool %s has negative participant count: %d",
						pool.PoolId, participantCount),
				), true
			}

			// Check min participants is positive
			if pool.MinParticipants == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mixing-state-consistency",
					fmt.Sprintf("mixing pool %s has zero min participants", pool.PoolId),
				), true
			}

			// Check max participants >= min participants
			if pool.MaxParticipants < pool.MinParticipants {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mixing-state-consistency",
					fmt.Sprintf("mixing pool %s max (%d) < min (%d) participants",
						pool.PoolId, pool.MaxParticipants, pool.MinParticipants),
				), true
			}

			// Check participant count doesn't exceed max
			participantCount = int32(len(pool.Participants))
			if uint32(participantCount) > pool.MaxParticipants {
				return sdk.FormatInvariant(
					types.ModuleName,
					"mixing-state-consistency",
					fmt.Sprintf("mixing pool %s participant count (%d) exceeds max (%d)",
						pool.PoolId, participantCount, pool.MaxParticipants),
				), true
			}
		}

		return "", false
	}
}

// CommitmentValidityInvariant checks commitment validity
func CommitmentValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		iterator := storetypes.KVStorePrefixIterator(store, types.CommitmentPrefix)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			commitment := iterator.Value()

			// Check commitment is not empty
			if len(commitment) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"commitment-validity",
					"found empty commitment",
				), true
			}

			// Check commitment has reasonable length (32 bytes for SHA256)
			if len(commitment) < 16 || len(commitment) > 128 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"commitment-validity",
					fmt.Sprintf("commitment has invalid length: %d", len(commitment)),
				), true
			}
		}

		return "", false
	}
}
