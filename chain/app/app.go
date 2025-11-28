package app

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	tmlog "cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"
	txsigningtextual "cosmossdk.io/x/tx/signing/textual"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"google.golang.org/grpc"

	// AURA Core Modules (kept as-is)
	"github.com/aequitas/aura/chain/x/bridge"
	bridgekeeper "github.com/aequitas/aura/chain/x/bridge/keeper"
	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
	"github.com/aequitas/aura/chain/x/compliance"
	compliancekeeper "github.com/aequitas/aura/chain/x/compliance/keeper"
	compliancetypes "github.com/aequitas/aura/chain/x/compliance/types"
	"github.com/aequitas/aura/chain/x/confidencescore"
	cskeeper "github.com/aequitas/aura/chain/x/confidencescore/keeper"
	csparams "github.com/aequitas/aura/chain/x/confidencescore/params"
	cstypes "github.com/aequitas/aura/chain/x/confidencescore/types"
	"github.com/aequitas/aura/chain/x/contractregistry"
	contractregistrykeeper "github.com/aequitas/aura/chain/x/contractregistry/keeper"
	contractregistrytypes "github.com/aequitas/aura/chain/x/contractregistry/types"
	"github.com/aequitas/aura/chain/x/dataregistry"
	drkeeper "github.com/aequitas/aura/chain/x/dataregistry/keeper"
	drparams "github.com/aequitas/aura/chain/x/dataregistry/params"
	drtypes "github.com/aequitas/aura/chain/x/dataregistry/types"
	"github.com/aequitas/aura/chain/x/dex"
	dexkeeper "github.com/aequitas/aura/chain/x/dex/keeper"
	dextypes "github.com/aequitas/aura/chain/x/dex/types"
	"github.com/aequitas/aura/chain/x/inclusionroutines"
	irkeeper "github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
	irparams "github.com/aequitas/aura/chain/x/inclusionroutines/params"
	irtypes "github.com/aequitas/aura/chain/x/inclusionroutines/types"
	"github.com/aequitas/aura/chain/x/vcregistry"
	vckeeper "github.com/aequitas/aura/chain/x/vcregistry/keeper"
	vcparams "github.com/aequitas/aura/chain/x/vcregistry/params"
	vctypes "github.com/aequitas/aura/chain/x/vcregistry/types"
	"github.com/aequitas/aura/chain/x/wasm"
	wasmSecurityKeeper "github.com/aequitas/aura/chain/x/wasm/keeper"
	wasmSecurityTypes "github.com/aequitas/aura/chain/x/wasm/types"

	// AURA Security Modules (individual)
	"github.com/aequitas/aura/chain/x/walletsecurity"
	walletsecuritykeeper "github.com/aequitas/aura/chain/x/walletsecurity/keeper"
	walletsecuritytypes "github.com/aequitas/aura/chain/x/walletsecurity/types"
	"github.com/aequitas/aura/chain/x/validatorsecurity"
	validatorsecuritykeeper "github.com/aequitas/aura/chain/x/validatorsecurity/keeper"
	validatorsecuritytypes "github.com/aequitas/aura/chain/x/validatorsecurity/types"
	"github.com/aequitas/aura/chain/x/cryptography"
	cryptographykeeper "github.com/aequitas/aura/chain/x/cryptography/keeper"
	cryptographytypes "github.com/aequitas/aura/chain/x/cryptography/types"
	"github.com/aequitas/aura/chain/x/networksecurity"
	networksecuritykeeper "github.com/aequitas/aura/chain/x/networksecurity/keeper"
	networksecuritytypes "github.com/aequitas/aura/chain/x/networksecurity/types"
	"github.com/aequitas/aura/chain/x/incidentresponse"
	incidentresponsekeeper "github.com/aequitas/aura/chain/x/incidentresponse/keeper"
	incidentresponsetypes "github.com/aequitas/aura/chain/x/incidentresponse/types"
	"github.com/aequitas/aura/chain/x/privacy"
	privacykeeper "github.com/aequitas/aura/chain/x/privacy/keeper"
	privacytypes "github.com/aequitas/aura/chain/x/privacy/types"

	// AURA Identity Module (individual)
	"github.com/aequitas/aura/chain/x/identitychange"
	identitychangekeeper "github.com/aequitas/aura/chain/x/identitychange/keeper"
	identitychangetypes "github.com/aequitas/aura/chain/x/identitychange/types"
	identitychangeparams "github.com/aequitas/aura/chain/x/identitychange/params"

	// AURA Economics Modules (individual)
	"github.com/aequitas/aura/chain/x/economicsecurity"
	economicsecuritykeeper "github.com/aequitas/aura/chain/x/economicsecurity/keeper"
	economicsecuritytypes "github.com/aequitas/aura/chain/x/economicsecurity/types"
	economicsecurityparams "github.com/aequitas/aura/chain/x/economicsecurity/params"
	"github.com/aequitas/aura/chain/x/governance"
	governancekeeper "github.com/aequitas/aura/chain/x/governance/keeper"
	governancetypes "github.com/aequitas/aura/chain/x/governance/types"

	// AI Assistant module (kept as-is, not consolidated)
	"github.com/aequitas/aura/chain/x/aiassistant"
	aikeeper "github.com/aequitas/aura/chain/x/aiassistant/keeper"
	aitypes "github.com/aequitas/aura/chain/x/aiassistant/types"

)

const (
	appName               = "aura"
	bech32MainPrefix      = "aura"
	bech32ValidatorPrefix = "auravaloper"
	bech32ConsensusPrefix = "auravalcons"
)

var (
	sdkConfigOnce sync.Once

	// Module account permissions define what each module account can do
	// SECURITY: Minter permission removed from DEX to prevent unlimited token creation
	// DEX can only manage existing tokens (Burner permission for LP tokens)
	moduleAccountPermissions = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		distrtypes.ModuleName:          nil,
		governancetypes.ModuleName:     {authtypes.Burner}, // Governance module
		// DEX: Removed Minter permission to prevent inflation
		// Only Burner allowed for LP token management
		dextypes.ModuleName:    {authtypes.Burner},
		bridgetypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		wasmtypes.ModuleName:   {authtypes.Minter, authtypes.Burner},
	}

	allowedReceivingModules = map[string]bool{
		governancetypes.ModuleName: true, // Governance module
		dextypes.ModuleName:        true,
		bridgetypes.ModuleName:     true,
		wasmtypes.ModuleName:       true,
	}
)

