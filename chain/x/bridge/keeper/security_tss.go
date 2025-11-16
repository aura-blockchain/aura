package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/aequitas/aura/chain/x/bridge/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// THRESHOLD SIGNATURE SCHEME (TSS) INTEGRATION
// ============================================================================

// VerifyTSSSignature verifies a threshold signature from bridge validators
// This ensures that enough validators have approved a cross-chain transfer
func (k Keeper) VerifyTSSSignature(
	ctx sdk.Context,
	message []byte,
	signature *types.TSSSignature,
) error {
	if signature == nil {
		return fmt.Errorf("TSS signature cannot be nil")
	}

	params := k.GetParams(ctx)

	// Verify we have enough signers
	if uint64(len(signature.Signers)) < signature.Threshold {
		return fmt.Errorf(
			"insufficient signers: got %d, required %d",
			len(signature.Signers),
			signature.Threshold,
		)
	}

	// Verify threshold is valid
	if signature.Threshold < params.MinValidatorSignatures {
		return fmt.Errorf(
			"threshold too low: got %d, minimum %d",
			signature.Threshold,
			params.MinValidatorSignatures,
		)
	}

	// Verify all signers are active validators
	totalPower := uint64(0)
	for _, signer := range signature.Signers {
		validator := k.GetBridgeValidator(ctx, signer)
		if validator == nil {
			return fmt.Errorf("signer is not a registered validator: %s", signer)
		}
		if !validator.Active {
			return fmt.Errorf("signer is not an active validator: %s", signer)
		}
		totalPower += validator.Power
	}

	// Verify signature voting power threshold (2/3+ of total power)
	allValidators := k.GetAllBridgeValidators(ctx)
	totalValidatorPower := uint64(0)
	for _, v := range allValidators {
		if v.Active {
			totalValidatorPower += v.Power
		}
	}

	requiredPower := (totalValidatorPower * 2) / 3
	if totalPower < requiredPower {
		return fmt.Errorf(
			"insufficient voting power: got %d, required %d",
			totalPower,
			requiredPower,
		)
	}

	// Verify nonce for replay protection
	expectedNonce := k.GetTSSNonce(ctx)
	if signature.Nonce != expectedNonce {
		return fmt.Errorf(
			"invalid nonce: got %d, expected %d",
			signature.Nonce,
			expectedNonce,
		)
	}

	// Verify the TSS signature cryptographically
	// Reconstruct the combined public key from validator public keys
	combinedPubKey, err := k.reconstructTSSPublicKey(ctx, signature.Signers)
	if err != nil {
		return fmt.Errorf("failed to reconstruct TSS public key: %w", err)
	}

	// Verify the aggregated signature
	if err := k.verifyTSSSignatureCrypto(message, signature.Signature, combinedPubKey); err != nil {
		return fmt.Errorf("TSS signature verification failed: %w", err)
	}

	// Increment nonce after successful verification
	k.SetTSSNonce(ctx, expectedNonce+1)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"tss_signature_verified",
			sdk.NewAttribute("signers_count", fmt.Sprintf("%d", len(signature.Signers))),
			sdk.NewAttribute("threshold", fmt.Sprintf("%d", signature.Threshold)),
			sdk.NewAttribute("total_power", fmt.Sprintf("%d", totalPower)),
			sdk.NewAttribute("nonce", fmt.Sprintf("%d", signature.Nonce)),
		),
	)

	return nil
}

// CreateTSSSignature creates a TSS signature structure for validators to sign
// This is typically called when initiating a cross-chain transfer
func (k Keeper) CreateTSSSignature(
	ctx sdk.Context,
	message []byte,
	signers []string,
) (*types.TSSSignature, error) {
	params := k.GetParams(ctx)

	// Verify we have enough signers
	if uint64(len(signers)) < params.MinValidatorSignatures {
		return nil, fmt.Errorf(
			"insufficient signers: got %d, required %d",
			len(signers),
			params.MinValidatorSignatures,
		)
	}

	// Verify all signers are active validators
	totalPower := uint64(0)
	for _, signer := range signers {
		validator := k.GetBridgeValidator(ctx, signer)
		if validator == nil {
			return nil, fmt.Errorf("signer is not a registered validator: %s", signer)
		}
		if !validator.Active {
			return nil, fmt.Errorf("signer is not an active validator: %s", signer)
		}
		totalPower += validator.Power
	}

	// Get current nonce
	nonce := k.GetTSSNonce(ctx)

	// Create message hash
	messageHash := sha256.Sum256(message)

	// In production, this would create an actual TSS signature
	// For now, we create the structure for validators to sign
	signature := &types.TSSSignature{
		Signature:         messageHash[:], // Placeholder
		PublicKey:         []byte{},       // Would be the combined public key
		Signers:           signers,
		Threshold:         params.MinValidatorSignatures,
		TotalParticipants: uint64(len(k.GetAllBridgeValidators(ctx))),
		Nonce:             nonce,
	}

	return signature, nil
}

