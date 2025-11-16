package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ============================
// VESTING SCHEDULES (Feature 4)
// ============================

// CreateVestingSchedule creates a new vesting schedule
func (k *Keeper) CreateVestingSchedule(
	beneficiary string,
	totalAmount string,
	startTime *timestamppb.Timestamp,
	cliffDuration, vestingDuration uint64,
	vestingType types.VestingType,
	scheduleType types.ScheduleType,
) (string, error) {
	if beneficiary == "" {
		return "", types.ErrInvalidBeneficiary
	}

	amount := new(big.Int)
	if _, ok := amount.SetString(totalAmount, 10); !ok || amount.Cmp(big.NewInt(0)) <= 0 {
		return "", types.ErrInvalidAmount
	}

	if vestingDuration == 0 {
		return "", types.ErrInvalidDuration
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Generate schedule ID
	scheduleID := k.generateScheduleID(beneficiary, totalAmount, startTime)

	schedule := &types.VestingSchedule{
		ScheduleId:         scheduleID,
		BeneficiaryAddress: beneficiary,
		TotalAmount:        totalAmount,
		VestedAmount:       "0",
		StartTime:          startTime,
		CliffDuration:      cliffDuration,
		VestingDuration:    vestingDuration,
		VestingType:        vestingType,
		ScheduleType:       scheduleType,
		Revoked:            false,
		RevokedAt:          nil,
		RevokedReason:      "",
	}

	k.vestingSchedules[scheduleID] = schedule
	k.userVestings[beneficiary] = append(k.userVestings[beneficiary], scheduleID)

	return scheduleID, nil
}

// ReleaseVestedTokens releases vested tokens to beneficiary
func (k *Keeper) ReleaseVestedTokens(beneficiary, scheduleID string) (string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	schedule, ok := k.vestingSchedules[scheduleID]
	if !ok {
		return "0", types.ErrVestingScheduleNotFound
	}

	if schedule.BeneficiaryAddress != beneficiary {
		return "0", types.ErrUnauthorized
	}

	if schedule.Revoked {
		return "0", types.ErrVestingAlreadyRevoked
	}

	// Calculate vested amount
	vestedAmount, err := k.calculateVestedAmount(schedule)
	if err != nil {
		return "0", err
	}

	vested := new(big.Int)
	vested.SetString(vestedAmount, 10)

	alreadyVested := new(big.Int)
	alreadyVested.SetString(schedule.VestedAmount, 10)

	// Calculate releasable amount
	releasable := new(big.Int).Sub(vested, alreadyVested)
	if releasable.Cmp(big.NewInt(0)) <= 0 {
		return "0", types.ErrNoVestedTokens
	}

	// Update schedule
	schedule.VestedAmount = vested.String()
	k.vestingSchedules[scheduleID] = schedule

	return releasable.String(), nil
}

// RevokeVestingSchedule revokes a vesting schedule
func (k *Keeper) RevokeVestingSchedule(scheduleID, reason string) (string, error) {
	k.mu.Lock()
	defer k.mu.Unlock()

	schedule, ok := k.vestingSchedules[scheduleID]
	if !ok {
		return "0", types.ErrVestingScheduleNotFound
	}

	if schedule.Revoked {
		return "0", types.ErrVestingAlreadyRevoked
	}

	// Calculate vested amount at revocation time
	vestedAmount, err := k.calculateVestedAmount(schedule)
	if err != nil {
		return "0", err
	}

	vested := new(big.Int)
	vested.SetString(vestedAmount, 10)

	totalAmount := new(big.Int)
	totalAmount.SetString(schedule.TotalAmount, 10)

	// Calculate unvested amount to return
	unvested := new(big.Int).Sub(totalAmount, vested)

	// Update schedule
	schedule.Revoked = true
	schedule.RevokedAt = timestamppb.New(time.Unix(k.currentTime, 0))
	schedule.RevokedReason = reason
	schedule.VestedAmount = vested.String()
	k.vestingSchedules[scheduleID] = schedule

	return unvested.String(), nil
}

// GetVestingSchedule retrieves a vesting schedule
func (k *Keeper) GetVestingSchedule(scheduleID string) (*types.VestingSchedule, bool) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	schedule, ok := k.vestingSchedules[scheduleID]
	return schedule, ok
}

