package keeper

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	k.mu.Lock()
	defer k.mu.Unlock()

	// Generate schedule ID
	scheduleID := fmt.Sprintf("schedule_%s_%d", keyID, time.Now().Unix())

	now := time.Now()
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

	// Store in state
	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(schedule)
	store.Set(types.GetKeyRotationScheduleKey(scheduleID), bz)

	// Cache
	k.rotationSchedules[scheduleID] = schedule

	k.Logger(ctx).Info("created key rotation schedule",
		"schedule_id", scheduleID,
		"key_id", keyID,
		"interval", rotationIntervalSeconds,
	)

	return scheduleID, nil
}

// GetKeyRotationSchedule retrieves a key rotation schedule by ID
func (k Keeper) GetKeyRotationSchedule(ctx context.Context, scheduleID string) (*cryptoproto.KeyRotationSchedule, error) {
	k.mu.RLock()
	if schedule, ok := k.rotationSchedules[scheduleID]; ok {
		k.mu.RUnlock()
		return schedule, nil
	}
	k.mu.RUnlock()

	store := k.getStore(ctx)
	bz := store.Get(types.GetKeyRotationScheduleKey(scheduleID))
	if bz == nil {
		return nil, types.ErrKeyRotationScheduleNotFound
	}

	var schedule cryptoproto.KeyRotationSchedule
	k.cdc.MustUnmarshal(bz, &schedule)

	k.mu.Lock()
	k.rotationSchedules[scheduleID] = &schedule
	k.mu.Unlock()

	return &schedule, nil
}

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

	k.mu.Lock()
	defer k.mu.Unlock()

	rotationID := fmt.Sprintf("rotation_%s_%d", keyID, time.Now().Unix())
	rotationTime := time.Now()

	// Update any associated rotation schedules
	for scheduleID, schedule := range k.rotationSchedules {
		if schedule.KeyId == keyID {
			schedule.LastRotation = timestamppb.New(rotationTime)
			nextRotation := rotationTime.Add(time.Duration(schedule.RotationIntervalSeconds) * time.Second)
			schedule.NextRotationTime = timestamppb.New(nextRotation)

			// Update in store
			store := k.getStore(ctx)
			bz := k.cdc.MustMarshal(schedule)
			store.Set(types.GetKeyRotationScheduleKey(scheduleID), bz)
		}
	}

	k.Logger(ctx).Info("rotated key",
		"rotation_id", rotationID,
		"key_id", keyID,
		"creator", creator,
	)

	return rotationID, rotationTime, nil
}

// ProcessScheduledRotations processes all scheduled key rotations that are due
func (k Keeper) ProcessScheduledRotations(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	now := time.Now()

	for scheduleID, schedule := range k.rotationSchedules {
		if !schedule.Enabled {
			continue
		}

		if !schedule.Policy.AutoRotate {
			continue
		}

		if schedule.NextRotationTime.AsTime().After(now) {
			continue
		}

		// Perform rotation
		k.Logger(ctx).Info("processing scheduled key rotation",
			"schedule_id", scheduleID,
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
		store := k.getStore(ctx)
		bz := k.cdc.MustMarshal(schedule)
		store.Set(types.GetKeyRotationScheduleKey(scheduleID), bz)
	}

	return nil
}

// DisableKeyRotationSchedule disables a key rotation schedule
func (k Keeper) DisableKeyRotationSchedule(ctx context.Context, scheduleID string) error {
	schedule, err := k.GetKeyRotationSchedule(ctx, scheduleID)
	if err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	schedule.Enabled = false

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(schedule)
	store.Set(types.GetKeyRotationScheduleKey(scheduleID), bz)

	k.rotationSchedules[scheduleID] = schedule

	k.Logger(ctx).Info("disabled key rotation schedule", "schedule_id", scheduleID)

	return nil
}

// EnableKeyRotationSchedule enables a key rotation schedule
func (k Keeper) EnableKeyRotationSchedule(ctx context.Context, scheduleID string) error {
	schedule, err := k.GetKeyRotationSchedule(ctx, scheduleID)
	if err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	schedule.Enabled = true

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(schedule)
	store.Set(types.GetKeyRotationScheduleKey(scheduleID), bz)

	k.rotationSchedules[scheduleID] = schedule

	k.Logger(ctx).Info("enabled key rotation schedule", "schedule_id", scheduleID)

	return nil
}

// GetSchedulesForKey returns all rotation schedules for a given key
func (k Keeper) GetSchedulesForKey(ctx context.Context, keyID string) []*cryptoproto.KeyRotationSchedule {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var schedules []*cryptoproto.KeyRotationSchedule
	for _, schedule := range k.rotationSchedules {
		if schedule.KeyId == keyID {
			schedules = append(schedules, schedule)
		}
	}

	return schedules
}

// SetKeyRotationSchedule stores a key rotation schedule (for genesis)
func (k *Keeper) SetKeyRotationSchedule(ctx context.Context, schedule *cryptoproto.KeyRotationSchedule) error {
	if schedule == nil {
		return nil
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	store := k.getStore(ctx)
	bz := k.cdc.MustMarshal(schedule)
	store.Set(types.GetKeyRotationScheduleKey(schedule.Id), bz)

	k.rotationSchedules[schedule.Id] = schedule
	return nil
}

// GetAllKeyRotationSchedules retrieves all key rotation schedules
func (k Keeper) GetAllKeyRotationSchedules(ctx context.Context) []*cryptoproto.KeyRotationSchedule {
	k.mu.RLock()
	defer k.mu.RUnlock()

	schedules := make([]*cryptoproto.KeyRotationSchedule, 0, len(k.rotationSchedules))
	for _, schedule := range k.rotationSchedules {
		schedules = append(schedules, schedule)
	}
	return schedules
}
