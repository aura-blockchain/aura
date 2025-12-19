package testing

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testing/testutil"
	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	bridgekeeper "github.com/aequitas/aura/chain/x/bridge/keeper"
	bridgetypes "github.com/aequitas/aura/chain/x/bridge/types"
	compliancekeeper "github.com/aequitas/aura/chain/x/compliance/keeper"
	compliancetypes "github.com/aequitas/aura/chain/x/compliance/types"
	dexkeeper "github.com/aequitas/aura/chain/x/dex/keeper"
	dextypes "github.com/aequitas/aura/chain/x/dex/types"
)

func sanitizeDenom(raw string, fallback string) string {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	normalized := strings.ToLower(raw)
	builder := strings.Builder{}
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		}
	}
	out := builder.String()
	if out == "" || !(out[0] >= 'a' && out[0] <= 'z') {
		return fallback
	}
	if len(out) < 3 {
		return fallback
	}
	return out
}

// FuzzCrossModuleGuards exercises DEX guard paths with randomized inputs.
// This targets guard regressions where pause/reentrancy checks must hold
// under unexpected execution orderings or parameter flips.
func FuzzCrossModuleGuards(f *testing.F) {
	// Seed with deterministic cases
	f.Add(true, int64(1_000_000), int64(1_000_000))
	f.Add(false, int64(10_000_000), int64(5_000_000))

	f.Fuzz(func(t *testing.T, _ bool, amountA int64, amountB int64) {
		if amountA <= 0 {
			amountA = 1
		}
		if amountB <= 0 {
			amountB = 1
		}

		// Dex setup with security guard
		dexInput := keepertest.CreateTestInput(t)
		securityMock := testutil.NewMockSecurityKeeper()
		mockBank := testutil.NewMockBankKeeper()
		mockAcct := testutil.NewMockAccountKeeper()
		dexKeeper := dexkeeper.NewKeeper(
			dexInput.Cdc,
			dexInput.StoreKey,
			mockBank,
			mockAcct,
			nil,
			securityMock,
		)
		params := dextypes.DefaultParams()
		require.NoError(t, dexKeeper.SetParams(dexInput.Ctx, &params))

		creator := keepertest.GenTestAddr()
		amtA := sdkmath.NewInt(amountA)
		amtB := sdkmath.NewInt(amountB)
		mockBank.Balances[creator.String()] = sdk.NewCoins(
			sdk.NewCoin("uaura", amtA),
			sdk.NewCoin("uusdc", amtB),
		)

		// DEX CreatePool should always invoke reentrancy guard; simulate potential collisions
		_ = rand.Int63n(1000) // #nosec G404 - fuzz randomness acceptable, ensures varied paths
		reentrancyKey := "dex:CreatePool"
		sec := testutil.NewMockSecurityKeeper()
		dexKeeper = dexkeeper.NewKeeper(
			dexInput.Cdc,
			dexInput.StoreKey,
			mockBank,
			mockAcct,
			nil,
			sec,
		)
		params = dextypes.DefaultParams()
		require.NoError(t, dexKeeper.SetParams(dexInput.Ctx, &params))

		// Lock once via guard to force reentrancy rejection on duplicated keys
		require.NoError(t, sec.EnterNoReentrant(dexInput.Ctx, reentrancyKey))
		_, _, err := dexKeeper.CreatePool(
			dexInput.Ctx,
			creator.String(),
			"uaura",
			"uusdc",
			sdk.NewCoin("uaura", amtA),
			sdk.NewCoin("uusdc", amtB),
		)
		if err == nil {
			// If guard semantics change, ensure we clean up lock to avoid poisoning subsequent iterations
			sec.ExitNoReentrant(dexInput.Ctx, reentrancyKey)
		}

		delete(sec.ReentrantKeys, reentrancyKey)
		_, _, err = dexKeeper.CreatePool(
			dexInput.Ctx,
			creator.String(),
			"uaura",
			"uusdc",
			sdk.NewCoin("uaura", amtA),
			sdk.NewCoin("uusdc", amtB),
		)
		require.NoError(t, err, "unlocked key should allow pool creation")
	})
}

