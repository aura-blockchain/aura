// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// TestVerifySignatureWithKey_Secp256k1 tests signature verification with secp256k1 keys
func TestVerifySignatureWithKey_Secp256k1(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Generate a secp256k1 key pair
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	// Create test DID and identity
	did := "did:aura:test123"
	owner := sdk.AccAddress(pubKey.Address()).String()

	// Encode public key as hex (verification method)
	verificationMethod := hex.EncodeToString(pubKey.Bytes())

	// Create identity with verification method
	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{verificationMethod},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}
	err := k.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Test message
	message := []byte("This is a test message for signature verification")

	// Hash and sign the message
	messageHash := sha256.Sum256(message)
	signature, err := privKey.Sign(messageHash[:])
	require.NoError(t, err)

	// Test valid signature verification
	err = k.VerifySignatureWithKey(ctx, did, verificationMethod, message, signature)
	require.NoError(t, err, "Valid secp256k1 signature should verify successfully")

	// Test invalid signature (wrong message)
	wrongMessage := []byte("This is a different message")
	err = k.VerifySignatureWithKey(ctx, did, verificationMethod, wrongMessage, signature)
	require.Error(t, err, "Signature verification should fail for wrong message")
	require.Contains(t, err.Error(), "signature verification failed")

	// Test invalid signature (corrupted signature)
	corruptedSignature := make([]byte, len(signature))
	copy(corruptedSignature, signature)
	corruptedSignature[0] ^= 0xFF // Flip bits
	err = k.VerifySignatureWithKey(ctx, did, verificationMethod, message, corruptedSignature)
	require.Error(t, err, "Signature verification should fail for corrupted signature")

	// Test empty message
	err = k.VerifySignatureWithKey(ctx, did, verificationMethod, []byte{}, signature)
	require.Error(t, err, "Should reject empty message")
	require.Contains(t, err.Error(), "message cannot be empty")

	// Test empty signature
	err = k.VerifySignatureWithKey(ctx, did, verificationMethod, message, []byte{})
	require.Error(t, err, "Should reject empty signature")
	require.Contains(t, err.Error(), "signature cannot be empty")
}

// TestVerifySignatureWithKey_Ed25519 tests signature verification with ed25519 keys
func TestVerifySignatureWithKey_Ed25519(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Generate an ed25519 key pair
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	// Create test DID and identity
	did := "did:aura:test-ed25519"
	owner := sdk.AccAddress(pubKey.Address()).String()

	// Encode public key as hex (verification method)
	verificationMethod := hex.EncodeToString(pubKey.Bytes())

	// Create identity with verification method
	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{verificationMethod},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}
	err := k.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Test message
	message := []byte("Ed25519 signature test message")

	// Hash and sign the message
	messageHash := sha256.Sum256(message)
	signature, err := privKey.Sign(messageHash[:])
	require.NoError(t, err)

	// Test valid signature verification
	err = k.VerifySignatureWithKey(ctx, did, verificationMethod, message, signature)
	require.NoError(t, err, "Valid ed25519 signature should verify successfully")

	// Test invalid signature (wrong message)
	wrongMessage := []byte("Wrong message for ed25519")
	err = k.VerifySignatureWithKey(ctx, did, verificationMethod, wrongMessage, signature)
	require.Error(t, err, "Signature verification should fail for wrong message")
}

// TestVerifySignatureWithKey_RevokedCredential tests that revoked credentials are rejected
func TestVerifySignatureWithKey_RevokedCredential(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Generate a secp256k1 key pair
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	// Create test DID and identity
	did := "did:aura:revoked-test"
	owner := sdk.AccAddress(pubKey.Address()).String()

	// Encode public key as hex (verification method)
	verificationMethod := hex.EncodeToString(pubKey.Bytes())

	// Create identity with verification method
	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{verificationMethod},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}
	err := k.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Revoke the credential
	err = k.RevokeCredential(ctx, verificationMethod, did, owner, "security_compromise", map[string]string{})
	require.NoError(t, err)

	// Test message and signature
	message := []byte("Test message")
	messageHash := sha256.Sum256(message)
	signature, err := privKey.Sign(messageHash[:])
	require.NoError(t, err)

	// Verify that revoked credential is rejected
	err = k.VerifySignatureWithKey(ctx, did, verificationMethod, message, signature)
	require.Error(t, err, "Revoked credentials should be rejected")
	require.Contains(t, err.Error(), "revoked")
}

// TestVerifySignatureWithKey_InvalidKey tests that invalid verification methods are rejected
func TestVerifySignatureWithKey_InvalidKey(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Generate a secp256k1 key pair
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	// Create test DID and identity
	did := "did:aura:invalid-key-test"
	owner := sdk.AccAddress(pubKey.Address()).String()

	// Encode public key as hex (verification method)
	verificationMethod := hex.EncodeToString(pubKey.Bytes())

	// Create identity with verification method
	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{verificationMethod},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}
	err := k.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Test message and signature
	message := []byte("Test message")
	messageHash := sha256.Sum256(message)
	signature, err := privKey.Sign(messageHash[:])
	require.NoError(t, err)

	// Use a different verification method (not associated with this DID)
	otherPrivKey := secp256k1.GenPrivKey()
	otherPubKey := otherPrivKey.PubKey()
	otherVerificationMethod := hex.EncodeToString(otherPubKey.Bytes())

	// Verify that unassociated key is rejected
	err = k.VerifySignatureWithKey(ctx, did, otherVerificationMethod, message, signature)
	require.Error(t, err, "Unassociated verification method should be rejected")
	require.Contains(t, err.Error(), "not valid")
}

