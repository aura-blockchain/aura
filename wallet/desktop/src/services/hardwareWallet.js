/**
 * Hardware Wallet Support for Aura Desktop Wallet
 * Supports Ledger (via Electron node-hid) and Trezor.
 */

const DEFAULT_PATH = "m/44'/118'/0'/0/0";

class HardwareWalletService {
  constructor() {
    this.connectedDevice = null;
    this.ledgerTransport = null;
    this.ledgerApp = null;
    this.connectionListeners = [];
    this.bech32Prefix = 'aura';
  }

  onConnectionChange(listener) {
    this.connectionListeners.push(listener);
    return () => {
      this.connectionListeners = this.connectionListeners.filter(l => l !== listener);
    };
  }

  notifyConnectionChange() {
    this.connectionListeners.forEach(l => l(this.connectedDevice));
  }

  async loadLedgerDeps() {
    // In Electron, we use @ledgerhq/hw-transport-node-hid for native USB access
    // Dynamic import for bundler compatibility
    if (this.TransportNodeHID && this.LedgerCosmosApp) return;

    try {
      const [transportMod, cosmosMod] = await Promise.all([
        import('@ledgerhq/hw-transport-node-hid'),
        import('@ledgerhq/hw-app-cosmos')
      ]);
      this.TransportNodeHID = transportMod.default;
      this.LedgerCosmosApp = cosmosMod.default;
    } catch (err) {
      // Fallback to WebHID for Electron with context isolation
      const [transportMod, cosmosMod] = await Promise.all([
        import('@ledgerhq/hw-transport-webhid'),
        import('@ledgerhq/hw-app-cosmos')
      ]);
      this.TransportNodeHID = transportMod.default;
      this.LedgerCosmosApp = cosmosMod.default;
    }
  }

  async detectDevices() {
    await this.loadLedgerDeps();
    const devices = [];

    try {
      if (this.TransportNodeHID.list) {
        const ledgerDevices = await this.TransportNodeHID.list();
        devices.push(...ledgerDevices.map(d => ({
          type: 'ledger',
          device: d,
          connected: true
        })));
      }
    } catch (err) {
      console.warn('Ledger detection failed:', err.message);
    }

    return devices;
  }

  async connect(type = 'ledger', timeoutMs = 30000) {
    if (this.connectedDevice && this.connectedDevice.type === type) {
      return this.connectedDevice;
    }

    if (this.connectedDevice) {
      await this.disconnect();
    }

    if (type === 'ledger') {
      return this.connectLedger(timeoutMs);
    }
    if (type === 'trezor') {
      return this.connectTrezor();
    }

    throw new Error(`Unsupported device type: ${type}`);
  }

  async connectLedger(timeoutMs = 30000) {
    await this.loadLedgerDeps();

    const timeoutPromise = new Promise((_, reject) => {
      setTimeout(() => reject(new Error('Connection timed out. Ensure Ledger is connected and Cosmos app is open.')), timeoutMs);
    });

    try {
      const transport = await Promise.race([
        this.TransportNodeHID.create(),
        timeoutPromise
      ]);

      this.ledgerTransport = transport;
      this.ledgerApp = new this.LedgerCosmosApp(transport);
      this.connectedDevice = {
        type: 'ledger',
        name: 'Ledger',
        device: transport.device
      };

      this.notifyConnectionChange();
      return this.connectedDevice;
    } catch (err) {
      console.error('Ledger connection failed:', err);
      this.reset();
      throw new Error(err.message || 'Failed to connect to Ledger');
    }
  }

  async connectTrezor() {
    // Trezor Connect integration for desktop
    if (!this.TrezorConnect) {
      const mod = await import('trezor-connect');
      this.TrezorConnect = mod.default;
      this.TrezorConnect.manifest({
        email: 'support@aura.network',
        appUrl: 'https://aura.network'
      });
    }

    const response = await this.TrezorConnect.getAddress({
      path: DEFAULT_PATH,
      coin: 'Cosmos',
      showOnTrezor: false
    });

    if (!response.success) {
      throw new Error(response.payload?.error || 'Trezor connection failed');
    }

    this.connectedDevice = {
      type: 'trezor',
      name: 'Trezor',
      device: { productName: 'Trezor' }
    };

    this.notifyConnectionChange();
    return this.connectedDevice;
  }

