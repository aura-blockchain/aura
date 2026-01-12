//go:build mvp

// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	tmlog "cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"
	upgrademodule "cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankmodule "github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensus "github.com/cosmos/cosmos-sdk/x/consensus"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutilmodule "github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	params "github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingmodule "github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingmodule "github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"

	// MVP AURA modules
	"github.com/aequitas/aura/chain/x/compliance"
	compliancekeeper "github.com/aequitas/aura/chain/x/compliance/keeper"
	compliancetypes "github.com/aequitas/aura/chain/x/compliance/types"
	"github.com/aequitas/aura/chain/x/dataregistry"
	drkeeper "github.com/aequitas/aura/chain/x/dataregistry/keeper"
	drparams "github.com/aequitas/aura/chain/x/dataregistry/params"
	drtypes "github.com/aequitas/aura/chain/x/dataregistry/types"
	"github.com/aequitas/aura/chain/x/governance"
	governancekeeper "github.com/aequitas/aura/chain/x/governance/keeper"
	governancetypes "github.com/aequitas/aura/chain/x/governance/types"
	"github.com/aequitas/aura/chain/x/identity"
	identitykeeper "github.com/aequitas/aura/chain/x/identity/keeper"
	identitytypes "github.com/aequitas/aura/chain/x/identity/types"
	"github.com/aequitas/aura/chain/x/prevalidation"
	prevalidationkeeper "github.com/aequitas/aura/chain/x/prevalidation/keeper"
	prevalidationtypes "github.com/aequitas/aura/chain/x/prevalidation/types"
	"github.com/aequitas/aura/chain/x/vcregistry"
	vckeeper "github.com/aequitas/aura/chain/x/vcregistry/keeper"
	vcparams "github.com/aequitas/aura/chain/x/vcregistry/params"
	vctypes "github.com/aequitas/aura/chain/x/vcregistry/types"
	"github.com/aequitas/aura/chain/x/wasm"
	wasmSecurityKeeper "github.com/aequitas/aura/chain/x/wasm/keeper"
	wasmSecurityTypes "github.com/aequitas/aura/chain/x/wasm/types"
)

// moduleAccountPermissionsMVP defines permissions for MVP modules only.
// Security: Only wasm has minter permission for MVP.
var moduleAccountPermissionsMVP = map[string][]string{
	authtypes.FeeCollectorName:     nil,
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	distrtypes.ModuleName:          nil,
	governancetypes.ModuleName:     {authtypes.Burner},
	wasmtypes.ModuleName:           {authtypes.Minter, authtypes.Burner},
}

// allowedReceivingModulesMVP defines which MVP modules can receive tokens.
var allowedReceivingModulesMVP = map[string]bool{
	governancetypes.ModuleName: true,
	wasmtypes.ModuleName:       true,
}

// AppMVP is the simplified AURA application for MVP release.
// It includes only the 12 essential modules.
type AppMVP struct {
	*baseapp.BaseApp

	moduleManager      *sdkmodule.Manager
	moduleConfigurator sdkmodule.Configurator
	encoding           EncodingConfig

	// Cosmos SDK keepers
	AccountKeeper      authkeeper.AccountKeeper
	BankKeeper         bankkeeper.BaseKeeper
	StakingKeeper      *stakingkeeper.Keeper
	SlashingKeeper     slashingkeeper.Keeper
	DistributionKeeper distrkeeper.Keeper
	ConsensusKeeper    consensuskeeper.Keeper
	UpgradeKeeper      *upgradekeeper.Keeper
	WasmKeeper         *wasmkeeper.Keeper

	// MVP AURA module keepers
	identityKeeper      *identitykeeper.Keeper
	vcKeeper            *vckeeper.Keeper
	drKeeper            *drkeeper.Keeper
	complianceKeeper    *compliancekeeper.Keeper
	prevalidationKeeper *prevalidationkeeper.Keeper
	governanceKeeper    *governancekeeper.Keeper
	wasmSecurityKeeper  wasmSecurityKeeper.Keeper

	storeKeys *storeKeysMVP
	memKeys   struct {
		vc *storetypes.MemoryStoreKey
	}
	transientKeys struct {
		params *storetypes.TransientStoreKey
	}
}

