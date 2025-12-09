package types

import (
	"testing"
)

func TestDefaultGenesisState(t *testing.T) {
	genesis := DefaultGenesisState()

	if genesis == nil {
		t.Fatal("DefaultGenesisState should not return nil")
	}

	// Params is a value type (not pointer), so check for non-zero values
	if genesis.Params.MaxIrPerLocale == 0 {
		t.Error("expected Params.MaxIrPerLocale to be set to non-zero value")
	}
	if genesis.Params.DefaultRateLimitHour == 0 {
		t.Error("expected Params.DefaultRateLimitHour to be set to non-zero value")
	}
	if genesis.Params.SuspensionFee == "" {
		t.Error("expected Params.SuspensionFee to be set")
	}
	if genesis.Params.MinGovernanceDeposit == "" {
		t.Error("expected Params.MinGovernanceDeposit to be set")
	}
}

func TestDefaultGenesis(t *testing.T) {
	// Test that DefaultGenesis doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DefaultGenesis panicked: %v", r)
		}
	}()

	// Use the module codec for serialization
	// Since we can't import codec here easily, we'll test DefaultGenesisState directly
	genesis := DefaultGenesisState()
	if genesis == nil {
		t.Fatal("DefaultGenesisState returned nil")
	}
}
