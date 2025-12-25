// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import pb "github.com/aequitas/aura/proto/aura/networksecurity/v1beta1"

// Re-export all proto types
type (
	// Core types
	PeerInfo                  = pb.PeerInfo
	TrustedPeer               = pb.TrustedPeer
	NodeReputation            = pb.NodeReputation
	ForkAlert                 = pb.ForkAlert
	PartitionAlert            = pb.PartitionAlert
	RateLimitEntry            = pb.RateLimitEntry
	MempoolStats              = pb.MempoolStats
	ConnectionConfig          = pb.ConnectionConfig
	GossipConfig              = pb.GossipConfig
	MempoolConfig             = pb.MempoolConfig
	RateLimitConfig           = pb.RateLimitConfig
	ReputationConfig          = pb.ReputationConfig
	ForkDetectionConfig       = pb.ForkDetectionConfig
	PartitionDetectionConfig  = pb.PartitionDetectionConfig
	Params                    = pb.Params

	// Message types
	MsgAddTrustedPeer              = pb.MsgAddTrustedPeer
	MsgAddTrustedPeerResponse      = pb.MsgAddTrustedPeerResponse
	MsgRemoveTrustedPeer           = pb.MsgRemoveTrustedPeer
	MsgRemoveTrustedPeerResponse   = pb.MsgRemoveTrustedPeerResponse
	MsgBanPeer                     = pb.MsgBanPeer
	MsgBanPeerResponse             = pb.MsgBanPeerResponse
	MsgUnbanPeer                   = pb.MsgUnbanPeer
	MsgUnbanPeerResponse           = pb.MsgUnbanPeerResponse
	MsgUpdatePeerReputation        = pb.MsgUpdatePeerReputation
	MsgUpdatePeerReputationResponse = pb.MsgUpdatePeerReputationResponse
	MsgResolveForkAlert            = pb.MsgResolveForkAlert
	MsgResolveForkAlertResponse    = pb.MsgResolveForkAlertResponse
	MsgResolvePartitionAlert       = pb.MsgResolvePartitionAlert
	MsgResolvePartitionAlertResponse = pb.MsgResolvePartitionAlertResponse
	MsgUpdateParams                = pb.MsgUpdateParams
	MsgUpdateParamsResponse        = pb.MsgUpdateParamsResponse

	// Query types
	QueryPeerInfoRequest            = pb.QueryPeerInfoRequest
	QueryPeerInfoResponse           = pb.QueryPeerInfoResponse
	QueryAllPeersRequest            = pb.QueryAllPeersRequest
	QueryAllPeersResponse           = pb.QueryAllPeersResponse
	QueryTrustedPeersRequest        = pb.QueryTrustedPeersRequest
	QueryTrustedPeersResponse       = pb.QueryTrustedPeersResponse
	QueryPeerReputationRequest      = pb.QueryPeerReputationRequest
	QueryPeerReputationResponse     = pb.QueryPeerReputationResponse
	QueryForkAlertsRequest          = pb.QueryForkAlertsRequest
	QueryForkAlertsResponse         = pb.QueryForkAlertsResponse
	QueryPartitionAlertsRequest     = pb.QueryPartitionAlertsRequest
	QueryPartitionAlertsResponse    = pb.QueryPartitionAlertsResponse
	QueryNetworkHealthRequest       = pb.QueryNetworkHealthRequest
	QueryNetworkHealthResponse      = pb.QueryNetworkHealthResponse
	QueryRateLimitStatusRequest     = pb.QueryRateLimitStatusRequest
	QueryRateLimitStatusResponse    = pb.QueryRateLimitStatusResponse
	QueryMempoolStatsRequest        = pb.QueryMempoolStatsRequest
	QueryMempoolStatsResponse       = pb.QueryMempoolStatsResponse
	QueryParamsRequest              = pb.QueryParamsRequest
	QueryParamsResponse             = pb.QueryParamsResponse

	// Genesis types
	GenesisState = pb.GenesisState

	// gRPC Server types
	MsgServer                = pb.MsgServer
	UnimplementedMsgServer   = pb.UnimplementedMsgServer
	QueryServer              = pb.QueryServer
	UnimplementedQueryServer = pb.UnimplementedQueryServer
)

// Re-export gRPC registration functions
var (
	RegisterMsgServer   = pb.RegisterMsgServer
	RegisterQueryServer = pb.RegisterQueryServer
)
