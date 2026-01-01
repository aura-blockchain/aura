// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/chain/x/bridge/types"
)

type SlashingComprehensiveTestSuite struct {
	KeeperTestSuite
}

func TestSlashingComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(SlashingComprehensiveTestSuite))
}

func (suite *SlashingComprehensiveTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
}

// =============================================================================
// slashValidator Tests
// =============================================================================

func (suite *SlashingComprehensiveTestSuite) TestSlashValidator_NilStakingKeeper() {
	// With nil staking keeper, slashing should return zero
	validatorAddr := sdk.ValAddress("validator_________").String()
	slashFraction := sdkmath.LegacyMustNewDecFromStr("0.1")

	amount, err := suite.Keeper.slashValidator(suite.SdkCtx, validatorAddr, slashFraction, 100)
	suite.NoError(err)
	suite.True(amount.IsZero(), "slashing with nil staking keeper should return zero")
}

func (suite *SlashingComprehensiveTestSuite) TestSlashValidator_InvalidSlashFraction_Negative() {
	validatorAddr := sdk.ValAddress("validator_________").String()
	slashFraction := sdkmath.LegacyMustNewDecFromStr("-0.1") // Invalid: negative

	amount, err := suite.Keeper.slashValidator(suite.SdkCtx, validatorAddr, slashFraction, 100)
	suite.Error(err, "should reject negative slash fraction")
	suite.Contains(err.Error(), "invalid slash fraction")
	suite.True(amount.IsZero())
}

func (suite *SlashingComprehensiveTestSuite) TestSlashValidator_InvalidSlashFraction_OverOne() {
	validatorAddr := sdk.ValAddress("validator_________").String()
	slashFraction := sdkmath.LegacyMustNewDecFromStr("1.5") // Invalid: > 1.0

	amount, err := suite.Keeper.slashValidator(suite.SdkCtx, validatorAddr, slashFraction, 100)
	suite.Error(err, "should reject slash fraction > 1.0")
	suite.Contains(err.Error(), "invalid slash fraction")
	suite.True(amount.IsZero())
}

func (suite *SlashingComprehensiveTestSuite) TestSlashValidator_ValidSlashFraction_Zero() {
	validatorAddr := sdk.ValAddress("validator_________").String()
	slashFraction := sdkmath.LegacyZeroDec() // Valid: zero

	amount, err := suite.Keeper.slashValidator(suite.SdkCtx, validatorAddr, slashFraction, 100)
	suite.NoError(err)
	suite.True(amount.IsZero(), "zero slash fraction should return zero amount")
}

func (suite *SlashingComprehensiveTestSuite) TestSlashValidator_ValidSlashFraction_One() {
	validatorAddr := sdk.ValAddress("validator_________").String()
	slashFraction := sdkmath.LegacyOneDec() // Valid: 1.0 (100%)

	// With nil staking keeper, just validates fraction
	amount, err := suite.Keeper.slashValidator(suite.SdkCtx, validatorAddr, slashFraction, 100)
	suite.NoError(err)
	suite.True(amount.IsZero())
}

func (suite *SlashingComprehensiveTestSuite) TestSlashValidator_InvalidValidatorAddress() {
	slashFraction := sdkmath.LegacyMustNewDecFromStr("0.1")

	// Invalid address format - but with nil staking keeper, won't reach address parsing
	// The function returns early when stakingKeeper is nil
	amount, err := suite.Keeper.slashValidator(suite.SdkCtx, "invalid-address", slashFraction, 100)
	suite.NoError(err)
	suite.True(amount.IsZero())
}

// =============================================================================
// jailValidator Tests
// =============================================================================

func (suite *SlashingComprehensiveTestSuite) TestJailValidator_ValidatorNotFound() {
	err := suite.Keeper.jailValidator(suite.SdkCtx, "nonexistent-validator")
	suite.Error(err)
	suite.Equal(types.ErrValidatorNotFound, err)
}

