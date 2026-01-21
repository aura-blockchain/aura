// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"

	"cosmossdk.io/math"
	"github.com/aequitas/aura/chain/x/confidencescore/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================
// SCORE MARKETPLACE
// Peer-to-peer marketplace for confidence score trading
// ============================

// ListingType defines types of marketplace listings
type ListingType string

const (
	ListingTypeSale    ListingType = "sale"    // Direct sale
	ListingTypeLease   ListingType = "lease"   // Temporary lease
	ListingTypeAuction ListingType = "auction" // Auction
)

// ScoreListing represents a marketplace listing
type ScoreListing struct {
	ListingID     string
	Seller        string
	ListingType   ListingType
	ScoreAmount   uint64
	PricePerPoint math.Int // Price per score point in uaura
	TotalPrice    math.Int
	LeaseDuration uint64   // blocks (for lease type)
	MinBid        math.Int // minimum bid (for auction type)
	CurrentBid    math.Int
	HighestBidder string
	ExpiresHeight uint64
	Active        bool
	CreatedHeight uint64
}

// ScorePurchase represents a completed transaction
type ScorePurchase struct {
	PurchaseID  string
	ListingID   string
	Buyer       string
	Seller      string
	ScoreAmount uint64
	Price       math.Int
	BlockHeight uint64
}

// CreateListing creates a new marketplace listing
func (k *Keeper) CreateListing(
	ctx sdk.Context,
	seller string,
	listingType ListingType,
	scoreAmount uint64,
	pricePerPoint math.Int,
	leaseDuration uint64,
	expirationBlocks uint64,
) (*ScoreListing, error) {
	// Validate inputs
	if seller == "" {
		return nil, types.ErrInvalidWalletAddress
	}

	if scoreAmount == 0 {
		return nil, fmt.Errorf("score amount must be positive")
	}

	if !pricePerPoint.IsPositive() {
		return nil, fmt.Errorf("price must be positive")
	}

	// Get seller record
	record, ok := k.GetUserRecord(ctx, seller)
	if !ok {
		return nil, types.ErrUserRecordNotFound
	}

	// Check if seller has enough score
	if record.TotalScore < scoreAmount {
		return nil, fmt.Errorf("insufficient score: have %d, need %d", record.TotalScore, scoreAmount)
	}

	// Check marketplace parameters
	if !false {
		return nil, fmt.Errorf("marketplace is disabled")
	}

	// Calculate total price
	totalPrice := pricePerPoint.Mul(math.NewInt(int64(scoreAmount)))

	// Create listing
	currentHeight := uint64(ctx.BlockHeight())
	listingID := fmt.Sprintf("listing-%s-%d", seller, currentHeight)
	expiresHeight := currentHeight + expirationBlocks

	listing := &ScoreListing{
		ListingID:     listingID,
		Seller:        seller,
		ListingType:   listingType,
		ScoreAmount:   scoreAmount,
		PricePerPoint: pricePerPoint,
		TotalPrice:    totalPrice,
		LeaseDuration: leaseDuration,
		MinBid:        totalPrice,
		CurrentBid:    math.ZeroInt(),
		HighestBidder: "",
		ExpiresHeight: expiresHeight,
		Active:        true,
		CreatedHeight: currentHeight,
	}

	// Lock seller's score
	record.TotalScore -= scoreAmount
	if err := k.SetUserRecord(ctx, record); err != nil {
		return nil, err
	}

	// Store listing
	k.storeListing(ctx, listing)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"score_listed",
			sdk.NewAttribute("listing_id", listingID),
			sdk.NewAttribute("seller", seller),
			sdk.NewAttribute("type", string(listingType)),
			sdk.NewAttribute("amount", fmt.Sprintf("%d", scoreAmount)),
			sdk.NewAttribute("price", totalPrice.String()),
		),
	)

	return listing, nil
}

