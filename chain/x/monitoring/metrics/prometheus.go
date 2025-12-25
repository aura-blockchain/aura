// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// metricsOnce ensures metrics are only registered once
	metricsOnce sync.Once
	// singletonMetrics holds the singleton instance
	singletonMetrics *MonitoringMetrics
)

// MonitoringMetrics holds all Prometheus metrics for the monitoring module
type MonitoringMetrics struct {
	// Transaction metrics
	TotalTransactions    *prometheus.CounterVec
	TransactionGasUsed   *prometheus.HistogramVec
	TransactionDuration  prometheus.Histogram
	LargeTransactions    prometheus.Counter
	FailedTransactions   *prometheus.CounterVec

	// Alert metrics
	TotalAlerts          *prometheus.CounterVec
	ActiveAlerts         *prometheus.GaugeVec
	AlertResolutionTime  prometheus.Histogram

	// Anomaly detection metrics
	AnomalyDetections    *prometheus.CounterVec
	AnomalyScore         prometheus.Histogram
	MLModelVersion       prometheus.Gauge
	MLModelAccuracy      prometheus.Gauge

	// Validator metrics
	ValidatorUptime      *prometheus.GaugeVec
	ValidatorMissedBlocks *prometheus.CounterVec
	JailedValidators     prometheus.Gauge
	ActiveValidators     prometheus.Gauge

	// Network health metrics
	BlockTime            prometheus.Gauge
	TransactionsPerSecond prometheus.Gauge
	MempoolSize          prometheus.Gauge
	PeerCount            prometheus.Gauge
	NetworkCongestion    prometheus.Gauge
	ConsensusHealth      prometheus.Gauge

	// Gas price metrics
	CurrentGasPrice      prometheus.Gauge
	AverageGasPrice      prometheus.Gauge
	GasPriceVolatility   prometheus.Gauge
	GasPriceSpikes       prometheus.Counter

	// TVL metrics
	TotalTVL             prometheus.Gauge
	TVLByModule          *prometheus.GaugeVec
	TVLChange24h         prometheus.Gauge
	TVLChange7d          prometheus.Gauge

	// Security metrics
	SecurityEvents       *prometheus.CounterVec
	ThreatLevel          *prometheus.GaugeVec
	MitigatedThreats     prometheus.Counter

	// Log aggregation metrics
	LogEntriesTotal      *prometheus.CounterVec
	LogProcessingErrors  prometheus.Counter

	// Explorer integration metrics
	ExplorerSyncHeight   prometheus.Gauge
	ExplorerSyncLag      prometheus.Gauge
	ExplorerAPIErrors    prometheus.Counter

	// SOC metrics
	SOCActiveOperators   prometheus.Gauge
	SOCIncidents         *prometheus.CounterVec
	SOCResponseTime      prometheus.Histogram
}

// NewMonitoringMetrics creates and registers all monitoring metrics.
// Uses a singleton pattern to prevent duplicate metric registration when
// the app is recreated (e.g., during node start when createAppWithDB is called).
func NewMonitoringMetrics() *MonitoringMetrics {
	metricsOnce.Do(func() {
		singletonMetrics = createMonitoringMetrics()
	})
	return singletonMetrics
}

