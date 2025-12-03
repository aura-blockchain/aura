package keeper

import (
	"fmt"
	"testing"
	"time"

	"github.com/aequitas/aura/chain/x/confidencescore/types"
)

// TestScoreHistoryPerAddress verifies that score history is tracked independently per address
func TestScoreHistoryPerAddress(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr1 := "aura1address1"
	walletAddr2 := "aura1address2"
	walletAddr3 := "aura1address3"

	// Test 1: Empty history for new addresses
	history1 := k.GetScoreHistory(ctx, walletAddr1, 0, 0, 0)
	if len(history1) != 0 {
		t.Errorf("expected 0 history entries for new wallet 1, got %d", len(history1))
	}

	history2 := k.GetScoreHistory(ctx, walletAddr2, 0, 0, 0)
	if len(history2) != 0 {
		t.Errorf("expected 0 history entries for new wallet 2, got %d", len(history2))
	}

	// Test 2: Add history entries for different addresses
	// Add 5 entries for wallet 1
	for i := uint64(1); i <= 5; i++ {
		change := types.ScoreChange{
			ScoreDelta:    int64(i * 100),
			NewTotal:      i * 500,
			Reason:        types.ChangeReasonIRCompletion,
			RelatedIrId:   fmt.Sprintf("IR-%03d", i),
			TxHash:        fmt.Sprintf("tx-wallet1-%d", i),
			PreviousScore: (i - 1) * 500,
		}
		if err := k.AddScoreChange(ctx, walletAddr1, change); err != nil {
			t.Fatalf("failed to add score change for wallet 1: %v", err)
		}
	}

	// Add 3 entries for wallet 2
	for i := uint64(1); i <= 3; i++ {
		change := types.ScoreChange{
			ScoreDelta:    int64(i * 200),
			NewTotal:      i * 1000,
			Reason:        types.ChangeReasonFraudSlash,
			RelatedIrId:   fmt.Sprintf("IR-%03d", i+10),
			TxHash:        fmt.Sprintf("tx-wallet2-%d", i),
			PreviousScore: (i - 1) * 1000,
		}
		if err := k.AddScoreChange(ctx, walletAddr2, change); err != nil {
			t.Fatalf("failed to add score change for wallet 2: %v", err)
		}
	}

	// Add 7 entries for wallet 3
	for i := uint64(1); i <= 7; i++ {
		change := types.ScoreChange{
			ScoreDelta:    int64(i * 50),
			NewTotal:      i * 250,
			Reason:        types.ChangeReasonGovernanceAdjustment,
			RelatedIrId:   fmt.Sprintf("IR-%03d", i+20),
			TxHash:        fmt.Sprintf("tx-wallet3-%d", i),
			PreviousScore: (i - 1) * 250,
		}
		if err := k.AddScoreChange(ctx, walletAddr3, change); err != nil {
			t.Fatalf("failed to add score change for wallet 3: %v", err)
		}
	}

	// Test 3: Verify correct count per address
	history1 = k.GetScoreHistory(ctx, walletAddr1, 0, 0, 0)
	if len(history1) != 5 {
		t.Errorf("expected 5 history entries for wallet 1, got %d", len(history1))
	}

	history2 = k.GetScoreHistory(ctx, walletAddr2, 0, 0, 0)
	if len(history2) != 3 {
		t.Errorf("expected 3 history entries for wallet 2, got %d", len(history2))
	}

	history3 := k.GetScoreHistory(ctx, walletAddr3, 0, 0, 0)
	if len(history3) != 7 {
		t.Errorf("expected 7 history entries for wallet 3, got %d", len(history3))
	}

	// Test 4: Verify wallet address is correctly set in each entry
	for i, h := range history1 {
		if h.WalletAddress != walletAddr1 {
			t.Errorf("history1[%d]: expected wallet address %s, got %s", i, walletAddr1, h.WalletAddress)
		}
		if h.TxHash != fmt.Sprintf("tx-wallet1-%d", i+1) {
			t.Errorf("history1[%d]: expected tx hash tx-wallet1-%d, got %s", i, i+1, h.TxHash)
		}
	}

	for i, h := range history2 {
		if h.WalletAddress != walletAddr2 {
			t.Errorf("history2[%d]: expected wallet address %s, got %s", i, walletAddr2, h.WalletAddress)
		}
		if h.TxHash != fmt.Sprintf("tx-wallet2-%d", i+1) {
			t.Errorf("history2[%d]: expected tx hash tx-wallet2-%d, got %s", i, i+1, h.TxHash)
		}
	}

	for i, h := range history3 {
		if h.WalletAddress != walletAddr3 {
			t.Errorf("history3[%d]: expected wallet address %s, got %s", i, walletAddr3, h.WalletAddress)
		}
		if h.TxHash != fmt.Sprintf("tx-wallet3-%d", i+1) {
			t.Errorf("history3[%d]: expected tx hash tx-wallet3-%d, got %s", i, i+1, h.TxHash)
		}
	}

	// Test 5: Verify histories are completely isolated
	for _, h := range history1 {
		if h.WalletAddress == walletAddr2 || h.WalletAddress == walletAddr3 {
			t.Errorf("wallet 1 history contains entries from other wallets!")
		}
	}

	for _, h := range history2 {
		if h.WalletAddress == walletAddr1 || h.WalletAddress == walletAddr3 {
			t.Errorf("wallet 2 history contains entries from other wallets!")
		}
	}

	for _, h := range history3 {
		if h.WalletAddress == walletAddr1 || h.WalletAddress == walletAddr2 {
			t.Errorf("wallet 3 history contains entries from other wallets!")
		}
	}
}

