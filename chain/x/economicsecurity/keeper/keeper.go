package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/aequitas/aura/chain/x/economicsecurity/params"
	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Keeper manages the economic security module state
type Keeper struct {
	mu sync.RWMutex

	// Parameter store
	paramsStore *params.Store

	// Vesting
	vestingSchedules map[string]*types.VestingSchedule // schedule_id -> VestingSchedule
	userVestings     map[string][]string               // beneficiary -> []schedule_id

	// Voting locks
	voteLocks     map[string]*types.VoteLock // lock_id -> VoteLock
	userVoteLocks map[string][]string        // owner -> []lock_id

	// Treasury
	pendingTreasuryTxs map[string]*types.PendingTreasuryTx // tx_id -> PendingTreasuryTx

	// Inflation monitoring
	inflationAlerts   []*types.InflationAlert
	previousInflation uint64

	// Whale protection
	largeTxRecords   []*types.LargeTxRecord
	lastLargeTxTimes map[string]int64    // address -> timestamp
	addressHoldings  map[string]*big.Int // address -> amount

	// MEV redistribution
	userMEVBalances map[string]*big.Int // address -> balance
	totalMEVPending *big.Int

	// Dynamic fees
	blockUtilization []uint64

	// Current state
	currentHeight uint64
	currentTime   int64
	totalBurned   *big.Int
}

// NewKeeper creates a new Keeper instance
func NewKeeper(store *params.Store) *Keeper {
	if store == nil {
		store = params.NewStore(*types.DefaultParams())
	}

	return &Keeper{
		paramsStore:        store,
		vestingSchedules:   make(map[string]*types.VestingSchedule),
		userVestings:       make(map[string][]string),
		voteLocks:          make(map[string]*types.VoteLock),
		userVoteLocks:      make(map[string][]string),
		pendingTreasuryTxs: make(map[string]*types.PendingTreasuryTx),
		inflationAlerts:    []*types.InflationAlert{},
		largeTxRecords:     []*types.LargeTxRecord{},
		lastLargeTxTimes:   make(map[string]int64),
		addressHoldings:    make(map[string]*big.Int),
		userMEVBalances:    make(map[string]*big.Int),
		totalMEVPending:    big.NewInt(0),
		blockUtilization:   []uint64{},
		currentTime:        time.Now().Unix(),
		totalBurned:        big.NewInt(0),
	}
}

// SetCurrentHeight sets the current block height
func (k *Keeper) SetCurrentHeight(height uint64) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.currentHeight = height
}

// SetCurrentTime sets the current time
func (k *Keeper) SetCurrentTime(t int64) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.currentTime = t
}

// GetParams returns the current module parameters
func (k *Keeper) GetParams() types.Params {
	return k.paramsStore.GetParams()
}

// SetParams sets new module parameters
func (k *Keeper) SetParams(params types.Params) error {
	return k.paramsStore.SetParams(params)
}

// ============================
// SUPPLY CAP ENFORCEMENT (Feature 1)
// ============================

// CheckSupplyCap validates that minting won't exceed max supply
func (k *Keeper) CheckSupplyCap(mintAmount string) error {
	params := k.GetParams()

	maxSupply := new(big.Int)
	if _, ok := maxSupply.SetString(params.Tokenomics.MaxSupply, 10); !ok {
		return types.ErrInvalidSupplyCap
	}

	currentSupply := new(big.Int)
	if _, ok := currentSupply.SetString(params.Tokenomics.CirculatingSupply, 10); !ok {
		return types.ErrInvalidAmount
	}

	mintAmt := new(big.Int)
	if _, ok := mintAmt.SetString(mintAmount, 10); !ok {
		return types.ErrInvalidAmount
	}

	newSupply := new(big.Int).Add(currentSupply, mintAmt)

	if newSupply.Cmp(maxSupply) > 0 {
		return types.ErrMaxSupplyExceeded
	}

	return nil
}

// UpdateCirculatingSupply updates the circulating supply (must check cap first)
func (k *Keeper) UpdateCirculatingSupply(delta string, increase bool) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	params := k.GetParams()

	currentSupply := new(big.Int)
	if _, ok := currentSupply.SetString(params.Tokenomics.CirculatingSupply, 10); !ok {
		return types.ErrInvalidAmount
	}

	deltaAmt := new(big.Int)
	if _, ok := deltaAmt.SetString(delta, 10); !ok {
		return types.ErrInvalidAmount
	}

	var newSupply *big.Int
	if increase {
		newSupply = new(big.Int).Add(currentSupply, deltaAmt)
	} else {
		newSupply = new(big.Int).Sub(currentSupply, deltaAmt)
		if newSupply.Cmp(big.NewInt(0)) < 0 {
			return types.ErrInvalidAmount
		}
	}

	params.Tokenomics.CirculatingSupply = newSupply.String()
	return k.SetParams(params)
}

// GetRemainingSupply returns how much supply is available to mint
func (k *Keeper) GetRemainingSupply() string {
	params := k.GetParams()

	maxSupply := new(big.Int)
	maxSupply.SetString(params.Tokenomics.MaxSupply, 10)

	currentSupply := new(big.Int)
	currentSupply.SetString(params.Tokenomics.CirculatingSupply, 10)

	remaining := new(big.Int).Sub(maxSupply, currentSupply)
	return remaining.String()
}

