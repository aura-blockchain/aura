package v1beta1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetSigners returns the expected signers for MsgSubmitKYC.
// The provider must be the signer for KYC submissions.
func (m *MsgSubmitKYC) GetSigners() []sdk.AccAddress {
	if m.Provider == "" {
		return []sdk.AccAddress{}
	}
	addr, err := sdk.AccAddressFromBech32(m.Provider)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// GetSigners returns the expected signers for MsgReportSuspiciousActivity.
// The reporter must be the signer.
func (m *MsgReportSuspiciousActivity) GetSigners() []sdk.AccAddress {
	if m.Reporter == "" {
		return []sdk.AccAddress{}
	}
	addr, err := sdk.AccAddressFromBech32(m.Reporter)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// GetSigners returns the expected signers for MsgScreenSanctions.
// The address being screened must be the signer (user-initiated screening).
func (m *MsgScreenSanctions) GetSigners() []sdk.AccAddress {
	if m.Address == "" {
		return []sdk.AccAddress{}
	}
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// GetSigners returns the expected signers for MsgRecordGDPRConsent.
// The address giving/withdrawing consent must be the signer.
func (m *MsgRecordGDPRConsent) GetSigners() []sdk.AccAddress {
	if m.Address == "" {
		return []sdk.AccAddress{}
	}
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// GetSigners returns the expected signers for MsgRequestGDPRData.
// The address requesting data must be the signer.
func (m *MsgRequestGDPRData) GetSigners() []sdk.AccAddress {
	if m.Address == "" {
		return []sdk.AccAddress{}
	}
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}

// GetSigners returns the expected signers for MsgGenerateTaxReport.
// The address requesting the tax report must be the signer.
func (m *MsgGenerateTaxReport) GetSigners() []sdk.AccAddress {
	if m.Address == "" {
		return []sdk.AccAddress{}
	}
	addr, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return []sdk.AccAddress{}
	}
	return []sdk.AccAddress{addr}
}
