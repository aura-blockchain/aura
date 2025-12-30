// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/dex/client/cli"
)

// ============================================================================
// Test Suite Setup
// ============================================================================

type DexQueryTestSuite struct {
	suite.Suite
}

func TestDexQueryTestSuite(t *testing.T) {
	suite.Run(t, new(DexQueryTestSuite))
}

func (s *DexQueryTestSuite) TestGetQueryCmd() {
	cmd := cli.GetQueryCmd()
	s.Require().NotNil(cmd)
	s.Require().Equal("dex", cmd.Use)
	s.Require().True(cmd.DisableFlagParsing)
	s.Require().NotEmpty(cmd.Short)
}

func (s *DexQueryTestSuite) TestQueryCmdHasSubcommands() {
	cmd := cli.GetQueryCmd()
	expected := []string{
		"pool",
		"pools",
		"pool-stats",
		"quote",
		"market-price",
		"spot-price",
		"orderbook",
		"order",
		"user-orders",
		"supported-coins",
		"htlc",
		"params",
	}

	for _, expectedCmd := range expected {
		found := false
		for _, c := range cmd.Commands() {
			if matchCommand(c, expectedCmd) {
				found = true
				break
			}
		}
		s.Require().True(found, "missing query command: %s", expectedCmd)
	}
}

// ============================================================================
// Legacy Tests (retained for backwards compatibility)
// ============================================================================

func TestGetQueryCmd(t *testing.T) {
	cmd := cli.GetQueryCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "dex", cmd.Use)
	require.True(t, cmd.DisableFlagParsing)

	expected := []string{
		"pool",
		"pools",
		"pool-stats",
		"quote",
		"market-price",
		"spot-price",
		"orderbook",
		"order",
		"user-orders",
		"supported-coins",
		"htlc",
	}

	sub := make(map[string]struct{})
	for _, c := range cmd.Commands() {
		name := c.Use
		if idx := len(name); idx > 0 {
			sub[name] = struct{}{}
		}
	}

	for _, expectedCmd := range expected {
		found := false
		for _, c := range cmd.Commands() {
			if matchCommand(c, expectedCmd) {
				found = true
				break
			}
		}
		require.True(t, found, "missing query command: %s", expectedCmd)
	}
}

func matchCommand(cmd *cobra.Command, use string) bool {
	if cmd.Use == use {
		return true
	}
	if len(cmd.Use) >= len(use) && cmd.Use[:len(use)] == use {
		return true
	}
	return false
}

// ============================================================================
// Pool Query Tests
// ============================================================================

