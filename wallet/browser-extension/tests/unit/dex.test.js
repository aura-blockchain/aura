import { describe, it, expect, beforeEach, vi } from 'vitest';

describe('DexModule', () => {
  let DexModule;
  const samplePool = {
    pool_id: '1',
    denom_a: 'uaura',
    denom_b: 'usdt',
    reserve_a: '1000000',
    reserve_b: '500000',
    total_lp_tokens: '100000000',
    fee_percentage: '0.003',
    protocol_fee_percentage: '0.0005',
  };

  beforeEach(async () => {
    global.COSMOS_SDK = {
      config: {
        bech32Prefix: 'aura',
        coinDenom: 'uaura',
        restEndpoint: 'http://localhost:1317',
      },
      queryPools: vi.fn().mockResolvedValue([samplePool]),
      queryPool: vi.fn().mockResolvedValue(samplePool),
      queryOraclePrices: vi.fn().mockResolvedValue([{ denom: 'uaura', price: '1.00' }]),
    };

    const module = await import('../../src/dex.js');
    DexModule = module.default || module;
  });

  it('normalizes and quotes pools correctly', () => {
    const formatted = DexModule.formatPool(samplePool);
    expect(formatted.poolId).toBe('1');
    expect(formatted.priceAB).toBeGreaterThan(0);
    expect(formatted.depth.uaura).toBe('1000000');

    const quote = DexModule.calculateSwapQuote(samplePool, 'uaura', 100000);
    expect(quote.denomOut).toBe('usdt');
    expect(quote.expectedAmountOut).toBe('45309');
    expect(quote.minAmountOut).toBe('45083');
    expect(quote.maxSlippageBps).toBe(50);
    expect(quote.priceImpactPercent).toBeCloseTo(9.06, 2);
  });

  it('builds swap and liquidity transactions', () => {
    const swapTx = DexModule.buildSwapExactInTx({
      sender: 'aura1senderaddress0000000000000000000000',
      poolId: '1',
      denomIn: 'uaura',
      amountIn: 100000,
      minAmountOut: 45000,
      maxSlippageBps: 75,
    });

    expect(swapTx.body.messages[0]['@type']).toBe('/aura.dex.v1beta1.MsgSwapExactIn');
    expect(swapTx.body.messages[0].max_slippage_bps).toBe(75);
    expect(swapTx.body.messages[0].coin_in.denom).toBe('uaura');

    const addLiquidityTx = DexModule.buildAddLiquidityTx({
      provider: 'aura1provider000000000000000000000000',
      poolId: '1',
      amountA: 500000,
      amountB: 250000,
      denomA: 'uaura',
      denomB: 'usdt',
    });

    expect(addLiquidityTx.body.messages[0]['@type']).toBe('/aura.dex.v1beta1.MsgAddLiquidity');
    expect(addLiquidityTx.body.messages[0].amount_a.amount).toBe('500000');

    const removeLiquidityTx = DexModule.buildRemoveLiquidityTx({
      provider: 'aura1provider000000000000000000000000',
      poolId: '1',
      lpTokens: '100000',
    });

    expect(removeLiquidityTx.body.messages[0]['@type']).toBe('/aura.dex.v1beta1.MsgRemoveLiquidity');
    expect(removeLiquidityTx.body.messages[0].lp_tokens).toBe('100000');
  });

  it('builds orderbook and HTLC transactions', () => {
    const createOrderTx = DexModule.buildCreateOrderTx({
      creator: 'aura1creator0000000000000000000000000',
      orderType: 1,
      auraAmount: 1000000,
      otherCoinDenom: 'usdt',
      otherAmount: 500,
    });

    expect(createOrderTx.body.messages[0]['@type']).toBe('/aura.dex.v1beta1.MsgCreateOrder');
    expect(createOrderTx.body.messages[0].order_type).toBe(1);

    const cancelOrderTx = DexModule.buildCancelOrderTx({
      creator: 'aura1creator0000000000000000000000000',
      orderId: '42',
    });

    expect(cancelOrderTx.body.messages[0]['@type']).toBe('/aura.dex.v1beta1.MsgCancelOrder');
    expect(cancelOrderTx.body.messages[0].order_id).toBe('42');

    const createHtlcTx = DexModule.buildCreateHtlcTx({
      sender: 'aura1sender000000000000000000000000000',
      recipient: 'aura1recipient000000000000000000000',
      denom: 'uaura',
      amount: 750000,
      secretHash: 'ABC123',
      timelockDuration: 3600,
    });

    expect(createHtlcTx.body.messages[0]['@type']).toBe('/aura.dex.v1beta1.MsgCreateHTLC');
    expect(createHtlcTx.body.messages[0].timelock_duration).toBe(3600);

    const claimHtlcTx = DexModule.buildClaimHtlcTx({
      recipient: 'aura1recipient000000000000000000000',
      htlcId: 'htlc-1',
      secret: 'secret',
    });

    expect(claimHtlcTx.body.messages[0]['@type']).toBe('/aura.dex.v1beta1.MsgClaimHTLC');

    const refundHtlcTx = DexModule.buildRefundHtlcTx({
      sender: 'aura1sender000000000000000000000000000',
      htlcId: 'htlc-1',
    });

    expect(refundHtlcTx.body.messages[0]['@type']).toBe('/aura.dex.v1beta1.MsgRefundHTLC');
  });

  it('estimates pool share and wraps query helpers', async () => {
    const share = DexModule.estimatePoolShare('2500000', '100000000');
    expect(share).toBeCloseTo(2.5, 1);

    await DexModule.queryPools();
    await DexModule.queryPool('1');
    await DexModule.queryOraclePrices();

    expect(COSMOS_SDK.queryPools).toHaveBeenCalled();
    expect(COSMOS_SDK.queryPool).toHaveBeenCalledWith('1');
    expect(COSMOS_SDK.queryOraclePrices).toHaveBeenCalled();
  });

  it('throws on invalid params', () => {
    expect(() => DexModule.calculateSwapQuote(samplePool, 'notadenom', 10)).toThrow('Input denom not part of pool');
    expect(() =>
      DexModule.buildAddLiquidityTx({
        provider: 'invalid-address',
        poolId: '1',
        amountA: 1,
        amountB: 1,
      })
    ).toThrow('provider must be a valid aura address');
  });
});
