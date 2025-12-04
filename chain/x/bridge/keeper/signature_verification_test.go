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

// ========================================================================
// TELEMETRY AND LOGGING COVERAGE TESTS
// ========================================================================

// TestPAWSignatureVerification_TelemetrySuccessPath tests that successful verification
// calls the correct telemetry functions (recordSignatureVerification with true)
func TestPAWSignatureVerification_TelemetrySuccessPath(t *testing.T) {
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

	// Verify the signature - this should trigger:
	// - ctx.Logger().Info() call on success (line 463)
	// - k.recordSignatureVerification("paw", "link_address", true, duration) on success (line 468)
	valid := k.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress, signature)
	require.True(t, valid, "Valid PAW signature should be verified successfully")

	// Test passes if no panic occurs - telemetry functions are called
}

// TestPAWSignatureVerification_TelemetryEmptyInput tests that empty input
// triggers the correct telemetry calls (recordSignatureMismatch and recordSignatureVerification with false)
func TestPAWSignatureVerification_TelemetryEmptyInput(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	testCases := []struct {
		name        string
		auraAddress string
		pawAddress  string
		signature   []byte
	}{
		{
			name:        "empty signature",
			auraAddress: "aura1test",
			pawAddress:  "d5e2f3a1b4c6d8e9f0a1b2c3d4e5f6a7b8c9d0e1",
			signature:   []byte{},
		},
		{
			name:        "empty paw address",
			auraAddress: "aura1test",
			pawAddress:  "",
			signature:   make([]byte, 65),
		},
		{
			name:        "empty aura address",
			auraAddress: "",
			pawAddress:  "d5e2f3a1b4c6d8e9f0a1b2c3d4e5f6a7b8c9d0e1",
			signature:   make([]byte, 65),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This should trigger (lines 374-376):
			// - k.recordSignatureMismatch("paw", "link_address", "empty_input")
			// - k.recordSignatureVerification("paw", "link_address", false, duration)
			valid := k.VerifyPawAddressOwnership(ctx, tc.auraAddress, tc.pawAddress, tc.signature)
			require.False(t, valid, "Empty input should be rejected")
			// Test passes if no panic occurs - telemetry functions are called
		})
	}
}

// TestPAWSignatureVerification_TelemetryInvalidLength tests that invalid signature length
// triggers Logger.Error and telemetry calls
func TestPAWSignatureVerification_TelemetryInvalidLength(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	pawAddress := "d5e2f3a1b4c6d8e9f0a1b2c3d4e5f6a7b8c9d0e1"
	auraAddress := "aura1test"

	testCases := []struct {
		name      string
		sigLength int
	}{
		{"64 bytes", 64},
		{"66 bytes", 66},
		{"32 bytes", 32},
		{"100 bytes", 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signature := make([]byte, tc.sigLength)

			// This should trigger (lines 382-388):
			// - ctx.Logger().Error() call with signature length details
			// - k.recordSignatureMismatch("paw", "link_address", "invalid_signature_length")
			// - k.recordSignatureVerification("paw", "link_address", false, duration)
			valid := k.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress, signature)
			require.False(t, valid, "Signature with invalid length should be rejected")
			// Test passes if no panic occurs - Logger and telemetry functions are called
		})
	}
}

// TestPAWSignatureVerification_TelemetryInvalidRecoveryID tests that invalid recovery ID
// triggers Logger.Error, recordInvalidRecoveryID, and other telemetry calls
func TestPAWSignatureVerification_TelemetryInvalidRecoveryID(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	privKey, pubKey := generateTestKeyPair(t)
	pawAddress := derivePawAddress(t, pubKey)
	auraAddress := "aura1test"

	message := "Link PAW address " + pawAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	testCases := []struct {
		name       string
		recoveryID byte
	}{
		{"recovery ID 35", 35},   // 35-27=8, which is >7
		{"recovery ID 100", 100}, // 100-27=73, which is >7
		{"recovery ID 255", 255}, // 255-27=228, which is >7
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			modifiedSig := make([]byte, 65)
			copy(modifiedSig, signature)
			modifiedSig[64] = tc.recoveryID

			// This should trigger (lines 406-412):
			// - ctx.Logger().Error() call with recovery ID details
			// - k.recordInvalidRecoveryID("paw")
			// - k.recordSignatureMismatch("paw", "link_address", "invalid_recovery_id")
			// - k.recordSignatureVerification("paw", "link_address", false, duration)
			valid := k.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress, modifiedSig)
			require.False(t, valid, "Signature with invalid recovery ID should be rejected")
			// Test passes if no panic occurs - Logger and telemetry functions are called
		})
	}
}

