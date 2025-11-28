package v1beta1

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/aequitas/aura/chain/x/common/validation"
)

const (
	// MaxVetoIDLength is the maximum length for veto IDs
	MaxVetoIDLength = 128
	// MaxSnapshotIDLength is the maximum length for snapshot IDs
	MaxSnapshotIDLength = 128
	// MaxVoteCommitmentLength is the maximum length for vote commitments
	MaxVoteCommitmentLength = 256
	// MaxSecretLength is the maximum length for secret votes
	MaxSecretLength = 256
	// MaxSignatureLength is the maximum length for cryptographic signatures (base64 encoded)
	MaxSignatureLength = 512
	// MaxWeightedVoteOptions is the maximum number of weighted vote options
	MaxWeightedVoteOptions = 10
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

// ValidateBasic implements the sdk.Msg interface for MsgSubmitProposal
func (m *MsgSubmitProposal) ValidateBasic() error {
	// Validate proposer address
	if err := validation.ValidateAccAddress(m.Proposer); err != nil {
		return fmt.Errorf("proposer: %w", err)
	}

	// Validate title
	if err := validation.ValidateBoundedString(m.Title, 1, validation.MaxNameLength, "title"); err != nil {
		return err
	}

	// Validate description
	if err := validation.ValidateBoundedString(m.Description, 1, validation.MaxDescriptionLength, "description"); err != nil {
		return err
	}

	// Category enum is validated at protobuf level

	// Validate initial deposit (if provided)
	if m.InitialDeposit != "" {
		if err := parseAndValidatePositiveInt(m.InitialDeposit, "initial_deposit"); err != nil {
			return err
		}
	}

	// IsEmergency is a boolean, no validation needed

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgDeposit
func (m *MsgDeposit) ValidateBasic() error {
	// Validate depositor address
	if err := validation.ValidateAccAddress(m.Depositor); err != nil {
		return fmt.Errorf("depositor: %w", err)
	}

	// Validate proposal ID
	if m.ProposalId == 0 {
		return fmt.Errorf("proposal_id must be greater than 0")
	}

	// Validate amount
	if err := parseAndValidatePositiveInt(m.Amount, "amount"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgVote
func (m *MsgVote) ValidateBasic() error {
	// Validate voter address
	if err := validation.ValidateAccAddress(m.Voter); err != nil {
		return fmt.Errorf("voter: %w", err)
	}

	// Validate proposal ID
	if m.ProposalId == 0 {
		return fmt.Errorf("proposal_id must be greater than 0")
	}

	// Option enum is validated at protobuf level

	// If secret ballot, validate vote commitment
	if m.IsSecret {
		if err := validation.ValidateBoundedString(m.VoteCommitment, 1, MaxVoteCommitmentLength, "vote_commitment"); err != nil {
			return err
		}
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgVoteWeighted
func (m *MsgVoteWeighted) ValidateBasic() error {
	// Validate voter address
	if err := validation.ValidateAccAddress(m.Voter); err != nil {
		return fmt.Errorf("voter: %w", err)
	}

	// Validate proposal ID
	if m.ProposalId == 0 {
		return fmt.Errorf("proposal_id must be greater than 0")
	}

	// Validate options
	if len(m.Options) == 0 {
		return fmt.Errorf("options cannot be empty")
	}

	if len(m.Options) > MaxWeightedVoteOptions {
		return fmt.Errorf("options cannot exceed %d, got %d", MaxWeightedVoteOptions, len(m.Options))
	}

	// Validate each option
	totalWeight := sdkmath.LegacyZeroDec()
	for i, opt := range m.Options {
		// Option enum is validated at protobuf level

		// Validate weight
		weight, err := sdkmath.LegacyNewDecFromStr(opt.Weight)
		if err != nil {
			return fmt.Errorf("options[%d].weight: invalid decimal: %w", i, err)
		}

		if weight.IsNegative() || weight.GT(sdkmath.LegacyOneDec()) {
			return fmt.Errorf("options[%d].weight must be between 0 and 1, got %s", i, weight.String())
		}

		totalWeight = totalWeight.Add(weight)
	}

	// Total weight should equal 1
	if !totalWeight.Equal(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("total weight must equal 1, got %s", totalWeight.String())
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgDelegateVote
func (m *MsgDelegateVote) ValidateBasic() error {
	// Validate delegator address
	if err := validation.ValidateAccAddress(m.Delegator); err != nil {
		return fmt.Errorf("delegator: %w", err)
	}

	// Validate delegate address
	if err := validation.ValidateAccAddress(m.Delegate); err != nil {
		return fmt.Errorf("delegate: %w", err)
	}

	// Cannot delegate to self
	if m.Delegator == m.Delegate {
		return fmt.Errorf("cannot delegate to self")
	}

	// Categories enum slice is validated at protobuf level

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgUndelegateVote
func (m *MsgUndelegateVote) ValidateBasic() error {
	// Validate delegator address
	if err := validation.ValidateAccAddress(m.Delegator); err != nil {
		return fmt.Errorf("delegator: %w", err)
	}

	// Categories enum slice is validated at protobuf level

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgSubmitVeto
func (m *MsgSubmitVeto) ValidateBasic() error {
	// Validate vetoer address (governance guardian with veto power)
	if err := validation.ValidateAccAddress(m.Vetoer); err != nil {
		return fmt.Errorf("vetoer: %w", err)
	}

	// Validate proposal ID
	if m.ProposalId == 0 {
		return fmt.Errorf("proposal_id must be greater than 0")
	}

	// Validate reason (required for transparency and governance records)
	if err := validation.ValidateBoundedString(m.Reason, 1, validation.MaxDescriptionLength, "reason"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgCosignVeto
func (m *MsgCosignVeto) ValidateBasic() error {
	// Validate cosigner address (must be authorized governance guardian)
	if err := validation.ValidateAccAddress(m.Cosigner); err != nil {
		return fmt.Errorf("cosigner: %w", err)
	}

	// Validate proposal ID
	if m.ProposalId == 0 {
		return fmt.Errorf("proposal_id must be greater than 0")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgExecuteProposal
func (m *MsgExecuteProposal) ValidateBasic() error {
	// Validate executor address
	if err := validation.ValidateAccAddress(m.Executor); err != nil {
		return fmt.Errorf("executor: %w", err)
	}

	// Validate proposal ID
	if m.ProposalId == 0 {
		return fmt.Errorf("proposal_id must be greater than 0")
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgSubmitSnapshotVote
func (m *MsgSubmitSnapshotVote) ValidateBasic() error {
	// Validate voter address
	if err := validation.ValidateAccAddress(m.Voter); err != nil {
		return fmt.Errorf("voter: %w", err)
	}

	// Validate proposal ID
	if m.ProposalId == 0 {
		return fmt.Errorf("proposal_id must be greater than 0")
	}

	// Option enum is validated at protobuf level

	// Validate cryptographic signature (ensures vote authenticity and non-repudiation)
	if err := validation.ValidateBoundedString(m.Signature, 1, MaxSignatureLength, "signature"); err != nil {
		return err
	}

	return nil
}

// ValidateBasic implements the sdk.Msg interface for MsgRevealSecretVote
func (m *MsgRevealSecretVote) ValidateBasic() error {
	// Validate voter address
	if err := validation.ValidateAccAddress(m.Voter); err != nil {
		return fmt.Errorf("voter: %w", err)
	}

	// Validate proposal ID
	if m.ProposalId == 0 {
		return fmt.Errorf("proposal_id must be greater than 0")
	}

	// Option enum is validated at protobuf level

	// Validate reveal key (cryptographic key to decrypt committed vote)
	if err := validation.ValidateBoundedString(m.RevealKey, 1, MaxSecretLength, "reveal_key"); err != nil {
		return err
	}

	return nil
}
