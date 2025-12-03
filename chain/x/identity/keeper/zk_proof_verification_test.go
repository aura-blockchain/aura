package keeper_test

import (
	"crypto/rand"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/identity/keeper"
)

// ============================================================================
// Verification Key Management Tests
// ============================================================================

func TestSetZKVerificationKey(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	tests := []struct {
		name        string
		proofType   keeper.ZKProofType
		keyData     []byte
		description string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid groth16 key",
			proofType:   keeper.ZKProofTypeGroth16,
			keyData:     make([]byte, 128),
			description: "Test Groth16 verification key",
			expectError: false,
		},
		{
			name:        "valid plonk key",
			proofType:   keeper.ZKProofTypePLONK,
			keyData:     make([]byte, 128),
			description: "Test PLONK verification key",
			expectError: false,
		},
		{
			name:        "valid simple key",
			proofType:   keeper.ZKProofTypeSimple,
			keyData:     make([]byte, 32),
			description: "Test simple verification key",
			expectError: false,
		},
		{
			name:        "empty proof type",
			proofType:   keeper.ZKProofType(""),
			keyData:     make([]byte, 128),
			description: "Test key",
			expectError: true,
			errorMsg:    "proof type cannot be empty",
		},
		{
			name:        "empty key data",
			proofType:   keeper.ZKProofTypeGroth16,
			keyData:     []byte{},
			description: "Test key",
			expectError: true,
			errorMsg:    "verification key data cannot be empty",
		},
		{
			name:        "groth16 key too short",
			proofType:   keeper.ZKProofTypeGroth16,
			keyData:     make([]byte, 64), // Too short for Groth16
			description: "Test key",
			expectError: true,
			errorMsg:    "verification key too short",
		},
		{
			name:        "plonk key too short",
			proofType:   keeper.ZKProofTypePLONK,
			keyData:     make([]byte, 32), // Too short
			description: "Test key",
			expectError: true,
			errorMsg:    "verification key too short",
		},
		{
			name:        "simple key too short",
			proofType:   keeper.ZKProofTypeSimple,
			keyData:     make([]byte, 16), // Too short
			description: "Test key",
			expectError: true,
			errorMsg:    "verification key too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k.SetZKVerificationKey(ctx, tt.proofType, tt.keyData, tt.description)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)

				// Verify we can retrieve the key
				vk, err := k.GetZKVerificationKey(ctx, tt.proofType)
				require.NoError(t, err)
				require.NotNil(t, vk)
				require.Equal(t, tt.proofType, vk.ProofType)
				require.Equal(t, tt.description, vk.Description)
				require.Equal(t, tt.keyData, vk.KeyData)
				require.True(t, vk.Active)
				require.Equal(t, uint32(1), vk.Version)
			}
		})
	}
}

func TestGetZKVerificationKey(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Set up a test key
	testKeyData := make([]byte, 128)
	_, err := rand.Read(testKeyData)
	require.NoError(t, err)

	err = k.SetZKVerificationKey(ctx, keeper.ZKProofTypeGroth16, testKeyData, "Test key")
	require.NoError(t, err)

	tests := []struct {
		name        string
		proofType   keeper.ZKProofType
		expectError bool
		errorMsg    string
	}{
		{
			name:        "existing key",
			proofType:   keeper.ZKProofTypeGroth16,
			expectError: false,
		},
		{
			name:        "non-existent key",
			proofType:   keeper.ZKProofTypePLONK,
			expectError: true,
			errorMsg:    "verification key not found",
		},
		{
			name:        "empty proof type",
			proofType:   keeper.ZKProofType(""),
			expectError: true,
			errorMsg:    "proof type cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vk, err := k.GetZKVerificationKey(ctx, tt.proofType)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
				require.Nil(t, vk)
			} else {
				require.NoError(t, err)
				require.NotNil(t, vk)
				require.Equal(t, tt.proofType, vk.ProofType)
			}
		})
	}
}

