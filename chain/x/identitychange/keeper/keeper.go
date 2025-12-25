// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	storetypes "cosmossdk.io/store/types"

	"github.com/aequitas/aura/chain/x/identitychange/params"
	"github.com/aequitas/aura/chain/x/identitychange/types"
)

// Keeper manages the state of the identitychange module using persistent KV store.
// All state is stored deterministically in the KV store to ensure consensus safety.
type Keeper struct {
	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	paramsStore  *params.Store
	authority    string
	logger       log.Logger
}

// NewKeeper creates a new Keeper instance with persistent KV store.
// All state is persisted to the KV store - no in-memory maps are used.
func NewKeeper(
	storeService store.KVStoreService,
	cdc codec.BinaryCodec,
	paramsStore *params.Store,
	authority string,
	logger log.Logger,
) *Keeper {
	if paramsStore == nil {
		paramsStore = params.NewStore(types.DefaultParams())
	}

	return &Keeper{
		storeService: storeService,
		cdc:          cdc,
		paramsStore:  paramsStore,
		authority:    authority,
		logger:       logger,
	}
}

// GetAuthority returns the governance module account address
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// GetParams returns the current module parameters
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	if k.paramsStore != nil {
		return k.paramsStore.GetParams(), nil
	}
	return types.DefaultParams(), nil
}

// SetParams sets new module parameters
func (k *Keeper) SetParams(params types.Params) error {
	if k.paramsStore == nil {
		return fmt.Errorf("params store not initialized")
	}
	return k.paramsStore.SetParams(params)
}

// ============================
// IDENTITY RECORD MANAGEMENT
// ============================

// GetIdentityRecord retrieves an identity record from KV store
func (k *Keeper) GetIdentityRecord(ctx sdk.Context, did string) (types.IdentityRecord, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte(types.RecordStoreKey(did)))
	if err != nil || bz == nil {
		return types.IdentityRecord{}, false
	}

	var record types.IdentityRecord
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return types.IdentityRecord{}, false
	}

	return record, true
}

// SetIdentityRecord stores an identity record to KV store
func (k *Keeper) SetIdentityRecord(ctx sdk.Context, record types.IdentityRecord) error {
	if record.Did == "" {
		return fmt.Errorf("DID is required")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return fmt.Errorf("failed to marshal identity record: %w", err)
	}

	if err := store.Set([]byte(types.RecordStoreKey(record.Did)), bz); err != nil {
		return fmt.Errorf("failed to store identity record: %w", err)
	}

	return nil
}

// ============================
// REQUEST MANAGEMENT
// ============================

// GetRequest retrieves an identity change request from KV store
func (k *Keeper) GetRequest(ctx sdk.Context, requestID string) (types.IdentityChangeRequest, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte(types.RequestStoreKey(requestID)))
	if err != nil || bz == nil {
		return types.IdentityChangeRequest{}, false
	}

	var request types.IdentityChangeRequest
	if err := k.cdc.Unmarshal(bz, &request); err != nil {
		return types.IdentityChangeRequest{}, false
	}

	return request, true
}

