// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

// ============================
// ECONOMIC ATTACK DETECTION (Feature 6)
// ============================

// DetectEconomicAttacks analyzes recent activity for potential economic attacks
// This is a comprehensive security system that monitors for various attack patterns
// including pump & dump, flash loans, sybil attacks, wash trading, and front-running
func (k *Keeper) DetectEconomicAttacks(ctx context.Context) ([]*types.AttackAlert, error) {
	params, _ := k.GetParams(ctx)
	alerts := []*types.AttackAlert{}

	// 1. Detect pump and dump attacks
	if alert, err := k.detectPumpAndDump(ctx, params); err != nil {
		return nil, err
	} else if alert != nil {
		alerts = append(alerts, alert)
	}

	// 2. Detect flash loan attacks (rapid large movements)
	if alert, err := k.detectFlashLoanAttack(ctx, params); err != nil {
		return nil, err
	} else if alert != nil {
		alerts = append(alerts, alert)
	}

	// 3. Detect sybil attacks (many small transactions from related accounts)
	if alert, err := k.detectSybilAttack(ctx, params); err != nil {
		return nil, err
	} else if alert != nil {
		alerts = append(alerts, alert)
	}

	// 4. Detect wash trading
	if alert, err := k.detectWashTrading(ctx, params); err != nil {
		return nil, err
	} else if alert != nil {
		alerts = append(alerts, alert)
	}

	// 5. Detect front-running attacks
	if alert, err := k.detectFrontRunning(ctx, params); err != nil {
		return nil, err
	} else if alert != nil {
		alerts = append(alerts, alert)
	}

	// Store all detected alerts
	for _, alert := range alerts {
		if err := k.RecordAttackAlert(ctx, alert); err != nil {
			return nil, err
		}
	}

	return alerts, nil
}

// detectPumpAndDump detects pump and dump patterns
// Pump and dump is when attackers artificially inflate the price and then sell
func (k *Keeper) detectPumpAndDump(ctx context.Context, params types.Params) (*types.AttackAlert, error) {
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}

	// Look at recent large transactions
	recentLargeTxs := 0
	cutoff := currentTime - 3600 // Last hour

	err = k.IterateLargeTxRecords(ctx, func(record *types.LargeTxRecord) bool {
		// Timestamp is time.Time, use Unix() to get seconds
		if record.Timestamp.Unix() >= cutoff {
			recentLargeTxs++
		}
		return false // Continue iterating
	})
	if err != nil {
		return nil, err
	}

	// If >5 large transactions in an hour, flag as potential pump and dump
	if recentLargeTxs > 5 {
		return k.createAttackAlert(
			ctx,
			types.AttackTypePumpAndDump,
			types.AlertSeverity_ALERT_SEVERITY_WARNING,
			fmt.Sprintf("Detected %d large transactions in the last hour", recentLargeTxs),
			"",
			uint64(recentLargeTxs),
		)
	}

	return nil, nil
}

// detectFlashLoanAttack detects flash loan attack patterns
// Flash loans allow borrowing large amounts without collateral for a single transaction
func (k *Keeper) detectFlashLoanAttack(ctx context.Context, params types.Params) (*types.AttackAlert, error) {
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}

	// Look for rapid large transactions from same address
	addressTxCounts := make(map[string]int)
	cutoff := currentTime - 60 // Last minute

	err = k.IterateLargeTxRecords(ctx, func(record *types.LargeTxRecord) bool {
		// Timestamp is time.Time, use Unix() to get seconds
		if record.Timestamp.Unix() >= cutoff {
			addressTxCounts[record.Sender]++
		}
		return false // Continue iterating
	})
	if err != nil {
		return nil, err
	}

	// Check for addresses with multiple large transactions in short time
	// Sort addresses to ensure deterministic iteration order
	addresses := make([]string, 0, len(addressTxCounts))
	for addr := range addressTxCounts {
		addresses = append(addresses, addr)
	}
	sort.Strings(addresses)

	for _, addr := range addresses {
		count := addressTxCounts[addr]
		if count >= 3 {
			return k.createAttackAlert(
				ctx,
				types.AttackTypeFlashLoan,
				types.AlertSeverity_ALERT_SEVERITY_CRITICAL,
				fmt.Sprintf("Address %s made %d large transactions in 1 minute", addr, count),
				addr,
				uint64(count),
			)
		}
	}

	return nil, nil
}

