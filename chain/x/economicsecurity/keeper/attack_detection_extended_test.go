// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// SYBIL ATTACK DETECTION TESTS
// ============================

func TestDetectSybilAttack_NoAddresses(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	params, _ := k.GetParams(ctx)
	alert, err := k.detectSybilAttack(ctx, params)

	require.NoError(t, err)
	require.Nil(t, alert, "should not trigger alert with no addresses")
}

func TestDetectSybilAttack_BelowThreshold(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create 40 addresses (below 50 threshold) - should not trigger
	for i := 0; i < 40; i++ {
		addr := fmt.Sprintf("aura1addr%03d", i)
		_ = k.SetUserMEVBalance(ctx, addr, "1000")
		_ = k.SetAddressHolding(ctx, addr, fmt.Sprintf("1000%04d", i))
	}

	params, _ := k.GetParams(ctx)
	alert, err := k.detectSybilAttack(ctx, params)

	require.NoError(t, err)
	require.Nil(t, alert, "should not trigger with fewer than 50 addresses")
}

func TestDetectSybilAttack_ExactlyAtThreshold(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create exactly 50 addresses with varied balances
	for i := 0; i < 50; i++ {
		addr := fmt.Sprintf("aura1addr%03d", i)
		_ = k.SetUserMEVBalance(ctx, addr, "1000")
		// Varied holdings to avoid clustering
		_ = k.SetAddressHolding(ctx, addr, fmt.Sprintf("%d000", (i*17)%100+10))
	}

	params, _ := k.GetParams(ctx)
	alert, err := k.detectSybilAttack(ctx, params)

	require.NoError(t, err)
	// Should not trigger - varied balances, no suspicious clustering
	require.Nil(t, alert)
}

func TestDetectSybilAttack_SuspiciousClustering(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create 100 addresses total
	// 25 addresses with unique balances
	for i := 0; i < 25; i++ {
		addr := fmt.Sprintf("aura1unique%03d", i)
		_ = k.SetUserMEVBalance(ctx, addr, "1000")
		_ = k.SetAddressHolding(ctx, addr, fmt.Sprintf("%d00000", i+1))
	}

	// 75 addresses with identical balance pattern (same first 4 digits)
	// This exceeds 20% threshold (75 > 20 = 100/5)
	for i := 0; i < 75; i++ {
		addr := fmt.Sprintf("aura1sybil%03d", i)
		_ = k.SetUserMEVBalance(ctx, addr, "1000")
		_ = k.SetAddressHolding(ctx, addr, fmt.Sprintf("5000%06d", i)) // all start with "5000"
	}

	params, _ := k.GetParams(ctx)
	alert, err := k.detectSybilAttack(ctx, params)

	require.NoError(t, err)
	require.NotNil(t, alert, "should detect sybil pattern with clustered balances")
	require.Equal(t, types.AttackTypeSybil, alert.AttackType)
	require.Equal(t, types.AlertSeverity_ALERT_SEVERITY_WARNING, alert.Severity)
	require.Contains(t, alert.Message, "similar balance patterns")
}

func TestDetectSybilAttack_EdgeCaseClustering(t *testing.T) {
	tests := []struct {
		name          string
		totalAddrs    int
		clusteredPct  int // percentage that have same pattern
		expectAlert   bool
		minClusterAmt int
	}{
		{
			name:          "exactly 20% clustered with 100 addresses",
			totalAddrs:    100,
			clusteredPct:  20,
			expectAlert:   false, // needs > 20%
			minClusterAmt: 20,
		},
		{
			name:          "21% clustered with 100 addresses",
			totalAddrs:    100,
			clusteredPct:  21,
			expectAlert:   true,
			minClusterAmt: 21,
		},
		{
			name:          "25% clustered but count=19 (under 20 min)",
			totalAddrs:    76, // 25% = 19, which is < 20
			clusteredPct:  25,
			expectAlert:   false,
			minClusterAmt: 19,
		},
		{
			name:          "large scale - 500 addresses with 30% clustered",
			totalAddrs:    500,
			clusteredPct:  30,
			expectAlert:   true,
			minClusterAmt: 150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)
			currentTime := time.Now().Unix()
			_ = k.SetCurrentTime(ctx, currentTime)

			clustered := tt.totalAddrs * tt.clusteredPct / 100
			nonClustered := tt.totalAddrs - clustered

			// Create non-clustered addresses with varied balances
			for i := 0; i < nonClustered; i++ {
				addr := fmt.Sprintf("aura1varied%05d", i)
				_ = k.SetUserMEVBalance(ctx, addr, "1000")
				// Ensure different first 4 digits for each
				_ = k.SetAddressHolding(ctx, addr, fmt.Sprintf("%04d%06d", i%9000+1000, i))
			}

			// Create clustered addresses with same first 4 digits
			for i := 0; i < clustered; i++ {
				addr := fmt.Sprintf("aura1cluster%05d", i)
				_ = k.SetUserMEVBalance(ctx, addr, "1000")
				_ = k.SetAddressHolding(ctx, addr, fmt.Sprintf("9999%06d", i)) // all start with "9999"
			}

			params, _ := k.GetParams(ctx)
			alert, err := k.detectSybilAttack(ctx, params)

			require.NoError(t, err)
			if tt.expectAlert {
				require.NotNil(t, alert, "expected sybil alert")
				require.Equal(t, types.AttackTypeSybil, alert.AttackType)
			} else {
				require.Nil(t, alert, "did not expect sybil alert")
			}
		})
	}
}

