package keeper

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/economics/types"
	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
)

// ============================
// TEST SUITE SETUP
// ============================

type MsgServerTestSuite struct {
	suite.Suite
	ctx       sdk.Context
	keeper    *Keeper
	msgServer economicspb.MsgServer
	authority string

	// Test accounts
	testAddrs []sdk.AccAddress
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (suite *MsgServerTestSuite) SetupTest() {
	// Create test context
	key := storetypes.NewKVStoreKey("economics")
	storeService := runtime.NewKVStoreService(key)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	suite.Require().NoError(err)

	suite.ctx = sdk.NewContext(stateStore, cmtproto.Header{
		Time: time.Now(),
	}, false, nil)

	// Create codec
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create keeper with aura prefix
	suite.authority = "aura1w3jhxap3ta047h6lta047h6lta047h6la3zjcr"
	suite.keeper = NewKeeper(cdc, storeService, suite.authority)

	// Initialize default params
	params := types.DefaultParams()
	// Set MinDeposit to 1,000,000 uaura to ensure proposals need deposits
	params.Governance.MinDeposit = sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(1000000)))
	err = suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Create msg server
	suite.msgServer = NewMsgServer(suite.keeper)

	// Create test accounts
	suite.testAddrs = []sdk.AccAddress{
		sdk.AccAddress("test_addr_1___________"),
		sdk.AccAddress("test_addr_2___________"),
		sdk.AccAddress("test_addr_3___________"),
		sdk.AccAddress("test_addr_4___________"),
		sdk.AccAddress("test_addr_5___________"),
	}
}

// ============================
// VESTING SCHEDULE TESTS
// ============================

