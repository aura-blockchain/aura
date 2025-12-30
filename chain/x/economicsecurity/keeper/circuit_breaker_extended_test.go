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

// =============================================================================
// Liquidity Crisis Breaker Tests (Target: 41.2% -> 100%)
// =============================================================================

func TestCheckLiquidityCrisisBreaker(t *testing.T) {
	tests := []struct {
		name               string
		setupParams        func(*types.Params)
		setup              func(*Keeper, context.Context)
		disabled           bool
		expectEvent        bool
		expectedSeverity   types.AlertSeverity
		expectedMsgContain string
	}{
		{
			name: "no event when disabled",
			setupParams: func(p *types.Params) {
				p.Tokenomics.CirculatingSupply = "1000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetTotalMEVPending(ctx, "500000000000") // 50% of supply
			},
			disabled:    true,
			expectEvent: false,
		},
		{
			name: "no event with zero MEV pending",
			setupParams: func(p *types.Params) {
				p.Tokenomics.CirculatingSupply = "1000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetTotalMEVPending(ctx, "0")
			},
			expectEvent: false,
		},
		{
			name: "no event with empty MEV pending",
			setupParams: func(p *types.Params) {
				p.Tokenomics.CirculatingSupply = "1000000000000"
			},
			setup:       func(k *Keeper, ctx context.Context) {},
			expectEvent: false,
		},
		{
			name: "no event with MEV below threshold",
			setupParams: func(p *types.Params) {
				p.Tokenomics.CirculatingSupply = "1000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetTotalMEVPending(ctx, "50000000000") // 5% of supply, below 10% threshold
			},
			expectEvent: false,
		},
		{
			name: "event triggered when MEV exceeds 10% of supply",
			setupParams: func(p *types.Params) {
				p.Tokenomics.CirculatingSupply = "1000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetTotalMEVPending(ctx, "150000000000") // 15% of supply
			},
			expectEvent:        true,
			expectedSeverity:   types.AlertSeverity_ALERT_SEVERITY_WARNING,
			expectedMsgContain: "High MEV pending",
		},
		{
			name: "event triggered at exactly threshold",
			setupParams: func(p *types.Params) {
				p.Tokenomics.CirculatingSupply = "1000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetTotalMEVPending(ctx, "100000000001") // Just over 10%
			},
			expectEvent:      true,
			expectedSeverity: types.AlertSeverity_ALERT_SEVERITY_WARNING,
		},
		{
			name: "event triggered with large MEV accumulation",
			setupParams: func(p *types.Params) {
				p.Tokenomics.CirculatingSupply = "1000000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetTotalMEVPending(ctx, "500000000000000") // 50% of supply - severe
			},
			expectEvent:      true,
			expectedSeverity: types.AlertSeverity_ALERT_SEVERITY_WARNING,
		},
		{
			name: "no event with invalid MEV pending format",
			setupParams: func(p *types.Params) {
				p.Tokenomics.CirculatingSupply = "1000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetTotalMEVPending(ctx, "invalid_number")
			},
			expectEvent: false,
		},
		{
			name: "no event with negative MEV pending",
			setupParams: func(p *types.Params) {
				p.Tokenomics.CirculatingSupply = "1000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetTotalMEVPending(ctx, "-100")
			},
			expectEvent: false,
		},
		{
			name: "error with invalid circulating supply",
			setupParams: func(p *types.Params) {
				p.Tokenomics.CirculatingSupply = "invalid"
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetTotalMEVPending(ctx, "150000000000")
			},
			expectEvent: false, // Returns error, handled gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.DefaultParams()
			if tt.setupParams != nil {
				tt.setupParams(params)
			}

			k, ctx := setupKeeperWithCustomParams(t, params)
			currentTime := time.Now().Unix()
			_ = k.SetCurrentTime(ctx, currentTime)

			if tt.setup != nil {
				tt.setup(k, ctx)
			}

			testParams, _ := k.GetParams(ctx)
			config := k.getCircuitBreakerConfig()
			if tt.disabled {
				config.LiquidityCrisisEnabled = false
			}

			event, err := k.checkLiquidityCrisisBreaker(ctx, testParams, config)

			if tt.expectEvent {
				require.NoError(t, err)
				require.NotNil(t, event)
				require.Equal(t, types.CircuitBreakerTypeLiquidityCrisis, event.BreakerType)
				require.Equal(t, tt.expectedSeverity, event.Severity)
				require.True(t, event.Active)
				if tt.expectedMsgContain != "" {
					require.Contains(t, event.Message, tt.expectedMsgContain)
				}
			} else {
				require.Nil(t, event)
			}
		})
	}
}

