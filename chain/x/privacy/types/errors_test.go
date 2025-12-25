// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrors_Defined(t *testing.T) {
	// Test that all error types are properly defined
	require.NotNil(t, ErrInvalidRingSize)
	require.NotNil(t, ErrInvalidMixingParams)
	require.NotNil(t, ErrInvalidCommitment)
	require.NotNil(t, ErrInvalidProof)
	require.NotNil(t, ErrNullifierExists)
	require.NotNil(t, ErrInvalidNullifier)
}

func TestErrors_Messages(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
	}{
		{"invalid ring size", ErrInvalidRingSize, "invalid ring size"},
		{"invalid mixing params", ErrInvalidMixingParams, "invalid mixing parameters"},
		{"invalid commitment", ErrInvalidCommitment, "invalid commitment"},
		{"invalid proof", ErrInvalidProof, "invalid proof"},
		{"nullifier exists", ErrNullifierExists, "nullifier already exists"},
		{"invalid nullifier", ErrInvalidNullifier, "invalid nullifier"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.message, tt.err.Error())
		})
	}
}

func TestErrors_Unique(t *testing.T) {
	// Ensure all errors are unique instances
	errors := []error{
		ErrInvalidRingSize,
		ErrInvalidMixingParams,
		ErrInvalidCommitment,
		ErrInvalidProof,
		ErrNullifierExists,
		ErrInvalidNullifier,
	}

	for i, err1 := range errors {
		for j, err2 := range errors {
			if i != j {
				require.NotEqual(t, err1, err2, "errors at index %d and %d should not be equal", i, j)
			}
		}
	}
}

func TestErrors_CanBeCompared(t *testing.T) {
	// Test that errors can be compared using ==
	err := ErrInvalidRingSize
	require.True(t, err == ErrInvalidRingSize)
	require.False(t, err == ErrInvalidMixingParams)
}