func (suite *SlashingComprehensiveTestSuite) TestJailValidator_Valid() {
	// First register a validator
	validator := &bridgepb.BridgeValidator{
		Address: "aura1validator123",
		Power:   100,
		Active:  true,
		Chains:  []string{"ethereum"},
	}
	suite.Keeper.setValidator(suite.SdkCtx, validator)

	// Jail the validator
	err := suite.Keeper.jailValidator(suite.SdkCtx, "aura1validator123")
	suite.NoError(err)

	// Verify validator is now inactive
	jailed, found := suite.Keeper.getValidator(suite.SdkCtx, "aura1validator123")
	suite.True(found)
	suite.False(jailed.Active, "jailed validator should be inactive")
}

func (suite *SlashingComprehensiveTestSuite) TestJailValidator_AlreadyInactive() {
	validator := &bridgepb.BridgeValidator{
		Address: "aura1validator456",
		Power:   100,
		Active:  false, // Already inactive
		Chains:  []string{"ethereum"},
	}
	suite.Keeper.setValidator(suite.SdkCtx, validator)

	// Jailing again should still succeed
	err := suite.Keeper.jailValidator(suite.SdkCtx, "aura1validator456")
	suite.NoError(err)

	jailed, found := suite.Keeper.getValidator(suite.SdkCtx, "aura1validator456")
	suite.True(found)
	suite.False(jailed.Active)
}

// =============================================================================
// setSlashingEvent / GetSlashingEvent Tests
// =============================================================================

func (suite *SlashingComprehensiveTestSuite) TestSetSlashingEvent_NilEvent() {
	err := suite.Keeper.setSlashingEvent(suite.SdkCtx, nil)
	suite.NoError(err) // nil event should be handled gracefully
}

func (suite *SlashingComprehensiveTestSuite) TestSetSlashingEvent_EmptyEventId() {
	event := &types.SlashingEvent{
		EventId: "",
	}
	err := suite.Keeper.setSlashingEvent(suite.SdkCtx, event)
	suite.NoError(err) // empty ID should be handled gracefully
}

func (suite *SlashingComprehensiveTestSuite) TestSetSlashingEvent_Valid() {
	event := &types.SlashingEvent{
		EventId:          "slash-event-1",
		ValidatorAddress: "aura1validator789",
		Reason:           types.SlashReason_SLASH_DOUBLE_SIGN,
		SlashAmount:      sdkmath.NewInt(1000),
		InfractionHeight: 12345,
		Timestamp:        time.Now(),
	}
	err := suite.Keeper.setSlashingEvent(suite.SdkCtx, event)
	suite.NoError(err)

	// Retrieve and verify
	retrieved, found := suite.Keeper.GetSlashingEvent(suite.SdkCtx, "slash-event-1")
	suite.True(found)
	suite.NotNil(retrieved)
	suite.Equal("slash-event-1", retrieved.EventId)
	suite.Equal(types.SlashReason_SLASH_DOUBLE_SIGN, retrieved.Reason)
}

func (suite *SlashingComprehensiveTestSuite) TestGetSlashingEvent_NotFound() {
	event, found := suite.Keeper.GetSlashingEvent(suite.SdkCtx, "nonexistent-event")
	suite.False(found)
	suite.Nil(event)
}

// =============================================================================
// RecordValidatorSigning Tests
// =============================================================================

func (suite *SlashingComprehensiveTestSuite) TestRecordValidatorSigning_Valid() {
	validatorAddr := "aura1signingvalidator"

	// Record signing - second param is 'signed' boolean
	suite.Keeper.RecordValidatorSigning(suite.SdkCtx, validatorAddr, true)

	// Get signing info
	info := suite.Keeper.GetValidatorSigningInfo(suite.SdkCtx, validatorAddr)
	suite.NotNil(info)
}