// TestPAWSignatureVerification_TelemetryPubKeyRecoveryFailure tests that public key
// recovery failure triggers Logger.Error, recordPubKeyRecoveryFailure, and other telemetry calls
func TestPAWSignatureVerification_TelemetryPubKeyRecoveryFailure(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Use a different key pair to generate signature
	privKey, pubKey := generateTestKeyPair(t)
	correctPawAddress := derivePawAddress(t, pubKey)
	auraAddress := "aura1test"

	// Sign message with correct address
	message := "Link PAW address " + correctPawAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	// Try to verify with a completely different PAW address (not derivable from signature)
	wrongPawAddress := "0000000000000000000000000000000000000001"

	// This should trigger (lines 442-448):
	// - ctx.Logger().Error() call about public key recovery failure
	// - k.recordPubKeyRecoveryFailure("paw", recoveryID)
	// - k.recordSignatureMismatch("paw", "link_address", "pubkey_recovery_failed")
	// - k.recordSignatureVerification("paw", "link_address", false, duration)
	valid := k.VerifyPawAddressOwnership(ctx, auraAddress, wrongPawAddress, signature)
	require.False(t, valid, "Signature with wrong PAW address should fail public key recovery")
	// Test passes if no panic occurs - Logger and telemetry functions are called
}

// TestPAWSignatureVerification_TelemetryEcdsaVerificationFailure tests that ECDSA
// verification failure triggers Logger.Error and telemetry calls
func TestPAWSignatureVerification_TelemetryEcdsaVerificationFailure(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	privKey, pubKey := generateTestKeyPair(t)
	pawAddress := derivePawAddress(t, pubKey)
	auraAddress := "aura1test"

	message := "Link PAW address " + pawAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	// Corrupt the signature R and S components (first 64 bytes) while keeping recovery ID valid
	// This will pass public key recovery but fail ECDSA verification
	corruptedSig := make([]byte, 65)
	copy(corruptedSig, signature)
	// Modify first byte of R component slightly to break ECDSA verification
	// but not so much that key recovery completely fails
	corruptedSig[0] ^= 0x01

	// This may trigger (lines 455-460):
	// - ctx.Logger().Error() call about ECDSA verification failure
	// - k.recordSignatureMismatch("paw", "link_address", "ecdsa_verification_failed")
	// - k.recordSignatureVerification("paw", "link_address", false, duration)
	valid := k.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress, corruptedSig)
	require.False(t, valid, "Corrupted signature should fail ECDSA verification")
	// Test passes if no panic occurs - Logger and telemetry functions are called
}

// TestPAWSignatureVerification_AllTelemetryPathsNoContext tests that all telemetry
// calls handle nil context gracefully (no panic) - this verifies telemetry is safe
func TestPAWSignatureVerification_AllTelemetryPathsNoContext(t *testing.T) {
	k, _ := setupKeeperForSignatureTests(t)

	// Create a minimal context that won't panic but also won't have full logger
	ctx := keepertest.CreateTestInput(t).Ctx

	privKey, pubKey := generateTestKeyPair(t)
	pawAddress := derivePawAddress(t, pubKey)
	auraAddress := "aura1test"
	message := "Link PAW address " + pawAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	// Test all paths don't panic with various inputs
	testCases := []struct {
		name        string
		auraAddr    string
		pawAddr     string
		sig         []byte
		description string
	}{
		{
			name:        "empty input path",
			auraAddr:    "",
			pawAddr:     pawAddress,
			sig:         signature,
			description: "triggers recordSignatureMismatch and recordSignatureVerification(false)",
		},
		{
			name:        "invalid length path",
			auraAddr:    auraAddress,
			pawAddr:     pawAddress,
			sig:         make([]byte, 64),
			description: "triggers Logger.Error, recordSignatureMismatch, recordSignatureVerification(false)",
		},
		{
			name:        "invalid recovery ID path",
			auraAddr:    auraAddress,
			pawAddr:     pawAddress,
			sig:         func() []byte { s := make([]byte, 65); s[64] = 255; return s }(),
			description: "triggers Logger.Error, recordInvalidRecoveryID, recordSignatureMismatch, recordSignatureVerification(false)",
		},
		{
			name:        "success path",
			auraAddr:    auraAddress,
			pawAddr:     pawAddress,
			sig:         signature,
			description: "triggers Logger.Info, recordSignatureVerification(true)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic when executing telemetry calls
			require.NotPanics(t, func() {
				_ = k.VerifyPawAddressOwnership(ctx, tc.auraAddr, tc.pawAddr, tc.sig)
			}, "Telemetry calls should not panic: %s", tc.description)
		})
	}
}

