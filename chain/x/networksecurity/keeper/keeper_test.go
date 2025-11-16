package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/networksecurity/keeper"
	"github.com/aequitas/aura/chain/x/networksecurity/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.Codec
}

func (suite *KeeperTestSuite) SetupTest() {
	// Create store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(suite.T(), stateStore.LoadLatestVersion())

	// Create codec
	registry := codectypes.NewInterfaceRegistry()
	suite.cdc = codec.NewProtoCodec(registry)

	// Create context
	suite.ctx = sdk.NewContext(stateStore, cmtproto.Header{Time: time.Now()}, false, log.NewNopLogger())

	// Create keeper
	suite.keeper = keeper.NewKeeper(
		suite.cdc,
		runtime.NewKVStoreService(storeKey),
		"authority",
		log.NewNopLogger(),
	)

	// Initialize with default params
	require.NoError(suite.T(), suite.keeper.SetParams(suite.ctx, types.DefaultParams()))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestParams() {
	params := types.DefaultParams()
	err := suite.keeper.SetParams(suite.ctx, params)
	require.NoError(suite.T(), err)

	retrievedParams, err := suite.keeper.GetParams(suite.ctx)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), params, retrievedParams)
}

func (suite *KeeperTestSuite) TestPeerInfo() {
	peerInfo := types.PeerInfo{
		PeerId:          "peer1",
		IpAddress:       "192.168.1.1",
		ConnectionType:  "inbound",
		ConnectedAt:     suite.ctx.BlockTime(),
		ReputationScore: 100,
		IsTrusted:       false,
		Asn:             12345,
		Region:          "US-East",
	}

	// Set peer info
	err := suite.keeper.SetPeerInfo(suite.ctx, peerInfo)
	require.NoError(suite.T(), err)

	// Get peer info
	retrieved, found := suite.keeper.GetPeerInfo(suite.ctx, "peer1")
	require.True(suite.T(), found)
	require.Equal(suite.T(), peerInfo.PeerId, retrieved.PeerId)
	require.Equal(suite.T(), peerInfo.IpAddress, retrieved.IpAddress)

	// Get all peers
	allPeers := suite.keeper.GetAllPeers(suite.ctx)
	require.Len(suite.T(), allPeers, 1)
}

func (suite *KeeperTestSuite) TestTrustedPeers() {
	trustedPeer := types.TrustedPeer{
		PeerId:      "trusted1",
		Address:     "192.168.1.100",
		PublicKey:   []byte("publickey"),
		Description: "Test trusted peer",
		AddedAt:     suite.ctx.BlockTime(),
	}

	// Add trusted peer
	err := suite.keeper.SetTrustedPeer(suite.ctx, trustedPeer)
	require.NoError(suite.T(), err)

	// Check if trusted
	isTrusted := suite.keeper.IsTrustedPeer(suite.ctx, "trusted1")
	require.True(suite.T(), isTrusted)

	// Get trusted peer
	retrieved, found := suite.keeper.GetTrustedPeer(suite.ctx, "trusted1")
	require.True(suite.T(), found)
	require.Equal(suite.T(), trustedPeer.PeerId, retrieved.PeerId)

	// Remove trusted peer
	err = suite.keeper.RemoveTrustedPeer(suite.ctx, "trusted1")
	require.NoError(suite.T(), err)

	isTrusted = suite.keeper.IsTrustedPeer(suite.ctx, "trusted1")
	require.False(suite.T(), isTrusted)
}

func (suite *KeeperTestSuite) TestBanPeer() {
	peerID := "badpeer"

	// Ban peer
	err := suite.keeper.BanPeer(suite.ctx, peerID, 3600, "misbehavior")
	require.NoError(suite.T(), err)

	// Check if banned
	isBanned := suite.keeper.IsBanned(suite.ctx, peerID)
	require.True(suite.T(), isBanned)

	// Unban peer
	err = suite.keeper.UnbanPeer(suite.ctx, peerID)
	require.NoError(suite.T(), err)

	isBanned = suite.keeper.IsBanned(suite.ctx, peerID)
	require.False(suite.T(), isBanned)
}

func (suite *KeeperTestSuite) TestReputation() {
	peerID := "peer1"

	// Set reputation
	reputation := types.NodeReputation{
		PeerId:            peerID,
		Score:             100,
		LastUpdatedHeight: suite.ctx.BlockHeight(),
		MessagesReceived:  50,
		ValidMessages:     45,
		InvalidMessages:   5,
		Uptime:            1000,
		MisbehaviorCount:  1,
	}

	err := suite.keeper.SetReputation(suite.ctx, reputation)
	require.NoError(suite.T(), err)

	// Get reputation
	retrieved, found := suite.keeper.GetReputation(suite.ctx, peerID)
	require.True(suite.T(), found)
	require.Equal(suite.T(), reputation.Score, retrieved.Score)

	// Penalize reputation
	suite.keeper.PenalizeReputation(suite.ctx, peerID, 20)
	retrieved, _ = suite.keeper.GetReputation(suite.ctx, peerID)
	require.Equal(suite.T(), int64(80), retrieved.Score)

	// Reward reputation
	suite.keeper.RewardReputation(suite.ctx, peerID, 10)
	retrieved, _ = suite.keeper.GetReputation(suite.ctx, peerID)
	require.Equal(suite.T(), int64(90), retrieved.Score)
}

