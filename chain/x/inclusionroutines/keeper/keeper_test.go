package keeper

import (
	"testing"

	"github.com/aequitas/aura/chain/x/inclusionroutines/params"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

func TestNewKeeper(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)

	if keeper == nil {
		t.Fatal("expected keeper to be non-nil")
	}

	if keeper.paramsStore != store {
		t.Error("keeper params store not set correctly")
	}
}

func TestCreateAndGetIR(t *testing.T) {
	keeper := NewKeeper(nil)

	ir := types.IRDefinition{
		ID:          "IR-001",
		Name:        "Test IR",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Description: "Test description",
		Score:       100,
		POIReward:   50,
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Version:     "1.0",
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}

	err := keeper.CreateIR(ir)
	if err != nil {
		t.Fatalf("failed to create IR: %v", err)
	}

	retrieved, ok := keeper.GetIR("IR-001")
	if !ok {
		t.Fatal("IR not found after creation")
	}

	if retrieved.Name != "Test IR" {
		t.Errorf("expected name 'Test IR', got '%s'", retrieved.Name)
	}
}

func TestDeleteIR(t *testing.T) {
	keeper := NewKeeper(nil)

	ir := types.IRDefinition{
		ID:          "IR-002",
		Name:        "Test IR 2",
		Arena:       inclusionroutinespb.Arena_ARENA_KNOWLEDGE,
		Description: "Test description",
		Score:       100,
		POIReward:   50,
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_LOW,
		Version:     "1.0",
		Status:      inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE,
	}

	keeper.CreateIR(ir)

	err := keeper.DeleteIR("IR-002")
	if err != nil {
		t.Fatalf("failed to delete IR: %v", err)
	}

	_, ok := keeper.GetIR("IR-002")
	if ok {
		t.Error("IR still exists after deletion")
	}
}

func TestSetPrerequisites(t *testing.T) {
	keeper := NewKeeper(nil)

	// Create two IRs
	ir1 := types.IRDefinition{
		ID:          "IR-100",
		Name:        "Base IR",
		Arena:       inclusionroutinespb.Arena_ARENA_ANCHOR,
		Description: "Base",
		Score:       0,
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Version:     "1.0",
	}
	ir2 := types.IRDefinition{
		ID:          "IR-101",
		Name:        "Dependent IR",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Description: "Dependent",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Version:     "1.0",
	}

	keeper.CreateIR(ir1)
	keeper.CreateIR(ir2)

	// Set prerequisite
	err := keeper.SetPrerequisites("IR-101", []string{"IR-100"})
	if err != nil {
		t.Fatalf("failed to set prerequisites: %v", err)
	}

	prereq, ok := keeper.GetPrerequisites("IR-101")
	if !ok {
		t.Fatal("prerequisites not found")
	}

	if len(prereq.RequiredIRIDs) != 1 || prereq.RequiredIRIDs[0] != "IR-100" {
		t.Error("prerequisite not set correctly")
	}
}

func TestCircularDependency(t *testing.T) {
	keeper := NewKeeper(nil)

	// Create three IRs
	ir1 := types.IRDefinition{
		ID:          "IR-200",
		Name:        "IR 1",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Description: "IR 1",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Version:     "1.0",
	}
	ir2 := types.IRDefinition{
		ID:          "IR-201",
		Name:        "IR 2",
		Arena:       inclusionroutinespb.Arena_ARENA_KNOWLEDGE,
		Description: "IR 2",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Version:     "1.0",
	}
	ir3 := types.IRDefinition{
		ID:          "IR-202",
		Name:        "IR 3",
		Arena:       inclusionroutinespb.Arena_ARENA_SOCIAL,
		Description: "IR 3",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Version:     "1.0",
	}

	keeper.CreateIR(ir1)
	keeper.CreateIR(ir2)
	keeper.CreateIR(ir3)

	// IR-201 requires IR-200
	keeper.SetPrerequisites("IR-201", []string{"IR-200"})

	// IR-202 requires IR-201
	keeper.SetPrerequisites("IR-202", []string{"IR-201"})

	// Try to make IR-200 require IR-202 (creates cycle)
	err := keeper.SetPrerequisites("IR-200", []string{"IR-202"})
	if err != types.ErrCircularDependency {
		t.Errorf("expected circular dependency error, got %v", err)
	}
}

func TestRateLimit(t *testing.T) {
	keeper := NewKeeper(nil)

	ir := types.IRDefinition{
		ID:          "IR-300",
		Name:        "Rate Limited IR",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Description: "Rate limited",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Version:     "1.0",
	}

	keeper.CreateIR(ir)

	// Set rate limit
	limit := types.IRRateLimit{
		IRID:             "IR-300",
		PerWalletPerHour: 2,
		PerWalletPerDay:  10,
		PerBlockGlobal:   100,
	}

	err := keeper.SetRateLimit(limit)
	if err != nil {
		t.Fatalf("failed to set rate limit: %v", err)
	}

	wallet := "wallet123"

	// First check should pass
	err = keeper.CheckRateLimit(wallet, "IR-300")
	if err != nil {
		t.Errorf("first rate limit check failed: %v", err)
	}

	// Increment usage
	keeper.IncrementRateLimit(wallet, "IR-300")
	keeper.IncrementRateLimit(wallet, "IR-300")

	// Third check should fail (hourly limit is 2)
	err = keeper.CheckRateLimit(wallet, "IR-300")
	if err == nil {
		t.Error("expected rate limit error after exceeding limit")
	}
}

func TestValidatePrerequisites(t *testing.T) {
	keeper := NewKeeper(nil)

	// Create base IR
	ir1 := types.IRDefinition{
		ID:          "IR-400",
		Name:        "Base",
		Arena:       inclusionroutinespb.Arena_ARENA_ANCHOR,
		Description: "Base",
		Score:       0,
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Version:     "1.0",
	}

	// Create dependent IR
	ir2 := types.IRDefinition{
		ID:          "IR-401",
		Name:        "Dependent",
		Arena:       inclusionroutinespb.Arena_ARENA_BIOMETRIC,
		Description: "Dependent",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH,
		Version:     "1.0",
	}

	keeper.CreateIR(ir1)
	keeper.CreateIR(ir2)
	keeper.SetPrerequisites("IR-401", []string{"IR-400"})

	// Test with prerequisites met
	completed := []string{"IR-400"}
	err := keeper.ValidatePrerequisites("IR-401", completed)
	if err != nil {
		t.Errorf("validation failed with completed prerequisites: %v", err)
	}

	// Test with prerequisites not met
	err = keeper.ValidatePrerequisites("IR-401", []string{})
	if err == nil {
		t.Error("expected prerequisite error when prerequisites not met")
	}
}