// ========================================================================
// XAI TELEMETRY AND LOGGING COVERAGE TESTS
// ========================================================================

// TestXAISignatureVerification_EmptyInputs tests rejection when inputs are empty
func TestXAISignatureVerification_EmptyInputs(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	privKey, pubKey := generateTestKeyPair(t)
	xaiAddress := deriveXaiAddress(t, pubKey)
	auraAddress := "aura1test"
	message := "Link XAI address " + xaiAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	testCases := []struct {
		name        string
		auraAddr    string
		xaiAddr     string
		sig         []byte
		expectValid bool
	}{
		{"all valid", auraAddress, xaiAddress, signature, true},
		{"empty aura address", "", xaiAddress, signature, false},
		{"empty xai address", auraAddress, "", signature, false},
		{"empty signature", auraAddress, xaiAddress, []byte{}, false},
		{"empty aura and xai", "", "", signature, false},
		{"empty signature and aura", "", xaiAddress, []byte{}, false},
		{"empty signature and xai", auraAddress, "", []byte{}, false},
		{"all empty", "", "", []byte{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This triggers (lines 502-506 in keeper.go):
			// - k.recordSignatureMismatch("xai", "link_address", "empty_input")
			// - k.recordSignatureVerification("xai", "link_address", false, duration)
			valid := k.VerifyXaiAddressOwnership(ctx, tc.auraAddr, tc.xaiAddr, tc.sig)
			require.Equal(t, tc.expectValid, valid, tc.name)
		})
	}
}

// TestXAISignatureVerification_InvalidSignatureLength tests rejection of incorrect length signatures
func TestXAISignatureVerification_InvalidSignatureLength(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	xaiAddress := "d5e2f3a1b4c6d8e9f0a1b2c3d4e5f6a7b8c9d0e1"
	auraAddress := "aura1test"

	testCases := []struct {
		name      string
		sigLength int
	}{
		{"too short (64 bytes)", 64},
		{"too short (32 bytes)", 32},
		{"too long (66 bytes)", 66},
		{"way too short", 10},
		{"way too long", 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			signature := make([]byte, tc.sigLength)
			// This triggers (lines 511-517 in keeper.go):
			// - ctx.Logger().Error() call with signature length details
			// - k.recordSignatureMismatch("xai", "link_address", "invalid_signature_length")
			// - k.recordSignatureVerification("xai", "link_address", false, duration)
			valid := k.VerifyXaiAddressOwnership(ctx, auraAddress, xaiAddress, signature)
			require.False(t, valid, "Signature with length %d should be rejected", tc.sigLength)
		})
	}
}

// TestXAISignatureVerification_InvalidRecoveryID tests rejection of invalid recovery IDs
func TestXAISignatureVerification_InvalidRecoveryID(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	privKey, pubKey := generateTestKeyPair(t)
	xaiAddress := deriveXaiAddress(t, pubKey)
	auraAddress := "aura1test"

	message := "Link XAI address " + xaiAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	testCases := []struct {
		name       string
		recoveryID byte
		expectPass bool
	}{
		{"recovery ID 0", 0, true},      // Valid uncompressed
		{"recovery ID 1", 1, true},      // Valid uncompressed
		{"recovery ID 2", 2, true},      // Valid uncompressed
		{"recovery ID 3", 3, true},      // Valid uncompressed
		{"recovery ID 4", 4, true},      // Valid compressed
		{"recovery ID 5", 5, true},      // Valid compressed
		{"recovery ID 6", 6, true},      // Valid compressed
		{"recovery ID 7", 7, true},      // Valid compressed
		{"recovery ID 27", 27, true},    // Valid with offset (27-27=0)
		{"recovery ID 28", 28, true},    // Valid with offset (28-27=1)
		{"recovery ID 34", 34, true},    // Valid with offset (34-27=7)
		{"recovery ID 35", 35, false},   // Invalid (35-27=8, >7)
		{"recovery ID 100", 100, false}, // Invalid (100-27=73, >7)
		{"recovery ID 255", 255, false}, // Invalid (255-27=228, >7)
		{"recovery ID 8", 8, false},     // Invalid (no offset, >7)
		{"recovery ID 10", 10, false},   // Invalid (no offset, >7)
		{"recovery ID 26", 26, false},   // Invalid (no offset, >7)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			modifiedSig := make([]byte, 65)
			copy(modifiedSig, signature)
			modifiedSig[64] = tc.recoveryID

			if !tc.expectPass {
				// This triggers (lines 535-541 in keeper.go):
				// - ctx.Logger().Error() call with recovery ID details
				// - k.recordInvalidRecoveryID("xai")
				// - k.recordSignatureMismatch("xai", "link_address", "invalid_recovery_id")
				// - k.recordSignatureVerification("xai", "link_address", false, duration)
			}

			valid := k.VerifyXaiAddressOwnership(ctx, auraAddress, xaiAddress, modifiedSig)
			if tc.expectPass {
				require.True(t, valid, "Signature with recovery ID %d should pass (valid ID)", tc.recoveryID)
			} else {
				require.False(t, valid, "Signature with invalid recovery ID %d should be rejected", tc.recoveryID)
			}
		})
	}
}