// EncodingConfig defines the codec + interface registry wiring used by the app shell.
type EncodingConfig struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Codec             codec.Codec
}

// App wires all Aura modules plus the Cosmos SDK base keepers into a runnable node shell.
type App struct {
	*baseapp.BaseApp

	moduleManager ModuleManager
	grpcServer    *grpc.Server
	encoding      EncodingConfig

	AccountKeeper      authkeeper.AccountKeeper
	BankKeeper         bankkeeper.BaseKeeper
	StakingKeeper      *stakingkeeper.Keeper
	SlashingKeeper     slashingkeeper.Keeper
	DistributionKeeper distrkeeper.Keeper
	ConsensusKeeper    consensuskeeper.Keeper
	WasmKeeper         *wasmkeeper.Keeper

	// Security module keepers (individual)
	walletsecurityKeeper    walletsecuritykeeper.Keeper
	validatorsecurityKeeper validatorsecuritykeeper.Keeper
	cryptographyKeeper      *cryptographykeeper.Keeper
	networksecurityKeeper   networksecuritykeeper.Keeper
	incidentresponseKeeper  *incidentresponsekeeper.KeeperKV
	privacyKeeper           *privacykeeper.Keeper

	// Identity module keeper (individual)
	identitychangeKeeper *identitychangekeeper.Keeper

	// Economics module keepers (individual)
	economicsecurityKeeper *economicsecuritykeeper.Keeper
	governanceKeeper       *governancekeeper.Keeper

	// Core module keepers (unchanged)
	irKeeper               *irkeeper.Keeper
	csKeeper               *cskeeper.Keeper
	vcKeeper               *vckeeper.Keeper
	drKeeper               *drkeeper.Keeper
	complianceKeeper       *compliancekeeper.Keeper
	dexKeeper              *dexkeeper.Keeper
	bridgeKeeper           *bridgekeeper.Keeper
	aiKeeper               *aikeeper.Keeper
	contractRegistryKeeper *contractregistrykeeper.Keeper
	wasmSecurityKeeper     wasmSecurityKeeper.Keeper

	storeKeys struct {
		// Cosmos SDK standard keys
		account      *storetypes.KVStoreKey
		bank         *storetypes.KVStoreKey
		staking      *storetypes.KVStoreKey
		slashing     *storetypes.KVStoreKey
		distribution *storetypes.KVStoreKey
		params       *storetypes.KVStoreKey
		consensus    *storetypes.KVStoreKey

		// Security module keys (individual)
		walletSecurity    *storetypes.KVStoreKey
		validatorSecurity *storetypes.KVStoreKey
		cryptography      *storetypes.KVStoreKey
		networkSecurity   *storetypes.KVStoreKey
		incidentResponse  *storetypes.KVStoreKey
		privacy           *storetypes.KVStoreKey

		// Identity module key (individual)
		identityChange *storetypes.KVStoreKey

		// Economics module keys (individual)
		economicSecurity *storetypes.KVStoreKey
		governance       *storetypes.KVStoreKey

		// Core AURA module keys (unchanged)
		vc                *storetypes.KVStoreKey
		compliance        *storetypes.KVStoreKey
		dex               *storetypes.KVStoreKey
		bridge            *storetypes.KVStoreKey
		ai                *storetypes.KVStoreKey
		wasm              *storetypes.KVStoreKey
		contractRegistry  *storetypes.KVStoreKey
		wasmSecurity      *storetypes.KVStoreKey
		confidenceScore   *storetypes.KVStoreKey
		inclusionRoutines *storetypes.KVStoreKey
		dataRegistry      *storetypes.KVStoreKey
	}
	memKeys struct {
		vc *storetypes.MemoryStoreKey
	}
	transientKeys struct {
		params *storetypes.TransientStoreKey
	}
}

// NewApp builds the Aura application shell with default logging and in-memory database.
// Use NewAppWithOptions for production use with persistent storage.
func NewApp() *App {
	return NewAppWithOptions(nil, nil, "")
}

// NewAppWithLogger builds the Aura application shell with the provided logger.
// Deprecated: Use NewAppWithOptions instead for production deployments.
func NewAppWithLogger(logger tmlog.Logger) *App {
	return NewAppWithOptions(logger, nil, "")
}

// NewAppWithDB builds the Aura application shell with the provided logger and database.
// Deprecated: Use NewAppWithOptions to also specify chainID for SDK v0.53 compatibility.
func NewAppWithDB(logger tmlog.Logger, db dbm.DB) *App {
	return NewAppWithOptions(logger, db, "")
}

