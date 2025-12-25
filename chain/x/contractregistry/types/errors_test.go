// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	// Test that all error constants are properly defined
	errorTests := []struct {
		name  string
		err   error
		check func(*testing.T, error)
	}{
		{
			name: "ErrContractNotFound",
			err:  ErrContractNotFound,
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Equal(t, "contract not found", err.Error())
			},
		},
		{
			name: "ErrContractAlreadyExists",
			err:  ErrContractAlreadyExists,
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Equal(t, "contract already exists", err.Error())
			},
		},
		{
			name: "ErrUnauthorized",
			err:  ErrUnauthorized,
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Equal(t, "unauthorized", err.Error())
			},
		},
		{
			name: "ErrRateLimitExceeded",
			err:  ErrRateLimitExceeded,
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Equal(t, "rate limit exceeded", err.Error())
			},
		},
		{
			name: "ErrKYCRequired",
			err:  ErrKYCRequired,
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Equal(t, "KYC required", err.Error())
			},
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.err)
		})
	}
}

func TestErrorIs(t *testing.T) {
	tests := []struct {
		name   string
		err1   error
		err2   error
		expect bool
	}{
		{
			name:   "same error",
			err1:   ErrContractNotFound,
			err2:   ErrContractNotFound,
			expect: true,
		},
		{
			name:   "different error",
			err1:   ErrContractNotFound,
			err2:   ErrUnauthorized,
			expect: false,
		},
		{
			name:   "wrapped error",
			err1:   errors.New("wrapped: " + ErrContractNotFound.Error()),
			err2:   ErrContractNotFound,
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err1, tt.err2)
			require.Equal(t, tt.expect, result)
		})
	}
}

func TestErrorTypes(t *testing.T) {
	// Test contract errors
	contractErrors := []error{
		ErrContractNotFound,
		ErrContractAlreadyExists,
		ErrContractNotActive,
		ErrContractPaused,
		ErrContractFrozen,
		ErrContractDeprecated,
	}

	for _, err := range contractErrors {
		require.Error(t, err)
	}

	// Test authorization errors
	authErrors := []error{
		ErrUnauthorized,
		ErrNotContractAdmin,
		ErrNotContractCreator,
		ErrInvalidSigner,
	}

	for _, err := range authErrors {
		require.Error(t, err)
	}

	// Test validation errors
	validationErrors := []error{
		ErrInvalidContractAddress,
		ErrInvalidCodeID,
		ErrInvalidMetadata,
		ErrInvalidSecurityPolicy,
		ErrInvalidCompliance,
		ErrInvalidParams,
	}

	for _, err := range validationErrors {
		require.Error(t, err)
	}
}
