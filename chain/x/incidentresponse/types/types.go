// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
	"time"
)

// IncidentSeverity represents the severity level of an incident
type IncidentSeverity string

const (
	SeverityLow      IncidentSeverity = "low"
	SeverityMedium   IncidentSeverity = "medium"
	SeverityHigh     IncidentSeverity = "high"
	SeverityCritical IncidentSeverity = "critical"
)

// IncidentStatus represents the status of an incident
type IncidentStatus string

const (
	StatusNew           IncidentStatus = "new"
	StatusInvestigation IncidentStatus = "investigating"
	StatusContained     IncidentStatus = "contained"
	StatusResolved      IncidentStatus = "resolved"
	StatusPostMortem    IncidentStatus = "post_mortem"
	StatusClosed        IncidentStatus = "closed"
)

// PauseLevel represents different levels of chain pause
type PauseLevel string

const (
	PauseLevelNone         PauseLevel = "none"
	PauseLevelTransactions PauseLevel = "transactions"
	PauseLevelModules      PauseLevel = "modules"
	PauseLevelFull         PauseLevel = "full"
)

// Incident represents a security incident
type Incident struct {
	ID              string                  `json:"id"`
	Title           string                  `json:"title"`
	Description     string                  `json:"description"`
	Severity        IncidentSeverity        `json:"severity"`
	Status          IncidentStatus          `json:"status"`
	ReportedBy      string                  `json:"reported_by"`
	ReportedAt      time.Time               `json:"reported_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	ResolvedAt      time.Time               `json:"resolved_at,omitempty"`
	AffectedSystems []string                `json:"affected_systems"`
	ResponseTeam    []string                `json:"response_team"`
	Timeline        []IncidentTimelineEntry `json:"timeline"`
	PostMortem      *PostMortem             `json:"post_mortem,omitempty"`
}

// IncidentTimelineEntry represents an entry in the incident timeline
type IncidentTimelineEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Actor       string    `json:"actor"`
}

// PostMortem represents a post-incident analysis
type PostMortem struct {
	CreatedAt      time.Time    `json:"created_at"`
	CreatedBy      string       `json:"created_by"`
	Summary        string       `json:"summary"`
	RootCause      string       `json:"root_cause"`
	Impact         string       `json:"impact"`
	Resolution     string       `json:"resolution"`
	LessonsLearned []string     `json:"lessons_learned"`
	ActionItems    []ActionItem `json:"action_items"`
	Timeline       string       `json:"timeline"`
}

// ActionItem represents a follow-up action from a post-mortem
type ActionItem struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Assignee    string    `json:"assignee"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	DueDate     time.Time `json:"due_date"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// ChainPauseState represents the pause state of the chain
type ChainPauseState struct {
	IsPaused        bool       `json:"is_paused"`
	PauseLevel      PauseLevel `json:"pause_level"`
	PausedAt        time.Time  `json:"paused_at,omitempty"`
	PausedBy        string     `json:"paused_by,omitempty"`
	Reason          string     `json:"reason,omitempty"`
	IncidentID      string     `json:"incident_id,omitempty"`
	PausedModules   []string   `json:"paused_modules,omitempty"`
	EstimatedResume time.Time  `json:"estimated_resume,omitempty"`
}

// WalletLimits represents security limits for hot wallets
type WalletLimits struct {
	Address            string    `json:"address"`
	MaxBalance         string    `json:"max_balance"`
	MaxTransactionSize string    `json:"max_transaction_size"`
	DailyLimit         string    `json:"daily_limit"`
	CurrentBalance     string    `json:"current_balance"`
	TodayTransferred   string    `json:"today_transferred"`
	LastReset          time.Time `json:"last_reset"`
}

// ColdStorageConfig represents cold storage configuration
type ColdStorageConfig struct {
	Enabled           bool      `json:"enabled"`
	MultiSigThreshold uint32    `json:"multisig_threshold"`
	MultiSigSigners   []string  `json:"multisig_signers"`
	TimeLockedUntil   time.Time `json:"time_locked_until,omitempty"`
	MinimumBalance    string    `json:"minimum_balance"`
	MaxHotWalletRatio float64   `json:"max_hot_wallet_ratio"`
}

// BackupValidatorConfig represents backup validator infrastructure
type BackupValidatorConfig struct {
	Enabled           bool          `json:"enabled"`
	PrimaryValidators []string      `json:"primary_validators"`
	BackupValidators  []string      `json:"backup_validators"`
	AutoFailover      bool          `json:"auto_failover"`
	FailoverThreshold int           `json:"failover_threshold"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	LastHealthCheck   time.Time     `json:"last_health_check"`
}

