package types

const ModuleName = "identitychange"

type IdentityChangeStatus int32

const (
	IdentityChangeStatusUnspecified IdentityChangeStatus = iota
	IdentityChangeStatusIdle
	IdentityChangeStatusPendingVerification
	IdentityChangeStatusReadyToApply
	IdentityChangeStatusRejected
	IdentityChangeStatusApplied
	IdentityChangeStatusSuspended
)

type IdentityRecord struct {
	DID               string               `json:"did"`
	Owner             string               `json:"owner"`
	ConfidenceScore   int64                `json:"confidence_score"`
	MetadataHash      string               `json:"metadata_hash"`
	LatestIRVersion   string               `json:"latest_ir_version"`
	LastChangedHeight int64                `json:"last_changed_height"`
	Status            IdentityChangeStatus `json:"status"`
}

type IdentityChangeRequest struct {
	RequestID       string               `json:"request_id"`
	TargetDID       string               `json:"target_did"`
	Requester       string               `json:"requester"`
	Assistant       string               `json:"assistant"`
	IRID            string               `json:"ir_id"`
	ProofHash       string               `json:"proof_hash"`
	RequestMetaHash string               `json:"request_meta_hash"`
	Status          IdentityChangeStatus `json:"status"`
	Reason          string               `json:"reason"`
	CreatedHeight   int64                `json:"created_height"`
	VerdictHeight   int64                `json:"verdict_height"`
}

type IdentityChangeHistory struct {
	RequestID           string `json:"request_id"`
	TargetDID           string `json:"target_did"`
	PrevConfidenceScore int64  `json:"prev_confidence_score"`
	NewConfidenceScore  int64  `json:"new_confidence_score"`
	TransitionReason    string `json:"transition_reason"`
	ChangedHeight       int64  `json:"changed_height"`
}
