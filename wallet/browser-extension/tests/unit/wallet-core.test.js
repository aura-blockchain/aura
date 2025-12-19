/**
 * Unit Tests for Wallet Core Module
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

// Mock chrome storage API
global.chrome = {
  storage: {
    local: {
      get: vi.fn(),
      set: vi.fn(),
      remove: vi.fn(),
    },
  },
};

// Mock COSMOS_SDK
global.COSMOS_SDK = {
  config: {
    bech32Prefix: 'aura',
    coinDenom: 'uaura',
    coinDecimals: 6,
  },
  generatePrivateKey: vi.fn(),
  getPublicKey: vi.fn(),
  publicKeyToAddress: vi.fn(),
  bytesToHex: vi.fn(),
  hexToBytes: vi.fn(),
};

describe('WalletCore', () => {
  let WalletCore;

  beforeEach(async () => {
    // Import module
    WalletCore = (await import('../../src/wallet-core.js')).default || (await import('../../src/wallet-core.js'));

    // Reset mocks
    vi.clearAllMocks();

    // Setup default mock implementations
    chrome.storage.local.get.mockImplementation((keys, callback) => {
      callback({});
    });

    chrome.storage.local.set.mockImplementation((data, callback) => {
      if (callback) callback();
    });

    chrome.storage.local.remove.mockImplementation((keys, callback) => {
      if (callback) callback();
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('generateWallet', () => {
    it('should generate a new wallet with valid address and private key', async () => {
      const mockPrivateKey = new Uint8Array(32).fill(1);
      const mockPublicKey = new Uint8Array(33).fill(2);
      const mockAddress = 'aura1test123456789012345678901234567890';

      COSMOS_SDK.generatePrivateKey.mockReturnValue(mockPrivateKey);
      COSMOS_SDK.getPublicKey.mockResolvedValue(mockPublicKey);
      COSMOS_SDK.publicKeyToAddress.mockReturnValue(mockAddress);
      COSMOS_SDK.bytesToHex.mockImplementation(bytes =>
        Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('')
      );

      const wallet = await WalletCore.generateWallet();

      expect(wallet).toHaveProperty('address');
      expect(wallet).toHaveProperty('privateKeyHex');
      expect(wallet).toHaveProperty('publicKey');
      expect(wallet.address).toBe(mockAddress);
      expect(COSMOS_SDK.generatePrivateKey).toHaveBeenCalled();
      expect(COSMOS_SDK.getPublicKey).toHaveBeenCalledWith(mockPrivateKey);
    });

    it('should throw error if private key length is invalid', async () => {
      const mockPrivateKey = new Uint8Array(16); // Invalid length

      COSMOS_SDK.generatePrivateKey.mockReturnValue(mockPrivateKey);

      await expect(WalletCore.generateWallet()).rejects.toThrow('Invalid private key length');
    });

    it('should throw error if address validation fails', async () => {
      const mockPrivateKey = new Uint8Array(32).fill(1);
      const mockPublicKey = new Uint8Array(33).fill(2);
      const mockAddress = 'invalid_address';

      COSMOS_SDK.generatePrivateKey.mockReturnValue(mockPrivateKey);
      COSMOS_SDK.getPublicKey.mockResolvedValue(mockPublicKey);
      COSMOS_SDK.publicKeyToAddress.mockReturnValue(mockAddress);
      COSMOS_SDK.bytesToHex.mockImplementation(bytes =>
        Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('')
      );

      await expect(WalletCore.generateWallet()).rejects.toThrow('address validation failed');
    });
  });

  describe('importWallet', () => {
    it('should import wallet from valid private key', async () => {
      const privateKeyHex = '0'.repeat(64); // 64 hex characters
      const mockPrivateKey = new Uint8Array(32);
      const mockPublicKey = new Uint8Array(33).fill(2);
      const mockAddress = 'aura1test123456789012345678901234567890';

      COSMOS_SDK.hexToBytes.mockReturnValue(mockPrivateKey);
      COSMOS_SDK.getPublicKey.mockResolvedValue(mockPublicKey);
      COSMOS_SDK.publicKeyToAddress.mockReturnValue(mockAddress);
      COSMOS_SDK.bytesToHex.mockImplementation(bytes =>
        Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('')
      );

      const wallet = await WalletCore.importWallet(privateKeyHex);

      expect(wallet.address).toBe(mockAddress);
      expect(wallet.privateKeyHex).toBe(privateKeyHex);
    });

    it('should throw error for invalid private key format', async () => {
      await expect(WalletCore.importWallet('invalid')).rejects.toThrow('Invalid private key format');
      await expect(WalletCore.importWallet('0'.repeat(63))).rejects.toThrow('Invalid private key format');
      await expect(WalletCore.importWallet('g'.repeat(64))).rejects.toThrow('Invalid private key format');
    });

    it('should throw error for empty private key', async () => {
      await expect(WalletCore.importWallet('')).rejects.toThrow('Private key is required');
      await expect(WalletCore.importWallet(null)).rejects.toThrow('Private key is required');
    });

    it('should handle private key with uppercase letters', async () => {
      const privateKeyHex = 'A'.repeat(64);
      const mockPrivateKey = new Uint8Array(32);
      const mockPublicKey = new Uint8Array(33).fill(2);
      const mockAddress = 'aura1test123456789012345678901234567890';

      COSMOS_SDK.hexToBytes.mockReturnValue(mockPrivateKey);
      COSMOS_SDK.getPublicKey.mockResolvedValue(mockPublicKey);
      COSMOS_SDK.publicKeyToAddress.mockReturnValue(mockAddress);
      COSMOS_SDK.bytesToHex.mockImplementation(bytes =>
        Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('')
      );

      const wallet = await WalletCore.importWallet(privateKeyHex);

      expect(wallet.address).toBe(mockAddress);
      // Should be normalized to lowercase
      expect(COSMOS_SDK.hexToBytes).toHaveBeenCalledWith(privateKeyHex.toLowerCase());
    });
  });

  describe('saveWallet', () => {
    it('should save wallet without encryption', async () => {
      const privateKeyHex = '0'.repeat(64);
      const address = 'aura1test123456789012345678901234567890';

      await WalletCore.saveWallet(privateKeyHex, address);

      expect(chrome.storage.local.set).toHaveBeenCalledWith(
        expect.objectContaining({
          walletPrivateKey: privateKeyHex,
          walletAddress: address,
          isKeyEncrypted: false,
        }),
        expect.any(Function)
      );
    });

    it('should save wallet with encryption when password provided', async () => {
      const privateKeyHex = '0'.repeat(64);
      const address = 'aura1test123456789012345678901234567890';
      const password = 'test-password';

      // Mock encryption
      const mockEncrypted = 'encrypted-data';
      const mockSalt = 'salt-data';
      vi.spyOn(WalletCore, 'encryptPrivateKey').mockResolvedValue({
        encrypted: mockEncrypted,
        salt: mockSalt,
      });

      await WalletCore.saveWallet(privateKeyHex, address, password);

      expect(WalletCore.encryptPrivateKey).toHaveBeenCalledWith(privateKeyHex, password);
      expect(chrome.storage.local.set).toHaveBeenCalledWith(
        expect.objectContaining({
          encryptedPrivateKey: mockEncrypted,
          keySalt: mockSalt,
          walletAddress: address,
          isKeyEncrypted: true,
        }),
        expect.any(Function)
      );
    });
  });

  describe('loadWallet', () => {
    it('should load unencrypted wallet', async () => {
      const privateKeyHex = '0'.repeat(64);
      const address = 'aura1test123456789012345678901234567890';
      const mockPrivateKey = new Uint8Array(32);
      const mockPublicKey = new Uint8Array(33).fill(2);

      chrome.storage.local.get.mockImplementation((keys, callback) => {
        callback({
          walletPrivateKey: privateKeyHex,
          walletAddress: address,
          isKeyEncrypted: false,
        });
      });

      COSMOS_SDK.hexToBytes.mockReturnValue(mockPrivateKey);
      COSMOS_SDK.getPublicKey.mockResolvedValue(mockPublicKey);
      COSMOS_SDK.publicKeyToAddress.mockReturnValue(address);
      COSMOS_SDK.bytesToHex.mockImplementation(bytes =>
        Array.from(bytes).map(b => b.toString(16).padStart(2, '0')).join('')
      );

      const wallet = await WalletCore.loadWallet();

      expect(wallet).not.toBeNull();
      expect(wallet.address).toBe(address);
      expect(wallet.privateKeyHex).toBe(privateKeyHex);
    });

    it('should return null if no wallet exists', async () => {
      chrome.storage.local.get.mockImplementation((keys, callback) => {
        callback({});
      });

      const wallet = await WalletCore.loadWallet();

      expect(wallet).toBeNull();
    });

    it('should throw error if encrypted wallet loaded without password', async () => {
      chrome.storage.local.get.mockImplementation((keys, callback) => {
        callback({
          encryptedPrivateKey: 'encrypted',
          keySalt: 'salt',
          walletAddress: 'aura1test',
          isKeyEncrypted: true,
        });
      });

      await expect(WalletCore.loadWallet()).rejects.toThrow('Password required');
    });
  });

  describe('deleteWallet', () => {
    it('should remove all wallet data from storage', async () => {
      await WalletCore.deleteWallet();

      expect(chrome.storage.local.remove).toHaveBeenCalledWith(
        expect.arrayContaining([
          'walletPrivateKey',
          'encryptedPrivateKey',
          'keySalt',
          'isKeyEncrypted',
          'walletAddress',
        ]),
        expect.any(Function)
      );
    });
  });

  describe('validateAddress', () => {
    it('should validate correct addresses', () => {
      expect(WalletCore.validateAddress('aura1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a')).toBe(true);
      expect(WalletCore.validateAddress('aura1test1234567890123456789012345678')).toBe(true);
    });

    it('should reject invalid addresses', () => {
      expect(WalletCore.validateAddress('cosmos1test')).toBe(false);
      expect(WalletCore.validateAddress('aura1')).toBe(false);
      expect(WalletCore.validateAddress('')).toBe(false);
      expect(WalletCore.validateAddress(null)).toBe(false);
      expect(WalletCore.validateAddress(undefined)).toBe(false);
      expect(WalletCore.validateAddress(123)).toBe(false);
    });
  });

  describe('encryptPrivateKey', () => {
    it('should encrypt private key with password', async () => {
      const privateKeyHex = '0'.repeat(64);
      const password = 'test-password';

      const result = await WalletCore.encryptPrivateKey(privateKeyHex, password);

      expect(result).toHaveProperty('encrypted');
      expect(result).toHaveProperty('salt');
      expect(typeof result.encrypted).toBe('string');
      expect(typeof result.salt).toBe('string');
    });
  });

  describe('decryptPrivateKey', () => {
    it('should decrypt private key with correct password', async () => {
      const privateKeyHex = '0'.repeat(64);
      const password = 'test-password';

      const encrypted = await WalletCore.encryptPrivateKey(privateKeyHex, password);
      const decrypted = await WalletCore.decryptPrivateKey(
        encrypted.encrypted,
        encrypted.salt,
        password
      );

      expect(decrypted).toBe(privateKeyHex);
    });

    it('should throw error with wrong password', async () => {
      const privateKeyHex = '0'.repeat(64);
      const password = 'test-password';
      const wrongPassword = 'wrong-password';

      const encrypted = await WalletCore.encryptPrivateKey(privateKeyHex, password);

      await expect(
        WalletCore.decryptPrivateKey(encrypted.encrypted, encrypted.salt, wrongPassword)
      ).rejects.toThrow();
    });
  });

  describe('bytesToBase64 and base64ToBytes', () => {
    it('should convert bytes to base64 and back', () => {
      const original = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8]);
      const base64 = WalletCore.bytesToBase64(original);
      const converted = WalletCore.base64ToBytes(base64);

      expect(converted).toEqual(original);
    });

    it('should handle empty arrays', () => {
      const original = new Uint8Array([]);
      const base64 = WalletCore.bytesToBase64(original);
      const converted = WalletCore.base64ToBytes(base64);

      expect(converted).toEqual(original);
    });
  });
});
