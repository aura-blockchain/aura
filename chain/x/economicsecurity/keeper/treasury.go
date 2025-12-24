package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// TREASURY MULTISIG (Feature 10)
// ============================

// ProposeTreasurySpend creates a proposal for treasury spending
func (k *Keeper) ProposeTreasurySpend(
	ctx context.Context,
	proposer, recipient, amount, description string,
) (txID string, executableAt time.Time, err error) {
	params, _ := k.GetParams(ctx)

	if params.TreasuryMultisig.TreasuryAddress == "" {
		return "", time.Time{}, types.ErrInvalidTreasuryAddress
	}

	// Validate proposer is a signer
	isSigner := false
	for _, signer := range params.TreasuryMultisig.Signers {
		if signer == proposer {
			isSigner = true
			break
		}
	}
	if !isSigner {
		return "", time.Time{}, types.ErrInvalidSigner
	}

	// Validate amount
	transferAmt := new(big.Int)
	if _, ok := transferAmt.SetString(amount, 10); !ok {
		return "", time.Time{}, types.ErrInvalidAmount
	}

	// Get current time
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return "", time.Time{}, err
	}

	// Generate transaction ID
	txID = k.generateTreasuryTxID(proposer, recipient, amount, currentTime)

	// Calculate timelock
	executableAtTime := time.Unix(currentTime+int64(params.TreasuryMultisig.TimelockDuration), 0)

	tx := &types.PendingTreasuryTx{
		TxId:         txID,
		Recipient:    recipient,
		Amount:       amount,
		Description:  description,
		Proposer:     proposer,
		Signatures:   []string{proposer}, // Proposer automatically signs
		CreatedAt:    time.Unix(currentTime, 0),
		ExecutableAt: executableAtTime,
		Executed:     false,
		Rejected:     false,
	}

	// Store the transaction
	if err := k.SetPendingTreasuryTx(ctx, tx); err != nil {
		return "", time.Time{}, err
	}

	return txID, tx.ExecutableAt, nil
}

// SignTreasurySpend signs a pending treasury transaction
func (k *Keeper) SignTreasurySpend(
	ctx context.Context,
	signer, txID string,
) (currentSignatures uint32, requiredThreshold uint32, err error) {
	params, _ := k.GetParams(ctx)

	// Validate signer
	isSigner := false
	for _, s := range params.TreasuryMultisig.Signers {
		if s == signer {
			isSigner = true
			break
		}
	}
	if !isSigner {
		return 0, 0, types.ErrInvalidSigner
	}

	// Get the transaction
	tx, err := k.GetPendingTreasuryTx(ctx, txID)
	if err != nil {
		return 0, 0, err
	}

	if tx.Executed {
		return 0, 0, types.ErrTxAlreadyExecuted
	}

	if tx.Rejected {
		return 0, 0, types.ErrTxAlreadyRejected
	}

	// Check if already signed
	for _, sig := range tx.Signatures {
		if sig == signer {
			return 0, 0, types.ErrAlreadySigned
		}
	}

	// Add signature
	tx.Signatures = append(tx.Signatures, signer)

	// Update the transaction in store
	if err := k.SetPendingTreasuryTx(ctx, tx); err != nil {
		return 0, 0, err
	}

	return uint32(len(tx.Signatures)), params.TreasuryMultisig.Threshold, nil
}

