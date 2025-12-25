// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"time"
)

// UnimplementedIncidentResponseServiceServer provides default implementations
type UnimplementedIncidentResponseServiceServer struct{}

func (UnimplementedIncidentResponseServiceServer) ReportIncident(req *ReportIncidentRequest) (*ReportIncidentResponse, error) {
	return nil, nil
}

func (UnimplementedIncidentResponseServiceServer) UpdateIncidentStatus(req *UpdateIncidentStatusRequest) (*UpdateIncidentStatusResponse, error) {
	return nil, nil
}

func (UnimplementedIncidentResponseServiceServer) GetIncident(req *GetIncidentRequest) (*GetIncidentResponse, error) {
	return nil, nil
}

func (UnimplementedIncidentResponseServiceServer) RequestChainPause(req *RequestChainPauseRequest) (*RequestChainPauseResponse, error) {
	return nil, nil
}

func (UnimplementedIncidentResponseServiceServer) ResumeChain(req *ResumeChainRequest) (*ResumeChainResponse, error) {
	return nil, nil
}

func (UnimplementedIncidentResponseServiceServer) GetChainPauseState(req *GetChainPauseStateRequest) (*GetChainPauseStateResponse, error) {
	return nil, nil
}

func (UnimplementedIncidentResponseServiceServer) SetWalletLimits(req *SetWalletLimitsRequest) (*SetWalletLimitsResponse, error) {
	return nil, nil
}

func (UnimplementedIncidentResponseServiceServer) CheckWalletLimit(req *CheckWalletLimitRequest) (*CheckWalletLimitResponse, error) {
	return nil, nil
}

func (UnimplementedIncidentResponseServiceServer) CreatePostMortem(req *CreatePostMortemRequest) (*CreatePostMortemResponse, error) {
	return nil, nil
}

func (UnimplementedIncidentResponseServiceServer) TriggerBackup(req *TriggerBackupRequest) (*TriggerBackupResponse, error) {
	return nil, nil
}

func (UnimplementedIncidentResponseServiceServer) TriggerInsuranceClaim(req *TriggerInsuranceClaimRequest) (*TriggerInsuranceClaimResponse, error) {
	return nil, nil
}

// RegisterIncidentResponseServiceServer registers the service
func RegisterIncidentResponseServiceServer(s interface{}, srv interface{}) {
	// In a real implementation, this would register with gRPC
}

// Request/Response types

type ReportIncidentRequest struct {
	Title           string
	Description     string
	Severity        IncidentSeverity
	ReportedBy      string
	AffectedSystems []string
}

type ReportIncidentResponse struct {
	IncidentId string
}

type UpdateIncidentStatusRequest struct {
	IncidentId string
	Status     IncidentStatus
	UpdatedBy  string
	Notes      string
}

type UpdateIncidentStatusResponse struct {
	Success bool
}

type GetIncidentRequest struct {
	IncidentId string
}

type GetIncidentResponse struct {
	Incident *Incident
}

type RequestChainPauseRequest struct {
	Requester  string
	PauseLevel PauseLevel
	Reason     string
	IncidentId string
	Duration   time.Duration
}

type RequestChainPauseResponse struct {
	Success bool
}

type ResumeChainRequest struct {
	ResumedBy string
	Reason    string
}

type ResumeChainResponse struct {
	Success bool
}

type GetChainPauseStateRequest struct{}

type GetChainPauseStateResponse struct {
	PauseState *ChainPauseState
}

type SetWalletLimitsRequest struct {
	Address            string
	MaxBalance         string
	MaxTransactionSize string
	DailyLimit         string
}

type SetWalletLimitsResponse struct {
	Success bool
}

type CheckWalletLimitRequest struct {
	Address        string
	Amount         string
	CurrentBalance string
}

type CheckWalletLimitResponse struct {
	Allowed bool
	Reason  string
}

type CreatePostMortemRequest struct {
	IncidentId     string
	CreatedBy      string
	Summary        string
	RootCause      string
	Impact         string
	Resolution     string
	LessonsLearned []string
	ActionItems    []ActionItem
}

type CreatePostMortemResponse struct {
	Success bool
}

type TriggerBackupRequest struct {
	BackupType  string
	TriggeredBy string
}

type TriggerBackupResponse struct {
	BackupId string
}

type TriggerInsuranceClaimRequest struct {
	IncidentId string
	Amount     string
	Signers    []string
}

type TriggerInsuranceClaimResponse struct {
	ClaimId string
}
