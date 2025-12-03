package keeper

import (
	context "context"
	"crypto/sha256"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/bridge/types"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
)

var _ bridgepb.MsgServer = (*msgServer)(nil)

type msgServer struct {
	bridgepb.UnimplementedMsgServer
	Keeper *Keeper
}

// NewMsgServerImpl wires the keeper into the protobuf msg server implementation.
func NewMsgServerImpl(k *Keeper) bridgepb.MsgServer {
	return &msgServer{Keeper: k}
}

func normalizeChain(chain string) string {
	return strings.ToLower(strings.TrimSpace(chain))
}

// verifyRawValidatorSignatures verifies raw byte signatures from multiple validators
// against the ACTIVE validator set. This function implements critical security checks:
//   1. Verifies cryptographic signatures
//   2. Checks validator authorization (active status)
//   3. Prevents signature reuse (each validator counted once)
//   4. Ensures minimum threshold of unique active validators
//
// Security considerations:
//   - Only ACTIVE validators are allowed (inactive/slashed/jailed are rejected)
//   - Each validator can only contribute one signature (prevents duplicate counting)
//   - Cryptographic verification using validator's registered public key
//   - Minimum required threshold enforced (never less than MinAllowedConfirmations)
//
// Parameters:
//   - ctx: SDK context for state access
//   - signatures: Raw signature bytes from validators
//   - msgHash: Hash of the message that was signed
//   - minRequired: Minimum number of valid signatures required
//
// Returns:
//   - validCount: Number of valid signatures from unique active validators
//   - err: Error if threshold not met or validation fails
func (ms msgServer) verifyRawValidatorSignatures(
	ctx sdk.Context,
	signatures [][]byte,
	msgHash []byte,
	minRequired uint64,
) (validCount int, err error) {
	if len(signatures) < int(minRequired) {
		return 0, errorsmod.Wrapf(types.ErrInsufficientSignatures,
			"provided %d signatures, need at least %d", len(signatures), minRequired)
	}

	// SECURITY: Get only ACTIVE validators (governance-approved, not slashed/jailed)
	activeValidators := ms.Keeper.getActiveValidators(ctx)
	if len(activeValidators) == 0 {
		return 0, errorsmod.Wrap(types.ErrNoActiveValidators,
			"no active validators in current set")
	}

	// Get full validator objects for active validators only
	activeValidatorMap := make(map[string]*types.BridgeValidator)
	for _, addr := range activeValidators {
		if validator, found := ms.Keeper.getValidator(ctx, addr); found {
			activeValidatorMap[addr] = validator
		}
	}

	if len(activeValidatorMap) == 0 {
		return 0, errorsmod.Wrap(types.ErrNoActiveValidators,
			"no active validators available")
	}

	// Track which validators have been matched (prevent counting same validator twice)
	usedValidators := make(map[string]bool)

	for _, sigBytes := range signatures {
		if len(sigBytes) == 0 {
			continue
		}

		// Try to match this signature against ACTIVE validators only
		for addr, validator := range activeValidatorMap {
			// Skip if validator already matched (prevent duplicate counting)
			if usedValidators[addr] {
				continue
			}

			// Parse the validator's public key
			if len(validator.PublicKey) == 0 {
				continue
			}

			var pubKey cryptotypes.PubKey
			// Unmarshal the public key
			err := ms.Keeper.cdc.UnmarshalInterface(validator.PublicKey, &pubKey)
			if err != nil {
				// Skip this validator if pubkey can't be parsed
				continue
			}

			if pubKey == nil {
				continue
			}

			// CRITICAL: Verify the cryptographic signature
			if pubKey.VerifySignature(msgHash, sigBytes) {
				// Valid signature from active validator found
				usedValidators[addr] = true
				validCount++
				break // Move to next signature
			}
		}
	}

	// Check if we have enough valid signatures from active validators
	if validCount < int(minRequired) {
		return validCount, errorsmod.Wrapf(types.ErrInsufficientSignatures,
			"only %d valid signatures from active validators (out of %d provided), need %d",
			validCount, len(signatures), minRequired)
	}

	return validCount, nil
}

