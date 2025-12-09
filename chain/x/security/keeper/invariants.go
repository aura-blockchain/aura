package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmath "cosmossdk.io/math"

	"github.com/aequitas/aura/chain/x/security/types"
)

// RegisterInvariants registers all security module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "spending-limits-validity", SpendingLimitsValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "audit-log-integrity", AuditLogIntegrityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "privacy-data-consistency", PrivacyDataConsistencyInvariant(k))
}

// AllInvariants runs all invariants of the security module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			SpendingLimitsValidityInvariant(k),
			AuditLogIntegrityInvariant(k),
			PrivacyDataConsistencyInvariant(k),
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

		// Validate network security parameters
		if params.NetworkSecurity != nil {
			// Max peers should be reasonable (1-1000)
			if params.NetworkSecurity.MaxPeers == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					"max peers cannot be zero",
				), true
			}

			if params.NetworkSecurity.MaxPeers > 1000 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("max peers too high: %d (max 1000)", params.NetworkSecurity.MaxPeers),
				), true
			}

			// Min reputation score should be 0-100
			if params.NetworkSecurity.MinReputationScore < 0 || params.NetworkSecurity.MinReputationScore > 100 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("min reputation score invalid: %d (must be 0-100)", params.NetworkSecurity.MinReputationScore),
				), true
			}

			// Rate limit should be positive
			if params.NetworkSecurity.RateLimitPerPeer == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					"rate limit per peer cannot be zero",
				), true
			}
		}

		// Validate validator security parameters
		if params.ValidatorSecurity != nil {
			// Slash fraction should be 0-100% (in basis points: 0-10000)
			if params.ValidatorSecurity.SlashFractionDoubleSign > 10000 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("slash fraction double sign too high: %d (max 10000)", params.ValidatorSecurity.SlashFractionDoubleSign),
				), true
			}

			if params.ValidatorSecurity.SlashFractionDowntime > 10000 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("slash fraction downtime too high: %d (max 10000)", params.ValidatorSecurity.SlashFractionDowntime),
				), true
			}

			// Jail duration should be reasonable (at least 1 minute)
			if params.ValidatorSecurity.JailDuration > 0 && params.ValidatorSecurity.JailDuration < 60 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("jail duration too short: %d seconds (min 60)", params.ValidatorSecurity.JailDuration),
				), true
			}

			// Signed blocks window should be positive
			if params.ValidatorSecurity.SignedBlocksWindow > 0 && params.ValidatorSecurity.SignedBlocksWindow < 100 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("signed blocks window too small: %d (min 100)", params.ValidatorSecurity.SignedBlocksWindow),
				), true
			}

			// Min signed per window should be reasonable (0-100%)
			minSigned, ok := sdkmath.LegacyNewDecFromStr(params.ValidatorSecurity.MinSignedPerWindow)
			if ok == nil && (minSigned.IsNegative() || minSigned.GT(sdkmath.LegacyOneDec())) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("min signed per window invalid: %s (must be 0.0-1.0)", params.ValidatorSecurity.MinSignedPerWindow),
				), true
			}
		}

		// Validate cryptography parameters
		if params.Cryptography != nil {
			// Supported algorithms should not be empty
			if len(params.Cryptography.SupportedAlgorithms) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					"supported algorithms cannot be empty",
				), true
			}
		}

		return "", false
	}
}

