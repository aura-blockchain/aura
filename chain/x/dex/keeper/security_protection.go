// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/dex/types"
)

// ============================================================================
// FRONT-RUNNING & MEV PROTECTION
// Extracted from security.go for better code organization
// ============================================================================

// CheckFrontRunningProtection ensures minimum block delay between swaps
func (k Keeper) CheckFrontRunningProtection(
	ctx sdk.Context,
	address string,
	poolID string,
) error {
	params := k.GetSecurityParams(ctx)
	if params.MinBlockDelay == 0 {
		return nil // Protection disabled
	}

	// Get last trade for this address/pool
	lastTrade := k.GetLastTradeBlock(ctx, address, poolID)
	currentBlock := uint64(ctx.BlockHeight())

	// If no previous trade (lastTrade == 0), allow the first trade
	if lastTrade > 0 && currentBlock-lastTrade < params.MinBlockDelay {
		return fmt.Errorf(
			"must wait %d blocks between trades (last: %d, current: %d): %w",
			params.MinBlockDelay,
			lastTrade,
			currentBlock,
			types.ErrFrontRunningDetected,
		)
	}

	return nil
}

// RecordTradeBlock records the block height of a trade
func (k Keeper) RecordTradeBlock(ctx sdk.Context, address string, poolID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.TradeBlockKey(address, poolID)
	height := uint64(ctx.BlockHeight())

	store.Set(key, sdk.Uint64ToBigEndian(height))

	// Update trade history
	k.UpdateTradeHistory(ctx, address, poolID)
}

// GetLastTradeBlock retrieves last trade block for address/pool
func (k Keeper) GetLastTradeBlock(ctx sdk.Context, address string, poolID string) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.TradeBlockKey(address, poolID)

	bz := store.Get(key)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// CheckMEVProtection implements MEV mitigation by address
func (k Keeper) CheckMEVProtection(ctx sdk.Context, address string) error {
	params := k.GetSecurityParams(ctx)
	if !params.MevProtectionEnabled {
		return nil
	}

	// Check maximum swaps per block
	swapsInBlock := k.GetSwapsInCurrentBlock(ctx, address)
	if swapsInBlock >= params.MaxSwapsPerBlock {
		return fmt.Errorf(
			"maximum swaps per block exceeded: %d >= %d: %w",
			swapsInBlock,
			params.MaxSwapsPerBlock,
			types.ErrMEVDetected,
		)
	}

	return nil
}

// GetSwapsInCurrentBlock counts swaps by address in current block
func (k Keeper) GetSwapsInCurrentBlock(ctx sdk.Context, address string) uint64 {
	history := k.GetTradeHistory(ctx, address)
	if history == nil {
		return 0
	}

	if history.LastTradeBlock == ctx.BlockHeight() {
		return history.TradesInBlock
	}

	return 0
}

// ============================================================================
// FLASH LOAN ATTACK PROTECTION
// ============================================================================

// CheckFlashLoanProtection prevents rapid add/remove liquidity
func (k Keeper) CheckFlashLoanProtection(
	ctx sdk.Context,
	provider string,
	poolID string,
	isAdding bool,
) error {
	params := k.GetSecurityParams(ctx)
	if params.MinLiquidityBlocks == 0 {
		return nil // Protection disabled
	}

	lastBlock := k.GetLastLiquidityBlock(ctx, provider, poolID)
	currentBlock := uint64(ctx.BlockHeight())

	if lastBlock > 0 && currentBlock-lastBlock < params.MinLiquidityBlocks {
		return fmt.Errorf(
			"must wait %d blocks between liquidity operations (last: %d, current: %d): %w",
			params.MinLiquidityBlocks,
			lastBlock,
			currentBlock,
			types.ErrFlashLoanDetected,
		)
	}

	return nil
}

// RecordLiquidityBlock records the block height of liquidity operation
func (k Keeper) RecordLiquidityBlock(ctx sdk.Context, provider string, poolID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.LiquidityBlockKey(provider, poolID)
	height := uint64(ctx.BlockHeight())

	store.Set(key, sdk.Uint64ToBigEndian(height))
}

// GetLastLiquidityBlock retrieves last liquidity operation block
func (k Keeper) GetLastLiquidityBlock(ctx sdk.Context, provider string, poolID string) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.LiquidityBlockKey(provider, poolID)

	bz := store.Get(key)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// ============================================================================
// SLIPPAGE & TRADE SIZE LIMITS
// ============================================================================

// CheckPoolSlippageLimit validates slippage against pool-specific limits
func (k Keeper) CheckPoolSlippageLimit(
	ctx sdk.Context,
	poolID string,
	priceImpact sdkmath.LegacyDec,
) error {
	params := k.GetSecurityParams(ctx)

	// MaxPriceImpactPercent is already LegacyDec
	maxPriceImpact := params.MaxPriceImpactPercent

	// Convert price impact percentage to decimal (5% = 5.0)
	if priceImpact.GT(maxPriceImpact) {
		return fmt.Errorf(
			"price impact %s%% exceeds maximum %s%%: %w",
			priceImpact.String(),
			maxPriceImpact.String(),
			types.ErrPriceImpactTooHigh,
		)
	}

	return nil
}

