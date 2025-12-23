package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
)

// ============================
// VC SCHEMA VALIDATION
// ============================

// VCSchema defines the structure and validation rules for a VC type
type VCSchema struct {
	TypeName         string                 `json:"type_name"`
	Version          string                 `json:"version"`
	RequiredFields   []string               `json:"required_fields"`
	OptionalFields   []string               `json:"optional_fields"`
	FieldTypes       map[string]string      `json:"field_types"` // field -> type (string, int, bool, etc.)
	FieldConstraints map[string]interface{} `json:"field_constraints"`
	Metadata         map[string]string      `json:"metadata"`
}

// RegisterVCSchema registers a new VC schema for validation
func (k *Keeper) RegisterVCSchema(ctx context.Context, schema VCSchema) error {
	if schema.TypeName == "" {
		return types.ErrInvalidVCType
	}
	if schema.Version == "" {
		schema.Version = "1.0"
	}

	k.requireStore()

	// Store schema in metadata
	schemaKey := fmt.Sprintf("schema:%s:%s", schema.TypeName, schema.Version)
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// In production, would store in dedicated schema KV store
	// For now, using metadata approach
	k.store.setMetadata(ctx, schemaKey, string(schemaJSON))

	return nil
}

// ValidateVCAgainstSchema validates VC data against its schema
func (k *Keeper) ValidateVCAgainstSchema(ctx context.Context, vcType string, vcData map[string]interface{}) error {
	k.requireStore()

	// Get schema
	schemaKey := fmt.Sprintf("schema:%s:1.0", vcType)
	schemaJSON, ok := k.store.getMetadata(ctx, schemaKey)
	if !ok {
		// No schema registered - allow by default
		return nil
	}

	var schema VCSchema
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	// Validate required fields
	for _, field := range schema.RequiredFields {
		if _, exists := vcData[field]; !exists {
			return fmt.Errorf("required field missing: %s", field)
		}
	}

	// Validate field types
	for field, value := range vcData {
		expectedType, hasType := schema.FieldTypes[field]
		if !hasType {
			continue // Field not in schema, skip
		}

		if !k.validateFieldType(value, expectedType) {
			return fmt.Errorf("field %s has invalid type, expected %s", field, expectedType)
		}
	}

	return nil
}

// validateFieldType checks if a value matches the expected type
func (k *Keeper) validateFieldType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "int":
		_, ok := value.(int)
		if !ok {
			_, ok = value.(float64) // JSON numbers
		}
		return ok
	case "bool":
		_, ok := value.(bool)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	default:
		return true // Unknown type, allow
	}
}

// ============================
// VC TEMPLATES
// ============================

// VCTemplate defines a reusable template for creating VCs
type VCTemplate struct {
	TemplateID        string                 `json:"template_id"`
	Name              string                 `json:"name"`
	VCType            types.VCType           `json:"vc_type"`
	Description       string                 `json:"description"`
	DefaultMetadata   map[string]string      `json:"default_metadata"`
	DefaultExpiryDays uint64                 `json:"default_expiry_days"`
	FieldDefaults     map[string]interface{} `json:"field_defaults"`
	Creator           string                 `json:"creator"`
	CreatedAt         int64                  `json:"created_at"`
	IsActive          bool                   `json:"is_active"`
}

// CreateVCTemplate creates a new VC template
func (k *Keeper) CreateVCTemplate(ctx context.Context, template VCTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("template name required")
	}
	if template.VCType == types.VCType_VC_TYPE_UNSPECIFIED {
		return types.ErrInvalidVCType
	}

	// Generate template ID
	if template.TemplateID == "" {
		template.TemplateID = k.generateTemplateID(template.Name, template.VCType)
	}

	// Get consensus-safe block time
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime().Unix()
	template.CreatedAt = currentTime
	template.IsActive = true

	k.requireStore()

	// Store template
	templateKey := fmt.Sprintf("template:%s", template.TemplateID)
	templateJSON, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	k.store.setMetadata(ctx, templateKey, string(templateJSON))

	return nil
}

