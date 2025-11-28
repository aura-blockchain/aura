package types

import (
	"testing"
	"time"
)

func TestAlertTypeConstants(t *testing.T) {
	alertTypes := []AlertType{
		AlertTypeSecurityThreat,
		AlertTypeAnomaly,
		AlertTypeValidatorDown,
		AlertTypeNetworkCongestion,
		AlertTypeGasPriceSpike,
		AlertTypeLargeTransaction,
		AlertTypeFailedTxPattern,
		AlertTypeTVLChange,
		AlertTypeSystemError,
	}

	seen := make(map[AlertType]bool)
	for _, alertType := range alertTypes {
		if seen[alertType] {
			t.Errorf("duplicate AlertType value: %v", alertType)
		}
		seen[alertType] = true
	}
}

func TestAlertSeverityConstants(t *testing.T) {
	severities := []AlertSeverity{
		SeverityCritical,
		SeverityHigh,
		SeverityMedium,
		SeverityLow,
		SeverityInfo,
	}

	seen := make(map[AlertSeverity]bool)
	for _, severity := range severities {
		if seen[severity] {
			t.Errorf("duplicate AlertSeverity value: %v", severity)
		}
		seen[severity] = true
	}
}

func TestAnomalyTypeConstants(t *testing.T) {
	anomalyTypes := []AnomalyType{
		AnomalyTypeTransaction,
		AnomalyTypeGasUsage,
		AnomalyTypeNetworkPattern,
		AnomalyTypeValidatorBehavior,
	}

	seen := make(map[AnomalyType]bool)
	for _, anomalyType := range anomalyTypes {
		if seen[anomalyType] {
			t.Errorf("duplicate AnomalyType value: %v", anomalyType)
		}
		seen[anomalyType] = true
	}
}

func TestSecurityEventTypeConstants(t *testing.T) {
	eventTypes := []SecurityEventType{
		SecurityEventSuspiciousTransaction,
		SecurityEventUnauthorizedAccess,
		SecurityEventDDoSAttempt,
		SecurityEventContractExploit,
		SecurityEventValidatorMisbehavior,
		SecurityEventKeyCompromise,
		SecurityEventDataBreach,
	}

	seen := make(map[SecurityEventType]bool)
	for _, eventType := range eventTypes {
		if seen[eventType] {
			t.Errorf("duplicate SecurityEventType value: %v", eventType)
		}
		seen[eventType] = true
	}
}

func TestLogLevelConstants(t *testing.T) {
	logLevels := []LogLevel{
		LogLevelDebug,
		LogLevelInfo,
		LogLevelWarning,
		LogLevelError,
		LogLevelFatal,
	}

	seen := make(map[LogLevel]bool)
	for _, logLevel := range logLevels {
		if seen[logLevel] {
			t.Errorf("duplicate LogLevel value: %v", logLevel)
		}
		seen[logLevel] = true
	}
}

func TestTransactionMonitorDataStruct(t *testing.T) {
	data := TransactionMonitorData{
		TxHash:      "0x123",
		Sender:      "aura1sender",
		Receiver:    "aura1receiver",
		Amount:      1000,
		GasUsed:     21000,
		GasPrice:    100,
		Status:      "success",
		Timestamp:   time.Now(),
		BlockHeight: 100,
		Module:      "bank",
	}

	if data.TxHash != "0x123" {
		t.Error("TxHash not set correctly")
	}

	if data.Sender != "aura1sender" {
		t.Error("Sender not set correctly")
	}
}

func TestAlertStruct(t *testing.T) {
	now := time.Now()
	alert := Alert{
		ID:           "alert1",
		Type:         AlertTypeSecurityThreat,
		Severity:     SeverityCritical,
		Message:      "Test alert",
		Details:      make(map[string]interface{}),
		Timestamp:    now,
		Acknowledged: false,
	}

	if alert.ID != "alert1" {
		t.Error("ID not set correctly")
	}

	if alert.Type != AlertTypeSecurityThreat {
		t.Error("Type not set correctly")
	}

	if alert.Severity != SeverityCritical {
		t.Error("Severity not set correctly")
	}

	if alert.Acknowledged {
		t.Error("Acknowledged should be false")
	}
}

func TestAnomalyDetectionStruct(t *testing.T) {
	detection := AnomalyDetection{
		ID:        "anomaly1",
		Type:      AnomalyTypeTransaction,
		Score:     0.85,
		Threshold: 0.75,
		IsAnomaly: true,
		Features:  make(map[string]float64),
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	if detection.Score < detection.Threshold {
		t.Error("Score should be above threshold for anomaly")
	}

	if !detection.IsAnomaly {
		t.Error("IsAnomaly should be true")
	}
}

func TestValidatorUptimeStruct(t *testing.T) {
	uptime := ValidatorUptime{
		ValidatorAddress:  "auravaloper1test",
		Moniker:           "Test Validator",
		TotalBlocks:       1000,
		SignedBlocks:      950,
		MissedBlocks:      50,
		UptimePercentage:  95.0,
		LastSeen:          time.Now(),
		Status:            "active",
		ConsecutiveMisses: 2,
		Jailed:            false,
	}

	expectedUptime := float64(uptime.SignedBlocks) / float64(uptime.TotalBlocks) * 100
	if uptime.UptimePercentage != 95.0 {
		t.Errorf("expected uptime to be 95.0, got %f (calculated: %f)", uptime.UptimePercentage, expectedUptime)
	}

	if uptime.Jailed {
		t.Error("Jailed should be false")
	}
}

func TestNetworkHealthStruct(t *testing.T) {
	health := NetworkHealth{
		Timestamp:         time.Now(),
		BlockHeight:       1000,
		BlockTime:         6.0,
		TPS:               100.0,
		ActiveValidators:  50,
		TotalValidators:   100,
		PeerCount:         25,
		MempoolSize:       500,
		NetworkCongestion: 0.3,
		ConsensusHealth:   0.95,
	}

	if health.ActiveValidators > health.TotalValidators {
		t.Error("ActiveValidators should not exceed TotalValidators")
	}

	if health.NetworkCongestion < 0 || health.NetworkCongestion > 1 {
		t.Error("NetworkCongestion should be between 0 and 1")
	}

	if health.ConsensusHealth < 0 || health.ConsensusHealth > 1 {
		t.Error("ConsensusHealth should be between 0 and 1")
	}
}

func TestSecurityEventStruct(t *testing.T) {
	event := SecurityEvent{
		ID:          "event1",
		EventType:   SecurityEventSuspiciousTransaction,
		Severity:    SeverityHigh,
		Source:      "detector1",
		Description: "Suspicious activity detected",
		RawData:     make(map[string]interface{}),
		Timestamp:   time.Now(),
		Indicators:  []string{"indicator1", "indicator2"},
		ThreatLevel: 8,
		Mitigated:   false,
	}

	if event.ThreatLevel < 1 || event.ThreatLevel > 10 {
		t.Error("ThreatLevel should be between 1 and 10")
	}

	if event.Mitigated {
		t.Error("Mitigated should be false")
	}

	if len(event.Indicators) != 2 {
		t.Error("Should have 2 indicators")
	}
}
