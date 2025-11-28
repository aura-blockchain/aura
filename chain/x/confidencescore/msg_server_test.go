package confidencescore

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/confidencescore/keeper"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMsgServer_RecordIRCompletion(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())
	k.SetCurrentHeight(100)

	// Setup IR registry
	registry := &mockIRRegistry{
		activeIRs: make(map[string]bool),
		irScores:  make(map[string]uint64),
		irArenas:  make(map[string]string),
	}
	registry.activeIRs["IR-000"] = true
	registry.irScores["IR-000"] = 500
	k.SetIRRegistry(registry)

	server := keeper.NewMsgServer(k)
	ctx := context.Background()

	proofHash := sha256.Sum256([]byte("proof"))
	verifierHash := sha256.Sum256([]byte("verifier"))

	msg := &confidencescorepb.MsgRecordIRCompletion{
		WalletAddress:    "aura1test",
		IrId:             "IR-000",
		AssistantAddress: "assistant1",
		ProofHash:        proofHash[:],
		VerifierHash:     verifierHash[:],
		Timestamp:        timestamppb.New(time.Now()),
	}

	resp, err := server.RecordIRCompletion(ctx, msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.ScoreEarned == 0 {
		t.Error("expected non-zero score earned")
	}

	if resp.NewTotalScore != resp.ScoreEarned {
		t.Errorf("expected new total to equal score earned for first completion")
	}
}

func TestMsgServer_RecordIRCompletion_NilMsg(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	server := keeper.NewMsgServer(k)
	ctx := context.Background()

	_, err := server.RecordIRCompletion(ctx, nil)
	if err == nil {
		t.Error("expected error for nil message")
	}
}

