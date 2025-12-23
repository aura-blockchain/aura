package types_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

func TestDefaultGenesis(t *testing.T) {
	genesis := types.DefaultGenesis()

	require.NotNil(t, genesis)
	require.NotNil(t, genesis.Params)
	require.Empty(t, genesis.KycRecords)
	require.Empty(t, genesis.AmlProfiles)
	require.Empty(t, genesis.SuspiciousActivities)
	require.Empty(t, genesis.SanctionsResults)
	require.Empty(t, genesis.GdprConsents)
	require.Empty(t, genesis.GdprRequests)
	require.Empty(t, genesis.TaxReports)
}

func TestValidateGenesis_Valid(t *testing.T) {
	genesis := types.DefaultGenesis()

	err := types.ValidateGenesis(genesis)
	require.NoError(t, err)
}

func TestValidateGenesis_ValidWithData(t *testing.T) {
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		KycRecords: []*types.KYCRecord{
			{
				Address:      "cosmos1test",
				Status:       types.KYCStatus_APPROVED,
				KycLevel:     1,
				Jurisdiction: "US",
				Provider:     "test-provider",
				SubmittedAt:  time.Now(),
				ApprovedAt:   time.Now(),
			},
		},
		AmlProfiles: []*types.AMLProfile{
			{
				Address:           "cosmos1test",
				RiskLevel:         types.RiskLevel_LOW,
				TotalTransactions: 100,
				TotalVolume:       sdkmath.NewInt(1000000),
				LastAssessment:    time.Now(),
			},
		},
		SuspiciousActivities: []*types.SuspiciousActivity{
			{
				Id:              "activity1",
				Address:         "cosmos1test",
				ActivityType:    "structuring",
				Amount:          sdkmath.NewInt(9000),
				DetectedAt:      time.Now(),
				FiledSar:        false,
				TransactionHash: "hash123",
				Description:     "test",
			},
		},
		SanctionsResults: []*types.SanctionsScreeningResult{
			{
				Address:   "cosmos1test",
				Flagged:   false,
				Reason:    "clear",
				CheckedAt: time.Now(),
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
		},
		GdprConsents: []*types.GDPRConsent{
			{
				Address:        "cosmos1test",
				ConsentType:    "data_processing",
				Consented:      true,
				ConsentGivenAt: time.Now(),
				ConsentVersion: "v1.0",
			},
		},
		GdprRequests: []*types.GDPRRequest{
			{
				Id:          "req1",
				Address:     "cosmos1test",
				RequestType: types.GDPRRequestType_ACCESS,
				RequestedAt: time.Now(),
				Status:      types.GDPRStatus_PENDING,
			},
		},
		TaxReports: []*types.TaxReport{
			{
				Address:       "cosmos1test",
				TaxYear:       2024,
				Jurisdiction:  "US",
				TotalIncome:   sdkmath.NewInt(100000),
				TotalGains:    sdkmath.NewInt(10000),
				TotalLosses:   sdkmath.NewInt(5000),
				GeneratedAt:   time.Now(),
				ReportVersion: "v1.0",
			},
		},
	}

	err := types.ValidateGenesis(genesis)
	require.NoError(t, err)
}

func TestValidateGenesis_NilGenesis(t *testing.T) {
	err := types.ValidateGenesis(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "genesis state cannot be nil")
}

func TestValidateGenesis_InvalidParams(t *testing.T) {
	genesis := types.DefaultGenesis()
	genesis.Params.KycExpiryDays = 0 // Invalid

	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
}

func TestValidateGenesis_DuplicateKYCRecords(t *testing.T) {
	genesis := types.DefaultGenesis()
	record := &types.KYCRecord{
		Address:      "cosmos1duplicate",
		Status:       types.KYCStatus_APPROVED,
		KycLevel:     1,
		Jurisdiction: "US",
		Provider:     "test-provider",
		SubmittedAt:  time.Now(),
	}

	genesis.KycRecords = []*types.KYCRecord{record, record}

	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate KYC record")
}

func TestValidateGenesis_DuplicateAMLProfiles(t *testing.T) {
	genesis := types.DefaultGenesis()
	profile := &types.AMLProfile{
		Address:           "cosmos1duplicate",
		RiskLevel:         types.RiskLevel_LOW,
		TotalTransactions: 100,
		TotalVolume:       sdkmath.NewInt(1000000),
		LastAssessment:    time.Now(),
	}

	genesis.AmlProfiles = []*types.AMLProfile{profile, profile}

	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate AML profile")
}

func TestValidateGenesis_DuplicateSuspiciousActivities(t *testing.T) {
	genesis := types.DefaultGenesis()
	activity := &types.SuspiciousActivity{
		Id:          "duplicate",
		Address:     "cosmos1test",
		ActivityType: "test",
		Amount:      sdkmath.NewInt(1000),
		DetectedAt:  time.Now(),
	}

	genesis.SuspiciousActivities = []*types.SuspiciousActivity{activity, activity}

	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate suspicious activity")
}

func TestValidateGenesis_DuplicateSanctionsResults(t *testing.T) {
	genesis := types.DefaultGenesis()
	result := &types.SanctionsScreeningResult{
		Address:   "cosmos1duplicate",
		Flagged:   false,
		Reason:    "clear",
		CheckedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	genesis.SanctionsResults = []*types.SanctionsScreeningResult{result, result}

	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate sanctions")
}

func TestValidateGenesis_DuplicateGDPRRequests(t *testing.T) {
	genesis := types.DefaultGenesis()
	request := &types.GDPRRequest{
		Id:          "duplicate",
		Address:     "cosmos1test",
		RequestType: types.GDPRRequestType_ACCESS,
		RequestedAt: time.Now(),
		Status:      types.GDPRStatus_PENDING,
	}

	genesis.GdprRequests = []*types.GDPRRequest{request, request}

	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate GDPR request")
}

func TestValidateGenesis_EmptyKYCAddress(t *testing.T) {
	genesis := types.DefaultGenesis()
	genesis.KycRecords = []*types.KYCRecord{
		{
			Address:      "",
			Status:       types.KYCStatus_APPROVED,
			KycLevel:     1,
			Jurisdiction: "US",
			Provider:     "test-provider",
			SubmittedAt:  time.Now(),
		},
	}

	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "address cannot be empty")
}

func TestValidateGenesis_EmptyAMLAddress(t *testing.T) {
	genesis := types.DefaultGenesis()
	genesis.AmlProfiles = []*types.AMLProfile{
		{
			Address:           "",
			RiskLevel:         types.RiskLevel_LOW,
			TotalTransactions: 100,
			TotalVolume:       sdkmath.NewInt(1000000),
			LastAssessment:    time.Now(),
		},
	}

	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "address cannot be empty")
}

func TestValidateGenesis_EmptySuspiciousActivityID(t *testing.T) {
	genesis := types.DefaultGenesis()
	genesis.SuspiciousActivities = []*types.SuspiciousActivity{
		{
			Id:          "",
			Address:     "cosmos1test",
			ActivityType: "test",
			Amount:      sdkmath.NewInt(1000),
			DetectedAt:  time.Now(),
		},
	}

	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "id cannot be empty")
}

func TestValidateGenesis_EmptyGDPRRequestID(t *testing.T) {
	genesis := types.DefaultGenesis()
	genesis.GdprRequests = []*types.GDPRRequest{
		{
			Id:          "",
			Address:     "cosmos1test",
			RequestType: types.GDPRRequestType_ACCESS,
			RequestedAt: time.Now(),
			Status:      types.GDPRStatus_PENDING,
		},
	}

	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
	require.Contains(t, err.Error(), "id cannot be empty")
}
