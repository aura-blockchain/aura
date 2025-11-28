package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmath "cosmossdk.io/math"
	storeprefix "cosmossdk.io/store/prefix"
	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

// RegisterInvariants registers all validatorsecurity module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "validator-monitoring-consistency", ValidatorMonitoringConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "jailing-state-validity", JailingStateValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "slashing-record-integrity", SlashingRecordIntegrityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "sentry-node-consistency", SentryNodeConsistencyInvariant(k))
}

// AllInvariants runs all invariants of the validatorsecurity module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			ValidatorMonitoringConsistencyInvariant(k),
			JailingStateValidityInvariant(k),
			SlashingRecordIntegrityInvariant(k),
			SentryNodeConsistencyInvariant(k),
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
		if err := types.ValidateParams(params); err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("invalid params: %s", err.Error()),
			), true
		}
		return "", false
	}
}

// ValidatorMonitoringConsistencyInvariant checks validator monitoring state
func ValidatorMonitoringConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.ValidatorMonitoringKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var monitoring types.ValidatorMonitoring
			if err := json.Unmarshal(iterator.Value(), &monitoring); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"validator-monitoring-consistency",
					fmt.Sprintf("failed to unmarshal validator monitoring: %s", err.Error()),
				), true
			}

			// Check validator address is valid
			if _, err := sdk.ValAddressFromBech32(monitoring.ValidatorAddress); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"validator-monitoring-consistency",
					fmt.Sprintf("invalid validator address: %s", monitoring.ValidatorAddress),
				), true
			}

			// Check uptime percentage is in valid range (0-100)
			if monitoring.UptimePercentage < 0 || monitoring.UptimePercentage > 100 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"validator-monitoring-consistency",
					fmt.Sprintf("validator %s has invalid uptime: %f",
						monitoring.ValidatorAddress, monitoring.UptimePercentage),
				), true
			}

			// Check missed blocks is non-negative
			if monitoring.MissedBlocks < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"validator-monitoring-consistency",
					fmt.Sprintf("validator %s has negative missed blocks: %d",
						monitoring.ValidatorAddress, monitoring.MissedBlocks),
				), true
			}

			// Check total blocks is non-negative
			if monitoring.TotalBlocks < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"validator-monitoring-consistency",
					fmt.Sprintf("validator %s has negative total blocks: %d",
						monitoring.ValidatorAddress, monitoring.TotalBlocks),
				), true
			}

			// Check missed blocks <= total blocks
			if monitoring.MissedBlocks > monitoring.TotalBlocks {
				return sdk.FormatInvariant(
					types.ModuleName,
					"validator-monitoring-consistency",
					fmt.Sprintf("validator %s missed blocks (%d) exceeds total (%d)",
						monitoring.ValidatorAddress, monitoring.MissedBlocks, monitoring.TotalBlocks),
				), true
			}

			// Check timestamp
			if monitoring.LastUpdated == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"validator-monitoring-consistency",
					fmt.Sprintf("validator %s has nil last_updated", monitoring.ValidatorAddress),
				), true
			}
		}

		return "", false
	}
}

// JailingStateValidityInvariant checks jailing state validity
func JailingStateValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.JailingKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var jailing types.JailingRecord
			if err := json.Unmarshal(iterator.Value(), &jailing); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"jailing-state-validity",
					fmt.Sprintf("failed to unmarshal jailing record: %s", err.Error()),
				), true
			}

			// Check validator address is valid
			if _, err := sdk.ValAddressFromBech32(jailing.ValidatorAddress); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"jailing-state-validity",
					fmt.Sprintf("invalid validator address in jailing: %s", jailing.ValidatorAddress),
				), true
			}

			// Check jail reason is not empty
			if jailing.Reason == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"jailing-state-validity",
					fmt.Sprintf("jailing for %s has empty reason", jailing.ValidatorAddress),
				), true
			}

			// Check timestamp
			if jailing.JailedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"jailing-state-validity",
					fmt.Sprintf("jailing for %s has nil jailed_at", jailing.ValidatorAddress),
				), true
			}

			// If temporary, must have release time
			if !jailing.Permanent && jailing.ReleaseTime == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"jailing-state-validity",
					fmt.Sprintf("temporary jailing for %s has nil release_time", jailing.ValidatorAddress),
				), true
			}

			// Release time should be after jail time
			if jailing.ReleaseTime != nil && jailing.JailedAt != nil {
				if jailing.ReleaseTime.AsTime().Before(jailing.JailedAt.AsTime()) {
					return sdk.FormatInvariant(
						types.ModuleName,
						"jailing-state-validity",
						fmt.Sprintf("validator %s release time before jail time", jailing.ValidatorAddress),
					), true
				}
			}

			// If released, must have actual release time
			if jailing.Released && jailing.ActualReleaseTime == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"jailing-state-validity",
					fmt.Sprintf("released jailing for %s has nil actual_release_time",
						jailing.ValidatorAddress),
				), true
			}
		}

		return "", false
	}
}

