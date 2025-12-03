package keeper_test

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ripemd160"
)

// setupKeeperForSignatureTests creates a keeper for signature verification tests
func setupKeeperForSignatureTests(t *testing.T) (*keeper.Keeper, keepertest.TestInput) {
	input := keepertest.CreateTestInput(t)

	// Create LegacyAmino codec for paramstore
	legacyAmino := codec.NewLegacyAmino()

	// Create a paramstore with KeyTable
	ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, "bridge")

	k := keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		&ps,
		nil, // bankKeeper
		nil, // accountKeeper
		nil, // vcKeeper
		nil, // stakingKeeper
	)

	return k, input
}

// TestPAWSignatureVerification_ValidSignature tests that a valid signature is correctly verified
func TestPAWSignatureVerification_ValidSignature(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Generate a test key pair
	privKey, pubKey := generateTestKeyPair(t)

	// Derive the PAW address from the public key
	pawAddress := derivePawAddress(t, pubKey)
	auraAddress := "aura1test"

	// Create the message and sign it
	message := "Link PAW address " + pawAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	// Verify the signature
	valid := k.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress, signature)
	require.True(t, valid, "Valid PAW signature should be verified successfully")
}

// TestPAWSignatureVerification_InvalidSignatureLength tests rejection of incorrect length signatures
func TestPAWSignatureVerification_InvalidSignatureLength(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	pawAddress := "d5e2f3a1b4c6d8e9f0a1b2c3d4e5f6a7b8c9d0e1"
	auraAddress := "aura1test"

	// Test various invalid signature lengths
	testCases := []struct {
		name      string
		sigLength int
	}{
		{"empty signature", 0},
		{"too short (64 bytes)", 64},
		{"too short (32 bytes)", 32},
		{"too long (66 bytes)", 66},
		{"way too short", 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signature := make([]byte, tc.sigLength)
			valid := k.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress, signature)
			require.False(t, valid, "Signature with length %d should be rejected", tc.sigLength)
		})
	}
}

// TestPAWSignatureVerification_WrongSigner tests that signature from different signer is rejected
func TestPAWSignatureVerification_WrongSigner(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Generate two different key pairs
	privKey1, pubKey1 := generateTestKeyPair(t)
	_, pubKey2 := generateTestKeyPair(t)

	// Derive addresses from both public keys
	pawAddress1 := derivePawAddress(t, pubKey1)
	pawAddress2 := derivePawAddress(t, pubKey2)
	auraAddress := "aura1test"

	// Sign message with privKey1 for pawAddress1
	message := "Link PAW address " + pawAddress1 + " to Aura address " + auraAddress
	signature := signMessage(t, privKey1, message)

	// Try to verify with pawAddress2 (should fail)
	valid := k.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress2, signature)
	require.False(t, valid, "Signature from different signer should be rejected")
}

// TestPAWSignatureVerification_InvalidRecoveryID tests rejection of invalid recovery IDs
func TestPAWSignatureVerification_InvalidRecoveryID(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	privKey, pubKey := generateTestKeyPair(t)
	pawAddress := derivePawAddress(t, pubKey)
	auraAddress := "aura1test"

	message := "Link PAW address " + pawAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	// Modify recovery ID to various values
	// Recovery IDs 0-7 are valid (0-3 uncompressed, 4-7 compressed)
	// Recovery IDs 27-34 are also valid (27+0 through 27+7, with +27 offset)
	// Only values >34 or <27 and >7 are invalid
	testCases := []struct {
		name       string
		recoveryID byte
		expectPass bool
	}{
		{"recovery ID 4", 4, true},    // Valid compressed
		{"recovery ID 5", 5, true},    // Valid compressed
		{"recovery ID 30", 30, true},  // Valid (30-27=3, uncompressed with offset)
		{"recovery ID 35", 35, false}, // Invalid (35-27=8, >7)
		{"recovery ID 255", 255, false}, // Invalid (255-27=228, >7)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			modifiedSig := make([]byte, 65)
			copy(modifiedSig, signature)
			modifiedSig[64] = tc.recoveryID

			valid := k.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress, modifiedSig)
			if tc.expectPass {
				require.True(t, valid, "Signature with recovery ID %d should pass (valid ID)", tc.recoveryID)
			} else {
				require.False(t, valid, "Signature with invalid recovery ID %d should be rejected", tc.recoveryID)
			}
		})
	}
}

// TestPAWSignatureVerification_MalformedSignature tests rejection of malformed signatures
func TestPAWSignatureVerification_MalformedSignature(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	pawAddress := "d5e2f3a1b4c6d8e9f0a1b2c3d4e5f6a7b8c9d0e1"
	auraAddress := "aura1test"

	testCases := []struct {
		name      string
		signature []byte
	}{
		{"all zeros", make([]byte, 65)},
		{"all ones", func() []byte {
			b := make([]byte, 65)
			for i := range b {
				b[i] = 0xFF
			}
			return b
		}()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid := k.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress, tc.signature)
			require.False(t, valid, "Malformed signature (%s) should be rejected", tc.name)
		})
	}
}

// TestXAISignatureVerification_ValidSignature tests that a valid XAI signature is correctly verified
func TestXAISignatureVerification_ValidSignature(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Generate a test key pair
	privKey, pubKey := generateTestKeyPair(t)

	// Derive the XAI address from the public key
	xaiAddress := deriveXaiAddress(t, pubKey)
	auraAddress := "aura1test"

	// Create the message and sign it
	message := "Link XAI address " + xaiAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	// Verify the signature
	valid := k.VerifyXaiAddressOwnership(ctx, auraAddress, xaiAddress, signature)
	require.True(t, valid, "Valid XAI signature should be verified successfully")
}

