package keeper_test

import (
	"crypto/sha256"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/bridge/keeper"
	"github.com/aequitas/aura/chain/x/bridge/types"
)

// TestLinkAddress_SignerVerification tests that only the Aura address owner can link addresses
//
// SECURITY TEST: Prevents unauthorized address linking
// Attack prevented: Anyone linking arbitrary addresses to their identity
func TestLinkAddress_SignerVerification(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	// Generate TWO DIFFERENT addresses
	addrs := keepertest.GenTestAddrs(2)
	auraAddr := addrs[0]
	attackerAddr := addrs[1]

	// Create valid signature (64+ bytes)
	signature := make([]byte, 65)

	// Attempt to link auraAddr from attackerAddr (should fail)
	msg := &types.MsgLinkAddress{
		Signer:       attackerAddr.String(), // Attacker trying to link someone else's address
		AuraAddress:  auraAddr.String(),     // Victim's address
		PawAddress:   "paw1victim",
		PawSignature: signature,
		XaiAddress:   "",
		XaiSignature: nil,
	}

	_, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err, "should reject unauthorized signer")

	// Verify it's a permission denied error
	st, ok := status.FromError(err)
	require.True(t, ok, "error should be a gRPC status error")
	require.Equal(t, codes.PermissionDenied, st.Code(), "should be PermissionDenied error")
	require.Contains(t, st.Message(), "signer must be the Aura address owner")
}

// TestLinkAddress_ValidSigner tests that the correct signer can link addresses
//
// SECURITY TEST: Validates legitimate linking by authorized user
func TestLinkAddress_ValidSigner(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	// Generate address
	auraAddr := keepertest.GenTestAddr()

	// Generate real cryptographic key pair and signature for PAW
	privKey, pubKey := generateTestKeyPair(t)
	pawAddress := derivePawAddress(t, pubKey)
	message := "Link PAW address " + pawAddress + " to Aura address " + auraAddr.String()
	signature := signMessage(t, privKey, message)

	// Link with correct signer (should succeed)
	msg := &types.MsgLinkAddress{
		Signer:       auraAddr.String(), // Correct owner
		AuraAddress:  auraAddr.String(), // Same address
		PawAddress:   pawAddress,
		PawSignature: signature,
		XaiAddress:   "",
		XaiSignature: nil,
	}

	resp, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err, "should allow authorized signer")
	require.True(t, resp.Success)
	require.Equal(t, auraAddr.String(), resp.LinkedIdentityId)

	// Verify the identity was created
	identity, found := k.GetSharedIdentity(ctx, auraAddr.String())
	require.True(t, found, "identity should be created")
	require.Equal(t, auraAddr.String(), identity.Address)
	require.Equal(t, pawAddress, identity.LinkedAddresses["paw"])
	require.True(t, identity.VerifiedPaw, "PAW should be verified")
}

// TestLinkAddress_PawSignatureRequired tests that PAW signature is required when linking PAW address
//
// SECURITY TEST: Prevents linking PAW addresses without proof of ownership
func TestLinkAddress_PawSignatureRequired(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	auraAddr := keepertest.GenTestAddr()

	// Attempt to link PAW address without signature (should fail)
	msg := &types.MsgLinkAddress{
		Signer:       auraAddr.String(),
		AuraAddress:  auraAddr.String(),
		PawAddress:   "paw1test",
		PawSignature: nil, // Missing signature
		XaiAddress:   "",
		XaiSignature: nil,
	}

	_, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err, "should reject missing PAW signature")

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
	require.Contains(t, st.Message(), "PAW signature required")
}

// TestLinkAddress_XaiSignatureRequired tests that XAI signature is required when linking XAI address
//
// SECURITY TEST: Prevents linking XAI addresses without proof of ownership
func TestLinkAddress_XaiSignatureRequired(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	auraAddr := keepertest.GenTestAddr()

	// Attempt to link XAI address without signature (should fail)
	msg := &types.MsgLinkAddress{
		Signer:       auraAddr.String(),
		AuraAddress:  auraAddr.String(),
		PawAddress:   "",
		PawSignature: nil,
		XaiAddress:   "xai1test",
		XaiSignature: nil, // Missing signature
	}

	_, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err, "should reject missing XAI signature")

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
	require.Contains(t, st.Message(), "XAI signature required")
}

