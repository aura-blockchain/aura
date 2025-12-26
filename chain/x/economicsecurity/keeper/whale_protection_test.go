// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/economicsecurity/types"
)

func TestCheckWhaleProtection(t *testing.T) {
	tests := []struct {
		name        string
		setupParams func(*types.Params)
		setup       func(*Keeper, context.Context)
		sender      string
		recipient   string
		amount      string
		wantErr     bool
		expectedErr error
	}{
		{
			name: "whale protection disabled allows all",
			setupParams: func(p *types.Params) {
				p.WhaleProtection.Enabled = false
			},
			sender:    "aura1sender",
			recipient: "aura1recipient",
			amount:    "999999999999999",
			wantErr:   false,
		},
		{
			name: "exempted sender bypasses protection",
			setupParams: func(p *types.Params) {
				p.WhaleProtection.Enabled = true
				p.WhaleProtection.ExemptedAddresses = []string{"aura1exempted"}
			},
			sender:    "aura1exempted",
			recipient: "aura1recipient",
			amount:    "999999999999999",
			wantErr:   false,
		},
		{
			name: "exempted recipient bypasses protection",
			setupParams: func(p *types.Params) {
				p.WhaleProtection.Enabled = true
				p.WhaleProtection.ExemptedAddresses = []string{"aura1exempted"}
			},
			sender:    "aura1sender",
			recipient: "aura1exempted",
			amount:    "999999999999999",
			wantErr:   false,
		},
		{
			name: "invalid amount format fails",
			setupParams: func(p *types.Params) {
				p.WhaleProtection.Enabled = true
			},
			sender:      "aura1sender",
			recipient:   "aura1recipient",
			amount:      "invalid",
			wantErr:     true,
			expectedErr: types.ErrInvalidAmount,
		},
		{
			name: "transaction exceeds max tx percentage",
			setupParams: func(p *types.Params) {
				p.WhaleProtection.Enabled = true
				p.WhaleProtection.MaxTxPercentage = 100              // 1%
				p.Tokenomics.CirculatingSupply = "1000000000000000" // 1 quadrillion
			},
			sender:      "aura1sender",
			recipient:   "aura1recipient",
			amount:      "50000000000000", // 5% of supply
			wantErr:     true,
			expectedErr: types.ErrWhaleTxLimitExceeded,
		},
		{
			name: "normal transaction succeeds",
			setupParams: func(p *types.Params) {
				p.WhaleProtection.Enabled = true
				p.WhaleProtection.MaxTxPercentage = 500             // 5%
				p.WhaleProtection.MaxHoldingPercentage = 1000       // 10%
				p.WhaleProtection.LargeTxThreshold = 100            // 1%
				p.Tokenomics.CirculatingSupply = "1000000000000000" // 1 quadrillion
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetCurrentTime(ctx, time.Now().Unix())
				_ = k.SetCurrentHeight(ctx, 100)
			},
			sender:    "aura1sender",
			recipient: "aura1recipient",
			amount:    "1000000000000", // 0.1% of supply
			wantErr:   false,
		},
		{
			name: "large tx during cooldown fails",
			setupParams: func(p *types.Params) {
				p.WhaleProtection.Enabled = true
				p.WhaleProtection.MaxTxPercentage = 500
				p.WhaleProtection.MaxHoldingPercentage = 1000
				p.WhaleProtection.LargeTxThreshold = 100
				p.WhaleProtection.LargeTxCooldown = 3600
				p.Tokenomics.CirculatingSupply = "1000000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)
				_ = k.SetCurrentHeight(ctx, 100)
				// Set last large tx time to 30 minutes ago (within 1 hour cooldown)
				_ = k.SetLastLargeTxTime(ctx, "aura1sender", currentTime-1800)
			},
			sender:      "aura1sender",
			recipient:   "aura1recipient",
			amount:      "20000000000000", // 2% - above threshold, triggers cooldown check
			wantErr:     true,
			expectedErr: types.ErrLargeTxCooldownActive,
		},
		{
			name: "large tx after cooldown succeeds",
			setupParams: func(p *types.Params) {
				p.WhaleProtection.Enabled = true
				p.WhaleProtection.MaxTxPercentage = 500
				p.WhaleProtection.MaxHoldingPercentage = 1000
				p.WhaleProtection.LargeTxThreshold = 100
				p.WhaleProtection.LargeTxCooldown = 3600
				p.Tokenomics.CirculatingSupply = "1000000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				currentTime := time.Now().Unix()
				_ = k.SetCurrentTime(ctx, currentTime)
				_ = k.SetCurrentHeight(ctx, 100)
				// Set last large tx time to 2 hours ago (past cooldown)
				_ = k.SetLastLargeTxTime(ctx, "aura1sender", currentTime-7200)
			},
			sender:    "aura1sender",
			recipient: "aura1recipient",
			amount:    "20000000000000", // 2%
			wantErr:   false,
		},
		{
			name: "recipient exceeds holding limit",
			setupParams: func(p *types.Params) {
				p.WhaleProtection.Enabled = true
				p.WhaleProtection.MaxTxPercentage = 500
				p.WhaleProtection.MaxHoldingPercentage = 500       // 5%
				p.WhaleProtection.LargeTxThreshold = 1000          // 10% - high to avoid cooldown
				p.Tokenomics.CirculatingSupply = "1000000000000000"
			},
			setup: func(k *Keeper, ctx context.Context) {
				_ = k.SetCurrentTime(ctx, time.Now().Unix())
				// Recipient already holds 4% of supply
				_ = k.SetAddressHolding(ctx, "aura1recipient", "40000000000000")
			},
			sender:      "aura1sender",
			recipient:   "aura1recipient",
			amount:      "20000000000000", // 2% - would push recipient to 6%
			wantErr:     true,
			expectedErr: types.ErrWhaleHoldingLimitExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.DefaultParams()
			tt.setupParams(params)

			k, ctx := setupKeeperWithCustomParams(t, params)

			if tt.setup != nil {
				tt.setup(k, ctx)
			}

			err := k.CheckWhaleProtection(ctx, tt.sender, tt.recipient, tt.amount)

			if tt.wantErr {
				require.Error(t, err)
				if tt.expectedErr != nil {
					require.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateAddressHolding(t *testing.T) {
	tests := []struct {
		name       string
		address    string
		newBalance string
		wantErr    bool
	}{
		{
			name:       "valid balance update",
			address:    "aura1user",
			newBalance: "1000000",
			wantErr:    false,
		},
		{
			name:       "zero balance update",
			address:    "aura1user",
			newBalance: "0",
			wantErr:    false,
		},
		{
			name:       "large balance update",
			address:    "aura1whale",
			newBalance: "99999999999999999",
			wantErr:    false,
		},
		{
			name:       "invalid balance format fails",
			address:    "aura1user",
			newBalance: "not_a_number",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, ctx := setupKeeperForTest(t)

			err := k.UpdateAddressHolding(ctx, tt.address, tt.newBalance)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify holding was updated
				holding, err := k.GetAddressHolding(ctx, tt.address)
				require.NoError(t, err)
				require.Equal(t, tt.newBalance, holding)
			}
		})
	}
}

func TestGetLargeTxRecords(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create some large tx records
	for i := 0; i < 15; i++ {
		record := &types.LargeTxRecord{
			TxHash:             "tx_record_" + string(rune('a'+i)),
			Sender:             "aura1sender" + string(rune('a'+i)),
			Recipient:          "aura1recipient",
			Amount:             "1000000000",
			PercentageOfSupply: uint64(50 + i*10),
			BlockHeight:        uint64(i + 1),
			Timestamp:          time.Unix(currentTime-int64(i*60), 0),
			Flagged:            i%2 == 0,
		}
		_ = k.SetLargeTxRecord(ctx, record)
	}

	tests := []struct {
		name     string
		limit    uint64
		expected int
	}{
		{
			name:     "default limit",
			limit:    0,
			expected: 15, // Default is 100, we have 15
		},
		{
			name:     "custom limit",
			limit:    5,
			expected: 5,
		},
		{
			name:     "limit exceeds records",
			limit:    100,
			expected: 15,
		},
		{
			name:     "max limit capped at 1000",
			limit:    2000,
			expected: 15, // Capped to 1000, we have 15
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, err := k.GetLargeTxRecords(ctx, tt.limit)

			require.NoError(t, err)
			require.Len(t, records, tt.expected)
		})
	}
}

func TestGetWhaleProtectionTriggers24h(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create transactions - some within 24h, some outside
	for i := 0; i < 10; i++ {
		record := &types.LargeTxRecord{
			TxHash:             "tx_24h_" + string(rune('a'+i)),
			Sender:             "aura1sender" + string(rune('a'+i)),
			Recipient:          "aura1recipient",
			Amount:             "1000000000",
			PercentageOfSupply: 100,
			BlockHeight:        uint64(i + 1),
			Flagged:            true,
		}

		if i < 5 {
			// Within 24 hours
			record.Timestamp = time.Unix(currentTime-int64(i*3600), 0)
		} else {
			// Outside 24 hours
			record.Timestamp = time.Unix(currentTime-int64(100000+i*3600), 0)
		}
		_ = k.SetLargeTxRecord(ctx, record)
	}

	count, err := k.GetWhaleProtectionTriggers24h(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(5), count)
}

func TestGetWhaleProtectionStatistics(t *testing.T) {
	k, ctx := setupKeeperForTest(t)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)

	// Create transactions with varying percentages
	percentages := []uint64{50, 100, 150, 200, 60}
	for i, pct := range percentages {
		record := &types.LargeTxRecord{
			TxHash:             "tx_stats_" + string(rune('a'+i)),
			Sender:             "aura1sender" + string(rune('a'+i)),
			Recipient:          "aura1recipient",
			Amount:             "1000000000",
			PercentageOfSupply: pct,
			BlockHeight:        uint64(i + 1),
			Timestamp:          time.Unix(currentTime-int64(i*60), 0),
			Flagged:            pct > 50, // Flag if > 0.5%
		}
		_ = k.SetLargeTxRecord(ctx, record)
	}

	total, flagged, avg, err := k.GetWhaleProtectionStatistics(ctx)

	require.NoError(t, err)
	require.Equal(t, uint64(5), total)
	require.Equal(t, uint64(4), flagged) // 4 transactions with pct > 50
	require.Equal(t, uint64(112), avg)   // (50+100+150+200+60)/5 = 112
}

func TestIsWhaleProtectionActive(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "protection enabled",
			enabled:  true,
			expected: true,
		},
		{
			name:     "protection disabled",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.DefaultParams()
			params.WhaleProtection.Enabled = tt.enabled

			k, ctx := setupKeeperWithCustomParams(t, params)

			active := k.IsWhaleProtectionActive(ctx)
			require.Equal(t, tt.expected, active)
		})
	}
}

