package cmd

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/spf13/cobra"
)

// GetAuraSignCommand wraps the SDK's sign command with LEGACY_AMINO_JSON enforcement.
// This ensures that transactions are signed with amino-json by default, not SIGN_MODE_DIRECT.
func GetAuraSignCommand() *cobra.Command {
	cmd := authcli.GetSignCommand()

	// Store original PreRunE
	originalPreRunE := cmd.PreRunE

	// Wrap PreRunE to enforce sign mode
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Execute original PreRunE if it exists
		if originalPreRunE != nil {
			if err := originalPreRunE(cmd, args); err != nil {
				return err
			}
		}

		// Force LEGACY_AMINO_JSON if --sign-mode was not explicitly provided
		if !cmd.Flags().Changed(flags.FlagSignMode) {
			_ = cmd.Flags().Set(flags.FlagSignMode, flags.SignModeLegacyAminoJSON)
		}

		// Ensure client context reflects the sign mode
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		signMode, _ := cmd.Flags().GetString(flags.FlagSignMode)
		if signMode == "" {
			signMode = flags.SignModeLegacyAminoJSON
		}

		clientCtx = clientCtx.WithSignModeStr(signMode)
		return client.SetCmdClientContext(cmd, clientCtx)
	}

	// Update flag default value
	if signModeFlag := cmd.Flags().Lookup(flags.FlagSignMode); signModeFlag != nil {
		signModeFlag.DefValue = flags.SignModeLegacyAminoJSON
	}

	return cmd
}
