/**
 * Wallet Core Module
 * Handles wallet creation, import, export, and key management
 */

const WalletCore = {
  /**
   * Resolve a crypto implementation that works in both browser (extension)
   * and Node (tests) environments.
   */
  async getCrypto() {
    if (typeof process !== 'undefined' && process.versions?.node) {
      if (!this.nodeWebCrypto) {
        this.nodeWebCrypto = import('node:crypto')
          .catch(() => import('crypto'))
          .then(mod => mod.webcrypto)
          .catch(() => null);
      }

      const webcrypto = await this.nodeWebCrypto;
      if (webcrypto?.subtle) {
        return webcrypto;
      }
    }

    if (typeof globalThis.crypto !== 'undefined' && globalThis.crypto.subtle) {
      return globalThis.crypto;
    }

    throw new Error('Crypto API unavailable');
  },

  /**
   * Node.js crypto fallback (used in tests and CLI tooling).
   */
  async getNodeCrypto() {
    if (typeof process !== 'undefined' && process.versions?.node) {
      if (!this.nodeCryptoModule) {
        this.nodeCryptoModule = import('node:crypto')
          .catch(() => import('crypto'))
          .catch(() => null);
      }
      return this.nodeCryptoModule;
    }
    return null;
  },

  /**
   * Normalize storage.get to work with both callback and promise APIs.
   */
  async storageGet(keys) {
    return new Promise((resolve, reject) => {
      try {
        const storage = chrome?.storage?.local;
        if (!storage || typeof storage.get !== 'function') {
          return reject(new Error('Storage API unavailable'));
        }

        let resolved = false;
        const done = (result = {}) => {
          if (resolved) return;
          resolved = true;
          resolve(result || {});
        };

        const maybePromise = storage.get(keys, done);

        if (maybePromise && typeof maybePromise.then === 'function') {
          maybePromise.then(done).catch(reject);
        } else if (storage.get.length === 1 && !resolved) {
          done(maybePromise);
        }
      } catch (error) {
        reject(error);
      }
    });
  },

  /**
   * Normalize storage.set to work with both callback and promise APIs.
   */
  async storageSet(data) {
    return new Promise((resolve, reject) => {
      try {
        const storage = chrome?.storage?.local;
        if (!storage || typeof storage.set !== 'function') {
          return reject(new Error('Storage API unavailable'));
        }

        const done = () => resolve();
        const maybePromise = storage.set(data, done);

        if (maybePromise && typeof maybePromise.then === 'function') {
          maybePromise.then(done).catch(reject);
        } else if (storage.set.length === 1) {
          done();
        }
      } catch (error) {
        reject(error);
      }
    });
  },

  /**
   * Normalize storage.remove to work with both callback and promise APIs.
   */
  async storageRemove(keys) {
    return new Promise((resolve, reject) => {
      try {
        const storage = chrome?.storage?.local;
        if (!storage || typeof storage.remove !== 'function') {
          return reject(new Error('Storage API unavailable'));
        }

        const done = () => resolve();
        const maybePromise = storage.remove(keys, done);

        if (maybePromise && typeof maybePromise.then === 'function') {
          maybePromise.then(done).catch(reject);
        } else if (storage.remove.length === 1) {
          done();
        }
      } catch (error) {
        reject(error);
      }
    });
  },

  /**
   * Storage keys
   */
  STORAGE_KEYS: {
    PRIVATE_KEY: 'walletPrivateKey',
    WALLET_ADDRESS: 'walletAddress',
    ENCRYPTED_KEY: 'encryptedPrivateKey',
    KEY_SALT: 'keySalt',
    IS_ENCRYPTED: 'isKeyEncrypted',
  },

  /**
   * Generate a new wallet
   * @returns {Promise<Object>} Wallet info (address, privateKey)
   */
  async generateWallet() {
    try {
      const privateKey = COSMOS_SDK.generatePrivateKey();

      if (privateKey.length !== 32) {
        throw new Error('Invalid private key length generated');
      }

      const privateKeyHex = COSMOS_SDK.bytesToHex(privateKey);
      const publicKey = await COSMOS_SDK.getPublicKey(privateKey);
      const address = COSMOS_SDK.publicKeyToAddress(publicKey);

      // Validate generated address
      if (!this.validateAddress(address)) {
        throw new Error('Generated address validation failed');
      }

      return {
        address,
        privateKeyHex,
        publicKey: COSMOS_SDK.bytesToHex(publicKey),
      };
    } catch (error) {
      console.error('Wallet generation error:', error);
      throw new Error(`Failed to generate wallet: ${error.message}`);
    }
  },

  /**
   * Import wallet from private key
   * @param {string} privateKeyHex - Private key in hex format
   * @returns {Promise<Object>} Wallet info
   */
  async importWallet(privateKeyHex) {
    try {
      // Validate input
      if (!privateKeyHex || typeof privateKeyHex !== 'string') {
        throw new Error('Private key is required');
      }

      // Clean and validate format
      privateKeyHex = privateKeyHex.trim().toLowerCase();
      if (!/^[0-9a-f]{64}$/i.test(privateKeyHex)) {
        throw new Error('Invalid private key format. Must be 64 hex characters (32 bytes)');
      }

      const privateKey = COSMOS_SDK.hexToBytes(privateKeyHex);

      if (privateKey.length !== 32) {
        throw new Error('Invalid private key length. Must be 32 bytes');
      }

      const publicKey = await COSMOS_SDK.getPublicKey(privateKey);
      const address = COSMOS_SDK.publicKeyToAddress(publicKey);

      if (!this.validateAddress(address)) {
        throw new Error('Generated address validation failed');
      }

      return {
        address,
        privateKeyHex,
        publicKey: COSMOS_SDK.bytesToHex(publicKey),
      };
    } catch (error) {
      console.error('Wallet import error:', error);
      throw new Error(`Failed to import wallet: ${error.message}`);
    }
  },

  /**
   * Save wallet to storage
   * @param {string} privateKeyHex - Private key in hex
   * @param {string} address - Wallet address
   * @param {string} password - Optional password for encryption
   * @returns {Promise<void>}
   */
  async saveWallet(privateKeyHex, address, password = null) {
    try {
      const storage = {};

      if (password) {
        // Encrypt private key with password
        const { encrypted, salt } = await this.encryptPrivateKey(privateKeyHex, password);
        storage[this.STORAGE_KEYS.ENCRYPTED_KEY] = encrypted;
        storage[this.STORAGE_KEYS.KEY_SALT] = salt;
        storage[this.STORAGE_KEYS.IS_ENCRYPTED] = true;
      } else {
        // Store unencrypted (not recommended for production)
        storage[this.STORAGE_KEYS.PRIVATE_KEY] = privateKeyHex;
        storage[this.STORAGE_KEYS.IS_ENCRYPTED] = false;
      }

      storage[this.STORAGE_KEYS.WALLET_ADDRESS] = address;

      await this.storageSet(storage);
    } catch (error) {
      console.error('Save wallet error:', error);
      throw new Error(`Failed to save wallet: ${error.message}`);
    }
  },

  /**
   * Load wallet from storage
   * @param {string} password - Optional password for decryption
   * @returns {Promise<Object>} Wallet info
   */
  async loadWallet(password = null) {
    try {
      const result = await this.storageGet([
        this.STORAGE_KEYS.PRIVATE_KEY,
        this.STORAGE_KEYS.ENCRYPTED_KEY,
        this.STORAGE_KEYS.KEY_SALT,
        this.STORAGE_KEYS.IS_ENCRYPTED,
        this.STORAGE_KEYS.WALLET_ADDRESS,
      ]);

      if (!result[this.STORAGE_KEYS.WALLET_ADDRESS]) {
        return null;
      }

      let privateKeyHex;

      if (result[this.STORAGE_KEYS.IS_ENCRYPTED]) {
        if (!password) {
          throw new Error('Password required to decrypt wallet');
        }

        privateKeyHex = await this.decryptPrivateKey(
          result[this.STORAGE_KEYS.ENCRYPTED_KEY],
          result[this.STORAGE_KEYS.KEY_SALT],
          password
        );
      } else {
        privateKeyHex = result[this.STORAGE_KEYS.PRIVATE_KEY];
      }

      if (!privateKeyHex) {
        return null;
      }

      const privateKey = COSMOS_SDK.hexToBytes(privateKeyHex);
      const publicKey = await COSMOS_SDK.getPublicKey(privateKey);
      const address = COSMOS_SDK.publicKeyToAddress(publicKey);

      return {
        address,
        privateKeyHex,
        publicKey: COSMOS_SDK.bytesToHex(publicKey),
      };
    } catch (error) {
      console.error('Load wallet error:', error);
      throw new Error(`Failed to load wallet: ${error.message}`);
    }
  },

  /**
   * Delete wallet from storage
   * @returns {Promise<void>}
   */
  async deleteWallet() {
    try {
      await this.storageRemove([
        this.STORAGE_KEYS.PRIVATE_KEY,
        this.STORAGE_KEYS.ENCRYPTED_KEY,
        this.STORAGE_KEYS.KEY_SALT,
        this.STORAGE_KEYS.IS_ENCRYPTED,
        this.STORAGE_KEYS.WALLET_ADDRESS,
      ]);
    } catch (error) {
      console.error('Delete wallet error:', error);
      throw new Error(`Failed to delete wallet: ${error.message}`);
    }
  },

  /**
   * Encrypt private key with password
   * @param {string} privateKeyHex - Private key in hex
  * @param {string} password - Password
  * @returns {Promise<Object>} Encrypted key and salt
  */
  async encryptPrivateKey(privateKeyHex, password) {
    try {
      const nodeCrypto = await this.getNodeCrypto();

      if (nodeCrypto) {
        const salt = nodeCrypto.randomBytes(16);
        const iv = nodeCrypto.randomBytes(12);
        const key = nodeCrypto.pbkdf2Sync(password, salt, 100000, 32, 'sha256');
        const cipher = nodeCrypto.createCipheriv('aes-256-gcm', key, iv);

        const encryptedBuffer = Buffer.concat([cipher.update(privateKeyHex, 'utf8'), cipher.final()]);
        const tag = cipher.getAuthTag();
        const combined = Buffer.concat([iv, encryptedBuffer, tag]);

        return {
          encrypted: this.bytesToBase64(combined),
          salt: this.bytesToBase64(salt),
        };
      }

      const cryptoApi = await this.getCrypto();

      // Generate salt
      const salt = cryptoApi.getRandomValues(new Uint8Array(16));

      // Derive key from password
      const encoder = new TextEncoder();
      const keyMaterial = await cryptoApi.subtle.importKey(
        'raw',
        encoder.encode(password),
        { name: 'PBKDF2' },
        false,
        ['deriveBits', 'deriveKey']
      );

      const key = await cryptoApi.subtle.deriveKey(
        {
          name: 'PBKDF2',
          salt: salt,
          iterations: 100000,
          hash: 'SHA-256',
        },
        keyMaterial,
        { name: 'AES-GCM', length: 256 },
        false,
        ['encrypt']
      );

      // Encrypt private key
      const iv = cryptoApi.getRandomValues(new Uint8Array(12));
      const encrypted = await cryptoApi.subtle.encrypt(
        { name: 'AES-GCM', iv: iv },
        key,
        encoder.encode(privateKeyHex)
      );

      // Combine IV and encrypted data
      const combined = new Uint8Array(iv.length + encrypted.byteLength);
      combined.set(iv, 0);
      combined.set(new Uint8Array(encrypted), iv.length);

      return {
        encrypted: this.bytesToBase64(combined),
        salt: this.bytesToBase64(salt),
      };
    } catch (error) {
      console.error('Encryption error:', error);
      throw new Error(`Failed to encrypt private key: ${error.message}`);
    }
  },

  /**
   * Decrypt private key with password
   * @param {string} encryptedBase64 - Encrypted key in base64
   * @param {string} saltBase64 - Salt in base64
   * @param {string} password - Password
   * @returns {Promise<string>} Decrypted private key in hex
  */
  async decryptPrivateKey(encryptedBase64, saltBase64, password) {
    const nodeCrypto = await this.getNodeCrypto();

    if (!nodeCrypto && process?.env?.NODE_ENV === 'test') {
      console.error('Node crypto unavailable in test runtime');
    }

    if (nodeCrypto) {
      try {
        const salt = this.base64ToBytes(saltBase64);
        const combined = this.base64ToBytes(encryptedBase64);
        const iv = combined.slice(0, 12);
        const cipherWithTag = combined.slice(12);

        if (process?.env?.NODE_ENV === 'test') {
          console.debug('Node decrypt path', {
            combinedLength: combined.length,
            cipherLength: cipherWithTag.length,
            ivLength: iv.length,
          });
        }

        if (cipherWithTag.length <= 16) {
          throw new Error('Invalid encrypted payload');
        }

        const authTag = cipherWithTag.slice(cipherWithTag.length - 16);
        const payload = cipherWithTag.slice(0, cipherWithTag.length - 16);

        const key = nodeCrypto.pbkdf2Sync(password, salt, 100000, 32, 'sha256');
        const decipher = nodeCrypto.createDecipheriv('aes-256-gcm', key, iv);
        decipher.setAuthTag(authTag);

        const decrypted = Buffer.concat([decipher.update(payload), decipher.final()]);
        return decrypted.toString('utf8');
      } catch (fallbackError) {
        console.error('Node crypto decrypt fallback failed:', fallbackError);
      }
    }

    try {
      const cryptoApi = await this.getCrypto();
      const encoder = new TextEncoder();
      const decoder = new TextDecoder();
      const salt = this.base64ToBytes(saltBase64);
      const combined = this.base64ToBytes(encryptedBase64);

      // Extract IV and encrypted data
      const iv = combined.slice(0, 12);
      const encrypted = combined.slice(12);

      if (encrypted.length <= 16) {
        const preview = typeof encryptedBase64 === 'string' ? encryptedBase64.slice(0, 24) : '[non-string]';
        throw new Error(
          `Encrypted payload too small (cipherLength=${encrypted.length}, total=${combined.length}, preview=${preview})`
        );
      }

      if (process?.env?.NODE_ENV === 'test') {
        console.error('WebCrypto decrypt path', {
          combinedLength: combined.length,
          cipherLength: encrypted.length,
          ivLength: iv.length,
        });
      }

      // Derive key from password
      const keyMaterial = await cryptoApi.subtle.importKey(
        'raw',
        encoder.encode(password),
        { name: 'PBKDF2' },
        false,
        ['deriveBits', 'deriveKey']
      );

      const key = await cryptoApi.subtle.deriveKey(
        {
          name: 'PBKDF2',
          salt: salt,
          iterations: 100000,
          hash: 'SHA-256',
        },
        keyMaterial,
        { name: 'AES-GCM', length: 256 },
        false,
        ['decrypt']
      );

      // Decrypt
      const ivBuffer = iv.buffer.slice(iv.byteOffset, iv.byteOffset + iv.byteLength);
      const cipherBuffer = encrypted.buffer.slice(
        encrypted.byteOffset,
        encrypted.byteOffset + encrypted.byteLength
      );

      const decrypted = await cryptoApi.subtle
        .decrypt({ name: 'AES-GCM', iv: ivBuffer }, key, cipherBuffer)
        .catch(err => {
          throw new Error(`${err.message} (cipherLength=${encrypted.length}, total=${combined.length})`);
        });

      return decoder.decode(decrypted);
    } catch (error) {
      console.error('Decryption error:', error);
      throw new Error(`Failed to decrypt private key. ${error.message || 'Wrong password?'}`);
    }
  },

  /**
   * Validate Cosmos address format
   * @param {string} address - Address to validate
   * @returns {boolean} Is valid
   */
  validateAddress(address) {
    if (!address || typeof address !== 'string') {
      return false;
    }
    const prefix = COSMOS_SDK?.config?.bech32Prefix || 'aura';
    const pattern = new RegExp(`^${prefix}1[0-9a-z]{10,90}$`, 'i');
    return pattern.test(address);
  },

  /**
   * Convert bytes to base64
   * @param {Uint8Array} bytes - Bytes
   * @returns {string} Base64 string
   */
  bytesToBase64(bytes) {
    if (typeof Buffer !== 'undefined') {
      return Buffer.from(bytes).toString('base64');
    }
    return btoa(String.fromCharCode(...bytes));
  },

  /**
   * Convert base64 to bytes
   * @param {string} base64 - Base64 string
   * @returns {Uint8Array} Bytes
   */
  base64ToBytes(base64) {
    if (typeof Buffer !== 'undefined') {
      return new Uint8Array(Buffer.from(base64, 'base64'));
    }
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  },
};

export default WalletCore;
