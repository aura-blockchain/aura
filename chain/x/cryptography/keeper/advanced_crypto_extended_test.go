// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/cryptography/keeper"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// generateTestCertificate generates a self-signed X.509 certificate for testing
func generateTestCertificate() []byte {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
	}
	certDER, _ := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	return certDER
}

// ============================================================================
// TestPinCertificate - Tests for PinCertificate function (advanced_crypto.go:213)
// ============================================================================

func TestPinCertificate(t *testing.T) {
	k, ctx := setupKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	t.Run("empty domain returns error", func(t *testing.T) {
		certificate := []byte("valid-certificate-data")
		fingerprint := "sha256:abc123"

		err := k.PinCertificate(sdkCtx, "", certificate, fingerprint)
		require.Error(t, err)
		require.Contains(t, err.Error(), "domain cannot be empty")
	})

	t.Run("empty certificate returns error", func(t *testing.T) {
		domain := "example.com"
		fingerprint := "sha256:abc123"

		err := k.PinCertificate(sdkCtx, domain, nil, fingerprint)
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate cannot be empty")

		err = k.PinCertificate(sdkCtx, domain, []byte{}, fingerprint)
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate cannot be empty")
	})

	t.Run("valid domain and certificate creates pin successfully", func(t *testing.T) {
		domain := "secure.example.com"
		certificate := []byte("valid-certificate-content-for-testing")
		fingerprint := "sha256:def456"

		err := k.PinCertificate(sdkCtx, domain, certificate, fingerprint)
		require.NoError(t, err)

		// Verify the pin was created by retrieving it
		pin, err := k.GetCertificatePin(ctx, domain)
		require.NoError(t, err)
		require.NotNil(t, pin)
		require.Equal(t, domain, pin.Hostname)
		require.True(t, pin.Enabled)
		require.Equal(t, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_FULL_CERT, pin.PinType)
	})

	t.Run("pin can be retrieved via GetCertificatePin", func(t *testing.T) {
		domain := "api.example.com"
		certificate := generateTestCertificate()
		fingerprint := "sha256:ghi789"

		err := k.PinCertificate(sdkCtx, domain, certificate, fingerprint)
		require.NoError(t, err)

		// Retrieve and verify
		pin, err := k.GetCertificatePin(ctx, domain)
		require.NoError(t, err)
		require.NotNil(t, pin)
		require.Equal(t, domain, pin.Hostname)
		require.Len(t, pin.CertificateHashes, 1)
		require.Equal(t, certificate, pin.CertificateHashes[0])
		require.NotNil(t, pin.ExpiresAt)
		// Verify expiry is approximately 1 year in the future
		expectedExpiry := sdkCtx.BlockTime().AddDate(1, 0, 0)
		require.WithinDuration(t, expectedExpiry, *pin.ExpiresAt, time.Second)
	})

	t.Run("multiple pins for different domains", func(t *testing.T) {
		domains := []string{"alpha.example.com", "beta.example.com", "gamma.example.com"}
		for i, domain := range domains {
			cert := []byte("certificate-" + domain)
			err := k.PinCertificate(sdkCtx, domain, cert, "fp"+string(rune(i)))
			require.NoError(t, err)
		}

		// Verify all pins exist
		for _, domain := range domains {
			pin, err := k.GetCertificatePin(ctx, domain)
			require.NoError(t, err)
			require.Equal(t, domain, pin.Hostname)
		}
	})
}

// ============================================================================
// TestEncryptAESGCM - Tests for encryptAESGCM function (advanced_crypto.go:255)
// ============================================================================