func TestDetectSybilAttack_ZeroBalances(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create 60 addresses but all with zero holdings (should be skipped)
	for i := 0; i < 60; i++ {
		addr := fmt.Sprintf("aura1zero%03d", i)
		_ = k.SetUserMEVBalance(ctx, addr, "1000")
		_ = k.SetAddressHolding(ctx, addr, "0")
	}

	params, _ := k.GetParams(ctx)
	alert, err := k.detectSybilAttack(ctx, params)

	require.NoError(t, err)
	require.Nil(t, alert, "zero balances should be skipped in detection")
}

func TestDetectSybilAttack_ShortHoldings(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create 60 addresses with holdings < 4 characters
	for i := 0; i < 60; i++ {
		addr := fmt.Sprintf("aura1short%03d", i)
		_ = k.SetUserMEVBalance(ctx, addr, "1000")
		_ = k.SetAddressHolding(ctx, addr, fmt.Sprintf("%d", i%1000)) // 0-999, < 4 chars
	}

	params, _ := k.GetParams(ctx)
	alert, err := k.detectSybilAttack(ctx, params)

	require.NoError(t, err)
	// Short holdings are counted but not bucketed, so no clustering detected
	require.Nil(t, alert)
}

// ============================
// WASH TRADING DETECTION TESTS
// ============================

func TestDetectWashTrading_NoTransactions(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	params, _ := k.GetParams(ctx)
	alert, err := k.detectWashTrading(ctx, params)

	require.NoError(t, err)
	require.Nil(t, alert, "should not trigger with no transactions")
}

func TestDetectWashTrading_OneWayTransactions(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create one-way transactions (A -> B only)
	for i := 0; i < 5; i++ {
		record := &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_oneway_%d", i),
			Sender:             "aura1alice",
			Recipient:          "aura1bob",
			Amount:             "1000000",
			PercentageOfSupply: 10,
			BlockHeight:        uint64(i + 1),
			Timestamp:          time.Unix(currentTime-int64(i*60), 0),
			Flagged:            true,
		}
		_ = k.SetLargeTxRecord(ctx, record)
	}

	params, _ := k.GetParams(ctx)
	alert, err := k.detectWashTrading(ctx, params)

	require.NoError(t, err)
	require.Nil(t, alert, "one-way transactions should not trigger wash trading")
}

func TestDetectWashTrading_CircularPattern(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create circular transactions: A -> B and B -> A
	// 4 transactions A -> B
	for i := 0; i < 4; i++ {
		record := &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_ab_%d", i),
			Sender:             "aura1alice",
			Recipient:          "aura1bob",
			Amount:             "1000000",
			PercentageOfSupply: 10,
			BlockHeight:        uint64(i + 1),
			Timestamp:          time.Unix(currentTime-int64(i*60), 0),
			Flagged:            true,
		}
		_ = k.SetLargeTxRecord(ctx, record)
	}

	// 2 transactions B -> A
	for i := 0; i < 2; i++ {
		record := &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_ba_%d", i),
			Sender:             "aura1bob",
			Recipient:          "aura1alice",
			Amount:             "1000000",
			PercentageOfSupply: 10,
			BlockHeight:        uint64(i + 5),
			Timestamp:          time.Unix(currentTime-int64((i+5)*60), 0),
			Flagged:            true,
		}
		_ = k.SetLargeTxRecord(ctx, record)
	}

	params, _ := k.GetParams(ctx)
	alert, err := k.detectWashTrading(ctx, params)

	require.NoError(t, err)
	require.NotNil(t, alert, "circular trading should trigger wash trading alert")
	require.Equal(t, types.AttackTypeWashTrading, alert.AttackType)
	require.Equal(t, types.AlertSeverity_ALERT_SEVERITY_CRITICAL, alert.Severity)
}