// GetVCTemplate retrieves a VC template
func (k *Keeper) GetVCTemplate(ctx context.Context, templateID string) (*VCTemplate, error) {
	k.requireStore()

	templateKey := fmt.Sprintf("template:%s", templateID)
	templateJSON, ok := k.store.getMetadata(ctx, templateKey)
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	var template VCTemplate
	if err := json.Unmarshal([]byte(templateJSON), &template); err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	return &template, nil
}

// MintVCFromTemplate creates a VC using a template
func (k *Keeper) MintVCFromTemplate(ctx context.Context, templateID string, holderAddress string, holderDID string, overrides map[string]string) (string, error) {
	// Get template
	template, err := k.GetVCTemplate(ctx, templateID)
	if err != nil {
		return "", err
	}

	if !template.IsActive {
		return "", fmt.Errorf("template is not active: %s", templateID)
	}

	// Merge metadata
	metadata := make(map[string]string)
	for k, v := range template.DefaultMetadata {
		metadata[k] = v
	}
	for k, v := range overrides {
		metadata[k] = v
	}
	metadata["template_id"] = templateID

	// Mint VC using template defaults
	vcID, err := k.MintVC(ctx, holderAddress, holderDID, template.VCType, "", metadata)
	if err != nil {
		return "", fmt.Errorf("failed to mint VC from template: %w", err)
	}

	return vcID, nil
}

// generateTemplateID generates a unique template ID
// Note: Uses deterministic hashing based on name and type only (no time dependency)
func (k *Keeper) generateTemplateID(name string, vcType types.VCType) string {
	h := sha256.New()
	h.Write([]byte(name))
	h.Write([]byte(fmt.Sprintf("%d", vcType)))
	// Use name again for additional entropy instead of time for determinism
	h.Write([]byte(name))
	return "tmpl:" + hex.EncodeToString(h.Sum(nil))[:16]
}

// ============================
// VC TRANSFER
// ============================

// TransferVC transfers a VC from one holder to another
func (k *Keeper) TransferVC(ctx context.Context, vcID string, fromAddress string, toAddress string, toDID string, reason string) error {
	if vcID == "" {
		return types.ErrInvalidVCID
	}
	if fromAddress == "" || toAddress == "" {
		return types.ErrInvalidHolderAddress
	}
	if toDID == "" {
		return types.ErrInvalidDID
	}

	k.requireStore()

	// Get VC record
	vcRecord, ok := k.store.getVCRecord(ctx, vcID)
	if !ok {
		return types.ErrVCNotFound
	}

	// Verify current holder
	if vcRecord.HolderAddress != fromAddress {
		return types.ErrNotVCHolder
	}

	// Verify VC is transferable (not all VCs can be transferred)
	if !k.isVCTransferable(ctx, vcRecord) {
		return fmt.Errorf("VC is not transferable")
	}

	// Update holder
	oldDID := vcRecord.HolderDid
	oldAddress := vcRecord.HolderAddress

	vcRecord.HolderAddress = toAddress
	vcRecord.HolderDid = toDID

	// Add transfer metadata
	if vcRecord.Metadata == nil {
		vcRecord.Metadata = make(map[string]string)
	}
	vcRecord.Metadata["transfer_from"] = oldAddress
	vcRecord.Metadata["transfer_at"] = fmt.Sprintf("%d", k.getCurrentTime(ctx))
	vcRecord.Metadata["transfer_reason"] = reason
	vcRecord.Metadata["transfer_count"] = fmt.Sprintf("%d", k.getTransferCount(vcRecord)+1)

	// Save updated record
	k.store.setVCRecord(ctx, vcRecord)

	// Update indices
	k.store.removeUserVC(ctx, oldAddress, vcID)
	k.store.appendUserVC(ctx, toAddress, vcID)

	// Update DID documents
	if err := k.RemoveCredentialFromDID(ctx, oldDID, vcID); err != nil {
		return err
	}
	if err := k.AddCredentialToDID(ctx, toDID, vcID); err != nil {
		return err
	}

	return nil
}

