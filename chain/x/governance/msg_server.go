package governance

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/governance/keeper"
	"github.com/aequitas/aura/chain/x/governance/types"
)

// MsgServer implements the governance Msg service
type MsgServer struct {
	types.UnimplementedMsgServer
	keeper *keeper.Keeper
}

// NewMsgServer creates a new MsgServer
func NewMsgServer(k *keeper.Keeper) *MsgServer {
	return &MsgServer{keeper: k}
}

// SubmitProposal handles proposal submission
func (m *MsgServer) SubmitProposal(ctx context.Context, msg *MsgSubmitProposal) (*MsgSubmitProposalResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	// Validate message
	if msg.Title == "" {
		return nil, fmt.Errorf("proposal title cannot be empty")
	}
	if msg.Description == "" {
		return nil, fmt.Errorf("proposal description cannot be empty")
	}
	if msg.Proposer == "" {
		return nil, fmt.Errorf("proposer cannot be empty")
	}

	proposalID, err := m.keeper.SubmitProposal(
		msg.Title,
		msg.Description,
		types.ProposalCategory(msg.Category),
		msg.Proposer,
		msg.InitialDeposit,
		msg.IsEmergency,
	)
	if err != nil {
		return nil, err
	}

	return &MsgSubmitProposalResponse{ProposalID: proposalID}, nil
}

// Deposit handles deposit additions
func (m *MsgServer) Deposit(ctx context.Context, msg *MsgDeposit) (*MsgDepositResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	if msg.Depositor == "" {
		return nil, fmt.Errorf("depositor cannot be empty")
	}
	if msg.Amount == "" || msg.Amount == "0" {
		return nil, fmt.Errorf("deposit amount must be positive")
	}

	err := m.keeper.AddDeposit(msg.ProposalID, msg.Depositor, msg.Amount)
	if err != nil {
		return nil, err
	}

	return &MsgDepositResponse{}, nil
}

// Vote handles vote casting
func (m *MsgServer) Vote(ctx context.Context, msg *MsgVote) (*MsgVoteResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	if msg.Voter == "" {
		return nil, fmt.Errorf("voter cannot be empty")
	}

	// In a real implementation, voting power would be retrieved from staking module
	votingPower := "1000000"

	err := m.keeper.CastVote(
		msg.ProposalID,
		msg.Voter,
		types.VoteOption(msg.Option),
		votingPower,
		msg.IsSecret,
		msg.VoteCommitment,
	)
	if err != nil {
		return nil, err
	}

	return &MsgVoteResponse{}, nil
}

// VoteWeighted handles weighted vote casting
func (m *MsgServer) VoteWeighted(ctx context.Context, msg *MsgVoteWeighted) (*MsgVoteWeightedResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	if msg.Voter == "" {
		return nil, fmt.Errorf("voter cannot be empty")
	}

	// Validate weights sum to 1
	totalWeight := 0.0
	for _, opt := range msg.Options {
		weight := 0.0
		fmt.Sscanf(opt.Weight, "%f", &weight)
		totalWeight += weight
	}
	if totalWeight != 1.0 {
		return nil, types.ErrInvalidWeight
	}

	// For simplicity, cast vote with highest weight option
	highestWeight := 0.0
	var selectedOption types.VoteOption
	for _, opt := range msg.Options {
		weight := 0.0
		fmt.Sscanf(opt.Weight, "%f", &weight)
		if weight > highestWeight {
			highestWeight = weight
			selectedOption = types.VoteOption(opt.Option)
		}
	}

	votingPower := "1000000"
	err := m.keeper.CastVote(msg.ProposalID, msg.Voter, selectedOption, votingPower, false, "")
	if err != nil {
		return nil, err
	}

	return &MsgVoteWeightedResponse{}, nil
}

// DelegateVote handles vote delegation
func (m *MsgServer) DelegateVote(ctx context.Context, msg *MsgDelegateVote) (*MsgDelegateVoteResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	if msg.Delegator == "" {
		return nil, fmt.Errorf("delegator cannot be empty")
	}
	if msg.Delegate == "" {
		return nil, fmt.Errorf("delegate cannot be empty")
	}

	// In a real implementation, voting power would be retrieved from staking module
	votingPower := "1000000"

	categories := make([]types.ProposalCategory, len(msg.Categories))
	for i, cat := range msg.Categories {
		categories[i] = types.ProposalCategory(cat)
	}

	err := m.keeper.DelegateVote(msg.Delegator, msg.Delegate, votingPower, categories)
	if err != nil {
		return nil, err
	}

	return &MsgDelegateVoteResponse{}, nil
}

// UndelegateVote handles vote undelegation
func (m *MsgServer) UndelegateVote(ctx context.Context, msg *MsgUndelegateVote) (*MsgUndelegateVoteResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	if msg.Delegator == "" {
		return nil, fmt.Errorf("delegator cannot be empty")
	}
	if msg.Delegate == "" {
		return nil, fmt.Errorf("delegate cannot be empty")
	}

	categories := make([]types.ProposalCategory, len(msg.Categories))
	for i, cat := range msg.Categories {
		categories[i] = types.ProposalCategory(cat)
	}

	err := m.keeper.UndelegateVote(msg.Delegator, msg.Delegate, categories)
	if err != nil {
		return nil, err
	}

	return &MsgUndelegateVoteResponse{}, nil
}

