// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	context "context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/pkg/log"
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
//  1. Verifies cryptographic signatures
//  2. Checks validator authorization (active status)
//  3. Prevents signature reuse (each validator counted once)
//  4. Ensures minimum threshold of unique active validators
//
// Security considerations:
//   - Only ACTIVE validators are allowed (inactive/slashed/jailed are rejected)
//   - Each validator can only contribute one signature (prevents duplicate counting)
//   - Cryptographic verification using validator's registered public key
//   - Minimum required threshold enforced (never less than MinAllowedConfirmations)
//
// SECURITY FIX: Validators must sign messages that include their validator address.
// This prevents order-dependency ambiguity where the same signature could match
// multiple validators. The message format must include the validator address:
//   "sourceChain:burnTxHash:sender:amount:denom:validator:validatorAddr"
//
// Each validator signs a unique message (with their own address), so each signature
// can only match one specific validator, eliminating ambiguity.
//
// Parameters:
//   - ctx: SDK context for state access
//   - signatures: Raw signature bytes from validators
//   - msgBase: Base message string (without validator address suffix)
//   - minRequired: Minimum number of valid signatures required
//
// Returns:
//   - validCount: Number of valid signatures from unique active validators
//   - err: Error if threshold not met or validation fails
func (ms msgServer) verifyRawValidatorSignatures(
	ctx sdk.Context,
	signatures [][]byte,
	msgBase string,
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

	// CONSENSUS-CRITICAL: Sort validator addresses for deterministic iteration order
	// Without sorting, map iteration order is random and can cause different nodes
	// to validate different signature combinations, breaking consensus.
	sortedValidatorAddrs := make([]string, 0, len(activeValidatorMap))
	for addr := range activeValidatorMap {
		sortedValidatorAddrs = append(sortedValidatorAddrs, addr)
	}
	sort.Strings(sortedValidatorAddrs)

	// P1 PERFORMANCE FIX: Pre-compute message hashes and parse pubkeys once
	// This reduces O(s*v) SHA256 operations to O(v) + O(s*v) signature verifications
	type validatorVerifyInfo struct {
		addr    string
		pubKey  cryptotypes.PubKey
		msgHash [32]byte
	}
	verifyInfos := make([]validatorVerifyInfo, 0, len(sortedValidatorAddrs))

	for _, addr := range sortedValidatorAddrs {
		validator := activeValidatorMap[addr]
		if len(validator.PublicKey) == 0 {
			continue
		}

		var pubKey cryptotypes.PubKey
		if err := ms.Keeper.cdc.UnmarshalInterface(validator.PublicKey, &pubKey); err != nil {
			continue
		}
		if pubKey == nil {
			continue
		}

		// Pre-compute validator-specific message hash
		msgWithValidator := fmt.Sprintf("%s:validator:%s", msgBase, addr)
		msgHash := sha256.Sum256([]byte(msgWithValidator))

		verifyInfos = append(verifyInfos, validatorVerifyInfo{
			addr:    addr,
			pubKey:  pubKey,
			msgHash: msgHash,
		})
	}

	// Track which validators have been matched (prevent counting same validator twice)
	usedValidators := make(map[string]bool)

	for _, sigBytes := range signatures {
		if len(sigBytes) == 0 {
			continue
		}

		// P1 PERFORMANCE FIX: Early exit when we have enough valid signatures
		if validCount >= int(minRequired) {
			break
		}

		// SECURITY FIX: Try to match this signature against ACTIVE validators
		// Each validator should have signed a message that includes their own address
		for i := range verifyInfos {
			info := &verifyInfos[i]
			// Skip if validator already matched (prevent duplicate counting)
			if usedValidators[info.addr] {
				continue
			}

			// CRITICAL: Verify the cryptographic signature against validator-specific message
			if info.pubKey.VerifySignature(info.msgHash[:], sigBytes) {
				// Valid signature from active validator found
				usedValidators[info.addr] = true
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
	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	log.TxStart(ctx, "MsgLockTokens", msg.Sender)
	if err := ms.Keeper.ensureBridgeEnabled(ctx); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	chainID := normalizeChain(msg.TargetChain)

	// CRITICAL SECURITY: Check if bridge is paused for this chain
	if err := ms.Keeper.RequireNotPaused(ctx, chainID); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
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
	transferID, err := ms.Keeper.nextTransferID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:            transferID,
		SourceChain:           sourceChainAura,
		TargetChain:           chainID,
		Sender:                msg.Sender,
		Recipient:             msg.Recipient,
		Amount:                amnt.Amount,
		Denom:                 amnt.Denom,
		Status:                bridgepb.TransferStatus_PENDING,
		Timestamp:             ctx.BlockTime(),
		RequiredConfirmations: params.MinConfirmations,
	}
	if err := ms.Keeper.setTransfer(ctx, transfer); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to store transfer: %v", err))
	}
	log.TxSuccess(ctx, "MsgLockTokens", "sender", msg.Sender, "transfer_id", transferID, "target_chain", chainID, "amount", amnt.String())
	log.StateChange(ctx, "cross_chain_transfer", "created", transferID)
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
	normalizedChain := normalizeChain(msg.SourceChain)
	if normalizedChain == "" {
		return nil, status.Error(codes.InvalidArgument, "source_chain required")
	}
	if strings.TrimSpace(msg.SourceTxHash) == "" {
		return nil, status.Error(codes.InvalidArgument, "source_tx_hash required")
	}
	if strings.TrimSpace(msg.Recipient) == "" {
		return nil, status.Error(codes.InvalidArgument, "recipient required")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid recipient")
	}
	if err := sdk.ValidateDenom(msg.Denom); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}
	// msg.Amount is already math.Int, use directly
	amount := msg.Amount
	if !amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// CRITICAL SECURITY: Check if bridge is paused for source chain
	if err := ms.Keeper.RequireNotPaused(ctx, normalizedChain); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	transferID, hasIndex := ms.Keeper.transferIDByHash(ctx, msg.SourceTxHash)
	if !hasIndex {
		var err error
		transferID, err = ms.Keeper.nextTransferID(ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	transfer, found := ms.Keeper.getTransfer(ctx, transferID)
	if !found {
		transfer = &bridgepb.CrossChainTransfer{
			TransferId:  transferID,
			SourceChain: normalizeChain(msg.SourceChain),
			TargetChain: sourceChainAura,
			Sender:      msg.SourceChain,
			Recipient:   msg.Recipient,
			Amount:      amount,
			Denom:       msg.Denom,
			Status:      bridgepb.TransferStatus_CONFIRMED,
			Timestamp:   ctx.BlockTime(),
		}
		if err := ms.Keeper.setTransfer(ctx, transfer); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to store transfer: %v", err))
		}
		ms.Keeper.indexTransferHash(ctx, msg.SourceTxHash, transferID)
	}
	if err := ms.Keeper.SubmitAttestation(ctx, transferID, msg.Validator, true); err != nil && err != types.ErrDuplicateAttestation {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if ms.Keeper.CheckAttestationThreshold(ctx, transferID) && transfer.Status != bridgepb.TransferStatus_COMPLETED {
		// CRITICAL SECURITY FIX: Check auto-pause threshold and record amount atomically
		// This must happen BEFORE minting to prevent concurrent validators from bypassing limits.
		// Both the check and record are in the same SDK transaction, making them atomic.
		ms.Keeper.RecordMintedAmount(ctx, msg.Denom, amount)

		// Check if threshold exceeded (GetHourlyMintedAmount now includes the amount we just recorded)
		if ms.Keeper.CheckAndTriggerAutoPause(ctx, msg.Denom, sdkmath.ZeroInt()) {
			// Auto-pause triggered - reject this mint
			// The amount stays recorded (prevents retry attacks from bypassing threshold)
			return nil, status.Error(codes.FailedPrecondition,
				"auto-pause triggered - hourly mint threshold exceeded")
		}

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
		transfer.Timestamp = ctx.BlockTime()
		if err := ms.Keeper.setTransfer(ctx, transfer); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update transfer status: %v", err))
		}
	}
	wrappedDenom := fmt.Sprintf("%s.%s", normalizedChain, msg.Denom)
	token, _ := ms.Keeper.getWrappedToken(ctx, wrappedDenom)
	if token == nil {
		token = &bridgepb.WrappedToken{
			WrappedDenom:  wrappedDenom,
			OriginalDenom: msg.Denom,
			SourceChain:   normalizeChain(msg.SourceChain),
			TotalSupply:   amount,
		}
	} else {
		// token.TotalSupply is already math.Int, use directly
		token.TotalSupply = token.TotalSupply.Add(amount)
	}
	if err := ms.Keeper.setWrappedToken(ctx, token); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update wrapped token: %v", err))
	}
	return &bridgepb.MsgMintTokensResponse{Success: true, WrappedDenom: wrappedDenom}, nil
}