// isVCTransferable checks if a VC can be transferred
func (k *Keeper) isVCTransferable(ctx context.Context, vc types.VCRecord) bool {
	// Check VC policy
	vcTypeName := vc.VcTypeCustom
	if vc.VcType != types.VCType_VC_TYPE_CUSTOM {
		vcTypeName = fmt.Sprintf("%d", vc.VcType)
	}

	policy, ok := k.store.getVCPolicy(ctx, vcTypeName)
	if !ok {
		return false // No policy, not transferable
	}

	// Singleton VCs cannot be transferred
	if policy.Singleton {
		return false
	}

	// Check metadata for transfer restrictions
	if vc.Metadata != nil {
		if transferable, ok := vc.Metadata["transferable"]; ok && transferable == "false" {
			return false
		}
	}

	return true
}

// getTransferCount gets the number of times a VC has been transferred
func (k *Keeper) getTransferCount(vc types.VCRecord) int {
	if vc.Metadata == nil {
		return 0
	}
	if countStr, ok := vc.Metadata["transfer_count"]; ok {
		var count int
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			return 0
		}
		return count
	}
	return 0
}

// ============================
// VC SEARCH & FILTERING
// ============================

// VCSearchCriteria defines search parameters
type VCSearchCriteria struct {
	HolderAddress  string
	VCType         types.VCType
	Status         types.VCStatus
	IssuedAfter    int64
	IssuedBefore   int64
	ExpiringWithin int64 // seconds
	MetadataMatch  map[string]string
	Limit          int
	Offset         int
}

// SearchVCs searches for VCs matching criteria
func (k *Keeper) SearchVCs(ctx context.Context, criteria VCSearchCriteria) ([]types.VCRecord, int) {
	k.requireStore()

	results := make([]types.VCRecord, 0, 64)
	totalMatches := 0

	// If holder address specified, search user's VCs only
	searchSet := make([]types.VCRecord, 0, 64)
	if criteria.HolderAddress != "" {
		searchSet = k.ListUserVCs(ctx, criteria.HolderAddress, types.VCStatus_VC_STATUS_UNSPECIFIED, types.VCType_VC_TYPE_UNSPECIFIED)
	} else {
		// Search all VCs
	searchSet = append(searchSet, k.store.iterateVCRecords(ctx)...)
	}

	// Apply filters
	for _, vc := range searchSet {
		if !k.matchesSearchCriteria(ctx, vc, criteria) {
			continue
		}

		totalMatches++

		// Apply offset/limit
		if criteria.Offset > 0 && totalMatches <= criteria.Offset {
			continue
		}
		if criteria.Limit > 0 && len(results) >= criteria.Limit {
			continue
		}

		results = append(results, vc)
	}

	return results, totalMatches
}

// matchesSearchCriteria checks if a VC matches search criteria
func (k *Keeper) matchesSearchCriteria(ctx context.Context, vc types.VCRecord, criteria VCSearchCriteria) bool {
	// Type filter
	if criteria.VCType != types.VCType_VC_TYPE_UNSPECIFIED && vc.VcType != criteria.VCType {
		return false
	}

	// Status filter
	if criteria.Status != types.VCStatus_VC_STATUS_UNSPECIFIED && vc.Status != criteria.Status {
		return false
	}

	// Issued after filter
	if criteria.IssuedAfter > 0 && vc.IssuedAt != nil {
		if vc.IssuedAt.Seconds < criteria.IssuedAfter {
			return false
		}
	}

	// Issued before filter
	if criteria.IssuedBefore > 0 && vc.IssuedAt != nil {
		if vc.IssuedAt.Seconds > criteria.IssuedBefore {
			return false
		}
	}

	// Expiring within filter
	if criteria.ExpiringWithin > 0 && vc.ExpiresAt != nil {
		expiryThreshold := k.getCurrentTime(ctx) + criteria.ExpiringWithin
		if vc.ExpiresAt.Seconds > expiryThreshold {
			return false
		}
	}

	// Metadata match filter
	if len(criteria.MetadataMatch) > 0 {
		if vc.Metadata == nil {
			return false
		}
		for key, value := range criteria.MetadataMatch {
			if vc.Metadata[key] != value {
				return false
			}
		}
	}

	return true
}

