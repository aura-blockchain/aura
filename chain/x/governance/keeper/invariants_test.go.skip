package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/governance/types"
)

type InvariantsTestSuite struct {
	suite.Suite

	Keeper *Keeper
	SdkCtx sdk.Context
}

func (suite *InvariantsTestSuite) SetupTest() {
	keeper, ctx := setupKeeper(suite.T())
	suite.Keeper = keeper
	suite.SdkCtx = ctx

	// Set default params
	params := types.DefaultParams()
	suite.Keeper.SetParams(suite.SdkCtx, params)
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

// Test all invariants on empty store
func (suite *InvariantsTestSuite) TestAllInvariantsEmptyStore() {
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

// Test invariant registration
func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	registry := sdk.NewInvariantRegistry()
	suite.NotPanics(func() {
		RegisterInvariants(registry, suite.Keeper)
	})
}

// ============================================================================
// ParamsInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestParamsInvariantValid() {
	inv := ParamsInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "params invariant should pass with default params")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestParamsInvariantInvalid() {
	// Set invalid params (negative min deposit)
	params := suite.Keeper.GetParams(suite.SdkCtx)
	params.MinDeposit = "-100" // Invalid
	suite.Keeper.SetParams(suite.SdkCtx, params)

	inv := ParamsInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "params invariant should fail with invalid params")
	suite.NotEmpty(msg)
}

// ============================================================================
// ProposalValidityInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestProposalValidityInvariantValid() {
	// Create a valid proposal
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal",
		Description: "Test Description",
		Proposer:    genTestAddr().String(),
		Status:      types.StatusDepositPeriod,
		SubmitTime:  timestamppb.Now(),
	}
	err := suite.Keeper.SetProposal(suite.SdkCtx, proposal)
	suite.Require().NoError(err)

	inv := ProposalValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "proposal validity invariant should pass with valid proposal")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestProposalValidityInvariantZeroID() {
	// Create proposal with zero ID
	proposal := &types.Proposal{
		Id:          0, // Invalid
		Title:       "Test",
		Description: "Test",
		Proposer:    genTestAddr().String(),
		Status:      types.StatusDepositPeriod,
		SubmitTime:  timestamppb.Now(),
	}
	err := suite.Keeper.SetProposal(suite.SdkCtx, proposal)
	suite.Require().NoError(err)

	inv := ProposalValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "proposal validity invariant should fail with zero ID")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestProposalValidityInvariantInvalidProposer() {
	// Create proposal with invalid proposer address
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test",
		Description: "Test",
		Proposer:    "invalid-address", // Invalid
		Status:      types.StatusDepositPeriod,
		SubmitTime:  timestamppb.Now(),
	}
	err := suite.Keeper.SetProposal(suite.SdkCtx, proposal)
	suite.Require().NoError(err)

	inv := ProposalValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "proposal validity invariant should fail with invalid proposer")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestProposalValidityInvariantEmptyTitle() {
	// Create proposal with empty title
	proposal := &types.Proposal{
		Id:          1,
		Title:       "", // Invalid
		Description: "Test",
		Proposer:    genTestAddr().String(),
		Status:      types.StatusDepositPeriod,
		SubmitTime:  timestamppb.Now(),
	}
	err := suite.Keeper.SetProposal(suite.SdkCtx, proposal)
	suite.Require().NoError(err)

	inv := ProposalValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "proposal validity invariant should fail with empty title")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestProposalValidityInvariantInvalidStatus() {
	// Create proposal with invalid status
	proposal := &types.Proposal{
		Id:          1,
		Title:       "Test",
		Description: "Test",
		Proposer:    genTestAddr().String(),
		Status:      999, // Invalid status
		SubmitTime:  timestamppb.Now(),
	}
	err := suite.Keeper.SetProposal(suite.SdkCtx, proposal)
	suite.Require().NoError(err)

	inv := ProposalValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "proposal validity invariant should fail with invalid status")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestProposalValidityInvariantActiveWithoutVotingPeriod() {
	// Create active proposal without voting period
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test",
		Description:     "Test",
		Proposer:        genTestAddr().String(),
		Status:          types.StatusVotingPeriod,
		SubmitTime:      timestamppb.Now(),
		VotingStartTime: nil, // Invalid for active
		VotingEndTime:   nil, // Invalid for active
	}
	err := suite.Keeper.SetProposal(suite.SdkCtx, proposal)
	suite.Require().NoError(err)

	inv := ProposalValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "proposal validity invariant should fail for active proposal without voting period")
	suite.NotEmpty(msg)
}

