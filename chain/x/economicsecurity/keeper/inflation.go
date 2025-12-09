package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// INFLATION MANAGEMENT
// ============================

// AdjustInflationRate manually adjusts the inflation rate with governance authorization.
//
// This function provides manual control over the inflation rate, allowing governance
// to override automatic adjustments. It enforces strict validation against minimum
// and maximum bounds, requires governance authority, and emits events for transparency.
//
// Security considerations:
//   - Authority validation: Only the governance authority can call this function
//   - Rate bounds: Enforces min/max inflation rate limits from params
//   - Audit trail: Emits events with old rate, new rate, and reason
//   - State consistency: Updates both params and previous inflation tracking
//
// Parameters:
//   - ctx: SDK context for state access and event emission
//   - authority: Address attempting the adjustment (must match keeper authority)
//   - newRate: New inflation rate in basis points (e.g., 500 = 5.00%)
//   - reason: Human-readable justification for the adjustment
//
// Returns:
//   - oldRate: Previous inflation rate before adjustment
//   - error: ErrUnauthorized, ErrInflationRateTooHigh, ErrInflationRateTooLow, or state errors
//
// Events emitted:
//   - EventTypeInflationAdjusted with old_rate, new_rate, reason, authority
func (k Keeper) AdjustInflationRate(ctx context.Context, authority string, newRate uint64, reason string) (oldRate uint64, err error) {
	// Validate authority - only governance can manually adjust inflation
	if authority != k.authority {
		return 0, types.ErrUnauthorized
	}

	// Validate reason is provided for audit trail
	if reason == "" {
		return 0, fmt.Errorf("reason is required for inflation rate adjustment")
	}

	// Get current params to access bounds and current rate
	params := k.GetParams()
	if params.Tokenomics == nil {
		return 0, fmt.Errorf("tokenomics config not initialized")
	}

	oldRate = params.Tokenomics.InflationRate

	// Validate new rate against configured bounds
	// This prevents governance from setting economically dangerous rates
	if newRate > params.Tokenomics.MaxInflationRate {
		return oldRate, fmt.Errorf("%w: new rate %d exceeds maximum %d",
			types.ErrInflationRateTooHigh, newRate, params.Tokenomics.MaxInflationRate)
	}

	if newRate < params.Tokenomics.MinInflationRate {
		return oldRate, fmt.Errorf("%w: new rate %d below minimum %d",
			types.ErrInflationRateTooLow, newRate, params.Tokenomics.MinInflationRate)
	}

	// Validate rate is reasonable (not equal to current, as that would be pointless)
	if newRate == oldRate {
		return oldRate, fmt.Errorf("new rate must differ from current rate %d", oldRate)
	}

	// Store previous rate for 24h change calculations
	if err := k.SetPreviousInflation(ctx, oldRate); err != nil {
		return oldRate, fmt.Errorf("failed to store previous inflation: %w", err)
	}

	// Update inflation rate in params
	params.Tokenomics.InflationRate = newRate
	params.Tokenomics.LastInflationAdjustment = time.Now()

	if err := k.SetParams(params); err != nil {
		return oldRate, fmt.Errorf("failed to update params: %w", err)
	}

	// Emit event for transparency and audit trail
	// This is critical for governance accountability
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeInflationAdjusted,
			sdk.NewAttribute(types.AttributeKeyOldRate, fmt.Sprintf("%d", oldRate)),
			sdk.NewAttribute(types.AttributeKeyNewRate, fmt.Sprintf("%d", newRate)),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
			sdk.NewAttribute(types.AttributeKeyAuthority, authority),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyTimestamp, sdkCtx.BlockTime().Format(time.RFC3339)),
		),
	)

	return oldRate, nil
}

