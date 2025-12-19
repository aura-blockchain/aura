import { buildKeplrChainInfo } from './keplr';
import DexModule from './dex';
const { CHAIN_CONFIG, COIN } = require('../../config/chain');

export async function ensureKeplrChain() {
  if (!window.keplr) {
    throw new Error('Keplr/Leap not available');
  }
  const chainInfo = buildKeplrChainInfo();
  await window.keplr.experimentalSuggestChain(chainInfo);
  return chainInfo;
}

export async function getKeplrOfflineSigner() {
  const chainInfo = await ensureKeplrChain();
  await window.keplr.enable(chainInfo.chainId);
  return window.getOfflineSignerAuto(chainInfo.chainId);
}

export function buildKeplrProvider() {
  return {
    async getKey() {
      const chainInfo = buildKeplrChainInfo();
      await window.keplr.enable(chainInfo.chainId);
      const key = await window.keplr.getKey(chainInfo.chainId);
      return key;
    },
    async signAmino(signer, signDoc) {
      const chainInfo = buildKeplrChainInfo();
      await window.keplr.enable(chainInfo.chainId);
      return window.keplr.signAmino(chainInfo.chainId, signer, signDoc);
    },
    async signDirect(signer, signDoc) {
      const chainInfo = buildKeplrChainInfo();
      await window.keplr.enable(chainInfo.chainId);
      return window.keplr.signDirect(chainInfo.chainId, signer, signDoc);
    },
    async sendTx(txBytes, mode = 'sync') {
      const rpc = CHAIN_CONFIG.rpc?.[0]?.address || 'http://localhost:26657';
      const res = await fetch(`${rpc}/cosmos/tx/v1beta1/txs`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ tx_bytes: Buffer.from(txBytes).toString('base64'), mode })
      });
      return res.json();
    },
    /**
     * Convenience: build Dex swap tx for dApp requests
     */
    buildSwapTx({ sender, poolId, denomIn, amountIn, minAmountOut, slippageBps = 50 }) {
      return DexModule.buildSwapExactInTx({
        sender,
        poolId: poolId?.toString(),
        denomIn,
        amountIn,
        minAmountOut,
        maxSlippageBps: Number(slippageBps) || 50,
        memo: 'Keplr swap via Aura extension',
      });
    },
    coinDenom: COIN.base,
  };
}