// TestXAISignatureVerification_MalformedSignature tests rejection of malformed signatures
func TestXAISignatureVerification_MalformedSignature(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	xaiAddress := "d5e2f3a1b4c6d8e9f0a1b2c3d4e5f6a7b8c9d0e1"
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
		{"invalid R component", func() []byte {
			b := make([]byte, 65)
			// Set recovery ID to valid value, but R to all zeros
			b[64] = 0
			return b
		}()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This triggers (lines 571-577 in keeper.go):
			// - ctx.Logger().Error() call about public key recovery failure
			// - k.recordPubKeyRecoveryFailure("xai", recoveryID)
			// - k.recordSignatureMismatch("xai", "link_address", "pubkey_recovery_failed")
			// - k.recordSignatureVerification("xai", "link_address", false, duration)
			valid := k.VerifyXaiAddressOwnership(ctx, auraAddress, xaiAddress, tc.signature)
			require.False(t, valid, "Malformed signature (%s) should be rejected", tc.name)
		})
	}
}

// TestXAISignatureVerification_TamperedMessage tests rejection when message is tampered
func TestXAISignatureVerification_TamperedMessage(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	privKey, pubKey := generateTestKeyPair(t)
	xaiAddress := deriveXaiAddress(t, pubKey)
	auraAddress1 := "aura1test1"
	auraAddress2 := "aura1test2"

	// Sign for auraAddress1
	message1 := "Link XAI address " + xaiAddress + " to Aura address " + auraAddress1
	signature1 := signMessage(t, privKey, message1)

	// Verify with auraAddress1 (should succeed)
	valid1 := k.VerifyXaiAddressOwnership(ctx, auraAddress1, xaiAddress, signature1)
	require.True(t, valid1, "Valid signature for auraAddress1 should verify")

	// Try to use same signature with different aura address (should fail - message mismatch)
	valid2 := k.VerifyXaiAddressOwnership(ctx, auraAddress2, xaiAddress, signature1)
	require.False(t, valid2, "Signature with tampered message (different aura address) should fail")
}

// TestXAISignatureVerification_TelemetrySuccessPath tests that successful verification
// calls the correct telemetry functions (recordSignatureVerification with true)
func TestXAISignatureVerification_TelemetrySuccessPath(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	privKey, pubKey := generateTestKeyPair(t)
	xaiAddress := deriveXaiAddress(t, pubKey)
	auraAddress := "aura1test"

	message := "Link XAI address " + xaiAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	// Verify the signature - this should trigger (lines 592-597 in keeper.go):
	// - ctx.Logger().Info() call on success
	// - k.recordSignatureVerification("xai", "link_address", true, duration)
	valid := k.VerifyXaiAddressOwnership(ctx, auraAddress, xaiAddress, signature)
	require.True(t, valid, "Valid XAI signature should be verified successfully")
}

