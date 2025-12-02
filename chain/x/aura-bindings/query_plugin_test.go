package aurabindings

import (
	"encoding/json"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"

	vckeeper "github.com/aequitas/aura/chain/x/vcregistry/keeper"
	vcparams "github.com/aequitas/aura/chain/x/vcregistry/params"
	vctypes "github.com/aequitas/aura/chain/x/vcregistry/types"
)

// Verifies the custom query plugin can return a VC record via AuraQuery.
func TestCustomQuerierReturnsVC(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	db := dbm.NewMemDB()
	metricsProvider := metrics.NewNoOpMetrics()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metricsProvider)
	vcStoreKey := storetypes.NewKVStoreKey(vctypes.StoreKey)
	stateStore.MountStoreWithDB(vcStoreKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, tmproto.Header{ChainID: "test-chain", Height: 1}, false, log.NewNopLogger())

	paramsStore := vcparams.NewStore(*vctypes.DefaultParams())
	vcKeeper := vckeeper.NewKeeperBuilder(paramsStore, "authority").WithStore(vcStoreKey, cdc).Build()

	holder := sdk.AccAddress([]byte("holder_query_12345")).String()
	vc := vctypes.VCRecord{
		VcId:           "vc-query-test",
		VcType:         vctypes.VCType_VC_TYPE_CUSTOM,
		VcTypeCustom:   "binding_tester_vc",
		HolderAddress:  holder,
		Status:         vctypes.VCStatus_VC_STATUS_ACTIVE,
		CredentialHash: []byte("dmMtdGVzdC1oYXNo"),
		Metadata:       map[string]string{"vc_base64": "dmMtdGVzdC1oYXNo"},
	}
	require.NoError(t, vcKeeper.SetVCRecord(ctx, vc))

	querier := CustomQuerier(vcKeeper, nil)

	query := AuraQuery{
		VCRegistry: &VCRegistryQuery{
			GetVC: &GetVCQuery{Address: holder},
		},
	}
	raw, err := json.Marshal(query)
	require.NoError(t, err)

	respBz, err := querier(ctx, raw)
	require.NoError(t, err)

	var resp VerifiableCredential
	require.NoError(t, json.Unmarshal(respBz, &resp))
	require.Equal(t, holder, resp.Address)
	require.Equal(t, "dmMtdGVzdC1oYXNo", resp.VCBase64)
}
