package keeper

import (
	"time"

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
	}
}

// initializeDefaultMonitoringRules sets up default transaction monitoring rules
func (k *Keeper) initializeDefaultMonitoringRules(ctx sdk.Context) error {
	params, err := k.GetParamsFromStore(ctx)
	if err != nil {
		return err
	}

	now := time.Now()

	// Velocity rule - detect rapid succession of transactions
	rule1 := &types.TransactionMonitoringRule{
		ID:          "velocity_24h",
		Name:        "24-Hour Velocity Limit",
		Description: "Triggers when transaction volume exceeds limit in 24 hours",
		RuleType:    "velocity",
		Parameters: map[string]string{
			"time_window": "24h",
			"limit":       k.params.VelocityLimit24h,
		},
		RiskLevel: types.TxRiskMedium,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Large transaction rule
	k.monitoringRules["large_transaction"] = &types.TransactionMonitoringRule{
		ID:          "large_transaction",
		Name:        "Large Transaction Alert",
		Description: "Triggers for transactions exceeding the single transaction limit",
		RuleType:    "threshold",
		Parameters: map[string]string{
			"threshold": k.params.SingleTransactionLimit,
		},
		RiskLevel: types.TxRiskHigh,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Structuring detection - multiple transactions just below reporting threshold
	k.monitoringRules["structuring"] = &types.TransactionMonitoringRule{
		ID:          "structuring",
		Name:        "Structuring Detection",
		Description: "Detects potential structuring behavior (multiple txs just below threshold)",
		RuleType:    "structuring",
		Parameters: map[string]string{
			"count_threshold":  string(rune(k.params.StructuringThresholdCount)),
			"amount_threshold": "9900", // Just below 10K reporting threshold
			"time_window":      "24h",
		},
		RiskLevel: types.TxRiskHigh,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Smurfing detection - coordinated transactions across multiple accounts
	k.monitoringRules["smurfing"] = &types.TransactionMonitoringRule{
		ID:          "smurfing",
		Name:        "Smurfing Detection",
		Description: "Detects coordinated transactions across multiple accounts",
		RuleType:    "smurfing",
		Parameters: map[string]string{
			"min_accounts":         "3",
			"time_window":          "1h",
			"similarity_threshold": "0.8",
		},
		RiskLevel: types.TxRiskCritical,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Round amount detection
	k.monitoringRules["round_amounts"] = &types.TransactionMonitoringRule{
		ID:          "round_amounts",
		Name:        "Round Amount Pattern",
		Description: "Detects suspicious patterns of round number transactions",
		RuleType:    "pattern",
		Parameters: map[string]string{
			"count_threshold": "5",
			"time_window":     "24h",
		},
		RiskLevel: types.TxRiskMedium,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Rapid movement detection
	k.monitoringRules["rapid_movement"] = &types.TransactionMonitoringRule{
		ID:          "rapid_movement",
		Name:        "Rapid Fund Movement",
		Description: "Detects funds moving rapidly through multiple accounts",
		RuleType:    "chain",
		Parameters: map[string]string{
			"min_hops":   "3",
			"max_time":   "1h",
			"min_amount": "1000",
		},
		RiskLevel: types.TxRiskHigh,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Geographic risk
	k.monitoringRules["high_risk_jurisdiction"] = &types.TransactionMonitoringRule{
		ID:          "high_risk_jurisdiction",
		Name:        "High-Risk Jurisdiction",
		Description: "Flags transactions involving high-risk jurisdictions",
		RuleType:    "jurisdiction",
		Parameters: map[string]string{
			"risk_countries": "KP,IR,SY", // Example: North Korea, Iran, Syria
		},
		RiskLevel: types.TxRiskCritical,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

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