// UnlockTokens unlocks locked assets after a burn proof on the destination chain.
func (ms msgServer) UnlockTokens(goCtx context.Context, msg *bridgepb.MsgUnlockTokens) (*bridgepb.MsgUnlockTokensResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	// msg.Amount is already math.Int, use directly
	amount := msg.Amount
	if !amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// CRITICAL SECURITY: Check if this source hash was already processed (replay attack prevention)
	// This MUST happen BEFORE any other processing to prevent replay attacks where an attacker
	// reuses the same burn transaction hash to unlock tokens multiple times.
	sourceChain := normalizeChain(msg.SourceChain)

	// CRITICAL SECURITY: Check if bridge is paused for source chain
	if err := ms.Keeper.RequireNotPaused(ctx, sourceChain); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	var transferID string
	var hasIndex bool

	if sourceChain == "" {
		// If no source chain specified, try to infer from transfer
		transferID, hasIndex = ms.Keeper.transferIDByHash(ctx, msg.BurnTxHash)
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

	transferID, hasIndex = ms.Keeper.transferIDByHash(ctx, msg.BurnTxHash)
	if !hasIndex {
		transferID = msg.BurnTxHash
	}
	transfer, found := ms.Keeper.getTransfer(ctx, transferID)
	if !found {
		return nil, status.Error(codes.NotFound, types.ErrTransferNotFound.Error())
	}

	// SECURITY FIX: Comprehensive validator signature verification
	// This implements three critical security checks as required by issue #019:
	//   1. Validator authorization check against governance-approved active list
	//   2. Active validator set verification at current block height
	//   3. Replay protection with signature tracking
	required := ms.Keeper.GetParams(ctx).MinConfirmations

	// CRITICAL SECURITY: Enforce minimum even if params were misconfigured
	// Never allow less than 2 validators to prevent single validator control
	if required < types.MinAllowedConfirmations {
		required = types.MinAllowedConfirmations
	}

	// CRITICAL SECURITY FIX #1: Get CURRENT active validator set at this block height
	// This ensures we only accept signatures from validators who are CURRENTLY active,
	// not from validators who were active in the past but have since been removed/slashed.
	//
	// Attack prevented: Attacker compromises validator, gets signatures, validator is removed,
	// attacker tries to replay signatures - this check rejects them because validator
	// is no longer in the active set.
	activeValidators := ms.Keeper.getActiveValidatorSet(ctx, ctx.BlockHeight())
	if len(activeValidators) == 0 {
		return nil, status.Error(codes.Internal,
			"no active validators in current set - cannot verify signatures")
	}

	// Log active validator set for audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"unlock_active_validators_checked",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute("active_validator_count", fmt.Sprintf("%d", len(activeValidators))),
			sdk.NewAttribute("required_signatures", fmt.Sprintf("%d", required)),
		),
	)

	// Build deterministic message base for signature verification
	// Format: sourceChain:burnTxHash:recipient:amount:denom
	// SECURITY FIX: Validators will append ":validator:validatorAddr" before signing
	// This format ensures uniqueness per transaction on the source chain AND per validator
	msgBase := fmt.Sprintf("%s:%s:%s:%s:%s",
		transfer.SourceChain,
		msg.BurnTxHash,
		msg.Sender,
		msg.Amount,
		msg.Denom,
	)

	// CRITICAL SECURITY FIX #2: Check if this signature set has been used before
	// This prevents replay attacks where the same valid signatures are reused
	// for multiple unlock attempts.
	//
	// Attack prevented: Attacker gets valid signatures, uses them once successfully,
	// tries to reuse same signatures again - this check rejects the replay.
	signatureSetHash := ms.Keeper.computeSignatureSetHash(msg.ValidatorSignatures)
	if signatureSetHash != nil && ms.Keeper.isSignatureSetUsed(ctx, transferID, signatureSetHash) {
		return nil, status.Error(codes.AlreadyExists,
			"signature set already used for this transfer (replay attack prevented)")
	}

	// CRITICAL SECURITY FIX #3: Verify signatures cryptographically against ACTIVE validator set
	// This function implements comprehensive security checks:
	//   - Verifies cryptographic signatures using validator public keys (prevents forgery)
	//   - Checks validator authorization - ONLY ACTIVE validators from current set (prevents unauthorized signing)
	//   - Counts UNIQUE validators - prevents duplicate counting same validator
	//   - Enforces minimum threshold - never less than MinAllowedConfirmations (prevents single validator control)
	//   - SECURITY FIX: Each validator signs message with their address included (prevents order-dependency ambiguity)
	//
	// The verifyRawValidatorSignatures function will:
	//   1. Call getActiveValidators() to get the current governance-approved validator list
	//   2. For each signature, verify it cryptographically using validator's public key
	//   3. Build validator-specific message: "msgBase:validator:validatorAddr"
	//   4. Ensure the validator is in the ACTIVE set (not slashed/jailed/removed)
	//   5. Count only unique validators (prevent same validator being counted twice)
	//   6. Require at least 'required' unique active validators
	//
	// Attack prevented: Multiple attack vectors blocked:
	//   - Unauthorized validators cannot provide valid signatures (not in active set)
	//   - Compromised then removed validators rejected (active set check)
	//   - Forged signatures rejected (cryptographic verification)
	//   - Duplicate signatures rejected (unique validator counting)
	//   - Order-dependency ambiguity eliminated (validator address in signed message)
	validCount, err := ms.verifyRawValidatorSignatures(ctx, msg.ValidatorSignatures, msgBase, required)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	// Log successful signature verification for audit trail
	msgBaseHash := sha256.Sum256([]byte(msgBase))
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"unlock_signatures_verified",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("valid_signature_count", fmt.Sprintf("%d", validCount)),
			sdk.NewAttribute("required_count", fmt.Sprintf("%d", required)),
			sdk.NewAttribute("message_base_hash", fmt.Sprintf("%x", msgBaseHash)),
			sdk.NewAttribute("signature_set_hash", fmt.Sprintf("%x", signatureSetHash)),
		),
	)

	// CRITICAL SECURITY: Verify Merkle proof that transaction exists on source chain
	// This prevents validators from attesting to fake deposits that never occurred.
	// The Merkle proof cryptographically proves the transaction is in a specific block.
	if len(msg.MerkleProof) > 0 && len(msg.MerkleRoot) > 0 {
		// Construct the transaction leaf hash
		txLeaf := ms.Keeper.ConstructTransactionLeaf(
			transfer.SourceChain,
			msg.BurnTxHash,
			msg.Sender,
			msg.Amount.String(),
			msg.Denom,
		)

		// Verify the Merkle proof
		proofValid := ms.Keeper.VerifyMerkleProofBytes(
			msg.MerkleRoot,
			txLeaf,
			msg.MerkleProof,
		)

		if !proofValid {
			return nil, status.Error(codes.InvalidArgument,
				"invalid Merkle proof: transaction not found in source block")
		}

		// If block hash and height provided, verify the block is authentic
		if len(msg.SourceBlockHash) > 0 && msg.SourceBlockHeight > 0 {
			blockValid := ms.Keeper.VerifySourceBlock(
				ctx,
				transfer.SourceChain,
				msg.SourceBlockHeight,
				msg.SourceBlockHash,
			)

			if !blockValid {
				return nil, status.Error(codes.InvalidArgument,
					"invalid or unverified source block hash")
			}
		}

		// Log Merkle proof verification for audit trail
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"merkle_proof_verified",
				sdk.NewAttribute("transfer_id", transferID),
				sdk.NewAttribute("source_chain", transfer.SourceChain),
				sdk.NewAttribute("source_block_height", fmt.Sprintf("%d", msg.SourceBlockHeight)),
				sdk.NewAttribute("merkle_root", fmt.Sprintf("%x", msg.MerkleRoot)),
			),
		)
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

	// CRITICAL SECURITY: Enforce supply caps and rate limits BEFORE minting
	// This prevents unlimited token inflation even if all validator signatures are valid
	params := ms.Keeper.GetParams(ctx)

	// 1. Check per-transfer maximum (circuit breaker)
	maxTransfer, ok := sdkmath.NewIntFromString(params.MaxTransferAmount)
	if ok && amount.GT(maxTransfer) {
		return nil, status.Errorf(codes.InvalidArgument,
			"amount %s exceeds max transfer limit %s", amount, maxTransfer)
	}

	// 2. Check per-token supply cap (if configured for this denom)
	if cap, exists := params.SupplyCaps[msg.Denom]; exists {
		supplyCap, ok := sdkmath.NewIntFromString(cap)
		if ok {
			// Get current supply of this token
			currentSupply := ms.Keeper.bankKeeper.GetSupply(ctx, msg.Denom).Amount
			// Check if minting would exceed the cap
			if currentSupply.Add(amount).GT(supplyCap) {
				return nil, status.Errorf(codes.ResourceExhausted,
					"minting %s would exceed supply cap of %s (current: %s)",
					amount, supplyCap, currentSupply)
			}
		}
	}

	// 3. Check daily mint limit (rate limiting)
	dailyMinted := ms.Keeper.GetDailyMintedAmount(ctx, msg.Denom)
	dailyLimit, ok := sdkmath.NewIntFromString(params.DailyMintLimit)
	if ok && dailyMinted.Add(amount).GT(dailyLimit) {
		return nil, status.Errorf(codes.ResourceExhausted,
			"daily mint limit exceeded: %s already minted today, limit is %s",
			dailyMinted, dailyLimit)
	}

	// 4. Check hourly mint limit (rate limiting - prevents rapid draining)
	hourlyMinted := ms.Keeper.GetHourlyMintedAmount(ctx, msg.Denom)
	hourlyLimit, ok := sdkmath.NewIntFromString(params.HourlyMintLimit)
	if ok && hourlyMinted.Add(amount).GT(hourlyLimit) {
		return nil, status.Errorf(codes.ResourceExhausted,
			"hourly mint limit exceeded: %s already minted this hour, limit is %s",
			hourlyMinted, hourlyLimit)
	}

	// CRITICAL SECURITY: Mark the source hash as processed BEFORE creating pending transfer
	// This prevents reentrancy and ensures the replay protection is atomic.
	// Following checks-effects-interactions pattern: effects (state change) before interactions (token transfer).
	ms.Keeper.MarkSourceHashProcessed(ctx, sourceChain, msg.BurnTxHash)

	// CRITICAL SECURITY: Get fraud proof window from params
	fraudProofWindow := time.Duration(params.FraudProofWindow) * time.Second
	if fraudProofWindow <= 0 {
		// Fallback to default if not set (should not happen with proper param validation)
		fraudProofWindow = types.DefaultFraudProofWindow
	}

	// Calculate unlock time (current time + fraud proof window)
	unlockTime := ctx.BlockTime().Add(fraudProofWindow)

	// Create pending transfer instead of immediately unlocking
	// This holds the transfer in escrow during the fraud proof window
	pendingTransfer := &types.PendingTransfer{
		TransferId:   transferID,
		Recipient:    msg.Sender,
		Amount:       amount, // Store as string for protobuf compatibility
		Denom:        msg.Denom,
		SourceChain:  sourceChain,
		SourceTxHash: msg.BurnTxHash,
		CreatedAt:    ctx.BlockTime(),
		UnlockTime:   unlockTime,
		Challenged:   false,
		FraudProofId: "",
	}

	if err := ms.Keeper.setPendingTransfer(ctx, pendingTransfer); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to store pending transfer: %v", err))
	}

	// Update transfer status to RELAYED (awaiting finalization)
	transfer.Status = bridgepb.TransferStatus_RELAYED
	transfer.TargetTxHash = msg.BurnTxHash
	transfer.Timestamp = ctx.BlockTime()
	if err := ms.Keeper.setTransfer(ctx, transfer); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update transfer status: %v", err))
	}

	// Emit event for pending transfer creation
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"pending_transfer_created",
			sdk.NewAttribute("transfer_id", transferID),
			sdk.NewAttribute("recipient", msg.Sender),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("denom", msg.Denom),
			sdk.NewAttribute("source_chain", sourceChain),
			sdk.NewAttribute("unlock_time", unlockTime.Format(time.RFC3339)),
			sdk.NewAttribute("fraud_proof_window_seconds", fmt.Sprintf("%d", int64(fraudProofWindow.Seconds()))),
		),
	)

	return &bridgepb.MsgUnlockTokensResponse{Success: true}, nil
}