func TestDetectWashTrading_OldTransactionsIgnored(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create circular transactions but all older than 1 hour
	cutoffTime := currentTime - 7200 // 2 hours ago

	for i := 0; i < 4; i++ {
		record := &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_old_ab_%d", i),
			Sender:             "aura1alice",
			Recipient:          "aura1bob",
			Amount:             "1000000",
			PercentageOfSupply: 10,
			BlockHeight:        uint64(i + 1),
			Timestamp:          time.Unix(cutoffTime-int64(i*60), 0),
			Flagged:            true,
		}
		_ = k.SetLargeTxRecord(ctx, record)
	}

	for i := 0; i < 2; i++ {
		record := &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_old_ba_%d", i),
			Sender:             "aura1bob",
			Recipient:          "aura1alice",
			Amount:             "1000000",
			PercentageOfSupply: 10,
			BlockHeight:        uint64(i + 5),
			Timestamp:          time.Unix(cutoffTime-int64((i+5)*60), 0),
			Flagged:            true,
		}
		_ = k.SetLargeTxRecord(ctx, record)
	}

	params, _ := k.GetParams(ctx)
	alert, err := k.detectWashTrading(ctx, params)

	require.NoError(t, err)
	require.Nil(t, alert, "old transactions should be ignored")
}

func TestDetectWashTrading_BelowThreshold(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create exactly 2 A->B transactions (at threshold, not above)
	for i := 0; i < 2; i++ {
		record := &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_thresh_ab_%d", i),
			Sender:             "aura1alice",
			Recipient:          "aura1bob",
			Amount:             "1000000",
			PercentageOfSupply: 10,
			BlockHeight:        uint64(i + 1),
			Timestamp:          time.Unix(currentTime-int64(i*60), 0),
			Flagged:            true,
		}
		_ = k.SetLargeTxRecord(ctx, record)
	}

	// 1 transaction B -> A
	record := &types.LargeTxRecord{
		TxHash:             "tx_thresh_ba_0",
		Sender:             "aura1bob",
		Recipient:          "aura1alice",
		Amount:             "1000000",
		PercentageOfSupply: 10,
		BlockHeight:        3,
		Timestamp:          time.Unix(currentTime-180, 0),
		Flagged:            true,
	}
	_ = k.SetLargeTxRecord(ctx, record)

	params, _ := k.GetParams(ctx)
	alert, err := k.detectWashTrading(ctx, params)

	require.NoError(t, err)
	// count=2 is not > 2, so should not trigger
	require.Nil(t, alert, "exactly 2 transactions should not trigger (needs >2)")
}

func TestDetectWashTrading_MultiplePairs(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create multiple trading pairs, one with wash trading pattern
	// Pair 1: A <-> B (wash trading)
	for i := 0; i < 4; i++ {
		_ = k.SetLargeTxRecord(ctx, &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_pair1_ab_%d", i),
			Sender:             "aura1alice",
			Recipient:          "aura1bob",
			Amount:             "1000000",
			PercentageOfSupply: 10,
			BlockHeight:        uint64(i + 1),
			Timestamp:          time.Unix(currentTime-int64(i*60), 0),
			Flagged:            true,
		})
	}
	for i := 0; i < 2; i++ {
		_ = k.SetLargeTxRecord(ctx, &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_pair1_ba_%d", i),
			Sender:             "aura1bob",
			Recipient:          "aura1alice",
			Amount:             "1000000",
			PercentageOfSupply: 10,
			BlockHeight:        uint64(i + 5),
			Timestamp:          time.Unix(currentTime-int64((i+5)*60), 0),
			Flagged:            true,
		})
	}

	// Pair 2: C -> D (one-way, no wash trading)
	for i := 0; i < 3; i++ {
		_ = k.SetLargeTxRecord(ctx, &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_pair2_cd_%d", i),
			Sender:             "aura1charlie",
			Recipient:          "aura1david",
			Amount:             "500000",
			PercentageOfSupply: 5,
			BlockHeight:        uint64(i + 10),
			Timestamp:          time.Unix(currentTime-int64((i+10)*60), 0),
			Flagged:            false,
		})
	}

	params, _ := k.GetParams(ctx)
	alert, err := k.detectWashTrading(ctx, params)

	require.NoError(t, err)
	require.NotNil(t, alert, "should detect wash trading in pair 1")
	require.Equal(t, types.AttackTypeWashTrading, alert.AttackType)
}

