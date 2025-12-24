package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// MEV AUCTION MECHANISM (Feature 2)
// ============================

// MEVAuctionType defines the type of auction
type MEVAuctionType string

const (
	MEVAuctionTypeFirstPrice  MEVAuctionType = "first_price"
	MEVAuctionTypeSecondPrice MEVAuctionType = "second_price"
	MEVAuctionTypeVickrey     MEVAuctionType = "vickrey"
	MEVAuctionTypeDutch       MEVAuctionType = "dutch"
)

// MEVAuctionBid represents a bid in the MEV auction
type MEVAuctionBid struct {
	BidID     string
	Bidder    string
	Amount    string
	BlockSlot uint64
	Priority  uint64
	Timestamp int64
	Status    string
}

// MEVAuction represents an MEV auction for a block slot
type MEVAuction struct {
	AuctionID    string
	BlockSlot    uint64
	Bids         []*MEVAuctionBid
	WinningBidID string
	TotalValue   string
	Status       string
	CreatedAt    int64
	ClosedAt     int64
	MinimumBid   string
	ReservePrice string
	AuctionType  MEVAuctionType
}

// In-memory auction state (in production, would use KV store with proto types)
// This is a simplified implementation for the keeper layer
type mevAuctionState struct {
	activeAuctions  map[string]*MEVAuction
	auctionHistory  []*MEVAuction
	totalAuctions   uint64
	totalRevenue    string
	enabled         bool
	minimumBid      string
	reservePrice    string
	auctionType     MEVAuctionType
}

// getMEVAuctionState retrieves or initializes MEV auction state
func (k *Keeper) getMEVAuctionState(ctx context.Context) *mevAuctionState {
	params, _ := k.GetParams(ctx)

	// Check if MEV is enabled in params
	enabled := params.Mev != nil && params.Mev.Enabled

	return &mevAuctionState{
		activeAuctions:  make(map[string]*MEVAuction),
		auctionHistory:  make([]*MEVAuction, 0),
		totalAuctions:   0,
		totalRevenue:    "0",
		enabled:         enabled,
		minimumBid:      "1000000",  // Default 1 token (assuming 6 decimals)
		reservePrice:    "5000000",  // Default 5 tokens
		auctionType:     MEVAuctionTypeFirstPrice,
	}
}

// CreateMEVAuction creates a new MEV auction for a block slot
func (k *Keeper) CreateMEVAuction(ctx context.Context, blockSlot uint64) (string, error) {
	state := k.getMEVAuctionState(ctx)

	if !state.enabled {
		return "", types.ErrMEVAuctionDisabled
	}

	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return "", err
	}

	// Generate auction ID
	auctionID := k.generateAuctionID(blockSlot, currentTime)

	// Create auction record
	auction := &MEVAuction{
		AuctionID:    auctionID,
		BlockSlot:    blockSlot,
		Bids:         make([]*MEVAuctionBid, 0),
		WinningBidID: "",
		TotalValue:   "0",
		Status:       "open",
		CreatedAt:    currentTime,
		ClosedAt:     0,
		MinimumBid:   state.minimumBid,
		ReservePrice: state.reservePrice,
		AuctionType:  state.auctionType,
	}

	state.activeAuctions[auctionID] = auction

	return auctionID, nil
}

// PlaceMEVBid places a bid in an MEV auction
func (k *Keeper) PlaceMEVBid(
	ctx context.Context,
	auctionID, bidder, amount string,
	priority uint64,
) (string, error) {
	state := k.getMEVAuctionState(ctx)

	if !state.enabled {
		return "", types.ErrMEVAuctionDisabled
	}

	auction, exists := state.activeAuctions[auctionID]
	if !exists {
		return "", types.ErrAuctionNotFound
	}

	if auction.Status != "open" {
		return "", types.ErrAuctionClosed
	}

	// Validate bid amount
	bidAmount := new(big.Int)
	if _, ok := bidAmount.SetString(amount, 10); !ok {
		return "", types.ErrInvalidAmount
	}

	minimumBid := new(big.Int)
	minimumBid.SetString(auction.MinimumBid, 10)

	if bidAmount.Cmp(minimumBid) < 0 {
		return "", types.ErrBidTooLow
	}

	// Check if bidder already has a bid
	for _, existingBid := range auction.Bids {
		if existingBid.Bidder == bidder {
			return "", types.ErrBidderAlreadyBid
		}
	}

	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return "", err
	}

	// Create bid
	bidID := k.generateBidID(auctionID, bidder, amount, currentTime)
	bid := &MEVAuctionBid{
		BidID:     bidID,
		Bidder:    bidder,
		Amount:    amount,
		Priority:  priority,
		Timestamp: currentTime,
		Status:    "active",
	}

	// Add bid to auction
	auction.Bids = append(auction.Bids, bid)

	// Update total value
	totalValue := new(big.Int)
	if auction.TotalValue != "" && auction.TotalValue != "0" {
		totalValue.SetString(auction.TotalValue, 10)
	}
	totalValue.Add(totalValue, bidAmount)
	auction.TotalValue = totalValue.String()

	return bidID, nil
}

