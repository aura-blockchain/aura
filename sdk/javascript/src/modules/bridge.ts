import { AuraClient } from '../client';
import {
  BridgeTransfer,
  BridgeParams,
  InitiateBridgeParams,
  CompleteBridgeParams,
  BridgeStats,
  BridgeSecurity,
  TxResult,
  GasOptions,
} from '../types';
import axios from 'axios';

/**
 * Bridge Module
 * Handles cross-chain bridge operations for transferring assets between chains
 */
export class BridgeModule {
  constructor(private client: AuraClient) {}

  /**
   * Initiate a bridge transfer to another chain
   * @param params - Bridge transfer parameters
   * @param options - Transaction options
   * @returns Transaction result
   */
  async initiateBridgeTransfer(
    params: InitiateBridgeParams,
    options?: GasOptions
  ): Promise<TxResult> {
    try {
      const msg = {
        typeUrl: '/aura.bridge.v1beta1.MsgInitiateBridge',
        value: {
          sender: params.sender,
          recipient: params.recipient,
          amount: {
            denom: params.denom,
            amount: params.amount,
          },
          targetChain: params.targetChain,
          timeout: params.timeout || 3600,
        },
      };

      const txBuilder = this.client.getTxBuilder();
      return await txBuilder.signAndBroadcast(params.sender, [msg], {
        ...options,
        memo: params.memo,
      });
    } catch (error) {
      throw new Error(`Failed to initiate bridge transfer: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Get bridge transfer by ID
   * @param id - Transfer ID
   * @returns Bridge transfer information
   */
  async getBridgeTransfer(id: string): Promise<BridgeTransfer> {
    try {
      const config = this.client.getConfig();
      if (!config.restEndpoint) {
        throw new Error('REST endpoint not configured');
      }

      const response = await axios.get(
        `${config.restEndpoint}/aura/bridge/v1beta1/transfers/${id}`
      );

      return this.parseBridgeTransfer(response.data.transfer);
    } catch (error) {
      throw new Error(`Failed to get bridge transfer: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Get all bridge transfers for an address
   * @param address - User address
   * @returns List of bridge transfers
   */
  async getBridgeTransfers(address: string): Promise<BridgeTransfer[]> {
    try {
      const config = this.client.getConfig();
      if (!config.restEndpoint) {
        throw new Error('REST endpoint not configured');
      }

      const response = await axios.get(
        `${config.restEndpoint}/aura/bridge/v1beta1/transfers`,
        { params: { address } }
      );

      return (response.data.transfers || []).map((t: any) =>
        this.parseBridgeTransfer(t)
      );
    } catch (error) {
      throw new Error(`Failed to get bridge transfers: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Complete a bridge transfer with proof
   * @param params - Complete bridge parameters
   * @param options - Transaction options
   * @returns Transaction result
   */
  async completeBridgeTransfer(
    params: CompleteBridgeParams,
    options?: GasOptions
  ): Promise<TxResult> {
    try {
      const msg = {
        typeUrl: '/aura.bridge.v1beta1.MsgCompleteBridge',
        value: {
          transferId: params.transferId,
          proof: params.proof,
          height: params.height,
          signatures: params.signatures,
        },
      };

      const client = this.client.getSigningClient();
      const accounts = await client.getAccount(await this.getFirstAccount());
      if (!accounts) {
        throw new Error('No account found');
      }

      const txBuilder = this.client.getTxBuilder();
      return await txBuilder.signAndBroadcast(accounts.address, [msg], options);
    } catch (error) {
      throw new Error(`Failed to complete bridge transfer: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Get bridge parameters
   * @returns Bridge parameters
   */
  async getBridgeParams(): Promise<BridgeParams> {
    try {
      const config = this.client.getConfig();
      if (!config.restEndpoint) {
        throw new Error('REST endpoint not configured');
      }

      const response = await axios.get(
        `${config.restEndpoint}/aura/bridge/v1beta1/params`
      );

      return response.data.params;
    } catch (error) {
      throw new Error(`Failed to get bridge params: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Get bridge statistics
   * @returns Bridge statistics
   */
  async getBridgeStats(): Promise<BridgeStats> {
    try {
      const config = this.client.getConfig();
      if (!config.restEndpoint) {
        throw new Error('REST endpoint not configured');
      }

      const response = await axios.get(
        `${config.restEndpoint}/aura/bridge/v1beta1/stats`
      );

      return response.data.stats;
    } catch (error) {
      throw new Error(`Failed to get bridge stats: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Get bridge security information
   * @returns Bridge security configuration
   */
  async getBridgeSecurity(): Promise<BridgeSecurity> {
    try {
      const config = this.client.getConfig();
      if (!config.restEndpoint) {
        throw new Error('REST endpoint not configured');
      }

      const response = await axios.get(
        `${config.restEndpoint}/aura/bridge/v1beta1/security`
      );

      return response.data.security;
    } catch (error) {
      throw new Error(`Failed to get bridge security: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Helper method to parse bridge transfer from API response
   */
  private parseBridgeTransfer(data: any): BridgeTransfer {
    return {
      id: data.id,
      sender: data.sender,
      recipient: data.recipient,
      amount: data.amount,
      sourceChain: data.source_chain,
      targetChain: data.target_chain,
      status: data.status,
      createdAt: new Date(data.created_at),
      completedAt: data.completed_at ? new Date(data.completed_at) : undefined,
      proof: data.proof,
      txHash: data.tx_hash,
      error: data.error,
    };
  }

  /**
   * Helper to get first account address
   */
  private async getFirstAccount(): Promise<string> {
    const client = this.client.getSigningClient();
    const accounts = await client.getAccount(
      (await client.getAccount('')).address
    );
    if (!accounts) {
      throw new Error('No accounts found');
    }
    return accounts.address;
  }
}
