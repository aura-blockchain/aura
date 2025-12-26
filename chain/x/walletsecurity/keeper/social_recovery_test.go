// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/suite"

	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

type SocialRecoveryTestSuite struct {
	KeeperTestSuite
}

func TestSocialRecoveryTestSuite(t *testing.T) {
	suite.Run(t, new(SocialRecoveryTestSuite))
}

func (suite *SocialRecoveryTestSuite) TestConfigureSocialRecovery() {
	tests := []struct {
		name              string
		walletID          string
		guardians         []*wsproto.Guardian
		recoveryThreshold int32
		recoveryDelay     *gogotypes.Duration
		wantErr           bool
		errContains       string
	}{
		{
			name:     "valid configuration with 3 guardians",
			walletID: "wallet_test_1",
			guardians: []*wsproto.Guardian{
				{Address: "aura1guardian1"},
				{Address: "aura1guardian2"},
				{Address: "aura1guardian3"},
			},
			recoveryThreshold: 2,
			wantErr:           false,
		},
		{
			name:     "minimum 2 guardians with threshold 2",
			walletID: "wallet_test_2",
			guardians: []*wsproto.Guardian{
				{Address: "aura1guardian1"},
				{Address: "aura1guardian2"},
			},
			recoveryThreshold: 2,
			wantErr:           false,
		},
		{
			name:              "empty guardians fails",
			walletID:          "wallet_test_3",
			guardians:         []*wsproto.Guardian{},
			recoveryThreshold: 2,
			wantErr:           true,
		},
		{
			name:     "threshold below minimum fails",
			walletID: "wallet_test_4",
			guardians: []*wsproto.Guardian{
				{Address: "aura1guardian1"},
				{Address: "aura1guardian2"},
			},
			recoveryThreshold: 1, // Minimum is 2
			wantErr:           true,
		},
		{
			name:     "threshold exceeds guardians fails",
			walletID: "wallet_test_5",
			guardians: []*wsproto.Guardian{
				{Address: "aura1guardian1"},
				{Address: "aura1guardian2"},
			},
			recoveryThreshold: 5,
			wantErr:           true,
		},
		{
			name:     "empty guardian address fails",
			walletID: "wallet_test_6",
			guardians: []*wsproto.Guardian{
				{Address: "aura1guardian1"},
				{Address: ""}, // Empty address
			},
			recoveryThreshold: 2,
			wantErr:           true,
		},
		{
			name:     "duplicate guardian fails",
			walletID: "wallet_test_7",
			guardians: []*wsproto.Guardian{
				{Address: "aura1guardian1"},
				{Address: "aura1guardian1"}, // Duplicate
			},
			recoveryThreshold: 2,
			wantErr:           true,
		},
		{
			name:     "with custom recovery delay",
			walletID: "wallet_test_8",
			guardians: []*wsproto.Guardian{
				{Address: "aura1guardian1"},
				{Address: "aura1guardian2"},
			},
			recoveryThreshold: 2,
			recoveryDelay:     &gogotypes.Duration{Seconds: 86400}, // 24 hours
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()
			k := suite.GetKeeper()
			ctx := suite.GetContext()

			config, err := k.ConfigureSocialRecovery(
				ctx,
				tt.walletID,
				tt.guardians,
				tt.recoveryThreshold,
				tt.recoveryDelay,
			)

			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errContains != "" {
					suite.Require().Contains(err.Error(), tt.errContains)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(config)
				suite.Require().Equal(tt.walletID, config.WalletId)
				suite.Require().Len(config.Guardians, len(tt.guardians))
				suite.Require().Equal(tt.recoveryThreshold, config.RecoveryThreshold)
				suite.Require().True(config.Enabled)

				// Verify all guardians are initially unconfirmed
				for _, g := range config.Guardians {
					suite.Require().False(g.Confirmed)
					suite.Require().Equal(int32(0), g.RecoveryRequestsCount)
				}
			}
		})
	}
}