// NewAppMVP builds the AURA MVP application with only essential modules.
func NewAppMVP(logger tmlog.Logger, db dbm.DB, chainID string) *AppMVP {
	ensureSDKConfig()

	if logger == nil {
		logger = tmlog.NewNopLogger()
	}

	encoding := MakeEncodingConfigMVP()

	if db == nil {
		db = dbm.NewMemDB()
	}

	baseAppOptions := []func(*baseapp.BaseApp){
		baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing)),
		baseapp.SetIAVLDisableFastNode(true),
		baseapp.SetIAVLCacheSize(10000),
	}
	if chainID != "" {
		baseAppOptions = append(baseAppOptions, baseapp.SetChainID(chainID))
	}

	base := baseapp.NewBaseApp(appName, logger, db, encoding.TxConfig.TxDecoder(), baseAppOptions...)
	base.SetInterfaceRegistry(encoding.InterfaceRegistry)
	base.SetTxEncoder(encoding.TxConfig.TxEncoder())

	// Initialize MVP store keys
	keys := initStoreKeysMVP()

	// Memory and transient keys
	vcMemKey := storetypes.NewMemoryStoreKey(vctypes.MemStoreKey)
	paramsTKey := storetypes.NewTransientStoreKey(paramstypes.TStoreKey)

	// Mount stores
	base.MountKVStores(keys.AsMap())
	base.MountTransientStores(map[string]*storetypes.TransientStoreKey{
		paramstypes.TStoreKey: paramsTKey,
	})
	base.MountMemoryStores(map[string]*storetypes.MemoryStoreKey{
		vctypes.MemStoreKey: vcMemKey,
	})

	paramsKeeper := paramskeeper.NewKeeper(encoding.Codec, codec.NewLegacyAmino(), keys.params, paramsTKey)

	accountCodec := address.NewBech32Codec(Bech32MainPrefix)
	validatorCodec := address.NewBech32Codec(Bech32ValidatorPrefix)
	consensusCodec := address.NewBech32Codec(Bech32ConsensusPrefix)
	authorityAddr := authtypes.NewModuleAddress(governancetypes.ModuleName).String()

	accountKeeper := authkeeper.NewAccountKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.account),
		authtypes.ProtoBaseAccount,
		moduleAccountPermissionsMVP,
		accountCodec,
		Bech32MainPrefix,
		authorityAddr,
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.bank),
		accountKeeper,
		blockedModuleAddressesMVP(moduleAccountPermissionsMVP),
		authorityAddr,
		logger,
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.staking),
		accountKeeper,
		bankKeeper,
		authorityAddr,
		validatorCodec,
		consensusCodec,
	)

	slashingKeeper := slashingkeeper.NewKeeper(
		encoding.Codec,
		codec.NewLegacyAmino(),
		runtime.NewKVStoreService(keys.slashing),
		stakingKeeper,
		authorityAddr,
	)

	distributionKeeper := distrkeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.distribution),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authorityAddr,
	)

	consensusKeeper := consensuskeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.consensus),
		authorityAddr,
		runtime.EventService{},
	)
	base.SetParamStore(consensusKeeper.ParamsStore)

	upgradeKeeper := upgradekeeper.NewKeeper(
		map[int64]bool{},
		runtime.NewKVStoreService(keys.upgrade),
		encoding.Codec,
		filepath.Join("/tmp", "aura-mvp-upgrades"),
		base,
		authorityAddr,
	)

	logger.Info("initializing MVP keepers", "phase", "compliance")

	// MVP AURA module keepers
	complianceKeeper := compliancekeeper.NewKeeper(encoding.Codec, keys.compliance)
	// Note: bankAdapter available for future compliance-monitored transfers
	_ = newMonitoredBankKeeperAdapter(bankKeeper, complianceKeeper)

	logger.Info("initializing MVP keepers", "phase", "identity")

	identityKeeper := identitykeeper.NewKeeper(
		runtime.NewKVStoreService(keys.identity),
		keys.identity,
		encoding.Codec,
		authorityAddr,
		logger.With("module", identitytypes.ModuleName),
	)

	logger.Info("initializing MVP keepers", "phase", "dataregistry")

	drParamsStore := drparams.NewStore(drtypes.DefaultParams())
	drKeeper := drkeeper.NewKeeper(
		runtime.NewKVStoreService(keys.dataRegistry),
		encoding.Codec,
		drParamsStore,
		authorityAddr,
		logger.With("module", drtypes.ModuleName),
	)

	logger.Info("initializing MVP keepers", "phase", "governance")

	// Governance keeper - using nil for security keeper since it's deferred
	governanceKeeper := governancekeeper.NewKeeper(encoding.Codec, keys.governance, stakingKeeper, bankKeeper, nil)

	logger.Info("initializing MVP keepers", "phase", "vcregistry")

	vcParamsStore := vcparams.NewStore(*vctypes.DefaultParams())
	vcKeeper := vckeeper.NewKeeperBuilder(vcParamsStore, authorityAddr).
		WithStore(keys.vc, encoding.Codec).
		Build()

	logger.Info("initializing MVP keepers", "phase", "prevalidation")

	prevalidationKeeper := prevalidationkeeper.NewKeeper(
		encoding.Codec,
		keys.prevalidation,
	)
	prevalidationKeeper.SetLogger(logger.With("module", prevalidationtypes.ModuleName))
	prevalidationKeeper.SetComplianceKeeper(complianceKeeper)

	logger.Info("initializing MVP keepers", "phase", "wasm")

	// WASM configuration
	wasmConfig := wasmtypes.DefaultWasmConfig()
	wasmConfig.SmartQueryGasLimit = 100000
	wasmConfig.MemoryCacheSize = 0

	wasmKeeper := wasmkeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.wasm),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		nil, // DistributionKeeper
		nil, // IBCKeeper
		nil, // ChannelKeeper
		nil, // PortKeeper
		nil, // ScopedWasmKeeper
		nil, // TransferKeeper
		base.MsgServiceRouter(),
		base.GRPCQueryRouter(),
		filepath.Join("/tmp", "wasm-mvp"),
		wasmConfig,
		"iterator",
		authorityAddr,
	)

	wasmSecurityKeeperInstance := wasmSecurityKeeper.NewKeeper(
		encoding.Codec,
		keys.wasmSecurity,
		&wasmKeeper,
		authorityAddr,
	)

	logger.Info("initializing MVP modules")

	// Create MVP modules
	identityModule := identity.NewAppModule(encoding.Codec, identityKeeper)
	vcModule := vcregistry.NewAppModule(vcKeeper)
	dataModule := dataregistry.NewAppModule(drKeeper)
	complianceModule := compliance.NewAppModule(complianceKeeper)
	governanceModule := governance.NewAppModule(governanceKeeper)
	prevalidationModule := prevalidation.NewAppModule(encoding.Codec, prevalidationKeeper)
	wasmSecurityModule := wasm.NewAppModule(encoding.Codec, wasmSecurityKeeperInstance)

	// Initialize metrics
	logger.Info("initializing Prometheus metrics for MVP modules")
	_ = identitykeeper.NewIdentityMetrics()
	logger.Info("Prometheus metrics initialized successfully")

	// Core SDK modules
	coreModules := []sdkmodule.AppModule{
		auth.NewAppModule(encoding.Codec, accountKeeper, nil, nil),
		bankmodule.NewAppModule(encoding.Codec, bankKeeper, accountKeeper, nil),
		stakingmodule.NewAppModule(encoding.Codec, stakingKeeper, accountKeeper, bankKeeper, nil),
		slashingmodule.NewAppModule(encoding.Codec, slashingKeeper, accountKeeper, bankKeeper, stakingKeeper, nil, encoding.InterfaceRegistry),
		distribution.NewAppModule(encoding.Codec, distributionKeeper, accountKeeper, bankKeeper, stakingKeeper, nil),
		params.NewAppModule(paramsKeeper),
		consensus.NewAppModule(encoding.Codec, consensusKeeper),
		upgrademodule.NewAppModule(upgradeKeeper, accountCodec),
		genutilmodule.NewAppModule(accountKeeper, stakingKeeper, base, encoding.TxConfig),
		wasmSecurityModule,
	}

	// MVP AURA modules
	auraModules := []sdkmodule.AppModule{
		identityModule,
		vcModule,
		dataModule,
		complianceModule,
		governanceModule,
		prevalidationModule,
	}

	moduleManager := sdkmodule.NewManager(append(coreModules, auraModules...)...)

	// Set genesis order for MVP modules only
	moduleManager.SetOrderInitGenesis(
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		consensustypes.ModuleName,
		upgradetypes.ModuleName,
		paramstypes.ModuleName,
		genutiltypes.ModuleName,
		wasmtypes.ModuleName,
		wasmSecurityTypes.ModuleName,
		identitytypes.ModuleName,
		vctypes.ModuleName,
		drtypes.ModuleName,
		compliancetypes.ModuleName,
		governancetypes.ModuleName,
		prevalidationtypes.ModuleName,
	)

	// Set begin blocker order for MVP
	moduleManager.SetOrderBeginBlockers(
		genutiltypes.ModuleName,
		consensustypes.ModuleName,
		upgradetypes.ModuleName,
		slashingtypes.ModuleName,
		stakingtypes.ModuleName,
		distrtypes.ModuleName,
		wasmtypes.ModuleName,
		wasmSecurityTypes.ModuleName,
		identitytypes.ModuleName,
		vctypes.ModuleName,
		drtypes.ModuleName,
		compliancetypes.ModuleName,
		governancetypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		paramstypes.ModuleName,
		prevalidationtypes.ModuleName,
	)

	// Set end blocker order for MVP
	moduleManager.SetOrderEndBlockers(
		upgradetypes.ModuleName,
		slashingtypes.ModuleName,
		stakingtypes.ModuleName,
		distrtypes.ModuleName,
		wasmtypes.ModuleName,
		wasmSecurityTypes.ModuleName,
		identitytypes.ModuleName,
		vctypes.ModuleName,
		drtypes.ModuleName,
		compliancetypes.ModuleName,
		governancetypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		paramstypes.ModuleName,
		genutiltypes.ModuleName,
		prevalidationtypes.ModuleName,
	)

	configurator := sdkmodule.NewConfigurator(encoding.Codec, base.MsgServiceRouter(), base.GRPCQueryRouter())
	moduleManager.RegisterServices(configurator)

	app := &AppMVP{
		BaseApp:             base,
		moduleManager:       moduleManager,
		moduleConfigurator:  configurator,
		encoding:            encoding,
		AccountKeeper:       accountKeeper,
		BankKeeper:          bankKeeper,
		StakingKeeper:       stakingKeeper,
		SlashingKeeper:      slashingKeeper,
		DistributionKeeper:  distributionKeeper,
		ConsensusKeeper:     consensusKeeper,
		UpgradeKeeper:       upgradeKeeper,
		WasmKeeper:          &wasmKeeper,
		identityKeeper:      identityKeeper,
		vcKeeper:            vcKeeper,
		drKeeper:            drKeeper,
		complianceKeeper:    complianceKeeper,
		prevalidationKeeper: prevalidationKeeper,
		governanceKeeper:    governanceKeeper,
		wasmSecurityKeeper:  wasmSecurityKeeperInstance,
		storeKeys:           keys,
		memKeys: struct {
			vc *storetypes.MemoryStoreKey
		}{vc: vcMemKey},
		transientKeys: struct {
			params *storetypes.TransientStoreKey
		}{params: paramsTKey},
	}

	app.SetPreBlocker(func(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
		return moduleManager.PreBlock(ctx)
	})
	app.SetInitChainer(func(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
		var genesisState map[string]json.RawMessage
		if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
			return nil, err
		}
		if err := canonicalizeAuthGenesis(encoding.Codec, genesisState); err != nil {
			return nil, err
		}
		res, err := moduleManager.InitGenesis(ctx, encoding.Codec, genesisState)
		if err != nil {
			return nil, err
		}
		ensureStoreInitMarkersMVP(ctx, app.allStoreKeysMVP())
		return res, nil
	})
	app.SetBeginBlocker(moduleManager.BeginBlock)
	app.SetEndBlocker(moduleManager.EndBlock)
	app.SetPrepareCheckStater(func(ctx sdk.Context) {
		if err := moduleManager.PrepareCheckState(ctx); err != nil {
			ctx.Logger().Error("prepare-check-state failed", "error", err)
			panic(err)
		}
	})
	app.SetPrecommiter(func(ctx sdk.Context) {
		if err := moduleManager.Precommit(ctx); err != nil {
			ctx.Logger().Error("precommit failed", "error", err)
			panic(err)
		}
	})

	app.SetupAnteHandlerMVP()

	// Set pruning on CommitMultiStore
	if cms, ok := base.CommitMultiStore().(interface {
		SetPruning(pruningtypes.PruningOptions)
	}); ok {
		cms.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
		logger.Info("pruning explicitly set to PruningNothing on CommitMultiStore")
	}

	logger.Info("MVP app initialized successfully", "modules", len(MVPModules))

	return app
}

