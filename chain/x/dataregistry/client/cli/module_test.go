// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetModuleRootCmd(t *testing.T) {
	cmd := GetModuleRootCmd()
	require.NotNil(t, cmd)
	require.Equal(t, "dataregistry", cmd.Use)
	require.Equal(t, "Data Registry module commands", cmd.Short)
	require.True(t, cmd.DisableFlagParsing)
	require.Equal(t, 2, cmd.SuggestionsMinimumDistance)
	require.NotNil(t, cmd.RunE)
}