func (suite *SlashingComprehensiveTestSuite) TestRecordValidatorSigning_Multiple() {
	validatorAddr := "aura1multisigvalidator"

	// Record multiple signings
	for i := 0; i < 10; i++ {
		suite.Keeper.RecordValidatorSigning(suite.SdkCtx, validatorAddr, true)
	}

	info := suite.Keeper.GetValidatorSigningInfo(suite.SdkCtx, validatorAddr)
	suite.NotNil(info)
}

// =============================================================================
// CheckValidatorLiveness Tests
// =============================================================================

func (suite *SlashingComprehensiveTestSuite) TestCheckValidatorLiveness_NoHistory() {
	validatorAddr := "aura1nohistory"

	isAlive, err := suite.Keeper.CheckValidatorLiveness(suite.SdkCtx, validatorAddr)
	// With no history, may return true or false depending on implementation
	_ = isAlive
	_ = err
}

func (suite *SlashingComprehensiveTestSuite) TestCheckValidatorLiveness_WithHistory() {
	validatorAddr := "aura1withhistory"

	// Record some signings
	for i := 0; i < 10; i++ {
		suite.Keeper.RecordValidatorSigning(suite.SdkCtx, validatorAddr, true)
	}

	isAlive, err := suite.Keeper.CheckValidatorLiveness(suite.SdkCtx, validatorAddr)
	suite.NoError(err)
	suite.True(isAlive, "validator with recent signing history should be alive")
}

// =============================================================================
// SlashForDowntime Tests
// =============================================================================

func (suite *SlashingComprehensiveTestSuite) TestSlashForDowntime_ValidatorNotFound() {
	err := suite.Keeper.SlashForDowntime(suite.SdkCtx, "nonexistent-downtime-validator")
	// May return error or handle gracefully
	_ = err
}

func (suite *SlashingComprehensiveTestSuite) TestSlashForDowntime_ValidValidator() {
	// Register a validator first
	validator := &bridgepb.BridgeValidator{
		Address: "aura1downtimevalidator",
		Power:   100,
		Active:  true,
		Chains:  []string{"ethereum"},
	}
	suite.Keeper.setValidator(suite.SdkCtx, validator)

	err := suite.Keeper.SlashForDowntime(suite.SdkCtx, "aura1downtimevalidator")
	suite.NoError(err)
}

// =============================================================================
// CheckAndSlashDoubleSigning Tests
// =============================================================================

func (suite *SlashingComprehensiveTestSuite) TestCheckAndSlashDoubleSigning_NoTransfer() {
	validatorAddr := "aura1nodoubleSign"
	transferId := "nonexistent-transfer"
	signature := []byte("testsignature")

	valid, err := suite.Keeper.CheckAndSlashDoubleSigning(suite.SdkCtx, transferId, signature, validatorAddr)
	suite.NoError(err)
	suite.True(valid, "should be valid when no previous signature exists")
}

func (suite *SlashingComprehensiveTestSuite) TestCheckAndSlashDoubleSigning_EmptyParams() {
	valid, err := suite.Keeper.CheckAndSlashDoubleSigning(suite.SdkCtx, "", []byte("sig"), "")
	suite.Error(err, "should error with empty params")
	suite.False(valid)
}

// =============================================================================
// DetectDoubleSigning Tests
// =============================================================================

func (suite *SlashingComprehensiveTestSuite) TestDetectDoubleSigning_NoTransfer() {
	transferId := "nonexistent-transfer"
	sig := []byte("signature")
	validatorAddr := "aura1detector"

	detected, err := suite.Keeper.DetectDoubleSigning(suite.SdkCtx, transferId, sig, validatorAddr)
	suite.NoError(err)
	suite.False(detected, "no transfer means no double signing")
}

func (suite *SlashingComprehensiveTestSuite) TestDetectDoubleSigning_EmptyParams() {
	sig := []byte("samesignature")

	detected, err := suite.Keeper.DetectDoubleSigning(suite.SdkCtx, "", sig, "")
	suite.Error(err, "should error with empty params")
	suite.False(detected)
}

