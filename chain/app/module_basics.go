// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"

	addrcodec "cosmossdk.io/core/address"
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	consensus "github.com/cosmos/cosmos-sdk/x/consensus"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil"
	params "github.com/cosmos/cosmos-sdk/x/params"
	slashing "github.com/cosmos/cosmos-sdk/x/slashing"
	staking "github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	// Aura modules
	aurabindings "github.com/aequitas/aura/chain/x/aura-bindings"
	"github.com/aequitas/aura/chain/x/aiassistant"
	"github.com/aequitas/aura/chain/x/bridge"
	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
	compliance "github.com/aequitas/aura/chain/x/compliance"
	"github.com/aequitas/aura/chain/x/confidencescore"
	cstypes "github.com/aequitas/aura/chain/x/confidencescore/types"
	"github.com/aequitas/aura/chain/x/contractregistry"
	"github.com/aequitas/aura/chain/x/cryptography"
	"github.com/aequitas/aura/chain/x/dataregistry"
	drtypes "github.com/aequitas/aura/chain/x/dataregistry/types"
	"github.com/aequitas/aura/chain/x/dex"
	dextypes "github.com/aequitas/aura/chain/x/dex/types"
	"github.com/aequitas/aura/chain/x/economics"
	"github.com/aequitas/aura/chain/x/economicsecurity"
	"github.com/aequitas/aura/chain/x/governance"
	"github.com/aequitas/aura/chain/x/identity"
	identitychange "github.com/aequitas/aura/chain/x/identitychange"
	"github.com/aequitas/aura/chain/x/incidentresponse"
	"github.com/aequitas/aura/chain/x/inclusionroutines"
	"github.com/aequitas/aura/chain/x/monitoring"
	"github.com/aequitas/aura/chain/x/networksecurity"
	"github.com/aequitas/aura/chain/x/prevalidation"
	"github.com/aequitas/aura/chain/x/privacy"
	"github.com/aequitas/aura/chain/x/security"
	"github.com/aequitas/aura/chain/x/validatorsecurity"
	"github.com/aequitas/aura/chain/x/vcregistry"
	vctypes "github.com/aequitas/aura/chain/x/vcregistry/types"
	"github.com/aequitas/aura/chain/x/walletsecurity"
	aurawasm "github.com/aequitas/aura/chain/x/wasm"
)

// ModuleBasics exposes the set of Cosmos SDK and Aura modules that implement
// module.AppModuleBasic so we can reuse the canonical genesis CLI (add-genesis-account,
// gentx, collect-gentxs, validate-genesis).
var ModuleBasics = module.NewBasicManager(
	// Cosmos SDK core modules
	auth.AppModuleBasic{},
	bank.AppModuleBasic{},
	staking.AppModuleBasic{},
	slashing.AppModuleBasic{},
	distribution.AppModuleBasic{},
	params.AppModuleBasic{},
	consensus.AppModuleBasic{},
	genutil.AppModuleBasic{},
	// CosmWasm base module
	wasm.AppModuleBasic{},
	// Aura modules registered for genesis CLI validation
	contractregistry.AppModuleBasic{},
	compliance.AppModuleBasic{},
	economicsecurity.AppModuleBasic{},
	economics.AppModuleBasic{},
	security.AppModuleBasic{},
	identity.AppModuleBasic{},
	identitychange.AppModuleBasic{},
	inclusionroutines.AppModuleBasic{},
	monitoring.AppModuleBasic{},
	prevalidation.AppModuleBasic{},
	aurabindings.AppModuleBasic{},
	governance.AppModuleBasic{},
	confidencescore.AppModuleBasic{},
	dataregistry.AppModuleBasic{},
	vcregistry.AppModuleBasic{},
	bridge.AppModuleBasic{},
	dex.AppModuleBasic{},
	cryptography.AppModuleBasic{},
	aurawasm.AppModuleBasic{},
	walletsecurity.AppModuleBasic{},
	networksecurity.AppModuleBasic{},
	incidentresponse.AppModuleBasic{},
	validatorsecurity.AppModuleBasic{},
	privacy.AppModuleBasic{},
	aiassistant.AppModuleBasic{},
)

// AccountAddressCodec returns the bech32 codec for account addresses using the
// app's configured prefixes.
func AccountAddressCodec() addrcodec.Codec {
	ensureSDKConfig()
	return sdkaddress.NewBech32Codec(bech32MainPrefix)
}

// ValidatorAddressCodec returns the bech32 codec for validator operator addresses.
func ValidatorAddressCodec() addrcodec.Codec {
	ensureSDKConfig()
	return sdkaddress.NewBech32Codec(bech32ValidatorPrefix)
}

// ConsensusAddressCodec returns the bech32 codec for validator consensus addresses.
func ConsensusAddressCodec() addrcodec.Codec {
	ensureSDKConfig()
	return sdkaddress.NewBech32Codec(bech32ConsensusPrefix)
}

// basicAdapter provides a minimal module.AppModuleBasic implementation for modules
// that do not expose one (legacy Aura modules).
type basicAdapter struct {
	name               string
	defaultGenesis     func(cdc codec.JSONCodec) json.RawMessage
	validateGenesis    func(cdc codec.JSONCodec, cfg client.TxEncodingConfig, bz json.RawMessage) error
	registerInterfaces func(codectypes.InterfaceRegistry)
}

func (b basicAdapter) Name() string { return b.name }

func (basicAdapter) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

func (b basicAdapter) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	if b.registerInterfaces != nil {
		b.registerInterfaces(reg)
	}
}

func (b basicAdapter) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	if b.defaultGenesis != nil {
		return b.defaultGenesis(cdc)
	}
	return json.RawMessage(`{}`)
}

func (b basicAdapter) ValidateGenesis(cdc codec.JSONCodec, cfg client.TxEncodingConfig, bz json.RawMessage) error {
	if b.validateGenesis != nil {
		return b.validateGenesis(cdc, cfg, bz)
	}
	return nil
}

func (basicAdapter) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

func (basicAdapter) GetTxCmd() *cobra.Command { return nil }

func (basicAdapter) GetQueryCmd() *cobra.Command { return nil }

// newDexModuleBasic wires dex message interfaces into ModuleBasics without needing a full AppModuleBasic.
func newDexModuleBasic() module.AppModuleBasic {
	return basicAdapter{
		name: dextypes.ModuleName,
	}
}

func newBridgeModuleBasic() module.AppModuleBasic {
	return basicAdapter{
		name: bridgetypes.ModuleName,
	}
}

func newConfidenceScoreModuleBasic() module.AppModuleBasic {
	return basicAdapter{
		name: cstypes.ModuleName,
	}
}

func newVCRegistryModuleBasic() module.AppModuleBasic {
	return basicAdapter{
		name: vctypes.ModuleName,
	}
}

func newDataRegistryModuleBasic() module.AppModuleBasic {
	return basicAdapter{
		name: drtypes.ModuleName,
	}
}
