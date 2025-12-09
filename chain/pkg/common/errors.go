package common

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
)

const (
	// DefaultCodespace is the default codespace for common errors
	DefaultCodespace = "common"
)

// Common error codes that can be reused across modules
const (
	CodeInvalidAddress         uint32 = 1001
	CodeInvalidAmount          uint32 = 1002
	CodeInvalidPagination      uint32 = 1003
	CodeInsufficientBalance    uint32 = 1004
	CodeUnauthorized           uint32 = 1005
	CodeNotFound               uint32 = 1006
	CodeAlreadyExists          uint32 = 1007
	CodeInvalidRequest         uint32 = 1008
	CodeInternalError          uint32 = 1009
	CodePermissionDenied       uint32 = 1010
	CodeInvalidSignature       uint32 = 1011
	CodeInvalidState           uint32 = 1012
	CodeTimeout                uint32 = 1013
	CodeRateLimitExceeded      uint32 = 1014
	CodeInvalidProof           uint32 = 1015
	CodeValidationFailed       uint32 = 1016
	CodeSerializationFailed    uint32 = 1017
	CodeDeserializationFailed  uint32 = 1018
	CodeOperationNotSupported  uint32 = 1019
)

var (
	// ErrInvalidAddress indicates an address is malformed or invalid
	ErrInvalidAddress = errorsmod.Register(DefaultCodespace, CodeInvalidAddress, "invalid address")

	// ErrInvalidAmount indicates an amount is zero, negative, or otherwise invalid
	ErrInvalidAmount = errorsmod.Register(DefaultCodespace, CodeInvalidAmount, "invalid amount")

	// ErrInvalidPagination indicates pagination parameters are invalid
	ErrInvalidPagination = errorsmod.Register(DefaultCodespace, CodeInvalidPagination, "invalid pagination")

	// ErrInsufficientBalance indicates insufficient funds for an operation
	ErrInsufficientBalance = errorsmod.Register(DefaultCodespace, CodeInsufficientBalance, "insufficient balance")

	// ErrUnauthorized indicates the caller lacks authorization
	ErrUnauthorized = errorsmod.Register(DefaultCodespace, CodeUnauthorized, "unauthorized")

	// ErrNotFound indicates a requested resource was not found
	ErrNotFound = errorsmod.Register(DefaultCodespace, CodeNotFound, "not found")

	// ErrAlreadyExists indicates a resource already exists
	ErrAlreadyExists = errorsmod.Register(DefaultCodespace, CodeAlreadyExists, "already exists")

	// ErrInvalidRequest indicates the request parameters are invalid
	ErrInvalidRequest = errorsmod.Register(DefaultCodespace, CodeInvalidRequest, "invalid request")

	// ErrInternalError indicates an unexpected internal error
	ErrInternalError = errorsmod.Register(DefaultCodespace, CodeInternalError, "internal error")

	// ErrPermissionDenied indicates the operation is not permitted
	ErrPermissionDenied = errorsmod.Register(DefaultCodespace, CodePermissionDenied, "permission denied")

	// ErrInvalidSignature indicates a cryptographic signature is invalid
	ErrInvalidSignature = errorsmod.Register(DefaultCodespace, CodeInvalidSignature, "invalid signature")

	// ErrInvalidState indicates the system is in an invalid state for the operation
	ErrInvalidState = errorsmod.Register(DefaultCodespace, CodeInvalidState, "invalid state")

	// ErrTimeout indicates an operation timed out
	ErrTimeout = errorsmod.Register(DefaultCodespace, CodeTimeout, "operation timeout")

	// ErrRateLimitExceeded indicates a rate limit was exceeded
	ErrRateLimitExceeded = errorsmod.Register(DefaultCodespace, CodeRateLimitExceeded, "rate limit exceeded")

	// ErrInvalidProof indicates a cryptographic proof is invalid
	ErrInvalidProof = errorsmod.Register(DefaultCodespace, CodeInvalidProof, "invalid proof")

	// ErrValidationFailed indicates validation checks failed
	ErrValidationFailed = errorsmod.Register(DefaultCodespace, CodeValidationFailed, "validation failed")

	// ErrSerializationFailed indicates data could not be serialized
	ErrSerializationFailed = errorsmod.Register(DefaultCodespace, CodeSerializationFailed, "serialization failed")

	// ErrDeserializationFailed indicates data could not be deserialized
	ErrDeserializationFailed = errorsmod.Register(DefaultCodespace, CodeDeserializationFailed, "deserialization failed")

	// ErrOperationNotSupported indicates the operation is not supported
	ErrOperationNotSupported = errorsmod.Register(DefaultCodespace, CodeOperationNotSupported, "operation not supported")
)

// WrapError wraps an error with additional context.
// This provides consistent error wrapping across modules.
//
// Parameters:
//   - baseErr: Base error to wrap
//   - format: Format string for context message
//   - args: Arguments for format string
//
// Returns:
//   - error: Wrapped error with context
//
// Example usage:
//   return common.WrapError(common.ErrInvalidAddress, "failed to validate sender: %s", msg.Sender)
func WrapError(baseErr error, format string, args ...interface{}) error {
	return errorsmod.Wrapf(baseErr, format, args...)
}

// WrapErrorf is an alias for WrapError for consistency with Cosmos SDK patterns
func WrapErrorf(baseErr error, format string, args ...interface{}) error {
	return errorsmod.Wrapf(baseErr, format, args...)
}

// NewError creates a new error with a message.
// Use this for module-specific errors that don't fit the common error types.
//
// Parameters:
//   - codespace: Module codespace
//   - code: Error code (should be unique within codespace)
//   - msg: Error message
//
// Returns:
//   - error: Registered error
//
// Example usage:
//   ErrPoolNotFound := common.NewError("dex", 2001, "liquidity pool not found")
func NewError(codespace string, code uint32, msg string) error {
	return errorsmod.Register(codespace, code, msg)
}

// IsNotFoundError checks if an error is a "not found" error.
// This provides consistent error checking across modules.
//
// Parameters:
//   - err: Error to check
//
// Returns:
//   - bool: True if error is a not found error
//
// Example usage:
//   if common.IsNotFoundError(err) {
//       return status.Error(codes.NotFound, err.Error())
//   }
func IsNotFoundError(err error) bool {
	return errorsmod.IsOf(err, ErrNotFound)
}

// IsUnauthorizedError checks if an error is an authorization error.
//
// Parameters:
//   - err: Error to check
//
// Returns:
//   - bool: True if error is an authorization error
func IsUnauthorizedError(err error) bool {
	return errorsmod.IsOf(err, ErrUnauthorized) || errorsmod.IsOf(err, ErrPermissionDenied)
}

// IsValidationError checks if an error is a validation error.
//
// Parameters:
//   - err: Error to check
//
// Returns:
//   - bool: True if error is a validation error
func IsValidationError(err error) bool {
	return errorsmod.IsOf(err, ErrValidationFailed) ||
		errorsmod.IsOf(err, ErrInvalidAddress) ||
		errorsmod.IsOf(err, ErrInvalidAmount) ||
		errorsmod.IsOf(err, ErrInvalidRequest)
}

// FormatError formats an error message with context.
// This provides consistent error message formatting.
//
// Parameters:
//   - operation: Operation that failed (e.g., "create pool", "transfer tokens")
//   - err: Original error
//
// Returns:
//   - string: Formatted error message
//
// Example usage:
//   errMsg := common.FormatError("create liquidity pool", err)
func FormatError(operation string, err error) string {
	return fmt.Sprintf("failed to %s: %s", operation, err.Error())
}
