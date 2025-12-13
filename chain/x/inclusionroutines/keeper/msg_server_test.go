package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

func TestMsgServerConstruction(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)
	server := NewMsgServer(keeper)

	require.NotNil(t, server)
	require.NotNil(t, ctx.Context())

	// Ensure interface compliance at runtime
	_, ok := server.(interface {
		inclusionroutinespb.MsgServer
	})
	require.True(t, ok)
	_ = context.Background()
}
