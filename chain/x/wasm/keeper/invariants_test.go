package keeper_test

import (
	"encoding/json"
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/wasm/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type InvariantsTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.BinaryCodec
}

func (suite *InvariantsTestSuite) SetupTest() {
	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create test context
	testCtx := testutil.DefaultContextWithDB(suite.T(), storeKey, storetypes.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx

	// Create codec
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	suite.cdc = codec.NewProtoCodec(registry)

	// Create keeper with valid authority address
	suite.keeper = keeper.NewKeeper(
		suite.cdc,
		storeKey,
		nil, // wasmd keeper not needed for invariant tests
		sdk.AccAddress("authority___________").String(),
	)
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

func (suite *InvariantsTestSuite) TestAllInvariants() {
	ctx := suite.ctx

	// Test: All invariants on empty store
	inv := keeper.AllInvariants(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	// Test individual invariants directly without registry
	// Cosmos SDK v0.50 doesn't have NewInvariantRegistry

	ctx := suite.ctx

	// Test ParamsInvariant
	inv := keeper.ParamsInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "params invariant should pass")
	suite.Empty(msg)

	// Test SecurityStatsInvariant
	inv = keeper.SecurityStatsInvariant(suite.keeper)
	msg, broken = inv(ctx)
	suite.False(broken, "security stats invariant should pass")
	suite.Empty(msg)

	// Test PausedContractsInvariant
	inv = keeper.PausedContractsInvariant(suite.keeper)
	msg, broken = inv(ctx)
	suite.False(broken, "paused contracts invariant should pass")
	suite.Empty(msg)

	// Test AuthorizedUploadersInvariant
	inv = keeper.AuthorizedUploadersInvariant(suite.keeper)
	msg, broken = inv(ctx)
	suite.False(broken, "authorized uploaders invariant should pass")
	suite.Empty(msg)

	// Test CodeSizeLimitsInvariant
	inv = keeper.CodeSizeLimitsInvariant(suite.keeper)
	msg, broken = inv(ctx)
	suite.False(broken, "code size limits invariant should pass")
	suite.Empty(msg)

	// Test UploadAuthEnforcementInvariant
	inv = keeper.UploadAuthEnforcementInvariant(suite.keeper)
	msg, broken = inv(ctx)
	suite.False(broken, "upload auth enforcement invariant should pass")
	suite.Empty(msg)

	// Test GasCapsInvariant
	inv = keeper.GasCapsInvariant(suite.keeper)
	msg, broken = inv(ctx)
	suite.False(broken, "gas caps invariant should pass")
	suite.Empty(msg)

	// Test AdminEnforcementInvariant
	inv = keeper.AdminEnforcementInvariant(suite.keeper)
	msg, broken = inv(ctx)
	suite.False(broken, "admin enforcement invariant should pass")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestCodeSizeLimitsInvariant() {
	ctx := suite.ctx

	// Test: Valid code size parameters
	params := types.DefaultParams()
	params.MaxWasmCodeSize = 600 * 1024 // 600KB - reasonable
	err := suite.keeper.SetParams(ctx, *params)
	suite.NoError(err)

	inv := keeper.CodeSizeLimitsInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "should pass with reasonable code size")
	suite.Empty(msg)

	// Test: Zero max code size (invalid)
	// SetParams validates and rejects 0, so we write directly to store
	invalidParams := *params
	invalidParams.MaxWasmCodeSize = 0
	suite.writeParamsDirectly(ctx, invalidParams)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with zero max code size")
	suite.Contains(msg, "cannot be zero")

	// Test: Extremely large max code size (invalid)
	// SetParams validates and rejects > 10MB, so we write directly to store
	invalidParams.MaxWasmCodeSize = 20 * 1024 * 1024 // 20MB - too large
	suite.writeParamsDirectly(ctx, invalidParams)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with excessive max code size")
	suite.Contains(msg, "exceeds reasonable limit")
}

// writeParamsDirectly writes params directly to store, bypassing validation
// This is needed to test invariants that should fail with invalid params
func (suite *InvariantsTestSuite) writeParamsDirectly(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(suite.keeper.GetStoreKey())
	bz, _ := json.Marshal(params)
	store.Set(types.ParamsKey, bz)
}

func (suite *InvariantsTestSuite) TestUploadAuthEnforcementInvariant() {
	ctx := suite.ctx

	// Use valid bech32 test addresses
	testAddr := sdk.AccAddress("test_address________").String()
	testAddr2 := sdk.AccAddress("test_address2_______").String()

	// Test: AccessTypeNobody with no authorized uploaders (valid)
	params := types.DefaultParams()
	params.CodeUploadAccess = types.AccessConfig{
		Permission: types.AccessTypeNobody,
	}
	err := suite.keeper.SetParams(ctx, *params)
	suite.NoError(err)

	inv := keeper.UploadAuthEnforcementInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "should pass when nobody can upload and no authorized uploaders")
	suite.Empty(msg)

	// Test: AccessTypeNobody but with authorized uploader (invalid)
	suite.keeper.AuthorizeUploader(ctx, testAddr)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail when authorized uploaders exist with NOBODY access")
	suite.Contains(msg, "authorized uploaders exist")

	// Clean up
	suite.keeper.RevokeUploader(ctx, testAddr)

	// Test: AccessTypeOnlyAddress with valid address
	params.CodeUploadAccess = types.AccessConfig{
		Permission: types.AccessTypeOnlyAddress,
		Address:    testAddr,
	}
	err = suite.keeper.SetParams(ctx, *params)
	suite.NoError(err)

	msg, broken = inv(ctx)
	suite.False(broken, "should pass with valid address")
	suite.Empty(msg)

	// Test: AccessTypeOnlyAddress with invalid address
	// Need to write directly to store to bypass validation
	invalidParams := *params
	invalidParams.CodeUploadAccess = types.AccessConfig{
		Permission: types.AccessTypeOnlyAddress,
		Address:    "invalid-address",
	}
	suite.writeParamsDirectly(ctx, invalidParams)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with invalid address")
	suite.Contains(msg, "invalid upload address")

	// Test: AccessTypeAnyOfAddresses with valid addresses
	params.CodeUploadAccess = types.AccessConfig{
		Permission: types.AccessTypeAnyOfAddresses,
		Addresses:  []string{testAddr, testAddr2},
	}
	err = suite.keeper.SetParams(ctx, *params)
	suite.NoError(err)

	msg, broken = inv(ctx)
	suite.False(broken, "should pass with valid addresses")
	suite.Empty(msg)

	// Test: AccessTypeAnyOfAddresses with invalid address in list
	// Write directly to store to bypass validation
	invalidParams = *params
	invalidParams.CodeUploadAccess = types.AccessConfig{
		Permission: types.AccessTypeAnyOfAddresses,
		Addresses:  []string{testAddr, "invalid-address"},
	}
	suite.writeParamsDirectly(ctx, invalidParams)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with invalid address in list")
	suite.Contains(msg, "invalid address in upload addresses list")
}

func (suite *InvariantsTestSuite) TestGasCapsInvariant() {
	ctx := suite.ctx

	// Test: Valid gas cap
	params := types.DefaultParams()
	params.MaxGasWasmExecution = 10_000_000 // 10M gas - reasonable
	err := suite.keeper.SetParams(ctx, *params)
	suite.NoError(err)

	inv := keeper.GasCapsInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "should pass with reasonable gas cap")
	suite.Empty(msg)

	// Test: Zero gas cap (invalid)
	// SetParams validates and rejects 0, so write directly
	invalidParams := *params
	invalidParams.MaxGasWasmExecution = 0
	suite.writeParamsDirectly(ctx, invalidParams)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with zero gas cap")
	suite.Contains(msg, "cannot be zero")

	// Test: Excessive gas cap (invalid)
	// SetParams may or may not reject this, write directly to be sure
	invalidParams.MaxGasWasmExecution = 100_000_000 // 100M gas - too high
	suite.writeParamsDirectly(ctx, invalidParams)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with excessive gas cap")
	suite.Contains(msg, "exceeds reasonable limit")

	// Test: Too low gas cap (invalid)
	invalidParams.MaxGasWasmExecution = 50_000 // 50K gas - too low
	suite.writeParamsDirectly(ctx, invalidParams)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with too low gas cap")
	suite.Contains(msg, "below minimum reasonable limit")
}

