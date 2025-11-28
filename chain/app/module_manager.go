package app

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	module "github.com/cosmos/cosmos-sdk/types/module"

	"github.com/aequitas/aura/chain/x/aiassistant"
	"github.com/aequitas/aura/chain/x/bridge"
	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
	"github.com/aequitas/aura/chain/x/compliance"
	"github.com/aequitas/aura/chain/x/confidencescore"
	"github.com/aequitas/aura/chain/x/contractregistry"
	"github.com/aequitas/aura/chain/x/cryptography"
	"github.com/aequitas/aura/chain/x/dataregistry"
	"github.com/aequitas/aura/chain/x/dex"
	"github.com/aequitas/aura/chain/x/economicsecurity"
	"github.com/aequitas/aura/chain/x/governance"
	"github.com/aequitas/aura/chain/x/identitychange"
	"github.com/aequitas/aura/chain/x/incidentresponse"
	"github.com/aequitas/aura/chain/x/inclusionroutines"
	"github.com/aequitas/aura/chain/x/networksecurity"
	"github.com/aequitas/aura/chain/x/privacy"
	"github.com/aequitas/aura/chain/x/validatorsecurity"
	"github.com/aequitas/aura/chain/x/vcregistry"
	"github.com/aequitas/aura/chain/x/walletsecurity"
	"github.com/aequitas/aura/chain/x/wasm"
	aipb "github.com/aequitas/aura/proto/aura/aiassistant/v1beta1"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
	compliancepb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
	dataregistrypb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
	identitypb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	"google.golang.org/grpc"
)

// ModuleManager wires Cosmos SDK modules (eventually) into the application lifecycle.
type ModuleManager struct {
	encoding                 EncodingConfig
	aiAssistantModules       []aiassistant.AppModule
	identityChangeModules    []identitychange.AppModule
	inclusionRoutinesModules []inclusionroutines.AppModule
	confidenceScoreModules   []confidencescore.AppModule
	vcRegistryModules        []vcregistry.AppModule
	dataRegistryModules      []dataregistry.AppModule
	complianceModules        []compliance.AppModule

	// Individual security modules
	walletsecurityModules    []walletsecurity.AppModule
	validatorsecurityModules []validatorsecurity.AppModule
	cryptographyModules      []cryptography.AppModule
	networksecurityModules   []networksecurity.AppModule
	incidentresponseModules  []incidentresponse.AppModule
	privacyModules           []privacy.AppModule

	// Individual economics modules
	economicsecurityModules []economicsecurity.AppModule
	governanceModules       []governance.AppModule

	dexModules              []dex.AppModule
	bridgeModules           []bridge.AppModule
	contractRegistryModules []contractregistry.AppModule
	wasmSecurityModules     []wasm.AppModule
}

// NewModuleManager initializes a tracker for the provided modules.
func NewModuleManager(
	encoding EncodingConfig,
	aiModules []aiassistant.AppModule,
	idModules []identitychange.AppModule,
	irModules []inclusionroutines.AppModule,
	csModules []confidencescore.AppModule,
	vcModules []vcregistry.AppModule,
	drModules []dataregistry.AppModule,
	compModules []compliance.AppModule,
	walletsecurityModules []walletsecurity.AppModule,
	validatorsecurityModules []validatorsecurity.AppModule,
	cryptographyModules []cryptography.AppModule,
	networksecurityModules []networksecurity.AppModule,
	incidentresponseModules []incidentresponse.AppModule,
	privacyModules []privacy.AppModule,
	economicsecurityModules []economicsecurity.AppModule,
	governanceModules []governance.AppModule,
	dexModules []dex.AppModule,
	bridgeModules []bridge.AppModule,
	contractRegistryModules []contractregistry.AppModule,
	wasmSecurityModules []wasm.AppModule,
) ModuleManager {
	return ModuleManager{
		encoding:                 encoding,
		aiAssistantModules:       aiModules,
		identityChangeModules:    idModules,
		inclusionRoutinesModules: irModules,
		confidenceScoreModules:   csModules,
		vcRegistryModules:        vcModules,
		dataRegistryModules:      drModules,
		complianceModules:        compModules,
		walletsecurityModules:    walletsecurityModules,
		validatorsecurityModules: validatorsecurityModules,
		cryptographyModules:      cryptographyModules,
		networksecurityModules:   networksecurityModules,
		incidentresponseModules:  incidentresponseModules,
		privacyModules:           privacyModules,
		economicsecurityModules:  economicsecurityModules,
		governanceModules:        governanceModules,
		dexModules:               dexModules,
		bridgeModules:            bridgeModules,
		contractRegistryModules:  contractRegistryModules,
		wasmSecurityModules:      wasmSecurityModules,
	}
}

