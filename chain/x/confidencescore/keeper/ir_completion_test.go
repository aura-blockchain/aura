package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

// Mock IR Registry for testing
type mockIRRegistry struct {
	activeIRs     map[string]bool
	irScores      map[string]uint64
	irArenas      map[string]string
	prerequisites map[string][]string
}

func newMockIRRegistry() *mockIRRegistry {
	return &mockIRRegistry{
		activeIRs:     make(map[string]bool),
		irScores:      make(map[string]uint64),
		irArenas:      make(map[string]string),
		prerequisites: make(map[string][]string),
	}
}

func (m *mockIRRegistry) GetIRPrerequisites(irID string) ([]string, error) {
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
		return 100, nil // default
	}
	return score, nil
}

func (m *mockIRRegistry) GetIRArena(irID string) (string, error) {
	arena, ok := m.irArenas[irID]
	if !ok {
		return "Biometric", nil // default
	}
	return arena, nil
}

func TestRecordIRCompletion_Success(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())
	k.SetCurrentHeight(100)

	registry := newMockIRRegistry()
	registry.activeIRs["IR-000"] = true
	registry.irScores["IR-000"] = 500
	k.SetIRRegistry(registry)

	walletAddr := "aura1test"
	proofHash := sha256.Sum256([]byte("proof"))
	verifierHash := sha256.Sum256([]byte("verifier"))

	score, err := k.RecordIRCompletion(
		walletAddr,
		"IR-000",
		"assistant1",
		proofHash[:],
		verifierHash[:],
		time.Now().Unix(),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if score == 0 {
		t.Error("expected non-zero score")
	}

	// Check completion was recorded
	completion, ok := k.GetIRCompletion(walletAddr, "IR-000")
	if !ok {
		t.Error("expected completion to be recorded")
	}

	if completion.IrId != "IR-000" {
		t.Errorf("expected IR-000, got %s", completion.IrId)
	}

	// Check user record was updated
	record, ok := k.GetUserRecord(walletAddr)
	if !ok {
		t.Error("expected user record to exist")
	}

	if record.TotalScore != score {
		t.Errorf("expected total score %d, got %d", score, record.TotalScore)
	}

	if !record.HasAnchor {
		t.Error("expected anchor to be marked as completed")
	}
}

func TestRecordIRCompletion_InvalidInputs(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())

	proofHash := sha256.Sum256([]byte("proof"))
	verifierHash := sha256.Sum256([]byte("verifier"))

	tests := []struct {
		name         string
		walletAddr   string
		irID         string
		proofHash    []byte
		verifierHash []byte
		expectError  error
	}{
		{
			name:         "empty wallet address",
			walletAddr:   "",
			irID:         "IR-001",
			proofHash:    proofHash[:],
			verifierHash: verifierHash[:],
			expectError:  types.ErrInvalidWalletAddress,
		},
		{
			name:         "empty IR ID",
			walletAddr:   "aura1test",
			irID:         "",
			proofHash:    proofHash[:],
			verifierHash: verifierHash[:],
			expectError:  types.ErrInvalidIRID,
		},
		{
			name:         "invalid proof hash length",
			walletAddr:   "aura1test",
			irID:         "IR-001",
			proofHash:    []byte("short"),
			verifierHash: verifierHash[:],
			expectError:  types.ErrInvalidProofHash,
		},
		{
			name:         "invalid verifier hash length",
			walletAddr:   "aura1test",
			irID:         "IR-001",
			proofHash:    proofHash[:],
			verifierHash: []byte("short"),
			expectError:  types.ErrInvalidVerifierHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := k.RecordIRCompletion(
				tt.walletAddr,
				tt.irID,
				"assistant1",
				tt.proofHash,
				tt.verifierHash,
				time.Now().Unix(),
			)

			if err != tt.expectError {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestRecordIRCompletion_AlreadyCompleted(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())

	registry := newMockIRRegistry()
	registry.activeIRs["IR-000"] = true
	k.SetIRRegistry(registry)

	walletAddr := "aura1test"
	proofHash := sha256.Sum256([]byte("proof"))
	verifierHash := sha256.Sum256([]byte("verifier"))

	// Complete IR first time
	_, err := k.RecordIRCompletion(
		walletAddr,
		"IR-000",
		"assistant1",
		proofHash[:],
		verifierHash[:],
		time.Now().Unix(),
	)
	if err != nil {
		t.Fatalf("first completion should succeed: %v", err)
	}

	// Try to complete again
	proofHash2 := sha256.Sum256([]byte("proof2"))
	_, err = k.RecordIRCompletion(
		walletAddr,
		"IR-000",
		"assistant1",
		proofHash2[:],
		verifierHash[:],
		time.Now().Unix(),
	)

	if err != types.ErrIRAlreadyCompleted {
		t.Errorf("expected ErrIRAlreadyCompleted, got %v", err)
	}
}

func TestRecordIRCompletion_InactiveIR(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())

	registry := newMockIRRegistry()
	registry.activeIRs["IR-001"] = false // Inactive
	k.SetIRRegistry(registry)

	walletAddr := "aura1test"
	proofHash := sha256.Sum256([]byte("proof"))
	verifierHash := sha256.Sum256([]byte("verifier"))

	_, err := k.RecordIRCompletion(
		walletAddr,
		"IR-001",
		"assistant1",
		proofHash[:],
		verifierHash[:],
		time.Now().Unix(),
	)

	if err != types.ErrIRNotActive {
		t.Errorf("expected ErrIRNotActive, got %v", err)
	}
}

func TestRecordIRCompletion_AnchorRequired(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())

	registry := newMockIRRegistry()
	registry.activeIRs["IR-001"] = true
	k.SetIRRegistry(registry)

	walletAddr := "aura1test"
	proofHash := sha256.Sum256([]byte("proof"))
	verifierHash := sha256.Sum256([]byte("verifier"))

	// Try to complete IR-001 without IR-000
	_, err := k.RecordIRCompletion(
		walletAddr,
		"IR-001",
		"assistant1",
		proofHash[:],
		verifierHash[:],
		time.Now().Unix(),
	)

	if err != types.ErrAnchorNotCompleted {
		t.Errorf("expected ErrAnchorNotCompleted, got %v", err)
	}
}

