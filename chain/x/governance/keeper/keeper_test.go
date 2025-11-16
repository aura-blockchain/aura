package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/governance/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubmitProposal(t *testing.T) {
	tests := []struct {
		name           string
		title          string
		description    string
		category       types.ProposalCategory
		proposer       string
		initialDeposit string
		isEmergency    bool
		expectErr      bool
		errContains    string
		expectedStatus types.ProposalStatus
	}{
		{
			name:           "valid text proposal",
			title:          "Test Proposal",
			description:    "This is a test proposal",
			category:       types.CategoryText,
			proposer:       "proposer1",
			initialDeposit: "5000000000",
			isEmergency:    false,
			expectErr:      false,
			expectedStatus: types.StatusVotingPeriod,
		},
		{
			name:           "valid emergency proposal",
			title:          "Emergency Proposal",
			description:    "This is an emergency proposal",
			category:       types.CategoryEmergency,
			proposer:       "proposer1",
			initialDeposit: "50000000000",
			isEmergency:    true,
			expectErr:      false,
			expectedStatus: types.StatusVotingPeriod,
		},
		{
			name:           "insufficient deposit for immediate voting",
			title:          "Test Proposal",
			description:    "This is a test proposal",
			category:       types.CategoryText,
			proposer:       "proposer1",
			initialDeposit: "1000",
			isEmergency:    false,
			expectErr:      false,
			expectedStatus: types.StatusDepositPeriod,
		},
		{
			name:           "empty title",
			title:          "",
			description:    "This is a test proposal",
			category:       types.CategoryText,
			proposer:       "proposer1",
			initialDeposit: "5000000000",
			isEmergency:    false,
			expectErr:      true,
			errContains:    "title cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := NewKeeper(types.DefaultParams())

			proposalID, err := k.SubmitProposal(
				tt.title,
				tt.description,
				tt.category,
				tt.proposer,
				tt.initialDeposit,
				tt.isEmergency,
			)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Greater(t, proposalID, uint64(0))

			// Verify proposal was created
			proposal, err := k.GetProposal(proposalID)
			require.NoError(t, err)
			assert.Equal(t, tt.title, proposal.Title)
			assert.Equal(t, tt.description, proposal.Description)
			assert.Equal(t, tt.category, proposal.Category)
			assert.Equal(t, tt.proposer, proposal.Proposer)

			// Verify status
			assert.Equal(t, tt.expectedStatus, proposal.Status)
		})
	}
}

func TestProposalCategoryThresholds(t *testing.T) {
	k := NewKeeper(types.DefaultParams())
	params := k.GetParams()

	tests := []struct {
		name     string
		category types.ProposalCategory
	}{
		{"text", types.CategoryText},
		{"parameter_change", types.CategoryParameterChange},
		{"software_upgrade", types.CategorySoftwareUpgrade},
		{"spending", types.CategorySpending},
		{"emergency", types.CategoryEmergency},
		{"constitution", types.CategoryConstitution},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categoryParams := params.GetCategoryParams(tt.category)

			// Verify category-specific parameters exist
			assert.NotEmpty(t, categoryParams.MinDeposit)
			assert.Greater(t, categoryParams.VotingPeriod, time.Duration(0))
			assert.NotEmpty(t, categoryParams.Quorum)
			assert.NotEmpty(t, categoryParams.Threshold)

			// Higher categories should have stricter requirements
			if tt.category == types.CategoryConstitution {
				assert.Equal(t, "0.667", categoryParams.Quorum)
				assert.Equal(t, "0.750", categoryParams.Threshold)
			} else if tt.category == types.CategoryEmergency {
				assert.Equal(t, "0.600", categoryParams.Quorum)
				assert.Equal(t, "0.750", categoryParams.Threshold)
			}
		})
	}
}

func TestAddDeposit(t *testing.T) {
	k := NewKeeper(types.DefaultParams())

	// Create a proposal with insufficient initial deposit to stay in deposit period
	proposalID, err := k.SubmitProposal(
		"Test Proposal",
		"Description",
		types.CategoryText,
		"proposer1",
		"1000000000",
		false,
	)
	require.NoError(t, err)

	// Verify proposal is in deposit period
	proposal, err := k.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, types.StatusDepositPeriod, proposal.Status)

	// Add deposit to reach minimum
	err = k.AddDeposit(proposalID, "depositor1", "4000000000")
	require.NoError(t, err)

	// Verify total deposit updated and status changed to voting
	proposal, err = k.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, "5000000000", proposal.TotalDeposit)
	assert.Equal(t, types.StatusVotingPeriod, proposal.Status)
}