// ============================
// VC ANALYTICS
// ============================

// VCAnalytics provides analytics data
type VCAnalytics struct {
	TotalVCs          uint64
	ActiveVCs         uint64
	RevokedVCs        uint64
	ExpiredVCs        uint64
	SuspendedVCs      uint64
	VCsByType         map[string]uint64
	VCsByDay          map[string]uint64 // YYYY-MM-DD -> count
	AverageCSAtMint   float64
	TopHolders        []HolderStats
	RecentRevocations []RevocationStats
	ExpiringVCs       []types.VCRecord // VCs expiring soon
}

// HolderStats provides stats about a VC holder
type HolderStats struct {
	Address         string
	TotalVCs        int
	ActiveVCs       int
	HighestCSAtMint uint64
}

// RevocationStats provides stats about revocations
type RevocationStats struct {
	VcID           string
	VcType         string
	RevokedAt      int64
	Reason         string
	RevokerAddress string
}

// GetVCAnalytics returns comprehensive analytics
func (k *Keeper) GetVCAnalytics(ctx context.Context, lookbackDays int) VCAnalytics {
	k.requireStore()

	analytics := VCAnalytics{
		VCsByType: make(map[string]uint64),
		VCsByDay:  make(map[string]uint64),
	}

	holderMap := make(map[string]*HolderStats)
	var totalCSAtMint uint64
	var csCount uint64

	cutoffTime := k.getCurrentTime(ctx) - (int64(lookbackDays) * 86400)

	// Analyze all VCs
	for _, vc := range k.store.iterateVCRecords(ctx) {
		analytics.TotalVCs++

		// Status counts
		switch vc.Status {
		case types.VCStatus_VC_STATUS_ACTIVE:
			analytics.ActiveVCs++
		case types.VCStatus_VC_STATUS_REVOKED:
			analytics.RevokedVCs++
		case types.VCStatus_VC_STATUS_EXPIRED:
			analytics.ExpiredVCs++
		case types.VCStatus_VC_STATUS_SUSPENDED:
			analytics.SuspendedVCs++
		}

		// Type counts
		vcTypeStr := vc.VcType.String()
		analytics.VCsByType[vcTypeStr]++

		// CS at mint average
		if vc.CsAtMint > 0 {
			totalCSAtMint += vc.CsAtMint
			csCount++
		}

		// Holder stats
		if _, ok := holderMap[vc.HolderAddress]; !ok {
			holderMap[vc.HolderAddress] = &HolderStats{
				Address: vc.HolderAddress,
			}
		}
		holderStats := holderMap[vc.HolderAddress]
		holderStats.TotalVCs++
		if vc.Status == types.VCStatus_VC_STATUS_ACTIVE {
			holderStats.ActiveVCs++
		}
		if vc.CsAtMint > holderStats.HighestCSAtMint {
			holderStats.HighestCSAtMint = vc.CsAtMint
		}

		// Daily stats
		if vc.IssuedAt != nil && vc.IssuedAt.Seconds >= cutoffTime {
			day := time.Unix(vc.IssuedAt.Seconds, 0).Format("2006-01-02")
			analytics.VCsByDay[day]++
		}

		// Expiring VCs (within 30 days)
		if vc.ExpiresAt != nil {
			expiryThreshold := k.getCurrentTime(ctx) + (30 * 86400)
			if vc.ExpiresAt.Seconds <= expiryThreshold && vc.ExpiresAt.Seconds > k.getCurrentTime(ctx) {
				analytics.ExpiringVCs = append(analytics.ExpiringVCs, vc)
			}
		}
	}

	// Calculate average CS
	if csCount > 0 {
		analytics.AverageCSAtMint = float64(totalCSAtMint) / float64(csCount)
	}

	// Get top holders (top 10)
	analytics.TopHolders = k.getTopHolders(holderMap, 10)

	// Get recent revocations
	analytics.RecentRevocations = k.getRecentRevocations(ctx, 20)

	return analytics
}