// CommunicationPlan represents incident communication settings
type CommunicationPlan struct {
	Enabled              bool          `json:"enabled"`
	NotificationChannels []string      `json:"notification_channels"`
	EscalationContacts   []Contact     `json:"escalation_contacts"`
	StatusPageURL        string        `json:"status_page_url"`
	UpdateInterval       time.Duration `json:"update_interval"`
}

// Contact represents an emergency contact
type Contact struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Telegram string `json:"telegram,omitempty"`
	Priority int    `json:"priority"`
}

// DisasterRecoveryPlan represents disaster recovery configuration
type DisasterRecoveryPlan struct {
	Enabled           bool          `json:"enabled"`
	BackupInterval    time.Duration `json:"backup_interval"`
	BackupLocations   []string      `json:"backup_locations"`
	LastBackupTime    time.Time     `json:"last_backup_time"`
	RPO               time.Duration `json:"rpo"` // Recovery Point Objective
	RTO               time.Duration `json:"rto"` // Recovery Time Objective
	SnapshotRetention int           `json:"snapshot_retention"`
	ValidatorBackups  bool          `json:"validator_backups"`
	StateBackups      bool          `json:"state_backups"`
	KeyBackups        bool          `json:"key_backups"`
}

// InsuranceIntegration represents insurance coverage integration
type InsuranceIntegration struct {
	Enabled         bool     `json:"enabled"`
	Provider        string   `json:"provider"`
	PolicyNumber    string   `json:"policy_number"`
	CoverageAmount  string   `json:"coverage_amount"`
	ClaimEndpoint   string   `json:"claim_endpoint"`
	RequiredSigners []string `json:"required_signers"`
	AutoClaim       bool     `json:"auto_claim"`
	ClaimThreshold  string   `json:"claim_threshold"`
}

// IncidentResponseParams represents the module parameters
type IncidentResponseParams struct {
	// Emergency pause settings
	EmergencyPauseEnabled bool          `json:"emergency_pause_enabled"`
	PauseAuthorizedKeys   []string      `json:"pause_authorized_keys"`
	PauseRequiredSigners  uint32        `json:"pause_required_signers"`
	MaxPauseDuration      time.Duration `json:"max_pause_duration"`

	// Hot wallet limits
	HotWalletLimitsEnabled bool   `json:"hot_wallet_limits_enabled"`
	GlobalMaxHotWallet     string `json:"global_max_hot_wallet"`
	GlobalDailyLimit       string `json:"global_daily_limit"`

	// Cold storage
	ColdStorage ColdStorageConfig `json:"cold_storage"`

	// Backup validators
	BackupValidators BackupValidatorConfig `json:"backup_validators"`

	// Communication plan
	Communication CommunicationPlan `json:"communication"`

	// Disaster recovery
	DisasterRecovery DisasterRecoveryPlan `json:"disaster_recovery"`

	// Insurance
	Insurance InsuranceIntegration `json:"insurance"`

	// Incident response team
	IncidentResponseTeam []string `json:"incident_response_team"`
}

