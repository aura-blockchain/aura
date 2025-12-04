package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"syscall"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	pruningtypes "cosmossdk.io/store/pruning/types"
	storetypes "cosmossdk.io/store/types"
	storewrapper "cosmossdk.io/store/wrapper"
	"github.com/cosmos/iavl"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aequitas/aura/chain/app"
	"github.com/aequitas/aura/chain/cmd/aurad/cmd/security"
	contractregistrytypes "github.com/aequitas/aura/chain/x/contractregistry/types"
	contractregistrypb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cmtcfg "github.com/cometbft/cometbft/config"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	// Flags for node configuration
	flagWithComet          = "with-comet"
	flagAddress            = "address"
	flagTransport          = "transport"
	flagTraceStore         = "trace-store"
	flagCPUProfile         = "cpu-profile"
	flagPruning            = "pruning"
	flagPruningKeepRecent  = "pruning-keep-recent"
	flagPruningInterval    = "pruning-interval"
	flagMinGasPrices       = "minimum-gas-prices"
	flagHaltHeight         = "halt-height"
	flagHaltTime           = "halt-time"
	flagInterBlockCache    = "inter-block-cache"
	flagUnsafeSkipUpgrades = "unsafe-skip-upgrades"
	flagTrace              = "trace"
	flagInvCheckPeriod     = "inv-check-period"
	flagAPIEnable          = "api.enable"
	flagAPIAddress         = "api.address"
	flagGRPCEnable         = "grpc.enable"
	flagGRPCAddress        = "grpc.address"
	flagRPCListenAddress   = "rpc.address"
	flagP2PListenAddress   = "p2p.address"
	flagGRPCWebEnable      = "grpc-web.enable"
	flagJSONRPCEnable      = "json-rpc.enable"
	flagJSONRPCAddress     = "json-rpc.address"

	// Default values
	DefaultGRPCAddress = "localhost:9090"
	DefaultAPIAddress  = "tcp://localhost:1317"
	DefaultRPCAddress  = "tcp://localhost:26657"
	DefaultP2PAddress  = "tcp://0.0.0.0:26656"
	DefaultABCIAddress = "tcp://0.0.0.0:26658"
)

// StartCmd returns the start command to run the node
func StartCmd(auraApp **app.App, logger log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Run the full node",
		Long: `Run the full node application with CometBFT in process.

The node will start the blockchain consensus engine (CometBFT), ABCI application,
gRPC server, and REST API server. The node will continue running until interrupted.

Pruning options can be provided via the '--pruning' flag or alternatively with
'--pruning-keep-recent' and '--pruning-interval' together.

For '--pruning' the options are as follows:

  default: the last 362880 states are kept, pruning at 10 block intervals
  nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
  everything: 2 latest states will be kept; pruning at 10 block intervals
  custom: allow pruning options to be manually specified through 'pruning-keep-recent' and 'pruning-interval'

Node halting configurations exist in the form of two flags: '--halt-height' and '--halt-time'.
During the ABCI Commit phase, the node will check if the current block height is greater than
or equal to the halt-height or if the current block time is greater than or equal to the halt-time.
If so, the node will attempt to gracefully shutdown and the block will not be committed.
`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			// Bind all flags to viper
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				return fmt.Errorf("failed to bind flags: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			withCMT := viper.GetBool(flagWithComet)
			if !withCMT {
				// Stand-alone mode: create app here if not provided
				if *auraApp == nil {
					*auraApp = app.NewAppWithLogger(logger)
				}
				logger.Info("starting ABCI without CometBFT")
				return startStandAlone(cmd, *auraApp, logger)
			}

			// CometBFT mode: app will be created in startInProcess with disk DB
			// Do NOT create app here - it would use in-memory DB
			logger.Info("starting node with ABCI CometBFT in-process")
			return startInProcess(cmd, *auraApp, logger)
		},
	}

	// Add flags
	addStartFlags(cmd)

	return cmd
}

// addStartFlags adds all flags for the start command
func addStartFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(flagWithComet, true, "Run with CometBFT in process")
	cmd.Flags().String(flagAddress, DefaultABCIAddress, "Listen address for ABCI")
	cmd.Flags().String(flagTransport, "socket", "Transport protocol: socket, grpc")
	cmd.Flags().String(flagTraceStore, "", "Enable KVStore tracing to an output file")
	cmd.Flags().String(flagCPUProfile, "", "Enable CPU profiling and write to file")

	// Pruning options
	cmd.Flags().String(flagPruning, pruningtypes.PruningOptionDefault, "Pruning strategy (default|nothing|everything|custom)")
	cmd.Flags().Uint64(flagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(flagPruningInterval, 0, "Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom')")

	// State sync
	cmd.Flags().Uint(flagInvCheckPeriod, 0, "Assert registered invariants every N blocks")

	// Halting
	cmd.Flags().Uint64(flagHaltHeight, 0, "Block height at which to gracefully halt the chain")
	cmd.Flags().Uint64(flagHaltTime, 0, "Minimum block time (in Unix seconds) at which to gracefully halt the chain")

	// MinGasPrices
	cmd.Flags().String(flagMinGasPrices, "", "Minimum gas prices to accept for transactions; Any fee in a tx must meet this minimum (e.g. 0.01photon;0.0001stake)")

	// gRPC flags
	cmd.Flags().Bool(flagGRPCEnable, true, "Enable the gRPC server")
	cmd.Flags().String(flagGRPCAddress, DefaultGRPCAddress, "the gRPC server address to listen on")

	// API flags
	cmd.Flags().Bool(flagAPIEnable, true, "Enable the API server")
	cmd.Flags().String(flagAPIAddress, DefaultAPIAddress, "the API server address to listen on")
	cmd.Flags().String(flagRPCListenAddress, DefaultRPCAddress, "the RPC (CometBFT) server address to listen on")
	cmd.Flags().String(flagP2PListenAddress, DefaultP2PAddress, "the P2P server address to listen on")

	// Other options
	cmd.Flags().Bool(flagInterBlockCache, true, "Enable inter-block caching")
	cmd.Flags().Bool(flagTrace, false, "Provide full stack traces for errors in ABCI Log")
	cmd.Flags().IntSlice(flagUnsafeSkipUpgrades, []int{}, "Skip a set of upgrade heights to continue the old binary")

}

