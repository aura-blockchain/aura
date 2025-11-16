package keeper

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// CreateKeyStretchingConfig creates a key stretching configuration
func (k Keeper) CreateKeyStretchingConfig(
	ctx context.Context,
	algorithm cryptoproto.KeyStretchingAlgorithm,
	iterations int32,
	memoryCost int32,
	parallelism int32,
	keyLength int32,
) (string, []byte, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return "", nil, err
	}

	// Validate parameters based on algorithm
	if err := k.validateKeyStretchingParams(params, algorithm, iterations, memoryCost, parallelism, keyLength); err != nil {
		return "", nil, err
	}

	// Generate random salt
	salt := make([]byte, params.MinSaltLengthBytes)
	_, err = rand.Read(salt)
	if err != nil {
		return "", nil, types.ErrRandomSourceFailed
	}

	// Generate config ID
	configID := fmt.Sprintf("ksc_%s_%d", algorithm.String(), time.Now().Unix())

	now := time.Now()
	config := &cryptoproto.KeyStretchingConfig{
		ConfigId:    configID,
		Algorithm:   algorithm,
		Iterations:  iterations,
		MemoryCost:  memoryCost,
		Parallelism: parallelism,
		KeyLength:   keyLength,
		Salt:        salt,
		CreatedAt:   now,
	}

	// Store in state
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(config)
	store.Set(types.GetKeyStretchingConfigKey(configID), bz)

	k.Logger(ctx).Info("created key stretching config",
		"config_id", configID,
		"algorithm", algorithm.String(),
		"iterations", iterations,
	)

	return configID, salt, nil
}

// StretchKey performs key stretching on a password/key
func (k Keeper) StretchKey(
	ctx context.Context,
	configID string,
	password []byte,
) ([]byte, error) {
	// Retrieve configuration
	store := k.getStore(ctx)
	bz := store.Get(types.GetKeyStretchingConfigKey(configID))
	if bz == nil {
		return nil, types.ErrInvalidKeyStretchingConfig
	}

	var config cryptoproto.KeyStretchingConfig
	k.cdc.MustUnmarshal(bz, &config)

	// Perform key stretching
	return k.performKeyStretching(password, &config)
}

// StretchKeyWithParams performs key stretching with custom parameters
func (k Keeper) StretchKeyWithParams(
	ctx context.Context,
	password []byte,
	salt []byte,
	algorithm cryptoproto.KeyStretchingAlgorithm,
	iterations int32,
	memoryCost int32,
	parallelism int32,
	keyLength int32,
) ([]byte, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	// Validate parameters
	if err := k.validateKeyStretchingParams(params, algorithm, iterations, memoryCost, parallelism, keyLength); err != nil {
		return nil, err
	}

	config := &cryptoproto.KeyStretchingConfig{
		Algorithm:   algorithm,
		Iterations:  iterations,
		MemoryCost:  memoryCost,
		Parallelism: parallelism,
		KeyLength:   keyLength,
		Salt:        salt,
	}

	return k.performKeyStretching(password, config)
}

// performKeyStretching performs the actual key stretching
func (k Keeper) performKeyStretching(
	password []byte,
	config *cryptoproto.KeyStretchingConfig,
) ([]byte, error) {
	switch config.Algorithm {
	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA256:
		return k.stretchPBKDF2SHA256(password, config)
	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA512:
		return k.stretchPBKDF2SHA512(password, config)
	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2I:
		return k.stretchArgon2i(password, config)
	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2D:
		return k.stretchArgon2d(password, config)
	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2ID:
		return k.stretchArgon2id(password, config)
	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_SCRYPT:
		return k.stretchScrypt(password, config)
	default:
		return nil, types.ErrInvalidKeyStretchingConfig
	}
}

// stretchPBKDF2SHA256 stretches key using PBKDF2 with SHA-256
func (k Keeper) stretchPBKDF2SHA256(password []byte, config *cryptoproto.KeyStretchingConfig) ([]byte, error) {
	return pbkdf2.Key(password, config.Salt, int(config.Iterations), int(config.KeyLength), sha256.New), nil
}

// stretchPBKDF2SHA512 stretches key using PBKDF2 with SHA-512
func (k Keeper) stretchPBKDF2SHA512(password []byte, config *cryptoproto.KeyStretchingConfig) ([]byte, error) {
	return pbkdf2.Key(password, config.Salt, int(config.Iterations), int(config.KeyLength), sha512.New), nil
}

// stretchArgon2i stretches key using Argon2i (optimized against side-channel attacks)
func (k Keeper) stretchArgon2i(password []byte, config *cryptoproto.KeyStretchingConfig) ([]byte, error) {
	// Argon2i parameters:
	// time: iterations
	// memory: memory cost in KB
	// threads: parallelism
	// keyLen: output key length
	return argon2.Key(
		password,
		config.Salt,
		uint32(config.Iterations),
		uint32(config.MemoryCost),
		uint8(config.Parallelism),
		uint32(config.KeyLength),
	), nil
}

