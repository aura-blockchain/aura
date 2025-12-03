package keeper_test

import (
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

	// Create keeper
	suite.keeper = keeper.NewKeeper(
		suite.cdc,
		storeKey,
		nil, // wasmd keeper not needed for invariant tests
		"aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
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
	suite.keeper.SetParams(ctx, *params)

	inv := keeper.CodeSizeLimitsInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "should pass with reasonable code size")
	suite.Empty(msg)

	// Test: Zero max code size (invalid)
	params.MaxWasmCodeSize = 0
	suite.keeper.SetParams(ctx, *params)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with zero max code size")
	suite.Contains(msg, "cannot be zero")

	// Test: Extremely large max code size (invalid)
	params.MaxWasmCodeSize = 20 * 1024 * 1024 // 20MB - too large
	suite.keeper.SetParams(ctx, *params)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with excessive max code size")
	suite.Contains(msg, "exceeds reasonable limit")
}

func (suite *InvariantsTestSuite) TestUploadAuthEnforcementInvariant() {
	ctx := suite.ctx

	// Test: AccessTypeNobody with no authorized uploaders (valid)
	params := types.DefaultParams()
	params.CodeUploadAccess = &types.AccessConfig{
		Permission: types.AccessTypeNobody,
	}
	suite.keeper.SetParams(ctx, *params)

	inv := keeper.UploadAuthEnforcementInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "should pass when nobody can upload and no authorized uploaders")
	suite.Empty(msg)

	// Test: AccessTypeNobody but with authorized uploader (invalid)
	testAddr := "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	suite.keeper.AuthorizeUploader(ctx, testAddr)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail when authorized uploaders exist with NOBODY access")
	suite.Contains(msg, "authorized uploaders exist")

	// Clean up
	suite.keeper.RevokeUploader(ctx, testAddr)

	// Test: AccessTypeOnlyAddress with valid address
	params.CodeUploadAccess = &types.AccessConfig{
		Permission: types.AccessTypeOnlyAddress,
		Address:    testAddr,
	}
	suite.keeper.SetParams(ctx, *params)

	msg, broken = inv(ctx)
	suite.False(broken, "should pass with valid address")
	suite.Empty(msg)

	// Test: AccessTypeOnlyAddress with invalid address
	params.CodeUploadAccess = &types.AccessConfig{
		Permission: types.AccessTypeOnlyAddress,
		Address:    "invalid-address",
	}
	suite.keeper.SetParams(ctx, *params)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with invalid address")
	suite.Contains(msg, "invalid upload address")

	// Test: AccessTypeAnyOfAddresses with valid addresses
	params.CodeUploadAccess = &types.AccessConfig{
		Permission: types.AccessTypeAnyOfAddresses,
		Addresses:  []string{testAddr, "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"},
	}
	suite.keeper.SetParams(ctx, *params)

	msg, broken = inv(ctx)
	suite.False(broken, "should pass with valid addresses")
	suite.Empty(msg)

	// Test: AccessTypeAnyOfAddresses with invalid address in list
	params.CodeUploadAccess = &types.AccessConfig{
		Permission: types.AccessTypeAnyOfAddresses,
		Addresses:  []string{testAddr, "invalid-address"},
	}
	suite.keeper.SetParams(ctx, *params)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with invalid address in list")
	suite.Contains(msg, "invalid address in upload addresses list")
}

func (suite *InvariantsTestSuite) TestGasCapsInvariant() {
	ctx := suite.ctx

	// Test: Valid gas cap
	params := types.DefaultParams()
	params.MaxGasWasmExecution = 10_000_000 // 10M gas - reasonable
	suite.keeper.SetParams(ctx, *params)

	inv := keeper.GasCapsInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "should pass with reasonable gas cap")
	suite.Empty(msg)

	// Test: Zero gas cap (invalid)
	params.MaxGasWasmExecution = 0
	suite.keeper.SetParams(ctx, *params)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with zero gas cap")
	suite.Contains(msg, "cannot be zero")

	// Test: Excessive gas cap (invalid)
	params.MaxGasWasmExecution = 100_000_000 // 100M gas - too high
	suite.keeper.SetParams(ctx, *params)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with excessive gas cap")
	suite.Contains(msg, "exceeds reasonable limit")

	// Test: Too low gas cap (invalid)
	params.MaxGasWasmExecution = 50_000 // 50K gas - too low
	suite.keeper.SetParams(ctx, *params)

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with too low gas cap")
	suite.Contains(msg, "below minimum reasonable limit")
}

