package keeper

import (
	"context"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/common/determinism"
)

// timestampToGogo converts a google protobuf Timestamp to gogoproto Timestamp
func timestampToGogo(ts *timestamppb.Timestamp) *gogotypes.Timestamp {
	if ts == nil {
		return nil
	}
	return &gogotypes.Timestamp{
		Seconds: ts.Seconds,
		Nanos:   ts.Nanos,
	}
}

// gogoTimestampNow returns the current time as a gogoproto Timestamp
//
// DEPRECATED: DO NOT USE IN PRODUCTION KEEPER CODE
// This function uses time.Now() which causes non-determinism in blockchain state.
// Each validator would get different wall-clock times, breaking consensus.
//
// For production keeper methods, use blockTimeToGogoTimestamp(ctx) instead.
// This function should ONLY be used in test code where determinism is not required.
func gogoTimestampNow() *gogotypes.Timestamp {
	now := time.Now()
	return &gogotypes.Timestamp{
		Seconds: now.Unix(),
		Nanos:   int32(now.Nanosecond()),
	}
}

// timeToGogoTimestamp converts a Go time.Time to gogoproto Timestamp
func timeToGogoTimestamp(t time.Time) *gogotypes.Timestamp {
	return &gogotypes.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

// blockTimeToGogoTimestamp gets the block time from context and converts to gogoproto Timestamp
func blockTimeToGogoTimestamp(ctx context.Context) *gogotypes.Timestamp {
	t := determinism.GetBlockTime(ctx)
	return &gogotypes.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

// blockTimeWithOffsetToGogoTimestamp gets the block time from context, adds a duration, and converts to gogoproto Timestamp
func blockTimeWithOffsetToGogoTimestamp(ctx context.Context, offset time.Duration) *gogotypes.Timestamp {
	t := determinism.GetBlockTime(ctx).Add(offset)
	return &gogotypes.Timestamp{
		Seconds: t.Unix(),
		Nanos:   int32(t.Nanosecond()),
	}
}

// gogoTimestampToTime converts a gogoproto Timestamp to time.Time
func gogoTimestampToTime(ts *gogotypes.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return time.Unix(ts.Seconds, int64(ts.Nanos))
}

// gogoDurationToTime converts a gogoproto Duration to time.Duration
func gogoDurationToTime(d *gogotypes.Duration) time.Duration {
	if d == nil {
		return 0
	}
	return time.Duration(d.Seconds)*time.Second + time.Duration(d.Nanos)*time.Nanosecond
}
