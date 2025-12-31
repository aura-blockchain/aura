// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringList_Marshal(t *testing.T) {
	tests := []struct {
		name   string
		list   *StringList
		hasErr bool
	}{
		{
			name:   "nil list",
			list:   nil,
			hasErr: false,
		},
		{
			name:   "empty list",
			list:   &StringList{Values: []string{}},
			hasErr: false,
		},
		{
			name:   "single value",
			list:   &StringList{Values: []string{"hello"}},
			hasErr: false,
		},
		{
			name:   "multiple values",
			list:   &StringList{Values: []string{"hello", "world", "test"}},
			hasErr: false,
		},
		{
			name:   "empty strings",
			list:   &StringList{Values: []string{"", "", ""}},
			hasErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.list.Marshal()
			if tt.hasErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.list == nil {
					require.Nil(t, data)
				} else {
					require.NotNil(t, data)
				}
			}
		})
	}
}

func TestStringList_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected []string
		hasErr   bool
	}{
		{
			name:     "empty data",
			data:     []byte{},
			expected: []string{},
			hasErr:   false,
		},
		{
			name:     "too short data",
			data:     []byte{0, 0},
			expected: []string{},
			hasErr:   false,
		},
		{
			name:     "zero count",
			data:     []byte{0, 0, 0, 0},
			expected: []string{},
			hasErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := &StringList{}
			err := list.Unmarshal(tt.data)
			if tt.hasErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, list.Values)
			}
		})
	}
}

func TestStringList_MarshalUnmarshalRoundtrip(t *testing.T) {
	tests := []struct {
		name   string
		values []string
	}{
		{
			name:   "empty list",
			values: []string{},
		},
		{
			name:   "single value",
			values: []string{"hello"},
		},
		{
			name:   "multiple values",
			values: []string{"hello", "world", "test", "foo", "bar"},
		},
		{
			name:   "with empty strings",
			values: []string{"", "a", "", "b", ""},
		},
		{
			name:   "unicode values",
			values: []string{"hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &StringList{Values: tt.values}
			data, err := original.Marshal()
			require.NoError(t, err)

			restored := &StringList{}
			err = restored.Unmarshal(data)
			require.NoError(t, err)

			require.Equal(t, original.Values, restored.Values)
		})
	}
}

func TestStringList_Unmarshal_Errors(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		hasErr bool
	}{
		{
			name: "truncated string length",
			// count=1 but not enough bytes for string length
			data:   []byte{0, 0, 0, 1, 0},
			hasErr: true,
		},
		{
			name: "truncated string data",
			// count=1, strlen=5 but only 2 bytes of data
			data:   []byte{0, 0, 0, 1, 0, 0, 0, 5, 'a', 'b'},
			hasErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := &StringList{}
			err := list.Unmarshal(tt.data)
			if tt.hasErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStringList_Reset(t *testing.T) {
	list := &StringList{Values: []string{"a", "b", "c"}}
	require.Len(t, list.Values, 3)

	list.Reset()
	require.Len(t, list.Values, 0)
}

func TestStringList_String(t *testing.T) {
	tests := []struct {
		name string
		list *StringList
	}{
		{
			name: "nil list",
			list: nil,
		},
		{
			name: "empty list",
			list: &StringList{Values: []string{}},
		},
		{
			name: "with values",
			list: &StringList{Values: []string{"a", "b"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.list.String()
			require.NotEmpty(t, result)
			require.Contains(t, result, "StringList")
		})
	}
}

func TestStringList_ProtoMessage(t *testing.T) {
	list := &StringList{}
	// Just verify it doesn't panic
	list.ProtoMessage()
}
