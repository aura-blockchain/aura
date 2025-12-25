// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
)

// GetTxCmd returns the transaction commands for the prevalidation module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "prevalidation",
		Short:                      "Pre-validation transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Note: The prevalidation module is primarily an internal optimization module
	// that runs automatically during off-peak hours. There are no user-facing
	// transaction commands as pre-validation is handled by the system automatically.
	//
	// Pre-validated transactions are created by validators during scheduled runs
	// and consumed automatically when matching transactions are submitted.
	//
	// Users benefit from pre-validation transparently through:
	// - Faster transaction execution
	// - Lower gas costs
	// - Reduced energy consumption
	//
	// For module configuration, governance proposals should be used to update params.

	return cmd
}
