package keeper

import (
	"testing"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

func TestListIRs(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	// Create multiple IRs
	for i := 0; i < 5; i++ {
		ir := types.IRDefinition{
			Id:          "list-" + string(rune('a'+i)),
			Name:        "IR " + string(rune('a'+i)),
			Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_BIOMETRIC),
			Description: "Test IR description",
			Score:       100,
			LocaleTags:  []string{"global"},
			PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
			Version:     "1.0",
			Status:      types.IRStatus(inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE),
		}
		err := keeper.CreateIR(ir)
		if err != nil {
			t.Fatalf("failed to create IR: %v", err)
		}
	}

	irs, _ := keeper.ListIRs(inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE, inclusionroutinespb.Arena_ARENA_UNSPECIFIED, "", 0, 100)
	if len(irs) < 5 {
		t.Errorf("expected at least 5 IRs, got %d", len(irs))
	}
}

func TestUpdateIR(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	ir := types.IRDefinition{
		Id:          "update-ir",
		Name:        "Original Name",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_BIOMETRIC),
		Description: "Original description",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
	}

	err := keeper.CreateIR(ir)
	if err != nil {
		t.Fatalf("failed to create IR: %v", err)
	}

	// Update
	ir.Name = "Updated Name"
	err = keeper.UpdateIR(ir)
	if err != nil {
		t.Fatalf("failed to update IR: %v", err)
	}

	retrieved, found := keeper.GetIR("update-ir")
	if !found {
		t.Fatal("IR not found after update")
	}
	if retrieved.Name != "Updated Name" {
		t.Errorf("expected 'Updated Name', got '%s'", retrieved.Name)
	}
}

func TestUpdateIRNotFound(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	ir := types.IRDefinition{
		Id: "nonexistent",
	}

	err := keeper.UpdateIR(ir)
	if err == nil {
		t.Error("expected error when updating nonexistent IR")
	}
}

func TestSuspendIR(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	ir := types.IRDefinition{
		Id:          "suspend-ir",
		Name:        "Suspend Test",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_BIOMETRIC),
		Description: "Suspend test description",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
		Status:      types.IRStatus(inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE),
	}

	err := keeper.CreateIR(ir)
	if err != nil {
		t.Fatalf("failed to create IR: %v", err)
	}

	err = keeper.SuspendIR("suspend-ir")
	if err != nil {
		t.Fatalf("failed to suspend IR: %v", err)
	}

	retrieved, found := keeper.GetIR("suspend-ir")
	if !found {
		t.Fatal("IR not found after suspend")
	}
	if retrieved.Status != types.IRStatus(inclusionroutinespb.IRStatus_IR_STATUS_SUSPENDED) {
		t.Errorf("expected suspended status, got %v", retrieved.Status)
	}
}

func TestActivateIR(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	ir := types.IRDefinition{
		Id:          "activate-ir",
		Name:        "Activate Test",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_BIOMETRIC),
		Description: "Activate test description",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
		Status:      types.IRStatus(inclusionroutinespb.IRStatus_IR_STATUS_SUSPENDED),
	}

	err := keeper.CreateIR(ir)
	if err != nil {
		t.Fatalf("failed to create IR: %v", err)
	}

	err = keeper.ActivateIR("activate-ir")
	if err != nil {
		t.Fatalf("failed to activate IR: %v", err)
	}

	retrieved, found := keeper.GetIR("activate-ir")
	if !found {
		t.Fatal("IR not found after activate")
	}
	if retrieved.Status != types.IRStatus(inclusionroutinespb.IRStatus_IR_STATUS_ACTIVE) {
		t.Errorf("expected active status, got %v", retrieved.Status)
	}
}

func TestSuspendIRNotFound(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	err := keeper.SuspendIR("nonexistent")
	if err == nil {
		t.Error("expected error when suspending nonexistent IR")
	}
}

func TestActivateIRNotFound(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	err := keeper.ActivateIR("nonexistent")
	if err == nil {
		t.Error("expected error when activating nonexistent IR")
	}
}

func TestGetIRNotFound(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	_, found := keeper.GetIR("nonexistent")
	if found {
		t.Error("expected IR not to be found")
	}
}

func TestDeleteIRNotFound(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	err := keeper.DeleteIR("nonexistent")
	if err == nil {
		t.Error("expected error when deleting nonexistent IR")
	}
}

func TestSetIRMultiple(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	// Create multiple IRs using SetIR directly
	for i := 0; i < 3; i++ {
		ir := types.IRDefinition{
			Id:          "set-" + string(rune('a'+i)),
			Name:        "Set IR " + string(rune('a'+i)),
			Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_KNOWLEDGE),
			Description: "Test IR description",
			Score:       100,
			LocaleTags:  []string{"global"},
			PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_LOW),
			Version:     "1.0",
		}
		err := keeper.SetIR(ir)
		if err != nil {
			t.Fatalf("failed to set IR: %v", err)
		}
	}

	irs, _ := keeper.ListIRs(inclusionroutinespb.IRStatus_IR_STATUS_UNSPECIFIED, inclusionroutinespb.Arena_ARENA_UNSPECIFIED, "", 0, 100)
	if len(irs) < 3 {
		t.Errorf("expected at least 3 IRs, got %d", len(irs))
	}
}

