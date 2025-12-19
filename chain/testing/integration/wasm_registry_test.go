package integration

import (
	"fmt"
	"testing"
	"time"

	wasmdkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	contractpb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

// WASMRegistryTestSuite exercises the full WASM hook flow with the real
// contract registry keeper wired in via WASMTestContext.
type WASMRegistryTestSuite struct {
	suite.Suite
	ctx WASMTestContext
}

func TestWASMRegistryIntegration(t *testing.T) {
	suite.Run(t, new(WASMRegistryTestSuite))
}

func (s *WASMRegistryTestSuite) SetupTest() {
	s.ctx = SetupTestAppWithWasm(s.T())
}

// TestUploadContractCode verifies auto-registration happens during instantiate
// and that metadata defaults are persisted in the registry.
func (s *WASMRegistryTestSuite) TestUploadContractCode() {
	uploader := s.ctx.CreateAuthorizedUploader()
	contractAddr := s.ctx.SetupCompleteContract(uploader, nil, nil, nil)

	info, found := s.ctx.RegistryKeeper.GetContractInfo(s.ctx.Ctx, contractAddr.String())
	s.Require().True(found, "contract record must exist")
	s.Equal(contractAddr.String(), info.Address)
	s.Equal(uploader.String(), info.Creator)
	s.Equal(contractpb.ContractStatus_CONTRACT_STATUS_ACTIVE, info.Status)
	s.Contains(info.Metadata.Tags, "wasm")
}

// TestComplianceAndVCEnforcement ensures the registry enforced requirements
// (KYC, VC, confidence score, sanctions) are honoured by the WASM hooks.
func (s *WASMRegistryTestSuite) TestComplianceAndVCEnforcement() {
	uploader := s.ctx.CreateAuthorizedUploader()
	metadata := contractpb.ContractMetadata{
		Name:               "compliance-contract",
		Description:        "Requires VC + high confidence",
		Tags:               []string{"dex", "kyc-required"},
		RequiresVc:         true,
		RequiredVcTypes:    []string{"KYCVerification"},
		MinConfidenceScore: 80,
		RequiredKycLevel:   2,
		CheckSanctions:     true,
	}
	policy := contractpb.SecurityPolicy{
		AllowPause:       true,
		MaxGasPerTx:      1_000_000,
		RateLimitPerUser: 5,
	}
	compliance := contractpb.ComplianceRequirements{
		EnforceKyc:            true,
		MinKycLevel:           2,
		EnforceSanctionsCheck: true,
	}

	contractAddr := s.ctx.SetupCompleteContractWithPolicy(uploader, metadata, policy, compliance)

	user := sdk.AccAddress([]byte("compliance-user-addr"))
	err := s.ctx.ExecuteAsUser(contractAddr, user, true, nil)
	s.Require().Error(err, "execution must fail until requirements satisfied")

	// Fulfil compliance + VC requirements.
	s.ctx.SetUserKYC(user, 3)
	s.ctx.SetUserSanction(user, false)
	s.ctx.SetUserConfidence(user, 90)
	s.ctx.GrantUserVC(user, "KYCVerification")

	s.ctx.AdvanceBlock(time.Second)
	err = s.ctx.ExecuteAsUser(contractAddr, user, true, nil)
	s.Require().NoError(err, "execution should pass once guards satisfied")
}

// TestRateLimitAndMetrics verifies per-user rate limits and that violations get
// reflected in the registry metrics.
func (s *WASMRegistryTestSuite) TestRateLimitAndMetrics() {
	uploader := s.ctx.CreateAuthorizedUploader()
	policy := contractpb.SecurityPolicy{
		AllowPause:       true,
		RateLimitPerUser: 2,
		MaxGasPerTx:      500_000,
	}
	contractAddr := s.ctx.SetupCompleteContractWithPolicy(
		uploader,
		registryMetadata("rate-limit-contract"),
		policy,
		contractpb.ComplianceRequirements{},
	)

	user := sdk.AccAddress([]byte("rate-limit-user"))
	s.Require().NoError(s.ctx.ExecuteAsUser(contractAddr, user, true, nil))
	s.ctx.AdvanceBlock(time.Second)
	s.Require().NoError(s.ctx.ExecuteAsUser(contractAddr, user, true, nil))
	s.ctx.AdvanceBlock(time.Second)

	err := s.ctx.ExecuteAsUser(contractAddr, user, true, nil)
	s.Require().Error(err, "third execution should trip rate limit")
	s.Contains(err.Error(), "rate limit")

	metrics, found := s.ctx.RegistryKeeper.GetContractMetrics(s.ctx.Ctx, contractAddr.String())
	s.Require().True(found, "metrics should exist after rate limit violation")
	s.Equal(uint64(1), metrics.RateLimitViolations)

	// Advance an hour to reset the window and confirm execution works again.
	s.ctx.AdvanceBlock(time.Hour)
	s.Require().NoError(s.ctx.ExecuteAsUser(contractAddr, user, true, nil))
}

func registryMetadata(name string) contractpb.ContractMetadata {
	return contractpb.ContractMetadata{
		Name:        name,
		Description: fmt.Sprintf("%s integration contract", name),
		Tags:        []string{"wasm", "integration"},
	}
}

// TestContractMigrationRequiresAdmin ensures migrations enforce admin checks
// even when the underlying wasmd keeper is not present by falling back to the
// contract registry metadata.
func (s *WASMRegistryTestSuite) TestContractMigrationRequiresAdmin() {
	uploader := s.ctx.CreateAuthorizedUploader()
	contractAddr := s.ctx.SetupCompleteContract(uploader, nil, nil, nil)

	params := s.ctx.WasmKeeper.GetParams(s.ctx.Ctx)
	params.RequireAdminForMigrate = true
	s.Require().NoError(s.ctx.WasmKeeper.SetParams(s.ctx.Ctx, params))

	mockOps := newMockContractOps()
	s.ctx.WasmKeeper.SetContractOpsFactory(func(_ *wasmdkeeper.Keeper) wasmtypes.ContractOpsKeeper {
		return mockOps
	})

	nonAdmin := s.ctx.CreateAuthorizedUploader()
	_, err := s.ctx.WasmKeeper.Migrate(s.ctx.Ctx, contractAddr, nonAdmin, 42, []byte("{}"))
	s.Require().Error(err)
	s.Contains(err.Error(), "not admin")
	s.Empty(mockOps.migrations, "migration ops should not run when admin check fails")
}

// TestContractMigrationSuccess wires the mock contract ops into the keeper so
// we can assert the migrate call is forwarded once admin checks pass.
func (s *WASMRegistryTestSuite) TestContractMigrationSuccess() {
	uploader := s.ctx.CreateAuthorizedUploader()
	contractAddr := s.ctx.SetupCompleteContract(uploader, nil, nil, nil)

	params := s.ctx.WasmKeeper.GetParams(s.ctx.Ctx)
	params.RequireAdminForMigrate = true
	s.Require().NoError(s.ctx.WasmKeeper.SetParams(s.ctx.Ctx, params))

	mockOps := newMockContractOps()
	mockOps.migrateResp = []byte("migrated-to-v2")
	s.ctx.WasmKeeper.SetContractOpsFactory(func(_ *wasmdkeeper.Keeper) wasmtypes.ContractOpsKeeper {
		return mockOps
	})

	payload := []byte(`{"upgrade":"v2"}`)
	data, err := s.ctx.WasmKeeper.Migrate(s.ctx.Ctx, contractAddr, uploader, 99, payload)
	s.Require().NoError(err)
	s.Equal([]byte("migrated-to-v2"), data)

	s.Require().Len(mockOps.migrations, 1)
	call := mockOps.migrations[0]
	s.Equal(contractAddr.String(), call.Contract.String())
	s.Equal(uploader.String(), call.Caller.String())
	s.Equal(uint64(99), call.CodeID)
	s.Equal(payload, call.Msg)
}

type migrationCall struct {
	Contract sdk.AccAddress
	Caller   sdk.AccAddress
	CodeID   uint64
	Msg      []byte
}

type mockContractOps struct {
	migrations  []migrationCall
	migrateResp []byte
	migrateErr  error
}

var _ wasmtypes.ContractOpsKeeper = (*mockContractOps)(nil)

func newMockContractOps() *mockContractOps {
	return &mockContractOps{
		migrateResp: []byte("migrated"),
	}
}

func (m *mockContractOps) Create(ctx sdk.Context, creator sdk.AccAddress, wasmCode []byte, instantiateAccess *wasmtypes.AccessConfig) (uint64, []byte, error) {
	return 0, nil, fmt.Errorf("create not implemented")
}

func (m *mockContractOps) Instantiate(ctx sdk.Context, codeID uint64, creator, admin sdk.AccAddress, initMsg []byte, label string, deposit sdk.Coins) (sdk.AccAddress, []byte, error) {
	return nil, nil, fmt.Errorf("instantiate not implemented")
}

func (m *mockContractOps) Instantiate2(ctx sdk.Context, codeID uint64, creator, admin sdk.AccAddress, initMsg []byte, label string, deposit sdk.Coins, salt []byte, fixMsg bool) (sdk.AccAddress, []byte, error) {
	return nil, nil, fmt.Errorf("instantiate2 not implemented")
}

func (m *mockContractOps) Execute(ctx sdk.Context, contractAddress, caller sdk.AccAddress, msg []byte, coins sdk.Coins) ([]byte, error) {
	return nil, fmt.Errorf("execute not implemented")
}

func (m *mockContractOps) Migrate(ctx sdk.Context, contractAddress, caller sdk.AccAddress, newCodeID uint64, msg []byte) ([]byte, error) {
	m.migrations = append(m.migrations, migrationCall{
		Contract: contractAddress,
		Caller:   caller,
		CodeID:   newCodeID,
		Msg:      append([]byte(nil), msg...),
	})
	if m.migrateErr != nil {
		return nil, m.migrateErr
	}
	return append([]byte(nil), m.migrateResp...), nil
}

func (m *mockContractOps) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
	return nil, fmt.Errorf("sudo not implemented")
}

func (m *mockContractOps) UpdateContractAdmin(ctx sdk.Context, contractAddress, caller, newAdmin sdk.AccAddress) error {
	return fmt.Errorf("update admin not implemented")
}

func (m *mockContractOps) ClearContractAdmin(ctx sdk.Context, contractAddress, caller sdk.AccAddress) error {
	return fmt.Errorf("clear admin not implemented")
}

func (m *mockContractOps) PinCode(ctx sdk.Context, codeID uint64) error {
	return fmt.Errorf("pin not implemented")
}

func (m *mockContractOps) UnpinCode(ctx sdk.Context, codeID uint64) error {
	return fmt.Errorf("unpin not implemented")
}

func (m *mockContractOps) SetContractInfoExtension(ctx sdk.Context, contract sdk.AccAddress, extra wasmtypes.ContractInfoExtension) error {
	return fmt.Errorf("set contract extension not implemented")
}

func (m *mockContractOps) SetAccessConfig(ctx sdk.Context, codeID uint64, caller sdk.AccAddress, newConfig wasmtypes.AccessConfig) error {
	return fmt.Errorf("set access config not implemented")
}