// GetTSSNonce retrieves the current TSS nonce for replay protection
func (k Keeper) GetTSSNonce(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TSSNonceKey())

	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetTSSNonce sets the TSS nonce
func (k Keeper) SetTSSNonce(ctx sdk.Context, nonce uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := sdk.Uint64ToBigEndian(nonce)
	store.Set(types.TSSNonceKey(), bz)
}

// VerifyValidatorSignature verifies an individual validator's signature
func (k Keeper) VerifyValidatorSignature(
	ctx sdk.Context,
	message []byte,
	validatorAddr string,
	signature []byte,
) error {
	validator := k.GetBridgeValidator(ctx, validatorAddr)
	if validator == nil {
		return fmt.Errorf("validator not found: %s", validatorAddr)
	}

	if !validator.Active {
		return fmt.Errorf("validator is not active: %s", validatorAddr)
	}

	if len(signature) == 0 {
		return fmt.Errorf("empty signature")
	}

	// Get validator's public key
	if len(validator.PublicKey) == 0 {
		return fmt.Errorf("validator has no public key: %s", validatorAddr)
	}

	// Parse public key and verify signature
	pubKey, err := k.parsePublicKey(validator.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	// Hash the message
	messageHash := sha256.Sum256(message)

	// Verify signature
	if !pubKey.VerifySignature(messageHash[:], signature) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// reconstructTSSPublicKey reconstructs the combined TSS public key from validator keys
func (k Keeper) reconstructTSSPublicKey(ctx sdk.Context, signers []string) (cryptotypes.PubKey, error) {
	if len(signers) == 0 {
		return nil, fmt.Errorf("no signers provided")
	}

	// For threshold signatures, we need to combine the public keys
	// This is a simplified implementation - real TSS would use proper key aggregation
	var pubKeys []cryptotypes.PubKey
	for _, signer := range signers {
		validator := k.GetBridgeValidator(ctx, signer)
		if validator == nil {
			return nil, fmt.Errorf("signer not found: %s", signer)
		}

		if len(validator.PublicKey) == 0 {
			return nil, fmt.Errorf("validator %s has no public key", signer)
		}

		pubKey, err := k.parsePublicKey(validator.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key for %s: %w", signer, err)
		}

		pubKeys = append(pubKeys, pubKey)
	}

	// For now, use the first validator's key as representative
	// Real TSS implementation would aggregate keys properly using BLS or Schnorr signatures
	if len(pubKeys) > 0 {
		return pubKeys[0], nil
	}

	return nil, fmt.Errorf("no valid public keys found")
}

// verifyTSSSignatureCrypto verifies the TSS signature cryptographically
func (k Keeper) verifyTSSSignatureCrypto(message, signature []byte, pubKey cryptotypes.PubKey) error {
	if pubKey == nil {
		return fmt.Errorf("public key is nil")
	}

	if len(signature) == 0 {
		return fmt.Errorf("signature is empty")
	}

	// Hash the message
	messageHash := sha256.Sum256(message)

	// Verify signature
	if !pubKey.VerifySignature(messageHash[:], signature) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// parsePublicKey parses a public key from bytes
func (k Keeper) parsePublicKey(pubKeyBytes []byte) (cryptotypes.PubKey, error) {
	if len(pubKeyBytes) == 0 {
		return nil, fmt.Errorf("empty public key")
	}

	// Try secp256k1 first (33 bytes compressed)
	if len(pubKeyBytes) == 33 {
		return &secp256k1.PubKey{Key: pubKeyBytes}, nil
	}

	// Try ed25519 (32 bytes)
	if len(pubKeyBytes) == 32 {
		return &ed25519.PubKey{Key: pubKeyBytes}, nil
	}

	// For other lengths, try secp256k1
	return &secp256k1.PubKey{Key: pubKeyBytes}, nil
}

// AggregateValidatorSignatures aggregates multiple validator signatures into a TSS signature
func (k Keeper) AggregateValidatorSignatures(
	ctx sdk.Context,
	transferId string,
	signatures []*types.ValidatorSignature,
) (*types.TSSSignature, error) {
	params := k.GetParams(ctx)

	if uint64(len(signatures)) < params.MinValidatorSignatures {
		return nil, fmt.Errorf(
			"insufficient signatures: got %d, required %d",
			len(signatures),
			params.MinValidatorSignatures,
		)
	}

	// Extract signer addresses
	signers := make([]string, len(signatures))
	for i, sig := range signatures {
		signers[i] = sig.ValidatorAddress
	}

	// Create TSS signature structure
	// In production, this would perform actual signature aggregation
	tssSignature := &types.TSSSignature{
		Signature:         []byte(transferId), // Placeholder
		PublicKey:         []byte{},
		Signers:           signers,
		Threshold:         params.MinValidatorSignatures,
		TotalParticipants: uint64(len(k.GetAllBridgeValidators(ctx))),
		Nonce:             k.GetTSSNonce(ctx),
	}

	return tssSignature, nil
}
