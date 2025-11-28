package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/privacy/keeper"
	"github.com/aequitas/aura/chain/x/privacy/types"
)

// MockAccountKeeper implements types.AccountKeeper for testing
type MockAccountKeeper struct{}

func (m MockAccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}

func (m MockAccountKeeper) SetAccount(ctx context.Context, acc sdk.AccountI) {}

func (m MockAccountKeeper) NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}

// MockBankKeeper implements types.BankKeeper for testing
type MockBankKeeper struct{}

func (m MockBankKeeper) SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

func (m MockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m MockBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	return nil
}

func (m MockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return sdk.NewInt64Coin(denom, 0)
}

// MockZKProofSystem implements types.ZKProofSystem for testing
type MockZKProofSystem struct{}

func (m MockZKProofSystem) VerifyProof(proof []byte, publicInputs []byte, verificationKey []byte) (bool, error) {
	return true, nil
}

func (m MockZKProofSystem) GenerateProof(witness []byte, publicInputs []byte, provingKey []byte) ([]byte, error) {
	return []byte("mock_proof"), nil
}

func (m MockZKProofSystem) VerifyRangeProof(commitment []byte, proof []byte, min, max uint64) (bool, error) {
	return true, nil
}

// MockMixingService implements types.MixingService for testing
type MockMixingService struct{}

func (m MockMixingService) CreateMixingPool(denomination string, minParticipants uint32) (string, error) {
	return "mock_pool_id", nil
}

func (m MockMixingService) JoinMixingPool(poolID string, participant string, commitment []byte) error {
	return nil
}

func (m MockMixingService) ExecuteMixing(poolID string) error {
	return nil
}

func (m MockMixingService) GetPoolStatus(poolID string) (string, error) {
	return "active", nil
}

// MockViewKeyManager implements types.ViewKeyManager for testing
type MockViewKeyManager struct{}

func (m MockViewKeyManager) GenerateViewKey(address string) ([]byte, error) {
	return []byte("mock_view_key"), nil
}

func (m MockViewKeyManager) StoreViewKey(address string, viewKey []byte) error {
	return nil
}

func (m MockViewKeyManager) GetViewKey(address string) ([]byte, bool) {
	return []byte("mock_view_key"), true
}

func (m MockViewKeyManager) DecryptWithViewKey(encryptedData []byte, viewKey []byte) ([]byte, error) {
	return []byte("decrypted"), nil
}

// MockNetworkPrivacy implements types.NetworkPrivacy for testing
type MockNetworkPrivacy struct{}

func (m MockNetworkPrivacy) ObfuscateTransaction(txData []byte) ([]byte, error) {
	return txData, nil
}

func (m MockNetworkPrivacy) RoutePrivately(destination string, data []byte) error {
	return nil
}

func (m MockNetworkPrivacy) GetPrivacyMetrics() map[string]interface{} {
	return map[string]interface{}{"test": "metrics"}
}

// MockMemoEncryptor implements types.MemoEncryptor for testing
type MockMemoEncryptor struct{}

func (m MockMemoEncryptor) EncryptMemo(memo string, recipientPubKey []byte) ([]byte, error) {
	return []byte("encrypted"), nil
}

func (m MockMemoEncryptor) DecryptMemo(encryptedMemo []byte, privateKey []byte) (string, error) {
	return "decrypted", nil
}

func (m MockMemoEncryptor) VerifyEncryptedMemo(encryptedMemo []byte) bool {
	return true
}

// setupTestKeeper creates a test keeper with mock dependencies
func setupTestKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	// Setup store
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	// Setup codec
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	// Create keeper with properly typed mock keepers
	authKeeper := MockAccountKeeper{}
	bankKeeper := MockBankKeeper{}

	k := keeper.NewKeeper(cdc, storeKey, authKeeper, bankKeeper)

	// Create context
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	return k, ctx
}

// TestKeeperTypeSafety verifies that the keeper is type-safe
func TestKeeperTypeSafety(t *testing.T) {
	k, ctx := setupTestKeeper(t)
	require.NotNil(t, k)
	require.NotNil(t, ctx)

	t.Run("Keeper created with proper types", func(t *testing.T) {
		// This test verifies that the keeper can be created with properly typed keepers
		// If there were any type mismatches, the code wouldn't compile
		require.NotNil(t, k)
	})

	t.Run("Optional dependencies can be set with proper types", func(t *testing.T) {
		// Test setting ZKProofSystem
		zkProof := MockZKProofSystem{}
		k.SetZKProofSystem(zkProof)

		// Test setting MixingService
		mixing := MockMixingService{}
		k.SetMixingService(mixing)

		// Test setting ViewKeyManager
		viewKey := MockViewKeyManager{}
		k.SetViewKeyManager(viewKey)

		// Test setting NetworkPrivacy
		network := MockNetworkPrivacy{}
		k.SetNetworkPrivacyService(network)

		// Test setting MemoEncryptor
		memo := MockMemoEncryptor{}
		k.SetMemoEncryptor(memo)
	})
}

