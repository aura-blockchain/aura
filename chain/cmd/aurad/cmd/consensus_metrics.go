// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// AURA Consensus Metrics (custom metrics with aura_ prefix to avoid conflicts with CometBFT built-in metrics)
var (
	// Round metrics
	consensusRound = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_round",
			Help: "Current consensus round for the given height",
		},
		[]string{"validator"},
	)

	consensusStep = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_step",
			Help: "Current consensus step (0=NewHeight, 1=NewRound, 2=Propose, 3=Prevote, 4=PrevoteWait, 5=Precommit, 6=PrecommitWait, 7=Commit)",
		},
		[]string{"validator"},
	)

	// Vote metrics
	consensusPrevotesReceived = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_prevotes_received",
			Help: "Number of prevotes received in current round",
		},
		[]string{"validator", "round"},
	)

	consensusPrecommitsReceived = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_precommits_received",
			Help: "Number of precommits received in current round",
		},
		[]string{"validator", "round"},
	)

	consensusPrevotesPower = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_prevotes_voting_power",
			Help: "Total voting power of prevotes received (out of total)",
		},
		[]string{"validator", "round"},
	)

	consensusPrecommitsPower = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_precommits_voting_power",
			Help: "Total voting power of precommits received (out of total)",
		},
		[]string{"validator", "round"},
	)

	// Height tracking
	consensusHeight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_height",
			Help: "Current consensus height",
		},
		[]string{"validator"},
	)

	consensusHeightRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_height_rate",
			Help: "Rate of height increase (blocks per minute)",
		},
		[]string{"validator"},
	)

	// Proposer tracking
	consensusIsProposer = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_is_proposer",
			Help: "Whether this validator is the proposer for current round (1=yes, 0=no)",
		},
		[]string{"validator"},
	)

	consensusProposerIndex = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_proposer_index",
			Help: "Index of the current proposer",
		},
		[]string{"validator"},
	)

	// Block proposal status
	consensusHasProposal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_has_proposal",
			Help: "Whether a block proposal has been received (1=yes, 0=no)",
		},
		[]string{"validator"},
	)

	consensusLockedBlockHash = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_has_locked_block",
			Help: "Whether this validator has locked on a block (1=yes, 0=no)",
		},
		[]string{"validator"},
	)

	consensusValidBlockHash = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_has_valid_block",
			Help: "Whether this validator has a valid block (1=yes, 0=no)",
		},
		[]string{"validator"},
	)

	// Validator participation
	consensusValidatorMissedBlocks = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aura_consensus_validator_missed_blocks_total",
			Help: "Total number of blocks missed by validator (based on nil votes)",
		},
		[]string{"validator", "validator_address"},
	)

	consensusValidatorParticipation = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "aura_consensus_validator_participation",
			Help: "Validator participation rate in consensus (0-1)",
		},
		[]string{"validator", "validator_address"},
	)

	// Round duration tracking
	consensusRoundDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aura_consensus_round_duration_seconds",
			Help:    "Duration of consensus rounds in seconds",
			Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60, 120},
		},
		[]string{"validator"},
	)

	// Consensus failures
	consensusTimeouts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aura_consensus_timeouts_total",
			Help: "Total number of consensus timeouts",
		},
		[]string{"validator", "timeout_type"},
	)
)

// ConsensusState represents the RPC response from /consensus_state
type ConsensusState struct {
	Result struct {
		RoundState struct {
			HeightRoundStep   string `json:"height/round/step"`
			StartTime         string `json:"start_time"`
			ProposalBlockHash string `json:"proposal_block_hash"`
			LockedBlockHash   string `json:"locked_block_hash"`
			ValidBlockHash    string `json:"valid_block_hash"`
			HeightVoteSet     []struct {
				Round              int      `json:"round"`
				Prevotes           []string `json:"prevotes"`
				PrevotesBitArray   string   `json:"prevotes_bit_array"`
				Precommits         []string `json:"precommits"`
				PrecommitsBitArray string   `json:"precommits_bit_array"`
			} `json:"height_vote_set"`
			Proposer struct {
				Address string `json:"address"`
				Index   int    `json:"index"`
			} `json:"proposer"`
		} `json:"round_state"`
	} `json:"result"`
}

// ConsensusMetricsCollector collects consensus metrics from CometBFT RPC
type ConsensusMetricsCollector struct {
	rpcURL     string
	moniker    string
	logger     cmtlog.Logger
	lastHeight int64
	lastTime   time.Time
	roundStart map[string]time.Time
}

// NewConsensusMetricsCollector creates a new consensus metrics collector
func NewConsensusMetricsCollector(rpcURL, moniker string, logger cmtlog.Logger) *ConsensusMetricsCollector {
	return &ConsensusMetricsCollector{
		rpcURL:     rpcURL,
		moniker:    moniker,
		logger:     logger,
		roundStart: make(map[string]time.Time),
	}
}

// Start begins collecting consensus metrics
func (c *ConsensusMetricsCollector) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	c.logger.Info("Starting consensus metrics collector", "rpc_url", c.rpcURL, "interval", interval)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Stopping consensus metrics collector")
			return
		case <-ticker.C:
			if err := c.collectMetrics(); err != nil {
				c.logger.Error("Failed to collect consensus metrics", "error", err)
			}
		}
	}
}