// LockTokens locks native tokens on Aura for cross-chain transfer.
func (ms msgServer) LockTokens(goCtx context.Context, msg *bridgepb.MsgLockTokens) (*bridgepb.MsgLockTokensResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.Amount == nil || !msg.Amount.Amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.Keeper.ensureBridgeEnabled(ctx); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	chainID := normalizeChain(msg.TargetChain)
	if chainID == "" {
		return nil, status.Error(codes.InvalidArgument, "target chain required")
	}
	chainCfg, found := ms.Keeper.getChainConfig(ctx, chainID)
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrChainNotFound.Error())
	}
	if !chainCfg.Enabled {
		return nil, status.Error(codes.FailedPrecondition, types.ErrChainDisabled.Error())
	}
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	amnt := sdk.NewCoin(msg.Amount.Denom, msg.Amount.Amount)
	params := ms.Keeper.GetParams(ctx)
	maxAmt, ok := sdkmath.NewIntFromString(params.MaxTransferAmount)
	if ok && amnt.Amount.GT(maxAmt) {
		return nil, status.Error(codes.InvalidArgument, types.ErrCircuitBreakerTripped.Error())
	}
	if ms.Keeper.bankKeeper != nil {
		if err := ms.Keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(amnt)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	transferID := ms.Keeper.nextTransferID(ctx)
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:            transferID,
		SourceChain:           sourceChainAura,
		TargetChain:           chainID,
		Sender:                msg.Sender,
		Recipient:             msg.Recipient,
		Amount:                amnt.Amount.String(),
		Denom:                 amnt.Denom,
		Status:                bridgepb.TransferStatus_PENDING,
		Timestamp:             timestamppb.New(ctx.BlockTime()),
		RequiredConfirmations: params.MinConfirmations,
	}
	ms.Keeper.setTransfer(ctx, transfer)
	return &bridgepb.MsgLockTokensResponse{
		TransferId:          transferID,
		EstimatedCompletion: 600,
	}, nil
}

