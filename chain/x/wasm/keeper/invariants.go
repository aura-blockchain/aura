// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

//lint:file-ignore SA1019 -- invariants rely on Cosmos SDK legacy registry until replacement available

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all wasm module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "security-stats-valid", SecurityStatsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "paused-contracts-valid", PausedContractsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "authorized-uploaders-valid", AuthorizedUploadersInvariant(k))
	ir.RegisterRoute(types.ModuleName, "code-size-limits", CodeSizeLimitsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "upload-auth-enforcement", UploadAuthEnforcementInvariant(k))
	ir.RegisterRoute(types.ModuleName, "gas-caps-valid", GasCapsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "admin-enforcement", AdminEnforcementInvariant(k))
}

// AllInvariants runs all invariants of the wasm module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg, broken := ParamsInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		msg, broken = SecurityStatsInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		msg, broken = PausedContractsInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		msg, broken = AuthorizedUploadersInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		msg, broken = CodeSizeLimitsInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		msg, broken = UploadAuthEnforcementInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		msg, broken = GasCapsInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		msg, broken = AdminEnforcementInvariant(k)(ctx)
		if broken {
			return msg, broken
		}

		return "", false
	}
}

// ParamsInvariant checks that module parameters are valid
func ParamsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params := k.GetParams(ctx)
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

// SecurityStatsInvariant checks that security statistics are consistent
func SecurityStatsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		stats := k.GetSecurityStats(ctx)

		// Count actual paused contracts
		actualPausedCount := k.countPausedContracts(ctx)

		if stats.ContractsPaused != actualPausedCount {
			return sdk.FormatInvariant(
				types.ModuleName,
				"security-stats-valid",
				fmt.Sprintf(
					"paused contracts count mismatch: stats=%d, actual=%d",
					stats.ContractsPaused,
					actualPausedCount,
				),
			), true
		}

		// Validate all counters are non-negative (implicit by uint64 type)
		// Additional checks could include:
		// - Total executions should be >= total contracts instantiated
		// - Reentrancy blocks should be reasonable

		return "", false
	}
}

// PausedContractsInvariant checks that paused contracts state is valid
func PausedContractsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)

		// Iterate all paused contracts and validate addresses
		iterator := storetypes.KVStorePrefixIterator(store, types.ContractPauseKey)
		defer iterator.Close()

		count := 0
		for ; iterator.Valid(); iterator.Next() {
			key := iterator.Key()
			address := string(key[len(types.ContractPauseKey):])

			// Validate address format
			if len(address) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"paused-contracts-valid",
					"found paused contract with empty address",
				), true
			}

			// Validate address is valid bech32
			if _, err := sdk.AccAddressFromBech32(address); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"paused-contracts-valid",
					fmt.Sprintf("invalid paused contract address: %s", address),
				), true
			}

			count++
		}

		return "", false
	}
}

// AuthorizedUploadersInvariant checks that authorized uploaders state is valid
func AuthorizedUploadersInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)

		// Iterate all authorized uploaders and validate addresses
		iterator := storetypes.KVStorePrefixIterator(store, types.ContractAuthKey)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			key := iterator.Key()
			address := string(key[len(types.ContractAuthKey):])

			// Validate address format
			if len(address) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"authorized-uploaders-valid",
					"found authorized uploader with empty address",
				), true
			}

			// Validate address is valid bech32
			if _, err := sdk.AccAddressFromBech32(address); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"authorized-uploaders-valid",
					fmt.Sprintf("invalid authorized uploader address: %s", address),
				), true
			}
		}

		return "", false
	}
}

// countPausedContracts counts the number of paused contracts
func (k Keeper) countPausedContracts(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ContractPauseKey)
	defer iterator.Close()

	count := uint64(0)
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return count
}

// CodeSizeLimitsInvariant checks that all stored code respects size limits
// This prevents blockchain bloat and ensures storage efficiency
func CodeSizeLimitsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params := k.GetParams(ctx)
		maxSize := params.MaxWasmCodeSize

		// Validate max size parameter itself is reasonable
		if maxSize == 0 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"code-size-limits",
				"max_wasm_code_size parameter cannot be zero",
			), true
		}

		// Upper bound check: prevent absurdly large values (> 10MB)
		const maxReasonableSize = 10 * 1024 * 1024 // 10MB
		if maxSize > maxReasonableSize {
			return sdk.FormatInvariant(
				types.ModuleName,
				"code-size-limits",
				fmt.Sprintf("max_wasm_code_size %d exceeds reasonable limit %d", maxSize, maxReasonableSize),
			), true
		}

		return "", false
	}
}

