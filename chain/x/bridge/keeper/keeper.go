// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	stdmath "math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	secp256k1Curve "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	lru "github.com/hashicorp/golang-lru/v2"
	// #nosec G507 -- RIPEMD160 is required for Bitcoin/Cosmos-style address derivation compatibility
	"golang.org/x/crypto/ripemd160" //nolint:staticcheck,gosec

	"github.com/aequitas/aura/chain/x/bridge/types"
)

const sourceChainAura = "aura"

const (
	// DefaultTransferCacheSize is the default size of the transfer LRU cache
	// Increased from 1000 to 5000 for better cache hit rate on high-traffic bridges
	DefaultTransferCacheSize = 5000
)

// Keeper of the bridge store
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace

	// Dependencies
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
	vcKeeper      types.VCRegistryKeeper // For shared identity verification
	stakingKeeper types.StakingKeeper    // For validator slashing

	// LRU cache for transfer lookups (improves performance for frequent hash-based queries)
	transferCache *lru.Cache[string, *types.CrossChainTransfer]
}

// NewKeeper creates a new bridge Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps *paramtypes.Subspace,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	vcKeeper types.VCRegistryKeeper,
	stakingKeeper types.StakingKeeper,
) *Keeper {
	var paramstore paramtypes.Subspace
	if ps != nil {
		if !ps.HasKeyTable() {
			paramstore = ps.WithKeyTable(types.ParamKeyTable())
		} else {
			paramstore = *ps
		}
	}

	// Initialize transfer cache
	transferCache, err := lru.New[string, *types.CrossChainTransfer](DefaultTransferCacheSize)
	if err != nil {
		// In production, we should handle this error, but for now we'll log and continue without cache
		// The keeper will still work, just without caching benefits
		transferCache = nil
	}

	return &Keeper{
		storeKey:      storeKey,
		cdc:           cdc,
		paramstore:    paramstore,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		vcKeeper:      vcKeeper,
		stakingKeeper: stakingKeeper,
		transferCache: transferCache,
	}
}

// Logger returns a module-specific logger
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k *Keeper) store(ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(k.storeKey)
}

// nextTransferID generates a deterministic transfer ID based on block height and transaction hash.
// This eliminates race conditions that could occur with counter-based IDs in concurrent execution.
//
// SECURITY CRITICAL: Deterministic IDs prevent race conditions where multiple transactions
// in the same block could generate duplicate IDs with counter-based approach.
//
// Format: transferID = SHA256(blockHeight + headerHash + txBytes)[:8] as uint64
// This encoding:
//   - Combines block height, header hash, and transaction bytes for uniqueness
//   - Guarantees uniqueness (no two distinct txs will produce same hash)
//   - Provides deterministic generation (same tx data → same ID)
//   - Truncates to 8 bytes (64-bit uint) for compact representation
//
// Example:
//   - Block 100, specific tx bytes → consistent deterministic ID
//
// Returns:
//   - string: Deterministic transfer ID in format "transfer-{id}"
//   - error: Error if unable to generate unique ID after max retries
func (k *Keeper) nextTransferID(ctx sdk.Context) (string, error) {
	blockHeight := ctx.BlockHeight()
	headerHash := ctx.HeaderHash()
	txBytes := ctx.TxBytes()

	// CRITICAL: Use defensive programming - ensure we have enough data
	if blockHeight < 0 {
		blockHeight = 0
	}

	// Build deterministic hash input: blockHeight + headerHash + txBytes
	// This combination ensures uniqueness across all transactions
	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, uint64(blockHeight))

	// Combine all components
	var hashInput []byte
	hashInput = append(hashInput, heightBytes...)
	hashInput = append(hashInput, headerHash...)
	hashInput = append(hashInput, txBytes...)

	// Compute SHA256 hash
	hash := sha256.Sum256(hashInput)

	// Take first 8 bytes and convert to uint64 for compact ID
	transferID := binary.BigEndian.Uint64(hash[:8])

	// DEFENSIVE: Check for duplicate IDs in store (extremely unlikely with SHA256)
	// This is a safety check to detect any potential collisions
	// If collision detected, retry with additional entropy up to maxRetries
	const maxRetries = 10
	for attempt := 0; attempt < maxRetries; attempt++ {
		idStr := fmt.Sprintf("transfer-%d", transferID)
		if _, found := k.getTransfer(ctx, idStr); !found {
			return idStr, nil // Unique ID found
		}

		// CRITICAL: Collision detected, add entropy and retry
		// This should never happen with SHA256 but we handle it defensively
		ctx.Logger().Error("RARE: Transfer ID collision detected, regenerating with nonce",
			"transfer_id", idStr,
			"block_height", blockHeight,
			"attempt", attempt+1,
			"max_retries", maxRetries)

		// Add timestamp nonce with attempt number for additional entropy
		nonceBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(nonceBytes, uint64(ctx.BlockTime().UnixNano()+int64(attempt)))
		hashInput = append(hashInput, nonceBytes...)

		hash = sha256.Sum256(hashInput)
		transferID = binary.BigEndian.Uint64(hash[:8])
	}

	return "", fmt.Errorf("failed to generate unique transfer ID after %d attempts", maxRetries)
}

// extractBlockHeightFromTransferID extracts information from a deterministic transfer ID.
// Note: With hash-based IDs, we cannot extract the original block height.
// Returns (0, 0, false) for new deterministic IDs, (height, 0, true) for legacy sequential IDs.
func (k *Keeper) extractBlockHeightFromTransferID(transferID string) (int64, int64, bool) {
	// Parse transfer ID: "transfer-{id}"
	if !strings.HasPrefix(transferID, "transfer-") {
		return 0, 0, false
	}

	idStr := strings.TrimPrefix(transferID, "transfer-")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, 0, false
	}

	// Legacy sequential IDs (old counter-based system) have small values
	// Modern hash-based IDs are much larger (64-bit hash values)
	// We use a threshold to distinguish between them
	const legacyIDThreshold = uint64(1 << 40) // 1 trillion

	if id < legacyIDThreshold {
		// Legacy sequential ID - return as-is for backward compatibility
		return int64(id), 0, true
	}

	// Modern hash-based ID - cannot extract height
	return 0, 0, false
}

func (k *Keeper) setTransfer(ctx sdk.Context, transfer *types.CrossChainTransfer) error {
	if transfer == nil || transfer.TransferId == "" {
		return nil
	}
	bz, err := k.cdc.Marshal(transfer)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal transfer %s: %v", transfer.TransferId, err)
	}
	k.store(ctx).Set(types.TransferKey(transfer.TransferId), bz)

	// Invalidate cache entry to ensure consistency
	if k.transferCache != nil {
		k.transferCache.Remove(transfer.TransferId)
	}

	return nil
}

func (k *Keeper) getTransfer(ctx sdk.Context, transferID string) (*types.CrossChainTransfer, bool) {
	if transferID == "" {
		return nil, false
	}

	// Check cache first
	if k.transferCache != nil {
		if cached, ok := k.transferCache.Get(transferID); ok {
			// Cache hit - return cached value
			return cached, true
		}
	}

	// Cache miss - fetch from store
	store := k.store(ctx)
	bz := store.Get(types.TransferKey(transferID))
	if bz == nil {
		return nil, false
	}
	var transfer types.CrossChainTransfer
	if err := k.cdc.Unmarshal(bz, &transfer); err != nil {
		return nil, false
	}

	// Update cache
	if k.transferCache != nil {
		k.transferCache.Add(transferID, &transfer)
	}

	return &transfer, true
}

// GetTransfer is a public exported method for getting a transfer (for tests).
func (k *Keeper) GetTransfer(ctx sdk.Context, transferID string) (*types.CrossChainTransfer, bool) {
	return k.getTransfer(ctx, transferID)
}

func (k *Keeper) deleteTransfer(ctx sdk.Context, transferID string) {
	if transferID == "" {
		return
	}
	k.store(ctx).Delete(types.TransferKey(transferID))

	// Invalidate cache
	if k.transferCache != nil {
		k.transferCache.Remove(transferID)
	}
}

// indexTransferHash indexes a transfer by its hash for fast lookup.
//
// ATOMICITY NOTE: This function creates an index entry mapping hash -> transferID.
// The operation is atomic within a single transaction because Cosmos SDK processes
// transactions serially and the KVStore changes are committed atomically.
// However, if this function is called after storing the transfer and then a subsequent
// operation fails, the index may point to a valid transfer that the transaction
// should have rolled back. Callers should ensure all related operations succeed
// or the transaction fails atomically.
func (k *Keeper) indexTransferHash(ctx sdk.Context, hash, transferID string) {
	if hash == "" || transferID == "" {
		return
	}
	k.store(ctx).Set(types.TransferHashIndexKey(strings.ToLower(hash)), []byte(transferID))
}

func (k *Keeper) transferIDByHash(ctx sdk.Context, hash string) (string, bool) {
	if hash == "" {
		return "", false
	}
	bz := k.store(ctx).Get(types.TransferHashIndexKey(strings.ToLower(hash)))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k *Keeper) setChainConfig(ctx sdk.Context, config types.ChainConfig) error {
	bz, err := k.cdc.Marshal(&config)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal chain config %s: %v", config.ChainId, err)
	}
	k.store(ctx).Set(types.ChainConfigKey(strings.ToLower(config.ChainId)), bz)
	return nil
}

func (k *Keeper) getChainConfig(ctx sdk.Context, chainID string) (types.ChainConfig, bool) {
	bz := k.store(ctx).Get(types.ChainConfigKey(strings.ToLower(chainID)))
	if bz == nil {
		return types.ChainConfig{}, false
	}
	var cfg types.ChainConfig
	if err := k.cdc.Unmarshal(bz, &cfg); err != nil {
		return types.ChainConfig{}, false
	}
	return cfg, true
}

func (k *Keeper) getAllChainConfigs(ctx sdk.Context) []types.ChainConfig {
	store := k.store(ctx)
	iterator := store.Iterator(types.ChainConfigPrefix, storetypes.PrefixEndBytes(types.ChainConfigPrefix))
	defer iterator.Close()
	configs := make([]types.ChainConfig, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var cfg types.ChainConfig
		if err := k.cdc.Unmarshal(iterator.Value(), &cfg); err != nil {
			// Log corrupted data but continue iteration
			k.Logger(ctx).Error("failed to unmarshal chain config",
				"key", hex.EncodeToString(iterator.Key()),
				"error", err.Error())
			continue
		}
		configs = append(configs, cfg)
	}
	return configs
}

func (k *Keeper) setSharedIdentity(ctx sdk.Context, identity *types.SharedIdentity) error {
	if identity == nil || identity.Address == "" {
		return nil
	}
	bz, err := k.cdc.Marshal(identity)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal shared identity %s: %v", identity.Address, err)
	}
	k.store(ctx).Set(types.SharedIdentityKey(identity.Address), bz)
	return nil
}

func (k *Keeper) getSharedIdentity(ctx sdk.Context, address string) (*types.SharedIdentity, bool) {
	bz := k.store(ctx).Get(types.SharedIdentityKey(address))
	if bz == nil {
		return nil, false
	}
	var identity types.SharedIdentity
	if err := k.cdc.Unmarshal(bz, &identity); err != nil {
		return nil, false
	}
	return &identity, true
}

// findSharedIdentityByLinkedAddress searches all shared identities to find one with the specified linked address.
// This prevents identity hijacking by checking if an address is already linked to another identity.
//
// Security considerations:
//   - Iterates all identities to find conflicts (acceptable for identity operations)
//   - Returns nil if no identity has this address linked
//   - Case-insensitive comparison for chain names
//
// Parameters:
//   - ctx: SDK context for state access
//   - chainName: Chain identifier ("paw", "xai", "aura")
//   - address: Address to search for on the specified chain
//
// Returns:
//   - SharedIdentity if found, nil otherwise
func (k *Keeper) findSharedIdentityByLinkedAddress(ctx sdk.Context, chainName string, address string) *types.SharedIdentity {
	if chainName == "" || address == "" {
		return nil
	}

	chainName = strings.ToLower(chainName)

	// Iterate all shared identities to find the one with this linked address
	identities := k.getAllSharedIdentities(ctx)
	for _, identity := range identities {
		if identity.LinkedAddresses != nil {
			if linkedAddr, exists := identity.LinkedAddresses[chainName]; exists {
				// Case-sensitive address comparison (addresses are case-sensitive)
				if linkedAddr == address {
					return identity
				}
			}
		}
	}

	return nil
}

