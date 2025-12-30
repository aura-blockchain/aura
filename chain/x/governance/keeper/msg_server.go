// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/governance/types"
	govpb "github.com/aequitas/aura/proto/aura/governance/v1beta1"
)

var _ govpb.MsgServer = (*msgServer)(nil)

type msgServer struct {
	govpb.UnimplementedMsgServer
	Keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper) govpb.MsgServer {
	return &msgServer{Keeper: keeper}
}

// SubmitProposal submits a new governance proposal
func (ms msgServer) SubmitProposal(goCtx context.Context, msg *govpb.MsgSubmitProposal) (*govpb.MsgSubmitProposalResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := validateProposal(msg); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	proposerAddr, err := sdk.AccAddressFromBech32(msg.Proposer)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !proposerAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "proposer must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get next proposal ID
	proposalID := ms.Keeper.GetNextProposalID(ctx)
	ms.Keeper.SetNextProposalID(ctx, proposalID+1)

	// Create proposal
	// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
	now := ctx.BlockTime()
	proposal := &types.Proposal{
		Id:          proposalID,
		Title:       msg.Title,
		Description: msg.Description,
		Category:    msg.Category,
		Status:      govpb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD,
		Proposer:    msg.Proposer,
		SubmitTime:  &gogotypes.Timestamp{Seconds: now.Unix(), Nanos: int32(now.Nanosecond())},
		IsEmergency: msg.IsEmergency,
	}

	// Set deposit and voting periods based on params
	params := ms.Keeper.GetParams(ctx)
	// Convert Duration to time.Duration manually
	depositDuration := time.Duration(params.MaxDepositPeriod.Seconds)*time.Second + time.Duration(params.MaxDepositPeriod.Nanos)*time.Nanosecond
	depositEndTime := ctx.BlockTime().Add(depositDuration)
	proposal.DepositEndTime = &gogotypes.Timestamp{Seconds: depositEndTime.Unix(), Nanos: int32(depositEndTime.Nanosecond())}

	// Store proposal
	if err := ms.Keeper.SetProposal(ctx, proposal); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Handle initial deposit if provided
	if msg.InitialDeposit != "" && msg.InitialDeposit != "0" {
		// Parse and validate deposit amount
		deposit, err := sdk.ParseCoinsNormalized(msg.InitialDeposit)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid deposit amount")
		}

		// Check minimum deposit requirement
		minDeposit, err := sdk.ParseCoinsNormalized(params.MinDeposit)
		if err != nil {
			return nil, status.Error(codes.Internal, "invalid minimum deposit parameter")
		}

		if deposit.IsAllLT(minDeposit) {
			return nil, status.Errorf(codes.InvalidArgument,
				"deposit %s below minimum %s", deposit, minDeposit)
		}

		// Actually transfer tokens from proposer to module account
		err = ms.Keeper.bankKeeper.SendCoinsFromAccountToModule(
			ctx,
			proposerAddr,
			types.ModuleName,
			deposit,
		)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition,
				"failed to transfer deposit: %s", err)
		}

		// Store deposit record
		// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
		depositNow := ctx.BlockTime()
		depositRecord := &types.Deposit{
			ProposalId: proposalID,
			Depositor:  msg.Proposer,
			Amount:     msg.InitialDeposit,
			Timestamp:  &gogotypes.Timestamp{Seconds: depositNow.Unix(), Nanos: int32(depositNow.Nanosecond())},
		}
		if err := ms.Keeper.SetDeposit(ctx, depositRecord); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitProposal,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute(types.AttributeKeyProposer, msg.Proposer),
		),
	)

	return &govpb.MsgSubmitProposalResponse{ProposalId: proposalID}, nil
}