func TestCastVote(t *testing.T) {
	k := NewKeeper(types.DefaultParams())

	// Create an emergency proposal (starts voting immediately)
	proposalID, err := k.SubmitProposal(
		"Emergency Proposal",
		"Description",
		types.CategoryEmergency,
		"proposer1",
		"50000000000",
		true,
	)
	require.NoError(t, err)

	tests := []struct {
		name        string
		voter       string
		option      types.VoteOption
		votingPower string
		isSecret    bool
		commitment  string
		expectErr   bool
		errContains string
	}{
		{
			name:        "valid yes vote",
			voter:       "voter1",
			option:      types.OptionYes,
			votingPower: "1000000",
			isSecret:    false,
			commitment:  "",
			expectErr:   false,
		},
		{
			name:        "valid no vote",
			voter:       "voter2",
			option:      types.OptionNo,
			votingPower: "2000000",
			isSecret:    false,
			commitment:  "",
			expectErr:   false,
		},
		{
			name:        "duplicate vote",
			voter:       "voter1",
			option:      types.OptionNo,
			votingPower: "1000000",
			isSecret:    false,
			commitment:  "",
			expectErr:   true,
			errContains: "already voted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k.CastVote(
				proposalID,
				tt.voter,
				tt.option,
				tt.votingPower,
				tt.isSecret,
				tt.commitment,
			)

			if tt.expectErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)

			// Verify vote was recorded
			vote, err := k.GetVote(proposalID, tt.voter)
			require.NoError(t, err)
			assert.Equal(t, tt.option, vote.Option)
			assert.Equal(t, tt.votingPower, vote.VotingPower)
		})
	}
}

func TestSecretBallotVoting(t *testing.T) {
	params := types.DefaultParams()
	params.SecretBallotEnabled = true
	k := NewKeeper(params)

	// Create an emergency proposal
	proposalID, err := k.SubmitProposal(
		"Secret Vote Proposal",
		"Description",
		types.CategoryEmergency,
		"proposer1",
		"50000000000",
		true,
	)
	require.NoError(t, err)

	// Compute commitment
	voter := "voter1"
	option := types.OptionYes
	revealKey := "secret123"
	commitment := computeCommitment(voter, option, revealKey)

	// Cast secret vote
	err = k.CastVote(proposalID, voter, option, "1000000", true, commitment)
	require.NoError(t, err)

	// Verify vote is hidden
	vote, err := k.GetVote(proposalID, voter)
	require.NoError(t, err)
	assert.Equal(t, types.OptionUnspecified, vote.Option)
	assert.True(t, vote.IsSecret)
	assert.Equal(t, commitment, vote.Commitment)

	// Mock end of voting period by modifying proposal
	proposal, _ := k.GetProposal(proposalID)
	proposal.VotingEndTime = time.Now().Add(-1 * time.Hour)
	k.proposals[proposalID] = proposal

	// Reveal vote
	err = k.RevealSecretVote(proposalID, voter, option, revealKey)
	require.NoError(t, err)

	// Verify vote is revealed
	vote, err = k.GetVote(proposalID, voter)
	require.NoError(t, err)
	assert.Equal(t, option, vote.Option)
}

