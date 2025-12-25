// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validator is a placeholder interface for validator operations
type Validator interface {
	GetOperator() sdk.ValAddress
	GetConsPubKey() (interface{}, error)
	GetConsAddr() (sdk.ConsAddress, error)
	GetStatus() int32
	GetTokens() math.Int
}

// StakingKeeper defines the expected staking keeper interface
type StakingKeeper interface {
	Validator(ctx context.Context, addr sdk.ValAddress) (Validator, error)
	ValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (Validator, error)
	Slash(ctx context.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor math.LegacyDec) (math.Int, error)
	Jail(ctx context.Context, consAddr sdk.ConsAddress) error
	Unjail(ctx context.Context, consAddr sdk.ConsAddress) error
	GetAllValidators(ctx context.Context) ([]Validator, error)
	PowerReduction(ctx context.Context) math.Int
}

// SlashingKeeper defines the expected slashing keeper interface
type SlashingKeeper interface {
	IsTombstoned(ctx context.Context, consAddr sdk.ConsAddress) bool
	Tombstone(ctx context.Context, consAddr sdk.ConsAddress) error
	JailUntil(ctx context.Context, consAddr sdk.ConsAddress, jailTime time.Time) error
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}
