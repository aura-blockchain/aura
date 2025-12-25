"""
Aura SDK Error Hierarchy

Typed exception classes for error handling throughout the SDK.
"""

from typing import Any, Optional


class AuraError(Exception):
    """Base exception class for all Aura SDK errors."""

    def __init__(
        self,
        message: str,
        code: str = "AURA_ERROR",
        details: Optional[dict[str, Any]] = None,
    ):
        super().__init__(message)
        self.message = message
        self.code = code
        self.details = details or {}

    def __repr__(self) -> str:
        return f"{self.__class__.__name__}(code={self.code!r}, message={self.message!r})"


# ============================================================================
# Connection Errors
# ============================================================================


class ConnectionError(AuraError):
    """Error connecting to the Aura network."""

    def __init__(self, message: str, details: Optional[dict[str, Any]] = None):
        super().__init__(message, "CONNECTION_ERROR", details)


class NotConnectedError(ConnectionError):
    """Client is not connected to the network."""

    def __init__(self, operation: Optional[str] = None):
        msg = "Client not connected"
        if operation:
            msg += f" for operation: {operation}"
        msg += ". Call connect() first."
        super().__init__(msg, {"operation": operation})


class EndpointError(ConnectionError):
    """Error with a specific endpoint."""

    def __init__(self, endpoint: str, reason: Optional[str] = None):
        msg = f"Endpoint error: {endpoint}"
        if reason:
            msg += f" - {reason}"
        super().__init__(msg, {"endpoint": endpoint, "reason": reason})


# ============================================================================
# Wallet Errors
# ============================================================================


class WalletError(AuraError):
    """Base class for wallet-related errors."""

    def __init__(self, message: str, details: Optional[dict[str, Any]] = None):
        super().__init__(message, "WALLET_ERROR", details)


class WalletNotInitializedError(WalletError):
    """Wallet has not been initialized."""

    def __init__(self, operation: Optional[str] = None):
        msg = "Wallet not initialized"
        if operation:
            msg += f" for operation: {operation}"
        super().__init__(msg, {"operation": operation})


class InvalidMnemonicError(WalletError):
    """Invalid mnemonic phrase provided."""

    def __init__(self):
        super().__init__("Invalid mnemonic phrase")


class NoAccountsError(WalletError):
    """No accounts found in wallet."""

    def __init__(self):
        super().__init__("No accounts found in wallet")


class SigningError(WalletError):
    """Error signing a transaction."""

    def __init__(self, reason: str):
        super().__init__(f"Signing failed: {reason}", {"reason": reason})


# ============================================================================
# Transaction Errors
# ============================================================================


class TransactionError(AuraError):
    """Base class for transaction-related errors."""

    def __init__(
        self,
        message: str,
        code: str = "TX_ERROR",
        tx_hash: Optional[str] = None,
        details: Optional[dict[str, Any]] = None,
    ):
        super().__init__(message, code, {**(details or {}), "tx_hash": tx_hash})
        self.tx_hash = tx_hash


class TransactionBroadcastError(TransactionError):
    """Error broadcasting a transaction."""

    def __init__(self, reason: str, tx_hash: Optional[str] = None):
        super().__init__(
            f"Transaction broadcast failed: {reason}",
            "TX_BROADCAST_ERROR",
            tx_hash,
            {"reason": reason},
        )


class TransactionTimeoutError(TransactionError):
    """Transaction timed out waiting for confirmation."""

    def __init__(self, tx_hash: str, timeout_ms: int):
        super().__init__(
            f"Transaction timed out after {timeout_ms}ms",
            "TX_TIMEOUT",
            tx_hash,
            {"timeout_ms": timeout_ms},
        )


class InsufficientFundsError(TransactionError):
    """Insufficient funds for transaction."""

    def __init__(self, required: str, available: str, denom: str):
        super().__init__(
            f"Insufficient funds: required {required} {denom}, available {available} {denom}",
            "INSUFFICIENT_FUNDS",
            details={"required": required, "available": available, "denom": denom},
        )
        self.required = required
        self.available = available
        self.denom = denom


class GasEstimationError(TransactionError):
    """Error estimating gas for transaction."""

    def __init__(self, reason: str):
        super().__init__(
            f"Gas estimation failed: {reason}",
            "GAS_ESTIMATION_ERROR",
            details={"reason": reason},
        )


# ============================================================================
# Query Errors
# ============================================================================


class QueryError(AuraError):
    """Base class for query-related errors."""

    def __init__(self, message: str, details: Optional[dict[str, Any]] = None):
        super().__init__(message, "QUERY_ERROR", details)