// CheckMaxTradeSize validates trade size against pool reserves
func (k Keeper) CheckMaxTradeSize(
	ctx sdk.Context,
	poolID string,
	amountIn sdkmath.Int,
) error {
	params := k.GetSecurityParams(ctx)
	pool := k.GetPool(ctx, poolID)
	if pool == nil {
		return types.ErrPoolNotFound
	}

	// Parse reserve
	reserveA, err := k.parseReserve(pool.ReserveA)
	if err != nil {
		return fmt.Errorf("failed to parseReserve: %w", err)
	}

	// MaxTradeSizePercent is already LegacyDec
	maxTradeSizePercent := params.MaxTradeSizePercent

	// Determine reserve based on input denom
	// Assuming input is denomA, check against reserveA
	maxAmount := reserveA.ToLegacyDec().Mul(maxTradeSizePercent).TruncateInt()

	if amountIn.GT(maxAmount) {
		return fmt.Errorf(
			"trade size %s exceeds maximum %s (%s%% of pool): %w",
			amountIn.String(),
			maxAmount.String(),
			maxTradeSizePercent.MulInt64(100).String(),
			types.ErrTradeTooLarge,
		)
	}

	return nil
}

// CheckPriceImpactThreshold validates price impact
func (k Keeper) CheckPriceImpactThreshold(
	ctx sdk.Context,
	priceImpact sdkmath.LegacyDec,
) error {
	params := k.GetSecurityParams(ctx)

	// MaxPriceImpactPercent is already LegacyDec
	maxPriceImpact := params.MaxPriceImpactPercent

	if priceImpact.GT(maxPriceImpact) {
		return fmt.Errorf(
			"price impact %s%% exceeds threshold %s%%: %w",
			priceImpact.String(),
			maxPriceImpact.String(),
			types.ErrPriceImpactTooHigh,
		)
	}

	return nil
}

// CheckDustAttack prevents dust attacks with minimum trade amounts
func (k Keeper) CheckDustAttack(ctx sdk.Context, amountIn sdkmath.Int) error {
	params := k.GetSecurityParams(ctx)

	// MinTradeAmount is already Int
	minTradeAmount := params.MinTradeAmount

	if amountIn.LT(minTradeAmount) {
		return fmt.Errorf(
			"trade amount %s below minimum %s: %w",
			amountIn.String(),
			minTradeAmount.String(),
			types.ErrDustAttack,
		)
	}

	return nil
}

// ============================================================================
// LIQUIDITY LOCK-UP PERIODS (PREVENTS RUG PULLS)
// ============================================================================

// CreateLiquidityLock locks LP tokens for specified period
func (k Keeper) CreateLiquidityLock(
	ctx sdk.Context,
	provider string,
	poolID string,
	lpTokens sdkmath.Int,
) error {
	params := k.GetSecurityParams(ctx)
	if params.LiquidityLockupSeconds == 0 {
		return nil // Lockup disabled
	}

	lockEnd := ctx.BlockTime().Add(time.Duration(params.LiquidityLockupSeconds) * time.Second)

	lock := &types.LiquidityLock{
		PoolId:         poolID,
		Provider:       provider,
		LockedLpTokens: lpTokens,
		LockStart:      ctx.BlockTime(),
		LockEnd:        lockEnd,
		IsActive:       true,
	}

	if err := k.SetLiquidityLock(ctx, lock); err != nil {
		return fmt.Errorf("error in CreateLiquidityLock for LiquidityLock: %w", err)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLiquidityLocked,
			sdk.NewAttribute(types.AttributeKeyPoolID, poolID),
			sdk.NewAttribute(types.AttributeKeyProvider, provider),
			sdk.NewAttribute(types.AttributeKeyLPTokens, lpTokens.String()),
			sdk.NewAttribute("lock_end", lockEnd.String()),
		),
	)

	return nil
}

// CheckLiquidityLock verifies if liquidity can be removed
func (k Keeper) CheckLiquidityLock(
	ctx sdk.Context,
	provider string,
	poolID string,
	lpTokens sdkmath.Int,
) error {
	lock := k.GetLiquidityLock(ctx, provider, poolID)
	if lock == nil || !lock.IsActive {
		return nil // No active lock
	}

	lockEndTime := lock.LockEnd
	if ctx.BlockTime().Before(lockEndTime) {
		return fmt.Errorf(
			"liquidity locked until %s (current: %s): %w",
			lockEndTime.String(),
			ctx.BlockTime().String(),
			types.ErrLiquidityLocked,
		)
	}

	// Lock expired, mark as inactive
	lock.IsActive = false
	if err := k.SetLiquidityLock(ctx, lock); err != nil {
		return fmt.Errorf("error in CheckLiquidityLock for liquidity: %w", err)
	}

	return nil
}

// SetLiquidityLock stores a liquidity lock
func (k Keeper) SetLiquidityLock(ctx sdk.Context, lock *types.LiquidityLock) error {
	store := ctx.KVStore(k.storeKey)
	key := types.LiquidityLockKey(lock.Provider, lock.PoolId)

	bz, err := k.cdc.Marshal(lock)
	if err != nil {
		return types.ErrMarshalFailed.Wrapf("failed to marshal liquidity lock for provider %s pool %s: %v", lock.Provider, lock.PoolId, err)
	}
	store.Set(key, bz)
	return nil
}

// GetLiquidityLock retrieves a liquidity lock
func (k Keeper) GetLiquidityLock(ctx sdk.Context, provider string, poolID string) *types.LiquidityLock {
	store := ctx.KVStore(k.storeKey)
	key := types.LiquidityLockKey(provider, poolID)

	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	var lock types.LiquidityLock
	if err := k.cdc.Unmarshal(bz, &lock); err != nil {
		ctx.Logger().Error("failed to unmarshal liquidity lock",
			"provider", provider,
			"pool_id", poolID,
			"error", err,
			"data_len", len(bz))
		return nil
	}
	return &lock
}
