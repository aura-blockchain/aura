package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	pruningtypes "cosmossdk.io/store/pruning/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	abci "github.com/cometbft/cometbft/abci/types"
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

	"github.com/aequitas/aura/chain/app"
	"github.com/aequitas/aura/chain/cmd/aurad/cmd/security"
)

const (
	// Flags for node configuration
	flagWithComet              = "with-comet"
	flagAddress                = "address"
	flagTransport              = "transport"
	flagTraceStore             = "trace-store"
	flagCPUProfile             = "cpu-profile"
	flagPruning                = "pruning"
	flagPruningKeepRecent      = "pruning-keep-recent"
	flagPruningInterval        = "pruning-interval"
	flagMinGasPrices           = "minimum-gas-prices"
	flagHaltHeight             = "halt-height"
	flagHaltTime               = "halt-time"
	flagInterBlockCache        = "inter-block-cache"
	flagUnsafeSkipUpgrades     = "unsafe-skip-upgrades"
	flagTrace                  = "trace"
	flagInvCheckPeriod         = "inv-check-period"
	flagAPIEnable              = "api.enable"
	flagAPIAddress             = "api.address"
	flagGRPCEnable             = "grpc.enable"
	flagGRPCAddress            = "grpc.address"
	flagGRPCWebEnable          = "grpc-web.enable"
	flagJSONRPCEnable          = "json-rpc.enable"
	flagJSONRPCAddress         = "json-rpc.address"

	// Default values
	DefaultGRPCAddress         = "localhost:9090"
	DefaultAPIAddress          = "tcp://localhost:1317"
	DefaultRPCAddress          = "tcp://localhost:26657"
	DefaultABCIAddress         = "tcp://0.0.0.0:26658"
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

	// Other options
	cmd.Flags().Bool(flagInterBlockCache, true, "Enable inter-block caching")
	cmd.Flags().Bool(flagTrace, false, "Provide full stack traces for errors in ABCI Log")
	cmd.Flags().IntSlice(flagUnsafeSkipUpgrades, []int{}, "Skip a set of upgrade heights to continue the old binary")

	// Home directory
	cmd.Flags().String(flags.FlagHome, getDefaultHomeDirFromGenesis(), "The application home directory")
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

	// Load CometBFT configuration
	cmtConfig, err := loadCometConfig(homeDir)
	if err != nil {
		return fmt.Errorf("failed to load CometBFT config: %w", err)
	}

	// Validate configuration
	if err := cmtConfig.ValidateBasic(); err != nil {
		return fmt.Errorf("invalid CometBFT config: %w", err)
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
	logger.Info("loaded chain-id from genesis", "chain_id", chainID)

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

	// Create server manager for graceful shutdown
	serverMgr := security.NewServerManager(secLogger)

	// Start gRPC server if enabled
	var grpcSrv *grpc.Server
	if viper.GetBool(flagGRPCEnable) {
		grpcAddress := viper.GetString(flagGRPCAddress)
		logger.Info("starting gRPC server", "address", grpcAddress)

		grpcSrv, err = startGRPCServer(auraApp, grpcAddress, serverMgr, logger)
		if err != nil {
			// Try to cleanup
			cmtNode.Stop()
			return fmt.Errorf("failed to start gRPC server: %w", err)
		}
	}

	// Start API server if enabled
	var apiSrv *http.Server
	if viper.GetBool(flagAPIEnable) {
		apiAddress := viper.GetString(flagAPIAddress)
		logger.Info("starting API server", "address", apiAddress)

		apiSrv, err = startAPIServer(apiAddress, serverMgr, secLogger)
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

	// Recreate app with proper database
	auraApp = createAppWithDB(logger, db, nil)

	// Create server manager
	serverMgr := security.NewServerManager(secLogger)

	// Start gRPC server
	var grpcSrv *grpc.Server
	if viper.GetBool(flagGRPCEnable) {
		grpcAddress := viper.GetString(flagGRPCAddress)
		grpcSrv, err = startGRPCServer(auraApp, grpcAddress, serverMgr, logger)
		if err != nil {
			return fmt.Errorf("failed to start gRPC server: %w", err)
		}
		logger.Info("gRPC server started", "address", grpcAddress)
	}

	// Start API server
	var apiSrv *http.Server
	if viper.GetBool(flagAPIEnable) {
		apiAddress := viper.GetString(flagAPIAddress)
		apiSrv, err = startAPIServer(apiAddress, serverMgr, secLogger)
		if err != nil {
			if grpcSrv != nil {
				grpcSrv.GracefulStop()
			}
			return fmt.Errorf("failed to start API server: %w", err)
		}
		logger.Info("API server started", "address", apiAddress)
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

	// Register services from the app - use the app's module manager
	// The app already has a grpc server, but we need to register services on our new one
	// For now, we'll use the app's built-in gRPC server handling
	// Note: In a full implementation, you'd register all module query services here

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
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}

		// Manually set some critical fields from viper
		if chainID := viper.GetString("chain-id"); chainID != "" {
			cmtConfig.BaseConfig.Moniker = chainID
		}
		if moniker := viper.GetString("moniker"); moniker != "" {
			cmtConfig.Moniker = moniker
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