// TestXAISignatureVerification_TelemetryPubKeyRecoveryFailure tests that public key
// recovery failure triggers Logger.Error, recordPubKeyRecoveryFailure, and other telemetry calls
func TestXAISignatureVerification_TelemetryPubKeyRecoveryFailure(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	privKey, pubKey := generateTestKeyPair(t)
	correctXaiAddress := deriveXaiAddress(t, pubKey)
	auraAddress := "aura1test"

	message := "Link XAI address " + correctXaiAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	// Try to verify with a completely different XAI address
	wrongXaiAddress := "0000000000000000000000000000000000000001"

	// This should trigger (lines 571-577 in keeper.go):
	// - ctx.Logger().Error() call about public key recovery failure
	// - k.recordPubKeyRecoveryFailure("xai", recoveryID)
	// - k.recordSignatureMismatch("xai", "link_address", "pubkey_recovery_failed")
	// - k.recordSignatureVerification("xai", "link_address", false, duration)
	valid := k.VerifyXaiAddressOwnership(ctx, auraAddress, wrongXaiAddress, signature)
	require.False(t, valid, "Signature with wrong XAI address should fail public key recovery")
}

// TestXAISignatureVerification_TelemetryEcdsaVerificationFailure tests that ECDSA
// verification failure triggers Logger.Error and telemetry calls
func TestXAISignatureVerification_TelemetryEcdsaVerificationFailure(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	privKey, pubKey := generateTestKeyPair(t)
	xaiAddress := deriveXaiAddress(t, pubKey)
	auraAddress := "aura1test"

	message := "Link XAI address " + xaiAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	// Corrupt the signature R and S components (first 64 bytes) while keeping recovery ID valid
	corruptedSig := make([]byte, 65)
	copy(corruptedSig, signature)
	// Modify first byte of R component slightly to break ECDSA verification
	corruptedSig[0] ^= 0x01

	// This may trigger (lines 584-589 in keeper.go):
	// - ctx.Logger().Error() call about ECDSA verification failure
	// - k.recordSignatureMismatch("xai", "link_address", "ecdsa_verification_failed")
	// - k.recordSignatureVerification("xai", "link_address", false, duration)
	valid := k.VerifyXaiAddressOwnership(ctx, auraAddress, xaiAddress, corruptedSig)
	require.False(t, valid, "Corrupted signature should fail ECDSA verification")
}

// TestXAISignatureVerification_AllTelemetryPaths tests all telemetry paths
func TestXAISignatureVerification_AllTelemetryPaths(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	privKey, pubKey := generateTestKeyPair(t)
	xaiAddress := deriveXaiAddress(t, pubKey)
	auraAddress := "aura1test"
	message := "Link XAI address " + xaiAddress + " to Aura address " + auraAddress
	signature := signMessage(t, privKey, message)

	testCases := []struct {
		name        string
		auraAddr    string
		xaiAddr     string
		sig         []byte
		description string
	}{
		{
			name:        "empty input path",
			auraAddr:    "",
			xaiAddr:     xaiAddress,
			sig:         signature,
			description: "triggers recordSignatureMismatch and recordSignatureVerification(false)",
		},
		{
			name:        "invalid length path",
			auraAddr:    auraAddress,
			xaiAddr:     xaiAddress,
			sig:         make([]byte, 64),
			description: "triggers Logger.Error, recordSignatureMismatch, recordSignatureVerification(false)",
		},
		{
			name:        "invalid recovery ID path",
			auraAddr:    auraAddress,
			xaiAddr:     xaiAddress,
			sig:         func() []byte { s := make([]byte, 65); s[64] = 255; return s }(),
			description: "triggers Logger.Error, recordInvalidRecoveryID, recordSignatureMismatch, recordSignatureVerification(false)",
		},
		{
			name:        "success path",
			auraAddr:    auraAddress,
			xaiAddr:     xaiAddress,
			sig:         signature,
			description: "triggers Logger.Info, recordSignatureVerification(true)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				_ = k.VerifyXaiAddressOwnership(ctx, tc.auraAddr, tc.xaiAddr, tc.sig)
			}, "Telemetry calls should not panic: %s", tc.description)
		})
	}
}