// =============================================================================
// Multiple Circuit Breakers Triggering Simultaneously
// =============================================================================

func TestMultipleCircuitBreakersTriggering(t *testing.T) {
	tests := []struct {
		name         string
		setupParams  func(*types.Params)
		setup        func(*Keeper, context.Context)
		expectEvents int
		eventTypes   []types.CircuitBreakerType
	}{
		{
			name: "gas spike and liquidity crisis together",
			setupParams: func(p *types.Params) {
				p.DynamicFees.CurrentMultiplier = 45000
				p.DynamicFees.MaxMultiplier = 50000
				p.Tokenomics.CirculatingSupply = "1000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetTotalMEVPending(ctx, "150000000000") // 15% of supply
			},
			expectEvents: 2,
			eventTypes: []types.CircuitBreakerType{
				types.CircuitBreakerTypeLiquidityCrisis,
				types.CircuitBreakerTypeGasSpike,
			},
		},
		{
			name: "price volatility and large transaction together",
			setupParams: func(p *types.Params) {
				p.Tokenomics.CirculatingSupply = "1000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)

				// Create 6 high-volatility transactions
				for i := 0; i < 6; i++ {
					record := &types.LargeTxRecord{
						TxHash:             "tx_multi_volatile_" + string(rune('a'+i)),
						Sender:             "aura1volatile" + string(rune('a'+i)),
						Recipient:          "aura1exchange",
						Amount:             "50000000000",
						PercentageOfSupply: 200, // 2% each
						BlockHeight:        uint64(i + 1),
						Timestamp:          time.Unix(currentTime-int64(i*60), 0),
						Flagged:            true,
					}
					_ = k.SetLargeTxRecord(ctx, record)
				}

				// Add one extremely large transaction (>5% of supply)
				hugeRecord := &types.LargeTxRecord{
					TxHash:             "tx_huge_attacker",
					Sender:             "aura1attacker",
					Recipient:          "aura1victim",
					Amount:             "100000000000",
					PercentageOfSupply: 700, // 7% of supply
					BlockHeight:        100,
					Timestamp:          time.Unix(currentTime-30, 0),
					Flagged:            true,
				}
				_ = k.SetLargeTxRecord(ctx, hugeRecord)
			},
			expectEvents: 2,
			eventTypes: []types.CircuitBreakerType{
				types.CircuitBreakerTypePriceVolatility,
				types.CircuitBreakerTypeLargeTransaction,
			},
		},
		{
			name: "all breakers except supply change",
			setupParams: func(p *types.Params) {
				p.DynamicFees.CurrentMultiplier = 45000
				p.DynamicFees.MaxMultiplier = 50000
				p.Tokenomics.CirculatingSupply = "1000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)
				_ = k.SetTotalMEVPending(ctx, "150000000000") // 15% of supply

				// Price volatility transactions
				for i := 0; i < 6; i++ {
					record := &types.LargeTxRecord{
						TxHash:             "tx_all_breakers_" + string(rune('a'+i)),
						Sender:             "aura1whale" + string(rune('a'+i)),
						Recipient:          "aura1exchange",
						Amount:             "50000000000",
						PercentageOfSupply: 250,
						BlockHeight:        uint64(i + 1),
						Timestamp:          time.Unix(currentTime-int64(i*50), 0),
						Flagged:            true,
					}
					_ = k.SetLargeTxRecord(ctx, record)
				}

				// Huge transaction
				hugeRecord := &types.LargeTxRecord{
					TxHash:             "tx_all_huge",
					Sender:             "aura1megawhale",
					Recipient:          "aura1target",
					Amount:             "100000000000",
					PercentageOfSupply: 600,
					BlockHeight:        200,
					Timestamp:          time.Unix(currentTime-20, 0),
					Flagged:            true,
				}
				_ = k.SetLargeTxRecord(ctx, hugeRecord)
			},
			expectEvents: 4,
			eventTypes: []types.CircuitBreakerType{
				types.CircuitBreakerTypePriceVolatility,
				types.CircuitBreakerTypeLargeTransaction,
				types.CircuitBreakerTypeLiquidityCrisis,
				types.CircuitBreakerTypeGasSpike,
			},
		},
		{
			name: "supply change breaker with others",
			setupParams: func(p *types.Params) {
				p.DynamicFees.CurrentMultiplier = 45000
				p.DynamicFees.MaxMultiplier = 50000
				p.Tokenomics.InflationRate = 2500 // 25%
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetPreviousInflation(ctx, 1000) // Previous was 10%, now 25% - >50% change
			},
			expectEvents: 2,
			eventTypes: []types.CircuitBreakerType{
				types.CircuitBreakerTypeSupplyChange,
				types.CircuitBreakerTypeGasSpike,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.DefaultParams()
			if tt.setupParams != nil {
				tt.setupParams(params)
			}

			k, ctx := setupKeeperWithCustomParams(t, params)
			currentTime := time.Now().Unix()
			_ = k.SetCurrentTime(ctx, currentTime)

			if tt.setup != nil {
				tt.setup(k, ctx)
			}

			events, err := k.CheckCircuitBreakers(ctx)

			require.NoError(t, err)
			require.Len(t, events, tt.expectEvents)

			// Verify expected event types are present
			eventTypeMap := make(map[types.CircuitBreakerType]bool)
			for _, event := range events {
				eventTypeMap[event.BreakerType] = true
			}

			for _, expectedType := range tt.eventTypes {
				require.True(t, eventTypeMap[expectedType], "Expected event type %v not found", expectedType)
			}
		})
	}
}

