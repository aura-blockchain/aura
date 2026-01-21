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

func TestCheckCircuitBreakers(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*Keeper, context.Context)
		expectEvents int
		wantErr      bool
	}{
		{
			name:         "no circuit breakers triggered with normal state",
			setup:        func(k *Keeper, ctx context.Context) {},
			expectEvents: 0,
			wantErr:      false,
		},
		{
			name: "gas spike circuit breaker triggered",
			setup: func(k *Keeper, ctx context.Context) {
				params, _ := k.GetParams(ctx)
				params.DynamicFees.CurrentMultiplier = 50000 // Very high multiplier
				params.DynamicFees.MaxMultiplier = 50000
				_ = k.SetParams(params)
			},
			expectEvents: 1, // Gas spike breaker
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)
			currentTime := time.Now().Unix()
			_ = k.SetCurrentTime(ctx, currentTime)

			tt.setup(k, ctx)

			events, err := k.CheckCircuitBreakers(ctx)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, events, tt.expectEvents)
			}
		})
	}
}

func TestCheckPriceVolatilityBreaker(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*Keeper, context.Context)
		disabled    bool
		expectEvent bool
	}{
		{
			name:        "no event with few transactions",
			setup:       func(k *Keeper, ctx context.Context) {},
			expectEvent: false,
		},
		{
			name: "event triggered with high volatility",
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)

				// Create 6 transactions with avg >1% of supply
				for i := 0; i < 6; i++ {
					record := &types.LargeTxRecord{
						TxHash:             "tx_volatile_" + string(rune('a'+i)),
						Sender:             "aura1whale" + string(rune('a'+i)),
						Recipient:          "aura1exchange",
						Amount:             "50000000000",
						PercentageOfSupply: 200, // 2% each
						BlockHeight:        uint64(i + 1),
						Timestamp:          time.Unix(currentTime-int64(i*100), 0),
						Flagged:            true,
					}
					_ = k.SetLargeTxRecord(ctx, record)
				}
			},
			expectEvent: true,
		},
		{
			name: "no event when disabled",
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)

				for i := 0; i < 6; i++ {
					record := &types.LargeTxRecord{
						TxHash:             "tx_disabled_" + string(rune('a'+i)),
						Sender:             "aura1whale" + string(rune('a'+i)),
						Recipient:          "aura1exchange",
						Amount:             "50000000000",
						PercentageOfSupply: 200,
						BlockHeight:        uint64(i + 1),
						Timestamp:          time.Unix(currentTime-int64(i*100), 0),
						Flagged:            true,
					}
					_ = k.SetLargeTxRecord(ctx, record)
				}
			},
			disabled:    true,
			expectEvent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)
			currentTime := time.Now().Unix()
			_ = k.SetCurrentTime(ctx, currentTime)

			tt.setup(k, ctx)

			params, _ := k.GetParams(ctx)
			config := k.getCircuitBreakerConfig()
			if tt.disabled {
				config.PriceVolatilityEnabled = false
			}

			event, err := k.checkPriceVolatilityBreaker(ctx, params, config)

			require.NoError(t, err)
			if tt.expectEvent {
				require.NotNil(t, event)
				require.Equal(t, types.CircuitBreakerTypePriceVolatility, event.BreakerType)
				require.Equal(t, types.AlertSeverity_ALERT_SEVERITY_CRITICAL, event.Severity)
				require.True(t, event.Active)
			} else {
				require.Nil(t, event)
			}
		})
	}
}

