package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Test Default Parameters

func TestDefaultParamsExtended(t *testing.T) {
	params := DefaultParams()
	require.NotNil(t, params)

	// Test deposit parameters
	assert.Equal(t, "10000000stake", params.MinDeposit)
	assert.NotNil(t, params.MaxDepositPeriod)
	assert.Equal(t, 7*24*time.Hour, params.MaxDepositPeriod.AsDuration())

	// Test voting parameters
	assert.NotNil(t, params.VotingPeriod)
	assert.Equal(t, 7*24*time.Hour, params.VotingPeriod.AsDuration())
	assert.Equal(t, "0.334", params.Quorum)
	assert.Equal(t, "0.50", params.Threshold)
	assert.Equal(t, "0.334", params.VetoThreshold)

	// Test execution delay
	assert.NotNil(t, params.ExecutionDelay)
	assert.Equal(t, 48*time.Hour, params.ExecutionDelay.AsDuration())

	// Test emergency parameters
	assert.NotNil(t, params.EmergencyVotingPeriod)
	assert.Equal(t, 24*time.Hour, params.EmergencyVotingPeriod.AsDuration())
	assert.Equal(t, "0.50", params.EmergencyQuorum)
	assert.Equal(t, "0.667", params.EmergencyThreshold)

	// Test veto parameters
	assert.Equal(t, uint32(3), params.VetoCosignersRequired)
	assert.NotNil(t, params.VetoAuthorizedAddresses)

	// Test token lock parameters
	assert.False(t, params.RequireTokenLock)
	assert.NotNil(t, params.TokenLockDuration)
	assert.Equal(t, 7*24*time.Hour, params.TokenLockDuration.AsDuration())

	// Test snapshot parameters
	assert.False(t, params.SnapshotVotingEnabled)
	assert.Equal(t, uint64(100), params.SnapshotLookbackBlocks)

	// Test secret ballot parameters
	assert.False(t, params.SecretBallotEnabled)
	assert.NotNil(t, params.RevealPeriod)
	assert.Equal(t, 24*time.Hour, params.RevealPeriod.AsDuration())
}

func TestCategoryParams(t *testing.T) {
	params := DefaultParams()
	require.NotNil(t, params.CategoryParams)

	tests := []struct {
		name              string
		category          ProposalCategory
		expectedQuorum    string
		expectedThreshold string
		expectedDelay     time.Duration
	}{
		{"Text", CategoryText, "0.334", "0.50", 48 * time.Hour},
		{"ParameterChange", CategoryParameterChange, "0.334", "0.50", 48 * time.Hour},
		{"SoftwareUpgrade", CategorySoftwareUpgrade, "0.334", "0.50", 48 * time.Hour},
		{"Spending", CategorySpending, "0.334", "0.50", 48 * time.Hour},
		{"Emergency", CategoryEmergency, "0.600", "0.750", 0},
		{"Constitution", CategoryConstitution, "0.667", "0.750", 48 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			categoryParams, exists := params.CategoryParams[tt.category.String()]
			require.True(t, exists, "Category %s should exist", tt.name)
			assert.Equal(t, tt.expectedQuorum, categoryParams.Quorum)
			assert.Equal(t, tt.expectedThreshold, categoryParams.Threshold)
			assert.Equal(t, tt.expectedDelay, categoryParams.ExecutionDelay.AsDuration())
		})
	}
}

// Test Proposal

func TestProposal(t *testing.T) {
	now := time.Now()

	proposal := &Proposal{
		Id:              1,
		Title:           "Test Proposal",
		Description:     "This is a test proposal",
		Category:        CategoryText,
		Proposer:        "aura1test",
		Status:          StatusDepositPeriod,
		SubmitTime:      timestamppb.New(now),
		DepositEndTime:  timestamppb.New(now.Add(7 * 24 * time.Hour)),
		TotalDeposit:    "1000000",
		VotingStartTime: nil,
		VotingEndTime:   nil,
		IsEmergency:     false,
	}

	assert.Equal(t, uint64(1), proposal.Id)
	assert.Equal(t, "Test Proposal", proposal.Title)
	assert.Equal(t, CategoryText, proposal.Category)
	assert.Equal(t, StatusDepositPeriod, proposal.Status)
}