type serviceConfig struct {
	grpcAddress string
	apiAddress  string
	grpcEnabled bool
	apiEnabled  bool
}

func loadStartServiceConfig(cmd *cobra.Command, homeDir string, logger log.Logger) (serviceConfig, error) {
	appCfg, cfgLoaded, err := loadAppConfig(homeDir, logger)
	if err != nil {
		return serviceConfig{}, err
	}

	return serviceConfig{
		grpcAddress: resolveStringFlag(cmd, flagGRPCAddress, appCfg.GRPC.Address, cfgLoaded, DefaultGRPCAddress),
		apiAddress:  resolveStringFlag(cmd, flagAPIAddress, appCfg.API.Address, cfgLoaded, DefaultAPIAddress),
		grpcEnabled: resolveBoolFlag(cmd, flagGRPCEnable, appCfg.GRPC.Enable, cfgLoaded, true),
		apiEnabled:  resolveBoolFlag(cmd, flagAPIEnable, appCfg.API.Enable, cfgLoaded, true),
	}, nil
}

func loadAppConfig(homeDir string, logger log.Logger) (serverconfig.Config, bool, error) {
	appConfigPath := filepath.Join(homeDir, "config", "app.toml")
	v := viper.New()
	v.SetConfigFile(appConfigPath)
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			logger.Info("app.toml not found; falling back to defaults", "path", appConfigPath)
			return *serverconfig.DefaultConfig(), false, nil
		}
		return serverconfig.Config{}, false, fmt.Errorf("failed to read app config: %w", err)
	}

	cfg, err := serverconfig.ParseConfig(v)
	if err != nil {
		return serverconfig.Config{}, true, fmt.Errorf("failed to parse app config: %w", err)
	}

	return *cfg, true, nil
}

func resolveStringFlag(cmd *cobra.Command, flagName, configValue string, configLoaded bool, fallback string) string {
	if flag := cmd.Flags().Lookup(flagName); flag != nil && flag.Changed {
		return viper.GetString(flagName)
	}
	if configLoaded && configValue != "" {
		return configValue
	}
	return fallback
}

func resolveBoolFlag(cmd *cobra.Command, flagName string, configValue bool, configLoaded bool, fallback bool) bool {
	if flag := cmd.Flags().Lookup(flagName); flag != nil && flag.Changed {
		return viper.GetBool(flagName)
	}
	if configLoaded {
		return configValue
	}
	return fallback
}

// startInProcess starts the node with CometBFT in the same process
func startInProcess(cmd *cobra.Command, auraApp *app.App, logger log.Logger) error {
	homeDir := GetHomeDir()
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Get security logger
	secLogger := GetSecurityLogger()

	// Drop any genesis gentxs for local single-node starts; multi-node setups
	// should run collect-gentxs outside this path.
	if err := dropGentxs(homeDir, secLogger); err != nil {
		return fmt.Errorf("failed to sanitize gentxs: %w", err)
	}

	// Reconcile staking pools and module accounts before any other processing.
	if err := reconcileGenesisState(homeDir, secLogger); err != nil {
		return fmt.Errorf("failed to reconcile genesis state: %w", err)
	}

	// Seed any KV stores that are missing on-disk versions (development safety net).
	if err := seedMissingStoreVersions(homeDir, logger, app.StoreKeyNames()); err != nil {
		return fmt.Errorf("failed to seed missing store versions: %w", err)
	}

	// Ensure genesis file formatting is compatible with CometBFT expectations
	if err := normalizeGenesisFile(homeDir, secLogger); err != nil {
		return fmt.Errorf("failed to normalize genesis file: %w", err)
	}

	// Load CometBFT configuration
	cmtConfig, err := loadCometConfig(homeDir)
	if err != nil {
		return fmt.Errorf("failed to load CometBFT config: %w", err)
	}

	serviceCfg, err := loadStartServiceConfig(cmd, homeDir, logger)
	if err != nil {
		return fmt.Errorf("failed to load service configuration: %w", err)
	}

	// Validate configuration
	if err := cmtConfig.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid CometBFT config: %w", err)
	}
	if cmd.Flags().Changed(flagRPCListenAddress) {
		cmtConfig.RPC.ListenAddress = viper.GetString(flagRPCListenAddress)
	}
	if cmd.Flags().Changed(flagP2PListenAddress) {
		cmtConfig.P2P.ListenAddress = viper.GetString(flagP2PListenAddress)
	}

	// Create CometBFT logger
	cmtLogger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))
	cmtLogger, err = createCometLogger(cmtConfig.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to create CometBFT logger: %w", err)
	}

	// Open application database
	db, err := openDB(homeDir)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Load genesis file to get chainID (required for SDK v0.53+)
	genesisPath := filepath.Join(homeDir, "config", "genesis.json")
	chainID, err := extractChainIDFromGenesis(genesisPath)
	if err != nil {
		return fmt.Errorf("failed to extract chain-id from genesis: %w", err)
	}
	if flagChainID := viper.GetString(flags.FlagChainID); flagChainID != "" {
		logger.Info("overriding chain-id from flag", "flag_chain_id", flagChainID, "genesis_chain_id", chainID)
		chainID = flagChainID
	}
	logger.Info("loaded chain-id", "chain_id", chainID)

	// Setup trace writer if enabled
	var traceWriter io.WriteCloser
	traceStoreFile := viper.GetString(flagTraceStore)
	if traceStoreFile != "" {
		traceWriter, err = os.OpenFile(
			traceStoreFile,
			os.O_WRONLY|os.O_APPEND|os.O_CREATE,
			0o644,
		)
		if err != nil {
			return fmt.Errorf("failed to open trace store: %w", err)
		}
		defer traceWriter.Close()
		logger.Info("trace store enabled", "file", traceStoreFile)
	}

	// Create app with proper database, tracing, and chainID
	auraApp = createAppWithOptions(logger, db, traceWriter, chainID)

	// Load latest version to initialize stores - REQUIRED before ABCI handshake
	if err := auraApp.LoadLatestVersion(); err != nil {
		return fmt.Errorf("failed to load latest app version: %w", err)
	}

	// Load node key
	nodeKey, err := p2p.LoadOrGenNodeKey(cmtConfig.NodeKeyFile())
	if err != nil {
		return fmt.Errorf("failed to load node key: %w", err)
	}
	logger.Info("loaded node key", "id", nodeKey.ID())

	// Setup genesis doc provider
	genDocProvider := node.DefaultGenesisDocProviderFunc(cmtConfig)

	// Load or generate private validator
	var privValidator cmttypes.PrivValidator
	// Always load private validator for now
	privValidator, err = loadOrGenPrivValidator(cmtConfig)
	if err != nil {
		return fmt.Errorf("failed to load private validator: %w", err)
	}
	logger.Info("loaded private validator")

	// Create ABCI application wrapper
	abciApp := NewCometABCIWrapper(auraApp)

	// Create CometBFT node
	cmtNode, err := node.NewNode(
		cmtConfig,
		privValidator,
		nodeKey,
		proxy.NewLocalClientCreator(abciApp),
		genDocProvider,
		cmtcfg.DefaultDBProvider,
		node.DefaultMetricsProvider(cmtConfig.Instrumentation),
		cmtLogger,
	)
	if err != nil {
		return fmt.Errorf("failed to create CometBFT node: %w", err)
	}

	// Start the CometBFT node
	logger.Info("starting CometBFT node", "chain_id", cmtConfig.BaseConfig.Moniker)
	if err := cmtNode.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	// Start Prometheus metrics HTTP server
	metricsPort := 26660
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsAddr := fmt.Sprintf(":%d", metricsPort)
		logger.Info("starting Prometheus metrics server", "address", metricsAddr)
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			logger.Error("metrics server failed", "error", err)
		}
	}()

	// Create server manager for graceful shutdown
	serverMgr := security.NewServerManager(secLogger)

	// Start gRPC server if enabled
	var grpcSrv *grpc.Server
	if serviceCfg.grpcEnabled {
		logger.Info("starting gRPC server", "address", serviceCfg.grpcAddress)

		grpcSrv, err = startGRPCServer(auraApp, serviceCfg.grpcAddress, serverMgr, logger)
		if err != nil {
			// Try to cleanup
			cmtNode.Stop()
			return fmt.Errorf("failed to start gRPC server: %w", err)
		}
	}

	// Start API server if enabled
	var apiSrv *http.Server
	if serviceCfg.apiEnabled {
		logger.Info("starting API server", "address", serviceCfg.apiAddress)

		apiSrv, err = startAPIServer(serviceCfg.apiAddress, serverMgr, secLogger)
		if err != nil {
			// Try to cleanup
			if grpcSrv != nil {
				grpcSrv.GracefulStop()
			}
			cmtNode.Stop()
			return fmt.Errorf("failed to start API server: %w", err)
		}
	}

	logger.Info("🚀 Aura node started successfully",
		"chain_id", cmtConfig.BaseConfig.Moniker,
		"node_id", nodeKey.ID(),
		"home", homeDir,
	)

	// Setup graceful shutdown
	return waitForShutdown(ctx, cmtNode, grpcSrv, apiSrv, serverMgr, logger)
}

