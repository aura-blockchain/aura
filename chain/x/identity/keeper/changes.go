package keeper

import (
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// ============================================================================
// Identity Record Management
// ============================================================================

// SetIdentityRecord stores an identity record in the KVStore
func (k *Keeper) SetIdentityRecord(ctx sdk.Context, record *types.IdentityRecord) error {
	if record.Did == "" {
		return types.ErrInvalidDID.Wrap("DID cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal identity record: %w", err)
	}

	key := types.GetIdentityRecordKey(record.Did)
	return store.Set(key, bz)
}

// GetIdentityRecord retrieves an identity record from the KVStore
func (k *Keeper) GetIdentityRecord(ctx sdk.Context, did string) (*types.IdentityRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetIdentityRecordKey(did)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrIdentityNotFound.Wrapf("identity not found: %s", did)
	}

	var record types.IdentityRecord
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal identity record: %w", err)
	}
	return &record, nil
}

// GetAllIdentityRecords retrieves all identity records
func (k *Keeper) GetAllIdentityRecords(ctx sdk.Context) ([]*types.IdentityRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.IdentityRecordPrefix, storetypes.PrefixEndBytes(types.IdentityRecordPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var records []*types.IdentityRecord
	for ; iterator.Valid(); iterator.Next() {
		var record types.IdentityRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return nil, fmt.Errorf("failed to unmarshal identity record: %w", err)
		}
		records = append(records, &record)
	}
	return records, nil
}

// DeleteIdentityRecord removes an identity record
// DEPRECATED: Use EraseIdentity for GDPR compliance instead
func (k *Keeper) DeleteIdentityRecord(ctx sdk.Context, did string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetIdentityRecordKey(did)
	return store.Delete(key)
}

// EraseIdentity implements GDPR Right to Erasure
// Marks identity as erased while preserving audit trail (commitment)
func (k *Keeper) EraseIdentity(ctx sdk.Context, did, requester, reason string) error {
	// Retrieve existing identity record
	record, err := k.GetIdentityRecord(ctx, did)
	if err != nil {
		return types.ErrIdentityNotFound.Wrapf("identity not found: %s", did)
	}

	// Check authorization: requester must be owner or have admin permission
	if record.Address != requester {
		if err := k.RequirePermission(ctx, requester, types.PermissionManageIdentity); err != nil {
			return types.ErrUnauthorized.Wrapf("requester %s not authorized to erase identity %s", requester, did)
		}
	}

	// Check if already erased
	if record.Erased {
		return types.ErrIdentityAlreadyErased.Wrapf("identity %s already erased at %s", did, record.ErasedAt.AsTime())
	}

	now := ctx.BlockTime()

	// Mark as erased (keep commitment for audit trail)
	record.Erased = true
	record.ErasedAt = timestamppb.New(now)
	record.Status = types.IdentityStatusErased
	record.UpdatedAt = timestamppb.New(now)

	// Clear off-chain data reference (if any)
	// The off-chain storage system should delete actual PII separately
	record.OffChainDataRef = ""
	record.OffChainDataType = ""

	// Keep: DID, address, commitments, status, timestamps for audit
	// Erased: off-chain references

	if err := k.SetIdentityRecord(ctx, record); err != nil {
		return fmt.Errorf("failed to update erased identity record: %w", err)
	}

	// Log audit trail
	k.LogAudit(ctx, requester, "erase_identity", did, "success", map[string]string{
		"reason":    reason,
		"erased_at": now.Format(time.RFC3339),
	}, "")

	return nil
}

// VerifyPIICommitment verifies that PII data matches the stored commitment
// This allows verification without storing raw PII on-chain
func (k *Keeper) VerifyPIICommitment(ctx sdk.Context, did string, piiData map[string]string) (bool, error) {
	record, err := k.GetIdentityRecord(ctx, did)
	if err != nil {
		return false, err
	}

	if record.Erased {
		return false, types.ErrIdentityErased.Wrapf("identity %s has been erased", did)
	}

	if len(record.PiiCommitment) == 0 {
		return false, types.ErrNoCommitment.Wrap("no PII commitment stored")
	}

	// Compute commitment from provided PII data
	computed := types.ComputePIICommitment(piiData, record.CommitmentSalt)

	// Compare with stored commitment
	return string(computed) == string(record.PiiCommitment), nil
}