func TestProposalAllStatuses(t *testing.T) {
	statuses := []ProposalStatus{
		StatusDepositPeriod,
		StatusVotingPeriod,
		StatusPassed,
		StatusRejected,
		StatusFailed,
		StatusVetoed,
		StatusExecutionDelay,
		StatusReadyForExecution,
		StatusExecuted,
	}

	for _, status := range statuses {
		proposal := &Proposal{
			Id:       1,
			Title:    "Test",
			Proposer: "aura1test",
			Status:   status,
		}
		assert.Equal(t, status, proposal.Status)
	}
}

func TestProposalAllCategories(t *testing.T) {
	categories := []ProposalCategory{
		CategoryText,
		CategoryParameterChange,
		CategorySoftwareUpgrade,
		CategorySpending,
		CategoryEmergency,
		CategoryConstitution,
	}

	for _, category := range categories {
		proposal := &Proposal{
			Id:       1,
			Title:    "Test",
			Proposer: "aura1test",
			Category: category,
		}
		assert.Equal(t, category, proposal.Category)
	}
}

// Test Vote

func TestVote(t *testing.T) {
	vote := &Vote{
		ProposalId:  1,
		Voter:       "aura1voter",
		Option:      OptionYes,
		VotingPower: "1000000",
		IsSecret:    false,
	}

	assert.Equal(t, uint64(1), vote.ProposalId)
	assert.Equal(t, "aura1voter", vote.Voter)
	assert.Equal(t, OptionYes, vote.Option)
	assert.Equal(t, "1000000", vote.VotingPower)
	assert.False(t, vote.IsSecret)
}

func TestVoteOptions(t *testing.T) {
	options := []VoteOption{
		OptionUnspecified,
		OptionYes,
		OptionAbstain,
		OptionNo,
		OptionNoWithVeto,
	}

	for _, option := range options {
		vote := &Vote{
			ProposalId:  1,
			Voter:       "aura1voter",
			Option:      option,
			VotingPower: "1000000",
		}
		assert.Equal(t, option, vote.Option)
	}
}

func TestSecretVote(t *testing.T) {
	commitment := "abcd1234"
	vote := &Vote{
		ProposalId:     1,
		Voter:          "aura1voter",
		Option:         OptionUnspecified, // Hidden until reveal
		VotingPower:    "1000000",
		IsSecret:       true,
		VoteCommitment: commitment,
	}

	assert.True(t, vote.IsSecret)
	assert.Equal(t, commitment, vote.VoteCommitment)
	assert.Equal(t, OptionUnspecified, vote.Option)
}

// Test Weighted Vote

func TestWeightedVoteOption(t *testing.T) {
	weighted := &WeightedVoteOption{
		Option: OptionYes,
		Weight: "0.50",
	}

	assert.Equal(t, OptionYes, weighted.Option)
	assert.Equal(t, "0.50", weighted.Weight)
}

// Test Deposit

func TestDeposit(t *testing.T) {
	deposit := &Deposit{
		ProposalId: 1,
		Depositor:  "aura1depositor",
		Amount:     "5000000",
	}

	assert.Equal(t, uint64(1), deposit.ProposalId)
	assert.Equal(t, "aura1depositor", deposit.Depositor)
	assert.Equal(t, "5000000", deposit.Amount)
}

// Test TallyResult

func TestTallyResult(t *testing.T) {
	tally := &TallyResult{
		Yes:              "60000000",
		Abstain:          "10000000",
		No:               "20000000",
		NoWithVeto:       "5000000",
		TotalVotingPower: "100000000",
	}

	assert.Equal(t, "60000000", tally.Yes)
	assert.Equal(t, "10000000", tally.Abstain)
	assert.Equal(t, "20000000", tally.No)
	assert.Equal(t, "5000000", tally.NoWithVeto)
	assert.Equal(t, "100000000", tally.TotalVotingPower)
}

func TestTallyResultEmpty(t *testing.T) {
	tally := &TallyResult{
		Yes:              "0",
		Abstain:          "0",
		No:               "0",
		NoWithVeto:       "0",
		TotalVotingPower: "0",
	}

	assert.Equal(t, "0", tally.Yes)
	assert.Equal(t, "0", tally.TotalVotingPower)
}

// Test VoteDelegation

func TestVoteDelegation(t *testing.T) {
	now := time.Now()
	delegation := &VoteDelegation{
		Delegator:      "aura1delegator",
		Delegate:       "aura1delegate",
		DelegatedPower: "5000000",
		DelegationTime: timestamppb.New(now),
		Categories:     []ProposalCategory{CategoryText, CategoryParameterChange},
	}

	assert.Equal(t, "aura1delegator", delegation.Delegator)
	assert.Equal(t, "aura1delegate", delegation.Delegate)
	assert.Equal(t, "5000000", delegation.DelegatedPower)
	assert.Len(t, delegation.Categories, 2)
}

