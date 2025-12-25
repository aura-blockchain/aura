// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/aequitas/aura/chain/x/aura-bindings/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "AuraBindings transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Note: The aura-bindings module primarily serves as a bridge for CosmWasm
	// custom bindings and doesn't define traditional user-facing transaction commands.
	// Transactions are handled through CosmWasm contract execution using the
	// message_plugin.go which processes custom CosmWasm messages.
	//
	// If future direct transaction commands are needed, they should be added here.

	return cmd
}
