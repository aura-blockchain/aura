package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// GenerateTaxReport generates a tax report for an address
// Feature 8: Tax reporting capabilities (1099 forms, etc.)
func (k *Keeper) GenerateTaxReport(
	address string,
	taxYear string,
	jurisdiction string,
	reportType string,
	transactions []*types.TaxTransaction,
) (string, error) {
	if !k.params.TaxReportingEnabled {
		return "", fmt.Errorf("tax reporting not enabled")
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Validate jurisdiction
	validJurisdiction := false
	for _, jur := range k.params.TaxJurisdictions {
		if jur == jurisdiction {
			validJurisdiction = true
			break
		}
	}
	if !validJurisdiction {
		return "", fmt.Errorf("invalid jurisdiction: %s", jurisdiction)
	}

	// Generate report ID
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%s", address, taxYear, jurisdiction, reportType)))
	reportID := hex.EncodeToString(hash[:])[:16]

	// Calculate tax totals
	totalIncome := big.NewInt(0)
	totalCapitalGains := big.NewInt(0)
	totalCapitalLosses := big.NewInt(0)

	for _, tx := range transactions {
		if tx.IsIncome {
			amount, ok := new(big.Int).SetString(tx.FairMarketValue, 10)
			if ok {
				totalIncome.Add(totalIncome, amount)
			}
		} else {
			gainLoss, ok := new(big.Int).SetString(tx.GainLoss, 10)
			if ok {
				if gainLoss.Sign() > 0 {
					totalCapitalGains.Add(totalCapitalGains, gainLoss)
				} else {
					loss := new(big.Int).Abs(gainLoss)
					totalCapitalLosses.Add(totalCapitalLosses, loss)
				}
			}
		}
	}

	report := &types.TaxReport{
		ID:                 reportID,
		Address:            address,
		TaxYear:            taxYear,
		Jurisdiction:       jurisdiction,
		ReportType:         reportType,
		Transactions:       transactions,
		TotalIncome:        totalIncome.String(),
		TotalCapitalGains:  totalCapitalGains.String(),
		TotalCapitalLosses: totalCapitalLosses.String(),
		GeneratedAt:        time.Now(),
		Filed:              false,
	}

	if err := report.Validate(); err != nil {
		return "", fmt.Errorf("invalid tax report: %w", err)
	}

	// Use registered generator if available
	if generator, exists := k.taxReportGenerators[jurisdiction]; exists {
		generatedReport, err := generator.GenerateReport(address, taxYear, reportType, transactions)
		if err == nil {
			report = generatedReport
			// Export to file
			filePath, err := generator.ExportToFile(report, "pdf")
			if err == nil {
				report.FilePath = filePath
			}
		}
	}

	// Store report
	if k.taxReports[address] == nil {
		k.taxReports[address] = make(map[string]*types.TaxReport)
	}
	k.taxReports[address][taxYear] = report

	return reportID, nil
}

// GetTaxReport retrieves a tax report
func (k *Keeper) GetTaxReport(address string, taxYear string) (*types.TaxReport, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.taxReports[address] == nil {
		return nil, fmt.Errorf("no tax reports found for address: %s", address)
	}

	report, exists := k.taxReports[address][taxYear]
	if !exists {
		return nil, fmt.Errorf("no tax report found for year %s", taxYear)
	}

	return report, nil
}

// MarkTaxReportFiled marks a tax report as filed
func (k *Keeper) MarkTaxReportFiled(address string, taxYear string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.taxReports[address] == nil {
		return fmt.Errorf("no tax reports found for address: %s", address)
	}

	report, exists := k.taxReports[address][taxYear]
	if !exists {
		return fmt.Errorf("no tax report found for year %s", taxYear)
	}

	report.Filed = true
	report.FiledAt = time.Now()

	return nil
}

// Generate1099MISC generates a 1099-MISC form for miscellaneous income
func (k *Keeper) Generate1099MISC(
	address string,
	taxYear string,
	incomeTransactions []*types.TaxTransaction,
) (string, error) {
	// Filter for income transactions only
	miscIncome := []*types.TaxTransaction{}
	for _, tx := range incomeTransactions {
		if tx.IsIncome {
			miscIncome = append(miscIncome, tx)
		}
	}

	return k.GenerateTaxReport(address, taxYear, "US", "1099-MISC", miscIncome)
}

// Generate1099K generates a 1099-K form for payment card and third-party network transactions
func (k *Keeper) Generate1099K(
	address string,
	taxYear string,
	paymentTransactions []*types.TaxTransaction,
) (string, error) {
	// 1099-K is required if gross payments exceed $20,000 and >200 transactions
	if len(paymentTransactions) < 200 {
		return "", fmt.Errorf("1099-K not required: less than 200 transactions")
	}

	totalPayments := big.NewInt(0)
	for _, tx := range paymentTransactions {
		amount, ok := new(big.Int).SetString(tx.Amount, 10)
		if ok {
			totalPayments.Add(totalPayments, amount)
		}
	}

	threshold := big.NewInt(20000)
	if totalPayments.Cmp(threshold) < 0 {
		return "", fmt.Errorf("1099-K not required: gross payments below $20,000")
	}

	return k.GenerateTaxReport(address, taxYear, "US", "1099-K", paymentTransactions)
}

// Generate8949 generates Form 8949 for capital gains and losses
func (k *Keeper) Generate8949(
	address string,
	taxYear string,
	capitalTransactions []*types.TaxTransaction,
) (string, error) {
	// Filter for capital gain/loss transactions
	capitalTxs := []*types.TaxTransaction{}
	for _, tx := range capitalTransactions {
		if !tx.IsIncome {
			capitalTxs = append(capitalTxs, tx)
		}
	}

	return k.GenerateTaxReport(address, taxYear, "US", "8949", capitalTxs)
}

// CalculateCapitalGains calculates capital gains/losses for a transaction
func (k *Keeper) CalculateCapitalGains(
	asset string,
	acquiredDate time.Time,
	soldDate time.Time,
	costBasis string,
	proceeds string,
) (string, bool, error) {
	costBasisInt, ok := new(big.Int).SetString(costBasis, 10)
	if !ok {
		return "", false, fmt.Errorf("invalid cost basis")
	}

	proceedsInt, ok := new(big.Int).SetString(proceeds, 10)
	if !ok {
		return "", false, fmt.Errorf("invalid proceeds")
	}

	gainLoss := new(big.Int).Sub(proceedsInt, costBasisInt)

	// Determine if long-term or short-term
	holdingPeriod := soldDate.Sub(acquiredDate)
	isLongTerm := holdingPeriod.Hours() > 24*365 // More than 1 year

	return gainLoss.String(), isLongTerm, nil
}

// ClassifyTransaction classifies a transaction for tax purposes
func (k *Keeper) ClassifyTransaction(
	txType string,
	amount string,
	timestamp time.Time,
) (bool, string, error) {
	// Classify transaction as income or capital gain/loss
	var isIncome bool
	var category string

	switch txType {
	case "stake_reward", "staking":
		isIncome = true
		category = "staking_rewards"
	case "airdrop":
		isIncome = true
		category = "airdrop"
	case "mining", "mining_reward":
		isIncome = true
		category = "mining"
	case "trade", "swap":
		isIncome = false
		category = "capital_transaction"
	case "transfer":
		isIncome = false
		category = "transfer"
	default:
		return false, "", fmt.Errorf("unknown transaction type: %s", txType)
	}

	return isIncome, category, nil
}

// GetTaxSummary generates a tax summary for an address
func (k *Keeper) GetTaxSummary(address string, taxYear string) (map[string]interface{}, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.taxReports[address] == nil || k.taxReports[address][taxYear] == nil {
		return nil, fmt.Errorf("no tax report found for %s in %s", address, taxYear)
	}

	report := k.taxReports[address][taxYear]

	summary := make(map[string]interface{})
	summary["address"] = address
	summary["tax_year"] = taxYear
	summary["total_income"] = report.TotalIncome
	summary["total_capital_gains"] = report.TotalCapitalGains
	summary["total_capital_losses"] = report.TotalCapitalLosses
	summary["transaction_count"] = len(report.Transactions)
	summary["report_generated"] = report.GeneratedAt
	summary["filed"] = report.Filed

	// Calculate net capital gains/losses
	gains, _ := new(big.Int).SetString(report.TotalCapitalGains, 10)
	losses, _ := new(big.Int).SetString(report.TotalCapitalLosses, 10)
	netCapital := new(big.Int).Sub(gains, losses)
	summary["net_capital_gain_loss"] = netCapital.String()

	// Breakdown by transaction type
	typeBreakdown := make(map[string]int)
	for _, tx := range report.Transactions {
		typeBreakdown[tx.TransactionType]++
	}
	summary["transaction_types"] = typeBreakdown

	return summary, nil
}

// GetTaxableEvents returns all taxable events for an address in a tax year
func (k *Keeper) GetTaxableEvents(address string, taxYear string) ([]*types.TaxTransaction, error) {
	report, err := k.GetTaxReport(address, taxYear)
	if err != nil {
		return nil, err
	}

	return report.Transactions, nil
}

// EstimateTaxLiability estimates tax liability (simplified)
func (k *Keeper) EstimateTaxLiability(
	address string,
	taxYear string,
	jurisdiction string,
) (map[string]string, error) {
	report, err := k.GetTaxReport(address, taxYear)
	if err != nil {
		return nil, err
	}

	liability := make(map[string]string)

	// Parse totals
	income, _ := new(big.Int).SetString(report.TotalIncome, 10)
	gains, _ := new(big.Int).SetString(report.TotalCapitalGains, 10)
	losses, _ := new(big.Int).SetString(report.TotalCapitalLosses, 10)

	// Simplified tax calculation for US (example)
	if jurisdiction == "US" {
		// Ordinary income tax (simplified - assume 22% bracket)
		ordinaryTax := new(big.Int).Mul(income, big.NewInt(22))
		ordinaryTax.Div(ordinaryTax, big.NewInt(100))
		liability["ordinary_income_tax"] = ordinaryTax.String()

		// Capital gains tax (simplified - assume 15% for long-term)
		netGains := new(big.Int).Sub(gains, losses)
		if netGains.Sign() > 0 {
			capitalGainsTax := new(big.Int).Mul(netGains, big.NewInt(15))
			capitalGainsTax.Div(capitalGainsTax, big.NewInt(100))
			liability["capital_gains_tax"] = capitalGainsTax.String()
		} else {
			liability["capital_gains_tax"] = "0"
		}

		// Total estimated tax
		totalTax := new(big.Int).Add(ordinaryTax, big.NewInt(0))
		if capTax, ok := new(big.Int).SetString(liability["capital_gains_tax"], 10); ok {
			totalTax.Add(totalTax, capTax)
		}
		liability["total_estimated_tax"] = totalTax.String()
	}

	liability["jurisdiction"] = jurisdiction
	liability["tax_year"] = taxYear
	liability["note"] = "This is a simplified estimate. Consult a tax professional for accurate calculation."

	return liability, nil
}

// ExportTaxReportCSV exports a tax report to CSV format
func (k *Keeper) ExportTaxReportCSV(address string, taxYear string) (string, error) {
	report, err := k.GetTaxReport(address, taxYear)
	if err != nil {
		return "", err
	}

	// Build CSV content
	csv := "Transaction Hash,Date,Type,Asset,Amount,Cost Basis,Fair Market Value,Gain/Loss,Is Income\n"

	for _, tx := range report.Transactions {
		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%t\n",
			tx.TransactionHash,
			tx.Timestamp.Format("2006-01-02"),
			tx.TransactionType,
			tx.Asset,
			tx.Amount,
			tx.CostBasis,
			tx.FairMarketValue,
			tx.GainLoss,
			tx.IsIncome,
		)
	}

	// In production, this would write to a file
	// For now, return the CSV string
	return csv, nil
}

