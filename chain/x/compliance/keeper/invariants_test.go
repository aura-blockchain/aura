package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/compliance/types"
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

	// Test: invalid params fail (set invalid params directly in store)
	// Note: This is a hypothetical test - in practice params validation happens before storage
}

func (suite *InvariantsTestSuite) TestKYCRecordConsistencyInvariant() {
	ctx := suite.SdkCtx

	// Test: valid KYC record passes
	validAddr := sdk.AccAddress("addr1_______________").String()
	validRecord := types.KYCRecord{
		Address:        validAddr,
		VerificationId: "kyc-123",
		Status:         "approved",
		SubmittedAt:    timestamppb.Now(),
		Tier:           1,
	}
	suite.Keeper.SetKYCRecord(ctx, validRecord)

	inv := KYCRecordConsistencyInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "valid KYC record should pass")
	suite.Empty(msg)

	// Test: invalid address in KYC record fails
	invalidRecord := types.KYCRecord{
		Address:        "invalid-address",
		VerificationId: "kyc-456",
		Status:         "approved",
		SubmittedAt:    timestamppb.Now(),
		Tier:           1,
	}
	suite.Keeper.SetKYCRecord(ctx, invalidRecord)

	msg, broken = inv(ctx)
	suite.True(broken, "invalid address should break invariant")
	suite.Contains(msg, "invalid address")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: empty verification ID fails
	emptyVerifRecord := types.KYCRecord{
		Address:        validAddr,
		VerificationId: "",
		Status:         "approved",
		SubmittedAt:    timestamppb.Now(),
		Tier:           1,
	}
	suite.Keeper.SetKYCRecord(ctx, emptyVerifRecord)

	msg, broken = inv(ctx)
	suite.True(broken, "empty verification ID should break invariant")
	suite.Contains(msg, "empty verification ID")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: invalid status fails
	invalidStatusRecord := types.KYCRecord{
		Address:        validAddr,
		VerificationId: "kyc-789",
		Status:         "invalid-status",
		SubmittedAt:    timestamppb.Now(),
		Tier:           1,
	}
	suite.Keeper.SetKYCRecord(ctx, invalidStatusRecord)

	msg, broken = inv(ctx)
	suite.True(broken, "invalid status should break invariant")
	suite.Contains(msg, "invalid status")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: nil submitted_at fails
	nilTimeRecord := types.KYCRecord{
		Address:        validAddr,
		VerificationId: "kyc-101",
		Status:         "approved",
		SubmittedAt:    nil,
		Tier:           1,
	}
	suite.Keeper.SetKYCRecord(ctx, nilTimeRecord)

	msg, broken = inv(ctx)
	suite.True(broken, "nil submitted_at should break invariant")
	suite.Contains(msg, "nil submitted_at")
}

func (suite *InvariantsTestSuite) TestSanctionsScreeningInvariant() {
	ctx := suite.SdkCtx

	// Test: valid sanctions screening passes
	validAddr := sdk.AccAddress("addr1_______________").String()
	validResult := types.SanctionsScreeningResult{
		Address:    validAddr,
		Flagged:    false,
		ScreenedAt: timestamppb.Now(),
		Matches:    []string{},
	}
	suite.Keeper.SetSanctionsScreening(ctx, validResult)

	inv := SanctionsScreeningInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "valid sanctions screening should pass")
	suite.Empty(msg)

	// Test: invalid address fails
	invalidResult := types.SanctionsScreeningResult{
		Address:    "invalid-address",
		Flagged:    false,
		ScreenedAt: timestamppb.Now(),
		Matches:    []string{},
	}
	suite.Keeper.SetSanctionsScreening(ctx, invalidResult)

	msg, broken = inv(ctx)
	suite.True(broken, "invalid address should break invariant")
	suite.Contains(msg, "invalid address")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: nil screened_at fails
	nilTimeResult := types.SanctionsScreeningResult{
		Address:    validAddr,
		Flagged:    false,
		ScreenedAt: nil,
		Matches:    []string{},
	}
	suite.Keeper.SetSanctionsScreening(ctx, nilTimeResult)

	msg, broken = inv(ctx)
	suite.True(broken, "nil screened_at should break invariant")
	suite.Contains(msg, "nil screened_at")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: flagged with no matches fails
	flaggedNoMatches := types.SanctionsScreeningResult{
		Address:    validAddr,
		Flagged:    true,
		ScreenedAt: timestamppb.Now(),
		Matches:    []string{},
	}
	suite.Keeper.SetSanctionsScreening(ctx, flaggedNoMatches)

	msg, broken = inv(ctx)
	suite.True(broken, "flagged with no matches should break invariant")
	suite.Contains(msg, "flagged but has no matches")
}

