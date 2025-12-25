// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/identity/types"
)

type IdentityCLISuite struct {
	suite.Suite
}

func TestIdentityCLISuite(t *testing.T) {
	suite.Run(t, new(IdentityCLISuite))
}

func (s *IdentityCLISuite) TestGetTxCmd() {
	cmd := GetTxCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal(types.ModuleName, cmd.Use)
	require.True(cmd.DisableFlagParsing)
}

func (s *IdentityCLISuite) TestGetQueryCmd() {
	cmd := GetQueryCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal(types.ModuleName, cmd.Use)
	require.True(cmd.DisableFlagParsing)
}
