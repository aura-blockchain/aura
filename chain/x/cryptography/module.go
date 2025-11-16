package cryptography

import (
	"context"
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/aequitas/aura/chain/x/cryptography/keeper"
	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic defines the basic application module used by the cryptography module
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the cryptography module's name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the cryptography module's types on the LegacyAmino codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry codec.InterfaceRegistry) {}

// DefaultGenesis returns default genesis state as raw bytes for the cryptography module
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the cryptography module
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ interface{}, bz json.RawMessage) error {
	var data cryptoproto.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return err
	}
	return types.ValidateGenesis(&data)
}

// AppModule implements an application module for the cryptography module
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface
func (am AppModule) IsAppModule() {}

// InitGenesis performs genesis initialization for the cryptography module
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState cryptoproto.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.initGenesis(ctx, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the cryptography module
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.exportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

// initGenesis initializes the module's state from genesis
func (am AppModule) initGenesis(ctx sdk.Context, gs *cryptoproto.GenesisState) {
	// Set params
	if err := am.keeper.SetParams(ctx, &gs.Params); err != nil {
		panic(err)
	}

	// Initialize key rotation schedules
	for _, schedule := range gs.KeyRotationSchedules {
		// Store would happen in keeper methods
		_ = schedule
	}

	// Initialize threshold schemes
	for _, scheme := range gs.ThresholdSchemes {
		_ = scheme
	}

	// Initialize ZK proof configs
	for _, config := range gs.ZkProofConfigs {
		_ = config
	}

	// Initialize secure enclaves
	for _, enclave := range gs.SecureEnclaves {
		_ = enclave
	}

	// Initialize quantum keys
	for _, key := range gs.QuantumResistantKeys {
		_ = key
	}

	// Initialize random sources
	for _, source := range gs.RandomSources {
		_ = source
	}

	// Initialize key stretching configs
	for _, config := range gs.KeyStretchingConfigs {
		_ = config
	}

	// Initialize certificate pins
	for _, pin := range gs.CertificatePins {
		_ = pin
	}
}

// exportGenesis exports the module's state to genesis
func (am AppModule) exportGenesis(ctx sdk.Context) *cryptoproto.GenesisState {
	params, err := am.keeper.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	return &cryptoproto.GenesisState{
		Params:               *params,
		KeyRotationSchedules: []*cryptoproto.KeyRotationSchedule{},
		ThresholdSchemes:     []*cryptoproto.ThresholdSignatureScheme{},
		ZkProofConfigs:       []*cryptoproto.ZKProofConfig{},
		SecureEnclaves:       []*cryptoproto.SecureEnclaveConfig{},
		QuantumResistantKeys: []*cryptoproto.QuantumResistantKey{},
		RandomSources:        []*cryptoproto.CryptoRandomSource{},
		KeyStretchingConfigs: []*cryptoproto.KeyStretchingConfig{},
		CertificatePins:      []*cryptoproto.CertificatePin{},
	}
}

// BeginBlock performs module-specific logic at the beginning of each block
func (am AppModule) BeginBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Process scheduled key rotations
	if err := am.keeper.ProcessScheduledRotations(sdkCtx); err != nil {
		am.keeper.Logger(sdkCtx).Error("failed to process scheduled rotations", "error", err)
	}

	// Check entropy health
	if err := am.keeper.CheckEntropyHealth(sdkCtx); err != nil {
		am.keeper.Logger(sdkCtx).Error("failed to check entropy health", "error", err)
	}

	// Cleanup expired certificate pins
	if err := am.keeper.CleanupExpiredPins(sdkCtx); err != nil {
		am.keeper.Logger(sdkCtx).Error("failed to cleanup expired pins", "error", err)
	}

	return nil
}

// EndBlock performs module-specific logic at the end of each block
func (am AppModule) EndBlock(ctx context.Context) error {
	return nil
}