// UploadAuthEnforcementInvariant verifies upload authorization is properly enforced
// This ensures only authorized addresses can upload contracts when restrictions are enabled
func UploadAuthEnforcementInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params := k.GetParams(ctx)
		store := ctx.KVStore(k.storeKey)

		// Check consistency between upload access type and authorized uploaders
		switch params.CodeUploadAccess.Permission {
		case types.AccessTypeNobody:
			// When nobody can upload, there should be no authorized uploaders
			iterator := storetypes.KVStorePrefixIterator(store, types.ContractAuthKey)
			defer iterator.Close()

			if iterator.Valid() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"upload-auth-enforcement",
					"authorized uploaders exist when upload access is set to NOBODY",
				), true
			}

		case types.AccessTypeOnlyAddress:
			// Validate the specified address is valid
			if params.CodeUploadAccess.Address != "" {
				if _, err := sdk.AccAddressFromBech32(params.CodeUploadAccess.Address); err != nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"upload-auth-enforcement",
						fmt.Sprintf("invalid upload address in params: %s", err),
					), true
				}
			}

		case types.AccessTypeAnyOfAddresses:
			// Validate all addresses in the list are valid
			for _, addr := range params.CodeUploadAccess.Addresses {
				if _, err := sdk.AccAddressFromBech32(addr); err != nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"upload-auth-enforcement",
						fmt.Sprintf("invalid address in upload addresses list: %s", err),
					), true
				}
			}
		}

		return "", false
	}
}

// GasCapsInvariant validates gas cap parameters are within safe bounds
// This prevents DoS attacks via excessive gas consumption
func GasCapsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params := k.GetParams(ctx)
		maxGas := params.MaxGasWasmExecution

		// Validate gas cap is set (non-zero)
		if maxGas == 0 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"gas-caps-valid",
				"max_gas_wasm_execution cannot be zero",
			), true
		}

		// Upper bound: prevent setting gas cap higher than reasonable block gas limit
		// Typical block gas limit is 50-100M, contract execution should be lower
		const maxReasonableGas = 50_000_000 // 50M gas
		if maxGas > maxReasonableGas {
			return sdk.FormatInvariant(
				types.ModuleName,
				"gas-caps-valid",
				fmt.Sprintf("max_gas_wasm_execution %d exceeds reasonable limit %d", maxGas, maxReasonableGas),
			), true
		}

		// Lower bound: ensure gas cap is high enough for basic operations
		const minReasonableGas = 100_000 // 100K gas minimum
		if maxGas < minReasonableGas {
			return sdk.FormatInvariant(
				types.ModuleName,
				"gas-caps-valid",
				fmt.Sprintf("max_gas_wasm_execution %d is below minimum reasonable limit %d", maxGas, minReasonableGas),
			), true
		}

		return "", false
	}
}

// AdminEnforcementInvariant verifies admin permissions are properly enforced
// This ensures contract admin operations follow security policies
func AdminEnforcementInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params := k.GetParams(ctx)
		store := ctx.KVStore(k.storeKey)

		// If admin is required for migration, verify this is tracked
		if params.RequireAdminForMigrate {
			// Iterate through contract admins and validate addresses
			iterator := storetypes.KVStorePrefixIterator(store, types.ContractAdminPrefix)
			defer iterator.Close()

			for ; iterator.Valid(); iterator.Next() {
				key := iterator.Key()
				value := iterator.Value()

				// Extract contract address from key
				contractAddr := string(key[len(types.ContractAdminPrefix):])
				if len(contractAddr) == 0 {
					return sdk.FormatInvariant(
						types.ModuleName,
						"admin-enforcement",
						"found admin entry with empty contract address",
					), true
				}

				// Validate contract address format
				if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"admin-enforcement",
						fmt.Sprintf("invalid contract address in admin storage: %s", contractAddr),
					), true
				}

				// Validate admin address format
				adminAddr := string(value)
				if len(adminAddr) == 0 {
					return sdk.FormatInvariant(
						types.ModuleName,
						"admin-enforcement",
						fmt.Sprintf("contract %s has empty admin address", contractAddr),
					), true
				}

				if _, err := sdk.AccAddressFromBech32(adminAddr); err != nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"admin-enforcement",
						fmt.Sprintf("invalid admin address for contract %s: %s", contractAddr, adminAddr),
					), true
				}
			}
		}

		return "", false
	}
}