func (suite *MsgServerTestSuite) TestMsgCreateVestingSchedule() {
	tests := []struct {
		name      string
		msg       *economicspb.MsgCreateVestingSchedule
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid linear vesting schedule",
			msg: &economicspb.MsgCreateVestingSchedule{
				Creator:            suite.testAddrs[0].String(),
				BeneficiaryAddress: suite.testAddrs[1].String(),
				TotalAmount: sdk.Coin{
					Denom:  "uaura",
					Amount: sdkmath.NewInt(1000000),
				},
				StartTime:       time.Now(),
				CliffDuration:   30 * 24 * 3600,  // 30 days
				VestingDuration: 365 * 24 * 3600, // 1 year
				VestingType:     economicspb.VestingType_VESTING_TYPE_LINEAR,
				ScheduleType:    economicspb.ScheduleType_SCHEDULE_TYPE_TEAM,
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty creator",
			msg: &economicspb.MsgCreateVestingSchedule{
				Creator:            "",
				BeneficiaryAddress: suite.testAddrs[1].String(),
				TotalAmount: sdk.Coin{
					Denom:  "uaura",
					Amount: sdkmath.NewInt(1000000),
				},
				StartTime:       time.Now(),
				VestingDuration: 365 * 24 * 3600,
				VestingType:     economicspb.VestingType_VESTING_TYPE_LINEAR,
			},
			shouldErr: true,
			errMsg:    "invalid creator address",
		},
		{
			name: "invalid - empty beneficiary",
			msg: &economicspb.MsgCreateVestingSchedule{
				Creator:            suite.testAddrs[0].String(),
				BeneficiaryAddress: "",
				TotalAmount: sdk.Coin{
					Denom:  "uaura",
					Amount: sdkmath.NewInt(1000000),
				},
				StartTime:       time.Now(),
				VestingDuration: 365 * 24 * 3600,
				VestingType:     economicspb.VestingType_VESTING_TYPE_LINEAR,
			},
			shouldErr: true,
			errMsg:    "invalid beneficiary address",
		},
		{
			name: "invalid - zero amount",
			msg: &economicspb.MsgCreateVestingSchedule{
				Creator:            suite.testAddrs[0].String(),
				BeneficiaryAddress: suite.testAddrs[1].String(),
				TotalAmount: sdk.Coin{
					Denom:  "uaura",
					Amount: sdkmath.ZeroInt(),
				},
				StartTime:       time.Now(),
				VestingDuration: 365 * 24 * 3600,
				VestingType:     economicspb.VestingType_VESTING_TYPE_LINEAR,
			},
			shouldErr: true,
			errMsg:    "invalid amount",
		},
		{
			name: "invalid - zero vesting duration",
			msg: &economicspb.MsgCreateVestingSchedule{
				Creator:            suite.testAddrs[0].String(),
				BeneficiaryAddress: suite.testAddrs[1].String(),
				TotalAmount: sdk.Coin{
					Denom:  "uaura",
					Amount: sdkmath.NewInt(1000000),
				},
				StartTime:       time.Now(),
				VestingDuration: 0,
				VestingType:     economicspb.VestingType_VESTING_TYPE_LINEAR,
			},
			shouldErr: true,
			errMsg:    "invalid vesting duration",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			goCtx := sdk.WrapSDKContext(suite.ctx)
			resp, err := suite.msgServer.CreateVestingSchedule(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().NotEmpty(resp.ScheduleId)
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgReleaseVestedTokens() {
	// First create a vesting schedule
	createMsg := &economicspb.MsgCreateVestingSchedule{
		Creator:            suite.testAddrs[0].String(),
		BeneficiaryAddress: suite.testAddrs[1].String(),
		TotalAmount: sdk.Coin{
			Denom:  "uaura",
			Amount: sdkmath.NewInt(1000000),
		},
		StartTime:       time.Now().Add(-100 * 24 * time.Hour), // Started 100 days ago
		CliffDuration:   0,
		VestingDuration: 365 * 24 * 3600,
		VestingType:     economicspb.VestingType_VESTING_TYPE_LINEAR,
		ScheduleType:    economicspb.ScheduleType_SCHEDULE_TYPE_TEAM,
	}

	goCtx := sdk.WrapSDKContext(suite.ctx)
	createResp, err := suite.msgServer.CreateVestingSchedule(goCtx, createMsg)
	suite.Require().NoError(err)
	scheduleID := createResp.ScheduleId

	tests := []struct {
		name      string
		msg       *economicspb.MsgReleaseVestedTokens
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid release",
			msg: &economicspb.MsgReleaseVestedTokens{
				Beneficiary: suite.testAddrs[1].String(),
				ScheduleId:  scheduleID,
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty beneficiary",
			msg: &economicspb.MsgReleaseVestedTokens{
				Beneficiary: "",
				ScheduleId:  scheduleID,
			},
			shouldErr: true,
			errMsg:    "invalid beneficiary",
		},
		{
			name: "invalid - empty schedule ID",
			msg: &economicspb.MsgReleaseVestedTokens{
				Beneficiary: suite.testAddrs[1].String(),
				ScheduleId:  "",
			},
			shouldErr: true,
			errMsg:    "invalid schedule ID",
		},
		{
			name: "invalid - non-existent schedule",
			msg: &economicspb.MsgReleaseVestedTokens{
				Beneficiary: suite.testAddrs[1].String(),
				ScheduleId:  "non-existent-id",
			},
			shouldErr: true,
			errMsg:    "schedule not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.msgServer.ReleaseVestedTokens(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				// Amount released should be positive (some tokens vested after 100 days)
				suite.Require().True(resp.AmountReleased.Amount.GT(sdkmath.ZeroInt()))
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgRevokeVestingSchedule() {
	// First create a vesting schedule
	createMsg := &economicspb.MsgCreateVestingSchedule{
		Creator:            suite.testAddrs[0].String(),
		BeneficiaryAddress: suite.testAddrs[1].String(),
		TotalAmount: sdk.Coin{
			Denom:  "uaura",
			Amount: sdkmath.NewInt(1000000),
		},
		StartTime:       time.Now(),
		VestingDuration: 365 * 24 * 3600,
		VestingType:     economicspb.VestingType_VESTING_TYPE_LINEAR,
		ScheduleType:    economicspb.ScheduleType_SCHEDULE_TYPE_TEAM,
	}

	goCtx := sdk.WrapSDKContext(suite.ctx)
	createResp, err := suite.msgServer.CreateVestingSchedule(goCtx, createMsg)
	suite.Require().NoError(err)
	scheduleID := createResp.ScheduleId

	tests := []struct {
		name      string
		msg       *economicspb.MsgRevokeVestingSchedule
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid revocation",
			msg: &economicspb.MsgRevokeVestingSchedule{
				Revoker:    suite.testAddrs[0].String(),
				ScheduleId: scheduleID,
				Reason:     "Employee left the company",
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty revoker",
			msg: &economicspb.MsgRevokeVestingSchedule{
				Revoker:    "",
				ScheduleId: scheduleID,
				Reason:     "Test revocation",
			},
			shouldErr: true,
			errMsg:    "invalid revoker",
		},
		{
			name: "invalid - non-existent schedule",
			msg: &economicspb.MsgRevokeVestingSchedule{
				Revoker:    suite.testAddrs[0].String(),
				ScheduleId: "non-existent",
				Reason:     "Test",
			},
			shouldErr: true,
			errMsg:    "schedule not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.msgServer.RevokeVestingSchedule(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
			}
		})
	}
}

// ============================
// GOVERNANCE TESTS
// ============================

func (suite *MsgServerTestSuite) TestMsgSubmitProposal() {
	tests := []struct {
		name      string
		msg       *economicspb.MsgSubmitProposal
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid text proposal",
			msg: &economicspb.MsgSubmitProposal{
				Title:          "Test Proposal",
				Description:    "This is a test proposal for the economics module",
				Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
				Proposer:       suite.testAddrs[0].String(),
				InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(1000000)}},
				IsEmergency:    false,
			},
			shouldErr: false,
		},
		{
			name: "valid parameter change proposal",
			msg: &economicspb.MsgSubmitProposal{
				Title:          "Update Fee Parameters",
				Description:    "Proposal to adjust fee parameters",
				Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE,
				Proposer:       suite.testAddrs[0].String(),
				InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(1000000)}},
				IsEmergency:    false,
			},
			shouldErr: false,
		},
		{
			name: "valid emergency proposal",
			msg: &economicspb.MsgSubmitProposal{
				Title:          "Emergency Security Fix",
				Description:    "Emergency proposal to fix security issue",
				Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_EMERGENCY,
				Proposer:       suite.testAddrs[0].String(),
				InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(5000000)}},
				IsEmergency:    true,
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty title",
			msg: &economicspb.MsgSubmitProposal{
				Title:          "",
				Description:    "Test description",
				Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
				Proposer:       suite.testAddrs[0].String(),
				InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(1000000)}},
			},
			shouldErr: true,
			errMsg:    "invalid title",
		},
		{
			name: "invalid - empty description",
			msg: &economicspb.MsgSubmitProposal{
				Title:          "Test Proposal",
				Description:    "",
				Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
				Proposer:       suite.testAddrs[0].String(),
				InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(1000000)}},
			},
			shouldErr: true,
			errMsg:    "invalid description",
		},
		{
			name: "invalid - empty proposer",
			msg: &economicspb.MsgSubmitProposal{
				Title:          "Test Proposal",
				Description:    "Test description",
				Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
				Proposer:       "",
				InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(1000000)}},
			},
			shouldErr: true,
			errMsg:    "invalid proposer",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			goCtx := sdk.WrapSDKContext(suite.ctx)
			resp, err := suite.msgServer.SubmitProposal(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Greater(resp.ProposalId, uint64(0))
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgDeposit() {
	// First create a proposal
	submitMsg := &economicspb.MsgSubmitProposal{
		Title:          "Test Proposal for Deposit",
		Description:    "Test proposal description",
		Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		Proposer:       suite.testAddrs[0].String(),
		InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(500000)}},
	}

	goCtx := sdk.WrapSDKContext(suite.ctx)
	submitResp, err := suite.msgServer.SubmitProposal(goCtx, submitMsg)
	suite.Require().NoError(err)
	proposalID := submitResp.ProposalId

	tests := []struct {
		name      string
		msg       *economicspb.MsgDeposit
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid deposit",
			msg: &economicspb.MsgDeposit{
				ProposalId: proposalID,
				Depositor:  suite.testAddrs[1].String(),
				Amount:     sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(500000)}},
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty depositor",
			msg: &economicspb.MsgDeposit{
				ProposalId: proposalID,
				Depositor:  "",
				Amount:     sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(100000)}},
			},
			shouldErr: true,
			errMsg:    "invalid depositor",
		},
		{
			name: "invalid - zero amount",
			msg: &economicspb.MsgDeposit{
				ProposalId: proposalID,
				Depositor:  suite.testAddrs[1].String(),
				Amount:     sdk.Coins{},
			},
			shouldErr: true,
			errMsg:    "invalid deposit amount",
		},
		{
			name: "invalid - non-existent proposal",
			msg: &economicspb.MsgDeposit{
				ProposalId: 99999,
				Depositor:  suite.testAddrs[1].String(),
				Amount:     sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(100000)}},
			},
			shouldErr: true,
			errMsg:    "proposal not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.msgServer.Deposit(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgVote() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// First lock tokens for each voter to give them voting power
	for i := 1; i <= 4; i++ {
		lockMsg := &economicspb.MsgLockVotingTokens{
			Owner: suite.testAddrs[i].String(),
			Amount: sdk.Coin{
				Denom:  "uaura",
				Amount: sdkmath.NewInt(1000000),
			},
			LockDuration: 365 * 24 * 3600, // 1 year
		}
		_, err := suite.msgServer.LockVotingTokens(goCtx, lockMsg)
		suite.Require().NoError(err)
	}

	// Create and deposit enough to activate voting
	submitMsg := &economicspb.MsgSubmitProposal{
		Title:          "Test Proposal for Voting",
		Description:    "Test proposal description",
		Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		Proposer:       suite.testAddrs[0].String(),
		InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(10000000)}}, // Large deposit to activate voting
	}

	submitResp, err := suite.msgServer.SubmitProposal(goCtx, submitMsg)
	suite.Require().NoError(err)
	proposalID := submitResp.ProposalId

	tests := []struct {
		name      string
		msg       *economicspb.MsgVote
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid yes vote",
			msg: &economicspb.MsgVote{
				ProposalId: proposalID,
				Voter:      suite.testAddrs[1].String(),
				Option:     economicspb.VoteOption_VOTE_OPTION_YES,
				IsSecret:   false,
			},
			shouldErr: false,
		},
		{
			name: "valid no vote",
			msg: &economicspb.MsgVote{
				ProposalId: proposalID,
				Voter:      suite.testAddrs[2].String(),
				Option:     economicspb.VoteOption_VOTE_OPTION_NO,
				IsSecret:   false,
			},
			shouldErr: false,
		},
		{
			name: "valid abstain vote",
			msg: &economicspb.MsgVote{
				ProposalId: proposalID,
				Voter:      suite.testAddrs[3].String(),
				Option:     economicspb.VoteOption_VOTE_OPTION_ABSTAIN,
				IsSecret:   false,
			},
			shouldErr: false,
		},
		{
			name: "valid no with veto vote",
			msg: &economicspb.MsgVote{
				ProposalId: proposalID,
				Voter:      suite.testAddrs[4].String(),
				Option:     economicspb.VoteOption_VOTE_OPTION_NO_WITH_VETO,
				IsSecret:   false,
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty voter",
			msg: &economicspb.MsgVote{
				ProposalId: proposalID,
				Voter:      "",
				Option:     economicspb.VoteOption_VOTE_OPTION_YES,
			},
			shouldErr: true,
			errMsg:    "invalid voter",
		},
		{
			name: "invalid - unspecified option",
			msg: &economicspb.MsgVote{
				ProposalId: proposalID,
				Voter:      suite.testAddrs[1].String(),
				Option:     economicspb.VoteOption_VOTE_OPTION_UNSPECIFIED,
			},
			shouldErr: true,
			errMsg:    "invalid vote option",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.msgServer.Vote(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgVoteWeighted() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// First lock tokens for the voter to give them voting power
	lockMsg := &economicspb.MsgLockVotingTokens{
		Owner: suite.testAddrs[1].String(),
		Amount: sdk.Coin{
			Denom:  "uaura",
			Amount: sdkmath.NewInt(1000000),
		},
		LockDuration: 365 * 24 * 3600, // 1 year
	}
	_, err := suite.msgServer.LockVotingTokens(goCtx, lockMsg)
	suite.Require().NoError(err)

	// Create proposal
	submitMsg := &economicspb.MsgSubmitProposal{
		Title:          "Test Proposal for Weighted Voting",
		Description:    "Test proposal description",
		Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		Proposer:       suite.testAddrs[0].String(),
		InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(10000000)}},
	}

	submitResp, err := suite.msgServer.SubmitProposal(goCtx, submitMsg)
	suite.Require().NoError(err)
	proposalID := submitResp.ProposalId

	tests := []struct {
		name      string
		msg       *economicspb.MsgVoteWeighted
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid weighted vote",
			msg: &economicspb.MsgVoteWeighted{
				ProposalId: proposalID,
				Voter:      suite.testAddrs[1].String(),
				Options: []economicspb.WeightedVoteOption{
					{
						Option: economicspb.VoteOption_VOTE_OPTION_YES,
						Weight: sdkmath.LegacyMustNewDecFromStr("0.7"),
					},
					{
						Option: economicspb.VoteOption_VOTE_OPTION_ABSTAIN,
						Weight: sdkmath.LegacyMustNewDecFromStr("0.3"),
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty voter",
			msg: &economicspb.MsgVoteWeighted{
				ProposalId: proposalID,
				Voter:      "",
				Options: []economicspb.WeightedVoteOption{
					{
						Option: economicspb.VoteOption_VOTE_OPTION_YES,
						Weight: sdkmath.LegacyMustNewDecFromStr("1.0"),
					},
				},
			},
			shouldErr: true,
			errMsg:    "invalid voter",
		},
		{
			name: "invalid - empty options",
			msg: &economicspb.MsgVoteWeighted{
				ProposalId: proposalID,
				Voter:      suite.testAddrs[1].String(),
				Options:    []economicspb.WeightedVoteOption{},
			},
			shouldErr: true,
			errMsg:    "invalid vote options",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.msgServer.VoteWeighted(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgDelegateVote() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// First lock tokens for the delegator to give them voting power
	lockMsg := &economicspb.MsgLockVotingTokens{
		Owner: suite.testAddrs[0].String(),
		Amount: sdk.Coin{
			Denom:  "uaura",
			Amount: sdkmath.NewInt(1000000),
		},
		LockDuration: 365 * 24 * 3600, // 1 year
	}
	_, err := suite.msgServer.LockVotingTokens(goCtx, lockMsg)
	suite.Require().NoError(err)

	tests := []struct {
		name      string
		msg       *economicspb.MsgDelegateVote
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid delegation - all categories",
			msg: &economicspb.MsgDelegateVote{
				Delegator:  suite.testAddrs[0].String(),
				Delegate:   suite.testAddrs[1].String(),
				Categories: []economicspb.ProposalCategory{}, // Empty = all categories
			},
			shouldErr: false,
		},
		{
			name: "valid delegation - specific categories",
			msg: &economicspb.MsgDelegateVote{
				Delegator: suite.testAddrs[0].String(),
				Delegate:  suite.testAddrs[1].String(),
				Categories: []economicspb.ProposalCategory{
					economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
					economicspb.ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE,
				},
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty delegator",
			msg: &economicspb.MsgDelegateVote{
				Delegator:  "",
				Delegate:   suite.testAddrs[1].String(),
				Categories: []economicspb.ProposalCategory{},
			},
			shouldErr: true,
			errMsg:    "invalid delegator",
		},
		{
			name: "invalid - empty delegate",
			msg: &economicspb.MsgDelegateVote{
				Delegator:  suite.testAddrs[0].String(),
				Delegate:   "",
				Categories: []economicspb.ProposalCategory{},
			},
			shouldErr: true,
			errMsg:    "invalid delegate",
		},
		{
			name: "invalid - self delegation",
			msg: &economicspb.MsgDelegateVote{
				Delegator:  suite.testAddrs[0].String(),
				Delegate:   suite.testAddrs[0].String(),
				Categories: []economicspb.ProposalCategory{},
			},
			shouldErr: true,
			errMsg:    "cannot delegate to self",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			goCtx := sdk.WrapSDKContext(suite.ctx)
			resp, err := suite.msgServer.DelegateVote(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgUndelegateVote() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// First lock tokens for the delegator to give them voting power
	lockMsg := &economicspb.MsgLockVotingTokens{
		Owner: suite.testAddrs[0].String(),
		Amount: sdk.Coin{
			Denom:  "uaura",
			Amount: sdkmath.NewInt(1000000),
		},
		LockDuration: 365 * 24 * 3600, // 1 year
	}
	_, err := suite.msgServer.LockVotingTokens(goCtx, lockMsg)
	suite.Require().NoError(err)

	// Create a delegation
	delegateMsg := &economicspb.MsgDelegateVote{
		Delegator:  suite.testAddrs[0].String(),
		Delegate:   suite.testAddrs[1].String(),
		Categories: []economicspb.ProposalCategory{},
	}

	_, err = suite.msgServer.DelegateVote(goCtx, delegateMsg)
	suite.Require().NoError(err)

	tests := []struct {
		name      string
		msg       *economicspb.MsgUndelegateVote
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid undelegation",
			msg: &economicspb.MsgUndelegateVote{
				Delegator:  suite.testAddrs[0].String(),
				Delegate:   suite.testAddrs[1].String(),
				Categories: []economicspb.ProposalCategory{},
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty delegator",
			msg: &economicspb.MsgUndelegateVote{
				Delegator:  "",
				Delegate:   suite.testAddrs[1].String(),
				Categories: []economicspb.ProposalCategory{},
			},
			shouldErr: true,
			errMsg:    "invalid delegator",
		},
		{
			name: "invalid - non-existent delegation",
			msg: &economicspb.MsgUndelegateVote{
				Delegator:  suite.testAddrs[2].String(),
				Delegate:   suite.testAddrs[3].String(),
				Categories: []economicspb.ProposalCategory{},
			},
			shouldErr: true,
			errMsg:    "delegation not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.msgServer.UndelegateVote(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgExecuteProposal() {
	// Create a proposal that will pass
	submitMsg := &economicspb.MsgSubmitProposal{
		Title:          "Test Proposal for Execution",
		Description:    "Test proposal description",
		Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		Proposer:       suite.testAddrs[0].String(),
		InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(10000000)}},
	}

	goCtx := sdk.WrapSDKContext(suite.ctx)
	submitResp, err := suite.msgServer.SubmitProposal(goCtx, submitMsg)
	suite.Require().NoError(err)
	proposalID := submitResp.ProposalId

	tests := []struct {
		name      string
		msg       *economicspb.MsgExecuteProposal
		shouldErr bool
		errMsg    string
	}{
		{
			name: "invalid - proposal not in passed status",
			msg: &economicspb.MsgExecuteProposal{
				ProposalId: proposalID,
				Executor:   suite.testAddrs[0].String(),
			},
			shouldErr: true,
			errMsg:    "proposal not in passed status",
		},
		{
			name: "invalid - empty executor",
			msg: &economicspb.MsgExecuteProposal{
				ProposalId: proposalID,
				Executor:   "",
			},
			shouldErr: true,
			errMsg:    "invalid executor",
		},
		{
			name: "invalid - non-existent proposal",
			msg: &economicspb.MsgExecuteProposal{
				ProposalId: 99999,
				Executor:   suite.testAddrs[0].String(),
			},
			shouldErr: true,
			errMsg:    "proposal not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.msgServer.ExecuteProposal(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgRevealSecretVote() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// First lock tokens for the voter to give them voting power
	lockMsg := &economicspb.MsgLockVotingTokens{
		Owner: suite.testAddrs[1].String(),
		Amount: sdk.Coin{
			Denom:  "uaura",
			Amount: sdkmath.NewInt(1000000),
		},
		LockDuration: 365 * 24 * 3600, // 1 year
	}
	_, err := suite.msgServer.LockVotingTokens(goCtx, lockMsg)
	suite.Require().NoError(err)

	// Create proposal with secret ballot enabled
	submitMsg := &economicspb.MsgSubmitProposal{
		Title:          "Test Proposal for Secret Voting",
		Description:    "Test proposal description",
		Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		Proposer:       suite.testAddrs[0].String(),
		InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(10000000)}},
	}

	submitResp, err := suite.msgServer.SubmitProposal(goCtx, submitMsg)
	suite.Require().NoError(err)
	proposalID := submitResp.ProposalId

	// Cast a secret vote
	// Generate a proper commitment: SHA256 hash of "{option}:{revealKey}"
	revealKey := "reveal_key_123"
	voteOption := economicspb.VoteOption_VOTE_OPTION_YES
	commitmentData := fmt.Sprintf("%d:%s", voteOption, revealKey)
	commitment := fmt.Sprintf("%x", sha256.Sum256([]byte(commitmentData)))

	voteMsg := &economicspb.MsgVote{
		ProposalId:     proposalID,
		Voter:          suite.testAddrs[1].String(),
		Option:         voteOption,
		IsSecret:       true,
		VoteCommitment: commitment,
	}
	_, err = suite.msgServer.Vote(goCtx, voteMsg)
	suite.Require().NoError(err)

	tests := []struct {
		name      string
		msg       *economicspb.MsgRevealSecretVote
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid reveal",
			msg: &economicspb.MsgRevealSecretVote{
				ProposalId: proposalID,
				Voter:      suite.testAddrs[1].String(),
				Option:     economicspb.VoteOption_VOTE_OPTION_YES,
				RevealKey:  "reveal_key_123",
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty voter",
			msg: &economicspb.MsgRevealSecretVote{
				ProposalId: proposalID,
				Voter:      "",
				Option:     economicspb.VoteOption_VOTE_OPTION_YES,
				RevealKey:  "reveal_key_123",
			},
			shouldErr: true,
			errMsg:    "invalid voter",
		},
		{
			name: "invalid - empty reveal key",
			msg: &economicspb.MsgRevealSecretVote{
				ProposalId: proposalID,
				Voter:      suite.testAddrs[1].String(),
				Option:     economicspb.VoteOption_VOTE_OPTION_YES,
				RevealKey:  "",
			},
			shouldErr: true,
			errMsg:    "invalid reveal key",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.msgServer.RevealSecretVote(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
			}
		})
	}
}

// ============================
// VOTE LOCK TESTS
// ============================

func (suite *MsgServerTestSuite) TestMsgLockVotingTokens() {
	tests := []struct {
		name      string
		msg       *economicspb.MsgLockVotingTokens
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid lock - 1 year",
			msg: &economicspb.MsgLockVotingTokens{
				Owner: suite.testAddrs[0].String(),
				Amount: sdk.Coin{
					Denom:  "uaura",
					Amount: sdkmath.NewInt(1000000),
				},
				LockDuration: 365 * 24 * 3600, // 1 year in seconds
			},
			shouldErr: false,
		},
		{
			name: "valid lock - 4 years",
			msg: &economicspb.MsgLockVotingTokens{
				Owner: suite.testAddrs[1].String(),
				Amount: sdk.Coin{
					Denom:  "uaura",
					Amount: sdkmath.NewInt(5000000),
				},
				LockDuration: 4 * 365 * 24 * 3600, // 4 years
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty owner",
			msg: &economicspb.MsgLockVotingTokens{
				Owner: "",
				Amount: sdk.Coin{
					Denom:  "uaura",
					Amount: sdkmath.NewInt(1000000),
				},
				LockDuration: 365 * 24 * 3600,
			},
			shouldErr: true,
			errMsg:    "invalid owner",
		},
		{
			name: "invalid - zero amount",
			msg: &economicspb.MsgLockVotingTokens{
				Owner: suite.testAddrs[0].String(),
				Amount: sdk.Coin{
					Denom:  "uaura",
					Amount: sdkmath.ZeroInt(),
				},
				LockDuration: 365 * 24 * 3600,
			},
			shouldErr: true,
			errMsg:    "invalid amount",
		},
		{
			name: "invalid - zero lock duration",
			msg: &economicspb.MsgLockVotingTokens{
				Owner: suite.testAddrs[0].String(),
				Amount: sdk.Coin{
					Denom:  "uaura",
					Amount: sdkmath.NewInt(1000000),
				},
				LockDuration: 0,
			},
			shouldErr: true,
			errMsg:    "invalid lock duration",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			goCtx := sdk.WrapSDKContext(suite.ctx)
			resp, err := suite.msgServer.LockVotingTokens(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().NotEmpty(resp.LockId)
				// Voting power should be greater than locked amount due to time multiplier
				suite.Require().True(resp.VotingPower.GTE(tt.msg.Amount.Amount))
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgUnlockVotingTokens() {
	// First create a lock
	lockMsg := &economicspb.MsgLockVotingTokens{
		Owner: suite.testAddrs[0].String(),
		Amount: sdk.Coin{
			Denom:  "uaura",
			Amount: sdkmath.NewInt(1000000),
		},
		LockDuration: 365 * 24 * 3600,
	}

	goCtx := sdk.WrapSDKContext(suite.ctx)
	lockResp, err := suite.msgServer.LockVotingTokens(goCtx, lockMsg)
	suite.Require().NoError(err)
	lockID := lockResp.LockId

	tests := []struct {
		name      string
		msg       *economicspb.MsgUnlockVotingTokens
		shouldErr bool
		errMsg    string
	}{
		{
			name: "invalid - lock period not ended",
			msg: &economicspb.MsgUnlockVotingTokens{
				Owner:  suite.testAddrs[0].String(),
				LockId: lockID,
			},
			shouldErr: true,
			errMsg:    "lock period not ended",
		},
		{
			name: "invalid - empty owner",
			msg: &economicspb.MsgUnlockVotingTokens{
				Owner:  "",
				LockId: lockID,
			},
			shouldErr: true,
			errMsg:    "invalid owner",
		},
		{
			name: "invalid - non-existent lock",
			msg: &economicspb.MsgUnlockVotingTokens{
				Owner:  suite.testAddrs[0].String(),
				LockId: "non-existent-lock",
			},
			shouldErr: true,
			errMsg:    "lock not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.msgServer.UnlockVotingTokens(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().True(resp.AmountUnlocked.Amount.GT(sdkmath.ZeroInt()))
			}
		})
	}
}

// ============================
// TREASURY TESTS
// ============================

func (suite *MsgServerTestSuite) TestMsgProposeTreasurySpend() {
	tests := []struct {
		name      string
		msg       *economicspb.MsgProposeTreasurySpend
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid treasury spend proposal",
			msg: &economicspb.MsgProposeTreasurySpend{
				Proposer:    suite.testAddrs[0].String(),
				Recipient:   suite.testAddrs[1].String(),
				Amount:      sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(100000)}},
				Description: "Development grant for team X",
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty proposer",
			msg: &economicspb.MsgProposeTreasurySpend{
				Proposer:    "",
				Recipient:   suite.testAddrs[1].String(),
				Amount:      sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(100000)}},
				Description: "Test",
			},
			shouldErr: true,
			errMsg:    "invalid proposer",
		},
		{
			name: "invalid - empty recipient",
			msg: &economicspb.MsgProposeTreasurySpend{
				Proposer:    suite.testAddrs[0].String(),
				Recipient:   "",
				Amount:      sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(100000)}},
				Description: "Test",
			},
			shouldErr: true,
			errMsg:    "invalid recipient",
		},
		{
			name: "invalid - zero amount",
			msg: &economicspb.MsgProposeTreasurySpend{
				Proposer:    suite.testAddrs[0].String(),
				Recipient:   suite.testAddrs[1].String(),
				Amount:      sdk.Coins{},
				Description: "Test",
			},
			shouldErr: true,
			errMsg:    "invalid amount",
		},
		{
			name: "invalid - empty description",
			msg: &economicspb.MsgProposeTreasurySpend{
				Proposer:  suite.testAddrs[0].String(),
				Recipient: suite.testAddrs[1].String(),
				Amount:    sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(100000)}},
			},
			shouldErr: true,
			errMsg:    "invalid description",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			goCtx := sdk.WrapSDKContext(suite.ctx)
			resp, err := suite.msgServer.ProposeTreasurySpend(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().NotEmpty(resp.TxId)
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgSignTreasurySpend() {
	// First create a treasury spend proposal
	proposeMsg := &economicspb.MsgProposeTreasurySpend{
		Proposer:    suite.testAddrs[0].String(),
		Recipient:   suite.testAddrs[1].String(),
		Amount:      sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(100000)}},
		Description: "Test treasury spend",
	}

	goCtx := sdk.WrapSDKContext(suite.ctx)
	proposeResp, err := suite.msgServer.ProposeTreasurySpend(goCtx, proposeMsg)
	suite.Require().NoError(err)
	txID := proposeResp.TxId

	tests := []struct {
		name      string
		msg       *economicspb.MsgSignTreasurySpend
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid signature",
			msg: &economicspb.MsgSignTreasurySpend{
				Signer: suite.testAddrs[0].String(),
				TxId:   txID,
			},
			shouldErr: false,
		},
		{
			name: "invalid - empty signer",
			msg: &economicspb.MsgSignTreasurySpend{
				Signer: "",
				TxId:   txID,
			},
			shouldErr: true,
			errMsg:    "invalid signer",
		},
		{
			name: "invalid - non-existent tx",
			msg: &economicspb.MsgSignTreasurySpend{
				Signer: suite.testAddrs[0].String(),
				TxId:   "non-existent-tx",
			},
			shouldErr: true,
			errMsg:    "transaction not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.msgServer.SignTreasurySpend(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Greater(resp.CurrentSignatures, uint32(0))
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgExecuteTreasurySpend() {
	// Create and sign a treasury spend
	proposeMsg := &economicspb.MsgProposeTreasurySpend{
		Proposer:    suite.testAddrs[0].String(),
		Recipient:   suite.testAddrs[1].String(),
		Amount:      sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(100000)}},
		Description: "Test treasury spend",
	}

	goCtx := sdk.WrapSDKContext(suite.ctx)
	proposeResp, err := suite.msgServer.ProposeTreasurySpend(goCtx, proposeMsg)
	suite.Require().NoError(err)
	txID := proposeResp.TxId

	tests := []struct {
		name      string
		msg       *economicspb.MsgExecuteTreasurySpend
		shouldErr bool
		errMsg    string
	}{
		{
			name: "invalid - insufficient signatures",
			msg: &economicspb.MsgExecuteTreasurySpend{
				Executor: suite.testAddrs[0].String(),
				TxId:     txID,
			},
			shouldErr: true,
			errMsg:    "insufficient signatures",
		},
		{
			name: "invalid - empty executor",
			msg: &economicspb.MsgExecuteTreasurySpend{
				Executor: "",
				TxId:     txID,
			},
			shouldErr: true,
			errMsg:    "invalid executor",
		},
		{
			name: "invalid - non-existent tx",
			msg: &economicspb.MsgExecuteTreasurySpend{
				Executor: suite.testAddrs[0].String(),
				TxId:     "non-existent",
			},
			shouldErr: true,
			errMsg:    "transaction not found",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			resp, err := suite.msgServer.ExecuteTreasurySpend(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().True(resp.Success)
			}
		})
	}
}

// ============================
// ADMIN TESTS
// ============================

func (suite *MsgServerTestSuite) TestMsgUpdateParams() {
	newParams := types.DefaultParams()
	newParams.Fees.BaseFee = sdkmath.NewInt(2000)

	tests := []struct {
		name      string
		msg       *economicspb.MsgUpdateParams
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid params update by authority",
			msg: &economicspb.MsgUpdateParams{
				Authority: suite.authority,
				Params:    *newParams,
			},
			shouldErr: false,
		},
		{
			name: "invalid - unauthorized caller",
			msg: &economicspb.MsgUpdateParams{
				Authority: suite.testAddrs[0].String(),
				Params:    *newParams,
			},
			shouldErr: true,
			errMsg:    "unauthorized",
		},
		{
			name: "invalid - empty authority",
			msg: &economicspb.MsgUpdateParams{
				Authority: "",
				Params:    *newParams,
			},
			shouldErr: true,
			errMsg:    "invalid authority",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			goCtx := sdk.WrapSDKContext(suite.ctx)
			resp, err := suite.msgServer.UpdateParams(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)

				// Verify params were updated
				params, err := suite.keeper.GetParams(suite.ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(newParams.Fees.BaseFee, params.Fees.BaseFee)
			}
		})
	}
}

