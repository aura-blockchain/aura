package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/identity/types"
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

// TestMsgServerUpdateParams_EdgeCases tests edge cases for UpdateParams
func TestMsgServerUpdateParams_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Keeper, sdk.Context)
		msg     *identitypb.MsgUpdateParams
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success - valid authority",
			setup: func(k *Keeper, ctx sdk.Context) {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
			},
			msg: &identitypb.MsgUpdateParams{
				Authority: "authority",
				Params:    types.DefaultParams(),
			},
			wantErr: false,
		},
		{
			name: "failure - invalid authority",
			setup: func(k *Keeper, ctx sdk.Context) {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
			},
			msg: &identitypb.MsgUpdateParams{
				Authority: "invalid",
				Params:    types.DefaultParams(),
			},
			wantErr: true,
			errCode: codes.PermissionDenied,
		},
		{
			name: "failure - empty authority",
			setup: func(k *Keeper, ctx sdk.Context) {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
			},
			msg: &identitypb.MsgUpdateParams{
				Authority: "",
				Params:    types.DefaultParams(),
			},
			wantErr: true,
			errCode: codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupKeeperForTest(t)
			msgServer := NewMsgServerImpl(keeper)

			if tt.setup != nil {
				tt.setup(keeper, ctx)
			}

			resp, err := msgServer.UpdateParams(ctx, tt.msg)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.errCode, st.Code())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// Verify params were updated
			params, err := keeper.GetParams(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.msg.Params, *params)
		})
	}
}

// TestMsgServerEraseIdentity_EdgeCases tests edge cases for EraseIdentity
func TestMsgServerEraseIdentity_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Keeper, sdk.Context) string
		msg     func(string) *identitypb.MsgEraseIdentity
		wantErr bool
	}{
		{
			name: "success - erase identity",
			setup: func(k *Keeper, ctx sdk.Context) string {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

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
				require.NoError(t, k.SetIdentityRecord(ctx, record))
				return did
			},
			msg: func(did string) *identitypb.MsgEraseIdentity {
				return &identitypb.MsgEraseIdentity{
					Requester: "aura1test",
					Did:       did,
					Reason:    "GDPR request",
				}
			},
			wantErr: false,
		},
		{
			name: "failure - identity not found",
			setup: func(k *Keeper, ctx sdk.Context) string {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
				return "did:aura:nonexistent"
			},
			msg: func(did string) *identitypb.MsgEraseIdentity {
				return &identitypb.MsgEraseIdentity{
					Requester: "aura1test",
					Did:       did,
					Reason:    "Test",
				}
			},
			wantErr: true,
		},
		{
			name: "failure - already erased",
			setup: func(k *Keeper, ctx sdk.Context) string {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

				// Create and erase identity
				did := "did:aura:test"
				now := ctx.BlockTime()
				record := &types.IdentityRecord{
					Did:       did,
					Address:   "aura1test",
					Status:    types.IdentityStatusActive,
					CreatedAt: now,
					UpdatedAt: &now,
				}
				require.NoError(t, k.SetIdentityRecord(ctx, record))

				// Erase it
				require.NoError(t, k.EraseIdentity(ctx, did, "aura1test", "first erasure"))
				return did
			},
			msg: func(did string) *identitypb.MsgEraseIdentity {
				return &identitypb.MsgEraseIdentity{
					Requester: "aura1test",
					Did:       did,
					Reason:    "Second erasure attempt",
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupKeeperForTest(t)
			msgServer := NewMsgServerImpl(keeper)

			var did string
			if tt.setup != nil {
				did = tt.setup(keeper, ctx)
			}

			resp, err := msgServer.EraseIdentity(ctx, tt.msg(did))

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// Verify identity was erased
			record, err := keeper.GetIdentityRecord(ctx, did)
			require.NoError(t, err)
			require.True(t, record.Erased)
			require.Equal(t, types.IdentityStatusErased, record.Status)
		})
	}
}

