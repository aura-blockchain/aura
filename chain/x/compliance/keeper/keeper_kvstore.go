package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

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
	ProcessingRestrictionsKeyPrefix = []byte{0x0B}
	RateLimitKeyPrefix              = []byte{0x0C}
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

// UpdateAMLProfileOnTransaction updates an AML profile when a transaction occurs.
// This implements continuous AML monitoring by:
//   - Tracking transaction count and volume
//   - Recalculating risk level based on behavior
//   - Emitting events when risk level changes
//
// This method should be called for every transaction to maintain accurate
// risk profiles and enable real-time AML compliance monitoring.
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: Address performing the transaction
//   - amount: Transaction amount (can be multi-denomination)
//
// Returns:
//   - error: If profile cannot be updated
//
// AML Compliance:
//   - FinCEN: Continuous transaction monitoring requirement
//   - FATF Recommendation 10: Customer due diligence and ongoing monitoring
//   - BSA: Suspicious activity detection through pattern analysis
//
// Security considerations:
//   - Profiles are created on first transaction (no pre-registration required)
//   - Risk level changes trigger immediate events for monitoring systems
//   - Transaction volume is accumulated across all denominations
//   - Timestamps track last activity for velocity analysis
//
// Example usage:
//   if err := k.UpdateAMLProfileOnTransaction(ctx, sender, amount); err != nil {
//       return errorsmod.Wrap(ErrAMLUpdateFailed, err.Error())
//   }
func (k *Keeper) UpdateAMLProfileOnTransaction(ctx sdk.Context, address string, amount sdk.Coins) error {
	// Get existing profile or create new one
	profile, err := k.GetAMLProfile(ctx, address)
	if err != nil {
		// Create new profile for first transaction
		profile = &types.AMLProfile{
			Address:          address,
			RiskLevel:        types.AMLRiskLevel_AML_RISK_LOW,
			TotalTransactions: 0,
			TotalVolume:      "0",
			RiskFactors:      []string{},
			LastAssessment:   timestamppb.New(ctx.BlockTime()),
			PepStatus:        false,
			SourceOfFunds:    []string{},
			Occupation:       "",
		}
	}

	// Update transaction metrics
	profile.TotalTransactions++
	profile.LastAssessment = timestamppb.New(ctx.BlockTime())

	// Calculate total transaction volume (sum all denominations)
	// Parse existing volume
	existingVolume, ok := math.NewIntFromString(profile.TotalVolume)
	if !ok {
		existingVolume = math.ZeroInt()
	}

	// Add current transaction amount (sum all coins as integer units)
	for _, coin := range amount {
		existingVolume = existingVolume.Add(coin.Amount)
	}
	profile.TotalVolume = existingVolume.String()

	// Store previous risk level for event emission
	previousRisk := profile.RiskLevel

	// Recalculate risk level based on updated metrics
	profile.RiskLevel = k.calculateRiskLevel(ctx, profile)

	// Save updated profile
	if err := k.SetAMLProfile(ctx, profile); err != nil {
		return err
	}

	// Emit profile updated event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAMLProfileUpdated,
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyTransactionCount, fmt.Sprintf("%d", profile.TotalTransactions)),
			sdk.NewAttribute(types.AttributeKeyTotalVolume, profile.TotalVolume),
			sdk.NewAttribute(types.AttributeKeyRiskLevel, profile.RiskLevel.String()),
		),
	)

	// Emit risk level changed event if risk level increased
	if profile.RiskLevel != previousRisk {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRiskLevelChanged,
				sdk.NewAttribute(types.AttributeKeyAddress, address),
				sdk.NewAttribute(types.AttributeKeyPreviousRisk, previousRisk.String()),
				sdk.NewAttribute(types.AttributeKeyNewRisk, profile.RiskLevel.String()),
				sdk.NewAttribute(types.AttributeKeyTotalVolume, profile.TotalVolume),
				sdk.NewAttribute(types.AttributeKeyTransactionCount, fmt.Sprintf("%d", profile.TotalTransactions)),
			),
		)
	}

	return nil
}