func TestGetWhaleHoldingPercentage(t *testing.T) {
	tests := []struct {
		name               string
		holding            string
		circulatingSupply  string
		expectedPercentage uint64
	}{
		{
			name:               "1% holding",
			holding:            "1000000000000",
			circulatingSupply:  "100000000000000",
			expectedPercentage: 100, // 1% = 100 basis points
		},
		{
			name:               "5% holding",
			holding:            "5000000000000",
			circulatingSupply:  "100000000000000",
			expectedPercentage: 500,
		},
		{
			name:               "0% holding",
			holding:            "0",
			circulatingSupply:  "100000000000000",
			expectedPercentage: 0,
		},
		{
			name:               "tiny holding",
			holding:            "1000",
			circulatingSupply:  "100000000000000",
			expectedPercentage: 0, // Rounds down
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := types.DefaultParams()
			params.Tokenomics.CirculatingSupply = tt.circulatingSupply

			k, ctx := setupKeeperWithCustomParams(t, params)
			_ = k.SetAddressHolding(ctx, "aura1testaddr", tt.holding)

			pct, err := k.GetWhaleHoldingPercentage(ctx, "aura1testaddr")

			require.NoError(t, err)
			require.Equal(t, tt.expectedPercentage, pct)
		})
	}
}

func TestRecordLargeTx(t *testing.T) {
	params := types.DefaultParams()
	params.Tokenomics.CirculatingSupply = "1000000000000000"

	k, ctx := setupKeeperWithCustomParams(t, params)
	currentTime := time.Now().Unix()
	_ = k.SetCurrentTime(ctx, currentTime)
	_ = k.SetCurrentHeight(ctx, 12345)

	// Parse big ints
	transferAmt := new(big.Int)
	transferAmt.SetString("50000000000000", 10)

	totalSupply := new(big.Int)
	totalSupply.SetString("1000000000000000", 10)

	// Test recording a large transaction
	err := k.recordLargeTx(
		ctx,
		"aura1senderxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		"aura1recipientxxxxxxxxxxxxxxxxxxxxxxxxx",
		"50000000000000", // 5% of supply
		transferAmt,
		totalSupply,
	)

	require.NoError(t, err)

	// Verify record was stored
	records, err := k.GetLargeTxRecords(ctx, 10)
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, "aura1senderxxxxxxxxxxxxxxxxxxxxxxxxxxx", records[0].Sender)
	require.True(t, records[0].Flagged) // Should be flagged as >0.5% of supply
}