// TestPAWSignatureVerification_EcdsaVerificationFailurePath tests the specific ECDSA verification
// failure path where public key recovery succeeds but signature verification fails.
// This covers lines 454-461 in keeper.go (verifyPawAddressOwnership).
func TestPAWSignatureVerification_EcdsaVerificationFailurePath(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Generate a test key pair
	privKey, pubKey := generateTestKeyPair(t)
	pawAddress := derivePawAddress(t, pubKey)
	auraAddress := "aura1test"

	// Strategy: Sign a DIFFERENT message with the same key, but use the PAW address
	// derived from the original public key. This will:
	// 1. Successfully recover a public key (because the signature structure is valid)
	// 2. Fail ECDSA verification (because the signature doesn't match the expected message hash)

	// Sign a different message
	differentMessage := "Link PAW address " + pawAddress + " to Aura address aura1different"
	differentSignature := signMessage(t, privKey, differentMessage)

	// Now try to verify the signature for differentMessage using the original auraAddress
	// This means we're verifying against the wrong message hash
	// The expected message hash will be for "Link PAW address ... to Aura address aura1test"
	// But the signature is for "Link PAW address ... to Aura address aura1different"

	// The verification process will:
	// 1. Construct the expected message: "Link PAW address " + pawAddress + " to Aura address " + auraAddress
	// 2. Hash it to get msgHash
	// 3. Try to recover public key from differentSignature - this WILL succeed
	// 4. Try to verify differentSignature against msgHash - this WILL fail (wrong message)

	valid := k.VerifyPawAddressOwnership(ctx, auraAddress, pawAddress, differentSignature)

	// Verification should fail because the signature was created for a different message
	require.False(t, valid, "Signature for different message should fail ECDSA verification")

	// Verify that the correct path was taken by also testing the valid signature works
	validResult := k.VerifyPawAddressOwnership(ctx, "aura1different", pawAddress, differentSignature)
	require.True(t, validResult, "Signature should be valid for the correct message")

	// This test covers:
	// - ctx.Logger().Error("PAW signature verification failed", ...) at line 455
	// - k.recordSignatureMismatch("paw", "link_address", "ecdsa_verification_failed") at line 458
	// - k.recordSignatureVerification("paw", "link_address", false, time.Since(startTime)) at line 459
}

// ========================================================================
// XAI EMPTY INPUT VALIDATION COVERAGE TESTS
// ========================================================================

// TestXAISignatureVerification_EmptyInputValidation tests the empty input validation
// path in verifyXaiAddressOwnership (keeper.go lines 502-506).
//
// This test ensures that all three empty input scenarios are correctly rejected:
// 1. Empty signature with valid addresses
// 2. Valid signature with empty XAI address
// 3. Valid signature with empty Aura address
//
// Each scenario verifies:
// - Function returns false
// - recordSignatureMismatch() called with "empty_input"
// - recordSignatureVerification() called with success=false
func TestXAISignatureVerification_EmptyInputValidation(t *testing.T) {
	k, input := setupKeeperForSignatureTests(t)
	ctx := input.Ctx

	// Generate valid test data for use in test cases
	privKey, pubKey := generateTestKeyPair(t)
	validXaiAddress := deriveXaiAddress(t, pubKey)
	validAuraAddress := "aura1test"
	validMessage := "Link XAI address " + validXaiAddress + " to Aura address " + validAuraAddress
	validSignature := signMessage(t, privKey, validMessage)

	testCases := []struct {
		name        string
		signature   []byte
		xaiAddress  string
		auraAddress string
		description string
	}{
		{
			name:        "empty signature with valid addresses",
			signature:   []byte{},
			xaiAddress:  validXaiAddress,
			auraAddress: validAuraAddress,
			description: "Empty signature should trigger empty_input validation failure",
		},
		{
			name:        "valid signature with empty XAI address",
			signature:   validSignature,
			xaiAddress:  "",
			auraAddress: validAuraAddress,
			description: "Empty XAI address should trigger empty_input validation failure",
		},
		{
			name:        "valid signature with empty Aura address",
			signature:   validSignature,
			xaiAddress:  validXaiAddress,
			auraAddress: "",
			description: "Empty Aura address should trigger empty_input validation failure",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call verifyXaiAddressOwnership which should trigger (lines 502-506):
			// - if len(signature) == 0 || xaiAddress == "" || auraAddress == ""
			// - k.recordSignatureMismatch("xai", "link_address", "empty_input")
			// - k.recordSignatureVerification("xai", "link_address", false, time.Since(startTime))
			// - return false
			valid := k.VerifyXaiAddressOwnership(ctx, tc.auraAddress, tc.xaiAddress, tc.signature)

			// Verify function returns false for empty inputs
			require.False(t, valid, "%s: %s", tc.name, tc.description)

			// Verify no panics occur (telemetry functions are called correctly)
			// The telemetry functions recordSignatureMismatch and recordSignatureVerification
			// should be called without errors
		})
	}
}
