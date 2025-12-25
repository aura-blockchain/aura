// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTxCmdStructure(t *testing.T) {
	cmd := NewTxCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "aiassistant", cmd.Use)
	require.True(t, cmd.DisableFlagParsing)

	// Ensure all expected transaction subcommands exist
	expected := map[string]bool{
		"register [assistant-address]":           false,
		"update-locales [assistant-address]":     false,
		"heartbeat [assistant-address]":          false,
		"report-misbehavior [assistant-address]": false,
	}

	for _, sub := range cmd.Commands() {
		if _, ok := expected[sub.Use]; ok {
			expected[sub.Use] = true
		}
	}

	for use, seen := range expected {
		require.Truef(t, seen, "missing tx subcommand: %s", use)
	}
}

func TestRegisterCmdArgs(t *testing.T) {
	cmd := newRegisterCmd()
	require.Equal(t, "register [assistant-address]", cmd.Use)

	err := cmd.ValidateArgs([]string{"aura1assistant"})
	require.NoError(t, err)

	err = cmd.ValidateArgs([]string{})
	require.Error(t, err)

	err = cmd.ValidateArgs([]string{"aura1assistant", "extra"})
	require.Error(t, err)
}

func TestUpdateLocalesCmdArgs(t *testing.T) {
	cmd := newUpdateLocalesCmd()

	err := cmd.ValidateArgs([]string{"aura1assistant"})
	require.NoError(t, err)

	err = cmd.ValidateArgs([]string{})
	require.Error(t, err)
}

func TestHeartbeatCmdArgs(t *testing.T) {
	cmd := newHeartbeatCmd()

	err := cmd.ValidateArgs([]string{"aura1assistant"})
	require.NoError(t, err)

	err = cmd.ValidateArgs([]string{})
	require.Error(t, err)
}

func TestReportMisbehaviorCmdArgs(t *testing.T) {
	cmd := newReportMisbehaviorCmd()

	err := cmd.ValidateArgs([]string{"aura1assistant"})
	require.NoError(t, err)

	err = cmd.ValidateArgs([]string{})
	require.Error(t, err)
}

func TestSplitAndNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "normal locales",
			input:    "en-US, es-ES ,EN-us",
			expected: []string{"en-us", "es-es"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    " , ",
			expected: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := splitAndNormalize(tc.input)
			require.Equal(t, tc.expected, out)
		})
	}
}