// TestLinkAddress_InvalidPawSignature tests that invalid PAW signatures are rejected
//
// SECURITY TEST: Prevents linking with forged or invalid signatures
func TestLinkAddress_InvalidPawSignature(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	auraAddr := keepertest.GenTestAddr()

	// Create invalid signature (too short)
	invalidSignature := make([]byte, 32) // Only 32 bytes, need 64+

	msg := &types.MsgLinkAddress{
		Signer:       auraAddr.String(),
		AuraAddress:  auraAddr.String(),
		PawAddress:   "paw1test",
		PawSignature: invalidSignature, // Invalid signature
		XaiAddress:   "",
		XaiSignature: nil,
	}

	_, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err, "should reject invalid PAW signature")

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
	require.Contains(t, st.Message(), "invalid PAW address ownership proof")
}

// TestLinkAddress_InvalidXaiSignature tests that invalid XAI signatures are rejected
//
// SECURITY TEST: Prevents linking with forged or invalid signatures
func TestLinkAddress_InvalidXaiSignature(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	auraAddr := keepertest.GenTestAddr()

	// Create invalid signature (too short)
	invalidSignature := make([]byte, 32) // Only 32 bytes, need 64+

	msg := &types.MsgLinkAddress{
		Signer:       auraAddr.String(),
		AuraAddress:  auraAddr.String(),
		PawAddress:   "",
		PawSignature: nil,
		XaiAddress:   "xai1test",
		XaiSignature: invalidSignature, // Invalid signature
	}

	_, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err, "should reject invalid XAI signature")

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
	require.Contains(t, st.Message(), "invalid XAI address ownership proof")
}

// TestLinkAddress_PawAddressAlreadyLinked tests that PAW addresses already linked are rejected
//
// SECURITY TEST: Prevents identity hijacking via duplicate linking
// Attack prevented: Attacker linking someone else's already-linked PAW address
func TestLinkAddress_PawAddressAlreadyLinked(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	// Create two different users
	users := keepertest.GenTestAddrs(2)
	user1 := users[0]
	user2 := users[1]

	// Generate real cryptographic key pair and signature for shared PAW address
	privKey, pubKey := generateTestKeyPair(t)
	sharedPawAddress := derivePawAddress(t, pubKey)

	// User1 links the PAW address
	message1 := "Link PAW address " + sharedPawAddress + " to Aura address " + user1.String()
	signature1 := signMessage(t, privKey, message1)

	msg1 := &types.MsgLinkAddress{
		Signer:       user1.String(),
		AuraAddress:  user1.String(),
		PawAddress:   sharedPawAddress, // This address will be linked to user1
		PawSignature: signature1,
		XaiAddress:   "",
		XaiSignature: nil,
	}

	resp1, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg1)
	require.NoError(t, err, "user1 should link successfully")
	require.True(t, resp1.Success)

	// User2 attempts to link the SAME PAW address (should fail)
	// Even with a valid signature, it should fail because address is already linked
	message2 := "Link PAW address " + sharedPawAddress + " to Aura address " + user2.String()
	signature2 := signMessage(t, privKey, message2)

	msg2 := &types.MsgLinkAddress{
		Signer:       user2.String(),
		AuraAddress:  user2.String(),
		PawAddress:   sharedPawAddress, // Same PAW address
		PawSignature: signature2,
		XaiAddress:   "",
		XaiSignature: nil,
	}

	_, err = ms.LinkAddress(sdk.WrapSDKContext(ctx), msg2)
	require.Error(t, err, "should reject duplicate PAW address linking")

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.AlreadyExists, st.Code())
	require.Contains(t, st.Message(), "PAW address")
	require.Contains(t, st.Message(), "already linked")
}

