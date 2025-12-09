package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
)

// setupKeeper creates a test keeper and context for contractregistry module tests.
// This function is used by tests in the keeper package that expect a simple
// setupKeeper(t) signature returning (Keeper, sdk.Context).
// NOTE: Cannot use global helpers due to import cycle (testutil/keeper -> wasm -> contractregistry)
func setupKeeper(t *testing.T) (Keeper, sdk.Context) {
	t.Helper()

	// Create in-memory database
	db := dbm.NewMemDB()

	// Create commit multi-store with metrics
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	// Create store key for contractregistry module
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)

	// Load latest version
	if err := cms.LoadLatestVersion(); err != nil {
		panic(err)
	}

	// Create context with block header and time
	ctx := sdk.NewContext(
		cms,
		cmtproto.Header{
			Height: 1,
			Time:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		false,
		log.NewNopLogger(),
	)

	// Create interface registry and codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create keeper with authority address
	// Using a standard test governance address
	authority := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"

	keeper := NewKeeper(
		storeKey,
		cdc,
		authority,
	)

	// Initialize with default params
	params := types.DefaultParams()
	if err := keeper.SetParams(ctx, &params); err != nil {
		panic(err)
	}

	return *keeper, ctx
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

func (m *mockConfidenceScoreKeeper) GetUserScore(ctx sdk.Context, address string) (uint64, bool) {
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
		Metadata: pb.ContractMetadata{
			Name:        "Test Contract",
			Description: "A test contract",
			Version:     "1.0.0",
			Tags:        []string{"test"},
		},
		SecurityPolicy: pb.SecurityPolicy{
			AllowPause:       true,
			MaxGasPerTx:      1000000,
			RateLimitPerUser: 100,
		},
		Compliance: pb.ComplianceRequirements{
			EnforceKyc: false,
		},
		Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
}

// advanceBlockTime advances the block time in the context by the given duration.
func advanceBlockTime(ctx sdk.Context, duration time.Duration) sdk.Context {
	header := ctx.BlockHeader()
	header.Time = header.Time.Add(duration)
	header.Height++
	return ctx.WithBlockHeader(header)
}

// advanceBlockHeight advances the block height by the given number of blocks.
func advanceBlockHeight(ctx sdk.Context, blocks int64) sdk.Context {
	header := ctx.BlockHeader()
	header.Height += blocks
	// Assume 5 seconds per block
	header.Time = header.Time.Add(time.Duration(blocks) * 5 * time.Second)
	return ctx.WithBlockHeader(header)
}
