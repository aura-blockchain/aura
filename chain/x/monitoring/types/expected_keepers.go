// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// StakingKeeper defines the expected interface for the staking module keeper.
// This is used to get validator information for deterministic network health metrics.
type StakingKeeper interface {
	// GetAllValidators returns all validators in the store (any status)
	GetAllValidators(ctx context.Context) ([]Validator, error)
}

// Validator defines the expected interface for a validator from the staking module.
type Validator interface {
	// GetOperator returns the validator's operator address
	GetOperator() string
	// GetStatus returns the validator's bond status (0=unbonded, 1=unbonding, 2=bonded)
	GetStatus() int32
	// GetTokens returns the validator's staked tokens
	GetTokens() math.Int
	// GetConsAddr returns the consensus address
	GetConsAddr() (sdk.ConsAddress, error)
	// GetMoniker returns the validator's moniker/name
	GetMoniker() string
	// IsBonded returns true if validator is bonded
	IsBonded() bool
	// IsJailed returns true if validator is jailed
	IsJailed() bool
}
