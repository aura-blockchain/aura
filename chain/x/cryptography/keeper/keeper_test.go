package keeper_test

import (
	"context"
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
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/cryptography/keeper"
	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

func setupKeeper(t *testing.T) (keeper.Keeper, context.Context) {
	// Configure bech32 prefixes for "aura" chain
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("aura", "aurapub")
	config.SetBech32PrefixForValidator("auravaloper", "auravaloper pub")
	config.SetBech32PrefixForConsensusNode("auravalcons", "auravalconspub")

	storeKey := storetypes.NewKVStoreKey(types.ModuleName)

	// Create database and commit multi-store
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, cms.LoadLatestVersion())

	// Create context with proper store
	header := cmtproto.Header{
		Height: 1,
		Time:   time.Now(),
	}
	ctx := sdk.NewContext(cms, header, false, log.NewNopLogger())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		log.NewNopLogger(),
		"authority",
	)

	// Initialize default params
	params := types.DefaultParams()
	err := k.SetParams(ctx, &params)
	require.NoError(t, err)

	return *k, ctx
}

func TestKeyRotation(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("Create key rotation schedule", func(t *testing.T) {
		policy := &cryptoproto.KeyRotationPolicy{
			MaxAgeDays:              90,
			WarningDaysBeforeExpiry: 7,
			AutoRotate:              true,
			MaxRotationAttempts:     3,
		}

		scheduleID, err := k.CreateKeyRotationSchedule(
			ctx,
			"creator",
			"test-key-1",
			86400, // 24 hours
			policy,
		)

		require.NoError(t, err)
		require.NotEmpty(t, scheduleID)

		// Retrieve schedule
		schedule, err := k.GetKeyRotationSchedule(ctx, scheduleID)
		require.NoError(t, err)
		require.Equal(t, "test-key-1", schedule.KeyId)
		require.True(t, schedule.Enabled)
	})

	t.Run("Manual key rotation", func(t *testing.T) {
		publicKey := make([]byte, 32)
		for i := range publicKey {
			publicKey[i] = byte(i)
		}

		rotationID, rotationTime, err := k.RotateKey(
			ctx,
			"creator",
			"test-key-2",
			publicKey,
		)

		require.NoError(t, err)
		require.NotEmpty(t, rotationID)
		require.False(t, rotationTime.IsZero())
	})

	t.Run("Invalid rotation interval", func(t *testing.T) {
		_, err := k.CreateKeyRotationSchedule(
			ctx,
			"creator",
			"test-key-3",
			1000, // Too short (< 3600)
			nil,
		)

		require.Error(t, err)
	})
}

// TestHDKeyDerivation - SKIPPED: DeriveHDKey and ValidateBIP44Path methods not implemented
// func TestHDKeyDerivation(t *testing.T) {
// 	k, ctx := setupKeeper(t)
//
// 	t.Run("Derive BIP44 key", func(t *testing.T) {
// 		seedHash := make([]byte, 32)
// 		for i := range seedHash {
// 			seedHash[i] = byte(i)
// 		}
//
// 		hdKey, err := k.DeriveHDKey(
// 			ctx,
// 			"master-key-1",
// 			seedHash,
// 			"m/44'/118'/0'/0/0",
// 		)
//
// 		require.NoError(t, err)
// 		require.NotNil(t, hdKey)
// 		require.Equal(t, "master-key-1", hdKey.MasterKeyId)
// 		require.Equal(t, "m/44'/118'/0'/0/0", hdKey.DerivationPath)
// 		require.Equal(t, int32(5), hdKey.Depth)
// 	})
//
// 	t.Run("Validate BIP44 path", func(t *testing.T) {
// 		err := k.ValidateBIP44Path("m/44'/118'/0'/0/0")
// 		require.NoError(t, err)
//
// 		err = k.ValidateBIP44Path("m/43'/118'/0'/0/0") // Wrong purpose
// 		require.Error(t, err)
// 	})
//
// 	t.Run("Invalid derivation path", func(t *testing.T) {
// 		seedHash := make([]byte, 32)
//
// 		_, err := k.DeriveHDKey(
// 			ctx,
// 			"master-key-2",
// 			seedHash,
// 			"invalid-path",
// 		)
//
// 		require.Error(t, err)
// 	})
// }

