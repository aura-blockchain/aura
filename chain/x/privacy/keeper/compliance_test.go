// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/privacy/keeper"
)

type ComplianceTestSuite struct {
	suite.Suite

	keeper *keeper.Keeper
	ctx    sdk.Context
}

func TestComplianceTestSuite(t *testing.T) {
	suite.Run(t, new(ComplianceTestSuite))
}

func (suite *ComplianceTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.keeper = keeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil,
		nil,
	)
	suite.ctx = input.Ctx
}

// CheckPrivacyCompliance Tests

func (suite *ComplianceTestSuite) TestCheckPrivacyCompliance_EU() {
	compliant, err := suite.keeper.CheckPrivacyCompliance(suite.ctx, "EU")

	suite.Require().NoError(err)
	suite.Require().True(compliant)
}

func (suite *ComplianceTestSuite) TestCheckPrivacyCompliance_US() {
	compliant, err := suite.keeper.CheckPrivacyCompliance(suite.ctx, "US")

	suite.Require().NoError(err)
	suite.Require().True(compliant)
}

func (suite *ComplianceTestSuite) TestCheckPrivacyCompliance_Unknown() {
	compliant, err := suite.keeper.CheckPrivacyCompliance(suite.ctx, "UNKNOWN")

	suite.Require().NoError(err)
	suite.Require().True(compliant) // Default to allowing
}

// RegisterViewKey Tests

func (suite *ComplianceTestSuite) TestRegisterViewKey_EmptyPublicKey() {
	owner := keepertest.GenTestAddr().String()

	err := suite.keeper.RegisterViewKey(suite.ctx, owner, []byte{})

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "public key cannot be empty")
}

func (suite *ComplianceTestSuite) TestRegisterViewKey_InvalidKeyLength() {
	owner := keepertest.GenTestAddr().String()
	invalidKey := make([]byte, 16) // Invalid length

	err := suite.keeper.RegisterViewKey(suite.ctx, owner, invalidKey)

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid public key length")
}

func (suite *ComplianceTestSuite) TestRegisterViewKey_ValidKey32Bytes() {
	owner := keepertest.GenTestAddr().String()
	publicKey := make([]byte, 32)
	for i := range publicKey {
		publicKey[i] = byte(i)
	}

	err := suite.keeper.RegisterViewKey(suite.ctx, owner, publicKey)

	suite.Require().NoError(err)

	// Verify key was stored
	viewKey, err := suite.keeper.GetViewKeyByPublic(suite.ctx, publicKey)
	suite.Require().NoError(err)
	suite.Require().NotNil(viewKey)
	suite.Require().Equal(publicKey, viewKey.PublicViewKey)
}

func (suite *ComplianceTestSuite) TestRegisterViewKey_ValidKey33Bytes() {
	owner := keepertest.GenTestAddr().String()
	publicKey := make([]byte, 33)
	for i := range publicKey {
		publicKey[i] = byte(i)
	}

	err := suite.keeper.RegisterViewKey(suite.ctx, owner, publicKey)

	suite.Require().NoError(err)

	// Verify key was stored
	viewKey, err := suite.keeper.GetViewKeyByPublic(suite.ctx, publicKey)
	suite.Require().NoError(err)
	suite.Require().NotNil(viewKey)
}

func (suite *ComplianceTestSuite) TestRegisterViewKey_ValidKey64Bytes() {
	owner := keepertest.GenTestAddr().String()
	publicKey := make([]byte, 64)
	for i := range publicKey {
		publicKey[i] = byte(i)
	}

	err := suite.keeper.RegisterViewKey(suite.ctx, owner, publicKey)

	suite.Require().NoError(err)

	// Verify key was stored
	viewKey, err := suite.keeper.GetViewKeyByPublic(suite.ctx, publicKey)
	suite.Require().NoError(err)
	suite.Require().NotNil(viewKey)
}

// SelectiveDisclose Tests

func (suite *ComplianceTestSuite) TestSelectiveDisclose_ViewKeyNotFound() {
	txID := "tx_123"
	viewKey := make([]byte, 32)

	details, err := suite.keeper.SelectiveDisclose(suite.ctx, txID, viewKey)

	suite.Require().Error(err)
	suite.Require().Nil(details)
	suite.Require().Contains(err.Error(), "view key not found")
}

func (suite *ComplianceTestSuite) TestSelectiveDisclose_Success() {
	owner := keepertest.GenTestAddr().String()
	publicKey := make([]byte, 32)
	for i := range publicKey {
		publicKey[i] = byte(i)
	}

	// Register view key first
	err := suite.keeper.RegisterViewKey(suite.ctx, owner, publicKey)
	suite.Require().NoError(err)

	txID := "tx_123"

	details, err := suite.keeper.SelectiveDisclose(suite.ctx, txID, publicKey)

	suite.Require().NoError(err)
	suite.Require().NotNil(details)
	suite.Require().Equal(txID, details["tx_id"])
	suite.Require().Equal(true, details["disclosed"])
	suite.Require().Equal(owner, details["address"])
}

// AuditPrivacyOperation Tests

func (suite *ComplianceTestSuite) TestAuditPrivacyOperation_Success() {
	operation := "view_key_registration"
	details := "User registered view key for compliance"

	err := suite.keeper.AuditPrivacyOperation(suite.ctx, operation, details)

	suite.Require().NoError(err)
}

func (suite *ComplianceTestSuite) TestAuditPrivacyOperation_MultipleAudits() {
	operations := []struct {
		operation string
		details   string
	}{
		{"view_key_registration", "User A registered view key"},
		{"selective_disclosure", "User B disclosed transaction"},
		{"compliance_check", "EU GDPR compliance verified"},
	}

	for _, op := range operations {
		err := suite.keeper.AuditPrivacyOperation(suite.ctx, op.operation, op.details)
		suite.Require().NoError(err)
	}
}

// Standalone tests for better coverage

func TestRegisterViewKey_AllLengths(t *testing.T) {
	testCases := []struct {
		name      string
		keyLength int
		shouldErr bool
	}{
		{"16 bytes - invalid", 16, true},
		{"32 bytes - valid", 32, false},
		{"33 bytes - valid", 33, false},
		{"64 bytes - valid", 64, false},
		{"65 bytes - invalid", 65, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := keepertest.CreateTestInput(t)
			k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

			owner := keepertest.GenTestAddr().String()
			publicKey := make([]byte, tc.keyLength)

			err := k.RegisterViewKey(input.Ctx, owner, publicKey)

			if tc.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSelectiveDisclose_WithDifferentOwners(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(input.Cdc, input.StoreKey, nil, nil)

	owner1 := keepertest.GenTestAddr().String()
	owner2 := keepertest.GenTestAddr().String()

	publicKey1 := make([]byte, 32)
	publicKey2 := make([]byte, 32)
	for i := range publicKey1 {
		publicKey1[i] = byte(i)
		publicKey2[i] = byte(i + 32)
	}

	// Register both view keys
	err := k.RegisterViewKey(input.Ctx, owner1, publicKey1)
	require.NoError(t, err)

	err = k.RegisterViewKey(input.Ctx, owner2, publicKey2)
	require.NoError(t, err)

	// Disclose with owner1's key
	txID := "tx_123"
	details1, err := k.SelectiveDisclose(input.Ctx, txID, publicKey1)
	require.NoError(t, err)
	require.Equal(t, owner1, details1["address"])

	// Disclose with owner2's key
	details2, err := k.SelectiveDisclose(input.Ctx, txID, publicKey2)
	require.NoError(t, err)
	require.Equal(t, owner2, details2["address"])
}