func TestVoteDelegationAllCategories(t *testing.T) {
	delegation := &VoteDelegation{
		Delegator:      "aura1delegator",
		Delegate:       "aura1delegate",
		DelegatedPower: "5000000",
		Categories:     []ProposalCategory{}, // Empty means all categories
	}

	assert.Empty(t, delegation.Categories)
}

// Test VetoRequest

func TestVetoRequest(t *testing.T) {
	now := time.Now()
	veto := &VetoRequest{
		ProposalId: 1,
		Vetoer:     "aura1vetoer",
		Reason:     "Security concerns",
		Timestamp:  timestamppb.New(now),
		Cosigners:  []string{"aura1vetoer", "aura1cosigner1", "aura1cosigner2"},
	}

	assert.Equal(t, uint64(1), veto.ProposalId)
	assert.Equal(t, "aura1vetoer", veto.Vetoer)
	assert.Equal(t, "Security concerns", veto.Reason)
	assert.Len(t, veto.Cosigners, 3)
}

// Test TokenLock

func TestTokenLock(t *testing.T) {
	now := time.Now()
	lock := &TokenLock{
		Owner:        "aura1owner",
		ProposalId:   1,
		LockedAmount: "1000000",
		LockTime:     timestamppb.New(now),
		UnlockTime:   timestamppb.New(now.Add(7 * 24 * time.Hour)),
	}

	assert.Equal(t, "aura1owner", lock.Owner)
	assert.Equal(t, uint64(1), lock.ProposalId)
	assert.Equal(t, "1000000", lock.LockedAmount)
	assert.NotNil(t, lock.LockTime)
	assert.NotNil(t, lock.UnlockTime)
	assert.True(t, lock.UnlockTime.AsTime().After(lock.LockTime.AsTime()))
}

// Test SnapshotVote

func TestSnapshotVote(t *testing.T) {
	now := time.Now()
	vote := &SnapshotVote{
		ProposalId:            1,
		Voter:                 "aura1voter",
		Option:                OptionYes,
		VotingPowerAtSnapshot: "2000000",
		SnapshotHeight:        1000,
		Signature:             "signature_data",
		Timestamp:             timestamppb.New(now),
	}

	assert.Equal(t, uint64(1), vote.ProposalId)
	assert.Equal(t, "aura1voter", vote.Voter)
	assert.Equal(t, OptionYes, vote.Option)
	assert.Equal(t, "2000000", vote.VotingPowerAtSnapshot)
	assert.Equal(t, uint64(1000), vote.SnapshotHeight)
	assert.NotEmpty(t, vote.Signature)
}

// Test CategoryParams

func TestCategoryParamsDefaults(t *testing.T) {
	categoryParams := &CategoryParams{
		MinDeposit:     "10000000",
		VotingPeriod:   durationpb.New(7 * 24 * time.Hour),
		Quorum:         "0.334",
		Threshold:      "0.50",
		VetoThreshold:  "0.334",
		ExecutionDelay: durationpb.New(48 * time.Hour),
	}

	assert.Equal(t, "10000000", categoryParams.MinDeposit)
	assert.Equal(t, 7*24*time.Hour, categoryParams.VotingPeriod.AsDuration())
	assert.Equal(t, "0.334", categoryParams.Quorum)
	assert.Equal(t, "0.50", categoryParams.Threshold)
	assert.Equal(t, "0.334", categoryParams.VetoThreshold)
	assert.Equal(t, 48*time.Hour, categoryParams.ExecutionDelay.AsDuration())
}

// Test Error Values

func TestErrors(t *testing.T) {
	errors := []error{
		ErrInvalidProposal,
		ErrInsufficientDeposit,
		ErrInvalidVote,
		ErrProposalNotFound,
		ErrInvalidProposalStatus,
		ErrVotingPeriodEnded,
		ErrVotingPeriodNotStarted,
		ErrAlreadyVoted,
		ErrInvalidDeposit,
		ErrDepositPeriodEnded,
		ErrUnauthorizedVeto,
		ErrInsufficientVetoCosigners,
		ErrExecutionDelayNotPassed,
		ErrProposalNotPassed,
		ErrAlreadyExecuted,
		ErrInvalidDelegation,
		ErrDelegationNotFound,
		ErrInsufficientTokens,
		ErrTokensLocked,
		ErrInvalidSnapshot,
		ErrInvalidReveal,
		ErrRevealPeriodNotStarted,
		ErrRevealPeriodEnded,
		ErrInvalidCommitment,
		ErrWeightedVoteNotEnabled,
		ErrInvalidWeight,
	}

	for _, err := range errors {
		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	}
}

