package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankexported "github.com/cosmos/cosmos-sdk/x/bank/exported"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/aequitas/aura/chain/cmd/aurad/cmd/security"
)

const (
	// FlagMoniker is the flag for the node moniker
	FlagMoniker = "moniker"
	// FlagChainID is the flag for the chain ID
	FlagChainID = "chain-id"
	// DefaultChainID is the default chain ID
	DefaultChainID = "aura-1"
	// DefaultMoniker is the default node moniker
	DefaultMoniker = "aura-node"
)

// bankBalancesIterator implements GenesisBalancesIterator for bank module
type bankBalancesIterator struct{}

// IterateGenesisBalances iterates over the genesis balances
func (b bankBalancesIterator) IterateGenesisBalances(
	cdc codec.JSONCodec,
	appGenesis map[string]json.RawMessage,
	cb func(bankexported.GenesisBalance) (stop bool),
) {
	for _, balance := range banktypes.GetGenesisStateFromAppState(cdc, appGenesis).Balances {
		if cb(balance) {
			break
		}
	}
}

// GenesisCmd returns the genesis command group with security-wrapped SDK commands
func GenesisCmd(mbm module.BasicManager, addressCodec address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genesis",
		Short: "Genesis file operations",
		Long: `Commands for working with the genesis file.

The genesis file defines the initial state of the blockchain including:
- Initial accounts and balances
- Genesis validators
- Module parameters and initial state`,
	}

	defaultNodeHome := getDefaultHomeDirForGenesis()

	// Add standard SDK genesis commands with security validation
	// Only add commands that require mbm and addressCodec if they are provided
	if mbm != nil && addressCodec != nil {
		cmd.AddCommand(
			wrapWithSecurity(genutilcli.AddGenesisAccountCmd(defaultNodeHome, addressCodec)),
			wrapWithSecurity(genutilcli.ValidateGenesisCmd(mbm)),
			wrapWithSecurity(genutilcli.CollectGenTxsCmd(
				bankBalancesIterator{},
				defaultNodeHome,
				genutiltypes.DefaultMessageValidator,
				addressCodec,
			)),
		)
	}

	// Always add quickstart command
	cmd.AddCommand(QuickstartCmd())

	return cmd
}

// getDefaultHomeDirForGenesis returns the default home directory
func getDefaultHomeDirForGenesis() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return ".aura"
	}
	return filepath.Join(userHomeDir, ".aura")
}

// getSecurityLoggerForGenesis returns a security logger instance
func getSecurityLoggerForGenesis() security.Logger {
	homeDir := getDefaultHomeDirForGenesis()
	logger, err := security.NewSecurityLogger(homeDir, true)
	if err != nil {
		// Fall back to console logger if file logging fails
		return security.NewConsoleLogger()
	}
	return logger
}

// wrapWithSecurity wraps a cobra command with security validation
func wrapWithSecurity(cmd *cobra.Command) *cobra.Command {
	// Store the original RunE
	originalRunE := cmd.RunE

	// Wrap with security validation
	cmd.RunE = func(c *cobra.Command, args []string) error {
		// Get security logger
		secLogger := getSecurityLoggerForGenesis()

		// Validate home directory path
		clientCtx := client.GetClientContextFromCmd(c)
		validator := security.NewPathValidator(secLogger)
		validHome, err := validator.ValidateAndCleanHomePath(clientCtx.HomeDir)
		if err != nil {
			return fmt.Errorf("invalid home directory: %w", err)
		}

		// Update the client context with validated home
		clientCtx = clientCtx.WithHomeDir(validHome)
		if err := client.SetCmdClientContext(c, clientCtx); err != nil {
			return err
		}

		// Update server context with validated home
		serverCtx := server.GetServerContextFromCmd(c)
		serverCtx.Config.SetRoot(validHome)

		// Call the original command
		return originalRunE(c, args)
	}

	return cmd
}

// QuickstartCmd returns a command that helps users quickly set up a single-node chain
func QuickstartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quickstart",
		Short: "Quickly set up a single-node development chain",
		Long: `Quickly set up a single-node development chain with a validator account.

This command provides step-by-step instructions for:
1. Creating a validator key
2. Adding the validator as a genesis account with tokens
3. Creating a genesis transaction (gentx)
4. Collecting the gentx into genesis.json
5. Starting the chain

This is ideal for development and testing purposes.

Example:
$ aurad genesis quickstart`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			validatorName := "validator"
			chainID := DefaultChainID
			moniker := DefaultMoniker
			homeDir := clientCtx.HomeDir

			fmt.Printf("🚀 Aura Blockchain Quick Start Guide\n\n")

			fmt.Printf("Configuration:\n")
			fmt.Printf("  Chain ID: %s\n", chainID)
			fmt.Printf("  Moniker: %s\n", moniker)
			fmt.Printf("  Validator: %s\n", validatorName)
			fmt.Printf("  Home: %s\n", homeDir)
			fmt.Printf("\n")

			fmt.Printf("Steps to set up your development chain:\n\n")

			fmt.Printf("1. Initialize the node:\n")
			fmt.Printf("   aurad init %s --chain-id %s\n\n", moniker, chainID)

			fmt.Printf("2. Create a validator key:\n")
			fmt.Printf("   aurad keys add %s --keyring-backend test\n\n", validatorName)

			fmt.Printf("3. Add genesis account with tokens:\n")
			fmt.Printf("   aurad genesis add-genesis-account %s 100000000stake --keyring-backend test\n\n", validatorName)

			fmt.Printf("4. Generate genesis transaction:\n")
			fmt.Printf("   aurad genesis gentx %s 90000000stake --chain-id %s --keyring-backend test\n\n", validatorName, chainID)

			fmt.Printf("5. Collect genesis transactions:\n")
			fmt.Printf("   aurad genesis collect-gentxs\n\n")

			fmt.Printf("6. Validate genesis file:\n")
			fmt.Printf("   aurad genesis validate-genesis\n\n")

			fmt.Printf("7. Start the chain:\n")
			fmt.Printf("   aurad start\n\n")

			fmt.Printf("For more information, run:\n")
			fmt.Printf("   aurad genesis --help\n")
			fmt.Printf("   aurad [command] --help\n")

			return nil
		},
	}

	return cmd
}
