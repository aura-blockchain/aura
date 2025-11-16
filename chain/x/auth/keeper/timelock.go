package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/auth/types"
	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
)

// ProposeTimeLockedAction proposes a new time-locked admin action
func (k *Keeper) ProposeTimeLockedAction(ctx context.Context, proposer, actionType string, payload []byte, delaySeconds uint64) (*authproto.TimeLockedAction, error) {
	// Validate proposer has permission
	if err := k.RequirePermission(ctx, proposer, types.PermissionManageTimeLock); err != nil {
		return nil, err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	now := time.Now()
	actionID := types.GenerateID("timelock", proposer, actionType, now.String())

	// Use provided delay or default
	if delaySeconds == 0 {
		delaySeconds = k.params.DefaultTimelockDelaySeconds
	}

	executableAt := now.Add(time.Duration(delaySeconds) * time.Second)

	action := &authproto.TimeLockedAction{
		Id:           actionID,
		ActionType:   actionType,
		Payload:      payload,
		Proposer:     proposer,
		ProposedAt:   &now,
		ExecutableAt: &executableAt,
		Status:       authproto.ActionStatus_ACTION_STATUS_PENDING,
		DelaySeconds: delaySeconds,
	}

	// Validate action
	if err := types.ValidateTimeLockedAction(action); err != nil {
		k.LogAudit(ctx, proposer, "propose_timelock_action", actionID, "failed", nil, err.Error())
		return nil, fmt.Errorf("%w: %v", types.ErrInvalidAction, err)
	}

	k.timeLockedActions[actionID] = action
	k.LogAudit(ctx, proposer, "propose_timelock_action", actionID, "success", map[string]string{
		"action_type":   actionType,
		"delay_seconds": fmt.Sprintf("%d", delaySeconds),
		"executable_at": executableAt.Format(time.RFC3339),
	}, "")

	return action, nil
}

// ExecuteTimeLockedAction executes a ready time-locked action
func (k *Keeper) ExecuteTimeLockedAction(ctx context.Context, executor, actionID string) error {
	// Validate executor has permission
	if err := k.RequirePermission(ctx, executor, types.PermissionManageTimeLock); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Get action
	action, ok := k.timeLockedActions[actionID]
	if !ok {
		k.LogAudit(ctx, executor, "execute_timelock_action", actionID, "failed", nil, "action not found")
		return types.ErrActionNotFound
	}

	// Check if already executed
	if action.Status == authproto.ActionStatus_ACTION_STATUS_EXECUTED {
		k.LogAudit(ctx, executor, "execute_timelock_action", actionID, "failed", nil, "already executed")
		return types.ErrActionAlreadyExecuted
	}

	// Check if action is cancelled
	if action.Status == authproto.ActionStatus_ACTION_STATUS_CANCELLED {
		k.LogAudit(ctx, executor, "execute_timelock_action", actionID, "failed", nil, "action cancelled")
		return fmt.Errorf("action has been cancelled")
	}

	// Check if action is ready
	if !types.IsActionReady(action) {
		k.LogAudit(ctx, executor, "execute_timelock_action", actionID, "failed", map[string]string{
			"executable_at": action.ExecutableAt.Format(time.RFC3339),
			"now":           time.Now().Format(time.RFC3339),
		}, "action not ready")
		return types.ErrActionNotReady
	}

	// Mark as executed
	now := time.Now()
	action.Status = authproto.ActionStatus_ACTION_STATUS_EXECUTED
	action.ExecutedAt = &now

	k.LogAudit(ctx, executor, "execute_timelock_action", actionID, "success", map[string]string{
		"action_type": action.ActionType,
		"proposer":    action.Proposer,
	}, "")

	// Note: In a real implementation, the payload would be decoded and executed here
	// This would involve unmarshaling the payload and executing the appropriate admin action
	// For example: updating module parameters, changing admin addresses, etc.

	return nil
}

// CancelTimeLockedAction cancels a pending time-locked action
func (k *Keeper) CancelTimeLockedAction(ctx context.Context, canceller, actionID string) error {
	// Validate canceller has permission (must be admin)
	if err := k.RequirePermission(ctx, canceller, types.PermissionAdmin); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	// Get action
	action, ok := k.timeLockedActions[actionID]
	if !ok {
		k.LogAudit(ctx, canceller, "cancel_timelock_action", actionID, "failed", nil, "action not found")
		return types.ErrActionNotFound
	}

	// Check if already executed
	if action.Status == authproto.ActionStatus_ACTION_STATUS_EXECUTED {
		k.LogAudit(ctx, canceller, "cancel_timelock_action", actionID, "failed", nil, "already executed")
		return types.ErrActionAlreadyExecuted
	}

	// Check if already cancelled
	if action.Status == authproto.ActionStatus_ACTION_STATUS_CANCELLED {
		k.LogAudit(ctx, canceller, "cancel_timelock_action", actionID, "failed", nil, "already cancelled")
		return fmt.Errorf("action already cancelled")
	}

	// Mark as cancelled
	action.Status = authproto.ActionStatus_ACTION_STATUS_CANCELLED

	k.LogAudit(ctx, canceller, "cancel_timelock_action", actionID, "success", map[string]string{
		"action_type": action.ActionType,
		"proposer":    action.Proposer,
	}, "")

	return nil
}

// GetTimeLockedAction retrieves a time-locked action by ID

// ListTimeLockedActions returns all time-locked actions
func (k *Keeper) ListTimeLockedActions(status authproto.ActionStatus) []*authproto.TimeLockedAction {
	k.mu.RLock()
	defer k.mu.RUnlock()

	actions := make([]*authproto.TimeLockedAction, 0)
	for _, action := range k.timeLockedActions {
		if status != authproto.ActionStatus_ACTION_STATUS_UNSPECIFIED && action.Status != status {
			continue
		}
		actions = append(actions, action)
	}

	return actions
}

// UpdateReadyActions updates status of actions that are now ready
func (k *Keeper) UpdateReadyActions() int {
	k.mu.Lock()
	defer k.mu.Unlock()

	count := 0
	for _, action := range k.timeLockedActions {
		if action.Status == authproto.ActionStatus_ACTION_STATUS_PENDING && types.IsActionReady(action) {
			action.Status = authproto.ActionStatus_ACTION_STATUS_READY
			count++
		}
	}

	return count
}

// ExecuteParameterChange executes a parameter change action
func (k *Keeper) ExecuteParameterChange(ctx context.Context, executor string, newParams *authproto.Params) error {
	// This would typically be called from ExecuteTimeLockedAction after decoding the payload

	// Validate executor has admin permission
	if err := k.RequirePermission(ctx, executor, types.PermissionAdmin); err != nil {
		return err
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	oldParams := k.params
	k.params = newParams

	k.LogAudit(ctx, executor, "execute_parameter_change", "params", "success", map[string]string{
		"old_session_timeout": fmt.Sprintf("%d", oldParams.SessionTimeoutSeconds),
		"new_session_timeout": fmt.Sprintf("%d", newParams.SessionTimeoutSeconds),
	}, "")

	return nil
}

// ScheduleParameterChange creates a time-locked action for parameter changes
func (k *Keeper) ScheduleParameterChange(ctx context.Context, proposer string, newParams *authproto.Params, delaySeconds uint64) (*authproto.TimeLockedAction, error) {
	// Serialize parameters (in real implementation, use proper serialization)
	payload := []byte(fmt.Sprintf("%+v", newParams))

	return k.ProposeTimeLockedAction(ctx, proposer, "UPDATE_PARAMS", payload, delaySeconds)
}
