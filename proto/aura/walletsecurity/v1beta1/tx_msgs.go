package v1beta1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Ensure messages implement sdk.Msg interface
var (
	_ sdk.Msg = &MsgRegisterHardwareWallet{}
	_ sdk.Msg = &MsgCreateMultiSigWallet{}
	_ sdk.Msg = &MsgSignMultiSigTransaction{}
	_ sdk.Msg = &MsgConfigureSocialRecovery{}
	_ sdk.Msg = &MsgInitiateRecovery{}
	_ sdk.Msg = &MsgApproveRecovery{}
	_ sdk.Msg = &MsgExecuteRecovery{}
	_ sdk.Msg = &MsgSimulateTransaction{}
	_ sdk.Msg = &MsgVerifyDomain{}
	_ sdk.Msg = &MsgSetSpendingLimit{}
	_ sdk.Msg = &MsgConfigureSession{}
	_ sdk.Msg = &MsgEnrollBiometric{}
	_ sdk.Msg = &MsgAuthenticateBiometric{}
)

// GetSigners implementations for wallet security messages
// These implement the sdk.Msg interface which requires returning []sdk.AccAddress

func (msg *MsgRegisterHardwareWallet) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// Note: MsgCreateMultiSigWallet has a Signers field which generates GetSigners() []string
// We need to implement the sdk.Msg version that returns []sdk.AccAddress
func (msg *MsgCreateMultiSigWallet) GetSignersSDK() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgSignMultiSigTransaction) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgConfigureSocialRecovery) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgInitiateRecovery) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Initiator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgApproveRecovery) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Guardian)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgExecuteRecovery) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Executor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgSimulateTransaction) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgVerifyDomain) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Verifier)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgSetSpendingLimit) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgConfigureSession) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgEnrollBiometric) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.WalletId)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func (msg *MsgAuthenticateBiometric) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.WalletId)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
