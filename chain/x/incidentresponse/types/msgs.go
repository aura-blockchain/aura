// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
)

// Message types for incident response module

// MsgServer interface definition
type MsgServer interface {
	ReportIncident(ctx interface{}, msg *MsgReportIncident) (*MsgReportIncidentResponse, error)
	UpdateIncidentStatus(ctx interface{}, msg *MsgUpdateIncidentStatus) (*MsgUpdateIncidentStatusResponse, error)
	RequestChainPause(ctx interface{}, msg *MsgRequestChainPause) (*MsgRequestChainPauseResponse, error)
	ResumeChain(ctx interface{}, msg *MsgResumeChain) (*MsgResumeChainResponse, error)
	SetWalletLimits(ctx interface{}, msg *MsgSetWalletLimits) (*MsgSetWalletLimitsResponse, error)
	CreatePostMortem(ctx interface{}, msg *MsgCreatePostMortem) (*MsgCreatePostMortemResponse, error)
	CloseIncident(ctx interface{}, msg *MsgCloseIncident) (*MsgCloseIncidentResponse, error)
	TriggerBackup(ctx interface{}, msg *MsgTriggerBackup) (*MsgTriggerBackupResponse, error)
	TriggerInsuranceClaim(ctx interface{}, msg *MsgTriggerInsuranceClaim) (*MsgTriggerInsuranceClaimResponse, error)
}

// QueryServer interface definition
type QueryServer interface {
	GetIncident(ctx interface{}, req *QueryGetIncidentRequest) (*QueryGetIncidentResponse, error)
	GetAllIncidents(ctx interface{}, req *QueryGetAllIncidentsRequest) (*QueryGetAllIncidentsResponse, error)
	GetChainPauseState(ctx interface{}, req *QueryGetChainPauseStateRequest) (*QueryGetChainPauseStateResponse, error)
	GetWalletLimits(ctx interface{}, req *QueryGetWalletLimitsRequest) (*QueryGetWalletLimitsResponse, error)
	GetParams(ctx interface{}, req *QueryGetParamsRequest) (*QueryGetParamsResponse, error)
	GetColdStorageConfig(ctx interface{}, req *QueryGetColdStorageConfigRequest) (*QueryGetColdStorageConfigResponse, error)
	GetBackupValidatorConfig(ctx interface{}, req *QueryGetBackupValidatorConfigRequest) (*QueryGetBackupValidatorConfigResponse, error)
	GetDisasterRecoveryPlan(ctx interface{}, req *QueryGetDisasterRecoveryPlanRequest) (*QueryGetDisasterRecoveryPlanResponse, error)
	GetCommunicationPlan(ctx interface{}, req *QueryGetCommunicationPlanRequest) (*QueryGetCommunicationPlanResponse, error)
	GetInsuranceIntegration(ctx interface{}, req *QueryGetInsuranceIntegrationRequest) (*QueryGetInsuranceIntegrationResponse, error)
}

// MsgReportIncident represents a request to report a new incident
type MsgReportIncident struct {
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Severity        string   `json:"severity"`
	Reporter        string   `json:"reporter"`
	AffectedSystems []string `json:"affected_systems"`
}

func (m *MsgReportIncident) ValidateBasic() error {
	if m.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if m.Description == "" {
		return fmt.Errorf("description cannot be empty")
	}
	if m.Reporter == "" {
		return fmt.Errorf("reporter cannot be empty")
	}
	if m.Severity == "" {
		return fmt.Errorf("severity cannot be empty")
	}
	return nil
}

type MsgReportIncidentResponse struct {
	IncidentId string `json:"incident_id"`
}

// MsgUpdateIncidentStatus represents a request to update incident status
type MsgUpdateIncidentStatus struct {
	IncidentId string `json:"incident_id"`
	Status     string `json:"status"`
	UpdatedBy  string `json:"updated_by"`
	Notes      string `json:"notes"`
}

func (m *MsgUpdateIncidentStatus) ValidateBasic() error {
	if m.IncidentId == "" {
		return fmt.Errorf("incident_id cannot be empty")
	}
	if m.Status == "" {
		return fmt.Errorf("status cannot be empty")
	}
	if m.UpdatedBy == "" {
		return fmt.Errorf("updated_by cannot be empty")
	}
	return nil
}

type MsgUpdateIncidentStatusResponse struct{}

// MsgRequestChainPause represents a request to pause the chain
type MsgRequestChainPause struct {
	Requester       string `json:"requester"`
	PauseLevel      string `json:"pause_level"`
	Reason          string `json:"reason"`
	IncidentId      string `json:"incident_id"`
	DurationSeconds int64  `json:"duration_seconds"`
}

func (m *MsgRequestChainPause) ValidateBasic() error {
	if m.Requester == "" {
		return fmt.Errorf("requester cannot be empty")
	}
	if m.PauseLevel == "" {
		return fmt.Errorf("pause_level cannot be empty")
	}
	if m.Reason == "" {
		return fmt.Errorf("reason cannot be empty")
	}
	if m.DurationSeconds <= 0 {
		return fmt.Errorf("duration_seconds must be positive")
	}
	return nil
}

type MsgRequestChainPauseResponse struct{}

// MsgResumeChain represents a request to resume the chain
type MsgResumeChain struct {
	Resumer string `json:"resumer"`
	Reason  string `json:"reason"`
}

func (m *MsgResumeChain) ValidateBasic() error {
	if m.Resumer == "" {
		return fmt.Errorf("resumer cannot be empty")
	}
	if m.Reason == "" {
		return fmt.Errorf("reason cannot be empty")
	}
	return nil
}