func TestEncryptAESGCM(t *testing.T) {
	k, _ := setupKeeper(t)

	t.Run("valid 16-byte key encrypts plaintext", func(t *testing.T) {
		key := make([]byte, 16) // AES-128
		for i := range key {
			key[i] = byte(i)
		}
		plaintext := []byte("Hello, World! This is a test message for AES-GCM encryption.")

		ciphertext, err := k.EncryptAESGCMExported(plaintext, key)
		require.NoError(t, err)
		require.NotNil(t, ciphertext)
		// Ciphertext should be longer than plaintext (includes 12-byte nonce + 16-byte tag)
		require.Greater(t, len(ciphertext), len(plaintext))
		// Minimum overhead: 12 (nonce) + 16 (tag) = 28 bytes
		require.GreaterOrEqual(t, len(ciphertext), len(plaintext)+12)
	})

	t.Run("valid 24-byte key encrypts plaintext", func(t *testing.T) {
		key := make([]byte, 24) // AES-192
		for i := range key {
			key[i] = byte(i + 10)
		}
		plaintext := []byte("Testing AES-192 encryption mode.")

		ciphertext, err := k.EncryptAESGCMExported(plaintext, key)
		require.NoError(t, err)
		require.NotNil(t, ciphertext)
		require.Greater(t, len(ciphertext), len(plaintext))
	})

	t.Run("valid 32-byte key encrypts plaintext", func(t *testing.T) {
		key := make([]byte, 32) // AES-256
		for i := range key {
			key[i] = byte(i + 20)
		}
		plaintext := []byte("Testing AES-256 encryption, the strongest AES variant.")

		ciphertext, err := k.EncryptAESGCMExported(plaintext, key)
		require.NoError(t, err)
		require.NotNil(t, ciphertext)
		require.Greater(t, len(ciphertext), len(plaintext))
	})

	t.Run("invalid key size returns error", func(t *testing.T) {
		invalidKeySizes := []int{0, 1, 8, 15, 17, 23, 25, 31, 33, 64}

		for _, size := range invalidKeySizes {
			key := make([]byte, size)
			for i := range key {
				key[i] = byte(i)
			}
			plaintext := []byte("test data")

			ciphertext, err := k.EncryptAESGCMExported(plaintext, key)
			require.Error(t, err, "Expected error for key size %d", size)
			require.Nil(t, ciphertext)
		}
	})

	t.Run("encrypted output is longer than plaintext", func(t *testing.T) {
		key := make([]byte, 32)
		for i := range key {
			key[i] = byte(i)
		}

		testCases := [][]byte{
			[]byte("a"),
			[]byte("short"),
			[]byte("medium length message here"),
			make([]byte, 1000), // 1KB
		}

		for _, plaintext := range testCases {
			ciphertext, err := k.EncryptAESGCMExported(plaintext, key)
			require.NoError(t, err)
			// AES-GCM adds 12-byte nonce prepended and 16-byte auth tag
			// Total overhead: 28 bytes minimum
			require.Greater(t, len(ciphertext), len(plaintext),
				"Ciphertext should be longer than plaintext")
		}
	})

	t.Run("empty plaintext can be encrypted", func(t *testing.T) {
		key := make([]byte, 32)
		for i := range key {
			key[i] = byte(i)
		}
		plaintext := []byte{}

		ciphertext, err := k.EncryptAESGCMExported(plaintext, key)
		require.NoError(t, err)
		require.NotNil(t, ciphertext)
		// Even empty plaintext produces output (nonce + tag)
		require.Greater(t, len(ciphertext), 0)
	})

	t.Run("same plaintext produces different ciphertext (random nonce)", func(t *testing.T) {
		key := make([]byte, 32)
		for i := range key {
			key[i] = byte(i)
		}
		plaintext := []byte("This message will be encrypted twice.")

		ciphertext1, err := k.EncryptAESGCMExported(plaintext, key)
		require.NoError(t, err)

		ciphertext2, err := k.EncryptAESGCMExported(plaintext, key)
		require.NoError(t, err)

		// Due to random nonce, ciphertexts should be different
		require.NotEqual(t, ciphertext1, ciphertext2,
			"Same plaintext should produce different ciphertexts due to random nonce")
	})
}

// ============================================================================
// TestExtractSPKIHash - Tests for extractSPKIHash function (cert_pinning.go:143)
// ============================================================================

