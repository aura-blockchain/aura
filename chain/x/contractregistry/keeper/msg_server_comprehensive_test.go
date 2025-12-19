package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/contractregistry/keeper"
	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MsgServerComprehensiveTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	keeper    *keeper.Keeper
	msgServer pb.MsgServer
	cdc       codec.BinaryCodec
}

func TestMsgServerComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerComprehensiveTestSuite))
}

func (suite *MsgServerComprehensiveTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx

	registry := codectypes.NewInterfaceRegistry()
	suite.cdc = codec.NewProtoCodec(registry)

	suite.keeper = keeper.NewKeeper(
		key,
		suite.cdc,
		"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
	)

	// Initialize with default params
	params := types.DefaultParams()
	suite.NoError(suite.keeper.SetParams(suite.ctx, &params))

	suite.msgServer = keeper.NewMsgServerImpl(*suite.keeper)
}

// ============================
// RegisterContract Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestRegisterContract_Success() {
	msg := &pb.MsgRegisterContract{
		Signer:          "cosmos1creator",
		ContractAddress: "cosmos1contract",
		CodeId:          1,
		Creator:         "cosmos1creator",
		Admin:           "cosmos1creator",
		Label:           "test-contract",
		Metadata: pb.ContractMetadata{
			Name:        "Test Contract",
			Description: "A test smart contract",
			Version:     "1.0.0",
			Tags:        []string{"test", "demo"},
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause:       true,
			MaxGasPerTx:      1000000,
			RateLimitPerUser: 100,
		},
		Compliance: pb.ComplianceRequirements{
			EnforceKyc:            false,
			MinKycLevel:           0,
			EnforceSanctionsCheck: false,
		},
	}

	resp, err := suite.msgServer.RegisterContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.NoError(err)
	suite.NotNil(resp)
	suite.True(resp.Success)
	suite.Equal(msg.ContractAddress, resp.ContractAddress)

	// Verify contract was stored
	info, found := suite.keeper.GetContractInfo(suite.ctx, msg.ContractAddress)
	suite.True(found)
	suite.Equal(msg.ContractAddress, info.Address)
	suite.Equal(msg.Creator, info.Creator)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_ACTIVE, info.Status)
}

func (suite *MsgServerComprehensiveTestSuite) TestRegisterContract_AdminDifferentFromCreator() {
	msg := &pb.MsgRegisterContract{
		Signer:          "cosmos1creator",
		ContractAddress: "cosmos1contract",
		CodeId:          1,
		Creator:         "cosmos1creator",
		Admin:           "cosmos1admin",
		Label:           "test-contract",
		Metadata: pb.ContractMetadata{
			Name:        "Test Contract",
			Description: "Test",
		},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance:     pb.ComplianceRequirements{},
	}

	resp, err := suite.msgServer.RegisterContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.NoError(err)
	suite.True(resp.Success)

	// Verify admin is set correctly
	info, _ := suite.keeper.GetContractInfo(suite.ctx, msg.ContractAddress)
	suite.Equal("cosmos1admin", info.Admin)
}

func (suite *MsgServerComprehensiveTestSuite) TestRegisterContract_AlreadyExists() {
	// Register first time
	msg := &pb.MsgRegisterContract{
		Signer:          "cosmos1creator",
		ContractAddress: "cosmos1contract",
		CodeId:          1,
		Creator:         "cosmos1creator",
		Admin:           "cosmos1creator",
		Label:           "test-contract",
		Metadata: pb.ContractMetadata{
			Name: "Test",
		},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance:     pb.ComplianceRequirements{},
	}

	_, err := suite.msgServer.RegisterContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.NoError(err)

	// Try to register again
	_, err = suite.msgServer.RegisterContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Error(err)
	suite.ErrorIs(err, types.ErrContractAlreadyExists)
}

func (suite *MsgServerComprehensiveTestSuite) TestRegisterContract_UnauthorizedSigner() {
	msg := &pb.MsgRegisterContract{
		Signer:          "cosmos1attacker",
		ContractAddress: "cosmos1contract",
		CodeId:          1,
		Creator:         "cosmos1creator",
		Admin:           "cosmos1admin",
		Label:           "test-contract",
		Metadata:        pb.ContractMetadata{Name: "Test"},
		SecurityPolicy:  pb.SecurityPolicy{},
		Compliance:      pb.ComplianceRequirements{},
	}

	_, err := suite.msgServer.RegisterContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Error(err)
	suite.Equal(codes.PermissionDenied, status.Code(err))
}

