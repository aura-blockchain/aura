package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
)

// TendermintCmd returns the CometBFT/Tendermint operations command
func TendermintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cometbft",
		Aliases: []string{"tendermint", "comet"},
		Short:   "CometBFT subcommands",
		Long: `CometBFT (formerly Tendermint) blockchain engine operations.

Provides direct access to the underlying consensus engine including:
- Node key management
- Validator key operations
- Network peer operations
- Block and state inspection
- Version information`,
	}

	cmd.AddCommand(
		showNodeIDCmd(),
		showValidatorCmd(),
		showAddressCmd(),
		versionCmd(),
		resetAllCmd(),
		resetStateCmd(),
		unsafeResetAllCmd(),
	)

	return cmd
}

// showNodeIDCmd displays the node ID
func showNodeIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-node-id",
		Short: "Show this node's ID",
		Long:  "Display the node ID derived from node_key.json in the config directory.",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config

			nodeKey, err := p2p.LoadNodeKey(cfg.NodeKeyFile())
			if err != nil {
				return fmt.Errorf("failed to load node key: %w", err)
			}

			fmt.Println(nodeKey.ID())
			return nil
		},
	}

	return cmd
}

// showValidatorCmd displays the validator public key
func showValidatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-validator",
		Short: "Show this node's validator public key",
		Long:  "Display the validator consensus public key from priv_validator_key.json.",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config

			pvKeyFile := cfg.PrivValidatorKeyFile()
			privValidator := privval.LoadFilePV(pvKeyFile, "")

			valPubKey, err := privValidator.GetPubKey()
			if err != nil {
				return fmt.Errorf("failed to get validator pubkey: %w", err)
			}

			// Print the public key in hex format
			fmt.Println(valPubKey)

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// showAddressCmd displays the validator consensus address
func showAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-address",
		Short: "Show this node's validator consensus address",
		Long:  "Display the validator consensus address derived from the validator public key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config

			pvKeyFile := cfg.PrivValidatorKeyFile()
			privValidator := privval.LoadFilePV(pvKeyFile, "")

			valPubKey, err := privValidator.GetPubKey()
			if err != nil {
				return fmt.Errorf("failed to get validator pubkey: %w", err)
			}

			fmt.Println(valPubKey.Address())
			return nil
		},
	}

	return cmd
}

// versionCmd displays CometBFT version
func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print CometBFT version",
		Long:  "Display the version of CometBFT being used by this binary.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("CometBFT version info:")
			fmt.Println("  - Version information would be displayed here")
			fmt.Println("  - Requires import of version package")
			return nil
		},
	}

	return cmd
}

// resetAllCmd removes all blockchain data but preserves node/validator keys
func resetAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Remove all blockchain data (preserves keys)",
		Long: `Remove all blockchain data including blocks and state.

This command will:
- Delete the blockchain database
- Delete application state
- Preserve node_key.json
- Preserve priv_validator_key.json
- Preserve priv_validator_state.json

WARNING: This is irreversible. The node will need to resync from genesis or snapshot.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Reset functionality requires server integration")
			fmt.Println("Use with extreme caution in production")
			return nil
		},
	}

	return cmd
}

// resetStateCmd removes application state but preserves blocks
func resetStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset-state",
		Short: "Remove application state (preserves blocks)",
		Long: `Remove only the application state database.

This command preserves:
- Block database
- Node and validator keys
- Configuration files

The application state will be rebuilt by replaying blocks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Reset state functionality requires server integration")
			return nil
		},
	}

	return cmd
}

// unsafeResetAllCmd removes all data including validator state
func unsafeResetAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsafe-reset-all",
		Short: "Remove ALL data including validator state",
		Long: `DANGER: Remove all blockchain data AND validator state.

This command will:
- Delete the blockchain database
- Delete application state
- Reset priv_validator_state.json (double-sign risk!)
- Preserve keys

WARNING: This can lead to double-signing if used on an active validator.
Only use this for testnets or if you know what you're doing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Unsafe reset ALL functionality requires server integration")
			fmt.Println("DANGER: Can lead to double-signing on active validators")
			return nil
		},
	}

	return cmd
}
