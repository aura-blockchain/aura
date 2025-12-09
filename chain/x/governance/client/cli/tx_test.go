package cli

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TxTestSuite struct {
	suite.Suite
}

func TestGovernanceTxTestSuite(t *testing.T) {
	suite.Run(t, new(TxTestSuite))
}

func (s *TxTestSuite) TestGetTxCmd() {
	cmd := GetTxCmd()
	require := s.Require()

	require.NotNil(cmd)
	require.Equal("governance", cmd.Use)
	require.True(cmd.DisableFlagParsing)

	expected := []string{
		"submit-proposal",
		"deposit",
		"vote",
		"vote-weighted",
		"reveal-secret-vote",
		"delegate-vote",
		"undelegate-vote",
		"submit-veto",
		"cosign-veto",
		"execute-proposal",
		"submit-snapshot-vote",
	}

	subCmds := cmd.Commands()
	nameSet := make(map[string]bool, len(subCmds))
	for _, c := range subCmds {
		nameSet[c.Name()] = true
	}

	for _, name := range expected {
		require.True(nameSet[name], "expected command %s to be registered", name)
	}
}

func (s *TxTestSuite) TestCmdSubmitProposalArgs() {
	cmd := CmdSubmitProposal()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"title", "desc", "text"}))
	require.Error(cmd.ValidateArgs([]string{"title"}))
}

func (s *TxTestSuite) TestCmdVoteArgs() {
	cmd := CmdVote()
	require := s.Require()

	require.NoError(cmd.ValidateArgs([]string{"1", "yes"}))
	require.Error(cmd.ValidateArgs([]string{"1"}))
}

func (s *TxTestSuite) TestParseProposalCategory() {
	valid := map[string]string{
		"text":             "text",
		"param-change":     "parameter-change",
		"software-upgrade": "software-upgrade",
		"spend":            "spending",
		"emergency":        "emergency",
		"constitution":     "constitution",
	}

	for input := range valid {
		_, err := parseProposalCategory(input)
		s.Require().NoError(err, "category %s should be valid", input)
	}

	_, err := parseProposalCategory("invalid-category")
	s.Require().Error(err)
}

func (s *TxTestSuite) TestParseVoteOption() {
	opts := []string{"yes", "no", "abstain", "no-with-veto", "y", "n", "a", "veto"}
	for _, opt := range opts {
		_, err := parseVoteOption(opt)
		s.Require().NoError(err, "option %s should be valid", opt)
	}

	_, err := parseVoteOption("maybe")
	s.Require().Error(err)
}

func (s *TxTestSuite) TestParseWeightedVoteOptions() {
	weights, err := parseWeightedVoteOptions("yes=0.6,no=0.4")
	require := s.Require()
	require.NoError(err)
	require.Len(weights, 2)

	_, err = parseWeightedVoteOptions("yes=0.5,no=0.25")
	require.Error(err, "weights must sum to 1.0")

	_, err = parseWeightedVoteOptions("badformat")
	require.Error(err)
}
