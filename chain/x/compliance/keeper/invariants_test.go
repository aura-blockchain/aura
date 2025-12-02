package keeper

import (
	"crypto/sha256"
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
	validRecord := &types.KYCRecord{
		Address:        validAddr,
		VerificationId: "kyc-123",
		KycLevel:       types.KYCLevel_KYC_LEVEL_BASIC,
		VerifiedAt:     timestamppb.Now(),
		ExpiresAt:      timestamppb.New(time.Now().Add(time.Hour)),
	}
	suite.Require().NoError(suite.Keeper.SetKYCRecord(ctx, validRecord))

	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Invalid address
	invalidRecord := &types.KYCRecord{
		Address:        "invalid-address",
		VerificationId: "kyc-456",
		KycLevel:       types.KYCLevel_KYC_LEVEL_BASIC,
		VerifiedAt:     timestamppb.Now(),
	}
	suite.Require().NoError(suite.Keeper.SetKYCRecord(ctx, invalidRecord))

	msg, broken = inv(ctx)
	suite.True(broken)
	suite.Contains(msg, "invalid address")

	// Reset context
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = KYCRecordConsistencyInvariant(suite.Keeper)

	// Empty verification ID
	emptyVerif := &types.KYCRecord{
		Address:        validAddr,
		VerificationId: "",
		KycLevel:       types.KYCLevel_KYC_LEVEL_BASIC,
		VerifiedAt:     timestamppb.Now(),
	}
	suite.Require().NoError(suite.Keeper.SetKYCRecord(ctx, emptyVerif))

	msg, broken = inv(ctx)
	suite.True(broken)
	suite.Contains(msg, "empty verification ID")

	// Reset context
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = KYCRecordConsistencyInvariant(suite.Keeper)

	// Invalid level
	badLevel := &types.KYCRecord{
		Address:        validAddr,
		VerificationId: "kyc-789",
		KycLevel:       types.KYCLevel(99),
		VerifiedAt:     timestamppb.Now(),
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
		Address:        validAddr,
		VerificationId: "kyc-101",
		KycLevel:       types.KYCLevel_KYC_LEVEL_BASIC,
	}
	suite.Require().NoError(suite.Keeper.SetKYCRecord(ctx, nilVerified))

	msg, broken = inv(ctx)
	suite.True(broken)
	suite.Contains(msg, "nil verified_at")
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
		ScreenedAt: timestamppb.Now(),
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
		ScreenedAt: timestamppb.Now(),
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

	// Test: nil screened_at fails
	validAddr2 := suite.addr("sanctions-niltime")
	nilTimeResult := &types.SanctionsScreeningResult{
		Address:    validAddr2,
		Status:     types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:    []*types.SanctionsMatch{},
		ScreenedAt: nil,
	}
	bz, err = suite.Keeper.cdc.Marshal(nilTimeResult)
	suite.Require().NoError(err)
	store = ctx.KVStore(suite.Keeper.storeKey)
	key = append(SanctionsResultsKeyPrefix, []byte(validAddr2)...)
	store.Set(key, bz)

	msg, broken = inv(ctx)
	suite.True(broken, "nil screened_at should break invariant")
	suite.Contains(msg, "nil screened_at")

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
		ScreenedAt: timestamppb.Now(),
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

	// Test: valid GDPR request passes
	validAddr := suite.addr("gdpr-valid")
	validRequest := &types.GDPRRequest{
		RequesterAddress: validAddr,
		RequestType:      "access",
		Status:           "pending",
		SubmittedAt:      timestamppb.Now(),
		ProcessedAt:      nil,
	}
	suite.Require().NoError(suite.Keeper.SetGDPRRequest(ctx, validRequest))

	inv := GDPRDataIntegrityInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "valid GDPR request should pass")
	suite.Empty(msg)

	// Test: invalid requester address fails
	invalidRequest := &types.GDPRRequest{
		RequesterAddress: "invalid-address",
		RequestType:      "access",
		Status:           "pending",
		SubmittedAt:      timestamppb.Now(),
	}
	suite.Require().NoError(suite.Keeper.SetGDPRRequest(ctx, invalidRequest))

	msg, broken = inv(ctx)
	suite.True(broken, "invalid requester address should break invariant")
	suite.Contains(msg, "invalid requester address")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = GDPRDataIntegrityInvariant(suite.Keeper)

	// Test: invalid request type fails
	invalidTypeRequest := &types.GDPRRequest{
		RequesterAddress: validAddr,
		RequestType:      "invalid-type",
		Status:           "pending",
		SubmittedAt:      timestamppb.Now(),
	}
	suite.Require().NoError(suite.Keeper.SetGDPRRequest(ctx, invalidTypeRequest))

	msg, broken = inv(ctx)
	suite.True(broken, "invalid request type should break invariant")
	suite.Contains(msg, "invalid type")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = GDPRDataIntegrityInvariant(suite.Keeper)

	// Test: nil submitted_at fails
	nilTimeRequest := &types.GDPRRequest{
		RequesterAddress: validAddr,
		RequestType:      "access",
		Status:           "pending",
		SubmittedAt:      nil,
	}
	suite.Require().NoError(suite.Keeper.SetGDPRRequest(ctx, nilTimeRequest))

	msg, broken = inv(ctx)
	suite.True(broken, "nil submitted_at should break invariant")
	suite.Contains(msg, "nil submitted_at")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = GDPRDataIntegrityInvariant(suite.Keeper)

	// Test: processed without processed_at fails
	processedNoTime := &types.GDPRRequest{
		RequesterAddress: validAddr,
		RequestType:      "access",
		Status:           "processed",
		SubmittedAt:      timestamppb.Now(),
		ProcessedAt:      nil,
	}
	suite.Require().NoError(suite.Keeper.SetGDPRRequest(ctx, processedNoTime))

	msg, broken = inv(ctx)
	suite.True(broken, "processed without processed_at should break invariant")
	suite.Contains(msg, "nil processed_at")
}

