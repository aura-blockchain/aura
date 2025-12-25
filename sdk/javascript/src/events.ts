/**
 * Event Subscription System for Aura SDK
 *
 * Provides typed event subscriptions for blockchain events.
 */

import { WebSocket } from 'ws';

// ============================================================================
// Event Types
// ============================================================================

export interface BlockEvent {
  type: 'block';
  height: number;
  hash: string;
  timestamp: string;
  txCount: number;
}

export interface TxEvent {
  type: 'tx';
  hash: string;
  height: number;
  code: number;
  gasUsed: number;
  gasWanted: number;
  events: EventAttribute[];
}

export interface EventAttribute {
  type: string;
  attributes: { key: string; value: string }[];
}

// Module-specific events
export interface TransferEvent {
  type: 'transfer';
  sender: string;
  recipient: string;
  amount: string;
  denom: string;
}

export interface BridgeTransferEvent {
  type: 'bridge_transfer';
  transferId: string;
  sender: string;
  recipient: string;
  amount: string;
  targetChain: string;
  status: 'initiated' | 'completed' | 'failed';
}

export interface IdentityEvent {
  type: 'identity';
  action: 'created' | 'updated' | 'deleted';
  did: string;
  owner: string;
}

export interface GovernanceEvent {
  type: 'governance';
  action: 'proposal_submitted' | 'vote_cast' | 'proposal_passed' | 'proposal_rejected';
  proposalId: string;
  voter?: string;
}

export interface DEXEvent {
  type: 'dex';
  action: 'swap' | 'add_liquidity' | 'remove_liquidity' | 'order_created' | 'order_filled';
  poolId?: string;
  sender: string;
  amounts?: string[];
}

export type AuraEvent =
  | BlockEvent
  | TxEvent
  | TransferEvent
  | BridgeTransferEvent
  | IdentityEvent
  | GovernanceEvent
  | DEXEvent;

// ============================================================================
// Event Filter Types
// ============================================================================

export interface EventFilter {
  type?: string | string[];
  sender?: string;
  recipient?: string;
  module?: string;
  action?: string;
  minHeight?: number;
  maxHeight?: number;
}

export type EventHandler<T extends AuraEvent = AuraEvent> = (event: T) => void | Promise<void>;
export type ErrorHandler = (error: Error) => void;

// ============================================================================
// Event Subscriber
// ============================================================================

interface Subscription {
  id: string;
  filter: EventFilter;
  handler: EventHandler;
}

/**
 * EventSubscriber manages WebSocket connections and event subscriptions.
 */
export class EventSubscriber {
  private wsEndpoint: string;
  private ws: WebSocket | null = null;
  private subscriptions: Map<string, Subscription> = new Map();
  private errorHandler: ErrorHandler | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private isConnected = false;
  private subscriptionCounter = 0;

  constructor(wsEndpoint: string) {
    this.wsEndpoint = wsEndpoint;
  }

