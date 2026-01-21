// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// GetTxCmd returns the transaction commands for the dex module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "dex",
		Aliases:                    []string{"swap", "exchange"},
		Short:                      "DEX transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		// AMM Liquidity Pool commands
		CmdCreatePool(),
		CmdAddLiquidity(),
		CmdRemoveLiquidity(),
		CmdSwap(),

		// P2P Orderbook commands
		CmdCreateOrder(),
		CmdMatchOrder(),
		CmdCancelOrder(),

		// HTLC Atomic Swap commands
		CmdCreateHTLC(),
		CmdClaimHTLC(),
		CmdRefundHTLC(),
	)

	return cmd
}

// ============================================================================
// AMM Liquidity Pool Commands
// ============================================================================

// CmdCreatePool creates a new liquidity pool
func CmdCreatePool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [denom-a] [amount-a] [denom-b] [amount-b]",
		Short: "Create a new AMM liquidity pool",
		Long: `Create a new AMM liquidity pool with initial liquidity.

Examples:
  aurad tx dex create-pool uaura 1000000 usdt 500000 --from alice
  aurad tx dex create-pool uaura 5000000 uosmo 2000000 --from alice

Note: Initial liquidity must meet minimum requirements based on current AURA price.
The pool ID will be automatically generated from the token pair (alphabetically ordered).
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse amounts
			amountA, err := sdk.ParseCoinNormalized(args[1] + args[0])
			if err != nil {
				return fmt.Errorf("invalid amount-a: %w", err)
			}

			amountB, err := sdk.ParseCoinNormalized(args[3] + args[2])
			if err != nil {
				return fmt.Errorf("invalid amount-b: %w", err)
			}

			msg := &types.MsgCreatePool{
				Creator: clientCtx.GetFromAddress().String(),
				DenomA:  args[0],
				DenomB:  args[2],
				AmountA: amountA,
				AmountB: amountB,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdAddLiquidity adds liquidity to an existing pool
func CmdAddLiquidity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-liquidity [pool-id] [denom-a] [amount-a] [denom-b] [amount-b]",
		Short: "Add liquidity to an existing pool",
		Long: `Add liquidity to an existing AMM pool and receive LP tokens.

Examples:
  aurad tx dex add-liquidity uaura-usdt uaura 500000 usdt 250000 --from alice
  aurad tx dex add-liquidity uaura-uosmo uaura 1000000 uosmo 400000 --from bob

Note: Amounts will be adjusted to match the current pool ratio.
Existing liquidity providers are grandfathered and can add any amount.
New liquidity providers must meet current minimum requirements.
`,
		Args: cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse amounts
			amountA, err := sdk.ParseCoinNormalized(args[2] + args[1])
			if err != nil {
				return fmt.Errorf("invalid amount-a: %w", err)
			}

			amountB, err := sdk.ParseCoinNormalized(args[4] + args[3])
			if err != nil {
				return fmt.Errorf("invalid amount-b: %w", err)
			}

			msg := &types.MsgAddLiquidity{
				Provider: clientCtx.GetFromAddress().String(),
				PoolId:   args[0],
				AmountA:  amountA,
				AmountB:  amountB,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRemoveLiquidity removes liquidity from a pool
func CmdRemoveLiquidity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-liquidity [pool-id] [lp-tokens]",
		Short: "Remove liquidity from a pool",
		Long: `Remove liquidity from an AMM pool by burning LP tokens.

Examples:
  aurad tx dex remove-liquidity uaura-usdt 1000000 --from alice
  aurad tx dex remove-liquidity uaura-uosmo 500000 --from bob

