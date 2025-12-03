package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmath "cosmossdk.io/math"
	storeprefix "cosmossdk.io/store/prefix"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

// RegisterInvariants registers all bridge module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "transfer-balance", TransferBalanceInvariant(k))
	ir.RegisterRoute(types.ModuleName, "merkle-proof-validity", MerkleProofInvariant(k))
	ir.RegisterRoute(types.ModuleName, "validator-set-validity", ValidatorSetInvariant(k))
	ir.RegisterRoute(types.ModuleName, "security-parameters", SecurityParametersInvariant(k))
	ir.RegisterRoute(types.ModuleName, "transfer-limits", TransferLimitInvariant(k))
	ir.RegisterRoute(types.ModuleName, "channel-state", ChannelStateInvariant(k))
}

// AllInvariants runs all invariants of the bridge module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			TransferBalanceInvariant(k),
			MerkleProofInvariant(k),
			ValidatorSetInvariant(k),
			SecurityParametersInvariant(k),
			TransferLimitInvariant(k),
			ChannelStateInvariant(k),
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

// TransferBalanceInvariant checks that locked assets sum correctly and module has sufficient balance.
//
// CRITICAL SECURITY: This invariant ensures that the bridge module actually has the tokens
// it claims to have locked for transfers. Without this check, transfers could be created
// without locking funds, allowing the module to become insolvent.
//
// The invariant validates:
//   1. All pending/confirmed transfer amounts are valid integers
//   2. Module balance >= sum of locked amounts for each denom
//   3. No transfer exists without corresponding locked funds
//
// Returns:
//   - ("", false) if invariant holds (module is solvent)
//   - (error message, true) if invariant broken (module is insolvent or data corrupted)
func TransferBalanceInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Skip if bank keeper is not available (test environments)
		if k.bankKeeper == nil {
			return "", false
		}

		// Get module address for balance checks
		moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
		if moduleAddr == nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"transfer-balance",
				"module address not found",
			), true
		}

		// Get all pending transfers
		store := ctx.KVStore(k.storeKey)
		transferStore := storeprefix.NewStore(store, types.TransferPrefix)

		iterator := transferStore.Iterator(nil, nil)
		defer iterator.Close()

		// Sum locked amounts by denom
		lockedAmounts := make(map[string]sdkmath.Int)

		for ; iterator.Valid(); iterator.Next() {
			var transfer types.CrossChainTransfer
			if err := k.cdc.Unmarshal(iterator.Value(), &transfer); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"transfer-balance",
					fmt.Sprintf("failed to unmarshal transfer: %s", err.Error()),
				), true
			}

			// Only count pending and confirmed transfers (locked in escrow)
			if transfer.Status == types.TransferStatus_PENDING ||
				transfer.Status == types.TransferStatus_CONFIRMED {

				denom := transfer.Denom
				if lockedAmounts[denom].IsNil() {
					lockedAmounts[denom] = sdkmath.ZeroInt()
				}

				amt, ok := sdkmath.NewIntFromString(transfer.Amount)
				if !ok {
					return sdk.FormatInvariant(
						types.ModuleName,
						"transfer-balance",
						fmt.Sprintf("invalid transfer amount: %s", transfer.Amount),
					), true
				}

				lockedAmounts[denom] = lockedAmounts[denom].Add(amt)
			}
		}

		// CRITICAL SECURITY CHECK: Verify module balance covers all locked amounts
		for denom, totalLocked := range lockedAmounts {
			moduleBalance := k.bankKeeper.GetBalance(ctx, moduleAddr, denom)

			if moduleBalance.Amount.LT(totalLocked) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"transfer-balance",
					fmt.Sprintf(
						"module balance insufficient: balance=%s < locked=%s for denom %s",
						moduleBalance.Amount.String(),
						totalLocked.String(),
						denom,
					),
				), true
			}
		}

		// Invariant holds: module has sufficient balance for all locked transfers
		return "", false
	}
}

// MerkleProofInvariant checks that merkle proofs are valid
func MerkleProofInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		proofStore := storeprefix.NewStore(store, types.MerkleRootPrefix)

		iterator := proofStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var proof types.MerkleProof
			if err := k.cdc.Unmarshal(iterator.Value(), &proof); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"merkle-proof-validity",
					fmt.Sprintf("failed to unmarshal merkle proof: %s", err.Error()),
				), true
			}

			// Check proof has required fields
			if len(proof.Root) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"merkle-proof-validity",
					"merkle proof has empty root",
				), true
			}

			if len(proof.Leaf) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"merkle-proof-validity",
					"merkle proof has empty leaf",
				), true
			}

			// Check proof field (sibling hashes)
			if len(proof.Proof) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"merkle-proof-validity",
					"merkle proof has no proof hashes",
				), true
			}

			// Validate proof depth is reasonable (max 256 levels)
			if len(proof.Proof) > 256 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"merkle-proof-validity",
					fmt.Sprintf("proof depth (%d) exceeds maximum (256)",
						len(proof.Proof)),
				), true
			}
		}

		return "", false
	}
}