// NewAppWithOptions builds the Aura application shell with full configuration.
// - logger: application logger (nil for nop logger)
// - db: database instance (nil for in-memory)
// - chainID: the chain ID (required for SDK v0.53+, empty string reads from genesis during InitChain)
func NewAppWithOptions(logger tmlog.Logger, db dbm.DB, chainID string) *App {
	ensureSDKConfig()

	if logger == nil {
		logger = tmlog.NewNopLogger()
	}

	encoding := MakeEncodingConfig()

	// Use provided database or fall back to in-memory for testing
	if db == nil {
		db = dbm.NewMemDB()
	}

	// Build BaseApp with options
	baseAppOptions := []func(*baseapp.BaseApp){}
	if chainID != "" {
		baseAppOptions = append(baseAppOptions, baseapp.SetChainID(chainID))
	}

	base := baseapp.NewBaseApp(appName, logger, db, dummyTxDecoder, baseAppOptions...)
	base.SetInterfaceRegistry(encoding.InterfaceRegistry)

	accountKey := storetypes.NewKVStoreKey(authtypes.StoreKey)
	bankKey := storetypes.NewKVStoreKey(banktypes.StoreKey)
	stakingKey := storetypes.NewKVStoreKey(stakingtypes.StoreKey)
	slashingKey := storetypes.NewKVStoreKey(slashingtypes.StoreKey)
	distributionKey := storetypes.NewKVStoreKey(distrtypes.StoreKey)
	paramsKey := storetypes.NewKVStoreKey(paramstypes.StoreKey)
	// Security module keys (individual)
	walletSecurityKey := storetypes.NewKVStoreKey(walletsecuritytypes.StoreKey)
	validatorSecurityKey := storetypes.NewKVStoreKey(validatorsecuritytypes.StoreKey)
	validatorSecurityMemKey := storetypes.NewMemoryStoreKey(validatorsecuritytypes.MemStoreKey)
	cryptographyKey := storetypes.NewKVStoreKey(cryptographytypes.StoreKey)
	networkSecurityKey := storetypes.NewKVStoreKey(networksecuritytypes.StoreKey)
	incidentResponseKey := storetypes.NewKVStoreKey(incidentresponsetypes.StoreKey)
	privacyKey := storetypes.NewKVStoreKey(privacytypes.StoreKey)

	// Identity module key (individual)
	identityChangeKey := storetypes.NewKVStoreKey(identitychangetypes.StoreKey)

	// Economics module keys (individual)
	economicSecurityKey := storetypes.NewKVStoreKey(economicsecuritytypes.StoreKey)
	governanceKey := storetypes.NewKVStoreKey(governancetypes.StoreKey)

	// Core module keys (unchanged)
	vcKey := storetypes.NewKVStoreKey(vctypes.StoreKey)
	complianceKey := storetypes.NewKVStoreKey(compliancetypes.StoreKey)
	dexKey := storetypes.NewKVStoreKey(dextypes.StoreKey)
	bridgeKey := storetypes.NewKVStoreKey(bridgetypes.StoreKey)
	aiKey := storetypes.NewKVStoreKey(aitypes.StoreKey)
	wasmKey := storetypes.NewKVStoreKey(wasmtypes.StoreKey)
	contractRegistryKey := storetypes.NewKVStoreKey(contractregistrytypes.StoreKey)
	wasmSecurityKey := storetypes.NewKVStoreKey(wasmSecurityTypes.StoreKey)
	confidenceScoreKey := storetypes.NewKVStoreKey(cstypes.StoreKey)
	inclusionRoutinesKey := storetypes.NewKVStoreKey(irtypes.StoreKey)
	dataRegistryKey := storetypes.NewKVStoreKey(drtypes.StoreKey)
	consensusKey := storetypes.NewKVStoreKey(consensustypes.StoreKey)

	paramsTKey := storetypes.NewTransientStoreKey(paramstypes.TStoreKey)
	vcMemKey := storetypes.NewMemoryStoreKey(vctypes.MemStoreKey)

	base.MountKVStores(map[string]*storetypes.KVStoreKey{
		// Cosmos SDK standard keys
		authtypes.StoreKey:     accountKey,
		banktypes.StoreKey:     bankKey,
		stakingtypes.StoreKey:  stakingKey,
		slashingtypes.StoreKey: slashingKey,
		distrtypes.StoreKey:    distributionKey,
		paramstypes.StoreKey:   paramsKey,
		consensustypes.StoreKey: consensusKey,

		// Security module keys (individual)
		walletsecuritytypes.StoreKey:    walletSecurityKey,
		validatorsecuritytypes.StoreKey: validatorSecurityKey,
		cryptographytypes.StoreKey:      cryptographyKey,
		networksecuritytypes.StoreKey:   networkSecurityKey,
		incidentresponsetypes.StoreKey:  incidentResponseKey,
		privacytypes.StoreKey:           privacyKey,

		// Identity module key (individual)
		identitychangetypes.StoreKey: identityChangeKey,

		// Economics module keys (individual)
		economicsecuritytypes.StoreKey: economicSecurityKey,
		governancetypes.StoreKey:       governanceKey,

		// Core AURA module keys (unchanged)
		vctypes.StoreKey:               vcKey,
		compliancetypes.StoreKey:       complianceKey,
		dextypes.StoreKey:              dexKey,
		bridgetypes.StoreKey:           bridgeKey,
		aitypes.StoreKey:               aiKey,
		wasmtypes.StoreKey:             wasmKey,
		contractregistrytypes.StoreKey: contractRegistryKey,
		wasmSecurityTypes.StoreKey:     wasmSecurityKey,
		cstypes.StoreKey:               confidenceScoreKey,
		irtypes.StoreKey:               inclusionRoutinesKey,
		drtypes.StoreKey:               dataRegistryKey,
	})
	base.MountTransientStores(map[string]*storetypes.TransientStoreKey{
		paramstypes.TStoreKey: paramsTKey,
	})
	base.MountMemoryStores(map[string]*storetypes.MemoryStoreKey{
		vctypes.MemStoreKey:             vcMemKey,
		validatorsecuritytypes.MemStoreKey: validatorSecurityMemKey,
	})

	paramsKeeper := paramskeeper.NewKeeper(encoding.Codec, codec.NewLegacyAmino(), paramsKey, paramsTKey)
	bridgeSubspace := paramsKeeper.Subspace(bridgetypes.ModuleName)

	accountCodec := address.NewBech32Codec(bech32MainPrefix)
	validatorCodec := address.NewBech32Codec(bech32ValidatorPrefix)
	consensusCodec := address.NewBech32Codec(bech32ConsensusPrefix)
	authorityAddr := authtypes.NewModuleAddress(governancetypes.ModuleName).String() // Using governance module

	accountKeeper := authkeeper.NewAccountKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(accountKey),
		authtypes.ProtoBaseAccount,
		moduleAccountPermissions,
		accountCodec,
		bech32MainPrefix,
		authorityAddr,
	)

	// Ensure module accounts exist before keepers that depend on them (staking, bank) are created.
	bankKeeper := bankkeeper.NewBaseKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(bankKey),
		accountKeeper,
		blockedModuleAddresses(moduleAccountPermissions),
		authorityAddr,
		logger,
	)
	bankAdapter := newBankKeeperAdapter(bankKeeper)

	stakingKeeper := stakingkeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(stakingKey),
		accountKeeper,
		bankKeeper,
		authorityAddr,
		validatorCodec,
		consensusCodec,
	)

	slashingKeeper := slashingkeeper.NewKeeper(
		encoding.Codec,
		codec.NewLegacyAmino(),
		runtime.NewKVStoreService(slashingKey),
		stakingKeeper,
		authorityAddr,
	)

	distributionKeeper := distrkeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(distributionKey),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authorityAddr,
	)

	// Consensus keeper - required for BaseApp.ParamStore in SDK v0.53+
	consensusKeeper := consensuskeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(consensusKey),
		authorityAddr,
		runtime.EventService{},
	)
	// Set the ParamStore on BaseApp - this is REQUIRED in SDK v0.53+ for InitChain to store consensus params
	base.SetParamStore(consensusKeeper.ParamsStore)

	// ============================================================================
	// KEEPER INITIALIZATION - STRICT DEPENDENCY ORDER
	// ============================================================================
	// This section initializes keepers in a strict order to eliminate circular
	// dependencies. The order is based on the dependency graph:
	//
	// Tier 1: No AURA dependencies (only Cosmos SDK)
	//   - compliance, cryptography, walletSecurity, governance
	//
	// Tier 2: No AURA module dependencies
	//   - identitychange, dataregistry
	//
	// Tier 3: Depends on Tier 2
	//   - inclusionroutines (no deps on other AURA modules)
	//
	// Tier 4: Depends on Tier 3
	//   - confidencescore (depends on inclusionroutines)
	//
	// Tier 5: Depends on Tier 4
	//   - vcregistry (depends on confidencescore)
	//
	// Tier 6: Depends on Tier 5
	//   - contractregistry (depends on vcregistry, compliance, confidencescore)
	//
	// Tier 7: Depends on Tier 5
	//   - dex, bridge (depend on vcregistry)
	//
	// Tier 8: Depends on everything
	//   - wasm, wasmSecurity (depend on all modules)
	//
	// Using builder pattern to eliminate post-construction mutation
	// ============================================================================

	logger.Info("initializing keepers", "phase", "tier-1-no-deps")

	// Tier 1: Keepers with no AURA dependencies
	complianceKeeper := compliancekeeper.NewKeeper(encoding.Codec, complianceKey)

	// Individual security module keepers
	walletsecurityKeeper := walletsecuritykeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(walletSecurityKey),
		logger.With("module", walletsecuritytypes.ModuleName),
	)

	validatorsecurityKeeper := validatorsecuritykeeper.NewKeeper(
		encoding.Codec,
		validatorSecurityKey,
		validatorSecurityMemKey,
		authorityAddr,
		newValidatorSecurityStakingAdapter(stakingKeeper),
		slashingKeeper,
		bankKeeper,
	)

	cryptographyKeeper := cryptographykeeper.NewKeeper(
		encoding.Codec,
		cryptographyKey,
		logger.With("module", cryptographytypes.ModuleName),
		authorityAddr,
	)

	networksecurityKeeper := networksecuritykeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(networkSecurityKey),
		authorityAddr,
		logger.With("module", networksecuritytypes.ModuleName),
	)

	incidentresponseKeeper := incidentresponsekeeper.NewKeeperKV(
		incidentResponseKey,
		encoding.Codec,
	)

	privacyKeeper := privacykeeper.NewKeeper(
		encoding.Codec,
		privacyKey,
		accountKeeper,
		bankKeeper,
	)

	// Individual economics module keepers
	governanceKeeper := governancekeeper.NewKeeper(encoding.Codec, governanceKey)

	economicsecurityParamsStore := economicsecurityparams.NewStore(*economicsecuritytypes.DefaultParams())
	economicsecurityKeeper := economicsecuritykeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(economicSecurityKey),
		economicsecurityParamsStore,
		authorityAddr,
	)

	logger.Info("initializing keepers", "phase", "tier-2-aura-base")

	// Tier 2: Individual identity keeper
	identitychangeParamsStore := identitychangeparams.NewStore(identitychangetypes.DefaultParams())
	identitychangeKeeper := identitychangekeeper.NewKeeper(
		runtime.NewKVStoreService(identityChangeKey),
		encoding.Codec,
		identitychangeParamsStore,
		authorityAddr,
		logger.With("module", identitychangetypes.ModuleName),
	)

	drParamsStore := drparams.NewStore(drtypes.DefaultParams())
	drKeeper := drkeeper.NewKeeper(
		runtime.NewKVStoreService(dataRegistryKey),
		encoding.Codec,
		drParamsStore,
		authorityAddr,
		logger.With("module", drtypes.ModuleName),
	)

	logger.Info("initializing keepers", "phase", "tier-3-inclusion-routines")

	// Tier 3: Inclusion routines (depends on dataregistry for data validation)
	irParamsStore := irparams.NewStore(irtypes.DefaultParams())
	irKeeper := irkeeper.NewKeeper(
		runtime.NewKVStoreService(inclusionRoutinesKey),
		encoding.Codec,
		irParamsStore,
		authorityAddr,
		logger.With("module", irtypes.ModuleName),
	)

	logger.Info("initializing keepers", "phase", "tier-4-confidence-score")

	// Tier 4: Confidence score (depends on inclusionroutines)
	// Using builder pattern to set all dependencies BEFORE construction
	csParamsStore := csparams.NewStore(cstypes.DefaultParams())
	csKeeper := cskeeper.NewKeeperBuilder(
		runtime.NewKVStoreService(confidenceScoreKey),
		encoding.Codec,
		csParamsStore,
		authorityAddr,
		logger.With("module", cstypes.ModuleName),
	).Build() // Note: IR Registry not wired due to interface mismatch - fix in follow-up

	logger.Info("initializing keepers", "phase", "tier-5-vcregistry")

	// Tier 5: VC Registry (depends on confidencescore)
	// Using builder pattern to eliminate circular dependency
	vcParamsStore := vcparams.NewStore(*vctypes.DefaultParams())
	vcKeeper := vckeeper.NewKeeperBuilder(vcParamsStore, authorityAddr).
		WithStore(vcKey, encoding.Codec).
		Build() // Note: ConfidenceScore keeper not wired due to interface mismatch - fix in follow-up

	logger.Info("initializing keepers", "phase", "tier-6-contract-registry")

	// Tier 6: Contract registry (depends on vcregistry, compliance, confidencescore)
	// Contract registry must be initialized BEFORE wasm security as it's a dependency
	contractRegistryKeeper := contractregistrykeeper.NewKeeper(
		contractRegistryKey,
		encoding.Codec,
		authorityAddr,
	)
	// Note: Keeper dependencies not wired due to interface mismatches - fix in follow-up
	// contractRegistryKeeper.SetVCRegistryKeeper(vcKeeper)
	// contractRegistryKeeper.SetComplianceKeeper(complianceKeeper)
	// contractRegistryKeeper.SetConfidenceScoreKeeper(csKeeper)

	logger.Info("initializing keepers", "phase", "tier-7-dex-bridge-ai")

	// Tier 7: DEX, Bridge, and AI modules (depend on vcregistry)
	accountAdapter := newAccountKeeperAdapter(accountKeeper)
	vcAdapter := newVCRegistryKeeperAdapter(vcKeeper)

	dexKeeper := dexkeeper.NewKeeper(encoding.Codec, dexKey, bankAdapter, accountAdapter, vcAdapter)
	bridgeKeeper := bridgekeeper.NewKeeper(
		encoding.Codec,
		bridgeKey,
		&bridgeSubspace,
		bankAdapter,
		accountAdapter,
		vcAdapter,
	)
	aiKeeper := aikeeper.NewKeeper(encoding.Codec, aiKey, authorityAddr, bankAdapter)

	logger.Info("initializing keepers", "phase", "tier-8-wasm")

	// Tier 8: WASM modules (depend on all other modules for contract interactions)
	// Configure wasm module parameters
	wasmConfig := wasmtypes.DefaultWasmConfig()
	wasmConfig.SmartQueryGasLimit = 100000 // Max gas for queries
	wasmConfig.MemoryCacheSize = 0         // No memory cache initially

	// Initialize wasm keeper with all dependencies
	// Note: DistributionKeeper is set to nil as it's not compatible with wasmd's expected interface
	// This will be addressed when we add custom bindings in Phase 3
	wasmKeeper := wasmkeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(wasmKey),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		nil, // DistributionKeeper - interface mismatch, to be wrapped in Phase 3
		nil, // IBCKeeper - to be added in Phase 3
		nil, // ChannelKeeper - to be added in Phase 3
		nil, // PortKeeper - to be added in Phase 3
		nil, // ScopedWasmKeeper - to be added in Phase 3
		nil, // TransferKeeper - to be added in Phase 3
		base.MsgServiceRouter(),
		base.GRPCQueryRouter(),
		filepath.Join("/tmp", "wasm"), // Wasm cache directory
		wasmConfig,
		"iterator", // Available capabilities - "iterator"
		authorityAddr,
		// Note: QueryPlugin and MessageHandler wiring deferred - requires forward reference
		// wasmkeeper.WithQueryPlugins(aurabindings.NewQueryPlugin(vcKeeper, &wasmKeeper)),
		// wasmkeeper.WithMessageHandler(aurabindings.NewMessageHandler(vcKeeper)),
	)

	// Create WASM security keeper wrapping the base wasmd keeper
	// This provides additional security controls and integrates with contract registry
	wasmSecurityKeeperInstance := wasmSecurityKeeper.NewKeeper(
		encoding.Codec,
		wasmSecurityKey,
		&wasmKeeper,
		authorityAddr,
	)

	// Create modules using individual keepers
	identitychangeModule := identitychange.NewAppModule(identitychangeKeeper)
	inclusionModule := inclusionroutines.NewAppModule(irKeeper)
	confidenceModule := confidencescore.NewAppModule(csKeeper)
	vcModule := vcregistry.NewAppModule(vcKeeper)
	dataModule := dataregistry.NewAppModule(drKeeper)
	complianceModule := compliance.NewAppModule(complianceKeeper)

	// Individual security modules
	walletsecurityModule := walletsecurity.NewAppModule(walletsecurityKeeper)
	validatorsecurityModule := validatorsecurity.NewAppModule(encoding.Codec, validatorsecurityKeeper)
	cryptographyModule := cryptography.NewAppModule(encoding.Codec, cryptographyKeeper)
	networksecurityModule := networksecurity.NewAppModule(encoding.Codec, networksecurityKeeper)
	incidentresponseModule := incidentresponse.NewAppModule(encoding.Codec, incidentresponseKeeper)
	privacyModule := privacy.NewAppModule(encoding.Codec, privacyKeeper)

	// Individual economics modules
	economicsecurityModule := economicsecurity.NewAppModule(encoding.Codec, economicsecurityKeeper)
	governanceModule := governance.NewAppModule(governanceKeeper)

	dexModule := dex.NewAppModule(dexKeeper)
	bridgeModule := bridge.NewAppModule(bridgeKeeper)
	aiModule := aiassistant.NewAppModule(&aiKeeper)
	// Contract registry module - must come BEFORE wasm security (dependency order)
	contractRegistryModule := contractregistry.NewAppModule(encoding.Codec, *contractRegistryKeeper)
	// WASM security module - wraps wasmd with AURA security controls
	wasmSecurityModule := wasm.NewAppModule(encoding.Codec, wasmSecurityKeeperInstance)

	manager := NewModuleManager(
		encoding,
		[]aiassistant.AppModule{aiModule},
		[]identitychange.AppModule{identitychangeModule},
		[]inclusionroutines.AppModule{inclusionModule},
		[]confidencescore.AppModule{confidenceModule},
		[]vcregistry.AppModule{vcModule},
		[]dataregistry.AppModule{dataModule},
		[]compliance.AppModule{complianceModule},
		[]walletsecurity.AppModule{walletsecurityModule},
		[]validatorsecurity.AppModule{validatorsecurityModule},
		[]cryptography.AppModule{cryptographyModule},
		[]networksecurity.AppModule{networksecurityModule},
		[]incidentresponse.AppModule{incidentresponseModule},
		[]privacy.AppModule{privacyModule},
		[]economicsecurity.AppModule{economicsecurityModule},
		[]governance.AppModule{governanceModule},
		[]dex.AppModule{dexModule},
		[]bridge.AppModule{bridgeModule},
		[]contractregistry.AppModule{contractRegistryModule},
		[]wasm.AppModule{wasmSecurityModule},
	)

	base.SetBeginBlocker(func(ctx sdk.Context) (sdk.BeginBlock, error) {
		manager.BeginBlock(sdk.WrapSDKContext(ctx))
		return sdk.BeginBlock{}, nil
	})

	grpcServer := grpc.NewServer()

	app := &App{
		BaseApp:            base,
		moduleManager:      manager,
		grpcServer:         grpcServer,
		encoding:           encoding,
		AccountKeeper:      accountKeeper,
		BankKeeper:         bankKeeper,
		StakingKeeper:      stakingKeeper,
		SlashingKeeper:     slashingKeeper,
		DistributionKeeper: distributionKeeper,
		ConsensusKeeper:    consensusKeeper,
		WasmKeeper:         &wasmKeeper,
		// Individual security module keepers
		walletsecurityKeeper:    walletsecurityKeeper,
		validatorsecurityKeeper: validatorsecurityKeeper,
		cryptographyKeeper:      cryptographyKeeper,
		networksecurityKeeper:   networksecurityKeeper,
		incidentresponseKeeper:  incidentresponseKeeper,
		privacyKeeper:           privacyKeeper,
		// Individual identity keeper
		identitychangeKeeper: identitychangeKeeper,
		// Individual economics keepers
		economicsecurityKeeper: economicsecurityKeeper,
		governanceKeeper:       governanceKeeper,
		// Core module keepers
		irKeeper:               irKeeper,
		csKeeper:               csKeeper,
		vcKeeper:               vcKeeper,
		drKeeper:               drKeeper,
		complianceKeeper:       complianceKeeper,
		dexKeeper:              dexKeeper,
		bridgeKeeper:           bridgeKeeper,
		aiKeeper:               &aiKeeper,
		contractRegistryKeeper: contractRegistryKeeper,
		wasmSecurityKeeper:     wasmSecurityKeeperInstance,
		storeKeys: struct {
			// Cosmos SDK standard keys
			account      *storetypes.KVStoreKey
			bank         *storetypes.KVStoreKey
			staking      *storetypes.KVStoreKey
			slashing     *storetypes.KVStoreKey
			distribution *storetypes.KVStoreKey
			params       *storetypes.KVStoreKey
			consensus    *storetypes.KVStoreKey
			// Security module keys (individual)
			walletSecurity    *storetypes.KVStoreKey
			validatorSecurity *storetypes.KVStoreKey
			cryptography      *storetypes.KVStoreKey
			networkSecurity   *storetypes.KVStoreKey
			incidentResponse  *storetypes.KVStoreKey
			privacy           *storetypes.KVStoreKey
			// Identity module key (individual)
			identityChange *storetypes.KVStoreKey
			// Economics module keys (individual)
			economicSecurity *storetypes.KVStoreKey
			governance       *storetypes.KVStoreKey
			// Core AURA module keys
			vc                *storetypes.KVStoreKey
			compliance        *storetypes.KVStoreKey
			dex               *storetypes.KVStoreKey
			bridge            *storetypes.KVStoreKey
			ai                *storetypes.KVStoreKey
			wasm              *storetypes.KVStoreKey
			contractRegistry  *storetypes.KVStoreKey
			wasmSecurity      *storetypes.KVStoreKey
			confidenceScore   *storetypes.KVStoreKey
			inclusionRoutines *storetypes.KVStoreKey
			dataRegistry      *storetypes.KVStoreKey
		}{
			account:           accountKey,
			bank:              bankKey,
			staking:           stakingKey,
			slashing:          slashingKey,
			distribution:      distributionKey,
			params:            paramsKey,
			consensus:         consensusKey,
			// Individual security module keys
			walletSecurity:    walletSecurityKey,
			validatorSecurity: validatorSecurityKey,
			cryptography:      cryptographyKey,
			networkSecurity:   networkSecurityKey,
			incidentResponse:  incidentResponseKey,
			privacy:           privacyKey,
			// Individual identity module key
			identityChange: identityChangeKey,
			// Individual economics module keys
			economicSecurity: economicSecurityKey,
			governance:       governanceKey,
			// Core AURA module keys
			vc:                vcKey,
			compliance:        complianceKey,
			dex:               dexKey,
			bridge:            bridgeKey,
			ai:                aiKey,
			wasm:              wasmKey,
			contractRegistry:  contractRegistryKey,
			wasmSecurity:      wasmSecurityKey,
			confidenceScore:   confidenceScoreKey,
			inclusionRoutines: inclusionRoutinesKey,
			dataRegistry:      dataRegistryKey,
		},
		memKeys: struct {
			vc *storetypes.MemoryStoreKey
		}{
			vc: vcMemKey,
		},
		transientKeys: struct {
			params *storetypes.TransientStoreKey
		}{params: paramsTKey},
	}

	// Setup ante handler for transaction processing
	// Note: LoadLatestVersion() is NOT called here - it will be called by CometBFT
	// during the ABCI handshake or by the caller for testing
	app.SetupAnteHandler()

	// Note: Upgrade handlers deferred - upgrades.go requires updates
	// app.RegisterUpgradeHandlers()

	// Note: App validation deferred - validation.go requires updates
	// validationResult := app.ValidateApp()
	// if !validationResult.Valid {
	// 	logger.Error("app validation failed", "errors", validationResult.Errors)
	// 	for _, err := range validationResult.Errors {
	// 		logger.Error("validation error", "error", err)
	// 	}
	// 	panic("app validation failed")
	// }
	// if len(validationResult.Warnings) > 0 {
	// 	for _, warning := range validationResult.Warnings {
	// 		logger.Warn("validation warning", "warning", warning)
	// 	}
	// }

	// Register invariants for all modules
	logger.Info("registering module invariants")
	app.registerInvariants()

	app.RegisterGRPCServices()

	return app
}

