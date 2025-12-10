package keeper

import (
	"context"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	economicsecuritypb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// MsgServer implements the economicsecurity Msg service
type MsgServer struct {
	economicsecuritypb.UnimplementedMsgServer
	keeper *Keeper
}

// NewMsgServer returns a new MsgServer
func NewMsgServer(k *Keeper) economicsecuritypb.MsgServer {
	return &MsgServer{keeper: k}
}

var _ economicsecuritypb.MsgServer = &MsgServer{}

// ============================
// VESTING SCHEDULE HANDLERS
// ============================

// CreateVestingSchedule creates a new vesting schedule
func (m *MsgServer) CreateVestingSchedule(
	goCtx context.Context,
	msg *types.MsgCreateVestingSchedule,
) (*types.MsgCreateVestingScheduleResponse, error) {
	// Validate message
	if msg == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "message cannot be nil")
	}

	// Validate creator address
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "invalid creator address")
	}

	// Validate beneficiary address
	if _, err := sdk.AccAddressFromBech32(msg.BeneficiaryAddress); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidBeneficiary, "invalid beneficiary address")
	}

	// Validate amounts and durations
	if msg.TotalAmount == "" || msg.TotalAmount == "0" {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "total amount must be positive")
	}

	if msg.VestingDuration == 0 {
		return nil, errorsmod.Wrap(types.ErrInvalidDuration, "vesting duration must be positive")
	}

	if msg.CliffDuration > msg.VestingDuration {
		return nil, errorsmod.Wrap(types.ErrInvalidDuration, "cliff duration cannot exceed vesting duration")
	}

	// Convert gogoproto Timestamp to time.Time
	// Use ctx.BlockTime() as fallback for determinism in consensus.
	// NEVER use time.Now() - it causes non-determinism across validators.
	ctx := sdk.UnwrapSDKContext(goCtx)
	var startTime time.Time
	if msg.StartTime != nil {
		startTime = time.Unix(msg.StartTime.Seconds, int64(msg.StartTime.Nanos))
	} else {
		startTime = ctx.BlockTime()
	}

	// Create the vesting schedule via keeper
	scheduleID, err := m.keeper.CreateVestingSchedule(
		goCtx,
		msg.BeneficiaryAddress,
		msg.TotalAmount,
		startTime,
		msg.CliffDuration,
		msg.VestingDuration,
		msg.VestingType,
		msg.ScheduleType,
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to create vesting schedule")
	}

	// Emit event (ctx already unwrapped above for startTime fallback)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vesting_schedule_created",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("beneficiary", msg.BeneficiaryAddress),
			sdk.NewAttribute("schedule_id", scheduleID),
			sdk.NewAttribute("total_amount", msg.TotalAmount),
			sdk.NewAttribute("vesting_type", msg.VestingType.String()),
			sdk.NewAttribute("schedule_type", msg.ScheduleType.String()),
		),
	)

	return &types.MsgCreateVestingScheduleResponse{
		ScheduleId: scheduleID,
	}, nil
}

// ReleaseVestedTokens releases vested tokens to beneficiary
func (m *MsgServer) ReleaseVestedTokens(
	goCtx context.Context,
	msg *types.MsgReleaseVestedTokens,
) (*types.MsgReleaseVestedTokensResponse, error) {
	// Validate message
	if msg == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "message cannot be nil")
	}

	// Validate beneficiary address (signer verification)
	if _, err := sdk.AccAddressFromBech32(msg.Beneficiary); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidBeneficiary, "invalid beneficiary address")
	}

	// Validate schedule ID
	if msg.ScheduleId == "" {
		return nil, errorsmod.Wrap(types.ErrInvalidScheduleID, "schedule ID cannot be empty")
	}

	// Release vested tokens via keeper
	amountReleased, err := m.keeper.ReleaseVestedTokens(goCtx, msg.Beneficiary, msg.ScheduleId)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to release vested tokens")
	}

	// Emit event
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vested_tokens_released",
			sdk.NewAttribute("beneficiary", msg.Beneficiary),
			sdk.NewAttribute("schedule_id", msg.ScheduleId),
			sdk.NewAttribute("amount_released", amountReleased),
		),
	)

	return &types.MsgReleaseVestedTokensResponse{
		AmountReleased: amountReleased,
	}, nil
}

