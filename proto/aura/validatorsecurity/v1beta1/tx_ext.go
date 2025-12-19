package v1beta1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m *MsgRegisterValidator) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ValidatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (m *MsgUpdateSecurityInfo) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ValidatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (m *MsgRegisterSentryNode) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ValidatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (m *MsgReportDoubleSign) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ReporterAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (m *MsgUnjail) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.ValidatorAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (m *MsgAcknowledgeAlert) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.AcknowledgerAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
