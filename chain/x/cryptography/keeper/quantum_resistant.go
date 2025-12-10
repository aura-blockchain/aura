package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// RegisterQuantumResistantKey registers a quantum-resistant public key that was generated off-chain.
//
// IMPORTANT: This function does NOT generate keys on-chain. Key generation MUST happen
// client-side to ensure:
// 1. Deterministic blockchain execution (crypto/rand breaks consensus)
// 2. Private key security (never expose private keys to validators)
// 3. Client control over entropy sources
//
// The client should:
// 1. Generate the quantum-resistant keypair off-chain using appropriate libraries
// 2. Submit only the PUBLIC key via MsgGenerateQuantumResistantKey
// 3. Store the PRIVATE key securely client-side
//
// This function validates and stores the public key on-chain for verification purposes.
func (k Keeper) RegisterQuantumResistantKey(
	ctx context.Context,
	creator string,
	algorithm cryptoproto.QuantumResistantAlgorithm,
	publicKey []byte,
	expiresAt *time.Time,
) (string, error) {
	if algorithm == cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_UNSPECIFIED {
		return "", types.ErrInvalidQuantumAlgorithm
	}

	// Validate public key based on algorithm
	if err := k.validateQuantumPublicKey(algorithm, publicKey); err != nil {
		return "", err
	}

	// Generate key ID using deterministic block time
	blockTime := sdk.UnwrapSDKContext(ctx).BlockTime()
	keyID := fmt.Sprintf("qr_%s_%d", algorithm.String(), blockTime.Unix())

	// Create metadata based on algorithm
	keyMetadata := k.getQuantumKeyMetadata(algorithm)

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
		return "", err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Logger(sdkCtx).Info("registered quantum-resistant key",
		"key_id", keyID,
		"algorithm", algorithm.String(),
		"creator", creator,
	)

	return keyID, nil
}

// validateQuantumPublicKey validates the public key format for the given algorithm
func (k Keeper) validateQuantumPublicKey(algorithm cryptoproto.QuantumResistantAlgorithm, publicKey []byte) error {
	var expectedMinLength, expectedMaxLength int

	switch algorithm {
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM:
		// CRYSTALS-Dilithium2 public key is 1312 bytes
		expectedMinLength = 1312
		expectedMaxLength = 1312
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER:
		// CRYSTALS-Kyber512 public key is 800 bytes
		expectedMinLength = 800
		expectedMaxLength = 800
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_FALCON:
		// Falcon-512 public key is 897 bytes
		expectedMinLength = 897
		expectedMaxLength = 897
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_SPHINCS_PLUS:
		// SPHINCS+-128s public key is 32 bytes
		expectedMinLength = 32
		expectedMaxLength = 32
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_NTRU:
		// NTRU-HPS-2048-509 public key is 1230 bytes
		expectedMinLength = 1230
		expectedMaxLength = 1230
	default:
		return types.ErrInvalidQuantumAlgorithm
	}

	if len(publicKey) < expectedMinLength {
		return fmt.Errorf("public key too short for %s: got %d bytes, expected %d",
			algorithm.String(), len(publicKey), expectedMinLength)
	}
	if len(publicKey) > expectedMaxLength {
		return fmt.Errorf("public key too long for %s: got %d bytes, expected %d",
			algorithm.String(), len(publicKey), expectedMaxLength)
	}

	return nil
}

// getQuantumKeyMetadata returns the metadata descriptor for a quantum algorithm
func (k Keeper) getQuantumKeyMetadata(algorithm cryptoproto.QuantumResistantAlgorithm) []byte {
	switch algorithm {
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM:
		return []byte("dilithium2")
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER:
		return []byte("kyber512")
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_FALCON:
		return []byte("falcon512")
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_SPHINCS_PLUS:
		return []byte("sphincs-sha256-128s-simple")
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_NTRU:
		return []byte("ntru-hps-2048-509")
	default:
		return []byte("unknown")
	}
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

// RotateQuantumResistantKey rotates a quantum-resistant key with a new public key.
//
// IMPORTANT: The new public key MUST be generated off-chain by the client.
// This function does NOT generate keys - it registers the client-provided public key.
//
// Parameters:
//   - keyID: The ID of the key being rotated
//   - newPublicKey: The new public key (generated client-side)
//   - newExpiresAt: Optional expiration time for the new key
//
// The client should:
// 1. Generate a new quantum-resistant keypair off-chain using the same algorithm
// 2. Call this function with the new PUBLIC key
// 3. Store the new PRIVATE key securely client-side
func (k Keeper) RotateQuantumResistantKey(
	ctx context.Context,
	keyID string,
	newPublicKey []byte,
	newExpiresAt *time.Time,
) (string, error) {
	oldKey, err := k.GetQuantumResistantKey(ctx, keyID)
	if err != nil {
		return "", err
	}

	// Register the new key with the same algorithm
	newKeyID, err := k.RegisterQuantumResistantKey(
		ctx,
		"system",
		oldKey.Algorithm,
		newPublicKey,
		newExpiresAt,
	)
	if err != nil {
		return "", err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Logger(sdkCtx).Info("rotated quantum-resistant key",
		"old_key_id", keyID,
		"new_key_id", newKeyID,
		"algorithm", oldKey.Algorithm.String(),
	)

	return newKeyID, nil
}

// Note: SetQuantumResistantKey is now implemented in keeper.go using KV store
