package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BankKeeper defines the expected interface needed for token transfers
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx sdk.Context, acc sdk.AccountI)
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// VCRegistryKeeper defines the expected interface for vcregistry keeper
// Used for shared identity verification
type VCRegistryKeeper interface {
	GetIRScore(ctx sdk.Context, address string) uint64
	IsVerified(ctx sdk.Context, address string) bool
}

// StakingKeeper defines the expected interface for the staking module keeper.
// Used for slashing malicious bridge validators.
//
// SECURITY CRITICAL: Slashing is the primary economic deterrent against
// validator fraud in the bridge. This interface allows the bridge module
// to slash validator stake for:
//   - Signing fraudulent transfers
//   - Double-signing (signing conflicting messages)
//   - Being offline (failing liveness checks)
type StakingKeeper interface {
	// GetValidator retrieves a validator by operator address
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)

	// Slash reduces a validator's stake by a fraction
	// Parameters:
	//   - consAddr: Consensus address of the validator
	//   - infractionHeight: Block height where the infraction occurred
	//   - power: Validator's power at the infraction height
	//   - slashFactor: Fraction of stake to slash (0.0 to 1.0)
	Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor sdkmath.LegacyDec) sdkmath.Int

	// Jail prevents a validator from participating in consensus
	// Parameters:
	//   - consAddr: Consensus address of the validator to jail
	Jail(ctx sdk.Context, consAddr sdk.ConsAddress)

	// Unjail allows a jailed validator to return to the active set
	// Parameters:
	//   - valAddr: Validator operator address
	Unjail(ctx sdk.Context, valAddr sdk.ValAddress)

	// IsBonded returns true if the validator is bonded (in active set)
	IsBonded(ctx sdk.Context, addr sdk.ValAddress) bool

	// GetValidatorByConsAddr retrieves a validator by consensus address
	GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (validator stakingtypes.Validator, found bool)
}
