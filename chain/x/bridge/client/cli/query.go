package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

// GetQueryCmd returns the query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "bridge",
		Short:                      "Querying commands for the Bridge module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQuerySharedIdentity(),
		CmdQueryTransfer(),
		CmdQueryAllTransfers(),
		CmdQueryUserTransfers(),
		CmdQueryWrappedToken(),
		CmdQueryAllWrappedTokens(),
		CmdQueryChainConfig(),
		CmdQueryAllChains(),
		CmdQueryCrossChainSwap(),
		CmdQueryBridgeStats(),
		CmdQueryValidators(),
		CmdQueryRelayerStats(),
	)

	return cmd
}

// CmdQuerySharedIdentity queries a shared identity across chains
func CmdQuerySharedIdentity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity [address]",
		Short: "Query shared identity across AURA, PAW, and XAI",
		Long: `Query a user's shared identity and verification status across chains.

Examples:
  aurad query bridge identity aura1abc...
  aurad query bridge identity $(aurad keys show alice -a)

Returns:
  - Linked addresses on PAW and XAI
  - Verification status on each chain
  - AURA IR score
  - Reputation score
  - Cross-chain verification timestamps
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QuerySharedIdentityRequest{
				Address: args[0],
			}

			res, err := queryClient.SharedIdentity(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryTransfer queries a specific cross-chain transfer by ID
func CmdQueryTransfer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer [transfer-id]",
		Short: "Query a cross-chain transfer by ID",
		Long: `Query details of a specific cross-chain transfer.

Examples:
  aurad query bridge transfer transfer-aura1abc...-paw-12345
  aurad query bridge transfer transfer-123

Returns:
  - Transfer ID
  - Source and target chains
  - Sender and recipient addresses
  - Amount and denomination
  - Status (PENDING, CONFIRMED, RELAYED, COMPLETED, FAILED, REFUNDED)
  - Transaction hashes on both chains
  - Validator confirmations
  - Timestamps
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryTransferRequest{
				TransferId: args[0],
			}

			res, err := queryClient.Transfer(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAllTransfers queries all cross-chain transfers
func CmdQueryAllTransfers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfers",
		Short: "List all cross-chain transfers",
		Long: `List all cross-chain transfers, optionally filtered by status.

Examples:
  aurad query bridge transfers
  aurad query bridge transfers --status PENDING
  aurad query bridge transfers --status COMPLETED

Status options: PENDING, CONFIRMED, RELAYED, COMPLETED, FAILED, REFUNDED
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			status, _ := cmd.Flags().GetString("status")

			req := &types.QueryAllTransfersRequest{
				Status: status,
			}

			res, err := queryClient.AllTransfers(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("status", "", "Filter by transfer status")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryUserTransfers queries transfers for a specific user
func CmdQueryUserTransfers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-transfers [address]",
		Short: "List transfers for a specific user",
		Long: `List all cross-chain transfers for a specific user.

Examples:
  aurad query bridge user-transfers aura1abc...
  aurad query bridge user-transfers aura1abc... --chain paw
  aurad query bridge user-transfers $(aurad keys show alice -a)

Options:
  --chain: Filter by source or target chain (paw, xai)
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			chain, _ := cmd.Flags().GetString("chain")

			req := &types.QueryUserTransfersRequest{
				Address: args[0],
				Chain:   chain,
			}

			res, err := queryClient.UserTransfers(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("chain", "", "Filter by chain (paw, xai)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryWrappedToken queries info about a wrapped token
func CmdQueryWrappedToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wrapped-token [wrapped-denom]",
		Short: "Query information about a wrapped token",
		Long: `Query details about a wrapped token from PAW or XAI.

Examples:
  aurad query bridge wrapped-token paw.token
  aurad query bridge wrapped-token xai.coin

Returns:
  - Wrapped denomination on AURA
  - Original denomination on source chain
  - Source chain
  - Total supply of wrapped tokens
  - Locked amount on source chain
  - Decimals
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryWrappedTokenRequest{
				WrappedDenom: args[0],
			}

			res, err := queryClient.WrappedToken(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAllWrappedTokens queries all wrapped tokens
func CmdQueryAllWrappedTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wrapped-tokens",
		Short: "List all wrapped tokens",
		Long: `List all wrapped tokens from PAW and XAI on AURA.

Examples:
  aurad query bridge wrapped-tokens

Returns a list of all wrapped tokens with their details:
  - Wrapped denomination
  - Original denomination
  - Source chain
  - Total supply
  - Locked amount
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAllWrappedTokensRequest{}

			res, err := queryClient.AllWrappedTokens(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryChainConfig queries configuration for a specific chain
func CmdQueryChainConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain-config [chain-id]",
		Short: "Query bridge configuration for a chain",
		Long: `Query bridge configuration and status for a connected chain.

Examples:
  aurad query bridge chain-config paw
  aurad query bridge chain-config xai

Returns:
  - Chain ID and name
  - RPC endpoint
  - Address prefix
  - Bridge contract address
  - Minimum confirmations required
  - Active validators
  - Enabled status
  - Native token denomination
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryChainConfigRequest{
				ChainId: args[0],
			}

			res, err := queryClient.ChainConfig(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryAllChains queries all connected chains
func CmdQueryAllChains() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chains",
		Short: "List all connected chains",
		Long: `List all chains connected to AURA via the bridge.

Examples:
  aurad query bridge chains

Returns a list of all connected chains with their configurations.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAllChainsRequest{}

			res, err := queryClient.AllChains(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryCrossChainSwap queries a cross-chain swap by ID
func CmdQueryCrossChainSwap() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap [swap-id]",
		Short: "Query a cross-chain swap by ID",
		Long: `Query details of a cross-chain swap.

Examples:
  aurad query bridge swap swap-123

Returns:
  - Swap ID
  - Source and target chains
  - Input and output tokens
  - Route (may involve multiple hops)
  - Status
  - Sender and recipient
  - Final amount received
  - Timestamps
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryCrossChainSwapRequest{
				SwapId: args[0],
			}

			res, err := queryClient.CrossChainSwap(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryBridgeStats queries bridge statistics
func CmdQueryBridgeStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Query bridge statistics",
		Long: `Query overall bridge statistics and metrics.

Examples:
  aurad query bridge stats

Returns:
  - Total transfers processed
  - Transfers by status
  - Total volume by chain
  - Total wrapped tokens
  - Active validators
  - Active relayers
  - Average completion time
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryBridgeStatsRequest{}

			res, err := queryClient.BridgeStats(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryValidators queries active bridge validators
func CmdQueryValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "List active bridge validators",
		Long: `List all active validators authorized to relay cross-chain transfers.

Examples:
  aurad query bridge validators

Returns:
  - Validator addresses
  - Public keys
  - Voting power
  - Active status
  - Chains they relay for
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryValidatorsRequest{}

			res, err := queryClient.Validators(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryRelayerStats queries relayer performance statistics
func CmdQueryRelayerStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relayer-stats [relayer-address]",
		Short: "Query relayer performance statistics",
		Long: `Query performance metrics for a specific relayer.

Examples:
  aurad query bridge relayer-stats aura1abc...
  aurad query bridge relayer-stats $(aurad keys show relayer -a)

Returns:
  - Total transfers relayed
  - Successful transfers
  - Failed transfers
  - Total volume processed
  - Last relay timestamp
  - Uptime percentage
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryRelayerStatsRequest{
				RelayerAddress: args[0],
			}

			res, err := queryClient.RelayerStats(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
