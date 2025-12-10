// +build integration_examples

package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// This file contains examples of how to integrate the DataProtectionService
// with existing keeper methods. These are reference implementations showing
// best practices for protecting sensitive compliance data.
//
// NOTE: These examples show the pattern. Actual integration would require:
// 1. Updating protobuf schemas with commitment fields
// 2. Modifying keeper methods to use commitments
// 3. Implementing off-chain provider integration
// 4. Migration of existing data
//
// This file is excluded from normal builds via build tag to avoid "declared and not used"
// errors. To view/build these examples, use: go build -tags integration_examples

// Example 1: Enhanced KYC Record Creation with PII Commitment
// ===========================================================

// SubmitKYCWithCommitment demonstrates secure KYC record creation
//
// This is the RECOMMENDED pattern for KYC record storage:
// 1. Provider collects PII off-chain
// 2. Provider generates SHA-256 commitment
// 3. Only commitment is stored on-chain
// 4. Original PII stored in provider's secure off-chain database
func (k *Keeper) SubmitKYCWithCommitmentExample(
	ctx sdk.Context,
	address string,
	kycLevel types.KYCLevel,
	provider string,
	offChainPII *PIIData, // Collected off-chain, passed for commitment only
) error {
	// Initialize data protection service
	dataProtection := NewDataProtectionService()

	// Generate cryptographic commitment for PII
	// This is the ONLY thing stored on-chain
	commitment, err := dataProtection.GeneratePIICommitment(offChainPII)
	if err != nil {
		return fmt.Errorf("failed to generate PII commitment: %w", err)
	}

	// Create KYC record with commitment (NO plaintext PII)
	record := &types.KYCRecord{
		Address:              address,
		KycLevel:             kycLevel,
		Provider:             provider,
		PiiCommitment:        commitment, // Only 32-byte hash stored
		VerifiedAt:           timestamppb.Now(),
		ExpiresAt:            timestamppb.New(ctx.BlockTime().Add(365 * 24 * time.Hour)),
		EnhancedDueDiligence: kycLevel >= types.KYCLevel_KYC_LEVEL_ADVANCED,
	}

	// Store record on-chain (commitment only)
	if err := k.SetKYCRecord(ctx, record); err != nil {
		return err
	}

	// NOTE: Provider must separately store offChainPII in their secure
	// off-chain database with the commitment as the key/reference
	// Example (pseudocode):
	//   providerDB.StorePII(address, offChainPII, commitment)

	// Emit event (with redacted data for privacy)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"kyc_submitted",
			sdk.NewAttribute("address", address),
			sdk.NewAttribute("kyc_level", kycLevel.String()),
			sdk.NewAttribute("provider", provider),
			sdk.NewAttribute("commitment", dataProtection.CommitmentToHex(commitment)),
			// Never emit PII in events!
		),
	)

	return nil
}

// VerifyKYCCommitment verifies that provided PII matches the on-chain commitment
//
// Use case: User or auditor wants to verify data integrity
func (k *Keeper) VerifyKYCCommitmentExample(
	ctx sdk.Context,
	address string,
	claimedPII *PIIData, // Retrieved from off-chain provider
) (bool, error) {
	// Retrieve KYC record with commitment
	record, err := k.GetKYCRecord(ctx, address)
	if err != nil {
		return false, fmt.Errorf("KYC record not found: %w", err)
	}

	// Verify commitment
	dataProtection := NewDataProtectionService()
	valid, err := dataProtection.VerifyPIICommitment(claimedPII, record.PiiCommitment)
	if err != nil {
		return false, fmt.Errorf("commitment verification failed: %w", err)
	}

	return valid, nil
}

// Example 2: AML Profile with Field-Level Commitments
// ====================================================

