package keeper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

// KeyMetadata holds metadata about an encryption key
type KeyMetadata struct {
	KeyID      string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	RotatedAt  *time.Time
	Status     KeyStatus
	Version    uint32
	Algorithm  string
	UsageCount uint64
	LastUsedAt time.Time
}

// KeyStatus represents the status of an encryption key
type KeyStatus int

const (
	KeyStatusActive KeyStatus = iota
	KeyStatusRotating
	KeyStatusDeprecated
	KeyStatusRevoked
)

func (s KeyStatus) String() string {
	switch s {
	case KeyStatusActive:
		return "active"
	case KeyStatusRotating:
		return "rotating"
	case KeyStatusDeprecated:
		return "deprecated"
	case KeyStatusRevoked:
		return "revoked"
	default:
		return "unknown"
	}
}

// KeyRotationSchedule defines when keys should be rotated
type KeyRotationSchedule struct {
	RotationIntervalHours uint32
	MaxKeyAge             time.Duration
	AutoRotateEnabled     bool
	GracePeriodHours      uint32 // Time to keep old keys for decryption
	MinKeysToRetain       int
}

// KMSInterface defines the interface for Key Management System integration
type KMSInterface interface {
	GenerateKey(keyID string, algorithm string) ([]byte, error)
	EncryptKey(keyData []byte) ([]byte, error)
	DecryptKey(encryptedKeyData []byte) ([]byte, error)
	RevokeKey(keyID string) error
	ListKeys() ([]string, error)
	GetKeyMetadata(keyID string) (*KeyMetadata, error)
}

// LocalKMS is a simple local implementation of KMS for development/testing
type LocalKMS struct {
	keys      map[string][]byte
	metadata  map[string]*KeyMetadata
	masterKey []byte
}

// NewLocalKMS creates a new local KMS instance
func NewLocalKMS() (*LocalKMS, error) {
	masterKey := make([]byte, 32)
	if _, err := rand.Read(masterKey); err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}

	return &LocalKMS{
		keys:      make(map[string][]byte),
		metadata:  make(map[string]*KeyMetadata),
		masterKey: masterKey,
	}, nil
}

// GenerateKey generates a new encryption key
func (kms *LocalKMS) GenerateKey(keyID string, algorithm string) ([]byte, error) {
	keySize := 32 // 256-bit for AES-256
	key := make([]byte, keySize)

	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Encrypt the key with master key before storing
	encryptedKey, err := kms.encryptWithMasterKey(key)
	if err != nil {
		return nil, err
	}

	kms.keys[keyID] = encryptedKey
	kms.metadata[keyID] = &KeyMetadata{
		KeyID:      keyID,
		CreatedAt:  time.Now(),
		Status:     KeyStatusActive,
		Version:    1,
		Algorithm:  algorithm,
		UsageCount: 0,
	}

	return key, nil
}

// EncryptKey encrypts a key for storage
func (kms *LocalKMS) EncryptKey(keyData []byte) ([]byte, error) {
	return kms.encryptWithMasterKey(keyData)
}

// DecryptKey decrypts a stored key
func (kms *LocalKMS) DecryptKey(encryptedKeyData []byte) ([]byte, error) {
	return kms.decryptWithMasterKey(encryptedKeyData)
}

// RevokeKey revokes a key
func (kms *LocalKMS) RevokeKey(keyID string) error {
	if metadata, ok := kms.metadata[keyID]; ok {
		metadata.Status = KeyStatusRevoked
		return nil
	}
	return fmt.Errorf("key not found: %s", keyID)
}

// ListKeys returns all key IDs
func (kms *LocalKMS) ListKeys() ([]string, error) {
	keys := make([]string, 0, len(kms.keys))
	for keyID := range kms.keys {
		keys = append(keys, keyID)
	}
	return keys, nil
}

// GetKeyMetadata returns metadata for a key
func (kms *LocalKMS) GetKeyMetadata(keyID string) (*KeyMetadata, error) {
	if metadata, ok := kms.metadata[keyID]; ok {
		return metadata, nil
	}
	return nil, fmt.Errorf("key metadata not found: %s", keyID)
}