// TestScoreHistoryQueryByAddress verifies querying history by specific address returns correct results
func TestScoreHistoryQueryByAddress(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr1 := "aura1querytest1"
	walletAddr2 := "aura1querytest2"

	// Add entries with different patterns
	for i := uint64(1); i <= 10; i++ {
		// Advance block height for each entry to ensure proper ordering
		ctx = ctx.WithBlockHeight(int64(100 + i))

		change1 := types.ScoreChange{
			ScoreDelta:    int64(i * 10),
			NewTotal:      i * 100,
			Reason:        types.ChangeReasonIRCompletion,
			RelatedIrId:   fmt.Sprintf("IR-%03d", i),
			TxHash:        fmt.Sprintf("tx1-%d", i),
			PreviousScore: (i - 1) * 100,
		}
		if err := k.AddScoreChange(ctx, walletAddr1, change1); err != nil {
			t.Fatalf("failed to add score change for wallet 1: %v", err)
		}

		change2 := types.ScoreChange{
			ScoreDelta:    int64(i * 20),
			NewTotal:      i * 200,
			Reason:        types.ChangeReasonAppealReversal,
			RelatedIrId:   fmt.Sprintf("IR-%03d", i+100),
			TxHash:        fmt.Sprintf("tx2-%d", i),
			PreviousScore: (i - 1) * 200,
		}
		if err := k.AddScoreChange(ctx, walletAddr2, change2); err != nil {
			t.Fatalf("failed to add score change for wallet 2: %v", err)
		}
	}

	// Query wallet 1
	history1 := k.GetScoreHistory(ctx, walletAddr1, 0, 0, 0)
	if len(history1) != 10 {
		t.Errorf("expected 10 entries for wallet 1, got %d", len(history1))
	}

	// Verify all entries belong to wallet 1
	for i, h := range history1 {
		if h.WalletAddress != walletAddr1 {
			t.Errorf("entry %d: expected wallet %s, got %s", i, walletAddr1, h.WalletAddress)
		}
		if h.Reason != types.ChangeReasonIRCompletion {
			t.Errorf("entry %d: expected reason IRCompletion, got %s", i, h.Reason)
		}
		// History entries are indexed from the loop (1 to 10), so entry i corresponds to loop iteration i+1
		expectedTotal := uint64((i + 1) * 100)
		if h.NewTotal != expectedTotal {
			t.Errorf("entry %d: expected new total %d, got %d", i, expectedTotal, h.NewTotal)
		}
	}

	// Query wallet 2
	history2 := k.GetScoreHistory(ctx, walletAddr2, 0, 0, 0)
	if len(history2) != 10 {
		t.Errorf("expected 10 entries for wallet 2, got %d", len(history2))
	}

	// Verify all entries belong to wallet 2
	for i, h := range history2 {
		if h.WalletAddress != walletAddr2 {
			t.Errorf("entry %d: expected wallet %s, got %s", i, walletAddr2, h.WalletAddress)
		}
		if h.Reason != types.ChangeReasonAppealReversal {
			t.Errorf("entry %d: expected reason AppealReversal, got %s", i, h.Reason)
		}
		// History entries are indexed from the loop (1 to 10), so entry i corresponds to loop iteration i+1
		expectedTotal := uint64((i + 1) * 200)
		if h.NewTotal != expectedTotal {
			t.Errorf("entry %d: expected new total %d, got %d", i, expectedTotal, h.NewTotal)
		}
	}

	// Query non-existent wallet
	historyNone := k.GetScoreHistory(ctx, "aura1nonexistent", 0, 0, 0)
	if len(historyNone) != 0 {
		t.Errorf("expected 0 entries for non-existent wallet, got %d", len(historyNone))
	}
}

