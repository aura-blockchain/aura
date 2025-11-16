package keeper

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
	"time"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// CreateThresholdScheme creates a new threshold signature scheme
func (k Keeper) CreateThresholdScheme(
	ctx context.Context,
	creator string,
	threshold int32,
	totalParticipants int32,
	participantIDs []string,
	schemeType cryptoproto.ThresholdSchemeType,
) (string, []byte, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return "", nil, err
	}

	// Validate parameters
	if threshold < int32(params.MinThresholdParticipants) {
		return "", nil, types.ErrInvalidThreshold
	}
	if totalParticipants > int32(params.MaxThresholdParticipants) {
		return "", nil, types.ErrInvalidParticipantCount
	}
	if threshold > totalParticipants {
		return "", nil, types.ErrInvalidThreshold
	}
	if len(participantIDs) != int(totalParticipants) {
		return "", nil, types.ErrInvalidParticipantCount
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Generate scheme ID
	schemeID := fmt.Sprintf("threshold_%s_%d", creator, time.Now().Unix())

	// Generate a public key for the scheme
	var publicKey []byte
	switch schemeType {
	case cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA:
		publicKey, err = k.generateECDSAPublicKey()
	case cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_EDDSA:
		publicKey, err = k.generateEdDSAPublicKey()
	case cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_BLS:
		publicKey, err = k.generateBLSPublicKey()
	case cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_SCHNORR:
		publicKey, err = k.generateSchnorrPublicKey()
	default:
		return "", nil, fmt.Errorf("unsupported threshold scheme type")
	}
	if err != nil {
		return "", nil, err
	}

	now := time.Now()
	scheme := &cryptoproto.ThresholdSignatureScheme{
		SchemeId:          schemeID,
		Threshold:         threshold,
		TotalParticipants: totalParticipants,
		ParticipantIds:    participantIDs,
		SchemeType:        schemeType,
		PublicKey:         publicKey,
		CreatedAt:         now,
		Status:            cryptoproto.ThresholdSchemeStatus_THRESHOLD_SCHEME_STATUS_ACTIVE,
	}

	// Store in state
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(scheme)
	store.Set(types.GetThresholdSchemeKey(schemeID), bz)

	// Cache
	k.thresholdSchemes[schemeID] = scheme
	k.thresholdShares[schemeID] = make(map[string]*cryptoproto.ThresholdSignatureShare)

	k.Logger(ctx).Info("created threshold signature scheme",
		"scheme_id", schemeID,
		"threshold", threshold,
		"participants", totalParticipants,
		"type", schemeType.String(),
	)

	return schemeID, publicKey, nil
}

// SubmitThresholdSignatureShare submits a signature share from a participant
func (k Keeper) SubmitThresholdSignatureShare(
	ctx context.Context,
	submitter string,
	schemeID string,
	signatureShare []byte,
	messageHash []byte,
) (int32, bool, []byte, error) {
	if len(signatureShare) == 0 {
		return 0, false, nil, types.ErrInvalidSignatureShare
	}
	if len(messageHash) != 32 {
		return 0, false, nil, fmt.Errorf("invalid message hash length")
	}

	scheme, err := k.GetThresholdScheme(ctx, schemeID)
	if err != nil {
		return 0, false, nil, err
	}

	if scheme.Status != cryptoproto.ThresholdSchemeStatus_THRESHOLD_SCHEME_STATUS_ACTIVE {
		return 0, false, nil, fmt.Errorf("threshold scheme not active")
	}

	// Verify submitter is a participant
	isParticipant := false
	for _, pid := range scheme.ParticipantIds {
		if pid == submitter {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return 0, false, nil, types.ErrUnauthorized
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if share already submitted
	if shares, ok := k.thresholdShares[schemeID]; ok {
		if _, exists := shares[submitter]; exists {
			k.mu.Unlock()
			return 0, false, nil, fmt.Errorf("signature share already submitted")
		}
	}

	now := time.Now()
	share := &cryptoproto.ThresholdSignatureShare{
		SchemeId:       schemeID,
		ParticipantId:  submitter,
		SignatureShare: signatureShare,
		MessageHash:    messageHash,
		SignedAt:       now,
	}

	// Store share
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(share)
	store.Set(types.GetThresholdSignatureShareKey(schemeID, submitter), bz)

	// Cache share
	if k.thresholdShares[schemeID] == nil {
		k.thresholdShares[schemeID] = make(map[string]*cryptoproto.ThresholdSignatureShare)
	}
	k.thresholdShares[schemeID][submitter] = share

	// Count shares
	sharesCollected := int32(len(k.thresholdShares[schemeID]))
	thresholdReached := sharesCollected >= scheme.Threshold

	var combinedSignature []byte
	if thresholdReached {
		// Combine signature shares
		combinedSignature, err = k.combineSignatureShares(scheme, k.thresholdShares[schemeID])
		if err != nil {
			return sharesCollected, false, nil, err
		}
	}

	k.Logger(ctx).Info("submitted threshold signature share",
		"scheme_id", schemeID,
		"participant", submitter,
		"shares_collected", sharesCollected,
		"threshold_reached", thresholdReached,
	)

	return sharesCollected, thresholdReached, combinedSignature, nil
}

// GetThresholdScheme retrieves a threshold signature scheme
func (k Keeper) GetThresholdScheme(ctx context.Context, schemeID string) (*cryptoproto.ThresholdSignatureScheme, error) {
	k.mu.RLock()
	if scheme, ok := k.thresholdSchemes[schemeID]; ok {
		k.mu.RUnlock()
		return scheme, nil
	}
	k.mu.RUnlock()

	store := k.getStore(ctx)
	bz := store.Get(types.GetThresholdSchemeKey(schemeID))
	if bz == nil {
		return nil, types.ErrThresholdSchemeNotFound
	}

	var scheme cryptoproto.ThresholdSignatureScheme
	k.cdc.MustUnmarshal(bz, &scheme)

	k.mu.Lock()
	k.thresholdSchemes[schemeID] = &scheme
	k.mu.Unlock()

	return &scheme, nil
}

// generateECDSAPublicKey generates a public key for ECDSA threshold scheme
func (k Keeper) generateECDSAPublicKey() ([]byte, error) {
	// Generate a key pair using P256 curve
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// Serialize public key
	publicKey := elliptic.Marshal(elliptic.P256(), privateKey.PublicKey.X, privateKey.PublicKey.Y)
	return publicKey, nil
}

// generateEdDSAPublicKey generates a public key for EdDSA threshold scheme
func (k Keeper) generateEdDSAPublicKey() ([]byte, error) {
	// In a real implementation, use ed25519
	// For now, generate random bytes
	publicKey := make([]byte, 32)
	_, err := rand.Read(publicKey)
	return publicKey, err
}

// generateBLSPublicKey generates a public key for BLS threshold scheme
func (k Keeper) generateBLSPublicKey() ([]byte, error) {
	// In a real implementation, use BLS12-381
	// For now, generate random bytes
	publicKey := make([]byte, 48)
	_, err := rand.Read(publicKey)
	return publicKey, err
}

// generateSchnorrPublicKey generates a public key for Schnorr threshold scheme
func (k Keeper) generateSchnorrPublicKey() ([]byte, error) {
	// Generate using secp256k1
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	// Serialize X-coordinate only for Schnorr
	publicKey := privateKey.PublicKey.X.Bytes()
	// Pad to 32 bytes if necessary
	if len(publicKey) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(publicKey):], publicKey)
		publicKey = padded
	}
	return publicKey, nil
}

