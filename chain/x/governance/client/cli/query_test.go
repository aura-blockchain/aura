// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
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
	require.Equal("Querying commands for the governance module", cmd.Short)
	require.True(cmd.DisableFlagParsing)
	require.Equal(2, cmd.SuggestionsMinimumDistance)

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

// ============================================================================
// Proposal Query Tests
// ============================================================================

func TestCmdQueryProposal(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid proposal ID",
			args:    []string{"1"},
			wantErr: false,
		},
		{
			name:    "valid large proposal ID",
			args:    []string{"999999"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryProposal()
			require.NotNil(t, cmd)
			require.Equal(t, "proposal [proposal-id]", cmd.Use)
			require.Contains(t, cmd.Short, "Query a proposal")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdQueryProposals(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args - valid",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "extra args - should fail",
			args:    []string{"unexpected"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryProposals()
			require.NotNil(t, cmd)
			require.Equal(t, "proposals", cmd.Use)
			require.Contains(t, cmd.Short, "Query all governance proposals")

			// Verify flags exist
			statusFlag := cmd.Flags().Lookup("status")
			require.NotNil(t, statusFlag)

			voterFlag := cmd.Flags().Lookup("voter")
			require.NotNil(t, voterFlag)

			depositorFlag := cmd.Flags().Lookup("depositor")
			require.NotNil(t, depositorFlag)

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdQueryProposalsFlags(t *testing.T) {
	cmd := CmdQueryProposals()

	// Test status flag values
	statusFlag := cmd.Flags().Lookup("status")
	require.NotNil(t, statusFlag)
	require.Equal(t, "", statusFlag.DefValue)

	// Test voter flag
	voterFlag := cmd.Flags().Lookup("voter")
	require.NotNil(t, voterFlag)
	require.Equal(t, "", voterFlag.DefValue)

	// Test depositor flag
	depositorFlag := cmd.Flags().Lookup("depositor")
	require.NotNil(t, depositorFlag)
	require.Equal(t, "", depositorFlag.DefValue)
}

// ============================================================================
// Vote Query Tests
// ============================================================================

func TestCmdQueryVote(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid proposal ID and voter",
			args:    []string{"1", "aura1abc123"},
			wantErr: false,
		},
		{
			name:    "valid with cosmos address format",
			args:    []string{"5", "aura1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"},
			wantErr: false,
		},
		{
			name:    "missing voter - should fail",
			args:    []string{"1"},
			wantErr: true,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"1", "voter", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryVote()
			require.NotNil(t, cmd)
			require.Equal(t, "vote [proposal-id] [voter]", cmd.Use)
			require.Contains(t, cmd.Short, "Query a vote")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdQueryVotes(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid proposal ID",
			args:    []string{"1"},
			wantErr: false,
		},
		{
			name:    "valid large proposal ID",
			args:    []string{"12345"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryVotes()
			require.NotNil(t, cmd)
			require.Equal(t, "votes [proposal-id]", cmd.Use)
			require.Contains(t, cmd.Short, "Query all votes")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Deposit Query Tests
// ============================================================================

func TestCmdQueryDeposit(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid proposal ID and depositor",
			args:    []string{"1", "aura1depositor123"},
			wantErr: false,
		},
		{
			name:    "missing depositor - should fail",
			args:    []string{"1"},
			wantErr: true,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"1", "depositor", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryDeposit()
			require.NotNil(t, cmd)
			require.Equal(t, "deposit [proposal-id] [depositor]", cmd.Use)
			require.Contains(t, cmd.Short, "Query a deposit")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdQueryDeposits(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid proposal ID",
			args:    []string{"1"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryDeposits()
			require.NotNil(t, cmd)
			require.Equal(t, "deposits [proposal-id]", cmd.Use)
			require.Contains(t, cmd.Short, "Query all deposits")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Tally Query Tests
// ============================================================================

func TestCmdQueryTallyResult(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid proposal ID",
			args:    []string{"1"},
			wantErr: false,
		},
		{
			name:    "valid large proposal ID",
			args:    []string{"999"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryTallyResult()
			require.NotNil(t, cmd)
			require.Equal(t, "tally [proposal-id]", cmd.Use)
			require.Contains(t, cmd.Short, "tally")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Params Query Tests
// ============================================================================

func TestCmdQueryParams(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args - valid",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "extra args - should fail",
			args:    []string{"extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryParams()
			require.NotNil(t, cmd)
			require.Equal(t, "params", cmd.Use)
			require.Contains(t, cmd.Short, "governance parameters")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Vote Delegation Query Tests
// ============================================================================

func TestCmdQueryVoteDelegations(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid delegator address",
			args:    []string{"aura1delegator123"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"addr1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryVoteDelegations()
			require.NotNil(t, cmd)
			require.Equal(t, "vote-delegations [delegator]", cmd.Use)
			require.Contains(t, cmd.Short, "vote delegations")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCmdQueryVotingPower(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid voter address",
			args:    []string{"aura1voter123"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"addr1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryVotingPower()
			require.NotNil(t, cmd)
			require.Equal(t, "voting-power [address]", cmd.Use)
			require.Contains(t, cmd.Short, "voting power")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Token Lock Query Tests
// ============================================================================

func TestCmdQueryTokenLocks(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid address",
			args:    []string{"aura1owner123"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"addr1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryTokenLocks()
			require.NotNil(t, cmd)
			require.Equal(t, "token-locks [address]", cmd.Use)
			require.Contains(t, cmd.Short, "token locks")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Veto Query Tests
// ============================================================================

func TestCmdQueryVetoRequests(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid proposal ID",
			args:    []string{"1"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQueryVetoRequests()
			require.NotNil(t, cmd)
			require.Equal(t, "veto-requests [proposal-id]", cmd.Use)
			require.Contains(t, cmd.Short, "veto requests")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Snapshot Voting Query Tests
// ============================================================================

func TestCmdQuerySnapshotVotes(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "valid proposal ID",
			args:    []string{"1"},
			wantErr: false,
		},
		{
			name:    "valid large proposal ID",
			args:    []string{"99999"},
			wantErr: false,
		},
		{
			name:    "no args - should fail",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "too many args - should fail",
			args:    []string{"1", "extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := CmdQuerySnapshotVotes()
			require.NotNil(t, cmd)
			require.Equal(t, "snapshot-votes [proposal-id]", cmd.Use)
			require.Contains(t, cmd.Short, "snapshot votes")

			err := cmd.Args(cmd, tt.args)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Command Configuration Tests
// ============================================================================

func TestQueryCommandsHaveQueryFlags(t *testing.T) {
	commands := []*struct {
		name string
		fn   func() *cobra.Command
	}{
		{"proposal", CmdQueryProposal},
		{"proposals", CmdQueryProposals},
		{"vote", CmdQueryVote},
		{"votes", CmdQueryVotes},
		{"deposit", CmdQueryDeposit},
		{"deposits", CmdQueryDeposits},
		{"tally", CmdQueryTallyResult},
		{"params", CmdQueryParams},
		{"vote-delegations", CmdQueryVoteDelegations},
		{"voting-power", CmdQueryVotingPower},
		{"token-locks", CmdQueryTokenLocks},
		{"veto-requests", CmdQueryVetoRequests},
		{"snapshot-votes", CmdQuerySnapshotVotes},
	}

	for _, tt := range commands {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.fn()
			require.NotNil(t, cmd)

			// All query commands should have node flag
			nodeFlag := cmd.Flags().Lookup("node")
			require.NotNil(t, nodeFlag, "command %s should have --node flag", tt.name)

			// All query commands should have output flag
			outputFlag := cmd.Flags().Lookup("output")
			require.NotNil(t, outputFlag, "command %s should have --output flag", tt.name)
		})
	}
}

func TestQueryCommandsHaveDescriptions(t *testing.T) {
	commands := []*struct {
		name string
		fn   func() *cobra.Command
	}{
		{"proposal", CmdQueryProposal},
		{"proposals", CmdQueryProposals},
		{"vote", CmdQueryVote},
		{"votes", CmdQueryVotes},
		{"deposit", CmdQueryDeposit},
		{"deposits", CmdQueryDeposits},
		{"tally", CmdQueryTallyResult},
		{"params", CmdQueryParams},
		{"vote-delegations", CmdQueryVoteDelegations},
		{"voting-power", CmdQueryVotingPower},
		{"token-locks", CmdQueryTokenLocks},
		{"veto-requests", CmdQueryVetoRequests},
		{"snapshot-votes", CmdQuerySnapshotVotes},
	}

	for _, tt := range commands {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.fn()
			require.NotNil(t, cmd)
			require.NotEmpty(t, cmd.Short, "command %s should have short description", tt.name)
			require.NotEmpty(t, cmd.Long, "command %s should have long description", tt.name)
		})
	}
}

// ============================================================================
// Suite Method Tests
// ============================================================================

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

func (s *QueryTestSuite) TestCmdQueryVoteArgs() {
	cmd := CmdQueryVote()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"1", "aura1voter"}))
	require.Error(cmd.ValidateArgs([]string{"1"}))
	require.Error(cmd.ValidateArgs([]string{}))
}

func (s *QueryTestSuite) TestCmdQueryVotesArgs() {
	cmd := CmdQueryVotes()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"1"}))
	require.Error(cmd.ValidateArgs([]string{}))
}

func (s *QueryTestSuite) TestCmdQueryDepositArgs() {
	cmd := CmdQueryDeposit()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"1", "aura1depositor"}))
	require.Error(cmd.ValidateArgs([]string{"1"}))
}

func (s *QueryTestSuite) TestCmdQueryDepositsArgs() {
	cmd := CmdQueryDeposits()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"1"}))
	require.Error(cmd.ValidateArgs([]string{}))
}

func (s *QueryTestSuite) TestCmdQueryTallyResultArgs() {
	cmd := CmdQueryTallyResult()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"1"}))
	require.Error(cmd.ValidateArgs([]string{}))
}

func (s *QueryTestSuite) TestCmdQueryParamsArgs() {
	cmd := CmdQueryParams()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{}))
	require.Error(cmd.ValidateArgs([]string{"unexpected"}))
}

func (s *QueryTestSuite) TestCmdQueryVoteDelegationsArgs() {
	cmd := CmdQueryVoteDelegations()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"aura1delegator"}))
	require.Error(cmd.ValidateArgs([]string{}))
}

func (s *QueryTestSuite) TestCmdQueryVotingPowerArgs() {
	cmd := CmdQueryVotingPower()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"aura1voter"}))
	require.Error(cmd.ValidateArgs([]string{}))
}

func (s *QueryTestSuite) TestCmdQueryTokenLocksArgs() {
	cmd := CmdQueryTokenLocks()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"aura1owner"}))
	require.Error(cmd.ValidateArgs([]string{}))
}

func (s *QueryTestSuite) TestCmdQueryVetoRequestsArgs() {
	cmd := CmdQueryVetoRequests()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"1"}))
	require.Error(cmd.ValidateArgs([]string{}))
}

func (s *QueryTestSuite) TestCmdQuerySnapshotVotesArgs() {
	cmd := CmdQuerySnapshotVotes()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"1"}))
	require.Error(cmd.ValidateArgs([]string{}))
}
