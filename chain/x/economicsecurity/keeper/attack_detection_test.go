// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

func TestDetectEconomicAttacks(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*Keeper, context.Context)
		expectedAlerts int
		wantErr        bool
	}{
		{
			name:           "no attacks with empty state",
			setup:          func(k *Keeper, ctx context.Context) {},
			expectedAlerts: 0,
			wantErr:        false,
		},
		{
			name: "pump and dump detection with many large txs",
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)

				// Create 6+ large transactions in the last hour
				for i := 0; i < 7; i++ {
					record := &types.LargeTxRecord{
						TxHash:             "tx_pump_" + string(rune('a'+i)),
						Sender:             "aura1sender" + string(rune('a'+i)),
						Recipient:          "aura1recipient",
						Amount:             "1000000000",
						PercentageOfSupply: 100,
						BlockHeight:        uint64(i + 1),
						Timestamp:          time.Unix(currentTime-int64(i*60), 0),
						Flagged:            true,
					}
					_ = k.SetLargeTxRecord(ctx, record)
				}
			},
			expectedAlerts: 1,
			wantErr:        false,
		},
		{
			name: "flash loan detection with rapid txs from same address",
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)

				// Create 3+ rapid transactions from same address within 1 minute
				for i := 0; i < 4; i++ {
					record := &types.LargeTxRecord{
						TxHash:             "tx_flash_" + string(rune('a'+i)),
						Sender:             "aura1flashattacker",
						Recipient:          "aura1victim",
						Amount:             "5000000000",
						PercentageOfSupply: 200,
						BlockHeight:        uint64(i + 1),
						Timestamp:          time.Unix(currentTime-int64(i*10), 0),
						Flagged:            true,
					}
					_ = k.SetLargeTxRecord(ctx, record)
				}
			},
			expectedAlerts: 1, // flash loan detection only (pump&dump requires >5 txs in hour)
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)
			tt.setup(k, ctx)

			alerts, err := k.DetectEconomicAttacks(ctx)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, alerts, tt.expectedAlerts)
			}
		})
	}
}

func TestDetectPumpAndDump(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*Keeper, context.Context)
		expectAlert bool
	}{
		{
			name: "no alert with few transactions",
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)

				// Only 3 transactions - below threshold
				for i := 0; i < 3; i++ {
					record := &types.LargeTxRecord{
						TxHash:             "tx_normal_" + string(rune('a'+i)),
						Sender:             "aura1sender" + string(rune('a'+i)),
						Recipient:          "aura1recipient",
						Amount:             "100000",
						PercentageOfSupply: 10,
						BlockHeight:        uint64(i + 1),
						Timestamp:          time.Unix(currentTime-int64(i*60), 0),
						Flagged:            false,
					}
					_ = k.SetLargeTxRecord(ctx, record)
				}
			},
			expectAlert: false,
		},
		{
			name: "alert with >5 transactions in hour",
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)

				// 6 transactions within the hour
				for i := 0; i < 6; i++ {
					record := &types.LargeTxRecord{
						TxHash:             "tx_pump_" + string(rune('a'+i)),
						Sender:             "aura1whale" + string(rune('a'+i)),
						Recipient:          "aura1exchange",
						Amount:             "10000000000",
						PercentageOfSupply: 150,
						BlockHeight:        uint64(i + 1),
						Timestamp:          time.Unix(currentTime-int64(i*300), 0),
						Flagged:            true,
					}
					_ = k.SetLargeTxRecord(ctx, record)
				}
			},
			expectAlert: true,
		},
		{
			name: "no alert with old transactions",
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)

				// Transactions older than 1 hour
				for i := 0; i < 10; i++ {
					record := &types.LargeTxRecord{
						TxHash:             "tx_old_" + string(rune('a'+i)),
						Sender:             "aura1oldsender" + string(rune('a'+i)),
						Recipient:          "aura1recipient",
						Amount:             "1000000000",
						PercentageOfSupply: 100,
						BlockHeight:        uint64(i + 1),
						Timestamp:          time.Unix(currentTime-7200-int64(i*60), 0), // 2+ hours ago
						Flagged:            true,
					}
					_ = k.SetLargeTxRecord(ctx, record)
				}
			},
			expectAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)
			tt.setup(k, ctx)

			params, _ := k.GetParams(ctx)
			alert, err := k.detectPumpAndDump(ctx, params)

			require.NoError(t, err)
			if tt.expectAlert {
				require.NotNil(t, alert)
				require.Equal(t, types.AttackTypePumpAndDump, alert.AttackType)
				require.Equal(t, types.AlertSeverity_ALERT_SEVERITY_WARNING, alert.Severity)
			} else {
				require.Nil(t, alert)
			}
		})
	}
}

