// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// NetworkHealthCollector collects real network health metrics from the node.
// It queries the local CometBFT RPC endpoint to gather peer count, block info,
// and other metrics that cannot be obtained from within the Cosmos SDK context.
type NetworkHealthCollector struct {
	rpcEndpoint string
	httpClient  *http.Client

	// Cache to avoid hammering the RPC endpoint
	mu              sync.RWMutex
	cachedHealth    *types.NetworkHealth
	cacheExpiry     time.Time
	cacheDuration   time.Duration

	// Block time calculation
	lastBlockHeight int64
	lastBlockTime   time.Time
	blockTimes      []float64 // Rolling window for average calculation
	maxBlockTimes   int
}

// NewNetworkHealthCollector creates a new network health collector.
// rpcEndpoint should be the local CometBFT RPC endpoint (e.g., "http://localhost:26657")
// The endpoint can be overridden via the MONITORING_RPC_ENDPOINT environment variable.
func NewNetworkHealthCollector(rpcEndpoint string) *NetworkHealthCollector {
	// Allow override via environment variable
	if envEndpoint := os.Getenv("MONITORING_RPC_ENDPOINT"); envEndpoint != "" {
		rpcEndpoint = envEndpoint
	}

	return &NetworkHealthCollector{
		rpcEndpoint: rpcEndpoint,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cacheDuration: 10 * time.Second, // Cache for 10 seconds
		blockTimes:    make([]float64, 0, 100),
		maxBlockTimes: 100, // Keep last 100 block times for averaging
	}
}

// SetRPCEndpoint updates the RPC endpoint
func (c *NetworkHealthCollector) SetRPCEndpoint(endpoint string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rpcEndpoint = endpoint
}