// TestXAISignatureVerification_WrongSigner tests that XAI signature from different signer is rejected
func TestXAISignatureVerification_WrongSigner(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Generate two different key pairs
	privKey1, pubKey1 := generateTestKeyPair(t)
	_, pubKey2 := generateTestKeyPair(t)

	// Derive addresses from both public keys
	xaiAddress1 := deriveXaiAddress(t, pubKey1)
	xaiAddress2 := deriveXaiAddress(t, pubKey2)
	auraAddress := "aura1test"

	// Sign message with privKey1 for xaiAddress1
	message := "Link XAI address " + xaiAddress1 + " to Aura address " + auraAddress
	signature := signMessage(t, privKey1, message)

	// Try to verify with xaiAddress2 (should fail)
	valid := k.VerifyXaiAddressOwnership(ctx, auraAddress, xaiAddress2, signature)
	require.False(t, valid, "XAI signature from different signer should be rejected")
}

// TestSignatureReplayAttack tests that same signature cannot be used for different addresses
func TestSignatureReplayAttack(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Generate key pair
	privKey, pubKey := generateTestKeyPair(t)
	pawAddress := derivePawAddress(t, pubKey)
	auraAddress1 := "aura1test1"
	auraAddress2 := "aura1test2"

	// Sign for auraAddress1
	message1 := "Link PAW address " + pawAddress + " to Aura address " + auraAddress1
	signature1 := signMessage(t, privKey, message1)

	// Verify with auraAddress1 (should succeed)
	valid1 := k.VerifyPawAddressOwnership(ctx, auraAddress1, pawAddress, signature1)
	require.True(t, valid1, "Valid signature for auraAddress1 should verify")

	// Try to replay signature for auraAddress2 (should fail - different message)
	valid2 := k.VerifyPawAddressOwnership(ctx, auraAddress2, pawAddress, signature1)
	require.False(t, valid2, "Signature replay for different Aura address should fail")
}

// TestEmptyAddresses tests rejection when addresses are empty
func TestEmptyAddresses(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	privKey, pubKey := generateTestKeyPair(t)
	pawAddress := derivePawAddress(t, pubKey)
	auraAddress := "aura1test"
	message := "Link PAW address " + pawAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	testCases := []struct {
		name        string
		auraAddr    string
		pawAddr     string
		expectValid bool
	}{
		{"both addresses valid", auraAddress, pawAddress, true},
		{"empty aura address", "", pawAddress, false},
		{"empty paw address", auraAddress, "", false},
		{"both addresses empty", "", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid := k.VerifyPawAddressOwnership(ctx, tc.auraAddr, tc.pawAddr, signature)
			require.Equal(t, tc.expectValid, valid, tc.name)
		})
	}
}

// ========================================================================
// HELPER FUNCTIONS FOR TESTING
// ========================================================================

// generateTestKeyPair generates a secp256k1 key pair for testing
func generateTestKeyPair(t *testing.T) (*secp256k1.PrivateKey, *secp256k1.PublicKey) {
	// Generate private key
	privKey, err := secp256k1.GeneratePrivateKey()
	require.NoError(t, err, "Failed to generate private key")

	// Get public key
	pubKey := privKey.PubKey()
	return privKey, pubKey
}

// signMessage signs a message with a private key and returns a 65-byte signature
func signMessage(t *testing.T, privKey *secp256k1.PrivateKey, message string) []byte {
	// Hash the message
	msgHash := sha256.Sum256([]byte(message))

	// Sign the message using SignCompact
	// SignCompact returns: [recovery_id + 27][R (32 bytes)][S (32 bytes)]
	compactSig := secp256k1ecdsa.SignCompact(privKey, msgHash[:], true)
	require.Equal(t, 65, len(compactSig), "SignCompact should return 65 bytes")

	// Rearrange to format expected by verification: [R][S][recovery_id]
	result := make([]byte, 65)
	copy(result[0:64], compactSig[1:65]) // Copy R and S
	result[64] = compactSig[0] - 27      // Convert recovery byte (27-30 -> 0-3)

	return result
}

// derivePawAddress derives a PAW address from a public key
func derivePawAddress(t *testing.T, pubKey *secp256k1.PublicKey) string {
	// Get compressed public key (33 bytes)
	pubKeyBytes := pubKey.SerializeCompressed()
	require.Equal(t, 33, len(pubKeyBytes), "Public key should be 33 bytes")

	// SHA256 hash
	sha256Hash := sha256.Sum256(pubKeyBytes)

	// RIPEMD160 hash
	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash[:])
	addressHash := ripemd160Hasher.Sum(nil)

	// Return hex-encoded address
	return hex.EncodeToString(addressHash)
}

// deriveXaiAddress derives an XAI address from a public key
func deriveXaiAddress(t *testing.T, pubKey *secp256k1.PublicKey) string {
	// Same derivation process as PAW
	return derivePawAddress(t, pubKey)
}

// goEcdsaPrivKeyToSecp256k1 converts a Go ecdsa.PrivateKey to secp256k1.PrivateKey
func goEcdsaPrivKeyToSecp256k1(privKey *ecdsa.PrivateKey) *secp256k1.PrivateKey {
	privKeyBytes := privKey.D.Bytes()
	// Pad to 32 bytes if necessary
	if len(privKeyBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(privKeyBytes):], privKeyBytes)
		privKeyBytes = padded
	}
	return secp256k1.PrivKeyFromBytes(privKeyBytes)
}