func TestCheckLargeTransactionBreaker(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*Keeper, context.Context)
		expectEvent bool
	}{
		{
			name:        "no event with no large transactions",
			setup:       func(k *Keeper, ctx context.Context) {},
			expectEvent: false,
		},
		{
			name: "no event with transactions below threshold",
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)

				record := &types.LargeTxRecord{
					TxHash:             "tx_normal",
					Sender:             "aura1sender",
					Recipient:          "aura1recipient",
					Amount:             "1000000",
					PercentageOfSupply: 100, // 1%, below 5% threshold
					BlockHeight:        1,
					Timestamp:          time.Unix(currentTime-60, 0),
					Flagged:            false,
				}
				_ = k.SetLargeTxRecord(ctx, record)
			},
			expectEvent: false,
		},
		{
			name: "event triggered with >5% transaction",
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)

				record := &types.LargeTxRecord{
					TxHash:             "tx_huge",
					Sender:             "aura1whaleattacker",
					Recipient:          "aura1victim",
					Amount:             "500000000000",
					PercentageOfSupply: 600, // 6%, above 5% threshold
					BlockHeight:        1,
					Timestamp:          time.Unix(currentTime-60, 0),
					Flagged:            true,
				}
				_ = k.SetLargeTxRecord(ctx, record)
			},
			expectEvent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)
			currentTime := time.Now().Unix()
			_ = k.SetCurrentTime(ctx, currentTime)

			tt.setup(k, ctx)

			params, _ := k.GetParams(ctx)
			config := k.getCircuitBreakerConfig()

			event, err := k.checkLargeTransactionBreaker(ctx, params, config)

			require.NoError(t, err)
			if tt.expectEvent {
				require.NotNil(t, event)
				require.Equal(t, types.CircuitBreakerTypeLargeTransaction, event.BreakerType)
				require.Equal(t, types.AlertSeverity_ALERT_SEVERITY_CRITICAL, event.Severity)
			} else {
				require.Nil(t, event)
			}
		})
	}
}

func TestCheckSupplyChangeBreaker(t *testing.T) {
	tests := []struct {
		name              string
		currentInflation  uint64
		previousInflation uint64
		expectEvent       bool
	}{
		{
			name:              "no event with stable inflation",
			currentInflation:  1000,
			previousInflation: 1000,
			expectEvent:       false,
		},
		{
			name:              "no event with no previous inflation",
			currentInflation:  1000,
			previousInflation: 0,
			expectEvent:       false,
		},
		{
			name:              "no event with small change",
			currentInflation:  1100,
			previousInflation: 1000,
			expectEvent:       false,
		},
		{
			name:              "event triggered with >50% increase",
			currentInflation:  2000,
			previousInflation: 1000,
			expectEvent:       true,
		},
		{
			name:              "event triggered with >50% decrease",
			currentInflation:  400,
			previousInflation: 1000,
			expectEvent:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.DefaultParams()
			params.Tokenomics.InflationRate = tt.currentInflation

			k, ctx := setupKeeperWithCustomParams(t, params)
			if tt.previousInflation > 0 {
				_ = k.SetPreviousInflation(ctx, tt.previousInflation)
			}

			testParams, _ := k.GetParams(ctx)
			config := k.getCircuitBreakerConfig()

			event, err := k.checkSupplyChangeBreaker(ctx, testParams, config)

			require.NoError(t, err)
			if tt.expectEvent {
				require.NotNil(t, event)
				require.Equal(t, types.CircuitBreakerTypeSupplyChange, event.BreakerType)
			} else {
				require.Nil(t, event)
			}
		})
	}
}

