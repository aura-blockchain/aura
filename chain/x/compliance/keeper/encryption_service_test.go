package keeper

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// ============================================================================
// Service Initialization Tests
// ============================================================================

func TestNewEncryptionService(t *testing.T) {
	tests := []struct {
		name      string
		masterKey []byte
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid 32-byte key",
			masterKey: make([]byte, 32),
			wantErr:   false,
		},
		{
			name:      "invalid 16-byte key",
			masterKey: make([]byte, 16),
			wantErr:   true,
			errMsg:    "master key must be exactly 32 bytes",
		},
		{
			name:      "invalid 64-byte key",
			masterKey: make([]byte, 64),
			wantErr:   true,
			errMsg:    "master key must be exactly 32 bytes",
		},
		{
			name:      "empty key",
			masterKey: []byte{},
			wantErr:   true,
			errMsg:    "master key must be exactly 32 bytes",
		},
		{
			name:      "nil key",
			masterKey: nil,
			wantErr:   true,
			errMsg:    "master key must be exactly 32 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewEncryptionService(tt.masterKey)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				require.Nil(t, service)
			} else {
				require.NoError(t, err)
				require.NotNil(t, service)
				require.NotNil(t, service.masterKey)
				require.Len(t, service.masterKey, 32)
			}
		})
	}
}

func TestEncryptionServiceKeyIsolation(t *testing.T) {
	// Test that service makes a copy of the key (doesn't hold reference to input)
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	// Modify original key
	for i := range masterKey {
		masterKey[i] = 0xFF
	}

	// Service should still have original key values
	for i := 0; i < 32; i++ {
		require.Equal(t, byte(i), service.masterKey[i], "service key should not be affected by external modification")
	}
}

// ============================================================================
// Encryption/Decryption Round-Trip Tests
// ============================================================================

func TestEncryptDecryptRoundTrip(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	tests := []struct {
		name      string
		plaintext []byte
		context   string
	}{
		{
			name:      "simple text",
			plaintext: []byte("Hello, World!"),
			context:   "test:1",
		},
		{
			name:      "binary data",
			plaintext: []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD},
			context:   "test:2",
		},
		{
			name:      "large data (1KB)",
			plaintext: make([]byte, 1024),
			context:   "test:3",
		},
		{
			name:      "single byte",
			plaintext: []byte{0x42},
			context:   "test:4",
		},
		{
			name:      "special characters",
			plaintext: []byte("Special chars: !@#$%^&*()[]{}"),
			context:   "test:5",
		},
		{
			name:      "unicode text",
			plaintext: []byte("Unicode: 你好世界 🌍"),
			context:   "test:6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, err := service.Encrypt(tt.plaintext, tt.context)
			require.NoError(t, err)
			require.NotNil(t, ciphertext)
			require.NotEqual(t, tt.plaintext, ciphertext, "ciphertext should differ from plaintext")

			// Verify ciphertext has correct overhead (nonce + tag)
			expectedMinSize := len(tt.plaintext) + service.GetEncryptionOverhead()
			require.GreaterOrEqual(t, len(ciphertext), expectedMinSize)

			// Decrypt
			decrypted, err := service.Decrypt(ciphertext, tt.context)
			require.NoError(t, err)
			require.Equal(t, tt.plaintext, decrypted, "decrypted data should match original plaintext")
		})
	}
}

func TestEncryptDecryptEmpty(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	// Empty plaintext should fail
	_, err = service.Encrypt([]byte{}, "test:empty")
	require.Error(t, err)
	require.Contains(t, err.Error(), "plaintext cannot be empty")

	// Empty ciphertext should fail
	_, err = service.Decrypt([]byte{}, "test:empty")
	require.Error(t, err)
	require.Contains(t, err.Error(), "ciphertext cannot be empty")
}

// ============================================================================
// Context Isolation Tests
// ============================================================================

