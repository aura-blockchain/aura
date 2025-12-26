// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/aequitas/aura/chain/x/common/determinism"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QRCodeData represents the JSON structure encoded in the QR code
type QRCodeData struct {
	Version        string          `json:"v"`   // Protocol version
	PresentationID string          `json:"p"`   // Presentation ID
	HolderDID      string          `json:"h"`   // Holder DID
	VCIDs          []string        `json:"vcs"` // VC IDs
	Context        map[string]bool `json:"ctx"` // Context flags
	ExpiresAt      int64           `json:"exp"` // Unix timestamp
	Nonce          uint64          `json:"n"`   // Anti-replay nonce
	Signature      string          `json:"sig"` // Signature (hex)
}

// ============================
// PRESENTATION MANAGEMENT
// ============================

// CreatePresentation creates a new VC presentation (QR code)
func (k *Keeper) CreatePresentation(
	ctx context.Context,
	holderAddress string,
	vcIDs []string,
	pctx *vcregistrypb.PresentationContext,
	expiresInSeconds uint64,
) (*vcregistrypb.VCPresentation, string, error) {
	if holderAddress == "" {
		return nil, "", types.ErrInvalidHolderAddress
	}
	if len(vcIDs) == 0 {
		return nil, "", types.ErrEmptyVCList
	}
	if expiresInSeconds == 0 {
		expiresInSeconds = 300 // Default: 5 minutes
	}
	if expiresInSeconds > 3600 {
		return nil, "", types.ErrInvalidExpirationTime
	}

	// Get consensus-safe block time
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime().Unix()

	// Verify all VCs exist and are valid
	holderDID := ""
	for _, vcID := range vcIDs {
		vc, ok := k.GetVCRecord(ctx, vcID)
		if !ok {
			return nil, "", fmt.Errorf("VC not found: %s", vcID)
		}
		if vc.HolderAddress != holderAddress {
			return nil, "", types.ErrNotVCHolder
		}
		if vc.Status != types.VCStatus_VC_STATUS_ACTIVE {
			return nil, "", fmt.Errorf("VC %s is not active (status: %s)", vcID, vc.Status.String())
		}
		// Check expiration
		if vc.ExpiresAt != nil && vc.ExpiresAt.Seconds <= currentTime {
			return nil, "", fmt.Errorf("VC %s has expired", vcID)
		}
		if holderDID == "" {
			holderDID = vc.HolderDid
		}
	}

	// Generate presentation ID using deterministic RNG
	presentationID := k.generatePresentationID(ctx, holderAddress)

	// Generate nonce using deterministic RNG
	nonce := k.generateNonce(ctx)

	// Create timestamps
	createdAt := &gogotypes.Timestamp{Seconds: currentTime, Nanos: 0}
	expiresAt := &gogotypes.Timestamp{Seconds: currentTime + int64(expiresInSeconds), Nanos: 0}

	// Create presentation
	presentation := &vcregistrypb.VCPresentation{
		PresentationId:   presentationID,
		HolderDid:        holderDID,
		HolderAddress:    holderAddress,
		VcIds:            vcIDs,
		CreatedAt:        createdAt,
		Nonce:            nonce,
		ExpiresInSeconds: expiresInSeconds,
		Signature:        nil, // Will be signed by user's wallet
		Context:          pctx,
	}

	// Generate QR code data
	qrData, err := k.generateQRCodeData(presentation, expiresAt.Seconds)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate QR code data: %w", err)
	}

	// Store presentation in KV store (REQUIRED - no fallback)
	k.requireStore()
	k.store.setPresentation(ctx, types.VCPresentation(*presentation))
	k.store.appendUserPresentation(ctx, holderAddress, presentation.PresentationId)

	// Mark nonce as used for replay protection
	k.markNonceUsed(ctx, nonce)

	return presentation, qrData, nil
}