// ============================
// INFLATION MONITORING (Feature 2)
// ============================

// CheckInflation monitors inflation and creates alerts if needed
func (k *Keeper) CheckInflation() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	params := k.GetParams()
	currentRate := params.Tokenomics.InflationRate
	targetRate := params.Tokenomics.TargetInflationRate

	// Check if rate exceeds bounds
	if currentRate > params.Tokenomics.MaxInflationRate {
		alert := k.createInflationAlert(
			types.InflationAlertTypeAboveMax,
			currentRate,
			targetRate,
			types.AlertSeverityCritical,
			fmt.Sprintf("Inflation rate %d exceeds maximum %d", currentRate, params.Tokenomics.MaxInflationRate),
		)
		k.inflationAlerts = append(k.inflationAlerts, alert)
		return types.ErrInflationRateTooHigh
	}

	if currentRate < params.Tokenomics.MinInflationRate {
		alert := k.createInflationAlert(
			types.InflationAlertTypeBelowMin,
			currentRate,
			targetRate,
			types.AlertSeverityCritical,
			fmt.Sprintf("Inflation rate %d below minimum %d", currentRate, params.Tokenomics.MinInflationRate),
		)
		k.inflationAlerts = append(k.inflationAlerts, alert)
		return types.ErrInflationRateTooLow
	}

	// Check deviation from target
	var deviation uint64
	if currentRate > targetRate {
		deviation = currentRate - targetRate
	} else {
		deviation = targetRate - currentRate
	}

	if deviation > params.InflationAlertThreshold {
		severity := types.AlertSeverityWarning
		if deviation > params.InflationAlertThreshold*2 {
			severity = types.AlertSeverityCritical
		}

		alertType := types.InflationAlertTypeAboveTarget
		if currentRate < targetRate {
			alertType = types.InflationAlertTypeBelowTarget
		}

		alert := k.createInflationAlert(
			alertType,
			currentRate,
			targetRate,
			severity,
			fmt.Sprintf("Inflation rate deviates from target by %d basis points", deviation),
		)
		k.inflationAlerts = append(k.inflationAlerts, alert)
	}

	// Check for rapid changes
	if k.previousInflation > 0 {
		var change uint64
		if currentRate > k.previousInflation {
			change = currentRate - k.previousInflation
		} else {
			change = k.previousInflation - currentRate
		}

		// Alert if change > 1% in one period
		if change > 100 {
			alert := k.createInflationAlert(
				types.InflationAlertTypeRapidChange,
				currentRate,
				targetRate,
				types.AlertSeverityWarning,
				fmt.Sprintf("Rapid inflation change detected: %d basis points", change),
			)
			k.inflationAlerts = append(k.inflationAlerts, alert)
		}
	}

	k.previousInflation = currentRate
	params.Tokenomics.LastInflationCheck = timestamppb.New(time.Unix(k.currentTime, 0))
	return k.SetParams(params)
}

func (k *Keeper) createInflationAlert(
	alertType types.InflationAlertType,
	currentRate, targetRate uint64,
	severity types.AlertSeverity,
	message string,
) *types.InflationAlert {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d-%d-%s", k.currentTime, currentRate, message)))
	alertID := hex.EncodeToString(h.Sum(nil))[:16]

	return &types.InflationAlert{
		AlertId:              alertID,
		AlertType:            alertType,
		CurrentInflationRate: currentRate,
		TargetInflationRate:  targetRate,
		TriggeredAt:          timestamppb.New(time.Unix(k.currentTime, 0)),
		Message:              message,
		Severity:             severity,
		ActionTaken:          false,
		ActionDescription:    "",
	}
}

// AdjustInflationRate adjusts the inflation rate
func (k *Keeper) AdjustInflationRate(newRate uint64, reason string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	params := k.GetParams()
	oldRate := params.Tokenomics.InflationRate

	if newRate > params.Tokenomics.MaxInflationRate {
		return types.ErrInflationRateTooHigh
	}
	if newRate < params.Tokenomics.MinInflationRate {
		return types.ErrInflationRateTooLow
	}

	params.Tokenomics.InflationRate = newRate
	params.Tokenomics.LastInflationAdjustment = timestamppb.New(time.Unix(k.currentTime, 0))

	// Create info alert for manual adjustment
	alert := k.createInflationAlert(
		types.InflationAlertTypeUnspecified,
		newRate,
		params.Tokenomics.TargetInflationRate,
		types.AlertSeverityInfo,
		fmt.Sprintf("Inflation adjusted from %d to %d: %s", oldRate, newRate, reason),
	)
	alert.ActionTaken = true
	alert.ActionDescription = reason
	k.inflationAlerts = append(k.inflationAlerts, alert)

	return k.SetParams(params)
}

// GetInflationAlerts returns recent inflation alerts
func (k *Keeper) GetInflationAlerts(limit uint64) []*types.InflationAlert {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if limit == 0 || limit > uint64(len(k.inflationAlerts)) {
		limit = uint64(len(k.inflationAlerts))
	}

	// Return most recent alerts
	start := uint64(len(k.inflationAlerts)) - limit
	return k.inflationAlerts[start:]
}

// To be continued in next part...