func TestDetectFlashLoanAttack(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*Keeper, context.Context)
		expectAlert bool
	}{
		{
			name: "no alert with transactions from different addresses",
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)

				for i := 0; i < 3; i++ {
					record := &types.LargeTxRecord{
						TxHash:             "tx_different_" + string(rune('a'+i)),
						Sender:             "aura1different" + string(rune('a'+i)),
						Recipient:          "aura1recipient",
						Amount:             "5000000000",
						PercentageOfSupply: 200,
						BlockHeight:        uint64(i + 1),
						Timestamp:          time.Unix(currentTime-int64(i*10), 0),
						Flagged:            true,
					}
					_ = k.SetLargeTxRecord(ctx, record)
				}
			},
			expectAlert: false,
		},
		{
			name: "alert with 3+ transactions from same address in 1 minute",
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)

				for i := 0; i < 4; i++ {
					record := &types.LargeTxRecord{
						TxHash:             "tx_flash_same_" + string(rune('a'+i)),
						Sender:             "aura1flashloanattacker",
						Recipient:          "aura1victim" + string(rune('a'+i)),
						Amount:             "100000000000",
						PercentageOfSupply: 500,
						BlockHeight:        uint64(i + 1),
						Timestamp:          time.Unix(currentTime-int64(i*10), 0),
						Flagged:            true,
					}
					_ = k.SetLargeTxRecord(ctx, record)
				}
			},
			expectAlert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)
			tt.setup(k, ctx)

			params, _ := k.GetParams(ctx)
			alert, err := k.detectFlashLoanAttack(ctx, params)

			require.NoError(t, err)
			if tt.expectAlert {
				require.NotNil(t, alert)
				require.Equal(t, types.AttackTypeFlashLoan, alert.AttackType)
				require.Equal(t, types.AlertSeverity_ALERT_SEVERITY_CRITICAL, alert.Severity)
			} else {
				require.Nil(t, alert)
			}
		})
	}
}

func TestDetectFrontRunning(t *testing.T) {
	tests := []struct {
		name         string
		setupParams  func(*types.Params)
		expectAlert  bool
		alertMessage string
	}{
		{
			name: "no alert with normal gas prices",
			setupParams: func(p *types.Params) {
				p.DynamicFees.BaseFee = "1000"
				p.DynamicFees.CurrentMultiplier = 10000 // 1x multiplier (100%)
			},
			expectAlert: false,
		},
		{
			name: "alert with >2x gas price spike",
			setupParams: func(p *types.Params) {
				p.DynamicFees.BaseFee = "1000"
				p.DynamicFees.CurrentMultiplier = 25000 // 2.5x multiplier
			},
			expectAlert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.DefaultParams()
			tt.setupParams(params)

			k, ctx := setupKeeperWithCustomParams(t, params)
			currentTime := time.Now().Unix()
			_ = k.SetCurrentTime(ctx, currentTime)

			testParams, _ := k.GetParams(ctx)
			alert, err := k.detectFrontRunning(ctx, testParams)

			require.NoError(t, err)
			if tt.expectAlert {
				require.NotNil(t, alert)
				require.Equal(t, types.AttackTypeFrontRunning, alert.AttackType)
			} else {
				require.Nil(t, alert)
			}
		})
	}
}

func TestCreateAttackAlert(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	tests := []struct {
		name          string
		attackType    types.AttackType
		severity      types.AlertSeverity
		message       string
		suspect       string
		evidenceCount uint64
	}{
		{
			name:          "pump and dump alert",
			attackType:    types.AttackTypePumpAndDump,
			severity:      types.AlertSeverity_ALERT_SEVERITY_WARNING,
			message:       "Detected pump and dump pattern",
			suspect:       "aura1suspectaddress",
			evidenceCount: 5,
		},
		{
			name:          "flash loan attack alert",
			attackType:    types.AttackTypeFlashLoan,
			severity:      types.AlertSeverity_ALERT_SEVERITY_CRITICAL,
			message:       "Flash loan attack detected",
			suspect:       "aura1flashattacker",
			evidenceCount: 10,
		},
		{
			name:          "sybil attack alert",
			attackType:    types.AttackTypeSybil,
			severity:      types.AlertSeverity_ALERT_SEVERITY_WARNING,
			message:       "Potential sybil attack",
			suspect:       "",
			evidenceCount: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert, err := k.createAttackAlert(
				ctx,
				tt.attackType,
				tt.severity,
				tt.message,
				tt.suspect,
				tt.evidenceCount,
			)

			require.NoError(t, err)
			require.NotNil(t, alert)
			require.NotEmpty(t, alert.AlertId)
			require.Equal(t, tt.attackType, alert.AttackType)
			require.Equal(t, tt.severity, alert.Severity)
			require.Equal(t, tt.message, alert.Message)
			require.Equal(t, tt.suspect, alert.SuspectAddress)
			require.Equal(t, tt.evidenceCount, alert.EvidenceCount)
			require.False(t, alert.AutoMitigated)
		})
	}
}

func TestRecordAttackAlert(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	alert := &types.AttackAlert{
		AlertId:        "test-alert-123",
		AttackType:     types.AttackTypePumpAndDump,
		Severity:       types.AlertSeverity_ALERT_SEVERITY_CRITICAL,
		Message:        "Test alert message",
		DetectedAt:     time.Now(),
		SuspectAddress: "aura1suspect",
		EvidenceCount:  5,
		AutoMitigated:  false,
	}

	err := k.RecordAttackAlert(ctx, alert)
	require.NoError(t, err)
}
