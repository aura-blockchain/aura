package v1beta1

import (
	"fmt"

	"github.com/aequitas/aura/proto/common/validation"
)

const (
	// MaxKeyLength is the maximum length for hot/cold keys
	MaxKeyLength = 256
	// MinKeyLength is the minimum length for hot/cold keys
	MinKeyLength = 32
	// MaxRegionLength is the maximum length for region strings
	MaxRegionLength = 128
	// MaxCountryCodeLength is the maximum length for country codes (ISO 3166-1 alpha-2)
	MaxCountryCodeLength = 3
	// MinCountryCodeLength is the minimum length for country codes
	MinCountryCodeLength = 2
	// MaxBackupValidators is the maximum number of backup validators
	MaxBackupValidators = 10
	// MaxIPAddressLength is the maximum length for IP addresses
	MaxIPAddressLength = 45
	// MinIPAddressLength is the minimum length for IP addresses
	MinIPAddressLength = 7
	// MinPort is the minimum port number
	MinPort = int32(1)
	// MaxPort is the maximum port number
	MaxPort = int32(65535)
	// MaxVoteSize is the maximum size for vote evidence
	MaxVoteSize = 4096
	// MinVoteSize is the minimum size for vote evidence
	MinVoteSize = 32
	// MaxAlertIDLength is the maximum length for alert IDs
	MaxAlertIDLength = 128
)

// ValidateBasic implements the sdk.Msg interface for MsgRegisterValidator
func (m *MsgRegisterValidator) ValidateBasic() error {
	// Validate validator address
	if err := validation.ValidateAccAddress(m.ValidatorAddress); err != nil {
		return fmt.Errorf("validator_address: %w", err)
	}

	// Validate hot key
	if err := validation.ValidateBoundedString(m.HotKey, MinKeyLength, MaxKeyLength, "hot_key"); err != nil {
		return err
	}

	// Validate cold key
	if err := validation.ValidateBoundedString(m.ColdKey, MinKeyLength, MaxKeyLength, "cold_key"); err != nil {
		return err
	}

	// Validate region (optional)
	if m.Region != "" {
		if err := validation.ValidateBoundedString(m.Region, 0, MaxRegionLength, "region"); err != nil {
			return err
		}
	}

	// Validate country code (optional)
	if m.CountryCode != "" {
		if err := validation.ValidateBoundedString(m.CountryCode, MinCountryCodeLength, MaxCountryCodeLength, "country_code"); err != nil {
			return err
		}
	}

	// Validate backup validators
	if len(m.BackupValidatorAddresses) > MaxBackupValidators {
		return fmt.Errorf("backup_validator_addresses cannot exceed %d, got %d", MaxBackupValidators, len(m.BackupValidatorAddresses))
	}

	for i, addr := range m.BackupValidatorAddresses {
		if err := validation.ValidateAccAddress(addr); err != nil {
			return fmt.Errorf("backup_validator_addresses[%d]: %w", i, err)
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateSecurityInfo
func (m *MsgUpdateSecurityInfo) ValidateBasic() error {
	// Validate validator address
	if err := validation.ValidateAccAddress(m.ValidatorAddress); err != nil {
		return fmt.Errorf("validator_address: %w", err)
	}

	// Validate hot key (optional)
	if m.HotKey != "" {
		if err := validation.ValidateBoundedString(m.HotKey, MinKeyLength, MaxKeyLength, "hot_key"); err != nil {
			return err
		}
	}

	// Validate cold key (optional)
	if m.ColdKey != "" {
		if err := validation.ValidateBoundedString(m.ColdKey, MinKeyLength, MaxKeyLength, "cold_key"); err != nil {
			return err
		}
	}

	// Validate region (optional)
	if m.Region != "" {
		if err := validation.ValidateBoundedString(m.Region, 0, MaxRegionLength, "region"); err != nil {
			return err
		}
	}

	// Validate country code (optional)
	if m.CountryCode != "" {
		if err := validation.ValidateBoundedString(m.CountryCode, MinCountryCodeLength, MaxCountryCodeLength, "country_code"); err != nil {
			return err
		}
	}

	// Validate backup validators
	if len(m.BackupValidatorAddresses) > MaxBackupValidators {
		return fmt.Errorf("backup_validator_addresses cannot exceed %d, got %d", MaxBackupValidators, len(m.BackupValidatorAddresses))
	}

	for i, addr := range m.BackupValidatorAddresses {
		if err := validation.ValidateAccAddress(addr); err != nil {
			return fmt.Errorf("backup_validator_addresses[%d]: %w", i, err)
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRegisterSentryNode
func (m *MsgRegisterSentryNode) ValidateBasic() error {
	// Validate validator address
	if err := validation.ValidateAccAddress(m.ValidatorAddress); err != nil {
		return fmt.Errorf("validator_address: %w", err)
	}

	// Validate sentry address
	if err := validation.ValidateAccAddress(m.SentryAddress); err != nil {
		return fmt.Errorf("sentry_address: %w", err)
	}

	// Validate IP address
	if err := validation.ValidateBoundedString(m.IpAddress, MinIPAddressLength, MaxIPAddressLength, "ip_address"); err != nil {
		return err
	}

	// Validate port
	if m.Port < MinPort || m.Port > MaxPort {
		return fmt.Errorf("port must be between %d and %d, got %d", MinPort, MaxPort, m.Port)
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgReportDoubleSign
func (m *MsgReportDoubleSign) ValidateBasic() error {
	// Validate reporter address
	if err := validation.ValidateAccAddress(m.ReporterAddress); err != nil {
		return fmt.Errorf("reporter_address: %w", err)
	}

	// Validate validator address
	if err := validation.ValidateAccAddress(m.ValidatorAddress); err != nil {
		return fmt.Errorf("validator_address: %w", err)
	}

	// Validate height
	if m.Height <= 0 {
		return fmt.Errorf("height must be positive, got %d", m.Height)
	}

	// Validate vote A
	if err := validation.ValidateBytes(m.VoteA, MinVoteSize, MaxVoteSize, "vote_a"); err != nil {
		return err
	}

	// Validate vote B
	if err := validation.ValidateBytes(m.VoteB, MinVoteSize, MaxVoteSize, "vote_b"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUnjail
func (m *MsgUnjail) ValidateBasic() error {
	// Validate validator address
	if err := validation.ValidateAccAddress(m.ValidatorAddress); err != nil {
		return fmt.Errorf("validator_address: %w", err)
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgAcknowledgeAlert
func (m *MsgAcknowledgeAlert) ValidateBasic() error {
	// Validate acknowledger address (validator acknowledging the security alert)
	if err := validation.ValidateAccAddress(m.AcknowledgerAddress); err != nil {
		return fmt.Errorf("acknowledger_address: %w", err)
	}

	// Validate alert ID
	if err := validation.ValidateBoundedString(m.AlertId, 1, MaxAlertIDLength, "alert_id"); err != nil {
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