// GetInflationMetrics returns comprehensive inflation metrics for monitoring and queries.
//
// This function provides a complete view of the inflation state including current rate,
// target rate, 24-hour change, and timing information. It's used by queries, dashboards,
// and monitoring systems to track inflation health.
//
// Calculation notes:
//   - 24h change: Computed by comparing current rate to previous rate stored 24h ago
//   - Next check: Calculated based on last check time + check interval
//   - Current height: Retrieved from keeper state
//   - Block time: Retrieved from SDK context
//
// Parameters:
//   - ctx: SDK context for state access
//
// Returns:
//   - currentRate: Current inflation rate in basis points
//   - targetRate: Target inflation rate from params
//   - change24h: Rate change over last 24 hours (signed integer, can be negative)
//   - lastAdjustment: Timestamp of last manual or automatic adjustment
//   - nextCheck: Projected timestamp for next automatic check
//   - error: State access errors or context errors
func (k Keeper) GetInflationMetrics(ctx context.Context) (
	currentRate uint64,
	targetRate uint64,
	change24h int64,
	lastAdjustment *timestamppb.Timestamp,
	nextCheck *timestamppb.Timestamp,
	err error,
) {
	// Get current params
	params := k.GetParams()
	if params.Tokenomics == nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("tokenomics config not initialized")
	}

	currentRate = params.Tokenomics.InflationRate
	targetRate = params.Tokenomics.TargetInflationRate
	lastAdjustment = timestamppb.New(params.Tokenomics.LastInflationAdjustment)

	// Calculate 24h change
	change24h, err = k.CalculateInflation24hChange(ctx)
	if err != nil {
		// Log error but continue - change might not be available yet
		change24h = 0
	}

	// Calculate next check time based on check interval
	// Check interval is in blocks, so we need to estimate time
	if !params.Tokenomics.LastInflationCheck.IsZero() {
		checkInterval := time.Duration(params.InflationCheckInterval) * 6 * time.Second // Assuming 6s block time
		nextCheckTime := params.Tokenomics.LastInflationCheck.Add(checkInterval)
		nextCheck = timestamppb.New(nextCheckTime)
	} else {
		// If never checked, use current time
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		nextCheck = timestamppb.New(sdkCtx.BlockTime())
	}

	return currentRate, targetRate, change24h, lastAdjustment, nextCheck, nil
}

// CalculateInflation24hChange computes the inflation rate change over the last 24 hours.
//
// This function calculates the difference between the current inflation rate and the
// rate from 24 hours ago. It's used for monitoring rate volatility and triggering
// alerts when changes exceed thresholds.
//
// Calculation method:
//   - Retrieves previous inflation rate stored 24h ago
//   - Compares with current rate from params
//   - Returns signed difference (positive = rate increased, negative = rate decreased)
//
// Edge cases:
//   - If no previous rate exists (chain just started): returns 0
//   - If previous rate is same as current: returns 0
//   - Handles wraparound correctly using signed integer arithmetic
//
// Parameters:
//   - ctx: SDK context for state access
//
// Returns:
//   - change: Signed rate change in basis points (can be negative)
//   - error: State access errors
//
// Example:
//   - Previous rate: 500 (5.00%), Current rate: 550 (5.50%) → Returns: 50
//   - Previous rate: 600 (6.00%), Current rate: 500 (5.00%) → Returns: -100
func (k Keeper) CalculateInflation24hChange(ctx context.Context) (int64, error) {
	// Get current inflation rate
	params := k.GetParams()
	if params.Tokenomics == nil {
		return 0, fmt.Errorf("tokenomics config not initialized")
	}
	currentRate := params.Tokenomics.InflationRate

	// Get previous inflation rate (stored 24h ago)
	previousRate, err := k.GetPreviousInflation(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get previous inflation: %w", err)
	}

	// If no previous rate exists, return 0 (chain just started)
	if previousRate == 0 {
		return 0, nil
	}

	// Calculate change as signed integer
	// This correctly handles both increases and decreases
	change := int64(currentRate) - int64(previousRate)

	return change, nil
}

// UpdateInflationCheckTimestamp updates the last inflation check timestamp.
//
// This is a helper function used by automated inflation adjustment mechanisms
// to track when the last check occurred. It's called by BeginBlock or other
// periodic processes.
//
// Parameters:
//   - ctx: SDK context for state access
//
// Returns:
//   - error: State update errors
func (k Keeper) UpdateInflationCheckTimestamp(ctx context.Context) error {
	params := k.GetParams()
	if params.Tokenomics == nil {
		return fmt.Errorf("tokenomics config not initialized")
	}

	params.Tokenomics.LastInflationCheck = timestamppb.Now()

	if err := k.SetParams(params); err != nil {
		return fmt.Errorf("failed to update check timestamp: %w", err)
	}

	return nil
}
