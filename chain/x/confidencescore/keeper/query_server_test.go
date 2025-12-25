// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

func TestQueryServerConstruction(t *testing.T) {
	ctx, k := setupConfKeeper(t)
	server := NewQueryServer(k)

	require.NotNil(t, server)
	require.NotNil(t, sdk.WrapSDKContext(ctx))

	_, ok := server.(interface {
		confidencescorepb.QueryServer
	})
	require.True(t, ok)
	_ = context.Background()
}
