// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package params

import (
	"testing"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

func TestStore_GetParams(t *testing.T) {
	defaultParams := types.DefaultParams()
	store := NewStore(defaultParams)

	params := store.GetParams()
	if params.VerificationThreshold != defaultParams.VerificationThreshold {
		t.Errorf("expected verification threshold %d, got %d",
			defaultParams.VerificationThreshold, params.VerificationThreshold)
	}
}

func TestStore_SetParams(t *testing.T) {
	store := NewStore(types.DefaultParams())

	newParams := types.DefaultParams()
	newParams.VerificationThreshold = 12000

	if err := store.SetParams(newParams); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	params := store.GetParams()
	if params.VerificationThreshold != 12000 {
		t.Errorf("expected verification threshold 12000, got %d", params.VerificationThreshold)
	}
}

func TestStore_SetParams_Invalid(t *testing.T) {
	store := NewStore(types.DefaultParams())

	invalidParams := types.DefaultParams()
	invalidParams.VerificationThreshold = 0 // Invalid

	if err := store.SetParams(invalidParams); err == nil {
		t.Error("expected error for invalid params, got nil")
	}
}

func TestStore_Concurrent(t *testing.T) {
	store := NewStore(types.DefaultParams())

	// Test concurrent reads and writes
	done := make(chan bool)

	// Start readers
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = store.GetParams()
			}
			done <- true
		}()
	}

	// Start writers
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				newParams := types.DefaultParams()
				newParams.VerificationThreshold = uint64(10000 + id)
				_ = store.SetParams(newParams)
			}
			done <- true
		}(i)
	}

	// Wait for completion
	for i := 0; i < 15; i++ {
		<-done
	}
}
