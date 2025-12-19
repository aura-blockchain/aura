import SignClient from '@walletconnect/sign-client';
import AsyncStorage from '@react-native-async-storage/async-storage';
import TransactionService from './TransactionService';

const SESSION_KEY = 'aura_wcv2_session';

class WalletConnectService {
  constructor() {
    this.client = null;
    this.session = null;
    this.statusListeners = new Set();
    this.signer = null;
  }

  onStatus(listener) {
    this.statusListeners.add(listener);
    return () => this.statusListeners.delete(listener);
  }

  emitStatus(status) {
    this.statusListeners.forEach(fn => fn(status));
  }

  setSigner({ address, privateKeyHex }) {
    if (!privateKeyHex) {
      throw new Error('Private key is required for WalletConnect signing');
    }
    this.signer = { address, privateKeyHex };
  }

  clearSigner() {
    this.signer = null;
  }

  requireSigner(expectedAddress) {
    if (!this.signer?.privateKeyHex) {
      throw new Error('No signer configured for WalletConnect');
    }
    if (expectedAddress && this.signer.address && this.signer.address !== expectedAddress) {
      throw new Error('Signer address mismatch');
    }
    return this.signer;
  }

  async persistSession(session) {
    try {
      await AsyncStorage.setItem(SESSION_KEY, JSON.stringify(session));
    } catch (err) {
      this.emitStatus(`Persist failed: ${err?.message}`);
    }
  }

  async loadSession() {
    try {
      const raw = await AsyncStorage.getItem(SESSION_KEY);
      return raw ? JSON.parse(raw) : null;
    } catch (err) {
      this.emitStatus(`Load failed: ${err?.message}`);
      return null;
    }
  }

  async clearSession() {
    try {
      await AsyncStorage.removeItem(SESSION_KEY);
    } catch (err) {
      this.emitStatus(`Clear failed: ${err?.message}`);
    }
  }

  async init(projectId = 'aura-mobile') {
    if (this.client) return this.client;
    this.client = await SignClient.init({
      projectId,
      metadata: {
        name: 'Aura Mobile',
        description: 'Aura mobile wallet',
        url: 'https://aura.network',
        icons: [],
      },
    });
    this.setupHandlers();
    await this.restore();
    return this.client;
  }

  setupHandlers() {
    if (!this.client?.on) return;
    this.client.on('session_request', async (event) => {
      const { topic, id, params } = event;
      const { request, chainId } = params;
      this.emitStatus(`Incoming ${request.method} (${chainId})`);
      try {
        const signer = this.requireSigner(request.params?.signerAddress);
        if (request.method === 'cosmos_signDirect') {
          const res = await TransactionService.signDirect(request.params.signDoc, signer.privateKeyHex);
          await this.client.respond({ topic, response: { id, jsonrpc: '2.0', result: res } });
          this.emitStatus('Handled signDirect');
          return;
        }
        if (request.method === 'cosmos_signAmino') {
          const res = await TransactionService.signAmino(request.params.signDoc, signer.privateKeyHex);
          await this.client.respond({ topic, response: { id, jsonrpc: '2.0', result: res } });
          this.emitStatus('Handled signAmino');
          return;
        }
        await this.client.respond({
          topic,
          response: { id, jsonrpc: '2.0', error: { code: -32000, message: 'Unsupported method' } },
        });
        this.emitStatus('Rejected unsupported request');
      } catch (err) {
        await this.client.respond({
          topic,
          response: { id, jsonrpc: '2.0', error: { code: -32001, message: err?.message || 'Signing failed' } },
        });
        this.emitStatus(`Request failed: ${err?.message}`);
      }
    });

    this.client.on('session_delete', () => {
      this.session = null;
      this.clearSession();
      this.emitStatus('Session closed');
    });
  }

  async connect(requiredChainId = 'aura-testnet-1', projectId = 'aura-mobile') {
    const client = await this.init(projectId);
    const { uri, approval } = await client.connect({
      requiredNamespaces: {
        cosmos: {
          methods: ['cosmos_signDirect', 'cosmos_signAmino'],
          chains: [`cosmos:${requiredChainId}`],
          events: [],
        },
      },
    });
    approval.then((session) => {
      this.session = session;
      this.persistSession(session);
      this.emitStatus('Session approved');
    }).catch(err => this.emitStatus(err?.message || 'Session rejected'));
    return { uri, approval };
  }

  async restore(projectId = 'aura-mobile') {
    await this.init(projectId);
    const persisted = await this.loadSession();
    if (persisted?.topic && this.client?.session?.get) {
      const session = this.client.session.get(persisted.topic);
      if (session) {
        this.session = session;
        this.emitStatus('Session restored');
        return session;
      }
    }
    return null;
  }

  async disconnect() {
    if (this.client && this.session?.topic && this.client.disconnect) {
      await this.client.disconnect({
        topic: this.session.topic,
        reason: { code: 6000, message: 'User disconnected' },
      });
    }
    this.session = null;
    this.clearSigner();
    await this.clearSession();
    this.emitStatus('Session closed');
  }
}

export default new WalletConnectService();
