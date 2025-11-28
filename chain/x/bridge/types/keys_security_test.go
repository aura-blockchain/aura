package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

func TestMerkleRootKey(t *testing.T) {
	chainID := "ethereum"
	blockHeight := uint64(12345)
	key := types.MerkleRootKey(chainID, blockHeight)

	require.NotNil(t, key)
	require.Contains(t, string(key), chainID)
	require.Contains(t, string(key), "12345")
}

func TestMerkleRootKey_DifferentHeights(t *testing.T) {
	chainID := "ethereum"
	key1 := types.MerkleRootKey(chainID, 100)
	key2 := types.MerkleRootKey(chainID, 200)

	require.NotEqual(t, key1, key2)
}

func TestTSSNonceKey(t *testing.T) {
	key := types.TSSNonceKey()

	require.NotNil(t, key)
	require.Equal(t, types.TSSNoncePrefix, key)
}

func TestBridgeValidatorKey(t *testing.T) {
	address := "aura1validator"
	key := types.BridgeValidatorKey(address)

	require.NotNil(t, key)
	require.Contains(t, string(key), address)
	require.Equal(t, append(types.BridgeValidatorPrefix, []byte(address)...), key)
}

func TestBridgeValidatorKey_Empty(t *testing.T) {
	key := types.BridgeValidatorKey("")
	require.NotNil(t, key)
	require.Equal(t, types.BridgeValidatorPrefix, key)
}

func TestValidatorRotationKey(t *testing.T) {
	rotationID := "rotation-1"
	key := types.ValidatorRotationKey(rotationID)

	require.NotNil(t, key)
	require.Contains(t, string(key), rotationID)
	require.Equal(t, append(types.ValidatorRotationPrefix, []byte(rotationID)...), key)
}

func TestSlashingEventKey(t *testing.T) {
	eventID := "slash-event-1"
	key := types.SlashingEventKey(eventID)

	require.NotNil(t, key)
	require.Contains(t, string(key), eventID)
	require.Equal(t, append(types.SlashingEventPrefix, []byte(eventID)...), key)
}

func TestFraudProofKey(t *testing.T) {
	proofID := "fraud-proof-1"
	key := types.FraudProofKey(proofID)

	require.NotNil(t, key)
	require.Contains(t, string(key), proofID)
	require.Equal(t, append(types.FraudProofPrefix, []byte(proofID)...), key)
}

func TestTimeLockKey(t *testing.T) {
	lockID := "timelock-1"
	key := types.TimeLockKey(lockID)

	require.NotNil(t, key)
	require.Contains(t, string(key), lockID)
	require.Equal(t, append(types.TimeLockPrefix, []byte(lockID)...), key)
}

func TestWithdrawalLimitKey(t *testing.T) {
	address := "aura1user"
	key := types.WithdrawalLimitKey(address)

	require.NotNil(t, key)
	require.Contains(t, string(key), address)
	require.Equal(t, append(types.WithdrawalLimitPrefix, []byte(address)...), key)
}

func TestCircuitBreakerKey(t *testing.T) {
	key := types.CircuitBreakerKey()

	require.NotNil(t, key)
	require.Equal(t, types.CircuitBreakerPrefix, key)
}

func TestNonceKey(t *testing.T) {
	address := "aura1user"
	chainID := "ethereum"
	key := types.NonceKey(address, chainID)

	require.NotNil(t, key)
	require.Contains(t, string(key), address)
	require.Contains(t, string(key), chainID)
}

func TestNonceKey_DifferentChains(t *testing.T) {
	address := "aura1user"
	key1 := types.NonceKey(address, "ethereum")
	key2 := types.NonceKey(address, "polygon")

	require.NotEqual(t, key1, key2)
}

func TestNonceTrackerKey(t *testing.T) {
	address := "aura1user"
	chainID := "ethereum"
	key := types.NonceTrackerKey(address, chainID)

	require.NotNil(t, key)
	require.Contains(t, string(key), address)
	require.Contains(t, string(key), chainID)
}

