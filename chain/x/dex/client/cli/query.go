package cli

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/chain/x/dex/types"
	dexv1beta1 "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

// GetQueryCmd returns the query commands for the dex module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "dex",
		Short:                      "Querying commands for the dex module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		// Pool queries
		CmdQueryPool(),
		CmdQueryAllPools(),
		CmdQueryPoolStats(),

		// Quote and price queries
		CmdQueryGetQuote(),
		CmdQueryMarketPrice(),
		CmdQuerySpotPrice(),

		// Orderbook queries
		CmdQueryOrderbook(),
		CmdQueryOrder(),
		CmdQueryUserOrders(),

		// System queries
		CmdQuerySupportedCoins(),

		// HTLC queries
		CmdQueryHTLC(),
	)

	return cmd
}

// ============================================================================
// Pool Query Commands
// ============================================================================

// CmdQueryPool queries a liquidity pool by ID
func CmdQueryPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [pool-id]",
		Short: "Query a liquidity pool by ID",
		Long: `Query detailed information about a specific AMM liquidity pool.

Examples:
  aurad query dex pool uaura-usdt
  aurad query dex pool uaura-uosmo

Returns:
  - Pool ID and token denominations
  - Current reserves for both tokens
  - Total LP tokens issued
  - Fee percentages
  - Trading statistics
  - List of liquidity providers
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dexv1beta1.NewQueryClient(clientCtx)

			req := &types.QueryPoolRequest{
				PoolId: args[0],
			}

			res, err := queryClient.Pool(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAllPools queries all liquidity pools
func CmdQueryAllPools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "Query all liquidity pools",
		Long: `Query all active AMM liquidity pools in the DEX.

Examples:
  aurad query dex pools

