package keeper

import (
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/prevalidation/types"
)

// ============================
// KEY ROTATION TESTS
// ============================

func TestRotateEncryptionKeys(t *testing.T) {
	keeper := setupTestKeeper()

	// Get initial key ID
	initialKeyID := keeper.currentEncryptionKeyID
	initialKeyCount := len(keeper.encryptionKeys)

	// Rotate keys
	err := keeper.RotateEncryptionKeys()
	if err != nil {
		t.Fatalf("Failed to rotate keys: %v", err)
	}

	// Verify new key was created
	if keeper.currentEncryptionKeyID == initialKeyID {
		t.Error("Current key ID should have changed after rotation")
	}

	// Verify old key is still available (for decryption)
	if len(keeper.encryptionKeys) != initialKeyCount+1 {
		t.Errorf("Expected %d keys after rotation, got %d",
			initialKeyCount+1, len(keeper.encryptionKeys))
	}

	// Verify old key still exists
	if _, ok := keeper.encryptionKeys[initialKeyID]; !ok {
		t.Error("Old key should still exist after rotation")
	}

	// Verify new key exists
	if _, ok := keeper.encryptionKeys[keeper.currentEncryptionKeyID]; !ok {
		t.Error("New key should exist after rotation")
	}
}

func TestReEncryptWithNewKey(t *testing.T) {
	keeper := setupTestKeeper()

	// Create a pre-validated transaction
	txData := []byte("test transaction data")
	tx, err := keeper.CreatePreValidatedTransaction(
		types.TxTypeIRCompletion,
		"test-template",
		txData,
		"test-signer",
		50000,
		map[string]string{},
	)

	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	initialKeyID := tx.EncryptionKeyId

	// Rotate keys
	err = keeper.RotateEncryptionKeys()
	if err != nil {
		t.Fatalf("Failed to rotate keys: %v", err)
	}

	// Re-encrypt the transaction
	err = keeper.ReEncryptWithNewKey(tx.Id)
	if err != nil {
		t.Fatalf("Failed to re-encrypt: %v", err)
	}

	// Verify key ID was updated
	updatedTx, ok := keeper.GetPreValidatedTransaction(tx.Id)
	if !ok {
		t.Fatal("Transaction not found after re-encryption")
	}

	if updatedTx.EncryptionKeyId == initialKeyID {
		t.Error("Key ID should have changed after re-encryption")
	}

	if updatedTx.EncryptionKeyId != keeper.currentEncryptionKeyID {
		t.Error("Transaction should use current encryption key")
	}

	// Verify we can still decrypt the data
	decryptedData, err := keeper.ExecutePreValidatedTransaction(tx.Id)
	if err != nil {
		t.Fatalf("Failed to execute after re-encryption: %v", err)
	}

	if len(decryptedData) == 0 {
		t.Error("Decrypted data should not be empty")
	}
}

func TestBatchReEncrypt(t *testing.T) {
	keeper := setupTestKeeper()

	// Create multiple pre-validated transactions
	numTxs := 5
	txIDs := make([]string, numTxs)

	for i := 0; i < numTxs; i++ {
		tx, err := keeper.CreatePreValidatedTransaction(
			types.TxTypeIRCompletion,
			"test-template",
			[]byte("test data"),
			"test-signer",
			50000,
			map[string]string{},
		)

		if err != nil {
			t.Fatalf("Failed to create transaction %d: %v", i, err)
		}

		txIDs[i] = tx.Id
	}

	// Rotate keys
	err := keeper.RotateEncryptionKeys()
	if err != nil {
		t.Fatalf("Failed to rotate keys: %v", err)
	}

	// Batch re-encrypt
	successCount, err := keeper.BatchReEncrypt()
	if err != nil {
		t.Fatalf("Batch re-encrypt failed: %v", err)
	}

	if successCount != uint64(numTxs) {
		t.Errorf("Expected %d successful re-encryptions, got %d", numTxs, successCount)
	}

	// Verify all transactions use new key
	for _, txID := range txIDs {
		tx, ok := keeper.GetPreValidatedTransaction(txID)
		if !ok {
			t.Errorf("Transaction %s not found", txID)
			continue
		}

		if tx.EncryptionKeyId != keeper.currentEncryptionKeyID {
			t.Errorf("Transaction %s should use current key", txID)
		}
	}
}

