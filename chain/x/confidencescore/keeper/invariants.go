package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all confidencescore module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "user-record-consistency", UserRecordConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "score-range", ScoreRangeInvariant(k))
}

// AllInvariants runs all invariants of the confidencescore module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			UserRecordConsistencyInvariant(k),
			ScoreRangeInvariant(k),
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
		params := k.GetParams()
		if err := types.ValidateParams(&params); err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("invalid params: %s", err.Error()),
			), true
		}
		return "", false
	}
}

// UserRecordConsistencyInvariant checks that all user records have valid state
func UserRecordConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := k.storeService.OpenKVStore(ctx)

		// Iterate over all user records
		prefix := []byte(types.UserRecordStoreKeyPrefix)
		endBytes := append([]byte(nil), prefix...)
		endBytes[len(endBytes)-1]++
		iterator, err := store.Iterator(prefix, endBytes)
		if err != nil {
			return "", false
		}
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var record types.UserConfidenceRecord
			if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"user-record-consistency",
					fmt.Sprintf("failed to unmarshal user record: %s", err.Error()),
				), true
			}

			// Check wallet address is not empty
			if record.WalletAddress == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"user-record-consistency",
					"user record has empty wallet address",
				), true
			}

			// Check score is within valid range
			if record.TotalScore > 1000000 { // reasonable upper limit
				return sdk.FormatInvariant(
					types.ModuleName,
					"user-record-consistency",
					fmt.Sprintf("user %s has excessive score: %d", record.WalletAddress, record.TotalScore),
				), true
			}

			// Check timestamps
			if record.LastUpdated != nil && record.LastUpdated.Seconds > ctx.BlockTime().Unix()+3600 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"user-record-consistency",
					fmt.Sprintf("user %s has future last update time", record.WalletAddress),
				), true
			}
		}

		return "", false
	}
}

// ScoreRangeInvariant checks that all scores are within reasonable bounds
func ScoreRangeInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := k.storeService.OpenKVStore(ctx)

		// Iterate over all user records
		prefix := []byte(types.UserRecordStoreKeyPrefix)
		endBytes := append([]byte(nil), prefix...)
		endBytes[len(endBytes)-1]++
		iterator, err := store.Iterator(prefix, endBytes)
		if err != nil {
			return "", false
		}
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var record types.UserConfidenceRecord
			if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
				continue
			}

			// Check total score is positive
			if record.TotalScore < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"score-range",
					fmt.Sprintf("user %s has negative score: %d", record.WalletAddress, record.TotalScore),
				), true
			}

			// Check arena scores are positive
			for arena, arenaScore := range record.ArenaScores {
				if arenaScore.TotalScore < 0 {
					return sdk.FormatInvariant(
						types.ModuleName,
						"score-range",
						fmt.Sprintf("user %s has negative arena score in %s: %d",
							record.WalletAddress, arena, arenaScore.TotalScore),
					), true
				}
			}
		}

		return "", false
	}
}