func (suite *InvariantsTestSuite) TestAdminEnforcementInvariant() {
	ctx := suite.ctx

	// Test: Admin enforcement disabled (should pass)
	params := types.DefaultParams()
	params.RequireAdminForMigrate = false
	suite.keeper.SetParams(ctx, *params)

	inv := keeper.AdminEnforcementInvariant(suite.keeper)
	msg, broken := inv(ctx)
	suite.False(broken, "should pass when admin enforcement is disabled")
	suite.Empty(msg)

	// Test: Admin enforcement enabled with valid admin entries
	params.RequireAdminForMigrate = true
	suite.keeper.SetParams(ctx, *params)

	contractAddr, _ := sdk.AccAddressFromBech32("aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")
	adminAddr, _ := sdk.AccAddressFromBech32("aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh")

	suite.keeper.SetContractAdmin(ctx, contractAddr, adminAddr)

	msg, broken = inv(ctx)
	suite.False(broken, "should pass with valid admin entry")
	suite.Empty(msg)

	// Test: Invalid contract address in admin storage
	store := ctx.KVStore(suite.keeper.GetStoreKey())
	invalidKey := append(types.ContractAdminPrefix, []byte("invalid-address")...)
	store.Set(invalidKey, []byte("aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"))

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with invalid contract address")
	suite.Contains(msg, "invalid contract address")

	// Clean up
	store.Delete(invalidKey)

	// Test: Empty admin address
	emptyAdminKey := append(types.ContractAdminPrefix, []byte("aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")...)
	store.Set(emptyAdminKey, []byte(""))

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with empty admin address")
	suite.Contains(msg, "empty admin address")

	// Clean up
	store.Delete(emptyAdminKey)

	// Test: Invalid admin address
	invalidAdminKey := append(types.ContractAdminPrefix, []byte("aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")...)
	store.Set(invalidAdminKey, []byte("invalid-admin"))

	msg, broken = inv(ctx)
	suite.True(broken, "should fail with invalid admin address")
	suite.Contains(msg, "invalid admin address")
}

func (suite *InvariantsTestSuite) TestCodeSizeLimitViolations() {
	ctx := suite.ctx

	// Test: Code upload respects size limits
	params := types.DefaultParams()
	maxSize := uint64(1000) // 1KB max
	params.MaxWasmCodeSize = maxSize
	suite.keeper.SetParams(ctx, *params)

	// Test valid size code
	validCode := make([]byte, maxSize-1)
	err := suite.keeper.ValidateContractUpload(ctx, "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", validCode)
	suite.NoError(err, "should accept code within size limit")

	// Test oversized code
	oversizedCode := make([]byte, maxSize+1)
	err = suite.keeper.ValidateContractUpload(ctx, "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", oversizedCode)
	suite.Error(err, "should reject oversized code")
	suite.Contains(err.Error(), "exceeds maximum")
}

func (suite *InvariantsTestSuite) TestUploadAuthorizationEnforcement() {
	ctx := suite.ctx

	// Test: Unauthorized uploader cannot upload when access is restricted
	params := types.DefaultParams()
	params.CodeUploadAccess = &types.AccessConfig{
		Permission: types.AccessTypeOnlyAddress,
		Address:    "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
	}
	suite.keeper.SetParams(ctx, *params)

	unauthorizedAddr := "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	code := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	err := suite.keeper.ValidateContractUpload(ctx, unauthorizedAddr, code)
	suite.Error(err, "should reject unauthorized uploader")
	suite.Contains(err.Error(), "not authorized")

	// Test: Authorized uploader can upload
	authorizedAddr := "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
	err = suite.keeper.ValidateContractUpload(ctx, authorizedAddr, code)
	suite.NoError(err, "should accept authorized uploader")
}

func (suite *InvariantsTestSuite) TestGasConsumptionTracking() {
	// Test: Gas consumption is tracked per contract execution
	execCtx := types.NewExecutionContext(10)

	contractAddr := "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	gasConsumed := uint64(50000)

	execCtx.RecordGasConsumption(contractAddr, gasConsumed)
	suite.Equal(gasConsumed, execCtx.GetGasConsumed(contractAddr), "gas should be tracked")

	// Test: Gas accumulates
	execCtx.RecordGasConsumption(contractAddr, gasConsumed)
	suite.Equal(gasConsumed*2, execCtx.GetGasConsumed(contractAddr), "gas should accumulate")

	// Test: Different contracts tracked separately
	contractAddr2 := "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
	execCtx.RecordGasConsumption(contractAddr2, gasConsumed)
	suite.Equal(gasConsumed, execCtx.GetGasConsumed(contractAddr2), "separate contract gas tracking")
	suite.Equal(gasConsumed*2, execCtx.GetGasConsumed(contractAddr), "first contract gas unchanged")
}

func (suite *InvariantsTestSuite) TestEventEmissionOnOperations() {
	ctx := suite.ctx

	// Test: Events are emitted for security operations
	contractAddr := "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	sender := "aura1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"

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
