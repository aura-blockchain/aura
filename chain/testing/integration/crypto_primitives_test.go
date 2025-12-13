package integration_test

import (
	"crypto/sha256"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/testing/testutil"
)

// CryptoPrimitivesTestSuite verifies that cryptographic primitives are correctly
// linked and configured in the Aura blockchain. This includes:
// - Key generation (secp256k1, ed25519)
// - Message signing
// - Signature verification
// - Address derivation
// - Hash functions
//
// These tests ensure the fundamental cryptographic operations work correctly
// and are properly integrated with the Cosmos SDK.
type CryptoPrimitivesTestSuite struct {
	suite.Suite
	ctx *testutil.TestContext
}

func (s *CryptoPrimitivesTestSuite) SetupSuite() {
	s.ctx = testutil.SetupTestContext(s.T())
}

func TestCryptoPrimitivesTestSuite(t *testing.T) {
	suite.Run(t, new(CryptoPrimitivesTestSuite))
}

// TestSecp256k1KeyGeneration tests secp256k1 key generation and properties
func (s *CryptoPrimitivesTestSuite) TestSecp256k1KeyGeneration() {
	// Generate a new secp256k1 private key
	privKey := secp256k1.GenPrivKey()
	require.NotNil(s.T(), privKey, "private key should not be nil")
	require.Equal(s.T(), 32, len(privKey.Bytes()), "secp256k1 private key should be 32 bytes")

	// Get public key
	pubKey := privKey.PubKey()
	require.NotNil(s.T(), pubKey, "public key should not be nil")
	require.Equal(s.T(), 33, len(pubKey.Bytes()), "compressed secp256k1 public key should be 33 bytes")

	// Verify address derivation
	addr := sdk.AccAddress(pubKey.Address())
	require.Equal(s.T(), 20, len(addr), "account address should be 20 bytes")
}

// TestEd25519KeyGeneration tests ed25519 key generation and properties
func (s *CryptoPrimitivesTestSuite) TestEd25519KeyGeneration() {
	// Generate a new ed25519 private key
	privKey := ed25519.GenPrivKey()
	require.NotNil(s.T(), privKey, "private key should not be nil")
	require.Equal(s.T(), 64, len(privKey.Bytes()), "ed25519 private key should be 64 bytes")

	// Get public key
	pubKey := privKey.PubKey()
	require.NotNil(s.T(), pubKey, "public key should not be nil")
	require.Equal(s.T(), 32, len(pubKey.Bytes()), "ed25519 public key should be 32 bytes")

	// Verify address derivation
	addr := sdk.AccAddress(pubKey.Address())
	require.Equal(s.T(), 20, len(addr), "account address should be 20 bytes")
}

// TestSecp256k1SigningAndVerification tests the complete signing and verification flow
func (s *CryptoPrimitivesTestSuite) TestSecp256k1SigningAndVerification() {
	// Generate key pair
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	// Create test message
	message := []byte("This is a test message for Aura blockchain cryptographic verification")

	// Sign the message
	signature, err := privKey.Sign(message)
	require.NoError(s.T(), err, "signing should succeed")
	require.NotNil(s.T(), signature, "signature should not be nil")
	require.NotEmpty(s.T(), signature, "signature should not be empty")

	// Verify the signature with correct public key
	valid := pubKey.VerifySignature(message, signature)
	require.True(s.T(), valid, "signature verification should succeed with correct public key")

	// Verify signature fails with different message
	differentMessage := []byte("Different message")
	valid = pubKey.VerifySignature(differentMessage, signature)
	require.False(s.T(), valid, "signature verification should fail with different message")

	// Verify signature fails with different public key
	differentPrivKey := secp256k1.GenPrivKey()
	differentPubKey := differentPrivKey.PubKey()
	valid = differentPubKey.VerifySignature(message, signature)
	require.False(s.T(), valid, "signature verification should fail with different public key")
}

// TestEd25519SigningAndVerification tests ed25519 signing and verification
func (s *CryptoPrimitivesTestSuite) TestEd25519SigningAndVerification() {
	// Generate key pair
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	// Create test message
	message := []byte("Aura blockchain ed25519 signature test")

	// Sign the message
	signature, err := privKey.Sign(message)
	require.NoError(s.T(), err, "signing should succeed")
	require.NotNil(s.T(), signature, "signature should not be nil")
	require.Equal(s.T(), 64, len(signature), "ed25519 signature should be 64 bytes")

	// Verify the signature with correct public key
	valid := pubKey.VerifySignature(message, signature)
	require.True(s.T(), valid, "signature verification should succeed with correct public key")

	// Verify signature fails with tampered message
	tamperedMessage := []byte("Aura blockchain ed25519 signature test!")
	valid = pubKey.VerifySignature(tamperedMessage, signature)
	require.False(s.T(), valid, "signature verification should fail with tampered message")
}