class NotFoundError(QueryError):
    """Requested resource not found."""

    def __init__(self, resource: str, identifier: str):
        super().__init__(
            f"{resource} not found: {identifier}",
            {"resource": resource, "identifier": identifier},
        )


class InvalidResponseError(QueryError):
    """Invalid response received from node."""

    def __init__(self, expected: str, received: str):
        super().__init__(
            f"Invalid response: expected {expected}, received {received}",
            {"expected": expected, "received": received},
        )


# ============================================================================
# Module-Specific Errors
# ============================================================================


class BridgeError(AuraError):
    """Base class for bridge-related errors."""

    def __init__(self, message: str, details: Optional[dict[str, Any]] = None):
        super().__init__(message, "BRIDGE_ERROR", details)


class BridgeTransferError(BridgeError):
    """Error with a bridge transfer."""

    def __init__(
        self,
        message: str,
        transfer_id: Optional[str] = None,
        details: Optional[dict[str, Any]] = None,
    ):
        super().__init__(message, {**(details or {}), "transfer_id": transfer_id})
        self.transfer_id = transfer_id


class UnsupportedChainError(BridgeError):
    """Chain is not supported by the bridge."""

    def __init__(self, chain_id: str):
        super().__init__(f"Unsupported chain: {chain_id}", {"chain_id": chain_id})


class IdentityError(AuraError):
    """Base class for identity-related errors."""

    def __init__(self, message: str, details: Optional[dict[str, Any]] = None):
        super().__init__(message, "IDENTITY_ERROR", details)


class DIDNotFoundError(IdentityError):
    """DID not found."""

    def __init__(self, did: str):
        super().__init__(f"DID not found: {did}", {"did": did})


class ComplianceError(AuraError):
    """Base class for compliance-related errors."""

    def __init__(self, message: str, details: Optional[dict[str, Any]] = None):
        super().__init__(message, "COMPLIANCE_ERROR", details)


class KYCRequiredError(ComplianceError):
    """KYC verification required for operation."""

    def __init__(self, operation: str, required_level: Optional[str] = None):
        super().__init__(
            f"KYC required for operation: {operation}",
            {"operation": operation, "required_level": required_level},
        )


class DEXError(AuraError):
    """Base class for DEX-related errors."""

    def __init__(self, message: str, details: Optional[dict[str, Any]] = None):
        super().__init__(message, "DEX_ERROR", details)


class SlippageExceededError(DEXError):
    """Slippage exceeded maximum tolerance."""

    def __init__(self, expected: str, actual: str, max_slippage: str):
        super().__init__(
            f"Slippage exceeded: expected {expected}, got {actual} (max: {max_slippage})",
            {"expected": expected, "actual": actual, "max_slippage": max_slippage},
        )


class PoolNotFoundError(DEXError):
    """Liquidity pool not found."""

    def __init__(self, pool_id: str):
        super().__init__(f"Pool not found: {pool_id}", {"pool_id": pool_id})


# ============================================================================
# Validation Errors
# ============================================================================


class ValidationError(AuraError):
    """Base class for validation errors."""

    def __init__(
        self,
        message: str,
        field: Optional[str] = None,
        details: Optional[dict[str, Any]] = None,
    ):
        super().__init__(message, "VALIDATION_ERROR", {**(details or {}), "field": field})
        self.field = field


class InvalidAddressError(ValidationError):
    """Invalid address format."""

    def __init__(self, address: str, expected_prefix: Optional[str] = None):
        msg = f"Invalid address: {address}"
        if expected_prefix:
            msg += f" (expected prefix: {expected_prefix})"
        super().__init__(msg, "address", {"address": address, "expected_prefix": expected_prefix})


class InvalidAmountError(ValidationError):
    """Invalid amount format or value."""

    def __init__(self, amount: str, reason: Optional[str] = None):
        msg = f"Invalid amount: {amount}"
        if reason:
            msg += f" - {reason}"
        super().__init__(msg, "amount", {"amount": amount, "reason": reason})


# ============================================================================
# Helper Functions
# ============================================================================


def is_aura_error(error: Exception) -> bool:
    """Check if an exception is an AuraError."""
    return isinstance(error, AuraError)


def wrap_error(error: Exception, context: Optional[str] = None) -> AuraError:
    """Wrap an unknown exception in an AuraError."""
    if isinstance(error, AuraError):
        return error

    message = str(error)
    if context:
        message = f"{context}: {message}"

    return AuraError(message, "UNKNOWN_ERROR", {"original_error": repr(error)})
