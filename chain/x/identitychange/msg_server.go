package identitychange

import (
	"context"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/types"
	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
)

type msgServer struct {
	identitychangepb.UnimplementedMsgServer
	keeper *keeper.Keeper
}

func NewMsgServer(k *keeper.Keeper) identitychangepb.MsgServer {
	return &msgServer{keeper: k}
}

func (s *msgServer) RequestIdentityChange(ctx context.Context, msg *identitychangepb.MsgRequestIdentityChange) (*identitychangepb.MsgRequestIdentityChangeResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}
	request := requestFromProto(msg)
	request.RequestID = fmt.Sprintf("%s-%d", msg.Requester, time.Now().UnixNano())
	request.CreatedHeight = time.Now().Unix()
	request.Status = types.IdentityChangeStatusPendingVerification
	created, err := s.keeper.CreateRequest(request)
	if err != nil {
		return nil, err
	}
	return &identitychangepb.MsgRequestIdentityChangeResponse{RequestId: created.RequestID}, nil
}

func (s *msgServer) SubmitAssistantProof(ctx context.Context, msg *identitychangepb.MsgSubmitAssistantProof) (*identitychangepb.MsgSubmitAssistantProofResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}
	_, err := s.keeper.SubmitProof(msg.RequestId, msg.Assistant, msg.Success, msg.ConfidenceDelta, "")
	if err != nil {
		return nil, err
	}
	return &identitychangepb.MsgSubmitAssistantProofResponse{}, nil
}

func (s *msgServer) ApplyIdentityChange(ctx context.Context, msg *identitychangepb.MsgApplyIdentityChange) (*identitychangepb.MsgApplyIdentityChangeResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}
	_, err := s.keeper.ApplyChange(msg.RequestId)
	if err != nil {
		return nil, err
	}
	return &identitychangepb.MsgApplyIdentityChangeResponse{}, nil
}

func (s *msgServer) RejectIdentityChange(ctx context.Context, msg *identitychangepb.MsgRejectIdentityChange) (*identitychangepb.MsgRejectIdentityChangeResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}
	_, err := s.keeper.RejectChange(msg.RequestId, msg.Reason)
	if err != nil {
		return nil, err
	}
	return &identitychangepb.MsgRejectIdentityChangeResponse{}, nil
}

func (s *msgServer) SuspendIdentityChanges(ctx context.Context, msg *identitychangepb.MsgSuspendIdentityChanges) (*identitychangepb.MsgSuspendIdentityChangesResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}
	s.keeper.SetSuspended(true)
	return &identitychangepb.MsgSuspendIdentityChangesResponse{}, nil
}