// TestMultipleKeysAndSignatures tests that multiple keys can coexist
func (s *CryptoPrimitivesTestSuite) TestMultipleKeysAndSignatures() {
	message := []byte("Multi-signature test message")

	// Generate multiple key pairs
	privKey1 := secp256k1.GenPrivKey()
	privKey2 := secp256k1.GenPrivKey()
	privKey3 := ed25519.GenPrivKey()

	pubKey1 := privKey1.PubKey()
	pubKey2 := privKey2.PubKey()
	pubKey3 := privKey3.PubKey()

	// Create signatures from each key
	sig1, err := privKey1.Sign(message)
	require.NoError(s.T(), err)

	sig2, err := privKey2.Sign(message)
	require.NoError(s.T(), err)

	sig3, err := privKey3.Sign(message)
	require.NoError(s.T(), err)

	// Verify each signature with its corresponding public key
	require.True(s.T(), pubKey1.VerifySignature(message, sig1))
	require.True(s.T(), pubKey2.VerifySignature(message, sig2))
	require.True(s.T(), pubKey3.VerifySignature(message, sig3))

	// Cross-verify: signatures should not verify with wrong public keys
	require.False(s.T(), pubKey1.VerifySignature(message, sig2))
	require.False(s.T(), pubKey2.VerifySignature(message, sig1))
	require.False(s.T(), pubKey1.VerifySignature(message, sig3))
}

// TestAddressDerivation tests that addresses are correctly derived from public keys
func (s *CryptoPrimitivesTestSuite) TestAddressDerivation() {
	// Generate key pair
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	// Derive address
	addr1 := sdk.AccAddress(pubKey.Address())
	require.NotEmpty(s.T(), addr1)

	// Derive address again - should be deterministic
	addr2 := sdk.AccAddress(pubKey.Address())
	require.Equal(s.T(), addr1, addr2, "address derivation should be deterministic")

	// Different public key should produce different address
	differentPrivKey := secp256k1.GenPrivKey()
	differentPubKey := differentPrivKey.PubKey()
	differentAddr := sdk.AccAddress(differentPubKey.Address())
	require.NotEqual(s.T(), addr1, differentAddr, "different keys should produce different addresses")
}

// TestHashingFunctions tests that hashing functions work correctly
func (s *CryptoPrimitivesTestSuite) TestHashingFunctions() {
	data := []byte("Test data for hashing")

	// SHA256 hashing
	hash1 := sha256.Sum256(data)
	require.Equal(s.T(), 32, len(hash1), "SHA256 hash should be 32 bytes")

	// Hashing should be deterministic
	hash2 := sha256.Sum256(data)
	require.Equal(s.T(), hash1, hash2, "hashing should be deterministic")

	// Different data should produce different hash
	differentData := []byte("Different data for hashing")
	hash3 := sha256.Sum256(differentData)
	require.NotEqual(s.T(), hash1, hash3, "different data should produce different hash")
}

// TestKeyEquality tests that keys can be compared for equality
func (s *CryptoPrimitivesTestSuite) TestKeyEquality() {
	// Generate a key
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	// Same key should equal itself
	require.True(s.T(), pubKey.Equals(pubKey))

	// Different key should not equal
	differentPrivKey := secp256k1.GenPrivKey()
	differentPubKey := differentPrivKey.PubKey()
	require.False(s.T(), pubKey.Equals(differentPubKey))

	// Keys of different types should not equal
	ed25519Key := ed25519.GenPrivKey().PubKey()
	require.False(s.T(), pubKey.Equals(ed25519Key))
}

// TestPublicKeyFromBytes tests public key deserialization
func (s *CryptoPrimitivesTestSuite) TestPublicKeyFromBytes() {
	// Generate original key
	privKey := secp256k1.GenPrivKey()
	originalPubKey := privKey.PubKey()
	pubKeyBytes := originalPubKey.Bytes()

	// Reconstruct public key from bytes
	var reconstructedPubKey cryptotypes.PubKey
	reconstructedPubKey = &secp256k1.PubKey{Key: pubKeyBytes}

	// Verify addresses match
	originalAddr := sdk.AccAddress(originalPubKey.Address())
	reconstructedAddr := sdk.AccAddress(reconstructedPubKey.Address())
	require.Equal(s.T(), originalAddr, reconstructedAddr, "addresses should match after reconstruction")

	// Verify signature verification works with reconstructed key
	message := []byte("Test message")
	signature, err := privKey.Sign(message)
	require.NoError(s.T(), err)

	valid := reconstructedPubKey.VerifySignature(message, signature)
	require.True(s.T(), valid, "signature should verify with reconstructed public key")
}

// TestSignatureNonMalleability tests that signatures cannot be easily modified
func (s *CryptoPrimitivesTestSuite) TestSignatureNonMalleability() {
	// Generate key pair
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	message := []byte("Important transaction data")

	// Sign message
	signature, err := privKey.Sign(message)
	require.NoError(s.T(), err)

	// Verify original signature
	require.True(s.T(), pubKey.VerifySignature(message, signature))

	// Attempt to modify signature (flip a bit)
	if len(signature) > 0 {
		tamperedSignature := make([]byte, len(signature))
		copy(tamperedSignature, signature)
		tamperedSignature[0] ^= 0x01 // Flip least significant bit

		// Tampered signature should not verify
		require.False(s.T(), pubKey.VerifySignature(message, tamperedSignature))
	}
}

// TestEmptyAndNilInputs tests handling of edge cases
func (s *CryptoPrimitivesTestSuite) TestEmptyAndNilInputs() {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	// Empty message signing
	emptyMessage := []byte{}
	signature, err := privKey.Sign(emptyMessage)
	require.NoError(s.T(), err, "should be able to sign empty message")
	require.True(s.T(), pubKey.VerifySignature(emptyMessage, signature), "should verify empty message signature")

	// Nil signature verification should fail gracefully
	require.False(s.T(), pubKey.VerifySignature([]byte("test"), nil), "nil signature should not verify")

	// Empty signature verification should fail gracefully
	require.False(s.T(), pubKey.VerifySignature([]byte("test"), []byte{}), "empty signature should not verify")
}
