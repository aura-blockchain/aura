package types

import "fmt"

// Event types for the identitychange module
const (
	EventTypeIdentityChangeRequested = "identity_change_requested"
	EventTypeIdentityChangeApproved  = "identity_change_approved"
	EventTypeIdentityChangeRejected  = "identity_change_rejected"
	EventTypeIdentityChangeExecuted  = "identity_change_executed"
	EventTypeIdentityChangeExpired   = "identity_change_expired"
	EventTypeParamsUpdated           = "params_updated"
)

// Event attribute keys
const (
	AttributeKeyRequestID    = "request_id"
	AttributeKeyRequester    = "requester"
	AttributeKeyOldIdentity  = "old_identity"
	AttributeKeyNewIdentity  = "new_identity"
	AttributeKeyChangeType   = "change_type"
	AttributeKeyApprovedBy   = "approved_by"
	AttributeKeyReason       = "reason"
	AttributeKeyBlockHeight  = "block_height"
	AttributeKeyBlockTime    = "block_time"
)

func formatInt64(i int64) string { return fmt.Sprintf("%d", i) }
