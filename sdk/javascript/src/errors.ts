/**
 * Aura SDK Error Hierarchy
 *
 * Base error classes for typed error handling throughout the SDK.
 */

/**
 * Base error class for all Aura SDK errors
 */
export class AuraError extends Error {
  public readonly code: string;
  public readonly details?: Record<string, unknown>;

  constructor(message: string, code: string, details?: Record<string, unknown>) {
    super(message);
    this.name = 'AuraError';
    this.code = code;
    this.details = details;
    Object.setPrototypeOf(this, new.target.prototype);
  }
}

// ============================================================================
// Connection Errors
// ============================================================================

export class ConnectionError extends AuraError {
  constructor(message: string, details?: Record<string, unknown>) {
    super(message, 'CONNECTION_ERROR', details);
    this.name = 'ConnectionError';
  }
}

export class NotConnectedError extends ConnectionError {
  constructor(operation?: string) {
    super(
      `Client not connected${operation ? ` for operation: ${operation}` : ''}. Call connect() first.`,
      { operation }
    );
    this.name = 'NotConnectedError';
  }
}

export class EndpointError extends ConnectionError {
  constructor(endpoint: string, reason?: string) {
    super(
      `Endpoint error: ${endpoint}${reason ? ` - ${reason}` : ''}`,
      { endpoint, reason }
    );
    this.name = 'EndpointError';
  }
}

// ============================================================================
// Wallet Errors
// ============================================================================

export class WalletError extends AuraError {
  constructor(message: string, details?: Record<string, unknown>) {
    super(message, 'WALLET_ERROR', details);
    this.name = 'WalletError';
  }
}

export class WalletNotInitializedError extends WalletError {
  constructor(operation?: string) {
    super(
      `Wallet not initialized${operation ? ` for operation: ${operation}` : ''}`,
      { operation }
    );
    this.name = 'WalletNotInitializedError';
  }
}

export class InvalidMnemonicError extends WalletError {
  constructor() {
    super('Invalid mnemonic phrase');
    this.name = 'InvalidMnemonicError';
  }
}

export class NoAccountsError extends WalletError {
  constructor() {
    super('No accounts found in wallet');
    this.name = 'NoAccountsError';
  }
}

export class SigningError extends WalletError {
  constructor(reason: string) {
    super(`Signing failed: ${reason}`, { reason });
    this.name = 'SigningError';
  }
}

// ============================================================================
// Transaction Errors
// ============================================================================

export class TransactionError extends AuraError {
  public readonly txHash?: string;

  constructor(message: string, code: string, txHash?: string, details?: Record<string, unknown>) {
    super(message, code, { ...details, txHash });
    this.name = 'TransactionError';
    this.txHash = txHash;
  }
}

export class TransactionBroadcastError extends TransactionError {
  constructor(reason: string, txHash?: string) {
    super(`Transaction broadcast failed: ${reason}`, 'TX_BROADCAST_ERROR', txHash, { reason });
    this.name = 'TransactionBroadcastError';
  }
}

export class TransactionTimeoutError extends TransactionError {
  constructor(txHash: string, timeout: number) {
    super(`Transaction timed out after ${timeout}ms`, 'TX_TIMEOUT', txHash, { timeout });
    this.name = 'TransactionTimeoutError';
  }
}

export class InsufficientFundsError extends TransactionError {
  public readonly required: string;
  public readonly available: string;

  constructor(required: string, available: string, denom: string) {
    super(
      `Insufficient funds: required ${required} ${denom}, available ${available} ${denom}`,
      'INSUFFICIENT_FUNDS',
      undefined,
      { required, available, denom }
    );
    this.name = 'InsufficientFundsError';
    this.required = required;
    this.available = available;
  }
}

export class GasEstimationError extends TransactionError {
  constructor(reason: string) {
    super(`Gas estimation failed: ${reason}`, 'GAS_ESTIMATION_ERROR', undefined, { reason });
    this.name = 'GasEstimationError';
  }
}

// ============================================================================
// Query Errors
// ============================================================================

export class QueryError extends AuraError {
  constructor(message: string, details?: Record<string, unknown>) {
    super(message, 'QUERY_ERROR', details);
    this.name = 'QueryError';
  }
}

