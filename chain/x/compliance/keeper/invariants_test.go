package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

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

	inv := AllInvariants(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Cosmos SDK v0.53 lacks sdk.NewInvariantRegistry; call invariants directly.
	invariants := []func(sdk.Context) (string, bool){
		ParamsInvariant(suite.Keeper),
		KYCRecordConsistencyInvariant(suite.Keeper),
		SanctionsScreeningInvariant(suite.Keeper),
		GDPRDataIntegrityInvariant(suite.Keeper),
		TaxRecordConsistencyInvariant(suite.Keeper),
	}

	for _, inv := range invariants {
		msg, broken := inv(suite.SdkCtx)
		suite.False(broken)
		suite.Empty(msg)
	}
}

func (suite *InvariantsTestSuite) TestKYCRecordConsistencyInvariant() {
	ctx := suite.SdkCtx
	inv := KYCRecordConsistencyInvariant(suite.Keeper)

	validAddr := suite.addr("kyc-valid")
	piiCommitment := make([]byte, 32)
	copy(piiCommitment, []byte("test_commitment_hash_32_bytes"))

	expiresAt := time.Now().Add(time.Hour)
	validRecord := &types.KYCRecord{
		Address:       validAddr,
		PiiCommitment: piiCommitment,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		VerifiedAt:    time.Now(),
		ExpiresAt:     &expiresAt,
	}
	suite.Require().NoError(suite.Keeper.SetKYCRecord(ctx, validRecord))

	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Invalid address
	invalidRecord := &types.KYCRecord{
		Address:       "invalid-address",
		PiiCommitment: piiCommitment,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		VerifiedAt:    time.Now(),
	}
	suite.Require().NoError(suite.Keeper.SetKYCRecord(ctx, invalidRecord))

	msg, broken = inv(ctx)
	suite.True(broken)
	suite.Contains(msg, "invalid address")

	// Reset context
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = KYCRecordConsistencyInvariant(suite.Keeper)

	// Invalid PII commitment length
	emptyCommitment := &types.KYCRecord{
		Address:       validAddr,
		PiiCommitment: []byte("short"),
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
		VerifiedAt:    time.Now(),
	}
	suite.Require().NoError(suite.Keeper.SetKYCRecord(ctx, emptyCommitment))

	msg, broken = inv(ctx)
	suite.True(broken)
	suite.Contains(msg, "invalid PII commitment length")

	// Reset context
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = KYCRecordConsistencyInvariant(suite.Keeper)

	// Invalid level
	badLevel := &types.KYCRecord{
		Address:       validAddr,
		PiiCommitment: piiCommitment,
		KycLevel:      types.KYCLevel(99),
		VerifiedAt:    time.Now(),
	}
	suite.Require().NoError(suite.Keeper.SetKYCRecord(ctx, badLevel))

	msg, broken = inv(ctx)
	suite.True(broken)
	suite.Contains(msg, "invalid KYC level")

	// Reset context
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = KYCRecordConsistencyInvariant(suite.Keeper)

	// Nil verified_at
	nilVerified := &types.KYCRecord{
		Address:       validAddr,
		PiiCommitment: piiCommitment,
		KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
	}
	suite.Require().NoError(suite.Keeper.SetKYCRecord(ctx, nilVerified))

	msg, broken = inv(ctx)
	suite.True(broken)
	suite.Contains(msg, "verified_at")
}