// =============================================================================
// Circuit Breaker Activation and Deactivation Cycles
// =============================================================================

func TestCircuitBreakerActivationDeactivationCycles(t *testing.T) {
	t.Run("activate and deactivate price volatility breaker", func(t *testing.T) {
		k, ctx := setupKeeperForTest(t)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Activate breaker
		err := k.ActivateCircuitBreaker(ctx, types.CircuitBreakerTypePriceVolatility, "Manual activation for testing")
		require.NoError(t, err)

		// Deactivate breaker
		err = k.DeactivateCircuitBreaker(ctx, "manual_cb_"+string(rune(currentTime)))
		require.NoError(t, err)
	})

	t.Run("activate multiple breakers sequentially", func(t *testing.T) {
		k, ctx := setupKeeperForTest(t)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		breakerTypes := []types.CircuitBreakerType{
			types.CircuitBreakerTypePriceVolatility,
			types.CircuitBreakerTypeLargeTransaction,
			types.CircuitBreakerTypeSupplyChange,
			types.CircuitBreakerTypeLiquidityCrisis,
			types.CircuitBreakerTypeGasSpike,
		}

		reasons := []string{
			"Market flash crash detected",
			"Whale manipulation suspected",
			"Unexpected inflation spike",
			"Liquidity pool drained",
			"Network congestion attack",
		}

		for i, bt := range breakerTypes {
			err := k.ActivateCircuitBreaker(ctx, bt, reasons[i])
			require.NoError(t, err)
		}
	})

	t.Run("deactivate non-existent breaker", func(t *testing.T) {
		k, ctx := setupKeeperForTest(t)

		err := k.DeactivateCircuitBreaker(ctx, "non_existent_breaker_id")
		require.NoError(t, err) // Should not error for non-existent breakers
	})

	t.Run("activate same breaker type multiple times", func(t *testing.T) {
		k, ctx := setupKeeperForTest(t)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Activate same breaker type multiple times with different reasons
		err := k.ActivateCircuitBreaker(ctx, types.CircuitBreakerTypeLiquidityCrisis, "First crisis")
		require.NoError(t, err)

		_ = k.SetCurrentTime(ctx, currentTime+1)
		err = k.ActivateCircuitBreaker(ctx, types.CircuitBreakerTypeLiquidityCrisis, "Second crisis")
		require.NoError(t, err)

		_ = k.SetCurrentTime(ctx, currentTime+2)
		err = k.ActivateCircuitBreaker(ctx, types.CircuitBreakerTypeLiquidityCrisis, "Third crisis")
		require.NoError(t, err)
	})
}

