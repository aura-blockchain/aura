// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"math"
	"testing"
)

func TestClampInt64ToInt32(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int32
	}{
		// Boundary cases - at limits
		{"max int32", math.MaxInt32, math.MaxInt32},
		{"min int32", math.MinInt32, math.MinInt32},

		// Beyond limits - should clamp
		{"above max int32", math.MaxInt32 + 1, math.MaxInt32},
		{"below min int32", math.MinInt32 - 1, math.MinInt32},
		{"max int64", math.MaxInt64, math.MaxInt32},
		{"min int64", math.MinInt64, math.MinInt32},
		{"large positive", int64(math.MaxInt32) * 2, math.MaxInt32},
		{"large negative", int64(math.MinInt32) * 2, math.MinInt32},

		// Normal values - no clamping needed
		{"zero", 0, 0},
		{"positive one", 1, 1},
		{"negative one", -1, -1},
		{"positive medium", 1000000, 1000000},
		{"negative medium", -1000000, -1000000},
		{"max int32 minus 1", math.MaxInt32 - 1, math.MaxInt32 - 1},
		{"min int32 plus 1", math.MinInt32 + 1, math.MinInt32 + 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampInt64ToInt32(tt.input)
			if result != tt.expected {
				t.Errorf("clampInt64ToInt32(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestClampNanosecondsToInt32(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int32
	}{
		// Boundary cases - at limits
		{"max int32", math.MaxInt32, math.MaxInt32},
		{"min int32", math.MinInt32, math.MinInt32},

		// Beyond limits - should clamp
		{"above max int32", math.MaxInt32 + 1, math.MaxInt32},
		{"below min int32", math.MinInt32 - 1, math.MinInt32},
		{"max int64", math.MaxInt64, math.MaxInt32},
		{"min int64", math.MinInt64, math.MinInt32},

		// Typical nanosecond values
		{"zero nanoseconds", 0, 0},
		{"one nanosecond", 1, 1},
		{"one microsecond in ns", 1000, 1000},
		{"one millisecond in ns", 1000000, 1000000},
		{"999 milliseconds in ns", 999999999, 999999999},

		// Edge cases for nanosecond representation
		{"max valid ns (just under 1 second)", 999999999, 999999999},
		{"negative nanoseconds", -500000000, -500000000},

		// Values that exceed int32 range (over ~2.1 seconds in ns)
		{"over 2 seconds in ns", 2500000000, math.MaxInt32},
		{"over 3 seconds in ns", 3000000000, math.MaxInt32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampNanosecondsToInt32(tt.input)
			if result != tt.expected {
				t.Errorf("clampNanosecondsToInt32(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