// verifyExternalAddressOwnership verifies that the signer owns an external chain address.
//
// SECURITY CRITICAL: This function verifies cross-chain ownership using cryptographic signatures.
// The signature proves that the signer has the private key for the external chain address.
//
// Expected signature format:
//   - Message: "Link <CHAIN> address <externalAddress> to Aura address <auraAddress>"
//   - Signature: secp256k1 signature (65 bytes: R[32] || S[32] || V[1])
//   - Recovery ID (V) is used to recover the public key from the signature
//
// Verification process:
//  1. Hash the expected message using SHA256
//  2. Recover the public key from signature using recovery ID
//  3. Derive the address from the recovered public key
//  4. Compare derived address with claimed external address
//
// Parameters:
//   - ctx: SDK context (for logging)
//   - chainName: Name of the external chain (e.g., "paw", "xai") - used for logging and message format
//   - auraAddress: The Aura address being linked
//   - externalAddress: The external chain address to verify ownership of
//   - signature: Cryptographic signature proving ownership (65 bytes)
//
// Returns:
//   - true if signature is valid and proves ownership
//   - false if signature is invalid or address format is wrong
func (k *Keeper) verifyExternalAddressOwnership(ctx sdk.Context, chainName string, auraAddress string, externalAddress string, signature []byte) bool {
	// TELEMETRY: Track signature verification start time
	// Use ctx.BlockTime() instead of time.Now() for blockchain determinism
	startTime := ctx.BlockTime()
	defer func() {
		duration := ctx.BlockTime().Sub(startTime)
		// Record will happen in specific failure/success paths below
		_ = duration
	}()

	// Normalize chain name for message formatting
	chainNameUpper := chainName
	if chainName == "paw" {
		chainNameUpper = "PAW"
	} else if chainName == "xai" {
		chainNameUpper = "XAI"
	}

	if len(signature) == 0 || externalAddress == "" || auraAddress == "" {
		k.recordSignatureMismatch(chainName, "link_address", "empty_input")
		k.recordSignatureVerification(chainName, "link_address", false, ctx.BlockTime().Sub(startTime))
		return false
	}

	// Validate signature format: secp256k1 signatures are 65 bytes (R[32] || S[32] || V[1])
	// where V is the recovery ID (0, 1, 2, or 3)
	if len(signature) != 65 {
		ctx.Logger().Error("Invalid signature length",
			"chain", chainName,
			"expected", 65,
			"actual", len(signature),
			"external_address", externalAddress)
		k.recordSignatureMismatch(chainName, "link_address", "invalid_signature_length")
		k.recordSignatureVerification(chainName, "link_address", false, ctx.BlockTime().Sub(startTime))
		return false
	}

	// Build the expected message that should have been signed
	// Format: "Link <CHAIN> address <externalAddress> to Aura address <auraAddress>"
	message := fmt.Sprintf("Link %s address %s to Aura address %s", chainNameUpper, externalAddress, auraAddress)

	// Hash the message using SHA256 (standard for Cosmos SDK chains)
	msgHash := sha256.Sum256([]byte(message))

	// Extract recovery ID (V) from the last byte
	// Recovery ID can be 0-7 (0-3 for uncompressed, 4-7 for compressed)
	// Some implementations add 27, making it 27-34
	recoveryID := signature[64]
	if recoveryID >= 27 {
		recoveryID -= 27
	}
	if recoveryID > 7 {
		ctx.Logger().Error("Invalid recovery ID in signature",
			"chain", chainName,
			"recovery_id", recoveryID,
			"external_address", externalAddress)
		k.recordInvalidRecoveryID(chainName)
		k.recordSignatureMismatch(chainName, "link_address", "invalid_recovery_id")
		k.recordSignatureVerification(chainName, "link_address", false, ctx.BlockTime().Sub(startTime))
		return false
	}

	// Extract R and S components (first 64 bytes)
	sigBytes := signature[:64]

	// SECURITY FIX: Check replay protection before attempting signature verification
	// This prevents the same signature from being reused
	signatureHash := sha256.Sum256(signature)
	if k.isSignatureUsed(ctx, signatureHash[:]) {
		ctx.Logger().Error("Signature replay attack detected",
			"chain", chainName,
			"external_address", externalAddress,
			"aura_address", auraAddress,
			"signature_hash", hex.EncodeToString(signatureHash[:]))
		k.recordSignatureMismatch(chainName, "link_address", "signature_replay")
		k.recordSignatureVerification(chainName, "link_address", false, ctx.BlockTime().Sub(startTime))
		return false
	}

	// SECURITY FIX: Check rate limiting before expensive cryptographic operations
	// This prevents DoS attacks via signature grinding
	if err := k.checkSignatureRateLimit(ctx, externalAddress); err != nil {
		ctx.Logger().Error("Signature rate limit exceeded",
			"chain", chainName,
			"external_address", externalAddress,
			"aura_address", auraAddress,
			"error", err.Error())
		k.recordSignatureMismatch(chainName, "link_address", "rate_limit_exceeded")
		k.recordSignatureVerification(chainName, "link_address", false, ctx.BlockTime().Sub(startTime))
		return false
	}

	// SECURITY FIX: Use only the claimed recovery ID, not all possible IDs
	// This prevents signature malleability and DoS amplification attacks
	pubKey, err := k.recoverPubKeyFromSignature(msgHash[:], sigBytes, recoveryID)
	if err != nil {
		ctx.Logger().Error("Failed to recover public key from signature",
			"chain", chainName,
			"external_address", externalAddress,
			"aura_address", auraAddress,
			"recovery_id", recoveryID,
			"error", err.Error())
		k.recordPubKeyRecoveryFailure(chainName, fmt.Sprintf("%d", recoveryID))
		k.recordSignatureMismatch(chainName, "link_address", "pubkey_recovery_failed")
		k.recordSignatureVerification(chainName, "link_address", false, ctx.BlockTime().Sub(startTime))
		return false
	}

	// Derive address from recovered public key using chain-specific derivation
	derivedAddress := k.deriveExternalAddressFromPubKey(pubKey, chainName)

	// Verify that the derived address matches the claimed external address
	if derivedAddress != externalAddress {
		ctx.Logger().Error("Address mismatch",
			"chain", chainName,
			"claimed", externalAddress,
			"derived", derivedAddress,
			"recovery_id", recoveryID)
		k.recordSignatureMismatch(chainName, "link_address", "address_mismatch")
		k.recordSignatureVerification(chainName, "link_address", false, ctx.BlockTime().Sub(startTime))
		return false
	}

	recoveredPubKey := pubKey

	// Additional verification: Verify the signature with the recovered public key
	// Use ecdsa library directly since external wallets use different signature formats
	// than Cosmos SDK's native secp256k1 implementation
	if !k.verifyEcdsaSignature(recoveredPubKey, msgHash[:], sigBytes) {
		ctx.Logger().Error("Signature verification failed",
			"chain", chainName,
			"external_address", externalAddress,
			"aura_address", auraAddress)
		k.recordSignatureMismatch(chainName, "link_address", "ecdsa_verification_failed")
		k.recordSignatureVerification(chainName, "link_address", false, ctx.BlockTime().Sub(startTime))
		return false
	}

	ctx.Logger().Info("Address ownership verified successfully",
		"chain", chainName,
		"external_address", externalAddress,
		"aura_address", auraAddress)

	// SECURITY FIX: Mark signature as used to prevent replay attacks
	k.markSignatureUsed(ctx, signatureHash[:], ctx.BlockHeight())

	// TELEMETRY: Record successful verification
	k.recordSignatureVerification(chainName, "link_address", true, ctx.BlockTime().Sub(startTime))

	return true
}

// verifyPawAddressOwnership verifies that the signer owns the PAW address.
// This is a thin wrapper around verifyExternalAddressOwnership for backwards compatibility.
//
// Deprecated: Use verifyExternalAddressOwnership directly for new code.
func (k *Keeper) verifyPawAddressOwnership(ctx sdk.Context, auraAddress string, pawAddress string, signature []byte) bool {
	return k.verifyExternalAddressOwnership(ctx, "paw", auraAddress, pawAddress, signature)
}

// verifyXaiAddressOwnership verifies that the signer owns the XAI address.
// This is a thin wrapper around verifyExternalAddressOwnership for backwards compatibility.
//
// Deprecated: Use verifyExternalAddressOwnership directly for new code.
func (k *Keeper) verifyXaiAddressOwnership(ctx sdk.Context, auraAddress string, xaiAddress string, signature []byte) bool {
	return k.verifyExternalAddressOwnership(ctx, "xai", auraAddress, xaiAddress, signature)
}

func (k *Keeper) setSwap(ctx sdk.Context, swap *types.CrossChainSwap) error {
	if swap == nil || swap.SwapId == "" {
		return nil
	}
	bz, err := k.cdc.Marshal(swap)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal swap %s: %v", swap.SwapId, err)
	}
	k.store(ctx).Set(types.SwapKey(swap.SwapId), bz)
	return nil
}

func (k *Keeper) getSwap(ctx sdk.Context, swapID string) (*types.CrossChainSwap, bool) {
	bz := k.store(ctx).Get(types.SwapKey(swapID))
	if bz == nil {
		return nil, false
	}
	var swap types.CrossChainSwap
	if err := k.cdc.Unmarshal(bz, &swap); err != nil {
		return nil, false
	}
	return &swap, true
}

func (k *Keeper) getWrappedToken(ctx sdk.Context, denom string) (*types.WrappedToken, bool) {
	bz := k.store(ctx).Get(types.WrappedTokenKey(denom))
	if bz == nil {
		return nil, false
	}
	var token types.WrappedToken
	if err := k.cdc.Unmarshal(bz, &token); err != nil {
		return nil, false
	}
	return &token, true
}

func (k *Keeper) setWrappedToken(ctx sdk.Context, token *types.WrappedToken) error {
	if token == nil || token.WrappedDenom == "" {
		return nil
	}
	bz, err := k.cdc.Marshal(token)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal wrapped token %s: %v", token.WrappedDenom, err)
	}
	k.store(ctx).Set(types.WrappedTokenKey(token.WrappedDenom), bz)
	return nil
}

