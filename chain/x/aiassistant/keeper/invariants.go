// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"strings"

	storeprefix "cosmossdk.io/store/prefix"
	"github.com/aequitas/aura/chain/x/aiassistant/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all aiassistant module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "assistant-state-consistency", AssistantStateConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "stake-balance-consistency", StakeBalanceConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "locale-index-consistency", LocaleIndexConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "status-validity", StatusValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "heartbeat-validity", HeartbeatValidityInvariant(k))
}

// AllInvariants runs all invariants of the aiassistant module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			AssistantStateConsistencyInvariant(k),
			StakeBalanceConsistencyInvariant(k),
			LocaleIndexConsistencyInvariant(k),
			StatusValidityInvariant(k),
			HeartbeatValidityInvariant(k),
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
func ParamsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params, err := k.GetParams(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("failed to get params: %s", err.Error()),
			), true
		}
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

// AssistantStateConsistencyInvariant checks that all assistants have valid state
func AssistantStateConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.AssistantKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var assistant types.Assistant
			if err := k.cdc.Unmarshal(iterator.Value(), &assistant); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"assistant-state-consistency",
					fmt.Sprintf("failed to unmarshal assistant: %s", err.Error()),
				), true
			}

			// Validate assistant address
			if _, err := sdk.AccAddressFromBech32(assistant.AssistantAddress); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"assistant-state-consistency",
					fmt.Sprintf("invalid assistant address: %s", assistant.AssistantAddress),
				), true
			}

			// Validate owner address
			if _, err := sdk.AccAddressFromBech32(assistant.OwnerAddress); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"assistant-state-consistency",
					fmt.Sprintf("invalid owner address: %s", assistant.OwnerAddress),
				), true
			}

			// Validate stake has denom
			if strings.TrimSpace(assistant.Stake.Denom) == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"assistant-state-consistency",
					fmt.Sprintf("assistant %s has empty stake denom", assistant.AssistantAddress),
				), true
			}

			// Validate sponsorship balance denom
			if strings.TrimSpace(assistant.SponsorshipBalance.Denom) == "" && assistant.SponsorshipBalance.Amount.IsPositive() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"assistant-state-consistency",
					fmt.Sprintf("assistant %s has empty sponsorship denom with funds", assistant.AssistantAddress),
				), true
			}

			// Validate locales
			if len(assistant.Locales) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"assistant-state-consistency",
					fmt.Sprintf("assistant %s has no locales", assistant.AssistantAddress),
				), true
			}

			// Validate last heartbeat is set
			if assistant.LastHeartbeat.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"assistant-state-consistency",
					fmt.Sprintf("assistant %s has zero last heartbeat", assistant.AssistantAddress),
				), true
			}
		}

		return "", false
	}
}

// StakeBalanceConsistencyInvariant checks that stake amounts are consistent
func StakeBalanceConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.AssistantKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var assistant types.Assistant
			if err := k.cdc.Unmarshal(iterator.Value(), &assistant); err != nil {
				continue
			}

			// Validate stake is parseable
			if _, err := balanceToCoin(assistant.Stake); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"stake-balance-consistency",
					fmt.Sprintf("assistant %s has invalid stake: %s", assistant.AssistantAddress, err.Error()),
				), true
			}

			// Validate sponsorship balance is parseable
			if _, err := balanceToCoin(assistant.SponsorshipBalance); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"stake-balance-consistency",
					fmt.Sprintf("assistant %s has invalid sponsorship balance: %s", assistant.AssistantAddress, err.Error()),
				), true
			}
		}

		return "", false
	}
}

// LocaleIndexConsistencyInvariant checks that locale indexes are consistent
func LocaleIndexConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.AssistantKeyPrefix)

		// Build map of assistant -> locales from primary storage
		assistantLocales := make(map[string]map[string]bool)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var assistant types.Assistant
			if err := k.cdc.Unmarshal(iterator.Value(), &assistant); err != nil {
				continue
			}

			localesMap := make(map[string]bool)
			for _, locale := range assistant.Locales {
				localesMap[locale] = true
			}
			assistantLocales[assistant.AssistantAddress] = localesMap
		}

		// Now check that locale indexes match
		localeStore := storeprefix.NewStore(store, types.LocaleKeyPrefix)
		localeIterator := localeStore.Iterator(nil, nil)
		defer localeIterator.Close()

		for ; localeIterator.Valid(); localeIterator.Next() {
			assistantAddr := string(localeIterator.Value())

			// Parse locale from key (this is simplified - actual parsing may differ)
			key := localeIterator.Key()
			// Extract locale from key (format: locale\x00address)
			nullIdx := -1
			for i, b := range key {
				if b == 0x00 {
					nullIdx = i
					break
				}
			}
			if nullIdx == -1 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"locale-index-consistency",
					"malformed locale index key",
				), true
			}
			locale := string(key[:nullIdx])

			// Check that the assistant exists and has this locale
			if locales, ok := assistantLocales[assistantAddr]; !ok {
				return sdk.FormatInvariant(
					types.ModuleName,
					"locale-index-consistency",
					fmt.Sprintf("locale index references non-existent assistant: %s", assistantAddr),
				), true
			} else if !locales[locale] {
				return sdk.FormatInvariant(
					types.ModuleName,
					"locale-index-consistency",
					fmt.Sprintf("assistant %s indexed for locale %s but doesn't declare it", assistantAddr, locale),
				), true
			}
		}

		return "", false
	}
}

// StatusValidityInvariant checks that assistant statuses are valid
func StatusValidityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.AssistantKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var assistant types.Assistant
			if err := k.cdc.Unmarshal(iterator.Value(), &assistant); err != nil {
				continue
			}

			// Check status is valid enum value
			validStatuses := []types.AssistantStatus{
				types.AssistantStatus_UNSPECIFIED,
				types.AssistantStatus_ACTIVE,
				types.AssistantStatus_JAILED,
				types.AssistantStatus_TOMBSTONED,
			}

			statusValid := false
			for _, vs := range validStatuses {
				if assistant.Status == vs {
					statusValid = true
					break
				}
			}

			if !statusValid {
				return sdk.FormatInvariant(
					types.ModuleName,
					"status-validity",
					fmt.Sprintf("assistant %s has invalid status: %d", assistant.AssistantAddress, assistant.Status),
				), true
			}

			// Tombstoned assistants should have zero stake
			if assistant.Status == types.AssistantStatus_TOMBSTONED {
				stake, err := balanceToCoin(assistant.Stake)
				if err == nil && !stake.Amount.IsZero() {
					return sdk.FormatInvariant(
						types.ModuleName,
						"status-validity",
						fmt.Sprintf("tombstoned assistant %s has non-zero stake: %s", assistant.AssistantAddress, stake.String()),
					), true
				}
			}
		}

		return "", false
	}
}

// HeartbeatValidityInvariant checks that heartbeat timestamps are valid
func HeartbeatValidityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.AssistantKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var assistant types.Assistant
			if err := k.cdc.Unmarshal(iterator.Value(), &assistant); err != nil {
				continue
			}

			heartbeatTime := assistant.LastHeartbeat
			if !heartbeatTime.IsZero() && heartbeatTime.After(ctx.BlockTime()) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"heartbeat-validity",
					fmt.Sprintf("assistant %s has future heartbeat: %s (current: %s)",
						assistant.AssistantAddress,
						heartbeatTime.String(),
						ctx.BlockTime().String()),
				), true
			}
		}

		return "", false
	}
}