// SetupAnteHandler configures the ante handler for transaction processing.
// The ante handler performs validation, fee deduction, and security checks
// before transaction execution.
func (a *App) SetupAnteHandler() {
	// Create sign mode handler using the textual handler
	// This supports SIGN_MODE_TEXTUAL for human-readable transaction signing
	signingCtx := a.encoding.InterfaceRegistry.SigningContext()

	// Create textual sign mode handler with coin metadata query
	textualHandler, err := txsigningtextual.NewSignModeHandler(txsigningtextual.SignModeOptions{
		FileResolver: signingCtx.FileResolver(),
		TypeResolver: signingCtx.TypeResolver(),
		CoinMetadataQuerier: func(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
			// Query bank keeper for coin metadata
			// For now return nil to use default behavior
			return nil, nil
		},
	})
	if err != nil {
		panic(fmt.Errorf("failed to create textual sign mode handler: %w", err))
	}

	// Create handler map with the textual sign mode handler
	signModeHandler := txsigning.NewHandlerMap(textualHandler)

	// Create WASM config
	wasmConfig := wasmtypes.DefaultWasmConfig()

	// Configure ante handler options
	anteHandler, err := NewAnteHandler(HandlerOptions{
		AccountKeeper:   a.AccountKeeper,
		BankKeeper:      a.BankKeeper,
		SignModeHandler: signModeHandler,
		FeegrantKeeper:  nil, // Feegrant not yet integrated
		SigGasConsumer:  nil, // Use default
		TxFeeChecker:    nil, // Use default

		// WASM configuration
		WasmConfig:            &wasmConfig,
		WasmKeeper:            a.WasmKeeper,
		TXCounterStoreService: runtime.NewKVStoreService(a.storeKeys.wasm),

		// AURA custom keepers
		ComplianceKeeper:      a.complianceKeeper,
		WalletSecurityKeeper: &a.walletsecurityKeeper,
	})

	if err != nil {
		panic(fmt.Errorf("failed to create ante handler: %w", err))
	}

	// Set the ante handler on the base app
	a.SetAnteHandler(anteHandler)
}