func (suite *KeeperTestSuite) TestRateLimiting() {
	peerID := "peer1"

	// Check rate limit (should pass initially)
	err := suite.keeper.CheckRateLimit(suite.ctx, peerID)
	require.NoError(suite.T(), err)

	// Get rate limiter
	limiter := suite.keeper.GetRateLimiter(suite.ctx, peerID)
	require.NotNil(suite.T(), limiter)

	// Exhaust rate limit
	for i := 0; i < 200; i++ {
		limiter.Allow()
	}

	// Should fail now
	err = suite.keeper.CheckRateLimit(suite.ctx, peerID)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), types.ErrRateLimitExceeded, err)
}

func (suite *KeeperTestSuite) TestConnectionManagement() {
	peerInfo := types.PeerInfo{
		PeerId:          "peer1",
		IpAddress:       "192.168.1.1",
		ConnectionType:  "inbound",
		ConnectedAt:     suite.ctx.BlockTime(),
		ReputationScore: 100,
		Asn:             12345,
	}

	// Accept connection
	err := suite.keeper.AcceptConnection(suite.ctx, peerInfo)
	require.NoError(suite.T(), err)

	// Check connection count
	count := suite.keeper.GetConnectionCount(suite.ctx, "192.168.1.1")
	require.Equal(suite.T(), uint32(1), count)

	// Disconnect peer
	err = suite.keeper.DisconnectPeer(suite.ctx, "peer1")
	require.NoError(suite.T(), err)

	// Connection count should be decremented
	count = suite.keeper.GetConnectionCount(suite.ctx, "192.168.1.1")
	require.Equal(suite.T(), uint32(0), count)
}

func (suite *KeeperTestSuite) TestForkDetection() {
	blockHash1 := []byte("hash1")
	blockHash2 := []byte("hash2")

	// First block at height 100
	err := suite.keeper.DetectFork(suite.ctx, 100, blockHash1)
	require.NoError(suite.T(), err)

	// Different hash at same height - should detect fork
	err = suite.keeper.DetectFork(suite.ctx, 100, blockHash2)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), types.ErrForkDetected, err)

	// Should have created a fork alert
	alerts := suite.keeper.GetAllForkAlerts(suite.ctx, false)
	require.Len(suite.T(), alerts, 1)
}

func (suite *KeeperTestSuite) TestPartitionDetection() {
	// Set expected peer count
	suite.keeper.UpdateExpectedPeerCount(suite.ctx, 20)

	// Add only a few peers
	for i := 0; i < 5; i++ {
		peerInfo := types.PeerInfo{
			PeerId:    fmt.Sprintf("peer%d", i),
			IpAddress: fmt.Sprintf("192.168.1.%d", i),
		}
		suite.keeper.SetPeerInfo(suite.ctx, peerInfo)
	}

	// Should detect partition
	err := suite.keeper.DetectPartition(suite.ctx)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), types.ErrPartitionDetected, err)

	// Should have created a partition alert
	alerts := suite.keeper.GetAllPartitionAlerts(suite.ctx, false)
	require.Len(suite.T(), alerts, 1)
}

func (suite *KeeperTestSuite) TestSybilDetection() {
	// Create peers all from same subnet
	for i := 0; i < 10; i++ {
		peerInfo := types.PeerInfo{
			PeerId:    fmt.Sprintf("peer%d", i),
			IpAddress: fmt.Sprintf("192.168.1.%d", i), // All in same /24
			Asn:       12345,                          // Same ASN
		}
		suite.keeper.SetPeerInfo(suite.ctx, peerInfo)
	}

	// Should detect Sybil attack
	err := suite.keeper.CheckSybilResistance(suite.ctx)
	require.Error(suite.T(), err)
}

func (suite *KeeperTestSuite) TestMempoolValidation() {
	// Create a mock transaction
	// In actual tests, you would use a real transaction
	txBytes := []byte("mock_tx")
	sender := "sender1"

	// Should pass validation initially
	err := suite.keeper.AntiSpamCheck(suite.ctx, nil, sender)
	require.NoError(suite.T(), err)

	// Add many transactions from same sender
	for i := 0; i < 150; i++ {
		suite.keeper.SetAccountMempoolTxCount(suite.ctx, sender, uint32(i+1))
	}

	// Should fail anti-spam check
	err = suite.keeper.AntiSpamCheck(suite.ctx, nil, sender)
	require.Error(suite.T(), err)
}

func (suite *KeeperTestSuite) TestGossipValidation() {
	msg := &keeper.GossipMessage{
		MessageID:   "msg1",
		Content:     []byte("test message"),
		SenderID:    "peer1",
		Timestamp:   time.Now(),
		TTL:         time.Minute * 5,
		MessageType: "test",
	}

	// Should validate successfully
	err := suite.keeper.ValidateGossipMessage(suite.ctx, msg)
	require.NoError(suite.T(), err)

	// Should detect duplicate
	err = suite.keeper.ValidateGossipMessage(suite.ctx, msg)
	require.Error(suite.T(), err)
	require.Equal(suite.T(), types.ErrDuplicateMessage, err)
}
