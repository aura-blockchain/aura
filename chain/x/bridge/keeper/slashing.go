package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

// ========================================================================
// VALIDATOR SLASHING
// ========================================================================

// SubmitSlashingEvidence submits evidence of validator misbehavior and executes slashing.
//
// SECURITY CRITICAL: This function implements economic deterrent against malicious validators.
// Validators can be slashed for:
//   - Signing fraudulent transfers (invalid Merkle proofs, fake deposits)
//   - Double-signing (signing conflicting messages for the same transfer)
//   - Being offline (failing to sign sufficient transfers)
//
// Parameters:
//   - ctx: SDK context for state access and block height
//   - validatorAddress: Address of the validator to slash
//   - reason: SlashReason enum (FRAUD, DOUBLE_SIGN, DOWNTIME)
//   - transferId: Transfer ID related to the infraction (if applicable)
//   - evidenceHash: Hash of the evidence proving the infraction
//   - submitter: Address of the entity submitting the evidence
//
// Returns:
//   - SlashingEvent: The created slashing event with details
//   - error: If validation fails or slashing cannot be executed
func (k Keeper) SubmitSlashingEvidence(
	ctx sdk.Context,
	validatorAddress string,
	reason types.SlashReason,
	transferId string,
	evidenceHash []byte,
	submitter string,
) (*types.SlashingEvent, error) {
	// 1. Input validation
	if validatorAddress == "" {
		return nil, fmt.Errorf("validator address cannot be empty")
	}
	if len(evidenceHash) == 0 {
		return nil, types.ErrInvalidEvidence
	}

	// 2. Verify validator exists and is registered as bridge validator
	validator, found := k.getValidator(ctx, validatorAddress)
	if !found {
		return nil, types.ErrValidatorNotFound
	}

	// 3. Check if validator is already jailed for this offense
	// (prevent double-slashing for the same infraction)
	if !validator.Active {
		return nil, types.ErrValidatorJailed
	}

	// 4. Get slashing parameters
	params := k.GetParams(ctx)
	var slashFraction sdkmath.LegacyDec
	var jailValidator bool

	switch reason {
	case types.SlashReason_SLASH_INVALID_PROOF, types.SlashReason_SLASH_FRAUD_ATTEMPT:
		// Fraudulent signature - slash fraud fraction
		var err error
		slashFraction, err = sdkmath.LegacyNewDecFromStr(params.SlashFraudSignature)
		if err != nil {
			return nil, fmt.Errorf("invalid slash fraud signature param: %w", err)
		}
		jailValidator = true // Jail validators who attempt fraud

	case types.SlashReason_SLASH_DOUBLE_SIGN:
		// Double signing - most severe punishment (tombstone)
		var err error
		slashFraction, err = sdkmath.LegacyNewDecFromStr(params.SlashDoubleSigning)
		if err != nil {
			return nil, fmt.Errorf("invalid slash double signing param: %w", err)
		}
		jailValidator = true // Permanently jail double-signers

	case types.SlashReason_SLASH_DOWNTIME:
		// Offline/downtime - minor slash
		var err error
		slashFraction, err = sdkmath.LegacyNewDecFromStr(params.SlashOffline)
		if err != nil {
			return nil, fmt.Errorf("invalid slash offline param: %w", err)
		}
		jailValidator = false // Don't jail for downtime

	default:
		return nil, fmt.Errorf("unsupported slash reason: %s", reason.String())
	}

	// 5. Calculate slash amount from validator's stake
	slashAmount, err := k.slashValidator(ctx, validatorAddress, slashFraction, ctx.BlockHeight())
	if err != nil {
		return nil, fmt.Errorf("failed to slash validator: %w", err)
	}

	// 6. Jail validator if required
	if jailValidator {
		if err := k.jailValidator(ctx, validatorAddress); err != nil {
			return nil, fmt.Errorf("failed to jail validator: %w", err)
		}
	}

	// 7. Create slashing event record
	eventID := fmt.Sprintf("slash-%s-%d", validatorAddress, ctx.BlockHeight())
	slashingEvent := &types.SlashingEvent{
		EventId:          eventID,
		ValidatorAddress: validatorAddress,
		Reason:           reason,
		SlashAmount:      slashAmount,
		EvidenceHash:     evidenceHash,
		InfractionHeight: uint64(ctx.BlockHeight()),
		Timestamp:        ctx.BlockTime(),
		Jailed:           jailValidator,
	}

	// 8. Store slashing event
	k.setSlashingEvent(ctx, slashingEvent)

	// 9. Emit event for indexing and audit trail
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_slashed",
			sdk.NewAttribute("event_id", eventID),
			sdk.NewAttribute("validator", validatorAddress),
			sdk.NewAttribute("reason", reason.String()),
			sdk.NewAttribute("slash_amount", slashAmount.String()),
			sdk.NewAttribute("jailed", fmt.Sprintf("%t", jailValidator)),
			sdk.NewAttribute("transfer_id", transferId),
			sdk.NewAttribute("submitter", submitter),
		),
	)

	return slashingEvent, nil
}