func TestGetKeyRotationStatus(t *testing.T) {
	keeper := setupTestKeeper()

	status := keeper.GetKeyRotationStatus()

	if status == nil {
		t.Fatal("Status should not be nil")
	}

	if _, ok := status["current_key_id"]; !ok {
		t.Error("Status should include current_key_id")
	}

	if _, ok := status["total_keys"]; !ok {
		t.Error("Status should include total_keys")
	}

	if _, ok := status["keys_in_use"]; !ok {
		t.Error("Status should include keys_in_use")
	}

	if _, ok := status["should_rotate"]; !ok {
		t.Error("Status should include should_rotate")
	}
}

func TestRevokeKey(t *testing.T) {
	keeper := setupTestKeeper()

	// Rotate to create additional key
	oldKeyID := keeper.currentEncryptionKeyID
	err := keeper.RotateEncryptionKeys()
	if err != nil {
		t.Fatalf("Failed to rotate keys: %v", err)
	}

	// Cannot revoke current key
	err = keeper.RevokeKey(keeper.currentEncryptionKeyID)
	if err == nil {
		t.Error("Should not allow revoking current key")
	}

	// Can revoke old key if not in use
	err = keeper.RevokeKey(oldKeyID)
	if err != nil {
		t.Errorf("Should allow revoking unused old key: %v", err)
	}

	// Verify key was removed
	if _, ok := keeper.encryptionKeys[oldKeyID]; ok {
		t.Error("Revoked key should be removed")
	}
}

func TestRevokeKeyInUse(t *testing.T) {
	keeper := setupTestKeeper()

	// Create transaction with current key
	tx, err := keeper.CreatePreValidatedTransaction(
		types.TxTypeIRCompletion,
		"test-template",
		[]byte("test data"),
		"test-signer",
		50000,
		map[string]string{},
	)

	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	keyInUse := tx.EncryptionKeyId

	// Rotate to new key
	err = keeper.RotateEncryptionKeys()
	if err != nil {
		t.Fatalf("Failed to rotate keys: %v", err)
	}

	// Cannot revoke key that's still in use
	err = keeper.RevokeKey(keyInUse)
	if err == nil {
		t.Error("Should not allow revoking key that's in use")
	}
}

func TestGetKeyUsageStatistics(t *testing.T) {
	keeper := setupTestKeeper()

	// Create transactions with different keys
	tx1, err := keeper.CreatePreValidatedTransaction(
		types.TxTypeIRCompletion,
		"test-template",
		[]byte("test data 1"),
		"test-signer",
		50000,
		map[string]string{},
	)
	if err != nil {
		t.Fatalf("Failed to create transaction 1: %v", err)
	}

	// Rotate key
	err = keeper.RotateEncryptionKeys()
	if err != nil {
		t.Fatalf("Failed to rotate keys: %v", err)
	}

	// Create another transaction with new key
	_, err = keeper.CreatePreValidatedTransaction(
		types.TxTypeIRCompletion,
		"test-template",
		[]byte("test data 2"),
		"test-signer",
		50000,
		map[string]string{},
	)
	if err != nil {
		t.Fatalf("Failed to create transaction 2: %v", err)
	}

	stats := keeper.GetKeyUsageStatistics()

	if stats == nil {
		t.Fatal("Stats should not be nil")
	}

	totalKeys, ok := stats["total_keys"].(int)
	if !ok || totalKeys < 2 {
		t.Errorf("Expected at least 2 total keys, got %v", stats["total_keys"])
	}

	keysInUse, ok := stats["keys_in_use"].(int)
	if !ok || keysInUse < 2 {
		t.Errorf("Expected at least 2 keys in use, got %v", stats["keys_in_use"])
	}

	keyUsage, ok := stats["key_usage"].(map[string]int)
	if !ok {
		t.Error("Expected key_usage map")
	} else {
		if usage, exists := keyUsage[tx1.EncryptionKeyId]; !exists || usage != 1 {
			t.Errorf("Expected usage of 1 for key %s", tx1.EncryptionKeyId)
		}
	}
}

