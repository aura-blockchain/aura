// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

func TestSetGetSession(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Test GetSession - not found
	session, found := keeper.GetSession(ctx, "session1")
	require.False(t, found)
	require.Nil(t, session)

	// Set and get a session
	expiresAt := ctx.BlockTime().Add(1 * time.Hour)
	keeper.SetSession(ctx, &types.WalletSession{
		Id:            "session1",
		WalletAddress: "aura1wallet1",
		ExpiresAt:     &expiresAt,
	})

	session, found = keeper.GetSession(ctx, "session1")
	require.True(t, found)
	require.NotNil(t, session)
	require.Equal(t, "session1", session.Id)
	require.Equal(t, "aura1wallet1", session.WalletAddress)
}

func TestGetAllSessions(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Initially empty
	sessions := keeper.GetAllSessions(ctx)
	require.Empty(t, sessions)

	// Add some sessions
	keeper.SetSession(ctx, &types.WalletSession{Id: "session1", WalletAddress: "aura1wallet1"})
	keeper.SetSession(ctx, &types.WalletSession{Id: "session2", WalletAddress: "aura1wallet2"})
	keeper.SetSession(ctx, &types.WalletSession{Id: "session3", WalletAddress: "aura1wallet3"})

	sessions = keeper.GetAllSessions(ctx)
	require.Len(t, sessions, 3)
}

func TestDeleteSession(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Set a session
	keeper.SetSession(ctx, &types.WalletSession{Id: "session1", WalletAddress: "aura1wallet1"})

	// Verify it exists
	_, found := keeper.GetSession(ctx, "session1")
	require.True(t, found)

	// Delete it
	keeper.DeleteSession(ctx, "session1")

	// Verify it's gone
	_, found = keeper.GetSession(ctx, "session1")
	require.False(t, found)
}

func TestValidateSession(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Test session not found
	err := keeper.ValidateSession(ctx, "nonexistent")
	require.Error(t, err)
	require.Equal(t, types.ErrInvalidSession, err)

	// Test valid session
	expiresAt := ctx.BlockTime().Add(1 * time.Hour)
	keeper.SetSession(ctx, &types.WalletSession{
		Id:            "session1",
		WalletAddress: "aura1wallet1",
		ExpiresAt:     &expiresAt,
	})

	err = keeper.ValidateSession(ctx, "session1")
	require.NoError(t, err)

	// Test expired session - note: this requires advancing block time
	expiredTime := ctx.BlockTime().Add(-1 * time.Hour)
	keeper.SetSession(ctx, &types.WalletSession{
		Id:            "expired_session",
		WalletAddress: "aura1wallet2",
		ExpiresAt:     &expiredTime,
	})

	err = keeper.ValidateSession(ctx, "expired_session")
	require.Error(t, err)
}

func TestSetGetAnomalyDetection(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Initially empty
	anomalies := keeper.GetAllAnomalyDetections(ctx)
	require.Empty(t, anomalies)

	// Set an anomaly
	keeper.SetAnomalyDetection(ctx, &types.AnomalyDetection{
		Id:            "anomaly1",
		WalletAddress: "aura1wallet1",
		AnomalyType:   "unusual_activity",
	})

	anomalies = keeper.GetAllAnomalyDetections(ctx)
	require.Len(t, anomalies, 1)
	require.Equal(t, "anomaly1", anomalies[0].Id)
}

func TestGetAllDeviceFingerprints(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Initially empty
	fps := keeper.GetAllDeviceFingerprints(ctx)
	require.Empty(t, fps)

	// Set some fingerprints
	keeper.SetDeviceFingerprint(ctx, &types.DeviceFingerprint{
		Id:            "fp1",
		WalletAddress: "aura1wallet1",
	})
	keeper.SetDeviceFingerprint(ctx, &types.DeviceFingerprint{
		Id:            "fp2",
		WalletAddress: "aura1wallet2",
	})

	fps = keeper.GetAllDeviceFingerprints(ctx)
	require.Len(t, fps, 2)
}

func TestIsMultiSigWallet(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Not a multisig wallet
	isMultiSig := keeper.IsMultiSigWallet(ctx, "nonexistent")
	require.False(t, isMultiSig)

	// Set a multisig wallet
	keeper.SetMultiSigWallet(ctx, &securitypb.MultiSigWallet{
		WalletId:  "multisig1",
		Threshold: 2,
		Signers:   []string{"signer1", "signer2", "signer3"},
	})

	// Now it should be recognized
	isMultiSig = keeper.IsMultiSigWallet(ctx, "multisig1")
	require.True(t, isMultiSig)
}

func TestValidateWallet(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Validate should succeed for any address (minimal validation)
	err := keeper.ValidateWallet(ctx, "aura1test123")
	require.NoError(t, err)
}

func TestLeaveMixingPool(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	// Create a mixing pool with math.Int
	keeper.SetMixingPool(ctx, &securitypb.MixingPool{
		PoolId:       "pool1",
		Participants: [][]byte{[]byte("participant1"), []byte("participant2")},
	})

	// Leave the pool
	err := keeper.LeaveMixingPool(ctx, "pool1")
	require.NoError(t, err)

	// Verify pool still exists
	_, found := keeper.GetMixingPool(ctx, "pool1")
	require.True(t, found)
}

func TestLeaveMixingPool_PoolNotFound(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)

	err := keeper.LeaveMixingPool(ctx, "nonexistent")
	require.Error(t, err)
}
