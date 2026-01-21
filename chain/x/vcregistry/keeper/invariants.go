// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all vcregistry module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "vc-consistency", VCConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "revocation-consistency", RevocationConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "vc-subject-integrity", VCSubjectIntegrityInvariant(k))
}

// AllInvariants runs all invariants of the vcregistry module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			VCConsistencyInvariant(k),
			RevocationConsistencyInvariant(k),
			VCSubjectIntegrityInvariant(k),
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
		params, _ := k.GetParams(ctx)
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

// VCConsistencyInvariant checks verifiable credential consistency
func VCConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		if k.store == nil {
			// Store not initialized, skip check
			return "", false
		}

		// Iterate through all VCs using the store iterator
		for _, vc := range k.store.iterateVCRecords(ctx) {
			// Check VC ID is not empty
			if vc.VcId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"vc-consistency",
					"VC has empty ID",
				), true
			}

			// Check holder address is valid
			if _, err := sdk.AccAddressFromBech32(vc.HolderAddress); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"vc-consistency",
					fmt.Sprintf("VC %s has invalid holder address: %s", vc.VcId, vc.HolderAddress),
				), true
			}

			// Check VC type is not unspecified
			if vc.VcType == types.VCType_VC_TYPE_UNSPECIFIED {
				return sdk.FormatInvariant(
					types.ModuleName,
					"vc-consistency",
					fmt.Sprintf("VC %s has unspecified type", vc.VcId),
				), true
			}

			// Check credential hash is not empty
			if len(vc.CredentialHash) == 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"vc-consistency",
					fmt.Sprintf("VC %s has empty credential hash", vc.VcId),
				), true
			}

			// Check timestamps
			if vc.IssuedAt == nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"vc-consistency",
					fmt.Sprintf("VC %s has nil issued_at", vc.VcId),
				), true
			}

			// Expiration should be after issuance
			if vc.ExpiresAt != nil && vc.IssuedAt != nil {
				expiresTime := time.Unix(vc.ExpiresAt.Seconds, int64(vc.ExpiresAt.Nanos))
				issuedTime := time.Unix(vc.IssuedAt.Seconds, int64(vc.IssuedAt.Nanos))
				if expiresTime.Before(issuedTime) {
					return sdk.FormatInvariant(
						types.ModuleName,
						"vc-consistency",
						fmt.Sprintf("VC %s expires before issuance", vc.VcId),
					), true
				}
			}

			// Revoked VCs should have revocation time
			if vc.Status == types.VCStatus_VC_STATUS_REVOKED {
				revRecord, found := k.GetRevocationRecord(ctx, vc.VcId)
				if !found || revRecord.RevokedAt == nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"vc-consistency",
						fmt.Sprintf("revoked VC %s has no revocation record", vc.VcId),
					), true
				}
			}
		}

		return "", false
	}
}

// RevocationConsistencyInvariant checks revocation list consistency
func RevocationConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		if k.store == nil {
			// Store not initialized, skip check
			return "", false
		}

		revocationList := k.GetRevocationList(ctx)
		if revocationList == nil {
			// No revocation list yet, this is valid
			return "", false
		}

		// Check that Merkle root is not empty if there are revocations
		if revocationList.TotalRevocations > 0 && len(revocationList.MerkleRoot) == 0 {
			return sdk.FormatInvariant(
				types.ModuleName,
				"revocation-consistency",
				"revocation list has entries but empty Merkle root",
			), true
		}

		// Check that last updated is not nil if there are revocations
		if revocationList.TotalRevocations > 0 && revocationList.LastUpdated == nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"revocation-consistency",
				"revocation list has entries but nil last_updated",
			), true
		}

		return "", false
	}
}

// VCSubjectIntegrityInvariant checks that all VCs reference existing DIDs
//
// CRITICAL SECURITY: This invariant ensures referential integrity between VCs and DIDs.
// Without this check, orphaned VCs could exist after DID deletion, leading to:
//   - Invalid credential references in the system
//   - Inability to verify credentials (missing DID document)
//   - Potential security vulnerabilities from dangling references
//
// The invariant validates that every VC's subject DID exists in the DID registry.
// This prevents the scenario where a DID is deleted but VCs still reference it.
//
// Returns:
//   - ("", false) if all VCs have valid subject DIDs
//   - (error message, true) if any VC references a non-existent DID
func VCSubjectIntegrityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		if k.store == nil {
			// Store not initialized, skip check
			return "", false
		}

		// Iterate through all VCs and verify their holder DIDs exist
		for _, vc := range k.store.iterateVCRecords(ctx) {
			// Check if the holder DID exists in the registry
			// VCs have a HolderDid field that should reference a valid DID
			if vc.HolderDid != "" {
				_, exists := k.GetDIDDocument(ctx, vc.HolderDid)
				if !exists {
					return sdk.FormatInvariant(
						types.ModuleName,
						"vc-subject-integrity",
						fmt.Sprintf("VC %s references non-existent DID %s", vc.VcId, vc.HolderDid),
					), true
				}
			}
		}

		return "", false
	}
}