// PurchaseListing executes a purchase from a listing
func (k *Keeper) PurchaseListing(
	ctx sdk.Context,
	listingID string,
	buyer string,
	bankKeeper BankKeeper,
) (*ScorePurchase, error) {
	listing, ok := k.getListing(ctx, listingID)
	if !ok {
		return nil, fmt.Errorf("listing not found")
	}

	if !listing.Active {
		return nil, fmt.Errorf("listing is not active")
	}

	if listing.ListingType != ListingTypeSale {
		return nil, fmt.Errorf("invalid listing type for purchase")
	}

	if uint64(ctx.BlockHeight()) > listing.ExpiresHeight {
		return nil, fmt.Errorf("listing has expired")
	}

	if buyer == listing.Seller {
		return nil, fmt.Errorf("cannot purchase own listing")
	}

	// Transfer payment from buyer to seller
	buyerAddr, err := sdk.AccAddressFromBech32(buyer)
	if err != nil {
		return nil, err
	}

	sellerAddr, err := sdk.AccAddressFromBech32(listing.Seller)
	if err != nil {
		return nil, err
	}

	// Calculate marketplace fee
	feePercent := uint64(2)
	fee := listing.TotalPrice.Mul(math.NewInt(int64(feePercent))).Quo(math.NewInt(100))
	sellerAmount := listing.TotalPrice.Sub(fee)

	// Transfer coins from buyer to module
	paymentCoins := sdk.NewCoins(sdk.NewCoin("uaura", listing.TotalPrice))
	if err := bankKeeper.SendCoinsFromAccountToModule(ctx, buyerAddr, types.ModuleName, paymentCoins); err != nil {
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	// Send to seller (minus fee)
	sellerCoins := sdk.NewCoins(sdk.NewCoin("uaura", sellerAmount))
	if err := bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sellerAddr, sellerCoins); err != nil {
		return nil, fmt.Errorf("seller payment failed: %w", err)
	}

	// Fee stays in module (could be distributed to stakers, treasury, etc.)

	// Transfer score to buyer
	buyerRecord := k.GetOrCreateUserRecord(ctx, buyer)
	buyerRecord.TotalScore += listing.ScoreAmount
	if err := k.SetUserRecord(ctx, buyerRecord); err != nil {
		return nil, err
	}

	// Mark listing as inactive
	listing.Active = false
	k.storeListing(ctx, listing)

	// Create purchase record
	purchase := &ScorePurchase{
		PurchaseID:  fmt.Sprintf("purchase-%s-%d", buyer, ctx.BlockHeight()),
		ListingID:   listingID,
		Buyer:       buyer,
		Seller:      listing.Seller,
		ScoreAmount: listing.ScoreAmount,
		Price:       listing.TotalPrice,
		BlockHeight: uint64(ctx.BlockHeight()),
	}

	k.storePurchase(ctx, purchase)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"score_purchased",
			sdk.NewAttribute("purchase_id", purchase.PurchaseID),
			sdk.NewAttribute("listing_id", listingID),
			sdk.NewAttribute("buyer", buyer),
			sdk.NewAttribute("seller", listing.Seller),
			sdk.NewAttribute("amount", fmt.Sprintf("%d", listing.ScoreAmount)),
			sdk.NewAttribute("price", listing.TotalPrice.String()),
			sdk.NewAttribute("fee", fee.String()),
		),
	)

	return purchase, nil
}

// PlaceBid places a bid on an auction listing
func (k *Keeper) PlaceBid(
	ctx sdk.Context,
	listingID string,
	bidder string,
	bidAmount math.Int,
	bankKeeper BankKeeper,
) error {
	listing, ok := k.getListing(ctx, listingID)
	if !ok {
		return fmt.Errorf("listing not found")
	}

	if !listing.Active {
		return fmt.Errorf("listing is not active")
	}

	if listing.ListingType != ListingTypeAuction {
		return fmt.Errorf("listing is not an auction")
	}

	if uint64(ctx.BlockHeight()) > listing.ExpiresHeight {
		return fmt.Errorf("auction has ended")
	}

	if bidder == listing.Seller {
		return fmt.Errorf("cannot bid on own auction")
	}

	// Check if bid is higher than current
	if bidAmount.LTE(listing.CurrentBid) && !listing.CurrentBid.IsZero() {
		return fmt.Errorf("bid must be higher than current bid: %s", listing.CurrentBid.String())
	}

	if bidAmount.LT(listing.MinBid) {
		return fmt.Errorf("bid must be at least minimum bid: %s", listing.MinBid.String())
	}

	// Lock bid amount
	bidderAddr, err := sdk.AccAddressFromBech32(bidder)
	if err != nil {
		return fmt.Errorf("failed to AccAddressFromBech32 for bid: %w", err)
	}

	bidCoins := sdk.NewCoins(sdk.NewCoin("uaura", bidAmount))
	if err := bankKeeper.SendCoinsFromAccountToModule(ctx, bidderAddr, types.ModuleName, bidCoins); err != nil {
		return fmt.Errorf("failed to lock bid: %w", err)
	}

	// Return previous bidder's funds
	if listing.HighestBidder != "" && listing.CurrentBid.IsPositive() {
		previousBidderAddr, err := sdk.AccAddressFromBech32(listing.HighestBidder)
		if err == nil {
			previousBidCoins := sdk.NewCoins(sdk.NewCoin("uaura", listing.CurrentBid))
			if err := bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, previousBidderAddr, previousBidCoins); err != nil {
				// Log error but continue - this shouldn't block the new bid
				ctx.Logger().Error("failed to return previous bidder's funds", "error", err)
			}
		}
	}

	// Update listing
	listing.CurrentBid = bidAmount
	listing.HighestBidder = bidder
	k.storeListing(ctx, listing)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bid_placed",
			sdk.NewAttribute("listing_id", listingID),
			sdk.NewAttribute("bidder", bidder),
			sdk.NewAttribute("amount", bidAmount.String()),
		),
	)

	return nil
}