// TestParseVerificationMethod tests the parsing of verification methods
func TestParseVerificationMethod(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	// Test secp256k1 key (33 bytes compressed)
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	hexEncoded := hex.EncodeToString(pubKey.Bytes())

	parsedKey, err := k.parseVerificationMethod(hexEncoded)
	require.NoError(t, err)
	require.NotNil(t, parsedKey)
	require.Equal(t, pubKey.Bytes(), parsedKey.Bytes())

	// Test ed25519 key (32 bytes)
	ed25519PrivKey := ed25519.GenPrivKey()
	ed25519PubKey := ed25519PrivKey.PubKey()
	ed25519Hex := hex.EncodeToString(ed25519PubKey.Bytes())

	parsedEd25519, err := k.parseVerificationMethod(ed25519Hex)
	require.NoError(t, err)
	require.NotNil(t, parsedEd25519)
	require.Equal(t, ed25519PubKey.Bytes(), parsedEd25519.Bytes())

	// Test invalid encoding
	_, err = k.parseVerificationMethod("not-valid-hex")
	require.Error(t, err)
	require.Contains(t, err.Error(), "hex or base64 encoded")

	// Test empty verification method
	_, err = k.parseVerificationMethod("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be empty")

	// Test invalid key length
	invalidKey := hex.EncodeToString([]byte{1, 2, 3, 4, 5}) // 5 bytes - invalid
	_, err = k.parseVerificationMethod(invalidKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported public key length")
}

// TestVerifySignature tests the low-level signature verification
func TestVerifySignature(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	// Test secp256k1 signature verification
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	message := []byte("Test message for low-level verification")
	messageHash := sha256.Sum256(message)
	signature, err := privKey.Sign(messageHash[:])
	require.NoError(t, err)

	// Valid signature
	valid := k.verifySignature(pubKey, messageHash[:], signature)
	require.True(t, valid, "Valid signature should verify")

	// Invalid signature (wrong hash)
	wrongHash := sha256.Sum256([]byte("different message"))
	valid = k.verifySignature(pubKey, wrongHash[:], signature)
	require.False(t, valid, "Signature should not verify with wrong hash")

	// Test with nil public key
	valid = k.verifySignature(nil, messageHash[:], signature)
	require.False(t, valid, "Nil public key should return false")

	// Test ed25519 signature verification
	ed25519PrivKey := ed25519.GenPrivKey()
	ed25519PubKey := ed25519PrivKey.PubKey()

	ed25519Signature, err := ed25519PrivKey.Sign(messageHash[:])
	require.NoError(t, err)

	valid = k.verifySignature(ed25519PubKey, messageHash[:], ed25519Signature)
	require.True(t, valid, "Valid ed25519 signature should verify")
}

// TestVerifySignatureWithKey_KeyRotation tests signature verification during key rotation
func TestVerifySignatureWithKey_KeyRotation(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Generate old and new key pairs
	oldPrivKey := secp256k1.GenPrivKey()
	oldPubKey := oldPrivKey.PubKey()
	oldVerificationMethod := hex.EncodeToString(oldPubKey.Bytes())

	newPrivKey := secp256k1.GenPrivKey()
	newPubKey := newPrivKey.PubKey()
	newVerificationMethod := hex.EncodeToString(newPubKey.Bytes())

	// Create test DID and identity
	did := "did:aura:rotation-test"
	owner := sdk.AccAddress(oldPubKey.Address()).String()

	// Create identity with old verification method
	now := ctx.BlockTime()
	identity := &types.IdentityRecord{
		Did:                 did,
		Address:             owner,
		Status:              types.IdentityStatusActive,
		VerificationMethods: []string{oldVerificationMethod},
		CreatedAt:           now,
		UpdatedAt:           &now,
	}
	err := k.SetIdentityRecord(ctx, identity)
	require.NoError(t, err)

	// Initiate key rotation
	rotation, err := k.RotateDIDKey(ctx, did, owner, newVerificationMethod, "regular_rotation")
	require.NoError(t, err)
	require.NotNil(t, rotation)

	// Test message
	message := []byte("Message during key rotation")

	// Old key signature (should still work during grace period)
	oldMessageHash := sha256.Sum256(message)
	oldSignature, err := oldPrivKey.Sign(oldMessageHash[:])
	require.NoError(t, err)

	err = k.VerifySignatureWithKey(ctx, did, oldVerificationMethod, message, oldSignature)
	require.NoError(t, err, "Old key should still work during grace period")

	// New key signature (should work as primary key)
	newMessageHash := sha256.Sum256(message)
	newSignature, err := newPrivKey.Sign(newMessageHash[:])
	require.NoError(t, err)

	err = k.VerifySignatureWithKey(ctx, did, newVerificationMethod, message, newSignature)
	require.NoError(t, err, "New key should work as primary key")
}
