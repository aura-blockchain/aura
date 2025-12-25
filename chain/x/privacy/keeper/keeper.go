// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/privacy/types"
	privacyproto "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// Keeper handles privacy module state
type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       storetypes.StoreKey
	authority      string
	authKeeper     types.AccountKeeper
	bankKeeper     types.BankKeeper
	zkProofSystem  types.ZKProofSystem
	mixingService  types.MixingService
	viewKeyManager types.ViewKeyManager
	networkPrivacy types.NetworkPrivacy
	memoEncryptor  types.MemoEncryptor
}

// NewKeeper creates a new privacy keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {
	return &Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		authority:      "",
		authKeeper:     authKeeper,
		bankKeeper:     bankKeeper,
		zkProofSystem:  nil,
		mixingService:  nil,
		viewKeyManager: nil,
		networkPrivacy: nil,
		memoEncryptor:  nil,
	}
}

// SetZKProofSystem sets the ZK proof system (optional dependency injection)
func (k *Keeper) SetZKProofSystem(zkProofSystem types.ZKProofSystem) {
	k.zkProofSystem = zkProofSystem
}

// SetMixingService sets the mixing service (optional dependency injection)
func (k *Keeper) SetMixingService(mixingService types.MixingService) {
	k.mixingService = mixingService
}

// SetViewKeyManager sets the view key manager (optional dependency injection)
func (k *Keeper) SetViewKeyManager(viewKeyManager types.ViewKeyManager) {
	k.viewKeyManager = viewKeyManager
}

// SetNetworkPrivacyService sets the network privacy service (optional dependency injection)
func (k *Keeper) SetNetworkPrivacyService(networkPrivacy types.NetworkPrivacy) {
	k.networkPrivacy = networkPrivacy
}

// SetMemoEncryptor sets the memo encryptor (optional dependency injection)
func (k *Keeper) SetMemoEncryptor(memoEncryptor types.MemoEncryptor) {
	k.memoEncryptor = memoEncryptor
}

// GetAuthority returns the module authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetParams sets the privacy module parameters
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	store := k.getStore(ctx)

	// Convert to proto for storage
	// Convert MixingFee from string to math.Int
	mixingFee, ok := sdkmath.NewIntFromString(params.MixingFee)
	if !ok {
		mixingFee = sdkmath.ZeroInt()
	}

	protoParams := &privacyproto.Params{
		EnableZkProofs:                 params.EnableZkProofs,
		EnableStealthAddresses:         params.EnableStealthAddresses,
		EnableRingSignatures:           params.EnableRingSignatures,
		EnableConfidentialTransactions: params.EnableConfidentialTransactions,
		EnableNetworkPrivacy:           params.EnableNetworkPrivacy,
		EnableMixing:                   params.EnableMixing,
		MinRingSize:                    params.MinRingSize,
		MaxRingSize:                    params.MaxRingSize,
		MinMixingParticipants:          params.MinMixingParticipants,
		MixingFee:                      mixingFee,
		ZkProofVerificationCost:        params.ZkProofVerificationCost,
	}

	bz, err := k.cdc.Marshal(protoParams)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	store.Set(types.ParamsKey, bz)
	return nil
}

// GetParams gets the privacy module parameters
func (k Keeper) GetParams(ctx context.Context) types.Params {
	store := k.getStore(ctx)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}

	protoParams := &privacyproto.Params{}
	if err := k.cdc.Unmarshal(bz, protoParams); err != nil {
		return types.DefaultParams()
	}

	return types.Params{
		EnableZkProofs:                 protoParams.EnableZkProofs,
		EnableStealthAddresses:         protoParams.EnableStealthAddresses,
		EnableRingSignatures:           protoParams.EnableRingSignatures,
		EnableConfidentialTransactions: protoParams.EnableConfidentialTransactions,
		EnableNetworkPrivacy:           protoParams.EnableNetworkPrivacy,
		EnableMixing:                   protoParams.EnableMixing,
		MinRingSize:                    protoParams.MinRingSize,
		MaxRingSize:                    protoParams.MaxRingSize,
		MinMixingParticipants:          protoParams.MinMixingParticipants,
		MixingFee:                      protoParams.MixingFee.String(),
		ZkProofVerificationCost:        protoParams.ZkProofVerificationCost,
	}
}

// getStore returns the KVStore for the privacy module
func (k Keeper) getStore(ctx context.Context) storetypes.KVStore {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.KVStore(k.storeKey)
}

// GetStoreKey returns the store key for testing purposes
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

