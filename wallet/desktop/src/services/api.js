import axios from 'axios';
import { SigningStargateClient, GasPrice, defaultRegistryTypes } from '@cosmjs/stargate';
import { DirectSecp256k1HdWallet, Registry } from '@cosmjs/proto-signing';
import { DEX_TYPE_REGISTRY } from './dexMessages';
const { CHAIN_CONFIG, COIN, GAS_PRICE_TIERS, REST_ENDPOINTS } = require('../../../config/chain');

const FEE_DENOM = COIN.base || 'uaura';
const MAX_BASIS_POINTS = 10_000;
const DEFAULT_API = REST_ENDPOINTS[0]?.address || 'http://localhost:1317';
const RPC_FROM_REST = (host) => host.replace('1317', '26657').replace('/cosmos', '');
const DEFAULT_GAS_PRICE = `${GAS_PRICE_TIERS.average ?? 0.025}${FEE_DENOM}`;
let BECH32_PREFIX = CHAIN_CONFIG?.bech32Prefix || 'aura';
let SLIP44 = CHAIN_CONFIG?.slip44 ?? 118;

const buildFee = (gas = '200000', amount = '5000') => ({
  amount: [{ denom: FEE_DENOM, amount }],
  gas
});

const normalizeAmount = (value, label) => {
  if (value === undefined || value === null || value === '') {
    throw new Error(`${label} is required`);
  }

  try {
    const normalized = BigInt(value.toString());
    if (normalized <= 0n) {
      throw new Error(`${label} must be greater than zero`);
    }
    return normalized.toString();
  } catch (err) {
    throw new Error(`${label} must be an integer amount`);
  }
};

const assertDenom = (denom, label) => {
  if (!denom || typeof denom !== 'string') {
    throw new Error(`${label} is required`);
  }
  return denom;
};

const assertAddress = (address, prefix, label) => {
  if (!address || !address.startsWith(prefix)) {
    throw new Error(`Invalid ${label}`);
  }
  return address;
};

const clampBasisPoints = (value) => {
  const numeric = Number(value);
  if (Number.isNaN(numeric)) {
    throw new Error('maxSlippageBps must be numeric');
  }
  if (numeric < 0 || numeric > MAX_BASIS_POINTS) {
    throw new Error('maxSlippageBps must be between 0 and 10000');
  }
  return Math.floor(numeric);
};

const mapOrderType = (orderType) => {
  if (typeof orderType === 'number' && (orderType === 0 || orderType === 1)) {
    return orderType;
  }

  if (typeof orderType === 'string') {
    const normalized = orderType.toLowerCase();
    if (normalized === 'buy') {
      return 0;
    }
    if (normalized === 'sell') {
      return 1;
    }
  }

  throw new Error('Invalid order type. Use "buy" or "sell"');
};

const buildCoin = (denom, amount) => ({
  denom: assertDenom(denom, 'denomination'),
  amount: normalizeAmount(amount, 'amount')
});

const assertString = (value, label) => {
  if (!value || typeof value !== 'string') {
    throw new Error(`${label} is required`);
  }
  return value;
};

const ensurePrivateKey = (privateKey) => {
  if (!privateKey) {
    throw new Error('Missing signing key');
  }
  return privateKey;
};

const hexToBytes = (value, label) => {
  const hex = value.startsWith('0x') ? value.slice(2) : value;
  if (!/^[0-9a-fA-F]+$/.test(hex) || hex.length === 0 || hex.length % 2 !== 0) {
    throw new Error(`${label} must be a hex string`);
  }
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < hex.length; i += 2) {
    bytes[i / 2] = parseInt(hex.substring(i, i + 2), 16);
  }
  return bytes;
};

const normalizeBytes = (value, label) => {
  if (value instanceof Uint8Array) {
    if (!value.length) {
      throw new Error(`${label} cannot be empty`);
    }
    return value;
  }

  if (typeof value === 'string') {
    return hexToBytes(value.trim(), label);
  }

  throw new Error(`${label} must be a Uint8Array or hex string`);
};

export class ApiService {
  constructor() {
    this.apiEndpoint = this.getApiEndpoint();
  }