// SetRequest stores an identity change request to KV store
func (k *Keeper) SetRequest(ctx sdk.Context, request types.IdentityChangeRequest) error {
	if request.RequestId == "" {
		return fmt.Errorf("request ID is required")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if err := store.Set([]byte(types.RequestStoreKey(request.RequestId)), bz); err != nil {
		return fmt.Errorf("failed to store request: %w", err)
	}

	return nil
}

// CreateRequest creates a new identity change request in KV store
func (k *Keeper) CreateRequest(ctx sdk.Context, request types.IdentityChangeRequest) (types.IdentityChangeRequest, error) {
	if k.IsSuspended(ctx) {
		return types.IdentityChangeRequest{}, fmt.Errorf("identity change requests are suspended")
	}

	if request.RequestId == "" {
		return types.IdentityChangeRequest{}, fmt.Errorf("request id required")
	}

	count := k.countRequests(ctx, request.Requester)
	params, _ := k.GetParams(ctx)
	if int32(count) >= params.MaxRequestsPerWalletPerMonth {
		return types.IdentityChangeRequest{}, fmt.Errorf("request limit exceeded for %s", request.Requester)
	}

	request.Status = types.IdentityChangeStatusPendingVerification
	if err := k.SetRequest(ctx, request); err != nil {
		return types.IdentityChangeRequest{}, err
	}

	return request, nil
}

// countRequests counts requests for a requester from KV store
func (k *Keeper) countRequests(ctx sdk.Context, requester string) int {
	store := k.storeService.OpenKVStore(ctx)

	prefix := []byte(types.RequestStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return 0
	}
	defer iterator.Close()

	count := 0
	for ; iterator.Valid(); iterator.Next() {
		var req types.IdentityChangeRequest
		if err := k.cdc.Unmarshal(iterator.Value(), &req); err != nil {
			continue
		}
		if req.Requester == requester {
			count++
		}
	}

	return count
}

// SubmitProof submits proof for an identity change request
func (k *Keeper) SubmitProof(ctx sdk.Context, requestID string, assistant string, success bool, confidenceDelta int64, reason string) (types.IdentityChangeRequest, error) {
	request, ok := k.GetRequest(ctx, requestID)
	if !ok {
		return types.IdentityChangeRequest{}, fmt.Errorf("request %s not found", requestID)
	}

	if request.Status != types.IdentityChangeStatusPendingVerification {
		return request, fmt.Errorf("request %s already processed", requestID)
	}

	request.Assistant = assistant
	request.VerdictHeight = int64(ctx.BlockHeight())

	if success {
		request.Status = types.IdentityChangeStatusReadyToApply
	} else {
		request.Status = types.IdentityChangeStatusRejected
		request.Reason = reason
	}

	if err := k.SetRequest(ctx, request); err != nil {
		return types.IdentityChangeRequest{}, err
	}

	return request, nil
}

// ApplyChange applies an identity change
func (k *Keeper) ApplyChange(ctx sdk.Context, requestID string) (types.IdentityRecord, error) {
	request, ok := k.GetRequest(ctx, requestID)
	if !ok {
		return types.IdentityRecord{}, fmt.Errorf("request %s missing", requestID)
	}

	if request.Status != types.IdentityChangeStatusReadyToApply {
		return types.IdentityRecord{}, fmt.Errorf("request %s not ready", requestID)
	}

	record, _ := k.GetIdentityRecord(ctx, request.TargetDid)
	prevScore := record.ConfidenceScore

	params, _ := k.GetParams(ctx)
	if int64(prevScore) < int64(params.MinConfidenceAfterChange) {
		record.ConfidenceScore = int64(params.MinConfidenceAfterChange)
	}

	record.Did = request.TargetDid
	record.Owner = request.Requester
	record.MetadataHash = request.RequestMetaHash
	record.LatestIrVersion = request.IrId
	record.LastChangedHeight++
	record.Status = types.IdentityChangeStatusApplied

	if err := k.SetIdentityRecord(ctx, record); err != nil {
		return types.IdentityRecord{}, err
	}

	// Add to history
	history := types.IdentityChangeHistory{
		RequestId:           request.RequestId,
		TargetDid:           request.TargetDid,
		PrevConfidenceScore: prevScore,
		NewConfidenceScore:  record.ConfidenceScore,
		TransitionReason:    "applied",
		ChangedHeight:       record.LastChangedHeight,
	}

	if err := k.AddHistory(ctx, history); err != nil {
		return types.IdentityRecord{}, err
	}

	request.Status = types.IdentityChangeStatusApplied
	if err := k.SetRequest(ctx, request); err != nil {
		return types.IdentityRecord{}, err
	}

	return record, nil
}

// RejectChange rejects an identity change request
func (k *Keeper) RejectChange(ctx sdk.Context, requestID, reason string) (types.IdentityChangeRequest, error) {
	request, ok := k.GetRequest(ctx, requestID)
	if !ok {
		return types.IdentityChangeRequest{}, fmt.Errorf("request %s missing", requestID)
	}

	if request.Status == types.IdentityChangeStatusApplied {
		return request, fmt.Errorf("cannot reject applied request %s", requestID)
	}

	request.Status = types.IdentityChangeStatusRejected
	request.Reason = reason

	if err := k.SetRequest(ctx, request); err != nil {
		return types.IdentityChangeRequest{}, err
	}

	return request, nil
}

// ============================
// HISTORY MANAGEMENT
// ============================

// AddHistory adds an identity change history entry to KV store
func (k *Keeper) AddHistory(ctx sdk.Context, history types.IdentityChangeHistory) error {
	store := k.storeService.OpenKVStore(ctx)

	// Use unique key combining DID and height
	key := fmt.Sprintf("%s%s/%d",
		types.HistoryStoreKeyPrefix,
		history.TargetDid,
		history.ChangedHeight)

	bz, err := k.cdc.Marshal(&history)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := store.Set([]byte(key), bz); err != nil {
		return fmt.Errorf("failed to store history: %w", err)
	}

	return nil
}

// ListHistory retrieves identity change history for a DID from KV store
func (k *Keeper) ListHistory(ctx sdk.Context, did string) []types.IdentityChangeHistory {
	store := k.storeService.OpenKVStore(ctx)

	prefix := []byte(types.HistoryStoreKeyPrefix + did + "/")
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return []types.IdentityChangeHistory{}
	}
	defer iterator.Close()

	entries := []types.IdentityChangeHistory{}
	for ; iterator.Valid(); iterator.Next() {
		var item types.IdentityChangeHistory
		if err := k.cdc.Unmarshal(iterator.Value(), &item); err != nil {
			continue
		}
		entries = append(entries, item)
	}

	return entries
}