func (k *Keeper) getAllWrappedTokens(ctx sdk.Context) []types.WrappedToken {
	store := k.store(ctx)
	iterator := store.Iterator(types.WrappedTokenPrefix, storetypes.PrefixEndBytes(types.WrappedTokenPrefix))
	defer iterator.Close()
	tokens := make([]types.WrappedToken, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var token types.WrappedToken
		if err := k.cdc.Unmarshal(iterator.Value(), &token); err != nil {
			// Log corrupted data but continue iteration
			k.Logger(ctx).Error("failed to unmarshal wrapped token",
				"key", hex.EncodeToString(iterator.Key()),
				"error", err.Error())
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

func (k *Keeper) getRelayerStats(ctx sdk.Context, relayer string) (*types.RelayerStats, bool) {
	bz := k.store(ctx).Get(types.RelayerKey(relayer))
	if bz == nil {
		return nil, false
	}
	var stats types.RelayerStats
	if err := k.cdc.Unmarshal(bz, &stats); err != nil {
		return nil, false
	}
	return &stats, true
}

func (k *Keeper) recordRelayerStats(ctx sdk.Context, relayer string, success bool, volume sdkmath.Int) {
	if relayer == "" {
		return
	}
	stats, _ := k.getRelayerStats(ctx, relayer)
	blockTime := ctx.BlockTime()
	if stats == nil {
		stats = &types.RelayerStats{
			RelayerAddress:   relayer,
			TotalVolume:      sdkmath.ZeroInt(),
			UptimePercentage: sdkmath.LegacyNewDec(100),
			LastRelay:        &blockTime,
		}
	}
	stats.TotalTransfersRelayed++
	if success {
		stats.SuccessfulTransfers++
	} else {
		stats.FailedTransfers++
	}
	// stats.TotalVolume is already math.Int, use directly
	stats.TotalVolume = stats.TotalVolume.Add(volume)
	stats.LastRelay = &blockTime
	_ = k.setRelayerStats(ctx, stats) // Best effort, stats are non-critical
}

func (k *Keeper) setRelayerStats(ctx sdk.Context, stats *types.RelayerStats) error {
	if stats == nil || stats.RelayerAddress == "" {
		return nil
	}
	bz, err := k.cdc.Marshal(stats)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal relayer stats %s: %v", stats.RelayerAddress, err)
	}
	k.store(ctx).Set(types.RelayerKey(stats.RelayerAddress), bz)
	return nil
}

func (k *Keeper) markTransferFraudulent(ctx sdk.Context, transferID string) (*types.CrossChainTransfer, error) {
	transfer, found := k.getTransfer(ctx, transferID)
	if !found {
		return nil, types.ErrTransferNotFound
	}
	transfer.Status = types.TransferStatus_FAILED
	transfer.Timestamp = ctx.BlockTime()
	if err := k.setTransfer(ctx, transfer); err != nil {
		return nil, fmt.Errorf("failed to mark transfer as fraudulent: %w", err)
	}
	return transfer, nil
}

func (k *Keeper) getAllRelayerStats(ctx sdk.Context) []*types.RelayerStats {
	store := k.store(ctx)
	iterator := store.Iterator(types.RelayerPrefix, storetypes.PrefixEndBytes(types.RelayerPrefix))
	defer iterator.Close()
	statsList := make([]*types.RelayerStats, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var stats types.RelayerStats
		if err := k.cdc.Unmarshal(iterator.Value(), &stats); err != nil {
			// Log corrupted data but continue iteration
			k.Logger(ctx).Error("failed to unmarshal relayer stats",
				"key", hex.EncodeToString(iterator.Key()),
				"error", err.Error())
			continue
		}
		statsCopy := stats
		statsList = append(statsList, &statsCopy)
	}
	return statsList
}

func (k *Keeper) setValidator(ctx sdk.Context, validator *types.BridgeValidator) error {
	if validator == nil || validator.Address == "" {
		return nil
	}
	bz, err := k.cdc.Marshal(validator)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal validator %s: %v", validator.Address, err)
	}
	k.store(ctx).Set(types.ValidatorKey(validator.Address), bz)
	return nil
}

func (k *Keeper) getValidator(ctx sdk.Context, address string) (*types.BridgeValidator, bool) {
	if address == "" {
		return nil, false
	}
	bz := k.store(ctx).Get(types.ValidatorKey(address))
	if bz == nil {
		return nil, false
	}
	var validator types.BridgeValidator
	if err := k.cdc.Unmarshal(bz, &validator); err != nil {
		return nil, false
	}
	return &validator, true
}

func (k *Keeper) getAllValidators(ctx sdk.Context) []*types.BridgeValidator {
	store := k.store(ctx)
	iterator := store.Iterator(types.ValidatorPrefix, storetypes.PrefixEndBytes(types.ValidatorPrefix))
	defer iterator.Close()
	validators := make([]*types.BridgeValidator, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var validator types.BridgeValidator
		if err := k.cdc.Unmarshal(iterator.Value(), &validator); err != nil {
			// Log corrupted data but continue iteration
			k.Logger(ctx).Error("failed to unmarshal bridge validator",
				"key", hex.EncodeToString(iterator.Key()),
				"error", err.Error())
			continue
		}
		valCopy := validator
		validators = append(validators, &valCopy)
	}
	return validators
}

func (k *Keeper) getAllSharedIdentities(ctx sdk.Context) []*types.SharedIdentity {
	store := k.store(ctx)
	iterator := store.Iterator(types.SharedIdentityPrefix, storetypes.PrefixEndBytes(types.SharedIdentityPrefix))
	defer iterator.Close()
	identities := make([]*types.SharedIdentity, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var identity types.SharedIdentity
		if err := k.cdc.Unmarshal(iterator.Value(), &identity); err != nil {
			// Log corrupted data but continue iteration
			k.Logger(ctx).Error("failed to unmarshal shared identity",
				"key", hex.EncodeToString(iterator.Key()),
				"error", err.Error())
			continue
		}
		idCopy := identity
		identities = append(identities, &idCopy)
	}
	return identities
}

func (k *Keeper) getAllSwaps(ctx sdk.Context) []*types.CrossChainSwap {
	store := k.store(ctx)
	iterator := store.Iterator(types.SwapPrefix, storetypes.PrefixEndBytes(types.SwapPrefix))
	defer iterator.Close()
	swaps := make([]*types.CrossChainSwap, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var swap types.CrossChainSwap
		if err := k.cdc.Unmarshal(iterator.Value(), &swap); err != nil {
			// Log corrupted data but continue iteration
			k.Logger(ctx).Error("failed to unmarshal cross-chain swap",
				"key", hex.EncodeToString(iterator.Key()),
				"error", err.Error())
			continue
		}
		swapCopy := swap
		swaps = append(swaps, &swapCopy)
	}
	return swaps
}

func (k *Keeper) getAllTransfers(ctx sdk.Context) []*types.CrossChainTransfer {
	store := k.store(ctx)
	iterator := store.Iterator(types.TransferPrefix, storetypes.PrefixEndBytes(types.TransferPrefix))
	defer iterator.Close()
	transfers := make([]*types.CrossChainTransfer, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var transfer types.CrossChainTransfer
		if err := k.cdc.Unmarshal(iterator.Value(), &transfer); err != nil {
			// Log corrupted data but continue iteration
			k.Logger(ctx).Error("failed to unmarshal cross-chain transfer",
				"key", hex.EncodeToString(iterator.Key()),
				"error", err.Error())
			continue
		}
		transferCopy := transfer
		transfers = append(transfers, &transferCopy)
	}
	return transfers
}

// GetParams returns the total set of bridge parameters.
func (k *Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	if k.paramstore.HasKeyTable() {
		defer func() {
			if r := recover(); r != nil {
				params = types.DefaultParams()
			}
		}()
		k.paramstore.GetParamSet(ctx, &params)
		return params
	}
	return types.DefaultParams()
}

// SetParams sets bridge parameters in the param store.
func (k *Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if k.paramstore.HasKeyTable() {
		k.paramstore.SetParamSet(ctx, &params)
	}
	return nil
}

// ensureBridgeEnabled returns error if the bridge is disabled.
func (k *Keeper) ensureBridgeEnabled(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	if !params.BridgeEnabled {
		return fmt.Errorf("bridge disabled")
	}
	return nil
}

// SubmitAttestation records a validator's attestation for a transfer
func (k *Keeper) SubmitAttestation(ctx sdk.Context, transferID string, validator string, approved bool) error {
	if validator == "" {
		return fmt.Errorf("validator address required")
	}
	store := k.store(ctx)
	key := types.AttestationKey(transferID, validator)
	if store.Has(key) {
		return types.ErrDuplicateAttestation
	}
	store.Set(key, []byte{1})

	transfer, found := k.getTransfer(ctx, transferID)
	if !found {
		return types.ErrTransferNotFound
	}
	transfer.Confirmations++
	if transfer.RequiredConfirmations == 0 {
		params := k.GetParams(ctx)
		transfer.RequiredConfirmations = params.MinConfirmations
	}
	if approved {
		transfer.ValidatorSignatures = append(transfer.ValidatorSignatures, types.ValidatorSignature{
			ValidatorAddress: validator,
		})
	}
	if err := k.setTransfer(ctx, transfer); err != nil {
		return fmt.Errorf("failed to update transfer with attestation: %w", err)
	}
	return nil
}

// GetAttestations returns all validator addresses that attested a transfer
func (k *Keeper) GetAttestations(ctx sdk.Context, transferID string) []string {
	store := k.store(ctx)
	prefix := append(types.AttestationPrefix, []byte(transferID)...)
	iterator := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	defer iterator.Close()
	validators := make([]string, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		parts := strings.Split(string(iterator.Key()), string([]byte{0x00}))
		if len(parts) == 2 {
			validators = append(validators, parts[1])
		}
	}
	return validators
}

// CheckAttestationThreshold returns true if a transfer has enough attestations
func (k *Keeper) CheckAttestationThreshold(ctx sdk.Context, transferID string) bool {
	transfer, found := k.getTransfer(ctx, transferID)
	if !found {
		return false
	}
	required := transfer.RequiredConfirmations
	if required == 0 {
		required = k.GetParams(ctx).MinConfirmations
	}
	return transfer.Confirmations >= required && required > 0
}

// ProcessWithdrawal processes a withdrawal request
func (k *Keeper) ProcessWithdrawal(ctx sdk.Context, recipient string, amount sdk.Coins, transferID string) error {
	if amount.Empty() {
		return fmt.Errorf("amount cannot be empty")
	}
	params := k.GetParams(ctx)
	maxAmt, ok := sdkmath.NewIntFromString(params.MaxTransferAmount)
	if !ok {
		return fmt.Errorf("invalid max transfer amount param")
	}
	if amount.AmountOf(amount[0].Denom).GT(maxAmt) {
		return types.ErrCircuitBreakerTripped
	}
	transfer, found := k.getTransfer(ctx, transferID)
	if !found {
		return types.ErrTransferNotFound
	}
	transfer.Status = types.TransferStatus_COMPLETED
	transfer.Timestamp = ctx.BlockTime()
	return k.setTransfer(ctx, transfer)
}

// InitiateTransfer initiates a cross-chain transfer
//
// This method creates a new cross-chain transfer request from the Aura chain to a target chain.
// It validates all inputs, checks bridge configuration, and creates a pending transfer record.
//
// Parameters:
//   - ctx: SDK context for state access
//   - sender: Address initiating the transfer (on Aura chain)
//   - recipient: Destination address (on target chain)
//   - amount: Coins to transfer (with denomination)
//   - targetChain: Target chain identifier (e.g., "ethereum", "paw")
//
// Returns:
//   - transferID: Unique identifier for this transfer
//   - error: Validation errors, unsupported chain, or max amount exceeded
//
// Security considerations:
//   - Input validation: All parameters are validated for non-empty/non-zero values
//   - Chain validation: Target chain must be supported and enabled
//   - Amount limits: Transfer amount must not exceed configured maximum
//   - Bridge status: Bridge must be enabled in params
func (k *Keeper) InitiateTransfer(ctx sdk.Context, sender string, recipient string, amount sdk.Coins, targetChain string) (string, error) {
	// Validate inputs
	if sender == "" {
		return "", fmt.Errorf("sender address required")
	}
	if recipient == "" {
		return "", fmt.Errorf("recipient address required")
	}
	if amount.Empty() || !amount.IsValid() {
		return "", fmt.Errorf("amount must be positive and valid")
	}
	if targetChain == "" {
		return "", fmt.Errorf("target chain required")
	}

	// Check if bridge is paused
	if err := k.RequireNotPaused(ctx, targetChain); err != nil {
		return "", err
	}

	// Validate target chain is supported and enabled
	chainConfig, found := k.getChainConfig(ctx, targetChain)
	if !found {
		return "", fmt.Errorf("unsupported chain: %s", targetChain)
	}
	if !chainConfig.Enabled {
		return "", fmt.Errorf("chain %s is currently disabled", targetChain)
	}

	// Validate amount does not exceed maximum
	params := k.GetParams(ctx)
	maxAmt, ok := sdkmath.NewIntFromString(params.MaxTransferAmount)
	if !ok {
		return "", fmt.Errorf("invalid max transfer amount param")
	}
	transferAmt := amount.AmountOf(amount[0].Denom)
	if transferAmt.GT(maxAmt) {
		return "", types.ErrCircuitBreakerTripped
	}

	// Generate unique transfer ID
	transferID, err := k.nextTransferID(ctx)
	if err != nil {
		return "", err
	}

	// Create cross-chain transfer record
	transfer := &types.CrossChainTransfer{
		TransferId:  transferID,
		SourceChain: sourceChainAura,
		TargetChain: targetChain,
		Sender:      sender,
		Recipient:   recipient,
		Amount:      transferAmt,
		Denom:       amount[0].Denom,
		Status:      types.TransferStatus_PENDING,
		Timestamp:   ctx.BlockTime(),
	}

	// Store transfer
	if err := k.setTransfer(ctx, transfer); err != nil {
		return "", err
	}

	// Emit event for transfer initiation
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bridge_transfer_initiated",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("sender", sender),
			sdk.NewAttribute("recipient", recipient),
			sdk.NewAttribute("amount", transferAmt.String()),
			sdk.NewAttribute("denom", amount[0].Denom),
			sdk.NewAttribute("target_chain", targetChain),
		),
	)

	return transferID, nil
}

// InitiateWithdrawal stores a pending withdrawal with timestamp metadata
func (k *Keeper) InitiateWithdrawal(ctx sdk.Context, recipient string, amount sdk.Coins) (string, error) {
	transferID, err := k.nextTransferID(ctx)
	if err != nil {
		return "", err
	}
	transfer := &types.CrossChainTransfer{
		TransferId:  transferID,
		SourceChain: sourceChainAura,
		TargetChain: sourceChainAura,
		Sender:      sourceChainAura,
		Recipient:   recipient,
		Amount:      amount.AmountOf(amount[0].Denom),
		Denom:       amount[0].Denom,
		Status:      types.TransferStatus_PENDING,
		Timestamp:   ctx.BlockTime(),
	}
	return transferID, k.setTransfer(ctx, transfer)
}

// ExecuteWithdrawal executes a withdrawal after timelock
func (k *Keeper) ExecuteWithdrawal(ctx sdk.Context, withdrawalID string) error {
	transfer, found := k.getTransfer(ctx, withdrawalID)
	if !found {
		return types.ErrWithdrawalNotFound
	}
	if transfer.Status != types.TransferStatus_PENDING {
		return nil
	}
	if !transfer.Timestamp.IsZero() {
		deadline := transfer.Timestamp.Add(types.DefaultTimelockDuration)
		if ctx.BlockTime().Before(deadline) {
			return types.ErrTimelockNotElapsed
		}
	}
	transfer.Status = types.TransferStatus_COMPLETED
	return k.setTransfer(ctx, transfer)
}

func (k *Keeper) setFraudProof(ctx sdk.Context, proof *types.FraudProof) error {
	if proof == nil || proof.ChallengedTransferId == "" {
		return nil
	}
	bz, err := k.cdc.Marshal(proof)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal fraud proof for transfer %s: %v", proof.ChallengedTransferId, err)
	}
	k.store(ctx).Set(types.FraudProofKey(proof.ChallengedTransferId), bz)
	return nil
}

func (k *Keeper) getFraudProof(ctx sdk.Context, transferID string) (*types.FraudProof, bool) {
	bz := k.store(ctx).Get(types.FraudProofKey(transferID))
	if bz == nil {
		return nil, false
	}
	var proof types.FraudProof
	if err := k.cdc.Unmarshal(bz, &proof); err != nil {
		return nil, false
	}
	return &proof, true
}

func (k *Keeper) getFraudProofWindow(ctx sdk.Context) time.Duration {
	window := types.DefaultFraudProofWindow
	if window <= 0 {
		return 0
	}
	return window
}

func (k *Keeper) getFraudProofReward(ctx sdk.Context) sdkmath.Int {
	params := types.DefaultSecurityParams()
	return params.FraudProofReward
}

func (k *Keeper) payoutFraudProofReward(ctx sdk.Context, challenger string, denom string, reward sdkmath.Int) error {
	if !reward.IsPositive() || challenger == "" || denom == "" {
		return nil
	}
	if k.bankKeeper == nil {
		return nil
	}
	addr, err := sdk.AccAddressFromBech32(challenger)
	if err != nil {
		return fmt.Errorf("failed to AccAddressFromBech32: %w", err)
	}
	coin := sdk.NewCoin(denom, reward)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return fmt.Errorf("failed to NewCoin: %w", err)
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(coin))
}

// SubmitFraudProof records a fraud challenge for a transfer and begins investigation.
//
// CRITICAL SECURITY ENHANCEMENT: This function now also marks the pending transfer
// as challenged, preventing finalization while the fraud proof is investigated.
func (k *Keeper) SubmitFraudProof(ctx sdk.Context, transferID string, submitter string, proof []byte) error {
	transferID = strings.TrimSpace(transferID)
	submitter = strings.TrimSpace(submitter)
	if transferID == "" || submitter == "" {
		return types.ErrInvalidParam
	}
	if len(proof) == 0 {
		return types.ErrInvalidEvidence
	}
	if _, found := k.getTransfer(ctx, transferID); !found {
		return types.ErrTransferNotFound
	}

	// CRITICAL: Check if there's a pending transfer to challenge
	// If the transfer is already finalized (no pending transfer), fraud proof window has passed
	pendingTransfer, hasPending := k.GetPendingTransfer(ctx, transferID)
	if !hasPending {
		return types.ErrFraudProofExpired
	}

	// Check if pending transfer is still within fraud proof window
	if k.IsPendingTransferExpired(ctx, pendingTransfer) {
		return types.ErrFraudProofExpired
	}

	if existing, found := k.getFraudProof(ctx, transferID); found {
		switch existing.Status {
		case types.FraudProofStatus_FRAUD_PROOF_INVESTIGATING, types.FraudProofStatus_FRAUD_PROOF_PENDING:
			return types.ErrFraudProofPending
		case types.FraudProofStatus_FRAUD_PROOF_VALID, types.FraudProofStatus_FRAUD_PROOF_INVALID:
			return types.ErrFraudProofAlreadyResolved
		}
	}

	fraudProofID := fmt.Sprintf("%s-%d", transferID, ctx.BlockHeight())

	fraudProof := &types.FraudProof{
		ProofId:              fraudProofID,
		ChallengedTransferId: transferID,
		Challenger:           submitter,
		FraudType:            types.FraudType_FRAUD_INVALID_SIGNATURE,
		Evidence:             proof,
		Status:               types.FraudProofStatus_FRAUD_PROOF_INVESTIGATING,
		SubmittedAt:          ctx.BlockTime(),
		RewardAmount:         sdkmath.ZeroInt(),
	}
	if err := k.setFraudProof(ctx, fraudProof); err != nil {
		return fmt.Errorf("failed to set fraud proof: %w", err)
	}

	// CRITICAL SECURITY: Mark pending transfer as challenged
	// This prevents finalization until the fraud proof is resolved
	if err := k.MarkPendingTransferChallenged(ctx, transferID, fraudProofID); err != nil {
		return fmt.Errorf("failed to mark pending transfer as challenged: %w", err)
	}

	return nil
}