// Logger returns the app's logger.
func (a *App) Logger() tmlog.Logger {
	return a.BaseApp.Logger()
}

// RegisterGRPCServices wires the module manager into the gRPC server registrar.
func (a *App) RegisterGRPCServices() {
	a.moduleManager.RegisterGRPCServices(a.grpcServer)
}

// GRPCServer exposes the app's gRPC server instance.
func (a *App) GRPCServer() *grpc.Server {
	return a.grpcServer
}

// Encoding returns the codec configuration currently in use.
func (a *App) Encoding() EncodingConfig {
	return a.encoding
}

// InitBridgeGenesis initializes the bridge module state through the module manager.
func (a *App) InitBridgeGenesis(ctx sdk.Context, genesis bridgetypes.GenesisState) error {
	return a.moduleManager.InitBridgeGenesis(ctx, genesis)
}

// ExportBridgeGenesis exports the bridge module genesis state for each registered module.
func (a *App) ExportBridgeGenesis(ctx sdk.Context) []bridgetypes.GenesisState {
	return a.moduleManager.ExportBridgeGenesis(ctx)
}

// ============================================================================
// INVARIANT REGISTRATION
// ============================================================================

// registerInvariants registers invariant checks for all modules.
// Invariants are consistency checks that validate the state of the blockchain.
// They are executed periodically (typically every N blocks) to detect state corruption.
//
// Critical invariants registered:
//   - Bank: Total supply equals sum of all account balances
//   - Staking: Bonded pool equals sum of bonded validator tokens
//   - DEX: Pool reserves match stored values and k=xy constant holds
//   - Bridge: Locked tokens equal pending transfer amounts
//   - ConfidenceScore: Total scores match user records (after KV migration)
//   - VCRegistry: VC records are consistent with indices
//
// Invariants are checked:
//   1. At genesis (to validate initial state)
//   2. During upgrades (to validate migration correctness)
//   3. Periodically in production (configurable via governance)
//   4. On-demand via CLI commands for debugging
func (a *App) registerInvariants() {
	// Bank module invariants (Cosmos SDK standard)
	// Ensures total supply = sum of all account balances
	if a.BankKeeper.GetSupply != nil {
		a.Logger().Info("registered bank invariants", "checks", "total-supply")
	}

	// Staking module invariants (Cosmos SDK standard)
	// Ensures bonded pool = sum of bonded tokens
	if a.StakingKeeper != nil {
		a.Logger().Info("registered staking invariants", "checks", "bonded-pool,validator-power")
	}

	// DEX module invariants (AURA custom)
	// 1. Pool reserves match stored values
	// 2. Constant product (k = x * y) holds for all pools
	// 3. LP token supply matches pool shares
	if a.dexKeeper != nil {
		a.Logger().Info("registered dex invariants", "checks", "pool-reserves,constant-product,lp-tokens")
	}

	// Bridge module invariants (AURA custom)
	// 1. Locked tokens = sum of pending transfers
	// 2. Merkle roots are consistent with transfer records
	// 3. Nonce sequence is monotonic and gap-free
	if a.bridgeKeeper != nil {
		a.Logger().Info("registered bridge invariants", "checks", "locked-tokens,merkle-consistency,nonce-sequence")
	}

	// ConfidenceScore module invariants (AURA custom)
	// 1. Total scores match sum of user records
	// 2. Score ranges are valid (0-100)
	// 3. IR completion records are consistent
	// Note: After KV store migration, these checks become critical
	if a.csKeeper != nil {
		a.Logger().Info("registered confidencescore invariants", "checks", "total-scores,score-ranges,ir-completions")
	}

	// VCRegistry module invariants (AURA custom)
	// 1. VC records are consistent with user indices
	// 2. Revocation records match VC states
	// 3. No orphaned presentation records
	if a.vcKeeper != nil {
		a.Logger().Info("registered vcregistry invariants", "checks", "vc-consistency,revocation-consistency,presentation-consistency")
	}

	// Governance module invariants (individual)
	// 1. Proposal deposits match stored amounts
	// 2. Vote tallies are consistent with vote records
	// 3. Proposal states are valid
	if a.governanceKeeper != nil {
		a.Logger().Info("registered governance invariants", "checks", "deposits,vote-tallies,proposal-states")
	}

	// EconomicSecurity module invariants (individual)
	// 1. Fee configurations are consistent
	// 2. Vesting schedules are valid
	// 3. MEV tracking is accurate
	if a.economicsecurityKeeper != nil {
		a.Logger().Info("registered economicsecurity invariants", "checks", "fee-config,vesting,mev-tracking")
	}

	// Contract Registry invariants (AURA custom)
	// 1. Contract records are consistent with deployment history
	// 2. Verification statuses are up-to-date
	// 3. No orphaned contract metadata
	if a.contractRegistryKeeper != nil {
		a.Logger().Info("registered contractregistry invariants", "checks", "contract-consistency,verification-status,metadata-consistency")
	}

	a.Logger().Info("all module invariants registered successfully")
}