func (suite *MsgServerTestSuite) TestMsgAdjustInflationRate() {
	tests := []struct {
		name      string
		msg       *economicspb.MsgAdjustInflationRate
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid inflation adjustment by authority",
			msg: &economicspb.MsgAdjustInflationRate{
				Authority: suite.authority,
				NewRate:   500, // 5% (basis points)
				Reason:    "Adjusting for economic conditions",
			},
			shouldErr: false,
		},
		{
			name: "invalid - unauthorized caller",
			msg: &economicspb.MsgAdjustInflationRate{
				Authority: suite.testAddrs[0].String(),
				NewRate:   500,
				Reason:    "Test",
			},
			shouldErr: true,
			errMsg:    "unauthorized",
		},
		{
			name: "invalid - rate too high",
			msg: &economicspb.MsgAdjustInflationRate{
				Authority: suite.authority,
				NewRate:   20000, // 200% - unrealistic
				Reason:    "Test",
			},
			shouldErr: true,
			errMsg:    "invalid inflation rate",
		},
		{
			name: "invalid - empty reason",
			msg: &economicspb.MsgAdjustInflationRate{
				Authority: suite.authority,
				NewRate:   500,
				Reason:    "",
			},
			shouldErr: true,
			errMsg:    "invalid reason",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			goCtx := sdk.WrapSDKContext(suite.ctx)
			resp, err := suite.msgServer.AdjustInflationRate(goCtx, tt.msg)

			if tt.shouldErr {
				suite.Require().Error(err)
				suite.Require().Nil(resp)
				if tt.errMsg != "" {
					suite.Require().Contains(err.Error(), tt.errMsg)
				}
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(resp)
				suite.Require().Equal(tt.msg.NewRate, resp.NewRate)
			}
		})
	}
}

