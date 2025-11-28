package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// WalletAnalytics provides analytics for a wallet
type WalletAnalytics struct {
	WalletID               string
	TotalTransactions      int64
	TotalVolume            math.Int
	AverageTransactionSize math.Int
	SecurityScore          float64
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
		return int64(sdk.BigEndianToUint64(countBytes))
	}

	return 0
}

func (k Keeper) calculateVolumes(ctx context.Context, walletID string) (math.Int, math.Int) {
	// Simplified calculation
	total := math.NewInt(0)
	count := k.countTransactions(ctx, walletID)

	if count == 0 {
		return total, math.NewInt(0)
	}

	average := total.Quo(math.NewInt(count))
	return total, average
}

func (k Keeper) calculateSecurityScore(ctx context.Context, walletID string) float64 {
	score := 100.0

	// Deduct for missing features
	if !k.hasMultiSig(ctx, walletID) {
		score -= 20
	}

	if !k.hasHardwareWallet(ctx, walletID) {
		score -= 15
	}

	if !k.hasSocialRecovery(ctx, walletID) {
		score -= 10
	}

	if !k.hasBiometric(ctx, walletID) {
		score -= 10
	}

	// Deduct for anomalies
	anomalies, _ := k.GetAnomalies(ctx, walletID)
	unresolvedAnomalies := 0
	for _, a := range anomalies {
		if !a.Resolved {
			unresolvedAnomalies++
		}
	}

	score -= float64(unresolvedAnomalies) * 5

	if score < 0 {
		score = 0
	}

	return score
}

func (k Keeper) determineRiskLevel(score float64) string {
	if score >= 80 {
		return "low"
	} else if score >= 60 {
		return "medium"
	} else if score >= 40 {
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

	report := &wsproto.SecurityReport{
		WalletId:        walletID,
		SecurityScore:   int32(analytics.SecurityScore),
		Recommendations: k.generateRecommendations(analytics),
		ActiveDevices:   int32(analytics.ActiveDevices),
	}

	return report, nil
}

func (k Keeper) generateRecommendations(analytics *WalletAnalytics) []string {
	recommendations := []string{}

	if analytics.SecurityScore < 80 {
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