// TestMsgServerCreateMultisigWallet_ValidationTests tests validation for CreateMultisigWallet
func TestMsgServerCreateMultisigWallet_ValidationTests(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Keeper, sdk.Context)
		msg     *identitypb.MsgCreateMultisigWallet
		wantErr bool
	}{
		{
			name: "success - minimum valid wallet",
			setup: func(k *Keeper, ctx sdk.Context) {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
			},
			msg: &identitypb.MsgCreateMultisigWallet{
				Creator:   "aura1creator",
				Signers:   []string{"aura1owner1", "aura1owner2"},
				Threshold: 1,
			},
			wantErr: false,
		},
		{
			name: "success - threshold equals owners",
			setup: func(k *Keeper, ctx sdk.Context) {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
			},
			msg: &identitypb.MsgCreateMultisigWallet{
				Creator:   "aura1creator",
				Signers:   []string{"aura1owner1", "aura1owner2", "aura1owner3"},
				Threshold: 3,
			},
			wantErr: false,
		},
		{
			name: "failure - threshold exceeds owners",
			setup: func(k *Keeper, ctx sdk.Context) {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
			},
			msg: &identitypb.MsgCreateMultisigWallet{
				Creator:   "aura1creator",
				Owners:    []string{"aura1owner1", "aura1owner2"},
				Threshold: 3,
			},
			wantErr: true,
		},
		{
			name: "failure - zero threshold",
			setup: func(k *Keeper, ctx sdk.Context) {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
			},
			msg: &identitypb.MsgCreateMultisigWallet{
				Creator:   "aura1creator",
				Owners:    []string{"aura1owner1", "aura1owner2"},
				Threshold: 0,
			},
			wantErr: true,
		},
		{
			name: "failure - empty owners",
			setup: func(k *Keeper, ctx sdk.Context) {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
			},
			msg: &identitypb.MsgCreateMultisigWallet{
				Creator:   "aura1creator",
				Signers:   []string{},
				Threshold: 1,
			},
			wantErr: true,
		},
		{
			name: "failure - single owner",
			setup: func(k *Keeper, ctx sdk.Context) {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
			},
			msg: &identitypb.MsgCreateMultisigWallet{
				Creator:   "aura1creator",
				Owners:    []string{"aura1owner1"},
				Threshold: 1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupKeeperForTest(t)
			msgServer := NewMsgServerImpl(keeper)

			if tt.setup != nil {
				tt.setup(keeper, ctx)
			}

			resp, err := msgServer.CreateMultisigWallet(ctx, tt.msg)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotEmpty(t, resp.WalletId)

			// Verify wallet was created
			wallet, err := keeper.GetMultisigWallet(ctx, resp.WalletId)
			require.NoError(t, err)
			require.Equal(t, tt.msg.Threshold, wallet.Threshold)
			require.Equal(t, len(tt.msg.Owners), len(wallet.Owners))
		})
	}
}

// TestMsgServerExecuteTimeLockedAction_TimingTests tests timing constraints
func TestMsgServerExecuteTimeLockedAction_TimingTests(t *testing.T) {
	tests := []struct {
		name       string
		setupDelay time.Duration
		wantErr    bool
	}{
		{
			name:       "failure - too early",
			setupDelay: 1 * time.Hour, // Action executable 1 hour in future
			wantErr:    true,
		},
		{
			name:       "success - exactly at time",
			setupDelay: 0, // Action executable now
			wantErr:    false,
		},
		{
			name:       "success - past executable time",
			setupDelay: -1 * time.Hour, // Action executable 1 hour ago
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupKeeperForTest(t)
			msgServer := NewMsgServerImpl(keeper)
			require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

			// Create time-locked action
			actionID := "action_test"
			now := ctx.BlockTime()
			action := &types.TimeLockedAction{
				ActionId:     actionID,
				Proposer:     "aura1proposer",
				Status:       types.ActionStatusPending,
				ProposedAt:   now,
				ExecutableAt: now.Add(tt.setupDelay),
				Action:       "test_action",
				ActionData:   []byte("test_data"),
			}
			require.NoError(t, keeper.SetTimeLockedAction(ctx, action))

			// Try to execute
			msg := &identitypb.MsgExecuteTimeLockedAction{
				ActionId: actionID,
				Executor: "aura1executor",
			}

			resp, err := msgServer.ExecuteTimeLockedAction(ctx, msg)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// Verify action status changed
			updatedAction, err := keeper.GetTimeLockedAction(ctx, actionID)
			require.NoError(t, err)
			require.Equal(t, types.ActionStatusExecuted, updatedAction.Status)
			require.NotNil(t, updatedAction.ExecutedAt)
		})
	}
}

