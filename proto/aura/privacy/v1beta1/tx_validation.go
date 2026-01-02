package v1beta1

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/aequitas/aura/proto/common/validation"
)

const (
	// MaxPoolIDLength is the maximum length for pool IDs
	MaxPoolIDLength = 128
	// MaxViewKeyLength is the maximum length for view keys
	MaxViewKeyLength = 256
	// MinParticipants is the minimum number of participants in a mixing pool
	MinParticipants = uint32(2)
	// MaxParticipants is the maximum number of participants in a mixing pool
	MaxParticipants = uint32(1000)
	// MinMixingRounds is the minimum number of mixing rounds
	MinMixingRounds = uint32(1)
	// MaxMixingRounds is the maximum number of mixing rounds
	MaxMixingRounds = uint32(100)
	// MinDeadlineDuration is the minimum deadline duration (1 hour in seconds)
	MinDeadlineDuration = uint64(60 * 60)
	// MaxDeadlineDuration is the maximum deadline duration (7 days in seconds)
	MaxDeadlineDuration = uint64(7 * 24 * 60 * 60)
	// MaxProofSize is the maximum size for ZK proofs
	MaxProofSize = 1024 * 1024
	// MinProofSize is the minimum size for ZK proofs
	MinProofSize = 32
	// MaxCommitmentSize is the maximum size for commitments
	MaxCommitmentSize = 256
	// MinCommitmentSize is the minimum size for commitments
	MinCommitmentSize = 32
)

// parseAndValidatePositiveInt parses a string to Int and validates it's positive
func parseAndValidatePositiveInt(s string, fieldName string) error {
	if s == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	val, ok := sdkmath.NewIntFromString(s)
	if !ok {
		return fmt.Errorf("%s must be a valid integer, got: %s", fieldName, s)
	}
	return validation.ValidatePositiveInt(val, fieldName)
}

// ValidateBasic implements the sdk.Msg interface for MsgSubmitPrivateTransaction
func (m *MsgSubmitPrivateTransaction) ValidateBasic() error {
	// Validate sender address
	if err := validation.ValidateAccAddress(m.Sender); err != nil {
		return fmt.Errorf("sender: %w", err)
	}

	// Validate private transaction (embedded message)
	// The private_transaction validation would be in the PrivateTransaction type
	// At minimum, ensure it's not nil
	if m.PrivateTransaction == nil {
		return fmt.Errorf("private_transaction cannot be nil")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCreateMixingPool
func (m *MsgCreateMixingPool) ValidateBasic() error {
	// Validate creator address
	if err := validation.ValidateAccAddress(m.Creator); err != nil {
		return fmt.Errorf("creator: %w", err)
	}

	// Validate min participants
	if m.MinParticipants < MinParticipants {
		return fmt.Errorf("min_participants must be >= %d, got %d", MinParticipants, m.MinParticipants)
	}

	// Validate max participants
	if m.MaxParticipants > MaxParticipants {
		return fmt.Errorf("max_participants cannot exceed %d, got %d", MaxParticipants, m.MaxParticipants)
	}

	// Min must be <= max
	if m.MinParticipants > m.MaxParticipants {
		return fmt.Errorf("min_participants must be <= max_participants, got %d > %d", m.MinParticipants, m.MaxParticipants)
	}

	// Validate denomination (customtype is already math.Int, not string)
	if err := validation.ValidatePositiveInt(m.Denomination, "denomination"); err != nil {
		return err
	}

	// Validate mixing rounds
	if m.MixingRounds < MinMixingRounds {
		return fmt.Errorf("mixing_rounds must be >= %d, got %d", MinMixingRounds, m.MixingRounds)
	}
	if m.MixingRounds > MaxMixingRounds {
		return fmt.Errorf("mixing_rounds cannot exceed %d, got %d", MaxMixingRounds, m.MixingRounds)
	}

	// Validate deadline duration
	if m.DeadlineDuration < MinDeadlineDuration {
		return fmt.Errorf("deadline_duration must be >= %d, got %d", MinDeadlineDuration, m.DeadlineDuration)
	}
	if m.DeadlineDuration > MaxDeadlineDuration {
		return fmt.Errorf("deadline_duration cannot exceed %d, got %d", MaxDeadlineDuration, m.DeadlineDuration)
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgJoinMixingPool
func (m *MsgJoinMixingPool) ValidateBasic() error {
	// Validate participant address
	if err := validation.ValidateAccAddress(m.Participant); err != nil {
		return fmt.Errorf("participant: %w", err)
	}

	// Validate pool ID
	if err := validation.ValidateBoundedString(m.PoolId, 1, MaxPoolIDLength, "pool_id"); err != nil {
		return err
	}

	// Validate commitment
	if err := validation.ValidateBytes(m.Commitment, MinCommitmentSize, MaxCommitmentSize, "commitment"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRegisterViewKey
func (m *MsgRegisterViewKey) ValidateBasic() error {
	// Validate owner address
	if err := validation.ValidateAccAddress(m.Owner); err != nil {
		return fmt.Errorf("owner: %w", err)
	}

	// Validate view key is not nil
	if m.ViewKey == nil {
		return fmt.Errorf("view_key cannot be nil")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRevokeViewKey
func (m *MsgRevokeViewKey) ValidateBasic() error {
	// Validate owner address
	if err := validation.ValidateAccAddress(m.Owner); err != nil {
		return fmt.Errorf("owner: %w", err)
	}

	// Validate public view key
	if err := validation.ValidateBytes(m.PublicViewKey, MinCommitmentSize, MaxCommitmentSize, "public_view_key"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUpdateNetworkPrivacy
func (m *MsgUpdateNetworkPrivacy) ValidateBasic() error {
	// Validate sender address
	if err := validation.ValidateAccAddress(m.Sender); err != nil {
		return fmt.Errorf("sender: %w", err)
	}

	// Validate network privacy is not nil
	if m.NetworkPrivacy == nil {
		return fmt.Errorf("network_privacy cannot be nil")
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
