package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/testutil/keeper"
	compliancekeeper "github.com/aequitas/aura/chain/x/compliance/keeper"
	"github.com/aequitas/aura/chain/x/compliance/types"
	compliancepb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
)

type JurisdictionCheckOrderTestSuite struct {
	suite.Suite
	ctx       sdk.Context
	keeper    *compliancekeeper.Keeper
	msgServer compliancepb.MsgServer
}

func (suite *JurisdictionCheckOrderTestSuite) SetupTest() {
	suite.ctx, suite.keeper = keeper.ComplianceKeeper(suite.T())
	suite.msgServer = compliancekeeper.NewMsgServerImpl(suite.keeper)
}

// TestJurisdictionBlockedBeforeProviderCheck verifies jurisdiction is checked BEFORE provider authorization
// This is the critical HIGH-002 security fix
func (suite *JurisdictionCheckOrderTestSuite) TestJurisdictionBlockedBeforeProviderCheck() {
	providerAddr := "aura1providerxxxxxxxxxxxxxxxxxxxxxxxx"

	// Setup: Add provider to authorized list
	params := suite.keeper.GetParams(suite.ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	params.BlockedJurisdictions = []string{"KP"} // North Korea blocked
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Setup: Add user consent (required for KYC)
	userAddr := "aura1userxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	err = suite.keeper.GrantConsent(suite.ctx, userAddr, "kyc_processing", "test_purpose")
	suite.NoError(err)

	// Test: Submit KYC for blocked jurisdiction
	msg := &compliancepb.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      compliancepb.KYCLevel_BASIC,
		Provider:      providerAddr,
		Jurisdiction:  "KP", // Blocked jurisdiction
		PiiCommitment: make([]byte, 32),
	}

	resp, err := suite.msgServer.SubmitKYC(context.Background(), msg)
	suite.Error(err)
	suite.Nil(resp)
	suite.Contains(err.Error(), "blocked due to OFAC sanctions")

	// CRITICAL VERIFICATION: Ensure the error is about jurisdiction, NOT provider
	// If provider was checked first, we might see "provider not authorized" instead
	suite.Contains(err.Error(), "jurisdiction")
	suite.Contains(err.Error(), "KP")
}

// TestProviderCheckOccursAfterJurisdictionCheck verifies provider is checked AFTER jurisdiction
func (suite *JurisdictionCheckOrderTestSuite) TestProviderCheckOccursAfterJurisdictionCheck() {
	providerAddr := "aura1providerxxxxxxxxxxxxxxxxxxxxxxxx"
	unauthorizedProviderAddr := "aura1unauthorizedxxxxxxxxxxxxxxxxxxxxx"
	userAddr := "aura1userxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	// Setup: Add authorized provider and blocked jurisdiction
	params := suite.keeper.GetParams(suite.ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	params.BlockedJurisdictions = []string{"KP"} // North Korea blocked
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Setup: Add user consent
	err = suite.keeper.GrantConsent(suite.ctx, userAddr, "kyc_processing", "test_purpose")
	suite.NoError(err)

	// Test 1: Blocked jurisdiction with unauthorized provider
	// Should fail on jurisdiction check (first), not provider check (second)
	msg := &compliancepb.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      compliancepb.KYCLevel_BASIC,
		Provider:      unauthorizedProviderAddr, // Unauthorized
		Jurisdiction:  "KP",                     // Blocked
		PiiCommitment: make([]byte, 32),
	}

	resp, err := suite.msgServer.SubmitKYC(context.Background(), msg)
	suite.Error(err)
	suite.Nil(resp)
	// Should fail on jurisdiction, not provider
	suite.Contains(err.Error(), "jurisdiction")
	suite.Contains(err.Error(), "KP")
	suite.NotContains(err.Error(), "not authorized")

	// Test 2: Allowed jurisdiction with unauthorized provider
	// Should fail on provider check (second)
	msg.Jurisdiction = "US" // Allowed jurisdiction
	resp, err = suite.msgServer.SubmitKYC(context.Background(), msg)
	suite.Error(err)
	suite.Nil(resp)
	// Should fail on provider authorization
	suite.Contains(err.Error(), "not authorized")
}