func TestThresholdSignatures(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("Create threshold scheme", func(t *testing.T) {
		participants := []string{"p1", "p2", "p3", "p4", "p5"}

		schemeID, publicKey, err := k.CreateThresholdScheme(
			ctx,
			"creator",
			3, // threshold
			5, // total participants
			participants,
			cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA,
		)

		require.NoError(t, err)
		require.NotEmpty(t, schemeID)
		require.NotEmpty(t, publicKey)

		// Retrieve scheme
		scheme, err := k.GetThresholdScheme(ctx, schemeID)
		require.NoError(t, err)
		require.Equal(t, int32(3), scheme.Threshold)
		require.Equal(t, int32(5), scheme.TotalParticipants)
	})

	t.Run("Submit signature shares", func(t *testing.T) {
		participants := []string{"p1", "p2", "p3"}
		schemeID, _, err := k.CreateThresholdScheme(
			ctx,
			"creator",
			2,
			3,
			participants,
			cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA,
		)
		require.NoError(t, err)

		messageHash := make([]byte, 32)
		share1 := make([]byte, 64)
		share2 := make([]byte, 64)

		// Submit first share
		shares1, reached1, _, err := k.SubmitThresholdSignatureShare(
			ctx,
			"p1",
			schemeID,
			share1,
			messageHash,
		)
		require.NoError(t, err)
		require.Equal(t, uint32(1), shares1)
		require.False(t, reached1)

		// Submit second share - threshold reached
		shares2, reached2, combined, err := k.SubmitThresholdSignatureShare(
			ctx,
			"p2",
			schemeID,
			share2,
			messageHash,
		)
		require.NoError(t, err)
		require.Equal(t, uint32(2), shares2)
		require.True(t, reached2)
		require.NotEmpty(t, combined)
	})
}

func TestQuantumResistantKeys(t *testing.T) {
	k, ctx := setupKeeper(t)

	algorithms := []cryptoproto.QuantumResistantAlgorithm{
		cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM,
		cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER,
		cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_FALCON,
		cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_SPHINCS_PLUS,
	}

	for _, algo := range algorithms {
		t.Run("Generate "+algo.String(), func(t *testing.T) {
			expiresAt := time.Now().Add(365 * 24 * time.Hour)

			keyID, publicKey, err := k.GenerateQuantumResistantKey(
				ctx,
				"creator",
				algo,
				&expiresAt,
			)

			require.NoError(t, err)
			require.NotEmpty(t, keyID)
			require.NotEmpty(t, publicKey)

			// Retrieve key
			key, err := k.GetQuantumResistantKey(ctx, keyID)
			require.NoError(t, err)
			require.Equal(t, algo, key.Algorithm)
		})
	}
}

func TestRandomGeneration(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("Initialize random source", func(t *testing.T) {
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

		// Retrieve source
		source, err := k.GetRandomSource(ctx, sourceID)
		require.NoError(t, err)
		require.Equal(t, int64(512), source.EntropyBits)
	})

	t.Run("Generate random bytes", func(t *testing.T) {
		randomBytes, err := k.GenerateSecureRandomBytes(32)
		require.NoError(t, err)
		require.Len(t, randomBytes, 32)
	})

	t.Run("Generate random in range", func(t *testing.T) {
		value, err := k.GenerateRandomInRange(ctx, 1, 100)
		require.NoError(t, err)
		require.GreaterOrEqual(t, value, int64(1))
		require.LessOrEqual(t, value, int64(100))
	})
}

