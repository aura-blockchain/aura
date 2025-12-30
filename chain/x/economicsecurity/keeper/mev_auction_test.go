// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// =============================================================================
// MEV Auction Tests
//
// Note: The MEV auction implementation uses in-memory state that resets each call
// to getMEVAuctionState. These tests focus on individual function behavior and
// helper functions that don't require state persistence across calls.
// =============================================================================

func TestCreateMEVAuction_Success(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: true,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)
	_ = k.SetCurrentTime(ctx, 1000)

	auctionID, err := k.CreateMEVAuction(ctx, 100)
	require.NoError(t, err)
	require.NotEmpty(t, auctionID)
	require.Contains(t, auctionID, "mev-")
}

func TestCreateMEVAuction_Disabled(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: false,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	_, err := k.CreateMEVAuction(ctx, 100)
	require.ErrorIs(t, err, types.ErrMEVAuctionDisabled)
}

func TestCreateMEVAuction_NilMEVParams(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = nil

	k, ctx := setupKeeperWithCustomParams(t, params)

	_, err := k.CreateMEVAuction(ctx, 100)
	require.ErrorIs(t, err, types.ErrMEVAuctionDisabled)
}

func TestPlaceMEVBid_Disabled(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: false,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	_, err := k.PlaceMEVBid(ctx, "auction-123", "aura1bidder1", "5000000", 10)
	require.ErrorIs(t, err, types.ErrMEVAuctionDisabled)
}

func TestPlaceMEVBid_AuctionNotFound(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: true,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	_, err := k.PlaceMEVBid(ctx, "nonexistent-auction", "aura1bidder1", "5000000", 10)
	require.ErrorIs(t, err, types.ErrAuctionNotFound)
}

func TestCloseMEVAuction_Disabled(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: false,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	_, _, err := k.CloseMEVAuction(ctx, "auction-123")
	require.ErrorIs(t, err, types.ErrMEVAuctionDisabled)
}

func TestCloseMEVAuction_NotFound(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: true,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	_, _, err := k.CloseMEVAuction(ctx, "nonexistent-auction")
	require.ErrorIs(t, err, types.ErrAuctionNotFound)
}

func TestSelectFirstPriceWinner_Empty(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	winner := k.selectFirstPriceWinner([]*MEVAuctionBid{})
	require.Nil(t, winner)
}

func TestSelectFirstPriceWinner_SingleBid(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	bids := []*MEVAuctionBid{
		{BidID: "bid-1", Bidder: "aura1bidder1", Amount: "5000000", Priority: 10},
	}

	winner := k.selectFirstPriceWinner(bids)
	require.NotNil(t, winner)
	require.Equal(t, "aura1bidder1", winner.Bidder)
}

func TestSelectFirstPriceWinner_MultipleBids(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	bids := []*MEVAuctionBid{
		{BidID: "bid-1", Bidder: "aura1bidder1", Amount: "5000000", Priority: 10},
		{BidID: "bid-2", Bidder: "aura1bidder2", Amount: "7000000", Priority: 20},
		{BidID: "bid-3", Bidder: "aura1bidder3", Amount: "3000000", Priority: 5},
	}

	winner := k.selectFirstPriceWinner(bids)
	require.NotNil(t, winner)
	require.Equal(t, "aura1bidder2", winner.Bidder)
	require.Equal(t, "7000000", winner.Amount)
}

func TestSelectFirstPriceWinner_TieBreakByPriority(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	bids := []*MEVAuctionBid{
		{BidID: "bid-1", Bidder: "aura1bidder1", Amount: "5000000", Priority: 10},
		{BidID: "bid-2", Bidder: "aura1bidder2", Amount: "5000000", Priority: 50}, // Higher priority
	}

	winner := k.selectFirstPriceWinner(bids)
	require.NotNil(t, winner)
	require.Equal(t, "aura1bidder2", winner.Bidder) // Higher priority wins
}

func TestSelectSecondPriceWinner_TwoBids(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	bids := []*MEVAuctionBid{
		{BidID: "bid-1", Bidder: "aura1bidder1", Amount: "5000000", Priority: 10},
		{BidID: "bid-2", Bidder: "aura1bidder2", Amount: "7000000", Priority: 20},
	}

	winner := k.selectSecondPriceWinner(bids)
	require.NotNil(t, winner)
	require.Equal(t, "aura1bidder2", winner.Bidder)
	require.Equal(t, "5000000", winner.Amount) // Pays second price
}