// ============================
// AUTHORIZATION TESTS
// ============================

func (suite *MsgServerTestSuite) TestAuthorizationChecks() {
	suite.Run("vesting schedule creation requires valid creator", func() {
		msg := &economicspb.MsgCreateVestingSchedule{
			Creator:            "invalid_address",
			BeneficiaryAddress: suite.testAddrs[1].String(),
			TotalAmount: sdk.Coin{
				Denom:  "uaura",
				Amount: sdkmath.NewInt(1000000),
			},
			StartTime:       time.Now(),
			VestingDuration: 365 * 24 * 3600,
			VestingType:     economicspb.VestingType_VESTING_TYPE_LINEAR,
		}

		goCtx := sdk.WrapSDKContext(suite.ctx)
		_, err := suite.msgServer.CreateVestingSchedule(goCtx, msg)
		suite.Require().Error(err)
		suite.Require().Contains(err.Error(), "invalid")
	})

	suite.Run("admin operations require authority", func() {
		msg := &economicspb.MsgUpdateParams{
			Authority: suite.testAddrs[0].String(), // Not the authority
			Params:    *types.DefaultParams(),
		}

		goCtx := sdk.WrapSDKContext(suite.ctx)
		_, err := suite.msgServer.UpdateParams(goCtx, msg)
		suite.Require().Error(err)
		suite.Require().Contains(err.Error(), "unauthorized")
	})
}

