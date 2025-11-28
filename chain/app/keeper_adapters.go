package app

import (
	"context"
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdkmath "cosmossdk.io/math"

	securitykeeper "github.com/aequitas/aura/chain/x/security/keeper"
	validatorsecuritykeeper "github.com/aequitas/aura/chain/x/validatorsecurity/keeper"
	validatorsecuritytypes "github.com/aequitas/aura/chain/x/validatorsecurity/types"
	vckeeper "github.com/aequitas/aura/chain/x/vcregistry/keeper"
)

// bankKeeperAdapter exposes sdk.Context-based helpers expected by legacy modules while delegating to the
// context.Context-driven bank keeper from Cosmos SDK v0.53.
type bankKeeperAdapter struct {
	inner bankkeeper.BaseKeeper
}

func newBankKeeperAdapter(inner bankkeeper.BaseKeeper) bankKeeperAdapter {
	return bankKeeperAdapter{inner: inner}
}

func (a bankKeeperAdapter) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	return a.inner.SendCoins(sdk.WrapSDKContext(ctx), fromAddr, toAddr, amt)
}

func (a bankKeeperAdapter) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return a.inner.SendCoinsFromAccountToModule(sdk.WrapSDKContext(ctx), senderAddr, recipientModule, amt)
}

func (a bankKeeperAdapter) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return a.inner.SendCoinsFromModuleToAccount(sdk.WrapSDKContext(ctx), senderModule, recipientAddr, amt)
}

func (a bankKeeperAdapter) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	return a.inner.SendCoinsFromModuleToModule(sdk.WrapSDKContext(ctx), senderModule, recipientModule, amt)
}

func (a bankKeeperAdapter) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return a.inner.MintCoins(sdk.WrapSDKContext(ctx), moduleName, amt)
}

func (a bankKeeperAdapter) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	return a.inner.BurnCoins(sdk.WrapSDKContext(ctx), moduleName, amt)
}

func (a bankKeeperAdapter) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return a.inner.GetBalance(sdk.WrapSDKContext(ctx), addr, denom)
}

// validatorSecurityBankAdapter exposes context.Context based methods for validatorsecurity.
type validatorSecurityBankAdapter struct {
	inner bankkeeper.BaseKeeper
}

func newValidatorSecurityBankAdapter(inner bankkeeper.BaseKeeper) validatorsecuritykeeper.BankKeeper {
	return validatorSecurityBankAdapter{inner: inner}
}

func (a validatorSecurityBankAdapter) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return a.inner.GetBalance(sdk.UnwrapSDKContext(ctx), addr, denom)
}

func (a validatorSecurityBankAdapter) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return a.inner.SendCoinsFromAccountToModule(sdk.UnwrapSDKContext(ctx), senderAddr, recipientModule, amt)
}

func (a validatorSecurityBankAdapter) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return a.inner.SendCoinsFromModuleToAccount(sdk.UnwrapSDKContext(ctx), senderModule, recipientAddr, amt)
}

// accountKeeperAdapter bridges sdk.Context expectations to the context-aware auth keeper.
type accountKeeperAdapter struct {
	inner authkeeper.AccountKeeper
}

func newAccountKeeperAdapter(inner authkeeper.AccountKeeper) accountKeeperAdapter {
	return accountKeeperAdapter{inner: inner}
}

func (a accountKeeperAdapter) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return a.inner.GetAccount(sdk.WrapSDKContext(ctx), addr)
}

func (a accountKeeperAdapter) SetAccount(ctx sdk.Context, acc sdk.AccountI) {
	a.inner.SetAccount(sdk.WrapSDKContext(ctx), acc)
}

func (a accountKeeperAdapter) NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return a.inner.NewAccountWithAddress(sdk.WrapSDKContext(ctx), addr)
}

// vcRegistryKeeperAdapter reuses the confidence score keeper hooks to provide the interface DEX/bridge expect.
type vcRegistryKeeperAdapter struct {
	inner *vckeeper.Keeper
}

func newVCRegistryKeeperAdapter(inner *vckeeper.Keeper) vcRegistryKeeperAdapter {
	return vcRegistryKeeperAdapter{inner: inner}
}

