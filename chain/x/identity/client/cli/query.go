package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/aequitas/aura/chain/x/identity/types"
)

// GetQueryCmd returns the query commands for the identity module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the identity module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Add query subcommands here
	// These will use AutoCLI under the hood, registered via module.go

	return cmd
}
