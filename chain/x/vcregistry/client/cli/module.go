package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
)

// GetModuleRootCmd returns the root command for the VC Registry module
func GetModuleRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "vcregistry",
		Short:                      "VC Registry module commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Add query and transaction subcommands
	cmd.AddCommand(
		GetQueryCmd(),
		GetTxCmd(),
	)

	return cmd
}
