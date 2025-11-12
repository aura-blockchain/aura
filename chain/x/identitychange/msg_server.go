package identitychange

import (
	"context"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/types"
)

type msgServer struct {
	keeper *keeper.Keeper
}

func NewMsgServer(k *keeper.Keeper) MsgServer {
	return &msgServer{keeper: k}
}

func (s *msgServer) RequestIdentityChange(ctx context.Context, msg *MsgRequestIdentityChange) (*MsgRequestIdentityChangeResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}
	request := types.IdentityChangeRequest{
		RequestID:       fmt.Sprintf("%s-%d", msg.Requester, time.Now().UnixNano()),
		TargetDID:       msg.TargetDID,
		Requester:       msg.Requester,
		ProofHash:       msg.ProofHash,
		IRID:            msg.IRID,
		RequestMetaHash: msg.MetadataHash,
		CreatedHeight:   time.Now().Unix(),
		Status:          types.IdentityChangeStatusIdle,
	}
	created, err := s.keeper.CreateRequest(request)
	if err != nil {
		return nil, err
	}
	return &MsgRequestIdentityChangeResponse{RequestID: created.RequestID}, nil
}

func (s *msgServer) SubmitAssistantProof(ctx context.Context, msg *MsgSubmitAssistantProof) (*MsgSubmitAssistantProofResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}
	_, err := s.keeper.SubmitProof(msg.RequestID, msg.Assistant, msg.Success, msg.ConfidenceDelta, "")
	if err != nil {
		return nil, err
	}
	return &MsgSubmitAssistantProofResponse{}, nil
}

func (s *msgServer) ApplyIdentityChange(ctx context.Context, msg *MsgApplyIdentityChange) (*MsgApplyIdentityChangeResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}
	_, err := s.keeper.ApplyChange(msg.RequestID)
	if err != nil {
		return nil, err
	}
	return &MsgApplyIdentityChangeResponse{}, nil
}

func (s *msgServer) RejectIdentityChange(ctx context.Context, msg *MsgRejectIdentityChange) (*MsgRejectIdentityChangeResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}
	_, err := s.keeper.RejectChange(msg.RequestID, msg.Reason)
	if err != nil {
		return nil, err
	}
	return &MsgRejectIdentityChangeResponse{}, nil
}

func (s *msgServer) SuspendIdentityChanges(ctx context.Context, msg *MsgSuspendIdentityChanges) (*MsgSuspendIdentityChangesResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}
	s.keeper.SetSuspended(true)
	return &MsgSuspendIdentityChangesResponse{}, nil
}