// Test Enum Constants

func TestProposalCategoryConstants(t *testing.T) {
	assert.Equal(t, ProposalCategory_PROPOSAL_CATEGORY_TEXT, CategoryText)
	assert.Equal(t, ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE, CategoryParameterChange)
	assert.Equal(t, ProposalCategory_PROPOSAL_CATEGORY_SOFTWARE_UPGRADE, CategorySoftwareUpgrade)
	assert.Equal(t, ProposalCategory_PROPOSAL_CATEGORY_SPENDING, CategorySpending)
	assert.Equal(t, ProposalCategory_PROPOSAL_CATEGORY_EMERGENCY, CategoryEmergency)
	assert.Equal(t, ProposalCategory_PROPOSAL_CATEGORY_CONSTITUTION, CategoryConstitution)
}

func TestProposalStatusConstants(t *testing.T) {
	assert.Equal(t, ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD, StatusDepositPeriod)
	assert.Equal(t, ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD, StatusVotingPeriod)
	assert.Equal(t, ProposalStatus_PROPOSAL_STATUS_PASSED, StatusPassed)
	assert.Equal(t, ProposalStatus_PROPOSAL_STATUS_REJECTED, StatusRejected)
	assert.Equal(t, ProposalStatus_PROPOSAL_STATUS_FAILED, StatusFailed)
	assert.Equal(t, ProposalStatus_PROPOSAL_STATUS_VETOED, StatusVetoed)
	assert.Equal(t, ProposalStatus_PROPOSAL_STATUS_EXECUTION_DELAY, StatusExecutionDelay)
	assert.Equal(t, ProposalStatus_PROPOSAL_STATUS_READY_FOR_EXECUTION, StatusReadyForExecution)
	assert.Equal(t, ProposalStatus_PROPOSAL_STATUS_EXECUTED, StatusExecuted)
}

func TestVoteOptionConstants(t *testing.T) {
	assert.Equal(t, VoteOption_VOTE_OPTION_UNSPECIFIED, OptionUnspecified)
	assert.Equal(t, VoteOption_VOTE_OPTION_YES, OptionYes)
	assert.Equal(t, VoteOption_VOTE_OPTION_ABSTAIN, OptionAbstain)
	assert.Equal(t, VoteOption_VOTE_OPTION_NO, OptionNo)
	assert.Equal(t, VoteOption_VOTE_OPTION_NO_WITH_VETO, OptionNoWithVeto)
}

// Test Edge Cases

func TestProposalWithLongDescription(t *testing.T) {
	longDesc := string(make([]byte, 10000))
	proposal := &Proposal{
		Id:          1,
		Title:       "Long Description",
		Description: longDesc,
		Proposer:    "aura1test",
	}

	assert.Equal(t, 10000, len(proposal.Description))
}

func TestProposalWithEmptyFields(t *testing.T) {
	proposal := &Proposal{
		Id:          0,
		Title:       "",
		Description: "",
		Proposer:    "",
	}

	assert.Equal(t, uint64(0), proposal.Id)
	assert.Empty(t, proposal.Title)
	assert.Empty(t, proposal.Description)
	assert.Empty(t, proposal.Proposer)
}

func TestVoteWithZeroPower(t *testing.T) {
	vote := &Vote{
		ProposalId:  1,
		Voter:       "aura1voter",
		Option:      OptionYes,
		VotingPower: "0",
	}

	assert.Equal(t, "0", vote.VotingPower)
}

func TestDepositWithZeroAmount(t *testing.T) {
	deposit := &Deposit{
		ProposalId: 1,
		Depositor:  "aura1depositor",
		Amount:     "0",
	}

	assert.Equal(t, "0", deposit.Amount)
}

func TestVoteDelegationEmptyDelegate(t *testing.T) {
	delegation := &VoteDelegation{
		Delegator:      "aura1delegator",
		Delegate:       "",
		DelegatedPower: "0",
	}

	assert.Empty(t, delegation.Delegate)
	assert.Equal(t, "0", delegation.DelegatedPower)
}