// DetectDoubleSigning checks if a validator has signed conflicting messages for the same transfer.
//
// SECURITY CRITICAL: Double-signing is a severe offense that indicates:
//   - Validator running multiple instances (security risk)
//   - Validator attempting to defraud the bridge
//   - Compromised validator keys
//
// Double-signing detection compares signatures for the same transfer:
//   - Same transferId
//   - Same validator
//   - Different signature bytes
//
// Parameters:
//   - ctx: SDK context
//   - transferId: Transfer being checked
//   - newSignature: New signature being submitted
//   - validatorAddress: Validator submitting the signature
//
// Returns:
//   - bool: true if double-signing detected
//   - error: If validation fails
func (k Keeper) DetectDoubleSigning(
	ctx sdk.Context,
	transferId string,
	newSignature []byte,
	validatorAddress string,
) (bool, error) {
	if transferId == "" || validatorAddress == "" || len(newSignature) == 0 {
		return false, fmt.Errorf("invalid parameters for double-signing detection")
	}

	// Get existing signatures for this transfer
	transfer, found := k.getTransfer(ctx, transferId)
	if !found {
		// No transfer found - this is the first signature
		return false, nil
	}

	// Check if this validator has already signed this transfer
	for _, existingSig := range transfer.ValidatorSignatures {
		if existingSig.ValidatorAddress == validatorAddress {
			// Validator has signed before - check if signatures match
			if !bytes.Equal(existingSig.Signature, newSignature) {
				// DOUBLE-SIGNING DETECTED!
				// Different signatures for the same transfer from same validator
				return true, nil
			}
			// Same signature - this is a duplicate (replay), not double-signing
			return false, nil
		}
	}

	// Validator has not signed this transfer yet
	return false, nil
}