// ============================
// EDGE CASE TESTS
// ============================

func (suite *MsgServerTestSuite) TestEdgeCases() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	suite.Run("nil message handling", func() {
		// This tests that the server handles nil messages gracefully
		// The actual implementation should validate this
		var nilMsg *economicspb.MsgCreateVestingSchedule
		_, err := suite.msgServer.CreateVestingSchedule(goCtx, nilMsg)
		suite.Require().Error(err)
	})

	suite.Run("duplicate operation prevention", func() {
		// Test that creating duplicate vesting schedules is handled
		msg := &economicspb.MsgCreateVestingSchedule{
			Creator:            suite.testAddrs[0].String(),
			BeneficiaryAddress: suite.testAddrs[1].String(),
			TotalAmount: sdk.Coin{
				Denom:  "uaura",
				Amount: sdkmath.NewInt(1000000),
			},
			StartTime:       time.Now(),
			VestingDuration: 365 * 24 * 3600,
			VestingType:     economicspb.VestingType_VESTING_TYPE_LINEAR,
		}

		// First creation should succeed
		resp1, err := suite.msgServer.CreateVestingSchedule(goCtx, msg)
		suite.Require().NoError(err)
		suite.Require().NotNil(resp1)

		// Second creation should also succeed (different schedule ID)
		resp2, err := suite.msgServer.CreateVestingSchedule(goCtx, msg)
		suite.Require().NoError(err)
		suite.Require().NotNil(resp2)
		suite.Require().NotEqual(resp1.ScheduleId, resp2.ScheduleId)
	})

	suite.Run("concurrent vote delegation", func() {
		// First lock tokens for the delegator to give them voting power
		lockMsg := &economicspb.MsgLockVotingTokens{
			Owner: suite.testAddrs[0].String(),
			Amount: sdk.Coin{
				Denom:  "uaura",
				Amount: sdkmath.NewInt(1000000),
			},
			LockDuration: 365 * 24 * 3600, // 1 year
		}
		_, err := suite.msgServer.LockVotingTokens(goCtx, lockMsg)
		suite.Require().NoError(err)

		// Test delegating to multiple delegates
		msg1 := &economicspb.MsgDelegateVote{
			Delegator:  suite.testAddrs[0].String(),
			Delegate:   suite.testAddrs[1].String(),
			Categories: []economicspb.ProposalCategory{economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT},
		}

		msg2 := &economicspb.MsgDelegateVote{
			Delegator:  suite.testAddrs[0].String(),
			Delegate:   suite.testAddrs[2].String(),
			Categories: []economicspb.ProposalCategory{economicspb.ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE},
		}

		_, err = suite.msgServer.DelegateVote(goCtx, msg1)
		suite.Require().NoError(err)

		_, err = suite.msgServer.DelegateVote(goCtx, msg2)
		suite.Require().NoError(err)
	})
}