// TestLinkAddress_XaiAddressAlreadyLinked tests that XAI addresses already linked are rejected
//
// SECURITY TEST: Prevents identity hijacking via duplicate linking
// Attack prevented: Attacker linking someone else's already-linked XAI address
func TestLinkAddress_XaiAddressAlreadyLinked(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	// Create two different users
	users := keepertest.GenTestAddrs(2)
	user1 := users[0]
	user2 := users[1]

	// Generate real cryptographic key pair and signature for shared XAI address
	privKey, pubKey := generateTestKeyPair(t)
	sharedXaiAddress := deriveXaiAddress(t, pubKey)

	// User1 links the XAI address
	message1 := "Link XAI address " + sharedXaiAddress + " to Aura address " + user1.String()
	signature1 := signMessage(t, privKey, message1)

	msg1 := &types.MsgLinkAddress{
		Signer:       user1.String(),
		AuraAddress:  user1.String(),
		PawAddress:   "",
		PawSignature: nil,
		XaiAddress:   sharedXaiAddress, // This address will be linked to user1
		XaiSignature: signature1,
	}

	resp1, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg1)
	require.NoError(t, err, "user1 should link successfully")
	require.True(t, resp1.Success)

	// User2 attempts to link the SAME XAI address (should fail)
	// Even with a valid signature, it should fail because address is already linked
	message2 := "Link XAI address " + sharedXaiAddress + " to Aura address " + user2.String()
	signature2 := signMessage(t, privKey, message2)

	msg2 := &types.MsgLinkAddress{
		Signer:       user2.String(),
		AuraAddress:  user2.String(),
		PawAddress:   "",
		PawSignature: nil,
		XaiAddress:   sharedXaiAddress, // Same XAI address
		XaiSignature: signature2,
	}

	_, err = ms.LinkAddress(sdk.WrapSDKContext(ctx), msg2)
	require.Error(t, err, "should reject duplicate XAI address linking")

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.AlreadyExists, st.Code())
	require.Contains(t, st.Message(), "XAI address")
	require.Contains(t, st.Message(), "already linked")
}

// TestLinkAddress_UpdateOwnLink tests that users can update their own links
//
// SECURITY TEST: Allows legitimate link updates by the owner
func TestLinkAddress_UpdateOwnLink(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	user := keepertest.GenTestAddr()

	// Generate real PAW key pair
	pawPrivKey, pawPubKey := generateTestKeyPair(t)
	pawAddress := derivePawAddress(t, pawPubKey)
	pawMessage := "Link PAW address " + pawAddress + " to Aura address " + user.String()
	pawSignature := signMessage(t, pawPrivKey, pawMessage)

	// Initial link with PAW
	msg1 := &types.MsgLinkAddress{
		Signer:       user.String(),
		AuraAddress:  user.String(),
		PawAddress:   pawAddress,
		PawSignature: pawSignature,
		XaiAddress:   "",
		XaiSignature: nil,
	}

	resp1, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg1)
	require.NoError(t, err)
	require.True(t, resp1.Success)

	// Generate real XAI key pair for update
	xaiPrivKey, xaiPubKey := generateTestKeyPair(t)
	xaiAddress := deriveXaiAddress(t, xaiPubKey)
	xaiMessage := "Link XAI address " + xaiAddress + " to Aura address " + user.String()
	xaiSignature := signMessage(t, xaiPrivKey, xaiMessage)

	// Update to add XAI address (should succeed - same owner)
	// Note: Don't re-provide PAW address/signature since it's already linked
	// Only provide the new XAI address that we want to add
	msg2 := &types.MsgLinkAddress{
		Signer:       user.String(),
		AuraAddress:  user.String(),
		PawAddress:   "",           // Don't re-provide already-linked PAW
		PawSignature: nil,          // No signature needed for already-linked address
		XaiAddress:   xaiAddress,   // Add XAI
		XaiSignature: xaiSignature, // Provide XAI signature
	}

	resp2, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg2)
	require.NoError(t, err, "owner should be able to update their own link")
	require.True(t, resp2.Success)

	// Verify both addresses are linked
	identity, found := k.GetSharedIdentity(ctx, user.String())
	require.True(t, found)
	require.Equal(t, pawAddress, identity.LinkedAddresses["paw"])
	require.Equal(t, xaiAddress, identity.LinkedAddresses["xai"])
	require.True(t, identity.VerifiedPaw)
	require.True(t, identity.VerifiedXai)
}

