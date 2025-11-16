package economicsecurity

import (
	"context"

	"github.com/aequitas/aura/chain/x/economicsecurity/keeper"
	economicsecuritypb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
)

// MsgServer implements the economic security module's Msg service
type MsgServer struct {
	keeper *keeper.Keeper
	economicsecuritypb.UnimplementedMsgServer
}

// NewMsgServer creates a new MsgServer instance
func NewMsgServer(k *keeper.Keeper) economicsecuritypb.MsgServer {
	return &MsgServer{keeper: k}
}
func (s *MsgServer) CreateVestingSchedule(ctx context.Context, msg *economicsecuritypb.MsgCreateVestingSchedule) (*economicsecuritypb.MsgCreateVestingScheduleResponse, error) {
	scheduleID, err := s.keeper.CreateVestingSchedule(
		msg.BeneficiaryAddress,
		msg.TotalAmount,
		msg.StartTime,
		msg.CliffDuration,
		msg.VestingDuration,
		msg.VestingType,
		msg.ScheduleType,
	)
	if err != nil {
		return nil, err
	}

	return &economicsecuritypb.MsgCreateVestingScheduleResponse{
		ScheduleId: scheduleID,
	}, nil
}

// ReleaseVestedTokens releases vested tokens to beneficiary
func (s *MsgServer) ReleaseVestedTokens(ctx context.Context, msg *economicsecuritypb.MsgReleaseVestedTokens) (*economicsecuritypb.MsgReleaseVestedTokensResponse, error) {
	amount, err := s.keeper.ReleaseVestedTokens(msg.Beneficiary, msg.ScheduleId)
	if err != nil {
		return nil, err
	}

	return &economicsecuritypb.MsgReleaseVestedTokensResponse{
		AmountReleased: amount,
	}, nil
}

// RevokeVestingSchedule revokes a vesting schedule
func (s *MsgServer) RevokeVestingSchedule(ctx context.Context, msg *economicsecuritypb.MsgRevokeVestingSchedule) (*economicsecuritypb.MsgRevokeVestingScheduleResponse, error) {
	unvested, err := s.keeper.RevokeVestingSchedule(msg.ScheduleId, msg.Reason)
	if err != nil {
		return nil, err
	}

	return &economicsecuritypb.MsgRevokeVestingScheduleResponse{
		UnvestedAmountReturned: unvested,
	}, nil
}

// LockVotingTokens locks tokens for voting power boost
func (s *MsgServer) LockVotingTokens(ctx context.Context, msg *economicsecuritypb.MsgLockVotingTokens) (*economicsecuritypb.MsgLockVotingTokensResponse, error) {
	lockID, votingPower, err := s.keeper.LockVotingTokens(msg.Owner, msg.Amount, msg.LockDuration)
	if err != nil {
		return nil, err
	}

	return &economicsecuritypb.MsgLockVotingTokensResponse{
		LockId:      lockID,
		VotingPower: votingPower,
	}, nil
}

// UnlockVotingTokens unlocks voting tokens after lock period
func (s *MsgServer) UnlockVotingTokens(ctx context.Context, msg *economicsecuritypb.MsgUnlockVotingTokens) (*economicsecuritypb.MsgUnlockVotingTokensResponse, error) {
	amount, err := s.keeper.UnlockVotingTokens(msg.Owner, msg.LockId)
	if err != nil {
		return nil, err
	}

	return &economicsecuritypb.MsgUnlockVotingTokensResponse{
		AmountUnlocked: amount,
	}, nil
}

// ProposeTreasurySpend proposes a treasury spend
func (s *MsgServer) ProposeTreasurySpend(ctx context.Context, msg *economicsecuritypb.MsgProposeTreasurySpend) (*economicsecuritypb.MsgProposeTreasurySpendResponse, error) {
	txID, executableAt, err := s.keeper.ProposeTreasurySpend(msg.Proposer, msg.Recipient, msg.Amount, msg.Description)
	if err != nil {
		return nil, err
	}

	return &economicsecuritypb.MsgProposeTreasurySpendResponse{
		TxId:         txID,
		ExecutableAt: executableAt,
	}, nil
}

// SignTreasurySpend signs a treasury spend proposal
func (s *MsgServer) SignTreasurySpend(ctx context.Context, msg *economicsecuritypb.MsgSignTreasurySpend) (*economicsecuritypb.MsgSignTreasurySpendResponse, error) {
	currentSigs, requiredSigs, err := s.keeper.SignTreasurySpend(msg.Signer, msg.TxId)
	if err != nil {
		return nil, err
	}

	return &economicsecuritypb.MsgSignTreasurySpendResponse{
		CurrentSignatures:  currentSigs,
		RequiredSignatures: requiredSigs,
	}, nil
}

// ExecuteTreasurySpend executes an approved treasury spend
func (s *MsgServer) ExecuteTreasurySpend(ctx context.Context, msg *economicsecuritypb.MsgExecuteTreasurySpend) (*economicsecuritypb.MsgExecuteTreasurySpendResponse, error) {
	// In a real implementation, would fetch treasury balance from bank module
	treasuryBalance := "1000000000000" // Placeholder

	err := s.keeper.ExecuteTreasurySpend(msg.Executor, msg.TxId, treasuryBalance)
	if err != nil {
		return nil, err
	}

	return &economicsecuritypb.MsgExecuteTreasurySpendResponse{
		Success: true,
	}, nil
}

// UpdateParams updates module parameters (governance only)
func (s *MsgServer) UpdateParams(ctx context.Context, msg *economicsecuritypb.MsgUpdateParams) (*economicsecuritypb.MsgUpdateParamsResponse, error) {
	// In a real implementation, would verify msg.Authority is governance module
	err := s.keeper.SetParams(*msg.Params)
	if err != nil {
		return nil, err
	}

	return &economicsecuritypb.MsgUpdateParamsResponse{}, nil
}

// AdjustInflationRate manually adjusts inflation rate (governance only)
func (s *MsgServer) AdjustInflationRate(ctx context.Context, msg *economicsecuritypb.MsgAdjustInflationRate) (*economicsecuritypb.MsgAdjustInflationRateResponse, error) {
	// In a real implementation, would verify msg.Authority is governance module
	params := s.keeper.GetParams()
	oldRate := params.Tokenomics.InflationRate

	err := s.keeper.AdjustInflationRate(msg.NewRate, msg.Reason)
	if err != nil {
		return nil, err
	}

	return &economicsecuritypb.MsgAdjustInflationRateResponse{
		OldRate: oldRate,
		NewRate: msg.NewRate,
	}, nil
}
