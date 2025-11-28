package types

// Event types for the wasm module
const (
	EventTypeStoreCode         = "store_code"
	EventTypeInstantiate       = "instantiate"
	EventTypeExecute           = "execute"
	EventTypeMigrate           = "migrate"
	EventTypeUpdateAdmin       = "update_admin"
	EventTypeClearAdmin        = "clear_admin"
	EventTypeAuthorizeUploader = "authorize_uploader"
	EventTypeRevokeUploader    = "revoke_uploader"
	EventTypePauseContract     = "pause_contract"
	EventTypeUnpauseContract   = "unpause_contract"
	EventTypeUpdateParams      = "update_params"
)

// Event attribute keys
const (
	AttributeKeyCodeID    = "code_id"
	AttributeKeyContract  = "contract"
	AttributeKeySender    = "sender"
	AttributeKeyNewAdmin  = "new_admin"
	AttributeKeyUploader  = "uploader"
	AttributeKeyParams    = "params"
)