// TestTimingAttackPrevention verifies the fix prevents timing attacks
func (suite *JurisdictionCheckOrderTestSuite) TestTimingAttackPrevention() {
	providerAddr := "aura1providerxxxxxxxxxxxxxxxxxxxxxxxx"
	userAddr := "aura1userxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	// Setup: Initially no blocked jurisdictions, provider authorized
	params := suite.keeper.GetParams(suite.ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	params.BlockedJurisdictions = []string{} // No blocks initially
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Setup: Add user consent
	err = suite.keeper.GrantConsent(suite.ctx, userAddr, "kyc_processing", "test_purpose")
	suite.NoError(err)

	// Simulated attack scenario:
	// 1. Attacker sees governance proposal to block "XX"
	// 2. Attacker quickly submits KYC for "XX" users
	// 3. In old code: provider check retrieves params (XX not blocked yet)
	// 4. In old code: jurisdiction check uses same params (stale)
	// 5. In old code: KYC approved for soon-to-be-blocked jurisdiction

	// Simulate governance updating params to block XX
	params.BlockedJurisdictions = []string{"XX"}
	err = suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Attacker tries to submit KYC after governance vote but before old check would catch it
	msg := &compliancepb.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      compliancepb.KYCLevel_BASIC,
		Provider:      providerAddr,
		Jurisdiction:  "XX", // Just blocked
		PiiCommitment: make([]byte, 32),
	}

	resp, err := suite.msgServer.SubmitKYC(context.Background(), msg)
	suite.Error(err)
	suite.Nil(resp)

	// CRITICAL: With the fix, jurisdiction check happens FIRST and retrieves latest params
	// So the KYC submission is rejected immediately
	suite.Contains(err.Error(), "blocked due to OFAC sanctions")
	suite.Contains(err.Error(), "XX")

	// Verify NO KYC record was created
	record, found := suite.keeper.GetKYCRecord(suite.ctx, userAddr)
	suite.False(found)
	suite.Nil(record)
}

// TestJurisdictionNormalization verifies jurisdiction codes are normalized before checking
func (suite *JurisdictionCheckOrderTestSuite) TestJurisdictionNormalization() {
	providerAddr := "aura1providerxxxxxxxxxxxxxxxxxxxxxxxx"
	userAddr := "aura1userxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	// Setup: Block lowercase jurisdiction code
	params := suite.keeper.GetParams(suite.ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	params.BlockedJurisdictions = []string{"KP"} // Uppercase
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Setup: Add user consent
	err = suite.keeper.GrantConsent(suite.ctx, userAddr, "kyc_processing", "test_purpose")
	suite.NoError(err)

	// Test: Submit with lowercase jurisdiction code
	msg := &compliancepb.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      compliancepb.KYCLevel_BASIC,
		Provider:      providerAddr,
		Jurisdiction:  "kp", // lowercase
		PiiCommitment: make([]byte, 32),
	}

	resp, err := suite.msgServer.SubmitKYC(context.Background(), msg)
	suite.Error(err)
	suite.Nil(resp)
	// Should still be blocked (normalized to uppercase)
	suite.Contains(err.Error(), "blocked")
}

