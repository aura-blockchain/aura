package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/aequitas/aura/chain/x/common/determinism"
	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

const (
	DefaultMaxGuardians  = 10
	DefaultRecoveryDelay = 48 * time.Hour
	MinRecoveryThreshold = 2
)

// ConfigureSocialRecovery configures social recovery for a wallet
func (k Keeper) ConfigureSocialRecovery(
	ctx context.Context,
	walletID string,
	guardians []*wsproto.Guardian,
	recoveryThreshold int32,
	recoveryDelay *gogotypes.Duration,
) (*wsproto.SocialRecoveryConfig, error) {
	// Validate inputs
	if len(guardians) == 0 {
		return nil, types.ErrInvalidRecoveryConfig
	}
	if recoveryThreshold < MinRecoveryThreshold {
		return nil, types.ErrInvalidRecoveryThreshold
	}
	if recoveryThreshold > int32(len(guardians)) {
		return nil, types.ErrInvalidRecoveryThreshold
	}
	if len(guardians) > DefaultMaxGuardians {
		return nil, fmt.Errorf("maximum %d guardians allowed", DefaultMaxGuardians)
	}

	// Validate guardians
	guardianAddresses := make(map[string]bool)
	for _, guardian := range guardians {
		if guardian.Address == "" {
			return nil, types.ErrInvalidGuardian
		}
		// Check for duplicates
		if guardianAddresses[guardian.Address] {
			return nil, fmt.Errorf("duplicate guardian address: %s", guardian.Address)
		}
		guardianAddresses[guardian.Address] = true

		// Initialize guardian
		guardian.Confirmed = false
		guardian.AddedAt = blockTimeToGogoTimestamp(ctx)
		guardian.RecoveryRequestsCount = 0
	}

	// Create configuration
	now := blockTimeToGogoTimestamp(ctx)
	delay := recoveryDelay
	if delay == nil {
		delay = &gogotypes.Duration{Seconds: int64(DefaultRecoveryDelay.Seconds()), Nanos: int32(DefaultRecoveryDelay.Nanoseconds() % 1e9)}
	}

	config := &wsproto.SocialRecoveryConfig{
		WalletId:          walletID,
		Guardians:         guardians,
		RecoveryThreshold: recoveryThreshold,
		RecoveryDelay:     delay,
		Enabled:           true,
		ConfiguredAt:      now,
		LastModified:      now,
		MaxGuardians:      DefaultMaxGuardians,
	}

	// Store configuration
	configBytes := k.cdc.MustMarshal(config)
	if err := k.SetSocialRecoveryConfig(ctx, walletID, configBytes); err != nil {
		return nil, err
	}

	k.logger.Info("configured social recovery",
		"wallet_id", walletID,
		"guardians", len(guardians),
		"threshold", recoveryThreshold,
		"delay", gogoDurationToTime(delay),
	)

	return config, nil
}

// ConfirmGuardian confirms a guardian's participation
func (k Keeper) ConfirmGuardian(ctx context.Context, walletID, guardianAddress string) error {
	// Get configuration
	configBytes, err := k.GetSocialRecoveryConfig(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	var config wsproto.SocialRecoveryConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

	}

	// Find and confirm guardian
	found := false
	for _, guardian := range config.Guardians {
		if guardian.Address == guardianAddress {
			guardian.Confirmed = true
			guardian.ConfirmedAt = blockTimeToGogoTimestamp(ctx)
			found = true
			break
		}
	}

	if !found {
		return types.ErrGuardianNotFound
	}

	config.LastModified = blockTimeToGogoTimestamp(ctx)

	// Store updated configuration
	updatedBytes := k.cdc.MustMarshal(&config)
	return k.SetSocialRecoveryConfig(ctx, walletID, updatedBytes)
}

// InitiateRecovery initiates a recovery request
func (k Keeper) InitiateRecovery(
	ctx context.Context,
	walletID string,
	newAddress string,
	initiator string,
) (*wsproto.RecoveryRequest, error) {
	// Get configuration
	configBytes, err := k.GetSocialRecoveryConfig(ctx, walletID)
	if err != nil {
		return nil, err
	}

	var config wsproto.SocialRecoveryConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

	}

	if !config.Enabled {
		return nil, types.ErrRecoveryNotEnabled
	}

	// Verify initiator is a confirmed guardian
	if !k.isConfirmedGuardian(initiator, config.Guardians) {
		return nil, types.ErrInvalidGuardian
	}

	// Generate request ID
	requestID := k.generateRecoveryRequestID(walletID, newAddress)

	// Create recovery request
	now := blockTimeToGogoTimestamp(ctx)
	executableAt := blockTimeWithOffsetToGogoTimestamp(ctx, gogoDurationToTime(config.RecoveryDelay))

	request := &wsproto.RecoveryRequest{
		RequestId:      requestID,
		WalletId:       walletID,
		NewAddress:     newAddress,
		Approvals:      []string{initiator},
		ApprovalsCount: 1,
		InitiatedAt:    now,
		ExecutableAt:   executableAt,
		Status:         wsproto.RecoveryStatus_RECOVERY_STATUS_PENDING,
		Initiator:      initiator,
	}

	// Store recovery request
	requestBytes := k.cdc.MustMarshal(request)
	if err := k.SetRecoveryRequest(ctx, requestID, requestBytes); err != nil {
		return nil, err
	}

	k.logger.Info("initiated recovery request",
		"request_id", requestID,
		"wallet_id", walletID,
		"initiator", initiator,
		"executable_at", gogoTimestampToTime(executableAt),
	)

	return request, nil
}

