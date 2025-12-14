import axios from 'axios';
import { SigningStargateClient, GasPrice } from '@cosmjs/stargate';
import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';

export class ApiService {
  constructor() {
    this.apiEndpoint = this.getApiEndpoint();
  }

  getApiEndpoint() {
    if (window.electron?.store) {
      return window.electron.store.get('apiEndpoint').then(endpoint =>
        endpoint || 'http://localhost:1317'
      );
    }
    return Promise.resolve('http://localhost:1317');
  }

  async getEndpoint() {
    return await this.apiEndpoint;
  }

  /**
   * Get account balance
   */
  async getBalance(address) {
    try {
      const endpoint = await this.getEndpoint();
      const response = await axios.get(
        `${endpoint}/cosmos/bank/v1beta1/balances/${address}`
      );
      return response.data;
    } catch (error) {
      console.error('Failed to get balance:', error);
      throw new Error(error.response?.data?.message || 'Failed to fetch balance');
    }
  }

  /**
   * Get account information (sequence, account number)
   */
  async getAccount(address) {
    try {
      const endpoint = await this.getEndpoint();
      const response = await axios.get(
        `${endpoint}/cosmos/auth/v1beta1/accounts/${address}`
      );
      return response.data.account;
    } catch (error) {
      console.error('Failed to get account:', error);
      throw new Error(error.response?.data?.message || 'Failed to fetch account');
    }
  }

  /**
   * Get transaction history
   */
  async getTransactions(address, limit = 50) {
    try {
      const endpoint = await this.getEndpoint();

      // Try to get transactions from the tx search endpoint
      const response = await axios.get(
        `${endpoint}/cosmos/tx/v1beta1/txs`,
        {
          params: {
            events: `message.sender='${address}'`,
            'pagination.limit': limit,
            order_by: 'ORDER_BY_DESC'
          }
        }
      );

      return response.data.txs || response.data.tx_responses || [];
    } catch (error) {
      console.error('Failed to get transactions:', error);
      // Return empty array if endpoint doesn't exist or fails
      return [];
    }
  }

  /**
   * Send tokens to another address
   */
  async sendTokens(fromAddress, toAddress, amount, denom, memo, privateKey) {
    try {
      const endpoint = await this.getEndpoint();
      const rpcEndpoint = endpoint.replace('1317', '26657').replace('/cosmos', '');

      // Create wallet from private key
      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: 'aura'
      });