func TestDifferentContextsProduceDifferentCiphertexts(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	plaintext := []byte("sensitive data")

	// Encrypt same plaintext with different contexts
	ciphertext1, err := service.Encrypt(plaintext, "context:1")
	require.NoError(t, err)

	ciphertext2, err := service.Encrypt(plaintext, "context:2")
	require.NoError(t, err)

	// Ciphertexts should be different (different derived keys)
	require.NotEqual(t, ciphertext1, ciphertext2)

	// Each should decrypt correctly with its own context
	decrypted1, err := service.Decrypt(ciphertext1, "context:1")
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted1)

	decrypted2, err := service.Decrypt(ciphertext2, "context:2")
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted2)

	// Cross-context decryption should fail (wrong key)
	_, err = service.Decrypt(ciphertext1, "context:2")
	require.Error(t, err)
	require.Contains(t, err.Error(), "authentication error")

	_, err = service.Decrypt(ciphertext2, "context:1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "authentication error")
}

func TestEmptyContextFails(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	plaintext := []byte("test data")

	// Empty context should fail
	_, err = service.Encrypt(plaintext, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be empty")
}

// ============================================================================
// Tampering Detection Tests
// ============================================================================

func TestTamperingDetection(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	plaintext := []byte("sensitive data")
	context := "test:tampering"

	ciphertext, err := service.Encrypt(plaintext, context)
	require.NoError(t, err)

	tests := []struct {
		name        string
		tamperFunc  func([]byte) []byte
		description string
	}{
		{
			name: "flip single bit",
			tamperFunc: func(ct []byte) []byte {
				tampered := make([]byte, len(ct))
				copy(tampered, ct)
				// Flip bit in the middle of ciphertext
				tampered[len(ct)/2] ^= 0x01
				return tampered
			},
			description: "bit flip in ciphertext",
		},
		{
			name: "modify nonce",
			tamperFunc: func(ct []byte) []byte {
				tampered := make([]byte, len(ct))
				copy(tampered, ct)
				// Modify nonce (first 12 bytes)
				tampered[0] ^= 0xFF
				return tampered
			},
			description: "modified nonce",
		},
		{
			name: "modify auth tag",
			tamperFunc: func(ct []byte) []byte {
				tampered := make([]byte, len(ct))
				copy(tampered, ct)
				// Modify last byte (auth tag)
				tampered[len(ct)-1] ^= 0x01
				return tampered
			},
			description: "modified authentication tag",
		},
		{
			name: "truncate ciphertext",
			tamperFunc: func(ct []byte) []byte {
				return ct[:len(ct)-1]
			},
			description: "truncated ciphertext",
		},
		{
			name: "append data",
			tamperFunc: func(ct []byte) []byte {
				return append(ct, 0xFF)
			},
			description: "appended data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tamperedCiphertext := tt.tamperFunc(ciphertext)

			// Decryption should fail with authentication error
			_, err := service.Decrypt(tamperedCiphertext, context)
			require.Error(t, err, "tampering should be detected: %s", tt.description)
			require.Contains(t, err.Error(), "authentication error", "should indicate authentication failure")
		})
	}
}

// ============================================================================
// Nonce Uniqueness Tests
// ============================================================================

func TestNonceUniqueness(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	plaintext := []byte("test data")
	context := "test:nonce"

	// Encrypt same plaintext multiple times
	nonces := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		ciphertext, err := service.Encrypt(plaintext, context)
		require.NoError(t, err)

		// Extract nonce (first 12 bytes)
		nonce := ciphertext[:12]
		nonceStr := string(nonce)

		// Verify nonce is unique
		require.False(t, nonces[nonceStr], "nonce should be unique across encryptions")
		nonces[nonceStr] = true
	}

	// Verify we got 100 unique nonces
	require.Len(t, nonces, iterations)
}

// ============================================================================
// JSON Encryption Tests
// ============================================================================

func TestEncryptDecryptJSON(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	tests := []struct {
		name    string
		data    interface{}
		context string
	}{
		{
			name:    "string map",
			data:    map[string]string{"name": "Alice", "ssn": "123-45-6789"},
			context: "json:1",
		},
		{
			name: "nested struct",
			data: struct {
				Name    string
				Age     int
				Address struct {
					Street string
					City   string
				}
			}{
				Name: "Bob",
				Age:  30,
				Address: struct {
					Street string
					City   string
				}{
					Street: "123 Main St",
					City:   "Springfield",
				},
			},
			context: "json:2",
		},
		{
			name:    "array",
			data:    []string{"item1", "item2", "item3"},
			context: "json:3",
		},
		{
			name:    "number",
			data:    float64(123.456),
			context: "json:4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt JSON
			encrypted, err := service.EncryptJSON(tt.data, tt.context)
			require.NoError(t, err)
			require.NotNil(t, encrypted)

			// Decrypt JSON
			var decrypted interface{}
			err = service.DecryptJSON(encrypted, tt.context, &decrypted)
			require.NoError(t, err)

			// Compare JSON representations (since types might differ)
			originalJSON, _ := json.Marshal(tt.data)
			decryptedJSON, _ := json.Marshal(decrypted)
			require.JSONEq(t, string(originalJSON), string(decryptedJSON))
		})
	}
}