// startStandAlone starts the application without CometBFT (for testing/development)
func startStandAlone(cmd *cobra.Command, auraApp *app.App, logger log.Logger) error {
	homeDir := GetHomeDir()
	secLogger := GetSecurityLogger()

	logger.Info("starting in stand-alone mode (no consensus)")

	// Open database
	db, err := openDB(homeDir)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	serviceCfg, err := loadStartServiceConfig(cmd, homeDir, logger)
	if err != nil {
		return fmt.Errorf("failed to load service configuration: %w", err)
	}

	// Recreate app with proper database
	auraApp = createAppWithDB(logger, db, nil)

	// Create server manager
	serverMgr := security.NewServerManager(secLogger)

	// Start gRPC server
	var grpcSrv *grpc.Server
	if serviceCfg.grpcEnabled {
		grpcSrv, err = startGRPCServer(auraApp, serviceCfg.grpcAddress, serverMgr, logger)
		if err != nil {
			return fmt.Errorf("failed to start gRPC server: %w", err)
		}
		logger.Info("gRPC server started", "address", serviceCfg.grpcAddress)
	}

	// Start API server
	var apiSrv *http.Server
	if serviceCfg.apiEnabled {
		apiSrv, err = startAPIServer(serviceCfg.apiAddress, serverMgr, secLogger)
		if err != nil {
			if grpcSrv != nil {
				grpcSrv.GracefulStop()
			}
			return fmt.Errorf("failed to start API server: %w", err)
		}
		logger.Info("API server started", "address", serviceCfg.apiAddress)
	}

	logger.Info("stand-alone mode started successfully")

	// Wait for shutdown signal
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	return waitForShutdown(ctx, nil, grpcSrv, apiSrv, serverMgr, logger)
}

// startGRPCServer starts the gRPC server
func startGRPCServer(
	auraApp *app.App,
	address string,
	serverMgr *security.ServerManager,
	logger log.Logger,
) (*grpc.Server, error) {
	if auraApp == nil {
		return nil, fmt.Errorf("app instance is nil: cannot start gRPC server")
	}

	// Parse address
	_, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid gRPC address: %w", err)
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	// Create gRPC server options
	grpcOpts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
		grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
	}

	// Add TLS if configured (optional for development)
	homeDir := GetHomeDir()
	tlsConfig := security.NewTLSConfig(homeDir, GetSecurityLogger())
	tlsCfg, err := tlsConfig.LoadTLSConfig()
	if err != nil {
		logger.Warn("TLS not configured for gRPC, using insecure connection", "error", err.Error())
		grpcOpts = append(grpcOpts, grpc.Creds(insecure.NewCredentials()))
	} else {
		logger.Info("gRPC server using TLS")
		creds := credentials.NewTLS(tlsCfg)
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}

	// Create server
	grpcSrv := grpc.NewServer(grpcOpts...)

	// Register ABCI-backed query services with proper SDK context injection so
	// client gRPC calls execute against the latest state.
	auraApp.BaseApp.RegisterGRPCServer(grpcSrv)
	auraApp.SetGRPCServer(grpcSrv)

	// Register with server manager
	serverMgr.RegisterGRPCServer(grpcSrv)

	// Start serving in background
	go func() {
		if err := grpcSrv.Serve(listener); err != nil {
			logger.Error("gRPC server error", "error", err)
		}
	}()

	logger.Info("gRPC server listening", "address", address, "port", portStr)
	return grpcSrv, nil
}