func TestSelectSecondPriceWinner_SingleBid(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	bids := []*MEVAuctionBid{
		{BidID: "bid-1", Bidder: "aura1bidder1", Amount: "5000000", Priority: 10},
	}

	winner := k.selectSecondPriceWinner(bids)
	require.NotNil(t, winner)
	require.Equal(t, "5000000", winner.Amount) // Pays own price when alone
}

func TestSelectSecondPriceWinner_ThreeBids(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	bids := []*MEVAuctionBid{
		{BidID: "bid-1", Bidder: "aura1bidder1", Amount: "5000000", Priority: 10},
		{BidID: "bid-2", Bidder: "aura1bidder2", Amount: "9000000", Priority: 20},
		{BidID: "bid-3", Bidder: "aura1bidder3", Amount: "7000000", Priority: 15},
	}

	winner := k.selectSecondPriceWinner(bids)
	require.NotNil(t, winner)
	require.Equal(t, "aura1bidder2", winner.Bidder)
	require.Equal(t, "7000000", winner.Amount) // Pays second highest
}

func TestSelectVickreyWinner(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	bids := []*MEVAuctionBid{
		{BidID: "bid-1", Bidder: "aura1bidder1", Amount: "5000000", Priority: 10},
		{BidID: "bid-2", Bidder: "aura1bidder2", Amount: "8000000", Priority: 20},
		{BidID: "bid-3", Bidder: "aura1bidder3", Amount: "6000000", Priority: 15},
	}

	winner := k.selectVickreyWinner(bids)
	require.NotNil(t, winner)
	require.Equal(t, "aura1bidder2", winner.Bidder)
	require.Equal(t, "6000000", winner.Amount) // Pays second highest
}

func TestSelectVickreyWinner_Empty(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	winner := k.selectVickreyWinner([]*MEVAuctionBid{})
	require.Nil(t, winner)
}

func TestSelectDutchWinner_MeetsReserve(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	bids := []*MEVAuctionBid{
		{BidID: "bid-1", Bidder: "aura1bidder1", Amount: "3000000", Priority: 10},
		{BidID: "bid-2", Bidder: "aura1bidder2", Amount: "6000000", Priority: 20},
	}

	winner := k.selectDutchWinner(bids, "5000000")
	require.NotNil(t, winner)
	require.Equal(t, "aura1bidder2", winner.Bidder)
}

func TestSelectDutchWinner_FirstMeetsReserve(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	bids := []*MEVAuctionBid{
		{BidID: "bid-1", Bidder: "aura1bidder1", Amount: "6000000", Priority: 10},
		{BidID: "bid-2", Bidder: "aura1bidder2", Amount: "7000000", Priority: 20},
	}

	// First bid that meets reserve wins
	winner := k.selectDutchWinner(bids, "5000000")
	require.NotNil(t, winner)
	require.Equal(t, "aura1bidder1", winner.Bidder)
}

func TestSelectDutchWinner_NoOneMeetsReserve(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	bids := []*MEVAuctionBid{
		{BidID: "bid-1", Bidder: "aura1bidder1", Amount: "3000000", Priority: 10},
		{BidID: "bid-2", Bidder: "aura1bidder2", Amount: "4000000", Priority: 20},
	}

	winner := k.selectDutchWinner(bids, "5000000")
	require.Nil(t, winner)
}

func TestSelectDutchWinner_ExactReserve(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	bids := []*MEVAuctionBid{
		{BidID: "bid-1", Bidder: "aura1bidder1", Amount: "5000000", Priority: 10},
	}

	winner := k.selectDutchWinner(bids, "5000000")
	require.NotNil(t, winner)
	require.Equal(t, "aura1bidder1", winner.Bidder)
}

func TestGetActiveAuctions_Empty(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: true,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	auctions := k.GetActiveAuctions(ctx)
	require.Empty(t, auctions)
}