func (suite *InvariantsTestSuite) TestGDPRDataIntegrityInvariant() {
	ctx := suite.SdkCtx

	// Test: valid GDPR request passes
	validAddr := sdk.AccAddress("addr1_______________").String()
	validRequest := types.GDPRRequest{
		RequesterAddress: validAddr,
		RequestType:      "access",
		Status:           "pending",
		SubmittedAt:      timestamppb.Now(),
		ProcessedAt:      nil,
	}
	suite.Keeper.SetGDPRRequest(ctx, validRequest)

	inv := GDPRDataIntegrityInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "valid GDPR request should pass")
	suite.Empty(msg)

	// Test: invalid requester address fails
	invalidRequest := types.GDPRRequest{
		RequesterAddress: "invalid-address",
		RequestType:      "access",
		Status:           "pending",
		SubmittedAt:      timestamppb.Now(),
	}
	suite.Keeper.SetGDPRRequest(ctx, invalidRequest)

	msg, broken = inv(ctx)
	suite.True(broken, "invalid requester address should break invariant")
	suite.Contains(msg, "invalid requester address")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: invalid request type fails
	invalidTypeRequest := types.GDPRRequest{
		RequesterAddress: validAddr,
		RequestType:      "invalid-type",
		Status:           "pending",
		SubmittedAt:      timestamppb.Now(),
	}
	suite.Keeper.SetGDPRRequest(ctx, invalidTypeRequest)

	msg, broken = inv(ctx)
	suite.True(broken, "invalid request type should break invariant")
	suite.Contains(msg, "invalid type")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: nil submitted_at fails
	nilTimeRequest := types.GDPRRequest{
		RequesterAddress: validAddr,
		RequestType:      "access",
		Status:           "pending",
		SubmittedAt:      nil,
	}
	suite.Keeper.SetGDPRRequest(ctx, nilTimeRequest)

	msg, broken = inv(ctx)
	suite.True(broken, "nil submitted_at should break invariant")
	suite.Contains(msg, "nil submitted_at")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: processed without processed_at fails
	processedNoTime := types.GDPRRequest{
		RequesterAddress: validAddr,
		RequestType:      "access",
		Status:           "processed",
		SubmittedAt:      timestamppb.Now(),
		ProcessedAt:      nil,
	}
	suite.Keeper.SetGDPRRequest(ctx, processedNoTime)

	msg, broken = inv(ctx)
	suite.True(broken, "processed without processed_at should break invariant")
	suite.Contains(msg, "nil processed_at")
}

func (suite *InvariantsTestSuite) TestTaxRecordConsistencyInvariant() {
	ctx := suite.SdkCtx

	// Test: valid tax record passes
	validAddr := sdk.AccAddress("addr1_______________").String()
	validRecord := types.TaxRecord{
		Address:      validAddr,
		TaxYear:      2024,
		Jurisdiction: "US",
		TotalIncome:  "1000.00",
		TotalTax:     "200.00",
	}
	suite.Keeper.SetTaxRecord(ctx, validRecord)

	inv := TaxRecordConsistencyInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "valid tax record should pass")
	suite.Empty(msg)

	// Test: invalid address fails
	invalidRecord := types.TaxRecord{
		Address:      "invalid-address",
		TaxYear:      2024,
		Jurisdiction: "US",
		TotalIncome:  "1000.00",
		TotalTax:     "200.00",
	}
	suite.Keeper.SetTaxRecord(ctx, invalidRecord)

	msg, broken = inv(ctx)
	suite.True(broken, "invalid address should break invariant")
	suite.Contains(msg, "invalid address")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: invalid tax year fails
	invalidYearRecord := types.TaxRecord{
		Address:      validAddr,
		TaxYear:      1900,
		Jurisdiction: "US",
		TotalIncome:  "1000.00",
		TotalTax:     "200.00",
	}
	suite.Keeper.SetTaxRecord(ctx, invalidYearRecord)

	msg, broken = inv(ctx)
	suite.True(broken, "invalid tax year should break invariant")
	suite.Contains(msg, "invalid year")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx

	// Test: empty jurisdiction fails
	emptyJurisdiction := types.TaxRecord{
		Address:      validAddr,
		TaxYear:      2024,
		Jurisdiction: "",
		TotalIncome:  "1000.00",
		TotalTax:     "200.00",
	}
	suite.Keeper.SetTaxRecord(ctx, emptyJurisdiction)

	msg, broken = inv(ctx)
	suite.True(broken, "empty jurisdiction should break invariant")
	suite.Contains(msg, "empty jurisdiction")
}