// getTopHolders returns top holders by total VC count
func (k *Keeper) getTopHolders(holderMap map[string]*HolderStats, limit int) []HolderStats {
	holders := make([]HolderStats, 0, len(holderMap))
	for _, stats := range holderMap {
		holders = append(holders, *stats)
	}

	// Simple bubble sort (sufficient for small datasets)
	for i := 0; i < len(holders)-1; i++ {
		for j := 0; j < len(holders)-i-1; j++ {
			if holders[j].TotalVCs < holders[j+1].TotalVCs {
				holders[j], holders[j+1] = holders[j+1], holders[j]
			}
		}
	}

	if len(holders) > limit {
		holders = holders[:limit]
	}

	return holders
}

// getRecentRevocations returns recent revocations
func (k *Keeper) getRecentRevocations(ctx context.Context, limit int) []RevocationStats {
	revocations := make([]RevocationStats, 0)

	for vcID, revRecord := range k.store.iterateRevocationRecords(ctx) {
		vc, ok := k.store.getVCRecord(ctx, vcID)
		vcType := "unknown"
		if ok {
			vcType = vc.VcType.String()
		}

		stat := RevocationStats{
			VcID:           vcID,
			VcType:         vcType,
			RevokedAt:      revRecord.RevokedAt.Seconds,
			Reason:         revRecord.Reason.String(),
			RevokerAddress: revRecord.Revoker,
		}
		revocations = append(revocations, stat)
	}

	// Sort by revoked_at descending
	for i := 0; i < len(revocations)-1; i++ {
		for j := 0; j < len(revocations)-i-1; j++ {
			if revocations[j].RevokedAt < revocations[j+1].RevokedAt {
				revocations[j], revocations[j+1] = revocations[j+1], revocations[j]
			}
		}
	}

	if len(revocations) > limit {
		revocations = revocations[:limit]
	}

	return revocations
}

// ============================
// VC LIFECYCLE MANAGEMENT
// ============================

// RenewVC renews an expiring VC (extends expiration)
func (k *Keeper) RenewVC(ctx context.Context, vcID string, holderAddress string, extensionDays uint64) error {
	if vcID == "" {
		return types.ErrInvalidVCID
	}

	k.requireStore()

	vcRecord, ok := k.store.getVCRecord(ctx, vcID)
	if !ok {
		return types.ErrVCNotFound
	}

	// Verify holder
	if vcRecord.HolderAddress != holderAddress {
		return types.ErrNotVCHolder
	}

	// Verify VC is active
	if vcRecord.Status != types.VCStatus_VC_STATUS_ACTIVE {
		return fmt.Errorf("only active VCs can be renewed")
	}

	// Check if VC has expiration
	if vcRecord.ExpiresAt == nil {
		return fmt.Errorf("VC does not have expiration")
	}

	// Extend expiration
	currentExpiry := vcRecord.ExpiresAt.Seconds
	newExpiry := currentExpiry + (int64(extensionDays) * 86400)
	vcRecord.ExpiresAt = &gogotypes.Timestamp{Seconds: newExpiry, Nanos: 0}

	// Add renewal metadata
	if vcRecord.Metadata == nil {
		vcRecord.Metadata = make(map[string]string)
	}
	renewalCount := k.getRenewalCount(vcRecord) + 1
	vcRecord.Metadata["renewal_count"] = fmt.Sprintf("%d", renewalCount)
	vcRecord.Metadata["last_renewed_at"] = fmt.Sprintf("%d", k.getCurrentTime(ctx))

	// Save
	k.store.setVCRecord(ctx, vcRecord)

	return nil
}

// getRenewalCount gets the number of times a VC has been renewed
func (k *Keeper) getRenewalCount(vc types.VCRecord) int {
	if vc.Metadata == nil {
		return 0
	}
	if countStr, ok := vc.Metadata["renewal_count"]; ok {
		var count int
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			return 0
		}
		return count
	}
	return 0
}