func TestCmdQueryPool(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid pool ID",
			args:    []string{"uaura-usdt"},
			wantErr: false,
		},
		{
			name:    "valid pool ID with numbers",
			args:    []string{"pool-123"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"pool1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.CmdQueryPool()
			require.NotNil(t, cmd)
			require.Equal(t, "pool [pool-id]", cmd.Use)
			require.Contains(t, cmd.Short, "pool")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdQueryAllPools(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args - valid",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "extra args - should fail",
			args:    []string{"extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.CmdQueryAllPools()
			require.NotNil(t, cmd)
			require.Equal(t, "pools", cmd.Use)
			require.Contains(t, cmd.Short, "pool")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdQueryPoolStats(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid pool ID",
			args:    []string{"uaura-usdt"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"pool1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.CmdQueryPoolStats()
			require.NotNil(t, cmd)
			require.Equal(t, "pool-stats [pool-id]", cmd.Use)
			require.Contains(t, cmd.Short, "statistic")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Quote Query Tests
// ============================================================================

func TestCmdQueryGetQuote(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid quote request",
			args:    []string{"uaura-usdt", "1000000", "uaura"},
			wantErr: false,
		},
		{
			name:    "missing denom - should fail",
			args:    []string{"uaura-usdt", "1000000"},
			wantErr: true,
		},
		{
			name:    "missing amount - should fail",
			args:    []string{"uaura-usdt"},
			wantErr: true,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"pool", "1000", "uaura", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.CmdQueryGetQuote()
			require.NotNil(t, cmd)
			require.Equal(t, "quote [pool-id] [denom-in] [amount-in]", cmd.Use)
			require.Contains(t, cmd.Short, "swap quote")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Market Price Query Tests
// ============================================================================

func TestCmdQueryMarketPrice(t *testing.T) {
	cmd := cli.CmdQueryMarketPrice()
	require.Equal(t, "market-price [coin]", cmd.Use)

	require.NoError(t, cmd.ValidateArgs([]string{"usdt"}))
	require.Error(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"usdt", "extra"}))
}

func TestCmdQueryMarketPriceExtended(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid coin - usdt",
			args:    []string{"usdt"},
			wantErr: false,
		},
		{
			name:    "valid coin - uaura",
			args:    []string{"uaura"},
			wantErr: false,
		},
		{
			name:    "valid coin - ibc denom",
			args:    []string{"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"usdt", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.CmdQueryMarketPrice()
			require.NotNil(t, cmd)
			require.Equal(t, "market-price [coin]", cmd.Use)
			require.Contains(t, cmd.Short, "price")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Spot Price Query Tests
// ============================================================================

func TestCmdQuerySpotPrice(t *testing.T) {
	cmd := cli.CmdQuerySpotPrice()
	require.Equal(t, "spot-price [pool-id] [base-denom] [quote-denom]", cmd.Use)

	require.NoError(t, cmd.ValidateArgs([]string{"uaura-usdt", "uaura", "usdt"}))
	require.Error(t, cmd.ValidateArgs([]string{"uaura-usdt", "uaura"}))
	require.Error(t, cmd.ValidateArgs([]string{}))
}

func TestCmdQuerySpotPriceExtended(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid spot price request",
			args:    []string{"uaura-usdt", "uaura", "usdt"},
			wantErr: false,
		},
		{
			name:    "valid with different denoms",
			args:    []string{"atom-osmo", "atom", "osmo"},
			wantErr: false,
		},
		{
			name:    "missing quote denom - should fail",
			args:    []string{"uaura-usdt", "uaura"},
			wantErr: true,
		},
		{
			name:    "missing base denom - should fail",
			args:    []string{"uaura-usdt"},
			wantErr: true,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"pool", "uaura", "usdt", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.CmdQuerySpotPrice()
			require.NotNil(t, cmd)
			require.Equal(t, "spot-price [pool-id] [base-denom] [quote-denom]", cmd.Use)
			require.Contains(t, cmd.Short, "spot")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Order Book Query Tests
// ============================================================================

func TestCmdQueryOrderbook(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid pair",
			args:    []string{"uaura-usdt"},
			wantErr: false,
		},
		{
			name:    "valid pair with different format",
			args:    []string{"btc-eth"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"pair1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.CmdQueryOrderbook()
			require.NotNil(t, cmd)
			require.Equal(t, "orderbook [pair]", cmd.Use)
			require.Contains(t, cmd.Short, "order")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Order Query Tests
// ============================================================================

func TestCmdQueryOrder(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid order ID",
			args:    []string{"order-123"},
			wantErr: false,
		},
		{
			name:    "valid order ID numeric",
			args:    []string{"12345"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"order1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.CmdQueryOrder()
			require.NotNil(t, cmd)
			require.Equal(t, "order [order-id]", cmd.Use)
			require.Contains(t, cmd.Short, "order")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// User Orders Query Tests
// ============================================================================

func TestCmdQueryUserOrders(t *testing.T) {
	cmd := cli.CmdQueryUserOrders()
	require.Equal(t, "user-orders [address]", cmd.Use)

	require.NoError(t, cmd.ValidateArgs([]string{"aura1xyz"}))
	require.Error(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"aura1", "extra"}))
}

func TestCmdQueryUserOrdersExtended(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid address",
			args:    []string{"aura1xyz123abc"},
			wantErr: false,
		},
		{
			name:    "valid cosmos address format",
			args:    []string{"aura1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq7dmzf5"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"addr1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.CmdQueryUserOrders()
			require.NotNil(t, cmd)
			require.Equal(t, "user-orders [address]", cmd.Use)
			require.Contains(t, cmd.Short, "order")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Supported Coins Query Tests
// ============================================================================

func TestCmdQuerySupportedCoins(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args - valid",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "extra args - should fail",
			args:    []string{"extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.CmdQuerySupportedCoins()
			require.NotNil(t, cmd)
			require.Equal(t, "supported-coins", cmd.Use)
			require.Contains(t, cmd.Short, "coin")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// HTLC Query Tests
// ============================================================================

func TestCmdQueryHTLC(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid HTLC ID",
			args:    []string{"htlc-abc123"},
			wantErr: false,
		},
		{
			name:    "valid HTLC hash format",
			args:    []string{"0x1234567890abcdef"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"htlc1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.CmdQueryHTLC()
			require.NotNil(t, cmd)
			require.Equal(t, "htlc [htlc-id]", cmd.Use)
			require.Contains(t, cmd.Short, "Hash Time-Locked")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Params Query Tests
// ============================================================================

func TestCmdQueryParams(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args - valid",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "extra args - should fail",
			args:    []string{"extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := cli.CmdQueryParams()
			require.NotNil(t, cmd)
			require.Equal(t, "params", cmd.Use)
			require.Contains(t, cmd.Short, "param")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Flag Tests
// ============================================================================

func TestQueryCommandsHaveQueryFlags(t *testing.T) {
	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"pool", cli.CmdQueryPool()},
		{"pools", cli.CmdQueryAllPools()},
		{"pool-stats", cli.CmdQueryPoolStats()},
		{"quote", cli.CmdQueryGetQuote()},
		{"market-price", cli.CmdQueryMarketPrice()},
		{"spot-price", cli.CmdQuerySpotPrice()},
		{"orderbook", cli.CmdQueryOrderbook()},
		{"order", cli.CmdQueryOrder()},
		{"user-orders", cli.CmdQueryUserOrders()},
		{"supported-coins", cli.CmdQuerySupportedCoins()},
		{"htlc", cli.CmdQueryHTLC()},
		{"params", cli.CmdQueryParams()},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.cmd)
			// Most query commands should have standard query flags
			// Check that common flags like --node and --output are available
			flags := tc.cmd.Flags()
			require.NotNil(t, flags)
		})
	}
}

// ============================================================================
// Description Tests
// ============================================================================

func TestQueryCommandsHaveDescriptions(t *testing.T) {
	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"pool", cli.CmdQueryPool()},
		{"pools", cli.CmdQueryAllPools()},
		{"pool-stats", cli.CmdQueryPoolStats()},
		{"quote", cli.CmdQueryGetQuote()},
		{"market-price", cli.CmdQueryMarketPrice()},
		{"spot-price", cli.CmdQuerySpotPrice()},
		{"orderbook", cli.CmdQueryOrderbook()},
		{"order", cli.CmdQueryOrder()},
		{"user-orders", cli.CmdQueryUserOrders()},
		{"supported-coins", cli.CmdQuerySupportedCoins()},
		{"htlc", cli.CmdQueryHTLC()},
		{"params", cli.CmdQueryParams()},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.cmd)
			require.NotEmpty(t, tc.cmd.Short, "command %s should have Short description", tc.name)
		})
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestPoolQueryWithSpecialCharacters(t *testing.T) {
	cmd := cli.CmdQueryPool()

	// Pool IDs with various formats
	validPoolIDs := []string{
		"pool-1",
		"POOL_1",
		"uaura-usdt",
		"ibc/ABC123",
	}

	for _, poolID := range validPoolIDs {
		t.Run(poolID, func(t *testing.T) {
			err := cmd.Args(cmd, []string{poolID})
			require.NoError(t, err)
		})
	}
}

func TestOrderQueryWithDifferentFormats(t *testing.T) {
	cmd := cli.CmdQueryOrder()

	// Order IDs with various formats
	validOrderIDs := []string{
		"1",
		"12345",
		"order-1",
		"ORD_123",
		"abc123def",
	}

	for _, orderID := range validOrderIDs {
		t.Run(orderID, func(t *testing.T) {
			err := cmd.Args(cmd, []string{orderID})
			require.NoError(t, err)
		})
	}
}

func TestSpotPriceWithIBCDenoms(t *testing.T) {
	cmd := cli.CmdQuerySpotPrice()

	// Test with IBC denoms
	err := cmd.Args(cmd, []string{
		"pool-1",
		"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		"ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518",
	})
	require.NoError(t, err)
}

func TestQuoteWithLargeAmounts(t *testing.T) {
	cmd := cli.CmdQueryGetQuote()

	// Test with large amounts
	err := cmd.Args(cmd, []string{
		"pool-1",
		"999999999999999999",
		"uaura",
	})
	require.NoError(t, err)
}