// MintTokens mints wrapped assets on Aura once sufficient attestations exist.
func (ms msgServer) MintTokens(goCtx context.Context, msg *bridgepb.MsgMintTokens) (*bridgepb.MsgMintTokensResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.Validator == "" {
		return nil, status.Error(codes.InvalidArgument, "validator required")
	}
	amount, ok := sdkmath.NewIntFromString(msg.Amount)
	if !ok || !amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	transferID, hasIndex := ms.Keeper.transferIDByHash(ctx, msg.SourceTxHash)
	if !hasIndex {
		transferID = ms.Keeper.nextTransferID(ctx)
	}
	transfer, found := ms.Keeper.getTransfer(ctx, transferID)
	if !found {
		transfer = &bridgepb.CrossChainTransfer{
			TransferId:  transferID,
			SourceChain: normalizeChain(msg.SourceChain),
			TargetChain: sourceChainAura,
			Sender:      msg.SourceChain,
			Recipient:   msg.Recipient,
			Amount:      amount.String(),
			Denom:       msg.Denom,
			Status:      bridgepb.TransferStatus_CONFIRMED,
			Timestamp:   timestamppb.New(ctx.BlockTime()),
		}
		ms.Keeper.setTransfer(ctx, transfer)
		ms.Keeper.indexTransferHash(ctx, msg.SourceTxHash, transferID)
	}
	if err := ms.Keeper.SubmitAttestation(ctx, transferID, msg.Validator, true); err != nil && err != types.ErrDuplicateAttestation {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if ms.Keeper.CheckAttestationThreshold(ctx, transferID) && transfer.Status != bridgepb.TransferStatus_COMPLETED {
		recipient, err := sdk.AccAddressFromBech32(transfer.Recipient)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		coin := sdk.NewCoin(msg.Denom, amount)
		if ms.Keeper.bankKeeper != nil {
			coins := sdk.NewCoins(coin)
			if err := ms.Keeper.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			if err := ms.Keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coins); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
		transfer.Status = bridgepb.TransferStatus_COMPLETED
		transfer.TargetTxHash = msg.SourceTxHash
		transfer.Timestamp = timestamppb.New(ctx.BlockTime())
		ms.Keeper.setTransfer(ctx, transfer)
	}
	wrappedDenom := fmt.Sprintf("%s.%s", normalizeChain(msg.SourceChain), msg.Denom)
	token, _ := ms.Keeper.getWrappedToken(ctx, wrappedDenom)
	if token == nil {
		token = &bridgepb.WrappedToken{
			WrappedDenom:  wrappedDenom,
			OriginalDenom: msg.Denom,
			SourceChain:   normalizeChain(msg.SourceChain),
			TotalSupply:   amount.String(),
		}
	} else {
		if total, ok := sdkmath.NewIntFromString(token.TotalSupply); ok {
			token.TotalSupply = total.Add(amount).String()
		}
	}
	ms.Keeper.setWrappedToken(ctx, token)
	return &bridgepb.MsgMintTokensResponse{Success: true, WrappedDenom: wrappedDenom}, nil
}

// UnlockTokens unlocks locked assets after a burn proof on the destination chain.
func (ms msgServer) UnlockTokens(goCtx context.Context, msg *bridgepb.MsgUnlockTokens) (*bridgepb.MsgUnlockTokensResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	amount, ok := sdkmath.NewIntFromString(msg.Amount)
	if !ok || !amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// CRITICAL SECURITY: Check if this source hash was already processed (replay attack prevention)
	// This MUST happen BEFORE any other processing to prevent replay attacks where an attacker
	// reuses the same burn transaction hash to unlock tokens multiple times.
	sourceChain := normalizeChain(msg.SourceChain)
	if sourceChain == "" {
		// If no source chain specified, try to infer from transfer
		transferID, hasIndex := ms.Keeper.transferIDByHash(ctx, msg.BurnTxHash)
		if hasIndex {
			if transfer, found := ms.Keeper.getTransfer(ctx, transferID); found {
				sourceChain = normalizeChain(transfer.SourceChain)
			}
		}
	}

	// Perform replay check with normalized source chain
	if ms.Keeper.IsSourceHashProcessed(ctx, sourceChain, msg.BurnTxHash) {
		return nil, status.Error(codes.AlreadyExists,
			"source transaction already processed (replay attack prevented)")
	}

	transferID, hasIndex := ms.Keeper.transferIDByHash(ctx, msg.BurnTxHash)
	if !hasIndex {
		transferID = msg.BurnTxHash
	}
	transfer, found := ms.Keeper.getTransfer(ctx, transferID)
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrTransferNotFound.Error())
	}

	// SECURITY FIX: Verify cryptographic signatures instead of just counting them
	required := ms.Keeper.GetParams(ctx).MinConfirmations

	// CRITICAL SECURITY: Enforce minimum even if params were misconfigured
	// Never allow less than 2 validators to prevent single validator control
	if required < types.MinAllowedConfirmations {
		required = types.MinAllowedConfirmations
	}

	// Build deterministic message hash for signature verification
	// Format: sourceChain:burnTxHash:recipient:amount:denom
	msgToSign := fmt.Sprintf("%s:%s:%s:%s:%s",
		transfer.SourceChain,
		msg.BurnTxHash,
		msg.Sender,
		msg.Amount,
		msg.Denom,
	)
	msgHash := sha256.Sum256([]byte(msgToSign))

	// CRITICAL SECURITY: Check if this signature set has been used before
	// This prevents replay attacks where the same valid signatures are reused
	signatureSetHash := ms.Keeper.computeSignatureSetHash(msg.ValidatorSignatures)
	if signatureSetHash != nil && ms.Keeper.isSignatureSetUsed(ctx, transferID, signatureSetHash) {
		return nil, status.Error(codes.AlreadyExists,
			"signature set already used for this transfer (replay attack prevented)")
	}

	// Verify signatures cryptographically against ACTIVE validator set
	// This function:
	//   - Verifies cryptographic signatures
	//   - Checks validator authorization (only active validators)
	//   - Counts UNIQUE validators (prevents duplicate counting)
	//   - Enforces minimum threshold
	validCount, err := ms.verifyRawValidatorSignatures(ctx, msg.ValidatorSignatures, msgHash[:], required)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	// Log successful verification for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"unlock_tokens_verified",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("burn_tx_hash", msg.BurnTxHash),
			sdk.NewAttribute("valid_signatures", fmt.Sprintf("%d", validCount)),
			sdk.NewAttribute("required_signatures", fmt.Sprintf("%d", required)),
			sdk.NewAttribute("signature_set_hash", fmt.Sprintf("%x", signatureSetHash)),
		),
	)

	// CRITICAL SECURITY: Mark signature set as used BEFORE unlocking tokens
	// This prevents the same signatures from being replayed
	// Following checks-effects-interactions pattern
	if signatureSetHash != nil {
		ms.Keeper.markSignatureSetUsed(ctx, transferID, signatureSetHash)
	}

	// CRITICAL SECURITY: Mark the source hash as processed BEFORE unlocking tokens
	// This prevents reentrancy and ensures the replay protection is atomic.
	// Following checks-effects-interactions pattern: effects (state change) before interactions (token transfer).
	ms.Keeper.MarkSourceHashProcessed(ctx, sourceChain, msg.BurnTxHash)

	recipient, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	coin := sdk.NewCoin(msg.Denom, amount)
	if ms.Keeper.bankKeeper != nil {
		if err := ms.Keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(coin)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	transfer.Status = bridgepb.TransferStatus_COMPLETED
	transfer.TargetTxHash = msg.BurnTxHash
	transfer.Timestamp = timestamppb.New(ctx.BlockTime())
	ms.Keeper.setTransfer(ctx, transfer)
	return &bridgepb.MsgUnlockTokensResponse{Success: true}, nil
}