// rpcStatusResponse represents the CometBFT /status response
type rpcStatusResponse struct {
	Result struct {
		NodeInfo struct {
			Network string `json:"network"`
			Moniker string `json:"moniker"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHeight string    `json:"latest_block_height"`
			LatestBlockTime   time.Time `json:"latest_block_time"`
			CatchingUp        bool      `json:"catching_up"`
		} `json:"sync_info"`
	} `json:"result"`
}

// rpcNetInfoResponse represents the CometBFT /net_info response
type rpcNetInfoResponse struct {
	Result struct {
		Listening bool   `json:"listening"`
		NPeers    string `json:"n_peers"`
		Peers     []struct {
			NodeInfo struct {
				Moniker string `json:"moniker"`
			} `json:"node_info"`
			RemoteIP string `json:"remote_ip"`
		} `json:"peers"`
	} `json:"result"`
}

// rpcUnconfirmedTxsResponse represents the CometBFT /unconfirmed_txs response
type rpcUnconfirmedTxsResponse struct {
	Result struct {
		NTxs       string `json:"n_txs"`
		Total      string `json:"total"`
		TotalBytes string `json:"total_bytes"`
	} `json:"result"`
}

// rpcValidatorsResponse represents the CometBFT /validators response
type rpcValidatorsResponse struct {
	Result struct {
		BlockHeight string `json:"block_height"`
		Validators  []struct {
			Address     string `json:"address"`
			VotingPower string `json:"voting_power"`
		} `json:"validators"`
		Count string `json:"count"`
		Total string `json:"total"`
	} `json:"result"`
}

// CollectNetworkHealth gathers network health metrics from the local RPC endpoint.
// This should be called from BeginBlocker to update metrics each block.
func (c *NetworkHealthCollector) CollectNetworkHealth(ctx context.Context) (*types.NetworkHealth, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check cache first
	c.mu.RLock()
	if c.cachedHealth != nil && time.Now().Before(c.cacheExpiry) {
		cached := c.cachedHealth
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	// Collect fresh data
	health := &types.NetworkHealth{
		Timestamp:   sdkCtx.BlockTime(),
		BlockHeight: sdkCtx.BlockHeight(),
	}

	// Calculate block time from current block
	c.updateBlockTime(sdkCtx.BlockHeight(), sdkCtx.BlockTime())
	health.BlockTime = c.getAverageBlockTime()

	// Fetch peer count from RPC
	peerCount, err := c.fetchPeerCount()
	if err != nil {
		// Log but don't fail - use 0 as fallback
		sdkCtx.Logger().Debug("failed to fetch peer count", "error", err)
	}
	health.PeerCount = peerCount

	// Fetch mempool size from RPC
	mempoolSize, err := c.fetchMempoolSize()
	if err != nil {
		sdkCtx.Logger().Debug("failed to fetch mempool size", "error", err)
	}
	health.MempoolSize = mempoolSize

	// Fetch validator count from RPC
	activeValidators, totalValidators, err := c.fetchValidatorCounts()
	if err != nil {
		sdkCtx.Logger().Debug("failed to fetch validator counts", "error", err)
	}
	health.ActiveValidators = activeValidators
	health.TotalValidators = totalValidators

	// Calculate consensus health based on available data
	health.ConsensusHealth = c.calculateConsensusHealth(health)

	// Calculate network congestion (simplified: based on mempool size)
	health.NetworkCongestion = c.calculateCongestion(health.MempoolSize)

	// Cache the result
	c.mu.Lock()
	c.cachedHealth = health
	c.cacheExpiry = time.Now().Add(c.cacheDuration)
	c.mu.Unlock()

	return health, nil
}

// updateBlockTime tracks block times for calculating average
func (c *NetworkHealthCollector) updateBlockTime(height int64, blockTime time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lastBlockHeight > 0 && height > c.lastBlockHeight {
		// Calculate time since last block
		timeDiff := blockTime.Sub(c.lastBlockTime).Seconds()
		if timeDiff > 0 && timeDiff < 60 { // Sanity check: ignore if > 60s (likely node restart)
			c.blockTimes = append(c.blockTimes, timeDiff)
			// Keep only the last N block times
			if len(c.blockTimes) > c.maxBlockTimes {
				c.blockTimes = c.blockTimes[1:]
			}
		}
	}

	c.lastBlockHeight = height
	c.lastBlockTime = blockTime
}

// getAverageBlockTime calculates the average block time from recent blocks
func (c *NetworkHealthCollector) getAverageBlockTime() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.blockTimes) == 0 {
		return 0
	}

	var sum float64
	for _, t := range c.blockTimes {
		sum += t
	}
	return sum / float64(len(c.blockTimes))
}

// fetchPeerCount queries the /net_info endpoint for peer count
func (c *NetworkHealthCollector) fetchPeerCount() (int, error) {
	c.mu.RLock()
	endpoint := c.rpcEndpoint
	c.mu.RUnlock()

	if endpoint == "" {
		return 0, fmt.Errorf("RPC endpoint not configured")
	}

	resp, err := c.httpClient.Get(endpoint + "/net_info")
	if err != nil {
		return 0, fmt.Errorf("failed to query net_info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("net_info returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var netInfo rpcNetInfoResponse
	if err := json.Unmarshal(body, &netInfo); err != nil {
		return 0, fmt.Errorf("failed to parse net_info: %w", err)
	}

	var peerCount int
	fmt.Sscanf(netInfo.Result.NPeers, "%d", &peerCount)
	return peerCount, nil
}

// fetchMempoolSize queries the /unconfirmed_txs endpoint for mempool size
func (c *NetworkHealthCollector) fetchMempoolSize() (int, error) {
	c.mu.RLock()
	endpoint := c.rpcEndpoint
	c.mu.RUnlock()

	if endpoint == "" {
		return 0, fmt.Errorf("RPC endpoint not configured")
	}

	resp, err := c.httpClient.Get(endpoint + "/unconfirmed_txs?limit=1")
	if err != nil {
		return 0, fmt.Errorf("failed to query unconfirmed_txs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unconfirmed_txs returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var txs rpcUnconfirmedTxsResponse
	if err := json.Unmarshal(body, &txs); err != nil {
		return 0, fmt.Errorf("failed to parse unconfirmed_txs: %w", err)
	}

	var total int
	fmt.Sscanf(txs.Result.Total, "%d", &total)
	return total, nil
}

// fetchValidatorCounts queries the /validators endpoint for validator counts
func (c *NetworkHealthCollector) fetchValidatorCounts() (active int, total int, err error) {
	c.mu.RLock()
	endpoint := c.rpcEndpoint
	c.mu.RUnlock()

	if endpoint == "" {
		return 0, 0, fmt.Errorf("RPC endpoint not configured")
	}

	resp, err := c.httpClient.Get(endpoint + "/validators?per_page=1")
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query validators: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("validators returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read response: %w", err)
	}

	var validators rpcValidatorsResponse
	if err := json.Unmarshal(body, &validators); err != nil {
		return 0, 0, fmt.Errorf("failed to parse validators: %w", err)
	}

	fmt.Sscanf(validators.Result.Total, "%d", &total)
	// Active validators = validators with voting power > 0
	// For simplicity, we assume all returned validators are active
	active = total

	return active, total, nil
}

// calculateConsensusHealth calculates a health score from 0-1
func (c *NetworkHealthCollector) calculateConsensusHealth(health *types.NetworkHealth) float64 {
	score := 1.0

	// Penalize for no peers
	if health.PeerCount == 0 {
		score -= 0.3
	} else if health.PeerCount < 3 {
		score -= 0.1
	}

	// Penalize for slow block times (> 5s is concerning)
	if health.BlockTime > 10 {
		score -= 0.3
	} else if health.BlockTime > 5 {
		score -= 0.1
	}

	// Penalize for large mempool
	if health.MempoolSize > 1000 {
		score -= 0.2
	} else if health.MempoolSize > 100 {
		score -= 0.1
	}

	// Clamp to 0-1
	if score < 0 {
		score = 0
	}
	return score
}

// calculateCongestion calculates network congestion from 0-1 based on mempool
func (c *NetworkHealthCollector) calculateCongestion(mempoolSize int) float64 {
	// Simple linear scaling: 0 at 0 txs, 1.0 at 10000+ txs
	if mempoolSize <= 0 {
		return 0
	}
	if mempoolSize >= 10000 {
		return 1.0
	}
	return float64(mempoolSize) / 10000.0
}
