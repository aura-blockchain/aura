package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/identity/types"
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

// TestMsgServerUpdateParams_Success tests successful param update
func TestMsgServerUpdateParams_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	// Set initial params
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Update with valid authority
	newParams := types.DefaultParams()
	msg := &identitypb.MsgUpdateParams{
		Authority: "authority",
		Params:    *newParams,
	}

	resp, err := msgServer.UpdateParams(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify params were updated
	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, *newParams, *params)
}

// TestMsgServerUpdateParams_InvalidAuthority tests invalid authority rejection
func TestMsgServerUpdateParams_InvalidAuthority(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Try update with invalid authority
	msg := &identitypb.MsgUpdateParams{
		Authority: "invalid",
		Params:    *types.DefaultParams(),
	}

	_, err := msgServer.UpdateParams(ctx, msg)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, st.Code())
}

// TestMsgServerEraseIdentity_Success tests successful identity erasure
func TestMsgServerEraseIdentity_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Create identity
	did := "did:aura:test"
	now := ctx.BlockTime()
	record := &types.IdentityRecord{
		Did:       did,
		Address:   "aura1test",
		Status:    types.IdentityStatusActive,
		CreatedAt: now,
		UpdatedAt: &now,
	}
	require.NoError(t, keeper.SetIdentityRecord(ctx, record))

	// Erase identity
	msg := &identitypb.MsgEraseIdentity{
		Requester: "aura1test",
		Did:       did,
		Reason:    "GDPR request",
	}

	resp, err := msgServer.EraseIdentity(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify erasure
	erasedRecord, err := keeper.GetIdentityRecord(ctx, did)
	require.NoError(t, err)
	require.True(t, erasedRecord.Erased)
	require.Equal(t, types.IdentityStatusErased, erasedRecord.Status)
}

// TestMsgServerEraseIdentity_NotFound tests erasure of non-existent identity
func TestMsgServerEraseIdentity_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	msg := &identitypb.MsgEraseIdentity{
		Requester: "aura1test",
		Did:       "did:aura:nonexistent",
		Reason:    "Test",
	}

	_, err := msgServer.EraseIdentity(ctx, msg)
	require.Error(t, err)
}

// TestMsgServerCreateMultisigWallet_Success tests successful wallet creation
func TestMsgServerCreateMultisigWallet_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	creator := "aura1creator"

	// Set up role with manage_multisig permission
	err := keeper.SetRole(ctx, &types.Role{
		Name:        "multisig_manager",
		Permissions: []string{types.PermissionManageMultisig},
	})
	require.NoError(t, err)

	// Assign role to creator
	err = keeper.SetRoleAssignment(ctx, &types.RoleAssignment{
		Address:  creator,
		RoleName: "multisig_manager",
	})
	require.NoError(t, err)

	msg := &identitypb.MsgCreateMultisigWallet{
		Creator:   creator,
		Signers:   []string{"aura1owner1", "aura1owner2"},
		Threshold: 2,
	}

	resp, err := msgServer.CreateMultisigWallet(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.WalletId)

	// Verify wallet was created
	wallet, err := keeper.GetMultisigWallet(ctx, resp.WalletId)
	require.NoError(t, err)
	require.Equal(t, uint32(2), wallet.Threshold)
	require.Equal(t, 2, len(wallet.Signers))
}

// TestMsgServerCreateMultisigWallet_InvalidThreshold tests invalid threshold
func TestMsgServerCreateMultisigWallet_InvalidThreshold(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Threshold exceeds number of signers
	msg := &identitypb.MsgCreateMultisigWallet{
		Creator:   "aura1creator",
		Signers:   []string{"aura1owner1", "aura1owner2"},
		Threshold: 3,
	}

	_, err := msgServer.CreateMultisigWallet(ctx, msg)
	require.Error(t, err)
}

// TestMsgServerCreateMultisigWallet_ZeroThreshold tests zero threshold rejection
func TestMsgServerCreateMultisigWallet_ZeroThreshold(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	msg := &identitypb.MsgCreateMultisigWallet{
		Creator:   "aura1creator",
		Signers:   []string{"aura1owner1", "aura1owner2"},
		Threshold: 0,
	}

	_, err := msgServer.CreateMultisigWallet(ctx, msg)
	require.Error(t, err)
}

// TestMsgServerCreateMultisigWallet_EmptySigners tests empty signers rejection
func TestMsgServerCreateMultisigWallet_EmptySigners(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	msg := &identitypb.MsgCreateMultisigWallet{
		Creator:   "aura1creator",
		Signers:   []string{},
		Threshold: 1,
	}

	_, err := msgServer.CreateMultisigWallet(ctx, msg)
	require.Error(t, err)
}

// TestMsgServerCreateSession_Success tests successful session creation
func TestMsgServerCreateSession_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	msg := &identitypb.MsgCreateSession{
		Address: "aura1test",
	}

	resp, err := msgServer.CreateSession(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.SessionId)

	// Verify session was created
	session, err := keeper.GetSession(ctx, resp.SessionId)
	require.NoError(t, err)
	require.True(t, session.IsActive)
	require.Equal(t, "aura1test", session.Address)
}

// TestMsgServerEndSession_Success tests successful session termination
func TestMsgServerEndSession_Success(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	// Create session first
	createMsg := &identitypb.MsgCreateSession{
		Address: "aura1test",
	}

	createResp, err := msgServer.CreateSession(ctx, createMsg)
	require.NoError(t, err)
	sessionID := createResp.SessionId

	// End session
	endMsg := &identitypb.MsgEndSession{
		Address:   "aura1test",
		SessionId: sessionID,
	}

	resp, err := msgServer.EndSession(ctx, endMsg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify session is deleted (RevokeSession deletes the session)
	_, err = keeper.GetSession(ctx, sessionID)
	require.Error(t, err, "session should be deleted after ending")
}

// TestMsgServerEndSession_NotFound tests ending non-existent session
func TestMsgServerEndSession_NotFound(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	msg := &identitypb.MsgEndSession{
		Address:   "aura1test",
		SessionId: "nonexistent",
	}

	_, err := msgServer.EndSession(ctx, msg)
	require.Error(t, err)
}