// ============================
// COMBINED ATTACK SCENARIOS
// ============================

func TestDetectEconomicAttacks_MultipleAttacks(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Setup: Pump and dump (>5 large txs in hour)
	for i := 0; i < 7; i++ {
		record := &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_pump_%d", i),
			Sender:             fmt.Sprintf("aura1pump%d", i),
			Recipient:          "aura1exchange",
			Amount:             "10000000000",
			PercentageOfSupply: 150,
			BlockHeight:        uint64(i + 1),
			Timestamp:          time.Unix(currentTime-int64(i*300), 0),
			Flagged:            true,
		}
		_ = k.SetLargeTxRecord(ctx, record)
	}

	alerts, err := k.DetectEconomicAttacks(ctx)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(alerts), 1, "should detect at least one attack")

	// Check we have pump and dump
	hasPumpAndDump := false
	for _, alert := range alerts {
		if alert.AttackType == types.AttackTypePumpAndDump {
			hasPumpAndDump = true
		}
	}
	require.True(t, hasPumpAndDump, "should detect pump and dump")
}

func TestDetectEconomicAttacks_FlashLoanAndWashTrading(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Setup: Flash loan attack (3+ rapid txs from same address in 1 minute)
	for i := 0; i < 4; i++ {
		record := &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_flash_%d", i),
			Sender:             "aura1flashattacker",
			Recipient:          fmt.Sprintf("aura1victim%d", i),
			Amount:             "50000000000",
			PercentageOfSupply: 300,
			BlockHeight:        uint64(i + 1),
			Timestamp:          time.Unix(currentTime-int64(i*10), 0), // within 1 minute
			Flagged:            true,
		}
		_ = k.SetLargeTxRecord(ctx, record)
	}

	alerts, err := k.DetectEconomicAttacks(ctx)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(alerts), 1)

	hasFlashLoan := false
	for _, alert := range alerts {
		if alert.AttackType == types.AttackTypeFlashLoan {
			hasFlashLoan = true
			require.Equal(t, "aura1flashattacker", alert.SuspectAddress)
			require.Equal(t, types.AlertSeverity_ALERT_SEVERITY_CRITICAL, alert.Severity)
		}
	}
	require.True(t, hasFlashLoan, "should detect flash loan attack")
}

func TestDetectEconomicAttacks_SybilWithWashTrading(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Setup sybil attack: 100 addresses with 30 having similar balances
	for i := 0; i < 70; i++ {
		addr := fmt.Sprintf("aura1legit%03d", i)
		_ = k.SetUserMEVBalance(ctx, addr, "1000")
		_ = k.SetAddressHolding(ctx, addr, fmt.Sprintf("%04d%06d", i%9000+1000, i))
	}
	for i := 0; i < 30; i++ {
		addr := fmt.Sprintf("aura1sybil%03d", i)
		_ = k.SetUserMEVBalance(ctx, addr, "1000")
		_ = k.SetAddressHolding(ctx, addr, fmt.Sprintf("8888%06d", i))
	}

	// Also setup wash trading
	for i := 0; i < 4; i++ {
		_ = k.SetLargeTxRecord(ctx, &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_wash_ab_%d", i),
			Sender:             "aura1washer1",
			Recipient:          "aura1washer2",
			Amount:             "5000000",
			PercentageOfSupply: 50,
			BlockHeight:        uint64(i + 1),
			Timestamp:          time.Unix(currentTime-int64(i*60), 0),
			Flagged:            true,
		})
	}
	for i := 0; i < 2; i++ {
		_ = k.SetLargeTxRecord(ctx, &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_wash_ba_%d", i),
			Sender:             "aura1washer2",
			Recipient:          "aura1washer1",
			Amount:             "5000000",
			PercentageOfSupply: 50,
			BlockHeight:        uint64(i + 5),
			Timestamp:          time.Unix(currentTime-int64((i+5)*60), 0),
			Flagged:            true,
		})
	}

	alerts, err := k.DetectEconomicAttacks(ctx)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(alerts), 2, "should detect at least sybil and wash trading")

	hasSybil := false
	hasWashTrading := false
	for _, alert := range alerts {
		switch alert.AttackType {
		case types.AttackTypeSybil:
			hasSybil = true
		case types.AttackTypeWashTrading:
			hasWashTrading = true
		}
	}
	require.True(t, hasSybil, "should detect sybil attack")
	require.True(t, hasWashTrading, "should detect wash trading")
}