func TestExtractSPKIHash(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("invalid certificate bytes returns error", func(t *testing.T) {
		invalidCerts := [][]byte{
			nil,
			{},
			[]byte("not a certificate"),
			[]byte{0x30, 0x82, 0x01}, // Truncated ASN.1
			make([]byte, 100),        // Random bytes
		}

		for i, invalidCert := range invalidCerts {
			hash, err := k.ExtractSPKIHashExported(invalidCert)
			require.Error(t, err, "Expected error for invalid cert case %d", i)
			require.Nil(t, hash)
		}
	})

	t.Run("valid X509 certificate returns 32-byte SHA256 hash", func(t *testing.T) {
		certDER := generateTestCertificate()
		require.NotEmpty(t, certDER)

		hash, err := k.ExtractSPKIHashExported(certDER)
		require.NoError(t, err)
		require.NotNil(t, hash)
		require.Len(t, hash, 32, "SHA256 hash should be exactly 32 bytes")
	})

	t.Run("same certificate produces consistent hash", func(t *testing.T) {
		certDER := generateTestCertificate()

		hash1, err := k.ExtractSPKIHashExported(certDER)
		require.NoError(t, err)

		hash2, err := k.ExtractSPKIHashExported(certDER)
		require.NoError(t, err)

		require.Equal(t, hash1, hash2, "Same certificate should produce same SPKI hash")
	})

	t.Run("different certificates produce different hashes", func(t *testing.T) {
		cert1 := generateTestCertificate()
		cert2 := generateTestCertificate()

		hash1, err := k.ExtractSPKIHashExported(cert1)
		require.NoError(t, err)

		hash2, err := k.ExtractSPKIHashExported(cert2)
		require.NoError(t, err)

		// Different keys should produce different SPKI hashes
		require.NotEqual(t, hash1, hash2, "Different certificates should produce different SPKI hashes")
	})

	t.Run("SPKI hash can be used for certificate pinning", func(t *testing.T) {
		certDER := generateTestCertificate()

		spkiHash, err := k.ExtractSPKIHashExported(certDER)
		require.NoError(t, err)
		require.Len(t, spkiHash, 32)

		// Use the SPKI hash to create a certificate pin
		_, err = k.AddCertificatePin(
			ctx,
			"creator",
			"spki-test.example.com",
			[][]byte{spkiHash},
			cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI,
			nil,
		)
		require.NoError(t, err)
	})
}

// ============================================================================
// TestUpdateRandomSourceStatus - Tests for updateRandomSourceStatus (random.go:184)
// ============================================================================

func TestUpdateRandomSourceStatus(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("non-existent source returns error", func(t *testing.T) {
		err := k.UpdateRandomSourceStatusExported(
			ctx,
			"non-existent-source-id",
			cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_FAILED,
		)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get")
	})

	t.Run("existing source status is updated successfully", func(t *testing.T) {
		// First, initialize a random source
		entropy := make([]byte, 64)
		for i := range entropy {
			entropy[i] = byte(i)
		}

		sourceID, err := k.InitializeRandomSource(
			ctx,
			cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_SYSTEM,
			entropy,
		)
		require.NoError(t, err)
		require.NotEmpty(t, sourceID)

		// Verify initial status is HEALTHY
		source, err := k.GetRandomSource(ctx, sourceID)
		require.NoError(t, err)
		require.Equal(t, cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY, source.Status)

		// Update status to FAILED
		err = k.UpdateRandomSourceStatusExported(
			ctx,
			sourceID,
			cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_FAILED,
		)
		require.NoError(t, err)

		// Verify status was updated
		source, err = k.GetRandomSource(ctx, sourceID)
		require.NoError(t, err)
		require.Equal(t, cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_FAILED, source.Status)
	})

	t.Run("update to LOW_ENTROPY status", func(t *testing.T) {
		entropy := make([]byte, 64)
		for i := range entropy {
			entropy[i] = byte(i + 100)
		}

		sourceID, err := k.InitializeRandomSource(
			ctx,
			cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_HARDWARE,
			entropy,
		)
		require.NoError(t, err)

		// Update to LOW_ENTROPY
		err = k.UpdateRandomSourceStatusExported(
			ctx,
			sourceID,
			cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_LOW_ENTROPY,
		)
		require.NoError(t, err)

		source, err := k.GetRandomSource(ctx, sourceID)
		require.NoError(t, err)
		require.Equal(t, cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_LOW_ENTROPY, source.Status)
	})

	t.Run("update back to HEALTHY status", func(t *testing.T) {
		entropy := make([]byte, 64)
		for i := range entropy {
			entropy[i] = byte(i + 200)
		}

		sourceID, err := k.InitializeRandomSource(
			ctx,
			cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_SYSTEM,
			entropy,
		)
		require.NoError(t, err)

		// Set to FAILED first
		err = k.UpdateRandomSourceStatusExported(
			ctx,
			sourceID,
			cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_FAILED,
		)
		require.NoError(t, err)

		// Recover to HEALTHY
		err = k.UpdateRandomSourceStatusExported(
			ctx,
			sourceID,
			cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY,
		)
		require.NoError(t, err)

		source, err := k.GetRandomSource(ctx, sourceID)
		require.NoError(t, err)
		require.Equal(t, cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY, source.Status)
	})

	t.Run("empty source ID returns error", func(t *testing.T) {
		err := k.UpdateRandomSourceStatusExported(
			ctx,
			"",
			cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY,
		)
		require.Error(t, err)
	})
}