// ============================================================================
// VoteConsistencyInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestVoteConsistencyInvariantValid() {
	// Create a valid vote
	vote := &types.Vote{
		ProposalId:  1,
		Voter:       genTestAddr().String(),
		Option:      types.OptionYes,
		VotingPower: "1000",
		Timestamp:   timestamppb.Now(),
	}
	err := suite.Keeper.SetVote(suite.SdkCtx, vote)
	suite.Require().NoError(err)

	inv := VoteConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "vote consistency invariant should pass with valid vote")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestVoteConsistencyInvariantZeroProposalID() {
	// Create vote with zero proposal ID
	vote := &types.Vote{
		ProposalId:  0, // Invalid
		Voter:       genTestAddr().String(),
		Option:      types.OptionYes,
		VotingPower: "1000",
		Timestamp:   timestamppb.Now(),
	}
	err := suite.Keeper.SetVote(suite.SdkCtx, vote)
	suite.Require().NoError(err)

	inv := VoteConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "vote consistency invariant should fail with zero proposal ID")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestVoteConsistencyInvariantInvalidVoter() {
	// Create vote with invalid voter address
	vote := &types.Vote{
		ProposalId:  1,
		Voter:       "invalid-address", // Invalid
		Option:      types.OptionYes,
		VotingPower: "1000",
		Timestamp:   timestamppb.Now(),
	}
	err := suite.Keeper.SetVote(suite.SdkCtx, vote)
	suite.Require().NoError(err)

	inv := VoteConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "vote consistency invariant should fail with invalid voter")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestVoteConsistencyInvariantInvalidOption() {
	// Create vote with invalid option
	vote := &types.Vote{
		ProposalId:  1,
		Voter:       genTestAddr().String(),
		Option:      999, // Invalid option
		VotingPower: "1000",
		Timestamp:   timestamppb.Now(),
	}
	err := suite.Keeper.SetVote(suite.SdkCtx, vote)
	suite.Require().NoError(err)

	inv := VoteConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "vote consistency invariant should fail with invalid option")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestVoteConsistencyInvariantInvalidVotingPower() {
	// Create vote with invalid voting power
	vote := &types.Vote{
		ProposalId:  1,
		Voter:       genTestAddr().String(),
		Option:      types.OptionYes,
		VotingPower: "-100", // Invalid (negative)
		Timestamp:   timestamppb.Now(),
	}
	err := suite.Keeper.SetVote(suite.SdkCtx, vote)
	suite.Require().NoError(err)

	inv := VoteConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "vote consistency invariant should fail with invalid voting power")
	suite.NotEmpty(msg)
}

// ============================================================================
// DepositConsistencyInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestDepositConsistencyInvariantValid() {
	// Create a valid deposit
	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  genTestAddr().String(),
		Amount:     "1000",
		Timestamp:  timestamppb.Now(),
	}
	err := suite.Keeper.SetDeposit(suite.SdkCtx, deposit)
	suite.Require().NoError(err)

	inv := DepositConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "deposit consistency invariant should pass with valid deposit")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestDepositConsistencyInvariantZeroProposalID() {
	// Create deposit with zero proposal ID
	deposit := &types.Deposit{
		ProposalId: 0, // Invalid
		Depositor:  genTestAddr().String(),
		Amount:     "1000",
		Timestamp:  timestamppb.Now(),
	}
	err := suite.Keeper.SetDeposit(suite.SdkCtx, deposit)
	suite.Require().NoError(err)

	inv := DepositConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "deposit consistency invariant should fail with zero proposal ID")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestDepositConsistencyInvariantInvalidDepositor() {
	// Create deposit with invalid depositor address
	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  "invalid-address", // Invalid
		Amount:     "1000",
		Timestamp:  timestamppb.Now(),
	}
	err := suite.Keeper.SetDeposit(suite.SdkCtx, deposit)
	suite.Require().NoError(err)

	inv := DepositConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "deposit consistency invariant should fail with invalid depositor")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestDepositConsistencyInvariantInvalidAmount() {
	// Create deposit with invalid amount
	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  genTestAddr().String(),
		Amount:     "-100", // Invalid (negative)
		Timestamp:  timestamppb.Now(),
	}
	err := suite.Keeper.SetDeposit(suite.SdkCtx, deposit)
	suite.Require().NoError(err)

	inv := DepositConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "deposit consistency invariant should fail with invalid amount")
	suite.NotEmpty(msg)
}

