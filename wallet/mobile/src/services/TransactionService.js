/**
 * Transaction Service
 * Handles transaction signing and broadcasting for staking, governance, and other operations
 * Uses native crypto libraries instead of CosmJS for React Native compatibility
 */

import {sha256} from 'js-sha256';
import {ec as EC} from 'elliptic';
const {CHAIN_CONFIG, COIN, GAS_PRICE_TIERS} = require('../../../config/chain');
import PawAPI from './PawAPI';

const ec = new EC('secp256k1');
const ADDRESS_PREFIX = CHAIN_CONFIG.bech32Prefix || 'aura';
const VALOPER_PREFIX = `${ADDRESS_PREFIX}valoper`;
const FEE_DENOM = COIN.base || 'uaura';
const MAX_BASIS_POINTS = 10_000;
const DEFAULT_GAS_PRICE = GAS_PRICE_TIERS?.average ?? 0.025;

const buildFee = (gas = '200000', amount = '5000') => ({
  amount: [{denom: FEE_DENOM, amount}],
  gas,
});

const ensureAddress = (address, prefix, label) => {
  if (!address || typeof address !== 'string' || !address.startsWith(prefix)) {
    throw new Error(`Invalid ${label}`);
  }
  return address;
};

const ensureDenom = (denom, label) => {
  if (!denom || typeof denom !== 'string') {
    throw new Error(`${label} is required`);
  }
  return denom;
};

const ensureString = (value, label) => {
  if (!value || typeof value !== 'string') {
    throw new Error(`${label} is required`);
  }
  return value;
};

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
  } catch (error) {
    throw new Error(`${label} must be an integer amount`);
  }
};

