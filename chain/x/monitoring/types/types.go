// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"time"
)

// TransactionMonitorData represents real-time transaction monitoring data
type TransactionMonitorData struct {
	TxHash           string    `json:"tx_hash"`
	Sender           string    `json:"sender"`
	Receiver         string    `json:"receiver"`
	Amount           uint64    `json:"amount"`
	GasUsed          uint64    `json:"gas_used"`
	GasPrice         uint64    `json:"gas_price"`
	Status           string    `json:"status"` // success, failed, pending
	Timestamp        time.Time `json:"timestamp"`
	BlockHeight      int64     `json:"block_height"`
	Module           string    `json:"module"`
	IsLargeTransfer  bool      `json:"is_large_transfer"`
	AnomalyScore     float64   `json:"anomaly_score"`
}

// Alert represents a security or operational alert
type Alert struct {
	ID               string                 `json:"id"`
	Type             AlertType              `json:"type"`
	Severity         AlertSeverity          `json:"severity"`
	Message          string                 `json:"message"`
	Details          map[string]interface{} `json:"details"`
	Timestamp        time.Time              `json:"timestamp"`
	Acknowledged     bool                   `json:"acknowledged"`
	AcknowledgedBy   string                 `json:"acknowledged_by,omitempty"`
	AcknowledgedAt   *time.Time             `json:"acknowledged_at,omitempty"`
	Resolved         bool                   `json:"resolved"`
	ResolvedAt       *time.Time             `json:"resolved_at,omitempty"`
	NotificationSent bool                   `json:"notification_sent"`
}

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypeSecurityThreat     AlertType = "security_threat"
	AlertTypeAnomaly            AlertType = "anomaly"
	AlertTypeValidatorDown      AlertType = "validator_down"
	AlertTypeNetworkCongestion  AlertType = "network_congestion"
	AlertTypeGasPriceSpike      AlertType = "gas_price_spike"
	AlertTypeLargeTransaction   AlertType = "large_transaction"
	AlertTypeFailedTxPattern    AlertType = "failed_tx_pattern"
	AlertTypeTVLChange          AlertType = "tvl_change"
	AlertTypeSystemError        AlertType = "system_error"
)

// AlertSeverity defines the severity level of an alert
type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "critical"
	SeverityHigh     AlertSeverity = "high"
	SeverityMedium   AlertSeverity = "medium"
	SeverityLow      AlertSeverity = "low"
	SeverityInfo     AlertSeverity = "info"
)

// AnomalyDetection represents the result of anomaly detection
type AnomalyDetection struct {
	ID            string                 `json:"id"`
	Type          AnomalyType            `json:"type"`
	Score         float64                `json:"score"` // 0-1, higher means more anomalous
	Threshold     float64                `json:"threshold"`
	IsAnomaly     bool                   `json:"is_anomaly"`
	Features      map[string]float64     `json:"features"`
	Metadata      map[string]interface{} `json:"metadata"`
	Timestamp     time.Time              `json:"timestamp"`
	ModelVersion  string                 `json:"model_version"`
}

// AnomalyType defines the type of anomaly
type AnomalyType string

const (
	AnomalyTypeTransaction    AnomalyType = "transaction"
	AnomalyTypeGasUsage       AnomalyType = "gas_usage"
	AnomalyTypeNetworkPattern AnomalyType = "network_pattern"
	AnomalyTypeValidatorBehavior AnomalyType = "validator_behavior"
)

// ValidatorUptime represents validator uptime tracking
type ValidatorUptime struct {
	ValidatorAddress  string    `json:"validator_address"`
	Moniker           string    `json:"moniker"`
	TotalBlocks       int64     `json:"total_blocks"`
	SignedBlocks      int64     `json:"signed_blocks"`
	MissedBlocks      int64     `json:"missed_blocks"`
	UptimeBasisPoints uint64    `json:"uptime_basis_points"` // 10000 = 100%, deterministic integer arithmetic
	LastSeen          time.Time `json:"last_seen"`
	Status            string    `json:"status"` // active, jailed, unbonding
	ConsecutiveMisses int64     `json:"consecutive_misses"`
	Jailed            bool      `json:"jailed"`
}

// UptimePercentage returns the uptime as a float64 percentage for display/metrics only.
// This should only be used for external APIs and Prometheus metrics, never for consensus state.
func (v *ValidatorUptime) UptimePercentage() float64 {
	return float64(v.UptimeBasisPoints) / 100.0
}

// NetworkHealth represents overall network health metrics
type NetworkHealth struct {
	Timestamp             time.Time `json:"timestamp"`
	BlockHeight           int64     `json:"block_height"`
	BlockTime             float64   `json:"block_time"` // average in seconds
	TPS                   float64   `json:"tps"`        // transactions per second
	ActiveValidators      int       `json:"active_validators"`
	TotalValidators       int       `json:"total_validators"`
	NetworkHashRate       float64   `json:"network_hash_rate"`
	PeerCount             int       `json:"peer_count"`
	MempoolSize           int       `json:"mempool_size"`
	AverageGasPrice       uint64    `json:"average_gas_price"`
	NetworkCongestion     float64   `json:"network_congestion"` // 0-1
	ConsensusHealth       float64   `json:"consensus_health"`   // 0-1
}