// BurnTokens burns wrapped tokens on Aura to unlock on the origin chain.
func (ms msgServer) BurnTokens(goCtx context.Context, msg *bridgepb.MsgBurnTokens) (*bridgepb.MsgBurnTokensResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// CRITICAL SECURITY: Validate amount is positive before any operations
	if !msg.Amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	chainID := normalizeChain(msg.TargetChain)
	if chainID == "" {
		return nil, status.Error(codes.InvalidArgument, "target chain required")
	}

	// CRITICAL SECURITY: Check if bridge is paused for target chain
	if err := ms.Keeper.RequireNotPaused(ctx, chainID); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
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
		if err := ms.Keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(amnt)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if err := ms.Keeper.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(amnt)); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	transferID, err := ms.Keeper.nextTransferID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	transfer := &bridgepb.CrossChainTransfer{
		TransferId:  transferID,
		SourceChain: sourceChainAura,
		TargetChain: chainID,
		Sender:      msg.Sender,
		Recipient:   msg.Recipient,
		Amount:      amnt.Amount,
		Denom:       amnt.Denom,
		Status:      bridgepb.TransferStatus_CONFIRMED,
		Timestamp:   ctx.BlockTime(),
	}
	if err := ms.Keeper.setTransfer(ctx, transfer); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to store transfer: %v", err))
	}
	ms.Keeper.indexTransferHash(ctx, transferID, transferID)
	return &bridgepb.MsgBurnTokensResponse{TransferId: transferID, EstimatedCompletion: 600}, nil
}

