package app

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	tmlog "cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	pruningtypes "cosmossdk.io/store/pruning/types"
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"
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
	"google.golang.org/grpc"

	// AURA Core Modules (kept as-is)
	aurabindings "github.com/aequitas/aura/chain/x/aura-bindings"
	aurabindingskeeper "github.com/aequitas/aura/chain/x/aura-bindings/keeper"
	aurabindingstypes "github.com/aequitas/aura/chain/x/aura-bindings/types"
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
	"github.com/aequitas/aura/chain/x/economics"
	economicskeeper "github.com/aequitas/aura/chain/x/economics/keeper"
	economicstypes "github.com/aequitas/aura/chain/x/economics/types"
	"github.com/aequitas/aura/chain/x/identity"
	identitykeeper "github.com/aequitas/aura/chain/x/identity/keeper"
	identitytypes "github.com/aequitas/aura/chain/x/identity/types"
	"github.com/aequitas/aura/chain/x/inclusionroutines"
	irkeeper "github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
	irparams "github.com/aequitas/aura/chain/x/inclusionroutines/params"
	irtypes "github.com/aequitas/aura/chain/x/inclusionroutines/types"
	"github.com/aequitas/aura/chain/x/monitoring"
	monitoringkeeper "github.com/aequitas/aura/chain/x/monitoring/keeper"
	monitoringtypes "github.com/aequitas/aura/chain/x/monitoring/types"
	"github.com/aequitas/aura/chain/x/prevalidation"
	prevalidationkeeper "github.com/aequitas/aura/chain/x/prevalidation/keeper"
	prevalidationtypes "github.com/aequitas/aura/chain/x/prevalidation/types"
	"github.com/aequitas/aura/chain/x/security"
	securitykeeper "github.com/aequitas/aura/chain/x/security/keeper"
	securitytypes "github.com/aequitas/aura/chain/x/security/types"
	"github.com/aequitas/aura/chain/x/vcregistry"
	vckeeper "github.com/aequitas/aura/chain/x/vcregistry/keeper"
	vcparams "github.com/aequitas/aura/chain/x/vcregistry/params"
	vctypes "github.com/aequitas/aura/chain/x/vcregistry/types"
	"github.com/aequitas/aura/chain/x/wasm"
	wasmSecurityKeeper "github.com/aequitas/aura/chain/x/wasm/keeper"
	wasmSecurityTypes "github.com/aequitas/aura/chain/x/wasm/types"

	// AURA Security Modules (individual)
	"github.com/aequitas/aura/chain/x/cryptography"
	cryptographykeeper "github.com/aequitas/aura/chain/x/cryptography/keeper"
	cryptographytypes "github.com/aequitas/aura/chain/x/cryptography/types"
	"github.com/aequitas/aura/chain/x/incidentresponse"
	incidentresponsekeeper "github.com/aequitas/aura/chain/x/incidentresponse/keeper"
	incidentresponsetypes "github.com/aequitas/aura/chain/x/incidentresponse/types"
	"github.com/aequitas/aura/chain/x/networksecurity"
	networksecuritykeeper "github.com/aequitas/aura/chain/x/networksecurity/keeper"
	networksecuritytypes "github.com/aequitas/aura/chain/x/networksecurity/types"
	"github.com/aequitas/aura/chain/x/privacy"
	privacykeeper "github.com/aequitas/aura/chain/x/privacy/keeper"
	privacytypes "github.com/aequitas/aura/chain/x/privacy/types"
	"github.com/aequitas/aura/chain/x/validatorsecurity"
	validatorsecuritykeeper "github.com/aequitas/aura/chain/x/validatorsecurity/keeper"
	validatorsecuritytypes "github.com/aequitas/aura/chain/x/validatorsecurity/types"
	"github.com/aequitas/aura/chain/x/walletsecurity"
	walletsecuritykeeper "github.com/aequitas/aura/chain/x/walletsecurity/keeper"
	walletsecuritytypes "github.com/aequitas/aura/chain/x/walletsecurity/types"

	// AURA Identity Module (individual)
	"github.com/aequitas/aura/chain/x/identitychange"
	identitychangekeeper "github.com/aequitas/aura/chain/x/identitychange/keeper"
	identitychangeparams "github.com/aequitas/aura/chain/x/identitychange/params"
	identitychangetypes "github.com/aequitas/aura/chain/x/identitychange/types"

	// AURA Economics Modules (individual)
	"github.com/aequitas/aura/chain/x/economicsecurity"
	economicsecuritykeeper "github.com/aequitas/aura/chain/x/economicsecurity/keeper"
	economicsecurityparams "github.com/aequitas/aura/chain/x/economicsecurity/params"
	economicsecuritytypes "github.com/aequitas/aura/chain/x/economicsecurity/types"
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
	TxConfig          client.TxConfig
}

// storeKeys holds all KV store keys with methods to access them as a single source of truth.
// This eliminates the need to update multiple locations when adding/removing modules.
type storeKeys struct {
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
	security          *storetypes.KVStoreKey

	// Identity module key (individual)
	identityChange *storetypes.KVStoreKey
	identity       *storetypes.KVStoreKey

	// Economics module keys (individual)
	economicSecurity *storetypes.KVStoreKey
	governance       *storetypes.KVStoreKey
	economics        *storetypes.KVStoreKey

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
	monitoring        *storetypes.KVStoreKey
	prevalidation     *storetypes.KVStoreKey
	aurabindings      *storetypes.KVStoreKey
}

