// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateAddress validates a bech32 address string and returns an AccAddress.
// This is the primary function for validating user addresses across all modules.
//
// Security considerations:
//   - Validates bech32 format and checksum
//   - Returns typed AccAddress for type safety
//   - Consistent error messages across all modules
//
// Parameters:
//   - addr: Bech32-encoded address string
//
// Returns:
//   - sdk.AccAddress: Parsed address if valid
//   - error: Descriptive error if invalid or empty
//
// Example usage:
//
//	address, err := common.ValidateAddress(msg.Sender)
//	if err != nil {
//	    return nil, status.Error(codes.InvalidArgument, err.Error())
//	}
func ValidateAddress(addr string) (sdk.AccAddress, error) {
	if addr == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	parsedAddr, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %w", err)
	}

	return parsedAddr, nil
}

// ValidateAddresses validates multiple bech32 addresses and returns typed AccAddress slice.
// Useful for batch validation in multisig, delegation, and governance operations.
//
// Security considerations:
//   - Validates all addresses before returning any
//   - Returns error on first invalid address with index
//   - Prevents partial validation that could lead to inconsistent state
//
// Parameters:
//   - addrs: Slice of bech32-encoded address strings
//
// Returns:
//   - []sdk.AccAddress: Parsed addresses if all valid
//   - error: Descriptive error with index of first invalid address
//
// Example usage:
//
//	addresses, err := common.ValidateAddresses(msg.Signers)
//	if err != nil {
//	    return nil, status.Error(codes.InvalidArgument, err.Error())
//	}
func ValidateAddresses(addrs []string) ([]sdk.AccAddress, error) {
	if len(addrs) == 0 {
		return nil, fmt.Errorf("address list cannot be empty")
	}

	result := make([]sdk.AccAddress, len(addrs))
	for i, addr := range addrs {
		parsedAddr, err := ValidateAddress(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid address at index %d: %w", i, err)
		}
		result[i] = parsedAddr
	}

	return result, nil
}

// MustValidateAddress validates an address and panics on error.
// Use only in contexts where address validity is guaranteed (e.g., genesis, tests).
// DO NOT use in message handlers or query servers where user input is involved.
//
// Parameters:
//   - addr: Bech32-encoded address string
//
// Returns:
//   - sdk.AccAddress: Parsed address
//
// Panics:
//   - If address is invalid
//
// Example usage:
//
//	moduleAddr := common.MustValidateAddress("aura1...")
func MustValidateAddress(addr string) sdk.AccAddress {
	parsedAddr, err := ValidateAddress(addr)
	if err != nil {
		panic(fmt.Sprintf("MustValidateAddress failed: %s", err))
	}
	return parsedAddr
}