// ResolveFraudProof finalizes an open fraud proof, rewarding challengers and marking transfers.
//
// SECURITY CRITICAL: When a fraud proof is validated as correct (valid=true), this function:
//  1. Marks the transfer as fraudulent
//  2. Slashes ALL validators who signed the fraudulent transfer
//  3. Pays out reward to the fraud proof challenger
//  4. Prevents the transfer from being finalized
func (k *Keeper) ResolveFraudProof(ctx sdk.Context, transferID string, valid bool) (types.FraudProof, error) {
	proof, found := k.getFraudProof(ctx, transferID)
	if !found {
		return types.FraudProof{}, types.ErrFraudProofNotFound
	}
	switch proof.Status {
	case types.FraudProofStatus_FRAUD_PROOF_VALID, types.FraudProofStatus_FRAUD_PROOF_INVALID:
		return types.FraudProof{}, types.ErrFraudProofAlreadyResolved
	case types.FraudProofStatus_FRAUD_PROOF_EXPIRED:
		return types.FraudProof{}, types.ErrFraudProofExpired
	}
	if !k.IsInFraudProofWindow(ctx, transferID) {
		proof.Status = types.FraudProofStatus_FRAUD_PROOF_EXPIRED
		if err := k.setFraudProof(ctx, proof); err != nil {
			return types.FraudProof{}, fmt.Errorf("failed to mark fraud proof as expired: %w", err)
		}
		return types.FraudProof{}, types.ErrFraudProofExpired
	}
	resolvedAt := ctx.BlockTime()
	proof.ResolvedAt = &resolvedAt
	reward := sdkmath.ZeroInt()
	if valid {
		proof.Status = types.FraudProofStatus_FRAUD_PROOF_VALID
		reward = k.getFraudProofReward(ctx)
		transfer, err := k.markTransferFraudulent(ctx, transferID)
		if err != nil {
			return types.FraudProof{}, err
		}

		// CRITICAL SECURITY: Slash all validators who signed this fraudulent transfer
		// This is the primary economic deterrent against validator fraud
		if err := k.slashValidatorsForFraudulentTransfer(ctx, transferID, proof.ProofId); err != nil {
			ctx.Logger().Error("failed to slash validators for fraudulent transfer",
				"transfer_id", transferID,
				"fraud_proof_id", proof.ProofId,
				"error", err)
			// Log error but continue - fraud proof is still valid even if slashing fails
		}

		if err := k.payoutFraudProofReward(ctx, proof.Challenger, transfer.Denom, reward); err != nil {
			return types.FraudProof{}, err
		}
	} else {
		proof.Status = types.FraudProofStatus_FRAUD_PROOF_INVALID
	}
	proof.RewardAmount = reward
	if err := k.setFraudProof(ctx, proof); err != nil {
		return types.FraudProof{}, fmt.Errorf("failed to save resolved fraud proof: %w", err)
	}
	return *proof, nil
}

// GetFraudProof retrieves a fraud proof for a transfer
func (k *Keeper) GetFraudProof(ctx sdk.Context, transferID string) (types.FraudProof, bool) {
	proof, found := k.getFraudProof(ctx, transferID)
	if !found {
		return types.FraudProof{}, false
	}
	return *proof, true
}

// IsInFraudProofWindow checks if a transfer is still in the fraud proof window
func (k *Keeper) IsInFraudProofWindow(ctx sdk.Context, transferID string) bool {
	transfer, found := k.getTransfer(ctx, transferID)
	if !found || transfer.Timestamp.IsZero() {
		return false
	}
	window := k.getFraudProofWindow(ctx)
	if window <= 0 {
		return false
	}
	return ctx.BlockTime().Sub(transfer.Timestamp) <= window
}

// AddSupportedChain adds a new supported chain configuration
func (k *Keeper) AddSupportedChain(ctx sdk.Context, chainConfig types.ChainConfig) error {
	if chainConfig.ChainId == "" {
		return types.ErrInvalidParam
	}
	return k.setChainConfig(ctx, chainConfig)
}

// GetSupportedChain retrieves a supported chain configuration
func (k *Keeper) GetSupportedChain(ctx sdk.Context, chainID string) (types.ChainConfig, bool) {
	return k.getChainConfig(ctx, chainID)
}

// RemoveSupportedChain removes a supported chain
func (k *Keeper) RemoveSupportedChain(ctx sdk.Context, chainID string) {
	k.store(ctx).Delete(types.ChainConfigKey(strings.ToLower(chainID)))
}

// DisableChain disables a supported chain
func (k *Keeper) DisableChain(ctx sdk.Context, chainID string) error {
	config, found := k.getChainConfig(ctx, chainID)
	if !found {
		return types.ErrChainNotFound
	}
	config.Enabled = false
	return k.setChainConfig(ctx, config)
}

// CalculateBridgeFee calculates the bridge fee for a transfer
func (k *Keeper) CalculateBridgeFee(ctx sdk.Context, amount sdkmath.Int, chainID string) sdkmath.Int {
	params := k.GetParams(ctx)
	if params.BridgeFeeBasisPoints == 0 {
		return sdkmath.ZeroInt()
	}
	basisPoints := int64(params.BridgeFeeBasisPoints)
	if params.BridgeFeeBasisPoints > stdmath.MaxInt64 {
		basisPoints = stdmath.MaxInt64
	}
	return amount.MulRaw(basisPoints).QuoRaw(10_000)
}

// GetCollectedFees returns the total collected fees
func (k *Keeper) GetCollectedFees(ctx sdk.Context) sdk.Coins {
	store := k.store(ctx)
	iterator := store.Iterator([]byte("collected-fees-"), []byte("collected-fees-\xff"))
	defer iterator.Close()
	fees := sdk.NewCoins()
	for ; iterator.Valid(); iterator.Next() {
		denom := strings.TrimPrefix(string(iterator.Key()), "collected-fees-")
		var amount sdkmath.Int
		if err := amount.Unmarshal(iterator.Value()); err != nil {
			k.Logger(ctx).Error("failed to unmarshal collected fee amount",
				"denom", denom,
				"error", err,
				"data_len", len(iterator.Value()))
			continue
		}
		fees = fees.Add(sdk.NewCoin(denom, amount))
	}
	return fees
}

// GetSharedIdentity is a public exported method for getting a shared identity
func (k *Keeper) GetSharedIdentity(ctx sdk.Context, address string) (*types.SharedIdentity, bool) {
	return k.getSharedIdentity(ctx, address)
}

// FindSharedIdentityByLinkedAddress is a public exported method for finding identities by linked address
func (k *Keeper) FindSharedIdentityByLinkedAddress(ctx sdk.Context, chainName string, address string) *types.SharedIdentity {
	return k.findSharedIdentityByLinkedAddress(ctx, chainName, address)
}

// VerifyPawAddressOwnership is a public exported method for PAW signature verification
func (k *Keeper) VerifyPawAddressOwnership(ctx sdk.Context, auraAddress string, pawAddress string, signature []byte) bool {
	return k.verifyPawAddressOwnership(ctx, auraAddress, pawAddress, signature)
}

// VerifyXaiAddressOwnership is a public exported method for XAI signature verification
func (k *Keeper) VerifyXaiAddressOwnership(ctx sdk.Context, auraAddress string, xaiAddress string, signature []byte) bool {
	return k.verifyXaiAddressOwnership(ctx, auraAddress, xaiAddress, signature)
}

// AddCollectedFee adds a collected fee to the total
func (k *Keeper) AddCollectedFee(ctx sdk.Context, fee sdk.Coin) {
	store := k.store(ctx)
	key := []byte("collected-fees-" + fee.Denom)
	current := sdkmath.ZeroInt()
	if bz := store.Get(key); bz != nil {
		if err := current.Unmarshal(bz); err != nil {
			k.Logger(ctx).Error("failed to unmarshal current collected fee, resetting to zero",
				"denom", fee.Denom,
				"error", err,
				"data_len", len(bz))
			// Continue with zero value - this is a recovery mechanism
		}
	}
	current = current.Add(fee.Amount)
	bz, err := current.Marshal()
	if err != nil {
		k.Logger(ctx).Error("failed to marshal collected fee amount",
			"denom", fee.Denom,
			"amount", current.String(),
			"error", err)
		return
	}
	store.Set(key, bz)
}

// IsSourceHashProcessed checks if a source transaction hash has already been processed.
// This prevents replay attacks where the same source transaction could be used
// multiple times to unlock tokens on Aura.
//
// Parameters:
//   - ctx: SDK context for state access
//   - sourceChain: Chain identifier where the transaction originated
//   - sourceHash: Transaction hash from the source chain
//
// Returns:
//   - bool: true if the hash has been processed, false otherwise
//
// Security: This is a CRITICAL function for preventing replay attacks.
// Always call this BEFORE processing any unlock/mint operation.
func (k *Keeper) IsSourceHashProcessed(ctx sdk.Context, sourceChain, sourceHash string) bool {
	// Normalize inputs to prevent bypass via case sensitivity
	sourceChain = strings.ToLower(strings.TrimSpace(sourceChain))
	sourceHash = strings.ToLower(strings.TrimSpace(sourceHash))

	if sourceChain == "" || sourceHash == "" {
		return false
	}

	store := k.store(ctx)
	key := types.ProcessedSourceHashKey(sourceChain, sourceHash)
	return store.Has(key)
}

// MarkSourceHashProcessed marks a source transaction hash as processed.
// This prevents the same source transaction from being used multiple times
// to unlock tokens (replay attack).
//
// Parameters:
//   - ctx: SDK context for state access
//   - sourceChain: Chain identifier where the transaction originated
//   - sourceHash: Transaction hash from the source chain
//
// Security considerations:
//   - This function MUST be called BEFORE minting/unlocking tokens
//   - Follows checks-effects-interactions pattern (mark before external call)
//   - State change is atomic and irreversible
//   - The hash is stored permanently in state (no expiration)
func (k *Keeper) MarkSourceHashProcessed(ctx sdk.Context, sourceChain, sourceHash string) {
	// Normalize inputs to ensure consistent storage
	sourceChain = strings.ToLower(strings.TrimSpace(sourceChain))
	sourceHash = strings.ToLower(strings.TrimSpace(sourceHash))

	if sourceChain == "" || sourceHash == "" {
		return
	}

	store := k.store(ctx)
	key := types.ProcessedSourceHashKey(sourceChain, sourceHash)
	// Store a simple marker byte - we only care about existence
	store.Set(key, []byte{1})

	// Emit event for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"source_hash_marked_processed",
			sdk.NewAttribute("source_chain", sourceChain),
			sdk.NewAttribute("source_hash", sourceHash),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}

// TryMarkSourceHashProcessing atomically checks if a source hash is already processed or processing,
// and if not, marks it as currently being processed. This prevents race conditions where multiple
// transactions with the same burn hash in the same block could all pass the replay check.
//
// This implements an atomic check-and-set pattern that is critical for preventing replay attacks
// in scenarios where multiple transactions are submitted in the same block.
//
// Parameters:
//   - ctx: SDK context for state access
//   - sourceChain: Chain identifier where the transaction originated
//   - sourceHash: Transaction hash from the source chain
//
// Returns:
//   - bool: true if successfully marked as processing (safe to proceed), false if already processed/processing
//
// Security: This function MUST be called at the very start of unlock processing, before any
// expensive validation or signature verification. The processing marker will be converted to
// a permanent processed marker by FinalizeSourceHashProcessing after successful completion.
func (k *Keeper) TryMarkSourceHashProcessing(ctx sdk.Context, sourceChain, sourceHash string) bool {
	// Normalize inputs to prevent bypass via case sensitivity
	sourceChain = strings.ToLower(strings.TrimSpace(sourceChain))
	sourceHash = strings.ToLower(strings.TrimSpace(sourceHash))

	if sourceChain == "" || sourceHash == "" {
		return false
	}

	store := k.store(ctx)

	// Check if already fully processed (permanent marker)
	processedKey := types.ProcessedSourceHashKey(sourceChain, sourceHash)
	if store.Has(processedKey) {
		return false // Already processed, reject
	}

	// Check if currently being processed by another tx in this block
	processingKey := types.ProcessingSourceHashKey(sourceChain, sourceHash)
	if store.Has(processingKey) {
		return false // Already being processed, reject to prevent race
	}

	// Atomically mark as processing
	// This ensures only one transaction can proceed past this point
	store.Set(processingKey, []byte{1})

	// Emit event for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"source_hash_marked_processing",
			sdk.NewAttribute("source_chain", sourceChain),
			sdk.NewAttribute("source_hash", sourceHash),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return true // Safe to proceed with processing
}

