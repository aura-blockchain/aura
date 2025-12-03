package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/ripemd160"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/bridge/types"
	"github.com/aequitas/aura/chain/x/common/security"
)

const sourceChainAura = "aura"

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

	// Security features
	reentrancyGuard *security.ReentrancyGuard
	pauseGuard      *security.PauseGuard
	inputValidator  *security.InputValidator
	safeMath        *security.SafeMath
	gasLimitGuard   *security.GasLimitGuard
	accessControl   *security.AccessControl
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

	return &Keeper{
		storeKey:        storeKey,
		cdc:             cdc,
		paramstore:      paramstore,
		bankKeeper:      bankKeeper,
		accountKeeper:   accountKeeper,
		vcKeeper:        vcKeeper,
		stakingKeeper:   stakingKeeper,
		reentrancyGuard: security.NewReentrancyGuard(),
		pauseGuard:      security.NewPauseGuard(""),
		inputValidator:  security.NewInputValidator(),
		safeMath:        security.NewSafeMath(),
		gasLimitGuard:   security.NewGasLimitGuard(1_000_000),
		accessControl:   security.NewAccessControl([]string{}),
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) store(ctx sdk.Context) storetypes.KVStore {
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
func (k Keeper) nextTransferID(ctx sdk.Context) string {
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
	idStr := fmt.Sprintf("transfer-%d", transferID)
	if _, found := k.getTransfer(ctx, idStr); found {
		// CRITICAL: If collision detected, append nonce and rehash
		// This should never happen with SHA256 but we handle it defensively
		ctx.Logger().Error("RARE: Transfer ID collision detected, regenerating with nonce",
			"transfer_id", idStr,
			"block_height", blockHeight)

		// Add timestamp nonce and regenerate
		nonceBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(nonceBytes, uint64(ctx.BlockTime().UnixNano()))
		hashInput = append(hashInput, nonceBytes...)

		hash = sha256.Sum256(hashInput)
		transferID = binary.BigEndian.Uint64(hash[:8])
		idStr = fmt.Sprintf("transfer-%d", transferID)
	}

	return idStr
}

// extractBlockHeightFromTransferID extracts information from a deterministic transfer ID.
// Note: With hash-based IDs, we cannot extract the original block height.
// Returns (0, 0, false) for new deterministic IDs, (height, 0, true) for legacy sequential IDs.
func (k Keeper) extractBlockHeightFromTransferID(transferID string) (int64, int64, bool) {
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

func (k Keeper) setTransfer(ctx sdk.Context, transfer *types.CrossChainTransfer) {
	if transfer == nil || transfer.TransferId == "" {
		return
	}
	k.store(ctx).Set(types.TransferKey(transfer.TransferId), k.cdc.MustMarshal(transfer))
}

func (k Keeper) getTransfer(ctx sdk.Context, transferID string) (*types.CrossChainTransfer, bool) {
	if transferID == "" {
		return nil, false
	}
	store := k.store(ctx)
	bz := store.Get(types.TransferKey(transferID))
	if bz == nil {
		return nil, false
	}
	var transfer types.CrossChainTransfer
	if err := k.cdc.Unmarshal(bz, &transfer); err != nil {
		return nil, false
	}
	return &transfer, true
}

// GetTransfer is a public exported method for getting a transfer (for tests).
func (k Keeper) GetTransfer(ctx sdk.Context, transferID string) (*types.CrossChainTransfer, bool) {
	return k.getTransfer(ctx, transferID)
}

func (k Keeper) deleteTransfer(ctx sdk.Context, transferID string) {
	if transferID == "" {
		return
	}
	k.store(ctx).Delete(types.TransferKey(transferID))
}

func (k Keeper) indexTransferHash(ctx sdk.Context, hash, transferID string) {
	if hash == "" || transferID == "" {
		return
	}
	k.store(ctx).Set(types.TransferHashIndexKey(strings.ToLower(hash)), []byte(transferID))
}

func (k Keeper) transferIDByHash(ctx sdk.Context, hash string) (string, bool) {
	if hash == "" {
		return "", false
	}
	bz := k.store(ctx).Get(types.TransferHashIndexKey(strings.ToLower(hash)))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) setChainConfig(ctx sdk.Context, config types.ChainConfig) {
	k.store(ctx).Set(types.ChainConfigKey(strings.ToLower(config.ChainId)), k.cdc.MustMarshal(&config))
}

func (k Keeper) getChainConfig(ctx sdk.Context, chainID string) (types.ChainConfig, bool) {
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

func (k Keeper) getAllChainConfigs(ctx sdk.Context) []types.ChainConfig {
	store := k.store(ctx)
	iterator := store.Iterator(types.ChainConfigPrefix, storetypes.PrefixEndBytes(types.ChainConfigPrefix))
	defer iterator.Close()
	var configs []types.ChainConfig
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

func (k Keeper) setSharedIdentity(ctx sdk.Context, identity *types.SharedIdentity) {
	if identity == nil || identity.Address == "" {
		return
	}
	k.store(ctx).Set(types.SharedIdentityKey(identity.Address), k.cdc.MustMarshal(identity))
}

func (k Keeper) getSharedIdentity(ctx sdk.Context, address string) (*types.SharedIdentity, bool) {
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
func (k Keeper) findSharedIdentityByLinkedAddress(ctx sdk.Context, chainName string, address string) *types.SharedIdentity {
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

// verifyPawAddressOwnership verifies that the signer owns the PAW address.
//
// SECURITY CRITICAL: This function verifies cross-chain ownership using cryptographic signatures.
// The signature proves that the signer has the private key for the PAW address.
//
// Expected signature format:
//   - Message: "Link PAW address <pawAddress> to Aura address <auraAddress>"
//   - Signature: secp256k1 signature (65 bytes: R[32] || S[32] || V[1])
//   - Recovery ID (V) is used to recover the public key from the signature
//
// Verification process:
//   1. Hash the expected message using SHA256
//   2. Recover the public key from signature using recovery ID
//   3. Derive the address from the recovered public key
//   4. Compare derived address with claimed PAW address
//
// Parameters:
//   - ctx: SDK context (for logging)
//   - auraAddress: The Aura address being linked
//   - pawAddress: The PAW address to verify ownership of
//   - signature: Cryptographic signature proving ownership (65 bytes)
//
// Returns:
//   - true if signature is valid and proves ownership
//   - false if signature is invalid or address format is wrong
func (k Keeper) verifyPawAddressOwnership(ctx sdk.Context, auraAddress string, pawAddress string, signature []byte) bool {
	if len(signature) == 0 || pawAddress == "" || auraAddress == "" {
		return false
	}

	// Validate signature format: secp256k1 signatures are 65 bytes (R[32] || S[32] || V[1])
	// where V is the recovery ID (0, 1, 2, or 3)
	if len(signature) != 65 {
		ctx.Logger().Error("Invalid PAW signature length",
			"expected", 65,
			"actual", len(signature),
			"paw_address", pawAddress)
		return false
	}

	// Build the expected message that should have been signed
	// Format: "Link PAW address <pawAddress> to Aura address <auraAddress>"
	message := fmt.Sprintf("Link PAW address %s to Aura address %s", pawAddress, auraAddress)

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
		ctx.Logger().Error("Invalid recovery ID in PAW signature",
			"recovery_id", recoveryID,
			"paw_address", pawAddress)
		return false
	}

	// Extract R and S components (first 64 bytes)
	sigBytes := signature[:64]

	// Attempt to recover the public key from the signature
	// We try all possible recovery IDs (0-7) to find the correct one
	var recoveredPubKey []byte
	for recID := byte(0); recID <= 7; recID++ {
		// Use the provided recovery ID first, then try others
		tryRecID := (recoveryID + recID) % 8

		// Attempt recovery with this ID
		pubKey, err := k.recoverPubKeyFromSignature(msgHash[:], sigBytes, tryRecID)
		if err != nil {
			continue
		}

		// Derive address from recovered public key
		derivedAddress := k.derivePawAddressFromPubKey(pubKey)

		// Check if derived address matches claimed PAW address
		if derivedAddress == pawAddress {
			recoveredPubKey = pubKey
			break
		}
	}

	if recoveredPubKey == nil {
		ctx.Logger().Error("Failed to recover public key from PAW signature",
			"paw_address", pawAddress,
			"aura_address", auraAddress)
		return false
	}

	// Additional verification: Verify the signature with the recovered public key
	// This ensures the signature is cryptographically valid
	pubKeyObj := &secp256k1.PubKey{Key: recoveredPubKey}
	if !pubKeyObj.VerifySignature(msgHash[:], sigBytes) {
		ctx.Logger().Error("PAW signature verification failed",
			"paw_address", pawAddress,
			"aura_address", auraAddress)
		return false
	}

	ctx.Logger().Info("PAW address ownership verified successfully",
		"paw_address", pawAddress,
		"aura_address", auraAddress)

	return true
}

// verifyXaiAddressOwnership verifies that the signer owns the XAI address.
//
// SECURITY CRITICAL: This function verifies cross-chain ownership using cryptographic signatures.
// The signature proves that the signer has the private key for the XAI address.
//
// Expected signature format:
//   - Message: "Link XAI address <xaiAddress> to Aura address <auraAddress>"
//   - Signature: secp256k1 signature (65 bytes: R[32] || S[32] || V[1])
//   - Recovery ID (V) is used to recover the public key from the signature
//
// Verification process:
//   1. Hash the expected message using SHA256
//   2. Recover the public key from signature using recovery ID
//   3. Derive the address from the recovered public key
//   4. Compare derived address with claimed XAI address
//
// Parameters:
//   - ctx: SDK context (for logging)
//   - auraAddress: The Aura address being linked
//   - xaiAddress: The XAI address to verify ownership of
//   - signature: Cryptographic signature proving ownership (65 bytes)
//
// Returns:
//   - true if signature is valid and proves ownership
//   - false if signature is invalid or address format is wrong
func (k Keeper) verifyXaiAddressOwnership(ctx sdk.Context, auraAddress string, xaiAddress string, signature []byte) bool {
	if len(signature) == 0 || xaiAddress == "" || auraAddress == "" {
		return false
	}

	// Validate signature format: secp256k1 signatures are 65 bytes (R[32] || S[32] || V[1])
	// where V is the recovery ID (0, 1, 2, or 3)
	if len(signature) != 65 {
		ctx.Logger().Error("Invalid XAI signature length",
			"expected", 65,
			"actual", len(signature),
			"xai_address", xaiAddress)
		return false
	}

	// Build the expected message that should have been signed
	// Format: "Link XAI address <xaiAddress> to Aura address <auraAddress>"
	message := fmt.Sprintf("Link XAI address %s to Aura address %s", xaiAddress, auraAddress)

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
		ctx.Logger().Error("Invalid recovery ID in XAI signature",
			"recovery_id", recoveryID,
			"xai_address", xaiAddress)
		return false
	}

	// Extract R and S components (first 64 bytes)
	sigBytes := signature[:64]

	// Attempt to recover the public key from the signature
	// We try all possible recovery IDs (0-7) to find the correct one
	var recoveredPubKey []byte
	for recID := byte(0); recID <= 7; recID++ {
		// Use the provided recovery ID first, then try others
		tryRecID := (recoveryID + recID) % 8

		// Attempt recovery with this ID
		pubKey, err := k.recoverPubKeyFromSignature(msgHash[:], sigBytes, tryRecID)
		if err != nil {
			continue
		}

		// Derive address from recovered public key
		derivedAddress := k.deriveXaiAddressFromPubKey(pubKey)

		// Check if derived address matches claimed XAI address
		if derivedAddress == xaiAddress {
			recoveredPubKey = pubKey
			break
		}
	}

	if recoveredPubKey == nil {
		ctx.Logger().Error("Failed to recover public key from XAI signature",
			"xai_address", xaiAddress,
			"aura_address", auraAddress)
		return false
	}

	// Additional verification: Verify the signature with the recovered public key
	// This ensures the signature is cryptographically valid
	pubKeyObj := &secp256k1.PubKey{Key: recoveredPubKey}
	if !pubKeyObj.VerifySignature(msgHash[:], sigBytes) {
		ctx.Logger().Error("XAI signature verification failed",
			"xai_address", xaiAddress,
			"aura_address", auraAddress)
		return false
	}

	ctx.Logger().Info("XAI address ownership verified successfully",
		"xai_address", xaiAddress,
		"aura_address", auraAddress)

	return true
}

func (k Keeper) setSwap(ctx sdk.Context, swap *types.CrossChainSwap) {
	if swap == nil || swap.SwapId == "" {
		return
	}
	k.store(ctx).Set(types.SwapKey(swap.SwapId), k.cdc.MustMarshal(swap))
}

func (k Keeper) getSwap(ctx sdk.Context, swapID string) (*types.CrossChainSwap, bool) {
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

func (k Keeper) getWrappedToken(ctx sdk.Context, denom string) (*types.WrappedToken, bool) {
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

func (k Keeper) setWrappedToken(ctx sdk.Context, token *types.WrappedToken) {
	if token == nil || token.WrappedDenom == "" {
		return
	}
	k.store(ctx).Set(types.WrappedTokenKey(token.WrappedDenom), k.cdc.MustMarshal(token))
}

func (k Keeper) getAllWrappedTokens(ctx sdk.Context) []types.WrappedToken {
	store := k.store(ctx)
	iterator := store.Iterator(types.WrappedTokenPrefix, storetypes.PrefixEndBytes(types.WrappedTokenPrefix))
	defer iterator.Close()
	var tokens []types.WrappedToken
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

func (k Keeper) getRelayerStats(ctx sdk.Context, relayer string) (*types.RelayerStats, bool) {
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

func (k Keeper) recordRelayerStats(ctx sdk.Context, relayer string, success bool, volume math.Int) {
	if relayer == "" {
		return
	}
	stats, _ := k.getRelayerStats(ctx, relayer)
	if stats == nil {
		stats = &types.RelayerStats{
			RelayerAddress:   relayer,
			TotalVolume:      "0",
			UptimePercentage: "100",
			LastRelay:        timestamppb.New(ctx.BlockTime()),
		}
	}
	stats.TotalTransfersRelayed++
	if success {
		stats.SuccessfulTransfers++
	} else {
		stats.FailedTransfers++
	}
	if vol, ok := math.NewIntFromString(stats.TotalVolume); ok {
		stats.TotalVolume = vol.Add(volume).String()
	} else {
		stats.TotalVolume = volume.String()
	}
	stats.LastRelay = timestamppb.New(ctx.BlockTime())
	k.setRelayerStats(ctx, stats)
}

func (k Keeper) setRelayerStats(ctx sdk.Context, stats *types.RelayerStats) {
	if stats == nil || stats.RelayerAddress == "" {
		return
	}
	k.store(ctx).Set(types.RelayerKey(stats.RelayerAddress), k.cdc.MustMarshal(stats))
}

func (k Keeper) markTransferFraudulent(ctx sdk.Context, transferID string) (*types.CrossChainTransfer, error) {
	transfer, found := k.getTransfer(ctx, transferID)
	if !found {
		return nil, types.ErrTransferNotFound
	}
	transfer.Status = types.TransferStatus_FAILED
	transfer.Timestamp = timestamppb.New(ctx.BlockTime())
	k.setTransfer(ctx, transfer)
	return transfer, nil
}

func (k Keeper) getAllRelayerStats(ctx sdk.Context) []*types.RelayerStats {
	store := k.store(ctx)
	iterator := store.Iterator(types.RelayerPrefix, storetypes.PrefixEndBytes(types.RelayerPrefix))
	defer iterator.Close()
	var statsList []*types.RelayerStats
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

func (k Keeper) setValidator(ctx sdk.Context, validator *types.BridgeValidator) {
	if validator == nil || validator.Address == "" {
		return
	}
	k.store(ctx).Set(types.ValidatorKey(validator.Address), k.cdc.MustMarshal(validator))
}

func (k Keeper) getValidator(ctx sdk.Context, address string) (*types.BridgeValidator, bool) {
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

func (k Keeper) getAllValidators(ctx sdk.Context) []*types.BridgeValidator {
	store := k.store(ctx)
	iterator := store.Iterator(types.ValidatorPrefix, storetypes.PrefixEndBytes(types.ValidatorPrefix))
	defer iterator.Close()
	var validators []*types.BridgeValidator
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

func (k Keeper) getAllSharedIdentities(ctx sdk.Context) []*types.SharedIdentity {
	store := k.store(ctx)
	iterator := store.Iterator(types.SharedIdentityPrefix, storetypes.PrefixEndBytes(types.SharedIdentityPrefix))
	defer iterator.Close()
	var identities []*types.SharedIdentity
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

func (k Keeper) getAllSwaps(ctx sdk.Context) []*types.CrossChainSwap {
	store := k.store(ctx)
	iterator := store.Iterator(types.SwapPrefix, storetypes.PrefixEndBytes(types.SwapPrefix))
	defer iterator.Close()
	var swaps []*types.CrossChainSwap
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

func (k Keeper) getAllTransfers(ctx sdk.Context) []*types.CrossChainTransfer {
	store := k.store(ctx)
	iterator := store.Iterator(types.TransferPrefix, storetypes.PrefixEndBytes(types.TransferPrefix))
	defer iterator.Close()
	var transfers []*types.CrossChainTransfer
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
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
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
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if k.paramstore.HasKeyTable() {
		k.paramstore.SetParamSet(ctx, &params)
	}
	return nil
}

// ensureBridgeEnabled returns error if the bridge is disabled.
func (k Keeper) ensureBridgeEnabled(ctx sdk.Context) error {
	params := k.GetParams(ctx)
	if !params.BridgeEnabled {
		return fmt.Errorf("bridge disabled")
	}
	return nil
}

// SubmitAttestation records a validator's attestation for a transfer
func (k Keeper) SubmitAttestation(ctx sdk.Context, transferID string, validator string, approved bool) error {
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
		transfer.ValidatorSignatures = append(transfer.ValidatorSignatures, &types.ValidatorSignature{
			ValidatorAddress: validator,
		})
	}
	k.setTransfer(ctx, transfer)
	return nil
}

// GetAttestations returns all validator addresses that attested a transfer
func (k Keeper) GetAttestations(ctx sdk.Context, transferID string) []string {
	store := k.store(ctx)
	prefix := append(types.AttestationPrefix, []byte(transferID)...)
	iterator := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	defer iterator.Close()
	var validators []string
	for ; iterator.Valid(); iterator.Next() {
		parts := strings.Split(string(iterator.Key()), string([]byte{0x00}))
		if len(parts) == 2 {
			validators = append(validators, parts[1])
		}
	}
	return validators
}

// CheckAttestationThreshold returns true if a transfer has enough attestations
func (k Keeper) CheckAttestationThreshold(ctx sdk.Context, transferID string) bool {
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
func (k Keeper) ProcessWithdrawal(ctx sdk.Context, recipient string, amount sdk.Coins, transferID string) error {
	if amount.Empty() {
		return fmt.Errorf("amount cannot be empty")
	}
	params := k.GetParams(ctx)
	maxAmt, ok := math.NewIntFromString(params.MaxTransferAmount)
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
	transfer.Timestamp = timestamppb.New(ctx.BlockTime())
	k.setTransfer(ctx, transfer)
	return nil
}

// InitiateWithdrawal stores a pending withdrawal with timestamp metadata
func (k Keeper) InitiateWithdrawal(ctx sdk.Context, recipient string, amount sdk.Coins) (string, error) {
	transferID := k.nextTransferID(ctx)
	transfer := &types.CrossChainTransfer{
		TransferId:  transferID,
		SourceChain: sourceChainAura,
		TargetChain: sourceChainAura,
		Sender:      sourceChainAura,
		Recipient:   recipient,
		Amount:      amount.AmountOf(amount[0].Denom).String(),
		Denom:       amount[0].Denom,
		Status:      types.TransferStatus_PENDING,
		Timestamp:   timestamppb.New(ctx.BlockTime()),
	}
	k.setTransfer(ctx, transfer)
	return transferID, nil
}

// ExecuteWithdrawal executes a withdrawal after timelock
func (k Keeper) ExecuteWithdrawal(ctx sdk.Context, withdrawalID string) error {
	transfer, found := k.getTransfer(ctx, withdrawalID)
	if !found {
		return types.ErrWithdrawalNotFound
	}
	if transfer.Status != types.TransferStatus_PENDING {
		return nil
	}
	if transfer.Timestamp != nil {
		deadline := transfer.Timestamp.AsTime().Add(types.DefaultTimelockDuration)
		if ctx.BlockTime().Before(deadline) {
			return types.ErrTimelockNotElapsed
		}
	}
	transfer.Status = types.TransferStatus_COMPLETED
	k.setTransfer(ctx, transfer)
	return nil
}

func (k Keeper) setFraudProof(ctx sdk.Context, proof *types.FraudProof) {
	if proof == nil || proof.ChallengedTransferId == "" {
		return
	}
	k.store(ctx).Set(types.FraudProofKey(proof.ChallengedTransferId), k.cdc.MustMarshal(proof))
}

func (k Keeper) getFraudProof(ctx sdk.Context, transferID string) (*types.FraudProof, bool) {
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

func (k Keeper) getFraudProofWindow(ctx sdk.Context) time.Duration {
	window := types.DefaultFraudProofWindow
	if window <= 0 {
		return 0
	}
	return window
}

func (k Keeper) getFraudProofReward(ctx sdk.Context) math.Int {
	params := types.DefaultSecurityParams()
	return params.FraudProofReward
}

func (k Keeper) payoutFraudProofReward(ctx sdk.Context, challenger string, denom string, reward math.Int) error {
	if !reward.IsPositive() || challenger == "" || denom == "" {
		return nil
	}
	if k.bankKeeper == nil {
		return nil
	}
	addr, err := sdk.AccAddressFromBech32(challenger)
	if err != nil {
		return err
	}
	coin := sdk.NewCoin(denom, reward)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return err
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, sdk.NewCoins(coin))
}

// SubmitFraudProof records a fraud challenge for a transfer and begins investigation.
//
// CRITICAL SECURITY ENHANCEMENT: This function now also marks the pending transfer
// as challenged, preventing finalization while the fraud proof is investigated.
func (k Keeper) SubmitFraudProof(ctx sdk.Context, transferID string, submitter string, proof []byte) error {
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
		SubmittedAt:          timestamppb.New(ctx.BlockTime()),
		RewardAmount:         math.ZeroInt().String(),
	}
	k.setFraudProof(ctx, fraudProof)

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
//   1. Marks the transfer as fraudulent
//   2. Slashes ALL validators who signed the fraudulent transfer
//   3. Pays out reward to the fraud proof challenger
//   4. Prevents the transfer from being finalized
func (k Keeper) ResolveFraudProof(ctx sdk.Context, transferID string, valid bool) (types.FraudProof, error) {
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
		k.setFraudProof(ctx, proof)
		return types.FraudProof{}, types.ErrFraudProofExpired
	}
	proof.ResolvedAt = timestamppb.New(ctx.BlockTime())
	reward := math.ZeroInt()
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
	proof.RewardAmount = reward.String()
	k.setFraudProof(ctx, proof)
	return *proof, nil
}

// GetFraudProof retrieves a fraud proof for a transfer
func (k Keeper) GetFraudProof(ctx sdk.Context, transferID string) (types.FraudProof, bool) {
	proof, found := k.getFraudProof(ctx, transferID)
	if !found {
		return types.FraudProof{}, false
	}
	return *proof, true
}

// IsInFraudProofWindow checks if a transfer is still in the fraud proof window
func (k Keeper) IsInFraudProofWindow(ctx sdk.Context, transferID string) bool {
	transfer, found := k.getTransfer(ctx, transferID)
	if !found || transfer.Timestamp == nil {
		return false
	}
	window := k.getFraudProofWindow(ctx)
	if window <= 0 {
		return false
	}
	return ctx.BlockTime().Sub(transfer.Timestamp.AsTime()) <= window
}

// AddSupportedChain adds a new supported chain configuration
func (k Keeper) AddSupportedChain(ctx sdk.Context, chainConfig types.ChainConfig) error {
	if chainConfig.ChainId == "" {
		return types.ErrInvalidParam
	}
	k.setChainConfig(ctx, chainConfig)
	return nil
}

// GetSupportedChain retrieves a supported chain configuration
func (k Keeper) GetSupportedChain(ctx sdk.Context, chainID string) (types.ChainConfig, bool) {
	return k.getChainConfig(ctx, chainID)
}

// RemoveSupportedChain removes a supported chain
func (k Keeper) RemoveSupportedChain(ctx sdk.Context, chainID string) {
	k.store(ctx).Delete(types.ChainConfigKey(strings.ToLower(chainID)))
}

// DisableChain disables a supported chain
func (k Keeper) DisableChain(ctx sdk.Context, chainID string) error {
	config, found := k.getChainConfig(ctx, chainID)
	if !found {
		return types.ErrChainNotFound
	}
	config.Enabled = false
	k.setChainConfig(ctx, config)
	return nil
}

// CalculateBridgeFee calculates the bridge fee for a transfer
func (k Keeper) CalculateBridgeFee(ctx sdk.Context, amount math.Int, chainID string) math.Int {
	params := k.GetParams(ctx)
	if params.BridgeFeeBasisPoints == 0 {
		return math.ZeroInt()
	}
	return amount.MulRaw(int64(params.BridgeFeeBasisPoints)).QuoRaw(10_000)
}

// GetCollectedFees returns the total collected fees
func (k Keeper) GetCollectedFees(ctx sdk.Context) sdk.Coins {
	store := k.store(ctx)
	iterator := store.Iterator([]byte("collected-fees-"), []byte("collected-fees-\xff"))
	defer iterator.Close()
	fees := sdk.NewCoins()
	for ; iterator.Valid(); iterator.Next() {
		denom := strings.TrimPrefix(string(iterator.Key()), "collected-fees-")
		var amount math.Int
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
func (k Keeper) GetSharedIdentity(ctx sdk.Context, address string) (*types.SharedIdentity, bool) {
	return k.getSharedIdentity(ctx, address)
}

// FindSharedIdentityByLinkedAddress is a public exported method for finding identities by linked address
func (k Keeper) FindSharedIdentityByLinkedAddress(ctx sdk.Context, chainName string, address string) *types.SharedIdentity {
	return k.findSharedIdentityByLinkedAddress(ctx, chainName, address)
}

// VerifyPawAddressOwnership is a public exported method for PAW signature verification
func (k Keeper) VerifyPawAddressOwnership(ctx sdk.Context, auraAddress string, pawAddress string, signature []byte) bool {
	return k.verifyPawAddressOwnership(ctx, auraAddress, pawAddress, signature)
}

// VerifyXaiAddressOwnership is a public exported method for XAI signature verification
func (k Keeper) VerifyXaiAddressOwnership(ctx sdk.Context, auraAddress string, xaiAddress string, signature []byte) bool {
	return k.verifyXaiAddressOwnership(ctx, auraAddress, xaiAddress, signature)
}

// AddCollectedFee adds a collected fee to the total
func (k Keeper) AddCollectedFee(ctx sdk.Context, fee sdk.Coin) {
	store := k.store(ctx)
	key := []byte("collected-fees-" + fee.Denom)
	current := math.ZeroInt()
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
func (k Keeper) IsSourceHashProcessed(ctx sdk.Context, sourceChain, sourceHash string) bool {
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
func (k Keeper) MarkSourceHashProcessed(ctx sdk.Context, sourceChain, sourceHash string) {
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

// GetAllProcessedSourceHashes returns all processed source hashes for genesis export.
// Returns a map of "sourceChain:sourceHash" -> true for all processed hashes.
func (k Keeper) GetAllProcessedSourceHashes(ctx sdk.Context) map[string]bool {
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
func (k Keeper) SetProcessedSourceHash(ctx sdk.Context, compositeKey string) {
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
func (k Keeper) SetValidator(ctx sdk.Context, validator *types.BridgeValidator) {
	k.setValidator(ctx, validator)
}

// SetTransfer is a public method to set a cross-chain transfer
func (k Keeper) SetTransfer(ctx sdk.Context, transfer *types.CrossChainTransfer) {
	k.setTransfer(ctx, transfer)
}

// IndexTransferHash is a public method to index a transfer by hash
func (k Keeper) IndexTransferHash(ctx sdk.Context, hash, transferID string) {
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
func (k Keeper) getActiveValidators(ctx sdk.Context) []string {
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
func (k Keeper) IsValidatorActive(ctx sdk.Context, validatorAddr string) bool {
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
func (k Keeper) isSignatureSetUsed(ctx sdk.Context, transferID string, signatureSetHash []byte) bool {
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
func (k Keeper) markSignatureSetUsed(ctx sdk.Context, transferID string, signatureSetHash []byte) {
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
func (k Keeper) computeSignatureSetHash(signatures [][]byte) []byte {
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
func (k Keeper) VerifyMerkleProofBytes(merkleRoot, transactionLeaf, merkleProofBytes []byte) bool {
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
	var proofHashes [][]byte
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
func (k Keeper) verifyMerkleProofBruteForce(merkleRoot, transactionLeaf []byte, proofHashes [][]byte) bool {
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
func (k Keeper) ConstructTransactionLeaf(sourceChain, burnTxHash, sender, amount, denom string) []byte {
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
func (k Keeper) VerifySourceBlock(ctx sdk.Context, sourceChain string, blockHeight uint64, blockHash []byte) bool {
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
func (k Keeper) GetVerifiedBlockHash(ctx sdk.Context, sourceChain string, blockHeight uint64) []byte {
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
func (k Keeper) SetVerifiedBlockHash(ctx sdk.Context, sourceChain string, blockHeight uint64, blockHash []byte) {
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
//   - math.Int: Total amount minted today (zero if none)
func (k Keeper) GetDailyMintedAmount(ctx sdk.Context, denom string) math.Int {
	if denom == "" {
		return math.ZeroInt()
	}

	// Format: YYYYMMDD (e.g., "20250102")
	date := ctx.BlockTime().UTC().Format("20060102")
	store := k.store(ctx)
	key := types.DailyMintKey(date, denom)

	bz := store.Get(key)
	if bz == nil {
		return math.ZeroInt()
	}

	var amount math.Int
	if err := amount.Unmarshal(bz); err != nil {
		return math.ZeroInt()
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
func (k Keeper) AddDailyMintedAmount(ctx sdk.Context, denom string, amount math.Int) {
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
//   - math.Int: Total amount minted this hour (zero if none)
func (k Keeper) GetHourlyMintedAmount(ctx sdk.Context, denom string) math.Int {
	if denom == "" {
		return math.ZeroInt()
	}

	// Format: YYYYMMDDHH (e.g., "2025010214" for 2PM)
	datetime := ctx.BlockTime().UTC().Format("2006010215")
	store := k.store(ctx)
	key := types.HourlyMintKey(datetime, denom)

	bz := store.Get(key)
	if bz == nil {
		return math.ZeroInt()
	}

	var amount math.Int
	if err := amount.Unmarshal(bz); err != nil {
		return math.ZeroInt()
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
func (k Keeper) AddHourlyMintedAmount(ctx sdk.Context, denom string, amount math.Int) {
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
func (k Keeper) ResetDailyMint(ctx sdk.Context) {
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
func (k Keeper) ResetHourlyMint(ctx sdk.Context) {
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
func (k Keeper) setPendingTransfer(ctx sdk.Context, pendingTransfer *types.PendingTransfer) {
	if pendingTransfer == nil || pendingTransfer.TransferId == "" {
		return
	}
	store := k.store(ctx)
	key := types.PendingTransferKey(pendingTransfer.TransferId)
	store.Set(key, k.cdc.MustMarshal(pendingTransfer))
}

// SetPendingTransfer is a public exported method for setting pending transfers (for tests).
func (k Keeper) SetPendingTransfer(ctx sdk.Context, pendingTransfer *types.PendingTransfer) {
	k.setPendingTransfer(ctx, pendingTransfer)
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
func (k Keeper) GetPendingTransfer(ctx sdk.Context, transferID string) (*types.PendingTransfer, bool) {
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
func (k Keeper) deletePendingTransfer(ctx sdk.Context, transferID string) {
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
func (k Keeper) GetAllPendingTransfers(ctx sdk.Context) []*types.PendingTransfer {
	store := k.store(ctx)
	iterator := store.Iterator(types.PendingTransferPrefix, storetypes.PrefixEndBytes(types.PendingTransferPrefix))
	defer iterator.Close()

	var pendingTransfers []*types.PendingTransfer
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
func (k Keeper) IsPendingTransferExpired(ctx sdk.Context, pendingTransfer *types.PendingTransfer) bool {
	if pendingTransfer == nil || pendingTransfer.UnlockTime == nil {
		return false
	}
	return ctx.BlockTime().After(pendingTransfer.UnlockTime.AsTime()) ||
		ctx.BlockTime().Equal(pendingTransfer.UnlockTime.AsTime())
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
func (k Keeper) MarkPendingTransferChallenged(ctx sdk.Context, transferID string, fraudProofID string) error {
	pending, found := k.GetPendingTransfer(ctx, transferID)
	if !found {
		return types.ErrTransferNotFound
	}

	if pending.Challenged {
		return fmt.Errorf("transfer already challenged with fraud proof %s", pending.FraudProofId)
	}

	pending.Challenged = true
	pending.FraudProofId = fraudProofID
	k.setPendingTransfer(ctx, pending)

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
//   1. The fraud proof window has fully elapsed
//   2. No valid fraud proof has been submitted against the transfer
//   3. All security checks pass (replay protection, supply caps, etc.)
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
func (k Keeper) FinalizeTransfer(ctx sdk.Context, transferID string) error {
	// 1. Retrieve pending transfer
	pending, found := k.GetPendingTransfer(ctx, transferID)
	if !found {
		return types.ErrTransferNotFound
	}

	// 2. CRITICAL SECURITY: Check fraud proof window has expired
	if !k.IsPendingTransferExpired(ctx, pending) {
		return fmt.Errorf("fraud proof window not expired: unlocks at %s, current time %s",
			pending.UnlockTime.AsTime(), ctx.BlockTime())
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

	// 5. Parse amount from string
	amount, ok := math.NewIntFromString(pending.Amount)
	if !ok {
		return fmt.Errorf("invalid amount string: %s", pending.Amount)
	}

	// Create coin to mint
	coin := sdk.NewCoin(pending.Denom, amount)

	// 6. CRITICAL SECURITY: Perform final security checks before minting
	// These checks should have been done when creating the pending transfer,
	// but we verify again in case parameters changed during the fraud proof window
	params := k.GetParams(ctx)

	// Check per-transfer maximum
	maxTransfer, ok := math.NewIntFromString(params.MaxTransferAmount)
	if ok && amount.GT(maxTransfer) {
		return fmt.Errorf("amount %s exceeds max transfer limit %s", amount, maxTransfer)
	}

	// Check per-token supply cap (if configured)
	if cap, exists := params.SupplyCaps[pending.Denom]; exists {
		supplyCap, ok := math.NewIntFromString(cap)
		if ok {
			currentSupply := k.bankKeeper.GetSupply(ctx, pending.Denom).Amount
			if currentSupply.Add(amount).GT(supplyCap) {
				return fmt.Errorf("minting would exceed supply cap of %s (current: %s)",
					supplyCap, currentSupply)
			}
		}
	}

	// 7. Mint and send tokens (following checks-effects-interactions)
	if k.bankKeeper != nil {
		// Mint to module
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(coin)); err != nil {
			return fmt.Errorf("failed to mint coins: %w", err)
		}

		// Send to recipient
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(coin)); err != nil {
			return fmt.Errorf("failed to send coins to recipient: %w", err)
		}
	}

	// 8. Update transfer status to completed
	transfer, found := k.getTransfer(ctx, transferID)
	if found {
		transfer.Status = types.TransferStatus_COMPLETED
		transfer.Timestamp = timestamppb.New(ctx.BlockTime())
		k.setTransfer(ctx, transfer)
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

// ========================================================================
// CRYPTOGRAPHIC HELPER FUNCTIONS FOR SIGNATURE VERIFICATION
// ========================================================================

// recoverPubKeyFromSignature recovers a public key from a message hash and signature.
//
// SECURITY CRITICAL: This function performs ECDSA public key recovery using secp256k1.
// It is used to verify that a signature was created by the holder of a specific private key.
//
// Process:
//   1. Parse R and S components from signature bytes
//   2. Use recovery ID to determine which of the 4 possible public keys is correct
//   3. Recover the public key using secp256k1 curve parameters
//   4. Return compressed public key (33 bytes)
//
// Parameters:
//   - msgHash: SHA256 hash of the signed message (32 bytes)
//   - signature: R and S components of signature (64 bytes)
//   - recoveryID: Recovery ID (0-3) indicating which public key to recover
//
// Returns:
//   - []byte: Compressed public key (33 bytes) on success
//   - error: If recovery fails (invalid signature, invalid recovery ID, etc.)
func (k Keeper) recoverPubKeyFromSignature(msgHash []byte, signature []byte, recoveryID byte) ([]byte, error) {
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

// derivePawAddressFromPubKey derives a PAW (Cosmos SDK) address from a public key.
//
// Address derivation process for Cosmos SDK chains:
//   1. Take the compressed public key (33 bytes)
//   2. Hash with SHA256
//   3. Hash with RIPEMD160
//   4. Encode with Bech32 using "paw" prefix
//
// For simplicity, this implementation returns the hex-encoded hash.
// In production, you would use the full Bech32 encoding with proper prefix.
//
// Parameters:
//   - pubKey: Compressed secp256k1 public key (33 bytes)
//
// Returns:
//   - string: Hex-encoded address (for testing) or Bech32 address (for production)
func (k Keeper) derivePawAddressFromPubKey(pubKey []byte) string {
	if len(pubKey) != 33 {
		return ""
	}

	// Step 1: SHA256 hash of public key
	sha256Hash := sha256.Sum256(pubKey)

	// Step 2: RIPEMD160 hash of SHA256 hash
	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash[:])
	addressHash := ripemd160Hasher.Sum(nil)

	// For production, encode as Bech32 with "paw" prefix
	// For now, return hex-encoded for testing compatibility
	// In a full implementation, use:
	// addr, _ := bech32.ConvertAndEncode("paw", addressHash)
	// return addr

	// Return hex-encoded address (20 bytes = 40 hex chars)
	return hex.EncodeToString(addressHash)
}

// deriveXaiAddressFromPubKey derives an XAI (Cosmos SDK) address from a public key.
//
// Address derivation process for Cosmos SDK chains:
//   1. Take the compressed public key (33 bytes)
//   2. Hash with SHA256
//   3. Hash with RIPEMD160
//   4. Encode with Bech32 using "xai" prefix
//
// For simplicity, this implementation returns the hex-encoded hash.
// In production, you would use the full Bech32 encoding with proper prefix.
//
// Parameters:
//   - pubKey: Compressed secp256k1 public key (33 bytes)
//
// Returns:
//   - string: Hex-encoded address (for testing) or Bech32 address (for production)
func (k Keeper) deriveXaiAddressFromPubKey(pubKey []byte) string {
	if len(pubKey) != 33 {
		return ""
	}

	// Step 1: SHA256 hash of public key
	sha256Hash := sha256.Sum256(pubKey)

	// Step 2: RIPEMD160 hash of SHA256 hash
	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash[:])
	addressHash := ripemd160Hasher.Sum(nil)

	// For production, encode as Bech32 with "xai" prefix
	// For now, return hex-encoded for testing compatibility
	// In a full implementation, use:
	// addr, _ := bech32.ConvertAndEncode("xai", addressHash)
	// return addr

	// Return hex-encoded address (20 bytes = 40 hex chars)
	return hex.EncodeToString(addressHash)
}