// collectMetrics fetches and processes consensus state from RPC
func (c *ConsensusMetricsCollector) collectMetrics() error {
	// Fetch consensus state
	resp, err := http.Get(c.rpcURL + "/consensus_state")
	if err != nil {
		return fmt.Errorf("failed to fetch consensus state: %w", err)
	}
	defer resp.Body.Close()

	var state ConsensusState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return fmt.Errorf("failed to decode consensus state: %w", err)
	}

	// Parse height/round/step
	parts := strings.Split(state.Result.RoundState.HeightRoundStep, "/")
	if len(parts) != 3 {
		return fmt.Errorf("invalid height/round/step format: %s", state.Result.RoundState.HeightRoundStep)
	}

	height, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse height: %w", err)
	}

	round, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("failed to parse round: %w", err)
	}

	step, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("failed to parse step: %w", err)
	}

	// Update basic metrics
	consensusHeight.WithLabelValues(c.moniker).Set(float64(height))
	consensusRound.WithLabelValues(c.moniker).Set(float64(round))
	consensusStep.WithLabelValues(c.moniker).Set(float64(step))

	// Calculate height rate (blocks per minute)
	now := time.Now()
	if c.lastHeight > 0 && !c.lastTime.IsZero() {
		duration := now.Sub(c.lastTime).Minutes()
		if duration > 0 {
			heightDiff := height - c.lastHeight
			rate := float64(heightDiff) / duration
			consensusHeightRate.WithLabelValues(c.moniker).Set(rate)
		}
	}
	c.lastHeight = height
	c.lastTime = now

	// Update proposer metrics
	consensusProposerIndex.WithLabelValues(c.moniker).Set(float64(state.Result.RoundState.Proposer.Index))

	// Check if we are the proposer (would need validator address comparison)
	// For now, we track if there is a proposer
	if state.Result.RoundState.Proposer.Address != "" {
		consensusIsProposer.WithLabelValues(c.moniker).Set(0) // Default to no
	}

	// Update block proposal status
	if state.Result.RoundState.ProposalBlockHash != "" {
		consensusHasProposal.WithLabelValues(c.moniker).Set(1)
	} else {
		consensusHasProposal.WithLabelValues(c.moniker).Set(0)
	}

	if state.Result.RoundState.LockedBlockHash != "" {
		consensusLockedBlockHash.WithLabelValues(c.moniker).Set(1)
	} else {
		consensusLockedBlockHash.WithLabelValues(c.moniker).Set(0)
	}

	if state.Result.RoundState.ValidBlockHash != "" {
		consensusValidBlockHash.WithLabelValues(c.moniker).Set(1)
	} else {
		consensusValidBlockHash.WithLabelValues(c.moniker).Set(0)
	}

	// Process vote sets
	for _, voteSet := range state.Result.RoundState.HeightVoteSet {
		roundStr := strconv.Itoa(voteSet.Round)

		// Count prevotes
		prevoteCount := 0
		for _, vote := range voteSet.Prevotes {
			if vote != "nil-Vote" {
				prevoteCount++
			}
		}
		consensusPrevotesReceived.WithLabelValues(c.moniker, roundStr).Set(float64(prevoteCount))

		// Count precommits
		precommitCount := 0
		for _, vote := range voteSet.Precommits {
			if vote != "nil-Vote" {
				precommitCount++
			}
		}
		consensusPrecommitsReceived.WithLabelValues(c.moniker, roundStr).Set(float64(precommitCount))

		// Parse voting power from bit array (format: "BA{4:____} 0/3600000 = 0.00")
		if power := parseBitArrayPower(voteSet.PrevotesBitArray); power >= 0 {
			consensusPrevotesPower.WithLabelValues(c.moniker, roundStr).Set(power)
		}

		if power := parseBitArrayPower(voteSet.PrecommitsBitArray); power >= 0 {
			consensusPrecommitsPower.WithLabelValues(c.moniker, roundStr).Set(power)
		}
	}

	// Track round duration
	roundKey := fmt.Sprintf("%d/%d", height, round)
	if startTime, exists := c.roundStart[roundKey]; exists {
		duration := now.Sub(startTime).Seconds()
		consensusRoundDuration.WithLabelValues(c.moniker).Observe(duration)
	} else {
		c.roundStart[roundKey] = now
	}

	// Clean up old round start times (keep last 10)
	if len(c.roundStart) > 10 {
		// Simple cleanup: remove all entries
		c.roundStart = make(map[string]time.Time)
		c.roundStart[roundKey] = now
	}

	return nil
}

// parseBitArrayPower parses voting power from bit array string
// Format: "BA{4:____} 0/3600000 = 0.00"
func parseBitArrayPower(bitArray string) float64 {
	parts := strings.Split(bitArray, " ")
	if len(parts) < 2 {
		return -1
	}

	// Extract the fraction part (e.g., "0/3600000")
	fractionParts := strings.Split(parts[1], "/")
	if len(fractionParts) != 2 {
		return -1
	}

	current, err := strconv.ParseFloat(fractionParts[0], 64)
	if err != nil {
		return -1
	}

	total, err := strconv.ParseFloat(fractionParts[1], 64)
	if err != nil || total == 0 {
		return -1
	}

	return current
}
