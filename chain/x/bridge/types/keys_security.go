package types

import (
	"fmt"

	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

// Additional KVStore key prefixes for security features
var (
	MerkleRootPrefix        = []byte{0x10}
	TSSNoncePrefix          = []byte{0x11}
	BridgeValidatorPrefix   = []byte{0x12}
	ValidatorRotationPrefix = []byte{0x13}
	SlashingEventPrefix     = []byte{0x14}
	FraudProofPrefix        = []byte{0x15}
	TimeLockPrefix          = []byte{0x16}
	WithdrawalLimitPrefix   = []byte{0x17}
	CircuitBreakerPrefix    = []byte{0x18}
	NonceTrackerPrefix      = []byte{0x19}
	AddressPermissionPrefix = []byte{0x1a}
	BridgeFeePrefix         = []byte{0x1b}
	InsuranceFundPrefix     = []byte{0x1c}
	InsuranceClaimPrefix    = []byte{0x1d}
	ValidatorSigningPrefix  = []byte{0x1e} // For liveness tracking
)

// MerkleRootKey returns the store key for a Merkle root
func MerkleRootKey(chainId string, blockHeight uint64) []byte {
	key := append(MerkleRootPrefix, []byte(chainId)...)
	key = append(key, []byte(fmt.Sprintf("-%d", blockHeight))...)
	return key
}

// TSSNonceKey returns the store key for the TSS nonce
func TSSNonceKey() []byte {
	return TSSNoncePrefix
}

// BridgeValidatorKey returns the store key for a bridge validator
func BridgeValidatorKey(address string) []byte {
	return append(BridgeValidatorPrefix, []byte(address)...)
}

// ValidatorRotationKey returns the store key for a validator rotation
func ValidatorRotationKey(rotationId string) []byte {
	return append(ValidatorRotationPrefix, []byte(rotationId)...)
}

// SlashingEventKey returns the store key for a slashing event
func SlashingEventKey(eventId string) []byte {
	return append(SlashingEventPrefix, []byte(eventId)...)
}

// FraudProofKey returns the store key for a fraud proof
func FraudProofKey(proofId string) []byte {
	return append(FraudProofPrefix, []byte(proofId)...)
}

// TimeLockKey returns the store key for a time lock
func TimeLockKey(lockId string) []byte {
	return append(TimeLockPrefix, []byte(lockId)...)
}

// WithdrawalLimitKey returns the store key for a withdrawal limit
func WithdrawalLimitKey(address string) []byte {
	return append(WithdrawalLimitPrefix, []byte(address)...)
}

// CircuitBreakerKey returns the store key for the circuit breaker
func CircuitBreakerKey() []byte {
	return CircuitBreakerPrefix
}

// NonceKey returns the store key for a nonce
func NonceKey(address string, chainId string) []byte {
	key := append([]byte(address), []byte("-")...)
	key = append(key, []byte(chainId)...)
	return append(NonceTrackerPrefix, key...)
}

// NonceTrackerKey returns the store key for a nonce tracker
func NonceTrackerKey(address string, chainId string) []byte {
	key := append([]byte(address), []byte("-")...)
	key = append(key, []byte(chainId)...)
	return append(NonceTrackerPrefix, key...)
}

// AddressPermissionKey returns the store key for an address permission
func AddressPermissionKey(address string) []byte {
	return append(AddressPermissionPrefix, []byte(address)...)
}

// BridgeFeeKey returns the store key for a bridge fee
func BridgeFeeKey(feeType bridgepb.FeeType) []byte {
	return append(BridgeFeePrefix, []byte(fmt.Sprintf("%d", feeType))...)
}

// InsuranceFundKey returns the store key for the insurance fund
func InsuranceFundKey() []byte {
	return InsuranceFundPrefix
}

// InsuranceClaimKey returns the store key for an insurance claim
func InsuranceClaimKey(claimId string) []byte {
	return append(InsuranceClaimPrefix, []byte(claimId)...)
}

// ValidatorSigningInfoKey returns the store key for validator signing info at a specific block height.
// Used for liveness tracking to determine if validators should be slashed for downtime.
//
// Storage format: ValidatorSigningPrefix + validatorAddress + "-" + blockHeight
//
// Parameters:
//   - validatorAddress: Address of the validator
//   - blockHeight: Block height
//
// Returns:
//   - []byte: Store key
func ValidatorSigningInfoKey(validatorAddress string, blockHeight int64) []byte {
	key := append(ValidatorSigningPrefix, []byte(validatorAddress)...)
	key = append(key, []byte(fmt.Sprintf("-%d", blockHeight))...)
	return key
}
