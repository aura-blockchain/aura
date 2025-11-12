// Code generated for placeholder purposes until buf/gogoproto generation is completed.
// Replace this stub with generated code after running `buf generate`.
package identitychange

import "context"

type IdentityChangeStatus int32

const (
	IdentityChangeStatusUnspecified         IdentityChangeStatus = 0
	IdentityChangeStatusIdle                IdentityChangeStatus = 1
	IdentityChangeStatusPendingVerification IdentityChangeStatus = 2
	IdentityChangeStatusReadyToApply        IdentityChangeStatus = 3
	IdentityChangeStatusRejected            IdentityChangeStatus = 4
	IdentityChangeStatusApplied             IdentityChangeStatus = 5
	IdentityChangeStatusSuspended           IdentityChangeStatus = 6
)

type IdentityRecord struct {
	DID               string
	Owner             string
	ConfidenceScore   int64
	MetadataHash      string
	LatestIRVersion   string
	LastChangedHeight int64
	Status            IdentityChangeStatus
}

type IdentityChangeRequest struct {
	RequestID       string
	TargetDID       string
	Requester       string
	Assistant       string
	IRID            string
	ProofHash       string
	RequestMetaHash string
	Status          IdentityChangeStatus
	Reason          string
	CreatedHeight   int64
	VerdictHeight   int64
}

type IdentityChangeHistory struct {
	RequestID           string
	TargetDID           string
	PrevConfidenceScore int64
	NewConfidenceScore  int64
	TransitionReason    string
	ChangedHeight       int64
}

type Params struct {
	MaxRequestsPerWalletPerMonth  int32
	MinConfidenceAfterChange      int64
	StalenessHeightThreshold      int64
	AssistantSlashOnFalsePositive bool
	StalenessInvestigatorChain    string
}

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
	Record *IdentityRecord
}

type QueryIdentityChangeRequestRequest struct {
	RequestID string
}

type QueryIdentityChangeRequestResponse struct {
	Request *IdentityChangeRequest
}

type PageRequest struct {
	Key        []byte
	Offset     uint64
	Limit      uint64
	CountTotal bool
}

type PageResponse struct {
	NextKey []byte
	Total   uint64
}

type QueryIdentityChangeHistoryRequest struct {
	DID        string
	Pagination *PageRequest
}

type QueryIdentityChangeHistoryResponse struct {
	Entries    []*IdentityChangeHistory
	Pagination *PageResponse
}

type QueryServer interface {
	IdentityRecord(context.Context, *QueryIdentityRecordRequest) (*QueryIdentityRecordResponse, error)
	IdentityChangeRequest(context.Context, *QueryIdentityChangeRequestRequest) (*QueryIdentityChangeRequestResponse, error)
	IdentityChangeHistory(context.Context, *QueryIdentityChangeHistoryRequest) (*QueryIdentityChangeHistoryResponse, error)
}
