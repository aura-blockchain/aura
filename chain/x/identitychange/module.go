// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package identitychange

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/types"
	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
)

type ModuleServices interface {
	RegisterMsgServer(identitychangepb.MsgServer)
	RegisterQueryServer(identitychangepb.QueryServer)
}

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterGRPCGatewayRoutes is a no-op placeholder to satisfy the AppModuleBasic interface.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

// RegisterLegacyAminoCodec registers amino types (no-op, protobuf-first).
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// RegisterInterfaces registers protobuf interfaces.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (AppModuleBasic) RegisterServices(ModuleServices) {}

// DefaultGenesis returns default genesis.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis validates genesis state.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gen identitychangepb.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gen); err != nil {
		return err
	}
	return types.ValidateGenesisState(&gen)
}

type AppModule struct {
	AppModuleBasic
	keeper *keeper.Keeper
}

func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{AppModuleBasic: AppModuleBasic{}, keeper: k}
}

func (AppModule) Name() string { return types.ModuleName }

func (m AppModule) RegisterServices(cfg module.Configurator) {
	identitychangepb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(m.keeper))
	identitychangepb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(m.keeper))
}

func (m AppModule) BeginBlock() {}

func (m AppModule) EndBlock() {}

// RegisterInvariants registers the module's invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	// NOTE: Future enhancement - Uncomment when invariants.go is fixed
	// keeper.RegisterInvariants(ir, am.keeper)
}

// InitGenesis initializes module state.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) ([]abci.ValidatorUpdate, error) {
	var genesis identitychangepb.GenesisState
	if err := cdc.UnmarshalJSON(data, &genesis); err != nil {
		return nil, err
	}
	if err := am.keeper.InitGenesis(ctx, genesis); err != nil {
		return nil, err
	}
	return []abci.ValidatorUpdate{}, nil
}

// ExportGenesis exports module state.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) (json.RawMessage, error) {
	gen := am.keeper.ExportGenesis(ctx)
	return cdc.MarshalJSON(&gen)
}

// ConsensusVersion implements appmodule.AppModule.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// IsOnePerModuleType tags this module for depinject compatibility.
func (AppModule) IsOnePerModuleType() {}

// IsAppModule tags this module for module manager compatibility.
func (AppModule) IsAppModule() {}
