package keeper_test

import (
	"time"

	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

func (suite *KeeperTestSuite) TestGetAllPeers() {
	assertions := suite.Require()
	// Add multiple peers
	for i := 0; i < 5; i++ {
		peerInfo := types.PeerInfo{
			PeerId:          string(rune('a' + i)),
			IpAddress:       "192.168.1.1",
			ReputationScore: 100,
		}
		assertions.NoError(suite.keeper.SetPeerInfo(suite.ctx, peerInfo))
	}

	peers := suite.keeper.GetAllPeers(suite.ctx)
	suite.Require().GreaterOrEqual(len(peers), 5)
}

func (suite *KeeperTestSuite) TestTrustedPeersExtended() {
	assertions := suite.Require()
	trusted := types.TrustedPeer{
		PeerId:  "trusted1",
		Address: "192.168.1.100",
	}

	// Set trusted peer
	assertions.NoError(suite.keeper.SetTrustedPeer(suite.ctx, trusted))

	// Check if trusted
	isTrusted := suite.keeper.IsTrustedPeer(suite.ctx, "trusted1")
	suite.Require().True(isTrusted)

	// Get trusted peer
	retrieved, found := suite.keeper.GetTrustedPeer(suite.ctx, "trusted1")
	suite.Require().True(found)
	suite.Require().Equal("trusted1", retrieved.PeerId)

	// Get all trusted peers
	allTrusted := suite.keeper.GetAllTrustedPeers(suite.ctx)
	suite.Require().GreaterOrEqual(len(allTrusted), 1)

	// Remove trusted peer
	assertions.NoError(suite.keeper.RemoveTrustedPeer(suite.ctx, "trusted1"))

	// Verify removed
	isTrusted = suite.keeper.IsTrustedPeer(suite.ctx, "trusted1")
	suite.Require().False(isTrusted)
}

func (suite *KeeperTestSuite) TestBanUnbanPeer() {
	peerId := "peer_to_ban"

	// Ban peer (ensure duration > 0 so ban is active)
	suite.Require().NoError(suite.keeper.BanPeer(suite.ctx, peerId, 3600, "misbehavior"))

	// Check if banned
	isBanned := suite.keeper.IsBanned(suite.ctx, peerId)
	suite.Require().True(isBanned)

	// Ensure banned peer shows up in exported list
	bannedPeers := suite.keeper.GetBannedPeers(suite.ctx)
	suite.Contains(bannedPeers, peerId)

	// When ban is expired, IsBanned should return false and entry should be cleaned on check.
	expired := suite.ctx.BlockTime().Add(-1 * time.Hour)
	entry, found := suite.keeper.GetRateLimitEntry(suite.ctx, peerId)
	suite.Require().True(found)
	entry.BanExpiresAt = &expired
	entry.IsBanned = true
	suite.Require().NoError(suite.keeper.SetRateLimitEntry(suite.ctx, entry))
	suite.False(suite.keeper.IsBanned(suite.ctx, peerId))
	entryAfter, _ := suite.keeper.GetRateLimitEntry(suite.ctx, peerId)
	suite.False(entryAfter.IsBanned)

	// Unban peer
	suite.Require().NoError(suite.keeper.UnbanPeer(suite.ctx, peerId))

	// Verify unbanned
	isBanned = suite.keeper.IsBanned(suite.ctx, peerId)
	suite.Require().False(isBanned)

	// Banned list should no longer include the peer
	bannedPeers = suite.keeper.GetBannedPeers(suite.ctx)
	suite.NotContains(bannedPeers, peerId)
}

func (suite *KeeperTestSuite) TestBanPersistsAcrossBlocksUntilExpiry() {
	peerID := "peer_persist"
	// duration is in nanoseconds inside BanPeer, so provide explicit nanoseconds to avoid premature expiry
	duration := int64((24 * time.Hour).Nanoseconds())
	suite.Require().NoError(suite.keeper.BanPeer(suite.ctx, peerID, duration, "test"))
	suite.True(suite.keeper.IsBanned(suite.ctx, peerID))

	// Advance block time but still within ban window
	nextCtx := suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(30 * time.Minute))
	suite.True(suite.keeper.IsBanned(nextCtx, peerID), "ban should persist across blocks until expiry")
}

