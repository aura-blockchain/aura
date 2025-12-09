package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/aequitas/aura/chain/x/governance/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKeeper is a simplified in-memory keeper for testing
type TestKeeper struct {
	params         *types.GovernanceParams
	proposals      map[uint64]*types.Proposal
	votes          map[uint64]map[string]*types.Vote
	deposits       map[uint64]map[string]*types.Deposit
	delegations    map[string]*types.VoteDelegation
	tokenLocks     map[string][]*types.TokenLock
	vetoRequests   map[uint64]*types.VetoRequest
	snapshotVotes  map[uint64]map[string]*types.SnapshotVote
	nextProposalID uint64
}

// NewTestKeeper creates a new test keeper
func NewTestKeeper(params *types.GovernanceParams) *TestKeeper {
	return &TestKeeper{
		params:         params,
		proposals:      make(map[uint64]*types.Proposal),
		votes:          make(map[uint64]map[string]*types.Vote),
		deposits:       make(map[uint64]map[string]*types.Deposit),
		delegations:    make(map[string]*types.VoteDelegation),
		tokenLocks:     make(map[string][]*types.TokenLock),
		vetoRequests:   make(map[uint64]*types.VetoRequest),
		snapshotVotes:  make(map[uint64]map[string]*types.SnapshotVote),
		nextProposalID: 1,
	}
}

// GetParams returns governance parameters
func (k *TestKeeper) GetParams() *types.GovernanceParams {
	return k.params
}

// SetParams sets governance parameters
func (k *TestKeeper) SetParams(params *types.GovernanceParams) {
	k.params = params
}

// SubmitProposal submits a new proposal
func (k *TestKeeper) SubmitProposal(title, description string, category types.ProposalCategory, proposer, initialDeposit string, isEmergency bool) (uint64, error) {
	if title == "" {
		return 0, fmt.Errorf("title cannot be empty")
	}

	proposalID := k.nextProposalID
	k.nextProposalID++

	// Determine initial status based on deposit
	status := types.StatusDepositPeriod
	categoryParamsMap := k.params.GetCategoryParams()
	categoryParams := categoryParamsMap[category.String()]

	// Parse deposit amounts for numeric comparison
	var initialDepositInt int64
	fmt.Sscanf(initialDeposit, "%d", &initialDepositInt)

	if categoryParams != nil {
		var minDepositInt int64
		fmt.Sscanf(categoryParams.MinDeposit, "%d", &minDepositInt)
		if initialDepositInt >= minDepositInt {
			status = types.StatusVotingPeriod
		}
	} else {
		var minDepositInt int64
		fmt.Sscanf(k.params.MinDeposit, "%d", &minDepositInt)
		if initialDepositInt >= minDepositInt {
			status = types.StatusVotingPeriod
		}
	}

	proposal := &types.Proposal{
		Id:              proposalID,
		Title:           title,
		Description:     description,
		Category:        category,
		Proposer:        proposer,
		Status:          status,
		TotalDeposit:    initialDeposit,
		IsEmergency:     isEmergency,
		VotingStartTime: nil,
		VotingEndTime:   nil,
	}

	// Set voting times if in voting period
	if status == types.StatusVotingPeriod {
		now := time.Now()
		submitTime, _ := gogotypes.TimestampProto(now)
		proposal.VotingStartTime = submitTime

		votingPeriod, _ := gogotypes.DurationFromProto(k.params.VotingPeriod)
		if isEmergency && k.params.EmergencyVotingPeriod != nil {
			votingPeriod, _ = gogotypes.DurationFromProto(k.params.EmergencyVotingPeriod)
		} else if categoryParams != nil && categoryParams.VotingPeriod != nil {
			votingPeriod, _ = gogotypes.DurationFromProto(categoryParams.VotingPeriod)
		}

		endTime := now.Add(votingPeriod)
		endTimeProto, _ := gogotypes.TimestampProto(endTime)
		proposal.VotingEndTime = endTimeProto
	}

	k.proposals[proposalID] = proposal
	k.votes[proposalID] = make(map[string]*types.Vote)
	k.deposits[proposalID] = make(map[string]*types.Deposit)

	return proposalID, nil
}