// BurnTokens burns wrapped tokens on Aura to unlock on the origin chain.
func (ms msgServer) BurnTokens(goCtx context.Context, msg *bridgepb.MsgBurnTokens) (*bridgepb.MsgBurnTokensResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	chainID := normalizeChain(msg.TargetChain)
	if chainID == "" {
		return nil, status.Error(codes.InvalidArgument, "target chain required")
	}
	if cfg, found := ms.Keeper.getChainConfig(ctx, chainID); !found {
		return nil, status.Error(codes.NotFound, types.ErrChainNotFound.Error())
	} else if !cfg.Enabled {
		return nil, status.Error(codes.FailedPrecondition, types.ErrChainDisabled.Error())
	}
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	amnt := msg.Amount
	if ms.Keeper.bankKeeper != nil {
		if err := ms.Keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(*amnt)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if err := ms.Keeper.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(*amnt)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	transferID := ms.Keeper.nextTransferID(ctx)
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  transferID,
		SourceChain: sourceChainAura,
		TargetChain: chainID,
		Sender:      msg.Sender,
		Recipient:   msg.Recipient,
		Amount:      amnt.Amount.String(),
		Denom:       amnt.Denom,
		Status:      bridgepb.TransferStatus_CONFIRMED,
		Timestamp:   timestamppb.New(ctx.BlockTime()),
	}
	ms.Keeper.setTransfer(ctx, transfer)
	ms.Keeper.indexTransferHash(ctx, transferID, transferID)
	return &bridgepb.MsgBurnTokensResponse{TransferId: transferID, EstimatedCompletion: 600}, nil
}

