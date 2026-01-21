// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/suite"

	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

type SpendingLimitTestSuite struct {
	KeeperTestSuite
}

func TestSpendingLimitTestSuite(t *testing.T) {
	suite.Run(t, new(SpendingLimitTestSuite))
}

func (suite *SpendingLimitTestSuite) TestSetSpendingLimit() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	tests := []struct {
		name         string
		walletID     string
		denom        string
		dailyLimit   string
		weeklyLimit  string
		monthlyLimit string
		wantErr      bool
	}{
		{
			name:         "valid spending limits",
			walletID:     "wallet1",
			denom:        "uaura",
			dailyLimit:   "1000000",
			weeklyLimit:  "5000000",
			monthlyLimit: "15000000",
			wantErr:      false,
		},
		{
			name:         "zero limits allowed",
			walletID:     "wallet2",
			denom:        "uaura",
			dailyLimit:   "0",
			weeklyLimit:  "0",
			monthlyLimit: "0",
			wantErr:      false,
		},
		{
			name:         "large limits",
			walletID:     "wallet3",
			denom:        "uaura",
			dailyLimit:   "999999999999999",
			weeklyLimit:  "999999999999999",
			monthlyLimit: "999999999999999",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			limit, err := k.SetSpendingLimit(ctx, tt.walletID, tt.denom, tt.dailyLimit, tt.weeklyLimit, tt.monthlyLimit)

			if tt.wantErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(limit)
				suite.Require().Equal(tt.walletID, limit.WalletId)
				suite.Require().Equal(tt.denom, limit.Denom)
				suite.Require().Equal(tt.dailyLimit, limit.DailyLimit)
				suite.Require().Equal(tt.weeklyLimit, limit.WeeklyLimit)
				suite.Require().Equal(tt.monthlyLimit, limit.MonthlyLimit)
				suite.Require().Equal("0", limit.CurrentDailySpent)
				suite.Require().Equal("0", limit.CurrentWeeklySpent)
				suite.Require().Equal("0", limit.CurrentMonthlySpent)
				suite.Require().True(limit.Enabled)
			}
		})
	}
}

func (suite *SpendingLimitTestSuite) TestCheckSpendingLimitWithinLimits() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Set spending limits
	_, err := k.SetSpendingLimit(ctx, "wallet1", "uaura", "1000000", "5000000", "15000000")
	suite.Require().NoError(err)

	// Spend within daily limit
	err = k.CheckSpendingLimit(ctx, "wallet1", "uaura", "500000")
	suite.Require().NoError(err)

	// Spend again within daily limit
	err = k.CheckSpendingLimit(ctx, "wallet1", "uaura", "400000")
	suite.Require().NoError(err)
}

func (suite *SpendingLimitTestSuite) TestCheckSpendingLimitExceedsDaily() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Set spending limit
	_, _ = k.SetSpendingLimit(ctx, "wallet1", "uaura", "1000000", "5000000", "15000000")

	// Try to spend more than daily limit in one transaction
	err := k.CheckSpendingLimit(ctx, "wallet1", "uaura", "2000000")
	suite.Require().Error(err)
}

func (suite *SpendingLimitTestSuite) TestCheckSpendingLimitExceedsWeekly() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Set limits: 2M daily, 3M weekly
	_, _ = k.SetSpendingLimit(ctx, "wallet1", "uaura", "2000000", "3000000", "15000000")

	// Spend within daily but pushing to weekly limit
	_ = k.CheckSpendingLimit(ctx, "wallet1", "uaura", "2000000")

	// Next spend exceeds weekly
	err := k.CheckSpendingLimit(ctx, "wallet1", "uaura", "2000000")
	suite.Require().Error(err)
}

func (suite *SpendingLimitTestSuite) TestCheckSpendingLimitNoLimit() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// No spending limit configured - should fail
	err := k.CheckSpendingLimit(ctx, "wallet_no_limit", "uaura", "1000000")
	suite.Require().Error(err)
}

func (suite *SpendingLimitTestSuite) TestCheckSpendingLimitInvalidAmount() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	_, _ = k.SetSpendingLimit(ctx, "wallet1", "uaura", "1000000", "5000000", "15000000")

	// Zero amount should fail
	err := k.CheckSpendingLimit(ctx, "wallet1", "uaura", "0")
	suite.Require().Error(err)

	// Negative amount should fail
	err = k.CheckSpendingLimit(ctx, "wallet1", "uaura", "-1000")
	suite.Require().Error(err)
}