// createMonitoringMetrics creates the actual metrics (called only once)
func createMonitoringMetrics() *MonitoringMetrics {
	return &MonitoringMetrics{
		// Transaction metrics
		TotalTransactions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "total_transactions",
				Help:      "Total number of transactions monitored",
			},
			[]string{"status", "module"},
		),
		TransactionGasUsed: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "transaction_gas_used",
				Help:      "Gas used by transactions",
				Buckets:   prometheus.ExponentialBuckets(1000, 2, 15),
			},
			[]string{"module"},
		),
		TransactionDuration: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "transaction_duration_seconds",
				Help:      "Transaction processing duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
		),
		LargeTransactions: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "large_transactions_total",
				Help:      "Total number of large transactions detected",
			},
		),
		FailedTransactions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "failed_transactions_total",
				Help:      "Total number of failed transactions",
			},
			[]string{"reason", "module"},
		),

		// Alert metrics
		TotalAlerts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "alerts_total",
				Help:      "Total number of alerts generated",
			},
			[]string{"type", "severity"},
		),
		ActiveAlerts: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "active_alerts",
				Help:      "Number of active unresolved alerts",
			},
			[]string{"type", "severity"},
		),
		AlertResolutionTime: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "alert_resolution_time_seconds",
				Help:      "Time taken to resolve alerts in seconds",
				Buckets:   prometheus.ExponentialBuckets(60, 2, 10),
			},
		),

		// Anomaly detection metrics
		AnomalyDetections: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "anomaly_detections_total",
				Help:      "Total number of anomalies detected",
			},
			[]string{"type"},
		),
		AnomalyScore: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "anomaly_score",
				Help:      "Distribution of anomaly scores",
				Buckets:   prometheus.LinearBuckets(0, 0.1, 11),
			},
		),
		MLModelVersion: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "ml_model_version",
				Help:      "Current ML model version",
			},
		),
		MLModelAccuracy: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "ml_model_accuracy",
				Help:      "Current ML model accuracy",
			},
		),

		// Validator metrics
		ValidatorUptime: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "validator_uptime_percentage",
				Help:      "Validator uptime percentage",
			},
			[]string{"validator_address", "moniker"},
		),
		ValidatorMissedBlocks: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "validator_missed_blocks_total",
				Help:      "Total number of blocks missed by validator",
			},
			[]string{"validator_address", "moniker"},
		),
		JailedValidators: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "jailed_validators",
				Help:      "Number of jailed validators",
			},
		),
		ActiveValidators: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "active_validators",
				Help:      "Number of active validators",
			},
		),

		// Network health metrics
		BlockTime: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "block_time_seconds",
				Help:      "Average block time in seconds",
			},
		),
		TransactionsPerSecond: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "transactions_per_second",
				Help:      "Network transactions per second",
			},
		),
		MempoolSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "mempool_size",
				Help:      "Number of transactions in mempool",
			},
		),
		PeerCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "peer_count",
				Help:      "Number of connected peers",
			},
		),
		NetworkCongestion: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "network_congestion",
				Help:      "Network congestion level (0-1)",
			},
		),
		ConsensusHealth: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "consensus_health",
				Help:      "Consensus health score (0-1)",
			},
		),

		// Gas price metrics
		CurrentGasPrice: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "current_gas_price",
				Help:      "Current gas price",
			},
		),
		AverageGasPrice: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "average_gas_price",
				Help:      "Average gas price",
			},
		),
		GasPriceVolatility: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "gas_price_volatility",
				Help:      "Gas price volatility score",
			},
		),
		GasPriceSpikes: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "gas_price_spikes_total",
				Help:      "Total number of gas price spikes detected",
			},
		),

		// TVL metrics
		TotalTVL: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "total_tvl",
				Help:      "Total Value Locked across all modules",
			},
		),
		TVLByModule: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "tvl_by_module",
				Help:      "Total Value Locked by module",
			},
			[]string{"module"},
		),
		TVLChange24h: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "tvl_change_24h_percentage",
				Help:      "TVL change over 24 hours in percentage",
			},
		),
		TVLChange7d: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "tvl_change_7d_percentage",
				Help:      "TVL change over 7 days in percentage",
			},
		),

		// Security metrics
		SecurityEvents: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "security_events_total",
				Help:      "Total number of security events",
			},
			[]string{"event_type", "severity"},
		),
		ThreatLevel: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "threat_level",
				Help:      "Threat level for security events",
			},
			[]string{"event_type"},
		),
		MitigatedThreats: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "mitigated_threats_total",
				Help:      "Total number of mitigated security threats",
			},
		),

		// Log aggregation metrics
		LogEntriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "log_entries_total",
				Help:      "Total number of log entries",
			},
			[]string{"level", "module"},
		),
		LogProcessingErrors: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "log_processing_errors_total",
				Help:      "Total number of log processing errors",
			},
		),

		// Explorer integration metrics
		ExplorerSyncHeight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "explorer_sync_height",
				Help:      "Explorer sync height",
			},
		),
		ExplorerSyncLag: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "explorer_sync_lag_blocks",
				Help:      "Explorer sync lag in blocks",
			},
		),
		ExplorerAPIErrors: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "explorer_api_errors_total",
				Help:      "Total number of explorer API errors",
			},
		),

		// SOC metrics
		SOCActiveOperators: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "soc_active_operators",
				Help:      "Number of active SOC operators",
			},
		),
		SOCIncidents: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "soc_incidents_total",
				Help:      "Total number of SOC incidents",
			},
			[]string{"severity"},
		),
		SOCResponseTime: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Namespace: "aura",
				Subsystem: "monitoring",
				Name:      "soc_response_time_seconds",
				Help:      "SOC incident response time in seconds",
				Buckets:   prometheus.ExponentialBuckets(60, 2, 8),
			},
		),
	}
}