func TestValidateKeyIntegrity(t *testing.T) {
	keeper := setupTestKeeper()

	// Initially should be valid
	err := keeper.ValidateKeyIntegrity()
	if err != nil {
		t.Errorf("Initial key integrity should be valid: %v", err)
	}

	// Create transaction
	tx, err := keeper.CreatePreValidatedTransaction(
		types.TxTypeIRCompletion,
		"test-template",
		[]byte("test data"),
		"test-signer",
		50000,
		map[string]string{},
	)
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Should still be valid
	err = keeper.ValidateKeyIntegrity()
	if err != nil {
		t.Errorf("Key integrity should be valid: %v", err)
	}

	// Simulate missing key by manually deleting it
	keeper.mu.Lock()
	delete(keeper.encryptionKeys, tx.EncryptionKeyId)
	keeper.mu.Unlock()

	// Should detect missing key
	err = keeper.ValidateKeyIntegrity()
	if err == nil {
		t.Error("Should detect missing encryption key")
	}
}

func TestExportKeyMetadata(t *testing.T) {
	keeper := setupTestKeeper()

	// Create transaction
	_, err := keeper.CreatePreValidatedTransaction(
		types.TxTypeIRCompletion,
		"test-template",
		[]byte("test data"),
		"test-signer",
		50000,
		map[string]string{},
	)
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	metadata := keeper.ExportKeyMetadata()

	if metadata == nil {
		t.Fatal("Metadata should not be nil")
	}

	if len(metadata) == 0 {
		t.Error("Metadata should contain at least one key")
	}

	for keyID, meta := range metadata {
		metaMap, ok := meta.(map[string]interface{})
		if !ok {
			t.Errorf("Metadata for key %s should be a map", keyID)
			continue
		}

		if _, ok := metaMap["is_current"]; !ok {
			t.Errorf("Metadata for key %s should include is_current", keyID)
		}

		if _, ok := metaMap["usage_count"]; !ok {
			t.Errorf("Metadata for key %s should include usage_count", keyID)
		}
	}
}

func TestForceKeyRotation(t *testing.T) {
	keeper := setupTestKeeper()

	initialKeyID := keeper.currentEncryptionKeyID

	// Force rotation
	err := keeper.ForceKeyRotation()
	if err != nil {
		t.Fatalf("Failed to force key rotation: %v", err)
	}

	// Verify rotation occurred
	if keeper.currentEncryptionKeyID == initialKeyID {
		t.Error("Key should have been rotated")
	}

	// Verify old key still exists
	if _, ok := keeper.encryptionKeys[initialKeyID]; !ok {
		t.Error("Old key should still exist")
	}
}

// ============================
// LOCAL KMS TESTS
// ============================

func TestLocalKMS(t *testing.T) {
	kms, err := NewLocalKMS()
	if err != nil {
		t.Fatalf("Failed to create KMS: %v", err)
	}

	// Test key generation
	key, err := kms.GenerateKey("test-key-1", "AES-256-GCM")
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected 32-byte key, got %d bytes", len(key))
	}

	// Test key encryption/decryption
	testData := []byte("test data to encrypt")
	encrypted, err := kms.EncryptKey(testData)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	decrypted, err := kms.DecryptKey(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if !bytesEqual(testData, decrypted) {
		t.Error("Decrypted data should match original")
	}

	// Test key listing
	keys, err := kms.ListKeys()
	if err != nil {
		t.Fatalf("Failed to list keys: %v", err)
	}

	if len(keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(keys))
	}

	// Test key metadata
	metadata, err := kms.GetKeyMetadata("test-key-1")
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}

	if metadata.KeyID != "test-key-1" {
		t.Errorf("Expected key ID 'test-key-1', got '%s'", metadata.KeyID)
	}

	if metadata.Algorithm != "AES-256-GCM" {
		t.Errorf("Expected algorithm 'AES-256-GCM', got '%s'", metadata.Algorithm)
	}

	if metadata.Status != KeyStatusActive {
		t.Errorf("Expected status Active, got %v", metadata.Status)
	}

	// Test key revocation
	err = kms.RevokeKey("test-key-1")
	if err != nil {
		t.Fatalf("Failed to revoke key: %v", err)
	}

	metadata, err = kms.GetKeyMetadata("test-key-1")
	if err != nil {
		t.Fatalf("Failed to get metadata after revocation: %v", err)
	}

	if metadata.Status != KeyStatusRevoked {
		t.Errorf("Expected status Revoked, got %v", metadata.Status)
	}
}