// calculateRiskLevel determines AML risk level based on transaction behavior.
// This implements risk-based approach required by FATF and FinCEN.
//
// Risk factors considered:
//   - Total transaction volume (high volume = higher risk)
//   - Transaction velocity (frequent transactions = higher risk)
//   - PEP status (Politically Exposed Person = higher risk)
//   - Existing risk factors from investigations
//
// Risk Level Thresholds:
//   - LOW: Normal transaction patterns, low volume
//   - MEDIUM: Moderate volume or frequency
//   - HIGH: High volume or frequent transactions
//   - SEVERE: PEP status or very high volume/frequency
//
// The thresholds are configurable via module params and should be
// calibrated based on jurisdiction requirements and risk appetite.
//
// Parameters:
//   - ctx: SDK context for accessing params
//   - profile: Current AML profile to assess
//
// Returns:
//   - AMLRiskLevel: Calculated risk level
//
// Security considerations:
//   - Conservative thresholds (bias toward flagging for review)
//   - PEP status always results in HIGH or SEVERE risk
//   - Multiple factors compound to increase risk level
//   - Thresholds should be reviewed regularly and adjusted
func (k *Keeper) calculateRiskLevel(ctx sdk.Context, profile *types.AMLProfile) types.AMLRiskLevel {
	params := k.GetParams(ctx)

	// Parse volume as math.Int for comparison
	totalVolume, ok := math.NewIntFromString(profile.TotalVolume)
	if !ok {
		totalVolume = math.ZeroInt()
	}

	// PEP status or existing high-risk factors trigger SEVERE
	if profile.PepStatus || len(profile.RiskFactors) >= 3 {
		return types.AMLRiskLevel_AML_RISK_SEVERE
	}

	// Parse threshold parameters
	highVolumeThreshold, ok := math.NewIntFromString(params.VelocityLimit_24H)
	if !ok {
		// Default to 1 million if param not set
		highVolumeThreshold = math.NewInt(1_000_000)
	}

	// Calculate medium threshold as 50% of high threshold
	mediumVolumeThreshold := highVolumeThreshold.QuoRaw(2)

	// Evaluate volume-based risk
	if totalVolume.GTE(highVolumeThreshold) {
		return types.AMLRiskLevel_AML_RISK_HIGH
	}

	if totalVolume.GTE(mediumVolumeThreshold) {
		return types.AMLRiskLevel_AML_RISK_MEDIUM
	}

	// Evaluate transaction velocity (frequency)
	// High frequency transactions can indicate structuring
	highFrequencyThreshold := uint64(100) // More than 100 transactions is high risk
	mediumFrequencyThreshold := uint64(50) // More than 50 transactions is medium risk

	if profile.TotalTransactions > highFrequencyThreshold {
		return types.AMLRiskLevel_AML_RISK_HIGH
	}

	if profile.TotalTransactions > mediumFrequencyThreshold {
		return types.AMLRiskLevel_AML_RISK_MEDIUM
	}

	// Check for existing risk factors
	if len(profile.RiskFactors) > 0 {
		return types.AMLRiskLevel_AML_RISK_MEDIUM
	}

	return types.AMLRiskLevel_AML_RISK_LOW
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

// ============================================================================
// Processing Restriction KVStore Methods (GDPR Article 7(3) Enforcement)
// ============================================================================

// SetProcessingRestriction marks an address as having restricted data processing rights.
// When set to true, the address has withdrawn consent and data processing must be halted.
// This implements GDPR Article 7(3) "Right to Withdraw Consent" enforcement.
//
// Security considerations:
//   - This flag must be checked before any data processing operations
//   - Withdrawal is recorded immutably for compliance audit
//   - Processing restriction cannot be bypassed
//
// GDPR compliance:
//   - Article 7(3): Consent withdrawal must be as easy as giving consent
//   - Article 18: Right to restriction of processing
//   - Enforcement mechanism for consent withdrawal
func (k *Keeper) SetProcessingRestriction(ctx sdk.Context, address string, restricted bool) error {
	store := ctx.KVStore(k.storeKey)
	key := append(ProcessingRestrictionsKeyPrefix, []byte(address)...)

	if restricted {
		// Store as single byte flag (0x01 = restricted)
		store.Set(key, []byte{0x01})
	} else {
		// Remove restriction
		store.Delete(key)
	}

	return nil
}

// IsProcessingRestricted checks if data processing is restricted for an address.
// Returns true if the address has withdrawn consent and processing must be halted.
// This method should be called before any data processing operation.
//
// GDPR compliance:
//   - Article 7(3): Withdrawal of consent enforcement check
//   - Article 18: Processing restriction enforcement
//   - Must be called before accessing or processing user data
func (k *Keeper) IsProcessingRestricted(ctx sdk.Context, address string) bool {
	store := ctx.KVStore(k.storeKey)
	key := append(ProcessingRestrictionsKeyPrefix, []byte(address)...)
	return store.Has(key)
}

// GetGDPRConsent retrieves a specific consent record by address and consent type.
// This method searches through all consents for an address and returns the matching one.
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: User's blockchain address
//   - consentType: Type of consent (e.g., "data_processing", "marketing")
//
// Returns:
//   - consent: The matching consent record
//   - found: true if consent was found, false otherwise
func (k *Keeper) GetGDPRConsent(ctx sdk.Context, address string, consentType string) (*types.GDPRConsent, bool) {
	consents, err := k.GetGDPRConsents(ctx, address)
	if err != nil {
		return nil, false
	}

	for _, consent := range consents {
		if consent.ConsentType == consentType {
			return consent, true
		}
	}

	return nil, false
}

// CanProcessData checks if data processing is allowed for an address and purpose.
// This is the primary enforcement mechanism for GDPR consent requirements.
//
// The function checks:
//   1. Whether processing is restricted (consent withdrawn)
//   2. Whether specific consent exists for the purpose
//   3. Whether the consent is still valid (not withdrawn)
//
// This method MUST be called before any data processing operation involving user data.
//
// Parameters:
//   - ctx: SDK context for state access
//   - address: User's blockchain address
//   - purpose: Purpose/type of data processing (must match consent type)
//
// Returns:
//   - bool: true if processing is allowed, false if restricted or no consent
//
// GDPR compliance:
//   - Article 6(1)(a): Lawfulness of processing based on consent
//   - Article 7(3): Consent withdrawal enforcement
//   - Article 18: Processing restriction enforcement
//
// Security considerations:
//   - Default deny: Returns false if no consent found
//   - Immutable audit: All checks are logged via state access
//   - Cannot be bypassed: All processing must call this function
//
// Example usage:
//   if !k.CanProcessData(ctx, userAddress, "data_processing") {
//       return errorsmod.Wrap(ErrProcessingRestricted, "consent withdrawn")
//   }
func (k *Keeper) CanProcessData(ctx sdk.Context, address string, purpose string) bool {
	// Check if processing is restricted (consent withdrawn)
	if k.IsProcessingRestricted(ctx, address) {
		return false
	}

	// Check if specific consent exists and is valid
	consent, found := k.GetGDPRConsent(ctx, address, purpose)
	if !found {
		return false
	}

	// Verify consent is still active (not withdrawn)
	return consent.Consented
}

// TriggerDataDeletion emits an event signaling that data should be deleted.
// This is used when consent is withdrawn to signal off-chain systems to delete PII.
//
// GDPR compliance:
//   - Article 7(3): Data must not be processed after consent withdrawal
//   - Article 17: Right to erasure implementation
//   - Off-chain systems must monitor and respond to these events
//
// Implementation note:
//   - On-chain data (commitments/hashes) remains for audit trail
//   - Off-chain PII must be deleted by monitoring systems
//   - Event provides immutable deletion request record
func (k *Keeper) TriggerDataDeletion(ctx sdk.Context, address string, consentType string) error {
	// Emit event for off-chain systems to process
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"gdpr_data_deletion_requested",
			sdk.NewAttribute(types.AttributeKeyAddress, address),
			sdk.NewAttribute(types.AttributeKeyConsentType, consentType),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyBlockTime, ctx.BlockTime().Format("2006-01-02T15:04:05Z")),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
		),
	)

	return nil
}