// SetupAnteHandlerMVP configures the ante handler for MVP transaction processing.
func (a *AppMVP) SetupAnteHandlerMVP() {
	signModeHandler := a.encoding.TxConfig.SignModeHandler()
	wasmConfig := wasmtypes.DefaultWasmConfig()

	anteHandler, err := NewAnteHandler(HandlerOptions{
		AccountKeeper:         a.AccountKeeper,
		BankKeeper:            a.BankKeeper,
		SignModeHandler:       signModeHandler,
		FeegrantKeeper:        nil,
		SigGasConsumer:        nil,
		TxFeeChecker:          nil,
		WasmConfig:            &wasmConfig,
		WasmKeeper:            a.WasmKeeper,
		TXCounterStoreService: runtime.NewKVStoreService(a.storeKeys.wasm),
		ComplianceKeeper:      a.complianceKeeper,
		WalletSecurityKeeper:  nil, // Deferred module
	})

	if err != nil {
		panic(fmt.Errorf("failed to create MVP ante handler: %w", err))
	}

	a.SetAnteHandler(anteHandler)
}

// Logger returns the app's logger.
func (a *AppMVP) Logger() tmlog.Logger {
	return a.BaseApp.Logger()
}

// Encoding returns the codec configuration.
func (a *AppMVP) Encoding() EncodingConfig {
	return a.encoding
}

