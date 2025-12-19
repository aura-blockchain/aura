package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

func TestMsgServerConstruction(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)
	server := NewMsgServer(keeper)

	require.NotNil(t, server)
	require.NotNil(t, ctx.Context())

	// Ensure interface compliance at runtime
	_, ok := server.(interface {
		inclusionroutinespb.MsgServer
	})
	require.True(t, ok)
	_ = context.Background()
}

func TestMsgServerCreateIR(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)
	ctx = ctx.WithEventManager(sdk.NewEventManager()).WithBlockHeight(42)
	server := NewMsgServer(keeper)

	msg := &inclusionroutinespb.MsgCreateIR{
		Authority:        "authority",
		Id:               "ir-create-1",
		Name:             "Biometric Liveness",
		Arena:            inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Description:      "Verify biometric liveness signal",
		Score:            120,
		PoiReward:        50,
		LocaleTags:       []string{"global"},
		PrivacyTier:      inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Version:          "1.0.0",
		MetadataHash:     "hash-1",
		ActivationHeight: 10,
		SunsetHeight:     1000,
	}

	resp, err := server.CreateIR(sdk.WrapSDKContext(ctx), msg)
	require.NoError(t, err)
	require.Equal(t, "ir-create-1", resp.Id)

	stored, ok := keeper.GetIR(ctx, "ir-create-1")
	require.True(t, ok)
	require.Equal(t, "Biometric Liveness", stored.Name)
	require.Equal(t, inclusionroutinespb.Arena_ARENA_BIOMETRIC, stored.Arena)
	require.Equal(t, inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH, stored.PrivacyTier)
	require.Equal(t, int64(120), stored.Score)
	require.Equal(t, int64(50), stored.PoiReward)
	require.Equal(t, []string{"global"}, stored.LocaleTags)

	events := ctx.EventManager().Events()
	require.NotEmpty(t, events)
	require.Equal(t, types.EventTypeIRCreated, events[len(events)-1].Type)
}

