package integration

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	contractregistrykeeper "github.com/aequitas/aura/chain/x/contractregistry/keeper"
	contractregistrytypes "github.com/aequitas/aura/chain/x/contractregistry/types"
	wasmkeeper "github.com/aequitas/aura/chain/x/wasm/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	contractpb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// WASMTestContext creates a fully wired in-memory Cosmos-SDK context with the
// Aura WASM security keeper and contract registry keeper. It mirrors the
// production wiring so integration tests can exercise the real hooks.
type WASMTestContext struct {
	T *testing.T

	Ctx             sdk.Context
	WasmKeeper      wasmkeeper.Keeper
	RegistryKeeper  *contractregistrykeeper.Keeper
	complianceMock  *mockComplianceKeeper
	vcMock          *mockVCKeeper
	confidenceMock  *mockConfidenceScoreKeeper
	nextCodeID      uint64
	nextContractNum uint64
}

const (
	testAuthorityAddress = "aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
)

// SetupTestAppWithWasm constructs WASM + contract registry keepers backed by a
// real KV store so integration tests can run the full hook logic.
func SetupTestAppWithWasm(t *testing.T) WASMTestContext {
	t.Helper()

	// Build codec shared by both keepers.
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	// Create stores for wasm + contract registry.
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	wasmStoreKey := storetypes.NewKVStoreKey(types.StoreKey)
	registryStoreKey := storetypes.NewKVStoreKey(contractregistrytypes.StoreKey)

	cms.MountStoreWithDB(wasmStoreKey, storetypes.StoreTypeIAVL, db)
	cms.MountStoreWithDB(registryStoreKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, cms.LoadLatestVersion())

	ctx := sdk.NewContext(
		cms,
		tmproto.Header{
			Height: 1,
			Time:   time.Now().UTC(),
		},
		false,
		log.NewNopLogger(),
	).WithGasMeter(storetypes.NewGasMeter(500_000))

	// Contract registry keeper with default params + mocks wired in.
	registryKeeper := contractregistrykeeper.NewKeeper(
		registryStoreKey,
		cdc,
		testAuthorityAddress,
	)
	require.NoError(t, registryKeeper.SetParams(ctx, contractregistrytypes.DefaultParams()))

	complianceMock := newMockComplianceKeeper()
	vcMock := newMockVCKeeper()
	confidenceMock := newMockConfidenceScoreKeeper()
	registryKeeper.SetComplianceKeeper(complianceMock)
	registryKeeper.SetVCKeeper(vcMock)
	registryKeeper.SetConfidenceScoreKeeper(confidenceMock)

	// WASM keeper configured with the registry hook (wasmd keeper mocked out).
	wasmKeeper := wasmkeeper.NewKeeper(
		cdc,
		wasmStoreKey,
		nil,
		testAuthorityAddress,
	)
	wasmKeeper.SetContractRegistry(registryKeeper)
	wasmKeeper.ResetCircuitBreaker(ctx)

	return WASMTestContext{
		T:              t,
		Ctx:            ctx,
		WasmKeeper:     wasmKeeper,
		RegistryKeeper: registryKeeper,
		complianceMock: complianceMock,
		vcMock:         vcMock,
		confidenceMock: confidenceMock,
		nextCodeID:     0,
	}
}

// GetContext returns the underlying sdk.Context for helpers that match the
// legacy README expectations.
func (w WASMTestContext) GetContext() sdk.Context {
	return w.Ctx
}

// AdvanceBlock increments the block height and time. Useful when exercising
// rate limits.
func (w *WASMTestContext) AdvanceBlock(duration time.Duration) {
	header := w.Ctx.BlockHeader()
	header.Height++
	header.Time = header.Time.Add(duration)
	w.Ctx = w.Ctx.WithBlockHeader(header)
}

// CreateAuthorizedUploader creates an account and authorizes it for uploads.
func (w *WASMTestContext) CreateAuthorizedUploader() sdk.AccAddress {
	addr := w.newAddress("uploader")
	err := w.WasmKeeper.AuthorizeUploader(w.Ctx, addr.String())
	require.NoError(w.T, err)
	return addr
}

// CreateMockWASMCode returns deterministic WASM bytes placeholder.
func (w *WASMTestContext) CreateMockWASMCode() []byte {
	return []byte(fmt.Sprintf("wasm-code-%d", time.Now().UnixNano()))
}

// UploadTestContract increments the code counter and returns a synthetic code ID.
func (w *WASMTestContext) UploadTestContract(_ interface{}) uint64 {
	w.nextCodeID++
	return w.nextCodeID
}

