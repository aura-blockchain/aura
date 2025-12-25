// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"
)

func TestDefaultGenesisState(t *testing.T) {
	genesis := DefaultGenesisState()

	if genesis == nil {
		t.Fatal("DefaultGenesisState should not return nil")
	}

	// Validate default params are set correctly
	if err := ValidateParams(genesis.Params); err != nil {
		t.Errorf("DefaultGenesisState should have valid params: %v", err)
	}
}

func TestGenesisStateValidate_Valid(t *testing.T) {
	genesis := DefaultGenesisState()

	err := genesis.Validate()
	if err != nil {
		t.Errorf("Validate should not return error for default genesis: %v", err)
	}
}

func TestGenesisStateValidate_InvalidParams(t *testing.T) {
	genesis := DefaultGenesisState()
	genesis.Params.LargeTransactionThreshold = 0

	err := genesis.Validate()
	if err == nil {
		t.Error("Validate should return error for invalid params")
	}
}

func TestGenesisStateConsistency(t *testing.T) {
	genesis1 := DefaultGenesisState()
	genesis2 := DefaultGenesisState()

	// Verify that multiple calls return consistent values
	if genesis1.Params.LargeTransactionThreshold != genesis2.Params.LargeTransactionThreshold {
		t.Error("DefaultGenesisState should return consistent values")
	}

	if genesis1.Params.AnomalyThreshold != genesis2.Params.AnomalyThreshold {
		t.Error("DefaultGenesisState should return consistent values")
	}
}
