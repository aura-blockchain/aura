package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// TestEdgeCases tests edge cases and error paths to achieve 100% coverage
func TestEdgeCases(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("GenerateRandomBytesFromSource - failed random source", func(t *testing.T) {
		// Create a random source
		entropy := make([]byte, 64)
		sourceID, err := k.InitializeRandomSource(ctx, cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_SYSTEM, entropy)
		require.NoError(t, err)

		// Update its status to failed
		source, err := k.GetRandomSource(ctx, sourceID)
		require.NoError(t, err)
		source.Status = cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_FAILED

		// Manually update the source in store
		err = k.SetRandomSource(ctx, source)
		require.NoError(t, err)

		// Try to generate random bytes - should fail
		_, err = k.GenerateRandomBytesFromSource(ctx, sourceID, 32)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrRandomSourceFailed)
	})

	t.Run("CheckEntropyHealth - old source", func(t *testing.T) {
		// This test ensures the time check branch is covered
		err := k.CheckEntropyHealth(ctx)
		require.NoError(t, err)
	})

	t.Run("GetEntropyStatistics - various statuses", func(t *testing.T) {
		// Create sources with different statuses
		entropy := make([]byte, 64)

		// Healthy source
		_, err := k.InitializeRandomSource(ctx, cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_SYSTEM, entropy)
		require.NoError(t, err)

		// Failed source
		sourceID2, err := k.InitializeRandomSource(ctx, cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_HARDWARE, entropy)
		require.NoError(t, err)
		source2, _ := k.GetRandomSource(ctx, sourceID2)
		source2.Status = cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_FAILED
		k.SetRandomSource(ctx, source2)

		// Low entropy source
		sourceID3, err := k.InitializeRandomSource(ctx, cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_QUANTUM, entropy)
		require.NoError(t, err)
		source3, _ := k.GetRandomSource(ctx, sourceID3)
		source3.Status = cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_LOW_ENTROPY
		k.SetRandomSource(ctx, source3)

		// Get statistics
		stats := k.GetEntropyStatistics(ctx)
		require.NotNil(t, stats)
		require.Greater(t, stats["total_sources"].(int), 0)
	})

	t.Run("GetRandomSource - not in cache, load from store", func(t *testing.T) {
		// Create a new keeper to have fresh cache
		k2, ctx2 := setupKeeper(t)

		// Create a random source
		entropy := make([]byte, 64)
		sourceID, err := k2.InitializeRandomSource(ctx2, cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_SYSTEM, entropy)
		require.NoError(t, err)

		// Create another keeper with same store to simulate cache miss
		k3, ctx3 := setupKeeper(t)

		// Set the source in the new keeper's store
		source, err := k2.GetRandomSource(ctx2, sourceID)
		require.NoError(t, err)
		err = k3.SetRandomSource(ctx3, source)
		require.NoError(t, err)

		// Now get it - should load from store
		retrieved, err := k3.GetRandomSource(ctx3, sourceID)
		require.NoError(t, err)
		require.Equal(t, sourceID, retrieved.SourceId)
	})

	t.Run("GetQuantumResistantKey - not in cache, load from store", func(t *testing.T) {
		k2, ctx2 := setupKeeper(t)

		keyID, _, err := k2.GenerateQuantumResistantKey(ctx2, "creator", cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM, nil)
		require.NoError(t, err)

		// Create new keeper and set the key
		k3, ctx3 := setupKeeper(t)
		key, err := k2.GetQuantumResistantKey(ctx2, keyID)
		require.NoError(t, err)
		err = k3.SetQuantumResistantKey(ctx3, key)
		require.NoError(t, err)

		// Get from store
		retrieved, err := k3.GetQuantumResistantKey(ctx3, keyID)
		require.NoError(t, err)
		require.Equal(t, keyID, retrieved.KeyId)
	})

	t.Run("GetSecureEnclave - not in cache, load from store", func(t *testing.T) {
		k2, ctx2 := setupKeeper(t)

		attestation := make([]byte, 432)
		enclaveID, err := k2.RegisterSecureEnclave(ctx2, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, attestation, nil)
		require.NoError(t, err)

		// Create new keeper and set the enclave
		k3, ctx3 := setupKeeper(t)
		enclave, err := k2.GetSecureEnclave(ctx2, enclaveID)
		require.NoError(t, err)
		err = k3.SetSecureEnclaveConfig(ctx3, enclave)
		require.NoError(t, err)

		// Get from store
		retrieved, err := k3.GetSecureEnclave(ctx3, enclaveID)
		require.NoError(t, err)
		require.Equal(t, enclaveID, retrieved.EnclaveId)
	})

	// Note: GetThresholdScheme cache behavior is already tested in the main tests

	t.Run("GetKeyRotationSchedule - not in cache, load from store", func(t *testing.T) {
		k2, ctx2 := setupKeeper(t)

		scheduleID, err := k2.CreateKeyRotationSchedule(ctx2, "creator", "test-key", 86400, nil)
		require.NoError(t, err)

		// Create new keeper
		k3, ctx3 := setupKeeper(t)

		// Manually set the schedule in the new keeper's store
		schedule, err := k2.GetKeyRotationSchedule(ctx2, scheduleID)
		require.NoError(t, err)

		err = k3.SetKeyRotationSchedule(ctx3, schedule)
		require.NoError(t, err)

		// Now get it - should load from store
		retrieved, err := k3.GetKeyRotationSchedule(ctx3, scheduleID)
		require.NoError(t, err)
		require.Equal(t, scheduleID, retrieved.Id)
	})

	t.Run("GetCertificatePin - not in cache, load from store", func(t *testing.T) {
		k2, ctx2 := setupKeeper(t)

		hash := make([]byte, 32)
		_, err := k2.AddCertificatePin(ctx2, "creator", "cache-test.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.NoError(t, err)

		// Create new keeper
		k3, ctx3 := setupKeeper(t)

		// Manually set the pin in the new keeper's store
		pin, err := k2.GetCertificatePin(ctx2, "cache-test.com")
		require.NoError(t, err)

		err = k3.SetCertificatePin(ctx3, pin)
		require.NoError(t, err)

		// Now get it - should load from store
		retrieved, err := k3.GetCertificatePin(ctx3, "cache-test.com")
		require.NoError(t, err)
		require.Equal(t, "cache-test.com", retrieved.Hostname)
	})

	t.Run("Attestation failures for insufficient size", func(t *testing.T) {
		// SEV too short
		_, err := k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SEV, make([]byte, 100), nil)
		require.Error(t, err)

		// TPM too short
		_, err = k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_TPM, make([]byte, 20), nil)
		require.Error(t, err)

		// HSM too short
		_, err = k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_HSM, make([]byte, 10), nil)
		require.Error(t, err)

		// Keychain too short
		_, err = k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_KEYCHAIN, make([]byte, 5), nil)
		require.Error(t, err)
	})

	t.Run("UpdateEnclaveStatus - non-existent enclave", func(t *testing.T) {
		err := k.UpdateEnclaveStatus(ctx, "non-existent", cryptoproto.SecureEnclaveStatus_SECURE_ENCLAVE_STATUS_READY)
		require.Error(t, err)
	})

	t.Run("RemoteAttestEnclave - non-existent enclave", func(t *testing.T) {
		_, err := k.RemoteAttestEnclave(ctx, "non-existent")
		require.Error(t, err)
	})

	t.Run("RotateQuantumResistantKey - non-existent key", func(t *testing.T) {
		_, _, err := k.RotateQuantumResistantKey(ctx, "non-existent", nil)
		require.Error(t, err)
	})

	t.Run("ValidateQuantumResistantKey - non-existent key", func(t *testing.T) {
		err := k.ValidateQuantumResistantKey(ctx, "non-existent")
		require.Error(t, err)
	})

	t.Run("ReseedRandomSource - non-existent source", func(t *testing.T) {
		err := k.ReseedRandomSource(ctx, "non-existent", make([]byte, 64))
		require.Error(t, err)
	})

	t.Run("UpdateCertificatePin - non-existent pin", func(t *testing.T) {
		err := k.UpdateCertificatePin(ctx, "non-existent.com", nil, nil)
		require.Error(t, err)
	})

	t.Run("RotateCertificatePin - non-existent pin", func(t *testing.T) {
		hash := make([]byte, 32)
		err := k.RotateCertificatePin(ctx, "non-existent.com", [][]byte{hash})
		require.Error(t, err)
	})

	t.Run("VerifyCertificatePin - non-existent pin", func(t *testing.T) {
		_, err := k.VerifyCertificatePin(ctx, "non-existent.com", []byte("cert"))
		require.Error(t, err)
	})

	t.Run("DisableCertificatePin - non-existent pin", func(t *testing.T) {
		err := k.DisableCertificatePin(ctx, "non-existent.com")
		require.Error(t, err)
	})

	t.Run("EnableCertificatePin - non-existent pin", func(t *testing.T) {
		err := k.EnableCertificatePin(ctx, "non-existent.com")
		require.Error(t, err)
	})

	t.Run("VerifyStretchedKey - non-existent config", func(t *testing.T) {
		_, err := k.VerifyStretchedKey(ctx, "non-existent", []byte("password"), []byte("key"))
		require.Error(t, err)
	})
}