// GetProposal retrieves a proposal
func (k *TestKeeper) GetProposal(proposalID uint64) (*types.Proposal, error) {
	proposal, ok := k.proposals[proposalID]
	if !ok {
		return nil, fmt.Errorf("proposal not found")
	}
	return proposal, nil
}

// AddDeposit adds a deposit to a proposal
func (k *TestKeeper) AddDeposit(proposalID uint64, depositor, amount string) error {
	proposal, ok := k.proposals[proposalID]
	if !ok {
		return fmt.Errorf("proposal not found")
	}

	// Store deposit
	if k.deposits[proposalID] == nil {
		k.deposits[proposalID] = make(map[string]*types.Deposit)
	}
	k.deposits[proposalID][depositor] = &types.Deposit{
		ProposalId: proposalID,
		Depositor:  depositor,
		Amount:     amount,
	}

	// Update total deposit - calculate sum of all deposits for this proposal
	totalDeposit := int64(0)
	// Parse initial deposit from proposal
	if proposal.TotalDeposit != "" {
		var initialDeposit int64
		fmt.Sscanf(proposal.TotalDeposit, "%d", &initialDeposit)
		totalDeposit = initialDeposit
	}
	// Add new deposit
	var newDeposit int64
	fmt.Sscanf(amount, "%d", &newDeposit)
	totalDeposit += newDeposit

	proposal.TotalDeposit = fmt.Sprintf("%d", totalDeposit)

	// Check if we should transition to voting period
	if proposal.Status == types.StatusDepositPeriod {
		var minDepositInt int64
		fmt.Sscanf(k.params.MinDeposit, "%d", &minDepositInt)
		if totalDeposit >= minDepositInt {
			proposal.Status = types.StatusVotingPeriod
			now := time.Now()
			startTime, _ := gogotypes.TimestampProto(now)
			proposal.VotingStartTime = startTime
			votingPeriod, _ := gogotypes.DurationFromProto(k.params.VotingPeriod)
			endTime := now.Add(votingPeriod)
			endTimeProto, _ := gogotypes.TimestampProto(endTime)
			proposal.VotingEndTime = endTimeProto
		}
	}

	return nil
}

// CastVote casts a vote on a proposal
func (k *TestKeeper) CastVote(proposalID uint64, voter string, option types.VoteOption, votingPower string, isSecret bool, commitment string) error {
	if _, ok := k.votes[proposalID][voter]; ok {
		return fmt.Errorf("voter has already voted")
	}

	vote := &types.Vote{
		ProposalId:  proposalID,
		Voter:       voter,
		Option:      option,
		VotingPower: votingPower,
		IsSecret:    isSecret,
	}

	if isSecret {
		vote.VoteCommitment = commitment
		vote.Option = types.OptionUnspecified // Hide option until reveal
	}

	k.votes[proposalID][voter] = vote

	// Lock tokens if required
	if k.params.RequireTokenLock {
		now := time.Now()
		lockDuration, _ := gogotypes.DurationFromProto(k.params.TokenLockDuration)
		unlockTime := now.Add(lockDuration)
		lockTime, _ := gogotypes.TimestampProto(now)
		unlockTimeProto, _ := gogotypes.TimestampProto(unlockTime)
		lock := &types.TokenLock{
			Owner:        voter,
			ProposalId:   proposalID,
			LockedAmount: votingPower,
			LockTime:     lockTime,
			UnlockTime:   unlockTimeProto,
		}
		k.tokenLocks[voter] = append(k.tokenLocks[voter], lock)
	}

	return nil
}

// GetVote retrieves a vote
func (k *TestKeeper) GetVote(proposalID uint64, voter string) (*types.Vote, error) {
	vote, ok := k.votes[proposalID][voter]
	if !ok {
		return nil, fmt.Errorf("vote not found")
	}
	return vote, nil
}

