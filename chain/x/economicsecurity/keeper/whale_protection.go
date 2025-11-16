package keeper

import (
	"fmt"
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================
// ANTI-WHALE MECHANISMS (Feature 5)
// ============================

// CheckWhaleProtection validates a transfer against whale protection rules
func (k *Keeper) CheckWhaleProtection(sender, recipient, amount string) error {
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
	totalSupply.SetString(params.Tokenomics.CirculatingSupply, 10)

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
		k.mu.RLock()
		lastTxTime, exists := k.lastLargeTxTimes[sender]
		k.mu.RUnlock()

		if exists {
			timeSinceLastTx := k.currentTime - lastTxTime
			if uint64(timeSinceLastTx) < params.WhaleProtection.LargeTxCooldown {
				return types.ErrLargeTxCooldownActive
			}
		}

		// Record this large transaction
		k.recordLargeTx(sender, recipient, amount, transferAmt, totalSupply)

		// Update last large tx time
		k.mu.Lock()
		k.lastLargeTxTimes[sender] = k.currentTime
		k.mu.Unlock()
	}

	// Check recipient holding limit after transfer
	k.mu.RLock()
	recipientHolding, exists := k.addressHoldings[recipient]
	if !exists {
		recipientHolding = big.NewInt(0)
	}
	k.mu.RUnlock()

	newHolding := new(big.Int).Add(recipientHolding, transferAmt)

	maxHolding := new(big.Int).Mul(totalSupply, big.NewInt(int64(params.WhaleProtection.MaxHoldingPercentage)))
	maxHolding.Div(maxHolding, big.NewInt(types.BasisPoints))

	if newHolding.Cmp(maxHolding) > 0 {
		return types.ErrWhaleHoldingLimitExceeded
	}

	return nil
}

// UpdateAddressHolding updates the holding amount for an address
func (k *Keeper) UpdateAddressHolding(address, newBalance string) {
	k.mu.Lock()
	defer k.mu.Unlock()

	balance := new(big.Int)
	balance.SetString(newBalance, 10)
	k.addressHoldings[address] = balance
}

// recordLargeTx records a large transaction for monitoring
func (k *Keeper) recordLargeTx(sender, recipient, amount string, transferAmt, totalSupply *big.Int) {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Calculate percentage
	percentage := new(big.Int).Mul(transferAmt, big.NewInt(types.BasisPoints))
	percentage.Div(percentage, totalSupply)

	record := &types.LargeTxRecord{
		TxHash:             fmt.Sprintf("tx_%d_%s", k.currentHeight, sender[:8]),
		Sender:             sender,
		Recipient:          recipient,
		Amount:             amount,
		PercentageOfSupply: percentage.Uint64(),
		BlockHeight:        k.currentHeight,
		Timestamp:          timestamppb.Now(),
		Flagged:            false,
	}

	// Flag if exceeds 0.5% of supply
	if percentage.Uint64() > 50 {
		record.Flagged = true
	}

	k.largeTxRecords = append(k.largeTxRecords, record)

	// Keep only last 1000 records
	if len(k.largeTxRecords) > 1000 {
		k.largeTxRecords = k.largeTxRecords[len(k.largeTxRecords)-1000:]
	}
}

// GetLargeTxRecords returns recent large transaction records
func (k *Keeper) GetLargeTxRecords(limit uint64) []*types.LargeTxRecord {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if limit == 0 || limit > uint64(len(k.largeTxRecords)) {
		limit = uint64(len(k.largeTxRecords))
	}

	start := uint64(len(k.largeTxRecords)) - limit
	return k.largeTxRecords[start:]
}

// GetWhaleProtectionTriggers24h returns count of whale protection triggers in last 24h
func (k *Keeper) GetWhaleProtectionTriggers24h() uint64 {
	k.mu.RLock()
	defer k.mu.RUnlock()

	cutoff := k.currentTime - 86400 // 24 hours ago
	count := uint64(0)

	for _, record := range k.largeTxRecords {
		if record.Timestamp.Seconds >= cutoff && record.Flagged {
			count++
		}
	}

	return count
}