func (suite *SocialRecoveryTestSuite) TestConfirmGuardian() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Configure recovery
	walletID := "wallet_confirm_test"
	guardians := []*wsproto.Guardian{
		{Address: "aura1guardian1"},
		{Address: "aura1guardian2"},
	}
	_, err := k.ConfigureSocialRecovery(ctx, walletID, guardians, 2, nil)
	suite.Require().NoError(err)

	tests := []struct {
		name            string
		guardianAddress string
		wantErr         bool
	}{
		{
			name:            "confirm existing guardian",
			guardianAddress: "aura1guardian1",
			wantErr:         false,
		},
		{
			name:            "confirm non-existent guardian fails",
			guardianAddress: "aura1unknown",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := k.ConfirmGuardian(ctx, walletID, tt.guardianAddress)

			if tt.wantErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *SocialRecoveryTestSuite) TestInitiateRecovery() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	walletID := "wallet_recovery_test"
	guardians := []*wsproto.Guardian{
		{Address: "aura1guardian1"},
		{Address: "aura1guardian2"},
	}
	_, err := k.ConfigureSocialRecovery(ctx, walletID, guardians, 2, nil)
	suite.Require().NoError(err)

	// Confirm guardians
	k.ConfirmGuardian(ctx, walletID, "aura1guardian1")
	k.ConfirmGuardian(ctx, walletID, "aura1guardian2")

	tests := []struct {
		name       string
		newAddress string
		initiator  string
		wantErr    bool
	}{
		{
			name:       "confirmed guardian initiates recovery",
			newAddress: "aura1newowner",
			initiator:  "aura1guardian1",
			wantErr:    false,
		},
		{
			name:       "unconfirmed guardian cannot initiate",
			newAddress: "aura1newowner2",
			initiator:  "aura1unknown",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			request, err := k.InitiateRecovery(ctx, walletID, tt.newAddress, tt.initiator)

			if tt.wantErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(request)
				suite.Require().NotEmpty(request.RequestId)
				suite.Require().Equal(walletID, request.WalletId)
				suite.Require().Equal(tt.newAddress, request.NewAddress)
				suite.Require().Equal(tt.initiator, request.Initiator)
				suite.Require().Contains(request.Approvals, tt.initiator)
				suite.Require().Equal(int32(1), request.ApprovalsCount)
				suite.Require().Equal(wsproto.RecoveryStatus_RECOVERY_STATUS_PENDING, request.Status)
			}
		})
	}
}

func (suite *SocialRecoveryTestSuite) TestInitiateRecoveryDisabled() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Non-existent wallet (no config)
	_, err := k.InitiateRecovery(ctx, "no_config_wallet", "aura1newowner", "aura1guardian")
	suite.Require().Error(err)
}

func (suite *SocialRecoveryTestSuite) TestApproveRecovery() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	walletID := "wallet_approve_test"
	guardians := []*wsproto.Guardian{
		{Address: "aura1guardian1"},
		{Address: "aura1guardian2"},
		{Address: "aura1guardian3"},
	}
	_, _ = k.ConfigureSocialRecovery(ctx, walletID, guardians, 2, nil)

	// Confirm all guardians
	for _, g := range guardians {
		k.ConfirmGuardian(ctx, walletID, g.Address)
	}

	// Initiate recovery
	request, _ := k.InitiateRecovery(ctx, walletID, "aura1newowner", "aura1guardian1")

	validSig := make([]byte, 64)

	tests := []struct {
		name      string
		requestID string
		guardian  string
		signature []byte
		wantErr   bool
		wantReady bool
	}{
		{
			name:      "second guardian approval reaches threshold",
			requestID: request.RequestId,
			guardian:  "aura1guardian2",
			signature: validSig,
			wantErr:   false,
			wantReady: true, // Now have 2 approvals
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ready, err := k.ApproveRecovery(ctx, tt.requestID, tt.guardian, tt.signature)

			if tt.wantErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tt.wantReady, ready)
			}
		})
	}
}

func (suite *SocialRecoveryTestSuite) TestApproveRecoveryDuplicate() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	walletID := "wallet_dup_approval"
	guardians := []*wsproto.Guardian{{Address: "aura1guardian1"}, {Address: "aura1guardian2"}}
	k.ConfigureSocialRecovery(ctx, walletID, guardians, 2, nil)
	k.ConfirmGuardian(ctx, walletID, "aura1guardian1")
	k.ConfirmGuardian(ctx, walletID, "aura1guardian2")

	request, _ := k.InitiateRecovery(ctx, walletID, "aura1newowner", "aura1guardian1")
	validSig := make([]byte, 64)

	// Guardian1 already approved (as initiator), try again
	_, err := k.ApproveRecovery(ctx, request.RequestId, "aura1guardian1", validSig)
	suite.Require().Error(err)
}

