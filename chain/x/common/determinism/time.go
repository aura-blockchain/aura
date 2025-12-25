// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package determinism

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetBlockTime returns the deterministic block time from context
// This should ALWAYS be used instead of time.Now() in keeper code
func GetBlockTime(ctx context.Context) time.Time {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.BlockTime()
}

// GetBlockHeight returns the current block height
func GetBlockHeight(ctx context.Context) int64 {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.BlockHeight()
}

// GetBlockTimestamp returns block time as Unix timestamp
func GetBlockTimestamp(ctx context.Context) int64 {
	return GetBlockTime(ctx).Unix()
}

// TimeSince returns deterministic duration since a past time
// Uses block time instead of time.Now()
func TimeSince(ctx context.Context, past time.Time) time.Duration {
	return GetBlockTime(ctx).Sub(past)
}

// TimeUntil returns deterministic duration until a future time
// Uses block time instead of time.Now()
func TimeUntil(ctx context.Context, future time.Time) time.Duration {
	return future.Sub(GetBlockTime(ctx))
}

// IsExpired checks if a deadline has passed using block time
func IsExpired(ctx context.Context, deadline time.Time) bool {
	return GetBlockTime(ctx).After(deadline)
}

// IsNotYetValid checks if a start time hasn't been reached
func IsNotYetValid(ctx context.Context, startTime time.Time) bool {
	return GetBlockTime(ctx).Before(startTime)
}

// IsInWindow checks if current block time is within a time window
func IsInWindow(ctx context.Context, start, end time.Time) bool {
	blockTime := GetBlockTime(ctx)
	return !blockTime.Before(start) && !blockTime.After(end)
}

// AddDuration adds a duration to block time
func AddDuration(ctx context.Context, duration time.Duration) time.Time {
	return GetBlockTime(ctx).Add(duration)
}

// GetDayTimestamp returns the Unix timestamp for start of the day
// containing the current block time (useful for daily rate limits)
func GetDayTimestamp(ctx context.Context) int64 {
	blockTime := GetBlockTime(ctx)
	year, month, day := blockTime.Date()
	dayStart := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return dayStart.Unix()
}

// GetHourTimestamp returns the Unix timestamp for start of the hour
func GetHourTimestamp(ctx context.Context) int64 {
	blockTime := GetBlockTime(ctx)
	hourStart := blockTime.Truncate(time.Hour)
	return hourStart.Unix()
}

// FormatBlockTime returns formatted block time string
func FormatBlockTime(ctx context.Context, layout string) string {
	return GetBlockTime(ctx).Format(layout)
}

// ValidateTimeWindow validates that end > start
func ValidateTimeWindow(start, end time.Time) error {
	if !end.After(start) {
		return fmt.Errorf("end time must be after start time: start=%s, end=%s", start, end)
	}
	return nil
}

// CalculateDeadline calculates a deadline from current block time
func CalculateDeadline(ctx context.Context, duration time.Duration) time.Time {
	return GetBlockTime(ctx).Add(duration)
}

// TimeRange represents a deterministic time range
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// NewTimeRange creates a validated time range
func NewTimeRange(start, end time.Time) (*TimeRange, error) {
	if err := ValidateTimeWindow(start, end); err != nil {
		return nil, err
	}
	return &TimeRange{Start: start, End: end}, nil
}

// Contains checks if a time falls within the range
func (tr TimeRange) Contains(t time.Time) bool {
	return !t.Before(tr.Start) && !t.After(tr.End)
}

// Duration returns the duration of the range
func (tr TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}

// IsActive checks if the range is currently active
func (tr TimeRange) IsActive(ctx context.Context) bool {
	return tr.Contains(GetBlockTime(ctx))
}

// DeterministicTimer provides deterministic time-based operations
type DeterministicTimer struct {
	startTime time.Time
	duration  time.Duration
}

// NewDeterministicTimer creates a timer based on block time
func NewDeterministicTimer(ctx context.Context, duration time.Duration) *DeterministicTimer {
	return &DeterministicTimer{
		startTime: GetBlockTime(ctx),
		duration:  duration,
	}
}

// IsExpired checks if the timer has expired
func (dt *DeterministicTimer) IsExpired(ctx context.Context) bool {
	return GetBlockTime(ctx).After(dt.startTime.Add(dt.duration))
}

// RemainingTime returns remaining time until expiration
func (dt *DeterministicTimer) RemainingTime(ctx context.Context) time.Duration {
	deadline := dt.startTime.Add(dt.duration)
	remaining := deadline.Sub(GetBlockTime(ctx))
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ElapsedTime returns time elapsed since timer start
func (dt *DeterministicTimer) ElapsedTime(ctx context.Context) time.Duration {
	return GetBlockTime(ctx).Sub(dt.startTime)
}

// Progress returns progress as a value between 0.0 and 1.0
func (dt *DeterministicTimer) Progress(ctx context.Context) float64 {
	elapsed := dt.ElapsedTime(ctx)
	if elapsed >= dt.duration {
		return 1.0
	}
	return float64(elapsed) / float64(dt.duration)
}
