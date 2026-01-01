// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TelemetryTestSuite struct {
	KeeperTestSuite
}

func TestTelemetryTestSuite(t *testing.T) {
	suite.Run(t, new(TelemetryTestSuite))
}

func (suite *TelemetryTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
}

// =============================================================================
// classifySignatureError Tests
// =============================================================================

func (suite *TelemetryTestSuite) TestClassifySignatureError_NilErrorNilPubKey() {
	result := classifySignatureError(nil, nil)
	suite.Equal("pubkey_recovery_failed", result)
}

func (suite *TelemetryTestSuite) TestClassifySignatureError_NilErrorWithPubKey() {
	result := classifySignatureError(nil, []byte("pubkey"))
	suite.Equal("address_mismatch", result)
}

func (suite *TelemetryTestSuite) TestClassifySignatureError_InvalidSignatureLength() {
	err := errors.New("invalid signature length: expected 65 bytes")
	result := classifySignatureError(err, nil)
	suite.Equal("invalid_signature_length", result)
}

func (suite *TelemetryTestSuite) TestClassifySignatureError_InvalidRecoveryID() {
	err := errors.New("invalid recovery id: must be 0-3")
	result := classifySignatureError(err, nil)
	suite.Equal("invalid_recovery_id", result)
}

func (suite *TelemetryTestSuite) TestClassifySignatureError_RecoveryFailed() {
	err := errors.New("recovery failed: could not recover public key")
	result := classifySignatureError(err, nil)
	suite.Equal("recovery_failed", result)
}

func (suite *TelemetryTestSuite) TestClassifySignatureError_VerificationFailed() {
	err := errors.New("ECDSA verification failed")
	result := classifySignatureError(err, nil)
	suite.Equal("ecdsa_verification_failed", result)
}

func (suite *TelemetryTestSuite) TestClassifySignatureError_MalformedSignature() {
	err := errors.New("malformed signature input")
	result := classifySignatureError(err, nil)
	suite.Equal("malformed_signature", result)
}

func (suite *TelemetryTestSuite) TestClassifySignatureError_OtherError() {
	err := errors.New("some unknown error")
	result := classifySignatureError(err, nil)
	suite.Equal("other", result)
}

// =============================================================================
// classifyStateError Tests
// =============================================================================

func (suite *TelemetryTestSuite) TestClassifyStateError_NilError() {
	result := classifyStateError(nil)
	suite.Equal("unknown", result)
}

func (suite *TelemetryTestSuite) TestClassifyStateError_UnmarshalError() {
	err := errors.New("failed to unmarshal proto message")
	result := classifyStateError(err)
	suite.Equal("unmarshal_error", result)
}

func (suite *TelemetryTestSuite) TestClassifyStateError_KeyNotFound() {
	err := errors.New("key not found in store")
	result := classifyStateError(err)
	suite.Equal("key_not_found", result)
}

func (suite *TelemetryTestSuite) TestClassifyStateError_DataCorrupted() {
	err := errors.New("data is corrupted")
	result := classifyStateError(err)
	suite.Equal("data_corrupted", result)
}

func (suite *TelemetryTestSuite) TestClassifyStateError_InvalidData() {
	err := errors.New("invalid data format")
	result := classifyStateError(err)
	suite.Equal("invalid_data", result)
}

func (suite *TelemetryTestSuite) TestClassifyStateError_DecodeError() {
	err := errors.New("failed to decode bytes")
	result := classifyStateError(err)
	suite.Equal("decode_error", result)
}

func (suite *TelemetryTestSuite) TestClassifyStateError_OtherError() {
	err := errors.New("some unknown state error")
	result := classifyStateError(err)
	suite.Equal("other", result)
}

// =============================================================================
// containsStr Tests
// =============================================================================

func (suite *TelemetryTestSuite) TestContainsStr_EmptySubstr() {
	result := containsStr("any string", "")
	suite.True(result)
}

func (suite *TelemetryTestSuite) TestContainsStr_Found() {
	result := containsStr("hello world", "world")
	suite.True(result)
}

func (suite *TelemetryTestSuite) TestContainsStr_NotFound() {
	result := containsStr("hello world", "foo")
	suite.False(result)
}

func (suite *TelemetryTestSuite) TestContainsStr_SubstrLongerThanStr() {
	result := containsStr("hi", "hello")
	suite.False(result)
}

func (suite *TelemetryTestSuite) TestContainsStr_ExactMatch() {
	result := containsStr("hello", "hello")
	suite.True(result)
}

func (suite *TelemetryTestSuite) TestContainsStr_AtStart() {
	result := containsStr("hello world", "hello")
	suite.True(result)
}

func (suite *TelemetryTestSuite) TestContainsStr_AtEnd() {
	result := containsStr("hello world", "world")
	suite.True(result)
}

func (suite *TelemetryTestSuite) TestContainsStr_CaseSensitive() {
	result := containsStr("Hello World", "hello")
	suite.False(result) // Case-sensitive
}

func (suite *TelemetryTestSuite) TestContainsStr_EmptyString() {
	result := containsStr("", "test")
	suite.False(result)
}

func (suite *TelemetryTestSuite) TestContainsStr_BothEmpty() {
	result := containsStr("", "")
	suite.True(result)
}

// =============================================================================
// Telemetry Recording Tests (these test that they don't panic)
// =============================================================================

func (suite *TelemetryTestSuite) TestRecordStateLoadError() {
	// Should not panic
	suite.Keeper.recordStateLoadError("bridge", "transfers", "key_not_found")
}

func (suite *TelemetryTestSuite) TestRecordUnmarshalError() {
	// Should not panic
	suite.Keeper.recordUnmarshalError("bridge", "CrossChainTransfer")
}

func (suite *TelemetryTestSuite) TestRecordStateCorruption() {
	// Should not panic
	suite.Keeper.recordStateCorruption("bridge", "validators")
}

func (suite *TelemetryTestSuite) TestRecordKVStoreIterationError() {
	// Should not panic
	suite.Keeper.recordKVStoreIterationError("transfers")
}

func (suite *TelemetryTestSuite) TestRecordBridgeTransfer() {
	// Should not panic
	suite.Keeper.recordBridgeTransfer("ethereum", "aura", "success", 0)
}

func (suite *TelemetryTestSuite) TestRecordIdentityLink() {
	// Should not panic
	suite.Keeper.recordIdentityLink("ethereum")
}

func (suite *TelemetryTestSuite) TestRecordIdentityLinkFailure() {
	// Should not panic
	suite.Keeper.recordIdentityLinkFailure("ethereum", "invalid_signature")
}

func (suite *TelemetryTestSuite) TestRecordInvalidRecoveryID() {
	// Should not panic
	suite.Keeper.recordInvalidRecoveryID("ethereum")
}

func (suite *TelemetryTestSuite) TestRecordPubKeyRecoveryFailure() {
	// Should not panic
	suite.Keeper.recordPubKeyRecoveryFailure("ethereum", "27")
}
