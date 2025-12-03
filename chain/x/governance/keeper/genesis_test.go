package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

func setupKeeper(t *testing.T) (*Keeper, sdk.Context) {
	input := keepertest.CreateTestInputWithKeys(t, "governance")
	mockStaking := &MockStakingKeeper{delegatorBonded: make(map[string]sdkmath.Int)}
	mockBank := &MockBankKeeper{
		balances:       make(map[string]sdk.Coins),
		moduleBalances: make(map[string]sdk.Coins),
	}
	keeper := NewKeeper(input.Cdc, input.StoreKey, mockStaking, mockBank)
	ctx := input.Ctx.WithKVGasConfig(storetypes.GasConfig{})
	return keeper, ctx
}


func TestInitGenesis(t *testing.T) {
	tests := []struct {
		name    string
		genesis types.GenesisState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid genesis with all data",
			genesis: types.GenesisState{
				Params: &types.GovernanceParams{
					MinDeposit:        "1000",
					MaxDepositPeriod:  durationpb.New(172800000000000),  // 2 days in nanoseconds
					VotingPeriod:      durationpb.New(604800000000000),  // 7 days in nanoseconds
					Quorum:            "0.334",
					Threshold:         "0.5",
					VetoThreshold:     "0.334",
				},
			},
			wantErr: false,
		},
		{
			name: "default genesis",
			genesis: types.GenesisState{
				Params: nil,
			},
			wantErr: false,
		},
		{
			name: "invalid genesis - nil params in validation",
			genesis: types.GenesisState{
				Params: nil,
			},
			wantErr: false, // InitGenesis uses defaults for nil params
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupKeeper(t)

			// Initialize genesis
			err := keeper.InitGenesis(ctx, tt.genesis)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)

			// Verify params were set
			p := keeper.GetParams(ctx)
			require.NotNil(t, p)
			require.NotEmpty(t, p.MinDeposit)
		})
	}
}

func TestExportGenesis(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Set custom params
	params := &types.GovernanceParams{
		MinDeposit:        "2000",
		MaxDepositPeriod:  durationpb.New(259200000000000),  // 3 days
		VotingPeriod:      durationpb.New(1209600000000000), // 14 days
		Quorum:            "0.4",
		Threshold:         "0.6",
		VetoThreshold:     "0.33",
	}
	keeper.SetParams(ctx, params)

	// Export genesis
	exported := keeper.ExportGenesis(ctx)

	// Verify exported data
	require.NotNil(t, exported.Params)
	require.Equal(t, "2000", exported.Params.MinDeposit)
	require.Equal(t, "0.4", exported.Params.Quorum)
	require.Equal(t, "0.6", exported.Params.Threshold)
	require.Equal(t, "0.33", exported.Params.VetoThreshold)
}

func TestGenesisRoundTrip(t *testing.T) {
	// Create first keeper with initial state
	keeper1, ctx1 := setupKeeper(t)

	// Set custom params
	params := &types.GovernanceParams{
		MinDeposit:        "5000",
		MaxDepositPeriod:  durationpb.New(345600000000000),  // 4 days
		VotingPeriod:      durationpb.New(1814400000000000), // 21 days
		Quorum:            "0.35",
		Threshold:         "0.55",
		VetoThreshold:     "0.35",
	}
	keeper1.SetParams(ctx1, params)

	// Export genesis from keeper1
	exported := keeper1.ExportGenesis(ctx1)

	// Create a new keeper and import the exported genesis
	keeper2, ctx2 := setupKeeper(t)
	err := keeper2.InitGenesis(ctx2, exported)
	require.NoError(t, err)

	// Verify all data was preserved
	params1 := keeper1.GetParams(ctx1)
	params2 := keeper2.GetParams(ctx2)
	require.Equal(t, params1.MinDeposit, params2.MinDeposit)
	require.Equal(t, params1.Quorum, params2.Quorum)
	require.Equal(t, params1.Threshold, params2.Threshold)
	require.Equal(t, params1.VetoThreshold, params2.VetoThreshold)

	// Export again and verify consistency
	exported2 := keeper2.ExportGenesis(ctx2)
	require.Equal(t, exported.Params.MinDeposit, exported2.Params.MinDeposit)
	require.Equal(t, exported.Params.Quorum, exported2.Params.Quorum)
	require.Equal(t, exported.Params.Threshold, exported2.Params.Threshold)
}