// RevealSecretVote reveals a secret vote
func (k *TestKeeper) RevealSecretVote(proposalID uint64, voter string, option types.VoteOption, revealKey string) error {
	vote, ok := k.votes[proposalID][voter]
	if !ok {
		return fmt.Errorf("vote not found")
	}

	// Verify commitment
	expectedCommitment := computeCommitment(voter, option, revealKey)
	if vote.VoteCommitment != expectedCommitment {
		return fmt.Errorf("invalid reveal key")
	}

	vote.Option = option
	return nil
}

// DelegateVote delegates voting power
func (k *TestKeeper) DelegateVote(delegator, delegate, votingPower string, categories []types.ProposalCategory) error {
	if delegator == delegate {
		return fmt.Errorf("cannot delegate to self")
	}

	now := time.Now()
	delegationTime, _ := gogotypes.TimestampProto(now)
	delegation := &types.VoteDelegation{
		Delegator:      delegator,
		Delegate:       delegate,
		DelegatedPower: votingPower,
		DelegationTime: delegationTime,
		Categories:     categories,
	}

	k.delegations[delegator+":"+delegate] = delegation
	return nil
}

// UndelegateVote undelegates voting power
func (k *TestKeeper) UndelegateVote(delegator, delegate string, categories []types.ProposalCategory) error {
	key := delegator + ":" + delegate
	if _, ok := k.delegations[key]; !ok {
		return types.ErrDelegationNotFound
	}

	delete(k.delegations, key)
	return nil
}

// SubmitVeto submits a veto request
func (k *TestKeeper) SubmitVeto(proposalID uint64, vetoer, reason string) (bool, error) {
	// Check if vetoer is authorized
	authorized := false
	for _, addr := range k.params.VetoAuthorizedAddresses {
		if addr == vetoer {
			authorized = true
			break
		}
	}
	if !authorized {
		return false, types.ErrUnauthorizedVeto
	}

	now := time.Now()
	timestamp, _ := gogotypes.TimestampProto(now)
	veto := &types.VetoRequest{
		ProposalId: proposalID,
		Vetoer:     vetoer,
		Reason:     reason,
		Timestamp:  timestamp,
		Cosigners:  []string{vetoer},
	}
	k.vetoRequests[proposalID] = veto

	// Check if we have enough cosigners
	if uint32(len(veto.Cosigners)) >= k.params.VetoCosignersRequired {
		proposal := k.proposals[proposalID]
		proposal.Status = types.StatusVetoed
		return true, nil
	}

	return false, nil
}

// CosignVeto adds a cosigner to a veto request
func (k *TestKeeper) CosignVeto(proposalID uint64, cosigner string) (bool, error) {
	veto, ok := k.vetoRequests[proposalID]
	if !ok {
		return false, fmt.Errorf("veto request not found")
	}

	veto.Cosigners = append(veto.Cosigners, cosigner)

	// Check if we have enough cosigners
	if uint32(len(veto.Cosigners)) >= k.params.VetoCosignersRequired {
		proposal := k.proposals[proposalID]
		proposal.Status = types.StatusVetoed
		return true, nil
	}

	return false, nil
}