// TestScoreHistoryMultipleChangesPerAddress verifies multiple changes for same address are stored correctly
func TestScoreHistoryMultipleChangesPerAddress(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	baseTime := time.Now()
	ctx = ctx.WithBlockHeight(100).WithBlockTime(baseTime)

	walletAddr := "aura1multichangetest"

	// Add 20 changes with different types
	changeTypes := []types.ChangeReason{
		types.ChangeReasonIRCompletion,
		types.ChangeReasonFraudSlash,
		types.ChangeReasonGovernanceAdjustment,
		types.ChangeReasonAppealReversal,
	}

	expectedChanges := []types.ScoreChange{}

	for i := uint64(1); i <= 20; i++ {
		// Advance block height and time for each change
		ctx = ctx.WithBlockHeight(int64(100 + i)).WithBlockTime(baseTime.Add(time.Duration(i) * time.Second))

		change := types.ScoreChange{
			ScoreDelta:    int64((i % 5) * 100),
			NewTotal:      i * 300,
			Reason:        changeTypes[i%4],
			RelatedIrId:   fmt.Sprintf("IR-%03d", i),
			TxHash:        fmt.Sprintf("tx-multi-%d", i),
			PreviousScore: (i - 1) * 300,
		}

		expectedChanges = append(expectedChanges, change)

		if err := k.AddScoreChange(ctx, walletAddr, change); err != nil {
			t.Fatalf("failed to add score change %d: %v", i, err)
		}
	}

	// Retrieve all changes
	history := k.GetScoreHistory(ctx, walletAddr, 0, 0, 0)
	if len(history) != 20 {
		t.Fatalf("expected 20 history entries, got %d", len(history))
	}

	// Verify all changes are present and correct
	for i, h := range history {
		if h.WalletAddress != walletAddr {
			t.Errorf("entry %d: expected wallet %s, got %s", i, walletAddr, h.WalletAddress)
		}
		if h.TxHash != fmt.Sprintf("tx-multi-%d", i+1) {
			t.Errorf("entry %d: expected tx hash tx-multi-%d, got %s", i, i+1, h.TxHash)
		}
		if h.Reason != changeTypes[(i+1)%4] {
			t.Errorf("entry %d: expected reason %s, got %s", i, changeTypes[(i+1)%4], h.Reason)
		}
	}

	// Test with limit
	limitedHistory := k.GetScoreHistory(ctx, walletAddr, 0, 0, 5)
	if len(limitedHistory) != 5 {
		t.Errorf("expected 5 entries with limit, got %d", len(limitedHistory))
	}

	// Verify limited history contains first 5 entries
	for i := 0; i < 5; i++ {
		if limitedHistory[i].TxHash != fmt.Sprintf("tx-multi-%d", i+1) {
			t.Errorf("limited entry %d: expected tx hash tx-multi-%d, got %s", i, i+1, limitedHistory[i].TxHash)
		}
	}
}

