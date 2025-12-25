// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModuleConstants(t *testing.T) {
	require.Equal(t, "cryptography", ModuleName)
	require.Equal(t, "cryptography", StoreKey)
	require.Equal(t, "cryptography", RouterKey)
	require.Equal(t, "cryptography", QuerierRoute)
}

func TestStoreKeys(t *testing.T) {
	// Test that all prefix keys are unique
	prefixes := [][]byte{
		ParamsKey,
		KeyRotationSchedulePrefix,
		HDKeyDerivationPrefix,
		ThresholdSchemePrefix,
		ThresholdSignatureSharePrefix,
		ZKProofConfigPrefix,
		ZKProofPrefix,
		SecureEnclavePrefix,
		QuantumResistantKeyPrefix,
		CryptoRandomSourcePrefix,
		SaltedHashPrefix,
		KeyStretchingConfigPrefix,
		CertificatePinPrefix,
	}

	// Ensure all prefixes are unique
	seen := make(map[byte]bool)
	for _, prefix := range prefixes {
		require.NotEmpty(t, prefix)
		require.False(t, seen[prefix[0]], "duplicate prefix found")
		seen[prefix[0]] = true
	}
}

func TestGetKeyRotationScheduleKey(t *testing.T) {
	id := "rotation-123"
	key := GetKeyRotationScheduleKey(id)

	require.NotEmpty(t, key)
	require.Equal(t, KeyRotationSchedulePrefix[0], key[0])
	require.Contains(t, string(key), id)
}

func TestGetHDKeyDerivationKey(t *testing.T) {
	masterKeyID := "master-key-1"
	key := GetHDKeyDerivationKey(masterKeyID)

	require.NotEmpty(t, key)
	require.Equal(t, HDKeyDerivationPrefix[0], key[0])
	require.Contains(t, string(key), masterKeyID)
}

func TestGetThresholdSchemeKey(t *testing.T) {
	schemeID := "scheme-123"
	key := GetThresholdSchemeKey(schemeID)

	require.NotEmpty(t, key)
	require.Equal(t, ThresholdSchemePrefix[0], key[0])
	require.Contains(t, string(key), schemeID)
}

func TestGetThresholdSignatureShareKey(t *testing.T) {
	schemeID := "scheme-123"
	participantID := "participant-456"
	key := GetThresholdSignatureShareKey(schemeID, participantID)

	require.NotEmpty(t, key)
	require.Equal(t, ThresholdSignatureSharePrefix[0], key[0])
	require.Contains(t, string(key), schemeID)
	require.Contains(t, string(key), participantID)
	require.Contains(t, string(key), "/") // Separator
}

func TestGetZKProofConfigKey(t *testing.T) {
	proofID := "proof-789"
	key := GetZKProofConfigKey(proofID)

	require.NotEmpty(t, key)
	require.Equal(t, ZKProofConfigPrefix[0], key[0])
	require.Contains(t, string(key), proofID)
}

func TestGetZKProofKey(t *testing.T) {
	proofID := "proof-789"
	key := GetZKProofKey(proofID)

	require.NotEmpty(t, key)
	require.Equal(t, ZKProofPrefix[0], key[0])
	require.Contains(t, string(key), proofID)
}

func TestGetSecureEnclaveKey(t *testing.T) {
	enclaveID := "enclave-abc"
	key := GetSecureEnclaveKey(enclaveID)

	require.NotEmpty(t, key)
	require.Equal(t, SecureEnclavePrefix[0], key[0])
	require.Contains(t, string(key), enclaveID)
}

func TestGetQuantumResistantKeyKey(t *testing.T) {
	keyID := "qr-key-def"
	key := GetQuantumResistantKeyKey(keyID)

	require.NotEmpty(t, key)
	require.Equal(t, QuantumResistantKeyPrefix[0], key[0])
	require.Contains(t, string(key), keyID)
}

func TestGetCryptoRandomSourceKey(t *testing.T) {
	sourceID := "source-ghi"
	key := GetCryptoRandomSourceKey(sourceID)

	require.NotEmpty(t, key)
	require.Equal(t, CryptoRandomSourcePrefix[0], key[0])
	require.Contains(t, string(key), sourceID)
}

func TestGetSaltedHashKey(t *testing.T) {
	hashID := "hash-jkl"
	key := GetSaltedHashKey(hashID)

	require.NotEmpty(t, key)
	require.Equal(t, SaltedHashPrefix[0], key[0])
	require.Contains(t, string(key), hashID)
}

func TestGetKeyStretchingConfigKey(t *testing.T) {
	configID := "config-mno"
	key := GetKeyStretchingConfigKey(configID)

	require.NotEmpty(t, key)
	require.Equal(t, KeyStretchingConfigPrefix[0], key[0])
	require.Contains(t, string(key), configID)
}

func TestGetCertificatePinKey(t *testing.T) {
	hostname := "example.com"
	key := GetCertificatePinKey(hostname)

	require.NotEmpty(t, key)
	require.Equal(t, CertificatePinPrefix[0], key[0])
	require.Contains(t, string(key), hostname)
}

func TestKeyUniqueness(t *testing.T) {
	// Test that different IDs produce different keys
	key1 := GetKeyRotationScheduleKey("id1")
	key2 := GetKeyRotationScheduleKey("id2")
	require.NotEqual(t, key1, key2)

	// Test that same ID with different functions produces different keys
	id := "test-id"
	rotationKey := GetKeyRotationScheduleKey(id)
	schemeKey := GetThresholdSchemeKey(id)
	require.NotEqual(t, rotationKey, schemeKey)
}