// CloseMEVAuction closes an MEV auction and determines winner
func (k *Keeper) CloseMEVAuction(
	ctx context.Context,
	auctionID string,
) (winningBid *MEVAuctionBid, winningAmount string, err error) {
	state := k.getMEVAuctionState(ctx)

	if !state.enabled {
		return nil, "", types.ErrMEVAuctionDisabled
	}

	auction, exists := state.activeAuctions[auctionID]
	if !exists {
		return nil, "", types.ErrAuctionNotFound
	}

	if auction.Status != "open" {
		return nil, "", types.ErrAuctionAlreadyClosed
	}

	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return nil, "", err
	}

	if len(auction.Bids) == 0 {
		auction.Status = "closed_no_bids"
		auction.ClosedAt = currentTime
		state.auctionHistory = append(state.auctionHistory, auction)
		delete(state.activeAuctions, auctionID)
		return nil, "0", nil
	}

	// Determine winner based on auction type
	var winner *MEVAuctionBid

	switch auction.AuctionType {
	case MEVAuctionTypeFirstPrice:
		winner = k.selectFirstPriceWinner(auction.Bids)
	case MEVAuctionTypeSecondPrice:
		winner = k.selectSecondPriceWinner(auction.Bids)
	case MEVAuctionTypeVickrey:
		winner = k.selectVickreyWinner(auction.Bids)
	case MEVAuctionTypeDutch:
		winner = k.selectDutchWinner(auction.Bids, auction.ReservePrice)
	default:
		winner = k.selectFirstPriceWinner(auction.Bids)
	}

	if winner == nil {
		auction.Status = "closed_no_winner"
		auction.ClosedAt = currentTime
		state.auctionHistory = append(state.auctionHistory, auction)
		delete(state.activeAuctions, auctionID)
		return nil, "0", nil
	}

	// Mark other bids as lost
	for _, bid := range auction.Bids {
		if bid.BidID != winner.BidID {
			bid.Status = "lost"
		} else {
			bid.Status = "won"
		}
	}

	// Update auction
	auction.WinningBidID = winner.BidID
	auction.Status = "closed"
	auction.ClosedAt = currentTime

	// Add to history
	state.auctionHistory = append(state.auctionHistory, auction)

	// Keep only last 100 auctions in history
	if len(state.auctionHistory) > 100 {
		state.auctionHistory = state.auctionHistory[len(state.auctionHistory)-100:]
	}

	// Update statistics
	state.totalAuctions++
	totalRevenue := new(big.Int)
	if state.totalRevenue != "" && state.totalRevenue != "0" {
		totalRevenue.SetString(state.totalRevenue, 10)
	}
	winningAmountBig := new(big.Int)
	winningAmountBig.SetString(winner.Amount, 10)
	totalRevenue.Add(totalRevenue, winningAmountBig)
	state.totalRevenue = totalRevenue.String()

	// Remove from active auctions
	delete(state.activeAuctions, auctionID)

	return winner, winner.Amount, nil
}

// selectFirstPriceWinner selects winner for first-price auction (highest bid wins, pays bid)
func (k *Keeper) selectFirstPriceWinner(bids []*MEVAuctionBid) *MEVAuctionBid {
	if len(bids) == 0 {
		return nil
	}

	var winner *MEVAuctionBid
	highestBid := big.NewInt(0)

	for _, bid := range bids {
		amount := new(big.Int)
		amount.SetString(bid.Amount, 10)

		if amount.Cmp(highestBid) > 0 {
			highestBid = amount
			winner = bid
		} else if amount.Cmp(highestBid) == 0 && winner != nil && bid.Priority > winner.Priority {
			// Tie-break by priority
			winner = bid
		}
	}

	return winner
}

// selectSecondPriceWinner selects winner for second-price auction (highest bid wins, pays 2nd highest)
func (k *Keeper) selectSecondPriceWinner(bids []*MEVAuctionBid) *MEVAuctionBid {
	if len(bids) == 0 {
		return nil
	}

	// Sort bids by amount descending
	sortedBids := make([]*MEVAuctionBid, len(bids))
	copy(sortedBids, bids)

	sort.Slice(sortedBids, func(i, j int) bool {
		amtI := new(big.Int)
		amtI.SetString(sortedBids[i].Amount, 10)
		amtJ := new(big.Int)
		amtJ.SetString(sortedBids[j].Amount, 10)
		return amtI.Cmp(amtJ) > 0
	})

	winner := sortedBids[0]

	// Winner pays second-highest price
	if len(sortedBids) > 1 {
		winner.Amount = sortedBids[1].Amount
	}

	return winner
}

// selectVickreyWinner selects winner for Vickrey auction (sealed-bid second-price)
func (k *Keeper) selectVickreyWinner(bids []*MEVAuctionBid) *MEVAuctionBid {
	// Same as second-price for our implementation
	return k.selectSecondPriceWinner(bids)
}