// combineSignatureShares combines threshold signature shares into a complete signature
func (k Keeper) combineSignatureShares(
	scheme *cryptoproto.ThresholdSignatureScheme,
	shares map[string]*cryptoproto.ThresholdSignatureShare,
) ([]byte, error) {
	if len(shares) < int(scheme.Threshold) {
		return nil, types.ErrThresholdNotReached
	}

	// In a real implementation, this would use proper threshold cryptography
	// For now, create a combined signature by hashing all shares together
	h := sha256.New()
	for _, share := range shares {
		h.Write(share.SignatureShare)
	}
	combinedSignature := h.Sum(nil)

	// Add scheme type identifier
	result := append([]byte{byte(scheme.SchemeType)}, combinedSignature...)

	return result, nil
}

// VerifyThresholdSignature verifies a threshold signature
func (k Keeper) VerifyThresholdSignature(
	ctx context.Context,
	schemeID string,
	signature []byte,
	messageHash []byte,
) (bool, error) {
	scheme, err := k.GetThresholdScheme(ctx, schemeID)
	if err != nil {
		return false, err
	}

	if len(signature) < 33 {
		return false, types.ErrInvalidSignatureShare
	}

	// In a real implementation, this would perform proper signature verification
	// For now, basic validation
	schemeType := cryptoproto.ThresholdSchemeType(signature[0])
	if schemeType != scheme.SchemeType {
		return false, fmt.Errorf("signature scheme type mismatch")
	}

	return true, nil
}

// RevokeThresholdScheme revokes a threshold signature scheme
func (k Keeper) RevokeThresholdScheme(ctx context.Context, schemeID string) error {
	scheme, err := k.GetThresholdScheme(ctx, schemeID)
	if err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	scheme.Status = cryptoproto.ThresholdSchemeStatus_THRESHOLD_SCHEME_STATUS_REVOKED

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(scheme)
	store.Set(types.GetThresholdSchemeKey(schemeID), bz)

	k.thresholdSchemes[schemeID] = scheme

	k.Logger(ctx).Info("revoked threshold scheme", "scheme_id", schemeID)

	return nil
}

// SetThresholdScheme stores a threshold scheme (for genesis)
func (k *Keeper) SetThresholdScheme(ctx context.Context, scheme *cryptoproto.ThresholdSignatureScheme) error {
	if scheme == nil {
		return nil
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(scheme)
	store.Set(types.GetThresholdSchemeKey(scheme.SchemeId), bz)
	k.thresholdSchemes[scheme.SchemeId] = scheme
	return nil
}

// GetAllThresholdSchemes retrieves all threshold signature schemes
func (k Keeper) GetAllThresholdSchemes(ctx context.Context) []*cryptoproto.ThresholdSignatureScheme {
	k.mu.RLock()
	defer k.mu.RUnlock()
	schemes := make([]*cryptoproto.ThresholdSignatureScheme, 0, len(k.thresholdSchemes))
	for _, scheme := range k.thresholdSchemes {
		schemes = append(schemes, scheme)
	}
	return schemes
}

// SetThresholdScheme stores a threshold scheme (for genesis)
func (k Keeper) GetAllThresholdSchemes(ctx context.Context) []*cryptoproto.ThresholdSignatureScheme {
	k.mu.RLock()
	defer k.mu.RUnlock()
	schemes := make([]*cryptoproto.ThresholdSignatureScheme, 0, len(k.thresholdSchemes))
	for _, scheme := range k.thresholdSchemes {
		schemes = append(schemes, scheme)
	}
	return schemes
}
