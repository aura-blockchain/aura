// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/aiassistant/types"
)

// setupKeeper creates a test keeper for aiassistant module
func setupKeeper(t *testing.T) (*Keeper, sdk.Context) {
	t.Helper()
	input := keepertest.CreateTestInputWithStoreKey(t, types.StoreKey)

	bank := newMockBankKeeper()
	k := NewKeeper(input.Cdc, input.StoreKey, "", bank)
	return &k, input.Ctx
}

// mockBankKeeper is a minimal bank keeper for testing
type mockBankKeeper struct{}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{}
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(_ sdk.Context, _ sdk.AccAddress, _ string, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToModule(_ sdk.Context, _, _ string, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) BurnCoins(_ sdk.Context, _ string, _ sdk.Coins) error {
	return nil
}

func TestRecordAnalytics(t *testing.T) {
	k, ctx := setupKeeper(t)

	tests := []struct {
		name    string
		record  AnalyticsRecord
		wantErr bool
	}{
		{
			name: "valid analytics record",
			record: AnalyticsRecord{
				UserAddress:   "user1",
				ModelHash:     "model-hash-1",
				ComputeUnits:  100,
				Cost:          sdkmath.NewInt(1000),
				OperationType: "query",
				Success:       true,
				ResponseTime:  500,
			},
			wantErr: false,
		},
		{
			name: "high compute units",
			record: AnalyticsRecord{
				UserAddress:   "user2",
				ModelHash:     "model-hash-2",
				ComputeUnits:  1000000,
				Cost:          sdkmath.NewInt(10000000),
				OperationType: "training",
				Success:       true,
				ResponseTime:  5000,
			},
			wantErr: false,
		},
		{
			name: "failed operation",
			record: AnalyticsRecord{
				UserAddress:   "user3",
				ModelHash:     "model-hash-1",
				ComputeUnits:  50,
				Cost:          sdkmath.NewInt(500),
				OperationType: "inference",
				Success:       false,
				ResponseTime:  100,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k.RecordAnalytics(ctx, tt.record)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecordAnalytics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