// startAPIServer starts the REST API server
func startAPIServer(
	address string,
	serverMgr *security.ServerManager,
	logger security.Logger,
) (*http.Server, error) {
	// Create rate limiter
	rateLimiter := security.NewDefaultRateLimiter(logger)

	// Create health checker
	healthChecker := security.NewHealthChecker(logger)
	healthChecker.Register("basic", func() error {
		return nil
	})

	// Create mux
	mux := http.NewServeMux()

	// Register routes
	mux.Handle("/health", healthChecker.HTTPHealthHandler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Ignore write error - client may have disconnected
		_, _ = fmt.Fprintf(w, `{"chain":"aura","status":"running","version":"%s"}`, getVersion())
	})

	// Apply middleware
	handler := security.SecurityHeadersMiddleware(mux)
	handler = rateLimiter.RateLimitMiddleware(handler)
	handler = security.TimeoutMiddleware(30 * time.Second)(handler)

	// Create request logger
	reqLogger := security.NewRequestLogger(logger)
	handler = reqLogger.Middleware(handler)

	// Parse address - handle both tcp:// and plain addresses
	listenAddr := address
	if len(address) > 6 && address[:6] == "tcp://" {
		listenAddr = address[6:]
	}

	// Create HTTP server
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	// Register with server manager
	serverMgr.RegisterHTTPServer(server)

	// Start serving in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("API server error: %v", err)
		}
	}()

	logger.Info("API server listening", "address", listenAddr)
	return server, nil
}

// waitForShutdown waits for interrupt signal and performs graceful shutdown
func waitForShutdown(
	ctx context.Context,
	cmtNode *node.Node,
	grpcSrv *grpc.Server,
	apiSrv *http.Server,
	serverMgr *security.ServerManager,
	logger log.Logger,
) error {
	// Setup signal catching
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	// Wait for signal
	select {
	case <-ctx.Done():
		logger.Info("context cancelled, shutting down")
	case sig := <-sigChan:
		logger.Info("received shutdown signal", "signal", sig.String())
	}

	// Perform graceful shutdown
	logger.Info("initiating graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown API server
	if apiSrv != nil {
		logger.Info("shutting down API server")
		if err := apiSrv.Shutdown(shutdownCtx); err != nil {
			logger.Error("API server shutdown error", "error", err)
		}
	}

	// Shutdown gRPC server
	if grpcSrv != nil {
		logger.Info("shutting down gRPC server")
		grpcSrv.GracefulStop()
	}

	// Shutdown CometBFT node
	if cmtNode != nil {
		logger.Info("stopping CometBFT node")
		if err := cmtNode.Stop(); err != nil {
			logger.Error("error stopping CometBFT node", "error", err)
		}
		cmtNode.Wait()
	}

	// Shutdown server manager
	if err := serverMgr.Shutdown(shutdownCtx); err != nil {
		logger.Error("server manager shutdown error", "error", err)
	}

	logger.Info("shutdown completed successfully")
	return nil
}

// loadCometConfig loads the CometBFT configuration from the home directory
func loadCometConfig(homeDir string) (*cmtcfg.Config, error) {
	configPath := filepath.Join(homeDir, "config")

	// Check if config directory exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		cmtConfig := cmtcfg.DefaultConfig()
		cmtConfig.SetRoot(homeDir)
		return cmtConfig, nil
	}

	// Load existing config
	cmtConfig := cmtcfg.DefaultConfig()
	cmtConfig.SetRoot(homeDir)

	// Try to read config.toml
	configFile := filepath.Join(configPath, "config.toml")
	if _, err := os.Stat(configFile); err == nil {
		cfgViper := viper.New()
		cfgViper.SetConfigFile(configFile)
		if err := cfgViper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}

		// Manually set some critical fields from viper
		if chainID := cfgViper.GetString("chain-id"); chainID != "" {
			cmtConfig.BaseConfig.Moniker = chainID
		}
		if moniker := cfgViper.GetString("moniker"); moniker != "" {
			cmtConfig.Moniker = moniker
		}
		if rpcAddr := cfgViper.GetString("rpc.laddr"); rpcAddr != "" {
			cmtConfig.RPC.ListenAddress = rpcAddr
		}
		if pprofAddr := cfgViper.GetString("rpc.pprof_laddr"); pprofAddr != "" {
			cmtConfig.RPC.PprofListenAddress = pprofAddr
		}
		if p2pAddr := cfgViper.GetString("p2p.laddr"); p2pAddr != "" {
			cmtConfig.P2P.ListenAddress = p2pAddr
		}
		if externalAddr := cfgViper.GetString("p2p.external_address"); externalAddr != "" {
			cmtConfig.P2P.ExternalAddress = externalAddr
		}
	}

	return cmtConfig, nil
}

// loadOrGenPrivValidator loads or generates a private validator
func loadOrGenPrivValidator(config *cmtcfg.Config) (cmttypes.PrivValidator, error) {
	keyFile := config.PrivValidatorKeyFile()
	stateFile := config.PrivValidatorStateFile()

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(keyFile), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create validator directory: %w", err)
	}

	// Load or generate file private validator
	pv := pvm.LoadOrGenFilePV(keyFile, stateFile)

	return pv, nil
}

// createCometLogger creates a CometBFT logger with the specified level
func createCometLogger(level string) (cmtlog.Logger, error) {
	logger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))

	// Parse and set log level
	option, err := cmtlog.AllowLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	logger = cmtlog.NewFilter(logger, option)
	return logger, nil
}

