package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/governance/types"
	govpb "github.com/aequitas/aura/proto/aura/governance/v1beta1"
)

var _ govpb.QueryServer = (*queryServer)(nil)

type queryServer struct {
	govpb.UnimplementedQueryServer
	Keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper *Keeper) govpb.QueryServer {
	return &queryServer{Keeper: keeper}
}

// Proposal queries a proposal based on proposal id
func (qs queryServer) Proposal(goCtx context.Context, req *govpb.QueryProposalRequest) (*govpb.QueryProposalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	proposal, err := qs.Keeper.GetProposal(ctx, req.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}

	return &govpb.QueryProposalResponse{Proposal: proposal}, nil
}

// Proposals queries all proposals
func (qs queryServer) Proposals(goCtx context.Context, req *govpb.QueryProposalsRequest) (*govpb.QueryProposalsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	proposals := qs.Keeper.GetAllProposals(ctx)

	// Filter by status if provided
	if req.Status != govpb.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED {
		var filtered []*types.Proposal
		for _, p := range proposals {
			if p.Status == req.Status {
				filtered = append(filtered, p)
			}
		}
		proposals = filtered
	}

	// Filter by voter if provided
	if req.Voter != "" {
		var filtered []*types.Proposal
		for _, p := range proposals {
			_, err := qs.Keeper.GetVote(ctx, p.Id, req.Voter)
			if err == nil {
				filtered = append(filtered, p)
			}
		}
		proposals = filtered
	}

	// Filter by depositor if provided
	if req.Depositor != "" {
		var filtered []*types.Proposal
		for _, p := range proposals {
			_, err := qs.Keeper.GetDeposit(ctx, p.Id, req.Depositor)
			if err == nil {
				filtered = append(filtered, p)
			}
		}
		proposals = filtered
	}

	return &govpb.QueryProposalsResponse{Proposals: proposals}, nil
}

// Vote queries a vote by proposal id and voter
func (qs queryServer) Vote(goCtx context.Context, req *govpb.QueryVoteRequest) (*govpb.QueryVoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id cannot be 0")
	}

	if req.Voter == "" {
		return nil, status.Error(codes.InvalidArgument, "voter cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	vote, err := qs.Keeper.GetVote(ctx, req.ProposalId, req.Voter)
	if err != nil {
		return nil, status.Error(codes.NotFound, "vote not found")
	}

	return &govpb.QueryVoteResponse{Vote: vote}, nil
}

// Votes queries all votes for a proposal
func (qs queryServer) Votes(goCtx context.Context, req *govpb.QueryVotesRequest) (*govpb.QueryVotesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	votes := qs.Keeper.GetVotes(ctx, req.ProposalId)

	return &govpb.QueryVotesResponse{Votes: votes}, nil
}

// Params queries the governance parameters
func (qs queryServer) Params(goCtx context.Context, req *govpb.QueryParamsRequest) (*govpb.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := qs.Keeper.GetParams(ctx)

	return &govpb.QueryParamsResponse{Params: params}, nil
}

// Deposit queries a deposit by proposal id and depositor
func (qs queryServer) Deposit(goCtx context.Context, req *govpb.QueryDepositRequest) (*govpb.QueryDepositResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id cannot be 0")
	}

	if req.Depositor == "" {
		return nil, status.Error(codes.InvalidArgument, "depositor cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	deposit, err := qs.Keeper.GetDeposit(ctx, req.ProposalId, req.Depositor)
	if err != nil {
		return nil, status.Error(codes.NotFound, "deposit not found")
	}

	return &govpb.QueryDepositResponse{Deposit: deposit}, nil
}

// Deposits queries all deposits for a proposal
func (qs queryServer) Deposits(goCtx context.Context, req *govpb.QueryDepositsRequest) (*govpb.QueryDepositsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	deposits := qs.Keeper.GetDeposits(ctx, req.ProposalId)

	return &govpb.QueryDepositsResponse{Deposits: deposits}, nil
}

// TallyResult queries the tally of a proposal
func (qs queryServer) TallyResult(goCtx context.Context, req *govpb.QueryTallyResultRequest) (*govpb.QueryTallyResultResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	proposal, err := qs.Keeper.GetProposal(ctx, req.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}

	// Calculate tally
	tally := qs.Keeper.CalculateTally(ctx, req.ProposalId)

	// Update proposal with tally if it doesn't have one
	if proposal.FinalTallyResult == nil {
		proposal.FinalTallyResult = tally
	}

	return &govpb.QueryTallyResultResponse{Tally: tally}, nil
}

// VoteDelegations queries all vote delegations for a delegator
func (qs queryServer) VoteDelegations(goCtx context.Context, req *govpb.QueryVoteDelegationsRequest) (*govpb.QueryVoteDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Delegator == "" {
		return nil, status.Error(codes.InvalidArgument, "delegator cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	delegations := qs.Keeper.GetVoteDelegations(ctx, req.Delegator)

	return &govpb.QueryVoteDelegationsResponse{Delegations: delegations}, nil
}

// VotingPower queries the voting power of an address
func (qs queryServer) VotingPower(goCtx context.Context, req *govpb.QueryVotingPowerRequest) (*govpb.QueryVotingPowerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get total voting power (includes delegations)
	totalPower, err := qs.Keeper.GetVotingPower(ctx, req.Address)
	if err != nil {
		return nil, err
	}

	// Get delegated power separately for breakdown
	delegatedPowerInt := qs.Keeper.GetDelegatedVotingPower(ctx, req.Address)

	// Get base power (direct stake)
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}
	basePower, err := qs.Keeper.stakingKeeper.GetDelegatorBonded(ctx, addr)
	if err != nil {
		basePower = sdkmath.ZeroInt()
	}

	return &govpb.QueryVotingPowerResponse{
		VotingPower:    basePower.String(),
		DelegatedPower: delegatedPowerInt.String(),
		TotalPower:     totalPower.String(),
	}, nil
}

// TokenLocks queries all token locks for an address
func (qs queryServer) TokenLocks(goCtx context.Context, req *govpb.QueryTokenLocksRequest) (*govpb.QueryTokenLocksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	locks := qs.Keeper.GetTokenLocks(ctx, req.Address)

	return &govpb.QueryTokenLocksResponse{Locks: locks}, nil
}

// VetoRequests queries all veto requests for a proposal
func (qs queryServer) VetoRequests(goCtx context.Context, req *govpb.QueryVetoRequestsRequest) (*govpb.QueryVetoRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	vetoRequests := qs.Keeper.GetVetoRequests(ctx, req.ProposalId)

	return &govpb.QueryVetoRequestsResponse{VetoRequests: vetoRequests}, nil
}

// SnapshotVotes queries snapshot votes for a proposal
func (qs queryServer) SnapshotVotes(goCtx context.Context, req *govpb.QuerySnapshotVotesRequest) (*govpb.QuerySnapshotVotesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ProposalId == 0 {
		return nil, status.Error(codes.InvalidArgument, "proposal id cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	votes := qs.Keeper.GetSnapshotVotes(ctx, req.ProposalId)

	return &govpb.QuerySnapshotVotesResponse{Votes: votes}, nil
}
