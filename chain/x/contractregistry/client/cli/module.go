// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the query commands for the contractregistry module
// NOTE: Future enhancement - Implement full query commands (query.go.skip has the full implementation)
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "contractregistry",
		Short:                      "Querying commands for the contractregistry module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}
	return cmd
}

// GetTxCmd returns the transaction commands for the contractregistry module
// NOTE: Future enhancement - Implement full tx commands (tx.go.skip has the full implementation)
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "contractregistry",
		Short:                      "Transaction commands for the contractregistry module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}
	return cmd
}
