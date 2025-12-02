package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
)

// setupKeeper creates a test keeper and context for contractregistry module tests.
// This function is used by tests in the keeper package that expect a simple
// setupKeeper(t) signature returning (Keeper, sdk.Context).
func setupKeeper(t *testing.T) (Keeper, sdk.Context) {
	t.Helper()

	// Use global test helpers for standard setup
	input := keepertest.CreateTestInputWithTime(t, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	// Create keeper with authority address
	// Using a standard test governance address
	authority := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"

	keeper := NewKeeper(
		input.StoreKey,
		input.Cdc,
		authority,
	)

	// Initialize with default params
	params := types.DefaultParams()
	if err := keeper.SetParams(input.Ctx, &params); err != nil {
		panic(err)
	}

	return *keeper, input.Ctx
}

// setupKeeperWithMocks creates a keeper with mock dependencies for testing.
// This is useful for tests that need to verify interactions with other modules.
func setupKeeperWithMocks(t *testing.T) (Keeper, sdk.Context, *mockComplianceKeeper, *mockVCKeeper, *mockConfidenceScoreKeeper) {
	t.Helper()

	k, ctx := setupKeeper(t)

	// Create mocks
	mockCompliance := &mockComplianceKeeper{}
	mockVC := &mockVCKeeper{}
	mockCS := &mockConfidenceScoreKeeper{}

	// Set mocks on keeper
	k.SetComplianceKeeper(mockCompliance)
	k.SetVCKeeper(mockVC)
	k.SetConfidenceScoreKeeper(mockCS)

	return k, ctx, mockCompliance, mockVC, mockCS
}

// Mock keepers for testing
// These implement the minimal interface needed for contractregistry keeper tests

type mockComplianceKeeper struct {
	kycLevel   uint32
	sanctioned bool
	err        error
}

func (m *mockComplianceKeeper) GetKYCLevel(ctx sdk.Context, address string) (uint32, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.kycLevel, nil
}

func (m *mockComplianceKeeper) ScreenForSanctions(ctx sdk.Context, address string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.sanctioned, nil
}

func (m *mockComplianceKeeper) SetKYCLevel(level uint32) {
	m.kycLevel = level
}

func (m *mockComplianceKeeper) SetSanctioned(sanctioned bool) {
	m.sanctioned = sanctioned
}

func (m *mockComplianceKeeper) SetError(err error) {
	m.err = err
}

type mockVCKeeper struct {
	hasVC bool
}

func (m *mockVCKeeper) HasVC(ctx interface{}, address string, vcType string) bool {
	return m.hasVC
}

func (m *mockVCKeeper) SetHasVC(has bool) {
	m.hasVC = has
}

type mockConfidenceScoreKeeper struct {
	score uint64
	found bool
}

func (m *mockConfidenceScoreKeeper) GetUserScore(address string) (uint64, bool) {
	return m.score, m.found
}

func (m *mockConfidenceScoreKeeper) SetScore(score uint64) {
	m.score = score
	m.found = true
}

// Helper functions for creating test data

// newTestContractInfo creates a ContractInfo for testing with sensible defaults.
func newTestContractInfo(address, creator, admin string) *pb.ContractInfo {
	return &pb.ContractInfo{
		Address: address,
		CodeId:  1,
		Creator: creator,
		Admin:   admin,
		Label:   "test-contract",
		Metadata: &pb.ContractMetadata{
			Name:        "Test Contract",
			Description: "A test contract",
			Version:     "1.0.0",
			Tags:        []string{"test"},
		},
		SecurityPolicy: &pb.SecurityPolicy{
			AllowPause:       true,
			MaxGasPerTx:      1000000,
			RateLimitPerUser: 100,
		},
		Compliance: &pb.ComplianceRequirements{
			EnforceKyc: false,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
}

// advanceBlockTime advances the block time in the context by the given duration.
// This is a module-specific wrapper over the global helper.
func advanceBlockTime(ctx sdk.Context, duration time.Duration) sdk.Context {
	return keepertest.AdvanceTime(ctx, duration).WithBlockHeight(ctx.BlockHeight() + 1)
}

// advanceBlockHeight advances the block height by the given number of blocks.
// This is a module-specific wrapper over the global helper.
func advanceBlockHeight(ctx sdk.Context, blocks int64) sdk.Context {
	return keepertest.AdvanceBlockHeight(ctx, blocks)
}
