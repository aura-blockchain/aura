// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SecurityScoreBasisPoints represents 100% as 10000 basis points for deterministic calculations
const SecurityScoreBasisPoints = int64(10000)

// WalletAnalytics provides analytics for a wallet
type WalletAnalytics struct {
	WalletID               string
	TotalTransactions      int64
	TotalVolume            sdkmath.Int
	AverageTransactionSize sdkmath.Int
	SecurityScore          int64 // Basis points (10000 = 100%)
	RiskLevel              string
	ActiveDevices          int
	EnabledFeatures        []string
}

// GetWalletAnalytics retrieves comprehensive analytics for a wallet
func (k Keeper) GetWalletAnalytics(ctx context.Context, walletID string) (*WalletAnalytics, error) {
	analytics := &WalletAnalytics{
		WalletID:        walletID,
		EnabledFeatures: []string{},
	}

	// Count transactions
	analytics.TotalTransactions = k.countTransactions(ctx, walletID)

	// Calculate volumes
	analytics.TotalVolume, analytics.AverageTransactionSize = k.calculateVolumes(ctx, walletID)

	// Calculate security score
	analytics.SecurityScore = k.calculateSecurityScore(ctx, walletID)

	// Determine risk level
	analytics.RiskLevel = k.determineRiskLevel(analytics.SecurityScore)

	// Count active devices
	devices, _ := k.GetDevices(ctx, walletID)
	for _, d := range devices {
		if d.Trusted {
			analytics.ActiveDevices++
		}
	}

	// List enabled features
	analytics.EnabledFeatures = k.getEnabledFeatures(ctx, walletID)

	return analytics, nil
}

func (k Keeper) countTransactions(ctx context.Context, walletID string) int64 {
	kvStore := k.getStore(ctx)
	key := []byte(fmt.Sprintf("tx_count_%s", walletID))

	countBytes, _ := kvStore.Get(key)
	if countBytes != nil {
		return clampUint64ToInt64(sdk.BigEndianToUint64(countBytes))
	}

	return 0
}

func (k Keeper) calculateVolumes(ctx context.Context, walletID string) (sdkmath.Int, sdkmath.Int) {
	// Simplified calculation
	total := sdkmath.NewInt(0)
	count := k.countTransactions(ctx, walletID)

	if count == 0 {
		return total, sdkmath.NewInt(0)
	}

	average := total.Quo(sdkmath.NewInt(count))
	return total, average
}

// calculateSecurityScore returns the security score in basis points (10000 = 100%)
// Uses integer arithmetic for deterministic cross-platform consensus
func (k Keeper) calculateSecurityScore(ctx context.Context, walletID string) int64 {
	// Start with 100% (10000 basis points)
	score := SecurityScoreBasisPoints

	// Deduct for missing features (in basis points)
	// 20% = 2000 basis points
	if !k.hasMultiSig(ctx, walletID) {
		score -= 2000
	}

	// 15% = 1500 basis points
	if !k.hasHardwareWallet(ctx, walletID) {
		score -= 1500
	}

	// 10% = 1000 basis points
	if !k.hasSocialRecovery(ctx, walletID) {
		score -= 1000
	}

	// 10% = 1000 basis points
	if !k.hasBiometric(ctx, walletID) {
		score -= 1000
	}

	// Deduct for anomalies
	anomalies, _ := k.GetAnomalies(ctx, walletID)
	unresolvedAnomalies := int64(0)
	for _, a := range anomalies {
		if !a.Resolved {
			unresolvedAnomalies++
		}
	}

	// 5% per anomaly = 500 basis points per anomaly
	score -= unresolvedAnomalies * 500

	if score < 0 {
		score = 0
	}

	return score
}

// determineRiskLevel converts a score in basis points to a risk level
// 80% = 8000 basis points, 60% = 6000 basis points, 40% = 4000 basis points
func (k Keeper) determineRiskLevel(score int64) string {
	if score >= 8000 {
		return "low"
	} else if score >= 6000 {
		return "medium"
	} else if score >= 4000 {
		return "high"
	}
	return "critical"
}

func (k Keeper) getEnabledFeatures(ctx context.Context, walletID string) []string {
	features := []string{}

	if k.hasMultiSig(ctx, walletID) {
		features = append(features, "multisig")
	}

	if k.hasHardwareWallet(ctx, walletID) {
		features = append(features, "hardware_wallet")
	}

	if k.hasSocialRecovery(ctx, walletID) {
		features = append(features, "social_recovery")
	}

	if k.hasBiometric(ctx, walletID) {
		features = append(features, "biometric")
	}

	return features
}

func (k Keeper) hasMultiSig(ctx context.Context, walletID string) bool {
	_, err := k.GetMultiSigWallet(ctx, walletID)
	return err == nil
}

func (k Keeper) hasHardwareWallet(ctx context.Context, walletID string) bool {
	_, err := k.GetHardwareWallet(ctx, walletID)
	return err == nil
}

func (k Keeper) hasSocialRecovery(ctx context.Context, walletID string) bool {
	_, err := k.GetSocialRecoveryConfig(ctx, walletID)
	return err == nil
}

func (k Keeper) hasBiometric(ctx context.Context, walletID string) bool {
	_, err := k.GetBiometricAuth(ctx, walletID)
	return err == nil
}

// GenerateSecurityReport generates a comprehensive security report
func (k Keeper) GenerateSecurityReport(ctx context.Context, walletID string) (*wsproto.SecurityReport, error) {
	analytics, err := k.GetWalletAnalytics(ctx, walletID)
	if err != nil {
		return nil, err
	}

	// Convert basis points to percentage for report (10000 basis points = 100%)
	scorePercent := int32(analytics.SecurityScore / 100)

	report := &wsproto.SecurityReport{
		WalletId:        walletID,
		SecurityScore:   scorePercent,
		Recommendations: k.generateRecommendations(analytics),
		ActiveDevices:   int32(analytics.ActiveDevices),
	}

	return report, nil
}

// generateRecommendations creates recommendations based on security score
// Score is in basis points (10000 = 100%), so 80% = 8000 basis points
func (k Keeper) generateRecommendations(analytics *WalletAnalytics) []string {
	recommendations := []string{}

	// 80% = 8000 basis points
	if analytics.SecurityScore < 8000 {
		if !k.contains(analytics.EnabledFeatures, "multisig") {
			recommendations = append(recommendations, "Enable multi-signature protection")
		}

		if !k.contains(analytics.EnabledFeatures, "hardware_wallet") {
			recommendations = append(recommendations, "Use a hardware wallet for enhanced security")
		}

		if !k.contains(analytics.EnabledFeatures, "social_recovery") {
			recommendations = append(recommendations, "Configure social recovery")
		}

		if !k.contains(analytics.EnabledFeatures, "biometric") {
			recommendations = append(recommendations, "Enable biometric authentication")
		}
	}

	return recommendations
}

func (k Keeper) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