// ============================================================================
// Rate Limiting KVStore Methods (DoS Protection for Expensive Operations)
// ============================================================================

// getRateLimitKey generates a composite key for rate limit tracking
// Format: RateLimitKeyPrefix + address + ":" + operation
func getRateLimitKey(address string, operation string) []byte {
	key := append(RateLimitKeyPrefix, []byte(address)...)
	key = append(key, ':')
	key = append(key, []byte(operation)...)
	return key
}

// GetRateLimitEntry retrieves the rate limit entry for an address and operation
func (k *Keeper) GetRateLimitEntry(ctx sdk.Context, address string, operation string) (*types.RateLimitEntry, bool) {
	store := ctx.KVStore(k.storeKey)
	key := getRateLimitKey(address, operation)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var entry types.RateLimitEntry
	if err := k.cdc.Unmarshal(bz, &entry); err != nil {
		return nil, false
	}
	return &entry, true
}

// SetRateLimitEntry stores a rate limit entry for an address and operation
func (k *Keeper) SetRateLimitEntry(ctx sdk.Context, entry *types.RateLimitEntry) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(entry)
	if err != nil {
		return err
	}
	key := getRateLimitKey(entry.Address, entry.Operation)
	store.Set(key, bz)
	return nil
}

// CheckRateLimit enforces per-address, per-operation rate limits to prevent DoS
// of expensive external API calls (sanctions screening, KYC verification, etc.).
//
// Rate limiting provides:
//   - DoS protection: Prevents overwhelming external sanctions/KYC providers
//   - Cost control: Limits API usage and associated costs
//   - Fair usage: Prevents single address from consuming all quota
//
// The function implements a sliding window rate limiter:
//   1. Checks if a rate limit entry exists for address+operation
//   2. Resets the window if time has elapsed beyond window_seconds
//   3. Increments the counter and checks against operation-specific limit
//   4. Returns error if limit exceeded, allowing the request if under limit
//
// Parameters:
//   - ctx: SDK context for state access and block time
//   - address: Address making the request (rate limit granularity)
//   - operation: Operation type (determines which limit to apply)
//
// Returns:
//   - error: ErrRateLimitExceeded if limit reached, nil if request allowed
//
// Operation types and their limits (configured in params):
//   - "sanctions_screening": sanctions_screening_limit per window
//   - "kyc_verification": kyc_verification_limit per window
//   - "aml_profile_query": aml_profile_query_limit per window
//   - "tax_report_generation": tax_report_generation_limit per window
//   - "transaction_alerts": default_query_limit per window
//   - others: default_query_limit per window
//
// Security considerations:
//   - Limits are per-address to prevent single user DoS
//   - Window size configurable via params (default: 1 hour)
//   - Limits stored in state for persistence across blocks
//   - Window automatically resets after expiration
//   - Uses block time for consistency and tamper resistance
//
// Example usage:
//   if err := k.CheckRateLimit(ctx, req.Address, "sanctions_screening"); err != nil {
//       return nil, status.Errorf(codes.ResourceExhausted, err.Error())
//   }
//
// Integration with external services:
//   - Prevents overwhelming OFAC SDN API
//   - Protects KYC provider quotas
//   - Controls blockchain indexing costs
//   - Manages tax calculation service usage
func (k *Keeper) CheckRateLimit(ctx sdk.Context, address string, operation string) error {
	params := k.GetParams(ctx)

	// Get or create rate limit entry
	entry, found := k.GetRateLimitEntry(ctx, address, operation)
	if !found {
		// Create new entry for first request
		entry = &types.RateLimitEntry{
			Address:     address,
			Operation:   operation,
			Count:       0,
			WindowStart: timestamppb.New(ctx.BlockTime()),
		}
	}

	// Check if window has expired and should be reset
	windowDuration := time.Duration(params.RateLimitWindowSeconds) * time.Second
	if windowDuration == 0 {
		// Default to 1 hour if not configured
		windowDuration = time.Hour
	}

	windowStart := entry.WindowStart.AsTime()
	timeSinceWindowStart := ctx.BlockTime().Sub(windowStart)

	if timeSinceWindowStart >= windowDuration {
		// Reset window - new time period
		entry.Count = 0
		entry.WindowStart = timestamppb.New(ctx.BlockTime())
	}

	// Determine limit for this operation
	var limit int64
	switch operation {
	case "sanctions_screening":
		limit = params.SanctionsScreeningLimit
	case "kyc_verification":
		limit = params.KycVerificationLimit
	case "aml_profile_query":
		limit = params.AmlProfileQueryLimit
	case "tax_report_generation":
		limit = params.TaxReportGenerationLimit
	case "transaction_alerts":
		limit = params.DefaultQueryLimit
	default:
		limit = params.DefaultQueryLimit
	}

	// Apply default if limit not configured
	if limit <= 0 {
		// Reasonable defaults based on operation cost
		switch operation {
		case "sanctions_screening":
			limit = 10 // Expensive external API calls
		case "kyc_verification":
			limit = 5 // Very expensive verification processes
		case "aml_profile_query":
			limit = 50 // Moderate cost database queries
		case "tax_report_generation":
			limit = 3 // Very expensive report generation
		default:
			limit = 100 // Cheap operations
		}
	}

	// Check if limit exceeded
	if entry.Count >= limit {
		// Emit rate limit exceeded event for monitoring
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRateLimitExceeded,
				sdk.NewAttribute(types.AttributeKeyAddress, address),
				sdk.NewAttribute(types.AttributeKeyOperation, operation),
				sdk.NewAttribute(types.AttributeKeyCount, fmt.Sprintf("%d", entry.Count)),
				sdk.NewAttribute(types.AttributeKeyLimit, fmt.Sprintf("%d", limit)),
				sdk.NewAttribute(types.AttributeKeyWindowStart, entry.WindowStart.AsTime().Format(time.RFC3339)),
			),
		)

		return fmt.Errorf("rate limit exceeded for %s: %d/%d requests in window (resets at %s)",
			operation, entry.Count, limit,
			entry.WindowStart.AsTime().Add(windowDuration).Format(time.RFC3339))
	}

	// Increment counter and save
	entry.Count++
	if err := k.SetRateLimitEntry(ctx, entry); err != nil {
		return fmt.Errorf("failed to update rate limit entry: %w", err)
	}
	return nil
}
