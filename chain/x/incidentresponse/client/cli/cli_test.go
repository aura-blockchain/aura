// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type IncidentResponseCLISuite struct {
	suite.Suite
}

func TestIncidentResponseCLISuite(t *testing.T) {
	suite.Run(t, new(IncidentResponseCLISuite))
}

func (s *IncidentResponseCLISuite) TestGetTxCmd() {
	cmd := GetTxCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("incidentresponse", cmd.Use)
	require.True(cmd.DisableFlagParsing)

	expected := []string{
		"report-incident",
		"update-status",
		"request-pause",
		"resume",
		"set-wallet-limits",
		"create-postmortem",
		"close",
		"trigger-backup",
		"trigger-insurance-claim",
	}

	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, name := range expected {
		require.True(names[name], "expected tx command %s to be registered", name)
	}

	// Validate basic arg count for a few commands.
	require.NoError(GetCmdReportIncident().ValidateArgs([]string{"title", "desc", "critical", "api"}))
	require.Error(GetCmdReportIncident().ValidateArgs([]string{"title"}))

	require.NoError(GetCmdResumeChain().ValidateArgs([]string{"ok"}))
	require.Error(GetCmdResumeChain().ValidateArgs([]string{}))
}

func (s *IncidentResponseCLISuite) TestGetQueryCmd() {
	cmd := GetQueryCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("incidentresponse", cmd.Use)
	require.True(cmd.DisableFlagParsing)

	expected := []string{
		"incident",
		"incidents",
		"pause-state",
		"wallet-limits",
		"params",
	}
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, name := range expected {
		require.True(names[name], "expected query command %s to be registered", name)
	}

	require.NoError(GetCmdQueryIncident().ValidateArgs([]string{"INC-1"}))
	require.Error(GetCmdQueryIncident().ValidateArgs([]string{}))
}
