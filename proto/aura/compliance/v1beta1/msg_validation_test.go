package v1beta1_test

import (
	"crypto/sha256"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	compliancepb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
)

func TestMsgSubmitKYC_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("test_address_123456")).String()
	validProvider := sdk.AccAddress([]byte("provider_address12")).String()
	validCommitment := sha256.Sum256([]byte("test pii data"))

	tests := []struct {
		name      string
		msg       *compliancepb.MsgSubmitKYC
		wantErr   bool
		errString string
	}{
		{
			name: "valid message",
			msg: &compliancepb.MsgSubmitKYC{
				Address:       validAddr,
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      validProvider,
				PiiCommitment: validCommitment[:],
				Jurisdiction:  "US",
			},
			wantErr: false,
		},
		{
			name: "empty provider",
			msg: &compliancepb.MsgSubmitKYC{
				Address:       validAddr,
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      "",
				PiiCommitment: validCommitment[:],
				Jurisdiction:  "US",
			},
			wantErr:   true,
			errString: "provider is required",
		},
		{
			name: "invalid provider address",
			msg: &compliancepb.MsgSubmitKYC{
				Address:       validAddr,
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      "invalid-address",
				PiiCommitment: validCommitment[:],
				Jurisdiction:  "US",
			},
			wantErr:   true,
			errString: "invalid provider address",
		},
		{
			name: "empty address",
			msg: &compliancepb.MsgSubmitKYC{
				Address:       "",
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      validProvider,
				PiiCommitment: validCommitment[:],
				Jurisdiction:  "US",
			},
			wantErr:   true,
			errString: "address is required",
		},
		{
			name: "invalid address",
			msg: &compliancepb.MsgSubmitKYC{
				Address:       "invalid-address",
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      validProvider,
				PiiCommitment: validCommitment[:],
				Jurisdiction:  "US",
			},
			wantErr:   true,
			errString: "invalid address",
		},
		{
			name: "unspecified kyc level",
			msg: &compliancepb.MsgSubmitKYC{
				Address:       validAddr,
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_UNSPECIFIED,
				Provider:      validProvider,
				PiiCommitment: validCommitment[:],
				Jurisdiction:  "US",
			},
			wantErr:   true,
			errString: "kyc_level must be specified",
		},
		{
			name: "invalid pii commitment length",
			msg: &compliancepb.MsgSubmitKYC{
				Address:       validAddr,
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      validProvider,
				PiiCommitment: []byte("short"),
				Jurisdiction:  "US",
			},
			wantErr:   true,
			errString: "pii_commitment must be 32 bytes",
		},
		{
			name: "empty jurisdiction",
			msg: &compliancepb.MsgSubmitKYC{
				Address:       validAddr,
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      validProvider,
				PiiCommitment: validCommitment[:],
				Jurisdiction:  "",
			},
			wantErr:   true,
			errString: "jurisdiction is required",
		},
		{
			name: "invalid jurisdiction length",
			msg: &compliancepb.MsgSubmitKYC{
				Address:       validAddr,
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      validProvider,
				PiiCommitment: validCommitment[:],
				Jurisdiction:  "USA",
			},
			wantErr:   true,
			errString: "jurisdiction must be 2-letter ISO 3166-1 alpha-2 country code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgReportSuspiciousActivity_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("test_address_123456")).String()
	validReporter := sdk.AccAddress([]byte("reporter_address12")).String()

	tests := []struct {
		name      string
		msg       *compliancepb.MsgReportSuspiciousActivity
		wantErr   bool
		errString string
	}{
		{
			name: "valid message",
			msg: &compliancepb.MsgReportSuspiciousActivity{
				Reporter:        validReporter,
				Address:         validAddr,
				TransactionHash: "hash123",
				ActivityType:    "structuring",
				Description:     "suspicious pattern detected",
				Indicators:      []string{"velocity", "amount"},
			},
			wantErr: false,
		},
		{
			name: "empty reporter",
			msg: &compliancepb.MsgReportSuspiciousActivity{
				Reporter:        "",
				Address:         validAddr,
				TransactionHash: "hash123",
				ActivityType:    "structuring",
			},
			wantErr:   true,
			errString: "reporter is required",
		},
		{
			name: "invalid reporter address",
			msg: &compliancepb.MsgReportSuspiciousActivity{
				Reporter:        "invalid-address",
				Address:         validAddr,
				TransactionHash: "hash123",
				ActivityType:    "structuring",
			},
			wantErr:   true,
			errString: "invalid reporter address",
		},
		{
			name: "empty address",
			msg: &compliancepb.MsgReportSuspiciousActivity{
				Reporter:        validReporter,
				Address:         "",
				TransactionHash: "hash123",
				ActivityType:    "structuring",
			},
			wantErr:   true,
			errString: "address is required",
		},
		{
			name: "empty transaction hash",
			msg: &compliancepb.MsgReportSuspiciousActivity{
				Reporter:     validReporter,
				Address:      validAddr,
				TransactionHash: "",
				ActivityType: "structuring",
			},
			wantErr:   true,
			errString: "transaction_hash is required",
		},
		{
			name: "empty activity type",
			msg: &compliancepb.MsgReportSuspiciousActivity{
				Reporter:        validReporter,
				Address:         validAddr,
				TransactionHash: "hash123",
				ActivityType:    "",
			},
			wantErr:   true,
			errString: "activity_type is required",
		},
		{
			name: "description too long",
			msg: &compliancepb.MsgReportSuspiciousActivity{
				Reporter:        validReporter,
				Address:         validAddr,
				TransactionHash: "hash123",
				ActivityType:    "structuring",
				Description:     strings.Repeat("a", 1001),
			},
			wantErr:   true,
			errString: "description too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgScreenSanctions_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("test_address_123456")).String()

	tests := []struct {
		name      string
		msg       *compliancepb.MsgScreenSanctions
		wantErr   bool
		errString string
	}{
		{
			name: "valid message",
			msg: &compliancepb.MsgScreenSanctions{
				Address:      validAddr,
				ForceRefresh: false,
			},
			wantErr: false,
		},
		{
			name: "valid message with force refresh",
			msg: &compliancepb.MsgScreenSanctions{
				Address:      validAddr,
				ForceRefresh: true,
			},
			wantErr: false,
		},
		{
			name: "empty address",
			msg: &compliancepb.MsgScreenSanctions{
				Address: "",
			},
			wantErr:   true,
			errString: "address is required",
		},
		{
			name: "invalid address",
			msg: &compliancepb.MsgScreenSanctions{
				Address: "invalid-address",
			},
			wantErr:   true,
			errString: "invalid address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgRecordGDPRConsent_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("test_address_123456")).String()

	tests := []struct {
		name      string
		msg       *compliancepb.MsgRecordGDPRConsent
		wantErr   bool
		errString string
	}{
		{
			name: "valid message - consent given",
			msg: &compliancepb.MsgRecordGDPRConsent{
				Address:        validAddr,
				ConsentType:    "data_processing",
				Consented:      true,
				ConsentVersion: "v1.0",
			},
			wantErr: false,
		},
		{
			name: "valid message - consent withdrawn",
			msg: &compliancepb.MsgRecordGDPRConsent{
				Address:        validAddr,
				ConsentType:    "marketing",
				Consented:      false,
				ConsentVersion: "v1.0",
			},
			wantErr: false,
		},
		{
			name: "empty address",
			msg: &compliancepb.MsgRecordGDPRConsent{
				Address:        "",
				ConsentType:    "data_processing",
				Consented:      true,
				ConsentVersion: "v1.0",
			},
			wantErr:   true,
			errString: "address is required",
		},
		{
			name: "invalid address",
			msg: &compliancepb.MsgRecordGDPRConsent{
				Address:        "invalid-address",
				ConsentType:    "data_processing",
				Consented:      true,
				ConsentVersion: "v1.0",
			},
			wantErr:   true,
			errString: "invalid address",
		},
		{
			name: "empty consent type",
			msg: &compliancepb.MsgRecordGDPRConsent{
				Address:        validAddr,
				ConsentType:    "",
				Consented:      true,
				ConsentVersion: "v1.0",
			},
			wantErr:   true,
			errString: "consent_type is required",
		},
		{
			name: "empty consent version",
			msg: &compliancepb.MsgRecordGDPRConsent{
				Address:        validAddr,
				ConsentType:    "data_processing",
				Consented:      true,
				ConsentVersion: "",
			},
			wantErr:   true,
			errString: "consent_version is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgRequestGDPRData_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("test_address_123456")).String()

	tests := []struct {
		name      string
		msg       *compliancepb.MsgRequestGDPRData
		wantErr   bool
		errString string
	}{
		{
			name: "valid message - access",
			msg: &compliancepb.MsgRequestGDPRData{
				Address:     validAddr,
				RequestType: "access",
			},
			wantErr: false,
		},
		{
			name: "valid message - rectification",
			msg: &compliancepb.MsgRequestGDPRData{
				Address:     validAddr,
				RequestType: "rectification",
			},
			wantErr: false,
		},
		{
			name: "valid message - erasure",
			msg: &compliancepb.MsgRequestGDPRData{
				Address:     validAddr,
				RequestType: "erasure",
			},
			wantErr: false,
		},
		{
			name: "valid message - portability",
			msg: &compliancepb.MsgRequestGDPRData{
				Address:     validAddr,
				RequestType: "portability",
			},
			wantErr: false,
		},
		{
			name: "valid message - restriction",
			msg: &compliancepb.MsgRequestGDPRData{
				Address:     validAddr,
				RequestType: "restriction",
			},
			wantErr: false,
		},
		{
			name: "valid message - objection",
			msg: &compliancepb.MsgRequestGDPRData{
				Address:     validAddr,
				RequestType: "objection",
			},
			wantErr: false,
		},
		{
			name: "empty address",
			msg: &compliancepb.MsgRequestGDPRData{
				Address:     "",
				RequestType: "access",
			},
			wantErr:   true,
			errString: "address is required",
		},
		{
			name: "invalid address",
			msg: &compliancepb.MsgRequestGDPRData{
				Address:     "invalid-address",
				RequestType: "access",
			},
			wantErr:   true,
			errString: "invalid address",
		},
		{
			name: "empty request type",
			msg: &compliancepb.MsgRequestGDPRData{
				Address:     validAddr,
				RequestType: "",
			},
			wantErr:   true,
			errString: "request_type is required",
		},
		{
			name: "invalid request type",
			msg: &compliancepb.MsgRequestGDPRData{
				Address:     validAddr,
				RequestType: "invalid",
			},
			wantErr:   true,
			errString: "invalid request_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgEraseGDPRData_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("test_address_123456")).String()

	tests := []struct {
		name      string
		msg       *compliancepb.MsgEraseGDPRData
		wantErr   bool
		errString string
	}{
		{
			name: "valid message with reason",
			msg: &compliancepb.MsgEraseGDPRData{
				Address:       validAddr,
				ErasureReason: "no longer using service",
			},
			wantErr: false,
		},
		{
			name: "valid message without reason",
			msg: &compliancepb.MsgEraseGDPRData{
				Address:       validAddr,
				ErasureReason: "",
			},
			wantErr: false,
		},
		{
			name: "empty address",
			msg: &compliancepb.MsgEraseGDPRData{
				Address:       "",
				ErasureReason: "test",
			},
			wantErr:   true,
			errString: "address is required",
		},
		{
			name: "invalid address",
			msg: &compliancepb.MsgEraseGDPRData{
				Address:       "invalid-address",
				ErasureReason: "test",
			},
			wantErr:   true,
			errString: "invalid address",
		},
		{
			name: "erasure reason too long",
			msg: &compliancepb.MsgEraseGDPRData{
				Address:       validAddr,
				ErasureReason: strings.Repeat("a", 501),
			},
			wantErr:   true,
			errString: "erasure_reason too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgGenerateTaxReport_ValidateBasic(t *testing.T) {
	validAddr := sdk.AccAddress([]byte("test_address_123456")).String()

	tests := []struct {
		name      string
		msg       *compliancepb.MsgGenerateTaxReport
		wantErr   bool
		errString string
	}{
		{
			name: "valid message - US 1099-MISC",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      validAddr,
				TaxYear:      "2024",
				Jurisdiction: "US",
				ReportType:   "1099-MISC",
			},
			wantErr: false,
		},
		{
			name: "valid message - US 1099-K",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      validAddr,
				TaxYear:      "2024",
				Jurisdiction: "US",
				ReportType:   "1099-K",
			},
			wantErr: false,
		},
		{
			name: "valid message - generic",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      validAddr,
				TaxYear:      "2024",
				Jurisdiction: "GB",
				ReportType:   "generic",
			},
			wantErr: false,
		},
		{
			name: "empty address",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      "",
				TaxYear:      "2024",
				Jurisdiction: "US",
				ReportType:   "1099-MISC",
			},
			wantErr:   true,
			errString: "address is required",
		},
		{
			name: "invalid address",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      "invalid-address",
				TaxYear:      "2024",
				Jurisdiction: "US",
				ReportType:   "1099-MISC",
			},
			wantErr:   true,
			errString: "invalid address",
		},
		{
			name: "empty tax year",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      validAddr,
				TaxYear:      "",
				Jurisdiction: "US",
				ReportType:   "1099-MISC",
			},
			wantErr:   true,
			errString: "tax_year is required",
		},
		{
			name: "invalid tax year format",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      validAddr,
				TaxYear:      "24",
				Jurisdiction: "US",
				ReportType:   "1099-MISC",
			},
			wantErr:   true,
			errString: "tax_year must be 4 digits",
		},
		{
			name: "non-numeric tax year",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      validAddr,
				TaxYear:      "202a",
				Jurisdiction: "US",
				ReportType:   "1099-MISC",
			},
			wantErr:   true,
			errString: "tax_year must be numeric",
		},
		{
			name: "empty jurisdiction",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      validAddr,
				TaxYear:      "2024",
				Jurisdiction: "",
				ReportType:   "1099-MISC",
			},
			wantErr:   true,
			errString: "jurisdiction is required",
		},
		{
			name: "invalid jurisdiction length",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      validAddr,
				TaxYear:      "2024",
				Jurisdiction: "USA",
				ReportType:   "1099-MISC",
			},
			wantErr:   true,
			errString: "jurisdiction must be 2-letter",
		},
		{
			name: "empty report type",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      validAddr,
				TaxYear:      "2024",
				Jurisdiction: "US",
				ReportType:   "",
			},
			wantErr:   true,
			errString: "report_type is required",
		},
		{
			name: "invalid report type",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      validAddr,
				TaxYear:      "2024",
				Jurisdiction: "US",
				ReportType:   "invalid",
			},
			wantErr:   true,
			errString: "invalid report_type",
		},
		{
			name: "US jurisdiction with generic report type",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      validAddr,
				TaxYear:      "2024",
				Jurisdiction: "US",
				ReportType:   "generic",
			},
			wantErr:   true,
			errString: "US jurisdiction should use specific US tax forms",
		},
		{
			name: "non-US jurisdiction with US report type",
			msg: &compliancepb.MsgGenerateTaxReport{
				Address:      validAddr,
				TaxYear:      "2024",
				Jurisdiction: "GB",
				ReportType:   "1099-MISC",
			},
			wantErr:   true,
			errString: "should typically use 'generic' report type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