// SetAMLProfileProtected demonstrates field-level commitment protection
//
// Pattern: Individual sensitive fields get their own commitments
// Benefits: Can selectively verify/reveal specific fields without exposing all data
func (k *Keeper) SetAMLProfileProtectedExample(
	ctx sdk.Context,
	address string,
	riskLevel types.AMLRiskLevel,
	sensitiveFields map[string]interface{}, // Off-chain sensitive data
) error {
	dataProtection := NewDataProtectionService()

	// Generate commitments for sensitive fields
	// Fields like: risk_factors, source_of_funds, occupation
	commitments, err := dataProtection.GenerateFieldCommitments(sensitiveFields)
	if err != nil {
		return fmt.Errorf("failed to generate field commitments: %w", err)
	}

	// For this example, we combine commitments into a single hash
	// In production, you might store individual commitments if proto schema supports it
	allCommitmentsData := make(map[string]string)
	for field, commitment := range commitments {
		allCommitmentsData[field] = dataProtection.CommitmentToHex(commitment)
	}

	// Generate master commitment for all sensitive fields
	masterCommitment, err := dataProtection.GenerateCommitment(allCommitmentsData)
	if err != nil {
		return fmt.Errorf("failed to generate master commitment: %w", err)
	}

	// Create AML profile with NO plaintext sensitive data
	profile := &types.AMLProfile{
		Address:        address,
		RiskLevel:      riskLevel,
		LastAssessment: timestamppb.Now(),
		// Note: In production, you'd add a sensitive_data_commitment field to the proto
		// and use: SensitiveDataCommitment: masterCommitment
	}

	if err := k.SetAMLProfile(ctx, profile); err != nil {
		return err
	}

	// Off-chain: Provider stores sensitiveFields mapped to commitments
	// Example: providerDB.StoreAMLFields(address, sensitiveFields, commitments)

	return nil
}

// Example 3: Suspicious Activity Report with Redaction
// =====================================================

// ReportSuspiciousActivityProtected demonstrates SAR with data protection
//
// Pattern: Store only commitment and risk indicators on-chain
// Full investigation details stored off-chain
func (k *Keeper) ReportSuspiciousActivityProtectedExample(
	ctx sdk.Context,
	activityID string,
	address string,
	activityType string,
	indicators []string, // Public risk indicators (not PII)
	detailedReport string, // Sensitive investigation notes (off-chain only)
) error {
	dataProtection := NewDataProtectionService()

	// Generate commitment for sensitive investigation details
	reportCommitment := dataProtection.GenerateCommitmentFromBytes([]byte(detailedReport))

	// Create suspicious activity record (minimal on-chain data)
	activity := &types.SuspiciousActivity{
		Id:           activityID,
		Address:      address,
		ActivityType: activityType,
		Indicators:   indicators, // Public, non-sensitive
		DetectedAt:   timestamppb.Now(),
		// Note: In production, add commitment field to proto
		// Description field should be deprecated or removed
	}

	if err := k.SetSuspiciousActivity(ctx, activity); err != nil {
		return err
	}

	// Off-chain: Store detailedReport in secure compliance database
	// Example: complianceDB.StoreSAR(activityID, detailedReport, reportCommitment)

	// Log with redaction
	redactedData := RedactSensitiveFields(
		map[string]interface{}{
			"activity_id":     activityID,
			"address":         address,
			"activity_type":   activityType,
			"detailed_report": detailedReport, // This will be redacted
		},
		[]string{"detailed_report"},
	)

	k.logger(ctx).Info("Suspicious activity reported", "data", redactedData)

	return nil
}

// Example 4: Tax Report with Privacy Protection
// ==============================================

// GenerateTaxReportProtected demonstrates privacy-preserving tax reporting
//
// Pattern: Store only aggregate values and commitment on-chain
// Individual transactions and PII stored off-chain
func (k *Keeper) GenerateTaxReportProtectedExample(
	ctx sdk.Context,
	address string,
	taxYear string,
	jurisdiction string,
	detailedTransactions []*types.TaxTransaction, // Off-chain only
	aggregateTotals map[string]string, // Safe to store (already aggregated)
) error {
	dataProtection := NewDataProtectionService()

	// Generate commitment for detailed transaction list
	// This allows verification without exposing individual transactions
	txCommitment, err := dataProtection.GenerateCommitment(detailedTransactions)
	if err != nil {
		return fmt.Errorf("failed to generate transaction commitment: %w", err)
	}

	// Create tax report with aggregate values only
	report := &types.TaxReport{
		Address:           address,
		TaxYear:           taxYear,
		Jurisdiction:      jurisdiction,
		ReportType:        "aggregate_summary",
		TotalIncome:       aggregateTotals["income"],
		TotalCapitalGains: aggregateTotals["gains"],
		TotalCapitalLosses: aggregateTotals["losses"],
		GeneratedAt:       timestamppb.Now(),
		// Note: Transactions field should not be populated with sensitive data
		// Add commitment field in production: TransactionsCommitment: txCommitment
	}

	// Store tax reports for address
	if err := k.SetTaxReport(ctx, report); err != nil {
		return err
	}

	// Off-chain: Store detailed transactions
	// Example: taxDB.StoreTransactions(address, taxYear, detailedTransactions, txCommitment)

	return nil
}