// SpendingLimitsValidityInvariant checks spending limits validity
func SpendingLimitsValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		limits := k.GetAllSpendingLimits(ctx)

		for _, limit := range limits {
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
				if limit.DailyLimit != "" && limit.DailyResetAt == nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"spending-limits-validity",
						fmt.Sprintf("wallet %s daily limit enabled but reset time is nil", limit.WalletId),
					), true
				}

				if limit.WeeklyLimit != "" && limit.WeeklyResetAt == nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"spending-limits-validity",
						fmt.Sprintf("wallet %s weekly limit enabled but reset time is nil", limit.WalletId),
					), true
				}

				if limit.MonthlyLimit != "" && limit.MonthlyResetAt == nil {
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

// AuditLogIntegrityInvariant checks audit log integrity
func AuditLogIntegrityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		logs := k.GetAllAuditLogEntries(ctx)

		// Track log IDs to detect duplicates
		logIDs := make(map[string]bool)

		for _, log := range logs {
			// Log ID should not be empty
			if log.LogId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"audit-log-integrity",
					"audit log has empty ID",
				), true
			}

			// Check for duplicate log IDs
			if logIDs[log.LogId] {
				return sdk.FormatInvariant(
					types.ModuleName,
					"audit-log-integrity",
					fmt.Sprintf("duplicate audit log ID: %s", log.LogId),
				), true
			}
			logIDs[log.LogId] = true

			// Timestamp should be set
			if log.Timestamp == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"audit-log-integrity",
					fmt.Sprintf("audit log %s has nil timestamp", log.LogId),
				), true
			}

			// Event type should not be empty
			if log.EventType == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"audit-log-integrity",
					fmt.Sprintf("audit log %s has empty event type", log.LogId),
				), true
			}

			// Actor should not be empty
			if log.Actor == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"audit-log-integrity",
					fmt.Sprintf("audit log %s has empty actor", log.LogId),
				), true
			}

			// Severity should be valid
			if log.Severity != "info" && log.Severity != "warning" && log.Severity != "error" && log.Severity != "critical" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"audit-log-integrity",
					fmt.Sprintf("audit log %s has invalid severity: %s", log.LogId, log.Severity),
				), true
			}
		}

		return "", false
	}
}

// PrivacyDataConsistencyInvariant checks privacy data consistency
func PrivacyDataConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Check stealth addresses
		stealthAddrs := k.GetAllStealthAddresses(ctx)

		for _, addr := range stealthAddrs {
			// One-time address should not be empty
			if len(addr.OneTimeAddress) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"privacy-data-consistency",
					"stealth address has empty one-time address",
				), true
			}

			// View tag should be valid (typically 1-8 bytes)
			if len(addr.ViewTag) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"privacy-data-consistency",
					"stealth address has empty view tag",
				), true
			}

			// Transaction hash should not be empty
			if addr.TxHash == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"privacy-data-consistency",
					"stealth address has empty transaction hash",
				), true
			}

			// Created at should be set
			if addr.CreatedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"privacy-data-consistency",
					"stealth address has nil created_at",
				), true
			}
		}

		// Check ring signatures
		ringSigs := k.GetAllRingSignatures(ctx)

		for _, sig := range ringSigs {
			// Key image should not be empty
			if len(sig.KeyImage) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"privacy-data-consistency",
					"ring signature has empty key image",
				), true
			}

			// Ring size should be reasonable (typically 11-16 for Monero-style)
			if sig.RingSize < 2 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"privacy-data-consistency",
					fmt.Sprintf("ring signature has too small ring size: %d (min 2)", sig.RingSize),
				), true
			}

			if sig.RingSize > 128 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"privacy-data-consistency",
					fmt.Sprintf("ring signature has too large ring size: %d (max 128)", sig.RingSize),
				), true
			}

			// Ring members should match ring size
			if uint32(len(sig.RingMembers)) != sig.RingSize {
				return sdk.FormatInvariant(
					types.ModuleName,
					"privacy-data-consistency",
					fmt.Sprintf("ring signature members (%d) doesn't match ring size (%d)",
						len(sig.RingMembers), sig.RingSize),
				), true
			}

			// Transaction hash should not be empty
			if sig.TxHash == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"privacy-data-consistency",
					"ring signature has empty transaction hash",
				), true
			}

			// Created at should be set
			if sig.CreatedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"privacy-data-consistency",
					"ring signature has nil created_at",
				), true
			}
		}

		return "", false
	}
}