// allStoreKeysMVP returns all MVP store keys for validation.
func (a *AppMVP) allStoreKeysMVP() []storetypes.StoreKey {
	return []storetypes.StoreKey{
		a.storeKeys.account,
		a.storeKeys.bank,
		a.storeKeys.staking,
		a.storeKeys.slashing,
		a.storeKeys.distribution,
		a.storeKeys.params,
		a.storeKeys.consensus,
		a.storeKeys.upgrade,
		a.storeKeys.identity,
		a.storeKeys.vc,
		a.storeKeys.dataRegistry,
		a.storeKeys.compliance,
		a.storeKeys.prevalidation,
		a.storeKeys.governance,
		a.storeKeys.wasm,
		a.storeKeys.wasmSecurity,
	}
}

// ensureStoreInitMarkersMVP writes markers to each KV store for MVP.
func ensureStoreInitMarkersMVP(ctx sdk.Context, keys []storetypes.StoreKey) {
	marker := []byte{0xFF, 'I', 'N', 'I', 'T'}
	for _, key := range keys {
		if key == nil {
			continue
		}
		store := ctx.KVStore(key)
		if store == nil {
			continue
		}
		if !store.Has(marker) {
			store.Set(marker, marker)
		}
	}
}

// blockedModuleAddressesMVP returns the bank blocklist for MVP module accounts.
func blockedModuleAddressesMVP(perms map[string][]string) map[string]bool {
	blocked := make(map[string]bool)

	names := make([]string, 0, len(perms))
	for name := range perms {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		addr := authtypes.NewModuleAddress(name).String()
		blocked[addr] = !allowedReceivingModulesMVP[name]
	}
	return blocked
}

