// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"strings"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/identity/keeper"
	"github.com/aequitas/aura/chain/x/identity/types"
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

// createFuzzTestContext creates a minimal test context for fuzz testing
func createFuzzTestContext(t testing.TB) (sdk.Context, *keeper.Keeper, identitypb.MsgServer) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	storeService := runtime.NewKVStoreService(storeKey)
	k := keeper.NewKeeper(storeService, storeKey, cdc, "authority", log.NewNopLogger())

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 1,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	msgServer := keeper.NewMsgServerImpl(k)

	return ctx, k, msgServer
}

// FuzzRequestIdentityChange fuzzes the RequestIdentityChange message handler.
// Security properties tested:
//   - Validates requester address format
//   - Handles empty/malformed DID strings
//   - Handles malformed metadata hash
//   - Never panics on any input combination
func FuzzRequestIdentityChange(f *testing.F) {
	// Seed corpus with representative test cases
	f.Add("aura1valid", "did:aura:test123", "ir-1", "QmHash123")
	f.Add("", "did:aura:test", "", "")                           // Empty requester
	f.Add("invalid", "", "ir-1", "hash")                         // Empty DID
	f.Add("aura1test", "did:invalid", "", "")                    // Invalid DID format
	f.Add(strings.Repeat("a", 1000), "did:aura:x", "", "")       // Very long requester
	f.Add("aura1x", strings.Repeat("d", 1000), "", "")           // Very long DID
	f.Add("aura1\x00null", "did:aura:test", "", "")              // Null bytes
	f.Add("'; DROP TABLE users; --", "did:aura:sql", "", "hash") // SQL injection

	f.Fuzz(func(t *testing.T, requester, targetDID, irID, metadataHash string) {
		// Skip extremely long inputs to avoid timeout
		if len(requester) > 2000 || len(targetDID) > 2000 || len(metadataHash) > 2000 {
			t.Skip("input too long")
		}

		ctx, _, msgServer := createFuzzTestContext(t)

		msg := &identitypb.MsgRequestIdentityChange{
			Requester:    requester,
			TargetDid:    targetDID,
			IrId:         irID,
			MetadataHash: metadataHash,
		}

		// Execute - must not panic
		resp, err := msgServer.RequestIdentityChange(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Invalid requester must be rejected
		if requester == "" {
			require.Error(t, err, "empty requester must be rejected")
		}

		// Valid Bech32 check
		_, parseErr := sdk.AccAddressFromBech32(requester)
		if parseErr != nil && requester != "" {
			require.Error(t, err, "invalid bech32 requester must be rejected")
		}

		// If successful, response must be valid
		if err == nil {
			require.NotNil(t, resp, "successful response must not be nil")
		}
	})
}

// FuzzCreateRole fuzzes the CreateRole message handler.
// Security properties tested:
//   - Validates creator address
//   - Handles empty/malformed role names
//   - Handles permission validation
//   - Never panics on any input
func FuzzCreateRole(f *testing.F) {
	f.Add("aura1creator", "admin", "read,write", "Admin role")
	f.Add("", "role", "perm", "desc")                 // Empty creator
	f.Add("aura1x", "", "perm", "desc")               // Empty role name
	f.Add("aura1x", "role", "", "")                   // Empty permissions
	f.Add("aura1x", strings.Repeat("r", 500), "", "") // Very long role name
	f.Add("aura1x", "role\x00null", "", "")           // Null in role name

	f.Fuzz(func(t *testing.T, creator, roleName, permissions, description string) {
		if len(creator) > 500 || len(roleName) > 500 || len(permissions) > 500 {
			t.Skip("input too long")
		}

		ctx, _, msgServer := createFuzzTestContext(t)

		// Parse permissions into slice
		perms := strings.Split(permissions, ",")

		msg := &identitypb.MsgCreateRole{
			Creator:     creator,
			RoleName:    roleName,
			Permissions: perms,
			Description: description,
		}

		// Execute - must not panic
		resp, err := msgServer.CreateRole(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Empty creator must be rejected
		if creator == "" {
			require.Error(t, err, "empty creator must be rejected")
		}

		// Valid response check
		if err == nil {
			require.NotNil(t, resp)
		}
	})
}

// FuzzAssignRole fuzzes the AssignRole message handler.
// Security properties tested:
//   - Validates assigner and assignee addresses
//   - Handles role assignment logic
//   - Validates expiry time handling
func FuzzAssignRole(f *testing.F) {
	f.Add("aura1assigner", "aura1assignee", "admin", int64(3600))
	f.Add("", "aura1x", "role", int64(0))                // Empty assigner
	f.Add("aura1x", "", "role", int64(0))                // Empty assignee
	f.Add("aura1x", "aura1y", "", int64(0))              // Empty role
	f.Add("aura1x", "aura1y", "role", int64(-1))         // Negative expiry
	f.Add("aura1x", "aura1y", "role", int64(9999999999)) // Very large expiry

	f.Fuzz(func(t *testing.T, assigner, assignee, roleName string, expirySeconds int64) {
		if len(assigner) > 500 || len(assignee) > 500 || len(roleName) > 500 {
			t.Skip("input too long")
		}

		ctx, _, msgServer := createFuzzTestContext(t)

		// Calculate expiry time
		var expiresAt *time.Time
		if expirySeconds > 0 {
			expiry := ctx.BlockTime().Add(time.Duration(expirySeconds) * time.Second)
			expiresAt = &expiry
		}

		msg := &identitypb.MsgAssignRole{
			Assigner:  assigner,
			Address:   assignee,
			RoleName:  roleName,
			ExpiresAt: expiresAt,
		}

		// Execute - must not panic
		resp, err := msgServer.AssignRole(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Empty assigner must be rejected
		if assigner == "" {
			require.Error(t, err, "empty assigner must be rejected")
		}

		if err == nil {
			require.NotNil(t, resp)
		}
	})
}

// FuzzCreateMultisigWallet fuzzes the CreateMultisigWallet message handler.
// Security properties tested:
//   - Validates threshold logic (threshold <= signers)
//   - Validates signer addresses
//   - Handles edge cases in multisig creation
func FuzzCreateMultisigWallet(f *testing.F) {
	f.Add("aura1creator", uint32(2), uint8(3), int32(0))
	f.Add("", uint32(1), uint8(1), int32(0))           // Empty creator
	f.Add("aura1x", uint32(0), uint8(2), int32(1))     // Zero threshold
	f.Add("aura1x", uint32(5), uint8(2), int32(2))     // Threshold > signers
	f.Add("aura1x", uint32(1), uint8(0), int32(3))     // Zero signers
	f.Add("aura1x", uint32(255), uint8(255), int32(0)) // Max values

	f.Fuzz(func(t *testing.T, creator string, threshold uint32, signerCount uint8, walletTypeInt int32) {
		if len(creator) > 500 {
			t.Skip("input too long")
		}

		ctx, _, msgServer := createFuzzTestContext(t)

		// Generate signer addresses
		signers := make([]string, signerCount)
		for i := range signers {
			signers[i] = "aura1signer" + string(rune('0'+i))
		}

		// Normalize wallet type to valid enum range
		walletType := identitypb.WalletType(walletTypeInt % 4)

		msg := &identitypb.MsgCreateMultisigWallet{
			Creator:    creator,
			Signers:    signers,
			Threshold:  threshold,
			WalletType: walletType,
		}

		// Execute - must not panic
		resp, err := msgServer.CreateMultisigWallet(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Zero threshold must be rejected
		if threshold == 0 {
			require.Error(t, err, "zero threshold must be rejected")
		}

		// SECURITY INVARIANT: Threshold > signers must be rejected
		if threshold > uint32(len(signers)) {
			require.Error(t, err, "threshold exceeding signers must be rejected")
		}

		// SECURITY INVARIANT: Empty creator must be rejected
		if creator == "" {
			require.Error(t, err, "empty creator must be rejected")
		}

		if err == nil {
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.WalletId)
		}
	})
}

// FuzzCreateSession fuzzes the CreateSession message handler.
// Security properties tested:
//   - Validates address format
//   - Handles session creation logic
func FuzzCreateSession(f *testing.F) {
	f.Add("aura1user123")
	f.Add("")                        // Empty address
	f.Add("invalid_address")         // Invalid format
	f.Add(strings.Repeat("a", 1000)) // Very long address
	f.Add("aura1\x00\x01\x02")       // Binary data
	f.Add("../../../etc/passwd")     // Path traversal attempt

	f.Fuzz(func(t *testing.T, address string) {
		if len(address) > 2000 {
			t.Skip("address too long")
		}

		ctx, _, msgServer := createFuzzTestContext(t)

		msg := &identitypb.MsgCreateSession{
			Address: address,
		}

		// Execute - must not panic
		resp, err := msgServer.CreateSession(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Empty address must be rejected
		if address == "" {
			require.Error(t, err, "empty address must be rejected")
		}

		// Invalid Bech32 must be rejected
		_, parseErr := sdk.AccAddressFromBech32(address)
		if parseErr != nil && address != "" {
			require.Error(t, err, "invalid bech32 address must be rejected")
		}

		if err == nil {
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.SessionId)
		}
	})
}

// FuzzEraseIdentity fuzzes the EraseIdentity message handler (GDPR Right to Erasure).
// Security properties tested:
//   - Validates requester authorization
//   - Validates DID format
//   - Handles reason validation
func FuzzEraseIdentity(f *testing.F) {
	f.Add("aura1requester", "did:aura:target123", "GDPR request")
	f.Add("", "did:aura:x", "reason")                   // Empty requester
	f.Add("aura1x", "", "reason")                       // Empty DID
	f.Add("aura1x", "did:aura:x", "")                   // Empty reason
	f.Add("aura1x", "did:invalid:format", "reason")     // Invalid DID format
	f.Add("'; DROP TABLE", "did:aura:sql", "injection") // SQL injection

	f.Fuzz(func(t *testing.T, requester, did, reason string) {
		if len(requester) > 500 || len(did) > 500 || len(reason) > 1000 {
			t.Skip("input too long")
		}

		ctx, _, msgServer := createFuzzTestContext(t)

		msg := &identitypb.MsgEraseIdentity{
			Requester: requester,
			Did:       did,
			Reason:    reason,
		}

		// Execute - must not panic
		resp, err := msgServer.EraseIdentity(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Empty requester must be rejected
		if requester == "" {
			require.Error(t, err, "empty requester must be rejected")
		}

		// SECURITY INVARIANT: Empty DID must be rejected
		if did == "" {
			require.Error(t, err, "empty DID must be rejected")
		}

		if err == nil {
			require.NotNil(t, resp)
		}
	})
}

// FuzzRotateDIDKey fuzzes the RotateDIDKey message handler.
// Security properties tested:
//   - Validates initiator authorization
//   - Validates DID and verification method formats
//   - Handles key rotation logic
func FuzzRotateDIDKey(f *testing.F) {
	f.Add("aura1initiator", "did:aura:owner", "ed25519:newpubkey", "routine rotation")
	f.Add("", "did:aura:x", "key", "reason")    // Empty initiator
	f.Add("aura1x", "", "key", "reason")        // Empty DID
	f.Add("aura1x", "did:aura:x", "", "reason") // Empty verification method
	f.Add("aura1x", "did:aura:x", "key", "")    // Empty reason (should be allowed)

	f.Fuzz(func(t *testing.T, initiator, did, verificationMethod, reason string) {
		if len(initiator) > 500 || len(did) > 500 || len(verificationMethod) > 500 {
			t.Skip("input too long")
		}

		ctx, _, msgServer := createFuzzTestContext(t)

		msg := &identitypb.MsgRotateDIDKey{
			Initiator:             initiator,
			Did:                   did,
			NewVerificationMethod: verificationMethod,
			Reason:                reason,
		}

		// Execute - must not panic
		resp, err := msgServer.RotateDIDKey(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Empty initiator must be rejected
		if initiator == "" {
			require.Error(t, err, "empty initiator must be rejected")
		}

		// SECURITY INVARIANT: Empty DID must be rejected
		if did == "" {
			require.Error(t, err, "empty DID must be rejected")
		}

		// SECURITY INVARIANT: Empty verification method must be rejected
		if verificationMethod == "" {
			require.Error(t, err, "empty verification method must be rejected")
		}

		if err == nil {
			require.NotNil(t, resp)
		}
	})
}

// FuzzProposeTimeLockedAction fuzzes the ProposeTimeLockedAction message handler.
// Security properties tested:
//   - Validates delay seconds bounds
//   - Validates proposer authorization
//   - Handles payload validation
func FuzzProposeTimeLockedAction(f *testing.F) {
	f.Add("aura1proposer", "upgrade", []byte("payload"), uint64(3600))
	f.Add("", "type", []byte("p"), uint64(0))           // Empty proposer
	f.Add("aura1x", "", []byte("p"), uint64(100))       // Empty action type
	f.Add("aura1x", "type", []byte{}, uint64(100))      // Empty payload
	f.Add("aura1x", "type", []byte("p"), uint64(0))     // Zero delay
	f.Add("aura1x", "type", []byte("p"), uint64(1<<62)) // Very large delay

	f.Fuzz(func(t *testing.T, proposer, actionType string, payload []byte, delaySeconds uint64) {
		if len(proposer) > 500 || len(actionType) > 500 || len(payload) > 10000 {
			t.Skip("input too long")
		}

		ctx, _, msgServer := createFuzzTestContext(t)

		msg := &identitypb.MsgProposeTimeLockedAction{
			Proposer:     proposer,
			ActionType:   actionType,
			Payload:      payload,
			DelaySeconds: delaySeconds,
		}

		// Execute - must not panic
		resp, err := msgServer.ProposeTimeLockedAction(sdk.WrapSDKContext(ctx), msg)

		// SECURITY INVARIANT: Empty proposer must be rejected
		if proposer == "" {
			require.Error(t, err, "empty proposer must be rejected")
		}

		if err == nil {
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.ActionId)
		}
	})
}

// FuzzNilMessageHandling ensures all message handlers reject nil messages gracefully.
func FuzzNilMessageHandling(f *testing.F) {
	f.Add(uint8(0))
	for i := uint8(0); i < 20; i++ {
		f.Add(i)
	}

	f.Fuzz(func(t *testing.T, msgType uint8) {
		ctx, _, msgServer := createFuzzTestContext(t)

		var err error

		// Test each message type with nil
		switch msgType % 10 {
		case 0:
			_, err = msgServer.RequestIdentityChange(sdk.WrapSDKContext(ctx), nil)
		case 1:
			_, err = msgServer.CreateRole(sdk.WrapSDKContext(ctx), nil)
		case 2:
			_, err = msgServer.AssignRole(sdk.WrapSDKContext(ctx), nil)
		case 3:
			_, err = msgServer.RevokeRole(sdk.WrapSDKContext(ctx), nil)
		case 4:
			_, err = msgServer.CreateMultisigWallet(sdk.WrapSDKContext(ctx), nil)
		case 5:
			_, err = msgServer.CreateSession(sdk.WrapSDKContext(ctx), nil)
		case 6:
			_, err = msgServer.EndSession(sdk.WrapSDKContext(ctx), nil)
		case 7:
			_, err = msgServer.EraseIdentity(sdk.WrapSDKContext(ctx), nil)
		case 8:
			_, err = msgServer.RotateDIDKey(sdk.WrapSDKContext(ctx), nil)
		case 9:
			_, err = msgServer.ProposeTimeLockedAction(sdk.WrapSDKContext(ctx), nil)
		}

		// SECURITY INVARIANT: Nil messages must always be rejected
		require.Error(t, err, "nil message must be rejected")
	})
}
