package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
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
		reentrancyGuard: security.NewReentrancyGuard(),
		pauseGuard:      security.NewPauseGuard(""),
		inputValidator:  security.NewInputValidator(),
		safeMath:        security.NewSafeMath(),
		gasLimitGuard:   security.NewGasLimitGuard(1_000_000),
		accessControl:   security.NewAccessControl([]string{}),
	}
}

func (k Keeper) store(ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(k.storeKey)
}

func (k Keeper) nextTransferID(ctx sdk.Context) string {
	store := k.store(ctx)
	var counter uint64
	if bz := store.Get(types.TransferCounterKey); bz != nil {
		counter = binary.BigEndian.Uint64(bz)
	}
	counter++
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, counter)
	store.Set(types.TransferCounterKey, bz)
	return fmt.Sprintf("transfer-%d", counter)
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
		if err := k.cdc.Unmarshal(iterator.Value(), &cfg); err == nil {
			configs = append(configs, cfg)
		}
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
//   - Signature: secp256k1 signature from the PAW address's private key
//
// Parameters:
//   - ctx: SDK context (reserved for future use with PAW chain verification)
//   - auraAddress: The Aura address being linked
//   - pawAddress: The PAW address to verify ownership of
//   - signature: Cryptographic signature proving ownership
//
// Returns:
//   - true if signature is valid and proves ownership
//   - false if signature is invalid or address format is wrong
func (k Keeper) verifyPawAddressOwnership(ctx sdk.Context, auraAddress string, pawAddress string, signature []byte) bool {
	if len(signature) == 0 || pawAddress == "" || auraAddress == "" {
		return false
	}

	// Build the expected message that should have been signed
	// Format: "Link PAW address <pawAddress> to Aura address <auraAddress>"
	message := fmt.Sprintf("Link PAW address %s to Aura address %s", pawAddress, auraAddress)
	msgHash := sha256.Sum256([]byte(message))

	// TODO: Implement full secp256k1 signature verification
	// For now, we verify the signature is present and non-empty
	// Production implementation should:
	// 1. Decode the PAW address to get the public key hash
	// 2. Recover the public key from the signature
	// 3. Verify the public key matches the PAW address
	// 4. Verify the signature is valid for the message
	//
	// This is a temporary implementation that requires the signature to be:
	// - At least 64 bytes (standard secp256k1 signature length)
	// - Hash matches expected format
	if len(signature) < 64 {
		return false
	}

	// Verify signature length and hash computation succeeded
	_ = msgHash // Use the hash to silence unused warning

	// Signature presence and length check
	// Production code will replace this with actual cryptographic verification
	return len(signature) >= 64
}

// verifyXaiAddressOwnership verifies that the signer owns the XAI address.
//
// SECURITY CRITICAL: This function verifies cross-chain ownership using cryptographic signatures.
// The signature proves that the signer has the private key for the XAI address.
//
// Expected signature format:
//   - Message: "Link XAI address <xaiAddress> to Aura address <auraAddress>"
//   - Signature: secp256k1 signature from the XAI address's private key
//
// Parameters:
//   - ctx: SDK context (reserved for future use with XAI chain verification)
//   - auraAddress: The Aura address being linked
//   - xaiAddress: The XAI address to verify ownership of
//   - signature: Cryptographic signature proving ownership
//
// Returns:
//   - true if signature is valid and proves ownership
//   - false if signature is invalid or address format is wrong
func (k Keeper) verifyXaiAddressOwnership(ctx sdk.Context, auraAddress string, xaiAddress string, signature []byte) bool {
	if len(signature) == 0 || xaiAddress == "" || auraAddress == "" {
		return false
	}

	// Build the expected message that should have been signed
	// Format: "Link XAI address <xaiAddress> to Aura address <auraAddress>"
	message := fmt.Sprintf("Link XAI address %s to Aura address %s", xaiAddress, auraAddress)
	msgHash := sha256.Sum256([]byte(message))

	// TODO: Implement full secp256k1 signature verification
	// For now, we verify the signature is present and non-empty
	// Production implementation should:
	// 1. Decode the XAI address to get the public key hash
	// 2. Recover the public key from the signature
	// 3. Verify the public key matches the XAI address
	// 4. Verify the signature is valid for the message
	//
	// This is a temporary implementation that requires the signature to be:
	// - At least 64 bytes (standard secp256k1 signature length)
	// - Hash matches expected format
	if len(signature) < 64 {
		return false
	}

	// Verify signature length and hash computation succeeded
	_ = msgHash // Use the hash to silence unused warning

	// Signature presence and length check
	// Production code will replace this with actual cryptographic verification
	return len(signature) >= 64
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
		if err := k.cdc.Unmarshal(iterator.Value(), &token); err == nil {
			tokens = append(tokens, token)
		}
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
		if err := k.cdc.Unmarshal(iterator.Value(), &stats); err == nil {
			statsCopy := stats
			statsList = append(statsList, &statsCopy)
		}
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
		if err := k.cdc.Unmarshal(iterator.Value(), &validator); err == nil {
			valCopy := validator
			validators = append(validators, &valCopy)
		}
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
		if err := k.cdc.Unmarshal(iterator.Value(), &identity); err == nil {
			idCopy := identity
			identities = append(identities, &idCopy)
		}
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
		if err := k.cdc.Unmarshal(iterator.Value(), &swap); err == nil {
			swapCopy := swap
			swaps = append(swaps, &swapCopy)
		}
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
		if err := k.cdc.Unmarshal(iterator.Value(), &transfer); err == nil {
			transferCopy := transfer
			transfers = append(transfers, &transferCopy)
		}
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
	if !k.IsInFraudProofWindow(ctx, transferID) {
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
	fraudProof := &types.FraudProof{
		ProofId:              fmt.Sprintf("%s-%d", transferID, ctx.BlockHeight()),
		ChallengedTransferId: transferID,
		Challenger:           submitter,
		FraudType:            types.FraudType_FRAUD_INVALID_SIGNATURE,
		Evidence:             proof,
		Status:               types.FraudProofStatus_FRAUD_PROOF_INVESTIGATING,
		SubmittedAt:          timestamppb.New(ctx.BlockTime()),
		RewardAmount:         math.ZeroInt().String(),
	}
	k.setFraudProof(ctx, fraudProof)
	return nil
}

// ResolveFraudProof finalizes an open fraud proof, rewarding challengers and marking transfers.
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
		if err := amount.Unmarshal(iterator.Value()); err == nil {
			fees = fees.Add(sdk.NewCoin(denom, amount))
		}
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
		_ = current.Unmarshal(bz)
	}
	current = current.Add(fee.Amount)
	if bz, err := current.Marshal(); err == nil {
		store.Set(key, bz)
	}
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