// RevokeVestingSchedule revokes a vesting schedule
func (m *MsgServer) RevokeVestingSchedule(
	goCtx context.Context,
	msg *types.MsgRevokeVestingSchedule,
) (*types.MsgRevokeVestingScheduleResponse, error) {
	// Validate message
	if msg == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "message cannot be nil")
	}

	// Validate revoker address (must be module authority or authorized signer)
	revokerAddr, err := sdk.AccAddressFromBech32(msg.Revoker)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "invalid revoker address")
	}

	// Check if revoker has authority
	authority := m.keeper.GetAuthority()
	if msg.Revoker != authority {
		return nil, errorsmod.Wrapf(types.ErrUnauthorized,
			"only module authority can revoke vesting schedules: expected %s, got %s",
			authority, msg.Revoker)
	}

	// Validate schedule ID
	if msg.ScheduleId == "" {
		return nil, errorsmod.Wrap(types.ErrInvalidScheduleID, "schedule ID cannot be empty")
	}

	// Validate reason
	if msg.Reason == "" {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "revocation reason cannot be empty")
	}

	// Revoke the vesting schedule via keeper
	unvestedAmount, err := m.keeper.RevokeVestingSchedule(goCtx, msg.ScheduleId, msg.Reason)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to revoke vesting schedule")
	}

	// Emit event
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vesting_schedule_revoked",
			sdk.NewAttribute("revoker", msg.Revoker),
			sdk.NewAttribute("schedule_id", msg.ScheduleId),
			sdk.NewAttribute("reason", msg.Reason),
			sdk.NewAttribute("unvested_amount_returned", unvestedAmount),
		),
	)

	// Ensure valid address for logging
	_ = revokerAddr

	return &types.MsgRevokeVestingScheduleResponse{
		UnvestedAmountReturned: unvestedAmount,
	}, nil
}

// ============================
// GOVERNANCE HANDLERS
// ============================

// LockVotingTokens locks tokens for voting power boost
func (m *MsgServer) LockVotingTokens(
	goCtx context.Context,
	msg *types.MsgLockVotingTokens,
) (*types.MsgLockVotingTokensResponse, error) {
	// Validate message
	if msg == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "message cannot be nil")
	}

	// Validate owner address (signer verification)
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "invalid owner address")
	}

	// Validate amount
	if msg.Amount == "" || msg.Amount == "0" {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "amount must be positive")
	}

	// Validate lock duration
	if msg.LockDuration == 0 {
		return nil, errorsmod.Wrap(types.ErrInvalidLockDuration, "lock duration must be positive")
	}

	// Lock voting tokens via keeper
	lockID, votingPower, err := m.keeper.LockVotingTokens(
		goCtx,
		msg.Owner,
		msg.Amount,
		msg.LockDuration,
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to lock voting tokens")
	}

	// Emit event
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"voting_tokens_locked",
			sdk.NewAttribute("owner", msg.Owner),
			sdk.NewAttribute("lock_id", lockID),
			sdk.NewAttribute("amount", msg.Amount),
			sdk.NewAttribute("lock_duration", fmt.Sprintf("%d", msg.LockDuration)),
			sdk.NewAttribute("voting_power", votingPower),
		),
	)

	return &types.MsgLockVotingTokensResponse{
		LockId:      lockID,
		VotingPower: votingPower,
	}, nil
}

// UnlockVotingTokens unlocks voting tokens after lock period
func (m *MsgServer) UnlockVotingTokens(
	goCtx context.Context,
	msg *types.MsgUnlockVotingTokens,
) (*types.MsgUnlockVotingTokensResponse, error) {
	// Validate message
	if msg == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "message cannot be nil")
	}

	// Validate owner address (signer verification)
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "invalid owner address")
	}

	// Validate lock ID
	if msg.LockId == "" {
		return nil, errorsmod.Wrap(types.ErrVoteLockNotFound, "lock ID cannot be empty")
	}

	// Unlock voting tokens via keeper
	amountUnlocked, err := m.keeper.UnlockVotingTokens(goCtx, msg.Owner, msg.LockId)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to unlock voting tokens")
	}

	// Emit event
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"voting_tokens_unlocked",
			sdk.NewAttribute("owner", msg.Owner),
			sdk.NewAttribute("lock_id", msg.LockId),
			sdk.NewAttribute("amount_unlocked", amountUnlocked),
		),
	)

	return &types.MsgUnlockVotingTokensResponse{
		AmountUnlocked: amountUnlocked,
	}, nil
}

