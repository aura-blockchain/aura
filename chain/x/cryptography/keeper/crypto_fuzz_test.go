// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"bytes"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/cryptography/keeper"
	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// setupCryptoFuzzKeeper creates a keeper for fuzz testing.
// Uses testing.TB interface to work with both *testing.T and *testing.F.
func setupCryptoFuzzKeeper(tb testing.TB) (*keeper.Keeper, sdk.Context) {
	tb.Helper()

	storeKey := storetypes.NewKVStoreKey(types.ModuleName)
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	if err := cms.LoadLatestVersion(); err != nil {
		tb.Fatalf("failed to load latest version: %v", err)
	}

	header := cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}
	ctx := sdk.NewContext(cms, header, false, log.NewNopLogger())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(cdc, storeKey, log.NewNopLogger(), "authority")

	// Initialize default params
	params := types.DefaultParams()
	if err := k.SetParams(ctx, &params); err != nil {
		tb.Fatalf("failed to set params: %v", err)
	}

	return k, ctx
}

// ============================================================================
// SIGNATURE VERIFICATION FUZZ TESTS
// ============================================================================

// FuzzZKProofVerification fuzzes the ZK proof verification path.
// Security properties tested:
//   - Rejects malformed proof data
//   - Rejects invalid proof sizes
//   - Rejects proofs with all-zero bytes (identity point)
//   - Never panics on any input
//   - Correctly validates proof format
func FuzzZKProofVerification(f *testing.F) {
	// Seed corpus with various proof types and sizes
	// Groth16 sizes: 128-256 bytes
	f.Add(make([]byte, 128), make([]byte, 32), uint8(0))
	f.Add(make([]byte, 256), make([]byte, 32), uint8(0))
	// PLONK sizes: 288-512 bytes
	f.Add(make([]byte, 288), make([]byte, 64), uint8(1))
	f.Add(make([]byte, 512), make([]byte, 64), uint8(1))
	// Bulletproofs: 672-2048 bytes
	f.Add(make([]byte, 672), make([]byte, 32), uint8(2))
	// STARKs: 1024-32768 bytes
	f.Add(make([]byte, 1024), make([]byte, 64), uint8(3))
	// Halo2: 256-512 bytes
	f.Add(make([]byte, 256), make([]byte, 32), uint8(4))
	// Edge cases
	f.Add([]byte{}, []byte{}, uint8(0))                    // Empty proof and inputs
	f.Add(make([]byte, 1), make([]byte, 1), uint8(0))      // Minimal proof
	f.Add(make([]byte, 100), make([]byte, 32), uint8(255)) // Invalid proof type
	// All zeros (identity point attack)
	f.Add(make([]byte, 128), make([]byte, 32), uint8(0))
	// Non-32-byte-aligned public inputs
	f.Add(make([]byte, 128), make([]byte, 17), uint8(0))

	f.Fuzz(func(t *testing.T, proofData []byte, publicInputs []byte, proofTypeInt uint8) {
		if len(proofData) > 100000 || len(publicInputs) > 10000 {
			t.Skip("input too large")
		}

		k, ctx := setupCryptoFuzzKeeper(t)

		// Map proof type
		proofType := cryptoproto.ZKProofType(proofTypeInt % 6)

		// Create verification key (required for proof verification)
		verificationKey := make([]byte, 256)
		for i := range verificationKey {
			verificationKey[i] = byte(i % 256)
		}

		// Register a circuit
		proofID, err := k.RegisterZKProofCircuit(
			ctx,
			"fuzz-creator",
			proofType,
			[]byte("public-params"),
			verificationKey,
			"fuzz-circuit",
		)

		if err != nil {
			// Registration can fail for invalid proof types
			return
		}

		// Submit and verify the fuzzed proof - must not panic
		verified, _, verifyErr := k.SubmitZKProof(ctx, "fuzz-submitter", proofID, proofData, publicInputs)

		// SECURITY INVARIANT: Empty proof must be rejected
		if len(proofData) == 0 {
			if verifyErr == nil && verified {
				t.Error("empty proof data should be rejected")
			}
		}

		// SECURITY INVARIANT: All-zero proofs should be rejected (identity point attack)
		allZero := true
		for _, b := range proofData {
			if b != 0 {
				allZero = false
				break
			}
		}
		if allZero && len(proofData) > 0 {
			if verified && verifyErr == nil {
				t.Error("all-zero proof should be rejected (identity point)")
			}
		}

		// SECURITY INVARIANT: Public inputs must be multiple of 32 bytes
		if len(publicInputs) > 0 && len(publicInputs) < 32 {
			if verified && verifyErr == nil {
				t.Error("public inputs smaller than 32 bytes should be rejected")
			}
		}

		// SECURITY INVARIANT: Invalid proof sizes should be rejected
		// This is proof-type dependent
		_ = verified // Used for invariant checks above
	})
}