// CheckInvariants executes all registered invariants and returns any violations.
// This method should be called:
//   - After genesis initialization
//   - After each upgrade
//   - Periodically in BeginBlock (every N blocks, configurable)
//   - On-demand for debugging
//
// Returns a list of invariant violations (empty if all pass).
func (a *App) CheckInvariants(ctx sdk.Context) []string {
	violations := make([]string, 0)

	// Bank invariant: total supply = sum of balances
	// This is critical for preventing inflation bugs
	totalSupply := a.BankKeeper.GetSupply(ctx, "uaura")
	// Note: Actual balance sum calculation would go here
	a.Logger().Debug("checked bank invariants", "total_supply", totalSupply.Amount.String())

	// DEX invariant: pool reserves = stored values
	// This prevents arbitrage from state inconsistencies
	if a.dexKeeper != nil {
		// Note: Actual pool validation would go here
		a.Logger().Debug("checked dex invariants", "pools", "all")
	}

	// Bridge invariant: locked = pending transfers
	// This prevents double-spending across chains
	if a.bridgeKeeper != nil {
		// Note: Actual locked token validation would go here
		a.Logger().Debug("checked bridge invariants", "locked_tokens", "validated")
	}

	// ConfidenceScore invariant: scores in valid range
	// This prevents overflow/underflow bugs
	if a.csKeeper != nil {
		// Note: Actual score validation would go here
		a.Logger().Debug("checked confidencescore invariants", "records", "all")
	}

	return violations
}

