// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	storetypes "cosmossdk.io/store/types"

	"github.com/aequitas/aura/chain/x/identitychange/types"
)

// RegisterInvariants registers all identitychange module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k *Keeper) {
	ir.RegisterRoute(types.ModuleName, "params-valid", ParamsInvariant(k))
	ir.RegisterRoute(types.ModuleName, "request-validity", RequestValidityInvariant(k))
	ir.RegisterRoute(types.ModuleName, "record-consistency", RecordConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "status-consistency", StatusConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "history-integrity", HistoryIntegrityInvariant(k))
}

// AllInvariants runs all invariants of the identitychange module
func AllInvariants(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		invariants := []sdk.Invariant{
			ParamsInvariant(k),
			RequestValidityInvariant(k),
			RecordConsistencyInvariant(k),
			StatusConsistencyInvariant(k),
			HistoryIntegrityInvariant(k),
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

// RequestValidityInvariant checks that all identity change requests are valid
func RequestValidityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := k.storeService.OpenKVStore(ctx)

		// Iterate through all requests
		prefix := []byte(types.RequestStoreKeyPrefix)
		iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"request-validity",
				fmt.Sprintf("failed to create iterator: %s", err.Error()),
			), true
		}
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var request types.IdentityChangeRequest
			if err := k.cdc.Unmarshal(iterator.Value(), &request); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"request-validity",
					fmt.Sprintf("failed to unmarshal identity change request: %s", err.Error()),
				), true
			}

			// Check request ID is not empty
			if request.RequestId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"request-validity",
					"identity change request has empty ID",
				), true
			}

			// Check requester is valid address
			if _, err := sdk.AccAddressFromBech32(request.Requester); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"request-validity",
					fmt.Sprintf("request %s has invalid requester: %s", request.RequestId, request.Requester),
				), true
			}

			// Check target DID is not empty
			if request.TargetDid == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"request-validity",
					fmt.Sprintf("request %s has empty target DID", request.RequestId),
				), true
			}

			// Check IR ID is not empty
			if request.IrId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"request-validity",
					fmt.Sprintf("request %s has empty IR ID", request.RequestId),
				), true
			}

			// Check status is valid
			if request.Status < types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_UNSPECIFIED ||
				request.Status > types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_SUSPENDED {
				return sdk.FormatInvariant(
					types.ModuleName,
					"request-validity",
					fmt.Sprintf("request %s has invalid status: %d", request.RequestId, request.Status),
				), true
			}

			// Verify height consistency
			if request.VerdictHeight > 0 && request.VerdictHeight < request.CreatedHeight {
				return sdk.FormatInvariant(
					types.ModuleName,
					"request-validity",
					fmt.Sprintf("request %s verdict height %d is before created height %d",
						request.RequestId, request.VerdictHeight, request.CreatedHeight),
				), true
			}
		}

		return "", false
	}
}

// RecordConsistencyInvariant checks identity record consistency
func RecordConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := k.storeService.OpenKVStore(ctx)

		// Iterate through all records
		prefix := []byte(types.RecordStoreKeyPrefix)
		iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"record-consistency",
				fmt.Sprintf("failed to create iterator: %s", err.Error()),
			), true
		}
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var record types.IdentityRecord
			if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"record-consistency",
					fmt.Sprintf("failed to unmarshal identity record: %s", err.Error()),
				), true
			}

			// Check DID is not empty
			if record.Did == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"record-consistency",
					"identity record has empty DID",
				), true
			}

			// Check owner is valid address
			if record.Owner != "" {
				if _, err := sdk.AccAddressFromBech32(record.Owner); err != nil {
					return sdk.FormatInvariant(
						types.ModuleName,
						"record-consistency",
						fmt.Sprintf("record %s has invalid owner: %s", record.Did, record.Owner),
					), true
				}
			}

			// Check confidence score is non-negative
			if record.ConfidenceScore < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"record-consistency",
					fmt.Sprintf("record %s has negative confidence score: %d", record.Did, record.ConfidenceScore),
				), true
			}

			// Check last changed height is non-negative
			if record.LastChangedHeight < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"record-consistency",
					fmt.Sprintf("record %s has negative last changed height: %d", record.Did, record.LastChangedHeight),
				), true
			}

			// Check status is valid
			if record.Status < types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_UNSPECIFIED ||
				record.Status > types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_SUSPENDED {
				return sdk.FormatInvariant(
					types.ModuleName,
					"record-consistency",
					fmt.Sprintf("record %s has invalid status: %d", record.Did, record.Status),
				), true
			}
		}

		return "", false
	}
}