// detectSybilAttack detects sybil attack patterns
// Sybil attacks involve creating many fake identities to gain disproportionate influence
func (k *Keeper) detectSybilAttack(ctx context.Context, params types.Params) (*types.AttackAlert, error) {
	// Detect many accounts with similar balances making coordinated transactions
	// This is a simplified heuristic using balance distribution analysis

	// Count addresses with holdings in similar ranges
	rangeCount := make(map[string]int)
	totalAddresses := 0

	// Iterate through all address holdings
	err := k.IterateUserMEVBalances(ctx, func(address string, balance string) bool {
		totalAddresses++

		// Get the actual holding for better analysis
		holding, err := k.GetAddressHolding(ctx, address)
		if err != nil || holding == "0" {
			return false // Continue iterating
		}

		// Bucket into ranges using first 4 significant digits
		if len(holding) >= 4 {
			rangeKey := holding[:4]
			rangeCount[rangeKey]++
		}
		return false // Continue iterating
	})
	if err != nil {
		return nil, err
	}

	// If we have very few addresses, don't trigger false positives
	if totalAddresses < 50 {
		return nil, nil
	}

	// Check for suspicious clustering
	// Sort range keys to ensure deterministic iteration order
	rangeKeys := make([]string, 0, len(rangeCount))
	for rangeKey := range rangeCount {
		rangeKeys = append(rangeKeys, rangeKey)
	}
	sort.Strings(rangeKeys)

	for _, rangeKey := range rangeKeys {
		count := rangeCount[rangeKey]
		// If >20% of addresses have similar balances, potential sybil attack
		threshold := totalAddresses / 5
		if count > threshold && count > 20 {
			return k.createAttackAlert(
				ctx,
				types.AttackTypeSybil,
				types.AlertSeverity_ALERT_SEVERITY_WARNING,
				fmt.Sprintf("Detected %d accounts with similar balance patterns (range: %s...) out of %d total", count, rangeKey, totalAddresses),
				"",
				uint64(count),
			)
		}
	}

	return nil, nil
}

// detectWashTrading detects wash trading patterns
// Wash trading is trading with yourself to create fake volume
func (k *Keeper) detectWashTrading(ctx context.Context, params types.Params) (*types.AttackAlert, error) {
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}

	// Detect circular trading patterns between same addresses
	// Track sender->recipient pairs
	pairCounts := make(map[string]int)
	reversePairMap := make(map[string]string)
	cutoff := currentTime - 3600 // Last hour

	err = k.IterateLargeTxRecords(ctx, func(record *types.LargeTxRecord) bool {
		// Timestamp is time.Time, use Unix() to get seconds
		if record.Timestamp.Unix() >= cutoff {
			// Create bidirectional key
			pair := record.Sender + ":" + record.Recipient
			reversePair := record.Recipient + ":" + record.Sender

			pairCounts[pair]++
			reversePairMap[pair] = reversePair

			// Check if there's reverse trading
			if pairCounts[reversePair] > 0 && pairCounts[pair] > 2 {
				// Found wash trading pattern - stop iterating
				return true
			}
		}
		return false // Continue iterating
	})
	if err != nil {
		return nil, err
	}

	// Check for circular patterns
	// Sort pairs to ensure deterministic iteration order
	pairs := make([]string, 0, len(pairCounts))
	for pair := range pairCounts {
		pairs = append(pairs, pair)
	}
	sort.Strings(pairs)

	for _, pair := range pairs {
		count := pairCounts[pair]
		reversePair := reversePairMap[pair]
		if pairCounts[reversePair] > 0 && count > 2 {
			// Extract addresses
			var sender, recipient string
			if _, err := fmt.Sscanf(pair, "%s:%s", &sender, &recipient); err != nil {
				return nil, fmt.Errorf("failed to parse pair %s: %w", pair, err)
			}

			return k.createAttackAlert(
				ctx,
				types.AttackTypeWashTrading,
				types.AlertSeverity_ALERT_SEVERITY_CRITICAL,
				fmt.Sprintf("Detected circular trading between %s and %s (%d transactions)", sender, recipient, count),
				sender,
				uint64(count),
			)
		}
	}

	return nil, nil
}