func (suite *InvariantsTestSuite) TestAdminEnforcementInvariant() {
	ctx := suite.ctx

	// Use valid bech32 test addresses
	contractAddr := sdk.AccAddress("test_contract_______")
	adminAddr := sdk.AccAddress("test_admin__________")

	// Test: Admin enforcement disabled (should pass)
	params := types.DefaultParams()
	params.RequireAdminForMigrate = false
	err := suite.keeper.SetParams(ctx, *params)
	suite.NoError(err)

	inv := keeper.AdminEnforcementInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "should pass when admin enforcement is disabled")
	suite.Empty(msg)

	// Test: Admin enforcement enabled with valid admin entries
	params.RequireAdminForMigrate = true
	err = suite.keeper.SetParams(ctx, *params)
	suite.NoError(err)

	err = suite.keeper.SetContractAdmin(ctx, contractAddr, adminAddr)
	suite.NoError(err)

	msg, broken = inv(ctx)
	suite.False(broken, "should pass with valid admin entry")
	suite.Empty(msg)

	// Clean up valid admin entry
	suite.keeper.DeleteContractAdmin(ctx, contractAddr)

	// Test: Invalid contract address in admin storage
	store := ctx.KVStore(suite.keeper.GetStoreKey())
	invalidKey := append(types.ContractAdminPrefix, []byte("invalid-address")...)
	store.Set(invalidKey, []byte(adminAddr.String()))

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with invalid contract address")
	suite.Contains(msg, "invalid contract address")

	// Clean up
	store.Delete(invalidKey)

	// Test: Empty admin address
	emptyAdminKey := append(types.ContractAdminPrefix, []byte(contractAddr.String())...)
	store.Set(emptyAdminKey, []byte(""))

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with empty admin address")
	suite.Contains(msg, "empty admin address")

	// Clean up
	store.Delete(emptyAdminKey)

	// Test: Invalid admin address
	invalidAdminKey := append(types.ContractAdminPrefix, []byte(contractAddr.String())...)
	store.Set(invalidAdminKey, []byte("invalid-admin"))

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with invalid admin address")
	suite.Contains(msg, "invalid admin address")
}