func (suite *InvariantsTestSuite) TestTaxRecordConsistencyInvariant() {
	ctx := suite.SdkCtx

	// Test: valid tax record passes
	validAddr := suite.addr("tax-valid")
	validRecord := &types.TaxRecord{
		Address:      validAddr,
		TaxYear:      2024,
		Jurisdiction: "US",
		TotalIncome:  "1000.00",
		TotalTax:     "200.00",
	}
	suite.Require().NoError(suite.Keeper.SetTaxRecord(ctx, validRecord))

	inv := TaxRecordConsistencyInvariant(suite.Keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "valid tax record should pass")
	suite.Empty(msg)

	// Test: invalid address fails
	invalidRecord := &types.TaxRecord{
		Address:      "invalid-address",
		TaxYear:      2024,
		Jurisdiction: "US",
		TotalIncome:  "1000.00",
		TotalTax:     "200.00",
	}
	suite.Require().NoError(suite.Keeper.SetTaxRecord(ctx, invalidRecord))

	msg, broken = inv(ctx)
	suite.True(broken, "invalid address should break invariant")
	suite.Contains(msg, "invalid address")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = TaxRecordConsistencyInvariant(suite.Keeper)

	// Test: invalid tax year fails
	invalidYearRecord := &types.TaxRecord{
		Address:      validAddr,
		TaxYear:      1900,
		Jurisdiction: "US",
		TotalIncome:  "1000.00",
		TotalTax:     "200.00",
	}
	suite.Require().NoError(suite.Keeper.SetTaxRecord(ctx, invalidYearRecord))

	msg, broken = inv(ctx)
	suite.True(broken, "invalid tax year should break invariant")
	suite.Contains(msg, "invalid year")

	// Clean up
	suite.SetupTest()
	ctx = suite.SdkCtx
	inv = TaxRecordConsistencyInvariant(suite.Keeper)

	// Test: empty jurisdiction fails
	emptyJurisdiction := &types.TaxRecord{
		Address:      validAddr,
		TaxYear:      2024,
		Jurisdiction: "",
		TotalIncome:  "1000.00",
		TotalTax:     "200.00",
	}
	suite.Require().NoError(suite.Keeper.SetTaxRecord(ctx, emptyJurisdiction))

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
