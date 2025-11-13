package identitychange

import (
	"github.com/aequitas/aura/chain/x/identitychange/types"
	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
	query "github.com/cosmos/cosmos-sdk/types/query"
)

func requestFromProto(msg *identitychangepb.MsgRequestIdentityChange) types.IdentityChangeRequest {
	if msg == nil {
		return types.IdentityChangeRequest{}
	}
	return types.IdentityChangeRequest{
		TargetDID:       msg.TargetDid,
		Requester:       msg.Requester,
		IRID:            msg.IrId,
		ProofHash:       msg.ProofHash,
		RequestMetaHash: msg.MetadataHash,
	}
}

func requestToProto(req types.IdentityChangeRequest) *identitychangepb.IdentityChangeRequest {
	return &identitychangepb.IdentityChangeRequest{
		RequestId:       req.RequestID,
		TargetDid:       req.TargetDID,
		Requester:       req.Requester,
		Assistant:       req.Assistant,
		IrId:            req.IRID,
		ProofHash:       req.ProofHash,
		RequestMetaHash: req.RequestMetaHash,
		Status:          toProtoStatus(req.Status),
		Reason:          req.Reason,
		CreatedHeight:   req.CreatedHeight,
		VerdictHeight:   req.VerdictHeight,
	}
}

func recordToProto(record types.IdentityRecord) *identitychangepb.IdentityRecord {
	return &identitychangepb.IdentityRecord{
		Did:               record.DID,
		Owner:             record.Owner,
		ConfidenceScore:   record.ConfidenceScore,
		MetadataHash:      record.MetadataHash,
		LatestIrVersion:   record.LatestIRVersion,
		LastChangedHeight: record.LastChangedHeight,
		Status:            toProtoStatus(record.Status),
	}
}

func historyToProto(history types.IdentityChangeHistory) *identitychangepb.IdentityChangeHistory {
	return &identitychangepb.IdentityChangeHistory{
		RequestId:           history.RequestID,
		TargetDid:           history.TargetDID,
		PrevConfidenceScore: history.PrevConfidenceScore,
		NewConfidenceScore:  history.NewConfidenceScore,
		TransitionReason:    history.TransitionReason,
		ChangedHeight:       history.ChangedHeight,
	}
}

func historySliceToProto(entries []types.IdentityChangeHistory) []*identitychangepb.IdentityChangeHistory {
	result := make([]*identitychangepb.IdentityChangeHistory, 0, len(entries))
	for _, entry := range entries {
		copy := entry
		result = append(result, historyToProto(copy))
	}
	return result
}

func toProtoStatus(status types.IdentityChangeStatus) identitychangepb.IdentityChangeStatus {
	converted := identitychangepb.IdentityChangeStatus(status)
	if converted < identitychangepb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_UNSPECIFIED || converted > identitychangepb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_SUSPENDED {
		return identitychangepb.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_UNSPECIFIED
	}
	return converted
}

func paginateHistory(entries []types.IdentityChangeHistory, page *query.PageRequest) ([]types.IdentityChangeHistory, *query.PageResponse) {
	offset := 0
	limit := len(entries)
	if page != nil {
		offset = int(page.Offset)
		if page.Limit > 0 {
			limit = int(page.Limit)
		}
	}
	if offset < 0 {
		offset = 0
	}
	if offset > len(entries) {
		offset = len(entries)
	}
	end := offset + limit
	if end > len(entries) {
		end = len(entries)
	}
	return entries[offset:end], &query.PageResponse{
		Total: uint64(len(entries)),
	}
}
