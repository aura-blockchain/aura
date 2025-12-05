package types

import (
	"time"
)

// Params defines the parameters for the monitoring module
type Params struct {
	// Transaction Monitoring
	EnableTransactionMonitoring bool   `json:"enable_transaction_monitoring"`
	LargeTransactionThreshold   uint64 `json:"large_transaction_threshold"`

	// Alert System
	EnableAlerts              bool          `json:"enable_alerts"`
	AlertRetentionPeriod      time.Duration `json:"alert_retention_period"`
	CriticalAlertCooldown     time.Duration `json:"critical_alert_cooldown"`

	// Anomaly Detection
	EnableAnomalyDetection    bool    `json:"enable_anomaly_detection"`
	AnomalyThreshold          float64 `json:"anomaly_threshold"`
	MLModelUpdateInterval     time.Duration `json:"ml_model_update_interval"`

	// Prometheus Integration
	EnablePrometheusMetrics   bool   `json:"enable_prometheus_metrics"`
	PrometheusPort            int    `json:"prometheus_port"`
	MetricsRetentionPeriod    time.Duration `json:"metrics_retention_period"`

	// Validator Monitoring
	EnableValidatorMonitoring bool    `json:"enable_validator_monitoring"`
	ValidatorUptimeWindow     int64   `json:"validator_uptime_window"`
	MaxConsecutiveMisses      int64   `json:"max_consecutive_misses"`

	// Network Health
	EnableNetworkHealthMonitoring bool    `json:"enable_network_health_monitoring"`
	NetworkHealthCheckInterval    time.Duration `json:"network_health_check_interval"`
	CongestionThreshold           float64 `json:"congestion_threshold"`

	// Gas Price Tracking
	EnableGasPriceTracking    bool    `json:"enable_gas_price_tracking"`
	GasPriceCheckInterval     time.Duration `json:"gas_price_check_interval"`
	GasPriceSpikeThreshold    float64 `json:"gas_price_spike_threshold"`
	GasPriceHistorySize       int     `json:"gas_price_history_size"`

	// TVL Monitoring
	EnableTVLMonitoring       bool    `json:"enable_tvl_monitoring"`
	TVLCheckInterval          time.Duration `json:"tvl_check_interval"`
	TVLChangeAlertThreshold   float64 `json:"tvl_change_alert_threshold"`
	TVLHistorySize            int     `json:"tvl_history_size"`

	// Failed Transaction Analysis
	EnableFailedTxAnalysis    bool    `json:"enable_failed_tx_analysis"`
	FailedTxPatternThreshold  int64   `json:"failed_tx_pattern_threshold"`
	FailedTxAnalysisWindow    time.Duration `json:"failed_tx_analysis_window"`

	// SIEM
	EnableSIEM                bool    `json:"enable_siem"`
	SIEMRetentionPeriod       time.Duration `json:"siem_retention_period"`
	ThreatLevelThreshold      int     `json:"threat_level_threshold"`

	// Log Aggregation
	EnableLogAggregation      bool    `json:"enable_log_aggregation"`
	LogRetentionPeriod        time.Duration `json:"log_retention_period"`
	MaxLogEntriesPerModule    int64   `json:"max_log_entries_per_module"`

	// Explorer Integration
	EnableExplorerIntegration bool   `json:"enable_explorer_integration"`
	ExplorerEndpoint          string `json:"explorer_endpoint"`
	ExplorerSyncInterval      time.Duration `json:"explorer_sync_interval"`

	// SOC Framework
	EnableSOC                 bool   `json:"enable_soc"`
	SOCRotationInterval       time.Duration `json:"soc_rotation_interval"`
	RequireSOCApproval        bool   `json:"require_soc_approval"`
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		// Transaction Monitoring
		EnableTransactionMonitoring: true,
		LargeTransactionThreshold:   1000000, // 1M tokens

		// Alert System
		EnableAlerts:              true,
		AlertRetentionPeriod:      30 * 24 * time.Hour, // 30 days
		CriticalAlertCooldown:     5 * time.Minute,

		// Anomaly Detection
		EnableAnomalyDetection:    true,
		AnomalyThreshold:          0.75,
		MLModelUpdateInterval:     24 * time.Hour,

		// Prometheus Integration
		EnablePrometheusMetrics:   true,
		PrometheusPort:            9090,
		MetricsRetentionPeriod:    7 * 24 * time.Hour, // 7 days

		// Validator Monitoring
		EnableValidatorMonitoring: true,
		ValidatorUptimeWindow:     10000, // blocks
		MaxConsecutiveMisses:      100,

		// Network Health
		EnableNetworkHealthMonitoring: true,
		NetworkHealthCheckInterval:    30 * time.Second,
		CongestionThreshold:           0.8,

		// Gas Price Tracking
		EnableGasPriceTracking:    true,
		GasPriceCheckInterval:     1 * time.Minute,
		GasPriceSpikeThreshold:    2.0, // 2x average
		GasPriceHistorySize:       1440, // 24 hours at 1-minute intervals

		// TVL Monitoring
		EnableTVLMonitoring:       true,
		TVLCheckInterval:          5 * time.Minute,
		TVLChangeAlertThreshold:   0.2, // 20% change
		TVLHistorySize:            288, // 24 hours at 5-minute intervals

		// Failed Transaction Analysis
		EnableFailedTxAnalysis:    true,
		FailedTxPatternThreshold:  10,
		FailedTxAnalysisWindow:    1 * time.Hour,

		// SIEM
		EnableSIEM:                true,
		SIEMRetentionPeriod:       90 * 24 * time.Hour, // 90 days
		ThreatLevelThreshold:      7,

		// Log Aggregation
		EnableLogAggregation:      true,
		LogRetentionPeriod:        14 * 24 * time.Hour, // 14 days
		MaxLogEntriesPerModule:    100000,

		// Explorer Integration
		EnableExplorerIntegration: true,
		ExplorerEndpoint:          "http://localhost:8080",
		ExplorerSyncInterval:      10 * time.Second,

		// SOC Framework
		EnableSOC:                 true,
		SOCRotationInterval:       8 * time.Hour,
		RequireSOCApproval:        false,
	}
}

// ValidateParams validates the parameters
func ValidateParams(p Params) error {
	if p.LargeTransactionThreshold == 0 {
		return ErrInvalidThreshold
	}

	if p.AnomalyThreshold < 0 || p.AnomalyThreshold > 1 {
		return ErrInvalidThreshold
	}

	if p.PrometheusPort < 1024 || p.PrometheusPort > 65535 {
		return ErrInvalidThreshold
	}

	if p.CongestionThreshold < 0 || p.CongestionThreshold > 1 {
		return ErrInvalidThreshold
	}

	if p.GasPriceSpikeThreshold <= 1 {
		return ErrInvalidThreshold
	}

	if p.TVLChangeAlertThreshold < 0 || p.TVLChangeAlertThreshold > 1 {
		return ErrInvalidThreshold
	}

	if p.ThreatLevelThreshold < 1 || p.ThreatLevelThreshold > 10 {
		return ErrInvalidThreshold
	}

	// Validate retention periods are non-negative
	if p.AlertRetentionPeriod < 0 {
		return ErrInvalidThreshold
	}
	if p.MetricsRetentionPeriod < 0 {
		return ErrInvalidThreshold
	}
	if p.SIEMRetentionPeriod < 0 {
		return ErrInvalidThreshold
	}
	if p.LogRetentionPeriod < 0 {
		return ErrInvalidThreshold
	}

	return nil
}
