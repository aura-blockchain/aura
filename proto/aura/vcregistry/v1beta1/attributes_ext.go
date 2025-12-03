package v1beta1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetSigners returns the expected signers for MsgCreateAttributeVC
func (m *MsgCreateAttributeVC) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the expected signers for MsgRevokeAttributeVC
func (m *MsgRevokeAttributeVC) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the expected signers for MsgUpdateDisclosurePolicy
func (m *MsgUpdateDisclosurePolicy) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the expected signers for MsgCreateDisclosureRequest
func (m *MsgCreateDisclosureRequest) GetSigners() []sdk.AccAddress {
	// Note: This is created by verifier, not the holder
	// The verifier is the one signing this transaction
	verifier, err := sdk.AccAddressFromBech32(m.Verifier)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{verifier}
}

// GetSigners returns the expected signers for MsgRespondToDisclosureRequest
func (m *MsgRespondToDisclosureRequest) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