Returns:
  List of all pools with:
  - Pool IDs and token pairs
  - Current reserves
  - Total LP tokens
  - Trading statistics
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dexv1beta1.NewQueryClient(clientCtx)

			req := &types.QueryAllPoolsRequest{}

			res, err := queryClient.AllPools(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryPoolStats queries pool statistics
func CmdQueryPoolStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool-stats [pool-id]",
		Short: "Query statistics for a liquidity pool",
		Long: `Query detailed trading statistics for a specific pool.

Examples:
  aurad query dex pool-stats uaura-usdt
  aurad query dex pool-stats uaura-uosmo

Returns:
  - Current reserves and total LP tokens
  - Number of liquidity providers
  - Current price
  - Total trading volume
  - Total fees collected
  - Number of swaps executed
  - Total Value Locked (TVL) for both tokens
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dexv1beta1.NewQueryClient(clientCtx)

			req := &types.QueryPoolStatsRequest{
				PoolId: args[0],
			}

			res, err := queryClient.PoolStats(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Quote and Price Query Commands
// ============================================================================

// CmdQueryGetQuote gets a swap quote without executing
func CmdQueryGetQuote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quote [pool-id] [denom-in] [amount-in]",
		Short: "Get a swap quote without executing the trade",
		Long: `Get an estimated output amount and price information for a potential swap.

Examples:
  aurad query dex quote uaura-usdt uaura 1000000
  aurad query dex quote uaura-uosmo uosmo 500000

Returns:
  - Estimated output amount
  - Effective price
  - Price impact percentage
  - Fee amount

Note: This is a read-only operation that doesn't execute the swap.
Use this to check prices before submitting a swap transaction.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Parse amount
			amountIn, ok := sdkmath.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid amount-in: %s", args[2])
			}

			queryClient := dexv1beta1.NewQueryClient(clientCtx)

			req := &types.QueryGetQuoteRequest{
				PoolId:   args[0],
				DenomIn:  args[1],
				AmountIn: amountIn.String(),
			}

			res, err := queryClient.GetQuote(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryMarketPrice queries current market price for a coin
func CmdQueryMarketPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market-price [coin]",
		Short: "Query current market price for a coin",
		Long: `Query the current market price for a specific coin based on recent swap history.

Examples:
  aurad query dex market-price usdt
  aurad query dex market-price btc
  aurad query dex market-price eth

Supported coins:
  usdt, usdc, dai, btc, eth, ltc, doge, bch, xmr, zec, dash,
  osmo, atom, paw, xai

Returns:
  - Coin symbol
  - Price in USD (reference)
  - Price in AURA
  - Last updated timestamp
  - Sample size (number of recent swaps)
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dexv1beta1.NewQueryClient(clientCtx)

			req := &types.QueryMarketPriceRequest{
				Coin: args[0],
			}

			res, err := queryClient.MarketPrice(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQuerySpotPrice calculates the instantaneous spot price within a pool.
func CmdQuerySpotPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spot-price [pool-id] [base-denom] [quote-denom]",
		Short: "Compute spot price between two denoms in a pool",
		Long: `Compute the instantaneous price for a specific trading pair inside a liquidity pool.

This command derives the price from current pool reserves (no swap is executed).`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			poolID := args[0]
			base := args[1]
			quote := args[2]

			queryClient := dexv1beta1.NewQueryClient(clientCtx)
			resp, err := queryClient.SpotPrice(cmd.Context(), &types.QuerySpotPriceRequest{
				PoolId:     poolID,
				BaseDenom:  base,
				QuoteDenom: quote,
			})
			if err != nil {
				return err
			}

			result := struct {
				PoolID     string `json:"pool_id"`
				BaseDenom  string `json:"base_denom"`
				QuoteDenom string `json:"quote_denom"`
				Price      string `json:"price"`
			}{
				PoolID:     poolID,
				BaseDenom:  strings.ToLower(base),
				QuoteDenom: strings.ToLower(quote),
				Price:      resp.Price,
			}

			return clientCtx.PrintObjectLegacy(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Orderbook Query Commands
// ============================================================================

// CmdQueryOrderbook queries the P2P orderbook for a trading pair
func CmdQueryOrderbook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orderbook [pair]",
		Short: "Query the P2P orderbook for a trading pair",
		Long: `Query the peer-to-peer orderbook showing all active buy and sell orders.

Examples:
  aurad query dex orderbook AURA/USDT
  aurad query dex orderbook AURA/BTC
  aurad query dex orderbook AURA/ETH

Returns:
  - Trading pair
  - Buy orders (sorted by price descending)
  - Sell orders (sorted by price ascending)
  - Total pending orders
  - Best bid and ask prices
  - Spread percentage
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dexv1beta1.NewQueryClient(clientCtx)

			req := &types.QueryOrderbookRequest{
				Pair: args[0],
			}

			res, err := queryClient.Orderbook(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryOrder queries a specific order by ID
func CmdQueryOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order [order-id]",
		Short: "Query a specific order by ID",
		Long: `Query detailed information about a specific P2P swap order.

Examples:
  aurad query dex order order-aura1abc...-12345
  aurad query dex order order-aura1def...-67890

Returns:
  - Order ID and type (buy/sell)
  - AURA amount and other coin amount
  - User address
  - Creation timestamp
  - Current status
  - Matched order (if applicable)
  - HTLC data (if applicable)
  - Price per AURA
  - Expiration time
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dexv1beta1.NewQueryClient(clientCtx)

			req := &types.QueryOrderRequest{
				OrderId: args[0],
			}

			res, err := queryClient.Order(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryUserOrders queries all orders for a user
func CmdQueryUserOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-orders [address]",
		Short: "Query all orders for a specific user",
		Long: `Query all P2P swap orders created by a specific user address.

Examples:
  aurad query dex user-orders aura1abc...
  aurad query dex user-orders aura1def...

Returns:
  List of all orders (pending, matched, completed, cancelled, expired) with:
  - Order IDs and types
  - Amounts and prices
  - Status and timestamps
  - Matched orders
  - HTLC data
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dexv1beta1.NewQueryClient(clientCtx)

			req := &types.QueryUserOrdersRequest{
				Address: args[0],
			}

			res, err := queryClient.UserOrders(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// System Query Commands
// ============================================================================

// CmdQuerySupportedCoins queries the list of supported coins
func CmdQuerySupportedCoins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supported-coins",
		Short: "Query the list of supported altcoins",
		Long: `Query the list of all supported altcoins for trading on the DEX.

Examples:
  aurad query dex supported-coins

Returns:
  List of supported coin denominations including:
  - Stablecoins: usdt, usdc, dai
  - Major cryptos: btc, eth, ltc, doge, bch, xmr, zec, dash
  - Cosmos ecosystem: osmo, atom
  - Cross-chain: paw, xai
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dexv1beta1.NewQueryClient(clientCtx)

			req := &types.QuerySupportedCoinsRequest{}

			res, err := queryClient.SupportedCoins(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// HTLC Query Commands
// ============================================================================

// CmdQueryHTLC queries a Hash Time-Locked Contract by ID
func CmdQueryHTLC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "htlc [htlc-id]",
		Short: "Query a Hash Time-Locked Contract by ID",
		Long: `Query detailed information about a specific HTLC for atomic swaps.

Examples:
  aurad query dex htlc htlc-aura1abc...-def123
  aurad query dex htlc htlc-aura1def...-ghi456

Returns:
  - HTLC ID
  - Secret hash (and secret if revealed)
  - Timelock expiration
  - Status (pending, claimed, refunded, expired)
  - Sender and recipient addresses
  - Locked amount and denomination
  - Refund address

Use this to:
  - Check HTLC status during atomic swap
  - Verify timelock hasn't expired before claiming
  - Monitor cross-chain swap progress
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := dexv1beta1.NewQueryClient(clientCtx)

			req := &types.QueryHTLCRequest{
				HtlcId: args[0],
			}

			res, err := queryClient.HTLC(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
