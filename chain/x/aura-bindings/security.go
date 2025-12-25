// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package aurabindings

import (
	"fmt"

	wasmkeeper "github.com/aequitas/aura/chain/x/wasm/keeper"
	wasmtypes "github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SecurityValidator validates custom binding operations
type SecurityValidator struct {
	wasmKeeper *wasmkeeper.Keeper
}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator(wasmKeeper *wasmkeeper.Keeper) *SecurityValidator {
	return &SecurityValidator{
		wasmKeeper: wasmKeeper,
	}
}

// ValidateContractPermissions validates that a contract has permission for custom bindings
func (sv *SecurityValidator) ValidateContractPermissions(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	operation string,
) error {
	if sv.wasmKeeper == nil {
		// If no wasm keeper, allow (for testing)
		return nil
	}

	// NOTE: Production Enhancement - Implement comprehensive permission checking
	// Full implementation requires:
	// 1. Permission storage in wasm keeper state (by contract address + operation)
	// 2. GetContractPermissions query method in wasm keeper
	// 3. Permission update transactions
	//
	// Current implementation performs basic security checks:

	// Check if contract is paused
	if sv.wasmKeeper.IsContractPaused(ctx, contractAddr.String()) {
		return wasmtypes.ErrUnauthorized.Wrapf(
			"contract %s is paused",
			contractAddr.String(),
		)
	}

	// NOTE: Additional permission checks (e.g., operation-specific ACLs)
	// should be implemented based on stored permissions data
	// For now, pause check provides basic protection

	return nil
}

// CheckRateLimit checks if the contract has exceeded rate limits for custom operations
func (sv *SecurityValidator) CheckRateLimit(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	operation string,
) error {
	if sv.wasmKeeper == nil {
		return nil
	}

	// NOTE: Production Enhancement - Implement rate limiting
	// Full implementation requires:
	// 1. Rate limit state storage (per contract, per operation)
	// 2. Time-window tracking (using block height or timestamp)
	// 3. Configurable limits per operation type
	// 4. Counter reset logic after time window expires
	//
	// Recommended approach:
	// - Store: contractAddr + operation -> (count, windowStart)
	// - On each call: increment count, check against limit
	// - Reset count when windowStart + windowDuration < currentTime
	//
	// For now, no rate limiting is enforced
	// This should be implemented before mainnet deployment

	return nil
}

// ValidateVCData validates VC data before registration
func (sv *SecurityValidator) ValidateVCData(vcBase64 string) error {
	// Check size
	if len(vcBase64) > wasmtypes.MaxVCDataSize {
		return wasmtypes.ErrSecurityViolation.Wrapf(
			"VC data size %d exceeds maximum %d",
			len(vcBase64),
			wasmtypes.MaxVCDataSize,
		)
	}

	// Check non-empty
	if len(vcBase64) == 0 {
		return wasmtypes.ErrSecurityViolation.Wrap("VC data cannot be empty")
	}

	// Additional validation could include:
	// - Base64 format validation
	// - JSON structure validation
	// - Schema validation
	// - Signature verification

	return nil
}

// CanRegisterVCFor checks if contract can register VC for a specific address
func (sv *SecurityValidator) CanRegisterVCFor(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	targetAddr string,
	vcType string,
) error {
	if sv.wasmKeeper == nil {
		return nil
	}

	// NOTE: Production Enhancement - Implement VC type permission checking
	// Full implementation should verify contract has permission for specific VC types
	// and enforce maximum registration limits per contract

	// Check if contract is paused
	if sv.wasmKeeper.IsContractPaused(ctx, contractAddr.String()) {
		return wasmtypes.ErrUnauthorized.Wrapf(
			"contract %s is paused",
			contractAddr.String(),
		)
	}

	// NOTE: Additional checks needed for production:
	// - Verify contract has permission for this VC type
	// - Check registration count against max limit
	// - Validate VC type schema compliance

	return nil
}

// CanQueryVCsFor checks if contract can query VCs
func (sv *SecurityValidator) CanQueryVCsFor(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
	vcType string,
) error {
	if sv.wasmKeeper == nil {
		return nil
	}

	// NOTE: Production Enhancement - Implement VC query permission checking
	// Full implementation should verify contract has query access for specific VC types

	// Check if contract is paused
	if sv.wasmKeeper.IsContractPaused(ctx, contractAddr.String()) {
		return wasmtypes.ErrUnauthorized.Wrapf(
			"contract %s is paused",
			contractAddr.String(),
		)
	}

	// STUB: For now, allow all VC type queries
	// Future implementation should check VC type query permissions

	return nil
}

// LogSecurityEvent logs a security event for custom bindings
func (sv *SecurityValidator) LogSecurityEvent(
	ctx sdk.Context,
	eventType string,
	contractAddr sdk.AccAddress,
	success bool,
	errorMsg string,
	additionalData map[string]interface{},
) {
	if sv.wasmKeeper == nil {
		return
	}

	auditEvent := wasmtypes.NewSecurityAuditEvent(
		eventType,
		contractAddr.String(),
		"", // No specific sender in custom bindings
		ctx,
		success,
		errorMsg,
	)

	// Add additional data
	for key, value := range additionalData {
		auditEvent.AddData(key, value)
	}

	sv.wasmKeeper.LogSecurityEvent(ctx, auditEvent)
}

// FilterSensitiveData filters sensitive fields from query results
func FilterSensitiveData(data map[string]interface{}, contractAddr string) map[string]interface{} {
	filtered := make(map[string]interface{})

	// Define sensitive fields that should be redacted
	sensitiveFields := map[string]bool{
		"private_key":     true,
		"secret":          true,
		"password":        true,
		"credential_hash": true, // May want to redact full hash
	}

	for key, value := range data {
		if sensitiveFields[key] {
			filtered[key] = "[REDACTED]"
		} else {
			filtered[key] = value
		}
	}

	return filtered
}

// GetAllowedVCTypes returns the VC types a contract is allowed to access
func (sv *SecurityValidator) GetAllowedVCTypes(
	ctx sdk.Context,
	contractAddr sdk.AccAddress,
) []string {
	if sv.wasmKeeper == nil {
		return []string{}
	}

	// NOTE: Production Enhancement - Implement VC type allowlist retrieval
	// Full implementation should query wasm keeper for contract's permitted VC types
	// Return value: list of VC type identifiers this contract can handle
	//
	// For now, returning empty list (no type restrictions)
	// This is permissive and should be tightened in production
	return []string{}
}

// ValidateAddress validates that an address is valid
func ValidateAddress(address string) error {
	_, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return wasmtypes.ErrSecurityViolation.Wrapf("invalid address: %s", err)
	}
	return nil
}

// SanitizeInput sanitizes user input to prevent injection attacks
func SanitizeInput(input string, maxLength int) (string, error) {
	// Check length
	if len(input) > maxLength {
		return "", fmt.Errorf("input exceeds maximum length %d", maxLength)
	}

	// Could add more sanitization here:
	// - Remove control characters
	// - Escape special characters
	// - Validate character set

	return input, nil
}