func TestGetAuction_NotFound(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: true,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	_, err := k.GetAuction(ctx, "nonexistent")
	require.ErrorIs(t, err, types.ErrAuctionNotFound)
}

func TestGetAuctionStatistics(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: true,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	completed, revenue, active, avgBid := k.GetAuctionStatistics(ctx)
	require.Equal(t, uint64(0), completed)
	require.Equal(t, "0", revenue)
	require.Equal(t, uint64(0), active)
	require.Equal(t, "0", avgBid)
}

func TestCancelAuction_NotFound(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: true,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	err := k.CancelAuction(ctx, "nonexistent", "testing")
	require.ErrorIs(t, err, types.ErrAuctionNotFound)
}

func TestGetAuctionHistory_Empty(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: true,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	history := k.GetAuctionHistory(ctx, 10)
	require.Empty(t, history)
}

func TestGetAuctionHistory_ZeroLimit(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: true,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	// Zero limit should return all
	history := k.GetAuctionHistory(ctx, 0)
	require.NotNil(t, history)
}

func TestConvertAuctionToProto(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	auction := &MEVAuction{
		AuctionID:    "mev-test123",
		BlockSlot:    100,
		Bids:         []*MEVAuctionBid{},
		WinningBidID: "",
		TotalValue:   "0",
		Status:       "open",
		CreatedAt:    1000,
		ClosedAt:     0,
		MinimumBid:   "1000000",
		ReservePrice: "5000000",
		AuctionType:  MEVAuctionTypeFirstPrice,
	}

	proto := k.ConvertAuctionToProto(auction)
	require.NotNil(t, proto)
	require.Equal(t, "mev-test123", proto["auction_id"])
	require.Equal(t, uint64(100), proto["block_slot"])
	require.Equal(t, "open", proto["status"])
	require.Equal(t, "1000000", proto["minimum_bid"])
	require.Equal(t, "5000000", proto["reserve_price"])
	require.Equal(t, "first_price", proto["auction_type"])
}

func TestConvertAuctionToProto_WithBids(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	auction := &MEVAuction{
		AuctionID: "mev-test123",
		BlockSlot: 100,
		Bids: []*MEVAuctionBid{
			{BidID: "bid-1", Bidder: "aura1bidder1", Amount: "5000000", Priority: 10, Timestamp: 1000, Status: "active"},
			{BidID: "bid-2", Bidder: "aura1bidder2", Amount: "6000000", Priority: 20, Timestamp: 1001, Status: "active"},
		},
		WinningBidID: "",
		TotalValue:   "11000000",
		Status:       "open",
		CreatedAt:    1000,
		ClosedAt:     0,
		MinimumBid:   "1000000",
		ReservePrice: "5000000",
		AuctionType:  MEVAuctionTypeFirstPrice,
	}

	proto := k.ConvertAuctionToProto(auction)
	require.NotNil(t, proto)

	bids, ok := proto["bids"].([]map[string]interface{})
	require.True(t, ok)
	require.Len(t, bids, 2)
	require.Equal(t, "aura1bidder1", bids[0]["bidder"])
	require.Equal(t, "aura1bidder2", bids[1]["bidder"])
}

func TestConvertAuctionToProto_ClosedAuction(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	auction := &MEVAuction{
		AuctionID:    "mev-test123",
		BlockSlot:    100,
		Bids:         []*MEVAuctionBid{},
		WinningBidID: "bid-1",
		TotalValue:   "5000000",
		Status:       "closed",
		CreatedAt:    1000,
		ClosedAt:     2000,
		MinimumBid:   "1000000",
		ReservePrice: "5000000",
		AuctionType:  MEVAuctionTypeSecondPrice,
	}

	proto := k.ConvertAuctionToProto(auction)
	require.NotNil(t, proto)
	require.Equal(t, "closed", proto["status"])
	require.Equal(t, "bid-1", proto["winning_bid_id"])
	require.Equal(t, "second_price", proto["auction_type"])
}

func TestGenerateAuctionID_Deterministic(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	id1 := k.generateAuctionID(100, 1000)
	id2 := k.generateAuctionID(100, 1000)
	id3 := k.generateAuctionID(101, 1000)

	require.Equal(t, id1, id2)
	require.NotEqual(t, id1, id3)
	require.Contains(t, id1, "mev-")
}

