package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// TestKeeperBasics tests basic keeper functionality
func TestKeeperBasics(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("Logger", func(t *testing.T) {
		logger := k.Logger(ctx)
		require.NotNil(t, logger)
	})

	t.Run("GetParams and SetParams", func(t *testing.T) {
		params, err := k.GetParams(ctx)
		require.NoError(t, err)
		require.NotNil(t, params)

		// Update params
		newParams := types.DefaultParams()
		newParams.DefaultRotationIntervalDays = 120
		err = k.SetParams(ctx, &newParams)
		require.NoError(t, err)

		// Verify update
		retrievedParams, err := k.GetParams(ctx)
		require.NoError(t, err)
		require.Equal(t, int32(120), retrievedParams.DefaultRotationIntervalDays)
	})

	t.Run("SetParams with invalid params", func(t *testing.T) {
		invalidParams := types.DefaultParams()
		invalidParams.MinThresholdParticipants = 1 // Invalid, should be >= 2
		err := k.SetParams(ctx, &invalidParams)
		require.Error(t, err)
	})

	t.Run("UpdateParams with authority", func(t *testing.T) {
		newParams := types.DefaultParams()
		newParams.DefaultRotationIntervalDays = 180

		// Test with correct authority
		err := k.UpdateParams(ctx, "authority", &newParams)
		require.NoError(t, err)

		// Test with wrong authority
		err = k.UpdateParams(ctx, "wrong-authority", &newParams)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrUnauthorized)
	})

	t.Run("ValidateAuthority", func(t *testing.T) {
		err := k.ValidateAuthority("authority")
		require.NoError(t, err)

		err = k.ValidateAuthority("wrong-authority")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrUnauthorized)
	})
}

// TestRandomGeneration tests random number generation
func TestRandomGenerationComprehensive(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("GenerateSecureRandomBytes - valid", func(t *testing.T) {
		for length := 1; length <= 128; length *= 2 {
			randomBytes, err := k.GenerateSecureRandomBytes(length)
			require.NoError(t, err)
			require.Len(t, randomBytes, length)
		}
	})

	t.Run("GenerateSecureRandomBytes - invalid length", func(t *testing.T) {
		_, err := k.GenerateSecureRandomBytes(0)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInsufficientEntropy)

		_, err = k.GenerateSecureRandomBytes(-1)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInsufficientEntropy)
	})

	t.Run("GenerateSecureRandomInt - valid", func(t *testing.T) {
		for max := int64(1); max <= 1000; max *= 10 {
			randomInt, err := k.GenerateSecureRandomInt(max)
			require.NoError(t, err)
			require.GreaterOrEqual(t, randomInt, int64(0))
			require.Less(t, randomInt, max)
		}
	})

	t.Run("GenerateSecureRandomInt - invalid max", func(t *testing.T) {
		_, err := k.GenerateSecureRandomInt(0)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInsufficientEntropy)

		_, err = k.GenerateSecureRandomInt(-1)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInsufficientEntropy)
	})

	t.Run("GenerateRandomUint64", func(t *testing.T) {
		randomUint, err := k.GenerateRandomUint64(ctx)
		require.NoError(t, err)
		require.NotEqual(t, uint64(0), randomUint) // Very unlikely to be zero
	})

	t.Run("GenerateRandomInRange - invalid range", func(t *testing.T) {
		// min >= max
		_, err := k.GenerateRandomInRange(ctx, 100, 100)
		require.Error(t, err)

		_, err = k.GenerateRandomInRange(ctx, 100, 50)
		require.Error(t, err)
	})

	t.Run("InitializeRandomSource - insufficient entropy", func(t *testing.T) {
		// Too little entropy (less than 256 bits)
		entropy := make([]byte, 16) // 128 bits
		_, err := k.InitializeRandomSource(ctx, cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_SYSTEM, entropy)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInsufficientEntropy)
	})

	t.Run("GenerateRandomBytesFromSource", func(t *testing.T) {
		entropy := make([]byte, 64)
		sourceID, err := k.InitializeRandomSource(ctx, cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_SYSTEM, entropy)
		require.NoError(t, err)

		randomBytes, err := k.GenerateRandomBytesFromSource(ctx, sourceID, 32)
		require.NoError(t, err)
		require.Len(t, randomBytes, 32)
	})

	t.Run("GenerateRandomBytesFromSource - non-existent source", func(t *testing.T) {
		_, err := k.GenerateRandomBytesFromSource(ctx, "non-existent", 32)
		require.Error(t, err)
	})

	t.Run("ReseedRandomSource", func(t *testing.T) {
		entropy := make([]byte, 64)
		sourceID, err := k.InitializeRandomSource(ctx, cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_SYSTEM, entropy)
		require.NoError(t, err)

		// Reseed with valid entropy
		newEntropy := make([]byte, 64)
		err = k.ReseedRandomSource(ctx, sourceID, newEntropy)
		require.NoError(t, err)

		// Reseed with insufficient entropy
		badEntropy := make([]byte, 16)
		err = k.ReseedRandomSource(ctx, sourceID, badEntropy)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInsufficientEntropy)
	})

	t.Run("CheckEntropyHealth", func(t *testing.T) {
		err := k.CheckEntropyHealth(ctx)
		require.NoError(t, err)
	})

	t.Run("GetEntropyStatistics", func(t *testing.T) {
		stats := k.GetEntropyStatistics(ctx)
		require.NotNil(t, stats)
		require.Contains(t, stats, "total_sources")
		require.Contains(t, stats, "healthy_sources")
	})

	t.Run("GetRandomSourceStatus", func(t *testing.T) {
		sources := k.GetRandomSourceStatus(ctx)
		require.NotNil(t, sources)
	})
}

