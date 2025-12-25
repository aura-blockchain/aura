// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/dex/client/cli"
)

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

func TestCmdQueryMarketPrice(t *testing.T) {
	cmd := cli.CmdQueryMarketPrice()
	require.Equal(t, "market-price [coin]", cmd.Use)

	require.NoError(t, cmd.ValidateArgs([]string{"usdt"}))
	require.Error(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"usdt", "extra"}))
}

func TestCmdQueryUserOrders(t *testing.T) {
	cmd := cli.CmdQueryUserOrders()
	require.Equal(t, "user-orders [address]", cmd.Use)

	require.NoError(t, cmd.ValidateArgs([]string{"aura1xyz"}))
	require.Error(t, cmd.ValidateArgs([]string{}))
	require.Error(t, cmd.ValidateArgs([]string{"aura1", "extra"}))
}

func TestCmdQuerySpotPrice(t *testing.T) {
	cmd := cli.CmdQuerySpotPrice()
	require.Equal(t, "spot-price [pool-id] [base-denom] [quote-denom]", cmd.Use)

	require.NoError(t, cmd.ValidateArgs([]string{"uaura-usdt", "uaura", "usdt"}))
	require.Error(t, cmd.ValidateArgs([]string{"uaura-usdt", "uaura"}))
	require.Error(t, cmd.ValidateArgs([]string{}))
}