// validatorSecurityStakingAdapter wraps the staking keeper for validatorsecurity expectations.
type validatorSecurityStakingAdapter struct {
	inner *stakingkeeper.Keeper
}

func newValidatorSecurityStakingAdapter(inner *stakingkeeper.Keeper) validatorsecuritykeeper.StakingKeeper {
	return validatorSecurityStakingAdapter{inner: inner}
}

func (a validatorSecurityStakingAdapter) Validator(ctx context.Context, addr sdk.ValAddress) (validatorsecuritykeeper.Validator, error) {
	validator, err := a.inner.GetValidator(ctx, addr)
	if err != nil {
		if errors.Is(err, stakingtypes.ErrNoValidatorFound) {
			return nil, validatorsecuritytypes.ErrValidatorNotFound
		}
		return nil, err
	}
	return validatorsecurityValidatorAdapter{inner: validator}, nil
}

func (a validatorSecurityStakingAdapter) ValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (validatorsecuritykeeper.Validator, error) {
	validator, err := a.inner.GetValidatorByConsAddr(ctx, consAddr)
	if err != nil {
		if errors.Is(err, stakingtypes.ErrNoValidatorFound) {
			return nil, validatorsecuritytypes.ErrValidatorNotFound
		}
		return nil, err
	}
	return validatorsecurityValidatorAdapter{inner: validator}, nil
}

func (a validatorSecurityStakingAdapter) Slash(ctx context.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor sdkmath.LegacyDec) (sdkmath.Int, error) {
	return a.inner.Slash(ctx, consAddr, infractionHeight, power, slashFactor)
}

func (a validatorSecurityStakingAdapter) Jail(ctx context.Context, consAddr sdk.ConsAddress) error {
	return a.inner.Jail(ctx, consAddr)
}

func (a validatorSecurityStakingAdapter) Unjail(ctx context.Context, consAddr sdk.ConsAddress) error {
	return a.inner.Unjail(ctx, consAddr)
}

func (a validatorSecurityStakingAdapter) GetAllValidators(ctx context.Context) ([]validatorsecuritykeeper.Validator, error) {
	validators, err := a.inner.GetAllValidators(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]validatorsecuritykeeper.Validator, 0, len(validators))
	for _, val := range validators {
		result = append(result, validatorsecurityValidatorAdapter{inner: val})
	}
	return result, nil
}

func (a validatorSecurityStakingAdapter) PowerReduction(ctx context.Context) sdkmath.Int {
	return a.inner.PowerReduction(ctx)
}

type validatorsecurityValidatorAdapter struct {
	inner stakingtypes.Validator
}

func (v validatorsecurityValidatorAdapter) GetOperator() sdk.ValAddress {
	return sdk.ValAddress(v.inner.GetOperator())
}

func (v validatorsecurityValidatorAdapter) GetConsPubKey() (interface{}, error) {
	return v.inner.ConsPubKey()
}

func (v validatorsecurityValidatorAdapter) GetConsAddr() (sdk.ConsAddress, error) {
	addr, err := v.inner.GetConsAddr()
	if err != nil {
		return nil, err
	}
	return sdk.ConsAddress(addr), nil
}

func (v validatorsecurityValidatorAdapter) GetStatus() int32 {
	return int32(v.inner.GetStatus())
}

func (v validatorsecurityValidatorAdapter) GetTokens() sdkmath.Int {
	return v.inner.GetTokens()
}

// validatorSecuritySlashingAdapter maps the slashing keeper to the expected interface.
type validatorSecuritySlashingAdapter struct {
	inner *slashingkeeper.Keeper
}

func newValidatorSecuritySlashingAdapter(inner *slashingkeeper.Keeper) validatorsecuritykeeper.SlashingKeeper {
	return validatorSecuritySlashingAdapter{inner: inner}
}

func (a validatorSecuritySlashingAdapter) IsTombstoned(ctx context.Context, consAddr sdk.ConsAddress) bool {
	return a.inner.IsTombstoned(ctx, consAddr)
}

