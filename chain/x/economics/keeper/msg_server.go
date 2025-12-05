package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	economicspb "github.com/aequitas/aura/proto/aura/economics/v1beta1"
	"github.com/aequitas/aura/chain/x/economics/types"
)

// Ensure MsgServer implements the MsgServer interface
var _ economicspb.MsgServer = msgServer{}

// msgServer is the message server implementation
type msgServer struct {
	economicspb.UnimplementedMsgServer
	keeper *Keeper
}

// NewMsgServer creates a new message server
func NewMsgServer(keeper *Keeper) economicspb.MsgServer {
	return &msgServer{keeper: keeper}
}

// CreateVestingSchedule creates a new vesting schedule
func (ms msgServer) CreateVestingSchedule(goCtx context.Context, msg *economicspb.MsgCreateVestingSchedule) (*economicspb.MsgCreateVestingScheduleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid creator address: %s", err)
	}

	// Validate beneficiary address
	beneficiary, err := sdk.AccAddressFromBech32(msg.BeneficiaryAddress)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid beneficiary address: %s", err)
	}

	// Create schedule via keeper
	scheduleID, err := ms.keeper.CreateVestingSchedule(
		ctx,
		creator.String(),
		beneficiary.String(),
		*msg.TotalAmount,
		msg.StartTime.AsTime(),
		msg.CliffDuration,
		msg.VestingDuration,
		msg.VestingType,
		msg.ScheduleType,
	)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateVestingSchedule,
			sdk.NewAttribute(types.AttributeKeyCreator, creator.String()),
			sdk.NewAttribute(types.AttributeKeyBeneficiary, beneficiary.String()),
			sdk.NewAttribute(types.AttributeKeyScheduleID, scheduleID),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.TotalAmount.String()),
		),
	)

	return &economicspb.MsgCreateVestingScheduleResponse{
		ScheduleId: scheduleID,
	}, nil
}

// ReleaseVestedTokens releases vested tokens to beneficiary
func (ms msgServer) ReleaseVestedTokens(goCtx context.Context, msg *economicspb.MsgReleaseVestedTokens) (*economicspb.MsgReleaseVestedTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate beneficiary address
	beneficiary, err := sdk.AccAddressFromBech32(msg.Beneficiary)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid beneficiary address: %s", err)
	}

	// Release tokens via keeper
	amountReleased, err := ms.keeper.ReleaseVestedTokens(ctx, beneficiary, msg.ScheduleId)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReleaseVestedTokens,
			sdk.NewAttribute(types.AttributeKeyBeneficiary, beneficiary.String()),
			sdk.NewAttribute(types.AttributeKeyScheduleID, msg.ScheduleId),
			sdk.NewAttribute(types.AttributeKeyAmount, amountReleased.String()),
		),
	)

	return &economicspb.MsgReleaseVestedTokensResponse{
		AmountReleased: &amountReleased,
	}, nil
}

// RevokeVestingSchedule revokes a vesting schedule
func (ms msgServer) RevokeVestingSchedule(goCtx context.Context, msg *economicspb.MsgRevokeVestingSchedule) (*economicspb.MsgRevokeVestingScheduleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate revoker address
	revoker, err := sdk.AccAddressFromBech32(msg.Revoker)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid revoker address: %s", err)
	}

	// Revoke schedule via keeper
	unvestedAmount, err := ms.keeper.RevokeVestingSchedule(ctx, revoker, msg.ScheduleId, msg.Reason)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRevokeVestingSchedule,
			sdk.NewAttribute(types.AttributeKeyRevoker, revoker.String()),
			sdk.NewAttribute(types.AttributeKeyScheduleID, msg.ScheduleId),
			sdk.NewAttribute(types.AttributeKeyAmount, unvestedAmount.String()),
			sdk.NewAttribute(types.AttributeKeyReason, msg.Reason),
		),
	)

	return &economicspb.MsgRevokeVestingScheduleResponse{
		UnvestedAmountReturned: &unvestedAmount,
	}, nil
}

