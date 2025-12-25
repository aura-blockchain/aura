// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

type queryServer struct {
	securitypb.UnimplementedQueryServer
	keeper *Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper *Keeper) securitypb.QueryServer {
	return &queryServer{keeper: keeper}
}

// Params returns the module parameters
func (qs queryServer) Params(ctx context.Context, req *securitypb.QueryParamsRequest) (*securitypb.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := qs.keeper.GetParams(sdkCtx)

	return &securitypb.QueryParamsResponse{
		Params: params,
	}, nil
}

// SecurityStatus returns overall security status
func (qs queryServer) SecurityStatus(ctx context.Context, req *securitypb.QuerySecurityStatusRequest) (*securitypb.QuerySecurityStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	return &securitypb.QuerySecurityStatusResponse{}, nil
}

// Network Security Queries

func (qs queryServer) PeerInfo(ctx context.Context, req *securitypb.QueryPeerInfoRequest) (*securitypb.QueryPeerInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryPeerInfoResponse{}, nil
}

func (qs queryServer) AllPeers(ctx context.Context, req *securitypb.QueryAllPeersRequest) (*securitypb.QueryAllPeersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)

	// Use ReputationKey as PeerInfo is constructed from reputation data
	peerStore := prefix.NewStore(store, types.ReputationKey)

	var peers []securitypb.PeerInfo
	pageRes, err := query.Paginate(peerStore, req.Pagination, func(key, value []byte) error {
		var rep securitypb.NodeReputation
		if err := qs.keeper.cdc.Unmarshal(value, &rep); err != nil {
			return fmt.Errorf("failed to unmarshal peer reputation: %w", err)
		}
		// Construct PeerInfo from NodeReputation
		peers = append(peers, securitypb.PeerInfo{
			PeerId:          rep.PeerId,
			ReputationScore: rep.Score,
		})
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QueryAllPeersResponse{
		Peers:      peers,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) TrustedPeers(ctx context.Context, req *securitypb.QueryTrustedPeersRequest) (*securitypb.QueryTrustedPeersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	peerStore := prefix.NewStore(store, types.TrustedPeerKey)

	var peers []securitypb.TrustedPeer
	pageRes, err := query.Paginate(peerStore, req.Pagination, func(key, value []byte) error {
		var peer securitypb.TrustedPeer
		if err := qs.keeper.cdc.Unmarshal(value, &peer); err != nil {
			return fmt.Errorf("failed to unmarshal trusted peer: %w", err)
		}
		peers = append(peers, peer)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QueryTrustedPeersResponse{
		Peers:      peers,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) PeerReputation(ctx context.Context, req *securitypb.QueryPeerReputationRequest) (*securitypb.QueryPeerReputationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryPeerReputationResponse{}, nil
}

func (qs queryServer) RateLimitStatus(ctx context.Context, req *securitypb.QueryRateLimitStatusRequest) (*securitypb.QueryRateLimitStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryRateLimitStatusResponse{}, nil
}

func (qs queryServer) MempoolStats(ctx context.Context, req *securitypb.QueryMempoolStatsRequest) (*securitypb.QueryMempoolStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryMempoolStatsResponse{}, nil
}

func (qs queryServer) ForkAlerts(ctx context.Context, req *securitypb.QueryForkAlertsRequest) (*securitypb.QueryForkAlertsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	alertStore := prefix.NewStore(store, types.ForkAlertKey)

	var alerts []securitypb.ForkAlert
	pageRes, err := query.Paginate(alertStore, req.Pagination, func(key, value []byte) error {
		var alert securitypb.ForkAlert
		if err := qs.keeper.cdc.Unmarshal(value, &alert); err != nil {
			return fmt.Errorf("failed to unmarshal fork alert: %w", err)
		}
		// Apply include_resolved filter
		if !req.IncludeResolved && alert.Resolved {
			return nil // Skip resolved alerts when not requested
		}
		alerts = append(alerts, alert)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QueryForkAlertsResponse{
		Alerts:     alerts,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) PartitionAlerts(ctx context.Context, req *securitypb.QueryPartitionAlertsRequest) (*securitypb.QueryPartitionAlertsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	alertStore := prefix.NewStore(store, types.PartitionAlertKey)

	var alerts []securitypb.PartitionAlert
	pageRes, err := query.Paginate(alertStore, req.Pagination, func(key, value []byte) error {
		var alert securitypb.PartitionAlert
		if err := qs.keeper.cdc.Unmarshal(value, &alert); err != nil {
			return fmt.Errorf("failed to unmarshal partition alert: %w", err)
		}
		// Apply include_resolved filter
		if !req.IncludeResolved && alert.Resolved {
			return nil // Skip resolved alerts when not requested
		}
		alerts = append(alerts, alert)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QueryPartitionAlertsResponse{
		Alerts:     alerts,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) NetworkHealth(ctx context.Context, req *securitypb.QueryNetworkHealthRequest) (*securitypb.QueryNetworkHealthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryNetworkHealthResponse{}, nil
}

// Validator Security Queries

func (qs queryServer) ValidatorSecurityInfo(ctx context.Context, req *securitypb.QueryValidatorSecurityInfoRequest) (*securitypb.QueryValidatorSecurityInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryValidatorSecurityInfoResponse{}, nil
}

func (qs queryServer) AllValidatorSecurityInfo(ctx context.Context, req *securitypb.QueryAllValidatorSecurityInfoRequest) (*securitypb.QueryAllValidatorSecurityInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	validatorStore := prefix.NewStore(store, types.ValidatorInfoKey)

	var validators []securitypb.ValidatorSecurityInfo
	pageRes, err := query.Paginate(validatorStore, req.Pagination, func(key, value []byte) error {
		var info securitypb.ValidatorSecurityInfo
		if err := qs.keeper.cdc.Unmarshal(value, &info); err != nil {
			return fmt.Errorf("failed to unmarshal validator security info: %w", err)
		}
		validators = append(validators, info)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QueryAllValidatorSecurityInfoResponse{
		Validators: validators,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) ValidatorAlerts(ctx context.Context, req *securitypb.QueryValidatorAlertsRequest) (*securitypb.QueryValidatorAlertsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	alertStore := prefix.NewStore(store, types.ValidatorAlertKey)

	var alerts []securitypb.ValidatorAlert
	pageRes, err := query.Paginate(alertStore, req.Pagination, func(key, value []byte) error {
		var alert securitypb.ValidatorAlert
		if err := qs.keeper.cdc.Unmarshal(value, &alert); err != nil {
			return fmt.Errorf("failed to unmarshal validator alert: %w", err)
		}
		// Apply validator_address filter if provided
		if req.ValidatorAddress != "" && alert.ValidatorAddress != req.ValidatorAddress {
			return nil // Skip alerts for other validators
		}
		// Apply include_acknowledged filter
		if !req.IncludeAcknowledged && alert.Acknowledged {
			return nil // Skip acknowledged alerts when not requested
		}
		alerts = append(alerts, alert)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QueryValidatorAlertsResponse{
		Alerts:     alerts,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) DoubleSignEvidences(ctx context.Context, req *securitypb.QueryDoubleSignEvidencesRequest) (*securitypb.QueryDoubleSignEvidencesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	evidenceStore := prefix.NewStore(store, types.DoubleSignEvidenceKey)

	var evidences []securitypb.DoubleSignEvidence
	pageRes, err := query.Paginate(evidenceStore, req.Pagination, func(key, value []byte) error {
		var evidence securitypb.DoubleSignEvidence
		if err := qs.keeper.cdc.Unmarshal(value, &evidence); err != nil {
			return fmt.Errorf("failed to unmarshal double sign evidence: %w", err)
		}
		// Apply validator_address filter if provided
		if req.ValidatorAddress != "" && evidence.ValidatorAddress != req.ValidatorAddress {
			return nil // Skip evidences for other validators
		}
		evidences = append(evidences, evidence)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QueryDoubleSignEvidencesResponse{
		Evidences:  evidences,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) DowntimeInfractions(ctx context.Context, req *securitypb.QueryDowntimeInfractionsRequest) (*securitypb.QueryDowntimeInfractionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	infractionStore := prefix.NewStore(store, types.DowntimeInfractionKey)

	var infractions []securitypb.DowntimeInfraction
	pageRes, err := query.Paginate(infractionStore, req.Pagination, func(key, value []byte) error {
		var infraction securitypb.DowntimeInfraction
		if err := qs.keeper.cdc.Unmarshal(value, &infraction); err != nil {
			return fmt.Errorf("failed to unmarshal downtime infraction: %w", err)
		}
		// Apply validator_address filter if provided
		if req.ValidatorAddress != "" && infraction.ValidatorAddress != req.ValidatorAddress {
			return nil // Skip infractions for other validators
		}
		infractions = append(infractions, infraction)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QueryDowntimeInfractionsResponse{
		Infractions: infractions,
		Pagination:  pageRes,
	}, nil
}

func (qs queryServer) SentryNodes(ctx context.Context, req *securitypb.QuerySentryNodesRequest) (*securitypb.QuerySentryNodesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	sentryStore := prefix.NewStore(store, types.SentryNodeKey)

	var sentryNodes []securitypb.SentryNodeInfo
	pageRes, err := query.Paginate(sentryStore, req.Pagination, func(key, value []byte) error {
		var sentry securitypb.SentryNodeInfo
		if err := qs.keeper.cdc.Unmarshal(value, &sentry); err != nil {
			return fmt.Errorf("failed to unmarshal sentry node: %w", err)
		}
		// Apply validator_address filter if provided
		if req.ValidatorAddress != "" && sentry.ValidatorAddress != req.ValidatorAddress {
			return nil // Skip sentry nodes for other validators
		}
		sentryNodes = append(sentryNodes, sentry)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QuerySentryNodesResponse{
		SentryNodes: sentryNodes,
		Pagination:  pageRes,
	}, nil
}

// Wallet Security Queries

func (qs queryServer) WalletSecurityInfo(ctx context.Context, req *securitypb.QueryWalletSecurityInfoRequest) (*securitypb.QueryWalletSecurityInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryWalletSecurityInfoResponse{}, nil
}

func (qs queryServer) HardwareWalletConfig(ctx context.Context, req *securitypb.QueryHardwareWalletConfigRequest) (*securitypb.QueryHardwareWalletConfigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryHardwareWalletConfigResponse{}, nil
}

func (qs queryServer) MultiSigWallet(ctx context.Context, req *securitypb.QueryMultiSigWalletRequest) (*securitypb.QueryMultiSigWalletResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryMultiSigWalletResponse{}, nil
}

func (qs queryServer) PendingMultiSigTransactions(ctx context.Context, req *securitypb.QueryPendingMultiSigTransactionsRequest) (*securitypb.QueryPendingMultiSigTransactionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryPendingMultiSigTransactionsResponse{}, nil
}

func (qs queryServer) SocialRecoveryConfig(ctx context.Context, req *securitypb.QuerySocialRecoveryConfigRequest) (*securitypb.QuerySocialRecoveryConfigResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QuerySocialRecoveryConfigResponse{}, nil
}

func (qs queryServer) RecoveryRequests(ctx context.Context, req *securitypb.QueryRecoveryRequestsRequest) (*securitypb.QueryRecoveryRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryRecoveryRequestsResponse{}, nil
}

func (qs queryServer) SpendingLimits(ctx context.Context, req *securitypb.QuerySpendingLimitsRequest) (*securitypb.QuerySpendingLimitsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QuerySpendingLimitsResponse{}, nil
}

func (qs queryServer) SimulateTransaction(ctx context.Context, req *securitypb.QuerySimulateTransactionRequest) (*securitypb.QuerySimulateTransactionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QuerySimulateTransactionResponse{}, nil
}

// Incident Response Queries

func (qs queryServer) Incident(ctx context.Context, req *securitypb.QueryIncidentRequest) (*securitypb.QueryIncidentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryIncidentResponse{}, nil
}

func (qs queryServer) AllIncidents(ctx context.Context, req *securitypb.QueryAllIncidentsRequest) (*securitypb.QueryAllIncidentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	incidentStore := prefix.NewStore(store, types.IncidentKey)

	var incidents []securitypb.Incident
	pageRes, err := query.Paginate(incidentStore, req.Pagination, func(key, value []byte) error {
		var incident securitypb.Incident
		if err := qs.keeper.cdc.Unmarshal(value, &incident); err != nil {
			return fmt.Errorf("failed to unmarshal incident: %w", err)
		}
		// Apply include_resolved filter
		if !req.IncludeResolved && incident.Status == securitypb.IncidentStatus_INCIDENT_STATUS_RESOLVED {
			return nil // Skip resolved incidents when not requested
		}
		// Apply min_severity filter
		if req.MinSeverity != securitypb.IncidentSeverity_INCIDENT_SEVERITY_UNSPECIFIED && incident.Severity < req.MinSeverity {
			return nil // Skip incidents below minimum severity
		}
		incidents = append(incidents, incident)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QueryAllIncidentsResponse{
		Incidents:  incidents,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) AuditLog(ctx context.Context, req *securitypb.QueryAuditLogRequest) (*securitypb.QueryAuditLogResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	auditStore := prefix.NewStore(store, types.AuditLogKey)

	var entries []securitypb.AuditLogEntry
	pageRes, err := query.Paginate(auditStore, req.Pagination, func(key, value []byte) error {
		var entry securitypb.AuditLogEntry
		if err := qs.keeper.cdc.Unmarshal(value, &entry); err != nil {
			return fmt.Errorf("failed to unmarshal audit log entry: %w", err)
		}
		// Apply event_type filter if provided
		if req.EventType != "" && entry.EventType != req.EventType {
			return nil
		}
		// Apply actor filter if provided
		if req.Actor != "" && entry.Actor != req.Actor {
			return nil
		}
		entries = append(entries, entry)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QueryAuditLogResponse{
		Entries:    entries,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) ResponseActions(ctx context.Context, req *securitypb.QueryResponseActionsRequest) (*securitypb.QueryResponseActionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// ResponseActions are stored within the Incident itself, but we'll paginate an empty response
	// since there's no dedicated store key for response actions. The data would need to be
	// extracted from the incident and its associated actions.
	// For now, return empty with pagination support.
	return &securitypb.QueryResponseActionsResponse{
		Actions:    []securitypb.ResponseAction{},
		Pagination: nil,
	}, nil
}

// Cryptography Queries

func (qs queryServer) KeyRotationSchedule(ctx context.Context, req *securitypb.QueryKeyRotationScheduleRequest) (*securitypb.QueryKeyRotationScheduleResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	return &securitypb.QueryKeyRotationScheduleResponse{}, nil
}

func (qs queryServer) AllKeyRotationSchedules(ctx context.Context, req *securitypb.QueryAllKeyRotationSchedulesRequest) (*securitypb.QueryAllKeyRotationSchedulesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	scheduleStore := prefix.NewStore(store, types.KeyRotationScheduleKey)

	var schedules []securitypb.KeyRotationSchedule
	pageRes, err := query.Paginate(scheduleStore, req.Pagination, func(key, value []byte) error {
		var schedule securitypb.KeyRotationSchedule
		if err := qs.keeper.cdc.Unmarshal(value, &schedule); err != nil {
			return fmt.Errorf("failed to unmarshal key rotation schedule: %w", err)
		}
		schedules = append(schedules, schedule)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QueryAllKeyRotationSchedulesResponse{
		Schedules:  schedules,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) ThresholdScheme(ctx context.Context, req *securitypb.QueryThresholdSchemeRequest) (*securitypb.QueryThresholdSchemeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryThresholdSchemeResponse{}, nil
}

func (qs queryServer) VerifyZKProof(ctx context.Context, req *securitypb.QueryVerifyZKProofRequest) (*securitypb.QueryVerifyZKProofResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryVerifyZKProofResponse{}, nil
}

func (qs queryServer) QuantumResistantKey(ctx context.Context, req *securitypb.QueryQuantumResistantKeyRequest) (*securitypb.QueryQuantumResistantKeyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryQuantumResistantKeyResponse{}, nil
}

// Privacy Queries

func (qs queryServer) MixingPool(ctx context.Context, req *securitypb.QueryMixingPoolRequest) (*securitypb.QueryMixingPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryMixingPoolResponse{}, nil
}

func (qs queryServer) AllMixingPools(ctx context.Context, req *securitypb.QueryAllMixingPoolsRequest) (*securitypb.QueryAllMixingPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(qs.keeper.storeKey)
	poolStore := prefix.NewStore(store, types.MixingPoolKey)

	var pools []securitypb.MixingPool
	pageRes, err := query.Paginate(poolStore, req.Pagination, func(key, value []byte) error {
		var pool securitypb.MixingPool
		if err := qs.keeper.cdc.Unmarshal(value, &pool); err != nil {
			return fmt.Errorf("failed to unmarshal mixing pool: %w", err)
		}
		// Apply status filter if provided
		if req.Status != "" && pool.Status != req.Status {
			return nil
		}
		pools = append(pools, pool)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &securitypb.QueryAllMixingPoolsResponse{
		Pools:      pools,
		Pagination: pageRes,
	}, nil
}

func (qs queryServer) StealthAddress(ctx context.Context, req *securitypb.QueryStealthAddressRequest) (*securitypb.QueryStealthAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryStealthAddressResponse{}, nil
}

func (qs queryServer) VerifyRingSignature(ctx context.Context, req *securitypb.QueryVerifyRingSignatureRequest) (*securitypb.QueryVerifyRingSignatureResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	return &securitypb.QueryVerifyRingSignatureResponse{}, nil
}