// TestHashingComprehensive tests all hashing functions
func TestHashingComprehensive(t *testing.T) {
	k, ctx := setupKeeper(t)
	data := []byte("test data for comprehensive hashing")

	allAlgorithms := []cryptoproto.HashAlgorithm{
		cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256,
		cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA512,
		cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA3_256,
		cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA3_512,
		cryptoproto.HashAlgorithm_HASH_ALGORITHM_BLAKE2B,
		cryptoproto.HashAlgorithm_HASH_ALGORITHM_BLAKE3,
	}

	for _, algo := range allAlgorithms {
		t.Run("CreateSaltedHash with "+algo.String(), func(t *testing.T) {
			hashID, salt, hash, err := k.CreateSaltedHash(ctx, data, algo, 1000)
			require.NoError(t, err)
			require.NotEmpty(t, hashID)
			require.NotEmpty(t, salt)
			require.NotEmpty(t, hash)

			// Verify hash
			valid, err := k.VerifySaltedHash(ctx, hashID, data)
			require.NoError(t, err)
			require.True(t, valid)

			// Verify with wrong data
			valid, err = k.VerifySaltedHash(ctx, hashID, []byte("wrong data"))
			require.NoError(t, err)
			require.False(t, valid)
		})
	}

	t.Run("VerifySaltedHash - non-existent hash", func(t *testing.T) {
		_, err := k.VerifySaltedHash(ctx, "non-existent", data)
		require.Error(t, err)
	})

	t.Run("HashWithCustomSalt", func(t *testing.T) {
		params, _ := k.GetParams(ctx)
		customSalt := make([]byte, params.MinSaltLengthBytes)

		for _, algo := range allAlgorithms {
			hash, err := k.HashWithCustomSalt(ctx, data, customSalt, algo, 100)
			require.NoError(t, err)
			require.NotEmpty(t, hash)
		}
	})

	t.Run("HashWithCustomSalt - invalid salt length", func(t *testing.T) {
		shortSalt := make([]byte, 4)
		_, err := k.HashWithCustomSalt(ctx, data, shortSalt, cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256, 100)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidSaltLength)
	})

	t.Run("GenerateSalt", func(t *testing.T) {
		salt, err := k.GenerateSalt(ctx, 32)
		require.NoError(t, err)
		require.Len(t, salt, 32)
	})

	t.Run("GenerateSalt - invalid length", func(t *testing.T) {
		params, _ := k.GetParams(ctx)
		_, err := k.GenerateSalt(ctx, params.MinSaltLengthBytes-1)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidSaltLength)
	})

	t.Run("BatchHashWithSalt", func(t *testing.T) {
		dataItems := [][]byte{
			[]byte("data1"),
			[]byte("data2"),
			[]byte("data3"),
		}

		results, err := k.BatchHashWithSalt(ctx, dataItems, cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256, 100)
		require.NoError(t, err)
		require.Len(t, results, 3)
		for _, result := range results {
			require.NotEmpty(t, result.Salt)
			require.NotEmpty(t, result.Hash)
		}
	})

	t.Run("CompareHashes", func(t *testing.T) {
		hash1 := []byte{1, 2, 3, 4}
		hash2 := []byte{1, 2, 3, 4}
		hash3 := []byte{1, 2, 3, 5}
		hash4 := []byte{1, 2, 3}

		require.True(t, k.CompareHashes(hash1, hash2))
		require.False(t, k.CompareHashes(hash1, hash3))
		require.False(t, k.CompareHashes(hash1, hash4))
	})

	t.Run("CreateSaltedHash - zero iterations", func(t *testing.T) {
		// Should default to 1
		_, _, hash, err := k.CreateSaltedHash(ctx, data, cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256, 0)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
	})
}