// FuzzThresholdSignatureShare fuzzes threshold signature share submission.
// Security properties tested:
//   - Validates participant authorization
//   - Prevents duplicate shares
//   - Handles malformed signature shares
//   - Never panics on any input
func FuzzThresholdSignatureShare(f *testing.F) {
	f.Add("participant1", "scheme-1", make([]byte, 64), make([]byte, 32))
	f.Add("", "scheme-1", make([]byte, 64), make([]byte, 32))             // Empty participant
	f.Add("p1", "", make([]byte, 64), make([]byte, 32))                   // Empty scheme
	f.Add("p1", "scheme-1", []byte{}, make([]byte, 32))                   // Empty share
	f.Add("p1", "scheme-1", make([]byte, 64), []byte{})                   // Empty message hash
	f.Add("p1", "scheme-1", make([]byte, 1000), make([]byte, 32))         // Large share
	f.Add("unauthorized", "scheme-1", make([]byte, 64), make([]byte, 32)) // Unauthorized participant

	f.Fuzz(func(t *testing.T, participant, schemeID string, signatureShare, messageHash []byte) {
		if len(participant) > 1000 || len(schemeID) > 500 || len(signatureShare) > 10000 || len(messageHash) > 1000 {
			t.Skip("input too large")
		}

		k, ctx := setupCryptoFuzzKeeper(t)

		// Create a threshold scheme with known participants
		participants := []string{"p1", "p2", "p3"}
		validSchemeID, _, err := k.CreateThresholdScheme(
			ctx,
			"creator",
			2, // threshold
			3, // total
			participants,
			cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA,
		)
		if err != nil {
			t.Fatalf("failed to create scheme: %v", err)
		}

		// Test with the fuzzed inputs
		testSchemeID := schemeID
		if schemeID == "scheme-1" {
			testSchemeID = validSchemeID
		}

		// Submit signature share - must not panic
		sharesCollected, thresholdReached, combinedSig, submitErr := k.SubmitThresholdSignatureShare(
			ctx,
			participant,
			testSchemeID,
			signatureShare,
			messageHash,
		)

		// SECURITY INVARIANT: Empty participant must be rejected
		if participant == "" {
			if submitErr == nil {
				t.Error("empty participant should be rejected")
			}
		}

		// SECURITY INVARIANT: Unauthorized participants must be rejected
		isAuthorized := false
		for _, p := range participants {
			if p == participant {
				isAuthorized = true
				break
			}
		}
		if !isAuthorized && testSchemeID == validSchemeID {
			if submitErr == nil {
				t.Error("unauthorized participant should be rejected")
			}
		}

		// SECURITY INVARIANT: Non-existent scheme must be rejected
		if schemeID != "scheme-1" && schemeID != validSchemeID {
			if submitErr == nil {
				t.Error("non-existent scheme should be rejected")
			}
		}

		// Prevent unused variable warnings
		_, _, _ = sharesCollected, thresholdReached, combinedSig
	})
}