type MsgResumeChainResponse struct{}

// MsgSetWalletLimits represents a request to set wallet limits
type MsgSetWalletLimits struct {
	Address            string `json:"address"`
	MaxBalance         string `json:"max_balance"`
	MaxTransactionSize string `json:"max_transaction_size"`
	DailyLimit         string `json:"daily_limit"`
}

func (m *MsgSetWalletLimits) ValidateBasic() error {
	if m.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if m.MaxBalance == "" {
		return fmt.Errorf("max_balance cannot be empty")
	}
	if m.MaxTransactionSize == "" {
		return fmt.Errorf("max_transaction_size cannot be empty")
	}
	if m.DailyLimit == "" {
		return fmt.Errorf("daily_limit cannot be empty")
	}
	return nil
}

type MsgSetWalletLimitsResponse struct{}

// MsgCreatePostMortem represents a request to create a post-mortem
type MsgCreatePostMortem struct {
	IncidentId     string       `json:"incident_id"`
	Creator        string       `json:"creator"`
	Summary        string       `json:"summary"`
	RootCause      string       `json:"root_cause"`
	Impact         string       `json:"impact"`
	Resolution     string       `json:"resolution"`
	LessonsLearned []string     `json:"lessons_learned"`
	ActionItems    []ActionItem `json:"action_items"`
}

func (m *MsgCreatePostMortem) ValidateBasic() error {
	if m.IncidentId == "" {
		return fmt.Errorf("incident_id cannot be empty")
	}
	if m.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}
	if m.Summary == "" {
		return fmt.Errorf("summary cannot be empty")
	}
	if m.RootCause == "" {
		return fmt.Errorf("root_cause cannot be empty")
	}
	if m.Impact == "" {
		return fmt.Errorf("impact cannot be empty")
	}
	if m.Resolution == "" {
		return fmt.Errorf("resolution cannot be empty")
	}
	return nil
}

type MsgCreatePostMortemResponse struct{}

// MsgCloseIncident represents a request to close an incident
type MsgCloseIncident struct {
	IncidentId string `json:"incident_id"`
	Closer     string `json:"closer"`
}

func (m *MsgCloseIncident) ValidateBasic() error {
	if m.IncidentId == "" {
		return fmt.Errorf("incident_id cannot be empty")
	}
	if m.Closer == "" {
		return fmt.Errorf("closer cannot be empty")
	}
	return nil
}

type MsgCloseIncidentResponse struct{}

// MsgTriggerBackup represents a request to trigger a backup
type MsgTriggerBackup struct {
	BackupType string `json:"backup_type"`
	Requester  string `json:"requester"`
}

func (m *MsgTriggerBackup) ValidateBasic() error {
	if m.BackupType == "" {
		return fmt.Errorf("backup_type cannot be empty")
	}
	if m.Requester == "" {
		return fmt.Errorf("requester cannot be empty")
	}
	return nil
}

type MsgTriggerBackupResponse struct {
	BackupId string `json:"backup_id"`
}

// MsgTriggerInsuranceClaim represents a request to trigger an insurance claim
type MsgTriggerInsuranceClaim struct {
	IncidentId string   `json:"incident_id"`
	Amount     string   `json:"amount"`
	Signers    []string `json:"signers"`
}

func (m *MsgTriggerInsuranceClaim) ValidateBasic() error {
	if m.IncidentId == "" {
		return fmt.Errorf("incident_id cannot be empty")
	}
	if m.Amount == "" {
		return fmt.Errorf("amount cannot be empty")
	}
	if len(m.Signers) == 0 {
		return fmt.Errorf("signers cannot be empty")
	}
	return nil
}

type MsgTriggerInsuranceClaimResponse struct {
	ClaimId string `json:"claim_id"`
}

// Query request/response types

type QueryGetIncidentRequest struct {
	IncidentId string `json:"incident_id"`
}

type QueryGetIncidentResponse struct {
	Incident *Incident `json:"incident"`
}

type QueryGetAllIncidentsRequest struct {
	Status   string `json:"status"`
	Severity string `json:"severity"`
}

type QueryGetAllIncidentsResponse struct {
	Incidents []*Incident `json:"incidents"`
}

type QueryGetChainPauseStateRequest struct{}

type QueryGetChainPauseStateResponse struct {
	PauseState *ChainPauseState `json:"pause_state"`
}

type QueryGetWalletLimitsRequest struct {
	Address string `json:"address"`
}

type QueryGetWalletLimitsResponse struct {
	Limits *WalletLimits `json:"limits"`
}

type QueryGetParamsRequest struct{}

type QueryGetParamsResponse struct {
	Params *IncidentResponseParams `json:"params"`
}

type QueryGetColdStorageConfigRequest struct{}

type QueryGetColdStorageConfigResponse struct {
	Config *ColdStorageConfig `json:"config"`
}

type QueryGetBackupValidatorConfigRequest struct{}

type QueryGetBackupValidatorConfigResponse struct {
	Config *BackupValidatorConfig `json:"config"`
}

type QueryGetDisasterRecoveryPlanRequest struct{}

type QueryGetDisasterRecoveryPlanResponse struct {
	Plan *DisasterRecoveryPlan `json:"plan"`
}

type QueryGetCommunicationPlanRequest struct{}

type QueryGetCommunicationPlanResponse struct {
	Plan *CommunicationPlan `json:"plan"`
}

type QueryGetInsuranceIntegrationRequest struct{}

type QueryGetInsuranceIntegrationResponse struct {
	Integration *InsuranceIntegration `json:"integration"`
}
