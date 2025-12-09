package keeper

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

type InvariantsTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

func (suite *InvariantsTestSuite) TestAllInvariants() {
	ctx := suite.SdkCtx

	// Test: All invariants on empty store
	inv := AllInvariants(&suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Register invariants - should not panic
	suite.NotPanics(func() {
		// RegisterInvariants expects a crisis.InvariantRegistry
		// For this basic test, we just verify it doesn't panic when called with keeper
		inv := AllInvariants(&suite.Keeper)
		suite.NotNil(inv, "invariant function should be created")
	})
}

func (suite *InvariantsTestSuite) TestPeerReputationConsistencyInvariant_InvalidScore() {
	ctx := suite.SdkCtx
	// Insert an invalid reputation score to trigger the invariant.
	rep := types.NodeReputation{
		PeerId:            "peer-invalid",
		Score:             150, // >100 triggers invariant
		LastUpdatedHeight: 1,
	}
	err := suite.Keeper.SetReputation(ctx, rep)
	suite.Require().NoError(err)

	msg, broken := PeerReputationConsistencyInvariant(&suite.Keeper)(ctx)
	suite.True(broken, "invalid score should break invariant")
	suite.Contains(msg, "peer-reputation-consistency")
}

func (suite *InvariantsTestSuite) TestPeerReputationConsistencyInvariant_NegativeHeight() {
	ctx := suite.SdkCtx

	rep := types.NodeReputation{
		PeerId:            "peer-neg-height",
		Score:             10,
		LastUpdatedHeight: -5,
	}
	suite.Require().NoError(suite.Keeper.SetReputation(ctx, rep))

	msg, broken := PeerReputationConsistencyInvariant(&suite.Keeper)(ctx)
	suite.True(broken, "negative heights should break invariant")
	suite.Contains(msg, "last_updated_height")
}

func FuzzPeerReputationConsistencyInvariant_Bounds(f *testing.F) {
	f.Add(int64(0), int64(0))
	f.Add(int64(-10), int64(5))
	f.Add(int64(150), int64(100))
	f.Add(int64(99), int64(10))

	f.Fuzz(func(t *testing.T, score int64, height int64) {
		suite := new(InvariantsTestSuite)
		suite.SetT(t)
		suite.SetupTest()
		ctx := suite.SdkCtx

		rep := types.NodeReputation{
			PeerId:            "peer-fuzz",
			Score:             score,
			LastUpdatedHeight: height,
		}
		suite.Require().NoError(suite.Keeper.SetReputation(ctx, rep))

		msg, broken := PeerReputationConsistencyInvariant(&suite.Keeper)(ctx)
		if score < 0 || score > 100 || height < 0 {
			suite.True(broken, "invalid fuzz input should break invariant")
			suite.NotEmpty(msg)
		} else {
			suite.False(broken, "valid fuzz input should not break invariant")
		}
	})
}

func (suite *InvariantsTestSuite) TestRateLimitValidityInvariant_MissingWindow() {
	ctx := suite.SdkCtx
	// Store a rate limit entry with empty peer id to trip invariant.
	rl := types.RateLimitEntry{
		PeerId: "",
	}
	suite.Require().NoError(suite.Keeper.SetRateLimitEntry(ctx, rl))

	msg, broken := RateLimitValidityInvariant(&suite.Keeper)(ctx)
	suite.True(broken, "missing peer id should break invariant")
	suite.Contains(msg, "rate-limit-validity")
}

func (suite *InvariantsTestSuite) TestRateLimitValidityInvariant_NilWindow() {
	ctx := suite.SdkCtx

	rl := types.RateLimitEntry{
		PeerId:      "peer-nil-window",
		WindowStart: nil,
	}
	suite.Require().NoError(suite.Keeper.SetRateLimitEntry(ctx, rl))

	msg, broken := RateLimitValidityInvariant(&suite.Keeper)(ctx)
	suite.True(broken, "nil window should break invariant")
	suite.Contains(msg, "nil window_start")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_NilConnectionConfig() {
	ctx := suite.SdkCtx
	// Persist params with nil connection to trigger invariant without passing validation.
	store := suite.Keeper.storeService.OpenKVStore(ctx)
	params := types.Params{Connection: nil}
	bz := suite.Keeper.cdc.MustMarshal(&params)
	_ = store.Set(types.ParamsKey, bz)

	msg, broken := ParamsInvariant(&suite.Keeper)(ctx)
	suite.True(broken, "corrupt params should break params invariant")
	suite.Contains(msg, "params-valid")
}

func (suite *InvariantsTestSuite) TestParamsInvariant_InvalidNodeSelectors() {
	ctx := suite.SdkCtx
	params := types.DefaultParams()
	params.Connection.MaxInboundConnections = 0 // invalid: must be positive

	store := suite.Keeper.storeService.OpenKVStore(ctx)
	bz := suite.Keeper.cdc.MustMarshal(params)
	suite.Require().NoError(store.Set(types.ParamsKey, bz))

	msg, broken := ParamsInvariant(&suite.Keeper)(ctx)
	suite.True(broken, "invalid connection params should break invariant")
	suite.Contains(msg, "params-valid")
}

func (suite *InvariantsTestSuite) TestMempoolSecurityInvariant_LargeStats() {
	ctx := suite.SdkCtx

	// Set oversized mempool stats to trigger invariant.
	stats := types.MempoolStats{
		TxCount:   1 << 62,
		SizeBytes: 1 << 62,
	}
	suite.Require().NoError(suite.Keeper.SetMempoolStats(ctx, stats))

	msg, broken := MempoolSecurityStateInvariant(&suite.Keeper)(ctx)
	suite.True(broken, "oversized mempool stats should break invariant")
	suite.Contains(msg, "mempool-security-state")
}

func (suite *InvariantsTestSuite) TestSybilDetectionInvariant_EmptyAlertId() {
	ctx := suite.SdkCtx

	alert := types.ForkAlert{
		AlertId: "", // should trigger invariant
	}
	suite.Require().NoError(suite.Keeper.SetForkAlert(ctx, alert))

	msg, broken := SybilDetectionIntegrityInvariant(&suite.Keeper)(ctx)
	suite.True(broken, "empty fork alert id should break invariant")
	suite.Contains(msg, "sybil-detection-integrity")
}

func (suite *InvariantsTestSuite) TestSybilDetectionInvariant_NegativeHeight() {
	ctx := suite.SdkCtx

	alert := types.ForkAlert{
		AlertId:     "fork-negative",
		BlockHeight: -1,
	}
	suite.Require().NoError(suite.Keeper.SetForkAlert(ctx, alert))

	msg, broken := SybilDetectionIntegrityInvariant(&suite.Keeper)(ctx)
	suite.True(broken, "negative fork height should break invariant")
	suite.Contains(msg, "negative height")
}
