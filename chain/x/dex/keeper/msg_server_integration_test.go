package keeper

import (
	"crypto/sha256"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	testutil "github.com/aequitas/aura/chain/testing/testutil"
	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
)

type MsgServerIntegrationSuite struct {
	suite.Suite

	ctx       sdk.Context
	keeper    *Keeper
	bank      *testutil.MockBankKeeper
	accounts  *testutil.MockAccountKeeper
	msgServer dexpb.MsgServer
}

func TestMsgServerIntegrationSuite(t *testing.T) {
	suite.Run(t, new(MsgServerIntegrationSuite))
}

func (suite *MsgServerIntegrationSuite) SetupTest() {
	// Configure SDK with Aura-specific prefixes (safe to call multiple times)
	keepertest.ConfigureSDK()

	input := keepertest.CreateTestInput(suite.T())
	suite.bank = testutil.NewMockBankKeeper()
	suite.accounts = testutil.NewMockAccountKeeper()

	suite.keeper = NewKeeper(input.Cdc, input.StoreKey, suite.bank, suite.accounts, nil, nil)
	suite.ctx = input.Ctx
	suite.msgServer = NewMsgServerImpl(suite.keeper)
}

func (suite *MsgServerIntegrationSuite) TestCreatePoolHappyPath() {
	creator := suite.addr("creator")
	suite.fund(creator, sdk.NewCoins(
		sdk.NewInt64Coin("uaura", 3_000_000_000),
		sdk.NewInt64Coin("usdt", 3_000_000_000),
	))

	resp, err := suite.msgServer.CreatePool(
		sdk.WrapSDKContext(suite.ctx),
		&dexpb.MsgCreatePool{
			Creator: creator,
			DenomA:  "uaura",
			DenomB:  "usdt",
			AmountA: sdk.NewInt64Coin("uaura", 1_200_000_000),
			AmountB: sdk.NewInt64Coin("usdt", 1_200_000_000),
		},
	)

	suite.Require().NoError(err)
	suite.NotEmpty(resp.PoolId)
	suite.NotEmpty(resp.LpTokens)

	pool := suite.keeper.GetPool(suite.ctx, resp.PoolId)
	suite.Require().NotNil(pool)
	suite.Equal("uaura", pool.DenomA)
	suite.Equal("usdt", pool.DenomB)
}

func (suite *MsgServerIntegrationSuite) TestAddLiquidityHappyPath() {
	creator := suite.addr("creator-add")
	poolID := suite.createPool(creator)

	// top up provider to add liquidity
	suite.fund(creator, sdk.NewCoins(
		sdk.NewInt64Coin("uaura", 800_000_000),
		sdk.NewInt64Coin("usdt", 800_000_000),
	))

	resp, err := suite.msgServer.AddLiquidity(
		sdk.WrapSDKContext(suite.ctx),
		&dexpb.MsgAddLiquidity{
			Provider: creator,
			PoolId:   poolID,
			AmountA:  sdk.NewInt64Coin("uaura", 500_000_000),
			AmountB:  sdk.NewInt64Coin("usdt", 500_000_000),
		},
	)

	suite.Require().NoError(err)
	suite.NotEmpty(resp.LpTokensMinted)

	pool := suite.keeper.GetPool(suite.ctx, poolID)
	suite.Require().NotNil(pool)
	suite.True(pool.TotalLpTokens.GT(sdkmath.ZeroInt()))
	suite.Len(pool.Providers, 1)
}

func (suite *MsgServerIntegrationSuite) TestSwapExactInHappyPath() {
	creator := suite.addr("creator-swap")
	poolID := suite.createPool(creator)

	// add more liquidity to make swap meaningful
	suite.fund(creator, sdk.NewCoins(
		sdk.NewInt64Coin("uaura", 800_000_000),
		sdk.NewInt64Coin("usdt", 800_000_000),
	))

	_, err := suite.msgServer.AddLiquidity(
		sdk.WrapSDKContext(suite.ctx),
		&dexpb.MsgAddLiquidity{
			Provider: creator,
			PoolId:   poolID,
			AmountA:  sdk.NewInt64Coin("uaura", 400_000_000),
			AmountB:  sdk.NewInt64Coin("usdt", 400_000_000),
		},
	)
	suite.Require().NoError(err)

	// fund swap sender with uaura to swap to usdt
	swapper := suite.addr("swapper")
	suite.fund(swapper, sdk.NewCoins(sdk.NewInt64Coin("uaura", 20_000_000)))

	swapResp, err := suite.msgServer.SwapExactIn(
		sdk.WrapSDKContext(suite.ctx),
		&dexpb.MsgSwapExactIn{
			Sender:         swapper,
			PoolId:         poolID,
			CoinIn:         sdk.NewInt64Coin("uaura", 5_000_000),
			MinAmountOut:   sdkmath.NewInt(1),
			MaxSlippageBps: 10000, // allow up to 100% for test simplicity
		},
	)

	suite.Require().NoError(err)
	suite.True(swapResp.AmountOut.GT(sdkmath.ZeroInt()))
}

func (suite *MsgServerIntegrationSuite) createPool(creator string) string {
	// ensure creator has funds and an account
	suite.fund(creator, sdk.NewCoins(
		sdk.NewInt64Coin("uaura", 3_000_000_000),
		sdk.NewInt64Coin("usdt", 3_000_000_000),
	))

	resp, err := suite.msgServer.CreatePool(
		sdk.WrapSDKContext(suite.ctx),
		&dexpb.MsgCreatePool{
			Creator: creator,
			DenomA:  "uaura",
			DenomB:  "usdt",
			AmountA: sdk.NewInt64Coin("uaura", 1_200_000_000),
			AmountB: sdk.NewInt64Coin("usdt", 1_200_000_000),
		},
	)
	suite.Require().NoError(err)
	return resp.PoolId
}

func (suite *MsgServerIntegrationSuite) fund(address string, coins sdk.Coins) {
	addr, err := sdk.AccAddressFromBech32(address)
	suite.Require().NoError(err)

	suite.accounts.NewAccountWithAddress(suite.ctx, addr)

	current := suite.bank.Balances[address]
	suite.bank.Balances[address] = current.Add(coins...)
}

func (suite *MsgServerIntegrationSuite) addr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	bech32, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	suite.Require().NoError(err)
	return bech32
}