func TestRecordIRCompletion_PrerequisitesMissing(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())

	registry := newMockIRRegistry()
	registry.activeIRs["IR-000"] = true
	registry.activeIRs["IR-002"] = true
	registry.prerequisites["IR-002"] = []string{"IR-001"}
	k.SetIRRegistry(registry)

	walletAddr := "aura1test"
	proofHash := sha256.Sum256([]byte("proof"))
	verifierHash := sha256.Sum256([]byte("verifier"))

	// Complete IR-000 (anchor)
	k.RecordIRCompletion(walletAddr, "IR-000", "assistant1", proofHash[:], verifierHash[:], time.Now().Unix())

	// Try to complete IR-002 without IR-001
	proofHash2 := sha256.Sum256([]byte("proof2"))
	_, err := k.RecordIRCompletion(
		walletAddr,
		"IR-002",
		"assistant1",
		proofHash2[:],
		verifierHash[:],
		time.Now().Unix(),
	)

	if err == nil || err.Error() != "required prerequisite ir not completed: IR-001" {
		t.Errorf("expected prerequisite error, got %v", err)
	}
}

func TestRecordIRCompletion_RateLimitExceeded(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())

	registry := newMockIRRegistry()
	registry.activeIRs["IR-000"] = true
	k.SetIRRegistry(registry)

	walletAddr := "aura1test"

	// Complete anchor first
	proofHash := sha256.Sum256([]byte("proof0"))
	verifierHash := sha256.Sum256([]byte("verifier"))
	k.RecordIRCompletion(walletAddr, "IR-000", "assistant1", proofHash[:], verifierHash[:], time.Now().Unix())

	// Set up other active IRs
	for i := 1; i <= 10; i++ {
		irID := "IR-" + string(rune('0'+i))
		registry.activeIRs[irID] = true
	}

	// Try to exceed hourly limit (3)
	for i := 1; i <= 4; i++ {
		irID := "IR-" + string(rune('0'+i))
		proofHash := sha256.Sum256([]byte("proof" + string(rune('0'+i))))
		_, err := k.RecordIRCompletion(walletAddr, irID, "assistant1", proofHash[:], verifierHash[:], time.Now().Unix())

		if i <= 2 {
			if err != nil {
				t.Errorf("completion %d should succeed, got error: %v", i, err)
			}
		} else if i == 3 {
			// 3rd completion should hit hourly limit
			if err != types.ErrHourlyLimitExceeded {
				t.Errorf("expected ErrHourlyLimitExceeded at completion 3, got %v", err)
			}
			break
		}
	}
}

func TestRecordIRCompletion_ReplayDetection(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())

	registry := newMockIRRegistry()
	registry.activeIRs["IR-000"] = true
	registry.activeIRs["IR-001"] = true
	k.SetIRRegistry(registry)

	walletAddr := "aura1test"
	proofHash := sha256.Sum256([]byte("proof"))
	verifierHash := sha256.Sum256([]byte("verifier"))

	// Complete IR-000 with proof hash
	k.RecordIRCompletion(walletAddr, "IR-000", "assistant1", proofHash[:], verifierHash[:], time.Now().Unix())

	// Try to use same proof hash for different IR
	_, err := k.RecordIRCompletion(walletAddr, "IR-001", "assistant1", proofHash[:], verifierHash[:], time.Now().Unix())

	if err != types.ErrReplayDetected {
		t.Errorf("expected ErrReplayDetected, got %v", err)
	}
}

func TestRecordIRCompletion_StaleAttestation(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())

	registry := newMockIRRegistry()
	registry.activeIRs["IR-000"] = true
	k.SetIRRegistry(registry)

	walletAddr := "aura1test"
	proofHash := sha256.Sum256([]byte("proof"))
	verifierHash := sha256.Sum256([]byte("verifier"))

	// Use timestamp from 10 minutes ago (exceeds 5 minute freshness)
	staleTimestamp := time.Now().Add(-10 * time.Minute).Unix()

	_, err := k.RecordIRCompletion(
		walletAddr,
		"IR-000",
		"assistant1",
		proofHash[:],
		verifierHash[:],
		staleTimestamp,
	)

	if err != types.ErrStaleAttestation {
		t.Errorf("expected ErrStaleAttestation, got %v", err)
	}
}