// MakeEncodingConfigMVP builds the encoding config for MVP.
func MakeEncodingConfigMVP() EncodingConfig {
	ensureSDKConfig()

	addrCodec := address.NewBech32Codec(Bech32MainPrefix)
	valCodec := address.NewBech32Codec(Bech32ValidatorPrefix)
	interfaceRegistry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: txsigning.Options{
			AddressCodec:          addrCodec,
			ValidatorAddressCodec: valCodec,
		},
	})
	if err != nil {
		panic(fmt.Errorf("failed to create interface registry: %w", err))
	}

	cryptocodec.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)
	slashingtypes.RegisterInterfaces(interfaceRegistry)
	distrtypes.RegisterInterfaces(interfaceRegistry)
	upgradetypes.RegisterInterfaces(interfaceRegistry)
	wasmtypes.RegisterInterfaces(interfaceRegistry)

	// Register MVP module interfaces
	ModuleBasicsMVP.RegisterInterfaces(interfaceRegistry)

	protoCodec := codec.NewProtoCodec(interfaceRegistry)

	signModes := []signingtypes.SignMode{
		signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		signingtypes.SignMode_SIGN_MODE_DIRECT,
		signingtypes.SignMode_SIGN_MODE_DIRECT_AUX,
	}

	txConfig, err := tx.NewTxConfigWithOptions(protoCodec, tx.ConfigOptions{
		EnabledSignModes: signModes,
		SigningOptions: &txsigning.Options{
			FileResolver:          proto.HybridResolver,
			AddressCodec:          addrCodec,
			ValidatorAddressCodec: valCodec,
		},
	})
	if err != nil {
		panic(fmt.Errorf("failed to build tx config: %w", err))
	}

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             protoCodec,
		TxConfig:          txConfig,
	}
}

