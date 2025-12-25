// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdkmath "cosmossdk.io/math"

	compliancekeeper "github.com/aequitas/aura/chain/x/compliance/keeper"
	confidencescorekeeper "github.com/aequitas/aura/chain/x/confidencescore/keeper"
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

// monitoredBankKeeperAdapter wraps the bank keeper with transaction monitoring for AML compliance.
// This adapter intercepts coin transfer methods to evaluate compliance rules before execution.
//
// IMPORTANT: This is the integration point for transaction monitoring as described in
// chain/x/compliance/TRANSACTION_MONITORING.md. All coin transfers go through this adapter.
type monitoredBankKeeperAdapter struct {
	inner            bankkeeper.BaseKeeper
	complianceKeeper *compliancekeeper.Keeper
}

func newMonitoredBankKeeperAdapter(bankKeeper bankkeeper.BaseKeeper, complianceKeeper *compliancekeeper.Keeper) monitoredBankKeeperAdapter {
	return monitoredBankKeeperAdapter{
		inner:            bankKeeper,
		complianceKeeper: complianceKeeper,
	}
}

func (a monitoredBankKeeperAdapter) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	// Monitor transaction before executing
	alerts, err := a.complianceKeeper.MonitorTransaction(ctx, fromAddr, toAddr, amt)
	if err != nil {
		// Log error but don't block transaction - monitoring failure shouldn't prevent valid transactions
		ctx.Logger().Error("transaction monitoring failed",
			"from", fromAddr.String(),
			"to", toAddr.String(),
			"amount", amt.String(),
			"error", err.Error(),
		)
	}

	// Check if transaction should be blocked based on alerts
	if shouldBlock, reason := a.complianceKeeper.ShouldBlockTransaction(alerts); shouldBlock {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"compliance_violation",
				sdk.NewAttribute("from", fromAddr.String()),
				sdk.NewAttribute("to", toAddr.String()),
				sdk.NewAttribute("amount", amt.String()),
				sdk.NewAttribute("reason", reason),
				sdk.NewAttribute("blocked", "true"),
			),
		)
		return fmt.Errorf("transaction blocked by compliance: %s", reason)
	}

	// Execute transfer via underlying bank keeper
	if err := a.inner.SendCoins(sdk.WrapSDKContext(ctx), fromAddr, toAddr, amt); err != nil {
		return err
	}

	// Update AML profiles after successful transfer
	if err := a.complianceKeeper.UpdateAMLProfileOnTransaction(ctx, fromAddr.String(), amt); err != nil {
		ctx.Logger().Error("failed to update sender AML profile",
			"address", fromAddr.String(),
			"error", err.Error(),
		)
	}

	if err := a.complianceKeeper.UpdateAMLProfileOnTransaction(ctx, toAddr.String(), amt); err != nil {
		ctx.Logger().Error("failed to update recipient AML profile",
			"address", toAddr.String(),
			"error", err.Error(),
		)
	}

	return nil
}

func (a monitoredBankKeeperAdapter) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	// Get module address for monitoring
	moduleAddr := sdk.AccAddress(sdk.MustBech32ifyAddressBytes("aura", sdk.AccAddress(recipientModule).Bytes()))

	// Monitor transaction
	alerts, err := a.complianceKeeper.MonitorTransaction(ctx, senderAddr, moduleAddr, amt)
	if err != nil {
		ctx.Logger().Error("transaction monitoring failed",
			"from", senderAddr.String(),
			"module", recipientModule,
			"amount", amt.String(),
			"error", err.Error(),
		)
	}

	// Check if transaction should be blocked
	if shouldBlock, reason := a.complianceKeeper.ShouldBlockTransaction(alerts); shouldBlock {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"compliance_violation",
				sdk.NewAttribute("from", senderAddr.String()),
				sdk.NewAttribute("module", recipientModule),
				sdk.NewAttribute("amount", amt.String()),
				sdk.NewAttribute("reason", reason),
				sdk.NewAttribute("blocked", "true"),
			),
		)
		return fmt.Errorf("module transfer blocked by compliance: %s", reason)
	}

	// Execute transfer
	if err := a.inner.SendCoinsFromAccountToModule(sdk.WrapSDKContext(ctx), senderAddr, recipientModule, amt); err != nil {
		return err
	}

	// Update sender AML profile
	if err := a.complianceKeeper.UpdateAMLProfileOnTransaction(ctx, senderAddr.String(), amt); err != nil {
		ctx.Logger().Error("failed to update sender AML profile",
			"address", senderAddr.String(),
			"error", err.Error(),
		)
	}

	return nil
}

