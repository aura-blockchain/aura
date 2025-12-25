// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/vcregistry/params"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
)

// ConfidenceScoreKeeper defines the interface for interacting with the confidencescore module
type ConfidenceScoreKeeper interface {
	GetUserScore(walletAddr string) (uint64, bool)
	HasCompletedIR(walletAddr, irID string) bool
	GetArenaScore(walletAddr, arena string) (uint64, error)
	GetAnchorInfo(walletAddr string) (interface{}, bool)
	IsVerified(walletAddr string) bool
}

// Keeper manages the state of the vcregistry module
// ALL STATE IS PERSISTED IN KV STORE - NO IN-MEMORY FALLBACKS
// Keeper is safe for concurrent use because Cosmos SDK processes transactions serially
type Keeper struct {
	store       *Store
	paramsStore *params.Store
	csKeeper    ConfidenceScoreKeeper
	authority   string // governance module account address

	// Test-only fields for deterministic time/height
	testCurrentTime   int64
	testCurrentHeight uint64
}

// NewKeeper creates a new Keeper instance
// All state must be in KV store. In-memory fallbacks have been removed.
func NewKeeper(store *params.Store, authority string) *Keeper {
	if store == nil {
		store = params.NewStore(*types.DefaultParams())
	}
	return &Keeper{
		store:       nil, // Set via WithStore
		paramsStore: store,
		authority:   authority,
	}
}

// WithStore wires a Cosmos KV store + codec for on-chain persistence.
// This MUST be called before the keeper is used. No in-memory fallback.
func (k *Keeper) WithStore(storeKey storetypes.StoreKey, cdc codec.BinaryCodec) *Keeper {
	s := NewStore(storeKey, cdc)
	k.store = &s
	return k
}

// requireStore panics if store is not initialized (production safety check)
func (k *Keeper) requireStore() {
	if k.store == nil {
		panic("vcregistry keeper: KV store not initialized. Call WithStore() first.")
	}
}

// sdkContext unwraps an SDK context; panics if unavailable (production safety check)
func (k *Keeper) sdkContext(ctx context.Context) sdk.Context {
	k.requireStore()
	return sdk.UnwrapSDKContext(ctx)
}

// GetAuthority returns the governance module account address
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// SetConfidenceScoreKeeper sets the confidence score keeper for validation
func (k *Keeper) SetConfidenceScoreKeeper(keeper ConfidenceScoreKeeper) {
	k.csKeeper = keeper
}

// getCurrentTime extracts current block time from context (consensus-safe)
func (k *Keeper) getCurrentTime(ctx context.Context) int64 {
	if k.testCurrentTime != 0 {
		return k.testCurrentTime
	}
	return sdk.UnwrapSDKContext(ctx).BlockTime().Unix()
}

// getCurrentHeight extracts current block height from context (consensus-safe)
func (k *Keeper) getCurrentHeight(ctx context.Context) uint64 {
	if k.testCurrentHeight != 0 {
		return k.testCurrentHeight
	}
	return uint64(sdk.UnwrapSDKContext(ctx).BlockHeight())
}

// SetCurrentTime sets the current time for testing purposes
// ONLY FOR TESTING - allows deterministic time in unit tests
func (k *Keeper) SetCurrentTime(t int64) {
	k.testCurrentTime = t
}

// SetCurrentHeight sets the current block height for testing purposes
// ONLY FOR TESTING - allows deterministic height in unit tests
func (k *Keeper) SetCurrentHeight(height uint64) {
	k.testCurrentHeight = height
}

// GetParams returns the current module parameters
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	if k.paramsStore != nil {
		return k.paramsStore.GetParams(), nil
	}
	return *types.DefaultParams(), nil
}

// SetParams sets new module parameters
func (k *Keeper) SetParams(params types.Params) error {
	if k.paramsStore == nil {
		return fmt.Errorf("params store not initialized")
	}
	return k.paramsStore.SetParams(params)
}

// ============================
// VC RECORD MANAGEMENT
// ============================

// GetVCRecord retrieves a VC record by ID (KV store only)
func (k *Keeper) GetVCRecord(ctx context.Context, vcID string) (types.VCRecord, bool) {
	k.requireStore()
	return k.store.getVCRecord(ctx, vcID)
}

