package keeper

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// GenerateQuantumResistantKey generates a quantum-resistant cryptographic key pair
func (k Keeper) GenerateQuantumResistantKey(
	ctx context.Context,
	creator string,
	algorithm cryptoproto.QuantumResistantAlgorithm,
	expiresAt *time.Time,
) (string, []byte, error) {
	if algorithm == cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_UNSPECIFIED {
		return "", nil, types.ErrInvalidQuantumAlgorithm
	}

	// Generate key pair based on algorithm
	var publicKey []byte
	var keyMetadata []byte
	var err error

	switch algorithm {
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM:
		publicKey, keyMetadata, err = k.generateDilithiumKey()
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER:
		publicKey, keyMetadata, err = k.generateKyberKey()
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_FALCON:
		publicKey, keyMetadata, err = k.generateFalconKey()
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_SPHINCS_PLUS:
		publicKey, keyMetadata, err = k.generateSPHINCSPlusKey()
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_NTRU:
		publicKey, keyMetadata, err = k.generateNTRUKey()
	default:
		return "", nil, types.ErrInvalidQuantumAlgorithm
	}

	if err != nil {
		return "", nil, err
	}

	// Generate key ID
	blockTime := sdk.UnwrapSDKContext(ctx).BlockTime()
	keyID := fmt.Sprintf("qr_%s_%d", algorithm.String(), blockTime.Unix())

	qrKey := &cryptoproto.QuantumResistantKey{
		KeyId:       keyID,
		Algorithm:   algorithm,
		PublicKey:   publicKey,
		KeyMetadata: keyMetadata,
		CreatedAt:   blockTime,
		ExpiresAt:   expiresAt,
	}

	// Store in KV store
	if err := k.SetQuantumResistantKey(ctx, qrKey); err != nil {
		return "", nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Logger(sdkCtx).Info("generated quantum-resistant key",
		"key_id", keyID,
		"algorithm", algorithm.String(),
	)

	return keyID, publicKey, nil
}

// Note: GetQuantumResistantKey is now implemented in keeper.go using KV store

// generateDilithiumKey generates a CRYSTALS-Dilithium key pair (NIST PQC standard)
func (k Keeper) generateDilithiumKey() ([]byte, []byte, error) {
	// CRYSTALS-Dilithium2 public key is 1312 bytes
	// In a real implementation, use the official Dilithium library
	publicKey := make([]byte, 1312)
	_, err := rand.Read(publicKey)
	if err != nil {
		return nil, nil, err
	}

	metadata := []byte("dilithium2")
	return publicKey, metadata, nil
}

// generateKyberKey generates a CRYSTALS-Kyber key pair (NIST PQC standard for KEM)
func (k Keeper) generateKyberKey() ([]byte, []byte, error) {
	// CRYSTALS-Kyber512 public key is 800 bytes
	// In a real implementation, use the official Kyber library
	publicKey := make([]byte, 800)
	_, err := rand.Read(publicKey)
	if err != nil {
		return nil, nil, err
	}

	metadata := []byte("kyber512")
	return publicKey, metadata, nil
}

// generateFalconKey generates a Falcon key pair (NIST PQC standard)
func (k Keeper) generateFalconKey() ([]byte, []byte, error) {
	// Falcon-512 public key is 897 bytes
	// In a real implementation, use the official Falcon library
	publicKey := make([]byte, 897)
	_, err := rand.Read(publicKey)
	if err != nil {
		return nil, nil, err
	}

	metadata := []byte("falcon512")
	return publicKey, metadata, nil
}

// generateSPHINCSPlusKey generates a SPHINCS+ key pair (NIST PQC standard)
func (k Keeper) generateSPHINCSPlusKey() ([]byte, []byte, error) {
	// SPHINCS+-128s public key is 32 bytes
	// In a real implementation, use the official SPHINCS+ library
	publicKey := make([]byte, 32)
	_, err := rand.Read(publicKey)
	if err != nil {
		return nil, nil, err
	}

	metadata := []byte("sphincs-sha256-128s-simple")
	return publicKey, metadata, nil
}

// generateNTRUKey generates an NTRU key pair
func (k Keeper) generateNTRUKey() ([]byte, []byte, error) {
	// NTRU public key size varies, typically around 1230 bytes for NTRU-HPS-2048-509
	// In a real implementation, use an NTRU library
	publicKey := make([]byte, 1230)
	_, err := rand.Read(publicKey)
	if err != nil {
		return nil, nil, err
	}

	metadata := []byte("ntru-hps-2048-509")
	return publicKey, metadata, nil
}

// ValidateQuantumResistantKey validates a quantum-resistant key
func (k Keeper) ValidateQuantumResistantKey(ctx context.Context, keyID string) error {
	key, err := k.GetQuantumResistantKey(ctx, keyID)
	if err != nil {
		return err
	}

	// Check expiration
	blockTime := sdk.UnwrapSDKContext(ctx).BlockTime()
	if key.ExpiresAt != nil && !key.ExpiresAt.IsZero() && key.ExpiresAt.Before(blockTime) {
		return types.ErrKeyExpired
	}

	// Validate public key length based on algorithm
	var expectedMinLength int
	switch key.Algorithm {
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM:
		expectedMinLength = 1000
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER:
		expectedMinLength = 700
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_FALCON:
		expectedMinLength = 800
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_SPHINCS_PLUS:
		expectedMinLength = 32
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_NTRU:
		expectedMinLength = 1000
	}

	if len(key.PublicKey) < expectedMinLength {
		return fmt.Errorf("invalid public key length for algorithm")
	}

	return nil
}

// RotateQuantumResistantKey rotates a quantum-resistant key
func (k Keeper) RotateQuantumResistantKey(
	ctx context.Context,
	keyID string,
	newExpiresAt *time.Time,
) (string, []byte, error) {
	oldKey, err := k.GetQuantumResistantKey(ctx, keyID)
	if err != nil {
		return "", nil, err
	}

	// Generate new key with same algorithm
	newKeyID, newPublicKey, err := k.GenerateQuantumResistantKey(
		ctx,
		"system",
		oldKey.Algorithm,
		newExpiresAt,
	)
	if err != nil {
		return "", nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Logger(sdkCtx).Info("rotated quantum-resistant key",
		"old_key_id", keyID,
		"new_key_id", newKeyID,
		"algorithm", oldKey.Algorithm.String(),
	)

	return newKeyID, newPublicKey, nil
}

// Note: SetQuantumResistantKey is now implemented in keeper.go using KV store