// GasPriceTracking represents gas price monitoring
type GasPriceTracking struct {
	Timestamp       time.Time          `json:"timestamp"`
	CurrentPrice    uint64             `json:"current_price"`
	AveragePrice    uint64             `json:"average_price"`
	MedianPrice     uint64             `json:"median_price"`
	MinPrice        uint64             `json:"min_price"`
	MaxPrice        uint64             `json:"max_price"`
	PriceHistory    []GasPricePoint    `json:"price_history"`
	TrendDirection  string             `json:"trend_direction"` // increasing, decreasing, stable
	VolatilityScore uint64             `json:"volatility_score"` // Coefficient of variation in basis points (10000 = 100%)
}

// GasPricePoint represents a single gas price data point
type GasPricePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Price     uint64    `json:"price"`
}

// TVLMonitoring represents Total Value Locked monitoring
type TVLMonitoring struct {
	Timestamp           time.Time          `json:"timestamp"`
	TotalTVL            uint64             `json:"total_tvl"`
	TVLByModule         map[string]uint64  `json:"tvl_by_module"`
	TVLChange24hBps     int64              `json:"tvl_change_24h_bps"` // basis points (10000 = 100%)
	TVLChange7dBps      int64              `json:"tvl_change_7d_bps"`  // basis points (10000 = 100%)
	TVLHistory          []TVLPoint         `json:"tvl_history"`
	LargestPools        []PoolTVL          `json:"largest_pools"`
}

// TVLPoint represents a single TVL data point
type TVLPoint struct {
	Timestamp time.Time `json:"timestamp"`
	TVL       uint64    `json:"tvl"`
}

// PoolTVL represents TVL for a specific pool
type PoolTVL struct {
	PoolID         string `json:"pool_id"`
	PoolName       string `json:"pool_name"`
	TVL            uint64 `json:"tvl"`
	PercentageBps  int64  `json:"percentage_bps"` // basis points (10000 = 100%)
}

// FailedTransactionPattern represents analysis of failed transaction patterns
type FailedTransactionPattern struct {
	ID              string                 `json:"id"`
	Pattern         string                 `json:"pattern"`
	FailureReason   string                 `json:"failure_reason"`
	Occurrences     int64                  `json:"occurrences"`
	AffectedAddresses []string             `json:"affected_addresses"`
	FirstSeen       time.Time              `json:"first_seen"`
	LastSeen        time.Time              `json:"last_seen"`
	Severity        AlertSeverity          `json:"severity"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// SecurityEvent represents a SIEM security event
type SecurityEvent struct {
	ID              string                 `json:"id"`
	EventType       SecurityEventType      `json:"event_type"`
	Severity        AlertSeverity          `json:"severity"`
	Source          string                 `json:"source"`
	Destination     string                 `json:"destination,omitempty"`
	Description     string                 `json:"description"`
	RawData         map[string]interface{} `json:"raw_data"`
	Timestamp       time.Time              `json:"timestamp"`
	Indicators      []string               `json:"indicators"`
	ThreatLevel     int                    `json:"threat_level"` // 1-10
	Mitigated       bool                   `json:"mitigated"`
	MitigationSteps []string               `json:"mitigation_steps,omitempty"`
}

// SecurityEventType defines the type of security event
type SecurityEventType string

const (
	SecurityEventSuspiciousTransaction SecurityEventType = "suspicious_transaction"
	SecurityEventUnauthorizedAccess    SecurityEventType = "unauthorized_access"
	SecurityEventDDoSAttempt           SecurityEventType = "ddos_attempt"
	SecurityEventContractExploit       SecurityEventType = "contract_exploit"
	SecurityEventValidatorMisbehavior  SecurityEventType = "validator_misbehavior"
	SecurityEventKeyCompromise         SecurityEventType = "key_compromise"
	SecurityEventDataBreach            SecurityEventType = "data_breach"
)

// LogEntry represents a centralized log entry
type LogEntry struct {
	ID        string                 `json:"id"`
	Level     LogLevel               `json:"level"`
	Module    string                 `json:"module"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields"`
	Timestamp time.Time              `json:"timestamp"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
}

// LogLevel defines the log severity level
type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelFatal   LogLevel = "fatal"
)

// ExplorerIntegration represents data for blockchain explorer integration
type ExplorerIntegration struct {
	Endpoint        string            `json:"endpoint"`
	APIKey          string            `json:"api_key,omitempty"`
	SyncStatus      string            `json:"sync_status"`
	LastSyncHeight  int64             `json:"last_sync_height"`
	LastSyncTime    time.Time         `json:"last_sync_time"`
	HealthStatus    string            `json:"health_status"`
	CustomMetadata  map[string]string `json:"custom_metadata,omitempty"`
}
