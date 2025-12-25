// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

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