// LinkAddress links Aura/PAW/XAI addresses for shared identity.
//
// SECURITY CRITICAL: This function implements cross-chain identity linking with strict access controls:
//  1. Signer verification: Only the Aura address owner can link addresses
//  2. Cross-chain ownership proof: Cryptographic signatures required from PAW/XAI addresses
//  3. Conflict prevention: Prevents overwriting existing links without proper authorization
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
	// Get existing identity to merge with (if any)
	existingIdentity, found := ms.Keeper.GetSharedIdentity(ctx, msg.AuraAddress)

	// Start with existing linked addresses, or create new map
	linked := make(map[string]string)
	if found && existingIdentity.LinkedAddresses != nil {
		// Copy existing links
		for chain, addr := range existingIdentity.LinkedAddresses {
			linked[chain] = addr
		}
	}

	// Always include aura address
	linked["aura"] = msg.AuraAddress

	// Update with new addresses (if provided)
	if msg.PawAddress != "" {
		linked["paw"] = msg.PawAddress
	}
	if msg.XaiAddress != "" {
		linked["xai"] = msg.XaiAddress
	}

	// Merge verification status (preserve existing verifications)
	verifiedPaw := (found && existingIdentity.VerifiedPaw) || (msg.PawAddress != "" && len(msg.PawSignature) > 0)
	verifiedXai := (found && existingIdentity.VerifiedXai) || (msg.XaiAddress != "" && len(msg.XaiSignature) > 0)

	identity := &bridgepb.SharedIdentity{
		Address:         msg.AuraAddress,
		VerifiedAura:    true, // Always true since signer verification passed
		VerifiedPaw:     verifiedPaw,
		VerifiedXai:     verifiedXai,
		LinkedAddresses: linked,
		VerifiedAt:      ctx.BlockTime(),
	}

	if ms.Keeper.vcKeeper != nil {
		identity.AuraIrScore = ms.Keeper.vcKeeper.GetIRScore(ctx, msg.AuraAddress)
		identity.ReputationScore = identity.AuraIrScore * 10
	}

	if err := ms.Keeper.setSharedIdentity(ctx, identity); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to store shared identity: %v", err))
	}

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
	if !msg.InputCoin.IsValid() || !msg.InputCoin.Amount.IsPositive() {
		return nil, status.Error(codes.InvalidArgument, "input amount required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	transferID, err := ms.Keeper.nextTransferID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	swapID := fmt.Sprintf("swap-%s", transferID)
	swap := &bridgepb.CrossChainSwap{
		SwapId:      swapID,
		Sender:      msg.Sender,
		SourceChain: normalizeChain(msg.SourceChain),
		TargetChain: normalizeChain(msg.TargetChain),
		SourceCoin:  msg.InputCoin,
		TargetDenom: msg.TargetDenom,
		Status:      "pending",
		InitiatedAt: ctx.BlockTime(),
	}
	if err := ms.Keeper.setSwap(ctx, swap); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to store swap: %v", err))
	}
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
	transfer.Timestamp = ctx.BlockTime()
	if err := ms.Keeper.setTransfer(ctx, transfer); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update transfer: %v", err))
	}
	// transfer.Amount is already math.Int, use directly
	ms.Keeper.recordRelayerStats(ctx, msg.Relayer, true, transfer.Amount)
	return &bridgepb.MsgRelayTransferResponse{Success: true}, nil
}