// ============================================================================
// VotingPowerConsistencyInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestVotingPowerConsistencyInvariantValid() {
	// Create proposal with voting power
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test",
		Description:     "Test",
		Proposer:        genTestAddr().String(),
		Status:          types.StatusVotingPeriod,
		SubmitTime:      timestamppb.Now(),
		VotingStartTime: timestamppb.Now(),
		VotingEndTime:   timestamppb.New(time.Now().Add(24 * time.Hour)),
		FinalTallyResult: &types.TallyResult{
			Yes:        "2000",
			No:         "0",
			Abstain:    "0",
			NoWithVeto: "0",
		},
	}
	err := suite.Keeper.SetProposal(suite.SdkCtx, proposal)
	suite.Require().NoError(err)

	// Create matching votes
	vote1 := &types.Vote{
		ProposalId:  1,
		Voter:       genTestAddr().String(),
		Option:      types.OptionYes,
		VotingPower: "1000",
		Timestamp:   timestamppb.Now(),
	}
	err = suite.Keeper.SetVote(suite.SdkCtx, vote1)
	suite.Require().NoError(err)

	vote2 := &types.Vote{
		ProposalId:  1,
		Voter:       genTestAddr().String(),
		Option:      types.OptionYes,
		VotingPower: "1000",
		Timestamp:   timestamppb.Now(),
	}
	err = suite.Keeper.SetVote(suite.SdkCtx, vote2)
	suite.Require().NoError(err)

	inv := VotingPowerConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "voting power consistency invariant should pass when totals match")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestVotingPowerConsistencyInvariantMismatch() {
	// Create proposal with mismatched voting power
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Test",
		Description:     "Test",
		Proposer:        genTestAddr().String(),
		Status:          types.StatusVotingPeriod,
		SubmitTime:      timestamppb.Now(),
		VotingStartTime: timestamppb.Now(),
		VotingEndTime:   timestamppb.New(time.Now().Add(24 * time.Hour)),
		FinalTallyResult: &types.TallyResult{
			Yes:        "5000", // Doesn't match actual votes
			No:         "0",
			Abstain:    "0",
			NoWithVeto: "0",
		},
	}
	err := suite.Keeper.SetProposal(suite.SdkCtx, proposal)
	suite.Require().NoError(err)

	// Create votes with different total
	vote := &types.Vote{
		ProposalId:  1,
		Voter:       genTestAddr().String(),
		Option:      types.OptionYes,
		VotingPower: "1000",
		Timestamp:   timestamppb.Now(),
	}
	err = suite.Keeper.SetVote(suite.SdkCtx, vote)
	suite.Require().NoError(err)

	inv := VotingPowerConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "voting power consistency invariant should fail when totals don't match")
	suite.NotEmpty(msg)
}

// ============================================================================
// All Invariants Integration Test
// ============================================================================

func (suite *InvariantsTestSuite) TestAllInvariantsWithValidData() {
	// Setup valid data
	proposal := &types.Proposal{
		Id:              1,
		Title:           "Integration Test Proposal",
		Description:     "Testing all invariants together",
		Proposer:        genTestAddr().String(),
		Status:          types.StatusVotingPeriod,
		SubmitTime:      timestamppb.Now(),
		VotingStartTime: timestamppb.Now(),
		VotingEndTime:   timestamppb.New(time.Now().Add(24 * time.Hour)),
		FinalTallyResult: &types.TallyResult{
			Yes:        "1000",
			No:         "0",
			Abstain:    "0",
			NoWithVeto: "0",
		},
	}
	err := suite.Keeper.SetProposal(suite.SdkCtx, proposal)
	suite.Require().NoError(err)

	vote := &types.Vote{
		ProposalId:  1,
		Voter:       genTestAddr().String(),
		Option:      types.OptionYes,
		VotingPower: "1000",
		Timestamp:   timestamppb.Now(),
	}
	err = suite.Keeper.SetVote(suite.SdkCtx, vote)
	suite.Require().NoError(err)

	deposit := &types.Deposit{
		ProposalId: 1,
		Depositor:  genTestAddr().String(),
		Amount:     "5000",
		Timestamp:  timestamppb.Now(),
	}
	err = suite.Keeper.SetDeposit(suite.SdkCtx, deposit)
	suite.Require().NoError(err)

	// Run all invariants
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "all invariants should pass with valid data")
	suite.Empty(msg)
}

// Helper function to generate test address
func genTestAddr() sdk.AccAddress {
	return sdk.AccAddress("test_address_____")
}
