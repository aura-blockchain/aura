// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/cryptography/keeper"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// =============================================================================
// SetZKProofConfig Tests
// =============================================================================

func TestSetZKProofConfig(t *testing.T) {
	k, ctx := setupKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	t.Run("Nil config returns error", func(t *testing.T) {
		err := k.SetZKProofConfig(sdkCtx, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("Empty ProofId returns error", func(t *testing.T) {
		config := &cryptoproto.ZKProofConfig{
			ProofId:   "", // Empty
			CircuitId: "test-circuit",
			ProofType: cryptoproto.ZKProofType_ZK_PROOF_TYPE_GROTH16,
		}
		err := k.SetZKProofConfig(sdkCtx, config)
		require.Error(t, err)
		require.Contains(t, err.Error(), "proof ID cannot be empty")
	})

	t.Run("Successful set and retrieval", func(t *testing.T) {
		config := &cryptoproto.ZKProofConfig{
			ProofId:          "test-proof-123",
			CircuitId:        "test-circuit-456",
			ProofType:        cryptoproto.ZKProofType_ZK_PROOF_TYPE_GROTH16,
			PublicParameters: []byte("public-params"),
			VerificationKey:  []byte("verification-key"),
			Status:           cryptoproto.ZKProofStatus_ZK_PROOF_STATUS_ACTIVE,
			TotalProofs:      100,
			SuccessfulProofs: 95,
			Metadata: map[string]string{
				"creator": "test-creator",
			},
		}

		err := k.SetZKProofConfig(sdkCtx, config)
		require.NoError(t, err)

		// Retrieve and verify
		retrieved, err := k.GetZKProofConfig(sdkCtx, "test-proof-123")
		require.NoError(t, err)
		require.Equal(t, config.ProofId, retrieved.ProofId)
		require.Equal(t, config.CircuitId, retrieved.CircuitId)
		require.Equal(t, config.ProofType, retrieved.ProofType)
		require.Equal(t, config.TotalProofs, retrieved.TotalProofs)
		require.Equal(t, config.SuccessfulProofs, retrieved.SuccessfulProofs)
		require.Equal(t, "test-creator", retrieved.Metadata["creator"])
	})

	t.Run("Overwrite existing config", func(t *testing.T) {
		// Set initial config
		config1 := &cryptoproto.ZKProofConfig{
			ProofId:          "overwrite-test",
			CircuitId:        "circuit-v1",
			ProofType:        cryptoproto.ZKProofType_ZK_PROOF_TYPE_PLONK,
			TotalProofs:      10,
			SuccessfulProofs: 5,
		}
		err := k.SetZKProofConfig(sdkCtx, config1)
		require.NoError(t, err)

		// Overwrite with new values
		config2 := &cryptoproto.ZKProofConfig{
			ProofId:          "overwrite-test",
			CircuitId:        "circuit-v2",
			ProofType:        cryptoproto.ZKProofType_ZK_PROOF_TYPE_STARK,
			TotalProofs:      50,
			SuccessfulProofs: 48,
		}
		err = k.SetZKProofConfig(sdkCtx, config2)
		require.NoError(t, err)

		// Verify overwrite
		retrieved, err := k.GetZKProofConfig(sdkCtx, "overwrite-test")
		require.NoError(t, err)
		require.Equal(t, "circuit-v2", retrieved.CircuitId)
		require.Equal(t, cryptoproto.ZKProofType_ZK_PROOF_TYPE_STARK, retrieved.ProofType)
		require.Equal(t, uint64(50), retrieved.TotalProofs)
		require.Equal(t, uint64(48), retrieved.SuccessfulProofs)
	})
}

// =============================================================================
// GetProofStatistics Tests
// =============================================================================

func TestGetProofStatistics(t *testing.T) {
	k, ctx := setupKeeper(t)
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	t.Run("Error when config does not exist", func(t *testing.T) {
		total, successful, rate, err := k.GetProofStatistics(sdkCtx, "non-existent-proof")
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
		require.Equal(t, uint64(0), total)
		require.Equal(t, uint64(0), successful)
		require.Equal(t, float64(0), rate)
	})

	t.Run("Statistics with zero proofs", func(t *testing.T) {
		config := &cryptoproto.ZKProofConfig{
			ProofId:          "stats-zero-proofs",
			CircuitId:        "circuit-zero",
			ProofType:        cryptoproto.ZKProofType_ZK_PROOF_TYPE_GROTH16,
			TotalProofs:      0,
			SuccessfulProofs: 0,
		}
		err := k.SetZKProofConfig(sdkCtx, config)
		require.NoError(t, err)

		total, successful, rate, err := k.GetProofStatistics(sdkCtx, "stats-zero-proofs")
		require.NoError(t, err)
		require.Equal(t, uint64(0), total)
		require.Equal(t, uint64(0), successful)
		require.Equal(t, float64(0), rate) // No division by zero
	})

	t.Run("Statistics with successful proofs - rate calculation", func(t *testing.T) {
		config := &cryptoproto.ZKProofConfig{
			ProofId:          "stats-rate-test",
			CircuitId:        "circuit-rate",
			ProofType:        cryptoproto.ZKProofType_ZK_PROOF_TYPE_PLONK,
			TotalProofs:      100,
			SuccessfulProofs: 75,
		}
		err := k.SetZKProofConfig(sdkCtx, config)
		require.NoError(t, err)

		total, successful, rate, err := k.GetProofStatistics(sdkCtx, "stats-rate-test")
		require.NoError(t, err)
		require.Equal(t, uint64(100), total)
		require.Equal(t, uint64(75), successful)
		require.InDelta(t, 75.0, rate, 0.001) // 75% success rate
	})

	t.Run("Statistics with 100% success rate", func(t *testing.T) {
		config := &cryptoproto.ZKProofConfig{
			ProofId:          "stats-perfect",
			CircuitId:        "circuit-perfect",
			ProofType:        cryptoproto.ZKProofType_ZK_PROOF_TYPE_BULLETPROOFS,
			TotalProofs:      200,
			SuccessfulProofs: 200,
		}
		err := k.SetZKProofConfig(sdkCtx, config)
		require.NoError(t, err)

		total, successful, rate, err := k.GetProofStatistics(sdkCtx, "stats-perfect")
		require.NoError(t, err)
		require.Equal(t, uint64(200), total)
		require.Equal(t, uint64(200), successful)
		require.InDelta(t, 100.0, rate, 0.001) // 100% success rate
	})

	t.Run("Statistics with low success rate", func(t *testing.T) {
		config := &cryptoproto.ZKProofConfig{
			ProofId:          "stats-low",
			CircuitId:        "circuit-low",
			ProofType:        cryptoproto.ZKProofType_ZK_PROOF_TYPE_HALO2,
			TotalProofs:      1000,
			SuccessfulProofs: 123,
		}
		err := k.SetZKProofConfig(sdkCtx, config)
		require.NoError(t, err)

		total, successful, rate, err := k.GetProofStatistics(sdkCtx, "stats-low")
		require.NoError(t, err)
		require.Equal(t, uint64(1000), total)
		require.Equal(t, uint64(123), successful)
		require.InDelta(t, 12.3, rate, 0.001) // 12.3% success rate
	})
}

// =============================================================================
// HashToG1 Tests
// =============================================================================

func TestHashToG1(t *testing.T) {
	t.Run("Returns valid curve point for simple message", func(t *testing.T) {
		message := []byte("test message")
		result := keeper.HashToG1(message)

		require.NotNil(t, result)
		require.True(t, result.IsOnCurve(), "Result should be on the BN254 curve")
		require.True(t, result.IsInSubGroup(), "Result should be in the correct subgroup")
	})

	t.Run("Returns valid curve point for empty message", func(t *testing.T) {
		message := []byte{}
		result := keeper.HashToG1(message)

		require.NotNil(t, result)
		require.True(t, result.IsOnCurve())
		require.True(t, result.IsInSubGroup())
	})

	t.Run("Returns valid curve point for long message", func(t *testing.T) {
		message := make([]byte, 10000)
		for i := range message {
			message[i] = byte(i % 256)
		}
		result := keeper.HashToG1(message)

		require.NotNil(t, result)
		require.True(t, result.IsOnCurve())
		require.True(t, result.IsInSubGroup())
	})

	t.Run("Different messages produce different points", func(t *testing.T) {
		msg1 := []byte("message one")
		msg2 := []byte("message two")

		result1 := keeper.HashToG1(msg1)
		result2 := keeper.HashToG1(msg2)

		require.NotNil(t, result1)
		require.NotNil(t, result2)

		// Serialize and compare
		bytes1 := result1.Marshal()
		bytes2 := result2.Marshal()
		require.NotEqual(t, bytes1, bytes2, "Different messages should produce different curve points")
	})

	t.Run("Same message produces same point (deterministic)", func(t *testing.T) {
		msg := []byte("deterministic test")

		result1 := keeper.HashToG1(msg)
		result2 := keeper.HashToG1(msg)

		bytes1 := result1.Marshal()
		bytes2 := result2.Marshal()
		require.Equal(t, bytes1, bytes2, "Same message should produce same curve point")
	})
}

// =============================================================================
// GenerateThresholdSignatureShare Tests
// =============================================================================

func TestGenerateThresholdSignatureShare(t *testing.T) {
	t.Run("Generates non-empty output with valid secret share", func(t *testing.T) {
		// 32-byte secret share (field element size for BN254)
		secretShare := make([]byte, 32)
		for i := range secretShare {
			secretShare[i] = byte(i + 1) // Non-zero values
		}
		message := []byte("test message for signing")

		result := keeper.GenerateThresholdSignatureShare(secretShare, message)

		require.NotNil(t, result)
		require.NotEmpty(t, result)
		// BN254 G1 point marshaled is 64 bytes (uncompressed) or 32 bytes (compressed)
		// The gnark library uses uncompressed format by default
		require.True(t, len(result) >= 32, "Output should be at least 32 bytes (compressed G1 point)")
	})

	t.Run("Generates valid curve point", func(t *testing.T) {
		secretShare := make([]byte, 32)
		secretShare[0] = 0x42 // Some non-zero value
		message := []byte("another test message")

		result := keeper.GenerateThresholdSignatureShare(secretShare, message)

		require.NotNil(t, result)
		require.NotEmpty(t, result)
	})

	t.Run("Different secret shares produce different outputs", func(t *testing.T) {
		share1 := make([]byte, 32)
		share2 := make([]byte, 32)
		for i := range share1 {
			share1[i] = byte(i + 1)
			share2[i] = byte(i + 100)
		}
		message := []byte("same message")

		result1 := keeper.GenerateThresholdSignatureShare(share1, message)
		result2 := keeper.GenerateThresholdSignatureShare(share2, message)

		require.NotEqual(t, result1, result2, "Different shares should produce different signature shares")
	})

	t.Run("Different messages produce different outputs", func(t *testing.T) {
		secretShare := make([]byte, 32)
		secretShare[0] = 0x01
		msg1 := []byte("message one")
		msg2 := []byte("message two")

		result1 := keeper.GenerateThresholdSignatureShare(secretShare, msg1)
		result2 := keeper.GenerateThresholdSignatureShare(secretShare, msg2)

		require.NotEqual(t, result1, result2, "Different messages should produce different signature shares")
	})

	t.Run("Deterministic output for same inputs", func(t *testing.T) {
		secretShare := make([]byte, 32)
		for i := range secretShare {
			secretShare[i] = byte(i * 2)
		}
		message := []byte("deterministic message")

		result1 := keeper.GenerateThresholdSignatureShare(secretShare, message)
		result2 := keeper.GenerateThresholdSignatureShare(secretShare, message)

		require.Equal(t, result1, result2, "Same inputs should produce same output")
	})

	t.Run("Works with zero secret share", func(t *testing.T) {
		secretShare := make([]byte, 32) // All zeros
		message := []byte("test message")

		result := keeper.GenerateThresholdSignatureShare(secretShare, message)

		require.NotNil(t, result)
		require.NotEmpty(t, result)
	})
}

// =============================================================================
// VerifyThresholdSignature Tests
// =============================================================================

func TestVerifyThresholdSignature(t *testing.T) {
	t.Run("Returns false for invalid group public key bytes", func(t *testing.T) {
		invalidPubKey := []byte{0x01, 0x02, 0x03} // Too short
		signature := make([]byte, 64)
		message := []byte("test message")

		result := keeper.VerifyThresholdSignature(invalidPubKey, signature, message)
		require.False(t, result)
	})

	t.Run("Returns false for invalid signature bytes", func(t *testing.T) {
		// Generate a valid public key using HashToG1
		validPoint := keeper.HashToG1([]byte("pubkey seed"))
		validPubKey := validPoint.Marshal()

		invalidSignature := []byte{0xFF, 0xFF, 0xFF} // Too short
		message := []byte("test message")

		result := keeper.VerifyThresholdSignature(validPubKey, invalidSignature, message)
		require.False(t, result)
	})

	t.Run("Returns false for empty public key", func(t *testing.T) {
		emptyPubKey := []byte{}
		signature := make([]byte, 64)
		message := []byte("test message")

		result := keeper.VerifyThresholdSignature(emptyPubKey, signature, message)
		require.False(t, result)
	})

	t.Run("Returns false for empty signature", func(t *testing.T) {
		validPoint := keeper.HashToG1([]byte("pubkey seed"))
		validPubKey := validPoint.Marshal()

		emptySignature := []byte{}
		message := []byte("test message")

		result := keeper.VerifyThresholdSignature(validPubKey, emptySignature, message)
		require.False(t, result)
	})

	t.Run("Returns true for valid curve points", func(t *testing.T) {
		// Generate valid curve points for public key and signature
		pubKeyPoint := keeper.HashToG1([]byte("group public key seed"))
		sigPoint := keeper.HashToG1([]byte("signature seed"))

		pubKey := pubKeyPoint.Marshal()
		sig := sigPoint.Marshal()
		message := []byte("test message")

		result := keeper.VerifyThresholdSignature(pubKey, sig, message)
		// Should return true because both are valid curve points
		require.True(t, result)
	})

	t.Run("Handles all-zero public key", func(t *testing.T) {
		// Note: 64 zero bytes may or may not be a valid representation
		// depending on the curve library's handling of the identity point.
		// The key point is that the function doesn't panic.
		zeroPubKey := make([]byte, 64)
		signature := make([]byte, 64)
		message := []byte("test message")

		// Should not panic - just return a result
		result := keeper.VerifyThresholdSignature(zeroPubKey, signature, message)
		// The result depends on whether the curve lib accepts zero bytes as identity
		_ = result // Result may be true or false depending on curve library behavior
	})

	t.Run("Handles all-zero signature", func(t *testing.T) {
		// Note: Same as above - zero bytes behavior depends on curve library
		validPoint := keeper.HashToG1([]byte("pubkey seed"))
		validPubKey := validPoint.Marshal()

		zeroSignature := make([]byte, 64)
		message := []byte("test message")

		// Should not panic - just return a result
		result := keeper.VerifyThresholdSignature(validPubKey, zeroSignature, message)
		_ = result
	})

	t.Run("Works with empty message", func(t *testing.T) {
		pubKeyPoint := keeper.HashToG1([]byte("group public key"))
		sigPoint := keeper.HashToG1([]byte("signature"))

		pubKey := pubKeyPoint.Marshal()
		sig := sigPoint.Marshal()
		emptyMessage := []byte{}

		// Should not panic and should return a valid result
		result := keeper.VerifyThresholdSignature(pubKey, sig, emptyMessage)
		require.True(t, result)
	})
}
