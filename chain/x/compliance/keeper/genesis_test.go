package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/compliance/types"
	compliancepb "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
)

type GenesisTestSuite struct {
	KeeperTestSuite
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestInitGenesis_DefaultAndCustom() {
	ctx := suite.SdkCtx

	// Default genesis should initialize cleanly
	suite.Require().NoError(suite.Keeper.InitGenesis(ctx, types.DefaultGenesis()))

	// Custom populated genesis
	addr := suite.addr("kyc-genesis")
	params := types.DefaultParams()
	piiCommitment := make([]byte, 32)
	copy(piiCommitment, []byte("test_commitment_hash_32_bytes"))
	expiresAt := time.Now().Add(24 * time.Hour)
	genesis := &compliancepb.GenesisState{
		Params: params,
		KycRecords: []*compliancepb.KYCRecord{
			{
				Address:       addr,
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				PiiCommitment: piiCommitment,
				VerifiedAt:    time.Now(),
				ExpiresAt:     &expiresAt,
			},
		},
		AmlProfiles: []*compliancepb.AMLProfile{
			{
				Address:        addr,
				RiskLevel:      compliancepb.AMLRiskLevel_AML_RISK_LOW,
				LastAssessment: time.Now(),
				TotalVolume:    "0",
			},
		},
	}

	suite.Require().NoError(suite.Keeper.InitGenesis(ctx, genesis))

	exported := suite.Keeper.ExportGenesis(ctx)
	suite.NotNil(exported)
	suite.Equal(1, len(exported.KycRecords))
	suite.Equal(1, len(exported.AmlProfiles))
}

func (suite *GenesisTestSuite) TestInitGenesis_Invalid() {
	ctx := suite.SdkCtx

	// Nil genesis should error
	suite.Error(suite.Keeper.InitGenesis(ctx, nil))

	// Genesis with invalid params should error
	// KycExpiryDays cannot be zero when KycRequired is true
	invalidGenesis := &compliancepb.GenesisState{
		Params: types.ComplianceParams{
			KycRequired:   true, // KYC is required
			KycExpiryDays: 0,    // invalid: zero when KYC is required
		},
	}
	suite.Error(suite.Keeper.InitGenesis(ctx, invalidGenesis))
}

func (suite *GenesisTestSuite) TestDefaultGenesisValidation() {
	suite.NoError(types.ValidateGenesis(types.DefaultGenesis()))
}

func (suite *GenesisTestSuite) TestGenesisRoundTrip() {
	ctx := suite.SdkCtx

	// Create a comprehensive genesis state with various data
	addr1 := suite.addr("kyc-addr-1")
	addr2 := suite.addr("kyc-addr-2")
	addr3 := suite.addr("kyc-addr-3")

	piiCommitment1 := make([]byte, 32)
	copy(piiCommitment1, []byte("pii_commitment_1_32_bytes_long"))
	piiCommitment2 := make([]byte, 32)
	copy(piiCommitment2, []byte("pii_commitment_2_32_bytes_long"))

	expiresAt1 := time.Now().Add(24 * time.Hour)
	expiresAt2 := time.Now().Add(48 * time.Hour)

	genesis := &compliancepb.GenesisState{
		Params: types.ComplianceParams{
			KycRequired:                  true,
			KycExpiryDays:                365,
			TransactionMonitoringEnabled: true,
			SanctionsScreeningEnabled:    true,
			MinimumKycLevel:              compliancepb.KYCLevel_KYC_LEVEL_BASIC,
			StructuringThresholdCount:    3,
			VelocityLimit_24H:            "1000000",
			SanctionsLists:               []string{"OFAC", "UN"},
		},
		KycRecords: []*compliancepb.KYCRecord{
			{
				Address:       addr1,
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				PiiCommitment: piiCommitment1,
				VerifiedAt:    time.Now(),
				ExpiresAt:     &expiresAt1,
			},
			{
				Address:       addr2,
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_ADVANCED,
				PiiCommitment: piiCommitment2,
				VerifiedAt:    time.Now().Add(-10 * time.Hour),
				ExpiresAt:     &expiresAt2,
			},
		},
		AmlProfiles: []*compliancepb.AMLProfile{
			{
				Address:        addr1,
				RiskLevel:      compliancepb.AMLRiskLevel_AML_RISK_LOW,
				LastAssessment: time.Now(),
				TotalVolume:    "100000",
			},
			{
				Address:        addr2,
				RiskLevel:      compliancepb.AMLRiskLevel_AML_RISK_MEDIUM,
				LastAssessment: time.Now().Add(-5 * time.Hour),
				TotalVolume:    "500000",
			},
			{
				Address:        addr3,
				RiskLevel:      compliancepb.AMLRiskLevel_AML_RISK_HIGH,
				LastAssessment: time.Now().Add(-1 * time.Hour),
				TotalVolume:    "1000000",
			},
		},
	}

	// Import genesis
	err := suite.Keeper.InitGenesis(ctx, genesis)
	suite.NoError(err)

	// Export genesis (first export)
	exported1 := suite.Keeper.ExportGenesis(ctx)
	suite.NotNil(exported1)

	// Verify exported data matches original
	suite.Equal(len(genesis.KycRecords), len(exported1.KycRecords))
	suite.Equal(len(genesis.AmlProfiles), len(exported1.AmlProfiles))
	suite.Equal(genesis.Params.KycRequired, exported1.Params.KycRequired)
	suite.Equal(genesis.Params.KycExpiryDays, exported1.Params.KycExpiryDays)

	// Create a fresh keeper for re-import
	suite.SetupTest()
	ctx2 := suite.SdkCtx

	// Re-import the exported genesis
	err = suite.Keeper.InitGenesis(ctx2, exported1)
	suite.NoError(err)

	// Export again (second export)
	exported2 := suite.Keeper.ExportGenesis(ctx2)
	suite.NotNil(exported2)

	// The two exports should be identical
	suite.Equal(len(exported1.KycRecords), len(exported2.KycRecords))
	suite.Equal(len(exported1.AmlProfiles), len(exported2.AmlProfiles))
	suite.Equal(exported1.Params.KycRequired, exported2.Params.KycRequired)
	suite.Equal(exported1.Params.KycExpiryDays, exported2.Params.KycExpiryDays)
	suite.Equal(exported1.Params.TransactionMonitoringEnabled, exported2.Params.TransactionMonitoringEnabled)
	suite.Equal(exported1.Params.SanctionsScreeningEnabled, exported2.Params.SanctionsScreeningEnabled)

	// Verify individual records match
	for i := range exported1.KycRecords {
		suite.Equal(exported1.KycRecords[i].Address, exported2.KycRecords[i].Address)
		suite.Equal(exported1.KycRecords[i].KycLevel, exported2.KycRecords[i].KycLevel)
		suite.Equal(exported1.KycRecords[i].PiiCommitment, exported2.KycRecords[i].PiiCommitment)
	}

	for i := range exported1.AmlProfiles {
		suite.Equal(exported1.AmlProfiles[i].Address, exported2.AmlProfiles[i].Address)
		suite.Equal(exported1.AmlProfiles[i].RiskLevel, exported2.AmlProfiles[i].RiskLevel)
		suite.Equal(exported1.AmlProfiles[i].TotalVolume, exported2.AmlProfiles[i].TotalVolume)
	}
}

func (suite *GenesisTestSuite) addr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	addr, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	suite.Require().NoError(err)
	return addr
}
