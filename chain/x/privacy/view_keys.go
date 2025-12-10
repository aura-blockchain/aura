package privacy

// This file implements OFF-CHAIN view key utilities for wallet software.
//
// IMPORTANT: All key generation functions in this file are OFF-CHAIN utilities
// and MUST NOT be called from consensus-critical code (message handlers,
// BeginBlocker, EndBlocker). These functions use crypto/rand which is
// non-deterministic and would break consensus if used on-chain.
//
// View keys allow selective disclosure of transaction information without
// compromising spending authority. This is a critical privacy feature
// that enables:
// - Regulatory compliance (auditors can view but not spend)
// - Read-only wallet sharing (view balance without spending risk)
// - Transaction monitoring for businesses
// - Selective transparency for tax purposes
//
// This implementation follows the dual-key model used in privacy-focused
// blockchains like Monero and Zcash, adapted for the AURA ecosystem.
//
// ON-CHAIN VS OFF-CHAIN SEPARATION:
// - OFF-CHAIN: GenerateViewKey(), DelegateViewKey(), BatchGenerateViewKeys()
//   These functions generate new cryptographic keys using crypto/rand.
//   They are utilities for wallet software to create view keys locally.
//
// - ON-CHAIN: MsgRegisterViewKey handler stores only PUBLIC view keys.
//   No private keys are ever transmitted to or stored on the blockchain.
//   No key generation occurs during consensus.
//
// The message handler (msg_server.go) only stores public view keys that were
// generated OFF-CHAIN by wallet software. Private view keys remain with the user.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ViewKeyType defines the type of view key and its capabilities
type ViewKeyType string

const (
	// ViewKeyTypeIncoming allows viewing incoming transactions only
	ViewKeyTypeIncoming ViewKeyType = "INCOMING"
	// ViewKeyTypeOutgoing allows viewing outgoing transactions only
	ViewKeyTypeOutgoing ViewKeyType = "OUTGOING"
	// ViewKeyTypeFull allows viewing all transactions (incoming and outgoing)
	ViewKeyTypeFull ViewKeyType = "FULL"
	// ViewKeyTypeAudit provides read-only audit access with additional metadata
	ViewKeyTypeAudit ViewKeyType = "AUDIT"
	// ViewKeyTypeBalance allows viewing balance only without transaction details
	ViewKeyTypeBalance ViewKeyType = "BALANCE"
)

// ViewKey represents a cryptographic view key for transaction visibility
type ViewKey struct {
	// PublicKey is the public portion that can be shared
	PublicKey []byte
	// PrivateKey is the secret portion used to derive decryption keys
	PrivateKey []byte
	// Type defines the access level this key provides
	Type ViewKeyType
	// OwnerAddress is the address this view key is associated with
	OwnerAddress []byte
	// Permissions are specific access rights granted by this key
	Permissions []string
	// CreatedAt is when this view key was generated
	CreatedAt time.Time
	// ExpiresAt is when this view key expires (nil for non-expiring)
	ExpiresAt *time.Time
	// Revoked indicates if this key has been revoked
	Revoked bool
	// RevokedAt is when this key was revoked (if applicable)
	RevokedAt *time.Time
	// Label is a human-readable identifier for this key
	Label string
	// Metadata contains additional key-specific information
	Metadata map[string]string
}

// ViewKeyManager manages view keys for addresses
type ViewKeyManager struct {
	curve    elliptic.Curve
	viewKeys map[string]*ViewKey // Keyed by hex-encoded public key
	byOwner  map[string][]*ViewKey // Keyed by hex-encoded owner address
	mu       sync.RWMutex
}

// NewViewKeyManager creates a new view key manager
func NewViewKeyManager() *ViewKeyManager {
	return &ViewKeyManager{
		curve:    elliptic.P256(),
		viewKeys: make(map[string]*ViewKey),
		byOwner:  make(map[string][]*ViewKey),
	}
}

