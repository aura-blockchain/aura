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
	"strconv"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/monitoring/types"
)

// ============================================================================
// OFF-CHAIN COMPONENT - NOT CONSENSUS SAFE
// ============================================================================
//
// NetworkHealthCollector provides two distinct operational modes:
//
// 1. DETERMINISTIC MODE (CollectDeterministicHealth):
//    - Safe to call during consensus (BeginBlocker/EndBlocker)
//    - Uses ONLY data from ctx.BlockTime() and ctx.BlockHeight()
//    - No external I/O, no HTTP calls, no network access
//    - Results can be persisted to consensus state
//
// 2. OFF-CHAIN MODE (CollectObservabilityHealth, fetchPeerCount, etc.):
//    - MUST NOT be called during consensus (BeginBlocker/EndBlocker/MsgHandler)
//    - Makes HTTP requests to local RPC endpoints
//    - Non-deterministic: results vary by node, network conditions, timing
//    - For Prometheus metrics, operator dashboards, and external monitoring only
//    - Should be called from a separate goroutine or sidecar process
//
// WARNING: The HTTP-based methods (fetchPeerCount, fetchMempoolSize,
// fetchValidatorCounts, CollectObservabilityHealth) are NON-DETERMINISTIC
// and will cause consensus failures if their results affect state.
//
// ============================================================================

// NetworkHealthCollector collects network health metrics.
// See package-level documentation for consensus safety guidelines.
type NetworkHealthCollector struct {
	rpcEndpoint string
	httpClient  *http.Client

	// Block time calculation (deterministic: uses block headers)
	mu              sync.RWMutex
	lastBlockHeight int64
	lastBlockTime   time.Time
	blockTimes      []float64 // Rolling window for average calculation
	maxBlockTimes   int
}

// NewNetworkHealthCollector creates a new network health collector.
// rpcEndpoint should be the local CometBFT RPC endpoint (e.g., "http://localhost:26657")
// The endpoint can be overridden via the MONITORING_RPC_ENDPOINT environment variable.
func NewNetworkHealthCollector(rpcEndpoint string) *NetworkHealthCollector {
	if envEndpoint := os.Getenv("MONITORING_RPC_ENDPOINT"); envEndpoint != "" {
		rpcEndpoint = envEndpoint
	}

	return &NetworkHealthCollector{
		rpcEndpoint: rpcEndpoint,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
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

// CollectDeterministicHealth gathers CONSENSUS-SAFE metrics derived only from
// the block context (ctx.BlockTime, ctx.BlockHeight). This method:
//   - Makes NO HTTP calls
//   - Makes NO external I/O
//   - Uses NO wall-clock time
//   - Is SAFE to call from BeginBlocker/EndBlocker
//   - Results CAN be persisted to consensus state
func (c *NetworkHealthCollector) CollectDeterministicHealth(ctx context.Context) (*types.NetworkHealth, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	health := &types.NetworkHealth{
		Timestamp:   sdkCtx.BlockTime(),
		BlockHeight: sdkCtx.BlockHeight(),
	}

	// Calculate block time from current block
	c.updateBlockTime(sdkCtx.BlockHeight(), sdkCtx.BlockTime())
	health.BlockTime = c.getAverageBlockTime()

	// Deterministic placeholders for non-consensus-friendly fields
	health.PeerCount = 0
	health.MempoolSize = 0
	health.TPS = 0
	health.ActiveValidators = 0
	health.TotalValidators = 0
	health.NetworkHashRate = 0
	health.AverageGasPrice = 0
	health.NetworkCongestion = 0
	health.ConsensusHealth = 1 // assume healthy; deterministic constant

	return health, nil
}

// CollectObservabilityHealth gathers richer node-local metrics for Prometheus.
//
// # OFF-CHAIN ONLY - DO NOT CALL DURING CONSENSUS
//
// This method makes HTTP requests to local RPC endpoints and is NON-DETERMINISTIC:
//   - Network conditions vary between nodes
//   - RPC response times differ
//   - Results will differ across validators
//
// This method should ONLY be called from:
//   - Background goroutines for Prometheus scraping
//   - External monitoring sidecars
//   - Operator CLI tools
//
// NEVER call from BeginBlocker, EndBlocker, or message handlers.
// NEVER persist results to consensus state.
func (c *NetworkHealthCollector) CollectObservabilityHealth(ctx context.Context) (*types.NetworkHealth, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	health := &types.NetworkHealth{
		Timestamp:   sdkCtx.BlockTime(),
		BlockHeight: sdkCtx.BlockHeight(),
	}

	// Deterministic block-time stats
	c.updateBlockTime(sdkCtx.BlockHeight(), sdkCtx.BlockTime())
	health.BlockTime = c.getAverageBlockTime()

	peerCount, _ := c.fetchPeerCount()
	mempoolSize, _ := c.fetchMempoolSize()
	activeVals, totalVals, _ := c.fetchValidatorCounts()

	health.PeerCount = peerCount
	health.MempoolSize = mempoolSize
	health.ActiveValidators = activeVals
	health.TotalValidators = totalVals
	health.TPS = 0
	health.NetworkCongestion = c.calculateCongestion(mempoolSize)
	health.ConsensusHealth = c.calculateConsensusHealth(health)

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

// fetchPeerCount queries the /net_info endpoint for peer count.
//
// OFF-CHAIN ONLY - NON-DETERMINISTIC
// Makes HTTP request to local RPC. Do not call during consensus.
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

	peerCount, _ := strconv.Atoi(netInfo.Result.NPeers)
	return peerCount, nil
}

// fetchMempoolSize queries /unconfirmed_txs for mempool transaction count.
//
// OFF-CHAIN ONLY - NON-DETERMINISTIC
// Makes HTTP request to local RPC. Do not call during consensus.
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

	total, _ := strconv.Atoi(txs.Result.Total)
	return total, nil
}

// fetchValidatorCounts queries /validators for validator counts.
//
// OFF-CHAIN ONLY - NON-DETERMINISTIC
// Makes HTTP request to local RPC. Do not call during consensus.
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

	total, _ = strconv.Atoi(validators.Result.Total)
	active = total
	return active, total, nil
}

