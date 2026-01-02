package v1beta1

import (
	"fmt"

	"github.com/aequitas/aura/proto/common/validation"
)

const (
	// MaxPeerIDLength is the maximum length for peer IDs
	MaxPeerIDLength = 128
	// MinPeerIDLength is the minimum length for peer IDs
	MinPeerIDLength = 1
	// MaxReasonLength is the maximum length for ban reasons
	MaxReasonLength = 500
	// MinBanDuration is the minimum ban duration (1 hour in seconds)
	MinBanDuration = int64(60 * 60)
	// MaxBanDuration is the maximum ban duration (365 days in seconds)
	MaxBanDuration = int64(365 * 24 * 60 * 60)
	// MinReputationScore is the minimum reputation score
	MinReputationScore = int64(0)
	// MaxReputationScore is the maximum reputation score
	MaxReputationScore = int64(100)
	// MaxAlertIDLength is the maximum length for alert IDs
	MaxAlertIDLength = 128
)

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

// ValidateBasic implements the sdk.Msg interface for MsgAddTrustedPeer
func (m *MsgAddTrustedPeer) ValidateBasic() error {
	// Validate authority address
	if err := validation.ValidateAccAddress(m.Authority); err != nil {
		return fmt.Errorf("authority: %w", err)
	}

	// Validate peer (embedded message)
	// The peer validation would be in the TrustedPeer type
	// At minimum, ensure it's not nil (enforced by gogoproto.nullable = false)

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRemoveTrustedPeer
func (m *MsgRemoveTrustedPeer) ValidateBasic() error {
	// Validate authority address
	if err := validation.ValidateAccAddress(m.Authority); err != nil {
		return fmt.Errorf("authority: %w", err)
	}

	// Validate peer ID
	if err := validation.ValidateBoundedString(m.PeerId, MinPeerIDLength, MaxPeerIDLength, "peer_id"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgBanPeer
func (m *MsgBanPeer) ValidateBasic() error {
	// Validate authority address
	if err := validation.ValidateAccAddress(m.Authority); err != nil {
		return fmt.Errorf("authority: %w", err)
	}

	// Validate peer ID
	if err := validation.ValidateBoundedString(m.PeerId, MinPeerIDLength, MaxPeerIDLength, "peer_id"); err != nil {
		return err
	}

	// Validate ban duration
	if m.DurationSeconds < MinBanDuration {
		return fmt.Errorf("duration_seconds must be >= %d, got %d", MinBanDuration, m.DurationSeconds)
	}
	if m.DurationSeconds > MaxBanDuration {
		return fmt.Errorf("duration_seconds must be <= %d, got %d", MaxBanDuration, m.DurationSeconds)
	}

	// Validate reason
	if err := validation.ValidateBoundedString(m.Reason, 1, MaxReasonLength, "reason"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUnbanPeer
func (m *MsgUnbanPeer) ValidateBasic() error {
	// Validate authority address
	if err := validation.ValidateAccAddress(m.Authority); err != nil {
		return fmt.Errorf("authority: %w", err)
	}

	// Validate peer ID
	if err := validation.ValidateBoundedString(m.PeerId, MinPeerIDLength, MaxPeerIDLength, "peer_id"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdatePeerReputation
func (m *MsgUpdatePeerReputation) ValidateBasic() error {
	// Validate authority address
	if err := validation.ValidateAccAddress(m.Authority); err != nil {
		return fmt.Errorf("authority: %w", err)
	}

	// Validate peer ID
	if err := validation.ValidateBoundedString(m.PeerId, MinPeerIDLength, MaxPeerIDLength, "peer_id"); err != nil {
		return err
	}

	// Validate reputation score
	if m.Score < MinReputationScore {
		return fmt.Errorf("score must be >= %d, got %d", MinReputationScore, m.Score)
	}
	if m.Score > MaxReputationScore {
		return fmt.Errorf("score must be <= %d, got %d", MaxReputationScore, m.Score)
	}

	// Validate reason (optional)
	if m.Reason != "" {
		if err := validation.ValidateBoundedString(m.Reason, 0, MaxReasonLength, "reason"); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgResolveForkAlert
func (m *MsgResolveForkAlert) ValidateBasic() error {
	// Validate authority address
	if err := validation.ValidateAccAddress(m.Authority); err != nil {
		return fmt.Errorf("authority: %w", err)
	}

	// Validate alert ID
	if err := validation.ValidateBoundedString(m.AlertId, 1, MaxAlertIDLength, "alert_id"); err != nil {
		return err
	}

	// Validate resolution details (optional)
	if m.ResolutionDetails != "" {
		if err := validation.ValidateBoundedString(m.ResolutionDetails, 0, validation.MaxDescriptionLength, "resolution_details"); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgResolvePartitionAlert
func (m *MsgResolvePartitionAlert) ValidateBasic() error {
	// Validate authority address
	if err := validation.ValidateAccAddress(m.Authority); err != nil {
		return fmt.Errorf("authority: %w", err)
	}

	// Validate alert ID
	if err := validation.ValidateBoundedString(m.AlertId, 1, MaxAlertIDLength, "alert_id"); err != nil {
		return err
	}

	return nil
}
