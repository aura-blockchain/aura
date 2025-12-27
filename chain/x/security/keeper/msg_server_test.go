// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// Test addresses - valid bech32 format (generated with proper checksums)
const (
	testAuthority     = "aura1v96hg6r0wf5hg72lta047h6lta047h6lxkle3a"
	testUserAddress   = "aura1w4ek2ujlta047h6lta047h6lta047h6l7xaskr"
	testUserAddress2  = "aura1w4ek2u3jta047h6lta047h6lta047h6ldahc99"
	testValidatorAddr = "aura1weskc6tyv96x7ujlta047h6lta047h6l3vv5yc"
	testGuardianAddr  = "aura1va6kzunyd9skuh6lta047h6lta047h6lc5q7g0"
	testReporterAddr  = "aura1wfjhqmmjw3jhyh6lta047h6lta047h6lx03q92"
)

// ========================
// NETWORK SECURITY MESSAGES
// ========================

func TestMsgServerAddTrustedPeer(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgAddTrustedPeer{
		Authority: testAuthority,
		PeerId:    "peer123",
		Address:   "192.168.1.1:26656",
	}

	resp, err := msgServer.AddTrustedPeer(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify peer was added
	peer, found := keeper.GetTrustedPeer(ctx, "peer123")
	require.True(t, found)
	require.Equal(t, "peer123", peer.PeerId)
}

func TestMsgServerRemoveTrustedPeer(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// First add a trusted peer
	keeper.SetTrustedPeer(ctx, &securitypb.TrustedPeer{
		PeerId:  "peer123",
		Address: "192.168.1.1:26656",
	})

	msg := &securitypb.MsgRemoveTrustedPeer{
		Authority: testAuthority,
		PeerId:    "peer123",
	}

	resp, err := msgServer.RemoveTrustedPeer(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify peer was removed
	_, found := keeper.GetTrustedPeer(ctx, "peer123")
	require.False(t, found)
}

func TestMsgServerBanPeer(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgBanPeer{
		Authority: testAuthority,
		PeerId:    "badpeer",
		Reason:    "malicious behavior",
	}

	resp, err := msgServer.BanPeer(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify peer is blacklisted
	require.True(t, keeper.IsBlacklisted(ctx, "badpeer"))
}

func TestMsgServerUnbanPeer(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// First ban the peer
	_, err := msgServer.BanPeer(ctx, &securitypb.MsgBanPeer{
		Authority: testAuthority,
		PeerId:    "peer123",
		Reason:    "test ban",
	})
	require.NoError(t, err)

	msg := &securitypb.MsgUnbanPeer{
		Authority: testAuthority,
		PeerId:    "peer123",
	}

	resp, err := msgServer.UnbanPeer(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify peer is no longer blacklisted
	require.False(t, keeper.IsBlacklisted(ctx, "peer123"))
}

func TestMsgServerUpdatePeerReputation(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgUpdatePeerReputation{
		Authority:       testAuthority,
		PeerId:          "peer123",
		ReputationDelta: 10,
		Reason:          "good behavior",
	}

	resp, err := msgServer.UpdatePeerReputation(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, int64(10), resp.NewReputation)
}

func TestMsgServerResolveForkAlert(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// First create a fork alert
	keeper.SetForkAlert(ctx, &securitypb.ForkAlert{
		AlertId:     "fork-1",
		BlockHeight: 100,
		Resolved:    false,
	})

	msg := &securitypb.MsgResolveForkAlert{
		Authority:         testAuthority,
		AlertId:           "fork-1",
		ResolutionDetails: "resolved via manual intervention",
	}

	resp, err := msgServer.ResolveForkAlert(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerResolvePartitionAlert(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// First create a partition alert
	keeper.SetPartitionAlert(ctx, &securitypb.PartitionAlert{
		AlertId:  "partition-1",
		Resolved: false,
	})

	msg := &securitypb.MsgResolvePartitionAlert{
		Authority: testAuthority,
		AlertId:   "partition-1",
	}

	resp, err := msgServer.ResolvePartitionAlert(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// ========================
// VALIDATOR SECURITY MESSAGES
// ========================

func TestMsgServerRegisterValidatorSecurity(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgRegisterValidatorSecurity{
		ValidatorAddress: testValidatorAddr,
		Region:           "us-east",
	}

	resp, err := msgServer.RegisterValidatorSecurity(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify registration
	info, found := keeper.GetValidatorSecurityInfo(ctx, testValidatorAddr)
	require.True(t, found)
	require.Equal(t, "us-east", info.Region)
}

func TestMsgServerUpdateValidatorSecurity(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// First register validator
	keeper.SetValidatorSecurityInfo(ctx, &securitypb.ValidatorSecurityInfo{
		ValidatorAddress: testValidatorAddr,
		Region:           "us-east",
	})

	msg := &securitypb.MsgUpdateValidatorSecurity{
		ValidatorAddress: testValidatorAddr,
		Region:           "us-west",
	}

	resp, err := msgServer.UpdateValidatorSecurity(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify update
	info, found := keeper.GetValidatorSecurityInfo(ctx, testValidatorAddr)
	require.True(t, found)
	require.Equal(t, "us-west", info.Region)
}

func TestMsgServerRegisterSentryNode(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgRegisterSentryNode{
		ValidatorAddress: testValidatorAddr,
		SentryAddress:    testUserAddress2,
		IpAddress:        "192.168.1.1",
		Port:             26656,
	}

	resp, err := msgServer.RegisterSentryNode(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerRemoveSentryNode(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// First register a sentry node
	keeper.SetSentryNode(ctx, &securitypb.SentryNodeInfo{
		ValidatorAddress: testValidatorAddr,
		Address:          testUserAddress2,
		IsActive:         true,
	})

	msg := &securitypb.MsgRemoveSentryNode{
		ValidatorAddress: testValidatorAddr,
		SentryAddress:    testUserAddress2,
	}

	resp, err := msgServer.RemoveSentryNode(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerReportDoubleSign(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgReportDoubleSign{
		Reporter:         testReporterAddr,
		ValidatorAddress: testValidatorAddr,
		Height:           100,
		VoteA:            []byte("vote a"),
		VoteB:            []byte("vote b"),
	}

	resp, err := msgServer.ReportDoubleSign(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerAcknowledgeValidatorAlert(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// First create a validator alert
	alertTime := ctx.BlockTime()
	keeper.SetValidatorAlert(ctx, &securitypb.ValidatorAlert{
		Id:               "alert-1",
		ValidatorAddress: testValidatorAddr,
		Timestamp:        &alertTime,
	})

	msg := &securitypb.MsgAcknowledgeValidatorAlert{
		ValidatorAddress: testValidatorAddr,
		AlertId:          "alert-1",
	}

	resp, err := msgServer.AcknowledgeValidatorAlert(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerTriggerFailover(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// First register validator security
	keeper.SetValidatorSecurityInfo(ctx, &securitypb.ValidatorSecurityInfo{
		ValidatorAddress: testValidatorAddr,
	})

	msg := &securitypb.MsgTriggerFailover{
		ValidatorAddress: testValidatorAddr,
		Reason:           "maintenance",
		BackupValidator:  testUserAddress2,
	}

	resp, err := msgServer.TriggerFailover(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// ========================
// WALLET SECURITY MESSAGES
// ========================

func TestMsgServerRegisterHardwareWallet(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgRegisterHardwareWallet{
		Address:  testUserAddress,
		DeviceId: "ledger-001",
	}

	resp, err := msgServer.RegisterHardwareWallet(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.WalletId)
}

func TestMsgServerCreateMultiSigWallet(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgCreateMultiSigWallet{
		Creator:   testUserAddress,
		Signers:   []string{testUserAddress, testUserAddress2, testGuardianAddr},
		Threshold: 2,
	}

	resp, err := msgServer.CreateMultiSigWallet(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.WalletId)
}

func TestMsgServerProposeMultiSigTransaction(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// First create a multisig wallet
	createdAt := ctx.BlockTime()
	walletId := "MS-testid"
	keeper.SetMultiSigWallet(ctx, &securitypb.MultiSigWallet{
		WalletId:     walletId,
		Signers:      []string{testUserAddress, testUserAddress2},
		Threshold:    2,
		TotalSigners: 2,
		CreatedAt:    &createdAt,
	})

	msg := &securitypb.MsgProposeMultiSigTransaction{
		Proposer: testUserAddress,
		WalletId: walletId,
		TxData:   []byte("test transaction"),
	}

	resp, err := msgServer.ProposeMultiSigTransaction(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.TxId)
}

func TestMsgServerSignMultiSigTransaction(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// Create wallet
	walletId := "MS-testid"
	createdAt := ctx.BlockTime()
	keeper.SetMultiSigWallet(ctx, &securitypb.MultiSigWallet{
		WalletId:     walletId,
		Signers:      []string{testUserAddress, testUserAddress2},
		Threshold:    2,
		TotalSigners: 2,
		CreatedAt:    &createdAt,
	})

	// Create pending tx
	txId := "TX-test123"
	keeper.SetPendingMultiSigTx(ctx, &securitypb.PendingMultiSigTransaction{
		TxId:     txId,
		WalletId: walletId,
		SignedBy: []string{testUserAddress},
	})

	msg := &securitypb.MsgSignMultiSigTransaction{
		Signer: testUserAddress2,
		TxId:   txId,
	}

	resp, err := msgServer.SignMultiSigTransaction(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerExecuteMultiSigTransaction(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// Create wallet
	walletId := "MS-testid"
	createdAt := ctx.BlockTime()
	keeper.SetMultiSigWallet(ctx, &securitypb.MultiSigWallet{
		WalletId:     walletId,
		Signers:      []string{testUserAddress, testUserAddress2},
		Threshold:    2,
		TotalSigners: 2,
		CreatedAt:    &createdAt,
	})

	// Create pending tx with enough signatures
	txId := "TX-test123"
	keeper.SetPendingMultiSigTx(ctx, &securitypb.PendingMultiSigTransaction{
		TxId:     txId,
		WalletId: walletId,
		SignedBy: []string{testUserAddress, testUserAddress2},
	})

	msg := &securitypb.MsgExecuteMultiSigTransaction{
		Executor: testUserAddress,
		TxId:     txId,
	}

	resp, err := msgServer.ExecuteMultiSigTransaction(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Success)
}

func TestMsgServerConfigureSocialRecovery(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgConfigureSocialRecovery{
		Address:  testUserAddress,
		WalletId: "wallet-1",
		Guardians: []*securitypb.Guardian{
			{Address: testGuardianAddr},
			{Address: testUserAddress2},
		},
		RecoveryThreshold: 2,
		RecoveryDelay:     time.Hour * 24,
	}

	resp, err := msgServer.ConfigureSocialRecovery(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerInitiateRecovery(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// First configure social recovery
	walletId := "wallet-1"
	configuredAt := ctx.BlockTime()
	keeper.SetSocialRecoveryConfig(ctx, &securitypb.SocialRecoveryConfig{
		WalletId: walletId,
		Guardians: []*securitypb.Guardian{
			{Address: testGuardianAddr},
			{Address: testUserAddress2},
		},
		RecoveryThreshold: 2,
		RecoveryDelay:     time.Hour,
		Enabled:           true,
		ConfiguredAt:      &configuredAt,
	})

	msg := &securitypb.MsgInitiateRecovery{
		Initiator:  testGuardianAddr,
		WalletId:   walletId,
		NewAddress: testReporterAddr,
	}

	resp, err := msgServer.InitiateRecovery(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.RequestId)
}

func TestMsgServerApproveRecovery(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// Set up recovery config and request
	walletId := "wallet-1"
	requestId := "REC-wallet-1-1"
	configuredAt := ctx.BlockTime()

	keeper.SetSocialRecoveryConfig(ctx, &securitypb.SocialRecoveryConfig{
		WalletId: walletId,
		Guardians: []*securitypb.Guardian{
			{Address: testGuardianAddr},
			{Address: testUserAddress2},
		},
		RecoveryThreshold: 2,
		ConfiguredAt:      &configuredAt,
	})

	initiatedAt := ctx.BlockTime()
	keeper.SetRecoveryRequest(ctx, &securitypb.RecoveryRequest{
		RequestId:      requestId,
		WalletId:       walletId,
		NewAddress:     testReporterAddr,
		Approvals:      []string{testGuardianAddr},
		ApprovalsCount: 1,
		InitiatedAt:    &initiatedAt,
		Status:         securitypb.RecoveryStatus_RECOVERY_STATUS_PENDING,
	})

	msg := &securitypb.MsgApproveRecovery{
		Guardian:  testUserAddress2,
		RequestId: requestId,
	}

	resp, err := msgServer.ApproveRecovery(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerExecuteRecovery(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// Set up complete recovery scenario
	walletId := "wallet-1"
	requestId := "REC-wallet-1-1"
	configuredAt := ctx.BlockTime()

	keeper.SetSocialRecoveryConfig(ctx, &securitypb.SocialRecoveryConfig{
		WalletId:          walletId,
		RecoveryThreshold: 2,
		Guardians: []*securitypb.Guardian{
			{Address: testGuardianAddr},
			{Address: testUserAddress2},
		},
		ConfiguredAt: &configuredAt,
	})

	initiatedAt := ctx.BlockTime().Add(-2 * time.Hour)
	executableAt := ctx.BlockTime().Add(-time.Hour)
	keeper.SetRecoveryRequest(ctx, &securitypb.RecoveryRequest{
		RequestId:      requestId,
		WalletId:       walletId,
		NewAddress:     testReporterAddr,
		Approvals:      []string{testGuardianAddr, testUserAddress2},
		ApprovalsCount: 2,
		InitiatedAt:    &initiatedAt,
		ExecutableAt:   &executableAt,
		Status:         securitypb.RecoveryStatus_RECOVERY_STATUS_PENDING,
	})

	msg := &securitypb.MsgExecuteRecovery{
		Executor:  testUserAddress,
		RequestId: requestId,
	}

	resp, err := msgServer.ExecuteRecovery(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Success)
}

func TestMsgServerSetSpendingLimits(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgSetSpendingLimits{
		Address:    testUserAddress,
		Denom:      "uaura",
		DailyLimit: "1000000",
	}

	resp, err := msgServer.SetSpendingLimits(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerRegisterBiometric(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgRegisterBiometric{
		Address:        testUserAddress,
		WalletId:       "wallet-1",
		Type:           securitypb.BiometricType_BIOMETRIC_TYPE_FINGERPRINT,
		EnrollmentHash: "fingerprint-hash-abc123",
	}

	resp, err := msgServer.RegisterBiometric(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// ========================
// INCIDENT RESPONSE MESSAGES
// ========================

func TestMsgServerCreateIncident(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgCreateIncident{
		Reporter:    testReporterAddr,
		Title:       "Security Incident",
		Description: "Detected unauthorized access attempt",
		Severity:    securitypb.IncidentSeverity_INCIDENT_SEVERITY_HIGH,
	}

	resp, err := msgServer.CreateIncident(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.IncidentId)
}

func TestMsgServerUpdateIncident(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// Create incident first
	incidentId := "INC-0001"
	detectedAt := ctx.BlockTime()
	keeper.SetIncident(ctx, &securitypb.Incident{
		IncidentId: incidentId,
		Title:      "Test Incident",
		Status:     securitypb.IncidentStatus_INCIDENT_STATUS_DETECTED,
		DetectedAt: detectedAt,
	})

	msg := &securitypb.MsgUpdateIncident{
		Updater:     testUserAddress,
		IncidentId:  incidentId,
		Status:      securitypb.IncidentStatus_INCIDENT_STATUS_INVESTIGATING,
		Description: "Investigating the issue",
	}

	resp, err := msgServer.UpdateIncident(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerResolveIncident(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// Create incident first
	incidentId := "INC-0001"
	detectedAt := ctx.BlockTime()
	keeper.SetIncident(ctx, &securitypb.Incident{
		IncidentId: incidentId,
		Title:      "Test Incident",
		Status:     securitypb.IncidentStatus_INCIDENT_STATUS_INVESTIGATING,
		DetectedAt: detectedAt,
	})

	msg := &securitypb.MsgResolveIncident{
		Resolver:          testUserAddress,
		IncidentId:        incidentId,
		ResolutionDetails: "Issue resolved by patching vulnerability",
	}

	resp, err := msgServer.ResolveIncident(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerExecuteResponseAction(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// Create incident first
	incidentId := "INC-0001"
	detectedAt := ctx.BlockTime()
	keeper.SetIncident(ctx, &securitypb.Incident{
		IncidentId: incidentId,
		Title:      "Test Incident",
		Status:     securitypb.IncidentStatus_INCIDENT_STATUS_INVESTIGATING,
		DetectedAt: detectedAt,
	})

	msg := &securitypb.MsgExecuteResponseAction{
		Executor:    testUserAddress,
		IncidentId:  incidentId,
		ActionType:  "notify",
		Description: "Send alerts to stakeholders",
	}

	resp, err := msgServer.ExecuteResponseAction(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Successful)
}

func TestMsgServerAddAuditLogEntry(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgAddAuditLogEntry{
		Actor:     testUserAddress,
		Action:    "key_rotation",
		Resource:  "master_key_001",
		EventType: "security",
		Success:   true,
	}

	resp, err := msgServer.AddAuditLogEntry(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.LogId)
}

// ========================
// CRYPTOGRAPHY MESSAGES
// ========================

func TestMsgServerCreateKeyRotationSchedule(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgCreateKeyRotationSchedule{
		Creator:                 testUserAddress,
		KeyId:                   "key123",
		RotationIntervalSeconds: 86400, // 24 hours
	}

	resp, err := msgServer.CreateKeyRotationSchedule(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "key123", resp.ScheduleId)
}

func TestMsgServerRotateKey(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// First create a key rotation schedule
	keeper.SetKeyRotationSchedule(ctx, &securitypb.KeyRotationSchedule{
		Id:      "key123",
		KeyId:   "key123",
		Enabled: true,
	})

	msg := &securitypb.MsgRotateKey{
		Creator: testUserAddress,
		KeyId:   "key123",
	}

	resp, err := msgServer.RotateKey(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.RotationId)
}

func TestMsgServerCreateThresholdScheme(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgCreateThresholdScheme{
		Creator:           testUserAddress,
		Threshold:         2,
		TotalParticipants: 3,
		ParticipantIds:    []string{testUserAddress, testUserAddress2, testGuardianAddr},
	}

	resp, err := msgServer.CreateThresholdScheme(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.SchemeId)
}

func TestMsgServerSubmitThresholdSignatureShare(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// Create threshold scheme first
	schemeId := "TSS-test123"
	keeper.SetThresholdScheme(ctx, &securitypb.ThresholdSignatureScheme{
		SchemeId:       schemeId,
		ParticipantIds: []string{testUserAddress, testUserAddress2},
		Threshold:      2,
	})

	msg := &securitypb.MsgSubmitThresholdSignatureShare{
		Submitter:      testUserAddress,
		SchemeId:       schemeId,
		SignatureShare: []byte("signature-share-data"),
	}

	resp, err := msgServer.SubmitThresholdSignatureShare(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerRegisterZKProofCircuit(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgRegisterZKProofCircuit{
		Creator:          testUserAddress,
		CircuitId:        "circuit123",
		ProofType:        securitypb.ZKProofType_ZK_PROOF_TYPE_GROTH16,
		VerificationKey:  []byte("vk-data"),
		PublicParameters: []byte("params"),
	}

	resp, err := msgServer.RegisterZKProofCircuit(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.ProofId)
}

func TestMsgServerSubmitZKProof(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// Create ZK proof config first
	proofId := "zkp-test123"
	keeper.SetZKProofConfig(ctx, &securitypb.ZKProofConfig{
		ProofId:   proofId,
		CircuitId: "circuit123",
	})

	msg := &securitypb.MsgSubmitZKProof{
		Submitter:    testUserAddress,
		ProofId:      proofId,
		ProofData:    []byte("proof data"),
		PublicInputs: []byte("public inputs"),
	}

	resp, err := msgServer.SubmitZKProof(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Verified)
}

func TestMsgServerGenerateQuantumResistantKey(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgGenerateQuantumResistantKey{
		Creator:   testUserAddress,
		Algorithm: securitypb.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM,
	}

	resp, err := msgServer.GenerateQuantumResistantKey(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.KeyId)
	require.NotEmpty(t, resp.PublicKey)
}

// ========================
// PRIVACY MESSAGES
// ========================

func TestMsgServerCreateMixingPool(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgCreateMixingPool{
		Creator:         testUserAddress,
		Denomination:    "uaura",
		MinParticipants: 3,
	}

	resp, err := msgServer.CreateMixingPool(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.PoolId)
}

func TestMsgServerJoinMixingPool(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// Create pool first with "active" status
	poolId := "POOL-test123"
	keeper.SetMixingPool(ctx, &securitypb.MixingPool{
		PoolId:          poolId,
		MinParticipants: 3,
		Status:          "active",
	})

	msg := &securitypb.MsgJoinMixingPool{
		Participant: testUserAddress,
		PoolId:      poolId,
	}

	resp, err := msgServer.JoinMixingPool(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgServerExecuteMixing(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	// Create pool with participants (as [][]byte)
	poolId := "POOL-test123"
	keeper.SetMixingPool(ctx, &securitypb.MixingPool{
		PoolId:          poolId,
		MinParticipants: 2,
		Participants:    [][]byte{[]byte(testUserAddress), []byte(testUserAddress2), []byte(testGuardianAddr)},
		Status:          "ready",
	})

	msg := &securitypb.MsgExecuteMixing{
		Executor: testUserAddress,
		PoolId:   poolId,
	}

	resp, err := msgServer.ExecuteMixing(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Success)
}

func TestMsgServerGenerateStealthAddress(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgGenerateStealthAddress{
		Creator:        testUserAddress,
		PublicSpendKey: []byte("public-spend-key"),
		PublicViewKey:  []byte("public-view-key"),
	}

	resp, err := msgServer.GenerateStealthAddress(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.OneTimeAddress)
}

func TestMsgServerCreateRingSignature(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgCreateRingSignature{
		Signer:      testUserAddress,
		Message:     []byte("message to sign"),
		RingMembers: [][]byte{[]byte("member1"), []byte("member2"), []byte("member3")},
	}

	resp, err := msgServer.CreateRingSignature(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.KeyImage)
}

func TestMsgServerCreateConfidentialTransaction(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgCreateConfidentialTransaction{
		Sender: testUserAddress,
	}

	resp, err := msgServer.CreateConfidentialTransaction(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.TxId)
}

// ========================
// PARAMS MESSAGE
// ========================

func TestMsgServerUpdateParams(t *testing.T) {
	keeper, ctx := newTestSecurityKeeper(t)
	msgServer := NewMsgServerImpl(&keeper)

	msg := &securitypb.MsgUpdateParams{
		Authority: testAuthority,
		Params:    securitypb.Params{},
	}

	resp, err := msgServer.UpdateParams(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}