func TestDefaultGenesis(t *testing.T) {
	// Test that default genesis is valid
	defaultGen := types.DefaultGenesis()
	require.NotNil(t, defaultGen)

	// Verify default params
	require.NotNil(t, defaultGen.Params)
	require.NotEmpty(t, defaultGen.Params.MinDeposit)
	require.NotEmpty(t, defaultGen.Params.Quorum)
	require.NotEmpty(t, defaultGen.Params.Threshold)
	require.NotEmpty(t, defaultGen.Params.VetoThreshold)
	require.NotNil(t, defaultGen.Params.MaxDepositPeriod)
	require.NotNil(t, defaultGen.Params.VotingPeriod)

	// Test importing default genesis
	keeper, ctx := setupKeeper(t)
	err := keeper.InitGenesis(ctx, *defaultGen)
	require.NoError(t, err)

	// Verify keeper state after importing default genesis
	p := keeper.GetParams(ctx)
	require.NotNil(t, p)
	require.NotEmpty(t, p.MinDeposit)
	require.NotEmpty(t, p.Quorum)
}

func TestInitGenesis_WithCustomParams(t *testing.T) {
	genesis := types.GenesisState{
		Params: &types.GovernanceParams{
			MinDeposit:        "10000",
			MaxDepositPeriod:  durationpb.New(432000000000000),  // 5 days
			VotingPeriod:      durationpb.New(2419200000000000), // 28 days
			Quorum:            "0.5",
			Threshold:         "0.667",
			VetoThreshold:     "0.4",
		},
	}

	keeper, ctx := setupKeeper(t)
	err := keeper.InitGenesis(ctx, genesis)
	require.NoError(t, err)

	// Verify custom params were set
	params := keeper.GetParams(ctx)
	require.NotNil(t, params)
	require.Equal(t, "10000", params.MinDeposit)
	require.Equal(t, "0.5", params.Quorum)
	require.Equal(t, "0.667", params.Threshold)
	require.Equal(t, "0.4", params.VetoThreshold)
}

func TestInitGenesis_NilParams(t *testing.T) {
	genesis := types.GenesisState{
		Params: nil,
	}

	keeper, ctx := setupKeeper(t)
	err := keeper.InitGenesis(ctx, genesis)
	require.NoError(t, err)

	// Verify default params were set
	params := keeper.GetParams(ctx)
	require.NotNil(t, params)
	require.NotEmpty(t, params.MinDeposit)
	require.NotEmpty(t, params.Quorum)
}

func TestExportGenesis_DefaultState(t *testing.T) {
	keeper, ctx := setupKeeper(t)

	// Don't set any params, should use defaults
	exported := keeper.ExportGenesis(ctx)

	// Verify default params are exported
	require.NotNil(t, exported.Params)
	require.NotEmpty(t, exported.Params.MinDeposit)
	require.NotEmpty(t, exported.Params.Quorum)
	require.NotEmpty(t, exported.Params.Threshold)
	require.NotEmpty(t, exported.Params.VetoThreshold)
}

func TestGenesisRoundTrip_MultipleIterations(t *testing.T) {
	// Test that multiple round trips preserve data integrity
	keeper1, ctx1 := setupKeeper(t)

	initialParams := &types.GovernanceParams{
		MinDeposit:        "3000",
		MaxDepositPeriod:  durationpb.New(259200000000000),  // 3 days
		VotingPeriod:      durationpb.New(1209600000000000), // 14 days
		Quorum:            "0.45",
		Threshold:         "0.65",
		VetoThreshold:     "0.38",
	}
	keeper1.SetParams(ctx1, initialParams)

	// First export
	exported1 := keeper1.ExportGenesis(ctx1)

	// Import to keeper2
	keeper2, ctx2 := setupKeeper(t)
	err := keeper2.InitGenesis(ctx2, exported1)
	require.NoError(t, err)

	// Second export
	exported2 := keeper2.ExportGenesis(ctx2)

	// Import to keeper3
	keeper3, ctx3 := setupKeeper(t)
	err = keeper3.InitGenesis(ctx3, exported2)
	require.NoError(t, err)

	// Third export
	exported3 := keeper3.ExportGenesis(ctx3)

	// Verify all exports are consistent
	require.Equal(t, exported1.Params.MinDeposit, exported2.Params.MinDeposit)
	require.Equal(t, exported2.Params.MinDeposit, exported3.Params.MinDeposit)
	require.Equal(t, exported1.Params.Quorum, exported2.Params.Quorum)
	require.Equal(t, exported2.Params.Quorum, exported3.Params.Quorum)
	require.Equal(t, exported1.Params.Threshold, exported2.Params.Threshold)
	require.Equal(t, exported2.Params.Threshold, exported3.Params.Threshold)
}

