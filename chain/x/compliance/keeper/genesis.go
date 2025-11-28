package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// InitGenesis initializes the module state from genesis data
func (k *Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) error {
	if err := types.ValidateGenesis(data); err != nil {
		return fmt.Errorf("invalid genesis state: %w", err)
	}

	// Set params
	if data.Params != nil {
		if err := k.SetParams(ctx, *data.Params); err != nil {
			return fmt.Errorf("failed to set params: %w", err)
		}
	}

	// Import KYC records
	for _, record := range data.KycRecords {
		if record == nil {
			continue
		}
		if err := k.SetKYCRecord(ctx, record); err != nil {
			return fmt.Errorf("failed to set KYC record: %w", err)
		}
	}

	// Import AML profiles
	for _, profile := range data.AmlProfiles {
		if profile == nil {
			continue
		}
		if err := k.SetAMLProfile(ctx, profile); err != nil {
			return fmt.Errorf("failed to set AML profile: %w", err)
		}
	}

	// Import suspicious activities
	for _, activity := range data.SuspiciousActivities {
		if activity == nil {
			continue
		}
		if err := k.SetSuspiciousActivity(ctx, activity); err != nil {
			return fmt.Errorf("failed to set suspicious activity: %w", err)
		}
	}

	// Import monitoring rules
	for _, rule := range data.MonitoringRules {
		if rule == nil {
			continue
		}
		if err := k.SetMonitoringRule(ctx, rule); err != nil {
			return fmt.Errorf("failed to set monitoring rule: %w", err)
		}
	}

	// Import transaction alerts
	for _, alert := range data.TransactionAlerts {
		if alert == nil {
			continue
		}
		// Extract address from alert; use a key field if address not directly available
		address := alert.Address
		if address == "" {
			// Fallback to using alert ID as key if address is not available
			address = alert.TransactionHash
		}
		if err := k.SetTransactionAlert(ctx, address, alert); err != nil {
			return fmt.Errorf("failed to set transaction alert: %w", err)
		}
	}

	// Import sanctions results
	for _, result := range data.SanctionsResults {
		if result == nil {
			continue
		}
		if err := k.SetSanctionsResult(ctx, result); err != nil {
			return fmt.Errorf("failed to set sanctions result: %w", err)
		}
	}

	// Import GDPR consents
	for _, consent := range data.GdprConsents {
		if consent == nil {
			continue
		}
		if err := k.SetGDPRConsent(ctx, consent); err != nil {
			return fmt.Errorf("failed to set GDPR consent: %w", err)
		}
	}

	// Import GDPR requests
	for _, request := range data.GdprRequests {
		if request == nil {
			continue
		}
		if err := k.SetGDPRRequest(ctx, request); err != nil {
			return fmt.Errorf("failed to set GDPR request: %w", err)
		}
	}

	// Import tax reports
	for _, report := range data.TaxReports {
		if report == nil {
			continue
		}
		if err := k.SetTaxReport(ctx, report); err != nil {
			return fmt.Errorf("failed to set tax report: %w", err)
		}
	}

	k.logger(ctx).Info("Compliance genesis imported",
		"kyc_records", len(data.KycRecords),
		"aml_profiles", len(data.AmlProfiles),
		"suspicious_activities", len(data.SuspiciousActivities))

	return nil
}

// ExportGenesis exports the current module state to genesis
func (k *Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)

	// Export KYC records
	kycRecords, err := k.GetAllKYCRecords(ctx)
	if err != nil {
		k.logger(ctx).Error("failed to get KYC records", "error", err)
		kycRecords = []*types.KYCRecord{}
	}

	// Export AML profiles
	amlProfiles, err := k.GetAllAMLProfiles(ctx)
	if err != nil {
		k.logger(ctx).Error("failed to get AML profiles", "error", err)
		amlProfiles = []*types.AMLProfile{}
	}

	// Export suspicious activities
	suspiciousActivities, err := k.GetAllSuspiciousActivities(ctx)
	if err != nil {
		k.logger(ctx).Error("failed to get suspicious activities", "error", err)
		suspiciousActivities = []*types.SuspiciousActivity{}
	}

	// Export monitoring rules
	monitoringRules, err := k.GetAllMonitoringRules(ctx)
	if err != nil {
		k.logger(ctx).Error("failed to get monitoring rules", "error", err)
		monitoringRules = []*types.TransactionMonitoringRule{}
	}

	// Export transaction alerts - convert map to slice for proto
	transactionAlertsMap, err := k.GetAllTransactionAlerts(ctx)
	if err != nil {
		k.logger(ctx).Error("failed to get transaction alerts", "error", err)
		transactionAlertsMap = make(map[string][]*types.TransactionAlert)
	}
	var transactionAlerts []*types.TransactionAlert
	for _, alerts := range transactionAlertsMap {
		transactionAlerts = append(transactionAlerts, alerts...)
	}

	// Export sanctions results
	sanctionsResults, err := k.GetAllSanctionsResults(ctx)
	if err != nil {
		k.logger(ctx).Error("failed to get sanctions results", "error", err)
		sanctionsResults = []*types.SanctionsScreeningResult{}
	}

	// Export GDPR consents - convert map to slice for proto
	gdprConsentsMap, err := k.GetAllGDPRConsents(ctx)
	if err != nil {
		k.logger(ctx).Error("failed to get GDPR consents", "error", err)
		gdprConsentsMap = make(map[string][]*types.GDPRConsent)
	}
	var gdprConsents []*types.GDPRConsent
	for _, consents := range gdprConsentsMap {
		gdprConsents = append(gdprConsents, consents...)
	}

	// Export GDPR requests
	gdprRequests, err := k.GetAllGDPRRequests(ctx)
	if err != nil {
		k.logger(ctx).Error("failed to get GDPR requests", "error", err)
		gdprRequests = []*types.GDPRDataRequest{}
	}

	// Export tax reports - convert map to slice for proto
	taxReportsMap, err := k.GetAllTaxReports(ctx)
	if err != nil {
		k.logger(ctx).Error("failed to get tax reports", "error", err)
		taxReportsMap = make(map[string][]*types.TaxReport)
	}
	var taxReports []*types.TaxReport
	for _, reports := range taxReportsMap {
		taxReports = append(taxReports, reports...)
	}

	return &types.GenesisState{
		Params:               &params,
		KycRecords:           kycRecords,
		AmlProfiles:          amlProfiles,
		SuspiciousActivities: suspiciousActivities,
		MonitoringRules:      monitoringRules,
		TransactionAlerts:    transactionAlerts,
		SanctionsResults:     sanctionsResults,
		GdprConsents:         gdprConsents,
		GdprRequests:         gdprRequests,
		TaxReports:           taxReports,
	}
}