// openDB opens the application database
func openDB(homeDir string) (dbm.DB, error) {
	dataDir := filepath.Join(homeDir, "data")

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database
	db, err := dbm.NewGoLevelDB("application", dataDir, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}

// loadBaseAppOptions loads base app options from configuration
func loadBaseAppOptions() []func(*baseapp.BaseApp) {
	var options []func(*baseapp.BaseApp)

	// Pruning options
	pruningStrategy := viper.GetString(flagPruning)
	if pruningStrategy != "" {
		var pruningOpts pruningtypes.PruningOptions
		switch pruningStrategy {
		case pruningtypes.PruningOptionDefault:
			pruningOpts = pruningtypes.NewPruningOptions(pruningtypes.PruningDefault)
		case pruningtypes.PruningOptionNothing:
			pruningOpts = pruningtypes.NewPruningOptions(pruningtypes.PruningNothing)
		case pruningtypes.PruningOptionEverything:
			pruningOpts = pruningtypes.NewPruningOptions(pruningtypes.PruningEverything)
		case pruningtypes.PruningOptionCustom:
			pruningOpts = pruningtypes.NewCustomPruningOptions(
				viper.GetUint64(flagPruningKeepRecent),
				viper.GetUint64(flagPruningInterval),
			)
		}
		options = append(options, baseapp.SetPruning(pruningOpts))
	}

	// Min gas prices
	if minGasPrices := viper.GetString(flagMinGasPrices); minGasPrices != "" {
		options = append(options, baseapp.SetMinGasPrices(minGasPrices))
	}

	// Halt height
	if haltHeight := viper.GetUint64(flagHaltHeight); haltHeight > 0 {
		options = append(options, baseapp.SetHaltHeight(haltHeight))
	}

	// Halt time
	if haltTime := viper.GetUint64(flagHaltTime); haltTime > 0 {
		options = append(options, baseapp.SetHaltTime(haltTime))
	}

	// Inter-block cache
	if interBlockCache := viper.GetBool(flagInterBlockCache); interBlockCache {
		options = append(options, baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager()))
	}

	// Trace
	if trace := viper.GetBool(flagTrace); trace {
		options = append(options, baseapp.SetTrace(true))
	}

	return options
}

// NewCometABCIWrapper wraps the app for CometBFT ABCI
func NewCometABCIWrapper(app *app.App) *CometABCIWrapper {
	return &CometABCIWrapper{app: app}
}

// CometABCIWrapper wraps the Cosmos SDK app for ABCI
type CometABCIWrapper struct {
	app *app.App
}

// Info implements ABCI
func (w *CometABCIWrapper) Info(ctx context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error) {
	return w.app.BaseApp.Info(req)
}

// Query implements ABCI
func (w *CometABCIWrapper) Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	return w.app.BaseApp.Query(ctx, req)
}

// CheckTx implements ABCI
func (w *CometABCIWrapper) CheckTx(ctx context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return w.app.BaseApp.CheckTx(req)
}

// InitChain implements ABCI
func (w *CometABCIWrapper) InitChain(ctx context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	return w.app.BaseApp.InitChain(req)
}

// PrepareProposal implements ABCI
func (w *CometABCIWrapper) PrepareProposal(ctx context.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	return w.app.BaseApp.PrepareProposal(req)
}

// ProcessProposal implements ABCI
func (w *CometABCIWrapper) ProcessProposal(ctx context.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	return w.app.BaseApp.ProcessProposal(req)
}

// FinalizeBlock implements ABCI
func (w *CometABCIWrapper) FinalizeBlock(ctx context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	return w.app.BaseApp.FinalizeBlock(req)
}

// Commit implements ABCI
func (w *CometABCIWrapper) Commit(ctx context.Context, req *abci.RequestCommit) (*abci.ResponseCommit, error) {
	resp, err := w.app.BaseApp.Commit()
	return resp, err
}

// ListSnapshots implements ABCI
func (w *CometABCIWrapper) ListSnapshots(ctx context.Context, req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	return w.app.BaseApp.ListSnapshots(req)
}

// OfferSnapshot implements ABCI
func (w *CometABCIWrapper) OfferSnapshot(ctx context.Context, req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	return w.app.BaseApp.OfferSnapshot(req)
}

// LoadSnapshotChunk implements ABCI
func (w *CometABCIWrapper) LoadSnapshotChunk(ctx context.Context, req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	return w.app.BaseApp.LoadSnapshotChunk(req)
}

// ApplySnapshotChunk implements ABCI
func (w *CometABCIWrapper) ApplySnapshotChunk(ctx context.Context, req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	return w.app.BaseApp.ApplySnapshotChunk(req)
}

// ExtendVote implements ABCI
func (w *CometABCIWrapper) ExtendVote(ctx context.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
	// Return empty vote extension for now
	return &abci.ResponseExtendVote{}, nil
}

// VerifyVoteExtension implements ABCI
func (w *CometABCIWrapper) VerifyVoteExtension(ctx context.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
	// Accept all vote extensions for now
	return &abci.ResponseVerifyVoteExtension{
		Status: abci.ResponseVerifyVoteExtension_ACCEPT,
	}, nil
}

// createAppWithDB creates a new app instance with the given database and trace writer
// Deprecated: Use createAppWithOptions to also specify chainID
func createAppWithDB(logger log.Logger, db dbm.DB, traceWriter io.WriteCloser) *app.App {
	return app.NewAppWithDB(logger, db)
}

// createAppWithOptions creates a new app instance with full configuration
func createAppWithOptions(logger log.Logger, db dbm.DB, traceWriter io.WriteCloser, chainID string) *app.App {
	return app.NewAppWithOptions(logger, db, chainID)
}

// extractChainIDFromGenesis reads the chain-id from the genesis.json file
func extractChainIDFromGenesis(genesisPath string) (string, error) {
	data, err := os.ReadFile(genesisPath)
	if err != nil {
		return "", fmt.Errorf("failed to read genesis file: %w", err)
	}

	var genesis struct {
		ChainID string `json:"chain_id"`
	}
	if err := json.Unmarshal(data, &genesis); err != nil {
		return "", fmt.Errorf("failed to parse genesis file: %w", err)
	}

	if genesis.ChainID == "" {
		return "", fmt.Errorf("chain_id not found in genesis file")
	}

	return genesis.ChainID, nil
}

// getVersion returns the application version
func getVersion() string {
	return "0.1.0"
}

