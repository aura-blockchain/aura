// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Privacy module error codes
var (
	ErrInvalidRingSize     = errorsmod.Register(ModuleName, 1, "invalid ring size")
	ErrInvalidMixingParams = errorsmod.Register(ModuleName, 2, "invalid mixing parameters")
	ErrInvalidCommitment   = errorsmod.Register(ModuleName, 3, "invalid commitment")
	ErrInvalidProof        = errorsmod.Register(ModuleName, 4, "invalid proof")
	ErrNullifierExists     = errorsmod.Register(ModuleName, 5, "nullifier already exists")
	ErrInvalidNullifier    = errorsmod.Register(ModuleName, 6, "invalid nullifier")
	ErrKeyImageAlreadyUsed = errorsmod.Register(ModuleName, 7, "key image already used")
)