// CancelListing cancels an active listing
func (k *Keeper) CancelListing(ctx sdk.Context, listingID string, caller string) error {
	listing, ok := k.getListing(ctx, listingID)
	if !ok {
		return fmt.Errorf("listing not found")
	}

	if !listing.Active {
		return fmt.Errorf("listing is not active")
	}

	if listing.Seller != caller {
		return types.ErrUnauthorized
	}

	// Return score to seller
	sellerRecord, ok := k.GetUserRecord(ctx, listing.Seller)
	if ok {
		sellerRecord.TotalScore += listing.ScoreAmount
		_ = k.SetUserRecord(ctx, sellerRecord) // Best effort
	}

	// Mark as inactive
	listing.Active = false
	k.storeListing(ctx, listing)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"listing_cancelled",
			sdk.NewAttribute("listing_id", listingID),
			sdk.NewAttribute("seller", listing.Seller),
		),
	)

	return nil
}

// GetActiveListings returns all active marketplace listings
func (k *Keeper) GetActiveListings(ctx sdk.Context, listingType ListingType) []ScoreListing {
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.MarketplaceListingStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return []ScoreListing{}
	}
	defer iterator.Close()

	listings := []ScoreListing{}
	for ; iterator.Valid(); iterator.Next() {
		// In production, properly unmarshal from protobuf
		listing := ScoreListing{Active: true}
		if listing.Active && (listingType == "" || listing.ListingType == listingType) {
			listings = append(listings, listing)
		}
	}

	return listings
}

// GetUserPurchases returns purchase history for a user
func (k *Keeper) GetUserPurchases(ctx sdk.Context, walletAddr string) []ScorePurchase {
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.MarketplacePurchaseStoreKeyPrefix)
	iterator, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return []ScorePurchase{}
	}
	defer iterator.Close()

	purchases := []ScorePurchase{}
	for ; iterator.Valid(); iterator.Next() {
		// In production, properly unmarshal from protobuf
		purchase := ScorePurchase{}
		if purchase.Buyer == walletAddr || purchase.Seller == walletAddr {
			purchases = append(purchases, purchase)
		}
	}

	return purchases
}

// Helper functions - in production, these would use proper KV store persistence

func (k *Keeper) storeListing(ctx sdk.Context, listing *ScoreListing) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.MarketplaceListingStoreKey(listing.ListingID)
	// In production, properly marshal to protobuf
	if err := store.Set([]byte(key), []byte(listing.Seller)); err != nil {
		ctx.Logger().Error("failed to store listing", "listing_id", listing.ListingID, "error", err)
	}
}

func (k *Keeper) getListing(ctx sdk.Context, listingID string) (*ScoreListing, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.MarketplaceListingStoreKey(listingID)
	bz, err := store.Get([]byte(key))
	if err != nil || len(bz) == 0 {
		return nil, false
	}
	// In production, properly unmarshal from protobuf
	return &ScoreListing{ListingID: listingID, Active: true}, true
}

func (k *Keeper) storePurchase(ctx sdk.Context, purchase *ScorePurchase) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.MarketplacePurchaseStoreKey(purchase.PurchaseID)
	// In production, properly marshal to protobuf
	if err := store.Set([]byte(key), []byte(purchase.Buyer)); err != nil {
		ctx.Logger().Error("failed to store purchase", "purchase_id", purchase.PurchaseID, "error", err)
	}
}
