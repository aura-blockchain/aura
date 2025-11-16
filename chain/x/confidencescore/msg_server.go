package confidencescore

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/confidencescore/keeper"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
)

type msgServer struct {
	confidencescorepb.UnimplementedMsgServer
	keeper *keeper.Keeper
}

// NewMsgServer creates a new message server
func NewMsgServer(k *keeper.Keeper) confidencescorepb.MsgServer {
	return &msgServer{keeper: k}
}

// RecordIRCompletion handles IR completion recording
func (s *msgServer) RecordIRCompletion(ctx context.Context, msg *confidencescorepb.MsgRecordIRCompletion) (*confidencescorepb.MsgRecordIRCompletionResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}

	// Validate inputs
	if msg.WalletAddress == "" {
		return nil, types.ErrInvalidWalletAddress
	}
	if msg.IrId == "" {
		return nil, types.ErrInvalidIRID
	}
	if len(msg.ProofHash) != 32 {
		return nil, types.ErrInvalidProofHash
	}

	// Record completion
	scoreEarned, err := s.keeper.RecordIRCompletion(
		msg.WalletAddress,
		msg.IrId,
		msg.AssistantAddress,
		msg.ProofHash,
		msg.VerifierHash,
		msg.Timestamp.AsTime().Unix(),
	)
	if err != nil {
		return nil, err
	}

	// Get updated user record
	record, ok := s.keeper.GetUserRecord(msg.WalletAddress)
	if !ok {
		return nil, types.ErrUserRecordNotFound
	}

	// Get the completion to extract multipliers
	completion, ok := s.keeper.GetIRCompletion(msg.WalletAddress, msg.IrId)
	if !ok {
		return nil, types.ErrCompletionNotFound
	}

	// Check if verification was achieved
	params := s.keeper.GetParams()
	verificationAchieved := record.TotalScore >= params.VerificationThreshold &&
		record.Status == types.VerificationStatusVerified

	return &confidencescorepb.MsgRecordIRCompletionResponse{
		ScoreEarned:          scoreEarned,
		NewTotalScore:        record.TotalScore,
		VerificationAchieved: verificationAchieved,
		VelocityMultiplier:   completion.VelocityBonus,
		ArenaMultiplier:      completion.ArenaBonus,
		JackpotMultiplier:    completion.JackpotBonus,
	}, nil
}

// RecalculateScore handles score recalculation (governance)
func (s *msgServer) RecalculateScore(ctx context.Context, msg *confidencescorepb.MsgRecalculateScore) (*confidencescorepb.MsgRecalculateScoreResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}

	// Verify authority is governance
	if msg.Authority != s.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", s.keeper.GetAuthority(), msg.Authority)
	}

	previousScore, recalculatedScore, discrepancies, err := s.keeper.RecalculateScore(msg.WalletAddress)
	if err != nil {
		return nil, err
	}

	return &confidencescorepb.MsgRecalculateScoreResponse{
		PreviousScore:     previousScore,
		RecalculatedScore: recalculatedScore,
		Discrepancies:     discrepancies,
	}, nil
}

// SlashScore handles score slashing (governance)
func (s *msgServer) SlashScore(ctx context.Context, msg *confidencescorepb.MsgSlashScore) (*confidencescorepb.MsgSlashScoreResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}

	// Verify authority is governance
	if msg.Authority != s.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", s.keeper.GetAuthority(), msg.Authority)
	}

	// Determine slash reason
	var reason types.SlashReason
	switch msg.Reason {
	case "fraud_detected":
		reason = types.SlashReasonFraudDetected
	case "false_attestation":
		reason = types.SlashReasonFalseAttestation
	case "collusion":
		reason = types.SlashReasonCollusion
	default:
		reason = types.SlashReasonFraudDetected
	}

	previousScore, newScore, verificationRevoked, slashTxHash, err := s.keeper.SlashScore(
		msg.WalletAddress,
		msg.IrId,
		msg.SlashAmount,
		reason,
		msg.Authority,
		msg.Evidence,
	)
	if err != nil {
		return nil, err
	}

	return &confidencescorepb.MsgSlashScoreResponse{
		PreviousScore:       previousScore,
		NewScore:            newScore,
		VerificationRevoked: verificationRevoked,
		SlashTxHash:         slashTxHash,
	}, nil
}

// AppealSlash handles slash appeals
func (s *msgServer) AppealSlash(ctx context.Context, msg *confidencescorepb.MsgAppealSlash) (*confidencescorepb.MsgAppealSlashResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}

	appealAccepted, _, err := s.keeper.AppealSlash(
		msg.WalletAddress,
		msg.SlashTxHash,
		msg.Evidence,
		msg.Deposit,
	)
	if err != nil {
		return nil, err
	}

	return &confidencescorepb.MsgAppealSlashResponse{
		AppealAccepted: appealAccepted,
		ReviewDeadline: nil, // TODO: Convert reviewDeadline to timestamp if needed
	}, nil
}

// ResolveAppeal handles appeal resolution (governance)
func (s *msgServer) ResolveAppeal(ctx context.Context, msg *confidencescorepb.MsgResolveAppeal) (*confidencescorepb.MsgResolveAppealResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}

	// Verify authority is governance
	if msg.Authority != s.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", s.keeper.GetAuthority(), msg.Authority)
	}

	restoredScore, depositReturned, err := s.keeper.ResolveAppeal(
		msg.WalletAddress,
		msg.SlashTxHash,
		msg.RestoreScore,
		msg.Authority,
		msg.ResolutionNotes,
	)
	if err != nil {
		return nil, err
	}

	return &confidencescorepb.MsgResolveAppealResponse{
		RestoredScore:   restoredScore,
		DepositReturned: depositReturned,
	}, nil
}
