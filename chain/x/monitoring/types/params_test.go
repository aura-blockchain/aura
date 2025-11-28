package types

import (
	"testing"
	"time"
)

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	// Transaction Monitoring
	if !params.EnableTransactionMonitoring {
		t.Error("expected EnableTransactionMonitoring to be true")
	}
	if params.LargeTransactionThreshold != 1000000 {
		t.Errorf("expected LargeTransactionThreshold to be 1000000, got %d", params.LargeTransactionThreshold)
	}

	// Alert System
	if !params.EnableAlerts {
		t.Error("expected EnableAlerts to be true")
	}
	if params.AlertRetentionPeriod != 30*24*time.Hour {
		t.Errorf("expected AlertRetentionPeriod to be 30 days, got %v", params.AlertRetentionPeriod)
	}

	// Anomaly Detection
	if !params.EnableAnomalyDetection {
		t.Error("expected EnableAnomalyDetection to be true")
	}
	if params.AnomalyThreshold != 0.75 {
		t.Errorf("expected AnomalyThreshold to be 0.75, got %f", params.AnomalyThreshold)
	}

	// Prometheus Integration
	if !params.EnablePrometheusMetrics {
		t.Error("expected EnablePrometheusMetrics to be true")
	}
	if params.PrometheusPort != 9090 {
		t.Errorf("expected PrometheusPort to be 9090, got %d", params.PrometheusPort)
	}

	// Validator Monitoring
	if !params.EnableValidatorMonitoring {
		t.Error("expected EnableValidatorMonitoring to be true")
	}

	// Network Health
	if !params.EnableNetworkHealthMonitoring {
		t.Error("expected EnableNetworkHealthMonitoring to be true")
	}

	// Gas Price Tracking
	if !params.EnableGasPriceTracking {
		t.Error("expected EnableGasPriceTracking to be true")
	}

	// TVL Monitoring
	if !params.EnableTVLMonitoring {
		t.Error("expected EnableTVLMonitoring to be true")
	}

	// SIEM
	if !params.EnableSIEM {
		t.Error("expected EnableSIEM to be true")
	}

	// Log Aggregation
	if !params.EnableLogAggregation {
		t.Error("expected EnableLogAggregation to be true")
	}

	// Explorer Integration
	if !params.EnableExplorerIntegration {
		t.Error("expected EnableExplorerIntegration to be true")
	}

	// SOC Framework
	if !params.EnableSOC {
		t.Error("expected EnableSOC to be true")
	}
}

func TestValidateParams_Valid(t *testing.T) {
	params := DefaultParams()

	err := ValidateParams(params)
	if err != nil {
		t.Errorf("ValidateParams should not return error for default params: %v", err)
	}
}

func TestValidateParams_ZeroThreshold(t *testing.T) {
	params := DefaultParams()
	params.LargeTransactionThreshold = 0

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for zero threshold")
	}
}

func TestValidateParams_InvalidAnomalyThreshold(t *testing.T) {
	params := DefaultParams()
	params.AnomalyThreshold = 1.5

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for anomaly threshold > 1")
	}

	params.AnomalyThreshold = -0.1
	err = ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for negative anomaly threshold")
	}
}

func TestValidateParams_InvalidPrometheusPort(t *testing.T) {
	params := DefaultParams()
	params.PrometheusPort = 500 // < 1024

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for port < 1024")
	}

	params.PrometheusPort = 70000 // > 65535
	err = ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for port > 65535")
	}
}

func TestValidateParams_InvalidCongestionThreshold(t *testing.T) {
	params := DefaultParams()
	params.CongestionThreshold = 1.5

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for congestion threshold > 1")
	}

	params.CongestionThreshold = -0.1
	err = ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for negative congestion threshold")
	}
}

func TestValidateParams_InvalidGasPriceSpikeThreshold(t *testing.T) {
	params := DefaultParams()
	params.GasPriceSpikeThreshold = 0.5 // <= 1

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for gas price spike threshold <= 1")
	}
}

func TestValidateParams_InvalidTVLChangeThreshold(t *testing.T) {
	params := DefaultParams()
	params.TVLChangeAlertThreshold = 1.5

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for TVL change threshold > 1")
	}

	params.TVLChangeAlertThreshold = -0.1
	err = ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for negative TVL change threshold")
	}
}

func TestValidateParams_InvalidThreatLevel(t *testing.T) {
	params := DefaultParams()
	params.ThreatLevelThreshold = 0 // < 1

	err := ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for threat level < 1")
	}

	params.ThreatLevelThreshold = 11 // > 10
	err = ValidateParams(params)
	if err == nil {
		t.Error("ValidateParams should return error for threat level > 10")
	}
}
