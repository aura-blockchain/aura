package testutil

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// MockBankKeeper is a mock implementation of the bank keeper
type MockBankKeeper struct {
	Balances map[string]sdk.Coins
}

// NewMockBankKeeper creates a new mock bank keeper
func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		Balances: make(map[string]sdk.Coins),
	}
}

// GetBalance returns the balance for an address and denom
func (m *MockBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if coins, ok := m.Balances[addr.String()]; ok {
		return sdk.NewCoin(denom, coins.AmountOf(denom))
	}
	return sdk.NewCoin(denom, math.ZeroInt())
}

// SpendableCoins returns spendable coins for an address
func (m *MockBankKeeper) SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	if coins, ok := m.Balances[addr.String()]; ok {
		return coins
	}
	return sdk.NewCoins()
}

// SendCoinsFromAccountToModule sends coins from account to module
func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	senderBalance := m.Balances[senderAddr.String()]
	if !senderBalance.IsAllGTE(amt) {
		return fmt.Errorf("insufficient funds")
	}
	m.Balances[senderAddr.String()] = senderBalance.Sub(amt...)
	return nil
}

// SendCoinsFromModuleToAccount sends coins from module to account
func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	current := m.Balances[recipientAddr.String()]
	m.Balances[recipientAddr.String()] = current.Add(amt...)
	return nil
}

// SendCoins is a passthrough for completeness
func (m *MockBankKeeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	currentFrom := m.Balances[fromAddr.String()]
	if !currentFrom.IsAllGTE(amt) {
		return fmt.Errorf("insufficient funds")
	}
	m.Balances[fromAddr.String()] = currentFrom.Sub(amt...)
	currentTo := m.Balances[toAddr.String()]
	m.Balances[toAddr.String()] = currentTo.Add(amt...)
	return nil
}

func (m *MockBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	// Track minting against the module name for completeness
	m.Balances[moduleName] = m.Balances[moduleName].Add(amt...)
	return nil
}

func (m *MockBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	current := m.Balances[moduleName]
	if !current.IsAllGTE(amt) {
		return fmt.Errorf("insufficient module balance")
	}
	m.Balances[moduleName] = current.Sub(amt...)
	return nil
}

// MockAccountKeeper is a mock implementation of the account keeper
type MockAccountKeeper struct {
	Accounts map[string]sdk.AccountI
}

// NewMockAccountKeeper creates a new mock account keeper
func NewMockAccountKeeper() *MockAccountKeeper {
	return &MockAccountKeeper{
		Accounts: make(map[string]sdk.AccountI),
	}
}

// GetAccount returns an account
func (m *MockAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return m.Accounts[addr.String()]
}

// SetAccount sets an account
func (m *MockAccountKeeper) SetAccount(ctx sdk.Context, acc sdk.AccountI) {
	m.Accounts[acc.GetAddress().String()] = acc
}

// NewAccountWithAddress creates a basic account for an address
func (m *MockAccountKeeper) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	acc := authtypes.NewBaseAccountWithAddress(addr)
	m.SetAccount(ctx, acc)
	return acc
}

// MockStakingKeeper is a mock implementation of the staking keeper
type MockStakingKeeper struct {
	Validators map[string]interface{}
}

// NewMockStakingKeeper creates a new mock staking keeper
func NewMockStakingKeeper() *MockStakingKeeper {
	return &MockStakingKeeper{
		Validators: make(map[string]interface{}),
	}
}

// GetValidator returns a validator
func (m *MockStakingKeeper) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (interface{}, error) {
	if val, ok := m.Validators[addr.String()]; ok {
		return val, nil
	}
	return nil, nil
}

// MockSlashingKeeper is a mock implementation of the slashing keeper
type MockSlashingKeeper struct {
	SlashedValidators map[string]bool
}

// NewMockSlashingKeeper creates a new mock slashing keeper
func NewMockSlashingKeeper() *MockSlashingKeeper {
	return &MockSlashingKeeper{
		SlashedValidators: make(map[string]bool),
	}
}

// Slash marks a validator as slashed
func (m *MockSlashingKeeper) Slash(ctx sdk.Context, addr sdk.ValAddress) error {
	m.SlashedValidators[addr.String()] = true
	return nil
}

// MockGovernanceKeeper is a mock implementation of the governance keeper
type MockGovernanceKeeper struct {
	Proposals map[uint64]interface{}
}

// NewMockGovernanceKeeper creates a new mock governance keeper
func NewMockGovernanceKeeper() *MockGovernanceKeeper {
	return &MockGovernanceKeeper{
		Proposals: make(map[uint64]interface{}),
	}
}

// GetProposal returns a proposal
func (m *MockGovernanceKeeper) GetProposal(ctx sdk.Context, proposalID uint64) (interface{}, bool) {
	prop, ok := m.Proposals[proposalID]
	return prop, ok
}

// SetProposal sets a proposal
func (m *MockGovernanceKeeper) SetProposal(ctx sdk.Context, proposalID uint64, proposal interface{}) {
	m.Proposals[proposalID] = proposal
}