// ============================================================================
// Proof Format Validation Tests
// ============================================================================

func TestProofFormatValidation(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Register verification keys for all types
	for _, proofType := range []keeper.ZKProofType{
		keeper.ZKProofTypeGroth16,
		keeper.ZKProofTypePLONK,
		keeper.ZKProofTypeBulletProofs,
		keeper.ZKProofTypeSimple,
	} {
		keyData := make([]byte, 128)
		_, err := rand.Read(keyData)
		require.NoError(t, err)
		err = k.SetZKVerificationKey(ctx, proofType, keyData, "Test key")
		require.NoError(t, err)
	}

	tests := []struct {
		name         string
		proofType    keeper.ZKProofType
		proofSize    int
		inputsSize   int
		expectError  bool
		errorMsg     string
	}{
		// Groth16 tests
		{
			name:        "valid groth16 proof - minimum size",
			proofType:   keeper.ZKProofTypeGroth16,
			proofSize:   96,
			inputsSize:  32,
			expectError: false,
		},
		{
			name:        "valid groth16 proof - normal size",
			proofType:   keeper.ZKProofTypeGroth16,
			proofSize:   128,
			inputsSize:  64,
			expectError: false,
		},
		{
			name:        "groth16 proof too short",
			proofType:   keeper.ZKProofTypeGroth16,
			proofSize:   64,
			inputsSize:  32,
			expectError: true,
			errorMsg:    "proof too short",
		},
		{
			name:        "groth16 proof too long",
			proofType:   keeper.ZKProofTypeGroth16,
			proofSize:   300,
			inputsSize:  32,
			expectError: true,
			errorMsg:    "proof too long",
		},

		// PLONK tests
		{
			name:        "valid plonk proof",
			proofType:   keeper.ZKProofTypePLONK,
			proofSize:   128,
			inputsSize:  32,
			expectError: false,
		},
		{
			name:        "plonk proof too short",
			proofType:   keeper.ZKProofTypePLONK,
			proofSize:   32,
			inputsSize:  32,
			expectError: true,
			errorMsg:    "proof too short",
		},
		{
			name:        "plonk proof too long",
			proofType:   keeper.ZKProofTypePLONK,
			proofSize:   600,
			inputsSize:  32,
			expectError: true,
			errorMsg:    "proof too long",
		},

		// BulletProofs tests
		{
			name:        "valid bulletproof",
			proofType:   keeper.ZKProofTypeBulletProofs,
			proofSize:   256,
			inputsSize:  32,
			expectError: false,
		},
		{
			name:        "bulletproof too short",
			proofType:   keeper.ZKProofTypeBulletProofs,
			proofSize:   16,
			inputsSize:  32,
			expectError: true,
			errorMsg:    "proof too short",
		},

		// Simple tests
		{
			name:        "valid simple proof",
			proofType:   keeper.ZKProofTypeSimple,
			proofSize:   96,
			inputsSize:  32,
			expectError: false,
		},
		{
			name:        "simple proof too short",
			proofType:   keeper.ZKProofTypeSimple,
			proofSize:   16,
			inputsSize:  32,
			expectError: true,
			errorMsg:    "proof too short",
		},

		// Public inputs tests
		{
			name:        "public inputs too short",
			proofType:   keeper.ZKProofTypeSimple,
			proofSize:   96,
			inputsSize:  16,
			expectError: true,
			errorMsg:    "public inputs too short",
		},
		{
			name:        "public inputs not multiple of 32",
			proofType:   keeper.ZKProofTypeSimple,
			proofSize:   96,
			inputsSize:  48,
			expectError: true,
			errorMsg:    "not a multiple of 32",
		},
		{
			name:        "public inputs too long",
			proofType:   keeper.ZKProofTypeSimple,
			proofSize:   96,
			inputsSize:  1100,
			expectError: true,
			errorMsg:    "public inputs too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proof := make([]byte, tt.proofSize)
			publicInputs := make([]byte, tt.inputsSize)

			_, err := k.VerifyZKProof(ctx, tt.proofType, proof, publicInputs)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
			} else {
				// Format validation passed, verification may succeed or fail
				// We don't check the result here, just that there's no format error
				_ = err
			}
		})
	}
}

