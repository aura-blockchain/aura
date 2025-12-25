// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package dataregistry

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/aequitas/aura/chain/x/dataregistry/keeper"
	"github.com/aequitas/aura/chain/x/dataregistry/types"
	pb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
)

// AppModuleBasic defines the basic application module
type AppModuleBasic struct{}

// Name returns the module name
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers amino types (no-op for proto-only module).
func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// RegisterGRPCGatewayRoutes is a no-op placeholder to satisfy the AppModuleBasic interface.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

// RegisterInterfaces wires the data registry msg service into the interface registry.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &pb.Msg_serviceDesc)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&pb.MsgStoreDataItem{},
		&pb.MsgUpdateDataItem{},
		&pb.MsgDeleteDataItem{},
		&pb.MsgVerifyDataItem{},
		&pb.MsgRevokeDataItem{},
	)
}

// AppModule implements the app module interface
type AppModule struct {
	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule instance
func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{keeper: k}
}

// Name returns the module name
func (AppModule) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers amino types (no-op for proto-only module)
func (AppModule) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register query handler client
	// This will be implemented when query services are added
}

// RegisterInterfaces registers the module's interface types
func (AppModule) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &pb.Msg_serviceDesc)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&pb.MsgStoreDataItem{},
		&pb.MsgUpdateDataItem{},
		&pb.MsgDeleteDataItem{},
		&pb.MsgVerifyDataItem{},
		&pb.MsgRevokeDataItem{},
	)
}

// RegisterServices registers the module's message and query servers
func (m AppModule) RegisterServices(cfg module.Configurator) {
	pb.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(m.keeper))
	pb.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(m.keeper))
}

// BeginBlock executes all ABCI BeginBlock logic
func (m AppModule) BeginBlock() {
	// Periodically check for expired data items
	// In production, implement expiration logic
}

// EndBlock executes all ABCI EndBlock logic
func (m AppModule) EndBlock() {}

// InitGenesis initializes module state from genesis
func (m AppModule) InitGenesis(ctx context.Context, genesis types.GenesisState) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return m.keeper.InitGenesis(sdkCtx, genesis)
}

// ExportGenesis exports module state for genesis
func (m AppModule) ExportGenesis(ctx context.Context) types.GenesisState {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return m.keeper.ExportGenesis(sdkCtx)
}

// IsOnePerModuleType tags this module for depinject one-per-module compatibility.
func (AppModule) IsOnePerModuleType() {}

// IsAppModule tags this module for Cosmos SDK module manager compatibility.
func (AppModule) IsAppModule() {}