// =============================================================================
// Circuit Breaker Configuration Edge Cases
// =============================================================================

func TestCircuitBreakerConfigEdgeCases(t *testing.T) {
	t.Run("all breakers disabled", func(t *testing.T) {
		params := types.DefaultParams()
		params.DynamicFees.CurrentMultiplier = 50000
		params.DynamicFees.MaxMultiplier = 50000
		params.Tokenomics.CirculatingSupply = "1000000000000"
		params.Tokenomics.InflationRate = 5000

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Set up conditions that would trigger all breakers
		_ = k.SetPreviousInflation(ctx, 1000)
		_ = k.SetTotalMEVPending(ctx, "500000000000")

		for i := 0; i < 10; i++ {
			record := &types.LargeTxRecord{
				TxHash:             "tx_disabled_" + string(rune('a'+i)),
				Sender:             "aura1whale" + string(rune('a'+i)),
				Recipient:          "aura1exchange",
				Amount:             "50000000000",
				PercentageOfSupply: 600,
				BlockHeight:        uint64(i + 1),
				Timestamp:          time.Unix(currentTime-int64(i*30), 0),
				Flagged:            true,
			}
			_ = k.SetLargeTxRecord(ctx, record)
		}

		// Manually check each breaker with disabled config
		testParams, _ := k.GetParams(ctx)
		config := &types.CircuitBreakerConfig{
			PriceVolatilityEnabled:  false,
			LargeTransactionEnabled: false,
			SupplyChangeEnabled:     false,
			LiquidityCrisisEnabled:  false,
			GasSpikeEnabled:         false,
		}

		event1, err := k.checkPriceVolatilityBreaker(ctx, testParams, config)
		require.NoError(t, err)
		require.Nil(t, event1)

		event2, err := k.checkLargeTransactionBreaker(ctx, testParams, config)
		require.NoError(t, err)
		require.Nil(t, event2)

		event3, err := k.checkSupplyChangeBreaker(ctx, testParams, config)
		require.NoError(t, err)
		require.Nil(t, event3)

		event4, err := k.checkLiquidityCrisisBreaker(ctx, testParams, config)
		require.NoError(t, err)
		require.Nil(t, event4)

		event5, err := k.checkGasSpikeBreaker(ctx, testParams, config)
		require.NoError(t, err)
		require.Nil(t, event5)
	})

	t.Run("selective breaker enabling", func(t *testing.T) {
		params := types.DefaultParams()
		params.DynamicFees.CurrentMultiplier = 50000
		params.DynamicFees.MaxMultiplier = 50000
		params.Tokenomics.CirculatingSupply = "1000000000000"

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		_ = k.SetTotalMEVPending(ctx, "500000000000")

		testParams, _ := k.GetParams(ctx)

		// Only gas spike enabled
		config := &types.CircuitBreakerConfig{
			PriceVolatilityEnabled:  false,
			LargeTransactionEnabled: false,
			SupplyChangeEnabled:     false,
			LiquidityCrisisEnabled:  false,
			GasSpikeEnabled:         true,
		}

		gasEvent, err := k.checkGasSpikeBreaker(ctx, testParams, config)
		require.NoError(t, err)
		require.NotNil(t, gasEvent)

		liquidityEvent, err := k.checkLiquidityCrisisBreaker(ctx, testParams, config)
		require.NoError(t, err)
		require.Nil(t, liquidityEvent)
	})

	t.Run("zero values in params", func(t *testing.T) {
		params := types.DefaultParams()
		params.DynamicFees.CurrentMultiplier = 0
		params.DynamicFees.MaxMultiplier = 0
		params.Tokenomics.CirculatingSupply = "0"

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		events, err := k.CheckCircuitBreakers(ctx)
		require.NoError(t, err)
		require.NotNil(t, events)
	})

	t.Run("extreme values in params", func(t *testing.T) {
		params := types.DefaultParams()
		params.DynamicFees.CurrentMultiplier = 18446744073709551615 // Max uint64
		params.DynamicFees.MaxMultiplier = 18446744073709551615
		params.Tokenomics.CirculatingSupply = "999999999999999999999999999999"

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		events, err := k.CheckCircuitBreakers(ctx)
		require.NoError(t, err)
		require.NotNil(t, events)
	})
}

