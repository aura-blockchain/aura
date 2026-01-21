// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"cosmossdk.io/errors"
)

// DONTCOVER

// Sentinel errors for the compliance module
var (
	ErrKYCNotFound              = errors.Register(ModuleName, 2, "KYC record not found")
	ErrKYCExpired               = errors.Register(ModuleName, 3, "KYC record has expired")
	ErrInsufficientKYCLevel     = errors.Register(ModuleName, 4, "insufficient KYC level for operation")
	ErrKYCNotRequired           = errors.Register(ModuleName, 5, "KYC not required")
	ErrInvalidKYCRecord         = errors.Register(ModuleName, 6, "invalid KYC record")
	ErrUnauthorizedProvider     = errors.Register(ModuleName, 7, "unauthorized KYC provider")
	ErrJurisdictionBlocked      = errors.Register(ModuleName, 8, "jurisdiction is blocked due to sanctions")
	ErrSanctionsMatch           = errors.Register(ModuleName, 9, "sanctions screening match found")
	ErrProcessingRestricted     = errors.Register(ModuleName, 10, "data processing restricted (consent withdrawn)")
	ErrGDPRViolation            = errors.Register(ModuleName, 11, "GDPR compliance violation")
	ErrInvalidPIICommitment     = errors.Register(ModuleName, 12, "invalid PII commitment")
	ErrAMLProfileNotFound       = errors.Register(ModuleName, 13, "AML profile not found")
	ErrHighRiskTransaction      = errors.Register(ModuleName, 14, "transaction flagged as high risk")
	ErrTransactionLimitExceeded = errors.Register(ModuleName, 15, "transaction limit exceeded")
	ErrInvalidTaxReport         = errors.Register(ModuleName, 16, "invalid tax report")
	ErrIBCNotEnabled            = errors.Register(ModuleName, 99, "IBC not enabled for compliance module - cross-chain compliance features will be available in v2.0")
)