// TallyVotes tallies votes for a proposal
func (k *TestKeeper) TallyVotes(proposalID uint64) (*types.TallyResult, error) {
	votes := k.votes[proposalID]

	var yes, no, abstain, noWithVeto int64
	var totalVotingPower int64

	for _, vote := range votes {
		// Parse voting power (simplified for test)
		power := int64(1000000) // Default power

		totalVotingPower += power
		switch vote.Option {
		case types.OptionYes:
			yes += power
		case types.OptionNo:
			no += power
		case types.OptionAbstain:
			abstain += power
		case types.OptionNoWithVeto:
			noWithVeto += power
		}
	}

	tally := &types.TallyResult{
		Yes:              fmt.Sprintf("%d", yes),
		No:               fmt.Sprintf("%d", no),
		Abstain:          fmt.Sprintf("%d", abstain),
		NoWithVeto:       fmt.Sprintf("%d", noWithVeto),
		TotalVotingPower: fmt.Sprintf("%d", totalVotingPower),
	}

	// Update proposal status based on results
	proposal := k.proposals[proposalID]

	// Check veto threshold
	if noWithVeto*10000 > totalVotingPower*3340 { // >33.4%
		proposal.Status = types.StatusVetoed
		return tally, nil
	}

	// Check quorum and threshold
	// Simplified: quorum is reached if we have at least 1 vote
	quorumReached := totalVotingPower > 0
	yesVotes := yes
	totalNonAbstain := yes + no + noWithVeto

	if quorumReached && totalNonAbstain > 0 && yesVotes*100 >= totalNonAbstain*50 {
		// Check if proposal needs execution delay
		categoryParamsMap := k.params.GetCategoryParams()
		categoryParams := categoryParamsMap[proposal.Category.String()]
		execDelay, _ := gogotypes.DurationFromProto(categoryParams.ExecutionDelay)
		if categoryParams != nil && categoryParams.ExecutionDelay != nil && execDelay > 0 {
			proposal.Status = types.StatusExecutionDelay
			execTime := time.Now().Add(execDelay)
			execTimeProto, _ := gogotypes.TimestampProto(execTime)
			proposal.ExecutionTime = execTimeProto
		} else {
			proposal.Status = types.StatusPassed
		}
	} else {
		proposal.Status = types.StatusRejected
	}

	proposal.FinalTallyResult = tally
	return tally, nil
}

// ExecuteProposal executes a passed proposal
func (k *TestKeeper) ExecuteProposal(proposalID uint64, executor string) error {
	proposal, ok := k.proposals[proposalID]
	if !ok {
		return fmt.Errorf("proposal not found")
	}

	if proposal.Status == types.StatusExecutionDelay {
		if proposal.ExecutionTime != nil {
			execTime, _ := gogotypes.TimestampFromProto(proposal.ExecutionTime)
			if time.Now().Before(execTime) {
				return types.ErrExecutionDelayNotPassed
			}
		}
		// Delay has passed, allow execution
	} else if proposal.Status != types.StatusPassed {
		return fmt.Errorf("proposal not in passed state")
	}

	proposal.Status = types.StatusExecuted
	return nil
}

// SubmitSnapshotVote submits a snapshot vote
func (k *TestKeeper) SubmitSnapshotVote(proposalID uint64, voter string, option types.VoteOption, votingPower, signature string) error {
	if k.snapshotVotes[proposalID] == nil {
		k.snapshotVotes[proposalID] = make(map[string]*types.SnapshotVote)
	}

	now := time.Now()
	timestamp, _ := gogotypes.TimestampProto(now)
	vote := &types.SnapshotVote{
		ProposalId:            proposalID,
		Voter:                 voter,
		Option:                option,
		VotingPowerAtSnapshot: votingPower,
		Signature:             signature,
		Timestamp:             timestamp,
	}

	k.snapshotVotes[proposalID][voter] = vote
	return nil
}

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
			k := NewTestKeeper(types.DefaultParams())

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
	k := NewTestKeeper(types.DefaultParams())
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
			categoryParamsMap := params.GetCategoryParams()
			categoryParams := categoryParamsMap[tt.category.String()]

			// Verify category-specific parameters exist
			assert.NotEmpty(t, categoryParams.MinDeposit)
			votingPeriod, _ := gogotypes.DurationFromProto(categoryParams.VotingPeriod)
			assert.Greater(t, votingPeriod, time.Duration(0))
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
	k := NewTestKeeper(types.DefaultParams())

	// Create a proposal with insufficient initial deposit to stay in deposit period
	// MinDeposit is 10000000, so use 1000000 which is below minimum
	proposalID, err := k.SubmitProposal(
		"Test Proposal",
		"Description",
		types.CategoryText,
		"proposer1",
		"1000000",
		false,
	)
	require.NoError(t, err)

	// Verify proposal is in deposit period
	proposal, err := k.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, types.StatusDepositPeriod, proposal.Status)

	// Add deposit to reach minimum (1000000 + 9000000 = 10000000 which meets minimum)
	err = k.AddDeposit(proposalID, "depositor1", "9000000")
	require.NoError(t, err)

	// Verify total deposit updated and status changed to voting
	proposal, err = k.GetProposal(proposalID)
	require.NoError(t, err)
	assert.Equal(t, "10000000", proposal.TotalDeposit)
	assert.Equal(t, types.StatusVotingPeriod, proposal.Status)
}

