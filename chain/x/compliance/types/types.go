package types

import (
	"fmt"
	"time"
)

// KYCLevel defines the level of KYC verification
type KYCLevel int

const (
	KYCLevelUnspecified KYCLevel = iota
	KYCLevelNone
	KYCLevelBasic
	KYCLevelIntermediate
	KYCLevelAdvanced
)

// AMLRiskLevel defines the AML risk classification
type AMLRiskLevel int

const (
	AMLRiskUnspecified AMLRiskLevel = iota
	AMLRiskLow
	AMLRiskMedium
	AMLRiskHigh
	AMLRiskSevere
)

// TransactionRiskLevel defines transaction risk classification
type TransactionRiskLevel int

const (
	TxRiskUnspecified TransactionRiskLevel = iota
	TxRiskLow
	TxRiskMedium
	TxRiskHigh
	TxRiskCritical
)

// SanctionsStatus defines sanctions screening status
type SanctionsStatus int

const (
	SanctionsUnspecified SanctionsStatus = iota
	SanctionsClear
	SanctionsPotentialMatch
	SanctionsConfirmed
	SanctionsPendingReview
)

// KYCRecord represents a KYC verification record
type KYCRecord struct {
	Address              string
	KYCLevel             KYCLevel
	Provider             string
	VerifiedAt           time.Time
	ExpiresAt            time.Time
	VerificationID       string
	Documents            []string
	Jurisdiction         string
	EnhancedDueDiligence bool
	RiskScore            string
}

// Validate performs basic validation on KYCRecord
func (k *KYCRecord) Validate() error {
	if k.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if k.KYCLevel == KYCLevelUnspecified {
		return fmt.Errorf("KYC level must be specified")
	}
	if k.Provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	if k.VerificationID == "" {
		return fmt.Errorf("verification ID cannot be empty")
	}
	if k.ExpiresAt.Before(k.VerifiedAt) {
		return fmt.Errorf("expiry time cannot be before verification time")
	}
	return nil
}

// IsExpired checks if the KYC record has expired
func (k *KYCRecord) IsExpired() bool {
	return time.Now().After(k.ExpiresAt)
}

// AMLProfile represents an AML risk profile
type AMLProfile struct {
	Address              string
	RiskLevel            AMLRiskLevel
	RiskFactors          []string
	LastAssessment       time.Time
	TotalTransactions    uint64
	TotalVolume          string
	SuspiciousActivities []*SuspiciousActivity
	PEPStatus            bool // Politically Exposed Person
	SourceOfFunds        []string
	Occupation           string
}

// Validate performs basic validation on AMLProfile
func (a *AMLProfile) Validate() error {
	if a.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if a.RiskLevel == AMLRiskUnspecified {
		return fmt.Errorf("risk level must be specified")
	}
	return nil
}

// SuspiciousActivity represents a suspicious activity report
type SuspiciousActivity struct {
	ID              string
	Address         string
	TransactionHash string
	ActivityType    string
	Description     string
	Amount          string
	DetectedAt      time.Time
	ReportedAt      time.Time
	FiledSAR        bool
	SARReference    string
	Indicators      []string
}

// Validate performs basic validation on SuspiciousActivity
func (s *SuspiciousActivity) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	if s.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if s.ActivityType == "" {
		return fmt.Errorf("activity type cannot be empty")
	}
	if s.Description == "" {
		return fmt.Errorf("description cannot be empty")
	}
	return nil
}

// TransactionMonitoringRule defines a monitoring rule
type TransactionMonitoringRule struct {
	ID          string
	Name        string
	Description string
	RuleType    string
	Parameters  map[string]string
	RiskLevel   TransactionRiskLevel
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate performs basic validation on TransactionMonitoringRule
func (t *TransactionMonitoringRule) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	if t.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if t.RuleType == "" {
		return fmt.Errorf("rule type cannot be empty")
	}
	if t.RiskLevel == TxRiskUnspecified {
		return fmt.Errorf("risk level must be specified")
	}
	return nil
}

// TransactionAlert represents a monitoring alert
type TransactionAlert struct {
	ID              string
	TransactionHash string
	Address         string
	RuleID          string
	RiskLevel       TransactionRiskLevel
	Description     string
	TriggeredAt     time.Time
	Reviewed        bool
	ReviewedAt      time.Time
	Reviewer        string
	Resolution      string
	Notes           string
}

// Validate performs basic validation on TransactionAlert
func (t *TransactionAlert) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	if t.TransactionHash == "" {
		return fmt.Errorf("transaction hash cannot be empty")
	}
	if t.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if t.RuleID == "" {
		return fmt.Errorf("rule ID cannot be empty")
	}
	return nil
}

// SanctionsScreeningResult represents a sanctions check result
type SanctionsScreeningResult struct {
	Address              string
	Status               SanctionsStatus
	Matches              []*SanctionsMatch
	ScreenedAt           time.Time
	ScreeningProvider    string
	RequiresManualReview bool
	ReviewedAt           time.Time
	Reviewer             string
	ReviewDecision       string
}

// Validate performs basic validation on SanctionsScreeningResult
func (s *SanctionsScreeningResult) Validate() error {
	if s.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if s.Status == SanctionsUnspecified {
		return fmt.Errorf("status must be specified")
	}
	if s.ScreeningProvider == "" {
		return fmt.Errorf("screening provider cannot be empty")
	}
	return nil
}

