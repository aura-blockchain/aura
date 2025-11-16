package keeper

import (
	"testing"

	"github.com/aequitas/aura/chain/x/inclusionroutines/params"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
	"github.com/stretchr/testify/require"
)

func TestGetIRPrerequisites(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	k := NewKeeper(store)

	// Create IR with prerequisites
	ir := types.IRDefinition{
		ID:          "IR-001",
		Name:        "Test IR",
		Description: "Test Description",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Score:       100,
		POIReward:   10,
		LocaleTags:  []string{"en"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}
	require.NoError(t, k.CreateIR(ir))

	// Create prerequisite IR
	prereqIR := types.IRDefinition{
		ID:          "IR-000",
		Name:        "Anchor IR",
		Description: "Anchor Description",
		Arena:       inclusionroutinespb.Arena_ARENA_ANCHOR,
		Score:       50,
		POIReward:   5,
		LocaleTags:  []string{"en"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_LOW,
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}
	require.NoError(t, k.CreateIR(prereqIR))

	// Set prerequisites
	require.NoError(t, k.SetPrerequisites("IR-001", []string{"IR-000"}))

	// Test GetIRPrerequisites
	prereqs, err := k.GetIRPrerequisites("IR-001")
	require.NoError(t, err)
	require.Equal(t, []string{"IR-000"}, prereqs)

	// Test with IR that has no prerequisites
	prereqs, err = k.GetIRPrerequisites("IR-000")
	require.NoError(t, err)
	require.Empty(t, prereqs)

	// Test with non-existent IR
	_, err = k.GetIRPrerequisites("IR-999")
	require.Error(t, err)
}

func TestIsIRActive(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	k := NewKeeper(store)

	// Create active IR
	activeIR := types.IRDefinition{
		ID:          "IR-001",
		Name:        "Active IR",
		Description: "Test Description",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Score:       100,
		POIReward:   10,
		LocaleTags:  []string{"en"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}
	require.NoError(t, k.CreateIR(activeIR))

	// Create suspended IR
	suspendedIR := types.IRDefinition{
		ID:          "IR-002",
		Name:        "Suspended IR",
		Description: "Test Description",
		Arena:       inclusionroutinespb.Arena_ARENA_KNOWLEDGE,
		Score:       100,
		POIReward:   10,
		LocaleTags:  []string{"en"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_MEDIUM,
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_SUSPENDED,
	}
	require.NoError(t, k.CreateIR(suspendedIR))

	// Test IsIRActive
	require.True(t, k.IsIRActive("IR-001"))
	require.False(t, k.IsIRActive("IR-002"))
	require.False(t, k.IsIRActive("IR-999")) // Non-existent IR
}

func TestGetIRScore(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	k := NewKeeper(store)

	// Create IR with score
	ir := types.IRDefinition{
		ID:          "IR-001",
		Name:        "Test IR",
		Description: "Test Description",
		Arena:       inclusionroutinespb.Arena_ARENA_POSSESSION,
		Score:       150,
		POIReward:   10,
		LocaleTags:  []string{"en"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_MEDIUM,
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}
	require.NoError(t, k.CreateIR(ir))

	// Test GetIRScore
	score, err := k.GetIRScore("IR-001")
	require.NoError(t, err)
	require.Equal(t, uint64(150), score)

	// Test with non-existent IR
	_, err = k.GetIRScore("IR-999")
	require.Error(t, err)
}

func TestGetIRArena(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	k := NewKeeper(store)

	// Create IR with arena
	ir := types.IRDefinition{
		ID:          "IR-001",
		Name:        "Test IR",
		Description: "Test Description",
		Arena:       inclusionroutinespb.Arena_ARENA_SOCIAL,
		Score:       100,
		POIReward:   10,
		LocaleTags:  []string{"en"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_LOW,
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}
	require.NoError(t, k.CreateIR(ir))

	// Test GetIRArena
	arena, err := k.GetIRArena("IR-001")
	require.NoError(t, err)
	require.Equal(t, "ARENA_SOCIAL", arena)

	// Test with non-existent IR
	_, err = k.GetIRArena("IR-999")
	require.Error(t, err)
}