func (a monitoredBankKeeperAdapter) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	// Get module address for monitoring
	moduleAddr := sdk.AccAddress(sdk.MustBech32ifyAddressBytes("aura", sdk.AccAddress(senderModule).Bytes()))

	// Monitor transaction (primarily for recipient)
	alerts, err := a.complianceKeeper.MonitorTransaction(ctx, moduleAddr, recipientAddr, amt)
	if err != nil {
		ctx.Logger().Error("transaction monitoring failed",
			"module", senderModule,
			"to", recipientAddr.String(),
			"amount", amt.String(),
			"error", err.Error(),
		)
	}

	// Check if transaction should be blocked
	if shouldBlock, reason := a.complianceKeeper.ShouldBlockTransaction(alerts); shouldBlock {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"compliance_violation",
				sdk.NewAttribute("module", senderModule),
				sdk.NewAttribute("to", recipientAddr.String()),
				sdk.NewAttribute("amount", amt.String()),
				sdk.NewAttribute("reason", reason),
				sdk.NewAttribute("blocked", "true"),
			),
		)
		return fmt.Errorf("module transfer blocked by compliance: %s", reason)
	}

	// Execute transfer
	if err := a.inner.SendCoinsFromModuleToAccount(sdk.WrapSDKContext(ctx), senderModule, recipientAddr, amt); err != nil {
		return err
	}

	// Update recipient AML profile
	if err := a.complianceKeeper.UpdateAMLProfileOnTransaction(ctx, recipientAddr.String(), amt); err != nil {
		ctx.Logger().Error("failed to update recipient AML profile",
			"address", recipientAddr.String(),
			"error", err.Error(),
		)
	}

	return nil
}

func (a monitoredBankKeeperAdapter) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	// Module-to-module transfers don't need monitoring (trusted internal transfers)
	return a.inner.SendCoinsFromModuleToModule(sdk.WrapSDKContext(ctx), senderModule, recipientModule, amt)
}

func (a monitoredBankKeeperAdapter) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	// Minting doesn't need monitoring (trusted operation)
	return a.inner.MintCoins(sdk.WrapSDKContext(ctx), moduleName, amt)
}

func (a monitoredBankKeeperAdapter) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	// Burning doesn't need monitoring (trusted operation)
	return a.inner.BurnCoins(sdk.WrapSDKContext(ctx), moduleName, amt)
}

func (a monitoredBankKeeperAdapter) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	// Balance queries don't need monitoring
	return a.inner.GetBalance(sdk.WrapSDKContext(ctx), addr, denom)
}

func (a monitoredBankKeeperAdapter) GetSupply(ctx sdk.Context, denom string) sdk.Coin {
	// Supply queries don't need monitoring
	return a.inner.GetSupply(sdk.WrapSDKContext(ctx), denom)
}

// securityStakingAdapter bridges staking keeper expectations for the security module.
type securityStakingAdapter struct {
	inner *stakingkeeper.Keeper
}

func newSecurityStakingAdapter(inner *stakingkeeper.Keeper) securitykeeper.StakingKeeper {
	return securityStakingAdapter{inner: inner}
}

func (a securityStakingAdapter) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (interface{}, bool) {
	validator, err := a.inner.GetValidator(sdk.WrapSDKContext(ctx), addr)
	return validator, err == nil
}

