package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/aequitas/aura/chain/x/vcregistry/params"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
	"google.golang.org/protobuf/types/known/timestamppb"
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
type Keeper struct {
	mu                sync.RWMutex
	vcRecords         map[string]types.VCRecord         // vc_id -> VCRecord
	userVCs           map[string][]string               // holder_address -> []vc_id
	revocationRecords map[string]types.RevocationRecord // vc_id -> RevocationRecord
	revocationList    types.RevocationList
	didDocuments      map[string]types.DIDDocument // did -> DIDDocument
	addressToDIDs     map[string][]string          // controller -> []did
	vcPolicies        map[string]types.VCPolicy    // vc_type_name -> VCPolicy
	userMintCounts    map[string]map[int64]uint64  // address -> day_timestamp -> count
	paramsStore       *params.Store
	csKeeper          ConfidenceScoreKeeper
	currentHeight     uint64
	currentTime       int64
	authority         string // governance module account address

	// Selective disclosure fields
	attributeVCs        map[string]interface{} // attribute_vc_id -> AttributeVC (interface{} for now)
	userAttributeVCs    map[string][]string    // holder_address -> []attribute_vc_id
	disclosurePolicies  map[string]interface{} // holder_address -> DisclosurePolicy
	disclosureRequests  map[string]interface{} // request_id -> DisclosureRequest
	disclosureResponses map[string]interface{} // request_id -> DisclosureResponse
	presentations       map[string]interface{} // presentation_id -> VCPresentation
	userPresentations   map[string][]string    // holder_address -> []presentation_id
	pendingDisclosures  map[string][]string    // holder_address -> []request_id
}

// NewKeeper creates a new Keeper instance
func NewKeeper(store *params.Store, authority string) *Keeper {
	if store == nil {
		store = params.NewStore(types.DefaultParams())
	}
	return &Keeper{
		vcRecords:         make(map[string]types.VCRecord),
		userVCs:           make(map[string][]string),
		revocationRecords: make(map[string]types.RevocationRecord),
		revocationList: types.RevocationList{
			MerkleRoot:        []byte{},
			TotalRevocations:  0,
			LastUpdatedHeight: 0,
			LastUpdated:       nil,
		},
		didDocuments:   make(map[string]types.DIDDocument),
		addressToDIDs:  make(map[string][]string),
		vcPolicies:     make(map[string]types.VCPolicy),
		userMintCounts: make(map[string]map[int64]uint64),
		paramsStore:    store,
		currentTime:    time.Now().Unix(),
		authority:      authority,

		// Initialize selective disclosure maps
		attributeVCs:        make(map[string]interface{}),
		userAttributeVCs:    make(map[string][]string),
		disclosurePolicies:  make(map[string]interface{}),
		disclosureRequests:  make(map[string]interface{}),
		disclosureResponses: make(map[string]interface{}),
		presentations:       make(map[string]interface{}),
		userPresentations:   make(map[string][]string),
		pendingDisclosures:  make(map[string][]string),
	}
}

// GetAuthority returns the governance module account address
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// SetConfidenceScoreKeeper sets the confidence score keeper for validation
func (k *Keeper) SetConfidenceScoreKeeper(keeper ConfidenceScoreKeeper) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.csKeeper = keeper
}

// SetCurrentHeight sets the current block height
func (k *Keeper) SetCurrentHeight(height uint64) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.currentHeight = height
}

// SetCurrentTime sets the current time
func (k *Keeper) SetCurrentTime(t int64) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.currentTime = t
}

// GetParams returns the current module parameters
func (k *Keeper) GetParams() types.Params {
	if k.paramsStore != nil {
		return k.paramsStore.GetParams()
	}
	return types.DefaultParams()
}

// SetParams sets new module parameters
func (k *Keeper) SetParams(params types.Params) error {
	if k.paramsStore == nil {
		return types.ErrUnauthorized
	}
	return k.paramsStore.SetParams(params)
}

// ============================
// VC RECORD MANAGEMENT
// ============================

// GetVCRecord retrieves a VC record by ID
func (k *Keeper) GetVCRecord(vcID string) (types.VCRecord, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	record, ok := k.vcRecords[vcID]
	return record, ok
}

