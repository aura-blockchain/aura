package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	bridgev1beta1 "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "bridge",
		Short:                      "Querying commands for the bridge module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryTransfer(),
		CmdQueryAllTransfers(),
		CmdQueryUserTransfers(),
		CmdQueryChainConfig(),
		CmdQueryAllChains(),
		CmdQueryWrappedToken(),
		CmdQueryAllWrappedTokens(),
		CmdQuerySharedIdentity(),
		CmdQueryCrossChainSwap(),
		CmdQueryBridgeStats(),
		CmdQueryValidators(),
		CmdQueryRelayerStats(),
	)

	return cmd
}

// CmdQueryTransfer queries a cross-chain transfer by ID
func CmdQueryTransfer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer [transfer-id]",
		Short: "Query a cross-chain transfer by ID",
		Long: `Query details of a specific cross-chain transfer using its transfer ID.

Examples:
  aurad query bridge transfer transfer-123
  aurad query bridge transfer 0x1a2b3c4d5e6f...

Returns:
  - Transfer ID
  - Source and target chains
  - Sender and recipient addresses
  - Amount and denomination
  - Transfer status
  - Confirmations
  - Validator signatures
  - Timestamps
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := bridgev1beta1.NewQueryClient(clientCtx)

			req := &bridgev1beta1.QueryTransferRequest{
				TransferId: args[0],
			}

			res, err := queryClient.Transfer(cmd.Context(), req)
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
		Short: "Query all cross-chain transfers",
		Long: `Query all cross-chain transfers with optional status filter.

Examples:
  aurad query bridge transfers
  aurad query bridge transfers --status PENDING
  aurad query bridge transfers --status COMPLETED

Status options: PENDING, CONFIRMED, RELAYED, COMPLETED, FAILED, REFUNDED