func TestGetAllIRs(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	// Create IRs
	for i := 0; i < 3; i++ {
		ir := types.IRDefinition{
			Id:          "getall-" + string(rune('a'+i)),
			Name:        "GetAll IR " + string(rune('a'+i)),
			Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_ANCHOR),
			Description: "Test IR description",
			Score:       100,
			LocaleTags:  []string{"global"},
			PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
			Version:     "1.0",
		}
		err := keeper.CreateIR(ir)
		if err != nil {
			t.Fatalf("failed to create IR: %v", err)
		}
	}

	irs := keeper.GetAllIRs()
	if len(irs) < 3 {
		t.Errorf("expected at least 3 IRs, got %d", len(irs))
	}
}

func TestGetAuthority(t *testing.T) {
	keeper := NewKeeper(nil, "test-authority")

	authority := keeper.GetAuthority()
	if authority != "test-authority" {
		t.Errorf("expected 'test-authority', got '%s'", authority)
	}
}

func TestGetPrerequisitesNotFound(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	prereqs, found := keeper.GetPrerequisites("nonexistent")
	if found {
		t.Error("expected prerequisites not found for nonexistent IR")
	}
	_ = prereqs // use variable
}

func TestGetIRGraphEmpty(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	graph := keeper.GetIRGraph("")
	if graph == nil {
		t.Error("expected non-nil graph")
	}
}

func TestCheckRateLimitNoLimit(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	// Create IR without rate limit
	ir := types.IRDefinition{
		Id:          "nolimit-ir",
		Name:        "No Limit",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_BIOMETRIC),
		Description: "No limit test",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
	}
	err := keeper.CreateIR(ir)
	if err != nil {
		t.Fatalf("failed to create IR: %v", err)
	}

	// Check rate limit - should allow (no error means allowed)
	err = keeper.CheckRateLimit("wallet1", "nolimit-ir")
	if err != nil {
		t.Error("expected rate limit check to pass when no limit set")
	}
}

func TestSetRateLimit(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	ir := types.IRDefinition{
		Id:          "ratelimit-ir",
		Name:        "Rate Limit Test",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_BIOMETRIC),
		Description: "Rate limit test",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
	}
	err := keeper.CreateIR(ir)
	if err != nil {
		t.Fatalf("failed to create IR: %v", err)
	}

	rateLimit := types.IRRateLimit{
		IrId:             "ratelimit-ir",
		PerWalletPerHour: 10,
		PerWalletPerDay:  100,
	}
	err = keeper.SetRateLimit(rateLimit)
	if err != nil {
		t.Fatalf("failed to set rate limit: %v", err)
	}

	_, found := keeper.GetRateLimit("ratelimit-ir")
	if !found {
		t.Error("rate limit not found after setting")
	}
}

func TestIncrementRateLimit(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	ir := types.IRDefinition{
		Id:          "increment-ir",
		Name:        "Increment Test",
		Arena:       types.Arena(inclusionroutinespb.Arena_ARENA_BIOMETRIC),
		Description: "Increment test",
		Score:       100,
		LocaleTags:  []string{"global"},
		PrivacyTier: types.PrivacyTier(inclusionroutinespb.PrivacyTier_PRIVACY_TIER_HIGH),
		Version:     "1.0",
	}
	err := keeper.CreateIR(ir)
	if err != nil {
		t.Fatalf("failed to create IR: %v", err)
	}
	rateLimit := types.IRRateLimit{
		IrId:             "increment-ir",
		PerWalletPerHour: 5,
		PerWalletPerDay:  50,
	}
	err = keeper.SetRateLimit(rateLimit)
	if err != nil {
		t.Fatalf("failed to set rate limit: %v", err)
	}

	err = keeper.IncrementRateLimit("wallet1", "increment-ir")
	if err != nil {
		t.Fatalf("failed to increment rate limit: %v", err)
	}

	hourly, _ := keeper.GetRateLimitUsage("wallet1", "increment-ir")
	if hourly != 1 {
		t.Errorf("expected usage 1, got %d", hourly)
	}
}

func TestCleanupOldRateLimits(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	// This function shouldn't panic
	keeper.CleanupExpiredRateLimits()
}

func TestGetRateLimitUsageNotFound(t *testing.T) {
	keeper := NewKeeper(nil, "authority")

	hourly, daily := keeper.GetRateLimitUsage("wallet1", "nonexistent")
	if hourly != 0 || daily != 0 {
		t.Errorf("expected 0 usage for nonexistent IR, got hourly=%d, daily=%d", hourly, daily)
	}
}