// SetupCompleteContract instantiates and auto-registers a contract, then allows
// callers to optionally override metadata/policies.
func (w *WASMTestContext) SetupCompleteContract(
	uploader sdk.AccAddress,
	metadata *contractpb.ContractMetadata,
	policy *contractpb.SecurityPolicy,
	compliance *contractpb.ComplianceRequirements,
) sdk.AccAddress {
	if uploader.Empty() {
		uploader = w.CreateAuthorizedUploader()
	}
	admin := uploader
	codeID := w.UploadTestContract(nil)
	label := fmt.Sprintf("contract-%d", codeID)

	require.NoError(w.T, w.WasmKeeper.BeforeInstantiateHook(
		w.Ctx,
		codeID,
		uploader,
		admin,
		label,
	))

	contractAddr := w.newContractAddress()
	require.NoError(w.T, w.WasmKeeper.AfterInstantiateHook(
		w.Ctx,
		contractAddr,
		codeID,
		uploader,
		admin,
		label,
	))

	// Override metadata/security/compliance if provided.
	if metadata != nil || policy != nil || compliance != nil {
		info, found := w.RegistryKeeper.GetContractInfo(w.Ctx, contractAddr.String())
		require.True(w.T, found, "contract must exist in registry")
		if metadata != nil {
			info.Metadata = *metadata
		}
		if policy != nil {
			info.SecurityPolicy = *policy
		}
		if compliance != nil {
			info.Compliance = *compliance
		}
		w.RegistryKeeper.SetContractInfo(w.Ctx, &info)
	}

	return contractAddr
}

// SetupCompleteContractWithPolicy is a convenience wrapper for readability in tests.
func (w *WASMTestContext) SetupCompleteContractWithPolicy(
	uploader sdk.AccAddress,
	metadata contractpb.ContractMetadata,
	policy contractpb.SecurityPolicy,
	compliance contractpb.ComplianceRequirements,
) sdk.AccAddress {
	return w.SetupCompleteContract(uploader, &metadata, &policy, &compliance)
}

// ExecuteAsUser runs the before/after execution hooks for a user. The caller can
// choose whether the execution should be recorded as success.
func (w *WASMTestContext) ExecuteAsUser(contractAddr sdk.AccAddress, user sdk.AccAddress, success bool, execErr error) error {
	gasBefore := w.Ctx.GasMeter().GasConsumed()
	if err := w.WasmKeeper.BeforeExecuteHook(w.Ctx, contractAddr, user); err != nil {
		return err
	}
	w.WasmKeeper.AfterExecuteHook(w.Ctx, contractAddr, gasBefore, success, execErr)
	return nil
}

// SetUserKYC configures the mock compliance keeper with a KYC level.
func (w *WASMTestContext) SetUserKYC(address sdk.AccAddress, level uint32) {
	w.complianceMock.kycLevels[address.String()] = level
}

// SetUserSanction toggles sanction status.
func (w *WASMTestContext) SetUserSanction(address sdk.AccAddress, sanctioned bool) {
	w.complianceMock.sanctions[address.String()] = sanctioned
}

// GrantUserVC grants a VC type to a user.
func (w *WASMTestContext) GrantUserVC(address sdk.AccAddress, vcType string) {
	w.vcMock.grant(address.String(), vcType)
}

// SetUserConfidence sets the confidence score for a user.
func (w *WASMTestContext) SetUserConfidence(address sdk.AccAddress, score uint64) {
	w.confidenceMock.scores[address.String()] = score
}

func (w *WASMTestContext) newContractAddress() sdk.AccAddress {
	w.nextContractNum++
	return sdk.AccAddress([]byte(fmt.Sprintf("contract-%d-addr", w.nextContractNum)))
}

func (w *WASMTestContext) newAddress(prefix string) sdk.AccAddress {
	w.nextContractNum++
	return sdk.AccAddress([]byte(fmt.Sprintf("%s-%d-addr", prefix, w.nextContractNum)))
}

// mockComplianceKeeper implements the compliance interface with simple maps.
type mockComplianceKeeper struct {
	kycLevels map[string]uint32
	sanctions map[string]bool
}

func newMockComplianceKeeper() *mockComplianceKeeper {
	return &mockComplianceKeeper{
		kycLevels: make(map[string]uint32),
		sanctions: make(map[string]bool),
	}
}

func (m *mockComplianceKeeper) GetKYCLevel(_ sdk.Context, address string) (uint32, error) {
	return m.kycLevels[address], nil
}

func (m *mockComplianceKeeper) ScreenForSanctions(_ sdk.Context, address string) (bool, error) {
	return m.sanctions[address], nil
}

// mockVCKeeper tracks which VC types each address holds.
type mockVCKeeper struct {
	records map[string]map[string]bool
}

func newMockVCKeeper() *mockVCKeeper {
	return &mockVCKeeper{
		records: make(map[string]map[string]bool),
	}
}

func (m *mockVCKeeper) grant(address, vcType string) {
	if _, ok := m.records[address]; !ok {
		m.records[address] = make(map[string]bool)
	}
	m.records[address][vcType] = true
}

func (m *mockVCKeeper) HasVC(_ interface{}, address string, vcType string) bool {
	if _, ok := m.records[address]; !ok {
		return false
	}
	return m.records[address][vcType]
}

// mockConfidenceScoreKeeper stores deterministic scores per address.
type mockConfidenceScoreKeeper struct {
	scores map[string]uint64
}

func newMockConfidenceScoreKeeper() *mockConfidenceScoreKeeper {
	return &mockConfidenceScoreKeeper{
		scores: make(map[string]uint64),
	}
}

func (m *mockConfidenceScoreKeeper) GetUserScore(_ sdk.Context, address string) (uint64, bool) {
	score, ok := m.scores[address]
	return score, ok
}
