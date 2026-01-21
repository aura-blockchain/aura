// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Data registry module error codes
var (
	// Data item errors (1-19)
	ErrInvalidDataID         = errorsmod.Register(ModuleName, 1, "invalid data ID")
	ErrDataItemNotFound      = errorsmod.Register(ModuleName, 2, "data item not found")
	ErrDataItemAlreadyExists = errorsmod.Register(ModuleName, 3, "data item already exists")
	ErrDataItemRevoked       = errorsmod.Register(ModuleName, 4, "data item is revoked")
	ErrDataItemExpired       = errorsmod.Register(ModuleName, 5, "data item is expired")

	// Ownership and access errors (20-29)
	ErrInvalidOwner = errorsmod.Register(ModuleName, 20, "invalid owner address")
	ErrUnauthorized = errorsmod.Register(ModuleName, 21, "unauthorized")
	ErrAccessDenied = errorsmod.Register(ModuleName, 22, "access denied")

	// Data validation errors (30-49)
	ErrInvalidDataType        = errorsmod.Register(ModuleName, 30, "invalid data type")
	ErrInvalidContentHash     = errorsmod.Register(ModuleName, 31, "invalid content hash")
	ErrInvalidStorageLocation = errorsmod.Register(ModuleName, 32, "invalid storage location")
	ErrInvalidGeoLocation     = errorsmod.Register(ModuleName, 33, "invalid geo location")
	ErrInvalidAccessPolicy    = errorsmod.Register(ModuleName, 34, "invalid access policy")

	// Verification errors (50-59)
	ErrInvalidVerifier          = errorsmod.Register(ModuleName, 50, "invalid verifier")
	ErrInvalidVerificationLevel = errorsmod.Register(ModuleName, 51, "invalid verification level")

	// Limit errors (60-69)
	ErrMaxDataItemsExceeded = errorsmod.Register(ModuleName, 60, "maximum data items per user exceeded")
	ErrStorageSizeExceeded  = errorsmod.Register(ModuleName, 61, "storage size limit exceeded")
)