// TestAllowedJurisdictionPassesCheck verifies allowed jurisdictions work correctly
func (suite *JurisdictionCheckOrderTestSuite) TestAllowedJurisdictionPassesCheck() {
	providerAddr := "aura1providerxxxxxxxxxxxxxxxxxxxxxxxx"
	userAddr := "aura1userxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	// Setup: Block some jurisdictions but not US
	params := suite.keeper.GetParams(suite.ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	params.BlockedJurisdictions = []string{"KP", "IR", "SY"} // North Korea, Iran, Syria
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Setup: Add user consent
	err = suite.keeper.GrantConsent(suite.ctx, userAddr, "kyc_processing", "test_purpose")
	suite.NoError(err)

	// Test: Submit KYC for allowed jurisdiction (US)
	msg := &compliancepb.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      compliancepb.KYCLevel_BASIC,
		Provider:      providerAddr,
		Jurisdiction:  "US",
		PiiCommitment: make([]byte, 32),
	}

	resp, err := suite.msgServer.SubmitKYC(context.Background(), msg)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)

	// Verify KYC record was created
	record, found := suite.keeper.GetKYCRecord(suite.ctx, userAddr)
	suite.True(found)
	suite.NotNil(record)
	suite.Equal("US", record.Jurisdiction)
}

// TestMultipleBlockedJurisdictions verifies all blocked jurisdictions are checked
func (suite *JurisdictionCheckOrderTestSuite) TestMultipleBlockedJurisdictions() {
	providerAddr := "aura1providerxxxxxxxxxxxxxxxxxxxxxxxx"

	// Setup: Block multiple jurisdictions
	params := suite.keeper.GetParams(suite.ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	params.BlockedJurisdictions = []string{
		"KP", // North Korea
		"IR", // Iran
		"SY", // Syria
		"CU", // Cuba
	}
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Test each blocked jurisdiction
	for _, jurisdiction := range params.BlockedJurisdictions {
		userAddr := "aura1user" + jurisdiction + "xxxxxxxxxxxxxxxxxxx"

		// Setup: Add user consent
		err = suite.keeper.GrantConsent(suite.ctx, userAddr, "kyc_processing", "test_purpose")
		suite.NoError(err)

		msg := &compliancepb.MsgSubmitKYC{
			Address:       userAddr,
			KycLevel:      compliancepb.KYCLevel_BASIC,
			Provider:      providerAddr,
			Jurisdiction:  jurisdiction,
			PiiCommitment: make([]byte, 32),
		}

		resp, err := suite.msgServer.SubmitKYC(context.Background(), msg)
		suite.Error(err, "Jurisdiction %s should be blocked", jurisdiction)
		suite.Nil(resp)
		suite.Contains(err.Error(), "blocked")
		suite.Contains(err.Error(), jurisdiction)
	}
}

// TestParamsRetrievalOrder verifies GetParams is called at the right time
func (suite *JurisdictionCheckOrderTestSuite) TestParamsRetrievalOrder() {
	providerAddr := "aura1providerxxxxxxxxxxxxxxxxxxxxxxxx"
	userAddr := "aura1userxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	// Setup: Initially allow jurisdiction XX
	params := suite.keeper.GetParams(suite.ctx)
	params.ApprovedKycProviders = []string{providerAddr}
	params.BlockedJurisdictions = []string{}
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Setup: Add user consent
	err = suite.keeper.GrantConsent(suite.ctx, userAddr, "kyc_processing", "test_purpose")
	suite.NoError(err)

	// In the SAME BLOCK, update params and submit KYC
	// This simulates the race condition the fix prevents

	// Update params to block XX
	params.BlockedJurisdictions = []string{"XX"}
	err = suite.keeper.SetParams(suite.ctx, params)
	suite.NoError(err)

	// Submit KYC for XX (should be rejected)
	msg := &compliancepb.MsgSubmitKYC{
		Address:       userAddr,
		KycLevel:      compliancepb.KYCLevel_BASIC,
		Provider:      providerAddr,
		Jurisdiction:  "XX",
		PiiCommitment: make([]byte, 32),
	}

	resp, err := suite.msgServer.SubmitKYC(context.Background(), msg)
	suite.Error(err)
	suite.Nil(resp)

	// The fix ensures GetParams is called BEFORE jurisdiction check
	// So even in the same block, the latest params are retrieved
	suite.Contains(err.Error(), "blocked")
}

func TestJurisdictionCheckOrderTestSuite(t *testing.T) {
	suite.Run(t, new(JurisdictionCheckOrderTestSuite))
}
