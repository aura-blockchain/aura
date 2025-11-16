package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================
// TREASURY MULTISIG (Feature 10)
// ============================

// ProposeTreasurySpend creates a proposal for treasury spending
func (k *Keeper) ProposeTreasurySpend(proposer, recipient, amount, description string) (string, *timestamppb.Timestamp, error) {
	params := k.GetParams()

	if params.TreasuryMultisig.TreasuryAddress == "" {
		return "", nil, types.ErrInvalidTreasuryAddress
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
		return "", nil, types.ErrInvalidSigner
	}

	transferAmt := new(big.Int)
	if _, ok := transferAmt.SetString(amount, 10); !ok {
		return "", nil, types.ErrInvalidAmount
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Generate transaction ID
	txID := k.generateTreasuryTxID(proposer, recipient, amount)

	// Calculate timelock
	executableAt := time.Unix(k.currentTime+int64(params.TreasuryMultisig.TimelockDuration), 0)

	tx := &types.PendingTreasuryTx{
		TxId:         txID,
		Recipient:    recipient,
		Amount:       amount,
		Description:  description,
		Proposer:     proposer,
		Signatures:   []string{proposer}, // Proposer automatically signs
		CreatedAt:    timestamppb.New(time.Unix(k.currentTime, 0)),
		ExecutableAt: timestamppb.New(executableAt),
		Executed:     false,
		Rejected:     false,
	}

	k.pendingTreasuryTxs[txID] = tx

	return txID, tx.ExecutableAt, nil
}

// SignTreasurySpend signs a pending treasury transaction
func (k *Keeper) SignTreasurySpend(signer, txID string) (uint32, uint32, error) {
	params := k.GetParams()

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

	k.mu.Lock()
	defer k.mu.Unlock()

	tx, ok := k.pendingTreasuryTxs[txID]
	if !ok {
		return 0, 0, types.ErrTxNotFound
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
	k.pendingTreasuryTxs[txID] = tx

	return uint32(len(tx.Signatures)), params.TreasuryMultisig.Threshold, nil
}

// ExecuteTreasurySpend executes an approved treasury transaction
func (k *Keeper) ExecuteTreasurySpend(executor, txID string, treasuryBalance string) error {
	params := k.GetParams()

	k.mu.Lock()
	defer k.mu.Unlock()

	tx, ok := k.pendingTreasuryTxs[txID]
	if !ok {
		return types.ErrTxNotFound
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
	executableAt := tx.ExecutableAt.AsTime().Unix()
	if k.currentTime < executableAt {
		return types.ErrTimelockNotExpired
	}

	// Check if treasury has sufficient balance
	balance := new(big.Int)
	balance.SetString(treasuryBalance, 10)

	amount := new(big.Int)
	amount.SetString(tx.Amount, 10)

	if balance.Cmp(amount) < 0 {
		return types.ErrInsufficientTreasuryBalance
	}

	// Mark as executed
	tx.Executed = true
	k.pendingTreasuryTxs[txID] = tx

	return nil
}

// GetPendingTreasuryTx retrieves a pending treasury transaction
func (k *Keeper) GetPendingTreasuryTx(txID string) (*types.PendingTreasuryTx, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	tx, ok := k.pendingTreasuryTxs[txID]
	return tx, ok
}

// GetAllPendingTreasuryTxs returns all pending treasury transactions
func (k *Keeper) GetAllPendingTreasuryTxs() []*types.PendingTreasuryTx {
	k.mu.RLock()
	defer k.mu.RUnlock()

	txs := make([]*types.PendingTreasuryTx, 0, len(k.pendingTreasuryTxs))
	for _, tx := range k.pendingTreasuryTxs {
		if !tx.Executed && !tx.Rejected {
			txs = append(txs, tx)
		}
	}

	return txs
}

// RejectTreasurySpend rejects a pending treasury transaction
func (k *Keeper) RejectTreasurySpend(txID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	tx, ok := k.pendingTreasuryTxs[txID]
	if !ok {
		return types.ErrTxNotFound
	}

	if tx.Executed {
		return types.ErrTxAlreadyExecuted
	}

	tx.Rejected = true
	k.pendingTreasuryTxs[txID] = tx

	return nil
}

func (k *Keeper) generateTreasuryTxID(proposer, recipient, amount string) string {
	h := sha256.New()
	h.Write([]byte(proposer))
	h.Write([]byte(recipient))
	h.Write([]byte(amount))
	h.Write([]byte(fmt.Sprintf("%d", k.currentTime)))
	return "ttx:" + hex.EncodeToString(h.Sum(nil))[:32]
}
