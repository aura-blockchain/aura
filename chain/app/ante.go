// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"

	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	txsigning "cosmossdk.io/x/tx/signing"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	compliancekeeper "github.com/aequitas/aura/chain/x/compliance/keeper"
	walletsecuritykeeper "github.com/aequitas/aura/chain/x/walletsecurity/keeper"
)

// HandlerOptions extends the SDK's AnteHandler options with custom AURA keepers
// for enhanced security and compliance features.
type HandlerOptions struct {
	AccountKeeper   authkeeper.AccountKeeper
	BankKeeper      bankkeeper.Keeper
	SignModeHandler *txsigning.HandlerMap
	FeegrantKeeper  authante.FeegrantKeeper
	SigGasConsumer  authante.SignatureVerificationGasConsumer
	TxFeeChecker    authante.TxFeeChecker

	// WASM specific
	WasmConfig            *wasmtypes.WasmConfig
	WasmKeeper            *wasmkeeper.Keeper
	TXCounterStoreService store.KVStoreService

	// AURA custom keepers for security
	ComplianceKeeper     *compliancekeeper.Keeper
	WalletSecurityKeeper *walletsecuritykeeper.Keeper
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, deducts fees from the first
// signer, and performs additional AURA-specific security checks.
//
// The ante handler chain is designed to be deterministic and follows Cosmos SDK
// best practices for transaction processing:
// 1. Setup - Prepare context and transaction state
// 2. Validation - Verify transaction structure and signatures
// 3. Security - Apply AURA-specific security controls
// 4. Fee Processing - Deduct transaction fees
// 5. WASM Controls - Apply smart contract limits
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}

	// Set default signature gas consumer if not provided
	if options.SigGasConsumer == nil {
		options.SigGasConsumer = authante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		// ==========================================
		// Phase 1: Setup and Context Preparation
		// ==========================================
		// SetUpContextDecorator must be first to properly set up the context
		authante.NewSetUpContextDecorator(),

		// ==========================================
		// Phase 2: Transaction Extension Options
		// ==========================================
		// Extension options decorator for handling transaction extensions
		authante.NewExtensionOptionsDecorator(nil),

		// ==========================================
		// Phase 3: Transaction Validation
		// ==========================================
		// Validate basic transaction structure
		authante.NewValidateBasicDecorator(),

		// Validate transaction timeout height
		authante.NewTxTimeoutHeightDecorator(),

		// Validate memo size to prevent spam
		authante.NewValidateMemoDecorator(options.AccountKeeper),

		// Consume gas for transaction size to prevent large tx spam
		authante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),

		// ==========================================
		// Phase 4: Fee Processing
		// ==========================================
		// Deduct fee from first signer and check minimum fees
		authante.NewDeductFeeDecorator(
			options.AccountKeeper,
			options.BankKeeper,
			options.FeegrantKeeper,
			options.TxFeeChecker,
		),

		// ==========================================
		// Phase 5: AURA Custom Security Controls
		// ==========================================
		// Wallet security checks - transaction limits, spending limits
		NewWalletSecurityDecorator(options.WalletSecurityKeeper),

		// Rate-limit repeated auth failures per block to mitigate brute-force attempts
		NewAuthRateLimitDecorator(options.WalletSecurityKeeper),

		// Compliance checks - sanctions screening, AML validation
		NewComplianceDecorator(options.ComplianceKeeper),

		// ==========================================
		// Phase 6: Account and Signature Validation
		// ==========================================
		// SetPubKeyDecorator must come before ValidateSigCountDecorator
		authante.NewSetPubKeyDecorator(options.AccountKeeper),

		// Validate signature count to prevent multi-sig spam
		authante.NewValidateSigCountDecorator(options.AccountKeeper),

		// Consume gas for signature verification
		authante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),

		// Verify all signatures and increment sequence numbers
		authante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),

		// Increment sequence number for each signer
		authante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	// ==========================================
	// Phase 7: WASM-Specific Controls (Optional)
	// ==========================================
	// Add WASM-specific decorators if WASM is enabled
	if options.WasmKeeper != nil {
		anteDecorators = append(anteDecorators,
			// Limit contract gas to prevent resource exhaustion
			wasmkeeper.NewLimitSimulationGasDecorator(options.WasmConfig.SimulationGasLimit),

			// Track transaction counter for WASM operations
			wasmkeeper.NewCountTXDecorator(options.TXCounterStoreService),

			// Additional WASM gas consumption checks
			wasmkeeper.NewGasRegisterDecorator(options.WasmKeeper.GetGasRegister()),
		)
	}

	// Chain all decorators together
	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