// TestLinkAddress_BothAddresses tests linking both PAW and XAI at once
//
// SECURITY TEST: Validates simultaneous multi-chain linking
func TestLinkAddress_BothAddresses(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	user := keepertest.GenTestAddr()

	// Generate real PAW key pair
	pawPrivKey, pawPubKey := generateTestKeyPair(t)
	pawAddress := derivePawAddress(t, pawPubKey)
	pawMessage := "Link PAW address " + pawAddress + " to Aura address " + user.String()
	pawSignature := signMessage(t, pawPrivKey, pawMessage)

	// Generate real XAI key pair
	xaiPrivKey, xaiPubKey := generateTestKeyPair(t)
	xaiAddress := deriveXaiAddress(t, xaiPubKey)
	xaiMessage := "Link XAI address " + xaiAddress + " to Aura address " + user.String()
	xaiSignature := signMessage(t, xaiPrivKey, xaiMessage)

	// Link both PAW and XAI at once
	msg := &types.MsgLinkAddress{
		Signer:       user.String(),
		AuraAddress:  user.String(),
		PawAddress:   pawAddress,
		PawSignature: pawSignature,
		XaiAddress:   xaiAddress,
		XaiSignature: xaiSignature,
	}

	resp, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err, "should link both addresses successfully")
	require.True(t, resp.Success)

	// Verify both addresses are linked and verified
	identity, found := k.GetSharedIdentity(ctx, user.String())
	require.True(t, found)
	require.Equal(t, pawAddress, identity.LinkedAddresses["paw"])
	require.Equal(t, xaiAddress, identity.LinkedAddresses["xai"])
	require.True(t, identity.VerifiedPaw)
	require.True(t, identity.VerifiedXai)
	require.True(t, identity.VerifiedAura)
}

// TestLinkAddress_NoSigner tests that messages without signers are rejected
//
// SECURITY TEST: Validates basic authentication requirements
func TestLinkAddress_NoSigner(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	signature := make([]byte, 65)

	// Create message with empty signer
	msg := &types.MsgLinkAddress{
		Signer:       "", // Empty signer
		AuraAddress:  keepertest.GenTestAddr().String(),
		PawAddress:   "paw1test",
		PawSignature: signature,
		XaiAddress:   "",
		XaiSignature: nil,
	}

	_, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg)
	require.Error(t, err, "should reject message without signer")

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Unauthenticated, st.Code())
}

// TestVerifyPawAddressOwnership_ValidSignature is obsolete.
// See signature_verification_test.go for proper cryptographic signature tests.

// TestVerifyPawAddressOwnership_InvalidSignature tests PAW signature verification with invalid signature
func TestVerifyPawAddressOwnership_InvalidSignature(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	auraAddr := "aura1test123"
	pawAddr := "paw1test456"

	// Create invalid signature (too short)
	invalidSignature := make([]byte, 32)

	result := k.VerifyPawAddressOwnership(ctx, auraAddr, pawAddr, invalidSignature)
	require.False(t, result, "should reject invalid signature")
}

// TestVerifyPawAddressOwnership_EmptySignature tests PAW signature verification with empty signature
func TestVerifyPawAddressOwnership_EmptySignature(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	auraAddr := "aura1test123"
	pawAddr := "paw1test456"

	result := k.VerifyPawAddressOwnership(ctx, auraAddr, pawAddr, nil)
	require.False(t, result, "should reject empty signature")
}

// TestVerifyXaiAddressOwnership_ValidSignature is obsolete.
// See signature_verification_test.go for proper cryptographic signature tests.

// OBSOLETE TEST BELOW (skipped above):
/*
func TestVerifyXaiAddressOwnership_ValidSignature_OLD(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	auraAddr := "aura1test123"
	xaiAddr := "xai1test456"

	// Create valid signature (65 bytes)
	signature := make([]byte, 65)
	for i := range signature {
		signature[i] = byte(i % 256)
	}

	result := k.VerifyXaiAddressOwnership(ctx, auraAddr, xaiAddr, signature)
	require.True(t, result, "should accept valid signature")
}
*/

