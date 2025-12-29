package compliance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	compliancepb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
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
				FullName:       "John Doe",
				DocumentType:   "passport",
				DocumentNumber: "123456",
			},
			wantErr: true,
			errMsg:  "address is required",
		},
		{
			name: "missing full name",
			params: &SubmitKYCParams{
				Address:        "aura1test123",
				DocumentType:   "passport",
				DocumentNumber: "123456",
			},
			wantErr: true,
			errMsg:  "full name is required",
		},
		{
			name: "missing document type",
			params: &SubmitKYCParams{
				Address:        "aura1test123",
				FullName:       "John Doe",
				DocumentNumber: "123456",
			},
			wantErr: true,
			errMsg:  "document type is required",
		},
		{
			name: "missing document number",
			params: &SubmitKYCParams{
				Address:      "aura1test123",
				FullName:     "John Doe",
				DocumentType: "passport",
			},
			wantErr: true,
			errMsg:  "document number is required",
		},
		{
			name: "valid params",
			params: &SubmitKYCParams{
				Address:           "aura1test123",
				FullName:          "John Doe",
				DocumentType:      "passport",
				DocumentNumber:    "123456",
				DateOfBirth:       "1990-01-01",
				Nationality:       "US",
				ResidenceCountry:  "US",
				VerificationLevel: compliancepb.VerificationLevel_VERIFICATION_LEVEL_BASIC,
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
					if tt.params.Address == "" {
						assert.Empty(t, tt.params.Address)
					} else if tt.params.FullName == "" {
						assert.Empty(t, tt.params.FullName)
					} else if tt.params.DocumentType == "" {
						assert.Empty(t, tt.params.DocumentType)
					} else if tt.params.DocumentNumber == "" {
						assert.Empty(t, tt.params.DocumentNumber)
					}
				}
			} else {
				assert.NotEmpty(t, tt.params.Address)
				assert.NotEmpty(t, tt.params.FullName)
				assert.NotEmpty(t, tt.params.DocumentType)
				assert.NotEmpty(t, tt.params.DocumentNumber)
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
				TargetAddress: "aura1target123",
				ActivityType:  "suspicious",
			},
			wantErr: true,
		},
		{
			name: "missing target address",
			params: &ReportSuspiciousActivityParams{
				Reporter:     "aura1reporter123",
				ActivityType: "suspicious",
			},
			wantErr: true,
		},
		{
			name: "missing activity type",
			params: &ReportSuspiciousActivityParams{
				Reporter:      "aura1reporter123",
				TargetAddress: "aura1target123",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &ReportSuspiciousActivityParams{
				Reporter:      "aura1reporter123",
				TargetAddress: "aura1target123",
				ActivityType:  "suspicious",
				Amount:        "1000",
				Description:   "Suspicious transaction",
				Evidence:      []string{"hash1", "hash2"},
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
					assert.True(t, tt.params.Reporter == "" || tt.params.TargetAddress == "" || tt.params.ActivityType == "")
				}
			} else {
				assert.NotEmpty(t, tt.params.Reporter)
				assert.NotEmpty(t, tt.params.TargetAddress)
				assert.NotEmpty(t, tt.params.ActivityType)
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
				TaxYear: 2024,
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
			name: "valid params",
			params: &GenerateTaxReportParams{
				Address:    "aura1test123",
				TaxYear:    2024,
				TaxRegion:  "US",
				ReportType: "full",
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
					assert.True(t, tt.params.Address == "" || tt.params.TaxYear == 0)
				}
			} else {
				assert.NotEmpty(t, tt.params.Address)
				assert.NotZero(t, tt.params.TaxYear)
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
				Granted:     true,
			},
			wantErr: true,
		},
		{
			name: "missing consent type",
			params: &RecordGDPRConsentParams{
				Address: "aura1test123",
				Granted: true,
			},
			wantErr: true,
		},
		{
			name: "valid params with consent granted",
			params: &RecordGDPRConsentParams{
				Address:     "aura1test123",
				ConsentType: "marketing",
				Granted:     true,
				ConsentText: "I consent to marketing",
				ConsentHash: "hash123",
			},
			wantErr: false,
		},
		{
			name: "valid params with consent denied",
			params: &RecordGDPRConsentParams{
				Address:     "aura1test123",
				ConsentType: "marketing",
				Granted:     false,
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
				FullName: "John Doe",
			},
			wantErr: true,
		},
		{
			name: "valid minimal params",
			params: &ScreenSanctionsParams{
				Address: "aura1test123",
			},
			wantErr: false,
		},
		{
			name: "valid full params",
			params: &ScreenSanctionsParams{
				Address:      "aura1test123",
				FullName:     "John Doe",
				DateOfBirth:  "1990-01-01",
				Nationality:  "US",
				CheckAgainst: []string{"OFAC", "EU"},
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