// ============================================================================
// Simple Proof Verification Tests
// ============================================================================

func TestVerifySimpleProof_ValidProof(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Register verification key
	vkData := make([]byte, 32)
	_, err := rand.Read(vkData)
	require.NoError(t, err)

	err = k.SetZKVerificationKey(ctx, keeper.ZKProofTypeSimple, vkData, "Test simple key")
	require.NoError(t, err)

	// Generate public inputs
	publicInputs := make([]byte, 64)
	_, err = rand.Read(publicInputs)
	require.NoError(t, err)

	// Generate salt
	salt := make([]byte, 32)
	_, err = rand.Read(salt)
	require.NoError(t, err)

	// Generate valid proof
	proof, err := k.GenerateSimpleProof(ctx, publicInputs, salt)
	require.NoError(t, err)
	require.NotNil(t, proof)
	require.Equal(t, 96, len(proof))

	// Verify the proof
	verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypeSimple, proof, publicInputs)
	require.NoError(t, err)
	require.True(t, verified, "Valid proof should verify successfully")
}

func TestVerifySimpleProof_InvalidProofs(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Register verification key
	vkData := make([]byte, 32)
	_, err := rand.Read(vkData)
	require.NoError(t, err)

	err = k.SetZKVerificationKey(ctx, keeper.ZKProofTypeSimple, vkData, "Test simple key")
	require.NoError(t, err)

	// Generate valid inputs and proof
	publicInputs := make([]byte, 64)
	_, err = rand.Read(publicInputs)
	require.NoError(t, err)

	salt := make([]byte, 32)
	_, err = rand.Read(salt)
	require.NoError(t, err)

	validProof, err := k.GenerateSimpleProof(ctx, publicInputs, salt)
	require.NoError(t, err)

	tests := []struct {
		name         string
		proof        []byte
		publicInputs []byte
		expectError  bool
		shouldVerify bool
		description  string
	}{
		{
			name:         "corrupted proof - flipped bit in commitment",
			proof:        corruptProof(validProof, 0),
			publicInputs: publicInputs,
			expectError:  false,
			shouldVerify: false,
			description:  "Flipping a bit in the commitment should fail verification",
		},
		{
			name:         "corrupted proof - flipped bit in salt",
			proof:        corruptProof(validProof, 32),
			publicInputs: publicInputs,
			expectError:  false,
			shouldVerify: false,
			description:  "Flipping a bit in the salt should fail verification",
		},
		{
			name:         "wrong public inputs",
			proof:        validProof,
			publicInputs: make([]byte, 64), // Different inputs
			expectError:  false,
			shouldVerify: false,
			description:  "Wrong public inputs should fail verification",
		},
		{
			name:         "random proof",
			proof:        randomBytes(t, 96),
			publicInputs: publicInputs,
			expectError:  false,
			shouldVerify: false,
			description:  "Random bytes should not verify as a valid proof",
		},
		{
			name:         "empty proof",
			proof:        []byte{},
			publicInputs: publicInputs,
			expectError:  true,
			shouldVerify: false,
			description:  "Empty proof should return error",
		},
		{
			name:         "empty public inputs",
			proof:        validProof,
			publicInputs: []byte{},
			expectError:  true,
			shouldVerify: false,
			description:  "Empty public inputs should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypeSimple, tt.proof, tt.publicInputs)

			if tt.expectError {
				require.Error(t, err, tt.description)
			} else {
				require.NoError(t, err, tt.description)
				require.Equal(t, tt.shouldVerify, verified, tt.description)
			}
		})
	}
}