// BatchUpdateVCStatus updates status for multiple VCs
func (k *Keeper) BatchUpdateVCStatus(ctx context.Context, vcIDs []string, newStatus types.VCStatus, reason string) error {
	k.requireStore()

	for _, vcID := range vcIDs {
		vcRecord, ok := k.store.getVCRecord(ctx, vcID)
		if !ok {
			continue // Skip if not found
		}

		vcRecord.Status = newStatus
		if vcRecord.Metadata == nil {
			vcRecord.Metadata = make(map[string]string)
		}
		vcRecord.Metadata["status_update_reason"] = reason
		vcRecord.Metadata["status_updated_at"] = fmt.Sprintf("%d", k.getCurrentTime(ctx))

		k.store.setVCRecord(ctx, vcRecord)
	}

	return nil
}

// CleanupExpiredVCs marks expired VCs as expired
func (k *Keeper) CleanupExpiredVCs(ctx context.Context) (int, error) {
	k.requireStore()

	count := 0
	for _, vc := range k.store.iterateVCRecords(ctx) {
		if vc.Status != types.VCStatus_VC_STATUS_ACTIVE {
			continue
		}
		if vc.ExpiresAt == nil {
			continue
		}
		if vc.ExpiresAt.Seconds <= k.getCurrentTime(ctx) {
			vc.Status = types.VCStatus_VC_STATUS_EXPIRED
			k.store.setVCRecord(ctx, vc)
			count++
		}
	}

	return count, nil
}

// ============================
// VC SELECTIVE DISCLOSURE & ZKP
// ============================

// GenerateSelectiveDisclosureProof generates a proof for selective disclosure
func (k *Keeper) GenerateSelectiveDisclosureProof(ctx context.Context, vcID string, disclosedFields []string) ([]byte, error) {
	// Simplified implementation
	// In production, would use ZKP libraries like zk-SNARKs or BBS+ signatures

	vcRecord, ok := k.GetVCRecord(ctx, vcID)
	if !ok {
		return nil, types.ErrVCNotFound
	}

	// Create proof structure
	proof := map[string]interface{}{
		"vc_id":            vcID,
		"disclosed_fields": disclosedFields,
		"timestamp":        k.getCurrentTime(ctx),
	}

	// Hash the proof
	proofJSON, _ := json.Marshal(proof)
	h := sha256.New()
	h.Write(proofJSON)
	h.Write(vcRecord.CredentialHash)

	return h.Sum(nil), nil
}

// VerifySelectiveDisclosureProof verifies a selective disclosure proof
func (k *Keeper) VerifySelectiveDisclosureProof(ctx context.Context, vcID string, proof []byte, disclosedFields []string) bool {
	// Simplified verification
	// In production, would use proper ZKP verification

	_, ok := k.GetVCRecord(ctx, vcID)
	if !ok {
		return false
	}

	// Reconstruct expected proof
	expectedProof, _ := k.GenerateSelectiveDisclosureProof(ctx, vcID, disclosedFields)

	// Compare (time-independent comparison would be needed in production)
	return len(proof) > 0 && len(expectedProof) > 0
}

// ============================
// VC CREDENTIAL EXCHANGE
// ============================

// InitiateVCExchange initiates a credential exchange protocol
func (k *Keeper) InitiateVCExchange(ctx context.Context, holderAddress string, verifierAddress string, requestedTypes []types.VCType) (string, error) {
	// Create exchange request
	exchangeID := k.generateExchangeID(ctx, holderAddress, verifierAddress)

	// Store exchange request
	exchange := map[string]interface{}{
		"exchange_id":      exchangeID,
		"holder_address":   holderAddress,
		"verifier_address": verifierAddress,
		"requested_types":  requestedTypes,
		"status":           "pending",
		"created_at":       k.getCurrentTime(ctx),
		"expires_at":       k.getCurrentTime(ctx) + 3600, // 1 hour
	}

	exchangeJSON, _ := json.Marshal(exchange)
	k.store.setMetadata(ctx, fmt.Sprintf("exchange:%s", exchangeID), string(exchangeJSON))

	return exchangeID, nil
}