// DefaultParams returns default parameters
func DefaultParams() IncidentResponseParams {
	return IncidentResponseParams{
		EmergencyPauseEnabled:  true,
		PauseAuthorizedKeys:    []string{"admin1", "admin2", "admin3"}, // Default authorized keys
		PauseRequiredSigners:   3,
		MaxPauseDuration:       24 * time.Hour,
		HotWalletLimitsEnabled: true,
		GlobalMaxHotWallet:     "10000000000", // 10B tokens
		GlobalDailyLimit:       "1000000000",  // 1B tokens per day
		ColdStorage: ColdStorageConfig{
			Enabled:           true,
			MultiSigThreshold: 5,
			MultiSigSigners:   []string{"signer1", "signer2", "signer3", "signer4", "signer5"}, // Default signers
			MinimumBalance:    "50000000000",                                                   // 50B tokens
			MaxHotWalletRatio: 0.20,                                                            // 20% max in hot wallets
		},
		BackupValidators: BackupValidatorConfig{
			Enabled:           true,
			AutoFailover:      true,
			FailoverThreshold: 3,
			HeartbeatInterval: 30 * time.Second,
		},
		Communication: CommunicationPlan{
			Enabled:        true,
			UpdateInterval: 30 * time.Minute,
		},
		DisasterRecovery: DisasterRecoveryPlan{
			Enabled:           true,
			BackupInterval:    6 * time.Hour,
			BackupLocations:   []string{"s3://default-backup"}, // Default backup location
			RPO:               15 * time.Minute,
			RTO:               2 * time.Hour,
			SnapshotRetention: 7,
			ValidatorBackups:  true,
			StateBackups:      true,
			KeyBackups:        false, // Keys should be backed up manually
		},
		Insurance: InsuranceIntegration{
			Enabled:        false,
			AutoClaim:      false,
			ClaimThreshold: "1000000000000", // 1T tokens
		},
		IncidentResponseTeam: []string{},
	}
}

// ValidateBasic validates the parameters
func (p IncidentResponseParams) ValidateBasic() error {
	if p.EmergencyPauseEnabled {
		if len(p.PauseAuthorizedKeys) == 0 {
			return fmt.Errorf("pause authorized keys cannot be empty when emergency pause is enabled")
		}
		if p.PauseRequiredSigners == 0 {
			return fmt.Errorf("pause required signers must be greater than 0")
		}
		if p.PauseRequiredSigners > uint32(len(p.PauseAuthorizedKeys)) {
			return fmt.Errorf("pause required signers cannot exceed number of authorized keys")
		}
	}

	if p.ColdStorage.Enabled {
		if p.ColdStorage.MultiSigThreshold == 0 {
			return fmt.Errorf("multisig threshold must be greater than 0")
		}
		if p.ColdStorage.MaxHotWalletRatio < 0 || p.ColdStorage.MaxHotWalletRatio > 1 {
			return fmt.Errorf("max hot wallet ratio must be between 0 and 1")
		}
	}

	if p.DisasterRecovery.Enabled {
		if p.DisasterRecovery.BackupInterval == 0 {
			return fmt.Errorf("backup interval must be greater than 0")
		}
		if len(p.DisasterRecovery.BackupLocations) == 0 {
			return fmt.Errorf("at least one backup location must be specified")
		}
	}

	return nil
}

// Errors
var (
	ErrIncidentNotFound         = fmt.Errorf("incident not found")
	ErrUnauthorizedPause        = fmt.Errorf("unauthorized to pause chain")
	ErrChainAlreadyPaused       = fmt.Errorf("chain is already paused")
	ErrChainNotPaused           = fmt.Errorf("chain is not paused")
	ErrWalletLimitExceeded      = fmt.Errorf("wallet limit exceeded")
	ErrInsufficientSigners      = fmt.Errorf("insufficient signers")
	ErrInvalidPauseLevel        = fmt.Errorf("invalid pause level")
	ErrMaxPauseDurationExceeded = fmt.Errorf("max pause duration exceeded")
	ErrPostMortemNotCompleted   = fmt.Errorf("post mortem not completed")
)