// ============================
// TREASURY MULTISIG HANDLERS
// ============================

// ProposeTreasurySpend proposes a treasury spend
func (m *MsgServer) ProposeTreasurySpend(
	goCtx context.Context,
	msg *types.MsgProposeTreasurySpend,
) (*types.MsgProposeTreasurySpendResponse, error) {
	// Validate message
	if msg == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "message cannot be nil")
	}

	// Validate proposer address
	if _, err := sdk.AccAddressFromBech32(msg.Proposer); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "invalid proposer address")
	}

	// Validate recipient address
	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "invalid recipient address")
	}

	// Validate amount
	if msg.Amount == "" || msg.Amount == "0" {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "amount must be positive")
	}

	// Validate description
	if msg.Description == "" {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "description cannot be empty")
	}

	// Propose treasury spend via keeper
	txID, executableAt, err := m.keeper.ProposeTreasurySpend(
		goCtx,
		msg.Proposer,
		msg.Recipient,
		msg.Amount,
		msg.Description,
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to propose treasury spend")
	}

	// Convert time.Time to gogoproto Timestamp for response
	executableAtProto := &gogotypes.Timestamp{
		Seconds: executableAt.Unix(),
		Nanos:   int32(executableAt.Nanosecond()),
	}

	// Emit event
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_spend_proposed",
			sdk.NewAttribute("proposer", msg.Proposer),
			sdk.NewAttribute("tx_id", txID),
			sdk.NewAttribute("recipient", msg.Recipient),
			sdk.NewAttribute("amount", msg.Amount),
			sdk.NewAttribute("description", msg.Description),
			sdk.NewAttribute("executable_at", fmt.Sprintf("%d", executableAt.Unix())),
		),
	)

	return &types.MsgProposeTreasurySpendResponse{
		TxId:         txID,
		ExecutableAt: executableAtProto,
	}, nil
}

// SignTreasurySpend signs a treasury spend proposal
func (m *MsgServer) SignTreasurySpend(
	goCtx context.Context,
	msg *types.MsgSignTreasurySpend,
) (*types.MsgSignTreasurySpendResponse, error) {
	// Validate message
	if msg == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "message cannot be nil")
	}

	// Validate signer address
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidSigner, "invalid signer address")
	}

	// Validate tx ID
	if msg.TxId == "" {
		return nil, errorsmod.Wrap(types.ErrTxNotFound, "transaction ID cannot be empty")
	}

	// Sign treasury spend via keeper
	currentSignatures, requiredThreshold, err := m.keeper.SignTreasurySpend(
		goCtx,
		msg.Signer,
		msg.TxId,
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to sign treasury spend")
	}

	// Emit event
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_spend_signed",
			sdk.NewAttribute("signer", msg.Signer),
			sdk.NewAttribute("tx_id", msg.TxId),
			sdk.NewAttribute("current_signatures", fmt.Sprintf("%d", currentSignatures)),
			sdk.NewAttribute("required_signatures", fmt.Sprintf("%d", requiredThreshold)),
		),
	)

	return &types.MsgSignTreasurySpendResponse{
		CurrentSignatures:  currentSignatures,
		RequiredSignatures: requiredThreshold,
	}, nil
}

// ExecuteTreasurySpend executes an approved treasury spend
func (m *MsgServer) ExecuteTreasurySpend(
	goCtx context.Context,
	msg *types.MsgExecuteTreasurySpend,
) (*types.MsgExecuteTreasurySpendResponse, error) {
	// Validate message
	if msg == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "message cannot be nil")
	}

	// Validate executor address
	if _, err := sdk.AccAddressFromBech32(msg.Executor); err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "invalid executor address")
	}

	// Validate tx ID
	if msg.TxId == "" {
		return nil, errorsmod.Wrap(types.ErrTxNotFound, "transaction ID cannot be empty")
	}

	// Get the pending transaction to retrieve recipient and amount for the event
	pendingTx, err := m.keeper.GetPendingTreasuryTx(goCtx, msg.TxId)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to retrieve treasury transaction")
	}

	// Execute treasury spend via keeper
	// Note: In a production system, you would need to integrate with bank keeper
	// to perform the actual token transfer. For now, we just update the state.
	err = m.keeper.ExecuteTreasurySpend(
		goCtx,
		msg.Executor,
		msg.TxId,
		"0", // Treasury balance check would be done by bank keeper integration
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to execute treasury spend")
	}

	// Emit event
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_spend_executed",
			sdk.NewAttribute("executor", msg.Executor),
			sdk.NewAttribute("tx_id", msg.TxId),
			sdk.NewAttribute("recipient", pendingTx.Recipient),
			sdk.NewAttribute("amount", pendingTx.Amount),
		),
	)

	return &types.MsgExecuteTreasurySpendResponse{
		Success: true,
	}, nil
}