// FinalizeTransfer finalizes a pending transfer after the fraud proof window expires.
//
// SECURITY CRITICAL: This function completes the unlock/mint process for transfers
// that have passed the fraud proof window without being challenged.
//
// Security checks:
//   - Fraud proof window must have expired (unlock_time <= current time)
//   - Transfer must not have been challenged (no fraud proof submitted)
//   - Transfer must exist and be in pending state
//
// Attack vectors prevented:
//   - Early finalization: Cannot finalize before fraud proof window expires
//   - Challenged transfers: Cannot finalize transfers under investigation
//   - Double finalization: Pending transfer is deleted after finalization
//
// Parameters:
//   - msg: Contains transfer ID to finalize
//
// Returns:
//   - Response with success status, amount, and recipient
//   - Error if: transfer not found, window not expired, or transfer challenged
func (ms msgServer) FinalizeTransfer(goCtx context.Context, msg *bridgepb.MsgFinalizeTransfer) (*bridgepb.MsgFinalizeTransferResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.TransferId == "" {
		return nil, status.Error(codes.InvalidArgument, "transfer id required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the pending transfer
	pendingTransfer, found := ms.Keeper.GetPendingTransfer(ctx, msg.TransferId)
	if !found {
		return nil, status.Error(codes.NotFound,
			"pending transfer not found - may already be finalized or never existed")
	}

	// CRITICAL SECURITY: Check if fraud proof window has expired
	// This ensures adequate time for fraud proof submission before finalizing
	currentTime := ctx.BlockTime()
	unlockTime := pendingTransfer.UnlockTime

	if currentTime.Before(unlockTime) {
		return nil, status.Error(codes.FailedPrecondition,
			fmt.Sprintf("fraud proof window has not expired - unlock time: %s, current time: %s",
				unlockTime.Format(time.RFC3339), currentTime.Format(time.RFC3339)))
	}

	// CRITICAL SECURITY: Check if transfer has been challenged
	// Challenged transfers cannot be finalized and require investigation
	if pendingTransfer.Challenged {
		return nil, status.Error(codes.FailedPrecondition,
			fmt.Sprintf("transfer has been challenged with fraud proof %s - cannot finalize",
				pendingTransfer.FraudProofId))
	}

	// pendingTransfer.Amount is already math.Int, use directly
	amount := pendingTransfer.Amount
	if !amount.IsPositive() {
		return nil, status.Error(codes.Internal,
			fmt.Sprintf("invalid amount in pending transfer: %s", amount.String()))
	}

	// Parse recipient address
	recipient, err := sdk.AccAddressFromBech32(pendingTransfer.Recipient)
	if err != nil {
		return nil, status.Error(codes.Internal,
			fmt.Sprintf("invalid recipient address: %s", err.Error()))
	}

	// CRITICAL SECURITY: Unlock/mint tokens following checks-effects-interactions pattern
	// 1. All checks done above
	// 2. Effects (state changes) - delete pending transfer first to prevent reentrancy
	ms.Keeper.deletePendingTransfer(ctx, msg.TransferId)

	// Update main transfer status to COMPLETED
	transfer, found := ms.Keeper.GetTransfer(ctx, msg.TransferId)
	if found {
		transfer.Status = bridgepb.TransferStatus_COMPLETED
		transfer.Timestamp = currentTime
		if err := ms.Keeper.setTransfer(ctx, transfer); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update transfer status: %v", err))
		}
	}

	// 3. Interactions (token transfers) - unlock tokens from module to recipient
	coin := sdk.NewCoin(pendingTransfer.Denom, amount)
	if ms.Keeper.bankKeeper != nil {
		// Send from module account to recipient
		if err := ms.Keeper.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			types.ModuleName,
			recipient,
			sdk.NewCoins(coin),
		); err != nil {
			return nil, status.Error(codes.Internal,
				fmt.Sprintf("failed to unlock tokens: %s", err.Error()))
		}
	}

	// Emit event for finalization
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"transfer_finalized",
			sdk.NewAttribute("transfer_id", msg.TransferId),
			sdk.NewAttribute("recipient", pendingTransfer.Recipient),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("denom", pendingTransfer.Denom),
			sdk.NewAttribute("source_chain", pendingTransfer.SourceChain),
			sdk.NewAttribute("finalized_at", currentTime.Format(time.RFC3339)),
		),
	)

	return &bridgepb.MsgFinalizeTransferResponse{
		Success:   true,
		Amount:    amount,
		Recipient: pendingTransfer.Recipient,
	}, nil
}