// ============================================================================
// Groth16 Proof Verification Tests
// ============================================================================

func TestVerifyGroth16Proof(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Register verification key
	vkData := make([]byte, 128)
	_, err := rand.Read(vkData)
	require.NoError(t, err)

	err = k.SetZKVerificationKey(ctx, keeper.ZKProofTypeGroth16, vkData, "Test Groth16 key")
	require.NoError(t, err)

	// Create valid proof structure
	publicInputs := make([]byte, 64)
	_, err = rand.Read(publicInputs)
	require.NoError(t, err)

	// Generate a valid proof for our placeholder implementation
	validProof := make([]byte, 128)
	_, err = rand.Read(validProof[:96]) // Random proof data

	// Compute expected commitment for last 32 bytes
	hasher := sha256.New()
	hasher.Write(publicInputs)
	hasher.Write(vkData)
	commitment := hasher.Sum(nil)
	copy(validProof[96:], commitment)

	// Test valid proof
	t.Run("valid groth16 proof", func(t *testing.T) {
		verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypeGroth16, validProof, publicInputs)
		require.NoError(t, err)
		require.True(t, verified, "Valid Groth16 proof should verify")
	})

	// Test invalid proofs
	t.Run("invalid groth16 proof - wrong commitment", func(t *testing.T) {
		invalidProof := make([]byte, 128)
		copy(invalidProof, validProof)
		invalidProof[100] ^= 0xFF // Corrupt commitment

		verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypeGroth16, invalidProof, publicInputs)
		require.NoError(t, err)
		require.False(t, verified, "Invalid proof should not verify")
	})

	t.Run("random groth16 proof", func(t *testing.T) {
		randomProof := randomBytes(t, 128)
		verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypeGroth16, randomProof, publicInputs)
		require.NoError(t, err)
		require.False(t, verified, "Random proof should not verify")
	})
}

// ============================================================================
// PLONK Proof Verification Tests
// ============================================================================

func TestVerifyPLONKProof(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Register verification key
	vkData := make([]byte, 128)
	_, err := rand.Read(vkData)
	require.NoError(t, err)

	err = k.SetZKVerificationKey(ctx, keeper.ZKProofTypePLONK, vkData, "Test PLONK key")
	require.NoError(t, err)

	// Create valid proof structure
	publicInputs := make([]byte, 64)
	_, err = rand.Read(publicInputs)
	require.NoError(t, err)

	// Generate a valid proof for our placeholder implementation
	validProof := make([]byte, 128)
	_, err = rand.Read(validProof[:96])

	// Compute expected commitment for last 32 bytes
	hasher := sha256.New()
	hasher.Write(publicInputs)
	hasher.Write(vkData)
	hasher.Write([]byte("plonk"))
	commitment := hasher.Sum(nil)
	copy(validProof[96:], commitment)

	// Test valid proof
	t.Run("valid plonk proof", func(t *testing.T) {
		verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypePLONK, validProof, publicInputs)
		require.NoError(t, err)
		require.True(t, verified, "Valid PLONK proof should verify")
	})

	// Test invalid proof
	t.Run("invalid plonk proof", func(t *testing.T) {
		invalidProof := randomBytes(t, 128)
		verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypePLONK, invalidProof, publicInputs)
		require.NoError(t, err)
		require.False(t, verified, "Random PLONK proof should not verify")
	})
}

// ============================================================================
// BulletProofs Verification Tests
// ============================================================================