func (suite *InvariantsTestSuite) TestParamsInvariant() {
	ctx := suite.SdkCtx

	// Test: valid params pass
	inv := ParamsInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "valid params should pass")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestSanctionsScreeningInvariant() {
	ctx := suite.SdkCtx

	// Test on empty store - should pass
	inv := SanctionsScreeningInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "sanctions invariant should pass on empty store")
	suite.Empty(msg)

	// Test with valid sanctions screening data (store directly)
	validAddr := suite.addr("sanctions-valid")
	validResult := &types.SanctionsScreeningResult{
		Address:    validAddr,
		Status:     types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:    []*types.SanctionsMatch{},
		ScreenedAt: time.Now(),
	}

	// Write directly to store since there's no public setter
	bz, err := suite.Keeper.cdc.Marshal(validResult)
	suite.Require().NoError(err)
	store := ctx.KVStore(suite.Keeper.storeKey)
	key := append(SanctionsResultsKeyPrefix, []byte(validAddr)...)
	store.Set(key, bz)

	msg, broken = inv(ctx)
	suite.False(broken, "valid sanctions screening should pass")
	suite.Empty(msg)

	// Test: invalid address fails
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = SanctionsScreeningInvariant(suite.Keeper)

	invalidResult := &types.SanctionsScreeningResult{
		Address:    "invalid-address",
		Status:     types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:    []*types.SanctionsMatch{},
		ScreenedAt: time.Now(),
	}
	bz, err = suite.Keeper.cdc.Marshal(invalidResult)
	suite.Require().NoError(err)
	store = ctx.KVStore(suite.Keeper.storeKey)
	key = append(SanctionsResultsKeyPrefix, []byte("invalid-address")...)
	store.Set(key, bz)

	msg, broken = inv(ctx)
	suite.True(broken, "invalid address should break invariant")
	suite.Contains(msg, "invalid address")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = SanctionsScreeningInvariant(suite.Keeper)

	// Test: zero screened_at fails
	validAddr2 := suite.addr("sanctions-zerotime")
	zeroTimeResult := &types.SanctionsScreeningResult{
		Address:    validAddr2,
		Status:     types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:    []*types.SanctionsMatch{},
		ScreenedAt: time.Time{}, // zero time
	}
	bz, err = suite.Keeper.cdc.Marshal(zeroTimeResult)
	suite.Require().NoError(err)
	store = ctx.KVStore(suite.Keeper.storeKey)
	key = append(SanctionsResultsKeyPrefix, []byte(validAddr2)...)
	store.Set(key, bz)

	msg, broken = inv(ctx)
	suite.True(broken, "zero screened_at should break invariant")
	suite.Contains(msg, "screened_at")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = SanctionsScreeningInvariant(suite.Keeper)

	// Test: status MATCH with no matches fails
	validAddr3 := suite.addr("sanctions-nomatch")
	matchNoMatches := &types.SanctionsScreeningResult{
		Address:    validAddr3,
		Status:     types.SanctionsStatus_SANCTIONS_MATCH,
		Matches:    []*types.SanctionsMatch{},
		ScreenedAt: time.Now(),
	}
	bz, err = suite.Keeper.cdc.Marshal(matchNoMatches)
	suite.Require().NoError(err)
	store = ctx.KVStore(suite.Keeper.storeKey)
	key = append(SanctionsResultsKeyPrefix, []byte(validAddr3)...)
	store.Set(key, bz)

	msg, broken = inv(ctx)
	suite.True(broken, "match status with no matches should break invariant")
	suite.Contains(msg, "has match status but no matches")
}

func (suite *InvariantsTestSuite) TestGDPRDataIntegrityInvariant() {
	ctx := suite.SdkCtx

	// Test on empty store - should pass
	inv := GDPRDataIntegrityInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "GDPR invariant should pass on empty store")
	suite.Empty(msg)

	// Test: valid GDPR request passes
	validAddr := suite.addr("gdpr-valid")
	validRequest := &types.GDPRDataRequest{
		Id:          "gdpr-req-001",
		Address:     validAddr,
		RequestType: "access",
		RequestedAt: time.Now(),
		CompletedAt: nil,
		Status:      "pending",
	}
	suite.Require().NoError(suite.Keeper.SetGDPRRequest(ctx, validRequest))

	msg, broken = inv(ctx)
	suite.False(broken, "valid GDPR request should pass")
	suite.Empty(msg)

	// Test: invalid address fails
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = GDPRDataIntegrityInvariant(suite.Keeper)

	invalidRequest := &types.GDPRDataRequest{
		Id:          "gdpr-req-002",
		Address:     "invalid-address",
		RequestType: "access",
		RequestedAt: time.Now(),
		Status:      "pending",
	}
	suite.Require().NoError(suite.Keeper.SetGDPRRequest(ctx, invalidRequest))

	msg, broken = inv(ctx)
	suite.True(broken, "invalid address should break invariant")
	suite.Contains(msg, "invalid address")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = GDPRDataIntegrityInvariant(suite.Keeper)

	// Test: invalid request type fails
	validAddr2 := suite.addr("gdpr-badtype")
	invalidTypeRequest := &types.GDPRDataRequest{
		Id:          "gdpr-req-003",
		Address:     validAddr2,
		RequestType: "invalid-type",
		RequestedAt: time.Now(),
		Status:      "pending",
	}
	suite.Require().NoError(suite.Keeper.SetGDPRRequest(ctx, invalidTypeRequest))

	msg, broken = inv(ctx)
	suite.True(broken, "invalid request type should break invariant")
	suite.Contains(msg, "invalid type")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = GDPRDataIntegrityInvariant(suite.Keeper)

	// Test: zero requested_at fails
	validAddr3 := suite.addr("gdpr-zerotime")
	zeroTimeRequest := &types.GDPRDataRequest{
		Id:          "gdpr-req-004",
		Address:     validAddr3,
		RequestType: "access",
		RequestedAt: time.Time{}, // zero time
		Status:      "pending",
	}
	suite.Require().NoError(suite.Keeper.SetGDPRRequest(ctx, zeroTimeRequest))

	msg, broken = inv(ctx)
	suite.True(broken, "zero requested_at should break invariant")
	suite.Contains(msg, "requested_at")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = GDPRDataIntegrityInvariant(suite.Keeper)

	// Test: completed status without completed_at fails
	validAddr4 := suite.addr("gdpr-nocompletetime")
	completedNoTime := &types.GDPRDataRequest{
		Id:          "gdpr-req-005",
		Address:     validAddr4,
		RequestType: "access",
		RequestedAt: time.Now(),
		CompletedAt: nil,
		Status:      "completed",
	}
	suite.Require().NoError(suite.Keeper.SetGDPRRequest(ctx, completedNoTime))

	msg, broken = inv(ctx)
	suite.True(broken, "completed status without completed_at should break invariant")
	suite.Contains(msg, "nil completed_at")
}