// UpdatePIICommitment updates the PII commitment for an identity
// This should be called when PII changes in off-chain storage
func (k *Keeper) UpdatePIICommitment(ctx sdk.Context, did, updater string, piiData map[string]string, offChainRef, offChainType string) error {
	record, err := k.GetIdentityRecord(ctx, did)
	if err != nil {
		return err
	}

	// Check authorization
	if record.Address != updater {
		if err := k.RequirePermission(ctx, updater, types.PermissionManageIdentity); err != nil {
			return types.ErrUnauthorized.Wrapf("updater %s not authorized", updater)
		}
	}

	if record.Erased {
		return types.ErrIdentityErased.Wrapf("cannot update erased identity %s", did)
	}

	// Generate new salt
	salt := types.GenerateCommitmentSalt()

	// Compute new commitment
	commitment := types.ComputePIICommitment(piiData, salt)

	// Update record
	record.PiiCommitment = commitment
	record.CommitmentSalt = salt
	record.OffChainDataRef = offChainRef
	record.OffChainDataType = offChainType
	record.UpdatedAt = timestamppb.New(ctx.BlockTime())

	if err := k.SetIdentityRecord(ctx, record); err != nil {
		return fmt.Errorf("failed to update PII commitment: %w", err)
	}

	// Log audit trail
	k.LogAudit(ctx, updater, "update_pii_commitment", did, "success", map[string]string{
		"off_chain_ref":  offChainRef,
		"off_chain_type": offChainType,
	}, "")

	return nil
}

// ============================================================================
// Change Request Management
// ============================================================================

// SetChangeRequest stores a change request in the KVStore
func (k *Keeper) SetChangeRequest(ctx sdk.Context, request *types.ChangeRequest) error {
	if request.Id == "" {
		return types.ErrInvalidChangeRequest.Wrap("request ID cannot be empty")
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal change request: %w", err)
	}

	key := types.GetChangeRequestKey(request.Id)
	return store.Set(key, bz)
}

// GetChangeRequest retrieves a change request from the KVStore
func (k *Keeper) GetChangeRequest(ctx sdk.Context, requestID string) (*types.ChangeRequest, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetChangeRequestKey(requestID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrChangeRequestNotFound.Wrapf("change request not found: %s", requestID)
	}

	var request types.ChangeRequest
	if err := k.cdc.Unmarshal(bz, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal change request: %w", err)
	}
	return &request, nil
}