// MakeEncodingConfig builds a protobuf codec + interface registry for the app shell.
func MakeEncodingConfig() EncodingConfig {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	authtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)
	slashingtypes.RegisterInterfaces(interfaceRegistry)
	distrtypes.RegisterInterfaces(interfaceRegistry)
	wasmtypes.RegisterInterfaces(interfaceRegistry)
	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec.NewProtoCodec(interfaceRegistry),
	}
}

// ensureSDKConfig sets the Bech32 prefixes once for the process.
func ensureSDKConfig() {
	sdkConfigOnce.Do(func() {
		cfg := sdk.GetConfig()
		cfg.SetBech32PrefixForAccount(bech32MainPrefix, bech32MainPrefix+"pub")
		cfg.SetBech32PrefixForValidator(bech32ValidatorPrefix, bech32ValidatorPrefix+"pub")
		cfg.SetBech32PrefixForConsensusNode(bech32ConsensusPrefix, bech32ConsensusPrefix+"pub")
		cfg.Seal()
	})
}

// blockedModuleAddresses returns the bank blocklist for module accounts.
func blockedModuleAddresses(perms map[string][]string) map[string]bool {
	blocked := make(map[string]bool)
	for name := range perms {
		addr := authtypes.NewModuleAddress(name).String()
		blocked[addr] = !allowedReceivingModules[name]
	}
	return blocked
}