// Deposit adds a deposit to a proposal
func (ms msgServer) Deposit(goCtx context.Context, msg *govpb.MsgDeposit) (*govpb.MsgDepositResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Amount == "" || msg.Amount == "0" {
		return nil, status.Error(codes.InvalidArgument, "deposit amount must be positive")
	}

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	depositorAddr, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !depositorAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "depositor must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if proposal exists
	proposal, err := ms.Keeper.GetProposal(ctx, msg.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}

	// Check if proposal is in deposit period
	if proposal.Status != govpb.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD {
		return nil, status.Error(codes.FailedPrecondition, "proposal not in deposit period")
	}

	// Parse and validate deposit amount
	depositAmount, err := sdk.ParseCoinsNormalized(msg.Amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid deposit amount")
	}

	// Actually transfer tokens from depositor to module account
	err = ms.Keeper.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		depositorAddr,
		types.ModuleName,
		depositAmount,
	)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition,
			"failed to transfer deposit: %s", err)
	}

	// Store deposit record
	// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
	depositNow := ctx.BlockTime()
	deposit := &types.Deposit{
		ProposalId: msg.ProposalId,
		Depositor:  msg.Depositor,
		Amount:     msg.Amount,
		Timestamp:  &gogotypes.Timestamp{Seconds: depositNow.Unix(), Nanos: int32(depositNow.Nanosecond())},
	}
	if err := ms.Keeper.SetDeposit(ctx, deposit); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDeposit,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyDepositor, msg.Depositor),
		),
	)

	return &govpb.MsgDepositResponse{}, nil
}

// Vote casts a vote on a proposal
func (ms msgServer) Vote(goCtx context.Context, msg *govpb.MsgVote) (*govpb.MsgVoteResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	voterAddr, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !voterAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "voter must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if proposal exists
	proposal, err := ms.Keeper.GetProposal(ctx, msg.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}

	// Check if proposal is in voting period
	if proposal.Status != govpb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
		return nil, status.Error(codes.FailedPrecondition, "proposal not in voting period")
	}

	// Get or create voting power snapshot for this voter
	// This is a performance optimization: voting power is cached at first vote
	// and reused for vote updates and tally calculations (O(1) instead of O(n))
	votingPower, err := ms.Keeper.GetOrCreateVotingPowerSnapshot(ctx, msg.ProposalId, msg.Voter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get voting power: %s", err)
	}

	// Check for existing vote to prevent double voting or allow vote update
	existingVote, err := ms.Keeper.GetVote(ctx, msg.ProposalId, msg.Voter)
	if err == nil && existingVote != nil {
		// Vote already exists - allow update by overwriting
		// This is standard governance behavior: users can change their vote during voting period
		// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
		voteUpdateNow := ctx.BlockTime()
		existingVote.Option = msg.Option
		existingVote.Timestamp = &gogotypes.Timestamp{Seconds: voteUpdateNow.Unix(), Nanos: int32(voteUpdateNow.Nanosecond())}
		existingVote.VotingPower = votingPower.String()

		// Handle secret ballot update
		if msg.IsSecret {
			existingVote.IsSecret = true
			existingVote.VoteCommitment = msg.VoteCommitment
		}

		if err := ms.Keeper.SetVote(ctx, existingVote); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Emit vote update event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeVote,
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
				sdk.NewAttribute(types.AttributeKeyVoter, msg.Voter),
				sdk.NewAttribute("action", "update"),
				sdk.NewAttribute("voting_power", votingPower.String()),
			),
		)

		return &govpb.MsgVoteResponse{}, nil
	}

	// Create new vote with cached voting power
	// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
	voteNow := ctx.BlockTime()
	vote := &types.Vote{
		ProposalId:  msg.ProposalId,
		Voter:       msg.Voter,
		Option:      msg.Option,
		Timestamp:   &gogotypes.Timestamp{Seconds: voteNow.Unix(), Nanos: int32(voteNow.Nanosecond())},
		IsSecret:    msg.IsSecret,
		VotingPower: votingPower.String(),
	}

	if msg.IsSecret {
		vote.VoteCommitment = msg.VoteCommitment
	}

	if err := ms.Keeper.SetVote(ctx, vote); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVote,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyVoter, msg.Voter),
			sdk.NewAttribute("action", "create"),
			sdk.NewAttribute("voting_power", votingPower.String()),
		),
	)

	return &govpb.MsgVoteResponse{}, nil
}

