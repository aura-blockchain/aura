// Package errors provides typed error definitions for the Aura SDK.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for type checking with errors.Is()
var (
	ErrConnection         = errors.New("connection error")
	ErrNotConnected       = errors.New("client not connected")
	ErrEndpoint           = errors.New("endpoint error")
	ErrWallet             = errors.New("wallet error")
	ErrWalletNotInit      = errors.New("wallet not initialized")
	ErrInvalidMnemonic    = errors.New("invalid mnemonic")
	ErrNoAccounts         = errors.New("no accounts in wallet")
	ErrSigning            = errors.New("signing error")
	ErrTransaction        = errors.New("transaction error")
	ErrTxBroadcast        = errors.New("transaction broadcast failed")
	ErrTxTimeout          = errors.New("transaction timeout")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrGasEstimation      = errors.New("gas estimation failed")
	ErrQuery              = errors.New("query error")
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidResponse    = errors.New("invalid response")
	ErrBridge             = errors.New("bridge error")
	ErrBridgeTransfer     = errors.New("bridge transfer error")
	ErrUnsupportedChain   = errors.New("unsupported chain")
	ErrIdentity           = errors.New("identity error")
	ErrDIDNotFound        = errors.New("DID not found")
	ErrCompliance         = errors.New("compliance error")
	ErrKYCRequired        = errors.New("KYC required")
	ErrDEX                = errors.New("DEX error")
	ErrSlippageExceeded   = errors.New("slippage exceeded")
	ErrPoolNotFound       = errors.New("pool not found")
	ErrValidation         = errors.New("validation error")
	ErrInvalidAddress     = errors.New("invalid address")
	ErrInvalidAmount      = errors.New("invalid amount")
)

// AuraError is the base error type with additional context.
type AuraError struct {
	Err     error
	Code    string
	Message string
	Details map[string]interface{}
}