func (suite *SpendingLimitTestSuite) TestGetSpendingLimit() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()

	// Set a limit
	_, _ = k.SetSpendingLimit(ctx, "wallet1", "uaura", "1000000", "5000000", "15000000")

	// Retrieve it
	limitBytes, err := k.GetSpendingLimit(ctx, "wallet1", "uaura")
	suite.Require().NoError(err)
	suite.Require().NotNil(limitBytes)

	// Non-existent limit
	_, err = k.GetSpendingLimit(ctx, "nonexistent", "uaura")
	suite.Require().Error(err)
}

func (suite *SpendingLimitTestSuite) TestCheckDustTransaction() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()
	cdc := suite.GetCodec()

	walletID := "wallet_dust_test"

	// No dust filter configured - should pass
	isDust, err := k.CheckDustTransaction(ctx, walletID, "tx1", "aura1sender", "aura1recipient", "1000", "uaura")
	suite.Require().NoError(err)
	suite.Require().False(isDust)

	// Configure dust filter
	filter := &wsproto.DustAttackFilter{
		WalletId:                   walletID,
		Enabled:                    true,
		MinimumAmount:              "10000",
		BlockedSenders:             []string{"aura1blockedsender"},
		SuspiciousPatternThreshold: 80,
	}
	filterBytes, _ := cdc.Marshal(filter)
	_ = k.SetDustFilter(ctx, walletID, filterBytes)

	tests := []struct {
		name        string
		txHash      string
		fromAddress string
		toAddress   string
		amount      string
		denom       string
		expectDust  bool
	}{
		{
			name:        "amount above minimum passes",
			txHash:      "tx_normal",
			fromAddress: "aura1normalsender",
			toAddress:   walletID,
			amount:      "50000",
			denom:       "uaura",
			expectDust:  false,
		},
		{
			name:        "amount below minimum is dust",
			txHash:      "tx_dust",
			fromAddress: "aura1normalsender",
			toAddress:   walletID,
			amount:      "100",
			denom:       "uaura",
			expectDust:  true,
		},
		{
			name:        "blocked sender is dust",
			txHash:      "tx_blocked",
			fromAddress: "aura1blockedsender",
			toAddress:   walletID,
			amount:      "50000",
			denom:       "uaura",
			expectDust:  true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			isDust, err := k.CheckDustTransaction(ctx, walletID, tt.txHash, tt.fromAddress, tt.toAddress, tt.amount, tt.denom)
			suite.Require().NoError(err)
			suite.Require().Equal(tt.expectDust, isDust)
		})
	}
}

func (suite *SpendingLimitTestSuite) TestCheckDustTransactionDisabled() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()
	cdc := suite.GetCodec()

	walletID := "wallet_dust_disabled"

	// Configure disabled dust filter
	filter := &wsproto.DustAttackFilter{
		WalletId:      walletID,
		Enabled:       false,
		MinimumAmount: "10000",
	}
	filterBytes, _ := cdc.Marshal(filter)
	_ = k.SetDustFilter(ctx, walletID, filterBytes)

	// Should pass even with tiny amount since disabled
	isDust, err := k.CheckDustTransaction(ctx, walletID, "tx1", "aura1sender", walletID, "1", "uaura")
	suite.Require().NoError(err)
	suite.Require().False(isDust)
}

func (suite *SpendingLimitTestSuite) TestCheckDustTransactionNegativeAmount() {
	suite.SetupTest()
	k := suite.GetKeeper()
	ctx := suite.GetContext()
	cdc := suite.GetCodec()

	walletID := "wallet_neg_test"

	filter := &wsproto.DustAttackFilter{
		WalletId:      walletID,
		Enabled:       true,
		MinimumAmount: "100",
	}
	filterBytes, _ := cdc.Marshal(filter)
	_ = k.SetDustFilter(ctx, walletID, filterBytes)

	// Negative amount should fail
	_, err := k.CheckDustTransaction(ctx, walletID, "tx1", "aura1sender", walletID, "-1000", "uaura")
	suite.Require().Error(err)
}