// TestKeyStretchingComprehensive tests all key stretching algorithms
func TestKeyStretchingComprehensive(t *testing.T) {
	k, ctx := setupKeeper(t)
	password := []byte("secure-password-for-testing")

	t.Run("PBKDF2 SHA256", func(t *testing.T) {
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

		key, err := k.StretchKey(ctx, configID, password)
		require.NoError(t, err)
		require.Len(t, key, 32)

		// Verify key
		valid, err := k.VerifyStretchedKey(ctx, configID, password, key)
		require.NoError(t, err)
		require.True(t, valid)

		// Wrong password
		valid, err = k.VerifyStretchedKey(ctx, configID, []byte("wrong"), key)
		require.NoError(t, err)
		require.False(t, valid)
	})

	t.Run("PBKDF2 SHA512", func(t *testing.T) {
		configID, _, err := k.CreateKeyStretchingConfig(
			ctx,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA512,
			100000,
			0,
			0,
			64,
		)
		require.NoError(t, err)

		key, err := k.StretchKey(ctx, configID, password)
		require.NoError(t, err)
		require.Len(t, key, 64)
	})

	t.Run("Argon2i", func(t *testing.T) {
		configID, _, err := k.CreateKeyStretchingConfig(
			ctx,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2I,
			3,
			65536,
			4,
			32,
		)
		require.NoError(t, err)

		key, err := k.StretchKey(ctx, configID, password)
		require.NoError(t, err)
		require.Len(t, key, 32)
	})

	t.Run("Argon2d", func(t *testing.T) {
		configID, _, err := k.CreateKeyStretchingConfig(
			ctx,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2D,
			3,
			65536,
			4,
			32,
		)
		require.NoError(t, err)

		key, err := k.StretchKey(ctx, configID, password)
		require.NoError(t, err)
		require.Len(t, key, 32)
	})

	t.Run("Scrypt", func(t *testing.T) {
		configID, _, err := k.CreateKeyStretchingConfig(
			ctx,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_SCRYPT,
			14, // N=2^14
			0,
			8,
			32,
		)
		require.NoError(t, err)

		key, err := k.StretchKey(ctx, configID, password)
		require.NoError(t, err)
		require.Len(t, key, 32)
	})

	t.Run("Invalid iteration counts", func(t *testing.T) {
		// PBKDF2 - too few iterations
		_, _, err := k.CreateKeyStretchingConfig(
			ctx,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA256,
			1000, // Less than MinPbkdf2Iterations
			0,
			0,
			32,
		)
		require.Error(t, err)

		// Argon2 - too few iterations
		_, _, err = k.CreateKeyStretchingConfig(
			ctx,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2ID,
			1, // Less than MinArgon2Iterations
			65536,
			4,
			32,
		)
		require.Error(t, err)

		// Scrypt - too few iterations
		_, _, err = k.CreateKeyStretchingConfig(
			ctx,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_SCRYPT,
			10, // Less than 14
			0,
			8,
			32,
		)
		require.Error(t, err)
	})

	t.Run("Invalid memory cost for Argon2", func(t *testing.T) {
		_, _, err := k.CreateKeyStretchingConfig(
			ctx,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2ID,
			3,
			1024, // Less than MinArgon2MemoryKb
			4,
			32,
		)
		require.Error(t, err)
	})

	t.Run("Invalid parallelism", func(t *testing.T) {
		_, _, err := k.CreateKeyStretchingConfig(
			ctx,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2ID,
			3,
			65536,
			0, // Invalid parallelism
			32,
		)
		require.Error(t, err)
	})

	t.Run("Invalid key length", func(t *testing.T) {
		_, _, err := k.CreateKeyStretchingConfig(
			ctx,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA256,
			100000,
			0,
			0,
			8, // Less than 16
		)
		require.Error(t, err)
	})

	t.Run("StretchKey - non-existent config", func(t *testing.T) {
		_, err := k.StretchKey(ctx, "non-existent", password)
		require.Error(t, err)
	})

	t.Run("StretchKeyWithParams", func(t *testing.T) {
		salt := make([]byte, 16)
		key, err := k.StretchKeyWithParams(
			ctx,
			password,
			salt,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA256,
			100000,
			0,
			0,
			32,
		)
		require.NoError(t, err)
		require.Len(t, key, 32)
	})

	t.Run("GetRecommendedStretchingConfig", func(t *testing.T) {
		algorithms := []cryptoproto.KeyStretchingAlgorithm{
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA256,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA512,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2ID,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_SCRYPT,
		}

		for _, algo := range algorithms {
			config, err := k.GetRecommendedStretchingConfig(ctx, algo)
			require.NoError(t, err)
			require.NotNil(t, config)
			require.NotEmpty(t, config.Salt)
		}

		// Unsupported algorithm
		_, err := k.GetRecommendedStretchingConfig(ctx, cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_UNSPECIFIED)
		require.Error(t, err)
	})
}

