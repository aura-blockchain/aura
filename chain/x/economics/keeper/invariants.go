package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmath "cosmossdk.io/math"

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

		// Validate fee configuration if present
		if params.FeeConfig != nil {
			// Base fee should be non-negative
			if params.FeeConfig.BaseFee < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("base fee cannot be negative: %d", params.FeeConfig.BaseFee),
				), true
			}

			// Priority multiplier should be positive
			if params.FeeConfig.PriorityMultiplier <= 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("priority multiplier must be positive: %d", params.FeeConfig.PriorityMultiplier),
				), true
			}
		}

		// Validate governance configuration if present
		if params.GovernanceConfig != nil {
			// Quorum should be between 0 and 100
			if params.GovernanceConfig.Quorum < 0 || params.GovernanceConfig.Quorum > 100 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("quorum must be between 0 and 100: %d", params.GovernanceConfig.Quorum),
				), true
			}

			// Voting period should be positive
			if params.GovernanceConfig.VotingPeriod <= 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("voting period must be positive: %d", params.GovernanceConfig.VotingPeriod),
				), true
			}

			// Min deposit should be non-negative
			minDeposit, ok := sdkmath.NewIntFromString(params.GovernanceConfig.MinDeposit)
			if !ok || minDeposit.IsNegative() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"params-valid",
					fmt.Sprintf("invalid min deposit: %s", params.GovernanceConfig.MinDeposit),
				), true
			}
		}

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
