package keeper_test

import (
	"testing"

	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	"github.com/stretchr/testify/require"
)

func (suite *KeeperTestSuite) TestValidateContractExecution() {
	contractAddr := "cosmos1contract"
	userAddr := "cosmos1user"
	admin := "cosmos1admin"

	// Register contract with various restrictions
	info := &pb.ContractInfo{
		Address: contractAddr,
		CodeId:  1,
		Creator: admin,
		Admin:   admin,
		Label:   "test-contract",
		Metadata: &pb.ContractMetadata{
			Name:               "Test Contract",
			RequiresVc:         false,
			MinConfidenceScore: 0,
		},
		SecurityPolicy: &pb.SecurityPolicy{
			AllowPause:           true, // Enable pause functionality
			MaxGasPerTx:          1000000,
			RateLimitPerUser:     10,
			BlacklistedAddresses: []string{},
			WhitelistedAddresses: []string{},
		},
		Compliance: &pb.ComplianceRequirements{
			EnforceKyc:            false,
			EnforceSanctionsCheck: false,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Test 1: Valid execution
	err := suite.keeper.ValidateContractExecution(suite.ctx, contractAddr, userAddr, 500000)
	suite.NoError(err)

	// Test 2: Contract not found
	err = suite.keeper.ValidateContractExecution(suite.ctx, "cosmos1nonexistent", userAddr, 500000)
	suite.ErrorIs(err, types.ErrContractNotFound)

	// Test 3: Contract paused
	suite.NoError(suite.keeper.PauseContract(suite.ctx, contractAddr, admin, "test"))
	err = suite.keeper.ValidateContractExecution(suite.ctx, contractAddr, userAddr, 500000)
	suite.ErrorIs(err, types.ErrContractPaused)
	suite.NoError(suite.keeper.UnpauseContract(suite.ctx, contractAddr, admin))

	// Test 4: Gas limit exceeded
	err = suite.keeper.ValidateContractExecution(suite.ctx, contractAddr, userAddr, 2000000)
	suite.ErrorIs(err, types.ErrGasLimitExceeded)

	// Test 5: Rate limit exceeded
	for i := 0; i < 10; i++ {
		suite.keeper.IncrementRateLimit(suite.ctx, contractAddr, userAddr)
	}
	err = suite.keeper.ValidateContractExecution(suite.ctx, contractAddr, userAddr, 500000)
	suite.ErrorIs(err, types.ErrRateLimitExceeded)
}

func (suite *KeeperTestSuite) TestValidateBlacklist() {
	contractAddr := "cosmos1contract"
	blacklistedUser := "cosmos1blacklisted"
	normalUser := "cosmos1normal"
	admin := "cosmos1admin"

	// Register contract with blacklist
	info := &pb.ContractInfo{
		Address: contractAddr,
		CodeId:  1,
		Creator: admin,
		Admin:   admin,
		Label:   "test-contract",
		Metadata: &pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: &pb.SecurityPolicy{
			BlacklistedAddresses: []string{blacklistedUser},
		},
		Compliance: &pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Blacklisted user should fail
	err := suite.keeper.ValidateContractExecution(suite.ctx, contractAddr, blacklistedUser, 100000)
	suite.ErrorIs(err, types.ErrBlacklisted)

	// Normal user should pass
	err = suite.keeper.ValidateContractExecution(suite.ctx, contractAddr, normalUser, 100000)
	suite.NoError(err)
}

func (suite *KeeperTestSuite) TestValidateWhitelist() {
	contractAddr := "cosmos1contract"
	whitelistedUser := "cosmos1whitelisted"
	normalUser := "cosmos1normal"
	admin := "cosmos1admin"

	// Register contract with whitelist
	info := &pb.ContractInfo{
		Address: contractAddr,
		CodeId:  1,
		Creator: admin,
		Admin:   admin,
		Label:   "test-contract",
		Metadata: &pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: &pb.SecurityPolicy{
			WhitelistedAddresses: []string{whitelistedUser},
		},
		Compliance: &pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Whitelisted user should pass
	err := suite.keeper.ValidateContractExecution(suite.ctx, contractAddr, whitelistedUser, 100000)
	suite.NoError(err)

	// Non-whitelisted user should fail
	err = suite.keeper.ValidateContractExecution(suite.ctx, contractAddr, normalUser, 100000)
	suite.ErrorIs(err, types.ErrNotWhitelisted)
}

func TestValidateMetadata(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	// Valid metadata
	validMetadata := &pb.ContractMetadata{
		Name:               "Valid Contract",
		Description:        "A valid contract",
		RequiredVcTypes:    []string{"kyc", "identity"},
		MinConfidenceScore: 50,
		RequiredKycLevel:   2,
	}
	err := suite.keeper.ValidateMetadataUpdate(validMetadata)
	require.NoError(t, err)

	// Invalid: empty name
	invalidMetadata := &pb.ContractMetadata{
		Name:        "",
		Description: "No name",
	}
	err = suite.keeper.ValidateMetadataUpdate(invalidMetadata)
	require.ErrorIs(t, err, types.ErrInvalidMetadata)

	// Invalid: CS too high
	invalidCS := &pb.ContractMetadata{
		Name:               "Test",
		MinConfidenceScore: 200,
	}
	err = suite.keeper.ValidateMetadataUpdate(invalidCS)
	require.ErrorIs(t, err, types.ErrInvalidMetadata)

	// Invalid: KYC level too high
	invalidKYC := &pb.ContractMetadata{
		Name:             "Test",
		RequiredKycLevel: 5,
	}
	err = suite.keeper.ValidateMetadataUpdate(invalidKYC)
	require.ErrorIs(t, err, types.ErrInvalidMetadata)

	// Invalid: empty VC type
	invalidVC := &pb.ContractMetadata{
		Name:            "Test",
		RequiredVcTypes: []string{"kyc", ""},
	}
	err = suite.keeper.ValidateMetadataUpdate(invalidVC)
	require.ErrorIs(t, err, types.ErrInvalidMetadata)
}

func TestValidateSecurityPolicy(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	// Valid policy
	validPolicy := &pb.SecurityPolicy{
		MaxGasPerTx:          1000000,
		RateLimitPerUser:     100,
		BlacklistedAddresses: []string{"cosmos1bad"},
		WhitelistedAddresses: []string{"cosmos1good"},
	}
	err := suite.keeper.ValidateSecurityPolicyUpdate(validPolicy)
	require.NoError(t, err)

	// Invalid: gas too high
	invalidGas := &pb.SecurityPolicy{
		MaxGasPerTx: 100000000,
	}
	err = suite.keeper.ValidateSecurityPolicyUpdate(invalidGas)
	require.ErrorIs(t, err, types.ErrInvalidSecurityPolicy)

	// Invalid: rate limit too high
	invalidRate := &pb.SecurityPolicy{
		RateLimitPerUser: 20000,
	}
	err = suite.keeper.ValidateSecurityPolicyUpdate(invalidRate)
	require.ErrorIs(t, err, types.ErrInvalidSecurityPolicy)

	// Invalid: empty blacklist address
	invalidBlacklist := &pb.SecurityPolicy{
		BlacklistedAddresses: []string{"cosmos1addr", ""},
	}
	err = suite.keeper.ValidateSecurityPolicyUpdate(invalidBlacklist)
	require.ErrorIs(t, err, types.ErrInvalidSecurityPolicy)
}

func TestValidateContractRegistration(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()

	// Valid registration
	validMsg := &types.MsgRegisterContract{
		Creator:         "cosmos1signer",
		ContractAddress: "cosmos1contract",
		CodeId:          1,
		Admin:           "cosmos1signer",
		Metadata: &types.ContractMetadata{
			Name:        "Test",
			Description: "A test contract",
		},
		SecurityPolicy: &types.SecurityPolicy{
			MaxGasPerExecution: 1000000,
			RateLimitPerUser:   100,
		},
	}
	err := suite.keeper.ValidateContractRegistration(suite.ctx, validMsg)
	require.NoError(t, err)

	// Invalid: empty contract address
	invalidAddr := &types.MsgRegisterContract{
		Creator:         "cosmos1signer",
		ContractAddress: "",
		CodeId:          1,
		Admin:           "cosmos1signer",
		Metadata:        &types.ContractMetadata{Name: "Test", Description: "Test"},
		SecurityPolicy:  &types.SecurityPolicy{},
	}
	err = suite.keeper.ValidateContractRegistration(suite.ctx, invalidAddr)
	require.ErrorIs(t, err, types.ErrInvalidContractAddress)

	// Invalid: zero code ID
	invalidCode := &types.MsgRegisterContract{
		Creator:         "cosmos1signer",
		ContractAddress: "cosmos1contract",
		CodeId:          0,
		Admin:           "cosmos1signer",
		Metadata:        &types.ContractMetadata{Name: "Test", Description: "Test"},
		SecurityPolicy:  &types.SecurityPolicy{},
	}
	err = suite.keeper.ValidateContractRegistration(suite.ctx, invalidCode)
	require.ErrorIs(t, err, types.ErrInvalidCodeID)

	// Test empty metadata
	invalidMetadata := &types.MsgRegisterContract{
		Creator:         "cosmos1signer",
		ContractAddress: "cosmos1contract",
		CodeId:          1,
		Admin:           "cosmos1signer",
		Metadata:        &types.ContractMetadata{}, // Empty metadata
		SecurityPolicy:  &types.SecurityPolicy{},
	}
	err = suite.keeper.ValidateContractRegistration(suite.ctx, invalidMetadata)
	require.ErrorIs(t, err, types.ErrInvalidMetadata)
}
