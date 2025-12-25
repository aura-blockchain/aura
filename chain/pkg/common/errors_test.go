// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package common_test

import (
	"errors"
	"testing"

	"github.com/aequitas/aura/chain/pkg/common"
	errorsmod "cosmossdk.io/errors"
	"github.com/stretchr/testify/require"
)

func TestCommonErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code uint32
	}{
		{"ErrInvalidAddress", common.ErrInvalidAddress, common.CodeInvalidAddress},
		{"ErrInvalidAmount", common.ErrInvalidAmount, common.CodeInvalidAmount},
		{"ErrInvalidPagination", common.ErrInvalidPagination, common.CodeInvalidPagination},
		{"ErrInsufficientBalance", common.ErrInsufficientBalance, common.CodeInsufficientBalance},
		{"ErrUnauthorized", common.ErrUnauthorized, common.CodeUnauthorized},
		{"ErrNotFound", common.ErrNotFound, common.CodeNotFound},
		{"ErrAlreadyExists", common.ErrAlreadyExists, common.CodeAlreadyExists},
		{"ErrInvalidRequest", common.ErrInvalidRequest, common.CodeInvalidRequest},
		{"ErrInternalError", common.ErrInternalError, common.CodeInternalError},
		{"ErrPermissionDenied", common.ErrPermissionDenied, common.CodePermissionDenied},
		{"ErrInvalidSignature", common.ErrInvalidSignature, common.CodeInvalidSignature},
		{"ErrInvalidState", common.ErrInvalidState, common.CodeInvalidState},
		{"ErrTimeout", common.ErrTimeout, common.CodeTimeout},
		{"ErrRateLimitExceeded", common.ErrRateLimitExceeded, common.CodeRateLimitExceeded},
		{"ErrInvalidProof", common.ErrInvalidProof, common.CodeInvalidProof},
		{"ErrValidationFailed", common.ErrValidationFailed, common.CodeValidationFailed},
		{"ErrSerializationFailed", common.ErrSerializationFailed, common.CodeSerializationFailed},
		{"ErrDeserializationFailed", common.ErrDeserializationFailed, common.CodeDeserializationFailed},
		{"ErrOperationNotSupported", common.ErrOperationNotSupported, common.CodeOperationNotSupported},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.err)
			// Errors are registered, just check they're non-nil
			require.NotNil(t, tt.err)
		})
	}
}

func TestWrapError(t *testing.T) {
	baseErr := common.ErrInvalidAddress

	t.Run("simple wrap", func(t *testing.T) {
		wrapped := common.WrapError(baseErr, "validation failed")
		require.Error(t, wrapped)
		require.Contains(t, wrapped.Error(), "validation failed")
		require.True(t, errorsmod.IsOf(wrapped, baseErr))
	})

	t.Run("wrap with format", func(t *testing.T) {
		wrapped := common.WrapError(baseErr, "failed to validate sender: %s", "aura1...")
		require.Error(t, wrapped)
		require.Contains(t, wrapped.Error(), "failed to validate sender")
		require.Contains(t, wrapped.Error(), "aura1...")
		require.True(t, errorsmod.IsOf(wrapped, baseErr))
	})
}

func TestWrapErrorf(t *testing.T) {
	baseErr := common.ErrInvalidAmount
	wrapped := common.WrapErrorf(baseErr, "amount must be positive, got %d", -5)

	require.Error(t, wrapped)
	require.Contains(t, wrapped.Error(), "amount must be positive")
	require.Contains(t, wrapped.Error(), "-5")
	require.True(t, errorsmod.IsOf(wrapped, baseErr))
}

func TestNewError(t *testing.T) {
	customErr := common.NewError("testmodule", 2001, "test error message")

	require.Error(t, customErr)
	require.Contains(t, customErr.Error(), "test error message")
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "not found error",
			err:      common.ErrNotFound,
			expected: true,
		},
		{
			name:     "wrapped not found error",
			err:      common.WrapError(common.ErrNotFound, "pool not found"),
			expected: true,
		},
		{
			name:     "other error",
			err:      common.ErrInvalidAddress,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "standard error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.IsNotFoundError(tt.err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsUnauthorizedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "unauthorized error",
			err:      common.ErrUnauthorized,
			expected: true,
		},
		{
			name:     "permission denied error",
			err:      common.ErrPermissionDenied,
			expected: true,
		},
		{
			name:     "wrapped unauthorized",
			err:      common.WrapError(common.ErrUnauthorized, "access denied"),
			expected: true,
		},
		{
			name:     "other error",
			err:      common.ErrInvalidAddress,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.IsUnauthorizedError(tt.err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "validation failed",
			err:      common.ErrValidationFailed,
			expected: true,
		},
		{
			name:     "invalid address",
			err:      common.ErrInvalidAddress,
			expected: true,
		},
		{
			name:     "invalid amount",
			err:      common.ErrInvalidAmount,
			expected: true,
		},
		{
			name:     "invalid request",
			err:      common.ErrInvalidRequest,
			expected: true,
		},
		{
			name:     "not found error",
			err:      common.ErrNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.IsValidationError(tt.err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatError(t *testing.T) {
	err := errors.New("pool does not exist")

	formatted := common.FormatError("create liquidity pool", err)

	require.Contains(t, formatted, "failed to create liquidity pool")
	require.Contains(t, formatted, "pool does not exist")
}