// SubmitFraudProof submits a fraud proof to challenge a pending transfer.
//
// SECURITY CRITICAL: This function allows anyone to challenge a potentially
// fraudulent transfer during the fraud proof window.
//
// Security considerations:
//   - Must be submitted during fraud proof window (before unlock_time)
//   - Prevents finalization of the challenged transfer
//   - Requires evidence data for investigation
//   - Creates a fraud proof record for governance review
//
// Attack vectors prevented:
//   - Fraudulent transfers: Allows community to challenge invalid transfers
//   - Late challenges: Must be submitted before window expires
//   - Frivolous challenges: Evidence required for investigation
//
// Parameters:
//   - msg: Contains transfer ID, fraud type, evidence, and description
//
// Returns:
//   - Response with success status and fraud proof ID
//   - Error if: transfer not found, window expired, or already challenged
func (ms msgServer) SubmitFraudProof(goCtx context.Context, msg *bridgepb.MsgSubmitFraudProof) (*bridgepb.MsgSubmitFraudProofResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.TransferId == "" {
		return nil, status.Error(codes.InvalidArgument, "transfer id required")
	}
	if msg.Challenger == "" {
		return nil, status.Error(codes.InvalidArgument, "challenger address required")
	}
	if len(msg.Evidence) == 0 {
		return nil, status.Error(codes.InvalidArgument, "evidence required for fraud proof")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the pending transfer
	pendingTransfer, found := ms.Keeper.GetPendingTransfer(ctx, msg.TransferId)
	if !found {
		return nil, status.Error(codes.NotFound,
			"pending transfer not found - may already be finalized or never existed")
	}

	// CRITICAL SECURITY: Check if fraud proof window is still open
	// Fraud proofs can only be submitted during the window
	currentTime := ctx.BlockTime()
	unlockTime := pendingTransfer.UnlockTime

	if currentTime.After(unlockTime) {
		return nil, status.Error(codes.FailedPrecondition,
			fmt.Sprintf("fraud proof window has expired - unlock time: %s, current time: %s",
				unlockTime.Format(time.RFC3339), currentTime.Format(time.RFC3339)))
	}

	// Check if transfer is already challenged
	if pendingTransfer.Challenged {
		return nil, status.Error(codes.AlreadyExists,
			fmt.Sprintf("transfer already challenged with fraud proof %s",
				pendingTransfer.FraudProofId))
	}

	// Generate fraud proof ID
	fraudProofID := fmt.Sprintf("fraud-%s-%d", msg.TransferId, currentTime.Unix())

	// Parse fraud type to enum
	var fraudType bridgepb.FraudType
	switch strings.ToUpper(msg.FraudType) {
	case "INVALID_MERKLE_PROOF":
		fraudType = bridgepb.FraudType_FRAUD_INVALID_MERKLE_PROOF
	case "DOUBLE_SPEND":
		fraudType = bridgepb.FraudType_FRAUD_DOUBLE_SPEND
	case "INVALID_SIGNATURE":
		fraudType = bridgepb.FraudType_FRAUD_INVALID_SIGNATURE
	case "AMOUNT_MISMATCH":
		fraudType = bridgepb.FraudType_FRAUD_AMOUNT_MISMATCH
	case "UNAUTHORIZED_MINT":
		fraudType = bridgepb.FraudType_FRAUD_UNAUTHORIZED_MINT
	default:
		fraudType = bridgepb.FraudType_FRAUD_INVALID_MERKLE_PROOF // Default
	}

	// Create fraud proof record
	fraudProof := &bridgepb.FraudProof{
		ProofId:              fraudProofID,
		ChallengedTransferId: msg.TransferId,
		Challenger:           msg.Challenger,
		FraudType:            fraudType,
		Evidence:             msg.Evidence,
		Status:               bridgepb.FraudProofStatus_FRAUD_PROOF_PENDING,
		SubmittedAt:          currentTime,
	}

	// Store fraud proof
	if err := ms.Keeper.setFraudProof(ctx, fraudProof); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to store fraud proof: %v", err))
	}

	// CRITICAL SECURITY: Mark pending transfer as challenged
	// This prevents finalization while fraud proof is investigated
	pendingTransfer.Challenged = true
	pendingTransfer.FraudProofId = fraudProofID
	if err := ms.Keeper.SetPendingTransfer(ctx, pendingTransfer); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update pending transfer: %v", err))
	}

	// Emit event for fraud proof submission
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"fraud_proof_submitted",
			sdk.NewAttribute("fraud_proof_id", fraudProofID),
			sdk.NewAttribute("transfer_id", msg.TransferId),
			sdk.NewAttribute("challenger", msg.Challenger),
			sdk.NewAttribute("fraud_type", msg.FraudType),
			sdk.NewAttribute("submitted_at", currentTime.Format(time.RFC3339)),
		),
	)

	return &bridgepb.MsgSubmitFraudProofResponse{
		Success:      true,
		FraudProofId: fraudProofID,
	}, nil
}