func (suite *KeeperTestSuite) TestReputationExtended() {
	peerId := "peer_rep"

	// Set reputation
	rep := types.NodeReputation{
		PeerId:            peerId,
		Score:             50,
		LastUpdatedHeight: suite.ctx.BlockHeight(),
	}
	suite.Require().NoError(suite.keeper.SetReputation(suite.ctx, rep))

	// Get reputation
	retrieved, found := suite.keeper.GetReputation(suite.ctx, peerId)
	suite.Require().True(found)
	suite.Require().Equal(int64(50), retrieved.Score)

	// Get all reputations
	allReps := suite.keeper.GetAllReputations(suite.ctx)
	suite.Require().GreaterOrEqual(len(allReps), 1)
}

func (suite *KeeperTestSuite) TestRateLimit() {
	peerId := "peer_rate"

	// Set rate limit entry
	entry := types.RateLimitEntry{
		PeerId:       peerId,
		RequestCount: 10,
		WindowStart:  suite.ctx.BlockTime(),
	}
	suite.Require().NoError(suite.keeper.SetRateLimitEntry(suite.ctx, entry))

	// Get rate limit entry
	retrieved, found := suite.keeper.GetRateLimitEntry(suite.ctx, peerId)
	suite.Require().True(found)
	suite.Require().Equal(peerId, retrieved.PeerId)
	suite.Require().Equal(uint64(10), retrieved.RequestCount)
}

func (suite *KeeperTestSuite) TestMempoolStats() {
	stats := types.MempoolStats{
		TxCount:   100,
		SizeBytes: 1000,
	}

	// Set mempool stats
	suite.Require().NoError(suite.keeper.SetMempoolStats(suite.ctx, stats))

	// Get mempool stats
	retrieved := suite.keeper.GetMempoolStats(suite.ctx)
	suite.Require().Equal(uint64(100), retrieved.TxCount)
	suite.Require().Equal(uint64(1000), retrieved.SizeBytes)
}

func (suite *KeeperTestSuite) TestForkAlert() {
	alert := types.ForkAlert{
		AlertId:     "fork1",
		BlockHeight: suite.ctx.BlockHeight(),
		DetectedAt:  suite.ctx.BlockTime(),
	}

	// Set fork alert
	suite.Require().NoError(suite.keeper.SetForkAlert(suite.ctx, alert))

	// Get fork alert
	retrieved, found := suite.keeper.GetForkAlert(suite.ctx, "fork1")
	suite.Require().True(found)
	suite.Require().Equal("fork1", retrieved.AlertId)

	// Get all fork alerts
	allAlerts := suite.keeper.GetAllForkAlerts(suite.ctx, false)
	suite.Require().GreaterOrEqual(len(allAlerts), 1)
}

func (suite *KeeperTestSuite) TestPartitionAlert() {
	alert := types.PartitionAlert{
		AlertId:        "partition1",
		ConnectedPeers: 5,
		DetectedAt:     suite.ctx.BlockTime(),
	}

	// Set partition alert
	suite.Require().NoError(suite.keeper.SetPartitionAlert(suite.ctx, alert))

	// Get partition alert
	retrieved, found := suite.keeper.GetPartitionAlert(suite.ctx, "partition1")
	suite.Require().True(found)
	suite.Require().Equal("partition1", retrieved.AlertId)

	// Get all partition alerts
	allAlerts := suite.keeper.GetAllPartitionAlerts(suite.ctx, false)
	suite.Require().GreaterOrEqual(len(allAlerts), 1)
}

func (suite *KeeperTestSuite) TestConnectionCount() {
	peerId := "peer_conn"

	// Set connection count
	suite.Require().NoError(suite.keeper.SetConnectionCount(suite.ctx, peerId, 5))

	// Get connection count
	count := suite.keeper.GetConnectionCount(suite.ctx, peerId)
	suite.Require().Equal(uint32(5), count)

	// Increment connection count
	suite.Require().NoError(suite.keeper.IncrementConnectionCount(suite.ctx, peerId))
	count = suite.keeper.GetConnectionCount(suite.ctx, peerId)
	suite.Require().Equal(uint32(6), count)

	// Decrement connection count
	suite.Require().NoError(suite.keeper.DecrementConnectionCount(suite.ctx, peerId))
	count = suite.keeper.GetConnectionCount(suite.ctx, peerId)
	suite.Require().Equal(uint32(5), count)

	// Stress increments should not overflow and stay monotonic
	for i := 0; i < 1000; i++ {
		suite.Require().NoError(suite.keeper.IncrementConnectionCount(suite.ctx, peerId))
	}
	count = suite.keeper.GetConnectionCount(suite.ctx, peerId)
	suite.Require().GreaterOrEqual(count, uint32(1005))
}