// SetVCRecord stores a VC record (KV store only)
func (k *Keeper) SetVCRecord(ctx context.Context, record types.VCRecord) error {
	if record.VcId == "" {
		return types.ErrInvalidVCID
	}
	if record.HolderAddress == "" {
		return types.ErrInvalidHolderAddress
	}

	k.requireStore()

	k.store.setVCRecord(ctx, record)
	k.store.appendUserVC(ctx, record.HolderAddress, record.VcId)

	return nil
}

// ListUserVCs returns all VCs for a user (KV store only)
func (k *Keeper) ListUserVCs(ctx context.Context, holderAddress string, statusFilter types.VCStatus, typeFilter types.VCType) []types.VCRecord {
	k.requireStore()

	vcIDs := k.store.listUserVCs(ctx, holderAddress)
	if len(vcIDs) == 0 {
		return []types.VCRecord{}
	}

	vcs := []types.VCRecord{}
	for _, vcID := range vcIDs {
		vc, ok := k.store.getVCRecord(ctx, vcID)
		if !ok {
			continue
		}
		// Apply filters
		if statusFilter != types.VCStatus_VC_STATUS_UNSPECIFIED && vc.Status != statusFilter {
			continue
		}
		if typeFilter != types.VCTypeUnspecified && vc.VcType != typeFilter {
			continue
		}
		vcs = append(vcs, vc)
	}

	return vcs
}

// ============================
// VC STATUS CHECKS
// ============================

// CheckVCStatus checks the current status of a VC
func (k *Keeper) CheckVCStatus(ctx context.Context, vcID string) (types.VCStatus, bool, error) {
	vc, ok := k.GetVCRecord(ctx, vcID)
	if !ok {
		return types.VCStatus_VC_STATUS_UNSPECIFIED, false, types.ErrVCNotFound
	}

	// Check expiration
	if vc.ExpiresAt != nil && vc.ExpiresAt.Seconds <= k.getCurrentTime(ctx) {
		// Auto-update status to expired
		vc.Status = types.VCStatus_VC_STATUS_EXPIRED
		_ = k.SetVCRecord(ctx, vc)
		return types.VCStatus_VC_STATUS_EXPIRED, false, nil
	}

	valid := vc.Status == types.VCStatus_VC_STATUS_ACTIVE
	return vc.Status, valid, nil
}

// IsVCValid checks if a VC is currently valid (active and not expired)
func (k *Keeper) IsVCValid(ctx context.Context, vcID string) bool {
	status, valid, err := k.CheckVCStatus(ctx, vcID)
	return err == nil && valid && status == types.VCStatus_VC_STATUS_ACTIVE
}

// ============================
// REVOCATION MANAGEMENT
// ============================

// RevokeVC revokes a verifiable credential (KV store only)
func (k *Keeper) RevokeVC(ctx context.Context, vcID string, reason types.RevocationReason, revoker string, evidence string) error {
	vc, ok := k.GetVCRecord(ctx, vcID)
	if !ok {
		return types.ErrVCNotFound
	}

	if vc.Status == types.VCStatus_VC_STATUS_REVOKED {
		return types.ErrVCAlreadyRevoked
	}

	k.requireStore()

	// Create revocation record
	revRecord := types.RevocationRecord{
		VcId:          vcID,
		RevokedAt:     &gogotypes.Timestamp{Seconds: k.getCurrentTime(ctx), Nanos: 0},
		RevokedHeight: k.getCurrentHeight(ctx),
		Reason:        reason,
		Revoker:       revoker,
		Evidence:      evidence,
	}

	// Update VC status
	vc.Status = types.VCStatus_VC_STATUS_REVOKED
	k.store.setVCRecord(ctx, vc)

	// Store revocation
	k.store.setRevocationRecord(ctx, revRecord)

	// Update Merkle tree
	k.updateRevocationMerkleRoot(ctx, vcID, revRecord)

	return nil
}

// GetRevocationRecord retrieves a revocation record
func (k *Keeper) GetRevocationRecord(ctx context.Context, vcID string) (types.RevocationRecord, bool) {
	k.requireStore()
	return k.store.getRevocationRecord(ctx, vcID)
}

// IsRevoked checks if a VC is revoked
func (k *Keeper) IsRevoked(ctx context.Context, vcID string) bool {
	_, ok := k.GetRevocationRecord(ctx, vcID)
	return ok
}