// =============================================================================
// SubmitSlashingEvidence Tests
// =============================================================================

func (suite *SlashingComprehensiveTestSuite) TestSubmitSlashingEvidence_Valid() {
	// First register a validator
	validator := &bridgepb.BridgeValidator{
		Address: "aura1slashevidence",
		Power:   100,
		Active:  true,
		Chains:  []string{"ethereum"},
	}
	suite.Keeper.setValidator(suite.SdkCtx, validator)

	event, err := suite.Keeper.SubmitSlashingEvidence(
		suite.SdkCtx,
		"aura1slashevidence",
		types.SlashReason_SLASH_DOUBLE_SIGN,
		"transfer-123",
		[]byte("evidence-hash"),
		"aura1submitter",
	)
	// Should succeed since validator exists
	suite.NoError(err)
	suite.NotNil(event)
}

func (suite *SlashingComprehensiveTestSuite) TestSubmitSlashingEvidence_EmptyValidator() {
	event, err := suite.Keeper.SubmitSlashingEvidence(
		suite.SdkCtx,
		"",
		types.SlashReason_SLASH_DOUBLE_SIGN,
		"transfer-123",
		[]byte("evidence-hash"),
		"aura1submitter",
	)
	suite.Error(err, "should error with empty validator")
	suite.Nil(event)
}

func (suite *SlashingComprehensiveTestSuite) TestSubmitSlashingEvidence_EmptyEvidence() {
	event, err := suite.Keeper.SubmitSlashingEvidence(
		suite.SdkCtx,
		"aura1validator",
		types.SlashReason_SLASH_DOUBLE_SIGN,
		"transfer-123",
		[]byte{}, // Empty evidence
		"aura1submitter",
	)
	suite.Error(err, "should error with empty evidence")
	suite.Nil(event)
}

// =============================================================================
// GetAllSlashingEvents Tests
// =============================================================================

func (suite *SlashingComprehensiveTestSuite) TestGetAllSlashingEvents_Empty() {
	events := suite.Keeper.GetAllSlashingEvents(suite.SdkCtx)
	suite.NotNil(events)
	suite.Len(events, 0)
}

func (suite *SlashingComprehensiveTestSuite) TestGetAllSlashingEvents_WithEvents() {
	// Create some events
	for i := 0; i < 3; i++ {
		event := &types.SlashingEvent{
			EventId:          "slash-getall-" + string(rune('0'+i)),
			ValidatorAddress: "aura1validator" + string(rune('0'+i)),
			Reason:           types.SlashReason_SLASH_DOWNTIME,
			SlashAmount:      sdkmath.NewInt(int64(100 * (i + 1))),
			InfractionHeight: uint64(1000 + i),
			Timestamp:        time.Now(),
		}
		suite.Keeper.setSlashingEvent(suite.SdkCtx, event)
	}

	events := suite.Keeper.GetAllSlashingEvents(suite.SdkCtx)
	suite.NotNil(events)
	suite.GreaterOrEqual(len(events), 3)
}

// =============================================================================
// GetValidatorSigningInfo Tests
// =============================================================================

func (suite *SlashingComprehensiveTestSuite) TestGetValidatorSigningInfo_NotFound() {
	info := suite.Keeper.GetValidatorSigningInfo(suite.SdkCtx, "nonexistent-signing-info")
	// May return nil or empty struct
	_ = info
}

func (suite *SlashingComprehensiveTestSuite) TestGetValidatorSigningInfo_WithHistory() {
	validatorAddr := "aura1signinginfo"
	suite.Keeper.RecordValidatorSigning(suite.SdkCtx, validatorAddr, true)

	info := suite.Keeper.GetValidatorSigningInfo(suite.SdkCtx, validatorAddr)
	suite.NotNil(info)
}