// TestKeeperInterfaceCompliance verifies all mock keepers satisfy their interfaces
func TestKeeperInterfaceCompliance(t *testing.T) {
	t.Run("AccountKeeper interface compliance", func(t *testing.T) {
		var _ types.AccountKeeper = MockAccountKeeper{}
	})

	t.Run("BankKeeper interface compliance", func(t *testing.T) {
		var _ types.BankKeeper = MockBankKeeper{}
	})

	t.Run("ZKProofSystem interface compliance", func(t *testing.T) {
		var _ types.ZKProofSystem = MockZKProofSystem{}
	})

	t.Run("MixingService interface compliance", func(t *testing.T) {
		var _ types.MixingService = MockMixingService{}
	})

	t.Run("ViewKeyManager interface compliance", func(t *testing.T) {
		var _ types.ViewKeyManager = MockViewKeyManager{}
	})

	t.Run("NetworkPrivacy interface compliance", func(t *testing.T) {
		var _ types.NetworkPrivacy = MockNetworkPrivacy{}
	})

	t.Run("MemoEncryptor interface compliance", func(t *testing.T) {
		var _ types.MemoEncryptor = MockMemoEncryptor{}
	})
}

// TestKeeperBasicOperations verifies basic keeper operations with typed dependencies
func TestKeeperBasicOperations(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	t.Run("Set and get params", func(t *testing.T) {
		params := types.DefaultParams()
		err := k.SetParams(ctx, params)
		require.NoError(t, err)

		retrieved := k.GetParams(ctx)
		require.Equal(t, params.EnableZkProofs, retrieved.EnableZkProofs)
		require.Equal(t, params.MinRingSize, retrieved.MinRingSize)
	})

	t.Run("Create commitment", func(t *testing.T) {
		commitment := []byte("test_commitment")
		id, err := k.CreateCommitment(ctx, "test_sender", commitment)
		require.NoError(t, err)
		require.NotEmpty(t, id)

		// Verify commitment can be retrieved
		record, found := k.GetCommitment(ctx, id)
		require.True(t, found)
		require.NotNil(t, record)
		require.Equal(t, id, record.ID)
	})

	t.Run("Nullifier operations", func(t *testing.T) {
		nullifier := []byte("test_nullifier")

		// Nullifier should not exist initially
		exists := k.NullifierExists(ctx, nullifier)
		require.False(t, exists)

		// Create nullifier
		err := k.CreateNullifier(ctx, nullifier)
		require.NoError(t, err)

		// Nullifier should now exist
		exists = k.NullifierExists(ctx, nullifier)
		require.True(t, exists)

		// Creating again should fail
		err = k.CreateNullifier(ctx, nullifier)
		require.Error(t, err)
	})

	t.Run("Merkle tree operations", func(t *testing.T) {
		leaf := []byte("test_leaf")
		index, err := k.AddLeaf(ctx, leaf)
		require.NoError(t, err)
		require.Equal(t, uint64(0), index)

		// Retrieve leaf
		retrieved, found := k.GetLeaf(ctx, index)
		require.True(t, found)
		require.Equal(t, leaf, retrieved)

		// Get merkle root
		root := k.GetMerkleRoot(ctx)
		require.NotNil(t, root)
	})
}

// TestTypeSafetyPreventsMisuse demonstrates compile-time type safety
func TestTypeSafetyPreventsMisuse(t *testing.T) {
	// This test demonstrates that the following would NOT compile:
	//
	// wrongKeeper := keeper.NewKeeper(cdc, storeKey, "string", 123)
	//                                                    ^^^^^^  ^^^
	//                                                    These would cause compile errors
	//
	// The keeper now requires properly typed AccountKeeper and BankKeeper interfaces,
	// preventing runtime type errors and providing compile-time guarantees.

	t.Run("Proper types are enforced at compile time", func(t *testing.T) {
		// If this test compiles, it means our type safety is working
		authKeeper := MockAccountKeeper{}
		bankKeeper := MockBankKeeper{}

		// These are the only valid types that can be passed
		var _ types.AccountKeeper = authKeeper
		var _ types.BankKeeper = bankKeeper

		// The following would NOT compile (uncomment to verify):
		// var _ types.AccountKeeper = "string"  // Compile error!
		// var _ types.BankKeeper = 123          // Compile error!
	})
}

// TestNilSafetyForOptionalDependencies verifies nil safety for optional dependencies
func TestNilSafetyForOptionalDependencies(t *testing.T) {
	k, ctx := setupTestKeeper(t)

	t.Run("Keeper works without optional dependencies", func(t *testing.T) {
		// The keeper should work for basic operations without optional dependencies
		// (zkProofSystem, mixingService, etc. can be nil)

		params := types.DefaultParams()
		err := k.SetParams(ctx, params)
		require.NoError(t, err)

		commitment := []byte("test")
		id, err := k.CreateCommitment(ctx, "sender", commitment)
		require.NoError(t, err)
		require.NotEmpty(t, id)
	})

	t.Run("Optional dependencies can be added later", func(t *testing.T) {
		// Verify that optional dependencies can be set after keeper creation
		k.SetZKProofSystem(MockZKProofSystem{})
		k.SetMixingService(MockMixingService{})
		k.SetViewKeyManager(MockViewKeyManager{})
		k.SetNetworkPrivacyService(MockNetworkPrivacy{})
		k.SetMemoEncryptor(MockMemoEncryptor{})

		// Keeper should still work after setting optional dependencies
		params := k.GetParams(ctx)
		require.NotNil(t, params)
	})
}
