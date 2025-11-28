package keeper

import (
	"context"
	"fmt"
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================
// ANTI-WHALE MECHANISMS (Feature 5)
// ============================

// CheckWhaleProtection validates a transfer against whale protection rules
// This function ensures that large token holders (whales) cannot manipulate
// the market through excessive transactions or holdings
func (k *Keeper) CheckWhaleProtection(ctx context.Context, sender, recipient, amount string) error {
	params := k.GetParams()

	if !params.WhaleProtection.Enabled {
		return nil
	}

	// Check if sender or recipient is exempted
	for _, exempted := range params.WhaleProtection.ExemptedAddresses {
		if sender == exempted || recipient == exempted {
			return nil
		}
	}

	transferAmt := new(big.Int)
	if _, ok := transferAmt.SetString(amount, 10); !ok {
		return types.ErrInvalidAmount
	}

	totalSupply := new(big.Int)
	if _, ok := totalSupply.SetString(params.Tokenomics.CirculatingSupply, 10); !ok {
		return types.ErrInvalidAmount
	}

	// Check transaction size limit
	maxTxAmount := new(big.Int).Mul(totalSupply, big.NewInt(int64(params.WhaleProtection.MaxTxPercentage)))
	maxTxAmount.Div(maxTxAmount, big.NewInt(types.BasisPoints))

	if transferAmt.Cmp(maxTxAmount) > 0 {
		return types.ErrWhaleTxLimitExceeded
	}

	// Check if this is a large transaction and cooldown applies
	largeTxThreshold := new(big.Int).Mul(totalSupply, big.NewInt(int64(params.WhaleProtection.LargeTxThreshold)))
	largeTxThreshold.Div(largeTxThreshold, big.NewInt(types.BasisPoints))

	if transferAmt.Cmp(largeTxThreshold) > 0 {
		// Check cooldown
		lastTxTime, err := k.GetLastLargeTxTime(ctx, sender)
		if err != nil {
			return err
		}

		currentTime, err := k.GetCurrentTime(ctx)
		if err != nil {
			return err
		}

		if lastTxTime > 0 {
			timeSinceLastTx := currentTime - lastTxTime
			if uint64(timeSinceLastTx) < params.WhaleProtection.LargeTxCooldown {
				return types.ErrLargeTxCooldownActive
			}
		}

		// Record this large transaction
		if err := k.recordLargeTx(ctx, sender, recipient, amount, transferAmt, totalSupply); err != nil {
			return err
		}

		// Update last large tx time
		if err := k.SetLastLargeTxTime(ctx, sender, currentTime); err != nil {
			return err
		}
	}

	// Check recipient holding limit after transfer
	recipientHoldingStr, err := k.GetAddressHolding(ctx, recipient)
	if err != nil {
		return err
	}

	recipientHolding := new(big.Int)
	if _, ok := recipientHolding.SetString(recipientHoldingStr, 10); !ok {
		recipientHolding = big.NewInt(0)
	}

	newHolding := new(big.Int).Add(recipientHolding, transferAmt)

	maxHolding := new(big.Int).Mul(totalSupply, big.NewInt(int64(params.WhaleProtection.MaxHoldingPercentage)))
	maxHolding.Div(maxHolding, big.NewInt(types.BasisPoints))

	if newHolding.Cmp(maxHolding) > 0 {
		return types.ErrWhaleHoldingLimitExceeded
	}

	return nil
}

// UpdateAddressHolding updates the holding amount for an address
// This tracks the current balance of each address for whale protection enforcement
func (k *Keeper) UpdateAddressHolding(ctx context.Context, address, newBalance string) error {
	// Validate the balance is a valid number
	balance := new(big.Int)
	if _, ok := balance.SetString(newBalance, 10); !ok {
		return types.ErrInvalidAmount
	}

	return k.SetAddressHolding(ctx, address, newBalance)
}

// recordLargeTx records a large transaction for monitoring and analytics
func (k *Keeper) recordLargeTx(ctx context.Context, sender, recipient, amount string, transferAmt, totalSupply *big.Int) error {
	// Calculate percentage
	percentage := new(big.Int).Mul(transferAmt, big.NewInt(types.BasisPoints))
	percentage.Div(percentage, totalSupply)

	currentHeight, err := k.GetCurrentHeight(ctx)
	if err != nil {
		return err
	}

	// Create a safe sender prefix (handle short addresses)
	senderPrefix := sender
	if len(sender) > 8 {
		senderPrefix = sender[:8]
	}

	record := &types.LargeTxRecord{
		TxHash:             fmt.Sprintf("tx_%d_%s", currentHeight, senderPrefix),
		Sender:             sender,
		Recipient:          recipient,
		Amount:             amount,
		PercentageOfSupply: percentage.Uint64(),
		BlockHeight:        currentHeight,
		Timestamp:          timestamppb.Now(),
		Flagged:            false,
	}

	// Flag if exceeds 0.5% of supply (50 basis points)
	if percentage.Uint64() > 50 {
		record.Flagged = true
	}

	// Store the record
	return k.SetLargeTxRecord(ctx, record)
}

// GetLargeTxRecords returns recent large transaction records
// Limited to prevent excessive memory usage in queries
func (k *Keeper) GetLargeTxRecords(ctx context.Context, limit uint64) ([]*types.LargeTxRecord, error) {
	if limit == 0 {
		limit = 100 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Maximum limit
	}

	records := []*types.LargeTxRecord{}
	count := uint64(0)

	err := k.IterateLargeTxRecords(ctx, func(record *types.LargeTxRecord) bool {
		if count >= limit {
			return false
		}
		records = append(records, record)
		count++
		return true
	})

	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetWhaleProtectionTriggers24h returns count of whale protection triggers in last 24h
// This is used for monitoring and alerting on potential whale activity
func (k *Keeper) GetWhaleProtectionTriggers24h(ctx context.Context) (uint64, error) {
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return 0, err
	}

	cutoff := currentTime - 86400 // 24 hours ago
	count := uint64(0)

	err = k.IterateLargeTxRecords(ctx, func(record *types.LargeTxRecord) bool {
		if record.Timestamp.Seconds >= cutoff && record.Flagged {
			count++
		}
		return true
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetWhaleProtectionStatistics returns comprehensive whale protection statistics
// This provides insights into large holder behavior and protection effectiveness
func (k *Keeper) GetWhaleProtectionStatistics(ctx context.Context) (totalLargeTx uint64, flaggedTx uint64, avgPercentage uint64, err error) {
	totalLargeTx = 0
	flaggedTx = 0
	totalPercentage := uint64(0)

	err = k.IterateLargeTxRecords(ctx, func(record *types.LargeTxRecord) bool {
		totalLargeTx++
		totalPercentage += record.PercentageOfSupply
		if record.Flagged {
			flaggedTx++
		}
		return true
	})

	if err != nil {
		return 0, 0, 0, err
	}

	if totalLargeTx > 0 {
		avgPercentage = totalPercentage / totalLargeTx
	}

	return totalLargeTx, flaggedTx, avgPercentage, nil
}

// IsWhaleProtectionActive checks if whale protection is currently active
func (k *Keeper) IsWhaleProtectionActive() bool {
	params := k.GetParams()
	return params.WhaleProtection.Enabled
}

// GetWhaleHoldingPercentage calculates what percentage of total supply an address holds
func (k *Keeper) GetWhaleHoldingPercentage(ctx context.Context, address string) (uint64, error) {
	params := k.GetParams()

	holdingStr, err := k.GetAddressHolding(ctx, address)
	if err != nil {
		return 0, err
	}

	holding := new(big.Int)
	if _, ok := holding.SetString(holdingStr, 10); !ok {
		return 0, nil
	}

	totalSupply := new(big.Int)
	if _, ok := totalSupply.SetString(params.Tokenomics.CirculatingSupply, 10); !ok {
		return 0, types.ErrInvalidAmount
	}

	// Calculate percentage in basis points
	percentage := new(big.Int).Mul(holding, big.NewInt(types.BasisPoints))
	percentage.Div(percentage, totalSupply)

	return percentage.Uint64(), nil
}