// updateRevocationMerkleRoot updates the Merkle root after a revocation
func (k *Keeper) updateRevocationMerkleRoot(ctx context.Context, vcID string, record types.RevocationRecord) {
	// Get current revocation list from KV store
	revocationList, ok := k.store.getRevocationList(ctx)
	if !ok {
		revocationList = types.RevocationList{
			MerkleRoot:        []byte{},
			TotalRevocations:  0,
			LastUpdatedHeight: 0,
			LastUpdated:       nil,
		}
	}

	oldRoot := append([]byte(nil), revocationList.MerkleRoot...)

	// Simple hash-based Merkle root (in production, use proper Merkle tree)
	h := sha256.New()
	h.Write([]byte(vcID))
	if record.RevokedAt != nil {
		h.Write([]byte(fmt.Sprintf("%d", record.RevokedAt.Seconds)))
	}
	h.Write([]byte(fmt.Sprintf("%d", record.Reason)))

	newHash := h.Sum(nil)

	// XOR with current root for simplicity (proper Merkle tree in production)
	if len(revocationList.MerkleRoot) == 0 {
		revocationList.MerkleRoot = newHash
	} else {
		combined := make([]byte, 32)
		for i := 0; i < 32; i++ {
			combined[i] = revocationList.MerkleRoot[i] ^ newHash[i]
		}
		revocationList.MerkleRoot = combined
	}

	revocationList.TotalRevocations++
	revocationList.LastUpdatedHeight = k.getCurrentHeight(ctx)
	revocationList.LastUpdated = &gogotypes.Timestamp{Seconds: k.getCurrentTime(ctx), Nanos: 0}

	k.store.setRevocationList(ctx, revocationList)

	newRoot := append([]byte(nil), revocationList.MerkleRoot...)
	k.emitMerkleRootUpdatedEvent(ctx, oldRoot, newRoot)
}

func (k *Keeper) emitMerkleRootUpdatedEvent(ctx context.Context, oldRoot, newRoot []byte) {
	if len(oldRoot) == 0 && len(newRoot) == 0 {
		return
	}
	sdkCtx := k.sdkContext(ctx)

	revocationList, _ := k.store.getRevocationList(ctx)

	attrs := types.NewEventMerkleRootUpdated(
		hex.EncodeToString(oldRoot),
		hex.EncodeToString(newRoot),
		fmt.Sprintf("%d", revocationList.TotalRevocations),
		fmt.Sprintf("%d", sdkCtx.BlockHeight()),
	)
	// Sort keys for deterministic iteration order (consensus-critical)
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	eventAttrs := make([]sdk.Attribute, 0, len(attrs))
	for _, key := range keys {
		value := attrs[key]
		if value == "" {
			continue
		}
		eventAttrs = append(eventAttrs, sdk.NewAttribute(key, value))
	}
	if len(eventAttrs) == 0 {
		return
	}
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeMerkleRootUpdated, eventAttrs...))
}

// GetRevocationList returns the current revocation list
func (k *Keeper) GetRevocationList(ctx context.Context) *types.RevocationList {
	k.requireStore()
	if list, ok := k.store.getRevocationList(ctx); ok {
		return &list
	}
	// Return empty list if not found
	return &types.RevocationList{
		MerkleRoot:        []byte{},
		TotalRevocations:  0,
		LastUpdatedHeight: 0,
		LastUpdated:       nil,
	}
}

// ============================
// DID MANAGEMENT
// ============================

// RegisterDID registers a new DID document
func (k *Keeper) RegisterDID(ctx context.Context, did string, controller string, verificationMethods []*types.VerificationMethod, metadataURI string) error {
	if did == "" {
		return types.ErrInvalidDID
	}
	if controller == "" {
		return types.ErrInvalidHolderAddress
	}

	k.requireStore()

	// Check if DID already exists
	if _, ok := k.store.getDIDDocument(ctx, did); ok {
		return types.ErrDIDAlreadyExists
	}

	doc := types.DIDDocument{
		Did:                 did,
		Controller:          controller,
		VerificationMethods: verificationMethods,
		CredentialIds:       []string{},
		Created:             &gogotypes.Timestamp{Seconds: k.getCurrentTime(ctx), Nanos: 0},
		Updated:             &gogotypes.Timestamp{Seconds: k.getCurrentTime(ctx), Nanos: 0},
		MetadataUri:         metadataURI,
		ServiceEndpoints:    make(map[string]string),
	}

	k.store.setDIDDocument(ctx, doc)
	k.store.appendAddressDID(ctx, controller, did)

	return nil
}

