package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// MonitoredBankKeeper wraps the standard bank keeper to add transaction monitoring.
//
// This keeper intercepts all coin transfers to evaluate compliance rules before
// execution. It provides real-time AML (Anti-Money Laundering) monitoring and
// can block transactions that violate compliance policies.
//
// Security considerations:
//   - All monitoring is performed BEFORE the transaction is executed (pre-flight checks)
//   - Critical risk transactions are blocked and never reach the underlying bank keeper
//   - Transaction monitoring does not affect gas costs (monitoring errors don't revert)
//   - All monitoring decisions are logged and stored for audit trail
//
// Architecture:
//   - Wraps bankkeeper.Keeper using composition
//   - Intercepts SendCoins, InputOutputCoins, and SendCoinsFromModuleToAccount
//   - Delegates to underlying keeper after monitoring passes
//   - Updates AML profiles after successful transactions
//
// Usage in app.go:
//
//	complianceKeeper := compliancekeeper.NewKeeper(...)
//	baseBankKeeper := bankkeeper.NewBaseKeeper(...)
//	monitoredBankKeeper := compliancekeeper.NewMonitoredBankKeeper(baseBankKeeper, complianceKeeper)
//	// Use monitoredBankKeeper instead of baseBankKeeper in other modules
type MonitoredBankKeeper struct {
	bankkeeper.Keeper
	complianceKeeper *Keeper
}

// NewMonitoredBankKeeper creates a new monitored bank keeper
//
// Parameters:
//   - baseKeeper: The underlying bank keeper to wrap
//   - complianceKeeper: The compliance keeper for transaction monitoring
//
// Returns:
//   - A new MonitoredBankKeeper that wraps the base keeper
func NewMonitoredBankKeeper(baseKeeper bankkeeper.Keeper, complianceKeeper *Keeper) *MonitoredBankKeeper {
	return &MonitoredBankKeeper{
		Keeper:           baseKeeper,
		complianceKeeper: complianceKeeper,
	}
}

// SendCoins transfers coins from one account to another with compliance monitoring.
//
// This method performs the following steps:
//  1. Evaluate transaction monitoring rules (large amount, velocity, structuring)
//  2. Check sanctions screening (if enabled)
//  3. Block transaction if critical risk detected
//  4. Execute transfer via underlying bank keeper
//  5. Update AML profiles for both sender and recipient
//
// Security considerations:
//   - Monitoring occurs BEFORE transfer (pre-flight checks)
//   - Critical risk alerts block the transaction
//   - Blocked transactions return an error and do not modify state
//   - AML profiles are updated only after successful transfer
//
// Parameters:
//   - ctx: SDK context for state access
//   - from: Sender address
//   - to: Recipient address
//   - amount: Coins to transfer
//
// Returns:
//   - error: Compliance violation error or underlying bank keeper error
//
// Events emitted:
//   - EventTypeTransactionAlert for each compliance alert (via MonitorTransaction)
//   - Standard bank module events (via underlying keeper)
func (k *MonitoredBankKeeper) SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	// Monitor transaction before executing
	alerts, err := k.complianceKeeper.MonitorTransaction(ctx, from, to, amount)
	if err != nil {
		// Monitoring error should not block transaction, but log the error
		k.complianceKeeper.logger(ctx).Error("transaction monitoring failed",
			"from", from.String(),
			"to", to.String(),
			"amount", amount.String(),
			"error", err.Error(),
		)
	}

	// Check if transaction should be blocked based on alerts
	if shouldBlock, reason := k.complianceKeeper.ShouldBlockTransaction(alerts); shouldBlock {
		// Emit compliance violation event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"compliance_violation",
				sdk.NewAttribute("from", from.String()),
				sdk.NewAttribute("to", to.String()),
				sdk.NewAttribute("amount", amount.String()),
				sdk.NewAttribute("reason", reason),
				sdk.NewAttribute("blocked", "true"),
			),
		)

		return fmt.Errorf("transaction blocked by compliance: %s", reason)
	}

	// Execute the transfer via underlying bank keeper
	if err := k.Keeper.SendCoins(ctx, from, to, amount); err != nil {
		return fmt.Errorf("error in SendCoins: %w", err)
	}

	// Update AML profiles after successful transfer
	if err := k.complianceKeeper.UpdateAMLProfileOnTransaction(ctx, from.String(), amount); err != nil {
		k.complianceKeeper.logger(ctx).Error("failed to update sender AML profile",
			"address", from.String(),
			"error", err.Error(),
		)
	}

	if err := k.complianceKeeper.UpdateAMLProfileOnTransaction(ctx, to.String(), amount); err != nil {
		k.complianceKeeper.logger(ctx).Error("failed to update recipient AML profile",
			"address", to.String(),
			"error", err.Error(),
		)
	}

	return nil
}