// FinalizeSourceHashProcessing converts a processing marker to a permanent processed marker
// and removes the temporary processing flag. This should be called after successful completion
// of the unlock operation.
//
// Parameters:
//   - ctx: SDK context for state access
//   - sourceChain: Chain identifier where the transaction originated
//   - sourceHash: Transaction hash from the source chain
//
// Security: This function completes the atomic check-and-set pattern by making the
// replay protection permanent. After this call, the source hash can never be processed again.
func (k *Keeper) FinalizeSourceHashProcessing(ctx sdk.Context, sourceChain, sourceHash string) {
	// Normalize inputs to ensure consistent storage
	sourceChain = strings.ToLower(strings.TrimSpace(sourceChain))
	sourceHash = strings.ToLower(strings.TrimSpace(sourceHash))

	if sourceChain == "" || sourceHash == "" {
		return
	}

	store := k.store(ctx)

	// Set permanent processed marker
	processedKey := types.ProcessedSourceHashKey(sourceChain, sourceHash)
	store.Set(processedKey, []byte{1})

	// Remove temporary processing marker
	processingKey := types.ProcessingSourceHashKey(sourceChain, sourceHash)
	store.Delete(processingKey)

	// Emit event for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"source_hash_processing_finalized",
			sdk.NewAttribute("source_chain", sourceChain),
			sdk.NewAttribute("source_hash", sourceHash),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}

// GetAllProcessedSourceHashes returns all processed source hashes for genesis export.
// Returns a map of "sourceChain:sourceHash" -> true for all processed hashes.
func (k *Keeper) GetAllProcessedSourceHashes(ctx sdk.Context) map[string]bool {
	store := k.store(ctx)
	iterator := store.Iterator(
		types.ProcessedSourceHashPrefix,
		storetypes.PrefixEndBytes(types.ProcessedSourceHashPrefix),
	)
	defer iterator.Close()

	processedHashes := make(map[string]bool)
	for ; iterator.Valid(); iterator.Next() {
		// Extract the composite key (sourceChain:sourceHash)
		fullKey := iterator.Key()
		compositeKey := string(fullKey[len(types.ProcessedSourceHashPrefix):])
		processedHashes[compositeKey] = true
	}

	return processedHashes
}

// SetProcessedSourceHash sets a processed source hash (for genesis import).
// The compositeKey should be in format "sourceChain:sourceHash".
func (k *Keeper) SetProcessedSourceHash(ctx sdk.Context, compositeKey string) {
	if compositeKey == "" {
		return
	}

	// Parse the composite key
	parts := strings.Split(compositeKey, ":")
	if len(parts) != 2 {
		return
	}

	sourceChain := strings.ToLower(strings.TrimSpace(parts[0]))
	sourceHash := strings.ToLower(strings.TrimSpace(parts[1]))

	if sourceChain == "" || sourceHash == "" {
		return
	}

	store := k.store(ctx)
	key := types.ProcessedSourceHashKey(sourceChain, sourceHash)
	store.Set(key, []byte{1})
}

// Public exported methods for external access

// SetValidator is a public method to set a bridge validator
func (k *Keeper) SetValidator(ctx sdk.Context, validator *types.BridgeValidator) error {
	return k.setValidator(ctx, validator)
}

// GetActiveValidatorSet is a public exported method for getting active validators (for tests)
func (k *Keeper) GetActiveValidatorSet(ctx sdk.Context, blockHeight int64) []string {
	return k.getActiveValidatorSet(ctx, blockHeight)
}

// ComputeSignatureSetHash is a public exported method for computing signature set hash (for tests)
func (k *Keeper) ComputeSignatureSetHash(signatures [][]byte) []byte {
	return k.computeSignatureSetHash(signatures)
}

// IsSignatureSetUsed is a public exported method for checking if signature set is used (for tests)
func (k *Keeper) IsSignatureSetUsed(ctx sdk.Context, transferID string, signatureSetHash []byte) bool {
	return k.isSignatureSetUsed(ctx, transferID, signatureSetHash)
}

// MarkSignatureSetUsed is a public exported method for marking signature set as used (for tests)
func (k *Keeper) MarkSignatureSetUsed(ctx sdk.Context, transferID string, signatureSetHash []byte) {
	k.markSignatureSetUsed(ctx, transferID, signatureSetHash)
}

// SetTransfer is a public method to set a cross-chain transfer
func (k *Keeper) SetTransfer(ctx sdk.Context, transfer *types.CrossChainTransfer) {
	if err := k.setTransfer(ctx, transfer); err != nil {
		k.Logger(ctx).Error("failed to set transfer", "transfer_id", transfer.GetTransferId(), "err", err)
	}
}

// IndexTransferHash is a public method to index a transfer by hash
func (k *Keeper) IndexTransferHash(ctx sdk.Context, hash, transferID string) {
	k.indexTransferHash(ctx, hash, transferID)
}

// getActiveValidators returns the list of currently active validators.
// Active validators are those with Active=true status.
//
// Security considerations:
//   - Only active validators should be allowed to sign bridge operations
//   - Inactive validators may have been slashed, jailed, or rotated out
//   - This function should be called at the current block height for authorization checks
//
// Returns:
//   - List of active validator addresses
func (k *Keeper) getActiveValidators(ctx sdk.Context) []string {
	allValidators := k.getAllValidators(ctx)
	activeAddresses := make([]string, 0, len(allValidators))

	for _, validator := range allValidators {
		if validator != nil && validator.Active {
			activeAddresses = append(activeAddresses, validator.Address)
		}
	}

	return activeAddresses
}

// IsValidatorActive checks if a validator is currently active and authorized.
//
// Security: This is a critical authorization check. A validator must be:
//   - Present in the validator registry
//   - Have Active=true status
//   - Not slashed or jailed (implementation can be extended)
//
// Parameters:
//   - ctx: SDK context for state access
//   - validatorAddr: Address of the validator to check
//
// Returns:
//   - bool: true if validator is active and authorized
func (k *Keeper) IsValidatorActive(ctx sdk.Context, validatorAddr string) bool {
	if validatorAddr == "" {
		return false
	}

	validator, found := k.getValidator(ctx, validatorAddr)
	if !found {
		return false
	}

	// Check if validator is active
	if !validator.Active {
		return false
	}

	// Additional checks could be added here:
	// - Check if validator is slashed
	// - Check if validator is jailed
	// - Check validator stake/power is above minimum

	return true
}

// isSignatureSetUsed checks if a specific signature set has already been used
// for a given transfer. This prevents replay attacks where an attacker reuses
// the same valid signatures to unlock tokens multiple times.
//
// Security considerations:
//   - This is CRITICAL for preventing signature replay attacks
//   - Must be checked BEFORE processing any unlock operation
//   - The signature set hash should be deterministic and cover all signatures
//
// Parameters:
//   - ctx: SDK context for state access
//   - transferID: Unique identifier of the transfer
//   - signatureSetHash: Hash of the signature set (SHA256 of concatenated signatures)
//
// Returns:
//   - bool: true if this signature set has been used before
func (k *Keeper) isSignatureSetUsed(ctx sdk.Context, transferID string, signatureSetHash []byte) bool {
	if transferID == "" || len(signatureSetHash) == 0 {
		return false
	}

	store := k.store(ctx)
	key := types.SignatureSetKey(transferID, signatureSetHash)
	return store.Has(key)
}

// markSignatureSetUsed marks a signature set as used for a given transfer.
// This prevents the same signature set from being replayed.
//
// Security considerations:
//   - Must be called AFTER successful verification but BEFORE token transfer
//   - Follows checks-effects-interactions pattern
//   - State change is permanent and irreversible
//
// Parameters:
//   - ctx: SDK context for state access
//   - transferID: Unique identifier of the transfer
//   - signatureSetHash: Hash of the signature set
func (k *Keeper) markSignatureSetUsed(ctx sdk.Context, transferID string, signatureSetHash []byte) {
	if transferID == "" || len(signatureSetHash) == 0 {
		return
	}

	store := k.store(ctx)
	key := types.SignatureSetKey(transferID, signatureSetHash)
	store.Set(key, []byte{1})

	// Emit event for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"signature_set_marked_used",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("signature_set_hash", fmt.Sprintf("%x", signatureSetHash)),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)
}

// computeSignatureSetHash computes a deterministic hash of a signature set.
// This hash is used to track which signature sets have been used.
//
// The hash is computed as: SHA256(sig1 || sig2 || ... || sigN) where signatures
// are sorted lexicographically to ensure determinism.
//
// Parameters:
//   - signatures: List of raw signature bytes
//
// Returns:
//   - []byte: SHA256 hash of the signature set
func (k *Keeper) computeSignatureSetHash(signatures [][]byte) []byte {
	if len(signatures) == 0 {
		return nil
	}

	// Sort signatures to ensure deterministic hash
	// This prevents different orderings from producing different hashes
	sortedSigs := make([][]byte, len(signatures))
	copy(sortedSigs, signatures)

	// Simple bubble sort (sufficient for small signature sets)
	for i := 0; i < len(sortedSigs)-1; i++ {
		for j := 0; j < len(sortedSigs)-i-1; j++ {
			if string(sortedSigs[j]) > string(sortedSigs[j+1]) {
				sortedSigs[j], sortedSigs[j+1] = sortedSigs[j+1], sortedSigs[j]
			}
		}
	}

	// Concatenate sorted signatures
	var combined []byte
	for _, sig := range sortedSigs {
		combined = append(combined, sig...)
	}

	// Return SHA256 hash
	hash := sha256.Sum256(combined)
	return hash[:]
}

// getActiveValidatorSet retrieves the current active validator set at a specific block height.
// This function ensures that signatures are verified against the CURRENT validator set,
// not a historical or future set.
//
// SECURITY CRITICAL: This function prevents attacks where:
//  1. Validators are compromised and then removed from the active set
//  2. Attacker replays signatures after validator set changes
//  3. Signatures from validators who are no longer active are accepted
//
// The function implements validator authorization by:
//   - Checking validator status at the current block height
//   - Ensuring only ACTIVE validators are included
//   - Filtering out slashed, jailed, or removed validators
//
// Parameters:
//   - ctx: SDK context for state access
//   - blockHeight: Block height to check validator set at (typically ctx.BlockHeight())
//
// Returns:
//   - []string: List of active validator addresses at the specified height
//
// Note: Currently uses the current validator set. In the future, this could be enhanced
// to support historical validator set snapshots if the validator set is persisted per block.
func (k *Keeper) getActiveValidatorSet(ctx sdk.Context, blockHeight int64) []string {
	// For now, return the current active validator set
	// In a production implementation with high validator rotation, you might:
	// 1. Store validator set snapshots per block
	// 2. Query historical state if available
	// 3. Use the blockHeight parameter to retrieve the exact set at that height

	// However, for security purposes, using the CURRENT active set is actually
	// more secure because it ensures signatures are from validators who are
	// CURRENTLY trusted, not validators who WERE trusted but have since been
	// slashed/removed.

	activeValidators := k.getActiveValidators(ctx)

	// Log validator set retrieval for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"active_validator_set_retrieved",
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", blockHeight)),
			sdk.NewAttribute("current_height", fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute("active_count", fmt.Sprintf("%d", len(activeValidators))),
		),
	)

	return activeValidators
}

// VerifyMerkleProofBytes verifies a Merkle proof given raw bytes.
// This is a wrapper around the existing VerifyMerkleProof that works with raw proof bytes.
//
// Security: This function verifies that a transaction is included in a block's Merkle tree.
// This prevents validators from attesting to fake deposits that never occurred on the source chain.
//
// IMPORTANT: This function expects the proof bytes to encode BOTH the sibling hashes AND
// the path indices. Each proof element consists of:
//   - 1 byte: index of the sibling (to determine left/right ordering)
//   - 32 bytes: hash of the sibling
//
// This matches the format used by GenerateMerkleProof when serializing to bytes.
//
// Parameters:
//   - merkleRoot: The Merkle root from the source block header
//   - transactionLeaf: The hash of the transaction being proven
//   - merkleProofBytes: Raw bytes of the Merkle proof (index + sibling hash per level)
//
// Returns:
//   - bool: true if the proof is valid and the transaction is in the block
func (k *Keeper) VerifyMerkleProofBytes(merkleRoot, transactionLeaf, merkleProofBytes []byte) bool {
	if len(merkleRoot) == 0 || len(transactionLeaf) == 0 {
		return false
	}

	// Special case: empty proof means single-element tree
	if len(merkleProofBytes) == 0 {
		// For a single element, the leaf hash should equal the root
		return bytes.Equal(transactionLeaf, merkleRoot)
	}

	// Parse proof bytes: each entry is 1 byte (index) + 32 bytes (hash) = 33 bytes
	// OR for backward compatibility, try 32-byte chunks (hash only)
	proofHashes := make([][]byte, 0, 64)
	var indices []uint64

	if len(merkleProofBytes)%33 == 0 {
		// New format: index + hash
		for i := 0; i < len(merkleProofBytes); i += 33 {
			idx := uint64(merkleProofBytes[i])
			// IMPORTANT: Copy the hash bytes to avoid aliasing issues
			hashCopy := make([]byte, 32)
			copy(hashCopy, merkleProofBytes[i+1:i+33])
			indices = append(indices, idx)
			proofHashes = append(proofHashes, hashCopy)
		}
	} else if len(merkleProofBytes)%32 == 0 {
		// Old format: hash only (try both orderings - brute force)
		// This is less secure but maintains backward compatibility
		for i := 0; i < len(merkleProofBytes); i += 32 {
			// IMPORTANT: Copy the hash bytes to avoid aliasing issues
			// If we just slice, the underlying array may be modified later
			hashCopy := make([]byte, 32)
			copy(hashCopy, merkleProofBytes[i:i+32])
			proofHashes = append(proofHashes, hashCopy)
		}
		// Try to verify by attempting both orderings at each level
		return k.verifyMerkleProofBruteForce(merkleRoot, transactionLeaf, proofHashes)
	} else {
		// Invalid format
		return false
	}

	// Verify using indices (new format)
	currentHash := transactionLeaf

	for i := 0; i < len(proofHashes); i++ {
		sibling := proofHashes[i]
		siblingIdx := indices[i]

		// IMPORTANT: Use explicit allocation to avoid corrupting input slices
		var combined []byte
		// The index stored is the sibling's position in the tree level.
		// If sibling index is even, sibling is on the left (before current).
		// If sibling index is odd, sibling is on the right (after current).
		if siblingIdx%2 == 1 {
			// Sibling is at odd position → sibling on RIGHT, current on LEFT
			combined = make([]byte, 0, 64)
			combined = append(combined, currentHash...)
			combined = append(combined, sibling...)
		} else {
			// Sibling is at even position → sibling on LEFT, current on RIGHT
			combined = make([]byte, 0, 64)
			combined = append(combined, sibling...)
			combined = append(combined, currentHash...)
		}

		hash := sha256.Sum256(combined)
		currentHash = hash[:]
	}

	// Check if final hash matches root
	return bytes.Equal(currentHash, merkleRoot)
}

