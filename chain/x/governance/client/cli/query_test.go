// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type QueryTestSuite struct {
	suite.Suite
}

func TestGovernanceQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}

func (s *QueryTestSuite) TestGetQueryCmd() {
	cmd := GetQueryCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("governance", cmd.Use)

	expected := []string{
		"proposal",
		"proposals",
		"vote",
		"votes",
		"deposit",
		"deposits",
		"tally",
		"params",
		"vote-delegations",
		"voting-power",
		"token-locks",
		"veto-requests",
		"snapshot-votes",
	}

	subCmds := cmd.Commands()
	nameSet := make(map[string]bool, len(subCmds))
	for _, c := range subCmds {
		nameSet[c.Name()] = true
	}

	for _, name := range expected {
		require.True(nameSet[name], "expected query command %s to be registered", name)
	}
}

func (s *QueryTestSuite) TestCmdQueryProposalArgs() {
	cmd := CmdQueryProposal()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"1"}))
	require.Error(cmd.ValidateArgs([]string{}))
}

func (s *QueryTestSuite) TestCmdQueryProposalsArgs() {
	cmd := CmdQueryProposals()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{}))
	require.Error(cmd.ValidateArgs([]string{"unexpected"}))
}
