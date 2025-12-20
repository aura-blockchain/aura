import DexModule from './src/dex.js';
import { buildKeplrChainInfo } from './src/keplr';
import { buildKeplrProvider, ensureKeplrChain } from './src/keplr-provider';
import { connectWalletConnect, getWalletConnectApproval, onWalletConnectStatus, restoreWalletConnectSession, disconnectWalletConnect, setWalletConnectProvider } from './src/walletconnect';
const { CHAIN_CONFIG, COIN, GAS_PRICE_TIERS, REST_ENDPOINTS, RPC_ENDPOINTS } = require('../config/chain');
const HWModule = require('./hardware-wallet');
let HardwareWalletManager = HWModule;
if (HWModule && typeof HWModule !== 'function') {
  HardwareWalletManager = HWModule.default || HWModule.HardwareWalletManager;
}
if (typeof HardwareWalletManager !== 'function') {
  HardwareWalletManager = class {
    isConnected() { return false; }
  };
}

// Import Cosmos SDK module
/* global COSMOS_SDK */

const API_KEY = 'apiHost';
const PRIVATE_KEY_STORAGE = 'walletPrivateKey';
const SESSION_TOKEN_KEY = 'walletSessionToken';
const SESSION_SECRET_KEY = 'walletSessionSecret';
const SESSION_ADDRESS_KEY = 'walletSessionAddress';
const HKDF_INFO = new TextEncoder().encode('walletconnect-trade');
const DEFAULT_SLIPPAGE_PERCENT = 0.5;
const DEFAULT_API_HOST = REST_ENDPOINTS[0]?.address || 'http://localhost:1317';
const DEFAULT_HW_PATH = `m/44'/${CHAIN_CONFIG.slip44 ?? 118}'/0'/0/0`;
const hwManager = new HardwareWalletManager();
const KEPLR_CHAIN_INFO = buildKeplrChainInfo();
let hwAddress = null;
let hwPublicKeyHex = null;
let useHardwareSigning = false;

let dexPoolsCache = [];
let lastSwapQuote = null;

function bufferToHex(buffer) {
  return Array.from(new Uint8Array(buffer))
    .map(byte => byte.toString(16).padStart(2, '0'))
    .join('');
}

function hexToBytes(hex) {
  const bytes = [];
  for (let c = 0; c < hex.length; c += 2) {
    bytes.push(parseInt(hex.substr(c, 2), 16));
  }
  return new Uint8Array(bytes);
}

function bufferToBase64(buffer) {
  return btoa(String.fromCharCode(...new Uint8Array(buffer)));
}