// slashValidator executes the actual slashing by calling the staking module.
//
// SECURITY CRITICAL: This function permanently reduces a validator's stake.
// The slashed amount cannot be recovered.
//
// Parameters:
//   - ctx: SDK context
//   - validatorAddress: Bridge validator address to slash
//   - slashFraction: Fraction of stake to slash (0.0 to 1.0)
//   - infractionHeight: Block height where infraction occurred
//
// Returns:
//   - sdkmath.Int: Amount slashed
//   - error: If slashing fails
func (k Keeper) slashValidator(
	ctx sdk.Context,
	validatorAddress string,
	slashFraction sdkmath.LegacyDec,
	infractionHeight int64,
) (sdkmath.Int, error) {
	// Validate inputs
	if slashFraction.IsNegative() || slashFraction.GT(sdkmath.LegacyOneDec()) {
		return sdkmath.ZeroInt(), fmt.Errorf("invalid slash fraction: %s (must be between 0 and 1)", slashFraction)
	}

	// If staking keeper is not available, record the slash but don't execute
	// (This happens in tests or if bridge validators are separate from consensus validators)
	if k.stakingKeeper == nil {
		// Return zero amount slashed
		return sdkmath.ZeroInt(), nil
	}

	// Convert validator address to ValAddress
	valAddr, err := sdk.ValAddressFromBech32(validatorAddress)
	if err != nil {
		return sdkmath.ZeroInt(), fmt.Errorf("invalid validator address: %w", err)
	}

	// Get validator from staking module
	validator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		// Validator not found in staking module
		// This is acceptable - bridge validators may not be consensus validators
		return sdkmath.ZeroInt(), nil
	}

	// Get consensus address for slashing
	consAddr, err := validator.GetConsAddr()
	if err != nil {
		return sdkmath.ZeroInt(), fmt.Errorf("failed to get consensus address: %w", err)
	}

	// Get validator's voting power at infraction height
	// For simplicity, use current power (production should use historical power)
	power := validator.ConsensusPower(sdk.DefaultPowerReduction)

	// Execute slash via staking module
	slashedAmount := k.stakingKeeper.Slash(
		ctx,
		consAddr,
		infractionHeight,
		power,
		slashFraction,
	)

	return slashedAmount, nil
}

// jailValidator jails a validator, removing them from the active set.
//
// SECURITY CRITICAL: Jailing prevents a malicious validator from continuing
// to participate in bridge operations until manually unjailed by governance.
//
// Parameters:
//   - ctx: SDK context
//   - validatorAddress: Bridge validator address to jail
//
// Returns:
//   - error: If jailing fails
func (k Keeper) jailValidator(ctx sdk.Context, validatorAddress string) error {
	// Mark validator as inactive in bridge module
	validator, found := k.getValidator(ctx, validatorAddress)
	if !found {
		return types.ErrValidatorNotFound
	}

	validator.Active = false
	k.setValidator(ctx, validator)

	// If staking keeper is available, also jail in staking module
	if k.stakingKeeper != nil {
		valAddr, err := sdk.ValAddressFromBech32(validatorAddress)
		if err != nil {
			return fmt.Errorf("invalid validator address: %w", err)
		}

		stakingValidator, found := k.stakingKeeper.GetValidator(ctx, valAddr)
		if found {
			consAddr, err := stakingValidator.GetConsAddr()
			if err != nil {
				return fmt.Errorf("failed to get consensus address: %w", err)
			}
			k.stakingKeeper.Jail(ctx, consAddr)
		}
	}

	return nil
}

// setSlashingEvent stores a slashing event
func (k Keeper) setSlashingEvent(ctx sdk.Context, event *types.SlashingEvent) {
	if event == nil || event.EventId == "" {
		return
	}
	store := k.store(ctx)
	key := types.SlashingEventKey(event.EventId)
	store.Set(key, k.cdc.MustMarshal(event))
}

// GetSlashingEvent retrieves a slashing event by ID
func (k Keeper) GetSlashingEvent(ctx sdk.Context, eventId string) (*types.SlashingEvent, bool) {
	if eventId == "" {
		return nil, false
	}
	store := k.store(ctx)
	key := types.SlashingEventKey(eventId)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var event types.SlashingEvent
	if err := k.cdc.Unmarshal(bz, &event); err != nil {
		return nil, false
	}
	return &event, true
}

// ========================================================================
// VALIDATOR LIVENESS TRACKING (for offline slashing)
// ========================================================================