func (a validatorSecuritySlashingAdapter) Tombstone(ctx context.Context, consAddr sdk.ConsAddress) error {
	return a.inner.Tombstone(ctx, consAddr)
}

func (a validatorSecuritySlashingAdapter) JailUntil(ctx context.Context, consAddr sdk.ConsAddress, jailTime time.Time) error {
	return a.inner.JailUntil(ctx, consAddr, jailTime)
}

func (a vcRegistryKeeperAdapter) GetIRScore(ctx sdk.Context, address string) uint64 {
	cs := a.inner.GetConfidenceScoreKeeper()
	if cs == nil {
		return 0
	}
	score, ok := cs.GetUserScore(address)
	if !ok {
		return 0
	}
	return score
}

func (a vcRegistryKeeperAdapter) IsVerified(ctx sdk.Context, address string) bool {
	cs := a.inner.GetConfidenceScoreKeeper()
	if cs == nil {
		return false
	}
	return cs.IsVerified(address)
}

// ============================================================================
// SECURITY KEEPER ADAPTERS
// ============================================================================

// securityBankKeeperAdapter wraps the bank keeper for the security module
type securityBankKeeperAdapter struct {
	inner bankkeeper.BaseKeeper
}

func newSecurityBankKeeperAdapter(inner bankkeeper.BaseKeeper) securitykeeper.BankKeeper {
	return securityBankKeeperAdapter{inner: inner}
}

func (a securityBankKeeperAdapter) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return a.inner.GetBalance(sdk.WrapSDKContext(ctx), addr, denom)
}

func (a securityBankKeeperAdapter) SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error {
	return a.inner.SendCoins(sdk.WrapSDKContext(ctx), from, to, amt)
}

func (a securityBankKeeperAdapter) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return a.inner.SendCoinsFromAccountToModule(sdk.WrapSDKContext(ctx), senderAddr, recipientModule, amt)
}

func (a securityBankKeeperAdapter) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return a.inner.SendCoinsFromModuleToAccount(sdk.WrapSDKContext(ctx), senderModule, recipientAddr, amt)
}

// securityStakingKeeperAdapter wraps the staking keeper for the security module
type securityStakingKeeperAdapter struct {
	inner *stakingkeeper.Keeper
}

func newSecurityStakingKeeperAdapter(inner *stakingkeeper.Keeper) securitykeeper.StakingKeeper {
	return securityStakingKeeperAdapter{inner: inner}
}

func (a securityStakingKeeperAdapter) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (interface{}, bool) {
	validator, err := a.inner.GetValidator(sdk.WrapSDKContext(ctx), addr)
	if err != nil {
		return nil, false
	}
	return validator, true
}

func (a securityStakingKeeperAdapter) Jail(ctx sdk.Context, consAddr sdk.ConsAddress) {
	_ = a.inner.Jail(sdk.WrapSDKContext(ctx), consAddr)
}

func (a securityStakingKeeperAdapter) Unjail(ctx sdk.Context, consAddr sdk.ConsAddress) {
	_ = a.inner.Unjail(sdk.WrapSDKContext(ctx), consAddr)
}

func (a securityStakingKeeperAdapter) Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor string) string {
	factor, err := sdkmath.LegacyNewDecFromStr(slashFactor)
	if err != nil {
		return "0"
	}
	slashed, err := a.inner.Slash(sdk.WrapSDKContext(ctx), consAddr, infractionHeight, power, factor)
	if err != nil {
		return "0"
	}
	return slashed.String()
}

// securityAccountKeeperAdapter wraps the account keeper for the security module
type securityAccountKeeperAdapter struct {
	inner authkeeper.AccountKeeper
}

func newSecurityAccountKeeperAdapter(inner authkeeper.AccountKeeper) securitykeeper.AccountKeeper {
	return securityAccountKeeperAdapter{inner: inner}
}

func (a securityAccountKeeperAdapter) GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI {
	return a.inner.GetAccount(sdk.WrapSDKContext(ctx), addr)
}

func (a securityAccountKeeperAdapter) SetAccount(ctx sdk.Context, acc sdk.AccountI) {
	a.inner.SetAccount(sdk.WrapSDKContext(ctx), acc)
}