// selectDutchWinner selects winner for Dutch auction (descending price)
func (k *Keeper) selectDutchWinner(bids []*MEVAuctionBid, reservePrice string) *MEVAuctionBid {
	// In Dutch auction, first bidder to accept current price wins
	// For simplified implementation, select first bid that meets reserve price
	reserve := new(big.Int)
	reserve.SetString(reservePrice, 10)

	for _, bid := range bids {
		amount := new(big.Int)
		amount.SetString(bid.Amount, 10)

		if amount.Cmp(reserve) >= 0 {
			return bid
		}
	}

	return nil
}

// GetActiveAuctions returns all active MEV auctions
func (k *Keeper) GetActiveAuctions(ctx context.Context) []*MEVAuction {
	state := k.getMEVAuctionState(ctx)
	auctions := make([]*MEVAuction, 0, len(state.activeAuctions))

	for _, auction := range state.activeAuctions {
		auctions = append(auctions, auction)
	}

	return auctions
}

// GetAuction returns a specific MEV auction
func (k *Keeper) GetAuction(ctx context.Context, auctionID string) (*MEVAuction, error) {
	state := k.getMEVAuctionState(ctx)

	// Check active auctions
	if auction, exists := state.activeAuctions[auctionID]; exists {
		return auction, nil
	}

	// Check history
	for _, auction := range state.auctionHistory {
		if auction.AuctionID == auctionID {
			return auction, nil
		}
	}

	return nil, types.ErrAuctionNotFound
}

// GetAuctionStatistics returns MEV auction statistics
func (k *Keeper) GetAuctionStatistics(ctx context.Context) (
	totalCompleted uint64,
	totalRevenue string,
	activeCount uint64,
	avgBid string,
) {
	state := k.getMEVAuctionState(ctx)

	totalCompleted = state.totalAuctions
	totalRevenue = state.totalRevenue
	activeCount = uint64(len(state.activeAuctions))

	// Calculate average winning bid
	avgBid = "0"
	if totalCompleted > 0 {
		revenue := new(big.Int)
		if totalRevenue != "" && totalRevenue != "0" {
			revenue.SetString(totalRevenue, 10)
		}
		avg := new(big.Int).Div(revenue, big.NewInt(int64(totalCompleted)))
		avgBid = avg.String()
	}

	return totalCompleted, totalRevenue, activeCount, avgBid
}

// generateAuctionID generates a unique auction ID
func (k *Keeper) generateAuctionID(blockSlot uint64, timestamp int64) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("auction-%d-%d", blockSlot, timestamp)))
	return "mev-" + hex.EncodeToString(h.Sum(nil))[:16]
}

// generateBidID generates a unique bid ID
func (k *Keeper) generateBidID(auctionID, bidder, amount string, timestamp int64) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s-%s-%s-%d", auctionID, bidder, amount, timestamp)))
	return "bid-" + hex.EncodeToString(h.Sum(nil))[:16]
}

// CancelAuction cancels an active MEV auction
func (k *Keeper) CancelAuction(ctx context.Context, auctionID string, reason string) error {
	state := k.getMEVAuctionState(ctx)

	auction, exists := state.activeAuctions[auctionID]
	if !exists {
		return types.ErrAuctionNotFound
	}

	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return err
	}

	auction.Status = "cancelled"
	auction.ClosedAt = currentTime

	// Mark all bids as cancelled
	for _, bid := range auction.Bids {
		bid.Status = "cancelled"
	}

	// Move to history
	state.auctionHistory = append(state.auctionHistory, auction)

	// Remove from active
	delete(state.activeAuctions, auctionID)

	return nil
}

// GetAuctionHistory returns recent auction history
func (k *Keeper) GetAuctionHistory(ctx context.Context, limit uint64) []*MEVAuction {
	state := k.getMEVAuctionState(ctx)

	if limit == 0 || limit > uint64(len(state.auctionHistory)) {
		limit = uint64(len(state.auctionHistory))
	}

	// Return most recent auctions
	start := uint64(len(state.auctionHistory)) - limit
	return state.auctionHistory[start:]
}

// ConvertAuctionToProto converts internal auction to protobuf compatible format
func (k *Keeper) ConvertAuctionToProto(auction *MEVAuction) map[string]interface{} {
	bids := make([]map[string]interface{}, 0, len(auction.Bids))
	for _, bid := range auction.Bids {
		bids = append(bids, map[string]interface{}{
			"bid_id":    bid.BidID,
			"bidder":    bid.Bidder,
			"amount":    bid.Amount,
			"priority":  bid.Priority,
			"timestamp": time.Unix(bid.Timestamp, 0),
			"status":    bid.Status,
		})
	}

	return map[string]interface{}{
		"auction_id":     auction.AuctionID,
		"block_slot":     auction.BlockSlot,
		"bids":           bids,
		"winning_bid_id": auction.WinningBidID,
		"total_value":    auction.TotalValue,
		"status":         auction.Status,
		"created_at":     time.Unix(auction.CreatedAt, 0),
		"closed_at":      time.Unix(auction.ClosedAt, 0),
		"minimum_bid":    auction.MinimumBid,
		"reserve_price":  auction.ReservePrice,
		"auction_type":   string(auction.AuctionType),
	}
}