// RegisterGRPCServicesMVP wires MVP module services to gRPC.
func (a *AppMVP) RegisterGRPCServicesMVP() {
	authtypes.RegisterQueryServer(a.GRPCQueryRouter(), authkeeper.NewQueryServer(a.AccountKeeper))
	banktypes.RegisterQueryServer(a.GRPCQueryRouter(), bankkeeper.NewQuerier(&a.BankKeeper))
	wasmSecurityTypes.RegisterQueryServer(a.GRPCQueryRouter(), wasmSecurityKeeper.NewQueryServerImpl(a.wasmSecurityKeeper))
}

// RegisterTxServiceMVP wires the Tx gRPC service for MVP.
func (a *AppMVP) RegisterTxServiceMVP(clientCtx client.Context) {
	tx.RegisterTxService(a.GRPCQueryRouter(), clientCtx, a.BaseApp.Simulate, a.encoding.InterfaceRegistry)
}

// canonicalizeAuthGenesis is shared with full app - uses the same implementation
// from app.go (non-MVP build won't have this file, so we need local copy)
func canonicalizeAuthGenesisMVP(cdc codec.JSONCodec, genesis map[string]json.RawMessage) error {
	raw := genesis[authtypes.ModuleName]
	if len(raw) == 0 {
		return nil
	}

	var authGen authtypes.GenesisState
	if err := cdc.UnmarshalJSON(raw, &authGen); err != nil {
		return fmt.Errorf("failed to unmarshal auth genesis: %w", err)
	}

	accounts, err := authtypes.UnpackAccounts(authGen.Accounts)
	if err != nil {
		return fmt.Errorf("failed to unpack auth accounts: %w", err)
	}

	sort.Slice(accounts, func(i, j int) bool {
		return bytes.Compare(accounts[i].GetAddress().Bytes(), accounts[j].GetAddress().Bytes()) < 0
	})

	for idx, acc := range accounts {
		if err := acc.SetAccountNumber(uint64(idx)); err != nil {
			return fmt.Errorf("failed to set account number for %s: %w", acc.GetAddress().String(), err)
		}
	}

	packed, err := authtypes.PackAccounts(accounts)
	if err != nil {
		return fmt.Errorf("failed to re-pack auth accounts: %w", err)
	}
	authGen.Accounts = packed

	updatedRaw, err := cdc.MarshalJSON(&authGen)
	if err != nil {
		return fmt.Errorf("failed to marshal canonical auth genesis: %w", err)
	}
	genesis[authtypes.ModuleName] = updatedRaw
	return nil
}