func TestMsgServer_RecordIRCompletion_InvalidInputs(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())

	server := keeper.NewMsgServer(k)
	ctx := context.Background()

	proofHash := sha256.Sum256([]byte("proof"))
	verifierHash := sha256.Sum256([]byte("verifier"))

	tests := []struct {
		name        string
		msg         *confidencescorepb.MsgRecordIRCompletion
		expectError error
	}{
		{
			name: "empty wallet address",
			msg: &confidencescorepb.MsgRecordIRCompletion{
				WalletAddress:    "",
				IrId:             "IR-001",
				AssistantAddress: "assistant1",
				ProofHash:        proofHash[:],
				VerifierHash:     verifierHash[:],
				Timestamp:        timestamppb.New(time.Now()),
			},
			expectError: types.ErrInvalidWalletAddress,
		},
		{
			name: "empty IR ID",
			msg: &confidencescorepb.MsgRecordIRCompletion{
				WalletAddress:    "aura1test",
				IrId:             "",
				AssistantAddress: "assistant1",
				ProofHash:        proofHash[:],
				VerifierHash:     verifierHash[:],
				Timestamp:        timestamppb.New(time.Now()),
			},
			expectError: types.ErrInvalidIRID,
		},
		{
			name: "invalid proof hash",
			msg: &confidencescorepb.MsgRecordIRCompletion{
				WalletAddress:    "aura1test",
				IrId:             "IR-001",
				AssistantAddress: "assistant1",
				ProofHash:        []byte("short"),
				VerifierHash:     verifierHash[:],
				Timestamp:        timestamppb.New(time.Now()),
			},
			expectError: types.ErrInvalidProofHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := server.RecordIRCompletion(ctx, tt.msg)
			if err != tt.expectError {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestMsgServer_RecalculateScore(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	server := keeper.NewMsgServer(k)
	ctx := context.Background()

	walletAddr := "aura1test"

	// Create user with score mismatch
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    5000, // Wrong
		CompletedIrs: []*types.IRCompletion{
			{IrId: "IR-001", FinalScore: 1000},
		},
	}
	k.SetUserRecord(record)

	msg := &confidencescorepb.MsgRecalculateScore{
		Authority:     "gov1",
		WalletAddress: walletAddr,
	}

	resp, err := server.RecalculateScore(ctx, msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.PreviousScore != 5000 {
		t.Errorf("expected previous score 5000, got %d", resp.PreviousScore)
	}

	if resp.RecalculatedScore != 1000 {
		t.Errorf("expected recalculated score 1000, got %d", resp.RecalculatedScore)
	}

	if len(resp.Discrepancies) == 0 {
		t.Error("expected discrepancies to be reported")
	}
}

func TestMsgServer_RecalculateScore_Unauthorized(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	server := keeper.NewMsgServer(k)
	ctx := context.Background()

	msg := &confidencescorepb.MsgRecalculateScore{
		Authority:     "unauthorized",
		WalletAddress: "aura1test",
	}

	_, err := server.RecalculateScore(ctx, msg)
	if err == nil {
		t.Error("expected error for unauthorized authority")
	}
}

func TestMsgServer_SlashScore(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	server := keeper.NewMsgServer(k)
	ctx := context.Background()

	walletAddr := "aura1test"

	// Create user
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
		Status:        types.VerificationStatusVerified,
	}
	k.SetUserRecord(record)

	msg := &confidencescorepb.MsgSlashScore{
		Authority:     "gov1",
		WalletAddress: walletAddr,
		IrId:          "IR-001",
		SlashAmount:   2000,
		Reason:        "fraud_detected",
		Evidence:      "evidence_hash",
	}

	resp, err := server.SlashScore(ctx, msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.PreviousScore != 10000 {
		t.Errorf("expected previous score 10000, got %d", resp.PreviousScore)
	}

	if resp.NewScore != 8000 {
		t.Errorf("expected new score 8000, got %d", resp.NewScore)
	}

	if resp.SlashTxHash == "" {
		t.Error("expected slash tx hash to be set")
	}
}

func TestMsgServer_SlashScore_DifferentReasons(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	server := keeper.NewMsgServer(k)
	ctx := context.Background()

	reasons := []string{"fraud_detected", "false_attestation", "collusion", "unknown"}

	for i, reason := range reasons {
		walletAddr := "aura1test" + string(rune('0'+i))

		record := types.UserConfidenceRecord{
			WalletAddress: walletAddr,
			TotalScore:    10000,
		}
		k.SetUserRecord(record)

		msg := &confidencescorepb.MsgSlashScore{
			Authority:     "gov1",
			WalletAddress: walletAddr,
			IrId:          "IR-001",
			SlashAmount:   1000,
			Reason:        reason,
			Evidence:      "evidence",
		}

		_, err := server.SlashScore(ctx, msg)
		if err != nil {
			t.Errorf("expected no error for reason %s, got %v", reason, err)
		}
	}
}

func TestMsgServer_AppealSlash(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	server := keeper.NewMsgServer(k)
	ctx := context.Background()

	walletAddr := "aura1test"

	// Create and slash user
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
	}
	k.SetUserRecord(record)

	slashMsg := &confidencescorepb.MsgSlashScore{
		Authority:     "gov1",
		WalletAddress: walletAddr,
		IrId:          "IR-001",
		SlashAmount:   2000,
		Reason:        "fraud_detected",
		Evidence:      "evidence",
	}

	slashResp, _ := server.SlashScore(ctx, slashMsg)

	// Appeal the slash
	params := k.GetParams()
	appealMsg := &confidencescorepb.MsgAppealSlash{
		WalletAddress: walletAddr,
		SlashTxHash:   slashResp.SlashTxHash,
		Evidence:      "counter_evidence",
		Deposit:       params.AppealDeposit,
	}

	appealResp, err := server.AppealSlash(ctx, appealMsg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !appealResp.AppealAccepted {
		t.Error("expected appeal to be accepted")
	}
}

func TestMsgServer_ResolveAppeal(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)
	k.SetCurrentTime(time.Now().Unix())

	server := keeper.NewMsgServer(k)
	ctx := context.Background()

	walletAddr := "aura1test"

	// Create, slash, and appeal
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		TotalScore:    10000,
	}
	k.SetUserRecord(record)

	slashMsg := &confidencescorepb.MsgSlashScore{
		Authority:     "gov1",
		WalletAddress: walletAddr,
		IrId:          "IR-001",
		SlashAmount:   2000,
		Reason:        "fraud_detected",
		Evidence:      "evidence",
	}
	slashResp, _ := server.SlashScore(ctx, slashMsg)

	params := k.GetParams()
	appealMsg := &confidencescorepb.MsgAppealSlash{
		WalletAddress: walletAddr,
		SlashTxHash:   slashResp.SlashTxHash,
		Evidence:      "counter_evidence",
		Deposit:       params.AppealDeposit,
	}
	server.AppealSlash(ctx, appealMsg)

	// Resolve appeal - restore score
	resolveMsg := &confidencescorepb.MsgResolveAppeal{
		Authority:       "gov1",
		WalletAddress:   walletAddr,
		SlashTxHash:     slashResp.SlashTxHash,
		RestoreScore:    true,
		ResolutionNotes: "appeal upheld",
	}

	resolveResp, err := server.ResolveAppeal(ctx, resolveMsg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resolveResp.RestoredScore != 2000 {
		t.Errorf("expected restored score 2000, got %d", resolveResp.RestoredScore)
	}

	if !resolveResp.DepositReturned {
		t.Error("expected deposit to be returned")
	}
}

func TestMsgServer_ResolveAppeal_Unauthorized(t *testing.T) {
	k := keeper.NewKeeper(nil, "gov1")
	server := keeper.NewMsgServer(k)
	ctx := context.Background()

	msg := &confidencescorepb.MsgResolveAppeal{
		Authority:       "unauthorized",
		WalletAddress:   "aura1test",
		SlashTxHash:     "hash",
		RestoreScore:    true,
		ResolutionNotes: "notes",
	}

	_, err := server.ResolveAppeal(ctx, msg)
	if err == nil {
		t.Error("expected error for unauthorized authority")
	}
}

// Mock IR Registry for tests
type mockIRRegistry struct {
	activeIRs     map[string]bool
	irScores      map[string]uint64
	irArenas      map[string]string
	prerequisites map[string][]string
}

func (m *mockIRRegistry) GetIRPrerequisites(irID string) ([]string, error) {
	if m.prerequisites == nil {
		return []string{}, nil
	}
	prereqs, ok := m.prerequisites[irID]
	if !ok {
		return []string{}, nil
	}
	return prereqs, nil
}

func (m *mockIRRegistry) IsIRActive(irID string) bool {
	active, ok := m.activeIRs[irID]
	return ok && active
}

func (m *mockIRRegistry) GetIRScore(irID string) (uint64, error) {
	score, ok := m.irScores[irID]
	if !ok {
		return 100, nil
	}
	return score, nil
}

func (m *mockIRRegistry) GetIRArena(irID string) (string, error) {
	arena, ok := m.irArenas[irID]
	if !ok {
		return "Biometric", nil
	}
	return arena, nil
}
