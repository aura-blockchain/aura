package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

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
	genesis := &compliancepb.GenesisState{
		Params: &params,
		KycRecords: []*compliancepb.KYCRecord{
			{
				Address:       addr,
				KycLevel:      compliancepb.KYCLevel_KYC_LEVEL_BASIC,
				PiiCommitment: piiCommitment,
				VerifiedAt:    timestamppb.Now(),
				ExpiresAt:     timestamppb.New(time.Now().Add(24 * time.Hour)),
			},
		},
		AmlProfiles: []*compliancepb.AMLProfile{
			{
				Address:        addr,
				RiskLevel:      compliancepb.AMLRiskLevel_AML_RISK_LOW,
				LastAssessment: timestamppb.Now(),
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

	suite.Error(suite.Keeper.InitGenesis(ctx, nil))

	// Nil params should error
	suite.Error(suite.Keeper.InitGenesis(ctx, &compliancepb.GenesisState{Params: nil}))
}

func (suite *GenesisTestSuite) TestDefaultGenesisValidation() {
	suite.NoError(types.ValidateGenesis(types.DefaultGenesis()))
}

func (suite *GenesisTestSuite) addr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	addr, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	suite.Require().NoError(err)
	return addr
}