func (suite *SocialRecoveryTestSuite) TestExecuteRecovery() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	walletID := "wallet_execute_test"
	guardians := []*wsproto.Guardian{{Address: "aura1guardian1"}, {Address: "aura1guardian2"}}
	// Use very short recovery delay for testing
	recoveryDelay := &gogotypes.Duration{Seconds: 0}
	k.ConfigureSocialRecovery(ctx, walletID, guardians, 2, recoveryDelay)
	k.ConfirmGuardian(ctx, walletID, "aura1guardian1")
	k.ConfirmGuardian(ctx, walletID, "aura1guardian2")

	request, _ := k.InitiateRecovery(ctx, walletID, "aura1newowner", "aura1guardian1")

	// Add second approval
	validSig := make([]byte, 64)
	k.ApproveRecovery(ctx, request.RequestId, "aura1guardian2", validSig)

	// Execute recovery (delay is 0)
	err := k.ExecuteRecovery(ctx, request.RequestId)
	suite.Require().NoError(err)
}

func (suite *SocialRecoveryTestSuite) TestExecuteRecoveryInsufficientApprovals() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	walletID := "wallet_insuff_test"
	guardians := []*wsproto.Guardian{{Address: "aura1guardian1"}, {Address: "aura1guardian2"}}
	k.ConfigureSocialRecovery(ctx, walletID, guardians, 2, nil)
	k.ConfirmGuardian(ctx, walletID, "aura1guardian1")

	request, _ := k.InitiateRecovery(ctx, walletID, "aura1newowner", "aura1guardian1")

	// Try to execute with only 1 approval (need 2)
	err := k.ExecuteRecovery(ctx, request.RequestId)
	suite.Require().Error(err)
}

func (suite *SocialRecoveryTestSuite) TestExecuteRecoveryAlreadyExecuted() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	walletID := "wallet_already_exec"
	guardians := []*wsproto.Guardian{{Address: "aura1guardian1"}, {Address: "aura1guardian2"}}
	recoveryDelay := &gogotypes.Duration{Seconds: 0}
	k.ConfigureSocialRecovery(ctx, walletID, guardians, 2, recoveryDelay)
	k.ConfirmGuardian(ctx, walletID, "aura1guardian1")
	k.ConfirmGuardian(ctx, walletID, "aura1guardian2")

	request, _ := k.InitiateRecovery(ctx, walletID, "aura1newowner", "aura1guardian1")
	validSig := make([]byte, 64)
	k.ApproveRecovery(ctx, request.RequestId, "aura1guardian2", validSig)
	k.ExecuteRecovery(ctx, request.RequestId)

	// Try to execute again
	err := k.ExecuteRecovery(ctx, request.RequestId)
	suite.Require().Error(err)
}

func (suite *SocialRecoveryTestSuite) TestCancelRecovery() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	walletID := "wallet_cancel_test"
	guardians := []*wsproto.Guardian{{Address: "aura1guardian1"}, {Address: "aura1guardian2"}}
	k.ConfigureSocialRecovery(ctx, walletID, guardians, 2, nil)
	k.ConfirmGuardian(ctx, walletID, "aura1guardian1")

	request, _ := k.InitiateRecovery(ctx, walletID, "aura1newowner", "aura1guardian1")

	err := k.CancelRecovery(ctx, request.RequestId, "aura1walletowner")
	suite.Require().NoError(err)

	// Verify status is cancelled
	requestBytes, _ := k.GetRecoveryRequest(ctx, request.RequestId)
	var updatedRequest wsproto.RecoveryRequest
	k.cdc.Unmarshal(requestBytes, &updatedRequest)
	suite.Require().Equal(wsproto.RecoveryStatus_RECOVERY_STATUS_CANCELLED, updatedRequest.Status)
}

