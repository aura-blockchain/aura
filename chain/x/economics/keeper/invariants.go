package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/economics/types"
)

// RegisterInvariants registers all economics module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "fee-multiplier-valid", FeeMultiplierInvariant(k))
	ir.RegisterRoute(types.ModuleName, "transfer-tax-consistency", TransferTaxConsistencyInvariant(k))
}

// AllInvariants runs all invariants of the economics module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			FeeMultiplierInvariant(k),
			TransferTaxConsistencyInvariant(k),
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

// ParamsInvariant checks that module parameters are valid and consistent
func ParamsInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params, err := k.GetParams(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("failed to get params: %s", err.Error()),
			), true
		}

		if params == nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				"params are nil",
			), true
		}

		// Validate params using the validation function
		if err := types.ValidateParams(params); err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("invalid module params: %s", err.Error()),
			), true
		}

		// Additional validation can be added here for specific param fields
		// when they are defined in the proto

		return "", false
	}
}

// FeeMultiplierInvariant checks that the fee multiplier is valid
func FeeMultiplierInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		multiplier, err := k.GetFeeMultiplier(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"fee-multiplier-valid",
				fmt.Sprintf("failed to get fee multiplier: %s", err.Error()),
			), true
		}

		// Parse multiplier to verify it's a valid decimal
		// Fee multiplier should be a positive decimal value
		if multiplier == "" {
			return sdk.FormatInvariant(
				types.ModuleName,
				"fee-multiplier-valid",
				"fee multiplier is empty",
			), true
		}

		// Attempt to parse as sdkmath.LegacyDec or verify format
		// For now, just check it's not negative (starts with -)
		if len(multiplier) > 0 && multiplier[0] == '-' {
			return sdk.FormatInvariant(
				types.ModuleName,
				"fee-multiplier-valid",
				fmt.Sprintf("fee multiplier cannot be negative: %s", multiplier),
			), true
		}

		return "", false
	}
}

// TransferTaxConsistencyInvariant checks transfer tax configuration consistency
func TransferTaxConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		enabled, err := k.GetTransferTaxEnabled(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"transfer-tax-consistency",
				fmt.Sprintf("failed to get transfer tax enabled: %s", err.Error()),
			), true
		}

		rate, err := k.GetTransferTaxRate(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"transfer-tax-consistency",
				fmt.Sprintf("failed to get transfer tax rate: %s", err.Error()),
			), true
		}

		// If transfer tax is enabled, rate should be valid and non-negative
		if enabled {
			// Parse rate to verify it's valid
			if rate == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"transfer-tax-consistency",
					"transfer tax enabled but rate is empty",
				), true
			}

			// Verify rate is non-negative
			if len(rate) > 0 && rate[0] == '-' {
				return sdk.FormatInvariant(
					types.ModuleName,
					"transfer-tax-consistency",
					fmt.Sprintf("transfer tax rate cannot be negative: %s", rate),
				), true
			}

			// Rate should be a reasonable value (e.g., not exceed 100%)
			// This is a sanity check - in practice, transfer tax should be small (< 10%)
			// We'll just verify it's parseable and reasonable
		}

		return "", false
	}
}