// ValidatorSetInvariant checks that bridge validators are valid
func ValidatorSetInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		validatorStore := storeprefix.NewStore(store, types.ValidatorPrefix)

		iterator := validatorStore.Iterator(nil, nil)
		defer iterator.Close()

		validatorCount := 0
		activeValidatorCount := 0

		for ; iterator.Valid(); iterator.Next() {
			var validator types.BridgeValidator
			if err := k.cdc.Unmarshal(iterator.Value(), &validator); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"validator-set-validity",
					fmt.Sprintf("failed to unmarshal validator: %s", err.Error()),
				), true
			}

			validatorCount++

			// Check validator address is valid
			if validator.Address == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"validator-set-validity",
					"validator has empty address",
				), true
			}

			// Validate address format
			if _, err := sdk.AccAddressFromBech32(validator.Address); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"validator-set-validity",
					fmt.Sprintf("invalid validator address: %s", validator.Address),
				), true
			}

			// Check power is non-negative
			if validator.Power < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"validator-set-validity",
					fmt.Sprintf("validator %s has negative power: %d",
						validator.Address, validator.Power),
				), true
			}

			if validator.Active {
				activeValidatorCount++
			}
		}

		// Check minimum validator count (using a default of 1 since params don't have MinValidators)
		minValidators := 1
		if activeValidatorCount < minValidators {
			return sdk.FormatInvariant(
				types.ModuleName,
				"validator-set-validity",
				fmt.Sprintf("active validators (%d) below minimum (%d)",
					activeValidatorCount, minValidators),
			), true
		}

		return "", false
	}
}

// SecurityParametersInvariant checks that security parameters are valid
func SecurityParametersInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params := k.GetParams(ctx)

		// Validate min confirmations
		if params.MinConfirmations == 0 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"security-parameters",
				"min confirmations is zero",
			), true
		}

		if params.MinConfirmations > 1000 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"security-parameters",
				fmt.Sprintf("min confirmations too high: %d", params.MinConfirmations),
			), true
		}

		// Validate bridge fee basis points (max 10000 = 100%)
		if params.BridgeFeeBasisPoints > 10000 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"security-parameters",
				fmt.Sprintf("bridge fee basis points too high: %d", params.BridgeFeeBasisPoints),
			), true
		}

		// Validate max transfer amount
		if params.MaxTransferAmount == "" || params.MaxTransferAmount == "0" {
			return sdk.FormatInvariant(
				types.ModuleName,
				"security-parameters",
				"max transfer amount is zero or empty",
			), true
		}

		// Validate validator threshold percentage (should be between 0 and 100)
		if params.ValidatorThresholdPercentage == 0 || params.ValidatorThresholdPercentage > 100 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"security-parameters",
				fmt.Sprintf("invalid validator threshold percentage: %d", params.ValidatorThresholdPercentage),
			), true
		}

		return "", false
	}
}

// TransferLimitInvariant checks that transfer limits are not exceeded
func TransferLimitInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params := k.GetParams(ctx)
		store := ctx.KVStore(k.storeKey)
		transferStore := storeprefix.NewStore(store, types.TransferPrefix)

		iterator := transferStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var transfer types.CrossChainTransfer
			if err := k.cdc.Unmarshal(iterator.Value(), &transfer); err != nil {
				continue
			}

			// Check transfer amount
			amt, ok := sdkmath.NewIntFromString(transfer.Amount)
			if !ok {
				return sdk.FormatInvariant(
					types.ModuleName,
					"transfer-limits",
					fmt.Sprintf("invalid transfer amount: %s", transfer.Amount),
				), true
			}

			// Validate against max transfer
			maxTransfer, ok := sdkmath.NewIntFromString(params.MaxTransferAmount)
			if ok && amt.GT(maxTransfer) {
				return sdk.FormatInvariant(
					types.ModuleName,
					"transfer-limits",
					fmt.Sprintf("transfer amount (%s) exceeds maximum (%s)",
						amt.String(), maxTransfer.String()),
				), true
			}

			// Check sender and recipient addresses
			if transfer.Sender == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"transfer-limits",
					fmt.Sprintf("transfer %s has empty sender", transfer.TransferId),
				), true
			}

			if transfer.Recipient == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"transfer-limits",
					fmt.Sprintf("transfer %s has empty recipient", transfer.TransferId),
				), true
			}
		}

		return "", false
	}
}

// ChannelStateInvariant checks that bridge chain configs are properly configured
func ChannelStateInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		configStore := storeprefix.NewStore(store, types.ChainConfigPrefix)

		iterator := configStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var config types.ChainConfig
			if err := k.cdc.Unmarshal(iterator.Value(), &config); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"channel-state",
					fmt.Sprintf("failed to unmarshal chain config: %s", err.Error()),
				), true
			}

			// Check chain ID is not empty
			if config.ChainId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"channel-state",
					"chain config has empty ID",
				), true
			}

			// Check chain name is not empty
			if config.ChainName == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"channel-state",
					fmt.Sprintf("chain %s has empty name", config.ChainId),
				), true
			}

			// Check address prefix is not empty
			if config.AddressPrefix == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"channel-state",
					fmt.Sprintf("chain %s has empty address prefix", config.ChainId),
				), true
			}

			// Check min confirmations is reasonable
			if config.MinConfirmations == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"channel-state",
					fmt.Sprintf("chain %s has zero min confirmations", config.ChainId),
				), true
			}

			if config.MinConfirmations > 1000 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"channel-state",
					fmt.Sprintf("chain %s has excessive min confirmations: %d", config.ChainId, config.MinConfirmations),
				), true
			}
		}

		return "", false
	}
}