  /**
   * Connect to the WebSocket endpoint
   */
  async connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.wsEndpoint);

        this.ws.onopen = () => {
          this.isConnected = true;
          this.reconnectAttempts = 0;
          this.resubscribeAll();
          resolve();
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event.data.toString());
        };

        this.ws.onerror = (error) => {
          if (this.errorHandler) {
            this.errorHandler(new Error(`WebSocket error: ${error.message}`));
          }
        };

        this.ws.onclose = () => {
          this.isConnected = false;
          this.attemptReconnect();
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Disconnect from the WebSocket endpoint
   */
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.isConnected = false;
    this.subscriptions.clear();
  }

  /**
   * Subscribe to events matching the filter
   */
  subscribe<T extends AuraEvent>(
    filter: EventFilter,
    handler: EventHandler<T>
  ): string {
    const id = `sub_${++this.subscriptionCounter}`;
    this.subscriptions.set(id, {
      id,
      filter,
      handler: handler as EventHandler,
    });

    if (this.isConnected && this.ws) {
      this.sendSubscription(filter);
    }

    return id;
  }

  /**
   * Unsubscribe from events
   */
  unsubscribe(subscriptionId: string): boolean {
    return this.subscriptions.delete(subscriptionId);
  }

  /**
   * Subscribe to new blocks
   */
  onBlock(handler: EventHandler<BlockEvent>): string {
    return this.subscribe({ type: 'block' }, handler);
  }

  /**
   * Subscribe to transactions
   */
  onTransaction(handler: EventHandler<TxEvent>, filter?: Partial<EventFilter>): string {
    return this.subscribe({ ...filter, type: 'tx' }, handler);
  }

  /**
   * Subscribe to transfers involving an address
   */
  onTransfer(
    address: string,
    handler: EventHandler<TransferEvent>
  ): string {
    return this.subscribe(
      { type: 'transfer', sender: address },
      handler
    );
  }

  /**
   * Subscribe to bridge transfer events
   */
  onBridgeTransfer(handler: EventHandler<BridgeTransferEvent>): string {
    return this.subscribe({ type: 'bridge_transfer', module: 'bridge' }, handler);
  }

  /**
   * Subscribe to identity events
   */
  onIdentity(handler: EventHandler<IdentityEvent>): string {
    return this.subscribe({ type: 'identity', module: 'identity' }, handler);
  }

  /**
   * Subscribe to governance events
   */
  onGovernance(handler: EventHandler<GovernanceEvent>): string {
    return this.subscribe({ type: 'governance', module: 'governance' }, handler);
  }

  /**
   * Subscribe to DEX events
   */
  onDEX(handler: EventHandler<DEXEvent>): string {
    return this.subscribe({ type: 'dex', module: 'dex' }, handler);
  }

  /**
   * Set error handler
   */
  onError(handler: ErrorHandler): void {
    this.errorHandler = handler;
  }

  /**
   * Check if connected
   */
  get connected(): boolean {
    return this.isConnected;
  }

  private handleMessage(data: string): void {
    try {
      const message = JSON.parse(data);
      const event = this.parseEvent(message);

      if (event) {
        for (const subscription of this.subscriptions.values()) {
          if (this.matchesFilter(event, subscription.filter)) {
            Promise.resolve(subscription.handler(event)).catch((err) => {
              if (this.errorHandler) {
                this.errorHandler(err);
              }
            });
          }
        }
      }
    } catch (error) {
      if (this.errorHandler) {
        this.errorHandler(error as Error);
      }
    }
  }

  private parseEvent(message: unknown): AuraEvent | null {
    if (!message || typeof message !== 'object') {
      return null;
    }

    const msg = message as Record<string, unknown>;
    const result = msg.result as Record<string, unknown> | undefined;

    if (!result || !result.data) {
      return null;
    }

    const data = result.data as Record<string, unknown>;
    const eventType = data.type as string;

    // Parse based on event type
    switch (eventType) {
      case 'tendermint/event/NewBlock':
        return this.parseBlockEvent(data);
      case 'tendermint/event/Tx':
        return this.parseTxEvent(data);
      default:
        return null;
    }
  }

  private parseBlockEvent(data: Record<string, unknown>): BlockEvent | null {
    const value = data.value as Record<string, unknown> | undefined;
    if (!value || !value.block) {
      return null;
    }

    const block = value.block as Record<string, unknown>;
    const header = block.header as Record<string, unknown>;

    return {
      type: 'block',
      height: parseInt(header.height as string, 10),
      hash: (block.last_commit as Record<string, unknown>)?.block_id as string || '',
      timestamp: header.time as string,
      txCount: ((block.data as Record<string, unknown>)?.txs as unknown[])?.length || 0,
    };
  }

  private parseTxEvent(data: Record<string, unknown>): TxEvent | null {
    const value = data.value as Record<string, unknown> | undefined;
    if (!value) {
      return null;
    }

    const txResult = value.TxResult as Record<string, unknown>;
    if (!txResult) {
      return null;
    }

    const result = txResult.result as Record<string, unknown>;

    return {
      type: 'tx',
      hash: txResult.hash as string || '',
      height: parseInt(txResult.height as string, 10),
      code: (result?.code as number) || 0,
      gasUsed: parseInt(result?.gas_used as string || '0', 10),
      gasWanted: parseInt(result?.gas_wanted as string || '0', 10),
      events: this.parseEventAttributes(result?.events as unknown[]),
    };
  }

  private parseEventAttributes(events: unknown[]): EventAttribute[] {
    if (!events || !Array.isArray(events)) {
      return [];
    }

    return events.map((e) => {
      const event = e as Record<string, unknown>;
      return {
        type: event.type as string,
        attributes: ((event.attributes as unknown[]) || []).map((a) => {
          const attr = a as Record<string, string>;
          return {
            key: attr.key || '',
            value: attr.value || '',
          };
        }),
      };
    });
  }

  private matchesFilter(event: AuraEvent, filter: EventFilter): boolean {
    if (filter.type) {
      const types = Array.isArray(filter.type) ? filter.type : [filter.type];
      if (!types.includes(event.type)) {
        return false;
      }
    }

    if (filter.minHeight && 'height' in event && event.height < filter.minHeight) {
      return false;
    }

    if (filter.maxHeight && 'height' in event && event.height > filter.maxHeight) {
      return false;
    }

    // Additional filter matching based on event type
    if (filter.sender && 'sender' in event && event.sender !== filter.sender) {
      return false;
    }

    if (filter.recipient && 'recipient' in event && event.recipient !== filter.recipient) {
      return false;
    }

    return true;
  }

  private sendSubscription(filter: EventFilter): void {
    if (!this.ws || !this.isConnected) {
      return;
    }

    // Build Tendermint query string
    const query = this.buildQuery(filter);
    const message = {
      jsonrpc: '2.0',
      method: 'subscribe',
      id: Date.now(),
      params: { query },
    };

    this.ws.send(JSON.stringify(message));
  }

  private buildQuery(filter: EventFilter): string {
    const conditions: string[] = [];

    if (filter.type === 'block') {
      conditions.push("tm.event='NewBlock'");
    } else if (filter.type === 'tx') {
      conditions.push("tm.event='Tx'");
    } else {
      conditions.push("tm.event='Tx'");
    }

    if (filter.module) {
      conditions.push(`message.module='${filter.module}'`);
    }

    if (filter.action) {
      conditions.push(`message.action='${filter.action}'`);
    }

    if (filter.sender) {
      conditions.push(`message.sender='${filter.sender}'`);
    }

    return conditions.join(' AND ');
  }

  private resubscribeAll(): void {
    for (const subscription of this.subscriptions.values()) {
      this.sendSubscription(subscription.filter);
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      if (this.errorHandler) {
        this.errorHandler(new Error('Max reconnection attempts reached'));
      }
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    setTimeout(() => {
      this.connect().catch((error) => {
        if (this.errorHandler) {
          this.errorHandler(error);
        }
      });
    }, delay);
  }
}

// ============================================================================
// Factory Function
// ============================================================================

/**
 * Create an event subscriber for the given WebSocket endpoint
 */
export function createEventSubscriber(wsEndpoint: string): EventSubscriber {
  return new EventSubscriber(wsEndpoint);
}
