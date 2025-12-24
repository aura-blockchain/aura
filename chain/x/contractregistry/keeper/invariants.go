package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	storeprefix "cosmossdk.io/store/prefix"
	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
)

// RegisterInvariants registers all contractregistry module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "contract-metadata-consistency", ContractMetadataConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "code-hash-validity", CodeHashValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "contract-address-validity", ContractAddressValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "version-consistency", VersionConsistencyInvariant(k))
}

// AllInvariants runs all invariants of the contractregistry module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			ContractMetadataConsistencyInvariant(k),
			CodeHashValidityInvariant(k),
			ContractAddressValidityInvariant(k),
			VersionConsistencyInvariant(k),
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
		params, err := k.GetParams(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("failed to get params: %s", err.Error()),
			), true
		}
		if err := params.Validate(); err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"params-valid",
				fmt.Sprintf("invalid params: %s", err.Error()),
			), true
		}
		return "", false
	}
}

// ContractMetadataConsistencyInvariant checks that all contract metadata is valid
func ContractMetadataConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.ContractInfoPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var info pb.ContractInfo
			if err := k.cdc.Unmarshal(iterator.Value(), &info); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"contract-metadata-consistency",
					fmt.Sprintf("failed to unmarshal contract info: %s", err.Error()),
				), true
			}

			// Metadata is a value type (not nullable), always present
			metadata := info.Metadata

			// Check name is not empty
			if metadata.Name == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"contract-metadata-consistency",
					fmt.Sprintf("contract %s has empty name", info.Address),
				), true
			}

			// Check version is not empty
			if metadata.Version == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"contract-metadata-consistency",
					fmt.Sprintf("contract %s has empty version", info.Address),
				), true
			}

			// Check creator address is valid if provided
			if info.Creator != "" {
				if _, err := sdk.AccAddressFromBech32(info.Creator); err != nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"contract-metadata-consistency",
						fmt.Sprintf("contract %s has invalid creator address: %s",
							info.Address, info.Creator),
					), true
				}
			}

			// Check timestamps - CreatedAt is time.Time (not pointer), check if zero
			if info.CreatedAt.IsZero() {
				return sdk.FormatInvariant(
					types.ModuleName,
					"contract-metadata-consistency",
					fmt.Sprintf("contract %s has zero created_at timestamp", info.Address),
				), true
			}
		}

		return "", false
	}
}

// CodeHashValidityInvariant checks that code hashes are properly formatted
func CodeHashValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.ContractMetadataKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var metadata types.ContractMetadata
			if err := json.Unmarshal(iterator.Value(), &metadata); err != nil {
				continue
			}

			// Check code hash length (SHA256 = 32 bytes)
			if len(metadata.CodeHash) != 32 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"code-hash-validity",
					fmt.Sprintf("contract %s has invalid code hash length: %d (expected 32)",
						metadata.ContractAddress, len(metadata.CodeHash)),
				), true
			}
		}

		return "", false
	}
}

// ContractAddressValidityInvariant checks contract addresses are valid
func ContractAddressValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.ContractAddressIndexKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			address := string(iterator.Value())
			if _, err := sdk.AccAddressFromBech32(address); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"contract-address-validity",
					fmt.Sprintf("invalid contract address in index: %s", address),
				), true
			}
		}

		return "", false
	}
}

// VersionConsistencyInvariant checks that version information is consistent
func VersionConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := ctx.KVStore(k.storeKey)
		prefixStore := storeprefix.NewStore(store, types.ContractInfoPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var info pb.ContractInfo
			if err := k.cdc.Unmarshal(iterator.Value(), &info); err != nil {
				continue
			}

			// Check metadata - always present as value type
			// Check version format (should be semver-like)
			if len(info.Metadata.Version) > 100 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"version-consistency",
					fmt.Sprintf("contract %s has overly long version string: %d chars",
						info.Address, len(info.Metadata.Version)),
				), true
			}

			// If updated, updated_at should be after created_at
			// Both are time.Time values (not pointers)
			if !info.UpdatedAt.IsZero() && !info.CreatedAt.IsZero() {
				if info.UpdatedAt.Before(info.CreatedAt) {
					return sdk.FormatInvariant(
						types.ModuleName,
						"version-consistency",
						fmt.Sprintf("contract %s updated_at is before created_at",
							info.Address),
					), true
				}
			}
		}

		return "", false
	}
}
