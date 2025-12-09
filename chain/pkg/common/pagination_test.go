package common_test

import (
	"testing"

	"github.com/aequitas/aura/chain/pkg/common"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
)

func TestNormalizePagination(t *testing.T) {
	tests := []struct {
		name           string
		input          *query.PageRequest
		expectedLimit  uint64
		expectedOffset uint64
		expectedKey    []byte
	}{
		{
			name:          "nil pagination",
			input:         nil,
			expectedLimit: common.DefaultPageLimit,
		},
		{
			name:          "zero limit",
			input:         &query.PageRequest{Limit: 0},
			expectedLimit: common.DefaultPageLimit,
		},
		{
			name:          "custom limit",
			input:         &query.PageRequest{Limit: 50},
			expectedLimit: 50,
		},
		{
			name:          "max limit",
			input:         &query.PageRequest{Limit: common.MaxPageLimit},
			expectedLimit: common.MaxPageLimit,
		},
		{
			name:          "above max limit (SDK will enforce)",
			input:         &query.PageRequest{Limit: 2000},
			expectedLimit: 2000, // NormalizePagination doesn't cap, SDK does
		},
		{
			name:           "with offset",
			input:          &query.PageRequest{Limit: 100, Offset: 50},
			expectedLimit:  100,
			expectedOffset: 50,
		},
		{
			name:          "with key",
			input:         &query.PageRequest{Limit: 100, Key: []byte("nextkey")},
			expectedLimit: 100,
			expectedKey:   []byte("nextkey"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.NormalizePagination(tt.input)

			require.NotNil(t, result)
			require.Equal(t, tt.expectedLimit, result.Limit)

			if tt.expectedOffset > 0 {
				require.Equal(t, tt.expectedOffset, result.Offset)
			}

			if tt.expectedKey != nil {
				require.Equal(t, tt.expectedKey, result.Key)
			}
		})
	}
}

func TestGetEffectiveLimit(t *testing.T) {
	tests := []struct {
		name          string
		input         *query.PageRequest
		expectedLimit uint64
	}{
		{
			name:          "nil pagination",
			input:         nil,
			expectedLimit: common.DefaultPageLimit,
		},
		{
			name:          "zero limit",
			input:         &query.PageRequest{Limit: 0},
			expectedLimit: common.DefaultPageLimit,
		},
		{
			name:          "custom limit",
			input:         &query.PageRequest{Limit: 50},
			expectedLimit: 50,
		},
		{
			name:          "max limit",
			input:         &query.PageRequest{Limit: common.MaxPageLimit},
			expectedLimit: common.MaxPageLimit,
		},
		{
			name:          "above max limit",
			input:         &query.PageRequest{Limit: 2000},
			expectedLimit: common.MaxPageLimit, // GetEffectiveLimit does cap
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.GetEffectiveLimit(tt.input)
			require.Equal(t, tt.expectedLimit, result)
		})
	}
}

func TestDefaultPageLimit(t *testing.T) {
	require.Equal(t, uint64(100), uint64(common.DefaultPageLimit))
}

func TestMaxPageLimit(t *testing.T) {
	require.Equal(t, uint64(1000), uint64(common.MaxPageLimit))
}