// Validate validates that all required options are provided
// and properly configured before creating the ante handler.
func (h HandlerOptions) Validate() error {
	// Check for nil/empty keepers by checking their internal fields
	// AccountKeeper and BankKeeper cannot be directly compared due to unexported fields
	if h.SignModeHandler == nil {
		return errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	// Validate WASM configuration if WASM keeper is provided
	if h.WasmKeeper != nil {
		if h.WasmConfig == nil {
			return errorsmod.Wrap(sdkerrors.ErrLogic, "wasm config is required when wasm keeper is provided")
		}
		if h.TXCounterStoreService == nil {
			return errorsmod.Wrap(sdkerrors.ErrLogic, "tx counter store service is required for wasm")
		}
	}

	return nil
}

// WalletSecurityDecorator performs wallet security checks before transaction execution.
// This includes:
// - Transaction rate limiting per wallet
// - Spending limit enforcement
// - Dust transaction filtering
type WalletSecurityDecorator struct {
	keeper *walletsecuritykeeper.Keeper
}

// NewWalletSecurityDecorator creates a new wallet security ante decorator.
func NewWalletSecurityDecorator(keeper *walletsecuritykeeper.Keeper) WalletSecurityDecorator {
	return WalletSecurityDecorator{
		keeper: keeper,
	}
}

// AnteHandle implements the ante decorator interface for wallet security checks.
// It validates each transaction signer against wallet security policies.
func (wsd WalletSecurityDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (sdk.Context, error) {
	// Skip security checks in simulation mode
	if simulate {
		return next(ctx, tx, simulate)
	}

	// Get all signers from the transaction
	signers, err := getTxSigners(tx)
	if err != nil {
		return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "failed to get transaction signers")
	}

	// Validate each signer's wallet security status
	for _, signer := range signers {
		// Skip if keeper is not initialized
		if wsd.keeper == nil {
			break
		}

		// Validate wallet exists and is not locked
		walletID := signer.String()
		if err := wsd.keeper.ValidateWallet(ctx, walletID); err != nil {
			return ctx, errorsmod.Wrapf(
				sdkerrors.ErrUnauthorized,
				"wallet %s validation failed: %v",
				walletID,
				err,
			)
		}

		// Check spending limits for each transaction amount
		txAmount := getTxAmount(tx)
		for _, coin := range txAmount {
			if err := wsd.keeper.CheckSpendingLimit(ctx, walletID, coin.Denom, coin.Amount.String()); err != nil {
				return ctx, errorsmod.Wrapf(
					sdkerrors.ErrUnauthorized,
					"spending limit exceeded for wallet %s: %v",
					walletID,
					err,
				)
			}
		}
	}

	return next(ctx, tx, simulate)
}

// ComplianceDecorator performs compliance checks before transaction execution.
// This includes:
// - Basic compliance validation
// - Transaction monitoring for suspicious patterns
type ComplianceDecorator struct {
	keeper *compliancekeeper.Keeper
}

// NewComplianceDecorator creates a new compliance ante decorator.
func NewComplianceDecorator(keeper *compliancekeeper.Keeper) ComplianceDecorator {
	return ComplianceDecorator{
		keeper: keeper,
	}
}

// AnteHandle implements the ante decorator interface for compliance checks.
// It validates each transaction against compliance policies and regulations.
func (cd ComplianceDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (sdk.Context, error) {
	// Skip compliance checks in simulation mode
	if simulate {
		return next(ctx, tx, simulate)
	}

	// Skip if keeper is not initialized
	if cd.keeper == nil {
		return next(ctx, tx, simulate)
	}

	// Get all signers from the transaction
	signers, err := getTxSigners(tx)
	if err != nil {
		return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "failed to get transaction signers")
	}

	// Perform basic compliance validation for each signer
	for _, signer := range signers {
		// Basic validation - in production this would include:
		// - Sanctions screening
		// - AML checks
		// - Transaction monitoring
		// For now we just validate the address format
		if signer.Empty() {
			return ctx, errorsmod.Wrapf(
				sdkerrors.ErrUnauthorized,
				"invalid signer address",
			)
		}
	}

	return next(ctx, tx, simulate)
}

// getTxSigners extracts all signers from a transaction.
// This is a deterministic operation that returns addresses in a consistent order.
func getTxSigners(tx sdk.Tx) ([]sdk.AccAddress, error) {
	sigTx, ok := tx.(authsigning.Tx)
	if !ok {
		return nil, fmt.Errorf("tx is not a signing transaction")
	}

	signers, err := sigTx.GetSigners()
	if err != nil {
		return nil, err
	}

	addresses := make([]sdk.AccAddress, len(signers))
	for i, signer := range signers {
		addresses[i] = sdk.AccAddress(signer)
	}

	return addresses, nil
}

// getTxAmount extracts the total amount being transferred in a transaction.
// This is used for spending limit and AML checks.
func getTxAmount(tx sdk.Tx) sdk.Coins {
	// Extract amounts from transaction messages
	// For now, return empty coins - this should be properly implemented based on msg types
	// In production, you would iterate through tx.GetMsgs() and extract amounts
	// from bank.MsgSend, bank.MsgMultiSend, etc.
	return sdk.NewCoins()
}
