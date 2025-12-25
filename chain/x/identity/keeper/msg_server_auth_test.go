// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/identity/types"
	identitypb "github.com/aequitas/aura/proto/aura/identity/v1beta1"
)

func TestMsgServerApplyIdentityChange_PermissionDenied(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)

	// Ensure params are initialized so keeper behavior matches production defaults
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	_, err := msgServer.ApplyIdentityChange(ctx, &identitypb.MsgApplyIdentityChange{
		RequestId: "req-unauthorized",
		Requester: "aura1unauth",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, st.Code())
}

func TestMsgServerRejectIdentityChange_PermissionDenied(t *testing.T) {
	keeper, ctx := setupKeeperForTest(t)
	msgServer := NewMsgServerImpl(keeper)
	require.NoError(t, keeper.SetParams(ctx, types.DefaultParams()))

	_, err := msgServer.RejectIdentityChange(ctx, &identitypb.MsgRejectIdentityChange{
		RequestId: "req-unauthorized",
		Actor:     "aura1unauth",
		Reason:    "not allowed",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, st.Code())
}
