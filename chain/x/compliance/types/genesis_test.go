package types_test

import (
	"testing"
	"time"

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
	now := time.Now()
	genesis := &types.GenesisState{
		Params: types.DefaultParams(),
		KycRecords: []*types.KYCRecord{
			{
				Address:       "cosmos1test",
				KycLevel:      types.KYCLevel_KYC_LEVEL_BASIC,
				Provider:      "test-provider",
				VerifiedAt:    now,
				PiiCommitment: make([]byte, 32),
				Jurisdiction:  "US",
			},
		},
		AmlProfiles: []*types.AMLProfile{
			{
				Address:           "cosmos1test",
				RiskLevel:         types.AMLRiskLevel_AML_RISK_LOW,
				TotalTransactions: 100,
				TotalVolume:       "1000000",
				LastAssessment:    now,
			},
		},
		SuspiciousActivities: []*types.SuspiciousActivity{
			{
				Id:              "activity1",
				Address:         "cosmos1test",
				ActivityType:    "structuring",
				Amount:          "9000",
				DetectedAt:      now,
				FiledSar:        false,
				TransactionHash: "hash123",
				Description:     "test",
			},
		},
		SanctionsResults: []*types.SanctionsScreeningResult{
			{
				Address:    "cosmos1test",
				Status:     types.SanctionsStatus_SANCTIONS_CLEAR,
				ScreenedAt: now,
			},
		},
		GdprConsents: []*types.GDPRConsent{
			{
				Address:        "cosmos1test",
				ConsentType:    "data_processing",
				Consented:      true,
				ConsentGivenAt: now,
				ConsentVersion: "v1.0",
			},
		},
		GdprRequests: []*types.GDPRDataRequest{
			{
				Id:          "req1",
				Address:     "cosmos1test",
				RequestType: "access",
				RequestedAt: now,
				Status:      "pending",
			},
		},
		TaxReports: []*types.TaxReport{
			{
				Id:                 "report1",
				Address:            "cosmos1test",
				TaxYear:            "2024",
				Jurisdiction:       "US",
				TotalIncome:        "100000",
				TotalCapitalGains:  "10000",
				TotalCapitalLosses: "5000",
				GeneratedAt:        now,
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
	// KycExpiryDays=0 is only invalid when KycRequired=true
	genesis.Params.KycRequired = true
	genesis.Params.KycExpiryDays = 0

	err := types.ValidateGenesis(genesis)
	require.Error(t, err)
}