// TestScoreHistoryOrdering verifies history is ordered by block height
func TestScoreHistoryOrdering(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	baseTime := time.Now()
	ctx = ctx.WithBlockHeight(100).WithBlockTime(baseTime)

	walletAddr := "aura1ordertest"

	// Add changes at different block heights (not in order)
	heights := []int64{105, 102, 110, 101, 108, 103, 107, 104, 106, 109}

	for i, height := range heights {
		ctx = ctx.WithBlockHeight(height).WithBlockTime(baseTime.Add(time.Duration(height) * time.Second))

		change := types.ScoreChange{
			ScoreDelta:    int64(i * 100),
			NewTotal:      uint64((i + 1) * 500),
			Reason:        types.ChangeReasonIRCompletion,
			RelatedIrId:   fmt.Sprintf("IR-%03d", i),
			TxHash:        fmt.Sprintf("tx-order-%d", height),
			PreviousScore: uint64(i * 500),
		}

		if err := k.AddScoreChange(ctx, walletAddr, change); err != nil {
			t.Fatalf("failed to add score change at height %d: %v", height, err)
		}
	}

	// Retrieve history - should be ordered by block height
	history := k.GetScoreHistory(ctx, walletAddr, 0, 0, 0)
	if len(history) != 10 {
		t.Fatalf("expected 10 history entries, got %d", len(history))
	}

	// Verify ordering by block height (ascending)
	for i := 0; i < len(history)-1; i++ {
		if history[i].BlockHeight > history[i+1].BlockHeight {
			t.Errorf("history not ordered: entry %d (height %d) > entry %d (height %d)",
				i, history[i].BlockHeight, i+1, history[i+1].BlockHeight)
		}
	}

	// Verify entries are at expected heights
	expectedHeights := []int64{101, 102, 103, 104, 105, 106, 107, 108, 109, 110}
	for i, h := range history {
		if int64(h.BlockHeight) != expectedHeights[i] {
			t.Errorf("entry %d: expected height %d, got %d", i, expectedHeights[i], h.BlockHeight)
		}
		if h.TxHash != fmt.Sprintf("tx-order-%d", expectedHeights[i]) {
			t.Errorf("entry %d: expected tx hash tx-order-%d, got %s", i, expectedHeights[i], h.TxHash)
		}
	}
}

// TestScoreHistoryHeightRangeFiltering verifies filtering by height range works correctly
func TestScoreHistoryHeightRangeFiltering(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	baseTime := time.Now()
	ctx = ctx.WithBlockHeight(100).WithBlockTime(baseTime)

	walletAddr := "aura1rangetest"

	// Add changes at heights 100-119
	for i := uint64(0); i < 20; i++ {
		height := int64(100 + i)
		ctx = ctx.WithBlockHeight(height).WithBlockTime(baseTime.Add(time.Duration(i) * time.Second))

		change := types.ScoreChange{
			ScoreDelta:    int64(i * 10),
			NewTotal:      (i + 1) * 100,
			Reason:        types.ChangeReasonIRCompletion,
			RelatedIrId:   fmt.Sprintf("IR-%03d", i),
			TxHash:        fmt.Sprintf("tx-range-%d", height),
			PreviousScore: i * 100,
		}

		if err := k.AddScoreChange(ctx, walletAddr, change); err != nil {
			t.Fatalf("failed to add score change at height %d: %v", height, err)
		}
	}

	// Test 1: Get all history (no filter)
	allHistory := k.GetScoreHistory(ctx, walletAddr, 0, 0, 0)
	if len(allHistory) != 20 {
		t.Errorf("expected 20 total entries, got %d", len(allHistory))
	}

	// Test 2: Filter by fromHeight only (>= 110)
	fromHistory := k.GetScoreHistory(ctx, walletAddr, 110, 0, 0)
	if len(fromHistory) != 10 {
		t.Errorf("expected 10 entries from height 110, got %d", len(fromHistory))
	}
	for _, h := range fromHistory {
		if h.BlockHeight < 110 {
			t.Errorf("found entry with height %d, expected >= 110", h.BlockHeight)
		}
	}

	// Test 3: Filter by toHeight only (<= 105)
	toHistory := k.GetScoreHistory(ctx, walletAddr, 0, 105, 0)
	if len(toHistory) != 6 {
		t.Errorf("expected 6 entries to height 105, got %d", len(toHistory))
	}
	for _, h := range toHistory {
		if h.BlockHeight > 105 {
			t.Errorf("found entry with height %d, expected <= 105", h.BlockHeight)
		}
	}

	// Test 4: Filter by range (105-115)
	rangeHistory := k.GetScoreHistory(ctx, walletAddr, 105, 115, 0)
	if len(rangeHistory) != 11 {
		t.Errorf("expected 11 entries in range 105-115, got %d", len(rangeHistory))
	}
	for _, h := range rangeHistory {
		if h.BlockHeight < 105 || h.BlockHeight > 115 {
			t.Errorf("found entry with height %d, expected in range 105-115", h.BlockHeight)
		}
	}

	// Test 5: Narrow range (110-112)
	narrowHistory := k.GetScoreHistory(ctx, walletAddr, 110, 112, 0)
	if len(narrowHistory) != 3 {
		t.Errorf("expected 3 entries in range 110-112, got %d", len(narrowHistory))
	}

	// Test 6: Range with limit
	limitedRangeHistory := k.GetScoreHistory(ctx, walletAddr, 105, 115, 5)
	if len(limitedRangeHistory) != 5 {
		t.Errorf("expected 5 entries with range and limit, got %d", len(limitedRangeHistory))
	}
	for _, h := range limitedRangeHistory {
		if h.BlockHeight < 105 || h.BlockHeight > 115 {
			t.Errorf("found entry with height %d, expected in range 105-115", h.BlockHeight)
		}
	}
}