// GetTaxReportingStatistics returns statistics about tax reports
func (k *Keeper) GetTaxReportingStatistics() map[string]interface{} {
	k.mu.RLock()
	defer k.mu.RUnlock()

	stats := make(map[string]interface{})

	totalReports := 0
	filedReports := 0
	unfiledReports := 0
	reportsByYear := make(map[string]int)
	reportsByJurisdiction := make(map[string]int)

	for _, reports := range k.taxReports {
		for year, report := range reports {
			totalReports++
			reportsByYear[year]++
			reportsByJurisdiction[report.Jurisdiction]++

			if report.Filed {
				filedReports++
			} else {
				unfiledReports++
			}
		}
	}

	stats["total_reports"] = totalReports
	stats["filed_reports"] = filedReports
	stats["unfiled_reports"] = unfiledReports
	stats["reports_by_year"] = reportsByYear
	stats["reports_by_jurisdiction"] = reportsByJurisdiction

	return stats
}

// CheckTaxReportingRequirements checks if tax reporting is required for an address
func (k *Keeper) CheckTaxReportingRequirements(address string, taxYear string) (bool, []string, error) {
	report, err := k.GetTaxReport(address, taxYear)
	if err != nil {
		// No report exists, check if one is needed based on activity
		return false, []string{}, nil
	}

	required := false
	forms := []string{}

	// Parse totals
	income, _ := new(big.Int).SetString(report.TotalIncome, 10)

	// Check if 1099-MISC is required (income >= $600)
	if income.Cmp(big.NewInt(600)) >= 0 {
		required = true
		forms = append(forms, "1099-MISC")
	}

	// Check if 1099-K is required
	if len(report.Transactions) >= 200 {
		totalPayments := big.NewInt(0)
		for _, tx := range report.Transactions {
			amount, ok := new(big.Int).SetString(tx.Amount, 10)
			if ok {
				totalPayments.Add(totalPayments, amount)
			}
		}
		if totalPayments.Cmp(big.NewInt(20000)) >= 0 {
			required = true
			forms = append(forms, "1099-K")
		}
	}

	// Check if Form 8949 is required (any capital transactions)
	hasCapitalTx := false
	for _, tx := range report.Transactions {
		if !tx.IsIncome {
			hasCapitalTx = true
			break
		}
	}
	if hasCapitalTx {
		required = true
		forms = append(forms, "8949")
	}

	return required, forms, nil
}