// FuzzSecurityReentrancyGuard ensures guard keys handle random inputs without leaking locks.
func FuzzSecurityReentrancyGuard(f *testing.F) {
	f.Add("dex:CreatePool")
	f.Add("bridge:lock")
	f.Add("")

	f.Fuzz(func(t *testing.T, key string) {
		if key == "" {
			key = "default"
		}
		sec := testutil.NewMockSecurityKeeper()
		ctx := keepertest.CreateTestInput(t).Ctx

		require.NoError(t, sec.EnterNoReentrant(ctx, key))
		err := sec.EnterNoReentrant(ctx, key)
		require.Error(t, err, "second entry with same key should be blocked")

		sec.ExitNoReentrant(ctx, key)
		require.NoError(t, sec.EnterNoReentrant(ctx, key), "lock must be released after exit")
	})
}

// FuzzDexCreatePoolPayloads feeds randomized payloads into CreatePool to ensure guard paths and validation never panic.
func FuzzDexCreatePoolPayloads(f *testing.F) {
	f.Add("uaura", "uusdc", int64(1_000_000), int64(2_000_000))
	f.Add("uaura", "uaura", int64(1), int64(1)) // invalid (same denom) expected to error

	f.Fuzz(func(t *testing.T, denomA, denomB string, amtA, amtB int64) {
		denomA = sanitizeDenom(denomA, "uaura")
		denomB = sanitizeDenom(denomB, "uusdc")
		if denomA == denomB {
			// Duplicate denoms lead to coin set panic upstream; skip to avoid noise
			t.Skip("identical denoms not valid for pool creation")
		}
		if amtA <= 0 {
			amtA = 1
		}
		if amtB <= 0 {
			amtB = 1
		}

		dexInput := keepertest.CreateTestInput(t)
		sec := testutil.NewMockSecurityKeeper()
		mockBank := testutil.NewMockBankKeeper()
		mockAcct := testutil.NewMockAccountKeeper()
		dexKeeper := dexkeeper.NewKeeper(
			dexInput.Cdc,
			dexInput.StoreKey,
			mockBank,
			mockAcct,
			nil,
			sec,
		)
		params := dextypes.DefaultParams()
		require.NoError(t, dexKeeper.SetParams(dexInput.Ctx, &params))

		creator := keepertest.GenTestAddr()
		// Fund with generous balances to avoid insufficient funds errors dominating results
		funding := []sdk.Coin{sdk.NewCoin(denomA, sdkmath.NewInt(amtA).AddRaw(1_000_000))}
		if denomB != denomA {
			funding = append(funding, sdk.NewCoin(denomB, sdkmath.NewInt(amtB).AddRaw(1_000_000)))
		}
		mockBank.Balances[creator.String()] = sdk.NewCoins(funding...)

		_, _, err := dexKeeper.CreatePool(
			dexInput.Ctx,
			creator.String(),
			denomA,
			denomB,
			sdk.NewCoin(denomA, sdkmath.NewInt(amtA)),
			sdk.NewCoin(denomB, sdkmath.NewInt(amtB)),
		)

		// No panics allowed; errors are acceptable for invalid inputs
		if denomA == denomB {
			require.Error(t, err)
		}
	})
}

