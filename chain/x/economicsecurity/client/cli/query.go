package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "economicsecurity",
		Short:                      "Querying commands for the economic security module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdQueryVestingSchedule(),
		CmdQueryVestingSchedulesByBeneficiary(),
		CmdQueryVoteLock(),
		CmdQueryVoteLocksByOwner(),
		CmdQueryVotingPower(),
		CmdQueryPendingTreasuryTx(),
		CmdQueryPendingTreasuryTxs(),
		CmdQueryInflationMetrics(),
		CmdQueryInflationAlerts(),
		CmdQueryLiquidityMiningStats(),
		CmdQueryMEVStats(),
		CmdQueryUserMEVBalance(),
		CmdQueryTokenomicsStats(),
	)

	return cmd
}

// CmdQueryParams queries module parameters
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query economic security module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &v1beta1.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryVestingSchedule queries a vesting schedule
func CmdQueryVestingSchedule() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vesting-schedule [schedule-id]",
		Short: "Query a vesting schedule by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryVestingScheduleRequest{
				ScheduleId: args[0],
			}

			res, err := queryClient.VestingSchedule(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryVestingSchedulesByBeneficiary queries all vesting schedules for a beneficiary
func CmdQueryVestingSchedulesByBeneficiary() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vesting-schedules [beneficiary-address]",
		Short: "Query all vesting schedules for a beneficiary",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryVestingSchedulesByBeneficiaryRequest{
				BeneficiaryAddress: args[0],
			}

			res, err := queryClient.VestingSchedulesByBeneficiary(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryVoteLock queries a vote lock
func CmdQueryVoteLock() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-lock [lock-id]",
		Short: "Query a vote lock by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryVoteLockRequest{
				LockId: args[0],
			}

			res, err := queryClient.VoteLock(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryVoteLocksByOwner queries all vote locks for an owner
func CmdQueryVoteLocksByOwner() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-locks [owner-address]",
		Short: "Query all vote locks for an owner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryVoteLocksByOwnerRequest{
				Owner: args[0],
			}

			res, err := queryClient.VoteLocksByOwner(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryVotingPower queries voting power for an address
func CmdQueryVotingPower() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voting-power [address]",
		Short: "Query voting power for an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryVotingPowerRequest{
				Address: args[0],
			}

			res, err := queryClient.VotingPower(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryPendingTreasuryTx queries a pending treasury transaction
func CmdQueryPendingTreasuryTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-treasury-tx [tx-id]",
		Short: "Query a pending treasury transaction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryPendingTreasuryTxRequest{
				TxId: args[0],
			}

			res, err := queryClient.PendingTreasuryTx(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryPendingTreasuryTxs queries all pending treasury transactions
func CmdQueryPendingTreasuryTxs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-treasury-txs",
		Short: "Query all pending treasury transactions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.PendingTreasuryTxs(context.Background(), &v1beta1.QueryPendingTreasuryTxsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryInflationMetrics queries inflation metrics
func CmdQueryInflationMetrics() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inflation-metrics",
		Short: "Query current inflation metrics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.InflationMetrics(context.Background(), &v1beta1.QueryInflationMetricsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryInflationAlerts queries inflation alerts
func CmdQueryInflationAlerts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inflation-alerts [limit]",
		Short: "Query inflation alerts",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			var limit uint64 = 10
			if len(args) > 0 {
				limit, err = strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid limit: %w", err)
				}
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryInflationAlertsRequest{
				Limit: limit,
			}

			res, err := queryClient.InflationAlerts(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryLiquidityMiningStats queries liquidity mining statistics
func CmdQueryLiquidityMiningStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidity-mining-stats",
		Short: "Query liquidity mining statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.LiquidityMiningStats(context.Background(), &v1beta1.QueryLiquidityMiningStatsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryMEVStats queries MEV redistribution statistics
func CmdQueryMEVStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mev-stats",
		Short: "Query MEV redistribution statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.MEVStats(context.Background(), &v1beta1.QueryMEVStatsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryUserMEVBalance queries user's MEV redistribution balance
func CmdQueryUserMEVBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-mev-balance [address]",
		Short: "Query user's MEV redistribution balance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			req := &v1beta1.QueryUserMEVBalanceRequest{
				Address: args[0],
			}

			res, err := queryClient.UserMEVBalance(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// CmdQueryTokenomicsStats queries overall tokenomics statistics
func CmdQueryTokenomicsStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokenomics-stats",
		Short: "Query overall tokenomics statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := v1beta1.NewQueryClient(clientCtx)

			res, err := queryClient.TokenomicsStats(context.Background(), &v1beta1.QueryTokenomicsStatsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
