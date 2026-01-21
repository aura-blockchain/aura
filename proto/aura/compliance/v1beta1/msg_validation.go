package v1beta1

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ValidateBasic performs stateless validation of MsgSubmitKYC.
// This method is called automatically by the Cosmos SDK before message processing.
func (m *MsgSubmitKYC) ValidateBasic() error {
	// Validate provider address (the signer)
	if m.Provider == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "provider is required")
	}
	_, err := sdk.AccAddressFromBech32(m.Provider)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid provider address: %s", err)
	}

	// Validate address being verified
	if m.Address == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "address is required")
	}
	_, err = sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address: %s", err)
	}

	// Validate KYC level
	if m.KycLevel == KYCLevel_KYC_LEVEL_UNSPECIFIED {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "kyc_level must be specified")
	}

	// Validate PII commitment (must be 32 bytes SHA-256 hash)
	if len(m.PiiCommitment) != 32 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"pii_commitment must be 32 bytes (SHA-256 hash), got %d bytes", len(m.PiiCommitment))
	}

	// Validate jurisdiction (ISO 3166-1 alpha-2 country code)
	if m.Jurisdiction == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "jurisdiction is required (ISO 3166-1 alpha-2 country code)")
	}
	if len(m.Jurisdiction) != 2 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"jurisdiction must be 2-letter ISO 3166-1 alpha-2 country code, got %d characters", len(m.Jurisdiction))
	}

	return nil
}

// ValidateBasic performs stateless validation of MsgReportSuspiciousActivity.
func (m *MsgReportSuspiciousActivity) ValidateBasic() error {
	// Validate reporter address (the signer)
	if m.Reporter == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "reporter is required")
	}
	_, err := sdk.AccAddressFromBech32(m.Reporter)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid reporter address: %s", err)
	}

	// Validate address being reported
	if m.Address == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "address is required")
	}
	_, err = sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address: %s", err)
	}

	// Validate transaction hash
	if m.TransactionHash == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "transaction_hash is required")
	}

	// Validate activity type
	if m.ActivityType == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "activity_type is required")
	}

	// Description is optional but validated for length if present
	if len(m.Description) > 1000 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"description too long (max 1000 characters), got %d", len(m.Description))
	}

	return nil
}

// ValidateBasic performs stateless validation of MsgScreenSanctions.
func (m *MsgScreenSanctions) ValidateBasic() error {
	// Validate address (must be the signer)
	if m.Address == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "address is required")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address: %s", err)
	}

	// ForceRefresh is a boolean flag, no validation needed

	return nil
}

// ValidateBasic performs stateless validation of MsgRecordGDPRConsent.
func (m *MsgRecordGDPRConsent) ValidateBasic() error {
	// Validate address (must be the signer)
	if m.Address == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "address is required")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address: %s", err)
	}

	// Validate consent type
	if m.ConsentType == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "consent_type is required")
	}

	// Validate consent version
	if m.ConsentVersion == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "consent_version is required")
	}

	// Consented is a boolean flag, no validation needed

	return nil
}

// ValidateBasic performs stateless validation of MsgRequestGDPRData.
func (m *MsgRequestGDPRData) ValidateBasic() error {
	// Validate address (must be the signer)
	if m.Address == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "address is required")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address: %s", err)
	}

	// Validate request type (access, rectification, erasure, portability)
	if m.RequestType == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "request_type is required")
	}

	// Validate request type is one of the valid GDPR rights
	validTypes := map[string]bool{
		"access":        true, // Article 15
		"rectification": true, // Article 16
		"erasure":       true, // Article 17
		"portability":   true, // Article 20
		"restriction":   true, // Article 18
		"objection":     true, // Article 21
	}
	if !validTypes[m.RequestType] {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"invalid request_type '%s', must be one of: access, rectification, erasure, portability, restriction, objection",
			m.RequestType)
	}

	return nil
}

// ValidateBasic performs stateless validation of MsgEraseGDPRData.
func (m *MsgEraseGDPRData) ValidateBasic() error {
	// Validate address (must be the signer)
	if m.Address == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "address is required")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address: %s", err)
	}

	// Erasure reason is optional but validated for length if present
	if len(m.ErasureReason) > 500 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"erasure_reason too long (max 500 characters), got %d", len(m.ErasureReason))
	}

	return nil
}

// ValidateBasic performs stateless validation of MsgGenerateTaxReport.
func (m *MsgGenerateTaxReport) ValidateBasic() error {
	// Validate address (must be the signer)
	if m.Address == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "address is required")
	}
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "invalid address: %s", err)
	}

	// Validate tax year (format: YYYY)
	if m.TaxYear == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "tax_year is required")
	}
	if len(m.TaxYear) != 4 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"tax_year must be 4 digits (YYYY format), got '%s'", m.TaxYear)
	}
	// Validate tax year is numeric
	for _, ch := range m.TaxYear {
		if ch < '0' || ch > '9' {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
				"tax_year must be numeric (YYYY format), got '%s'", m.TaxYear)
		}
	}

	// Validate jurisdiction (ISO 3166-1 alpha-2 country code)
	if m.Jurisdiction == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "jurisdiction is required (ISO 3166-1 alpha-2 country code)")
	}
	if len(m.Jurisdiction) != 2 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"jurisdiction must be 2-letter ISO 3166-1 alpha-2 country code, got %d characters", len(m.Jurisdiction))
	}

	// Validate report type
	if m.ReportType == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "report_type is required")
	}

	// Validate report type is one of the valid tax forms
	validReportTypes := map[string]bool{
		"1099-MISC":  true, // US: Miscellaneous Income
		"1099-K":     true, // US: Payment Card and Third Party Network Transactions
		"1099-B":     true, // US: Proceeds from Broker and Barter Exchange Transactions
		"8949":       true, // US: Sales and Other Dispositions of Capital Assets
		"Schedule D": true, // US: Capital Gains and Losses
		"generic":    true, // Generic format for non-US jurisdictions
	}
	if !validReportTypes[m.ReportType] {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"invalid report_type '%s', must be one of: 1099-MISC, 1099-K, 1099-B, 8949, Schedule D, generic",
			m.ReportType)
	}

	// Validate jurisdiction-reportType compatibility
	if m.Jurisdiction == "US" {
		// US jurisdiction should use US tax forms
		if m.ReportType == "generic" {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest,
				"US jurisdiction should use specific US tax forms (1099-MISC, 1099-K, 1099-B, 8949, Schedule D)")
		}
	} else {
		// Non-US jurisdictions should typically use generic format
		// (unless they have specific US reporting requirements)
		if m.ReportType != "generic" {
			return fmt.Errorf("non-US jurisdiction '%s' should typically use 'generic' report type", m.Jurisdiction)
		}
	}

	return nil
}