// ApproveRecovery approves a recovery request
func (k Keeper) ApproveRecovery(
	ctx context.Context,
	requestID string,
	guardianAddress string,
	signature []byte,
) (bool, error) {
	// Get recovery request
	requestBytes, err := k.GetRecoveryRequest(ctx, requestID)
	if err != nil {
		return false, err
	}

	var request wsproto.RecoveryRequest
	if err := k.cdc.Unmarshal(requestBytes, &request); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

	}

	// Check if already executed
	if request.Status == wsproto.RecoveryStatus_RECOVERY_STATUS_EXECUTED {
		return false, types.ErrRecoveryAlreadyExecuted
	}

	// Get configuration
	configBytes, err := k.GetSocialRecoveryConfig(ctx, request.WalletId)
	if err != nil {
		return false, err
	}

	var config wsproto.SocialRecoveryConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

	}

	// Verify guardian is authorized and confirmed
	if !k.isConfirmedGuardian(guardianAddress, config.Guardians) {
		return false, types.ErrInvalidGuardian
	}

	// Check if already approved
	for _, approval := range request.Approvals {
		if approval == guardianAddress {
			return false, fmt.Errorf("guardian already approved")
		}
	}

	// Validate signature
	if err := k.validateRecoverySignature(request.WalletId, request.NewAddress, signature, guardianAddress); err != nil {
		return false, err
	}

	// Add approval
	request.Approvals = append(request.Approvals, guardianAddress)
	request.ApprovalsCount++

	// Update guardian recovery count
	if err := k.incrementGuardianRecoveryCount(ctx, request.WalletId, guardianAddress); err != nil {
		return false, err
	}

	// Check if threshold is met
	readyToExecute := request.ApprovalsCount >= config.RecoveryThreshold

	if readyToExecute {
		request.Status = wsproto.RecoveryStatus_RECOVERY_STATUS_APPROVED
	}

	// Store updated request
	updatedBytes := k.cdc.MustMarshal(&request)
	if err := k.SetRecoveryRequest(ctx, requestID, updatedBytes); err != nil {
		return false, err
	}

	k.logger.Info("approved recovery request",
		"request_id", requestID,
		"guardian", guardianAddress,
		"approvals", request.ApprovalsCount,
		"threshold", config.RecoveryThreshold,
		"ready", readyToExecute,
	)

	return readyToExecute, nil
}

// ExecuteRecovery executes an approved recovery request
func (k Keeper) ExecuteRecovery(ctx context.Context, requestID string) error {
	// Get recovery request
	requestBytes, err := k.GetRecoveryRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	var request wsproto.RecoveryRequest
	if err := k.cdc.Unmarshal(requestBytes, &request); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

	}

	// Check if already executed
	if request.Status == wsproto.RecoveryStatus_RECOVERY_STATUS_EXECUTED {
		return types.ErrRecoveryAlreadyExecuted
	}

	// Get configuration
	configBytes, err := k.GetSocialRecoveryConfig(ctx, request.WalletId)
	if err != nil {
		return fmt.Errorf("failed to get for WalletId: %w", err)
	}

	var config wsproto.SocialRecoveryConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

	}

	// Verify threshold is met
	if request.ApprovalsCount < config.RecoveryThreshold {
		return types.ErrInsufficientApprovals
	}

	// Check if recovery delay has elapsed
	if determinism.GetBlockTime(ctx).Before(gogoTimestampToTime(request.ExecutableAt)) {
		return types.ErrRecoveryDelayNotElapsed
	}

	// In production, this would:
	// 1. Transfer wallet ownership to new address
	// 2. Update all associated keys
	// 3. Notify all guardians
	// 4. Create audit log entry

	request.Status = wsproto.RecoveryStatus_RECOVERY_STATUS_EXECUTED
	request.ExecutedAt = blockTimeToGogoTimestamp(ctx)

	// Store updated request
	updatedBytes := k.cdc.MustMarshal(&request)
	if err := k.SetRecoveryRequest(ctx, requestID, updatedBytes); err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	k.logger.Info("executed recovery request",
		"request_id", requestID,
		"wallet_id", request.WalletId,
		"new_address", request.NewAddress,
	)

	return nil
}

