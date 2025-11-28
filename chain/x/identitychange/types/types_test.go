package types

import (
	"testing"
)

func TestTypeExports(t *testing.T) {
	var _ IdentityChangeStatus
	var _ IdentityRecord
	var _ IdentityChangeRequest
	var _ IdentityChangeHistory
	var _ Params
	var _ GenesisState
}

func TestMessageTypeExports(t *testing.T) {
	var _ MsgRequestIdentityChange
	var _ MsgSubmitAssistantProof
	var _ MsgApplyIdentityChange
	var _ MsgRejectIdentityChange
	var _ MsgSuspendIdentityChanges
}

func TestQueryTypeExports(t *testing.T) {
	var _ QueryIdentityRecordRequest
	var _ QueryIdentityRecordResponse
	var _ QueryIdentityChangeRequestRequest
	var _ QueryIdentityChangeRequestResponse
	var _ QueryIdentityChangeHistoryRequest
	var _ QueryIdentityChangeHistoryResponse
}

func TestIdentityChangeStatusEnums(t *testing.T) {
	statuses := []IdentityChangeStatus{
		IdentityChangeStatus_IDENTITY_CHANGE_STATUS_UNSPECIFIED,
		IdentityChangeStatus_IDENTITY_CHANGE_STATUS_IDLE,
		IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
		IdentityChangeStatus_IDENTITY_CHANGE_STATUS_READY_TO_APPLY,
		IdentityChangeStatus_IDENTITY_CHANGE_STATUS_REJECTED,
		IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED,
		IdentityChangeStatus_IDENTITY_CHANGE_STATUS_SUSPENDED,
	}

	seen := make(map[IdentityChangeStatus]bool)
	for _, status := range statuses {
		if seen[status] {
			t.Errorf("duplicate IdentityChangeStatus value: %v", status)
		}
		seen[status] = true
	}

	if len(seen) != len(statuses) {
		t.Error("not all IdentityChangeStatus values are unique")
	}
}