// RegisterGRPCServices registers each module's gRPC handlers with the provided registrar.
func (m ModuleManager) RegisterGRPCServices(server grpc.ServiceRegistrar) {
	if server == nil {
		panic("module manager: nil gRPC server registrar")
	}

	grpcServer, ok := server.(*grpc.Server)
	if !ok {
		// Only panic if we have modules that require *grpc.Server
		if len(m.walletsecurityModules) > 0 || len(m.validatorsecurityModules) > 0 ||
			len(m.cryptographyModules) > 0 || len(m.networksecurityModules) > 0 ||
			len(m.incidentresponseModules) > 0 || len(m.privacyModules) > 0 ||
			len(m.economicsecurityModules) > 0 || len(m.governanceModules) > 0 {
			panic("module manager: registrar must be *grpc.Server when security/economics modules are registered")
		}
	}

	// Register aiassistant modules
	aiServices := &aiAssistantServices{registrar: server}
	for _, module := range m.aiAssistantModules {
		module.RegisterServices(aiServices)
	}

	// Register identitychange modules
	idServices := &identityChangeServices{registrar: server}
	for _, module := range m.identityChangeModules {
		module.RegisterServices(idServices)
	}

	// Register inclusionroutines modules
	irServices := &inclusionRoutinesServices{registrar: server}
	for _, module := range m.inclusionRoutinesModules {
		module.RegisterServices(irServices)
	}

	// Register confidencescore modules
	csServices := &confidenceScoreServices{registrar: server}
	for _, module := range m.confidenceScoreModules {
		module.RegisterServices(csServices)
	}

	// Register vcregistry modules
	vcServices := &vcRegistryServices{registrar: server}
	for _, module := range m.vcRegistryModules {
		module.RegisterServices(vcServices)
	}

	// Register dataregistry modules
	drServices := &dataRegistryServices{registrar: server}
	for _, module := range m.dataRegistryModules {
		module.RegisterServices(drServices)
	}

	// Register compliance modules
	compServices := &complianceServices{registrar: server}
	for _, module := range m.complianceModules {
		module.RegisterServices(compServices)
	}

	// Register individual security modules (using Configurator pattern)
	if ok {
		for _, module := range m.walletsecurityModules {
			m.registerConfiguratorModule(grpcServer, module)
		}
		for _, module := range m.validatorsecurityModules {
			m.registerConfiguratorModule(grpcServer, module)
		}
		for _, module := range m.cryptographyModules {
			m.registerConfiguratorModule(grpcServer, module)
		}
		for _, module := range m.networksecurityModules {
			m.registerConfiguratorModule(grpcServer, module)
		}
		for _, module := range m.incidentresponseModules {
			m.registerConfiguratorModule(grpcServer, module)
		}
		for _, module := range m.privacyModules {
			m.registerConfiguratorModule(grpcServer, module)
		}
	}

	// Register individual economics modules (using Configurator pattern)
	if ok {
		for _, module := range m.economicsecurityModules {
			m.registerConfiguratorModule(grpcServer, module)
		}
		for _, module := range m.governanceModules {
			m.registerConfiguratorModule(grpcServer, module)
		}
	}

	// Register dex modules
	dexServices := &dexServices{registrar: server}
	for _, module := range m.dexModules {
		module.RegisterServices(dexServices)
	}

	// Register bridge modules
	bridgeServices := &bridgeServices{registrar: server}
	for _, module := range m.bridgeModules {
		module.RegisterServices(bridgeServices)
	}

	// Register contract registry modules (using Configurator pattern)
	if ok {
		for _, module := range m.contractRegistryModules {
			m.registerConfiguratorModule(grpcServer, module)
		}
	}

	// Register wasm security modules (using Configurator pattern)
	if ok {
		for _, module := range m.wasmSecurityModules {
			m.registerConfiguratorModule(grpcServer, module)
		}
	}
}

// BeginBlock fans out the BeginBlock call (with context when available) to all modules.
func (m ModuleManager) BeginBlock(ctx context.Context) {
	for _, module := range m.aiAssistantModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.identityChangeModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.inclusionRoutinesModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.confidenceScoreModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.vcRegistryModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.dataRegistryModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.complianceModules {
		callBeginBlock(ctx, module)
	}

	// Individual security modules
	for _, module := range m.walletsecurityModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.validatorsecurityModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.cryptographyModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.networksecurityModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.incidentresponseModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.privacyModules {
		callBeginBlock(ctx, module)
	}

	// Individual economics modules
	for _, module := range m.economicsecurityModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.governanceModules {
		callBeginBlock(ctx, module)
	}

	for _, module := range m.dexModules {
		callBeginBlock(ctx, module)
	}
	for _, module := range m.bridgeModules {
		callBeginBlock(ctx, module)
	}

	// Contract registry cleanup every 100 blocks
	for _, module := range m.contractRegistryModules {
		callBeginBlock(ctx, module)
	}

	for _, module := range m.wasmSecurityModules {
		callBeginBlock(ctx, module)
	}
}