// VoteWeighted casts a weighted vote on a proposal
func (ms msgServer) VoteWeighted(goCtx context.Context, msg *govpb.MsgVoteWeighted) (*govpb.MsgVoteWeightedResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(msg.Options) == 0 {
		return nil, status.Error(codes.InvalidArgument, "weighted vote options cannot be empty")
	}

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	voterAddr, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !voterAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "voter must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if proposal exists
	proposal, err := ms.Keeper.GetProposal(ctx, msg.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}

	// Check if proposal is in voting period
	if proposal.Status != govpb.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD {
		return nil, status.Error(codes.FailedPrecondition, "proposal not in voting period")
	}

	// Convert message options to vote weighted options
	// SECURITY: Store ALL weighted options, not just the first one
	// This is critical for proper weighted voting where users can split their vote
	weightedOptions := make([]*govpb.WeightedVoteOption, len(msg.Options))
	for i, opt := range msg.Options {
		weightedOptions[i] = &govpb.WeightedVoteOption{
			Option: opt.Option,
			Weight: opt.Weight,
		}
	}

	// Check for existing vote to prevent double voting or allow vote update
	existingVote, err := ms.Keeper.GetVote(ctx, msg.ProposalId, msg.Voter)
	if err == nil && existingVote != nil {
		// Vote already exists - allow update by overwriting
		// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
		voteWeightedUpdateNow := ctx.BlockTime()
		existingVote.Option = msg.Options[0].Option
		existingVote.WeightedOptions = weightedOptions // Store ALL weighted options
		existingVote.Timestamp = &gogotypes.Timestamp{Seconds: voteWeightedUpdateNow.Unix(), Nanos: int32(voteWeightedUpdateNow.Nanosecond())}

		if err := ms.Keeper.SetVote(ctx, existingVote); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Emit vote update event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeVote,
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
				sdk.NewAttribute(types.AttributeKeyVoter, msg.Voter),
				sdk.NewAttribute("action", "update"),
			),
		)

		return &govpb.MsgVoteWeightedResponse{}, nil
	}

	// Create new weighted vote with ALL options stored
	// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
	voteWeightedNow := ctx.BlockTime()
	vote := &types.Vote{
		ProposalId:      msg.ProposalId,
		Voter:           msg.Voter,
		Option:          msg.Options[0].Option,          // Primary option for backward compatibility
		WeightedOptions: weightedOptions,                 // ALL weighted options for proper tallying
		Timestamp:       &gogotypes.Timestamp{Seconds: voteWeightedNow.Unix(), Nanos: int32(voteWeightedNow.Nanosecond())},
	}

	if err := ms.Keeper.SetVote(ctx, vote); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVote,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyVoter, msg.Voter),
			sdk.NewAttribute("action", "create"),
		),
	)

	return &govpb.MsgVoteWeightedResponse{}, nil
}

// DelegateVote delegates voting power to another address
func (ms msgServer) DelegateVote(goCtx context.Context, msg *govpb.MsgDelegateVote) (*govpb.MsgDelegateVoteResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Delegator == msg.Delegate {
		return nil, status.Error(codes.InvalidArgument, "cannot delegate to self")
	}

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	delegatorAddr, err := sdk.AccAddressFromBech32(msg.Delegator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !delegatorAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "delegator must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Store vote delegation
	delegation := &types.VoteDelegation{
		Delegator:  msg.Delegator,
		Delegate:   msg.Delegate,
		Categories: msg.Categories,
	}

	if err := ms.Keeper.SetVoteDelegation(ctx, delegation); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDelegateVote,
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.Delegator),
			sdk.NewAttribute(types.AttributeKeyDelegate, msg.Delegate),
		),
	)

	return &govpb.MsgDelegateVoteResponse{}, nil
}