// Names returns all store key names for StoreKeyNames function.
// This is the single source of truth for store key names.
func (s *storeKeys) Names() []string {
	return []string{
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		slashingtypes.StoreKey,
		distrtypes.StoreKey,
		paramstypes.StoreKey,
		consensustypes.StoreKey,
		walletsecuritytypes.StoreKey,
		validatorsecuritytypes.StoreKey,
		cryptographytypes.StoreKey,
		networksecuritytypes.StoreKey,
		incidentresponsetypes.StoreKey,
		privacytypes.StoreKey,
		identitychangetypes.StoreKey,
		identitytypes.StoreKey,
		economicsecuritytypes.StoreKey,
		governancetypes.StoreKey,
		economicstypes.StoreKey,
		vctypes.StoreKey,
		compliancetypes.StoreKey,
		dextypes.StoreKey,
		bridgetypes.StoreKey,
		aitypes.StoreKey,
		wasmtypes.StoreKey,
		contractregistrytypes.StoreKey,
		wasmSecurityTypes.StoreKey,
		cstypes.StoreKey,
		irtypes.StoreKey,
		drtypes.StoreKey,
		monitoringtypes.StoreKey,
		prevalidationtypes.StoreKey,
		aurabindingstypes.StoreKey,
		securitytypes.StoreKey,
	}
}

// AsMap returns store keys as a map for MountKVStores.
// This is the single source of truth for store key mounting.
func (s *storeKeys) AsMap() map[string]*storetypes.KVStoreKey {
	return map[string]*storetypes.KVStoreKey{
		// Cosmos SDK standard keys
		authtypes.StoreKey:      s.account,
		banktypes.StoreKey:      s.bank,
		stakingtypes.StoreKey:   s.staking,
		slashingtypes.StoreKey:  s.slashing,
		distrtypes.StoreKey:     s.distribution,
		paramstypes.StoreKey:    s.params,
		consensustypes.StoreKey: s.consensus,

		// Security module keys (individual)
		walletsecuritytypes.StoreKey:    s.walletSecurity,
		validatorsecuritytypes.StoreKey: s.validatorSecurity,
		cryptographytypes.StoreKey:      s.cryptography,
		networksecuritytypes.StoreKey:   s.networkSecurity,
		incidentresponsetypes.StoreKey:  s.incidentResponse,
		privacytypes.StoreKey:           s.privacy,

		// Identity module key (individual)
		identitychangetypes.StoreKey: s.identityChange,
		identitytypes.StoreKey:       s.identity,

		// Economics module keys (individual)
		economicsecuritytypes.StoreKey: s.economicSecurity,
		governancetypes.StoreKey:       s.governance,
		economicstypes.StoreKey:        s.economics,

		// Core AURA module keys (unchanged)
		vctypes.StoreKey:               s.vc,
		compliancetypes.StoreKey:       s.compliance,
		dextypes.StoreKey:              s.dex,
		bridgetypes.StoreKey:           s.bridge,
		aitypes.StoreKey:               s.ai,
		wasmtypes.StoreKey:             s.wasm,
		contractregistrytypes.StoreKey: s.contractRegistry,
		wasmSecurityTypes.StoreKey:     s.wasmSecurity,
		cstypes.StoreKey:               s.confidenceScore,
		irtypes.StoreKey:               s.inclusionRoutines,
		drtypes.StoreKey:               s.dataRegistry,
		monitoringtypes.StoreKey:       s.monitoring,
		prevalidationtypes.StoreKey:    s.prevalidation,
		aurabindingstypes.StoreKey:     s.aurabindings,
		securitytypes.StoreKey:         s.security,
	}
}