// TestKeyRotationComprehensive tests all key rotation functionality
func TestKeyRotationComprehensive(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("DisableKeyRotationSchedule", func(t *testing.T) {
		scheduleID, err := k.CreateKeyRotationSchedule(ctx, "creator", "test-key", 86400, nil)
		require.NoError(t, err)

		err = k.DisableKeyRotationSchedule(ctx, scheduleID)
		require.NoError(t, err)

		schedule, err := k.GetKeyRotationSchedule(ctx, scheduleID)
		require.NoError(t, err)
		require.False(t, schedule.Enabled)
	})

	t.Run("EnableKeyRotationSchedule", func(t *testing.T) {
		scheduleID, err := k.CreateKeyRotationSchedule(ctx, "creator", "test-key", 86400, nil)
		require.NoError(t, err)

		err = k.DisableKeyRotationSchedule(ctx, scheduleID)
		require.NoError(t, err)

		err = k.EnableKeyRotationSchedule(ctx, scheduleID)
		require.NoError(t, err)

		schedule, err := k.GetKeyRotationSchedule(ctx, scheduleID)
		require.NoError(t, err)
		require.True(t, schedule.Enabled)
	})

	t.Run("GetSchedulesForKey", func(t *testing.T) {
		keyID1 := "multi-schedule-key-1"
		keyIDShared := "shared-schedule-key"

		_, err := k.CreateKeyRotationSchedule(ctx, "creator", keyID1, 86400, nil)
		require.NoError(t, err)

		// Sleep to ensure different Unix timestamp
		time.Sleep(time.Second + 10*time.Millisecond)

		_, err = k.CreateKeyRotationSchedule(ctx, "creator", keyIDShared, 172800, nil)
		require.NoError(t, err)

		// Sleep again
		time.Sleep(time.Second + 10*time.Millisecond)

		_, err = k.CreateKeyRotationSchedule(ctx, "creator", keyIDShared, 259200, nil)
		require.NoError(t, err)

		// Get schedules for the shared key - should have 2
		schedules := k.GetSchedulesForKey(ctx, keyIDShared)
		require.Len(t, schedules, 2)

		// Get schedules for key1 - should have 1
		schedules1 := k.GetSchedulesForKey(ctx, keyID1)
		require.Len(t, schedules1, 1)
	})

	t.Run("ProcessScheduledRotations", func(t *testing.T) {
		err := k.ProcessScheduledRotations(ctx)
		require.NoError(t, err)
	})

	t.Run("GetAllKeyRotationSchedules", func(t *testing.T) {
		schedules := k.GetAllKeyRotationSchedules(ctx)
		require.NotNil(t, schedules)
	})

	t.Run("GetKeyRotationSchedule - non-existent", func(t *testing.T) {
		_, err := k.GetKeyRotationSchedule(ctx, "non-existent")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrKeyRotationScheduleNotFound)
	})

	t.Run("DisableKeyRotationSchedule - non-existent", func(t *testing.T) {
		err := k.DisableKeyRotationSchedule(ctx, "non-existent")
		require.Error(t, err)
	})

	t.Run("EnableKeyRotationSchedule - non-existent", func(t *testing.T) {
		err := k.EnableKeyRotationSchedule(ctx, "non-existent")
		require.Error(t, err)
	})

	t.Run("RotateKey - invalid key ID", func(t *testing.T) {
		_, _, err := k.RotateKey(ctx, "creator", "", []byte("public-key"))
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidKeyID)
	})

	t.Run("RotateKey - short public key", func(t *testing.T) {
		_, _, err := k.RotateKey(ctx, "creator", "test-key", []byte("short"))
		require.Error(t, err)
	})

	t.Run("CreateKeyRotationSchedule - empty key ID", func(t *testing.T) {
		_, err := k.CreateKeyRotationSchedule(ctx, "creator", "", 86400, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidKeyID)
	})
}