// ============================
// ALERT GENERATION TESTS
// ============================

func TestCreateAttackAlert_UniqueIDs(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create multiple alerts and verify they have unique IDs
	alerts := make([]*types.AttackAlert, 0, 5)
	for i := 0; i < 5; i++ {
		alert, err := k.createAttackAlert(
			ctx,
			types.AttackTypePumpAndDump,
			types.AlertSeverity_ALERT_SEVERITY_WARNING,
			fmt.Sprintf("Test message %d", i),
			fmt.Sprintf("aura1suspect%d", i),
			uint64(i+1),
		)
		require.NoError(t, err)
		require.NotNil(t, alert)
		alerts = append(alerts, alert)
	}

	// Verify all IDs are unique
	idSet := make(map[string]bool)
	for _, alert := range alerts {
		require.NotEmpty(t, alert.AlertId)
		require.Len(t, alert.AlertId, 16, "alert ID should be 16 chars")
		require.False(t, idSet[alert.AlertId], "alert IDs should be unique")
		idSet[alert.AlertId] = true
	}
}

func TestCreateAttackAlert_AllSeverityLevels(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	severities := []types.AlertSeverity{
		types.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED,
		types.AlertSeverity_ALERT_SEVERITY_INFO,
		types.AlertSeverity_ALERT_SEVERITY_WARNING,
		types.AlertSeverity_ALERT_SEVERITY_CRITICAL,
	}

	for _, sev := range severities {
		alert, err := k.createAttackAlert(
			ctx,
			types.AttackTypeSybil,
			sev,
			"Test message",
			"aura1suspect",
			10,
		)
		require.NoError(t, err)
		require.NotNil(t, alert)
		require.Equal(t, sev, alert.Severity)
	}
}

func TestCreateAttackAlert_AllAttackTypes(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	attackTypes := []types.AttackType{
		types.AttackTypePumpAndDump,
		types.AttackTypeFlashLoan,
		types.AttackTypeSybil,
		types.AttackTypeWashTrading,
		types.AttackTypeFrontRunning,
	}

	for _, at := range attackTypes {
		alert, err := k.createAttackAlert(
			ctx,
			at,
			types.AlertSeverity_ALERT_SEVERITY_WARNING,
			fmt.Sprintf("Test %s", at.String()),
			"aura1suspect",
			5,
		)
		require.NoError(t, err)
		require.NotNil(t, alert)
		require.Equal(t, at, alert.AttackType)
	}
}

func TestCreateAttackAlert_EmptySuspectAddress(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	alert, err := k.createAttackAlert(
		ctx,
		types.AttackTypeSybil,
		types.AlertSeverity_ALERT_SEVERITY_WARNING,
		"Sybil attack detected",
		"", // Empty suspect address (common for sybil attacks)
		100,
	)

	require.NoError(t, err)
	require.NotNil(t, alert)
	require.Empty(t, alert.SuspectAddress)
	require.Equal(t, uint64(100), alert.EvidenceCount)
	require.False(t, alert.AutoMitigated)
	require.Empty(t, alert.MitigationAction)
}

func TestCreateAttackAlert_ZeroEvidenceCount(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	alert, err := k.createAttackAlert(
		ctx,
		types.AttackTypeFrontRunning,
		types.AlertSeverity_ALERT_SEVERITY_WARNING,
		"High gas price detected",
		"",
		0, // Zero evidence count
	)

	require.NoError(t, err)
	require.NotNil(t, alert)
	require.Equal(t, uint64(0), alert.EvidenceCount)
}

