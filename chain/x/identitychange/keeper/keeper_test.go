package keeper_test

import (
	"testing"

	"github.com/aequitas/aura/chain/x/identitychange/keeper"
	"github.com/aequitas/aura/chain/x/identitychange/params"
	"github.com/aequitas/aura/chain/x/identitychange/types"
)

func TestCreateRequestRateLimit(t *testing.T) {
	store := params.NewStore(types.Params{
		MaxRequestsPerWalletPerMonth:  1,
		MinConfidenceAfterChange:      1000,
		StalenessHeightThreshold:      1000,
		AssistantSlashOnFalsePositive: true,
		StalenessInvestigatorChain:    "",
	})
	k := keeper.NewKeeper(store)
	req := types.IdentityChangeRequest{RequestID: "r1", Requester: "alice", TargetDID: "did:alice"}
	if _, err := k.CreateRequest(req); err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if _, err := k.CreateRequest(types.IdentityChangeRequest{RequestID: "r2", Requester: "alice", TargetDID: "did:alice"}); err == nil {
		t.Fatalf("expected limit error")
	}
}

func TestApplyChangeEnforcesMinConfidence(t *testing.T) {
	store := params.NewStore(types.Params{
		MaxRequestsPerWalletPerMonth:  5,
		MinConfidenceAfterChange:      1100,
		StalenessHeightThreshold:      1000,
		AssistantSlashOnFalsePositive: true,
		StalenessInvestigatorChain:    "",
	})
	k := keeper.NewKeeper(store)
	req := types.IdentityChangeRequest{
		RequestID:       "r-ready",
		Requester:       "bob",
		TargetDID:       "did:bob",
		IRID:            "ir1",
		ProofHash:       "hash",
		RequestMetaHash: "meta",
	}
	created, err := k.CreateRequest(req)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if _, err := k.SubmitProof(created.RequestID, "assistant", true, 0, ""); err != nil {
		t.Fatalf("submit proof: %v", err)
	}
	record, err := k.ApplyChange(created.RequestID)
	if err != nil {
		t.Fatalf("apply change: %v", err)
	}
	if record.ConfidenceScore != 1100 {
		t.Fatalf("expected min confidence 1100, got %d", record.ConfidenceScore)
	}
}
