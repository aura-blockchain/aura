package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

func TestMsgServerConstruction(t *testing.T) {
	ctx, k := setupConfKeeper(t)
	server := NewMsgServer(k)

	require.NotNil(t, server)
	require.NotNil(t, sdk.WrapSDKContext(ctx))

	_, ok := server.(interface {
		confidencescorepb.MsgServer
	})
	require.True(t, ok)
	_ = context.Background()
}
