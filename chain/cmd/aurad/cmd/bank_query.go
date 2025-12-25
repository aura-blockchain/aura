// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"
)

// GetBankQueryCmd returns the query commands for the bank module
func GetBankQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        "bank",
		Short:                      "Querying commands for the bank module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		GetCmdQueryBalance(),
		GetCmdQueryTotalSupply(),
		GetCmdQueryDenomMetadata(),
	)

	return queryCmd
}

// GetCmdQueryBalance returns the balance query command
func GetCmdQueryBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balances [address]",
		Short: "Query an account's balance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			address := args[0]

			queryClient := banktypes.NewQueryClient(clientCtx)
			res, err := queryClient.AllBalances(
				context.Background(),
				&banktypes.QueryAllBalancesRequest{Address: address},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryTotalSupply returns the total supply query command
func GetCmdQueryTotalSupply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total [denom]",
		Short: "Query the total supply of coins or supply of a single coin denomination",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := banktypes.NewQueryClient(clientCtx)

			if len(args) == 0 {
				// Query total supply for all denominations
				res, err := queryClient.TotalSupply(
					context.Background(),
					&banktypes.QueryTotalSupplyRequest{},
				)
				if err != nil {
					return err
				}
				return clientCtx.PrintProto(res)
			}

			// Query total supply for a specific denom
			res, err := queryClient.SupplyOf(
				context.Background(),
				&banktypes.QuerySupplyOfRequest{Denom: args[0]},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryDenomMetadata returns the denom metadata query command
func GetCmdQueryDenomMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denom-metadata [denom]",
		Short: "Query the client metadata for coin denominations",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := banktypes.NewQueryClient(clientCtx)

			if len(args) == 0 {
				// Query metadata for all denominations
				res, err := queryClient.DenomsMetadata(
					context.Background(),
					&banktypes.QueryDenomsMetadataRequest{},
				)
				if err != nil {
					return err
				}
				return clientCtx.PrintProto(res)
			}

			// Query metadata for a specific denom
			res, err := queryClient.DenomMetadata(
				context.Background(),
				&banktypes.QueryDenomMetadataRequest{Denom: args[0]},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