func (suite *MsgServerComprehensiveTestSuite) TestRegisterContract_ExceedMaxContracts() {
	// Set low limit
	params := suite.keeper.GetParams(suite.ctx)
	params.MaxContractsPerCreator = 1
	suite.NoError(suite.keeper.SetParams(suite.ctx, params))

	creator := "cosmos1creator"

	// Register first contract
	msg1 := &pb.MsgRegisterContract{
		Signer:          creator,
		ContractAddress: "cosmos1contract1",
		CodeId:          1,
		Creator:         creator,
		Admin:           creator,
		Label:           "contract1",
		Metadata:        pb.ContractMetadata{Name: "Contract 1"},
		SecurityPolicy:  pb.SecurityPolicy{},
		Compliance:      pb.ComplianceRequirements{},
	}
	_, err := suite.msgServer.RegisterContract(sdk.WrapSDKContext(suite.ctx), msg1)
	suite.NoError(err)

	// Try to register second contract
	msg2 := &pb.MsgRegisterContract{
		Signer:          creator,
		ContractAddress: "cosmos1contract2",
		CodeId:          2,
		Creator:         creator,
		Admin:           creator,
		Label:           "contract2",
		Metadata:        pb.ContractMetadata{Name: "Contract 2"},
		SecurityPolicy:  pb.SecurityPolicy{},
		Compliance:      pb.ComplianceRequirements{},
	}
	_, err = suite.msgServer.RegisterContract(sdk.WrapSDKContext(suite.ctx), msg2)
	suite.Error(err)
	suite.ErrorIs(err, types.ErrTooManyContracts)
}

// ============================
// UpdateContractMetadata Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestUpdateContractMetadata_Success() {
	// Register contract first
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        "cosmos1creator",
		Admin:          admin,
		Label:          "test",
		Metadata:       pb.ContractMetadata{Name: "Original"},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance:     pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Update metadata
	msg := &pb.MsgUpdateContractMetadata{
		Signer:          admin,
		ContractAddress: contractAddr,
		Metadata: pb.ContractMetadata{
			Name:        "Updated Contract",
			Description: "Updated description",
			Version:     "2.0.0",
			Tags:        []string{"updated"},
		},
	}

	resp, err := suite.msgServer.UpdateContractMetadata(sdk.WrapSDKContext(suite.ctx), msg)
	suite.NoError(err)
	suite.True(resp.Success)

	// Verify update
	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal("Updated Contract", stored.Metadata.Name)
	suite.Equal("2.0.0", stored.Metadata.Version)
}

func (suite *MsgServerComprehensiveTestSuite) TestUpdateContractMetadata_NotAdmin() {
	// Register contract
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        "cosmos1creator",
		Admin:          admin,
		Label:          "test",
		Metadata:       pb.ContractMetadata{Name: "Original"},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance:     pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Try to update as non-admin
	msg := &pb.MsgUpdateContractMetadata{
		Signer:          "cosmos1notadmin",
		ContractAddress: contractAddr,
		Metadata:        pb.ContractMetadata{Name: "Hacked"},
	}

	_, err := suite.msgServer.UpdateContractMetadata(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Error(err)
	suite.Equal(codes.PermissionDenied, status.Code(err))
}

func (suite *MsgServerComprehensiveTestSuite) TestUpdateContractMetadata_ContractNotFound() {
	msg := &pb.MsgUpdateContractMetadata{
		Signer:          "cosmos1admin",
		ContractAddress: "cosmos1nonexistent",
		Metadata:        pb.ContractMetadata{Name: "Updated"},
	}

	_, err := suite.msgServer.UpdateContractMetadata(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Error(err)
	suite.ErrorIs(err, types.ErrContractNotFound)
}

// ============================
// UpdateSecurityPolicy Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestUpdateSecurityPolicy_Success() {
	// Register contract
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata:       pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{MaxGasPerTx: 1000000},
		Compliance:     pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Update security policy
	msg := &pb.MsgUpdateSecurityPolicy{
		Signer:          admin,
		ContractAddress: contractAddr,
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause:       true,
			MaxGasPerTx:      2000000,
			RateLimitPerUser: 50,
		},
	}

	resp, err := suite.msgServer.UpdateSecurityPolicy(sdk.WrapSDKContext(suite.ctx), msg)
	suite.NoError(err)
	suite.True(resp.Success)

	// Verify update
	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.True(stored.SecurityPolicy.AllowPause)
	suite.Equal(uint64(2000000), stored.SecurityPolicy.MaxGasPerTx)
	suite.Equal(uint64(50), stored.SecurityPolicy.RateLimitPerUser)
}

func (suite *MsgServerComprehensiveTestSuite) TestUpdateSecurityPolicy_NotAdmin() {
	// Register contract
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata:       pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance:     pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Try to update as non-admin
	msg := &pb.MsgUpdateSecurityPolicy{
		Signer:          "cosmos1attacker",
		ContractAddress: contractAddr,
		SecurityPolicy:  pb.SecurityPolicy{MaxGasPerTx: 999999999},
	}

	_, err := suite.msgServer.UpdateSecurityPolicy(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Error(err)
	suite.Equal(codes.PermissionDenied, status.Code(err))
}

// ============================
// PauseContract Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestPauseContract_Success() {
	// Register contract with pause enabled
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:  contractAddr,
		CodeId:   1,
		Creator:  admin,
		Admin:    admin,
		Label:    "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: true,
		},
		Compliance: pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Pause contract
	msg := &pb.MsgPauseContract{
		Signer:          admin,
		ContractAddress: contractAddr,
		Reason:          "maintenance",
	}

	resp, err := suite.msgServer.PauseContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.NoError(err)
	suite.True(resp.Success)

	// Verify paused
	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_PAUSED, stored.Status)
}