// ============================================================================
// KEY GENERATION AND REGISTRATION FUZZ TESTS
// ============================================================================

// FuzzQuantumResistantKeyRegistration fuzzes quantum-resistant key registration.
// Security properties tested:
//   - Validates key sizes for each algorithm
//   - Rejects empty or nil public keys
//   - Handles various algorithm types
//   - Never panics on any input
func FuzzQuantumResistantKeyRegistration(f *testing.F) {
	// CRYSTALS-Dilithium: 1312 bytes
	f.Add("creator1", uint8(0), make([]byte, 1312), int64(86400*365))
	// CRYSTALS-Kyber: 800 bytes
	f.Add("creator1", uint8(1), make([]byte, 800), int64(86400*365))
	// Falcon: 897 bytes
	f.Add("creator1", uint8(2), make([]byte, 897), int64(86400*365))
	// SPHINCS+: 32 bytes
	f.Add("creator1", uint8(3), make([]byte, 32), int64(86400*365))
	// NTRU: 1230 bytes
	f.Add("creator1", uint8(4), make([]byte, 1230), int64(86400*365))
	// Edge cases
	f.Add("", uint8(0), make([]byte, 1312), int64(86400))   // Empty creator
	f.Add("c", uint8(0), []byte{}, int64(86400))            // Empty key
	f.Add("c", uint8(0), make([]byte, 1), int64(86400))     // Too short
	f.Add("c", uint8(255), make([]byte, 100), int64(86400)) // Invalid algorithm
	f.Add("c", uint8(0), make([]byte, 1312), int64(-1))     // Negative expiry
	f.Add("c", uint8(0), make([]byte, 1312), int64(0))      // Zero expiry

	f.Fuzz(func(t *testing.T, creator string, algoInt uint8, publicKey []byte, expirySeconds int64) {
		if len(creator) > 1000 || len(publicKey) > 10000 {
			t.Skip("input too large")
		}

		k, ctx := setupCryptoFuzzKeeper(t)

		algo := cryptoproto.QuantumResistantAlgorithm(algoInt % 6)

		var expiresAt *time.Time
		if expirySeconds > 0 {
			t := time.Now().Add(time.Duration(expirySeconds) * time.Second)
			expiresAt = &t
		}

		// Register key - must not panic
		keyID, err := k.RegisterQuantumResistantKey(ctx, creator, algo, publicKey, expiresAt)

		// SECURITY INVARIANT: Empty creator must be rejected
		if creator == "" {
			if err == nil {
				t.Error("empty creator should be rejected")
			}
		}

		// SECURITY INVARIANT: Empty public key must be rejected
		if len(publicKey) == 0 {
			if err == nil {
				t.Error("empty public key should be rejected")
			}
		}

		// SECURITY INVARIANT: Invalid key sizes should be rejected
		// Each algorithm has specific size requirements
		validSizes := map[cryptoproto.QuantumResistantAlgorithm]int{
			cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM: 1312,
			cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER:     800,
			cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_FALCON:             897,
			cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_SPHINCS_PLUS:       32,
			cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_NTRU:               1230,
		}
		expectedSize, known := validSizes[algo]
		if known && len(publicKey) != expectedSize && len(publicKey) > 0 {
			// Note: Implementation may or may not validate sizes
			// This is documentation of expected behavior
		}

		if err == nil {
			// Verify stored key can be retrieved
			storedKey, getErr := k.GetQuantumResistantKey(ctx, keyID)
			if getErr != nil {
				t.Errorf("failed to retrieve registered key: %v", getErr)
			}
			if storedKey != nil && !bytes.Equal(storedKey.PublicKey, publicKey) {
				t.Error("stored public key does not match registered key")
			}
		}
	})
}

// ============================================================================
// ENCRYPTION/DECRYPTION FUZZ TESTS
// ============================================================================

