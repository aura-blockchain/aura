/**
 * Aura Blockchain API Service
 * Provides REST API client for Cosmos SDK endpoints
 * Handles balance queries, transactions, staking, DEX, and oracle data
 */

import axios from 'axios';

class PawAPIService {
  constructor() {
    this.baseURL = 'http://localhost:1317'; // Default to local testnet
    this.client = axios.create({
      baseURL: this.baseURL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }

  /**
   * Set base URL for API
   * @param {string} url - Base URL
   */
  setBaseURL(url) {
    this.baseURL = url;
    this.client.defaults.baseURL = url;
  }

  /**
   * Handle API errors
   * @param {Error} error - Axios error
   * @throws {Error} Formatted error
   */
  handleError(error) {
    if (error.response) {
      const message = error.response.data?.message || 'API Error';
      throw new Error(`API Error: ${message}`);
    } else if (error.request) {
      throw new Error('Network Error: No response from server');
    } else {
      throw new Error('Request error: ' + error.message);
    }
  }

  /**
   * Get node info
   * @returns {Promise<Object>} Node information
   */
  async getNodeInfo() {
    try {
      const response = await this.client.get(
        '/cosmos/base/tendermint/v1beta1/node_info',
      );
      return response.data;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get account balance
   * @param {string} address - Account address
   * @returns {Promise<Object>} Balance data
   */
  async getBalance(address) {
    try {
      const response = await this.client.get(
        `/cosmos/bank/v1beta1/balances/${address}`,
      );
      return response.data;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get account information
   * @param {string} address - Account address
   * @returns {Promise<Object>} Account data (sequence, account_number)
   */
  async getAccount(address) {
    try {
      const response = await this.client.get(
        `/cosmos/auth/v1beta1/accounts/${address}`,
      );
      return response.data.account;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get transaction by hash
   * @param {string} hash - Transaction hash
   * @returns {Promise<Object>} Transaction data
   */
  async getTransaction(hash) {
    try {
      const response = await this.client.get(`/cosmos/tx/v1beta1/txs/${hash}`);
      return response.data;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get transactions by address
   * @param {string} address - Account address
   * @param {number} limit - Max transactions to return
   * @returns {Promise<Array>} Transaction array
   */
  async getTransactionsByAddress(address, limit = 20) {
    try {
      const response = await this.client.get('/cosmos/tx/v1beta1/txs', {
        params: {
          events: `message.sender='${address}'`,
          'pagination.limit': limit,
          order_by: 'ORDER_BY_DESC',
        },
      });
      return response.data.txs || [];
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Broadcast transaction
   * @param {Object} txBytes - Signed transaction bytes
   * @param {string} mode - Broadcast mode (sync/async/block)
   * @returns {Promise<Object>} Broadcast result
   */
  async broadcastTx(txBytes, mode = 'sync') {
    try {
      const response = await this.client.post('/cosmos/tx/v1beta1/txs', {
        tx_bytes: txBytes,
        mode: `BROADCAST_MODE_${mode.toUpperCase()}`,
      });
      return response.data;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Simulate transaction to estimate gas
   * @param {Object} tx - Transaction object
   * @returns {Promise<Object>} Gas estimate
   */
  async simulateTx(tx) {
    try {
      const response = await this.client.post('/cosmos/tx/v1beta1/simulate', {
        tx,
      });
      return response.data.gas_info;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get DEX pools
   * @returns {Promise<Array>} Pool array
   */
  async getDexPools() {
    try {
      const response = await this.client.get('/aura/dex/v1/pools');
      return response.data.pools || [];
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get specific DEX pool
   * @param {string} poolId - Pool ID
   * @returns {Promise<Object>} Pool data
   */
  async getPool(poolId) {
    try {
      const response = await this.client.get(`/aura/dex/v1/pools/${poolId}`);
      return response.data.pool;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get pool liquidity
   * @param {string} poolId - Pool ID
   * @returns {Promise<Object>} Liquidity data
   */
  async getPoolLiquidity(poolId) {
    try {
      const response = await this.client.get(
        `/aura/dex/v1/pools/${poolId}/liquidity`,
      );
      return response.data;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get oracle prices
   * @returns {Promise<Array>} Price array
   */
  async getOraclePrices() {
    try {
      const response = await this.client.get('/aura/oracle/v1/prices');
      return response.data.prices || [];
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get specific oracle price
   * @param {string} symbol - Price symbol (e.g., 'Aura/USD')
   * @returns {Promise<Object>} Price data
   */
  async getOraclePrice(symbol) {
    try {
      const response = await this.client.get(
        `/aura/oracle/v1/prices/${symbol}`,
      );
      return response.data.price;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get validators
   * @param {string} status - Validator status filter
   * @returns {Promise<Array>} Validator array
   */
  async getValidators(status = 'BOND_STATUS_BONDED') {
    try {
      const response = await this.client.get('/cosmos/staking/v1beta1/validators', {
        params: {status},
      });
      return response.data.validators || [];
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get specific validator
   * @param {string} validatorAddr - Validator operator address
   * @returns {Promise<Object>} Validator data
   */
  async getValidator(validatorAddr) {
    try {
      const response = await this.client.get(
        `/cosmos/staking/v1beta1/validators/${validatorAddr}`,
      );
      return response.data.validator;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get delegations for an address
   * @param {string} address - Delegator address
   * @returns {Promise<Array>} Delegation array
   */
  async getDelegations(address) {
    try {
      const response = await this.client.get(
        `/cosmos/staking/v1beta1/delegations/${address}`,
      );
      return response.data.delegation_responses || [];
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get delegation rewards
   * @param {string} address - Delegator address
   * @returns {Promise<Object>} Rewards data
   */
  async getRewards(address) {
    try {
      const response = await this.client.get(
        `/cosmos/distribution/v1beta1/delegators/${address}/rewards`,
      );
      return response.data;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get unbonding delegations
   * @param {string} address - Delegator address
   * @returns {Promise<Array>} Unbonding delegation array
   */
  async getUnbondingDelegations(address) {
    try {
      const response = await this.client.get(
        `/cosmos/staking/v1beta1/delegators/${address}/unbonding_delegations`,
      );
      return response.data.unbonding_responses || [];
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get supply information
   * @returns {Promise<Object>} Supply data
   */
  async getSupply() {
    try {
      const response = await this.client.get('/cosmos/bank/v1beta1/supply');
      return response.data;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get staking pool
   * @returns {Promise<Object>} Pool data
   */
  async getStakingPool() {
    try {
      const response = await this.client.get('/cosmos/staking/v1beta1/pool');
      return response.data.pool;
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get governance proposals
   * @param {string} status - Proposal status filter
   * @returns {Promise<Array>} Proposal array
   */
  async getProposals(status = '') {
    try {
      const params = status ? {proposal_status: status} : {};
      const response = await this.client.get('/cosmos/gov/v1beta1/proposals', {
        params,
      });
      return response.data.proposals || [];
    } catch (error) {
      this.handleError(error);
    }
  }

  /**
   * Get specific proposal
   * @param {string} proposalId - Proposal ID
   * @returns {Promise<Object>} Proposal data
   */
  async getProposal(proposalId) {
    try {
      const response = await this.client.get(
        `/cosmos/gov/v1beta1/proposals/${proposalId}`,
      );
      return response.data.proposal;
    } catch (error) {
      this.handleError(error);
    }
  }
}

// Export singleton instance
const PawAPI = new PawAPIService();
export default PawAPI;