func TestGenesisRoundTrip_CompleteState(t *testing.T) {
	// Test that all governance data is preserved through export/import
	keeper1, ctx1 := setupKeeper(t)

	// Set custom params
	params := &types.GovernanceParams{
		MinDeposit:        "5000auracoin",
		MaxDepositPeriod:  durationpb.New(172800000000000),  // 2 days
		VotingPeriod:      durationpb.New(604800000000000),  // 7 days
		Quorum:            "0.334",
		Threshold:         "0.5",
		VetoThreshold:     "0.334",
	}
	keeper1.SetParams(ctx1, params)

	// Create test addresses
	addr1 := sdk.AccAddress("addr1_______________")
	addr2 := sdk.AccAddress("addr2_______________")
	addr3 := sdk.AccAddress("addr3_______________")

	// Set starting proposal ID
	keeper1.SetNextProposalID(ctx1, 10)

	// Create and store proposals
	proposal1 := &types.Proposal{
		Id:          1,
		Title:       "Test Proposal 1",
		Description: "Description for test proposal 1",
		Category:    types.ProposalCategory_PROPOSAL_CATEGORY_TEXT,
		Status:      types.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Proposer:    addr1.String(),
		TotalDeposit: "1000auracoin",
	}
	require.NoError(t, keeper1.SetProposal(ctx1, proposal1))

	proposal2 := &types.Proposal{
		Id:          2,
		Title:       "Test Proposal 2",
		Description: "Description for test proposal 2",
		Category:    types.ProposalCategory_PROPOSAL_CATEGORY_PARAMETER_CHANGE,
		Status:      types.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
		Proposer:    addr2.String(),
		TotalDeposit: "500auracoin",
	}
	require.NoError(t, keeper1.SetProposal(ctx1, proposal2))

	// Create and store deposits
	deposit1 := &types.Deposit{
		ProposalId: 1,
		Depositor:  addr1.String(),
		Amount:     "1000auracoin",
	}
	require.NoError(t, keeper1.SetDeposit(ctx1, deposit1))

	deposit2 := &types.Deposit{
		ProposalId: 2,
		Depositor:  addr2.String(),
		Amount:     "500auracoin",
	}
	require.NoError(t, keeper1.SetDeposit(ctx1, deposit2))

	// Create and store votes
	vote1 := &types.Vote{
		ProposalId: 1,
		Voter:      addr1.String(),
		Option:     types.VoteOption_VOTE_OPTION_YES,
	}
	require.NoError(t, keeper1.SetVote(ctx1, vote1))

	vote2 := &types.Vote{
		ProposalId: 1,
		Voter:      addr2.String(),
		Option:     types.VoteOption_VOTE_OPTION_NO,
	}
	require.NoError(t, keeper1.SetVote(ctx1, vote2))

	// Create and store vote delegations
	delegation1 := &types.VoteDelegation{
		Delegator: addr1.String(),
		Delegate:  addr3.String(),
	}
	require.NoError(t, keeper1.SetVoteDelegation(ctx1, delegation1))

	// Create and store token locks
	tokenLock1 := &types.TokenLock{
		Owner:        addr1.String(),
		ProposalId:   1,
		LockedAmount: "100auracoin",
	}
	require.NoError(t, keeper1.SetTokenLock(ctx1, tokenLock1))

	// Create and store veto requests
	vetoRequest1 := &types.VetoRequest{
		ProposalId: 1,
		Vetoer:     addr3.String(),
		Reason:     "Test veto reason",
	}
	require.NoError(t, keeper1.SetVetoRequest(ctx1, vetoRequest1))

	// Export genesis from keeper1
	exported := keeper1.ExportGenesis(ctx1)

	// Verify exported data completeness
	require.NotNil(t, exported.Params)
	require.Len(t, exported.Proposals, 2, "Should export 2 proposals")
	require.Len(t, exported.Deposits, 2, "Should export 2 deposits")
	require.Len(t, exported.Votes, 2, "Should export 2 votes")
	require.Len(t, exported.VoteDelegations, 1, "Should export 1 delegation")
	require.Len(t, exported.TokenLocks, 1, "Should export 1 token lock")
	require.Len(t, exported.VetoRequests, 1, "Should export 1 veto request")
	require.Equal(t, uint64(10), exported.StartingProposalId, "Should export starting proposal ID")

	// Create a new keeper and import the exported genesis
	keeper2, ctx2 := setupKeeper(t)
	err := keeper2.InitGenesis(ctx2, exported)
	require.NoError(t, err)

	// Verify all proposals were imported
	importedProposal1, err := keeper2.GetProposal(ctx2, 1)
	require.NoError(t, err)
	require.Equal(t, proposal1.Id, importedProposal1.Id)
	require.Equal(t, proposal1.Title, importedProposal1.Title)
	require.Equal(t, proposal1.Proposer, importedProposal1.Proposer)

	importedProposal2, err := keeper2.GetProposal(ctx2, 2)
	require.NoError(t, err)
	require.Equal(t, proposal2.Id, importedProposal2.Id)
	require.Equal(t, proposal2.Title, importedProposal2.Title)

	// Verify all deposits were imported
	importedDeposit1, err := keeper2.GetDeposit(ctx2, 1, addr1.String())
	require.NoError(t, err)
	require.Equal(t, deposit1.ProposalId, importedDeposit1.ProposalId)
	require.Equal(t, deposit1.Amount, importedDeposit1.Amount)

	importedDeposit2, err := keeper2.GetDeposit(ctx2, 2, addr2.String())
	require.NoError(t, err)
	require.Equal(t, deposit2.Amount, importedDeposit2.Amount)

	// Verify all votes were imported
	importedVote1, err := keeper2.GetVote(ctx2, 1, addr1.String())
	require.NoError(t, err)
	require.Equal(t, vote1.Option, importedVote1.Option)

	importedVote2, err := keeper2.GetVote(ctx2, 1, addr2.String())
	require.NoError(t, err)
	require.Equal(t, vote2.Option, importedVote2.Option)

	// Verify vote delegations were imported
	importedDelegations := keeper2.GetVoteDelegations(ctx2, addr1.String())
	require.Len(t, importedDelegations, 1)
	require.Equal(t, delegation1.Delegate, importedDelegations[0].Delegate)

	// Verify token locks were imported
	importedLocks := keeper2.GetTokenLocks(ctx2, addr1.String())
	require.Len(t, importedLocks, 1)
	require.Equal(t, tokenLock1.LockedAmount, importedLocks[0].LockedAmount)

	// Verify veto requests were imported
	importedVeto, err := keeper2.GetVetoRequest(ctx2, 1)
	require.NoError(t, err)
	require.Equal(t, vetoRequest1.Vetoer, importedVeto.Vetoer)
	require.Equal(t, vetoRequest1.Reason, importedVeto.Reason)

	// Verify starting proposal ID was imported
	nextID := keeper2.GetNextProposalID(ctx2)
	require.Equal(t, uint64(10), nextID, "Starting proposal ID should be preserved")

	// Export again and verify consistency (second round trip)
	exported2 := keeper2.ExportGenesis(ctx2)
	require.Len(t, exported2.Proposals, 2, "Second export should have 2 proposals")
	require.Len(t, exported2.Deposits, 2, "Second export should have 2 deposits")
	require.Len(t, exported2.Votes, 2, "Second export should have 2 votes")
	require.Len(t, exported2.VoteDelegations, 1, "Second export should have 1 delegation")
	require.Len(t, exported2.TokenLocks, 1, "Second export should have 1 token lock")
	require.Len(t, exported2.VetoRequests, 1, "Second export should have 1 veto request")
	require.Equal(t, exported.StartingProposalId, exported2.StartingProposalId, "Starting proposal ID should match")

	// Verify proposal data matches
	require.Equal(t, exported.Proposals[0].Id, exported2.Proposals[0].Id)
	require.Equal(t, exported.Proposals[0].Title, exported2.Proposals[0].Title)
	require.Equal(t, exported.Proposals[1].Id, exported2.Proposals[1].Id)
	require.Equal(t, exported.Proposals[1].Title, exported2.Proposals[1].Title)
}

