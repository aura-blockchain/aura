package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

func TestQueryServerConstruction(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)
	server := NewQueryServer(keeper)

	require.NotNil(t, server)
	require.NotNil(t, ctx.Context())

	_, ok := server.(interface {
		inclusionroutinespb.QueryServer
	})
	require.True(t, ok)
	_ = context.Background()
}