// SanctionsMatch represents a potential sanctions list match
type SanctionsMatch struct {
	ListName    string
	MatchScore  string
	MatchedName string
	MatchedID   string
	Aliases     []string
	Country     string
	Program     string
	Remarks     string
}

// GDPRConsent represents GDPR consent record
type GDPRConsent struct {
	Address            string
	ConsentType        string
	Consented          bool
	ConsentGivenAt     time.Time
	ConsentWithdrawnAt time.Time
	ConsentVersion     string
	IPAddress          string
	UserAgent          string
}

// Validate performs basic validation on GDPRConsent
func (g *GDPRConsent) Validate() error {
	if g.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if g.ConsentType == "" {
		return fmt.Errorf("consent type cannot be empty")
	}
	if g.ConsentVersion == "" {
		return fmt.Errorf("consent version cannot be empty")
	}
	return nil
}

// GDPRDataRequest represents a GDPR data request
type GDPRDataRequest struct {
	ID              string
	Address         string
	RequestType     string
	RequestedAt     time.Time
	CompletedAt     time.Time
	Status          string
	FulfillmentData string
	Notes           string
}

// Validate performs basic validation on GDPRDataRequest
func (g *GDPRDataRequest) Validate() error {
	if g.ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	if g.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if g.RequestType == "" {
		return fmt.Errorf("request type cannot be empty")
	}
	validTypes := map[string]bool{
		"access":        true,
		"rectification": true,
		"erasure":       true,
		"portability":   true,
	}
	if !validTypes[g.RequestType] {
		return fmt.Errorf("invalid request type: %s", g.RequestType)
	}
	return nil
}

// TaxReport represents a tax reporting record
type TaxReport struct {
	ID                 string
	Address            string
	TaxYear            string
	Jurisdiction       string
	ReportType         string
	Transactions       []*TaxTransaction
	TotalIncome        string
	TotalCapitalGains  string
	TotalCapitalLosses string
	GeneratedAt        time.Time
	FilePath           string
	Filed              bool
	FiledAt            time.Time
}

// Validate performs basic validation on TaxReport
func (t *TaxReport) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	if t.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if t.TaxYear == "" {
		return fmt.Errorf("tax year cannot be empty")
	}
	if t.Jurisdiction == "" {
		return fmt.Errorf("jurisdiction cannot be empty")
	}
	if t.ReportType == "" {
		return fmt.Errorf("report type cannot be empty")
	}
	return nil
}

// TaxTransaction represents a taxable transaction
type TaxTransaction struct {
	TransactionHash string
	Timestamp       time.Time
	TransactionType string
	Asset           string
	Amount          string
	CostBasis       string
	FairMarketValue string
	GainLoss        string
	IsIncome        bool
}

// ComplianceParams defines module parameters
type ComplianceParams struct {
	// KYC/AML settings
	KYCRequired     bool
	MinimumKYCLevel KYCLevel
	KYCExpiryDays   uint64

	// Transaction monitoring
	TransactionMonitoringEnabled bool
	VelocityLimit24h             string
	SingleTransactionLimit       string
	StructuringThresholdCount    uint32

	// Sanctions screening
	SanctionsScreeningEnabled bool
	SanctionsList             []string
	ScreeningCacheHours       uint64

	// GDPR
	GDPREnabled        bool
	DataRetentionDays  uint64
	ProcessingPurposes []string

	// Tax reporting
	TaxReportingEnabled bool
	TaxJurisdictions    []string
	TaxYearEnd          string
}

// DefaultParams returns default compliance parameters
func DefaultParams() ComplianceParams {
	return ComplianceParams{
		KYCRequired:                  false,
		MinimumKYCLevel:              KYCLevelBasic,
		KYCExpiryDays:                365,
		TransactionMonitoringEnabled: true,
		VelocityLimit24h:             "1000000", // 1M units
		SingleTransactionLimit:       "100000",  // 100K units
		StructuringThresholdCount:    5,
		SanctionsScreeningEnabled:    true,
		SanctionsList:                []string{"OFAC_SDN", "EU_SANCTIONS", "UN_SANCTIONS"},
		ScreeningCacheHours:          24,
		GDPREnabled:                  true,
		DataRetentionDays:            2555, // ~7 years
		ProcessingPurposes:           []string{"transaction_processing", "compliance", "analytics"},
		TaxReportingEnabled:          true,
		TaxJurisdictions:             []string{"US", "EU"},
		TaxYearEnd:                   "12-31",
	}
}

// Validate performs basic validation on ComplianceParams
func (p *ComplianceParams) Validate() error {
	if p.KYCExpiryDays == 0 {
		return fmt.Errorf("KYC expiry days must be greater than 0")
	}
	if p.StructuringThresholdCount == 0 {
		return fmt.Errorf("structuring threshold count must be greater than 0")
	}
	if p.ScreeningCacheHours == 0 {
		return fmt.Errorf("screening cache hours must be greater than 0")
	}
	if p.DataRetentionDays == 0 {
		return fmt.Errorf("data retention days must be greater than 0")
	}
	return nil
}
