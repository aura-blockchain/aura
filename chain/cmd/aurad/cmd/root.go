package cmd

import (
	"fmt"
	"os"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aequitas/aura/chain/app"
)

const (
	// EnvPrefix is the prefix for environment variables
	EnvPrefix = "AURA"

	// DefaultHomePath is the default home directory for the daemon
	DefaultHomePath = ".aura"

	// Flag constants
	FlagAPIAddress  = "api.address"
	FlagGRPCAddress = "grpc.address"
)

var (
	// homeDir is the home directory for the daemon
	homeDir string

	// configFile is the path to the config file
	configFile string
)

// NewRootCmd creates a new root command for the Aura daemon
func NewRootCmd() *cobra.Command {
	// Initialize encoding config for client context
	encodingConfig := app.MakeEncodingConfig()

	// Create initial client context with codec, TxConfig, and AccountRetriever
	// AccountRetriever is required for transaction signing to query account info
	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithInput(os.Stdin)

	rootCmd := &cobra.Command{
		Use:   "aurad",
		Short: "Aura Blockchain Daemon",
		Long: `Aura is a next-generation blockchain focused on decentralized identity,
verifiable credentials, and secure data management.

The Aura daemon (aurad) is the main entry point for running an Aura blockchain node.
It provides commands for initialization, starting the node, querying state, and more.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Build client context from flags and client.toml so tx/query commands
			// inherit proper home/node/chain-id defaults.
			clientCtx := initClientCtx.WithHomeDir(homeDir).WithViper(EnvPrefix)

			var err error
			clientCtx, err = client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			homeDir = clientCtx.HomeDir
			if homeDir == "" {
				homeDir = getDefaultHomeDir()
				clientCtx = clientCtx.WithHomeDir(homeDir)
			}

			// Initialize configuration (server + viper-based settings)
			if err := initConfig(); err != nil {
				return err
			}

			clientCtx, err = clientconfig.ReadFromClientConfig(clientCtx)
			if err != nil {
				return err
			}

			// Ensure CLI defaults to LEGACY_AMINO_JSON signing unless explicitly overridden.
			if clientCtx.SignModeStr == "" {
				clientCtx = clientCtx.WithSignModeStr(flags.SignModeLegacyAminoJSON)
			}

			// Set the client context with encoding config on the command
			if err := client.SetCmdClientContext(cmd, clientCtx); err != nil {
				return err
			}

			return nil
		},
	}

	// Add persistent flags
	rootCmd.PersistentFlags().StringVar(&homeDir, flags.FlagHome, getDefaultHomeDir(), "directory for config and data")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.aura/config.toml)")
	rootCmd.PersistentFlags().String("log-level", "info", "logging level (trace|debug|info|warn|error|fatal|panic)")
	rootCmd.PersistentFlags().String("log-format", "plain", "logging format (json|plain)")
	rootCmd.PersistentFlags().Bool("verbose", false, "enable verbose output")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug output")
	rootCmd.PersistentFlags().String(flags.FlagOutput, "text", "output format (text|json|yaml|csv)")
	rootCmd.PersistentFlags().String(flags.FlagNode, "tcp://localhost:26657", "<host>:<port> to CometBFT RPC interface for this chain")
	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	// Bind flags to viper
	viper.BindPFlag(flags.FlagHome, rootCmd.PersistentFlags().Lookup(flags.FlagHome))
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log-format", rootCmd.PersistentFlags().Lookup("log-format"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag(flags.FlagOutput, rootCmd.PersistentFlags().Lookup(flags.FlagOutput))
	viper.BindPFlag(flags.FlagNode, rootCmd.PersistentFlags().Lookup(flags.FlagNode))
	viper.BindPFlag(flags.FlagChainID, rootCmd.PersistentFlags().Lookup(flags.FlagChainID))
	// Force legacy amino JSON as the default signing mode so CLI transactions
	// never fall back to SIGN_MODE_DIRECT when the user does not pass --sign-mode.
	viper.SetDefault(flags.FlagSignMode, flags.SignModeLegacyAminoJSON)

	// Add subcommands
	addSubcommands(rootCmd, initClientCtx.TxConfig)

	return rootCmd
}

// addSubcommands adds all subcommands to the root command
func addSubcommands(rootCmd *cobra.Command, txEncCfg client.TxConfig) {
	// Create logger
	logger := log.NewLogger(os.Stdout)

	// Note: App initialization will be done lazily when needed
	// to avoid errors during command help/version checks
	var auraApp *app.App

	accountCodec := app.AccountAddressCodec()
	validatorCodec := app.ValidatorAddressCodec()
	genesisCmd := GenesisCmd(app.ModuleBasics, accountCodec, validatorCodec, txEncCfg)

	rootCmd.AddCommand(
		InitCmd(),
		StartCmd(&auraApp, logger),
		genesisCmd,
		VersionCmd(),
		StatusCmd(),
		KeysCmd(),
		QueryCmd(),
		TxCmd(),
		InteractiveCmd(),
		ConfigCmd(),
		CompletionCmd(),
		BatchCmd(),
		HelpCmd(),
		MnemonicSignerCmd(), // Bypass SDK 0.53.x keyring bug
	)
}

// initConfig initializes the daemon configuration
func initConfig() error {
	// Set environment variable prefix
	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()

	// Set config file if specified
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		// Use default config path
		configPath := fmt.Sprintf("%s/config", homeDir)
		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
	}

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; ignore error if desired
	}

	return nil
}

// getDefaultHomeDir returns the default home directory
func getDefaultHomeDir() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultHomePath
	}
	return fmt.Sprintf("%s/%s", userHomeDir, DefaultHomePath)
}

// GetHomeDirVar returns the package-level homeDir variable set by the --home flag
func GetHomeDirVar() string {
	return homeDir
}
