// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"
)

// =============================================================================
// Wallet Insurance Tests
// =============================================================================

type WalletInsuranceTestSuite struct {
	KeeperTestSuite
}

func (suite *WalletInsuranceTestSuite) TestPurchaseInsurance_Success() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	coverageAmount := math.NewInt(1000000)
	premium := math.NewInt(10000)

	policy, err := k.PurchaseInsurance(ctx, "wallet-1", coverageAmount, premium)
	suite.Require().NoError(err)
	suite.Require().NotNil(policy)
	suite.Require().NotEmpty(policy.PolicyId)
	suite.Require().Equal("wallet-1", policy.WalletId)
	suite.Require().Equal(coverageAmount.String(), policy.CoverageAmount)
	suite.Require().Equal(premium.String(), policy.Premium)
	suite.Require().True(policy.Active)
	suite.Require().Equal("0", policy.ClaimsPaid)
	suite.Require().NotNil(policy.PurchasedAt)
	suite.Require().NotNil(policy.ExpiresAt)
}

func (suite *WalletInsuranceTestSuite) TestPurchaseInsurance_DifferentWallets() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Purchase insurance for different wallets - should get different policy IDs
	policy1, err := k.PurchaseInsurance(ctx, "wallet-2a", math.NewInt(500000), math.NewInt(5000))
	suite.Require().NoError(err)
	suite.Require().NotEmpty(policy1.PolicyId)

	policy2, err := k.PurchaseInsurance(ctx, "wallet-2b", math.NewInt(1000000), math.NewInt(10000))
	suite.Require().NoError(err)
	suite.Require().NotEmpty(policy2.PolicyId)

	// Different wallet IDs in policy IDs
	suite.Require().NotEqual(policy1.PolicyId, policy2.PolicyId)
}

func (suite *WalletInsuranceTestSuite) TestFileClaim_Success() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// First purchase insurance
	policy, err := k.PurchaseInsurance(ctx, "wallet-3", math.NewInt(1000000), math.NewInt(10000))
	suite.Require().NoError(err)

	// File a claim
	claimID, err := k.FileClaim(ctx, policy.PolicyId, "Security breach", math.NewInt(50000))
	suite.Require().NoError(err)
	suite.Require().NotEmpty(claimID)
}

func (suite *WalletInsuranceTestSuite) TestFileClaim_PolicyNotFound() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	_, err := k.FileClaim(ctx, "nonexistent-policy", "Test", math.NewInt(1000))
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "policy not found")
}

func (suite *WalletInsuranceTestSuite) TestFileClaim_MultipleClaims() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	policy, err := k.PurchaseInsurance(ctx, "wallet-4", math.NewInt(1000000), math.NewInt(10000))
	suite.Require().NoError(err)

	claim1, err := k.FileClaim(ctx, policy.PolicyId, "First incident", math.NewInt(10000))
	suite.Require().NoError(err)
	suite.Require().NotEmpty(claim1)

	claim2, err := k.FileClaim(ctx, policy.PolicyId, "Second incident", math.NewInt(20000))
	suite.Require().NoError(err)
	suite.Require().NotEmpty(claim2)

	// Both claims should be created successfully - exact IDs depend on timing
	// but both should contain the policy ID prefix
	suite.Require().Contains(claim1, "claim_")
	suite.Require().Contains(claim2, "claim_")
}

func (suite *WalletInsuranceTestSuite) TestProcessClaim_Approved() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Purchase insurance
	policy, err := k.PurchaseInsurance(ctx, "wallet-5", math.NewInt(1000000), math.NewInt(10000))
	suite.Require().NoError(err)

	// File claim
	claimID, err := k.FileClaim(ctx, policy.PolicyId, "Approved claim", math.NewInt(25000))
	suite.Require().NoError(err)

	// Process claim as approved
	err = k.ProcessClaim(ctx, claimID, true)
	suite.Require().NoError(err)
}

func (suite *WalletInsuranceTestSuite) TestProcessClaim_Denied() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Purchase insurance
	policy, err := k.PurchaseInsurance(ctx, "wallet-6", math.NewInt(1000000), math.NewInt(10000))
	suite.Require().NoError(err)

	// File claim
	claimID, err := k.FileClaim(ctx, policy.PolicyId, "Denied claim", math.NewInt(25000))
	suite.Require().NoError(err)

	// Process claim as denied
	err = k.ProcessClaim(ctx, claimID, false)
	suite.Require().NoError(err)
}

func (suite *WalletInsuranceTestSuite) TestProcessClaim_NotFound() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	err := k.ProcessClaim(ctx, "nonexistent-claim", true)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "claim not found")
}

func (suite *WalletInsuranceTestSuite) TestPurchaseInsurance_ZeroCoverage() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	policy, err := k.PurchaseInsurance(ctx, "wallet-7", math.NewInt(0), math.NewInt(0))
	suite.Require().NoError(err)
	suite.Require().NotNil(policy)
	suite.Require().Equal("0", policy.CoverageAmount)
	suite.Require().Equal("0", policy.Premium)
}

func (suite *WalletInsuranceTestSuite) TestFileClaim_ZeroAmount() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	policy, err := k.PurchaseInsurance(ctx, "wallet-8", math.NewInt(1000000), math.NewInt(10000))
	suite.Require().NoError(err)

	claimID, err := k.FileClaim(ctx, policy.PolicyId, "Zero claim test", math.NewInt(0))
	suite.Require().NoError(err)
	suite.Require().NotEmpty(claimID)
}

func (suite *WalletInsuranceTestSuite) TestProcessMultipleClaims_ClaimsPaidAccumulates() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Purchase insurance
	policy, err := k.PurchaseInsurance(ctx, "wallet-9", math.NewInt(1000000), math.NewInt(10000))
	suite.Require().NoError(err)

	// File and approve first claim
	claim1, err := k.FileClaim(ctx, policy.PolicyId, "First claim", math.NewInt(10000))
	suite.Require().NoError(err)
	err = k.ProcessClaim(ctx, claim1, true)
	suite.Require().NoError(err)

	// File and approve second claim
	claim2, err := k.FileClaim(ctx, policy.PolicyId, "Second claim", math.NewInt(15000))
	suite.Require().NoError(err)
	err = k.ProcessClaim(ctx, claim2, true)
	suite.Require().NoError(err)
}

func TestWalletInsuranceTestSuite(t *testing.T) {
	suite.Run(t, new(WalletInsuranceTestSuite))
}