func (suite *InvariantsTestSuite) TestCodeSizeLimitViolations() {
	ctx := suite.ctx

	// Use valid test address
	testAddr := sdk.AccAddress("test_uploader______").String()

	// Authorize the uploader first
	suite.keeper.AuthorizeUploader(ctx, testAddr)

	// Test: Code upload respects size limits
	params := types.DefaultParams()
	maxSize := uint64(1000) // 1KB max
	params.MaxWasmCodeSize = maxSize
	err := suite.keeper.SetParams(ctx, *params)
	suite.NoError(err)

	// Test valid size code
	validCode := make([]byte, maxSize-1)
	err = suite.keeper.ValidateContractUpload(ctx, testAddr, validCode)
	suite.NoError(err, "should accept code within size limit")

	// Test oversized code
	oversizedCode := make([]byte, maxSize+1)
	err = suite.keeper.ValidateContractUpload(ctx, testAddr, oversizedCode)
	suite.Error(err, "should reject oversized code")
	suite.Contains(err.Error(), "exceeds maximum")
}

func (suite *InvariantsTestSuite) TestUploadAuthorizationEnforcement() {
	ctx := suite.ctx

	// Use valid test addresses
	authorizedAddr := sdk.AccAddress("authorized_addr_____").String()
	unauthorizedAddr := sdk.AccAddress("unauthorized_addr___").String()

	// Test: Unauthorized uploader cannot upload when access is restricted
	// Note: ValidateContractUpload checks IsAuthorizedUploader (store-based), not params.CodeUploadAccess.Address
	params := types.DefaultParams()
	params.CodeUploadAccess = types.AccessConfig{
		Permission: types.AccessTypeOnlyAddress,
		Address:    authorizedAddr,
	}
	err := suite.keeper.SetParams(ctx, *params)
	suite.NoError(err)

	code := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	err = suite.keeper.ValidateContractUpload(ctx, unauthorizedAddr, code)
	suite.Error(err, "should reject unauthorized uploader")
	suite.Contains(err.Error(), "not authorized")

	// Test: Authorized uploader can upload
	// Must explicitly authorize the address via AuthorizeUploader
	suite.keeper.AuthorizeUploader(ctx, authorizedAddr)
	err = suite.keeper.ValidateContractUpload(ctx, authorizedAddr, code)
	suite.NoError(err, "should accept authorized uploader")
}

func (suite *InvariantsTestSuite) TestGasConsumptionTracking() {
	// Test: Gas consumption is tracked per contract execution
	execCtx := types.NewExecutionContext(10)

	// Use valid test addresses
	contractAddr := sdk.AccAddress("test_contract1______").String()
	contractAddr2 := sdk.AccAddress("test_contract2______").String()
	gasConsumed := uint64(50000)

	execCtx.RecordGasConsumption(contractAddr, gasConsumed)
	suite.Equal(gasConsumed, execCtx.GetGasConsumed(contractAddr), "gas should be tracked")

	// Test: Gas accumulates
	execCtx.RecordGasConsumption(contractAddr, gasConsumed)
	suite.Equal(gasConsumed*2, execCtx.GetGasConsumed(contractAddr), "gas should accumulate")

	// Test: Different contracts tracked separately
	execCtx.RecordGasConsumption(contractAddr2, gasConsumed)
	suite.Equal(gasConsumed, execCtx.GetGasConsumed(contractAddr2), "separate contract gas tracking")
	suite.Equal(gasConsumed*2, execCtx.GetGasConsumed(contractAddr), "first contract gas unchanged")
}

func (suite *InvariantsTestSuite) TestEventEmissionOnOperations() {
	ctx := suite.ctx

	// Test: Events are emitted for security operations
	// Use valid test addresses
	contractAddr := sdk.AccAddress("event_contract______").String()
	sender := sdk.AccAddress("event_sender________").String()

	// Create security audit event
	event := types.NewSecurityAuditEvent(
		types.EventTypeStoreCode,
		contractAddr,
		sender,
		ctx,
		true,
		"",
	)

	suite.Equal(types.EventTypeStoreCode, event.EventType)
	suite.Equal(contractAddr, event.ContractAddr)
	suite.Equal(sender, event.Sender)
	suite.True(event.Success)

	// Test: Event data can be added
	event.AddData("code_id", uint64(1))
	event.AddData("code_size", uint64(5000))

	suite.Equal(uint64(1), event.AdditionalData["code_id"])
	suite.Equal(uint64(5000), event.AdditionalData["code_size"])

	// Log the event (this would emit SDK events in real execution)
	suite.keeper.LogSecurityEvent(ctx, event)

	// Verify event was logged via SDK event manager
	events := ctx.EventManager().Events()
	suite.NotEmpty(events, "events should be emitted")
}
