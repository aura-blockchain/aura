package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

func TestNewKeeper(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	if keeper == nil {
		t.Fatal("expected keeper to be non-nil")
	}

	if keeper.paramsStore == nil {
		t.Error("keeper params store not set correctly")
	}

	// validate params are accessible
	params := keeper.GetParams()
	if params.MaxIrPerLocale == 0 {
		t.Errorf("expected default params to be set, got %+v", params)
	}

	_ = ctx
}

func TestCreateAndGetIR(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	ir := types.IRDefinition{
		Id:          "IR-001",
		Name:        "Test IR",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_BIOMETRIC),
		Description: "Test description",
		Score:       100,
		PoiReward:   50,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
		Status:      types.IRStatus(inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE),
	}

	err := keeper.CreateIR(ctx, ir)
	if err != nil {
		t.Fatalf("failed to create IR: %v", err)
	}

	retrieved, ok := keeper.GetIR(ctx, "IR-001")
	if !ok {
		t.Fatal("IR not found after creation")
	}

	if retrieved.Name != "Test IR" {
		t.Errorf("expected name 'Test IR', got '%s'", retrieved.Name)
	}
}

func TestDeleteIR(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	ir := types.IRDefinition{
		Id:          "IR-002",
		Name:        "Test IR 2",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_KNOWLEDGE),
		Description: "Test description",
		Score:       100,
		PoiReward:   50,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_LOW),
		Version:     "1.0",
		Status:      types.IRStatus(inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE),
	}

	keeper.CreateIR(ctx, ir)

	err := keeper.DeleteIR(ctx, "IR-002")
	if err != nil {
		t.Fatalf("failed to delete IR: %v", err)
	}

	_, ok := keeper.GetIR(ctx, "IR-002")
	if ok {
		t.Error("IR still exists after deletion")
	}
}

func TestSetPrerequisites(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	// Create two IRs
	ir1 := types.IRDefinition{
		Id:          "IR-100",
		Name:        "Base IR",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_ANCHOR),
		Description: "Base",
		Score:       0,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
	}
	ir2 := types.IRDefinition{
		Id:          "IR-101",
		Name:        "Dependent IR",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_BIOMETRIC),
		Description: "Dependent",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
	}

	keeper.CreateIR(ctx, ir1)
	keeper.CreateIR(ctx, ir2)

	// Set prerequisite
	err := keeper.SetPrerequisites(ctx, "IR-101", []string{"IR-100"})
	if err != nil {
		t.Fatalf("failed to set prerequisites: %v", err)
	}

	prereq, ok := keeper.GetPrerequisites(ctx, "IR-101")
	if !ok {
		t.Fatal("prerequisites not found")
	}

	if len(prereq.RequiredIrIds) != 1 || prereq.RequiredIrIds[0] != "IR-100" {
		t.Error("prerequisite not set correctly")
	}
}

func TestCircularDependency(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	// Create three IRs
	ir1 := types.IRDefinition{
		Id:          "IR-200",
		Name:        "IR 1",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_BIOMETRIC),
		Description: "IR 1",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
	}
	ir2 := types.IRDefinition{
		Id:          "IR-201",
		Name:        "IR 2",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_KNOWLEDGE),
		Description: "IR 2",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
	}
	ir3 := types.IRDefinition{
		Id:          "IR-202",
		Name:        "IR 3",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_SOCIAL),
		Description: "IR 3",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
	}

	keeper.CreateIR(ctx, ir1)
	keeper.CreateIR(ctx, ir2)
	keeper.CreateIR(ctx, ir3)

	// IR-201 requires IR-200
	keeper.SetPrerequisites(ctx, "IR-201", []string{"IR-200"})

	// IR-202 requires IR-201
	keeper.SetPrerequisites(ctx, "IR-202", []string{"IR-201"})

	// Try to make IR-200 require IR-202 (creates cycle)
	err := keeper.SetPrerequisites(ctx, "IR-200", []string{"IR-202"})
	if err != types.ErrCircularDependency {
		t.Errorf("expected circular dependency error, got %v", err)
	}
}

func TestRateLimit(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	ir := types.IRDefinition{
		Id:          "IR-300",
		Name:        "Rate Limited IR",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_BIOMETRIC),
		Description: "Rate limited",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
	}

	keeper.CreateIR(ctx, ir)

	// Set rate limit
	limit := types.IRRateLimit{
		IrId:             "IR-300",
		PerWalletPerHour: 2,
		PerWalletPerDay:  10,
		PerBlockGlobal:   100,
	}

	err := keeper.SetRateLimit(ctx, limit)
	if err != nil {
		t.Fatalf("failed to set rate limit: %v", err)
	}

	wallet := "wallet123"

	// First check should pass
	err = keeper.CheckRateLimit(ctx, wallet, "IR-300")
	if err != nil {
		t.Errorf("first rate limit check failed: %v", err)
	}

	// Increment usage
	require.NoError(t, keeper.IncrementRateLimitCounters(ctx, wallet, "IR-300"))
	require.NoError(t, keeper.IncrementRateLimitCounters(ctx, wallet, "IR-300"))

	// Third check should fail (hourly limit is 2)
	err = keeper.CheckRateLimit(ctx, wallet, "IR-300")
	if err == nil {
		t.Error("expected rate limit error after exceeding limit")
	}
}

func TestValidatePrerequisites(t *testing.T) {
	ctx, keeper := setupInclusionKeeper(t)

	// Create base IR
	ir1 := types.IRDefinition{
		Id:          "IR-400",
		Name:        "Base",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_ANCHOR),
		Description: "Base",
		Score:       0,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
	}

	// Create dependent IR
	ir2 := types.IRDefinition{
		Id:          "IR-401",
		Name:        "Dependent",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_BIOMETRIC),
		Description: "Dependent",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
	}

	keeper.CreateIR(ctx, ir1)
	keeper.CreateIR(ctx, ir2)
	keeper.SetPrerequisites(ctx, "IR-401", []string{"IR-400"})

	// Test with prerequisites met
	completed := []string{"IR-400"}
	err := keeper.ValidatePrerequisites(ctx, "IR-401", completed)
	if err != nil {
		t.Errorf("validation failed with completed prerequisites: %v", err)
	}

	// Test with prerequisites not met
	err = keeper.ValidatePrerequisites(ctx, "IR-401", []string{})
	if err == nil {
		t.Error("expected prerequisite error when prerequisites not met")
	}
}