function base64ToBytes(base64) {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

function buildFallbackWalletConnectProvider() {
  const err = () => Promise.reject(new Error('WalletConnect signing requires Keplr/Leap or an imported/hardware key.'));
  return {
    async getKey() {
      throw new Error('No WalletConnect signer available. Register Keplr/Leap or load hardware key.');
    },
    signAmino: err,
    signDirect: err,
    sendTx: err,
  };
}

function buildLocalSignerProvider() {
  return {
    async signDirect(signer, signDoc) {
      try {
        const current = $('#walletAddress')?.value?.trim();
        if (current && signer && signer !== current) {
          throw new Error('Signer address mismatch. Switch wallet to match request.');
        }
        const privateKeyHex = await getPrivateKey();
        const hasHardware = useHardwareSigning && hwManager.isConnected() && hwPublicKeyHex;
        if (!privateKeyHex && !hasHardware) {
          throw new Error('No signer available. Connect hardware wallet or import a key.');
        }
        const tx = signDoc?.bodyBytes && signDoc?.authInfoBytes
          ? COSMOS_SDK.decodeTx(signDoc.bodyBytes, signDoc.authInfoBytes)
          : signDoc;
        const signedTx = await signWithAvailableWallet(tx, signer, null, { preferHardware: hasHardware });
        const sigBytes = signedTx?.signatures?.[0];
        const pubKeyAny = signedTx?.auth_info?.signer_infos?.[0]?.public_key;
        const pubKeyHex = pubKeyAny?.value || pubKeyAny?.key || hwPublicKeyHex || '';
        const signatureHex = Buffer.from(sigBytes || []).toString('hex');
        return {
          signed: signDoc,
          signature: {
            signature: signatureHex,
            pub_key: {
              type: 'tendermint/PubKeySecp256k1',
              value: pubKeyHex,
            },
          },
        };
      } catch (err) {
        showMessage('walletConnectStatus', err.message, true);
        throw err;
      }
    },
    async signAmino(signer, signDoc) {
      try {
        const current = $('#walletAddress')?.value?.trim();
        if (current && signer && signer !== current) {
          throw new Error('Signer address mismatch. Switch wallet to match request.');
        }
        const privateKeyHex = await getPrivateKey();
        const hasHardware = useHardwareSigning && hwManager.isConnected() && hwPublicKeyHex;
        if (!privateKeyHex && !hasHardware) {
          throw new Error('No signer available. Connect hardware wallet or import a key.');
        }
        const tx = signDoc?.msgs ? COSMOS_SDK.encodeAminoTx(signDoc) : signDoc;
        const signedTx = await signWithAvailableWallet(tx, signer, null, { preferHardware: hasHardware });
        const sigBytes = signedTx?.signatures?.[0];
        const pubKeyAny = signedTx?.auth_info?.signer_infos?.[0]?.public_key;
        const pubKeyHex = pubKeyAny?.value || pubKeyAny?.key || hwPublicKeyHex || '';
        const signatureHex = Buffer.from(sigBytes || []).toString('hex');
        return {
          signed: signDoc,
          signature: {
            signature: signatureHex,
            pub_key: {
              type: 'tendermint/PubKeySecp256k1',
              value: pubKeyHex,
            },
          },
        };
      } catch (err) {
        showMessage('walletConnectStatus', err.message, true);
        throw err;
      }
    },
  };
}

function getHardwareWalletType() {
  const select = $('#hardwareWalletType');
  return (select?.value || 'ledger').toLowerCase();
}

function stableStringify(value) {
  if (value === null) {return 'null';}
  if (Array.isArray(value)) {
    return `[${value.map(item => stableStringify(item)).join(',')}]`;
  }
  if (typeof value === 'object') {
    const keys = Object.keys(value).sort();
    return `{${keys.map(key => `"${key}":${stableStringify(value[key])}`).join(',')}}`;
  }
  return JSON.stringify(value);
}

function getMicroFactor() {
  const decimals = Number(COSMOS_SDK?.config?.coinDecimals ?? COIN.exponent ?? 6);
  return 10 ** decimals;
}

function amountToBaseUnits(amount) {
  if (!Number.isFinite(amount) || amount <= 0) {
    throw new Error('Amount must be greater than zero');
  }
  return BigInt(Math.floor(amount * getMicroFactor()));
}

function baseUnitsToDisplay(value) {
  const asBigInt = typeof value === 'string' ? BigInt(value) : BigInt(value || 0);
  return Number(asBigInt) / getMicroFactor();
}

function normalizeDenom(denom) {
  return (denom || '').trim().toLowerCase();
}

async function loadPoolById(poolId) {
  const idStr = poolId?.toString();
  if (!idStr) {
    return null;
  }

  let pool = dexPoolsCache.find(p => p.pool_id === idStr || p.id === idStr);
  if (!pool) {
    pool = await DexModule.queryPool(idStr);
    if (pool) {
      dexPoolsCache = [pool, ...dexPoolsCache.filter(existing => (existing.pool_id || existing.id) !== idStr)];
    }
  }
  return pool;
}

function updateSwapQuoteDetails(quote, denomOut) {
  const quoteElement = $('#swapQuoteDetails');
  if (!quoteElement || !quote) {
    return;
  }

  const expectedDisplay = baseUnitsToDisplay(quote.expectedAmountOut).toFixed(6);
  const minDisplay = baseUnitsToDisplay(quote.minAmountOut).toFixed(6);
  quoteElement.textContent = `Expected: ${expectedDisplay} ${denomOut} | Min: ${minDisplay} ${denomOut} | Price impact: ${quote.priceImpactPercent.toFixed(2)}%`;
}

async function signPayload(payloadStr, secretHex) {
  const encoder = new TextEncoder();
  const key = await crypto.subtle.importKey(
    'raw',
    hexToBytes(secretHex),
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign']
  );
  const sig = await crypto.subtle.sign('HMAC', key, encoder.encode(payloadStr));
  return bufferToHex(sig);
}

async function saveSession(token, secret, address) {
  return new Promise(resolve => {
    chrome.storage.local.set(
      {
        [SESSION_TOKEN_KEY]: token,
        [SESSION_SECRET_KEY]: secret,
        [SESSION_ADDRESS_KEY]: address,
      },
      () => resolve()
    );
  });
}

async function getSession() {
  return new Promise(resolve => {
    chrome.storage.local.get(
      [SESSION_TOKEN_KEY, SESSION_SECRET_KEY, SESSION_ADDRESS_KEY],
      result => {
        resolve({
          sessionToken: result[SESSION_TOKEN_KEY],
          sessionSecret: result[SESSION_SECRET_KEY],
          walletAddress: result[SESSION_ADDRESS_KEY],
        });
      }
    );
  });
}

async function registerSession(address) {
  const host = await getApiHost();
  const response = await fetch(`${host}/wallet-trades/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ wallet_address: address }),
  });
  const payload = await response.json();
  if (payload.success) {
    await saveSession(payload.session_token, payload.session_secret, address);
    return {
      sessionToken: payload.session_token,
      sessionSecret: payload.session_secret,
      walletAddress: address,
    };
  }
  return null;
}

// eslint-disable-next-line no-unused-vars
async function beginWalletConnectHandshake(address) {
  const host = await getApiHost();
  const response = await fetch(`${host}/wallet-trades/wc/handshake`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ wallet_address: address }),
  });
  const payload = await response.json();
  return payload.success ? payload : null;
}

// eslint-disable-next-line no-unused-vars
async function confirmWalletConnectHandshake(address, handshake) {
  const host = await getApiHost();
  const clientKeyPair = await crypto.subtle.generateKey(
    { name: 'ECDH', namedCurve: 'P-256' },
    true,
    ['deriveBits']
  );
  const clientPublicRaw = await crypto.subtle.exportKey('raw', clientKeyPair.publicKey);
  const serverPublicBytes = base64ToBytes(handshake.server_public);
  const serverKey = await crypto.subtle.importKey(
    'raw',
    serverPublicBytes,
    { name: 'ECDH', namedCurve: 'P-256' },
    false,
    []
  );
  const sharedBits = await crypto.subtle.deriveBits(
    { name: 'ECDH', public: serverKey },
    clientKeyPair.privateKey,
    256
  );
  const hkdfKey = await crypto.subtle.importKey(
    'raw',
    sharedBits,
    { name: 'HKDF', hash: 'SHA-256' },
    false,
    ['deriveBits']
  );
  const derivedBits = await crypto.subtle.deriveBits(
    {
      name: 'HKDF',
      hash: 'SHA-256',
      salt: new TextEncoder().encode(handshake.handshake_id),
      info: HKDF_INFO,
    },
    hkdfKey,
    256
  );
  const derivedHex = bufferToHex(derivedBits);
  const clientPublicBase64 = bufferToBase64(clientPublicRaw);
  const response = await fetch(`${host}/wallet-trades/wc/confirm`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      handshake_id: handshake.handshake_id,
      wallet_address: address,
      client_public: clientPublicBase64,
    }),
  });
  const payload = await response.json();
  if (!payload.success) {return null;}
  await saveSession(payload.session_token, derivedHex, address);
  return { sessionToken: payload.session_token, sessionSecret: derivedHex, walletAddress: address };
}

async function ensureSession() {
  const address = $('#walletAddress').value.trim();
  if (!address) {return null;}
  const session = await getSession();
  if (
    session &&
    session.walletAddress === address &&
    session.sessionToken &&
    session.sessionSecret
  ) {
    return session;
  }
  return registerSession(address);
}

async function getApiHost() {
  return new Promise(resolve => {
    chrome.storage.local.get([API_KEY], result => {
      resolve(result[API_KEY] || 'http://localhost:1317');
    });
  });
}

/**
 * Cosmos SDK Integration Functions
 */

async function getPrivateKey() {
  return new Promise(resolve => {
    chrome.storage.local.get([PRIVATE_KEY_STORAGE], result => {
      resolve(result[PRIVATE_KEY_STORAGE] || null);
    });
  });
}

async function savePrivateKey(privateKeyHex) {
  return new Promise(resolve => {
    chrome.storage.local.set({ [PRIVATE_KEY_STORAGE]: privateKeyHex }, () => resolve());
  });
}

async function signWithAvailableWallet(tx, signerAddress, messageId = null, options = {}) {
  const accountInfo = await COSMOS_SDK.getAccount(signerAddress);

  // Prefer hardware if connected and address matches
  const preferHardware = options.preferHardware || (useHardwareSigning && hwManager.isConnected());
  if (preferHardware && hwManager.isConnected()) {
    if (!hwAddress || !hwPublicKeyHex) {
      updateHardwareStatus('Hardware connected but address not loaded. Click "Get Address" to sync.', true);
      throw new Error('Hardware wallet connected but address not loaded. Click "Get Address" first.');
    }
    if (hwAddress !== signerAddress) {
      updateHardwareStatus('Hardware address does not match current wallet. Load address again.', true);
      throw new Error('Hardware wallet address does not match the selected signer address.');
    }
    const publicKey = hexToBytes(hwPublicKeyHex);
    const { signBytes } = COSMOS_SDK.buildSignDoc(tx, accountInfo, publicKey);
    let sigHex = '';
    if (hwManager.connectedDevice?.type === 'ledger') {
      const ledgerSig = await hwManager.signWithLedger(signBytes, DEFAULT_HW_PATH);
      sigHex = ledgerSig.signature;
    } else {
      const resp = await hwManager.signTransaction(signBytes, DEFAULT_HW_PATH);
      sigHex = resp.signature;
    }
    const sigBytes = hexToBytes(sigHex);
    return COSMOS_SDK.applySignature(tx, sigBytes, publicKey, accountInfo.sequence);
  }

  // Fallback to software key
  const privateKeyHex = await getPrivateKey();
  if (!privateKeyHex) {
    const errMsg = 'No signing key available. Connect hardware wallet or import a software key.';
    if (messageId) {
      showMessage(messageId, errMsg, true);
    }
    throw new Error(errMsg);
  }
  const privateKey = COSMOS_SDK.hexToBytes(privateKeyHex);
  const publicKey = await COSMOS_SDK.getPublicKey(privateKey);
  return COSMOS_SDK.signTx(tx, privateKey, accountInfo, publicKey);
}

async function generateNewWallet() {
  try {
    // Confirm wallet creation
    const existingKey = await getPrivateKey();
    if (existingKey) {
      const confirmed = confirm(
        'You already have a wallet. Creating a new one will replace it. ' +
        'Make sure you have backed up your current private key! Continue?'
      );
      if (!confirmed) {
        return;
      }
    }

    showMessage('walletMessage', 'Generating new wallet...');

    const privateKey = COSMOS_SDK.generatePrivateKey();
    if (privateKey.length !== 32) {
      throw new Error('Invalid private key length');
    }

    const privateKeyHex = COSMOS_SDK.bytesToHex(privateKey);
    const publicKey = await COSMOS_SDK.getPublicKey(privateKey);
    const address = COSMOS_SDK.publicKeyToAddress(publicKey);

    if (!validateCosmosAddress(address)) {
      throw new Error('Generated address validation failed');
    }

    await savePrivateKey(privateKeyHex);
    $('#walletAddress').value = address;
    chrome.storage.local.set({ walletAddress: address });

    showMessage('walletMessage', `New wallet created: ${address}`);

    // Show backup warning
    setTimeout(() => {
      alert(
        'IMPORTANT: Back up your private key!\n\n' +
        'Use the "Export Private Key" button to view and save your private key. ' +
        'Without it, you cannot recover your wallet if you lose access to this browser.'
      );
    }, 500);

    await updateBalance();
  } catch (error) {
    showMessage('walletMessage', `Error creating wallet: ${error.message}`, true);
    console.error('Wallet generation error:', error);
  }
}

async function importWallet(privateKeyHex) {
  try {
    // Validate input
    if (!privateKeyHex || typeof privateKeyHex !== 'string') {
      throw new Error('Private key is required');
    }

    // Remove whitespace and validate hex format
    privateKeyHex = privateKeyHex.trim().toLowerCase();
    if (!/^[0-9a-f]{64}$/i.test(privateKeyHex)) {
      throw new Error('Invalid private key format. Must be 64 hex characters (32 bytes)');
    }

    showMessage('walletMessage', 'Importing wallet...');

    const privateKey = COSMOS_SDK.hexToBytes(privateKeyHex);
    if (privateKey.length !== 32) {
      throw new Error('Invalid private key length. Must be 32 bytes');
    }

    const publicKey = await COSMOS_SDK.getPublicKey(privateKey);
    const address = COSMOS_SDK.publicKeyToAddress(publicKey);

    if (!validateCosmosAddress(address)) {
      throw new Error('Generated address validation failed');
    }

    // Confirm if overwriting existing wallet
    const existingKey = await getPrivateKey();
    if (existingKey && existingKey !== privateKeyHex) {
      const confirmed = confirm(
        'You already have a wallet. Importing will replace it. ' +
        'Make sure you have backed up your current private key! Continue?'
      );
      if (!confirmed) {
        return;
      }
    }

    await savePrivateKey(privateKeyHex);
    $('#walletAddress').value = address;
    chrome.storage.local.set({ walletAddress: address });

    showMessage('walletMessage', `Wallet imported successfully: ${address}`);
    await updateBalance();
  } catch (error) {
    showMessage('walletMessage', `Error importing wallet: ${error.message}`, true);
    console.error('Wallet import error:', error);
  }
}

async function updateBalance() {
  const address = $('#walletAddress').value.trim();
  if (!address) {
    $('#balanceDisplay').textContent = 'Balance: Enter address';
    return;
  }

  try {
    const balances = await COSMOS_SDK.getBalance(address);
    if (balances.length === 0) {
      $('#balanceDisplay').textContent = 'Balance: 0 AURA';
      return;
    }

    const balanceText = balances
      .map(b => {
        const amount = parseInt(b.amount) / Math.pow(10, COSMOS_SDK.config.coinDecimals);
        const denom = b.denom === 'uaura' ? 'AURA' : b.denom;
        return `${amount} ${denom}`;
      })
      .join(', ');

    $('#balanceDisplay').textContent = `Balance: ${balanceText}`;
  } catch (error) {
    $('#balanceDisplay').textContent = `Balance: Error - ${error.message}`;
  }
}

async function sendTokens(toAddress, amount, denom = 'upaw') {
  const fromAddress = $('#walletAddress').value.trim();

  // Validation
  if (!fromAddress) {
    showMessage('transactionMessage', 'Please enter your wallet address', true);
    return null;
  }

  if (!validateCosmosAddress(fromAddress)) {
    showMessage('transactionMessage', 'Invalid sender address format', true);
    return null;
  }

  if (!validateCosmosAddress(toAddress)) {
    showMessage('transactionMessage', 'Invalid recipient address format', true);
    return null;
  }

  if (!amount || amount <= 0) {
    showMessage('transactionMessage', 'Invalid amount. Must be greater than 0', true);
    return null;
  }

  try {
    showMessage('transactionMessage', 'Preparing transaction...');

    const amountInMicroDenom = Math.floor(amount * Math.pow(10, COSMOS_SDK.config.coinDecimals));

    if (amountInMicroDenom <= 0) {
      throw new Error('Amount too small to send');
    }

    const tx = COSMOS_SDK.buildTransferTx({
      fromAddress,
      toAddress,
      amount: amountInMicroDenom,
      denom,
      memo: 'Sent from Aura Browser Wallet',
    });

    showMessage('transactionMessage', 'Signing transaction...');
    const signedTx = await signWithAvailableWallet(tx, fromAddress, 'transactionMessage');

    showMessage('transactionMessage', 'Broadcasting transaction...');
    const result = await COSMOS_SDK.broadcastTx(signedTx);

    showMessage('transactionMessage', `Transaction successful! Hash: ${result.txhash}`);
    await updateBalance();
    await refreshTradeHistory();
    return result;
  } catch (error) {
    const errorMsg = error.message || 'Unknown error occurred';
    showMessage('transactionMessage', `Transaction failed: ${errorMsg}`, true);
    console.error('Send tokens error:', error);
    return null;
  }
}

async function executeSwap(poolId, tokenInDenom, tokenInAmount, tokenOutDenom, minAmountOut, slippageBps = 50) {
  const sender = $('#walletAddress').value.trim();

  // Validation
  if (!sender) {
    showMessage('tradeMessage', 'Please enter your wallet address', true);
    return null;
  }

  if (!validateCosmosAddress(sender)) {
    showMessage('tradeMessage', 'Invalid wallet address format', true);
    return null;
  }

  if (!poolId || poolId <= 0) {
    showMessage('tradeMessage', 'Invalid pool ID', true);
    return null;
  }

  if (!tokenInAmount || tokenInAmount <= 0) {
    showMessage('tradeMessage', 'Invalid swap amount. Must be greater than 0', true);
    return null;
  }

  if (!tokenInDenom || !tokenOutDenom) {
    showMessage('tradeMessage', 'Token denominations required', true);
    return null;
  }

  try {
    showMessage('tradeMessage', 'Preparing swap transaction...');

    const tx = DexModule.buildSwapExactInTx({
      sender,
      poolId: poolId.toString(),
      denomIn: normalizeDenom(tokenInDenom),
      amountIn: tokenInAmount,
      minAmountOut: minAmountOut,
      maxSlippageBps: Number(slippageBps) || 50,
      memo: 'DEX Swap from Aura Browser Wallet',
    });

    showMessage('tradeMessage', 'Signing swap transaction...');
    const signedTx = await signWithAvailableWallet(tx, sender, 'tradeMessage');

    showMessage('tradeMessage', 'Broadcasting swap transaction...');
    const result = await COSMOS_SDK.broadcastTx(signedTx);

    showMessage('tradeMessage', `Swap successful! Hash: ${result.txhash}`);
    await updateBalance();
    await refreshPools();
    await refreshTradeHistory();
    return result;
  } catch (error) {
    const errorMsg = error.message || 'Unknown swap error occurred';
    showMessage('tradeMessage', `Swap failed: ${errorMsg}`, true);
    console.error('Swap execution error:', error);
    return null;
  }
}

async function quoteSwapFromForm() {
  try {
    const fromDenom = normalizeDenom($('#swapFromToken').value);
    const toDenom = normalizeDenom($('#swapToToken').value);
    const amountFloat = parseFloat($('#swapFromAmount').value);
    const poolId = $('#swapPoolId').value.trim();
    const slippagePercent = parseFloat($('#swapSlippage').value || DEFAULT_SLIPPAGE_PERCENT);

    if (!fromDenom || !toDenom || !Number.isFinite(amountFloat) || amountFloat <= 0 || !poolId) {
      throw new Error('Enter valid pool ID, tokens, and amount before quoting');
    }

    const pool = await loadPoolById(poolId);
    if (!pool) {
      throw new Error('Pool not found');
    }

    const slippageBps = Math.max(1, Math.round(slippagePercent * 100));
    const amountIn = amountToBaseUnits(amountFloat);

    const quote = DexModule.calculateSwapQuote(pool, fromDenom, amountIn, slippageBps);
    lastSwapQuote = {
      ...quote,
      slippageBps,
      denomIn: fromDenom,
      denomOut: toDenom,
    };

    const minAmountDisplay = baseUnitsToDisplay(quote.minAmountOut).toFixed(6);
    $('#swapMinAmount').value = minAmountDisplay;
    updateSwapQuoteDetails(quote, toDenom);
    showMessage('tradeMessage', 'Quote updated. Review and execute when ready.');
  } catch (error) {
    lastSwapQuote = null;
    $('#swapQuoteDetails').textContent = `Quote error: ${error.message}`;
    showMessage('tradeMessage', `Quote failed: ${error.message}`, true);
  }
}

async function handleSwapSubmit(event) {
  event.preventDefault();
  try {
    const fromDenom = normalizeDenom($('#swapFromToken').value);
    const toDenom = normalizeDenom($('#swapToToken').value);
    const amountFloat = parseFloat($('#swapFromAmount').value);
    const minAmountFloat = parseFloat($('#swapMinAmount').value);
    const poolId = parseInt($('#swapPoolId').value, 10);
    const slippagePercent = parseFloat($('#swapSlippage').value || DEFAULT_SLIPPAGE_PERCENT);
    const slippageBps = Math.max(1, Math.round(slippagePercent * 100));

    if (!Number.isFinite(amountFloat) || amountFloat <= 0) {
      throw new Error('Enter a valid from amount before swapping');
    }

    const amountIn = amountToBaseUnits(amountFloat);
    const minAmountOut = Number.isFinite(minAmountFloat) && minAmountFloat > 0
      ? amountToBaseUnits(minAmountFloat)
      : amountIn;

    await executeSwap(poolId, fromDenom, amountIn, toDenom, minAmountOut, slippageBps);
  } catch (error) {
    showMessage('tradeMessage', `Swap failed: ${error.message}`, true);
  }
}

async function handleAddLiquiditySubmit(event) {
  event.preventDefault();
  try {
    const provider = $('#walletAddress').value.trim();
    if (!validateCosmosAddress(provider)) {
      throw new Error('Enter a valid wallet address before providing liquidity');
    }

    const poolId = $('#addLiquidityPoolId').value.trim();
    const denomA = normalizeDenom($('#addLiquidityDenomA').value);
    const denomB = normalizeDenom($('#addLiquidityDenomB').value);
    const amountAFloat = parseFloat($('#addLiquidityAmountA').value);
    const amountBFloat = parseFloat($('#addLiquidityAmountB').value);

    if (!denomA || !denomB || !Number.isFinite(amountAFloat) || !Number.isFinite(amountBFloat)) {
      throw new Error('Provide both token denoms and amounts');
    }

    const tx = DexModule.buildAddLiquidityTx({
      provider,
      poolId,
      denomA,
      denomB,
      amountA: amountToBaseUnits(amountAFloat),
      amountB: amountToBaseUnits(amountBFloat),
      memo: 'Provide liquidity via Aura Browser Wallet',
    });

    showMessage('liquidityMessage', 'Signing add-liquidity transaction...');
    const signedTx = await signWithAvailableWallet(tx, provider, 'liquidityMessage');

    showMessage('liquidityMessage', 'Broadcasting add-liquidity transaction...');
    const result = await COSMOS_SDK.broadcastTx(signedTx);

    showMessage('liquidityMessage', `Liquidity added! Hash: ${result.txhash}`);
    await refreshPools();
    await refreshTradeHistory();
  } catch (error) {
    showMessage('liquidityMessage', `Add liquidity failed: ${error.message}`, true);
  }
}

async function handleRemoveLiquiditySubmit(event) {
  event.preventDefault();
  try {
    const provider = $('#walletAddress').value.trim();
    if (!validateCosmosAddress(provider)) {
      throw new Error('Enter a valid wallet address before removing liquidity');
    }

    const poolId = $('#removeLiquidityPoolId').value.trim();
    const lpTokensFloat = parseFloat($('#removeLiquidityLpTokens').value);
    if (!Number.isFinite(lpTokensFloat) || lpTokensFloat <= 0) {
      throw new Error('Enter LP token amount to withdraw');
    }

    const tx = DexModule.buildRemoveLiquidityTx({
      provider,
      poolId,
      lpTokens: amountToBaseUnits(lpTokensFloat),
      memo: 'Remove liquidity via Aura Browser Wallet',
    });

    showMessage('liquidityMessage', 'Signing remove-liquidity transaction...');
    const signedTx = await signWithAvailableWallet(tx, provider, 'liquidityMessage');

    showMessage('liquidityMessage', 'Broadcasting remove-liquidity transaction...');
    const result = await COSMOS_SDK.broadcastTx(signedTx);

    showMessage('liquidityMessage', `Liquidity removed! Hash: ${result.txhash}`);
    await refreshPools();
    await refreshTradeHistory();
  } catch (error) {
    showMessage('liquidityMessage', `Remove liquidity failed: ${error.message}`, true);
  }
}

function showMessage(elementId, message, isError = false) {
  const element = $(`#${elementId}`);
  if (element) {
    element.textContent = message;
    element.classList.toggle('error', isError);
    setTimeout(() => {
      element.textContent = '';
      element.classList.remove('error');
    }, 10000);
  }
}

onWalletConnectStatus((status) => {
  if (!status) return;
  showMessage('walletConnectStatus', status, status.toLowerCase().includes('fail') || status.toLowerCase().includes('reject'));
});

async function setApiHost(host) {
  chrome.storage.local.set({ [API_KEY]: host });
  // Update Cosmos SDK config
  COSMOS_SDK.config.restEndpoint = host;
  COSMOS_SDK.config.rpcEndpoint = host.replace('1317', '26657');
}

/**
 * Hardware wallet helpers
 */
function updateHardwareStatus(message, isError = false) {
  showMessage('hardwareWalletMessage', message, isError);
  const status = $('#hardwareWalletStatus');
  if (status) {
    status.textContent = message;
    status.classList.toggle('error', isError);
  }
}

function setHardwareInfo(info) {
  const infoBox = $('#hardwareWalletInfo');
  const typeEl = $('#hwDeviceType');
  const nameEl = $('#hwDeviceName');
  if (infoBox && typeEl && nameEl) {
    infoBox.classList.remove('hidden');
    typeEl.textContent = `Type: ${info?.type || 'Unknown'}`;
    nameEl.textContent = `Device: ${info?.name || 'Unknown'}`;
  }
}

async function restoreWalletConnect() {
  const existing = await restoreWalletConnectSession('aura-local');
  if (existing?.topic) {
    showMessage('walletConnectStatus', 'Restored WalletConnect session', false);
    const uriEl = $('#walletConnectUri');
    if (uriEl) {
      uriEl.textContent = `topic:${existing.topic}`;
    }
  }
}

async function suggestKeplrChain() {
  if (!window.keplr) {
    updateHardwareStatus('Keplr not detected. Install Keplr to enable dApp provider.', true);
    return;
  }
  try {
    await ensureKeplrChain();
    // Expose a lightweight provider shim for dApps if needed
    const provider = buildKeplrProvider();
    window.auraKeplrProvider = provider;
    setWalletConnectProvider(provider);
    updateHardwareStatus('Keplr/Leap chain info registered.');
  } catch (err) {
    updateHardwareStatus(err.message || 'Failed to register chain with Keplr', true);
  }
}

async function initiateWalletConnect() {
  try {
    showMessage('walletConnectStatus', 'Initializing WalletConnect...', false);
    const { uri } = await connectWalletConnect({ projectId: 'aura-local' });
    const uriEl = $('#walletConnectUri');
    if (uriEl) {
      uriEl.textContent = uri || 'No URI';
    }
    showMessage('walletConnectStatus', 'WalletConnect URI generated. Scan with your wallet.', false);
    const approval = getWalletConnectApproval();
    if (approval?.then) {
      approval.then(() => {
        showMessage('walletConnectStatus', 'WalletConnect session approved.', false);
      }).catch(err => {
        showMessage('walletConnectStatus', err?.message || 'WalletConnect approval failed', true);
      });
    }
  } catch (err) {
    showMessage('walletConnectStatus', err.message || 'WalletConnect init failed', true);
  }
}

async function detectHardwareWallet() {
  try {
    const hwType = getHardwareWalletType();
    if (hwType === 'trezor') {
      updateHardwareStatus('Trezor detection uses Connect prompt. Click Connect to continue.');
      return;
    }
    updateHardwareStatus('Detecting devices...');
    const devices = await hwManager.detectDevices();
    if (!devices.length) {
      updateHardwareStatus('No hardware wallets found', true);
      return;
    }
    updateHardwareStatus(`Detected ${devices.length} device(s)`);
  } catch (err) {
    updateHardwareStatus(err.message || 'Hardware detection failed', true);
  }
}

async function connectHardwareWallet() {
  const connectBtn = $('#connectHardwareWallet');
  const statusEl = $('#hardwareWalletStatus');
  try {
    const hwType = getHardwareWalletType();
    if (connectBtn) {
      connectBtn.disabled = true;
      connectBtn.textContent = 'Connecting...';
      connectBtn.classList.add('hw-connecting');
    }
    if (statusEl) statusEl.classList.add('hw-connecting');
    updateHardwareStatus(`Connecting to ${hwType === 'trezor' ? 'Trezor' : 'Ledger'}... (30s timeout)`);
    const device = await hwManager.connect(hwType);
    setHardwareInfo(hwManager.getDeviceInfo());
    updateHardwareStatus(`Connected: ${device.type}`);
    useHardwareSigning = true;
    const toggle = $('#useHardwareSigning');
    if (toggle) {
      toggle.checked = true;
    }
  } catch (err) {
    updateHardwareStatus(err.message || 'Hardware connection failed', true);
  } finally {
    if (connectBtn) {
      connectBtn.disabled = false;
      connectBtn.textContent = 'Connect';
      connectBtn.classList.remove('hw-connecting');
    }
    if (statusEl) statusEl.classList.remove('hw-connecting');
  }
}

function resetHardwareWallet() {
  hwManager.reset();
  useHardwareSigning = false;
  hwAddress = null;
  hwPublicKeyHex = null;
  updateHardwareStatus('Reset complete. Ready to reconnect.');
  const toggle = $('#useHardwareSigning');
  if (toggle) toggle.checked = false;
  const connectBtn = $('#connectHardwareWallet');
  if (connectBtn) {
    connectBtn.disabled = false;
    connectBtn.textContent = 'Connect';
    connectBtn.classList.remove('hw-connecting');
  }
  const infoBox = $('#hardwareWalletInfo');
  if (infoBox) infoBox.classList.add('hidden');
}

async function disconnectHardwareWallet() {
  await hwManager.disconnect();
  await disconnectWalletConnect();
  useHardwareSigning = false;
  hwAddress = null;
  hwPublicKeyHex = null;
  updateHardwareStatus('Disconnected');
  const toggle = $('#useHardwareSigning');
  if (toggle) {
    toggle.checked = false;
  }
}

async function fetchHardwareAddress() {
  try {
    updateHardwareStatus('Requesting address on device...');
    const hwType = getHardwareWalletType();
    const resp = await hwManager.getAddress(DEFAULT_HW_PATH, hwType === 'ledger');
    if (resp?.address) {
      hwAddress = resp.address;
      hwPublicKeyHex = resp.publicKey;
      useHardwareSigning = true;
      const input = $('#walletAddress');
      if (input) {
        input.value = resp.address;
        chrome.storage.local.set({ walletAddress: resp.address });
      }
      updateHardwareStatus(`Address: ${resp.address}`);
      const status = $('#hardwareSigningStatus');
      if (status) {
        status.textContent = 'Hardware signing: enabled';
      }
      const toggle = $('#useHardwareSigning');
      if (toggle) {
        toggle.checked = true;
      }
    } else {
      updateHardwareStatus('No address returned', true);
    }
  } catch (err) {
    updateHardwareStatus(err.message || 'Failed to fetch address', true);
  }
}

/**
 * Additional Helper Functions
 */

async function exportPrivateKey() {
  const privateKeyHex = await getPrivateKey();
  if (!privateKeyHex) {
    showMessage('walletMessage', 'No private key found', true);
    return;
  }

  const confirmed = confirm(
    'WARNING: Never share your private key with anyone! ' +
    'Anyone with access to your private key can steal your funds. ' +
    'Are you sure you want to view it?'
  );

  if (confirmed) {
    alert(`Your private key:\n\n${privateKeyHex}\n\nStore this securely and never share it!`);
  }
}

async function deleteWallet() {
  const confirmed = confirm(
    'WARNING: This will delete your private key from this browser. ' +
    'Make sure you have backed up your private key first! ' +
    'This action cannot be undone. Continue?'
  );

  if (confirmed) {
    await chrome.storage.local.remove([PRIVATE_KEY_STORAGE, 'walletAddress']);
    $('#walletAddress').value = '';
    $('#balanceDisplay').textContent = 'Balance: Wallet deleted';
    showMessage('walletMessage', 'Wallet deleted successfully');
  }
}

async function queryAccountInfo() {
  const address = $('#walletAddress').value.trim();
  if (!address) {
    showMessage('walletMessage', 'Enter a wallet address first', true);
    return;
  }

  try {
    const accountInfo = await COSMOS_SDK.getAccount(address);
    const message = `Account Number: ${accountInfo.accountNumber}\n` +
                   `Sequence: ${accountInfo.sequence}\n` +
                   `Address: ${accountInfo.address}`;
    alert(message);
  } catch (error) {
    showMessage('walletMessage', `Error fetching account: ${error.message}`, true);
  }
}

function validateCosmosAddress(address) {
  // Basic validation for Cosmos Bech32 addresses
  if (!address || typeof address !== 'string') {
    return false;
  }
  return address.startsWith(COSMOS_SDK.config.bech32Prefix) && address.length >= 39;
}

async function checkNetworkConnection() {
  try {
    const response = await fetch(`${COSMOS_SDK.config.rpcEndpoint}/status`);
    if (response.ok) {
      const data = await response.json();
      return {
        connected: true,
        chainId: data.result?.node_info?.network || CHAIN_CONFIG.chainId,
        latestHeight: data.result?.sync_info?.latest_block_height,
      };
    }
    return { connected: false };
  } catch (error) {
    return { connected: false, error: error.message };
  }
}

function $(selector) {
  return document.querySelector(selector);
}

async function updateMiningStatus() {
  const address = $('#walletAddress').value.trim();
  if (!address) {
    $('#miningStatus').textContent = 'Status: wallet address required';
    return;
  }

  try {
    // Query validator status from Cosmos SDK
    const validatorUrl = `${COSMOS_SDK.config.rpcEndpoint}/validators`;
    const statusRes = await fetch(validatorUrl);
    if (!statusRes.ok) {
      $('#miningStatus').textContent = 'Status: unable to reach validator API';
      return;
    }

    const data = await statusRes.json();
    $('#miningStatus').textContent = 'Status: Network connected';
    $('#miningMeta').textContent = `Validators: ${data.result?.validators?.length || 0}`;
  } catch (error) {
    $('#miningStatus').textContent = `Status: ${error.message}`;
    $('#miningMeta').textContent = 'Network unavailable';
  }
}

async function startMining() {
  const address = $('#walletAddress').value.trim();
  if (!address) {
    showMessage('miningMessage', 'Enter a wallet address first', true);
    return;
  }

  showMessage('miningMessage', 'Note: Aura uses Proof-of-Stake. Use staking controls instead of mining.');
  await updateMiningStatus();
}

async function stopMining() {
  showMessage('miningMessage', 'Note: Aura uses Proof-of-Stake. Check staking section.');
  await updateMiningStatus();
}

async function refreshPools() {
  const list = $('#poolsList');
  if (!list) {
    return;
  }

  try {
    const pools = await DexModule.queryPools();
    dexPoolsCache = pools || [];

    if (!dexPoolsCache.length) {
      list.innerHTML = '<h3>Liquidity Pools</h3><div class="list-placeholder">No pools available</div>';
      return;
    }

    const entries = dexPoolsCache.slice(0, 5).map(pool => {
      const formatted = DexModule.formatPool(pool);
      return `
        <div class="entry">
          <strong>Pool #${formatted.poolId}</strong><br />
          ${formatted.denomA}/${formatted.denomB} · Price ${formatted.priceAB.toFixed(4)} ${formatted.denomB}<br />
          Depth: ${formatted.depth[formatted.denomA]} ${formatted.denomA} / ${formatted.depth[formatted.denomB]} ${formatted.denomB}
        </div>`;
    });

    list.innerHTML = `<h3>Liquidity Pools</h3>${entries.join('')}`;
  } catch (error) {
    list.innerHTML = `<h3>Liquidity Pools</h3><div class="list-placeholder">Error loading pools: ${error.message}</div>`;
  }
}

async function refreshMatches() {
  try {
    const prices = await COSMOS_SDK.queryOraclePrices();
    const list = $('#matchesList');

    if (prices && prices.length > 0) {
      list.innerHTML = prices
        .slice(0, 10)
        .map(
          price => `
        <div class="entry">
          ${price.symbol}: $${price.price}
          <br />
          Updated: ${new Date(price.timestamp * 1000).toLocaleTimeString()}
        </div>`
        )
        .join('');
    } else {
      list.innerHTML = '<div class="list-placeholder">No price feeds available</div>';
    }
  } catch (error) {
    $('#matchesList').innerHTML = `<div class="list-placeholder">Error loading prices: ${error.message}</div>`;
  }
}

async function refreshTradeHistory() {
  const address = $('#walletAddress').value.trim();
  if (!address) {
    $('#tradeHistory').innerHTML = '<div class="list-placeholder">Enter wallet address</div>';
    return;
  }

  try {
    const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/tx/v1beta1/txs?events=message.sender='${address}'&order_by=ORDER_BY_DESC&limit=5`;
    const res = await fetch(url);

    if (!res.ok) {
      $('#tradeHistory').innerHTML = '<div class="list-placeholder">Unable to fetch history</div>';
      return;
    }

    const data = await res.json();
    const container = $('#tradeHistory');

    if (data.txs && data.txs.length > 0) {
      container.innerHTML = data.txs
        .map(
          tx => `
        <div class="entry">
          TX @ height ${tx.height || 'pending'}
          <br />
          Hash: ${tx.txhash?.substring(0, 12)}... | Fee: ${tx.auth_info?.fee?.amount?.[0]?.amount || 0}
        </div>`
        )
        .join('');
    } else {
      container.innerHTML = '<div class="list-placeholder">No transactions found</div>';
    }
  } catch (error) {
    $('#tradeHistory').innerHTML = `<div class="list-placeholder">Error: ${error.message}</div>`;
  }
}

async function refreshMinerStats() {
  const address = $('#walletAddress').value.trim();
  if (!address) {
    $('#minerStats').textContent = 'Enter address for staking stats';
    return;
  }

  try {
    const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/staking/v1beta1/delegations/${address}`;
    const res = await fetch(url);

    if (!res.ok) {
      $('#minerStats').textContent = 'No staking data';
      $('#minerHistory').textContent = 'Not staking';
      return;
    }

    const data = await res.json();
    const delegations = data.delegation_responses || [];

    if (delegations.length > 0) {
      const totalStaked = delegations.reduce((sum, del) => {
        return sum + parseInt(del.balance?.amount || 0);
      }, 0) / Math.pow(10, COSMOS_SDK.config.coinDecimals);

      $('#minerStats').textContent = `Staked: ${totalStaked} AURA | Delegations: ${delegations.length}`;
      $('#minerHistory').textContent = 'Active validator delegations';
    } else {
      $('#minerStats').textContent = 'No active delegations';
      $('#minerHistory').textContent = 'Not staking';
    }
  } catch (error) {
    $('#minerStats').textContent = `Error: ${error.message}`;
    $('#minerHistory').textContent = 'Unable to fetch staking data';
  }
}

async function submitOrder(event) {
  event.preventDefault();
  const form = event.target;
  const formData = new FormData(form);

  const tokenOffered = formData.get('tokenOffered');
  const amountOffered = parseFloat(formData.get('amountOffered'));
  const tokenRequested = formData.get('tokenRequested');
  const amountRequested = parseFloat(formData.get('amountRequested'));

  if (!tokenOffered || !tokenRequested || !amountOffered || !amountRequested) {
    showMessage('tradeMessage', 'Please fill in all fields', true);
    return;
  }

  try {
    // For now, use the swap function
    // In a real implementation, this would map to DEX order creation
    const poolId = 1; // Default pool - should be dynamically selected
    const slippageBps = 500;
    const amountIn = amountToBaseUnits(amountOffered);
    const minAmountOut = amountToBaseUnits(amountRequested * 0.95);

    const result = await executeSwap(
      poolId,
      tokenOffered,
      amountIn,
      tokenRequested,
      minAmountOut,
      slippageBps
    );

    if (result) {
      showMessage('tradeMessage', `Swap successful! Hash: ${result.txhash}`);
      await refreshPools();
      await refreshTradeHistory();
    }
  } catch (error) {
    showMessage('tradeMessage', `Swap failed: ${error.message}`, true);
  }
}

function setAiStatus(message, isError = false) {
  const aiStatus = $('#aiStatus');
  aiStatus.textContent = message;
  aiStatus.classList.toggle('error', isError);
}

function setKeyDeletionNotice(message) {
  $('#aiKeyDeleted').textContent = message;
}

function clearAiKeyField() {
  const keyInput = $('#aiApiKey');
  keyInput.value = '';
  setKeyDeletionNotice('Your AI API key has been deleted from this wallet.');
}

async function runPersonalAiSwap() {
  const userAddress = $('#walletAddress').value.trim();
  const mode = $('#aiKeyMode').value;
  let apiKey = $('#aiApiKey').value.trim();
  const provider = $('#aiProvider').value.trim() || 'anthropic';
  const model = $('#aiModel').value.trim() || 'claude-sonnet-4-5';
  const swapDetails = {
    from_coin: $('#aiFromCoin').value.trim() || 'AURA',
    to_coin: $('#aiToCoin').value.trim() || 'USDC',
    amount: parseFloat($('#aiAmount').value) || 0,
    recipient_address: $('#aiRecipient').value.trim() || userAddress,
  };

  if (!userAddress) {
    setAiStatus('Provide your wallet address before using the assistant', true);
    return;
  }
  if (mode === 'session') {
    apiKey = apiKey || (await getStoredAiKey());
  }
  if (!apiKey) {
    setAiStatus('Enter your AI API key for this session', true);
    return;
  }
  if (mode === 'session') {
    storeAiKey(apiKey);
  }
  if (!swapDetails.amount) {
    setAiStatus('Enter a swap amount before running the assistant', true);
    return;
  }

  setAiStatus('Preparing AI-assisted swap...');

  try {
    // Execute the swap using Cosmos SDK
    const poolId = 1; // Default pool
    const amountInMicroDenom = Math.floor(
      swapDetails.amount * Math.pow(10, COSMOS_SDK.config.coinDecimals)
    );
    const minAmountOut = Math.floor(amountInMicroDenom * 0.95); // 5% slippage

    const result = await executeSwap(
      poolId,
      swapDetails.from_coin.toLowerCase(),
      amountInMicroDenom,
      swapDetails.to_coin.toLowerCase(),
      minAmountOut
    );

    if (result) {
      setAiStatus(`Swap successful! Hash: ${result.txhash}`);
      setKeyDeletionNotice('Transaction complete. API key removed from this extension.');
      await updateBalance();
      await refreshTradeHistory();
    } else {
      throw new Error('Swap transaction failed');
    }
  } catch (error) {
    setAiStatus(`Swap error: ${error.message}`, true);
  } finally {
    if (mode === 'temporary' || mode === 'external') {
      clearAiKeyField();
    } else if (mode === 'session') {
      setKeyDeletionNotice('Transaction complete. Stored key remains until you click Clear Key.');
    }
  }
}

async function getStoredAiKey() {
  return new Promise(resolve => {
    chrome.storage.local.get(['personalAiApiKey'], result => {
      resolve(result.personalAiApiKey || '');
    });
  });
}

function storeAiKey(value) {
  chrome.storage.local.set({ personalAiApiKey: value });
}

function clearStoredAiKey() {
  chrome.storage.local.remove('personalAiApiKey');
  clearAiKeyField();
  setKeyDeletionNotice('Stored Personal AI key removed.');
}

function bindActions() {
  // Wallet management
  const generateWalletBtn = $('#generateWallet');
  const importWalletBtn = $('#importWallet');
  const refreshBalanceBtn = $('#refreshBalance');
  const exportKeyBtn = $('#exportPrivateKey');
  const deleteWalletBtn = $('#deleteWallet');
  const accountInfoBtn = $('#accountInfo');

  if (generateWalletBtn) {
    generateWalletBtn.addEventListener('click', generateNewWallet);
  }
  if (importWalletBtn) {
    importWalletBtn.addEventListener('click', () => {
      const privateKey = prompt('Enter your private key (hex):');
      if (privateKey) {
        importWallet(privateKey);
      }
    });
  }
  if (refreshBalanceBtn) {
    refreshBalanceBtn.addEventListener('click', updateBalance);
  }
  if (exportKeyBtn) {
    exportKeyBtn.addEventListener('click', exportPrivateKey);
  }
  if (deleteWalletBtn) {
    deleteWalletBtn.addEventListener('click', deleteWallet);
  }
  if (accountInfoBtn) {
    accountInfoBtn.addEventListener('click', queryAccountInfo);
  }

  // Mining/Staking
  const startMiningBtn = $('#startMining');
  const stopMiningBtn = $('#stopMining');
  if (startMiningBtn) startMiningBtn.addEventListener('click', startMining);
  if (stopMiningBtn) stopMiningBtn.addEventListener('click', stopMining);

  // Trading
  const refreshPoolsBtn = $('#refreshPools');
  const refreshMatchesBtn = $('#refreshMatches');
  const refreshTradeHistoryBtn = $('#refreshTradeHistory');
  const swapForm = $('#swapForm');
  const swapQuoteBtn = $('#swapQuoteBtn');
  const addLiquidityForm = $('#addLiquidityForm');
  const removeLiquidityForm = $('#removeLiquidityForm');
  const orderForm = $('#orderForm');

  if (refreshPoolsBtn) refreshPoolsBtn.addEventListener('click', refreshPools);
  if (refreshMatchesBtn) refreshMatchesBtn.addEventListener('click', refreshMatches);
  if (refreshTradeHistoryBtn) refreshTradeHistoryBtn.addEventListener('click', refreshTradeHistory);
  if (swapQuoteBtn) swapQuoteBtn.addEventListener('click', quoteSwapFromForm);
  if (swapForm) swapForm.addEventListener('submit', handleSwapSubmit);
  if (addLiquidityForm) addLiquidityForm.addEventListener('submit', handleAddLiquiditySubmit);
  if (removeLiquidityForm) removeLiquidityForm.addEventListener('submit', handleRemoveLiquiditySubmit);
  if (orderForm) orderForm.addEventListener('submit', submitOrder);

  // AI Assistant
  const runAiSwapBtn = $('#runAiSwap');
  const clearAiKeyBtn = $('#clearAiKey');

  if (runAiSwapBtn) runAiSwapBtn.addEventListener('click', runPersonalAiSwap);
  if (clearAiKeyBtn) clearAiKeyBtn.addEventListener('click', clearStoredAiKey);

  // Theme toggle
  const themeBtn = $('#themeToggle');
  if (themeBtn) themeBtn.addEventListener('click', toggleTheme);

  // Address book
  const addContactForm = $('#addContactForm');
  const selectContactsBtn = $('#selectFromContacts');
  const contactPickerClose = $('#contactPickerClose');
  const addressBookToggle = $('#addressBookToggle');
  if (addContactForm) addContactForm.addEventListener('submit', handleAddContact);
  if (selectContactsBtn) selectContactsBtn.addEventListener('click', openContactPicker);
  if (contactPickerClose) contactPickerClose.addEventListener('click', () => $('#contactPickerModal').classList.add('hidden'));
  if (addressBookToggle) {
    addressBookToggle.addEventListener('click', () => {
      addressBookToggle.closest('.card').classList.toggle('collapsed');
    });
  }

  // CSV export
  const exportCsvBtn = $('#exportCsv');
  if (exportCsvBtn) exportCsvBtn.addEventListener('click', exportTransactionsCsv);

  // Transaction preview modal
  const txPreviewCancel = $('#txPreviewCancel');
  const txPreviewConfirm = $('#txPreviewConfirm');
  if (txPreviewCancel) txPreviewCancel.addEventListener('click', () => closeTxPreview(false));
  if (txPreviewConfirm) txPreviewConfirm.addEventListener('click', () => closeTxPreview(true));

  // API Host
  const apiHostInput = $('#apiHost');
  if (apiHostInput) {
    apiHostInput.placeholder = DEFAULT_API_HOST;
    apiHostInput.addEventListener('change', event => {
      const newHost = event.target.value.trim();
      setApiHost(newHost);
      const host = newHost || DEFAULT_API_HOST;
      COSMOS_SDK.config.restEndpoint = host;
      COSMOS_SDK.config.rpcEndpoint = host.replace('1317', '26657');
    });
  }

  // Hardware wallet actions
  const detectHwBtn = $('#detectHardwareWallet');
  const connectHwBtn = $('#connectHardwareWallet');
  const disconnectHwBtn = $('#disconnectHardwareWallet');
  const getHwAddressBtn = $('#getHwAddress');
  const hwStatus = $('#hardwareSigningStatus');
  const hwToggle = $('#useHardwareSigning');
  const registerKeplrBtn = $('#registerKeplr');
  const wcConnectBtn = $('#connectWalletConnect');
  const wcDisconnectBtn = $('#disconnectWalletConnect');
  setWalletConnectProvider(buildFallbackWalletConnectProvider());
  setWalletConnectProvider(buildLocalSignerProvider());
  restoreWalletConnect();
  if (window.keplr) {
    const keplrProvider = buildKeplrProvider();
    setWalletConnectProvider(keplrProvider);
  }

  const resetHwBtn = $('#resetHardwareWallet');
  if (detectHwBtn) detectHwBtn.addEventListener('click', detectHardwareWallet);
  if (connectHwBtn) connectHwBtn.addEventListener('click', connectHardwareWallet);
  if (resetHwBtn) resetHwBtn.addEventListener('click', resetHardwareWallet);
  if (disconnectHwBtn) disconnectHwBtn.addEventListener('click', disconnectHardwareWallet);
  if (getHwAddressBtn) getHwAddressBtn.addEventListener('click', fetchHardwareAddress);
  if (hwStatus) {
    hwStatus.textContent = useHardwareSigning ? 'Hardware signing: enabled' : 'Hardware signing: disabled';
  }
  if (registerKeplrBtn) registerKeplrBtn.addEventListener('click', suggestKeplrChain);
  if (wcConnectBtn) wcConnectBtn.addEventListener('click', initiateWalletConnect);
  if (wcDisconnectBtn) wcDisconnectBtn.addEventListener('click', async () => {
    await disconnectWalletConnect();
    showMessage('walletConnectStatus', 'WalletConnect session disconnected', false);
  });
  if (hwToggle) {
    hwToggle.addEventListener('change', event => {
      useHardwareSigning = Boolean(event.target.checked);
      if (hwStatus) {
        hwStatus.textContent = useHardwareSigning ? 'Hardware signing: enabled' : 'Hardware signing: disabled';
      }
    });
  }
}

function restoreSettings() {
  chrome.storage.local.get(['walletAddress', API_KEY], result => {
    const walletAddressInput = $('#walletAddress');
    const apiHostInput = $('#apiHost');

    if (result.walletAddress && walletAddressInput) {
      walletAddressInput.value = result.walletAddress;
    }
    if (result[API_KEY] && apiHostInput) {
      apiHostInput.value = result[API_KEY];
      COSMOS_SDK.config.restEndpoint = result[API_KEY] || DEFAULT_API_HOST;
      COSMOS_SDK.config.rpcEndpoint = (result[API_KEY] || DEFAULT_API_HOST).replace('1317', '26657');
    } else if (apiHostInput) {
      apiHostInput.value = DEFAULT_API_HOST;
      COSMOS_SDK.config.restEndpoint = DEFAULT_API_HOST;
      COSMOS_SDK.config.rpcEndpoint = DEFAULT_API_HOST.replace('1317', '26657');
    }
  });

  const walletAddressInput = $('#walletAddress');
  if (walletAddressInput) {
    walletAddressInput.addEventListener('change', async event => {
      chrome.storage.local.set({ walletAddress: event.target.value.trim() });
      await updateBalance();
    });
  }
}

async function initializeWallet() {
  try {
    // Check network connection first
    const networkStatus = await checkNetworkConnection();
    const statusElement = $('#networkStatus');

    if (networkStatus.connected) {
      if (statusElement) {
        statusElement.textContent = `Connected to ${networkStatus.chainId || 'Aura'} | Block: ${networkStatus.latestHeight || 'N/A'}`;
        statusElement.classList.remove('error');
      }
    } else {
      if (statusElement) {
        statusElement.textContent = `Disconnected: ${networkStatus.error || 'Network unavailable'}`;
        statusElement.classList.add('error');
      }
    }

    // Check if we have a stored private key
    const privateKeyHex = await getPrivateKey();
    if (privateKeyHex) {
      const privateKey = COSMOS_SDK.hexToBytes(privateKeyHex);
      const publicKey = await COSMOS_SDK.getPublicKey(privateKey);
      const address = COSMOS_SDK.publicKeyToAddress(publicKey);

      const walletAddressInput = $('#walletAddress');
      if (walletAddressInput && !walletAddressInput.value) {
        walletAddressInput.value = address;
        chrome.storage.local.set({ walletAddress: address });
      }

      // Validate the address format
      if (!validateCosmosAddress(address)) {
        console.warn('Generated address may be invalid:', address);
      }
    }
  } catch (error) {
    console.error('Error initializing wallet:', error);
    showMessage('walletMessage', `Initialization error: ${error.message}`, true);
  }
}

async function updateNetworkStatus() {
  const networkStatus = await checkNetworkConnection();
  const statusElement = $('#networkStatus');

  if (statusElement) {
    if (networkStatus.connected) {
      statusElement.textContent = `Connected to ${networkStatus.chainId || 'Aura'} | Block: ${networkStatus.latestHeight || 'N/A'}`;
      statusElement.classList.remove('error');
    } else {
      statusElement.textContent = `Disconnected: ${networkStatus.error || 'Network unavailable'}`;
      statusElement.classList.add('error');
    }
  }
}

document.addEventListener('DOMContentLoaded', async () => {
  try {
    // Expose Keplr/Leap provider shim for dApps (chain registration still user-driven)
    window.auraKeplrProvider = buildKeplrProvider();

    // Load theme and contacts first
    loadTheme();
    await loadContacts();
    renderContactList();

    bindActions();
    restoreSettings();

    const apiHostInput = $('#apiHost');
    if (apiHostInput) {
      const host = await getApiHost();
      apiHostInput.value = host;
      COSMOS_SDK.config.restEndpoint = host;
      COSMOS_SDK.config.rpcEndpoint = host.replace('1317', '26657');
    }

    await initializeWallet();

    // Safe async calls with error handling
    await safeAsyncCall(updateBalance, 'balance update');
    await safeAsyncCall(refreshPools, 'pools refresh');
    await safeAsyncCall(refreshMatches, 'prices refresh');
    await safeAsyncCall(updateMiningStatus, 'network status');
    await safeAsyncCall(refreshMinerStats, 'staking stats');
    await safeAsyncCall(refreshTradeHistory, 'transaction history');

    // Set up auto-refresh every 30 seconds
    setInterval(async () => {
      await safeAsyncCall(updateNetworkStatus, 'network status');
      await safeAsyncCall(updateBalance, 'balance update');
      await safeAsyncCall(refreshPools, 'pools refresh');
      await safeAsyncCall(refreshMatches, 'prices refresh');
      await safeAsyncCall(updateMiningStatus, 'network status');
    }, 30000);

    console.log('Aura Browser Wallet initialized successfully');
  } catch (error) {
    console.error('Fatal initialization error:', error);
    showMessage('walletMessage', `Failed to initialize wallet: ${error.message}`, true);
  }
});

export {
  suggestKeplrChain,
  initiateWalletConnect
};

/**
 * Safe async call wrapper with error handling
 */
async function safeAsyncCall(fn, operationName) {
  try {
    await fn();
  } catch (error) {
    console.error(`Error during ${operationName}:`, error);
    // Don't show UI errors for background operations to avoid spam
  }
}

// ==================== ADDRESS BOOK ====================
const CONTACTS_KEY = 'walletContacts';
let contacts = [];

async function loadContacts() {
  return new Promise(resolve => {
    chrome.storage.local.get([CONTACTS_KEY], result => {
      contacts = result[CONTACTS_KEY] || [];
      resolve(contacts);
    });
  });
}

async function saveContacts() {
  return new Promise(resolve => {
    chrome.storage.local.set({ [CONTACTS_KEY]: contacts }, resolve);
  });
}

function renderContactList() {
  const list = $('#addressBookList');
  if (!list) return;

  if (!contacts.length) {
    list.innerHTML = '<div class="list-placeholder">No saved contacts yet.</div>';
    return;
  }

  list.innerHTML = contacts.map((c, i) => `
    <div class="entry" data-index="${i}">
      <div class="contact-info">
        <div class="contact-name">${escapeHtml(c.name)}</div>
        <div class="contact-address">${c.address}</div>
      </div>
      <div class="contact-actions">
        <button class="secondary contact-use" data-address="${c.address}">Use</button>
        <button class="danger-outline contact-delete" data-index="${i}">✗</button>
      </div>
    </div>
  `).join('');

  // Bind use buttons
  list.querySelectorAll('.contact-use').forEach(btn => {
    btn.addEventListener('click', e => {
      const addr = e.target.dataset.address;
      const input = $('#recipientAddress');
      if (input) input.value = addr;
    });
  });

  // Bind delete buttons
  list.querySelectorAll('.contact-delete').forEach(btn => {
    btn.addEventListener('click', async e => {
      const idx = parseInt(e.target.dataset.index, 10);
      contacts.splice(idx, 1);
      await saveContacts();
      renderContactList();
      showMessage('addressBookMessage', 'Contact deleted');
    });
  });
}

async function handleAddContact(e) {
  e.preventDefault();
  const name = $('#contactName').value.trim();
  const address = $('#contactAddress').value.trim();

  if (!name || !address) {
    showMessage('addressBookMessage', 'Name and address required', true);
    return;
  }

  if (!validateCosmosAddress(address)) {
    showMessage('addressBookMessage', 'Invalid address format', true);
    return;
  }

  if (contacts.find(c => c.address === address)) {
    showMessage('addressBookMessage', 'Contact already exists', true);
    return;
  }

  contacts.push({ name, address });
  await saveContacts();
  renderContactList();
  $('#contactName').value = '';
  $('#contactAddress').value = '';
  showMessage('addressBookMessage', `${name} saved to contacts`);
}

function openContactPicker() {
  const modal = $('#contactPickerModal');
  const list = $('#contactPickerList');

  if (!contacts.length) {
    list.innerHTML = '<div class="list-placeholder">No contacts saved. Add some in Address Book.</div>';
  } else {
    list.innerHTML = contacts.map(c => `
      <div class="entry contact-pick" data-address="${c.address}">
        <div class="contact-name">${escapeHtml(c.name)}</div>
        <div class="contact-address">${c.address}</div>
      </div>
    `).join('');

    list.querySelectorAll('.contact-pick').forEach(el => {
      el.addEventListener('click', () => {
        const addr = el.dataset.address;
        $('#recipientAddress').value = addr;
        modal.classList.add('hidden');
      });
    });
  }

  modal.classList.remove('hidden');
}

// ==================== THEME TOGGLE ====================
const THEME_KEY = 'walletTheme';

function loadTheme() {
  chrome.storage.local.get([THEME_KEY], result => {
    const theme = result[THEME_KEY] || 'dark';
    applyTheme(theme);
  });
}

function applyTheme(theme) {
  const btn = $('#themeToggle');
  if (theme === 'light') {
    document.body.classList.add('light-theme');
    if (btn) btn.textContent = '☀️';
  } else {
    document.body.classList.remove('light-theme');
    if (btn) btn.textContent = '🌙';
  }
}

function toggleTheme() {
  const isLight = document.body.classList.contains('light-theme');
  const newTheme = isLight ? 'dark' : 'light';
  applyTheme(newTheme);
  chrome.storage.local.set({ [THEME_KEY]: newTheme });
}

// ==================== TRANSACTION PREVIEW ====================
let pendingTxCallback = null;

function showTxPreview(details, onConfirm) {
  const modal = $('#txPreviewModal');
  $('#previewAction').textContent = details.action || 'Transfer';
  $('#previewFrom').textContent = truncateAddress(details.from);
  $('#previewTo').textContent = truncateAddress(details.to);
  $('#previewAmount').textContent = `${details.amount} ${details.denom || 'AURA'}`;
  $('#previewGas').textContent = details.gas || '~0.005 AURA';

  const warning = $('#previewWarning');
  if (details.warning) {
    warning.classList.remove('hidden');
    $('#previewWarningText').textContent = details.warning;
  } else {
    warning.classList.add('hidden');
  }

  pendingTxCallback = onConfirm;
  modal.classList.remove('hidden');
}

function closeTxPreview(confirmed) {
  $('#txPreviewModal').classList.add('hidden');
  if (confirmed && pendingTxCallback) {
    pendingTxCallback();
  }
  pendingTxCallback = null;
}

function truncateAddress(addr) {
  if (!addr || addr.length < 20) return addr || '--';
  return `${addr.slice(0, 12)}...${addr.slice(-8)}`;
}

// ==================== CSV EXPORT ====================
async function exportTransactionsCsv() {
  const address = $('#walletAddress').value.trim();
  if (!address) {
    showMessage('tradeMessage', 'Enter wallet address first', true);
    return;
  }

  try {
    showMessage('tradeMessage', 'Fetching transactions for export...');
    const url = `${COSMOS_SDK.config.restEndpoint}/cosmos/tx/v1beta1/txs?events=message.sender='${address}'&order_by=ORDER_BY_DESC&limit=100`;
    const res = await fetch(url);

    if (!res.ok) throw new Error('Failed to fetch transactions');

    const data = await res.json();
    const txs = data.txs || [];

    if (!txs.length) {
      showMessage('tradeMessage', 'No transactions to export', true);
      return;
    }

    // Build CSV
    const headers = ['Hash', 'Height', 'Timestamp', 'Type', 'Fee', 'Status'];
    const rows = txs.map(tx => {
      const msgs = tx.body?.messages || [];
      const msgType = msgs[0]?.['@type']?.split('.').pop() || 'Unknown';
      const fee = tx.auth_info?.fee?.amount?.[0]?.amount || '0';
      return [
        tx.txhash,
        tx.height || '',
        tx.timestamp || '',
        msgType,
        fee,
        tx.code === 0 ? 'Success' : 'Failed'
      ].map(v => `"${v}"`).join(',');
    });

    const csv = [headers.join(','), ...rows].join('\n');

    // Download
    const blob = new Blob([csv], { type: 'text/csv' });
    const urlObj = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = urlObj;
    a.download = `aura-transactions-${Date.now()}.csv`;
    a.click();
    URL.revokeObjectURL(urlObj);

    showMessage('tradeMessage', `Exported ${txs.length} transactions`);
  } catch (error) {
    showMessage('tradeMessage', `Export failed: ${error.message}`, true);
  }
}

// ==================== UTILITY ====================
function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}