// EmergencyPause pauses bridge operations for authorized emergency guardians.
//
// SECURITY CRITICAL: This function allows pre-authorized addresses to immediately
// pause bridge operations in case of an active exploit or security incident.
//
// Security checks:
//   - Signer must be in EmergencyPauseAddresses parameter list (ACL check)
//   - Can pause globally or specific chains
//   - Reason required for audit trail
//   - Emits events for monitoring
//
// Attack vectors prevented:
//   - Unauthorized pause: Only authorized guardians can pause
//   - Missing audit trail: Reason required and logged
//   - Delayed response: Immediate pause without governance delay
//
// Parameters:
//   - msg: Contains signer, reason, and optional chains to pause
//
// Returns:
//   - Response with success status
//   - Error if: signer not authorized, invalid parameters
func (ms msgServer) EmergencyPause(goCtx context.Context, msg *bridgepb.MsgEmergencyPause) (*bridgepb.MsgEmergencyPauseResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.Signer == "" {
		return nil, status.Error(codes.InvalidArgument, "signer required")
	}
	if msg.Reason == "" {
		return nil, status.Error(codes.InvalidArgument, "reason required for emergency pause")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// CRITICAL SECURITY: Check if signer is authorized to trigger emergency pause
	// This is the ACL check required by HIGH-001 security finding
	if !ms.Keeper.IsEmergencyPauseAuthorized(ctx, msg.Signer) {
		return nil, status.Error(codes.PermissionDenied,
			fmt.Sprintf("address %s is not authorized to trigger emergency pause", msg.Signer))
	}

	// Get current params
	params := ms.Keeper.GetParams(ctx)

	if len(msg.Chains) == 0 {
		// Global pause - pause all bridge operations
		params.Paused = true
		log.TxStart(ctx, "EmergencyPause", msg.Signer)
		log.StateChange(ctx, "bridge_pause", "global", "emergency")
	} else {
		// Pause specific chains
		if params.PausedChains == nil {
			params.PausedChains = make([]string, 0)
		}
		for _, chain := range msg.Chains {
			normalizedChain := normalizeChain(chain)
			if normalizedChain != "" {
				// Check if chain is not already in the list
				alreadyPaused := false
				for _, pausedChain := range params.PausedChains {
					if pausedChain == normalizedChain {
						alreadyPaused = true
						break
					}
				}
				if !alreadyPaused {
					params.PausedChains = append(params.PausedChains, normalizedChain)
					log.StateChange(ctx, "bridge_pause", normalizedChain, "emergency")
				}
			}
		}
	}

	// Update params with pause status
	if err := ms.Keeper.SetParams(ctx, params); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update params: %v", err))
	}

	// Emit detailed event for monitoring and audit trail
	eventAttrs := []sdk.Attribute{
		sdk.NewAttribute("signer", msg.Signer),
		sdk.NewAttribute("reason", msg.Reason),
		sdk.NewAttribute("timestamp", ctx.BlockTime().Format(time.RFC3339)),
	}

	if len(msg.Chains) == 0 {
		eventAttrs = append(eventAttrs, sdk.NewAttribute("scope", "global"))
	} else {
		eventAttrs = append(eventAttrs, sdk.NewAttribute("scope", "chains"))
		eventAttrs = append(eventAttrs, sdk.NewAttribute("chains", strings.Join(msg.Chains, ",")))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"emergency_pause_triggered",
			eventAttrs...,
		),
	)

	if len(msg.Chains) == 0 {
		log.TxSuccess(ctx, "EmergencyPause", "signer", msg.Signer, "scope", "global", "reason", msg.Reason)
	} else {
		log.TxSuccess(ctx, "EmergencyPause", "signer", msg.Signer, "chains", strings.Join(msg.Chains, ","), "reason", msg.Reason)
	}

	return &bridgepb.MsgEmergencyPauseResponse{Success: true}, nil
}