// TestThresholdSignaturesComprehensive tests threshold signatures
func TestThresholdSignaturesComprehensive(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("CreateThresholdScheme - invalid threshold", func(t *testing.T) {
		_, _, err := k.CreateThresholdScheme(ctx, "creator", 5, 3, []string{"p1", "p2", "p3"}, cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA)
		require.Error(t, err)
	})

	t.Run("CreateThresholdScheme - zero threshold", func(t *testing.T) {
		_, _, err := k.CreateThresholdScheme(ctx, "creator", 0, 3, []string{"p1", "p2", "p3"}, cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA)
		require.Error(t, err)
	})

	t.Run("CreateThresholdScheme - mismatched participant count", func(t *testing.T) {
		_, _, err := k.CreateThresholdScheme(ctx, "creator", 2, 5, []string{"p1", "p2", "p3"}, cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA)
		require.Error(t, err)
	})

	t.Run("SubmitThresholdSignatureShare - non-existent scheme", func(t *testing.T) {
		_, _, _, err := k.SubmitThresholdSignatureShare(ctx, "p1", "non-existent", []byte("share"), []byte("hash"))
		require.Error(t, err)
	})

	t.Run("SubmitThresholdSignatureShare - non-participant", func(t *testing.T) {
		participants := []string{"p1", "p2", "p3"}
		schemeID, _, err := k.CreateThresholdScheme(ctx, "creator", 2, 3, participants, cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA)
		require.NoError(t, err)

		_, _, _, err = k.SubmitThresholdSignatureShare(ctx, "p4", schemeID, []byte("share"), []byte("hash"))
		require.Error(t, err)
	})

	t.Run("GetThresholdScheme - non-existent", func(t *testing.T) {
		_, err := k.GetThresholdScheme(ctx, "non-existent")
		require.Error(t, err)
	})

	t.Run("GetAllThresholdSchemes", func(t *testing.T) {
		schemes := k.GetAllThresholdSchemes(ctx)
		require.NotNil(t, schemes)
	})

	t.Run("Multiple threshold scheme types", func(t *testing.T) {
		types := []cryptoproto.ThresholdSchemeType{
			cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA,
			cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_EDDSA,
			cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_BLS,
		}

		for _, schemeType := range types {
			_, _, err := k.CreateThresholdScheme(ctx, "creator", 2, 3, []string{"p1", "p2", "p3"}, schemeType)
			require.NoError(t, err)
		}
	})
}

// TestQuantumResistantKeysComprehensive tests quantum-resistant key operations
func TestQuantumResistantKeysComprehensive(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("GenerateQuantumResistantKey - unspecified algorithm", func(t *testing.T) {
		_, _, err := k.GenerateQuantumResistantKey(ctx, "creator", cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_UNSPECIFIED, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidQuantumAlgorithm)
	})

	t.Run("GenerateQuantumResistantKey - all algorithms with expiration", func(t *testing.T) {
		algorithms := []cryptoproto.QuantumResistantAlgorithm{
			cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM,
			cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER,
			cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_FALCON,
			cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_SPHINCS_PLUS,
			cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_NTRU,
		}

		expiresAt := time.Now().Add(365 * 24 * time.Hour)
		for _, algo := range algorithms {
			keyID, publicKey, err := k.GenerateQuantumResistantKey(ctx, "creator", algo, &expiresAt)
			require.NoError(t, err)
			require.NotEmpty(t, keyID)
			require.NotEmpty(t, publicKey)

			// Validate the key
			err = k.ValidateQuantumResistantKey(ctx, keyID)
			require.NoError(t, err)
		}
	})

	t.Run("GetQuantumResistantKey - non-existent", func(t *testing.T) {
		_, err := k.GetQuantumResistantKey(ctx, "non-existent")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrQuantumKeyNotFound)
	})

	// Note: Testing expired key retrieval is complex due to caching
	// The ValidateQuantumResistantKey test below covers the expiry logic

	t.Run("ValidateQuantumResistantKey - expired", func(t *testing.T) {
		pastTime := time.Now().Add(-24 * time.Hour)
		keyID, _, err := k.GenerateQuantumResistantKey(ctx, "creator", cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_FALCON, &pastTime)
		require.NoError(t, err)

		err = k.ValidateQuantumResistantKey(ctx, keyID)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrKeyExpired)
	})

	t.Run("RotateQuantumResistantKey", func(t *testing.T) {
		expiresAt := time.Now().Add(365 * 24 * time.Hour)
		keyID, publicKey1, err := k.GenerateQuantumResistantKey(ctx, "creator", cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER, &expiresAt)
		require.NoError(t, err)

		newExpiresAt := time.Now().Add(730 * 24 * time.Hour)
		newKeyID, newPublicKey, err := k.RotateQuantumResistantKey(ctx, keyID, &newExpiresAt)
		require.NoError(t, err)
		require.NotEmpty(t, newKeyID)
		require.NotEmpty(t, newPublicKey)
		// Key IDs will be different due to different timestamps
		// Public keys will be different as they're randomly generated
		require.NotEqual(t, publicKey1, newPublicKey)
	})

	t.Run("DeleteQuantumResistantKey", func(t *testing.T) {
		keyID, _, err := k.GenerateQuantumResistantKey(ctx, "creator", cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_SPHINCS_PLUS, nil)
		require.NoError(t, err)

		err = k.DeleteQuantumResistantKey(ctx, keyID)
		require.NoError(t, err)
	})

	t.Run("DeleteZKProofConfig", func(t *testing.T) {
		// This tests the DeleteZKProofConfig method
		err := k.DeleteZKProofConfig(ctx, "some-proof-id")
		require.NoError(t, err)
	})

	t.Run("GetAllZKProofConfigs", func(t *testing.T) {
		configs := k.GetAllZKProofConfigs(ctx)
		require.NotNil(t, configs)
	})
}