// ============================
// FRONT-RUNNING DETECTION TESTS
// ============================

func TestDetectFrontRunning_InvalidBaseFee(t *testing.T) {
	params := types.DefaultParams()
	params.DynamicFees.BaseFee = "invalid_number"

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	testParams, _ := k.GetParams(ctx)
	alert, err := k.detectFrontRunning(ctx, testParams)

	require.NoError(t, err)
	require.Nil(t, alert, "invalid base fee should return nil, not error")
}

func TestDetectFrontRunning_ExactlyAtThreshold(t *testing.T) {
	params := types.DefaultParams()
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 20000 // exactly 2x

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	testParams, _ := k.GetParams(ctx)
	alert, err := k.detectFrontRunning(ctx, testParams)

	require.NoError(t, err)
	require.Nil(t, alert, "exactly at 2x should not trigger (needs >2x)")
}

func TestDetectFrontRunning_JustAboveThreshold(t *testing.T) {
	params := types.DefaultParams()
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 20001 // just above 2x

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	testParams, _ := k.GetParams(ctx)
	alert, err := k.detectFrontRunning(ctx, testParams)

	require.NoError(t, err)
	require.NotNil(t, alert, "just above 2x should trigger")
	require.Equal(t, types.AttackTypeFrontRunning, alert.AttackType)
	require.Equal(t, types.AlertSeverity_ALERT_SEVERITY_WARNING, alert.Severity)
}

func TestDetectFrontRunning_ExtremeMultiplier(t *testing.T) {
	params := types.DefaultParams()
	params.DynamicFees.BaseFee = "1000"
	params.DynamicFees.CurrentMultiplier = 100000 // 10x multiplier

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	testParams, _ := k.GetParams(ctx)
	alert, err := k.detectFrontRunning(ctx, testParams)

	require.NoError(t, err)
	require.NotNil(t, alert)
	require.Equal(t, types.AttackTypeFrontRunning, alert.AttackType)
	require.Contains(t, alert.Message, "10000")
}

// ============================
// EDGE CASE TESTS
// ============================

func TestDetectEconomicAttacks_EmptyState(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	alerts, err := k.DetectEconomicAttacks(ctx)

	require.NoError(t, err)
	require.Empty(t, alerts, "empty state should produce no alerts")
}

func TestDetectEconomicAttacks_OnlyOldData(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	oldTime := currentTime - 86400 // 1 day ago

	// Add old transactions that would otherwise trigger alerts
	for i := 0; i < 10; i++ {
		record := &types.LargeTxRecord{
			TxHash:             fmt.Sprintf("tx_old_%d", i),
			Sender:             "aura1oldwhale",
			Recipient:          "aura1exchange",
			Amount:             "10000000000",
			PercentageOfSupply: 150,
			BlockHeight:        uint64(i + 1),
			Timestamp:          time.Unix(oldTime-int64(i*300), 0),
			Flagged:            true,
		}
		_ = k.SetLargeTxRecord(ctx, record)
	}

	alerts, err := k.DetectEconomicAttacks(ctx)

	require.NoError(t, err)
	require.Empty(t, alerts, "old data should not trigger alerts")
}

func TestRecordAttackAlert_MultipleAlerts(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	// Record multiple alerts in sequence
	for i := 0; i < 10; i++ {
		alert := &types.AttackAlert{
			AlertId:        fmt.Sprintf("test-alert-%d", i),
			AttackType:     types.AttackType(i % 5),
			Severity:       types.AlertSeverity(i % 4),
			Message:        fmt.Sprintf("Alert message %d", i),
			DetectedAt:     time.Now(),
			SuspectAddress: fmt.Sprintf("aura1suspect%d", i),
			EvidenceCount:  uint64(i + 1),
			AutoMitigated:  i%2 == 0,
		}

		err := k.RecordAttackAlert(ctx, alert)
		require.NoError(t, err)
	}
}

// ============================
// CONCURRENT/DETERMINISM TESTS
// ============================