// GetHomeDir returns the configured home directory with path validation
func GetHomeDir() string {
	// Try multiple sources for home directory:
	// Priority order (flag value takes precedence over viper defaults):
	// 1. Package-level homeDir variable from root.go (set by --home flag)
	// 2. Viper with cosmos-sdk flag key
	// 3. Viper with simple "home" key
	// 4. Fall back to default

	// First check the flag variable - this is set directly by cobra flag parsing
	dir := GetHomeDirVar()

	// If not set, try viper
	if dir == "" {
		dir = viper.GetString(flags.FlagHome)
	}
	if dir == "" {
		dir = viper.GetString("home")
	}

	if dir != "" {
		// Validate the home directory path
		logger := security.NewConsoleLogger()
		validator := security.NewPathValidator(logger)
		validPath, err := validator.ValidateAndCleanHomePath(dir)
		if err != nil {
			// Log error but fall back to default
			logger.Error("Invalid home directory path: %v", err)
			return getDefaultHomeDirFromGenesis()
		}
		return validPath
	}
	return getDefaultHomeDirFromGenesis()
}

// GetSecurityLogger returns a security logger instance
func GetSecurityLogger() security.Logger {
	homeDir := GetHomeDir()
	logger, err := security.NewSecurityLogger(homeDir, true)
	if err != nil {
		// Fall back to console logger if file logging fails
		return security.NewConsoleLogger()
	}
	return logger
}

// getDefaultHomeDirFromGenesis returns the default home directory (imported from genesis.go)
func getDefaultHomeDirFromGenesis() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return ".aura"
	}
	return filepath.Join(userHomeDir, ".aura")
}

// normalizeGenesisFile enforces legacy-compatible encoding for fields CometBFT expects.
// Some tooling rewrites genesis.json with numeric initial_height values and nests validators
// under consensus.*, which breaks the default CometBFT loader. This function corrects those issues.
func normalizeGenesisFile(homeDir string, logger security.Logger) error {
	genesisPath := filepath.Join(homeDir, "config", "genesis.json")

	data, err := os.ReadFile(genesisPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to read genesis file: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	doc := make(map[string]interface{})
	if err := decoder.Decode(&doc); err != nil {
		return fmt.Errorf("failed to parse genesis file: %w", err)
	}

	changed := false

	// Ensure initial_height is encoded as a JSON string
	if value, ok := doc["initial_height"]; ok {
		switch v := value.(type) {
		case json.Number:
			doc["initial_height"] = v.String()
			changed = true
		case float64:
			doc["initial_height"] = strconv.FormatInt(int64(v), 10)
			changed = true
		case string:
			// Already encoded properly
		default:
			// Leave other unexpected types untouched
		}
	}

	// Rehydrate legacy validators/consensus_params fields if genesis only contains consensus block
	if consensusRaw, ok := doc["consensus"]; ok {
		if consensus, ok := consensusRaw.(map[string]interface{}); ok {
			if _, hasValidators := doc["validators"]; !hasValidators {
				if consensusValidators, ok := consensus["validators"]; ok {
					doc["validators"] = consensusValidators
					changed = true
				}
			}
			if _, hasParams := doc["consensus_params"]; !hasParams {
				if consensusParams, ok := consensus["params"]; ok {
					doc["consensus_params"] = consensusParams
					changed = true
				}
			}
		}
	}

	if !changed {
		return nil
	}

	normalizedData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal normalized genesis file: %w", err)
	}

	if err := os.WriteFile(genesisPath, normalizedData, security.ConfigFilePerms); err != nil {
		return fmt.Errorf("failed to write normalized genesis file: %w", err)
	}
	if err := os.Chmod(genesisPath, security.ConfigFilePerms); err != nil {
		return fmt.Errorf("failed to set permissions on genesis file: %w", err)
	}

	logger.Info("normalized genesis metadata in %s", genesisPath)
	return nil
}