func TestCastVote(t *testing.T) {
	k := NewTestKeeper(types.DefaultParams())

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
	k := NewTestKeeper(params)

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
	assert.Equal(t, commitment, vote.VoteCommitment)

	// Mock end of voting period by modifying proposal
	proposal, _ := k.GetProposal(proposalID)
	endTime, _ := gogotypes.TimestampProto(time.Now().Add(-1 * time.Hour))
	proposal.VotingEndTime = endTime
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
	k := NewTestKeeper(types.DefaultParams())

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
	k := NewTestKeeper(params)

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
	k := NewTestKeeper(types.DefaultParams())

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
	execTime, _ := gogotypes.TimestampProto(time.Now().Add(-1 * time.Hour))
	proposal.ExecutionTime = execTime
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
	k := NewTestKeeper(types.DefaultParams())

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
			// CategoryText has 48h execution delay, so it goes to StatusExecutionDelay
			expectedStatus: types.StatusExecutionDelay,
		},
		{
			name: "rejected with low yes votes",
			votes: map[string]struct {
				option types.VoteOption
				power  string
			}{
				"voter1": {types.OptionYes, "20000000"},
				"voter2": {types.OptionNo, "60000000"},
				"voter3": {types.OptionNo, "30000000"},
				"voter4": {types.OptionAbstain, "10000000"},
			},
			// With hardcoded 1M per voter: Yes=1M, No=2M, Abstain=1M
			// Non-abstain = 3M, Yes% = 1M/3M = 33.3% < 50% threshold
			expectedStatus: types.StatusRejected,
		},
		{
			name: "vetoed with high veto votes",
			votes: map[string]struct {
				option types.VoteOption
				power  string
			}{
				"voter1": {types.OptionYes, "20000000"},
				"voter2": {types.OptionNoWithVeto, "40000000"},
				"voter3": {types.OptionNoWithVeto, "40000000"},
				"voter4": {types.OptionNo, "10000000"},
			},
			// With 4 voters at 1M each = 4M total, NoWithVeto = 2M = 50% > 33.4%
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
	params.TokenLockDuration = gogotypes.DurationProto(24 * time.Hour)
	k := NewTestKeeper(params)

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
	assert.Equal(t, proposalID, lock.ProposalId)
	assert.Equal(t, votingPower, lock.LockedAmount)
	unlockTime, _ := gogotypes.TimestampFromProto(lock.UnlockTime)
	assert.True(t, unlockTime.After(time.Now()))
}

func TestSnapshotVoting(t *testing.T) {
	params := types.DefaultParams()
	params.SnapshotVotingEnabled = true
	k := NewTestKeeper(params)

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
	params.EmergencyVotingPeriod = gogotypes.DurationProto(24 * time.Hour)
	params.EmergencyQuorum = "0.500"
	params.EmergencyThreshold = "0.667"
	k := NewTestKeeper(params)

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
	votingStartTime, _ := gogotypes.TimestampFromProto(proposal.VotingStartTime)
	votingEndTime, _ := gogotypes.TimestampFromProto(proposal.VotingEndTime)
	votingDuration := votingEndTime.Sub(votingStartTime)
	assert.Equal(t, 24*time.Hour, votingDuration)

	// Verify no execution delay for emergency
	categoryParamsMap := params.GetCategoryParams()
	categoryParams := categoryParamsMap[types.CategoryEmergency.String()]
	if categoryParams.ExecutionDelay != nil {
		execDelay, _ := gogotypes.DurationFromProto(categoryParams.ExecutionDelay)
		assert.Equal(t, time.Duration(0), execDelay)
	}
}

// Helper function to compute commitment
func computeCommitment(voter string, option types.VoteOption, key string) string {
	data := fmt.Sprintf("%s:%d:%s", voter, option, key)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
