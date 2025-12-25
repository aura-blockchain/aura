// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTxCmd(t *testing.T) {
	cmd := GetTxCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "prevalidation", cmd.Use)
	require.Contains(t, cmd.Short, "Pre-validation")
	require.True(t, cmd.DisableFlagParsing)

	// Prevalidation module has no user-facing transaction commands
	// as it operates automatically
	subcommands := cmd.Commands()
	require.Len(t, subcommands, 0, "prevalidation should have no tx subcommands")
}

func TestTxCommandStructure(t *testing.T) {
	cmd := GetTxCmd()

	// Test command properties
	require.Equal(t, "prevalidation", cmd.Use)
	require.NotEmpty(t, cmd.Short)
	require.True(t, cmd.DisableFlagParsing)
	require.Equal(t, 2, cmd.SuggestionsMinimumDistance)
}

func TestPrevalidationIsAutomated(t *testing.T) {
	// This test documents that prevalidation is an automated system
	// with no user-facing transactions
	cmd := GetTxCmd()

	// Should have no subcommands
	require.Empty(t, cmd.Commands())

	// Command should exist but be informational only
	require.NotNil(t, cmd)
	require.NotEmpty(t, cmd.Use)
}