func TestNonceTrackerKey_SameAsNonceKey(t *testing.T) {
	// NonceTrackerKey and NonceKey should produce the same result
	address := "aura1user"
	chainID := "ethereum"

	nonceKey := types.NonceKey(address, chainID)
	trackerKey := types.NonceTrackerKey(address, chainID)

	require.Equal(t, nonceKey, trackerKey)
}

func TestAddressPermissionKey(t *testing.T) {
	address := "aura1user"
	key := types.AddressPermissionKey(address)

	require.NotNil(t, key)
	require.Contains(t, string(key), address)
	require.Equal(t, append(types.AddressPermissionPrefix, []byte(address)...), key)
}

func TestBridgeFeeKey(t *testing.T) {
	feeType := bridgepb.FeeType_FEE_TRANSFER
	key := types.BridgeFeeKey(feeType)

	require.NotNil(t, key)
}

func TestBridgeFeeKey_DifferentTypes(t *testing.T) {
	key1 := types.BridgeFeeKey(bridgepb.FeeType_FEE_TRANSFER)
	key2 := types.BridgeFeeKey(bridgepb.FeeType_FEE_MINT_WRAPPED)
	key3 := types.BridgeFeeKey(bridgepb.FeeType_FEE_BURN_WRAPPED)

	require.NotEqual(t, key1, key2)
	require.NotEqual(t, key1, key3)
	require.NotEqual(t, key2, key3)
}

func TestInsuranceFundKey(t *testing.T) {
	key := types.InsuranceFundKey()

	require.NotNil(t, key)
	require.Equal(t, types.InsuranceFundPrefix, key)
}

func TestInsuranceClaimKey(t *testing.T) {
	claimID := "claim-1"
	key := types.InsuranceClaimKey(claimID)

	require.NotNil(t, key)
	require.Contains(t, string(key), claimID)
	require.Equal(t, append(types.InsuranceClaimPrefix, []byte(claimID)...), key)
}

func TestSecurityKeyUniqueness(t *testing.T) {
	// Ensure different security key types don't collide
	merkleKey := types.MerkleRootKey("eth", 1)
	tssKey := types.TSSNonceKey()
	validatorKey := types.BridgeValidatorKey("test")
	fraudKey := types.FraudProofKey("test")
	timelockKey := types.TimeLockKey("test")
	cbKey := types.CircuitBreakerKey()
	insuranceKey := types.InsuranceFundKey()

	keys := [][]byte{merkleKey, tssKey, validatorKey, fraudKey, timelockKey, cbKey, insuranceKey}

	// Check all keys are unique
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			require.NotEqual(t, keys[i], keys[j], "Keys at index %d and %d should be different", i, j)
		}
	}
}

func TestMerkleRootKey_ZeroHeight(t *testing.T) {
	chainID := "ethereum"
	key := types.MerkleRootKey(chainID, 0)

	require.NotNil(t, key)
	require.Contains(t, string(key), chainID)
	require.Contains(t, string(key), "-0")
}

func TestMerkleRootKey_LargeHeight(t *testing.T) {
	chainID := "ethereum"
	blockHeight := uint64(999999999999)
	key := types.MerkleRootKey(chainID, blockHeight)

	require.NotNil(t, key)
	require.Contains(t, string(key), chainID)
	require.Contains(t, string(key), "999999999999")
}

func TestNonceKey_EmptyInputs(t *testing.T) {
	key := types.NonceKey("", "")
	require.NotNil(t, key)
}

func TestAddressPermissionKey_LongAddress(t *testing.T) {
	address := "aura1abcdefghijklmnopqrstuvwxyz0123456789"
	key := types.AddressPermissionKey(address)

	require.NotNil(t, key)
	require.Contains(t, string(key), address)
}
