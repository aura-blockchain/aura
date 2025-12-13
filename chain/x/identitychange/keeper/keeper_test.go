package keeper_test

import (
	"testing"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/types"
	"github.com/stretchr/testify/require"
)

func TestCreateRequestRateLimit(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		nil,
		"authority",
		keepertest.Logger(),
	)

	// Set custom params with rate limit of 1
	customParams := types.DefaultParams()
	customParams.MaxRequestsPerWalletPerMonth = 1
	require.NoError(t, k.SetParams(customParams))

	// Create first request - should succeed
	req := types.IdentityChangeRequest{
		RequestId:       "r1",
		Requester:       "aura1requester",
		TargetDid:       "did:alice",
		IrId:            "ir1",
		ProofHash:       "proof1",
		RequestMetaHash: "meta1",
		Status:          types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
		CreatedHeight:   1000,
	}
	if _, err := k.CreateRequest(input.Ctx, req); err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	// Create second request - should fail due to rate limit
	req2 := types.IdentityChangeRequest{
		RequestId:       "r2",
		Requester:       "aura1requester",
		TargetDid:       "did:alice",
		IrId:            "ir2",
		ProofHash:       "proof2",
		RequestMetaHash: "meta2",
		Status:          types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
		CreatedHeight:   1001,
	}
	if _, err := k.CreateRequest(input.Ctx, req2); err == nil {
		t.Fatalf("expected limit error")
	}
}

func TestApplyChangeEnforcesMinConfidence(t *testing.T) {
	input := keepertest.CreateTestInput(t)
	k := keeper.NewKeeper(
		keepertest.WrapStoreService(input.StoreKey),
		input.Cdc,
		nil,
		"authority",
		keepertest.Logger(),
	)

	// Set custom params with min confidence of 100
	customParams := types.DefaultParams()
	customParams.MinConfidenceAfterChange = 100
	require.NoError(t, k.SetParams(customParams))

	// Create and prepare request
	req := types.IdentityChangeRequest{
		RequestId:       "r-ready",
		Requester:       "aura1requester",
		TargetDid:       "did:bob",
		IrId:            "ir1",
		ProofHash:       "hash",
		RequestMetaHash: "meta",
		Status:          types.IdentityChangeStatus_IDENTITY_CHANGE_STATUS_PENDING_VERIFICATION,
		CreatedHeight:   1000,
	}

	// Create the request
	created, err := k.CreateRequest(input.Ctx, req)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	// Submit proof to mark it ready to apply
	if _, err := k.SubmitProof(input.Ctx, created.RequestId, "aura1assistant", true, 0, ""); err != nil {
		t.Fatalf("submit proof: %v", err)
	}

	// Apply the change
	record, err := k.ApplyChange(input.Ctx, created.RequestId)
	if err != nil {
		t.Fatalf("apply change: %v", err)
	}

	// Verify min confidence is enforced
	if record.ConfidenceScore != 100 {
		t.Fatalf("expected min confidence 100, got %d", record.ConfidenceScore)
	}
}