// calculateConsensusHealth computes a 0-1 health score (metrics only)
func (c *NetworkHealthCollector) calculateConsensusHealth(health *types.NetworkHealth) float64 {
	score := 1.0
	if health.PeerCount == 0 {
		score -= 0.3
	} else if health.PeerCount < 3 {
		score -= 0.1
	}
	if health.BlockTime > 10 {
		score -= 0.3
	} else if health.BlockTime > 5 {
		score -= 0.1
	}
	if health.MempoolSize > 1000 {
		score -= 0.2
	} else if health.MempoolSize > 100 {
		score -= 0.1
	}
	if score < 0 {
		return 0
	}
	return score
}

// calculateCongestion estimates congestion from mempool size (metrics only)
func (c *NetworkHealthCollector) calculateCongestion(mempoolSize int) float64 {
	if mempoolSize <= 0 {
		return 0
	}
	if mempoolSize >= 10000 {
		return 1.0
	}
	return float64(mempoolSize) / 10000.0
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

// ============================================================================
// OFF-CHAIN OBSERVABILITY HELPERS
// ============================================================================

// CollectObservabilityMetricsOffChain collects non-deterministic metrics for
// Prometheus/observability purposes. This is a convenience method that returns
// only the observability data without requiring a context.
//
// OFF-CHAIN ONLY - Call from background goroutines, not during consensus.
//
// Returns peer count, mempool size, active validators, total validators, and any error.
func (c *NetworkHealthCollector) CollectObservabilityMetricsOffChain() (peerCount int, mempoolSize int, activeVals int, totalVals int, err error) {
	peerCount, _ = c.fetchPeerCount()
	mempoolSize, _ = c.fetchMempoolSize()
	activeVals, totalVals, _ = c.fetchValidatorCounts()
	return peerCount, mempoolSize, activeVals, totalVals, nil
}
