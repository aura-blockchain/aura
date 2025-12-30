// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// =============================================================================
// Wallet Analytics Tests
// =============================================================================

type WalletAnalyticsTestSuite struct {
	KeeperTestSuite
}

func (suite *WalletAnalyticsTestSuite) TestGetWalletAnalytics_NewWallet() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	analytics, err := k.GetWalletAnalytics(ctx, "new-wallet")
	suite.Require().NoError(err)
	suite.Require().NotNil(analytics)
	suite.Require().Equal("new-wallet", analytics.WalletID)
	suite.Require().Equal(int64(0), analytics.TotalTransactions)
	suite.Require().True(analytics.TotalVolume.IsZero())
	suite.Require().True(analytics.AverageTransactionSize.IsZero())
	suite.Require().NotNil(analytics.EnabledFeatures)
}

func (suite *WalletAnalyticsTestSuite) TestGetWalletAnalytics_WithDevices() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Register trusted devices
	_, err := k.RegisterDevice(ctx, "wallet-with-devices", "device-1", "Phone", "mobile", []byte("fp1"))
	suite.Require().NoError(err)
	_, err = k.RegisterDevice(ctx, "wallet-with-devices", "device-2", "Laptop", "desktop", []byte("fp2"))
	suite.Require().NoError(err)

	analytics, err := k.GetWalletAnalytics(ctx, "wallet-with-devices")
	suite.Require().NoError(err)
	suite.Require().Equal(2, analytics.ActiveDevices)
}

func (suite *WalletAnalyticsTestSuite) TestCalculateSecurityScore_NewWallet() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// New wallet with no features should have reduced score
	score := k.calculateSecurityScore(ctx, "new-wallet-score")
	// No multisig (-20), no hardware wallet (-15), no social recovery (-10), no biometric (-10) = 45
	suite.Require().Equal(45.0, score)
}

func (suite *WalletAnalyticsTestSuite) TestCalculateSecurityScore_WithMultiSig() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Set up multisig wallet - walletID is generated from creator and signers
	msWallet, err := k.CreateMultiSigWallet(ctx, "creator1", []string{"owner1", "owner2", "owner3"}, 2, nil, 0, nil)
	suite.Require().NoError(err)
	suite.Require().NotNil(msWallet)

	// Check security score using the actual wallet ID
	score := k.calculateSecurityScore(ctx, msWallet.WalletId)
	// With multisig (+20) but no hardware wallet (-15), no social recovery (-10), no biometric (-10) = 65
	suite.Require().Equal(65.0, score)
}

func (suite *WalletAnalyticsTestSuite) TestDetermineRiskLevel_LowRisk() {
	k := suite.GetKeeper()

	// High score = low risk
	suite.Require().Equal("low", k.determineRiskLevel(85))
	suite.Require().Equal("low", k.determineRiskLevel(100))
	suite.Require().Equal("low", k.determineRiskLevel(80))
}

func (suite *WalletAnalyticsTestSuite) TestDetermineRiskLevel_MediumRisk() {
	k := suite.GetKeeper()

	suite.Require().Equal("medium", k.determineRiskLevel(79))
	suite.Require().Equal("medium", k.determineRiskLevel(65))
	suite.Require().Equal("medium", k.determineRiskLevel(60))
}

func (suite *WalletAnalyticsTestSuite) TestDetermineRiskLevel_HighRisk() {
	k := suite.GetKeeper()

	suite.Require().Equal("high", k.determineRiskLevel(59))
	suite.Require().Equal("high", k.determineRiskLevel(45))
	suite.Require().Equal("high", k.determineRiskLevel(40))
}

func (suite *WalletAnalyticsTestSuite) TestDetermineRiskLevel_CriticalRisk() {
	k := suite.GetKeeper()

	suite.Require().Equal("critical", k.determineRiskLevel(39))
	suite.Require().Equal("critical", k.determineRiskLevel(20))
	suite.Require().Equal("critical", k.determineRiskLevel(0))
}

func (suite *WalletAnalyticsTestSuite) TestGetEnabledFeatures_Empty() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	features := k.getEnabledFeatures(ctx, "empty-features-wallet")
	suite.Require().NotNil(features)
	suite.Require().Len(features, 0)
}

