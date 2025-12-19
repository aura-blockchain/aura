import SignClient from '@walletconnect/sign-client';
import { buildKeplrProvider } from './keplr-provider';

let wcClient = null;
let wcApproval = null;
let wcSession = null;
const statusListeners = new Set();
const SESSION_KEY = 'aura_wcv2_session';
let customProvider = null;

function hasChromeStorage() {
  return typeof chrome !== 'undefined' && chrome?.storage?.local;
}

async function persistSession(session) {
  try {
    if (!session) return;
    if (hasChromeStorage()) {
      await new Promise(resolve => chrome.storage.local.set({ [SESSION_KEY]: session }, resolve));
    } else if (typeof localStorage !== 'undefined') {
      localStorage.setItem(SESSION_KEY, JSON.stringify(session));
    }
  } catch (err) {
    console.warn('Unable to persist WalletConnect session', err);
  }
}

async function loadPersistedSession() {
  try {
    if (hasChromeStorage()) {
      return await new Promise(resolve => chrome.storage.local.get([SESSION_KEY], data => resolve(data?.[SESSION_KEY])));
    }
    if (typeof localStorage !== 'undefined') {
      const raw = localStorage.getItem(SESSION_KEY);
      return raw ? JSON.parse(raw) : null;
    }
    return null;
  } catch (err) {
    console.warn('Unable to load WalletConnect session', err);
    return null;
  }
}

async function clearPersistedSession() {
  try {
    if (hasChromeStorage()) {
      await new Promise(resolve => chrome.storage.local.remove([SESSION_KEY], resolve));
    } else if (typeof localStorage !== 'undefined') {
      localStorage.removeItem(SESSION_KEY);
    }
  } catch (err) {
    console.warn('Unable to clear WalletConnect session', err);
  }
}

function notifyStatus(status) {
  statusListeners.forEach(fn => fn(status));
}

export function onWalletConnectStatus(listener) {
  statusListeners.add(listener);
  return () => statusListeners.delete(listener);
}

async function ensureClient(projectId = 'local-dev') {
  if (wcClient) return wcClient;
  wcClient = await SignClient.init({
    projectId,
    metadata: {
      name: 'Aura Wallet',
      description: 'Aura browser wallet',
      url: 'https://aura.dev',
      icons: [],
    },
  });
  // Attempt to restore if session exists in client storage
  try {
    const persisted = await loadPersistedSession();
    if (persisted?.topic && wcClient.session?.get) {
      const existing = wcClient.session.get(persisted.topic);
      if (existing) {
        wcSession = existing;
        notifyStatus('WalletConnect session restored');
      }
    }
  } catch (err) {
    console.warn('WalletConnect restore skipped', err);
  }
  return wcClient;
}

function setupRequestHandlers(provider) {
  if (!wcClient?.on) return;
  wcClient.on('session_request', async (event) => {
    const { topic, id, params } = event;
    const { request, chainId } = params;
    const signerProvider = customProvider || provider;
    notifyStatus(`Incoming request: ${request.method} (${chainId})`);
    try {
      if (request.method === 'cosmos_signDirect' && signerProvider?.signDirect) {
        const res = await signerProvider.signDirect(request.params.signerAddress, request.params.signDoc);
        await wcClient.respond({ topic, response: { id, jsonrpc: '2.0', result: res } });
        notifyStatus(`Request handled: signDirect (${chainId})`);
        return;
      }
      if (request.method === 'cosmos_signAmino' && signerProvider?.signAmino) {
        const res = await signerProvider.signAmino(request.params.signerAddress, request.params.signDoc);
        await wcClient.respond({ topic, response: { id, jsonrpc: '2.0', result: res } });
        notifyStatus(`Request handled: signAmino (${chainId})`);
        return;
      }
      await wcClient.respond({
        topic,
        response: { id, jsonrpc: '2.0', error: { code: -32000, message: 'Unsupported WalletConnect request' } }
      });
      notifyStatus('Rejected unsupported request');
    } catch (err) {
      await wcClient.respond({
        topic,
        response: { id, jsonrpc: '2.0', error: { code: -32001, message: err?.message || 'Signing failed' } }
      });
      notifyStatus(`Request failed: ${err?.message}`);
    }
  });

  wcClient.on('session_delete', () => {
    wcSession = null;
    clearPersistedSession();
    notifyStatus('Session closed');
  });
}

export async function connectWalletConnect(options = {}) {
  const client = await ensureClient(options.projectId);
  const provider = customProvider || buildKeplrProvider();
  setupRequestHandlers(provider);
  const { uri, approval } = await client.connect({
    requiredNamespaces: {
      cosmos: {
        methods: ['cosmos_signDirect', 'cosmos_signAmino'],
        chains: [`cosmos:${provider.chainId || 'aura-testnet-1'}`],
        events: [],
      },
    },
  });
  wcApproval = approval;
  approval.then(session => {
    wcSession = session;
    persistSession(session);
    notifyStatus('Session approved');
  }).catch(err => {
    notifyStatus(err?.message || 'Session rejected');
  });
  return { uri, approval };
}

export async function disconnectWalletConnect() {
  if (wcClient && wcSession?.topic && wcClient.disconnect) {
    await wcClient.disconnect({
      topic: wcSession.topic,
      reason: { code: 6000, message: 'User disconnected' }
    });
  }
  wcSession = null;
  await clearPersistedSession();
  notifyStatus('Session closed');
}

export function getWalletConnectClient() {
  return wcClient;
}

export function getWalletConnectApproval() {
  return wcApproval;
}

export function getWalletConnectSession() {
  return wcSession;
}

export async function restoreWalletConnectSession(projectId = 'local-dev') {
  const client = await ensureClient(projectId);
  const persisted = await loadPersistedSession();
  if (persisted?.topic && client.session?.get) {
    const existing = client.session.get(persisted.topic);
    if (existing) {
      wcSession = existing;
      notifyStatus('Session restored');
      return existing;
    }
  }
  return null;
}

export function setWalletConnectProvider(provider) {
  customProvider = provider;
}
