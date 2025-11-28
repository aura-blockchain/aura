package testutil

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// TestFixtures provides common test data for all modules
type TestFixtures struct {
	Addresses      []sdk.AccAddress
	ValidatorAddrs []sdk.ValAddress
	Amounts        []sdk.Coin
	Timestamps     []time.Time
}

// NewTestFixtures creates a new set of test fixtures
func NewTestFixtures() *TestFixtures {
	return &TestFixtures{
		Addresses:      GenerateTestAddresses(10),
		ValidatorAddrs: GenerateTestValidatorAddresses(5),
		Amounts:        GenerateTestCoins(),
		Timestamps:     GenerateTestTimestamps(10),
	}
}

// GenerateTestValidatorAddresses generates test validator addresses
func GenerateTestValidatorAddresses(count int) []sdk.ValAddress {
	addrs := make([]sdk.ValAddress, count)
	for i := 0; i < count; i++ {
		addrs[i] = sdk.ValAddress([]byte("test_validator_" + string(rune(i))))
	}
	return addrs
}

// GenerateTestCoins generates test coin amounts
func GenerateTestCoins() []sdk.Coin {
	return []sdk.Coin{
		sdk.NewCoin("aura", math.NewInt(1000000)),
		sdk.NewCoin("aura", math.NewInt(5000000)),
		sdk.NewCoin("aura", math.NewInt(10000000)),
		sdk.NewCoin("stake", math.NewInt(100000)),
		sdk.NewCoin("atom", math.NewInt(50000)),
	}
}

// GenerateTestTimestamps generates test timestamps
func GenerateTestTimestamps(count int) []time.Time {
	base := MockTime()
	timestamps := make([]time.Time, count)
	for i := 0; i < count; i++ {
		timestamps[i] = base.Add(time.Duration(i) * time.Hour)
	}
	return timestamps
}

// TestAccount represents a test account with balance
type TestAccount struct {
	Address sdk.AccAddress
	Balance sdk.Coins
}

// GenerateTestAccounts creates test accounts with balances
func GenerateTestAccounts(count int) []TestAccount {
	accounts := make([]TestAccount, count)
	for i := 0; i < count; i++ {
		accounts[i] = TestAccount{
			Address: sdk.AccAddress([]byte("test_account_" + string(rune(i)))),
			Balance: sdk.NewCoins(sdk.NewCoin("aura", math.NewInt(1000000))),
		}
	}
	return accounts
}

// TestValidator represents a test validator
type TestValidator struct {
	Address    sdk.ValAddress
	Tokens     math.Int
	Commission math.LegacyDec
	Jailed     bool
}

// GenerateTestValidators creates test validators
func GenerateTestValidators(count int) []TestValidator {
	validators := make([]TestValidator, count)
	for i := 0; i < count; i++ {
		validators[i] = TestValidator{
			Address:    sdk.ValAddress([]byte("test_validator_" + string(rune(i)))),
			Tokens:     math.NewInt(1000000),
			Commission: math.LegacyMustNewDecFromStr("0.1"),
			Jailed:     false,
		}
	}
	return validators
}

// CreateTestBankGenesisState creates a bank genesis state for testing
func CreateTestBankGenesisState(accounts []TestAccount) *banktypes.GenesisState {
	balances := make([]banktypes.Balance, len(accounts))
	for i, acc := range accounts {
		balances[i] = banktypes.Balance{
			Address: acc.Address.String(),
			Coins:   acc.Balance,
		}
	}

	return &banktypes.GenesisState{
		Balances: balances,
		Supply:   calculateTotalSupply(balances),
	}
}

func calculateTotalSupply(balances []banktypes.Balance) sdk.Coins {
	total := sdk.NewCoins()
	for _, balance := range balances {
		total = total.Add(balance.Coins...)
	}
	return total
}

// TestData provides common test strings and data
var TestData = struct {
	ValidDID       string
	InvalidDID     string
	ValidIPFSHash  string
	InvalidIPFSHash string
	ValidProof     []byte
	InvalidProof   []byte
	TestMetadata   string
}{
	ValidDID:        "did:aura:test123",
	InvalidDID:      "",
	ValidIPFSHash:   "QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG",
	InvalidIPFSHash: "invalid",
	ValidProof:      []byte("valid_proof_data"),
	InvalidProof:    []byte{},
	TestMetadata:    `{"key":"value"}`,
}