// InitBridgeGenesis runs InitGenesis for each registered bridge module.
func (m ModuleManager) InitBridgeGenesis(ctx sdk.Context, genesis bridgetypes.GenesisState) error {
	for _, module := range m.bridgeModules {
		if err := module.InitGenesis(ctx, genesis); err != nil {
			return err
		}
	}
	return nil
}

// ExportBridgeGenesis exports the genesis state for each registered bridge module.
func (m ModuleManager) ExportBridgeGenesis(ctx sdk.Context) []bridgetypes.GenesisState {
	if len(m.bridgeModules) == 0 {
		return nil
	}
	states := make([]bridgetypes.GenesisState, 0, len(m.bridgeModules))
	for _, module := range m.bridgeModules {
		states = append(states, module.ExportGenesis(ctx))
	}
	return states
}

type beginBlockWithContext interface {
	BeginBlock(context.Context)
}

type beginBlockNoContext interface {
	BeginBlock()
}

func callBeginBlock(ctx context.Context, module interface{}) {
	if module == nil {
		return
	}
	if bbCtx, ok := module.(beginBlockWithContext); ok {
		bbCtx.BeginBlock(ctx)
		return
	}
	if bb, ok := module.(beginBlockNoContext); ok {
		bb.BeginBlock()
	}
}

type moduleConfigurator interface {
	RegisterServices(module.Configurator)
}

func (m ModuleManager) registerConfiguratorModule(grpcServer *grpc.Server, mod moduleConfigurator) {
	cfg := module.NewConfigurator(m.encoding.Codec, grpcServer, grpcServer)
	mod.RegisterServices(cfg)
	if err := cfg.Error(); err != nil {
		panic(err)
	}
}

type aiAssistantServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *aiAssistantServices) RegisterMsgServer(server aipb.MsgServer) {
	aipb.RegisterMsgServer(s.registrar, server)
}

func (s *aiAssistantServices) RegisterQueryServer(server aipb.QueryServer) {
	aipb.RegisterQueryServer(s.registrar, server)
}

// identityChangeServices implements identitychange.ModuleServices
type identityChangeServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *identityChangeServices) RegisterMsgServer(server identitypb.MsgServer) {
	identitypb.RegisterMsgServer(s.registrar, server)
}

func (s *identityChangeServices) RegisterQueryServer(server identitypb.QueryServer) {
	identitypb.RegisterQueryServer(s.registrar, server)
}

// inclusionRoutinesServices implements inclusionroutines.ModuleServices
type inclusionRoutinesServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *inclusionRoutinesServices) RegisterMsgServer(server inclusionroutinespb.MsgServer) {
	inclusionroutinespb.RegisterMsgServer(s.registrar, server)
}

func (s *inclusionRoutinesServices) RegisterQueryServer(server inclusionroutinespb.QueryServer) {
	inclusionroutinespb.RegisterQueryServer(s.registrar, server)
}

// confidenceScoreServices implements confidencescore.ModuleServices
type confidenceScoreServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *confidenceScoreServices) RegisterMsgServer(server confidencescorepb.MsgServer) {
	confidencescorepb.RegisterMsgServer(s.registrar, server)
}

func (s *confidenceScoreServices) RegisterQueryServer(server confidencescorepb.QueryServer) {
	confidencescorepb.RegisterQueryServer(s.registrar, server)
}

// vcRegistryServices implements vcregistry.ModuleServices
type vcRegistryServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *vcRegistryServices) RegisterMsgServer(server vcregistrypb.MsgServer) {
	vcregistrypb.RegisterMsgServer(s.registrar, server)
}

func (s *vcRegistryServices) RegisterQueryServer(server vcregistrypb.QueryServer) {
	vcregistrypb.RegisterQueryServer(s.registrar, server)
}

// dataRegistryServices implements dataregistry.ModuleServices
type dataRegistryServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *dataRegistryServices) RegisterMsgServer(server dataregistrypb.MsgServer) {
	dataregistrypb.RegisterMsgServer(s.registrar, server)
}

func (s *dataRegistryServices) RegisterQueryServer(server dataregistrypb.QueryServer) {
	dataregistrypb.RegisterQueryServer(s.registrar, server)
}

// complianceServices implements compliance.ModuleServices
type complianceServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *complianceServices) RegisterMsgServer(server compliancepb.MsgServer) {
	compliancepb.RegisterMsgServer(s.registrar, server)
}

func (s *complianceServices) RegisterQueryServer(server compliancepb.QueryServer) {
	compliancepb.RegisterQueryServer(s.registrar, server)
}

// dexServices implements dex.ModuleServices
type dexServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *dexServices) RegisterMsgServer(server dexpb.MsgServer) {
	if server == nil {
		panic("dex: nil msg server")
	}
	dexpb.RegisterMsgServer(s.registrar, server)
}

