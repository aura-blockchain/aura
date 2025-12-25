// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "valid address",
			addr:    "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			wantErr: false,
		},
		{
			name:    "empty address",
			addr:    "",
			wantErr: true,
		},
		{
			name:    "whitespace address",
			addr:    "   ",
			wantErr: true,
		},
		{
			name:    "invalid bech32",
			addr:    "invalid-address",
			wantErr: true,
		},
		{
			name:    "valid aura address",
			addr:    "aura1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5y5muy9",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddress(tt.addr)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateAccAddress(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "valid cosmos address",
			addr:    "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			wantErr: false,
		},
		{
			name:    "empty address",
			addr:    "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			addr:    "not-an-address",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAccAddress(tt.addr)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidatePositiveInt(t *testing.T) {
	tests := []struct {
		name      string
		val       sdkmath.Int
		fieldName string
		wantErr   bool
	}{
		{
			name:      "positive value",
			val:       sdkmath.NewInt(100),
			fieldName: "amount",
			wantErr:   false,
		},
		{
			name:      "zero value",
			val:       sdkmath.NewInt(0),
			fieldName: "amount",
			wantErr:   true,
		},
		{
			name:      "negative value",
			val:       sdkmath.NewInt(-50),
			fieldName: "amount",
			wantErr:   true,
		},
		{
			name:      "large positive value",
			val:       sdkmath.NewInt(1000000000000),
			fieldName: "amount",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveInt(tt.val, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateNonNegativeInt(t *testing.T) {
	tests := []struct {
		name      string
		val       sdkmath.Int
		fieldName string
		wantErr   bool
	}{
		{
			name:      "positive value",
			val:       sdkmath.NewInt(100),
			fieldName: "amount",
			wantErr:   false,
		},
		{
			name:      "zero value",
			val:       sdkmath.NewInt(0),
			fieldName: "amount",
			wantErr:   false,
		},
		{
			name:      "negative value",
			val:       sdkmath.NewInt(-1),
			fieldName: "amount",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNonNegativeInt(tt.val, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateBoundedInt(t *testing.T) {
	min := sdkmath.NewInt(10)
	max := sdkmath.NewInt(100)

	tests := []struct {
		name      string
		val       sdkmath.Int
		fieldName string
		wantErr   bool
	}{
		{
			name:      "within bounds",
			val:       sdkmath.NewInt(50),
			fieldName: "amount",
			wantErr:   false,
		},
		{
			name:      "at min bound",
			val:       sdkmath.NewInt(10),
			fieldName: "amount",
			wantErr:   false,
		},
		{
			name:      "at max bound",
			val:       sdkmath.NewInt(100),
			fieldName: "amount",
			wantErr:   false,
		},
		{
			name:      "below min",
			val:       sdkmath.NewInt(5),
			fieldName: "amount",
			wantErr:   true,
		},
		{
			name:      "above max",
			val:       sdkmath.NewInt(150),
			fieldName: "amount",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBoundedInt(tt.val, min, max, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidatePositiveDec(t *testing.T) {
	tests := []struct {
		name      string
		val       sdkmath.LegacyDec
		fieldName string
		wantErr   bool
	}{
		{
			name:      "positive decimal",
			val:       sdkmath.LegacyMustNewDecFromStr("1.5"),
			fieldName: "rate",
			wantErr:   false,
		},
		{
			name:      "zero decimal",
			val:       sdkmath.LegacyNewDec(0),
			fieldName: "rate",
			wantErr:   true,
		},
		{
			name:      "negative decimal",
			val:       sdkmath.LegacyMustNewDecFromStr("-1.5"),
			fieldName: "rate",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveDec(tt.val, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateNonNegativeDec(t *testing.T) {
	tests := []struct {
		name      string
		val       sdkmath.LegacyDec
		fieldName string
		wantErr   bool
	}{
		{
			name:      "positive decimal",
			val:       sdkmath.LegacyMustNewDecFromStr("1.5"),
			fieldName: "rate",
			wantErr:   false,
		},
		{
			name:      "zero decimal",
			val:       sdkmath.LegacyNewDec(0),
			fieldName: "rate",
			wantErr:   false,
		},
		{
			name:      "negative decimal",
			val:       sdkmath.LegacyMustNewDecFromStr("-0.1"),
			fieldName: "rate",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNonNegativeDec(tt.val, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateBoundedDec(t *testing.T) {
	min := sdkmath.LegacyMustNewDecFromStr("0.1")
	max := sdkmath.LegacyMustNewDecFromStr("10.0")

	tests := []struct {
		name      string
		val       sdkmath.LegacyDec
		fieldName string
		wantErr   bool
	}{
		{
			name:      "within bounds",
			val:       sdkmath.LegacyMustNewDecFromStr("5.0"),
			fieldName: "rate",
			wantErr:   false,
		},
		{
			name:      "at min bound",
			val:       sdkmath.LegacyMustNewDecFromStr("0.1"),
			fieldName: "rate",
			wantErr:   false,
		},
		{
			name:      "at max bound",
			val:       sdkmath.LegacyMustNewDecFromStr("10.0"),
			fieldName: "rate",
			wantErr:   false,
		},
		{
			name:      "below min",
			val:       sdkmath.LegacyMustNewDecFromStr("0.05"),
			fieldName: "rate",
			wantErr:   true,
		},
		{
			name:      "above max",
			val:       sdkmath.LegacyMustNewDecFromStr("15.0"),
			fieldName: "rate",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBoundedDec(tt.val, min, max, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateNonEmptyString(t *testing.T) {
	tests := []struct {
		name      string
		str       string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid string",
			str:       "hello",
			fieldName: "name",
			wantErr:   false,
		},
		{
			name:      "empty string",
			str:       "",
			fieldName: "name",
			wantErr:   true,
		},
		{
			name:      "whitespace only",
			str:       "   ",
			fieldName: "name",
			wantErr:   true,
		},
		{
			name:      "string with content and spaces",
			str:       "  hello  ",
			fieldName: "name",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNonEmptyString(tt.str, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateBoundedString(t *testing.T) {
	tests := []struct {
		name      string
		str       string
		minLen    int
		maxLen    int
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid length",
			str:       "hello",
			minLen:    1,
			maxLen:    10,
			fieldName: "name",
			wantErr:   false,
		},
		{
			name:      "too short",
			str:       "hi",
			minLen:    5,
			maxLen:    10,
			fieldName: "name",
			wantErr:   true,
		},
		{
			name:      "too long",
			str:       "this is a very long string",
			minLen:    1,
			maxLen:    10,
			fieldName: "name",
			wantErr:   true,
		},
		{
			name:      "at min length",
			str:       "hello",
			minLen:    5,
			maxLen:    10,
			fieldName: "name",
			wantErr:   false,
		},
		{
			name:      "at max length",
			str:       "helloworld",
			minLen:    5,
			maxLen:    10,
			fieldName: "name",
			wantErr:   false,
		},
		{
			name:      "with control characters",
			str:       "hello\x00world",
			minLen:    1,
			maxLen:    20,
			fieldName: "name",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBoundedString(tt.str, tt.minLen, tt.maxLen, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid http URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			url:     "https://example.com/path",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "no scheme",
			url:     "example.com",
			wantErr: true,
		},
		{
			name:    "invalid scheme",
			url:     "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "no host",
			url:     "http://",
			wantErr: true,
		},
		{
			name:    "with query params",
			url:     "https://example.com/path?param=value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateHash(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{
			name:    "valid SHA256 hash",
			hash:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr: false,
		},
		{
			name:    "valid SHA512 hash",
			hash:    "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
			wantErr: false,
		},
		{
			name:    "empty hash",
			hash:    "",
			wantErr: true,
		},
		{
			name:    "too short",
			hash:    "abc123",
			wantErr: true,
		},
		{
			name:    "too long",
			hash:    strings.Repeat("a", 150),
			wantErr: true,
		},
		{
			name:    "non-hex characters",
			hash:    "g3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantErr: true,
		},
		{
			name:    "uppercase hex valid",
			hash:    "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHash(tt.hash)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateDenom(t *testing.T) {
	tests := []struct {
		name    string
		denom   string
		wantErr bool
	}{
		{
			name:    "valid simple denom",
			denom:   "uatom",
			wantErr: false,
		},
		{
			name:    "valid with dots",
			denom:   "ibc.token",
			wantErr: false,
		},
		{
			name:    "valid with slashes",
			denom:   "factory/addr/token",
			wantErr: false,
		},
		{
			name:    "empty denom",
			denom:   "",
			wantErr: true,
		},
		{
			name:    "starts with number",
			denom:   "1atom",
			wantErr: true,
		},
		{
			name:    "too short",
			denom:   "ab",
			wantErr: true,
		},
		{
			name:    "valid uppercase",
			denom:   "ATOM",
			wantErr: false,
		},
		{
			name:    "invalid characters",
			denom:   "atom@token",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDenom(tt.denom)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateCoin(t *testing.T) {
	tests := []struct {
		name      string
		coin      sdk.Coin
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid coin",
			coin:      sdk.NewInt64Coin("uatom", 1000),
			fieldName: "amount",
			wantErr:   false,
		},
		{
			name:      "zero amount",
			coin:      sdk.NewInt64Coin("uatom", 0),
			fieldName: "amount",
			wantErr:   true,
		},
		{
			name:      "negative amount",
			coin:      sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(-100)},
			fieldName: "amount",
			wantErr:   true,
		},
		{
			name:      "invalid denom",
			coin:      sdk.Coin{Denom: "1invalid", Amount: sdkmath.NewInt(100)},
			fieldName: "amount",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCoin(tt.coin, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateCoins(t *testing.T) {
	tests := []struct {
		name      string
		coins     sdk.Coins
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid single coin",
			coins:     sdk.NewCoins(sdk.NewInt64Coin("uatom", 1000)),
			fieldName: "amounts",
			wantErr:   false,
		},
		{
			name:      "valid multiple coins",
			coins:     sdk.NewCoins(sdk.NewInt64Coin("uatom", 1000), sdk.NewInt64Coin("stake", 500)),
			fieldName: "amounts",
			wantErr:   false,
		},
		{
			name:      "empty coins",
			coins:     sdk.Coins{},
			fieldName: "amounts",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCoins(tt.coins, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateAlphanumeric(t *testing.T) {
	tests := []struct {
		name      string
		str       string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid alphanumeric",
			str:       "abc123",
			fieldName: "id",
			wantErr:   false,
		},
		{
			name:      "with underscores",
			str:       "user_123",
			fieldName: "id",
			wantErr:   false,
		},
		{
			name:      "with hyphens",
			str:       "user-123",
			fieldName: "id",
			wantErr:   false,
		},
		{
			name:      "with spaces",
			str:       "user 123",
			fieldName: "id",
			wantErr:   true,
		},
		{
			name:      "with special chars",
			str:       "user@123",
			fieldName: "id",
			wantErr:   true,
		},
		{
			name:      "empty",
			str:       "",
			fieldName: "id",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAlphanumeric(tt.str, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid ID",
			id:        "transfer-123",
			fieldName: "transfer_id",
			wantErr:   false,
		},
		{
			name:      "empty ID",
			id:        "",
			fieldName: "transfer_id",
			wantErr:   true,
		},
		{
			name:      "too long ID",
			id:        strings.Repeat("a", 150),
			fieldName: "transfer_id",
			wantErr:   true,
		},
		{
			name:      "with special characters",
			id:        "transfer@123",
			fieldName: "transfer_id",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateJurisdictionCode(t *testing.T) {
	valid := []string{"US", "GB", "JP", "US-NY", "DE-BW"}
	for _, code := range valid {
		require.NoError(t, ValidateJurisdictionCode(code), "expected valid code %s", code)
	}

	invalid := []string{"", "usa", "U", "USNYC", "1A", "US_", "US--NY", "US-N"}
	for _, code := range invalid {
		require.Error(t, ValidateJurisdictionCode(code), "expected invalid code %s", code)
	}
}

func TestValidateUint32Positive(t *testing.T) {
	tests := []struct {
		name      string
		val       uint32
		fieldName string
		wantErr   bool
	}{
		{
			name:      "positive value",
			val:       10,
			fieldName: "threshold",
			wantErr:   false,
		},
		{
			name:      "zero value",
			val:       0,
			fieldName: "threshold",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUint32Positive(tt.val, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateUint64Positive(t *testing.T) {
	tests := []struct {
		name      string
		val       uint64
		fieldName string
		wantErr   bool
	}{
		{
			name:      "positive value",
			val:       100,
			fieldName: "duration",
			wantErr:   false,
		},
		{
			name:      "zero value",
			val:       0,
			fieldName: "duration",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUint64Positive(tt.val, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateBoundedUint32(t *testing.T) {
	tests := []struct {
		name      string
		val       uint32
		min       uint32
		max       uint32
		fieldName string
		wantErr   bool
	}{
		{
			name:      "within bounds",
			val:       50,
			min:       10,
			max:       100,
			fieldName: "threshold",
			wantErr:   false,
		},
		{
			name:      "at min",
			val:       10,
			min:       10,
			max:       100,
			fieldName: "threshold",
			wantErr:   false,
		},
		{
			name:      "at max",
			val:       100,
			min:       10,
			max:       100,
			fieldName: "threshold",
			wantErr:   false,
		},
		{
			name:      "below min",
			val:       5,
			min:       10,
			max:       100,
			fieldName: "threshold",
			wantErr:   true,
		},
		{
			name:      "above max",
			val:       150,
			min:       10,
			max:       100,
			fieldName: "threshold",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBoundedUint32(tt.val, tt.min, tt.max, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateBoundedUint64(t *testing.T) {
	tests := []struct {
		name      string
		val       uint64
		min       uint64
		max       uint64
		fieldName string
		wantErr   bool
	}{
		{
			name:      "within bounds",
			val:       500,
			min:       100,
			max:       1000,
			fieldName: "duration",
			wantErr:   false,
		},
		{
			name:      "below min",
			val:       50,
			min:       100,
			max:       1000,
			fieldName: "duration",
			wantErr:   true,
		},
		{
			name:      "above max",
			val:       1500,
			min:       100,
			max:       1000,
			fieldName: "duration",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBoundedUint64(tt.val, tt.min, tt.max, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateChainID(t *testing.T) {
	tests := []struct {
		name    string
		chainID string
		wantErr bool
	}{
		{
			name:    "valid simple chain ID",
			chainID: "paw",
			wantErr: false,
		},
		{
			name:    "valid chain ID with number",
			chainID: "osmosis-1",
			wantErr: false,
		},
		{
			name:    "valid chain ID",
			chainID: "cosmoshub-4",
			wantErr: false,
		},
		{
			name:    "empty chain ID",
			chainID: "",
			wantErr: true,
		},
		{
			name:    "too short",
			chainID: "a",
			wantErr: true,
		},
		{
			name:    "uppercase",
			chainID: "PAW",
			wantErr: true,
		},
		{
			name:    "special characters",
			chainID: "chain_id",
			wantErr: true,
		},
		{
			name:    "starts with hyphen",
			chainID: "-chain",
			wantErr: true,
		},
		{
			name:    "ends with hyphen",
			chainID: "chain-",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateChainID(tt.chainID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove leading/trailing spaces",
			input:    "  hello  ",
			expected: "hello",
		},
		{
			name:     "convert tabs to spaces",
			input:    "hello\tworld",
			expected: "hello world",
		},
		{
			name:     "convert newlines to spaces",
			input:    "hello\nworld",
			expected: "hello world",
		},
		{
			name:     "remove control characters",
			input:    "hello\x00world",
			expected: "helloworld",
		},
		{
			name:     "collapse multiple spaces",
			input:    "hello     world",
			expected: "hello world",
		},
		{
			name:     "complex sanitization",
			input:    "  hello\t\n  world  \x00  ",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestValidatePercentage(t *testing.T) {
	tests := []struct {
		name      string
		val       uint32
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid 0%",
			val:       0,
			fieldName: "slippage",
			wantErr:   false,
		},
		{
			name:      "valid 50%",
			val:       50,
			fieldName: "slippage",
			wantErr:   false,
		},
		{
			name:      "valid 100%",
			val:       100,
			fieldName: "slippage",
			wantErr:   false,
		},
		{
			name:      "invalid > 100%",
			val:       101,
			fieldName: "slippage",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePercentage(tt.val, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateBasisPoints(t *testing.T) {
	tests := []struct {
		name      string
		val       uint64
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid 0 bps",
			val:       0,
			fieldName: "slippage_bps",
			wantErr:   false,
		},
		{
			name:      "valid 5% (500 bps)",
			val:       500,
			fieldName: "slippage_bps",
			wantErr:   false,
		},
		{
			name:      "valid 100% (10000 bps)",
			val:       10000,
			fieldName: "slippage_bps",
			wantErr:   false,
		},
		{
			name:      "invalid > 100%",
			val:       10001,
			fieldName: "slippage_bps",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBasisPoints(tt.val, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		ts        int64
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid zero",
			ts:        0,
			fieldName: "timestamp",
			wantErr:   false,
		},
		{
			name:      "valid positive",
			ts:        1234567890,
			fieldName: "timestamp",
			wantErr:   false,
		},
		{
			name:      "invalid negative",
			ts:        -1,
			fieldName: "timestamp",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimestamp(tt.ts, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidatePositiveTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		ts        int64
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid positive",
			ts:        1234567890,
			fieldName: "timestamp",
			wantErr:   false,
		},
		{
			name:      "invalid zero",
			ts:        0,
			fieldName: "timestamp",
			wantErr:   true,
		},
		{
			name:      "invalid negative",
			ts:        -1,
			fieldName: "timestamp",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveTimestamp(tt.ts, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateBytes(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		minLen    int
		maxLen    int
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid length",
			data:      []byte("hello"),
			minLen:    1,
			maxLen:    10,
			fieldName: "signature",
			wantErr:   false,
		},
		{
			name:      "too short",
			data:      []byte("hi"),
			minLen:    5,
			maxLen:    10,
			fieldName: "signature",
			wantErr:   true,
		},
		{
			name:      "too long",
			data:      []byte("this is a very long byte array"),
			minLen:    1,
			maxLen:    10,
			fieldName: "signature",
			wantErr:   true,
		},
		{
			name:      "at min",
			data:      []byte("hello"),
			minLen:    5,
			maxLen:    10,
			fieldName: "signature",
			wantErr:   false,
		},
		{
			name:      "at max",
			data:      []byte("helloworld"),
			minLen:    5,
			maxLen:    10,
			fieldName: "signature",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBytes(tt.data, tt.minLen, tt.maxLen, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateStringSlice(t *testing.T) {
	tests := []struct {
		name      string
		slice     []string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid slice",
			slice:     []string{"addr1", "addr2", "addr3"},
			fieldName: "signers",
			wantErr:   false,
		},
		{
			name:      "empty slice",
			slice:     []string{},
			fieldName: "signers",
			wantErr:   true,
		},
		{
			name:      "slice with empty string",
			slice:     []string{"addr1", "", "addr3"},
			fieldName: "signers",
			wantErr:   true,
		},
		{
			name:      "slice with whitespace",
			slice:     []string{"addr1", "   ", "addr3"},
			fieldName: "signers",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringSlice(tt.slice, tt.fieldName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestContainsControlCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "normal string",
			input:    "hello world",
			expected: false,
		},
		{
			name:     "with newline (allowed)",
			input:    "hello\nworld",
			expected: false,
		},
		{
			name:     "with tab (allowed)",
			input:    "hello\tworld",
			expected: false,
		},
		{
			name:     "with null byte",
			input:    "hello\x00world",
			expected: true,
		},
		{
			name:     "with bell character",
			input:    "hello\x07world",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsControlCharacters(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}