// RecordValidatorSigning records that a validator signed a transfer at the current block.
//
// SECURITY: Liveness tracking ensures validators remain active and responsive.
// Validators who fail to sign sufficient transfers over a time window can be slashed.
//
// Parameters:
//   - ctx: SDK context
//   - validatorAddress: Validator who signed
//   - signed: Whether the validator signed (true) or missed (false)
func (k Keeper) RecordValidatorSigning(ctx sdk.Context, validatorAddress string, signed bool) {
	if validatorAddress == "" {
		return
	}

	params := k.GetParams(ctx)
	if params.MinSigningWindow == 0 {
		// Liveness tracking disabled
		return
	}

	blockHeight := ctx.BlockHeight()

	// Store signing info for this block
	key := types.ValidatorSigningInfoKey(validatorAddress, blockHeight)
	store := k.store(ctx)

	if signed {
		store.Set(key, []byte{1})
	} else {
		store.Set(key, []byte{0})
	}

	// Cleanup old signing info beyond the window
	oldestHeight := blockHeight - params.MinSigningWindow
	if oldestHeight > 0 {
		oldKey := types.ValidatorSigningInfoKey(validatorAddress, oldestHeight-1)
		store.Delete(oldKey)
	}
}

// CheckValidatorLiveness checks if a validator meets the minimum signing requirement.
//
// SECURITY CRITICAL: This function determines if a validator should be slashed
// for being offline. The validator must have signed at least MinSignedPerWindow
// percent of blocks in the last MinSigningWindow blocks.
//
// Parameters:
//   - ctx: SDK context
//   - validatorAddress: Validator to check
//
// Returns:
//   - bool: true if validator meets liveness requirements
//   - error: If validation fails
func (k Keeper) CheckValidatorLiveness(ctx sdk.Context, validatorAddress string) (bool, error) {
	if validatorAddress == "" {
		return false, fmt.Errorf("validator address cannot be empty")
	}

	params := k.GetParams(ctx)
	if params.MinSigningWindow == 0 {
		// Liveness tracking disabled - always pass
		return true, nil
	}

	minSignedPerWindow, err := sdkmath.LegacyNewDecFromStr(params.MinSignedPerWindow)
	if err != nil {
		return false, fmt.Errorf("invalid MinSignedPerWindow param: %w", err)
	}

	blockHeight := ctx.BlockHeight()
	store := k.store(ctx)

	// Count signatures in the window
	signedCount := int64(0)
	totalBlocks := params.MinSigningWindow

	// Check if we have enough blocks (might be early in chain life)
	if blockHeight < totalBlocks {
		totalBlocks = blockHeight
	}

	// Count how many blocks the validator signed
	for i := int64(0); i < totalBlocks; i++ {
		height := blockHeight - i
		if height <= 0 {
			break
		}

		key := types.ValidatorSigningInfoKey(validatorAddress, height)
		if bz := store.Get(key); bz != nil && len(bz) > 0 && bz[0] == 1 {
			signedCount++
		}
	}

	// Calculate percentage signed
	if totalBlocks == 0 {
		return true, nil // No blocks to check
	}

	signedFraction := sdkmath.LegacyNewDec(signedCount).Quo(sdkmath.LegacyNewDec(totalBlocks))

	// Check if meets minimum
	meetsRequirement := signedFraction.GTE(minSignedPerWindow)

	return meetsRequirement, nil
}

// SlashForDowntime slashes a validator for being offline.
//
// SECURITY: This is automatically called when a validator fails liveness checks.
// The slash amount is typically small (e.g., 1%) to encourage uptime without
// being overly punitive for temporary outages.
//
// Parameters:
//   - ctx: SDK context
//   - validatorAddress: Validator to slash
//
// Returns:
//   - error: If slashing fails
func (k Keeper) SlashForDowntime(ctx sdk.Context, validatorAddress string) error {
	// Check liveness first
	meetsRequirement, err := k.CheckValidatorLiveness(ctx, validatorAddress)
	if err != nil {
		return err
	}

	if meetsRequirement {
		// Validator is fine, no slashing needed
		return nil
	}

	// Validator failed liveness check - slash for downtime
	evidenceHash := []byte(fmt.Sprintf("downtime-slash-%s-%d", validatorAddress, ctx.BlockHeight()))

	_, err = k.SubmitSlashingEvidence(
		ctx,
		validatorAddress,
		types.SlashReason_SLASH_DOWNTIME,
		"", // No specific transfer ID for downtime
		evidenceHash,
		"system", // System-initiated slash
	)

	return err
}

