// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

// Enum constants for easier use
const (
	IRCompletionStatusUnspecified = IRCompletionStatus_IR_COMPLETION_STATUS_UNSPECIFIED
	IRCompletionStatusPending     = IRCompletionStatus_IR_COMPLETION_STATUS_PENDING
	IRCompletionStatusVerified    = IRCompletionStatus_IR_COMPLETION_STATUS_VERIFIED
	IRCompletionStatusRejected    = IRCompletionStatus_IR_COMPLETION_STATUS_REJECTED
	IRCompletionStatusAppealed    = IRCompletionStatus_IR_COMPLETION_STATUS_APPEALED

	VerificationStatusUnspecified = VerificationStatus_VERIFICATION_STATUS_UNSPECIFIED
	VerificationStatusUnverified  = VerificationStatus_VERIFICATION_STATUS_UNVERIFIED
	VerificationStatusVerified    = VerificationStatus_VERIFICATION_STATUS_VERIFIED
	VerificationStatusSuspended   = VerificationStatus_VERIFICATION_STATUS_SUSPENDED
	VerificationStatusRevoked     = VerificationStatus_VERIFICATION_STATUS_REVOKED

	SlashReasonUnspecified         = SlashReason_SLASH_REASON_UNSPECIFIED
	SlashReasonFraudDetected       = SlashReason_SLASH_REASON_FRAUD_DETECTED
	SlashReasonFalseAttestation    = SlashReason_SLASH_REASON_FALSE_ATTESTATION
	SlashReasonCollusion           = SlashReason_SLASH_REASON_COLLUSION
	SlashReasonDuplicateCompletion = SlashReason_SLASH_REASON_DUPLICATE_COMPLETION
)
