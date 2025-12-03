package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

type InvariantsTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

func (suite *InvariantsTestSuite) TestAllInvariants() {
	ctx := suite.SdkCtx

	// Test: All invariants on empty store
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Create a mock invariant registry
	registry := sdk.NewInvariantRegistry()

	// Register invariants - should not panic
	suite.NotPanics(func() {
		RegisterInvariants(registry, suite.Keeper)
	})
}

func (suite *InvariantsTestSuite) TestParamsInvariant() {
	ctx := suite.SdkCtx

	// Test: valid params pass
	inv := ParamsInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "valid params should pass")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestKeyRotationValidityInvariant() {
	ctx := suite.SdkCtx

	// Test: empty store passes
	inv := KeyRotationValidityInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "empty store should pass")
	suite.Empty(msg)

	// Test: valid key rotation schedule passes
	validSchedule := &cryptoproto.KeyRotationSchedule{
		KeyId:                 "key-1",
		RotationPeriodSeconds: 3600,
		LastRotationTime:      timestamppb.Now(),
		NextRotationTime:      timestamppb.Now(),
	}
	suite.storeKeyRotationSchedule(ctx, validSchedule)

	msg, broken = inv(ctx)
	suite.False(broken, "valid key rotation schedule should pass")
	suite.Empty(msg)

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: empty key ID fails
	emptyKeySchedule := &cryptoproto.KeyRotationSchedule{
		KeyId:                 "",
		RotationPeriodSeconds: 3600,
		LastRotationTime:      timestamppb.Now(),
		NextRotationTime:      timestamppb.Now(),
	}
	suite.storeKeyRotationSchedule(ctx, emptyKeySchedule)

	msg, broken = inv(ctx)
	suite.True(broken, "empty key ID should break invariant")
	suite.Contains(msg, "empty key ID")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: zero rotation period fails
	zeroRotationSchedule := &cryptoproto.KeyRotationSchedule{
		KeyId:                 "key-1",
		RotationPeriodSeconds: 0,
		LastRotationTime:      timestamppb.Now(),
		NextRotationTime:      timestamppb.Now(),
	}
	suite.storeKeyRotationSchedule(ctx, zeroRotationSchedule)

	msg, broken = inv(ctx)
	suite.True(broken, "zero rotation period should break invariant")
	suite.Contains(msg, "zero rotation period")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: nil next rotation time fails
	nilNextRotationSchedule := &cryptoproto.KeyRotationSchedule{
		KeyId:                 "key-1",
		RotationPeriodSeconds: 3600,
		LastRotationTime:      timestamppb.Now(),
		NextRotationTime:      nil,
	}
	suite.storeKeyRotationSchedule(ctx, nilNextRotationSchedule)

	msg, broken = inv(ctx)
	suite.True(broken, "nil next rotation time should break invariant")
	suite.Contains(msg, "nil next rotation time")
}

func (suite *InvariantsTestSuite) TestThresholdSchemeConsistencyInvariant() {
	ctx := suite.SdkCtx

	// Test: empty store passes
	inv := ThresholdSchemeConsistencyInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "empty store should pass")
	suite.Empty(msg)

	// Test: valid threshold scheme passes
	validScheme := &cryptoproto.ThresholdSignatureScheme{
		SchemeId:          "scheme-1",
		Threshold:         2,
		TotalParticipants: 3,
		PublicKey:         []byte("public-key"),
	}
	suite.storeThresholdScheme(ctx, validScheme)

	msg, broken = inv(ctx)
	suite.False(broken, "valid threshold scheme should pass")
	suite.Empty(msg)

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: empty scheme ID fails
	emptyIDScheme := &cryptoproto.ThresholdSignatureScheme{
		SchemeId:          "",
		Threshold:         2,
		TotalParticipants: 3,
		PublicKey:         []byte("public-key"),
	}
	suite.storeThresholdScheme(ctx, emptyIDScheme)

	msg, broken = inv(ctx)
	suite.True(broken, "empty scheme ID should break invariant")
	suite.Contains(msg, "empty ID")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: threshold > total participants fails
	invalidThresholdScheme := &cryptoproto.ThresholdSignatureScheme{
		SchemeId:          "scheme-1",
		Threshold:         5,
		TotalParticipants: 3,
		PublicKey:         []byte("public-key"),
	}
	suite.storeThresholdScheme(ctx, invalidThresholdScheme)

	msg, broken = inv(ctx)
	suite.True(broken, "threshold > total participants should break invariant")
	suite.Contains(msg, "invalid threshold")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: zero threshold fails
	zeroThresholdScheme := &cryptoproto.ThresholdSignatureScheme{
		SchemeId:          "scheme-1",
		Threshold:         0,
		TotalParticipants: 3,
		PublicKey:         []byte("public-key"),
	}
	suite.storeThresholdScheme(ctx, zeroThresholdScheme)

	msg, broken = inv(ctx)
	suite.True(broken, "zero threshold should break invariant")
	suite.Contains(msg, "invalid threshold")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: zero participants fails
	zeroParticipantsScheme := &cryptoproto.ThresholdSignatureScheme{
		SchemeId:          "scheme-1",
		Threshold:         1,
		TotalParticipants: 0,
		PublicKey:         []byte("public-key"),
	}
	suite.storeThresholdScheme(ctx, zeroParticipantsScheme)

	msg, broken = inv(ctx)
	suite.True(broken, "zero participants should break invariant")
	suite.Contains(msg, "zero participants")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: empty public key fails
	emptyPubKeyScheme := &cryptoproto.ThresholdSignatureScheme{
		SchemeId:          "scheme-1",
		Threshold:         2,
		TotalParticipants: 3,
		PublicKey:         []byte{},
	}
	suite.storeThresholdScheme(ctx, emptyPubKeyScheme)

	msg, broken = inv(ctx)
	suite.True(broken, "empty public key should break invariant")
	suite.Contains(msg, "empty public key")
}

