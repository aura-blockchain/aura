package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

type InvariantsComprehensiveTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsComprehensiveTestSuite))
}

func (suite *InvariantsComprehensiveTestSuite) TestKYCRecordConsistencyInvariant() {
	ctx := suite.SdkCtx
	inv := KYCRecordConsistencyInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Create valid KYC record
	addr := sdk.AccAddress("test_address______")
	record := types.KYCRecord{
		Address:        addr.String(),
		VerificationId: "kyc-12345",
		Status:         "approved",
		Level:          "basic",
		VerifiedAt:     "2024-01-01T00:00:00Z",
	}
	suite.storeKYCRecord(ctx, &record)

	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestKYCRecordInvalidAddress() {
	ctx := suite.SdkCtx
	inv := KYCRecordConsistencyInvariant(suite.Keeper)

	// Create record with invalid address
	record := types.KYCRecord{
		Address:        "invalid-address",
		VerificationId: "kyc-12345",
		Status:         "approved",
		Level:          "basic",
	}
	suite.storeKYCRecord(ctx, &record)

	msg, broken := inv(ctx)
	suite.True(broken, "record with invalid address should break invariant")
	suite.Contains(msg, "invalid address")
}

func (suite *InvariantsComprehensiveTestSuite) TestKYCRecordEmptyVerificationID() {
	ctx := suite.SdkCtx
	inv := KYCRecordConsistencyInvariant(suite.Keeper)

	addr := sdk.AccAddress("test_address______")
	record := types.KYCRecord{
		Address:        addr.String(),
		VerificationId: "",
		Status:         "approved",
		Level:          "basic",
	}
	suite.storeKYCRecord(ctx, &record)

	msg, broken := inv(ctx)
	suite.True(broken, "record with empty verification ID should break invariant")
	suite.Contains(msg, "empty verification ID")
}

func (suite *InvariantsComprehensiveTestSuite) TestKYCRecordInvalidStatus() {
	ctx := suite.SdkCtx
	inv := KYCRecordConsistencyInvariant(suite.Keeper)

	addr := sdk.AccAddress("test_address______")
	record := types.KYCRecord{
		Address:        addr.String(),
		VerificationId: "kyc-12345",
		Status:         "invalid-status",
		Level:          "basic",
	}
	suite.storeKYCRecord(ctx, &record)

	msg, broken := inv(ctx)
	suite.True(broken, "record with invalid status should break invariant")
	suite.Contains(msg, "invalid status")
}

func (suite *InvariantsComprehensiveTestSuite) TestSanctionsScreeningInvariant() {
	ctx := suite.SdkCtx
	inv := SanctionsScreeningInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestGDPRDataIntegrityInvariant() {
	ctx := suite.SdkCtx
	inv := GDPRDataIntegrityInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTaxRecordConsistencyInvariant() {
	ctx := suite.SdkCtx
	inv := TaxRecordConsistencyInvariant(suite.Keeper)

	// Test: Empty store should not break invariant
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestAllInvariants() {
	ctx := suite.SdkCtx
	inv := AllInvariants(suite.Keeper)

	// Test: All invariants on empty store
	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	// Add valid data
	addr := sdk.AccAddress("test_address______")
	record := types.KYCRecord{
		Address:        addr.String(),
		VerificationId: "kyc-12345",
		Status:         "approved",
		Level:          "basic",
	}
	suite.storeKYCRecord(ctx, &record)

	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

// Helper methods
func (suite *InvariantsComprehensiveTestSuite) storeKYCRecord(ctx sdk.Context, record *types.KYCRecord) {
	store := ctx.KVStore(suite.Keeper.storeKey)
	bz := suite.Keeper.cdc.MustMarshal(record)
	store.Set(append(types.KYCRecordKeyPrefix, []byte(record.Address)...), bz)
}
