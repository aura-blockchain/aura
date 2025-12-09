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

// CreateMultiSigWallet creates a new multi-signature wallet
func (k Keeper) CreateMultiSigWallet(
	ctx context.Context,
	creator string,
	signers []string,
	threshold int32,
	signerWeights map[string]int32,
	weightThreshold int32,
	timeLock *gogotypes.Duration,
) (*wsproto.MultiSigWallet, error) {
	// Validate inputs
	if len(signers) == 0 {
		return nil, types.ErrInvalidMultiSigConfig
	}
	if threshold <= 0 || threshold > int32(len(signers)) {
		return nil, types.ErrInvalidThreshold
	}
	if creator == "" {
		return nil, types.ErrInvalidInput
	}

	// Validate weights if using weighted multisig
	if len(signerWeights) > 0 {
		totalWeight := int32(0)
		for _, weight := range signerWeights {
			if weight <= 0 {
				return nil, types.ErrInvalidWeights
			}
			totalWeight += weight
		}
		if weightThreshold <= 0 || weightThreshold > totalWeight {
			return nil, types.ErrInvalidWeights
		}
	}

	// Generate wallet ID
	walletID := k.generateMultiSigWalletID(creator, signers)

	// Check if wallet already exists
	if _, err := k.GetMultiSigWallet(ctx, walletID); err == nil {
		return nil, fmt.Errorf("multi-sig wallet already exists")
	}

	// Create multi-sig wallet
	now := blockTimeToGogoTimestamp(ctx)
	wallet := &wsproto.MultiSigWallet{
		WalletId:        walletID,
		Signers:         signers,
		Threshold:       threshold,
		TotalSigners:    int32(len(signers)),
		CreatedAt:       now,
		Creator:         creator,
		SignerWeights:   signerWeights,
		WeightThreshold: weightThreshold,
		TimeLocked:      timeLock != nil,
		PendingSigners:  []string{},
	}

	if timeLock != nil {
		unlockTime := determinism.GetBlockTime(ctx).Add(gogoDurationToTime(timeLock))
		wallet.UnlockTime = timeToGogoTimestamp(unlockTime)
	}

	// Store the wallet
	walletBytes := k.cdc.MustMarshal(wallet)
	if err := k.SetMultiSigWallet(ctx, walletID, walletBytes); err != nil {
		return nil, err
	}

	k.logger.Info("created multi-sig wallet",
		"wallet_id", walletID,
		"creator", creator,
		"signers", len(signers),
		"threshold", threshold,
	)

	return wallet, nil
}

// CreatePendingMultiSigTransaction creates a pending transaction for multi-sig wallet
func (k Keeper) CreatePendingMultiSigTransaction(
	ctx context.Context,
	walletID string,
	txData []byte,
	txType string,
	description string,
	expirationDuration time.Duration,
) (*wsproto.PendingMultiSigTransaction, error) {
	// Get wallet configuration
	walletBytes, err := k.GetMultiSigWallet(ctx, walletID)
	if err != nil {
		return nil, err
	}

	var wallet wsproto.MultiSigWallet
	if err := k.cdc.Unmarshal(walletBytes, &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal multi-sig wallet: %w", err)
	}

	// Check time lock
	if wallet.TimeLocked && determinism.GetBlockTime(ctx).Before(gogoTimestampToTime(wallet.UnlockTime)) {
		return nil, fmt.Errorf("wallet is time-locked until %s", gogoTimestampToTime(wallet.UnlockTime))
	}

	// Generate transaction ID
	txID := k.generateTxID(walletID, txData)

	// Create pending transaction
	now := blockTimeToGogoTimestamp(ctx)
	expiresAt := blockTimeWithOffsetToGogoTimestamp(ctx, expirationDuration)

	pendingTx := &wsproto.PendingMultiSigTransaction{
		TxId:          txID,
		WalletId:      walletID,
		TxData:        txData,
		Signatures:    []string{},
		SignedBy:      []string{},
		CurrentWeight: 0,
		CreatedAt:     now,
		ExpiresAt:     expiresAt,
		TxType:        txType,
		Description:   description,
	}

	// Store pending transaction
	txBytes := k.cdc.MustMarshal(pendingTx)
	if err := k.SetPendingMultiSigTx(ctx, txID, txBytes); err != nil {
		return nil, err
	}

	k.logger.Info("created pending multi-sig transaction",
		"tx_id", txID,
		"wallet_id", walletID,
		"type", txType,
	)

	return pendingTx, nil
}

