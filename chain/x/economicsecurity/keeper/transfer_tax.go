package keeper

import (
	"math/big"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// TRANSFER TAX (Feature 6)
// ============================

// CalculateTransferTax calculates the tax amount for a transfer
func (k *Keeper) CalculateTransferTax(sender, amount string) (string, string, string, error) {
	params := k.GetParams()

	if !params.TransferTax.Enabled {
		return "0", "0", "0", nil
	}

	// Check if sender is exempted
	for _, exempted := range params.TransferTax.ExemptedAddresses {
		if sender == exempted {
			return "0", "0", "0", nil
		}
	}

	transferAmt := new(big.Int)
	if _, ok := transferAmt.SetString(amount, 10); !ok {
		return "0", "0", "0", types.ErrInvalidAmount
	}

	// Calculate tax rate (may be dynamic)
	taxRate := params.TransferTax.BaseTaxRate
	if params.TransferTax.DynamicAdjustmentEnabled {
		taxRate = k.calculateDynamicTaxRate(params.TransferTax)
	}

	// Calculate total tax
	tax := new(big.Int).Mul(transferAmt, big.NewInt(int64(taxRate)))
	tax.Div(tax, big.NewInt(types.BasisPoints))

	// Split tax according to distribution percentages
	burnAmt := new(big.Int).Mul(tax, big.NewInt(int64(params.TransferTax.BurnPercentage)))
	burnAmt.Div(burnAmt, big.NewInt(types.BasisPoints))

	treasuryAmt := new(big.Int).Mul(tax, big.NewInt(int64(params.TransferTax.TreasuryPercentage)))
	treasuryAmt.Div(treasuryAmt, big.NewInt(types.BasisPoints))

	redistributeAmt := new(big.Int).Mul(tax, big.NewInt(int64(params.TransferTax.RedistributePercentage)))
	redistributeAmt.Div(redistributeAmt, big.NewInt(types.BasisPoints))

	return tax.String(), burnAmt.String(), treasuryAmt.String(), nil
}

// ProcessTransferTax processes the transfer tax (burns, sends to treasury, redistributes)
func (k *Keeper) ProcessTransferTax(burnAmount, treasuryAmount string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Update total burned
	burnAmt := new(big.Int)
	if _, ok := burnAmt.SetString(burnAmount, 10); !ok {
		return types.ErrInvalidAmount
	}
	k.totalBurned.Add(k.totalBurned, burnAmt)

	// Update circulating supply (burn reduces supply)
	return k.UpdateCirculatingSupply(burnAmount, false)
}

// calculateDynamicTaxRate calculates tax rate based on market conditions
func (k *Keeper) calculateDynamicTaxRate(config *types.TransferTaxConfig) uint64 {
	// Simple implementation: could be enhanced with on-chain volatility metrics
	// For now, return base rate
	// In production, this could adjust based on:
	// - Transaction volume
	// - Price volatility
	// - Liquidity metrics
	return config.BaseTaxRate
}

// GetTaxCollected24h returns the amount of tax collected in last 24 hours
func (k *Keeper) GetTaxCollected24h() string {
	// This would track tax collection over time
	// For simplicity, returning a placeholder
	// In production, would maintain a rolling window of tax events
	return "0"
}