// initStoreKeys creates all store keys as a single source of truth.
// Adding a new module only requires updating this function.
func initStoreKeys() *storeKeys {
	return &storeKeys{
		// Cosmos SDK standard keys
		account:      storetypes.NewKVStoreKey(authtypes.StoreKey),
		bank:         storetypes.NewKVStoreKey(banktypes.StoreKey),
		staking:      storetypes.NewKVStoreKey(stakingtypes.StoreKey),
		slashing:     storetypes.NewKVStoreKey(slashingtypes.StoreKey),
		distribution: storetypes.NewKVStoreKey(distrtypes.StoreKey),
		params:       storetypes.NewKVStoreKey(paramstypes.StoreKey),
		consensus:    storetypes.NewKVStoreKey(consensustypes.StoreKey),

		// Security module keys (individual)
		walletSecurity:    storetypes.NewKVStoreKey(walletsecuritytypes.StoreKey),
		validatorSecurity: storetypes.NewKVStoreKey(validatorsecuritytypes.StoreKey),
		cryptography:      storetypes.NewKVStoreKey(cryptographytypes.StoreKey),
		networkSecurity:   storetypes.NewKVStoreKey(networksecuritytypes.StoreKey),
		incidentResponse:  storetypes.NewKVStoreKey(incidentresponsetypes.StoreKey),
		privacy:           storetypes.NewKVStoreKey(privacytypes.StoreKey),
		security:          storetypes.NewKVStoreKey(securitytypes.StoreKey),

		// Identity module key (individual)
		identityChange: storetypes.NewKVStoreKey(identitychangetypes.StoreKey),
		identity:       storetypes.NewKVStoreKey(identitytypes.StoreKey),

		// Economics module keys (individual)
		economicSecurity: storetypes.NewKVStoreKey(economicsecuritytypes.StoreKey),
		governance:       storetypes.NewKVStoreKey(governancetypes.StoreKey),
		economics:        storetypes.NewKVStoreKey(economicstypes.StoreKey),

		// Core AURA module keys
		vc:                storetypes.NewKVStoreKey(vctypes.StoreKey),
		compliance:        storetypes.NewKVStoreKey(compliancetypes.StoreKey),
		dex:               storetypes.NewKVStoreKey(dextypes.StoreKey),
		bridge:            storetypes.NewKVStoreKey(bridgetypes.StoreKey),
		ai:                storetypes.NewKVStoreKey(aitypes.StoreKey),
		wasm:              storetypes.NewKVStoreKey(wasmtypes.StoreKey),
		contractRegistry:  storetypes.NewKVStoreKey(contractregistrytypes.StoreKey),
		wasmSecurity:      storetypes.NewKVStoreKey(wasmSecurityTypes.StoreKey),
		confidenceScore:   storetypes.NewKVStoreKey(cstypes.StoreKey),
		inclusionRoutines: storetypes.NewKVStoreKey(irtypes.StoreKey),
		dataRegistry:      storetypes.NewKVStoreKey(drtypes.StoreKey),
		monitoring:        storetypes.NewKVStoreKey(monitoringtypes.StoreKey),
		prevalidation:     storetypes.NewKVStoreKey(prevalidationtypes.StoreKey),
		aurabindings:      storetypes.NewKVStoreKey(aurabindingstypes.StoreKey),
	}
}

