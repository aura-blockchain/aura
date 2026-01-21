// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PendingWebhookKeyPrefix is the key prefix for pending webhook events in the KV store
var PendingWebhookKeyPrefix = []byte{0x0E}

// WebhookConfig represents a webhook configuration stored in the KV store
type WebhookConfig struct {
	ID          string            `json:"id"`
	URL         string            `json:"url"`
	Secret      string            `json:"secret"`
	Events      []string          `json:"events"`
	Enabled     bool              `json:"enabled"`
	RetryCount  int               `json:"retry_count"`
	Timeout     int64             `json:"timeout"` // duration in seconds
	Headers     map[string]string `json:"headers,omitempty"`
	LastTrigger time.Time         `json:"last_trigger"`
	FailCount   int               `json:"fail_count"`
}

// WebhookEvent represents an event to be sent via webhook
type WebhookEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
}

var WebhookKeyPrefix = []byte{0x0D}

// RegisterWebhook registers a new webhook configuration
func (k Keeper) RegisterWebhook(ctx context.Context, url, secret string, events []string) (string, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	webhookID := k.generateID(ctx, "webhook")

	config := &WebhookConfig{
		ID:         webhookID,
		URL:        url,
		Secret:     secret,
		Events:     events,
		Enabled:    true,
		RetryCount: 3,
		Timeout:    30,
		Headers:    make(map[string]string),
	}

	// Store webhook configuration in KV store
	store := k.storeService.OpenKVStore(ctx)
	key := append(WebhookKeyPrefix, []byte(webhookID)...)

	bz, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	if err := store.Set(key, bz); err != nil {
		return "", err
	}

	// Emit event
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"webhook_registered",
		sdk.NewAttribute("webhook_id", webhookID),
		sdk.NewAttribute("url", url),
	))

	return webhookID, nil
}

// GetWebhook retrieves a webhook configuration from the KV store
func (k Keeper) GetWebhook(ctx context.Context, webhookID string) (*WebhookConfig, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(WebhookKeyPrefix, []byte(webhookID)...)

	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, fmt.Errorf("webhook not found: %s", webhookID)
	}

	var config WebhookConfig
	if err := json.Unmarshal(bz, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// UpdateWebhook updates a webhook configuration
func (k Keeper) UpdateWebhook(ctx context.Context, webhookID string, url string, events []string, enabled bool) error {
	config, err := k.GetWebhook(ctx, webhookID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	// Update fields
	if url != "" {
		config.URL = url
	}
	if events != nil {
		config.Events = events
	}
	config.Enabled = enabled

	// Save updated config
	store := k.storeService.OpenKVStore(ctx)
	key := append(WebhookKeyPrefix, []byte(webhookID)...)

	bz, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	return store.Set(key, bz)
}

// RemoveWebhook removes a webhook configuration
func (k Keeper) RemoveWebhook(ctx context.Context, webhookID string) error {
	store := k.storeService.OpenKVStore(ctx)
	key := append(WebhookKeyPrefix, []byte(webhookID)...)

	// Check if exists
	has, err := store.Has(key)
	if err != nil {
		return fmt.Errorf("failed to Has: %w", err)
	}
	if !has {
		return fmt.Errorf("webhook not found: %s", webhookID)
	}

	return store.Delete(key)
}

// PendingWebhookEvent represents a webhook event queued for off-chain delivery.
// These events are stored in the KV store and emitted via SDK events for
// off-chain processors to pick up and deliver.
type PendingWebhookEvent struct {
	ID          string            `json:"id"`
	WebhookID   string            `json:"webhook_id"`
	WebhookURL  string            `json:"webhook_url"`
	Secret      string            `json:"secret"`
	Headers     map[string]string `json:"headers,omitempty"`
	Event       *WebhookEvent     `json:"event"`
	RetryCount  int               `json:"retry_count"`
	Timeout     int64             `json:"timeout"`
	CreatedAt   time.Time         `json:"created_at"`
	BlockHeight int64             `json:"block_height"`
}

// TriggerWebhook queues a webhook event for off-chain delivery.
//
// CONSENSUS SAFETY: This function does NOT make HTTP calls. External HTTP calls
// are inherently non-deterministic because:
//   - Network conditions vary between validators
//   - External service responses differ across nodes
//   - Timing and latency vary unpredictably
//
// Instead, webhook events are:
//  1. Stored in the KV store as pending events (deterministic)
//  2. Emitted via SDK events for off-chain processors to consume
//  3. Delivered by off-chain relay services outside consensus
//
// This ensures all validators reach the same state while still enabling
// webhook notifications through off-chain infrastructure.
func (k Keeper) TriggerWebhook(ctx context.Context, webhookID string, event *WebhookEvent) error {
	config, err := k.GetWebhook(ctx, webhookID)
	if err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	if !config.Enabled {
		return fmt.Errorf("webhook is disabled")
	}

	return k.queueWebhookEvent(ctx, config, event)
}

// queueWebhookEvent stores a webhook event for off-chain delivery and emits an SDK event.
// This is consensus-safe because it only performs deterministic operations:
// - Stores data in the KV store (deterministic)
// - Emits SDK events (deterministic)
// The actual HTTP delivery happens off-chain via relay services.
func (k Keeper) queueWebhookEvent(ctx context.Context, config *WebhookConfig, event *WebhookEvent) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Generate unique ID for the pending event
	pendingID := k.generateID(ctx, "pending_webhook")

	pending := &PendingWebhookEvent{
		ID:          pendingID,
		WebhookID:   config.ID,
		WebhookURL:  config.URL,
		Secret:      config.Secret,
		Headers:     config.Headers,
		Event:       event,
		RetryCount:  config.RetryCount,
		Timeout:     config.Timeout,
		CreatedAt:   sdkCtx.BlockTime(),
		BlockHeight: sdkCtx.BlockHeight(),
	}

	// Store in KV store for persistence and queryability
	store := k.storeService.OpenKVStore(ctx)
	key := append(PendingWebhookKeyPrefix, []byte(pendingID)...)

	bz, err := json.Marshal(pending)
	if err != nil {
		return fmt.Errorf("failed to marshal pending webhook: %w", err)
	}

	if err := store.Set(key, bz); err != nil {
		return fmt.Errorf("failed to store pending webhook: %w", err)
	}

	// Emit SDK event for off-chain relay services to consume.
	// Off-chain indexers/relayers can listen for these events and deliver webhooks.
	eventPayload, _ := json.Marshal(event)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"webhook_queued",
		sdk.NewAttribute("pending_id", pendingID),
		sdk.NewAttribute("webhook_id", config.ID),
		sdk.NewAttribute("webhook_url", config.URL),
		sdk.NewAttribute("event_type", event.Type),
		sdk.NewAttribute("event_id", event.ID),
		sdk.NewAttribute("payload", string(eventPayload)),
		sdk.NewAttribute("block_height", fmt.Sprintf("%d", sdkCtx.BlockHeight())),
	))

	return nil
}