// ========================================================================
// DOUBLE-SIGNING DETECTION INTEGRATION
// ========================================================================

// CheckAndSlashDoubleSigning is called when submitting validator signatures.
// It automatically detects and slashes double-signing.
//
// SECURITY CRITICAL: This should be called BEFORE accepting any validator signature.
// If double-signing is detected, the signature is rejected and the validator is slashed.
//
// Parameters:
//   - ctx: SDK context
//   - transferId: Transfer being signed
//   - signature: Signature being submitted
//   - validatorAddress: Validator submitting signature
//
// Returns:
//   - bool: true if signature is valid (not a double-sign)
//   - error: If validation fails or slashing fails
func (k Keeper) CheckAndSlashDoubleSigning(
	ctx sdk.Context,
	transferId string,
	signature []byte,
	validatorAddress string,
) (bool, error) {
	// Detect double-signing
	isDoubleSigning, err := k.DetectDoubleSigning(ctx, transferId, signature, validatorAddress)
	if err != nil {
		return false, fmt.Errorf("failed to detect double-signing: %w", err)
	}

	if !isDoubleSigning {
		// No double-signing detected - signature is valid
		return true, nil
	}

	// DOUBLE-SIGNING DETECTED - Slash the validator
	evidenceHash := []byte(fmt.Sprintf("double-sign-%s-%s", transferId, validatorAddress))

	_, err = k.SubmitSlashingEvidence(
		ctx,
		validatorAddress,
		types.SlashReason_SLASH_DOUBLE_SIGN,
		transferId,
		evidenceHash,
		"system", // System-detected double-signing
	)

	if err != nil {
		return false, fmt.Errorf("failed to slash for double-signing: %w", err)
	}

	// Reject the signature
	return false, fmt.Errorf("validator %s double-signed transfer %s - signature rejected and validator slashed",
		validatorAddress, transferId)
}

// GetValidatorSigningInfo retrieves signing history for a validator
func (k Keeper) GetValidatorSigningInfo(ctx sdk.Context, validatorAddress string) map[int64]bool {
	if validatorAddress == "" {
		return nil
	}

	params := k.GetParams(ctx)
	if params.MinSigningWindow == 0 {
		return nil
	}

	blockHeight := ctx.BlockHeight()
	store := k.store(ctx)
	signingInfo := make(map[int64]bool)

	// Retrieve signing info for the window
	for i := int64(0); i < params.MinSigningWindow; i++ {
		height := blockHeight - i
		if height <= 0 {
			break
		}

		key := types.ValidatorSigningInfoKey(validatorAddress, height)
		if bz := store.Get(key); bz != nil && len(bz) > 0 {
			signingInfo[height] = bz[0] == 1
		}
	}

	return signingInfo
}

// ========================================================================
// FRAUD PROOF INTEGRATION
// ========================================================================

