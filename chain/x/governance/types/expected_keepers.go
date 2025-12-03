package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected bank keeper interface for managing token transfers
type BankKeeper interface {
	// SendCoinsFromAccountToModule transfers coins from an account to a module account
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error

	// SendCoinsFromModuleToAccount transfers coins from a module account to an account
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	// GetBalance returns the balance of a specific denom for an account
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}