// verifyMerkleProofBruteForce tries all possible orderings to verify a proof
// when indices are not available. This is less efficient but maintains backward
// compatibility with proofs that only contain hashes.
//
// NOTE: This approach tries 2^n possible paths for n proof elements, which is
// exponential. For production use, proofs should include indices.
func (k *Keeper) verifyMerkleProofBruteForce(merkleRoot, transactionLeaf []byte, proofHashes [][]byte) bool {
	// For small proofs (< 10 levels), try all 2^n possible orderings
	if len(proofHashes) > 10 {
		return false // Too many possibilities to brute force
	}

	// Try all possible bit patterns for left/right choices
	numPatterns := 1 << uint(len(proofHashes))
	for pattern := 0; pattern < numPatterns; pattern++ {
		currentHash := transactionLeaf

		for i := 0; i < len(proofHashes); i++ {
			sibling := proofHashes[i]

			// Use bit i of pattern to determine ordering
			// IMPORTANT: Use explicit allocation to avoid corrupting input slices
			var combined []byte
			if (pattern & (1 << uint(i))) != 0 {
				// Bit is 1: current on left, sibling on right
				combined = make([]byte, 0, 64)
				combined = append(combined, currentHash...)
				combined = append(combined, sibling...)
			} else {
				// Bit is 0: sibling on left, current on right
				combined = make([]byte, 0, 64)
				combined = append(combined, sibling...)
				combined = append(combined, currentHash...)
			}

			hash := sha256.Sum256(combined)
			currentHash = hash[:]
		}

		// Check if this ordering produces the correct root
		if bytes.Equal(currentHash, merkleRoot) {
			return true
		}
	}

	return false
}

// ConstructTransactionLeaf constructs a deterministic hash of a transaction
// that can be verified against a Merkle proof.
//
// Security: The transaction leaf must be constructed the same way on both
// the source chain and Aura to ensure proof verification works correctly.
//
// Format: SHA256(sourceChain:burnTxHash:sender:amount:denom)
//
// Parameters:
//   - sourceChain: Chain where the transaction occurred
//   - burnTxHash: Transaction hash on source chain
//   - sender: Address that initiated the transaction
//   - amount: Amount of tokens involved
//   - denom: Token denomination
//
// Returns:
//   - []byte: SHA256 hash of the transaction data
func (k *Keeper) ConstructTransactionLeaf(sourceChain, burnTxHash, sender, amount, denom string) []byte {
	// Build deterministic transaction data string
	// Format matches what source chain should use for Merkle tree construction
	txData := fmt.Sprintf("%s:%s:%s:%s:%s",
		strings.ToLower(strings.TrimSpace(sourceChain)),
		strings.ToLower(strings.TrimSpace(burnTxHash)),
		sender,
		amount,
		denom,
	)

	// Return SHA256 hash
	hash := sha256.Sum256([]byte(txData))
	return hash[:]
}

// VerifySourceBlock verifies that a source block hash is valid for a given height.
// This prevents validators from providing fake block headers.
//
// Security: In a production system, this should verify the block hash against:
//   - An oracle that tracks source chain headers
//   - A light client that maintains source chain state
//   - IBC connection if using IBC for cross-chain communication
//
// Current implementation: Stores verified block hashes in state.
// Validators must submit block headers that are verified by consensus.
//
// Parameters:
//   - ctx: SDK context for state access
//   - sourceChain: Chain identifier (e.g., "paw", "xai")
//   - blockHeight: Block height on source chain
//   - blockHash: Block hash to verify
//
// Returns:
//   - bool: true if block hash is verified for this height
func (k *Keeper) VerifySourceBlock(ctx sdk.Context, sourceChain string, blockHeight uint64, blockHash []byte) bool {
	if sourceChain == "" || blockHeight == 0 || len(blockHash) == 0 {
		return false
	}

	// Normalize chain name
	sourceChain = strings.ToLower(strings.TrimSpace(sourceChain))

	// Get stored/verified block hash for this height
	storedHash := k.GetVerifiedBlockHash(ctx, sourceChain, blockHeight)
	if storedHash == nil {
		// Block not verified yet - in production, this should trigger
		// verification via light client or oracle
		return false
	}

	// Compare provided hash with stored verified hash
	if len(storedHash) != len(blockHash) {
		return false
	}

	for i := range storedHash {
		if storedHash[i] != blockHash[i] {
			return false
		}
	}

	return true
}

// GetVerifiedBlockHash retrieves a verified block hash for a given chain and height.
//
// Storage key format: VerifiedBlockHashPrefix + sourceChain + ":" + height
//
// Parameters:
//   - ctx: SDK context for state access
//   - sourceChain: Chain identifier
//   - blockHeight: Block height
//
// Returns:
//   - []byte: Verified block hash, or nil if not found
func (k *Keeper) GetVerifiedBlockHash(ctx sdk.Context, sourceChain string, blockHeight uint64) []byte {
	sourceChain = strings.ToLower(strings.TrimSpace(sourceChain))
	if sourceChain == "" || blockHeight == 0 {
		return nil
	}

	store := k.store(ctx)
	key := types.VerifiedBlockHashKey(sourceChain, blockHeight)
	return store.Get(key)
}

// SetVerifiedBlockHash stores a verified block hash for a given chain and height.
// This should only be called after the block hash has been verified through:
//   - Light client verification
//   - Oracle consensus
//   - IBC proof verification
//
// Security: Access to this function should be restricted to authorized validators
// or governance proposals.
//
// Parameters:
//   - ctx: SDK context for state access
//   - sourceChain: Chain identifier
//   - blockHeight: Block height
//   - blockHash: Verified block hash to store
func (k *Keeper) SetVerifiedBlockHash(ctx sdk.Context, sourceChain string, blockHeight uint64, blockHash []byte) {
	sourceChain = strings.ToLower(strings.TrimSpace(sourceChain))
	if sourceChain == "" || blockHeight == 0 || len(blockHash) == 0 {
		return
	}

	store := k.store(ctx)
	key := types.VerifiedBlockHashKey(sourceChain, blockHeight)
	store.Set(key, blockHash)

	// Emit event for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"verified_block_hash_stored",
			sdk.NewAttribute("source_chain", sourceChain),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", blockHeight)),
			sdk.NewAttribute("block_hash", fmt.Sprintf("%x", blockHash)),
		),
	)
}

// GetDailyMintedAmount returns the total amount minted for a specific denom today.
//
// Security considerations:
//   - Used to enforce daily mint limits (supply cap security)
//   - Automatically resets at midnight UTC via BeginBlocker
//   - Tracked separately per denom to prevent cross-denom abuse
//
// Parameters:
//   - ctx: SDK context for state access and time
//   - denom: Token denomination to check
//
// Returns:
//   - sdkmath.Int: Total amount minted today (zero if none)
func (k *Keeper) GetDailyMintedAmount(ctx sdk.Context, denom string) sdkmath.Int {
	if denom == "" {
		return sdkmath.ZeroInt()
	}

	// Format: YYYYMMDD (e.g., "20250102")
	date := ctx.BlockTime().UTC().Format("20060102")
	store := k.store(ctx)
	key := types.DailyMintKey(date, denom)

	bz := store.Get(key)
	if bz == nil {
		return sdkmath.ZeroInt()
	}

	var amount sdkmath.Int
	if err := amount.Unmarshal(bz); err != nil {
		return sdkmath.ZeroInt()
	}

	return amount
}

// AddDailyMintedAmount adds to the daily minted amount for a specific denom.
//
// Security considerations:
//   - MUST be called AFTER successful mint operation (checks-effects-interactions)
//   - Updates persistent state to track against daily limits
//   - Separate tracking per denom prevents abuse via multiple tokens
//
// Parameters:
//   - ctx: SDK context for state access
//   - denom: Token denomination being minted
//   - amount: Amount to add to today's total
func (k *Keeper) AddDailyMintedAmount(ctx sdk.Context, denom string, amount sdkmath.Int) {
	if denom == "" || !amount.IsPositive() {
		return
	}

	date := ctx.BlockTime().UTC().Format("20060102")
	store := k.store(ctx)
	key := types.DailyMintKey(date, denom)

	current := k.GetDailyMintedAmount(ctx, denom)
	newTotal := current.Add(amount)

	bz, err := newTotal.Marshal()
	if err != nil {
		return
	}

	store.Set(key, bz)
}

// GetHourlyMintedAmount returns the total amount minted for a specific denom in the current hour.
//
// Security considerations:
//   - Used to enforce hourly rate limits (prevents rapid draining)
//   - Automatically resets each hour via BeginBlocker
//   - Tracked separately per denom
//
// Parameters:
//   - ctx: SDK context for state access and time
//   - denom: Token denomination to check
//
// Returns:
//   - sdkmath.Int: Total amount minted this hour (zero if none)
func (k *Keeper) GetHourlyMintedAmount(ctx sdk.Context, denom string) sdkmath.Int {
	if denom == "" {
		return sdkmath.ZeroInt()
	}

	// Format: YYYYMMDDHH (e.g., "2025010214" for 2PM)
	datetime := ctx.BlockTime().UTC().Format("2006010215")
	store := k.store(ctx)
	key := types.HourlyMintKey(datetime, denom)

	bz := store.Get(key)
	if bz == nil {
		return sdkmath.ZeroInt()
	}

	var amount sdkmath.Int
	if err := amount.Unmarshal(bz); err != nil {
		return sdkmath.ZeroInt()
	}

	return amount
}

// AddHourlyMintedAmount adds to the hourly minted amount for a specific denom.
//
// Security considerations:
//   - MUST be called AFTER successful mint operation
//   - Updates persistent state to track against hourly rate limits
//   - Separate tracking per denom prevents cross-token abuse
//
// Parameters:
//   - ctx: SDK context for state access
//   - denom: Token denomination being minted
//   - amount: Amount to add to this hour's total
func (k *Keeper) AddHourlyMintedAmount(ctx sdk.Context, denom string, amount sdkmath.Int) {
	if denom == "" || !amount.IsPositive() {
		return
	}

	datetime := ctx.BlockTime().UTC().Format("2006010215")
	store := k.store(ctx)
	key := types.HourlyMintKey(datetime, denom)

	current := k.GetHourlyMintedAmount(ctx, denom)
	newTotal := current.Add(amount)

	bz, err := newTotal.Marshal()
	if err != nil {
		return
	}

	store.Set(key, bz)
}

// ResetDailyMint resets daily mint counters (called in BeginBlocker at midnight UTC).
//
// Security considerations:
//   - Allows fresh daily limit after reset
//   - Only resets counters for the previous day
//   - Prevents accumulation of stale data
func (k *Keeper) ResetDailyMint(ctx sdk.Context) {
	store := k.store(ctx)
	currentDate := ctx.BlockTime().UTC().Format("20060102")

	// Iterate all daily mint keys
	iterator := store.Iterator(types.DailyMintPrefix, storetypes.PrefixEndBytes(types.DailyMintPrefix))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		// Extract the date from the key
		compositeKey := string(key[len(types.DailyMintPrefix):])
		parts := strings.Split(compositeKey, ":")
		if len(parts) == 2 {
			keyDate := parts[0]
			// Delete if not current date (cleanup old data)
			if keyDate < currentDate {
				store.Delete(key)
			}
		}
	}
}

// ResetHourlyMint resets hourly mint counters (called in BeginBlocker each hour).
//
// Security considerations:
//   - Allows fresh hourly limit after reset
//   - Only resets counters for previous hours
//   - Prevents accumulation of stale data
func (k *Keeper) ResetHourlyMint(ctx sdk.Context) {
	store := k.store(ctx)
	currentDatetime := ctx.BlockTime().UTC().Format("2006010215")

	// Iterate all hourly mint keys
	iterator := store.Iterator(types.HourlyMintPrefix, storetypes.PrefixEndBytes(types.HourlyMintPrefix))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		// Extract the datetime from the key
		compositeKey := string(key[len(types.HourlyMintPrefix):])
		parts := strings.Split(compositeKey, ":")
		if len(parts) == 2 {
			keyDatetime := parts[0]
			// Delete if not current hour (cleanup old data)
			if keyDatetime < currentDatetime {
				store.Delete(key)
			}
		}
	}
}

// ========================================================================
// PENDING TRANSFER MANAGEMENT (Fraud Proof Window)
// ========================================================================

// setPendingTransfer stores a pending transfer awaiting fraud proof window expiry.
//
// Security considerations:
//   - CRITICAL: Pending transfers are held in escrow during fraud proof window
//   - Prevents immediate token release to allow time for fraud proof challenges
//   - State change must be atomic with transfer creation
//
// Parameters:
//   - ctx: SDK context for state access
//   - pendingTransfer: The pending transfer to store
func (k *Keeper) setPendingTransfer(ctx sdk.Context, pendingTransfer *types.PendingTransfer) error {
	if pendingTransfer == nil || pendingTransfer.TransferId == "" {
		return nil
	}
	store := k.store(ctx)
	key := types.PendingTransferKey(pendingTransfer.TransferId)
	bz, err := k.cdc.Marshal(pendingTransfer)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal pending transfer %s: %v", pendingTransfer.TransferId, err)
	}
	store.Set(key, bz)
	return nil
}

