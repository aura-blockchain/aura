package params

import (
	"testing"

	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
)

func TestNewStore(t *testing.T) {
	defaults := types.DefaultParams()
	store := NewStore(defaults)

	if store == nil {
		t.Fatal("expected store to be non-nil")
	}

	params := store.GetParams()
	if params.MaxIRPerLocale != defaults.MaxIRPerLocale {
		t.Errorf("expected MaxIRPerLocale=%d, got %d", defaults.MaxIRPerLocale, params.MaxIRPerLocale)
	}
}

func TestSetParams(t *testing.T) {
	store := NewStore(types.DefaultParams())

	newParams := types.Params{
		MaxIRPerLocale:       100,
		DefaultRateLimitHour: 20,
		SuspensionFee:        "2000000uaura",
		MinGovernanceDeposit: "20000000uaura",
	}

	err := store.SetParams(newParams)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	retrieved := store.GetParams()
	if retrieved.MaxIRPerLocale != 100 {
		t.Errorf("expected MaxIRPerLocale=100, got %d", retrieved.MaxIRPerLocale)
	}
	if retrieved.DefaultRateLimitHour != 20 {
		t.Errorf("expected DefaultRateLimitHour=20, got %d", retrieved.DefaultRateLimitHour)
	}
}

func TestSetParamsValidation(t *testing.T) {
	store := NewStore(types.DefaultParams())

	invalidParams := types.Params{
		MaxIRPerLocale:       -1, // Invalid
		DefaultRateLimitHour: 10,
		SuspensionFee:        "1000000uaura",
		MinGovernanceDeposit: "10000000uaura",
	}

	err := store.SetParams(invalidParams)
	if err == nil {
		t.Error("expected validation error for negative MaxIRPerLocale")
	}
}

func TestPanicOnInvalidDefaults(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid default params")
		}
	}()

	invalidDefaults := types.Params{
		MaxIRPerLocale:       -1,
		DefaultRateLimitHour: 10,
		SuspensionFee:        "1000000uaura",
		MinGovernanceDeposit: "10000000uaura",
	}

	NewStore(invalidDefaults)
}
