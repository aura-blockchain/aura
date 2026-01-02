package v1beta1

import (
	"fmt"

	"github.com/aequitas/aura/proto/common/validation"
)

const (
	// MaxLocales is the maximum number of locales an assistant can support
	MaxLocales = 50
	// MaxLocaleLength is the maximum length for locale strings (e.g., "en-US")
	MaxLocaleLength = 10
	// MinLocaleLength is the minimum length for locale strings
	MinLocaleLength = 2
	// MaxModelHashLength is the maximum length for model hashes
	MaxModelHashLength = 128
	// MinModelHashLength is the minimum length for model hashes
	MinModelHashLength = 32
	// MaxAPIKeyFingerprintLength is the maximum length for API key fingerprints
	MaxAPIKeyFingerprintLength = 128
	// MinAPIKeyFingerprintLength is the minimum length for API key fingerprints
	MinAPIKeyFingerprintLength = 32
	// MaxAttestationHashLength is the maximum length for attestation hashes
	MaxAttestationHashLength = 128
	// MinAttestationHashLength is the minimum length for attestation hashes
	MinAttestationHashLength = 32
	// MaxInfractionLength is the maximum length for infraction descriptions
	MaxInfractionLength = 500
	// MaxEvidenceHashLength is the maximum length for evidence hashes
	MaxEvidenceHashLength = 128
	// MinEvidenceHashLength is the minimum length for evidence hashes
	MinEvidenceHashLength = 32
)

// ValidateBasic implements the sdk.Msg interface for MsgRegisterAssistant
func (m *MsgRegisterAssistant) ValidateBasic() error {
	// Validate assistant address
	if err := validation.ValidateAccAddress(m.AssistantAddress); err != nil {
		return fmt.Errorf("assistant_address: %w", err)
	}

	// Validate owner address
	if err := validation.ValidateAccAddress(m.OwnerAddress); err != nil {
		return fmt.Errorf("owner_address: %w", err)
	}

	// Validate locales
	if len(m.Locales) == 0 {
		return fmt.Errorf("locales cannot be empty, at least one locale is required")
	}

	if len(m.Locales) > MaxLocales {
		return fmt.Errorf("locales cannot exceed %d, got %d", MaxLocales, len(m.Locales))
	}

	// Validate each locale
	for i, locale := range m.Locales {
		if err := validation.ValidateBoundedString(locale, MinLocaleLength, MaxLocaleLength, fmt.Sprintf("locales[%d]", i)); err != nil {
			return err
		}
	}

	// Validate model hash
	if err := validation.ValidateBoundedString(m.ModelHash, MinModelHashLength, MaxModelHashLength, "model_hash"); err != nil {
		return err
	}

	// Validate API key fingerprint
	if err := validation.ValidateBoundedString(m.ApiKeyFingerprint, MinAPIKeyFingerprintLength, MaxAPIKeyFingerprintLength, "api_key_fingerprint"); err != nil {
		return err
	}

	// Stake and sponsorship validation would be in the Balance type
	// At minimum, ensure they're not nil (enforced by gogoproto.nullable = false)

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateLocales
func (m *MsgUpdateLocales) ValidateBasic() error {
	// Validate assistant address
	if err := validation.ValidateAccAddress(m.AssistantAddress); err != nil {
		return fmt.Errorf("assistant_address: %w", err)
	}

	// Validate owner address
	if err := validation.ValidateAccAddress(m.OwnerAddress); err != nil {
		return fmt.Errorf("owner_address: %w", err)
	}

	// Validate locales
	if len(m.Locales) == 0 {
		return fmt.Errorf("locales cannot be empty, at least one locale is required")
	}

	if len(m.Locales) > MaxLocales {
		return fmt.Errorf("locales cannot exceed %d, got %d", MaxLocales, len(m.Locales))
	}

	// Validate each locale
	for i, locale := range m.Locales {
		if err := validation.ValidateBoundedString(locale, MinLocaleLength, MaxLocaleLength, fmt.Sprintf("locales[%d]", i)); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgHeartbeat
func (m *MsgHeartbeat) ValidateBasic() error {
	// Validate assistant address
	if err := validation.ValidateAccAddress(m.AssistantAddress); err != nil {
		return fmt.Errorf("assistant_address: %w", err)
	}

	// Validate operator address
	if err := validation.ValidateAccAddress(m.OperatorAddress); err != nil {
		return fmt.Errorf("operator_address: %w", err)
	}

	// Validate attestation hash
	if err := validation.ValidateBoundedString(m.AttestationHash, MinAttestationHashLength, MaxAttestationHashLength, "attestation_hash"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgReportMisbehavior
func (m *MsgReportMisbehavior) ValidateBasic() error {
	// Validate reporter address
	if err := validation.ValidateAccAddress(m.Reporter); err != nil {
		return fmt.Errorf("reporter: %w", err)
	}

	// Validate assistant address
	if err := validation.ValidateAccAddress(m.AssistantAddress); err != nil {
		return fmt.Errorf("assistant_address: %w", err)
	}

	// Validate infraction description
	if err := validation.ValidateBoundedString(m.Infraction, 1, MaxInfractionLength, "infraction"); err != nil {
		return err
	}

	// Validate evidence hash
	if err := validation.ValidateBoundedString(m.EvidenceHash, MinEvidenceHashLength, MaxEvidenceHashLength, "evidence_hash"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateParams
func (m *MsgUpdateParams) ValidateBasic() error {
	// Validate authority address
	if err := validation.ValidateAccAddress(m.Authority); err != nil {
		return fmt.Errorf("authority: %w", err)
	}

	// Params validation would be done by the keeper
	// Here we just ensure authority is valid

	return nil
}