// =============================================================================
// Recovery from Triggered States
// =============================================================================

func TestRecoveryFromTriggeredStates(t *testing.T) {
	t.Run("gas spike recovery - multiplier decreases", func(t *testing.T) {
		// Initial state: gas spike triggered
		params := types.DefaultParams()
		params.DynamicFees.CurrentMultiplier = 45000
		params.DynamicFees.MaxMultiplier = 50000

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		events, err := k.CheckCircuitBreakers(ctx)
		require.NoError(t, err)
		require.Len(t, events, 1)
		require.Equal(t, types.CircuitBreakerTypeGasSpike, events[0].BreakerType)

		// Recovery: gas multiplier decreases
		params.DynamicFees.CurrentMultiplier = 30000 // Below 80% threshold
		_ = k.SetParams(*params)

		events, err = k.CheckCircuitBreakers(ctx)
		require.NoError(t, err)
		require.Len(t, events, 0)
	})

	t.Run("liquidity crisis recovery - MEV pending decreases", func(t *testing.T) {
		params := types.DefaultParams()
		params.Tokenomics.CirculatingSupply = "1000000000000"

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Initial state: liquidity crisis
		_ = k.SetTotalMEVPending(ctx, "150000000000")

		events, err := k.CheckCircuitBreakers(ctx)
		require.NoError(t, err)

		hasLiquidityCrisis := false
		for _, e := range events {
			if e.BreakerType == types.CircuitBreakerTypeLiquidityCrisis {
				hasLiquidityCrisis = true
				break
			}
		}
		require.True(t, hasLiquidityCrisis)

		// Recovery: MEV pending decreases
		_ = k.SetTotalMEVPending(ctx, "50000000000") // 5% - below threshold

		events, err = k.CheckCircuitBreakers(ctx)
		require.NoError(t, err)

		hasLiquidityCrisis = false
		for _, e := range events {
			if e.BreakerType == types.CircuitBreakerTypeLiquidityCrisis {
				hasLiquidityCrisis = true
				break
			}
		}
		require.False(t, hasLiquidityCrisis)
	})

	t.Run("supply change recovery - inflation stabilizes", func(t *testing.T) {
		params := types.DefaultParams()
		params.Tokenomics.InflationRate = 2500 // 25%

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Initial state: rapid inflation change
		_ = k.SetPreviousInflation(ctx, 1000) // Previous 10%, now 25% - >50% change

		events, err := k.CheckCircuitBreakers(ctx)
		require.NoError(t, err)

		hasSupplyChange := false
		for _, e := range events {
			if e.BreakerType == types.CircuitBreakerTypeSupplyChange {
				hasSupplyChange = true
				break
			}
		}
		require.True(t, hasSupplyChange)

		// Recovery: inflation stabilizes
		_ = k.SetPreviousInflation(ctx, 2400) // Now only 4% change from 2500

		events, err = k.CheckCircuitBreakers(ctx)
		require.NoError(t, err)

		hasSupplyChange = false
		for _, e := range events {
			if e.BreakerType == types.CircuitBreakerTypeSupplyChange {
				hasSupplyChange = true
				break
			}
		}
		require.False(t, hasSupplyChange)
	})

	t.Run("price volatility recovery - transactions age out", func(t *testing.T) {
		params := types.DefaultParams()

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Create volatile transactions from 2 hours ago (outside 1 hour window)
		for i := 0; i < 6; i++ {
			record := &types.LargeTxRecord{
				TxHash:             "tx_aged_" + string(rune('a'+i)),
				Sender:             "aura1whale" + string(rune('a'+i)),
				Recipient:          "aura1exchange",
				Amount:             "50000000000",
				PercentageOfSupply: 200,
				BlockHeight:        uint64(i + 1),
				Timestamp:          time.Unix(currentTime-7200-int64(i*100), 0), // 2+ hours ago
				Flagged:            true,
			}
			_ = k.SetLargeTxRecord(ctx, record)
		}

		events, err := k.CheckCircuitBreakers(ctx)
		require.NoError(t, err)

		hasPriceVolatility := false
		for _, e := range events {
			if e.BreakerType == types.CircuitBreakerTypePriceVolatility {
				hasPriceVolatility = true
				break
			}
		}
		require.False(t, hasPriceVolatility) // Old transactions should not trigger
	})

	t.Run("large transaction recovery - transaction ages out", func(t *testing.T) {
		params := types.DefaultParams()

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Create huge transaction from 10 minutes ago (outside 5 minute window)
		record := &types.LargeTxRecord{
			TxHash:             "tx_old_huge",
			Sender:             "aura1oldwhale",
			Recipient:          "aura1target",
			Amount:             "100000000000",
			PercentageOfSupply: 700, // 7% of supply
			BlockHeight:        1,
			Timestamp:          time.Unix(currentTime-600, 0), // 10 minutes ago
			Flagged:            true,
		}
		_ = k.SetLargeTxRecord(ctx, record)

		events, err := k.CheckCircuitBreakers(ctx)
		require.NoError(t, err)

		hasLargeTx := false
		for _, e := range events {
			if e.BreakerType == types.CircuitBreakerTypeLargeTransaction {
				hasLargeTx = true
				break
			}
		}
		require.False(t, hasLargeTx) // Old transaction should not trigger
	})
}

