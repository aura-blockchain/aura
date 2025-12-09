package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	compliancepb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
)

// RegisterLegacyAminoCodec registers the module messages on the legacy amino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	_ = cdc
}

// RegisterInterfaces registers compliance protobuf interfaces.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(registry, &compliancepb.Msg_serviceDesc)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&compliancepb.MsgSubmitKYC{},
		&compliancepb.MsgReportSuspiciousActivity{},
		&compliancepb.MsgScreenSanctions{},
		&compliancepb.MsgRecordGDPRConsent{},
		&compliancepb.MsgRequestGDPRData{},
		&compliancepb.MsgEraseGDPRData{},
		&compliancepb.MsgGenerateTaxReport{},
	)

	registry.RegisterImplementations((*sdktx.MsgResponse)(nil),
		&compliancepb.MsgSubmitKYCResponse{},
		&compliancepb.MsgReportSuspiciousActivityResponse{},
		&compliancepb.MsgScreenSanctionsResponse{},
		&compliancepb.MsgRecordGDPRConsentResponse{},
		&compliancepb.MsgRequestGDPRDataResponse{},
		&compliancepb.MsgEraseGDPRDataResponse{},
		&compliancepb.MsgGenerateTaxReportResponse{},
	)
}
