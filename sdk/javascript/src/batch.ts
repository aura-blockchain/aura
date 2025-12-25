/**
 * Batch Operation Helpers for Aura SDK
 *
 * Provides utilities for batching multiple transactions and queries.
 */

import type { Coin } from '@cosmjs/stargate';
import type { EncodeObject } from '@cosmjs/proto-signing';

// ============================================================================
// Types
// ============================================================================

export interface BatchTransferItem {
  recipient: string;
  amount: Coin[];
}

export interface BatchResult<T> {
  success: boolean;
  results: T[];
  errors: { index: number; error: Error }[];
  txHash?: string;
}

export interface BatchQueryResult<T> {
  results: T[];
  errors: { index: number; error: Error }[];
}

export interface BatchConfig {
  maxBatchSize?: number;
  parallelQueries?: number;
  retryOnFailure?: boolean;
  retryAttempts?: number;
  delayBetweenRetries?: number;
}

const DEFAULT_CONFIG: Required<BatchConfig> = {
  maxBatchSize: 50,
  parallelQueries: 10,
  retryOnFailure: true,
  retryAttempts: 3,
  delayBetweenRetries: 1000,
};

// ============================================================================
// Transaction Batching
// ============================================================================

/**
 * BatchTransactionBuilder helps construct multi-message transactions
 */
export class BatchTransactionBuilder {
  private messages: EncodeObject[] = [];
  private memo: string = '';

  /**
   * Add a message to the batch
   */
  addMessage(msg: EncodeObject): this {
    this.messages.push(msg);
    return this;
  }

  /**
   * Add multiple messages to the batch
   */
  addMessages(msgs: EncodeObject[]): this {
    this.messages.push(...msgs);
    return this;
  }

  /**
   * Add a bank send message
   */
  addSend(sender: string, recipient: string, amount: Coin[]): this {
    this.messages.push({
      typeUrl: '/cosmos.bank.v1beta1.MsgSend',
      value: {
        fromAddress: sender,
        toAddress: recipient,
        amount,
      },
    });
    return this;
  }

  /**
   * Add multiple send messages for batch transfers
   */
  addBatchSends(sender: string, transfers: BatchTransferItem[]): this {
    for (const transfer of transfers) {
      this.addSend(sender, transfer.recipient, transfer.amount);
    }
    return this;
  }

  /**
   * Add a delegate message
   */
  addDelegate(delegator: string, validator: string, amount: Coin): this {
    this.messages.push({
      typeUrl: '/cosmos.staking.v1beta1.MsgDelegate',
      value: {
        delegatorAddress: delegator,
        validatorAddress: validator,
        amount,
      },
    });
    return this;
  }

  /**
   * Add an undelegate message
   */
  addUndelegate(delegator: string, validator: string, amount: Coin): this {
    this.messages.push({
      typeUrl: '/cosmos.staking.v1beta1.MsgUndelegate',
      value: {
        delegatorAddress: delegator,
        validatorAddress: validator,
        amount,
      },
    });
    return this;
  }

  /**
   * Add a governance vote message
   */
  addVote(voter: string, proposalId: string, option: number): this {
    this.messages.push({
      typeUrl: '/cosmos.gov.v1beta1.MsgVote',
      value: {
        proposalId: BigInt(proposalId),
        voter,
        option,
      },
    });
    return this;
  }

  /**
   * Set transaction memo
   */
  withMemo(memo: string): this {
    this.memo = memo;
    return this;
  }

  /**
   * Get all messages in the batch
   */
  getMessages(): EncodeObject[] {
    return [...this.messages];
  }

  /**
   * Get the memo
   */
  getMemo(): string {
    return this.memo;
  }

  /**
   * Get the number of messages
   */
  size(): number {
    return this.messages.length;
  }

  /**
   * Clear all messages
   */
  clear(): this {
    this.messages = [];
    this.memo = '';
    return this;
  }
}

// ============================================================================
// Query Batching
// ============================================================================

type QueryFunction<T> = () => Promise<T>;

/**
 * Execute multiple queries in parallel with batching
 */
export async function batchQueries<T>(
  queries: QueryFunction<T>[],
  config: BatchConfig = {}
): Promise<BatchQueryResult<T>> {
  const cfg = { ...DEFAULT_CONFIG, ...config };
  const results: T[] = [];
  const errors: { index: number; error: Error }[] = [];

  // Split into batches
  const batches: QueryFunction<T>[][] = [];
  for (let i = 0; i < queries.length; i += cfg.parallelQueries) {
    batches.push(queries.slice(i, i + cfg.parallelQueries));
  }

  let globalIndex = 0;

  for (const batch of batches) {
    const batchPromises = batch.map(async (query, localIndex) => {
      const index = globalIndex + localIndex;
      try {
        const result = await executeWithRetry(query, cfg);
        return { index, result, error: null };
      } catch (err) {
        return { index, result: null, error: err as Error };
      }
    });

    const batchResults = await Promise.all(batchPromises);

    for (const res of batchResults) {
      if (res.error) {
        errors.push({ index: res.index, error: res.error });
      } else if (res.result !== null) {
        results[res.index] = res.result;
      }
    }

    globalIndex += batch.length;
  }

  return { results, errors };
}

/**
 * Execute a function with retry logic
 */
async function executeWithRetry<T>(
  fn: () => Promise<T>,
  config: Required<BatchConfig>
): Promise<T> {
  let lastError: Error | null = null;

  for (let attempt = 0; attempt < config.retryAttempts; attempt++) {
    try {
      return await fn();
    } catch (err) {
      lastError = err as Error;
      if (!config.retryOnFailure || attempt === config.retryAttempts - 1) {
        throw lastError;
      }
      await delay(config.delayBetweenRetries * Math.pow(2, attempt));
    }
  }

  throw lastError;
}

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Split an array into chunks of specified size
 */
export function chunk<T>(array: T[], size: number): T[][] {
  const chunks: T[][] = [];
  for (let i = 0; i < array.length; i += size) {
    chunks.push(array.slice(i, i + size));
  }
  return chunks;
}

/**
 * Create a batch transfer builder
 */
export function createBatchTransfers(): BatchTransactionBuilder {
  return new BatchTransactionBuilder();
}

/**
 * Helper to validate batch size
 */
export function validateBatchSize(
  items: unknown[],
  maxSize: number = DEFAULT_CONFIG.maxBatchSize
): void {
  if (items.length > maxSize) {
    throw new Error(
      `Batch size ${items.length} exceeds maximum allowed size of ${maxSize}`
    );
  }
}

/**
 * Calculate estimated gas for batch operations
 */
export function estimateBatchGas(
  messageCount: number,
  baseGasPerMessage: number = 100000,
  overheadGas: number = 50000
): number {
  return overheadGas + messageCount * baseGasPerMessage;
}

// ============================================================================
// Multi-Sig Batch Helpers
// ============================================================================

export interface MultiSigBatchItem {
  messages: EncodeObject[];
  memo?: string;
}

/**
 * Create multiple transactions for multi-sig signing
 */
export function createMultiSigBatch(items: MultiSigBatchItem[]): {
  transactions: { messages: EncodeObject[]; memo: string }[];
  totalMessages: number;
} {
  const transactions = items.map((item) => ({
    messages: item.messages,
    memo: item.memo || '',
  }));

  const totalMessages = transactions.reduce(
    (sum, tx) => sum + tx.messages.length,
    0
  );

  return { transactions, totalMessages };
}