func (suite *InvariantsTestSuite) TestZKProofConfigValidityInvariant() {
	ctx := suite.SdkCtx

	// Test: empty store passes
	inv := ZKProofConfigValidityInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "empty store should pass")
	suite.Empty(msg)

	// Test: valid ZK proof config passes
	validConfig := &cryptoproto.ZKProofConfig{
		ConfigId:           "config-1",
		ProofType:          "groth16",
		CircuitParameters:  []byte("params"),
		VerificationKey:    []byte("vkey"),
	}
	suite.storeZKProofConfig(ctx, validConfig)

	msg, broken = inv(ctx)
	suite.False(broken, "valid ZK proof config should pass")
	suite.Empty(msg)

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: empty config ID fails
	emptyIDConfig := &cryptoproto.ZKProofConfig{
		ConfigId:           "",
		ProofType:          "groth16",
		CircuitParameters:  []byte("params"),
		VerificationKey:    []byte("vkey"),
	}
	suite.storeZKProofConfig(ctx, emptyIDConfig)

	msg, broken = inv(ctx)
	suite.True(broken, "empty config ID should break invariant")
	suite.Contains(msg, "empty ID")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: invalid proof type fails
	invalidTypeConfig := &cryptoproto.ZKProofConfig{
		ConfigId:           "config-1",
		ProofType:          "invalid-type",
		CircuitParameters:  []byte("params"),
		VerificationKey:    []byte("vkey"),
	}
	suite.storeZKProofConfig(ctx, invalidTypeConfig)

	msg, broken = inv(ctx)
	suite.True(broken, "invalid proof type should break invariant")
	suite.Contains(msg, "invalid proof type")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: empty circuit parameters fails
	emptyParamsConfig := &cryptoproto.ZKProofConfig{
		ConfigId:           "config-1",
		ProofType:          "groth16",
		CircuitParameters:  []byte{},
		VerificationKey:    []byte("vkey"),
	}
	suite.storeZKProofConfig(ctx, emptyParamsConfig)

	msg, broken = inv(ctx)
	suite.True(broken, "empty circuit parameters should break invariant")
	suite.Contains(msg, "empty circuit parameters")
}

func (suite *InvariantsTestSuite) TestSecureEnclaveValidityInvariant() {
	ctx := suite.SdkCtx

	// Test: empty store passes
	inv := SecureEnclaveValidityInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "empty store should pass")
	suite.Empty(msg)

	// Test: valid secure enclave passes
	validEnclave := &cryptoproto.SecureEnclaveConfig{
		EnclaveId:   "enclave-1",
		EnclaveType: "sgx",
		Attestation: []byte("attestation"),
	}
	suite.storeSecureEnclave(ctx, validEnclave)

	msg, broken = inv(ctx)
	suite.False(broken, "valid secure enclave should pass")
	suite.Empty(msg)

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: empty enclave ID fails
	emptyIDEnclave := &cryptoproto.SecureEnclaveConfig{
		EnclaveId:   "",
		EnclaveType: "sgx",
		Attestation: []byte("attestation"),
	}
	suite.storeSecureEnclave(ctx, emptyIDEnclave)

	msg, broken = inv(ctx)
	suite.True(broken, "empty enclave ID should break invariant")
	suite.Contains(msg, "empty ID")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: empty attestation fails
	emptyAttestationEnclave := &cryptoproto.SecureEnclaveConfig{
		EnclaveId:   "enclave-1",
		EnclaveType: "sgx",
		Attestation: []byte{},
	}
	suite.storeSecureEnclave(ctx, emptyAttestationEnclave)

	msg, broken = inv(ctx)
	suite.True(broken, "empty attestation should break invariant")
	suite.Contains(msg, "empty attestation")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: invalid enclave type fails
	invalidTypeEnclave := &cryptoproto.SecureEnclaveConfig{
		EnclaveId:   "enclave-1",
		EnclaveType: "invalid-type",
		Attestation: []byte("attestation"),
	}
	suite.storeSecureEnclave(ctx, invalidTypeEnclave)

	msg, broken = inv(ctx)
	suite.True(broken, "invalid enclave type should break invariant")
	suite.Contains(msg, "invalid type")
}