// SignMultiSigTransaction adds a signature to a pending multi-sig transaction
func (k Keeper) SignMultiSigTransaction(
	ctx context.Context,
	txID string,
	signer string,
	signature []byte,
) (bool, error) {
	// Get pending transaction
	txBytes, err := k.GetPendingMultiSigTx(ctx, txID)
	if err != nil {
		return false, err
	}

	var pendingTx wsproto.PendingMultiSigTransaction
	if err := k.cdc.Unmarshal(txBytes, &pendingTx); err != nil {

		k.logger.Error("failed to unmarshal", "error", err)

	}

	// Check expiration
	if determinism.GetBlockTime(ctx).After(gogoTimestampToTime(pendingTx.ExpiresAt)) {
		return false, types.ErrMultiSigTxExpired
	}

	// Get wallet configuration
	walletBytes, err := k.GetMultiSigWallet(ctx, pendingTx.WalletId)
	if err != nil {
		return false, err
	}

	var wallet wsproto.MultiSigWallet
	if err := k.cdc.Unmarshal(walletBytes, &wallet); err != nil {
		return false, fmt.Errorf("failed to unmarshal multi-sig wallet: %w", err)
	}

	// Verify signer is authorized
	if !k.isAuthorizedSigner(signer, wallet.Signers) {
		return false, types.ErrInvalidSigner
	}

	// Check if already signed
	if k.hasAlreadySigned(signer, pendingTx.SignedBy) {
		return false, types.ErrSignatureExists
	}

	// Validate signature
	if err := k.validateMultiSigSignature(pendingTx.TxData, signature, signer); err != nil {
		return false, err
	}

	// Add signature
	pendingTx.Signatures = append(pendingTx.Signatures, hex.EncodeToString(signature))
	pendingTx.SignedBy = append(pendingTx.SignedBy, signer)

	// Update weight
	if len(wallet.SignerWeights) > 0 {
		weight := wallet.SignerWeights[signer]
		pendingTx.CurrentWeight += weight
	} else {
		pendingTx.CurrentWeight = int32(len(pendingTx.SignedBy))
	}

	// Store updated transaction
	updatedTxBytes := k.cdc.MustMarshal(&pendingTx)
	if err := k.SetPendingMultiSigTx(ctx, txID, updatedTxBytes); err != nil {
		return false, err
	}

	// Check if ready to execute
	readyToExecute := k.isReadyToExecute(&pendingTx, &wallet)

	k.logger.Info("signed multi-sig transaction",
		"tx_id", txID,
		"signer", signer,
		"signatures", len(pendingTx.SignedBy),
		"ready", readyToExecute,
	)

	return readyToExecute, nil
}

// ExecuteMultiSigTransaction executes a fully signed multi-sig transaction
func (k Keeper) ExecuteMultiSigTransaction(ctx context.Context, txID string) error {
	// Get pending transaction
	txBytes, err := k.GetPendingMultiSigTx(ctx, txID)
	if err != nil {
		return err
	}

	var pendingTx wsproto.PendingMultiSigTransaction
	if err := k.cdc.Unmarshal(txBytes, &pendingTx); err != nil {
		return fmt.Errorf("failed to unmarshal pending multi-sig transaction: %w", err)
	}

	// Check expiration
	if determinism.GetBlockTime(ctx).After(gogoTimestampToTime(pendingTx.ExpiresAt)) {
		return types.ErrMultiSigTxExpired
	}

	// Get wallet configuration
	walletBytes, err := k.GetMultiSigWallet(ctx, pendingTx.WalletId)
	if err != nil {
		return err
	}

	var wallet wsproto.MultiSigWallet
	if err := k.cdc.Unmarshal(walletBytes, &wallet); err != nil {
		return fmt.Errorf("failed to unmarshal multi-sig wallet: %w", err)
	}

	// Verify ready to execute
	if !k.isReadyToExecute(&pendingTx, &wallet) {
		return types.ErrInsufficientSignatures
	}

	// In production, execute the actual transaction here
	// This would involve:
	// 1. Decoding the transaction data
	// 2. Executing the transaction
	// 3. Broadcasting to the network

	// Remove pending transaction
	if err := k.DeletePendingMultiSigTx(ctx, txID); err != nil {
		return err
	}

	k.logger.Info("executed multi-sig transaction",
		"tx_id", txID,
		"wallet_id", pendingTx.WalletId,
	)

	return nil
}