// detectFrontRunning detects front-running patterns
// Front-running is when someone sees a pending transaction and submits their own first
func (k *Keeper) detectFrontRunning(ctx context.Context, params types.Params) (*types.AttackAlert, error) {
	// Detect transactions with very high gas prices followed by similar transactions
	// This analyzes gas price spikes that may indicate front-running

	currentGasPrice := new(big.Int)
	if _, ok := currentGasPrice.SetString(params.DynamicFees.BaseFee, 10); !ok {
		return nil, nil
	}

	// Apply current multiplier
	currentGasPrice.Mul(currentGasPrice, big.NewInt(int64(params.DynamicFees.CurrentMultiplier)))
	currentGasPrice.Div(currentGasPrice, big.NewInt(types.BasisPoints))

	avgGasPrice := new(big.Int)
	if _, ok := avgGasPrice.SetString(params.DynamicFees.BaseFee, 10); !ok {
		return nil, nil
	}

	// If current gas price is >200% of base (indicating someone is paying premium), potential front-running
	threshold := new(big.Int).Mul(avgGasPrice, big.NewInt(2))

	if currentGasPrice.Cmp(threshold) > 0 {
		return k.createAttackAlert(
			ctx,
			types.AttackTypeFrontRunning,
			types.AlertSeverity_ALERT_SEVERITY_WARNING,
			fmt.Sprintf("Unusually high gas price detected: %s (base: %s)", currentGasPrice.String(), avgGasPrice.String()),
			"",
			0,
		)
	}

	return nil, nil
}

// createAttackAlert creates an attack alert with unique ID
func (k *Keeper) createAttackAlert(
	ctx context.Context,
	attackType types.AttackType,
	severity types.AlertSeverity,
	message string,
	suspectAddress string,
	evidenceCount uint64,
) (*types.AttackAlert, error) {
	currentTime, err := k.GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}

	// Generate unique alert ID using hash
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d-%s-%s", currentTime, attackType.String(), message)))
	alertID := hex.EncodeToString(h.Sum(nil))[:16]

	return &types.AttackAlert{
		AlertId:          alertID,
		AttackType:       attackType,
		Severity:         severity,
		Message:          message,
		DetectedAt:       time.Unix(currentTime, 0),
		SuspectAddress:   suspectAddress,
		EvidenceCount:    evidenceCount,
		AutoMitigated:    false,
		MitigationAction: "",
	}, nil
}

// RecordAttackAlert stores an attack alert
// In production this would persist to KV store for audit trail
func (k *Keeper) RecordAttackAlert(ctx context.Context, alert *types.AttackAlert) error {
	// In a production system, this would store the alert in KV store
	// For now, we accept the alert which can be logged/emitted
	_ = alert
	return nil
}

// GetAttackAlerts returns recent attack alerts with optional filtering
// Filters by severity level and limits the number of results
func (k *Keeper) GetAttackAlerts(limit uint64, severityFilter types.AlertSeverity) []*types.AttackAlert {
	// In production this would query the KV store with filters
	return []*types.AttackAlert{}
}

// GetAttackStatistics returns attack detection statistics
// Returns total detected, mitigated, critical count, and warning count
func (k *Keeper) GetAttackStatistics() (totalDetected, totalMitigated, criticalCount, warningCount uint64) {
	// In production this would aggregate statistics from KV store
	return 0, 0, 0, 0
}

// GetAttacksByType returns alerts grouped by attack type
func (k *Keeper) GetAttacksByType(attackType types.AttackType) []*types.AttackAlert {
	// In production this would query KV store filtered by type
	return []*types.AttackAlert{}
}

// GetRecentCriticalAttacks returns critical attacks from the last 24 hours
// This is used for real-time monitoring and alerting
func (k *Keeper) GetRecentCriticalAttacks(ctx context.Context) ([]*types.AttackAlert, error) {
	// In production this would query KV store with time range filter
	return []*types.AttackAlert{}, nil
}