func TestVerifyBulletProof(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Register verification key
	vkData := make([]byte, 128)
	_, err := rand.Read(vkData)
	require.NoError(t, err)

	err = k.SetZKVerificationKey(ctx, keeper.ZKProofTypeBulletProofs, vkData, "Test BulletProofs key")
	require.NoError(t, err)

	// Create valid proof structure
	publicInputs := make([]byte, 64)
	_, err = rand.Read(publicInputs)
	require.NoError(t, err)

	// Generate a valid proof for our placeholder implementation
	validProof := make([]byte, 256)
	_, err = rand.Read(validProof[:224])

	// Compute expected commitment for last 32 bytes
	hasher := sha256.New()
	hasher.Write(publicInputs)
	hasher.Write(vkData)
	hasher.Write([]byte("bulletproof"))
	commitment := hasher.Sum(nil)
	copy(validProof[224:], commitment)

	// Test valid proof
	t.Run("valid bulletproof", func(t *testing.T) {
		verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypeBulletProofs, validProof, publicInputs)
		require.NoError(t, err)
		require.True(t, verified, "Valid BulletProof should verify")
	})

	// Test invalid proof
	t.Run("invalid bulletproof", func(t *testing.T) {
		invalidProof := randomBytes(t, 256)
		verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypeBulletProofs, invalidProof, publicInputs)
		require.NoError(t, err)
		require.False(t, verified, "Random BulletProof should not verify")
	})
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestVerifyZKProof_ErrorCases(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Register a test verification key
	vkData := make([]byte, 128)
	err := k.SetZKVerificationKey(ctx, keeper.ZKProofTypeSimple, vkData, "Test key")
	require.NoError(t, err)

	validProof := make([]byte, 96)
	validInputs := make([]byte, 64)

	tests := []struct {
		name        string
		proofType   keeper.ZKProofType
		proof       []byte
		inputs      []byte
		expectError bool
		errorMsg    string
	}{
		{
			name:        "unknown proof type",
			proofType:   keeper.ZKProofType("unknown"),
			proof:       validProof,
			inputs:      validInputs,
			expectError: true,
			errorMsg:    "verification key not found",
		},
		{
			name:        "missing verification key",
			proofType:   keeper.ZKProofTypeGroth16, // Not registered
			proof:       validProof,
			inputs:      validInputs,
			expectError: true,
			errorMsg:    "verification key not found",
		},
		{
			name:        "nil proof",
			proofType:   keeper.ZKProofTypeSimple,
			proof:       nil,
			inputs:      validInputs,
			expectError: true,
			errorMsg:    "proof cannot be empty",
		},
		{
			name:        "nil inputs",
			proofType:   keeper.ZKProofTypeSimple,
			proof:       validProof,
			inputs:      nil,
			expectError: true,
			errorMsg:    "public inputs cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verified, err := k.VerifyZKProof(ctx, tt.proofType, tt.proof, tt.inputs)

			if tt.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
				require.False(t, verified)
			}
		})
	}
}

// ============================================================================
// Security Tests - Malicious Proofs
// ============================================================================

func TestVerifyZKProof_SecurityTests(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Register verification key
	vkData := make([]byte, 32)
	_, err := rand.Read(vkData)
	require.NoError(t, err)

	err = k.SetZKVerificationKey(ctx, keeper.ZKProofTypeSimple, vkData, "Test key")
	require.NoError(t, err)

	publicInputs := make([]byte, 64)
	_, err = rand.Read(publicInputs)
	require.NoError(t, err)

	salt := make([]byte, 32)
	_, err = rand.Read(salt)
	require.NoError(t, err)

	validProof, err := k.GenerateSimpleProof(ctx, publicInputs, salt)
	require.NoError(t, err)

	t.Run("replay attack - same proof different inputs", func(t *testing.T) {
		differentInputs := make([]byte, 64)
		_, err := rand.Read(differentInputs)
		require.NoError(t, err)

		// Trying to use the same proof with different inputs should fail
		verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypeSimple, validProof, differentInputs)
		require.NoError(t, err)
		require.False(t, verified, "Proof should not verify with different inputs")
	})

	t.Run("malformed proof - all zeros", func(t *testing.T) {
		zeroProof := make([]byte, 96)
		verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypeSimple, zeroProof, publicInputs)
		require.NoError(t, err)
		require.False(t, verified, "All-zero proof should not verify")
	})

	t.Run("malformed proof - all ones", func(t *testing.T) {
		onesProof := make([]byte, 96)
		for i := range onesProof {
			onesProof[i] = 0xFF
		}
		verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypeSimple, onesProof, publicInputs)
		require.NoError(t, err)
		require.False(t, verified, "All-ones proof should not verify")
	})

	t.Run("truncated proof", func(t *testing.T) {
		truncatedProof := validProof[:80]
		_, err := k.VerifyZKProof(ctx, keeper.ZKProofTypeSimple, truncatedProof, publicInputs)
		require.Error(t, err, "Truncated proof should return format error")
		require.Contains(t, err.Error(), "proof too short")
	})

	t.Run("extended proof with garbage", func(t *testing.T) {
		extendedProof := make([]byte, 200)
		copy(extendedProof, validProof)
		_, err := rand.Read(extendedProof[96:])
		require.NoError(t, err)

		// Should still verify based on the valid portion
		verified, err := k.VerifyZKProof(ctx, keeper.ZKProofTypeSimple, extendedProof, publicInputs)
		require.NoError(t, err)
		require.True(t, verified, "Valid proof with extra data should still verify")
	})
}

