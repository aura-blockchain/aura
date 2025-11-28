package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "monitoring",
		Short:                      "Monitoring transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdAcknowledgeAlert(),
		CmdResolveAlert(),
	)

	return cmd
}

// CmdAcknowledgeAlert acknowledges an alert
func CmdAcknowledgeAlert() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "acknowledge-alert [alert-id]",
		Short: "Acknowledge a monitoring alert",
		Long: `Acknowledge a monitoring alert to indicate it has been reviewed.

Example:
  aurad tx monitoring acknowledge-alert "alert-123" --from alice
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Note: This would require a proper Msg type to be defined
			// For now, this is a placeholder structure
			// In production, you would create a proper MsgAcknowledgeAlert proto message

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), nil)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// CmdResolveAlert resolves an alert
func CmdResolveAlert() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve-alert [alert-id] [resolution-notes]",
		Short: "Resolve a monitoring alert",
		Long: `Mark a monitoring alert as resolved with optional notes.

Example:
  aurad tx monitoring resolve-alert "alert-123" "Issue fixed by restarting validator" --from alice
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Note: This would require a proper Msg type to be defined
			// For now, this is a placeholder structure
			// In production, you would create a proper MsgResolveAlert proto message

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), nil)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
