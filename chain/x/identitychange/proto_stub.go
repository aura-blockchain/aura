package identitychange

import (
	"context"

	"github.com/aequitas/aura/chain/x/identitychange/types"
)

type MsgRequestIdentityChange struct {
	Requester    string
	TargetDID    string
	MetadataHash string
	IRID         string
	ProofHash    string
}

type MsgRequestIdentityChangeResponse struct {
	RequestID string
}

type MsgSubmitAssistantProof struct {
	Assistant       string
	RequestID       string
	ProofHash       string
	ConfidenceDelta int64
	Success         bool
}

type MsgSubmitAssistantProofResponse struct{}

type MsgApplyIdentityChange struct {
	Requester string
	RequestID string
}

type MsgApplyIdentityChangeResponse struct{}

type MsgRejectIdentityChange struct {
	Actor     string
	RequestID string
	Reason    string
}

type MsgRejectIdentityChangeResponse struct{}

type MsgSuspendIdentityChanges struct {
	Authority string
	Reason    string
}

type MsgSuspendIdentityChangesResponse struct{}

type MsgServer interface {
	RequestIdentityChange(context.Context, *MsgRequestIdentityChange) (*MsgRequestIdentityChangeResponse, error)
	SubmitAssistantProof(context.Context, *MsgSubmitAssistantProof) (*MsgSubmitAssistantProofResponse, error)
	ApplyIdentityChange(context.Context, *MsgApplyIdentityChange) (*MsgApplyIdentityChangeResponse, error)
	RejectIdentityChange(context.Context, *MsgRejectIdentityChange) (*MsgRejectIdentityChangeResponse, error)
	SuspendIdentityChanges(context.Context, *MsgSuspendIdentityChanges) (*MsgSuspendIdentityChangesResponse, error)
}

type QueryIdentityRecordRequest struct {
	DID string
}

type QueryIdentityRecordResponse struct {
	Record *types.IdentityRecord
}

type QueryIdentityChangeRequestRequest struct {
	RequestID string
}

type QueryIdentityChangeRequestResponse struct {
	Request *types.IdentityChangeRequest
}

type QueryIdentityChangeHistoryRequest struct {
	DID string
}

type QueryIdentityChangeHistoryResponse struct {
	Entries []*types.IdentityChangeHistory
}

type QueryServer interface {
	IdentityRecord(context.Context, *QueryIdentityRecordRequest) (*QueryIdentityRecordResponse, error)
	IdentityChangeRequest(context.Context, *QueryIdentityChangeRequestRequest) (*QueryIdentityChangeRequestResponse, error)
	IdentityChangeHistory(context.Context, *QueryIdentityChangeHistoryRequest) (*QueryIdentityChangeHistoryResponse, error)
}