func TestGenerateAuctionID_DifferentTimestamps(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	id1 := k.generateAuctionID(100, 1000)
	id2 := k.generateAuctionID(100, 1001)

	require.NotEqual(t, id1, id2)
}

func TestGenerateBidID_Deterministic(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	id1 := k.generateBidID("auction1", "bidder1", "1000", 1000)
	id2 := k.generateBidID("auction1", "bidder1", "1000", 1000)
	id3 := k.generateBidID("auction1", "bidder2", "1000", 1000)

	require.Equal(t, id1, id2)
	require.NotEqual(t, id1, id3)
	require.Contains(t, id1, "bid-")
}

func TestGenerateBidID_DifferentAmounts(t *testing.T) {
	k, _ := setupKeeperForTest(t)

	id1 := k.generateBidID("auction1", "bidder1", "1000", 1000)
	id2 := k.generateBidID("auction1", "bidder1", "2000", 1000)

	require.NotEqual(t, id1, id2)
}

func TestGetMEVAuctionState_Enabled(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: true,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	state := k.getMEVAuctionState(ctx)
	require.NotNil(t, state)
	require.True(t, state.enabled)
	require.NotNil(t, state.activeAuctions)
}

func TestGetMEVAuctionState_Disabled(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: false,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	state := k.getMEVAuctionState(ctx)
	require.NotNil(t, state)
	require.False(t, state.enabled)
}

func TestGetMEVAuctionState_DefaultValues(t *testing.T) {
	params := types.DefaultParams()
	params.Mev = &types.MEVConfig{
		Enabled: true,
	}

	k, ctx := setupKeeperWithCustomParams(t, params)

	state := k.getMEVAuctionState(ctx)
	require.NotNil(t, state)
	require.Equal(t, "1000000", state.minimumBid)
	require.Equal(t, "5000000", state.reservePrice)
	require.Equal(t, MEVAuctionTypeFirstPrice, state.auctionType)
	require.Equal(t, uint64(0), state.totalAuctions)
	require.Equal(t, "0", state.totalRevenue)
}

func TestMEVAuctionTypes(t *testing.T) {
	require.Equal(t, MEVAuctionType("first_price"), MEVAuctionTypeFirstPrice)
	require.Equal(t, MEVAuctionType("second_price"), MEVAuctionTypeSecondPrice)
	require.Equal(t, MEVAuctionType("vickrey"), MEVAuctionTypeVickrey)
	require.Equal(t, MEVAuctionType("dutch"), MEVAuctionTypeDutch)
}

func TestMEVAuctionBid_Fields(t *testing.T) {
	bid := MEVAuctionBid{
		BidID:     "bid-123",
		Bidder:    "aura1bidder",
		Amount:    "5000000",
		BlockSlot: 100,
		Priority:  10,
		Timestamp: 1000,
		Status:    "active",
	}

	require.Equal(t, "bid-123", bid.BidID)
	require.Equal(t, "aura1bidder", bid.Bidder)
	require.Equal(t, "5000000", bid.Amount)
	require.Equal(t, uint64(100), bid.BlockSlot)
	require.Equal(t, uint64(10), bid.Priority)
	require.Equal(t, int64(1000), bid.Timestamp)
	require.Equal(t, "active", bid.Status)
}

func TestMEVAuction_Fields(t *testing.T) {
	auction := MEVAuction{
		AuctionID:    "mev-123",
		BlockSlot:    100,
		Bids:         []*MEVAuctionBid{},
		WinningBidID: "",
		TotalValue:   "0",
		Status:       "open",
		CreatedAt:    1000,
		ClosedAt:     0,
		MinimumBid:   "1000000",
		ReservePrice: "5000000",
		AuctionType:  MEVAuctionTypeFirstPrice,
	}

	require.Equal(t, "mev-123", auction.AuctionID)
	require.Equal(t, uint64(100), auction.BlockSlot)
	require.Equal(t, "open", auction.Status)
	require.Equal(t, "1000000", auction.MinimumBid)
	require.Equal(t, "5000000", auction.ReservePrice)
	require.Equal(t, MEVAuctionTypeFirstPrice, auction.AuctionType)
}