func TestVetoRequestNoCosigners(t *testing.T) {
	veto := &VetoRequest{
		ProposalId: 1,
		Vetoer:     "aura1vetoer",
		Cosigners:  []string{},
	}

	assert.Empty(t, veto.Cosigners)
}

// Test Governance Params GetCategoryParams Helper

func TestGovernanceParamsGetCategoryParams(t *testing.T) {
	params := DefaultParams()
	categoryParams := params.GetCategoryParams()

	require.NotNil(t, categoryParams)
	assert.Len(t, categoryParams, 6) // 6 categories

	// Test each category exists
	_, ok := categoryParams[CategoryText.String()]
	assert.True(t, ok)

	_, ok = categoryParams[CategoryEmergency.String()]
	assert.True(t, ok)

	_, ok = categoryParams[CategoryConstitution.String()]
	assert.True(t, ok)
}

// Test Time-based Fields

func TestProposalTimeFields(t *testing.T) {
	now := time.Now()
	submitTime := timestamppb.New(now)
	depositEndTime := timestamppb.New(now.Add(7 * 24 * time.Hour))
	votingStartTime := timestamppb.New(now.Add(7 * 24 * time.Hour))
	votingEndTime := timestamppb.New(now.Add(14 * 24 * time.Hour))
	executionTime := timestamppb.New(now.Add(16 * 24 * time.Hour))

	proposal := &Proposal{
		Id:              1,
		Title:           "Time Test",
		Proposer:        "aura1test",
		SubmitTime:      submitTime,
		DepositEndTime:  depositEndTime,
		VotingStartTime: votingStartTime,
		VotingEndTime:   votingEndTime,
		ExecutionTime:   executionTime,
	}

	assert.NotNil(t, proposal.SubmitTime)
	assert.NotNil(t, proposal.DepositEndTime)
	assert.NotNil(t, proposal.VotingStartTime)
	assert.NotNil(t, proposal.VotingEndTime)
	assert.NotNil(t, proposal.ExecutionTime)

	assert.True(t, proposal.DepositEndTime.AsTime().After(proposal.SubmitTime.AsTime()))
	assert.True(t, proposal.VotingEndTime.AsTime().After(proposal.VotingStartTime.AsTime()))
}

func TestDurationFields(t *testing.T) {
	params := DefaultParams()

	durations := []struct {
		name     string
		duration *durationpb.Duration
		expected time.Duration
	}{
		{"MaxDepositPeriod", params.MaxDepositPeriod, 7 * 24 * time.Hour},
		{"VotingPeriod", params.VotingPeriod, 7 * 24 * time.Hour},
		{"ExecutionDelay", params.ExecutionDelay, 48 * time.Hour},
		{"EmergencyVotingPeriod", params.EmergencyVotingPeriod, 24 * time.Hour},
		{"TokenLockDuration", params.TokenLockDuration, 7 * 24 * time.Hour},
		{"RevealPeriod", params.RevealPeriod, 24 * time.Hour},
	}

	for _, d := range durations {
		t.Run(d.name, func(t *testing.T) {
			assert.NotNil(t, d.duration)
			assert.Equal(t, d.expected, d.duration.AsDuration())
		})
	}
}

// Test Large Numbers

func TestLargeVotingPower(t *testing.T) {
	vote := &Vote{
		ProposalId:  1,
		Voter:       "aura1voter",
		Option:      OptionYes,
		VotingPower: "999999999999999999", // Large number
	}

	assert.Equal(t, "999999999999999999", vote.VotingPower)
}

func TestLargeDepositAmount(t *testing.T) {
	deposit := &Deposit{
		ProposalId: 1,
		Depositor:  "aura1depositor",
		Amount:     "999999999999999999", // Large number
	}

	assert.Equal(t, "999999999999999999", deposit.Amount)
}

// Test Multiple Weighted Votes

func TestMultipleWeightedVotes(t *testing.T) {
	weights := []*WeightedVoteOption{
		{Option: OptionYes, Weight: "0.50"},
		{Option: OptionNo, Weight: "0.30"},
		{Option: OptionAbstain, Weight: "0.20"},
	}

	for _, w := range weights {
		assert.NotEmpty(t, w.Weight)
		// In real impl, would parse and sum weights
	}

	// Weights should sum to 1.0 in production
	assert.Len(t, weights, 3)
}
