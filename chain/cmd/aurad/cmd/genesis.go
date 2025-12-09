package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	"github.com/aequitas/aura/chain/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankexported "github.com/cosmos/cosmos-sdk/x/bank/exported"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/aequitas/aura/chain/cmd/aurad/cmd/security"
	contractregistrytypes "github.com/aequitas/aura/chain/x/contractregistry/types"
	incidentresponsetypes "github.com/aequitas/aura/chain/x/incidentresponse/types"
	contractregistrypb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
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
func GenesisCmd(
	mbm module.BasicManager,
	accountCodec address.Codec,
	validatorCodec address.Codec,
	txEncCfg client.TxEncodingConfig,
) *cobra.Command {
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
	if mbm != nil && accountCodec != nil && validatorCodec != nil {
		validateCmd := genutilcli.ValidateGenesisCmd(mbm)
		validateCmd.RunE = func(cmd *cobra.Command, args []string) error {
			encCfg := app.MakeEncodingConfig()
			clientCtx := client.GetClientContextFromCmd(cmd).
				WithTxConfig(encCfg.TxConfig).
				WithCodec(encCfg.Codec).
				WithInterfaceRegistry(encCfg.InterfaceRegistry)
			if err := client.SetCmdClientContext(cmd, clientCtx); err != nil {
				return err
			}

			if err := ensureGenesisSections(cmd, args, mbm); err != nil {
				return err
			}

			serverCtx := server.GetServerContextFromCmd(cmd)
			var genesisPath string
			if len(args) == 0 {
				if serverCtx == nil || serverCtx.Config == nil {
					return fmt.Errorf("server context not initialized; unable to locate genesis file")
				}
				genesisPath = serverCtx.Config.GenesisFile()
			} else {
				genesisPath = args[0]
			}

			appGenesis, err := genutiltypes.AppGenesisFromFile(genesisPath)
			if err != nil {
				return enrichUnmarshalError(err)
			}

			if err := appGenesis.ValidateAndComplete(); err != nil {
				const chainUpgradeGuide = "https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md"
				return fmt.Errorf("make sure that you have correctly migrated all CometBFT consensus params. Refer the UPGRADING.md (%s): %w", chainUpgradeGuide, err)
			}

			var genState map[string]json.RawMessage
			if err := json.Unmarshal(appGenesis.AppState, &genState); err != nil {
				if strings.Contains(err.Error(), "unexpected end of JSON input") {
					return fmt.Errorf("app_state is missing in the genesis file: %s", err.Error())
				}
				return fmt.Errorf("error unmarshalling genesis doc %s: %w", genesisPath, err)
			}

			txCfg := encCfg.TxConfig

			// Validate genutil/gentx explicitly with the safe decoder.
			genutilGs := genutiltypes.GetGenesisStateFromAppState(clientCtx.Codec, genState)
			if genutilGs != nil {
				if err := genutiltypes.ValidateGenesis(genutilGs, txCfg.TxJSONDecoder(), genutiltypes.DefaultMessageValidator); err != nil {
					errStr := fmt.Sprintf("error validating genesis file %s (module %s): %s", genesisPath, genutiltypes.ModuleName, err.Error())
					if errors.Is(err, io.EOF) {
						errStr = fmt.Sprintf("%s: section is missing in the app_state", errStr)
					}
					return fmt.Errorf("%s", errStr)
				}
			}

			for name, mod := range mbm {
				if name == genutiltypes.ModuleName {
					continue
				}
				vg, ok := mod.(interface {
					ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error
				})
				if !ok {
					continue
				}

				if err := vg.ValidateGenesis(clientCtx.Codec, txCfg, genState[name]); err != nil {
					errStr := fmt.Sprintf("error validating genesis file %s (module %s): %s", genesisPath, name, err.Error())
					if errors.Is(err, io.EOF) {
						errStr = fmt.Sprintf("%s: section is missing in the app_state", errStr)
					}
					return fmt.Errorf("%s", errStr)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "File at %s is a valid genesis file\n", genesisPath)
			return nil
		}
		validateCmd = wrapWithSecurity(validateCmd)
		validateCmd.Use = "validate-genesis"
		validateCmd.Aliases = append(validateCmd.Aliases, "validate")

		cmd.AddCommand(
			wrapWithSecurity(genutilcli.AddGenesisAccountCmd(defaultNodeHome, accountCodec)),
			wrapWithSecurity(genutilcli.GenTxCmd(
				mbm,
				txEncCfg,
				bankBalancesIterator{},
				defaultNodeHome,
				validatorCodec,
			)),
			validateCmd,
			wrapWithSecurity(genutilcli.CollectGenTxsCmd(
				bankBalancesIterator{},
				defaultNodeHome,
				genutiltypes.DefaultMessageValidator,
				validatorCodec,
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
		if clientCtx.Codec == nil || clientCtx.InterfaceRegistry == nil {
			encCfg := app.MakeEncodingConfig()
			clientCtx = clientCtx.
				WithTxConfig(encCfg.TxConfig).
				WithCodec(encCfg.Codec).
				WithInterfaceRegistry(encCfg.InterfaceRegistry)
		}
		homeDir := clientCtx.HomeDir
		if homeFlag := c.Flag(flags.FlagHome); homeFlag != nil {
			if flagHome := homeFlag.Value.String(); flagHome != "" {
				homeDir = flagHome
			}
		}

		validator := security.NewPathValidator(secLogger)
		validHome, err := validator.ValidateAndCleanHomePath(homeDir)
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
		if serverCtx != nil && serverCtx.Config != nil {
			serverCtx.Config.SetRoot(validHome)
		}

		// Call the original command
		return originalRunE(c, args)
	}

	return cmd
}

// ensureGenesisSections backfills missing module genesis sections with defaults
// so validation succeeds even if app_state omitted registered modules.
func ensureGenesisSections(cmd *cobra.Command, args []string, mbm module.BasicManager) error {
	clientCtx := client.GetClientContextFromCmd(cmd)
	if clientCtx.Codec == nil || clientCtx.InterfaceRegistry == nil {
		encCfg := app.MakeEncodingConfig()
		clientCtx = clientCtx.
			WithCodec(encCfg.Codec).
			WithInterfaceRegistry(encCfg.InterfaceRegistry).
			WithTxConfig(encCfg.TxConfig)
		if err := client.SetCmdClientContext(cmd, clientCtx); err != nil {
			return err
		}
	}

	serverCtx := server.GetServerContextFromCmd(cmd)
	var genesisPath string
	if len(args) == 0 {
		if serverCtx == nil || serverCtx.Config == nil {
			return fmt.Errorf("server context not initialized; unable to locate genesis file")
		}
		genesisPath = serverCtx.Config.GenesisFile()
	} else {
		genesisPath = args[0]
	}

	appGenesis, err := genutiltypes.AppGenesisFromFile(genesisPath)
	if err != nil {
		return fmt.Errorf("failed to load genesis file %s: %w", genesisPath, err)
	}

	var genState map[string]json.RawMessage
	if err := json.Unmarshal(appGenesis.AppState, &genState); err != nil {
		return fmt.Errorf("error unmarshalling genesis doc %s: %w", genesisPath, err)
	}

	updated := false
	for name, mod := range mbm {
		if len(genState[name]) == 0 {
			genState[name] = defaultGenesisForModule(clientCtx.Codec, mod)
			updated = true
		}
	}

	// Fix up known-invalid defaults for incidentresponse (emergency pause with no keys).
	if raw := genState[incidentresponsetypes.ModuleName]; len(raw) != 0 {
		var gs incidentresponsetypes.GenesisState
		if err := json.Unmarshal(raw, &gs); err == nil {
			if gs.Params != nil && gs.Params.EmergencyPauseEnabled && len(gs.Params.PauseAuthorizedKeys) == 0 {
				gs.Params.EmergencyPauseEnabled = false
				patched, _ := json.Marshal(gs)
				genState[incidentresponsetypes.ModuleName] = patched
				updated = true
				raw = genState[incidentresponsetypes.ModuleName]
			}
		}

		if err := json.Unmarshal(raw, &gs); err == nil {
			if gs.Params != nil && gs.Params.DisasterRecovery.Enabled && len(gs.Params.DisasterRecovery.BackupLocations) == 0 {
				gs.Params.DisasterRecovery.Enabled = false
				patched, _ := json.Marshal(gs)
				genState[incidentresponsetypes.ModuleName] = patched
				updated = true
			}
		}
	}

	// Ensure contractregistry has params to satisfy InitGenesis.
	if raw := genState[contractregistrytypes.ModuleName]; len(raw) != 0 {
		var gs contractregistrypb.GenesisState
		if err := json.Unmarshal(raw, &gs); err == nil {
			// Params is a value type, check if it's zero-valued by checking a field
			if gs.Params.MaxContractsPerCreator == 0 {
				gs.Params = *contractregistrytypes.DefaultParams()
				genState[contractregistrytypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&gs)
				updated = true
			}
		}
	}

	if !updated {
		return nil
	}

	newAppState, err := json.Marshal(genState)
	if err != nil {
		return fmt.Errorf("failed to marshal patched app_state: %w", err)
	}

	appGenesis.AppState = newAppState
	if err := genutil.ExportGenesisFile(appGenesis, genesisPath); err != nil {
		return fmt.Errorf("failed to write patched genesis file %s: %w", genesisPath, err)
	}

	return nil
}

// defaultGenesisForModule returns a module's default genesis if available.
func defaultGenesisForModule(cdc codec.JSONCodec, mod module.AppModuleBasic) json.RawMessage {
	if mod == nil {
		return json.RawMessage(`{}`)
	}
	if dg, ok := mod.(interface {
		DefaultGenesis(codec.JSONCodec) json.RawMessage
	}); ok {
		if raw := dg.DefaultGenesis(cdc); len(raw) > 0 {
			return raw
		}
	}
	return json.RawMessage(`{}`)
}

// enrichUnmarshalError mirrors the SDK helper to improve syntax error context.
func enrichUnmarshalError(err error) error {
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return fmt.Errorf("error at offset %d: %s", syntaxErr.Offset, syntaxErr.Error())
	}
	return err
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
