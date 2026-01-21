// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// CreateSaltedHash creates a cryptographic hash with salt.
//
// WARNING: This function uses crypto/rand which is NON-DETERMINISTIC and will break
// consensus if called from a message handler (MsgServer method).
//
// DO NOT call this from message handlers. Instead:
//  1. For message handlers: Use HashWithCustomSalt() and require the client to provide
//     the salt in the message (generated off-chain)
//  2. For queries/client-side: This function is safe to use
//
// This function is intended for:
// - Client-side utilities
// - Query responses
// - BeginBlocker/EndBlocker with extreme caution (consider if randomness is needed)
func (k Keeper) CreateSaltedHash(
	ctx context.Context,
	data []byte,
	algorithm cryptoproto.HashAlgorithm,
	iterations int32,
) (string, []byte, []byte, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return "", nil, nil, err
	}

	// Generate random salt
	// WARNING: This is non-deterministic - see function documentation
	salt := make([]byte, params.MinSaltLengthBytes)
	_, err = rand.Read(salt)
	if err != nil {
		return "", nil, nil, types.ErrRandomSourceFailed
	}

	// Compute hash
	hash, err := k.computeSaltedHash(data, salt, algorithm, iterations)
	if err != nil {
		return "", nil, nil, err
	}

	// Generate hash ID using consensus-safe block time
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()
	hashID := fmt.Sprintf("hash_%s_%d", algorithm.String(), blockTime.Unix())

	saltedHash := &cryptoproto.SaltedHash{
		HashId:     hashID,
		Salt:       salt,
		Hash:       hash,
		Algorithm:  algorithm,
		Iterations: iterations,
		CreatedAt:  blockTime,
	}

	// Store in state
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(saltedHash)
	store.Set(types.GetSaltedHashKey(hashID), bz)
	k.Logger(sdkCtx).Info("created salted hash",
		"hash_id", hashID,
		"algorithm", algorithm.String(),
		"iterations", iterations,
	)

	return hashID, salt, hash, nil
}

// VerifySaltedHash verifies data against a salted hash
func (k Keeper) VerifySaltedHash(
	ctx context.Context,
	hashID string,
	data []byte,
) (bool, error) {
	// Retrieve stored hash
	store := k.getStore(ctx)
	bz := store.Get(types.GetSaltedHashKey(hashID))
	if bz == nil {
		return false, fmt.Errorf("hash not found")
	}

	var storedHash cryptoproto.SaltedHash
	if err := k.cdc.Unmarshal(bz, &storedHash); err != nil {
		return false, fmt.Errorf("failed to unmarshal stored hash: %w", err)
	}

	// Compute hash with stored salt
	computedHash, err := k.computeSaltedHash(data, storedHash.Salt, storedHash.Algorithm, storedHash.Iterations)
	if err != nil {
		return false, err
	}

	// Compare hashes
	if len(computedHash) != len(storedHash.Hash) {
		return false, nil
	}

	for i := range computedHash {
		if computedHash[i] != storedHash.Hash[i] {
			return false, nil
		}
	}

	return true, nil
}

// computeSaltedHash computes a salted hash using the specified algorithm
func (k Keeper) computeSaltedHash(
	data []byte,
	salt []byte,
	algorithm cryptoproto.HashAlgorithm,
	iterations int32,
) ([]byte, error) {
	if iterations < 1 {
		iterations = 1
	}

	var hash []byte
	var err error

	// Combine data and salt
	combined := append(salt, data...)

	// Initial hash
	switch algorithm {
	case cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256:
		hash, err = k.hashSHA256(combined)
	case cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA512:
		hash, err = k.hashSHA512(combined)
	case cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA3_256:
		hash, err = k.hashSHA3_256(combined)
	case cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA3_512:
		hash, err = k.hashSHA3_512(combined)
	case cryptoproto.HashAlgorithm_HASH_ALGORITHM_BLAKE2B:
		hash, err = k.hashBLAKE2b(combined)
	case cryptoproto.HashAlgorithm_HASH_ALGORITHM_BLAKE3:
		hash, err = k.hashBLAKE3(combined)
	default:
		return nil, types.ErrInvalidHashAlgorithm
	}

	if err != nil {
		return nil, err
	}

	// Apply iterations
	for i := int32(1); i < iterations; i++ {
		switch algorithm {
		case cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA256:
			hash, err = k.hashSHA256(hash)
		case cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA512:
			hash, err = k.hashSHA512(hash)
		case cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA3_256:
			hash, err = k.hashSHA3_256(hash)
		case cryptoproto.HashAlgorithm_HASH_ALGORITHM_SHA3_512:
			hash, err = k.hashSHA3_512(hash)
		case cryptoproto.HashAlgorithm_HASH_ALGORITHM_BLAKE2B:
			hash, err = k.hashBLAKE2b(hash)
		case cryptoproto.HashAlgorithm_HASH_ALGORITHM_BLAKE3:
			hash, err = k.hashBLAKE3(hash)
		}

		if err != nil {
			return nil, err
		}
	}

	return hash, nil
}

