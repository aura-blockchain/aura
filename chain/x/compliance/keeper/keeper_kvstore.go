package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// Key prefixes for KVStore
var (
	KYCRecordsKeyPrefix           = []byte{0x01}
	AMLProfilesKeyPrefix          = []byte{0x02}
	SuspiciousActivitiesKeyPrefix = []byte{0x03}
	MonitoringRulesKeyPrefix      = []byte{0x04}
	TransactionAlertsKeyPrefix    = []byte{0x05}
	SanctionsResultsKeyPrefix     = []byte{0x06}
	GDPRConsentsKeyPrefix         = []byte{0x07}
	GDPRRequestsKeyPrefix         = []byte{0x08}
	TaxReportsKeyPrefix           = []byte{0x09}
	ParamsKeyPrefix               = []byte{0x0A}
)

// ============================================================================
// KYC Record KVStore Methods
// ============================================================================

func (k *Keeper) SetKYCRecord(ctx sdk.Context, record *types.KYCRecord) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(record)
	if err != nil {
		return err
	}
	key := append(KYCRecordsKeyPrefix, []byte(record.Address)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetKYCRecord(ctx sdk.Context, address string) (*types.KYCRecord, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(KYCRecordsKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("KYC record not found: %s", address)
	}

	var record types.KYCRecord
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (k *Keeper) GetAllKYCRecords(ctx sdk.Context) ([]*types.KYCRecord, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, KYCRecordsKeyPrefix)
	defer iterator.Close()

	var records []*types.KYCRecord
	for ; iterator.Valid(); iterator.Next() {
		var record types.KYCRecord
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return nil, err
		}
		records = append(records, &record)
	}
	return records, nil
}

// ============================================================================
// AML Profile KVStore Methods
// ============================================================================