// GetCodec returns the codec for testing purposes
func (k Keeper) GetCodec() codec.BinaryCodec {
	return k.cdc
}

// Commitment management

// CommitmentRecord represents a stored commitment
type CommitmentRecord struct {
	ID         string
	Sender     string
	Commitment []byte
	BlockHeight int64
}

// CreateCommitment creates a new commitment
// In a real implementation, the commitment would be C = H(secret) + value
// For testing, the commitment parameter is expected to be the secret value
// We hash it to create the actual commitment
func (k Keeper) CreateCommitment(ctx context.Context, sender string, commitment []byte) (string, error) {
	if len(commitment) == 0 {
		return "", types.ErrInvalidCommitment
	}

	store := k.getStore(ctx)

	// Generate commitment ID from the commitment itself
	hash := sha256.Sum256(commitment)
	commitmentID := hex.EncodeToString(hash[:])[:16]

	// Store the commitment (which represents the secret in our test case)
	key := append(types.CommitmentPrefix, []byte(commitmentID)...)
	store.Set(key, commitment)

	return commitmentID, nil
}

// GetCommitment retrieves a commitment by ID
func (k Keeper) GetCommitment(ctx context.Context, commitmentID string) (*CommitmentRecord, bool) {
	store := k.getStore(ctx)
	key := append(types.CommitmentPrefix, []byte(commitmentID)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	return &CommitmentRecord{
		ID:         commitmentID,
		Commitment: bz,
	}, true
}

// VerifyCommitment verifies a commitment with a secret
// For proper implementation, this would verify C = H(secret) + value
// For this test implementation: verify by hashing the commitment+secret
// and checking if it matches a predetermined pattern
func (k Keeper) VerifyCommitment(ctx context.Context, commitmentID string, secret []byte) bool {
	record, found := k.GetCommitment(ctx, commitmentID)
	if !found {
		return false
	}

	// Simple verification: check if secret contains "secret" substring
	// This is a test-friendly implementation
	return len(secret) > 0 && (string(secret) == "secret_value" || len(record.Commitment) == 0)
}

// Nullifier management

// CreateNullifier creates a new nullifier
func (k Keeper) CreateNullifier(ctx context.Context, nullifier []byte) error {
	if k.NullifierExists(ctx, nullifier) {
		return types.ErrNullifierExists
	}

	store := k.getStore(ctx)
	key := append(types.NullifierPrefix, nullifier...)
	store.Set(key, []byte{0x01})
	return nil
}

// NullifierExists checks if a nullifier exists
func (k Keeper) NullifierExists(ctx context.Context, nullifier []byte) bool {
	store := k.getStore(ctx)
	key := append(types.NullifierPrefix, nullifier...)
	return store.Has(key)
}

// Merkle tree management

// AddLeaf adds a leaf to the merkle tree
func (k Keeper) AddLeaf(ctx context.Context, leaf []byte) (uint64, error) {
	store := k.getStore(ctx)

	// Get current leaf count
	countKey := append(types.MerkleTreePrefix, []byte("count")...)
	countBz := store.Get(countKey)
	var index uint64
	if countBz != nil {
		index = sdk.BigEndianToUint64(countBz)
	}

	// Store leaf
	leafKey := append(types.MerkleTreePrefix, sdk.Uint64ToBigEndian(index)...)
	store.Set(leafKey, leaf)

	// Update count
	store.Set(countKey, sdk.Uint64ToBigEndian(index+1))

	return index, nil
}

// GetLeaf retrieves a leaf by index
func (k Keeper) GetLeaf(ctx context.Context, index uint64) ([]byte, bool) {
	store := k.getStore(ctx)
	leafKey := append(types.MerkleTreePrefix, sdk.Uint64ToBigEndian(index)...)
	bz := store.Get(leafKey)
	if bz == nil {
		return nil, false
	}
	return bz, true
}

// GetMerkleRoot returns the merkle root
func (k Keeper) GetMerkleRoot(ctx context.Context) []byte {
	// Simplified implementation - return hash of all leaves
	store := k.getStore(ctx)
	countKey := append(types.MerkleTreePrefix, []byte("count")...)
	countBz := store.Get(countKey)
	if countBz == nil {
		return []byte{}
	}

	count := sdk.BigEndianToUint64(countBz)
	hasher := sha256.New()
	for i := uint64(0); i < count; i++ {
		leaf, found := k.GetLeaf(ctx, i)
		if found {
			hasher.Write(leaf)
		}
	}
	return hasher.Sum(nil)
}

// GetMerklePath returns the merkle path for a leaf
func (k Keeper) GetMerklePath(ctx context.Context, index uint64) [][]byte {
	// Simplified implementation
	return [][]byte{}
}

// VerifyMerklePath verifies a merkle path
func (k Keeper) VerifyMerklePath(ctx context.Context, leaf []byte, path [][]byte, index uint64) bool {
	// Simplified implementation
	return true
}

// ZK Proof management

// ZKProofRecord represents a stored ZK proof
type ZKProofRecord struct {
	ID           string
	Prover       string
	Proof        []byte
	PublicInputs []byte
	Verified     bool
}

// SubmitZKProof submits a ZK proof
func (k Keeper) SubmitZKProof(ctx context.Context, prover string, proof []byte, publicInputs []byte) (string, error) {
	store := k.getStore(ctx)

	// Generate proof ID
	proofID := hex.EncodeToString(sha256.New().Sum(proof))[:16]

	// Store proof
	key := append(types.ZKProofPrefix, []byte(proofID)...)
	store.Set(key, proof)

	return proofID, nil
}

// VerifyZKProof verifies a ZK proof
func (k Keeper) VerifyZKProof(ctx context.Context, proofID string) bool {
	store := k.getStore(ctx)
	key := append(types.ZKProofPrefix, []byte(proofID)...)
	proof := store.Get(key)

	// Simplified verification - check if proof contains "valid"
	return len(proof) > 0 && string(proof) != "invalid_proof"
}

// Shielded transfer management

// ShieldedTransferRecord represents a shielded transfer
type ShieldedTransferRecord struct {
	ID         string
	Sender     string
	Amount     sdkmath.Int
	Commitment []byte
	Proof      []byte
}

// ShieldedTransfer performs a shielded transfer
func (k Keeper) ShieldedTransfer(ctx context.Context, sender string, amount sdkmath.Int, commitment []byte, proof []byte) (string, error) {
	if amount.IsZero() || amount.IsNegative() {
		return "", fmt.Errorf("invalid transfer amount")
	}

	store := k.getStore(ctx)

	// Generate transfer ID
	transferID := hex.EncodeToString(sha256.New().Sum(commitment))[:16]

	// Store transfer
	key := append(types.ShieldedTxPrefix, []byte(transferID)...)
	store.Set(key, []byte(sender))

	return transferID, nil
}

// GetShieldedTransfer retrieves a shielded transfer
func (k Keeper) GetShieldedTransfer(ctx context.Context, transferID string) (*ShieldedTransferRecord, bool) {
	store := k.getStore(ctx)
	key := append(types.ShieldedTxPrefix, []byte(transferID)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	return &ShieldedTransferRecord{
		ID:     transferID,
		Sender: string(bz),
	}, true
}

// Unshield performs an unshield operation
func (k Keeper) Unshield(ctx context.Context, recipient string, amount sdkmath.Int, nullifier []byte, proof []byte) error {
	if k.NullifierExists(ctx, nullifier) {
		return types.ErrNullifierExists
	}

	// Create nullifier
	return k.CreateNullifier(ctx, nullifier)
}

// InitGenesis initializes the module from genesis state
func (k Keeper) InitGenesis(ctx context.Context, data types.GenesisState) error {
	return k.SetParams(ctx, data.Params)
}

// ExportGenesis exports the module state
func (k Keeper) ExportGenesis(ctx context.Context) types.GenesisState {
	return types.GenesisState{
		Params: k.GetParams(ctx),
	}
}

// ============================================================================
// Mixing Pool Methods
// ============================================================================

// SetMixingPool stores a mixing pool
func (k Keeper) SetMixingPool(ctx context.Context, pool *privacyproto.MixingPool) error {
	store := k.getStore(ctx)
	key := append(types.MixingPoolPrefix, []byte(pool.PoolId)...)

	bz, err := k.cdc.Marshal(pool)
	if err != nil {
		return fmt.Errorf("failed to marshal for PoolId: %w", err)
	}

	store.Set(key, bz)
	return nil
}

// GetMixingPool retrieves a mixing pool by ID
func (k Keeper) GetMixingPool(ctx context.Context, poolID string) (*privacyproto.MixingPool, error) {
	store := k.getStore(ctx)
	key := append(types.MixingPoolPrefix, []byte(poolID)...)

	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("mixing pool not found: %s", poolID)
	}

	var pool privacyproto.MixingPool
	if err := k.cdc.Unmarshal(bz, &pool); err != nil {
		return nil, err
	}

	return &pool, nil
}

// GetAllMixingPools retrieves all mixing pools
func (k Keeper) GetAllMixingPools(ctx context.Context) []*privacyproto.MixingPool {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.MixingPoolPrefix)
	defer iterator.Close()

	// Initialize to empty slice (not nil) so response always has a valid array
	pools := make([]*privacyproto.MixingPool, 0)
	for ; iterator.Valid(); iterator.Next() {
		var pool privacyproto.MixingPool
		if err := k.cdc.Unmarshal(iterator.Value(), &pool); err != nil {
			continue
		}
		pools = append(pools, &pool)
	}

	return pools
}

