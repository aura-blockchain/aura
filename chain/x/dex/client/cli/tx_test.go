package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTxCmd(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "root command exists",
			run: func(t *testing.T) {
				cmd := GetTxCmd()
				require.NotNil(t, cmd)
				require.Equal(t, "dex", cmd.Use)
				require.Contains(t, cmd.Aliases, "swap")
				require.Contains(t, cmd.Aliases, "exchange")
				require.Equal(t, "DEX transaction subcommands", cmd.Short)
				require.True(t, cmd.DisableFlagParsing)
				require.Equal(t, 2, cmd.SuggestionsMinimumDistance)
			},
		},
		{
			name: "has all subcommands",
			run: func(t *testing.T) {
				cmd := GetTxCmd()
				subcommands := cmd.Commands()
				require.Len(t, subcommands, 10)

				cmdNames := make(map[string]bool)
				for _, subcmd := range subcommands {
					cmdNames[subcmd.Use] = true
				}

				// AMM commands
				require.True(t, cmdNames["create-pool [denom-a] [amount-a] [denom-b] [amount-b]"])
				require.True(t, cmdNames["add-liquidity [pool-id] [denom-a] [amount-a] [denom-b] [amount-b]"])
				require.True(t, cmdNames["remove-liquidity [pool-id] [lp-tokens]"])
				require.True(t, cmdNames["swap [pool-id] [coin-in] [min-amount-out] [max-slippage-bps]"])

				// P2P orderbook commands
				require.True(t, cmdNames["create-order [order-type] [aura-amount] [other-coin] [other-amount]"])
				require.True(t, cmdNames["match-order [order-id]"])
				require.True(t, cmdNames["cancel-order [order-id]"])

				// HTLC commands
				require.True(t, cmdNames["create-htlc [recipient] [amount] [secret-hash] [timelock-seconds]"])
				require.True(t, cmdNames["claim-htlc [htlc-id] [secret]"])
				require.True(t, cmdNames["refund-htlc [htlc-id]"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestCmdCreatePool(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid pool creation",
			args:    []string{"uaura", "1000000", "usdt", "500000"},
			wantErr: false,
		},
		{
			name:    "valid with different pair",
			args:    []string{"uaura", "5000000", "uosmo", "2000000"},
			wantErr: false,
		},
		{
			name:    "missing amount-b",
			args:    []string{"uaura", "1000000", "usdt"},
			wantErr: true,
		},
		{
			name:    "missing denom-b and amounts",
			args:    []string{"uaura", "1000000"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"uaura", "1000000", "usdt", "500000", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdCreatePool()
			require.NotNil(t, cmd)
			require.Equal(t, "create-pool [denom-a] [amount-a] [denom-b] [amount-b]", cmd.Use)
			require.Contains(t, cmd.Short, "Create a new AMM liquidity pool")

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdAddLiquidity(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid add liquidity",
			args:    []string{"uaura-usdt", "uaura", "500000", "usdt", "250000"},
			wantErr: false,
		},
		{
			name:    "valid different pool",
			args:    []string{"uaura-uosmo", "uaura", "1000000", "uosmo", "400000"},
			wantErr: false,
		},
		{
			name:    "missing amount-b",
			args:    []string{"uaura-usdt", "uaura", "100000", "usdt"},
			wantErr: true,
		},
		{
			name:    "missing denom-b and amount-b",
			args:    []string{"uaura-usdt", "uaura", "100000"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"uaura-usdt", "uaura", "100000", "usdt", "50000", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdAddLiquidity()
			require.NotNil(t, cmd)
			require.Equal(t, "add-liquidity [pool-id] [denom-a] [amount-a] [denom-b] [amount-b]", cmd.Use)
			require.Contains(t, cmd.Short, "Add liquidity")

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdRemoveLiquidity(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid remove liquidity",
			args:    []string{"uaura-usdt", "1000000"},
			wantErr: false,
		},
		{
			name:    "valid different pool",
			args:    []string{"uaura-uosmo", "500000"},
			wantErr: false,
		},
		{
			name:    "missing lp amount",
			args:    []string{"uaura-usdt"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"uaura-usdt", "1000", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdRemoveLiquidity()
			require.NotNil(t, cmd)
			require.Equal(t, "remove-liquidity [pool-id] [lp-tokens]", cmd.Use)
			require.Contains(t, cmd.Short, "Remove liquidity")

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdSwap(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid swap",
			args:    []string{"uaura-usdt", "100000uaura", "48000", "500"},
			wantErr: false,
		},
		{
			name:    "valid reverse swap",
			args:    []string{"uaura-usdt", "50000usdt", "95000", "1000"},
			wantErr: false,
		},
		{
			name:    "missing max slippage",
			args:    []string{"uaura-usdt", "100000uaura", "48000"},
			wantErr: true,
		},
		{
			name:    "missing min out and slippage",
			args:    []string{"uaura-usdt", "100000uaura"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"uaura-usdt", "100000uaura", "48000", "500", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdSwap()
			require.NotNil(t, cmd)
			require.Equal(t, "swap [pool-id] [coin-in] [min-amount-out] [max-slippage-bps]", cmd.Use)
			require.Contains(t, cmd.Short, "swap")

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdCreateOrder(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid buy order",
			args:    []string{"buy", "2000000", "usdt", "500000"},
			wantErr: false,
		},
		{
			name:    "valid sell order",
			args:    []string{"sell", "1000000", "usdt", "500000"},
			wantErr: false,
		},
		{
			name:    "missing other amount",
			args:    []string{"buy", "1000000", "usdt"},
			wantErr: true,
		},
		{
			name:    "missing other coin and amount",
			args:    []string{"buy", "1000000"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"buy", "1000000", "usdt", "500000", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdCreateOrder()
			require.NotNil(t, cmd)
			require.Equal(t, "create-order [order-type] [aura-amount] [other-coin] [other-amount]", cmd.Use)
			require.Contains(t, cmd.Short, "Create a P2P swap order")

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdMatchOrder(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid match",
			args:    []string{"order-aura1abc...-12345"},
			wantErr: false,
		},
		{
			name:    "valid match different order",
			args:    []string{"order-aura1def...-67890"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"order-123", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdMatchOrder()
			require.NotNil(t, cmd)
			require.Equal(t, "match-order [order-id]", cmd.Use)
			require.Contains(t, cmd.Short, "Match")

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdCancelOrder(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid cancel",
			args:    []string{"order-123"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"order-123", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdCancelOrder()
			require.NotNil(t, cmd)
			require.Equal(t, "cancel-order [order-id]", cmd.Use)
			require.Contains(t, cmd.Short, "Cancel a pending P2P order")

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdCreateHTLC(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid htlc creation",
			args:    []string{"aura1recipient", "1000000uaura", "abc123def456...", "3600"},
			wantErr: false,
		},
		{
			name:    "valid with different coin",
			args:    []string{"aura1recipient", "500000usdt", "def456ghi789...", "7200"},
			wantErr: false,
		},
		{
			name:    "missing timelock",
			args:    []string{"aura1recipient", "1000000uaura", "abc123hashlock"},
			wantErr: true,
		},
		{
			name:    "missing hash and timelock",
			args:    []string{"aura1recipient", "1000000uaura"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"aura1recipient", "1000000uaura", "abc123", "100", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdCreateHTLC()
			require.NotNil(t, cmd)
			require.Equal(t, "create-htlc [recipient] [amount] [secret-hash] [timelock-seconds]", cmd.Use)
			require.Contains(t, cmd.Short, "Create a Hash Time-Locked Contract")

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdClaimHTLC(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid claim",
			args:    []string{"htlc-123", "secretpreimage"},
			wantErr: false,
		},
		{
			name:    "missing secret",
			args:    []string{"htlc-123"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"htlc-123", "secret", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdClaimHTLC()
			require.NotNil(t, cmd)
			require.Equal(t, "claim-htlc [htlc-id] [secret]", cmd.Use)
			require.Contains(t, cmd.Short, "Claim an HTLC")

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdRefundHTLC(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid refund",
			args:    []string{"htlc-123"},
			wantErr: false,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args",
			args:    []string{"htlc-123", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdRefundHTLC()
			require.NotNil(t, cmd)
			require.Equal(t, "refund-htlc [htlc-id]", cmd.Use)
			require.Contains(t, cmd.Short, "Refund an expired HTLC")

			cmd.SetArgs(tt.args)
			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTxCommandStructure(t *testing.T) {
	cmd := GetTxCmd()

	// Test command properties
	require.Equal(t, "dex", cmd.Use)
	require.Contains(t, cmd.Aliases, "swap")
	require.Contains(t, cmd.Aliases, "exchange")
	require.NotEmpty(t, cmd.Short)
	require.True(t, cmd.DisableFlagParsing)
	require.Equal(t, 2, cmd.SuggestionsMinimumDistance)

	// Test all subcommands are registered
	subcommands := cmd.Commands()
	require.NotEmpty(t, subcommands)
	require.Len(t, subcommands, 10)

	// Verify each subcommand has proper structure
	for _, subcmd := range subcommands {
		require.NotEmpty(t, subcmd.Use, "subcommand Use should not be empty")
		require.NotEmpty(t, subcmd.Short, "subcommand Short should not be empty")
		require.NotNil(t, subcmd.RunE, "subcommand RunE should not be nil")
	}
}