// ExecuteTreasurySpend executes an approved treasury transaction
func (k *Keeper) ExecuteTreasurySpend(
	ctx context.Context,
	executor, txID string,
	treasuryBalance string,
) error {
	params, _ := k.GetParams(ctx)

	// Get the transaction
	tx, err := k.GetPendingTreasuryTx(ctx, txID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	if tx.Executed {
		return types.ErrTxAlreadyExecuted
	}

	if tx.Rejected {
		return types.ErrTxAlreadyRejected
	}

	// Check if sufficient signatures
	if uint32(len(tx.Signatures)) < params.TreasuryMultisig.Threshold {
		return types.ErrInsufficientSignatures
	}

	// Check if timelock has passed
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	executableAt := tx.ExecutableAt.Unix()
	if currentTime < executableAt {
		return types.ErrTimelockNotExpired
	}

	// Check if treasury has sufficient balance
	balance := new(big.Int)
	if treasuryBalance != "" && treasuryBalance != "0" {
		balance.SetString(treasuryBalance, 10)
	}

	amount := new(big.Int)
	amount.SetString(tx.Amount, 10)

	if balance.Cmp(amount) < 0 {
		return types.ErrInsufficientTreasuryBalance
	}

	// Mark as executed
	tx.Executed = true

	// Update the transaction in store
	return k.SetPendingTreasuryTx(ctx, tx)
}

// GetPendingTreasuryTxByID retrieves a pending treasury transaction
func (k *Keeper) GetPendingTreasuryTxByID(ctx context.Context, txID string) (*types.PendingTreasuryTx, bool) {
	tx, err := k.GetPendingTreasuryTx(ctx, txID)
	if err != nil {
		return nil, false
	}
	return tx, true
}

// GetAllPendingTreasuryTxs returns all pending treasury transactions
func (k *Keeper) GetAllPendingTreasuryTxs(ctx context.Context) ([]*types.PendingTreasuryTx, error) {
	txs := make([]*types.PendingTreasuryTx, 0)

	err := k.IteratePendingTreasuryTxs(ctx, func(tx *types.PendingTreasuryTx) bool {
		if !tx.Executed && !tx.Rejected {
			txs = append(txs, tx)
		}
		return true
	})

	return txs, err
}

// RejectTreasurySpend rejects a pending treasury transaction
func (k *Keeper) RejectTreasurySpend(ctx context.Context, txID string) error {
	tx, err := k.GetPendingTreasuryTx(ctx, txID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	if tx.Executed {
		return types.ErrTxAlreadyExecuted
	}

	tx.Rejected = true

	return k.SetPendingTreasuryTx(ctx, tx)
}

// generateTreasuryTxID generates a unique transaction ID
func (k *Keeper) generateTreasuryTxID(proposer, recipient, amount string, timestamp int64) string {
	h := sha256.New()
	h.Write([]byte(proposer))
	h.Write([]byte(recipient))
	h.Write([]byte(amount))
	h.Write([]byte(fmt.Sprintf("%d", timestamp)))
	return "ttx:" + hex.EncodeToString(h.Sum(nil))[:32]
}

// GetTreasuryStatistics returns treasury multisig statistics
func (k *Keeper) GetTreasuryStatistics(ctx context.Context) (
	enabled bool,
	treasuryAddress string,
	threshold uint32,
	signerCount uint32,
	pendingTxCount uint64,
	executedTxCount uint64,
) {
	params, _ := k.GetParams(ctx)

	if params.TreasuryMultisig.TreasuryAddress == "" {
		return false, "", 0, 0, 0, 0
	}

	pendingCount := uint64(0)
	executedCount := uint64(0)

	if err := k.IteratePendingTreasuryTxs(ctx, func(tx *types.PendingTreasuryTx) bool {
		if tx.Executed {
			executedCount++
		} else if !tx.Rejected {
			pendingCount++
		}
		return true
	}); err != nil {
		return false, "", 0, 0, 0, 0
	}

	return true,
		params.TreasuryMultisig.TreasuryAddress,
		params.TreasuryMultisig.Threshold,
		uint32(len(params.TreasuryMultisig.Signers)),
		pendingCount,
		executedCount
}

// CleanupExecutedTreasuryTxs removes old executed treasury transactions
func (k *Keeper) CleanupExecutedTreasuryTxs(ctx context.Context, retentionPeriod int64) error {
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	txsToDelete := make([]string, 0)

	err = k.IteratePendingTreasuryTxs(ctx, func(tx *types.PendingTreasuryTx) bool {
		if tx.Executed || tx.Rejected {
			createdAt := tx.CreatedAt.Unix()
			if currentTime-createdAt > retentionPeriod {
				txsToDelete = append(txsToDelete, tx.TxId)
			}
		}
		return true
	})

	if err != nil {
		return fmt.Errorf("error in CleanupExecutedTreasuryTxs for TxId: %w", err)
	}

	// Delete old transactions
	for _, txID := range txsToDelete {
		if err := k.DeletePendingTreasuryTx(ctx, txID); err != nil {
			return fmt.Errorf("failed to delete: %w", err)
		}
	}

	return nil
}