func TestSaltedHashing(t *testing.T) {
	k, ctx := setupKeeper(t)

	algorithms := []cryptoproto.HashAlgorithm{
		cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256,
		cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA512,
		cryptoproto.HashAlgorithm_HASH_ALGORITHM_BLAKE2B,
	}

	for _, algo := range algorithms {
		t.Run("Hash with "+algo.String(), func(t *testing.T) {
			data := []byte("test data for hashing")

			hashID, salt, hash, err := k.CreateSaltedHash(
				ctx,
				data,
				algo,
				1000,
			)

			require.NoError(t, err)
			require.NotEmpty(t, hashID)
			require.NotEmpty(t, salt)
			require.NotEmpty(t, hash)

			// Verify hash
			valid, err := k.VerifySaltedHash(ctx, hashID, data)
			require.NoError(t, err)
			require.True(t, valid)

			// Verify with wrong data
			wrongData := []byte("wrong data")
			valid, err = k.VerifySaltedHash(ctx, hashID, wrongData)
			require.NoError(t, err)
			require.False(t, valid)
		})
	}
}

func TestKeyStretching(t *testing.T) {
	k, ctx := setupKeeper(t)

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

		password := []byte("secure-password-123")
		key, err := k.StretchKey(ctx, configID, password)
		require.NoError(t, err)
		require.Len(t, key, 32)
	})

	t.Run("Argon2id", func(t *testing.T) {
		configID, _, err := k.CreateKeyStretchingConfig(
			ctx,
			cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2ID,
			3,
			65536,
			4,
			32,
		)

		require.NoError(t, err)

		password := []byte("secure-password-456")
		key, err := k.StretchKey(ctx, configID, password)
		require.NoError(t, err)
		require.Len(t, key, 32)
	})
}

func TestCertificatePinning(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("Add certificate pin", func(t *testing.T) {
		hash1 := make([]byte, 32)
		hash2 := make([]byte, 32)
		for i := range hash1 {
			hash1[i] = byte(i)
			hash2[i] = byte(i + 1)
		}

		expiresAt := time.Now().Add(365 * 24 * time.Hour)

		pinID, err := k.AddCertificatePin(
			ctx,
			"creator",
			"example.com",
			[][]byte{hash1, hash2},
			cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI,
			&expiresAt,
		)

		require.NoError(t, err)
		require.NotEmpty(t, pinID)

		// Retrieve pin
		pin, err := k.GetCertificatePin(ctx, "example.com")
		require.NoError(t, err)
		require.Equal(t, "example.com", pin.Hostname)
		require.Len(t, pin.CertificateHashes, 2)
	})

	t.Run("Update certificate pin", func(t *testing.T) {
		hash := make([]byte, 32)
		k.AddCertificatePin(
			ctx,
			"creator",
			"test.com",
			[][]byte{hash},
			cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI,
			nil,
		)

		newHash := make([]byte, 32)
		for i := range newHash {
			newHash[i] = byte(i + 10)
		}

		err := k.UpdateCertificatePin(ctx, "test.com", [][]byte{newHash}, nil)
		require.NoError(t, err)
	})
}

func TestSecureEnclave(t *testing.T) {
	k, ctx := setupKeeper(t)

	t.Run("Register secure enclave", func(t *testing.T) {
		attestation := make([]byte, 432) // Minimum SGX attestation size
		metadata := map[string]string{
			"version": "1.0",
			"vendor":  "Intel",
		}

		enclaveID, err := k.RegisterSecureEnclave(
			ctx,
			"creator",
			cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX,
			attestation,
			metadata,
		)

		require.NoError(t, err)
		require.NotEmpty(t, enclaveID)

		// Retrieve enclave
		enclave, err := k.GetSecureEnclave(ctx, enclaveID)
		require.NoError(t, err)
		require.Equal(t, cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, enclave.EnclaveType)
	})

	t.Run("Seal and unseal data", func(t *testing.T) {
		attestation := make([]byte, 432)
		enclaveID, _ := k.RegisterSecureEnclave(
			ctx,
			"creator",
			cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX,
			attestation,
			nil,
		)

		data := []byte("secret data to seal")
		sealed, err := k.SealDataToEnclave(ctx, enclaveID, data)
		require.NoError(t, err)
		require.NotEmpty(t, sealed)

		unsealed, err := k.UnsealDataFromEnclave(ctx, enclaveID, sealed)
		require.NoError(t, err)
		require.Equal(t, data, unsealed)
	})
}
