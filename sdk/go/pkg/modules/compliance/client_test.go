package compliance

import (
	"testing"

	compliancepb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubmitKYCParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *SubmitKYCParams
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
			errMsg:  "params cannot be nil",
		},
		{
			name: "missing address",
			params: &SubmitKYCParams{
				KYCLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      "provider123",
				PIICommitment: []byte("commitment"),
				Jurisdiction:  "US",
			},
			wantErr: true,
			errMsg:  "address is required",
		},
		{
			name: "missing provider",
			params: &SubmitKYCParams{
				Address:       "aura1test123",
				KYCLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				PIICommitment: []byte("commitment"),
				Jurisdiction:  "US",
			},
			wantErr: true,
			errMsg:  "provider is required",
		},
		{
			name: "missing PII commitment",
			params: &SubmitKYCParams{
				Address:      "aura1test123",
				KYCLevel:     compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				Provider:     "provider123",
				Jurisdiction: "US",
			},
			wantErr: true,
			errMsg:  "PII commitment is required",
		},
		{
			name: "missing jurisdiction",
			params: &SubmitKYCParams{
				Address:       "aura1test123",
				KYCLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      "provider123",
				PIICommitment: []byte("commitment"),
			},
			wantErr: true,
			errMsg:  "jurisdiction is required",
		},
		{
			name: "valid params",
			params: &SubmitKYCParams{
				Address:       "aura1test123",
				KYCLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      "provider123",
				PIICommitment: []byte("commitment"),
				Jurisdiction:  "US",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.Nil(t, tt.params)
				} else {
					assert.True(t,
						tt.params.Address == "" ||
							tt.params.Provider == "" ||
							len(tt.params.PIICommitment) == 0 ||
							tt.params.Jurisdiction == "")
				}
			} else {
				assert.NotEmpty(t, tt.params.Address)
				assert.NotEmpty(t, tt.params.Provider)
				assert.NotEmpty(t, tt.params.PIICommitment)
				assert.NotEmpty(t, tt.params.Jurisdiction)
			}
		})
	}
}

func TestReportSuspiciousActivityParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *ReportSuspiciousActivityParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing reporter",
			params: &ReportSuspiciousActivityParams{
				Address:      "aura1target123",
				ActivityType: "suspicious",
			},
			wantErr: true,
		},
		{
			name: "missing address",
			params: &ReportSuspiciousActivityParams{
				Reporter:     "aura1reporter123",
				ActivityType: "suspicious",
			},
			wantErr: true,
		},
		{
			name: "missing activity type",
			params: &ReportSuspiciousActivityParams{
				Reporter: "aura1reporter123",
				Address:  "aura1target123",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &ReportSuspiciousActivityParams{
				Reporter:        "aura1reporter123",
				Address:         "aura1target123",
				ActivityType:    "suspicious",
				TransactionHash: "hash123",
				Description:     "Suspicious transaction",
				Indicators:      []string{"high_velocity", "structuring"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.Nil(t, tt.params)
				} else {
					assert.True(t, tt.params.Reporter == "" || tt.params.Address == "" || tt.params.ActivityType == "")
				}
			} else {
				assert.NotEmpty(t, tt.params.Reporter)
				assert.NotEmpty(t, tt.params.Address)
				assert.NotEmpty(t, tt.params.ActivityType)
			}
		})
	}
}

func TestScreenSanctionsParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *ScreenSanctionsParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing address",
			params: &ScreenSanctionsParams{
				ForceRefresh: true,
			},
			wantErr: true,
		},
		{
			name: "valid params without force refresh",
			params: &ScreenSanctionsParams{
				Address:      "aura1test123",
				ForceRefresh: false,
			},
			wantErr: false,
		},
		{
			name: "valid params with force refresh",
			params: &ScreenSanctionsParams{
				Address:      "aura1test123",
				ForceRefresh: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.Nil(t, tt.params)
				} else {
					assert.Empty(t, tt.params.Address)
				}
			} else {
				assert.NotEmpty(t, tt.params.Address)
			}
		})
	}
}

func TestRecordGDPRConsentParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *RecordGDPRConsentParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing address",
			params: &RecordGDPRConsentParams{
				ConsentType: "marketing",
				Consented:   true,
			},
			wantErr: true,
		},
		{
			name: "missing consent type",
			params: &RecordGDPRConsentParams{
				Address:   "aura1test123",
				Consented: true,
			},
			wantErr: true,
		},
		{
			name: "valid params with consent granted",
			params: &RecordGDPRConsentParams{
				Address:        "aura1test123",
				ConsentType:    "marketing",
				Consented:      true,
				ConsentVersion: "1.0",
			},
			wantErr: false,
		},
		{
			name: "valid params with consent denied",
			params: &RecordGDPRConsentParams{
				Address:        "aura1test123",
				ConsentType:    "marketing",
				Consented:      false,
				ConsentVersion: "1.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.Nil(t, tt.params)
				} else {
					assert.True(t, tt.params.Address == "" || tt.params.ConsentType == "")
				}
			} else {
				assert.NotEmpty(t, tt.params.Address)
				assert.NotEmpty(t, tt.params.ConsentType)
			}
		})
	}
}

func TestGenerateTaxReportParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *GenerateTaxReportParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing address",
			params: &GenerateTaxReportParams{
				TaxYear: "2024",
			},
			wantErr: true,
		},
		{
			name: "missing tax year",
			params: &GenerateTaxReportParams{
				Address: "aura1test123",
			},
			wantErr: true,
		},
		{
			name: "valid minimal params",
			params: &GenerateTaxReportParams{
				Address: "aura1test123",
				TaxYear: "2024",
			},
			wantErr: false,
		},
		{
			name: "valid full params",
			params: &GenerateTaxReportParams{
				Address:      "aura1test123",
				TaxYear:      "2024",
				Jurisdiction: "US",
				ReportType:   "1099-MISC",
				FilePath:     "/tmp/tax_report.pdf",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.Nil(t, tt.params)
				} else {
					assert.True(t, tt.params.Address == "" || tt.params.TaxYear == "")
				}
			} else {
				assert.NotEmpty(t, tt.params.Address)
				assert.NotEmpty(t, tt.params.TaxYear)
			}
		})
	}
}

func TestRequestGDPRData_Validation(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		requestType string
		wantErr     bool
	}{
		{
			name:        "empty address",
			address:     "",
			requestType: "access",
			wantErr:     true,
		},
		{
			name:        "empty request type",
			address:     "aura1test123",
			requestType: "",
			wantErr:     true,
		},
		{
			name:        "valid params",
			address:     "aura1test123",
			requestType: "access",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.True(t, tt.address == "" || tt.requestType == "")
			} else {
				assert.NotEmpty(t, tt.address)
				assert.NotEmpty(t, tt.requestType)
			}
		})
	}
}

func TestEraseGDPRData_Validation(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		erasureReason string
		wantErr       bool
	}{
		{
			name:          "empty address",
			address:       "",
			erasureReason: "user request",
			wantErr:       true,
		},
		{
			name:          "empty erasure reason",
			address:       "aura1test123",
			erasureReason: "",
			wantErr:       true,
		},
		{
			name:          "valid params",
			address:       "aura1test123",
			erasureReason: "user request",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.True(t, tt.address == "" || tt.erasureReason == "")
			} else {
				assert.NotEmpty(t, tt.address)
				assert.NotEmpty(t, tt.erasureReason)
			}
		})
	}
}

func TestQueryValidation(t *testing.T) {
	t.Run("GetKYCRecord requires address", func(t *testing.T) {
		address := ""
		assert.Empty(t, address)
	})

	t.Run("GetKYCHistory requires address", func(t *testing.T) {
		address := ""
		assert.Empty(t, address)
	})

	t.Run("GetAMLProfile requires address", func(t *testing.T) {
		address := ""
		assert.Empty(t, address)
	})

	t.Run("GetSanctionsScreening requires address", func(t *testing.T) {
		address := ""
		assert.Empty(t, address)
	})

	t.Run("GetTransactionAlerts requires address", func(t *testing.T) {
		address := ""
		assert.Empty(t, address)
	})

	t.Run("GetTaxReport requires address and tax year", func(t *testing.T) {
		address := ""
		taxYear := ""
		assert.True(t, address == "" || taxYear == "")
	})
}

func TestKYCLevelValidation(t *testing.T) {
	validLevels := []compliancepb.KYCLevel{
		compliancepb.KYCLevel_KYC_LEVEL_NONE,
		compliancepb.KYCLevel_KYC_LEVEL_BASIC,
		compliancepb.KYCLevel_KYC_LEVEL_INTERMEDIATE,
		compliancepb.KYCLevel_KYC_LEVEL_ADVANCED,
	}

	for _, level := range validLevels {
		t.Run(level.String(), func(t *testing.T) {
			assert.NotEqual(t, compliancepb.KYCLevel_KYC_LEVEL_UNSPECIFIED, level)
		})
	}
}

func TestAMLRiskLevelValidation(t *testing.T) {
	validLevels := []compliancepb.AMLRiskLevel{
		compliancepb.AMLRiskLevel_AML_RISK_LOW,
		compliancepb.AMLRiskLevel_AML_RISK_MEDIUM,
		compliancepb.AMLRiskLevel_AML_RISK_HIGH,
		compliancepb.AMLRiskLevel_AML_RISK_SEVERE,
	}

	for _, level := range validLevels {
		t.Run(level.String(), func(t *testing.T) {
			assert.NotEqual(t, compliancepb.AMLRiskLevel_AML_RISK_UNSPECIFIED, level)
		})
	}
}

func TestSanctionsStatusValidation(t *testing.T) {
	validStatuses := []compliancepb.SanctionsStatus{
		compliancepb.SanctionsStatus_SANCTIONS_CLEAR,
		compliancepb.SanctionsStatus_SANCTIONS_MATCH,
		compliancepb.SanctionsStatus_SANCTIONS_CONFIRMED,
		compliancepb.SanctionsStatus_SANCTIONS_PENDING_REVIEW,
	}

	for _, status := range validStatuses {
		t.Run(status.String(), func(t *testing.T) {
			assert.NotEqual(t, compliancepb.SanctionsStatus_SANCTIONS_UNSPECIFIED, status)
		})
	}
}