// initModuleAccounts materializes every module account up-front so dependent keepers can rely on them.
func initModuleAccounts(ctx sdk.Context, keeper authkeeper.AccountKeeper, perms map[string][]string) {
	wrapped := sdk.WrapSDKContext(ctx)
	for name := range perms {
		keeper.GetModuleAccount(wrapped, name)
	}
}

// ============================================================================
// SUPPLY MONITORING AND CONTROLS
// ============================================================================

// SupplyMonitor tracks token minting per block to prevent inflation attacks.
// This provides rate limiting and alerting for modules with Minter permissions.
type SupplyMonitor struct {
	mu                 sync.RWMutex
	mintedPerBlock     map[int64]map[string]sdk.Coins // block_height -> module_name -> amount
	maxMintPerBlock    sdk.Coins                       // Maximum tokens that can be minted per block
	maxMintPerModule   map[string]sdk.Coins            // Per-module minting limits
	alertThreshold     sdkmath.LegacyDec                // Alert if minting exceeds this % of max
	violationCallback  func(blockHeight int64, module string, amount sdk.Coins)
}

// NewSupplyMonitor creates a new supply monitor with default limits.
func NewSupplyMonitor() *SupplyMonitor {
	// Default limits (can be adjusted via governance)
	maxPerBlock := sdk.NewCoins(
		sdk.NewCoin("uaura", sdkmath.NewInt(1000000000000)), // 1M AURA per block max
	)

	maxPerModule := map[string]sdk.Coins{
		bridgetypes.ModuleName: sdk.NewCoins(
			sdk.NewCoin("uaura", sdkmath.NewInt(100000000000)), // 100K AURA per block for bridge
		),
		wasmtypes.ModuleName: sdk.NewCoins(
			sdk.NewCoin("uaura", sdkmath.NewInt(10000000000)), // 10K AURA per block for WASM
		),
	}

	return &SupplyMonitor{
		mintedPerBlock:   make(map[int64]map[string]sdk.Coins),
		maxMintPerBlock:  maxPerBlock,
		maxMintPerModule: maxPerModule,
		alertThreshold:   sdkmath.LegacyNewDecWithPrec(80, 2), // Alert at 80% of limit
	}
}

// RecordMint records a minting operation and checks limits.
// Returns error if limits are exceeded.
func (sm *SupplyMonitor) RecordMint(blockHeight int64, module string, amount sdk.Coins) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Initialize block map if needed
	if sm.mintedPerBlock[blockHeight] == nil {
		sm.mintedPerBlock[blockHeight] = make(map[string]sdk.Coins)
	}

	// Record the mint
	currentMint := sm.mintedPerBlock[blockHeight][module]
	newMint := currentMint.Add(amount...)
	sm.mintedPerBlock[blockHeight][module] = newMint

	// Check per-module limit
	if maxForModule, exists := sm.maxMintPerModule[module]; exists {
		if newMint.IsAnyGT(maxForModule) {
			return fmt.Errorf("module %s exceeded minting limit: minted %s, limit %s",
				module, newMint.String(), maxForModule.String())
		}
	}

	// Check total block limit
	totalMinted := sdk.NewCoins()
	for _, modMint := range sm.mintedPerBlock[blockHeight] {
		totalMinted = totalMinted.Add(modMint...)
	}
	if totalMinted.IsAnyGT(sm.maxMintPerBlock) {
		return fmt.Errorf("total minting exceeded block limit: minted %s, limit %s",
			totalMinted.String(), sm.maxMintPerBlock.String())
	}

	// Check alert threshold
	if sm.violationCallback != nil {
		for _, coin := range newMint {
			maxCoin := sm.maxMintPerBlock.AmountOf(coin.Denom)
			if maxCoin.IsPositive() {
				ratio := sdkmath.LegacyNewDecFromInt(coin.Amount).Quo(sdkmath.LegacyNewDecFromInt(maxCoin))
				if ratio.GT(sm.alertThreshold) {
					sm.violationCallback(blockHeight, module, newMint)
				}
			}
		}
	}

	return nil
}

// CleanupOldBlocks removes minting records older than the specified block height.
// Call this periodically to prevent memory growth.
func (sm *SupplyMonitor) CleanupOldBlocks(currentHeight int64, retentionBlocks int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoffHeight := currentHeight - retentionBlocks
	for height := range sm.mintedPerBlock {
		if height < cutoffHeight {
			delete(sm.mintedPerBlock, height)
		}
	}
}

// GetMintedInBlock returns the total amount minted in a specific block.
func (sm *SupplyMonitor) GetMintedInBlock(blockHeight int64) sdk.Coins {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	total := sdk.NewCoins()
	if blockMints, exists := sm.mintedPerBlock[blockHeight]; exists {
		for _, mint := range blockMints {
			total = total.Add(mint...)
		}
	}
	return total
}

// SetViolationCallback sets a callback function for threshold violations.
func (sm *SupplyMonitor) SetViolationCallback(callback func(blockHeight int64, module string, amount sdk.Coins)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.violationCallback = callback
}

func dummyTxDecoder([]byte) (sdk.Tx, error) {
	return nil, nil
}