// GetAllChangeRequests retrieves all change requests
func (k *Keeper) GetAllChangeRequests(ctx sdk.Context) ([]*types.ChangeRequest, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ChangeRequestPrefix, storetypes.PrefixEndBytes(types.ChangeRequestPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var requests []*types.ChangeRequest
	for ; iterator.Valid(); iterator.Next() {
		var request types.ChangeRequest
		if err := k.cdc.Unmarshal(iterator.Value(), &request); err != nil {
			return nil, fmt.Errorf("failed to unmarshal change request: %w", err)
		}
		requests = append(requests, &request)
	}
	return requests, nil
}

// DeleteChangeRequest removes a change request
func (k *Keeper) DeleteChangeRequest(ctx sdk.Context, requestID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetChangeRequestKey(requestID)
	return store.Delete(key)
}

// CreateChangeRequest creates a new identity change request
func (k *Keeper) CreateChangeRequest(ctx sdk.Context, requester, targetDID, irID, metadataHash string) (*types.ChangeRequest, error) {
	// Check if identity changes are suspended
	if k.IsSuspended(ctx) {
		return nil, types.ErrIdentityChangeSuspended.Wrap("identity changes are currently suspended")
	}

	// Check request limit
	params, _ := k.GetParams(ctx)
	count := k.countChangeRequests(ctx, requester)
	if uint32(count) >= uint32(params.Change.MaxRequestsPerWalletPerMonth) {
		return nil, types.ErrChangeRequestLimitExceeded.Wrapf("request limit exceeded for %s", requester)
	}

	// Generate request ID
	store := k.storeService.OpenKVStore(ctx)
	counterBz, _ := store.Get(types.ChangeRequestCounterPrefix)
	var requestID uint64 = 1
	if counterBz != nil {
		requestID = sdk.BigEndianToUint64(counterBz)
	}

	now := ctx.BlockTime()
	nowProto := timestamppb.New(now)
	request := &types.ChangeRequest{
		Id:           fmt.Sprintf("req-%d", requestID),
		Requester:    requester,
		Did:          targetDID,
		IrId:         irID,
		Status:       types.ChangeStatusPending,
		RequestedAt:  nowProto,
		ChangeType:   types.ChangeTypeUpdateMetadata,
		ProofHash:    metadataHash,
	}

	if err := k.SetChangeRequest(ctx, request); err != nil {
		return nil, err
	}

	// Increment counter
	store.Set(types.ChangeRequestCounterPrefix, sdk.Uint64ToBigEndian(requestID+1))

	// Log audit trail
	k.LogAudit(ctx, requester, "create_change_request", targetDID, "success", map[string]string{
		"request_id": request.Id,
		"ir_id":      irID,
	}, "")

	return request, nil
}

// SubmitVerification submits verification for an identity change request
func (k *Keeper) SubmitVerification(ctx sdk.Context, requestID, assistant string, approved bool, reason string) (*types.ChangeRequest, error) {
	request, err := k.GetChangeRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if request.Status != types.ChangeStatusPending {
		return nil, types.ErrChangeRequestInvalid.Wrapf("request %s not pending verification", requestID)
	}

	// Check verifier has permission
	if err := k.RequirePermission(ctx, assistant, types.PermissionVerifyIdentity); err != nil {
		return nil, err
	}

	// Check if the identity being verified has any revoked credentials
	// This prevents verification of identities with compromised credentials
	identity, err := k.GetIdentityRecord(ctx, request.Did)
	if err == nil && identity != nil {
		// Check all verification methods for revocation
		for _, verificationMethod := range identity.VerificationMethods {
			if k.IsCredentialRevoked(ctx, verificationMethod) {
				return nil, types.ErrCredentialRevoked.Wrapf(
					"identity %s has revoked verification method %s",
					request.Did, verificationMethod,
				)
			}
		}
	}

	request.Assistant = assistant
	request.VerdictHeight = ctx.BlockHeight()

	if approved {
		request.Status = types.ChangeStatusApproved
	} else {
		request.Status = types.ChangeStatusRejected
	}

	if err := k.SetChangeRequest(ctx, request); err != nil {
		return nil, err
	}

	// Log audit trail
	k.LogAudit(ctx, assistant, "verify_change_request", request.Id, request.Status.String(), map[string]string{
		"approved": fmt.Sprintf("%t", approved),
		"reason":   reason,
	}, "")

	return request, nil
}

// ApplyChange applies an approved identity change
func (k *Keeper) ApplyChange(ctx sdk.Context, requestID, applier string) (*types.IdentityRecord, error) {
	// Check applier has permission
	if err := k.RequirePermission(ctx, applier, types.PermissionApproveChangeRequest); err != nil {
		return nil, err
	}

	request, err := k.GetChangeRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if request.Status != types.ChangeStatusApproved {
		return nil, types.ErrChangeRequestInvalid.Wrapf("request %s not ready to apply", requestID)
	}

	// Get or create identity record
	record, err := k.GetIdentityRecord(ctx, request.Did)
	if err != nil {
		// Create new record if not found
		now := ctx.BlockTime()
		record = &types.IdentityRecord{
			Did:       request.Did,
			Address:   request.Requester,
			Status:    types.IdentityStatusActive,
			CreatedAt: timestamppb.New(now),
			UpdatedAt: timestamppb.New(now),
		}
	}

	prevScore := record.ConfidenceScore

	// Update record
	params, _ := k.GetParams(ctx)
	record.Address = request.Requester
	record.MetadataHash = request.ProofHash
	record.LatestIrVersion = request.IrId
	record.Status = types.IdentityStatusActive
	record.UpdatedAt = timestamppb.New(ctx.BlockTime())

	// Apply minimum confidence score
	if params != nil && params.Change != nil && record.ConfidenceScore < int64(params.Change.MinConfidenceAfterChange) {
		record.ConfidenceScore = int64(params.Change.MinConfidenceAfterChange)
	}

	if err := k.SetIdentityRecord(ctx, record); err != nil {
		return nil, err
	}

	// Add to history
	history := &types.ChangeHistory{
		RequestId:           request.Id,
		TargetDid:           request.Did,
		PrevConfidenceScore: prevScore,
		NewConfidenceScore:  record.ConfidenceScore,
		TransitionReason:    "applied",
		ChangedHeight:       ctx.BlockHeight(),
		ChangedAt:           timestamppb.New(ctx.BlockTime()),
	}

	if err := k.SetChangeHistory(ctx, history); err != nil {
		return nil, err
	}

	// Update request status
	request.Status = types.ChangeStatusExecuted
	
	if err := k.SetChangeRequest(ctx, request); err != nil {
		return nil, err
	}

	// Log audit trail
	k.LogAudit(ctx, applier, "apply_change_request", request.Id, "success", map[string]string{
		"did":              request.Did,
		"prev_score":       fmt.Sprintf("%d", prevScore),
		"new_score":        fmt.Sprintf("%d", record.ConfidenceScore),
	}, "")

	return record, nil
}

// RejectChange rejects an identity change request
func (k *Keeper) RejectChange(ctx sdk.Context, requestID, rejecter, reason string) (*types.ChangeRequest, error) {
	// Check rejecter has permission
	if err := k.RequirePermission(ctx, rejecter, types.PermissionApproveChangeRequest); err != nil {
		return nil, err
	}

	request, err := k.GetChangeRequest(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if request.Status == types.ChangeStatusExecuted {
		return nil, types.ErrChangeRequestAlreadyApplied.Wrapf("cannot reject applied request %s", requestID)
	}

	request.Status = types.ChangeStatusRejected
	request.Reason = reason
	

	if err := k.SetChangeRequest(ctx, request); err != nil {
		return nil, err
	}

	// Log audit trail
	k.LogAudit(ctx, rejecter, "reject_change_request", request.Id, "success", map[string]string{
		"reason": reason,
	}, "")

	return request, nil
}

// countChangeRequests counts recent change requests for a requester
func (k *Keeper) countChangeRequests(ctx sdk.Context, requester string) int {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ChangeRequestPrefix, storetypes.PrefixEndBytes(types.ChangeRequestPrefix))
	if err != nil {
		return 0
	}
	defer iterator.Close()

	count := 0
	cutoff := ctx.BlockTime().Add(-30 * 24 * time.Hour) // 30 days

	for ; iterator.Valid(); iterator.Next() {
		var req types.ChangeRequest
		if err := k.cdc.Unmarshal(iterator.Value(), &req); err != nil {
			continue
		}
		if req.Requester == requester && req.RequestedAt.AsTime().After(cutoff) {
			count++
		}
	}

	return count
}

// ============================================================================
// Change History Management
// ============================================================================

// SetChangeHistory stores a change history entry
func (k *Keeper) SetChangeHistory(ctx sdk.Context, history *types.ChangeHistory) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal change history: %w", err)
	}

	key := types.GetChangeHistoryKey(history.TargetDid, uint64(history.ChangedHeight))
	return store.Set(key, bz)
}