// isAuthorizedSigner checks if an address is an authorized signer
func (k Keeper) isAuthorizedSigner(address string, signers []string) bool {
	for _, signer := range signers {
		if signer == address {
			return true
		}
	}
	return false
}

// hasAlreadySigned checks if a signer has already signed
func (k Keeper) hasAlreadySigned(address string, signedBy []string) bool {
	for _, signer := range signedBy {
		if signer == address {
			return true
		}
	}
	return false
}

// isReadyToExecute checks if a transaction has enough signatures
func (k Keeper) isReadyToExecute(tx *wsproto.PendingMultiSigTransaction, wallet *wsproto.MultiSigWallet) bool {
	if len(wallet.SignerWeights) > 0 {
		// Weighted multisig
		return tx.CurrentWeight >= wallet.WeightThreshold
	}
	// Standard multisig
	return int32(len(tx.SignedBy)) >= wallet.Threshold
}

// validateMultiSigSignature validates a multi-sig transaction signature
func (k Keeper) validateMultiSigSignature(txData, signature []byte, signer string) error {
	// In production, this would:
	// 1. Verify the signature using the signer's public key
	// 2. Check that the signature is over the transaction hash
	// 3. Validate the signature format

	if len(signature) < 64 {
		return types.ErrInvalidDeviceSignature
	}

	return nil
}

// generateMultiSigWalletID generates a unique wallet ID for a multi-sig wallet
func (k Keeper) generateMultiSigWalletID(creator string, signers []string) string {
	data := creator
	for _, signer := range signers {
		data += ":" + signer
	}
	hash := sha256.Sum256([]byte(data))
	return "multisig_" + hex.EncodeToString(hash[:16])
}

// generateTxID generates a unique transaction ID
func (k Keeper) generateTxID(walletID string, txData []byte) string {
	data := append([]byte(walletID), txData...)
	hash := sha256.Sum256(data)
	return "tx_" + hex.EncodeToString(hash[:16])
}

// AddSignerToMultiSigWallet adds a new signer to a multi-sig wallet
func (k Keeper) AddSignerToMultiSigWallet(
	ctx context.Context,
	walletID string,
	newSigner string,
	requester string,
) error {
	// Get wallet
	walletBytes, err := k.GetMultiSigWallet(ctx, walletID)
	if err != nil {
		return err
	}

	var wallet wsproto.MultiSigWallet
	if err := k.cdc.Unmarshal(walletBytes, &wallet); err != nil {
		return fmt.Errorf("failed to unmarshal multi-sig wallet: %w", err)
	}

	// Verify requester is authorized
	if !k.isAuthorizedSigner(requester, wallet.Signers) {
		return types.ErrInvalidSigner
	}

	// Check if signer already exists
	if k.isAuthorizedSigner(newSigner, wallet.Signers) {
		return fmt.Errorf("signer already exists")
	}

	// Add to pending signers (requires multi-sig approval)
	wallet.PendingSigners = append(wallet.PendingSigners, newSigner)

	// Store updated wallet
	updatedBytes := k.cdc.MustMarshal(&wallet)
	return k.SetMultiSigWallet(ctx, walletID, updatedBytes)
}

// RemoveSignerFromMultiSigWallet removes a signer from a multi-sig wallet
func (k Keeper) RemoveSignerFromMultiSigWallet(
	ctx context.Context,
	walletID string,
	signerToRemove string,
	requester string,
) error {
	// Get wallet
	walletBytes, err := k.GetMultiSigWallet(ctx, walletID)
	if err != nil {
		return err
	}

	var wallet wsproto.MultiSigWallet
	if err := k.cdc.Unmarshal(walletBytes, &wallet); err != nil {
		return fmt.Errorf("failed to unmarshal multi-sig wallet: %w", err)
	}

	// Verify requester is authorized
	if !k.isAuthorizedSigner(requester, wallet.Signers) {
		return types.ErrInvalidSigner
	}

	// Remove signer
	newSigners := make([]string, 0, len(wallet.Signers)-1)
	for _, signer := range wallet.Signers {
		if signer != signerToRemove {
			newSigners = append(newSigners, signer)
		}
	}

	// Ensure threshold is still valid
	if wallet.Threshold > int32(len(newSigners)) {
		return types.ErrInvalidThreshold
	}

	wallet.Signers = newSigners
	wallet.TotalSigners = int32(len(newSigners))

	// Store updated wallet
	updatedBytes := k.cdc.MustMarshal(&wallet)
	return k.SetMultiSigWallet(ctx, walletID, updatedBytes)
}