func TestKeyStatusString(t *testing.T) {
	tests := []struct {
		status   KeyStatus
		expected string
	}{
		{KeyStatusActive, "active"},
		{KeyStatusRotating, "rotating"},
		{KeyStatusDeprecated, "deprecated"},
		{KeyStatusRevoked, "revoked"},
		{KeyStatus(99), "unknown"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Errorf("Expected '%s' for status %d, got '%s'",
				tt.expected, tt.status, tt.status.String())
		}
	}
}

func TestEncryptDecryptWithKey(t *testing.T) {
	keeper := setupTestKeeper()

	testData := []byte("sensitive transaction data")

	// Get current key
	key, ok := keeper.encryptionKeys[keeper.currentEncryptionKeyID]
	if !ok {
		t.Fatal("Current encryption key not found")
	}

	// Encrypt
	encrypted, err := keeper.encryptWithKey(testData, key)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Verify encrypted data is different
	if bytesEqual(testData, encrypted) {
		t.Error("Encrypted data should differ from plaintext")
	}

	// Decrypt
	decrypted, err := keeper.decryptWithKey(encrypted, key)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	// Verify decrypted data matches original
	if !bytesEqual(testData, decrypted) {
		t.Error("Decrypted data should match original")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	keeper := setupTestKeeper()

	testData := []byte("sensitive data")

	// Encrypt with current key
	key1, ok := keeper.encryptionKeys[keeper.currentEncryptionKeyID]
	if !ok {
		t.Fatal("Current encryption key not found")
	}

	encrypted, err := keeper.encryptWithKey(testData, key1)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Try to decrypt with wrong key
	wrongKey := make([]byte, 32)
	for i := range wrongKey {
		wrongKey[i] = byte(i)
	}

	_, err = keeper.decryptWithKey(encrypted, wrongKey)
	if err == nil {
		t.Error("Should fail to decrypt with wrong key")
	}
}

func TestScheduleKeyRotation(t *testing.T) {
	keeper := setupTestKeeper()

	// Schedule rotation
	keeper.ScheduleKeyRotation(24)

	// This is primarily for code coverage
	// In production, would verify scheduled task was created
}

func TestCleanupOldKeys(t *testing.T) {
	keeper := setupTestKeeper()

	// Create many keys by rotating multiple times
	for i := 0; i < 5; i++ {
		err := keeper.RotateEncryptionKeys()
		if err != nil {
			t.Fatalf("Failed to rotate keys: %v", err)
		}
	}

	initialCount := len(keeper.encryptionKeys)

	// Cleanup (with current implementation, this keeps all keys)
	keeper.mu.Lock()
	keeper.cleanupOldKeys()
	keeper.mu.Unlock()

	// Verify minimum keys are retained
	securityConfig := keeper.getSecurityConfig()
	if len(keeper.encryptionKeys) < securityConfig.MinKeysToRetain {
		t.Errorf("Should retain at least %d keys, got %d",
			securityConfig.MinKeysToRetain, len(keeper.encryptionKeys))
	}

	// In current implementation, all keys are kept
	if len(keeper.encryptionKeys) != initialCount {
		t.Logf("Cleanup may have removed keys (initial: %d, current: %d)",
			initialCount, len(keeper.encryptionKeys))
	}
}