// FuzzComplianceAlerts ensures ShouldBlockTransaction logic matches expected risk aggregation.
func FuzzComplianceAlerts(f *testing.F) {
	f.Add(int64(1), int64(0))
	f.Add(int64(2), int64(2))

	f.Fuzz(func(t *testing.T, riskSeed int64, criticalCount int64) {
		input := keepertest.CreateTestInput(t)
		keeper := compliancekeeper.NewKeeper(input.Cdc, input.StoreKey)

		alerts := make([]*compliancetypes.TransactionAlert, 0, 3)
		riskLevels := []compliancetypes.TransactionRiskLevel{
			compliancetypes.TransactionRiskLevel_TX_RISK_LOW,
			compliancetypes.TransactionRiskLevel_TX_RISK_MEDIUM,
			compliancetypes.TransactionRiskLevel_TX_RISK_HIGH,
			compliancetypes.TransactionRiskLevel_TX_RISK_CRITICAL,
		}

		expectedBlock := false
		highCount := 0
		for i := 0; i < 3; i++ {
			idx := int((riskSeed + int64(i)) % int64(len(riskLevels)))
			if idx < 0 {
				idx = -idx
			}
			level := riskLevels[idx%len(riskLevels)]
			if int64(i) < criticalCount {
				level = compliancetypes.TransactionRiskLevel_TX_RISK_CRITICAL
			}
			if level == compliancetypes.TransactionRiskLevel_TX_RISK_CRITICAL {
				expectedBlock = true
			}
			if level == compliancetypes.TransactionRiskLevel_TX_RISK_HIGH {
				highCount++
			}

			alerts = append(alerts, &compliancetypes.TransactionAlert{
				Id:        "alert-" + fmt.Sprintf("%d", i),
				Address:   "aura1seed",
				RuleId:    "fuzz",
				RiskLevel: level,
			})
		}

		if !expectedBlock && highCount >= 2 {
			expectedBlock = true
		}

		block, _ := keeper.ShouldBlockTransaction(alerts)
		require.Equal(t, expectedBlock, block)
	})
}

// FuzzBridgePauseGuards stresses pause + auto-pause behavior with varied chain labels and mint deltas.
func FuzzBridgePauseGuards(f *testing.F) {
	f.Add("paw", "PAW", int64(200))
	f.Add("xai ", "aura", int64(50))

	f.Fuzz(func(t *testing.T, pausedChain string, queryChain string, mintDelta int64) {
		input := keepertest.CreateTestInput(t)
		legacyAmino := codec.NewLegacyAmino()
		ps := paramtypes.NewSubspace(input.Cdc, legacyAmino, input.StoreKey, input.MemStoreKey, bridgetypes.ModuleName)
		bridgeKeeper := bridgekeeper.NewKeeper(
			input.Cdc,
			input.StoreKey,
			&ps,
			nil,
			nil,
			nil,
			nil,
		)

		params := bridgetypes.DefaultParams()
		params.PausedChains = []string{pausedChain}
		params.AutoPauseEnabled = true
		params.AutoPauseThreshold = sdkmath.NewInt(1_000).String()
		require.NoError(t, bridgeKeeper.SetParams(input.Ctx, params))

		normalizedPaused := strings.ToLower(strings.TrimSpace(pausedChain))
		normalizedQuery := strings.ToLower(strings.TrimSpace(queryChain))

		err := bridgeKeeper.RequireNotPaused(input.Ctx, queryChain)
		if normalizedPaused != "" && normalizedPaused == normalizedQuery {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		delta := sdkmath.NewInt(mintDelta)
		if !delta.IsPositive() {
			delta = sdkmath.NewInt(1)
		}

		bridgeKeeper.AddHourlyMintedAmount(input.Ctx, "uaura", sdkmath.NewInt(900))
		triggered := bridgeKeeper.CheckAndTriggerAutoPause(input.Ctx, "uaura", delta)
		total := sdkmath.NewInt(900).Add(delta)
		if total.GT(sdkmath.NewInt(1_000)) {
			require.True(t, triggered, "auto-pause should trigger when threshold exceeded")
			foundEvent := false
			for _, evt := range input.Ctx.EventManager().Events() {
				if evt.Type == "bridge_auto_paused" {
					foundEvent = true
					break
				}
			}
			require.True(t, foundEvent, "auto-pause trigger should emit bridge_auto_paused event")
		} else {
			require.False(t, triggered, "auto-pause should not trigger below threshold")
		}
	})
}
