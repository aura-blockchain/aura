// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/aequitas/aura/chain/x/aura-bindings/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the aura-bindings module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryStats(),
		GetCmdMessageStats(),
		GetCmdAllStats(),
	)

	return cmd
}

// GetCmdQueryStats queries the query statistics
func GetCmdQueryStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-stats",
		Short: "Query the query usage statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryStatsRequest{}

			res, err := queryClient.QueryStats(context.Background(), req)
			if err != nil {
				return err
			}

			// Use JSON marshaling since response doesn't implement proto.Message
			bz, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}
			return clientCtx.PrintString(string(bz))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdMessageStats queries the message statistics
func GetCmdMessageStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "message-stats",
		Short: "Query the message usage statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.MessageStatsRequest{}

			res, err := queryClient.MessageStats(context.Background(), req)
			if err != nil {
				return err
			}

			// Use JSON marshaling since response doesn't implement proto.Message
			bz, err := json.MarshalIndent(res, "", "  ")
			if err != nil {
				return err
			}
			return clientCtx.PrintString(string(bz))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdAllStats queries both query and message statistics
func GetCmdAllStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all-stats",
		Short: "Query all usage statistics (queries and messages)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			req := &types.AllStatsRequest{}

			res, err := queryClient.AllStats(context.Background(), req)
			if err != nil {
				return err
			}

			fmt.Printf("Query Statistics:\n")
			for queryType, count := range res.QueryStats {
				fmt.Printf("  %s: %d\n", queryType, count)
			}

			fmt.Printf("\nMessage Statistics:\n")
			for msgType, count := range res.MessageStats {
				fmt.Printf("  %s: %d\n", msgType, count)
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
