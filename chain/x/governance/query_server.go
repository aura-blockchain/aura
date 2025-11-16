package governance

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/governance/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

// QueryServer implements the governance Query service
type QueryServer struct {
	types.UnimplementedQueryServer
	keeper *keeper.Keeper
}

// NewQueryServer creates a new QueryServer
func NewQueryServer(k *keeper.Keeper) *QueryServer {
	return &QueryServer{keeper: k}
}

// Proposal queries a proposal by ID
func (q *QueryServer) Proposal(ctx context.Context, req *QueryProposalRequest) (*QueryProposalResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	proposal, err := q.keeper.GetProposal(req.ProposalID)
	if err != nil {
		return nil, err
	}

	return &QueryProposalResponse{
		Proposal: convertProposalToProto(proposal),
	}, nil
}

// Proposals queries all proposals
func (q *QueryServer) Proposals(ctx context.Context, req *QueryProposalsRequest) (*QueryProposalsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	proposals := q.keeper.GetProposals()
	protoProposals := make([]*ProposalProto, 0, len(proposals))

	for _, proposal := range proposals {
		// Filter by status if specified
		if req.Status != 0 && types.ProposalStatus(req.Status) != proposal.Status {
			continue
		}

		protoProposals = append(protoProposals, convertProposalToProto(proposal))
	}

	return &QueryProposalsResponse{Proposals: protoProposals}, nil
}

// Vote queries a vote
func (q *QueryServer) Vote(ctx context.Context, req *QueryVoteRequest) (*QueryVoteResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	if req.Voter == "" {
		return nil, fmt.Errorf("voter cannot be empty")
	}

	vote, err := q.keeper.GetVote(req.ProposalID, req.Voter)
	if err != nil {
		return nil, err
	}

	return &QueryVoteResponse{
		Vote: convertVoteToProto(vote),
	}, nil
}

// Votes queries all votes for a proposal
func (q *QueryServer) Votes(ctx context.Context, req *QueryVotesRequest) (*QueryVotesResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	votes := q.keeper.GetVotes(req.ProposalID)
	protoVotes := make([]*VoteProto, 0, len(votes))

	for _, vote := range votes {
		protoVotes = append(protoVotes, convertVoteToProto(vote))
	}

	return &QueryVotesResponse{Votes: protoVotes}, nil
}

// Params queries governance parameters
func (q *QueryServer) Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	params := q.keeper.GetParams()

	return &QueryParamsResponse{
		Params: &ParamsProto{
			MinDeposit:         params.MinDeposit,
			Quorum:             params.Quorum,
			Threshold:          params.Threshold,
			VetoThreshold:      params.VetoThreshold,
			EmergencyQuorum:    params.EmergencyQuorum,
			EmergencyThreshold: params.EmergencyThreshold,
		},
	}, nil
}

// TallyResult queries the tally of a proposal
func (q *QueryServer) TallyResult(ctx context.Context, req *QueryTallyResultRequest) (*QueryTallyResultResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	tally, err := q.keeper.TallyVotes(req.ProposalID)
	if err != nil {
		return nil, err
	}

	return &QueryTallyResultResponse{
		Tally: &TallyResultProto{
			Yes:            tally.Yes,
			No:             tally.No,
			Abstain:        tally.Abstain,
			NoWithVeto:     tally.NoWithVeto,
			TurnoutPercent: tally.TurnoutPercent,
		},
	}, nil
}

// Helper conversion functions
func convertProposalToProto(p *types.Proposal) *ProposalProto {
	return &ProposalProto{
		ID:             p.ID,
		Title:          p.Title,
		Description:    p.Description,
		Category:       int32(p.Category),
		Status:         int32(p.Status),
		Proposer:       p.Proposer,
		TotalDeposit:   p.TotalDeposit,
		IsEmergency:    p.IsEmergency,
		SnapshotHeight: p.SnapshotHeight,
	}
}

func convertVoteToProto(v *types.Vote) *VoteProto {
	return &VoteProto{
		ProposalID:  v.ProposalID,
		Voter:       v.Voter,
		Option:      int32(v.Option),
		IsSecret:    v.IsSecret,
		VotingPower: v.VotingPower,
	}
}

// Query message types (would normally be generated from proto)
type QueryProposalRequest struct {
	ProposalID uint64
}

type QueryProposalResponse struct {
	Proposal *ProposalProto
}

type QueryProposalsRequest struct {
	Status    int32
	Voter     string
	Depositor string
}

type QueryProposalsResponse struct {
	Proposals []*ProposalProto
}

type QueryVoteRequest struct {
	ProposalID uint64
	Voter      string
}

type QueryVoteResponse struct {
	Vote *VoteProto
}

type QueryVotesRequest struct {
	ProposalID uint64
}

type QueryVotesResponse struct {
	Votes []*VoteProto
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params *ParamsProto
}

type QueryTallyResultRequest struct {
	ProposalID uint64
}

type QueryTallyResultResponse struct {
	Tally *TallyResultProto
}

// Proto types
type ProposalProto struct {
	ID             uint64
	Title          string
	Description    string
	Category       int32
	Status         int32
	Proposer       string
	TotalDeposit   string
	IsEmergency    bool
	SnapshotHeight uint64
}

type VoteProto struct {
	ProposalID  uint64
	Voter       string
	Option      int32
	IsSecret    bool
	VotingPower string
}

type ParamsProto struct {
	MinDeposit         string
	Quorum             string
	Threshold          string
	VetoThreshold      string
	EmergencyQuorum    string
	EmergencyThreshold string
}

type TallyResultProto struct {
	Yes            string
	No             string
	Abstain        string
	NoWithVeto     string
	TurnoutPercent string
}