func (suite *MsgServerComprehensiveTestSuite) TestPauseContract_NotAllowed() {
	// Register contract with pause disabled
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:  contractAddr,
		CodeId:   1,
		Creator:  admin,
		Admin:    admin,
		Label:    "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: false,
		},
		Compliance: pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Try to pause
	msg := &pb.MsgPauseContract{
		Signer:          admin,
		ContractAddress: contractAddr,
		Reason:          "maintenance",
	}

	_, err := suite.msgServer.PauseContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Error(err)
}

func (suite *MsgServerComprehensiveTestSuite) TestPauseContract_NotAdmin() {
	// Register contract
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:  contractAddr,
		CodeId:   1,
		Creator:  admin,
		Admin:    admin,
		Label:    "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: true,
		},
		Compliance: pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Try to pause as non-admin
	msg := &pb.MsgPauseContract{
		Signer:          "cosmos1attacker",
		ContractAddress: contractAddr,
		Reason:          "attack",
	}

	_, err := suite.msgServer.PauseContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Error(err)
	suite.Equal(codes.PermissionDenied, status.Code(err))
}

// ============================
// UnpauseContract Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestUnpauseContract_Success() {
	// Register and pause contract
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:  contractAddr,
		CodeId:   1,
		Creator:  admin,
		Admin:    admin,
		Label:    "test",
		Metadata: pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause: true,
		},
		Compliance: pb.ComplianceRequirements{},
		Status:     pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))
	suite.NoError(suite.keeper.PauseContract(suite.ctx, contractAddr, admin, "test"))

	// Unpause contract
	msg := &pb.MsgUnpauseContract{
		Signer:          admin,
		ContractAddress: contractAddr,
	}

	resp, err := suite.msgServer.UnpauseContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.NoError(err)
	suite.True(resp.Success)

	// Verify active
	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_ACTIVE, stored.Status)
}

func (suite *MsgServerComprehensiveTestSuite) TestUnpauseContract_NotPaused() {
	// Register active contract
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata:       pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance:     pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Try to unpause
	msg := &pb.MsgUnpauseContract{
		Signer:          admin,
		ContractAddress: contractAddr,
	}

	_, err := suite.msgServer.UnpauseContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Error(err)
}

// ============================
// DeprecateContract Tests
// ============================

func (suite *MsgServerComprehensiveTestSuite) TestDeprecateContract_Success() {
	// Register contract
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata:       pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance:     pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Deprecate contract
	msg := &pb.MsgDeprecateContract{
		Signer:          admin,
		ContractAddress: contractAddr,
		Reason:          "v2 available",
		MigrationTarget: "cosmos1newcontract",
	}

	resp, err := suite.msgServer.DeprecateContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.NoError(err)
	suite.True(resp.Success)

	// Verify deprecated
	stored, _ := suite.keeper.GetContractInfo(suite.ctx, contractAddr)
	suite.Equal(pb.ContractStatus_CONTRACT_STATUS_DEPRECATED, stored.Status)
}

func (suite *MsgServerComprehensiveTestSuite) TestDeprecateContract_WithoutMigrationTarget() {
	// Register contract
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata:       pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance:     pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Deprecate without migration target
	msg := &pb.MsgDeprecateContract{
		Signer:          admin,
		ContractAddress: contractAddr,
		Reason:          "end of life",
		MigrationTarget: "",
	}

	resp, err := suite.msgServer.DeprecateContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.NoError(err)
	suite.True(resp.Success)
}

func (suite *MsgServerComprehensiveTestSuite) TestDeprecateContract_NotAdmin() {
	// Register contract
	contractAddr := "cosmos1contract"
	admin := "cosmos1admin"

	info := &pb.ContractInfo{
		Address:        contractAddr,
		CodeId:         1,
		Creator:        admin,
		Admin:          admin,
		Label:          "test",
		Metadata:       pb.ContractMetadata{Name: "Test"},
		SecurityPolicy: pb.SecurityPolicy{},
		Compliance:     pb.ComplianceRequirements{},
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	suite.NoError(suite.keeper.RegisterContract(suite.ctx, info))

	// Try to deprecate as non-admin
	msg := &pb.MsgDeprecateContract{
		Signer:          "cosmos1attacker",
		ContractAddress: contractAddr,
		Reason:          "malicious",
		MigrationTarget: "",
	}

	_, err := suite.msgServer.DeprecateContract(sdk.WrapSDKContext(suite.ctx), msg)
	suite.Error(err)
	suite.Equal(codes.PermissionDenied, status.Code(err))
}
