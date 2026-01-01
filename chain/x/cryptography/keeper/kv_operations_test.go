// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// TestDeleteKeyRotationSchedule tests the DeleteKeyRotationSchedule function
func TestDeleteKeyRotationSchedule(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("delete existing schedule", func(t *testing.T) {
		// Create a key rotation schedule
		policy := &cryptoproto.KeyRotationPolicy{
			MaxAgeDays:              90,
			WarningDaysBeforeExpiry: 7,
			AutoRotate:              true,
			MaxRotationAttempts:     3,
		}

		scheduleID, err := k.CreateKeyRotationSchedule(
			ctx,
			"creator",
			"test-key-delete-1",
			86400, // 24 hours
			policy,
		)
		require.NoError(t, err)
		require.NotEmpty(t, scheduleID)

		// Verify it exists
		schedule, err := k.GetKeyRotationSchedule(ctx, scheduleID)
		require.NoError(t, err)
		require.NotNil(t, schedule)
		require.Equal(t, "test-key-delete-1", schedule.KeyId)

		// Delete the schedule
		err = k.DeleteKeyRotationSchedule(ctx, scheduleID)
		require.NoError(t, err)

		// Verify it's gone
		_, err = k.GetKeyRotationSchedule(ctx, scheduleID)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrKeyRotationScheduleNotFound)
	})

	t.Run("delete non-existent schedule is idempotent", func(t *testing.T) {
		// Deleting a non-existent schedule should not return an error
		err := k.DeleteKeyRotationSchedule(ctx, "non-existent-schedule")
		require.NoError(t, err)
	})
}

// TestGetKeyStretchingConfig tests the GetKeyStretchingConfig function
func TestGetKeyStretchingConfig(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("not found case", func(t *testing.T) {
		_, err := k.GetKeyStretchingConfig(ctx, "non-existent-config")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidKeyStretchingConfig)
	})

	t.Run("valid retrieval", func(t *testing.T) {
		// Create a key stretching config
		configID, salt, err := k.CreateKeyStretchingConfig(
			ctx,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA256,
			100000,
			0,
			0,
			32,
		)
		require.NoError(t, err)
		require.NotEmpty(t, configID)
		require.NotEmpty(t, salt)

		// Retrieve the config
		config, err := k.GetKeyStretchingConfig(ctx, configID)
		require.NoError(t, err)
		require.NotNil(t, config)
		require.Equal(t, cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA256, config.Algorithm)
		require.Equal(t, int32(100000), config.Iterations)
		require.Equal(t, int32(32), config.KeyLength)
	})
}

