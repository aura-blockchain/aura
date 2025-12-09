package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface for the bank keeper
type BankKeeper interface {
	// MintCoins mints coins to a module account
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error

	// SendCoinsFromModuleToAccount sends coins from a module account to a user account
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	// SendCoinsFromAccountToModule sends coins from a user account to a module account
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	// BurnCoins burns coins from a module account
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error

	// GetBalance retrieves the balance of an account for a specific denomination
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// IRRegistry defines the expected interface for interacting with the inclusionroutines module
type IRRegistry interface {
	// GetIRPrerequisites retrieves the list of prerequisite IR IDs for a given IR
	GetIRPrerequisites(irID string) ([]string, error)

	// IsIRActive checks if an inclusion routine is currently active
	IsIRActive(irID string) bool

	// GetIRScore retrieves the confidence score value for completing a specific IR
	GetIRScore(irID string) (uint64, error)

	// GetIRArena retrieves the arena identifier for a given IR
	GetIRArena(irID string) (string, error)

	// ValidateIRCompletion validates that an IR completion is legitimate
	ValidateIRCompletion(ctx sdk.Context, walletAddr string, irID string, proofData []byte) (bool, error)
}