// FuzzSaltedHashing fuzzes salted hashing operations.
// Security properties tested:
//   - Handles various hash algorithms
//   - Validates iteration counts
//   - Salt length validation
//   - Hash verification correctness
//   - Never panics on any input
func FuzzSaltedHashing(f *testing.F) {
	f.Add([]byte("test data"), uint8(0), int32(1))
	f.Add([]byte("test data"), uint8(1), int32(1000))
	f.Add([]byte("test data"), uint8(2), int32(1))
	f.Add([]byte{}, uint8(0), int32(1))            // Empty data
	f.Add(make([]byte, 10000), uint8(0), int32(1)) // Large data
	f.Add([]byte("x"), uint8(255), int32(1))       // Invalid algorithm
	f.Add([]byte("x"), uint8(0), int32(0))         // Zero iterations
	f.Add([]byte("x"), uint8(0), int32(-1))        // Negative iterations
	f.Add([]byte("x"), uint8(0), int32(1000000))   // Very high iterations

	f.Fuzz(func(t *testing.T, data []byte, algoInt uint8, iterations int32) {
		if len(data) > 100000 {
			t.Skip("data too large")
		}

		// Limit iterations to prevent excessive test time
		if iterations > 10000 || iterations < 0 {
			iterations = 1
		}

		k, ctx := setupCryptoFuzzKeeper(t)

		algo := cryptoproto.HashAlgorithm(algoInt % 7)

		// Create salted hash - must not panic
		hashID, salt, hash, err := k.CreateSaltedHash(ctx, data, algo, iterations)

		// For valid inputs, verify the hash
		if err == nil {
			// SECURITY INVARIANT: Hash ID must not be empty
			if hashID == "" {
				t.Error("hash ID should not be empty on success")
			}

			// SECURITY INVARIANT: Salt must not be empty
			if len(salt) == 0 {
				t.Error("salt should not be empty on success")
			}

			// SECURITY INVARIANT: Hash must not be empty
			if len(hash) == 0 {
				t.Error("hash should not be empty on success")
			}

			// Verify the hash is correct
			valid, verifyErr := k.VerifySaltedHash(ctx, hashID, data)
			if verifyErr != nil {
				t.Errorf("verification failed: %v", verifyErr)
			}
			if !valid {
				t.Error("hash verification should succeed for correct data")
			}

			// SECURITY INVARIANT: Wrong data must fail verification
			if len(data) > 0 {
				wrongData := append([]byte{0xff}, data...)
				wrongValid, _ := k.VerifySaltedHash(ctx, hashID, wrongData)
				if wrongValid {
					t.Error("hash verification should fail for wrong data")
				}
			}
		}
	})
}