func (suite *SocialRecoveryTestSuite) TestAddGuardian() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	walletID := "wallet_add_guardian"
	guardians := []*wsproto.Guardian{{Address: "aura1guardian1"}, {Address: "aura1guardian2"}}
	k.ConfigureSocialRecovery(ctx, walletID, guardians, 2, nil)

	newGuardian := &wsproto.Guardian{Address: "aura1guardian3"}
	err := k.AddGuardian(ctx, walletID, newGuardian)
	suite.Require().NoError(err)

	// Verify guardian was added
	configBytes, _ := k.GetSocialRecoveryConfig(ctx, walletID)
	var config wsproto.SocialRecoveryConfig
	k.cdc.Unmarshal(configBytes, &config)
	suite.Require().Len(config.Guardians, 3)
}

func (suite *SocialRecoveryTestSuite) TestAddDuplicateGuardian() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	walletID := "wallet_dup_guardian"
	guardians := []*wsproto.Guardian{{Address: "aura1guardian1"}, {Address: "aura1guardian2"}}
	k.ConfigureSocialRecovery(ctx, walletID, guardians, 2, nil)

	duplicateGuardian := &wsproto.Guardian{Address: "aura1guardian1"}
	err := k.AddGuardian(ctx, walletID, duplicateGuardian)
	suite.Require().Error(err)
}

func (suite *SocialRecoveryTestSuite) TestRemoveGuardian() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	walletID := "wallet_remove_guardian"
	guardians := []*wsproto.Guardian{
		{Address: "aura1guardian1"},
		{Address: "aura1guardian2"},
		{Address: "aura1guardian3"},
	}
	k.ConfigureSocialRecovery(ctx, walletID, guardians, 2, nil)

	err := k.RemoveGuardian(ctx, walletID, "aura1guardian3")
	suite.Require().NoError(err)

	// Verify guardian was removed
	configBytes, _ := k.GetSocialRecoveryConfig(ctx, walletID)
	var config wsproto.SocialRecoveryConfig
	k.cdc.Unmarshal(configBytes, &config)
	suite.Require().Len(config.Guardians, 2)
}

func (suite *SocialRecoveryTestSuite) TestRemoveGuardianBreaksThreshold() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	walletID := "wallet_threshold_break"
	guardians := []*wsproto.Guardian{{Address: "aura1guardian1"}, {Address: "aura1guardian2"}}
	k.ConfigureSocialRecovery(ctx, walletID, guardians, 2, nil)

	// Removing would leave only 1 guardian, but threshold is 2
	err := k.RemoveGuardian(ctx, walletID, "aura1guardian2")
	suite.Require().Error(err)
}

func (suite *SocialRecoveryTestSuite) TestIsConfirmedGuardian() {
	k := suite.GetKeeper()

	guardians := []*wsproto.Guardian{
		{Address: "aura1confirmed", Confirmed: true},
		{Address: "aura1unconfirmed", Confirmed: false},
	}

	tests := []struct {
		address  string
		expected bool
	}{
		{"aura1confirmed", true},
		{"aura1unconfirmed", false},
		{"aura1unknown", false},
	}

	for _, tt := range tests {
		result := k.isConfirmedGuardian(tt.address, guardians)
		suite.Require().Equal(tt.expected, result, "address: %s", tt.address)
	}
}

func (suite *SocialRecoveryTestSuite) TestMaxGuardians() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Create more than max guardians (10)
	guardians := make([]*wsproto.Guardian, 11)
	for i := 0; i < 11; i++ {
		guardians[i] = &wsproto.Guardian{Address: "aura1guardian" + string(rune('a'+i))}
	}

	_, err := k.ConfigureSocialRecovery(ctx, "wallet_max_test", guardians, 2, nil)
	suite.Require().Error(err)
}

func (suite *SocialRecoveryTestSuite) TestValidateRecoverySignature() {
	k := suite.GetKeeper()

	tests := []struct {
		name      string
		signature []byte
		wantErr   bool
	}{
		{
			name:      "valid signature length",
			signature: make([]byte, 64),
			wantErr:   false,
		},
		{
			name:      "longer signature is valid",
			signature: make([]byte, 128),
			wantErr:   false,
		},
		{
			name:      "short signature fails",
			signature: make([]byte, 32),
			wantErr:   true,
		},
		{
			name:      "empty signature fails",
			signature: []byte{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := k.validateRecoverySignature("wallet", "newaddr", tt.signature, "guardian")

			if tt.wantErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