// =============================================================================
// Circuit Breaker Event Creation Tests
// =============================================================================

func TestCreateCircuitBreakerEventDetails(t *testing.T) {
	tests := []struct {
		name         string
		breakerType  types.CircuitBreakerType
		severity     types.AlertSeverity
		message      string
		currentValue string
		threshold    string
	}{
		{
			name:         "price volatility event",
			breakerType:  types.CircuitBreakerTypePriceVolatility,
			severity:     types.AlertSeverity_ALERT_SEVERITY_CRITICAL,
			message:      "Extreme price movement detected",
			currentValue: "350",
			threshold:    "100",
		},
		{
			name:         "large transaction event",
			breakerType:  types.CircuitBreakerTypeLargeTransaction,
			severity:     types.AlertSeverity_ALERT_SEVERITY_CRITICAL,
			message:      "Suspicious whale activity",
			currentValue: "750",
			threshold:    "500",
		},
		{
			name:         "supply change event",
			breakerType:  types.CircuitBreakerTypeSupplyChange,
			severity:     types.AlertSeverity_ALERT_SEVERITY_WARNING,
			message:      "Inflation rate changed rapidly",
			currentValue: "7500",
			threshold:    "5000",
		},
		{
			name:         "liquidity crisis event",
			breakerType:  types.CircuitBreakerTypeLiquidityCrisis,
			severity:     types.AlertSeverity_ALERT_SEVERITY_WARNING,
			message:      "High MEV accumulation detected",
			currentValue: "150000000000",
			threshold:    "100000000000",
		},
		{
			name:         "gas spike event",
			breakerType:  types.CircuitBreakerTypeGasSpike,
			severity:     types.AlertSeverity_ALERT_SEVERITY_WARNING,
			message:      "Gas price nearing maximum",
			currentValue: "45000",
			threshold:    "40000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)
			currentTime := time.Now().Unix()
			_ = k.SetCurrentTime(ctx, currentTime)

			event, err := k.createCircuitBreakerEvent(
				ctx,
				tt.breakerType,
				tt.severity,
				tt.message,
				tt.currentValue,
				tt.threshold,
			)

			require.NoError(t, err)
			require.NotNil(t, event)
			require.Contains(t, event.BreakerId, "cb_")
			require.Equal(t, tt.breakerType, event.BreakerType)
			require.Equal(t, tt.severity, event.Severity)
			require.Equal(t, tt.message, event.Message)
			require.Equal(t, tt.currentValue, event.CurrentValue)
			require.Equal(t, tt.threshold, event.Threshold)
			require.True(t, event.Active)
			require.False(t, event.AutoMitigated)
		})
	}
}

