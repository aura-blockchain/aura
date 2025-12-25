// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

type QueryServerTestSuite struct {
	KeeperTestSuite
	queryServer types.QueryServer
}

func TestQueryServerTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

func (suite *QueryServerTestSuite) SetupTest() {
	suite.KeeperTestSuite.SetupTest()
	suite.queryServer = NewQueryServerImpl(suite.Keeper)
}

func (suite *QueryServerTestSuite) TestQueryServerImplementation() {
	suite.NotNil(suite.queryServer, "query server should be created")
}

func (suite *QueryServerTestSuite) TestAllPeersPagination() {
	// Add test peers
	for i := 0; i < 15; i++ {
		peerID := "peer" + string(rune('A'+i))
		peer := types.PeerInfo{
			PeerId:          peerID,
			IpAddress:       "192.168.1." + string(rune('1'+i)),
			ConnectionType:  "inbound",
			ConnectedAt:     time.Now(),
			ReputationScore: 100,
			IsTrusted:       false,
		}
		suite.Keeper.SetPeerInfo(suite.SdkCtx, peer)
	}

	// Test first page
	resp, err := suite.queryServer.AllPeers(suite.SdkCtx, &types.QueryAllPeersRequest{
		Pagination: &query.PageRequest{
			Limit: 5,
		},
	})
	suite.NoError(err)
	suite.Len(resp.Peers, 5)
	suite.NotNil(resp.Pagination)
	suite.NotNil(resp.Pagination.NextKey)

	// Test second page
	resp2, err := suite.queryServer.AllPeers(suite.SdkCtx, &types.QueryAllPeersRequest{
		Pagination: &query.PageRequest{
			Key:   resp.Pagination.NextKey,
			Limit: 5,
		},
	})
	suite.NoError(err)
	suite.Len(resp2.Peers, 5)

	// Test all results without pagination
	respAll, err := suite.queryServer.AllPeers(suite.SdkCtx, &types.QueryAllPeersRequest{})
	suite.NoError(err)
	suite.Len(respAll.Peers, 15)
}

func (suite *QueryServerTestSuite) TestTrustedPeersPagination() {
	// Add test trusted peers
	for i := 0; i < 12; i++ {
		peerID := "trusted" + string(rune('A'+i))
		peer := types.TrustedPeer{
			PeerId:      peerID,
			Address:     "192.168.1." + string(rune('1'+i)),
			PublicKey:   []byte("pubkey" + string(rune('A'+i))),
			Description: "test peer",
			AddedAt:     time.Now(),
		}
		suite.Keeper.SetTrustedPeer(suite.SdkCtx, peer)
	}

	// Test first page
	resp, err := suite.queryServer.TrustedPeers(suite.SdkCtx, &types.QueryTrustedPeersRequest{
		Pagination: &query.PageRequest{
			Limit: 5,
		},
	})
	suite.NoError(err)
	suite.Len(resp.Peers, 5)
	suite.NotNil(resp.Pagination)

	// Test all without pagination
	respAll, err := suite.queryServer.TrustedPeers(suite.SdkCtx, &types.QueryTrustedPeersRequest{})
	suite.NoError(err)
	suite.Len(respAll.Peers, 12)
}

func (suite *QueryServerTestSuite) TestForkAlertsPagination() {
	// Add test fork alerts
	for i := 0; i < 10; i++ {
		alertID := "fork" + string(rune('A'+i))
		alert := types.ForkAlert{
			AlertId:     alertID,
			BlockHeight: int64(100 + i),
			ChainAHash:  []byte("hash1"),
			ChainBHash:  []byte("hash2"),
			DetectedAt:  time.Now(),
			Resolved:    i%2 == 0, // Alternate resolved/unresolved
		}
		suite.Keeper.SetForkAlert(suite.SdkCtx, alert)
	}

	// Test pagination with filter
	resp, err := suite.queryServer.ForkAlerts(suite.SdkCtx, &types.QueryForkAlertsRequest{
		IncludeResolved: false,
		Pagination: &query.PageRequest{
			Limit: 3,
		},
	})
	suite.NoError(err)
	suite.LessOrEqual(len(resp.Alerts), 5) // Only unresolved ones

	// Test all including resolved
	respAll, err := suite.queryServer.ForkAlerts(suite.SdkCtx, &types.QueryForkAlertsRequest{
		IncludeResolved: true,
	})
	suite.NoError(err)
	suite.Len(respAll.Alerts, 10)
}

func (suite *QueryServerTestSuite) TestPartitionAlertsPagination() {
	// Add test partition alerts
	for i := 0; i < 8; i++ {
		alertID := "partition" + string(rune('A'+i))
		alert := types.PartitionAlert{
			AlertId:        alertID,
			ConnectedPeers: uint32(5),
			ExpectedPeers:  uint32(10),
			MissingPeerIds: []string{"peer1", "peer2"},
			DetectedAt:     time.Now(),
			Resolved:       i >= 5, // Last 3 resolved
		}
		suite.Keeper.SetPartitionAlert(suite.SdkCtx, alert)
	}

	// Test pagination without resolved
	resp, err := suite.queryServer.PartitionAlerts(suite.SdkCtx, &types.QueryPartitionAlertsRequest{
		IncludeResolved: false,
		Pagination: &query.PageRequest{
			Limit: 3,
		},
	})
	suite.NoError(err)
	suite.LessOrEqual(len(resp.Alerts), 5) // Only unresolved

	// Test all including resolved
	respAll, err := suite.queryServer.PartitionAlerts(suite.SdkCtx, &types.QueryPartitionAlertsRequest{
		IncludeResolved: true,
	})
	suite.NoError(err)
	suite.Len(respAll.Alerts, 8)
}