func (suite *InvariantsTestSuite) TestQuantumKeyValidityInvariant() {
	ctx := suite.SdkCtx

	// Test: empty store passes
	inv := QuantumKeyValidityInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "empty store should pass")
	suite.Empty(msg)

	// Test: valid quantum key passes
	validKey := &cryptoproto.QuantumResistantKey{
		KeyId:     "key-1",
		Algorithm: "dilithium",
		PublicKey: []byte("public-key"),
	}
	suite.storeQuantumKey(ctx, validKey)

	msg, broken = inv(ctx)
	suite.False(broken, "valid quantum key should pass")
	suite.Empty(msg)

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: empty key ID fails
	emptyIDKey := &cryptoproto.QuantumResistantKey{
		KeyId:     "",
		Algorithm: "dilithium",
		PublicKey: []byte("public-key"),
	}
	suite.storeQuantumKey(ctx, emptyIDKey)

	msg, broken = inv(ctx)
	suite.True(broken, "empty key ID should break invariant")
	suite.Contains(msg, "empty ID")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: invalid algorithm fails
	invalidAlgKey := &cryptoproto.QuantumResistantKey{
		KeyId:     "key-1",
		Algorithm: "invalid-alg",
		PublicKey: []byte("public-key"),
	}
	suite.storeQuantumKey(ctx, invalidAlgKey)

	msg, broken = inv(ctx)
	suite.True(broken, "invalid algorithm should break invariant")
	suite.Contains(msg, "invalid algorithm")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: empty public key fails
	emptyPubKey := &cryptoproto.QuantumResistantKey{
		KeyId:     "key-1",
		Algorithm: "dilithium",
		PublicKey: []byte{},
	}
	suite.storeQuantumKey(ctx, emptyPubKey)

	msg, broken = inv(ctx)
	suite.True(broken, "empty public key should break invariant")
	suite.Contains(msg, "empty public key")
}

// Helper methods to store data directly to the store
func (suite *InvariantsTestSuite) storeKeyRotationSchedule(ctx sdk.Context, schedule *cryptoproto.KeyRotationSchedule) {
	store := ctx.KVStore(suite.Keeper.storeService.OpenKVStore(ctx))
	bz := suite.Keeper.cdc.MustMarshal(schedule)
	key := append(cryptoproto.KeyRotationScheduleKeyPrefix, []byte(schedule.KeyId)...)
	store.Set(key, bz)
}

func (suite *InvariantsTestSuite) storeThresholdScheme(ctx sdk.Context, scheme *cryptoproto.ThresholdSignatureScheme) {
	store := ctx.KVStore(suite.Keeper.storeService.OpenKVStore(ctx))
	bz := suite.Keeper.cdc.MustMarshal(scheme)
	key := append(cryptoproto.ThresholdSchemeKeyPrefix, []byte(scheme.SchemeId)...)
	store.Set(key, bz)
}

func (suite *InvariantsTestSuite) storeZKProofConfig(ctx sdk.Context, config *cryptoproto.ZKProofConfig) {
	store := ctx.KVStore(suite.Keeper.storeService.OpenKVStore(ctx))
	bz := suite.Keeper.cdc.MustMarshal(config)
	key := append(cryptoproto.ZKProofConfigKeyPrefix, []byte(config.ConfigId)...)
	store.Set(key, bz)
}

func (suite *InvariantsTestSuite) storeSecureEnclave(ctx sdk.Context, enclave *cryptoproto.SecureEnclaveConfig) {
	store := ctx.KVStore(suite.Keeper.storeService.OpenKVStore(ctx))
	bz := suite.Keeper.cdc.MustMarshal(enclave)
	key := append(cryptoproto.SecureEnclaveKeyPrefix, []byte(enclave.EnclaveId)...)
	store.Set(key, bz)
}

func (suite *InvariantsTestSuite) storeQuantumKey(ctx sdk.Context, qkey *cryptoproto.QuantumResistantKey) {
	store := ctx.KVStore(suite.Keeper.storeService.OpenKVStore(ctx))
	bz := suite.Keeper.cdc.MustMarshal(qkey)
	key := append(cryptoproto.QuantumKeyKeyPrefix, []byte(qkey.KeyId)...)
	store.Set(key, bz)
}
