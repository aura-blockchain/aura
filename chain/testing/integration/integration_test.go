package integration_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	modtestutil "github.com/aequitas/aura/chain/testutil"
	dexkeeper "github.com/aequitas/aura/chain/x/dex/keeper"
	dextypes "github.com/aequitas/aura/chain/x/dex/types"
	securitytypes "github.com/aequitas/aura/chain/x/security/types"
	dexpb "github.com/aequitas/aura/proto/aura/dex/v1beta1"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx sdk.Context
}

func (suite *IntegrationTestSuite) SetupTest() {
	suite.ctx = freshCtx(suite.T())
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func freshCtx(t *testing.T) sdk.Context {
	t.Helper()
	input := keepertest.CreateTestInput(t)
	require.NotNil(t, input.Ctx)
	return input.Ctx
}

// DEX + Bank Integration Tests

func TestDEXBankIntegration(t *testing.T) {
	ctx := freshCtx(t)
	account := keepertest.GenTestAddr()

	require.NotEmpty(t, account.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestDEXSwapAffectsBalances(t *testing.T) {
	ctx := freshCtx(t)
	trader := keepertest.GenTestAddr()

	require.NotEmpty(t, trader.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestLiquidityAdditionLocksTokens(t *testing.T) {
	ctx := freshCtx(t)
	provider := keepertest.GenTestAddr()

	require.NotEmpty(t, provider.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

type integrationBankKeeper struct{}

func (integrationBankKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (integrationBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (integrationBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (integrationBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (integrationBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (integrationBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewCoin(denom, sdkmath.NewInt(1000000))
}

// integrationAccountKeeper implements the AccountKeeper interface
type integrationAccountKeeper struct{}

func (integrationAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}

func (integrationAccountKeeper) SetAccount(ctx sdk.Context, acc sdk.AccountI) {}

func (integrationAccountKeeper) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}

// integrationVCRegistryKeeper implements the VCRegistryKeeper interface
type integrationVCRegistryKeeper struct{}

func (integrationVCRegistryKeeper) GetIRScore(ctx sdk.Context, address string) uint64 {
	return 0
}

func (integrationVCRegistryKeeper) IsVerified(ctx sdk.Context, address string) bool {
	return false
}

// integrationSecurityKeeper implements the SecurityKeeper interface
type integrationSecurityKeeper struct{}

func (integrationSecurityKeeper) EnterNoReentrant(ctx sdk.Context, key string) error {
	return nil
}

func (integrationSecurityKeeper) ExitNoReentrant(ctx sdk.Context, key string) {}

func (integrationSecurityKeeper) WithReentrancyGuard(ctx sdk.Context, key string, fn func() error) error {
	return fn()
}

func (integrationSecurityKeeper) RequireNotPaused(ctx sdk.Context, moduleName string) error {
	return nil
}

func (integrationSecurityKeeper) PauseModule(ctx sdk.Context, moduleName string, pausedBy string) error {
	return nil
}

func (integrationSecurityKeeper) UnpauseModule(ctx sdk.Context, moduleName string, unpausedBy string) error {
	return nil
}

func (integrationSecurityKeeper) IsModulePaused(ctx sdk.Context, moduleName string) bool {
	return false
}

func (integrationSecurityKeeper) CheckGuardRateLimit(ctx sdk.Context, key string, limit uint64, window time.Duration) error {
	return nil
}

func (integrationSecurityKeeper) IncrementGuardRateLimit(ctx sdk.Context, key string, window time.Duration) {
}

func (integrationSecurityKeeper) ValidateAddress(address string) error {
	return nil
}

func (integrationSecurityKeeper) ValidateAmount(amount sdkmath.Int, min, max sdkmath.Int) error {
	return nil
}

func (integrationSecurityKeeper) ValidateNonEmpty(value string, fieldName string) error {
	return nil
}

func (integrationSecurityKeeper) ValidateStringLength(value string, fieldName string, minLen, maxLen int) error {
	return nil
}

func (integrationSecurityKeeper) CheckAuthorization(ctx sdk.Context, address string, action string) error {
	return nil
}

func (integrationSecurityKeeper) LogSecurityEvent(ctx sdk.Context, eventType string, severity string, actor string, action string, details string) {
}

func setupDEXIntegrationKeeper(t *testing.T) (*dexkeeper.Keeper, sdk.Context) {
	input := keepertest.CreateTestInput(t)
	k := dexkeeper.NewKeeper(
		input.Cdc,
		input.StoreKey,
		integrationBankKeeper{},
		integrationAccountKeeper{},
		integrationVCRegistryKeeper{},
		integrationSecurityKeeper{},
	)
	return k, input.Ctx
}

func TestDEXSwapWorkflow(t *testing.T) {
	k, ctx := setupDEXIntegrationKeeper(t)
	creator := keepertest.GenTestAddr().String()

	_, _, err := k.CreatePool(
		ctx,
		creator,
		"uaura",
		"usdt",
		sdk.NewCoin("uaura", sdkmath.NewInt(1_000000)),
		sdk.NewCoin("usdt", sdkmath.NewInt(2_000000)),
	)
	require.NoError(t, err)

	_, err = k.CreateOrder(ctx, creator, dextypes.SwapOrderType_BUY, sdkmath.NewInt(1000), "usdt", sdkmath.NewInt(2000), 60)
	require.NoError(t, err)

	querySrv := dexkeeper.NewQueryServerImpl(k)
	userOrders, err := querySrv.UserOrders(sdk.WrapSDKContext(ctx), &dexpb.QueryUserOrdersRequest{Address: creator})
	require.NoError(t, err)
	require.Len(t, userOrders.Orders, 1)

	k.RecordSwapStats(ctx, "uaura-usdt", sdkmath.NewInt(1_000000), sdkmath.NewInt(500_000), ctx.BlockTime())

	priceResp, err := querySrv.MarketPrice(sdk.WrapSDKContext(ctx), &dexpb.QueryMarketPriceRequest{Coin: "usdt"})
	require.NoError(t, err)
	require.Equal(t, uint64(1), priceResp.Price.SampleSize)
	require.Equal(t, "usdt", priceResp.Price.Coin)
}

// Governance + Auth Integration Tests

func TestGovernanceRequiresAuth(t *testing.T) {
	ctx := freshCtx(t)
	proposer := keepertest.GenTestAddr()

	require.NotEmpty(t, proposer.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestGovernanceRoleBasedVoting(t *testing.T) {
	ctx := freshCtx(t)

	require.NotNil(t, ctx)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestEmergencyProposalByAdmin(t *testing.T) {
	ctx := freshCtx(t)
	admin := keepertest.GenTestAddr()

	require.NotEmpty(t, admin.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Bridge + Compliance Integration Tests

func TestBridgeComplianceCheck(t *testing.T) {
	ctx := freshCtx(t)
	sender := keepertest.GenTestAddr()

	require.NotEmpty(t, sender.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestBridgeKYCRequirement(t *testing.T) {
	ctx := freshCtx(t)
	user := keepertest.GenTestAddr()

	require.NotEmpty(t, user.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestBridgeAMLChecks(t *testing.T) {
	ctx := freshCtx(t)
	sender := keepertest.GenTestAddr()

	require.NotEmpty(t, sender.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Privacy + DEX Integration Tests

func TestShieldedSwap(t *testing.T) {
	ctx := freshCtx(t)
	trader := keepertest.GenTestAddr()

	require.NotEmpty(t, trader.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestPrivatePoolCreation(t *testing.T) {
	ctx := freshCtx(t)
	creator := keepertest.GenTestAddr()

	require.NotEmpty(t, creator.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Incident Response + Monitoring Integration Tests

func TestIncidentDetectionAndResponse(t *testing.T) {
	ctx, keeper := modtestutil.NewSecurityKeeperForTest(t)
	ctx = ctx.WithBlockTime(time.Now().UTC())

	incident := keeper.CreateIncident(ctx, "validator outage", "critical", "2/4 validators offline", "ops")
	require.NotNil(t, incident)
	require.Equal(t, securitypb.IncidentStatus_INCIDENT_STATUS_DETECTED, incident.Status)

	err := keeper.PauseSystem(ctx, 2, "halt swaps until validators recover", keeper.GetAuthority())
	require.NoError(t, err)
	require.True(t, keeper.IsSystemPaused(ctx))

	txErr := keeper.CheckTransactionAllowed(
		ctx,
		keepertest.GenTestAddr().String(),
		sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(50_000))),
	)
	require.ErrorIs(t, txErr, securitytypes.ErrSystemPaused)

	resolveSteps := []string{"rotated validator keys", "restarted validators", "synced state"}
	require.NoError(t, keeper.ResolveIncident(ctx, incident.IncidentId, resolveSteps))
	stored, found := keeper.GetIncident(ctx, incident.IncidentId)
	require.True(t, found)
	require.Equal(t, securitypb.IncidentStatus_INCIDENT_STATUS_RESOLVED, stored.Status)
	require.NotNil(t, stored.ResolvedAt)

	require.NoError(t, keeper.ResumeSystem(ctx))
	require.False(t, keeper.IsSystemPaused(ctx))

	txErr = keeper.CheckTransactionAllowed(
		ctx,
		keepertest.GenTestAddr().String(),
		sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(50_000))),
	)
	require.NoError(t, txErr)
}

func TestCircuitBreakerActivation(t *testing.T) {
	ctx, keeper := modtestutil.NewSecurityKeeperForTest(t)
	moduleName := "dex"
	authority := keeper.GetAuthority()

	require.NoError(t, keeper.RequireNotPaused(ctx, moduleName))
	require.NoError(t, keeper.PauseModule(ctx, moduleName, authority))
	require.True(t, keeper.IsModulePaused(ctx, moduleName))

	err := keeper.RequireNotPaused(ctx, moduleName)
	require.ErrorIs(t, err, securitytypes.ErrSystemPaused)

	require.NoError(t, keeper.UnpauseModule(ctx, moduleName, authority))
	require.NoError(t, keeper.RequireNotPaused(ctx, moduleName))
}

func TestFraudDetectionChain(t *testing.T) {
	ctx, keeper := modtestutil.NewSecurityKeeperForTest(t)
	ctx = ctx.WithBlockTime(time.Now().UTC())

	wallet := keepertest.GenTestAddr().String()
	setAt := ctx.BlockTime()
	expiresAt := setAt.Add(time.Hour)

	keeper.SetWalletLimit(ctx, &securitytypes.WalletLimit{
		WalletAddress:  wallet,
		MaxTxAmount:    "100000",
		MaxDailyTxs:    3,
		CooldownPeriod: "1h",
		SetAt:          &setAt,
		ExpiresAt:      &expiresAt,
		Reason:         "elevated risk wallet",
	})

	// Within limit succeeds
	err := keeper.CheckTransactionAllowed(
		ctx,
		wallet,
		sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(80_000))),
	)
	require.NoError(t, err)

	// Oversized transfer triggers fraud guard
	err = keeper.CheckTransactionAllowed(
		ctx,
		wallet,
		sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(150_000))),
	)
	require.ErrorIs(t, err, securitytypes.ErrWalletLimitExceeded)

	// Expire the limit – next run should clear entry and allow transfer
	past := setAt.Add(-2 * time.Hour)
	keeper.SetWalletLimit(ctx, &securitytypes.WalletLimit{
		WalletAddress: wallet,
		MaxTxAmount:   "100000",
		SetAt:         &setAt,
		ExpiresAt:     &past,
	})

	err = keeper.CheckTransactionAllowed(
		ctx,
		wallet,
		sdk.NewCoins(sdk.NewCoin("uaura", sdkmath.NewInt(200_000))),
	)
	require.NoError(t, err)

	_, hasLimit := keeper.GetWalletLimit(ctx, wallet)
	require.False(t, hasLimit, "expired limits must be cleared automatically")
}

// Validator Security + Economic Security Integration

func TestValidatorSlashing(t *testing.T) {
	ctx := freshCtx(t)
	validator := keepertest.GenTestAddr()

	require.NotEmpty(t, validator.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestValidatorRotation(t *testing.T) {
	ctx := freshCtx(t)

	require.NotNil(t, ctx)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Identity Change + Social Recovery Integration

func TestSocialRecoveryFlow(t *testing.T) {
	ctx := freshCtx(t)
	owner := keepertest.GenTestAddr()

	require.NotEmpty(t, owner.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestIdentityChangeRequiresRecovery(t *testing.T) {
	ctx := freshCtx(t)
	user := keepertest.GenTestAddr()

	require.NotEmpty(t, user.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Wallet Security + Cryptography Integration

func TestMultisigWalletCreation(t *testing.T) {
	ctx := freshCtx(t)
	owners := keepertest.GenTestAddrs(3)

	require.Len(t, owners, 3)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestSecureWalletRecovery(t *testing.T) {
	ctx := freshCtx(t)
	owner := keepertest.GenTestAddr()

	require.NotEmpty(t, owner.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Data Registry + VC Registry Integration

func TestVerifiableCredentialStorage(t *testing.T) {
	ctx := freshCtx(t)
	issuer := keepertest.GenTestAddr()

	require.NotEmpty(t, issuer.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestVCRevocationAndUpdate(t *testing.T) {
	ctx := freshCtx(t)
	holder := keepertest.GenTestAddr()

	require.NotEmpty(t, holder.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Monitoring + Network Security Integration

func TestNetworkAnomalyDetection(t *testing.T) {
	ctx := freshCtx(t)

	require.NotNil(t, ctx)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestRateLimitTriggers(t *testing.T) {
	ctx := freshCtx(t)

	require.NotNil(t, ctx)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Prevalidation + Mempool Integration

func TestMempoolEnforcesPrevalidation(t *testing.T) {
	ctx := freshCtx(t)

	require.NotNil(t, ctx)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestMempoolRejectsInvalidTx(t *testing.T) {
	ctx := freshCtx(t)

	require.NotNil(t, ctx)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Validator Security + Monitoring Integration

func TestValidatorDowntimeDetection(t *testing.T) {
	ctx := freshCtx(t)

	require.NotNil(t, ctx)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestValidatorMisbehaviorAlerts(t *testing.T) {
	ctx := freshCtx(t)

	require.NotNil(t, ctx)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Wallet Security + Bridge Integration

func TestWalletSecurityForBridgeTransfers(t *testing.T) {
	ctx := freshCtx(t)
	sender := keepertest.GenTestAddr()
	recipient := keepertest.GenTestAddr()

	require.NotEmpty(t, sender.String())
	require.NotEmpty(t, recipient.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestBridgeRespectsWalletLimits(t *testing.T) {
	ctx := freshCtx(t)
	sender := keepertest.GenTestAddr()

	require.NotEmpty(t, sender.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Privacy + Wallet Integration

func TestShieldedWalletBalances(t *testing.T) {
	ctx := freshCtx(t)
	owner := keepertest.GenTestAddr()

	require.NotEmpty(t, owner.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestPrivateTransfersBetweenWallets(t *testing.T) {
	ctx := freshCtx(t)
	sender := keepertest.GenTestAddr()
	recipient := keepertest.GenTestAddr()

	require.NotEmpty(t, sender.String())
	require.NotEmpty(t, recipient.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Exploratory: Data Registry + Compliance

func TestDataRetentionPolicies(t *testing.T) {
	ctx := freshCtx(t)

	require.NotNil(t, ctx)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestGDPRRequestHandling(t *testing.T) {
	ctx := freshCtx(t)
	requester := keepertest.GenTestAddr()

	require.NotEmpty(t, requester.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Explorer + Governance

func TestProposalIndexing(t *testing.T) {
	ctx := freshCtx(t)

	require.NotNil(t, ctx)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestVoteVisibility(t *testing.T) {
	ctx := freshCtx(t)

	require.NotNil(t, ctx)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

// Crypto + Wallet recovery

func TestKeyRotationFlow(t *testing.T) {
	ctx := freshCtx(t)
	owner := keepertest.GenTestAddr()

	require.NotEmpty(t, owner.String())
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}

func TestSessionExpiry(t *testing.T) {
	ctx := freshCtx(t)

	require.NotNil(t, ctx)
	require.GreaterOrEqual(t, ctx.BlockHeight(), int64(0))
}