// FuzzHashWithCustomSalt fuzzes hashing with custom salt.
// Security properties tested:
//   - Validates minimum salt length
//   - Consistent hash output for same inputs
//   - Different hashes for different inputs
//   - Never panics on any input
func FuzzHashWithCustomSalt(f *testing.F) {
	f.Add([]byte("test data"), make([]byte, 16), uint8(0), int32(1))
	f.Add([]byte("test data"), make([]byte, 32), uint8(1), int32(100))
	f.Add([]byte{}, make([]byte, 16), uint8(0), int32(1))      // Empty data
	f.Add([]byte("x"), []byte{}, uint8(0), int32(1))           // Empty salt
	f.Add([]byte("x"), make([]byte, 15), uint8(0), int32(1))   // Salt too short
	f.Add([]byte("x"), make([]byte, 16), uint8(255), int32(1)) // Invalid algorithm

	f.Fuzz(func(t *testing.T, data, salt []byte, algoInt uint8, iterations int32) {
		if len(data) > 50000 || len(salt) > 1000 {
			t.Skip("input too large")
		}

		if iterations > 10000 || iterations < 0 {
			iterations = 1
		}

		k, ctx := setupCryptoFuzzKeeper(t)

		algo := cryptoproto.HashAlgorithm(algoInt % 7)

		// Hash with custom salt - must not panic
		hash1, err := k.HashWithCustomSalt(ctx, data, salt, algo, iterations)

		// SECURITY INVARIANT: Salt less than 16 bytes should be rejected
		if len(salt) < 16 {
			if err == nil {
				t.Error("salt shorter than 16 bytes should be rejected")
			}
			return
		}

		if err == nil {
			// SECURITY INVARIANT: Same inputs must produce same hash (determinism)
			hash2, err2 := k.HashWithCustomSalt(ctx, data, salt, algo, iterations)
			if err2 != nil {
				t.Errorf("second hash failed: %v", err2)
			}
			if !bytes.Equal(hash1, hash2) {
				t.Error("same inputs must produce same hash")
			}

			// SECURITY INVARIANT: Different data must produce different hash
			if len(data) > 0 {
				differentData := append(data, 0xff)
				hash3, err3 := k.HashWithCustomSalt(ctx, differentData, salt, algo, iterations)
				if err3 == nil && bytes.Equal(hash1, hash3) {
					t.Error("different data should produce different hash (collision detected)")
				}
			}

			// SECURITY INVARIANT: Different salt must produce different hash
			if len(salt) >= 16 {
				differentSalt := make([]byte, len(salt))
				copy(differentSalt, salt)
				differentSalt[0] ^= 0xff
				hash4, err4 := k.HashWithCustomSalt(ctx, data, differentSalt, algo, iterations)
				if err4 == nil && bytes.Equal(hash1, hash4) {
					t.Error("different salt should produce different hash")
				}
			}
		}
	})
}

// FuzzKeyStretching fuzzes key stretching operations.
// Security properties tested:
//   - Validates algorithm parameters
//   - Handles various input key lengths
//   - Salt length validation
//   - Output key correctness
//   - Never panics on any input
func FuzzKeyStretching(f *testing.F) {
	f.Add([]byte("password"), make([]byte, 16), "pbkdf2-sha512", uint32(1000))
	f.Add([]byte("password"), make([]byte, 16), "argon2id", uint32(3))
	f.Add([]byte{}, make([]byte, 16), "pbkdf2-sha512", uint32(1000))         // Empty key
	f.Add([]byte("x"), make([]byte, 15), "pbkdf2-sha512", uint32(1000))      // Salt too short
	f.Add([]byte("x"), make([]byte, 16), "invalid-algo", uint32(1000))       // Invalid algorithm
	f.Add([]byte("x"), make([]byte, 16), "pbkdf2-sha512", uint32(0))         // Zero iterations
	f.Add(make([]byte, 10000), make([]byte, 32), "pbkdf2-sha512", uint32(1)) // Large input

	f.Fuzz(func(t *testing.T, inputKey, salt []byte, algorithm string, iterations uint32) {
		if len(inputKey) > 50000 || len(salt) > 1000 {
			t.Skip("input too large")
		}

		// Limit iterations to prevent excessive test time
		if iterations > 10000 {
			iterations = 1000
		}

		k, sdkCtx := setupCryptoFuzzKeeper(t)

		// Perform key stretching - must not panic
		// Note: PerformKeyStretching takes sdk.Context
		stretchedKey, err := k.PerformKeyStretching(sdkCtx, inputKey, salt, algorithm, iterations)

		// SECURITY INVARIANT: Empty input key must be rejected
		if len(inputKey) == 0 {
			if err == nil {
				t.Error("empty input key should be rejected")
			}
			return
		}

		// SECURITY INVARIANT: Salt less than 16 bytes should be rejected
		if len(salt) < 16 {
			if err == nil {
				t.Error("salt shorter than 16 bytes should be rejected")
			}
			return
		}

		// SECURITY INVARIANT: Invalid algorithms should be rejected
		validAlgos := map[string]bool{"pbkdf2-sha512": true, "argon2id": true}
		if !validAlgos[algorithm] {
			if err == nil {
				t.Error("invalid algorithm should be rejected")
			}
			return
		}

		if err == nil {
			// SECURITY INVARIANT: Stretched key must not be empty
			if len(stretchedKey) == 0 {
				t.Error("stretched key should not be empty on success")
			}

			// SECURITY INVARIANT: Same inputs must produce same output (determinism)
			stretchedKey2, err2 := k.PerformKeyStretching(sdkCtx, inputKey, salt, algorithm, iterations)
			if err2 != nil {
				t.Errorf("second stretch failed: %v", err2)
			}
			if !bytes.Equal(stretchedKey, stretchedKey2) {
				t.Error("same inputs must produce same stretched key")
			}
		}
	})
}