// GetDIDDocument retrieves a DID document
func (k *Keeper) GetDIDDocument(ctx context.Context, did string) (types.DIDDocument, bool) {
	k.requireStore()
	return k.store.getDIDDocument(ctx, did)
}

// UpdateDIDDocument updates a DID document
func (k *Keeper) UpdateDIDDocument(ctx context.Context, did string, verificationMethods []*types.VerificationMethod, metadataURI string) error {
	k.requireStore()

	doc, found := k.store.getDIDDocument(ctx, did)
	if !found {
		return types.ErrDIDNotFound
	}

	doc.VerificationMethods = verificationMethods
	doc.MetadataUri = metadataURI
	doc.Updated = &gogotypes.Timestamp{Seconds: k.getCurrentTime(ctx), Nanos: 0}
	k.store.setDIDDocument(ctx, doc)
	return nil
}

// GetDIDsByAddress retrieves all DIDs for a controller address
func (k *Keeper) GetDIDsByAddress(ctx context.Context, controller string) []string {
	k.requireStore()

	dids := k.store.listAddressDIDs(ctx, controller)
	if dids == nil {
		return []string{}
	}
	return dids
}

// AddCredentialToDID adds a credential ID to a DID document
func (k *Keeper) AddCredentialToDID(ctx context.Context, did string, vcID string) error {
	k.requireStore()

	doc, found := k.store.getDIDDocument(ctx, did)
	if !found {
		return types.ErrDIDNotFound
	}

	doc.CredentialIds = append(doc.CredentialIds, vcID)
	doc.Updated = &gogotypes.Timestamp{Seconds: k.getCurrentTime(ctx), Nanos: 0}
	k.store.setDIDDocument(ctx, doc)
	return nil
}

// RemoveCredentialFromDID removes a credential ID from a DID document
func (k *Keeper) RemoveCredentialFromDID(ctx context.Context, did string, vcID string) error {
	k.requireStore()

	doc, found := k.store.getDIDDocument(ctx, did)
	if !found {
		return types.ErrDIDNotFound
	}

	filtered := make([]string, 0, len(doc.CredentialIds))
	for _, id := range doc.CredentialIds {
		if id != vcID {
			filtered = append(filtered, id)
		}
	}
	doc.CredentialIds = filtered
	doc.Updated = &gogotypes.Timestamp{Seconds: k.getCurrentTime(ctx), Nanos: 0}
	k.store.setDIDDocument(ctx, doc)
	return nil
}

// ============================
// ATTRIBUTE VC MANAGEMENT
// ============================

// CreateAttributeVC issues a new attribute VC.
func (k *Keeper) CreateAttributeVC(ctx context.Context, avc types.AttributeVC) error {
	if avc.AttributeType == types.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED {
		return fmt.Errorf("attribute_type required")
	}
	if avc.AttributeVcId == "" {
		return fmt.Errorf("attribute_vc_id required")
	}
	if avc.HolderAddress == "" {
		return types.ErrInvalidHolderAddress
	}
	// Prevent duplicate IDs
	if _, ok := k.GetAttributeVC(ctx, avc.AttributeVcId); ok {
		return fmt.Errorf("attribute VC already exists: %s", avc.AttributeVcId)
	}
	if len(avc.EncryptedValue) == 0 && len(avc.ValueHash) == 0 {
		return fmt.Errorf("encrypted_value or value_hash required")
	}
	if avc.Status == types.VCStatus_VC_STATUS_UNSPECIFIED {
		avc.Status = types.VCStatus_VC_STATUS_ACTIVE
	}

	// Enforce expiry if set
	if avc.ExpiresAt != nil && avc.ExpiresAt.Seconds <= k.getCurrentTime(ctx) {
		return fmt.Errorf("attribute VC already expired")
	}

	// Disallow multiple active VCs of the same attribute type for a holder
	existing := k.ListAttributeVCs(ctx, avc.HolderAddress, nil)
	for _, e := range existing {
		if e.AttributeType == avc.AttributeType && e.Status == types.VCStatus_VC_STATUS_ACTIVE {
			return fmt.Errorf("attribute VC of type %s already active for holder", e.AttributeType.String())
		}
	}

	// Set issued_at if missing
	if avc.IssuedAt == nil {
		avc.IssuedAt = &gogotypes.Timestamp{Seconds: k.getCurrentTime(ctx), Nanos: 0}
	}

	k.requireStore()

	k.store.setAttributeVC(ctx, avc)
	k.store.appendUserAttributeVC(ctx, avc.HolderAddress, avc.AttributeVcId)

	return nil
}

