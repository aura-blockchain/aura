package keeper

import (
	"sync"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
)

var configureComplianceSDK sync.Once

// KeeperTestSuite provides a shared harness for compliance keeper tests.
type KeeperTestSuite struct {
	suite.Suite

	Keeper   *Keeper
	SdkCtx   sdk.Context
	Cdc      codec.Codec
	StoreKey storetypes.StoreKey
}

func (suite *KeeperTestSuite) SetupTest() {
	configureComplianceSDK.Do(func() {
		cfg := sdk.GetConfig()
		cfg.SetBech32PrefixForAccount("aura", "aurapub")
		cfg.SetBech32PrefixForValidator("auravaloper", "auravaloperpub")
		cfg.SetBech32PrefixForConsensusNode("auravalcons", "auravalconspub")
		cfg.Seal()
	})

	input := keepertest.CreateTestInputWithKeys(suite.T(), "compliance")

	suite.Keeper = NewKeeper(input.Cdc, input.StoreKey)
	suite.SdkCtx = input.Ctx
	suite.Cdc = input.Cdc
	suite.StoreKey = input.StoreKey
}