// SlashingRecordIntegrityInvariant checks slashing record integrity
func SlashingRecordIntegrityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.SlashingKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var slashing types.SlashingRecord
			if err := json.Unmarshal(iterator.Value(), &slashing); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"slashing-record-integrity",
					fmt.Sprintf("failed to unmarshal slashing record: %s", err.Error()),
				), true
			}

			// Check validator address is valid
			if _, err := sdk.ValAddressFromBech32(slashing.ValidatorAddress); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"slashing-record-integrity",
					fmt.Sprintf("invalid validator address in slashing: %s", slashing.ValidatorAddress),
				), true
			}

			// Check slash amount is positive
			amount, ok := sdkmath.NewIntFromString(slashing.SlashAmount)
			if !ok || !amount.IsPositive() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"slashing-record-integrity",
					fmt.Sprintf("validator %s has invalid slash amount: %s",
						slashing.ValidatorAddress, slashing.SlashAmount),
				), true
			}

			// Check slash fraction is in valid range (0-1)
			fraction, err := sdkmath.LegacyNewDecFromStr(slashing.SlashFraction)
			if err != nil || fraction.IsNegative() || fraction.GT(sdkmath.LegacyOneDec()) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"slashing-record-integrity",
					fmt.Sprintf("validator %s has invalid slash fraction: %s",
						slashing.ValidatorAddress, slashing.SlashFraction),
				), true
			}

			// Check reason is not empty
			if slashing.Reason == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"slashing-record-integrity",
					fmt.Sprintf("slashing for %s has empty reason", slashing.ValidatorAddress),
				), true
			}

			// Check timestamp
			if slashing.SlashedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"slashing-record-integrity",
					fmt.Sprintf("slashing for %s has nil slashed_at", slashing.ValidatorAddress),
				), true
			}

			// Check infraction height is positive
			if slashing.InfractionHeight == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"slashing-record-integrity",
					fmt.Sprintf("slashing for %s has zero infraction height", slashing.ValidatorAddress),
				), true
			}
		}

		return "", false
	}
}

// SentryNodeConsistencyInvariant checks sentry node configuration consistency
func SentryNodeConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.SentryNodeKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var sentry types.SentryNode
			if err := json.Unmarshal(iterator.Value(), &sentry); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sentry-node-consistency",
					fmt.Sprintf("failed to unmarshal sentry node: %s", err.Error()),
				), true
			}

			// Check node ID is not empty
			if sentry.NodeId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sentry-node-consistency",
					"sentry node has empty ID",
				), true
			}

			// Check validator address is valid
			if _, err := sdk.ValAddressFromBech32(sentry.ValidatorAddress); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sentry-node-consistency",
					fmt.Sprintf("sentry node %s has invalid validator address: %s",
						sentry.NodeId, sentry.ValidatorAddress),
				), true
			}

			// Check IP address is not empty
			if sentry.IpAddress == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sentry-node-consistency",
					fmt.Sprintf("sentry node %s has empty IP address", sentry.NodeId),
				), true
			}

			// Check port is in valid range
			if sentry.Port == 0 || sentry.Port > 65535 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sentry-node-consistency",
					fmt.Sprintf("sentry node %s has invalid port: %d", sentry.NodeId, sentry.Port),
				), true
			}

			// Check timestamp
			if sentry.RegisteredAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sentry-node-consistency",
					fmt.Sprintf("sentry node %s has nil registered_at", sentry.NodeId),
				), true
			}

			// Active nodes should have last heartbeat
			if sentry.Active && sentry.LastHeartbeat == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"sentry-node-consistency",
					fmt.Sprintf("active sentry node %s has nil last_heartbeat", sentry.NodeId),
				), true
			}
		}

		return "", false
	}
}