// Example 5: GDPR Data Erasure Implementation
// ============================================

// EraseGDPRDataImplementation demonstrates proper erasure with audit trail
//
// Pattern: Emit erasure event on-chain, actually delete off-chain
// Maintains immutable audit trail while satisfying right to erasure
func (k *Keeper) EraseGDPRDataImplementationExample(
	ctx sdk.Context,
	address string,
	erasureReason string,
) (string, error) {
	// Generate unique erasure event ID
	erasureEventID := fmt.Sprintf("erasure_%s_%d", address, ctx.BlockHeight())

	// Emit immutable erasure event (audit trail)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"gdpr_erasure_requested",
			sdk.NewAttribute("address", address),
			sdk.NewAttribute("erasure_event_id", erasureEventID),
			sdk.NewAttribute("reason", erasureReason),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute("timestamp", ctx.BlockTime().UTC().Format(time.RFC3339)),
		),
	)

	// Note: On-chain commitments remain (they reveal no PII due to pre-image resistance)
	// Off-chain providers MUST delete actual PII upon seeing this event:
	//
	// Example off-chain handler (pseudocode):
	//   eventListener.On("gdpr_erasure_requested", func(event) {
	//       address := event.GetAttribute("address")
	//       kycDB.DeletePII(address)
	//       amlDB.DeletePII(address)
	//       taxDB.DeletePII(address)
	//       auditLog.RecordErasure(address, event.GetAttribute("erasure_event_id"))
	//   })

	// Log erasure
	k.logger(ctx).Info(
		"GDPR erasure event emitted",
		"address", address,
		"event_id", erasureEventID,
	)

	return erasureEventID, nil
}

// Example 6: Query with Commitment Verification
// ==============================================

// QueryKYCWithVerification demonstrates secure query pattern
//
// Pattern: Return on-chain commitment, user verifies off-chain data
func (k *Keeper) QueryKYCWithVerificationExample(
	ctx sdk.Context,
	address string,
) (*types.KYCRecord, []byte, error) {
	// Retrieve KYC record
	record, err := k.GetKYCRecord(ctx, address)
	if err != nil {
		return nil, nil, err
	}

	// Return record and commitment separately
	// User can then:
	// 1. Request PII from off-chain provider
	// 2. Verify PII matches the commitment
	// 3. Trust the data only if verification passes

	return record, record.PiiCommitment, nil
}

// Example 7: Bulk Verification for Auditing
// ==========================================

// AuditKYCRecordsIntegrity performs bulk verification for compliance audits
//
// Use case: Auditor wants to verify data integrity across all records
func (k *Keeper) AuditKYCRecordsIntegrityExample(
	ctx sdk.Context,
	providerPIIData map[string]*PIIData, // Retrieved from off-chain provider
) (map[string]bool, []string, error) {
	// Get all KYC records
	records, err := k.GetAllKYCRecords(ctx)
	if err != nil {
		return nil, nil, err
	}

	dataProtection := NewDataProtectionService()
	verificationResults := make(map[string]bool)
	var failures []string

	// Verify each record
	for _, record := range records {
		pii, exists := providerPIIData[record.Address]
		if !exists {
			failures = append(failures, fmt.Sprintf("%s: PII not provided by provider", record.Address))
			verificationResults[record.Address] = false
			continue
		}

		// Verify commitment
		valid, err := dataProtection.VerifyPIICommitment(pii, record.PiiCommitment)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: verification error: %v", record.Address, err))
			verificationResults[record.Address] = false
			continue
		}

		verificationResults[record.Address] = valid
		if !valid {
			failures = append(failures, fmt.Sprintf("%s: commitment mismatch (data tampering detected)", record.Address))
		}
	}

	return verificationResults, failures, nil
}

// Example 8: Secure Logging
// ==========================

// LogComplianceOperation demonstrates secure logging with redaction
//
// Pattern: Always redact sensitive fields before logging
func (k *Keeper) LogComplianceOperationExample(
	ctx sdk.Context,
	operation string,
	data map[string]interface{},
	dataType string,
) {
	// Get sensitive fields list for this data type
	sensitiveFields := GetSensitiveFieldsList(dataType)

	// Redact sensitive data
	redactedData := RedactSensitiveFields(data, sensitiveFields)

	// Safe to log
	k.logger(ctx).Info(
		operation,
		"data_type", dataType,
		"data", redactedData,
		"block_height", ctx.BlockHeight(),
	)
}