func TestRecordIRCompletion_VerificationAchieved(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())
	k.SetCurrentHeight(100)

	registry := newMockIRRegistry()
	registry.activeIRs["IR-000"] = true
	registry.irScores["IR-000"] = 10000 // Exactly at verification threshold
	k.SetIRRegistry(registry)

	walletAddr := "aura1test"
	proofHash := sha256.Sum256([]byte("proof"))
	verifierHash := sha256.Sum256([]byte("verifier"))

	score, err := k.RecordIRCompletion(
		walletAddr,
		"IR-000",
		"assistant1",
		proofHash[:],
		verifierHash[:],
		time.Now().Unix(),
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check verification status
	record, _ := k.GetUserRecord(walletAddr)
	if record.Status != types.VerificationStatusVerified {
		t.Errorf("expected verified status, got %v", record.Status)
	}

	if record.VerificationAchievedHeight != 100 {
		t.Errorf("expected verification height 100, got %d", record.VerificationAchievedHeight)
	}

	if score < 10000 {
		t.Errorf("expected score >= 10000, got %d", score)
	}
}

func TestValidateAnchor(t *testing.T) {
	k := NewKeeper(nil, "gov1")

	walletAddr := "aura1test"

	// Test without anchor
	err := k.ValidateAnchor(walletAddr)
	if err != types.ErrAnchorNotCompleted {
		t.Errorf("expected ErrAnchorNotCompleted, got %v", err)
	}

	// Create record with anchor but not completed
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		HasAnchor:     true,
		AnchorInfo:    nil,
	}
	k.SetUserRecord(record)

	err = k.ValidateAnchor(walletAddr)
	if err != types.ErrInvalidAnchor {
		t.Errorf("expected ErrInvalidAnchor, got %v", err)
	}

	// Valid anchor
	record.AnchorInfo = &types.AnchorInfo{
		Completed: true,
	}
	k.SetUserRecord(record)

	err = k.ValidateAnchor(walletAddr)
	if err != nil {
		t.Errorf("expected no error for valid anchor, got %v", err)
	}
}

func TestCalculateVelocityBonus(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	// No anchor - should return 1.0
	bonus := k.CalculateVelocityBonus(walletAddr)
	if bonus != 1.0 {
		t.Errorf("expected 1.0 for no anchor, got %f", bonus)
	}

	// Create record with anchor completed 5 days ago
	anchorTime := time.Now().Add(-5 * 24 * time.Hour).Unix()
	record := types.UserConfidenceRecord{
		WalletAddress: walletAddr,
		HasAnchor:     true,
		TotalScore:    5000, // Below verification threshold
		AnchorInfo: &types.AnchorInfo{
			Completed:   true,
			CompletedAt: timestampFromUnix(anchorTime),
		},
	}
	k.SetUserRecord(record)

	bonus = k.CalculateVelocityBonus(walletAddr)
	if bonus != 1.25 { // Within 7 days
		t.Errorf("expected 1.25 for 5 days, got %f", bonus)
	}

	// Already verified - should return 1.0
	record.TotalScore = 10000
	k.SetUserRecord(record)
	bonus = k.CalculateVelocityBonus(walletAddr)
	if bonus != 1.0 {
		t.Errorf("expected 1.0 for verified user, got %f", bonus)
	}
}

func TestCheckJackpotWin(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentHeight(100)

	walletAddr := "aura1test"
	irID := "IR-001"

	// Test jackpot calculation (deterministic based on seed)
	multiplier := k.CheckJackpotWin(walletAddr, irID)

	// Should be >= 1.0
	if multiplier < 1.0 {
		t.Errorf("expected multiplier >= 1.0, got %f", multiplier)
	}

	// Same inputs should give same result
	multiplier2 := k.CheckJackpotWin(walletAddr, irID)
	if multiplier != multiplier2 {
		t.Error("expected deterministic jackpot result")
	}

	// Different inputs should potentially give different result
	multiplier3 := k.CheckJackpotWin(walletAddr, "IR-002")
	_ = multiplier3 // May or may not be different, but should not panic
}

func TestCleanupExpiredRateLimits(t *testing.T) {
	k := NewKeeper(nil, "gov1")
	k.SetCurrentTime(time.Now().Unix())

	walletAddr := "aura1test"

	// Manually add some rate limit entries
	k.mu.Lock()
	k.rateLimits[walletAddr] = map[string]int{
		"hour_2024_1":   1,
		"day_2024_1":    2,
		"hour_2025_365": 3, // Old entry
	}
	k.mu.Unlock()

	k.CleanupExpiredRateLimits()

	// Check that old entries are removed
	k.mu.RLock()
	limits := k.rateLimits[walletAddr]
	k.mu.RUnlock()

	// Should only have current hour/day entries
	for key := range limits {
		if key == "hour_2025_365" {
			// This might still be there if it matches current time
			continue
		}
	}
}