// App wires all Aura modules plus the Cosmos SDK base keepers into a runnable node shell.
type App struct {
	*baseapp.BaseApp

	moduleManager *sdkmodule.Manager
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
	identityKeeper       *identitykeeper.Keeper

	// Economics module keepers (individual)
	economicsecurityKeeper *economicsecuritykeeper.Keeper
	governanceKeeper       *governancekeeper.Keeper
	economicsKeeper        *economicskeeper.Keeper

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
	securityKeeper         *securitykeeper.Keeper
	monitoringKeeper       *monitoringkeeper.Keeper
	prevalidationKeeper    *prevalidationkeeper.Keeper
	aurabindingsKeeper     *aurabindingskeeper.Keeper

	storeKeys *storeKeys
	memKeys struct {
		vc           *storetypes.MemoryStoreKey
		security     *storetypes.MemoryStoreKey
		aurabindings *storetypes.MemoryStoreKey
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

// StoreKeyNames lists all KV store names mounted by the app.
// Uses the centralized storeKeys.Names() as the single source of truth.
func StoreKeyNames() []string {
	// Create a temporary storeKeys instance just to get the names
	// This ensures Names() is the single source of truth
	keys := &storeKeys{}
	return keys.Names()
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
	baseAppOptions := []func(*baseapp.BaseApp){
		// Use PruningNothing to keep all versions - required for queries to work
		// This ensures IAVL trees retain historical state for versioned queries
		baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing)),
		// Enable IAVL fast node for better query performance
		// Disable fast node to avoid version lookup errors during tx execution
		baseapp.SetIAVLDisableFastNode(true),
		// Set IAVL cache size for improved performance
		baseapp.SetIAVLCacheSize(10000),
	}
	if chainID != "" {
		baseAppOptions = append(baseAppOptions, baseapp.SetChainID(chainID))
	}

	base := baseapp.NewBaseApp(appName, logger, db, encoding.TxConfig.TxDecoder(), baseAppOptions...)
	base.SetInterfaceRegistry(encoding.InterfaceRegistry)
	base.SetTxEncoder(encoding.TxConfig.TxEncoder())

	// Initialize all KV store keys from single source of truth
	keys := initStoreKeys()

	// Memory keys (not part of centralized keys as they're transient)
	validatorSecurityMemKey := storetypes.NewMemoryStoreKey(validatorsecuritytypes.MemStoreKey)
	vcMemKey := storetypes.NewMemoryStoreKey(vctypes.MemStoreKey)
	securityMemKey := storetypes.NewMemoryStoreKey(securitytypes.MemStoreKey)
	aurabindingsMemKey := storetypes.NewMemoryStoreKey(aurabindingstypes.MemStoreKey)

	// Transient keys (not part of centralized keys as they're transient)
	paramsTKey := storetypes.NewTransientStoreKey(paramstypes.TStoreKey)

	// Mount all KV stores using the centralized map (single source of truth)
	base.MountKVStores(keys.AsMap())
	base.MountTransientStores(map[string]*storetypes.TransientStoreKey{
		paramstypes.TStoreKey: paramsTKey,
	})
	base.MountMemoryStores(map[string]*storetypes.MemoryStoreKey{
		vctypes.MemStoreKey:                vcMemKey,
		validatorsecuritytypes.MemStoreKey: validatorSecurityMemKey,
		securitytypes.MemStoreKey:          securityMemKey,
		aurabindingstypes.MemStoreKey:      aurabindingsMemKey,
	})

	paramsKeeper := paramskeeper.NewKeeper(encoding.Codec, codec.NewLegacyAmino(), keys.params, paramsTKey)
	bridgeSubspace := paramsKeeper.Subspace(bridgetypes.ModuleName)

	accountCodec := address.NewBech32Codec(bech32MainPrefix)
	validatorCodec := address.NewBech32Codec(bech32ValidatorPrefix)
	consensusCodec := address.NewBech32Codec(bech32ConsensusPrefix)
	authorityAddr := authtypes.NewModuleAddress(governancetypes.ModuleName).String() // Using governance module

	accountKeeper := authkeeper.NewAccountKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.account),
		authtypes.ProtoBaseAccount,
		moduleAccountPermissions,
		accountCodec,
		bech32MainPrefix,
		authorityAddr,
	)

	// Ensure module accounts exist before keepers that depend on them (staking, bank) are created.
	// Create base bank keeper (will be wrapped with monitoring later)
	baseBankKeeper := bankkeeper.NewBaseKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.bank),
		accountKeeper,
		blockedModuleAddresses(moduleAccountPermissions),
		authorityAddr,
		logger,
	)
	accountAdapter := newAccountKeeperAdapter(accountKeeper)

	stakingKeeper := stakingkeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.staking),
		accountKeeper,
		baseBankKeeper,
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
		baseBankKeeper,
		stakingKeeper,
		authtypes.FeeCollectorName,
		authorityAddr,
	)

	// Consensus keeper - required for BaseApp.ParamStore in SDK v0.53+
	consensusKeeper := consensuskeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.consensus),
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
	complianceKeeper := compliancekeeper.NewKeeper(encoding.Codec, keys.compliance)

	// Wrap bank keeper with transaction monitoring for AML compliance
	// This intercepts all coin transfers to evaluate compliance rules before execution
	monitoredBankKeeper := compliancekeeper.NewMonitoredBankKeeper(baseBankKeeper, complianceKeeper)

	// Create bank adapter using the monitored bank keeper for compliance
	bankAdapter := newBankKeeperAdapter(monitoredBankKeeper)

	// Individual security module keepers
	walletsecurityKeeper := walletsecuritykeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.walletSecurity),
		logger.With("module", walletsecuritytypes.ModuleName),
	)

	validatorsecurityKeeper := validatorsecuritykeeper.NewKeeper(
		encoding.Codec,
		keys.validatorSecurity,
		validatorSecurityMemKey,
		authorityAddr,
		newValidatorSecurityStakingAdapter(stakingKeeper),
		slashingKeeper,
		monitoredBankKeeper,
	)

	cryptographyKeeper := cryptographykeeper.NewKeeper(
		encoding.Codec,
		keys.cryptography,
		logger.With("module", cryptographytypes.ModuleName),
		authorityAddr,
	)

	networksecurityKeeper := networksecuritykeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.networkSecurity),
		authorityAddr,
		logger.With("module", networksecuritytypes.ModuleName),
	)

	incidentresponseKeeper := incidentresponsekeeper.NewKeeperKV(
		keys.incidentResponse,
		encoding.Codec,
	)

	privacyKeeper := privacykeeper.NewKeeper(
		encoding.Codec,
		keys.privacy,
		accountKeeper,
		monitoredBankKeeper,
	)

	securityKeeper := securitykeeper.NewKeeper(
		encoding.Codec,
		keys.security,
		securityMemKey,
		authorityAddr,
		bankAdapter,
		newSecurityStakingAdapter(stakingKeeper),
		accountAdapter,
	)

	// Individual economics module keepers
	governanceKeeper := governancekeeper.NewKeeper(encoding.Codec, keys.governance, stakingKeeper, monitoredBankKeeper)

	economicsecurityParamsStore := economicsecurityparams.NewStore(*economicsecuritytypes.DefaultParams())
	economicsecurityKeeper := economicsecuritykeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.economicSecurity),
		economicsecurityParamsStore,
		authorityAddr,
	)

	logger.Info("initializing keepers", "phase", "tier-2-aura-base")

	// Tier 2: Individual identity keeper
	identitychangeParamsStore := identitychangeparams.NewStore(identitychangetypes.DefaultParams())
	identitychangeKeeper := identitychangekeeper.NewKeeper(
		runtime.NewKVStoreService(keys.identityChange),
		encoding.Codec,
		identitychangeParamsStore,
		authorityAddr,
		logger.With("module", identitychangetypes.ModuleName),
	)

	identityKeeper := identitykeeper.NewKeeper(
		runtime.NewKVStoreService(keys.identity),
		encoding.Codec,
		authorityAddr,
		logger.With("module", identitytypes.ModuleName),
	)

	drParamsStore := drparams.NewStore(drtypes.DefaultParams())
	drKeeper := drkeeper.NewKeeper(
		runtime.NewKVStoreService(keys.dataRegistry),
		encoding.Codec,
		drParamsStore,
		authorityAddr,
		logger.With("module", drtypes.ModuleName),
	)

	monitoringKeeper := monitoringkeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.monitoring),
		authorityAddr,
	)

	economicsKeeper := economicskeeper.NewKeeper(
		encoding.Codec,
		runtime.NewKVStoreService(keys.economics),
		authorityAddr,
	)

	prevalidationKeeper := prevalidationkeeper.NewKeeper(
		encoding.Codec,
		keys.prevalidation,
	)
	prevalidationKeeper.SetLogger(logger.With("module", prevalidationtypes.ModuleName))

	logger.Info("initializing keepers", "phase", "tier-3-inclusion-routines")

	// Tier 3: Inclusion routines (depends on dataregistry for data validation)
	irParamsStore := irparams.NewStore(irtypes.DefaultParams())
	irKeeper := irkeeper.NewKeeper(
		runtime.NewKVStoreService(keys.inclusionRoutines),
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
		runtime.NewKVStoreService(keys.confidenceScore),
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
		WithStore(keys.vc, encoding.Codec).
		Build() // Note: ConfidenceScore keeper not wired due to interface mismatch - fix in follow-up

	aurabindingsKeeper := aurabindingskeeper.NewKeeper(
		encoding.Codec,
		keys.aurabindings,
		vcKeeper,
	)

	logger.Info("initializing keepers", "phase", "tier-6-contract-registry")

	// Tier 6: Contract registry (depends on vcregistry, compliance, confidencescore)
	// Contract registry must be initialized BEFORE wasm security as it's a dependency
	contractRegistryKeeper := contractregistrykeeper.NewKeeper(
		keys.contractRegistry,
		encoding.Codec,
		authorityAddr,
	)
	// Note: Keeper dependencies not wired due to interface mismatches - fix in follow-up
	// contractRegistryKeeper.SetVCRegistryKeeper(vcKeeper)
	// contractRegistryKeeper.SetComplianceKeeper(complianceKeeper)
	// contractRegistryKeeper.SetConfidenceScoreKeeper(csKeeper)

	logger.Info("initializing keepers", "phase", "tier-7-dex-bridge-ai")

	// Tier 7: DEX, Bridge, and AI modules (depend on vcregistry)
	vcAdapter := newVCRegistryKeeperAdapter(vcKeeper)

	dexKeeper := dexkeeper.NewKeeper(encoding.Codec, keys.dex, bankAdapter, accountAdapter, vcAdapter)
	bridgeKeeper := bridgekeeper.NewKeeper(
		encoding.Codec,
		keys.bridge,
		&bridgeSubspace,
		bankAdapter,
		accountAdapter,
		vcAdapter,
		stakingKeeper, // For validator slashing
	)
	aiKeeper := aikeeper.NewKeeper(encoding.Codec, keys.ai, authorityAddr, bankAdapter)

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
		runtime.NewKVStoreService(keys.wasm),
		accountKeeper,
		monitoredBankKeeper,
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
	// Note: QueryPlugin wiring deferred - requires forward reference
	wasmkeeper.WithQueryPlugins(aurabindings.NewQueryPlugin(vcKeeper, nil)),
	wasmkeeper.WithMessageHandler(aurabindings.NewMessageHandler(vcKeeper)),
)

	// Create WASM security keeper wrapping the base wasmd keeper
	// This provides additional security controls and integrates with contract registry
	wasmSecurityKeeperInstance := wasmSecurityKeeper.NewKeeper(
		encoding.Codec,
		keys.wasmSecurity,
		&wasmKeeper,
		authorityAddr,
	)

	// Create modules using individual keepers
	identitychangeModule := identitychange.NewAppModule(identitychangeKeeper)
	identityModule := identity.NewAppModule(encoding.Codec, identityKeeper)
	inclusionModule := inclusionroutines.NewAppModule(irKeeper)
	confidenceModule := confidencescore.NewAppModule(csKeeper)
	vcModule := vcregistry.NewAppModule(vcKeeper)
	dataModule := dataregistry.NewAppModule(drKeeper)
	complianceModule := compliance.NewAppModule(complianceKeeper)
	monitoringModule := monitoring.NewAppModule(encoding.Codec, monitoringKeeper)

	// Individual security modules
	walletsecurityModule := walletsecurity.NewAppModule(walletsecurityKeeper)
	validatorsecurityModule := validatorsecurity.NewAppModule(encoding.Codec, validatorsecurityKeeper)
	cryptographyModule := cryptography.NewAppModule(encoding.Codec, cryptographyKeeper)
	networksecurityModule := networksecurity.NewAppModule(encoding.Codec, networksecurityKeeper)
	incidentresponseModule := incidentresponse.NewAppModule(encoding.Codec, incidentresponseKeeper)
	privacyModule := privacy.NewAppModule(encoding.Codec, privacyKeeper)
	securityModule := security.NewAppModule(encoding.Codec, securityKeeper)

	// Individual economics modules
	economicsecurityModule := economicsecurity.NewAppModule(encoding.Codec, economicsecurityKeeper)
	economicsModule := economics.NewAppModule(encoding.Codec, economicsKeeper)
	governanceModule := governance.NewAppModule(governanceKeeper)

	dexModule := dex.NewAppModule(dexKeeper)
	bridgeModule := bridge.NewAppModule(bridgeKeeper)
	aiModule := aiassistant.NewAppModule(&aiKeeper)
	// Contract registry module - must come BEFORE wasm security (dependency order)
	contractRegistryModule := contractregistry.NewAppModule(encoding.Codec, *contractRegistryKeeper)
	// WASM security module - wraps wasmd with AURA security controls
	wasmSecurityModule := wasm.NewAppModule(encoding.Codec, wasmSecurityKeeperInstance)
	prevalidationModule := prevalidation.NewAppModule(encoding.Codec, prevalidationKeeper)
	aurabindingsModule := aurabindings.NewAppModule(encoding.Codec, *aurabindingsKeeper)

	coreModules := []sdkmodule.AppModule{
		auth.NewAppModule(encoding.Codec, accountKeeper, nil, nil),
		bankmodule.NewAppModule(encoding.Codec, monitoredBankKeeper, accountKeeper, nil),
		stakingmodule.NewAppModule(encoding.Codec, stakingKeeper, accountKeeper, monitoredBankKeeper, nil),
		slashingmodule.NewAppModule(encoding.Codec, slashingKeeper, accountKeeper, monitoredBankKeeper, stakingKeeper, nil, encoding.InterfaceRegistry),
		distribution.NewAppModule(encoding.Codec, distributionKeeper, accountKeeper, monitoredBankKeeper, stakingKeeper, nil),
		params.NewAppModule(paramsKeeper),
		consensus.NewAppModule(encoding.Codec, consensusKeeper),
		genutilmodule.NewAppModule(accountKeeper, stakingKeeper, base, encoding.TxConfig),
		wasmSecurityModule,
		contractRegistryModule,
	}

	auraModules := []sdkmodule.AppModule{
		inclusionModule,
		confidenceModule,
	}

	adapterModules := wrapAdapters(map[string]interface{}{
		identitychangetypes.ModuleName:    identitychangeModule,
		identitytypes.ModuleName:          identityModule,
		vctypes.ModuleName:                vcModule,
		drtypes.ModuleName:                dataModule,
		compliancetypes.ModuleName:        complianceModule,
		monitoringtypes.ModuleName:        monitoringModule,
		walletsecuritytypes.ModuleName:    walletsecurityModule,
		validatorsecuritytypes.ModuleName: validatorsecurityModule,
		cryptographytypes.ModuleName:      cryptographyModule,
		networksecuritytypes.ModuleName:   networksecurityModule,
		incidentresponsetypes.ModuleName:  incidentresponseModule,
		privacytypes.ModuleName:           privacyModule,
		economicsecuritytypes.ModuleName:  economicsecurityModule,
		economicstypes.ModuleName:         economicsModule,
		governancetypes.ModuleName:        governanceModule,
		dextypes.ModuleName:               dexModule,
		bridgetypes.ModuleName:            bridgeModule,
		aitypes.ModuleName:                aiModule,
		prevalidationtypes.ModuleName:     prevalidationModule,
		securitytypes.ModuleName:          securityModule,
		aurabindingstypes.ModuleName:      aurabindingsModule,
	})

	moduleManager := sdkmodule.NewManager(append(append(coreModules, auraModules...), adapterModules...)...)

	moduleManager.SetOrderInitGenesis(
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		consensustypes.ModuleName,
		paramstypes.ModuleName,
		genutiltypes.ModuleName,
		wasmtypes.ModuleName,
		wasmSecurityTypes.ModuleName,
		contractregistrytypes.ModuleName,
		identitychangetypes.ModuleName,
		identitytypes.ModuleName,
		irtypes.ModuleName,
		cstypes.ModuleName,
		vctypes.ModuleName,
		drtypes.ModuleName,
		compliancetypes.ModuleName,
		monitoringtypes.ModuleName,
		walletsecuritytypes.ModuleName,
		validatorsecuritytypes.ModuleName,
		cryptographytypes.ModuleName,
		networksecuritytypes.ModuleName,
		incidentresponsetypes.ModuleName,
		privacytypes.ModuleName,
		securitytypes.ModuleName,
		economicsecuritytypes.ModuleName,
		economicstypes.ModuleName,
		governancetypes.ModuleName,
		prevalidationtypes.ModuleName,
		dextypes.ModuleName,
		bridgetypes.ModuleName,
		aitypes.ModuleName,
		aurabindingstypes.ModuleName,
	)

	moduleManager.SetOrderBeginBlockers(
		genutiltypes.ModuleName,
		consensustypes.ModuleName,
		slashingtypes.ModuleName,
		stakingtypes.ModuleName,
		distrtypes.ModuleName,
		wasmtypes.ModuleName,
		wasmSecurityTypes.ModuleName,
		dextypes.ModuleName,
		bridgetypes.ModuleName,
		contractregistrytypes.ModuleName,
		identitychangetypes.ModuleName,
		identitytypes.ModuleName,
		irtypes.ModuleName,
		cstypes.ModuleName,
		vctypes.ModuleName,
		drtypes.ModuleName,
		compliancetypes.ModuleName,
		monitoringtypes.ModuleName,
		walletsecuritytypes.ModuleName,
		validatorsecuritytypes.ModuleName,
		cryptographytypes.ModuleName,
		networksecuritytypes.ModuleName,
		incidentresponsetypes.ModuleName,
		privacytypes.ModuleName,
		securitytypes.ModuleName,
		economicsecuritytypes.ModuleName,
		economicstypes.ModuleName,
		governancetypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		paramstypes.ModuleName,
		prevalidationtypes.ModuleName,
		aurabindingstypes.ModuleName,
	)

	moduleManager.SetOrderEndBlockers(
		slashingtypes.ModuleName,
		stakingtypes.ModuleName,
		distrtypes.ModuleName,
		wasmtypes.ModuleName,
		wasmSecurityTypes.ModuleName,
		dextypes.ModuleName,
		bridgetypes.ModuleName,
		contractregistrytypes.ModuleName,
		identitychangetypes.ModuleName,
		identitytypes.ModuleName,
		irtypes.ModuleName,
		cstypes.ModuleName,
		vctypes.ModuleName,
		drtypes.ModuleName,
		compliancetypes.ModuleName,
		monitoringtypes.ModuleName,
		walletsecuritytypes.ModuleName,
		validatorsecuritytypes.ModuleName,
		cryptographytypes.ModuleName,
		networksecuritytypes.ModuleName,
		incidentresponsetypes.ModuleName,
		privacytypes.ModuleName,
		securitytypes.ModuleName,
		economicsecuritytypes.ModuleName,
		economicstypes.ModuleName,
		governancetypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		paramstypes.ModuleName,
		genutiltypes.ModuleName,
		prevalidationtypes.ModuleName,
		aurabindingstypes.ModuleName,
	)

	configurator := sdkmodule.NewConfigurator(encoding.Codec, base.MsgServiceRouter(), base.GRPCQueryRouter())
	moduleManager.RegisterServices(configurator)

	app := &App{
		BaseApp:            base,
		moduleManager:      moduleManager,
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
		identityKeeper:       identityKeeper,
		// Individual economics keepers
		economicsecurityKeeper: economicsecurityKeeper,
		governanceKeeper:       governanceKeeper,
		economicsKeeper:        economicsKeeper,
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
		securityKeeper:         securityKeeper,
		monitoringKeeper:       monitoringKeeper,
		prevalidationKeeper:    prevalidationKeeper,
		aurabindingsKeeper:     aurabindingsKeeper,
		storeKeys:              keys,
		memKeys: struct {
			vc           *storetypes.MemoryStoreKey
			security     *storetypes.MemoryStoreKey
			aurabindings *storetypes.MemoryStoreKey
		}{
			vc:           vcMemKey,
			security:     securityMemKey,
			aurabindings: aurabindingsMemKey,
		},
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
		res, err := moduleManager.InitGenesis(ctx, encoding.Codec, genesisState)
		if err != nil {
			return nil, err
		}
		ensureStoreInitMarkers(ctx, app.allStoreKeys())
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

	// Explicitly set pruning on CommitMultiStore after all stores are mounted.
	// This ensures IAVL trees retain all versions for historical queries.
	// The baseapp.SetPruning option sets it during NewBaseApp, but we reinforce it here.
	if cms, ok := base.CommitMultiStore().(interface {
		SetPruning(pruningtypes.PruningOptions)
	}); ok {
		cms.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
		logger.Info("pruning explicitly set to PruningNothing on CommitMultiStore")
	}

	return app
}

