// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

func TestValidatorSecurityInfo(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Not found initially
	info, found := keeper.GetValidatorSecurityInfo(ctx, "val1")
	require.False(t, found)
	require.Nil(t, info)

	// Set info
	keeper.SetValidatorSecurityInfo(ctx, &securitypb.ValidatorSecurityInfo{
		ValidatorAddress:    "val1",
		MissedBlocksCounter: 5,
	})

	info, found = keeper.GetValidatorSecurityInfo(ctx, "val1")
	require.True(t, found)
	require.Equal(t, int64(5), info.MissedBlocksCounter)

	// Get all
	all := keeper.GetAllValidatorSecurityInfos(ctx)
	require.Len(t, all, 1)
}

func TestTrackMissedBlock(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Track missed block for new validator
	keeper.TrackMissedBlock(ctx, "val1")
	info, _ := keeper.GetValidatorSecurityInfo(ctx, "val1")
	require.Equal(t, int64(1), info.MissedBlocksCounter)

	// Track another missed block
	keeper.TrackMissedBlock(ctx, "val1")
	info, _ = keeper.GetValidatorSecurityInfo(ctx, "val1")
	require.Equal(t, int64(2), info.MissedBlocksCounter)
}

func TestTrackSignedBlock(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Set initial missed blocks
	keeper.SetValidatorSecurityInfo(ctx, &securitypb.ValidatorSecurityInfo{
		ValidatorAddress:    "val1",
		MissedBlocksCounter: 10,
	})

	// Track signed block - resets counter
	keeper.TrackSignedBlock(ctx, "val1")
	info, _ := keeper.GetValidatorSecurityInfo(ctx, "val1")
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	// Track for new validator
	keeper.TrackSignedBlock(ctx, "val2")
	info, _ = keeper.GetValidatorSecurityInfo(ctx, "val2")
	require.Equal(t, int64(0), info.MissedBlocksCounter)
}

func TestDoubleSignEvidence(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	keeper.SetDoubleSignEvidence(ctx, &securitypb.DoubleSignEvidence{
		ValidatorAddress: "val1",
		Height:           100,
	})

	all := keeper.GetAllDoubleSignEvidence(ctx)
	require.Len(t, all, 1)
	require.Equal(t, "val1", all[0].ValidatorAddress)
	require.Equal(t, int64(100), all[0].Height)
}

func TestDowntimeInfraction(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	keeper.SetDowntimeInfraction(ctx, &securitypb.DowntimeInfraction{
		ValidatorAddress: "val1",
		MissedBlocks:     50,
	})

	all := keeper.GetAllDowntimeInfractions(ctx)
	require.Len(t, all, 1)
	require.Equal(t, int64(50), all[0].MissedBlocks)
}

func TestValidatorAlert(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	keeper.SetValidatorAlert(ctx, &securitypb.ValidatorAlert{
		Id:      "alert-1",
		Message: "Low uptime",
	})

	all := keeper.GetAllValidatorAlerts(ctx)
	require.Len(t, all, 1)
	require.Equal(t, "alert-1", all[0].Id)
}

func TestSentryNode(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	keeper.SetSentryNode(ctx, &securitypb.SentryNodeInfo{
		ValidatorAddress: "val1",
		Address:          "sentry1",
		IsActive:         true,
	})

	all := keeper.GetAllSentryNodes(ctx)
	require.Len(t, all, 1)
	require.True(t, all[0].IsActive)
}
