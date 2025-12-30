// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package security

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/security/keeper"
	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// =============================================================================
// Test Suite Setup
// =============================================================================

type InternalFunctionsTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper *keeper.Keeper
	module AppModule
	cdc    codec.Codec
}

func TestInternalFunctionsTestSuite(t *testing.T) {
	suite.Run(t, new(InternalFunctionsTestSuite))
}

func (suite *InternalFunctionsTestSuite) SetupTest() {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)
	db := dbm.NewMemDB()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	binaryCodec := codec.NewProtoCodec(interfaceRegistry)

	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memKey, storetypes.StoreTypeMemory, nil)
	suite.Require().NoError(stateStore.LoadLatestVersion())

	// Set up context with block time
	baseTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	suite.ctx = sdk.NewContext(stateStore, cmtproto.Header{Time: baseTime, Height: 100}, false, log.NewNopLogger())

	suite.keeper = keeper.NewKeeper(binaryCodec, storeKey, memKey, "authority", nil, nil, nil)
	suite.module = NewAppModule(binaryCodec, suite.keeper)
	suite.cdc = binaryCodec

	// Initialize default params
	params := types.DefaultParams()
	suite.keeper.SetParams(suite.ctx, params)
}

// =============================================================================
// processKeyRotations Tests
// =============================================================================