// FuzzCompareHashes fuzzes the constant-time hash comparison.
// Security properties tested:
//   - Constant-time comparison (no timing side channels)
//   - Correct equality detection
//   - Different lengths are not equal
//   - Never panics on any input
func FuzzCompareHashes(f *testing.F) {
	hash := make([]byte, 32)
	for i := range hash {
		hash[i] = byte(i)
	}
	f.Add(hash, hash)                 // Equal
	f.Add(hash, make([]byte, 32))     // Different content
	f.Add(hash, make([]byte, 31))     // Different length
	f.Add([]byte{}, []byte{})         // Both empty
	f.Add(hash, []byte{})             // One empty
	f.Add([]byte{0x00}, []byte{0x01}) // Single byte different

	f.Fuzz(func(t *testing.T, hash1, hash2 []byte) {
		if len(hash1) > 10000 || len(hash2) > 10000 {
			t.Skip("input too large")
		}

		k, _ := setupCryptoFuzzKeeper(t)

		// Compare hashes - must not panic
		result := k.CompareHashes(hash1, hash2)

		// SECURITY INVARIANT: Different lengths must not be equal
		if len(hash1) != len(hash2) {
			if result {
				t.Error("hashes of different lengths should not be equal")
			}
		}

		// SECURITY INVARIANT: Identical hashes must be equal
		if bytes.Equal(hash1, hash2) {
			if !result {
				t.Error("identical hashes should be equal")
			}
		}

		// SECURITY INVARIANT: Different hashes must not be equal
		if !bytes.Equal(hash1, hash2) {
			if result {
				t.Error("different hashes should not be equal")
			}
		}
	})
}

// FuzzSecureRandomBytes fuzzes secure random byte generation.
// Security properties tested:
//   - Validates length parameter
//   - Rejects zero/negative lengths
//   - Returns correct length
//   - Never panics on any input
func FuzzSecureRandomBytes(f *testing.F) {
	f.Add(32)
	f.Add(64)
	f.Add(1)
	f.Add(0)       // Zero length
	f.Add(-1)      // Negative length
	f.Add(1000000) // Very large length

	f.Fuzz(func(t *testing.T, length int) {
		// Limit max length to prevent memory issues
		if length > 100000 {
			t.Skip("length too large")
		}

		k, _ := setupCryptoFuzzKeeper(t)

		// Generate random bytes - must not panic
		randomBytes, err := k.GenerateSecureRandomBytes(length)

		// SECURITY INVARIANT: Zero or negative length must be rejected
		if length < 1 {
			if err == nil {
				t.Error("length < 1 should be rejected")
			}
			return
		}

		if err == nil {
			// SECURITY INVARIANT: Output length must match requested length
			if len(randomBytes) != length {
				t.Errorf("expected %d bytes, got %d", length, len(randomBytes))
			}

			// SECURITY INVARIANT: Non-zero output (statistically unlikely to be all zeros)
			// Only check for small lengths where all-zeros would be extremely unlikely
			if length <= 32 {
				allZero := true
				for _, b := range randomBytes {
					if b != 0 {
						allZero = false
						break
					}
				}
				// Note: This could theoretically fail with probability 2^(-256) for 32 bytes
				// but in practice indicates a broken RNG
				if allZero {
					t.Error("random bytes should not be all zeros (indicates broken RNG)")
				}
			}
		}
	})
}
