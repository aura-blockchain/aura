package keeper

import (
	"testing"

	"cosmossdk.io/log"
	runtime "github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/inclusionroutines/params"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

func setupKeeper(t *testing.T) (*Keeper, sdk.Context) {
	input := keepertest.CreateTestInputWithStoreKey(t, types.StoreKey)

	paramsStore := params.NewStore(types.DefaultParams())
	storeService := runtime.NewKVStoreService(input.StoreKey)

	keeper := NewKeeper(
		storeService,
		input.Cdc,
		paramsStore,
		"authority",
		log.NewNopLogger(),
	)

	return keeper, input.Ctx
}

func TestGetIRPrerequisites(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Create IR with prerequisites
	ir := types.IRDefinition{
		Id:          "IR-001",
		Name:        "Test IR",
		Description: "Test Description",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Score:       100,
		PoiReward:   10,
		LocaleTags:  []string{"en"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}
	require.NoError(t, keeper.CreateIR(ctx, ir))

	// Create prerequisite IR
	prereqIR := types.IRDefinition{
		Id:          "IR-000",
		Name:        "Anchor IR",
		Description: "Anchor Description",
		Arena:       inclusionroutinespb.Arena_ARENA_ANCHOR,
		Score:       50,
		PoiReward:   5,
		LocaleTags:  []string{"en"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_LOW,
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}
	require.NoError(t, keeper.CreateIR(ctx, prereqIR))

	// Set prerequisites
	require.NoError(t, keeper.SetPrerequisites(ctx, "IR-001", []string{"IR-000"}))

	// Test GetIRPrerequisites
	prereqs, err := keeper.GetIRPrerequisites(ctx, "IR-001")
	require.NoError(t, err)
	require.Equal(t, []string{"IR-000"}, prereqs)

	// Test with IR that has no prerequisites
	prereqs, err = keeper.GetIRPrerequisites(ctx, "IR-000")
	require.NoError(t, err)
	require.Empty(t, prereqs)

	// Test with non-existent IR
	_, err = keeper.GetIRPrerequisites(ctx, "IR-999")
	require.Error(t, err)
}

func TestIsIRActive(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Create active IR
	activeIR := types.IRDefinition{
		Id:          "IR-001",
		Name:        "Active IR",
		Description: "Test Description",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Score:       100,
		PoiReward:   10,
		LocaleTags:  []string{"en"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}
	require.NoError(t, keeper.CreateIR(ctx, activeIR))

	// Create suspended IR
	suspendedIR := types.IRDefinition{
		Id:          "IR-002",
		Name:        "Suspended IR",
		Description: "Test Description",
		Arena:       inclusionroutinespb.Arena_ARENA_KNOWLEDGE,
		Score:       100,
		PoiReward:   10,
		LocaleTags:  []string{"en"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_MEDIUM,
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_SUSPENDED,
	}
	require.NoError(t, keeper.CreateIR(ctx, suspendedIR))

	// Test IsIRActive
	require.True(t, keeper.IsIRActive(ctx, "IR-001"))
	require.False(t, keeper.IsIRActive(ctx, "IR-002"))
	require.False(t, keeper.IsIRActive(ctx, "IR-999")) // Non-existent IR
}

func TestGetIRScore(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Create IR with score
	ir := types.IRDefinition{
		Id:          "IR-001",
		Name:        "Test IR",
		Description: "Test Description",
		Arena:       inclusionroutinespb.Arena_ARENA_POSSESSION,
		Score:       150,
		PoiReward:   10,
		LocaleTags:  []string{"en"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_MEDIUM,
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}
	require.NoError(t, keeper.CreateIR(ctx, ir))

	// Test GetIRScore
	score, err := keeper.GetIRScore(ctx, "IR-001")
	require.NoError(t, err)
	require.Equal(t, uint64(150), score)

	// Test with non-existent IR
	_, err = keeper.GetIRScore(ctx, "IR-999")
	require.Error(t, err)
}

func TestGetIRArena(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Create IR with arena
	ir := types.IRDefinition{
		Id:          "IR-001",
		Name:        "Test IR",
		Description: "Test Description",
		Arena:       inclusionroutinespb.Arena_ARENA_SOCIAL,
		Score:       100,
		PoiReward:   10,
		LocaleTags:  []string{"en"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_LOW,
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}
	require.NoError(t, keeper.CreateIR(ctx, ir))

	// Test GetIRArena
	arena, err := keeper.GetIRArena(ctx, "IR-001")
	require.NoError(t, err)
	require.Equal(t, "ARENA_SOCIAL", arena)

	// Test with non-existent IR
	_, err = keeper.GetIRArena(ctx, "IR-999")
	require.Error(t, err)
}