// GetAttributeVC retrieves an attribute VC by ID.
func (k *Keeper) GetAttributeVC(ctx context.Context, avcID string) (types.AttributeVC, bool) {
	k.requireStore()
	return k.store.getAttributeVC(ctx, avcID)
}

// ListAttributeVCs returns attribute VCs for a holder, optionally filtered by types.
func (k *Keeper) ListAttributeVCs(ctx context.Context, holder string, filter []types.AttributeType) []types.AttributeVC {
	k.requireStore()

	ids := k.store.listUserAttributeVCs(ctx, holder)
	if len(ids) == 0 {
		return nil
	}

	// filter set
	filterSet := make(map[types.AttributeType]bool)
	for _, t := range filter {
		filterSet[t] = true
	}

	res := make([]types.AttributeVC, 0, len(ids))
	for _, id := range ids {
		avc, ok := k.store.getAttributeVC(ctx, id)
		if !ok {
			continue
		}
		if len(filterSet) > 0 && !filterSet[avc.AttributeType] {
			continue
		}
		res = append(res, avc)
	}

	return res
}

// RevokeAttributeVC revokes an attribute VC.
func (k *Keeper) RevokeAttributeVC(ctx context.Context, avcID string, reason string) error {
	avc, ok := k.GetAttributeVC(ctx, avcID)
	if !ok {
		return fmt.Errorf("attribute VC not found")
	}
	if avc.Status == types.VCStatus_VC_STATUS_REVOKED {
		return fmt.Errorf("attribute VC already revoked")
	}

	avc.Status = types.VCStatus_VC_STATUS_REVOKED

	k.requireStore()

	k.store.setAttributeVC(ctx, avc)
	return nil
}

// ============================
// DISCLOSURE MANAGEMENT
// ============================

// SetDisclosurePolicy creates or updates a holder's disclosure policy.
func (k *Keeper) SetDisclosurePolicy(ctx context.Context, policy types.DisclosurePolicy) error {
	if policy.HolderAddress == "" {
		return types.ErrInvalidHolderAddress
	}
	if policy.DefaultMode == 0 {
		policy.DefaultMode = types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_DENY
	}
	if policy.DefaultMode > types.DisclosurePolicyMode_DISCLOSURE_POLICY_MODE_CONDITIONAL {
		return fmt.Errorf("invalid default_mode")
	}

	// Validate rules: unique attribute types and required fields
	seen := make(map[types.AttributeType]struct{})
	for i, rule := range policy.Rules {
		if rule.AttributeType == types.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED {
			return fmt.Errorf("rule %d attribute_type required", i)
		}
		if _, dup := seen[rule.AttributeType]; dup {
			return fmt.Errorf("duplicate rule for attribute_type %s", rule.AttributeType.String())
		}
		seen[rule.AttributeType] = struct{}{}
	}
	policy.UpdatedAt = &gogotypes.Timestamp{Seconds: k.getCurrentTime(ctx), Nanos: 0}

	k.requireStore()

	k.store.setDisclosurePolicy(ctx, policy)
	return nil
}

// GetDisclosurePolicy retrieves a holder's disclosure policy.
func (k *Keeper) GetDisclosurePolicy(ctx context.Context, holder string) (types.DisclosurePolicy, bool) {
	k.requireStore()
	return k.store.getDisclosurePolicy(ctx, holder)
}

// GetDisclosureResponse retrieves a disclosure response by request ID.
func (k *Keeper) GetDisclosureResponse(ctx context.Context, requestID string) (types.DisclosureResponse, bool) {
	k.requireStore()
	return k.store.getDisclosureResponse(ctx, requestID)
}

