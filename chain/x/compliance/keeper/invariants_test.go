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

func (suite *InvariantsTestSuite) addr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	bech32, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	suite.Require().NoError(err)
	return bech32
}
