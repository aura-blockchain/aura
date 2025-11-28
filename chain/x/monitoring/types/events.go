package types

import "fmt"

// Event types for the monitoring module
const (
	EventTypeAlertTriggered  = "alert_triggered"
	EventTypeAlertResolved   = "alert_resolved"
	EventTypeAnomalyDetected = "anomaly_detected"
	EventTypeMetricRecorded  = "metric_recorded"
	EventTypeThresholdBreach = "threshold_breach"
	EventTypeParamsUpdated   = "params_updated"
)

// Event attribute keys
const (
	AttributeKeyAlertID     = "alert_id"
	AttributeKeyMetricName  = "metric_name"
	AttributeKeySeverity    = "severity"
	AttributeKeyValue       = "value"
	AttributeKeyThreshold   = "threshold"
	AttributeKeyBlockHeight = "block_height"
	AttributeKeyBlockTime   = "block_time"
)

func formatInt64(i int64) string   { return fmt.Sprintf("%d", i) }
func formatFloat64(f float64) string { return fmt.Sprintf("%.6f", f) }