// UndelegateVote removes vote delegation
func (ms msgServer) UndelegateVote(goCtx context.Context, msg *govpb.MsgUndelegateVote) (*govpb.MsgUndelegateVoteResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	delegatorAddr, err := sdk.AccAddressFromBech32(msg.Delegator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !delegatorAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "delegator must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Remove vote delegation
	if err := ms.Keeper.DeleteVoteDelegation(ctx, msg.Delegator, msg.Delegate); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUndelegateVote,
			sdk.NewAttribute(types.AttributeKeyDelegator, msg.Delegator),
			sdk.NewAttribute(types.AttributeKeyDelegate, msg.Delegate),
		),
	)

	return &govpb.MsgUndelegateVoteResponse{}, nil
}

// SubmitVeto submits a veto request
func (ms msgServer) SubmitVeto(goCtx context.Context, msg *govpb.MsgSubmitVeto) (*govpb.MsgSubmitVetoResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	vetoerAddr, err := sdk.AccAddressFromBech32(msg.Vetoer)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !vetoerAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "vetoer must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if proposal exists
	_, err = ms.Keeper.GetProposal(ctx, msg.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}

	// Store veto request
	// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
	vetoNow := ctx.BlockTime()
	veto := &types.VetoRequest{
		ProposalId: msg.ProposalId,
		Vetoer:     msg.Vetoer,
		Reason:     msg.Reason,
		Timestamp:  &gogotypes.Timestamp{Seconds: vetoNow.Unix(), Nanos: int32(vetoNow.Nanosecond())},
		Cosigners:  []string{},
	}

	if err := ms.Keeper.SetVetoRequest(ctx, veto); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVeto,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyVetoer, msg.Vetoer),
		),
	)

	return &govpb.MsgSubmitVetoResponse{VetoExecuted: false}, nil
}

// CosignVeto cosigns an existing veto request
func (ms msgServer) CosignVeto(goCtx context.Context, msg *govpb.MsgCosignVeto) (*govpb.MsgCosignVetoResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// Validate address BEFORE calling GetSigners() to avoid panic
	cosignerAddr, err := sdk.AccAddressFromBech32(msg.Cosigner)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	if !cosignerAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "cosigner must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get veto request
	veto, err := ms.Keeper.GetVetoRequest(ctx, msg.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "veto request not found")
	}

	// Check for duplicate cosigner - prevent same address from cosigning twice
	for _, existingCosigner := range veto.Cosigners {
		if existingCosigner == msg.Cosigner {
			return nil, status.Error(codes.AlreadyExists, types.ErrDuplicateCosigner.Error())
		}
	}

	// Add cosigner
	veto.Cosigners = append(veto.Cosigners, msg.Cosigner)

	// Update veto request
	if err := ms.Keeper.SetVetoRequest(ctx, veto); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVeto,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyVetoer, msg.Cosigner),
		),
	)

	return &govpb.MsgCosignVetoResponse{VetoExecuted: false}, nil
}

// ExecuteProposal executes a passed proposal after time-lock
func (ms msgServer) ExecuteProposal(goCtx context.Context, msg *govpb.MsgExecuteProposal) (*govpb.MsgExecuteProposalResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	executorAddr, err := sdk.AccAddressFromBech32(msg.Executor)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !executorAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "executor must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get proposal
	proposal, err := ms.Keeper.GetProposal(ctx, msg.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}

	// Check if proposal is ready for execution
	if proposal.Status != govpb.ProposalStatus_PROPOSAL_STATUS_READY_FOR_EXECUTION &&
		proposal.Status != govpb.ProposalStatus_PROPOSAL_STATUS_PASSED {
		return nil, status.Error(codes.FailedPrecondition, "proposal not ready for execution")
	}

	// Update proposal status
	proposal.Status = govpb.ProposalStatus_PROPOSAL_STATUS_EXECUTED
	if err := ms.Keeper.SetProposal(ctx, proposal); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecuteProposal,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyExecutor, msg.Executor),
		),
	)

	return &govpb.MsgExecuteProposalResponse{}, nil
}

