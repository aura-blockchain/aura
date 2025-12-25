// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
)

// GetModuleRootCmd returns the root command for the Data Registry module
func GetModuleRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "dataregistry",
		Short:                      "Data Registry module commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Add query and transaction subcommands
// 	cmd.AddCommand(
// 		GetQueryCmd(),
// 		GetTxCmd(),
// 	)

	return cmd
}