// encryptWithMasterKey encrypts data with the master key
func (kms *LocalKMS) encryptWithMasterKey(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(kms.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// decryptWithMasterKey decrypts data with the master key
func (kms *LocalKMS) decryptWithMasterKey(encryptedData []byte) ([]byte, error) {
	block, err := aes.NewCipher(kms.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// ============================
// KEY ROTATION IMPLEMENTATION
// ============================

// SetKMS sets the KMS implementation for the keeper
func (k *Keeper) SetKMS(kms KMSInterface) {
	k.mu.Lock()
	defer k.mu.Unlock()
	// In production, store KMS reference
	// For now, we'll use the built-in encryption
}

// ShouldRotateKeys checks if keys should be rotated
func (k *Keeper) ShouldRotateKeys() bool {
	securityConfig := k.getSecurityConfig()

	if securityConfig.KeyRotationIntervalHours == 0 {
		return false // Rotation disabled
	}

	// Check if enough time has passed since initialization
	// In production, track last rotation time in state
	return true // Simplified for demonstration
}

// RotateEncryptionKeys performs key rotation
func (k *Keeper) RotateEncryptionKeys() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Generate new key
	newKeyID := k.generateKeyID()
	newKey := make([]byte, 32)
	if _, err := rand.Read(newKey); err != nil {
		return fmt.Errorf("failed to generate new key: %w", err)
	}

	// Store new key
	k.encryptionKeys[newKeyID] = newKey

	// Mark old key as deprecated but keep for decryption
	if k.currentEncryptionKeyID != "" {
		// Old key remains in map for decryption of existing data
		k.auditAction("key_rotated", "system", k.currentEncryptionKeyID,
			fmt.Sprintf("Rotated to new key %s", newKeyID),
			true, map[string]string{
				"old_key_id": k.currentEncryptionKeyID,
				"new_key_id": newKeyID,
			})
	}

	// Set new key as current
	k.currentEncryptionKeyID = newKeyID

	// Clean up old keys if we have too many
	k.cleanupOldKeys()

	return nil
}

// cleanupOldKeys removes keys that are no longer needed
func (k *Keeper) cleanupOldKeys() {
	securityConfig := k.getSecurityConfig()

	if len(k.encryptionKeys) <= securityConfig.MinKeysToRetain {
		return // Keep minimum number of keys
	}

	// In production, track key metadata and remove keys that:
	// 1. Are no longer referenced by any pre-validations
	// 2. Are past the grace period
	// 3. Exceed the minimum retention count

	// For now, keep all keys for backward compatibility
}

// GetKeyRotationStatus returns the status of key rotation
func (k *Keeper) GetKeyRotationStatus() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	status := map[string]interface{}{
		"current_key_id": k.currentEncryptionKeyID,
		"total_keys":     len(k.encryptionKeys),
		"keys_in_use":    k.countKeysInUse(),
		"should_rotate":  k.ShouldRotateKeys(),
	}

	return status
}

// countKeysInUse counts how many keys are currently being used
func (k *Keeper) countKeysInUse() int {
	keysInUse := make(map[string]bool)

	for _, tx := range k.preValidatedTxs {
		keysInUse[tx.EncryptionKeyId] = true
	}

	return len(keysInUse)
}

// ReEncryptWithNewKey re-encrypts data with a new key
func (k *Keeper) ReEncryptWithNewKey(txID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	tx, ok := k.preValidatedTxs[txID]
	if !ok {
		return types.ErrPreValidationNotFound
	}

	// Decrypt with old key
	oldKeyID := tx.EncryptionKeyId
	oldKey, ok := k.encryptionKeys[oldKeyID]
	if !ok {
		return fmt.Errorf("old encryption key not found: %s", oldKeyID)
	}

	plaintext, err := k.decryptWithKey(tx.EncryptedData, oldKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt with old key: %w", err)
	}

	// Encrypt with new key
	newKey, ok := k.encryptionKeys[k.currentEncryptionKeyID]
	if !ok {
		return fmt.Errorf("new encryption key not found")
	}

	newEncryptedData, err := k.encryptWithKey(plaintext, newKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt with new key: %w", err)
	}

	// Update transaction
	tx.EncryptedData = newEncryptedData
	tx.EncryptionKeyId = k.currentEncryptionKeyID

	k.auditAction("data_reencrypted", "system", txID,
		fmt.Sprintf("Re-encrypted from key %s to %s", oldKeyID, k.currentEncryptionKeyID),
		true, nil)

	return nil
}

// decryptWithKey decrypts data with a specific key
func (k *Keeper) decryptWithKey(encryptedData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, types.ErrDecryptionFailed
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, types.ErrDecryptionFailed
	}

	return plaintext, nil
}

// encryptWithKey encrypts data with a specific key
func (k *Keeper) encryptWithKey(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// BatchReEncrypt re-encrypts all pre-validated transactions with the current key
func (k *Keeper) BatchReEncrypt() (uint64, error) {
	k.mu.Lock()
	txIDs := make([]string, 0, len(k.preValidatedTxs))
	for txID := range k.preValidatedTxs {
		txIDs = append(txIDs, txID)
	}
	k.mu.Unlock()

	successCount := uint64(0)
	errorCount := uint64(0)

	for _, txID := range txIDs {
		if err := k.ReEncryptWithNewKey(txID); err != nil {
			errorCount++
			k.auditAction("reencryption_failed", "system", txID,
				fmt.Sprintf("Failed to re-encrypt: %v", err),
				false, nil)
		} else {
			successCount++
		}
	}

	k.auditAction("batch_reencryption", "system", "",
		fmt.Sprintf("Re-encrypted %d transactions (%d failed)", successCount, errorCount),
		errorCount == 0,
		map[string]string{
			"success_count": fmt.Sprintf("%d", successCount),
			"error_count":   fmt.Sprintf("%d", errorCount),
		})

	return successCount, nil
}

// ScheduleKeyRotation sets up automatic key rotation
func (k *Keeper) ScheduleKeyRotation(intervalHours uint32) {
	// In production, this would set up a periodic task
	// For now, it's called manually or by the scheduler
	k.auditAction("key_rotation_scheduled", "system", "",
		fmt.Sprintf("Scheduled key rotation every %d hours", intervalHours),
		true, map[string]string{
			"interval_hours": fmt.Sprintf("%d", intervalHours),
		})
}

// GetKeyAge returns the age of the current encryption key
func (k *Keeper) GetKeyAge() time.Duration {
	// In production, track key creation time in metadata
	// For now, return a placeholder
	return 0
}

// ForceKeyRotation immediately rotates keys (admin function)
func (k *Keeper) ForceKeyRotation() error {
	k.auditAction("key_rotation_forced", "admin", "",
		"Forced key rotation initiated",
		true, nil)

	if err := k.RotateEncryptionKeys(); err != nil {
		return err
	}

	// Optionally re-encrypt all existing data
	// This can be done asynchronously for large datasets
	// _, _ = k.BatchReEncrypt()

	return nil
}

// GetKeyUsageStatistics returns statistics about key usage
func (k *Keeper) GetKeyUsageStatistics() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	keyUsage := make(map[string]int)
	for _, tx := range k.preValidatedTxs {
		keyUsage[tx.EncryptionKeyId]++
	}

	return map[string]interface{}{
		"total_keys":     len(k.encryptionKeys),
		"current_key_id": k.currentEncryptionKeyID,
		"key_usage":      keyUsage,
		"keys_in_use":    len(keyUsage),
	}
}

// RevokeKey revokes a specific encryption key
func (k *Keeper) RevokeKey(keyID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Check if key exists
	if _, ok := k.encryptionKeys[keyID]; !ok {
		return fmt.Errorf("key not found: %s", keyID)
	}

	// Don't allow revoking the current key
	if keyID == k.currentEncryptionKeyID {
		return fmt.Errorf("cannot revoke current encryption key")
	}

	// Check if any pre-validations still use this key
	for _, tx := range k.preValidatedTxs {
		if tx.EncryptionKeyId == keyID {
			return fmt.Errorf("key still in use by pre-validations")
		}
	}

	// Remove the key
	delete(k.encryptionKeys, keyID)

	k.auditAction("key_revoked", "admin", keyID,
		"Encryption key revoked",
		true, nil)

	return nil
}

// ExportKeyMetadata exports metadata about all keys (for auditing)
func (k *Keeper) ExportKeyMetadata() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	metadata := make(map[string]interface{})

	for keyID := range k.encryptionKeys {
		usage := 0
		for _, tx := range k.preValidatedTxs {
			if tx.EncryptionKeyId == keyID {
				usage++
			}
		}

		metadata[keyID] = map[string]interface{}{
			"is_current":  keyID == k.currentEncryptionKeyID,
			"usage_count": usage,
		}
	}

	return metadata
}

// ValidateKeyIntegrity checks the integrity of encryption keys
func (k *Keeper) ValidateKeyIntegrity() error {
	k.mu.RLock()
	defer k.mu.RUnlock()

	// Check current key exists
	if k.currentEncryptionKeyID == "" {
		return fmt.Errorf("no current encryption key set")
	}

	if _, ok := k.encryptionKeys[k.currentEncryptionKeyID]; !ok {
		return fmt.Errorf("current encryption key not found in key store")
	}

	// Check all referenced keys exist
	missingKeys := make(map[string]bool)
	for _, tx := range k.preValidatedTxs {
		if _, ok := k.encryptionKeys[tx.EncryptionKeyId]; !ok {
			missingKeys[tx.EncryptionKeyId] = true
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("missing encryption keys: %v", getMapKeys(missingKeys))
	}

	return nil
}

// Helper function to get map keys
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