  setBech32Prefix(prefix) {
    if (prefix) {
      BECH32_PREFIX = prefix;
    }
  }

  setSlip44(slip44) {
    if (Number.isInteger(slip44)) {
      SLIP44 = slip44;
    }
  }

  getApiEndpoint() {
    if (window.electron?.store) {
      return window.electron.store.get('apiEndpoint').then(endpoint =>
        endpoint || DEFAULT_API
      );
    }
    return Promise.resolve(DEFAULT_API);
  }

  async getEndpoint() {
    return await this.apiEndpoint;
  }

  getDexRegistry() {
    if (!this.dexRegistry) {
      const registry = new Registry(defaultRegistryTypes);
      if (Array.isArray(DEX_TYPE_REGISTRY)) {
        DEX_TYPE_REGISTRY.forEach(([typeUrl, mod]) => {
          registry.register(typeUrl, mod);
        });
      }
      this.dexRegistry = registry;
    }
    return this.dexRegistry;
  }

  async broadcastDexTx(signerAddress, messages, memo, privateKey, fee = buildFee()) {
    assertAddress(signerAddress, BECH32_PREFIX, 'signer address');
    ensurePrivateKey(privateKey);

    const endpoint = await this.getEndpoint();
    const rpcEndpoint = RPC_FROM_REST(endpoint);

    const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
      prefix: BECH32_PREFIX
    });

    const client = await SigningStargateClient.connectWithSigner(
      rpcEndpoint,
      wallet,
      {
        gasPrice: GasPrice.fromString(DEFAULT_GAS_PRICE),
        registry: this.getDexRegistry()
      }
    );

    return client.signAndBroadcast(
      signerAddress,
      messages,
      fee,
      memo
    );
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
      if (!fromAddress || !fromAddress.startsWith(CHAIN_CONFIG.bech32Prefix)) {
        throw new Error('Invalid sender address');
      }
      if (!toAddress || !toAddress.startsWith(CHAIN_CONFIG.bech32Prefix)) {
        throw new Error('Invalid recipient address');
      }
      if (!amount || Number(amount) <= 0) {
        throw new Error('Amount must be greater than zero');
      }
      if (!privateKey) {
        throw new Error('Missing signing key');
      }

      const endpoint = await this.getEndpoint();
      const rpcEndpoint = RPC_FROM_REST(endpoint);

      // Create wallet from private key
      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: CHAIN_CONFIG.bech32Prefix
      });

      // Get signing client
      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString(DEFAULT_GAS_PRICE)
        }
      );

      // Send transaction
      const result = await client.sendTokens(
        fromAddress,
        toAddress,
        [{ denom, amount: amount.toString() }],
        {
          amount: [{ denom: FEE_DENOM, amount: '5000' }],
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
      if (!delegatorAddress || !delegatorAddress.startsWith(CHAIN_CONFIG.bech32Prefix)) {
        throw new Error('Invalid delegator address');
      }
      if (!validatorAddress || !validatorAddress.startsWith('auravaloper')) {
        throw new Error('Invalid validator address');
      }
      if (!amount || Number(amount) <= 0) {
        throw new Error('Delegate amount must be greater than zero');
      }
      if (!privateKey) {
        throw new Error('Missing signing key');
      }

      const endpoint = await this.getEndpoint();
      const rpcEndpoint = RPC_FROM_REST(endpoint);

      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: CHAIN_CONFIG.bech32Prefix
      });

      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString(DEFAULT_GAS_PRICE)
        }
      );

      const result = await client.delegateTokens(
        delegatorAddress,
        validatorAddress,
        { denom, amount: amount.toString() },
        {
          amount: [{ denom: FEE_DENOM, amount: '5000' }],
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
      if (!delegatorAddress || !delegatorAddress.startsWith(CHAIN_CONFIG.bech32Prefix)) {
        throw new Error('Invalid delegator address');
      }
      if (!validatorAddress || !validatorAddress.startsWith('auravaloper')) {
        throw new Error('Invalid validator address');
      }
      if (!amount || Number(amount) <= 0) {
        throw new Error('Undelegate amount must be greater than zero');
      }
      if (!privateKey) {
        throw new Error('Missing signing key');
      }

      const endpoint = await this.getEndpoint();
      const rpcEndpoint = RPC_FROM_REST(endpoint);

      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: CHAIN_CONFIG.bech32Prefix
      });

      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString(DEFAULT_GAS_PRICE)
        }
      );

      const result = await client.undelegateTokens(
        delegatorAddress,
        validatorAddress,
        { denom, amount: amount.toString() },
        {
          amount: [{ denom: FEE_DENOM, amount: '5000' }],
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
      if (!delegatorAddress || !delegatorAddress.startsWith(CHAIN_CONFIG.bech32Prefix)) {
        throw new Error('Invalid delegator address');
      }
      if (!validatorAddress || !validatorAddress.startsWith('auravaloper')) {
        throw new Error('Invalid validator address');
      }
      if (!privateKey) {
        throw new Error('Missing signing key');
      }

      const endpoint = await this.getEndpoint();
      const rpcEndpoint = RPC_FROM_REST(endpoint);

      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: CHAIN_CONFIG.bech32Prefix
      });

      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString(DEFAULT_GAS_PRICE)
        }
      );

      const result = await client.withdrawRewards(
        delegatorAddress,
        validatorAddress,
        {
          amount: [{ denom: FEE_DENOM, amount: '5000' }],
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
      if (!delegatorAddress || !delegatorAddress.startsWith('aura')) {
        throw new Error('Invalid delegator address');
      }
      if (!srcValidatorAddress || !srcValidatorAddress.startsWith('auravaloper') ||
        !dstValidatorAddress || !dstValidatorAddress.startsWith('auravaloper')) {
        throw new Error('Invalid validator address');
      }
      if (srcValidatorAddress === dstValidatorAddress) {
        throw new Error('Source and destination validators must be different');
      }
      if (!amount || Number(amount) <= 0) {
        throw new Error('Redelegate amount must be greater than zero');
      }
      if (!privateKey) {
        throw new Error('Missing signing key');
      }

      const endpoint = await this.getEndpoint();
      const rpcEndpoint = RPC_FROM_REST(endpoint);

      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: CHAIN_CONFIG.bech32Prefix
      });

      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString(DEFAULT_GAS_PRICE)
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
          amount: [{ denom: FEE_DENOM, amount: '5000' }],
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
      if (!voterAddress || !voterAddress.startsWith(CHAIN_CONFIG.bech32Prefix)) {
        throw new Error('Invalid voter address');
      }
      if (!proposalId || Number(proposalId) <= 0) {
        throw new Error('Invalid proposal id');
      }
      if (!privateKey) {
        throw new Error('Missing signing key');
      }

      const endpoint = await this.getEndpoint();
      const rpcEndpoint = RPC_FROM_REST(endpoint);

      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: CHAIN_CONFIG.bech32Prefix
      });

      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString(DEFAULT_GAS_PRICE)
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
          amount: [{ denom: FEE_DENOM, amount: '5000' }],
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
   * Deposit tokens to a governance proposal
   */
  async deposit(depositorAddress, proposalId, amount, denom, memo, privateKey) {
    try {
      if (!depositorAddress || !depositorAddress.startsWith(CHAIN_CONFIG.bech32Prefix)) {
        throw new Error('Invalid depositor address');
      }
      if (!proposalId || Number(proposalId) <= 0) {
        throw new Error('Invalid proposal id');
      }
      if (!amount || Number(amount) <= 0) {
        throw new Error('Deposit amount must be greater than zero');
      }
      if (!privateKey) {
        throw new Error('Missing signing key');
      }

      const endpoint = await this.getEndpoint();
      const rpcEndpoint = RPC_FROM_REST(endpoint);

      const wallet = await DirectSecp256k1HdWallet.fromMnemonic(privateKey, {
        prefix: CHAIN_CONFIG.bech32Prefix
      });

      const client = await SigningStargateClient.connectWithSigner(
        rpcEndpoint,
        wallet,
        {
          gasPrice: GasPrice.fromString(DEFAULT_GAS_PRICE)
        }
      );

      const msgDeposit = {
        typeUrl: '/cosmos.gov.v1beta1.MsgDeposit',
        value: {
          proposalId: BigInt(proposalId),
          depositor: depositorAddress,
          amount: [{ denom, amount: amount.toString() }]
        }
      };

      const result = await client.signAndBroadcast(
        depositorAddress,
        [msgDeposit],
        {
          amount: [{ denom: FEE_DENOM, amount: '5000' }],
          gas: '200000'
        },
        memo
      );

      return result;
    } catch (error) {
      console.error('Failed to deposit to proposal:', error);
      throw new Error(error.message || 'Failed to deposit to proposal');
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

  /**
   * Create a new DEX liquidity pool
  */
  async createDexPool(creatorAddress, denomA, denomB, amountA, amountB, memo = '', privateKey) {
    try {
      assertAddress(creatorAddress, BECH32_PREFIX, 'creator address');
      if (denomA === denomB) {
        throw new Error('Pool denominations must be different');
      }

      const msg = {
        typeUrl: '/aura.dex.v1beta1.MsgCreatePool',
        value: {
          creator: creatorAddress,
          denomA: assertDenom(denomA, 'denomA'),
          denomB: assertDenom(denomB, 'denomB'),
          amountA: buildCoin(denomA, amountA),
          amountB: buildCoin(denomB, amountB)
        }
      };

      return await this.broadcastDexTx(
        creatorAddress,
        [msg],
        memo,
        privateKey,
        buildFee('500000', '25000')
      );
    } catch (error) {
      console.error('Failed to create DEX pool:', error);
      throw new Error(error.message || 'Failed to create DEX pool');
    }
  }

  /**
   * Provide liquidity to an existing DEX pool
   */
  async addDexLiquidity(providerAddress, poolId, amountA, denomA, amountB, denomB, memo = '', privateKey) {
    try {
      assertAddress(providerAddress, BECH32_PREFIX, 'provider address');
      const poolIdentifier = poolId?.toString();
      if (!poolIdentifier) {
        throw new Error('poolId is required');
      }

      const msg = {
        typeUrl: '/aura.dex.v1beta1.MsgAddLiquidity',
        value: {
          provider: providerAddress,
          poolId: poolIdentifier,
          amountA: buildCoin(denomA, amountA),
          amountB: buildCoin(denomB, amountB)
        }
      };

      return await this.broadcastDexTx(
        providerAddress,
        [msg],
        memo,
        privateKey,
        buildFee('350000', '15000')
      );
    } catch (error) {
      console.error('Failed to add DEX liquidity:', error);
      throw new Error(error.message || 'Failed to add DEX liquidity');
    }
  }

  /**
   * Remove liquidity from a DEX pool
   */
  async removeDexLiquidity(providerAddress, poolId, lpTokens, memo = '', privateKey) {
    try {
      assertAddress(providerAddress, BECH32_PREFIX, 'provider address');
      const poolIdentifier = poolId?.toString();
      if (!poolIdentifier) {
        throw new Error('poolId is required');
      }

      const msg = {
        typeUrl: '/aura.dex.v1beta1.MsgRemoveLiquidity',
        value: {
          provider: providerAddress,
          poolId: poolIdentifier,
          lpTokens: normalizeAmount(lpTokens, 'lpTokens')
        }
      };

      return await this.broadcastDexTx(
        providerAddress,
        [msg],
        memo,
        privateKey,
        buildFee('320000', '12000')
      );
    } catch (error) {
      console.error('Failed to remove DEX liquidity:', error);
      throw new Error(error.message || 'Failed to remove DEX liquidity');
    }
  }

  /**
   * Execute a swap within a DEX pool
   */
  async swapDexExactIn(senderAddress, poolId, denomIn, amountIn, minAmountOut, maxSlippageBps = 50, memo = '', privateKey) {
    try {
      assertAddress(senderAddress, BECH32_PREFIX, 'sender address');
      const poolIdentifier = poolId?.toString();
      if (!poolIdentifier) {
        throw new Error('poolId is required');
      }

      const msg = {
        typeUrl: '/aura.dex.v1beta1.MsgSwapExactIn',
        value: {
          sender: senderAddress,
          poolId: poolIdentifier,
          coinIn: buildCoin(denomIn, amountIn),
          minAmountOut: minAmountOut
            ? normalizeAmount(minAmountOut, 'minAmountOut')
            : normalizeAmount(amountIn, 'amountIn'),
          maxSlippageBps: clampBasisPoints(maxSlippageBps)
        }
      };

      return await this.broadcastDexTx(
        senderAddress,
        [msg],
        memo,
        privateKey,
        buildFee('300000', '12000')
      );
    } catch (error) {
      console.error('Failed to execute swap:', error);
      throw new Error(error.message || 'Failed to execute swap');
    }
  }

  /**
   * Create a DEX order (orderbook)
   */
  async createDexOrder(creatorAddress, orderType, auraAmount, otherCoinDenom, otherAmount, memo = '', privateKey) {
    try {
      assertAddress(creatorAddress, BECH32_PREFIX, 'creator address');

      const msg = {
        typeUrl: '/aura.dex.v1beta1.MsgCreateOrder',
        value: {
          creator: creatorAddress,
          orderType: mapOrderType(orderType),
          auraAmount: normalizeAmount(auraAmount, 'auraAmount'),
          otherCoin: assertDenom(otherCoinDenom, 'other coin'),
          otherAmount: normalizeAmount(otherAmount, 'otherAmount')
        }
      };

      return await this.broadcastDexTx(
        creatorAddress,
        [msg],
        memo,
        privateKey,
        buildFee('250000', '10000')
      );
    } catch (error) {
      console.error('Failed to create DEX order:', error);
      throw new Error(error.message || 'Failed to create DEX order');
    }
  }

  /**
   * Cancel an orderbook order
   */
  async cancelDexOrder(creatorAddress, orderId, memo = '', privateKey) {
    try {
      assertAddress(creatorAddress, BECH32_PREFIX, 'creator address');
      const orderIdentifier = orderId?.toString();
      if (!orderIdentifier) {
        throw new Error('orderId is required');
      }

      const msg = {
        typeUrl: '/aura.dex.v1beta1.MsgCancelOrder',
        value: {
          creator: creatorAddress,
          orderId: orderIdentifier
        }
      };

      return await this.broadcastDexTx(
        creatorAddress,
        [msg],
        memo,
        privateKey,
        buildFee('160000', '8000')
      );
    } catch (error) {
      console.error('Failed to cancel DEX order:', error);
      throw new Error(error.message || 'Failed to cancel DEX order');
    }
  }

  /**
   * Commit to an order hash for anti-frontrunning
   */
  async commitDexOrder(senderAddress, commitHash, memo = '', privateKey) {
    try {
      assertAddress(senderAddress, BECH32_PREFIX, 'sender address');

      const msg = {
        typeUrl: '/aura.dex.v1beta1.MsgCommitOrder',
        value: {
          sender: senderAddress,
          commitHash: normalizeBytes(commitHash, 'commit hash')
        }
      };

      return await this.broadcastDexTx(
        senderAddress,
        [msg],
        memo,
        privateKey,
        buildFee('150000', '7000')
      );
    } catch (error) {
      console.error('Failed to commit DEX order:', error);
      throw new Error(error.message || 'Failed to commit DEX order');
    }
  }

  /**
   * Reveal an order commitment
   */
  async revealDexOrder(senderAddress, commitId, orderType, auraAmount, otherCoinDenom, otherAmount, salt, memo = '', privateKey) {
    try {
      assertAddress(senderAddress, BECH32_PREFIX, 'sender address');
      const commitIdentifier = commitId?.toString();
      if (!commitIdentifier) {
        throw new Error('commitId is required');
      }

      const msg = {
        typeUrl: '/aura.dex.v1beta1.MsgRevealOrder',
        value: {
          sender: senderAddress,
          commitId: commitIdentifier,
          orderType: mapOrderType(orderType),
          auraAmount: normalizeAmount(auraAmount, 'auraAmount'),
          otherCoin: assertDenom(otherCoinDenom, 'other coin'),
          otherAmount: normalizeAmount(otherAmount, 'otherAmount'),
          salt: normalizeBytes(salt, 'salt')
        }
      };

      return await this.broadcastDexTx(
        senderAddress,
        [msg],
        memo,
        privateKey,
        buildFee('280000', '11000')
      );
    } catch (error) {
      console.error('Failed to reveal DEX order:', error);
      throw new Error(error.message || 'Failed to reveal DEX order');
    }
  }

  /**
   * Execute an HTLC-backed swap after matching
   */
  async executeDexSwap(initiatorAddress, orderId, secret, memo = '', privateKey) {
    try {
      assertAddress(initiatorAddress, BECH32_PREFIX, 'initiator address');
      const orderIdentifier = orderId?.toString();
      if (!orderIdentifier) {
        throw new Error('orderId is required');
      }

      const msg = {
        typeUrl: '/aura.dex.v1beta1.MsgExecuteSwap',
        value: {
          initiator: initiatorAddress,
          orderId: orderIdentifier,
          secret: assertString(secret, 'secret')
        }
      };

      return await this.broadcastDexTx(
        initiatorAddress,
        [msg],
        memo,
        privateKey,
        buildFee('220000', '12000')
      );
    } catch (error) {
      console.error('Failed to execute DEX swap:', error);
      throw new Error(error.message || 'Failed to execute DEX swap');
    }
  }

  /**
   * Create a DEX HTLC
   */
  async createDexHTLC(senderAddress, recipientAddress, denom, amount, secretHash, timelockDuration, memo = '', privateKey) {
    try {
      assertAddress(senderAddress, BECH32_PREFIX, 'sender address');
      assertAddress(recipientAddress, BECH32_PREFIX, 'recipient address');
      const duration = BigInt(timelockDuration ?? 0);
      if (duration <= 0n) {
        throw new Error('timelockDuration must be greater than zero');
      }

      const msg = {
        typeUrl: '/aura.dex.v1beta1.MsgCreateHTLC',
        value: {
          sender: senderAddress,
          recipient: recipientAddress,
          amount: buildCoin(denom, amount),
          secretHash: assertString(secretHash, 'secret hash'),
          timelockDuration: duration
        }
      };

      return await this.broadcastDexTx(
        senderAddress,
        [msg],
        memo,
        privateKey,
        buildFee('320000', '14000')
      );
    } catch (error) {
      console.error('Failed to create DEX HTLC:', error);
      throw new Error(error.message || 'Failed to create DEX HTLC');
    }
  }

  /**
   * Claim an HTLC with the secret
   */
  async claimDexHTLC(recipientAddress, htlcId, secret, memo = '', privateKey) {
    try {
      assertAddress(recipientAddress, BECH32_PREFIX, 'recipient address');
      const htlcIdentifier = htlcId?.toString();
      if (!htlcIdentifier) {
        throw new Error('htlcId is required');
      }

      const msg = {
        typeUrl: '/aura.dex.v1beta1.MsgClaimHTLC',
        value: {
          recipient: recipientAddress,
          htlcId: htlcIdentifier,
          secret: assertString(secret, 'secret')
        }
      };

      return await this.broadcastDexTx(
        recipientAddress,
        [msg],
        memo,
        privateKey,
        buildFee('200000', '9000')
      );
    } catch (error) {
      console.error('Failed to claim DEX HTLC:', error);
      throw new Error(error.message || 'Failed to claim DEX HTLC');
    }
  }

  /**
   * Refund an expired HTLC
   */
  async refundDexHTLC(senderAddress, htlcId, memo = '', privateKey) {
    try {
      assertAddress(senderAddress, BECH32_PREFIX, 'sender address');
      const htlcIdentifier = htlcId?.toString();
      if (!htlcIdentifier) {
        throw new Error('htlcId is required');
      }

      const msg = {
        typeUrl: '/aura.dex.v1beta1.MsgRefundHTLC',
        value: {
          sender: senderAddress,
          htlcId: htlcIdentifier
        }
      };

      return await this.broadcastDexTx(
        senderAddress,
        [msg],
        memo,
        privateKey,
        buildFee('200000', '9000')
      );
    } catch (error) {
      console.error('Failed to refund DEX HTLC:', error);
      throw new Error(error.message || 'Failed to refund DEX HTLC');
    }
  }
}
