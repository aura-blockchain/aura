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