// reconcileGenesisState ensures module accounts exist and staking pool balances line up with bonded tokens.
func reconcileGenesisState(homeDir string, logger security.Logger) error {
	genesisPath := filepath.Join(homeDir, "config", "genesis.json")
	if _, err := os.Stat(genesisPath); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	appGenesis, err := genutiltypes.AppGenesisFromFile(genesisPath)
	if err != nil {
		return fmt.Errorf("failed to load genesis: %w", err)
	}

	var genState map[string]json.RawMessage
	if err := json.Unmarshal(appGenesis.AppState, &genState); err != nil {
		return fmt.Errorf("failed to decode app_state: %w", err)
	}

	encCfg := app.MakeEncodingConfig()

	authGenesis := authtypes.GetGenesisStateFromAppState(encCfg.Codec, genState)
	bankGenesis := banktypes.GetGenesisStateFromAppState(encCfg.Codec, genState)
	stakingGenesis := stakingtypes.GetGenesisStateFromAppState(encCfg.Codec, genState)
	genutilGenesis := genutiltypes.GetGenesisStateFromAppState(encCfg.Codec, genState)
	var slashingGenesis *slashingtypes.GenesisState
	switch {
	case len(genState[slashingtypes.ModuleName]) == 0:
		slashingGenesis = slashingtypes.DefaultGenesisState()
	default:
		slashingGenesis = &slashingtypes.GenesisState{}
		if err := encCfg.Codec.UnmarshalJSON(genState[slashingtypes.ModuleName], slashingGenesis); err != nil {
			return fmt.Errorf("failed to decode slashing genesis: %w", err)
		}
	}
	var contractGenesis contractregistrypb.GenesisState

	switch {
	case len(genState[contractregistrytypes.ModuleName]) == 0:
		contractGenesis = *contractregistrytypes.DefaultGenesis()
	default:
		if err := encCfg.Codec.UnmarshalJSON(genState[contractregistrytypes.ModuleName], &contractGenesis); err != nil {
			return fmt.Errorf("failed to decode contractregistry genesis: %w", err)
		}
	}
	if contractGenesis.Params == nil {
		contractGenesis.Params = contractregistrytypes.DefaultParams()
		logger.Info("populated default contractregistry params")
	}
	genState[contractregistrytypes.ModuleName] = encCfg.Codec.MustMarshalJSON(&contractGenesis)

	bondDenom := stakingGenesis.Params.BondDenom
	requiredModuleAccounts := map[string][]string{
		authtypes.FeeCollectorName:     nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}

	for name, perms := range requiredModuleAccounts {
		if !moduleAccountExists(&authGenesis, encCfg.Codec, name) {
			ma := authtypes.NewEmptyModuleAccount(name, perms...)
			any, err := codectypes.NewAnyWithValue(ma)
			if err != nil {
				return fmt.Errorf("failed to wrap module account %s: %w", name, err)
			}
			authGenesis.Accounts = append(authGenesis.Accounts, any)
			logger.Info("added missing module account", "module", name)
		}
	}

	powerReduction := sdk.DefaultPowerReduction
	newConsensusVals := make([]cmttypes.GenesisValidator, 0, len(stakingGenesis.Validators))
	newLastPowers := make([]stakingtypes.LastValidatorPower, 0, len(stakingGenesis.Validators))
	var totalPower int64
	for _, val := range stakingGenesis.Validators {
		pk, err := val.ConsPubKey()
		if err != nil {
			return fmt.Errorf("failed to unpack validator consensus key: %w", err)
		}
		cmtPk, err := cryptocodec.ToCmtPubKeyInterface(pk)
		if err != nil {
			return fmt.Errorf("failed to convert validator consensus key: %w", err)
		}
		power := val.ConsensusPower(powerReduction)
		totalPower += power

		newConsensusVals = append(newConsensusVals, cmttypes.GenesisValidator{
			Address: sdk.ConsAddress(cmtPk.Address()).Bytes(),
			PubKey:  cmtPk,
			Power:   power,
			Name:    val.Description.Moniker,
		})
		newLastPowers = append(newLastPowers, stakingtypes.LastValidatorPower{
			Address: val.OperatorAddress,
			Power:   power,
		})
	}
	stakingGenesis.LastValidatorPowers = newLastPowers
	stakingGenesis.LastTotalPower = sdkmath.NewInt(totalPower)
	if appGenesis.Consensus == nil {
		appGenesis.Consensus = &genutiltypes.ConsensusGenesis{}
	}
	appGenesis.Consensus.Validators = newConsensusVals

	if slashingGenesis == nil {
		slashingGenesis = slashingtypes.DefaultGenesisState()
	}
	existingSigningInfo := make(map[string]bool, len(slashingGenesis.SigningInfos))
	for _, info := range slashingGenesis.SigningInfos {
		existingSigningInfo[info.Address] = true
	}
	for _, val := range stakingGenesis.Validators {
		consAddr, err := val.GetConsAddr()
		if err != nil {
			return fmt.Errorf("failed to derive consensus address: %w", err)
		}
		addrStr, err := sdk.Bech32ifyAddressBytes(sdk.GetConfig().GetBech32ConsensusAddrPrefix(), consAddr)
		if err != nil {
			return fmt.Errorf("failed to encode consensus address: %w", err)
		}
		if existingSigningInfo[addrStr] {
			continue
		}
		slashingGenesis.SigningInfos = append(slashingGenesis.SigningInfos, slashingtypes.SigningInfo{
			Address: addrStr,
			ValidatorSigningInfo: slashingtypes.ValidatorSigningInfo{
				StartHeight:         0,
				IndexOffset:         0,
				JailedUntil:         time.Time{},
				Tombstoned:          false,
				MissedBlocksCounter: 0,
			},
		})
	}

	// Build a mutable balance map.
	balanceMap := make(map[string]*banktypes.Balance, len(bankGenesis.Balances)+len(requiredModuleAccounts))
	for i := range bankGenesis.Balances {
		bal := &bankGenesis.Balances[i]
		balanceCopy := *bal
		balanceMap[bal.Address] = &balanceCopy
	}

	ensureBalance := func(addr string) *banktypes.Balance {
		if bal, ok := balanceMap[addr]; ok {
			return bal
		}
		newBal := &banktypes.Balance{Address: addr, Coins: sdk.NewCoins()}
		balanceMap[addr] = newBal
		return newBal
	}

	// Move delegated stake from delegator balances into the bonded pool.
	for _, del := range stakingGenesis.Delegations {
		amt := del.Shares.TruncateInt()
		if amt.IsZero() {
			continue
		}
		bal := ensureBalance(del.DelegatorAddress)
		current := bal.Coins.AmountOf(bondDenom)
		newAmt := current.Sub(amt)
		if newAmt.IsNegative() {
			newAmt = sdkmath.ZeroInt()
		}
		setCoinAmount(bal, bondDenom, newAmt)
	}

	// Sum bonded tokens from bonded validators.
	bondedTokens := sdkmath.NewInt(0)
	notBondedTokens := sdkmath.NewInt(0)
	for _, val := range stakingGenesis.Validators {
		if val.Status == stakingtypes.Bonded {
			bondedTokens = bondedTokens.Add(val.Tokens)
		} else {
			notBondedTokens = notBondedTokens.Add(val.Tokens)
		}
	}

	for _, ubd := range stakingGenesis.UnbondingDelegations {
		for _, entry := range ubd.Entries {
			notBondedTokens = notBondedTokens.Add(entry.Balance)
		}
	}

	bondedAddr := authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String()
	notBondedAddr := authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName).String()

	// Place bonded tokens into the bonded pool and non-bonded tokens into the not bonded pool.
	setCoinAmount(ensureBalance(bondedAddr), bondDenom, bondedTokens)
	setCoinAmount(ensureBalance(notBondedAddr), bondDenom, notBondedTokens)

	// Rebuild balances and supply.
	balances := make([]banktypes.Balance, 0, len(balanceMap))
	supplyByDenom := make(map[string]sdkmath.Int)
	for _, bal := range balanceMap {
		bal.Coins = bal.Coins.Sort()
		balances = append(balances, *bal)
		for _, coin := range bal.Coins {
			if coin.Amount.IsNegative() {
				continue
			}
			if _, ok := supplyByDenom[coin.Denom]; !ok {
				supplyByDenom[coin.Denom] = sdkmath.ZeroInt()
			}
			supplyByDenom[coin.Denom] = supplyByDenom[coin.Denom].Add(coin.Amount)
		}
	}

	sort.Slice(balances, func(i, j int) bool {
		return balances[i].Address < balances[j].Address
	})

	newSupply := sdk.NewCoins()
	for denom, amt := range supplyByDenom {
		if amt.IsZero() {
			continue
		}
		newSupply = newSupply.Add(sdk.NewCoin(denom, amt))
	}

	bankGenesis.Balances = balances
	bankGenesis.Supply = newSupply.Sort()

	genState[authtypes.ModuleName] = encCfg.Codec.MustMarshalJSON(&authGenesis)
	genState[banktypes.ModuleName] = encCfg.Codec.MustMarshalJSON(bankGenesis)
	genState[stakingtypes.ModuleName] = encCfg.Codec.MustMarshalJSON(stakingGenesis)
	genState[slashingtypes.ModuleName] = encCfg.Codec.MustMarshalJSON(slashingGenesis)
	if genutilGenesis == nil {
		genutilGenesis = genutiltypes.DefaultGenesisState()
	}
	if len(genutilGenesis.GenTxs) > 0 {
		genutilGenesis.GenTxs = []json.RawMessage{}
		logger.Info("removed genutil gentxs for local start (using patched staking state)")
	}
	genState[genutiltypes.ModuleName] = encCfg.Codec.MustMarshalJSON(genutilGenesis)

	newAppState, err := json.Marshal(genState)
	if err != nil {
		return fmt.Errorf("failed to marshal patched app_state: %w", err)
	}

	appGenesis.AppState = newAppState
	if err := genutil.ExportGenesisFile(appGenesis, genesisPath); err != nil {
		return fmt.Errorf("failed to write patched genesis: %w", err)
	}

	logger.Info("reconciled module accounts and staking pools in %s", genesisPath)
	return nil
}

