// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

func TestNewQueryServerImpl(t *testing.T) {
	keeper, _ := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)
	require.NotNil(t, queryServer)
}

func TestQueryParams(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.Params(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.Params(ctx, &securitypb.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Params)
}

func TestQuerySecurityStatus(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.SecurityStatus(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.SecurityStatus(ctx, &securitypb.QuerySecurityStatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryPeerInfo(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.PeerInfo(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.PeerInfo(ctx, &securitypb.QueryPeerInfoRequest{PeerId: "peer1"})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryAllPeers(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.AllPeers(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.AllPeers(ctx, &securitypb.QueryAllPeersRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryTrustedPeers(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.TrustedPeers(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Add some trusted peers
	keeper.SetTrustedPeer(ctx, &securitypb.TrustedPeer{
		PeerId:  "trusted1",
		Address: "192.168.1.1:26656",
	})

	// Test valid request
	resp, err = queryServer.TrustedPeers(ctx, &securitypb.QueryTrustedPeersRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryPeerReputation(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.PeerReputation(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.PeerReputation(ctx, &securitypb.QueryPeerReputationRequest{PeerId: "peer1"})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryRateLimitStatus(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.RateLimitStatus(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.RateLimitStatus(ctx, &securitypb.QueryRateLimitStatusRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryMempoolStats(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.MempoolStats(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.MempoolStats(ctx, &securitypb.QueryMempoolStatsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryForkAlerts(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.ForkAlerts(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.ForkAlerts(ctx, &securitypb.QueryForkAlertsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryPartitionAlerts(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.PartitionAlerts(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.PartitionAlerts(ctx, &securitypb.QueryPartitionAlertsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryNetworkHealth(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.NetworkHealth(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.NetworkHealth(ctx, &securitypb.QueryNetworkHealthRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryValidatorSecurityInfo(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.ValidatorSecurityInfo(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.ValidatorSecurityInfo(ctx, &securitypb.QueryValidatorSecurityInfoRequest{
		ValidatorAddress: "auravaloper1test",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryAllValidatorSecurityInfo(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.AllValidatorSecurityInfo(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.AllValidatorSecurityInfo(ctx, &securitypb.QueryAllValidatorSecurityInfoRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryValidatorAlerts(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.ValidatorAlerts(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.ValidatorAlerts(ctx, &securitypb.QueryValidatorAlertsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryDoubleSignEvidences(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.DoubleSignEvidences(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.DoubleSignEvidences(ctx, &securitypb.QueryDoubleSignEvidencesRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryDowntimeInfractions(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.DowntimeInfractions(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.DowntimeInfractions(ctx, &securitypb.QueryDowntimeInfractionsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQuerySentryNodes(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.SentryNodes(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.SentryNodes(ctx, &securitypb.QuerySentryNodesRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryWalletSecurityInfo(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.WalletSecurityInfo(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.WalletSecurityInfo(ctx, &securitypb.QueryWalletSecurityInfoRequest{
		WalletId: "wallet1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryHardwareWalletConfig(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.HardwareWalletConfig(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.HardwareWalletConfig(ctx, &securitypb.QueryHardwareWalletConfigRequest{
		WalletId: "hw1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryMultiSigWallet(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.MultiSigWallet(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.MultiSigWallet(ctx, &securitypb.QueryMultiSigWalletRequest{
		WalletId: "multisig1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryPendingMultiSigTransactions(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.PendingMultiSigTransactions(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.PendingMultiSigTransactions(ctx, &securitypb.QueryPendingMultiSigTransactionsRequest{
		WalletId: "multisig1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQuerySocialRecoveryConfig(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.SocialRecoveryConfig(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.SocialRecoveryConfig(ctx, &securitypb.QuerySocialRecoveryConfigRequest{
		WalletId: "wallet1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryRecoveryRequests(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.RecoveryRequests(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.RecoveryRequests(ctx, &securitypb.QueryRecoveryRequestsRequest{
		WalletId: "wallet1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQuerySpendingLimits(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.SpendingLimits(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.SpendingLimits(ctx, &securitypb.QuerySpendingLimitsRequest{
		WalletId: "wallet1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQuerySimulateTransaction(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.SimulateTransaction(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.SimulateTransaction(ctx, &securitypb.QuerySimulateTransactionRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryIncident(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.Incident(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.Incident(ctx, &securitypb.QueryIncidentRequest{
		IncidentId: "incident1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryAllIncidents(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.AllIncidents(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.AllIncidents(ctx, &securitypb.QueryAllIncidentsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryAuditLog(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.AuditLog(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.AuditLog(ctx, &securitypb.QueryAuditLogRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryResponseActions(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.ResponseActions(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.ResponseActions(ctx, &securitypb.QueryResponseActionsRequest{
		IncidentId: "incident1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryKeyRotationSchedule(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.KeyRotationSchedule(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.KeyRotationSchedule(ctx, &securitypb.QueryKeyRotationScheduleRequest{
		Id: "schedule1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryAllKeyRotationSchedules(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.AllKeyRotationSchedules(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.AllKeyRotationSchedules(ctx, &securitypb.QueryAllKeyRotationSchedulesRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryThresholdScheme(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.ThresholdScheme(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.ThresholdScheme(ctx, &securitypb.QueryThresholdSchemeRequest{
		SchemeId: "scheme1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryVerifyZKProof(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.VerifyZKProof(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.VerifyZKProof(ctx, &securitypb.QueryVerifyZKProofRequest{
		ProofId: "proof1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryQuantumResistantKey(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.QuantumResistantKey(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.QuantumResistantKey(ctx, &securitypb.QueryQuantumResistantKeyRequest{
		KeyId: "key1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryMixingPool(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.MixingPool(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.MixingPool(ctx, &securitypb.QueryMixingPoolRequest{
		PoolId: "pool1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryAllMixingPools(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.AllMixingPools(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.AllMixingPools(ctx, &securitypb.QueryAllMixingPoolsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryStealthAddress(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.StealthAddress(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.StealthAddress(ctx, &securitypb.QueryStealthAddressRequest{
		Address: []byte("stealth1"),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestQueryVerifyRingSignature(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Test nil request
	resp, err := queryServer.VerifyRingSignature(ctx, nil)
	require.Error(t, err)
	require.Nil(t, resp)

	// Test valid request
	resp, err = queryServer.VerifyRingSignature(ctx, &securitypb.QueryVerifyRingSignatureRequest{
		KeyImage: []byte("keyimage1"),
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// Ensure queryServer implements the context.Context interface requirement
func TestQueryServerContextUsage(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	queryServer := NewQueryServerImpl(&keeper)

	// Use context.Context interface
	var goCtx context.Context = ctx

	resp, err := queryServer.Params(goCtx, &securitypb.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}