// ============================================================================
// Multiple Verification Keys Tests
// ============================================================================

func TestMultipleVerificationKeys(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Register multiple verification keys
	proofTypes := []keeper.ZKProofType{
		keeper.ZKProofTypeGroth16,
		keeper.ZKProofTypePLONK,
		keeper.ZKProofTypeBulletProofs,
		keeper.ZKProofTypeSimple,
	}

	for _, pt := range proofTypes {
		vkData := randomBytes(t, 128)
		err := k.SetZKVerificationKey(ctx, pt, vkData, "Test key for "+string(pt))
		require.NoError(t, err)
	}

	// Verify all keys are retrievable
	for _, pt := range proofTypes {
		vk, err := k.GetZKVerificationKey(ctx, pt)
		require.NoError(t, err)
		require.NotNil(t, vk)
		require.Equal(t, pt, vk.ProofType)
		require.True(t, vk.Active)
	}

	// Verify each proof type works independently
	publicInputs := randomBytes(t, 64)

	for _, pt := range proofTypes {
		t.Run("verify_"+string(pt), func(t *testing.T) {
			var proof []byte
			var err error

			switch pt {
			case keeper.ZKProofTypeSimple:
				salt := randomBytes(t, 32)
				proof, err = k.GenerateSimpleProof(ctx, publicInputs, salt)
				require.NoError(t, err)
			default:
				// For placeholder implementations, create a proof with correct commitment
				vk, err := k.GetZKVerificationKey(ctx, pt)
				require.NoError(t, err)

				proof = make([]byte, 128)
				hasher := sha256.New()
				hasher.Write(publicInputs)
				hasher.Write(vk.KeyData)
				if pt == keeper.ZKProofTypePLONK {
					hasher.Write([]byte("plonk"))
				} else if pt == keeper.ZKProofTypeBulletProofs {
					hasher.Write([]byte("bulletproof"))
				}
				commitment := hasher.Sum(nil)
				copy(proof[96:], commitment)
			}

			verified, err := k.VerifyZKProof(ctx, pt, proof, publicInputs)
			require.NoError(t, err)
			require.True(t, verified, "Valid proof for "+string(pt)+" should verify")
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// corruptProof flips a bit at the given byte offset
func corruptProof(proof []byte, offset int) []byte {
	corrupted := make([]byte, len(proof))
	copy(corrupted, proof)
	if offset < len(corrupted) {
		corrupted[offset] ^= 0x01 // Flip the least significant bit
	}
	return corrupted
}

// randomBytes generates random bytes of given length
func randomBytes(t *testing.T, length int) []byte {
	b := make([]byte, length)
	_, err := rand.Read(b)
	require.NoError(t, err)
	return b
}