func TestCheckGasSpikeBreaker(t *testing.T) {
	tests := []struct {
		name              string
		currentMultiplier uint64
		maxMultiplier     uint64
		expectEvent       bool
	}{
		{
			name:              "no event with normal gas",
			currentMultiplier: 10000, // 1x
			maxMultiplier:     50000, // 5x max
			expectEvent:       false,
		},
		{
			name:              "event triggered at 80%+ of max",
			currentMultiplier: 45000, // 4.5x
			maxMultiplier:     50000, // 5x max (80% = 40000)
			expectEvent:       true,
		},
		{
			name:              "event triggered exactly at threshold",
			currentMultiplier: 40001, // Just above 80%
			maxMultiplier:     50000,
			expectEvent:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.DefaultParams()
			params.DynamicFees.CurrentMultiplier = tt.currentMultiplier
			params.DynamicFees.MaxMultiplier = tt.maxMultiplier

			k, ctx := setupKeeperWithCustomParams(t, params)

			testParams, _ := k.GetParams(ctx)
			config := k.getCircuitBreakerConfig()

			event, err := k.checkGasSpikeBreaker(ctx, testParams, config)

			require.NoError(t, err)
			if tt.expectEvent {
				require.NotNil(t, event)
				require.Equal(t, types.CircuitBreakerTypeGasSpike, event.BreakerType)
				require.Equal(t, types.AlertSeverity_ALERT_SEVERITY_WARNING, event.Severity)
			} else {
				require.Nil(t, event)
			}
		})
	}
}

func TestActivateCircuitBreaker(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	tests := []struct {
		name        string
		breakerType types.CircuitBreakerType
		reason      string
		wantErr     bool
	}{
		{
			name:        "activate price volatility breaker",
			breakerType: types.CircuitBreakerTypePriceVolatility,
			reason:      "Emergency market conditions",
			wantErr:     false,
		},
		{
			name:        "activate liquidity crisis breaker",
			breakerType: types.CircuitBreakerTypeLiquidityCrisis,
			reason:      "Liquidity pool drained",
			wantErr:     false,
		},
		{
			name:        "activate gas spike breaker",
			breakerType: types.CircuitBreakerTypeGasSpike,
			reason:      "Network congestion attack",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k.ActivateCircuitBreaker(ctx, tt.breakerType, tt.reason)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeactivateCircuitBreaker(t *testing.T) {
	k, ctx := setupKeeperForTest(t)

	err := k.DeactivateCircuitBreaker(ctx, "test-breaker-id")
	require.NoError(t, err)
}

func TestCreateCircuitBreakerEvent(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	event, err := k.createCircuitBreakerEvent(
		ctx,
		types.CircuitBreakerTypePriceVolatility,
		types.AlertSeverity_ALERT_SEVERITY_CRITICAL,
		"High price volatility detected",
		"250",
		"100",
	)

	require.NoError(t, err)
	require.NotNil(t, event)
	require.NotEmpty(t, event.BreakerId)
	require.Equal(t, types.CircuitBreakerTypePriceVolatility, event.BreakerType)
	require.Equal(t, types.AlertSeverity_ALERT_SEVERITY_CRITICAL, event.Severity)
	require.Equal(t, "High price volatility detected", event.Message)
	require.Equal(t, "250", event.CurrentValue)
	require.Equal(t, "100", event.Threshold)
	require.True(t, event.Active)
	require.False(t, event.AutoMitigated)
}

func TestGetCircuitBreakerConfig(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	config := k.getCircuitBreakerConfig()

	require.NotNil(t, config)
	require.True(t, config.PriceVolatilityEnabled)
	require.True(t, config.LargeTransactionEnabled)
	require.True(t, config.SupplyChangeEnabled)
	require.True(t, config.LiquidityCrisisEnabled)
	require.True(t, config.GasSpikeEnabled)
	require.Equal(t, uint64(0), config.TotalTriggered)
}

func TestGetActiveCircuitBreakers(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	breakers := k.GetActiveCircuitBreakers()
	require.NotNil(t, breakers)
	require.Len(t, breakers, 0)
}

func TestGetCircuitBreakerStatistics(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	total, active, autoMitigated, manual := k.GetCircuitBreakerStatistics()

	require.Equal(t, uint64(0), total)
	require.Equal(t, uint64(0), active)
	require.Equal(t, uint64(0), autoMitigated)
	require.Equal(t, uint64(0), manual)
}

func TestGetCircuitBreakerHistory(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	history := k.GetCircuitBreakerHistory(10)
	require.NotNil(t, history)
	require.Len(t, history, 0)
}