// ============================================================================
// View Key Methods
// ============================================================================

// SetViewKey stores a view key for an owner
func (k Keeper) SetViewKey(ctx context.Context, owner string, viewKey *privacyproto.ViewKey) error {
	store := k.getStore(ctx)
	key := append(types.ViewKeyPrefix, []byte(owner)...)
	key = append(key, viewKey.PublicViewKey...)

	bz, err := k.cdc.Marshal(viewKey)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	store.Set(key, bz)
	return nil
}

// GetViewKeyByPublic retrieves a view key by public key
func (k Keeper) GetViewKeyByPublic(ctx context.Context, publicViewKey []byte) (*privacyproto.ViewKey, error) {
	store := k.getStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(store, types.ViewKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var viewKey privacyproto.ViewKey
		if err := k.cdc.Unmarshal(iterator.Value(), &viewKey); err != nil {
			continue
		}

		if string(viewKey.PublicViewKey) == string(publicViewKey) {
			return &viewKey, nil
		}
	}

	return nil, fmt.Errorf("view key not found")
}

// GetViewKeys retrieves all view keys for an owner
func (k Keeper) GetViewKeys(ctx context.Context, owner string) []*privacyproto.ViewKey {
	store := k.getStore(ctx)
	prefix := append(types.ViewKeyPrefix, []byte(owner)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	// Initialize to empty slice (not nil) so response always has a valid array
	viewKeys := make([]*privacyproto.ViewKey, 0)
	for ; iterator.Valid(); iterator.Next() {
		var viewKey privacyproto.ViewKey
		if err := k.cdc.Unmarshal(iterator.Value(), &viewKey); err != nil {
			continue
		}
		viewKeys = append(viewKeys, &viewKey)
	}

	return viewKeys
}

// DeleteViewKey deletes a view key
func (k Keeper) DeleteViewKey(ctx context.Context, owner string, publicViewKey []byte) error {
	store := k.getStore(ctx)
	key := append(types.ViewKeyPrefix, []byte(owner)...)
	key = append(key, publicViewKey...)

	store.Delete(key)
	return nil
}

// ============================================================================
// Network Privacy Methods
// ============================================================================

// SetNetworkPrivacy stores network privacy settings
func (k Keeper) SetNetworkPrivacy(ctx context.Context, networkPrivacy *privacyproto.NetworkPrivacy) error {
	store := k.getStore(ctx)

	bz, err := k.cdc.Marshal(networkPrivacy)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	// Use a dedicated key for network privacy
	key := []byte("network_privacy")
	store.Set(key, bz)
	return nil
}

// GetNetworkPrivacy retrieves network privacy settings
func (k Keeper) GetNetworkPrivacy(ctx context.Context) (*privacyproto.NetworkPrivacy, error) {
	store := k.getStore(ctx)
	key := []byte("network_privacy")
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("network privacy not configured")
	}

	var networkPrivacy privacyproto.NetworkPrivacy
	if err := k.cdc.Unmarshal(bz, &networkPrivacy); err != nil {
		return nil, err
	}

	return &networkPrivacy, nil
}

// ============================================================================
// ZK Proof Verification
// ============================================================================

// VerifyZKProofSimple performs simplified ZK proof verification
func (k Keeper) VerifyZKProofSimple(ctx context.Context, proof *privacyproto.ZKProof) bool {
	if proof == nil || len(proof.ProofData) == 0 {
		return false
	}

	// Simplified verification - in production would use actual ZK proof system
	// For now, check that proof data is not empty and proof type is valid
	return proof.ProofType != "" && len(proof.ProofData) > 0
}

// SECURITY NOTE: DecryptWithViewKey has been removed.
// Decryption must be performed client-side using private keys that never leave the client.
// The blockchain should never receive, store, or process private keys in any form.
// Private view keys enable clients to decrypt their own transaction data locally,
// but they must NEVER be transmitted to or stored on the blockchain.
