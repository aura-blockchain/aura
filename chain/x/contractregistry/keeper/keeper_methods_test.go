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
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestIsContractRegistered tests the IsContractRegistered method
func TestIsContractRegistered(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	k := keeper.NewKeeper(key, cdc, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")

	// Test with non-existent contract
	exists := k.IsContractRegistered(ctx, "cosmos1nonexistent")
	require.False(t, exists, "Non-existent contract should return false")

	// Register a contract
	info := &pb.ContractInfo{
		Address:   "cosmos1contract",
		CodeId:    1,
		Creator:   "cosmos1creator",
		Admin:     "cosmos1admin",
		Label:     "test-contract",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Status:    pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	k.SetContractInfo(ctx, info)

	// Test with existing contract
	exists = k.IsContractRegistered(ctx, "cosmos1contract")
	require.True(t, exists, "Existing contract should return true")
}

// TestGetCreatorContracts tests the GetCreatorContracts method
func TestGetCreatorContracts(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	k := keeper.NewKeeper(key, cdc, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")

	creator := "cosmos1creator"

	// Test with no contracts
	contracts := k.GetCreatorContracts(ctx, creator)
	require.Empty(t, contracts, "Should return empty slice when no contracts exist")

	// Add contracts
	for i := 1; i <= 3; i++ {
		contractAddr := sdk.AccAddress([]byte("contract" + string(rune('0'+i)))).String()
		info := &pb.ContractInfo{
			Address:   contractAddr,
			CodeId:    uint64(i),
			Creator:   creator,
			Admin:     creator,
			Label:     "test",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Status:    pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
		}
		k.SetContractInfo(ctx, info)
		k.AddCreatorContract(ctx, creator, contractAddr)
	}

	// Test retrieval
	contracts = k.GetCreatorContracts(ctx, creator)
	require.Len(t, contracts, 3, "Should return 3 contracts for creator")

	// Verify all contracts have the correct creator
	for _, contract := range contracts {
		require.Equal(t, creator, contract.Creator, "All contracts should have correct creator")
	}
}

// TestGetTagContracts tests the GetTagContracts method
func TestGetTagContracts(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	k := keeper.NewKeeper(key, cdc, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")

	tag := "defi"

	// Test with no contracts
	contracts := k.GetTagContracts(ctx, tag)
	require.Empty(t, contracts, "Should return empty slice when no contracts exist")

	// Add contracts with the tag
	for i := 1; i <= 2; i++ {
		contractAddr := sdk.AccAddress([]byte("contract" + string(rune('0'+i)))).String()
		info := &pb.ContractInfo{
			Address:   contractAddr,
			CodeId:    uint64(i),
			Creator:   "cosmos1creator",
			Admin:     "cosmos1admin",
			Label:     "test",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Metadata: &pb.ContractMetadata{
				Name:        "Test Contract",
				Description: "A test contract",
				Version:     "1.0.0",
				Tags:        []string{tag},
			},
			Status: pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
		}
		k.SetContractInfo(ctx, info)
		k.AddTagContract(ctx, tag, contractAddr)
	}

	// Test retrieval
	contracts = k.GetTagContracts(ctx, tag)
	require.Len(t, contracts, 2, "Should return 2 contracts with tag")

	// Verify all contracts have the tag in metadata
	for _, contract := range contracts {
		require.NotNil(t, contract.Metadata, "Contract metadata should not be nil")
		require.Contains(t, contract.Metadata.Tags, tag, "Contract should have the tag")
	}
}

// TestGetAllContracts tests the GetAllContracts method
func TestGetAllContracts(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	k := keeper.NewKeeper(key, cdc, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")

	// Test with no contracts
	contracts := k.GetAllContracts(ctx)
	require.Empty(t, contracts, "Should return empty slice when no contracts exist")

	// Add multiple contracts
	for i := 1; i <= 5; i++ {
		contractAddr := sdk.AccAddress([]byte("contract" + string(rune('0'+i)))).String()
		info := &pb.ContractInfo{
			Address:   contractAddr,
			CodeId:    uint64(i),
			Creator:   "cosmos1creator",
			Admin:     "cosmos1admin",
			Label:     "test",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			Status:    pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
		}
		k.SetContractInfo(ctx, info)
	}

	// Test retrieval
	contracts = k.GetAllContracts(ctx)
	require.Len(t, contracts, 5, "Should return all 5 contracts")
}

// TestDeleteContractInfo tests the DeleteContractInfo method
func TestDeleteContractInfo(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	k := keeper.NewKeeper(key, cdc, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")

	contractAddr := "cosmos1contract"

	// Add a contract
	info := &pb.ContractInfo{
		Address:   contractAddr,
		CodeId:    1,
		Creator:   "cosmos1creator",
		Admin:     "cosmos1admin",
		Label:     "test-contract",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
		Status:    pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
	}
	k.SetContractInfo(ctx, info)

	// Verify it exists
	exists := k.IsContractRegistered(ctx, contractAddr)
	require.True(t, exists, "Contract should exist after creation")

	// Delete the contract
	k.DeleteContractInfo(ctx, contractAddr)

	// Verify it's deleted
	exists = k.IsContractRegistered(ctx, contractAddr)
	require.False(t, exists, "Contract should not exist after deletion")
}
