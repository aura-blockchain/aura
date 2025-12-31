/**
 * Hardware Wallet Support for Aura Web Wallet
 * Supports Ledger via WebHID (Chrome/Edge)
 */

const AURA_PREFIX = 'aura';
const DEFAULT_PATH = "m/44'/118'/0'/0/0";
const ALLOWED_FEE_DENOMS = ['uaura'];
const MAX_ACCOUNT_INDEX = 4;

class HardwareWalletService {
  constructor() {
    this.transport = null;
    this.app = null;
    this.connected = false;
    this.connectionListeners = [];
  }

  onConnectionChange(listener) {
    this.connectionListeners.push(listener);
    return () => {
      this.connectionListeners = this.connectionListeners.filter(l => l !== listener);
    };
  }

  notifyConnectionChange() {
    this.connectionListeners.forEach(l => l(this.connected));
  }

  async loadLedgerDeps() {
    if (this.TransportWebHID && this.CosmosApp) return;

    try {
      const [transportMod, cosmosMod] = await Promise.all([
        import('https://esm.sh/@ledgerhq/hw-transport-webhid@6.29.15'),
        import('https://esm.sh/@ledgerhq/hw-app-cosmos@6.29.15')
      ]);
      this.TransportWebHID = transportMod.default;
      this.CosmosApp = cosmosMod.default;
    } catch (err) {
      throw new Error(`Failed to load Ledger libraries: ${err.message}`);
    }
  }

  async connect(timeoutMs = 30000) {
    if (this.connected) {
      return { connected: true };
    }

    await this.loadLedgerDeps();

    const timeoutPromise = new Promise((_, reject) => {
      setTimeout(() => reject(new Error('Connection timed out')), timeoutMs);
    });

    try {
      this.transport = await Promise.race([
        this.TransportWebHID.create(),
        timeoutPromise
      ]);
      this.app = new this.CosmosApp(this.transport);
      this.connected = true;
      this.notifyConnectionChange();
      return { connected: true };
    } catch (err) {
      this.reset();
      throw new Error(err.message || 'Failed to connect to Ledger');
    }
  }

  reset() {
    this.transport = null;
    this.app = null;
    this.connected = false;
    this.notifyConnectionChange();
  }

  async disconnect() {
    try {
      if (this.transport) {
        await this.transport.close();
      }
    } catch (err) {
      console.warn('Error closing transport:', err.message);
    } finally {
      this.reset();
    }
  }

  normalizePath(path, maxAccount = MAX_ACCOUNT_INDEX) {
    const sanitized = path.startsWith('m/') ? path.slice(2) : path;
    const segments = sanitized.split('/');
    if (segments.length !== 5) {
      throw new Error(`Invalid derivation path: ${path}`);
    }
    const accountSeg = segments[2];
    const account = parseInt(accountSeg.replace("'", ''), 10);
    if (isNaN(account) || account > maxAccount) {
      throw new Error(`Account index exceeds maximum (${maxAccount})`);
    }
    return segments.map((seg, idx) => {
      const hardened = seg.endsWith("'");
      const val = parseInt(hardened ? seg.slice(0, -1) : seg, 10);
      if (isNaN(val)) {
        throw new Error(`Invalid path segment: ${seg}`);
      }
      return hardened || idx < 3 ? `${val}'` : `${val}`;
    }).join('/');
  }

  async getAddress(accountIndex = 0, confirm = true) {
    if (!this.app) {
      await this.connect();
    }
    if (accountIndex < 0 || accountIndex > MAX_ACCOUNT_INDEX) {
      throw new Error('Account index must be between 0 and 4');
    }
    const path = this.normalizePath(`m/44'/118'/${accountIndex}'/0/0`);
    try {
      const response = await this.app.getAddress(path, AURA_PREFIX, confirm);
      this.validateBech32Prefix(response.address);
      return {
        address: response.address,
        publicKey: response.publicKey
      };
    } catch (err) {
      throw new Error(err.message || 'Failed to get address');
    }
  }

  validateBech32Prefix(address) {
    if (!address.startsWith(AURA_PREFIX + '1')) {
      throw new Error(`Invalid address prefix: expected ${AURA_PREFIX}`);
    }
  }

  validateSignDoc(signDoc, expectedChainId) {
    if (!signDoc?.chain_id) {
      throw new Error('chain_id is required for Ledger signing');
    }
    if (expectedChainId && signDoc.chain_id !== expectedChainId) {
      throw new Error(`Chain-id mismatch (expected ${expectedChainId})`);
    }
    const fee = signDoc?.fee || {};
    if (!fee.gas || isNaN(Number(fee.gas)) || Number(fee.gas) <= 0) {
      throw new Error('Invalid gas value');
    }
    if (!Array.isArray(fee.amount) || !fee.amount.length) {
      throw new Error('Fee amount required');
    }
    fee.amount.forEach(coin => {
      if (!ALLOWED_FEE_DENOMS.includes(coin.denom)) {
        throw new Error(`Fee denom ${coin.denom} not permitted`);
      }
    });
  }

  async signAmino(signDoc, accountIndex = 0, expectedChainId = null) {
    if (!this.app) {
      await this.connect();
    }
    this.validateSignDoc(signDoc, expectedChainId);
    const path = this.normalizePath(`m/44'/118'/${accountIndex}'/0/0`);
    try {
      const response = await this.app.sign(path, JSON.stringify(signDoc));
      return {
        signature: response.signature
      };
    } catch (err) {
      throw new Error(err.message || 'Failed to sign transaction');
    }
  }

  isConnected() {
    return this.connected;
  }

  static isSupported() {
    return typeof navigator !== 'undefined' && !!navigator.hid;
  }
}

const hardwareWalletService = new HardwareWalletService();

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { HardwareWalletService, hardwareWalletService };
}
