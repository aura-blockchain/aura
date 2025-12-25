// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
	compliancepb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
)

func TestRegisterLegacyAminoCodec(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	require.NotPanics(t, func() {
		types.RegisterLegacyAminoCodec(cdc)
	})
}

func TestRegisterInterfaces(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()

	require.NotPanics(t, func() {
		types.RegisterInterfaces(registry)
	})

	// Verify message implementations are registered
	msgSubmitKYC := &compliancepb.MsgSubmitKYC{}
	require.Implements(t, (*sdk.Msg)(nil), msgSubmitKYC)

	msgReportSuspicious := &compliancepb.MsgReportSuspiciousActivity{}
	require.Implements(t, (*sdk.Msg)(nil), msgReportSuspicious)

	msgScreenSanctions := &compliancepb.MsgScreenSanctions{}
	require.Implements(t, (*sdk.Msg)(nil), msgScreenSanctions)

	msgRecordGDPR := &compliancepb.MsgRecordGDPRConsent{}
	require.Implements(t, (*sdk.Msg)(nil), msgRecordGDPR)

	msgRequestGDPR := &compliancepb.MsgRequestGDPRData{}
	require.Implements(t, (*sdk.Msg)(nil), msgRequestGDPR)

	msgEraseGDPR := &compliancepb.MsgEraseGDPRData{}
	require.Implements(t, (*sdk.Msg)(nil), msgEraseGDPR)

	msgGenerateTax := &compliancepb.MsgGenerateTaxReport{}
	require.Implements(t, (*sdk.Msg)(nil), msgGenerateTax)
}

func TestCodecRegistration(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)

	// Test that we can resolve registered message types
	msgURL := sdk.MsgTypeURL(&compliancepb.MsgSubmitKYC{})
	require.NotEmpty(t, msgURL)

	msgURL = sdk.MsgTypeURL(&compliancepb.MsgReportSuspiciousActivity{})
	require.NotEmpty(t, msgURL)

	msgURL = sdk.MsgTypeURL(&compliancepb.MsgScreenSanctions{})
	require.NotEmpty(t, msgURL)
}