export class NotFoundError extends QueryError {
  constructor(resource: string, identifier: string) {
    super(`${resource} not found: ${identifier}`, { resource, identifier });
    this.name = 'NotFoundError';
  }
}

export class InvalidResponseError extends QueryError {
  constructor(expected: string, received: string) {
    super(`Invalid response: expected ${expected}, received ${received}`, { expected, received });
    this.name = 'InvalidResponseError';
  }
}

// ============================================================================
// Module-Specific Errors
// ============================================================================

export class BridgeError extends AuraError {
  constructor(message: string, details?: Record<string, unknown>) {
    super(message, 'BRIDGE_ERROR', details);
    this.name = 'BridgeError';
  }
}

export class BridgeTransferError extends BridgeError {
  public readonly transferId?: string;

  constructor(message: string, transferId?: string, details?: Record<string, unknown>) {
    super(message, { ...details, transferId });
    this.name = 'BridgeTransferError';
    this.transferId = transferId;
  }
}

export class UnsupportedChainError extends BridgeError {
  constructor(chainId: string) {
    super(`Unsupported chain: ${chainId}`, { chainId });
    this.name = 'UnsupportedChainError';
  }
}

export class IdentityError extends AuraError {
  constructor(message: string, details?: Record<string, unknown>) {
    super(message, 'IDENTITY_ERROR', details);
    this.name = 'IdentityError';
  }
}

export class DIDNotFoundError extends IdentityError {
  constructor(did: string) {
    super(`DID not found: ${did}`, { did });
    this.name = 'DIDNotFoundError';
  }
}

export class ComplianceError extends AuraError {
  constructor(message: string, details?: Record<string, unknown>) {
    super(message, 'COMPLIANCE_ERROR', details);
    this.name = 'ComplianceError';
  }
}

export class KYCRequiredError extends ComplianceError {
  constructor(operation: string, requiredLevel?: string) {
    super(`KYC required for operation: ${operation}`, { operation, requiredLevel });
    this.name = 'KYCRequiredError';
  }
}

export class DEXError extends AuraError {
  constructor(message: string, details?: Record<string, unknown>) {
    super(message, 'DEX_ERROR', details);
    this.name = 'DEXError';
  }
}

export class SlippageExceededError extends DEXError {
  constructor(expected: string, actual: string, maxSlippage: string) {
    super(`Slippage exceeded: expected ${expected}, got ${actual} (max: ${maxSlippage})`, {
      expected,
      actual,
      maxSlippage,
    });
    this.name = 'SlippageExceededError';
  }
}

export class PoolNotFoundError extends DEXError {
  constructor(poolId: string) {
    super(`Pool not found: ${poolId}`, { poolId });
    this.name = 'PoolNotFoundError';
  }
}

// ============================================================================
// Validation Errors
// ============================================================================

export class ValidationError extends AuraError {
  public readonly field?: string;

  constructor(message: string, field?: string, details?: Record<string, unknown>) {
    super(message, 'VALIDATION_ERROR', { ...details, field });
    this.name = 'ValidationError';
    this.field = field;
  }
}

export class InvalidAddressError extends ValidationError {
  constructor(address: string, expectedPrefix?: string) {
    super(
      `Invalid address: ${address}${expectedPrefix ? ` (expected prefix: ${expectedPrefix})` : ''}`,
      'address',
      { address, expectedPrefix }
    );
    this.name = 'InvalidAddressError';
  }
}

export class InvalidAmountError extends ValidationError {
  constructor(amount: string, reason?: string) {
    super(`Invalid amount: ${amount}${reason ? ` - ${reason}` : ''}`, 'amount', { amount, reason });
    this.name = 'InvalidAmountError';
  }
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Type guard to check if error is an AuraError
 */
export function isAuraError(error: unknown): error is AuraError {
  return error instanceof AuraError;
}

/**
 * Wrap unknown error in AuraError
 */
export function wrapError(error: unknown, context?: string): AuraError {
  if (isAuraError(error)) {
    return error;
  }

  const message = error instanceof Error ? error.message : String(error);
  return new AuraError(
    context ? `${context}: ${message}` : message,
    'UNKNOWN_ERROR',
    { originalError: error }
  );
}