// TestSecureEnclaveOperations tests SetSecureEnclaveConfig, GetSecureEnclave, DeleteSecureEnclave
func TestSecureEnclaveOperations(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("not found case", func(t *testing.T) {
		_, err := k.GetSecureEnclave(ctx, "non-existent-enclave")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrSecureEnclaveNotFound)
	})

	t.Run("set, get, delete enclave", func(t *testing.T) {
		enclaveConfig := &cryptoproto.SecureEnclaveConfig{
			EnclaveId:       "test-enclave-1",
			EnclaveType:     cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX,
			AttestationData: make([]byte, 432),
			EnclaveMetadata: map[string]string{
				"version": "1.0",
				"vendor":  "Intel",
			},
		}

		// Set the enclave config
		err := k.SetSecureEnclaveConfig(ctx, enclaveConfig)
		require.NoError(t, err)

		// Get and verify
		retrieved, err := k.GetSecureEnclave(ctx, "test-enclave-1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, "test-enclave-1", retrieved.EnclaveId)
		require.Equal(t, cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, retrieved.EnclaveType)
		require.Equal(t, "1.0", retrieved.EnclaveMetadata["version"])

		// Delete the enclave
		err = k.DeleteSecureEnclave(ctx, "test-enclave-1")
		require.NoError(t, err)

		// Verify it's gone
		_, err = k.GetSecureEnclave(ctx, "test-enclave-1")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrSecureEnclaveNotFound)
	})

	t.Run("set nil enclave config is no-op", func(t *testing.T) {
		err := k.SetSecureEnclaveConfig(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("delete non-existent enclave is idempotent", func(t *testing.T) {
		err := k.DeleteSecureEnclave(ctx, "non-existent-enclave")
		require.NoError(t, err)
	})
}

// TestRandomSourceOperations tests SetRandomSource, GetRandomSource, DeleteRandomSource
func TestRandomSourceOperations(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("not found case", func(t *testing.T) {
		_, err := k.GetRandomSource(ctx, "non-existent-source")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrRandomSourceFailed)
	})

	t.Run("set, get, delete random source", func(t *testing.T) {
		source := &cryptoproto.CryptoRandomSource{
			SourceId:    "test-source-1",
			SourceType:  cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_HARDWARE,
			EntropyBits: 256,
			Status:      cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY,
		}

		// Set the random source
		err := k.SetRandomSource(ctx, source)
		require.NoError(t, err)

		// Get and verify
		retrieved, err := k.GetRandomSource(ctx, "test-source-1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, "test-source-1", retrieved.SourceId)
		require.Equal(t, cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_HARDWARE, retrieved.SourceType)
		require.Equal(t, int64(256), retrieved.EntropyBits)
		require.Equal(t, cryptoproto.RandomSourceStatus_RANDOM_SOURCE_STATUS_HEALTHY, retrieved.Status)

		// Delete the source
		err = k.DeleteRandomSource(ctx, "test-source-1")
		require.NoError(t, err)

		// Verify it's gone
		_, err = k.GetRandomSource(ctx, "test-source-1")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrRandomSourceFailed)
	})

	t.Run("set nil random source is no-op", func(t *testing.T) {
		err := k.SetRandomSource(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("delete non-existent source is idempotent", func(t *testing.T) {
		err := k.DeleteRandomSource(ctx, "non-existent-source")
		require.NoError(t, err)
	})
}

// TestThresholdSchemeOperations tests SetThresholdScheme, GetThresholdScheme, DeleteThresholdScheme
func TestThresholdSchemeOperations(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("not found case", func(t *testing.T) {
		_, err := k.GetThresholdScheme(ctx, "non-existent-scheme")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrThresholdSchemeNotFound)
	})

	t.Run("set, get, delete threshold scheme", func(t *testing.T) {
		scheme := &cryptoproto.ThresholdSignatureScheme{
			SchemeId:          "test-scheme-1",
			Threshold:         3,
			TotalParticipants: 5,
			SchemeType:        cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA,
			ParticipantIds:    []string{"p1", "p2", "p3", "p4", "p5"},
			PublicKey:         make([]byte, 33),
		}

		// Set the scheme
		err := k.SetThresholdScheme(ctx, scheme)
		require.NoError(t, err)

		// Get and verify
		retrieved, err := k.GetThresholdScheme(ctx, "test-scheme-1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, "test-scheme-1", retrieved.SchemeId)
		require.Equal(t, int32(3), retrieved.Threshold)
		require.Equal(t, int32(5), retrieved.TotalParticipants)

		// Delete the scheme
		err = k.DeleteThresholdScheme(ctx, "test-scheme-1")
		require.NoError(t, err)

		// Verify it's gone
		_, err = k.GetThresholdScheme(ctx, "test-scheme-1")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrThresholdSchemeNotFound)
	})

	t.Run("set nil scheme is no-op", func(t *testing.T) {
		err := k.SetThresholdScheme(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("delete non-existent scheme is idempotent", func(t *testing.T) {
		err := k.DeleteThresholdScheme(ctx, "non-existent-scheme")
		require.NoError(t, err)
	})
}

// TestThresholdSignatureShareOperations tests Set, Get, Delete for threshold signature shares
func TestThresholdSignatureShareOperations(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("not found case", func(t *testing.T) {
		_, err := k.GetThresholdSignatureShare(ctx, "non-existent-scheme", "non-existent-participant")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidSignatureShare)
	})

	t.Run("set, get, delete signature share", func(t *testing.T) {
		share := &cryptoproto.ThresholdSignatureShare{
			SchemeId:       "scheme-share-1",
			ParticipantId:  "participant-1",
			SignatureShare: make([]byte, 64),
			MessageHash:    make([]byte, 32),
		}

		// Set the share
		err := k.SetThresholdSignatureShare(ctx, share)
		require.NoError(t, err)

		// Get and verify
		retrieved, err := k.GetThresholdSignatureShare(ctx, "scheme-share-1", "participant-1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, "scheme-share-1", retrieved.SchemeId)
		require.Equal(t, "participant-1", retrieved.ParticipantId)

		// Delete the share
		err = k.DeleteThresholdSignatureShare(ctx, "scheme-share-1", "participant-1")
		require.NoError(t, err)

		// Verify it's gone
		_, err = k.GetThresholdSignatureShare(ctx, "scheme-share-1", "participant-1")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidSignatureShare)
	})

	t.Run("set nil share is no-op", func(t *testing.T) {
		err := k.SetThresholdSignatureShare(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("delete non-existent share is idempotent", func(t *testing.T) {
		err := k.DeleteThresholdSignatureShare(ctx, "non-existent-scheme", "non-existent-participant")
		require.NoError(t, err)
	})
}

// TestZKProofCRUD tests SetZKProof, GetZKProof, DeleteZKProof
func TestZKProofCRUD(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("not found case", func(t *testing.T) {
		_, err := k.GetZKProof(ctx, "non-existent-proof")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidZKProof)
	})

	t.Run("set, get, delete ZK proof", func(t *testing.T) {
		proof := &cryptoproto.ZKProof{
			ProofId:      "test-proof-1",
			ProofData:    make([]byte, 128),
			PublicInputs: make([]byte, 64),
			Verified:     false,
		}

		// Set the proof
		err := k.SetZKProof(ctx, proof)
		require.NoError(t, err)

		// Get and verify
		retrieved, err := k.GetZKProof(ctx, "test-proof-1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, "test-proof-1", retrieved.ProofId)
		require.Len(t, retrieved.ProofData, 128)

		// Delete the proof
		err = k.DeleteZKProof(ctx, "test-proof-1")
		require.NoError(t, err)

		// Verify it's gone
		_, err = k.GetZKProof(ctx, "test-proof-1")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidZKProof)
	})

	t.Run("set nil proof is no-op", func(t *testing.T) {
		err := k.SetZKProof(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("delete non-existent proof is idempotent", func(t *testing.T) {
		err := k.DeleteZKProof(ctx, "non-existent-proof")
		require.NoError(t, err)
	})

	t.Run("multiple proofs CRUD", func(t *testing.T) {
		// Create multiple proofs
		for i := 0; i < 3; i++ {
			proof := &cryptoproto.ZKProof{
				ProofId:      "multi-proof-" + string(rune('a'+i)),
				ProofData:    make([]byte, 64),
				PublicInputs: make([]byte, 32),
			}
			err := k.SetZKProof(ctx, proof)
			require.NoError(t, err)
		}

		// Verify all exist
		for i := 0; i < 3; i++ {
			retrieved, err := k.GetZKProof(ctx, "multi-proof-"+string(rune('a'+i)))
			require.NoError(t, err)
			require.NotNil(t, retrieved)
		}

		// Delete all
		for i := 0; i < 3; i++ {
			err := k.DeleteZKProof(ctx, "multi-proof-"+string(rune('a'+i)))
			require.NoError(t, err)
		}

		// Verify all are gone
		for i := 0; i < 3; i++ {
			_, err := k.GetZKProof(ctx, "multi-proof-"+string(rune('a'+i)))
			require.Error(t, err)
		}
	})
}

// TestZKProofVerificationCRUD tests SetZKProofVerification, GetZKProofVerification, DeleteZKProofVerification
func TestZKProofVerificationCRUD(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("not found case", func(t *testing.T) {
		_, err := k.GetZKProofVerification(ctx, "non-existent-verification")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidZKProof)
	})

	t.Run("set, get, delete ZK proof verification", func(t *testing.T) {
		verification := &cryptoproto.ZKProofVerification{
			VerificationId: "test-verification-1",
			ProofId:        "proof-1",
			Verified:       true,
			Submitter:      "submitter-address",
			VerifierNode:   "verifier-node",
		}

		// Set the verification
		err := k.SetZKProofVerification(ctx, verification)
		require.NoError(t, err)

		// Get and verify
		retrieved, err := k.GetZKProofVerification(ctx, "test-verification-1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, "test-verification-1", retrieved.VerificationId)
		require.Equal(t, "proof-1", retrieved.ProofId)
		require.True(t, retrieved.Verified)
		require.Equal(t, "submitter-address", retrieved.Submitter)

		// Delete the verification
		err = k.DeleteZKProofVerification(ctx, "test-verification-1")
		require.NoError(t, err)

		// Verify it's gone
		_, err = k.GetZKProofVerification(ctx, "test-verification-1")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidZKProof)
	})

	t.Run("set nil verification is no-op", func(t *testing.T) {
		err := k.SetZKProofVerification(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("delete non-existent verification is idempotent", func(t *testing.T) {
		err := k.DeleteZKProofVerification(ctx, "non-existent-verification")
		require.NoError(t, err)
	})

	t.Run("verification with invalid result", func(t *testing.T) {
		verification := &cryptoproto.ZKProofVerification{
			VerificationId: "test-verification-invalid",
			ProofId:        "proof-invalid",
			Verified:       false,
			Submitter:      "submitter-address",
			ErrorMessage:   "proof verification failed: invalid witness",
		}

		err := k.SetZKProofVerification(ctx, verification)
		require.NoError(t, err)

		retrieved, err := k.GetZKProofVerification(ctx, "test-verification-invalid")
		require.NoError(t, err)
		require.False(t, retrieved.Verified)
		require.Equal(t, "proof verification failed: invalid witness", retrieved.ErrorMessage)

		// Cleanup
		err = k.DeleteZKProofVerification(ctx, "test-verification-invalid")
		require.NoError(t, err)
	})
}

// TestSaltedHashCRUD tests SetSaltedHash, GetSaltedHash, DeleteSaltedHash
func TestSaltedHashCRUD(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("not found case", func(t *testing.T) {
		_, err := k.GetSaltedHash(ctx, "non-existent-hash")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidHashAlgorithm)
	})

	t.Run("set, get, delete salted hash", func(t *testing.T) {
		hash := &cryptoproto.SaltedHash{
			HashId:     "test-hash-1",
			Algorithm:  cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256,
			Salt:       make([]byte, 16),
			Hash:       make([]byte, 32),
			Iterations: 1000,
		}

		// Set the hash
		err := k.SetSaltedHash(ctx, hash)
		require.NoError(t, err)

		// Get and verify
		retrieved, err := k.GetSaltedHash(ctx, "test-hash-1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, "test-hash-1", retrieved.HashId)
		require.Equal(t, cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256, retrieved.Algorithm)
		require.Equal(t, int32(1000), retrieved.Iterations)

		// Delete the hash
		err = k.DeleteSaltedHash(ctx, "test-hash-1")
		require.NoError(t, err)

		// Verify it's gone
		_, err = k.GetSaltedHash(ctx, "test-hash-1")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidHashAlgorithm)
	})

	t.Run("set nil hash is no-op", func(t *testing.T) {
		err := k.SetSaltedHash(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("delete non-existent hash is idempotent", func(t *testing.T) {
		err := k.DeleteSaltedHash(ctx, "non-existent-hash")
		require.NoError(t, err)
	})

	t.Run("different hash algorithms", func(t *testing.T) {
		algorithms := []cryptoproto.HashAlgorithm{
			cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256,
			cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA512,
			cryptoproto.HashAlgorithm_HASH_ALGORITHM_BLAKE2B,
		}

		for _, algo := range algorithms {
			hash := &cryptoproto.SaltedHash{
				HashId:     "hash-" + algo.String(),
				Algorithm:  algo,
				Salt:       make([]byte, 16),
				Hash:       make([]byte, 32),
				Iterations: 500,
			}

			err := k.SetSaltedHash(ctx, hash)
			require.NoError(t, err)

			retrieved, err := k.GetSaltedHash(ctx, "hash-"+algo.String())
			require.NoError(t, err)
			require.Equal(t, algo, retrieved.Algorithm)

			err = k.DeleteSaltedHash(ctx, "hash-"+algo.String())
			require.NoError(t, err)
		}
	})
}

// TestHDKeyDerivationCRUD tests SetHDKeyDerivation, GetHDKeyDerivation, DeleteHDKeyDerivation
func TestHDKeyDerivationCRUD(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("not found case", func(t *testing.T) {
		_, err := k.GetHDKeyDerivation(ctx, "non-existent-master-key")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrHDKeyDerivationFailed)
	})

	t.Run("set, get, delete HD key derivation", func(t *testing.T) {
		derivation := &cryptoproto.HDKeyDerivation{
			MasterKeyId:    "master-key-1",
			DerivationPath: "m/44'/118'/0'/0/0",
			Depth:          5,
			SeedHash:       make([]byte, 32),
			ChainCode:      make([]byte, 32),
		}

		// Set the derivation
		err := k.SetHDKeyDerivation(ctx, derivation)
		require.NoError(t, err)

		// Get and verify
		retrieved, err := k.GetHDKeyDerivation(ctx, "master-key-1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Equal(t, "master-key-1", retrieved.MasterKeyId)
		require.Equal(t, "m/44'/118'/0'/0/0", retrieved.DerivationPath)
		require.Equal(t, int32(5), retrieved.Depth)

		// Delete the derivation
		err = k.DeleteHDKeyDerivation(ctx, "master-key-1")
		require.NoError(t, err)

		// Verify it's gone
		_, err = k.GetHDKeyDerivation(ctx, "master-key-1")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrHDKeyDerivationFailed)
	})

	t.Run("set nil derivation is no-op", func(t *testing.T) {
		err := k.SetHDKeyDerivation(ctx, nil)
		require.NoError(t, err)
	})

	t.Run("delete non-existent derivation is idempotent", func(t *testing.T) {
		err := k.DeleteHDKeyDerivation(ctx, "non-existent-master-key")
		require.NoError(t, err)
	})

	t.Run("multiple derivations with different paths", func(t *testing.T) {
		paths := []struct {
			masterKeyID string
			path        string
			depth       int32
		}{
			{"master-eth", "m/44'/60'/0'/0/0", 5},
			{"master-cosmos", "m/44'/118'/0'/0/0", 5},
			{"master-bitcoin", "m/44'/0'/0'/0/0", 5},
		}

		for _, p := range paths {
			derivation := &cryptoproto.HDKeyDerivation{
				MasterKeyId:    p.masterKeyID,
				DerivationPath: p.path,
				Depth:          p.depth,
				SeedHash:       make([]byte, 32),
				ChainCode:      make([]byte, 32),
			}

			err := k.SetHDKeyDerivation(ctx, derivation)
			require.NoError(t, err)

			retrieved, err := k.GetHDKeyDerivation(ctx, p.masterKeyID)
			require.NoError(t, err)
			require.Equal(t, p.path, retrieved.DerivationPath)
		}

		// Cleanup
		for _, p := range paths {
			err := k.DeleteHDKeyDerivation(ctx, p.masterKeyID)
			require.NoError(t, err)

			_, err = k.GetHDKeyDerivation(ctx, p.masterKeyID)
			require.Error(t, err)
		}
	})

	t.Run("update existing derivation", func(t *testing.T) {
		// Set initial derivation
		derivation := &cryptoproto.HDKeyDerivation{
			MasterKeyId:    "update-test",
			DerivationPath: "m/44'/118'/0'/0/0",
			Depth:          5,
			SeedHash:       make([]byte, 32),
			ChainCode:      make([]byte, 32),
		}
		err := k.SetHDKeyDerivation(ctx, derivation)
		require.NoError(t, err)

		// Update with new path
		derivation.DerivationPath = "m/44'/118'/0'/0/1"
		derivation.Depth = 6
		err = k.SetHDKeyDerivation(ctx, derivation)
		require.NoError(t, err)

		// Verify updated
		retrieved, err := k.GetHDKeyDerivation(ctx, "update-test")
		require.NoError(t, err)
		require.Equal(t, "m/44'/118'/0'/0/1", retrieved.DerivationPath)
		require.Equal(t, int32(6), retrieved.Depth)

		// Cleanup
		err = k.DeleteHDKeyDerivation(ctx, "update-test")
		require.NoError(t, err)
	})
}