Returns a list of all transfers matching the criteria.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := bridgev1beta1.NewQueryClient(clientCtx)

			status, _ := cmd.Flags().GetString("status")

			req := &bridgev1beta1.QueryAllTransfersRequest{
				Status: status,
			}

			res, err := queryClient.AllTransfers(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("status", "", "Filter transfers by status (PENDING, CONFIRMED, RELAYED, COMPLETED, FAILED, REFUNDED)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryUserTransfers queries transfers for a specific user
func CmdQueryUserTransfers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-transfers [address]",
		Short: "Query transfers for a specific user",
		Long: `Query all cross-chain transfers for a specific user address.

Examples:
  aurad query bridge user-transfers aura1abc...
  aurad query bridge user-transfers paw1def... --chain paw
  aurad query bridge user-transfers xai1ghi... --chain xai

Options:
  --chain: Filter transfers by specific chain (aura, paw, xai)

Returns all transfers where the user is sender or recipient.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := bridgev1beta1.NewQueryClient(clientCtx)

			chain, _ := cmd.Flags().GetString("chain")

			req := &bridgev1beta1.QueryUserTransfersRequest{
				Address: args[0],
				Chain:   chain,
			}

			res, err := queryClient.UserTransfers(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("chain", "", "Filter by specific chain (aura, paw, xai)")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryChainConfig queries configuration for a specific chain
func CmdQueryChainConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain-config [chain-id]",
		Short: "Query configuration for a specific chain",
		Long: `Query bridge configuration for a connected chain.

Examples:
  aurad query bridge chain-config aura
  aurad query bridge chain-config paw
  aurad query bridge chain-config xai

Returns:
  - Chain ID and name
  - RPC endpoint
  - Address prefix
  - Bridge contract address
  - Minimum confirmations
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

			queryClient := bridgev1beta1.NewQueryClient(clientCtx)

			req := &bridgev1beta1.QueryChainConfigRequest{
				ChainId: args[0],
			}

			res, err := queryClient.ChainConfig(cmd.Context(), req)
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
		Short: "Query all connected chains",
		Long: `Query configuration for all chains connected to the bridge.

Example:
  aurad query bridge chains

Returns a list of all configured chains (AURA, PAW, XAI) with their settings.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := bridgev1beta1.NewQueryClient(clientCtx)

			req := &bridgev1beta1.QueryAllChainsRequest{}

			res, err := queryClient.AllChains(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryWrappedToken queries information about a wrapped token
func CmdQueryWrappedToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wrapped-token [wrapped-denom]",
		Short: "Query information about a wrapped token",
		Long: `Query details of a wrapped token on AURA.

Examples:
  aurad query bridge wrapped-token paw.token
  aurad query bridge wrapped-token xai.coin

Returns:
  - Wrapped denomination on AURA
  - Original denomination on source chain
  - Source chain
  - Total supply of wrapped tokens
  - Decimals
  - Locked amount on source chain
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := bridgev1beta1.NewQueryClient(clientCtx)

			req := &bridgev1beta1.QueryWrappedTokenRequest{
				WrappedDenom: args[0],
			}

			res, err := queryClient.WrappedToken(cmd.Context(), req)
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
		Short: "Query all wrapped tokens",
		Long: `Query all tokens from PAW and XAI that are wrapped on AURA.

Example:
  aurad query bridge wrapped-tokens

Returns a list of all wrapped tokens with their mappings and supply information.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := bridgev1beta1.NewQueryClient(clientCtx)

			req := &bridgev1beta1.QueryAllWrappedTokensRequest{}

			res, err := queryClient.AllWrappedTokens(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQuerySharedIdentity queries shared identity across chains
func CmdQuerySharedIdentity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shared-identity [address]",
		Short: "Query shared identity across AURA, PAW, and XAI",
		Long: `Query shared identity information for an address across all chains.

Examples:
  aurad query bridge shared-identity aura1abc...
  aurad query bridge shared-identity paw1def...
  aurad query bridge shared-identity xai1ghi...

Returns:
  - Verification status on each chain
  - AURA IR score
  - Linked addresses on other chains
  - Reputation score
  - Verification timestamp

This allows checking if an address on one chain has verified identities on other chains.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := bridgev1beta1.NewQueryClient(clientCtx)

			req := &bridgev1beta1.QuerySharedIdentityRequest{
				Address: args[0],
			}

			res, err := queryClient.SharedIdentity(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryCrossChainSwap queries cross-chain swap status
func CmdQueryCrossChainSwap() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cross-chain-swap [swap-id]",
		Short: "Query cross-chain swap status",
		Long: `Query the status and details of a cross-chain swap.

Examples:
  aurad query bridge cross-chain-swap swap-123
  aurad query bridge cross-chain-swap 0x1a2b3c...

Returns:
  - Swap ID
  - Source chain and token
  - Target chain and token
  - Sender and recipient
  - Swap route (chains involved)
  - Status
  - Initiated and completion timestamps
  - Final amount received
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := bridgev1beta1.NewQueryClient(clientCtx)

			req := &bridgev1beta1.QueryCrossChainSwapRequest{
				SwapId: args[0],
			}

			res, err := queryClient.CrossChainSwap(cmd.Context(), req)
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

Example:
  aurad query bridge stats

Returns:
  - Total transfers processed
  - Transfers by status (pending, completed, failed, etc.)
  - Total volume by chain
  - Total wrapped tokens
  - Active validators count
  - Active relayers count
  - Average completion time

Useful for monitoring bridge health and performance.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := bridgev1beta1.NewQueryClient(clientCtx)

			req := &bridgev1beta1.QueryBridgeStatsRequest{}

			res, err := queryClient.BridgeStats(cmd.Context(), req)
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
		Short: "Query active bridge validators",
		Long: `Query all validators authorized to relay cross-chain transfers.

Example:
  aurad query bridge validators

Returns:
  - Validator address
  - Public key
  - Voting power/weight
  - Active status
  - Chains the validator relays for

Only active validators can sign cross-chain transfer proofs.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := bridgev1beta1.NewQueryClient(clientCtx)

			req := &bridgev1beta1.QueryValidatorsRequest{}

			res, err := queryClient.Validators(cmd.Context(), req)
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
		Long: `Query performance statistics for a specific bridge relayer.

Examples:
  aurad query bridge relayer-stats aura1relayer...

Returns:
  - Relayer address
  - Total transfers relayed
  - Successful transfers
  - Failed transfers
  - Total volume relayed
  - Last relay timestamp
  - Uptime percentage

Useful for monitoring relayer performance and reliability.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := bridgev1beta1.NewQueryClient(clientCtx)

			req := &bridgev1beta1.QueryRelayerStatsRequest{
				RelayerAddress: args[0],
			}

			res, err := queryClient.RelayerStats(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// FormatTransferStatus formats transfer status for display
func FormatTransferStatus(status int32) string {
	statuses := map[int32]string{
		0: "PENDING",
		1: "CONFIRMED",
		2: "RELAYED",
		3: "COMPLETED",
		4: "FAILED",
		5: "REFUNDED",
	}
	if s, ok := statuses[status]; ok {
		return s
	}
	return fmt.Sprintf("UNKNOWN(%d)", status)
}

// FormatChainList formats a list of chains for display
func FormatChainList(chains []string) string {
	if len(chains) == 0 {
		return "none"
	}
	return strings.Join(chains, ", ")
}