const clampBasisPoints = (value = 0) => {
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

const encodeBytesBase64 = (value, label) => {
  if (value instanceof Uint8Array) {
    if (!value.length) {
      throw new Error(`${label} cannot be empty`);
    }
    return Buffer.from(value).toString('base64');
  }

  if (typeof value === 'string') {
    const normalized = value.startsWith('0x') ? value.slice(2) : value;
    if (!/^[0-9a-fA-F]+$/.test(normalized) || normalized.length === 0 || normalized.length % 2 !== 0) {
      throw new Error(`${label} must be a hex string`);
    }
    return Buffer.from(normalized, 'hex').toString('base64');
  }

  throw new Error(`${label} must be a Uint8Array or hex string`);
};

const buildCoin = (denom, amount, label = 'amount') => ({
  denom: ensureDenom(denom, 'denomination'),
  amount: normalizeAmount(amount, label),
});

class TransactionServiceClass {
  /**
   * Create a signed transaction
   * @param {Object} params - Transaction parameters
   * @returns {Promise<Object>} Signed transaction ready for broadcast
   */
  async createSignedTransaction({
    messages,
    fee,
    memo,
    accountNumber,
    sequence,
    chainId,
    privateKeyHex,
  }) {
    try {
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }
      if (!Array.isArray(messages) || messages.length === 0) {
        throw new Error('At least one message is required');
      }
      if (accountNumber === undefined || sequence === undefined) {
        throw new Error('Account number and sequence are required');
      }
      if (!chainId) {
        throw new Error('Chain ID is required');
      }

      // Create transaction body
      const txBody = {
        messages,
        memo: memo || '',
        timeout_height: '0',
        extension_options: [],
        non_critical_extension_options: [],
      };

      // Create auth info
      const authInfo = {
        signer_infos: [
          {
            public_key: {
              type_url: '/cosmos.crypto.secp256k1.PubKey',
              value: this.getPubKeyBytes(privateKeyHex),
            },
            mode_info: {
              single: {
                mode: 1, // SIGN_MODE_DIRECT
              },
            },
            sequence: sequence.toString(),
          },
        ],
        fee: {
          amount: fee.amount || [{denom: FEE_DENOM, amount: '5000'}],
          gas_limit: fee.gas || '200000',
          payer: '',
          granter: '',
        },
      };

      // Create sign doc
      const signDoc = {
        body_bytes: this.encodeTxBody(txBody),
        auth_info_bytes: this.encodeAuthInfo(authInfo),
        chain_id: chainId,
        account_number: accountNumber.toString(),
      };

      // Sign the transaction
      const signature = this.signTransaction(signDoc, privateKeyHex);

      // Create final transaction
      const tx = {
        body: txBody,
        auth_info: authInfo,
        signatures: [signature],
      };

      return tx;
    } catch (error) {
      console.error('Failed to create signed transaction:', error);
      throw error;
    }
  }

  /**
   * Helper to sign and broadcast a single message transaction
   */
  async broadcastSingleMessage({
    message,
    fee = buildFee(),
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    const tx = await this.createSignedTransaction({
      messages: Array.isArray(message) ? message : [message],
      fee,
      memo,
      accountNumber,
      sequence,
      chainId,
      privateKeyHex,
    });

    return PawAPI.broadcastTx(tx, 'sync');
  }

  /**
   * Get public key bytes from private key
   */
  getPubKeyBytes(privateKeyHex) {
    const keyPair = ec.keyFromPrivate(privateKeyHex, 'hex');
    const pubKey = keyPair.getPublic();
    return Buffer.from(pubKey.encode('array', true)).toString('base64');
  }

  /**
   * Encode transaction body (simplified version)
   */
  encodeTxBody(txBody) {
    // In a real implementation, use protobuf encoding
    // For now, return a simple JSON string representation
    return Buffer.from(JSON.stringify(txBody)).toString('base64');
  }

  /**
   * Encode auth info (simplified version)
   */
  encodeAuthInfo(authInfo) {
    // In a real implementation, use protobuf encoding
    return Buffer.from(JSON.stringify(authInfo)).toString('base64');
  }

  /**
   * Sign transaction
   */
  signTransaction(signDoc, privateKeyHex) {
    const keyPair = ec.keyFromPrivate(privateKeyHex, 'hex');

    // Create sign bytes
    const signBytes = Buffer.from(
      JSON.stringify({
        chain_id: signDoc.chain_id,
        account_number: signDoc.account_number,
        sequence: signDoc.sequence,
        fee: signDoc.fee,
        msgs: signDoc.messages,
        memo: signDoc.memo || '',
      }),
    );

    // Hash and sign
    const hash = sha256.array(signBytes);
    const signature = keyPair.sign(hash);

    // Return DER encoded signature
    return Buffer.from(signature.toDER()).toString('base64');
  }

  signDirect(signDoc, privateKeyHex) {
    if (!signDoc) {
      throw new Error('signDoc is required');
    }
    if (!privateKeyHex) {
      throw new Error('Private key is required');
    }
    const normalizedAccount = (signDoc.account_number ?? signDoc.accountNumber ?? '0').toString();
    const normalizedSequence = (signDoc.sequence ?? signDoc.auth_info?.signer_infos?.[0]?.sequence ?? '0').toString();
    const normalizedFee = signDoc.fee || signDoc.auth_info?.fee || {};
    const normalizedMsgs = signDoc.msgs || signDoc.body?.messages || [];
    const normalizedMemo = signDoc.memo || signDoc.body?.memo || '';
    const normalizedChainId = signDoc.chain_id || signDoc.chainId;

    const signature = this.signTransaction(
      {
        chain_id: normalizedChainId,
        account_number: normalizedAccount,
        sequence: normalizedSequence,
        fee: normalizedFee,
        messages: normalizedMsgs,
        memo: normalizedMemo,
      },
      privateKeyHex,
    );

    return {
      signed: {
        ...signDoc,
        account_number: normalizedAccount,
        sequence: normalizedSequence,
      },
      signature: {
        pub_key: {
          type: 'tendermint/PubKeySecp256k1',
          value: this.getPubKeyBytes(privateKeyHex),
        },
        signature,
      },
    };
  }

  signAmino(signDoc, privateKeyHex) {
    if (!signDoc) {
      throw new Error('signDoc is required');
    }
    if (!privateKeyHex) {
      throw new Error('Private key is required');
    }

    const aminoDoc = {
      chain_id: signDoc.chain_id || signDoc.chainId,
      account_number: (signDoc.account_number ?? signDoc.accountNumber ?? '0').toString(),
      sequence: (signDoc.sequence ?? '0').toString(),
      fee: signDoc.fee || {},
      msgs: signDoc.msgs || [],
      memo: signDoc.memo || '',
    };

    const signature = this.signTransaction(
      {
        chain_id: aminoDoc.chain_id,
        account_number: aminoDoc.account_number,
        sequence: aminoDoc.sequence,
        fee: aminoDoc.fee,
        messages: aminoDoc.msgs,
        memo: aminoDoc.memo,
      },
      privateKeyHex,
    );

    return {
      signed: aminoDoc,
      signature: {
        pub_key: {
          type: 'tendermint/PubKeySecp256k1',
          value: this.getPubKeyBytes(privateKeyHex),
        },
        signature,
      },
    };
  }

  /**
   * Undelegate tokens from a validator
   */
  async undelegate({
    delegatorAddress,
    validatorAddress,
    amount,
    denom,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      if (!delegatorAddress || !delegatorAddress.startsWith(ADDRESS_PREFIX)) {
        throw new Error('Invalid delegator address');
      }
      if (!validatorAddress || !validatorAddress.startsWith(VALOPER_PREFIX)) {
        throw new Error('Invalid validator address');
      }
      if (!amount || Number(amount) <= 0) {
        throw new Error('Undelegate amount must be greater than zero');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/cosmos.staking.v1beta1.MsgUndelegate',
        value: {
          delegator_address: delegatorAddress,
          validator_address: validatorAddress,
          amount: {denom, amount: amount.toString()},
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: FEE_DENOM, amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to undelegate:', error);
      throw new Error(error.message || 'Failed to undelegate tokens');
    }
  }

  /**
   * Withdraw delegation rewards
   */
  async withdrawRewards({
    delegatorAddress,
    validatorAddress,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      if (!delegatorAddress || !delegatorAddress.startsWith(ADDRESS_PREFIX)) {
        throw new Error('Invalid delegator address');
      }
      if (!validatorAddress || !validatorAddress.startsWith(VALOPER_PREFIX)) {
        throw new Error('Invalid validator address');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward',
        value: {
          delegator_address: delegatorAddress,
          validator_address: validatorAddress,
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: FEE_DENOM, amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to withdraw rewards:', error);
      throw new Error(error.message || 'Failed to withdraw rewards');
    }
  }

  /**
   * Redelegate tokens from one validator to another
   */
  async redelegate({
    delegatorAddress,
    srcValidatorAddress,
    dstValidatorAddress,
    amount,
    denom,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      if (!delegatorAddress || !delegatorAddress.startsWith(ADDRESS_PREFIX)) {
        throw new Error('Invalid delegator address');
      }
      if (
        !srcValidatorAddress ||
        !srcValidatorAddress.startsWith(VALOPER_PREFIX) ||
        !dstValidatorAddress ||
        !dstValidatorAddress.startsWith(VALOPER_PREFIX)
      ) {
        throw new Error('Invalid validator address');
      }
      if (srcValidatorAddress === dstValidatorAddress) {
        throw new Error('Source and destination validators must be different');
      }
      if (!amount || Number(amount) <= 0) {
        throw new Error('Redelegate amount must be greater than zero');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/cosmos.staking.v1beta1.MsgBeginRedelegate',
        value: {
          delegator_address: delegatorAddress,
          validator_src_address: srcValidatorAddress,
          validator_dst_address: dstValidatorAddress,
          amount: {denom, amount: amount.toString()},
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: FEE_DENOM, amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to redelegate:', error);
      throw new Error(error.message || 'Failed to redelegate tokens');
    }
  }

  /**
   * Vote on a governance proposal
   */
  async vote({
    voterAddress,
    proposalId,
    option,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      if (!voterAddress || !voterAddress.startsWith(ADDRESS_PREFIX)) {
        throw new Error('Invalid voter address');
      }
      if (!proposalId) {
        throw new Error('Proposal ID is required');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      // Map option string to VoteOption enum
      const voteOptionMap = {
        yes: 1,
        abstain: 2,
        no: 3,
        no_with_veto: 4,
      };

      const voteOption =
        typeof option === 'string'
          ? voteOptionMap[option.toLowerCase()]
          : option;

      if (!voteOption) {
        throw new Error(
          'Invalid vote option. Must be: yes, abstain, no, or no_with_veto',
        );
      }

      const message = {
        type_url: '/cosmos.gov.v1beta1.MsgVote',
        value: {
          proposal_id: proposalId.toString(),
          voter: voterAddress,
          option: voteOption,
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: FEE_DENOM, amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to vote:', error);
      throw new Error(error.message || 'Failed to submit vote');
    }
  }

  /**
   * Deposit tokens to a governance proposal
   */
  async deposit({
    depositorAddress,
    proposalId,
    amount,
    denom,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      if (!depositorAddress || !depositorAddress.startsWith(ADDRESS_PREFIX)) {
        throw new Error('Invalid depositor address');
      }
      if (!proposalId || Number(proposalId) <= 0) {
        throw new Error('Invalid proposal id');
      }
      if (!amount || Number(amount) <= 0) {
        throw new Error('Deposit amount must be greater than zero');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/cosmos.gov.v1beta1.MsgDeposit',
        value: {
          proposal_id: proposalId.toString(),
          depositor: depositorAddress,
          amount: [{denom, amount: amount.toString()}],
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: FEE_DENOM, amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to deposit to proposal:', error);
      throw new Error(error.message || 'Failed to deposit to proposal');
    }
  }

  /**
   * Delegate tokens to a validator
   */
  async delegate({
    delegatorAddress,
    validatorAddress,
    amount,
    denom,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      if (!delegatorAddress || !delegatorAddress.startsWith(ADDRESS_PREFIX)) {
        throw new Error('Invalid delegator address');
      }
      if (!validatorAddress || !validatorAddress.startsWith(VALOPER_PREFIX)) {
        throw new Error('Invalid validator address');
      }
      if (!amount || Number(amount) <= 0) {
        throw new Error('Delegate amount must be greater than zero');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/cosmos.staking.v1beta1.MsgDelegate',
        value: {
          delegator_address: delegatorAddress,
          validator_address: validatorAddress,
          amount: {denom, amount: amount.toString()},
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: FEE_DENOM, amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to delegate:', error);
      throw new Error(error.message || 'Failed to delegate tokens');
    }
  }

  /**
   * Send tokens
   */
  async sendTokens({
    fromAddress,
    toAddress,
    amount,
    denom,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      if (!fromAddress || !fromAddress.startsWith(ADDRESS_PREFIX)) {
        throw new Error('Invalid sender address');
      }
      if (!toAddress || !toAddress.startsWith(ADDRESS_PREFIX)) {
        throw new Error('Invalid recipient address');
      }
      if (!amount || Number(amount) <= 0) {
        throw new Error('Send amount must be greater than zero');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/cosmos.bank.v1beta1.MsgSend',
        value: {
          from_address: fromAddress,
          to_address: toAddress,
          amount: [{denom, amount: amount.toString()}],
        },
      };

      const tx = await this.createSignedTransaction({
        messages: [message],
        fee: {
          amount: [{denom: FEE_DENOM, amount: '5000'}],
          gas: '200000',
        },
        memo,
        accountNumber,
        sequence,
        chainId,
        privateKeyHex,
      });

      const result = await PawAPI.broadcastTx(tx, 'sync');
      return result;
    } catch (error) {
      console.error('Failed to send tokens:', error);
      throw new Error(error.message || 'Failed to send tokens');
    }
  }

  /**
   * Create a DEX liquidity pool
   */
  async createDexPool({
    creatorAddress,
    denomA,
    denomB,
    amountA,
    amountB,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      ensureAddress(creatorAddress, ADDRESS_PREFIX, 'creator address');
      const normalizedDenomA = ensureDenom(denomA, 'denomA');
      const normalizedDenomB = ensureDenom(denomB, 'denomB');
      if (normalizedDenomA === normalizedDenomB) {
        throw new Error('Pool denominations must be different');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/aura.dex.v1beta1.MsgCreatePool',
        value: {
          creator: creatorAddress,
          denom_a: normalizedDenomA,
          denom_b: normalizedDenomB,
          amount_a: buildCoin(normalizedDenomA, amountA, 'amountA'),
          amount_b: buildCoin(normalizedDenomB, amountB, 'amountB'),
        },
      };

      return await this.broadcastSingleMessage({
        message,
        fee: buildFee('500000', '25000'),
        memo,
        privateKeyHex,
        accountNumber,
        sequence,
        chainId,
      });
    } catch (error) {
      console.error('Failed to create DEX pool:', error);
      throw new Error(error.message || 'Failed to create DEX pool');
    }
  }

  /**
   * Add liquidity to an existing DEX pool
   */
  async addDexLiquidity({
    providerAddress,
    poolId,
    amountA,
    denomA,
    amountB,
    denomB,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      ensureAddress(providerAddress, ADDRESS_PREFIX, 'provider address');
      const poolIdentifier = poolId?.toString();
      if (!poolIdentifier) {
        throw new Error('poolId is required');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/aura.dex.v1beta1.MsgAddLiquidity',
        value: {
          provider: providerAddress,
          pool_id: poolIdentifier,
          amount_a: buildCoin(denomA, amountA, 'amountA'),
          amount_b: buildCoin(denomB, amountB, 'amountB'),
        },
      };

      return await this.broadcastSingleMessage({
        message,
        fee: buildFee('350000', '15000'),
        memo,
        privateKeyHex,
        accountNumber,
        sequence,
        chainId,
      });
    } catch (error) {
      console.error('Failed to add DEX liquidity:', error);
      throw new Error(error.message || 'Failed to add DEX liquidity');
    }
  }

  /**
   * Remove liquidity from a DEX pool
   */
  async removeDexLiquidity({
    providerAddress,
    poolId,
    lpTokens,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      ensureAddress(providerAddress, ADDRESS_PREFIX, 'provider address');
      const poolIdentifier = poolId?.toString();
      if (!poolIdentifier) {
        throw new Error('poolId is required');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/aura.dex.v1beta1.MsgRemoveLiquidity',
        value: {
          provider: providerAddress,
          pool_id: poolIdentifier,
          lp_tokens: normalizeAmount(lpTokens, 'lpTokens'),
        },
      };

      return await this.broadcastSingleMessage({
        message,
        fee: buildFee('320000', '12000'),
        memo,
        privateKeyHex,
        accountNumber,
        sequence,
        chainId,
      });
    } catch (error) {
      console.error('Failed to remove DEX liquidity:', error);
      throw new Error(error.message || 'Failed to remove DEX liquidity');
    }
  }

  /**
   * Execute an exact-in swap on a DEX pool
   */
  async swapDexExactIn({
    senderAddress,
    poolId,
    denomIn,
    amountIn,
    minAmountOut,
    maxSlippageBps = 50,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      ensureAddress(senderAddress, ADDRESS_PREFIX, 'sender address');
      const poolIdentifier = poolId?.toString();
      if (!poolIdentifier) {
        throw new Error('poolId is required');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/aura.dex.v1beta1.MsgSwapExactIn',
        value: {
          sender: senderAddress,
          pool_id: poolIdentifier,
          coin_in: buildCoin(denomIn, amountIn, 'amountIn'),
          min_amount_out: minAmountOut
            ? normalizeAmount(minAmountOut, 'minAmountOut')
            : normalizeAmount(amountIn, 'amountIn'),
          max_slippage_bps: clampBasisPoints(maxSlippageBps),
        },
      };

      return await this.broadcastSingleMessage({
        message,
        fee: buildFee('300000', '12000'),
        memo,
        privateKeyHex,
        accountNumber,
        sequence,
        chainId,
      });
    } catch (error) {
      console.error('Failed to execute DEX swap:', error);
      throw new Error(error.message || 'Failed to execute DEX swap');
    }
  }

  /**
   * Create an order on the DEX orderbook
   */
  async createDexOrder({
    creatorAddress,
    orderType,
    auraAmount,
    otherCoinDenom,
    otherAmount,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      ensureAddress(creatorAddress, ADDRESS_PREFIX, 'creator address');
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/aura.dex.v1beta1.MsgCreateOrder',
        value: {
          creator: creatorAddress,
          order_type: mapOrderType(orderType),
          aura_amount: normalizeAmount(auraAmount, 'auraAmount'),
          other_coin: ensureDenom(otherCoinDenom, 'other coin'),
          other_amount: normalizeAmount(otherAmount, 'otherAmount'),
        },
      };

      return await this.broadcastSingleMessage({
        message,
        fee: buildFee('250000', '10000'),
        memo,
        privateKeyHex,
        accountNumber,
        sequence,
        chainId,
      });
    } catch (error) {
      console.error('Failed to create DEX order:', error);
      throw new Error(error.message || 'Failed to create DEX order');
    }
  }

  /**
   * Cancel an existing DEX order
   */
  async cancelDexOrder({
    creatorAddress,
    orderId,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      ensureAddress(creatorAddress, ADDRESS_PREFIX, 'creator address');
      const orderIdentifier = orderId?.toString();
      if (!orderIdentifier) {
        throw new Error('orderId is required');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/aura.dex.v1beta1.MsgCancelOrder',
        value: {
          creator: creatorAddress,
          order_id: orderIdentifier,
        },
      };

      return await this.broadcastSingleMessage({
        message,
        fee: buildFee('160000', '8000'),
        memo,
        privateKeyHex,
        accountNumber,
        sequence,
        chainId,
      });
    } catch (error) {
      console.error('Failed to cancel DEX order:', error);
      throw new Error(error.message || 'Failed to cancel DEX order');
    }
  }

  /**
   * Commit to a DEX order hash (commit-reveal phase 1)
   */
  async commitDexOrder({
    senderAddress,
    commitHash,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      ensureAddress(senderAddress, ADDRESS_PREFIX, 'sender address');
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/aura.dex.v1beta1.MsgCommitOrder',
        value: {
          sender: senderAddress,
          commit_hash: encodeBytesBase64(commitHash, 'commit hash'),
        },
      };

      return await this.broadcastSingleMessage({
        message,
        fee: buildFee('150000', '7000'),
        memo,
        privateKeyHex,
        accountNumber,
        sequence,
        chainId,
      });
    } catch (error) {
      console.error('Failed to commit DEX order:', error);
      throw new Error(error.message || 'Failed to commit DEX order');
    }
  }

  /**
   * Reveal a previously committed DEX order (commit-reveal phase 2)
   */
  async revealDexOrder({
    senderAddress,
    commitId,
    orderType,
    auraAmount,
    otherCoinDenom,
    otherAmount,
    salt,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      ensureAddress(senderAddress, ADDRESS_PREFIX, 'sender address');
      const commitIdentifier = commitId?.toString();
      if (!commitIdentifier) {
        throw new Error('commitId is required');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/aura.dex.v1beta1.MsgRevealOrder',
        value: {
          sender: senderAddress,
          commit_id: commitIdentifier,
          order_type: mapOrderType(orderType),
          aura_amount: normalizeAmount(auraAmount, 'auraAmount'),
          other_coin: ensureDenom(otherCoinDenom, 'other coin'),
          other_amount: normalizeAmount(otherAmount, 'otherAmount'),
          salt: encodeBytesBase64(salt, 'salt'),
        },
      };

      return await this.broadcastSingleMessage({
        message,
        fee: buildFee('280000', '11000'),
        memo,
        privateKeyHex,
        accountNumber,
        sequence,
        chainId,
      });
    } catch (error) {
      console.error('Failed to reveal DEX order:', error);
      throw new Error(error.message || 'Failed to reveal DEX order');
    }
  }

  /**
   * Execute a matched DEX order (HTLC settlement)
   */
  async executeDexSwap({
    initiatorAddress,
    orderId,
    secret,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      ensureAddress(initiatorAddress, ADDRESS_PREFIX, 'initiator address');
      const orderIdentifier = orderId?.toString();
      if (!orderIdentifier) {
        throw new Error('orderId is required');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/aura.dex.v1beta1.MsgExecuteSwap',
        value: {
          initiator: initiatorAddress,
          order_id: orderIdentifier,
          secret: ensureString(secret, 'secret'),
        },
      };

      return await this.broadcastSingleMessage({
        message,
        fee: buildFee('220000', '12000'),
        memo,
        privateKeyHex,
        accountNumber,
        sequence,
        chainId,
      });
    } catch (error) {
      console.error('Failed to execute DEX swap:', error);
      throw new Error(error.message || 'Failed to execute DEX swap');
    }
  }

  /**
   * Create a DEX HTLC
   */
  async createDexHTLC({
    senderAddress,
    recipientAddress,
    denom,
    amount,
    secretHash,
    timelockDuration,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      ensureAddress(senderAddress, ADDRESS_PREFIX, 'sender address');
      ensureAddress(recipientAddress, ADDRESS_PREFIX, 'recipient address');
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/aura.dex.v1beta1.MsgCreateHTLC',
        value: {
          sender: senderAddress,
          recipient: recipientAddress,
          amount: buildCoin(denom, amount, 'amount'),
          secret_hash: ensureString(secretHash, 'secret hash'),
          timelock_duration: normalizeAmount(
            timelockDuration,
            'timelockDuration',
          ),
        },
      };

      return await this.broadcastSingleMessage({
        message,
        fee: buildFee('320000', '14000'),
        memo,
        privateKeyHex,
        accountNumber,
        sequence,
        chainId,
      });
    } catch (error) {
      console.error('Failed to create DEX HTLC:', error);
      throw new Error(error.message || 'Failed to create DEX HTLC');
    }
  }

  /**
   * Claim a DEX HTLC
   */
  async claimDexHTLC({
    recipientAddress,
    htlcId,
    secret,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      ensureAddress(recipientAddress, ADDRESS_PREFIX, 'recipient address');
      const htlcIdentifier = htlcId?.toString();
      if (!htlcIdentifier) {
        throw new Error('htlcId is required');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/aura.dex.v1beta1.MsgClaimHTLC',
        value: {
          recipient: recipientAddress,
          htlc_id: htlcIdentifier,
          secret: ensureString(secret, 'secret'),
        },
      };

      return await this.broadcastSingleMessage({
        message,
        fee: buildFee('200000', '9000'),
        memo,
        privateKeyHex,
        accountNumber,
        sequence,
        chainId,
      });
    } catch (error) {
      console.error('Failed to claim DEX HTLC:', error);
      throw new Error(error.message || 'Failed to claim DEX HTLC');
    }
  }

  /**
   * Refund an expired DEX HTLC
   */
  async refundDexHTLC({
    senderAddress,
    htlcId,
    memo,
    privateKeyHex,
    accountNumber,
    sequence,
    chainId,
  }) {
    try {
      ensureAddress(senderAddress, ADDRESS_PREFIX, 'sender address');
      const htlcIdentifier = htlcId?.toString();
      if (!htlcIdentifier) {
        throw new Error('htlcId is required');
      }
      if (!privateKeyHex) {
        throw new Error('Private key is required');
      }

      const message = {
        type_url: '/aura.dex.v1beta1.MsgRefundHTLC',
        value: {
          sender: senderAddress,
          htlc_id: htlcIdentifier,
        },
      };

      return await this.broadcastSingleMessage({
        message,
        fee: buildFee('200000', '9000'),
        memo,
        privateKeyHex,
        accountNumber,
        sequence,
        chainId,
      });
    } catch (error) {
      console.error('Failed to refund DEX HTLC:', error);
      throw new Error(error.message || 'Failed to refund DEX HTLC');
    }
  }
}

// Export singleton instance
const TransactionService = new TransactionServiceClass();
export default TransactionService;