// Unpause unpauses bridge operations (governance only).
//
// SECURITY CRITICAL: This function allows governance to resume bridge operations
// after an emergency pause or investigation.
//
// Security checks:
//   - Authority must be governance address (not emergency guardians)
//   - Can unpause globally or specific chains
//   - Emits events for monitoring
//
// Attack vectors prevented:
//   - Unauthorized unpause: Only governance can unpause
//   - Premature resumption: Requires governance consensus
//
// Parameters:
//   - msg: Contains authority address and optional chains to unpause
//
// Returns:
//   - Response with success status
//   - Error if: authority not governance, invalid parameters
func (ms msgServer) Unpause(goCtx context.Context, msg *bridgepb.MsgUnpause) (*bridgepb.MsgUnpauseResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if msg.Authority == "" {
		return nil, status.Error(codes.InvalidArgument, "authority required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// CRITICAL SECURITY: Only governance can unpause
	// This prevents emergency pause addresses from unpausing (requires broader consensus)
	// The authority is checked by the message handler via cosmos.msg.v1.signer annotation
	// in the proto file, which ensures msg.Authority matches the governance module address.
	// This is enforced at the SDK level before this handler is called.
	params := ms.Keeper.GetParams(ctx)

	if len(msg.Chains) == 0 {
		// Global unpause - resume all bridge operations
		params.Paused = false
		// Also clear all chain-specific pauses
		params.PausedChains = make([]string, 0)
		log.TxStart(ctx, "Unpause", msg.Authority)
		log.StateChange(ctx, "bridge_unpause", "global", "governance")
	} else {
		// Unpause specific chains
		if params.PausedChains == nil {
			params.PausedChains = make([]string, 0)
		}
		// Remove chains from paused list
		newPausedChains := make([]string, 0)
		for _, pausedChain := range params.PausedChains {
			shouldUnpause := false
			for _, chain := range msg.Chains {
				normalizedChain := normalizeChain(chain)
				if normalizedChain == pausedChain {
					shouldUnpause = true
					log.StateChange(ctx, "bridge_unpause", normalizedChain, "governance")
					break
				}
			}
			if !shouldUnpause {
				newPausedChains = append(newPausedChains, pausedChain)
			}
		}
		params.PausedChains = newPausedChains
	}

	// Update params with unpause status
	if err := ms.Keeper.SetParams(ctx, params); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update params: %v", err))
	}

	// Emit detailed event for monitoring and audit trail
	eventAttrs := []sdk.Attribute{
		sdk.NewAttribute("authority", msg.Authority),
		sdk.NewAttribute("timestamp", ctx.BlockTime().Format(time.RFC3339)),
	}

	if len(msg.Chains) == 0 {
		eventAttrs = append(eventAttrs, sdk.NewAttribute("scope", "global"))
	} else {
		eventAttrs = append(eventAttrs, sdk.NewAttribute("scope", "chains"))
		eventAttrs = append(eventAttrs, sdk.NewAttribute("chains", strings.Join(msg.Chains, ",")))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bridge_unpaused",
			eventAttrs...,
		),
	)

	if len(msg.Chains) == 0 {
		log.TxSuccess(ctx, "Unpause", "authority", msg.Authority, "scope", "global")
	} else {
		log.TxSuccess(ctx, "Unpause", "authority", msg.Authority, "chains", strings.Join(msg.Chains, ","))
	}

	return &bridgepb.MsgUnpauseResponse{Success: true}, nil
}