// =============================================================================
// Statistics and History Functions Tests
// =============================================================================

func TestCircuitBreakerStatisticsAndHistory(t *testing.T) {
	t.Run("get statistics returns zeros initially", func(t *testing.T) {
		k, _ := setupKeeperForTest(t)

		total, active, autoMitigated, manual := k.GetCircuitBreakerStatistics()

		require.Equal(t, uint64(0), total)
		require.Equal(t, uint64(0), active)
		require.Equal(t, uint64(0), autoMitigated)
		require.Equal(t, uint64(0), manual)
	})

	t.Run("get history returns empty initially", func(t *testing.T) {
		k, _ := setupKeeperForTest(t)

		history := k.GetCircuitBreakerHistory(100)

		require.NotNil(t, history)
		require.Len(t, history, 0)
	})

	t.Run("get active breakers returns empty initially", func(t *testing.T) {
		k, _ := setupKeeperForTest(t)

		active := k.GetActiveCircuitBreakers()

		require.NotNil(t, active)
		require.Len(t, active, 0)
	})

	t.Run("get history with different limits", func(t *testing.T) {
		k, _ := setupKeeperForTest(t)

		history0 := k.GetCircuitBreakerHistory(0)
		require.NotNil(t, history0)

		history1 := k.GetCircuitBreakerHistory(1)
		require.NotNil(t, history1)

		history1000 := k.GetCircuitBreakerHistory(1000)
		require.NotNil(t, history1000)
	})
}

// =============================================================================
// Boundary Condition Tests
// =============================================================================

func TestCircuitBreakerBoundaryConditions(t *testing.T) {
	t.Run("price volatility exactly at threshold", func(t *testing.T) {
		params := types.DefaultParams()

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Create exactly 5 transactions (minimum for check)
		for i := 0; i < 5; i++ {
			record := &types.LargeTxRecord{
				TxHash:             "tx_boundary_" + string(rune('a'+i)),
				Sender:             "aura1boundary" + string(rune('a'+i)),
				Recipient:          "aura1exchange",
				Amount:             "10000000000",
				PercentageOfSupply: 100, // Exactly 1%
				BlockHeight:        uint64(i + 1),
				Timestamp:          time.Unix(currentTime-int64(i*100), 0),
				Flagged:            false,
			}
			_ = k.SetLargeTxRecord(ctx, record)
		}

		testParams, _ := k.GetParams(ctx)
		config := k.getCircuitBreakerConfig()

		event, err := k.checkPriceVolatilityBreaker(ctx, testParams, config)

		require.NoError(t, err)
		require.Nil(t, event) // Exactly 100 is not > 100
	})

	t.Run("price volatility just above threshold", func(t *testing.T) {
		params := types.DefaultParams()

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Create 5 transactions just above threshold
		for i := 0; i < 5; i++ {
			record := &types.LargeTxRecord{
				TxHash:             "tx_above_" + string(rune('a'+i)),
				Sender:             "aura1above" + string(rune('a'+i)),
				Recipient:          "aura1exchange",
				Amount:             "10100000000",
				PercentageOfSupply: 101, // Just above 1%
				BlockHeight:        uint64(i + 1),
				Timestamp:          time.Unix(currentTime-int64(i*100), 0),
				Flagged:            true,
			}
			_ = k.SetLargeTxRecord(ctx, record)
		}

		testParams, _ := k.GetParams(ctx)
		config := k.getCircuitBreakerConfig()

		event, err := k.checkPriceVolatilityBreaker(ctx, testParams, config)

		require.NoError(t, err)
		require.NotNil(t, event) // 101 > 100, triggers
	})

	t.Run("large transaction exactly at threshold", func(t *testing.T) {
		params := types.DefaultParams()

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		record := &types.LargeTxRecord{
			TxHash:             "tx_exact_threshold",
			Sender:             "aura1exact",
			Recipient:          "aura1target",
			Amount:             "50000000000",
			PercentageOfSupply: 500, // Exactly 5%
			BlockHeight:        1,
			Timestamp:          time.Unix(currentTime-60, 0),
			Flagged:            true,
		}
		_ = k.SetLargeTxRecord(ctx, record)

		testParams, _ := k.GetParams(ctx)
		config := k.getCircuitBreakerConfig()

		event, err := k.checkLargeTransactionBreaker(ctx, testParams, config)

		require.NoError(t, err)
		require.Nil(t, event) // Exactly 500 is not > 500
	})

	t.Run("gas spike exactly at 80% threshold", func(t *testing.T) {
		params := types.DefaultParams()
		params.DynamicFees.CurrentMultiplier = 40000 // Exactly 80% of 50000
		params.DynamicFees.MaxMultiplier = 50000

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		testParams, _ := k.GetParams(ctx)
		config := k.getCircuitBreakerConfig()

		event, err := k.checkGasSpikeBreaker(ctx, testParams, config)

		require.NoError(t, err)
		require.Nil(t, event) // Exactly at threshold is not > threshold
	})

	t.Run("supply change exactly at 50% threshold", func(t *testing.T) {
		params := types.DefaultParams()
		params.Tokenomics.InflationRate = 1500 // 15%

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		_ = k.SetPreviousInflation(ctx, 1000) // 10% -> 15% = 50% change

		testParams, _ := k.GetParams(ctx)
		config := k.getCircuitBreakerConfig()

		event, err := k.checkSupplyChangeBreaker(ctx, testParams, config)

		require.NoError(t, err)
		require.Nil(t, event) // Exactly 5000 is not > 5000
	})
}