func (k *Keeper) SetAMLProfile(ctx sdk.Context, profile *types.AMLProfile) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(profile)
	if err != nil {
		return err
	}
	key := append(AMLProfilesKeyPrefix, []byte(profile.Address)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetAMLProfile(ctx sdk.Context, address string) (*types.AMLProfile, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(AMLProfilesKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("AML profile not found: %s", address)
	}

	var profile types.AMLProfile
	if err := k.cdc.Unmarshal(bz, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (k *Keeper) GetAllAMLProfiles(ctx sdk.Context) ([]*types.AMLProfile, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, AMLProfilesKeyPrefix)
	defer iterator.Close()

	var profiles []*types.AMLProfile
	for ; iterator.Valid(); iterator.Next() {
		var profile types.AMLProfile
		if err := k.cdc.Unmarshal(iterator.Value(), &profile); err != nil {
			return nil, err
		}
		profiles = append(profiles, &profile)
	}
	return profiles, nil
}

// ============================================================================
// Suspicious Activity KVStore Methods
// ============================================================================

func (k *Keeper) SetSuspiciousActivity(ctx sdk.Context, activity *types.SuspiciousActivity) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(activity)
	if err != nil {
		return err
	}
	key := append(SuspiciousActivitiesKeyPrefix, []byte(activity.Id)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetSuspiciousActivity(ctx sdk.Context, id string) (*types.SuspiciousActivity, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(SuspiciousActivitiesKeyPrefix, []byte(id)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("suspicious activity not found: %s", id)
	}

	var activity types.SuspiciousActivity
	if err := k.cdc.Unmarshal(bz, &activity); err != nil {
		return nil, err
	}
	return &activity, nil
}

func (k *Keeper) GetAllSuspiciousActivities(ctx sdk.Context) ([]*types.SuspiciousActivity, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, SuspiciousActivitiesKeyPrefix)
	defer iterator.Close()

	var activities []*types.SuspiciousActivity
	for ; iterator.Valid(); iterator.Next() {
		var activity types.SuspiciousActivity
		if err := k.cdc.Unmarshal(iterator.Value(), &activity); err != nil {
			return nil, err
		}
		activities = append(activities, &activity)
	}
	return activities, nil
}

// ============================================================================
// Transaction Monitoring Rule KVStore Methods
// ============================================================================

func (k *Keeper) SetMonitoringRule(ctx sdk.Context, rule *types.TransactionMonitoringRule) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(rule)
	if err != nil {
		return err
	}
	key := append(MonitoringRulesKeyPrefix, []byte(rule.Id)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetMonitoringRule(ctx sdk.Context, id string) (*types.TransactionMonitoringRule, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(MonitoringRulesKeyPrefix, []byte(id)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("monitoring rule not found: %s", id)
	}

	var rule types.TransactionMonitoringRule
	if err := k.cdc.Unmarshal(bz, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

func (k *Keeper) GetAllMonitoringRules(ctx sdk.Context) ([]*types.TransactionMonitoringRule, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, MonitoringRulesKeyPrefix)
	defer iterator.Close()

	var rules []*types.TransactionMonitoringRule
	for ; iterator.Valid(); iterator.Next() {
		var rule types.TransactionMonitoringRule
		if err := k.cdc.Unmarshal(iterator.Value(), &rule); err != nil {
			return nil, err
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}

// ============================================================================
// Transaction Alert KVStore Methods (stored per address)
// ============================================================================

func (k *Keeper) AddTransactionAlert(ctx sdk.Context, address string, alert *types.TransactionAlert) error {
	alerts, err := k.GetTransactionAlerts(ctx, address)
	if err != nil {
		return err
	}
	alerts = append(alerts, alert)
	list := &types.TransactionAlertList{Alerts: alerts}
	bz, err := k.cdc.Marshal(list)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	key := append(TransactionAlertsKeyPrefix, []byte(address)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetTransactionAlerts(ctx sdk.Context, address string) ([]*types.TransactionAlert, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(TransactionAlertsKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return []*types.TransactionAlert{}, nil
	}
	var list types.TransactionAlertList
	if err := k.cdc.Unmarshal(bz, &list); err != nil {
		return nil, err
	}
	return list.Alerts, nil
}

// SetTransactionAlert sets transaction alerts for an address (alias for AddTransactionAlert)
func (k *Keeper) SetTransactionAlert(ctx sdk.Context, address string, alert *types.TransactionAlert) error {
	return k.AddTransactionAlert(ctx, address, alert)
}

// GetAllTransactionAlerts retrieves all transaction alerts across all addresses
func (k *Keeper) GetAllTransactionAlerts(ctx sdk.Context) (map[string][]*types.TransactionAlert, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, TransactionAlertsKeyPrefix)
	defer iterator.Close()

	alerts := make(map[string][]*types.TransactionAlert)
	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Key()[len(TransactionAlertsKeyPrefix):])
		var list types.TransactionAlertList
		if err := k.cdc.Unmarshal(iterator.Value(), &list); err != nil {
			return nil, err
		}
		alerts[address] = list.Alerts
	}
	return alerts, nil
}

// ============================================================================
// Sanctions Screening Result KVStore Methods
// ============================================================================

func (k *Keeper) SetSanctionsResult(ctx sdk.Context, result *types.SanctionsScreeningResult) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(result)
	if err != nil {
		return err
	}
	key := append(SanctionsResultsKeyPrefix, []byte(result.Address)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetSanctionsResult(ctx sdk.Context, address string) (*types.SanctionsScreeningResult, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(SanctionsResultsKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("sanctions result not found: %s", address)
	}

	var result types.SanctionsScreeningResult
	if err := k.cdc.Unmarshal(bz, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (k *Keeper) GetAllSanctionsResults(ctx sdk.Context) ([]*types.SanctionsScreeningResult, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, SanctionsResultsKeyPrefix)
	defer iterator.Close()

	var results []*types.SanctionsScreeningResult
	for ; iterator.Valid(); iterator.Next() {
		var result types.SanctionsScreeningResult
		if err := k.cdc.Unmarshal(iterator.Value(), &result); err != nil {
			return nil, err
		}
		results = append(results, &result)
	}
	return results, nil
}

// ============================================================================
// GDPR Consent KVStore Methods (nested by address and consent type)
// ============================================================================

func (k *Keeper) SetGDPRConsent(ctx sdk.Context, consent *types.GDPRConsent) error {
	consents, err := k.GetGDPRConsents(ctx, consent.Address)
	if err != nil {
		return err
	}

	// Update or add consent
	found := false
	for i, existing := range consents {
		if existing.ConsentType == consent.ConsentType {
			consents[i] = consent
			found = true
			break
		}
	}
	if !found {
		consents = append(consents, consent)
	}

	store := ctx.KVStore(k.storeKey)
	list := &types.GDPRConsentList{Consents: consents}
	bz, err := k.cdc.Marshal(list)
	if err != nil {
		return err
	}
	key := append(GDPRConsentsKeyPrefix, []byte(consent.Address)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetGDPRConsents(ctx sdk.Context, address string) ([]*types.GDPRConsent, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(GDPRConsentsKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return []*types.GDPRConsent{}, nil
	}
	var list types.GDPRConsentList
	if err := k.cdc.Unmarshal(bz, &list); err != nil {
		return nil, err
	}
	return list.Consents, nil
}

// GetAllGDPRConsents retrieves all GDPR consents across all addresses
func (k *Keeper) GetAllGDPRConsents(ctx sdk.Context) (map[string][]*types.GDPRConsent, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, GDPRConsentsKeyPrefix)
	defer iterator.Close()

	consents := make(map[string][]*types.GDPRConsent)
	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Key()[len(GDPRConsentsKeyPrefix):])
		var list types.GDPRConsentList
		if err := k.cdc.Unmarshal(iterator.Value(), &list); err != nil {
			return nil, err
		}
		consents[address] = list.Consents
	}
	return consents, nil
}

// ============================================================================
// GDPR Data Request KVStore Methods
// ============================================================================

func (k *Keeper) SetGDPRRequest(ctx sdk.Context, request *types.GDPRDataRequest) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(request)
	if err != nil {
		return err
	}
	key := append(GDPRRequestsKeyPrefix, []byte(request.Id)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetGDPRRequest(ctx sdk.Context, requestID string) (*types.GDPRDataRequest, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(GDPRRequestsKeyPrefix, []byte(requestID)...)
	bz := store.Get(key)
	if bz == nil {
		return nil, fmt.Errorf("GDPR request not found: %s", requestID)
	}

	var request types.GDPRDataRequest
	if err := k.cdc.Unmarshal(bz, &request); err != nil {
		return nil, err
	}
	return &request, nil
}

func (k *Keeper) GetAllGDPRRequests(ctx sdk.Context) ([]*types.GDPRDataRequest, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, GDPRRequestsKeyPrefix)
	defer iterator.Close()

	var requests []*types.GDPRDataRequest
	for ; iterator.Valid(); iterator.Next() {
		var request types.GDPRDataRequest
		if err := k.cdc.Unmarshal(iterator.Value(), &request); err != nil {
			return nil, err
		}
		requests = append(requests, &request)
	}
	return requests, nil
}

// ============================================================================
// Tax Report KVStore Methods (nested by address and tax year)
// ============================================================================

func (k *Keeper) SetTaxReport(ctx sdk.Context, report *types.TaxReport) error {
	reports, err := k.GetTaxReports(ctx, report.Address)
	if err != nil {
		return err
	}

	// Update or add report
	found := false
	for i, existing := range reports {
		if existing.TaxYear == report.TaxYear {
			reports[i] = report
			found = true
			break
		}
	}
	if !found {
		reports = append(reports, report)
	}

	store := ctx.KVStore(k.storeKey)
	list := &types.TaxReportList{Reports: reports}
	bz, err := k.cdc.Marshal(list)
	if err != nil {
		return err
	}
	key := append(TaxReportsKeyPrefix, []byte(report.Address)...)
	store.Set(key, bz)
	return nil
}

func (k *Keeper) GetTaxReports(ctx sdk.Context, address string) ([]*types.TaxReport, error) {
	store := ctx.KVStore(k.storeKey)
	key := append(TaxReportsKeyPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return []*types.TaxReport{}, nil
	}
	var list types.TaxReportList
	if err := k.cdc.Unmarshal(bz, &list); err != nil {
		return nil, err
	}
	return list.Reports, nil
}

// GetAllTaxReports retrieves all tax reports across all addresses
func (k *Keeper) GetAllTaxReports(ctx sdk.Context) (map[string][]*types.TaxReport, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, TaxReportsKeyPrefix)
	defer iterator.Close()

	reports := make(map[string][]*types.TaxReport)
	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Key()[len(TaxReportsKeyPrefix):])
		var list types.TaxReportList
		if err := k.cdc.Unmarshal(iterator.Value(), &list); err != nil {
			return nil, err
		}
		reports[address] = list.Reports
	}
	return reports, nil
}

// ============================================================================
// Params KVStore Methods
// ============================================================================

func (k *Keeper) GetParamsFromStore(ctx sdk.Context) (types.ComplianceParams, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ParamsKeyPrefix)
	if bz == nil {
		return types.ComplianceParams{}, nil
	}

	var params types.ComplianceParams
	if err := k.cdc.Unmarshal(bz, &params); err != nil {
		return types.ComplianceParams{}, err
	}
	return params, nil
}

func (k *Keeper) SetParamsToStore(ctx sdk.Context, params types.ComplianceParams) error {
	if err := types.ValidateParams(params); err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(ParamsKeyPrefix, bz)
	return nil
}