// CreateDisclosureRequest stores a new disclosure request and indexes it as pending.
func (k *Keeper) CreateDisclosureRequest(ctx context.Context, holderAddress string, req types.DisclosureRequest) error {
	if req.RequestId == "" {
		return fmt.Errorf("request_id required")
	}
	if holderAddress == "" {
		return types.ErrInvalidHolderAddress
	}
	if req.VerifierAddress == "" {
		return fmt.Errorf("verifier_address required")
	}
	if len(req.RequestedAttributes) == 0 {
		return fmt.Errorf("requested_attributes required")
	}
	if req.ExpiresInSeconds == 0 {
		req.ExpiresInSeconds = 300 // default 5m
	}
	if req.ExpiresInSeconds > 86400 {
		return fmt.Errorf("expires_in_seconds too long")
	}
	if req.RequestedAt == nil {
		req.RequestedAt = &gogotypes.Timestamp{Seconds: k.getCurrentTime(ctx), Nanos: 0}
	}
	if req.RequestedAt.Seconds+int64(req.ExpiresInSeconds) <= k.getCurrentTime(ctx) {
		return fmt.Errorf("disclosure request already expired")
	}

	k.requireStore()

	k.store.setDisclosureRequest(ctx, req)
	k.store.appendPendingDisclosure(ctx, holderAddress, req.RequestId)
	return nil
}

// GetDisclosureRequest retrieves a disclosure request.
func (k *Keeper) GetDisclosureRequest(ctx context.Context, requestID string) (types.DisclosureRequest, bool) {
	k.requireStore()
	return k.store.getDisclosureRequest(ctx, requestID)
}

// RespondToDisclosureRequest stores a disclosure response and removes from pending.
func (k *Keeper) RespondToDisclosureRequest(ctx context.Context, resp types.DisclosureResponse) error {
	if resp.RequestId == "" {
		return fmt.Errorf("request_id required")
	}
	if resp.HolderAddress == "" {
		return types.ErrInvalidHolderAddress
	}

	req, ok := k.GetDisclosureRequest(ctx, resp.RequestId)
	if !ok {
		return fmt.Errorf("disclosure request not found")
	}
	if req.RequestedAt != nil && req.RequestedAt.Seconds+int64(req.ExpiresInSeconds) <= k.getCurrentTime(ctx) {
		return fmt.Errorf("disclosure request expired")
	}

	// If approved, ensure disclosed attributes are a subset of requested
	if resp.Approved {
		allowed := make(map[types.AttributeType]struct{}, len(req.RequestedAttributes))
		for _, a := range req.RequestedAttributes {
			allowed[a] = struct{}{}
		}
		for _, da := range resp.DisclosedAttributes {
			if da == nil {
				return fmt.Errorf("disclosed attribute entry missing")
			}
			if _, ok := allowed[da.AttributeType]; !ok {
				return fmt.Errorf("disclosed attribute %s was not requested", da.AttributeType.String())
			}
		}
	}

	k.requireStore()

	// Ensure the request is currently pending for this holder
	pendingForHolder := false
	for _, id := range k.store.listPendingDisclosures(ctx, resp.HolderAddress) {
		if id == resp.RequestId {
			pendingForHolder = true
			break
		}
	}
	if !pendingForHolder {
		return fmt.Errorf("request not pending for holder")
	}

	// Prevent duplicate responses
	if _, exists := k.store.getDisclosureResponse(ctx, resp.RequestId); exists {
		return fmt.Errorf("disclosure response already recorded")
	}

	resp.RespondedAt = &gogotypes.Timestamp{Seconds: k.getCurrentTime(ctx), Nanos: 0}

	k.store.setDisclosureResponse(ctx, resp)
	// remove pending marker if present
	reqs := k.store.listPendingDisclosures(ctx, resp.HolderAddress)
	for _, r := range reqs {
		if r == resp.RequestId {
			k.store.deletePendingDisclosure(ctx, resp.HolderAddress, resp.RequestId)
		}
	}

	return nil
}

// ============================
// VC POLICY MANAGEMENT
// ============================

// GetVCPolicy retrieves a VC policy by type name
func (k *Keeper) GetVCPolicy(ctx context.Context, vcTypeName string) (types.VCPolicy, bool) {
	k.requireStore()
	return k.store.getVCPolicy(ctx, vcTypeName)
}

// SetVCPolicy stores a VC policy
func (k *Keeper) SetVCPolicy(ctx context.Context, policy types.VCPolicy) error {
	if policy.VcTypeName == "" {
		return types.ErrInvalidVCType
	}

	k.requireStore()

	k.store.setVCPolicy(ctx, policy)
	return nil
}