func (suite *InvariantsTestSuite) TestTaxRecordConsistencyInvariant() {
	ctx := suite.SdkCtx

	// Test on empty store - should pass
	inv := TaxRecordConsistencyInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "tax invariant should pass on empty store")
	suite.Empty(msg)

	// Test: valid tax report passes
	validAddr := suite.addr("tax-valid")
	validReport := &types.TaxReport{
		Id:           "tax-report-001",
		Address:      validAddr,
		TaxYear:      "2024",
		Jurisdiction: "US",
		ReportType:   "1099-MISC",
		TotalIncome:  "1000.00",
		GeneratedAt:  time.Now(),
	}
	suite.Require().NoError(suite.Keeper.SetTaxReport(ctx, validReport))

	msg, broken = inv(ctx)
	suite.False(broken, "valid tax report should pass")
	suite.Empty(msg)

	// Test: invalid address fails
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = TaxRecordConsistencyInvariant(suite.Keeper)

	invalidReport := &types.TaxReport{
		Id:           "tax-report-002",
		Address:      "invalid-address",
		TaxYear:      "2024",
		Jurisdiction: "US",
		ReportType:   "1099-MISC",
		TotalIncome:  "1000.00",
		GeneratedAt:  time.Now(),
	}
	suite.Require().NoError(suite.Keeper.SetTaxReport(ctx, invalidReport))

	msg, broken = inv(ctx)
	suite.True(broken, "invalid address should break invariant")
	suite.Contains(msg, "invalid address")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = TaxRecordConsistencyInvariant(suite.Keeper)

	// Test: invalid tax year fails
	validAddr2 := suite.addr("tax-badyear")
	invalidYearReport := &types.TaxReport{
		Id:           "tax-report-003",
		Address:      validAddr2,
		TaxYear:      "1900",
		Jurisdiction: "US",
		ReportType:   "1099-MISC",
		TotalIncome:  "1000.00",
		GeneratedAt:  time.Now(),
	}
	suite.Require().NoError(suite.Keeper.SetTaxReport(ctx, invalidYearReport))

	msg, broken = inv(ctx)
	suite.True(broken, "invalid tax year should break invariant")
	suite.Contains(msg, "invalid year")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = TaxRecordConsistencyInvariant(suite.Keeper)

	// Test: empty jurisdiction fails
	validAddr3 := suite.addr("tax-nojuris")
	emptyJurisdiction := &types.TaxReport{
		Id:           "tax-report-004",
		Address:      validAddr3,
		TaxYear:      "2024",
		Jurisdiction: "",
		ReportType:   "1099-MISC",
		TotalIncome:  "1000.00",
		GeneratedAt:  time.Now(),
	}
	suite.Require().NoError(suite.Keeper.SetTaxReport(ctx, emptyJurisdiction))

	msg, broken = inv(ctx)
	suite.True(broken, "empty jurisdiction should break invariant")
	suite.Contains(msg, "empty jurisdiction")
}

func (suite *InvariantsTestSuite) addr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	bech32, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	suite.Require().NoError(err)
	return bech32
}
