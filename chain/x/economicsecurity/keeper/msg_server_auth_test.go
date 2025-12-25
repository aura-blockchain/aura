// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

func TestMsgServerUpdateParams_PermissionDenied(t *testing.T) {
	k, _ := setupKeeperForTest(t)
	msgServer := NewMsgServer(k)

	_, err := msgServer.UpdateParams(nil, &types.MsgUpdateParams{
		Authority: "aura1bad",
		Params:    types.DefaultParams(),
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, st.Code())
}

func TestMsgServerAdjustInflationRate_PermissionDenied(t *testing.T) {
	k, _ := setupKeeperForTest(t)
	msgServer := NewMsgServer(k)

	_, err := msgServer.AdjustInflationRate(nil, &types.MsgAdjustInflationRate{
		Authority: "aura1bad",
		Reason:    "test",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.PermissionDenied, st.Code())
}