func (suite *InternalFunctionsTestSuite) TestProcessKeyRotations_NoDueSchedules() {
	ctx := suite.ctx

	// No schedules exist - should complete without error
	err := suite.module.processKeyRotations(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestProcessKeyRotations_WithDueSchedules() {
	ctx := suite.ctx
	k := suite.keeper

	pastTime := ctx.BlockTime().Add(-1 * time.Hour)

	// Create a schedule that is due
	schedule := &securitypb.KeyRotationSchedule{
		Id:                      "due-schedule-1",
		KeyId:                   "key-1",
		RotationIntervalSeconds: 3600,
		NextRotationTime:        pastTime,
		Enabled:                 true,
	}
	k.SetKeyRotationSchedule(ctx, schedule)

	err := suite.module.processKeyRotations(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestProcessKeyRotations_WithFutureSchedules() {
	ctx := suite.ctx
	k := suite.keeper

	futureTime := ctx.BlockTime().Add(1 * time.Hour)

	// Create a schedule that is not due yet
	schedule := &securitypb.KeyRotationSchedule{
		Id:                      "future-schedule",
		KeyId:                   "key-future",
		RotationIntervalSeconds: 3600,
		NextRotationTime:        futureTime,
		Enabled:                 true,
	}
	k.SetKeyRotationSchedule(ctx, schedule)

	err := suite.module.processKeyRotations(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestProcessKeyRotations_DisabledSchedules() {
	ctx := suite.ctx
	k := suite.keeper

	pastTime := ctx.BlockTime().Add(-1 * time.Hour)

	// Create a disabled schedule that is past due
	schedule := &securitypb.KeyRotationSchedule{
		Id:                      "disabled-schedule",
		KeyId:                   "key-disabled",
		RotationIntervalSeconds: 3600,
		NextRotationTime:        pastTime,
		Enabled:                 false, // Disabled
	}
	k.SetKeyRotationSchedule(ctx, schedule)

	err := suite.module.processKeyRotations(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestProcessKeyRotations_MultipleSchedules() {
	ctx := suite.ctx
	k := suite.keeper

	pastTime := ctx.BlockTime().Add(-1 * time.Hour)
	futureTime := ctx.BlockTime().Add(1 * time.Hour)

	// Create mixed schedules
	schedules := []*securitypb.KeyRotationSchedule{
		{Id: "due-1", KeyId: "k1", NextRotationTime: pastTime, Enabled: true},
		{Id: "due-2", KeyId: "k2", NextRotationTime: pastTime, Enabled: true},
		{Id: "future-1", KeyId: "k3", NextRotationTime: futureTime, Enabled: true},
		{Id: "disabled-1", KeyId: "k4", NextRotationTime: pastTime, Enabled: false},
	}

	for _, s := range schedules {
		k.SetKeyRotationSchedule(ctx, s)
	}

	err := suite.module.processKeyRotations(ctx)
	suite.Require().NoError(err)
}

// =============================================================================
// updateNetworkMetrics Tests
// =============================================================================

func (suite *InternalFunctionsTestSuite) TestUpdateNetworkMetrics_NoRateLimits() {
	ctx := suite.ctx

	// No rate limits exist
	err := suite.module.updateNetworkMetrics(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestUpdateNetworkMetrics_ExpiredBan() {
	ctx := suite.ctx
	k := suite.keeper

	// Create a rate limit with expired ban
	expiredTime := ctx.BlockTime().Add(-1 * time.Hour)
	rl := &securitypb.RateLimitEntry{
		PeerId:       "peer-expired-ban",
		RequestCount: 100,
		IsBanned:     true,
		BanExpiresAt: &expiredTime,
	}
	k.SetRateLimit(ctx, rl)

	err := suite.module.updateNetworkMetrics(ctx)
	suite.Require().NoError(err)

	// Verify ban was lifted
	updatedRL, found := k.GetRateLimit(ctx, "peer-expired-ban")
	suite.Require().True(found)
	suite.Require().False(updatedRL.IsBanned)
}

func (suite *InternalFunctionsTestSuite) TestUpdateNetworkMetrics_ActiveBan() {
	ctx := suite.ctx
	k := suite.keeper

	// Create a rate limit with active ban
	futureTime := ctx.BlockTime().Add(1 * time.Hour)
	rl := &securitypb.RateLimitEntry{
		PeerId:       "peer-active-ban",
		RequestCount: 100,
		IsBanned:     true,
		BanExpiresAt: &futureTime,
	}
	k.SetRateLimit(ctx, rl)

	err := suite.module.updateNetworkMetrics(ctx)
	suite.Require().NoError(err)

	// Verify ban is still active
	updatedRL, found := k.GetRateLimit(ctx, "peer-active-ban")
	suite.Require().True(found)
	suite.Require().True(updatedRL.IsBanned)
}

func (suite *InternalFunctionsTestSuite) TestUpdateNetworkMetrics_MultiplePeers() {
	ctx := suite.ctx
	k := suite.keeper

	expiredTime := ctx.BlockTime().Add(-1 * time.Hour)
	futureTime := ctx.BlockTime().Add(1 * time.Hour)

	// Create multiple rate limits with different states
	rateLimits := []*securitypb.RateLimitEntry{
		{PeerId: "peer-1", IsBanned: true, BanExpiresAt: &expiredTime},  // Should be unbanned
		{PeerId: "peer-2", IsBanned: true, BanExpiresAt: &futureTime},  // Should stay banned
		{PeerId: "peer-3", IsBanned: false},                            // Not banned
		{PeerId: "peer-4", IsBanned: true, BanExpiresAt: nil},          // Banned with no expiry
	}

	for _, rl := range rateLimits {
		k.SetRateLimit(ctx, rl)
	}

	err := suite.module.updateNetworkMetrics(ctx)
	suite.Require().NoError(err)

	// Verify peer-1 was unbanned
	rl1, _ := k.GetRateLimit(ctx, "peer-1")
	suite.Require().False(rl1.IsBanned)

	// Verify peer-2 is still banned
	rl2, _ := k.GetRateLimit(ctx, "peer-2")
	suite.Require().True(rl2.IsBanned)
}

func (suite *InternalFunctionsTestSuite) TestUpdateNetworkMetrics_BanWithNilExpiry() {
	ctx := suite.ctx
	k := suite.keeper

	// Create a rate limit with ban but nil expiry
	rl := &securitypb.RateLimitEntry{
		PeerId:       "peer-nil-expiry",
		RequestCount: 100,
		IsBanned:     true,
		BanExpiresAt: nil, // No expiry set
	}
	k.SetRateLimit(ctx, rl)

	err := suite.module.updateNetworkMetrics(ctx)
	suite.Require().NoError(err)

	// Verify ban is still active (no automatic unban without expiry)
	updatedRL, found := k.GetRateLimit(ctx, "peer-nil-expiry")
	suite.Require().True(found)
	suite.Require().True(updatedRL.IsBanned)
}

// =============================================================================
// checkValidatorSecurity Tests
// =============================================================================

func (suite *InternalFunctionsTestSuite) TestCheckValidatorSecurity_NoAlerts() {
	ctx := suite.ctx

	// No alerts exist
	err := suite.module.checkValidatorSecurity(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestCheckValidatorSecurity_WithAlerts() {
	ctx := suite.ctx
	k := suite.keeper

	// Create validator alerts
	alert := &securitypb.ValidatorAlert{
		Id:               "alert-1",
		ValidatorAddress: "auravaloper1test",
		AlertType:        securitypb.ValidatorAlert_DOWNTIME,
		Severity:         securitypb.ValidatorAlert_CRITICAL,
	}
	k.SetValidatorAlert(ctx, alert)

	err := suite.module.checkValidatorSecurity(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestCheckValidatorSecurity_WithDoubleSignEvidence() {
	ctx := suite.ctx
	k := suite.keeper

	// Create double-sign evidence
	evidence := &securitypb.DoubleSignEvidence{
		ValidatorAddress: "auravaloper1test",
		Height:           100,
	}
	k.SetDoubleSignEvidence(ctx, evidence)

	err := suite.module.checkValidatorSecurity(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestCheckValidatorSecurity_MultipleAlertsAndEvidence() {
	ctx := suite.ctx
	k := suite.keeper

	// Create multiple alerts
	for i := 1; i <= 5; i++ {
		alert := &securitypb.ValidatorAlert{
			Id:               "alert-" + string(rune('0'+i)),
			ValidatorAddress: "auravaloper" + string(rune('0'+i)),
			AlertType:        securitypb.ValidatorAlert_DOWNTIME,
		}
		k.SetValidatorAlert(ctx, alert)
	}

	// Create multiple double-sign evidence
	for i := 1; i <= 3; i++ {
		evidence := &securitypb.DoubleSignEvidence{
			ValidatorAddress: "auravaloper" + string(rune('0'+i)),
			Height:           int64(100 + i),
		}
		k.SetDoubleSignEvidence(ctx, evidence)
	}

	err := suite.module.checkValidatorSecurity(ctx)
	suite.Require().NoError(err)
}

// =============================================================================
// processWalletSecurity Tests
// =============================================================================

func (suite *InternalFunctionsTestSuite) TestProcessWalletSecurity_NoPendingTxs() {
	ctx := suite.ctx

	// No pending transactions
	err := suite.module.processWalletSecurity(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestProcessWalletSecurity_ExpiredPendingTx() {
	ctx := suite.ctx
	k := suite.keeper

	expiredTime := ctx.BlockTime().Add(-1 * time.Hour)

	// Create expired pending transaction
	tx := &securitypb.PendingMultiSigTransaction{
		TxId:      "tx-expired",
		WalletId:  "wallet-1",
		ExpiresAt: &expiredTime,
	}
	k.SetPendingMultiSigTx(ctx, tx)

	err := suite.module.processWalletSecurity(ctx)
	suite.Require().NoError(err)

	// Verify transaction was deleted
	_, found := k.GetPendingMultiSigTx(ctx, "tx-expired")
	suite.Require().False(found)
}

func (suite *InternalFunctionsTestSuite) TestProcessWalletSecurity_ActivePendingTx() {
	ctx := suite.ctx
	k := suite.keeper

	futureTime := ctx.BlockTime().Add(1 * time.Hour)

	// Create active pending transaction
	tx := &securitypb.PendingMultiSigTransaction{
		TxId:      "tx-active",
		WalletId:  "wallet-2",
		ExpiresAt: &futureTime,
	}
	k.SetPendingMultiSigTx(ctx, tx)

	err := suite.module.processWalletSecurity(ctx)
	suite.Require().NoError(err)

	// Verify transaction still exists
	_, found := k.GetPendingMultiSigTx(ctx, "tx-active")
	suite.Require().True(found)
}

func (suite *InternalFunctionsTestSuite) TestProcessWalletSecurity_PendingTxNoExpiry() {
	ctx := suite.ctx
	k := suite.keeper

	// Create pending transaction with no expiry
	tx := &securitypb.PendingMultiSigTransaction{
		TxId:      "tx-no-expiry",
		WalletId:  "wallet-3",
		ExpiresAt: nil,
	}
	k.SetPendingMultiSigTx(ctx, tx)

	err := suite.module.processWalletSecurity(ctx)
	suite.Require().NoError(err)

	// Verify transaction still exists
	_, found := k.GetPendingMultiSigTx(ctx, "tx-no-expiry")
	suite.Require().True(found)
}

func (suite *InternalFunctionsTestSuite) TestProcessWalletSecurity_MultiplePendingTxs() {
	ctx := suite.ctx
	k := suite.keeper

	expiredTime := ctx.BlockTime().Add(-1 * time.Hour)
	futureTime := ctx.BlockTime().Add(1 * time.Hour)

	txs := []*securitypb.PendingMultiSigTransaction{
		{TxId: "tx-1", WalletId: "w1", ExpiresAt: &expiredTime},  // Should be deleted
		{TxId: "tx-2", WalletId: "w2", ExpiresAt: &futureTime},   // Should remain
		{TxId: "tx-3", WalletId: "w3", ExpiresAt: nil},           // Should remain
		{TxId: "tx-4", WalletId: "w4", ExpiresAt: &expiredTime},  // Should be deleted
	}

	for _, tx := range txs {
		k.SetPendingMultiSigTx(ctx, tx)
	}

	err := suite.module.processWalletSecurity(ctx)
	suite.Require().NoError(err)

	// Verify correct transactions were deleted
	_, found := k.GetPendingMultiSigTx(ctx, "tx-1")
	suite.Require().False(found, "expired tx-1 should be deleted")

	_, found = k.GetPendingMultiSigTx(ctx, "tx-2")
	suite.Require().True(found, "active tx-2 should remain")

	_, found = k.GetPendingMultiSigTx(ctx, "tx-3")
	suite.Require().True(found, "tx-3 with no expiry should remain")

	_, found = k.GetPendingMultiSigTx(ctx, "tx-4")
	suite.Require().False(found, "expired tx-4 should be deleted")
}

func (suite *InternalFunctionsTestSuite) TestProcessWalletSecurity_WithRecoveryRequests() {
	ctx := suite.ctx
	k := suite.keeper

	// Create recovery requests
	request := &securitypb.RecoveryRequest{
		RequestId: "recovery-1",
		WalletId:  "wallet-1",
		Status:    securitypb.RecoveryStatus_RECOVERY_STATUS_PENDING,
	}
	k.SetRecoveryRequest(ctx, request)

	err := suite.module.processWalletSecurity(ctx)
	suite.Require().NoError(err)
}

// =============================================================================
// updateIncidentState Tests
// =============================================================================

func (suite *InternalFunctionsTestSuite) TestUpdateIncidentState_NoIncidents() {
	ctx := suite.ctx

	// No incidents exist
	err := suite.module.updateIncidentState(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestUpdateIncidentState_WithIncidents() {
	ctx := suite.ctx
	k := suite.keeper

	// Create incident
	incident := &securitypb.Incident{
		IncidentId:  "INC-1",
		Title:       "Test Incident",
		Status:      securitypb.IncidentStatus_INCIDENT_STATUS_DETECTED,
		DetectedAt:  ctx.BlockTime(),
	}
	k.SetIncident(ctx, incident)

	err := suite.module.updateIncidentState(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestUpdateIncidentState_SystemNotPaused() {
	ctx := suite.ctx

	// System is not paused
	err := suite.module.updateIncidentState(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestUpdateIncidentState_SystemPausedWithin24Hours() {
	ctx := suite.ctx
	k := suite.keeper

	// Pause system less than 24 hours ago
	pausedAt := ctx.BlockTime().Add(-12 * time.Hour)
	pauseState := &types.PauseState{
		IsPaused:   true,
		PauseLevel: 1,
		PausedAt:   &pausedAt,
		PausedBy:   "admin",
		Reason:     "security incident",
	}
	k.SetPauseState(ctx, pauseState)

	err := suite.module.updateIncidentState(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestUpdateIncidentState_SystemPausedOver24Hours() {
	ctx := suite.ctx
	k := suite.keeper

	// Pause system more than 24 hours ago
	pausedAt := ctx.BlockTime().Add(-36 * time.Hour)
	pauseState := &types.PauseState{
		IsPaused:   true,
		PauseLevel: 2,
		PausedAt:   &pausedAt,
		PausedBy:   "admin",
		Reason:     "critical security incident",
	}
	k.SetPauseState(ctx, pauseState)

	err := suite.module.updateIncidentState(ctx)
	suite.Require().NoError(err)
	// The function logs a warning but does not error
}

func (suite *InternalFunctionsTestSuite) TestUpdateIncidentState_PausedWithNilPausedAt() {
	ctx := suite.ctx
	k := suite.keeper

	// Pause state with nil PausedAt
	pauseState := &types.PauseState{
		IsPaused:   true,
		PauseLevel: 1,
		PausedAt:   nil,
		PausedBy:   "admin",
		Reason:     "test",
	}
	k.SetPauseState(ctx, pauseState)

	err := suite.module.updateIncidentState(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestUpdateIncidentState_MultipleIncidents() {
	ctx := suite.ctx
	k := suite.keeper

	// Create multiple incidents
	for i := 1; i <= 5; i++ {
		incident := &securitypb.Incident{
			IncidentId:  "INC-" + string(rune('0'+i)),
			Title:       "Incident " + string(rune('0'+i)),
			Status:      securitypb.IncidentStatus_INCIDENT_STATUS_DETECTED,
			DetectedAt:  ctx.BlockTime(),
		}
		k.SetIncident(ctx, incident)
	}

	err := suite.module.updateIncidentState(ctx)
	suite.Require().NoError(err)
}

// =============================================================================
// refreshPrivacyPools Tests
// =============================================================================

func (suite *InternalFunctionsTestSuite) TestRefreshPrivacyPools_NoPools() {
	ctx := suite.ctx

	// No pools exist
	err := suite.module.refreshPrivacyPools(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestRefreshPrivacyPools_WithActivePools() {
	ctx := suite.ctx
	k := suite.keeper

	// Create active pool
	pool := &securitypb.MixingPool{
		PoolId:          "pool-1",
		MinParticipants: 5,
		MaxParticipants: 10,
		Status:          "active",
	}
	k.SetMixingPool(ctx, pool)

	err := suite.module.refreshPrivacyPools(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestRefreshPrivacyPools_MultiplePools() {
	ctx := suite.ctx
	k := suite.keeper

	// Create multiple pools
	pools := []*securitypb.MixingPool{
		{PoolId: "pool-1", Status: "active", MinParticipants: 5},
		{PoolId: "pool-2", Status: "inactive", MinParticipants: 10},
		{PoolId: "pool-3", Status: "pending", MinParticipants: 3},
	}

	for _, p := range pools {
		k.SetMixingPool(ctx, p)
	}

	err := suite.module.refreshPrivacyPools(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestRefreshPrivacyPools_PoolWithParticipants() {
	ctx := suite.ctx
	k := suite.keeper

	// Create pool with participants
	pool := &securitypb.MixingPool{
		PoolId:          "pool-participants",
		MinParticipants: 3,
		MaxParticipants: 10,
		Status:          "active",
		Participants:    [][]byte{[]byte("p1"), []byte("p2"), []byte("p3")},
	}
	k.SetMixingPool(ctx, pool)

	err := suite.module.refreshPrivacyPools(ctx)
	suite.Require().NoError(err)
}

// =============================================================================
// cleanupExpiredSessions Tests
// =============================================================================

func (suite *InternalFunctionsTestSuite) TestCleanupExpiredSessions_NoSessions() {
	ctx := suite.ctx

	// No sessions exist
	err := suite.module.cleanupExpiredSessions(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestCleanupExpiredSessions_ExpiredSession() {
	ctx := suite.ctx
	k := suite.keeper

	expiredTime := ctx.BlockTime().Add(-1 * time.Hour)

	// Create expired session
	session := &types.WalletSession{
		Id:            "session-expired",
		WalletAddress: "aura1test",
		ExpiresAt:     &expiredTime,
	}
	k.SetSession(ctx, session)

	err := suite.module.cleanupExpiredSessions(ctx)
	suite.Require().NoError(err)

	// Verify session was deleted
	_, found := k.GetSession(ctx, "session-expired")
	suite.Require().False(found)
}

func (suite *InternalFunctionsTestSuite) TestCleanupExpiredSessions_ActiveSession() {
	ctx := suite.ctx
	k := suite.keeper

	futureTime := ctx.BlockTime().Add(1 * time.Hour)

	// Create active session
	session := &types.WalletSession{
		Id:            "session-active",
		WalletAddress: "aura1test",
		ExpiresAt:     &futureTime,
	}
	k.SetSession(ctx, session)

	err := suite.module.cleanupExpiredSessions(ctx)
	suite.Require().NoError(err)

	// Verify session still exists
	_, found := k.GetSession(ctx, "session-active")
	suite.Require().True(found)
}

func (suite *InternalFunctionsTestSuite) TestCleanupExpiredSessions_SessionWithNilExpiry() {
	ctx := suite.ctx
	k := suite.keeper

	// Create session with no expiry
	session := &types.WalletSession{
		Id:            "session-no-expiry",
		WalletAddress: "aura1test",
		ExpiresAt:     nil,
	}
	k.SetSession(ctx, session)

	err := suite.module.cleanupExpiredSessions(ctx)
	suite.Require().NoError(err)

	// Verify session still exists (no expiry means it doesn't expire)
	_, found := k.GetSession(ctx, "session-no-expiry")
	suite.Require().True(found)
}

func (suite *InternalFunctionsTestSuite) TestCleanupExpiredSessions_MultipleSessions() {
	ctx := suite.ctx
	k := suite.keeper

	expiredTime := ctx.BlockTime().Add(-1 * time.Hour)
	futureTime := ctx.BlockTime().Add(1 * time.Hour)
	now := ctx.BlockTime()

	sessions := []*types.WalletSession{
		{Id: "s1", WalletAddress: "w1", ExpiresAt: &expiredTime},  // Should be deleted
		{Id: "s2", WalletAddress: "w2", ExpiresAt: &futureTime},   // Should remain
		{Id: "s3", WalletAddress: "w3", ExpiresAt: nil},           // Should remain
		{Id: "s4", WalletAddress: "w4", ExpiresAt: &expiredTime},  // Should be deleted
		{Id: "s5", WalletAddress: "w5", ExpiresAt: &now},          // Edge case: expires at current time, should NOT be deleted
	}

	for _, s := range sessions {
		k.SetSession(ctx, s)
	}

	err := suite.module.cleanupExpiredSessions(ctx)
	suite.Require().NoError(err)

	// Verify correct sessions were deleted
	_, found := k.GetSession(ctx, "s1")
	suite.Require().False(found, "expired s1 should be deleted")

	_, found = k.GetSession(ctx, "s2")
	suite.Require().True(found, "active s2 should remain")

	_, found = k.GetSession(ctx, "s3")
	suite.Require().True(found, "s3 with no expiry should remain")

	_, found = k.GetSession(ctx, "s4")
	suite.Require().False(found, "expired s4 should be deleted")

	_, found = k.GetSession(ctx, "s5")
	suite.Require().True(found, "s5 at exact time should remain (Before returns false)")
}

func (suite *InternalFunctionsTestSuite) TestCleanupExpiredSessions_LargeBatch() {
	ctx := suite.ctx
	k := suite.keeper

	expiredTime := ctx.BlockTime().Add(-1 * time.Hour)
	futureTime := ctx.BlockTime().Add(1 * time.Hour)

	// Create many sessions
	for i := 0; i < 100; i++ {
		var expiresAt *time.Time
		if i%2 == 0 {
			expiresAt = &expiredTime // Even indices expire
		} else {
			expiresAt = &futureTime // Odd indices are active
		}
		session := &types.WalletSession{
			Id:            "batch-session-" + string(rune('0'+i/10)) + string(rune('0'+i%10)),
			WalletAddress: "aura1batch" + string(rune('0'+i)),
			ExpiresAt:     expiresAt,
		}
		k.SetSession(ctx, session)
	}

	err := suite.module.cleanupExpiredSessions(ctx)
	suite.Require().NoError(err)

	// Verify remaining sessions count
	remaining := k.GetAllSessions(ctx)
	suite.Require().Equal(50, len(remaining), "50 active sessions should remain")
}

// =============================================================================
// Integration Tests - BeginBlock and EndBlock
// =============================================================================

func (suite *InternalFunctionsTestSuite) TestBeginBlock_AllFunctionsExecute() {
	ctx := suite.ctx

	// Test that BeginBlock executes all internal functions without error
	err := suite.module.BeginBlock(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestEndBlock_AllFunctionsExecute() {
	ctx := suite.ctx

	// Test that EndBlock executes all internal functions without error
	err := suite.module.EndBlock(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestBeginBlock_WithFullState() {
	ctx := suite.ctx
	k := suite.keeper

	pastTime := ctx.BlockTime().Add(-1 * time.Hour)
	futureTime := ctx.BlockTime().Add(1 * time.Hour)

	// Set up key rotation schedule
	k.SetKeyRotationSchedule(ctx, &securitypb.KeyRotationSchedule{
		Id: "kr-1", KeyId: "k1", NextRotationTime: pastTime, Enabled: true,
	})

	// Set up rate limits
	k.SetRateLimit(ctx, &securitypb.RateLimitEntry{
		PeerId: "peer-1", IsBanned: true, BanExpiresAt: &pastTime,
	})

	// Set up validator alerts
	k.SetValidatorAlert(ctx, &securitypb.ValidatorAlert{
		Id: "alert-1", ValidatorAddress: "val1",
	})

	// Set up pending transactions
	k.SetPendingMultiSigTx(ctx, &securitypb.PendingMultiSigTransaction{
		TxId: "tx-1", WalletId: "w1", ExpiresAt: &pastTime,
	})

	// Set up incidents
	k.SetIncident(ctx, &securitypb.Incident{
		IncidentId: "INC-1", Status: securitypb.IncidentStatus_INCIDENT_STATUS_DETECTED,
	})

	// Set up mixing pools
	k.SetMixingPool(ctx, &securitypb.MixingPool{
		PoolId: "pool-1", Status: "active",
	})

	// Execute BeginBlock
	err := suite.module.BeginBlock(ctx)
	suite.Require().NoError(err)

	// Verify state changes
	rl, _ := k.GetRateLimit(ctx, "peer-1")
	suite.Require().False(rl.IsBanned, "expired ban should be lifted")

	_, found := k.GetPendingMultiSigTx(ctx, "tx-1")
	suite.Require().False(found, "expired transaction should be deleted")

	// Set up sessions for EndBlock test
	k.SetSession(ctx, &types.WalletSession{
		Id: "sess-1", WalletAddress: "w1", ExpiresAt: &pastTime,
	})
	k.SetSession(ctx, &types.WalletSession{
		Id: "sess-2", WalletAddress: "w2", ExpiresAt: &futureTime,
	})

	// Execute EndBlock
	err = suite.module.EndBlock(ctx)
	suite.Require().NoError(err)

	// Verify session cleanup
	_, found = k.GetSession(ctx, "sess-1")
	suite.Require().False(found, "expired session should be deleted")

	_, found = k.GetSession(ctx, "sess-2")
	suite.Require().True(found, "active session should remain")
}

// =============================================================================
// Edge Cases and Error Handling Tests
// =============================================================================

func (suite *InternalFunctionsTestSuite) TestProcessKeyRotations_ZeroTime() {
	ctx := suite.ctx
	k := suite.keeper

	zeroTime := time.Time{}

	// Create schedule with zero time
	schedule := &securitypb.KeyRotationSchedule{
		Id:               "zero-time-schedule",
		KeyId:            "key-zero",
		NextRotationTime: zeroTime,
		Enabled:          true,
	}
	k.SetKeyRotationSchedule(ctx, schedule)

	err := suite.module.processKeyRotations(ctx)
	suite.Require().NoError(err)
}

func (suite *InternalFunctionsTestSuite) TestUpdateNetworkMetrics_ExactExpiryTime() {
	ctx := suite.ctx
	k := suite.keeper

	// Create rate limit with ban expiring exactly at block time
	exactTime := ctx.BlockTime()
	rl := &securitypb.RateLimitEntry{
		PeerId:       "peer-exact-time",
		IsBanned:     true,
		BanExpiresAt: &exactTime,
	}
	k.SetRateLimit(ctx, rl)

	err := suite.module.updateNetworkMetrics(ctx)
	suite.Require().NoError(err)

	// With Before(), exact time should NOT trigger unban
	updatedRL, _ := k.GetRateLimit(ctx, "peer-exact-time")
	suite.Require().True(updatedRL.IsBanned, "exact time should not trigger unban (Before returns false)")
}

func (suite *InternalFunctionsTestSuite) TestCleanupExpiredSessions_ExactExpiryTime() {
	ctx := suite.ctx
	k := suite.keeper

	// Create session expiring exactly at block time
	exactTime := ctx.BlockTime()
	session := &types.WalletSession{
		Id:        "session-exact",
		ExpiresAt: &exactTime,
	}
	k.SetSession(ctx, session)

	err := suite.module.cleanupExpiredSessions(ctx)
	suite.Require().NoError(err)

	// With Before(), exact time should NOT delete session
	_, found := k.GetSession(ctx, "session-exact")
	suite.Require().True(found, "session at exact time should remain")
}

// =============================================================================
// Standalone Test Functions
// =============================================================================

// setupModuleWithTime creates a module with a proper block time set
func setupModuleWithTime(t *testing.T) (AppModule, *keeper.Keeper, sdk.Context, codec.JSONCodec) {
	module, k, ctx, cdc := setupModule(t)
	// Set a proper block time to avoid protobuf timestamp issues
	baseTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(baseTime).WithBlockHeight(100)
	return module, k, ctx, cdc
}

func TestProcessKeyRotations_Standalone(t *testing.T) {
	module, k, ctx, _ := setupModuleWithTime(t)

	// Add due schedule
	pastTime := ctx.BlockTime().Add(-1 * time.Hour)
	schedule := &securitypb.KeyRotationSchedule{
		Id:               "standalone-schedule",
		KeyId:            "standalone-key",
		NextRotationTime: pastTime,
		Enabled:          true,
	}
	k.SetKeyRotationSchedule(ctx, schedule)

	err := module.processKeyRotations(ctx)
	require.NoError(t, err)
}

func TestUpdateNetworkMetrics_Standalone(t *testing.T) {
	module, k, ctx, _ := setupModuleWithTime(t)

	// Add rate limit with expired ban
	expiredTime := ctx.BlockTime().Add(-1 * time.Hour)
	rl := &securitypb.RateLimitEntry{
		PeerId:       "standalone-peer",
		IsBanned:     true,
		BanExpiresAt: &expiredTime,
	}
	k.SetRateLimit(ctx, rl)

	err := module.updateNetworkMetrics(ctx)
	require.NoError(t, err)

	updated, _ := k.GetRateLimit(ctx, "standalone-peer")
	require.False(t, updated.IsBanned)
}

func TestCheckValidatorSecurity_Standalone(t *testing.T) {
	module, k, ctx, _ := setupModuleWithTime(t)

	alert := &securitypb.ValidatorAlert{
		Id:               "standalone-alert",
		ValidatorAddress: "val1",
	}
	k.SetValidatorAlert(ctx, alert)

	err := module.checkValidatorSecurity(ctx)
	require.NoError(t, err)
}

func TestProcessWalletSecurity_Standalone(t *testing.T) {
	module, k, ctx, _ := setupModuleWithTime(t)

	expiredTime := ctx.BlockTime().Add(-1 * time.Hour)
	tx := &securitypb.PendingMultiSigTransaction{
		TxId:      "standalone-tx",
		WalletId:  "w1",
		ExpiresAt: &expiredTime,
	}
	k.SetPendingMultiSigTx(ctx, tx)

	err := module.processWalletSecurity(ctx)
	require.NoError(t, err)

	_, found := k.GetPendingMultiSigTx(ctx, "standalone-tx")
	require.False(t, found)
}

func TestUpdateIncidentState_Standalone(t *testing.T) {
	module, k, ctx, _ := setupModuleWithTime(t)

	incident := &securitypb.Incident{
		IncidentId: "standalone-incident",
		Status:     securitypb.IncidentStatus_INCIDENT_STATUS_DETECTED,
	}
	k.SetIncident(ctx, incident)

	err := module.updateIncidentState(ctx)
	require.NoError(t, err)
}

func TestRefreshPrivacyPools_Standalone(t *testing.T) {
	module, k, ctx, _ := setupModuleWithTime(t)

	pool := &securitypb.MixingPool{
		PoolId: "standalone-pool",
		Status: "active",
	}
	k.SetMixingPool(ctx, pool)

	err := module.refreshPrivacyPools(ctx)
	require.NoError(t, err)
}

func TestCleanupExpiredSessions_Standalone(t *testing.T) {
	module, k, ctx, _ := setupModuleWithTime(t)

	expiredTime := ctx.BlockTime().Add(-1 * time.Hour)
	session := &types.WalletSession{
		Id:        "standalone-session",
		ExpiresAt: &expiredTime,
	}
	k.SetSession(ctx, session)

	err := module.cleanupExpiredSessions(ctx)
	require.NoError(t, err)

	_, found := k.GetSession(ctx, "standalone-session")
	require.False(t, found)
}
