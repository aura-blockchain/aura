package aurabindings_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/app"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdkmath "cosmossdk.io/math"
)

type IntegrationTestSuite struct {
	suite.Suite
	App *app.App
	Ctx sdk.Context
	MsgServer wasmtypes.MsgServer
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupTest() {
	s.App = app.NewApp()
	s.Ctx = s.App.NewUncachedContext(false, tmproto.Header{Time: time.Now()})
	s.MsgServer = keeper.NewMsgServerImpl(s.App.WasmKeeper)

	// Set up creator account with funds
	creator := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc := s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, creator)
	s.App.AccountKeeper.SetAccount(s.Ctx, acc)

	// Set wasm parameters
	params := wasmtypes.DefaultParams() // Start with default parameters
	params.CodeUploadAccess = wasmtypes.AllowEverybody
	params.InstantiateDefaultPermission = wasmtypes.AccessTypeEverybody
	s.App.WasmKeeper.SetParams(s.Ctx, params)

	err := s.App.BankKeeper.MintCoins(s.Ctx, wasmtypes.ModuleName, sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1000000000))))
	require.NoError(s.T(), err)
	err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, wasmtypes.ModuleName, creator, sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1000000000))))
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) TestRegisterAndGetVC() {
	// 1. Store the contract
	wasm, err := os.ReadFile("../../../contracts/binding-tester/target/wasm32-unknown-unknown/release/binding_tester.wasm")
	require.NoError(s.T(), err)

	creator := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	storeMsg := &wasmtypes.MsgStoreCode{
		Sender:       creator.String(),
		WASMByteCode: wasm,
	}

	storeResult, err := s.MsgServer.StoreCode(s.Ctx, storeMsg)
	require.NoError(s.T(), err)
	codeID := storeResult.CodeID

	// 2. Instantiate the contract
	instantiateMsg := &wasmtypes.MsgInstantiateContract{
		Sender: creator.String(),
		Admin:  creator.String(),
		CodeID: codeID,
		Label:  "binding-tester",
		Msg:    []byte(`{}`),
		Funds:  sdk.Coins{},
	}
	instantiateResult, err := s.MsgServer.InstantiateContract(s.Ctx, instantiateMsg)
	require.NoError(s.T(), err)
	addr := instantiateResult.Address

	// 3. Execute the contract
	executeMsg := &wasmtypes.MsgExecuteContract{
		Sender:   creator.String(),
		Contract: addr,
		Msg:      []byte(`{"register_vc":{"address":"` + creator.String() + `","vc_base64":"dGVzdA=="}}`),
		Funds:    sdk.Coins{},
	}
	_, err = s.MsgServer.ExecuteContract(s.Ctx, executeMsg)
	require.NoError(s.T(), err)

	// 4. Query the contract
	queryMsg := []byte(`{"get_vc":{"address":"` + creator.String() + `"}}`)
	contractAddr, err := sdk.AccAddressFromBech32(addr)
	require.NoError(s.T(), err)
	res, err := s.App.WasmKeeper.QuerySmart(s.Ctx, contractAddr, queryMsg)
	require.NoError(s.T(), err)

	var vc struct {
		Address  string `json:"address"`
		VCBase64 string `json:"vc_base64"`
	}
	err = json.Unmarshal(res, &vc)
	require.NoError(s.T(), err)

	require.Equal(s.T(), creator.String(), vc.Address)
	require.Equal(s.T(), "dGVzdA==", vc.VCBase64)
}
