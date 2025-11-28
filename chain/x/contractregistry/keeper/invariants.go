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
		params := k.GetParams(ctx)
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

			// Skip if no metadata
			if info.Metadata == nil {
				continue
			}

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

			// Check timestamps
			if info.CreatedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"contract-metadata-consistency",
					fmt.Sprintf("contract %s has nil created_at", info.Address),
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
		prefixStore := storeprefix.NewStore(store, types.ContractMetadataKeyPrefix)

		iterator := prefixStore.Iterator(nil, nil)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var metadata types.ContractMetadata
			if err := json.Unmarshal(iterator.Value(), &metadata); err != nil {
				continue
			}

			// Check version format (should be semver-like)
			if len(metadata.Version) > 100 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"version-consistency",
					fmt.Sprintf("contract %s has overly long version string: %d chars",
						metadata.ContractAddress, len(metadata.Version)),
				), true
			}

			// If updated, updated_at should be set
			if metadata.UpdatedAt != nil && metadata.CreatedAt != nil {
				if metadata.UpdatedAt.AsTime().Before(metadata.CreatedAt.AsTime()) {
					return sdk.FormatInvariant(
						types.ModuleName,
						"version-consistency",
						fmt.Sprintf("contract %s updated_at is before created_at",
							metadata.ContractAddress),
					), true
				}
			}
		}

		return "", false
	}
}