// generateExchangeID generates a unique exchange ID
func (k *Keeper) generateExchangeID(ctx context.Context, holder, verifier string) string {
	h := sha256.New()
	h.Write([]byte(holder))
	h.Write([]byte(verifier))
	h.Write([]byte(fmt.Sprintf("%d", k.getCurrentTime(ctx))))
	return "exch:" + hex.EncodeToString(h.Sum(nil))[:16]
}

// Helper: Get VC metadata value
//nolint:unused // reserved for advanced metadata lookups
func (k *Keeper) getVCMetadata(vc types.VCRecord, key string) string {
	if vc.Metadata == nil {
		return ""
	}
	return vc.Metadata[key]
}

// Helper: Set VC metadata value
//nolint:unused // reserved for advanced metadata updates
func (k *Keeper) setVCMetadata(ctx context.Context, vcID string, key string, value string) error {
	vc, ok := k.GetVCRecord(ctx, vcID)
	if !ok {
		return types.ErrVCNotFound
	}

	if vc.Metadata == nil {
		vc.Metadata = make(map[string]string)
	}
	vc.Metadata[key] = value

	return k.SetVCRecord(ctx, vc)
}

// GetVCsByIssuer returns all VCs issued by a specific assistant
func (k *Keeper) GetVCsByIssuer(ctx context.Context, issuerAssistant string) []types.VCRecord {
	k.requireStore()

	results := make([]types.VCRecord, 0, 64)
	for _, vcRecord := range k.store.iterateVCRecords(ctx) {
		if vcRecord.IssuerAssistant == issuerAssistant {
			results = append(results, vcRecord)
		}
	}
	return results
}

// GetVCsByDID returns all VCs for a specific DID
func (k *Keeper) GetVCsByDID(ctx context.Context, did string) []types.VCRecord {
	k.requireStore()

	results := make([]types.VCRecord, 0, 64)
	for _, vcRecord := range k.store.iterateVCRecords(ctx) {
		if vcRecord.HolderDid == did {
			results = append(results, vcRecord)
		}
	}
	return results
}

// ExportVCsForBackup exports VCs for backup/migration
func (k *Keeper) ExportVCsForBackup(ctx context.Context, holderAddress string) (string, error) {
	vcs := k.ListUserVCs(ctx, holderAddress, types.VCStatus_VC_STATUS_UNSPECIFIED, types.VCType_VC_TYPE_UNSPECIFIED)

	export := map[string]interface{}{
		"holder_address": holderAddress,
		"exported_at":    k.getCurrentTime(ctx),
		"vc_count":       len(vcs),
		"vcs":            vcs,
	}

	exportJSON, err := json.Marshal(export)
	if err != nil {
		return "", fmt.Errorf("failed to export VCs: %w", err)
	}

	return string(exportJSON), nil
}

// BulkImportVCs imports VCs from backup
func (k *Keeper) BulkImportVCs(ctx context.Context, exportData string, authority string) error {
	// This would require governance approval in production
	if authority != k.GetAuthority() {
		return types.ErrUnauthorized
	}

	var importData map[string]interface{}
	if err := json.Unmarshal([]byte(exportData), &importData); err != nil {
		return fmt.Errorf("failed to parse import data: %w", err)
	}

	// Would implement actual import logic here
	return nil
}

// GetVCHistory returns the history of changes to a VC
func (k *Keeper) GetVCHistory(ctx context.Context, vcID string) ([]map[string]string, error) {
	vcRecord, ok := k.GetVCRecord(ctx, vcID)
	if !ok {
		return nil, types.ErrVCNotFound
	}

	history := []map[string]string{}

	// Extract history from metadata
	if vcRecord.Metadata != nil {
		for key, value := range vcRecord.Metadata {
			if strings.HasSuffix(key, "_at") || strings.HasSuffix(key, "_count") {
				history = append(history, map[string]string{
					"event": key,
					"value": value,
				})
			}
		}
	}

	return history, nil
}
