/**
 * Hardware Wallet Support for Aura Browser Extension
 * Supports Ledger (WebHID) and placeholder hooks for Trezor/Keystone.
 */
const DEFAULT_PATH = "m/44'/118'/0'/0/0";

class HardwareWalletManager {
  constructor() {
    this.connectedDevice = null;
    this.ledgerTransport = null;
    this.ledgerApp = null;
    this.trezor = null;
    this.TrezorConnect = null;
    this.trezorManifestSet = false;
    this.chainConfig = null;
    this.keystoneSdk = null;
  }

  async loadChainConfig() {
    if (this.chainConfig) return this.chainConfig;
    const mod = await import('../config/chain.js');
    const cfg = mod.CHAIN_CONFIG || mod.default?.CHAIN_CONFIG || mod;
    this.chainConfig = cfg;
    return cfg;
  }

  async loadLedgerDeps() {
    if (this.LedgerTransport && this.LedgerCosmosApp) {
      return;
    }
    const [{ default: TransportWebHID }, { default: LedgerCosmosApp }] = await Promise.all([
      import('@ledgerhq/hw-transport-webhid'),
      import('@ledgerhq/hw-app-cosmos')
    ]);
    this.LedgerTransport = TransportWebHID;
    this.LedgerCosmosApp = LedgerCosmosApp;
  }

  async loadTrezorDeps() {
    if (this.TrezorConnect) {
      return;
    }
    const { default: TrezorConnect } = await import('trezor-connect');
    this.TrezorConnect = TrezorConnect;
    if (TrezorConnect?.manifest && !this.trezorManifestSet) {
      const cfg = await this.loadChainConfig();
      TrezorConnect.manifest({
        email: 'dev@aura.network',
        appUrl: cfg?.explorerUrl || 'https://aura.network',
      });
      this.trezorManifestSet = true;
    }
  }

  async loadKeystoneDeps() {
    if (this.keystoneSdk) {
      return;
    }
    const { default: KeystoneSDK } = await import('@keystonehq/keystone-sdk');
    this.keystoneSdk = new KeystoneSDK({
      origin: 'aura-browser-extension',
    });
  }

  pathToArray(path = DEFAULT_PATH) {
    return path
      .split('/')
      .filter(Boolean)
      .filter(part => part !== 'm')
      .map(part => {
        const hardened = part.endsWith("'");
        const clean = hardened ? part.slice(0, -1) : part;
        const num = parseInt(clean, 10);
        return hardened ? ((0x80000000 | num) >>> 0) : num;
      });
  }

  async detectDevices() {
    await this.loadLedgerDeps();
    const devices = [];
    try {
      const ledgerDevices = await this.LedgerTransport.list();
      devices.push(
        ...ledgerDevices.map(device => ({
          device,
          type: 'ledger',
          connected: true
        }))
      );
    } catch (err) {
      console.error('Failed to detect Ledger devices:', err);
    }
    // Trezor detection will be handled via connect request; Trezor Connect does not expose a list() API.
    return devices;
  }

  async requestDevice(type = 'ledger', timeoutMs = 30000) {
    if (type === 'trezor') {
      return this.connectTrezor();
    }
    if (type === 'keystone') {
      return this.connectKeystone();
    }
    await this.loadLedgerDeps();

    // Create timeout promise for WebHID request
    const timeoutPromise = new Promise((_, reject) => {
      setTimeout(() => reject(new Error('Connection timed out. Click Reset and try again.')), timeoutMs);
    });

    try {
      const transport = await Promise.race([
        this.LedgerTransport.create(),
        timeoutPromise
      ]);
      this.ledgerTransport = transport;
      this.ledgerApp = new this.LedgerCosmosApp(transport);
      this.connectedDevice = { type: 'ledger', device: transport.device };
      return this.connectedDevice;
    } catch (err) {
      console.error('Failed to request Ledger device:', err);
      this.reset();
      if (err.message?.includes('timed out')) {
        throw err;
      }
      throw new Error('User cancelled or device not available');
    }
  }

  reset() {
    this.connectedDevice = null;
    this.ledgerTransport = null;
    this.ledgerApp = null;
    this.trezor = null;
  }

  async connect(type = 'ledger') {
    if (this.connectedDevice) {
      if (this.connectedDevice.type === type) {
        return this.connectedDevice;
      }
      await this.disconnect();
    }
    return this.requestDevice(type);
  }

  async connectTrezor() {
    await this.loadTrezorDeps();
    const cfg = await this.loadChainConfig();
    const hrp = cfg?.bech32Prefix || 'cosmos';
    const resp = this.TrezorConnect.cosmosGetAddress
      ? await this.TrezorConnect.cosmosGetAddress({
        path: this.pathToArray(DEFAULT_PATH),
        showOnTrezor: false,
        hrp,
      })
      : await this.TrezorConnect.getAddress({
        path: DEFAULT_PATH,
        coin: 'Cosmos',
        showOnTrezor: false,
      });
    if (!resp?.success) {
      throw new Error(resp?.payload?.error || 'Unable to reach Trezor');
    }
    this.connectedDevice = { type: 'trezor', device: { productName: 'Trezor' } };
    return this.connectedDevice;
  }

  async connectKeystone() {
    await this.loadKeystoneDeps();
    this.connectedDevice = { type: 'keystone', device: { productName: 'Keystone' } };
    return this.connectedDevice;
  }

  async disconnect() {
    try {
      if (this.ledgerTransport) {
        await this.ledgerTransport.close();
      }
      if (this.trezor) {
        this.trezor = null;
      }
    } catch (err) {
      console.error('Error closing transport:', err);
    } finally {
      this.ledgerTransport = null;
      this.ledgerApp = null;
      this.connectedDevice = null;
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
      return this.getTrezorAddress(path);
    }
    if (this.connectedDevice.type === 'keystone') {
      return this.getKeystoneAddress(path);
    }
    throw new Error('Unsupported device type');
  }

