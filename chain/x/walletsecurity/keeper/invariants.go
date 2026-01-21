// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// RegisterInvariants registers all walletsecurity module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "spending-limits-validity", SpendingLimitsValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "multi-sig-consistency", MultiSigConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "security-features-validity", SecurityFeaturesValidityInvariant(k))
}

// AllInvariants runs all invariants of the walletsecurity module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			SpendingLimitsValidityInvariant(k),
			MultiSigConsistencyInvariant(k),
			SecurityFeaturesValidityInvariant(k),
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
		// Note: This module may not have params, so we'll do basic validation
		// If params are added in the future, validate them here

		// For now, just return success
		return "", false
	}
}

// SpendingLimitsValidityInvariant checks spending limits validity
func SpendingLimitsValidityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Iterate through all spending limits in the store
		store := k.getStore(ctx)
		iterator, err := store.Iterator(types.SpendingLimitPrefix, storetypes.PrefixEndBytes(types.SpendingLimitPrefix))
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"spending-limits-validity",
				fmt.Sprintf("failed to create iterator: %s", err.Error()),
			), true
		}
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var limit wsproto.SpendingLimit
			if err := k.cdc.Unmarshal(iterator.Value(), &limit); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"spending-limits-validity",
					fmt.Sprintf("failed to unmarshal spending limit: %s", err.Error()),
				), true
			}

			// Wallet ID should not be empty
			if limit.WalletId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"spending-limits-validity",
					"spending limit has empty wallet ID",
				), true
			}

			// Validate amounts are valid and non-negative
			if limit.DailyLimit != "" {
				dailyLimit, ok := sdkmath.NewIntFromString(limit.DailyLimit)
				if !ok || dailyLimit.IsNegative() {
					return sdk.FormatInvariant(
						types.ModuleName,
						"spending-limits-validity",
						fmt.Sprintf("wallet %s has invalid daily limit: %s", limit.WalletId, limit.DailyLimit),
					), true
				}
			}

			if limit.WeeklyLimit != "" {
				weeklyLimit, ok := sdkmath.NewIntFromString(limit.WeeklyLimit)
				if !ok || weeklyLimit.IsNegative() {
					return sdk.FormatInvariant(
						types.ModuleName,
						"spending-limits-validity",
						fmt.Sprintf("wallet %s has invalid weekly limit: %s", limit.WalletId, limit.WeeklyLimit),
					), true
				}
			}

			if limit.MonthlyLimit != "" {
				monthlyLimit, ok := sdkmath.NewIntFromString(limit.MonthlyLimit)
				if !ok || monthlyLimit.IsNegative() {
					return sdk.FormatInvariant(
						types.ModuleName,
						"spending-limits-validity",
						fmt.Sprintf("wallet %s has invalid monthly limit: %s", limit.WalletId, limit.MonthlyLimit),
					), true
				}
			}

			// Current spent amounts should be non-negative
			if limit.CurrentDailySpent != "" {
				currentDaily, ok := sdkmath.NewIntFromString(limit.CurrentDailySpent)
				if !ok || currentDaily.IsNegative() {
					return sdk.FormatInvariant(
						types.ModuleName,
						"spending-limits-validity",
						fmt.Sprintf("wallet %s has invalid current daily spent: %s", limit.WalletId, limit.CurrentDailySpent),
					), true
				}
			}

			if limit.CurrentWeeklySpent != "" {
				currentWeekly, ok := sdkmath.NewIntFromString(limit.CurrentWeeklySpent)
				if !ok || currentWeekly.IsNegative() {
					return sdk.FormatInvariant(
						types.ModuleName,
						"spending-limits-validity",
						fmt.Sprintf("wallet %s has invalid current weekly spent: %s", limit.WalletId, limit.CurrentWeeklySpent),
					), true
				}
			}

			if limit.CurrentMonthlySpent != "" {
				currentMonthly, ok := sdkmath.NewIntFromString(limit.CurrentMonthlySpent)
				if !ok || currentMonthly.IsNegative() {
					return sdk.FormatInvariant(
						types.ModuleName,
						"spending-limits-validity",
						fmt.Sprintf("wallet %s has invalid current monthly spent: %s", limit.WalletId, limit.CurrentMonthlySpent),
					), true
				}
			}

			// Validate reset times are set if limits are enabled
			if limit.Enabled {
				if limit.DailyLimit != "" && limit.DailyLimit != "0" && limit.DailyResetAt == nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"spending-limits-validity",
						fmt.Sprintf("wallet %s daily limit enabled but reset time is nil", limit.WalletId),
					), true
				}

				if limit.WeeklyLimit != "" && limit.WeeklyLimit != "0" && limit.WeeklyResetAt == nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"spending-limits-validity",
						fmt.Sprintf("wallet %s weekly limit enabled but reset time is nil", limit.WalletId),
					), true
				}

				if limit.MonthlyLimit != "" && limit.MonthlyLimit != "0" && limit.MonthlyResetAt == nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"spending-limits-validity",
						fmt.Sprintf("wallet %s monthly limit enabled but reset time is nil", limit.WalletId),
					), true
				}
			}
		}

		return "", false
	}
}

