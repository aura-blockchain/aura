// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"errors"
	"fmt"
	"testing"
)

func TestMonitoringErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrInvalidTransaction", ErrInvalidTransaction, "invalid transaction data"},
		{"ErrAlertNotFound", ErrAlertNotFound, "alert not found"},
		{"ErrAnomalyDetectionFailed", ErrAnomalyDetectionFailed, "anomaly detection failed"},
		{"ErrValidatorNotFound", ErrValidatorNotFound, "validator not found"},
		{"ErrInvalidThreshold", ErrInvalidThreshold, "invalid threshold value"},
		{"ErrMetricsNotAvailable", ErrMetricsNotAvailable, "metrics not available"},
		{"ErrInvalidGasPrice", ErrInvalidGasPrice, "invalid gas price"},
		{"ErrTVLCalculationFailed", ErrTVLCalculationFailed, "TVL calculation failed"},
		{"ErrSecurityEventInvalid", ErrSecurityEventInvalid, "invalid security event"},
		{"ErrLogAggregationFailed", ErrLogAggregationFailed, "log aggregation failed"},
		{"ErrMLModelNotTrained", ErrMLModelNotTrained, "ML model not trained"},
		{"ErrSIEMProcessingFailed", ErrSIEMProcessingFailed, "SIEM processing failed"},
		{"ErrExplorerIntegrationFailed", ErrExplorerIntegrationFailed, "explorer integration failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected error message %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestErrorsAreErrors(t *testing.T) {
	var _ error = ErrInvalidTransaction
	var _ error = ErrAlertNotFound
	var _ error = ErrAnomalyDetectionFailed
	var _ error = ErrValidatorNotFound
	var _ error = ErrInvalidThreshold
	var _ error = ErrMetricsNotAvailable
	var _ error = ErrInvalidGasPrice
	var _ error = ErrTVLCalculationFailed
	var _ error = ErrSecurityEventInvalid
	var _ error = ErrLogAggregationFailed
	var _ error = ErrMLModelNotTrained
	var _ error = ErrSIEMProcessingFailed
	var _ error = ErrExplorerIntegrationFailed
}

func TestErrorComparison(t *testing.T) {
	err := ErrAlertNotFound
	if !errors.Is(err, ErrAlertNotFound) {
		t.Error("errors.Is should return true for same error")
	}

	if errors.Is(err, ErrValidatorNotFound) {
		t.Error("errors.Is should return false for different error")
	}
}

func TestErrorWrapping(t *testing.T) {
	baseErr := ErrAnomalyDetectionFailed
	wrappedErr := fmt.Errorf("wrapped: %w", baseErr)

	if !errors.Is(wrappedErr, ErrAnomalyDetectionFailed) {
		t.Error("wrapped error should be detectable with errors.Is")
	}
}
