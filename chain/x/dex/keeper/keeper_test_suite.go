package keeper

import (
	"sync"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
)

var configureSDK sync.Once

// KeeperTestSuite is a base test suite for keeper tests in the keeper package
type KeeperTestSuite struct {
	suite.Suite

	Keeper   *Keeper
	SdkCtx   sdk.Context
	Cdc      codec.BinaryCodec
	StoreKey storetypes.StoreKey
}

// SetupTest initializes the test suite before each test
func (suite *KeeperTestSuite) SetupTest() {
	configureSDK.Do(func() {
		cfg := sdk.GetConfig()
		cfg.SetBech32PrefixForAccount("aura", "aurapub")
		cfg.SetBech32PrefixForValidator("auravaloper", "auravaloperpub")
		cfg.SetBech32PrefixForConsensusNode("auravalcons", "auravalconspub")
		cfg.Seal()
	})

	input := keepertest.CreateTestInput(suite.T())

	suite.Keeper = NewKeeper(
		input.Cdc,
		input.StoreKey,
		nil, // bankKeeper
		nil, // accountKeeper
		nil, // vcKeeper
	)
	suite.SdkCtx = input.Ctx
	suite.Cdc = input.Cdc
	suite.StoreKey = input.StoreKey
}