// SubmitSnapshotVote submits an off-chain snapshot vote
func (ms msgServer) SubmitSnapshotVote(goCtx context.Context, msg *govpb.MsgSubmitSnapshotVote) (*govpb.MsgSubmitSnapshotVoteResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	voterAddr, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if !voterAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "voter must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if proposal exists
	_, err = ms.Keeper.GetProposal(ctx, msg.ProposalId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "proposal not found")
	}

	// Check for existing snapshot vote to prevent double voting or allow vote update
	existingVote, err := ms.Keeper.GetSnapshotVote(ctx, msg.ProposalId, msg.Voter)
	if err == nil && existingVote != nil {
		// Snapshot vote already exists - allow update
		// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
		snapshotUpdateNow := ctx.BlockTime()
		existingVote.Option = msg.Option
		existingVote.Signature = msg.Signature
		existingVote.Timestamp = &gogotypes.Timestamp{Seconds: snapshotUpdateNow.Unix(), Nanos: int32(snapshotUpdateNow.Nanosecond())}

		if err := ms.Keeper.SetSnapshotVote(ctx, existingVote); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		// Emit snapshot vote update event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSnapshotVote,
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
				sdk.NewAttribute(types.AttributeKeyVoter, msg.Voter),
				sdk.NewAttribute("action", "update"),
			),
		)

		return &govpb.MsgSubmitSnapshotVoteResponse{}, nil
	}

	// Create new snapshot vote
	// Use ctx.BlockTime() for determinism - NEVER use time.Now() in consensus code
	snapshotNow := ctx.BlockTime()
	snapshotVote := &types.SnapshotVote{
		ProposalId: msg.ProposalId,
		Voter:      msg.Voter,
		Option:     msg.Option,
		Signature:  msg.Signature,
		Timestamp:  &gogotypes.Timestamp{Seconds: snapshotNow.Unix(), Nanos: int32(snapshotNow.Nanosecond())},
	}

	if err := ms.Keeper.SetSnapshotVote(ctx, snapshotVote); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSnapshotVote,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyVoter, msg.Voter),
			sdk.NewAttribute("action", "create"),
		),
	)

	return &govpb.MsgSubmitSnapshotVoteResponse{}, nil
}

// RevealSecretVote reveals a secret ballot vote
func (ms msgServer) RevealSecretVote(goCtx context.Context, msg *govpb.MsgRevealSecretVote) (*govpb.MsgRevealSecretVoteResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// Validate address BEFORE calling GetSigners() to avoid panic
	voterAddr, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	if !voterAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "voter must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get existing vote
	vote, err := ms.Keeper.GetVote(ctx, msg.ProposalId, msg.Voter)
	if err != nil {
		return nil, status.Error(codes.NotFound, "vote not found")
	}

	// Verify vote was secret
	if !vote.IsSecret {
		return nil, status.Error(codes.InvalidArgument, "vote is not secret")
	}

	// Update vote with revealed option
	vote.Option = msg.Option
	vote.IsSecret = false
	if err := ms.Keeper.SetVote(ctx, vote); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRevealVote,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyVoter, msg.Voter),
		),
	)

	return &govpb.MsgRevealSecretVoteResponse{}, nil
}

// validateProposal validates a proposal submission
func validateProposal(msg *govpb.MsgSubmitProposal) error {
	if msg.Title == "" {
		return fmt.Errorf("proposal title cannot be empty")
	}
	if msg.Description == "" {
		return fmt.Errorf("proposal description cannot be empty")
	}
	if msg.Proposer == "" {
		return fmt.Errorf("proposer cannot be empty")
	}
	return nil
}