// GetUserVestingSchedules returns all vesting schedules for a user
func (k *Keeper) GetUserVestingSchedules(beneficiary string) []*types.VestingSchedule {
	k.mu.RLock()
	defer k.mu.RUnlock()

	scheduleIDs := k.userVestings[beneficiary]
	schedules := make([]*types.VestingSchedule, 0, len(scheduleIDs))

	for _, id := range scheduleIDs {
		if schedule, ok := k.vestingSchedules[id]; ok {
			schedules = append(schedules, schedule)
		}
	}

	return schedules
}

// calculateVestedAmount calculates how much has vested for a schedule
func (k *Keeper) calculateVestedAmount(schedule *types.VestingSchedule) (string, error) {
	if schedule.Revoked {
		return schedule.VestedAmount, nil
	}

	startTime := schedule.StartTime.AsTime().Unix()
	cliffEnd := startTime + int64(schedule.CliffDuration)
	vestingEnd := startTime + int64(schedule.VestingDuration)

	// Before cliff, nothing vested
	if k.currentTime < cliffEnd {
		return "0", types.ErrCliffNotReached
	}

	totalAmount := new(big.Int)
	totalAmount.SetString(schedule.TotalAmount, 10)

	// After vesting period, everything vested
	if k.currentTime >= vestingEnd {
		return totalAmount.String(), nil
	}

	// During vesting period
	switch schedule.VestingType {
	case types.VestingTypeLinear:
		elapsed := k.currentTime - startTime
		total := int64(schedule.VestingDuration)

		vested := new(big.Int).Mul(totalAmount, big.NewInt(elapsed))
		vested.Div(vested, big.NewInt(total))
		return vested.String(), nil

	case types.VestingTypeCliffThenLinear:
		elapsed := k.currentTime - cliffEnd
		total := int64(schedule.VestingDuration - schedule.CliffDuration)

		vested := new(big.Int).Mul(totalAmount, big.NewInt(elapsed))
		vested.Div(vested, big.NewInt(total))
		return vested.String(), nil

	case types.VestingTypeMilestone:
		// For milestone vesting, would need additional milestone data
		// For now, treat as linear
		elapsed := k.currentTime - startTime
		total := int64(schedule.VestingDuration)

		vested := new(big.Int).Mul(totalAmount, big.NewInt(elapsed))
		vested.Div(vested, big.NewInt(total))
		return vested.String(), nil

	default:
		return "0", types.ErrInvalidAmount
	}
}

// GetTotalVesting returns total vesting amounts for all users
func (k *Keeper) GetTotalVesting() (string, string) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	totalVested := big.NewInt(0)
	totalVesting := big.NewInt(0)

	for _, schedule := range k.vestingSchedules {
		if schedule.Revoked {
			continue
		}

		vested := new(big.Int)
		vested.SetString(schedule.VestedAmount, 10)
		totalVested.Add(totalVested, vested)

		total := new(big.Int)
		total.SetString(schedule.TotalAmount, 10)
		totalVesting.Add(totalVesting, total)
	}

	return totalVested.String(), totalVesting.String()
}

func (k *Keeper) generateScheduleID(beneficiary, amount string, startTime *timestamppb.Timestamp) string {
	h := sha256.New()
	h.Write([]byte(beneficiary))
	h.Write([]byte(amount))
	h.Write([]byte(fmt.Sprintf("%d", startTime.Seconds)))
	h.Write([]byte(fmt.Sprintf("%d", k.currentTime)))
	return "vs:" + hex.EncodeToString(h.Sum(nil))[:32]
}