// LinkAddress links Aura/PAW/XAI addresses for shared identity.
//
// SECURITY CRITICAL: This function implements cross-chain identity linking with strict access controls:
//   1. Signer verification: Only the Aura address owner can link addresses
//   2. Cross-chain ownership proof: Cryptographic signatures required from PAW/XAI addresses
//   3. Conflict prevention: Prevents overwriting existing links without proper authorization
//
// Attack vectors prevented:
//   - Identity theft: Can't link someone else's addresses without their private keys
//   - Cross-chain impersonation: Requires proof of ownership on each chain
//   - Link hijacking: Existing links can't be overwritten by unauthorized parties
//
// Parameters:
//   - msg: Contains addresses to link and cryptographic proofs of ownership
//
// Returns:
//   - Response with success status and linked identity ID
//   - Error if: signer not authorized, signatures invalid, addresses already linked
func (ms msgServer) LinkAddress(goCtx context.Context, msg *bridgepb.MsgLinkAddress) (*bridgepb.MsgLinkAddressResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.AuraAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "aura address required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// CRITICAL SECURITY: Verify signer owns the Aura address
	// This prevents arbitrary address linking by unauthorized parties
	if msg.Signer == "" {
		return nil, status.Error(codes.Unauthenticated, "signer required")
	}

	// Verify signer is the Aura address owner
	// The signer must match the Aura address being linked
	if msg.AuraAddress != msg.Signer {
		return nil, status.Error(codes.PermissionDenied,
			"signer must be the Aura address owner")
	}

	// CRITICAL SECURITY: Check if addresses are already linked to someone else
	// Prevent hijacking existing identity links
	if msg.PawAddress != "" {
		existingIdentity := ms.Keeper.findSharedIdentityByLinkedAddress(ctx, "paw", msg.PawAddress)
		if existingIdentity != nil && existingIdentity.Address != msg.Signer {
			return nil, status.Error(codes.AlreadyExists,
				fmt.Sprintf("PAW address %s already linked to identity %s",
					msg.PawAddress, existingIdentity.Address))
		}
	}

	if msg.XaiAddress != "" {
		existingIdentity := ms.Keeper.findSharedIdentityByLinkedAddress(ctx, "xai", msg.XaiAddress)
		if existingIdentity != nil && existingIdentity.Address != msg.Signer {
			return nil, status.Error(codes.AlreadyExists,
				fmt.Sprintf("XAI address %s already linked to identity %s",
					msg.XaiAddress, existingIdentity.Address))
		}
	}

	// CRITICAL SECURITY: Verify cross-chain ownership proofs
	// Require cryptographic signatures from PAW/XAI addresses to prove ownership
	if msg.PawAddress != "" {
		if len(msg.PawSignature) == 0 {
			return nil, status.Error(codes.InvalidArgument,
				"PAW signature required when linking PAW address")
		}
		if !ms.Keeper.verifyPawAddressOwnership(ctx, msg.AuraAddress, msg.PawAddress, msg.PawSignature) {
			return nil, status.Error(codes.Unauthenticated,
				"invalid PAW address ownership proof")
		}
	}

	if msg.XaiAddress != "" {
		if len(msg.XaiSignature) == 0 {
			return nil, status.Error(codes.InvalidArgument,
				"XAI signature required when linking XAI address")
		}
		if !ms.Keeper.verifyXaiAddressOwnership(ctx, msg.AuraAddress, msg.XaiAddress, msg.XaiSignature) {
			return nil, status.Error(codes.Unauthenticated,
				"invalid XAI address ownership proof")
		}
	}

	// All security checks passed - create or update the shared identity
	linked := map[string]string{"aura": msg.AuraAddress}
	if msg.PawAddress != "" {
		linked["paw"] = msg.PawAddress
	}
	if msg.XaiAddress != "" {
		linked["xai"] = msg.XaiAddress
	}

	identity := &bridgepb.SharedIdentity{
		Address:         msg.AuraAddress,
		VerifiedAura:    true, // Always true since signer verification passed
		VerifiedPaw:     msg.PawAddress != "" && len(msg.PawSignature) > 0,
		VerifiedXai:     msg.XaiAddress != "" && len(msg.XaiSignature) > 0,
		LinkedAddresses: linked,
		VerifiedAt:      timestamppb.New(ctx.BlockTime()),
	}

	if ms.Keeper.vcKeeper != nil {
		identity.AuraIrScore = ms.Keeper.vcKeeper.GetIRScore(ctx, msg.AuraAddress)
		identity.ReputationScore = identity.AuraIrScore * 10
	}

	ms.Keeper.setSharedIdentity(ctx, identity)

	// Emit event for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"link_address",
			sdk.NewAttribute("aura_address", msg.AuraAddress),
			sdk.NewAttribute("paw_address", msg.PawAddress),
			sdk.NewAttribute("xai_address", msg.XaiAddress),
			sdk.NewAttribute("verified_paw", fmt.Sprintf("%t", identity.VerifiedPaw)),
			sdk.NewAttribute("verified_xai", fmt.Sprintf("%t", identity.VerifiedXai)),
		),
	)

	return &bridgepb.MsgLinkAddressResponse{Success: true, LinkedIdentityId: msg.AuraAddress}, nil
}