      // Get signing client
      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString('0.025uaura')
        }
      );

      // Send transaction
      const result = await client.sendTokens(
        fromAddress,
        toAddress,
        [{ denom, amount: amount.toString() }],
        {
          amount: [{ denom: 'uaura', amount: '5000' }],
          gas: '200000'
        },
        memo
      );

      return result;
    } catch (error) {
      console.error('Failed to send tokens:', error);
      throw new Error(error.message || 'Failed to send transaction');
    }
  }

  /**
   * Get validator list
   */
  async getValidators() {
    try {
      const endpoint = await this.getEndpoint();
      const response = await axios.get(
        `${endpoint}/cosmos/staking/v1beta1/validators`
      );
      return response.data.validators || [];
    } catch (error) {
      console.error('Failed to get validators:', error);
      throw new Error(error.response?.data?.message || 'Failed to fetch validators');
    }
  }

  /**
   * Delegate tokens to a validator
   */
  async delegate(delegatorAddress, validatorAddress, amount, denom, memo, privateKey) {
    try {
      const endpoint = await this.getEndpoint();
      const rpcEndpoint = endpoint.replace('1317', '26657').replace('/cosmos', '');

      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: 'aura'
      });

      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString('0.025uaura')
        }
      );

      const result = await client.delegateTokens(
        delegatorAddress,
        validatorAddress,
        { denom, amount: amount.toString() },
        {
          amount: [{ denom: 'uaura', amount: '5000' }],
          gas: '200000'
        },
        memo
      );

      return result;
    } catch (error) {
      console.error('Failed to delegate:', error);
      throw new Error(error.message || 'Failed to delegate tokens');
    }
  }

  /**
   * Get node information
   */
  async getNodeInfo() {
    try {
      const endpoint = await this.getEndpoint();
      const response = await axios.get(
        `${endpoint}/cosmos/base/tendermint/v1beta1/node_info`
      );
      return response.data.node_info;
    } catch (error) {
      console.error('Failed to get node info:', error);
      throw new Error(error.response?.data?.message || 'Failed to fetch node info');
    }
  }

  /**
   * Get latest block
   */
  async getLatestBlock() {
    try {
      const endpoint = await this.getEndpoint();
      const response = await axios.get(
        `${endpoint}/cosmos/base/tendermint/v1beta1/blocks/latest`
      );
      return response.data.block;
    } catch (error) {
      console.error('Failed to get latest block:', error);
      throw new Error(error.response?.data?.message || 'Failed to fetch latest block');
    }
  }

  /**
   * Undelegate tokens from a validator
   */
  async undelegate(delegatorAddress, validatorAddress, amount, denom, memo, privateKey) {
    try {
      const endpoint = await this.getEndpoint();
      const rpcEndpoint = endpoint.replace('1317', '26657').replace('/cosmos', '');

      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: 'aura'
      });

      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString('0.025uaura')
        }
      );

      const result = await client.undelegateTokens(
        delegatorAddress,
        validatorAddress,
        { denom, amount: amount.toString() },
        {
          amount: [{ denom: 'uaura', amount: '5000' }],
          gas: '200000'
        },
        memo
      );

      return result;
    } catch (error) {
      console.error('Failed to undelegate:', error);
      throw new Error(error.message || 'Failed to undelegate tokens');
    }
  }

  /**
   * Withdraw delegation rewards from a validator
   */
  async withdrawRewards(delegatorAddress, validatorAddress, memo, privateKey) {
    try {
      const endpoint = await this.getEndpoint();
      const rpcEndpoint = endpoint.replace('1317', '26657').replace('/cosmos', '');

      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: 'aura'
      });

      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString('0.025uaura')
        }
      );

      const result = await client.withdrawRewards(
        delegatorAddress,
        validatorAddress,
        {
          amount: [{ denom: 'uaura', amount: '5000' }],
          gas: '200000'
        },
        memo
      );

      return result;
    } catch (error) {
      console.error('Failed to withdraw rewards:', error);
      throw new Error(error.message || 'Failed to withdraw rewards');
    }
  }

  /**
   * Redelegate tokens from one validator to another
   */
  async redelegate(delegatorAddress, srcValidatorAddress, dstValidatorAddress, amount, denom, memo, privateKey) {
    try {
      const endpoint = await this.getEndpoint();
      const rpcEndpoint = endpoint.replace('1317', '26657').replace('/cosmos', '');

      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: 'aura'
      });

      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString('0.025uaura')
        }
      );

      const msgRedelegate = {
        typeUrl: '/cosmos.staking.v1beta1.MsgBeginRedelegate',
        value: {
          delegatorAddress: delegatorAddress,
          validatorSrcAddress: srcValidatorAddress,
          validatorDstAddress: dstValidatorAddress,
          amount: { denom, amount: amount.toString() }
        }
      };

      const result = await client.signAndBroadcast(
        delegatorAddress,
        [msgRedelegate],
        {
          amount: [{ denom: 'uaura', amount: '5000' }],
          gas: '200000'
        },
        memo
      );

      return result;
    } catch (error) {
      console.error('Failed to redelegate:', error);
      throw new Error(error.message || 'Failed to redelegate tokens');
    }
  }

  /**
   * Vote on a governance proposal
   */
  async vote(voterAddress, proposalId, option, memo, privateKey) {
    try {
      const endpoint = await this.getEndpoint();
      const rpcEndpoint = endpoint.replace('1317', '26657').replace('/cosmos', '');

      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: 'aura'
      });

      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString('0.025uaura')
        }
      );

      // Map option string to VoteOption enum
      const voteOptionMap = {
        'yes': 1,
        'abstain': 2,
        'no': 3,
        'no_with_veto': 4
      };

      const voteOption = typeof option === 'string'
        ? voteOptionMap[option.toLowerCase()]
        : option;

      if (!voteOption) {
        throw new Error('Invalid vote option. Must be: yes, abstain, no, or no_with_veto');
      }

      const msgVote = {
        typeUrl: '/cosmos.gov.v1beta1.MsgVote',
        value: {
          proposalId: BigInt(proposalId),
          voter: voterAddress,
          option: voteOption
        }
      };

      const result = await client.signAndBroadcast(
        voterAddress,
        [msgVote],
        {
          amount: [{ denom: 'uaura', amount: '5000' }],
          gas: '200000'
        },
        memo
      );

      return result;
    } catch (error) {
      console.error('Failed to vote:', error);
      throw new Error(error.message || 'Failed to submit vote');
    }
  }

  /**
   * Get delegations for an address
   */
  async getDelegations(address) {
    try {
      const endpoint = await this.getEndpoint();
      const response = await axios.get(
        `${endpoint}/cosmos/staking/v1beta1/delegations/${address}`
      );
      return response.data.delegation_responses || [];
    } catch (error) {
      console.error('Failed to get delegations:', error);
      throw new Error(error.response?.data?.message || 'Failed to fetch delegations');
    }
  }

  /**
   * Get delegation rewards
   */
  async getRewards(address) {
    try {
      const endpoint = await this.getEndpoint();
      const response = await axios.get(
        `${endpoint}/cosmos/distribution/v1beta1/delegators/${address}/rewards`
      );
      return response.data;
    } catch (error) {
      console.error('Failed to get rewards:', error);
      throw new Error(error.response?.data?.message || 'Failed to fetch rewards');
    }
  }

  /**
   * Get unbonding delegations
   */
  async getUnbondingDelegations(address) {
    try {
      const endpoint = await this.getEndpoint();
      const response = await axios.get(
        `${endpoint}/cosmos/staking/v1beta1/delegators/${address}/unbonding_delegations`
      );
      return response.data.unbonding_responses || [];
    } catch (error) {
      console.error('Failed to get unbonding delegations:', error);
      throw new Error(error.response?.data?.message || 'Failed to fetch unbonding delegations');
    }
  }

  /**
   * Get governance proposals
   */
  async getProposals(status = '') {
    try {
      const endpoint = await this.getEndpoint();
      const params = status ? { proposal_status: status } : {};
      const response = await axios.get(
        `${endpoint}/cosmos/gov/v1beta1/proposals`,
        { params }
      );
      return response.data.proposals || [];
    } catch (error) {
      console.error('Failed to get proposals:', error);
      throw new Error(error.response?.data?.message || 'Failed to fetch proposals');
    }
  }

  /**
   * Get specific governance proposal
   */
  async getProposal(proposalId) {
    try {
      const endpoint = await this.getEndpoint();
      const response = await axios.get(
        `${endpoint}/cosmos/gov/v1beta1/proposals/${proposalId}`
      );
      return response.data.proposal;
    } catch (error) {
      console.error('Failed to get proposal:', error);
      throw new Error(error.response?.data?.message || 'Failed to fetch proposal');
    }
  }

  /**
   * Get DEX pools (Aura-specific)
   */
  async getDexPools() {
    try {
      const endpoint = await this.getEndpoint();
      const response = await axios.get(`${endpoint}/aura/dex/v1/pools`);
      return response.data.pools || [];
    } catch (error) {
      console.error('Failed to get DEX pools:', error);
      // Return empty array if DEX module not available
      return [];
    }
  }

  /**
   * Get oracle prices (Aura-specific)
   */
  async getOraclePrices() {
    try {
      const endpoint = await this.getEndpoint();
      const response = await axios.get(`${endpoint}/aura/oracle/v1/prices`);
      return response.data.prices || [];
    } catch (error) {
      console.error('Failed to get oracle prices:', error);
      // Return empty array if oracle module not available
      return [];
    }
  }
}