func (a securityStakingAdapter) Jail(ctx sdk.Context, consAddr sdk.ConsAddress) error {
	if err := a.inner.Jail(sdk.WrapSDKContext(ctx), consAddr); err != nil {
		return fmt.Errorf("staking jail failed for %s: %w", consAddr, err)
	}
	return nil
}

func (a securityStakingAdapter) Unjail(ctx sdk.Context, consAddr sdk.ConsAddress) error {
	if err := a.inner.Unjail(sdk.WrapSDKContext(ctx), consAddr); err != nil {
		return fmt.Errorf("staking unjail failed for %s: %w", consAddr, err)
	}
	return nil
}

func (a securityStakingAdapter) Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor string) (string, error) {
	dec, err := sdkmath.LegacyNewDecFromStr(slashFactor)
	if err != nil {
		return "", fmt.Errorf("invalid slash factor %q: %w", slashFactor, err)
	}
	tokens, err := a.inner.Slash(sdk.WrapSDKContext(ctx), consAddr, infractionHeight, power, dec)
	if err != nil {
		return "", fmt.Errorf("staking slash failed for %s: %w", consAddr, err)
	}
	return tokens.String(), nil
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

func (a accountKeeperAdapter) GetModuleAddress(moduleName string) sdk.AccAddress {
	return a.inner.GetModuleAddress(moduleName)
}

// bridgeStakingAdapter wraps the staking keeper for bridge module expectations.
type bridgeStakingAdapter struct {
	inner *stakingkeeper.Keeper
}

func newBridgeStakingAdapter(inner *stakingkeeper.Keeper) bridgeStakingAdapter {
	return bridgeStakingAdapter{inner: inner}
}

func (a bridgeStakingAdapter) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (stakingtypes.Validator, bool) {
	validator, err := a.inner.GetValidator(sdk.WrapSDKContext(ctx), addr)
	if err != nil {
		return stakingtypes.Validator{}, false
	}
	return validator, true
}

func (a bridgeStakingAdapter) Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor sdkmath.LegacyDec) sdkmath.Int {
	slashed, _ := a.inner.Slash(sdk.WrapSDKContext(ctx), consAddr, infractionHeight, power, slashFactor)
	return slashed
}

func (a bridgeStakingAdapter) Jail(ctx sdk.Context, consAddr sdk.ConsAddress) {
	_ = a.inner.Jail(sdk.WrapSDKContext(ctx), consAddr)
}

func (a bridgeStakingAdapter) Unjail(ctx sdk.Context, valAddr sdk.ValAddress) {
	validator, found := a.GetValidator(ctx, valAddr)
	if !found {
		return
	}
	consAddrBytes, err := validator.GetConsAddr()
	if err != nil {
		return
	}
	_ = a.inner.Unjail(sdk.WrapSDKContext(ctx), consAddrBytes)
}

func (a bridgeStakingAdapter) IsBonded(ctx sdk.Context, addr sdk.ValAddress) bool {
	validator, found := a.GetValidator(ctx, addr)
	if !found {
		return false
	}
	return validator.IsBonded()
}

