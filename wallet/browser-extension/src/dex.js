/**
 * DEX Module
 * Provides helpers for Aura AMM + orderbook actions from the browser wallet.
 */

const BASIS_POINTS = 10_000;

const DexModule = {
  BASIS_POINTS,

  /**
   * Normalize pool data into numeric helpers used by quoting functions.
   * @param {Object} pool - Pool details from REST.
   * @returns {Object} Normalized pool data.
   */
  normalizePool(pool) {
    if (!pool) {
      throw new Error('Pool data is required');
    }

    const reserveA = BigInt(pool.reserve_a ?? '0');
    const reserveB = BigInt(pool.reserve_b ?? '0');
    const feePercent = parseFloat(pool.fee_percentage ?? '0');
    const protocolFeePercent = parseFloat(pool.protocol_fee_percentage ?? '0');

    return {
      poolId: pool.pool_id?.toString() ?? '',
      denomA: pool.denom_a,
      denomB: pool.denom_b,
      reserveA,
      reserveB,
      feePercent,
      protocolFeePercent,
      totalLpTokens: BigInt(pool.total_lp_tokens ?? '0'),
    };
  },

  /**
   * Format pool with derived metrics (price, depth, share).
   * @param {Object} pool - Pool information.
   * @returns {Object} Formatted pool metrics.
   */
  formatPool(pool) {
    const normalized = this.normalizePool(pool);
    const { reserveA, reserveB } = normalized;

    const priceAB = reserveA > 0n ? Number(reserveB) / Number(reserveA) : 0;
    const priceBA = reserveB > 0n ? Number(reserveA) / Number(reserveB) : 0;

    return {
      ...normalized,
      priceAB,
      priceBA,
      depth: {
        [normalized.denomA]: reserveA.toString(),
        [normalized.denomB]: reserveB.toString(),
      },
    };
  },

  /**
   * Estimate swap results using a constant-product formula.
   * @param {Object} pool - Pool details.
   * @param {string} denomIn - Input denom.
   * @param {string|number|bigint} amountIn - Amount of input denom.
   * @param {number} slippageBps - Allowed slippage in basis points.
   * @returns {Object} Quote result with min amount out + price impact.
   */
  calculateSwapQuote(pool, denomIn, amountIn, slippageBps = 50) {
    const normalized = this.normalizePool(pool);
    const inputAmount = this.toBigInt(amountIn, 'amountIn');

    if (inputAmount <= 0n) {
      throw new Error('amountIn must be positive');
    }

    const feeBps = Math.round(normalized.feePercent * BASIS_POINTS);
    const protocolFeeBps = Math.round(normalized.protocolFeePercent * BASIS_POINTS);
    const effectiveBps = BASIS_POINTS - feeBps - protocolFeeBps;

    if (effectiveBps <= 0) {
      throw new Error('Invalid fee configuration');
    }

    let reserveIn;
    let reserveOut;
    let denomOut;

    if (denomIn === normalized.denomA) {
      reserveIn = normalized.reserveA;
      reserveOut = normalized.reserveB;
      denomOut = normalized.denomB;
    } else if (denomIn === normalized.denomB) {
      reserveIn = normalized.reserveB;
      reserveOut = normalized.reserveA;
      denomOut = normalized.denomA;
    } else {
      throw new Error('Input denom not part of pool');
    }

    if (reserveIn <= 0n || reserveOut <= 0n) {
      throw new Error('Pool has insufficient liquidity');
    }

    const amountInAfterFee = (inputAmount * BigInt(effectiveBps)) / BigInt(BASIS_POINTS);
    const numerator = amountInAfterFee * reserveOut;
    const denominator = reserveIn + amountInAfterFee;
    const expectedAmountOut = denominator === 0n ? 0n : numerator / denominator;

    const slippage = BigInt(slippageBps);
    let minAmountOut = expectedAmountOut - ((expectedAmountOut * slippage) / BigInt(BASIS_POINTS));
    if (minAmountOut < 0n) {
      minAmountOut = 0n;
    }

    const feePaid = inputAmount - amountInAfterFee;
    const priceImpact = Number(amountInAfterFee) / Number(reserveIn + amountInAfterFee);

    return {
      poolId: normalized.poolId,
      denomIn,
      denomOut,
      expectedAmountOut: expectedAmountOut.toString(),
      minAmountOut: minAmountOut.toString(),
      feeAmount: feePaid.toString(),
      priceImpactPercent: Number((priceImpact || 0) * 100),
      maxSlippageBps: Number(slippage),
    };
  },

  /**
   * Estimate pool share percentage from LP tokens.
   * @param {string|number|bigint} userLpTokens - LP tokens user holds.
   * @param {string|number|bigint} totalLpTokens - Total LP tokens in pool.
   * @returns {number} Share percentage (0-100).
   */
  estimatePoolShare(userLpTokens, totalLpTokens) {
    const user = this.toBigInt(userLpTokens, 'userLpTokens');
    const total = this.toBigInt(totalLpTokens, 'totalLpTokens');

    if (total === 0n) {
      return 0;
    }

    return Number((user * 10000n) / total) / 100;
  },

  /**
   * Build MsgSwapExactIn transaction.
   */
  buildSwapExactInTx(params) {
    const {
      sender,
      poolId,
      denomIn,
      amountIn,
      minAmountOut,
      maxSlippageBps = 50,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '12000' }], gas: '300000' },
    } = params;

    this.assertAddress(sender, 'sender');
    this.assertPositive(amountIn, 'amountIn');
    this.assertPositive(minAmountOut ?? amountIn, 'minAmountOut');

    return this.buildTxWrapper({
      memo,
      fee,
      typeUrl: '/aura.dex.v1beta1.MsgSwapExactIn',
      value: {
        sender,
        pool_id: poolId?.toString(),
        coin_in: this.normalizeCoin(denomIn, amountIn),
        min_amount_out: minAmountOut?.toString() ?? amountIn.toString(),
        max_slippage_bps: Number(maxSlippageBps),
      },
    });
  },

  /**
   * Build MsgAddLiquidity transaction.
   */
  buildAddLiquidityTx(params) {
    const {
      provider,
      poolId,
      amountA,
      amountB,
      denomA,
      denomB,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '15000' }], gas: '350000' },
    } = params;

    this.assertAddress(provider, 'provider');
    this.assertPositive(amountA, 'amountA');
    this.assertPositive(amountB, 'amountB');

    return this.buildTxWrapper({
      memo,
      fee,
      typeUrl: '/aura.dex.v1beta1.MsgAddLiquidity',
      value: {
        provider,
        pool_id: poolId?.toString(),
        amount_a: this.normalizeCoin(denomA ?? COSMOS_SDK.config.coinDenom, amountA),
        amount_b: this.normalizeCoin(denomB, amountB),
      },
    });
  },

  /**
   * Build MsgRemoveLiquidity transaction.
   */
  buildRemoveLiquidityTx(params) {
    const {
      provider,
      poolId,
      lpTokens,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '12000' }], gas: '300000' },
    } = params;

    this.assertAddress(provider, 'provider');
    this.assertPositive(lpTokens, 'lpTokens');

    return this.buildTxWrapper({
      memo,
      fee,
      typeUrl: '/aura.dex.v1beta1.MsgRemoveLiquidity',
      value: {
        provider,
        pool_id: poolId?.toString(),
        lp_tokens: lpTokens.toString(),
      },
    });
  },

  /**
   * Build MsgCreatePool transaction.
   */
  buildCreatePoolTx(params) {
    const {
      creator,
      denomA,
      denomB,
      amountA,
      amountB,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '25000' }], gas: '500000' },
    } = params;

    this.assertAddress(creator, 'creator');
    this.assertPositive(amountA, 'amountA');
    this.assertPositive(amountB, 'amountB');

    return this.buildTxWrapper({
      memo,
      fee,
      typeUrl: '/aura.dex.v1beta1.MsgCreatePool',
      value: {
        creator,
        denom_a: denomA ?? COSMOS_SDK.config.coinDenom,
        denom_b: denomB,
        amount_a: this.normalizeCoin(denomA ?? COSMOS_SDK.config.coinDenom, amountA),
        amount_b: this.normalizeCoin(denomB, amountB),
      },
    });
  },

  /**
   * Build MsgCreateOrder transaction for the orderbook.
   */
  buildCreateOrderTx(params) {
    const {
      creator,
      orderType = 0,
      auraAmount,
      otherCoinDenom,
      otherAmount,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '10000' }], gas: '250000' },
    } = params;

    this.assertAddress(creator, 'creator');
    this.assertPositive(auraAmount, 'auraAmount');
    this.assertPositive(otherAmount, 'otherAmount');

    return this.buildTxWrapper({
      memo,
      fee,
      typeUrl: '/aura.dex.v1beta1.MsgCreateOrder',
      value: {
        creator,
        order_type: Number(orderType),
        aura_amount: auraAmount.toString(),
        other_coin: otherCoinDenom,
        other_amount: otherAmount.toString(),
      },
    });
  },

  /**
   * Build MsgCancelOrder transaction.
   */
  buildCancelOrderTx(params) {
    const {
      creator,
      orderId,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '8000' }], gas: '160000' },
    } = params;

    this.assertAddress(creator, 'creator');

    return this.buildTxWrapper({
      memo,
      fee,
      typeUrl: '/aura.dex.v1beta1.MsgCancelOrder',
      value: {
        creator,
        order_id: orderId?.toString(),
      },
    });
  },

  /**
   * Build MsgCreateHTLC transaction.
   */
  buildCreateHtlcTx(params) {
    const {
      sender,
      recipient,
      denom,
      amount,
      secretHash,
      timelockDuration,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '14000' }], gas: '320000' },
    } = params;

    this.assertAddress(sender, 'sender');
    this.assertAddress(recipient, 'recipient');
    this.assertPositive(amount, 'amount');

    return this.buildTxWrapper({
      memo,
      fee,
      typeUrl: '/aura.dex.v1beta1.MsgCreateHTLC',
      value: {
        sender,
        recipient,
        amount: this.normalizeCoin(denom, amount),
        secret_hash: secretHash,
        timelock_duration: Number(timelockDuration ?? 0),
      },
    });
  },

  /**
   * Build MsgClaimHTLC transaction.
   */
  buildClaimHtlcTx(params) {
    const {
      recipient,
      htlcId,
      secret,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '9000' }], gas: '200000' },
    } = params;

    this.assertAddress(recipient, 'recipient');

    return this.buildTxWrapper({
      memo,
      fee,
      typeUrl: '/aura.dex.v1beta1.MsgClaimHTLC',
      value: {
        recipient,
        htlc_id: htlcId?.toString(),
        secret,
      },
    });
  },

  /**
   * Build MsgRefundHTLC transaction.
   */
  buildRefundHtlcTx(params) {
    const {
      sender,
      htlcId,
      memo = '',
      fee = { amount: [{ denom: 'uaura', amount: '9000' }], gas: '200000' },
    } = params;

    this.assertAddress(sender, 'sender');

    return this.buildTxWrapper({
      memo,
      fee,
      typeUrl: '/aura.dex.v1beta1.MsgRefundHTLC',
      value: {
        sender,
        htlc_id: htlcId?.toString(),
      },
    });
  },

  /**
   * Query helpers (wrap COSMOS_SDK for consistency).
   */
  async queryPools() {
    return COSMOS_SDK.queryPools();
  },

  async queryPool(poolId) {
    return COSMOS_SDK.queryPool(poolId);
  },

  async queryOraclePrices() {
    return COSMOS_SDK.queryOraclePrices();
  },

  /**
   * Internal: wrap a message shard into a transaction skeleton.
   */
  buildTxWrapper({ memo, fee, typeUrl, value }) {
    return {
      body: {
        messages: [{
          '@type': typeUrl,
          ...value,
        }],
        memo,
        timeout_height: '0',
        extension_options: [],
        non_critical_extension_options: [],
      },
      auth_info: {
        signer_infos: [],
        fee,
      },
      signatures: [],
    };
  },

  normalizeCoin(denom = COSMOS_SDK.config.coinDenom, amount) {
    this.assertPositive(amount, 'amount');
    if (!denom) {
      throw new Error('Denom is required');
    }
    return { denom, amount: amount.toString() };
  },

  toBigInt(value, field) {
    try {
      return BigInt(value ?? 0);
    } catch (error) {
      throw new Error(`${field} must be an integer-like value`);
    }
  },

  assertAddress(value, field) {
    if (!value || typeof value !== 'string' || !value.startsWith(COSMOS_SDK.config.bech32Prefix)) {
      throw new Error(`${field} must be a valid ${COSMOS_SDK.config.bech32Prefix} address`);
    }
  },

  assertPositive(value, field) {
    if (this.toBigInt(value, field) <= 0n) {
      throw new Error(`${field} must be positive`);
    }
  },
};

export default DexModule;

if (typeof module !== 'undefined' && module.exports) {
  module.exports = DexModule;
}