// ============================
// SUSPENDED FLAG MANAGEMENT
// ============================

// IsSuspended checks if identity change requests are suspended
func (k *Keeper) IsSuspended(ctx sdk.Context) bool {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte(types.SuspendedKey))
	if err != nil || bz == nil {
		return false
	}
	return bz[0] == 0x01
}

// SetSuspended sets the suspended flag in KV store
func (k *Keeper) SetSuspended(ctx sdk.Context, suspended bool) error {
	store := k.storeService.OpenKVStore(ctx)
	var bz []byte
	if suspended {
		bz = []byte{0x01}
	} else {
		bz = []byte{0x00}
	}

	if err := store.Set([]byte(types.SuspendedKey), bz); err != nil {
		return fmt.Errorf("failed to set suspended flag: %w", err)
	}

	return nil
}

// ============================
// RECOVERY MANAGEMENT
// ============================
// NOTE: Future enhancement - Uncomment when IdentityRecovery proto type is defined

// GetRecovery retrieves identity recovery record from KV store
// func (k *Keeper) GetRecovery(ctx sdk.Context, did string) (types.IdentityRecovery, bool) {
// 	store := k.storeService.OpenKVStore(ctx)
// 	bz, err := store.Get([]byte(types.RecoveryStoreKey(did)))
// 	if err != nil || bz == nil {
// 		return types.IdentityRecovery{}, false
// 	}
//
// 	var recovery types.IdentityRecovery
// 	if err := k.cdc.Unmarshal(bz, &recovery); err != nil {
// 		return types.IdentityRecovery{}, false
// 	}
//
// 	return recovery, true
// }

// SetRecovery stores identity recovery record to KV store
// func (k *Keeper) SetRecovery(ctx sdk.Context, recovery types.IdentityRecovery) error {
// 	if recovery.Did == "" {
// 		return fmt.Errorf("DID is required")
// 	}
//
// 	store := k.storeService.OpenKVStore(ctx)
// 	bz, err := k.cdc.Marshal(&recovery)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal recovery: %w", err)
// 	}
//
// 	if err := store.Set([]byte(types.RecoveryStoreKey(recovery.Did)), bz); err != nil {
// 		return fmt.Errorf("failed to store recovery: %w", err)
// 	}
//
// 	return nil
// }

// ============================
// VERIFICATION MANAGEMENT
// ============================
// NOTE: Future enhancement - Uncomment when IdentityVerification proto type is defined

// GetVerification retrieves identity verification record from KV store
// func (k *Keeper) GetVerification(ctx sdk.Context, did string) (types.IdentityVerification, bool) {
// 	store := k.storeService.OpenKVStore(ctx)
// 	bz, err := store.Get([]byte(types.VerificationStoreKey(did)))
// 	if err != nil || bz == nil {
// 		return types.IdentityVerification{}, false
// 	}
//
// 	var verification types.IdentityVerification
// 	if err := k.cdc.Unmarshal(bz, &verification); err != nil {
// 		return types.IdentityVerification{}, false
// 	}
//
// 	return verification, true
// }