// dropGentxs removes any collected gentxs for local single-node runs to avoid
// signature verification errors when no validator keyring context is present.
func dropGentxs(homeDir string, logger security.Logger) error {
	genesisPath := filepath.Join(homeDir, "config", "genesis.json")
	if _, err := os.Stat(genesisPath); errors.Is(err, os.ErrNotExist) {
		return nil
	}

	appGenesis, err := genutiltypes.AppGenesisFromFile(genesisPath)
	if err != nil {
		return fmt.Errorf("failed to load genesis: %w", err)
	}

	var genState map[string]json.RawMessage
	if err := json.Unmarshal(appGenesis.AppState, &genState); err != nil {
		return fmt.Errorf("failed to decode app_state: %w", err)
	}

	encCfg := app.MakeEncodingConfig()
	genutilGenesis := genutiltypes.GetGenesisStateFromAppState(encCfg.Codec, genState)
	if genutilGenesis == nil {
		genutilGenesis = genutiltypes.DefaultGenesisState()
	}
	if len(genutilGenesis.GenTxs) == 0 {
		return nil
	}

	genutilGenesis.GenTxs = []json.RawMessage{}
	genState[genutiltypes.ModuleName] = encCfg.Codec.MustMarshalJSON(genutilGenesis)

	newAppState, err := json.Marshal(genState)
	if err != nil {
		return fmt.Errorf("failed to marshal patched app_state: %w", err)
	}

	appGenesis.AppState = newAppState
	if err := genutil.ExportGenesisFile(appGenesis, genesisPath); err != nil {
		return fmt.Errorf("failed to write patched genesis: %w", err)
	}

	logger.Info("removed gentxs from %s for local start", genesisPath)
	return nil
}

func moduleAccountExists(gs *authtypes.GenesisState, cdc codectypes.AnyUnpacker, name string) bool {
	for _, acc := range gs.Accounts {
		var ma authtypes.ModuleAccount
		if err := cdc.UnpackAny(acc, &ma); err == nil && ma.Name == name {
			return true
		}
	}
	return false
}

func setCoinAmount(bal *banktypes.Balance, denom string, amount sdkmath.Int) {
	coins := bal.Coins
	filtered := sdk.NewCoins()
	for _, coin := range coins {
		if coin.Denom == denom {
			continue
		}
		filtered = filtered.Add(coin)
	}
	if amount.IsPositive() {
		filtered = filtered.Add(sdk.NewCoin(denom, amount))
	}
	bal.Coins = filtered
}

// seedMissingStoreVersions ensures every mounted store has an on-disk version.
// Some dev builds produced commit infos that referenced stores without persisted
// IAVL versions, causing "version does not exist" failures when loading
// CacheMultiStoreWithVersion. This function detects those stores and seeds an
// empty tree at the expected version so the node can start.
func seedMissingStoreVersions(homeDir string, logger log.Logger, expectedStores []string) error {
	dataDir := filepath.Join(homeDir, "data")
	db, err := dbm.NewGoLevelDB("application", dataDir, nil)
	if err != nil {
		// If the DB doesn't exist yet (fresh init), nothing to do.
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer db.Close()

	latestBz, err := db.Get([]byte("s/latest"))
	if err != nil || len(latestBz) == 0 {
		// nothing to seed on a brand new DB
		return nil
	}
	var ci storetypes.CommitInfo
	if err := ci.Unmarshal(latestBz); err != nil {
		logger.Warn("failed to parse latest commit info; skipping store seeding", "error", err)
		return nil
	}

	ciKey := []byte(fmt.Sprintf("s/%d", ci.Version))
	if ciBz, err := db.Get(ciKey); err == nil && len(ciBz) > 0 {
		var existing storetypes.CommitInfo
		if err := existing.Unmarshal(ciBz); err == nil {
			ci = existing
		}
	}

	updated := false
	for i, si := range ci.StoreInfos {
		prefix := []byte("s/k:" + si.Name + "/")
		pdb := dbm.NewPrefixDB(db, prefix)
		tree := iavl.NewMutableTree(storewrapper.NewDBWrapper(pdb), 0, true, log.NewNopLogger())

		if _, err := tree.LoadVersion(si.CommitId.GetVersion()); err == nil {
			continue
		}

		version := si.CommitId.GetVersion()
		tree.SetInitialVersion(uint64(version))
		_, _ = tree.Set([]byte("seed"), []byte{}) // ensure a version is written
		if _, _, err := tree.SaveVersion(); err != nil {
			return fmt.Errorf("seed store %s: %w", si.Name, err)
		}

		ci.StoreInfos[i].CommitId.Version = version
		ci.StoreInfos[i].CommitId.Hash = tree.Hash()
		updated = true
		logger.Info("seeded missing store version", "store", si.Name, "version", version)
	}

	if !updated {
		return nil
	}

	newCIBz, err := ci.Marshal()
	if err != nil {
		return err
	}
	if err := db.Set(ciKey, newCIBz); err != nil {
		return err
	}
	if err := db.Set([]byte("s/latest"), newCIBz); err != nil {
		return err
	}

	logger.Info("updated commit info to include seeded stores", "version", ci.Version)
	return nil
}