// TestScoreHistoryValidation verifies validation of AddScoreChange
func TestScoreHistoryValidation(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	// Test 1: Empty wallet address should fail
	err := k.AddScoreChange(ctx, "", types.ScoreChange{
		ScoreDelta: 100,
		NewTotal:   100,
		Reason:     types.ChangeReasonIRCompletion,
		TxHash:     "tx-test",
	})
	if err == nil {
		t.Error("expected error when adding score change with empty wallet address")
	}
	if err != types.ErrInvalidWalletAddress {
		t.Errorf("expected ErrInvalidWalletAddress, got %v", err)
	}

	// Test 2: Valid wallet address should succeed
	validAddr := "aura1validtest"
	err = k.AddScoreChange(ctx, validAddr, types.ScoreChange{
		ScoreDelta: 100,
		NewTotal:   100,
		Reason:     types.ChangeReasonIRCompletion,
		TxHash:     "tx-valid",
	})
	if err != nil {
		t.Errorf("expected no error with valid wallet address, got %v", err)
	}

	// Verify it was stored
	history := k.GetScoreHistory(ctx, validAddr, 0, 0, 0)
	if len(history) != 1 {
		t.Errorf("expected 1 history entry after valid add, got %d", len(history))
	}
	if len(history) > 0 && history[0].WalletAddress != validAddr {
		t.Errorf("stored entry has wrong wallet address: expected %s, got %s", validAddr, history[0].WalletAddress)
	}
}

// TestScoreHistoryTimestampAndHeight verifies timestamps and block heights are set correctly
func TestScoreHistoryTimestampAndHeight(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	baseTime := time.Now()
	ctx = ctx.WithBlockHeight(100).WithBlockTime(baseTime)

	walletAddr := "aura1timestamptest"

	// Add changes at different heights and times
	for i := uint64(1); i <= 5; i++ {
		height := int64(100 + i)
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		ctx = ctx.WithBlockHeight(height).WithBlockTime(timestamp)

		change := types.ScoreChange{
			ScoreDelta:    int64(i * 100),
			NewTotal:      i * 500,
			Reason:        types.ChangeReasonIRCompletion,
			RelatedIrId:   fmt.Sprintf("IR-%03d", i),
			TxHash:        fmt.Sprintf("tx-time-%d", i),
			PreviousScore: (i - 1) * 500,
		}

		if err := k.AddScoreChange(ctx, walletAddr, change); err != nil {
			t.Fatalf("failed to add score change %d: %v", i, err)
		}
	}

	// Retrieve and verify timestamps and heights
	history := k.GetScoreHistory(ctx, walletAddr, 0, 0, 0)
	if len(history) != 5 {
		t.Fatalf("expected 5 history entries, got %d", len(history))
	}

	for i, h := range history {
		expectedHeight := uint64(101 + i)
		if h.BlockHeight != expectedHeight {
			t.Errorf("entry %d: expected height %d, got %d", i, expectedHeight, h.BlockHeight)
		}

		// Verify timestamp is set (not nil)
		if h.Timestamp == nil {
			t.Errorf("entry %d: timestamp is nil", i)
		}

		// Verify wallet address is set in the change record
		if h.WalletAddress != walletAddr {
			t.Errorf("entry %d: expected wallet %s, got %s", i, walletAddr, h.WalletAddress)
		}
	}
}