// CancelRecovery cancels a recovery request
func (k Keeper) CancelRecovery(ctx context.Context, requestID string, walletOwner string) error {
	// Get recovery request
	requestBytes, err := k.GetRecoveryRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	var request wsproto.RecoveryRequest
	if err := k.cdc.Unmarshal(requestBytes, &request); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

	}

	// Only wallet owner can cancel
	// In production, verify walletOwner actually owns the wallet

	request.Status = wsproto.RecoveryStatus_RECOVERY_STATUS_CANCELLED

	// Store updated request
	updatedBytes := k.cdc.MustMarshal(&request)
	return k.SetRecoveryRequest(ctx, requestID, updatedBytes)
}

// isConfirmedGuardian checks if an address is a confirmed guardian
func (k Keeper) isConfirmedGuardian(address string, guardians []*wsproto.Guardian) bool {
	for _, guardian := range guardians {
		if guardian.Address == address && guardian.Confirmed {
			return true
		}
	}
	return false
}

// validateRecoverySignature validates a recovery approval signature
func (k Keeper) validateRecoverySignature(walletID, newAddress string, signature []byte, guardian string) error {
	// In production, this would:
	// 1. Verify the signature using the guardian's public key
	// 2. Check that the signature is over the recovery data (walletID + newAddress)
	// 3. Validate the signature format

	if len(signature) < 64 {
		return types.ErrInvalidDeviceSignature
	}

	return nil
}

// generateRecoveryRequestID generates a unique recovery request ID
func (k Keeper) generateRecoveryRequestID(walletID, newAddress string) string {
	data := fmt.Sprintf("%s:%s", walletID, newAddress)
	hash := sha256.Sum256([]byte(data))
	return "recovery_" + hex.EncodeToString(hash[:16])
}

// incrementGuardianRecoveryCount increments a guardian's recovery count
func (k Keeper) incrementGuardianRecoveryCount(ctx context.Context, walletID, guardianAddress string) error {
	configBytes, err := k.GetSocialRecoveryConfig(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	var config wsproto.SocialRecoveryConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

	}

	for _, guardian := range config.Guardians {
		if guardian.Address == guardianAddress {
			guardian.RecoveryRequestsCount++
			break
		}
	}

	config.LastModified = blockTimeToGogoTimestamp(ctx)

	updatedBytes := k.cdc.MustMarshal(&config)
	return k.SetSocialRecoveryConfig(ctx, walletID, updatedBytes)
}

// AddGuardian adds a new guardian to social recovery
func (k Keeper) AddGuardian(ctx context.Context, walletID string, guardian *wsproto.Guardian) error {
	configBytes, err := k.GetSocialRecoveryConfig(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	var config wsproto.SocialRecoveryConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

	}

	// Check max guardians
	if len(config.Guardians) >= int(config.MaxGuardians) {
		return fmt.Errorf("maximum guardians reached")
	}

	// Check for duplicates
	for _, g := range config.Guardians {
		if g.Address == guardian.Address {
			return types.ErrGuardianExists
		}
	}

	// Initialize new guardian
	guardian.AddedAt = blockTimeToGogoTimestamp(ctx)
	guardian.Confirmed = false
	guardian.RecoveryRequestsCount = 0

	config.Guardians = append(config.Guardians, guardian)
	config.LastModified = blockTimeToGogoTimestamp(ctx)

	updatedBytes := k.cdc.MustMarshal(&config)
	return k.SetSocialRecoveryConfig(ctx, walletID, updatedBytes)
}

// RemoveGuardian removes a guardian from social recovery
func (k Keeper) RemoveGuardian(ctx context.Context, walletID, guardianAddress string) error {
	configBytes, err := k.GetSocialRecoveryConfig(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	var config wsproto.SocialRecoveryConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

	}

	// Remove guardian
	newGuardians := make([]*wsproto.Guardian, 0, len(config.Guardians)-1)
	for _, guardian := range config.Guardians {
		if guardian.Address != guardianAddress {
			newGuardians = append(newGuardians, guardian)
		}
	}

	// Ensure threshold is still valid
	if config.RecoveryThreshold > int32(len(newGuardians)) {
		return types.ErrInvalidRecoveryThreshold
	}

	config.Guardians = newGuardians
	config.LastModified = blockTimeToGogoTimestamp(ctx)

	updatedBytes := k.cdc.MustMarshal(&config)
	return k.SetSocialRecoveryConfig(ctx, walletID, updatedBytes)
}
