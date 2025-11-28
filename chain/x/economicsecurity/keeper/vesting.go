package keeper

import (
	"context"
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
	ctx context.Context,
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

	// Get current time for ID generation
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return "", err
	}

	// Generate schedule ID
	scheduleID := generateScheduleID(beneficiary, totalAmount, startTime, currentTime)

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

	// Store the schedule
	if err := k.SetVestingSchedule(ctx, schedule); err != nil {
		return "", err
	}

	// Add to user's vesting index
	if err := k.AddUserVestingSchedule(ctx, beneficiary, scheduleID); err != nil {
		return "", err
	}

	return scheduleID, nil
}

// ReleaseVestedTokens releases vested tokens to beneficiary
func (k *Keeper) ReleaseVestedTokens(ctx context.Context, beneficiary, scheduleID string) (string, error) {
	schedule, err := k.GetVestingSchedule(ctx, scheduleID)
	if err != nil {
		return "0", err
	}

	if schedule.BeneficiaryAddress != beneficiary {
		return "0", types.ErrUnauthorized
	}

	if schedule.Revoked {
		return "0", types.ErrVestingAlreadyRevoked
	}

	// Calculate vested amount
	vestedAmount, err := k.calculateVestedAmount(ctx, schedule)
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
	if err := k.SetVestingSchedule(ctx, schedule); err != nil {
		return "0", err
	}

	return releasable.String(), nil
}

// RevokeVestingSchedule revokes a vesting schedule
func (k *Keeper) RevokeVestingSchedule(ctx context.Context, scheduleID, reason string) (string, error) {
	schedule, err := k.GetVestingSchedule(ctx, scheduleID)
	if err != nil {
		return "0", err
	}

	if schedule.Revoked {
		return "0", types.ErrVestingAlreadyRevoked
	}

	// Calculate vested amount at revocation time
	vestedAmount, err := k.calculateVestedAmount(ctx, schedule)
	if err != nil {
		return "0", err
	}

	vested := new(big.Int)
	vested.SetString(vestedAmount, 10)

	totalAmount := new(big.Int)
	totalAmount.SetString(schedule.TotalAmount, 10)

	// Calculate unvested amount to return
	unvested := new(big.Int).Sub(totalAmount, vested)

	// Get current time
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return "0", err
	}

	// Update schedule
	schedule.Revoked = true
	schedule.RevokedAt = timestamppb.New(time.Unix(currentTime, 0))
	schedule.RevokedReason = reason
	schedule.VestedAmount = vested.String()

	if err := k.SetVestingSchedule(ctx, schedule); err != nil {
		return "0", err
	}

	return unvested.String(), nil
}

// GetUserVestingSchedules returns all vesting schedules for a user
func (k *Keeper) GetUserVestingSchedules(ctx context.Context, beneficiary string) ([]*types.VestingSchedule, error) {
	scheduleIDs, err := k.GetUserVestingIndex(ctx, beneficiary)
	if err != nil {
		return nil, err
	}

	schedules := make([]*types.VestingSchedule, 0, len(scheduleIDs))
	for _, id := range scheduleIDs {
		schedule, err := k.GetVestingSchedule(ctx, id)
		if err == nil {
			schedules = append(schedules, schedule)
		}
	}

	return schedules, nil
}

// calculateVestedAmount calculates how much has vested for a schedule
func (k *Keeper) calculateVestedAmount(ctx context.Context, schedule *types.VestingSchedule) (string, error) {
	if schedule.Revoked {
		return schedule.VestedAmount, nil
	}

	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return "0", err
	}

	startTime := schedule.StartTime.AsTime().Unix()
	cliffEnd := startTime + int64(schedule.CliffDuration)
	vestingEnd := startTime + int64(schedule.VestingDuration)

	// Before cliff, nothing vested
	if currentTime < cliffEnd {
		return "0", types.ErrCliffNotReached
	}

	totalAmount := new(big.Int)
	totalAmount.SetString(schedule.TotalAmount, 10)

	// After vesting period, everything vested
	if currentTime >= vestingEnd {
		return totalAmount.String(), nil
	}

	// During vesting period
	switch schedule.VestingType {
	case types.VestingTypeLinear:
		elapsed := currentTime - startTime
		total := int64(schedule.VestingDuration)

		vested := new(big.Int).Mul(totalAmount, big.NewInt(elapsed))
		vested.Div(vested, big.NewInt(total))
		return vested.String(), nil

	case types.VestingTypeCliffThenLinear:
		elapsed := currentTime - cliffEnd
		total := int64(schedule.VestingDuration - schedule.CliffDuration)

		vested := new(big.Int).Mul(totalAmount, big.NewInt(elapsed))
		vested.Div(vested, big.NewInt(total))
		return vested.String(), nil

	case types.VestingTypeMilestone:
		// For milestone vesting, would need additional milestone data
		// For now, treat as linear
		elapsed := currentTime - startTime
		total := int64(schedule.VestingDuration)

		vested := new(big.Int).Mul(totalAmount, big.NewInt(elapsed))
		vested.Div(vested, big.NewInt(total))
		return vested.String(), nil

	default:
		return "0", types.ErrInvalidAmount
	}
}

// GetTotalVesting returns total vesting amounts for all users
func (k *Keeper) GetTotalVesting(ctx context.Context) (string, string, error) {
	totalVested := big.NewInt(0)
	totalVesting := big.NewInt(0)

	err := k.IterateVestingSchedules(ctx, func(schedule *types.VestingSchedule) bool {
		if schedule.Revoked {
			return true
		}

		vested := new(big.Int)
		vested.SetString(schedule.VestedAmount, 10)
		totalVested.Add(totalVested, vested)

		total := new(big.Int)
		total.SetString(schedule.TotalAmount, 10)
		totalVesting.Add(totalVesting, total)

		return true
	})

	if err != nil {
		return "0", "0", err
	}

	return totalVested.String(), totalVesting.String(), nil
}

// GetVestingScheduleInfo returns detailed info about a vesting schedule including current vested amount
func (k *Keeper) GetVestingScheduleInfo(ctx context.Context, scheduleID string) (*types.VestingSchedule, string, error) {
	schedule, err := k.GetVestingSchedule(ctx, scheduleID)
	if err != nil {
		return nil, "0", err
	}

	currentVested, err := k.calculateVestedAmount(ctx, schedule)
	if err != nil {
		// If error is cliff not reached, return 0 without error
		if err == types.ErrCliffNotReached {
			return schedule, "0", nil
		}
		return nil, "0", err
	}

	return schedule, currentVested, nil
}

func generateScheduleID(beneficiary, amount string, startTime *timestamppb.Timestamp, currentTime int64) string {
	h := sha256.New()
	h.Write([]byte(beneficiary))
	h.Write([]byte(amount))
	h.Write([]byte(fmt.Sprintf("%d", startTime.Seconds)))
	h.Write([]byte(fmt.Sprintf("%d", currentTime)))
	return "vs:" + hex.EncodeToString(h.Sum(nil))[:32]
}
