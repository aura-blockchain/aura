package keeper

import (
	"testing"

	"github.com/aequitas/aura/chain/x/confidencescore/params"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

func TestNewKeeper(t *testing.T) {
	store := params.NewStore(types.DefaultParams())
	keeper := NewKeeper(store)

	if keeper == nil {
		t.Fatal("expected non-nil keeper")
	}

	if keeper.paramsStore == nil {
		t.Error("expected params store to be set")
	}
}

func TestKeeper_GetSetUserRecord(t *testing.T) {
	keeper := NewKeeper(nil)

	walletAddr := "aura1test"
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    5000,
		HasAnchor:     true,
		Status:        types.VerificationStatusUnverified,
	}

	// Set record
	if err := keeper.SetUserRecord(record); err != nil {
		t.Fatalf("failed to set user record: %v", err)
	}

	// Get record
	retrieved, ok := keeper.GetUserRecord(walletAddr)
	if !ok {
		t.Fatal("expected to find user record")
	}

	if retrieved.TotalScore != 5000 {
		t.Errorf("expected score 5000, got %d", retrieved.TotalScore)
	}
}

func TestKeeper_IsVerified(t *testing.T) {
	keeper := NewKeeper(nil)

	walletAddr := "aura1test"

	// User with score below threshold
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    9000,
		Status:        types.VerificationStatusUnverified,
	}
	keeper.SetUserRecord(record)

	if keeper.IsVerified(walletAddr) {
		t.Error("expected user to not be verified")
	}

	// User with score at threshold
	record.TotalScore = 10000
	record.Status = types.VerificationStatusVerified
	keeper.SetUserRecord(record)

	if !keeper.IsVerified(walletAddr) {
		t.Error("expected user to be verified")
	}
}

func TestKeeper_HasCompletedIR(t *testing.T) {
	keeper := NewKeeper(nil)

	walletAddr := "aura1test"
	irID := "IR-100"

	// User hasn't completed IR
	if keeper.HasCompletedIR(walletAddr, irID) {
		t.Error("expected IR to not be completed")
	}

	// Complete IR
	completion := types.IRCompletion{
		IRID:       irID,
		FinalScore: 500,
	}
	keeper.SetIRCompletion(walletAddr, completion)

	if !keeper.HasCompletedIR(walletAddr, irID) {
		t.Error("expected IR to be completed")
	}
}

func TestKeeper_InitExportGenesis(t *testing.T) {
	keeper := NewKeeper(nil)

	// Create test genesis
	genesis := types.TestGenesisState()

	// Init genesis
	if err := keeper.InitGenesis(genesis); err != nil {
		t.Fatalf("failed to init genesis: %v", err)
	}

	// Export genesis
	exported := keeper.ExportGenesis()

	if len(exported.UserRecords) != len(genesis.UserRecords) {
		t.Errorf("expected %d user records, got %d",
			len(genesis.UserRecords), len(exported.UserRecords))
	}
}

func TestKeeper_GetArenaScore(t *testing.T) {
	keeper := NewKeeper(nil)

	walletAddr := "aura1test"
	arena := "Biometric"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		ArenaScores: map[string]types.ArenaScore{
			arena: {
				ArenaType:  arena,
				TotalScore: 3000,
				IRCount:    5,
			},
		},
	}
	keeper.SetUserRecord(record)

	score, err := keeper.GetArenaScore(walletAddr, arena)
	if err != nil {
		t.Fatalf("failed to get arena score: %v", err)
	}

	if score != 3000 {
		t.Errorf("expected arena score 3000, got %d", score)
	}
}
