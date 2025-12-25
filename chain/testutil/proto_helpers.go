// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

// Package testutil provides helpers for creating correctly-typed test data.
//
// IMPORTANT: This project uses gogoproto annotations that transform proto types.
// See docs/GOGOPROTO_TYPES.md for full documentation.
//
// Use these helpers instead of timestamppb, durationpb, or string literals
// when creating test data for proto-generated types.
package testutil

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// -----------------------------------------------------------------------------
// Timestamp Helpers
// -----------------------------------------------------------------------------

// Now returns the current time as time.Time (for non-nullable timestamp fields).
func Now() time.Time {
	return time.Now()
}

// NowPtr returns a pointer to the current time (for nullable timestamp fields).
func NowPtr() *time.Time {
	t := time.Now()
	return &t
}

// TimeAgo returns a time in the past (for non-nullable timestamp fields).
func TimeAgo(d time.Duration) time.Time {
	return time.Now().Add(-d)
}

// TimeAgoPtr returns a pointer to a time in the past (for nullable timestamp fields).
func TimeAgoPtr(d time.Duration) *time.Time {
	t := time.Now().Add(-d)
	return &t
}

// TimeFromNow returns a time in the future (for non-nullable timestamp fields).
func TimeFromNow(d time.Duration) time.Time {
	return time.Now().Add(d)
}

// TimeFromNowPtr returns a pointer to a time in the future (for nullable timestamp fields).
func TimeFromNowPtr(d time.Duration) *time.Time {
	t := time.Now().Add(d)
	return &t
}

// TimePtr converts a time.Time to *time.Time (for nullable timestamp fields).
func TimePtr(t time.Time) *time.Time {
	return &t
}

// ZeroTime returns the zero value for time.Time.
// Use this instead of nil when testing empty/unset timestamp values.
func ZeroTime() time.Time {
	return time.Time{}
}

// -----------------------------------------------------------------------------
// Math Helpers
// -----------------------------------------------------------------------------

// NewInt creates a math.Int from an int64.
// Use this instead of string literals like "1000".
func NewInt(i int64) sdkmath.Int {
	return sdkmath.NewInt(i)
}

// NewIntFromString creates a math.Int from a string.
func NewIntFromString(s string) sdkmath.Int {
	i, ok := sdkmath.NewIntFromString(s)
	if !ok {
		panic("invalid int string: " + s)
	}
	return i
}

// ZeroInt returns a zero math.Int.
func ZeroInt() sdkmath.Int {
	return sdkmath.ZeroInt()
}

// NewDec creates a math.LegacyDec from a string like "0.05".
// Use this instead of string literals.
func NewDec(s string) sdkmath.LegacyDec {
	return sdkmath.LegacyMustNewDecFromStr(s)
}

// NewDecFromInt creates a math.LegacyDec from an int64.
func NewDecFromInt(i int64) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDec(i)
}

// NewDecWithPrec creates a math.LegacyDec with precision.
// Example: NewDecWithPrec(5, 2) = 0.05
func NewDecWithPrec(i int64, prec int64) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecWithPrec(i, prec)
}

// ZeroDec returns a zero math.LegacyDec.
func ZeroDec() sdkmath.LegacyDec {
	return sdkmath.LegacyZeroDec()
}

// -----------------------------------------------------------------------------
// Coin Helpers
// -----------------------------------------------------------------------------

// NewCoin creates a sdk.Coin (value type, not pointer).
// Use this instead of &sdk.Coin{} when nullable=false.
func NewCoin(denom string, amount int64) sdk.Coin {
	return sdk.NewCoin(denom, sdkmath.NewInt(amount))
}

// NewCoinFromInt creates a sdk.Coin from a math.Int.
func NewCoinFromInt(denom string, amount sdkmath.Int) sdk.Coin {
	return sdk.NewCoin(denom, amount)
}

// ZeroCoin returns a zero-value coin (empty denom, zero amount).
// Use this instead of nil when testing empty coin values.
func ZeroCoin() sdk.Coin {
	return sdk.Coin{}
}

// NewCoins creates a sdk.Coins slice.
func NewCoins(coins ...sdk.Coin) sdk.Coins {
	return sdk.NewCoins(coins...)
}

// -----------------------------------------------------------------------------
// Duration Helpers
// -----------------------------------------------------------------------------

// These are just aliases for clarity - durations don't need special handling
// since gogoproto.stdduration=true converts to time.Duration directly.

// Hours returns a duration in hours.
func Hours(h int) time.Duration {
	return time.Duration(h) * time.Hour
}

// Minutes returns a duration in minutes.
func Minutes(m int) time.Duration {
	return time.Duration(m) * time.Minute
}

// Seconds returns a duration in seconds.
func Seconds(s int) time.Duration {
	return time.Duration(s) * time.Second
}

// Days returns a duration in days.
func Days(d int) time.Duration {
	return time.Duration(d) * 24 * time.Hour
}
