// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// DEXMetrics holds all Prometheus metrics for the DEX module
type DEXMetrics struct {
	// Swap metrics
	SwapsTotal        *prometheus.CounterVec
	SwapVolume        *prometheus.CounterVec
	SwapLatency       prometheus.Histogram
	SwapSlippage      prometheus.Histogram
	SwapFeesCollected *prometheus.CounterVec

	// Orderbook metrics
	OrderbookOrdersPlaced    *prometheus.CounterVec
	OrderbookOrdersFilled    *prometheus.CounterVec
	OrderbookOrdersCancelled prometheus.Counter
	LimitOrdersActive        *prometheus.GaugeVec
	MarketOrdersExecuted     prometheus.Counter

	// Liquidity metrics
	LiquidityAdded   *prometheus.CounterVec
	LiquidityRemoved *prometheus.CounterVec
	PoolReserves     *prometheus.GaugeVec
	LPTokenSupply    *prometheus.GaugeVec
	PoolTVL          *prometheus.GaugeVec

	// Pool metrics
	PoolsTotal         prometheus.Gauge
	PoolCreationRate   prometheus.Counter
	PoolImbalanceRatio *prometheus.GaugeVec
	PoolFeeTier        *prometheus.GaugeVec

	// HTLC metrics
	HTLCCreated  prometheus.Counter
	HTLCClaimed  prometheus.Counter
	HTLCRefunded prometheus.Counter
	HTLCActive   prometheus.Gauge

	// Security metrics
	CircuitBreakerActive   *prometheus.GaugeVec
	CircuitBreakerTriggers *prometheus.CounterVec
	MEVProtections         *prometheus.CounterVec
	RateLimitExceeds       *prometheus.CounterVec
	SuspiciousActivity     *prometheus.CounterVec

	// TWAP metrics
	TWAPUpdates               prometheus.Counter
	TWAPValue                 *prometheus.GaugeVec
	PriceManipulationAttempts prometheus.Counter

	// Cross-chain metrics
	IBCSwapsSent      *prometheus.CounterVec
	IBCSwapsReceived  *prometheus.CounterVec
	IBCTimeouts       *prometheus.CounterVec
	CrossChainLatency *prometheus.HistogramVec
	IBCPacketsPending prometheus.Gauge
}

var (
	dexMetricsOnce sync.Once
	dexMetrics     *DEXMetrics
)