// ============================================================================
// TestRegisterInvariantsExtended - Extended tests for RegisterInvariants (invariants.go:18)
// ============================================================================

func TestRegisterInvariantsExtended(t *testing.T) {
	k, ctx := setupKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	t.Run("can register invariants without panic using mock registry", func(t *testing.T) {
		// Create a mock invariant registry
		registry := &mockInvariantRegistryExtended{
			routes: make(map[string]sdk.Invariant),
		}

		// This should not panic
		require.NotPanics(t, func() {
			keeper.RegisterInvariants(registry, k)
		})

		// Verify invariants were registered
		require.Greater(t, len(registry.routes), 0, "At least one invariant should be registered")

		// Check for expected invariant routes
		expectedRoutes := []string{
			"cryptography/params-valid",
			"cryptography/key-rotation-validity",
			"cryptography/threshold-scheme-consistency",
			"cryptography/zk-proof-config-validity",
			"cryptography/secure-enclave-validity",
			"cryptography/quantum-key-validity",
		}

		for _, route := range expectedRoutes {
			_, exists := registry.routes[route]
			require.True(t, exists, "Expected invariant route %s to be registered", route)
		}
	})

	t.Run("invariants pass with valid populated data", func(t *testing.T) {
		// Create valid key rotation schedule
		policy := &cryptoproto.KeyRotationPolicy{
			MaxAgeDays:              90,
			WarningDaysBeforeExpiry: 7,
			AutoRotate:              true,
			MaxRotationAttempts:     3,
		}
		_, err := k.CreateKeyRotationSchedule(ctx, "creator", "test-key-inv", 86400, policy)
		require.NoError(t, err)

		// Create valid threshold scheme
		participants := []string{"p1", "p2", "p3"}
		_, _, err = k.CreateThresholdScheme(
			ctx,
			"creator",
			2,
			3,
			participants,
			cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA,
		)
		require.NoError(t, err)

		// Create valid quantum key
		pubKey := GenerateDummyQuantumPublicKey(cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM)
		_, err = k.RegisterQuantumResistantKey(
			ctx,
			"creator",
			cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM,
			pubKey,
			nil,
		)
		require.NoError(t, err)

		// Register secure enclave
		attestation := make([]byte, 432)
		_, err = k.RegisterSecureEnclave(
			ctx,
			"creator",
			cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX,
			attestation,
			nil,
		)
		require.NoError(t, err)

		// Run all invariants
		allInv := keeper.AllInvariants(k)
		msg, broken := allInv(sdkCtx)
		require.False(t, broken, "All invariants should pass with valid data: %s", msg)
	})

	t.Run("registered invariants can be executed individually", func(t *testing.T) {
		registry := &mockInvariantRegistryExtended{
			routes: make(map[string]sdk.Invariant),
		}

		keeper.RegisterInvariants(registry, k)

		// Execute each registered invariant
		for route, inv := range registry.routes {
			msg, broken := inv(sdkCtx)
			require.False(t, broken, "Invariant %s should pass: %s", route, msg)
		}
	})
}

// mockInvariantRegistryExtended implements sdk.InvariantRegistry for testing
type mockInvariantRegistryExtended struct {
	routes map[string]sdk.Invariant
}

func (m *mockInvariantRegistryExtended) RegisterRoute(moduleName string, invariantName string, invariant sdk.Invariant) {
	key := moduleName + "/" + invariantName
	m.routes[key] = invariant
}
