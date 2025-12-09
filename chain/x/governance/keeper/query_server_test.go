package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/aequitas/aura/chain/testutil/keeper"
	govkeeper "github.com/aequitas/aura/chain/x/governance/keeper"
	govpb "github.com/aequitas/aura/proto/aura/governance/v1beta1"
)

// TestQueryProposalsPagination tests that Proposals query supports Cosmos SDK pagination
func TestQueryProposalsPagination(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	keeper := suite.keeper
	queryServer := suite.queryServer

	// Create multiple proposals for pagination testing
	for i := 0; i < 5; i++ {
		proposal := &govpb.Proposal{
			Title:       "Test Proposal",
			Description: "Test Description",
			Status:      govpb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		}
		proposal.Id = keeper.GetNextProposalID(ctx)
		keeper.SetNextProposalID(ctx, proposal.Id+1)
		err := keeper.SetProposal(ctx, proposal)
		require.NoError(t, err)
	}

	// Test 1: Query without pagination (should get default limit)
	resp1, err := queryServer.Proposals(sdk.WrapSDKContext(ctx), &govpb.QueryProposalsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp1)
	require.Len(t, resp1.Proposals, 5)
	require.NotNil(t, resp1.Pagination)

	// Test 2: Query with limit=3
	resp2, err := queryServer.Proposals(sdk.WrapSDKContext(ctx), &govpb.QueryProposalsRequest{
		Pagination: &query.PageRequest{
			Limit: 3,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp2.Proposals, 3)
	require.NotNil(t, resp2.Pagination)
	require.NotEmpty(t, resp2.Pagination.NextKey)

	// Test 3: Query next page using NextKey
	resp3, err := queryServer.Proposals(sdk.WrapSDKContext(ctx), &govpb.QueryProposalsRequest{
		Pagination: &query.PageRequest{
			Key:   resp2.Pagination.NextKey,
			Limit: 3,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp3.Proposals, 2)
	require.NotNil(t, resp3.Pagination)

	// Test 4: Verify no duplicates across pages
	allProposalIDs := make(map[uint64]bool)
	for _, proposal := range resp2.Proposals {
		allProposalIDs[proposal.Id] = true
	}
	for _, proposal := range resp3.Proposals {
		require.False(t, allProposalIDs[proposal.Id], "duplicate proposal found across pages")
	}
}

// TestQueryProposalsPaginationWithFilters tests pagination with status filters
func TestQueryProposalsPaginationWithFilters(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	keeper := suite.keeper
	queryServer := suite.queryServer

	// Create proposals with different statuses
	for i := 0; i < 3; i++ {
		proposal := &govpb.Proposal{
			Title:       "Voting Proposal",
			Description: "Test Description",
			Status:      govpb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		}
		proposal.Id = keeper.GetNextProposalID(ctx)
		keeper.SetNextProposalID(ctx, proposal.Id+1)
		err := keeper.SetProposal(ctx, proposal)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		proposal := &govpb.Proposal{
			Title:       "Passed Proposal",
			Description: "Test Description",
			Status:      govpb.ProposalStatus_PROPOSAL_STATUS_PASSED,
		}
		proposal.Id = keeper.GetNextProposalID(ctx)
		keeper.SetNextProposalID(ctx, proposal.Id+1)
		err := keeper.SetProposal(ctx, proposal)
		require.NoError(t, err)
	}

	// Test 1: Query only voting proposals with pagination
	resp1, err := queryServer.Proposals(sdk.WrapSDKContext(ctx), &govpb.QueryProposalsRequest{
		Status: govpb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		Pagination: &query.PageRequest{
			Limit: 2,
		},
	})
	require.NoError(t, err)
	require.Len(t, resp1.Proposals, 2)
	require.NotNil(t, resp1.Pagination)

	// Verify all returned proposals have correct status
	for _, proposal := range resp1.Proposals {
		require.Equal(t, govpb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD, proposal.Status)
	}

	// Test 2: Query only passed proposals
	resp2, err := queryServer.Proposals(sdk.WrapSDKContext(ctx), &govpb.QueryProposalsRequest{
		Status: govpb.ProposalStatus_PROPOSAL_STATUS_PASSED,
	})
	require.NoError(t, err)
	require.Len(t, resp2.Proposals, 2)

	// Verify all returned proposals have correct status
	for _, proposal := range resp2.Proposals {
		require.Equal(t, govpb.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)
	}
}

// TestQueryProposalsPaginationNilRequest tests that nil request is handled gracefully
func TestQueryProposalsPaginationNilRequest(t *testing.T) {
	suite := setupTestSuite(t)
	ctx := suite.ctx
	queryServer := suite.queryServer

	// Test with nil request - should use default pagination
	resp, err := queryServer.Proposals(sdk.WrapSDKContext(ctx), nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Pagination)
}

// testSuite holds the test dependencies
type testSuite struct {
	ctx         sdk.Context
	keeper      *govkeeper.Keeper
	queryServer govpb.QueryServer
}

// Helper function to setup test suite
func setupTestSuite(t *testing.T) *testSuite {
	t.Helper()

	// Create keeper and context using testutil
	k, ctx := testkeeper.GovernanceKeeper(t)

	// Create query server
	queryServer := govkeeper.NewQueryServerImpl(k)

	return &testSuite{
		ctx:         ctx,
		keeper:      k,
		queryServer: queryServer,
	}
}