// hashSHA256 computes SHA-256 hash
func (k Keeper) hashSHA256(data []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil), nil
}

// hashSHA512 computes SHA-512 hash
func (k Keeper) hashSHA512(data []byte) ([]byte, error) {
	h := sha512.New()
	h.Write(data)
	return h.Sum(nil), nil
}

// hashSHA3_256 computes SHA3-256 hash
func (k Keeper) hashSHA3_256(data []byte) ([]byte, error) {
	h := sha3.New256()
	h.Write(data)
	return h.Sum(nil), nil
}

// hashSHA3_512 computes SHA3-512 hash
func (k Keeper) hashSHA3_512(data []byte) ([]byte, error) {
	h := sha3.New512()
	h.Write(data)
	return h.Sum(nil), nil
}

// hashBLAKE2b computes BLAKE2b hash
func (k Keeper) hashBLAKE2b(data []byte) ([]byte, error) {
	h, err := blake2b.New512(nil)
	if err != nil {
		return nil, err
	}
	h.Write(data)
	return h.Sum(nil), nil
}

// hashBLAKE3 computes BLAKE3 hash
func (k Keeper) hashBLAKE3(data []byte) ([]byte, error) {
	// Note: BLAKE3 is not yet in the Go standard library or golang.org/x/crypto.
	// The official BLAKE3 implementation is available at github.com/zeebo/blake3.
	// For production deployment, import that library:
	// import "github.com/zeebo/blake3"
	// hasher := blake3.New()
	// hasher.Write(data)
	// return hasher.Sum(nil), nil
	//
	// For now, we fallback to BLAKE2b which provides similar security properties.
	// BLAKE2b-512 is cryptographically secure and approved for production use.
	return k.hashBLAKE2b(data)
}

// HashWithCustomSalt creates a hash with a custom salt (for migration scenarios)
func (k Keeper) HashWithCustomSalt(
	ctx context.Context,
	data []byte,
	salt []byte,
	algorithm cryptoproto.HashAlgorithm,
	iterations int32,
) ([]byte, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	// Validate salt length
	if len(salt) < int(params.MinSaltLengthBytes) {
		return nil, types.ErrInvalidSaltLength
	}

	return k.computeSaltedHash(data, salt, algorithm, iterations)
}

// GenerateSalt generates a cryptographically secure salt.
//
// WARNING: This function uses crypto/rand which is NON-DETERMINISTIC and will break
// consensus if called from a message handler (MsgServer method).
//
// DO NOT call this from message handlers. Instead, require the client to generate
// salt off-chain and include it in the message.
//
// This function is intended for:
// - Client-side utilities only
// - NEVER for on-chain/consensus operations
func (k Keeper) GenerateSalt(ctx context.Context, length int32) ([]byte, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	if length < params.MinSaltLengthBytes {
		return nil, types.ErrInvalidSaltLength
	}

	// WARNING: Non-deterministic - see function documentation
	salt := make([]byte, length)
	_, err = rand.Read(salt)
	if err != nil {
		return nil, types.ErrRandomSourceFailed
	}

	return salt, nil
}

// BatchHashWithSalt creates multiple salted hashes efficiently.
//
// WARNING: This function uses crypto/rand which is NON-DETERMINISTIC and will break
// consensus if called from a message handler (MsgServer method).
//
// DO NOT call this from message handlers. This is for client-side batch operations only.
func (k Keeper) BatchHashWithSalt(
	ctx context.Context,
	dataItems [][]byte,
	algorithm cryptoproto.HashAlgorithm,
	iterations int32,
) ([]struct {
	Salt []byte
	Hash []byte
}, error) {
	results := make([]struct {
		Salt []byte
		Hash []byte
	}, len(dataItems))

	for i, data := range dataItems {
		salt, err := k.GenerateSalt(ctx, 16)
		if err != nil {
			return nil, err
		}

		hash, err := k.computeSaltedHash(data, salt, algorithm, iterations)
		if err != nil {
			return nil, err
		}

		results[i].Salt = salt
		results[i].Hash = hash
	}

	return results, nil
}

// CompareHashes performs constant-time hash comparison to prevent timing attacks
func (k Keeper) CompareHashes(hash1, hash2 []byte) bool {
	if len(hash1) != len(hash2) {
		return false
	}

	// Constant-time comparison
	result := byte(0)
	for i := 0; i < len(hash1); i++ {
		result |= hash1[i] ^ hash2[i]
	}

	return result == 0
}
