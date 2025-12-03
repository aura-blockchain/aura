package v1beta1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetSigners returns the signers for MsgCreateKeyRotationSchedule
func (m *MsgCreateKeyRotationSchedule) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the signers for MsgRotateKey
func (m *MsgRotateKey) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the signers for MsgCreateThresholdScheme
func (m *MsgCreateThresholdScheme) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the signers for MsgSubmitThresholdSignatureShare
func (m *MsgSubmitThresholdSignatureShare) GetSigners() []sdk.AccAddress {
	submitter, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{submitter}
}

// GetSigners returns the signers for MsgRegisterZKProofCircuit
func (m *MsgRegisterZKProofCircuit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the signers for MsgSubmitZKProof
func (m *MsgSubmitZKProof) GetSigners() []sdk.AccAddress {
	submitter, err := sdk.AccAddressFromBech32(m.Submitter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{submitter}
}

// GetSigners returns the signers for MsgRegisterSecureEnclave
func (m *MsgRegisterSecureEnclave) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the signers for MsgGenerateQuantumResistantKey
func (m *MsgGenerateQuantumResistantKey) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the signers for MsgAddCertificatePin
func (m *MsgAddCertificatePin) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the signers for MsgUpdateParams
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}
