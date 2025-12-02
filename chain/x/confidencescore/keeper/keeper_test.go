package keeper

import (
	"testing"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

func TestNewKeeper(t *testing.T) {
	ctx, keeper := setupConfKeeper(t)

	if keeper == nil {
		t.Fatal("expected non-nil keeper")
	}

	if keeper.paramsStore == nil {
		t.Error("expected params store to be set")
	}
	_ = ctx
}

func TestKeeper_GetSetUserRecord(t *testing.T) {
	ctx, keeper := setupConfKeeperWithTime(t)

	walletAddr := "aura1test"
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    5000,
		HasAnchor:     true,
		Status:        types.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED,
	}

	// Set record
	if err := keeper.SetUserRecord(ctx, record); err != nil {
		t.Fatalf("failed to set user record: %v", err)
	}

	// Get record
	retrieved, ok := keeper.GetUserRecord(ctx, walletAddr)
	if !ok {
		t.Fatal("expected to find user record")
	}

	if retrieved.TotalScore != 5000 {
		t.Errorf("expected score 5000, got %d", retrieved.TotalScore)
	}
}

func TestKeeper_IsVerified(t *testing.T) {
	ctx, keeper := setupConfKeeperWithTime(t)

	walletAddr := "aura1test"

	// User with score below threshold
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    9000,
		Status:        types.VerificationStatus_VERIFICATION_STATUS_UNVERIFIED,
	}
	keeper.SetUserRecord(ctx, record)

	if keeper.IsVerified(ctx, walletAddr) {
		t.Error("expected user to not be verified")
	}

	// User with score at threshold
	record.TotalScore = 10000
	record.Status = types.VerificationStatus_VERIFICATION_STATUS_VERIFIED
	keeper.SetUserRecord(ctx, record)

	if !keeper.IsVerified(ctx, walletAddr) {
		t.Error("expected user to be verified")
	}
}

func TestKeeper_HasCompletedIR(t *testing.T) {
	ctx, keeper := setupConfKeeperWithTime(t)

	walletAddr := "aura1test"
	irID := "IR-100"

	// User hasn't completed IR
	if keeper.HasCompletedIR(ctx, walletAddr, irID) {
		t.Error("expected IR to not be completed")
	}

	// Complete IR
	completion := types.IRCompletion{
		IrId:       irID,
		FinalScore: 500,
	}
	keeper.SetIRCompletion(ctx, walletAddr, completion)

	if !keeper.HasCompletedIR(ctx, walletAddr, irID) {
		t.Error("expected IR to be completed")
	}
}

func TestKeeper_InitExportGenesis(t *testing.T) {
	ctx, keeper := setupConfKeeperWithTime(t)

	// Create test genesis state
	defaultParams := types.DefaultParams()
	genesis := types.GenesisState{
		Params:       &defaultParams,
		UserRecords:  []*types.UserConfidenceRecord{},
		SlashRecords: []*types.SlashRecord{},
	}

	// Init genesis
	if err := keeper.InitGenesis(ctx, genesis); err != nil {
		t.Fatalf("failed to init genesis: %v", err)
	}

	// Export genesis
	exported := keeper.ExportGenesis(ctx)

	if len(exported.UserRecords) != len(genesis.UserRecords) {
		t.Errorf("expected %d user records, got %d",
			len(genesis.UserRecords), len(exported.UserRecords))
	}
}

func TestKeeper_GetArenaScore(t *testing.T) {
	ctx, keeper := setupConfKeeperWithTime(t)

	walletAddr := "aura1test"
	arena := "Biometric"

	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		ArenaScores: map[string]*types.ArenaScore{
			arena: {
				ArenaType:  arena,
				TotalScore: 3000,
				IrCount:    5,
			},
		},
	}
	keeper.SetUserRecord(ctx, record)

	score, err := keeper.GetArenaScore(ctx, walletAddr, arena)
	if err != nil {
		t.Fatalf("failed to get arena score: %v", err)
	}

	if score != 3000 {
		t.Errorf("expected arena score 3000, got %d", score)
	}
}