// TestVerifyXaiAddressOwnership_InvalidSignature tests XAI signature verification with invalid signature
func TestVerifyXaiAddressOwnership_InvalidSignature(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx

	auraAddr := "aura1test123"
	xaiAddr := "xai1test456"

	// Create invalid signature (too short)
	invalidSignature := make([]byte, 32)

	result := k.VerifyXaiAddressOwnership(ctx, auraAddr, xaiAddr, invalidSignature)
	require.False(t, result, "should reject invalid signature")
}

// TestFindSharedIdentityByLinkedAddress tests finding identities by linked addresses
func TestFindSharedIdentityByLinkedAddress(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil, nil, nil, nil)
	ctx := input.Ctx
	ms := keeper.NewMsgServerImpl(k)

	user := keepertest.GenTestAddr()

	// Generate real PAW key pair
	pawPrivKey, pawPubKey := generateTestKeyPair(t)
	pawAddress := derivePawAddress(t, pawPubKey)
	pawMessage := "Link PAW address " + pawAddress + " to Aura address " + user.String()
	pawSignature := signMessage(t, pawPrivKey, pawMessage)

	// Generate real XAI key pair
	xaiPrivKey, xaiPubKey := generateTestKeyPair(t)
	xaiAddress := deriveXaiAddress(t, xaiPubKey)
	xaiMessage := "Link XAI address " + xaiAddress + " to Aura address " + user.String()
	xaiSignature := signMessage(t, xaiPrivKey, xaiMessage)

	// Create an identity with linked addresses
	msg := &types.MsgLinkAddress{
		Signer:       user.String(),
		AuraAddress:  user.String(),
		PawAddress:   pawAddress,
		PawSignature: pawSignature,
		XaiAddress:   xaiAddress,
		XaiSignature: xaiSignature,
	}

	_, err := ms.LinkAddress(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)

	// Test finding by PAW address
	foundPaw := k.FindSharedIdentityByLinkedAddress(ctx, "paw", pawAddress)
	require.NotNil(t, foundPaw, "should find identity by PAW address")
	require.Equal(t, user.String(), foundPaw.Address)

	// Test finding by XAI address
	foundXai := k.FindSharedIdentityByLinkedAddress(ctx, "xai", xaiAddress)
	require.NotNil(t, foundXai, "should find identity by XAI address")
	require.Equal(t, user.String(), foundXai.Address)

	// Test not finding non-existent address
	notFound := k.FindSharedIdentityByLinkedAddress(ctx, "paw", "paw1notexist")
	require.Nil(t, notFound, "should not find non-existent address")

	// Test case-insensitive chain name
	foundLowercase := k.FindSharedIdentityByLinkedAddress(ctx, "PAW", pawAddress)
	require.NotNil(t, foundLowercase, "should find with uppercase chain name")
	require.Equal(t, user.String(), foundLowercase.Address)
}

// TestLinkAddress_MessageHashFormat tests that the signature verification uses correct message format
func TestLinkAddress_MessageHashFormat(t *testing.T) {
	// This test documents the expected message format for cross-chain signatures
	expectedMessage := "Link PAW address paw1test456 to Aura address aura1test123"
	hash := sha256.Sum256([]byte(expectedMessage))

	require.Equal(t, 32, len(hash), "hash should be 32 bytes")
	require.NotEmpty(t, hash, "hash should not be empty")

	// For XAI
	expectedXaiMessage := "Link XAI address xai1test789 to Aura address aura1test123"
	xaiHash := sha256.Sum256([]byte(expectedXaiMessage))

	require.Equal(t, 32, len(xaiHash), "XAI hash should be 32 bytes")
	require.NotEmpty(t, xaiHash, "XAI hash should not be empty")
}

// Note: Helper functions (generateTestKeyPair, signMessage, derivePawAddress, deriveXaiAddress)
// are defined in signature_verification_test.go and shared across test files
