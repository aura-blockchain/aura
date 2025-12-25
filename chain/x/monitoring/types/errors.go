// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Monitoring module error codes
var (
	// Transaction errors (1-9)
	ErrInvalidTransaction = errorsmod.Register(ModuleName, 1, "invalid transaction data")

	// Alert errors (10-19)
	ErrAlertNotFound = errorsmod.Register(ModuleName, 10, "alert not found")

	// Detection errors (20-29)
	ErrAnomalyDetectionFailed = errorsmod.Register(ModuleName, 20, "anomaly detection failed")

	// Validator errors (30-39)
	ErrValidatorNotFound = errorsmod.Register(ModuleName, 30, "validator not found")

	// Threshold errors (40-49)
	ErrInvalidThreshold = errorsmod.Register(ModuleName, 40, "invalid threshold value")

	// Metrics errors (50-59)
	ErrMetricsNotAvailable = errorsmod.Register(ModuleName, 50, "metrics not available")

	// Gas price errors (60-69)
	ErrInvalidGasPrice = errorsmod.Register(ModuleName, 60, "invalid gas price")

	// TVL errors (70-79)
	ErrTVLCalculationFailed = errorsmod.Register(ModuleName, 70, "TVL calculation failed")

	// Security event errors (80-89)
	ErrSecurityEventInvalid = errorsmod.Register(ModuleName, 80, "invalid security event")

	// Log aggregation errors (90-99)
	ErrLogAggregationFailed = errorsmod.Register(ModuleName, 90, "log aggregation failed")

	// ML model errors (100-109)
	ErrMLModelNotTrained = errorsmod.Register(ModuleName, 100, "ML model not trained")

	// SIEM errors (110-119)
	ErrSIEMProcessingFailed = errorsmod.Register(ModuleName, 110, "SIEM processing failed")

	// Explorer integration errors (120-129)
	ErrExplorerIntegrationFailed = errorsmod.Register(ModuleName, 120, "explorer integration failed")
)