func (s *dexServices) RegisterQueryServer(server dexpb.QueryServer) {
	if server == nil {
		panic("dex: nil query server")
	}
	dexpb.RegisterQueryServer(s.registrar, server)
}

// bridgeServices implements bridge.ModuleServices
type bridgeServices struct {
	registrar grpc.ServiceRegistrar
}

func (s *bridgeServices) RegisterMsgServer(server bridgepb.MsgServer) {
	if server == nil {
		panic("bridge: nil msg server")
	}
	bridgepb.RegisterMsgServer(s.registrar, server)
}

func (s *bridgeServices) RegisterQueryServer(server bridgepb.QueryServer) {
	if server == nil {
		panic("bridge: nil query server")
	}
	bridgepb.RegisterQueryServer(s.registrar, server)
}

// ============================================================================
// MODULE MIGRATION SUPPORT
// ============================================================================

// RunMigrations executes migrations for all registered modules.
func (m ModuleManager) RunMigrations(
	ctx context.Context,
	configurator module.Configurator,
	fromVM module.VersionMap,
) (module.VersionMap, error) {
	if fromVM == nil {
		fromVM = make(module.VersionMap)
	}

	// Create a new version map for tracking updates
	toVM := make(module.VersionMap)
	for k, v := range fromVM {
		toVM[k] = v
	}

	// Migration order based on dependency graph
	// Tier 1: Compliance, Security modules, Economics modules
	if err := m.migrateModules(ctx, "compliance", m.complianceModules, toVM); err != nil {
		return nil, err
	}
	if err := m.migrateModules(ctx, "walletsecurity", m.walletsecurityModules, toVM); err != nil {
		return nil, err
	}
	if err := m.migrateModules(ctx, "validatorsecurity", m.validatorsecurityModules, toVM); err != nil {
		return nil, err
	}
	if err := m.migrateModules(ctx, "cryptography", m.cryptographyModules, toVM); err != nil {
		return nil, err
	}
	if err := m.migrateModules(ctx, "networksecurity", m.networksecurityModules, toVM); err != nil {
		return nil, err
	}
	if err := m.migrateModules(ctx, "incidentresponse", m.incidentresponseModules, toVM); err != nil {
		return nil, err
	}
	if err := m.migrateModules(ctx, "privacy", m.privacyModules, toVM); err != nil {
		return nil, err
	}
	if err := m.migrateModules(ctx, "economicsecurity", m.economicsecurityModules, toVM); err != nil {
		return nil, err
	}
	if err := m.migrateModules(ctx, "governance", m.governanceModules, toVM); err != nil {
		return nil, err
	}

	// Tier 2: Identity, DataRegistry
	if err := m.migrateModules(ctx, "identitychange", m.identityChangeModules, toVM); err != nil {
		return nil, err
	}
	if err := m.migrateModules(ctx, "dataregistry", m.dataRegistryModules, toVM); err != nil {
		return nil, err
	}

	// Tier 3: InclusionRoutines
	if err := m.migrateModules(ctx, "inclusionroutines", m.inclusionRoutinesModules, toVM); err != nil {
		return nil, err
	}

	// Tier 4: ConfidenceScore
	if err := m.migrateModules(ctx, "confidencescore", m.confidenceScoreModules, toVM); err != nil {
		return nil, err
	}

	// Tier 5: VCRegistry
	if err := m.migrateModules(ctx, "vcregistry", m.vcRegistryModules, toVM); err != nil {
		return nil, err
	}

	// Tier 6: ContractRegistry
	if err := m.migrateModules(ctx, "contractregistry", m.contractRegistryModules, toVM); err != nil {
		return nil, err
	}

	// Tier 7: DEX, Bridge, AIAssistant
	if err := m.migrateModules(ctx, "dex", m.dexModules, toVM); err != nil {
		return nil, err
	}
	if err := m.migrateModules(ctx, "bridge", m.bridgeModules, toVM); err != nil {
		return nil, err
	}
	if err := m.migrateModules(ctx, "aiassistant", m.aiAssistantModules, toVM); err != nil {
		return nil, err
	}

	// Tier 8: WASM modules (last to ensure all dependencies are migrated)
	if err := m.migrateModules(ctx, "wasmsecurity", m.wasmSecurityModules, toVM); err != nil {
		return nil, err
	}

	return toVM, nil
}

// migrateModules is a helper that attempts to migrate a list of modules.
func (m ModuleManager) migrateModules(
	ctx context.Context,
	moduleName string,
	modules interface{},
	versionMap module.VersionMap,
) error {
	// For now, we just mark the module as version 1 if not already in the map
	if _, exists := versionMap[moduleName]; !exists {
		versionMap[moduleName] = 1
	}
	return nil
}