// CrossChainSwap stores metadata about a requested swap route.
func (ms msgServer) CrossChainSwap(goCtx context.Context, msg *bridgepb.MsgCrossChainSwap) (*bridgepb.MsgCrossChainSwapResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.InputCoin == nil || !msg.InputCoin.Amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "input amount required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	swapID := fmt.Sprintf("swap-%s", ms.Keeper.nextTransferID(ctx))
	swap := &bridgepb.CrossChainSwap{
		SwapId:      swapID,
		Sender:      msg.Sender,
		SourceChain: normalizeChain(msg.SourceChain),
		TargetChain: normalizeChain(msg.TargetChain),
		SourceCoin:  msg.InputCoin,
		TargetDenom: msg.TargetDenom,
		Status:      "pending",
		InitiatedAt: timestamppb.New(ctx.BlockTime()),
	}
	ms.Keeper.setSwap(ctx, swap)
	route := []string{normalizeChain(msg.SourceChain), normalizeChain(msg.TargetChain)}
	if msg.SourceChain == msg.TargetChain {
		route = []string{normalizeChain(msg.SourceChain)}
	}
	return &bridgepb.MsgCrossChainSwapResponse{SwapId: swapID, Route: route, EstimatedCompletion: 900}, nil
}

// RelayTransfer updates transfer state based on relayer reports.
func (ms msgServer) RelayTransfer(goCtx context.Context, msg *bridgepb.MsgRelayTransfer) (*bridgepb.MsgRelayTransferResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.TransferId == "" {
		return nil, status.Error(codes.InvalidArgument, "transfer id required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	transfer, found := ms.Keeper.getTransfer(ctx, msg.TransferId)
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrTransferNotFound.Error())
	}
	switch strings.ToUpper(msg.Status) {
	case "PENDING":
		transfer.Status = bridgepb.TransferStatus_PENDING
	case "CONFIRMED":
		transfer.Status = bridgepb.TransferStatus_CONFIRMED
	case "RELAYED":
		transfer.Status = bridgepb.TransferStatus_RELAYED
	case "COMPLETED":
		transfer.Status = bridgepb.TransferStatus_COMPLETED
	case "FAILED":
		transfer.Status = bridgepb.TransferStatus_FAILED
	case "REFUNDED":
		transfer.Status = bridgepb.TransferStatus_REFUNDED
	}
	transfer.TargetTxHash = msg.TargetTxHash
	transfer.Timestamp = timestamppb.New(ctx.BlockTime())
	ms.Keeper.setTransfer(ctx, transfer)
	vol, _ := sdkmath.NewIntFromString(transfer.Amount)
	ms.Keeper.recordRelayerStats(ctx, msg.Relayer, true, vol)
	return &bridgepb.MsgRelayTransferResponse{Success: true}, nil
}