func TestInitGenesis_ErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		genesis types.GenesisState
		wantErr bool
		errMsg  string
	}{
		{
			name: "nil proposal in list",
			genesis: types.GenesisState{
				Params:    types.DefaultParams(),
				Proposals: []*types.Proposal{nil},
			},
			wantErr: false, // Should skip nil proposals with warning
		},
		{
			name: "nil deposit in list",
			genesis: types.GenesisState{
				Params:   types.DefaultParams(),
				Deposits: []*types.Deposit{nil},
			},
			wantErr: false, // Should skip nil deposits with warning
		},
		{
			name: "nil vote in list",
			genesis: types.GenesisState{
				Params: types.DefaultParams(),
				Votes:  []*types.Vote{nil},
			},
			wantErr: false, // Should skip nil votes with warning
		},
		{
			name: "nil delegation in list",
			genesis: types.GenesisState{
				Params:          types.DefaultParams(),
				VoteDelegations: []*types.VoteDelegation{nil},
			},
			wantErr: false, // Should skip nil delegations with warning
		},
		{
			name: "nil token lock in list",
			genesis: types.GenesisState{
				Params:     types.DefaultParams(),
				TokenLocks: []*types.TokenLock{nil},
			},
			wantErr: false, // Should skip nil token locks with warning
		},
		{
			name: "nil veto request in list",
			genesis: types.GenesisState{
				Params:       types.DefaultParams(),
				VetoRequests: []*types.VetoRequest{nil},
			},
			wantErr: false, // Should skip nil veto requests with warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper, ctx := setupKeeper(t)
			err := keeper.InitGenesis(ctx, tt.genesis)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