  async getLedgerAddress(path = DEFAULT_PATH, confirm = true) {
    try {
      await this.loadLedgerDeps();
      if (!this.ledgerApp) {
        await this.connect();
      }
      const cfg = await this.loadChainConfig();
      const hrp = cfg?.bech32Prefix || 'cosmos';
      const resp = await this.ledgerApp.getAddress(path, hrp, confirm);
      return {
        address: resp.bech32_address,
        publicKey: Buffer.from(resp.publicKey).toString('hex')
      };
    } catch (err) {
      console.error('Failed to get Ledger address:', err);
      throw new Error(err.message || 'Failed to retrieve address from Ledger');
    }
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
    if (this.connectedDevice.type === 'keystone') {
      return this.signWithKeystone(txBytes, path);
    }
    throw new Error('Unsupported device type');
  }

  async signWithLedger(txBytes, path = DEFAULT_PATH) {
    try {
      await this.loadLedgerDeps();
      if (!this.ledgerApp) {
        await this.connect();
      }
      const bytes = Buffer.isBuffer(txBytes) ? txBytes : Buffer.from(txBytes);
      const { signature } = await this.ledgerApp.sign(path, bytes);
      return {
        signature: Buffer.from(signature).toString('hex')
      };
    } catch (err) {
      console.error('Failed to sign with Ledger:', err);
      throw new Error(err.message || 'Failed to sign transaction');
    }
  }

  // Placeholder for future Trezor integration
  async signWithTrezor(txBytes, path = DEFAULT_PATH) {
    await this.loadTrezorDeps();
    if (!this.connectedDevice) {
      await this.connect('trezor');
    }
    const trezorPath = this.pathToArray(path);
    const bytes = Buffer.isBuffer(txBytes) ? txBytes : Buffer.from(txBytes);
    const hex = Buffer.from(bytes).toString('hex');
    const resp = this.TrezorConnect.cosmosSignTx
      ? await this.TrezorConnect.cosmosSignTx({ path: trezorPath, rawTx: hex })
      : await this.TrezorConnect.signTransaction({ path: trezorPath, rawTxHex: hex });
    if (!resp?.success) {
      throw new Error(resp?.payload?.error || 'Trezor signing failed');
    }
    const payload = resp.payload || resp;
    const signatureHex = payload.signatureHex || payload.signature || payload.sig || payload.signedTransaction;
    return { signature: signatureHex || '' };
  }

  async signWithKeystone(txBytes, path = DEFAULT_PATH) {
    await this.loadKeystoneDeps();
    const cfg = await this.loadChainConfig();
    const hrp = cfg?.bech32Prefix || 'cosmos';
    const data = Buffer.isBuffer(txBytes) ? txBytes : Buffer.from(txBytes);
    const result = await this.keystoneSdk.sign({
      requestId: Date.now().toString(),
      signData: data.toString('hex'),
      xfp: '',
      derivationPath: path,
      address: '',
      hrp,
      isTestNet: false,
    });
    return { signature: result?.signature || '' };
  }

  async getTrezorAddress(path = DEFAULT_PATH) {
    await this.loadTrezorDeps();
    if (!this.connectedDevice) {
      await this.connect('trezor');
    }
    const trezorPath = this.pathToArray(path);
    const cfg = await this.loadChainConfig();
    const hrp = cfg?.bech32Prefix || 'cosmos';
    const resp = this.TrezorConnect.cosmosGetAddress
      ? await this.TrezorConnect.cosmosGetAddress({ path: trezorPath, showOnTrezor: true, hrp })
      : await this.TrezorConnect.getAddress({ path, coin: 'Cosmos', showOnTrezor: true });
    if (!resp?.success) {
      throw new Error(resp?.payload?.error || 'Trezor address retrieval failed');
    }
    const payload = resp.payload || resp;
    const address = payload.address || payload.bech32_address || payload.bech32;
    const publicKey = payload.publicKey || payload.public_key || payload.public_key_hex;
    return {
      address,
      publicKey: typeof publicKey === 'string' ? publicKey : Buffer.from(publicKey || []).toString('hex'),
    };
  }

  async getKeystoneAddress(path = DEFAULT_PATH) {
    await this.loadKeystoneDeps();
    const cfg = await this.loadChainConfig();
    const hrp = cfg?.bech32Prefix || 'cosmos';
    const qrCode = this.keystoneSdk.getConnectedWalletAddress({
      derivationPath: path,
      xfp: '',
      hrp,
      isTestNet: false,
    });
    this.connectedDevice = { type: 'keystone', device: { productName: 'Keystone' } };
    return {
      address: qrCode?.address || '',
      publicKey: qrCode?.publicKey || '',
    };
  }

  async verifyAddress(path = DEFAULT_PATH) {
    const addr = this.connectedDevice?.type === 'trezor'
      ? await this.getTrezorAddress(path)
      : this.connectedDevice?.type === 'keystone'
        ? await this.getKeystoneAddress(path)
        : await this.getLedgerAddress(path, true);
    return Boolean(addr?.address);
  }

  isConnected() {
    return this.connectedDevice !== null;
  }

  getDeviceInfo() {
    if (!this.connectedDevice) {
      return null;
    }
    return {
      type: this.connectedDevice.type,
      name: this.connectedDevice.type === 'ledger' ? 'Ledger' : this.connectedDevice.type === 'trezor' ? 'Trezor' : 'Keystone',
      manufacturer: this.connectedDevice.device?.manufacturerName,
      product: this.connectedDevice.device?.productName
    };
  }
}

export default HardwareWalletManager;
