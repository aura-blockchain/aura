// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

// TriggerWebhook sends an event to a webhook (non-blocking)
func (k Keeper) TriggerWebhook(ctx context.Context, webhookID string, event *WebhookEvent) error {
	config, err := k.GetWebhook(ctx, webhookID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	if !config.Enabled {
		return fmt.Errorf("webhook is disabled")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Send webhook asynchronously (non-consensus operation)
	go func() {
		if err := k.sendWebhook(config, event); err != nil {
			sdkCtx.Logger().Error("webhook send failed", "webhook_id", webhookID, "error", err)
		}
	}()

	return nil
}

// sendWebhook sends a webhook HTTP request with retries
func (k Keeper) sendWebhook(config *WebhookConfig, event *WebhookEvent) error {
	// Prepare payload
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", config.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", config.Secret)
	req.Header.Set("X-Webhook-ID", config.ID)
	req.Header.Set("X-Event-Type", event.Type)

	// Apply custom headers
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	// Send request with retries
	var lastErr error
	maxRetries := config.RetryCount
	if maxRetries <= 0 {
		maxRetries = 3
	}

	for i := 0; i < maxRetries; i++ {
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		body, _ := io.ReadAll(resp.Body)
		lastErr = fmt.Errorf("webhook failed with status %d: %s", resp.StatusCode, string(body))
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return lastErr
}

// NotifyWebhooks sends an event to all registered webhooks for a specific event type
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
			go func(cfg WebhookConfig) {
				if err := k.sendWebhook(&cfg, event); err != nil {
					sdkCtx.Logger().Error("webhook broadcast failed", "webhook_id", cfg.ID, "error", err)
				}
			}(config)
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