// ============================
// EVENT EMISSION TESTS
// ============================

func (suite *MsgServerTestSuite) TestEventEmission() {
	goCtx := sdk.WrapSDKContext(suite.ctx)

	suite.Run("vesting schedule creation emits events", func() {
		msg := &economicspb.MsgCreateVestingSchedule{
			Creator:            suite.testAddrs[0].String(),
			BeneficiaryAddress: suite.testAddrs[1].String(),
			TotalAmount: sdk.Coin{
				Denom:  "uaura",
				Amount: sdkmath.NewInt(1000000),
			},
			StartTime:       time.Now(),
			VestingDuration: 365 * 24 * 3600,
			VestingType:     economicspb.VestingType_VESTING_TYPE_LINEAR,
		}

		eventsBefore := suite.ctx.EventManager().Events()
		_, err := suite.msgServer.CreateVestingSchedule(goCtx, msg)
		suite.Require().NoError(err)
		eventsAfter := suite.ctx.EventManager().Events()

		// Should have emitted at least one event
		suite.Require().Greater(len(eventsAfter), len(eventsBefore))
	})

	suite.Run("proposal submission emits events", func() {
		msg := &economicspb.MsgSubmitProposal{
			Title:          "Event Test Proposal",
			Description:    "Testing event emission",
			Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
			Proposer:       suite.testAddrs[0].String(),
			InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(1000000)}},
		}

		eventsBefore := suite.ctx.EventManager().Events()
		_, err := suite.msgServer.SubmitProposal(goCtx, msg)
		suite.Require().NoError(err)
		eventsAfter := suite.ctx.EventManager().Events()

		// Should have emitted at least one event
		suite.Require().Greater(len(eventsAfter), len(eventsBefore))
	})
}

// ============================
// HELPER FUNCTIONS (for future use if needed)
// ============================

// Helper function to advance time in context (if needed in future tests)
func (suite *MsgServerTestSuite) advanceTime(duration time.Duration) {
	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(duration))
}

// Helper function to create a test proposal that reaches voting period
func (suite *MsgServerTestSuite) createVotingProposal() uint64 {
	msg := &economicspb.MsgSubmitProposal{
		Title:          "Test Voting Proposal",
		Description:    "Test proposal for voting tests",
		Category:       economicspb.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		Proposer:       suite.testAddrs[0].String(),
		InitialDeposit: sdk.Coins{sdk.Coin{Denom: "uaura", Amount: sdkmath.NewInt(10000000)}},
	}

	goCtx := sdk.WrapSDKContext(suite.ctx)
	resp, err := suite.msgServer.SubmitProposal(goCtx, msg)
	suite.Require().NoError(err)
	return resp.ProposalId
}

// ============================
// STANDALONE TESTS (non-suite)
// ============================

func TestMsgServerNilKeeper(t *testing.T) {
	// Test that creating msg server with nil keeper is handled
	require.NotPanics(t, func() {
		_ = NewMsgServer(nil)
	})
}
