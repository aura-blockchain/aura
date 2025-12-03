package v1beta1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetSigners returns the expected signers for MsgMintVC
func (m *MsgMintVC) GetSigners() []sdk.AccAddress {
	holder, err := sdk.AccAddressFromBech32(m.HolderAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{holder}
}

// GetSigners returns the expected signers for MsgRevokeVC
func (m *MsgRevokeVC) GetSigners() []sdk.AccAddress {
	holder, err := sdk.AccAddressFromBech32(m.HolderAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{holder}
}

// GetSigners returns the expected signers for MsgAdminRevokeVC
func (m *MsgAdminRevokeVC) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSigners returns the expected signers for MsgSuspendVC
func (m *MsgSuspendVC) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSigners returns the expected signers for MsgReactivateVC
func (m *MsgReactivateVC) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSigners returns the expected signers for MsgCreateVCPolicy
func (m *MsgCreateVCPolicy) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSigners returns the expected signers for MsgUpdateVCPolicy
func (m *MsgUpdateVCPolicy) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSigners returns the expected signers for MsgDeprecateVCPolicy
func (m *MsgDeprecateVCPolicy) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// GetSigners returns the expected signers for MsgRegisterDID
func (m *MsgRegisterDID) GetSigners() []sdk.AccAddress {
	controller, err := sdk.AccAddressFromBech32(m.Controller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{controller}
}

// GetSigners returns the expected signers for MsgUpdateDIDDocument
func (m *MsgUpdateDIDDocument) GetSigners() []sdk.AccAddress {
	controller, err := sdk.AccAddressFromBech32(m.Controller)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{controller}
}
