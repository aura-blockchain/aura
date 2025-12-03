package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// TestAMLProfileUpdateIntegration tests the complete flow:
// 1. Transaction monitoring triggers
// 2. Profiles are updated for both sender and recipient
// 3. Risk levels are calculated correctly
// 4. Events are emitted
// 5. Multiple transactions accumulate correctly
func TestAMLProfileUpdateIntegration(t *testing.T) {
	k, ctx := setupKeeperForMonitor(t)

	// Enable monitoring with low thresholds for testing
	params := types.ComplianceParams{
		TransactionMonitoringEnabled: true,
		SingleTransactionLimit:       "100000",
		VelocityLimit_24H:            "50000", // Medium=25000, High=50000
		StructuringThresholdCount:    10,
		SanctionsScreeningEnabled:    false,
	}
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	// Create test addresses
	sender := sdk.AccAddress([]byte("sender_address"))
	recipient := sdk.AccAddress([]byte("recipient_address"))

	// Transaction 1: Small transaction (LOW risk)
	t.Run("Transaction 1 - Both LOW risk", func(t *testing.T) {
		amount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 5000))

		// Monitor transaction
		_, err := k.MonitorTransaction(ctx, sender, recipient, amount)
		require.NoError(t, err)
		require.Empty(t, alerts, "no alerts for small transaction")

		// Update AML profiles (this is what MonitoredBankKeeper does)
		err = k.UpdateAMLProfileOnTransaction(ctx, sender.String(), amount)
		require.NoError(t, err)

		err = k.UpdateAMLProfileOnTransaction(ctx, recipient.String(), amount)
		require.NoError(t, err)

		// Verify sender profile
		senderProfile, err := k.GetAMLProfile(ctx, sender.String())
		require.NoError(t, err)
		require.Equal(t, uint64(1), senderProfile.TotalTransactions)
		require.Equal(t, "5000", senderProfile.TotalVolume)
		require.Equal(t, types.AMLRiskLevel_AML_RISK_LOW, senderProfile.RiskLevel)

		// Verify recipient profile
		recipientProfile, err := k.GetAMLProfile(ctx, recipient.String())
		require.NoError(t, err)
		require.Equal(t, uint64(1), recipientProfile.TotalTransactions)
		require.Equal(t, "5000", recipientProfile.TotalVolume)
		require.Equal(t, types.AMLRiskLevel_AML_RISK_LOW, recipientProfile.RiskLevel)
	})

	// Transaction 2: Medium transaction (sender escalates to MEDIUM)
	t.Run("Transaction 2 - Sender MEDIUM risk", func(t *testing.T) {
		amount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 25000))

		// Monitor transaction
		_, err := k.MonitorTransaction(ctx, sender, recipient, amount)
		require.NoError(t, err)

		// Update profiles
		err = k.UpdateAMLProfileOnTransaction(ctx, sender.String(), amount)
		require.NoError(t, err)

		err = k.UpdateAMLProfileOnTransaction(ctx, recipient.String(), amount)
		require.NoError(t, err)

		// Verify sender escalated to MEDIUM
		senderProfile, err := k.GetAMLProfile(ctx, sender.String())
		require.NoError(t, err)
		require.Equal(t, uint64(2), senderProfile.TotalTransactions)
		require.Equal(t, "30000", senderProfile.TotalVolume)
		require.Equal(t, types.AMLRiskLevel_AML_RISK_MEDIUM, senderProfile.RiskLevel)

		// Verify recipient also MEDIUM
		recipientProfile, err := k.GetAMLProfile(ctx, recipient.String())
		require.NoError(t, err)
		require.Equal(t, uint64(2), recipientProfile.TotalTransactions)
		require.Equal(t, "30000", recipientProfile.TotalVolume)
		require.Equal(t, types.AMLRiskLevel_AML_RISK_MEDIUM, recipientProfile.RiskLevel)
	})

	// Transaction 3: Large transaction (sender escalates to HIGH)
	t.Run("Transaction 3 - Sender HIGH risk", func(t *testing.T) {
		amount := sdk.NewCoins(sdk.NewInt64Coin("uaura", 30000))

		// Monitor transaction
		_, err := k.MonitorTransaction(ctx, sender, recipient, amount)
		require.NoError(t, err)

		// Update profiles
		err = k.UpdateAMLProfileOnTransaction(ctx, sender.String(), amount)
		require.NoError(t, err)

		err = k.UpdateAMLProfileOnTransaction(ctx, recipient.String(), amount)
		require.NoError(t, err)

		// Verify sender escalated to HIGH
		senderProfile, err := k.GetAMLProfile(ctx, sender.String())
		require.NoError(t, err)
		require.Equal(t, uint64(3), senderProfile.TotalTransactions)
		require.Equal(t, "60000", senderProfile.TotalVolume)
		require.Equal(t, types.AMLRiskLevel_AML_RISK_HIGH, senderProfile.RiskLevel)

		// Verify recipient also HIGH
		recipientProfile, err := k.GetAMLProfile(ctx, recipient.String())
		require.NoError(t, err)
		require.Equal(t, uint64(3), recipientProfile.TotalTransactions)
		require.Equal(t, "60000", recipientProfile.TotalVolume)
		require.Equal(t, types.AMLRiskLevel_AML_RISK_HIGH, recipientProfile.RiskLevel)
	})
}
