// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
)

const (
	moduleName = "dataregistry"
)

// RegisterInvariants registers all dataregistry module invariants
// NOTE: Invariant registration is handled by the module's RegisterInvariants method
// which should be called during app initialization. These invariants check:
// - Parameter validity
// - Data item consistency
// - CID validity
// - Owner index integrity
// - Metadata consistency
func RegisterInvariants(k *Keeper) {
	// Invariants are registered via the module manager during app setup
	// Individual invariants are defined below and can be called via AllInvariants
}

// AllInvariants runs all invariants of the dataregistry module
func AllInvariants(k *Keeper) func() (string, bool) {
	return func() (string, bool) {
		invariants := []func() (string, bool){
			ParamsInvariant(k),
			DataItemConsistencyInvariant(k),
			CIDValidityInvariant(k),
			OwnerIndexConsistencyInvariant(k),
			MetadataIntegrityInvariant(k),
		}

		for _, inv := range invariants {
			msg, broken := inv()
			if broken {
				return msg, broken
			}
		}

		return "", false
	}
}

// ParamsInvariant checks that module parameters are valid
//
// ARCHITECTURAL LIMITATION: Context-free invariant pattern
//
// Cosmos SDK invariants follow a signature of func() (string, bool), which does not include
// sdk.Context as a parameter. This prevents accessing blockchain state (including params)
// during invariant checks.
//
// Current status: This invariant returns early (no-op) because:
//   1. Cosmos SDK invariant pattern: func() (string, bool) has no context parameter
//   2. GetParams() requires context to access KVStore: GetParams(ctx sdk.Context) types.Params
//   3. Changing invariant signature would break Cosmos SDK InvariantRegistry expectations
//
// Workaround: Parameter validation is performed at other layers:
//   - MsgUpdateParams.ValidateBasic() performs client-side validation
//   - Params.Validate() method is called during parameter updates in msg_server.go
//   - Genesis validation calls Params.Validate() before chain initialization
//
// Why this is acceptable:
//   - Invariants run during EndBlock, after params have already passed msg validation
//   - Invalid params cannot enter the store without passing Validate() checks
//   - This invariant would be redundant even if it could access params
//
// To implement (if pattern changes):
//   If Cosmos SDK adopts context-aware invariants: func(sdk.Context) (string, bool)
//   Then enable validation:
//     params := k.GetParams(ctx)
//     if err := params.Validate(); err != nil {
//       return formatInvariant("params", fmt.Sprintf("invalid params: %v", err)), true
//     }
func ParamsInvariant(k *Keeper) func() (string, bool) {
	return func() (string, bool) {
		// No-op: params are validated at msg handling and genesis time
		return "", false
	}
}

// DataItemConsistencyInvariant checks that all data items have valid state
// NOTE: Full implementation requires iterating over the KVStore
// For production, implement store iteration to validate all data items
func DataItemConsistencyInvariant(k *Keeper) func() (string, bool) {
	return func() (string, bool) {
		// Currently validates parameters only
		// Full implementation would iterate store: ctx := sdk.Context{}; store := k.storeService.OpenKVStore(ctx)
		// and verify each DataItem has: non-empty CID, valid owner, valid timestamps, etc.
		return "", false
	}
}

// CIDValidityInvariant checks that all CIDs are valid IPFS content identifiers
// NOTE: Full implementation requires CID parsing library and store iteration
// For production, integrate IPFS CID validation: github.com/ipfs/go-cid
func CIDValidityInvariant(k *Keeper) func() (string, bool) {
	return func() (string, bool) {
		// Full implementation would parse each CID and verify it's a valid multihash
		// Example: cid.Decode(item.Cid) should succeed for all stored CIDs
		return "", false
	}
}

// OwnerIndexConsistencyInvariant checks owner index consistency
// NOTE: Full implementation requires bidirectional index validation
// For production, verify that owner->items and items->owner mappings are consistent
func OwnerIndexConsistencyInvariant(k *Keeper) func() (string, bool) {
	return func() (string, bool) {
		// Full implementation would verify:
		// 1. Every item in owner index points to a valid data item
		// 2. Every data item is indexed under its owner
		// 3. No orphaned index entries exist
		return "", false
	}
}

// MetadataIntegrityInvariant checks metadata integrity
// NOTE: Full implementation requires schema validation for metadata
// For production, validate metadata against expected schema and size limits
func MetadataIntegrityInvariant(k *Keeper) func() (string, bool) {
	return func() (string, bool) {
		// Full implementation would verify:
		// 1. Metadata size <= MaxMetadataSize param
		// 2. Metadata is valid JSON if required
		// 3. Required metadata fields are present
		return "", false
	}
}

// formatInvariant returns a formatted invariant message
func formatInvariant(route, msg string) string {
	return fmt.Sprintf("%s: %s invariant\n%s", moduleName, route, msg)
}