// MultiSigConsistencyInvariant checks multi-sig wallet consistency
func MultiSigConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Iterate through all multi-sig wallets
		store := k.getStore(ctx)
		iterator, err := store.Iterator(types.MultiSigWalletPrefix, storetypes.PrefixEndBytes(types.MultiSigWalletPrefix))
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"multi-sig-consistency",
				fmt.Sprintf("failed to create iterator: %s", err.Error()),
			), true
		}
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var wallet wsproto.MultiSigWallet
			if err := k.cdc.Unmarshal(iterator.Value(), &wallet); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"multi-sig-consistency",
					fmt.Sprintf("failed to unmarshal multi-sig wallet: %s", err.Error()),
				), true
			}

			// Wallet ID should not be empty
			if wallet.WalletId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"multi-sig-consistency",
					"multi-sig wallet has empty ID",
				), true
			}

			// Signers should not be empty
			if len(wallet.Signers) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"multi-sig-consistency",
					fmt.Sprintf("multi-sig wallet %s has no signers", wallet.WalletId),
				), true
			}

			// Threshold should be valid (1 <= threshold <= signers)
			if wallet.Threshold == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"multi-sig-consistency",
					fmt.Sprintf("multi-sig wallet %s has zero threshold", wallet.WalletId),
				), true
			}

			// Note: signers is repeated string, so len gives count directly
			if int32(len(wallet.Signers)) > 0 && wallet.Threshold > int32(len(wallet.Signers)) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"multi-sig-consistency",
					fmt.Sprintf("multi-sig wallet %s threshold (%d) exceeds signers (%d)",
						wallet.WalletId, wallet.Threshold, len(wallet.Signers)),
				), true
			}

			// Validate all signer addresses (signers is []string in proto)
			seenSigners := make(map[string]bool)
			for _, signerAddr := range wallet.Signers {
				if signerAddr == "" {
					return sdk.FormatInvariant(
						types.ModuleName,
						"multi-sig-consistency",
						fmt.Sprintf("multi-sig wallet %s has signer with empty address", wallet.WalletId),
					), true
				}

				// Check for duplicate signers
				if seenSigners[signerAddr] {
					return sdk.FormatInvariant(
						types.ModuleName,
						"multi-sig-consistency",
						fmt.Sprintf("multi-sig wallet %s has duplicate signer: %s", wallet.WalletId, signerAddr),
					), true
				}
				seenSigners[signerAddr] = true

				// Validate address format
				if _, err := sdk.AccAddressFromBech32(signerAddr); err != nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"multi-sig-consistency",
						fmt.Sprintf("multi-sig wallet %s has invalid signer address: %s", wallet.WalletId, signerAddr),
					), true
				}
			}

			// Created at should be set
			if wallet.CreatedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"multi-sig-consistency",
					fmt.Sprintf("multi-sig wallet %s has nil created_at", wallet.WalletId),
				), true
			}
		}

		// Pending multi-sig transactions validation can be added when
		// PendingMultiSigTx proto is fully defined

		return "", false
	}
}

// SecurityFeaturesValidityInvariant checks security features validity
func SecurityFeaturesValidityInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Check hardware wallet configurations
		store := k.getStore(ctx)
		hwIterator, err := store.Iterator(types.HardwareWalletPrefix, storetypes.PrefixEndBytes(types.HardwareWalletPrefix))
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"security-features-validity",
				fmt.Sprintf("failed to create hardware wallet iterator: %s", err.Error()),
			), true
		}
		defer hwIterator.Close()

		for ; hwIterator.Valid(); hwIterator.Next() {
			var hwConfig wsproto.HardwareWalletConfig
			if err := k.cdc.Unmarshal(hwIterator.Value(), &hwConfig); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"security-features-validity",
					fmt.Sprintf("failed to unmarshal hardware wallet config: %s", err.Error()),
				), true
			}

			// Wallet ID should not be empty
			if hwConfig.WalletId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"security-features-validity",
					"hardware wallet config has empty wallet ID",
				), true
			}

			// Additional field validation can be added when fields are defined in proto
		}

		// Check social recovery configurations
		recoveryIterator, err := store.Iterator(types.SocialRecoveryPrefix, storetypes.PrefixEndBytes(types.SocialRecoveryPrefix))
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"security-features-validity",
				fmt.Sprintf("failed to create social recovery iterator: %s", err.Error()),
			), true
		}
		defer recoveryIterator.Close()

		for ; recoveryIterator.Valid(); recoveryIterator.Next() {
			var recovery wsproto.SocialRecoveryConfig
			if err := k.cdc.Unmarshal(recoveryIterator.Value(), &recovery); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"security-features-validity",
					fmt.Sprintf("failed to unmarshal social recovery config: %s", err.Error()),
				), true
			}

			// Wallet ID should not be empty
			if recovery.WalletId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"security-features-validity",
					"social recovery config has empty wallet ID",
				), true
			}

			// Guardians should not be empty
			if len(recovery.Guardians) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"security-features-validity",
					fmt.Sprintf("social recovery %s has no guardians", recovery.WalletId),
				), true
			}

			// Additional field validation can be added when fields are defined in proto

			// Configured at should be set
			if recovery.ConfiguredAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"security-features-validity",
					fmt.Sprintf("social recovery %s has nil configured_at", recovery.WalletId),
				), true
			}
		}

		return "", false
	}
}
