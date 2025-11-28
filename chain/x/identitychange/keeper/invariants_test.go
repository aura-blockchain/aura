package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

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
	suite.Keeper = NewKeeper(input.Cdc, input.StoreKey)
	suite.SdkCtx = input.Ctx

	// Set default params
	params := types.DefaultParams()
	suite.Keeper.SetParams(suite.SdkCtx, params)
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
	registry := sdk.NewInvariantRegistry()
	suite.NotPanics(func() {
		RegisterInvariants(registry, suite.Keeper)
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

func (suite *InvariantsTestSuite) TestParamsInvariantInvalid() {
	// Set invalid params (negative timeout)
	params := suite.Keeper.GetParams(suite.SdkCtx)
	params.RequestTimeoutSeconds = -100 // Invalid
	suite.Keeper.SetParams(suite.SdkCtx, params)

	inv := ParamsInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "params invariant should fail with invalid params")
	suite.NotEmpty(msg)
}

// ============================================================================
// RequestValidityInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestRequestValidityInvariantValid() {
	// Create a valid request
	request := &types.IdentityChangeRequest{
		RequestId:   "req-1",
		Requester:   keepertest.GenTestAddr().String(),
		OldIdentity: "did:old:identity",
		NewIdentity: "did:new:identity",
		ChangeType:  "key_rotation",
		Status:      "pending",
		RequestTime: timestamppb.Now(),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := RequestValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "request validity invariant should pass with valid request")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestRequestValidityInvariantEmptyID() {
	// Create request with empty ID
	request := &types.IdentityChangeRequest{
		RequestId:   "", // Invalid
		Requester:   keepertest.GenTestAddr().String(),
		OldIdentity: "did:old",
		NewIdentity: "did:new",
		ChangeType:  "key_rotation",
		Status:      "pending",
		RequestTime: timestamppb.Now(),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := RequestValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "request validity invariant should fail with empty ID")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestRequestValidityInvariantInvalidRequester() {
	// Create request with invalid requester address
	request := &types.IdentityChangeRequest{
		RequestId:   "req-1",
		Requester:   "invalid-address", // Invalid
		OldIdentity: "did:old",
		NewIdentity: "did:new",
		ChangeType:  "key_rotation",
		Status:      "pending",
		RequestTime: timestamppb.Now(),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := RequestValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "request validity invariant should fail with invalid requester")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestRequestValidityInvariantSameIdentity() {
	// Create request with identical old and new identities
	request := &types.IdentityChangeRequest{
		RequestId:   "req-1",
		Requester:   keepertest.GenTestAddr().String(),
		OldIdentity: "did:same", // Same as new
		NewIdentity: "did:same", // Invalid - must be different
		ChangeType:  "key_rotation",
		Status:      "pending",
		RequestTime: timestamppb.Now(),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := RequestValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "request validity invariant should fail when old and new identities are same")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestRequestValidityInvariantInvalidChangeType() {
	// Create request with invalid change type
	request := &types.IdentityChangeRequest{
		RequestId:   "req-1",
		Requester:   keepertest.GenTestAddr().String(),
		OldIdentity: "did:old",
		NewIdentity: "did:new",
		ChangeType:  "invalid-type", // Invalid
		Status:      "pending",
		RequestTime: timestamppb.Now(),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := RequestValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "request validity invariant should fail with invalid change type")
	suite.NotEmpty(msg)
}

// ============================================================================
// ProofConsistencyInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestProofConsistencyInvariantValid() {
	// Create approved request with proof
	request := &types.IdentityChangeRequest{
		RequestId:    "req-1",
		Requester:    keepertest.GenTestAddr().String(),
		OldIdentity:  "did:old",
		NewIdentity:  "did:new",
		ChangeType:   "key_rotation",
		Status:       "approved",
		Proof:        []byte("valid-proof-data"),
		ApprovedBy:   keepertest.GenTestAddr().String(),
		ApprovalTime: timestamppb.Now(),
		RequestTime:  timestamppb.New(time.Now().Add(-1 * time.Hour)),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := ProofConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "proof consistency invariant should pass with valid approved request")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestProofConsistencyInvariantApprovedWithoutProof() {
	// Create approved request without proof
	request := &types.IdentityChangeRequest{
		RequestId:    "req-1",
		Requester:    keepertest.GenTestAddr().String(),
		OldIdentity:  "did:old",
		NewIdentity:  "did:new",
		ChangeType:   "key_rotation",
		Status:       "approved",
		Proof:        nil, // Invalid for approved
		ApprovedBy:   keepertest.GenTestAddr().String(),
		ApprovalTime: timestamppb.Now(),
		RequestTime:  timestamppb.Now(),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := ProofConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "proof consistency invariant should fail for approved request without proof")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestProofConsistencyInvariantExecutedBeforeApproval() {
	// Create request executed before approval
	now := time.Now()
	request := &types.IdentityChangeRequest{
		RequestId:    "req-1",
		Requester:    keepertest.GenTestAddr().String(),
		OldIdentity:  "did:old",
		NewIdentity:  "did:new",
		ChangeType:   "key_rotation",
		Status:       "executed",
		Proof:        []byte("proof"),
		ApprovedBy:   keepertest.GenTestAddr().String(),
		ApprovalTime: timestamppb.New(now),
		ExecutedAt:   timestamppb.New(now.Add(-1 * time.Hour)), // Before approval - invalid
		RequestTime:  timestamppb.New(now.Add(-2 * time.Hour)),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := ProofConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "proof consistency invariant should fail when executed before approval")
	suite.NotEmpty(msg)
}

// ============================================================================
// StatusConsistencyInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestStatusConsistencyInvariantValid() {
	// Create request with valid status
	request := &types.IdentityChangeRequest{
		RequestId:   "req-1",
		Requester:   keepertest.GenTestAddr().String(),
		OldIdentity: "did:old",
		NewIdentity: "did:new",
		ChangeType:  "key_rotation",
		Status:      "pending",
		RequestTime: timestamppb.Now(),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := StatusConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "status consistency invariant should pass with valid status")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestStatusConsistencyInvariantInvalidStatus() {
	// Create request with invalid status
	request := &types.IdentityChangeRequest{
		RequestId:   "req-1",
		Requester:   keepertest.GenTestAddr().String(),
		OldIdentity: "did:old",
		NewIdentity: "did:new",
		ChangeType:  "key_rotation",
		Status:      "invalid-status", // Invalid
		RequestTime: timestamppb.Now(),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := StatusConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "status consistency invariant should fail with invalid status")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestStatusConsistencyInvariantRejectedWithoutReason() {
	// Create rejected request without reason
	request := &types.IdentityChangeRequest{
		RequestId:       "req-1",
		Requester:       keepertest.GenTestAddr().String(),
		OldIdentity:     "did:old",
		NewIdentity:     "did:new",
		ChangeType:      "key_rotation",
		Status:          "rejected",
		RejectionReason: "", // Invalid for rejected
		RequestTime:     timestamppb.Now(),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := StatusConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "status consistency invariant should fail for rejected request without reason")
	suite.NotEmpty(msg)
}

// ============================================================================
// TimelineIntegrityInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestTimelineIntegrityInvariantValid() {
	// Create request with valid timeline
	now := time.Now()
	request := &types.IdentityChangeRequest{
		RequestId:    "req-1",
		Requester:    keepertest.GenTestAddr().String(),
		OldIdentity:  "did:old",
		NewIdentity:  "did:new",
		ChangeType:   "key_rotation",
		Status:       "approved",
		Proof:        []byte("proof"),
		RequestTime:  timestamppb.New(now.Add(-2 * time.Hour)),
		ApprovalTime: timestamppb.New(now.Add(-1 * time.Hour)),
		ExpiryTime:   timestamppb.New(now.Add(24 * time.Hour)),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := TimelineIntegrityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "timeline integrity invariant should pass with valid timeline")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestTimelineIntegrityInvariantApprovedBeforeRequest() {
	// Create request approved before requested
	now := time.Now()
	request := &types.IdentityChangeRequest{
		RequestId:    "req-1",
		Requester:    keepertest.GenTestAddr().String(),
		OldIdentity:  "did:old",
		NewIdentity:  "did:new",
		ChangeType:   "key_rotation",
		Status:       "approved",
		Proof:        []byte("proof"),
		RequestTime:  timestamppb.New(now),
		ApprovalTime: timestamppb.New(now.Add(-1 * time.Hour)), // Before request - invalid
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := TimelineIntegrityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "timeline integrity invariant should fail when approved before requested")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestTimelineIntegrityInvariantExecutedBeforeRequest() {
	// Create request executed before requested
	now := time.Now()
	request := &types.IdentityChangeRequest{
		RequestId:   "req-1",
		Requester:   keepertest.GenTestAddr().String(),
		OldIdentity: "did:old",
		NewIdentity: "did:new",
		ChangeType:  "key_rotation",
		Status:      "executed",
		Proof:       []byte("proof"),
		RequestTime: timestamppb.New(now),
		ExecutedAt:  timestamppb.New(now.Add(-1 * time.Hour)), // Before request - invalid
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := TimelineIntegrityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "timeline integrity invariant should fail when executed before requested")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestTimelineIntegrityInvariantExpiryBeforeRequest() {
	// Create request with expiry before request time
	now := time.Now()
	request := &types.IdentityChangeRequest{
		RequestId:   "req-1",
		Requester:   keepertest.GenTestAddr().String(),
		OldIdentity: "did:old",
		NewIdentity: "did:new",
		ChangeType:  "key_rotation",
		Status:      "pending",
		RequestTime: timestamppb.New(now),
		ExpiryTime:  timestamppb.New(now.Add(-1 * time.Hour)), // Before request - invalid
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request)
	suite.Require().NoError(err)

	inv := TimelineIntegrityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "timeline integrity invariant should fail when expiry is before request")
	suite.NotEmpty(msg)
}

// ============================================================================
// All Invariants Integration Test
// ============================================================================

func (suite *InvariantsTestSuite) TestAllInvariantsWithValidData() {
	// Setup multiple valid requests
	now := time.Now()

	request1 := &types.IdentityChangeRequest{
		RequestId:   "req-1",
		Requester:   keepertest.GenTestAddr().String(),
		OldIdentity: "did:old:1",
		NewIdentity: "did:new:1",
		ChangeType:  "key_rotation",
		Status:      "pending",
		RequestTime: timestamppb.New(now),
	}
	err := suite.Keeper.SetRequest(suite.SdkCtx, request1)
	suite.Require().NoError(err)

	request2 := &types.IdentityChangeRequest{
		RequestId:    "req-2",
		Requester:    keepertest.GenTestAddr().String(),
		OldIdentity:  "did:old:2",
		NewIdentity:  "did:new:2",
		ChangeType:   "wallet_migration",
		Status:       "approved",
		Proof:        []byte("proof-data"),
		ApprovedBy:   keepertest.GenTestAddr().String(),
		RequestTime:  timestamppb.New(now.Add(-1 * time.Hour)),
		ApprovalTime: timestamppb.New(now),
	}
	err = suite.Keeper.SetRequest(suite.SdkCtx, request2)
	suite.Require().NoError(err)

	// Run all invariants
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "all invariants should pass with valid data")
	suite.Empty(msg)
}