// TestMsgServerCreateSession_EdgeCases tests edge cases for CreateSession
func TestMsgServerCreateSession_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Keeper, sdk.Context)
		msg      *identitypb.MsgCreateSession
		wantErr  bool
		validate func(*testing.T, *Keeper, sdk.Context, *identitypb.MsgCreateSessionResponse)
	}{
		{
			name: "success - create session with duration",
			setup: func(k *Keeper, ctx sdk.Context) {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
			},
			msg: &identitypb.MsgCreateSession{
				Address:      "aura1test",
				DurationSecs: 3600, // 1 hour
				SessionData:  []byte("test_data"),
			},
			wantErr: false,
			validate: func(t *testing.T, k *Keeper, ctx sdk.Context, resp *identitypb.MsgCreateSessionResponse) {
				require.NotEmpty(t, resp.SessionId)

				// Verify session was created
				session, err := k.GetSession(ctx, resp.SessionId)
				require.NoError(t, err)
				require.True(t, session.IsActive)
				require.Equal(t, "aura1test", session.Address)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupKeeperForTest(t)
			msgServer := NewMsgServerImpl(keeper)

			if tt.setup != nil {
				tt.setup(keeper, ctx)
			}

			resp, err := msgServer.CreateSession(ctx, tt.msg)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tt.validate != nil {
				tt.validate(t, keeper, ctx, resp)
			}
		})
	}
}

// TestMsgServerEndSession_EdgeCases tests edge cases for EndSession
func TestMsgServerEndSession_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Keeper, sdk.Context) string
		msg     func(string) *identitypb.MsgEndSession
		wantErr bool
	}{
		{
			name: "success - end active session",
			setup: func(k *Keeper, ctx sdk.Context) string {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

				// Create active session
				sessionID := "session_test"
				now := ctx.BlockTime()
				session := &types.Session{
					SessionId: sessionID,
					Address:   "aura1test",
					CreatedAt: now,
					ExpiresAt: now.Add(1 * time.Hour),
					IsActive:  true,
				}
				require.NoError(t, k.SetSession(ctx, session))
				return sessionID
			},
			msg: func(sessionID string) *identitypb.MsgEndSession {
				return &identitypb.MsgEndSession{
					SessionId: sessionID,
					Address:   "aura1test",
				}
			},
			wantErr: false,
		},
		{
			name: "failure - session not found",
			setup: func(k *Keeper, ctx sdk.Context) string {
				require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
				return "nonexistent_session"
			},
			msg: func(sessionID string) *identitypb.MsgEndSession {
				return &identitypb.MsgEndSession{
					SessionId: sessionID,
					Address:   "aura1test",
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupKeeperForTest(t)
			msgServer := NewMsgServerImpl(keeper)

			var sessionID string
			if tt.setup != nil {
				sessionID = tt.setup(keeper, ctx)
			}

			resp, err := msgServer.EndSession(ctx, tt.msg(sessionID))

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// Verify session is no longer active
			session, err := keeper.GetSession(ctx, sessionID)
			require.NoError(t, err)
			require.False(t, session.IsActive)
		})
	}
}