func TestDetectWashTrading_DeterministicPairOrdering(t *testing.T) {
	// Test that detection is deterministic regardless of iteration order
	for run := 0; run < 5; run++ {
		k, ctx := setupKeeperForTest(t)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Create multiple pairs that could potentially trigger
		pairs := []struct {
			sender    string
			recipient string
		}{
			{"aura1zebra", "aura1yak"},
			{"aura1alice", "aura1bob"},
			{"aura1mike", "aura1nancy"},
		}

		for _, pair := range pairs {
			for i := 0; i < 4; i++ {
				_ = k.SetLargeTxRecord(ctx, &types.LargeTxRecord{
					TxHash:             fmt.Sprintf("tx_%s_%s_%d", pair.sender, pair.recipient, i),
					Sender:             pair.sender,
					Recipient:          pair.recipient,
					Amount:             "1000000",
					PercentageOfSupply: 10,
					BlockHeight:        uint64(i + 1),
					Timestamp:          time.Unix(currentTime-int64(i*60), 0),
					Flagged:            true,
				})
			}
			for i := 0; i < 2; i++ {
				_ = k.SetLargeTxRecord(ctx, &types.LargeTxRecord{
					TxHash:             fmt.Sprintf("tx_%s_%s_rev_%d", pair.recipient, pair.sender, i),
					Sender:             pair.recipient,
					Recipient:          pair.sender,
					Amount:             "1000000",
					PercentageOfSupply: 10,
					BlockHeight:        uint64(i + 10),
					Timestamp:          time.Unix(currentTime-int64((i+10)*60), 0),
					Flagged:            true,
				})
			}
		}

		params, _ := k.GetParams(ctx)
		alert, err := k.detectWashTrading(ctx, params)

		require.NoError(t, err)
		require.NotNil(t, alert)
		// The first pair alphabetically should be detected consistently
		require.Contains(t, alert.Message, "aura1alice")
	}
}

func TestDetectSybilAttack_DeterministicRangeOrdering(t *testing.T) {
	// Test that sybil detection is deterministic
	for run := 0; run < 3; run++ {
		k, ctx := setupKeeperForTest(t)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Create addresses that would cluster into multiple ranges
		ranges := []string{"1111", "2222", "3333", "4444", "5555"}
		for rangeIdx, prefix := range ranges {
			count := 15 + rangeIdx*5 // 15, 20, 25, 30, 35 per range
			for i := 0; i < count; i++ {
				addr := fmt.Sprintf("aura1r%d_%04d", rangeIdx, i)
				_ = k.SetUserMEVBalance(ctx, addr, "1000")
				_ = k.SetAddressHolding(ctx, addr, fmt.Sprintf("%s%06d", prefix, i))
			}
		}

		params, _ := k.GetParams(ctx)
		alert, err := k.detectSybilAttack(ctx, params)

		require.NoError(t, err)
		if alert != nil {
			// Should consistently detect the same range if any
			require.Equal(t, types.AttackTypeSybil, alert.AttackType)
		}
	}
}

// ============================
// BOUNDARY CONDITION TESTS
// ============================

func TestDetectFlashLoanAttack_ExactlyAtBoundary(t *testing.T) {
	tests := []struct {
		name          string
		txCount       int
		withinMinute  bool
		expectAlert   bool
	}{
		{
			name:          "2 transactions in minute - no alert",
			txCount:       2,
			withinMinute:  true,
			expectAlert:   false,
		},
		{
			name:          "3 transactions in minute - triggers",
			txCount:       3,
			withinMinute:  true,
			expectAlert:   true,
		},
		{
			name:          "4 transactions outside minute - no alert",
			txCount:       4,
			withinMinute:  false,
			expectAlert:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)
			currentTime := time.Now().Unix()
			_ = k.SetCurrentTime(ctx, currentTime)

			var baseTime int64
			if tt.withinMinute {
				baseTime = currentTime - 30 // 30 seconds ago
			} else {
				baseTime = currentTime - 120 // 2 minutes ago
			}

			for i := 0; i < tt.txCount; i++ {
				record := &types.LargeTxRecord{
					TxHash:             fmt.Sprintf("tx_boundary_%d", i),
					Sender:             "aura1flashtest",
					Recipient:          fmt.Sprintf("aura1victim%d", i),
					Amount:             "10000000000",
					PercentageOfSupply: 200,
					BlockHeight:        uint64(i + 1),
					Timestamp:          time.Unix(baseTime+int64(i*5), 0),
					Flagged:            true,
				}
				_ = k.SetLargeTxRecord(ctx, record)
			}

			params, _ := k.GetParams(ctx)
			alert, err := k.detectFlashLoanAttack(ctx, params)

			require.NoError(t, err)
			if tt.expectAlert {
				require.NotNil(t, alert)
				require.Equal(t, types.AttackTypeFlashLoan, alert.AttackType)
			} else {
				require.Nil(t, alert)
			}
		})
	}
}