// ============================
// GOVERNANCE PARAMETER HANDLERS
// ============================

// UpdateParams updates module parameters (governance only)
func (m *MsgServer) UpdateParams(
	goCtx context.Context,
	msg *types.MsgUpdateParams,
) (*types.MsgUpdateParamsResponse, error) {
	// Validate message
	if msg == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "message cannot be nil")
	}

	// Verify authority (must be governance module)
	authority := m.keeper.GetAuthority()
	if msg.Authority != authority {
		return nil, errorsmod.Wrapf(
			govtypes.ErrInvalidSigner,
			"invalid authority; expected %s, got %s",
			authority,
			msg.Authority,
		)
	}

	// Validate parameters
	if msg.Params == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "params cannot be nil")
	}

	// Validate governance configuration if present
	if msg.Params.Governance != nil {
		if err := m.keeper.ValidateGovernanceParameters(goCtx, msg.Params.Governance); err != nil {
			return nil, errorsmod.Wrap(err, "invalid governance parameters")
		}
	}

	// Update parameters via keeper
	if err := m.keeper.SetParams(*msg.Params); err != nil {
		return nil, errorsmod.Wrap(err, "failed to set parameters")
	}

	// Emit event
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeParamsUpdated,
			sdk.NewAttribute("authority", msg.Authority),
			sdk.NewAttribute("module", types.ModuleName),
		),
	)

	return &types.MsgUpdateParamsResponse{}, nil
}

// AdjustInflationRate manually adjusts inflation rate (governance only)
func (m *MsgServer) AdjustInflationRate(
	goCtx context.Context,
	msg *types.MsgAdjustInflationRate,
) (*types.MsgAdjustInflationRateResponse, error) {
	// Validate message
	if msg == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, "message cannot be nil")
	}

	// Verify authority (must be governance module)
	authority := m.keeper.GetAuthority()
	if msg.Authority != authority {
		return nil, errorsmod.Wrapf(
			govtypes.ErrInvalidSigner,
			"invalid authority; expected %s, got %s",
			authority,
			msg.Authority,
		)
	}

	// Validate reason
	if msg.Reason == "" {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "reason cannot be empty")
	}

	// Get current parameters
	params := m.keeper.GetParams()
	if params.Tokenomics == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidInflationRate, "tokenomics config not initialized")
	}

	// Validate new rate against min/max bounds
	if msg.NewRate < params.Tokenomics.MinInflationRate {
		return nil, errorsmod.Wrapf(
			types.ErrInflationRateTooLow,
			"new rate %d below minimum %d",
			msg.NewRate,
			params.Tokenomics.MinInflationRate,
		)
	}

	if msg.NewRate > params.Tokenomics.MaxInflationRate {
		return nil, errorsmod.Wrapf(
			types.ErrInflationRateTooHigh,
			"new rate %d exceeds maximum %d",
			msg.NewRate,
			params.Tokenomics.MaxInflationRate,
		)
	}

	// Store old rate for response and event
	oldRate := params.Tokenomics.InflationRate

	// Update inflation rate
	params.Tokenomics.InflationRate = msg.NewRate

	// Update parameters
	if err := m.keeper.SetParams(params); err != nil {
		return nil, errorsmod.Wrap(err, "failed to update inflation rate")
	}

	// Emit event
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"inflation_rate_adjusted",
			sdk.NewAttribute("authority", msg.Authority),
			sdk.NewAttribute("old_rate", fmt.Sprintf("%d", oldRate)),
			sdk.NewAttribute("new_rate", fmt.Sprintf("%d", msg.NewRate)),
			sdk.NewAttribute("reason", msg.Reason),
		),
	)

	return &types.MsgAdjustInflationRateResponse{
		OldRate: oldRate,
		NewRate: msg.NewRate,
	}, nil
}