func (suite *WalletAnalyticsTestSuite) TestGetEnabledFeatures_WithMultiSig() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Create multisig wallet - walletID is generated from creator + signers
	msWallet, err := k.CreateMultiSigWallet(ctx, "wallet-features", []string{"owner1", "owner2"}, 2, nil, 0, nil)
	suite.Require().NoError(err)

	features := k.getEnabledFeatures(ctx, msWallet.WalletId)
	suite.Require().Contains(features, "multisig")
}

func (suite *WalletAnalyticsTestSuite) TestContainsHelper() {
	k := suite.GetKeeper()

	slice := []string{"multisig", "hardware_wallet", "biometric"}

	suite.Require().True(k.contains(slice, "multisig"))
	suite.Require().True(k.contains(slice, "hardware_wallet"))
	suite.Require().False(k.contains(slice, "social_recovery"))
	suite.Require().False(k.contains(nil, "anything"))
	suite.Require().False(k.contains([]string{}, "anything"))
}

func (suite *WalletAnalyticsTestSuite) TestGenerateSecurityReport() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	report, err := k.GenerateSecurityReport(ctx, "wallet-report")
	suite.Require().NoError(err)
	suite.Require().NotNil(report)
	suite.Require().Equal("wallet-report", report.WalletId)
	suite.Require().NotNil(report.Recommendations)
}

func (suite *WalletAnalyticsTestSuite) TestGenerateSecurityReport_WithRecommendations() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	report, err := k.GenerateSecurityReport(ctx, "wallet-needs-recs")
	suite.Require().NoError(err)
	suite.Require().NotNil(report)
	// Score is 45, below 80, so should have recommendations
	suite.Require().True(len(report.Recommendations) > 0)
}

func (suite *WalletAnalyticsTestSuite) TestGenerateRecommendations_LowScore() {
	k := suite.GetKeeper()

	analytics := &WalletAnalytics{
		SecurityScore:   45,
		EnabledFeatures: []string{},
	}

	recommendations := k.generateRecommendations(analytics)
	suite.Require().NotNil(recommendations)
	suite.Require().Greater(len(recommendations), 0)
	suite.Require().Contains(recommendations, "Enable multi-signature protection")
	suite.Require().Contains(recommendations, "Use a hardware wallet for enhanced security")
}

func (suite *WalletAnalyticsTestSuite) TestGenerateRecommendations_HighScore() {
	k := suite.GetKeeper()

	analytics := &WalletAnalytics{
		SecurityScore:   95,
		EnabledFeatures: []string{"multisig", "hardware_wallet", "social_recovery", "biometric"},
	}

	recommendations := k.generateRecommendations(analytics)
	suite.Require().NotNil(recommendations)
	suite.Require().Len(recommendations, 0)
}

func (suite *WalletAnalyticsTestSuite) TestCountTransactions_Empty() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	count := k.countTransactions(ctx, "wallet-no-txs")
	suite.Require().Equal(int64(0), count)
}

func (suite *WalletAnalyticsTestSuite) TestCalculateVolumes_NoTransactions() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	total, average := k.calculateVolumes(ctx, "wallet-no-volume")
	suite.Require().True(total.IsZero())
	suite.Require().True(average.IsZero())
}

func (suite *WalletAnalyticsTestSuite) TestHasMultiSig_False() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	result := k.hasMultiSig(ctx, "wallet-no-multisig")
	suite.Require().False(result)
}

func (suite *WalletAnalyticsTestSuite) TestHasMultiSig_True() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	msWallet, err := k.CreateMultiSigWallet(ctx, "wallet-with-multisig", []string{"owner1", "owner2"}, 2, nil, 0, nil)
	suite.Require().NoError(err)

	result := k.hasMultiSig(ctx, msWallet.WalletId)
	suite.Require().True(result)
}

func (suite *WalletAnalyticsTestSuite) TestHasHardwareWallet_False() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	result := k.hasHardwareWallet(ctx, "wallet-no-hw")
	suite.Require().False(result)
}

func (suite *WalletAnalyticsTestSuite) TestHasSocialRecovery_False() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	result := k.hasSocialRecovery(ctx, "wallet-no-social")
	suite.Require().False(result)
}

func (suite *WalletAnalyticsTestSuite) TestHasBiometric_False() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	result := k.hasBiometric(ctx, "wallet-no-biometric")
	suite.Require().False(result)
}

func TestWalletAnalyticsTestSuite(t *testing.T) {
	suite.Run(t, new(WalletAnalyticsTestSuite))
}
