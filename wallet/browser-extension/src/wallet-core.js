/**
 * Wallet Core Module
 * Handles wallet creation, import, export, and key management
 */

const WalletCore = {
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

      await chrome.storage.local.set(storage);
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
      const result = await chrome.storage.local.get([
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
      await chrome.storage.local.remove([
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
      // Generate salt
      const salt = crypto.getRandomValues(new Uint8Array(16));

      // Derive key from password
      const encoder = new TextEncoder();
      const keyMaterial = await crypto.subtle.importKey(
        'raw',
        encoder.encode(password),
        { name: 'PBKDF2' },
        false,
        ['deriveBits', 'deriveKey']
      );

      const key = await crypto.subtle.deriveKey(
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
      const iv = crypto.getRandomValues(new Uint8Array(12));
      const encrypted = await crypto.subtle.encrypt(
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
    try {
      const encoder = new TextEncoder();
      const decoder = new TextDecoder();
      const salt = this.base64ToBytes(saltBase64);
      const combined = this.base64ToBytes(encryptedBase64);

      // Extract IV and encrypted data
      const iv = combined.slice(0, 12);
      const encrypted = combined.slice(12);

      // Derive key from password
      const keyMaterial = await crypto.subtle.importKey(
        'raw',
        encoder.encode(password),
        { name: 'PBKDF2' },
        false,
        ['deriveBits', 'deriveKey']
      );

      const key = await crypto.subtle.deriveKey(
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
      const decrypted = await crypto.subtle.decrypt(
        { name: 'AES-GCM', iv: iv },
        key,
        encrypted
      );

      return decoder.decode(decrypted);
    } catch (error) {
      console.error('Decryption error:', error);
      throw new Error('Failed to decrypt private key. Wrong password?');
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
    // Check prefix and length
    return address.startsWith(COSMOS_SDK.config.bech32Prefix) && address.length >= 39;
  },

  /**
   * Convert bytes to base64
   * @param {Uint8Array} bytes - Bytes
   * @returns {string} Base64 string
   */
  bytesToBase64(bytes) {
    return btoa(String.fromCharCode(...bytes));
  },

  /**
   * Convert base64 to bytes
   * @param {string} base64 - Base64 string
   * @returns {Uint8Array} Bytes
   */
  base64ToBytes(base64) {
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  },
};

// Export for use in extension
if (typeof module !== 'undefined' && module.exports) {
  module.exports = WalletCore;
}