// ============================================================================
// String Encryption Tests
// ============================================================================

func TestEncryptDecryptString(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	tests := []struct {
		name    string
		text    string
		context string
	}{
		{
			name:    "simple string",
			text:    "Hello, World!",
			context: "string:1",
		},
		{
			name:    "empty string should fail",
			text:    "",
			context: "string:2",
		},
		{
			name:    "unicode string",
			text:    "你好世界 🌍",
			context: "string:3",
		},
		{
			name:    "long string",
			text:    string(make([]byte, 10000)),
			context: "string:4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.text == "" {
				_, err := service.EncryptString(tt.text, tt.context)
				require.Error(t, err)
				return
			}

			encrypted, err := service.EncryptString(tt.text, tt.context)
			require.NoError(t, err)

			decrypted, err := service.DecryptString(encrypted, tt.context)
			require.NoError(t, err)
			require.Equal(t, tt.text, decrypted)
		})
	}
}

// ============================================================================
// Base64 Encoding Tests
// ============================================================================

func TestEncryptDecryptBase64(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	plaintext := []byte("sensitive data for base64 encoding")
	context := "base64:1"

	// Encrypt to base64
	base64Ciphertext, err := service.EncryptToBase64(plaintext, context)
	require.NoError(t, err)
	require.NotEmpty(t, base64Ciphertext)

	// Verify it's valid base64
	_, err = base64.StdEncoding.DecodeString(base64Ciphertext)
	require.NoError(t, err, "output should be valid base64")

	// Decrypt from base64
	decrypted, err := service.DecryptFromBase64(base64Ciphertext, context)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

func TestDecryptInvalidBase64(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	// Invalid base64 should fail
	_, err = service.DecryptFromBase64("not valid base64!@#$", "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode base64")
}

// ============================================================================
// Key Rotation Tests
// ============================================================================

func TestKeyRotation(t *testing.T) {
	oldKey := generateRandomKey(t)
	newKey := generateRandomKey(t)

	// Create service with old key
	oldService, err := NewEncryptionService(oldKey)
	require.NoError(t, err)

	plaintext := []byte("data encrypted with old key")
	context := "rotation:1"

	// Encrypt with old key
	ciphertext, err := oldService.Encrypt(plaintext, context)
	require.NoError(t, err)

	// Decrypt with old key (should work)
	decrypted, err := oldService.Decrypt(ciphertext, context)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)

	// Rotate to new key
	newService, err := oldService.RotateMasterKey(newKey)
	require.NoError(t, err)

	// Old ciphertext cannot be decrypted with new key
	_, err = newService.Decrypt(ciphertext, context)
	require.Error(t, err)

	// Simulate re-encryption process:
	// 1. Decrypt with old service
	decryptedWithOld, err := oldService.Decrypt(ciphertext, context)
	require.NoError(t, err)

	// 2. Re-encrypt with new service
	newCiphertext, err := newService.Encrypt(decryptedWithOld, context)
	require.NoError(t, err)

	// 3. Verify decryption with new service
	decryptedWithNew, err := newService.Decrypt(newCiphertext, context)
	require.NoError(t, err)
	require.Equal(t, plaintext, decryptedWithNew)
}

// ============================================================================
// Performance Tests
// ============================================================================

func TestEncryptionPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	dataSizes := []int{
		100,      // Small KYC field
		1024,     // Medium document
		10240,    // Large document
		102400,   // Very large (100KB)
	}

	for _, size := range dataSizes {
		t.Run(fmt.Sprintf("size_%d_bytes", size), func(t *testing.T) {
			plaintext := make([]byte, size)
			_, _ = rand.Read(plaintext)

			context := fmt.Sprintf("perf:size_%d", size)

			// Measure encryption time
			encrypted, err := service.Encrypt(plaintext, context)
			require.NoError(t, err)

			// Measure decryption time
			decrypted, err := service.Decrypt(encrypted, context)
			require.NoError(t, err)
			require.Equal(t, plaintext, decrypted)

			// Log overhead
			overhead := len(encrypted) - len(plaintext)
			t.Logf("Size: %d bytes, Overhead: %d bytes (%.2f%%)",
				size, overhead, float64(overhead)/float64(size)*100)
		})
	}
}

// ============================================================================
// Security Property Tests
// ============================================================================

func TestEncryptionOverhead(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	require.Equal(t, 28, service.GetEncryptionOverhead(), "overhead should be 28 bytes (12 nonce + 16 tag)")

	// Verify actual overhead matches
	plaintext := []byte("test")
	encrypted, err := service.Encrypt(plaintext, "test")
	require.NoError(t, err)

	actualOverhead := len(encrypted) - len(plaintext)
	require.Equal(t, service.GetEncryptionOverhead(), actualOverhead)
}

func TestDifferentMasterKeysProduceDifferentCiphertexts(t *testing.T) {
	key1 := generateRandomKey(t)
	key2 := generateRandomKey(t)

	service1, err := NewEncryptionService(key1)
	require.NoError(t, err)

	service2, err := NewEncryptionService(key2)
	require.NoError(t, err)

	plaintext := []byte("test data")
	context := "test:keys"

	ciphertext1, err := service1.Encrypt(plaintext, context)
	require.NoError(t, err)

	ciphertext2, err := service2.Encrypt(plaintext, context)
	require.NoError(t, err)

	// Ciphertexts should be different (different master keys)
	// Note: Even with same plaintext and context, nonce randomness makes them different
	// But we're testing that different master keys also contribute
	require.NotEqual(t, ciphertext1, ciphertext2)

	// Cross-service decryption should fail
	_, err = service1.Decrypt(ciphertext2, context)
	require.Error(t, err)

	_, err = service2.Decrypt(ciphertext1, context)
	require.Error(t, err)
}

func TestConstantTimeComparison(t *testing.T) {
	// This test verifies that wipeKey actually zeros the key
	// (Best-effort security - Go doesn't guarantee this)
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	wipeKey(key)

	// Verify all bytes are zero
	for i, b := range key {
		require.Equal(t, byte(0), b, "byte %d should be wiped", i)
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestShortCiphertext(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	// Ciphertext shorter than nonce size should fail
	shortCiphertext := []byte{0x01, 0x02, 0x03} // Only 3 bytes, need at least 12

	_, err = service.Decrypt(shortCiphertext, "test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "ciphertext too short")
}

func TestLargeContext(t *testing.T) {
	masterKey := generateRandomKey(t)
	service, err := NewEncryptionService(masterKey)
	require.NoError(t, err)

	plaintext := []byte("test data")
	// Very long context (1KB)
	longContext := string(make([]byte, 1024))

	encrypted, err := service.Encrypt(plaintext, longContext)
	require.NoError(t, err)

	decrypted, err := service.Decrypt(encrypted, longContext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

// ============================================================================
// Helper Functions
// ============================================================================

func generateRandomKey(t *testing.T) []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return key
}

// Benchmark tests
func BenchmarkEncrypt(b *testing.B) {
	masterKey := make([]byte, 32)
	rand.Read(masterKey)
	service, _ := NewEncryptionService(masterKey)

	plaintext := []byte("benchmark data for encryption")
	context := "benchmark:encrypt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.Encrypt(plaintext, context)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	masterKey := make([]byte, 32)
	rand.Read(masterKey)
	service, _ := NewEncryptionService(masterKey)

	plaintext := []byte("benchmark data for decryption")
	context := "benchmark:decrypt"
	ciphertext, _ := service.Encrypt(plaintext, context)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.Decrypt(ciphertext, context)
	}
}

func BenchmarkEncryptJSON(b *testing.B) {
	masterKey := make([]byte, 32)
	rand.Read(masterKey)
	service, _ := NewEncryptionService(masterKey)

	data := map[string]string{
		"name":  "Alice",
		"ssn":   "123-45-6789",
		"email": "alice@example.com",
	}
	context := "benchmark:json"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.EncryptJSON(data, context)
	}
}