// TestCertificatePinningComprehensive tests certificate pinning
func TestCertificatePinningComprehensive(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("AddCertificatePin - empty hostname", func(t *testing.T) {
		hash := make([]byte, 32)
		_, err := k.AddCertificatePin(ctx, "creator", "", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.Error(t, err)
	})

	t.Run("AddCertificatePin - empty hashes", func(t *testing.T) {
		_, err := k.AddCertificatePin(ctx, "creator", "example.com", [][]byte{}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.Error(t, err)
	})

	t.Run("AddCertificatePin - invalid hash length", func(t *testing.T) {
		hash := make([]byte, 16) // Wrong length, should be 32
		_, err := k.AddCertificatePin(ctx, "creator", "example.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidCertificateHash)
	})

	t.Run("GetCertificatePin - non-existent", func(t *testing.T) {
		_, err := k.GetCertificatePin(ctx, "non-existent.com")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrCertificatePinNotFound)
	})

	t.Run("DisableCertificatePin", func(t *testing.T) {
		hash := make([]byte, 32)
		_, err := k.AddCertificatePin(ctx, "creator", "disable-test.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.NoError(t, err)

		err = k.DisableCertificatePin(ctx, "disable-test.com")
		require.NoError(t, err)

		pin, err := k.GetCertificatePin(ctx, "disable-test.com")
		require.NoError(t, err)
		require.False(t, pin.Enabled)
	})

	t.Run("EnableCertificatePin", func(t *testing.T) {
		hash := make([]byte, 32)
		_, err := k.AddCertificatePin(ctx, "creator", "enable-test.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.NoError(t, err)

		err = k.DisableCertificatePin(ctx, "enable-test.com")
		require.NoError(t, err)

		err = k.EnableCertificatePin(ctx, "enable-test.com")
		require.NoError(t, err)

		pin, err := k.GetCertificatePin(ctx, "enable-test.com")
		require.NoError(t, err)
		require.True(t, pin.Enabled)
	})

	t.Run("RemoveCertificatePin", func(t *testing.T) {
		hash := make([]byte, 32)
		_, err := k.AddCertificatePin(ctx, "creator", "remove-test.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.NoError(t, err)

		err = k.RemoveCertificatePin(ctx, "remove-test.com")
		require.NoError(t, err)

		_, err = k.GetCertificatePin(ctx, "remove-test.com")
		require.Error(t, err)
	})

	t.Run("RotateCertificatePin", func(t *testing.T) {
		hash1 := make([]byte, 32)
		_, err := k.AddCertificatePin(ctx, "creator", "rotate-test.com", [][]byte{hash1}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.NoError(t, err)

		hash2 := make([]byte, 32)
		for i := range hash2 {
			hash2[i] = byte(i + 10)
		}
		err = k.RotateCertificatePin(ctx, "rotate-test.com", [][]byte{hash2})
		require.NoError(t, err)

		pin, err := k.GetCertificatePin(ctx, "rotate-test.com")
		require.NoError(t, err)
		require.Len(t, pin.CertificateHashes, 2)
	})

	t.Run("RotateCertificatePin - invalid hash", func(t *testing.T) {
		hash := make([]byte, 32)
		_, err := k.AddCertificatePin(ctx, "creator", "rotate-invalid.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.NoError(t, err)

		badHash := make([]byte, 16)
		err = k.RotateCertificatePin(ctx, "rotate-invalid.com", [][]byte{badHash})
		require.Error(t, err)
	})

	t.Run("UpdateCertificatePin - invalid hash", func(t *testing.T) {
		hash := make([]byte, 32)
		_, err := k.AddCertificatePin(ctx, "creator", "update-test.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.NoError(t, err)

		badHash := make([]byte, 16)
		err = k.UpdateCertificatePin(ctx, "update-test.com", [][]byte{badHash}, nil)
		require.Error(t, err)
	})

	t.Run("ListCertificatePins", func(t *testing.T) {
		pins := k.ListCertificatePins(ctx)
		require.NotNil(t, pins)
	})

	t.Run("CleanupExpiredPins", func(t *testing.T) {
		hash := make([]byte, 32)
		pastTime := time.Now().Add(-24 * time.Hour)
		_, err := k.AddCertificatePin(ctx, "creator", "expired-test.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, &pastTime)
		require.NoError(t, err)

		err = k.CleanupExpiredPins(ctx)
		require.NoError(t, err)
	})

	t.Run("VerifyCertificatePin - disabled", func(t *testing.T) {
		hash := make([]byte, 32)
		_, err := k.AddCertificatePin(ctx, "creator", "verify-disabled.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.NoError(t, err)

		err = k.DisableCertificatePin(ctx, "verify-disabled.com")
		require.NoError(t, err)

		_, err = k.VerifyCertificatePin(ctx, "verify-disabled.com", []byte("cert"))
		require.Error(t, err)
	})

	t.Run("VerifyCertificatePin - expired", func(t *testing.T) {
		hash := make([]byte, 32)
		pastTime := time.Now().Add(-24 * time.Hour)
		_, err := k.AddCertificatePin(ctx, "creator", "verify-expired.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, &pastTime)
		require.NoError(t, err)

		_, err = k.VerifyCertificatePin(ctx, "verify-expired.com", []byte("cert"))
		require.Error(t, err)
	})

	t.Run("VerifyCertificatePin - full cert type", func(t *testing.T) {
		cert := []byte("test certificate data")
		hash := make([]byte, 32)
		_, err := k.AddCertificatePin(ctx, "creator", "full-cert-test.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_FULL_CERT, nil)
		require.NoError(t, err)

		// Should fail because hash won't match
		_, err = k.VerifyCertificatePin(ctx, "full-cert-test.com", cert)
		require.Error(t, err)
	})

	t.Run("VerifyCertificatePin - intermediate type", func(t *testing.T) {
		cert := []byte("test intermediate certificate")
		hash := make([]byte, 32)
		_, err := k.AddCertificatePin(ctx, "creator", "intermediate-test.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_INTERMEDIATE, nil)
		require.NoError(t, err)

		// Should fail because hash won't match
		_, err = k.VerifyCertificatePin(ctx, "intermediate-test.com", cert)
		require.Error(t, err)
	})
}

// TestSecureEnclaveComprehensive tests secure enclave functionality
func TestSecureEnclaveComprehensive(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("RegisterSecureEnclave - unspecified type", func(t *testing.T) {
		_, err := k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_UNSPECIFIED, []byte("attestation"), nil)
		require.Error(t, err)
	})

	t.Run("RegisterSecureEnclave - empty attestation", func(t *testing.T) {
		_, err := k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, []byte{}, nil)
		require.Error(t, err)
	})

	t.Run("RegisterSecureEnclave - SGX attestation too short", func(t *testing.T) {
		attestation := make([]byte, 100) // Less than 432
		_, err := k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, attestation, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrEnclaveAttestationFailed)
	})

	t.Run("RegisterSecureEnclave - all types with valid attestation", func(t *testing.T) {
		types := []cryptoproto.SecureEnclaveType{
			cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX,
			cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SEV,
			cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_TPM,
			cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_HSM,
			cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_KEYCHAIN,
		}

		for _, enclaveType := range types {
			var attestation []byte
			switch enclaveType {
			case cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX:
				attestation = make([]byte, 432)
			case cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SEV:
				attestation = make([]byte, 144)
			case cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_TPM:
				attestation = make([]byte, 48)
			case cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_HSM:
				attestation = make([]byte, 32)
			case cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_KEYCHAIN:
				attestation = make([]byte, 16)
			}

			metadata := map[string]string{"version": "1.0"}
			enclaveID, err := k.RegisterSecureEnclave(ctx, "creator", enclaveType, attestation, metadata)
			require.NoError(t, err)
			require.NotEmpty(t, enclaveID)

			// Verify enclave
			enclave, err := k.GetSecureEnclave(ctx, enclaveID)
			require.NoError(t, err)
			require.Equal(t, enclaveType, enclave.EnclaveType)
		}
	})

	t.Run("GetSecureEnclave - non-existent", func(t *testing.T) {
		_, err := k.GetSecureEnclave(ctx, "non-existent")
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrSecureEnclaveNotFound)
	})

	t.Run("SealDataToEnclave - enclave not ready", func(t *testing.T) {
		attestation := make([]byte, 432)
		enclaveID, err := k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, attestation, nil)
		require.NoError(t, err)

		err = k.UpdateEnclaveStatus(ctx, enclaveID, cryptoproto.SecureEnclaveStatus_SECURE_ENCLAVE_STATUS_ERROR)
		require.NoError(t, err)

		_, err = k.SealDataToEnclave(ctx, enclaveID, []byte("data"))
		require.Error(t, err)
	})

	t.Run("UnsealDataFromEnclave - enclave not ready", func(t *testing.T) {
		attestation := make([]byte, 432)
		enclaveID, err := k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, attestation, nil)
		require.NoError(t, err)

		data := []byte("secret data")
		sealed, err := k.SealDataToEnclave(ctx, enclaveID, data)
		require.NoError(t, err)

		err = k.UpdateEnclaveStatus(ctx, enclaveID, cryptoproto.SecureEnclaveStatus_SECURE_ENCLAVE_STATUS_ERROR)
		require.NoError(t, err)

		_, err = k.UnsealDataFromEnclave(ctx, enclaveID, sealed)
		require.Error(t, err)
	})

	t.Run("UnsealDataFromEnclave - invalid sealed data", func(t *testing.T) {
		attestation := make([]byte, 432)
		enclaveID, err := k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, attestation, nil)
		require.NoError(t, err)

		// Too short sealed data
		_, err = k.UnsealDataFromEnclave(ctx, enclaveID, []byte("short"))
		require.Error(t, err)
	})

	t.Run("ListSecureEnclaves", func(t *testing.T) {
		enclaves := k.ListSecureEnclaves(ctx)
		require.NotNil(t, enclaves)
	})

	t.Run("RemoteAttestEnclave", func(t *testing.T) {
		attestation := make([]byte, 432)
		enclaveID, err := k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, attestation, nil)
		require.NoError(t, err)

		report, err := k.RemoteAttestEnclave(ctx, enclaveID)
		require.NoError(t, err)
		require.NotEmpty(t, report)
	})

	t.Run("UpdateEnclaveStatus", func(t *testing.T) {
		attestation := make([]byte, 432)
		enclaveID, err := k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, attestation, nil)
		require.NoError(t, err)

		statuses := []cryptoproto.SecureEnclaveStatus{
			cryptoproto.SecureEnclaveStatus_SECURE_ENCLAVE_STATUS_UNINITIALIZED,
			cryptoproto.SecureEnclaveStatus_SECURE_ENCLAVE_STATUS_SEALED,
			cryptoproto.SecureEnclaveStatus_SECURE_ENCLAVE_STATUS_ERROR,
			cryptoproto.SecureEnclaveStatus_SECURE_ENCLAVE_STATUS_READY,
		}

		for _, status := range statuses {
			err = k.UpdateEnclaveStatus(ctx, enclaveID, status)
			require.NoError(t, err)

			enclave, err := k.GetSecureEnclave(ctx, enclaveID)
			require.NoError(t, err)
			require.Equal(t, status, enclave.Status)
		}
	})
}

// TestZKProofOperations tests ZK proof operations
func TestZKProofOperations(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("RegisterZKProofCircuit - not implemented", func(t *testing.T) {
		_, err := k.RegisterZKProofCircuit(ctx, "creator", cryptoproto.ZKProofType_ZK_PROOF_TYPE_GROTH16, []byte("params"), []byte("key"), "circuit-1")
		require.Error(t, err)
	})

	t.Run("SubmitZKProof - not implemented", func(t *testing.T) {
		_, _, err := k.SubmitZKProof(ctx, "submitter", "proof-1", []byte("data"), []byte("inputs"))
		require.Error(t, err)
	})

	t.Run("VerifyZKProof - not implemented", func(t *testing.T) {
		_, err := k.VerifyZKProof(ctx, "proof-1", []byte("data"), []byte("inputs"))
		require.Error(t, err)
	})
}