func (suite *KeeperTestSuite) TestGetAuthority() {
	authority := suite.keeper.GetAuthority()
	suite.Require().Equal("authority", authority)
}

func (suite *KeeperTestSuite) TestLogger() {
	logger := suite.keeper.Logger(suite.ctx)
	suite.Require().NotNil(logger)
}

func (suite *KeeperTestSuite) TestGetPeerInfoNotFound() {
	_, found := suite.keeper.GetPeerInfo(suite.ctx, "nonexistent")
	suite.Require().False(found)
}

func (suite *KeeperTestSuite) TestGetTrustedPeerNotFound() {
	_, found := suite.keeper.GetTrustedPeer(suite.ctx, "nonexistent")
	suite.Require().False(found)
}

func (suite *KeeperTestSuite) TestGetForkAlertNotFound() {
	_, found := suite.keeper.GetForkAlert(suite.ctx, "nonexistent")
	suite.Require().False(found)
}

func (suite *KeeperTestSuite) TestGetPartitionAlertNotFound() {
	_, found := suite.keeper.GetPartitionAlert(suite.ctx, "nonexistent")
	suite.Require().False(found)
}

func (suite *KeeperTestSuite) TestGetRateLimitEntryNotFound() {
	_, found := suite.keeper.GetRateLimitEntry(suite.ctx, "nonexistent")
	suite.Require().False(found)
}

func (suite *KeeperTestSuite) TestIsBannedNotFound() {
	isBanned := suite.keeper.IsBanned(suite.ctx, "nonexistent")
	suite.Require().False(isBanned)
}

func (suite *KeeperTestSuite) TestIsTrustedPeerNotFound() {
	isTrusted := suite.keeper.IsTrustedPeer(suite.ctx, "nonexistent")
	suite.Require().False(isTrusted)
}

func (suite *KeeperTestSuite) TestGetReputationNotFound() {
	_, found := suite.keeper.GetReputation(suite.ctx, "nonexistent")
	suite.Require().False(found)
}

func (suite *KeeperTestSuite) TestGetConnectionCountNotFound() {
	count := suite.keeper.GetConnectionCount(suite.ctx, "nonexistent")
	suite.Require().Equal(uint32(0), count)
}

func (suite *KeeperTestSuite) TestDecrementConnectionCountZero() {
	peerId := "peer_zero"

	// Decrement when count is already 0
	suite.Require().NoError(suite.keeper.DecrementConnectionCount(suite.ctx, peerId))

	// Should stay at 0
	count := suite.keeper.GetConnectionCount(suite.ctx, peerId)
	suite.Require().Equal(uint32(0), count)
}

func (suite *KeeperTestSuite) TestCheckGossipMessage() {
	msgID := []byte("msg123")

	// First check - message is new
	isNew, hash, err := suite.keeper.CheckGossipMessage(suite.ctx, msgID)
	suite.Require().NoError(err)
	suite.Require().True(isNew)
	suite.Require().NotEmpty(hash)

	// Second check - duplicate detected with explicit error
	isNew, _, err = suite.keeper.CheckGossipMessage(suite.ctx, msgID)
	suite.Require().False(isNew)
	suite.Require().ErrorIs(err, types.ErrDuplicateMessage)

	stats := suite.keeper.GetMessageCacheStats()
	suite.Require().Equal(uint64(1), stats.Hits)
	suite.Require().Equal(uint64(1), stats.Misses)
}

func (suite *KeeperTestSuite) TestGetMessageCacheStats() {
	// Record some messages including a duplicate to exercise counters.
	_, _, err := suite.keeper.CheckGossipMessage(suite.ctx, []byte("msg1"))
	suite.Require().NoError(err)
	isNew, _, err := suite.keeper.CheckGossipMessage(suite.ctx, []byte("msg2"))
	suite.Require().NoError(err)
	suite.Require().True(isNew)
	_, _, err = suite.keeper.CheckGossipMessage(suite.ctx, []byte("msg3"))
	suite.Require().NoError(err)
	// msg2 again - should be duplicate
	isNew2, _, err := suite.keeper.CheckGossipMessage(suite.ctx, []byte("msg2"))
	suite.Require().ErrorIs(err, types.ErrDuplicateMessage)
	suite.Require().False(isNew2)

	stats := suite.keeper.GetMessageCacheStats()
	suite.Require().Equal(3, stats.Size)
	suite.Require().Equal(uint64(1), stats.Hits)
	suite.Require().Equal(uint64(3), stats.Misses)
	suite.Require().Positive(stats.MaxSize)
}