func TestVoteDelegation(t *testing.T) {
	k := NewKeeper(types.DefaultParams())

	delegator := "delegator1"
	delegate := "delegate1"
	votingPower := "1000000"

	// Test delegation
	err := k.DelegateVote(delegator, delegate, votingPower, []types.ProposalCategory{})
	require.NoError(t, err)

	// Test self-delegation (should fail)
	err = k.DelegateVote(delegator, delegator, votingPower, []types.ProposalCategory{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delegate to self")

	// Test category-specific delegation
	err = k.DelegateVote(
		"delegator2",
		"delegate2",
		votingPower,
		[]types.ProposalCategory{types.CategoryText, types.CategoryParameterChange},
	)
	require.NoError(t, err)

	// Test undelegation
	err = k.UndelegateVote(delegator, delegate, []types.ProposalCategory{})
	require.NoError(t, err)

	// Test undelegation of non-existent delegation
	err = k.UndelegateVote(delegator, delegate, []types.ProposalCategory{})
	require.Error(t, err)
	assert.Equal(t, types.ErrDelegationNotFound, err)
}

func TestVetoMechanism(t *testing.T) {
	params := types.DefaultParams()
	params.VetoAuthorizedAddresses = []string{"vetoer1", "vetoer2", "vetoer3", "vetoer4", "vetoer5"}
	params.VetoCosignersRequired = 3
	k := NewKeeper(params)

	// Create a proposal
	proposalID, err := k.SubmitProposal(
		"Veto Test Proposal",
		"Description",
		types.CategoryEmergency,
		"proposer1",
		"50000000000",
		true,
	)
	require.NoError(t, err)

	// Submit veto
	executed, err := k.SubmitVeto(proposalID, "vetoer1", "Security concern")
	require.NoError(t, err)
	assert.False(t, executed) // Should not execute with only 1 cosigner

	// Unauthorized veto should fail
	_, err = k.SubmitVeto(proposalID, "unauthorized", "Reason")
	require.Error(t, err)
	assert.Equal(t, types.ErrUnauthorizedVeto, err)

	// Add cosigners
	executed, err = k.CosignVeto(proposalID, "vetoer2")
	require.NoError(t, err)
	assert.False(t, executed) // Still need 1 more

	executed, err = k.CosignVeto(proposalID, "vetoer3")
	require.NoError(t, err)
	assert.True(t, executed) // Should execute with 3 cosigners

	// Verify proposal is vetoed
	proposal, err := k.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, types.StatusVetoed, proposal.Status)
}

func TestTimeLockExecution(t *testing.T) {
	k := NewKeeper(types.DefaultParams())

	// Create a proposal with execution delay
	proposalID, err := k.SubmitProposal(
		"Time-Lock Proposal",
		"Description",
		types.CategoryParameterChange,
		"proposer1",
		"15000000000",
		true,
	)
	require.NoError(t, err)

	// Cast votes to pass the proposal
	err = k.CastVote(proposalID, "voter1", types.OptionYes, "60000000", false, "")
	require.NoError(t, err)

	// Tally votes
	tally, err := k.TallyVotes(proposalID)
	require.NoError(t, err)
	assert.NotNil(t, tally)

	// Get proposal and check status
	proposal, err := k.GetProposal(proposalID)
	require.NoError(t, err)

	// Proposal should be in execution delay status
	assert.Equal(t, types.StatusExecutionDelay, proposal.Status)

	// Try to execute immediately (should fail)
	err = k.ExecuteProposal(proposalID, "executor1")
	require.Error(t, err)
	assert.Equal(t, types.ErrExecutionDelayNotPassed, err)

	// Mock passage of time
	proposal.ExecutionTime = time.Now().Add(-1 * time.Hour)
	proposal.Status = types.StatusPassed
	k.proposals[proposalID] = proposal

	// Now execution should succeed
	err = k.ExecuteProposal(proposalID, "executor1")
	require.NoError(t, err)

	// Verify proposal is executed
	proposal, err = k.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, types.StatusExecuted, proposal.Status)
}

func TestQuorumAndThreshold(t *testing.T) {
	k := NewKeeper(types.DefaultParams())

	tests := []struct {
		name  string
		votes map[string]struct {
			option types.VoteOption
			power  string
		}
		expectedStatus types.ProposalStatus
	}{
		{
			name: "passes with majority",
			votes: map[string]struct {
				option types.VoteOption
				power  string
			}{
				"voter1": {types.OptionYes, "60000000"},
				"voter2": {types.OptionNo, "20000000"},
				"voter3": {types.OptionAbstain, "10000000"},
			},
			expectedStatus: types.StatusPassed,
		},
		{
			name: "rejected with low yes votes",
			votes: map[string]struct {
				option types.VoteOption
				power  string
			}{
				"voter1": {types.OptionYes, "20000000"},
				"voter2": {types.OptionNo, "60000000"},
				"voter3": {types.OptionAbstain, "10000000"},
			},
			expectedStatus: types.StatusRejected,
		},
		{
			name: "vetoed with high veto votes",
			votes: map[string]struct {
				option types.VoteOption
				power  string
			}{
				"voter1": {types.OptionYes, "30000000"},
				"voter2": {types.OptionNoWithVeto, "40000000"},
				"voter3": {types.OptionNo, "20000000"},
			},
			expectedStatus: types.StatusVetoed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new proposal for each test
			pid, err := k.SubmitProposal(
				"Test "+tt.name,
				"Description",
				types.CategoryText,
				"proposer1",
				"5000000000",
				true,
			)
			require.NoError(t, err)

			// Cast votes
			for voter, voteData := range tt.votes {
				err := k.CastVote(pid, voter, voteData.option, voteData.power, false, "")
				require.NoError(t, err)
			}

			// Tally votes
			_, err = k.TallyVotes(pid)
			require.NoError(t, err)

			// Check status
			proposal, err := k.GetProposal(pid)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, proposal.Status)
		})
	}
}