// GetChangeHistory retrieves change history for a DID
func (k *Keeper) GetChangeHistory(ctx sdk.Context, did string) ([]*types.ChangeHistory, error) {
	store := k.storeService.OpenKVStore(ctx)

	// Create prefix for this DID's history
	prefix := append(types.ChangeHistoryPrefix, []byte(did)...)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var entries []*types.ChangeHistory
	for ; iterator.Valid(); iterator.Next() {
		var entry types.ChangeHistory
		if err := k.cdc.Unmarshal(iterator.Value(), &entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshal change history: %w", err)
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}

// GetAllChangeHistory retrieves all change history entries
func (k *Keeper) GetAllChangeHistory(ctx sdk.Context) ([]*types.ChangeHistory, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.ChangeHistoryPrefix, storetypes.PrefixEndBytes(types.ChangeHistoryPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var entries []*types.ChangeHistory
	for ; iterator.Valid(); iterator.Next() {
		var entry types.ChangeHistory
		if err := k.cdc.Unmarshal(iterator.Value(), &entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshal change history: %w", err)
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}

// ============================================================================
// Suspended Flag Management
// ============================================================================

// IsSuspended checks if identity changes are suspended
func (k *Keeper) IsSuspended(ctx sdk.Context) bool {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.SuspendedKey)
	if err != nil || bz == nil || len(bz) == 0 {
		return false
	}
	return bz[0] == 0x01
}

// SetSuspended sets the suspended flag
func (k *Keeper) SetSuspended(ctx sdk.Context, suspended bool) error {
	store := k.storeService.OpenKVStore(ctx)
	var bz []byte
	if suspended {
		bz = []byte{0x01}
	} else {
		bz = []byte{0x00}
	}
	return store.Set(types.SuspendedKey, bz)
}

// ============================================================================
// Additional Identity Record Types
// ============================================================================

// Recovery Records
func (k *Keeper) SetRecoveryRecord(ctx sdk.Context, record types.RecoveryRecord) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return fmt.Errorf("failed to marshal recovery record: %w", err)
	}
	key := types.GetRecoveryRecordKey(record.DID)
	return store.Set(key, bz)
}

func (k *Keeper) GetRecoveryRecord(ctx sdk.Context, did string) (types.RecoveryRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetRecoveryRecordKey(did)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.RecoveryRecord{}, types.ErrIdentityNotFound.Wrap("recovery record not found")
	}
	var record types.RecoveryRecord
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return types.RecoveryRecord{}, err
	}
	return record, nil
}

