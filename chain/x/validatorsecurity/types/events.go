package types

// Event types for the validatorsecurity module
const (
	EventTypeValidatorJailed   = "validator_jailed"
	EventTypeValidatorUnjailed = "validator_unjailed"
	EventTypeValidatorSlashed  = "validator_slashed"
	EventTypeSentryNodeAdded   = "sentry_node_added"
	EventTypeSentryNodeRemoved = "sentry_node_removed"
	EventTypeParamsUpdated     = "params_updated"
)

// Event attribute keys
const (
	AttributeKeyValidatorAddress = "validator_address"
	AttributeKeyReason           = "reason"
	AttributeKeySlashAmount      = "slash_amount"
	AttributeKeySlashFraction    = "slash_fraction"
	AttributeKeyJailDuration     = "jail_duration_seconds"
	AttributeKeyNodeID           = "node_id"
	AttributeKeyBlockHeight      = "block_height"
	AttributeKeyBlockTime        = "block_time"
)
