package keeper_test

import (
	"testing"

	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/params"
	"github.com/aequitas/aura/chain/x/identitychange/types"
)

func TestCreateRequestRateLimit(t *testing.T) {
	customParams := types.DefaultParams()
	customParams.MaxRequestsPerWalletPerMonth = 1
	store := params.NewStore(customParams)
	k := keeper.NewKeeper(store)
	req := types.IdentityChangeRequest{RequestId: "r1", Requester: "alice", TargetDid: "did:alice"}
	if _, err := k.CreateRequest(req); err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if _, err := k.CreateRequest(types.IdentityChangeRequest{RequestId: "r2", Requester: "alice", TargetDid: "did:alice"}); err == nil {
		t.Fatalf("expected limit error")
	}
}

func TestApplyChangeEnforcesMinConfidence(t *testing.T) {
	customParams := types.DefaultParams()
	customParams.MinConfidenceAfterChange = 100
	store := params.NewStore(customParams)
	k := keeper.NewKeeper(store)
	req := types.IdentityChangeRequest{
		RequestId:       "r-ready",
		Requester:       "bob",
		TargetDid:       "did:bob",
		IrId:            "ir1",
		ProofHash:       "hash",
		RequestMetaHash: "meta",
	}
	created, err := k.CreateRequest(req)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if _, err := k.SubmitProof(created.RequestId, "assistant", true, 0, ""); err != nil {
		t.Fatalf("submit proof: %v", err)
	}
	record, err := k.ApplyChange(created.RequestId)
	if err != nil {
		t.Fatalf("apply change: %v", err)
	}
	if record.ConfidenceScore != 100 {
		t.Fatalf("expected min confidence 100, got %d", record.ConfidenceScore)
	}
}