// StatusConsistencyInvariant checks request status consistency
func StatusConsistencyInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := k.storeService.OpenKVStore(ctx)

		// Iterate through all requests
		prefix := []byte(types.RequestStoreKeyPrefix)
		iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"status-consistency",
				fmt.Sprintf("failed to create iterator: %s", err.Error()),
			), true
		}
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var request types.IdentityChangeRequest
			if err := k.cdc.Unmarshal(iterator.Value(), &request); err != nil {
				continue
			}

			// Requests with READY_TO_APPLY or APPLIED status should have assistant set
			if (request.Status == types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_READY_TO_APPLY ||
				request.Status == types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED) &&
				request.Assistant == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"status-consistency",
					fmt.Sprintf("request %s with status %d has empty assistant", request.RequestId, request.Status),
				), true
			}

			// Requests with READY_TO_APPLY or APPLIED status should have verdict height
			if (request.Status == types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_READY_TO_APPLY ||
				request.Status == types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED) &&
				request.VerdictHeight <= 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"status-consistency",
					fmt.Sprintf("request %s with status %d has invalid verdict height %d",
						request.RequestId, request.Status, request.VerdictHeight),
				), true
			}

			// Rejected requests should have reason
			if request.Status == types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_REJECTED &&
				request.Reason == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"status-consistency",
					fmt.Sprintf("rejected request %s has empty rejection reason", request.RequestId),
				), true
			}
		}

		return "", false
	}
}

// HistoryIntegrityInvariant checks history integrity
func HistoryIntegrityInvariant(k *Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		store := k.storeService.OpenKVStore(ctx)

		// Iterate through all history entries
		prefix := []byte(types.HistoryStoreKeyPrefix)
		iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName,
				"history-integrity",
				fmt.Sprintf("failed to create iterator: %s", err.Error()),
			), true
		}
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			var history types.IdentityChangeHistory
			if err := k.cdc.Unmarshal(iterator.Value(), &history); err != nil {
				return sdk.FormatInvariant(
					types.ModuleName,
					"history-integrity",
					fmt.Sprintf("failed to unmarshal history entry: %s", err.Error()),
				), true
			}

			// Check request ID is not empty
			if history.RequestId == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"history-integrity",
					"history entry has empty request ID",
				), true
			}

			// Check target DID is not empty
			if history.TargetDid == "" {
				return sdk.FormatInvariant(
					types.ModuleName,
					"history-integrity",
					fmt.Sprintf("history entry for request %s has empty target DID", history.RequestId),
				), true
			}

			// Check changed height is non-negative
			if history.ChangedHeight < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"history-integrity",
					fmt.Sprintf("history entry for request %s has negative changed height: %d",
						history.RequestId, history.ChangedHeight),
				), true
			}

			// Check confidence scores are non-negative
			if history.PrevConfidenceScore < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"history-integrity",
					fmt.Sprintf("history entry for request %s has negative prev confidence score: %d",
						history.RequestId, history.PrevConfidenceScore),
				), true
			}

			if history.NewConfidenceScore < 0 {
				return sdk.FormatInvariant(
					types.ModuleName,
					"history-integrity",
					fmt.Sprintf("history entry for request %s has negative new confidence score: %d",
						history.RequestId, history.NewConfidenceScore),
				), true
			}
		}

		return "", false
	}
}
