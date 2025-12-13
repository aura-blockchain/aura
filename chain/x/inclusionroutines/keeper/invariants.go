package keeper

//lint:file-ignore SA1019 -- use invariant registry until Cosmos SDK removes the legacy APIs

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all inclusionroutines module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "ir-consistency", IRConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "prerequisite-validity", PrerequisiteValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "rate-limit-validity", RateLimitValidityInvariant(k))
}

// AllInvariants runs all invariants of the inclusionroutines module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			IRConsistencyInvariant(k),
			PrerequisiteValidityInvariant(k),
			RateLimitValidityInvariant(k),
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

// IRConsistencyInvariant checks that IR definitions are valid
func IRConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		allIRs := k.GetAllIRs(ctx)

		for _, ir := range allIRs {
			if err := k.validateIR(ir); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"ir-consistency",
					fmt.Sprintf("invalid IR %s: %s", ir.Id, err.Error()),
				), true
			}
		}

		return "", false
	}
}

// PrerequisiteValidityInvariant checks prerequisite relationships
func PrerequisiteValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		allIRs := k.GetAllIRs(ctx)
		validIRs := make(map[string]bool)

		for _, ir := range allIRs {
			validIRs[ir.Id] = true
		}

		for _, ir := range allIRs {
			prereq, exists := k.GetPrerequisite(ctx, ir.Id)
			if !exists {
				continue
			}

			for _, reqID := range prereq.RequiredIrIds {
				if !validIRs[reqID] {
					return sdk.FormatInvariant(
						types.ModuleName,
						"prerequisite-validity",
						fmt.Sprintf("IR %s references non-existent prerequisite %s", ir.Id, reqID),
					), true
				}
			}
		}

		return "", false
	}
}

// RateLimitValidityInvariant checks rate limit configurations
func RateLimitValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		allIRs := k.GetAllIRs(ctx)

		for _, ir := range allIRs {
			rateLimit, exists := k.GetRateLimit(ctx, ir.Id)
			if !exists {
				continue
			}

			if err := k.ValidateRateLimit(rateLimit); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"rate-limit-validity",
					fmt.Sprintf("invalid rate limit for IR %s: %s", ir.Id, err.Error()),
				), true
			}
		}

		return "", false
	}
}