func TestMsgServerCreateIR_Unauthorized(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	server := NewMsgServer(keeper)

	_, err := server.CreateIR(sdk.WrapSDKContext(ctx), &inclusionroutinespb.MsgCreateIR{
		Authority:        "not-authority",
		Id:               "ir-create-unauth",
		Name:             "Invalid",
		Arena:            inclusionroutinespb.Arena_ARENA_ANCHOR,
		Description:      "test",
		Score:            10,
		PoiReward:        5,
		LocaleTags:       []string{"global"},
		PrivacyTier:      inclusionroutinespb.PrivacyTier_PRIVACY_TIER_LOW,
		Version:          "1.0.0",
		MetadataHash:     "hash",
		ActivationHeight: 1,
	})
	require.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestMsgServerUpdateIRPartial(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	base := types.IRDefinition{
		Id:               "ir-update-1",
		Name:             "Original Name",
		Arena:            inclusionroutinespb.Arena_ARENA_KNOWLEDGE,
		Description:      "Original description",
		Score:            200,
		PoiReward:        90,
		LocaleTags:       []string{"us"},
		PrivacyTier:      inclusionroutinespb.PrivacyTier_PRIVACY_TIER_MEDIUM,
		Version:          "1.0.0",
		MetadataHash:     "hash-old",
		Status:           inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
		ActivationHeight: 5,
		SunsetHeight:     2000,
	}
	require.NoError(t, keeper.CreateIR(ctx, base))

	server := NewMsgServer(keeper)
	updateMsg := &inclusionroutinespb.MsgUpdateIR{
		Authority:    "authority",
		Id:           "ir-update-1",
		Name:         "Updated Name",
		Description:  "Refined description",
		PoiReward:    120,
		LocaleTags:   []string{"us", "ca"},
		Version:      "1.1.0",
		MetadataHash: "hash-new",
	}

	_, err := server.UpdateIR(sdk.WrapSDKContext(ctx), updateMsg)
	require.NoError(t, err)

	updated, ok := keeper.GetIR(ctx, "ir-update-1")
	require.True(t, ok)
	require.Equal(t, "Updated Name", updated.Name)
	require.Equal(t, "Refined description", updated.Description)
	require.Equal(t, int64(200), updated.Score, "score should remain unchanged without explicit override")
	require.Equal(t, int64(120), updated.PoiReward)
	require.Equal(t, []string{"us", "ca"}, updated.LocaleTags)
	require.Equal(t, "1.1.0", updated.Version)
	require.Equal(t, "hash-new", updated.MetadataHash)
}

func TestMsgServerSetPrerequisitesAndRateLimit(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	server := NewMsgServer(keeper)

	require.NoError(t, keeper.CreateIR(ctx, types.IRDefinition{
		Id:               "ir-parent",
		Name:             "Parent",
		Arena:            inclusionroutinespb.Arena_ARENA_ANCHOR,
		Description:      "Parent IR",
		Score:            10,
		PoiReward:        5,
		LocaleTags:       []string{"global"},
		PrivacyTier:      inclusionroutinespb.PrivacyTier_PRIVACY_TIER_LOW,
		ActivationHeight: 1,
		SunsetHeight:     100,
	}))
	require.NoError(t, keeper.CreateIR(ctx, types.IRDefinition{
		Id:               "ir-child",
		Name:             "Child",
		Arena:            inclusionroutinespb.Arena_ARENA_SOCIAL,
		Description:      "Child IR",
		Score:            15,
		PoiReward:        6,
		LocaleTags:       []string{"global"},
		PrivacyTier:      inclusionroutinespb.PrivacyTier_PRIVACY_TIER_MEDIUM,
		ActivationHeight: 1,
		SunsetHeight:     200,
	}))

	_, err := server.SetIRPrerequisites(sdk.WrapSDKContext(ctx), &inclusionroutinespb.MsgSetIRPrerequisites{
		Authority:     "authority",
		IrId:          "ir-child",
		RequiredIrIds: []string{"ir-parent"},
	})
	require.NoError(t, err)

	prereq, ok := keeper.GetPrerequisite(ctx, "ir-child")
	require.True(t, ok)
	require.Equal(t, []string{"ir-parent"}, prereq.RequiredIrIds)

	_, err = server.SetIRRateLimit(sdk.WrapSDKContext(ctx), &inclusionroutinespb.MsgSetIRRateLimit{
		Authority:        "authority",
		IrId:             "ir-child",
		PerWalletPerHour: 5,
		PerWalletPerDay:  10,
		PerBlockGlobal:   50,
	})
	require.NoError(t, err)

	limit, ok := keeper.GetRateLimitConfig(ctx, "ir-child")
	require.True(t, ok)
	require.Equal(t, int32(5), limit.PerWalletPerHour)
	require.Equal(t, int32(10), limit.PerWalletPerDay)
	require.Equal(t, int32(50), limit.PerBlockGlobal)
}

func TestMsgServerSuspendActivateDelete(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)
	ctx = ctx.WithEventManager(sdk.NewEventManager()).WithBlockHeight(5)
	server := NewMsgServer(keeper)

	require.NoError(t, keeper.CreateIR(ctx, types.IRDefinition{
		Id:               "ir-lifecycle",
		Name:             "Lifecycle IR",
		Arena:            inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Description:      "Lifecycle test",
		Score:            30,
		PoiReward:        12,
		LocaleTags:       []string{"global"},
		PrivacyTier:      inclusionroutinespb.PrivacyTier_PRIVACY_TIER_MEDIUM,
		ActivationHeight: 1,
		SunsetHeight:     500,
	}))

	_, err := server.ActivateIR(sdk.WrapSDKContext(ctx), &inclusionroutinespb.MsgActivateIR{
		Authority: "authority",
		IrId:      "ir-lifecycle",
	})
	require.NoError(t, err)

	ir, ok := keeper.GetIR(ctx, "ir-lifecycle")
	require.True(t, ok)
	require.Equal(t, inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE, ir.Status)

	_, err = server.SuspendIR(sdk.WrapSDKContext(ctx), &inclusionroutinespb.MsgSuspendIR{
		Authority: "authority",
		IrId:      "ir-lifecycle",
		Reason:    "maintenance",
	})
	require.NoError(t, err)

	ir, ok = keeper.GetIR(ctx, "ir-lifecycle")
	require.True(t, ok)
	require.Equal(t, inclusionroutinespb.IRStatus_IR_STATUS_SUSPENDED, ir.Status)

	_, err = server.DeleteIR(sdk.WrapSDKContext(ctx), &inclusionroutinespb.MsgDeleteIR{
		Authority: "authority",
		Id:        "ir-lifecycle",
	})
	require.NoError(t, err)

	_, exists := keeper.GetIR(ctx, "ir-lifecycle")
	require.False(t, exists)
}