func (e *AuraError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

func (e *AuraError) Unwrap() error {
	return e.Err
}

// New creates a new AuraError.
func New(err error, code, message string) *AuraError {
	return &AuraError{
		Err:     err,
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetails adds details to the error.
func (e *AuraError) WithDetails(key string, value interface{}) *AuraError {
	e.Details[key] = value
	return e
}

// ConnectionError represents connection-related errors.
type ConnectionError struct {
	AuraError
	Endpoint string
}

func NewConnectionError(message string) *ConnectionError {
	return &ConnectionError{
		AuraError: AuraError{
			Err:     ErrConnection,
			Code:    "CONNECTION_ERROR",
			Message: message,
			Details: make(map[string]interface{}),
		},
	}
}

func NewNotConnectedError(operation string) *ConnectionError {
	msg := "client not connected"
	if operation != "" {
		msg = fmt.Sprintf("client not connected for operation: %s", operation)
	}
	return &ConnectionError{
		AuraError: AuraError{
			Err:     ErrNotConnected,
			Code:    "NOT_CONNECTED",
			Message: msg,
			Details: map[string]interface{}{"operation": operation},
		},
	}
}

func NewEndpointError(endpoint, reason string) *ConnectionError {
	msg := fmt.Sprintf("endpoint error: %s", endpoint)
	if reason != "" {
		msg = fmt.Sprintf("%s - %s", msg, reason)
	}
	return &ConnectionError{
		AuraError: AuraError{
			Err:     ErrEndpoint,
			Code:    "ENDPOINT_ERROR",
			Message: msg,
			Details: map[string]interface{}{"endpoint": endpoint, "reason": reason},
		},
		Endpoint: endpoint,
	}
}

// WalletError represents wallet-related errors.
type WalletError struct {
	AuraError
}

func NewWalletError(message string) *WalletError {
	return &WalletError{
		AuraError: AuraError{
			Err:     ErrWallet,
			Code:    "WALLET_ERROR",
			Message: message,
			Details: make(map[string]interface{}),
		},
	}
}

func NewWalletNotInitializedError(operation string) *WalletError {
	msg := "wallet not initialized"
	if operation != "" {
		msg = fmt.Sprintf("wallet not initialized for operation: %s", operation)
	}
	return &WalletError{
		AuraError: AuraError{
			Err:     ErrWalletNotInit,
			Code:    "WALLET_NOT_INITIALIZED",
			Message: msg,
			Details: map[string]interface{}{"operation": operation},
		},
	}
}

func NewInvalidMnemonicError() *WalletError {
	return &WalletError{
		AuraError: AuraError{
			Err:     ErrInvalidMnemonic,
			Code:    "INVALID_MNEMONIC",
			Message: "invalid mnemonic phrase",
			Details: make(map[string]interface{}),
		},
	}
}

func NewNoAccountsError() *WalletError {
	return &WalletError{
		AuraError: AuraError{
			Err:     ErrNoAccounts,
			Code:    "NO_ACCOUNTS",
			Message: "no accounts found in wallet",
			Details: make(map[string]interface{}),
		},
	}
}

func NewSigningError(reason string) *WalletError {
	return &WalletError{
		AuraError: AuraError{
			Err:     ErrSigning,
			Code:    "SIGNING_ERROR",
			Message: fmt.Sprintf("signing failed: %s", reason),
			Details: map[string]interface{}{"reason": reason},
		},
	}
}

// TransactionError represents transaction-related errors.
type TransactionError struct {
	AuraError
	TxHash string
}

func NewTransactionError(message string) *TransactionError {
	return &TransactionError{
		AuraError: AuraError{
			Err:     ErrTransaction,
			Code:    "TX_ERROR",
			Message: message,
			Details: make(map[string]interface{}),
		},
	}
}

func NewTxBroadcastError(reason, txHash string) *TransactionError {
	return &TransactionError{
		AuraError: AuraError{
			Err:     ErrTxBroadcast,
			Code:    "TX_BROADCAST_ERROR",
			Message: fmt.Sprintf("transaction broadcast failed: %s", reason),
			Details: map[string]interface{}{"reason": reason, "tx_hash": txHash},
		},
		TxHash: txHash,
	}
}

func NewTxTimeoutError(txHash string, timeoutMs int64) *TransactionError {
	return &TransactionError{
		AuraError: AuraError{
			Err:     ErrTxTimeout,
			Code:    "TX_TIMEOUT",
			Message: fmt.Sprintf("transaction timed out after %dms", timeoutMs),
			Details: map[string]interface{}{"tx_hash": txHash, "timeout_ms": timeoutMs},
		},
		TxHash: txHash,
	}
}

func NewInsufficientFundsError(required, available, denom string) *TransactionError {
	return &TransactionError{
		AuraError: AuraError{
			Err:     ErrInsufficientFunds,
			Code:    "INSUFFICIENT_FUNDS",
			Message: fmt.Sprintf("insufficient funds: required %s %s, available %s %s", required, denom, available, denom),
			Details: map[string]interface{}{"required": required, "available": available, "denom": denom},
		},
	}
}

func NewGasEstimationError(reason string) *TransactionError {
	return &TransactionError{
		AuraError: AuraError{
			Err:     ErrGasEstimation,
			Code:    "GAS_ESTIMATION_ERROR",
			Message: fmt.Sprintf("gas estimation failed: %s", reason),
			Details: map[string]interface{}{"reason": reason},
		},
	}
}

// QueryError represents query-related errors.
type QueryError struct {
	AuraError
}

func NewQueryError(message string) *QueryError {
	return &QueryError{
		AuraError: AuraError{
			Err:     ErrQuery,
			Code:    "QUERY_ERROR",
			Message: message,
			Details: make(map[string]interface{}),
		},
	}
}

func NewNotFoundError(resource, identifier string) *QueryError {
	return &QueryError{
		AuraError: AuraError{
			Err:     ErrNotFound,
			Code:    "NOT_FOUND",
			Message: fmt.Sprintf("%s not found: %s", resource, identifier),
			Details: map[string]interface{}{"resource": resource, "identifier": identifier},
		},
	}
}

func NewInvalidResponseError(expected, received string) *QueryError {
	return &QueryError{
		AuraError: AuraError{
			Err:     ErrInvalidResponse,
			Code:    "INVALID_RESPONSE",
			Message: fmt.Sprintf("invalid response: expected %s, received %s", expected, received),
			Details: map[string]interface{}{"expected": expected, "received": received},
		},
	}
}

// BridgeError represents bridge-related errors.
type BridgeError struct {
	AuraError
	TransferID string
}

func NewBridgeError(message string) *BridgeError {
	return &BridgeError{
		AuraError: AuraError{
			Err:     ErrBridge,
			Code:    "BRIDGE_ERROR",
			Message: message,
			Details: make(map[string]interface{}),
		},
	}
}

func NewBridgeTransferError(message, transferID string) *BridgeError {
	return &BridgeError{
		AuraError: AuraError{
			Err:     ErrBridgeTransfer,
			Code:    "BRIDGE_TRANSFER_ERROR",
			Message: message,
			Details: map[string]interface{}{"transfer_id": transferID},
		},
		TransferID: transferID,
	}
}

func NewUnsupportedChainError(chainID string) *BridgeError {
	return &BridgeError{
		AuraError: AuraError{
			Err:     ErrUnsupportedChain,
			Code:    "UNSUPPORTED_CHAIN",
			Message: fmt.Sprintf("unsupported chain: %s", chainID),
			Details: map[string]interface{}{"chain_id": chainID},
		},
	}
}

// IdentityError represents identity-related errors.
type IdentityError struct {
	AuraError
	DID string
}

func NewIdentityError(message string) *IdentityError {
	return &IdentityError{
		AuraError: AuraError{
			Err:     ErrIdentity,
			Code:    "IDENTITY_ERROR",
			Message: message,
			Details: make(map[string]interface{}),
		},
	}
}

func NewDIDNotFoundError(did string) *IdentityError {
	return &IdentityError{
		AuraError: AuraError{
			Err:     ErrDIDNotFound,
			Code:    "DID_NOT_FOUND",
			Message: fmt.Sprintf("DID not found: %s", did),
			Details: map[string]interface{}{"did": did},
		},
		DID: did,
	}
}

// DEXError represents DEX-related errors.
type DEXError struct {
	AuraError
	PoolID string
}

func NewDEXError(message string) *DEXError {
	return &DEXError{
		AuraError: AuraError{
			Err:     ErrDEX,
			Code:    "DEX_ERROR",
			Message: message,
			Details: make(map[string]interface{}),
		},
	}
}

func NewSlippageExceededError(expected, actual, maxSlippage string) *DEXError {
	return &DEXError{
		AuraError: AuraError{
			Err:     ErrSlippageExceeded,
			Code:    "SLIPPAGE_EXCEEDED",
			Message: fmt.Sprintf("slippage exceeded: expected %s, got %s (max: %s)", expected, actual, maxSlippage),
			Details: map[string]interface{}{"expected": expected, "actual": actual, "max_slippage": maxSlippage},
		},
	}
}

func NewPoolNotFoundError(poolID string) *DEXError {
	return &DEXError{
		AuraError: AuraError{
			Err:     ErrPoolNotFound,
			Code:    "POOL_NOT_FOUND",
			Message: fmt.Sprintf("pool not found: %s", poolID),
			Details: map[string]interface{}{"pool_id": poolID},
		},
		PoolID: poolID,
	}
}

// ValidationError represents validation-related errors.
type ValidationError struct {
	AuraError
	Field string
}

func NewValidationError(message, field string) *ValidationError {
	return &ValidationError{
		AuraError: AuraError{
			Err:     ErrValidation,
			Code:    "VALIDATION_ERROR",
			Message: message,
			Details: map[string]interface{}{"field": field},
		},
		Field: field,
	}
}

func NewInvalidAddressError(address, expectedPrefix string) *ValidationError {
	msg := fmt.Sprintf("invalid address: %s", address)
	if expectedPrefix != "" {
		msg = fmt.Sprintf("%s (expected prefix: %s)", msg, expectedPrefix)
	}
	return &ValidationError{
		AuraError: AuraError{
			Err:     ErrInvalidAddress,
			Code:    "INVALID_ADDRESS",
			Message: msg,
			Details: map[string]interface{}{"address": address, "expected_prefix": expectedPrefix},
		},
		Field: "address",
	}
}

func NewInvalidAmountError(amount, reason string) *ValidationError {
	msg := fmt.Sprintf("invalid amount: %s", amount)
	if reason != "" {
		msg = fmt.Sprintf("%s - %s", msg, reason)
	}
	return &ValidationError{
		AuraError: AuraError{
			Err:     ErrInvalidAmount,
			Code:    "INVALID_AMOUNT",
			Message: msg,
			Details: map[string]interface{}{"amount": amount, "reason": reason},
		},
		Field: "amount",
	}
}

// IsAuraError checks if an error is an AuraError.
func IsAuraError(err error) bool {
	var auraErr *AuraError
	return errors.As(err, &auraErr)
}

// Wrap wraps an error with additional context.
func Wrap(err error, context string) *AuraError {
	if err == nil {
		return nil
	}

	var auraErr *AuraError
	if errors.As(err, &auraErr) {
		return auraErr
	}

	msg := err.Error()
	if context != "" {
		msg = fmt.Sprintf("%s: %s", context, msg)
	}

	return &AuraError{
		Err:     err,
		Code:    "UNKNOWN_ERROR",
		Message: msg,
		Details: make(map[string]interface{}),
	}
}