// ListVcPolicies returns all VC policies
func (k *Keeper) ListVcPolicies(ctx context.Context, statusFilter types.VCPolicyStatus) []types.VCPolicy {
	k.requireStore()

	policies := []types.VCPolicy{}
	for _, policy := range k.store.iterateVCPolicies(ctx) {
		if statusFilter != types.VCPolicyStatusUnspecified && policy.Status != statusFilter {
			continue
		}
		policies = append(policies, policy)
	}
	return policies
}

// ============================
// RATE LIMITING
// ============================

// CheckMintRateLimit checks if user has exceeded minting rate limits
func (k *Keeper) CheckMintRateLimit(ctx context.Context, holderAddress string) error {
	params, _ := k.GetParams(ctx)
	if !params.RateLimitingEnabled {
		return nil
	}

	k.requireStore()
	dayTimestamp := k.getCurrentTime(ctx) / 86400


	sdkCtx := k.sdkContext(ctx)
	count, found := k.store.getMintCount(sdkCtx, holderAddress, dayTimestamp)
	if !found {
		return nil
	}
	if count >= params.MaxMintPerDay {
		return types.ErrRateLimitExceeded
	}
	return nil
}

// IncrementMintCount increments the mint count for rate limiting
func (k *Keeper) IncrementMintCount(ctx context.Context, holderAddress string) {
	k.requireStore()

	dayTimestamp := k.getCurrentTime(ctx) / 86400
	sdkCtx := k.sdkContext(ctx)
	count, _ := k.store.getMintCount(sdkCtx, holderAddress, dayTimestamp)
	k.store.setMintCount(sdkCtx, holderAddress, dayTimestamp, count+1)
}

// CleanupOldMintCounts removes old mint count entries (older than 7 days).
// This function is consensus-safe: it uses deterministic iteration order
// to ensure all validators produce the same state changes.
func (k *Keeper) CleanupOldMintCounts(ctx context.Context) {
	k.requireStore()

	cutoffDay := (k.getCurrentTime(ctx) / 86400) - 7
	sdkCtx := k.sdkContext(ctx)

	// Use the deterministic cleanup method that iterates the KVStore
	// in lexicographic order (deterministic across all validators)
	k.store.cleanupOldMintCountsDeterministic(sdkCtx, cutoffDay)
}

// ============================
// STATISTICS
// ============================

// GetStats returns registry statistics
func (k *Keeper) GetStats(ctx context.Context) types.RegistryStats {
	k.requireStore()

	stats := types.RegistryStats{VCsByType: make(map[types.VCType]uint64)}

	for _, vc := range k.store.iterateVCRecords(ctx) {
		switch vc.Status {
		case types.VCStatus_VC_STATUS_ACTIVE:
			stats.ActiveVCs++
		case types.VCStatus_VC_STATUS_REVOKED:
			stats.RevokedVCs++
		case types.VCStatus_VC_STATUS_EXPIRED:
			stats.ExpiredVCs++
		}
		stats.TotalVCs++
		stats.VCsByType[vc.VcType]++
	}
	stats.TotalDIDs = uint64(len(k.store.iterateDIDDocuments(ctx)))
	stats.TotalPolicies = uint64(len(k.store.iterateVCPolicies(ctx)))
	return stats
}

// ============================
// GENESIS
// ============================
// Note: InitGenesis and ExportGenesis are implemented in genesis.go

// GenerateVCID generates a unique VC ID using context for deterministic time/height
func (k *Keeper) GenerateVCID(ctx context.Context, holderAddress string, vcType types.VCType) string {
	h := sha256.New()
	h.Write([]byte(holderAddress))
	h.Write([]byte(fmt.Sprintf("%d", vcType)))
	h.Write([]byte(fmt.Sprintf("%d", k.getCurrentTime(ctx))))
	h.Write([]byte(fmt.Sprintf("%d", k.getCurrentHeight(ctx))))
	return "vc:" + hex.EncodeToString(h.Sum(nil))[:32]
}

// GenerateAttributeVCID generates a unique ID for attribute VCs using context for deterministic time/height
func (k *Keeper) GenerateAttributeVCID(ctx context.Context, holderAddress string, attrType types.AttributeType) string {
	h := sha256.New()
	h.Write([]byte(holderAddress))
	h.Write([]byte(attrType.String()))
	h.Write([]byte(fmt.Sprintf("%d", k.getCurrentTime(ctx))))
	h.Write([]byte(fmt.Sprintf("%d", k.getCurrentHeight(ctx))))
	return "avc:" + hex.EncodeToString(h.Sum(nil))[:32]
}