// SubmitProposal submits a new governance proposal
func (ms msgServer) SubmitProposal(goCtx context.Context, msg *economicspb.MsgSubmitProposal) (*economicspb.MsgSubmitProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate proposer address
	proposer, err := sdk.AccAddressFromBech32(msg.Proposer)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid proposer address: %s", err)
	}

	// Convert coins from proto format
	depositCoins := make([]sdk.Coin, len(msg.InitialDeposit))
	for i, coin := range msg.InitialDeposit {
		depositCoins[i] = *coin
	}
	initialDeposit := sdk.NewCoins(depositCoins...)

	// Submit proposal via keeper
	proposalID, err := ms.keeper.SubmitProposal(
		ctx,
		msg.Title,
		msg.Description,
		msg.Category,
		proposer,
		initialDeposit,
		msg.IsEmergency,
	)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitProposal,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute(types.AttributeKeyProposer, proposer.String()),
			sdk.NewAttribute(types.AttributeKeyProposalTitle, msg.Title),
		),
	)

	return &economicspb.MsgSubmitProposalResponse{
		ProposalId: proposalID,
	}, nil
}

// Deposit adds a deposit to a proposal
func (ms msgServer) Deposit(goCtx context.Context, msg *economicspb.MsgDeposit) (*economicspb.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate depositor address
	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid depositor address: %s", err)
	}

	// Convert coins from proto format
	coins := make([]sdk.Coin, len(msg.Amount))
	for i, coin := range msg.Amount {
		coins[i] = *coin
	}
	amount := sdk.NewCoins(coins...)

	// Add deposit via keeper
	if err := ms.keeper.AddDeposit(ctx, msg.ProposalId, depositor, amount); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDeposit,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	)

	return &economicspb.MsgDepositResponse{}, nil
}

// Vote casts a vote on a proposal
func (ms msgServer) Vote(goCtx context.Context, msg *economicspb.MsgVote) (*economicspb.MsgVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate voter address
	voter, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid voter address: %s", err)
	}

	// Cast vote via keeper
	if err := ms.keeper.AddVote(ctx, msg.ProposalId, voter, msg.Option, msg.IsSecret, msg.VoteCommitment); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVote,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyVoter, voter.String()),
			sdk.NewAttribute(types.AttributeKeyOption, msg.Option.String()),
		),
	)

	return &economicspb.MsgVoteResponse{}, nil
}

// VoteWeighted casts a weighted vote on a proposal
func (ms msgServer) VoteWeighted(goCtx context.Context, msg *economicspb.MsgVoteWeighted) (*economicspb.MsgVoteWeightedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate voter address
	voter, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid voter address: %s", err)
	}

	// Cast weighted vote via keeper
	if err := ms.keeper.AddWeightedVote(ctx, msg.ProposalId, voter, msg.Options); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVote,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyVoter, voter.String()),
			sdk.NewAttribute(types.AttributeKeyOption, "weighted"),
		),
	)

	return &economicspb.MsgVoteWeightedResponse{}, nil
}

// DelegateVote delegates voting power to another address
func (ms msgServer) DelegateVote(goCtx context.Context, msg *economicspb.MsgDelegateVote) (*economicspb.MsgDelegateVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate addresses
	delegator, err := sdk.AccAddressFromBech32(msg.Delegator)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid delegator address: %s", err)
	}

	delegate, err := sdk.AccAddressFromBech32(msg.Delegate)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid delegate address: %s", err)
	}

	// Delegate vote via keeper
	if err := ms.keeper.DelegateVote(ctx, delegator, delegate, msg.Categories); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDelegateVote,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegator.String()),
			sdk.NewAttribute(types.AttributeKeyDelegate, delegate.String()),
		),
	)

	return &economicspb.MsgDelegateVoteResponse{}, nil
}

