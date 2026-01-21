// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package aurabindings

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	vckeeper "github.com/aequitas/aura/chain/x/vcregistry/keeper"
	vcparams "github.com/aequitas/aura/chain/x/vcregistry/params"
	vctypes "github.com/aequitas/aura/chain/x/vcregistry/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// Ensures custom AuraMsg Cosmos messages are dispatched through the message handler
// into the VC registry keeper.
func TestMessageHandlerRegistersVC(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Set up in-memory KV store and context.
	db := dbm.NewMemDB()
	vcStoreKey := storetypes.NewKVStoreKey(vctypes.StoreKey)
	metricsProvider := metrics.NewNoOpMetrics()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metricsProvider)
	stateStore.MountStoreWithDB(vcStoreKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	sdkCtx := sdk.NewContext(stateStore, tmproto.Header{ChainID: "test-chain", Height: 1}, false, log.NewNopLogger())

	// Build VC registry keeper with backing store.
	vcParamsStore := vcparams.NewStore(*vctypes.DefaultParams())
	vcKeeper := vckeeper.NewKeeperBuilder(vcParamsStore, "authority").WithStore(vcStoreKey, cdc).Build()

	handler := NewMessageHandler(vcKeeper)

	holder := sdk.AccAddress([]byte("holder_address_1234"))
	contractAddr := sdk.AccAddress([]byte("contract_addr_12345"))

	custom := AuraMsg{
		VCRegistry: &VCRegistryMsg{
			RegisterVC: &RegisterVCMsg{
				Address:  holder.String(),
				VCBase64: "ZGVtbw==",
			},
		},
	}
	bz, err := json.Marshal(custom)
	require.NoError(t, err)

	events, resp, err := handler.DispatchMsg(sdkCtx, contractAddr, "", wasmvmtypes.CosmosMsg{Custom: bz})
	require.NoError(t, err)
	require.Empty(t, events)
	require.Empty(t, resp)

	// Ensure the VC was persisted for the holder.
	records := vcKeeper.ListUserVCs(sdkCtx, holder.String(), vctypes.VCStatus_VC_STATUS_UNSPECIFIED, vctypes.VCTypeUnspecified)
	require.Len(t, records, 1)
	require.Equal(t, holder.String(), records[0].HolderAddress)
	require.Equal(t, vctypes.VCType_VC_TYPE_CUSTOM, records[0].VcType)
	require.Equal(t, "binding_tester_vc", records[0].VcTypeCustom)
	require.Equal(t, []byte("ZGVtbw=="), records[0].CredentialHash)
	require.Equal(t, "ZGVtbw==", records[0].Metadata["vc_base64"])
}