// =============================================================================
// Circuit Breaker Type String Tests
// =============================================================================

func TestCircuitBreakerTypeStrings(t *testing.T) {
	tests := []struct {
		breakerType types.CircuitBreakerType
		expected    string
	}{
		{types.CircuitBreakerTypeUnspecified, "UNSPECIFIED"},
		{types.CircuitBreakerTypePriceVolatility, "PRICE_VOLATILITY"},
		{types.CircuitBreakerTypeLargeTransaction, "LARGE_TRANSACTION"},
		{types.CircuitBreakerTypeSupplyChange, "SUPPLY_CHANGE"},
		{types.CircuitBreakerTypeLiquidityCrisis, "LIQUIDITY_CRISIS"},
		{types.CircuitBreakerTypeGasSpike, "GAS_SPIKE"},
		{types.CircuitBreakerType(99), "UNSPECIFIED"}, // Unknown type
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.breakerType.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// Stress Tests
// =============================================================================

func TestCircuitBreakerStressConditions(t *testing.T) {
	t.Run("many large transactions in window", func(t *testing.T) {
		params := types.DefaultParams()

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Create 100 transactions
		for i := 0; i < 100; i++ {
			record := &types.LargeTxRecord{
				TxHash:             "tx_stress_" + string(rune('a'+i%26)) + string(rune('0'+i/26)),
				Sender:             "aura1stress" + string(rune('a'+i%26)),
				Recipient:          "aura1exchange",
				Amount:             "1000000000",
				PercentageOfSupply: 50,
				BlockHeight:        uint64(i + 1),
				Timestamp:          time.Unix(currentTime-int64(i*10), 0),
				Flagged:            false,
			}
			_ = k.SetLargeTxRecord(ctx, record)
		}

		events, err := k.CheckCircuitBreakers(ctx)
		require.NoError(t, err)
		require.NotNil(t, events)
	})

	t.Run("rapid circuit breaker checks", func(t *testing.T) {
		params := types.DefaultParams()
		params.DynamicFees.CurrentMultiplier = 45000
		params.DynamicFees.MaxMultiplier = 50000

		k, ctx := setupKeeperWithCustomParams(t, params)
		currentTime := time.Now().Unix()
		_ = k.SetCurrentTime(ctx, currentTime)

		// Call CheckCircuitBreakers many times
		for i := 0; i < 50; i++ {
			events, err := k.CheckCircuitBreakers(ctx)
			require.NoError(t, err)
			require.NotNil(t, events)
		}
	})
}