// GenerateViewKey generates a new view key for an address
func (vkm *ViewKeyManager) GenerateViewKey(
	keyType ViewKeyType,
	ownerAddress []byte,
	permissions []string,
	expiresAt *time.Time,
	now time.Time, // Accept time parameter for determinism
) (*ViewKey, error) {
	vkm.mu.Lock()
	defer vkm.mu.Unlock()

	if len(ownerAddress) == 0 {
		return nil, errors.New("owner address cannot be empty")
	}

	if !isValidViewKeyType(keyType) {
		return nil, fmt.Errorf("invalid view key type: %s", keyType)
	}

	// Generate ECDSA key pair for the view key
	privateKey, err := ecdsa.GenerateKey(vkm.curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Serialize keys
	privKeyBytes := privateKey.D.Bytes()
	pubKeyBytes := elliptic.MarshalCompressed(vkm.curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)

	// Apply default permissions based on key type if none provided
	if len(permissions) == 0 {
		permissions = getDefaultPermissions(keyType)
	}

	viewKey := &ViewKey{
		PublicKey:    pubKeyBytes,
		PrivateKey:   privKeyBytes,
		Type:         keyType,
		OwnerAddress: ownerAddress,
		Permissions:  permissions,
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
		Revoked:      false,
		Metadata:     make(map[string]string),
	}

	// Store the view key
	pubKeyHex := hex.EncodeToString(pubKeyBytes)
	vkm.viewKeys[pubKeyHex] = viewKey

	// Index by owner
	ownerHex := hex.EncodeToString(ownerAddress)
	vkm.byOwner[ownerHex] = append(vkm.byOwner[ownerHex], viewKey)

	return viewKey, nil
}

// GetViewKey retrieves a view key by its public key
func (vkm *ViewKeyManager) GetViewKey(publicKey []byte, now time.Time) (*ViewKey, error) {
	vkm.mu.RLock()
	defer vkm.mu.RUnlock()

	pubKeyHex := hex.EncodeToString(publicKey)
	viewKey, exists := vkm.viewKeys[pubKeyHex]
	if !exists {
		return nil, errors.New("view key not found")
	}

	// Check if revoked
	if viewKey.Revoked {
		return nil, errors.New("view key has been revoked")
	}

	// Check if expired
	if viewKey.ExpiresAt != nil && now.After(*viewKey.ExpiresAt) {
		return nil, errors.New("view key has expired")
	}

	return viewKey, nil
}

// RevokeViewKey revokes a view key
func (vkm *ViewKeyManager) RevokeViewKey(publicKey []byte, now time.Time) error {
	vkm.mu.Lock()
	defer vkm.mu.Unlock()

	pubKeyHex := hex.EncodeToString(publicKey)
	viewKey, exists := vkm.viewKeys[pubKeyHex]
	if !exists {
		return errors.New("view key not found")
	}

	if viewKey.Revoked {
		return errors.New("view key is already revoked")
	}

	viewKey.Revoked = true
	viewKey.RevokedAt = &now

	// Remove from byOwner index
	ownerHex := hex.EncodeToString(viewKey.OwnerAddress)
	ownerKeys := vkm.byOwner[ownerHex]
	for i, key := range ownerKeys {
		if hex.EncodeToString(key.PublicKey) == pubKeyHex {
			vkm.byOwner[ownerHex] = append(ownerKeys[:i], ownerKeys[i+1:]...)
			break
		}
	}

	// Remove from main index
	delete(vkm.viewKeys, pubKeyHex)

	return nil
}

// VerifyPermission checks if a view key has a specific permission
func (vkm *ViewKeyManager) VerifyPermission(publicKey []byte, permission string, now time.Time) (bool, error) {
	vkm.mu.RLock()
	defer vkm.mu.RUnlock()

	pubKeyHex := hex.EncodeToString(publicKey)
	viewKey, exists := vkm.viewKeys[pubKeyHex]
	if !exists {
		return false, errors.New("view key not found")
	}

	// Check if revoked
	if viewKey.Revoked {
		return false, errors.New("view key has been revoked")
	}

	// Check if expired
	if viewKey.ExpiresAt != nil && now.After(*viewKey.ExpiresAt) {
		return false, errors.New("view key has expired")
	}

	// Check permission
	for _, p := range viewKey.Permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// ListActiveViewKeys returns all active (non-revoked, non-expired) view keys for an address
func (vkm *ViewKeyManager) ListActiveViewKeys(ownerAddress []byte, now time.Time) []*ViewKey {
	vkm.mu.RLock()
	defer vkm.mu.RUnlock()

	ownerHex := hex.EncodeToString(ownerAddress)
	ownerKeys := vkm.byOwner[ownerHex]

	activeKeys := make([]*ViewKey, 0)
	for _, key := range ownerKeys {
		if !key.Revoked && (key.ExpiresAt == nil || now.Before(*key.ExpiresAt)) {
			activeKeys = append(activeKeys, key)
		}
	}

	return activeKeys
}

// UpdatePermissions updates the permissions for a view key
func (vkm *ViewKeyManager) UpdatePermissions(publicKey []byte, newPermissions []string) error {
	vkm.mu.Lock()
	defer vkm.mu.Unlock()

	pubKeyHex := hex.EncodeToString(publicKey)
	viewKey, exists := vkm.viewKeys[pubKeyHex]
	if !exists {
		return errors.New("view key not found")
	}

	if viewKey.Revoked {
		return errors.New("cannot update revoked view key")
	}

	viewKey.Permissions = newPermissions
	return nil
}

// SetLabel sets a label for a view key
func (vkm *ViewKeyManager) SetLabel(publicKey []byte, label string) error {
	vkm.mu.Lock()
	defer vkm.mu.Unlock()

	pubKeyHex := hex.EncodeToString(publicKey)
	viewKey, exists := vkm.viewKeys[pubKeyHex]
	if !exists {
		return errors.New("view key not found")
	}

	viewKey.Label = label
	return nil
}

// SetMetadata sets metadata for a view key
func (vkm *ViewKeyManager) SetMetadata(publicKey []byte, key, value string) error {
	vkm.mu.Lock()
	defer vkm.mu.Unlock()

	pubKeyHex := hex.EncodeToString(publicKey)
	viewKey, exists := vkm.viewKeys[pubKeyHex]
	if !exists {
		return errors.New("view key not found")
	}

	viewKey.Metadata[key] = value
	return nil
}

// ExtendExpiry extends the expiry time for a view key
func (vkm *ViewKeyManager) ExtendExpiry(publicKey []byte, newExpiryTime time.Time, now time.Time) error {
	vkm.mu.Lock()
	defer vkm.mu.Unlock()

	pubKeyHex := hex.EncodeToString(publicKey)
	viewKey, exists := vkm.viewKeys[pubKeyHex]
	if !exists {
		return errors.New("view key not found")
	}

	if viewKey.Revoked {
		return errors.New("cannot extend expiry for revoked view key")
	}

	if viewKey.ExpiresAt != nil && now.After(*viewKey.ExpiresAt) {
		return errors.New("view key has already expired")
	}

	viewKey.ExpiresAt = &newExpiryTime
	return nil
}

// DeriveSharedSecret derives a shared secret using the view key for decryption
func (vkm *ViewKeyManager) DeriveSharedSecret(publicKey []byte, txPublicKey []byte) ([]byte, error) {
	vkm.mu.RLock()
	defer vkm.mu.RUnlock()

	pubKeyHex := hex.EncodeToString(publicKey)
	viewKey, exists := vkm.viewKeys[pubKeyHex]
	if !exists {
		return nil, errors.New("view key not found")
	}

	if viewKey.Revoked {
		return nil, errors.New("view key has been revoked")
	}

	// Unmarshal the transaction public key
	txPubX, txPubY := elliptic.UnmarshalCompressed(vkm.curve, txPublicKey)
	if txPubX == nil {
		return nil, errors.New("invalid transaction public key")
	}

	// Perform ECDH: viewPrivateKey * txPublicKey
	viewPriv := viewKey.PrivateKey
	sharedX, sharedY := vkm.curve.ScalarMult(txPubX, txPubY, viewPriv)

	// Hash the shared secret for use as decryption key
	hasher := sha256.New()
	hasher.Write(elliptic.Marshal(vkm.curve, sharedX, sharedY))
	hasher.Write([]byte("view_key_shared_secret"))

	return hasher.Sum(nil), nil
}

// CanViewTransaction checks if a view key can view a specific transaction type
func (vkm *ViewKeyManager) CanViewTransaction(publicKey []byte, isIncoming bool, now time.Time) (bool, error) {
	viewKey, err := vkm.GetViewKey(publicKey, now)
	if err != nil {
		return false, err
	}

	switch viewKey.Type {
	case ViewKeyTypeFull, ViewKeyTypeAudit:
		return true, nil
	case ViewKeyTypeIncoming:
		return isIncoming, nil
	case ViewKeyTypeOutgoing:
		return !isIncoming, nil
	case ViewKeyTypeBalance:
		return false, nil // Balance-only keys cannot view transaction details
	default:
		return false, nil
	}
}

// GetViewKeyStats returns statistics about view keys for an address
func (vkm *ViewKeyManager) GetViewKeyStats(ownerAddress []byte, now time.Time) map[string]interface{} {
	vkm.mu.RLock()
	defer vkm.mu.RUnlock()

	ownerHex := hex.EncodeToString(ownerAddress)
	ownerKeys := vkm.byOwner[ownerHex]

	stats := map[string]interface{}{
		"total_keys":   len(ownerKeys),
		"active_keys":  0,
		"revoked_keys": 0,
		"expired_keys": 0,
		"by_type":      make(map[ViewKeyType]int),
	}

	byType := stats["by_type"].(map[ViewKeyType]int)

	for _, key := range ownerKeys {
		if key.Revoked {
			stats["revoked_keys"] = stats["revoked_keys"].(int) + 1
		} else if key.ExpiresAt != nil && now.After(*key.ExpiresAt) {
			stats["expired_keys"] = stats["expired_keys"].(int) + 1
		} else {
			stats["active_keys"] = stats["active_keys"].(int) + 1
		}
		byType[key.Type]++
	}

	return stats
}

// ViewKeyProof generates a cryptographic proof that a view key is valid for an address
type ViewKeyProof struct {
	PublicKey    []byte
	OwnerAddress []byte
	Signature    []byte
	Timestamp    time.Time
}

// GenerateOwnershipProof generates a proof that the view key belongs to an address
func (vkm *ViewKeyManager) GenerateOwnershipProof(publicKey []byte, now time.Time) (*ViewKeyProof, error) {
	viewKey, err := vkm.GetViewKey(publicKey, now)
	if err != nil {
		return nil, err
	}

	// Create message to sign
	message := make([]byte, 0)
	message = append(message, viewKey.PublicKey...)
	message = append(message, viewKey.OwnerAddress...)
	message = append(message, []byte(now.String())...)

	// Sign the message with the view key's private key
	hasher := sha256.New()
	hasher.Write(message)
	messageHash := hasher.Sum(nil)

	// Create a simple signature (in production, use proper ECDSA signing)
	signatureHasher := sha256.New()
	signatureHasher.Write(messageHash)
	signatureHasher.Write(viewKey.PrivateKey)
	signature := signatureHasher.Sum(nil)

	return &ViewKeyProof{
		PublicKey:    viewKey.PublicKey,
		OwnerAddress: viewKey.OwnerAddress,
		Signature:    signature,
		Timestamp:    now,
	}, nil
}

// VerifyOwnershipProof verifies a view key ownership proof
func VerifyOwnershipProof(proof *ViewKeyProof) bool {
	if proof == nil {
		return false
	}

	if len(proof.PublicKey) == 0 || len(proof.OwnerAddress) == 0 || len(proof.Signature) == 0 {
		return false
	}

	// Basic structural verification
	// In production, this would verify the ECDSA signature
	return len(proof.Signature) == 32
}

// Helper functions

func isValidViewKeyType(keyType ViewKeyType) bool {
	switch keyType {
	case ViewKeyTypeIncoming, ViewKeyTypeOutgoing, ViewKeyTypeFull, ViewKeyTypeAudit, ViewKeyTypeBalance:
		return true
	default:
		return false
	}
}

func getDefaultPermissions(keyType ViewKeyType) []string {
	switch keyType {
	case ViewKeyTypeIncoming:
		return []string{"view_incoming", "view_balance"}
	case ViewKeyTypeOutgoing:
		return []string{"view_outgoing"}
	case ViewKeyTypeFull:
		return []string{"view_incoming", "view_outgoing", "view_balance", "view_history"}
	case ViewKeyTypeAudit:
		return []string{"view_incoming", "view_outgoing", "view_balance", "view_history", "view_metadata", "export_reports"}
	case ViewKeyTypeBalance:
		return []string{"view_balance"}
	default:
		return []string{}
	}
}

// ViewKeyDelegation allows delegating view key capabilities to another party
type ViewKeyDelegation struct {
	OriginalViewKey  []byte
	DelegatedKey     []byte
	DelegatedTo      []byte
	Permissions      []string
	CreatedAt        time.Time
	ExpiresAt        *time.Time
	DelegationProof  []byte
}

// DelegateViewKey creates a delegated view key with reduced permissions
func (vkm *ViewKeyManager) DelegateViewKey(
	originalPubKey []byte,
	delegateTo []byte,
	permissions []string,
	expiresAt *time.Time,
	now time.Time, // Accept time parameter for determinism
) (*ViewKeyDelegation, error) {
	vkm.mu.Lock()
	defer vkm.mu.Unlock()

	pubKeyHex := hex.EncodeToString(originalPubKey)
	originalKey, exists := vkm.viewKeys[pubKeyHex]
	if !exists {
		return nil, errors.New("original view key not found")
	}

	if originalKey.Revoked {
		return nil, errors.New("cannot delegate revoked view key")
	}

	// Verify requested permissions are subset of original
	for _, reqPerm := range permissions {
		found := false
		for _, origPerm := range originalKey.Permissions {
			if reqPerm == origPerm {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("cannot delegate permission %s: not in original key", reqPerm)
		}
	}

	// Generate new key pair for delegation
	delegatedPriv, err := ecdsa.GenerateKey(vkm.curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate delegated key: %w", err)
	}

	delegatedPubKey := elliptic.MarshalCompressed(vkm.curve, delegatedPriv.PublicKey.X, delegatedPriv.PublicKey.Y)

	// Create delegation proof
	hasher := sha256.New()
	hasher.Write(originalPubKey)
	hasher.Write(delegatedPubKey)
	hasher.Write(delegateTo)
	for _, p := range permissions {
		hasher.Write([]byte(p))
	}
	delegationProof := hasher.Sum(nil)

	// Store delegated view key
	delegatedViewKey := &ViewKey{
		PublicKey:    delegatedPubKey,
		PrivateKey:   delegatedPriv.D.Bytes(),
		Type:         originalKey.Type,
		OwnerAddress: delegateTo,
		Permissions:  permissions,
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
		Revoked:      false,
		Metadata: map[string]string{
			"delegation_source": pubKeyHex,
		},
	}

	delegatedPubKeyHex := hex.EncodeToString(delegatedPubKey)
	vkm.viewKeys[delegatedPubKeyHex] = delegatedViewKey

	delegateToHex := hex.EncodeToString(delegateTo)
	vkm.byOwner[delegateToHex] = append(vkm.byOwner[delegateToHex], delegatedViewKey)

	return &ViewKeyDelegation{
		OriginalViewKey: originalPubKey,
		DelegatedKey:    delegatedPubKey,
		DelegatedTo:     delegateTo,
		Permissions:     permissions,
		CreatedAt:       now,
		ExpiresAt:       expiresAt,
		DelegationProof: delegationProof,
	}, nil
}

// BatchGenerateViewKeys generates multiple view keys for an address
func (vkm *ViewKeyManager) BatchGenerateViewKeys(
	ownerAddress []byte,
	keyTypes []ViewKeyType,
	now time.Time, // Accept time parameter for determinism
) ([]*ViewKey, error) {
	if len(keyTypes) == 0 {
		return nil, errors.New("no key types specified")
	}

	viewKeys := make([]*ViewKey, 0, len(keyTypes))
	for _, keyType := range keyTypes {
		viewKey, err := vkm.GenerateViewKey(keyType, ownerAddress, nil, nil, now)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s view key: %w", keyType, err)
		}
		viewKeys = append(viewKeys, viewKey)
	}

	return viewKeys, nil
}

// CleanupExpiredKeys removes expired view keys from the manager
func (vkm *ViewKeyManager) CleanupExpiredKeys(now time.Time) int {
	vkm.mu.Lock()
	defer vkm.mu.Unlock()

	removed := 0

	for pubKeyHex, viewKey := range vkm.viewKeys {
		if viewKey.ExpiresAt != nil && now.After(*viewKey.ExpiresAt) {
			delete(vkm.viewKeys, pubKeyHex)
			removed++

			// Remove from owner index
			ownerHex := hex.EncodeToString(viewKey.OwnerAddress)
			ownerKeys := vkm.byOwner[ownerHex]
			for i, key := range ownerKeys {
				if hex.EncodeToString(key.PublicKey) == pubKeyHex {
					vkm.byOwner[ownerHex] = append(ownerKeys[:i], ownerKeys[i+1:]...)
					break
				}
			}
		}
	}

	return removed
}

// ExportViewKey exports a view key in a serialized format
func (vkm *ViewKeyManager) ExportViewKey(publicKey []byte, includePrivate bool, now time.Time) (map[string]interface{}, error) {
	viewKey, err := vkm.GetViewKey(publicKey, now)
	if err != nil {
		return nil, err
	}

	export := map[string]interface{}{
		"public_key":    hex.EncodeToString(viewKey.PublicKey),
		"type":          string(viewKey.Type),
		"owner_address": hex.EncodeToString(viewKey.OwnerAddress),
		"permissions":   viewKey.Permissions,
		"created_at":    viewKey.CreatedAt.Unix(),
		"label":         viewKey.Label,
		"metadata":      viewKey.Metadata,
	}

	if viewKey.ExpiresAt != nil {
		export["expires_at"] = viewKey.ExpiresAt.Unix()
	}

	if includePrivate {
		export["private_key"] = hex.EncodeToString(viewKey.PrivateKey)
	}

	return export, nil
}