// TestScoreHistoryDifferentReasons verifies different change reasons are tracked correctly
func TestScoreHistoryDifferentReasons(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1reasontest"

	// Add changes with different reasons
	reasons := []types.ChangeReason{
		types.ChangeReasonIRCompletion,
		types.ChangeReasonFraudSlash,
		types.ChangeReasonGovernanceAdjustment,
		types.ChangeReasonAppealReversal,
	}

	for i, reason := range reasons {
		ctx = ctx.WithBlockHeight(int64(100 + i + 1))

		change := types.ScoreChange{
			ScoreDelta:    int64((i + 1) * 100),
			NewTotal:      uint64((i + 1) * 500),
			Reason:        reason,
			RelatedIrId:   fmt.Sprintf("IR-%03d", i+1),
			TxHash:        fmt.Sprintf("tx-reason-%d", i+1),
			PreviousScore: uint64(i * 500),
		}

		if err := k.AddScoreChange(ctx, walletAddr, change); err != nil {
			t.Fatalf("failed to add score change with reason %s: %v", reason, err)
		}
	}

	// Retrieve and verify reasons
	history := k.GetScoreHistory(ctx, walletAddr, 0, 0, 0)
	if len(history) != len(reasons) {
		t.Fatalf("expected %d history entries, got %d", len(reasons), len(history))
	}

	for i, h := range history {
		if h.Reason != reasons[i] {
			t.Errorf("entry %d: expected reason %s, got %s", i, reasons[i], h.Reason)
		}
		if h.RelatedIrId != fmt.Sprintf("IR-%03d", i+1) {
			t.Errorf("entry %d: expected IR ID IR-%03d, got %s", i, i+1, h.RelatedIrId)
		}
	}
}

// TestScoreHistoryNegativeDeltas verifies negative score changes (slashes) are tracked correctly
func TestScoreHistoryNegativeDeltas(t *testing.T) {
	ctx, k := setupConfKeeperWithTime(t)
	ctx = ctx.WithBlockHeight(100).WithBlockTime(time.Now())

	walletAddr := "aura1negativedelta"

	// Start with high score and apply slashes
	currentScore := uint64(10000)

	// Add positive changes
	for i := uint64(1); i <= 5; i++ {
		ctx = ctx.WithBlockHeight(int64(100 + i))

		change := types.ScoreChange{
			ScoreDelta:    int64(i * 100),
			NewTotal:      currentScore + i*100,
			Reason:        types.ChangeReasonIRCompletion,
			RelatedIrId:   fmt.Sprintf("IR-%03d", i),
			TxHash:        fmt.Sprintf("tx-positive-%d", i),
			PreviousScore: currentScore,
		}
		currentScore += i * 100

		if err := k.AddScoreChange(ctx, walletAddr, change); err != nil {
			t.Fatalf("failed to add positive score change %d: %v", i, err)
		}
	}

	// Add negative changes (slashes)
	for i := uint64(1); i <= 5; i++ {
		ctx = ctx.WithBlockHeight(int64(105 + i))

		slashAmount := i * 200
		change := types.ScoreChange{
			ScoreDelta:    -int64(slashAmount),
			NewTotal:      currentScore - slashAmount,
			Reason:        types.ChangeReasonFraudSlash,
			RelatedIrId:   fmt.Sprintf("SLASH-%03d", i),
			TxHash:        fmt.Sprintf("tx-negative-%d", i),
			PreviousScore: currentScore,
		}
		currentScore -= slashAmount

		if err := k.AddScoreChange(ctx, walletAddr, change); err != nil {
			t.Fatalf("failed to add negative score change %d: %v", i, err)
		}
	}

	// Retrieve and verify
	history := k.GetScoreHistory(ctx, walletAddr, 0, 0, 0)
	if len(history) != 10 {
		t.Fatalf("expected 10 history entries, got %d", len(history))
	}

	// Verify first 5 are positive
	for i := 0; i < 5; i++ {
		if history[i].ScoreDelta <= 0 {
			t.Errorf("entry %d: expected positive delta, got %d", i, history[i].ScoreDelta)
		}
		if history[i].Reason != types.ChangeReasonIRCompletion {
			t.Errorf("entry %d: expected IRCompletion reason, got %s", i, history[i].Reason)
		}
	}

	// Verify last 5 are negative
	for i := 5; i < 10; i++ {
		if history[i].ScoreDelta >= 0 {
			t.Errorf("entry %d: expected negative delta, got %d", i, history[i].ScoreDelta)
		}
		if history[i].Reason != types.ChangeReasonFraudSlash {
			t.Errorf("entry %d: expected FraudSlash reason, got %s", i, history[i].Reason)
		}
	}
}