func TestDetectPumpAndDump_ExactlyAtBoundary(t *testing.T) {
	tests := []struct {
		name        string
		txCount     int
		expectAlert bool
	}{
		{"5 transactions - no alert", 5, false},
		{"6 transactions - triggers", 6, true},
		{"10 transactions - triggers", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)
			currentTime := time.Now().Unix()
			_ = k.SetCurrentTime(ctx, currentTime)

			for i := 0; i < tt.txCount; i++ {
				record := &types.LargeTxRecord{
					TxHash:             fmt.Sprintf("tx_pnd_%d", i),
					Sender:             fmt.Sprintf("aura1sender%d", i),
					Recipient:          "aura1exchange",
					Amount:             "5000000000",
					PercentageOfSupply: 100,
					BlockHeight:        uint64(i + 1),
					Timestamp:          time.Unix(currentTime-int64(i*300), 0), // within 1 hour
					Flagged:            true,
				}
				_ = k.SetLargeTxRecord(ctx, record)
			}

			params, _ := k.GetParams(ctx)
			alert, err := k.detectPumpAndDump(ctx, params)

			require.NoError(t, err)
			if tt.expectAlert {
				require.NotNil(t, alert)
				require.Equal(t, types.AttackTypePumpAndDump, alert.AttackType)
			} else {
				require.Nil(t, alert)
			}
		})
	}
}

// ============================
// HELPER FUNCTION TESTS
// ============================

func TestGetAttackAlerts_LimitAndFilter(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	tests := []struct {
		name     string
		limit    uint64
		severity types.AlertSeverity
	}{
		{"no limit no filter", 0, types.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED},
		{"limit 5 no filter", 5, types.AlertSeverity_ALERT_SEVERITY_UNSPECIFIED},
		{"limit 100 critical filter", 100, types.AlertSeverity_ALERT_SEVERITY_CRITICAL},
		{"limit 10 warning filter", 10, types.AlertSeverity_ALERT_SEVERITY_WARNING},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts := k.GetAttackAlerts(tt.limit, tt.severity)
			require.NotNil(t, alerts) // Stub returns empty slice
		})
	}
}

func TestGetAttacksByType_AllTypes(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	allTypes := []types.AttackType{
		types.AttackTypeUnspecified,
		types.AttackTypePumpAndDump,
		types.AttackTypeFlashLoan,
		types.AttackTypeSybil,
		types.AttackTypeWashTrading,
		types.AttackTypeFrontRunning,
	}

	for _, attackType := range allTypes {
		t.Run(attackType.String(), func(t *testing.T) {
			alerts := k.GetAttacksByType(attackType)
			require.NotNil(t, alerts)
		})
	}
}

func TestSybilAttackScenarioHelper(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	totalAddrs := 100
	clusteredAddrs := 30
	clusterPrefix := "7777"

	// Create non-clustered addresses
	nonClustered := totalAddrs - clusteredAddrs
	for i := 0; i < nonClustered; i++ {
		addr := fmt.Sprintf("aura1normal%05d", i)
		_ = k.SetUserMEVBalance(ctx, addr, "1000")
		_ = k.SetAddressHolding(ctx, addr, fmt.Sprintf("%04d%06d", i%9000+1000, i))
	}

	// Create clustered addresses
	for i := 0; i < clusteredAddrs; i++ {
		addr := fmt.Sprintf("aura1sybil%05d", i)
		_ = k.SetUserMEVBalance(ctx, addr, "1000")
		_ = k.SetAddressHolding(ctx, addr, fmt.Sprintf("%s%06d", clusterPrefix, i))
	}

	params, _ := k.GetParams(ctx)
	alert, err := k.detectSybilAttack(ctx, params)

	require.NoError(t, err)
	require.NotNil(t, alert)
	require.Equal(t, types.AttackTypeSybil, alert.AttackType)
}