func (k *Keeper) GetAllRecoveryRecords(ctx sdk.Context) ([]types.RecoveryRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.RecoveryRecordPrefix, storetypes.PrefixEndBytes(types.RecoveryRecordPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var records []types.RecoveryRecord
	for ; iterator.Valid(); iterator.Next() {
		var record types.RecoveryRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

// Verification Records
func (k *Keeper) SetVerificationRecord(ctx sdk.Context, record types.VerificationRecord) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return fmt.Errorf("failed to marshal verification record: %w", err)
	}
	key := types.GetVerificationKey(record.DID)
	return store.Set(key, bz)
}

func (k *Keeper) GetVerificationRecord(ctx sdk.Context, did string) (types.VerificationRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetVerificationKey(did)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.VerificationRecord{}, types.ErrIdentityNotFound.Wrap("verification record not found")
	}
	var record types.VerificationRecord
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return types.VerificationRecord{}, err
	}
	return record, nil
}

func (k *Keeper) GetAllVerificationRecords(ctx sdk.Context) ([]types.VerificationRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.VerificationPrefix, storetypes.PrefixEndBytes(types.VerificationPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var records []types.VerificationRecord
	for ; iterator.Valid(); iterator.Next() {
		var record types.VerificationRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

// Delegation Records
func (k *Keeper) SetDelegationRecord(ctx sdk.Context, record types.DelegationRecord) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return fmt.Errorf("failed to marshal delegation record: %w", err)
	}
	key := types.GetDelegationKey(record.DID)
	return store.Set(key, bz)
}

func (k *Keeper) GetAllDelegationRecords(ctx sdk.Context) ([]types.DelegationRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.DelegationPrefix, storetypes.PrefixEndBytes(types.DelegationPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var records []types.DelegationRecord
	for ; iterator.Valid(); iterator.Next() {
		var record types.DelegationRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

// Federation Records
func (k *Keeper) SetFederationRecord(ctx sdk.Context, record types.FederationRecord) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return fmt.Errorf("failed to marshal federation record: %w", err)
	}
	key := types.GetFederationKey(record.DID)
	return store.Set(key, bz)
}

func (k *Keeper) GetAllFederationRecords(ctx sdk.Context) ([]types.FederationRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.FederationPrefix, storetypes.PrefixEndBytes(types.FederationPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var records []types.FederationRecord
	for ; iterator.Valid(); iterator.Next() {
		var record types.FederationRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

// Cross-Chain Links
func (k *Keeper) SetCrossChainLink(ctx sdk.Context, link types.CrossChainLink) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&link)
	if err != nil {
		return fmt.Errorf("failed to marshal cross-chain link: %w", err)
	}
	key := types.GetCrossChainLinkKey(link.DID)
	return store.Set(key, bz)
}

func (k *Keeper) GetAllCrossChainLinks(ctx sdk.Context) ([]types.CrossChainLink, error) {
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(types.CrossChainLinkPrefix, storetypes.PrefixEndBytes(types.CrossChainLinkPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var links []types.CrossChainLink
	for ; iterator.Valid(); iterator.Next() {
		var link types.CrossChainLink
		if err := k.cdc.Unmarshal(iterator.Value(), &link); err != nil {
			continue
		}
		links = append(links, link)
	}
	return links, nil
}
