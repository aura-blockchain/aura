package types

// Event types for the privacy module
const (
	EventTypeEncryptionKeyGenerated = "encryption_key_generated"
	EventTypeEncryptionKeyRotated   = "encryption_key_rotated"
	EventTypeMixingPoolCreated      = "mixing_pool_created"
	EventTypeMixingCompleted        = "mixing_completed"
	EventTypeRingSignatureCreated   = "ring_signature_created"
	EventTypeParamsUpdated          = "params_updated"
	EventTypePrivateTransaction     = "private_transaction"
	EventTypeMixingPool             = "mixing_pool"
	EventTypeViewKey                = "view_key"
	EventTypeNetworkPrivacy         = "network_privacy"
	EventTypeUpdateParams           = "update_params"
)

// Event attribute keys
const (
	AttributeKeyKeyID       = "key_id"
	AttributeKeyOwner       = "owner"
	AttributeKeyAlgorithm   = "algorithm"
	AttributeKeyPoolID      = "pool_id"
	AttributeKeyRingSize    = "ring_size"
	AttributeKeyBlockHeight = "block_height"
	AttributeKeyBlockTime   = "block_time"
	AttributeKeySender      = "sender"
	AttributeKeyTxHash      = "tx_hash"
)
