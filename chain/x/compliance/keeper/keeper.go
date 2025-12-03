package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// Keeper maintains the state of the compliance module
type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec

	// External service integrations (not persisted in state)
	kycProviders        map[string]KYCProvider
	sanctionsProviders  map[string]SanctionsProvider
	taxReportGenerators map[string]TaxReportGenerator
	sanctionsCache      map[string]time.Time // address -> last screening time (transient)

	// Data protection service for PII commitments (GDPR Article 32 compliance)
	dataProtection *DataProtectionService
}

// NewKeeper creates a new compliance keeper
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) *Keeper {
	return &Keeper{
		cdc:                 cdc,
		storeKey:            storeKey,
		kycProviders:        make(map[string]KYCProvider),
		sanctionsProviders:  make(map[string]SanctionsProvider),
		taxReportGenerators: make(map[string]TaxReportGenerator),
		sanctionsCache:      make(map[string]time.Time),
		dataProtection:      NewDataProtectionService(),
	}
}

// GetDataProtectionService returns the data protection service for PII commitments
//
// Use this service to:
// - Generate SHA-256 commitments for sensitive data before on-chain storage
// - Verify data integrity against stored commitments
// - Protect PII according to GDPR Article 32 requirements
//
// See: chain/x/compliance/keeper/DATA_PROTECTION_ARCHITECTURE.md for usage patterns
func (k *Keeper) GetDataProtectionService() *DataProtectionService {
	return k.dataProtection
}

// Logger returns a module-specific logger
func (k Keeper) logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// StoreKey returns the keeper's store key.
func (k Keeper) StoreKey() storetypes.StoreKey {
	return k.storeKey
}

// initializeDefaultMonitoringRules sets up default transaction monitoring rules
func (k *Keeper) initializeDefaultMonitoringRules(ctx sdk.Context) error {
	params, err := k.GetParamsFromStore(ctx)
	if err != nil {
		return err
	}

	defaultRules := []*types.TransactionMonitoringRule{
		{
			Id:          "large_transaction",
			Name:        "Large Transaction Monitor",
			Description: "Monitor large transactions",
			RuleType:    "velocity",
			Parameters: map[string]string{
				"threshold": params.SingleTransactionLimit,
				"action":    "flag_for_review",
			},
			RiskLevel: types.TransactionRiskLevel_TX_RISK_HIGH,
			Enabled:   true,
		},
		{
			Id:          "high_frequency",
			Name:        "High Frequency Monitor",
			Description: "Monitor high frequency transactions",
			RuleType:    "velocity",
			Parameters: map[string]string{
				"threshold_24h": params.VelocityLimit_24H,
				"action":        "flag_for_review",
			},
			RiskLevel: types.TransactionRiskLevel_TX_RISK_MEDIUM,
			Enabled:   true,
		},
		{
			Id:          "suspicious_pattern",
			Name:        "Suspicious Pattern Monitor",
			Description: "Monitor suspicious transaction patterns",
			RuleType:    "structuring",
			Parameters: map[string]string{
				"count_threshold": "10",
				"action":          "automatic_hold",
			},
			RiskLevel: types.TransactionRiskLevel_TX_RISK_CRITICAL,
			Enabled:   true,
		},
	}

	for _, rule := range defaultRules {
		if err := k.SetMonitoringRule(ctx, rule); err != nil {
			return fmt.Errorf("failed to set monitoring rule %s: %w", rule.Id, err)
		}
	}

	return nil
}

// Monitoring rule CRUD operations are implemented in keeper_kvstore.go using proper proto types

// GetParams returns the current module parameters
func (k *Keeper) GetParams(ctx sdk.Context) types.ComplianceParams {
	params, _ := k.GetParamsFromStore(ctx)
	return params
}

// SetParams updates the module parameters
func (k *Keeper) SetParams(ctx sdk.Context, params types.ComplianceParams) error {
	return k.SetParamsToStore(ctx, params)
}

// RegisterKYCProvider registers a KYC provider for integration
func (k *Keeper) RegisterKYCProvider(name string, provider KYCProvider) {
	k.kycProviders[name] = provider
}

// RegisterSanctionsProvider registers a sanctions screening provider
func (k *Keeper) RegisterSanctionsProvider(name string, provider SanctionsProvider) {
	k.sanctionsProviders[name] = provider
}

// RegisterTaxReportGenerator registers a tax report generator
func (k *Keeper) RegisterTaxReportGenerator(jurisdiction string, generator TaxReportGenerator) {
	k.taxReportGenerators[jurisdiction] = generator
}

// IsJurisdictionBlocked checks if a jurisdiction is blocked due to OFAC sanctions.
// Jurisdiction codes use ISO 3166-1 alpha-2 format (e.g., "US", "KP", "IR").
//
// OFAC Compliance:
//   - Validates against the blocked_jurisdictions list in module params
//   - Returns true if jurisdiction is sanctioned (user from that country cannot use platform)
//   - Case-insensitive comparison for robustness
//   - Governance can update the blocked list via params
//
// Security considerations:
//   - Empty jurisdiction string is treated as blocked (fail-safe)
//   - Invalid country codes should be blocked by upstream validation
//   - This check should occur before accepting KYC records
//
// Returns:
//   - true: Jurisdiction is blocked (OFAC sanctioned)
//   - false: Jurisdiction is allowed
func (k Keeper) IsJurisdictionBlocked(ctx sdk.Context, jurisdiction string) bool {
	if jurisdiction == "" {
		return true // Fail-safe: empty jurisdiction is blocked
	}

	params := k.GetParams(ctx)

	// Case-insensitive check for blocked jurisdictions
	jurisdictionUpper := toUpperASCII(jurisdiction)
	for _, blocked := range params.BlockedJurisdictions {
		if toUpperASCII(blocked) == jurisdictionUpper {
			return true
		}
	}

	return false
}

// toUpperASCII converts ASCII letters to uppercase without allocating (simple case for country codes)
func toUpperASCII(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - 32 // Convert to uppercase
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// Interface definitions for external service providers
type KYCProvider interface {
	VerifyIdentity(address, documentType string, documents [][]byte) (*types.KYCRecord, error)
	GetVerificationStatus(verificationID string) (*types.KYCRecord, error)
	UpdateRiskScore(address string) (string, error)
}

type SanctionsProvider interface {
	ScreenAddress(address string) (*types.SanctionsScreeningResult, error)
	CheckLists(lists []string) ([]*types.SanctionsMatch, error)
}

type TaxReportGenerator interface {
	GenerateReport(address, taxYear, reportType string, transactions []*types.TaxTransaction) (*types.TaxReport, error)
	ExportToFile(report *types.TaxReport, format string) (string, error)
}