  reset() {
    this.connectedDevice = null;
    this.ledgerTransport = null;
    this.ledgerApp = null;
    this.notifyConnectionChange();
  }

  async disconnect() {
    try {
      if (this.ledgerTransport) {
        await this.ledgerTransport.close();
      }
    } catch (err) {
      console.warn('Error closing transport:', err.message);
    } finally {
      this.reset();
    }
  }

  async getAddress(path = DEFAULT_PATH, confirm = true) {
    if (!this.connectedDevice) {
      throw new Error('No device connected');
    }

    if (this.connectedDevice.type === 'ledger') {
      return this.getLedgerAddress(path, confirm);
    }
    if (this.connectedDevice.type === 'trezor') {
      return this.getTrezorAddress(path, confirm);
    }

    throw new Error('Unsupported device');
  }

  async getLedgerAddress(path = DEFAULT_PATH, confirm = true) {
    if (!this.ledgerApp) {
      await this.connect('ledger');
    }

    try {
      const response = await this.ledgerApp.getAddress(path, this.bech32Prefix, confirm);
      return {
        address: response.bech32_address,
        publicKey: Buffer.from(response.publicKey).toString('hex')
      };
    } catch (err) {
      console.error('Failed to get Ledger address:', err);
      throw new Error(err.message || 'Failed to retrieve address');
    }
  }

  async getTrezorAddress(path = DEFAULT_PATH, confirm = true) {
    const response = await this.TrezorConnect.getAddress({
      path: path,
      coin: 'Cosmos',
      showOnTrezor: confirm
    });

    if (!response.success) {
      throw new Error(response.payload?.error || 'Failed to get address');
    }

    return {
      address: response.payload.address,
      publicKey: response.payload.publicKey || ''
    };
  }

  async signTransaction(txBytes, path = DEFAULT_PATH) {
    if (!this.connectedDevice) {
      throw new Error('No device connected');
    }

    if (this.connectedDevice.type === 'ledger') {
      return this.signWithLedger(txBytes, path);
    }
    if (this.connectedDevice.type === 'trezor') {
      return this.signWithTrezor(txBytes, path);
    }

    throw new Error('Unsupported device');
  }

  async signWithLedger(txBytes, path = DEFAULT_PATH) {
    if (!this.ledgerApp) {
      await this.connect('ledger');
    }

    try {
      const bytes = Buffer.isBuffer(txBytes) ? txBytes : Buffer.from(txBytes);
      const response = await this.ledgerApp.sign(path, bytes);
      return {
        signature: Buffer.from(response.signature).toString('hex')
      };
    } catch (err) {
      console.error('Ledger signing failed:', err);
      throw new Error(err.message || 'Failed to sign transaction');
    }
  }

  async signWithTrezor(txBytes, path = DEFAULT_PATH) {
    const bytes = Buffer.isBuffer(txBytes) ? txBytes : Buffer.from(txBytes);
    const hex = bytes.toString('hex');

    const response = await this.TrezorConnect.signTransaction({
      path: path,
      rawTxHex: hex
    });

    if (!response.success) {
      throw new Error(response.payload?.error || 'Trezor signing failed');
    }

    return {
      signature: response.payload.signature || response.payload.signedTransaction || ''
    };
  }

  isConnected() {
    return this.connectedDevice !== null;
  }

  getDeviceInfo() {
    return this.connectedDevice;
  }

  async verifyAddress(path = DEFAULT_PATH) {
    try {
      const result = await this.getAddress(path, true);
      return Boolean(result?.address);
    } catch {
      return false;
    }
  }
}

export const hardwareWalletService = new HardwareWalletService();
export default HardwareWalletService;
