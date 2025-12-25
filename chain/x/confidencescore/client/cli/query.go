// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

// GetQueryCmd returns the cli query commands for the confidencescore module
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        "confidencescore",
		Short:                      "Querying commands for the confidencescore module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		CmdQueryUserScore(),
		CmdQueryUserCompletions(),
		CmdQueryScoreHistory(),
		CmdQueryThresholds(),
		CmdQueryVerifiedUsers(),
		CmdQueryArenaBreakdown(),
		CmdQuerySlashRecords(),
		CmdQueryParams(),
		CmdQueryIRCompletion(),
	)

	return queryCmd
}

// CmdQueryUserScore queries a user's confidence score and status
func CmdQueryUserScore() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "score [wallet-address]",
		Short: "Query a user's confidence score and verification status",
		Long: `Query the confidence score, verification status, and basic statistics for a user.

Example:
  $ aurad query confidencescore score aura1abc...

This returns:
- Total confidence score
- Verification status (verified if >= 10,000 CS)
- Anchor (IR-000) completion info
- Arena score breakdown
- Number of IRs completed
- Last update timestamp`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryUserScoreRequest{
				WalletAddress: args[0],
			}

			res, err := queryClient.UserScore(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryUserCompletions queries a user's IR completion history
func CmdQueryUserCompletions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completions [wallet-address]",
		Short: "Query a user's IR completion history",
		Long: `Query all IR completions for a user, including scores, bonuses, and timestamps.

Example:
  $ aurad query confidencescore completions aura1abc... --arena Biometric
  $ aurad query confidencescore completions aura1abc... --page 1 --limit 50

This returns for each IR:
- IR ID (e.g., IR-102)
- Base and final scores
- Completion timestamp and block height
- Assistant verifier address
- Velocity, arena, and jackpot multipliers
- Arena type
- Completion status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			arenaFilter, _ := cmd.Flags().GetString("arena")

			queryClient := v1beta1.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &v1beta1.QueryUserCompletionsRequest{
				WalletAddress: args[0],
				ArenaFilter:   arenaFilter,
				Pagination:    pageReq,
			}

			res, err := queryClient.UserCompletions(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String("arena", "", "Filter by arena type (Biometric, Possession, Knowledge, etc.)")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "completions")

	return cmd
}

// CmdQueryScoreHistory queries a user's score change history
func CmdQueryScoreHistory() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history [wallet-address]",
		Short: "Query a user's confidence score change history",
		Long: `Query the complete history of confidence score changes for a user.

Example:
  $ aurad query confidencescore history aura1abc...
  $ aurad query confidencescore history aura1abc... --from-height 1000 --to-height 5000

This returns for each change:
- Block height and timestamp
- Score delta (positive or negative)
- New total score
- Change reason (IR completion, slash, governance adjustment, etc.)
- Related IR ID (if applicable)
- Transaction hash`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			fromHeight, _ := cmd.Flags().GetUint64("from-height")
			toHeight, _ := cmd.Flags().GetUint64("to-height")

			queryClient := v1beta1.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &v1beta1.QueryScoreHistoryRequest{
				WalletAddress: args[0],
				FromHeight:    fromHeight,
				ToHeight:      toHeight,
				Pagination:    pageReq,
			}

			res, err := queryClient.ScoreHistory(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Uint64("from-height", 0, "Starting block height for history query")
	cmd.Flags().Uint64("to-height", 0, "Ending block height for history query")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "history")

	return cmd
}

// CmdQueryThresholds queries verification thresholds
func CmdQueryThresholds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thresholds",
		Short: "Query confidence score verification thresholds",
		Long: `Query the confidence score thresholds for various verification levels and arena focus bonuses.

Example:
  $ aurad query confidencescore thresholds

This returns:
- Verified Human threshold (default: 10,000 CS)
- VC (Verifiable Credential) thresholds by type
- Arena focus thresholds (default: 5,000 CS per arena)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryThresholdsRequest{}

			res, err := queryClient.Thresholds(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryVerifiedUsers queries verified users
func CmdQueryVerifiedUsers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verified-users",
		Short: "Query list of verified users (CS >= 10,000)",
		Long: `Query all users who have achieved verified status (confidence score >= 10,000).

Example:
  $ aurad query confidencescore verified-users
  $ aurad query confidencescore verified-users --min-score 15000

This returns:
- List of wallet addresses
- Corresponding confidence scores
- Pagination support for large result sets`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			minScore, _ := cmd.Flags().GetUint64("min-score")

			queryClient := v1beta1.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &v1beta1.QueryVerifiedUsersRequest{
				MinScore:   minScore,
				Pagination: pageReq,
			}

			res, err := queryClient.VerifiedUsers(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Uint64("min-score", 0, "Minimum score filter (default: 10000)")
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "verified-users")

	return cmd
}

// CmdQueryArenaBreakdown queries a user's arena score breakdown
func CmdQueryArenaBreakdown() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "arena-breakdown [wallet-address]",
		Short: "Query a user's score breakdown by arena",
		Long: `Query detailed arena-by-arena score breakdown for a user.

Example:
  $ aurad query confidencescore arena-breakdown aura1abc...

This returns for each arena:
- Arena type (Biometric, Possession, Knowledge, Social, GeoLocation, HighAssurance, Persistence, Specialized)
- Total score in that arena
- Number of IRs completed in that arena
- Focus bonus status (active if arena score >= 5,000)

Also returns:
- List of arenas where focus bonus is active`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryArenaBreakdownRequest{
				WalletAddress: args[0],
			}

			res, err := queryClient.ArenaBreakdown(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQuerySlashRecords queries slash records for a user
func CmdQuerySlashRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slash-records [wallet-address]",
		Short: "Query slash records for a user",
		Long: `Query all score slashing records for a user.

Example:
  $ aurad query confidencescore slash-records aura1abc...

This returns for each slash:
- Slash amount (CS points deducted)
- Reason (fraud, false attestation, collusion, duplicate completion)
- Block height and timestamp
- Related IR ID
- Slash transaction hash
- Appeal deadline
- Appeal status (filed/not filed)
- Resolution status
- Authority who executed the slash
- Evidence (IPFS hash or metadata)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			req := &v1beta1.QuerySlashRecordRequest{
				WalletAddress: args[0],
				Pagination:    pageReq,
			}

			res, err := queryClient.SlashRecord(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "slash-records")

	return cmd
}

// CmdQueryParams queries module parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query confidence score module parameters",
		Long: `Query the current parameters for the confidence score module.

Example:
  $ aurad query confidencescore params

This returns:
- Verification thresholds (verified human, high assurance, arena focus)
- Velocity bonus configuration (days and multipliers)
- Arena multipliers (by threshold)
- Slashing parameters (percentage, appeal deposit)
- Rate limits (max IRs per day/hour)
- Jackpot configuration (odds and multipliers)
- Staleness settings (degradation rate)
- PoI rewards configuration (user/operator split, velocity bonus toggle)`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryParamsRequest{}

			res, err := queryClient.Params(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryIRCompletion queries a specific IR completion
func CmdQueryIRCompletion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ir-completion [wallet-address] [ir-id]",
		Short: "Query a specific IR completion record",
		Long: `Query detailed information about a specific IR completion for a user.

Example:
  $ aurad query confidencescore ir-completion aura1abc... IR-102

This returns:
- Whether the IR is completed
- Completion details:
  * Base and final scores
  * Completion timestamp and block height
  * Assistant verifier address
  * Proof and verifier hashes
  * Transaction hash
  * Velocity bonus applied (1.0-1.25x)
  * Arena bonus applied (1.0-1.5x)
  * Jackpot bonus applied (1.0x, 5.0x, or 25.0x)
  * Completion status (pending, verified, rejected, appealed)
  * Arena type`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryIRCompletionRequest{
				WalletAddress: args[0],
				IrId:          args[1],
			}

			res, err := queryClient.IRCompletion(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