// SubmitVeto handles veto submission
func (m *MsgServer) SubmitVeto(ctx context.Context, msg *MsgSubmitVeto) (*MsgSubmitVetoResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	if msg.Vetoer == "" {
		return nil, fmt.Errorf("vetoer cannot be empty")
	}
	if msg.Reason == "" {
		return nil, fmt.Errorf("veto reason cannot be empty")
	}

	vetoExecuted, err := m.keeper.SubmitVeto(msg.ProposalID, msg.Vetoer, msg.Reason)
	if err != nil {
		return nil, err
	}

	return &MsgSubmitVetoResponse{VetoExecuted: vetoExecuted}, nil
}

// CosignVeto handles veto cosigning
func (m *MsgServer) CosignVeto(ctx context.Context, msg *MsgCosignVeto) (*MsgCosignVetoResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	if msg.Cosigner == "" {
		return nil, fmt.Errorf("cosigner cannot be empty")
	}

	vetoExecuted, err := m.keeper.CosignVeto(msg.ProposalID, msg.Cosigner)
	if err != nil {
		return nil, err
	}

	return &MsgCosignVetoResponse{VetoExecuted: vetoExecuted}, nil
}

// ExecuteProposal handles proposal execution
func (m *MsgServer) ExecuteProposal(ctx context.Context, msg *MsgExecuteProposal) (*MsgExecuteProposalResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	if msg.Executor == "" {
		return nil, fmt.Errorf("executor cannot be empty")
	}

	err := m.keeper.ExecuteProposal(msg.ProposalID, msg.Executor)
	if err != nil {
		return nil, err
	}

	return &MsgExecuteProposalResponse{}, nil
}

// SubmitSnapshotVote handles snapshot vote submission
func (m *MsgServer) SubmitSnapshotVote(ctx context.Context, msg *MsgSubmitSnapshotVote) (*MsgSubmitSnapshotVoteResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	if msg.Voter == "" {
		return nil, fmt.Errorf("voter cannot be empty")
	}
	if msg.Signature == "" {
		return nil, fmt.Errorf("signature cannot be empty")
	}

	// In a real implementation, voting power would be calculated at snapshot height
	votingPower := "1000000"

	err := m.keeper.SubmitSnapshotVote(
		msg.ProposalID,
		msg.Voter,
		types.VoteOption(msg.Option),
		votingPower,
		msg.Signature,
	)
	if err != nil {
		return nil, err
	}

	return &MsgSubmitSnapshotVoteResponse{}, nil
}

// RevealSecretVote handles secret vote reveal
func (m *MsgServer) RevealSecretVote(ctx context.Context, msg *MsgRevealSecretVote) (*MsgRevealSecretVoteResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil message")
	}

	if msg.Voter == "" {
		return nil, fmt.Errorf("voter cannot be empty")
	}
	if msg.RevealKey == "" {
		return nil, fmt.Errorf("reveal key cannot be empty")
	}

	err := m.keeper.RevealSecretVote(
		msg.ProposalID,
		msg.Voter,
		types.VoteOption(msg.Option),
		msg.RevealKey,
	)
	if err != nil {
		return nil, err
	}

	return &MsgRevealSecretVoteResponse{}, nil
}

// Message types (would normally be generated from proto)
type MsgSubmitProposal struct {
	Title          string
	Description    string
	Category       int32
	Proposer       string
	InitialDeposit string
	IsEmergency    bool
}

type MsgSubmitProposalResponse struct {
	ProposalID uint64
}

type MsgDeposit struct {
	ProposalID uint64
	Depositor  string
	Amount     string
}

type MsgDepositResponse struct{}

type MsgVote struct {
	ProposalID     uint64
	Voter          string
	Option         int32
	IsSecret       bool
	VoteCommitment string
}

type MsgVoteResponse struct{}

type WeightedVoteOption struct {
	Option int32
	Weight string
}

type MsgVoteWeighted struct {
	ProposalID uint64
	Voter      string
	Options    []WeightedVoteOption
}

type MsgVoteWeightedResponse struct{}

type MsgDelegateVote struct {
	Delegator  string
	Delegate   string
	Categories []int32
}

type MsgDelegateVoteResponse struct{}

type MsgUndelegateVote struct {
	Delegator  string
	Delegate   string
	Categories []int32
}

type MsgUndelegateVoteResponse struct{}

type MsgSubmitVeto struct {
	ProposalID uint64
	Vetoer     string
	Reason     string
}

type MsgSubmitVetoResponse struct {
	VetoExecuted bool
}

type MsgCosignVeto struct {
	ProposalID uint64
	Cosigner   string
}

type MsgCosignVetoResponse struct {
	VetoExecuted bool
}

type MsgExecuteProposal struct {
	ProposalID uint64
	Executor   string
}

type MsgExecuteProposalResponse struct{}

type MsgSubmitSnapshotVote struct {
	ProposalID uint64
	Voter      string
	Option     int32
	Signature  string
}

type MsgSubmitSnapshotVoteResponse struct{}

type MsgRevealSecretVote struct {
	ProposalID uint64
	Voter      string
	Option     int32
	RevealKey  string
}

type MsgRevealSecretVoteResponse struct{}