// SetVerification stores identity verification record to KV store
// func (k *Keeper) SetVerification(ctx sdk.Context, verification types.IdentityVerification) error {
// 	if verification.Did == "" {
// 		return fmt.Errorf("DID is required")
// 	}
//
// 	store := k.storeService.OpenKVStore(ctx)
// 	bz, err := k.cdc.Marshal(&verification)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal verification: %w", err)
// 	}
//
// 	if err := store.Set([]byte(types.VerificationStoreKey(verification.Did)), bz); err != nil {
// 		return fmt.Errorf("failed to store verification: %w", err)
// 	}
//
// 	return nil
// }

// ============================
// DELEGATION MANAGEMENT
// ============================
// NOTE: Future enhancement - Uncomment when IdentityDelegation, IdentityFederation, CrossChainIdentity proto types are defined
/*
// GetDelegation retrieves identity delegation record from KV store
func (k *Keeper) GetDelegation(ctx sdk.Context, did string) (types.IdentityDelegation, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte(types.DelegationStoreKey(did)))
	if err != nil || bz == nil {
		return types.IdentityDelegation{}, false
	}

	var delegation types.IdentityDelegation
	if err := k.cdc.Unmarshal(bz, &delegation); err != nil {
		return types.IdentityDelegation{}, false
	}

	return delegation, true
}

// SetDelegation stores identity delegation record to KV store
func (k *Keeper) SetDelegation(ctx sdk.Context, delegation types.IdentityDelegation) error {
	if delegation.Did == "" {
		return fmt.Errorf("DID is required")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&delegation)
	if err != nil {
		return fmt.Errorf("failed to marshal delegation: %w", err)
	}

	if err := store.Set([]byte(types.DelegationStoreKey(delegation.Did)), bz); err != nil {
		return fmt.Errorf("failed to store delegation: %w", err)
	}

	return nil
}

// ============================
// FEDERATION MANAGEMENT
// ============================

// GetFederation retrieves identity federation record from KV store
func (k *Keeper) GetFederation(ctx sdk.Context, did string) (types.IdentityFederation, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte(types.FederationStoreKey(did)))
	if err != nil || bz == nil {
		return types.IdentityFederation{}, false
	}

	var federation types.IdentityFederation
	if err := k.cdc.Unmarshal(bz, &federation); err != nil {
		return types.IdentityFederation{}, false
	}

	return federation, true
}

// SetFederation stores identity federation record to KV store
func (k *Keeper) SetFederation(ctx sdk.Context, federation types.IdentityFederation) error {
	if federation.Did == "" {
		return fmt.Errorf("DID is required")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&federation)
	if err != nil {
		return fmt.Errorf("failed to marshal federation: %w", err)
	}

	if err := store.Set([]byte(types.FederationStoreKey(federation.Did)), bz); err != nil {
		return fmt.Errorf("failed to store federation: %w", err)
	}

	return nil
}

// ============================
// CROSS-CHAIN LINK MANAGEMENT
// ============================

// GetCrossChainLink retrieves cross-chain identity link from KV store
func (k *Keeper) GetCrossChainLink(ctx sdk.Context, did string) (types.CrossChainIdentity, bool) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get([]byte(types.CrossChainLinkStoreKey(did)))
	if err != nil || bz == nil {
		return types.CrossChainIdentity{}, false
	}

	var link types.CrossChainIdentity
	if err := k.cdc.Unmarshal(bz, &link); err != nil {
		return types.CrossChainIdentity{}, false
	}

	return link, true
}

// SetCrossChainLink stores cross-chain identity link to KV store
func (k *Keeper) SetCrossChainLink(ctx sdk.Context, link types.CrossChainIdentity) error {
	if link.Did == "" {
		return fmt.Errorf("DID is required")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&link)
	if err != nil {
		return fmt.Errorf("failed to marshal cross-chain link: %w", err)
	}

	if err := store.Set([]byte(types.CrossChainLinkStoreKey(link.Did)), bz); err != nil {
		return fmt.Errorf("failed to store cross-chain link: %w", err)
	}

	return nil
}
*/

// ============================
// GENESIS MANAGEMENT
// ============================
// Note: InitGenesis and ExportGenesis are implemented in genesis.go