// VerifyPresentation verifies a QR code presentation
func (k *Keeper) VerifyPresentation(
	ctx context.Context,
	qrCodeData string,
	verifierAddress string,
) (*vcregistrypb.VerificationResult, error) {
	if qrCodeData == "" {
		return nil, types.ErrInvalidQRCodeData
	}

	// Get consensus-safe block time
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime().Unix()

	// Parse QR code data
	qrData, err := k.parseQRCodeData(qrCodeData)
	if err != nil {
		return &vcregistrypb.VerificationResult{
			IsValid:           false,
			VerificationError: fmt.Sprintf("Invalid QR code format: %v", err),
			VerifiedAt:        &gogotypes.Timestamp{Seconds: currentTime, Nanos: 0},
		}, nil
	}

	// Check expiration
	if qrData.ExpiresAt <= currentTime {
		return &vcregistrypb.VerificationResult{
			IsValid:           false,
			HolderDid:         qrData.HolderDID,
			VerificationError: "Presentation has expired",
			VerifiedAt:        &gogotypes.Timestamp{Seconds: currentTime, Nanos: 0},
		}, nil
	}

	// Check nonce (prevent replay attacks)
	if k.isNonceUsed(ctx, qrData.Nonce) {
		return &vcregistrypb.VerificationResult{
			IsValid:           false,
			HolderDid:         qrData.HolderDID,
			VerificationError: "Nonce has already been used (replay attack detected)",
			VerifiedAt:        &gogotypes.Timestamp{Seconds: currentTime, Nanos: 0},
		}, nil
	}

	// Verify signature
	isValidSig := k.verifyPresentationSignature(qrData)
	if !isValidSig {
		return &vcregistrypb.VerificationResult{
			IsValid:           false,
			HolderDid:         qrData.HolderDID,
			VerificationError: "Invalid signature",
			VerifiedAt:        &gogotypes.Timestamp{Seconds: currentTime, Nanos: 0},
		}, nil
	}

	// Verify each VC
	vcDetails := []*vcregistrypb.VCVerificationDetail{}
	allValid := true

	for _, vcID := range qrData.VCIDs {
		vc, ok := k.GetVCRecord(ctx, vcID)
		if !ok {
			allValid = false
			vcDetails = append(vcDetails, &vcregistrypb.VCVerificationDetail{
				VcId:    vcID,
				IsValid: false,
			})
			continue
		}

		isExpired := vc.ExpiresAt != nil && vc.ExpiresAt.Seconds <= currentTime
		isRevoked := vc.Status == types.VCStatus_VC_STATUS_REVOKED
		isValid := vc.Status == types.VCStatus_VC_STATUS_ACTIVE && !isExpired && !isRevoked

		if !isValid {
			allValid = false
		}

		vcDetails = append(vcDetails, &vcregistrypb.VCVerificationDetail{
			VcId:      vcID,
			VcType:    vcregistrypb.VCType(vc.VcType),
			Status:    vcregistrypb.VCStatus(vc.Status),
			IsValid:   isValid,
			IsExpired: isExpired,
			IsRevoked: isRevoked,
			IssuedAt:  vc.IssuedAt,
			ExpiresAt: vc.ExpiresAt,
		})
	}

	// Extract disclosed attributes based on context
	attributes := k.extractDiscloseableAttributes(qrData, vcDetails)

	result := &vcregistrypb.VerificationResult{
		IsValid:    allValid,
		HolderDid:  qrData.HolderDID,
		VcDetails:  vcDetails,
		VerifiedAt: &gogotypes.Timestamp{Seconds: currentTime, Nanos: 0},
		Attributes: attributes,
	}

	if !allValid {
		result.VerificationError = "One or more VCs are invalid, expired, or revoked"
	}

	return result, nil
}

// ============================
// HELPER METHODS
// ============================

// generatePresentationID generates a unique presentation ID using deterministic RNG
func (k *Keeper) generatePresentationID(ctx context.Context, holderAddress string) string {
	// Use deterministic ID generation based on block context
	return determinism.GenerateDeterministicID(ctx, "pres", []byte(holderAddress))
}