// InputOutputCoins performs multi-input, multi-output transfers with compliance monitoring.
//
// This method monitors all senders in the inputs list for compliance violations.
// If any sender triggers a critical alert, the entire multi-send transaction is blocked.
//
// Security considerations:
//   - ALL inputs are monitored, not just the first one
//   - A critical alert from ANY sender blocks the entire transaction
//   - Atomic operation: either all transfers succeed or all fail
//   - AML profiles updated for all participants after success
//
// Parameters:
//   - ctx: SDK context for state access
//   - inputs: List of input accounts and amounts
//   - outputs: List of output accounts and amounts
//
// Returns:
//   - error: Compliance violation or underlying keeper error
func (k *MonitoredBankKeeper) InputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	// Monitor each input for compliance
	allAlerts := make([]*types.TransactionAlert, 0, 64)
	for _, input := range inputs {
		fromAddr := sdk.MustAccAddressFromBech32(input.Address)

		// For multi-output transactions, monitor against each output
		for _, output := range outputs {
			toAddr := sdk.MustAccAddressFromBech32(output.Address)

			alerts, err := k.complianceKeeper.MonitorTransaction(ctx, fromAddr, toAddr, input.Coins)
			if err != nil {
				k.complianceKeeper.logger(ctx).Error("transaction monitoring failed",
					"from", input.Address,
					"to", output.Address,
					"amount", input.Coins.String(),
					"error", err.Error(),
				)
			}

			allAlerts = append(allAlerts, alerts...)
		}
	}

	// Check if any transaction should be blocked
	if shouldBlock, reason := k.complianceKeeper.ShouldBlockTransaction(allAlerts); shouldBlock {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"compliance_violation",
				sdk.NewAttribute("operation", "multi_send"),
				sdk.NewAttribute("reason", reason),
				sdk.NewAttribute("blocked", "true"),
			),
		)

		return fmt.Errorf("multi-send blocked by compliance: %s", reason)
	}

	// Execute the multi-send via underlying bank keeper
	for _, input := range inputs {
		if err := k.Keeper.InputOutputCoins(ctx, input, outputs); err != nil {
			return fmt.Errorf("error in InputOutputCoins: %w", err)
		}
	}

	// Update AML profiles for all participants
	for _, input := range inputs {
		if err := k.complianceKeeper.UpdateAMLProfileOnTransaction(ctx, input.Address, input.Coins); err != nil {
			k.complianceKeeper.logger(ctx).Error("failed to update AML profile",
				"address", input.Address,
				"error", err.Error(),
			)
		}
	}

	for _, output := range outputs {
		if err := k.complianceKeeper.UpdateAMLProfileOnTransaction(ctx, output.Address, output.Coins); err != nil {
			k.complianceKeeper.logger(ctx).Error("failed to update AML profile",
				"address", output.Address,
				"error", err.Error(),
			)
		}
	}

	return nil
}

