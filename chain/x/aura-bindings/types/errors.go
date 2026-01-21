// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"cosmossdk.io/errors"
)

// x/aura-bindings module sentinel errors
var (
	ErrInvalidQuery             = errors.Register(ModuleName, 1, "invalid query")
	ErrInvalidMessage           = errors.Register(ModuleName, 2, "invalid message")
	ErrQueryRateLimitExceeded   = errors.Register(ModuleName, 3, "query rate limit exceeded")
	ErrMessageRateLimitExceeded = errors.Register(ModuleName, 4, "message rate limit exceeded")
	ErrUnauthorized             = errors.Register(ModuleName, 5, "unauthorized")
	ErrInvalidParam             = errors.Register(ModuleName, 6, "invalid parameter")
	ErrQueryFailed              = errors.Register(ModuleName, 7, "query failed")
	ErrMessageFailed            = errors.Register(ModuleName, 8, "message failed")
	ErrInvalidAddress           = errors.Register(ModuleName, 9, "invalid address")
	ErrNotFound                 = errors.Register(ModuleName, 10, "not found")
)