// UndelegateVote removes vote delegation
func (ms msgServer) UndelegateVote(goCtx context.Context, msg *economicspb.MsgUndelegateVote) (*economicspb.MsgUndelegateVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate addresses
	delegator, err := sdk.AccAddressFromBech32(msg.Delegator)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid delegator address: %s", err)
	}

	delegate, err := sdk.AccAddressFromBech32(msg.Delegate)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid delegate address: %s", err)
	}

	// Undelegate vote via keeper
	if err := ms.keeper.UndelegateVote(ctx, delegator, delegate, msg.Categories); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUndelegateVote,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegator.String()),
			sdk.NewAttribute(types.AttributeKeyDelegate, delegate.String()),
		),
	)

	return &economicspb.MsgUndelegateVoteResponse{}, nil
}

// ExecuteProposal executes a passed proposal after timelock
func (ms msgServer) ExecuteProposal(goCtx context.Context, msg *economicspb.MsgExecuteProposal) (*economicspb.MsgExecuteProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate executor address
	executor, err := sdk.AccAddressFromBech32(msg.Executor)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid executor address: %s", err)
	}

	// Execute proposal via keeper
	if err := ms.keeper.ExecuteProposal(ctx, msg.ProposalId, executor); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecuteProposal,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyExecutor, executor.String()),
		),
	)

	return &economicspb.MsgExecuteProposalResponse{}, nil
}

// RevealSecretVote reveals a secret ballot vote
func (ms msgServer) RevealSecretVote(goCtx context.Context, msg *economicspb.MsgRevealSecretVote) (*economicspb.MsgRevealSecretVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate voter address
	voter, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid voter address: %s", err)
	}

	// Reveal vote via keeper
	if err := ms.keeper.RevealSecretVote(ctx, msg.ProposalId, voter, msg.Option, msg.RevealKey); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRevealVote,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", msg.ProposalId)),
			sdk.NewAttribute(types.AttributeKeyVoter, voter.String()),
			sdk.NewAttribute(types.AttributeKeyOption, msg.Option.String()),
		),
	)

	return &economicspb.MsgRevealSecretVoteResponse{}, nil
}

// LockVotingTokens locks tokens for voting power boost
func (ms msgServer) LockVotingTokens(goCtx context.Context, msg *economicspb.MsgLockVotingTokens) (*economicspb.MsgLockVotingTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate owner address
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid owner address: %s", err)
	}

	// Lock tokens via keeper
	lockID, votingPower, err := ms.keeper.LockVotingTokens(ctx, owner, *msg.Amount, msg.LockDuration)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLockVotingTokens,
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
			sdk.NewAttribute(types.AttributeKeyLockID, lockID),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyVotingPower, votingPower.String()),
		),
	)

	return &economicspb.MsgLockVotingTokensResponse{
		LockId:      lockID,
		VotingPower: votingPower.String(),
	}, nil
}

// UnlockVotingTokens unlocks voting tokens after lock period
func (ms msgServer) UnlockVotingTokens(goCtx context.Context, msg *economicspb.MsgUnlockVotingTokens) (*economicspb.MsgUnlockVotingTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate owner address
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid owner address: %s", err)
	}

	// Unlock tokens via keeper
	amountUnlocked, err := ms.keeper.UnlockVotingTokens(ctx, owner, msg.LockId)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnlockVotingTokens,
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
			sdk.NewAttribute(types.AttributeKeyLockID, msg.LockId),
			sdk.NewAttribute(types.AttributeKeyAmount, amountUnlocked.String()),
		),
	)

	return &economicspb.MsgUnlockVotingTokensResponse{
		AmountUnlocked: &amountUnlocked,
	}, nil
}

// ProposeTreasurySpend proposes a treasury spend
func (ms msgServer) ProposeTreasurySpend(goCtx context.Context, msg *economicspb.MsgProposeTreasurySpend) (*economicspb.MsgProposeTreasurySpendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate proposer address
	proposer, err := sdk.AccAddressFromBech32(msg.Proposer)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid proposer address: %s", err)
	}

	// Validate recipient address
	recipient, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid recipient address: %s", err)
	}

	// Convert coins from proto format
	treasuryCoins := make([]sdk.Coin, len(msg.Amount))
	for i, coin := range msg.Amount {
		treasuryCoins[i] = *coin
	}
	amount := sdk.NewCoins(treasuryCoins...)

	// Propose treasury spend via keeper
	txID, executableAt, err := ms.keeper.ProposeTreasurySpend(ctx, proposer, recipient, amount, msg.Description)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposeTreasurySpend,
			sdk.NewAttribute(types.AttributeKeyProposer, proposer.String()),
			sdk.NewAttribute(types.AttributeKeyRecipient, recipient.String()),
			sdk.NewAttribute(types.AttributeKeyTxID, txID),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	)

	return &economicspb.MsgProposeTreasurySpendResponse{
		TxId:         txID,
		ExecutableAt: executableAt,
	}, nil
}