func TestTokenLockDuringVoting(t *testing.T) {
	params := types.DefaultParams()
	params.RequireTokenLock = true
	params.TokenLockDuration = 24 * time.Hour
	k := NewKeeper(params)

	// Create a proposal
	proposalID, err := k.SubmitProposal(
		"Token Lock Test",
		"Description",
		types.CategoryEmergency,
		"proposer1",
		"50000000000",
		true,
	)
	require.NoError(t, err)

	// Cast vote (should lock tokens)
	voter := "voter1"
	votingPower := "1000000"
	err = k.CastVote(proposalID, voter, types.OptionYes, votingPower, false, "")
	require.NoError(t, err)

	// Verify tokens are locked
	locks := k.tokenLocks[voter]
	require.NotNil(t, locks)
	require.Len(t, locks, 1)

	lock := locks[0]
	assert.Equal(t, voter, lock.Owner)
	assert.Equal(t, proposalID, lock.ProposalID)
	assert.Equal(t, votingPower, lock.LockedAmount)
	assert.True(t, lock.UnlockTime.After(time.Now()))
}

func TestSnapshotVoting(t *testing.T) {
	params := types.DefaultParams()
	params.SnapshotVotingEnabled = true
	k := NewKeeper(params)

	// Create a proposal
	proposalID, err := k.SubmitProposal(
		"Snapshot Test",
		"Description",
		types.CategoryText,
		"proposer1",
		"5000000000",
		false,
	)
	require.NoError(t, err)

	// Submit snapshot vote
	voter := "voter1"
	votingPower := "1000000"
	signature := "mock_signature"

	err = k.SubmitSnapshotVote(proposalID, voter, types.OptionYes, votingPower, signature)
	require.NoError(t, err)

	// Verify snapshot vote was recorded
	snapshotVotes := k.snapshotVotes[proposalID]
	require.NotNil(t, snapshotVotes)
	require.Contains(t, snapshotVotes, voter)

	vote := snapshotVotes[voter]
	assert.Equal(t, types.OptionYes, vote.Option)
	assert.Equal(t, votingPower, vote.VotingPowerAtSnapshot)
	assert.Equal(t, signature, vote.Signature)
}

func TestEmergencyFastTrack(t *testing.T) {
	params := types.DefaultParams()
	params.EmergencyVotingPeriod = 24 * time.Hour
	params.EmergencyQuorum = "0.500"
	params.EmergencyThreshold = "0.667"
	k := NewKeeper(params)

	// Create emergency proposal
	proposalID, err := k.SubmitProposal(
		"Emergency Fast-Track",
		"Critical security fix",
		types.CategoryEmergency,
		"proposer1",
		"50000000000",
		true,
	)
	require.NoError(t, err)

	// Verify it starts voting immediately
	proposal, err := k.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, types.StatusVotingPeriod, proposal.Status)
	assert.True(t, proposal.IsEmergency)

	// Verify shorter voting period
	votingDuration := proposal.VotingEndTime.Sub(proposal.VotingStartTime)
	assert.Equal(t, 24*time.Hour, votingDuration)

	// Verify no execution delay for emergency
	categoryParams := params.GetCategoryParams(types.CategoryEmergency)
	assert.Equal(t, time.Duration(0), categoryParams.ExecutionDelay)
}

// Helper function to compute commitment
func computeCommitment(voter string, option types.VoteOption, key string) string {
	data := fmt.Sprintf("%s:%d:%s", voter, option, key)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
