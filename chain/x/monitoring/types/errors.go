package types

import (
	"fmt"
)

var (
	ErrInvalidTransaction       = fmt.Errorf("invalid transaction data")
	ErrAlertNotFound           = fmt.Errorf("alert not found")
	ErrAnomalyDetectionFailed  = fmt.Errorf("anomaly detection failed")
	ErrValidatorNotFound       = fmt.Errorf("validator not found")
	ErrInvalidThreshold        = fmt.Errorf("invalid threshold value")
	ErrMetricsNotAvailable     = fmt.Errorf("metrics not available")
	ErrInvalidGasPrice         = fmt.Errorf("invalid gas price")
	ErrTVLCalculationFailed    = fmt.Errorf("TVL calculation failed")
	ErrSecurityEventInvalid    = fmt.Errorf("invalid security event")
	ErrLogAggregationFailed    = fmt.Errorf("log aggregation failed")
	ErrMLModelNotTrained       = fmt.Errorf("ML model not trained")
	ErrSIEMProcessingFailed    = fmt.Errorf("SIEM processing failed")
	ErrExplorerIntegrationFailed = fmt.Errorf("explorer integration failed")
)
