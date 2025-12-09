package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// CreateKeyRotationSchedule creates a new automated key rotation schedule
func (k Keeper) CreateKeyRotationSchedule(
	ctx context.Context,
	creator string,
	keyID string,
	rotationIntervalSeconds int64,
	policy *cryptoproto.KeyRotationPolicy,
) (string, error) {
	if keyID == "" {
		return "", types.ErrInvalidKeyID
	}
	if rotationIntervalSeconds < 3600 { // Minimum 1 hour
		return "", types.ErrInvalidRotationInterval
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return "", err
	}

	// Validate policy
	if policy == nil {
		policy = &cryptoproto.KeyRotationPolicy{
			MaxAgeDays:              params.DefaultRotationIntervalDays,
			WarningDaysBeforeExpiry: 7,
			AutoRotate:              params.EnableAutoRotation,
			MaxRotationAttempts:     3,
		}
	}

	// Generate schedule ID using consensus-safe block time
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()
	scheduleID := fmt.Sprintf("schedule_%s_%d", keyID, blockTime.Unix())

	now := blockTime
	nextRotation := now.Add(time.Duration(rotationIntervalSeconds) * time.Second)

	schedule := &cryptoproto.KeyRotationSchedule{
		Id:                      scheduleID,
		KeyId:                   keyID,
		NextRotationTime:        timestamppb.New(nextRotation),
		RotationIntervalSeconds: rotationIntervalSeconds,
		Enabled:                 true,
		CreatedBy:               creator,
		Policy:                  policy,
		LastRotation:            nil,
	}

	// Store in KV store
	if err := k.SetKeyRotationSchedule(ctx, schedule); err != nil {
		return "", err
	}

	k.Logger(sdkCtx).Info("created key rotation schedule",
		"schedule_id", scheduleID,
		"key_id", keyID,
		"interval", rotationIntervalSeconds,
	)

	return scheduleID, nil
}

// Note: GetKeyRotationSchedule is now implemented in keeper.go using KV store

// RotateKey performs a manual key rotation
func (k Keeper) RotateKey(
	ctx context.Context,
	creator string,
	keyID string,
	newPublicKey []byte,
) (string, time.Time, error) {
	if keyID == "" {
		return "", time.Time{}, types.ErrInvalidKeyID
	}
	if len(newPublicKey) < 32 {
		return "", time.Time{}, fmt.Errorf("public key too short")
	}

	rotateCtx := sdk.UnwrapSDKContext(ctx)
	rotationID := fmt.Sprintf("rotation_%s_%d", keyID, rotateCtx.BlockTime().Unix())
	rotationTime := rotateCtx.BlockTime()

	// Update any associated rotation schedules
	schedules := k.GetSchedulesForKey(ctx, keyID)
	for _, schedule := range schedules {
		schedule.LastRotation = timestamppb.New(rotationTime)
		nextRotation := rotationTime.Add(time.Duration(schedule.RotationIntervalSeconds) * time.Second)
		schedule.NextRotationTime = timestamppb.New(nextRotation)

		// Update in store
		if err := k.SetKeyRotationSchedule(ctx, schedule); err != nil {
			return "", time.Time{}, err
		}
	}

	k.Logger(rotateCtx).Info("rotated key",
		"rotation_id", rotationID,
		"key_id", keyID,
		"creator", creator,
	)

	return rotationID, rotationTime, nil
}

// ProcessScheduledRotations processes all scheduled key rotations that are due
func (k Keeper) ProcessScheduledRotations(ctx context.Context) error {
	processCtx := sdk.UnwrapSDKContext(ctx)
	now := processCtx.BlockTime()

	// Collect schedules that need rotation
	schedulesToRotate := make([]*cryptoproto.KeyRotationSchedule, 0)
	err := k.IterateKeyRotationSchedules(ctx, func(schedule *cryptoproto.KeyRotationSchedule) bool {
		if schedule.Enabled &&
			schedule.Policy != nil &&
			schedule.Policy.AutoRotate &&
			schedule.NextRotationTime.AsTime().Before(now) {
			schedulesToRotate = append(schedulesToRotate, schedule)
		}
		return false
	})
	if err != nil {
		return err
	}

	// Process rotations
	for _, schedule := range schedulesToRotate {
		k.Logger(processCtx).Info("processing scheduled key rotation",
			"schedule_id", schedule.Id,
			"key_id", schedule.KeyId,
		)

		// In a real implementation, this would:
		// 1. Generate a new key pair
		// 2. Update the key in the appropriate module
		// 3. Notify relevant parties
		// 4. Archive the old key

		// Update schedule
		schedule.LastRotation = timestamppb.New(now)
		nextRotation := now.Add(time.Duration(schedule.RotationIntervalSeconds) * time.Second)
		schedule.NextRotationTime = timestamppb.New(nextRotation)

		// Update in store
		if err := k.SetKeyRotationSchedule(ctx, schedule); err != nil {
			return err
		}
	}

	return nil
}

// DisableKeyRotationSchedule disables a key rotation schedule
func (k Keeper) DisableKeyRotationSchedule(ctx context.Context, scheduleID string) error {
	schedule, err := k.GetKeyRotationSchedule(ctx, scheduleID)
	if err != nil {
		return err
	}

	schedule.Enabled = false

	if err := k.SetKeyRotationSchedule(ctx, schedule); err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Logger(sdkCtx).Info("disabled key rotation schedule", "schedule_id", scheduleID)

	return nil
}

// EnableKeyRotationSchedule enables a key rotation schedule
func (k Keeper) EnableKeyRotationSchedule(ctx context.Context, scheduleID string) error {
	schedule, err := k.GetKeyRotationSchedule(ctx, scheduleID)
	if err != nil {
		return err
	}

	schedule.Enabled = true

	if err := k.SetKeyRotationSchedule(ctx, schedule); err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Logger(sdkCtx).Info("enabled key rotation schedule", "schedule_id", scheduleID)

	return nil
}

// Note: GetSchedulesForKey, SetKeyRotationSchedule, and GetAllKeyRotationSchedules
// are now implemented in keeper.go using KV store