// SetPendingTransfer is a public exported method for setting pending transfers (for tests).
func (k *Keeper) SetPendingTransfer(ctx sdk.Context, pendingTransfer *types.PendingTransfer) error {
	return k.setPendingTransfer(ctx, pendingTransfer)
}

// GetPendingTransfer retrieves a pending transfer by ID.
//
// Parameters:
//   - ctx: SDK context for state access
//   - transferID: Unique identifier of the transfer
//
// Returns:
//   - *PendingTransfer: The pending transfer if found
//   - bool: true if found, false otherwise
func (k *Keeper) GetPendingTransfer(ctx sdk.Context, transferID string) (*types.PendingTransfer, bool) {
	if transferID == "" {
		return nil, false
	}
	store := k.store(ctx)
	key := types.PendingTransferKey(transferID)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var pending types.PendingTransfer
	if err := k.cdc.Unmarshal(bz, &pending); err != nil {
		return nil, false
	}
	return &pending, true
}

// deletePendingTransfer removes a pending transfer from storage.
//
// Security considerations:
//   - Called after finalization or cancellation
//   - Ensures cleanup of completed transfers
//   - Prevents double-finalization
//
// Parameters:
//   - ctx: SDK context for state access
//   - transferID: Unique identifier of the transfer to delete
func (k *Keeper) deletePendingTransfer(ctx sdk.Context, transferID string) {
	if transferID == "" {
		return
	}
	store := k.store(ctx)
	key := types.PendingTransferKey(transferID)
	store.Delete(key)
}

// GetAllPendingTransfers retrieves all pending transfers.
//
// Returns:
//   - []*PendingTransfer: List of all pending transfers
func (k *Keeper) GetAllPendingTransfers(ctx sdk.Context) []*types.PendingTransfer {
	store := k.store(ctx)
	iterator := store.Iterator(types.PendingTransferPrefix, storetypes.PrefixEndBytes(types.PendingTransferPrefix))
	defer iterator.Close()

	pendingTransfers := make([]*types.PendingTransfer, 0, 64)
	for ; iterator.Valid(); iterator.Next() {
		var pending types.PendingTransfer
		if err := k.cdc.Unmarshal(iterator.Value(), &pending); err != nil {
			// Log corrupted data but continue iteration
			k.Logger(ctx).Error("failed to unmarshal pending transfer",
				"key", hex.EncodeToString(iterator.Key()),
				"error", err.Error())
			continue
		}
		pendingCopy := pending
		pendingTransfers = append(pendingTransfers, &pendingCopy)
	}
	return pendingTransfers
}

// IsPendingTransferExpired checks if a pending transfer's fraud proof window has expired.
//
// Security considerations:
//   - CRITICAL: Must only return true after fraud proof window has fully elapsed
//   - Used to determine if transfer can be finalized
//   - Time-based security check
//
// Parameters:
//   - ctx: SDK context for current block time
//   - pendingTransfer: The pending transfer to check
//
// Returns:
//   - bool: true if window has expired and transfer can be finalized
func (k *Keeper) IsPendingTransferExpired(ctx sdk.Context, pendingTransfer *types.PendingTransfer) bool {
	if pendingTransfer == nil || pendingTransfer.UnlockTime.IsZero() {
		return false
	}
	return ctx.BlockTime().After(pendingTransfer.UnlockTime) ||
		ctx.BlockTime().Equal(pendingTransfer.UnlockTime)
}

// MarkPendingTransferChallenged marks a pending transfer as challenged with a fraud proof.
//
// Security considerations:
//   - CRITICAL: Prevents finalization of challenged transfers
//   - Must be called when fraud proof is submitted
//   - State change is permanent until fraud proof is resolved
//
// Parameters:
//   - ctx: SDK context for state access
//   - transferID: Unique identifier of the transfer
//   - fraudProofID: ID of the fraud proof challenging this transfer
//
// Returns:
//   - error: If transfer not found or already challenged
func (k *Keeper) MarkPendingTransferChallenged(ctx sdk.Context, transferID string, fraudProofID string) error {
	pending, found := k.GetPendingTransfer(ctx, transferID)
	if !found {
		return types.ErrTransferNotFound
	}

	if pending.Challenged {
		return fmt.Errorf("transfer already challenged with fraud proof %s", pending.FraudProofId)
	}

	pending.Challenged = true
	pending.FraudProofId = fraudProofID
	if err := k.setPendingTransfer(ctx, pending); err != nil {
		return fmt.Errorf("failed to mark pending transfer as challenged: %w", err)
	}

	// Emit event for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"pending_transfer_challenged",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("fraud_proof_id", fraudProofID),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return nil
}

// FinalizeTransfer completes a pending transfer after fraud proof window expires.
//
// SECURITY CRITICAL: This function releases tokens to the recipient ONLY if:
//  1. The fraud proof window has fully elapsed
//  2. No valid fraud proof has been submitted against the transfer
//  3. All security checks pass (replay protection, supply caps, etc.)
//
// This implements the fraud proof challenge period mechanism:
//   - Transfers are held in pending state during the window
//   - Anyone can submit fraud proofs during this time
//   - Only after window expiry with no valid challenges can tokens be released
//
// Parameters:
//   - ctx: SDK context for state access and time checks
//   - transferID: Unique identifier of the transfer to finalize
//
// Returns:
//   - error: If transfer cannot be finalized (window not expired, challenged, not found, etc.)
func (k *Keeper) FinalizeTransfer(ctx sdk.Context, transferID string) error {
	// 1. Retrieve pending transfer
	pending, found := k.GetPendingTransfer(ctx, transferID)
	if !found {
		return types.ErrTransferNotFound
	}

	// 2. CRITICAL SECURITY: Check fraud proof window has expired
	if !k.IsPendingTransferExpired(ctx, pending) {
		return fmt.Errorf("fraud proof window not expired: unlocks at %s, current time %s",
			pending.UnlockTime, ctx.BlockTime())
	}

	// 3. CRITICAL SECURITY: Check if challenged by fraud proof
	if pending.Challenged {
		return fmt.Errorf("transfer challenged with fraud proof %s - cannot finalize",
			pending.FraudProofId)
	}

	// 4. Parse recipient address
	recipient, err := sdk.AccAddressFromBech32(pending.Recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}

	// 5. pending.Amount is already sdkmath.Int, use directly
	amount := pending.Amount

	// Create coin to mint
	coin := sdk.NewCoin(pending.Denom, amount)

	// 6. CRITICAL SECURITY: Perform final security checks before minting
	// These checks should have been done when creating the pending transfer,
	// but we verify again in case parameters changed during the fraud proof window
	params := k.GetParams(ctx)

	// Check per-transfer maximum
	maxTransfer, ok := sdkmath.NewIntFromString(params.MaxTransferAmount)
	if ok && amount.GT(maxTransfer) {
		return fmt.Errorf("amount %s exceeds max transfer limit %s", amount, maxTransfer)
	}

	// Check per-token supply cap (if configured)
	if cap, exists := params.SupplyCaps[pending.Denom]; exists {
		supplyCap, ok := sdkmath.NewIntFromString(cap)
		if ok {
			currentSupply := k.bankKeeper.GetSupply(ctx, pending.Denom).Amount
			if currentSupply.Add(amount).GT(supplyCap) {
				return fmt.Errorf("minting would exceed supply cap of %s (current: %s)",
					supplyCap, currentSupply)
			}
		}
	}

	// 7. Unlock or mint tokens (following checks-effects-interactions)
	// Determine if this is an unlock (native tokens) or mint (wrapped tokens)
	// Wrapped tokens have format "chain.denom" (e.g., "paw.token", "xai.coin")
	// Native tokens are unlocked from module, wrapped tokens are minted
	isWrapped := strings.Contains(pending.Denom, ".")

	if k.bankKeeper != nil {
		if isWrapped {
			// WRAPPED TOKENS: Mint new wrapped tokens
			// Mint to module
			if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(coin)); err != nil {
				return fmt.Errorf("failed to mint wrapped tokens: %w", err)
			}

			// Send to recipient
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(coin)); err != nil {
				return fmt.Errorf("failed to send wrapped tokens to recipient: %w", err)
			}
		} else {
			// NATIVE TOKENS: Unlock from module account (already locked from source chain)
			// The tokens were locked in the module when LockTokens was called on the source
			// We just need to send them from module to recipient
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(coin)); err != nil {
				return fmt.Errorf("failed to unlock native tokens: %w", err)
			}
		}
	}

	// 8. Update transfer status to completed
	transfer, found := k.getTransfer(ctx, transferID)
	if found {
		transfer.Status = types.TransferStatus_COMPLETED
		transfer.Timestamp = ctx.BlockTime()
		if err := k.setTransfer(ctx, transfer); err != nil {
			return fmt.Errorf("failed to persist finalized transfer: %w", err)
		}
	}

	// 9. Delete pending transfer (cleanup)
	k.deletePendingTransfer(ctx, transferID)

	// 10. Record minted amounts for rate limiting
	k.AddDailyMintedAmount(ctx, pending.Denom, amount)
	k.AddHourlyMintedAmount(ctx, pending.Denom, amount)

	// 11. Emit completion event for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"transfer_finalized",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("recipient", pending.Recipient),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("denom", pending.Denom),
			sdk.NewAttribute("source_chain", pending.SourceChain),
			sdk.NewAttribute("source_tx_hash", pending.SourceTxHash),
			sdk.NewAttribute("finalized_at_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	return nil
}

// ProcessExpiredPendingTransfers automatically finalizes pending transfers whose
// fraud proof window has expired. This function is called from EndBlock.
//
// SECURITY CRITICAL: This implements the automated completion of transfers after
// the fraud proof window expires, ensuring users receive their funds without
// requiring manual finalization.
//
// Process:
//  1. Retrieve all pending transfers from state
//  2. Check each transfer's unlock time against current block time
//  3. Skip challenged transfers (require governance resolution)
//  4. Finalize unchallenged expired transfers
//  5. Log all operations for audit trail
//
// Security considerations:
//   - Only processes transfers with expired fraud proof windows
//   - Respects fraud proofs (challenged transfers are not finalized)
//   - Uses existing FinalizeTransfer which has full security checks
//   - Gas-limited to prevent chain halt (skips to next if one fails)
//
// Parameters:
//   - ctx: SDK context for current block time and state access
func (k *Keeper) ProcessExpiredPendingTransfers(ctx sdk.Context) {
	// Get all pending transfers
	allPending := k.GetAllPendingTransfers(ctx)

	if len(allPending) == 0 {
		// No pending transfers to process
		return
	}

	// Track processing statistics for monitoring
	var processed, finalized, challenged, errors, skipped int

	// SECURITY: Limit processing to MaxPendingTransfersPerBlock to prevent
	// unbounded O(n) iteration that could cause chain halts under high load.
	// Remaining transfers will be processed in subsequent blocks.
	maxToProcess := types.MaxPendingTransfersPerBlock

	// Process each pending transfer (up to batch limit)
	for _, pending := range allPending {
		// Check batch limit - only count actual processing attempts
		if processed >= maxToProcess {
			skipped = len(allPending) - processed
			break
		}

		processed++

		// Check if transfer has been challenged
		if pending.Challenged {
			challenged++
			// Log challenged transfer for monitoring (not an error)
			k.Logger(ctx).Info("skipping challenged pending transfer",
				"transfer_id", pending.TransferId,
				"fraud_proof_id", pending.FraudProofId,
				"unlock_time", pending.UnlockTime)
			continue
		}

		// Check if fraud proof window has expired
		if !k.IsPendingTransferExpired(ctx, pending) {
			// Not expired yet - skip silently (this is normal)
			continue
		}

		// Fraud proof window has expired and no challenge - finalize the transfer
		if err := k.FinalizeTransfer(ctx, pending.TransferId); err != nil {
			errors++
			// Log error but continue processing other transfers
			// We don't want one failed transfer to block all others
			k.Logger(ctx).Error("failed to auto-finalize expired pending transfer",
				"transfer_id", pending.TransferId,
				"recipient", pending.Recipient,
				"amount", pending.Amount,
				"denom", pending.Denom,
				"unlock_time", pending.UnlockTime,
				"error", err.Error())
			continue
		}

		finalized++

		// Log successful finalization
		k.Logger(ctx).Info("auto-finalized expired pending transfer",
			"transfer_id", pending.TransferId,
			"recipient", pending.Recipient,
			"amount", pending.Amount,
			"denom", pending.Denom,
			"unlock_time", pending.UnlockTime)
	}

	// Emit summary event if any transfers were processed
	if finalized > 0 || errors > 0 || challenged > 0 || skipped > 0 {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"pending_transfers_processed",
				sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
				sdk.NewAttribute("total_pending", fmt.Sprintf("%d", len(allPending))),
				sdk.NewAttribute("processed", fmt.Sprintf("%d", processed)),
				sdk.NewAttribute("finalized", fmt.Sprintf("%d", finalized)),
				sdk.NewAttribute("challenged", fmt.Sprintf("%d", challenged)),
				sdk.NewAttribute("errors", fmt.Sprintf("%d", errors)),
				sdk.NewAttribute("skipped_batch_limit", fmt.Sprintf("%d", skipped)),
			),
		)
	}
}

// ========================================================================
// CRYPTOGRAPHIC HELPER FUNCTIONS FOR SIGNATURE VERIFICATION
// ========================================================================