// SignTreasurySpend signs a treasury spend proposal
func (ms msgServer) SignTreasurySpend(goCtx context.Context, msg *economicspb.MsgSignTreasurySpend) (*economicspb.MsgSignTreasurySpendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate signer address
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid signer address: %s", err)
	}

	// Sign treasury spend via keeper
	currentSigs, requiredSigs, err := ms.keeper.SignTreasurySpend(ctx, signer, msg.TxId)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSignTreasurySpend,
			sdk.NewAttribute(types.AttributeKeySigner, signer.String()),
			sdk.NewAttribute(types.AttributeKeyTxID, msg.TxId),
			sdk.NewAttribute(types.AttributeKeySignatures, fmt.Sprintf("%d/%d", currentSigs, requiredSigs)),
		),
	)

	return &economicspb.MsgSignTreasurySpendResponse{
		CurrentSignatures:  currentSigs,
		RequiredSignatures: requiredSigs,
	}, nil
}

// ExecuteTreasurySpend executes an approved treasury spend
func (ms msgServer) ExecuteTreasurySpend(goCtx context.Context, msg *economicspb.MsgExecuteTreasurySpend) (*economicspb.MsgExecuteTreasurySpendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate executor address
	executor, err := sdk.AccAddressFromBech32(msg.Executor)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid executor address: %s", err)
	}

	// Execute treasury spend via keeper
	success, err := ms.keeper.ExecuteTreasurySpend(ctx, executor, msg.TxId)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecuteTreasurySpend,
			sdk.NewAttribute(types.AttributeKeyExecutor, executor.String()),
			sdk.NewAttribute(types.AttributeKeyTxID, msg.TxId),
			sdk.NewAttribute(types.AttributeKeySuccess, fmt.Sprintf("%t", success)),
		),
	)

	return &economicspb.MsgExecuteTreasurySpendResponse{
		Success: success,
	}, nil
}

// UpdateParams updates module parameters (governance only)
func (ms msgServer) UpdateParams(goCtx context.Context, msg *economicspb.MsgUpdateParams) (*economicspb.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority address
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid authority address: %s", err)
	}

	// Update params via keeper
	if err := ms.keeper.UpdateParams(ctx, authority, msg.Params); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateParams,
			sdk.NewAttribute(types.AttributeKeyAuthority, authority.String()),
		),
	)

	return &economicspb.MsgUpdateParamsResponse{}, nil
}

// AdjustInflationRate manually adjusts inflation rate (governance only)
func (ms msgServer) AdjustInflationRate(goCtx context.Context, msg *economicspb.MsgAdjustInflationRate) (*economicspb.MsgAdjustInflationRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate authority address
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "invalid authority address: %s", err)
	}

	// Adjust inflation rate via keeper
	oldRate, newRate, err := ms.keeper.AdjustInflationRate(ctx, authority, msg.NewRate, msg.Reason)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAdjustInflationRate,
			sdk.NewAttribute(types.AttributeKeyAuthority, authority.String()),
			sdk.NewAttribute(types.AttributeKeyOldRate, fmt.Sprintf("%d", oldRate)),
			sdk.NewAttribute(types.AttributeKeyNewRate, fmt.Sprintf("%d", newRate)),
			sdk.NewAttribute(types.AttributeKeyReason, msg.Reason),
		),
	)

	return &economicspb.MsgAdjustInflationRateResponse{
		OldRate: oldRate,
		NewRate: newRate,
	}, nil
}
