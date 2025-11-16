package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// GetQueryCmd returns the query commands for the dex module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "dex",
		Short:                      "Querying commands for the DEX module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		// Pool queries
		CmdQueryPools(),
		CmdQueryPool(),
		CmdQueryQuote(),

		// Orderbook queries
		CmdQueryOrders(),
		CmdQueryOrder(),
		CmdQueryOrderbook(),

		// HTLC queries
		CmdQueryHTLCs(),
		CmdQueryHTLC(),

		// Atomic swap queries
		CmdQuerySwaps(),

		// Module params
		CmdQueryParams(),
	)

	return cmd
}

// ============================================================================
// Pool Query Commands
// ============================================================================

// CmdQueryPools lists all liquidity pools
func CmdQueryPools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "List all liquidity pools",
		Long: `Query all AMM liquidity pools in the DEX.

Examples:
  aurad query dex pools
  aurad query dex pools --output json

Returns information about all pools including:
  - Pool ID
  - Token pair (denom A and denom B)
  - Reserves
  - Total LP tokens
  - Fee percentages
  - Volume and swap statistics
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAllPoolsRequest{}

			res, err := queryClient.AllPools(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryPool gets a specific pool by ID
func CmdQueryPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [pool-id]",
		Short: "Query a specific liquidity pool by ID",
		Long: `Query details of a specific AMM liquidity pool.

Examples:
  aurad query dex pool uaura-usdt
  aurad query dex pool uaura-uosmo --output json

Returns detailed information including:
  - Current reserves
  - LP token supply
  - Liquidity providers
  - Fee structure
  - Trading statistics
  - Pool creation time
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryPoolRequest{
				PoolId: args[0],
			}

			res, err := queryClient.Pool(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryQuote gets a swap quote without executing
func CmdQueryQuote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quote [pool-id] [denom-in] [amount-in]",
		Short: "Get a swap quote without executing the trade",
		Long: `Calculate the expected output and price impact for a swap without executing it.

Examples:
  aurad query dex quote uaura-usdt uaura 100000
  aurad query dex quote uaura-uosmo uosmo 50000 --output json

Returns:
  - Estimated output amount
  - Effective price
  - Price impact percentage
  - Fee amount (including IR boost if applicable)

Note: This is a simulation - actual swap results may vary slightly.
Verified users (100+ IR points) will see 40% higher fees in their simulations.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			// Parse amount
			amountIn, ok := sdk.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid amount-in: %s", args[2])
			}

			req := &types.QueryGetQuoteRequest{
				PoolId:   args[0],
				DenomIn:  args[1],
				AmountIn: amountIn,
			}

			res, err := queryClient.GetQuote(context.Background(), req)
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
// Orderbook Query Commands
// ============================================================================

// CmdQueryOrders lists all orders
func CmdQueryOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orders",
		Short: "List all P2P swap orders",
		Long: `Query all P2P swap orders in the orderbook.

Examples:
  aurad query dex orders
  aurad query dex orders --status pending
  aurad query dex orders --user aura1abc... --output json

Flags:
  --status: Filter by status (pending, matched, completed, cancelled)
  --user: Filter by user address

Returns all matching orders with their details.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			// Get optional user filter
			userAddr, _ := cmd.Flags().GetString("user")
			if userAddr != "" {
				req := &types.QueryUserOrdersRequest{
					Address: userAddr,
				}

				res, err := queryClient.UserOrders(context.Background(), req)
				if err != nil {
					return err
				}

				return clientCtx.PrintProto(res)
			}

			// List all orders (would need to implement in proto/keeper)
			// For now, return an informative message
			return fmt.Errorf("listing all orders not yet implemented - use --user flag to query specific user's orders")
		},
	}

	cmd.Flags().String("status", "", "Filter by order status (pending, matched, completed, cancelled)")
	cmd.Flags().String("user", "", "Filter by user address")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryOrder gets a specific order by ID
func CmdQueryOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order [order-id]",
		Short: "Query a specific P2P order by ID",
		Long: `Query details of a specific P2P swap order.

Examples:
  aurad query dex order order-aura1abc...-12345
  aurad query dex order order-aura1def...-67890 --output json

Returns:
  - Order ID
  - Order type (buy/sell)
  - AURA amount
  - Other coin and amount
  - Status
  - Creator and matcher addresses
  - Creation and expiration times
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryOrderRequest{
				OrderId: args[0],
			}

			res, err := queryClient.Order(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryOrderbook gets the orderbook for a trading pair
func CmdQueryOrderbook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orderbook [trading-pair]",
		Short: "Query the orderbook for a trading pair",
		Long: `Query all pending orders for a specific trading pair.

Examples:
  aurad query dex orderbook AURA/USDT
  aurad query dex orderbook AURA/OSMO --output json

The trading pair should be in format: TOKEN1/TOKEN2

Returns:
  - All pending buy orders
  - All pending sell orders
  - Best bid and ask prices
  - Order depth
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryOrderbookRequest{
				Pair: args[0],
			}

			res, err := queryClient.Orderbook(context.Background(), req)
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

// CmdQueryHTLCs lists all HTLCs
func CmdQueryHTLCs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "htlcs",
		Short: "List all Hash Time-Locked Contracts",
		Long: `Query all HTLCs in the system.

Examples:
  aurad query dex htlcs
  aurad query dex htlcs --status pending
  aurad query dex htlcs --sender aura1abc... --output json

Flags:
  --status: Filter by status (pending, completed, refunded)
  --sender: Filter by sender address
  --recipient: Filter by recipient address

Returns all matching HTLCs with their details.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// This would need a custom query implementation
			// For now, return an informative message
			return fmt.Errorf("listing all HTLCs not yet implemented - use 'query dex htlc [htlc-id]' for specific HTLC")
		},
	}

	cmd.Flags().String("status", "", "Filter by HTLC status (pending, completed, refunded)")
	cmd.Flags().String("sender", "", "Filter by sender address")
	cmd.Flags().String("recipient", "", "Filter by recipient address")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryHTLC gets a specific HTLC by ID
func CmdQueryHTLC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "htlc [htlc-id]",
		Short: "Query a specific HTLC by ID",
		Long: `Query details of a specific Hash Time-Locked Contract.

Examples:
  aurad query dex htlc htlc-aura1abc...-def123
  aurad query dex htlc htlc-aura1def...-ghi456 --output json

Returns:
  - HTLC ID
  - Sender and recipient addresses
  - Locked amount
  - Secret hash
  - Timelock expiration
  - Status (pending, completed, refunded)
  - Secret (if claimed)
  - Timestamps
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryHTLCRequest{
				HtlcId: args[0],
			}

			res, err := queryClient.HTLC(context.Background(), req)
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
// Atomic Swap Query Commands
// ============================================================================

// CmdQuerySwaps lists atomic swaps
func CmdQuerySwaps() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swaps",
		Short: "List atomic swaps",
		Long: `Query atomic swap records in the system.

Examples:
  aurad query dex swaps
  aurad query dex swaps --initiator aura1abc...
  aurad query dex swaps --status completed --output json

Flags:
  --status: Filter by status (initiated, completed, failed)
  --initiator: Filter by initiator address
  --counterparty: Filter by counterparty address

Returns all matching atomic swaps with their details.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// This would need a custom query implementation
			// For now, return an informative message
			return fmt.Errorf("listing atomic swaps not yet implemented in proto/keeper")
		},
	}

	cmd.Flags().String("status", "", "Filter by swap status (initiated, completed, failed)")
	cmd.Flags().String("initiator", "", "Filter by initiator address")
	cmd.Flags().String("counterparty", "", "Filter by counterparty address")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// ============================================================================
// Module Parameters
// ============================================================================

// CmdQueryParams queries the module parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query DEX module parameters",
		Long: `Query the current parameters of the DEX module.

Examples:
  aurad query dex params
  aurad query dex params --output json

Returns:
  - Trading fee percentage
  - Protocol fee percentage
  - Minimum liquidity tiers (dynamic based on AURA price)
  - IR boost settings
  - Authority/governance configuration
  - Supported altcoins
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// This would need a params query implementation
			// For now, return an informative message
			return fmt.Errorf("params query not yet implemented in proto - needs QueryParamsRequest/Response")
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