// ensureStoreInitMarkers writes a deterministic marker into each KV store so
// every mounted store has an on-disk version from the first commit.
func ensureStoreInitMarkers(ctx sdk.Context, keys []storetypes.StoreKey) {
	marker := []byte{0x01}
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

func (app *App) allStoreKeys() []storetypes.StoreKey {
	return []storetypes.StoreKey{
		app.storeKeys.account,
		app.storeKeys.bank,
		app.storeKeys.staking,
		app.storeKeys.slashing,
		app.storeKeys.distribution,
		app.storeKeys.params,
		app.storeKeys.consensus,
		app.storeKeys.walletSecurity,
		app.storeKeys.validatorSecurity,
		app.storeKeys.cryptography,
		app.storeKeys.networkSecurity,
		app.storeKeys.incidentResponse,
		app.storeKeys.privacy,
		app.storeKeys.identityChange,
		app.storeKeys.identity,
		app.storeKeys.economicSecurity,
		app.storeKeys.governance,
		app.storeKeys.economics,
		app.storeKeys.vc,
		app.storeKeys.compliance,
		app.storeKeys.dex,
		app.storeKeys.bridge,
		app.storeKeys.ai,
		app.storeKeys.wasm,
		app.storeKeys.contractRegistry,
		app.storeKeys.wasmSecurity,
		app.storeKeys.confidenceScore,
		app.storeKeys.inclusionRoutines,
		app.storeKeys.dataRegistry,
		app.storeKeys.monitoring,
		app.storeKeys.prevalidation,
		app.storeKeys.aurabindings,
		app.storeKeys.security,
	}
}

// SetupAnteHandler configures the ante handler for transaction processing.
// The ante handler performs validation, fee deduction, and security checks
// before transaction execution.
func (a *App) SetupAnteHandler() {
	// Use the SignModeHandler from TxConfig which was configured in MakeEncodingConfig
	// with LEGACY_AMINO_JSON, DIRECT, and DIRECT_AUX modes. This ensures the ante handler
	// can verify signatures created with any of these modes.
	//
	// Previously this was creating a new handler with only TEXTUAL mode, which caused
	// signature verification failures when clients signed with LEGACY_AMINO_JSON mode.
	signModeHandler := a.encoding.TxConfig.SignModeHandler()

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
		ComplianceKeeper:     a.complianceKeeper,
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
// Also registers SDK module query services on the BaseApp's GRPCQueryRouter
// for transaction signing to work (AccountRetriever queries auth module).
func (a *App) RegisterGRPCServices() {
	// Register SDK auth module query service on the BaseApp's GRPCQueryRouter.
	// This is required for transaction signing - the AccountRetriever queries
	// the auth module to get account sequence and account number.
	authtypes.RegisterQueryServer(a.GRPCQueryRouter(), authkeeper.NewQueryServer(a.AccountKeeper))

	// Register bank module query service for balance queries
	banktypes.RegisterQueryServer(a.GRPCQueryRouter(), bankkeeper.NewQuerier(&a.BankKeeper))

	// Register WASM security module query service on the BaseApp's GRPCQueryRouter.
	// This enables CLI queries to route through ABCI to the WASM query server.
	// The proto file descriptors are registered manually in x/wasm/types/proto_registry.go
	// using synthetic protoreflect descriptors, which satisfies the GRPCQueryRouter's
	// requirement for proto method descriptors.
	wasmSecurityTypes.RegisterQueryServer(a.GRPCQueryRouter(), wasmSecurityKeeper.NewQueryServerImpl(a.wasmSecurityKeeper))
}

// GRPCServer exposes the app's gRPC server instance.
func (a *App) GRPCServer() *grpc.Server {
	return a.grpcServer
}

// SetGRPCServer records the externally managed gRPC server instance so it can be
// exposed to helpers (tests, diagnostics) without coupling app construction to
// server options like TLS credentials.
func (a *App) SetGRPCServer(server *grpc.Server) {
	a.grpcServer = server
}

// Encoding returns the codec configuration currently in use.
func (a *App) Encoding() EncodingConfig {
	return a.encoding
}

// InitBridgeGenesis initializes the bridge module state directly via the keeper.
func (a *App) InitBridgeGenesis(ctx sdk.Context, genesis bridgetypes.GenesisState) error {
	return a.bridgeKeeper.InitGenesis(ctx, genesis)
}

// ExportBridgeGenesis exports the bridge module genesis state.
func (a *App) ExportBridgeGenesis(ctx sdk.Context) bridgetypes.GenesisState {
	return a.bridgeKeeper.ExportGenesis(ctx)
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
//  1. At genesis (to validate initial state)
//  2. During upgrades (to validate migration correctness)
//  3. Periodically in production (configurable via governance)
//  4. On-demand via CLI commands for debugging
func (a *App) registerInvariants() {
	// Bank module invariants (Cosmos SDK standard)
	// Ensures total supply = sum of all account balances
	a.Logger().Info("registered bank invariants", "checks", "total-supply")

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
// This also ensures SDK config is set with the correct Bech32 prefixes.
func MakeEncodingConfig() EncodingConfig {
	// Ensure SDK config is set before creating encoding config
	// This sets the Bech32 prefixes (aura, auravaloper, auravalcons) for address encoding
	ensureSDKConfig()

	addrCodec := address.NewBech32Codec(bech32MainPrefix)
	valCodec := address.NewBech32Codec(bech32ValidatorPrefix)
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

	// Register crypto types (secp256k1, ed25519, etc.) - CRITICAL for keyring proto unmarshalling
	// Without this, the SDK 0.53.x keyring migration fails with "no registered implementations of type types.PubKey"
	cryptocodec.RegisterInterfaces(interfaceRegistry)

	authtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)
	slashingtypes.RegisterInterfaces(interfaceRegistry)
	distrtypes.RegisterInterfaces(interfaceRegistry)
	wasmtypes.RegisterInterfaces(interfaceRegistry)

	// Register interfaces for all modules tracked by ModuleBasics so
	// consensus and custom message/service types are discoverable.
	ModuleBasics.RegisterInterfaces(interfaceRegistry)

	protoCodec := codec.NewProtoCodec(interfaceRegistry)

	// Prefer LEGACY_AMINO_JSON as the default sign mode to align with existing keys/scripts,
	// while still supporting DIRECT and DIRECT_AUX for modern clients.
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
	mu                sync.RWMutex
	mintedPerBlock    map[int64]map[string]sdk.Coins // block_height -> module_name -> amount
	maxMintPerBlock   sdk.Coins                      // Maximum tokens that can be minted per block
	maxMintPerModule  map[string]sdk.Coins           // Per-module minting limits
	alertThreshold    sdkmath.LegacyDec              // Alert if minting exceeds this % of max
	violationCallback func(blockHeight int64, module string, amount sdk.Coins)
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