func (a bridgeStakingAdapter) GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (stakingtypes.Validator, bool) {
	validator, err := a.inner.GetValidatorByConsAddr(sdk.WrapSDKContext(ctx), consAddr)
	if err != nil {
		return stakingtypes.Validator{}, false
	}
	return validator, true
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
// CONTRACT REGISTRY KEEPER ADAPTERS
// ============================================================================

// contractRegistryComplianceAdapter wraps the compliance keeper for contract registry expectations.
type contractRegistryComplianceAdapter struct {
	inner *compliancekeeper.Keeper
}

func newContractRegistryComplianceAdapter(inner *compliancekeeper.Keeper) contractRegistryComplianceAdapter {
	return contractRegistryComplianceAdapter{inner: inner}
}

// GetKYCLevel retrieves the KYC level for an address.
// Returns the KYC level as uint32 (matching the KYCLevel enum values).
func (a contractRegistryComplianceAdapter) GetKYCLevel(ctx sdk.Context, address string) (uint32, error) {
	record, err := a.inner.GetKYCRecord(ctx, address)
	if err != nil {
		// Return NONE level if no record found
		return 1, nil // KYC_LEVEL_NONE = 1
	}
	return uint32(record.KycLevel), nil
}

// ScreenForSanctions checks if an address is sanctioned.
// Returns true if the address is sanctioned (blocked), false otherwise.
func (a contractRegistryComplianceAdapter) ScreenForSanctions(ctx sdk.Context, address string) (bool, error) {
	isSanctioned := a.inner.IsAddressSanctioned(ctx, address)
	return isSanctioned, nil
}

// contractRegistryVCAdapter wraps the VC registry keeper for contract registry expectations.
type contractRegistryVCAdapter struct {
	inner *vckeeper.Keeper
}

func newContractRegistryVCAdapter(inner *vckeeper.Keeper) contractRegistryVCAdapter {
	return contractRegistryVCAdapter{inner: inner}
}

// HasVC checks if a user has a specific type of verifiable credential.
// The ctx parameter is interface{} to match the expected interface, but we unwrap it to sdk.Context.
func (a contractRegistryVCAdapter) HasVC(ctx interface{}, address string, vcType string) bool {
	// Unwrap context - support both sdk.Context and context.Context
	var sdkCtx sdk.Context
	switch c := ctx.(type) {
	case sdk.Context:
		sdkCtx = c
	case context.Context:
		sdkCtx = sdk.UnwrapSDKContext(c)
	default:
		return false
	}

	// List all VCs for the user
	vcs := a.inner.ListUserVCs(sdkCtx, address, 0, 0)

	// Check if any active VC matches the requested type
	for _, vc := range vcs {
		if vc.Status == 1 { // VCStatus_VC_STATUS_ACTIVE = 1
			// Match by VC type string representation
			vcTypeStr := vc.VcType.String()
			if vcTypeStr == vcType {
				return true
			}
		}
	}

	return false
}

// contractRegistryConfidenceScoreAdapter wraps the confidence score keeper for contract registry expectations.
// This adapter forwards calls to the actual confidence score keeper with proper context handling.
type contractRegistryConfidenceScoreAdapter struct {
	inner *confidencescorekeeper.Keeper
}

func newContractRegistryConfidenceScoreAdapter(inner *confidencescorekeeper.Keeper) contractRegistryConfidenceScoreAdapter {
	return contractRegistryConfidenceScoreAdapter{inner: inner}
}

// GetUserScore retrieves the total confidence score for a user.
// The ctx parameter is passed through to the underlying keeper for proper state access.
func (a contractRegistryConfidenceScoreAdapter) GetUserScore(ctx sdk.Context, address string) (uint64, bool) {
	return a.inner.GetUserScore(ctx, address)
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

func (a securityStakingKeeperAdapter) Jail(ctx sdk.Context, consAddr sdk.ConsAddress) error {
	if err := a.inner.Jail(sdk.WrapSDKContext(ctx), consAddr); err != nil {
		return fmt.Errorf("staking jail failed for %s: %w", consAddr, err)
	}
	return nil
}

func (a securityStakingKeeperAdapter) Unjail(ctx sdk.Context, consAddr sdk.ConsAddress) error {
	if err := a.inner.Unjail(sdk.WrapSDKContext(ctx), consAddr); err != nil {
		return fmt.Errorf("staking unjail failed for %s: %w", consAddr, err)
	}
	return nil
}

func (a securityStakingKeeperAdapter) Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor string) (string, error) {
	factor, err := sdkmath.LegacyNewDecFromStr(slashFactor)
	if err != nil {
		return "", fmt.Errorf("invalid slash factor %q: %w", slashFactor, err)
	}
	slashed, err := a.inner.Slash(sdk.WrapSDKContext(ctx), consAddr, infractionHeight, power, factor)
	if err != nil {
		return "", fmt.Errorf("staking slash failed for %s: %w", consAddr, err)
	}
	return slashed.String(), nil
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
