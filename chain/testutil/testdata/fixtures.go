// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package testdata

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
)

// Test addresses - predefined for consistent testing
var (
	TestAddr1 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	TestAddr2 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	TestAddr3 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	TestAddr4 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	TestAddr5 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
)

// Test validator addresses
var (
	TestValAddr1 = sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address())
	TestValAddr2 = sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address())
	TestValAddr3 = sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address())
	TestValAddr4 = sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address())
)

// Test amounts - common amounts for testing
var (
	TestAmount1     = math.NewInt(1)
	TestAmount10    = math.NewInt(10)
	TestAmount100   = math.NewInt(100)
	TestAmount1000  = math.NewInt(1000)
	TestAmount10000 = math.NewInt(10000)
	TestAmount1M    = math.NewInt(1000000)
	TestAmount1B    = math.NewInt(1000000000)
)

// Test decimal amounts
var (
	TestDec0     = math.LegacyNewDec(0)
	TestDec1     = math.LegacyNewDec(1)
	TestDec10    = math.LegacyNewDec(10)
	TestDec100   = math.LegacyNewDec(100)
	TestDecHalf  = math.LegacyNewDecWithPrec(5, 1)  // 0.5
	TestDecQuarter = math.LegacyNewDecWithPrec(25, 2) // 0.25
)

// Test coins - predefined coin sets
var (
	TestCoinAura100  = sdk.NewCoin("uaura", TestAmount100)
	TestCoinAura1000 = sdk.NewCoin("uaura", TestAmount1000)
	TestCoinAura1M   = sdk.NewCoin("uaura", TestAmount1M)

	TestCoinUSDT100  = sdk.NewCoin("uusdt", TestAmount100)
	TestCoinUSDT1000 = sdk.NewCoin("uusdt", TestAmount1000)
	TestCoinUSDT1M   = sdk.NewCoin("uusdt", TestAmount1M)

	TestCoinStake100  = sdk.NewCoin("stake", TestAmount100)
	TestCoinStake1000 = sdk.NewCoin("stake", TestAmount1000)
	TestCoinStake1M   = sdk.NewCoin("stake", TestAmount1M)
)

// Test coin sets
var (
	TestCoinsAura    = sdk.NewCoins(TestCoinAura1000)
	TestCoinsUSDT    = sdk.NewCoins(TestCoinUSDT1000)
	TestCoinsMixed   = sdk.NewCoins(TestCoinAura1000, TestCoinUSDT1000)
	TestCoinsStake   = sdk.NewCoins(TestCoinStake1000)
	TestCoinsEmpty   = sdk.NewCoins()
	TestCoinsLarge   = sdk.NewCoins(TestCoinAura1M, TestCoinUSDT1M)
)

// Test timestamps
var (
	TestTime1 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	TestTime2 = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	TestTime3 = time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	TestTimeGenesis = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
)

// Test durations
var (
	TestDuration1Hour  = time.Hour
	TestDuration1Day   = 24 * time.Hour
	TestDuration1Week  = 7 * 24 * time.Hour
	TestDuration1Month = 30 * 24 * time.Hour
	TestDuration1Year  = 365 * 24 * time.Hour
)

// Test strings
var (
	TestChainID       = "aura-testnet-1"
	TestMemo          = "test memo"
	TestDescription   = "test description"
	TestMoniker       = "test-validator"
	TestWebsite       = "https://test.aura.network"
	TestSecurityContact = "security@test.aura.network"
	TestDetails       = "Test validator details"
)

// Test IDs
var (
	TestProposalID1 = uint64(1)
	TestProposalID2 = uint64(2)
	TestProposalID3 = uint64(3)
)

// Helper functions

// GenTestAddr generates a random test address
func GenTestAddr() sdk.AccAddress {
	return sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

// GenTestValidatorAddr generates a random validator address
func GenTestValidatorAddr() sdk.ValAddress {
	return sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address())
}

// GenTestAddrs generates n test addresses
func GenTestAddrs(n int) []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, n)
	for i := 0; i < n; i++ {
		addrs[i] = GenTestAddr()
	}
	return addrs
}

// GenTestValAddrs generates n validator addresses
func GenTestValAddrs(n int) []sdk.ValAddress {
	addrs := make([]sdk.ValAddress, n)
	for i := 0; i < n; i++ {
		addrs[i] = GenTestValidatorAddr()
	}
	return addrs
}

// MakeTestCoins creates test coins with given denom and amount
func MakeTestCoins(denom string, amount int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(amount)))
}

// MakeTestCoin creates a single test coin
func MakeTestCoin(denom string, amount int64) sdk.Coin {
	return sdk.NewCoin(denom, math.NewInt(amount))
}

// MakeTestMultiCoins creates multiple test coins
func MakeTestMultiCoins(denoms []string, amounts []int64) sdk.Coins {
	if len(denoms) != len(amounts) {
		panic("denoms and amounts length mismatch")
	}

	coins := sdk.NewCoins()
	for i, denom := range denoms {
		coins = coins.Add(sdk.NewCoin(denom, math.NewInt(amounts[i])))
	}
	return coins
}

// TimeFromNow returns a time offset from now
func TimeFromNow(d time.Duration) time.Time {
	return time.Now().UTC().Add(d)
}

// TimeAgo returns a time in the past
func TimeAgo(d time.Duration) time.Time {
	return time.Now().UTC().Add(-d)
}

// IsTestAddr checks if address is one of the predefined test addresses
func IsTestAddr(addr sdk.AccAddress) bool {
	return addr.Equals(TestAddr1) ||
	       addr.Equals(TestAddr2) ||
	       addr.Equals(TestAddr3) ||
	       addr.Equals(TestAddr4) ||
	       addr.Equals(TestAddr5)
}

// Constants for testing
const (
	DefaultGasLimit = uint64(200000)
	DefaultGasPrice = uint64(1)
	MaxGasLimit     = uint64(10000000)

	// Common test parameters
	TestBondDenom   = "uaura"
	TestStakeDenom  = "stake"
	TestUSDTDenom   = "uusdt"

	// Validation constants
	MinValidatorCount = 4
	MaxValidatorCount = 100
	MinDelegation     = 1000000 // 1 AURA

	// Governance constants
	TestVotingPeriod  = 24 * time.Hour
	TestDepositPeriod = 24 * time.Hour
	MinDeposit        = 10000000 // 10 AURA
)

// Default test parameters
var (
	DefaultTestCoins = sdk.NewCoins(
		sdk.NewCoin(TestBondDenom, TestAmount1M),
		sdk.NewCoin(TestUSDTDenom, TestAmount1M),
	)
)