// slashValidatorsForFraudulentTransfer slashes all validators who signed a fraudulent transfer.
//
// SECURITY CRITICAL: This is called when a fraud proof is validated as correct.
// All validators who attested to the fraudulent transfer must be punished economically.
//
// This function:
//   1. Retrieves the transfer and all validator signatures
//   2. For each validator who signed, submits slashing evidence
//   3. Slashes the validator's stake and jails them
//   4. Records slashing events for audit trail
//
// Parameters:
//   - ctx: SDK context
//   - transferID: The fraudulent transfer ID
//   - fraudProofID: The fraud proof ID that proved the transfer fraudulent
//
// Returns:
//   - error: If slashing fails for all validators (partial failures are logged but don't cause error)
func (k Keeper) slashValidatorsForFraudulentTransfer(ctx sdk.Context, transferID string, fraudProofID string) error {
	// Get the transfer
	transfer, found := k.getTransfer(ctx, transferID)
	if !found {
		return fmt.Errorf("transfer not found: %s", transferID)
	}

	// Check if transfer has validator signatures
	if len(transfer.ValidatorSignatures) == 0 {
		return fmt.Errorf("no validator signatures found on fraudulent transfer %s", transferID)
	}

	// Track slashing results
	slashedCount := 0
	var lastError error

	// Slash each validator who signed this fraudulent transfer
	for _, sig := range transfer.ValidatorSignatures {
		if sig.ValidatorAddress == "" {
			continue
		}

		// Create evidence hash (combines transfer ID, fraud proof ID, and validator address)
		evidenceData := []byte(fmt.Sprintf("fraud:%s:proof:%s:validator:%s",
			transferID, fraudProofID, sig.ValidatorAddress))

		// Submit slashing evidence for this validator
		_, err := k.SubmitSlashingEvidence(
			ctx,
			sig.ValidatorAddress,
			types.SlashReason_SLASH_FRAUD_ATTEMPT,
			transferID,
			evidenceData,
			fmt.Sprintf("fraud-proof:%s", fraudProofID),
		)

		if err != nil {
			ctx.Logger().Error("failed to slash validator for fraudulent transfer",
				"validator", sig.ValidatorAddress,
				"transfer_id", transferID,
				"fraud_proof_id", fraudProofID,
				"error", err)
			lastError = err
			// Continue to slash other validators even if one fails
			continue
		}

		slashedCount++
		ctx.Logger().Info("validator slashed for signing fraudulent transfer",
			"validator", sig.ValidatorAddress,
			"transfer_id", transferID,
			"fraud_proof_id", fraudProofID)
	}

	// If no validators were slashed, return error
	if slashedCount == 0 {
		if lastError != nil {
			return fmt.Errorf("failed to slash any validators: %w", lastError)
		}
		return fmt.Errorf("failed to slash any validators for fraudulent transfer %s", transferID)
	}

	ctx.Logger().Info("completed validator slashing for fraudulent transfer",
		"transfer_id", transferID,
		"fraud_proof_id", fraudProofID,
		"slashed_count", slashedCount,
		"total_signers", len(transfer.ValidatorSignatures))

	return nil
}

// GetAllSlashingEvents retrieves all slashing events from state.
//
// Returns:
//   - []*SlashingEvent: List of all slashing events
func (k Keeper) GetAllSlashingEvents(ctx sdk.Context) []*types.SlashingEvent {
	store := k.store(ctx)
	iterator := store.Iterator(types.SlashingEventPrefix, storetypes.PrefixEndBytes(types.SlashingEventPrefix))
	defer iterator.Close()

	var events []*types.SlashingEvent
	for ; iterator.Valid(); iterator.Next() {
		var event types.SlashingEvent
		if err := k.cdc.Unmarshal(iterator.Value(), &event); err != nil {
			// Log corrupted data but continue iteration
			k.Logger(ctx).Error("failed to unmarshal slashing event",
				"key", hex.EncodeToString(iterator.Key()),
				"error", err.Error())
			continue
		}
		eventCopy := event
		events = append(events, &eventCopy)
	}
	return events
}

// GetValidatorSlashingHistory retrieves all slashing events for a specific validator.
//
// Parameters:
//   - ctx: SDK context
//   - validatorAddress: Validator address
//
// Returns:
//   - []*SlashingEvent: List of slashing events for this validator
func (k Keeper) GetValidatorSlashingHistory(ctx sdk.Context, validatorAddress string) []*types.SlashingEvent {
	allEvents := k.GetAllSlashingEvents(ctx)
	var validatorEvents []*types.SlashingEvent

	for _, event := range allEvents {
		if event != nil && event.ValidatorAddress == validatorAddress {
			validatorEvents = append(validatorEvents, event)
		}
	}

	return validatorEvents
}