// NewDEXMetrics creates and registers DEX metrics (singleton pattern)
func NewDEXMetrics() *DEXMetrics {
	dexMetricsOnce.Do(func() {
		dexMetrics = &DEXMetrics{
			// Swap metrics
			SwapsTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "swaps_total",
					Help:      "Total number of swaps executed",
				},
				[]string{"pool", "status"},
			),
			SwapVolume: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "swap_volume_total",
					Help:      "Total swap volume",
				},
				[]string{"denom"},
			),
			SwapLatency: promauto.NewHistogram(
				prometheus.HistogramOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "swap_latency_seconds",
					Help:      "Swap execution latency",
					Buckets:   prometheus.DefBuckets,
				},
			),
			SwapSlippage: promauto.NewHistogram(
				prometheus.HistogramOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "swap_slippage_percent",
					Help:      "Swap slippage percentage",
					Buckets:   []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
				},
			),
			SwapFeesCollected: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "swap_fees_collected_total",
					Help:      "Total swap fees collected",
				},
				[]string{"pool", "denom"},
			),

			// Orderbook metrics
			OrderbookOrdersPlaced: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "orderbook_orders_placed_total",
					Help:      "Orders placed on orderbook",
				},
				[]string{"type"},
			),
			OrderbookOrdersFilled: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "orderbook_orders_filled_total",
					Help:      "Orders filled from orderbook",
				},
				[]string{"type"},
			),
			OrderbookOrdersCancelled: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "orderbook_orders_cancelled_total",
					Help:      "Orders cancelled",
				},
			),
			LimitOrdersActive: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "limit_orders_active",
					Help:      "Active limit orders",
				},
				[]string{"pool"},
			),
			MarketOrdersExecuted: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "market_orders_executed_total",
					Help:      "Market orders executed",
				},
			),

			// Liquidity metrics
			LiquidityAdded: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "liquidity_added_total",
					Help:      "Total liquidity added",
				},
				[]string{"pool", "denom"},
			),
			LiquidityRemoved: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "liquidity_removed_total",
					Help:      "Total liquidity removed",
				},
				[]string{"pool", "denom"},
			),
			PoolReserves: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "pool_reserves",
					Help:      "Current pool reserves",
				},
				[]string{"pool", "denom"},
			),
			LPTokenSupply: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "lp_token_supply",
					Help:      "LP token supply per pool",
				},
				[]string{"pool"},
			),
			PoolTVL: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "pool_tvl_total",
					Help:      "Total Value Locked in pool",
				},
				[]string{"pool"},
			),

			// Pool metrics
			PoolsTotal: promauto.NewGauge(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "pools_total",
					Help:      "Total number of liquidity pools",
				},
			),
			PoolCreationRate: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "pool_creations_total",
					Help:      "Total pools created",
				},
			),
			PoolImbalanceRatio: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "pool_imbalance_ratio",
					Help:      "Pool reserve ratio",
				},
				[]string{"pool"},
			),
			PoolFeeTier: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "pool_fee_tier",
					Help:      "Pool fee tier in basis points",
				},
				[]string{"pool"},
			),

			// HTLC metrics
			HTLCCreated: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "htlc_created_total",
					Help:      "Hash time-locked contracts created",
				},
			),
			HTLCClaimed: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "htlc_claimed_total",
					Help:      "HTLCs successfully claimed",
				},
			),
			HTLCRefunded: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "htlc_refunded_total",
					Help:      "HTLCs refunded after timeout",
				},
			),
			HTLCActive: promauto.NewGauge(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "htlc_active",
					Help:      "Currently active HTLCs",
				},
			),

			// Security metrics
			CircuitBreakerActive: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "circuit_breaker_active",
					Help:      "Circuit breaker activation status",
				},
				[]string{"pool"},
			),
			CircuitBreakerTriggers: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "circuit_breaker_triggers_total",
					Help:      "Circuit breaker trigger events",
				},
				[]string{"pool", "reason"},
			),
			MEVProtections: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "mev_protections_triggered_total",
					Help:      "MEV protection mechanisms triggered",
				},
				[]string{"type"},
			),
			RateLimitExceeds: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "rate_limits_exceeded_total",
					Help:      "Rate limit violations",
				},
				[]string{"operation"},
			),
			SuspiciousActivity: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "suspicious_activity_detected_total",
					Help:      "Suspicious activity detections",
				},
				[]string{"type"},
			),

			// TWAP metrics
			TWAPUpdates: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "twap_updates_total",
					Help:      "TWAP update operations",
				},
			),
			TWAPValue: promauto.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "twap_price",
					Help:      "Time-weighted average price",
				},
				[]string{"pool"},
			),
			PriceManipulationAttempts: promauto.NewCounter(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "price_manipulation_attempts_total",
					Help:      "Detected price manipulation attempts",
				},
			),

			// Cross-chain metrics
			IBCSwapsSent: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "ibc_swaps_sent_total",
					Help:      "Cross-chain swaps sent",
				},
				[]string{"dest_chain"},
			),
			IBCSwapsReceived: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "ibc_swaps_received_total",
					Help:      "Cross-chain swaps received",
				},
				[]string{"source_chain"},
			),
			IBCTimeouts: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "ibc_timeouts_total",
					Help:      "IBC packet timeouts",
				},
				[]string{"chain"},
			),
			CrossChainLatency: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "cross_chain_latency_seconds",
					Help:      "Cross-chain operation latency",
					Buckets:   []float64{5, 10, 30, 60, 120, 300},
				},
				[]string{"chain_pair"},
			),
			IBCPacketsPending: promauto.NewGauge(
				prometheus.GaugeOpts{
					Namespace: "aura",
					Subsystem: "dex",
					Name:      "ibc_packets_pending",
					Help:      "Pending IBC packets",
				},
			),
		}
	})
	return dexMetrics
}

// GetDEXMetrics returns the singleton DEX metrics instance
func GetDEXMetrics() *DEXMetrics {
	if dexMetrics == nil {
		return NewDEXMetrics()
	}
	return dexMetrics
}
