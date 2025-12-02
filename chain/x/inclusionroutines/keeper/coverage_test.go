package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

func TestIRCRUDAndListing(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	ir := types.IRDefinition{
		Id:          "ir-crud-1",
		Name:        "Test IR",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Description: "basic IR",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Version:     "1.0",
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}

	require.NoError(t, keeper.CreateIR(ctx, ir))

	// Update name and ensure persistence
	ir.Name = "Updated IR"
	require.NoError(t, keeper.UpdateIR(ctx, ir))

	fetched, ok := keeper.GetIR(ctx, ir.Id)
	require.True(t, ok)
	require.Equal(t, "Updated IR", fetched.Name)

	// List active IRs and ensure our IR is present
	list, total := keeper.ListIRs(ctx, inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE, inclusionroutinespb.Arena_ARENA_UNSPECIFIED, "", 0, 10)
	require.GreaterOrEqual(t, total, 1)
	found := false
	for _, item := range list {
		if item.Id == ir.Id {
			found = true
			break
		}
	}
	require.True(t, found, "created IR should appear in listing")

	require.NoError(t, keeper.DeleteIR(ctx, ir.Id))
	_, ok = keeper.GetIR(ctx, ir.Id)
	require.False(t, ok)
}

func TestPrerequisitesAndCircularDetection(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	base := types.IRDefinition{
		Id:          "ir-base",
		Name:        "Base",
		Arena:       inclusionroutinespb.Arena_ARENA_ANCHOR,
		Version:     "1.0",
		Description: "base flow",
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
	}
	child := types.IRDefinition{
		Id:          "ir-child",
		Name:        "Child",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Version:     "1.0",
		Description: "child flow",
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_MEDIUM,
		Score:       50,
	}
	other := types.IRDefinition{
		Id:          "ir-other",
		Name:        "Other",
		Arena:       inclusionroutinespb.Arena_ARENA_KNOWLEDGE,
		Version:     "1.0",
		Description: "other flow",
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_LOW,
		Score:       30,
	}

	require.NoError(t, keeper.CreateIR(ctx, base))
	require.NoError(t, keeper.CreateIR(ctx, child))
	require.NoError(t, keeper.CreateIR(ctx, other))

	require.NoError(t, keeper.SetPrerequisites(ctx, child.Id, []string{base.Id}))

	prereq, ok := keeper.GetPrerequisites(ctx, child.Id)
	require.True(t, ok)
	require.Equal(t, []string{base.Id}, prereq.RequiredIrIds)

	// Create a cycle: other -> child, then base -> other should fail
	require.NoError(t, keeper.SetPrerequisites(ctx, other.Id, []string{child.Id}))
	err := keeper.SetPrerequisites(ctx, base.Id, []string{other.Id})
	require.ErrorIs(t, err, types.ErrCircularDependency)
}

func TestRateLimitConfigurationAndUsage(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	ir := types.IRDefinition{
		Id:          "ir-rate",
		Name:        "Rate IR",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Version:     "1.0",
		Description: "rate limited flow",
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_MEDIUM,
		Score:       10,
	}
	require.NoError(t, keeper.CreateIR(ctx, ir))

	limit := types.IRRateLimit{
		IrId:             ir.Id,
		PerWalletPerHour: 1,
		PerWalletPerDay:  2,
		PerBlockGlobal:   0,
	}
	require.NoError(t, keeper.SetRateLimitConfig(ctx, limit))

	wallet := "wallet1"
	require.NoError(t, keeper.CheckRateLimit(ctx, wallet, ir.Id))
	require.NoError(t, keeper.IncrementRateLimitCounters(ctx, wallet, ir.Id))

	// Next hourly attempt should hit limit
	err := keeper.CheckRateLimit(ctx, wallet, ir.Id)
	require.ErrorIs(t, err, types.ErrRateLimitExceeded)

	// Move context time forward one hour to reset hourly counter
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour))
	require.NoError(t, keeper.CheckRateLimit(ctx, wallet, ir.Id))
}