// generateNonce generates a deterministic nonce
func (k *Keeper) generateNonce(ctx context.Context) uint64 {
	rng := determinism.NewDeterministicRNG(ctx, []byte("vc-presentation-nonce"))
	return rng.Uint64()
}

// generateQRCodeData generates the QR code data string
func (k *Keeper) generateQRCodeData(
	presentation *vcregistrypb.VCPresentation,
	expiresAt int64,
) (string, error) {
	// Convert context to map
	contextMap := make(map[string]bool)
	if presentation.Context != nil {
		contextMap["show_full_name"] = presentation.Context.ShowFullName
		contextMap["show_age"] = presentation.Context.ShowAge
		contextMap["show_age_over_18"] = presentation.Context.ShowAgeOver_18
		contextMap["show_age_over_21"] = presentation.Context.ShowAgeOver_21
		contextMap["show_address"] = presentation.Context.ShowAddress
		contextMap["show_city_state_only"] = presentation.Context.ShowCityStateOnly
		contextMap["show_professional_license"] = presentation.Context.ShowProfessionalLicense
	}

	qrData := QRCodeData{
		Version:        "1.0",
		PresentationID: presentation.PresentationId,
		HolderDID:      presentation.HolderDid,
		VCIDs:          presentation.VcIds,
		Context:        contextMap,
		ExpiresAt:      expiresAt,
		Nonce:          presentation.Nonce,
		Signature:      "", // Will be filled by user's wallet
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(qrData)
	if err != nil {
		return "", err
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(jsonData)

	// Format as aura:// URI
	return fmt.Sprintf("aura://verify?data=%s", encoded), nil
}

// parseQRCodeData parses QR code data string
func (k *Keeper) parseQRCodeData(qrCodeData string) (*QRCodeData, error) {
	// Extract base64 data from URI
	const prefix = "aura://verify?data="
	if len(qrCodeData) < len(prefix) {
		return nil, types.ErrInvalidQRCodeData
	}
	if qrCodeData[:len(prefix)] != prefix {
		return nil, types.ErrInvalidQRCodeData
	}

	encodedData := qrCodeData[len(prefix):]

	// Decode from base64
	jsonData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Unmarshal JSON
	var qrData QRCodeData
	if err := json.Unmarshal(jsonData, &qrData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate version
	if qrData.Version != "1.0" {
		return nil, fmt.Errorf("unsupported QR code version: %s", qrData.Version)
	}

	return &qrData, nil
}

// verifyPresentationSignature verifies the signature on a presentation
func (k *Keeper) verifyPresentationSignature(qrData *QRCodeData) bool {
	if qrData.Signature == "" {
		return false
	}

	// Decode signature from hex
	signatureBytes, err := hex.DecodeString(qrData.Signature)
	if err != nil {
		return false
	}

	// Construct the message that was signed
	// Message format: presentationID || nonce || expiresAt || vcIDs
	msgHash := sha256.New()
	msgHash.Write([]byte(qrData.PresentationID))
	msgHash.Write([]byte(fmt.Sprintf("%d", qrData.Nonce)))
	msgHash.Write([]byte(fmt.Sprintf("%d", qrData.ExpiresAt)))
	for _, vcID := range qrData.VCIDs {
		msgHash.Write([]byte(vcID))
	}
	messageBytes := msgHash.Sum(nil)

	// Get holder's public key from DID document
	// In a real implementation, this would query the DID registry
	// For now, we'll derive it from the DID string
	pubKey, err := k.getPublicKeyFromDID(qrData.HolderDID)
	if err != nil {
		return false
	}

	// Verify signature using the appropriate algorithm
	return k.verifySignatureWithKey(pubKey, messageBytes, signatureBytes)
}

// getPublicKeyFromDID retrieves the public key from a DID
// In production, this would query a DID registry or resolver
func (k *Keeper) getPublicKeyFromDID(did string) (cryptotypes.PubKey, error) {
	// Extract the address from the DID
	// DID format: did:aura:<address>
	if len(did) < 10 {
		return nil, fmt.Errorf("invalid DID format")
	}

	// Parse DID to get address
	// In production, implement proper DID resolution
	addrStr := did[9:] // Skip "did:aura:"

	// Convert address string to sdk.AccAddress
	addr, err := sdk.AccAddressFromBech32(addrStr)
	if err != nil {
		return nil, fmt.Errorf("invalid address in DID: %w", err)
	}

	// For now, create a secp256k1 public key from address
	// This is a simplification - real implementation would query DID registry
	pubKeyBytes := addr.Bytes()
	if len(pubKeyBytes) >= 33 {
		return &secp256k1.PubKey{Key: pubKeyBytes[:33]}, nil
	}

	return nil, fmt.Errorf("cannot derive public key from address for DID: %s", did)
}

// verifySignatureWithKey verifies a signature using a public key
func (k *Keeper) verifySignatureWithKey(pubKey cryptotypes.PubKey, message, signature []byte) bool {
	if pubKey == nil {
		return false
	}

	// Try to verify with the public key
	// This supports both secp256k1 and ed25519 keys
	switch pk := pubKey.(type) {
	case *secp256k1.PubKey:
		// secp256k1 signature verification
		return pk.VerifySignature(message, signature)
	case *ed25519.PubKey:
		// ed25519 signature verification
		return pk.VerifySignature(message, signature)
	default:
		// Fallback: try direct verification
		return pubKey.VerifySignature(message, signature)
	}
}

// extractDiscloseableAttributes extracts attributes based on context
func (k *Keeper) extractDiscloseableAttributes(
	qrData *QRCodeData,
	vcDetails []*vcregistrypb.VCVerificationDetail,
) *vcregistrypb.DiscloseableAttributes {
	attributes := &vcregistrypb.DiscloseableAttributes{
		CustomAttributes: make(map[string]string),
	}

	// Extract attributes from VCs based on context
	// This is a simplified implementation
	// In production, would extract actual data from VC claims

	if qrData.Context["show_full_name"] {
		attributes.FullName = "John Doe" // Placeholder
	}

	if qrData.Context["show_age"] {
		attributes.Age = 25 // Placeholder
	}

	if qrData.Context["show_age_over_18"] {
		attributes.IsOver_18 = true
	}

	if qrData.Context["show_age_over_21"] {
		attributes.IsOver_21 = true
	}

	if qrData.Context["show_address"] {
		attributes.FullAddress = "123 Main St, City, State 12345" // Placeholder
	}

	if qrData.Context["show_city_state_only"] {
		attributes.CityState = "City, State" // Placeholder
	}

	return attributes
}

// storePresentationTemp stores a presentation temporarily for verification
//nolint:unused // placeholder for future stateful verification
func (k *Keeper) storePresentationTemp(presentation *vcregistrypb.VCPresentation, expiresAt int64) {
	// In a real implementation, this would store in the KV store
	// For now, we don't need to store it since verification is stateless
	// (it reads VCs directly from the chain)
}

// markNonceUsed marks a nonce as used in the KV store for replay protection
func (k *Keeper) markNonceUsed(ctx context.Context, nonce uint64) {
	k.requireStore()
	k.store.setUsedNonce(ctx, nonce)
}

// isNonceUsed checks if a nonce has been used (replay attack protection)
func (k *Keeper) isNonceUsed(ctx context.Context, nonce uint64) bool {
	k.requireStore()
	return k.store.isNonceUsed(ctx, nonce)
}

// GetPresentation retrieves a presentation by ID (if stored)
func (k *Keeper) GetPresentation(ctx context.Context, presentationID string) (*vcregistrypb.VCPresentation, bool) {
	k.requireStore()
	if pres, ok := k.store.getPresentation(ctx, presentationID); ok {
		p := vcregistrypb.VCPresentation(pres)
		return &p, true
	}

	return nil, false
}
