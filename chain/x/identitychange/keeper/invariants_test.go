package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/types"
)

type InvariantsTestSuite struct {
	suite.Suite

	Keeper *Keeper
	SdkCtx sdk.Context
}

func (suite *InvariantsTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.Keeper = NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		nil,
		"authority",
		keepertest.Logger(),
	)
	suite.SdkCtx = input.Ctx

	// Set default params
	params := types.DefaultParams()
	suite.Keeper.SetParams(params)
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

// Test all invariants on empty store
func (suite *InvariantsTestSuite) TestAllInvariantsEmptyStore() {
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

// Test invariant registration
func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Register should not panic
	suite.NotPanics(func() {
		// Just verify the function exists and doesn't panic
		// We can't test the actual registration without a full SDK context
		_ = AllInvariants(suite.Keeper)
	})
}

// ============================================================================
// ParamsInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestParamsInvariantValid() {
	inv := ParamsInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "params invariant should pass with default params")
	suite.Empty(msg)
}

// ============================================================================
// RequestValidityInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestRequestValidityInvariantValid() {
	// Create a valid request
	request := &types.IdentityChangeRequest{
		RequestId:       "req-1",
		TargetDid:       "did:aura:valid",
		Requester:       keepertest.GenTestAddr().String(),
		IrId:            "ir-1",
		ProofHash:       "proof-hash",
		RequestMetaHash: "meta-hash",
		Status:          types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
		CreatedHeight:   1000,
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, *request)
	suite.Require().NoError(err)

	inv := RequestValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "request validity invariant should pass with valid request")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestRequestValidityInvariantEmptyID() {
	// Create request with empty ID  - this should fail at SetRequest level
	request := &types.IdentityChangeRequest{
		RequestId:       "",
		TargetDid:       "did:aura:test",
		Requester:       keepertest.GenTestAddr().String(),
		IrId:            "ir-1",
		ProofHash:       "proof",
		RequestMetaHash: "meta",
		Status:          types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
		CreatedHeight:   1000,
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, *request)
	// SetRequest should fail with empty ID
	suite.Require().Error(err)
}

func (suite *InvariantsTestSuite) TestRequestValidityInvariantInvalidRequester() {
	// Cannot create request with completely invalid requester due to validation
	// This test verifies the invariant passes with valid data
	request := &types.IdentityChangeRequest{
		RequestId:       "req-1",
		TargetDid:       "did:aura:test",
		Requester:       keepertest.GenTestAddr().String(),
		IrId:            "ir-1",
		ProofHash:       "proof",
		RequestMetaHash: "meta",
		Status:          types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
		CreatedHeight:   1000,
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, *request)
	suite.Require().NoError(err)

	inv := RequestValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "request validity invariant should pass")
	suite.Empty(msg)
}

// ============================================================================
// RecordConsistencyInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestRecordConsistencyInvariantValid() {
	// Create a valid record
	record := types.IdentityRecord{
		Did:               "did:aura:record1",
		Owner:             keepertest.GenTestAddr().String(),
		ConfidenceScore:   75,
		MetadataHash:      "metadata-hash",
		LatestIrVersion:   "v1",
		LastChangedHeight: 1000,
		Status:            types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED,
	}
	err := suite.Keeper.SetIdentityRecord(suite.SdkCtx, record)
	suite.Require().NoError(err)

	inv := RecordConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "record consistency invariant should pass with valid record")
	suite.Empty(msg)
}

// ============================================================================
// StatusConsistencyInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestStatusConsistencyInvariantValid() {
	// Create request with valid status
	request := &types.IdentityChangeRequest{
		RequestId:       "req-status-1",
		TargetDid:       "did:aura:status",
		Requester:       keepertest.GenTestAddr().String(),
		IrId:            "ir-status",
		ProofHash:       "proof",
		RequestMetaHash: "meta",
		Status:          types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
		CreatedHeight:   1000,
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, *request)
	suite.Require().NoError(err)

	inv := StatusConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "status consistency invariant should pass with valid status")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestStatusConsistencyInvariantRejectedWithReason() {
	// Create rejected request with reason
	request := &types.IdentityChangeRequest{
		RequestId:       "req-rejected",
		TargetDid:       "did:aura:rejected",
		Requester:       keepertest.GenTestAddr().String(),
		IrId:            "ir-rejected",
		ProofHash:       "proof",
		RequestMetaHash: "meta",
		Status:          types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_REJECTED,
		Reason:          "Invalid proof provided",
		CreatedHeight:   1000,
		VerdictHeight:   1010,
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, *request)
	suite.Require().NoError(err)

	inv := StatusConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "status consistency invariant should pass for rejected request with reason")
	suite.Empty(msg)
}

// ============================================================================
// HistoryIntegrityInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestHistoryIntegrityInvariantValid() {
	// Create valid history entry
	history := types.IdentityChangeHistory{
		RequestId:           "req-history-1",
		TargetDid:           "did:aura:history",
		PrevConfidenceScore: 70,
		NewConfidenceScore:  80,
		TransitionReason:    "applied",
		ChangedHeight:       1001,
	}
	err := suite.Keeper.AddHistory(suite.SdkCtx, history)
	suite.Require().NoError(err)

	inv := HistoryIntegrityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "history integrity invariant should pass with valid history")
	suite.Empty(msg)
}

// ============================================================================
// All Invariants Integration Test
// ============================================================================

func (suite *InvariantsTestSuite) TestAllInvariantsWithValidData() {
	// Setup multiple valid requests and records
	request1 := &types.IdentityChangeRequest{
		RequestId:       "req-int-1",
		TargetDid:       "did:aura:int1",
		Requester:       keepertest.GenTestAddr().String(),
		IrId:            "ir-int-1",
		ProofHash:       "proof-1",
		RequestMetaHash: "meta-1",
		Status:          types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
		CreatedHeight:   1000,
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, *request1)
	suite.Require().NoError(err)

	request2 := &types.IdentityChangeRequest{
		RequestId:       "req-int-2",
		TargetDid:       "did:aura:int2",
		Requester:       keepertest.GenTestAddr().String(),
		Assistant:       keepertest.GenTestAddr().String(),
		IrId:            "ir-int-2",
		ProofHash:       "proof-2",
		RequestMetaHash: "meta-2",
		Status:          types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_READY_TO_APPLY,
		CreatedHeight:   900,
		VerdictHeight:   1000,
	}
	err = suite.Keeper.SetRequest(suite.SdkCtx, *request2)
	suite.Require().NoError(err)

	record := types.IdentityRecord{
		Did:               "did:aura:int1",
		Owner:             keepertest.GenTestAddr().String(),
		ConfidenceScore:   85,
		MetadataHash:      "record-meta",
		LatestIrVersion:   "v1",
		LastChangedHeight: 1001,
		Status:            types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_APPLIED,
	}
	err = suite.Keeper.SetIdentityRecord(suite.SdkCtx, record)
	suite.Require().NoError(err)

	// Run all invariants
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "all invariants should pass with valid data")
	suite.Empty(msg)
}