// stretchArgon2d stretches key using Argon2d (optimized against GPU attacks)
func (k Keeper) stretchArgon2d(password []byte, config *cryptoproto.KeyStretchingConfig) ([]byte, error) {
	// Note: crypto/argon2 doesn't have separate Argon2d function
	// Using IDKey as a placeholder (in real implementation, use proper Argon2d)
	return argon2.IDKey(
		password,
		config.Salt,
		uint32(config.Iterations),
		uint32(config.MemoryCost),
		uint8(config.Parallelism),
		uint32(config.KeyLength),
	), nil
}

// stretchArgon2id stretches key using Argon2id (hybrid, recommended)
func (k Keeper) stretchArgon2id(password []byte, config *cryptoproto.KeyStretchingConfig) ([]byte, error) {
	return argon2.IDKey(
		password,
		config.Salt,
		uint32(config.Iterations),
		uint32(config.MemoryCost),
		uint8(config.Parallelism),
		uint32(config.KeyLength),
	), nil
}

// stretchScrypt stretches key using scrypt
func (k Keeper) stretchScrypt(password []byte, config *cryptoproto.KeyStretchingConfig) ([]byte, error) {
	// scrypt parameters:
	// N: CPU/memory cost (use 2^iterations)
	// r: block size (use parallelism)
	// p: parallelization (use parallelism)
	N := 1 << uint(config.Iterations)
	r := int(config.Parallelism)
	p := int(config.Parallelism)

	return scrypt.Key(password, config.Salt, N, r, p, int(config.KeyLength))
}

// validateKeyStretchingParams validates key stretching parameters
func (k Keeper) validateKeyStretchingParams(
	params *cryptoproto.Params,
	algorithm cryptoproto.KeyStretchingAlgorithm,
	iterations int32,
	memoryCost int32,
	parallelism int32,
	keyLength int32,
) error {
	switch algorithm {
	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA256,
		cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA512:
		if iterations < params.MinPbkdf2Iterations {
			return types.ErrInvalidIterationCount
		}

	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2I,
		cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2D,
		cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2ID:
		if iterations < params.MinArgon2Iterations {
			return types.ErrInvalidIterationCount
		}
		if memoryCost < params.MinArgon2MemoryKb {
			return types.ErrInvalidKeyStretchingConfig
		}
		if parallelism < 1 {
			return types.ErrInvalidKeyStretchingConfig
		}

	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_SCRYPT:
		if iterations < 14 { // Minimum N=2^14 for scrypt
			return types.ErrInvalidIterationCount
		}
		if parallelism < 1 {
			return types.ErrInvalidKeyStretchingConfig
		}
	}

	if keyLength < 16 {
		return types.ErrInvalidKeyStretchingConfig
	}

	return nil
}

// GetRecommendedStretchingConfig returns recommended key stretching configuration
func (k Keeper) GetRecommendedStretchingConfig(
	ctx context.Context,
	algorithm cryptoproto.KeyStretchingAlgorithm,
) (*cryptoproto.KeyStretchingConfig, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	salt := make([]byte, params.MinSaltLengthBytes)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, types.ErrRandomSourceFailed
	}

	var config *cryptoproto.KeyStretchingConfig

	switch algorithm {
	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA256:
		config = &cryptoproto.KeyStretchingConfig{
			Algorithm:  algorithm,
			Iterations: 100000, // OWASP recommendation
			KeyLength:  32,
			Salt:       salt,
		}

	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_PBKDF2_SHA512:
		config = &cryptoproto.KeyStretchingConfig{
			Algorithm:  algorithm,
			Iterations: 100000,
			KeyLength:  64,
			Salt:       salt,
		}

	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_ARGON2ID:
		config = &cryptoproto.KeyStretchingConfig{
			Algorithm:   algorithm,
			Iterations:  3,     // time cost
			MemoryCost:  65536, // 64 MB
			Parallelism: 4,     // threads
			KeyLength:   32,
			Salt:        salt,
		}

	case cryptoproto.KeyStretchingAlgorithm_KEY_STRETCHING_ALGORITHM_SCRYPT:
		config = &cryptoproto.KeyStretchingConfig{
			Algorithm:   algorithm,
			Iterations:  15, // N=2^15
			Parallelism: 8,  // r and p
			KeyLength:   32,
			Salt:        salt,
		}

	default:
		return nil, types.ErrInvalidKeyStretchingConfig
	}

	return config, nil
}

// VerifyStretchedKey verifies a password against a stretched key
func (k Keeper) VerifyStretchedKey(
	ctx context.Context,
	configID string,
	password []byte,
	expectedKey []byte,
) (bool, error) {
	derivedKey, err := k.StretchKey(ctx, configID, password)
	if err != nil {
		return false, err
	}

	return k.CompareHashes(derivedKey, expectedKey), nil
}

// SetKeyStretchingConfig stores a key stretching config (for genesis)
func (k *Keeper) SetKeyStretchingConfig(ctx context.Context, config *cryptoproto.KeyStretchingConfig) error {
	if config == nil {
		return nil
	}
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(config)
	store.Set(types.GetKeyStretchingConfigKey(config.ConfigId), bz)
	return nil
}