// SendCoinsFromModuleToAccount transfers coins from a module to an account with monitoring.
//
// Module-to-account transfers are monitored to detect large withdrawals or
// suspicious patterns. Module accounts are considered trusted senders, but
// recipient accounts are still monitored.
//
// Security considerations:
//   - Module accounts are trusted (no sanctions screening on sender)
//   - Recipient account is fully monitored
//   - Large withdrawals still trigger alerts
//   - Useful for detecting reward farming or other abuse patterns
//
// Parameters:
//   - ctx: SDK context for state access
//   - senderModule: Name of the sender module
//   - recipientAddr: Recipient account address
//   - amount: Coins to transfer
//
// Returns:
//   - error: Compliance violation or underlying keeper error
func (k *MonitoredBankKeeper) SendCoinsFromModuleToAccount(
	ctx sdk.Context,
	senderModule string,
	recipientAddr sdk.AccAddress,
	amount sdk.Coins,
) error {
	// Get module address
	moduleAddr := k.GetModuleAddress(senderModule)
	if moduleAddr == nil {
		return fmt.Errorf("module account not found: %s", senderModule)
	}

	// Monitor transaction (primarily for recipient)
	alerts, err := k.complianceKeeper.MonitorTransaction(ctx, moduleAddr, recipientAddr, amount)
	if err != nil {
		k.complianceKeeper.logger(ctx).Error("transaction monitoring failed",
			"module", senderModule,
			"to", recipientAddr.String(),
			"amount", amount.String(),
			"error", err.Error(),
		)
	}

	// Check if transaction should be blocked
	if shouldBlock, reason := k.complianceKeeper.ShouldBlockTransaction(alerts); shouldBlock {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"compliance_violation",
				sdk.NewAttribute("module", senderModule),
				sdk.NewAttribute("to", recipientAddr.String()),
				sdk.NewAttribute("amount", amount.String()),
				sdk.NewAttribute("reason", reason),
				sdk.NewAttribute("blocked", "true"),
			),
		)

		return fmt.Errorf("module transfer blocked by compliance: %s", reason)
	}

	// Execute transfer via underlying keeper
	if err := k.Keeper.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amount); err != nil {
		return fmt.Errorf("error in SendCoinsFromModuleToAccount: %w", err)
	}

	// Update recipient AML profile
	if err := k.complianceKeeper.UpdateAMLProfileOnTransaction(ctx, recipientAddr.String(), amount); err != nil {
		k.complianceKeeper.logger(ctx).Error("failed to update recipient AML profile",
			"address", recipientAddr.String(),
			"error", err.Error(),
		)
	}

	return nil
}

// SendCoinsFromAccountToModule transfers coins from an account to a module with monitoring.
//
// Account-to-module transfers are monitored to detect unusual deposit patterns.
// Module accounts are considered trusted recipients, but sender accounts are
// still monitored for AML compliance.
//
// Parameters:
//   - ctx: SDK context for state access
//   - senderAddr: Sender account address
//   - recipientModule: Name of the recipient module
//   - amount: Coins to transfer
//
// Returns:
//   - error: Compliance violation or underlying keeper error
func (k *MonitoredBankKeeper) SendCoinsFromAccountToModule(
	ctx sdk.Context,
	senderAddr sdk.AccAddress,
	recipientModule string,
	amount sdk.Coins,
) error {
	// Get module address
	moduleAddr := k.GetModuleAddress(recipientModule)
	if moduleAddr == nil {
		return fmt.Errorf("module account not found: %s", recipientModule)
	}

	// Monitor transaction
	alerts, err := k.complianceKeeper.MonitorTransaction(ctx, senderAddr, moduleAddr, amount)
	if err != nil {
		k.complianceKeeper.logger(ctx).Error("transaction monitoring failed",
			"from", senderAddr.String(),
			"module", recipientModule,
			"amount", amount.String(),
			"error", err.Error(),
		)
	}

	// Check if transaction should be blocked
	if shouldBlock, reason := k.complianceKeeper.ShouldBlockTransaction(alerts); shouldBlock {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"compliance_violation",
				sdk.NewAttribute("from", senderAddr.String()),
				sdk.NewAttribute("module", recipientModule),
				sdk.NewAttribute("amount", amount.String()),
				sdk.NewAttribute("reason", reason),
				sdk.NewAttribute("blocked", "true"),
			),
		)

		return fmt.Errorf("module transfer blocked by compliance: %s", reason)
	}

	// Execute transfer via underlying keeper
	if err := k.Keeper.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amount); err != nil {
		return fmt.Errorf("error in SendCoinsFromAccountToModule: %w", err)
	}

	// Update sender AML profile
	if err := k.complianceKeeper.UpdateAMLProfileOnTransaction(ctx, senderAddr.String(), amount); err != nil {
		k.complianceKeeper.logger(ctx).Error("failed to update sender AML profile",
			"address", senderAddr.String(),
			"error", err.Error(),
		)
	}

	return nil
}

// GetModuleAddress returns the address for a module account
func (k *MonitoredBankKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	// Module address is derived from the module name using the auth module
	return authtypes.NewModuleAddress(moduleName)
}