// recoverPubKeyFromSignature recovers a public key from a message hash and signature.
//
// SECURITY CRITICAL: This function performs ECDSA public key recovery using secp256k1.
// It is used to verify that a signature was created by the holder of a specific private key.
//
// Process:
//  1. Parse R and S components from signature bytes
//  2. Use recovery ID to determine which of the 4 possible public keys is correct
//  3. Recover the public key using secp256k1 curve parameters
//  4. Return compressed public key (33 bytes)
//
// Parameters:
//   - msgHash: SHA256 hash of the signed message (32 bytes)
//   - signature: R and S components of signature (64 bytes)
//   - recoveryID: Recovery ID (0-3) indicating which public key to recover
//
// Returns:
//   - []byte: Compressed public key (33 bytes) on success
//   - error: If recovery fails (invalid signature, invalid recovery ID, etc.)
func (k *Keeper) recoverPubKeyFromSignature(msgHash []byte, signature []byte, recoveryID byte) ([]byte, error) {
	// Validate input lengths
	if len(msgHash) != 32 {
		return nil, fmt.Errorf("invalid message hash length: expected 32, got %d", len(msgHash))
	}
	if len(signature) != 64 {
		return nil, fmt.Errorf("invalid signature length: expected 64, got %d", len(signature))
	}
	if recoveryID > 7 {
		return nil, fmt.Errorf("invalid recovery ID: %d (must be 0-7)", recoveryID)
	}

	// Parse R and S from signature
	rBytes := signature[:32]
	sBytes := signature[32:64]

	// Recover public key using the library's RecoverCompact function
	// The recovery ID indicates which of 4 possible public keys is correct
	// Format recovery byte: 27 + recoveryID (for compatibility)
	recoveryByte := byte(27 + recoveryID)

	// Combine recovery byte with R and S for RecoverCompact format
	// RecoverCompact expects: [recovery_byte][R (32 bytes)][S (32 bytes)]
	compactSig := make([]byte, 65)
	compactSig[0] = recoveryByte
	copy(compactSig[1:33], rBytes)
	copy(compactSig[33:65], sBytes)

	// Recover the public key
	pubKey, _, err := ecdsa.RecoverCompact(compactSig, msgHash)
	if err != nil {
		return nil, fmt.Errorf("failed to recover public key: %w", err)
	}

	// Return compressed public key (33 bytes)
	return pubKey.SerializeCompressed(), nil
}

// deriveExternalAddressFromPubKey derives an external chain address from a public key.
//
// Address derivation process for Cosmos SDK-compatible chains:
//  1. Take the compressed public key (33 bytes)
//  2. Hash with SHA256
//  3. Hash with RIPEMD160
//  4. Encode with Bech32 using chain-specific prefix
//
// For simplicity, this implementation returns the hex-encoded hash.
// In production, you would use the full Bech32 encoding with proper prefix.
//
// Parameters:
//   - pubKey: Compressed secp256k1 public key (33 bytes)
//   - chainName: Name of the chain (e.g., "paw", "xai") - currently unused but reserved for future Bech32 encoding
//
// Returns:
//   - string: Hex-encoded address (for testing) or Bech32 address (for production)
func (k *Keeper) deriveExternalAddressFromPubKey(pubKey []byte, chainName string) string {
	if len(pubKey) != 33 {
		return ""
	}

	// Step 1: SHA256 hash of public key
	sha256Hash := sha256.Sum256(pubKey)

	// Step 2: RIPEMD160 hash of SHA256 hash
	ripemd160Hasher := ripemd160.New() // #nosec G406 -- RIPEMD160 required for Cosmos-style address derivation compatibility
	ripemd160Hasher.Write(sha256Hash[:])
	addressHash := ripemd160Hasher.Sum(nil)

	// For production, encode as Bech32 with chain-specific prefix
	// For now, return hex-encoded for testing compatibility
	// In a full implementation, use:
	// addr, _ := bech32.ConvertAndEncode(chainName, addressHash)
	// return addr

	// Note: chainName parameter is currently unused but reserved for future Bech32 encoding
	_ = chainName

	// Return hex-encoded address (20 bytes = 40 hex chars)
	return hex.EncodeToString(addressHash)
}

// derivePawAddressFromPubKey derives a PAW (Cosmos SDK) address from a public key.
// This is a thin wrapper around deriveExternalAddressFromPubKey for backwards compatibility.
//
// Deprecated: Use deriveExternalAddressFromPubKey directly for new code.
func (k *Keeper) derivePawAddressFromPubKey(pubKey []byte) string {
	return k.deriveExternalAddressFromPubKey(pubKey, "paw")
}

// deriveXaiAddressFromPubKey derives an XAI (Cosmos SDK) address from a public key.
// This is a thin wrapper around deriveExternalAddressFromPubKey for backwards compatibility.
//
// Deprecated: Use deriveExternalAddressFromPubKey directly for new code.
func (k *Keeper) deriveXaiAddressFromPubKey(pubKey []byte) string {
	return k.deriveExternalAddressFromPubKey(pubKey, "xai")
}

// normalizeLowS normalizes a signature to low-S form for Cosmos SDK compatibility.
//
// SECURITY CRITICAL: The Cosmos SDK rejects high-S signatures for malleability protection.
// For secp256k1, a signature (R, S) has a malleability counterpart (R, N-S) where N is the curve order.
// To prevent malleability, only one form is accepted: the one where S is in the lower half of the curve order.
//
// If S > N/2, we normalize it to S' = N - S.
//
// Parameters:
//   - signature: R and S components (64 bytes: R[32] || S[32])
//
// Returns:
//   - []byte: Normalized signature (64 bytes) in low-S form
func (k *Keeper) normalizeLowS(signature []byte) []byte {
	if len(signature) != 64 {
		return signature // Invalid length, return as-is
	}

	// Secp256k1 curve order (N)
	// N = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
	curveOrderBytes := []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE,
		0xBA, 0xAE, 0xDC, 0xE6, 0xAF, 0x48, 0xA0, 0x3B,
		0xBF, 0xD2, 0x5E, 0x8C, 0xD0, 0x36, 0x41, 0x41,
	}

	// Half of curve order (N/2)
	halfOrderBytes := []byte{
		0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0x5D, 0x57, 0x6E, 0x73, 0x57, 0xA4, 0x50, 0x1D,
		0xDF, 0xE9, 0x2F, 0x46, 0x68, 0x1B, 0x20, 0xA0,
	}

	// Extract S component (bytes 32-63)
	sBytes := signature[32:64]

	// Check if S > N/2 by comparing bytes
	isHighS := false
	for i := 0; i < 32; i++ {
		if sBytes[i] > halfOrderBytes[i] {
			isHighS = true
			break
		} else if sBytes[i] < halfOrderBytes[i] {
			break
		}
	}

	// If S is already low, return original signature
	if !isHighS {
		result := make([]byte, 64)
		copy(result, signature)
		return result
	}

	// Normalize: S' = N - S
	normalized := make([]byte, 64)
	copy(normalized[0:32], signature[0:32]) // Keep R unchanged

	// Compute N - S using big integer subtraction
	sInt := new(big.Int)
	sInt.SetBytes(sBytes)

	nInt := new(big.Int)
	nInt.SetBytes(curveOrderBytes)

	sPrime := new(big.Int)
	sPrime.Sub(nInt, sInt)

	// Convert S' back to 32-byte big-endian
	sPrimeBytes := sPrime.Bytes()
	// Pad with leading zeros if necessary
	if len(sPrimeBytes) < 32 {
		copy(normalized[32+32-len(sPrimeBytes):64], sPrimeBytes)
	} else {
		copy(normalized[32:64], sPrimeBytes[len(sPrimeBytes)-32:])
	}

	return normalized
}

// verifyEcdsaSignature verifies an ECDSA signature using the secp256k1 curve.
//
// SECURITY CRITICAL: This function verifies signatures from external wallets (PAW, XAI)
// which use standard ECDSA secp256k1 signatures. These signatures may not be compatible
// with Cosmos SDK's native secp256k1 implementation due to different encoding formats.
//
// Process:
//  1. Parse R and S components from signature bytes
//  2. Create secp256k1.PublicKey from compressed public key bytes
//  3. Create ecdsa.Signature from R and S
//  4. Verify signature against message hash
//
// Parameters:
//   - pubKeyBytes: Compressed public key (33 bytes)
//   - msgHash: SHA256 hash of the message (32 bytes)
//   - signature: R and S components (64 bytes: R[32] || S[32])
//
// Returns:
//   - true if signature is cryptographically valid
//   - false if signature is invalid or parameters are malformed
func (k *Keeper) verifyEcdsaSignature(pubKeyBytes []byte, msgHash []byte, signature []byte) bool {
	// Validate input lengths
	if len(pubKeyBytes) != 33 {
		return false
	}
	if len(msgHash) != 32 {
		return false
	}
	if len(signature) != 64 {
		return false
	}

	// Parse compressed public key
	pubKey, err := secp256k1Curve.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false
	}

	// Extract R and S components
	rBytes := signature[:32]
	sBytes := signature[32:64]

	// Create ModNScalar values for R and S
	var r, s secp256k1Curve.ModNScalar
	if overflow := r.SetByteSlice(rBytes); overflow {
		return false
	}
	if overflow := s.SetByteSlice(sBytes); overflow {
		return false
	}

	// Create ECDSA signature
	sig := ecdsa.NewSignature(&r, &s)

	// Verify signature
	return sig.Verify(msgHash, pubKey)
}

// ========================================================================
// SIGNATURE REPLAY PROTECTION AND RATE LIMITING
// ========================================================================

// isSignatureUsed checks if a signature has already been used.
//
// SECURITY CRITICAL: This function prevents replay attacks by tracking used signatures.
// Each signature can only be accepted once, even if it's cryptographically valid.
//
// The signature hash is stored in the KV store after successful verification.
// This prevents an attacker from reusing the same signature to perform the same
// action multiple times (e.g., linking the same address multiple times).
//
// Parameters:
//   - ctx: SDK context for state access
//   - signatureHash: SHA256 hash of the complete signature (32 bytes)
//
// Returns:
//   - bool: true if signature has been used before, false if it's fresh
func (k *Keeper) isSignatureUsed(ctx sdk.Context, signatureHash []byte) bool {
	if len(signatureHash) != 32 {
		return true // Invalid hash length, reject to be safe
	}

	store := k.store(ctx)
	key := types.SignatureUsedKey(signatureHash)
	return store.Has(key)
}

// markSignatureUsed marks a signature as used to prevent replay attacks.
//
// SECURITY CRITICAL: This function stores the signature hash in the KV store
// along with the block height at which it was used. This creates an immutable
// audit trail of signature usage.
//
// Once marked as used, the signature will be rejected by isSignatureUsed(),
// preventing replay attacks.
//
// Storage format:
//   - Key: SignatureUsedPrefix + signatureHash (32 bytes)
//   - Value: blockHeight (8 bytes, big-endian uint64)
//
// Parameters:
//   - ctx: SDK context for state access
//   - signatureHash: SHA256 hash of the complete signature (32 bytes)
//   - blockHeight: Block height at which signature was verified
//
// Implementation note: The block height is stored for potential future cleanup
// of old signature records, though this is not currently implemented to maintain
// maximum security (no automatic deletion of replay protection data).
func (k *Keeper) markSignatureUsed(ctx sdk.Context, signatureHash []byte, blockHeight int64) {
	if len(signatureHash) != 32 {
		// Invalid hash length, log error but don't panic
		// This should never happen if called correctly
		ctx.Logger().Error("Attempted to mark invalid signature hash",
			"hash_length", len(signatureHash),
			"expected", 32)
		return
	}

	store := k.store(ctx)
	key := types.SignatureUsedKey(signatureHash)

	// Store block height as 8-byte big-endian uint64
	value := make([]byte, 8)
	binary.BigEndian.PutUint64(value, uint64(blockHeight))

	store.Set(key, value)

	// Emit event for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"signature_marked_used",
			sdk.NewAttribute("signature_hash", hex.EncodeToString(signatureHash)),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", blockHeight)),
		),
	)
}

// checkSignatureRateLimit enforces rate limiting on signature verification attempts.
//
// SECURITY CRITICAL: This function prevents DoS attacks where an attacker floods
// the system with signature verification requests, consuming computational resources.
//
// Rate limiting is enforced per address with a sliding window approach:
//   - Window size: 100 blocks (~10 minutes at 6s block time)
//   - Max attempts per window: 10 signatures
//
// The rate limit is tracked separately from signature replay protection because:
//  1. It tracks attempts, not just successful verifications
//  2. It uses a sliding window that can be cleaned up
//  3. It provides DoS protection even for invalid signatures
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: Address attempting signature verification
//
// Returns:
//   - error: ErrSignatureRateLimit if rate limit exceeded, nil otherwise
func (k *Keeper) checkSignatureRateLimit(ctx sdk.Context, address string) error {
	const (
		// Rate limit parameters
		maxAttemptsPerWindow = 10  // Maximum signature attempts per window
		rateLimitWindowSize  = 100 // Window size in blocks (~10 minutes)
	)

	currentHeight := ctx.BlockHeight()
	windowStart := (currentHeight / rateLimitWindowSize) * rateLimitWindowSize

	store := k.store(ctx)
	key := types.SignatureRateLimitKey(address, windowStart)

	// Get current attempt count for this window
	var attemptCount uint64
	bz := store.Get(key)
	if bz != nil && len(bz) == 8 {
		attemptCount = binary.BigEndian.Uint64(bz)
	}

	// Check if rate limit exceeded
	if attemptCount >= maxAttemptsPerWindow {
		ctx.Logger().Warn("Signature verification rate limit exceeded",
			"address", address,
			"window_start", windowStart,
			"current_height", currentHeight,
			"attempts", attemptCount,
			"max_allowed", maxAttemptsPerWindow)

		return types.ErrSignatureRateLimit
	}

	// Increment attempt count
	attemptCount++
	newValue := make([]byte, 8)
	binary.BigEndian.PutUint64(newValue, attemptCount)
	store.Set(key, newValue)

	// Log rate limit check for monitoring
	if attemptCount > maxAttemptsPerWindow/2 {
		ctx.Logger().Info("Signature rate limit warning",
			"address", address,
			"attempts", attemptCount,
			"max_allowed", maxAttemptsPerWindow,
			"window_blocks_remaining", rateLimitWindowSize-(currentHeight-windowStart))
	}

	return nil
}