Note: You will receive tokens proportional to your share of the pool.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse LP tokens amount
			lpTokens, ok := sdkmath.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid lp-tokens amount: %s", args[1])
			}

			msg := &types.MsgRemoveLiquidity{
				Provider: clientCtx.GetFromAddress().String(),
				PoolId:   args[0],
				LpTokens: lpTokens,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdSwap executes a token swap
func CmdSwap() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap [pool-id] [coin-in] [min-amount-out] [max-slippage-bps]",
		Short: "Execute a token swap in AMM pool",
		Long: `Execute a token swap using the constant product AMM formula.

Examples:
  aurad tx dex swap uaura-usdt 100000uaura 48000 500 --from alice
  aurad tx dex swap uaura-uosmo 50000uosmo 120000 1000 --from bob

Arguments:
  pool-id: The liquidity pool ID (e.g., uaura-usdt)
  coin-in: Amount and denomination to swap (e.g., 100000uaura)
  min-amount-out: Minimum output amount (slippage protection)
  max-slippage-bps: Maximum allowed price impact in basis points (500 = 5%)

Note: Verified users (100+ IR points) earn 40% more fees!
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse coin in
			coinIn, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid coin-in: %w", err)
			}

			// Parse min amount out
			minAmountOut, ok := sdkmath.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid min-amount-out: %s", args[2])
			}

			// Parse max slippage bps
			maxSlippageBps, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid max-slippage-bps: %w", err)
			}

			msg := &types.MsgSwapExactIn{
				Sender:         clientCtx.GetFromAddress().String(),
				PoolId:         args[0],
				CoinIn:         coinIn,
				MinAmountOut:   minAmountOut,
				MaxSlippageBps: maxSlippageBps,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// P2P Orderbook Commands
// ============================================================================

// CmdCreateOrder creates a new P2P swap order
func CmdCreateOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-order [order-type] [aura-amount] [other-coin] [other-amount]",
		Short: "Create a P2P swap order in the orderbook",
		Long: `Create a peer-to-peer swap order that can be matched by other users.

Examples:
  aurad tx dex create-order sell 1000000 usdt 500000 --from alice
  aurad tx dex create-order buy 2000000 btc 5000 --from bob

Arguments:
  order-type: "buy" or "sell" (AURA)
  aura-amount: Amount of AURA tokens
  other-coin: Denomination of the other token (usdt, btc, eth, etc.)
  other-amount: Amount of the other token

Note: Funds will be locked until the order is matched or cancelled.
Orders expire after 24 hours by default.
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse order type
			var orderType types.SwapOrderType
			switch args[0] {
			case "sell":
				orderType = types.SwapOrderType_SELL
			case "buy":
				orderType = types.SwapOrderType_BUY
			default:
				return fmt.Errorf("invalid order-type: must be 'buy' or 'sell'")
			}

			// Parse amounts
			auraAmount, ok := sdkmath.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid aura-amount: %s", args[1])
			}

			otherAmount, ok := sdkmath.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("invalid other-amount: %s", args[3])
			}

			msg := &types.MsgCreateOrder{
				Creator:     clientCtx.GetFromAddress().String(),
				OrderType:   orderType,
				AuraAmount:  auraAmount,
				OtherCoin:   args[2],
				OtherAmount: otherAmount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdMatchOrder matches an existing order
func CmdMatchOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "match-order [order-id]",
		Short: "Match and execute an existing P2P order",
		Long: `Match an existing order in the orderbook and execute the swap.

Examples:
  aurad tx dex match-order order-aura1abc...-12345 --from bob
  aurad tx dex match-order order-aura1def...-67890 --from alice

Note: You must have sufficient funds to match the order.
The swap will be executed immediately upon matching.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgExecuteSwap{
				Initiator: clientCtx.GetFromAddress().String(),
				OrderId:   args[0],
				Secret:    "", // Not used for simple P2P matching
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdCancelOrder cancels a pending order
func CmdCancelOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-order [order-id]",
		Short: "Cancel a pending P2P order",
		Long: `Cancel your pending order and unlock the locked funds.

Examples:
  aurad tx dex cancel-order order-aura1abc...-12345 --from alice
  aurad tx dex cancel-order order-aura1def...-67890 --from bob

Note: Only pending orders can be cancelled.
Locked funds will be returned to your account.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgCancelOrder{
				Creator: clientCtx.GetFromAddress().String(),
				OrderId: args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// HTLC Atomic Swap Commands
// ============================================================================

// CmdCreateHTLC creates a Hash Time-Locked Contract
func CmdCreateHTLC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-htlc [recipient] [amount] [secret-hash] [timelock-seconds]",
		Short: "Create a Hash Time-Locked Contract for atomic swaps",
		Long: `Create an HTLC for trustless cross-chain atomic swaps.

Examples:
  aurad tx dex create-htlc aura1def... 1000000uaura abc123def456... 3600 --from alice
  aurad tx dex create-htlc aura1ghi... 500000usdt 789abc012def... 7200 --from bob

Arguments:
  recipient: Address that can claim the HTLC with the secret
  amount: Amount to lock (e.g., 1000000uaura)
  secret-hash: SHA-256 hash of the secret (hex encoded)
  timelock-seconds: Lock duration in seconds

Workflow:
  1. Alice creates HTLC on Chain A with secret hash
  2. Bob creates HTLC on Chain B with same secret hash
  3. Alice claims Bob's HTLC, revealing the secret
  4. Bob uses revealed secret to claim Alice's HTLC
  5. If timeout expires, funds are refunded
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse amount
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}

			// Parse timelock duration
			timelockDuration, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid timelock-seconds: %w", err)
			}

			msg := &types.MsgCreateHTLC{
				Sender:           clientCtx.GetFromAddress().String(),
				Recipient:        args[0],
				Amount:           amount,
				SecretHash:       args[2],
				TimelockDuration: timelockDuration,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdClaimHTLC claims an HTLC by revealing the secret
func CmdClaimHTLC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-htlc [htlc-id] [secret]",
		Short: "Claim an HTLC by revealing the secret preimage",
		Long: `Claim funds from an HTLC by providing the secret that matches the hash.

Examples:
  aurad tx dex claim-htlc htlc-aura1abc...-def123 mysecretpreimage123 --from bob
  aurad tx dex claim-htlc htlc-aura1def...-ghi456 anothersecret456 --from alice

Note: The secret must hash to the secret-hash used when creating the HTLC.
The HTLC must not be expired.
Only the designated recipient can claim the HTLC.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgClaimHTLC{
				Recipient: clientCtx.GetFromAddress().String(),
				HtlcId:    args[0],
				Secret:    args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdRefundHTLC refunds an expired HTLC
func CmdRefundHTLC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refund-htlc [htlc-id]",
		Short: "Refund an expired HTLC back to the sender",
		Long: `Refund funds from an expired HTLC back to the original sender.

Examples:
  aurad tx dex refund-htlc htlc-aura1abc...-def123 --from alice
  aurad tx dex refund-htlc htlc-aura1def...-ghi456 --from bob

Note: The HTLC must have expired (past the timelock).
Only the original sender can refund the HTLC.
If the HTLC was claimed before expiry, refund will fail.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgRefundHTLC{
				Sender: clientCtx.GetFromAddress().String(),
				HtlcId: args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