// SetVCRecord stores a VC record
func (k *Keeper) SetVCRecord(record types.VCRecord) error {
	if record.VcId == "" {
		return types.ErrInvalidVCID
	}
	if record.HolderAddress == "" {
		return types.ErrInvalidHolderAddress
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	k.vcRecords[record.VcId] = record

	// Index by user
	k.userVCs[record.HolderAddress] = append(k.userVCs[record.HolderAddress], record.VcId)

	return nil
}

// ListUserVCs returns all VCs for a user
func (k *Keeper) ListUserVCs(holderAddress string, statusFilter types.VCStatus, typeFilter types.VCType) []types.VCRecord {
	k.mu.RLock()
	defer k.mu.RUnlock()

	vcIDs, ok := k.userVCs[holderAddress]
	if !ok {
		return []types.VCRecord{}
	}

	vcs := []types.VCRecord{}
	for _, vcID := range vcIDs {
		if vc, ok := k.vcRecords[vcID]; ok {
			// Apply filters
			if statusFilter != types.VCStatusUnspecified && vc.Status != statusFilter {
				continue
			}
			if typeFilter != types.VCTypeUnspecified && vc.VcType != typeFilter {
				continue
			}
			vcs = append(vcs, vc)
		}
	}

	return vcs
}

// ============================
// VC STATUS CHECKS
// ============================

// CheckVCStatus checks the current status of a VC
func (k *Keeper) CheckVCStatus(vcID string) (types.VCStatus, bool, error) {
	vc, ok := k.GetVCRecord(vcID)
	if !ok {
		return types.VCStatusUnspecified, false, types.ErrVCNotFound
	}

	// Check expiration
	if vc.ExpiresAt != nil && vc.ExpiresAt.Seconds <= k.currentTime {
		// Auto-update status to expired
		vc.Status = types.VCStatusExpired
		k.SetVCRecord(vc)
		return types.VCStatusExpired, false, nil
	}

	valid := vc.Status == types.VCStatusActive
	return vc.Status, valid, nil
}

// IsVCValid checks if a VC is currently valid (active and not expired)
func (k *Keeper) IsVCValid(vcID string) bool {
	status, valid, err := k.CheckVCStatus(vcID)
	return err == nil && valid && status == types.VCStatusActive
}

// ============================
// REVOCATION MANAGEMENT
// ============================

// RevokeVC revokes a verifiable credential
func (k *Keeper) RevokeVC(vcID string, reason types.RevocationReason, revoker string, evidence string) error {
	vc, ok := k.GetVCRecord(vcID)
	if !ok {
		return types.ErrVCNotFound
	}

	if vc.Status == types.VCStatusRevoked {
		return types.ErrVCAlreadyRevoked
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Create revocation record
	revRecord := types.RevocationRecord{
		VcId:          vcID,
		RevokedAt:     timestamppb.New(time.Unix(k.currentTime, 0)),
		RevokedHeight: k.currentHeight,
		Reason:        reason,
		Revoker:       revoker,
		Evidence:      evidence,
	}

	// Update VC status
	vc.Status = types.VCStatusRevoked
	k.vcRecords[vcID] = vc

	// Store revocation
	k.revocationRecords[vcID] = revRecord

	// Update Merkle tree
	k.updateRevocationMerkleRoot(vcID, revRecord)

	return nil
}

// GetRevocationRecord retrieves a revocation record
func (k *Keeper) GetRevocationRecord(vcID string) (types.RevocationRecord, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	record, ok := k.revocationRecords[vcID]
	return record, ok
}

// IsRevoked checks if a VC is revoked
func (k *Keeper) IsRevoked(vcID string) bool {
	_, ok := k.GetRevocationRecord(vcID)
	return ok
}

// updateRevocationMerkleRoot updates the Merkle root after a revocation
func (k *Keeper) updateRevocationMerkleRoot(vcID string, record types.RevocationRecord) {
	// Simple hash-based Merkle root (in production, use proper Merkle tree)
	h := sha256.New()
	h.Write([]byte(vcID))
	if record.RevokedAt != nil {
		h.Write([]byte(fmt.Sprintf("%d", record.RevokedAt.Seconds)))
	}
	h.Write([]byte(fmt.Sprintf("%d", record.Reason)))

	newHash := h.Sum(nil)

	// XOR with current root for simplicity (proper Merkle tree in production)
	if len(k.revocationList.MerkleRoot) == 0 {
		k.revocationList.MerkleRoot = newHash
	} else {
		combined := make([]byte, 32)
		for i := 0; i < 32; i++ {
			combined[i] = k.revocationList.MerkleRoot[i] ^ newHash[i]
		}
		k.revocationList.MerkleRoot = combined
	}

	k.revocationList.TotalRevocations++
	k.revocationList.LastUpdatedHeight = k.currentHeight
	k.revocationList.LastUpdated = timestamppb.New(time.Unix(k.currentTime, 0))
}

// GetRevocationList returns the current revocation list
func (k *Keeper) GetRevocationList() *types.RevocationList {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return &k.revocationList
}

// ============================
// DID MANAGEMENT
// ============================

// RegisterDID registers a new DID document
func (k *Keeper) RegisterDID(did string, controller string, verificationMethods []*types.VerificationMethod, metadataURI string) error {
	if did == "" {
		return types.ErrInvalidDID
	}
	if controller == "" {
		return types.ErrInvalidHolderAddress
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if DID already exists
	if _, ok := k.didDocuments[did]; ok {
		return types.ErrDIDAlreadyExists
	}

	doc := types.DIDDocument{
		Did:                 did,
		Controller:          controller,
		VerificationMethods: verificationMethods,
		CredentialIds:       []string{},
		Created:             timestamppb.New(time.Unix(k.currentTime, 0)),
		Updated:             timestamppb.New(time.Unix(k.currentTime, 0)),
		MetadataUri:         metadataURI,
		ServiceEndpoints:    make(map[string]string),
	}

	k.didDocuments[did] = doc
	k.addressToDIDs[controller] = append(k.addressToDIDs[controller], did)

	return nil
}

// GetDIDDocument retrieves a DID document
func (k *Keeper) GetDIDDocument(did string) (types.DIDDocument, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	doc, ok := k.didDocuments[did]
	return doc, ok
}

// UpdateDIDDocument updates a DID document
func (k *Keeper) UpdateDIDDocument(did string, verificationMethods []*types.VerificationMethod, metadataURI string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	doc, ok := k.didDocuments[did]
	if !ok {
		return types.ErrDIDNotFound
	}

	doc.VerificationMethods = verificationMethods
	doc.MetadataUri = metadataURI
	doc.Updated = timestamppb.New(time.Unix(k.currentTime, 0))

	k.didDocuments[did] = doc
	return nil
}

// GetDIDsByAddress retrieves all DIDs for a controller address
func (k *Keeper) GetDIDsByAddress(controller string) []string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.addressToDIDs[controller]
}

// AddCredentialToDID adds a credential ID to a DID document
func (k *Keeper) AddCredentialToDID(did string, vcID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	doc, ok := k.didDocuments[did]
	if !ok {
		return types.ErrDIDNotFound
	}

	doc.CredentialIds = append(doc.CredentialIds, vcID)
	doc.Updated = timestamppb.New(time.Unix(k.currentTime, 0))
	k.didDocuments[did] = doc

	return nil
}

// RemoveCredentialFromDID removes a credential ID from a DID document
func (k *Keeper) RemoveCredentialFromDID(did string, vcID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	doc, ok := k.didDocuments[did]
	if !ok {
		return types.ErrDIDNotFound
	}

	newCredIDs := []string{}
	for _, id := range doc.CredentialIds {
		if id != vcID {
			newCredIDs = append(newCredIDs, id)
		}
	}

	doc.CredentialIds = newCredIDs
	doc.Updated = timestamppb.New(time.Unix(k.currentTime, 0))
	k.didDocuments[did] = doc

	return nil
}

// ============================
// VC POLICY MANAGEMENT
// ============================

// GetVCPolicy retrieves a VC policy by type name
func (k *Keeper) GetVCPolicy(vcTypeName string) (types.VCPolicy, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	policy, ok := k.vcPolicies[vcTypeName]
	return policy, ok
}

// SetVCPolicy stores a VC policy
func (k *Keeper) SetVCPolicy(policy types.VCPolicy) error {
	if policy.VcTypeName == "" {
		return types.ErrInvalidVCType
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	k.vcPolicies[policy.VcTypeName] = policy
	return nil
}

// ListVCPolicies returns all VC policies
func (k *Keeper) ListVCPolicies(statusFilter types.VCPolicyStatus) []types.VCPolicy {
	k.mu.RLock()
	defer k.mu.RUnlock()

	policies := []types.VCPolicy{}
	for _, policy := range k.vcPolicies {
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
func (k *Keeper) CheckMintRateLimit(holderAddress string) error {
	params := k.GetParams()
	if !params.RateLimitingEnabled {
		return nil
	}

	k.mu.RLock()
	defer k.mu.RUnlock()

	// Get current day timestamp
	dayTimestamp := k.currentTime / 86400

	counts, ok := k.userMintCounts[holderAddress]
	if !ok {
		return nil // No mints yet
	}

	todayCount := counts[dayTimestamp]
	if todayCount >= params.MaxMintPerDay {
		return types.ErrRateLimitExceeded
	}

	return nil
}

// IncrementMintCount increments the mint count for rate limiting
func (k *Keeper) IncrementMintCount(holderAddress string) {
	k.mu.Lock()
	defer k.mu.Unlock()

	dayTimestamp := k.currentTime / 86400

	if k.userMintCounts[holderAddress] == nil {
		k.userMintCounts[holderAddress] = make(map[int64]uint64)
	}

	k.userMintCounts[holderAddress][dayTimestamp]++
}

// CleanupOldMintCounts removes old mint count entries (older than 7 days)
func (k *Keeper) CleanupOldMintCounts() {
	k.mu.Lock()
	defer k.mu.Unlock()

	cutoff := (k.currentTime / 86400) - 7

	for addr, counts := range k.userMintCounts {
		for day := range counts {
			if day < cutoff {
				delete(k.userMintCounts[addr], day)
			}
		}
	}
}

// ============================
// STATISTICS
// ============================

// GetStats returns registry statistics
func (k *Keeper) GetStats() types.RegistryStats {
	k.mu.RLock()
	defer k.mu.RUnlock()

	stats := types.RegistryStats{
		TotalVCs:      uint64(len(k.vcRecords)),
		TotalDIDs:     uint64(len(k.didDocuments)),
		TotalPolicies: uint64(len(k.vcPolicies)),
		VCsByType:     make(map[string]uint64),
	}

	for _, vc := range k.vcRecords {
		switch vc.Status {
		case types.VCStatusActive:
			stats.ActiveVCs++
		case types.VCStatusRevoked:
			stats.RevokedVCs++
		case types.VCStatusExpired:
			stats.ExpiredVCs++
		}

		vcTypeName := fmt.Sprintf("%d", vc.VcType)
		if vc.VcTypeCustom != "" {
			vcTypeName = vc.VcTypeCustom
		}
		stats.VCsByType[vcTypeName]++
	}

	return stats
}

// ============================
// GENESIS
// ============================

// InitGenesis initializes the keeper from genesis state
func (k *Keeper) InitGenesis(genesis types.GenesisState) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Set params
	if err := k.paramsStore.SetParams(genesis.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}

	// Load VC records
	for _, record := range genesis.VCRecords {
		k.vcRecords[record.VcId] = record
		k.userVCs[record.HolderAddress] = append(k.userVCs[record.HolderAddress], record.VcId)
	}

	// Load revocations
	for _, revRecord := range genesis.RevocationRecords {
		k.revocationRecords[revRecord.VcId] = revRecord
	}

	// Load revocation list
	if genesis.RevocationList != nil {
		k.revocationList = *genesis.RevocationList
	}

	// Load DID documents
	for _, doc := range genesis.DIDDocuments {
		k.didDocuments[doc.Did] = doc
		k.addressToDIDs[doc.Controller] = append(k.addressToDIDs[doc.Controller], doc.Did)
	}

	// Load policies
	for _, policy := range genesis.VCPolicies {
		k.vcPolicies[policy.VcTypeName] = policy
	}

	// Load mint counts
	k.userMintCounts = genesis.UserMintCounts

	return nil
}

// ExportGenesis exports the current state for genesis
func (k *Keeper) ExportGenesis() types.GenesisState {
	k.mu.RLock()
	defer k.mu.RUnlock()

	vcRecords := []types.VCRecord{}
	for _, record := range k.vcRecords {
		vcRecords = append(vcRecords, record)
	}

	revRecords := []types.RevocationRecord{}
	for _, record := range k.revocationRecords {
		revRecords = append(revRecords, record)
	}

	didDocs := []types.DIDDocument{}
	for _, doc := range k.didDocuments {
		didDocs = append(didDocs, doc)
	}

	policies := []types.VCPolicy{}
	for _, policy := range k.vcPolicies {
		policies = append(policies, policy)
	}

	return types.GenesisState{
		Params:            k.GetParams(),
		VCRecords:         vcRecords,
		RevocationRecords: revRecords,
		RevocationList:    &k.revocationList,
		DIDDocuments:      didDocs,
		VCPolicies:        policies,
		UserMintCounts:    k.userMintCounts,
	}
}

// GenerateVCID generates a unique VC ID
func (k *Keeper) GenerateVCID(holderAddress string, vcType types.VCType) string {
	h := sha256.New()
	h.Write([]byte(holderAddress))
	h.Write([]byte(fmt.Sprintf("%d", vcType)))
	h.Write([]byte(fmt.Sprintf("%d", k.currentTime)))
	h.Write([]byte(fmt.Sprintf("%d", k.currentHeight)))
	return "vc:" + hex.EncodeToString(h.Sum(nil))[:32]
}