// GetPendingWebhook retrieves a pending webhook event from the KV store
func (k Keeper) GetPendingWebhook(ctx context.Context, pendingID string) (*PendingWebhookEvent, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := append(PendingWebhookKeyPrefix, []byte(pendingID)...)

	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, fmt.Errorf("pending webhook not found: %s", pendingID)
	}

	var pending PendingWebhookEvent
	if err := json.Unmarshal(bz, &pending); err != nil {
		return nil, err
	}

	return &pending, nil
}

// MarkWebhookDelivered removes a pending webhook after successful off-chain delivery.
// This should be called by governance or an authorized relayer to clean up delivered events.
func (k Keeper) MarkWebhookDelivered(ctx context.Context, pendingID string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.storeService.OpenKVStore(ctx)
	key := append(PendingWebhookKeyPrefix, []byte(pendingID)...)

	// Verify it exists
	has, err := store.Has(key)
	if err != nil {
		return fmt.Errorf("failed to check pending webhook: %w", err)
	}
	if !has {
		return fmt.Errorf("pending webhook not found: %s", pendingID)
	}

	if err := store.Delete(key); err != nil {
		return fmt.Errorf("failed to delete pending webhook: %w", err)
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"webhook_delivered",
		sdk.NewAttribute("pending_id", pendingID),
	))

	return nil
}

// GetAllPendingWebhooks retrieves all pending webhook events for off-chain processing
func (k Keeper) GetAllPendingWebhooks(ctx context.Context) ([]*PendingWebhookEvent, error) {
	var pending []*PendingWebhookEvent

	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(PendingWebhookKeyPrefix, storetypes.PrefixEndBytes(PendingWebhookKeyPrefix))
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var event PendingWebhookEvent
		if err := json.Unmarshal(iterator.Value(), &event); err != nil {
			continue
		}
		pending = append(pending, &event)
	}

	return pending, nil
}

// NotifyWebhooks queues events to all registered webhooks for a specific event type.
//
// CONSENSUS SAFETY: This function does NOT make HTTP calls. External HTTP calls
// are inherently non-deterministic because network conditions, external service
// responses, and timing vary between validators.
//
// Instead, webhook events are queued deterministically:
//  1. Each matching webhook gets a pending event stored in the KV store
//  2. SDK events are emitted for off-chain relay services to consume
//  3. Actual HTTP delivery happens off-chain, outside consensus
//
// This ensures all validators reach the same state while enabling webhook
// notifications through off-chain infrastructure.
func (k Keeper) NotifyWebhooks(ctx context.Context, eventType string, data map[string]interface{}) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	event := &WebhookEvent{
		ID:        k.generateID(ctx, "event"),
		Type:      eventType,
		Timestamp: sdkCtx.BlockTime(),
		Data:      data,
		Metadata: map[string]interface{}{
			"block_height": sdkCtx.BlockHeight(),
		},
	}

	// Iterate through all webhooks
	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(WebhookKeyPrefix, storetypes.PrefixEndBytes(WebhookKeyPrefix))
	if err != nil {
		return fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var config WebhookConfig
		if err := json.Unmarshal(iterator.Value(), &config); err != nil {
			continue
		}

		// Check if webhook is interested in this event
		if config.Enabled && k.webhookInterestedInEvent(&config, eventType) {
			// Queue for off-chain delivery - NO HTTP calls during consensus
			if err := k.queueWebhookEvent(ctx, &config, event); err != nil {
				sdkCtx.Logger().Error("failed to queue webhook event", "webhook_id", config.ID, "error", err)
				// Continue processing other webhooks - individual queue failures
				// should not block other notifications
			}
		}
	}

	return nil
}

// webhookInterestedInEvent checks if a webhook is interested in an event type
func (k Keeper) webhookInterestedInEvent(config *WebhookConfig, eventType string) bool {
	for _, event := range config.Events {
		if event == eventType || event == "*" {
			return true
		}
	}
	return false
}

// GetAllWebhooks retrieves all webhook configurations
func (k Keeper) GetAllWebhooks(ctx context.Context) ([]*WebhookConfig, error) {
	var webhooks []*WebhookConfig

	store := k.storeService.OpenKVStore(ctx)
	iterator, err := store.Iterator(WebhookKeyPrefix, storetypes.PrefixEndBytes(WebhookKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var config WebhookConfig
		if err := json.Unmarshal(iterator.Value(), &config); err != nil {
			continue
		}
		webhooks = append(webhooks, &config)
	}

	return webhooks, nil
}
